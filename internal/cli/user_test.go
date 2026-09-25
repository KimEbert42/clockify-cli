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

func TestUserGetHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"user", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "get") {
		t.Errorf("user help should mention 'get', got:\n%s", output)
	}
	if !strings.Contains(output, "list") {
		t.Errorf("user help should mention 'list', got:\n%s", output)
	}
	if !strings.Contains(output, "info") {
		t.Errorf("user help should mention 'info', got:\n%s", output)
	}
}

func TestUserGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Errorf("expected path /user, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		user := clockify.User{
			ID:    "user123",
			Email: "test@example.com",
			Name:  "Test User",
			Status: "ACTIVE",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"user", "get", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var user clockify.User
	if err := json.Unmarshal(buf.Bytes(), &user); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if user.ID != "user123" {
		t.Errorf("expected ID user123, got %s", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
}

func TestUserListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/users" {
			t.Errorf("expected path /workspaces/ws123/users, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		users := []clockify.User{
			{ID: "u1", Name: "User 1", Email: "user1@example.com", Status: "ACTIVE"},
			{ID: "u2", Name: "User 2", Email: "user2@example.com", Status: "INACTIVE"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_WORKSPACE_ID", "ws123")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	cfgPath := filepath.Join(t.TempDir(), "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
  workspace_id: ws123
`), 0600)
	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"user", "list", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var users []clockify.User
	if err := json.Unmarshal(buf.Bytes(), &users); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Status != "ACTIVE" {
		t.Errorf("expected first user ACTIVE, got %s", users[0].Status)
	}
}

func TestUserInfoByEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users := []clockify.User{
			{ID: "u1", Name: "Alice Smith", Email: "alice@example.com", Status: "ACTIVE"},
			{ID: "u2", Name: "Bob Jones", Email: "bob@example.com", Status: "ACTIVE"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_WORKSPACE_ID", "ws123")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	cfgPath := filepath.Join(t.TempDir(), "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
  workspace_id: ws123
`), 0600)
	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"user", "info", "--email", "alice@example.com", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var user clockify.User
	if err := json.Unmarshal(buf.Bytes(), &user); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if user.ID != "u1" {
		t.Errorf("expected ID u1, got %s", user.ID)
	}
	if user.Name != "Alice Smith" {
		t.Errorf("expected name Alice Smith, got %s", user.Name)
	}
}

func TestUserInfoByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users := []clockify.User{
			{ID: "u1", Name: "Alice Smith", Email: "alice@example.com", Status: "ACTIVE"},
			{ID: "u2", Name: "Alice Johnson", Email: "alicej@example.com", Status: "ACTIVE"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_WORKSPACE_ID", "ws123")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	cfgPath := filepath.Join(t.TempDir(), "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
  workspace_id: ws123
`), 0600)
	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"user", "info", "--name", "Alice", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var users []clockify.User
	if err := json.Unmarshal(buf.Bytes(), &users); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users matching 'Alice', got %d", len(users))
	}
}

func TestUserInfoNoFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"user", "info"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing flags, got nil")
	}
	if !strings.Contains(err.Error(), "--email") && !strings.Contains(err.Error(), "--name") {
		t.Errorf("expected error about --email or --name, got: %v", err)
	}
}

func TestUserListWithStatusFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users := []clockify.User{
			{ID: "u1", Name: "Active User", Email: "active@example.com", Status: "ACTIVE"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_WORKSPACE_ID", "ws123")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	cfgPath := filepath.Join(t.TempDir(), "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
  workspace_id: ws123
`), 0600)
	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(filepath.Dir(cfgPath)))

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"user", "list", "--status", "ACTIVE", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var users []clockify.User
	if err := json.Unmarshal(buf.Bytes(), &users); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Status != "ACTIVE" {
		t.Errorf("expected user ACTIVE, got %s", users[0].Status)
	}
}
