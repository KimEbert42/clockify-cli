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
	"github.com/kim/clockify-cli/pkg/duration"
	"github.com/spf13/cobra"
)

var validTaskStatuses = map[string]bool{
	"TODO":        true,
	"IN_PROGRESS": true,
	"DONE":        true,
}

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
		Long:  `List, create, update, and delete Clockify tasks within projects.`,
	}
	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskCreateCmd())
	cmd.AddCommand(newTaskUpdateCmd())
	cmd.AddCommand(newTaskDeleteCmd())
	return cmd
}

func newTaskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			projectRef, _ := cmd.Flags().GetString("project")
			isActive, _ := cmd.Flags().GetBool("is-active")
			page, _ := cmd.Flags().GetInt("page")
			pageSize, _ := cmd.Flags().GetInt("page-size")
			noPagination, _ := cmd.Flags().GetBool("no-pagination")

			if projectRef == "" {
				return fmt.Errorf("must specify --project")
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

			proj, err := resolveProjectRef(ctx, client, projectRef)
			if err != nil {
				return err
			}

			opts := api.ListTasksOpts{}
			if cmd.Flags().Changed("is-active") {
				opts.IsActive = &isActive
			}

			tasks, pageInfo, err := fetchPaginatedTasks(ctx, client, proj.ID, opts, page, pageSize, noPagination)
			if err != nil {
				return err
			}

			if len(tasks) == 0 && !quiet {
				cmd.Println("No tasks found.")
				return nil
			}

			type taskDisplay struct {
				ID        string `json:"id" yaml:"id"`
				Name      string `json:"name" yaml:"name"`
				Status    string `json:"status" yaml:"status"`
				Estimate  string `json:"estimate" yaml:"estimate"`
				Assignee  string `json:"assignee" yaml:"assignee"`
				Billable  bool   `json:"billable" yaml:"billable"`
				Active    bool   `json:"active" yaml:"active"`
				ProjectID string `json:"projectId" yaml:"projectId"`
			}

			var displays []taskDisplay
			for _, t := range tasks {
				est := ""
				if t.Estimate != nil && t.Estimate.Estimate > 0 {
					est = duration.Format(t.Estimate.Estimate)
				}
				assignee := ""
				if len(t.AssigneeIDs) > 0 {
					assignee = strings.Join(t.AssigneeIDs, ",")
				}
				displays = append(displays, taskDisplay{
					ID:        t.ID,
					Name:      t.Name,
					Status:    t.Status,
					Estimate:  est,
					Assignee:  assignee,
					Billable:  t.IsBillable,
					Active:    t.Active,
					ProjectID: t.ProjectID,
				})
			}

			if err := formatOutput(cmd, displays, outputFormat, []string{"ID", "NAME", "STATUS", "ESTIMATE", "ASSIGNEE", "BILLABLE"}); err != nil {
				return err
			}

			if !quiet && outputFormat != "json" && outputFormat != "csv" && outputFormat != "yaml" && pageInfo != nil && pageInfo.TotalPages > 1 {
				cmd.Printf("\nPage %d of %d (%d results)\n", pageInfo.Page, pageInfo.TotalPages, pageInfo.TotalResults)
			}
			return nil
		},
	}
	cmd.Flags().String("project", "", "project ID or name (required)")
	cmd.Flags().Bool("is-active", false, "show only active tasks")
	return cmd
}

func newTaskCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task in a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			projectRef, _ := cmd.Flags().GetString("project")
			name, _ := cmd.Flags().GetString("name")
			estimate, _ := cmd.Flags().GetString("estimate")
			status, _ := cmd.Flags().GetString("status")
			assignee, _ := cmd.Flags().GetString("assignee")
			billable, _ := cmd.Flags().GetBool("billable")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")
			ifNotExists, _ := cmd.Flags().GetBool("if-not-exists")

			if projectRef == "" {
				return fmt.Errorf("must specify --project")
			}

			// Handle JSON input from file or stdin before auth/API calls
			if filePath != "" || stdin {
				var inputData []byte
				var err error
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

				var jsonReq clockify.CreateTaskRequest
				if err := json.Unmarshal(inputData, &jsonReq); err != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				if jsonReq.Name == "" {
					return fmt.Errorf("task name is required")
				}
				name = jsonReq.Name
				status = jsonReq.Status
				billable = jsonReq.IsBillable
				if jsonReq.Estimate != nil && jsonReq.Estimate.Estimate > 0 {
					estimate = fmt.Sprintf("%d", jsonReq.Estimate.Estimate)
				}
				if len(jsonReq.AssigneeIDs) > 0 {
					assignee = jsonReq.AssigneeIDs[0]
				}
			}

			if name == "" {
				return fmt.Errorf("must specify --name")
			}

			if status != "" {
				var err error
				status, err = parseTaskStatus(status)
				if err != nil {
					return err
				}
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

			proj, err := resolveProjectRef(ctx, client, projectRef)
			if err != nil {
				return err
			}

			var assigneeIDs []string
			if assignee != "" {
				uid, err := resolveUserRef(ctx, client, assignee)
				if err != nil {
					return err
				}
				assigneeIDs = []string{uid}
			}

			// Handle if-not-exists (scoped to project)
			if ifNotExists {
				existing, err := resolveTaskByName(ctx, client, proj.ID, name)
				if err == nil {
					cmd.Printf("Task %s already exists (ID: %s).\n", existing.Name, existing.ID)
					return formatOutput(cmd, existing, outputFormat, []string{"ID", "NAME", "STATUS", "ESTIMATE", "ASSIGNEE", "BILLABLE"})
				}
			}

			req := clockify.CreateTaskRequest{
				Name:        name,
				Status:      status,
				AssigneeIDs: assigneeIDs,
				IsBillable:  billable,
			}

			if estimate != "" {
				secs, err := duration.Parse(estimate)
				if err != nil {
					return fmt.Errorf("invalid estimate: %w", err)
				}
				req.Estimate = &clockify.Estimate{
					Type:     "AUTO",
					Estimate: secs,
				}
			}

			task, err := client.CreateTask(ctx, proj.ID, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, task, outputFormat, []string{"ID", "NAME", "STATUS", "ESTIMATE", "ASSIGNEE", "BILLABLE"})
		},
	}
	cmd.Flags().String("project", "", "project ID or name (required)")
	cmd.Flags().String("name", "", "task name (required)")
	cmd.Flags().String("estimate", "", "task estimate (seconds, or e.g. 1h, 1.5h, 30m)")
	cmd.Flags().String("status", "", "task status: TODO, IN_PROGRESS, DONE")
	cmd.Flags().String("assignee", "", "assignee user ID, email, or name")
	cmd.Flags().Bool("billable", false, "task is billable")
	cmd.Flags().String("file", "", "JSON file with task data")
	cmd.Flags().Bool("stdin", false, "read task data from stdin")
	cmd.Flags().Bool("if-not-exists", false, "return existing task if one with the same name exists in the project")
	return cmd
}

func newTaskUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			projectRef, _ := cmd.Flags().GetString("project")
			id, _ := cmd.Flags().GetString("id")
			newName, _ := cmd.Flags().GetString("name")
			estimate, _ := cmd.Flags().GetString("estimate")
			status, _ := cmd.Flags().GetString("status")
			assignee, _ := cmd.Flags().GetString("assignee")
			billable, _ := cmd.Flags().GetBool("billable")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if projectRef == "" {
				return fmt.Errorf("must specify --project")
			}
			if id == "" {
				return fmt.Errorf("must specify --id")
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

			proj, err := resolveProjectRef(ctx, client, projectRef)
			if err != nil {
				return err
			}

			// Get existing task to merge updates
			existing, err := client.GetTask(ctx, proj.ID, id)
			if err != nil {
				return err
			}

			req := clockify.UpdateTaskRequest{
				Name:        existing.Name,
				Estimate:    existing.Estimate,
				Status:      existing.Status,
				AssigneeIDs: existing.AssigneeIDs,
				IsBillable:  &existing.IsBillable,
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

				var jsonReq clockify.UpdateTaskRequest
				if err := json.Unmarshal(inputData, &jsonReq); err != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				if jsonReq.Name != "" {
					req.Name = jsonReq.Name
				}
				if jsonReq.Estimate != nil {
					req.Estimate = jsonReq.Estimate
				}
				if jsonReq.Status != "" {
					status, err = parseTaskStatus(jsonReq.Status)
					if err != nil {
						return err
					}
					req.Status = status
				}
				if jsonReq.AssigneeIDs != nil {
					req.AssigneeIDs = jsonReq.AssigneeIDs
				}
				if jsonReq.IsBillable != nil {
					req.IsBillable = jsonReq.IsBillable
				}
			} else {
				// Apply flag updates
				if newName != "" {
					req.Name = newName
				}
				if estimate != "" {
					secs, err := duration.Parse(estimate)
					if err != nil {
						return fmt.Errorf("invalid estimate: %w", err)
					}
					req.Estimate = &clockify.Estimate{
						Type:     "AUTO",
						Estimate: secs,
					}
				}
				if status != "" {
					status, err = parseTaskStatus(status)
					if err != nil {
						return err
					}
					req.Status = status
				}
				if assignee != "" {
					uid, err := resolveUserRef(ctx, client, assignee)
					if err != nil {
						return err
					}
					req.AssigneeIDs = []string{uid}
				}
				if cmd.Flags().Changed("billable") {
					req.IsBillable = &billable
				}
			}

			task, err := client.UpdateTask(ctx, proj.ID, id, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, task, outputFormat, []string{"ID", "NAME", "STATUS", "ESTIMATE", "ASSIGNEE", "BILLABLE"})
		},
	}
	cmd.Flags().String("project", "", "project ID or name (required)")
	cmd.Flags().String("id", "", "task ID (required)")
	cmd.Flags().String("name", "", "new task name")
	cmd.Flags().String("estimate", "", "task estimate (seconds, or e.g. 1h, 1.5h, 30m)")
	cmd.Flags().String("status", "", "task status: TODO, IN_PROGRESS, DONE")
	cmd.Flags().String("assignee", "", "assignee user ID, email, or name")
	cmd.Flags().Bool("billable", false, "task is billable")
	cmd.Flags().String("file", "", "JSON file with task data")
	cmd.Flags().Bool("stdin", false, "read task data from stdin")
	return cmd
}

func newTaskDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			projectRef, _ := cmd.Flags().GetString("project")
			id, _ := cmd.Flags().GetString("id")
			confirm, _ := cmd.Flags().GetBool("confirm")

			if projectRef == "" {
				return fmt.Errorf("must specify --project")
			}
			if id == "" {
				return fmt.Errorf("must specify --id")
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

			proj, err := resolveProjectRef(ctx, client, projectRef)
			if err != nil {
				return err
			}

			// Get task info for confirmation message
			task, err := client.GetTask(ctx, proj.ID, id)
			if err != nil {
				return err
			}

			if !confirm {
				cmd.Printf("Delete task %s (%s)?\n", task.ID, task.Name)
				cmd.Print("Are you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					cmd.Println("Aborted.")
					return nil
				}
			}

			if err := client.DeleteTask(ctx, proj.ID, id); err != nil {
				return err
			}

			cmd.Printf("Task %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("project", "", "project ID or name (required)")
	cmd.Flags().String("id", "", "task ID (required)")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}

func parseTaskStatus(s string) (string, error) {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), "-", "_"))
	if validTaskStatuses[normalized] {
		return normalized, nil
	}
	return "", fmt.Errorf("invalid status %q; must be one of: TODO, IN_PROGRESS, DONE", s)
}

func resolveProjectRef(ctx context.Context, client *api.ClockifyClient, ref string) (*clockify.Project, error) {
	// Try as project ID first (fast path — spec examples use IDs)
	proj, err := client.GetProject(ctx, ref)
	if err == nil {
		return proj, nil
	}
	// Fall back to name resolution
	proj, err = resolveProjectByName(ctx, client, ref)
	if err != nil {
		return nil, fmt.Errorf("no project found matching %q", ref)
	}
	return proj, nil
}

func resolveTaskByName(ctx context.Context, client *api.ClockifyClient, projectID, name string) (*clockify.Task, error) {
	tasks, err := client.ListTasks(ctx, projectID, api.ListTasksOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, t := range tasks {
		if t.Name == name {
			return &t, nil
		}
	}

	var fuzzyMatches []clockify.Task
	for _, t := range tasks {
		if containsIgnoreCase(t.Name, name) {
			fuzzyMatches = append(fuzzyMatches, t)
		}
	}

	if len(fuzzyMatches) == 1 {
		return &fuzzyMatches[0], nil
	}
	if len(fuzzyMatches) > 1 {
		return nil, fmt.Errorf("multiple tasks match %q; use --id to disambiguate", name)
	}

	return nil, fmt.Errorf("no task found matching %q", name)
}

func resolveUserRef(ctx context.Context, client *api.ClockifyClient, ref string) (string, error) {
	users, err := client.ListUsers(ctx, api.ListUsersOpts{})
	if err != nil {
		return "", fmt.Errorf("failed to list users: %w", err)
	}

	// Exact ID match
	for _, u := range users {
		if u.ID == ref {
			return u.ID, nil
		}
	}
	// Exact email match
	for _, u := range users {
		if strings.EqualFold(u.Email, ref) {
			return u.ID, nil
		}
	}
	// Exact name match
	for _, u := range users {
		if u.Name == ref {
			return u.ID, nil
		}
	}
	// Fuzzy name match
	var matches []clockify.User
	for _, u := range users {
		if containsIgnoreCase(u.Name, ref) {
			matches = append(matches, u)
		}
	}
	if len(matches) == 1 {
		return matches[0].ID, nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple users match %q; use user ID to disambiguate", ref)
	}

	// Not found by name/email — treat as raw user ID
	return ref, nil
}
