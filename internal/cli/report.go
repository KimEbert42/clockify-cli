package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/internal/output"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/kim/clockify-cli/pkg/duration"
	"github.com/spf13/cobra"
)

// CLI status values accepted by --status (spec §8.11).
var validReportStatuses = map[string]bool{
	"IN_PROGRESS": true,
	"NEW":         true,
	"APPROVED":    true,
	"REJECTED":    true,
}

func newReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate time reports",
		Long: `Generate Clockify summary, detailed, and weekly reports.

Reports use the Clockify Reports API (separate host/path from the main API).
Date shortcuts: today, yesterday, this-week, last-week, this-month, last-month.

Presets (summary with matching date range):
  report today|yesterday|this-week|last-week|this-month|last-month`,
	}
	cmd.AddCommand(newReportSummaryCmd())
	cmd.AddCommand(newReportDetailedCmd())
	cmd.AddCommand(newReportWeeklyCmd())
	cmd.AddCommand(newReportPresetCmd("today", "today", "today"))
	cmd.AddCommand(newReportPresetCmd("yesterday", "yesterday", "yesterday"))
	cmd.AddCommand(newReportPresetCmd("this-week", "this-week", "today"))
	cmd.AddCommand(newReportPresetCmd("last-week", "last-week", "last-week+6d"))
	cmd.AddCommand(newReportPresetCmd("this-month", "this-month", "today"))
	cmd.AddCommand(newReportPresetCmd("last-month", "last-month", "last-month+1m-1d"))
	return cmd
}

// newReportPresetCmd builds a summary-report shortcut for a date preset.
func newReportPresetCmd(name, startRel, endRel string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Summary report for " + name,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("start") {
				_ = cmd.Flags().Set("start", startRel)
			}
			if !cmd.Flags().Changed("end") {
				end := endRel
				// Resolve pseudo end expressions like last-week+6d
				if strings.Contains(end, "+") {
					base := strings.SplitN(end, "+", 2)[0]
					delta := strings.SplitN(end, "+", 2)[1]
					t, err := duration.ParseRelativeDate(base, "monday")
					if err != nil {
						return err
					}
					switch {
					case strings.HasSuffix(delta, "d") && !strings.Contains(delta, "m"):
						var n int
						if _, err := fmt.Sscanf(delta, "%dd", &n); err == nil {
							t = t.AddDate(0, 0, n)
						}
					case strings.Contains(delta, "m-1d"):
						var n int
						if _, err := fmt.Sscanf(delta, "%dm-1d", &n); err == nil {
							t = t.AddDate(0, n, -1)
						}
					}
					end = t.Format("2006-01-02")
				}
				_ = cmd.Flags().Set("end", end)
			}
			kind := "summary"
			if detailed, _ := cmd.Flags().GetBool("detailed"); detailed {
				kind = "detailed"
			}
			return runReport(kind, true)(cmd, args)
		},
	}
	addReportCommonFlags(cmd, true)
	cmd.Flags().Bool("billable", false, "show only billable time")
	cmd.Flags().String("description", "", "filter by description")
	cmd.Flags().Bool("detailed", false, "use detailed report instead of summary")
	return cmd
}

func addReportCommonFlags(cmd *cobra.Command, requireEnd bool) {
	cmd.Flags().String("start", "", "start date (YYYY-MM-DD or relative)")
	if requireEnd {
		cmd.Flags().String("end", "", "end date (YYYY-MM-DD or relative)")
	} else {
		cmd.Flags().String("end", "", "end date (YYYY-MM-DD or relative; default: start + 6 days)")
	}
	cmd.Flags().String("project", "", "filter by project (ID or name)")
	cmd.Flags().String("task", "", "filter by task (ID or name; requires --project)")
	cmd.Flags().String("user", "", "filter by user (ID, email, or name)")
	cmd.Flags().String("client", "", "filter by client (ID or name)")
	cmd.Flags().String("group", "", "filter by user group ID")
	cmd.Flags().String("status", "", "filter by approval state: IN_PROGRESS, NEW, APPROVED, REJECTED")
	cmd.Flags().String("export", "", "export format: csv or pdf")
	cmd.Flags().String("file", "", "output path for --export, or JSON file with report parameters")
	cmd.Flags().Bool("stdin", false, "read report parameters JSON from stdin")
	_ = cmd.RegisterFlagCompletionFunc("project", completeProjectRefC)
	_ = cmd.RegisterFlagCompletionFunc("user", completeUserRefC)
	_ = cmd.RegisterFlagCompletionFunc("client", completeClientRefC)
	_ = cmd.RegisterFlagCompletionFunc("task", completeTaskRefC)
}

func newReportSummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Summary report by project/user",
		RunE:  runReport("summary", true),
	}
	addReportCommonFlags(cmd, true)
	cmd.Flags().Bool("billable", false, "show only billable time")
	cmd.Flags().String("description", "", "filter by description")
	return cmd
}

func newReportDetailedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detailed",
		Short: "Detailed time entry report",
		RunE:  runReport("detailed", true),
	}
	addReportCommonFlags(cmd, true)
	return cmd
}

func newReportWeeklyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weekly",
		Short: "Weekly report starting from a given date",
		RunE:  runReport("weekly", false),
	}
	addReportCommonFlags(cmd, false)
	return cmd
}

func runReport(kind string, requireEnd bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		outputFormat, _ := cmd.Flags().GetString("output")
		quiet, _ := cmd.Flags().GetBool("quiet")
		startRaw, _ := cmd.Flags().GetString("start")
		endRaw, _ := cmd.Flags().GetString("end")
		projectRef, _ := cmd.Flags().GetString("project")
		taskRef, _ := cmd.Flags().GetString("task")
		userRef, _ := cmd.Flags().GetString("user")
		clientRef, _ := cmd.Flags().GetString("client")
		groupID, _ := cmd.Flags().GetString("group")
		statusRaw, _ := cmd.Flags().GetString("status")
		exportFormat, _ := cmd.Flags().GetString("export")
		filePath, _ := cmd.Flags().GetString("file")
		stdinParams, _ := cmd.Flags().GetBool("stdin")
		billableOnly, _ := cmd.Flags().GetBool("billable")
		description, _ := cmd.Flags().GetString("description")

		// Optional JSON params from --file or --stdin (spec §8.11 --file as input).
		// When --export is set, --file is the output path instead.
		var fileParams map[string]any
		if stdinParams && exportFormat == "" {
			params := map[string]any{}
			if err := readJSONInput("", true, &params); err != nil {
				return err
			}
			fileParams = params
		} else if filePath != "" && exportFormat == "" {
			params, err := readReportParamsFile(filePath)
			if err != nil {
				return err
			}
			fileParams = params
		}

		if startRaw == "" {
			if v, ok := fileParams["start"].(string); ok && v != "" {
				startRaw = v
			}
		}
		if startRaw == "" {
			return fmt.Errorf("must specify --start")
		}
		if endRaw == "" {
			if v, ok := fileParams["end"].(string); ok && v != "" {
				endRaw = v
			}
		}
		if requireEnd && endRaw == "" && kind != "weekly" {
			return fmt.Errorf("must specify --end")
		}

		apiKey, region, cfg, err := loadAuthWithConfig(profile)
		if err != nil {
			return err
		}
		if cfg.WorkspaceID == "" {
			return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
		}

		ctx := context.Background()
		mainClient := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
		reportsClient := api.NewClient(apiKey, api.ReportsBaseURL(region), cfg.WorkspaceID)

		startTime, err := parseReportDate(startRaw)
		if err != nil {
			return fmt.Errorf("invalid start date: %w", err)
		}
		var endTime time.Time
		if endRaw != "" {
			endTime, err = parseReportDate(endRaw)
			if err != nil {
				return fmt.Errorf("invalid end date: %w", err)
			}
		} else if kind == "weekly" {
			endTime = startTime.AddDate(0, 0, 6)
		} else {
			return fmt.Errorf("must specify --end")
		}

		// Validate status before network calls
		approvalState, err := mapReportStatus(statusRaw)
		if err != nil {
			return err
		}

		req := clockify.ReportRequest{
			DateRangeStart: formatReportDate(startTime, false),
			DateRangeEnd:   formatReportDate(endTime, true),
			WeekStart:      clockify.ReportWeekStartMonday,
			ApprovalState:  approvalState,
		}

		// Resolve filters via main API client
		if projectRef != "" {
			proj, err := resolveProjectRef(ctx, mainClient, projectRef)
			if err != nil {
				return err
			}
			req.Projects = &clockify.ContainsFilter{
				Contains: clockify.ReportContains,
				IDs:      []string{proj.ID},
				Status:   clockify.ReportStatusActive,
			}
		}
		if taskRef != "" {
			if req.Projects == nil || len(req.Projects.IDs) == 0 {
				return fmt.Errorf("--task requires --project")
			}
			task, err := resolveTaskByName(ctx, mainClient, req.Projects.IDs[0], taskRef)
			if err != nil {
				return err
			}
			req.Tasks = &clockify.ContainsFilter{
				Contains: clockify.ReportContains,
				IDs:      []string{task.ID},
				Status:   clockify.ReportStatusActive,
			}
		}
		if userRef != "" {
			userID, err := resolveUserRef(ctx, mainClient, userRef)
			if err != nil {
				return err
			}
			req.Users = &clockify.ContainsFilter{
				Contains: clockify.ReportContains,
				IDs:      []string{userID},
				Status:   clockify.ReportStatusActive,
			}
		}
		if clientRef != "" {
			client, err := resolveClientByName(ctx, mainClient, clientRef)
			if err != nil {
				return err
			}
			req.Clients = &clockify.ContainsFilter{
				Contains: clockify.ReportContains,
				IDs:      []string{client.ID},
				Status:   clockify.ReportStatusActive,
			}
		}
		if groupID != "" {
			req.UserGroups = &clockify.ContainsFilter{
				Contains: clockify.ReportContains,
				IDs:      []string{groupID},
				Status:   clockify.ReportStatusActive,
			}
		}
		if kind == "summary" {
			if billableOnly {
				req.Billable = boolPtr(true)
			}
			if description != "" {
				req.Description = description
			}
			req.SummaryFilter = &clockify.SummaryFilter{SummaryChartType: "PROJECT"}
		}
		if kind == "detailed" {
			req.DetailedFilter = &clockify.DetailedFilter{PageSize: 500}
		}
		if kind == "weekly" {
			req.WeeklyFilter = &clockify.WeeklyFilter{}
		}

		// Export handling
		exportFormat = strings.ToLower(strings.TrimSpace(exportFormat))
		if exportFormat != "" && exportFormat != "csv" && exportFormat != "pdf" {
			return fmt.Errorf("invalid --export %q; must be csv or pdf", exportFormat)
		}
		if exportFormat != "" {
			req.ExportType = strings.ToUpper(exportFormat)
		}

		// Merge any remaining file params that map onto known request fields
		if fileParams != nil {
			mergeReportFileParams(&req, fileParams)
		}

		var body []byte
		var contentType string
		switch kind {
		case "summary":
			body, contentType, err = reportsClient.GenerateSummaryReport(ctx, req)
		case "detailed":
			body, contentType, err = reportsClient.GenerateDetailedReport(ctx, req)
		case "weekly":
			body, contentType, err = reportsClient.GenerateWeeklyReport(ctx, req)
		default:
			return fmt.Errorf("unknown report type %q", kind)
		}
		if err != nil {
			return err
		}

		// Export path
		if exportFormat == "csv" {
			return writeReportCSV(cmd, body, contentType, filePath, kind, outputFormat, quiet)
		}
		if exportFormat == "pdf" {
			return writeReportPDF(cmd, body, contentType, filePath, quiet)
		}

		// Normal display path
		if api.LooksLikeCSV(contentType, body) {
			// API returned CSV even without --export — print as-is
			if filePath != "" {
				return os.WriteFile(filePath, body, 0644)
			}
			cmd.OutOrStdout().Write(body)
			return nil
		}

		resp, err := api.ParseReportResponse(body)
		if err != nil {
			// Fall back to raw JSON dump
			if outputFormat == "json" || outputFormat == "" {
				var pretty any
				if json.Unmarshal(body, &pretty) == nil {
					return formatOutput(cmd, pretty, outputFormat, nil)
				}
			}
			return err
		}

		displays, headers := buildReportDisplays(kind, resp)
		if len(displays) == 0 && !quiet {
			cmd.Println("No report data returned.")
			return nil
		}

		// JSON mode: pass structured data (with raw if useful)
		if outputFormat == "json" {
			if resp.Raw != nil {
				return formatOutput(cmd, resp.Raw, outputFormat, nil)
			}
		}

		return formatOutput(cmd, displays, outputFormat, headers)
	}
}

func boolPtr(b bool) *bool { return &b }

func readReportParamsFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read report params file: %w", err)
	}
	var params map[string]any
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, fmt.Errorf("invalid JSON in report params file: %w", err)
	}
	return params, nil
}

func mergeReportFileParams(req *clockify.ReportRequest, params map[string]any) {
	if v, ok := params["billable"].(bool); ok && req.Billable == nil {
		req.Billable = &v
	}
	if v, ok := params["description"].(string); ok && req.Description == "" {
		req.Description = v
	}
	if v, ok := params["exportType"].(string); ok && req.ExportType == "" {
		req.ExportType = v
	}
	if v, ok := params["approvalState"].(string); ok && req.ApprovalState == "" {
		req.ApprovalState = v
	}
	if v, ok := params["dateRangeType"].(string); ok {
		req.DateRangeType = v
	}
}

// parseReportDate accepts relative shortcuts, human times, and YYYY-MM-DD / RFC3339 dates.
func parseReportDate(s string) (time.Time, error) {
	return parseHumanStart(s, nil)
}

// formatReportDate renders times in the Clockify reports API format
// (YYYY-MM-DDTHH:MM:SS.ssssss — microsecond precision).
func formatReportDate(t time.Time, endOfDay bool) string {
	if endOfDay {
		return t.Format("2006-01-02") + "T23:59:59.999999"
	}
	return t.Format("2006-01-02") + "T00:00:00.000000"
}

// mapReportStatus maps CLI --status values to Clockify approvalState.
func mapReportStatus(status string) (string, error) {
	if status == "" {
		return "", nil
	}
	if !validReportStatuses[status] {
		return "", fmt.Errorf("invalid status %q; must be one of: IN_PROGRESS, NEW, APPROVED, REJECTED", status)
	}
	switch status {
	case "APPROVED":
		return clockify.ReportApprovalApproved, nil
	case "NEW", "IN_PROGRESS":
		// Clockify reports API exposes APPROVED/UNAPPROVED/ALL.
		// NEW and IN_PROGRESS both map to unapproved entries.
		return clockify.ReportApprovalUnapproved, nil
	case "REJECTED":
		// Not represented in the reports API approvalState enum.
		return "", fmt.Errorf("status REJECTED is not supported by the Clockify reports API; filter rejected entries client-side or omit --status")
	}
	return "", nil
}

type reportDisplay struct {
	Group       string `json:"group" yaml:"group"`
	Date        string `json:"date" yaml:"date"`
	Project     string `json:"project" yaml:"project"`
	User        string `json:"user" yaml:"user"`
	Description string `json:"description" yaml:"description"`
	Duration    string `json:"duration" yaml:"duration"`
	Billable    string `json:"billable" yaml:"billable"`
	Amount      string `json:"amount" yaml:"amount"`
}

func buildReportDisplays(kind string, resp *clockify.ReportResponse) ([]reportDisplay, []string) {
	headers := []string{"GROUP", "DATE", "DURATION", "BILLABLE", "AMOUNT", "DESCRIPTION"}

	switch kind {
	case "detailed":
		headers = []string{"DATE", "DESCRIPTION", "DURATION", "BILLABLE"}
		var rows []reportDisplay
		entries := resp.TimeEntries
		if len(entries) == 0 {
			// Some payloads nest entries under groups
			for _, g := range resp.Groups {
				entries = append(entries, g.TimeEntries...)
			}
		}
		for _, e := range entries {
			start := ""
			dur := ""
			billable := "false"
			if e.TimeInterval != nil {
				start = e.TimeInterval.Start
				if start != "" {
					if t, err := time.Parse(time.RFC3339, start); err == nil {
						start = t.Format("2006-01-02")
					} else if t, err := time.Parse("2006-01-02T15:04:05.000000", start); err == nil {
						start = t.Format("2006-01-02")
					}
				}
				if e.TimeInterval.Duration != nil {
					dur = duration.Format(*e.TimeInterval.Duration)
				}
			}
			if e.Billable {
				billable = "true"
			}
			rows = append(rows, reportDisplay{
				Date:        start,
				Description: e.Description,
				Duration:    dur,
				Billable:    billable,
			})
		}
		return rows, headers

	case "summary", "weekly":
		var rows []reportDisplay
		groups := resp.Groups
		if len(groups) == 0 {
			groups = resp.Days
		}
		for _, g := range groups {
			row := reportDisplay{Group: g.Name, Date: g.Date}
			if g.Totals != nil {
				if g.Totals.Duration != nil {
					row.Duration = duration.Format(*g.Totals.Duration)
				}
				if g.Totals.BillableDuration != nil {
					row.Billable = duration.Format(*g.Totals.BillableDuration)
				}
				if g.Totals.Amount != nil {
					row.Amount = fmt.Sprintf("%.2f", *g.Totals.Amount)
				}
			} else if len(g.TimeEntries) > 0 {
				var total int64
				for _, e := range g.TimeEntries {
					if e.TimeInterval != nil && e.TimeInterval.Duration != nil {
						total += *e.TimeInterval.Duration
					}
				}
				row.Duration = duration.Format(total)
			}
			if row.Group == "" && row.Date == "" && row.Duration == "" {
				continue
			}
			rows = append(rows, row)
		}

		// Top-level totals as a final row
		if resp.Totals != nil {
			totalDur := ""
			if resp.Totals.Duration != nil {
				totalDur = duration.Format(*resp.Totals.Duration)
			}
			amount := ""
			if resp.Totals.Amount != nil {
				amount = fmt.Sprintf("%.2f", *resp.Totals.Amount)
			}
			if totalDur != "" || amount != "" {
				rows = append(rows, reportDisplay{
					Group:    "TOTAL",
					Duration: totalDur,
					Amount:   amount,
				})
			}
		}
		return rows, headers
	}
	return nil, headers
}

func writeReportCSV(cmd *cobra.Command, body []byte, contentType, filePath, kind, outputFormat string, quiet bool) error {
	var csvBytes []byte

	if api.LooksLikeCSV(contentType, body) {
		csvBytes = body
	} else {
		// Client-side CSV from JSON response
		resp, err := api.ParseReportResponse(body)
		if err != nil {
			return fmt.Errorf("report response is neither CSV nor valid JSON: %w", err)
		}
		displays, headers := buildReportDisplays(kind, resp)
		var buf strings.Builder
		f := output.CSVFormatter{}
		// CSVFormatter needs io.Writer; use a small adapter
		if err := writeCSVTo(&buf, f, displays, headers); err != nil {
			return err
		}
		csvBytes = []byte(buf.String())
	}

	if filePath != "" {
		if err := os.WriteFile(filePath, csvBytes, 0644); err != nil {
			return fmt.Errorf("failed to write CSV file: %w", err)
		}
		if !quiet {
			cmd.Printf("Report exported to %s\n", filePath)
		}
		return nil
	}

	_, err := cmd.OutOrStdout().Write(csvBytes)
	if err != nil {
		return err
	}
	if len(csvBytes) > 0 && csvBytes[len(csvBytes)-1] != '\n' {
		cmd.Println()
	}
	return nil
}

func writeReportPDF(cmd *cobra.Command, body []byte, contentType, filePath string, quiet bool) error {
	if !api.LooksLikePDF(contentType, body) && filePath == "" {
		return fmt.Errorf("report API did not return PDF data; try --export csv or --output json")
	}
	if filePath == "" {
		return fmt.Errorf("--export pdf requires --file <path>")
	}
	if err := os.WriteFile(filePath, body, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}
	if !quiet {
		cmd.Printf("Report exported to %s\n", filePath)
	}
	return nil
}

type csvWriterAdapter struct {
	buf *strings.Builder
}

func (w csvWriterAdapter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func writeCSVTo(buf *strings.Builder, f output.CSVFormatter, data any, headers []string) error {
	return f.Format(csvWriterAdapter{buf: buf}, data, headers)
}
