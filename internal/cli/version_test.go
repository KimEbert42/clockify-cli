package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionTextOutput(t *testing.T) {
	SetBuildInfo("1.2.3", "abc1234", "2026-10-04T00:00:00Z")
	t.Cleanup(func() { SetBuildInfo("dev", "none", "unknown") })

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"1.2.3", "abc1234", "2026-10-04", "go", "platform"} {
		if !strings.Contains(out, want) {
			t.Errorf("version output missing %q:\n%s", want, out)
		}
	}
}

func TestVersionJSONOutput(t *testing.T) {
	SetBuildInfo("9.9.9", "deadbeef", "2026-01-01T00:00:00Z")
	t.Cleanup(func() { SetBuildInfo("dev", "none", "unknown") })

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"version", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var info BuildInfo
	if err := json.Unmarshal(buf.Bytes(), &info); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if info.Version != "9.9.9" {
		t.Errorf("Version = %q, want 9.9.9", info.Version)
	}
	if info.Commit != "deadbeef" {
		t.Errorf("Commit = %q, want deadbeef", info.Commit)
	}
	if info.Go == "" || info.Platform == "" {
		t.Errorf("Go/Platform empty: %+v", info)
	}
}

func TestGetBuildInfoDefaults(t *testing.T) {
	SetBuildInfo("dev", "none", "unknown")
	info := GetBuildInfo()
	if info.Version != "dev" {
		t.Errorf("Version = %q, want dev", info.Version)
	}
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Platform = %q, want GOOS/GOARCH", info.Platform)
	}
}
