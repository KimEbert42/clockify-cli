package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kim/clockify-cli/pkg/clockify"
)

func TestTimeEntryHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "list") {
		t.Errorf("time-entry help should mention 'list', got:\n%s", output)
	}
	if !strings.Contains(output, "get") {
		t.Errorf("time-entry help should mention 'get', got:\n%s", output)
	}
	if !strings.Contains(output, "create") {
		t.Errorf("time-entry help should mention 'create', got:\n%s", output)
	}
	if !strings.Contains(output, "start") {
		t.Errorf("time-entry help should mention 'start', got:\n%s", output)
	}
	if !strings.Contains(output, "stop") {
		t.Errorf("time-entry help should mention 'stop', got:\n%s", output)
	}
}

func TestTimeEntryListJSON(t *testing.T) {
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			user := clockify.User{ID: "u1", Name: "Test User", Email: "test@example.com"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}

		if r.URL.Path != "/workspaces/ws123/user/u1/time-entries" {
			t.Errorf("expected path /workspaces/ws123/user/u1/time-entries, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		entries := []clockify.TimeEntry{
			{
				ID:          "te1",
				Description: "Working on feature",
				ProjectID:   "p1",
				UserID:      "u1",
				Billable:    true,
				TimeInterval: clockify.TimeInterval{
					Start:    start,
					End:      &end,
					Duration: 3600,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "list", "--start", "today", "--end", "today", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var entries []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestTimeEntryGetJSON(t *testing.T) {
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/time-entries/te1" {
			t.Errorf("expected path /workspaces/ws123/time-entries/te1, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		entry := clockify.TimeEntry{
			ID:          "te1",
			Description: "Working on feature",
			ProjectID:   "p1",
			UserID:      "u1",
			Billable:    true,
			TimeInterval: clockify.TimeInterval{
				Start:    start,
				End:      &end,
				Duration: 3600,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entry)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "get", "--id", "te1", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var entry clockify.TimeEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if entry.ID != "te1" {
		t.Errorf("expected ID te1, got %s", entry.ID)
	}
	if entry.Description != "Working on feature" {
		t.Errorf("expected description 'Working on feature', got %s", entry.Description)
	}
}

func TestTimeEntryGetNoID(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "get"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --id flag, got nil")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("expected 'must specify' error, got: %v", err)
	}
}

func TestTimeEntryCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			user := clockify.User{ID: "u1", Name: "Test User", Email: "test@example.com"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1" && r.Method == "GET" {
			proj := clockify.Project{ID: "p1", Name: "Project One", ClientID: "c1"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(proj)
			return
		}
		if r.Method != "POST" || r.URL.Path != "/workspaces/ws123/time-entries" {
			t.Errorf("expected POST /workspaces/ws123/time-entries, got %s %s", r.Method, r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var req clockify.TimeEntry
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		entry := clockify.TimeEntry{
			ID:          "te-new",
			Description: req.Description,
			ProjectID:   req.ProjectID,
			UserID:      "u1",
			Billable:    req.Billable,
			TimeInterval: clockify.TimeInterval{
				Start:    time.Now().Add(-1 * time.Hour),
				End:      &time.Time{},
				Duration: 3600,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(entry)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	setupTestConfig(t, "ws123")

	startTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"time-entry", "create",
		"--start", startTime,
		"--duration", "1h",
		"--project", "p1",
		"--description", "Test entry",
		"--billable",
		"--output", "json",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var entry clockify.TimeEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if entry.ID != "te-new" {
		t.Errorf("expected ID te-new, got %s", entry.ID)
	}
}

func TestTimeEntryCreateStartByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			user := clockify.User{ID: "u1", Name: "Test User"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects" && r.Method == "GET" {
			projects := []clockify.Project{
				{ID: "p-uuid-1", Name: "Website Redesign", ClientID: "c1"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(projects)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/workspaces/ws123/time-entries" {
			var req clockify.TimeEntry
			_ = json.NewDecoder(r.Body).Decode(&req)
			entry := clockify.TimeEntry{
				ID:          "te-named",
				Description: req.Description,
				ProjectID:   req.ProjectID,
				UserID:      "u1",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(entry)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)
	setupTestConfig(t, "ws123")

	startTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"time-entry", "create",
		"--start", startTime,
		"--duration", "1h",
		"--project", "Website Redesign",
		"--output", "json",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var entry clockify.TimeEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON: %v\nraw: %s", err, buf.String())
	}
	if entry.ProjectID != "p-uuid-1" {
		t.Errorf("expected resolved project p-uuid-1, got %q", entry.ProjectID)
	}
}

func TestTimeEntryCreateNoStart(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "create", "--duration", "1h"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --start flag, got nil")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("expected 'must specify' error, got: %v", err)
	}
}

func TestTimeEntryCreateNoDurationOrEnd(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "create", "--start", startTime})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --duration or --end flag, got nil")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("expected 'must specify' error, got: %v", err)
	}
}

func TestTimeEntryDeleteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/time-entries/te1" {
			if r.Method == "GET" {
				entry := clockify.TimeEntry{
					ID:          "te1",
					Description: "Test entry",
					TimeInterval: clockify.TimeInterval{
						Start: time.Now().Add(-1 * time.Hour),
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(entry)
				return
			}
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "delete", "--id", "te1", "--confirm"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got:\n%s", output)
	}
}

func TestTimeEntryStartJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			user := clockify.User{ID: "u1", Name: "Test User"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}

		if r.URL.Path == "/workspaces/ws123/projects/p1" && r.Method == "GET" {
			proj := clockify.Project{ID: "p1", Name: "Project One"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(proj)
			return
		}

		if r.URL.Path == "/workspaces/ws123/time-entries/status/in-progress" {
			entries := []clockify.TimeEntry{}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/workspaces/ws123/time-entries" {
			entry := clockify.TimeEntry{
				ID:          "te-running",
				Description: "Timer started",
				ProjectID:   "p1",
				UserID:      "u1",
				TimeInterval: clockify.TimeInterval{
					Start: time.Now(),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(entry)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "start", "--project", "p1", "--description", "Timer started", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "te-running") {
		t.Errorf("expected 'te-running' in output, got:\n%s", output)
	}
}

func TestTimeEntryStopJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			user := clockify.User{ID: "u1", Name: "Test User"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}

		if r.URL.Path == "/workspaces/ws123/time-entries/status/in-progress" {
			entries := []clockify.TimeEntry{
				{
					ID:          "te-running",
					Description: "Running timer",
					ProjectID:   "p1",
					UserID:      "u1",
					TimeInterval: clockify.TimeInterval{
						Start: time.Now().Add(-1 * time.Hour),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)
			return
		}

		if r.Method == "PATCH" && r.URL.Path == "/workspaces/ws123/user/u1/time-entries" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.URL.Path == "/workspaces/ws123/time-entries/te-running" {
			now := time.Now()
			start := now.Add(-1 * time.Hour)
			entry := clockify.TimeEntry{
				ID:          "te-running",
				Description: "Running timer",
				ProjectID:   "p1",
				UserID:      "u1",
				TimeInterval: clockify.TimeInterval{
					Start:    start,
					End:      &now,
					Duration: 3600,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entry)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "stop", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1h") {
		t.Errorf("expected '1h' duration in output, got:\n%s", output)
	}
}

func TestTimeEntryStopNoRunning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			user := clockify.User{ID: "u1", Name: "Test User"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}

		if r.URL.Path == "/workspaces/ws123/time-entries/status/in-progress" {
			entries := []clockify.TimeEntry{}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "stop"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no running timer, got nil")
	}
	if !strings.Contains(err.Error(), "no timer running") {
		t.Errorf("expected 'no timer running' error, got: %v", err)
	}
}

func TestTimeEntryListInProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			user := clockify.User{ID: "u1", Name: "Test User"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}

		if r.URL.Path == "/workspaces/ws123/time-entries/status/in-progress" {
			entries := []clockify.TimeEntry{
				{
					ID:          "te-running",
					Description: "Running timer",
					ProjectID:   "p1",
					UserID:      "u1",
					TimeInterval: clockify.TimeInterval{
						Start: time.Now().Add(-30 * time.Minute),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)

	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "list", "--in-progress", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var entries []clockify.TimeEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "te-running" {
		t.Errorf("expected ID te-running, got %s", entries[0].ID)
	}
}

func TestTimeEntryBulkNoInput(t *testing.T) {
	setupTestConfig(t, "ws123")

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "bulk"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing input, got nil")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("expected 'must specify' error, got: %v", err)
	}
}

func TestTimeEntryBulkInvalidJSON(t *testing.T) {
	// Create a temp file with invalid JSON
	tmpFile := t.TempDir() + "/invalid.json"
	os.WriteFile(tmpFile, []byte("not json"), 0644)

	setupTestConfig(t, "ws123")

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "bulk", "--file", tmpFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' error, got: %v", err)
	}
}

func TestTimeEntryBulkEmptyArray(t *testing.T) {
	tmpFile := t.TempDir() + "/empty.json"
	os.WriteFile(tmpFile, []byte("[]"), 0644)

	setupTestConfig(t, "ws123")

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-entry", "bulk", "--file", tmpFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty array, got nil")
	}
	if !strings.Contains(err.Error(), "no time entries provided") {
		t.Errorf("expected 'no time entries provided' error, got: %v", err)
	}
}
