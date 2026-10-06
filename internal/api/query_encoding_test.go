package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppendQueryParamsEncodesSpecialChars(t *testing.T) {
	got := appendQueryParams("/projects", map[string]string{
		"search":   "web design & more",
		"clientId": "c 1",
		"page":     "1",
	})
	if strings.Contains(got, "web design") {
		t.Errorf("query not encoded: %q", got)
	}
	if !strings.Contains(got, "search=web+design") && !strings.Contains(got, "search=web%20design") {
		t.Errorf("missing encoded search: %q", got)
	}
	if !strings.Contains(got, "clientId=c+1") && !strings.Contains(got, "clientId=c%201") {
		t.Errorf("missing encoded clientId: %q", got)
	}
}

func TestAppendQueryParamsOmitsEmpty(t *testing.T) {
	got := appendQueryParams("/tags", map[string]string{"search": "", "page": "2"})
	if strings.Contains(got, "search") {
		t.Errorf("empty search should be omitted: %q", got)
	}
	if !strings.Contains(got, "page=2") {
		t.Errorf("page=2 missing: %q", got)
	}
}

func TestBuildTimeEntryQueryParamsEncodes(t *testing.T) {
	q := buildTimeEntryQueryParams(ListTimeEntriesOpts{
		Start:       "2026-01-01T00:00:00Z",
		End:         "2026-01-02T23:59:59Z",
		Description: "a b",
		Tags:        "x,y",
		Page:        2,
		PageSize:    50,
	})
	if strings.Contains(q, "description=a b") {
		t.Errorf("description not encoded: %q", q)
	}
	if !strings.Contains(q, "page=2") || !strings.Contains(q, "pageSize=50") {
		t.Errorf("page params missing: %q", q)
	}
}

func TestListProjectsPagedSingleFetch(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if tp := r.Header.Get("X-Pagination-Total-Pages"); tp != "" {
			// already set by client — headers come from server
		}
		w.Header().Set("X-Pagination-Total-Pages", "2")
		w.Header().Set("X-Pagination-Total-Results", "7")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"p1","name":"A"},{"id":"p2","name":"B"}]`))
	}))
	t.Cleanup(srv.Close)

	client := NewClient("k", srv.URL, "ws")
	projects, info, err := client.ListProjectsPaged(context.Background(), ListProjectsOpts{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Errorf("projects=%d", len(projects))
	}
	if calls != 1 {
		t.Errorf("expected 1 HTTP call, got %d", calls)
	}
	if info == nil || info.TotalResults != 7 || info.TotalPages != 2 {
		t.Errorf("pageInfo=%+v", info)
	}
}
