package clockify

// Report filter enums (Clockify Reports API v1).
const (
	ReportExportJSON   = "JSON"
	ReportExportJSONV1 = "JSON_V1"
	ReportExportCSV    = "CSV"
	ReportExportPDF    = "PDF"
	ReportExportXLSX   = "XLSX"
	ReportExportZIP    = "ZIP"

	ReportApprovalApproved   = "APPROVED"
	ReportApprovalUnapproved = "UNAPPROVED"
	ReportApprovalAll        = "ALL"

	ReportContains        = "CONTAINS"
	ReportContainsOnly    = "CONTAINS_ONLY"
	ReportStatusActive    = "ACTIVE"
	ReportSortAscending   = "ASCENDING"
	ReportSortDescending  = "DESCENDING"
	ReportWeekStartMonday = "MONDAY"
)

// ContainsFilter is used for projects, clients, users, userGroups, tasks.
type ContainsFilter struct {
	Contains string   `json:"contains,omitempty"`
	IDs      []string `json:"ids,omitempty"`
	Status   string   `json:"status,omitempty"`
}

// TagFilter extends ContainsFilter with containedInTimeentry.
type TagFilter struct {
	ContainsFilter
	ContainedInTimeentry string `json:"containedInTimeentry,omitempty"`
}

// DetailedFilter is required for detailed reports.
type DetailedFilter struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	SortColumn string `json:"sortColumn,omitempty"`
	Options  *DetailedOptions `json:"options,omitempty"`
}

type DetailedOptions struct {
	Totals string `json:"totals,omitempty"` // e.g. CALCULATE
}

// SummaryFilter is required for summary reports.
type SummaryFilter struct {
	Groups           []string `json:"groups,omitempty"`
	SortColumn       string   `json:"sortColumn,omitempty"`
	SummaryChartType string   `json:"summaryChartType,omitempty"` // e.g. PROJECT
}

// WeeklyFilter is required for weekly reports.
type WeeklyFilter struct {
	Group    string `json:"group,omitempty"`
	Subgroup string `json:"subgroup,omitempty"`
}

// ReportRequest is the shared body for detailed/summary/weekly report generation.
// Per-endpoint required filters are set by the caller (DetailedFilter / SummaryFilter / WeeklyFilter).
type ReportRequest struct {
	DateRangeStart string `json:"dateRangeStart"`
	DateRangeEnd   string `json:"dateRangeEnd"`

	DateRangeType string `json:"dateRangeType,omitempty"`
	DateFormat    string `json:"dateFormat,omitempty"`
	TimeFormat    string `json:"timeFormat,omitempty"`
	TimeZone      string `json:"timeZone,omitempty"`
	WeekStart     string `json:"weekStart,omitempty"`

	Description     string `json:"description,omitempty"`
	WithoutDescription *bool `json:"withoutDescription,omitempty"`
	Billable        *bool  `json:"billable,omitempty"`

	Projects  *ContainsFilter `json:"projects,omitempty"`
	Clients   *ContainsFilter `json:"clients,omitempty"`
	Users     *ContainsFilter `json:"users,omitempty"`
	UserGroups *ContainsFilter `json:"userGroups,omitempty"`
	Tasks     *ContainsFilter `json:"tasks,omitempty"`
	Tags      *TagFilter      `json:"tags,omitempty"`

	ApprovalState  string `json:"approvalState,omitempty"`
	InvoicingState string `json:"invoicingState,omitempty"`
	ExportType     string `json:"exportType,omitempty"`
	SortOrder      string `json:"sortOrder,omitempty"`
	Rounding       *bool  `json:"rounding,omitempty"`
	ZoomLevel      string `json:"zoomLevel,omitempty"`
	Archived       *bool  `json:"archived,omitempty"`

	DetailedFilter *DetailedFilter `json:"detailedFilter,omitempty"`
	SummaryFilter  *SummaryFilter  `json:"summaryFilter,omitempty"`
	WeeklyFilter   *WeeklyFilter   `json:"weeklyFilter,omitempty"`
}

// ReportTimeInterval mirrors time entry intervals in report payloads.
type ReportTimeInterval struct {
	Start    string  `json:"start,omitempty"`
	End      string  `json:"end,omitempty"`
	Duration *int64  `json:"duration,omitempty"`
}

// ReportEntry is a flexible time-entry row in report responses.
// Field names vary between Clockify response versions (id vs _id).
type ReportEntry struct {
	ID          string              `json:"id,omitempty"`
	IDAlt       string              `json:"_id,omitempty"`
	Description string              `json:"description,omitempty"`
	UserID      string              `json:"userId,omitempty"`
	ProjectID   string              `json:"projectId,omitempty"`
	TaskID      string              `json:"taskId,omitempty"`
	Billable    bool                `json:"billable,omitempty"`
	TimeInterval *ReportTimeInterval `json:"timeInterval,omitempty"`
}

func (e ReportEntry) EntryID() string {
	if e.ID != "" {
		return e.ID
	}
	return e.IDAlt
}

// ReportTotals holds duration/amount aggregates when present.
type ReportTotals struct {
	Duration         *int64 `json:"duration,omitempty"`
	BillableDuration *int64 `json:"billableDuration,omitempty"`
	Amount           *float64 `json:"amount,omitempty"`
}

// ReportGroup is a summary/weekly grouping (project, user, client, day, etc.).
type ReportGroup struct {
	ID          string        `json:"id,omitempty"`
	Name        string        `json:"name,omitempty"`
	Date        string        `json:"date,omitempty"`
	TimeEntries []ReportEntry `json:"timeentries,omitempty"`
	Totals      *ReportTotals `json:"totals,omitempty"`
}

// ReportResponse is a best-effort typed view of reports API output.
// Unknown fields are preserved via Raw for JSON pipelines.
type ReportResponse struct {
	Groups     []ReportGroup  `json:"groups,omitempty"`
	Days       []ReportGroup  `json:"days,omitempty"`
	TimeEntries []ReportEntry `json:"timeentries,omitempty"`
	Totals     *ReportTotals  `json:"totals,omitempty"`
	Raw        map[string]any `json:"-"`
}
