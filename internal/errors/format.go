package errors

import (
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
	s := fmt.Sprintf("ERROR [%d]: %s", int(err.Code), err.Message)
	if err.Detail != "" {
		s += fmt.Sprintf("\n\n  %s", err.Detail)
	}
	if err.Suggestion != "" {
		s += fmt.Sprintf("\n\n  Suggested fix: %s", err.Suggestion)
	}
	return s
}

func Fprint(w io.Writer, err error) {
	fmt.Fprintln(w, Format(err))
}

func As(err error, target any) bool {
	if target == nil {
		return false
	}
	if cliErr, ok := target.(*CLIError); ok {
		if e, ok := err.(*CLIError); ok {
			*cliErr = *e
			return true
		}
	}
	return false
}
