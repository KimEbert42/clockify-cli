package cli

import (
	"github.com/spf13/cobra"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func SetBuildInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

var rootCmd = &cobra.Command{
	Use:   "clockify",
	Short: "A CLI tool for Clockify time tracking",
	Long: `Clockify CLI - A command-line interface for Clockify time tracking.

Manage workspaces, projects, time entries, and reports directly from your terminal.
Supports multiple output formats (table, JSON, CSV, YAML) and shell completion.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().String("profile", "", "config profile to use")
	rootCmd.PersistentFlags().String("workspace", "", "workspace ID or name")
	rootCmd.PersistentFlags().String("output", "", "output format: table, json, csv, yaml (default: auto-detect)")
	rootCmd.PersistentFlags().String("region", "", "API region: global, eu-de, us, uk, au")
	rootCmd.PersistentFlags().String("api-key", "", "Clockify API key (overrides config)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().CountP("verbose", "v", "verbose output (can be repeated for debug)")
	rootCmd.PersistentFlags().Bool("quiet", false, "suppress non-essential output")
	rootCmd.PersistentFlags().Bool("confirm", false, "skip confirmation prompts")
	rootCmd.PersistentFlags().Int("page", 1, "page number for paginated results")
	rootCmd.PersistentFlags().Int("page-size", 50, "number of items per page")
	rootCmd.PersistentFlags().Bool("no-pagination", false, "fetch all pages automatically")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(completionCmd)
	registerAliases(rootCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("clockify %s (%s) built on %s\n", buildVersion, buildCommit, buildDate)
	},
}
