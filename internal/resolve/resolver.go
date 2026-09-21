package resolve

import (
	"context"
	"fmt"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
)

type Resolver interface {
	Resolve(ctx context.Context, client *api.ClockifyClient, workspaceID string, input string) (string, error)
}

func ResolveInput(ctx context.Context, client *api.ClockifyClient, workspaceID string, input string, listFn func(ctx context.Context, client *api.ClockifyClient, workspaceID string) ([]map[string]any, error)) (string, error) {
	if isUUID(input) {
		return input, nil
	}

	items, err := listFn(ctx, client, workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to list resources: %w", err)
	}

	for _, item := range items {
		if id, ok := item["id"].(string); ok {
			if strings.HasPrefix(id, input) {
				return id, nil
			}
		}
	}

	var matches []map[string]any
	for _, item := range items {
		if name, ok := item["name"].(string); ok {
			if strings.EqualFold(name, input) || strings.Contains(strings.ToLower(name), strings.ToLower(input)) {
				matches = append(matches, item)
			}
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no resource found matching %q", input)
	case 1:
		return matches[0]["id"].(string), nil
	default:
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = fmt.Sprintf("%s (%s)", m["name"], m["id"])
		}
		return "", fmt.Errorf("multiple matches for %q:\n%s\nUse the full ID to disambiguate", input, strings.Join(names, "\n"))
	}
}

func isUUID(s string) bool {
	s = strings.ReplaceAll(s, "-", "")
	return len(s) == 32
}
