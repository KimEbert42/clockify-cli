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
	"github.com/kim/clockify-cli/internal/config"
	clierrors "github.com/kim/clockify-cli/internal/errors"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/kim/clockify-cli/pkg/duration"
	"github.com/spf13/cobra"
)

// parseStartTimeFlexible keeps backward-compatible parsing and delegates to the
// human-time grammar (local timezone).
func parseStartTimeFlexible(s string) (time.Time, error) {
	return parseStartTimeFlexibleWith(s, nil)
}

func parseStartTimeFlexibleWith(s string, cfg *config.Config) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date/time value")
	}
	return duration.ParseHumanDateTime(s, locationFromConfig(cfg), time.Now())
}

// resolveTimeEntryRefs resolves project/task/tags flags to API IDs.
func resolveTimeEntryRefs(ctx context.Context, client *api.ClockifyClient, projectRef, taskRef, tagsRef string) (projectID, taskID string, tagIDs []string, err error) {
	if projectRef != "" {
		p, rerr := resolveProjectRef(ctx, client, projectRef)
		if rerr != nil {
			return "", "", nil, rerr
		}
		projectID = p.ID
	}
	if taskRef != "" {
		if projectID == "" {
			return "", "", nil, fmt.Errorf("--task requires --project (task names are project-scoped)")
		}
		tid, rerr := resolveTaskRef(ctx, client, projectID, taskRef)
		if rerr != nil {
			return "", "", nil, rerr
		}
		taskID = tid
	}
	if tagsRef != "" {
		ids, rerr := resolveTagRefs(ctx, client, splitTagRefs(tagsRef))
		if rerr != nil {
			return "", "", nil, rerr
		}
		tagIDs = ids
	}
	return projectID, taskID, tagIDs, nil
}

// cloneFlags describes which fields were explicitly set on the CLI.
type cloneFlags struct {
	Project     string
	Task        string
	Tags        string
	Description string
	Billable    *bool
	Start       string
	End         string
	Duration    string
}

func runTimeEntryClone(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	outputFormat, _ := cmd.Flags().GetString("output")
	srcRef := ""
	if len(args) > 0 {
		srcRef = args[0]
	} else {
		srcRef, _ = cmd.Flags().GetString("id")
	}
	if srcRef == "" {
		srcRef = "last"
	}

	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	durationStr, _ := cmd.Flags().GetString("duration")
	projectRef, _ := cmd.Flags().GetString("project")
	taskRef, _ := cmd.Flags().GetString("task")
	tagsStr, _ := cmd.Flags().GetString("tags")
	description, _ := cmd.Flags().GetString("description")
	billable, _ := cmd.Flags().GetBool("billable")
	billableSet := cmd.Flags().Changed("billable")

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

	srcID, err := resolveTimeEntryRef(ctx, client, srcRef)
	if err != nil {
		return err
	}
	src, err := client.GetTimeEntry(ctx, srcID)
	if err != nil {
		return err
	}

	// Defaults from source
	if projectRef == "" && src.ProjectID != "" {
		projectRef = src.ProjectID
	}
	if taskRef == "" && src.TaskID != "" {
		taskRef = src.TaskID
	}
	if tagsStr == "" && len(src.TagIDs) > 0 {
		tagsStr = strings.Join(src.TagIDs, ",")
	}
	if description == "" {
		description = src.Description
	}
	if !billableSet {
		billable = src.Billable
	}

	projectID, taskID, tagIDs, err := resolveTimeEntryRefs(ctx, client, projectRef, taskRef, tagsStr)
	if err != nil {
		return err
	}

	now := time.Now()
	startTime := now
	if startStr != "" {
		startTime, err = parseStartTimeFlexibleWith(startStr, cfg)
		if err != nil {
			return err
		}
	}

	req := api.CreateTimeEntryRequest{
		Start:       startTime.Format(time.RFC3339),
		ProjectID:   projectID,
		TaskID:      taskID,
		Description: description,
		Billable:    billable,
		TagIDs:      tagIDs,
	}

	// If --start/--end/--duration given → completed entry; else start timer (no end).
	if startStr != "" || endStr != "" || durationStr != "" {
		if durationStr != "" {
			secs, err := duration.Parse(durationStr)
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
			req.End = startTime.Add(time.Duration(secs) * time.Second).Format(time.RFC3339)
		} else if endStr != "" {
			endTime, err := parseStartTimeFlexibleWith(endStr, cfg)
			if err != nil {
				return err
			}
			req.End = endTime.Format(time.RFC3339)
		} else if startStr != "" && endStr == "" && durationStr == "" {
			// clone with only --start: treat as start of new timer? PLAN says
			// --start/--end creates completed entry. With only --start, start a timer at that time.
			// Leave End empty.
		}
	}

	// Conflict check when starting a timer
	if req.End == "" {
		running, err := client.GetTimeEntryInProgress(ctx, user.ID)
		if err != nil {
			return err
		}
		if len(running) > 0 {
			return fmt.Errorf("timer already running (ID: %s); stop it first with 'clockify time-entry stop'", running[0].ID)
		}
	}

	entry, err := client.CreateTimeEntry(ctx, req)
	if err != nil {
		return err
	}

	if req.End == "" {
		cmd.Printf("Timer started (ID: %s) cloned from %s\n", entry.ID, srcID)
	} else {
		cmd.Printf("Time entry created (ID: %s) cloned from %s\n", entry.ID, srcID)
	}
	return formatOutput(cmd, entry, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "DURATION", "START", "END", "BILLABLE"})
}

func runTimeEntrySplit(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	outputFormat, _ := cmd.Flags().GetString("output")
	confirm, _ := cmd.Flags().GetBool("confirm")
	if len(args) < 1 {
		return fmt.Errorf("must specify time entry ID (or current/last)")
	}
	id := args[0]
	atStrs, _ := cmd.Flags().GetStringArray("at")
	if len(atStrs) == 0 {
		return fmt.Errorf("must specify at least one --at HH:MM split point")
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

	entryID, err := resolveTimeEntryRef(ctx, client, id)
	if err != nil {
		return err
	}
	entry, err := client.GetTimeEntry(ctx, entryID)
	if err != nil {
		return err
	}
	if entry.TimeInterval.End == nil {
		return fmt.Errorf("cannot split a running timer; stop it first")
	}

	loc := locationFromConfig(cfg)
	baseDate := entry.TimeInterval.Start.In(loc)
	var points []time.Time
	for _, at := range atStrs {
		t, err := parseStartTimeFlexibleWith(at, cfg)
		if err != nil {
			return fmt.Errorf("invalid --at %q: %w", at, err)
		}
		// Interpret HH:MM as same local date as entry start when only time given.
		if isTimeOnly(at) {
			t = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
		}
		points = append(points, t)
	}

	// Validate split points strictly inside the entry
	start := entry.TimeInterval.Start
	end := *entry.TimeInterval.End
	for i := 0; i < len(points); i++ {
		if !points[i].After(start) || !points[i].Before(end) {
			return &clierrors.CLIError{
				Code:       clierrors.E004Validation,
				Message:    fmt.Sprintf("split point %s is outside the entry range", points[i].Format(time.RFC3339)),
				Detail:     fmt.Sprintf("entry starts %s and ends %s", start.Format(time.RFC3339), end.Format(time.RFC3339)),
				Suggestion: "choose --at times strictly between start and end",
			}
		}
		if i > 0 && !points[i].After(points[i-1]) {
			return &clierrors.CLIError{
				Code:       clierrors.E004Validation,
				Message:    "split points must be strictly increasing",
				Suggestion: "sort --at times or remove duplicates",
			}
		}
	}

	if !confirm {
		cmd.Printf("Split time entry %s into %d segments at %s?\n",
			entry.ID, len(points)+1, strings.Join(atStrs, ", "))
		cmd.Print("Are you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			cmd.Println("Aborted.")
			return nil
		}
	}

	// Update original to end at first split point
	firstReq := api.UpdateTimeEntryRequest{
		Start:       entry.TimeInterval.Start.Format(time.RFC3339),
		End:         points[0].Format(time.RFC3339),
		ProjectID:   entry.ProjectID,
		TaskID:      entry.TaskID,
		Description: entry.Description,
		Billable:    entry.Billable,
		TagIDs:      entry.TagIDs,
	}
	if _, err := client.UpdateTimeEntry(ctx, entry.ID, firstReq); err != nil {
		return err
	}

	var created []clockify.TimeEntry
	prev := points[0]
	for i := 1; i < len(points); i++ {
		req := api.CreateTimeEntryRequest{
			Start:       prev.Format(time.RFC3339),
			End:         points[i].Format(time.RFC3339),
			ProjectID:   entry.ProjectID,
			TaskID:      entry.TaskID,
			Description: entry.Description,
			Billable:    entry.Billable,
			TagIDs:      entry.TagIDs,
		}
		e, err := client.CreateTimeEntry(ctx, req)
		if err != nil {
			return err
		}
		created = append(created, *e)
		prev = points[i]
	}
	// Final remainder
	remainder := api.CreateTimeEntryRequest{
		Start:       prev.Format(time.RFC3339),
		End:         end.Format(time.RFC3339),
		ProjectID:   entry.ProjectID,
		TaskID:      entry.TaskID,
		Description: entry.Description,
		Billable:    entry.Billable,
		TagIDs:      entry.TagIDs,
	}
	e, err := client.CreateTimeEntry(ctx, remainder)
	if err != nil {
		return err
	}
	created = append(created, *e)

	cmd.Printf("Split %s into %d segments.\n", entry.ID, len(created))
	return formatOutput(cmd, created, outputFormat, []string{"ID", "DESCRIPTION", "START", "END", "DURATION"})
}

func isTimeOnly(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if strings.HasSuffix(s, "am") || strings.HasSuffix(s, "pm") {
		s = strings.TrimSpace(s[:len(s)-2])
	}
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

// runTimeEntryUpdateMulti applies the same flags to several entries.
func runTimeEntryUpdateMulti(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	outputFormat, _ := cmd.Flags().GetString("output")
	confirm, _ := cmd.Flags().GetBool("confirm")
	stdinIDs, _ := cmd.Flags().GetBool("stdin-ids")

	ids := append([]string{}, args...)
	if stdinIDs {
		data, err := readAllStdin()
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			// Support JSON array on one line or raw IDs per line
			if strings.HasPrefix(line, "[") {
				var arr []string
				if err := jsonUnmarshal(line, &arr); err == nil {
					ids = append(ids, arr...)
					continue
				}
			}
			ids = append(ids, line)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("must specify one or more time entry IDs (or use --stdin-ids)")
	}

	// Collect shared update flags
	startStr, _ := cmd.Flags().GetString("start")
	durationStr, _ := cmd.Flags().GetString("duration")
	endStr, _ := cmd.Flags().GetString("end")
	projectRef, _ := cmd.Flags().GetString("project")
	taskRef, _ := cmd.Flags().GetString("task")
	description, _ := cmd.Flags().GetString("description")
	tagsStr, _ := cmd.Flags().GetString("tags")
	billable, _ := cmd.Flags().GetBool("billable")
	billableSet := cmd.Flags().Changed("billable")

	if startStr == "" && durationStr == "" && endStr == "" && projectRef == "" && taskRef == "" && description == "" && tagsStr == "" && !billableSet {
		return fmt.Errorf("specify at least one field to update")
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

	var resolvedIDs []string
	for _, id := range ids {
		rid, err := resolveTimeEntryRef(ctx, client, id)
		if err != nil {
			return err
		}
		resolvedIDs = append(resolvedIDs, rid)
	}

	projectID, taskID, tagIDs, err := resolveTimeEntryRefs(ctx, client, projectRef, taskRef, tagsStr)
	if err != nil {
		return err
	}

	if !confirm {
		cmd.Printf("Update %d time entries?\n", len(resolvedIDs))
		cmd.Print("Are you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			cmd.Println("Aborted.")
			return nil
		}
	}

	var updated []clockify.TimeEntry
	var failures []string
	for _, id := range resolvedIDs {
		existing, err := client.GetTimeEntry(ctx, id)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", id, err))
			continue
		}
		req := api.UpdateTimeEntryRequest{
			Start:       existing.TimeInterval.Start.Format(time.RFC3339),
			End:         "",
			ProjectID:   existing.ProjectID,
			TaskID:      existing.TaskID,
			Description: existing.Description,
			Billable:    existing.Billable,
			TagIDs:      existing.TagIDs,
		}
		if existing.TimeInterval.End != nil {
			req.End = existing.TimeInterval.End.Format(time.RFC3339)
		}
		if startStr != "" {
			st, err := parseStartTimeFlexibleWith(startStr, cfg)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: %v", id, err))
				continue
			}
			req.Start = st.Format(time.RFC3339)
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
		if tagIDs != nil {
			req.TagIDs = tagIDs
		}
		if billableSet {
			req.Billable = billable
		}
		if durationStr != "" {
			secs, err := duration.Parse(durationStr)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: invalid duration: %v", id, err))
				continue
			}
			st, _ := time.Parse(time.RFC3339, req.Start)
			req.End = st.Add(time.Duration(secs) * time.Second).Format(time.RFC3339)
		} else if endStr != "" {
			et, err := parseStartTimeFlexibleWith(endStr, cfg)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: %v", id, err))
				continue
			}
			req.End = et.Format(time.RFC3339)
		}
		entry, err := client.UpdateTimeEntry(ctx, id, req)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", id, err))
			continue
		}
		updated = append(updated, *entry)
	}

	if len(failures) > 0 {
		cmd.Printf("Updated %d entries, %d failed:\n", len(updated), len(failures))
		for _, f := range failures {
			cmd.Printf("  - %s\n", f)
		}
	} else {
		cmd.Printf("Updated %d time entries.\n", len(updated))
	}
	if err := formatOutput(cmd, updated, outputFormat, []string{"ID", "DESCRIPTION", "PROJECT", "START", "END"}); err != nil {
		return err
	}
	if len(failures) > 0 {
		return fmt.Errorf("multi-update completed with %d failures", len(failures))
	}
	return nil
}

func readAllStdin() ([]byte, error) {
	return io.ReadAll(os.Stdin)
}

func jsonUnmarshal(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}
