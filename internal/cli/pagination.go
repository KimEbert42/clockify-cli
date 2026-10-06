package cli

import (
	"context"
	"strconv"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
)

func fetchPaginatedProjects(ctx context.Context, client *api.ClockifyClient, opts api.ListProjectsOpts, page, pageSize int, noPagination bool) ([]clockify.Project, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		projects, pageInfo, err := client.ListProjectsPaged(ctx, opts)
		return projects, pageInfo, err
	}

	var allProjects []clockify.Project
	currentPage := 1
	var lastPageInfo *clockify.PageInfo

	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		projects, pageInfo, err := client.ListProjectsPaged(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(projects) == 0 {
			break
		}
		allProjects = append(allProjects, projects...)
		lastPageInfo = pageInfo
		if len(projects) < pageSize {
			break
		}
		if pageInfo != nil && pageInfo.TotalPages > 0 && currentPage >= pageInfo.TotalPages {
			break
		}
		currentPage++
	}

	if lastPageInfo != nil && lastPageInfo.TotalResults > 0 {
		return allProjects, &clockify.PageInfo{
			Page:         1,
			PageSize:     pageSize,
			TotalPages:   lastPageInfo.TotalPages,
			TotalResults: lastPageInfo.TotalResults,
		}, nil
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
		clients, pageInfo, err := client.ListClientsPaged(ctx, opts)
		return clients, pageInfo, err
	}

	var allClients []clockify.Client
	currentPage := 1
	var lastPageInfo *clockify.PageInfo

	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		clients, pageInfo, err := client.ListClientsPaged(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(clients) == 0 {
			break
		}
		allClients = append(allClients, clients...)
		lastPageInfo = pageInfo
		if len(clients) < pageSize {
			break
		}
		if pageInfo != nil && pageInfo.TotalPages > 0 && currentPage >= pageInfo.TotalPages {
			break
		}
		currentPage++
	}

	if lastPageInfo != nil && lastPageInfo.TotalResults > 0 {
		return allClients, &clockify.PageInfo{
			Page:         1,
			PageSize:     pageSize,
			TotalPages:   lastPageInfo.TotalPages,
			TotalResults: lastPageInfo.TotalResults,
		}, nil
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
		tags, pageInfo, err := client.ListTagsPaged(ctx, opts)
		return tags, pageInfo, err
	}

	var allTags []clockify.Tag
	currentPage := 1
	var lastPageInfo *clockify.PageInfo

	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		tags, pageInfo, err := client.ListTagsPaged(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(tags) == 0 {
			break
		}
		allTags = append(allTags, tags...)
		lastPageInfo = pageInfo
		if len(tags) < pageSize {
			break
		}
		if pageInfo != nil && pageInfo.TotalPages > 0 && currentPage >= pageInfo.TotalPages {
			break
		}
		currentPage++
	}

	if lastPageInfo != nil && lastPageInfo.TotalResults > 0 {
		return allTags, &clockify.PageInfo{
			Page:         1,
			PageSize:     pageSize,
			TotalPages:   lastPageInfo.TotalPages,
			TotalResults: lastPageInfo.TotalResults,
		}, nil
	}
	return allTags, &clockify.PageInfo{
		Page:         1,
		PageSize:     pageSize,
		TotalPages:   currentPage,
		TotalResults: len(allTags),
	}, nil
}

func fetchPaginatedTasks(ctx context.Context, client *api.ClockifyClient, projectID string, opts api.ListTasksOpts, page, pageSize int, noPagination bool) ([]clockify.Task, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		tasks, pageInfo, err := client.ListTasksPaged(ctx, projectID, opts)
		return tasks, pageInfo, err
	}

	var allTasks []clockify.Task
	currentPage := 1
	var lastPageInfo *clockify.PageInfo

	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		tasks, pageInfo, err := client.ListTasksPaged(ctx, projectID, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(tasks) == 0 {
			break
		}
		allTasks = append(allTasks, tasks...)
		lastPageInfo = pageInfo
		if len(tasks) < pageSize {
			break
		}
		if pageInfo != nil && pageInfo.TotalPages > 0 && currentPage >= pageInfo.TotalPages {
			break
		}
		currentPage++
	}

	if lastPageInfo != nil && lastPageInfo.TotalResults > 0 {
		return allTasks, &clockify.PageInfo{
			Page:         1,
			PageSize:     pageSize,
			TotalPages:   lastPageInfo.TotalPages,
			TotalResults: lastPageInfo.TotalResults,
		}, nil
	}
	return allTasks, &clockify.PageInfo{
		Page:         1,
		PageSize:     pageSize,
		TotalPages:   currentPage,
		TotalResults: len(allTasks),
	}, nil
}

func buildProjectQueryParams(opts api.ListProjectsOpts) string {
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
	return api.EncodeQueryParams(params)
}
