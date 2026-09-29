package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kim/clockify-cli/pkg/clockify"
)

func TestProjectListPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("X-Pagination-Total-Pages", "2")
		w.Header().Set("X-Pagination-Total-Results", "3")

		var projects []clockify.Project
		if callCount == 1 {
			projects = []clockify.Project{
				{ID: "p1", Name: "Project 1", Status: "ACTIVE"},
				{ID: "p2", Name: "Project 2", Status: "ACTIVE"},
			}
		} else {
			projects = []clockify.Project{
				{ID: "p3", Name: "Project 3", Status: "ACTIVE"},
			}
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
	cmd.SetArgs([]string{"project", "list", "--no-pagination", "--page-size", "2", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var projects []clockify.Project
	if err := json.Unmarshal(buf.Bytes(), &projects); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(projects) != 3 {
		t.Fatalf("expected 3 projects (all pages), got %d", len(projects))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

func TestProjectListSinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Pagination-Total-Pages", "1")
		w.Header().Set("X-Pagination-Total-Results", "1")

		projects := []clockify.Project{
			{ID: "p1", Name: "Only Project", Status: "ACTIVE"},
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
	cmd.SetArgs([]string{"project", "list", "--no-pagination", "--output", "json"})

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
}

func TestClientListPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("X-Pagination-Total-Pages", "2")
		w.Header().Set("X-Pagination-Total-Results", "3")

		var clients []clockify.Client
		if callCount == 1 {
			clients = []clockify.Client{
				{ID: "c1", Name: "Client 1"},
				{ID: "c2", Name: "Client 2"},
			}
		} else {
			clients = []clockify.Client{
				{ID: "c3", Name: "Client 3"},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clients)
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
	cmd.SetArgs([]string{"client", "list", "--no-pagination", "--page-size", "2", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var clients []clockify.Client
	if err := json.Unmarshal(buf.Bytes(), &clients); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(clients) != 3 {
		t.Fatalf("expected 3 clients (all pages), got %d", len(clients))
	}
}

func TestTagListPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("X-Pagination-Total-Pages", "2")
		w.Header().Set("X-Pagination-Total-Results", "3")

		var tags []clockify.Tag
		if callCount == 1 {
			tags = []clockify.Tag{
				{ID: "t1", Name: "bug", Color: "#FF0000"},
				{ID: "t2", Name: "feature", Color: "#00FF00"},
			}
		} else {
			tags = []clockify.Tag{
				{ID: "t3", Name: "urgent", Color: "#0000FF"},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tags)
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
	cmd.SetArgs([]string{"tag", "list", "--no-pagination", "--page-size", "2", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var tags []clockify.Tag
	if err := json.Unmarshal(buf.Bytes(), &tags); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(tags) != 3 {
		t.Fatalf("expected 3 tags (all pages), got %d", len(tags))
	}
}
