package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kim/clockify-cli/pkg/clockify"
)

type ListProjectsOpts struct {
	Status string
	ClientID string
	Search string
	Page    int
	PageSize int
}

type ListClientsOpts struct {
	Search string
	Page    int
	PageSize int
}

type ListTagsOpts struct {
	Search string
	Page    int
	PageSize int
}

type ListUsersOpts struct {
	Status string
	Page    int
	PageSize int
}

type ListTasksOpts struct {
	IsActive  *bool
	Page      int
	PageSize  int
}

type ListTimeEntriesOpts struct {
	Start       string
	End         string
	ProjectID   string
	TaskID      string
	Tags        string
	Description string
	InProgress  string
	Hydrated    bool
	Page        int
	PageSize    int
}

type CreateTimeEntryRequest struct {
	Start       string   `json:"start"`
	End         string   `json:"end,omitempty"`
	ProjectID   string   `json:"projectId,omitempty"`
	TaskID      string   `json:"taskId,omitempty"`
	Description string   `json:"description,omitempty"`
	Billable    bool     `json:"billable"`
	TagIDs      []string `json:"tagIds,omitempty"`
	Type        string   `json:"type,omitempty"`
}

type UpdateTimeEntryRequest struct {
	Start       string   `json:"start"`
	End         string   `json:"end,omitempty"`
	ProjectID   string   `json:"projectId,omitempty"`
	TaskID      string   `json:"taskId,omitempty"`
	Description string   `json:"description,omitempty"`
	Billable    bool     `json:"billable"`
	TagIDs      []string `json:"tagIds,omitempty"`
}

type StopTimerRequest struct {
	End string `json:"end"`
}

func (c *ClockifyClient) ListWorkspaces(ctx context.Context) ([]clockify.Workspace, error) {
	var workspaces []clockify.Workspace
	err := c.Do(ctx, "GET", "/workspaces", nil, &workspaces)
	return workspaces, err
}

func (c *ClockifyClient) GetWorkspace(ctx context.Context, id string) (*clockify.Workspace, error) {
	var ws clockify.Workspace
	err := c.Do(ctx, "GET", "/workspaces/"+id, nil, &ws)
	return &ws, err
}

func (c *ClockifyClient) GetCurrentUser(ctx context.Context) (*clockify.User, error) {
	var user clockify.User
	err := c.Do(ctx, "GET", "/user", nil, &user)
	return &user, err
}

func (c *ClockifyClient) ListUsers(ctx context.Context, opts ListUsersOpts) ([]clockify.User, error) {
	path := fmt.Sprintf("/workspaces/%s/users", c.workspaceID)
	params := map[string]string{"status": opts.Status}
	if opts.Page > 0 {
		params["page"] = strconv.Itoa(opts.Page)
	}
	if opts.PageSize > 0 {
		params["pageSize"] = strconv.Itoa(opts.PageSize)
	}
	path = appendQueryParams(path, params)

	var users []clockify.User
	err := c.Do(ctx, "GET", path, nil, &users)
	return users, err
}

func (c *ClockifyClient) ListProjects(ctx context.Context, opts ListProjectsOpts) ([]clockify.Project, error) {
	projects, _, err := c.ListProjectsPaged(ctx, opts)
	return projects, err
}

// ListProjectsPaged fetches one page of projects and returns header-derived PageInfo.
func (c *ClockifyClient) ListProjectsPaged(ctx context.Context, opts ListProjectsOpts) ([]clockify.Project, *clockify.PageInfo, error) {
	path := fmt.Sprintf("/workspaces/%s/projects", c.workspaceID)
	params := map[string]string{
		"status":   opts.Status,
		"clientId": opts.ClientID,
		"search":   opts.Search,
	}
	if opts.Page > 0 {
		params["page"] = strconv.Itoa(opts.Page)
	}
	if opts.PageSize > 0 {
		params["pageSize"] = strconv.Itoa(opts.PageSize)
	}
	path = appendQueryParams(path, params)

	var projects []clockify.Project
	pageInfo, err := c.DoWithPagination(ctx, "GET", path, nil, &projects)
	if err != nil {
		return nil, nil, err
	}
	if pageInfo != nil {
		if pageInfo.Page == 0 {
			pageInfo.Page = opts.Page
		}
		if pageInfo.PageSize == 0 {
			pageInfo.PageSize = opts.PageSize
		}
	}
	return projects, pageInfo, nil
}

func (c *ClockifyClient) GetProject(ctx context.Context, id string) (*clockify.Project, error) {
	path := fmt.Sprintf("/workspaces/%s/projects/%s", c.workspaceID, id)
	var proj clockify.Project
	err := c.Do(ctx, "GET", path, nil, &proj)
	return &proj, err
}

func (c *ClockifyClient) CreateProject(ctx context.Context, req clockify.CreateProjectRequest) (*clockify.Project, error) {
	path := fmt.Sprintf("/workspaces/%s/projects", c.workspaceID)
	var proj clockify.Project
	err := c.Do(ctx, "POST", path, req, &proj)
	return &proj, err
}

func (c *ClockifyClient) UpdateProject(ctx context.Context, id string, req clockify.UpdateProjectRequest) (*clockify.Project, error) {
	path := fmt.Sprintf("/workspaces/%s/projects/%s", c.workspaceID, id)
	var proj clockify.Project
	err := c.Do(ctx, "PUT", path, req, &proj)
	return &proj, err
}

func (c *ClockifyClient) DeleteProject(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/projects/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

func (c *ClockifyClient) ListClients(ctx context.Context, opts ListClientsOpts) ([]clockify.Client, error) {
	clients, _, err := c.ListClientsPaged(ctx, opts)
	return clients, err
}

// ListClientsPaged fetches one page of clients with header-derived PageInfo.
func (c *ClockifyClient) ListClientsPaged(ctx context.Context, opts ListClientsOpts) ([]clockify.Client, *clockify.PageInfo, error) {
	path := fmt.Sprintf("/workspaces/%s/clients", c.workspaceID)
	params := map[string]string{"search": opts.Search}
	if opts.Page > 0 {
		params["page"] = strconv.Itoa(opts.Page)
	}
	if opts.PageSize > 0 {
		params["pageSize"] = strconv.Itoa(opts.PageSize)
	}
	path = appendQueryParams(path, params)

	var clients []clockify.Client
	pageInfo, err := c.DoWithPagination(ctx, "GET", path, nil, &clients)
	if err != nil {
		return nil, nil, err
	}
	if pageInfo != nil {
		if pageInfo.Page == 0 {
			pageInfo.Page = opts.Page
		}
		if pageInfo.PageSize == 0 {
			pageInfo.PageSize = opts.PageSize
		}
	}
	return clients, pageInfo, nil
}

func (c *ClockifyClient) GetClient(ctx context.Context, id string) (*clockify.Client, error) {
	path := fmt.Sprintf("/workspaces/%s/clients/%s", c.workspaceID, id)
	var client clockify.Client
	err := c.Do(ctx, "GET", path, nil, &client)
	return &client, err
}

func (c *ClockifyClient) CreateClient(ctx context.Context, req clockify.CreateClientRequest) (*clockify.Client, error) {
	path := fmt.Sprintf("/workspaces/%s/clients", c.workspaceID)
	var client clockify.Client
	err := c.Do(ctx, "POST", path, req, &client)
	return &client, err
}

func (c *ClockifyClient) UpdateClient(ctx context.Context, id string, req clockify.UpdateClientRequest) (*clockify.Client, error) {
	path := fmt.Sprintf("/workspaces/%s/clients/%s", c.workspaceID, id)
	var client clockify.Client
	err := c.Do(ctx, "PUT", path, req, &client)
	return &client, err
}

func (c *ClockifyClient) DeleteClient(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/clients/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

func (c *ClockifyClient) ListTags(ctx context.Context, opts ListTagsOpts) ([]clockify.Tag, error) {
	tags, _, err := c.ListTagsPaged(ctx, opts)
	return tags, err
}

// ListTagsPaged fetches one page of tags with header-derived PageInfo.
func (c *ClockifyClient) ListTagsPaged(ctx context.Context, opts ListTagsOpts) ([]clockify.Tag, *clockify.PageInfo, error) {
	path := fmt.Sprintf("/workspaces/%s/tags", c.workspaceID)
	params := map[string]string{"search": opts.Search}
	if opts.Page > 0 {
		params["page"] = strconv.Itoa(opts.Page)
	}
	if opts.PageSize > 0 {
		params["pageSize"] = strconv.Itoa(opts.PageSize)
	}
	path = appendQueryParams(path, params)

	var tags []clockify.Tag
	pageInfo, err := c.DoWithPagination(ctx, "GET", path, nil, &tags)
	if err != nil {
		return nil, nil, err
	}
	if pageInfo != nil {
		if pageInfo.Page == 0 {
			pageInfo.Page = opts.Page
		}
		if pageInfo.PageSize == 0 {
			pageInfo.PageSize = opts.PageSize
		}
	}
	return tags, pageInfo, nil
}

func (c *ClockifyClient) GetTag(ctx context.Context, id string) (*clockify.Tag, error) {
	path := fmt.Sprintf("/workspaces/%s/tags/%s", c.workspaceID, id)
	var tag clockify.Tag
	err := c.Do(ctx, "GET", path, nil, &tag)
	return &tag, err
}

func (c *ClockifyClient) CreateTag(ctx context.Context, req clockify.CreateTagRequest) (*clockify.Tag, error) {
	path := fmt.Sprintf("/workspaces/%s/tags", c.workspaceID)
	var tag clockify.Tag
	err := c.Do(ctx, "POST", path, req, &tag)
	return &tag, err
}

func (c *ClockifyClient) UpdateTag(ctx context.Context, id string, req clockify.UpdateTagRequest) (*clockify.Tag, error) {
	path := fmt.Sprintf("/workspaces/%s/tags/%s", c.workspaceID, id)
	var tag clockify.Tag
	err := c.Do(ctx, "PUT", path, req, &tag)
	return &tag, err
}

func (c *ClockifyClient) DeleteTag(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/tags/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

func (c *ClockifyClient) ListTasks(ctx context.Context, projectID string, opts ListTasksOpts) ([]clockify.Task, error) {
	tasks, _, err := c.ListTasksPaged(ctx, projectID, opts)
	return tasks, err
}

// ListTasksPaged fetches one page of tasks with header-derived PageInfo.
func (c *ClockifyClient) ListTasksPaged(ctx context.Context, projectID string, opts ListTasksOpts) ([]clockify.Task, *clockify.PageInfo, error) {
	path := fmt.Sprintf("/workspaces/%s/projects/%s/tasks", c.workspaceID, projectID)
	params := map[string]string{}
	if opts.Page > 0 {
		params["page"] = strconv.Itoa(opts.Page)
	}
	if opts.PageSize > 0 {
		params["pageSize"] = strconv.Itoa(opts.PageSize)
	}
	if opts.IsActive != nil {
		params["is-active"] = strconv.FormatBool(*opts.IsActive)
	}
	path = appendQueryParams(path, params)

	var tasks []clockify.Task
	pageInfo, err := c.DoWithPagination(ctx, "GET", path, nil, &tasks)
	if err != nil {
		return nil, nil, err
	}
	if pageInfo != nil {
		if pageInfo.Page == 0 {
			pageInfo.Page = opts.Page
		}
		if pageInfo.PageSize == 0 {
			pageInfo.PageSize = opts.PageSize
		}
	}
	return tasks, pageInfo, nil
}

func (c *ClockifyClient) GetTask(ctx context.Context, projectID, taskID string) (*clockify.Task, error) {
	path := fmt.Sprintf("/workspaces/%s/projects/%s/tasks/%s", c.workspaceID, projectID, taskID)
	var task clockify.Task
	err := c.Do(ctx, "GET", path, nil, &task)
	return &task, err
}

func (c *ClockifyClient) CreateTask(ctx context.Context, projectID string, req clockify.CreateTaskRequest) (*clockify.Task, error) {
	path := fmt.Sprintf("/workspaces/%s/projects/%s/tasks", c.workspaceID, projectID)
	var task clockify.Task
	err := c.Do(ctx, "POST", path, req, &task)
	return &task, err
}

func (c *ClockifyClient) UpdateTask(ctx context.Context, projectID, taskID string, req clockify.UpdateTaskRequest) (*clockify.Task, error) {
	path := fmt.Sprintf("/workspaces/%s/projects/%s/tasks/%s", c.workspaceID, projectID, taskID)
	var task clockify.Task
	err := c.Do(ctx, "PUT", path, req, &task)
	return &task, err
}

func (c *ClockifyClient) DeleteTask(ctx context.Context, projectID, taskID string) error {
	path := fmt.Sprintf("/workspaces/%s/projects/%s/tasks/%s", c.workspaceID, projectID, taskID)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

func (c *ClockifyClient) ListTimeEntries(ctx context.Context, userID string, opts ListTimeEntriesOpts) ([]clockify.TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%s/user/%s/time-entries", c.workspaceID, userID)
	path = appendQueryParams(path, map[string]string{
		"start":       opts.Start,
		"end":         opts.End,
		"project":     opts.ProjectID,
		"task":        opts.TaskID,
		"tags":        opts.Tags,
		"description": opts.Description,
		"in-progress": opts.InProgress,
		"page":        strconv.Itoa(opts.Page),
		"pageSize":    strconv.Itoa(opts.PageSize),
	})

	var entries []clockify.TimeEntry
	err := c.Do(ctx, "GET", path, nil, &entries)
	return entries, err
}

func (c *ClockifyClient) ListTimeEntriesWithPagination(ctx context.Context, userID string, opts ListTimeEntriesOpts) ([]clockify.TimeEntry, *clockify.PageInfo, error) {
	path := fmt.Sprintf("/workspaces/%s/user/%s/time-entries?%s", c.workspaceID, userID, buildTimeEntryQueryParams(opts))
	pageInfo, err := c.DoWithPagination(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	entries, err := c.ListTimeEntries(ctx, userID, opts)
	return entries, pageInfo, err
}

func (c *ClockifyClient) GetTimeEntry(ctx context.Context, id string) (*clockify.TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%s/time-entries/%s", c.workspaceID, id)
	var entry clockify.TimeEntry
	err := c.Do(ctx, "GET", path, nil, &entry)
	return &entry, err
}

func (c *ClockifyClient) GetTimeEntryInProgress(ctx context.Context, userID string) ([]clockify.TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%s/time-entries/status/in-progress", c.workspaceID)
	var entries []clockify.TimeEntry
	err := c.Do(ctx, "GET", path, nil, &entries)
	return entries, err
}

func (c *ClockifyClient) BatchGetTimeEntries(ctx context.Context, ids []string, hydrated bool) ([]clockify.TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%s/time-entries/batch", c.workspaceID)
	body := map[string]interface{}{
		"timeEntryIds": ids,
		"hydrated":     hydrated,
	}
	var entries []clockify.TimeEntry
	err := c.Do(ctx, "POST", path, body, &entries)
	return entries, err
}

func (c *ClockifyClient) CreateTimeEntry(ctx context.Context, req CreateTimeEntryRequest) (*clockify.TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%s/time-entries", c.workspaceID)
	var entry clockify.TimeEntry
	err := c.Do(ctx, "POST", path, req, &entry)
	return &entry, err
}

func (c *ClockifyClient) UpdateTimeEntry(ctx context.Context, id string, req UpdateTimeEntryRequest) (*clockify.TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%s/time-entries/%s", c.workspaceID, id)
	var entry clockify.TimeEntry
	err := c.Do(ctx, "PUT", path, req, &entry)
	return &entry, err
}

func (c *ClockifyClient) DeleteTimeEntry(ctx context.Context, id string) error {
	path := fmt.Sprintf("/workspaces/%s/time-entries/%s", c.workspaceID, id)
	return c.Do(ctx, "DELETE", path, nil, nil)
}

func (c *ClockifyClient) StopTimeEntry(ctx context.Context, userID string, req StopTimerRequest) error {
	path := fmt.Sprintf("/workspaces/%s/user/%s/time-entries", c.workspaceID, userID)
	return c.Do(ctx, "PATCH", path, req, nil)
}

func (c *ClockifyClient) BulkUpdateTimeEntries(ctx context.Context, userID string, entries []UpdateTimeEntryRequest) ([]clockify.TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%s/user/%s/time-entries", c.workspaceID, userID)
	var result []clockify.TimeEntry
	err := c.Do(ctx, "PUT", path, entries, &result)
	return result, err
}

func appendQueryParams(path string, params map[string]string) string {
	vals := url.Values{}
	for k, v := range params {
		if v == "" {
			continue
		}
		vals.Set(k, v)
	}
	enc := vals.Encode()
	if enc == "" {
		return path
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + enc
}

func buildTimeEntryQueryParams(opts ListTimeEntriesOpts) string {
	params := map[string]string{
		"start":       opts.Start,
		"end":         opts.End,
		"project":     opts.ProjectID,
		"task":        opts.TaskID,
		"tags":        opts.Tags,
		"description": opts.Description,
		"in-progress": opts.InProgress,
	}
	if opts.Page > 0 {
		params["page"] = strconv.Itoa(opts.Page)
	}
	if opts.PageSize > 0 {
		params["pageSize"] = strconv.Itoa(opts.PageSize)
	}
	return encodeQueryParams(params)
}

// encodeQueryParams encodes non-empty params with url.Values.
func encodeQueryParams(params map[string]string) string {
	return EncodeQueryParams(params)
}

// EncodeQueryParams encodes non-empty params with url.Values.
func EncodeQueryParams(params map[string]string) string {
	vals := url.Values{}
	for k, v := range params {
		if v == "" {
			continue
		}
		vals.Set(k, v)
	}
	return vals.Encode()
}
