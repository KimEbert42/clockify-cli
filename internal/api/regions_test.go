package api

import "testing"

func TestReportsBaseURLGlobal(t *testing.T) {
	t.Setenv("CLOCKIFY_REPORTS_URL", "")
	got := ReportsBaseURL("global")
	want := "https://reports.api.clockify.me/v1"
	if got != want {
		t.Errorf("ReportsBaseURL(global) = %q, want %q", got, want)
	}
}

func TestReportsBaseURLRegional(t *testing.T) {
	t.Setenv("CLOCKIFY_REPORTS_URL", "")
	cases := map[string]string{
		"eu-de": "https://euc1.clockify.me/report/v1",
		"us":    "https://use2.clockify.me/report/v1",
		"uk":    "https://euw2.clockify.me/report/v1",
		"au":    "https://apse2.clockify.me/report/v1",
	}
	for region, want := range cases {
		if got := ReportsBaseURL(region); got != want {
			t.Errorf("ReportsBaseURL(%s) = %q, want %q", region, got, want)
		}
	}
}

func TestReportsBaseURLUnknownFallsBackGlobal(t *testing.T) {
	t.Setenv("CLOCKIFY_REPORTS_URL", "")
	got := ReportsBaseURL("nope")
	want := "https://reports.api.clockify.me/v1"
	if got != want {
		t.Errorf("ReportsBaseURL(nope) = %q, want %q", got, want)
	}
}

func TestReportsBaseURLEnvOverride(t *testing.T) {
	t.Setenv("CLOCKIFY_REPORTS_URL", "http://127.0.0.1:9999/v1")
	got := ReportsBaseURL("global")
	want := "http://127.0.0.1:9999/v1"
	if got != want {
		t.Errorf("ReportsBaseURL(env) = %q, want %q", got, want)
	}
}

func TestBaseURLUnchangedByReports(t *testing.T) {
	t.Setenv("CLOCKIFY_API_URL", "")
	if got := BaseURL("global"); got != "https://api.clockify.me/api/v1" {
		t.Errorf("BaseURL(global) = %q", got)
	}
}
