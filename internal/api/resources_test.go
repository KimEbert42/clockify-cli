package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kim/clockify-cli/internal/errors"
	"github.com/kim/clockify-cli/pkg/clockify"
)

func testAPI(t *testing.T, handler http.HandlerFunc) (*ClockifyClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient("secret-api-key-12345", srv.URL, "ws-1", WithRetry(RetryPolicy{
		MaxRetries: 0,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Millisecond,
		RetryableFn: func(statusCode int, err error) bool { return false },
	})), srv
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"short", "****"},
		{"12345678", "****"},
		{"secret-api-key-12345", "secr****2345"},
	}
	for _, tt := range tests {
		if got := MaskAPIKey(tt.in); got != tt.want {
			t.Errorf("MaskAPIKey(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestClientDebugRedactsAPIKey(t *testing.T) {
	var out bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	client := NewClient("super-secret-key-abcdef", srv.URL, "ws", WithDebug(true), WithRetry(RetryPolicy{
		MaxRetries:  0,
		BaseDelay:   time.Millisecond,
		RetryableFn: func(int, error) bool { return false },
	}))
	var result map[string]any
	_ = client.Do(context.Background(), "GET", "/user", nil, &result)

	_ = w.Close()
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	os.Stdout = old
	out.Write(buf[:n])
	combined := out.String()

	if strings.Contains(combined, "super-secret-key-abcdef") {
		t.Errorf("debug output leaked raw API key: %q", combined)
	}
	if !strings.Contains(combined, "api-key=") {
		t.Errorf("debug output missing api-key marker: %q", combined)
	}
	if !strings.Contains(combined, "supe****cdef") {
		t.Errorf("debug output should show masked key: %q", combined)
	}
}

func TestClientHandleErrorResponseMapping(t *testing.T) {
	tests := []struct {
		status int
		code   errors.ErrorCode
		sugg   string
	}{
		{401, errors.E001Auth, "configure"},
		{403, errors.E001Auth, "configure"},
		{404, errors.E002NotFound, "Check the resource"},
		{429, errors.E005RateLimit, "Wait a moment"},
		{400, errors.E004Validation, ""},
		{500, errors.E009Server, ""},
	}
	for _, tt := range tests {
		client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tt.status)
			_, _ = w.Write([]byte(`{"message":"boom"}`))
		})
		var result any
		err := client.Do(context.Background(), "GET", "/x", nil, &result)
		if err == nil {
			t.Fatalf("status %d: expected error", tt.status)
		}
		var cliErr *errors.CLIError
		if !errors.As(err, &cliErr) {
			t.Fatalf("status %d: not CLIError: %v", tt.status, err)
		}
		if cliErr.Code != tt.code {
			t.Errorf("status %d: code=%v want %v", tt.status, cliErr.Code, tt.code)
		}
		if tt.sugg != "" && !strings.Contains(cliErr.Suggestion, tt.sugg) {
			t.Errorf("status %d: suggestion=%q want contains %q", tt.status, cliErr.Suggestion, tt.sugg)
		}
	}
}

func TestClientDoWithPagination(t *testing.T) {
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Pagination-Total-Pages", "3")
		w.Header().Set("X-Pagination-Total-Results", "250")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"1"}]`))
	})
	var result []map[string]any
	info, err := client.DoWithPagination(context.Background(), "GET", "/workspaces/ws-1/projects", nil, &result)
	if err != nil {
		t.Fatalf("DoWithPagination() error = %v", err)
	}
	if info.TotalPages != 3 || info.TotalResults != 250 {
		t.Errorf("pageInfo=%+v", info)
	}
	if len(result) != 1 {
		t.Errorf("result=%v", result)
	}
	if client.WorkspaceID() != "ws-1" {
		t.Errorf("WorkspaceID()=%q", client.WorkspaceID())
	}
}

func TestClientDoWithPaginationMarshalError(t *testing.T) {
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {})
	_, err := client.DoWithPagination(context.Background(), "GET", "/x", make(chan int), nil)
	if err == nil {
		t.Fatal("expected marshal error")
	}
	var cliErr *errors.CLIError
	if !errors.As(err, &cliErr) || cliErr.Code != errors.E010Internal {
		t.Errorf("err=%v", err)
	}
}

func TestClientNetworkError(t *testing.T) {
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {})
	// Point at closed server
	client.baseURL = "http://127.0.0.1:1"
	err := client.Do(context.Background(), "GET", "/x", nil, nil)
	if err == nil {
		t.Fatal("expected network error")
	}
	var cliErr *errors.CLIError
	if !errors.As(err, &cliErr) || cliErr.Code != errors.E006Network {
		t.Errorf("err=%v", err)
	}
}

func TestClientWithOptions(t *testing.T) {
	c := NewClient("k", "u", "w",
		WithVerbose(true),
		WithDebug(true),
		WithRateLimit(100),
		WithRetry(RetryPolicy{MaxRetries: 2}),
	)
	if !c.verbose || !c.debug || c.retryPolicy.MaxRetries != 2 {
		t.Fatalf("options not applied: %+v", c)
	}
	if c.rateLimiter == nil {
		t.Fatal("rate limiter missing")
	}
}

func TestResourcePathsAndMethods(t *testing.T) {
	type call struct {
		method string
		path   string
	}
	var got []call
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		got = append(got, call{method: r.Method, path: r.URL.Path})
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(p, "/workspaces") && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"id":"ws-1","name":"WS"}]`))
		case strings.HasSuffix(p, "/workspaces/ws-1") && r.Method == "GET":
			_, _ = w.Write([]byte(`{"id":"ws-1","name":"WS"}`))
		case p == "/user" && r.Method == "GET":
			_, _ = w.Write([]byte(`{"id":"u1","email":"a@b.c","name":"A"}`))
		case strings.HasSuffix(p, "/users") && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"id":"u1","email":"a@b.c","name":"A"}]`))
		case strings.HasSuffix(p, "/projects") && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"id":"p1","name":"Project"}]`))
		case strings.HasSuffix(p, "/projects") && (r.Method == "POST" || r.Method == "PUT"):
			_, _ = w.Write([]byte(`{"id":"p1","name":"Project"}`))
		case strings.Contains(p, "/projects/") && strings.HasSuffix(p, "/tasks") && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"id":"t1","name":"Task"}]`))
		case strings.Contains(p, "/projects/") && strings.Contains(p, "/tasks/") && r.Method == "GET":
			_, _ = w.Write([]byte(`{"id":"t1","name":"Task"}`))
		case strings.Contains(p, "/projects/") && strings.HasSuffix(p, "/tasks") && r.Method != "GET":
			_, _ = w.Write([]byte(`{"id":"t1","name":"Task"}`))
		case strings.Contains(p, "/projects/") && r.Method == "GET":
			_, _ = w.Write([]byte(`{"id":"p1","name":"Project"}`))
		case strings.Contains(p, "/projects/") && r.Method != "GET":
			_, _ = w.Write([]byte(`{"id":"p1","name":"Project"}`))
		case strings.HasSuffix(p, "/clients") && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"id":"c1","name":"Client"}]`))
		case strings.HasSuffix(p, "/clients") && r.Method != "GET":
			_, _ = w.Write([]byte(`{"id":"c1","name":"Client"}`))
		case strings.Contains(p, "/clients/") && r.Method == "GET":
			_, _ = w.Write([]byte(`{"id":"c1","name":"Client"}`))
		case strings.Contains(p, "/clients/") && r.Method != "GET":
			_, _ = w.Write([]byte(`{"id":"c1","name":"Client"}`))
		case strings.HasSuffix(p, "/tags") && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"id":"tag1","name":"billable"}]`))
		case strings.HasSuffix(p, "/tags") && r.Method != "GET":
			_, _ = w.Write([]byte(`{"id":"tag1","name":"billable"}`))
		case strings.Contains(p, "/tags/") && r.Method == "GET":
			_, _ = w.Write([]byte(`{"id":"tag1","name":"billable"}`))
		case strings.Contains(p, "/tags/") && r.Method != "GET":
			_, _ = w.Write([]byte(`{"id":"tag1","name":"billable"}`))
		case strings.Contains(p, "/time-entries/status/in-progress") && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"id":"te1"}]`))
		case strings.Contains(p, "/time-entries/batch") && r.Method == "POST":
			_, _ = w.Write([]byte(`[{"id":"te1"}]`))
		case strings.Contains(p, "/user/") && strings.HasSuffix(p, "/time-entries") && r.Method == "PUT":
			_, _ = w.Write([]byte(`[{"id":"te1"}]`))
		case strings.HasSuffix(p, "/time-entries") && r.Method == "GET":
			_, _ = w.Write([]byte(`[{"id":"te1"}]`))
		case strings.HasSuffix(p, "/time-entries") && r.Method == "PATCH":
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(p, "/time-entries") && r.Method != "GET":
			_, _ = w.Write([]byte(`{"id":"te1"}`))
		case strings.Contains(p, "/time-entries/") && r.Method == "GET":
			_, _ = w.Write([]byte(`{"id":"te1"}`))
		case strings.Contains(p, "/time-entries/") && r.Method != "GET":
			_, _ = w.Write([]byte(`{"id":"te1"}`))
		case strings.HasSuffix(p, "/user/u1/time-entries") && r.Method == "PUT":
			_, _ = w.Write([]byte(`[{"id":"te1"}]`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	})

	ctx := context.Background()
	if _, err := client.ListWorkspaces(ctx); err != nil {
		t.Fatalf("ListWorkspaces: %v", err)
	}
	if _, err := client.GetWorkspace(ctx, "ws-1"); err != nil {
		t.Fatalf("GetWorkspace: %v", err)
	}
	if _, err := client.GetCurrentUser(ctx); err != nil {
		t.Fatalf("GetCurrentUser: %v", err)
	}
	if _, err := client.ListUsers(ctx, ListUsersOpts{Status: "ACTIVE", Page: 1, PageSize: 50}); err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if _, err := client.ListProjects(ctx, ListProjectsOpts{Status: "ACTIVE", Search: "web"}); err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if _, err := client.GetProject(ctx, "p1"); err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if _, err := client.CreateProject(ctx, clockify.CreateProjectRequest{Name: "P"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if _, err := client.UpdateProject(ctx, "p1", clockify.UpdateProjectRequest{Name: "P2"}); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if err := client.DeleteProject(ctx, "p1"); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	if _, err := client.ListClients(ctx, ListClientsOpts{Search: "acme"}); err != nil {
		t.Fatalf("ListClients: %v", err)
	}
	if _, err := client.GetClient(ctx, "c1"); err != nil {
		t.Fatalf("GetClient: %v", err)
	}
	if _, err := client.CreateClient(ctx, clockify.CreateClientRequest{Name: "C"}); err != nil {
		t.Fatalf("CreateClient: %v", err)
	}
	if _, err := client.UpdateClient(ctx, "c1", clockify.UpdateClientRequest{Name: "C2"}); err != nil {
		t.Fatalf("UpdateClient: %v", err)
	}
	if err := client.DeleteClient(ctx, "c1"); err != nil {
		t.Fatalf("DeleteClient: %v", err)
	}
	if _, err := client.ListTags(ctx, ListTagsOpts{Search: "bill"}); err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if _, err := client.GetTag(ctx, "tag1"); err != nil {
		t.Fatalf("GetTag: %v", err)
	}
	if _, err := client.CreateTag(ctx, clockify.CreateTagRequest{Name: "t"}); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if _, err := client.UpdateTag(ctx, "tag1", clockify.UpdateTagRequest{Name: "t2"}); err != nil {
		t.Fatalf("UpdateTag: %v", err)
	}
	if err := client.DeleteTag(ctx, "tag1"); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}

	active := true
	if _, err := client.ListTasks(ctx, "p1", ListTasksOpts{IsActive: &active, Page: 1}); err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if _, err := client.GetTask(ctx, "p1", "t1"); err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if _, err := client.CreateTask(ctx, "p1", clockify.CreateTaskRequest{Name: "T"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := client.UpdateTask(ctx, "p1", "t1", clockify.UpdateTaskRequest{Name: "T2"}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if err := client.DeleteTask(ctx, "p1", "t1"); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	if _, err := client.ListTimeEntries(ctx, "u1", ListTimeEntriesOpts{Start: "2026-01-01", End: "2026-01-02"}); err != nil {
		t.Fatalf("ListTimeEntries: %v", err)
	}
	if _, _, err := client.ListTimeEntriesWithPagination(ctx, "u1", ListTimeEntriesOpts{Page: 1, PageSize: 10}); err != nil {
		t.Fatalf("ListTimeEntriesWithPagination: %v", err)
	}
	if _, err := client.GetTimeEntry(ctx, "te1"); err != nil {
		t.Fatalf("GetTimeEntry: %v", err)
	}
	if _, err := client.GetTimeEntryInProgress(ctx, "u1"); err != nil {
		t.Fatalf("GetTimeEntryInProgress: %v", err)
	}
	if _, err := client.BatchGetTimeEntries(ctx, []string{"te1"}, false); err != nil {
		t.Fatalf("BatchGetTimeEntries: %v", err)
	}
	if _, err := client.CreateTimeEntry(ctx, CreateTimeEntryRequest{Start: "2026-01-01T10:00:00Z", Description: "x"}); err != nil {
		t.Fatalf("CreateTimeEntry: %v", err)
	}
	if _, err := client.UpdateTimeEntry(ctx, "te1", UpdateTimeEntryRequest{Start: "2026-01-01T11:00:00Z", Description: "y"}); err != nil {
		t.Fatalf("UpdateTimeEntry: %v", err)
	}
	if err := client.DeleteTimeEntry(ctx, "te1"); err != nil {
		t.Fatalf("DeleteTimeEntry: %v", err)
	}
	if err := client.StopTimeEntry(ctx, "u1", StopTimerRequest{End: "2026-01-01T12:00:00Z"}); err != nil {
		t.Fatalf("StopTimeEntry: %v", err)
	}
	if _, err := client.BulkUpdateTimeEntries(ctx, "u1", []UpdateTimeEntryRequest{{Start: "x", Description: "z"}}); err != nil {
		t.Fatalf("BulkUpdateTimeEntries: %v", err)
	}

	if len(got) == 0 {
		t.Fatal("handler never called")
	}
}

func TestAppendQueryParamsAndBuildTimeEntryQuery(t *testing.T) {
	got := appendQueryParams("/a", map[string]string{"x": "1", "y": "", "z": "2"})
	if !strings.Contains(got, "x=1") || strings.Contains(got, "y=") || !strings.Contains(got, "z=2") {
		t.Errorf("appendQueryParams=%q", got)
	}
	if got != "/a?x=1&z=2" && got != "/a?z=2&x=1" {
		t.Errorf("appendQueryParams order/path unexpected: %q", got)
	}
	if appendQueryParams("/a", map[string]string{}) != "/a" {
		t.Error("empty params should not append ?")
	}

	q := buildTimeEntryQueryParams(ListTimeEntriesOpts{Start: "s", End: "e", ProjectID: "p", TaskID: "t", Tags: "g", Description: "d", InProgress: "TRUE", Page: 2, PageSize: 50})
	for _, part := range []string{"start=s", "end=e", "project=p", "task=t", "tags=g", "description=d", "in-progress=TRUE", "page=2", "pageSize=50"} {
		if !strings.Contains(q, part) {
			t.Errorf("query missing %s: %q", part, q)
		}
	}
	if strings.HasSuffix(q, "&") {
		t.Errorf("trailing amp: %q", q)
	}
}

func TestReportsLooksLikeCSV(t *testing.T) {
	if !LooksLikeCSV("text/csv", nil) {
		t.Error("content-type csv should be true")
	}
	if !LooksLikeCSV("", []byte("a,b,c\n1,2,3")) {
		t.Error("csv body should be true")
	}
	if LooksLikeCSV("application/json", []byte(`{"a":1}`)) {
		t.Error("json body should be false")
	}
	if LooksLikeCSV("", nil) {
		t.Error("empty should be false")
	}
}

func TestReportErrorFromStatus(t *testing.T) {
	err := reportErrorFromStatus(401, []byte(`{"message":"nope"}`))
	var cliErr *errors.CLIError
	if !errors.As(err, &cliErr) || cliErr.Code != errors.E001Auth {
		t.Fatalf("err=%v", err)
	}
	err = reportErrorFromStatus(404, []byte(`{}`))
	if !errors.As(err, &cliErr) || cliErr.Code != errors.E002NotFound {
		t.Fatalf("err=%v", err)
	}
	err = reportErrorFromStatus(429, []byte(`{}`))
	if !errors.As(err, &cliErr) || cliErr.Code != errors.E005RateLimit {
		t.Fatalf("err=%v", err)
	}
	err = reportErrorFromStatus(400, []byte(`{"message":"bad filter"}`))
	if err == nil || !strings.Contains(err.Error(), "bad filter") {
		t.Fatalf("err=%v", err)
	}
	err = reportErrorFromStatus(500, []byte(`{}`))
	if !errors.As(err, &cliErr) || cliErr.Code != errors.E009Server {
		t.Fatalf("err=%v", err)
	}
}

func TestLooksLikeCSVHelperExported(t *testing.T) {
	if !LooksLikeCSV("text/plain", []byte("col1,col2")) {
		t.Fatal("expected csv-like body")
	}
}

func TestResourceMethodPathsExact(t *testing.T) {
	var paths []string
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"x"}]`))
	})
	ctx := context.Background()
	_, _ = client.ListProjects(ctx, ListProjectsOpts{})
	_, _ = client.ListClients(ctx, ListClientsOpts{})
	_, _ = client.ListTags(ctx, ListTagsOpts{})
	_, _ = client.ListUsers(ctx, ListUsersOpts{})
	_, _ = client.ListTasks(ctx, "p", ListTasksOpts{})
	_, _ = client.ListTimeEntries(ctx, "u", ListTimeEntriesOpts{})

	want := map[string]bool{
		"GET /workspaces/ws-1/projects":        true,
		"GET /workspaces/ws-1/clients":         true,
		"GET /workspaces/ws-1/tags":            true,
		"GET /workspaces/ws-1/users":           true,
		"GET /workspaces/ws-1/projects/p/tasks": true,
		"GET /workspaces/ws-1/user/u/time-entries": true,
	}
	for _, p := range paths {
		delete(want, p)
	}
	if len(want) != 0 {
		t.Errorf("missing expected calls: %v (got %v)", want, paths)
	}
}

func TestDoJSONUnmarshalError(t *testing.T) {
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	})
	var result map[string]any
	err := client.Do(context.Background(), "GET", "/x", nil, &result)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
	var cliErr *errors.CLIError
	if !errors.As(err, &cliErr) || cliErr.Code != errors.E010Internal {
		t.Errorf("err=%v", err)
	}
}

func TestDo204SkipsBody(t *testing.T) {
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	var result map[string]any
	if err := client.Do(context.Background(), "DELETE", "/x", nil, &result); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}

func TestM7InvoiceGetUpdateDelete(t *testing.T) {
	var method, path string
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET", "PUT":
			_, _ = w.Write([]byte(`{"id":"inv1","invoiceStatus":"DRAFT"}`))
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		case "PATCH":
			_, _ = w.Write([]byte(`{"id":"inv1","invoiceStatus":"SENT"}`))
		default:
			_, _ = w.Write([]byte(`[{"id":"inv1"}]`))
		}
	})
	ctx := context.Background()
	inv, err := client.GetInvoice(ctx, "inv1")
	if err != nil || inv.ID != "inv1" {
		t.Fatalf("GetInvoice=%v %v", inv, err)
	}
	if method != "GET" || path != "/workspaces/ws-1/invoices/inv1" {
		t.Errorf("get path=%s %s", method, path)
	}
	inv, err = client.UpdateInvoice(ctx, "inv1", clockify.UpdateInvoiceRequest{Number: "1"})
	if err != nil {
		t.Fatalf("UpdateInvoice: %v", err)
	}
	_ = inv
	if err := client.DeleteInvoice(ctx, "inv1"); err != nil {
		t.Fatalf("DeleteInvoice: %v", err)
	}
	inv, err = client.ChangeInvoiceStatus(ctx, "inv1", "SENT")
	if err != nil {
		t.Fatalf("ChangeInvoiceStatus: %v", err)
	}
	if inv.InvoiceStatus != "SENT" {
		t.Errorf("status=%q", inv.InvoiceStatus)
	}
}

func TestDecodeInvoiceListWrapped(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"invoices": []map[string]any{{"id": "a"}},
		"total":    1,
	})
	inv, info, err := decodeInvoiceList(raw, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(inv) != 1 || info.TotalResults != 1 {
		t.Fatalf("inv=%v info=%+v", inv, info)
	}
	// Array form
	raw2, _ := json.Marshal([]map[string]any{{"id": "b"}})
	inv2, info2, err := decodeInvoiceList(raw2, 2, 5)
	if err != nil || len(inv2) != 1 || info2.TotalResults != 1 {
		t.Fatalf("array form: inv=%v info=%+v err=%v", inv2, info2, err)
	}
	// Invalid JSON fails both paths
	if _, _, err := decodeInvoiceList([]byte(`[1,2`), 1, 1); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestM7OtherResourcesPaths(t *testing.T) {
	var paths []string
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if strings.Contains(r.URL.Path, "list") || r.Method == "GET" && !strings.Contains(r.URL.Path, "expense") {
			// return object-shaped when path ends with id-like segment
			if strings.Count(r.URL.Path, "/") >= 4 && !strings.HasSuffix(r.URL.Path, "expenses") && !strings.HasSuffix(r.URL.Path, "webhooks") && !strings.HasSuffix(r.URL.Path, "user-groups") {
				_, _ = w.Write([]byte(`{"id":"x","name":"X"}`))
				return
			}
		}
		_, _ = w.Write([]byte(`[{"id":"x","name":"X"}]`))
	})
	ctx := context.Background()

	// Exercise remaining M7 wrappers for path coverage
	_, _ = client.GetExpense(ctx, "e1")
	_, _ = client.UpdateExpense(ctx, "e1", clockify.UpdateExpenseRequest{})
	_, _ = client.GetWebhook(ctx, "w1")
	_, _ = client.UpdateWebhook(ctx, "w1", clockify.UpdateWebhookRequest{})
	_, _ = client.GetUserGroup(ctx, "g1")
	_, _ = client.UpdateUserGroup(ctx, "g1", clockify.UpdateUserGroupRequest{})
	_ = client.DeleteUserGroup(ctx, "g1")
	_, _ = client.SubmitApprovalRequestForUser(ctx, "u1", "p1", clockify.SubmitApprovalRequest{})

	if len(paths) == 0 {
		t.Fatal("no M7 calls recorded")
	}
	joined := strings.Join(paths, "\n")
	for _, want := range []string{"/expenses/e1", "/webhooks/w1", "/user-groups/g1"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing path fragment %s in:\n%s", want, joined)
		}
	}
}

func TestClientDoMarshalBody(t *testing.T) {
	var body map[string]any
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	var result map[string]any
	if err := client.Do(context.Background(), "POST", "/x", map[string]string{"a": "b"}, &result); err != nil {
		t.Fatal(err)
	}
	if body["a"] != "b" {
		t.Errorf("body=%v", body)
	}
}

func TestClientRetryOn429(t *testing.T) {
	attempts := 0
	client, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	client.retryPolicy = &RetryPolicy{
		MaxRetries:  3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    2 * time.Millisecond,
		RetryableFn: func(code int, err error) bool { return code == 429 },
	}
	var result map[string]any
	if err := client.Do(context.Background(), "GET", "/x", nil, &result); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Errorf("attempts=%d", attempts)
	}
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("k", "https://api.example", "ws")
	if c.baseURL != "https://api.example" || c.workspaceID != "ws" || c.apiKey != "k" {
		t.Fatalf("defaults wrong: %+v", c)
	}
	if c.retryPolicy == nil || c.rateLimiter == nil {
		t.Fatal("defaults missing retry/rate limit")
	}
}

func TestStatusCodeConstantsUsedInTestsCompile(t *testing.T) {
	// Guard against accidental removal of helpers used above.
	_ = fmt.Sprintf
	_ = errors.E010Internal
}
