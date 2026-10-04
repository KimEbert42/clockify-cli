package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/kim/clockify-cli/pkg/duration"
)

// parseDateTimeCLI accepts relative dates, YYYY-MM-DD, or RFC3339/ISO8601.
func parseDateTimeCLI(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty date/time value")
	}
	if duration.IsRelativeDate(s) {
		t, err := duration.ParseRelativeDate(s, "monday")
		if err != nil {
			return "", err
		}
		return t.Format("2006-01-02T15:04:05.000000Z"), nil
	}
	// YYYY-MM-DD
	if len(s) == 10 {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return "", fmt.Errorf("invalid date %q (use YYYY-MM-DD or relative date)", s)
		}
		return t.Format("2006-01-02T15:04:05.000000Z"), nil
	}
	// Already ISO8601-ish
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format("2006-01-02T15:04:05.000000Z"), nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t.UTC().Format("2006-01-02T15:04:05.000000Z"), nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05Z", s); err == nil {
		return t.UTC().Format("2006-01-02T15:04:05.000000Z"), nil
	}
	return "", fmt.Errorf("invalid date/time %q (use YYYY-MM-DD, relative date, or ISO8601)", s)
}

func parseDateOnlyCLI(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty date value")
	}
	if duration.IsRelativeDate(s) {
		t, err := duration.ParseRelativeDate(s, "monday")
		if err != nil {
			return "", err
		}
		return t.Format("2006-01-02"), nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.Format("2006-01-02"), nil
	}
	if full, err := parseDateTimeCLI(s); err == nil {
		if t, err := time.Parse("2006-01-02T15:04:05.000000Z", full); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("invalid date %q (use YYYY-MM-DD or relative date)", s)
}

// parseHoursPerDay converts duration strings like "8h" or "480" (minutes) or "8" to hours float.
func parseHoursPerDay(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty hours value")
	}
	// Pure number → hours
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}
	secs, err := duration.Parse(s)
	if err != nil {
		return 0, fmt.Errorf("invalid hours-per-day %q (use hours like 8h, or a number): %w", s, err)
	}
	hours := float64(secs) / 3600.0
	return hours, nil
}

func parseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", s, err)
	}
	return f, nil
}

func readJSONInput(filePath string, stdin bool, out any) error {
	var inputData []byte
	var err error
	if filePath != "" {
		inputData, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
	} else if stdin {
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else {
		return nil
	}
	if err := json.Unmarshal(inputData, out); err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}
	return nil
}

func confirmDestructive(cmd interface {
	Println(a ...any)
	Printf(format string, a ...any)
	Print(a ...any)
}, confirm bool, kind, id, name string) bool {
	if confirm {
		return true
	}
	cmd.Printf("Delete %s %s", kind, id)
	if name != "" {
		cmd.Printf(" (%s)", name)
	}
	cmd.Println("?")
	cmd.Print("Are you sure? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(strings.TrimSpace(response)) != "y" {
		cmd.Println("Aborted.")
		return false
	}
	return true
}

func newM7Client(profile string) (*api.ClockifyClient, *clockifyConfigView, error) {
	apiKey, region, cfg, err := loadAuthWithConfig(profile)
	if err != nil {
		return nil, nil, err
	}
	if cfg.WorkspaceID == "" {
		return nil, nil, fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
	}
	return api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID), &clockifyConfigView{WorkspaceID: cfg.WorkspaceID}, nil
}

type clockifyConfigView struct {
	WorkspaceID string
}

func resolveClientRef(ctx context.Context, client *api.ClockifyClient, ref string) (string, error) {
	if c, err := client.GetClient(ctx, ref); err == nil && c.ID != "" {
		return c.ID, nil
	}
	c, err := resolveClientByName(ctx, client, ref)
	if err != nil {
		return "", fmt.Errorf("no client found matching %q", ref)
	}
	return c.ID, nil
}

func resolveExpenseCategory(ctx context.Context, client *api.ClockifyClient, ref string) (string, error) {
	if ref == "" {
		return "", nil
	}
	cats, err := client.ListExpenseCategories(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list expense categories: %w", err)
	}
	for _, cat := range cats {
		if cat.ID == ref || strings.EqualFold(cat.Name, ref) {
			return cat.ID, nil
		}
	}
	var matches []clockify.ExpenseCategory
	for _, cat := range cats {
		if containsIgnoreCase(cat.Name, ref) {
			matches = append(matches, cat)
		}
	}
	if len(matches) == 1 {
		return matches[0].ID, nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple expense categories match %q; use the category ID", ref)
	}
	return "", fmt.Errorf("no expense category found matching %q (list categories in the Clockify UI)", ref)
}

func resolvePolicyRef(ctx context.Context, client *api.ClockifyClient, ref string) (string, error) {
	policies, err := client.ListTimeOffPolicies(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list time-off policies: %w", err)
	}
	for _, p := range policies {
		if p.ID == ref || strings.EqualFold(p.Name, ref) {
			return p.ID, nil
		}
	}
	// Type label match (VACATION, SICK, ...)
	upper := strings.ToUpper(strings.TrimSpace(ref))
	if err := clockify.ValidateTimeOffType(upper); err == nil {
		for _, p := range policies {
			if strings.Contains(strings.ToUpper(p.Name), upper) {
				return p.ID, nil
			}
		}
	}
	var matches []clockify.TimeOffPolicy
	for _, p := range policies {
		if containsIgnoreCase(p.Name, ref) {
			matches = append(matches, p)
		}
	}
	if len(matches) == 1 {
		return matches[0].ID, nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple time-off policies match %q; use --policy with the policy ID", ref)
	}
	return "", fmt.Errorf("no time-off policy found matching %q; use --policy with a policy ID from the Clockify UI", ref)
}

func resolveGroupByName(ctx context.Context, client *api.ClockifyClient, name string) (*clockify.UserGroup, error) {
	groups, err := client.ListUserGroups(ctx, api.ListUserGroupsOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	for _, g := range groups {
		if g.Name == name {
			return &g, nil
		}
	}
	var fuzzy []clockify.UserGroup
	for _, g := range groups {
		if containsIgnoreCase(g.Name, name) {
			fuzzy = append(fuzzy, g)
		}
	}
	if len(fuzzy) == 1 {
		return &fuzzy[0], nil
	}
	if len(fuzzy) > 1 {
		return nil, fmt.Errorf("multiple groups match %q; use --id to disambiguate", name)
	}
	return nil, fmt.Errorf("no group found matching %q", name)
}

func resolveCustomFieldByName(ctx context.Context, client *api.ClockifyClient, name string) (*clockify.CustomField, error) {
	fields, err := client.ListCustomFields(ctx, api.ListCustomFieldsOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to list custom fields: %w", err)
	}
	for _, f := range fields {
		if f.Name == name {
			return &f, nil
		}
	}
	var fuzzy []clockify.CustomField
	for _, f := range fields {
		if containsIgnoreCase(f.Name, name) {
			fuzzy = append(fuzzy, f)
		}
	}
	if len(fuzzy) == 1 {
		return &fuzzy[0], nil
	}
	if len(fuzzy) > 1 {
		return nil, fmt.Errorf("multiple custom fields match %q; use --id to disambiguate", name)
	}
	return nil, fmt.Errorf("no custom field found matching %q", name)
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func invoiceStatusDisplay(inv clockify.Invoice) string {
	if inv.InvoiceStatus != "" {
		return inv.InvoiceStatus
	}
	return inv.Status
}

func printPaginationSummary(cmd interface {
	Println(a ...any)
	Printf(format string, a ...any)
}, quiet bool, outputFormat string, pageInfo *clockify.PageInfo) {
	if quiet || pageInfo == nil || pageInfo.TotalPages <= 1 {
		return
	}
	if outputFormat == "json" || outputFormat == "csv" || outputFormat == "yaml" {
		return
	}
	cmd.Printf("\nPage %d of %d (%d results)\n", pageInfo.Page, pageInfo.TotalPages, pageInfo.TotalResults)
}
