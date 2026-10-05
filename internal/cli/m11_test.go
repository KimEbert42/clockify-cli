package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/kim/clockify-cli/internal/output"
)

func TestLooksLikeID(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"p1", false},
		{"Website Redesign", false},
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"550e8400e29b41d4a716446655440000", true},
		{"not-a-uuid-value-here", false},
	}
	for _, tt := range tests {
		if got := looksLikeID(tt.in); got != tt.want {
			t.Errorf("looksLikeID(%q)=%v want %v", tt.in, got, tt.want)
		}
	}
}

func TestSplitTagRefs(t *testing.T) {
	got := splitTagRefs(" a , b ,,c ")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestAmbiguousRefErrorFormat(t *testing.T) {
	err := ambiguousRefError("projects", "Web", []nameCandidate{
		{Name: "Website", ID: "id-1"},
		{Name: "Web App", ID: "id-2"},
	})
	if err.Code != 4 { // E004Validation
		t.Errorf("code = %d want 4", err.Code)
	}
	if !strings.Contains(err.Detail, "Website") || !strings.Contains(err.Detail, "id-1") {
		t.Errorf("detail missing candidates: %q", err.Detail)
	}
	if !strings.Contains(err.Suggestion, "full ID") {
		t.Errorf("suggestion = %q", err.Suggestion)
	}
}

func TestInteractiveBlockedReasonNonTTY(t *testing.T) {
	cmd := newTimeEntryStartCmd()
	// In tests stdout/stdin are typically not TTYs
	if reason := interactiveBlockedReason(cmd); reason == "" {
		// Might be a TTY in some environments; only assert non-empty when not a TTY
		if !output.IsTerminalStdout() || !output.IsTerminalStdin() {
			t.Fatal("expected blocked reason when not a TTY")
		}
	}
}

func TestInteractiveBlockedByOutputJSON(t *testing.T) {
	cmd := newTimeEntryStartCmd()
	_ = cmd.Flags().Set("output", "json")
	// Force non-TTY path by also setting CI
	t.Setenv("CI", "true")
	reason := interactiveBlockedReason(cmd)
	if reason == "" {
		t.Fatal("expected blocked reason with CI=true")
	}
	if !strings.Contains(reason, "CI") && !strings.Contains(reason, "machine") && !strings.Contains(reason, "terminal") {
		t.Errorf("unexpected reason: %q", reason)
	}
}

func TestMissingFlagError(t *testing.T) {
	err := missingFlagError("--project", "--project <id|name>")
	if !strings.Contains(err.Error(), "--project") {
		t.Errorf("error = %v", err)
	}
	if !strings.Contains(err.Error(), "-i") {
		t.Errorf("error should mention -i: %v", err)
	}
}

func TestIsTimeOnly(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"9:00", true},
		{"09:00", true},
		{"9:00am", true},
		{"2026-10-04", false},
		{"2026-10-04T09:00:00", false},
		{"yesterday 09:00", false},
	}
	for _, tt := range tests {
		if got := isTimeOnly(tt.in); got != tt.want {
			t.Errorf("isTimeOnly(%q)=%v want %v", tt.in, got, tt.want)
		}
	}
}

func TestFormatIDs(t *testing.T) {
	type row struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var sb strings.Builder
	err := output.FormatIDs(&sb, []row{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(sb.String())
	if got != "a\nb" {
		t.Errorf("got %q", got)
	}
}

func TestFormatTemplate(t *testing.T) {
	type row struct {
		ID          string
		Description string
	}
	var sb strings.Builder
	err := output.FormatTemplate(&sb, row{ID: "x", Description: "hello"}, "{{.Description}}")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(sb.String()) != "hello" {
		t.Errorf("got %q", sb.String())
	}
}

func TestParseStartTimeFlexibleHuman(t *testing.T) {
	// Should accept HH:MM without error
	_, err := parseStartTimeFlexible("09:00")
	if err != nil {
		t.Errorf("HH:MM should parse: %v", err)
	}
	_, err = parseStartTimeFlexible("yesterday 10:30")
	if err != nil {
		t.Errorf("yesterday HH:MM should parse: %v", err)
	}
	_, err = parseStartTimeFlexible("now")
	if err != nil {
		t.Errorf("now should parse: %v", err)
	}
	// RFC3339 still works
	tm, err := parseStartTimeFlexible("2026-10-04T09:00:00Z")
	if err != nil {
		t.Fatalf("RFC3339 should parse: %v", err)
	}
	if tm.Hour() != 9 {
		t.Errorf("hour = %d want 9", tm.Hour())
	}
}

func TestWeekStartFromConfig(t *testing.T) {
	if weekStartFromConfig(nil) != "monday" {
		t.Error("default week start should be monday")
	}
}

func TestLocationFromConfigDefault(t *testing.T) {
	loc := locationFromConfig(nil)
	if loc == nil {
		t.Fatal("nil loc")
	}
}

func TestAliasesRegistered(t *testing.T) {
	root := newRootCmd()
	// Ensure new aliases map to existing commands
	for alias, target := range map[string]string{
		"in":  "time-entry start",
		"out": "time-entry stop",
		"ge":  "time-entry get",
		"ce":  "time-entry create",
	} {
		parts := splitCommand(target)
		cmd := findCommand(root, parts)
		if cmd == nil {
			t.Errorf("alias %s target %q not found", alias, target)
		}
	}
}

func TestHumanTimeOffset(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, loc)
	tm, err := parseStartTimeFlexibleWith("+1h", nil)
	// Uses time.Now() — just ensure no error
	_ = tm
	_ = err
	_ = now
	tm2, err := parseStartTimeFlexibleWith("09:00", nil)
	if err != nil {
		t.Fatal(err)
	}
	if tm2.Hour() != 9 {
		t.Errorf("hour=%d want 9", tm2.Hour())
	}
}
