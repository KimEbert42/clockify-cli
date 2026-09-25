package api

import (
	"context"
	"fmt"
	"strconv"

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
	path = appendQueryParams(path, map[string]string{
		"status":   opts.Status,
		"page":     strconv.Itoa(opts.Page),
		"pageSize": strconv.Itoa(opts.PageSize),
	})

	var users []clockify.User
	err := c.Do(ctx, "GET", path, nil, &users)
	return users, err
}

func (c *ClockifyClient) ListProjects(ctx context.Context, opts ListProjectsOpts) ([]clockify.Project, error) {
	path := fmt.Sprintf("/workspaces/%s/projects", c.workspaceID)
	path = appendQueryParams(path, map[string]string{
		"status":   opts.Status,
		"clientId": opts.ClientID,
		"search":   opts.Search,
		"page":     strconv.Itoa(opts.Page),
		"pageSize": strconv.Itoa(opts.PageSize),
	})

	var projects []clockify.Project
	err := c.Do(ctx, "GET", path, nil, &projects)
	return projects, err
}

func (c *ClockifyClient) GetProject(ctx context.Context, id string) (*clockify.Project, error) {
	path := fmt.Sprintf("/workspaces/%s/projects/%s", c.workspaceID, id)
	var proj clockify.Project
	err := c.Do(ctx, "GET", path, nil, &proj)
	return &proj, err
}

func (c *ClockifyClient) ListClients(ctx context.Context, opts ListClientsOpts) ([]clockify.Client, error) {
	path := fmt.Sprintf("/workspaces/%s/clients", c.workspaceID)
	path = appendQueryParams(path, map[string]string{
		"search":   opts.Search,
		"page":     strconv.Itoa(opts.Page),
		"pageSize": strconv.Itoa(opts.PageSize),
	})

	var clients []clockify.Client
	err := c.Do(ctx, "GET", path, nil, &clients)
	return clients, err
}

func (c *ClockifyClient) GetClient(ctx context.Context, id string) (*clockify.Client, error) {
	path := fmt.Sprintf("/workspaces/%s/clients/%s", c.workspaceID, id)
	var client clockify.Client
	err := c.Do(ctx, "GET", path, nil, &client)
	return &client, err
}

func (c *ClockifyClient) ListTags(ctx context.Context, opts ListTagsOpts) ([]clockify.Tag, error) {
	path := fmt.Sprintf("/workspaces/%s/tags", c.workspaceID)
	path = appendQueryParams(path, map[string]string{
		"search":   opts.Search,
		"page":     strconv.Itoa(opts.Page),
		"pageSize": strconv.Itoa(opts.PageSize),
	})

	var tags []clockify.Tag
	err := c.Do(ctx, "GET", path, nil, &tags)
	return tags, err
}

func appendQueryParams(path string, params map[string]string) string {
	first := true
	for k, v := range params {
		if v == "" {
			continue
		}
		if first {
			path += "?"
			first = false
		} else {
			path += "&"
		}
		path += k + "=" + v
	}
	return path
}
