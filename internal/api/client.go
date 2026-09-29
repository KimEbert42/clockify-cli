package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kim/clockify-cli/internal/errors"
	"github.com/kim/clockify-cli/pkg/clockify"
)

type ClockifyClient struct {
	httpClient  *http.Client
	baseURL     string
	apiKey      string
	workspaceID string
	rateLimiter *RateLimiter
	retryPolicy *RetryPolicy
	verbose     bool
	debug       bool
}

type ClientOption func(*ClockifyClient)

func WithRetry(policy RetryPolicy) ClientOption {
	return func(c *ClockifyClient) {
		c.retryPolicy = &policy
	}
}

func WithRateLimit(requestsPerSecond int) ClientOption {
	return func(c *ClockifyClient) {
		c.rateLimiter = NewRateLimiter(requestsPerSecond)
	}
}

func WithVerbose(verbose bool) ClientOption {
	return func(c *ClockifyClient) {
		c.verbose = verbose
	}
}

func WithDebug(debug bool) ClientOption {
	return func(c *ClockifyClient) {
		c.debug = debug
	}
}

func NewClient(apiKey, baseURL, workspaceID string, opts ...ClientOption) *ClockifyClient {
	c := &ClockifyClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     baseURL,
		apiKey:      apiKey,
		workspaceID: workspaceID,
		retryPolicy: &DefaultRetryPolicy,
		rateLimiter: NewRateLimiter(45),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *ClockifyClient) Do(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return errors.Wrap(errors.E010Internal, "failed to marshal request body", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return errors.Wrap(errors.E010Internal, "failed to create request", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.verbose {
		fmt.Printf(">>> %s %s\n", method, url)
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		c.rateLimiter.Take()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if c.retryPolicy.RetryableFn(0, err) && attempt < c.retryPolicy.MaxRetries {
				time.Sleep(c.retryPolicy.Delay(attempt))
				continue
			}
			return errors.Wrap(errors.E006Network, "request failed", err)
		}

		defer resp.Body.Close()

		if c.debug {
			fmt.Printf("<<< %d %s\n", resp.StatusCode, resp.Status)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if result != nil && resp.StatusCode != 204 {
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return errors.Wrap(errors.E010Internal, "failed to read response body", err)
				}
				if err := json.Unmarshal(respBody, result); err != nil {
					return errors.Wrap(errors.E010Internal, "failed to unmarshal response", err)
				}
			}
			return nil
		}

		if c.retryPolicy.RetryableFn(resp.StatusCode, nil) && attempt < c.retryPolicy.MaxRetries {
			time.Sleep(c.retryPolicy.Delay(attempt))
			continue
		}

		return c.handleErrorResponse(resp)
	}

	return errors.Wrap(errors.E006Network, "max retries exceeded", lastErr)
}

func (c *ClockifyClient) handleErrorResponse(resp *http.Response) error {
	respBody, _ := io.ReadAll(resp.Body)
	code := errors.FromHTTPStatus(resp.StatusCode)

	msg := fmt.Sprintf("API request failed with status %d", resp.StatusCode)
	detail := string(respBody)

	switch code {
	case errors.E001Auth:
		return &errors.CLIError{Code: code, Message: msg, Detail: detail, Suggestion: "Run 'clockify configure' to set up your API key"}
	case errors.E002NotFound:
		return &errors.CLIError{Code: code, Message: msg, Detail: detail, Suggestion: "Check the resource ID or name"}
	case errors.E005RateLimit:
		return &errors.CLIError{Code: code, Message: msg, Detail: detail, Suggestion: "Wait a moment and try again"}
	default:
		return &errors.CLIError{Code: code, Message: msg, Detail: detail}
	}
}

func (c *ClockifyClient) WorkspaceID() string {
	return c.workspaceID
}

func (c *ClockifyClient) DoWithPagination(ctx context.Context, method, path string, body any, result any) (*clockify.PageInfo, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(errors.E010Internal, "failed to marshal request body", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, errors.Wrap(errors.E010Internal, "failed to create request", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.verbose {
		fmt.Printf(">>> %s %s\n", method, url)
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		c.rateLimiter.Take()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if c.retryPolicy.RetryableFn(0, err) && attempt < c.retryPolicy.MaxRetries {
				time.Sleep(c.retryPolicy.Delay(attempt))
				continue
			}
			return nil, errors.Wrap(errors.E006Network, "request failed", err)
		}

		defer resp.Body.Close()

		if c.debug {
			fmt.Printf("<<< %d %s\n", resp.StatusCode, resp.Status)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if result != nil && resp.StatusCode != 204 {
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, errors.Wrap(errors.E010Internal, "failed to read response body", err)
				}
				if err := json.Unmarshal(respBody, result); err != nil {
					return nil, errors.Wrap(errors.E010Internal, "failed to unmarshal response", err)
				}
			}

			pageInfo := &clockify.PageInfo{}
			if tp := resp.Header.Get("X-Pagination-Total-Pages"); tp != "" {
				fmt.Sscanf(tp, "%d", &pageInfo.TotalPages)
			}
			if tr := resp.Header.Get("X-Pagination-Total-Results"); tr != "" {
				fmt.Sscanf(tr, "%d", &pageInfo.TotalResults)
			}
			return pageInfo, nil
		}

		if c.retryPolicy.RetryableFn(resp.StatusCode, nil) && attempt < c.retryPolicy.MaxRetries {
			time.Sleep(c.retryPolicy.Delay(attempt))
			continue
		}

		return nil, c.handleErrorResponse(resp)
	}

	return nil, errors.Wrap(errors.E006Network, "max retries exceeded", lastErr)
}
