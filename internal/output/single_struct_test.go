package output

import (
	"bytes"
	"strings"
	"testing"
)

type singleEntity struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name,omitempty"`
}

func TestCSVFormatterSingleStruct(t *testing.T) {
	data := singleEntity{ID: "p1", Name: "Website"}
	var buf bytes.Buffer
	f := &CSVFormatter{}
	if err := f.Format(&buf, data, []string{"ID", "NAME"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID,NAME") || !strings.Contains(out, "p1,Website") {
		t.Errorf("csv=%q", out)
	}
}

func TestCSVFormatterSingleStructPointer(t *testing.T) {
	data := &singleEntity{ID: "p2", Name: "App"}
	var buf bytes.Buffer
	f := &CSVFormatter{}
	if err := f.Format(&buf, data, []string{"ID", "NAME"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "p2,App") {
		t.Errorf("csv=%q", buf.String())
	}
}

func TestTableFormatterSingleStruct(t *testing.T) {
	data := singleEntity{ID: "e1", Name: "Timer"}
	var buf bytes.Buffer
	f := &TableFormatter{}
	if err := f.Format(&buf, data, []string{"ID", "NAME"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "e1") || !strings.Contains(out, "Timer") {
		t.Errorf("table=%q", out)
	}
	if strings.Contains(out, "No results found") {
		t.Errorf("single struct should not be empty: %q", out)
	}
}

func TestFromMapsNilAndEmpty(t *testing.T) {
	if toMaps(nil) != nil {
		t.Error("nil should be nil")
	}
	var p *singleEntity
	if toMaps(p) != nil {
		t.Error("nil struct ptr should be nil")
	}
}
