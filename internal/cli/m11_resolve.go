package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/internal/config"
	clierrors "github.com/kim/clockify-cli/internal/errors"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/kim/clockify-cli/pkg/duration"
)

type nameCandidate struct {
	Name string
	ID   string
}

func ambiguousRefError(kind, ref string, cands []nameCandidate) *clierrors.CLIError {
	var b strings.Builder
	for _, c := range cands {
		b.WriteString(fmt.Sprintf("  - %s  (%s)\n", c.Name, c.ID))
	}
	return &clierrors.CLIError{
		Code:       clierrors.E004Validation,
		Message:    fmt.Sprintf("multiple %s match %q", kind, ref),
		Detail:     strings.TrimSpace(b.String()),
		Suggestion: fmt.Sprintf("use the full ID to disambiguate the %s", kind),
	}
}

func notFoundRefError(kind, ref string) *clierrors.CLIError {
	return &clierrors.CLIError{
		Code:       clierrors.E002NotFound,
		Message:    fmt.Sprintf("no %s found matching %q", kind, ref),
		Suggestion: fmt.Sprintf("check the spelling or list available %ss", kind),
	}
}

// looksLikeID reports whether s is a UUID-shaped identifier (with or without dashes).
func looksLikeID(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	hex := strings.ReplaceAll(s, "-", "")
	if len(hex) != 32 {
		return false
	}
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func splitTagRefs(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// resolveTaskRef resolves a task ID or name within a project.
func resolveTaskRef(ctx context.Context, client *api.ClockifyClient, projectID, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", nil
	}
	if looksLikeID(ref) {
		t, err := client.GetTask(ctx, projectID, ref)
		if err == nil && t != nil && t.ID != "" {
			return t.ID, nil
		}
	}
	t, err := resolveTaskByName(ctx, client, projectID, ref)
	if err != nil {
		return "", err
	}
	return t.ID, nil
}

// resolveTagRefs resolves comma-separated tag names/IDs to tag IDs.
func resolveTagRefs(ctx context.Context, client *api.ClockifyClient, refs []string) ([]string, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		if looksLikeID(ref) {
			ids = append(ids, ref)
			continue
		}
		t, err := resolveTagByName(ctx, client, ref)
		if err != nil {
			return nil, err
		}
		ids = append(ids, t.ID)
	}
	return ids, nil
}

// resolveTimeEntrySelector resolves time-entry refs: UUID passthrough, "current", or "last".
// When ref is not current/last, no API calls are made (userID may be empty).
func resolveTimeEntrySelector(ctx context.Context, client *api.ClockifyClient, userID, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("time entry ID, current, or last is required")
	}
	switch strings.ToLower(ref) {
	case "current":
		if userID == "" {
			return "", fmt.Errorf("current selector requires a user context")
		}
		running, err := client.GetTimeEntryInProgress(ctx, userID)
		if err != nil {
			return "", err
		}
		if len(running) == 0 {
			return "", &clierrors.CLIError{
				Code:       clierrors.E002NotFound,
				Message:    "no running timer",
				Suggestion: "start one with 'clockify time-entry start' or pass a time entry ID / last",
			}
		}
		return running[0].ID, nil
	case "last":
		if userID == "" {
			return "", fmt.Errorf("last selector requires a user context")
		}
		return mostRecentCompletedEntryID(ctx, client, userID)
	default:
		return ref, nil
	}
}

// resolveTimeEntryRef resolves a ref, fetching the current user only when needed.
func resolveTimeEntryRef(ctx context.Context, client *api.ClockifyClient, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	lower := strings.ToLower(ref)
	if lower != "current" && lower != "last" {
		if ref == "" {
			return "", fmt.Errorf("time entry ID, current, or last is required")
		}
		return ref, nil
	}
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return resolveTimeEntrySelector(ctx, client, user.ID, ref)
}

func mostRecentCompletedEntryID(ctx context.Context, client *api.ClockifyClient, userID string) (string, error) {
	entries, err := client.ListTimeEntries(ctx, userID, api.ListTimeEntriesOpts{PageSize: 50})
	if err != nil {
		return "", err
	}
	var best *clockify.TimeEntry
	for i := range entries {
		e := entries[i]
		if e.TimeInterval.End == nil {
			continue
		}
		if best == nil || e.TimeInterval.Start.After(best.TimeInterval.Start) {
			best = &entries[i]
		}
	}
	if best == nil {
		return "", &clierrors.CLIError{
			Code:       clierrors.E002NotFound,
			Message:    "no completed time entries found",
			Suggestion: "create a time entry first, or pass an explicit ID",
		}
	}
	return best.ID, nil
}

func weekStartFromConfig(cfg *config.Config) string {
	if cfg != nil {
		switch strings.ToLower(strings.TrimSpace(cfg.WeekStarts)) {
		case "sunday":
			return "sunday"
		case "monday", "":
			return "monday"
		}
	}
	return "monday"
}

func locationFromConfig(cfg *config.Config) *time.Location {
	if cfg != nil && strings.TrimSpace(cfg.TimeZone) != "" {
		if loc, err := time.LoadLocation(cfg.TimeZone); err == nil {
			return loc
		}
	}
	return time.Local
}

func parseHumanStart(s string, cfg *config.Config) (time.Time, error) {
	return parseHumanStartAt(s, cfg, time.Now())
}

func parseHumanStartAt(s string, cfg *config.Config, now time.Time) (time.Time, error) {
	loc := locationFromConfig(cfg)
	return duration.ParseHumanDateTime(s, loc, now)
}
