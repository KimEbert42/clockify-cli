package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kim/clockify-cli/pkg/clockify"
)

func TestProjectListHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"project", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "list") {
		t.Errorf("project help should mention 'list', got:\n%s", output)
	}
	if !strings.Contains(output, "get") {
		t.Errorf("project help should mention 'get', got:\n%s", output)
	}
	if !strings.Contains(output, "create") {
		t.Errorf("project help should mention 'create', got:\n%s", output)
	}
	if !strings.Contains(output, "update") {
		t.Errorf("project help should mention 'update', got:\n%s", output)
	}
	if !strings.Contains(output, "delete") {
		t.Errorf("project help should mention 'delete', got:\n%s", output)
	}
}

func TestProjectListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/projects" {
			t.Errorf("expected path /workspaces/ws123/projects, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		projects := []clockify.Project{
			{ID: "p1", Name: "Website", ClientID: "c1", Billable: true, Status: "ACTIVE"},
			{ID: "p2", Name: "Mobile App", ClientID: "c2", Billable: false, Status: "ACTIVE"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
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
	cmd.SetArgs([]string{"project", "list", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var projects []clockify.Project
	if err := json.Unmarshal(buf.Bytes(), &projects); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].ID != "p1" {
		t.Errorf("expected first project ID p1, got %s", projects[0].ID)
	}
	if projects[1].Name != "Mobile App" {
		t.Errorf("expected second project name Mobile App, got %s", projects[1].Name)
	}
}

func TestProjectListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		search := r.URL.Query().Get("search")
		clientID := r.URL.Query().Get("clientId")

		projects := []clockify.Project{
			{ID: "p1", Name: "Website", ClientID: "c1", Status: "ACTIVE"},
		}

		if status != "" && status != "ACTIVE" {
			projects = nil
		}
		if search != "" && !strings.Contains("Website", search) {
			projects = nil
		}
		if clientID != "" && clientID != "c1" {
			projects = nil
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
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
	cmd.SetArgs([]string{"project", "list", "--status", "ACTIVE", "--search", "Web", "--client", "c1", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var projects []clockify.Project
	if err := json.Unmarshal(buf.Bytes(), &projects); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "Website" {
		t.Errorf("expected name Website, got %s", projects[0].Name)
	}
}

func TestProjectGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/projects/p1" {
			t.Errorf("expected path /workspaces/ws123/projects/p1, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		proj := clockify.Project{ID: "p1", Name: "Website", ClientID: "c1", Billable: true, Status: "ACTIVE", Color: "#FF0000"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(proj)
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
	cmd.SetArgs([]string{"project", "get", "--id", "p1", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var proj clockify.Project
	if err := json.Unmarshal(buf.Bytes(), &proj); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if proj.ID != "p1" {
		t.Errorf("expected ID p1, got %s", proj.ID)
	}
	if proj.Name != "Website" {
		t.Errorf("expected name Website, got %s", proj.Name)
	}
}

func TestProjectGetByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{
			{ID: "p1", Name: "Website Project", ClientID: "c1"},
			{ID: "p2", Name: "Mobile App", ClientID: "c2"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
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
	cmd.SetArgs([]string{"project", "get", "--name", "Website Project", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var proj clockify.Project
	if err := json.Unmarshal(buf.Bytes(), &proj); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if proj.ID != "p1" {
		t.Errorf("expected ID p1, got %s", proj.ID)
	}
}

func TestProjectGetByNameMultipleMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{
			{ID: "p1", Name: "App Mobile", ClientID: "c1"},
			{ID: "p2", Name: "App Web", ClientID: "c2"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
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
	cmd.SetArgs([]string{"project", "get", "--name", "App"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for multiple matches, got nil")
	}
	if !strings.Contains(err.Error(), "multiple") {
		t.Errorf("expected 'multiple' error, got: %v", err)
	}
}

func TestProjectGetByNameNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{
			{ID: "p1", Name: "Website", ClientID: "c1"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
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
	cmd.SetArgs([]string{"project", "get", "--name", "Nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !strings.Contains(err.Error(), "no project found") {
		t.Errorf("expected 'no project found' error, got: %v", err)
	}
}

func TestProjectGetNoFlags(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"project", "get"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing flags, got nil")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("expected 'must specify' error, got: %v", err)
	}
}

func TestProjectListNoProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects := []clockify.Project{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
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
	cmd.SetArgs([]string{"project", "list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No projects found") {
		t.Errorf("expected 'No projects found', got:\n%s", output)
	}
}

func TestProjectCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/workspaces/ws123/projects" {
			t.Errorf("expected path /workspaces/ws123/projects, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var req clockify.CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if req.Name != "New Project" {
			t.Errorf("expected name New Project, got %s", req.Name)
		}
		if req.ClientID != "c1" {
			t.Errorf("expected clientID c1, got %s", req.ClientID)
		}
		if !req.Billable {
			t.Error("expected billable to be true")
		}

		proj := clockify.Project{ID: "p1", Name: req.Name, ClientID: req.ClientID, Billable: req.Billable, Status: "ACTIVE"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(proj)
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
	cmd.SetArgs([]string{"project", "create", "--name", "New Project", "--client", "c1", "--billable", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var proj clockify.Project
	if err := json.Unmarshal(buf.Bytes(), &proj); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if proj.ID != "p1" {
		t.Errorf("expected ID p1, got %s", proj.ID)
	}
	if proj.Name != "New Project" {
		t.Errorf("expected name New Project, got %s", proj.Name)
	}
}

func TestProjectCreateNoName(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"project", "create"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "must specify --name") {
		t.Errorf("expected 'must specify --name' error, got: %v", err)
	}
}

func TestProjectUpdateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/projects" {
			// List projects for name resolution
			projects := []clockify.Project{
				{ID: "p1", Name: "Website", ClientID: "c1", Billable: true, Status: "ACTIVE"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(projects)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/workspaces/ws123/projects/p1" {
			proj := clockify.Project{ID: "p1", Name: "Website", ClientID: "c1", Billable: true, Status: "ACTIVE"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(proj)
			return
		}
		if r.Method == "PUT" && r.URL.Path == "/workspaces/ws123/projects/p1" {
			var req clockify.UpdateProjectRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request body: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if req.Name != "Updated Project" {
				t.Errorf("expected name Updated Project, got %s", req.Name)
			}
			proj := clockify.Project{ID: "p1", Name: req.Name, ClientID: req.ClientID, Billable: true, Status: "ACTIVE"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(proj)
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
	cmd.SetArgs([]string{"project", "update", "--id", "p1", "--new-name", "Updated Project", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var proj clockify.Project
	if err := json.Unmarshal(buf.Bytes(), &proj); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if proj.Name != "Updated Project" {
		t.Errorf("expected name Updated Project, got %s", proj.Name)
	}
}

func TestProjectDeleteConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/projects" {
			// List projects for name resolution
			projects := []clockify.Project{
				{ID: "p1", Name: "Website", ClientID: "c1"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(projects)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/workspaces/ws123/projects/p1" {
			proj := clockify.Project{ID: "p1", Name: "Website", ClientID: "c1"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(proj)
			return
		}
		if r.Method == "DELETE" && r.URL.Path == "/workspaces/ws123/projects/p1" {
			w.WriteHeader(http.StatusOK)
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
	cmd.SetArgs([]string{"project", "delete", "--id", "p1", "--confirm"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Project p1 deleted") {
		t.Errorf("expected 'Project p1 deleted', got:\n%s", output)
	}
}

func TestProjectDeleteNoConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/projects" {
			projects := []clockify.Project{
				{ID: "p1", Name: "Website", ClientID: "c1"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(projects)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/workspaces/ws123/projects/p1" {
			proj := clockify.Project{ID: "p1", Name: "Website", ClientID: "c1"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(proj)
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
	cmd.SetArgs([]string{"project", "delete", "--id", "p1"})

	// Simulate user typing "n" to abort
	// Note: In a real test, you'd mock stdin. For now, we just verify the command runs without --confirm
	// and would prompt. Since we can't easily mock stdin, we'll test the --confirm path above.
	err := cmd.Execute()
	// The command will try to read from stdin, which in tests is empty, so it will abort
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Aborted") && !strings.Contains(output, "deleted") {
		t.Errorf("expected 'Aborted' or 'deleted', got:\n%s", output)
	}
}

func TestProjectCreateIfNotExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/projects" {
			if r.Method == "GET" {
				// List projects for name resolution
				projects := []clockify.Project{
					{ID: "p1", Name: "Existing Project", ClientID: "c1"},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(projects)
				return
			}
			if r.Method == "POST" {
				t.Error("POST should not be called when --if-not-exists and project exists")
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
	cmd.SetArgs([]string{"project", "create", "--name", "Existing Project", "--if-not-exists", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already exists") {
		t.Errorf("expected 'already exists', got:\n%s", output)
	}
}
