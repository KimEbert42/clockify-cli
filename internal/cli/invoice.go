package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newInvoiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoice",
		Short: "Manage invoices",
		Long: `List, get, create, update, and delete Clockify invoices.

Status mapping (CLI → API):
  DRAFT, FINALIZED → UNSENT; SENT → SENT; ACCEPTED, PAID → PAID; VOID → VOID
  Official statuses also accepted: UNSENT, PARTIALLY_PAID, OVERDUE`,
	}
	cmd.AddCommand(newInvoiceListCmd())
	cmd.AddCommand(newInvoiceGetCmd())
	cmd.AddCommand(newInvoiceCreateCmd())
	cmd.AddCommand(newInvoiceUpdateCmd())
	cmd.AddCommand(newInvoiceDeleteCmd())
	return cmd
}

func newInvoiceListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoices",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			clientRef, _ := cmd.Flags().GetString("client")
			status, _ := cmd.Flags().GetString("status")
			page, _ := cmd.Flags().GetInt("page")
			pageSize, _ := cmd.Flags().GetInt("page-size")
			noPagination, _ := cmd.Flags().GetBool("no-pagination")

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			opts := api.ListInvoicesOpts{Page: page, PageSize: pageSize}
			if clientRef != "" {
				cid, err := resolveClientRef(ctx, client, clientRef)
				if err != nil {
					return err
				}
				opts.ClientID = cid
			}
			if status != "" {
				mapped, err := clockify.MapInvoiceStatus(status)
				if err != nil {
					return err
				}
				opts.Status = mapped
			}

			invoices, pageInfo, err := fetchPaginatedInvoices(ctx, client, opts, page, pageSize, noPagination)
			if err != nil {
				return err
			}
			if len(invoices) == 0 && !quiet {
				cmd.Println("No invoices found.")
				return nil
			}
			type invDisplay struct {
				ID      string  `json:"id" yaml:"id"`
				Number  string  `json:"number" yaml:"number"`
				Client  string  `json:"clientId" yaml:"clientId"`
				Status  string  `json:"status" yaml:"status"`
				Currency string `json:"currency" yaml:"currency"`
				Issued  string  `json:"issuedDate" yaml:"issuedDate"`
				Due     string  `json:"dueDate" yaml:"dueDate"`
				Total   float64 `json:"total" yaml:"total"`
			}
			displays := make([]invDisplay, 0, len(invoices))
			for _, inv := range invoices {
				displays = append(displays, invDisplay{
					ID: inv.ID, Number: inv.Number, Client: inv.ClientID,
					Status: invoiceStatusDisplay(inv), Currency: inv.Currency,
					Issued: inv.IssuedDate, Due: inv.DueDate, Total: inv.Total,
				})
			}
			if err := formatOutput(cmd, displays, outputFormat, []string{"ID", "NUMBER", "CLIENT_ID", "STATUS", "CURRENCY", "ISSUED", "DUE", "TOTAL"}); err != nil {
				return err
			}
			printPaginationSummary(cmd, quiet, outputFormat, pageInfo)
			return nil
		},
	}
	cmd.Flags().String("client", "", "filter by client ID or name")
	cmd.Flags().String("status", "", "filter by invoice status")
	return cmd
}

func newInvoiceGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an invoice by ID",
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
			inv, err := client.GetInvoice(context.Background(), id)
			if err != nil {
				return err
			}
			return formatOutput(cmd, inv, outputFormat, []string{"ID", "NUMBER", "CLIENT_ID", "STATUS", "CURRENCY", "ISSUED", "DUE", "TOTAL", "NOTE"})
		},
	}
	cmd.Flags().String("id", "", "invoice ID")
	return cmd
}

func newInvoiceCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an invoice",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			clientRef, _ := cmd.Flags().GetString("client")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			currency, _ := cmd.Flags().GetString("currency")
			number, _ := cmd.Flags().GetString("number")
			status, _ := cmd.Flags().GetString("status")
			note, _ := cmd.Flags().GetString("note")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if filePath != "" || stdin {
				var raw map[string]any
				if err := readJSONInput(filePath, stdin, &raw); err != nil {
					return err
				}
				if v, ok := raw["clientId"].(string); ok && clientRef == "" {
					clientRef = v
				}
				if v, ok := raw["issuedDate"].(string); ok && start == "" {
					start = v
				}
				if v, ok := raw["start"].(string); ok && start == "" {
					start = v
				}
				if v, ok := raw["dueDate"].(string); ok && end == "" {
					end = v
				}
				if v, ok := raw["end"].(string); ok && end == "" {
					end = v
				}
				if v, ok := raw["currency"].(string); ok && currency == "" {
					currency = v
				}
				if v, ok := raw["number"].(string); ok && number == "" {
					number = v
				}
				if v, ok := raw["note"].(string); ok && note == "" {
					note = v
				}
			}

			if clientRef == "" {
				return fmt.Errorf("must specify --client")
			}
			if start == "" {
				return fmt.Errorf("must specify --start (maps to issuedDate)")
			}
			if end == "" {
				return fmt.Errorf("must specify --end (maps to dueDate)")
			}
			if currency == "" {
				return fmt.Errorf("must specify --currency (e.g. USD, EUR)")
			}

			issued, err := parseDateOnlyCLI(start)
			if err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}
			due, err := parseDateOnlyEndCLI(end)
			if err != nil {
				return fmt.Errorf("invalid --end: %w", err)
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()
			cid, err := resolveClientRef(ctx, client, clientRef)
			if err != nil {
				return err
			}

			if number == "" {
				number = fmt.Sprintf("INV-%s", time.Now().Format("20060102-150405"))
			}

			req := clockify.CreateInvoiceRequest{
				ClientID:   cid,
				Currency:   currency,
				IssuedDate: issued,
				DueDate:    due,
				Number:     number,
				Note:       note,
			}
			if status != "" {
				mapped, err := clockify.MapInvoiceStatus(status)
				if err != nil {
					return err
				}
				req.InvoiceStatus = mapped
			}

			inv, err := client.CreateInvoice(ctx, req)
			if err != nil {
				return err
			}
			return formatOutput(cmd, inv, outputFormat, []string{"ID", "NUMBER", "CLIENT_ID", "STATUS", "CURRENCY", "ISSUED", "DUE", "TOTAL", "NOTE"})
		},
	}
	cmd.Flags().String("client", "", "client ID or name (required)")
	cmd.Flags().String("start", "", "issued date YYYY-MM-DD or relative (required)")
	cmd.Flags().String("end", "", "due date YYYY-MM-DD or relative (required)")
	cmd.Flags().String("currency", "", "currency code e.g. USD (required)")
	cmd.Flags().String("number", "", "invoice number (auto-generated if omitted)")
	cmd.Flags().String("status", "", "invoice status")
	cmd.Flags().String("note", "", "invoice note")
	cmd.Flags().String("file", "", "JSON file with invoice data")
	cmd.Flags().Bool("stdin", false, "read invoice data from stdin")
	return cmd
}

func newInvoiceUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an invoice or change its status",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			clientRef, _ := cmd.Flags().GetString("client")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			currency, _ := cmd.Flags().GetString("currency")
			number, _ := cmd.Flags().GetString("number")
			status, _ := cmd.Flags().GetString("status")
			note, _ := cmd.Flags().GetString("note")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			if status != "" && clientRef == "" && start == "" && end == "" && currency == "" && number == "" && note == "" {
				mapped, err := clockify.MapInvoiceStatus(status)
				if err != nil {
					return err
				}
				inv, err := client.ChangeInvoiceStatus(ctx, id, mapped)
				if err != nil {
					return err
				}
				return formatOutput(cmd, inv, outputFormat, []string{"ID", "NUMBER", "CLIENT_ID", "STATUS", "CURRENCY", "ISSUED", "DUE", "TOTAL", "NOTE"})
			}

			existing, err := client.GetInvoice(ctx, id)
			if err != nil {
				return err
			}
			req := clockify.UpdateInvoiceRequest{
				ClientID:      existing.ClientID,
				Currency:      existing.Currency,
				IssuedDate:    existing.IssuedDate,
				DueDate:       existing.DueDate,
				Number:        existing.Number,
				InvoiceStatus: invoiceStatusDisplay(*existing),
				Note:          existing.Note,
			}
			if clientRef != "" {
				cid, err := resolveClientRef(ctx, client, clientRef)
				if err != nil {
					return err
				}
				req.ClientID = cid
			}
			if start != "" {
				issued, err := parseDateTimeCLI(start)
				if err != nil {
					return err
				}
				req.IssuedDate = issued
			}
			if end != "" {
				due, err := parseDateOnlyEndCLI(end)
				if err != nil {
					return err
				}
				req.DueDate = due
			}
			if currency != "" {
				req.Currency = currency
			}
			if number != "" {
				req.Number = number
			}
			if note != "" {
				req.Note = note
			}
			if status != "" {
				mapped, err := clockify.MapInvoiceStatus(status)
				if err != nil {
					return err
				}
				req.InvoiceStatus = mapped
			}

			inv, err := client.UpdateInvoice(ctx, id, req)
			if err != nil {
				return err
			}
			return formatOutput(cmd, inv, outputFormat, []string{"ID", "NUMBER", "CLIENT_ID", "STATUS", "CURRENCY", "ISSUED", "DUE", "TOTAL", "NOTE"})
		},
	}
	cmd.Flags().String("id", "", "invoice ID")
	cmd.Flags().String("client", "", "new client ID or name")
	cmd.Flags().String("start", "", "new issued date")
	cmd.Flags().String("end", "", "new due date")
	cmd.Flags().String("currency", "", "new currency code")
	cmd.Flags().String("number", "", "new invoice number")
	cmd.Flags().String("status", "", "new invoice status")
	cmd.Flags().String("note", "", "new note")
	return cmd
}

func newInvoiceDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an invoice",
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
			if inv, err := client.GetInvoice(ctx, id); err == nil {
				name = inv.Number
			}
			if !confirmDestructive(cmd, confirm, "invoice", id, name) {
				return nil
			}
			if err := client.DeleteInvoice(ctx, id); err != nil {
				return err
			}
			cmd.Printf("Invoice %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "invoice ID")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}
