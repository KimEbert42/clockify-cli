package output

import (
	"io"
)

type Formatter interface {
	Format(w io.Writer, data any, headers []string) error
}

func Detect(format string, isTerminal bool) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "csv":
		return &CSVFormatter{}
	case "yaml":
		return &YAMLFormatter{}
	case "table":
		return &TableFormatter{}
	case "":
		if isTerminal {
			return &TableFormatter{}
		}
		return &JSONFormatter{}
	default:
		return &TableFormatter{}
	}
}
