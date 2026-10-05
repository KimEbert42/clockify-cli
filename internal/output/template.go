package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"
)

// FormatTemplate renders data with a Go text/template.
func FormatTemplate(w io.Writer, data any, tmplStr string) error {
	tmpl, err := template.New("clockify").Funcs(templateFuncs()).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}
	// Ensure trailing newline for CLI friendliness
	return nil
}

// FormatIDs prints only the ID field(s) from data, one per line.
func FormatIDs(w io.Writer, data any) error {
	ids := extractIDs(data)
	for _, id := range ids {
		if _, err := fmt.Fprintln(w, id); err != nil {
			return err
		}
	}
	return nil
}

func extractIDs(data any) []string {
	if data == nil {
		return nil
	}
	v := reflect.ValueOf(data)
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		var out []string
		for i := 0; i < v.Len(); i++ {
			out = append(out, idsFromValue(v.Index(i))...)
		}
		return out
	default:
		return idsFromValue(v)
	}
}

func idsFromValue(v reflect.Value) []string {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if strings.EqualFold(name, "ID") || strings.EqualFold(name, "Id") {
			f := v.Field(i)
			if f.Kind() == reflect.String {
				return []string{f.String()}
			}
		}
	}
	return nil
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}
}
