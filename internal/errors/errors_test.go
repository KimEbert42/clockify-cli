package errors

import (
	"strings"
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

func TestErrorCodeString(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want string
	}{
		{E001Auth, "E001"},
		{E002NotFound, "E002"},
		{E003Permission, "E003"},
		{E004Validation, "E004"},
		{E005RateLimit, "E005"},
		{E006Network, "E006"},
		{E007Config, "E007"},
		{E008Conflict, "E008"},
		{E009Server, "E009"},
		{E010Internal, "E010"},
		{ErrorCode(42), "E042"},
		{ErrorCode(-1), "E000"},
	}
	for _, tt := range tests {
		if got := tt.code.String(); got != tt.want {
			t.Errorf("String(%d)=%q want %q", tt.code, got, tt.want)
		}
	}
}

func TestLearnMoreURL(t *testing.T) {
	url := E001Auth.LearnMoreURL()
	if !strings.Contains(url, "docs/errors/E001-auth.md") {
		t.Errorf("LearnMoreURL=%q", url)
	}
	if !strings.HasPrefix(url, "https://github.com/kim/clockify-cli/") {
		t.Errorf("LearnMoreURL should be repo-hosted: %q", url)
	}
	if got := E002NotFound.ShortName(); got != "not-found" {
		t.Errorf("ShortName(E002)=%q", got)
	}
	if got := E004Validation.ShortName(); got != "validation" {
		t.Errorf("ShortName(E004)=%q", got)
	}
}

func TestFromHTTPStatus(t *testing.T) {
	tests := []struct {
		status int
		want   ErrorCode
	}{
		{401, E001Auth},
		{403, E003Permission},
		{404, E002NotFound},
		{400, E004Validation},
		{422, E004Validation},
		{429, E005RateLimit},
		{409, E008Conflict},
		{500, E009Server},
		{502, E009Server},
		{418, E010Internal},
	}
	for _, tt := range tests {
		if got := FromHTTPStatus(tt.status); got != tt.want {
			t.Errorf("FromHTTPStatus(%d) = %d, want %d", tt.status, got, tt.want)
		}
	}
}

func TestFormatErrorFull(t *testing.T) {
	err := &CLIError{
		Code:       E001Auth,
		Message:    "authentication failed",
		Detail:     "API key is invalid",
		Suggestion: "Run 'clockify configure' to set up your API key",
	}
	out := Format(err)
	for _, want := range []string{
		"ERROR [E001]: authentication failed",
		"API key is invalid",
		"Suggested fix: Run 'clockify configure'",
		"Learn more:",
		"docs/errors/E001-auth.md",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Format() missing %q in:\n%s", want, out)
		}
	}
}

func TestFormatErrorWithoutOptionalFields(t *testing.T) {
	out := Format(&CLIError{Code: E004Validation, Message: "bad input"})
	if !strings.Contains(out, "ERROR [E004]: bad input") {
		t.Errorf("out=%q", out)
	}
	if strings.Contains(out, "Suggested fix") {
		t.Errorf("unexpected suggestion: %q", out)
	}
	if !strings.Contains(out, "Learn more") {
		t.Errorf("expected learn more: %q", out)
	}
}

func TestFormatPlainError(t *testing.T) {
	out := Format(errString("boom"))
	if out != "ERROR: boom" {
		t.Errorf("out=%q", out)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestFprint(t *testing.T) {
	var sb strings.Builder
	Fprint(&sb, &CLIError{Code: E007Config, Message: "missing config"})
	if !strings.Contains(sb.String(), "E007") {
		t.Errorf("out=%q", sb.String())
	}
}

func TestAs(t *testing.T) {
	var target CLIError
	if !As(&CLIError{Code: E001Auth, Message: "x"}, &target) {
		t.Fatal("As failed")
	}
	if target.Code != E001Auth {
		t.Errorf("code=%v", target.Code)
	}
	if As(errString("x"), &target) {
		t.Error("As should fail for non-CLIError")
	}
	if As(&CLIError{}, nil) {
		t.Error("As with nil target should fail")
	}
}

func TestNewAndWrap(t *testing.T) {
	err := New(E004Validation, "invalid")
	if err.Code != E004Validation || err.Error() != "invalid" {
		t.Fatalf("New=%+v", err)
	}
	cause := errString("root")
	w := Wrap(E006Network, "request failed", cause)
	if w.Unwrap() != error(cause) {
		t.Error("Unwrap mismatch")
	}
	if w.Error() != "request failed: root" {
		t.Errorf("Error()=%q", w.Error())
	}
}

func TestCLIErrorErrorWithCause(t *testing.T) {
	err := &CLIError{Code: E001Auth, Message: "auth failed", Cause: errString("inner")}
	if err.Error() != "auth failed: inner" {
		t.Errorf("Error()=%q", err.Error())
	}
}
