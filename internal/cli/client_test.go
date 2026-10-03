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

func TestClientCreateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/workspaces/ws123/clients" {
			t.Errorf("expected path /workspaces/ws123/clients, got %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var req clockify.CreateClientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if req.Name != "New Client" {
			t.Errorf("expected name New Client, got %s", req.Name)
		}
		if req.Email != "new@example.com" {
			t.Errorf("expected email new@example.com, got %s", req.Email)
		}

		client := clockify.Client{ID: "c1", Name: req.Name, Email: req.Email, Address: req.Address, Note: req.Note}
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
	cmd.SetArgs([]string{"client", "create", "--name", "New Client", "--email", "new@example.com", "--output", "json"})

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
	if client.Name != "New Client" {
		t.Errorf("expected name New Client, got %s", client.Name)
	}
}

func TestClientCreateNoName(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"client", "create"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "must specify --name") {
		t.Errorf("expected 'must specify --name' error, got: %v", err)
	}
}

func TestClientUpdateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/clients" {
			clients := []clockify.Client{
				{ID: "c1", Name: "Acme Corp", Email: "acme@example.com"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clients)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/workspaces/ws123/clients/c1" {
			client := clockify.Client{ID: "c1", Name: "Acme Corp", Email: "acme@example.com"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(client)
			return
		}
		if r.Method == "PUT" && r.URL.Path == "/workspaces/ws123/clients/c1" {
			var req clockify.UpdateClientRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request body: %v", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if req.Name != "Updated Client" {
				t.Errorf("expected name Updated Client, got %s", req.Name)
			}
			client := clockify.Client{ID: "c1", Name: req.Name, Email: req.Email}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(client)
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
	cmd.SetArgs([]string{"client", "update", "--id", "c1", "--new-name", "Updated Client", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var client clockify.Client
	if err := json.Unmarshal(buf.Bytes(), &client); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if client.Name != "Updated Client" {
		t.Errorf("expected name Updated Client, got %s", client.Name)
	}
}

func TestClientDeleteConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/clients" {
			clients := []clockify.Client{
				{ID: "c1", Name: "Acme Corp"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clients)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/workspaces/ws123/clients/c1" {
			client := clockify.Client{ID: "c1", Name: "Acme Corp"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(client)
			return
		}
		if r.Method == "DELETE" && r.URL.Path == "/workspaces/ws123/clients/c1" {
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
	cmd.SetArgs([]string{"client", "delete", "--id", "c1", "--confirm"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Client c1 deleted") {
		t.Errorf("expected 'Client c1 deleted', got:\n%s", output)
	}
}

func TestClientCreateIfNotExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/workspaces/ws123/clients" {
			if r.Method == "GET" {
				clients := []clockify.Client{
					{ID: "c1", Name: "Existing Client"},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(clients)
				return
			}
			if r.Method == "POST" {
				t.Error("POST should not be called when --if-not-exists and client exists")
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
	cmd.SetArgs([]string{"client", "create", "--name", "Existing Client", "--if-not-exists", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already exists") {
		t.Errorf("expected 'already exists', got:\n%s", output)
	}
}
