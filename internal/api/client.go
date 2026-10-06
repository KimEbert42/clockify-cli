package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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

// MaskAPIKey returns a redacted form of an API key for logs and diagnostics.
// Short keys become "****"; longer keys keep a 4-char prefix and suffix.
func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
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
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return errors.Wrap(errors.E010Internal, "failed to marshal request body", err)
		}
	}

	url := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return errors.Wrap(errors.E010Internal, "failed to create request", err)
		}
		if bodyBytes != nil {
			req.ContentLength = int64(len(bodyBytes))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(bodyBytes)), nil
			}
		}

		req.Header.Set("X-Api-Key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		if c.verbose {
			fmt.Fprintf(os.Stderr, ">>> %s %s\n", method, url)
		}

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

		if c.debug {
			fmt.Fprintf(os.Stderr, "<<< %d %s (api-key=%s)\n", resp.StatusCode, resp.Status, MaskAPIKey(c.apiKey))
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if result != nil && resp.StatusCode != 204 {
				respBody, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					return errors.Wrap(errors.E010Internal, "failed to read response body", err)
				}
				if err := json.Unmarshal(respBody, result); err != nil {
					return errors.Wrap(errors.E010Internal, "failed to unmarshal response", err)
				}
			} else {
				resp.Body.Close()
			}
			return nil
		}

		if c.retryPolicy.RetryableFn(resp.StatusCode, nil) && attempt < c.retryPolicy.MaxRetries {
			resp.Body.Close()
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
					time.Sleep(time.Duration(secs) * time.Second)
					continue
				}
			}
			time.Sleep(c.retryPolicy.Delay(attempt))
			continue
		}

		err = c.handleErrorResponse(resp)
		resp.Body.Close()
		return err
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
	case errors.E003Permission:
		return &errors.CLIError{Code: code, Message: msg, Detail: detail, Suggestion: "Ask a workspace admin for a role with the required permissions"}
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
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(errors.E010Internal, "failed to marshal request body", err)
		}
	}

	url := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, errors.Wrap(errors.E010Internal, "failed to create request", err)
		}
		if bodyBytes != nil {
			req.ContentLength = int64(len(bodyBytes))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(bodyBytes)), nil
			}
		}

		req.Header.Set("X-Api-Key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		if c.verbose {
			fmt.Fprintf(os.Stderr, ">>> %s %s\n", method, url)
		}

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

		if c.debug {
			fmt.Fprintf(os.Stderr, "<<< %d %s (api-key=%s)\n", resp.StatusCode, resp.Status, MaskAPIKey(c.apiKey))
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			pageInfo := &clockify.PageInfo{}
			if tp := resp.Header.Get("X-Pagination-Total-Pages"); tp != "" {
				fmt.Sscanf(tp, "%d", &pageInfo.TotalPages)
			}
			if tr := resp.Header.Get("X-Pagination-Total-Results"); tr != "" {
				fmt.Sscanf(tr, "%d", &pageInfo.TotalResults)
			}
			if p := resp.Header.Get("X-Pagination-Page"); p != "" {
				fmt.Sscanf(p, "%d", &pageInfo.Page)
			}
			if ps := resp.Header.Get("X-Pagination-Page-Size"); ps != "" {
				fmt.Sscanf(ps, "%d", &pageInfo.PageSize)
			}
			if result != nil && resp.StatusCode != 204 {
				respBody, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					return nil, errors.Wrap(errors.E010Internal, "failed to read response body", err)
				}
				if err := json.Unmarshal(respBody, result); err != nil {
					return nil, errors.Wrap(errors.E010Internal, "failed to unmarshal response", err)
				}
			} else {
				resp.Body.Close()
			}
			return pageInfo, nil
		}

		if c.retryPolicy.RetryableFn(resp.StatusCode, nil) && attempt < c.retryPolicy.MaxRetries {
			resp.Body.Close()
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
					time.Sleep(time.Duration(secs) * time.Second)
					continue
				}
			}
			time.Sleep(c.retryPolicy.Delay(attempt))
			continue
		}

		err = c.handleErrorResponse(resp)
		resp.Body.Close()
		return nil, err
	}

	return nil, errors.Wrap(errors.E006Network, "max retries exceeded", lastErr)
}
