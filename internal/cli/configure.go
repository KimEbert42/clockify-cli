package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/internal/config"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactive setup wizard for Clockify CLI",
	Long:  `Set up your Clockify API key, workspace, and preferences.`,
	RunE:  runConfigure,
}

func runConfigure(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	if profile == "" {
		profile = "default"
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your Clockify API key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	fmt.Print("Enter region (global/eu-de/us/uk/au) [global]: ")
	region, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read region: %w", err)
	}
	region = strings.TrimSpace(region)
	if region == "" {
		region = "global"
	}

	fmt.Println("Validating API key...")
	user, err := validateAPIKey(apiKey, region)
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}
	fmt.Printf("  Authenticated as: %s (%s)\n", user["name"], user["email"])

	fmt.Print("Enter workspace ID (or leave empty to auto-detect): ")
	workspaceID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read workspace ID: %w", err)
	}
	workspaceID = strings.TrimSpace(workspaceID)

	if workspaceID == "" {
		fmt.Println("Attempting to auto-detect workspace...")
		workspaces, err := fetchWorkspaces(apiKey, region)
		if err != nil {
			fmt.Printf("Warning: could not fetch workspaces: %v\n", err)
			fmt.Print("Please enter workspace ID manually: ")
			workspaceID, _ = reader.ReadString('\n')
			workspaceID = strings.TrimSpace(workspaceID)
		} else if len(workspaces) > 0 {
			fmt.Printf("Found %d workspace(s):\n", len(workspaces))
			for i, ws := range workspaces {
				fmt.Printf("  %d. %s (%s)\n", i+1, ws["name"], ws["id"])
			}
			fmt.Print("Select workspace number [1]: ")
			choiceStr, _ := reader.ReadString('\n')
			choiceStr = strings.TrimSpace(choiceStr)
			if choiceStr == "" {
				choiceStr = "1"
			}
			choice, err := strconv.Atoi(choiceStr)
			if err != nil || choice < 1 || choice > len(workspaces) {
				return fmt.Errorf("invalid workspace selection: %s", choiceStr)
			}
			workspaceID = workspaces[choice-1]["id"].(string)
		}
	}

	if err := config.SetAPIKey(profile, apiKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	pc, err := config.LoadRaw()
	if err != nil {
		return fmt.Errorf("failed to load existing config: %w", err)
	}

	newCfg := config.Config{
		Region:      region,
		WorkspaceID: workspaceID,
	}

	if profile == "default" {
		pc.Default = newCfg
	} else {
		if pc.Profiles == nil {
			pc.Profiles = make(map[string]config.Config)
		}
		pc.Profiles[profile] = newCfg
	}

	if err := config.Save(pc); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\nConfiguration saved successfully! (profile: %s)\n", profile)
	fmt.Printf("  Region:    %s\n", region)
	fmt.Printf("  Workspace: %s\n", workspaceID)
	return nil
}

func validateAPIKey(apiKey, region string) (map[string]any, error) {
	client := api.NewClient(apiKey, api.BaseURL(region), "")
	var user map[string]any
	err := client.Do(context.Background(), "GET", "/user", nil, &user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func fetchWorkspaces(apiKey, region string) ([]map[string]any, error) {
	client := api.NewClient(apiKey, api.BaseURL(region), "")
	var workspaces []map[string]any
	err := client.Do(context.Background(), "GET", "/workspaces", nil, &workspaces)
	return workspaces, err
}
