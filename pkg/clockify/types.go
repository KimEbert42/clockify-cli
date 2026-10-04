package clockify

import "time"

type Workspace struct {
	ID                string  `json:"id" yaml:"id"`
	Name              string  `json:"name" yaml:"name"`
	HourlyRate        Rate    `json:"hourlyRate" yaml:"hourlyRate"`
	Currency          string  `json:"currency" yaml:"currency"`
	IsMember          bool    `json:"isMember" yaml:"isMember"`
	MembershipType    string  `json:"membershipType" yaml:"membershipType"`
	WorkspaceSettings Settings `json:"workspaceSettings" yaml:"workspaceSettings"`
}

type Rate struct {
	Amount   int    `json:"amount" yaml:"amount"`
	Currency string `json:"currency" yaml:"currency"`
}

type Settings struct {
	AdminOnlyPage         bool   `json:"adminOnlyPage" yaml:"adminOnlyPage"`
	AdminOnlyTagging      bool   `json:"adminOnlyTagging" yaml:"adminOnlyTagging"`
	AdminOnlyCreatingTag  bool   `json:"adminOnlyCreatingTag" yaml:"adminOnlyCreatingTag"`
	AdminOnlyCreatingProject bool `json:"adminOnlyCreatingProject" yaml:"adminOnlyCreatingProject"`
	AllowTrackTimeWithTimer bool `json:"allowTrackTimeWithTimer" yaml:"allowTrackTimeWithTimer"`
	WeekStart             string `json:"weekStart" yaml:"weekStart"`
	DateFormat            string `json:"dateFormat" yaml:"dateFormat"`
	TimeFormat            string `json:"timeFormat" yaml:"timeFormat"`
}

type User struct {
	ID             string         `json:"id" yaml:"id"`
	Email          string         `json:"email" yaml:"email"`
	Name           string         `json:"name" yaml:"name"`
	Status         string         `json:"status" yaml:"status"`
	Profile        UserProfile    `json:"profile" yaml:"profile"`
	ActiveWorkspace string       `json:"activeWorkspace" yaml:"activeWorkspace"`
	DefaultWorkspace string     `json:"defaultWorkspace" yaml:"defaultWorkspace"`
}

type UserProfile struct {
	WeekStart         string `json:"weekStart" yaml:"weekStart"`
	TimeZone          string `json:"timeZone" yaml:"timeZone"`
	DateFormat        string `json:"dateFormat" yaml:"dateFormat"`
	TimeFormat        string `json:"timeFormat" yaml:"timeFormat"`
	StartPage         string `json:"startPage" yaml:"startPage"`
	Billable          bool   `json:"billable" yaml:"billable"`
	Locale            string `json:"locale" yaml:"locale"`
}

type Project struct {
	ID              string        `json:"id" yaml:"id"`
	Name            string        `json:"name" yaml:"name"`
	ClientID        string        `json:"clientId" yaml:"clientId"`
	WorkspaceID     string        `json:"workspaceId" yaml:"workspaceId"`
	Billable        bool          `json:"billable" yaml:"billable"`
	Status          string        `json:"status" yaml:"status"`
	Archived        bool          `json:"archived" yaml:"archived"`
	Duration        *int64        `json:"duration" yaml:"duration"`
	Estimate        *Estimate     `json:"estimate" yaml:"estimate"`
	ActiveTaskCount int           `json:"activeTaskCount" yaml:"activeTaskCount"`
	Color           string        `json:"color" yaml:"color"`
	NameSimilarTo   string        `json:"nameSimilarTo" yaml:"nameSimilarTo"`
	Note            string        `json:"note" yaml:"note"`
}

type Estimate struct {
	Type     string `json:"type" yaml:"type"`
	Estimate int64  `json:"estimate" yaml:"estimate"`
}

type Client struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	WorkspaceID string `json:"workspaceId" yaml:"workspaceId"`
	Email       string `json:"email" yaml:"email"`
	Address     string `json:"address" yaml:"address"`
	Phone       string `json:"phone" yaml:"phone"`
	Note        string `json:"note" yaml:"note"`
}

type Tag struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	WorkspaceID string `json:"workspaceId" yaml:"workspaceId"`
	Color       string `json:"color" yaml:"color"`
}

type TimeEntry struct {
	ID              string         `json:"id" yaml:"id"`
	Description     string         `json:"description" yaml:"description"`
	WorkspaceID     string         `json:"workspaceId" yaml:"workspaceId"`
	ProjectID       string         `json:"projectId" yaml:"projectId"`
	TaskID          string         `json:"taskId" yaml:"taskId"`
	UserID          string         `json:"userId" yaml:"userId"`
	Billable        bool           `json:"billable" yaml:"billable"`
	TimeInterval    TimeInterval   `json:"timeInterval" yaml:"timeInterval"`
	TagIDs          []string       `json:"tagIds" yaml:"tagIds"`
}

type TimeInterval struct {
	Start    time.Time  `json:"start" yaml:"start"`
	End      *time.Time `json:"end" yaml:"end"`
	Duration int64      `json:"duration" yaml:"duration"`
}

type PageInfo struct {
	Page         int `json:"page" yaml:"page"`
	PageSize     int `json:"pageSize" yaml:"pageSize"`
	TotalPages   int `json:"totalPages" yaml:"totalPages"`
	TotalResults int `json:"totalResults" yaml:"totalResults"`
}

type CreateProjectRequest struct {
	Name     string    `json:"name"`
	ClientID string    `json:"clientId,omitempty"`
	Billable bool      `json:"billable"`
	Color    string    `json:"color,omitempty"`
	Note     string    `json:"note,omitempty"`
	Estimate *Estimate `json:"estimate,omitempty"`
}

type UpdateProjectRequest struct {
	Name     string `json:"name,omitempty"`
	ClientID string `json:"clientId,omitempty"`
	Billable *bool  `json:"billable,omitempty"`
	Color    string `json:"color,omitempty"`
	Note     string `json:"note,omitempty"`
	Archived *bool  `json:"archived,omitempty"`
}

type CreateClientRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email,omitempty"`
	Address string `json:"address,omitempty"`
	Note    string `json:"note,omitempty"`
}

type UpdateClientRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Address  string `json:"address,omitempty"`
	Note     string `json:"note,omitempty"`
	Archived *bool  `json:"archived,omitempty"`
}

type CreateTagRequest struct {
	Name string `json:"name"`
}

type UpdateTagRequest struct {
	Name     string `json:"name,omitempty"`
	Archived *bool  `json:"archived,omitempty"`
}

type Task struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	ProjectID   string    `json:"projectId" yaml:"projectId"`
	WorkspaceID string    `json:"workspaceId" yaml:"workspaceId"`
	AssigneeIDs []string  `json:"assigneeIds" yaml:"assigneeIds"`
	Estimate    *Estimate `json:"estimate" yaml:"estimate"`
	Status      string    `json:"status" yaml:"status"`
	IsBillable  bool      `json:"isBillable" yaml:"isBillable"`
	Active      bool      `json:"active" yaml:"active"`
}

type CreateTaskRequest struct {
	Name        string    `json:"name"`
	Estimate    *Estimate `json:"estimate,omitempty"`
	Status      string    `json:"status,omitempty"`
	AssigneeIDs []string  `json:"assigneeIds,omitempty"`
	IsBillable  bool      `json:"isBillable,omitempty"`
}

type UpdateTaskRequest struct {
	Name        string    `json:"name,omitempty"`
	Estimate    *Estimate `json:"estimate,omitempty"`
	Status      string    `json:"status,omitempty"`
	AssigneeIDs []string  `json:"assigneeIds,omitempty"`
	IsBillable  *bool     `json:"isBillable,omitempty"`
}
