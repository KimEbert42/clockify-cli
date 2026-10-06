package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type CSVFormatter struct{}

func (f *CSVFormatter) Format(w io.Writer, data any, headers []string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if len(headers) > 0 {
		writer.Write(headers)
	}

	rows := toMaps(data)
	for _, row := range rows {
		var record []string
		for _, h := range headers {
			key := strings.ToLower(strings.ReplaceAll(h, " ", "_"))
			record = append(record, fmt.Sprintf("%v", row[key]))
		}
		writer.Write(record)
	}
	return nil
}

func toMaps(data any) []map[string]any {
	switch v := data.(type) {
	case []map[string]any:
		return v
	case []any:
		var result []map[string]any
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				result = append(result, m)
			}
		}
		return result
	default:
		return fromStruct(data)
	}
}

func fromStruct(data any) []map[string]any {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		if m := structToMap(val); m != nil {
			return []map[string]any{m}
		}
		return nil
	}
	if val.Kind() != reflect.Slice {
		return nil
	}
	var result []map[string]any
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i)
		if item.Kind() == reflect.Ptr {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}
		if item.Kind() == reflect.Struct {
			if m := structToMap(item); m != nil {
				result = append(result, m)
			}
		}
	}
	return result
}

func structToMap(item reflect.Value) map[string]any {
	if item.Kind() != reflect.Struct {
		return nil
	}
	m := make(map[string]any)
	t := item.Type()
	for j := 0; j < t.NumField(); j++ {
		field := t.Field(j)
		if field.PkgPath != "" { // unexported
			continue
		}
		tag := field.Tag.Get("yaml")
		if tag == "" || tag == "-" {
			tag = strings.ToLower(field.Name)
		} else {
			// Strip options like ",omitempty" so keys match header names
			if i := strings.IndexByte(tag, ','); i >= 0 {
				tag = tag[:i]
			}
			if tag == "" || tag == "-" {
				tag = strings.ToLower(field.Name)
			}
		}
		m[tag] = item.Field(j).Interface()
	}
	return m
}
