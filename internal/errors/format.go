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

func As(err error, target any) bool {
	if err == nil || target == nil {
		return false
	}
	switch t := target.(type) {
	case *CLIError:
		if e, ok := err.(*CLIError); ok {
			*t = *e
			return true
		}
	case **CLIError:
		if e, ok := err.(*CLIError); ok {
			*t = e
			return true
		}
	}
	return false
}
