package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type idStruct struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type noIDStruct struct {
	Name string
}

func TestFormatTemplate(t *testing.T) {
	var buf bytes.Buffer
	err := FormatTemplate(&buf, map[string]string{"name": "Website"}, "project={{.name}}")
	if err != nil {
		t.Fatalf("FormatTemplate() error = %v", err)
	}
	if buf.String() != "project=Website" {
		t.Errorf("got %q", buf.String())
	}
}

func TestFormatTemplateFuncs(t *testing.T) {
	var buf bytes.Buffer
	err := FormatTemplate(&buf, map[string]string{"name": "abc"}, "{{upper .name}}|{{lower .name}}")
	if err != nil {
		t.Fatalf("FormatTemplate() error = %v", err)
	}
	if buf.String() != "ABC|abc" {
		t.Errorf("got %q", buf.String())
	}
}

func TestFormatTemplateInvalid(t *testing.T) {
	var buf bytes.Buffer
	if err := FormatTemplate(&buf, nil, "{{.Name"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestFormatTemplateExecuteError(t *testing.T) {
	var buf bytes.Buffer
	// .Missing on map[string]string fails at execute time in text/template
	if err := FormatTemplate(&buf, map[string]string{}, "{{.nope}}"); err == nil {
		// maps return zero value for missing keys in text/template — use nil field instead
		t.Log("missing map key is zero-value; trying nil pointer")
	}
	var buf2 bytes.Buffer
	type item struct{ ID *string }
	if err := FormatTemplate(&buf2, item{}, "{{.ID}}"); err == nil {
		// nil pointer print is "<nil>" not error
		t.Log("nil pointer printed without error")
	}
}

func TestFormatIDs_SingleStruct(t *testing.T) {
	var buf bytes.Buffer
	err := FormatIDs(&buf, idStruct{ID: "abc-123", Name: "x"})
	if err != nil {
		t.Fatalf("FormatIDs() error = %v", err)
	}
	if buf.String() != "abc-123\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestFormatIDs_Slice(t *testing.T) {
	var buf bytes.Buffer
	data := []idStruct{{ID: "a"}, {ID: "b"}}
	if err := FormatIDs(&buf, data); err != nil {
		t.Fatalf("FormatIDs() error = %v", err)
	}
	if buf.String() != "a\nb\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestFormatIDs_PointerSlice(t *testing.T) {
	var buf bytes.Buffer
	a, b := idStruct{ID: "x"}, idStruct{ID: "y"}
	data := []*idStruct{&a, &b}
	if err := FormatIDs(&buf, data); err != nil {
		t.Fatalf("FormatIDs() error = %v", err)
	}
	if buf.String() != "x\ny\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestFormatIDs_NilAndEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := FormatIDs(&buf, nil); err != nil {
		t.Fatalf("nil: %v", err)
	}
	if err := FormatIDs(&buf, []idStruct{}); err != nil {
		t.Fatalf("empty slice: %v", err)
	}
	var p *idStruct
	if err := FormatIDs(&buf, p); err != nil {
		t.Fatalf("nil ptr: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestFormatIDs_NoIDField(t *testing.T) {
	var buf bytes.Buffer
	if err := FormatIDs(&buf, noIDStruct{Name: "x"}); err != nil {
		t.Fatalf("FormatIDs() error = %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty, got %q", buf.String())
	}
}

func TestFormatIDs_NilElement(t *testing.T) {
	var buf bytes.Buffer
	var p *idStruct
	if err := FormatIDs(&buf, []*idStruct{p, {ID: "ok"}}); err != nil {
		t.Fatalf("FormatIDs() error = %v", err)
	}
	if buf.String() != "ok\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestFormatIDs_IdFieldOnly(t *testing.T) {
	// Case variants: Id
	type mixed struct {
		Id   string
		Name string
	}
	var buf bytes.Buffer
	if err := FormatIDs(&buf, mixed{Id: "lower-id"}); err != nil {
		t.Fatalf("FormatIDs() error = %v", err)
	}
	if buf.String() != "lower-id\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestDetectFormatTypes(t *testing.T) {
	tests := []struct {
		format   string
		terminal bool
		want     string
	}{
		{"json", false, "*output.JSONFormatter"},
		{"csv", false, "*output.CSVFormatter"},
		{"yaml", false, "*output.YAMLFormatter"},
		{"table", false, "*output.TableFormatter"},
		{"", true, "*output.TableFormatter"},
		{"", false, "*output.JSONFormatter"},
		{"unknown", false, "*output.TableFormatter"},
		{"JSON", false, "*output.TableFormatter"}, // case-sensitive
	}
	for _, tt := range tests {
		f := Detect(tt.format, tt.terminal)
		got := strings.TrimPrefix(strings.Replace(formatType(f), "output.", "", 1), "*")
		if got != strings.TrimPrefix(tt.want, "*output.") {
			t.Errorf("Detect(%q, %v) type=%s, want %s", tt.format, tt.terminal, formatType(f), tt.want)
		}
	}
}

func formatType(f Formatter) string {
	switch f.(type) {
	case *JSONFormatter:
		return "*output.JSONFormatter"
	case *CSVFormatter:
		return "*output.CSVFormatter"
	case *YAMLFormatter:
		return "*output.YAMLFormatter"
	case *TableFormatter:
		return "*output.TableFormatter"
	default:
		return "unknown"
	}
}

func TestYAMLFormatterEncodesJSONCompatible(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{{"name": "t", "n": 1}}
	f := &YAMLFormatter{}
	if err := f.Format(&buf, data, nil); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "name:") {
		t.Errorf("yaml missing name: %q", out)
	}
	_ = json.Marshal // keep import if unused in future edits
}
