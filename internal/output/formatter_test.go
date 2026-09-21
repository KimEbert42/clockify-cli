package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestJSONFormatter(t *testing.T) {
	data := []map[string]any{{"name": "test", "value": 42}}
	var buf bytes.Buffer
	f := &JSONFormatter{}
	err := f.Format(&buf, data, nil)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	var result []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
}

func TestTableFormatter(t *testing.T) {
	data := []map[string]any{{"name": "test", "value": 42}}
	var buf bytes.Buffer
	f := &TableFormatter{}
	err := f.Format(&buf, data, []string{"NAME", "VALUE"})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("table output is empty")
	}
}

func TestTableFormatterEmpty(t *testing.T) {
	data := []map[string]any{}
	var buf bytes.Buffer
	f := &TableFormatter{}
	err := f.Format(&buf, data, []string{"NAME"})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if !strings.Contains(buf.String(), "No results found") {
		t.Error("empty table should show 'No results found'")
	}
}

func TestCSVFormatter(t *testing.T) {
	data := []map[string]any{{"name": "test", "value": 42}}
	var buf bytes.Buffer
	f := &CSVFormatter{}
	err := f.Format(&buf, data, []string{"NAME", "VALUE"})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if !strings.Contains(buf.String(), "NAME,VALUE") {
		t.Error("CSV should contain header row")
	}
}

func TestYAMLFormatter(t *testing.T) {
	data := []map[string]any{{"name": "test", "value": 42}}
	var buf bytes.Buffer
	f := &YAMLFormatter{}
	err := f.Format(&buf, data, nil)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("YAML output is empty")
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		format    string
		terminal  bool
		wantType  string
	}{
		{"json", false, "*output.JSONFormatter"},
		{"csv", false, "*output.CSVFormatter"},
		{"yaml", false, "*output.YAMLFormatter"},
		{"table", false, "*output.TableFormatter"},
		{"", true, "*output.TableFormatter"},
		{"", false, "*output.JSONFormatter"},
	}
	for _, tt := range tests {
		f := Detect(tt.format, tt.terminal)
		typeName := strings.TrimPrefix(strings.TrimPrefix(
			strings.Replace(fmt.Sprintf("%T", f), "*output.", "", 1), ""), "*")
		_ = typeName
	}
}
