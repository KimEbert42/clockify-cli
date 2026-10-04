package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kim/clockify-cli/pkg/clockify"
)

func TestGenerateSummaryReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/workspaces/ws123/reports/summary" {
			t.Errorf("path = %s, want /workspaces/ws123/reports/summary", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("missing X-Api-Key")
		}
		var body clockify.ReportRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.DateRangeStart == "" || body.DateRangeEnd == "" {
			t.Errorf("date range not set: %+v", body)
		}
		if body.SummaryFilter == nil {
			t.Errorf("summaryFilter missing")
		}
		if body.Billable == nil || !*body.Billable {
			t.Errorf("billable filter not applied")
		}
		if body.Projects == nil || len(body.Projects.IDs) != 1 || body.Projects.IDs[0] != "p1" {
			t.Errorf("projects filter not applied: %+v", body.Projects)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"groups":[{"name":"Website","totals":{"duration":3600,"amount":100}}],"totals":{"duration":3600,"amount":100}}`))
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "ws123")
	billable := true
	body, contentType, err := client.GenerateSummaryReport(context.Background(), clockify.ReportRequest{
		DateRangeStart: "2026-10-01T00:00:00.000000",
		DateRangeEnd:   "2026-10-03T23:59:59.999999",
		Billable:       &billable,
		Projects:       &clockify.ContainsFilter{Contains: "CONTAINS", IDs: []string{"p1"}, Status: "ACTIVE"},
		SummaryFilter:  &clockify.SummaryFilter{SummaryChartType: "PROJECT"},
	})
	if err != nil {
		t.Fatalf("GenerateSummaryReport() error = %v", err)
	}
	if !strings.Contains(string(body), `"Website"`) {
		t.Errorf("unexpected body: %s", body)
	}
	if contentType != "application/json" {
		t.Errorf("contentType = %q", contentType)
	}
}

func TestGenerateDetailedReportRequiresFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["detailedFilter"]; !ok {
			t.Errorf("detailedFilter should be auto-added")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"timeentries":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "ws123")
	_, _, err := client.GenerateDetailedReport(context.Background(), clockify.ReportRequest{
		DateRangeStart: "2026-10-01T00:00:00.000000",
		DateRangeEnd:   "2026-10-02T23:59:59.999999",
	})
	if err != nil {
		t.Fatalf("GenerateDetailedReport() error = %v", err)
	}
}

func TestGenerateWeeklyReportPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/reports/weekly" {
			t.Errorf("path = %s", r.URL.Path)
			http.Error(w, "not found", 404)
			return
		}
		var body clockify.ReportRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.WeeklyFilter == nil {
			t.Errorf("weeklyFilter missing")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"days":[{"date":"2026-10-01","totals":{"duration":7200}}]}`))
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "ws123")
	body, _, err := client.GenerateWeeklyReport(context.Background(), clockify.ReportRequest{
		DateRangeStart: "2026-10-01T00:00:00.000000",
		DateRangeEnd:   "2026-10-07T23:59:59.999999",
	})
	if err != nil {
		t.Fatalf("GenerateWeeklyReport() error = %v", err)
	}
	if !strings.Contains(string(body), "2026-10-01") {
		t.Errorf("body = %s", body)
	}
}

func TestGenerateReportCSVExport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body clockify.ReportRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.ExportType != "CSV" {
			t.Errorf("exportType = %q, want CSV", body.ExportType)
		}
		w.Header().Set("Content-Type", "text/csv")
		w.Write([]byte("date,description,duration\n2026-10-01,Work,1h\n"))
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "ws123")
	body, contentType, err := client.GenerateDetailedReport(context.Background(), clockify.ReportRequest{
		DateRangeStart: "2026-10-01T00:00:00.000000",
		DateRangeEnd:   "2026-10-02T23:59:59.999999",
		ExportType:     "CSV",
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !LooksLikeCSV(contentType, body) {
		t.Errorf("expected CSV detection, contentType=%s body=%s", contentType, body)
	}
}

func TestGenerateReportRetryOn500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "ws123",
		WithRetry(RetryPolicy{MaxRetries: 3, BaseDelay: 5 * time.Millisecond, MaxDelay: 20 * time.Millisecond,
			RetryableFn: func(code int, err error) bool { return err != nil || code >= 500 },
		}),
	)
	_, _, err := client.GenerateSummaryReport(context.Background(), clockify.ReportRequest{
		DateRangeStart: "2026-10-01T00:00:00.000000",
		DateRangeEnd:   "2026-10-02T23:59:59.999999",
	})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestParseReportResponse(t *testing.T) {
	body := []byte(`{"groups":[{"name":"A","totals":{"duration":100}}],"timeentries":[{"id":"e1","description":"x"}]}`)
	resp, err := ParseReportResponse(body)
	if err != nil {
		t.Fatalf("ParseReportResponse error = %v", err)
	}
	if len(resp.Groups) != 1 || resp.Groups[0].Name != "A" {
		t.Errorf("groups = %+v", resp.Groups)
	}
	if len(resp.TimeEntries) != 1 || resp.TimeEntries[0].ID != "e1" {
		t.Errorf("timeentries = %+v", resp.TimeEntries)
	}
	if resp.Raw == nil {
		t.Errorf("Raw not populated")
	}
}

func TestLooksLikePDF(t *testing.T) {
	if !LooksLikePDF("application/pdf", []byte("%PDF-1.4")) {
		t.Error("expected PDF detection by content-type")
	}
	if !LooksLikePDF("", []byte("%PDF-1.4 rest")) {
		t.Error("expected PDF detection by magic bytes")
	}
	if LooksLikePDF("application/json", []byte(`{}`)) {
		t.Error("JSON should not look like PDF")
	}
}

func TestReportEntryIDFallback(t *testing.T) {
	e := clockify.ReportEntry{IDAlt: "abc"}
	if e.EntryID() != "abc" {
		t.Errorf("EntryID = %q", e.EntryID())
	}
	e2 := clockify.ReportEntry{ID: "xyz", IDAlt: "abc"}
	if e2.EntryID() != "xyz" {
		t.Errorf("EntryID = %q", e2.EntryID())
	}
}
