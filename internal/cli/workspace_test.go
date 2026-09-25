package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kim/clockify-cli/pkg/clockify"
)

func setupTestConfig(t *testing.T, workspaceID string) {
	t.Helper()
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	t.Cleanup(func() { os.Unsetenv("XDG_CONFIG_HOME") })

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
  workspace_id: `+workspaceID+`
`), 0600)
}

func TestWorkspaceListHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"workspace", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "list") {
		t.Errorf("workspace help should mention 'list', got:\n%s", output)
	}
	if !strings.Contains(output, "get") {
		t.Errorf("workspace help should mention 'get', got:\n%s", output)
	}
}

func TestWorkspaceListNoConfig(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
`), 0600)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"workspace", "list"})

	err := cmd.Execute()
	if err != nil {
		if !strings.Contains(err.Error(), "API key") {
			t.Fatalf("expected API key error, got: %v", err)
		}
		return
	}
}

func TestWorkspaceListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces" {
			t.Errorf("expected path /workspaces, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		workspaces := []clockify.Workspace{
			{ID: "ws1", Name: "Workspace 1", Currency: "USD"},
			{ID: "ws2", Name: "Workspace 2", Currency: "EUR"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaces)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"workspace", "list", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var workspaces []clockify.Workspace
	if err := json.Unmarshal(buf.Bytes(), &workspaces); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(workspaces))
	}
	if workspaces[0].ID != "ws1" {
		t.Errorf("expected first workspace ID ws1, got %s", workspaces[0].ID)
	}
	if workspaces[1].Name != "Workspace 2" {
		t.Errorf("expected second workspace name Workspace 2, got %s", workspaces[1].Name)
	}
}

func TestWorkspaceGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123" {
			t.Errorf("expected path /workspaces/ws123, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		ws := clockify.Workspace{ID: "ws123", Name: "Test Workspace", Currency: "USD"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"workspace", "get", "--id", "ws123", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var ws clockify.Workspace
	if err := json.Unmarshal(buf.Bytes(), &ws); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if ws.ID != "ws123" {
		t.Errorf("expected ID ws123, got %s", ws.ID)
	}
	if ws.Name != "Test Workspace" {
		t.Errorf("expected name Test Workspace, got %s", ws.Name)
	}
}

func TestWorkspaceGetByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workspaces := []clockify.Workspace{
			{ID: "ws1", Name: "Website Project", Currency: "USD"},
			{ID: "ws2", Name: "Mobile App", Currency: "EUR"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaces)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"workspace", "get", "--name", "Website Project", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var ws clockify.Workspace
	if err := json.Unmarshal(buf.Bytes(), &ws); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if ws.ID != "ws1" {
		t.Errorf("expected ID ws1, got %s", ws.ID)
	}
}

func TestWorkspaceGetByNameMultipleMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workspaces := []clockify.Workspace{
			{ID: "ws1", Name: "App Mobile", Currency: "USD"},
			{ID: "ws2", Name: "App Web", Currency: "EUR"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaces)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"workspace", "get", "--name", "App"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for multiple matches, got nil")
	}
	if !strings.Contains(err.Error(), "multiple") {
		t.Errorf("expected 'multiple' error, got: %v", err)
	}
}

func TestWorkspaceGetByNameNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workspaces := []clockify.Workspace{
			{ID: "ws1", Name: "Workspace 1", Currency: "USD"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaces)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"workspace", "get", "--name", "Nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !strings.Contains(err.Error(), "no workspace found") {
		t.Errorf("expected 'no workspace found' error, got: %v", err)
	}
}
