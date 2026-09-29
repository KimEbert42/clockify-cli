package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
		Long:  `List and get Clockify projects.`,
	}
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectGetCmd())
	return cmd
}

func newProjectListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects in current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			statusFilter, _ := cmd.Flags().GetString("status")
			clientFilter, _ := cmd.Flags().GetString("client")
			searchFilter, _ := cmd.Flags().GetString("search")
			page, _ := cmd.Flags().GetInt("page")
			pageSize, _ := cmd.Flags().GetInt("page-size")
			noPagination, _ := cmd.Flags().GetBool("no-pagination")

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			projects, pageInfo, err := fetchPaginatedProjects(context.Background(), client, api.ListProjectsOpts{
				Status:   statusFilter,
				ClientID: clientFilter,
				Search:   searchFilter,
			}, page, pageSize, noPagination)
			if err != nil {
				return err
			}

			if len(projects) == 0 && !quiet {
				cmd.Println("No projects found.")
				return nil
			}

			if err := formatOutput(cmd, projects, outputFormat, []string{"ID", "NAME", "CLIENT_ID", "BILLABLE", "STATUS"}); err != nil {
				return err
			}

			if !quiet && outputFormat != "json" && outputFormat != "csv" && outputFormat != "yaml" && pageInfo != nil && pageInfo.TotalPages > 1 {
				cmd.Printf("\nPage %d of %d (%d results)\n", pageInfo.Page, pageInfo.TotalPages, pageInfo.TotalResults)
			}
			return nil
		},
	}
	cmd.Flags().String("status", "", "filter by status: ACTIVE, ARCHIVED")
	cmd.Flags().String("client", "", "filter by client ID")
	cmd.Flags().String("search", "", "search by project name")
	return cmd
}

func newProjectGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get project by ID or name",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")

			if id == "" && name == "" {
				return fmt.Errorf("must specify --id or --name")
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			ctx := context.Background()

			if id != "" {
				proj, err := client.GetProject(ctx, id)
				if err != nil {
					return err
				}
				return formatOutput(cmd, proj, outputFormat, []string{"ID", "NAME", "CLIENT_ID", "BILLABLE", "STATUS", "COLOR"})
			}

			proj, err := resolveProjectByName(ctx, client, name)
			if err != nil {
				return err
			}
			return formatOutput(cmd, proj, outputFormat, []string{"ID", "NAME", "CLIENT_ID", "BILLABLE", "STATUS", "COLOR"})
		},
	}
	cmd.Flags().String("id", "", "project ID")
	cmd.Flags().String("name", "", "project name (fuzzy match)")
	return cmd
}

func resolveProjectByName(ctx context.Context, client *api.ClockifyClient, name string) (*clockify.Project, error) {
	projects, err := client.ListProjects(ctx, api.ListProjectsOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, p := range projects {
		if p.Name == name {
			return &p, nil
		}
	}

	var fuzzyMatches []clockify.Project
	for _, p := range projects {
		if containsIgnoreCase(p.Name, name) {
			fuzzyMatches = append(fuzzyMatches, p)
		}
	}

	if len(fuzzyMatches) == 1 {
		return &fuzzyMatches[0], nil
	}
	if len(fuzzyMatches) > 1 {
		return nil, fmt.Errorf("multiple projects match %q; use --id to disambiguate", name)
	}

	return nil, fmt.Errorf("no project found matching %q", name)
}

func projectHasConflict(projects []clockify.Project, name string) bool {
	for _, p := range projects {
		if strings.EqualFold(p.Name, name) {
			return true
		}
	}
	return false
}
