package clockify

import (
	"encoding/json"
	"testing"
)

func TestWorkspaceJSON(t *testing.T) {
	data := `{
		"id": "5f8a1b2c3d4e5f6a7b8c9d0e",
		"name": "My Workspace",
		"hourlyRate": {"amount": 5000, "currency": "USD"},
		"currency": "USD",
		"isMember": true,
		"membershipType": "OWNER",
		"workspaceSettings": {
			"weekStart": "MONDAY",
			"dateFormat": "YYYY-MM-DD",
			"timeFormat": "HH:mm",
			"allowTrackTimeWithTimer": true
		}
	}`

	var ws Workspace
	if err := json.Unmarshal([]byte(data), &ws); err != nil {
		t.Fatalf("failed to unmarshal workspace: %v", err)
	}

	if ws.ID != "5f8a1b2c3d4e5f6a7b8c9d0e" {
		t.Errorf("expected ID 5f8a1b2c3d4e5f6a7b8c9d0e, got %s", ws.ID)
	}
	if ws.Name != "My Workspace" {
		t.Errorf("expected name My Workspace, got %s", ws.Name)
	}
	if ws.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", ws.Currency)
	}
	if !ws.IsMember {
		t.Error("expected IsMember true")
	}
	if ws.HourlyRate.Amount != 5000 {
		t.Errorf("expected hourly rate 5000, got %d", ws.HourlyRate.Amount)
	}
}

func TestUserJSON(t *testing.T) {
	data := `{
		"id": "user123",
		"email": "test@example.com",
		"name": "Test User",
		"status": "ACTIVE",
		"profile": {
			"weekStart": "MONDAY",
			"timeZone": "America/New_York",
			"dateFormat": "YYYY-MM-DD",
			"timeFormat": "HH:mm",
			"billable": true,
			"locale": "en-US"
		},
		"activeWorkspace": "ws123",
		"defaultWorkspace": "ws123"
	}`

	var user User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		t.Fatalf("failed to unmarshal user: %v", err)
	}

	if user.ID != "user123" {
		t.Errorf("expected ID user123, got %s", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.Status != "ACTIVE" {
		t.Errorf("expected status ACTIVE, got %s", user.Status)
	}
	if user.Profile.TimeZone != "America/New_York" {
		t.Errorf("expected timezone America/New_York, got %s", user.Profile.TimeZone)
	}
}

func TestProjectJSON(t *testing.T) {
	data := `{
		"id": "proj123",
		"name": "Website Redesign",
		"clientId": "client456",
		"workspaceId": "ws123",
		"billable": true,
		"status": "ACTIVE",
		"archived": false,
		"duration": null,
		"estimate": {"type": "AUTO", "estimate": 7200},
		"activeTaskCount": 5,
		"color": "#e74c3c"
	}`

	var proj Project
	if err := json.Unmarshal([]byte(data), &proj); err != nil {
		t.Fatalf("failed to unmarshal project: %v", err)
	}

	if proj.ID != "proj123" {
		t.Errorf("expected ID proj123, got %s", proj.ID)
	}
	if proj.Name != "Website Redesign" {
		t.Errorf("expected name Website Redesign, got %s", proj.Name)
	}
	if proj.ClientID != "client456" {
		t.Errorf("expected clientId client456, got %s", proj.ClientID)
	}
	if !proj.Billable {
		t.Error("expected billable true")
	}
	if proj.ActiveTaskCount != 5 {
		t.Errorf("expected activeTaskCount 5, got %d", proj.ActiveTaskCount)
	}
}

func TestClientJSON(t *testing.T) {
	data := `{
		"id": "client456",
		"name": "Acme Corp",
		"workspaceId": "ws123",
		"email": "contact@acme.com",
		"address": "123 Main St",
		"phone": "555-0123",
		"note": "Important client"
	}`

	var client Client
	if err := json.Unmarshal([]byte(data), &client); err != nil {
		t.Fatalf("failed to unmarshal client: %v", err)
	}

	if client.ID != "client456" {
		t.Errorf("expected ID client456, got %s", client.ID)
	}
	if client.Name != "Acme Corp" {
		t.Errorf("expected name Acme Corp, got %s", client.Name)
	}
	if client.Email != "contact@acme.com" {
		t.Errorf("expected email contact@acme.com, got %s", client.Email)
	}
}

func TestTagJSON(t *testing.T) {
	data := `{
		"id": "tag789",
		"name": "meeting",
		"workspaceId": "ws123",
		"color": "#3498db"
	}`

	var tag Tag
	if err := json.Unmarshal([]byte(data), &tag); err != nil {
		t.Fatalf("failed to unmarshal tag: %v", err)
	}

	if tag.ID != "tag789" {
		t.Errorf("expected ID tag789, got %s", tag.ID)
	}
	if tag.Name != "meeting" {
		t.Errorf("expected name meeting, got %s", tag.Name)
	}
	if tag.Color != "#3498db" {
		t.Errorf("expected color #3498db, got %s", tag.Color)
	}
}

func TestWorkspaceListJSON(t *testing.T) {
	data := `[{"id":"ws1","name":"Workspace 1"},{"id":"ws2","name":"Workspace 2"}]`

	var workspaces []Workspace
	if err := json.Unmarshal([]byte(data), &workspaces); err != nil {
		t.Fatalf("failed to unmarshal workspace list: %v", err)
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

func TestProjectListJSON(t *testing.T) {
	data := `[
		{"id":"p1","name":"Project A","status":"ACTIVE","billable":true},
		{"id":"p2","name":"Project B","status":"ARCHIVED","billable":false}
	]`

	var projects []Project
	if err := json.Unmarshal([]byte(data), &projects); err != nil {
		t.Fatalf("failed to unmarshal project list: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Status != "ACTIVE" {
		t.Errorf("expected first project ACTIVE, got %s", projects[0].Status)
	}
	if projects[1].Status != "ARCHIVED" {
		t.Errorf("expected second project ARCHIVED, got %s", projects[1].Status)
	}
}
