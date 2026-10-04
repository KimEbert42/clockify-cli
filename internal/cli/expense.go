package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newExpenseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expense",
		Short: "Manage expenses",
		Long:  `List, get, create, update, and delete Clockify expenses.`,
	}
	cmd.AddCommand(newExpenseListCmd())
	cmd.AddCommand(newExpenseGetCmd())
	cmd.AddCommand(newExpenseCreateCmd())
	cmd.AddCommand(newExpenseUpdateCmd())
	cmd.AddCommand(newExpenseDeleteCmd())
	return cmd
}

func newExpenseListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List expenses",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			projectRef, _ := cmd.Flags().GetString("project")
			userRef, _ := cmd.Flags().GetString("user")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			page, _ := cmd.Flags().GetInt("page")
			pageSize, _ := cmd.Flags().GetInt("page-size")
			noPagination, _ := cmd.Flags().GetBool("no-pagination")

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			opts := api.ListExpensesOpts{Page: page, PageSize: pageSize}
			if projectRef != "" {
				proj, err := resolveProjectRef(ctx, client, projectRef)
				if err != nil {
					return err
				}
				opts.ProjectID = proj.ID
			}
			if userRef != "" {
				uid, err := resolveUserRef(ctx, client, userRef)
				if err != nil {
					return err
				}
				opts.UserID = uid
			}
			if start != "" {
				s, err := parseDateTimeCLI(start)
				if err != nil {
					return err
				}
				opts.Start = s
			}
			if end != "" {
				e, err := parseDateTimeCLI(end)
				if err != nil {
					return err
				}
				opts.End = e
			}

			expenses, err := fetchPaginatedExpenses(ctx, client, opts, page, pageSize, noPagination)
			if err != nil {
				return err
			}
			if len(expenses) == 0 && !quiet {
				cmd.Println("No expenses found.")
				return nil
			}
			type expDisplay struct {
				ID         string  `json:"id" yaml:"id"`
				Date       string  `json:"date" yaml:"date"`
				Total      float64 `json:"total" yaml:"total"`
				CategoryID string  `json:"categoryId" yaml:"categoryId"`
				ProjectID  string  `json:"projectId" yaml:"projectId"`
				UserID     string  `json:"userId" yaml:"userId"`
				Billable   bool    `json:"billable" yaml:"billable"`
				Notes      string  `json:"notes" yaml:"notes"`
			}
			displays := make([]expDisplay, 0, len(expenses))
			for _, e := range expenses {
				displays = append(displays, expDisplay{
					ID: e.ID, Date: e.Date, Total: e.Total, CategoryID: e.CategoryID,
					ProjectID: e.ProjectID, UserID: e.UserID, Billable: e.Billable, Notes: e.Notes,
				})
			}
			return formatOutput(cmd, displays, outputFormat, []string{"ID", "DATE", "TOTAL", "CATEGORY_ID", "PROJECT_ID", "BILLABLE", "NOTES"})
		},
	}
	cmd.Flags().String("project", "", "filter by project ID or name")
	cmd.Flags().String("user", "", "filter by user ID, email, or name")
	cmd.Flags().String("start", "", "start date (YYYY-MM-DD or relative)")
	cmd.Flags().String("end", "", "end date (YYYY-MM-DD or relative)")
	return cmd
}

func newExpenseGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an expense by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			e, err := client.GetExpense(context.Background(), id)
			if err != nil {
				return err
			}
			return formatOutput(cmd, e, outputFormat, []string{"ID", "DATE", "TOTAL", "CATEGORY_ID", "PROJECT_ID", "USER_ID", "BILLABLE", "NOTES"})
		},
	}
	cmd.Flags().String("id", "", "expense ID")
	return cmd
}

func newExpenseCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an expense",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			projectRef, _ := cmd.Flags().GetString("project")
			dateFlag, _ := cmd.Flags().GetString("date")
			amountFlag, _ := cmd.Flags().GetString("amount")
			typeFlag, _ := cmd.Flags().GetString("type")
			categoryFlag, _ := cmd.Flags().GetString("category")
			description, _ := cmd.Flags().GetString("description")
			userRef, _ := cmd.Flags().GetString("user")
			billable, _ := cmd.Flags().GetBool("billable")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if filePath != "" || stdin {
				var raw map[string]any
				if err := readJSONInput(filePath, stdin, &raw); err != nil {
					return err
				}
				if v, ok := raw["projectId"].(string); ok && projectRef == "" {
					projectRef = v
				}
				if v, ok := raw["date"].(string); ok && dateFlag == "" {
					dateFlag = v
				}
				if v, ok := raw["total"].(float64); ok && amountFlag == "" {
					amountFlag = fmt.Sprint(v)
				}
				if v, ok := raw["amount"].(float64); ok && amountFlag == "" {
					amountFlag = fmt.Sprint(v)
				}
				if v, ok := raw["categoryId"].(string); ok && categoryFlag == "" {
					categoryFlag = v
				}
				if v, ok := raw["notes"].(string); ok && description == "" {
					description = v
				}
				if v, ok := raw["description"].(string); ok && description == "" {
					description = v
				}
				if v, ok := raw["billable"].(bool); ok {
					billable = v
				}
			}

			if projectRef == "" {
				return fmt.Errorf("must specify --project")
			}
			if dateFlag == "" {
				return fmt.Errorf("must specify --date")
			}
			if amountFlag == "" {
				return fmt.Errorf("must specify --amount")
			}
			if typeFlag == "" && categoryFlag == "" {
				return fmt.Errorf("must specify --type or --category")
			}

			date, err := parseDateOnlyCLI(dateFlag)
			if err != nil {
				return err
			}
			amount, err := parseAmount(amountFlag)
			if err != nil {
				return err
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			proj, err := resolveProjectRef(ctx, client, projectRef)
			if err != nil {
				return err
			}

			catRef := categoryFlag
			if catRef == "" {
				catRef = typeFlag
			}
			catID, err := resolveExpenseCategory(ctx, client, catRef)
			if err != nil {
				return err
			}

			userID := ""
			if userRef != "" {
				userID, err = resolveUserRef(ctx, client, userRef)
				if err != nil {
					return err
				}
			} else {
				u, err := client.GetCurrentUser(ctx)
				if err != nil {
					return err
				}
				userID = u.ID
			}

			e, err := client.CreateExpense(ctx, clockify.CreateExpenseRequest{
				ProjectID:  proj.ID,
				UserID:     userID,
				CategoryID: catID,
				Date:       date,
				Total:      amount,
				Billable:   billable,
				Notes:      description,
			})
			if err != nil {
				return err
			}
			return formatOutput(cmd, e, outputFormat, []string{"ID", "DATE", "TOTAL", "CATEGORY_ID", "PROJECT_ID", "BILLABLE", "NOTES"})
		},
	}
	cmd.Flags().String("project", "", "project ID or name (required)")
	cmd.Flags().String("date", "", "expense date YYYY-MM-DD or relative (required)")
	cmd.Flags().String("amount", "", "expense amount (required)")
	cmd.Flags().String("type", "", "expense type/category name (e.g. TRAVEL, MEAL)")
	cmd.Flags().String("category", "", "expense category ID or name (overrides --type)")
	cmd.Flags().String("description", "", "expense notes")
	cmd.Flags().String("user", "", "user ID, email, or name (defaults to current user)")
	cmd.Flags().Bool("billable", false, "expense is billable")
	cmd.Flags().String("file", "", "JSON file with expense data")
	cmd.Flags().Bool("stdin", false, "read expense data from stdin")
	return cmd
}

func newExpenseUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an expense",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			amountFlag, _ := cmd.Flags().GetString("amount")
			categoryFlag, _ := cmd.Flags().GetString("category")
			typeFlag, _ := cmd.Flags().GetString("type")
			description, _ := cmd.Flags().GetString("description")
			dateFlag, _ := cmd.Flags().GetString("date")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			existing, err := client.GetExpense(ctx, id)
			if err != nil {
				return err
			}

			req := clockify.UpdateExpenseRequest{
				ProjectID:  existing.ProjectID,
				TaskID:     existing.TaskID,
				UserID:     existing.UserID,
				CategoryID: existing.CategoryID,
				Date:       existing.Date,
				Total:      existing.Total,
				Quantity:   existing.Quantity,
				Billable:   existing.Billable,
				Notes:      existing.Notes,
			}

			if filePath != "" || stdin {
				var raw map[string]any
				if err := readJSONInput(filePath, stdin, &raw); err != nil {
					return err
				}
				if v, ok := raw["total"].(float64); ok {
					req.Total = v
				}
				if v, ok := raw["amount"].(float64); ok {
					req.Total = v
				}
				if v, ok := raw["categoryId"].(string); ok && v != "" {
					req.CategoryID = v
				}
				if v, ok := raw["notes"].(string); ok {
					req.Notes = v
				}
				if v, ok := raw["date"].(string); ok && v != "" {
					d, err := parseDateOnlyCLI(v)
					if err != nil {
						return err
					}
					req.Date = d
				}
			} else {
				if amountFlag != "" {
					amount, err := parseAmount(amountFlag)
					if err != nil {
						return err
					}
					req.Total = amount
				}
				catRef := categoryFlag
				if catRef == "" {
					catRef = typeFlag
				}
				if catRef != "" {
					catID, err := resolveExpenseCategory(ctx, client, catRef)
					if err != nil {
						return err
					}
					req.CategoryID = catID
				}
				if description != "" {
					req.Notes = description
				}
				if dateFlag != "" {
					d, err := parseDateOnlyCLI(dateFlag)
					if err != nil {
						return err
					}
					req.Date = d
				}
			}

			e, err := client.UpdateExpense(ctx, id, req)
			if err != nil {
				return err
			}
			return formatOutput(cmd, e, outputFormat, []string{"ID", "DATE", "TOTAL", "CATEGORY_ID", "PROJECT_ID", "BILLABLE", "NOTES"})
		},
	}
	cmd.Flags().String("id", "", "expense ID")
	cmd.Flags().String("amount", "", "new amount")
	cmd.Flags().String("category", "", "new category ID or name")
	cmd.Flags().String("type", "", "new expense type/category name")
	cmd.Flags().String("description", "", "new notes")
	cmd.Flags().String("date", "", "new date YYYY-MM-DD or relative")
	cmd.Flags().String("file", "", "JSON file with update data")
	cmd.Flags().Bool("stdin", false, "read update data from stdin")
	return cmd
}

func newExpenseDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an expense",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			confirm, _ := cmd.Flags().GetBool("confirm")
			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()
			name := ""
			if e, err := client.GetExpense(ctx, id); err == nil {
				name = e.Notes
				if len(name) > 40 {
					name = strings.TrimSpace(name[:40]) + "..."
				}
			}
			if !confirmDestructive(cmd, confirm, "expense", id, name) {
				return nil
			}
			if err := client.DeleteExpense(ctx, id); err != nil {
				return err
			}
			cmd.Printf("Expense %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "expense ID")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}
