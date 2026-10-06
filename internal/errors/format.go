package errors

import (
	"errors"
	"fmt"
	"io"
)

func Format(err error) string {
	var cliErr *CLIError
	if As(err, &cliErr) {
		return formatCLIError(cliErr)
	}
	return fmt.Sprintf("ERROR: %v", err)
}

func formatCLIError(err *CLIError) string {
	s := fmt.Sprintf("ERROR [%s]: %s", err.Code, err.Message)
	if err.Detail != "" {
		s += fmt.Sprintf("\n\n  %s", err.Detail)
	}
	if err.Suggestion != "" {
		s += fmt.Sprintf("\n\n  Suggested fix: %s", err.Suggestion)
	}
	if err.Code >= E001Auth && err.Code <= E010Internal {
		s += fmt.Sprintf("\n\n  Learn more: %s", err.Code.LearnMoreURL())
	}
	return s
}

func Fprint(w io.Writer, err error) {
	fmt.Fprintln(w, Format(err))
}

// As finds the first CLIError in err's chain, including wrapped %w causes.
func As(err error, target any) bool {
	if err == nil || target == nil {
		return false
	}
	// Prefer stdlib walk so %w chains are honored.
	switch t := target.(type) {
	case *CLIError:
		var e *CLIError
		if errors.As(err, &e) {
			*t = *e
			return true
		}
	case **CLIError:
		var e *CLIError
		if errors.As(err, &e) {
			*t = e
			return true
		}
	}
	return false
}
