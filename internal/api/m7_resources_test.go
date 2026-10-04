package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kim/clockify-cli/pkg/clockify"
)

func m7TestClient(handler http.HandlerFunc) (*ClockifyClient, func()) {
	server := httptest.NewServer(handler)
	return NewClient("test-key", server.URL, "ws123"), server.Close
}

func TestListUserGroups(t *testing.T) {
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws123/user-groups" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]clockify.UserGroup{
			{ID: "g1", Name: "Engineering", UserIDs: []string{"u1", "u2"}},
		})
	})
	defer done()

	groups, err := client.ListUserGroups(context.Background(), ListUserGroupsOpts{})
	if err != nil {
		t.Fatalf("ListUserGroups() error = %v", err)
	}
	if len(groups) != 1 || groups[0].Name != "Engineering" {
		t.Errorf("groups = %+v", groups)
	}
}

func TestCreateUserGroupAndMembers(t *testing.T) {
	var addCalls int
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/user-groups":
			json.NewEncoder(w).Encode(clockify.UserGroup{ID: "g1", Name: "Eng"})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/user-groups/g1/users":
			addCalls++
			var body clockify.AddUserToGroupRequest
			json.NewDecoder(r.Body).Decode(&body)
			if body.UserID != "u1" {
				t.Errorf("userId = %s", body.UserID)
			}
			json.NewEncoder(w).Encode(clockify.UserGroup{ID: "g1", Name: "Eng", UserIDs: []string{"u1"}})
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/users/u1"):
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	defer done()

	g, err := client.CreateUserGroup(context.Background(), clockify.CreateUserGroupRequest{Name: "Eng"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	g, err = client.AddUserToGroup(context.Background(), g.ID, "u1")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	if addCalls != 1 || len(g.UserIDs) != 1 {
		t.Errorf("addCalls=%d users=%v", addCalls, g.UserIDs)
	}
	if err := client.RemoveUserFromGroup(context.Background(), g.ID, "u1"); err != nil {
		t.Fatalf("remove user: %v", err)
	}
}

func TestListWebhooksAndCreate(t *testing.T) {
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/webhooks":
			json.NewEncoder(w).Encode([]clockify.Webhook{{ID: "w1", URL: "https://x", WebhookEvent: "NEW_TIME_ENTRY", Enabled: true}})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/webhooks":
			var req clockify.CreateWebhookRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.WebhookEvent != "NEW_TIME_ENTRY" {
				t.Errorf("event = %s", req.WebhookEvent)
			}
			json.NewEncoder(w).Encode(clockify.Webhook{ID: "w2", URL: req.URL, WebhookEvent: req.WebhookEvent, Enabled: req.Enabled})
		case r.Method == http.MethodDelete && r.URL.Path == "/workspaces/ws123/webhooks/w1":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	defer done()

	hooks, err := client.ListWebhooks(context.Background(), ListWebhooksOpts{})
	if err != nil || len(hooks) != 1 {
		t.Fatalf("list: %v %v", hooks, err)
	}
	h, err := client.CreateWebhook(context.Background(), clockify.CreateWebhookRequest{
		URL: "https://example.com/hook", WebhookEvent: "NEW_TIME_ENTRY", Enabled: true,
	})
	if err != nil || h.ID != "w2" {
		t.Fatalf("create: %v %+v", err, h)
	}
	if err := client.DeleteWebhook(context.Background(), "w1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestCustomFieldCRUD(t *testing.T) {
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/custom-fields":
			json.NewEncoder(w).Encode([]clockify.CustomField{{ID: "cf1", Name: "Department", Type: "TXT"}})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/custom-fields":
			var req clockify.CreateCustomFieldRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Type != "DROPDOWN_SINGLE" {
				t.Errorf("type = %s", req.Type)
			}
			json.NewEncoder(w).Encode(clockify.CustomField{ID: "cf2", Name: req.Name, Type: req.Type, AllowedValues: req.AllowedValues})
		case r.Method == http.MethodPut && r.URL.Path == "/workspaces/ws123/custom-fields/cf1":
			var req clockify.UpdateCustomFieldRequest
			json.NewDecoder(r.Body).Decode(&req)
			json.NewEncoder(w).Encode(clockify.CustomField{ID: "cf1", Name: req.Name, Type: req.Type})
		case r.Method == http.MethodDelete && r.URL.Path == "/workspaces/ws123/custom-fields/cf1":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	})
	defer done()

	fields, err := client.ListCustomFields(context.Background(), ListCustomFieldsOpts{})
	if err != nil || len(fields) != 1 {
		t.Fatalf("list: %v %v", fields, err)
	}
	created, err := client.CreateCustomField(context.Background(), clockify.CreateCustomFieldRequest{
		Name: "Dept", Type: "DROPDOWN_SINGLE", AllowedValues: []string{"Eng", "Sales"},
	})
	if err != nil || created.ID != "cf2" {
		t.Fatalf("create: %v %+v", err, created)
	}
	updated, err := client.UpdateCustomField(context.Background(), "cf1", clockify.UpdateCustomFieldRequest{Name: "Dept2", Type: "TXT"})
	if err != nil || updated.Name != "Dept2" {
		t.Fatalf("update: %v %+v", err, updated)
	}
	if err := client.DeleteCustomField(context.Background(), "cf1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestExpenseListWrappedAndCRUD(t *testing.T) {
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/expenses":
			json.NewEncoder(w).Encode(clockify.ExpenseListResponse{
				Expenses: []clockify.Expense{{ID: "e1", Total: 45.5, CategoryID: "cat1"}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/expenses/categories":
			json.NewEncoder(w).Encode([]clockify.ExpenseCategory{{ID: "cat1", Name: "TRAVEL"}})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/expenses":
			var req clockify.CreateExpenseRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Total != 45.5 {
				t.Errorf("total = %v", req.Total)
			}
			json.NewEncoder(w).Encode(clockify.Expense{ID: "e2", Total: req.Total, CategoryID: req.CategoryID})
		case r.Method == http.MethodDelete && r.URL.Path == "/workspaces/ws123/expenses/e1":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	defer done()

	expenses, err := client.ListExpenses(context.Background(), ListExpensesOpts{})
	if err != nil || len(expenses) != 1 || expenses[0].Total != 45.5 {
		t.Fatalf("list: %v %v", expenses, err)
	}
	cats, err := client.ListExpenseCategories(context.Background())
	if err != nil || len(cats) != 1 || cats[0].Name != "TRAVEL" {
		t.Fatalf("categories: %v %v", cats, err)
	}
	created, err := client.CreateExpense(context.Background(), clockify.CreateExpenseRequest{
		ProjectID: "p1", UserID: "u1", CategoryID: "cat1", Date: "2026-09-21", Total: 45.5, Billable: true,
	})
	if err != nil || created.ID != "e2" {
		t.Fatalf("create: %v %+v", err, created)
	}
	if err := client.DeleteExpense(context.Background(), "e1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestInvoiceListWrappedAndStatus(t *testing.T) {
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/invoices":
			json.NewEncoder(w).Encode(clockify.InvoiceListResponse{
				Invoices: []clockify.Invoice{{ID: "i1", Number: "INV-1", InvoiceStatus: "UNSENT", Currency: "USD"}},
				Total:    1,
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/invoices":
			var req clockify.CreateInvoiceRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.ClientID != "c1" || req.Currency != "USD" || req.Number == "" {
				t.Errorf("req = %+v", req)
			}
			json.NewEncoder(w).Encode(clockify.Invoice{ID: "i2", Number: req.Number, ClientID: req.ClientID, InvoiceStatus: req.InvoiceStatus})
		case r.Method == http.MethodPatch && r.URL.Path == "/workspaces/ws123/invoices/i1/status":
			var req clockify.ChangeInvoiceStatusRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.InvoiceStatus != "SENT" {
				t.Errorf("status = %s", req.InvoiceStatus)
			}
			json.NewEncoder(w).Encode(clockify.Invoice{ID: "i1", InvoiceStatus: "SENT"})
		case r.Method == http.MethodDelete && r.URL.Path == "/workspaces/ws123/invoices/i1":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	defer done()

	invoices, _, err := client.ListInvoices(context.Background(), ListInvoicesOpts{})
	if err != nil || len(invoices) != 1 {
		t.Fatalf("list: %v %v", invoices, err)
	}
	created, err := client.CreateInvoice(context.Background(), clockify.CreateInvoiceRequest{
		ClientID: "c1", Currency: "USD", IssuedDate: "2026-09-01T00:00:00.000000Z",
		DueDate: "2026-09-30T00:00:00.000000Z", Number: "INV-9",
	})
	if err != nil || created.ID != "i2" {
		t.Fatalf("create: %v %+v", err, created)
	}
	updated, err := client.ChangeInvoiceStatus(context.Background(), "i1", "SENT")
	if err != nil || updated.InvoiceStatus != "SENT" {
		t.Fatalf("status: %v %+v", err, updated)
	}
	if err := client.DeleteInvoice(context.Background(), "i1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestTimeOffListPostAndPolicyOps(t *testing.T) {
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/time-off/policies":
			json.NewEncoder(w).Encode([]clockify.TimeOffPolicy{{ID: "pol1", Name: "Vacation Policy"}})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/time-off/requests":
			var body clockify.ListTimeOffRequestsRequest
			json.NewDecoder(r.Body).Decode(&body)
			if body.Start == "" || body.End == "" {
				t.Errorf("list body missing dates: %+v", body)
			}
			json.NewEncoder(w).Encode(clockify.TimeOffListResponse{
				Count: 1,
				Requests: []clockify.TimeOffRequest{
					{ID: "to1", Status: "PENDING", PolicyID: "pol1", Start: body.Start, End: body.End},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/time-off/policies/pol1/requests":
			var req clockify.CreateTimeOffRequestRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.TimeOffPeriod.Start == "" {
				t.Errorf("missing period start")
			}
			json.NewEncoder(w).Encode(clockify.TimeOffRequest{ID: "to2", PolicyID: "pol1", Status: "PENDING"})
		case r.Method == http.MethodPatch && r.URL.Path == "/workspaces/ws123/time-off/policies/pol1/requests/to2":
			var req clockify.UpdateTimeOffStatusRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Status != "APPROVED" {
				t.Errorf("status = %s", req.Status)
			}
			json.NewEncoder(w).Encode(clockify.TimeOffRequest{ID: "to2", PolicyID: "pol1", Status: "APPROVED"})
		case r.Method == http.MethodDelete && r.URL.Path == "/workspaces/ws123/time-off/policies/pol1/requests/to2":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	defer done()

	policies, err := client.ListTimeOffPolicies(context.Background())
	if err != nil || len(policies) != 1 {
		t.Fatalf("policies: %v %v", policies, err)
	}
	requests, err := client.ListTimeOffRequests(context.Background(), clockify.ListTimeOffRequestsRequest{
		Start: "2026-09-01T00:00:00.000000Z", End: "2026-09-30T23:59:59.000000Z", Statuses: []string{"PENDING"},
	})
	if err != nil || len(requests) != 1 {
		t.Fatalf("list: %v %v", requests, err)
	}
	created, err := client.CreateTimeOffRequest(context.Background(), "pol1", clockify.CreateTimeOffRequestRequest{
		Note: "out",
		TimeOffPeriod: clockify.TimeOffRequestPeriod{
			Start: "2026-12-20T00:00:00.000000Z", End: "2026-12-31T23:59:59.000000Z",
		},
	})
	if err != nil || created.ID != "to2" {
		t.Fatalf("create: %v %+v", err, created)
	}
	updated, err := client.UpdateTimeOffStatus(context.Background(), "pol1", "to2", clockify.UpdateTimeOffStatusRequest{Status: "APPROVED", Note: "ok"})
	if err != nil || updated.Status != "APPROVED" {
		t.Fatalf("update-status: %v %+v", err, updated)
	}
	if err := client.DeleteTimeOffRequest(context.Background(), "pol1", "to2"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestScheduleAssignmentOps(t *testing.T) {
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/scheduling/assignments/all":
			if r.URL.Query().Get("start") == "" || r.URL.Query().Get("end") == "" {
				t.Errorf("missing start/end query")
			}
			json.NewEncoder(w).Encode([]clockify.ScheduleAssignment{
				{ID: "a1", ProjectID: "p1", UserID: "u1", HoursPerDay: 8},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/scheduling/assignments/recurring":
			var req clockify.CreateScheduleAssignmentRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.HoursPerDay != 8 || req.ProjectID == "" || req.UserID == "" {
				t.Errorf("req = %+v", req)
			}
			json.NewEncoder(w).Encode(clockify.ScheduleAssignment{ID: "a2", ProjectID: req.ProjectID, UserID: req.UserID, HoursPerDay: req.HoursPerDay})
		case r.Method == http.MethodPatch && r.URL.Path == "/workspaces/ws123/scheduling/assignments/recurring/a2":
			var req clockify.UpdateScheduleAssignmentRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.HoursPerDay == nil || *req.HoursPerDay != 6 {
				t.Errorf("hours = %+v", req.HoursPerDay)
			}
			json.NewEncoder(w).Encode(clockify.ScheduleAssignment{ID: "a2", HoursPerDay: 6})
		case r.Method == http.MethodDelete && r.URL.Path == "/workspaces/ws123/scheduling/assignments/recurring/a2":
			if r.URL.Query().Get("seriesUpdateOption") != "THIS_ONE" {
				t.Errorf("series = %s", r.URL.Query().Get("seriesUpdateOption"))
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	defer done()

	assignments, err := client.ListAssignments(context.Background(), ListAssignmentsOpts{
		Start: "2026-09-01T00:00:00.000000Z", End: "2026-09-30T23:59:59.000000Z",
	})
	if err != nil || len(assignments) != 1 {
		t.Fatalf("list: %v %v", assignments, err)
	}
	created, err := client.CreateAssignment(context.Background(), clockify.CreateScheduleAssignmentRequest{
		ProjectID: "p1", UserID: "u1",
		Start: "2026-09-21T00:00:00.000000Z", End: "2026-09-25T23:59:59.000000Z",
		HoursPerDay: 8,
	})
	if err != nil || created.ID != "a2" {
		t.Fatalf("create: %v %+v", err, created)
	}
	hpd := 6.0
	updated, err := client.UpdateAssignment(context.Background(), "a2", clockify.UpdateScheduleAssignmentRequest{HoursPerDay: &hpd, SeriesUpdateOption: "THIS_ONE"})
	if err != nil || updated.HoursPerDay != 6 {
		t.Fatalf("update: %v %+v", err, updated)
	}
	if err := client.DeleteAssignment(context.Background(), "a2", "THIS_ONE"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestApprovalOps(t *testing.T) {
	client, done := m7TestClient(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workspaces/ws123/approval-requests":
			if r.URL.Query().Get("status") != "PENDING" {
				t.Errorf("status = %s", r.URL.Query().Get("status"))
			}
			json.NewEncoder(w).Encode([]clockify.ApprovalRequest{
				{ID: "ar1", Status: "PENDING", Type: "TIMESHEET", UserID: "u1"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workspaces/ws123/approval-requests/TIMESHEET":
			var req clockify.SubmitApprovalRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Period != "WEEKLY" || req.PeriodStart == "" {
				t.Errorf("req = %+v", req)
			}
			json.NewEncoder(w).Encode(clockify.ApprovalRequest{ID: "ar2", Status: "PENDING", Type: "TIMESHEET"})
		case r.Method == http.MethodPatch && r.URL.Path == "/workspaces/ws123/approval-requests/ar1":
			var req clockify.UpdateApprovalStatusRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.State != "APPROVED" {
				t.Errorf("state = %s", req.State)
			}
			json.NewEncoder(w).Encode(clockify.ApprovalRequest{ID: "ar1", Status: "APPROVED"})
		default:
			http.NotFound(w, r)
		}
	})
	defer done()

	requests, err := client.ListApprovalRequests(context.Background(), ListApprovalRequestsOpts{Status: "PENDING"})
	if err != nil || len(requests) != 1 {
		t.Fatalf("list: %v %v", requests, err)
	}
	submitted, err := client.SubmitApprovalRequest(context.Background(), "TIMESHEET", clockify.SubmitApprovalRequest{
		Period: "WEEKLY", PeriodStart: "2026-09-29T00:00:00.000000Z",
	})
	if err != nil || submitted.ID != "ar2" {
		t.Fatalf("submit: %v %+v", err, submitted)
	}
	updated, err := client.UpdateApprovalStatus(context.Background(), "ar1", clockify.UpdateApprovalStatusRequest{State: "APPROVED"})
	if err != nil || updated.Status != "APPROVED" {
		t.Fatalf("update: %v %+v", err, updated)
	}
}
