package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newTimeOffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "time-off",
		Aliases: []string{"timeoff", "pto"},
		Short:   "Manage time-off requests",
		Long: `List, create, update status, and delete Clockify time-off requests.

Time-off is policy-scoped in the official API. Use --policy with a policy ID or name.
Spec types (VACATION, SICK, PERSONAL, UNPAID, OTHER) can be used with --type to resolve a matching policy.`,
	}
	cmd.AddCommand(newTimeOffListCmd())
	cmd.AddCommand(newTimeOffCreateCmd())
	cmd.AddCommand(newTimeOffUpdateStatusCmd())
	cmd.AddCommand(newTimeOffDeleteCmd())
	return cmd
}

func newTimeOffListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time-off requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			status, _ := cmd.Flags().GetString("status")
			userRef, _ := cmd.Flags().GetString("user")

			if start == "" {
				start = "this-month"
			}
			if end == "" {
				end = "today"
			}
			startISO, err := parseDateTimeCLI(start)
			if err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}
			endISO, err := parseDateTimeCLI(end)
			if err != nil {
				return fmt.Errorf("invalid --end: %w", err)
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			req := clockify.ListTimeOffRequestsRequest{
				Start: startISO,
				End:   endISO,
			}
			if status != "" {
				mapped, err := clockify.MapTimeOffStatus(status)
				if err != nil {
					return err
				}
				req.Statuses = []string{mapped}
			}
			if userRef != "" {
				uid, err := resolveUserRef(ctx, client, userRef)
				if err != nil {
					return err
				}
				req.Users = []string{uid}
			}

			requests, err := client.ListTimeOffRequests(ctx, req)
			if err != nil {
				return err
			}
			if len(requests) == 0 && !quiet {
				cmd.Println("No time-off requests found.")
				return nil
			}
			type toDisplay struct {
				ID     string `json:"id" yaml:"id"`
				User   string `json:"userId" yaml:"userId"`
				Policy string `json:"policyId" yaml:"policyId"`
				Start  string `json:"start" yaml:"start"`
				End    string `json:"end" yaml:"end"`
				Status string `json:"status" yaml:"status"`
				Note   string `json:"note" yaml:"note"`
			}
			displays := make([]toDisplay, 0, len(requests))
			for _, r := range requests {
				displays = append(displays, toDisplay{
					ID: r.ID, User: r.UserID, Policy: r.PolicyID,
					Start: r.Start, End: r.End, Status: r.Status, Note: r.Note,
				})
			}
			return formatOutput(cmd, displays, outputFormat, []string{"ID", "USER_ID", "POLICY_ID", "START", "END", "STATUS", "NOTE"})
		},
	}
	cmd.Flags().String("start", "", "start of range (default: this-month)")
	cmd.Flags().String("end", "", "end of range (default: today)")
	cmd.Flags().String("status", "", "filter by status: PENDING, APPROVED, REJECTED")
	cmd.Flags().String("user", "", "filter by user ID, email, or name")
	return cmd
}

func newTimeOffCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a time-off request",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			policyRef, _ := cmd.Flags().GetString("policy")
			typeFlag, _ := cmd.Flags().GetString("type")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			note, _ := cmd.Flags().GetString("note")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if filePath != "" || stdin {
				var raw map[string]any
				if err := readJSONInput(filePath, stdin, &raw); err != nil {
					return err
				}
				if v, ok := raw["policyId"].(string); ok && policyRef == "" {
					policyRef = v
				}
				if v, ok := raw["policy"].(string); ok && policyRef == "" {
					policyRef = v
				}
				if v, ok := raw["type"].(string); ok && typeFlag == "" {
					typeFlag = v
				}
				if v, ok := raw["start"].(string); ok && start == "" {
					start = v
				}
				if v, ok := raw["end"].(string); ok && end == "" {
					end = v
				}
				if v, ok := raw["note"].(string); ok && note == "" {
					note = v
				}
			}

			if policyRef == "" && typeFlag == "" {
				return fmt.Errorf("must specify --policy (policy ID or name) or --type (e.g. VACATION)")
			}
			if start == "" || end == "" {
				return fmt.Errorf("must specify --start and --end")
			}

			if typeFlag != "" {
				if err := clockify.ValidateTimeOffType(typeFlag); err != nil {
					return err
				}
			}

			startISO, err := parseDateTimeCLI(start)
			if err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}
			endISO, err := parseDateTimeCLI(end)
			if err != nil {
				return fmt.Errorf("invalid --end: %w", err)
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			ref := policyRef
			if ref == "" {
				ref = typeFlag
			}
			policyID, err := resolvePolicyRef(ctx, client, ref)
			if err != nil {
				return err
			}

			r, err := client.CreateTimeOffRequest(ctx, policyID, clockify.CreateTimeOffRequestRequest{
				Note: note,
				TimeOffPeriod: clockify.TimeOffRequestPeriod{
					Start: startISO,
					End:   endISO,
				},
			})
			if err != nil {
				return err
			}
			return formatOutput(cmd, r, outputFormat, []string{"ID", "USER_ID", "POLICY_ID", "START", "END", "STATUS", "NOTE"})
		},
	}
	cmd.Flags().String("policy", "", "time-off policy ID or name")
	cmd.Flags().String("type", "", "time-off type label: VACATION, SICK, PERSONAL, UNPAID, OTHER (resolves a matching policy)")
	cmd.Flags().String("start", "", "request start date/time (required)")
	cmd.Flags().String("end", "", "request end date/time (required)")
	cmd.Flags().String("note", "", "request note")
	cmd.Flags().String("file", "", "JSON file with time-off data")
	cmd.Flags().Bool("stdin", false, "read time-off data from stdin")
	return cmd
}

func newTimeOffUpdateStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-status",
		Short: "Approve or reject a time-off request",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			status, _ := cmd.Flags().GetString("status")
			note, _ := cmd.Flags().GetString("note")
			policyRef, _ := cmd.Flags().GetString("policy")
			typeFlag, _ := cmd.Flags().GetString("type")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			if status == "" {
				return fmt.Errorf("must specify --status APPROVED or REJECTED")
			}
			mapped, err := clockify.MapTimeOffStatus(status)
			if err != nil {
				return err
			}
			if mapped == "PENDING" {
				return fmt.Errorf("--status must be APPROVED or REJECTED for update-status")
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			ref := policyRef
			if ref == "" {
				ref = typeFlag
			}
			if ref == "" {
				return fmt.Errorf("must specify --policy (policy ID or name); time-off status updates are policy-scoped in the Clockify API")
			}
			policyID, err := resolvePolicyRef(ctx, client, ref)
			if err != nil {
				return err
			}

			r, err := client.UpdateTimeOffStatus(ctx, policyID, id, clockify.UpdateTimeOffStatusRequest{
				Note:   note,
				Status: mapped,
			})
			if err != nil {
				return err
			}
			return formatOutput(cmd, r, outputFormat, []string{"ID", "USER_ID", "POLICY_ID", "START", "END", "STATUS", "NOTE"})
		},
	}
	cmd.Flags().String("id", "", "time-off request ID")
	cmd.Flags().String("status", "", "APPROVED or REJECTED")
	cmd.Flags().String("note", "", "note for the status change")
	cmd.Flags().String("policy", "", "policy ID or name (required)")
	cmd.Flags().String("type", "", "time-off type label used to resolve a policy")
	return cmd
}

func newTimeOffDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a time-off request",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			policyRef, _ := cmd.Flags().GetString("policy")
			typeFlag, _ := cmd.Flags().GetString("type")
			confirm, _ := cmd.Flags().GetBool("confirm")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			ref := policyRef
			if ref == "" {
				ref = typeFlag
			}
			if ref == "" {
				return fmt.Errorf("must specify --policy (policy ID or name)")
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()
			policyID, err := resolvePolicyRef(ctx, client, ref)
			if err != nil {
				return err
			}

			if !confirmDestructive(cmd, confirm, "time-off request", id, strings.TrimSpace(ref)) {
				return nil
			}
			if err := client.DeleteTimeOffRequest(ctx, policyID, id); err != nil {
				return err
			}
			cmd.Printf("Time-off request %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "time-off request ID")
	cmd.Flags().String("policy", "", "policy ID or name (required)")
	cmd.Flags().String("type", "", "time-off type label used to resolve a policy")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}
