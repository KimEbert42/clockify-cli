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

var rootCmd = newRootCmd()

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clockify",
		Short: "A CLI tool for Clockify time tracking",
		Long: `Clockify CLI - A command-line interface for Clockify time tracking.

Manage workspaces, projects, time entries, and reports directly from your terminal.
Supports multiple output formats (table, JSON, CSV, YAML) and shell completion.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().String("profile", "", "config profile to use")
	cmd.PersistentFlags().String("workspace", "", "workspace ID or name")
	cmd.PersistentFlags().String("output", "", "output format: table, json, csv, yaml (default: auto-detect)")
	cmd.PersistentFlags().String("region", "", "API region: global, eu-de, us, uk, au")
	cmd.PersistentFlags().String("api-key", "", "Clockify API key (overrides config)")
	cmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	cmd.PersistentFlags().CountP("verbose", "v", "verbose output (can be repeated for debug)")
	cmd.PersistentFlags().Bool("quiet", false, "suppress non-essential output")
	cmd.PersistentFlags().Bool("confirm", false, "skip confirmation prompts")
	cmd.PersistentFlags().Int("page", 1, "page number for paginated results")
	cmd.PersistentFlags().Int("page-size", 50, "number of items per page")
	cmd.PersistentFlags().Bool("no-pagination", false, "fetch all pages automatically")

	cmd.AddCommand(versionCmd)
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(configureCmd)
	cmd.AddCommand(statusCmd)
	cmd.AddCommand(completionCmd)
	cmd.AddCommand(newWorkspaceCmd())
	cmd.AddCommand(newUserCmd())
	registerAliases(cmd)

	return cmd
}

func Execute() error {
	return rootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("clockify %s (%s) built on %s\n", buildVersion, buildCommit, buildDate)
	},
}
