package errors

import (
	"testing"
)

func TestExitCodes(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want int
	}{
		{E001Auth, 2},
		{E002NotFound, 3},
		{E003Permission, 4},
		{E004Validation, 10},
		{E005RateLimit, 5},
		{E006Network, 6},
		{E007Config, 7},
		{E008Conflict, 8},
		{E009Server, 9},
		{E010Internal, 11},
	}
	for _, tt := range tests {
		err := &CLIError{Code: tt.code, Message: "test"}
		if got := err.ExitCode(); got != tt.want {
			t.Errorf("ExitCode(%d) = %d, want %d", tt.code, got, tt.want)
		}
	}
}

func TestFromHTTPStatus(t *testing.T) {
	tests := []struct {
		status int
		want   ErrorCode
	}{
		{401, E001Auth},
		{403, E001Auth},
		{404, E002NotFound},
		{400, E004Validation},
		{422, E004Validation},
		{429, E005RateLimit},
		{409, E008Conflict},
		{500, E009Server},
		{502, E009Server},
	}
	for _, tt := range tests {
		if got := FromHTTPStatus(tt.status); got != tt.want {
			t.Errorf("FromHTTPStatus(%d) = %d, want %d", tt.status, got, tt.want)
		}
	}
}

func TestFormatError(t *testing.T) {
	err := &CLIError{
		Code:       E001Auth,
		Message:    "authentication failed",
		Detail:     "API key is invalid",
		Suggestion: "Run 'clockify configure' to set up your API key",
	}
	output := Format(err)
	if output == "" {
		t.Error("Format() returned empty string")
	}
}

func TestCLIErrorError(t *testing.T) {
	err := &CLIError{Code: E001Auth, Message: "auth failed"}
	if err.Error() != "auth failed" {
		t.Errorf("Error() = %q, want %q", err.Error(), "auth failed")
	}
}

func TestCLIErrorUnwrap(t *testing.T) {
	cause := &CLIError{Code: E001Auth, Message: "inner"}
	err := Wrap(E001Auth, "outer", cause)
	if err.Unwrap() != cause {
		t.Error("Unwrap() did not return cause")
	}
}
