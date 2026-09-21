package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("help output is empty")
	}
}

func TestVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("version output is empty")
	}
}

func TestConfigPath(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "path"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("config path output is empty")
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"not configured", "(not configured)", "(not configured)"},
		{"short key", "abc123", "****"},
		{"long key", "ck_live_abcdefghijklmnopqrstuvwxyz", "ck_l****wxyz"},
		{"exactly 8", "12345678", "****"},
		{"9 chars", "123456789", "1234****6789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskAPIKey(tt.input)
			if got != tt.want {
				t.Errorf("maskAPIKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatValidIndicator(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"valid", "valid"},
		{"invalid", "invalid - run 'clockify configure' to re-authenticate"},
		{"unreachable", "unable to validate (offline?)"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := formatValidIndicator(tt.status)
			if got != tt.want {
				t.Errorf("formatValidIndicator(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestProfileOrDefault(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "default"},
		{"default", "default"},
		{"work", "work"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := profileOrDefault(tt.input)
			if got != tt.want {
				t.Errorf("profileOrDefault(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStatusNoAPIKey(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
`), 0600)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "(not configured)") {
		t.Errorf("status output should show '(not configured)' for missing API key, got:\n%s", output)
	}
	if !strings.Contains(output, "unknown") {
		t.Errorf("status output should show 'unknown' validation status, got:\n%s", output)
	}
}

func TestStatusJSONOutput(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
  workspace_id: ws-test
`), 0600)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"status", "--output", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var out statusOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if out.Profile != "default" {
		t.Errorf("profile = %q, want %q", out.Profile, "default")
	}
	if out.Region != "global" {
		t.Errorf("region = %q, want %q", out.Region, "global")
	}
	if out.Workspace != "ws-test" {
		t.Errorf("workspace = %q, want %q", out.Workspace, "ws-test")
	}
	if out.Valid != "unknown" {
		t.Errorf("valid = %q, want %q", out.Valid, "unknown")
	}
}

func TestConfigListProfiles(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
profiles:
  work:
    region: eu-de
    workspace_id: ws-work
  personal:
    region: us
`), 0600)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[default]") {
		t.Errorf("config list should show [default] section, got:\n%s", output)
	}
	if !strings.Contains(output, "[work]") {
		t.Errorf("config list should show [work] section, got:\n%s", output)
	}
	if !strings.Contains(output, "[personal]") {
		t.Errorf("config list should show [personal] section, got:\n%s", output)
	}
}

func TestConfigDeleteProfile(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
profiles:
  work:
    region: eu-de
`), 0600)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "delete", "--profile", "work", "--confirm"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("delete output should confirm deletion, got:\n%s", output)
	}

	buf.Reset()
	cmd2 := newRootCmd()
	cmd2.SetOut(buf)
	cmd2.SetErr(buf)
	cmd2.SetArgs([]string{"config", "list"})
	cmd2.Execute()
	listOutput := buf.String()
	if strings.Contains(listOutput, "[work]") {
		t.Errorf("config list should not show [work] after deletion, got:\n%s", listOutput)
	}
}

func TestConfigSetPreservesProfiles(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Ensure env vars don't interfere
	for _, k := range []string{"CLOCKIFY_REGION", "CLOCKIFY_OUTPUT", "CLOCKIFY_WORKSPACE_ID", "CLOCKIFY_API_KEY"} {
		if v := os.Getenv(k); v != "" {
			os.Unsetenv(k)
			defer os.Setenv(k, v)
		}
	}

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
profiles:
  work:
    region: eu-de
    workspace_id: ws-work
`), 0600)

	buf := new(bytes.Buffer)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"config", "set", "region", "us", "--profile", "work"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	buf.Reset()
	cmd2 := newRootCmd()
	cmd2.SetOut(buf)
	cmd2.SetErr(buf)
	cmd2.SetArgs([]string{"config", "get", "region", "--profile", "work"})
	cmd2.Execute()
	output := buf.String()
	if !strings.Contains(output, "us") {
		t.Errorf("work profile region should be 'us', got:\n%s", output)
	}

	buf.Reset()
	cmd3 := newRootCmd()
	cmd3.SetOut(buf)
	cmd3.SetErr(buf)
	cmd3.SetArgs([]string{"config", "get", "region"})
	cmd3.Execute()
	output = buf.String()
	if !strings.Contains(output, "global") {
		t.Errorf("default profile region should still be 'global', got:\n%s", output)
	}
}
