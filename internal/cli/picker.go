package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/internal/config"
	"github.com/kim/clockify-cli/internal/output"
	"github.com/spf13/cobra"
)

// interactiveRequested reports whether the user opted into pickers.
func interactiveRequested(cmd *cobra.Command, cfg *config.Config) bool {
	if cmd.Flags().Lookup("interactive") != nil {
		if v, err := cmd.Flags().GetBool("interactive"); err == nil && v {
			return true
		}
	}
	return cfg != nil && strings.EqualFold(strings.TrimSpace(cfg.Interactive), "true")
}

// interactiveBlockedReason returns a non-empty reason when interactive prompts
// must not run. Empty means prompting is allowed.
func interactiveBlockedReason(cmd *cobra.Command) string {
	if quietEnabled(cmd) {
		return "--quiet is set"
	}
	if cmd.Flags().Lookup("confirm") != nil {
		if v, _ := cmd.Flags().GetBool("confirm"); v {
			return "--confirm is set"
		}
	}
	if out, _ := cmd.Flags().GetString("output"); out != "" && out != "table" {
		return "--output " + out + " is machine-readable"
	}
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("BUILD_ID") != "" {
		return "CI environment detected"
	}
	if !output.IsTerminalStdout() {
		return "stdout is not a terminal"
	}
	if !output.IsTerminalStdin() {
		return "stdin is not a terminal"
	}
	return ""
}

type pickerClient struct {
	client *api.ClockifyClient
	userID string
}

// maybePromptMissing fills missing time-entry fields via interactive pickers.
// Returns updated refs; empty strings mean still missing.
// If interactive is not requested or prompts are blocked, missing fields are
// left as-is so callers can emit actionable missing-flag errors.
func maybePromptMissing(cmd *cobra.Command, cfg *config.Config, client *api.ClockifyClient, userID, projectRef, taskRef, tagsRef, description string) (string, string, string, string, error) {
	needsProject := projectRef == ""
	// Task/tags/description pickers are optional enhancements once project is known.
	if !interactiveRequested(cmd, cfg) {
		return projectRef, taskRef, tagsRef, description, nil
	}
	if needsProject {
		if reason := interactiveBlockedReason(cmd); reason != "" {
			return projectRef, taskRef, tagsRef, description, fmt.Errorf("interactive pickers unavailable: %s; pass --project <id|name> (not in CI/pipes)", reason)
		}
	} else if reason := interactiveBlockedReason(cmd); reason != "" {
		// Project already provided — skip optional pickers, do not fail.
		return projectRef, taskRef, tagsRef, description, nil
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if projectRef == "" {
		items, err := loadProjectItems(ctx, cmd, client)
		if err != nil {
			return projectRef, taskRef, tagsRef, description, err
		}
		id, err := pickItem(cmd, "project", items)
		if err != nil {
			return projectRef, taskRef, tagsRef, description, err
		}
		if id != "" {
			projectRef = id
		}
	}

	if projectRef != "" && taskRef == "" {
		pid := projectRef
		if p, err := resolveProjectRef(ctx, client, projectRef); err == nil {
			pid = p.ID
		}
		items, err := loadTaskItems(ctx, cmd, client, pid)
		if err != nil {
			return projectRef, taskRef, tagsRef, description, err
		}
		if len(items) > 0 {
			id, err := pickItem(cmd, "task", items)
			if err != nil {
				return projectRef, taskRef, tagsRef, description, err
			}
			if id != "" {
				taskRef = id
			}
		}
	}

	if tagsRef == "" {
		items, err := loadTagItems(ctx, cmd, client)
		if err != nil {
			return projectRef, taskRef, tagsRef, description, err
		}
		if len(items) > 0 {
			tags, err := pickTags(cmd, items)
			if err != nil {
				return projectRef, taskRef, tagsRef, description, err
			}
			if tags != "" {
				tagsRef = tags
			}
		}
	}

	if description == "" {
		desc, err := promptLine(cmd, "Description (optional)")
		if err != nil {
			return projectRef, taskRef, tagsRef, description, err
		}
		description = desc
	}
	return projectRef, taskRef, tagsRef, description, nil
}

// missingFlagError builds an actionable error when required flags are absent.
func missingFlagError(flag, example string) error {
	return fmt.Errorf("missing %s; pass %s, or run interactively with -i on a TTY (not in CI/pipes)", flag, example)
}

func promptLine(cmd *cobra.Command, label string) (string, error) {
	cmd.Printf("%s: ", label)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// pickItem presents a numbered fuzzy list and returns the selected item's ID.
// Empty input skips (returns ""). Filtering is case-insensitive substring match.
func pickItem(cmd *cobra.Command, label string, items []completionItem) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no %s available to pick from", label)
	}

	reader := bufio.NewReader(os.Stdin)
	filter := ""
	for attempt := 0; attempt < 5; attempt++ {
		matches := filterItems(items, filter)
		if len(matches) == 0 {
			return "", fmt.Errorf("no %s match filter %q", label, filter)
		}
		cmd.Printf("\nSelect %s", label)
		if filter != "" {
			cmd.Printf(" (filter: %q)", filter)
		}
		cmd.Println(":")
		for i, it := range matches {
			cmd.Printf("  %d) %s\n", i+1, it.Name)
		}
		cmd.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		choice := strings.TrimSpace(line)
		if choice == "" {
			return "", nil
		}

		var n int
		if _, err := fmt.Sscanf(choice, "%d", &n); err == nil && n >= 1 && n <= len(matches) {
			return matches[n-1].ID, nil
		}

		for _, it := range matches {
			if strings.EqualFold(it.Name, choice) {
				return it.ID, nil
			}
		}

		sub := filterItems(items, choice)
		if len(sub) == 1 {
			return sub[0].ID, nil
		}

		filter = choice
		cmd.Printf("  %d matches; enter a number, exact name, or narrower filter\n", len(sub))
	}
	return "", fmt.Errorf("selection cancelled for %s", label)
}

func filterItems(items []completionItem, filter string) []completionItem {
	if filter == "" {
		return items
	}
	var out []completionItem
	f := strings.ToLower(filter)
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Name), f) || strings.Contains(strings.ToLower(it.ID), f) {
			out = append(out, it)
		}
	}
	return out
}

func pickTags(cmd *cobra.Command, items []completionItem) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	cmd.Println("\nSelect tags (comma-separated numbers or names; empty to skip):")
	for i, it := range items {
		cmd.Printf("  %d) %s\n", i+1, it.Name)
	}
	cmd.Print("> ")
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	choice := strings.TrimSpace(line)
	if choice == "" {
		return "", nil
	}
	var ids []string
	for _, part := range strings.Split(choice, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(part, "%d", &n); err == nil && n >= 1 && n <= len(items) {
			ids = append(ids, items[n-1].ID)
			continue
		}
		matched := false
		for _, it := range items {
			if strings.EqualFold(it.Name, part) {
				ids = append(ids, it.ID)
				matched = true
				break
			}
		}
		if !matched {
			sub := filterItems(items, part)
			if len(sub) == 1 {
				ids = append(ids, sub[0].ID)
			} else {
				return "", fmt.Errorf("ambiguous or unknown tag %q", part)
			}
		}
	}
	return strings.Join(ids, ","), nil
}

func loadProjectItems(ctx context.Context, cmd *cobra.Command, client *api.ClockifyClient) ([]completionItem, error) {
	projects, err := client.ListProjects(ctx, api.ListProjectsOpts{})
	if err != nil {
		return nil, err
	}
	return namesFromProjects(projects), nil
}

func loadTagItems(ctx context.Context, cmd *cobra.Command, client *api.ClockifyClient) ([]completionItem, error) {
	tags, err := client.ListTags(ctx, api.ListTagsOpts{})
	if err != nil {
		return nil, err
	}
	return namesFromTags(tags), nil
}

func loadTaskItems(ctx context.Context, cmd *cobra.Command, client *api.ClockifyClient, projectID string) ([]completionItem, error) {
	if projectID == "" {
		return nil, nil
	}
	tasks, err := client.ListTasks(ctx, projectID, api.ListTasksOpts{})
	if err != nil {
		return nil, err
	}
	items := make([]completionItem, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, completionItem{ID: t.ID, Name: t.Name})
	}
	return items, nil
}
