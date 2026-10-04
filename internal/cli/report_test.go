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
	"time"

	"github.com/kim/clockify-cli/pkg/clockify"
)

func TestReportHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"report", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	out := buf.String()
	for _, sub := range []string{"summary", "detailed", "weekly"} {
		if !strings.Contains(out, sub) {
			t.Errorf("report help should mention %q, got:\n%s", sub, out)
		}
	}
}

func TestReportSummaryJSON(t *testing.T) {
	var reportsPath string
	var reportsBody clockify.ReportRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Main API: resolve project by name via list
		if r.URL.Path == "/workspaces/ws123/projects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]clockify.Project{{ID: "p1", Name: "Website"}})
			return
		}
		if r.URL.Path == "/workspaces/ws123/projects/p1" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clockify.Project{ID: "p1", Name: "Website"})
			return
		}
		// Reports API (same test server via CLOCKIFY_REPORTS_URL)
		if strings.HasSuffix(r.URL.Path, "/reports/summary") {
			reportsPath = r.URL.Path
			json.NewDecoder(r.Body).Decode(&reportsBody)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"groups": []map[string]any{
					{"name": "Website", "totals": map[string]any{"duration": 3600, "amount": 42.5}},
				},
				"totals": map[string]any{"duration": 3600, "amount": 42.5},
			})
			return
		}
		http.Error(w, "not found: "+r.URL.Path, http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)
	t.Setenv("CLOCKIFY_REPORTS_URL", server.URL)
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"report", "summary",
		"--start", "2026-10-01",
		"--end", "2026-10-03",
		"--project", "Website",
		"--billable",
		"--output", "json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\noutput: %s", err, buf.String())
	}

	if !strings.HasSuffix(reportsPath, "/reports/summary") {
		t.Errorf("reports path = %q", reportsPath)
	}
	if reportsBody.DateRangeStart == "" {
		t.Errorf("dateRangeStart empty")
	}
	if reportsBody.Billable == nil || !*reportsBody.Billable {
		t.Errorf("billable not sent: %+v", reportsBody)
	}
	if reportsBody.Projects == nil || len(reportsBody.Projects.IDs) != 1 || reportsBody.Projects.IDs[0] != "p1" {
		t.Errorf("project filter = %+v", reportsBody.Projects)
	}

	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
}

func TestReportDetailedExportCSVFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/reports/detailed") {
			var body clockify.ReportRequest
			json.NewDecoder(r.Body).Decode(&body)
			if body.ExportType != "CSV" {
				t.Errorf("exportType = %q", body.ExportType)
			}
			if body.DetailedFilter == nil {
				t.Errorf("detailedFilter missing")
			}
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte("date,description,duration\n2026-10-01,Work,1h\n"))
			return
		}
		http.Error(w, "not found", 404)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)
	t.Setenv("CLOCKIFY_REPORTS_URL", server.URL)
	setupTestConfig(t, "ws123")

	outFile := filepath.Join(t.TempDir(), "report.csv")
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"report", "detailed",
		"--start", "2026-10-01",
		"--end", "2026-10-02",
		"--export", "csv",
		"--file", outFile,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, buf.String())
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read export file: %v", err)
	}
	if !strings.Contains(string(data), "2026-10-01") {
		t.Errorf("CSV content = %s", data)
	}
}

func TestReportDetailedClientSideCSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/reports/detailed") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"timeentries": []map[string]any{
					{
						"id":          "e1",
						"description": "Design work",
						"billable":    true,
						"timeInterval": map[string]any{
							"start":    "2026-10-01T09:00:00Z",
							"end":      "2026-10-01T10:00:00Z",
							"duration": 3600,
						},
					},
				},
			})
			return
		}
		http.Error(w, "not found", 404)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)
	t.Setenv("CLOCKIFY_REPORTS_URL", server.URL)
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"report", "detailed",
		"--start", "2026-10-01",
		"--end", "2026-10-01",
		"--export", "csv",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Design work") {
		t.Errorf("CSV output = %s", out)
	}
}

func TestReportWeeklyDefaultsEnd(t *testing.T) {
	var start, end string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/reports/weekly") {
			var body clockify.ReportRequest
			json.NewDecoder(r.Body).Decode(&body)
			start = body.DateRangeStart
			end = body.DateRangeEnd
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"days": []map[string]any{
					{"date": "2026-09-15", "totals": map[string]any{"duration": 1800}},
				},
			})
			return
		}
		http.Error(w, "not found", 404)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)
	t.Setenv("CLOCKIFY_REPORTS_URL", server.URL)
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"report", "weekly",
		"--start", "2026-09-15",
		"--output", "json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, buf.String())
	}
	if !strings.HasPrefix(start, "2026-09-15") {
		t.Errorf("start = %q", start)
	}
	if !strings.HasPrefix(end, "2026-09-21") {
		t.Errorf("end = %q, want start+6 days", end)
	}
}

func TestReportRequiresStart(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_API_URL", "http://127.0.0.1:1")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"report", "summary", "--end", "2026-10-01"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--start") {
		t.Fatalf("expected --start error, got %v", err)
	}
}

func TestReportInvalidStatus(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_API_URL", "http://127.0.0.1:1")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"report", "summary",
		"--start", "2026-10-01",
		"--end", "2026-10-02",
		"--status", "BOGUS",
	})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestReportRejectedStatusUnsupported(t *testing.T) {
	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_API_URL", "http://127.0.0.1:1")
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"report", "summary",
		"--start", "2026-10-01",
		"--end", "2026-10-02",
		"--status", "REJECTED",
	})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "REJECTED") {
		t.Fatalf("expected REJECTED error, got %v", err)
	}
}

func TestReportRelativeDates(t *testing.T) {
	var start, end string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/reports/summary") {
			var body clockify.ReportRequest
			json.NewDecoder(r.Body).Decode(&body)
			start = body.DateRangeStart
			end = body.DateRangeEnd
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"groups": []any{}})
			return
		}
		http.Error(w, "not found", 404)
	}))
	defer server.Close()

	t.Setenv("CLOCKIFY_API_KEY", "test-key")
	t.Setenv("CLOCKIFY_REGION", "global")
	t.Setenv("CLOCKIFY_API_URL", server.URL)
	t.Setenv("CLOCKIFY_REPORTS_URL", server.URL)
	setupTestConfig(t, "ws123")

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"report", "summary",
		"--start", "this-month",
		"--end", "today",
		"--output", "json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v\n%s", err, buf.String())
	}
	if start == "" || end == "" {
		t.Fatalf("dates not parsed: start=%q end=%q", start, end)
	}
	if !strings.HasSuffix(start, "T00:00:00.000000") {
		t.Errorf("start format = %q", start)
	}
	if !strings.HasSuffix(end, "T23:59:59.999999") {
		t.Errorf("end format = %q", end)
	}
}

func TestParseReportDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	got, err := parseReportDate("2026-10-01")
	if err != nil {
		t.Fatalf("parseReportDate(2026-10-01) error = %v", err)
	}
	if got.Format("2006-01-02") != "2026-10-01" {
		t.Errorf("got %s, want 2026-10-01", got.Format("2006-01-02"))
	}

	got, err = parseReportDate("today")
	if err != nil {
		t.Fatalf("parseReportDate(today) error = %v", err)
	}
	if !got.Equal(today) {
		t.Errorf("today = %s, want %s", got.Format("2006-01-02"), today.Format("2006-01-02"))
	}

	if _, err := parseReportDate("not-a-date"); err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestFormatReportDate(t *testing.T) {
	day := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	if got := formatReportDate(day, false); got != "2026-10-01T00:00:00.000000" {
		t.Errorf("start format = %q", got)
	}
	if got := formatReportDate(day, true); got != "2026-10-01T23:59:59.999999" {
		t.Errorf("end format = %q", got)
	}
}

func TestMapReportStatus(t *testing.T) {
	if got, err := mapReportStatus("APPROVED"); err != nil || got != "APPROVED" {
		t.Errorf("APPROVED → %q, %v", got, err)
	}
	if got, err := mapReportStatus("NEW"); err != nil || got != "UNAPPROVED" {
		t.Errorf("NEW → %q, %v", got, err)
	}
	if got, err := mapReportStatus("IN_PROGRESS"); err != nil || got != "UNAPPROVED" {
		t.Errorf("IN_PROGRESS → %q, %v", got, err)
	}
	if _, err := mapReportStatus("REJECTED"); err == nil {
		t.Error("REJECTED should error")
	}
	if _, err := mapReportStatus("nope"); err == nil {
		t.Error("nope should error")
	}
	if got, err := mapReportStatus(""); err != nil || got != "" {
		t.Errorf("empty → %q, %v", got, err)
	}
}
