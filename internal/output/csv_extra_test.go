package output

import (
	"bytes"
	"strings"
	"testing"
)

type rowStruct struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name,omitempty"`
}

type rowStructNoTags struct {
	ID   string
	Name string
}

func TestCSVFormatterFromStructSlice(t *testing.T) {
	data := []rowStruct{{ID: "1", Name: "a"}, {ID: "2", Name: "b"}}
	var buf bytes.Buffer
	f := &CSVFormatter{}
	if err := f.Format(&buf, data, []string{"ID", "NAME"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID,NAME") || !strings.Contains(out, "1,a") || !strings.Contains(out, "2,b") {
		t.Errorf("csv=%q", out)
	}
}

func TestCSVFormatterFromAnySlice(t *testing.T) {
	data := []any{map[string]any{"name": "x", "value": 1}}
	var buf bytes.Buffer
	f := &CSVFormatter{}
	if err := f.Format(&buf, data, []string{"NAME", "VALUE"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "x,1") {
		t.Errorf("csv=%q", buf.String())
	}
}

func TestCSVFormatterStructNoTags(t *testing.T) {
	data := []rowStructNoTags{{ID: "9", Name: "z"}}
	var buf bytes.Buffer
	f := &CSVFormatter{}
	if err := f.Format(&buf, data, []string{"ID", "NAME"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "9,z") {
		t.Errorf("csv=%q", buf.String())
	}
}

func TestCSVFormatterEmpty(t *testing.T) {
	var buf bytes.Buffer
	f := &CSVFormatter{}
	if err := f.Format(&buf, []map[string]any{}, []string{"A"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "A") {
		t.Errorf("csv=%q", buf.String())
	}
}

func TestToMapsFromStructSlice(t *testing.T) {
	rows := toMaps([]rowStruct{{ID: "1", Name: "n"}})
	if len(rows) != 1 || rows[0]["id"] != "1" || rows[0]["name"] != "n" {
		t.Errorf("rows=%v", rows)
	}
	rows2 := toMaps(map[string]string{"a": "b"})
	if rows2 != nil {
		t.Errorf("non-slice should be nil: %v", rows2)
	}
	var p *[]rowStruct
	if toMaps(p) != nil {
		t.Error("nil ptr slice should be nil")
	}
}

func TestTableFormatterQuiet(t *testing.T) {
	data := []map[string]any{{"name": "x", "value": 1}}
	var buf bytes.Buffer
	f := &TableFormatter{Quiet: true}
	if err := f.Format(&buf, data, []string{"NAME", "VALUE"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "NAME") {
		t.Errorf("quiet table should omit headers: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "x") {
		t.Errorf("table missing row: %q", buf.String())
	}
}

func TestTableFormatterFromStruct(t *testing.T) {
	data := []rowStruct{{ID: "1", Name: "n"}}
	var buf bytes.Buffer
	f := &TableFormatter{}
	if err := f.Format(&buf, data, []string{"ID", "NAME"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "1") || !strings.Contains(buf.String(), "n") {
		t.Errorf("table=%q", buf.String())
	}
}

func TestTerminalHelpersDoNotPanic(t *testing.T) {
	// In CI tests stdout/stdin are usually pipes — just ensure they return bool.
	_ = IsTerminal()
	_ = IsTerminalStdout()
	_ = IsTerminalStdin()
}
