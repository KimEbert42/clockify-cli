package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/internal/config"
	"github.com/kim/clockify-cli/internal/output"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage workspaces",
		Long:  `List and get Clockify workspaces.`,
	}
	cmd.AddCommand(newWorkspaceListCmd())
	cmd.AddCommand(newWorkspaceGetCmd())
	return cmd
}

func newWorkspaceListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")

			var cached []clockify.Workspace
			if cachedList(cmd, "", compResourceWorkspaces, &cached) && len(cached) > 0 {
				return formatOutput(cmd, cached, outputFormat, []string{"ID", "NAME", "CURRENCY"})
			}

			apiKey, region, err := loadAuth(profile)
			if err != nil {
				return err
			}

			client := api.NewClient(apiKey, api.BaseURL(region), "")
			workspaces, err := client.ListWorkspaces(context.Background())
			if err != nil {
				return err
			}

			rememberCompletionNames(cmd, "", compResourceWorkspaces, namesFromWorkspaces(workspaces))
			rememberList(cmd, "", compResourceWorkspaces, workspaces)

			if len(workspaces) == 0 && !quiet {
				cmd.Println("No workspaces found.")
				return nil
			}

			return formatOutput(cmd, workspaces, outputFormat, []string{"ID", "NAME", "CURRENCY"})
		},
	}
	_ = cmd.RegisterFlagCompletionFunc("name", completeWorkspaceFlagC)
	return cmd
}

func newWorkspaceGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get workspace by ID, name, or current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")

			apiKey, region, err := loadAuth(profile)
			if err != nil {
				return err
			}

			client := api.NewClient(apiKey, api.BaseURL(region), "")
			ctx := context.Background()

			if id == "" && name == "" {
				cfg, err := config.Load(profile)
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				if cfg.WorkspaceID == "" {
					return fmt.Errorf("no workspace specified; use --id, --name, or configure a default workspace")
				}
				id = cfg.WorkspaceID
			}

			if id == "" && name != "" {
				ws, err := resolveWorkspaceByName(ctx, client, name)
				if err != nil {
					return err
				}
				return formatOutput(cmd, ws, outputFormat, []string{"ID", "NAME", "CURRENCY"})
			}

			ws, err := client.GetWorkspace(ctx, id)
			if err != nil {
				return err
			}

			return formatOutput(cmd, ws, outputFormat, []string{"ID", "NAME", "CURRENCY"})
		},
	}
	cmd.Flags().String("id", "", "workspace ID")
	cmd.Flags().String("name", "", "workspace name (fuzzy match)")
	_ = cmd.RegisterFlagCompletionFunc("id", completeWorkspaceFlagC)
	_ = cmd.RegisterFlagCompletionFunc("name", completeWorkspaceFlagC)
	return cmd
}

func resolveWorkspaceByName(ctx context.Context, client *api.ClockifyClient, name string) (*clockify.Workspace, error) {
	workspaces, err := client.ListWorkspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	for _, ws := range workspaces {
		if ws.Name == name {
			return &ws, nil
		}
	}

	var fuzzyMatches []clockify.Workspace
	for _, ws := range workspaces {
		if containsIgnoreCase(ws.Name, name) {
			fuzzyMatches = append(fuzzyMatches, ws)
		}
	}

	if len(fuzzyMatches) == 1 {
		return &fuzzyMatches[0], nil
	}
	if len(fuzzyMatches) > 1 {
		return nil, fmt.Errorf("multiple workspaces match %q; use --id to disambiguate", name)
	}

	return nil, fmt.Errorf("no workspace found matching %q", name)
}

func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return len(s) >= len(substr) && strings.Contains(s, substr)
}

func loadAuth(profile string) (apiKey, region string, err error) {
	apiKey, err = config.GetAPIKey(profile)
	if err != nil {
		return "", "", fmt.Errorf("failed to get API key: %w", err)
	}

	cfg, err := config.Load(profile)
	if err != nil {
		return "", "", fmt.Errorf("failed to load config: %w", err)
	}

	return apiKey, cfg.Region, nil
}

func formatOutput(cmd *cobra.Command, data any, format string, headers []string) error {
	out := cmd.OutOrStdout()
	formatter := output.Detect(format, output.IsTerminal())
	if tf, ok := formatter.(*output.TableFormatter); ok {
		tf.Quiet = quietEnabled(cmd)
	}
	return formatter.Format(out, data, headers)
}

func quietEnabled(cmd *cobra.Command) bool {
	if cmd.Flags().Lookup("quiet") == nil {
		return false
	}
	v, _ := cmd.Flags().GetBool("quiet")
	return v
}

func colorDisabled(cmd *cobra.Command) bool {
	if cmd.Flags().Lookup("no-color") != nil {
		if v, _ := cmd.Flags().GetBool("no-color"); v {
			return true
		}
	}
	if v, ok := os.LookupEnv("NO_COLOR"); ok && v != "" {
		return true
	}
	return false
}
