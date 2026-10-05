package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
		Long:  `List, get, create, update, and delete Clockify projects.`,
	}
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectGetCmd())
	cmd.AddCommand(newProjectCreateCmd())
	cmd.AddCommand(newProjectUpdateCmd())
	cmd.AddCommand(newProjectDeleteCmd())
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

			unfiltered := statusFilter == "" && clientFilter == "" && searchFilter == ""
			wsID := workspaceIDFromCmd(cmd)
			if unfiltered && wsID != "" {
				var cached []clockify.Project
				if cachedList(cmd, wsID, compResourceProjects, &cached) && len(cached) > 0 {
					return formatOutput(cmd, cached, outputFormat, []string{"ID", "NAME", "CLIENT_ID", "BILLABLE", "STATUS"})
				}
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}
			wsID = cfg.WorkspaceID

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			projects, pageInfo, err := fetchPaginatedProjects(context.Background(), client, api.ListProjectsOpts{
				Status:   statusFilter,
				ClientID: clientFilter,
				Search:   searchFilter,
			}, page, pageSize, noPagination)
			if err != nil {
				return err
			}

			if unfiltered {
				rememberCompletionNames(cmd, wsID, compResourceProjects, namesFromProjects(projects))
				rememberList(cmd, wsID, compResourceProjects, projects)
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
	_ = cmd.RegisterFlagCompletionFunc("client", completeClientRefC)
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
	_ = cmd.RegisterFlagCompletionFunc("id", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("name", completeProjectRefC)
	return cmd
}

func newProjectCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			name, _ := cmd.Flags().GetString("name")
			clientID, _ := cmd.Flags().GetString("client")
			billable, _ := cmd.Flags().GetBool("billable")
			color, _ := cmd.Flags().GetString("color")
			estimate, _ := cmd.Flags().GetString("estimate")
			note, _ := cmd.Flags().GetString("note")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")
			ifNotExists, _ := cmd.Flags().GetBool("if-not-exists")

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			ctx := context.Background()

			// Handle JSON input from file or stdin
			if filePath != "" || stdin {
				var inputData []byte
				if filePath != "" {
					inputData, err = os.ReadFile(filePath)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", filePath, err)
					}
				} else {
					inputData, err = io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read from stdin: %w", err)
					}
				}

				var req clockify.CreateProjectRequest
				if err := json.Unmarshal(inputData, &req); err != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				if req.Name == "" {
					return fmt.Errorf("project name is required")
				}
				name = req.Name
				clientID = req.ClientID
				billable = req.Billable
				color = req.Color
				note = req.Note
			}

			if name == "" {
				return fmt.Errorf("must specify --name")
			}

			// Handle if-not-exists
			if ifNotExists {
				existing, err := resolveProjectByName(ctx, client, name)
				if err == nil {
					cmd.Printf("Project %s already exists (ID: %s).\n", existing.Name, existing.ID)
					return formatOutput(cmd, existing, outputFormat, []string{"ID", "NAME", "CLIENT_ID", "BILLABLE", "STATUS", "COLOR"})
				}
			}

			req := clockify.CreateProjectRequest{
				Name:     name,
				ClientID: clientID,
				Billable: billable,
				Color:    color,
				Note:     note,
			}

			if estimate != "" {
				// Parse estimate duration
				secs, err := parseEstimateDuration(estimate)
				if err != nil {
					return fmt.Errorf("invalid estimate: %w", err)
				}
				req.Estimate = &clockify.Estimate{
					Type:     "AUTO",
					Estimate: secs,
				}
			}

			proj, err := client.CreateProject(ctx, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, proj, outputFormat, []string{"ID", "NAME", "CLIENT_ID", "BILLABLE", "STATUS", "COLOR"})
		},
	}
	cmd.Flags().String("name", "", "project name (required)")
	cmd.Flags().String("client", "", "client ID")
	cmd.Flags().Bool("billable", false, "project is billable")
	cmd.Flags().String("color", "", "project color (hex format: #RRGGBB)")
	cmd.Flags().String("estimate", "", "project estimate (e.g., 10h, 2d)")
	cmd.Flags().String("note", "", "project note")
	cmd.Flags().String("file", "", "JSON file with project data")
	cmd.Flags().Bool("stdin", false, "read project data from stdin")
	cmd.Flags().Bool("if-not-exists", false, "return existing project if one with the same name exists")
	_ = cmd.RegisterFlagCompletionFunc("client", completeClientRefC)
	return cmd
}

func newProjectUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			newName, _ := cmd.Flags().GetString("new-name")
			clientID, _ := cmd.Flags().GetString("client")
			billable, _ := cmd.Flags().GetBool("billable")
			color, _ := cmd.Flags().GetString("color")
			note, _ := cmd.Flags().GetString("note")
			archive, _ := cmd.Flags().GetBool("archive")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

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

			// Resolve project ID
			if id == "" {
				proj, err := resolveProjectByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = proj.ID
			}

			// Get existing project to merge updates
			existing, err := client.GetProject(ctx, id)
			if err != nil {
				return err
			}

			req := clockify.UpdateProjectRequest{
				Name:     existing.Name,
				ClientID: existing.ClientID,
				Billable: &existing.Billable,
				Color:    existing.Color,
				Note:     existing.Note,
			}

			// Handle JSON input from file or stdin
			if filePath != "" || stdin {
				var inputData []byte
				if filePath != "" {
					inputData, err = os.ReadFile(filePath)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", filePath, err)
					}
				} else {
					inputData, err = io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read from stdin: %w", err)
					}
				}

				var jsonReq clockify.UpdateProjectRequest
				if err := json.Unmarshal(inputData, &jsonReq); err != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				// Merge JSON input with existing
				if jsonReq.Name != "" {
					req.Name = jsonReq.Name
				}
				if jsonReq.ClientID != "" {
					req.ClientID = jsonReq.ClientID
				}
				if jsonReq.Billable != nil {
					req.Billable = jsonReq.Billable
				}
				if jsonReq.Color != "" {
					req.Color = jsonReq.Color
				}
				if jsonReq.Note != "" {
					req.Note = jsonReq.Note
				}
				if jsonReq.Archived != nil {
					req.Archived = jsonReq.Archived
				}
			} else {
				// Apply flag updates
				if newName != "" {
					req.Name = newName
				}
				if clientID != "" {
					req.ClientID = clientID
				}
				if cmd.Flags().Changed("billable") {
					req.Billable = &billable
				}
				if color != "" {
					req.Color = color
				}
				if note != "" {
					req.Note = note
				}
				if archive {
					req.Archived = &archive
				}
			}

			proj, err := client.UpdateProject(ctx, id, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, proj, outputFormat, []string{"ID", "NAME", "CLIENT_ID", "BILLABLE", "STATUS", "COLOR"})
		},
	}
	cmd.Flags().String("id", "", "project ID")
	cmd.Flags().String("name", "", "project name (fuzzy match)")
	cmd.Flags().String("new-name", "", "new project name")
	cmd.Flags().String("client", "", "client ID")
	cmd.Flags().Bool("billable", false, "project is billable")
	cmd.Flags().String("color", "", "project color (hex format: #RRGGBB)")
	cmd.Flags().String("note", "", "project note")
	cmd.Flags().Bool("archive", false, "archive the project")
	cmd.Flags().String("file", "", "JSON file with project data")
	cmd.Flags().Bool("stdin", false, "read project data from stdin")
	_ = cmd.RegisterFlagCompletionFunc("id", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("name", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("new-name", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("client", completeClientRefC)
	return cmd
}

func newProjectDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			confirm, _ := cmd.Flags().GetBool("confirm")

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

			// Resolve project ID if needed
			if id == "" {
				proj, err := resolveProjectByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = proj.ID
				name = proj.Name
			} else {
				// Get project info for confirmation message
				proj, err := client.GetProject(ctx, id)
				if err != nil {
					return err
				}
				name = proj.Name
			}

			if !confirm {
				cmd.Printf("Delete project %s (%s)?\n", id, name)
				cmd.Print("Are you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					cmd.Println("Aborted.")
					return nil
				}
			}

			if err := client.DeleteProject(ctx, id); err != nil {
				return err
			}

			cmd.Printf("Project %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "project ID")
	cmd.Flags().String("name", "", "project name (fuzzy match)")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	_ = cmd.RegisterFlagCompletionFunc("id", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("name", completeProjectRefC)
	return cmd
}

func parseEstimateDuration(estimate string) (int64, error) {
	// Simple estimate parsing - supports hours (h) and days (d)
	estimate = strings.TrimSpace(estimate)
	if estimate == "" {
		return 0, fmt.Errorf("empty estimate")
	}

	// Check for days
	if strings.HasSuffix(estimate, "d") {
		var days float64
		if _, err := fmt.Sscanf(estimate, "%fd", &days); err != nil {
			return 0, fmt.Errorf("invalid estimate format: %s", estimate)
		}
		return int64(days * 8 * 3600), nil // 8 hours per day
	}

	// Check for hours
	if strings.HasSuffix(estimate, "h") {
		var hours float64
		if _, err := fmt.Sscanf(estimate, "%fh", &hours); err != nil {
			return 0, fmt.Errorf("invalid estimate format: %s", estimate)
		}
		return int64(hours * 3600), nil
	}

	// Assume minutes if no suffix
	var minutes float64
	if _, err := fmt.Sscanf(estimate, "%fm", &minutes); err == nil {
		return int64(minutes * 60), nil
	}

	return 0, fmt.Errorf("invalid estimate format: %s (use h, d, or m suffix)", estimate)
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
		cands := make([]nameCandidate, 0, len(fuzzyMatches))
		for _, p := range fuzzyMatches {
			cands = append(cands, nameCandidate{Name: p.Name, ID: p.ID})
		}
		return nil, ambiguousRefError("projects", name, cands)
	}

	return nil, notFoundRefError("project", name)
}

func projectHasConflict(projects []clockify.Project, name string) bool {
	for _, p := range projects {
		if strings.EqualFold(p.Name, name) {
			return true
		}
	}
	return false
}
