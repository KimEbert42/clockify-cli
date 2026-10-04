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

func TestM7HelpCommands(t *testing.T) {
	cmds := []string{
		"invoice", "schedule", "time-off", "expense",
		"custom-field", "webhook", "approval", "group",
	}
	for _, c := range cmds {
		buf := new(bytes.Buffer)
		cmd := newRootCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{c, "--help"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%s --help error = %v", c, err)
		}
		out := buf.String()
		if out == "" {
			t.Errorf("%s help is empty", c)
		}
	}
}

func TestGroupListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/user-groups" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode([]clockify.UserGroup{
			{ID: "g1", Name: "Engineering", UserIDs: []string{"u1", "u2"}},
		})
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
	cmd.SetArgs([]string{"group", "list", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var groups []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &groups); err != nil {
		t.Fatalf("parse JSON: %v\n%s", err, buf.String())
	}
	if len(groups) != 1 || groups[0]["name"] != "Engineering" {
		t.Errorf("groups = %v", groups)
	}
	if groups[0]["memberCount"] != float64(2) {
		t.Errorf("memberCount = %v", groups[0]["memberCount"])
	}
}

func TestWebhookCreateValidatesEvents(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", "http://127.0.0.1:1")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"webhook", "create", "--url", "https://x", "--events", "NOT_REAL"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid event")
	}
	if !strings.Contains(err.Error(), "webhook event") {
		t.Errorf("error = %v", err)
	}
}

func TestCustomFieldTypeMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workspaces/ws123/custom-fields" {
			http.NotFound(w, r)
			return
		}
		var req clockify.CreateCustomFieldRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Type != "TXT" {
			t.Errorf("type = %s, want TXT", req.Type)
		}
		json.NewEncoder(w).Encode(clockify.CustomField{ID: "cf1", Name: req.Name, Type: req.Type})
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
	cmd.SetArgs([]string{"custom-field", "create", "--name", "Dept", "--type", "TEXT", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var f clockify.CustomField
	if err := json.Unmarshal(buf.Bytes(), &f); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if f.Type != "TXT" {
		t.Errorf("type = %s", f.Type)
	}
}

func TestCustomFieldDateRejected(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", "http://127.0.0.1:1")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"custom-field", "create", "--name", "When", "--type", "DATE"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "DATE") {
		t.Errorf("error = %v", err)
	}
}

func TestInvoiceStatusMappingInCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/clients/c1":
			json.NewEncoder(w).Encode(clockify.Client{ID: "c1", Name: "Acme"})
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/clients":
			json.NewEncoder(w).Encode([]clockify.Client{{ID: "c1", Name: "Acme"}})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/invoices":
			var req clockify.CreateInvoiceRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.InvoiceStatus != "UNSENT" {
				t.Errorf("invoiceStatus = %s, want UNSENT", req.InvoiceStatus)
			}
			if req.ClientID != "c1" {
				t.Errorf("clientId = %s", req.ClientID)
			}
			json.NewEncoder(w).Encode(clockify.Invoice{ID: "i1", Number: req.Number, InvoiceStatus: req.InvoiceStatus})
		default:
			http.NotFound(w, r)
		}
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
	cmd.SetArgs([]string{
		"invoice", "create",
		"--client", "c1",
		"--start", "2026-09-01",
		"--end", "2026-09-30",
		"--currency", "USD",
		"--status", "DRAFT",
		"--output", "json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, buf.String())
	}
	var inv clockify.Invoice
	if err := json.Unmarshal(buf.Bytes(), &inv); err != nil {
		t.Fatalf("parse: %v\n%s", err, buf.String())
	}
	if inv.InvoiceStatus != "UNSENT" {
		t.Errorf("status = %s", inv.InvoiceStatus)
	}
}

func TestApprovalSubmitRejectsTimeEntry(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", "http://127.0.0.1:1")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"approval", "submit", "--time-entry", "e1", "--period", "WEEKLY", "--start", "2026-09-29"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "time-entry") {
		t.Errorf("error = %v", err)
	}
}

func TestApprovalSubmitPeriod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workspaces/ws123/approval-requests/TIMESHEET" {
			http.NotFound(w, r)
			return
		}
		var req clockify.SubmitApprovalRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Period != "WEEKLY" || req.PeriodStart == "" {
			t.Errorf("req = %+v", req)
		}
		json.NewEncoder(w).Encode(clockify.ApprovalRequest{ID: "ar1", Status: "PENDING", Type: "TIMESHEET"})
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
	cmd.SetArgs([]string{
		"approval", "submit",
		"--type", "TIMESHEET",
		"--period", "WEEKLY",
		"--start", "2026-09-29",
		"--output", "json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, buf.String())
	}
	var ar clockify.ApprovalRequest
	if err := json.Unmarshal(buf.Bytes(), &ar); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ar.ID != "ar1" {
		t.Errorf("id = %s", ar.ID)
	}
}

func TestScheduleCreateHoursParse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/users":
			json.NewEncoder(w).Encode([]clockify.User{{ID: "u1", Name: "Jane", Email: "jane@x.com"}})
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/projects/p1":
			json.NewEncoder(w).Encode(clockify.Project{ID: "p1", Name: "Website"})
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/projects":
			json.NewEncoder(w).Encode([]clockify.Project{{ID: "p1", Name: "Website"}})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/scheduling/assignments/recurring":
			var req clockify.CreateScheduleAssignmentRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.HoursPerDay != 8 {
				t.Errorf("hoursPerDay = %v, want 8", req.HoursPerDay)
			}
			json.NewEncoder(w).Encode(clockify.ScheduleAssignment{ID: "a1", HoursPerDay: req.HoursPerDay})
		default:
			http.NotFound(w, r)
		}
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
	cmd.SetArgs([]string{
		"schedule", "create",
		"--user", "u1",
		"--project", "p1",
		"--start", "2026-09-21",
		"--end", "2026-09-25",
		"--duration", "8h",
		"--output", "json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, buf.String())
	}
	var a clockify.ScheduleAssignment
	if err := json.Unmarshal(buf.Bytes(), &a); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if a.HoursPerDay != 8 {
		t.Errorf("hoursPerDay = %v", a.HoursPerDay)
	}
}

func TestTimeOffCreateRequiresPolicyOrType(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", "http://127.0.0.1:1")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"time-off", "create", "--start", "2026-12-20", "--end", "2026-12-31"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "policy") {
		t.Errorf("error = %v", err)
	}
}

func TestExpenseCreateRequiresAmount(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", "http://127.0.0.1:1")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"expense", "create", "--project", "p1", "--date", "2026-09-21", "--type", "TRAVEL"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "amount") {
		t.Errorf("error = %v", err)
	}
}

func TestParseHoursPerDay(t *testing.T) {
	tests := []struct {
		in      string
		want    float64
		wantErr bool
	}{
		{"8", 8, false},
		{"8h", 8, false},
		{"1.5h", 1.5, false},
		{"", 0, true},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		got, err := parseHoursPerDay(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseHoursPerDay(%q) err = %v", tt.in, err)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseHoursPerDay(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseDateTimeCLI(t *testing.T) {
	got, err := parseDateTimeCLI("2026-09-21")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.HasPrefix(got, "2026-09-21T") {
		t.Errorf("got %q", got)
	}
	if _, err := parseDateTimeCLI("not-a-date"); err == nil {
		t.Error("expected error")
	}
	got2, err := parseDateTimeCLI("today")
	if err != nil || !strings.Contains(got2, "T") {
		t.Errorf("today = %q, %v", got2, err)
	}
}
