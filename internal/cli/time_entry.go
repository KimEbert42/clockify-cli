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
		Long:    `List, get, create, update, and delete Clockify time entries.

Selectors for get/update/delete/stop/clone/split:
  <id>     time entry UUID
  current  the running timer
  last     most recent completed entry`,
	}
	cmd.AddCommand(newTimeEntryListCmd())
	cmd.AddCommand(newTimeEntryGetCmd())
	cmd.AddCommand(newTimeEntryCreateCmd())
	cmd.AddCommand(newTimeEntryUpdateCmd())
	cmd.AddCommand(newTimeEntryDeleteCmd())
	cmd.AddCommand(newTimeEntryStartCmd())
	cmd.AddCommand(newTimeEntryStopCmd())
	cmd.AddCommand(newTimeEntryBulkCmd())
	cmd.AddCommand(newTimeEntryCloneCmd())
	cmd.AddCommand(newTimeEntrySplitCmd())
	return cmd
}

func newTimeEntryListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List time entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			startFilter, _ := cmd.Flags().GetString("start")
			endFilter, _ := cmd.Flags().GetString("end")
			projectRef, _ := cmd.Flags().GetString("project")
			taskRef, _ := cmd.Flags().GetString("task")
			tagsRef, _ := cmd.Flags().GetString("tags")
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
			ctx := context.Background()

			user, err := client.GetCurrentUser(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			// Parse relative dates (local timezone)
			weekStart := weekStartFromConfig(cfg)
			if startFilter != "" && duration.IsRelativeDate(startFilter) {
				t, err := duration.ParseRelativeDateIn(startFilter, weekStart, locationFromConfig(cfg), time.Now())
				if err != nil {
					return fmt.Errorf("invalid start date: %w", err)
				}
				startFilter = t.Format(time.RFC3339)
			}
			if endFilter != "" && duration.IsRelativeDate(endFilter) {
				t, err := duration.ParseRelativeDateIn(endFilter, weekStart, locationFromConfig(cfg), time.Now())
				if err != nil {
					return fmt.Errorf("invalid end date: %w", err)
				}
				endFilter = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()).Format(time.RFC3339)
			}

			// Default to today if no date filters provided
			if startFilter == "" && endFilter == "" && !inProgress {
				today := time.Now().In(locationFromConfig(cfg))
				startFilter = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()).Format(time.RFC3339)
				endFilter = time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, today.Location()).Format(time.RFC3339)
			}

			// Resolve name filters; raw IDs pass through resolveProjectRef fast path.
			projectID := projectRef
			taskID := taskRef
			var tagIDs []string
			if projectRef != "" || taskRef != "" || tagsRef != "" {
				var err2 error
				projectID, taskID, tagIDs, err2 = resolveTimeEntryRefs(ctx, client, projectRef, taskRef, tagsRef)
				if err2 != nil {
					return err2
				}
			}
			tagsFilter := ""
			if len(tagIDs) > 0 {
				tagsFilter = strings.Join(tagIDs, ",")
			} else if tagsRef != "" {
				tagsFilter = tagsRef
			}
			if projectRef != "" && projectID == "" {
				projectID = projectRef
			}

			if inProgress {
				entries, err := client.GetTimeEntryInProgress(ctx, user.ID)
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
				ProjectID:   projectID,
				TaskID:      taskID,
				Tags:        tagsFilter,
				Description: descriptionFilter,
				Hydrated:    hydrated,
			}

			entries, pageInfo, err := fetchPaginatedTimeEntries(ctx, client, user.ID, opts, page, pageSize, noPagination)
			if err != nil {
				return err
			}

			if len(entries) == 0 && !quiet {
				cmd.Println("No time entries found.")
				return nil
			}

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
	cmd.Flags().String("start", "", "start date (YYYY-MM-DD, HH:MM, relative: today, yesterday, this-week, last-week, this-month, last-month, or RFC3339)")
	cmd.Flags().String("end", "", "end date (same formats as --start)")
	cmd.Flags().String("project", "", "filter by project ID or name")
	cmd.Flags().String("task", "", "filter by task ID or name (requires --project)")
	cmd.Flags().String("tags", "", "filter by tag IDs or names (comma-separated)")
	cmd.Flags().String("description", "", "search by description")
	cmd.Flags().Bool("in-progress", false, "show only running timers")
	cmd.Flags().Bool("hydrated", false, "include full project/task/user objects in output")
	_ = cmd.RegisterFlagCompletionFunc("project", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("task", completeTaskRefC)
	_ = cmd.RegisterFlagCompletionFunc("tags", completeTagRefC)
	return cmd
}

func newTimeEntryGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id|current|last]",
		Short: "Get time entry by ID, current, or last",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			if len(args) > 0 {
				id = args[0]
			}
			hydrated, _ := cmd.Flags().GetBool("hydrated")

			if id == "" {
				return fmt.Errorf("must specify time entry ID, current, or last")
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

			id, err = resolveTimeEntryRef(ctx, client, id)
			if err != nil {
				return err
			}

			var entry *clockify.TimeEntry
			if hydrated {
				entries, err := client.BatchGetTimeEntries(ctx, []string{id}, true)
				if err != nil {
					return err
				}
				if len(entries) == 0 {
					return fmt.Errorf("time entry %s not found", id)
				}
				entry = &entries[0]
			} else {
				entry, err = client.GetTimeEntry(ctx, id)
				if err != nil {
					return err
				}
			}

			return formatOutput(cmd, entry, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "TASK", "USER", "DURATION", "START", "END", "BILLABLE", "TAGS"})
		},
	}
	cmd.Flags().String("id", "", "time entry ID, current, or last")
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
			projectRef, _ := cmd.Flags().GetString("project")
			taskRef, _ := cmd.Flags().GetString("task")
			description, _ := cmd.Flags().GetString("description")
			tagsStr, _ := cmd.Flags().GetString("tags")
			billable, _ := cmd.Flags().GetBool("billable")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			// JSON input from --file/--stdin; explicit flags override JSON fields.
			var jsonData map[string]any
			if filePath != "" || stdin {
				jsonData = map[string]any{}
				if err := readJSONInput(filePath, stdin, &jsonData); err != nil {
					return err
				}
				if startStr == "" {
					if v, ok := jsonData["start"].(string); ok {
						startStr = v
					}
				}
				if durationStr == "" {
					if v, ok := jsonData["duration"].(string); ok {
						durationStr = v
					}
				}
				if endStr == "" {
					if v, ok := jsonData["end"].(string); ok {
						endStr = v
					}
				}
				if projectRef == "" {
					if v, ok := jsonData["projectId"].(string); ok {
						projectRef = v
					}
					if v, ok := jsonData["project"].(string); ok && projectRef == "" {
						projectRef = v
					}
				}
				if taskRef == "" {
					if v, ok := jsonData["taskId"].(string); ok {
						taskRef = v
					}
				}
				if description == "" {
					if v, ok := jsonData["description"].(string); ok {
						description = v
					}
				}
				if tagsStr == "" {
					if v, ok := jsonData["tagIds"].([]any); ok {
						parts := make([]string, 0, len(v))
						for _, raw := range v {
							if s, ok := raw.(string); ok && s != "" {
								parts = append(parts, s)
							}
						}
						tagsStr = strings.Join(parts, ",")
					}
				}
				if !cmd.Flags().Changed("billable") {
					if v, ok := jsonData["billable"].(bool); ok {
						billable = v
					}
				}
			}

			// Validate required flags before network/auth (fail fast for scripts).
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
			ctx := context.Background()

			user, err := client.GetCurrentUser(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			// Interactive pickers for missing fields
			projectRef, taskRef, tagsStr, description, err = maybePromptMissing(cmd, cfg, client, user.ID, projectRef, taskRef, tagsStr, description)
			if err != nil {
				return err
			}

			if projectRef == "" {
				return missingFlagError("--project", "--project <id|name>")
			}

			projectID, taskID, tagIDs, err := resolveTimeEntryRefs(ctx, client, projectRef, taskRef, tagsStr)
			if err != nil {
				return err
			}

			startTime, err := parseStartTimeFlexibleWith(startStr, cfg)
			if err != nil {
				return err
			}

			req := api.CreateTimeEntryRequest{
				Start:       startTime.Format(time.RFC3339),
				ProjectID:   projectID,
				TaskID:      taskID,
				Description: description,
				Billable:    billable,
				TagIDs:      tagIDs,
			}

			if durationStr != "" {
				secs, err := duration.Parse(durationStr)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
				endTime := startTime.Add(time.Duration(secs) * time.Second)
				req.End = endTime.Format(time.RFC3339)
			} else {
				endTime, err := parseStartTimeFlexibleWith(endStr, cfg)
				if err != nil {
					return err
				}
				req.End = endTime.Format(time.RFC3339)
			}

			entry, err := client.CreateTimeEntry(ctx, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, entry, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "DURATION", "START", "END", "BILLABLE"})
		},
	}
	cmd.Flags().String("start", "", "start time (RFC3339, YYYY-MM-DD, HH:MM, now, ±offsets, or relative date)")
	cmd.Flags().String("duration", "", "duration (e.g., 1h, 30m, 1h30m)")
	cmd.Flags().String("end", "", "end time (same formats as --start)")
	cmd.Flags().String("project", "", "project ID or name")
	cmd.Flags().String("task", "", "task ID or name")
	cmd.Flags().String("description", "", "description")
	cmd.Flags().String("tags", "", "tag IDs or names (comma-separated)")
	cmd.Flags().Bool("billable", false, "mark as billable")
	cmd.Flags().String("file", "", "JSON file with time entry data")
	cmd.Flags().Bool("stdin", false, "read time entry data from stdin")
	cmd.Flags().BoolP("interactive", "i", false, "prompt interactively for missing fields (TTY only)")
	_ = cmd.RegisterFlagCompletionFunc("project", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("task", completeTaskRefC)
	_ = cmd.RegisterFlagCompletionFunc("tags", completeTagRefC)
	return cmd
}

func newTimeEntryUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id|current|last] [id...]",
		Short: "Update one or more time entries",
		Args:  cobra.ArbitraryArgs,
		RunE:  runTimeEntryUpdateMulti,
	}
	cmd.Flags().String("id", "", "time entry ID, current, or last")
	cmd.Flags().Bool("stdin-ids", false, "read time entry IDs from stdin (one per line or JSON array)")
	cmd.Flags().String("start", "", "start time (RFC3339, YYYY-MM-DD, HH:MM, now, ±offsets, or relative date)")
	cmd.Flags().String("duration", "", "duration (e.g., 1h, 30m, 1h30m)")
	cmd.Flags().String("end", "", "end time (same formats as --start)")
	cmd.Flags().String("project", "", "project ID or name")
	cmd.Flags().String("task", "", "task ID or name")
	cmd.Flags().String("description", "", "description")
	cmd.Flags().String("tags", "", "tag IDs or names (comma-separated)")
	cmd.Flags().Bool("billable", false, "mark as billable")
	cmd.Flags().String("file", "", "JSON file with time entry data")
	cmd.Flags().Bool("stdin", false, "read time entry data from stdin")
	_ = cmd.RegisterFlagCompletionFunc("project", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("task", completeTaskRefC)
	_ = cmd.RegisterFlagCompletionFunc("tags", completeTagRefC)
	return cmd
}

func newTimeEntryDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id|current|last]",
		Short: "Delete a time entry",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			confirm, _ := cmd.Flags().GetBool("confirm")
			if len(args) > 0 {
				id = args[0]
			}

			if id == "" {
				return fmt.Errorf("must specify time entry ID, current, or last")
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

			id, err = resolveTimeEntryRef(ctx, client, id)
			if err != nil {
				return err
			}

			entry, err := client.GetTimeEntry(ctx, id)
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

			if err := client.DeleteTimeEntry(ctx, id); err != nil {
				return err
			}

			cmd.Printf("Time entry %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "time entry ID, current, or last")
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
			projectRef, _ := cmd.Flags().GetString("project")
			taskRef, _ := cmd.Flags().GetString("task")
			description, _ := cmd.Flags().GetString("description")
			tagsStr, _ := cmd.Flags().GetString("tags")
			billable, _ := cmd.Flags().GetBool("billable")
			startStr, _ := cmd.Flags().GetString("start")

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			ctx := context.Background()

			user, err := client.GetCurrentUser(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			projectRef, taskRef, tagsStr, description, err = maybePromptMissing(cmd, cfg, client, user.ID, projectRef, taskRef, tagsStr, description)
			if err != nil {
				return err
			}

			if projectRef == "" {
				return missingFlagError("--project", "--project <id|name>")
			}

			running, err := client.GetTimeEntryInProgress(ctx, user.ID)
			if err != nil {
				return err
			}

			if len(running) > 0 {
				return fmt.Errorf("timer already running (ID: %s); stop it first with 'clockify time-entry stop'", running[0].ID)
			}

			startTime := time.Now()
			if startStr != "" {
				startTime, err = parseStartTimeFlexibleWith(startStr, cfg)
				if err != nil {
					return err
				}
			}

			// Resolve project name/ID; pass through raw IDs when resolution is
			// unnecessary (UUID-shaped) via resolveProjectRef fast path.
			projectID, taskID, tagIDs, err := resolveTimeEntryRefs(ctx, client, projectRef, taskRef, tagsStr)
			if err != nil {
				return err
			}

			req := api.CreateTimeEntryRequest{
				Start:       startTime.Format(time.RFC3339),
				ProjectID:   projectID,
				TaskID:      taskID,
				Description: description,
				Billable:    billable,
				TagIDs:      tagIDs,
			}

			entry, err := client.CreateTimeEntry(ctx, req)
			if err != nil {
				return err
			}

			cmd.Printf("Timer started (ID: %s)\n", entry.ID)
			return formatOutput(cmd, entry, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "START"})
		},
	}
	cmd.Flags().String("project", "", "project ID or name")
	cmd.Flags().String("task", "", "task ID or name")
	cmd.Flags().String("description", "", "description")
	cmd.Flags().String("tags", "", "tag IDs or names (comma-separated)")
	cmd.Flags().Bool("billable", false, "mark as billable")
	cmd.Flags().String("start", "", "optional start time (default: now)")
	cmd.Flags().BoolP("interactive", "i", false, "prompt interactively for missing fields (TTY only)")
	_ = cmd.RegisterFlagCompletionFunc("project", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("task", completeTaskRefC)
	_ = cmd.RegisterFlagCompletionFunc("tags", completeTagRefC)
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
			ctx := context.Background()

			user, err := client.GetCurrentUser(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			running, err := client.GetTimeEntryInProgress(ctx, user.ID)
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

			if err := client.StopTimeEntry(ctx, user.ID, req); err != nil {
				return err
			}

			updated, err := client.GetTimeEntry(ctx, entry.ID)
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
    "project": "project-id-or-name",
    "description": "Working on feature"
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
			ctx := context.Background()

			var created []clockify.TimeEntry
			var errors []string

			for i, input := range inputEntries {
				startTime, err := parseStartTimeFlexibleWith(input.Start, cfg)
				if err != nil {
					errors = append(errors, fmt.Sprintf("entry %d: invalid start time: %v", i+1, err))
					continue
				}

				projectID, taskID, tagIDs, err := resolveTimeEntryRefs(ctx, client, input.ProjectID, input.TaskID, strings.Join(input.TagIDs, ","))
				if err != nil {
					errors = append(errors, fmt.Sprintf("entry %d: %v", i+1, err))
					continue
				}

				req := api.CreateTimeEntryRequest{
					Start:       startTime.Format(time.RFC3339),
					ProjectID:   projectID,
					TaskID:      taskID,
					Description: input.Description,
					Billable:    input.Billable,
					TagIDs:      tagIDs,
				}

				if input.Duration != "" {
					secs, err := duration.Parse(input.Duration)
					if err != nil {
						errors = append(errors, fmt.Sprintf("entry %d: invalid duration: %v", i+1, err))
						continue
					}
					endTime := startTime.Add(time.Duration(secs) * time.Second)
					req.End = endTime.Format(time.RFC3339)
				} else if input.End != "" {
					endTime, err := parseStartTimeFlexibleWith(input.End, cfg)
					if err != nil {
						errors = append(errors, fmt.Sprintf("entry %d: invalid end time: %v", i+1, err))
						continue
					}
					req.End = endTime.Format(time.RFC3339)
				}

				entry, err := client.CreateTimeEntry(ctx, req)
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

func newTimeEntryCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone [id|last]",
		Short: "Clone a time entry (start timer or create completed entry)",
		Long: `Copy project, task, tags, description, and billable from an existing entry.

Without --start/--end/--duration: starts a new timer with copied metadata.
With --start and --end/--duration: creates a completed entry.
Any flag overrides the copied field.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTimeEntryClone,
	}
	cmd.Flags().String("id", "", "source time entry ID or last")
	cmd.Flags().String("start", "", "start time for the new entry (enables completed-entry mode with --end/--duration)")
	cmd.Flags().String("end", "", "end time for the new entry")
	cmd.Flags().String("duration", "", "duration for the new entry")
	cmd.Flags().String("project", "", "override project (ID or name)")
	cmd.Flags().String("task", "", "override task (ID or name)")
	cmd.Flags().String("tags", "", "override tags (comma-separated IDs or names)")
	cmd.Flags().String("description", "", "override description")
	cmd.Flags().Bool("billable", false, "override billable flag")
	_ = cmd.RegisterFlagCompletionFunc("project", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("task", completeTaskRefC)
	_ = cmd.RegisterFlagCompletionFunc("tags", completeTagRefC)
	return cmd
}

func newTimeEntrySplitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "split <id|current|last> --at HH:MM [--at HH:MM...]",
		Short: "Split a time entry at one or more times",
		Long: `Divide an entry at the given local times. The original entry ends at the
first split point; remaining segments are created with identical metadata.
Split points must fall strictly inside the entry.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runTimeEntrySplit,
	}
	cmd.Flags().StringArray("at", nil, "split point (HH:MM or full timestamp; repeatable)")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
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
