package clockify

import (
	"fmt"
	"strings"
)

// Official custom-field type enum (docs.clockify.me).
var officialCustomFieldTypes = map[string]bool{
	"TXT":              true,
	"NUMBER":           true,
	"DROPDOWN_SINGLE":  true,
	"DROPDOWN_MULTIPLE": true,
	"CHECKBOX":         true,
	"LINK":             true,
}

// Spec/CLI-friendly names → official API types.
var customFieldTypeAliases = map[string]string{
	"TEXT":             "TXT",
	"TXT":              "TXT",
	"NUMBER":           "NUMBER",
	"DROPDOWN":         "DROPDOWN_SINGLE",
	"DROPDOWN_SINGLE":  "DROPDOWN_SINGLE",
	"DROPDOWN_MULTIPLE": "DROPDOWN_MULTIPLE",
	"CHECKBOX":         "CHECKBOX",
	"LINK":             "LINK",
}

// MapCustomFieldType maps a CLI/spec custom-field type to the official API type.
func MapCustomFieldType(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if key == "DATE" {
		return "", fmt.Errorf("custom field type DATE is not supported by the Clockify API; supported types: TXT, NUMBER, DROPDOWN, DROPDOWN_SINGLE, DROPDOWN_MULTIPLE, CHECKBOX, LINK")
	}
	if mapped, ok := customFieldTypeAliases[key]; ok {
		return mapped, nil
	}
	return "", fmt.Errorf("invalid custom field type %q; supported: TEXT, NUMBER, DROPDOWN, DROPDOWN_SINGLE, DROPDOWN_MULTIPLE, CHECKBOX, LINK", s)
}

// Invoice status: spec names + official API names.
var invoiceStatusAliases = map[string]string{
	// Spec names
	"DRAFT":     "UNSENT",
	"FINALIZED": "UNSENT",
	"SENT":      "SENT",
	"ACCEPTED":  "PAID",
	"PAID":      "PAID",
	"VOID":      "VOID",
	// Official API names
	"UNSENT":         "UNSENT",
	"PARTIALLY_PAID": "PARTIALLY_PAID",
	"OVERDUE":        "OVERDUE",
}

// MapInvoiceStatus maps CLI/spec invoice status to the official API status.
func MapInvoiceStatus(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if mapped, ok := invoiceStatusAliases[key]; ok {
		return mapped, nil
	}
	return "", fmt.Errorf("invalid invoice status %q; supported: DRAFT, FINALIZED, SENT, ACCEPTED, PAID, VOID, UNSENT, PARTIALLY_PAID, OVERDUE", s)
}

// Official webhook event enum.
var officialWebhookEvents = map[string]bool{
	"NEW_PROJECT": true, "PROJECT_UPDATED": true, "PROJECT_DELETED": true,
	"NEW_TASK": true, "TASK_UPDATED": true, "TASK_DELETED": true,
	"NEW_CLIENT": true, "CLIENT_UPDATED": true, "CLIENT_DELETED": true,
	"NEW_TIMER_STARTED": true, "TIMER_STOPPED": true,
	"NEW_TIME_ENTRY": true, "TIME_ENTRY_UPDATED": true, "TIME_ENTRY_DELETED": true,
	"TIME_ENTRY_SPLIT": true, "TIME_ENTRY_RESTORED": true,
	"NEW_TAG": true, "TAG_UPDATED": true, "TAG_DELETED": true,
	"USER_JOINED_WORKSPACE": true, "USER_DELETED_FROM_WORKSPACE": true,
	"USER_DEACTIVATED_ON_WORKSPACE": true, "USER_ACTIVATED_ON_WORKSPACE": true,
	"USER_EMAIL_CHANGED": true, "USER_UPDATED": true,
	"NEW_INVOICE": true, "INVOICE_UPDATED": true,
	"NEW_APPROVAL_REQUEST": true, "APPROVAL_REQUEST_STATUS_UPDATED": true,
	"TIME_OFF_REQUESTED": true, "TIME_OFF_REQUEST_UPDATED": true,
	"TIME_OFF_REQUEST_APPROVED": true, "TIME_OFF_REQUEST_REJECTED": true,
	"TIME_OFF_REQUEST_STARTED": true, "TIME_OFF_REQUEST_WITHDRAWN": true,
	"BALANCE_UPDATED": true,
	"EXPENSE_CREATED": true, "EXPENSE_UPDATED": true, "EXPENSE_DELETED": true, "EXPENSE_RESTORED": true,
	"ASSIGNMENT_CREATED": true, "ASSIGNMENT_UPDATED": true, "ASSIGNMENT_DELETED": true, "ASSIGNMENT_PUBLISHED": true,
	"USER_GROUP_CREATED": true, "USER_GROUP_UPDATED": true, "USER_GROUP_DELETED": true,
	"USERS_INVITED_TO_WORKSPACE": true, "LIMITED_USERS_ADDED_TO_WORKSPACE": true,
	"COST_RATE_UPDATED": true, "BILLABLE_RATE_UPDATED": true,
}

// Spec webhook event aliases → official events.
var webhookEventAliases = map[string]string{
	"TIME_ENTRY_CREATED": "NEW_TIME_ENTRY",
	"PROJECT_CREATED":    "NEW_PROJECT",
	"TASK_CREATED":       "NEW_TASK",
	"CLIENT_CREATED":     "NEW_CLIENT",
	"TAG_CREATED":        "NEW_TAG",
	"INVOICE_CREATED":    "NEW_INVOICE",
	"USER_CREATED":       "USER_JOINED_WORKSPACE",
}

// MapWebhookEvent validates/aliases a webhook event name.
func MapWebhookEvent(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if mapped, ok := webhookEventAliases[key]; ok {
		return mapped, nil
	}
	if officialWebhookEvents[key] {
		return key, nil
	}
	return "", fmt.Errorf("invalid webhook event %q; see 'clockify webhook create --help' for supported events", s)
}

// MapWebhookEvents validates a comma-separated event list.
func MapWebhookEvents(s string) ([]string, error) {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		mapped, err := MapWebhookEvent(p)
		if err != nil {
			return nil, err
		}
		if !seen[mapped] {
			out = append(out, mapped)
			seen[mapped] = true
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one webhook event is required")
	}
	return out, nil
}

// ValidWebhookEvents returns the official event names (sorted stable order for help text).
func ValidWebhookEvents() []string {
	return []string{
		"NEW_PROJECT", "PROJECT_UPDATED", "PROJECT_DELETED",
		"NEW_TASK", "TASK_UPDATED", "TASK_DELETED",
		"NEW_CLIENT", "CLIENT_UPDATED", "CLIENT_DELETED",
		"NEW_TIMER_STARTED", "TIMER_STOPPED",
		"NEW_TIME_ENTRY", "TIME_ENTRY_UPDATED", "TIME_ENTRY_DELETED",
		"TIME_ENTRY_SPLIT", "TIME_ENTRY_RESTORED",
		"NEW_TAG", "TAG_UPDATED", "TAG_DELETED",
		"USER_JOINED_WORKSPACE", "USER_DELETED_FROM_WORKSPACE",
		"USER_DEACTIVATED_ON_WORKSPACE", "USER_ACTIVATED_ON_WORKSPACE",
		"USER_EMAIL_CHANGED", "USER_UPDATED",
		"NEW_INVOICE", "INVOICE_UPDATED",
		"NEW_APPROVAL_REQUEST", "APPROVAL_REQUEST_STATUS_UPDATED",
		"TIME_OFF_REQUESTED", "TIME_OFF_REQUEST_UPDATED",
		"TIME_OFF_REQUEST_APPROVED", "TIME_OFF_REQUEST_REJECTED",
		"TIME_OFF_REQUEST_STARTED", "TIME_OFF_REQUEST_WITHDRAWN",
		"BALANCE_UPDATED",
		"EXPENSE_CREATED", "EXPENSE_UPDATED", "EXPENSE_DELETED", "EXPENSE_RESTORED",
		"ASSIGNMENT_CREATED", "ASSIGNMENT_UPDATED", "ASSIGNMENT_DELETED", "ASSIGNMENT_PUBLISHED",
		"USER_GROUP_CREATED", "USER_GROUP_UPDATED", "USER_GROUP_DELETED",
		"USERS_INVITED_TO_WORKSPACE", "LIMITED_USERS_ADDED_TO_WORKSPACE",
		"COST_RATE_UPDATED", "BILLABLE_RATE_UPDATED",
	}
}

// Approval states accepted by the official update endpoint.
var approvalStates = map[string]bool{
	"PENDING":              true,
	"APPROVED":             true,
	"WITHDRAWN_SUBMISSION": true,
	"WITHDRAWN_APPROVAL":   true,
	"REJECTED":             true,
}

// MapApprovalState validates an approval state.
func MapApprovalState(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	// Spec used PENDING for list filter; official list uses PENDING, APPROVED, WITHDRAWN_APPROVAL
	if key == "NEW" || key == "IN_PROGRESS" {
		return "PENDING", nil
	}
	if approvalStates[key] {
		return key, nil
	}
	return "", fmt.Errorf("invalid approval status %q; supported: PENDING, APPROVED, REJECTED, WITHDRAWN_SUBMISSION, WITHDRAWN_APPROVAL", s)
}

// Approval list statuses (official GET filter enum).
var approvalListStatuses = map[string]bool{
	"PENDING":            true,
	"APPROVED":           true,
	"WITHDRAWN_APPROVAL": true,
}

// MapApprovalListStatus validates a list filter status.
func MapApprovalListStatus(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if key == "REJECTED" || key == "NEW" || key == "IN_PROGRESS" {
		// REJECTED is not a list filter enum; still allow filtering client-side.
		return key, nil
	}
	if approvalListStatuses[key] {
		return key, nil
	}
	return "", fmt.Errorf("invalid approval list status %q; supported: PENDING, APPROVED, WITHDRAWN_APPROVAL", s)
}

// Time-off statuses.
var timeOffStatuses = map[string]bool{
	"PENDING":  true,
	"APPROVED": true,
	"REJECTED": true,
}

// MapTimeOffStatus validates a time-off status.
func MapTimeOffStatus(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if timeOffStatuses[key] {
		return key, nil
	}
	return "", fmt.Errorf("invalid time-off status %q; supported: PENDING, APPROVED, REJECTED", s)
}

// Spec time-off type labels (display/filter only; API is policy-scoped).
var timeOffTypes = map[string]bool{
	"VACATION": true,
	"SICK":     true,
	"PERSONAL": true,
	"UNPAID":   true,
	"OTHER":    true,
}

// ValidTimeOffTypes returns spec type labels.
func ValidTimeOffTypes() []string {
	return []string{"VACATION", "SICK", "PERSONAL", "UNPAID", "OTHER"}
}

// ValidateTimeOffType validates a spec type label.
func ValidateTimeOffType(s string) error {
	key := strings.ToUpper(strings.TrimSpace(s))
	if timeOffTypes[key] {
		return nil
	}
	return fmt.Errorf("invalid time-off type %q; supported: %s", s, strings.Join(ValidTimeOffTypes(), ", "))
}

// Approval submit period values.
var approvalPeriods = map[string]bool{
	"WEEKLY":  true,
	"MONTHLY": true,
}

// MapApprovalPeriod validates an approval period.
func MapApprovalPeriod(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if approvalPeriods[key] {
		return key, nil
	}
	return "", fmt.Errorf("invalid approval period %q; supported: WEEKLY, MONTHLY", s)
}

// Approval request types.
var approvalTypes = map[string]bool{
	"TIMESHEET":            true,
	"EXPENSE":              true,
	"TIMESHEET_AND_EXPENSE": true,
}

// MapApprovalType validates an approval type.
func MapApprovalType(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if approvalTypes[key] {
		return key, nil
	}
	return "", fmt.Errorf("invalid approval type %q; supported: TIMESHEET, EXPENSE, TIMESHEET_AND_EXPENSE", s)
}

// Schedule series update options.
var seriesUpdateOptions = map[string]bool{
	"THIS_ONE":            true,
	"THIS_AND_FOLLOWING":  true,
	"ALL":                 true,
}

// MapSeriesUpdateOption validates a series option.
func MapSeriesUpdateOption(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if key == "" {
		return "THIS_ONE", nil
	}
	if seriesUpdateOptions[key] {
		return key, nil
	}
	return "", fmt.Errorf("invalid series option %q; supported: THIS_ONE, THIS_AND_FOLLOWING, ALL", s)
}

// Custom field statuses.
var customFieldStatuses = map[string]bool{
	"VISIBLE":   true,
	"INVISIBLE": true,
	"INACTIVE":  true,
}

// MapCustomFieldStatus validates a custom field status.
func MapCustomFieldStatus(s string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(s))
	if key == "" {
		return "", nil
	}
	if customFieldStatuses[key] {
		return key, nil
	}
	return "", fmt.Errorf("invalid custom field status %q; supported: VISIBLE, INVISIBLE, INACTIVE", s)
}
