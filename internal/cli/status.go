package cli

import (
	"fmt"

	"github.com/kim/clockify-cli/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication and configuration status",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	cfg, err := config.Load(profile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey, err := config.GetAPIKey(profile)
	if err != nil {
		apiKey = "(not configured)"
	}

	maskedKey := maskAPIKey(apiKey)
	validStatus := "Unknown"
	if apiKey != "(not configured)" && apiKey != "" {
		if _, err := validateAPIKey(apiKey, cfg.Region); err != nil {
			validStatus = "Invalid"
		} else {
			validStatus = "Valid"
		}
	}

	output := fmt.Sprintf(`Profile:    %s
API Key:    %s
Workspace:  %s
Region:     %s
Output:     %s
Valid:      %s
Config:     %s`,
		profileOrDefault(profile),
		maskedKey,
		cfg.WorkspaceID,
		cfg.Region,
		cfg.Output,
		validStatus,
		config.ConfigPath(),
	)

	cmd.Println(output)
	return nil
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
