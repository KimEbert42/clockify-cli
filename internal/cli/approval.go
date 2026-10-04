package cli

import (
	"context"
	"fmt"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newApprovalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "approval",
		Aliases: []string{"approvals"},
		Short:   "Manage approval requests",
		Long: `List, submit, and update Clockify approval requests.

Note: the official Clockify API uses period-based approval submission (WEEKLY/MONTHLY),
not per-time-entry IDs as shown in older spec examples.`,
	}
	cmd.AddCommand(newApprovalListCmd())
	cmd.AddCommand(newApprovalSubmitCmd())
	cmd.AddCommand(newApprovalUpdateCmd())
	return cmd
}

func newApprovalListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List approval requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			status, _ := cmd.Flags().GetString("status")
			types, _ := cmd.Flags().GetString("type")
			userRef, _ := cmd.Flags().GetString("user")

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			opts := api.ListApprovalRequestsOpts{}
			if status != "" {
				mapped, err := clockify.MapApprovalListStatus(status)
				if err != nil {
					return err
				}
				opts.Status = mapped
			}
			if types != "" {
				mapped, err := clockify.MapApprovalType(types)
				if err != nil {
					return err
				}
				opts.Types = mapped
			}

			requests, err := client.ListApprovalRequests(ctx, opts)
			if err != nil {
				return err
			}

			if userRef != "" {
				uid, err := resolveUserRef(ctx, client, userRef)
				if err != nil {
					return err
				}
				var filtered []clockify.ApprovalRequest
				for _, r := range requests {
					if r.UserID == uid {
						filtered = append(filtered, r)
					}
				}
				requests = filtered
			}

			// Client-side filter for statuses not accepted by the API list enum.
			if status != "" {
				want, _ := clockify.MapApprovalListStatus(status)
				var filtered []clockify.ApprovalRequest
				for _, r := range requests {
					if r.Status == want {
						filtered = append(filtered, r)
					}
				}
				// Only apply if API returned unfiltered results we still want to narrow.
				if want == "REJECTED" || want == "NEW" || want == "IN_PROGRESS" {
					requests = filtered
				}
			}

			if len(requests) == 0 && !quiet {
				cmd.Println("No approval requests found.")
				return nil
			}
			type apDisplay struct {
				ID     string `json:"id" yaml:"id"`
				User   string `json:"userId" yaml:"userId"`
				Status string `json:"status" yaml:"status"`
				Type   string `json:"type" yaml:"type"`
				Start  string `json:"start" yaml:"start"`
				End    string `json:"end" yaml:"end"`
				Note   string `json:"note" yaml:"note"`
			}
			displays := make([]apDisplay, 0, len(requests))
			for _, r := range requests {
				displays = append(displays, apDisplay{
					ID: r.ID, User: r.UserID, Status: r.Status, Type: r.Type,
					Start: r.Start, End: r.End, Note: r.Note,
				})
			}
			return formatOutput(cmd, displays, outputFormat, []string{"ID", "USER_ID", "STATUS", "TYPE", "START", "END", "NOTE"})
		},
	}
	cmd.Flags().String("status", "", "filter by status: PENDING, APPROVED, REJECTED, WITHDRAWN_APPROVAL")
	cmd.Flags().String("type", "", "filter by type: TIMESHEET, EXPENSE, TIMESHEET_AND_EXPENSE")
	cmd.Flags().String("user", "", "filter by user ID, email, or name (client-side)")
	return cmd
}

func newApprovalSubmitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit an approval request for a period",
		Long: `Submit a period-based approval request.

Official API model: POST /approval-requests/{type} with period + periodStart.
Per-time-entry submission (--time-entry) is not supported by the official API.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			types, _ := cmd.Flags().GetString("type")
			period, _ := cmd.Flags().GetString("period")
			start, _ := cmd.Flags().GetString("start")
			userRef, _ := cmd.Flags().GetString("user")
			note, _ := cmd.Flags().GetString("note")
			timeEntry, _ := cmd.Flags().GetString("time-entry")

			if timeEntry != "" {
				return fmt.Errorf("--time-entry is not supported by the official Clockify approval API; use --type --period --start for period-based submission (e.g. --type TIMESHEET --period WEEKLY --start 2026-09-29)")
			}
			if types == "" {
				types = "TIMESHEET"
			}
			if period == "" {
				return fmt.Errorf("must specify --period (WEEKLY or MONTHLY)")
			}
			if start == "" {
				return fmt.Errorf("must specify --start (period start date)")
			}

			mappedType, err := clockify.MapApprovalType(types)
			if err != nil {
				return err
			}
			mappedPeriod, err := clockify.MapApprovalPeriod(period)
			if err != nil {
				return err
			}
			startISO, err := parseDateTimeCLI(start)
			if err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			req := clockify.SubmitApprovalRequest{
				Period:      mappedPeriod,
				PeriodStart: startISO,
				Note:        note,
			}

			var r *clockify.ApprovalRequest
			if userRef != "" {
				uid, err := resolveUserRef(ctx, client, userRef)
				if err != nil {
					return err
				}
				r, err = client.SubmitApprovalRequestForUser(ctx, uid, mappedType, req)
			} else {
				r, err = client.SubmitApprovalRequest(ctx, mappedType, req)
			}
			if err != nil {
				return err
			}
			return formatOutput(cmd, r, outputFormat, []string{"ID", "USER_ID", "STATUS", "TYPE", "START", "END", "NOTE"})
		},
	}
	cmd.Flags().String("type", "", "TIMESHEET, EXPENSE, or TIMESHEET_AND_EXPENSE (default TIMESHEET)")
	cmd.Flags().String("period", "", "WEEKLY or MONTHLY (required)")
	cmd.Flags().String("start", "", "period start date (required)")
	cmd.Flags().String("user", "", "submit on behalf of user ID, email, or name")
	cmd.Flags().String("note", "", "submission note")
	cmd.Flags().String("time-entry", "", "unsupported (official API uses period-based submission)")
	return cmd
}

func newApprovalUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an approval request status",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			status, _ := cmd.Flags().GetString("status")
			note, _ := cmd.Flags().GetString("note")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			if status == "" {
				return fmt.Errorf("must specify --status (APPROVED, REJECTED, PENDING, WITHDRAWN_SUBMISSION, WITHDRAWN_APPROVAL)")
			}
			mapped, err := clockify.MapApprovalState(status)
			if err != nil {
				return err
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			r, err := client.UpdateApprovalStatus(context.Background(), id, clockify.UpdateApprovalStatusRequest{
				Note:  note,
				State: mapped,
			})
			if err != nil {
				return err
			}
			return formatOutput(cmd, r, outputFormat, []string{"ID", "USER_ID", "STATUS", "TYPE", "START", "END", "NOTE"})
		},
	}
	cmd.Flags().String("id", "", "approval request ID")
	cmd.Flags().String("status", "", "PENDING, APPROVED, REJECTED, WITHDRAWN_SUBMISSION, WITHDRAWN_APPROVAL")
	cmd.Flags().String("note", "", "note for the status change")
	return cmd
}
