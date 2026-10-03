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

func TestTagListHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"tag", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "list") {
		t.Errorf("tag help should mention 'list', got:\n%s", output)
	}
}

func TestTagListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/tags" {
			t.Errorf("expected path /workspaces/ws123/tags, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		tags := []clockify.Tag{
			{ID: "t1", Name: "bug", Color: "#FF0000"},
			{ID: "t2", Name: "feature", Color: "#00FF00"},
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
	cmd.SetArgs([]string{"tag", "list", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var tags []clockify.Tag
	if err := json.Unmarshal(buf.Bytes(), &tags); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0].ID != "t1" {
		t.Errorf("expected first tag ID t1, got %s", tags[0].ID)
	}
	if tags[1].Name != "feature" {
		t.Errorf("expected second tag name feature, got %s", tags[1].Name)
	}
}

func TestTagListWithSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")

		tags := []clockify.Tag{
			{ID: "t1", Name: "bug", Color: "#FF0000"},
		}

		if search != "" && search != "bug" {
			tags = nil
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
	cmd.SetArgs([]string{"tag", "list", "--search", "bug", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var tags []clockify.Tag
	if err := json.Unmarshal(buf.Bytes(), &tags); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "bug" {
		t.Errorf("expected name bug, got %s", tags[0].Name)
	}
}

func TestTagListNoTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tags := []clockify.Tag{}
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
	cmd.SetArgs([]string{"tag", "list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No tags found") {
		t.Errorf("expected 'No tags found', got:\n%s", output)
	}
}

func TestTagCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/workspaces/ws123/tags" {
			t.Errorf("expected path /workspaces/ws123/tags, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var req clockify.CreateTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if req.Name != "new-tag" {
			t.Errorf("expected name new-tag, got %s", req.Name)
		}

		tag := clockify.Tag{ID: "t1", Name: req.Name}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tag)
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
	cmd.SetArgs([]string{"tag", "create", "--name", "new-tag", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var tag clockify.Tag
	if err := json.Unmarshal(buf.Bytes(), &tag); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if tag.ID != "t1" {
		t.Errorf("expected ID t1, got %s", tag.ID)
	}
	if tag.Name != "new-tag" {
		t.Errorf("expected name new-tag, got %s", tag.Name)
	}
}

func TestTagCreateNoName(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"tag", "create"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "must specify --name") {
		t.Errorf("expected 'must specify --name' error, got: %v", err)
	}
}

func TestTagUpdateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/tags" {
			tags := []clockify.Tag{
				{ID: "t1", Name: "bug"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tags)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/workspaces/ws123/tags/t1" {
			tag := clockify.Tag{ID: "t1", Name: "bug"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tag)
			return
		}
		if r.Method == "PUT" && r.URL.Path == "/workspaces/ws123/tags/t1" {
			var req clockify.UpdateTagRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request body: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if req.Name != "updated-tag" {
				t.Errorf("expected name updated-tag, got %s", req.Name)
			}
			tag := clockify.Tag{ID: "t1", Name: req.Name}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tag)
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
	cmd.SetArgs([]string{"tag", "update", "--id", "t1", "--new-name", "updated-tag", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var tag clockify.Tag
	if err := json.Unmarshal(buf.Bytes(), &tag); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if tag.Name != "updated-tag" {
		t.Errorf("expected name updated-tag, got %s", tag.Name)
	}
}

func TestTagDeleteConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/tags" {
			tags := []clockify.Tag{
				{ID: "t1", Name: "bug"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tags)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/workspaces/ws123/tags/t1" {
			tag := clockify.Tag{ID: "t1", Name: "bug"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tag)
			return
		}
		if r.Method == "DELETE" && r.URL.Path == "/workspaces/ws123/tags/t1" {
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
	cmd.SetArgs([]string{"tag", "delete", "--id", "t1", "--confirm"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Tag t1 deleted") {
		t.Errorf("expected 'Tag t1 deleted', got:\n%s", output)
	}
}

func TestTagCreateIfNotExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/tags" {
			if r.Method == "GET" {
				tags := []clockify.Tag{
					{ID: "t1", Name: "existing-tag"},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tags)
				return
			}
			if r.Method == "POST" {
				t.Error("POST should not be called when --if-not-exists and tag exists")
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
	cmd.SetArgs([]string{"tag", "create", "--name", "existing-tag", "--if-not-exists", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already exists") {
		t.Errorf("expected 'already exists', got:\n%s", output)
	}
}
