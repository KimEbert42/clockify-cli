package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/kim/clockify-cli/pkg/duration"
	"github.com/spf13/cobra"
)

func newTimeEntryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "time-entry",
		Aliases: []string{"te"},
		Short:   "Manage time entries",
		Long:    `List, get, create, update, and delete Clockify time entries.`,
	}
	cmd.AddCommand(newTimeEntryListCmd())
	cmd.AddCommand(newTimeEntryGetCmd())
	cmd.AddCommand(newTimeEntryCreateCmd())
	cmd.AddCommand(newTimeEntryUpdateCmd())
	cmd.AddCommand(newTimeEntryDeleteCmd())
	cmd.AddCommand(newTimeEntryStartCmd())
	cmd.AddCommand(newTimeEntryStopCmd())
	cmd.AddCommand(newTimeEntryBulkCmd())
	return cmd
}

func newTimeEntryListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			startFilter, _ := cmd.Flags().GetString("start")
			endFilter, _ := cmd.Flags().GetString("end")
			projectFilter, _ := cmd.Flags().GetString("project")
			taskFilter, _ := cmd.Flags().GetString("task")
			tagsFilter, _ := cmd.Flags().GetString("tags")
			descriptionFilter, _ := cmd.Flags().GetString("description")
			inProgress, _ := cmd.Flags().GetBool("in-progress")
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

			user, err := client.GetCurrentUser(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			// Parse relative dates
			if startFilter != "" && duration.IsRelativeDate(startFilter) {
				t, err := duration.ParseRelativeDate(startFilter, "monday")
				if err != nil {
					return fmt.Errorf("invalid start date: %w", err)
				}
				startFilter = t.Format(time.RFC3339)
			}
			if endFilter != "" && duration.IsRelativeDate(endFilter) {
				t, err := duration.ParseRelativeDate(endFilter, "monday")
				if err != nil {
					return fmt.Errorf("invalid end date: %w", err)
				}
				// For end dates, use end of day
				endFilter = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()).Format(time.RFC3339)
			}

			// Default to today if no date filters provided
			if startFilter == "" && endFilter == "" && !inProgress {
				today := time.Now()
				startFilter = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()).Format(time.RFC3339)
				endFilter = time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, today.Location()).Format(time.RFC3339)
			}

			if inProgress {
				entries, err := client.GetTimeEntryInProgress(context.Background(), user.ID)
				if err != nil {
					return err
				}

				if len(entries) == 0 && !quiet {
					cmd.Println("No running timers.")
					return nil
				}

				return formatOutput(cmd, entries, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "START", "BILLABLE"})
			}

			hydrated, _ := cmd.Flags().GetBool("hydrated")

			opts := api.ListTimeEntriesOpts{
				Start:       startFilter,
				End:         endFilter,
				ProjectID:   projectFilter,
				TaskID:      taskFilter,
				Tags:        tagsFilter,
				Description: descriptionFilter,
				Hydrated:    hydrated,
			}

			entries, pageInfo, err := fetchPaginatedTimeEntries(context.Background(), client, user.ID, opts, page, pageSize, noPagination)
			if err != nil {
				return err
			}

			if len(entries) == 0 && !quiet {
				cmd.Println("No time entries found.")
				return nil
			}

			// Format entries for display
			type entryDisplay struct {
				ID          string `json:"id" yaml:"id"`
				Description string `json:"description" yaml:"description"`
				Project     string `json:"project" yaml:"project"`
				Task        string `json:"task" yaml:"task"`
				Duration    string `json:"duration" yaml:"duration"`
				Start       string `json:"start" yaml:"start"`
				End         string `json:"end" yaml:"end"`
				Billable    bool   `json:"billable" yaml:"billable"`
			}

			var displays []entryDisplay
			for _, e := range entries {
				dur := ""
				if e.TimeInterval.Duration > 0 {
					dur = duration.Format(e.TimeInterval.Duration)
				} else if e.TimeInterval.End == nil {
					dur = "running"
				}

				start := ""
				if !e.TimeInterval.Start.IsZero() {
					start = e.TimeInterval.Start.Format("2006-01-02 15:04")
				}

				end := ""
				if e.TimeInterval.End != nil && !e.TimeInterval.End.IsZero() {
					end = e.TimeInterval.End.Format("2006-01-02 15:04")
				}

				displays = append(displays, entryDisplay{
					ID:          e.ID,
					Description: e.Description,
					Project:     e.ProjectID,
					Task:        e.TaskID,
					Duration:    dur,
					Start:       start,
					End:         end,
					Billable:    e.Billable,
				})
			}

			if err := formatOutput(cmd, displays, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "TASK", "DURATION", "START", "END", "BILLABLE"}); err != nil {
				return err
			}

			if !quiet && outputFormat != "json" && outputFormat != "csv" && outputFormat != "yaml" && pageInfo != nil && pageInfo.TotalPages > 1 {
				cmd.Printf("\nPage %d of %d (%d results)\n", pageInfo.Page, pageInfo.TotalPages, pageInfo.TotalResults)
			}
			return nil
		},
	}
	cmd.Flags().String("start", "", "start date (YYYY-MM-DD or relative: today, yesterday, this-week, last-week, this-month, last-month)")
	cmd.Flags().String("end", "", "end date (YYYY-MM-DD or relative)")
	cmd.Flags().String("project", "", "filter by project ID")
	cmd.Flags().String("task", "", "filter by task ID")
	cmd.Flags().String("tags", "", "filter by tag IDs (comma-separated)")
	cmd.Flags().String("description", "", "search by description")
	cmd.Flags().Bool("in-progress", false, "show only running timers")
	cmd.Flags().Bool("hydrated", false, "include full project/task/user objects in output")
	return cmd
}

func newTimeEntryGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get time entry by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			hydrated, _ := cmd.Flags().GetBool("hydrated")

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

			var entry *clockify.TimeEntry
			if hydrated {
				// Use batch endpoint with hydrated flag
				entries, err := client.BatchGetTimeEntries(context.Background(), []string{id}, true)
				if err != nil {
					return err
				}
				if len(entries) == 0 {
					return fmt.Errorf("time entry %s not found", id)
				}
				entry = &entries[0]
			} else {
				entry, err = client.GetTimeEntry(context.Background(), id)
				if err != nil {
					return err
				}
			}

			return formatOutput(cmd, entry, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "TASK", "USER", "DURATION", "START", "END", "BILLABLE", "TAGS"})
		},
	}
	cmd.Flags().String("id", "", "time entry ID")
	cmd.Flags().Bool("hydrated", false, "include full project/task/user objects in output")
	return cmd
}

func newTimeEntryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			startStr, _ := cmd.Flags().GetString("start")
			durationStr, _ := cmd.Flags().GetString("duration")
			endStr, _ := cmd.Flags().GetString("end")
			projectID, _ := cmd.Flags().GetString("project")
			taskID, _ := cmd.Flags().GetString("task")
			description, _ := cmd.Flags().GetString("description")
			tagsStr, _ := cmd.Flags().GetString("tags")
			billable, _ := cmd.Flags().GetBool("billable")

			if startStr == "" {
				return fmt.Errorf("must specify --start")
			}

			if durationStr == "" && endStr == "" {
				return fmt.Errorf("must specify --duration or --end")
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)

			// Parse start time
			startTime, err := time.Parse(time.RFC3339, startStr)
			if err != nil {
				return fmt.Errorf("invalid start time %q: use format 2006-01-02T15:04:05Z", startStr)
			}

			req := api.CreateTimeEntryRequest{
				Start:       startTime.Format(time.RFC3339),
				ProjectID:   projectID,
				TaskID:      taskID,
				Description: description,
				Billable:    billable,
			}

			// Parse tags
			if tagsStr != "" {
				req.TagIDs = strings.Split(tagsStr, ",")
			}

			// Handle duration or end time
			if durationStr != "" {
				secs, err := duration.Parse(durationStr)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
				endTime := startTime.Add(time.Duration(secs) * time.Second)
				req.End = endTime.Format(time.RFC3339)
			} else {
				endTime, err := time.Parse(time.RFC3339, endStr)
				if err != nil {
					return fmt.Errorf("invalid end time %q: use format 2006-01-02T15:04:05Z", endStr)
				}
				req.End = endTime.Format(time.RFC3339)
			}

			entry, err := client.CreateTimeEntry(context.Background(), req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, entry, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "DURATION", "START", "END", "BILLABLE"})
		},
	}
	cmd.Flags().String("start", "", "start time (RFC3339 format: 2006-01-02T15:04:05Z)")
	cmd.Flags().String("duration", "", "duration (e.g., 1h, 30m, 1h30m)")
	cmd.Flags().String("end", "", "end time (RFC3339 format)")
	cmd.Flags().String("project", "", "project ID")
	cmd.Flags().String("task", "", "task ID")
	cmd.Flags().String("description", "", "description")
	cmd.Flags().String("tags", "", "tag IDs (comma-separated)")
	cmd.Flags().Bool("billable", false, "mark as billable")
	return cmd
}

func newTimeEntryUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a time entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			startStr, _ := cmd.Flags().GetString("start")
			durationStr, _ := cmd.Flags().GetString("duration")
			endStr, _ := cmd.Flags().GetString("end")
			projectID, _ := cmd.Flags().GetString("project")
			taskID, _ := cmd.Flags().GetString("task")
			description, _ := cmd.Flags().GetString("description")
			tagsStr, _ := cmd.Flags().GetString("tags")
			billable, _ := cmd.Flags().GetBool("billable")

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

			// Get existing entry to merge updates
			existing, err := client.GetTimeEntry(context.Background(), id)
			if err != nil {
				return err
			}

			req := api.UpdateTimeEntryRequest{
				Start:       existing.TimeInterval.Start.Format(time.RFC3339),
				ProjectID:   existing.ProjectID,
				TaskID:      existing.TaskID,
				Description: existing.Description,
				Billable:    existing.Billable,
				TagIDs:      existing.TagIDs,
			}

			// Apply updates
			if startStr != "" {
				startTime, err := time.Parse(time.RFC3339, startStr)
				if err != nil {
					return fmt.Errorf("invalid start time %q: use format 2006-01-02T15:04:05Z", startStr)
				}
				req.Start = startTime.Format(time.RFC3339)
			}

			if projectID != "" {
				req.ProjectID = projectID
			}

			if taskID != "" {
				req.TaskID = taskID
			}

			if description != "" {
				req.Description = description
			}

			if tagsStr != "" {
				req.TagIDs = strings.Split(tagsStr, ",")
			}

			if cmd.Flags().Changed("billable") {
				req.Billable = billable
			}

			// Handle duration or end time
			if durationStr != "" {
				secs, err := duration.Parse(durationStr)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
				startTime, _ := time.Parse(time.RFC3339, req.Start)
				endTime := startTime.Add(time.Duration(secs) * time.Second)
				req.End = endTime.Format(time.RFC3339)
			} else if endStr != "" {
				endTime, err := time.Parse(time.RFC3339, endStr)
				if err != nil {
					return fmt.Errorf("invalid end time %q: use format 2006-01-02T15:04:05Z", endStr)
				}
				req.End = endTime.Format(time.RFC3339)
			} else if existing.TimeInterval.End != nil {
				req.End = existing.TimeInterval.End.Format(time.RFC3339)
			}

			entry, err := client.UpdateTimeEntry(context.Background(), id, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, entry, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "DURATION", "START", "END", "BILLABLE"})
		},
	}
	cmd.Flags().String("id", "", "time entry ID")
	cmd.Flags().String("start", "", "start time (RFC3339 format)")
	cmd.Flags().String("duration", "", "duration (e.g., 1h, 30m, 1h30m)")
	cmd.Flags().String("end", "", "end time (RFC3339 format)")
	cmd.Flags().String("project", "", "project ID")
	cmd.Flags().String("task", "", "task ID")
	cmd.Flags().String("description", "", "description")
	cmd.Flags().String("tags", "", "tag IDs (comma-separated)")
	cmd.Flags().Bool("billable", false, "mark as billable")
	return cmd
}

func newTimeEntryDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a time entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			confirm, _ := cmd.Flags().GetBool("confirm")

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

			// Get entry info before deleting
			entry, err := client.GetTimeEntry(context.Background(), id)
			if err != nil {
				return err
			}

			if !confirm {
				cmd.Printf("Delete time entry %s? (%s - %s)\n",
					entry.ID,
					entry.Description,
					entry.TimeInterval.Start.Format("2006-01-02 15:04"))
				cmd.Print("Are you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					cmd.Println("Aborted.")
					return nil
				}
			}

			if err := client.DeleteTimeEntry(context.Background(), id); err != nil {
				return err
			}

			cmd.Printf("Time entry %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "time entry ID")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}

func newTimeEntryStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a timer",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			projectID, _ := cmd.Flags().GetString("project")
			taskID, _ := cmd.Flags().GetString("task")
			description, _ := cmd.Flags().GetString("description")
			tagsStr, _ := cmd.Flags().GetString("tags")
			billable, _ := cmd.Flags().GetBool("billable")

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)

			// Check for running timer
			user, err := client.GetCurrentUser(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			running, err := client.GetTimeEntryInProgress(context.Background(), user.ID)
			if err != nil {
				return err
			}

			if len(running) > 0 {
				return fmt.Errorf("timer already running (ID: %s); stop it first with 'clockify time-entry stop'", running[0].ID)
			}

			req := api.CreateTimeEntryRequest{
				Start:       time.Now().Format(time.RFC3339),
				ProjectID:   projectID,
				TaskID:      taskID,
				Description: description,
				Billable:    billable,
			}

			if tagsStr != "" {
				req.TagIDs = strings.Split(tagsStr, ",")
			}

			entry, err := client.CreateTimeEntry(context.Background(), req)
			if err != nil {
				return err
			}

			cmd.Printf("Timer started (ID: %s)\n", entry.ID)
			return formatOutput(cmd, entry, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "START"})
		},
	}
	cmd.Flags().String("project", "", "project ID")
	cmd.Flags().String("task", "", "task ID")
	cmd.Flags().String("description", "", "description")
	cmd.Flags().String("tags", "", "tag IDs (comma-separated)")
	cmd.Flags().Bool("billable", false, "mark as billable")
	return cmd
}

func newTimeEntryStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running timer",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)

			user, err := client.GetCurrentUser(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			running, err := client.GetTimeEntryInProgress(context.Background(), user.ID)
			if err != nil {
				return err
			}

			if len(running) == 0 {
				return fmt.Errorf("no timer running")
			}

			entry := running[0]
			req := api.StopTimerRequest{
				End: time.Now().Format(time.RFC3339),
			}

			if err := client.StopTimeEntry(context.Background(), user.ID, req); err != nil {
				return err
			}

			// Get updated entry
			updated, err := client.GetTimeEntry(context.Background(), entry.ID)
			if err != nil {
				return err
			}

			dur := ""
			if updated.TimeInterval.Duration > 0 {
				dur = duration.Format(updated.TimeInterval.Duration)
			}

			cmd.Printf("Timer stopped. Duration: %s\n", dur)
			return formatOutput(cmd, updated, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "DURATION", "START", "END"})
		},
	}
	return cmd
}

func newTimeEntryBulkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk create time entries from JSON",
		Long: `Create multiple time entries from a JSON file or stdin.

Input format (JSON array):
[
  {
    "start": "2024-01-15T09:00:00Z",
    "duration": "1h",
    "project": "project-id",
    "description": "Working on feature"
  },
  {
    "start": "2024-01-15T10:00:00Z",
    "end": "2024-01-15T12:00:00Z",
    "project": "project-id",
    "description": "Meeting"
  }
]`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if filePath == "" && !stdin {
				return fmt.Errorf("must specify --file or --stdin")
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			// Read input
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

			// Parse JSON input
			var inputEntries []struct {
				Start       string   `json:"start"`
				End         string   `json:"end,omitempty"`
				Duration    string   `json:"duration,omitempty"`
				ProjectID   string   `json:"project"`
				TaskID      string   `json:"task,omitempty"`
				Description string   `json:"description,omitempty"`
				Billable    bool     `json:"billable,omitempty"`
				TagIDs      []string `json:"tags,omitempty"`
			}

			if err := json.Unmarshal(inputData, &inputEntries); err != nil {
				return fmt.Errorf("invalid JSON input: %w", err)
			}

			if len(inputEntries) == 0 {
				return fmt.Errorf("no time entries provided")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)

			var created []clockify.TimeEntry
			var errors []string

			for i, input := range inputEntries {
				// Parse start time
				startTime, err := time.Parse(time.RFC3339, input.Start)
				if err != nil {
					errors = append(errors, fmt.Sprintf("entry %d: invalid start time: %v", i+1, err))
					continue
				}

				req := api.CreateTimeEntryRequest{
					Start:       startTime.Format(time.RFC3339),
					ProjectID:   input.ProjectID,
					TaskID:      input.TaskID,
					Description: input.Description,
					Billable:    input.Billable,
					TagIDs:      input.TagIDs,
				}

				// Handle duration or end time
				if input.Duration != "" {
					secs, err := duration.Parse(input.Duration)
					if err != nil {
						errors = append(errors, fmt.Sprintf("entry %d: invalid duration: %v", i+1, err))
						continue
					}
					endTime := startTime.Add(time.Duration(secs) * time.Second)
					req.End = endTime.Format(time.RFC3339)
				} else if input.End != "" {
					endTime, err := time.Parse(time.RFC3339, input.End)
					if err != nil {
						errors = append(errors, fmt.Sprintf("entry %d: invalid end time: %v", i+1, err))
						continue
					}
					req.End = endTime.Format(time.RFC3339)
				}

				entry, err := client.CreateTimeEntry(context.Background(), req)
				if err != nil {
					errors = append(errors, fmt.Sprintf("entry %d: %v", i+1, err))
					continue
				}
				created = append(created, *entry)
			}

			if len(errors) > 0 {
				cmd.Printf("Created %d entries with %d errors:\n", len(created), len(errors))
				for _, e := range errors {
					cmd.Printf("  - %s\n", e)
				}
			} else {
				cmd.Printf("Created %d time entries.\n", len(created))
			}

			return formatOutput(cmd, created, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "DURATION", "START", "END"})
		},
	}
	cmd.Flags().String("file", "", "JSON file with time entries")
	cmd.Flags().Bool("stdin", false, "read time entries from stdin")
	return cmd
}

func fetchPaginatedTimeEntries(ctx context.Context, client *api.ClockifyClient, userID string, opts api.ListTimeEntriesOpts, page, pageSize int, noPagination bool) ([]clockify.TimeEntry, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		entries, pageInfo, err := client.ListTimeEntriesWithPagination(ctx, userID, opts)
		return entries, pageInfo, err
	}

	var allEntries []clockify.TimeEntry
	currentPage := 1

	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		entries, err := client.ListTimeEntries(ctx, userID, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(entries) == 0 {
			break
		}
		allEntries = append(allEntries, entries...)
		if len(entries) < pageSize {
			break
		}
		currentPage++
	}

	return allEntries, &clockify.PageInfo{
		Page:         1,
		PageSize:     pageSize,
		TotalPages:   currentPage,
		TotalResults: len(allEntries),
	}, nil
}
