package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type TableFormatter struct {
	Quiet bool
}

func (f *TableFormatter) Format(w io.Writer, data any, headers []string) error {
	rows := toMaps(data)
	if len(rows) == 0 {
		fmt.Fprintln(w, "No results found")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if !f.Quiet && len(headers) > 0 {
		fmt.Fprintln(tw, strings.Join(headers, "\t"))
	}

	for _, row := range rows {
		var vals []string
		for _, h := range headers {
			key := strings.ToLower(strings.ReplaceAll(h, " ", "_"))
			val := fmt.Sprintf("%v", row[key])
			vals = append(vals, val)
		}
		fmt.Fprintln(tw, strings.Join(vals, "\t"))
	}

	return tw.Flush()
}
