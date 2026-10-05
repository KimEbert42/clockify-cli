package resolve

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kim/clockify-cli/internal/api"
)

func TestIsUUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"550e8400e29b41d4a716446655440000", true},
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"abc123", false},
		{"", false},
		{"550e8400-e29b-41d4-a716-44665544000", false}, // 31 hex chars
	}
	for _, tt := range tests {
		if got := isUUID(tt.input); got != tt.want {
			t.Errorf("isUUID(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestResolveInput_UUIDPassthrough(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	called := false
	listFn := func(ctx context.Context, client *api.ClockifyClient, workspaceID string) ([]map[string]any, error) {
		called = true
		return nil, nil
	}
	got, err := ResolveInput(context.Background(), nil, "ws", id, listFn)
	if err != nil {
		t.Fatalf("ResolveInput() error = %v", err)
	}
	if got != id {
		t.Errorf("got %q, want %q", got, id)
	}
	if called {
		t.Error("listFn should not be called for UUID input")
	}
}

func TestResolveInput_PrefixMatch(t *testing.T) {
	listFn := func(ctx context.Context, client *api.ClockifyClient, workspaceID string) ([]map[string]any, error) {
		return []map[string]any{
			{"id": "abc123def456", "name": "Website"},
			{"id": "zzz999", "name": "Other"},
		}, nil
	}
	got, err := ResolveInput(context.Background(), nil, "ws", "abc1", listFn)
	if err != nil {
		t.Fatalf("ResolveInput() error = %v", err)
	}
	if got != "abc123def456" {
		t.Errorf("got %q, want abc123def456", got)
	}
}

func TestResolveInput_ExactName(t *testing.T) {
	listFn := func(ctx context.Context, client *api.ClockifyClient, workspaceID string) ([]map[string]any, error) {
		return []map[string]any{
			{"id": "id-1", "name": "Website Redesign"},
			{"id": "id-2", "name": "Mobile App"},
		}, nil
	}
	got, err := ResolveInput(context.Background(), nil, "ws", "website redesign", listFn)
	if err != nil {
		t.Fatalf("ResolveInput() error = %v", err)
	}
	if got != "id-1" {
		t.Errorf("got %q, want id-1", got)
	}
}

func TestResolveInput_ContainsName(t *testing.T) {
	listFn := func(ctx context.Context, client *api.ClockifyClient, workspaceID string) ([]map[string]any, error) {
		return []map[string]any{
			{"id": "id-1", "name": "Website Redesign"},
			{"id": "id-2", "name": "Mobile App"},
		}, nil
	}
	got, err := ResolveInput(context.Background(), nil, "ws", "redes", listFn)
	if err != nil {
		t.Fatalf("ResolveInput() error = %v", err)
	}
	if got != "id-1" {
		t.Errorf("got %q, want id-1", got)
	}
}

func TestResolveInput_NotFound(t *testing.T) {
	listFn := func(ctx context.Context, client *api.ClockifyClient, workspaceID string) ([]map[string]any, error) {
		return []map[string]any{{"id": "id-1", "name": "Website"}}, nil
	}
	_, err := ResolveInput(context.Background(), nil, "ws", "missing", listFn)
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if !strings.Contains(err.Error(), "no resource found") {
		t.Errorf("error = %v, want no resource found", err)
	}
}

func TestResolveInput_MultipleMatches(t *testing.T) {
	listFn := func(ctx context.Context, client *api.ClockifyClient, workspaceID string) ([]map[string]any, error) {
		return []map[string]any{
			{"id": "id-1", "name": "Website A"},
			{"id": "id-2", "name": "Website B"},
		}, nil
	}
	_, err := ResolveInput(context.Background(), nil, "ws", "website", listFn)
	if err == nil {
		t.Fatal("expected multiple-match error")
	}
	if !strings.Contains(err.Error(), "multiple matches") {
		t.Errorf("error = %v, want multiple matches", err)
	}
	if !strings.Contains(err.Error(), "id-1") || !strings.Contains(err.Error(), "id-2") {
		t.Errorf("error should list candidates: %v", err)
	}
}

func TestResolveInput_ListError(t *testing.T) {
	want := errors.New("api down")
	listFn := func(ctx context.Context, client *api.ClockifyClient, workspaceID string) ([]map[string]any, error) {
		return nil, want
	}
	_, err := ResolveInput(context.Background(), nil, "ws", "name", listFn)
	if err == nil {
		t.Fatal("expected list error")
	}
	if !strings.Contains(err.Error(), "failed to list resources") {
		t.Errorf("error = %v, want failed to list resources", err)
	}
}
