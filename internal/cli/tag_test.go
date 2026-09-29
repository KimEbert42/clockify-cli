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
