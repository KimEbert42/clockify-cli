package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kim/clockify-cli/pkg/clockify"
)

// ---------- Invoice ----------

type ListInvoicesOpts struct {
	ClientID string
	Number   string
	Status   string
	Page     int
	PageSize int
}

func (c *ClockifyClient) ListInvoices(ctx context.Context, opts ListInvoicesOpts) ([]clockify.Invoice, *clockify.PageInfo, error) {
	path := fmt.Sprintf("/workspaces/%s/invoices", c.workspaceID)
	params := map[string]string{
		"clientId":  opts.ClientID,
		"number":    opts.Number,
		"status":    opts.Status,
		"page":      strconv.Itoa(opts.Page),
		"pageSize":  strconv.Itoa(opts.PageSize),
	}
	path = appendQueryParams(path, params)

	var raw json.RawMessage
	if err := c.Do(ctx, "GET", path, nil, &raw); err != nil {
		return nil, nil, err
	}
	invoices, pageInfo, err := decodeInvoiceList(raw, opts.Page, opts.PageSize)
	return invoices, pageInfo, err
}

func decodeInvoiceList(raw json.RawMessage, page, pageSize int) ([]clockify.Invoice, *clockify.PageInfo, error) {
	var arr []clockify.Invoice
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, &clockify.PageInfo{Page: page, PageSize: pageSize, TotalResults: len(arr)}, nil
	}
	var wrapped clockify.InvoiceListResponse
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		return wrapped.Invoices, &clockify.PageInfo{Page: page, PageSize: pageSize, TotalResults: wrapped.Total}, nil
	}
	return nil, nil, fmt.Errorf("unexpected invoice list response")
}

func (c *ClockifyClient) GetInvoice(ctx context.Context, id string) (*clockify.Invoice, error) {
	path := fmt.Sprintf("/workspaces/%s/invoices/%s", c.workspaceID, id)
	var inv clockify.Invoice
	err := c.Do(ctx, "GET", path, nil, &inv)
	return &inv, err
}

func (c *ClockifyClient) CreateInvoice(ctx context.Context, req clockify.CreateInvoiceRequest) (*clockify.Invoice, error) {
	path := fmt.Sprintf("/workspaces/%s/invoices", c.workspaceID)
	var inv clockify.Invoice
	err := c.Do(ctx, "POST", path, req, &inv)
	return &inv, err
}

func (c *ClockifyClient) UpdateInvoice(ctx context.Context, id string, req clockify.UpdateInvoiceRequest) (*clockify.Invoice, error) {
	path := fmt.Sprintf("/workspaces/%s/invoices/%s", c.workspaceID, id)
	var inv clockify.Invoice
	err := c.Do(ctx, "PUT", path, req, &inv)
	return &inv, err
}

func (c *ClockifyClient) DeleteInvoice(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/invoices/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

func (c *ClockifyClient) ChangeInvoiceStatus(ctx context.Context, id string, status string) (*clockify.Invoice, error) {
	path := fmt.Sprintf("/workspaces/%s/invoices/%s/status", c.workspaceID, id)
	var inv clockify.Invoice
	err := c.Do(ctx, "PATCH", path, clockify.ChangeInvoiceStatusRequest{InvoiceStatus: status}, &inv)
	if err != nil {
		return nil, err
	}
	if inv.InvoiceStatus == "" {
		inv.InvoiceStatus = status
	}
	return &inv, nil
}

// ---------- Schedule ----------

type ListAssignmentsOpts struct {
	Start    string
	End      string
	Name     string
	UserID   string
	ProjectID string
	Page     int
	PageSize int
}

func (c *ClockifyClient) ListAssignments(ctx context.Context, opts ListAssignmentsOpts) ([]clockify.ScheduleAssignment, error) {
	path := fmt.Sprintf("/workspaces/%s/scheduling/assignments/all", c.workspaceID)
	params := map[string]string{
		"start":     opts.Start,
		"end":       opts.End,
		"name":      opts.Name,
		"user":      opts.UserID,
		"project":   opts.ProjectID,
		"page":      strconv.Itoa(opts.Page),
		"pageSize":  strconv.Itoa(opts.PageSize),
	}
	path = appendQueryParams(path, params)
	var assignments []clockify.ScheduleAssignment
	err := c.Do(ctx, "GET", path, nil, &assignments)
	return assignments, err
}

func (c *ClockifyClient) CreateAssignment(ctx context.Context, req clockify.CreateScheduleAssignmentRequest) (*clockify.ScheduleAssignment, error) {
	path := fmt.Sprintf("/workspaces/%s/scheduling/assignments/recurring", c.workspaceID)
	var out clockify.ScheduleAssignment
	err := c.Do(ctx, "POST", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) UpdateAssignment(ctx context.Context, id string, req clockify.UpdateScheduleAssignmentRequest) (*clockify.ScheduleAssignment, error) {
	path := fmt.Sprintf("/workspaces/%s/scheduling/assignments/recurring/%s", c.workspaceID, id)
	var out clockify.ScheduleAssignment
	err := c.Do(ctx, "PATCH", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) DeleteAssignment(ctx context.Context, id, seriesOption string) error {
	path := fmt.Sprintf("/workspaces/%s/scheduling/assignments/recurring/%s", c.workspaceID, id)
	if seriesOption != "" {
		path = appendQueryParams(path, map[string]string{"seriesUpdateOption": seriesOption})
	}
	return c.Do(ctx, "DELETE", path, nil, nil)
}

// ---------- Time-off ----------

func (c *ClockifyClient) ListTimeOffPolicies(ctx context.Context) ([]clockify.TimeOffPolicy, error) {
	path := fmt.Sprintf("/workspaces/%s/time-off/policies", c.workspaceID)
	var policies []clockify.TimeOffPolicy
	err := c.Do(ctx, "GET", path, nil, &policies)
	return policies, err
}

func (c *ClockifyClient) ListTimeOffRequests(ctx context.Context, req clockify.ListTimeOffRequestsRequest) ([]clockify.TimeOffRequest, error) {
	path := fmt.Sprintf("/workspaces/%s/time-off/requests", c.workspaceID)
	var raw json.RawMessage
	if err := c.Do(ctx, "POST", path, req, &raw); err != nil {
		return nil, err
	}
	var arr []clockify.TimeOffRequest
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var wrapped clockify.TimeOffListResponse
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		return wrapped.Requests, nil
	}
	return nil, fmt.Errorf("unexpected time-off list response")
}

func (c *ClockifyClient) CreateTimeOffRequest(ctx context.Context, policyID string, req clockify.CreateTimeOffRequestRequest) (*clockify.TimeOffRequest, error) {
	path := fmt.Sprintf("/workspaces/%s/time-off/policies/%s/requests", c.workspaceID, policyID)
	var out clockify.TimeOffRequest
	err := c.Do(ctx, "POST", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) UpdateTimeOffStatus(ctx context.Context, policyID, requestID string, req clockify.UpdateTimeOffStatusRequest) (*clockify.TimeOffRequest, error) {
	path := fmt.Sprintf("/workspaces/%s/time-off/policies/%s/requests/%s", c.workspaceID, policyID, requestID)
	var out clockify.TimeOffRequest
	err := c.Do(ctx, "PATCH", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) DeleteTimeOffRequest(ctx context.Context, policyID, requestID string) error {
	path := fmt.Sprintf("/workspaces/%s/time-off/policies/%s/requests/%s", c.workspaceID, policyID, requestID)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

// ---------- Expense ----------

type ListExpensesOpts struct {
	ProjectID string
	UserID    string
	Start     string
	End       string
	Page      int
	PageSize  int
}

func (c *ClockifyClient) ListExpenses(ctx context.Context, opts ListExpensesOpts) ([]clockify.Expense, error) {
	path := fmt.Sprintf("/workspaces/%s/expenses", c.workspaceID)
	params := map[string]string{
		"project":  opts.ProjectID,
		"user-id":  opts.UserID,
		"start":    opts.Start,
		"end":      opts.End,
		"page":     strconv.Itoa(opts.Page),
		"pageSize": strconv.Itoa(opts.PageSize),
	}
	path = appendQueryParams(path, params)
	var raw json.RawMessage
	if err := c.Do(ctx, "GET", path, nil, &raw); err != nil {
		return nil, err
	}
	var arr []clockify.Expense
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var wrapped clockify.ExpenseListResponse
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		return wrapped.Expenses, nil
	}
	return nil, fmt.Errorf("unexpected expense list response")
}

func (c *ClockifyClient) ListExpenseCategories(ctx context.Context) ([]clockify.ExpenseCategory, error) {
	path := fmt.Sprintf("/workspaces/%s/expenses/categories", c.workspaceID)
	var cats []clockify.ExpenseCategory
	err := c.Do(ctx, "GET", path, nil, &cats)
	return cats, err
}

func (c *ClockifyClient) GetExpense(ctx context.Context, id string) (*clockify.Expense, error) {
	path := fmt.Sprintf("/workspaces/%s/expenses/%s", c.workspaceID, id)
	var out clockify.Expense
	err := c.Do(ctx, "GET", path, nil, &out)
	return &out, err
}

func (c *ClockifyClient) CreateExpense(ctx context.Context, req clockify.CreateExpenseRequest) (*clockify.Expense, error) {
	path := fmt.Sprintf("/workspaces/%s/expenses", c.workspaceID)
	var out clockify.Expense
	err := c.Do(ctx, "POST", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) UpdateExpense(ctx context.Context, id string, req clockify.UpdateExpenseRequest) (*clockify.Expense, error) {
	path := fmt.Sprintf("/workspaces/%s/expenses/%s", c.workspaceID, id)
	var out clockify.Expense
	err := c.Do(ctx, "PUT", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) DeleteExpense(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/expenses/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

// ---------- Custom fields ----------

type ListCustomFieldsOpts struct {
	Name     string
	Status   string
	Page     int
	PageSize int
}

func (c *ClockifyClient) ListCustomFields(ctx context.Context, opts ListCustomFieldsOpts) ([]clockify.CustomField, error) {
	path := fmt.Sprintf("/workspaces/%s/custom-fields", c.workspaceID)
	params := map[string]string{
		"name":     opts.Name,
		"status":   opts.Status,
		"page":     strconv.Itoa(opts.Page),
		"pageSize": strconv.Itoa(opts.PageSize),
	}
	path = appendQueryParams(path, params)
	var fields []clockify.CustomField
	err := c.Do(ctx, "GET", path, nil, &fields)
	return fields, err
}

func (c *ClockifyClient) CreateCustomField(ctx context.Context, req clockify.CreateCustomFieldRequest) (*clockify.CustomField, error) {
	path := fmt.Sprintf("/workspaces/%s/custom-fields", c.workspaceID)
	var out clockify.CustomField
	err := c.Do(ctx, "POST", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) UpdateCustomField(ctx context.Context, id string, req clockify.UpdateCustomFieldRequest) (*clockify.CustomField, error) {
	path := fmt.Sprintf("/workspaces/%s/custom-fields/%s", c.workspaceID, id)
	var out clockify.CustomField
	err := c.Do(ctx, "PUT", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) DeleteCustomField(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/custom-fields/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

// ---------- Webhooks ----------

type ListWebhooksOpts struct {
	Page     int
	PageSize int
}

func (c *ClockifyClient) ListWebhooks(ctx context.Context, opts ListWebhooksOpts) ([]clockify.Webhook, error) {
	path := fmt.Sprintf("/workspaces/%s/webhooks", c.workspaceID)
	params := map[string]string{
		"page":     strconv.Itoa(opts.Page),
		"pageSize": strconv.Itoa(opts.PageSize),
	}
	path = appendQueryParams(path, params)
	var hooks []clockify.Webhook
	err := c.Do(ctx, "GET", path, nil, &hooks)
	return hooks, err
}

func (c *ClockifyClient) GetWebhook(ctx context.Context, id string) (*clockify.Webhook, error) {
	path := fmt.Sprintf("/workspaces/%s/webhooks/%s", c.workspaceID, id)
	var out clockify.Webhook
	err := c.Do(ctx, "GET", path, nil, &out)
	return &out, err
}

func (c *ClockifyClient) CreateWebhook(ctx context.Context, req clockify.CreateWebhookRequest) (*clockify.Webhook, error) {
	path := fmt.Sprintf("/workspaces/%s/webhooks", c.workspaceID)
	var out clockify.Webhook
	err := c.Do(ctx, "POST", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) UpdateWebhook(ctx context.Context, id string, req clockify.UpdateWebhookRequest) (*clockify.Webhook, error) {
	path := fmt.Sprintf("/workspaces/%s/webhooks/%s", c.workspaceID, id)
	var out clockify.Webhook
	err := c.Do(ctx, "PUT", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) DeleteWebhook(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/webhooks/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

// ---------- Approval ----------

type ListApprovalRequestsOpts struct {
	Status   string
	Types    string
	Page     int
	PageSize int
}

func (c *ClockifyClient) ListApprovalRequests(ctx context.Context, opts ListApprovalRequestsOpts) ([]clockify.ApprovalRequest, error) {
	path := fmt.Sprintf("/workspaces/%s/approval-requests", c.workspaceID)
	params := map[string]string{
		"status":   opts.Status,
		"types":    opts.Types,
		"page":     strconv.Itoa(opts.Page),
		"pageSize": strconv.Itoa(opts.PageSize),
	}
	path = appendQueryParams(path, params)
	var raw json.RawMessage
	if err := c.Do(ctx, "GET", path, nil, &raw); err != nil {
		return nil, err
	}
	var arr []clockify.ApprovalRequest
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var wrapped struct {
		Requests []clockify.ApprovalRequest `json:"requests"`
		Items    []clockify.ApprovalRequest `json:"items"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		if wrapped.Requests != nil {
			return wrapped.Requests, nil
		}
		return wrapped.Items, nil
	}
	return nil, fmt.Errorf("unexpected approval list response")
}

func (c *ClockifyClient) SubmitApprovalRequest(ctx context.Context, approvalType string, req clockify.SubmitApprovalRequest) (*clockify.ApprovalRequest, error) {
	path := fmt.Sprintf("/workspaces/%s/approval-requests/%s", c.workspaceID, approvalType)
	var out clockify.ApprovalRequest
	err := c.Do(ctx, "POST", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) SubmitApprovalRequestForUser(ctx context.Context, userID, approvalType string, req clockify.SubmitApprovalRequest) (*clockify.ApprovalRequest, error) {
	path := fmt.Sprintf("/workspaces/%s/approval-requests/users/%s/%s", c.workspaceID, userID, approvalType)
	var out clockify.ApprovalRequest
	err := c.Do(ctx, "POST", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) UpdateApprovalStatus(ctx context.Context, id string, req clockify.UpdateApprovalStatusRequest) (*clockify.ApprovalRequest, error) {
	path := fmt.Sprintf("/workspaces/%s/approval-requests/%s", c.workspaceID, id)
	var out clockify.ApprovalRequest
	err := c.Do(ctx, "PATCH", path, req, &out)
	if err != nil {
		return nil, err
	}
	if out.Status == "" {
		out.Status = req.State
	}
	return &out, nil
}

// ---------- User groups ----------

type ListUserGroupsOpts struct {
	Name     string
	ProjectID string
	Page     int
	PageSize int
}

func (c *ClockifyClient) ListUserGroups(ctx context.Context, opts ListUserGroupsOpts) ([]clockify.UserGroup, error) {
	path := fmt.Sprintf("/workspaces/%s/user-groups", c.workspaceID)
	params := map[string]string{
		"name":      opts.Name,
		"project-id": opts.ProjectID,
		"page":      strconv.Itoa(opts.Page),
		"pageSize":  strconv.Itoa(opts.PageSize),
	}
	path = appendQueryParams(path, params)
	var groups []clockify.UserGroup
	err := c.Do(ctx, "GET", path, nil, &groups)
	return groups, err
}

func (c *ClockifyClient) GetUserGroup(ctx context.Context, id string) (*clockify.UserGroup, error) {
	path := fmt.Sprintf("/workspaces/%s/user-groups/%s", c.workspaceID, id)
	var out clockify.UserGroup
	err := c.Do(ctx, "GET", path, nil, &out)
	return &out, err
}

func (c *ClockifyClient) CreateUserGroup(ctx context.Context, req clockify.CreateUserGroupRequest) (*clockify.UserGroup, error) {
	path := fmt.Sprintf("/workspaces/%s/user-groups", c.workspaceID)
	var out clockify.UserGroup
	err := c.Do(ctx, "POST", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) UpdateUserGroup(ctx context.Context, id string, req clockify.UpdateUserGroupRequest) (*clockify.UserGroup, error) {
	path := fmt.Sprintf("/workspaces/%s/user-groups/%s", c.workspaceID, id)
	var out clockify.UserGroup
	err := c.Do(ctx, "PUT", path, req, &out)
	return &out, err
}

func (c *ClockifyClient) DeleteUserGroup(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/user-groups/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

func (c *ClockifyClient) AddUserToGroup(ctx context.Context, groupID, userID string) (*clockify.UserGroup, error) {
	path := fmt.Sprintf("/workspaces/%s/user-groups/%s/users", c.workspaceID, groupID)
	var out clockify.UserGroup
	err := c.Do(ctx, "POST", path, clockify.AddUserToGroupRequest{UserID: userID}, &out)
	return &out, err
}

func (c *ClockifyClient) RemoveUserFromGroup(ctx context.Context, groupID, userID string) error {
	path := fmt.Sprintf("/workspaces/%s/user-groups/%s/users/%s", c.workspaceID, groupID, userID)
	return c.Do(ctx, "DELETE", path, nil, nil)
}
