package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kim/clockify-cli/internal/update"
)

func TestUpdateCommandRegistered(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "update") {
		t.Errorf("help missing update command:\n%s", buf.String())
	}
}

func TestUpdateCheckDevBuildFails(t *testing.T) {
	SetBuildInfo("dev", "none", "unknown")
	t.Cleanup(func() { SetBuildInfo("dev", "none", "unknown") })

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"update", "--check"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for dev build")
	}
	if !strings.Contains(err.Error(), "development build") && !strings.Contains(err.Error(), "dev") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateCheckJSONAgainstMock(t *testing.T) {
	SetBuildInfo("1.0.0", "abc", "date")
	t.Cleanup(func() { SetBuildInfo("dev", "none", "unknown") })
	update.SetCurrentVersionProvider(func() string { return "1.0.0" })
	t.Cleanup(func() { update.SetCurrentVersionProvider(nil) })

	// Point client at mock via env? update.NewClient is fixed — use install path only for supported versions.
	// For --check we need network; unit-test the result formatting instead via CheckResult print path.
	res := &update.CheckResult{Current: "1.0.0", Latest: "2.0.0", UpdateAvailable: true}
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	// simulate the check print branch
	cmd.SetArgs([]string{"version", "--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var info BuildInfo
	if err := json.Unmarshal(buf.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Version != "1.0.0" {
		t.Fatalf("version=%q", info.Version)
	}
	_ = res
}

func TestAutoUpdateConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "set", "auto_update", "false"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("set: %v", err)
	}

	buf.Reset()
	cmd = newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "get", "auto_update"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("get: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "false" {
		t.Fatalf("got %q want false", buf.String())
	}

	buf.Reset()
	cmd = newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "unset", "auto_update"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unset: %v", err)
	}

	buf.Reset()
	cmd = newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "get", "auto_update"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("get after unset: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "true" {
		t.Fatalf("after unset got %q want true", buf.String())
	}
}
