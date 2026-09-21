package cli

import (
	"fmt"

	"github.com/kim/clockify-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `View and modify Clockify CLI configuration settings.`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		cfg, err := config.Load(profile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		fmt.Printf("Region:      %s\n", cfg.Region)
		fmt.Printf("Output:      %s\n", cfg.Output)
		fmt.Printf("Workspace:   %s\n", cfg.WorkspaceID)
		fmt.Printf("Date Format: %s\n", cfg.DateFormat)
		fmt.Printf("Time Format: %s\n", cfg.TimeFormat)
		fmt.Printf("Week Starts: %s\n", cfg.WeekStarts)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		cfg, err := config.Load(profile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		key := args[0]
		switch key {
		case "region":
			fmt.Println(cfg.Region)
		case "output":
			fmt.Println(cfg.Output)
		case "workspace", "workspace_id":
			fmt.Println(cfg.WorkspaceID)
		case "date_format":
			fmt.Println(cfg.DateFormat)
		case "time_format":
			fmt.Println(cfg.TimeFormat)
		case "week_starts":
			fmt.Println(cfg.WeekStarts)
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		cfg, err := config.Load(profile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		key, value := args[0], args[1]
		switch key {
		case "region":
			cfg.Region = value
		case "output":
			cfg.Output = value
		case "workspace", "workspace_id":
			cfg.WorkspaceID = value
		case "date_format":
			cfg.DateFormat = value
		case "time_format":
			cfg.TimeFormat = value
		case "week_starts":
			cfg.WeekStarts = value
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
		pc := &config.ProfileConfig{
			Version: 1,
			Default: *cfg,
		}
		if err := config.Save(pc); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset [key]",
	Short: "Remove a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		cfg, err := config.Load(profile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		key := args[0]
		switch key {
		case "region":
			cfg.Region = "global"
		case "output":
			cfg.Output = ""
		case "workspace", "workspace_id":
			cfg.WorkspaceID = ""
		case "date_format":
			cfg.DateFormat = "2006-01-02"
		case "time_format":
			cfg.TimeFormat = "15:04"
		case "week_starts":
			cfg.WeekStarts = "MONDAY"
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
		pc := &config.ProfileConfig{
			Version: 1,
			Default: *cfg,
		}
		if err := config.Save(pc); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Unset %s\n", key)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(config.ConfigPath())
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configPathCmd)
}
