package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/internal/config"
	"github.com/spf13/cobra"
)

type statusOutput struct {
	Profile   string `json:"profile"`
	APIKey    string `json:"api_key"`
	Workspace string `json:"workspace"`
	Region    string `json:"region"`
	Output    string `json:"output"`
	Valid     string `json:"valid"`
	Config    string `json:"config"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication and configuration status",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	outputFormat, _ := cmd.Flags().GetString("output")

	cfg, err := config.Load(profile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey, err := config.GetAPIKey(profile)
	if err != nil {
		apiKey = "(not configured)"
	}

	maskedKey := maskAPIKey(apiKey)
	validStatus := "unknown"
	if apiKey != "(not configured)" && apiKey != "" {
		validStatus = validateKeyStatus(apiKey, cfg.Region)
	}

	out := statusOutput{
		Profile:   profileOrDefault(profile),
		APIKey:    maskedKey,
		Workspace: cfg.WorkspaceID,
		Region:    cfg.Region,
		Output:    cfg.Output,
		Valid:     validStatus,
		Config:    config.ConfigPath(),
	}

	if strings.ToLower(outputFormat) == "json" {
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}
		cmd.Println(string(data))
		return nil
	}

	validIndicator := formatValidIndicator(validStatus)
	cmd.Printf(`Profile:    %s
API Key:    %s
Workspace:  %s
Region:     %s
Output:     %s
Valid:      %s
Config:     %s`,
		out.Profile,
		out.APIKey,
		out.Workspace,
		out.Region,
		out.Output,
		validIndicator,
		out.Config,
	)
	cmd.Println()
	return nil
}

func validateKeyStatus(apiKey, region string) string {
	client := api.NewClient(apiKey, api.BaseURL(region), "")
	var user map[string]any
	err := client.Do(context.Background(), "GET", "/user", nil, &user)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "context deadline exceeded") ||
			strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "no such host") ||
			strings.Contains(errStr, "request canceled") {
			return "unreachable"
		}
		return "invalid"
	}
	return "valid"
}

func formatValidIndicator(status string) string {
	switch status {
	case "valid":
		return "valid"
	case "invalid":
		return "invalid - run 'clockify configure' to re-authenticate"
	case "unreachable":
		return "unable to validate (offline?)"
	default:
		return "unknown"
	}
}

func maskAPIKey(key string) string {
	if key == "" || key == "(not configured)" {
		return key
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func profileOrDefault(profile string) string {
	if profile == "" {
		return "default"
	}
	return profile
}
