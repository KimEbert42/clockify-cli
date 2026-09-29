package cli

import (
	"context"
	"fmt"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
)

func fetchPaginatedProjects(ctx context.Context, client *api.ClockifyClient, opts api.ListProjectsOpts, page, pageSize int, noPagination bool) ([]clockify.Project, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
	 pageInfo, err := client.DoWithPagination(ctx, "GET",
			fmt.Sprintf("/workspaces/%s/projects?%s", client.WorkspaceID(), buildProjectQueryParams(opts)),
			nil, nil)
		if err != nil {
			return nil, nil, err
		}
		projects, err := client.ListProjects(ctx, opts)
		return projects, pageInfo, err
	}

	var allProjects []clockify.Project
	currentPage := 1
	var lastPageInfo *clockify.PageInfo

	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		projects, err := client.ListProjects(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(projects) == 0 {
			break
		}
		allProjects = append(allProjects, projects...)
		if len(projects) < pageSize {
			break
		}
		currentPage++
		_ = lastPageInfo
	}

	return allProjects, &clockify.PageInfo{
		Page:         1,
		PageSize:     pageSize,
		TotalPages:   currentPage,
		TotalResults: len(allProjects),
	}, nil
}

func fetchPaginatedClients(ctx context.Context, client *api.ClockifyClient, opts api.ListClientsOpts, page, pageSize int, noPagination bool) ([]clockify.Client, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		clients, err := client.ListClients(ctx, opts)
		return clients, &clockify.PageInfo{Page: page, PageSize: pageSize}, err
	}

	var allClients []clockify.Client
	currentPage := 1

	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		clients, err := client.ListClients(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(clients) == 0 {
			break
		}
		allClients = append(allClients, clients...)
		if len(clients) < pageSize {
			break
		}
		currentPage++
	}

	return allClients, &clockify.PageInfo{
		Page:         1,
		PageSize:     pageSize,
		TotalPages:   currentPage,
		TotalResults: len(allClients),
	}, nil
}

func fetchPaginatedTags(ctx context.Context, client *api.ClockifyClient, opts api.ListTagsOpts, page, pageSize int, noPagination bool) ([]clockify.Tag, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		tags, err := client.ListTags(ctx, opts)
		return tags, &clockify.PageInfo{Page: page, PageSize: pageSize}, err
	}

	var allTags []clockify.Tag
	currentPage := 1

	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		tags, err := client.ListTags(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(tags) == 0 {
			break
		}
		allTags = append(allTags, tags...)
		if len(tags) < pageSize {
			break
		}
		currentPage++
	}

	return allTags, &clockify.PageInfo{
		Page:         1,
		PageSize:     pageSize,
		TotalPages:   currentPage,
		TotalResults: len(allTags),
	}, nil
}

func buildProjectQueryParams(opts api.ListProjectsOpts) string {
	params := ""
	if opts.Status != "" {
		params += "status=" + opts.Status + "&"
	}
	if opts.ClientID != "" {
		params += "clientId=" + opts.ClientID + "&"
	}
	if opts.Search != "" {
		params += "search=" + opts.Search + "&"
	}
	if opts.Page > 0 {
		params += fmt.Sprintf("page=%d&", opts.Page)
	}
	if opts.PageSize > 0 {
		params += fmt.Sprintf("pageSize=%d&", opts.PageSize)
	}
	if len(params) > 0 && params[len(params)-1] == '&' {
		params = params[:len(params)-1]
	}
	return params
}
