package cli

import (
	"context"

	"github.com/kim/clockify-cli/internal/config"
	"github.com/kim/clockify-cli/internal/update"
	"github.com/spf13/cobra"
)

func init() {
	update.SetCurrentVersionProvider(func() string { return buildVersion })
}

// maybeCheckForUpdate runs the passive GitHub Releases check. Errors are swallowed.
func maybeCheckForUpdate(cmd *cobra.Command) {
	profile := ""
	if f := cmd.Flags().Lookup("profile"); f != nil {
		profile, _ = cmd.Flags().GetString("profile")
	}
	quiet := quietEnabled(cmd)
	outputFormat := ""
	if f := cmd.Flags().Lookup("output"); f != nil {
		outputFormat, _ = cmd.Flags().GetString("output")
	}

	cfg, err := config.Load(profile)
	autoUpdate := true
	if err == nil && cfg != nil {
		autoUpdate = cfg.AutoUpdateEnabled()
	}

	res := update.MaybeCheck(
		context.Background(),
		update.NewClient(),
		profile,
		buildVersion,
		quiet,
		outputFormat,
		autoUpdate,
		cmd.Name(),
	)
	if res == nil {
		return
	}
	// Hint goes to stderr so JSON pipelines stay clean.
	cmd.PrintErrln(update.Hint(res))
}

func newUpdateCmd() *cobra.Command {
	var (
		checkOnly bool
		version   string
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update clockify to a newer release",
		Long: `Download and install the latest clockify release from GitHub.

Examples:
  clockify update
  clockify update --check
  clockify update --version 1.2.3

Self-update replaces the running binary. Dev builds (version=dev) cannot self-update;
use your package manager or go install instead.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			current := buildVersion
			if update.IsUnsupportedInstall(current) {
				return newUpdateUnsupportedError(current)
			}

			client := update.NewClient()

			if checkOnly {
				ctx := context.Background()
				var res *update.CheckResult
				var err error
				if version != "" {
					rel, lerr := client.ReleaseByVersion(ctx, version)
					if lerr != nil {
						return lerr
					}
					res = &update.CheckResult{
						Current:         update.NormalizeVersion(current),
						Latest:          update.NormalizeVersion(rel.TagName),
						UpdateAvailable: update.CompareVersions(current, rel.TagName) < 0,
					}
				} else {
					res, err = update.Check(ctx, client, current)
					if err != nil {
						return err
					}
				}
				format, _ := cmd.Flags().GetString("output")
				if format == "json" {
					return formatOutput(cmd, res, "json", nil)
				}
				cmd.Printf("current: %s\n", res.Current)
				cmd.Printf("latest:  %s\n", res.Latest)
				if res.UpdateAvailable {
					cmd.Println("update available: yes")
				} else {
					cmd.Println("update available: no")
				}
				return nil
			}

			oldV, newV, err := update.Install(context.Background(), client, version, runtimeGoos(), runtimeGoarch())
			if err != nil {
				return err
			}
			quiet := quietEnabled(cmd)
			if !quiet {
				cmd.Printf("Updated clockify %s -> %s\n", oldV, newV)
				cmd.Println("Restart your shell if the binary is already loaded.")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check", false, "check for updates without installing")
	cmd.Flags().StringVar(&version, "version", "", "install a specific version (e.g. 1.2.3)")
	return cmd
}

func runtimeGoos() string { return update.RuntimeGoos() }
func runtimeGoarch() string { return update.RuntimeGoarch() }

func newUpdateUnsupportedError(current string) error {
	return &updateUnsupportedError{current: current}
}

type updateUnsupportedError struct {
	current string
}

func (e *updateUnsupportedError) Error() string {
	return "cannot self-update a development build (version=" + e.current + "); install a release binary or use go install / package manager"
}

func (e *updateUnsupportedError) ExitCode() int { return 10 }
