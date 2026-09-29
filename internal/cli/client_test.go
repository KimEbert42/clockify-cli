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

func TestClientListHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"client", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "list") {
		t.Errorf("client help should mention 'list', got:\n%s", output)
	}
	if !strings.Contains(output, "get") {
		t.Errorf("client help should mention 'get', got:\n%s", output)
	}
}

func TestClientListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/clients" {
			t.Errorf("expected path /workspaces/ws123/clients, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		clients := []clockify.Client{
			{ID: "c1", Name: "Acme Corp", Email: "acme@example.com"},
			{ID: "c2", Name: "Globex Inc", Email: "globex@example.com"},
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
	cmd.SetArgs([]string{"client", "list", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var clients []clockify.Client
	if err := json.Unmarshal(buf.Bytes(), &clients); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(clients))
	}
	if clients[0].ID != "c1" {
		t.Errorf("expected first client ID c1, got %s", clients[0].ID)
	}
	if clients[1].Name != "Globex Inc" {
		t.Errorf("expected second client name Globex Inc, got %s", clients[1].Name)
	}
}

func TestClientListWithSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")

		clients := []clockify.Client{
			{ID: "c1", Name: "Acme Corp", Email: "acme@example.com"},
		}

		if search != "" && search != "Acme" {
			clients = nil
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
	cmd.SetArgs([]string{"client", "list", "--search", "Acme", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var clients []clockify.Client
	if err := json.Unmarshal(buf.Bytes(), &clients); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
	if clients[0].Name != "Acme Corp" {
		t.Errorf("expected name Acme Corp, got %s", clients[0].Name)
	}
}

func TestClientGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/clients/c1" {
			t.Errorf("expected path /workspaces/ws123/clients/c1, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		client := clockify.Client{ID: "c1", Name: "Acme Corp", Email: "acme@example.com", Address: "123 Main St", Phone: "555-0100", Note: "Important client"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(client)
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
	cmd.SetArgs([]string{"client", "get", "--id", "c1", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var client clockify.Client
	if err := json.Unmarshal(buf.Bytes(), &client); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if client.ID != "c1" {
		t.Errorf("expected ID c1, got %s", client.ID)
	}
	if client.Name != "Acme Corp" {
		t.Errorf("expected name Acme Corp, got %s", client.Name)
	}
	if client.Email != "acme@example.com" {
		t.Errorf("expected email acme@example.com, got %s", client.Email)
	}
}

func TestClientGetNoID(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"client", "get"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --id, got nil")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("expected 'must specify' error, got: %v", err)
	}
}

func TestClientGetNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Client not found"}`, http.StatusNotFound)
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
	cmd.SetArgs([]string{"client", "get", "--id", "nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
}

func TestClientListNoClients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clients := []clockify.Client{}
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
	cmd.SetArgs([]string{"client", "list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No clients found") {
		t.Errorf("expected 'No clients found', got:\n%s", output)
	}
}
