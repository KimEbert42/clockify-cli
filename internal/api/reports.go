package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kim/clockify-cli/internal/errors"
	"github.com/kim/clockify-cli/pkg/clockify"
)

// GenerateDetailedReport POSTs to /workspaces/{id}/reports/detailed.
// detailedFilter must be set by the caller (required by the API).
func (c *ClockifyClient) GenerateDetailedReport(ctx context.Context, req clockify.ReportRequest) ([]byte, string, error) {
	if req.DetailedFilter == nil {
		req.DetailedFilter = &clockify.DetailedFilter{}
	}
	return c.postReport(ctx, "detailed", req)
}

// GenerateSummaryReport POSTs to /workspaces/{id}/reports/summary.
// summaryFilter must be set by the caller (required by the API).
// Clockify requires summaryFilter.groups with 1-3 values; default to PROJECT.
func (c *ClockifyClient) GenerateSummaryReport(ctx context.Context, req clockify.ReportRequest) ([]byte, string, error) {
	if req.SummaryFilter == nil {
		req.SummaryFilter = &clockify.SummaryFilter{}
	}
	if len(req.SummaryFilter.Groups) == 0 {
		req.SummaryFilter.Groups = []string{"PROJECT"}
		if req.SummaryFilter.SummaryChartType == "" {
			req.SummaryFilter.SummaryChartType = "PROJECT"
		}
	}
	return c.postReport(ctx, "summary", req)
}

// GenerateWeeklyReport POSTs to /workspaces/{id}/reports/weekly.
// weeklyFilter must be set by the caller (required by the API).
func (c *ClockifyClient) GenerateWeeklyReport(ctx context.Context, req clockify.ReportRequest) ([]byte, string, error) {
	if req.WeeklyFilter == nil {
		req.WeeklyFilter = &clockify.WeeklyFilter{}
	}
	return c.postReport(ctx, "weekly", req)
}

// postReport sends a report generation request and returns the raw response body
// plus Content-Type. Unlike Do(), it does not force JSON unmarshaling so that
// CSV/PDF export responses can pass through.
func (c *ClockifyClient) postReport(ctx context.Context, kind string, body any) ([]byte, string, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, "", errors.Wrap(errors.E010Internal, "failed to marshal report request", err)
	}

	path := fmt.Sprintf("/workspaces/%s/reports/%s", c.workspaceID, kind)
	url := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.retryPolicy.Delay(attempt - 1)
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(delay):
			}
		}

		c.rateLimiter.Take()

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
		if err != nil {
			return nil, "", errors.Wrap(errors.E010Internal, "failed to create report request", err)
		}
		req.Header.Set("X-Api-Key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/csv, application/pdf, */*")

		if c.verbose {
			fmt.Fprintf(os.Stderr, ">>> POST %s\n", url)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if c.retryPolicy.RetryableFn(0, err) {
				continue
			}
			return nil, "", errors.Wrap(errors.E005RateLimit, "report request failed", err)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if c.debug {
			fmt.Fprintf(os.Stderr, "<<< %s (%d bytes)\n", resp.Status, len(respBody))
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, resp.Header.Get("Content-Type"), nil
		}

		if c.retryPolicy.RetryableFn(resp.StatusCode, nil) {
			lastErr = fmt.Errorf("report API returned %s", resp.Status)
			continue
		}

		apiErr := reportErrorFromStatus(resp.StatusCode, respBody)
		return nil, "", apiErr
	}

	if lastErr != nil {
		return nil, "", fmt.Errorf("report request failed after retries: %w", lastErr)
	}
	return nil, "", fmt.Errorf("report request failed after retries")
}

func reportErrorFromStatus(statusCode int, body []byte) error {
	code := errors.FromHTTPStatus(statusCode)
	msg := fmt.Sprintf("Report API request failed with status %d", statusCode)
	detail := string(body)

	var apiMsg string
	var apiErrObj struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &apiErrObj) == nil && apiErrObj.Message != "" {
		apiMsg = apiErrObj.Message
	}

	switch code {
	case errors.E001Auth:
		return &errors.CLIError{Code: code, Message: msg, Detail: detail, Suggestion: "Run 'clockify configure' to set up your API key"}
	case errors.E002NotFound:
		return &errors.CLIError{Code: code, Message: msg, Detail: detail, Suggestion: "Check the workspace ID and report filters"}
	case errors.E005RateLimit:
		return &errors.CLIError{Code: code, Message: msg, Detail: detail, Suggestion: "Wait a moment and try again"}
	default:
		if apiMsg != "" {
			return fmt.Errorf("%s: %s", msg, apiMsg)
		}
		return &errors.CLIError{Code: code, Message: msg, Detail: detail}
	}
}

// ParseReportResponse decodes a reports API JSON body into a flexible ReportResponse.
func ParseReportResponse(body []byte) (*clockify.ReportResponse, error) {
	if len(body) == 0 {
		return &clockify.ReportResponse{}, nil
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid report response: %w", err)
	}

	typedJSON, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var resp clockify.ReportResponse
	if err := json.Unmarshal(typedJSON, &resp); err != nil {
		return nil, fmt.Errorf("failed to map report response: %w", err)
	}
	resp.Raw = raw
	return &resp, nil
}

// LooksLikeCSV reports whether a response body/content-type indicates CSV export output.
func LooksLikeCSV(contentType string, body []byte) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "csv") {
		return true
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return false
	}
	firstLine := trimmed
	if i := strings.IndexByte(trimmed, '\n'); i >= 0 {
		firstLine = trimmed[:i]
	}
	if strings.Contains(firstLine, ",") && !strings.HasPrefix(firstLine, "{") && !strings.HasPrefix(firstLine, "[") {
		return true
	}
	return false
}

// LooksLikePDF reports whether a response indicates PDF export output.
func LooksLikePDF(contentType string, body []byte) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "pdf") {
		return true
	}
	return len(body) >= 4 && string(body[:4]) == "%PDF"
}
