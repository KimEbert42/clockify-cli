package clockify

// M7 secondary resources — types per docs.clockify.me v1.

type Invoice struct {
	ID            string  `json:"id" yaml:"id"`
	WorkspaceID   string  `json:"workspaceId" yaml:"workspaceId"`
	ClientID      string  `json:"clientId" yaml:"clientId"`
	Number        string  `json:"number" yaml:"number"`
	Currency      string  `json:"currency" yaml:"currency"`
	IssuedDate    string  `json:"issuedDate" yaml:"issuedDate"`
	DueDate       string  `json:"dueDate" yaml:"dueDate"`
	InvoiceStatus string  `json:"invoiceStatus" yaml:"invoiceStatus"`
	Status        string  `json:"status" yaml:"status"`
	Note          string  `json:"note" yaml:"note"`
	Total         float64 `json:"total" yaml:"total"`
	TimeViewMode  string  `json:"timeViewMode" yaml:"timeViewMode"`
}

type CreateInvoiceRequest struct {
	ClientID     string  `json:"clientId"`
	Currency     string  `json:"currency"`
	IssuedDate   string  `json:"issuedDate"`
	DueDate      string  `json:"dueDate"`
	Number       string  `json:"number"`
	InvoiceStatus string `json:"invoiceStatus,omitempty"`
	Note         string  `json:"note,omitempty"`
	TimeViewMode string  `json:"timeViewMode,omitempty"`
}

type UpdateInvoiceRequest struct {
	ClientID      string  `json:"clientId,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	IssuedDate    string  `json:"issuedDate,omitempty"`
	DueDate       string  `json:"dueDate,omitempty"`
	Number        string  `json:"number,omitempty"`
	InvoiceStatus string  `json:"invoiceStatus,omitempty"`
	Note          string  `json:"note,omitempty"`
}

type ChangeInvoiceStatusRequest struct {
	InvoiceStatus string `json:"invoiceStatus"`
}

type InvoiceListResponse struct {
	Invoices []Invoice `json:"invoices"`
	Total    int       `json:"total"`
}

type ScheduleAssignment struct {
	ID                  string  `json:"id" yaml:"id"`
	WorkspaceID         string  `json:"workspaceId" yaml:"workspaceId"`
	ProjectID           string  `json:"projectId" yaml:"projectId"`
	ProjectName         string  `json:"projectName" yaml:"projectName"`
	ProjectColor        string  `json:"projectColor" yaml:"projectColor"`
	ProjectBillable     bool    `json:"projectBillable" yaml:"projectBillable"`
	ProjectArchived     bool    `json:"projectArchived" yaml:"projectArchived"`
	TaskID              string  `json:"taskId" yaml:"taskId"`
	TaskName            string  `json:"taskName" yaml:"taskName"`
	UserID              string  `json:"userId" yaml:"userId"`
	UserName            string  `json:"userName" yaml:"userName"`
	ClientID            string  `json:"clientId" yaml:"clientId"`
	ClientName          string  `json:"clientName" yaml:"clientName"`
	HoursPerDay         float64 `json:"hoursPerDay" yaml:"hoursPerDay"`
	StartTime           string  `json:"startTime" yaml:"startTime"`
	Billable            bool    `json:"billable" yaml:"billable"`
	Note                string  `json:"note" yaml:"note"`
	Published           bool    `json:"published" yaml:"published"`
	IncludeNonWorkingDays bool   `json:"includeNonWorkingDays" yaml:"includeNonWorkingDays"`
}

type CreateScheduleAssignmentRequest struct {
	ProjectID             string  `json:"projectId"`
	UserID                string  `json:"userId"`
	Start                 string  `json:"start"`
	End                   string  `json:"end"`
	HoursPerDay           float64 `json:"hoursPerDay"`
	Billable              bool    `json:"billable"`
	Note                  string  `json:"note,omitempty"`
	TaskID                string  `json:"taskId,omitempty"`
	StartTime             string  `json:"startTime,omitempty"`
	IncludeNonWorkingDays bool    `json:"includeNonWorkingDays,omitempty"`
}

type UpdateScheduleAssignmentRequest struct {
	ProjectID             *string  `json:"projectId,omitempty"`
	UserID                *string  `json:"userId,omitempty"`
	Start                 string   `json:"start,omitempty"`
	End                   string   `json:"end,omitempty"`
	HoursPerDay           *float64 `json:"hoursPerDay,omitempty"`
	Billable              *bool    `json:"billable,omitempty"`
	Note                  string   `json:"note,omitempty"`
	TaskID                string   `json:"taskId,omitempty"`
	StartTime             string   `json:"startTime,omitempty"`
	IncludeNonWorkingDays *bool    `json:"includeNonWorkingDays,omitempty"`
	SeriesUpdateOption    string   `json:"seriesUpdateOption,omitempty"`
}

type TimeOffRequest struct {
	ID          string `json:"id" yaml:"id"`
	WorkspaceID string `json:"workspaceId" yaml:"workspaceId"`
	UserID      string `json:"userId" yaml:"userId"`
	PolicyID    string `json:"policyId" yaml:"policyId"`
	Start       string `json:"start" yaml:"start"`
	End         string `json:"end" yaml:"end"`
	Status      string `json:"status" yaml:"status"`
	Note        string `json:"note" yaml:"note"`
	Description string `json:"description" yaml:"description"`
}

type TimeOffPolicy struct {
	ID          string `json:"id" yaml:"id"`
	WorkspaceID string `json:"workspaceId" yaml:"workspaceId"`
	Name        string `json:"name" yaml:"name"`
	Archived    bool   `json:"archived" yaml:"archived"`
}

type TimeOffListResponse struct {
	Count    int              `json:"count"`
	Requests []TimeOffRequest `json:"requests"`
}

type TimeOffRequestPeriod struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	TimeZone string `json:"timeZone,omitempty"`
}

type CreateTimeOffRequestRequest struct {
	Note          string              `json:"note"`
	TimeOffPeriod TimeOffRequestPeriod `json:"timeOffPeriod"`
}

type ListTimeOffRequestsRequest struct {
	Start      string   `json:"start"`
	End        string   `json:"end"`
	Statuses   []string `json:"statuses,omitempty"`
	Users      []string `json:"users,omitempty"`
	UserGroups []string `json:"userGroups,omitempty"`
	Page       int      `json:"page,omitempty"`
	PageSize   int      `json:"pageSize,omitempty"`
}

type UpdateTimeOffStatusRequest struct {
	Note   string `json:"note,omitempty"`
	Status string `json:"status"`
}

type Expense struct {
	ID          string  `json:"id" yaml:"id"`
	WorkspaceID string  `json:"workspaceId" yaml:"workspaceId"`
	ProjectID   string  `json:"projectId" yaml:"projectId"`
	TaskID      string  `json:"taskId" yaml:"taskId"`
	UserID      string  `json:"userId" yaml:"userId"`
	CategoryID  string  `json:"categoryId" yaml:"categoryId"`
	Date        string  `json:"date" yaml:"date"`
	Total       float64 `json:"total" yaml:"total"`
	Quantity    float64 `json:"quantity" yaml:"quantity"`
	Billable    bool    `json:"billable" yaml:"billable"`
	Notes       string  `json:"notes" yaml:"notes"`
	FileID      string  `json:"fileId" yaml:"fileId"`
	Locked      bool    `json:"locked" yaml:"locked"`
}

type ExpenseCategory struct {
	ID          string `json:"id" yaml:"id"`
	WorkspaceID string `json:"workspaceId" yaml:"workspaceId"`
	Name        string `json:"name" yaml:"name"`
	Archived    bool   `json:"archived" yaml:"archived"`
}

type ExpenseListResponse struct {
	Expenses     []Expense `json:"expenses"`
	DailyTotals  any       `json:"dailyTotals"`
	WeeklyTotals any       `json:"weeklyTotals"`
}

type CreateExpenseRequest struct {
	ProjectID  string  `json:"projectId"`
	TaskID     string  `json:"taskId,omitempty"`
	UserID     string  `json:"userId"`
	CategoryID string  `json:"categoryId"`
	Date       string  `json:"date"`
	Total      float64 `json:"total"`
	Quantity   float64 `json:"quantity,omitempty"`
	Billable   bool    `json:"billable"`
	Notes      string  `json:"notes,omitempty"`
}

type UpdateExpenseRequest struct {
	ProjectID  string  `json:"projectId"`
	TaskID     string  `json:"taskId,omitempty"`
	UserID     string  `json:"userId"`
	CategoryID string  `json:"categoryId"`
	Date       string  `json:"date"`
	Total      float64 `json:"total"`
	Quantity   float64 `json:"quantity,omitempty"`
	Billable   bool    `json:"billable"`
	Notes      string  `json:"notes,omitempty"`
}

type CustomField struct {
	ID                string   `json:"id" yaml:"id"`
	WorkspaceID       string   `json:"workspaceId" yaml:"workspaceId"`
	Name              string   `json:"name" yaml:"name"`
	Type              string   `json:"type" yaml:"type"`
	Status            string   `json:"status" yaml:"status"`
	AllowedValues     []string `json:"allowedValues" yaml:"allowedValues"`
	Description       string   `json:"description" yaml:"description"`
	OnlyAdminCanEdit  bool     `json:"onlyAdminCanEdit" yaml:"onlyAdminCanEdit"`
	Placeholder       string   `json:"placeholder" yaml:"placeholder"`
}

type CreateCustomFieldRequest struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	AllowedValues    []string `json:"allowedValues,omitempty"`
	Status           string   `json:"status,omitempty"`
	OnlyAdminCanEdit bool     `json:"onlyAdminCanEdit,omitempty"`
	Placeholder      string   `json:"placeholder,omitempty"`
	Description      string   `json:"description,omitempty"`
}

type UpdateCustomFieldRequest struct {
	Name             string   `json:"name"`
	Type             string   `json:"type,omitempty"`
	AllowedValues    []string `json:"allowedValues,omitempty"`
	Status           string   `json:"status,omitempty"`
	OnlyAdminCanEdit bool     `json:"onlyAdminCanEdit,omitempty"`
	Placeholder      string   `json:"placeholder,omitempty"`
	Description      string   `json:"description,omitempty"`
}

type Webhook struct {
	ID              string   `json:"id" yaml:"id"`
	WorkspaceID     string   `json:"workspaceId" yaml:"workspaceId"`
	Name            string   `json:"name" yaml:"name"`
	URL             string   `json:"url" yaml:"url"`
	WebhookEvent    string   `json:"webhookEvent" yaml:"webhookEvent"`
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	DeliveryEnabled bool     `json:"deliveryEnabled" yaml:"deliveryEnabled"`
	UserID          string   `json:"userId" yaml:"userId"`
	TriggerSource   []string `json:"triggerSource" yaml:"triggerSource"`
	TriggerSourceType string `json:"triggerSourceType" yaml:"triggerSourceType"`
	PlanEnabled     bool     `json:"planEnabled" yaml:"planEnabled"`
}

type CreateWebhookRequest struct {
	Name              string   `json:"name,omitempty"`
	URL               string   `json:"url"`
	WebhookEvent      string   `json:"webhookEvent"`
	Enabled           bool     `json:"enabled"`
	TriggerSource     []string `json:"triggerSource,omitempty"`
	TriggerSourceType string   `json:"triggerSourceType,omitempty"`
}

type UpdateWebhookRequest struct {
	Name              string   `json:"name,omitempty"`
	URL               string   `json:"url,omitempty"`
	WebhookEvent      string   `json:"webhookEvent,omitempty"`
	Enabled           *bool    `json:"enabled,omitempty"`
	TriggerSource     []string `json:"triggerSource,omitempty"`
	TriggerSourceType string   `json:"triggerSourceType,omitempty"`
}

type ApprovalRequest struct {
	ID          string `json:"id" yaml:"id"`
	WorkspaceID string `json:"workspaceId" yaml:"workspaceId"`
	UserID      string `json:"userId" yaml:"userId"`
	Status      string `json:"status" yaml:"status"`
	Type        string `json:"type" yaml:"type"`
	Start       string `json:"start" yaml:"start"`
	End         string `json:"end" yaml:"end"`
	Note        string `json:"note" yaml:"note"`
	CreatedAt   string `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   string `json:"updatedAt" yaml:"updatedAt"`
}

type SubmitApprovalRequest struct {
	Period      string `json:"period"`
	PeriodStart string `json:"periodStart"`
	Note        string `json:"note,omitempty"`
	Type        string `json:"type,omitempty"`
}

type UpdateApprovalStatusRequest struct {
	Note  string `json:"note,omitempty"`
	State string `json:"state"`
}

type UserGroup struct {
	ID           string   `json:"id" yaml:"id"`
	WorkspaceID  string   `json:"workspaceId" yaml:"workspaceId"`
	Name         string   `json:"name" yaml:"name"`
	UserIDs      []string `json:"userIds" yaml:"userIds"`
	TeamManagers []struct {
		ID   string `json:"id" yaml:"id"`
		Name string `json:"name" yaml:"name"`
	} `json:"teamManagers" yaml:"teamManagers"`
}

type CreateUserGroupRequest struct {
	Name string `json:"name"`
}

type UpdateUserGroupRequest struct {
	Name string `json:"name"`
}

type AddUserToGroupRequest struct {
	UserID string `json:"userId"`
}
