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
	"golang.org/x/term"
)

// readSecretLine prints prompt and reads a line without echoing when stdin is a TTY.
func readSecretLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return line, err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

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

	apiKey, err := readSecretLine("Enter your Clockify API key: ")
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
				name, _ := ws["name"].(string)
				id, _ := ws["id"].(string)
				fmt.Printf("  %d. %s (%s)\n", i+1, name, id)
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
			id, _ := workspaces[choice-1]["id"].(string)
			if id == "" {
				return fmt.Errorf("workspace %d has no usable id field", choice)
			}
			workspaceID = id
		}
	}

	if err := config.SetAPIKey(profile, apiKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	pc, err := config.LoadRaw()
	if err != nil {
		return fmt.Errorf("failed to load existing config: %w", err)
	}

	// Merge-on-save: only overlay keys the wizard collected.
	var base config.Config
	if profile == "default" {
		base = pc.Default
	} else {
		base = pc.Profiles[profile]
	}
	if region != "" {
		base.Region = region
	}
	if workspaceID != "" {
		base.WorkspaceID = workspaceID
	}

	if profile == "default" {
		pc.Default = base
	} else {
		if pc.Profiles == nil {
			pc.Profiles = make(map[string]config.Config)
		}
		pc.Profiles[profile] = base
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
