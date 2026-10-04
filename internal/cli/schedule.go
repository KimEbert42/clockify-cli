package cli

import (
	"context"
	"fmt"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schedule",
		Aliases: []string{"schedules", "assignment"},
		Short:   "Manage schedule assignments",
		Long: `List, create, update, and delete Clockify schedule assignments.

List requires a date range (--start/--end). If omitted, defaults to this-month → today.`,
	}
	cmd.AddCommand(newScheduleListCmd())
	cmd.AddCommand(newScheduleCreateCmd())
	cmd.AddCommand(newScheduleUpdateCmd())
	cmd.AddCommand(newScheduleDeleteCmd())
	return cmd
}

func newScheduleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List schedule assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			userRef, _ := cmd.Flags().GetString("user")
			projectRef, _ := cmd.Flags().GetString("project")
			name, _ := cmd.Flags().GetString("name")

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

			opts := api.ListAssignmentsOpts{Start: startISO, End: endISO, Name: name}
			if userRef != "" {
				uid, err := resolveUserRef(ctx, client, userRef)
				if err != nil {
					return err
				}
				opts.UserID = uid
			}
			if projectRef != "" {
				proj, err := resolveProjectRef(ctx, client, projectRef)
				if err != nil {
					return err
				}
				opts.ProjectID = proj.ID
			}

			assignments, err := client.ListAssignments(ctx, opts)
			if err != nil {
				return err
			}
			if len(assignments) == 0 && !quiet {
				cmd.Println("No schedule assignments found.")
				return nil
			}
			type schDisplay struct {
				ID          string  `json:"id" yaml:"id"`
				Project     string  `json:"projectId" yaml:"projectId"`
				ProjectName string  `json:"projectName" yaml:"projectName"`
				User        string  `json:"userId" yaml:"userId"`
				UserName    string  `json:"userName" yaml:"userName"`
				HoursPerDay float64 `json:"hoursPerDay" yaml:"hoursPerDay"`
				Billable    bool    `json:"billable" yaml:"billable"`
				Note        string  `json:"note" yaml:"note"`
				Published   bool    `json:"published" yaml:"published"`
			}
			displays := make([]schDisplay, 0, len(assignments))
			for _, a := range assignments {
				displays = append(displays, schDisplay{
					ID: a.ID, Project: a.ProjectID, ProjectName: a.ProjectName,
					User: a.UserID, UserName: a.UserName, HoursPerDay: a.HoursPerDay,
					Billable: a.Billable, Note: a.Note, Published: a.Published,
				})
			}
			return formatOutput(cmd, displays, outputFormat, []string{"ID", "PROJECT_ID", "PROJECT_NAME", "USER_ID", "USER_NAME", "HOURS_PER_DAY", "BILLABLE", "PUBLISHED", "NOTE"})
		},
	}
	cmd.Flags().String("start", "", "range start (default: this-month)")
	cmd.Flags().String("end", "", "range end (default: today)")
	cmd.Flags().String("user", "", "filter by user ID, email, or name")
	cmd.Flags().String("project", "", "filter by project ID or name")
	cmd.Flags().String("name", "", "filter by assignment name")
	return cmd
}

func newScheduleCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a schedule assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			userRef, _ := cmd.Flags().GetString("user")
			projectRef, _ := cmd.Flags().GetString("project")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			hours, _ := cmd.Flags().GetString("hours-per-day")
			durationFlag, _ := cmd.Flags().GetString("duration")
			note, _ := cmd.Flags().GetString("note")
			billable, _ := cmd.Flags().GetBool("billable")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if filePath != "" || stdin {
				var raw map[string]any
				if err := readJSONInput(filePath, stdin, &raw); err != nil {
					return err
				}
				if v, ok := raw["userId"].(string); ok && userRef == "" {
					userRef = v
				}
				if v, ok := raw["projectId"].(string); ok && projectRef == "" {
					projectRef = v
				}
				if v, ok := raw["start"].(string); ok && start == "" {
					start = v
				}
				if v, ok := raw["end"].(string); ok && end == "" {
					end = v
				}
				if v, ok := raw["hoursPerDay"].(float64); ok && hours == "" {
					hours = fmt.Sprint(v)
				}
				if v, ok := raw["note"].(string); ok && note == "" {
					note = v
				}
			}

			if userRef == "" {
				return fmt.Errorf("must specify --user")
			}
			if projectRef == "" {
				return fmt.Errorf("must specify --project")
			}
			if start == "" || end == "" {
				return fmt.Errorf("must specify --start and --end")
			}
			hoursVal := hours
			if hoursVal == "" {
				hoursVal = durationFlag
			}
			if hoursVal == "" {
				return fmt.Errorf("must specify --hours-per-day or --duration (e.g. 8h)")
			}
			hpd, err := parseHoursPerDay(hoursVal)
			if err != nil {
				return err
			}
			if hpd <= 0 {
				return fmt.Errorf("hours-per-day must be > 0")
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

			uid, err := resolveUserRef(ctx, client, userRef)
			if err != nil {
				return err
			}
			proj, err := resolveProjectRef(ctx, client, projectRef)
			if err != nil {
				return err
			}

			a, err := client.CreateAssignment(ctx, clockify.CreateScheduleAssignmentRequest{
				ProjectID:   proj.ID,
				UserID:      uid,
				Start:       startISO,
				End:         endISO,
				HoursPerDay: hpd,
				Billable:    billable,
				Note:        note,
			})
			if err != nil {
				return err
			}
			return formatOutput(cmd, a, outputFormat, []string{"ID", "PROJECT_ID", "PROJECT_NAME", "USER_ID", "USER_NAME", "HOURS_PER_DAY", "BILLABLE", "PUBLISHED", "NOTE"})
		},
	}
	cmd.Flags().String("user", "", "user ID, email, or name (required)")
	cmd.Flags().String("project", "", "project ID or name (required)")
	cmd.Flags().String("start", "", "assignment start date (required)")
	cmd.Flags().String("end", "", "assignment end date (required)")
	cmd.Flags().String("hours-per-day", "", "hours per day (e.g. 8 or 8h)")
	cmd.Flags().String("duration", "", "alias for --hours-per-day (e.g. 8h)")
	cmd.Flags().String("note", "", "assignment note")
	cmd.Flags().Bool("billable", false, "assignment is billable")
	cmd.Flags().String("file", "", "JSON file with assignment data")
	cmd.Flags().Bool("stdin", false, "read assignment data from stdin")
	return cmd
}

func newScheduleUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a schedule assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			start, _ := cmd.Flags().GetString("start")
			end, _ := cmd.Flags().GetString("end")
			hours, _ := cmd.Flags().GetString("hours-per-day")
			durationFlag, _ := cmd.Flags().GetString("duration")
			note, _ := cmd.Flags().GetString("note")
			series, _ := cmd.Flags().GetString("series")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			seriesOpt, err := clockify.MapSeriesUpdateOption(series)
			if err != nil {
				return err
			}

			req := clockify.UpdateScheduleAssignmentRequest{SeriesUpdateOption: seriesOpt}
			if start != "" {
				s, err := parseDateTimeCLI(start)
				if err != nil {
					return err
				}
				req.Start = s
			}
			if end != "" {
				e, err := parseDateTimeCLI(end)
				if err != nil {
					return err
				}
				req.End = e
			}
			hoursVal := hours
			if hoursVal == "" {
				hoursVal = durationFlag
			}
			if hoursVal != "" {
				hpd, err := parseHoursPerDay(hoursVal)
				if err != nil {
					return err
				}
				req.HoursPerDay = &hpd
			}
			if note != "" {
				req.Note = note
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			a, err := client.UpdateAssignment(context.Background(), id, req)
			if err != nil {
				return err
			}
			return formatOutput(cmd, a, outputFormat, []string{"ID", "PROJECT_ID", "PROJECT_NAME", "USER_ID", "USER_NAME", "HOURS_PER_DAY", "BILLABLE", "PUBLISHED", "NOTE"})
		},
	}
	cmd.Flags().String("id", "", "assignment ID")
	cmd.Flags().String("start", "", "new start date")
	cmd.Flags().String("end", "", "new end date")
	cmd.Flags().String("hours-per-day", "", "new hours per day")
	cmd.Flags().String("duration", "", "alias for --hours-per-day")
	cmd.Flags().String("note", "", "new note")
	cmd.Flags().String("series", "", "series option: THIS_ONE, THIS_AND_FOLLOWING, ALL (default THIS_ONE)")
	return cmd
}

func newScheduleDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a schedule assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			series, _ := cmd.Flags().GetString("series")
			confirm, _ := cmd.Flags().GetBool("confirm")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			seriesOpt, err := clockify.MapSeriesUpdateOption(series)
			if err != nil {
				return err
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			if !confirmDestructive(cmd, confirm, "schedule assignment", id, seriesOpt) {
				return nil
			}
			if err := client.DeleteAssignment(context.Background(), id, seriesOpt); err != nil {
				return err
			}
			cmd.Printf("Schedule assignment %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "assignment ID")
	cmd.Flags().String("series", "", "series option: THIS_ONE, THIS_AND_FOLLOWING, ALL (default THIS_ONE)")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}
