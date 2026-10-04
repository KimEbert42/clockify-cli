package cli

import (
	"context"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
)

func fetchPaginatedGroups(ctx context.Context, client *api.ClockifyClient, opts api.ListUserGroupsOpts, page, pageSize int, noPagination bool) ([]clockify.UserGroup, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		groups, err := client.ListUserGroups(ctx, opts)
		return groups, &clockify.PageInfo{Page: page, PageSize: pageSize}, err
	}

	var all []clockify.UserGroup
	currentPage := 1
	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		groups, err := client.ListUserGroups(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(groups) == 0 {
			break
		}
		all = append(all, groups...)
		if len(groups) < pageSize {
			break
		}
		currentPage++
	}
	return all, &clockify.PageInfo{Page: 1, PageSize: pageSize, TotalPages: currentPage, TotalResults: len(all)}, nil
}

func fetchPaginatedWebhooks(ctx context.Context, client *api.ClockifyClient, opts api.ListWebhooksOpts, page, pageSize int, noPagination bool) ([]clockify.Webhook, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		hooks, err := client.ListWebhooks(ctx, opts)
		return hooks, &clockify.PageInfo{Page: page, PageSize: pageSize}, err
	}

	var all []clockify.Webhook
	currentPage := 1
	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		hooks, err := client.ListWebhooks(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(hooks) == 0 {
			break
		}
		all = append(all, hooks...)
		if len(hooks) < pageSize {
			break
		}
		currentPage++
	}
	return all, &clockify.PageInfo{Page: 1, PageSize: pageSize, TotalPages: currentPage, TotalResults: len(all)}, nil
}

func fetchPaginatedCustomFields(ctx context.Context, client *api.ClockifyClient, opts api.ListCustomFieldsOpts, page, pageSize int, noPagination bool) ([]clockify.CustomField, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		fields, err := client.ListCustomFields(ctx, opts)
		return fields, &clockify.PageInfo{Page: page, PageSize: pageSize}, err
	}

	var all []clockify.CustomField
	currentPage := 1
	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		fields, err := client.ListCustomFields(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(fields) == 0 {
			break
		}
		all = append(all, fields...)
		if len(fields) < pageSize {
			break
		}
		currentPage++
	}
	return all, &clockify.PageInfo{Page: 1, PageSize: pageSize, TotalPages: currentPage, TotalResults: len(all)}, nil
}

func fetchPaginatedExpenses(ctx context.Context, client *api.ClockifyClient, opts api.ListExpensesOpts, page, pageSize int, noPagination bool) ([]clockify.Expense, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		return client.ListExpenses(ctx, opts)
	}

	var all []clockify.Expense
	currentPage := 1
	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		expenses, err := client.ListExpenses(ctx, opts)
		if err != nil {
			return nil, err
		}
		if len(expenses) == 0 {
			break
		}
		all = append(all, expenses...)
		if len(expenses) < pageSize {
			break
		}
		currentPage++
	}
	return all, nil
}

func fetchPaginatedInvoices(ctx context.Context, client *api.ClockifyClient, opts api.ListInvoicesOpts, page, pageSize int, noPagination bool) ([]clockify.Invoice, *clockify.PageInfo, error) {
	if !noPagination {
		opts.Page = page
		opts.PageSize = pageSize
		return client.ListInvoices(ctx, opts)
	}

	var all []clockify.Invoice
	currentPage := 1
	for {
		opts.Page = currentPage
		opts.PageSize = pageSize
		invoices, pageInfo, err := client.ListInvoices(ctx, opts)
		if err != nil {
			return nil, nil, err
		}
		if len(invoices) == 0 {
			break
		}
		all = append(all, invoices...)
		if len(invoices) < pageSize {
			break
		}
		currentPage++
		_ = pageInfo
	}
	return all, &clockify.PageInfo{Page: 1, PageSize: pageSize, TotalPages: currentPage, TotalResults: len(all)}, nil
}
