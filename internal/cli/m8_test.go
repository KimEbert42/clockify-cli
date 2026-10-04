package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kim/clockify-cli/internal/config"
	"github.com/kim/clockify-cli/pkg/clockify"
)

func TestListCacheServeAndNoCache(t *testing.T) {
	dir := t.TempDir()
	isolateCacheEnv(t, dir)
	setupTestConfig(t, "ws-cache")

	projects := []clockify.Project{
		{ID: "p1", Name: "Cached Project", Status: "ACTIVE"},
	}
	if err := config.WriteTTL("default", listCacheKey("ws-cache", compResourceProjects), projects); err != nil {
		t.Fatalf("seed list cache: %v", err)
	}

	// Served from cache without hitting API
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"project", "list", "--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("project list from cache error = %v", err)
	}
	var got []clockify.Project
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("parse cache output: %v raw=%s", err, buf.String())
	}
	if len(got) != 1 || got[0].Name != "Cached Project" {
		t.Fatalf("expected cached project, got %+v", got)
	}

	// --no-cache bypasses cache and should fail without a live API
	buf2 := new(bytes.Buffer)
	cmd2 := newRootCmd()
	cmd2.SetOut(buf2)
	cmd2.SetErr(buf2)
	cmd2.SetArgs([]string{"project", "list", "--output", "json", "--no-cache"})
	if err := cmd2.Execute(); err == nil {
		t.Fatal("expected --no-cache project list to fail without API")
	}
}

func TestCompletionRefreshWritesNames(t *testing.T) {
	dir := t.TempDir()
	isolateCacheEnv(t, dir)
	setupTestConfig(t, "ws-refresh")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/workspaces":
			_ = json.NewEncoder(w).Encode([]clockify.Workspace{{ID: "ws-refresh", Name: "Main WS"}})
		case "/workspaces/ws-refresh/projects":
			_ = json.NewEncoder(w).Encode([]clockify.Project{{ID: "p1", Name: "Alpha"}})
		case "/workspaces/ws-refresh/clients":
			_ = json.NewEncoder(w).Encode([]clockify.Client{{ID: "c1", Name: "Acme"}})
		case "/workspaces/ws-refresh/tags":
			_ = json.NewEncoder(w).Encode([]clockify.Tag{{ID: "t1", Name: "urgent"}})
		case "/workspaces/ws-refresh/users":
			_ = json.NewEncoder(w).Encode([]clockify.User{{ID: "u1", Name: "Ada", Email: "ada@example.com"}})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion", "refresh", "--quiet"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("completion refresh error = %v", err)
	}

	got, _ := completeCachedNames(newRootCmd(), "ws-refresh", compResourceProjects, "Al")
	if len(got) != 1 || got[0] != "Alpha" {
		t.Fatalf("expected Alpha from refresh cache, got %v", got)
	}

	gotUsers, _ := completeCachedNames(newRootCmd(), "ws-refresh", compResourceUsers, "Ada")
	if len(gotUsers) != 1 || gotUsers[0] != "Ada" {
		t.Fatalf("expected Ada from refresh cache, got %v", gotUsers)
	}
}

func TestTimeEntryCreateRequiresStart(t *testing.T) {
	dir := t.TempDir()
	isolateCacheEnv(t, dir)
	setupTestConfig(t, "ws123")

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"time-entry", "create", "--duration", "1h"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when --start missing")
	}
}

func TestColorDisabledEnv(t *testing.T) {
	cmd := newRootCmd()
	if colorDisabled(cmd) {
		t.Error("expected false without NO_COLOR")
	}
	t.Setenv("NO_COLOR", "1")
	if !colorDisabled(cmd) {
		t.Error("expected true with NO_COLOR")
	}
}

func TestProfileFlagCompletion(t *testing.T) {
	dir := t.TempDir()
	isolateCacheEnv(t, dir)
	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
profiles:
  work:
    region: eu-de
`), 0600)

	cmd := newRootCmd()
	got, _ := completeProfileFlag(cmd, "w")
	found := false
	for _, n := range got {
		if n == "work" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected work in profile completions, got %v", got)
	}
}

func TestProjectListPopulatesCompletionCache(t *testing.T) {
	dir := t.TempDir()
	isolateCacheEnv(t, dir)
	setupTestConfig(t, "ws-pop")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws-pop/projects" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]clockify.Project{
			{ID: "p1", Name: "Website"},
			{ID: "p2", Name: "Mobile"},
		})
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"project", "list", "--output", "json", "--quiet"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("project list error = %v", err)
	}

	got, _ := completeCachedNames(newRootCmd(), "ws-pop", compResourceProjects, "Web")
	if len(got) != 1 || got[0] != "Website" {
		t.Fatalf("expected Website after list populated cache, got %v", got)
	}
}
