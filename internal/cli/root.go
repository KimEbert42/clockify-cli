package cli

import (
	"encoding/json"
	"runtime"

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

// BuildInfo returns the build metadata injected at compile time.
type BuildInfo struct {
	Version string `json:"version" yaml:"version"`
	Commit  string `json:"commit" yaml:"commit"`
	Date    string `json:"date" yaml:"date"`
	Go      string `json:"go" yaml:"go"`
	Platform string `json:"platform" yaml:"platform"`
}

func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:  buildVersion,
		Commit:   buildCommit,
		Date:     buildDate,
		Go:       runtime.Version(),
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}
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
	cmd.PersistentFlags().String("output", "", "output format: table, json, csv, yaml, template, ids (default: auto-detect)")
	cmd.PersistentFlags().String("template", "", "template string for --output template (Go text/template)")
	cmd.PersistentFlags().String("region", "", "API region: global, eu-de, us, uk, au")
	cmd.PersistentFlags().String("api-key", "", "Clockify API key (overrides config)")
	cmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	cmd.PersistentFlags().CountP("verbose", "v", "verbose output (can be repeated for debug)")
	cmd.PersistentFlags().Bool("quiet", false, "suppress non-essential output")
	cmd.PersistentFlags().Bool("confirm", false, "skip confirmation prompts")
	cmd.PersistentFlags().Int("page", 1, "page number for paginated results")
	cmd.PersistentFlags().Int("page-size", 50, "number of items per page")
	cmd.PersistentFlags().Bool("no-pagination", false, "fetch all pages automatically")
	cmd.PersistentFlags().Bool("no-cache", false, "bypass local cache and fetch fresh data")

	// Passive update check — never fails the primary command.
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		maybeCheckForUpdate(c)
	}

	cmd.AddCommand(versionCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(configureCmd)
	cmd.AddCommand(statusCmd)
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newWorkspaceCmd())
	cmd.AddCommand(newUserCmd())
	cmd.AddCommand(newProjectCmd())
	cmd.AddCommand(newClientCmd())
	cmd.AddCommand(newTagCmd())
	cmd.AddCommand(newTaskCmd())
	cmd.AddCommand(newTimeEntryCmd())
	cmd.AddCommand(newReportCmd())
	cmd.AddCommand(newInvoiceCmd())
	cmd.AddCommand(newScheduleCmd())
	cmd.AddCommand(newTimeOffCmd())
	cmd.AddCommand(newExpenseCmd())
	cmd.AddCommand(newCustomFieldCmd())
	cmd.AddCommand(newWebhookCmd())
	cmd.AddCommand(newApprovalCmd())
	cmd.AddCommand(newGroupCmd())
	cmd.AddCommand(newUpdateCmd())
	registerAliases(cmd)

	// Dynamic completion for --workspace and resource --name/--project/--client flags.
	_ = cmd.RegisterFlagCompletionFunc("workspace", completeWorkspaceFlagC)
	_ = cmd.RegisterFlagCompletionFunc("profile", completeProfileFlagC)

	return cmd
}

func Execute() error {
	return rootCmd.Execute()
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long: `Show version, commit, build date, Go version, and platform.

Examples:
  clockify version
  clockify version --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := GetBuildInfo()
			format, _ := cmd.Flags().GetString("output")
			if format == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}
			cmd.Printf("clockify %s\n", info.Version)
			cmd.Printf("commit:     %s\n", info.Commit)
			cmd.Printf("built:      %s\n", info.Date)
			cmd.Printf("go:         %s\n", info.Go)
			cmd.Printf("platform:   %s\n", info.Platform)
			return nil
		},
	}
}
