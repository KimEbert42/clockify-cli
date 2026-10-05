package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/kim/clockify-cli/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  `View and modify Clockify CLI configuration settings.`,
	}
	cmd.AddCommand(newConfigListCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigUnsetCmd())
	cmd.AddCommand(newConfigPathCmd())
	cmd.AddCommand(newConfigDeleteCmd())
	cmd.AddCommand(newConfigClearCacheCmd())
	return cmd
}

func printConfig(w io.Writer, name string, cfg *config.Config) {
	fmt.Fprintf(w, "[%s]\n", name)
	fmt.Fprintf(w, "  Region:      %s\n", cfg.Region)
	fmt.Fprintf(w, "  Output:      %s\n", cfg.Output)
	fmt.Fprintf(w, "  Workspace:   %s\n", cfg.WorkspaceID)
	fmt.Fprintf(w, "  Date Format: %s\n", cfg.DateFormat)
	fmt.Fprintf(w, "  Time Format: %s\n", cfg.TimeFormat)
	fmt.Fprintf(w, "  Week Starts: %s\n", cfg.WeekStarts)
	fmt.Fprintf(w, "  Auto Update: %s\n", boolLabel(cfg.AutoUpdateEnabled()))
	fmt.Fprintf(w, "  Interactive: %s\n", boolLabel(cfg.InteractiveEnabled()))
	fmt.Fprintf(w, "  Time Zone:   %s\n", cfg.TimeZone)
}

func boolLabel(enabled bool) string {
	if enabled {
		return "true"
	}
	return "false"
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			out := cmd.OutOrStdout()

			if profile != "" {
				cfg, err := config.Load(profile)
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				printConfig(out, profile, cfg)
				return nil
			}

			pc, err := config.LoadRaw()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			printConfig(out, "default", &pc.Default)

			names := make([]string, 0, len(pc.Profiles))
			for name := range pc.Profiles {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				p := pc.Profiles[name]
				fmt.Fprintln(out)
				printConfig(out, name, &p)
			}
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
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
				cmd.Println(cfg.Region)
			case "output":
				cmd.Println(cfg.Output)
			case "workspace", "workspace_id":
				cmd.Println(cfg.WorkspaceID)
			case "date_format":
				cmd.Println(cfg.DateFormat)
			case "time_format":
				cmd.Println(cfg.TimeFormat)
			case "week_starts":
				cmd.Println(cfg.WeekStarts)
			case "auto_update":
				cmd.Println(boolLabel(cfg.AutoUpdateEnabled()))
			case "interactive":
				cmd.Println(boolLabel(cfg.InteractiveEnabled()))
			case "timezone":
				cmd.Println(cfg.TimeZone)
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			if profile == "" {
				profile = "default"
			}

			key, value := args[0], args[1]

			pc, err := config.LoadRaw()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			var cfg *config.Config
			if profile == "default" {
				cfg = &pc.Default
			} else {
				p, ok := pc.Profiles[profile]
				cfg = &p
				if !ok {
					if pc.Profiles == nil {
						pc.Profiles = make(map[string]config.Config)
					}
				}
			}

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
			case "auto_update":
				norm, err := config.ParseAutoUpdate(value)
				if err != nil {
					return err
				}
				cfg.AutoUpdate = norm
			case "interactive":
				norm, err := config.ParseInteractive(value)
				if err != nil {
					return err
				}
				cfg.Interactive = norm
			case "timezone":
				cfg.TimeZone = value
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			if profile == "default" {
				pc.Default = *cfg
			} else {
				pc.Profiles[profile] = *cfg
			}

			if err := config.Save(pc); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			cmd.Printf("Set %s = %s (profile: %s)\n", key, value, profile)
			return nil
		},
	}
}

func newConfigUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset [key]",
		Short: "Remove a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			if profile == "" {
				profile = "default"
			}

			key := args[0]

			pc, err := config.LoadRaw()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			var cfg *config.Config
			if profile == "default" {
				cfg = &pc.Default
			} else {
				p := pc.Profiles[profile]
				cfg = &p
			}

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
			case "auto_update":
				cfg.AutoUpdate = ""
			case "interactive":
				cfg.Interactive = ""
			case "timezone":
				cfg.TimeZone = ""
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			if profile == "default" {
				pc.Default = *cfg
			} else {
				pc.Profiles[profile] = *cfg
			}

			if err := config.Save(pc); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			cmd.Printf("Unset %s (profile: %s)\n", key, profile)
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(config.ConfigPath())
		},
	}
}

func newConfigDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a configuration profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			if profile == "" || profile == "default" {
				return fmt.Errorf("must specify a non-default profile with --profile")
			}

			confirm, _ := cmd.Flags().GetBool("confirm")
			if !confirm {
				cmd.Printf("Delete profile %q? This cannot be undone.\n", profile)
				cmd.Print("Type 'yes' to confirm: ")
				var input string
				fmt.Scanln(&input)
				if input != "yes" {
					cmd.Println("Aborted.")
					return nil
				}
			}

			if err := config.DeleteProfile(profile); err != nil {
				return err
			}

			if err := config.DeleteAPIKey(profile); err != nil {
				cmd.Printf("Warning: could not delete API key from keychain: %v\n", err)
			}

			cmd.Printf("Profile %q deleted.\n", profile)
			return nil
		},
	}
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}

func newConfigClearCacheCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear-cache",
		Short: "Clear cached data",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")

			cleared, err := config.ClearCache(profile)
			if err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}

			if cleared {
				if profile != "" && profile != "default" {
					cmd.Printf("Cache cleared for profile %q.\n", profile)
				} else {
					cmd.Println("Cache cleared.")
				}
			} else {
				cmd.Println("No cache to clear.")
			}
			return nil
		},
	}
}
