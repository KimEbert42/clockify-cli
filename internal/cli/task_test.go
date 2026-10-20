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

// serveProject resolves either project list or get-by-ID requests.
func serveProject(w http.ResponseWriter, r *http.Request, projects []clockify.Project) bool {
	if r.URL.Path == "/workspaces/ws123/projects" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
		return true
	}
	// GET /workspaces/ws123/projects/{id}
	for _, p := range projects {
		if r.URL.Path == "/workspaces/ws123/projects/"+p.ID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p)
			return true
		}
	}
	return false
}

func TestTaskHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"task", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	for _, sub := range []string{"list", "create", "update", "delete"} {
		if !strings.Contains(output, sub) {
			t.Errorf("task help should mention %q, got:\n%s", sub, output)
		}
	}
}

func TestTaskListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks" {
			tasks := []clockify.Task{
				{ID: "t1", Name: "Design", ProjectID: "p1", Status: "TODO", Active: true},
				{ID: "t2", Name: "Develop", ProjectID: "p1", Status: "IN_PROGRESS", Active: true},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tasks)
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
	cmd.SetArgs([]string{"task", "list", "--project", "p1", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var tasks []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0]["id"] != "t1" {
		t.Errorf("expected first task ID t1, got %v", tasks[0]["id"])
	}
	if tasks[1]["name"] != "Develop" {
		t.Errorf("expected second task name Develop, got %v", tasks[1]["name"])
	}
}

func TestTaskListIsActive(t *testing.T) {
	var gotIsActive string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks" {
			gotIsActive = r.URL.Query().Get("is-active")
			tasks := []clockify.Task{
				{ID: "t1", Name: "Design", ProjectID: "p1", Status: "TODO", Active: true},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tasks)
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
	cmd.SetArgs([]string{"task", "list", "--project", "p1", "--is-active", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if gotIsActive != "true" {
		t.Errorf("expected is-active=true query param, got %q", gotIsActive)
	}
}

func TestTaskListMissingProject(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"task", "list"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --project, got nil")
	}
	if !strings.Contains(err.Error(), "must specify --project") {
		t.Errorf("expected 'must specify --project' error, got: %v", err)
	}
}

func TestTaskListByProjectName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{
			{ID: "p1", Name: "Website Project"},
			{ID: "p2", Name: "Mobile App"},
		}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks" {
			tasks := []clockify.Task{
				{ID: "t1", Name: "Design", ProjectID: "p1", Status: "TODO", Active: true},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tasks)
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
	cmd.SetArgs([]string{"task", "list", "--project", "Website", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var tasks []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0]["projectId"] != "p1" {
		t.Errorf("expected projectId p1, got %v", tasks[0]["projectId"])
	}
}

func TestTaskCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks" && r.Method == "POST" {
			var req clockify.CreateTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request body: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if req.Name != "Design Phase" {
				t.Errorf("expected name Design Phase, got %s", req.Name)
			}
			if req.Status != "ACTIVE" {
				t.Errorf("expected status ACTIVE (TODO maps to ACTIVE), got %s", req.Status)
			}
			if req.Estimate != "PT1H" {
				t.Errorf("expected estimate PT1H, got %q", req.Estimate)
			}
			task := clockify.Task{ID: "t1", Name: req.Name, ProjectID: "p1", Status: req.Status, Estimate: 3600, Active: true}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
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
	cmd.SetArgs([]string{"task", "create", "--project", "p1", "--name", "Design Phase", "--status", "TODO", "--estimate", "3600", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var task clockify.Task
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}
	if task.ID != "t1" {
		t.Errorf("expected ID t1, got %s", task.ID)
	}
	if task.Name != "Design Phase" {
		t.Errorf("expected name Design Phase, got %s", task.Name)
	}
}

func TestTaskCreateEstimateDurationFormat(t *testing.T) {
	var gotEstimate string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks" && r.Method == "POST" {
			var req clockify.CreateTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request body: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if req.Estimate != "" {
				gotEstimate = req.Estimate
			}
			task := clockify.Task{ID: "t1", Name: req.Name, ProjectID: "p1", Estimate: 3600, Active: true}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
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
	cmd.SetArgs([]string{"task", "create", "--project", "p1", "--name", "Design", "--estimate", "1h", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if gotEstimate != "PT1H" {
		t.Errorf("expected estimate PT1H for 1h, got %q", gotEstimate)
	}
}

func TestTaskCreateAssigneeByName(t *testing.T) {
	var gotAssignees []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/users" && r.Method == "GET" {
			users := []clockify.User{
				{ID: "u1", Name: "Alice Smith", Email: "alice@example.com"},
				{ID: "u2", Name: "Bob Jones", Email: "bob@example.com"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks" && r.Method == "POST" {
			var req clockify.CreateTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request body: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			gotAssignees = req.AssigneeIDs
			task := clockify.Task{ID: "t1", Name: req.Name, ProjectID: "p1", AssigneeIDs: req.AssigneeIDs, Active: true}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
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
	cmd.SetArgs([]string{"task", "create", "--project", "p1", "--name", "Design", "--assignee", "Alice", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(gotAssignees) != 1 || gotAssignees[0] != "u1" {
		t.Errorf("expected assignee IDs [u1], got %v", gotAssignees)
	}
}

func TestTaskCreateMissingName(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"task", "create", "--project", "p1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --name, got nil")
	}
	if !strings.Contains(err.Error(), "must specify --name") {
		t.Errorf("expected 'must specify --name' error, got: %v", err)
	}
}

func TestTaskCreateInvalidStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("API should not be called for task create when status is invalid")
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
	cmd.SetArgs([]string{"task", "create", "--project", "p1", "--name", "Design", "--status", "INVALID"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("expected 'invalid status' error, got: %v", err)
	}
}

func TestTaskCreateIfNotExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks" {
			if r.Method == "GET" {
				tasks := []clockify.Task{
					{ID: "t1", Name: "Existing Task", ProjectID: "p1", Status: "TODO", Active: true},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tasks)
				return
			}
			if r.Method == "POST" {
				t.Error("POST should not be called when --if-not-exists and task exists")
				http.Error(w, "should not create", http.StatusBadRequest)
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
	cmd.SetArgs([]string{"task", "create", "--project", "p1", "--name", "Existing Task", "--if-not-exists", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already exists") {
		t.Errorf("expected 'already exists', got:\n%s", output)
	}
}

func TestTaskCreateFromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks" && r.Method == "POST" {
			var req clockify.CreateTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request body: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if req.Name != "From File" {
				t.Errorf("expected name From File, got %s", req.Name)
			}
			if req.Status != "DONE" {
				t.Errorf("expected status DONE, got %s", req.Status)
			}
			task := clockify.Task{ID: "t1", Name: req.Name, ProjectID: "p1", Status: req.Status, Active: true}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)
	setupTestConfig(t, "ws123")

	dir := t.TempDir()
	path := filepath.Join(dir, "task.json")
	if err := os.WriteFile(path, []byte(`{"name":"From File","status":"DONE"}`), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"task", "create", "--project", "p1", "--name", "Ignored", "--file", path, "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var task clockify.Task
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}
	if task.Name != "From File" {
		t.Errorf("expected name From File (file should win over flag), got %s", task.Name)
	}
}

func TestTaskUpdateStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks/t1" {
			if r.Method == "GET" {
				task := clockify.Task{ID: "t1", Name: "Design", ProjectID: "p1", Status: "TODO", Active: true}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(task)
				return
			}
			if r.Method == "PUT" {
				var req clockify.UpdateTaskRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request body: %v", err)
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				if req.Status != "DONE" {
					t.Errorf("expected status DONE, got %s", req.Status)
				}
				task := clockify.Task{ID: "t1", Name: req.Name, ProjectID: "p1", Status: req.Status, Active: true}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(task)
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
	cmd.SetArgs([]string{"task", "update", "--project", "p1", "--id", "t1", "--status", "DONE", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var task clockify.Task
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}
	if task.Status != "DONE" {
		t.Errorf("expected status DONE, got %s", task.Status)
	}
}

func TestTaskUpdateFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks/t1" {
			if r.Method == "GET" {
				task := clockify.Task{ID: "t1", Name: "Old Name", ProjectID: "p1", Status: "TODO", Active: true}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(task)
				return
			}
			if r.Method == "PUT" {
				var req clockify.UpdateTaskRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request body: %v", err)
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				if req.Name != "Renamed Task" {
					t.Errorf("expected name Renamed Task, got %s", req.Name)
				}
				task := clockify.Task{ID: "t1", Name: req.Name, ProjectID: "p1", Status: req.Status, Active: true}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(task)
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

	dir := t.TempDir()
	path := filepath.Join(dir, "task.json")
	if err := os.WriteFile(path, []byte(`{"name":"Renamed Task"}`), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"task", "update", "--project", "p1", "--id", "t1", "--file", path, "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var task clockify.Task
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}
	if task.Name != "Renamed Task" {
		t.Errorf("expected name Renamed Task, got %s", task.Name)
	}
}

func TestTaskDeleteConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks/t1" {
			if r.Method == "GET" {
				task := clockify.Task{ID: "t1", Name: "Design", ProjectID: "p1", Status: "TODO", Active: true}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(task)
				return
			}
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusOK)
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
	cmd.SetArgs([]string{"task", "delete", "--project", "p1", "--id", "t1", "--confirm"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Task t1 deleted") {
		t.Errorf("expected 'Task t1 deleted', got:\n%s", output)
	}
}

func TestTaskDeleteNoConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{{ID: "p1", Name: "Website"}}
		if serveProject(w, r, projects) {
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1/tasks/t1" && r.Method == "GET" {
			task := clockify.Task{ID: "t1", Name: "Design", ProjectID: "p1", Status: "TODO", Active: true}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
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
	cmd.SetArgs([]string{"task", "delete", "--project", "p1", "--id", "t1"})

	// stdin is empty in tests, so Scanln fails and command aborts
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Aborted") && !strings.Contains(output, "deleted") {
		t.Errorf("expected 'Aborted' or 'deleted', got:\n%s", output)
	}
}

func TestTaskCreateMissingProject(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"task", "create", "--name", "Design"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --project, got nil")
	}
	if !strings.Contains(err.Error(), "must specify --project") {
		t.Errorf("expected 'must specify --project' error, got: %v", err)
	}
}
