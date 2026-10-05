package errors

import (
	"fmt"
)

type ErrorCode int

const (
	E001Auth         ErrorCode = 1
	E002NotFound     ErrorCode = 2
	E003Permission   ErrorCode = 3
	E004Validation   ErrorCode = 4
	E005RateLimit    ErrorCode = 5
	E006Network      ErrorCode = 6
	E007Config       ErrorCode = 7
	E008Conflict     ErrorCode = 8
	E009Server       ErrorCode = 9
	E010Internal     ErrorCode = 10
)

// String returns the stable public error code (E001–E010).
func (c ErrorCode) String() string {
	switch c {
	case E001Auth:
		return "E001"
	case E002NotFound:
		return "E002"
	case E003Permission:
		return "E003"
	case E004Validation:
		return "E004"
	case E005RateLimit:
		return "E005"
	case E006Network:
		return "E006"
	case E007Config:
		return "E007"
	case E008Conflict:
		return "E008"
	case E009Server:
		return "E009"
	case E010Internal:
		return "E010"
	default:
		if c < 0 {
			return "E000"
		}
		return fmt.Sprintf("E%03d", int(c))
	}
}

// ShortName returns a stable slug used in docs/errors filenames.
func (c ErrorCode) ShortName() string {
	switch c {
	case E001Auth:
		return "auth"
	case E002NotFound:
		return "not-found"
	case E003Permission:
		return "permission"
	case E004Validation:
		return "validation"
	case E005RateLimit:
		return "rate-limit"
	case E006Network:
		return "network"
	case E007Config:
		return "config"
	case E008Conflict:
		return "conflict"
	case E009Server:
		return "server"
	case E010Internal:
		return "internal"
	default:
		return "unknown"
	}
}

// LearnMoreURL returns the documentation link for this error code.
func (c ErrorCode) LearnMoreURL() string {
	return "https://github.com/kim/clockify-cli/blob/main/docs/errors/" + c.String() + "-" + c.ShortName() + ".md"
}

type CLIError struct {
	Code       ErrorCode
	Message    string
	Detail     string
	Suggestion string
	Cause      error
}

func (e *CLIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *CLIError) Unwrap() error {
	return e.Cause
}

func (e *CLIError) ExitCode() int {
	switch e.Code {
	case E001Auth:
		return 2
	case E002NotFound:
		return 3
	case E003Permission:
		return 4
	case E004Validation:
		return 10
	case E005RateLimit:
		return 5
	case E006Network:
		return 6
	case E007Config:
		return 7
	case E008Conflict:
		return 8
	case E009Server:
		return 9
	case E010Internal:
		return 11
	default:
		return 1
	}
}

func FromHTTPStatus(status int) ErrorCode {
	switch {
	case status == 401 || status == 403:
		return E001Auth
	case status == 404:
		return E002NotFound
	case status == 400 || status == 422:
		return E004Validation
	case status == 429:
		return E005RateLimit
	case status == 409:
		return E008Conflict
	case status >= 500:
		return E009Server
	default:
		return E010Internal
	}
}

func New(code ErrorCode, message string) *CLIError {
	return &CLIError{Code: code, Message: message}
}

func Wrap(code ErrorCode, message string, err error) *CLIError {
	return &CLIError{Code: code, Message: message, Cause: err}
}
