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
