package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfg, err := Load("default")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Region != "global" {
		t.Errorf("Region = %q, want %q", cfg.Region, "global")
	}
	if cfg.DateFormat != "2006-01-02" {
		t.Errorf("DateFormat = %q, want %q", cfg.DateFormat, "2006-01-02")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: eu-de
  output: json
  date_format: 02/01/2006
`), 0600)

	cfg, err := Load("default")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Region != "eu-de" {
		t.Errorf("Region = %q, want %q", cfg.Region, "eu-de")
	}
	if cfg.Output != "json" {
		t.Errorf("Output = %q, want %q", cfg.Output, "json")
	}
	if cfg.DateFormat != "02/01/2006" {
		t.Errorf("DateFormat = %q, want %q", cfg.DateFormat, "02/01/2006")
	}
}

func TestLoadEnvOverride(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: eu-de
`), 0600)

	os.Setenv("CLOCKIFY_REGION", "us")
	defer os.Unsetenv("CLOCKIFY_REGION")

	cfg, err := Load("default")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Region != "us" {
		t.Errorf("Region = %q, want %q (env should override file)", cfg.Region, "us")
	}
}

func TestLoadProfile(t *testing.T) {
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
    workspace_id: ws-123
`), 0600)

	cfg, err := Load("work")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Region != "eu-de" {
		t.Errorf("Region = %q, want %q", cfg.Region, "eu-de")
	}
	if cfg.WorkspaceID != "ws-123" {
		t.Errorf("WorkspaceID = %q, want %q", cfg.WorkspaceID, "ws-123")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfg := &ProfileConfig{
		Version: 1,
		Default: Config{
			Region: "us",
			Output: "json",
		},
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load("default")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Region != "us" {
		t.Errorf("Region = %q, want %q", loaded.Region, "us")
	}
	if loaded.Output != "json" {
		t.Errorf("Output = %q, want %q", loaded.Output, "json")
	}
}

func TestMergeFlags(t *testing.T) {
	cfg := &Config{Region: "global", Output: "table"}
	merged := MergeFlags(cfg, map[string]interface{}{
		"region": "us",
	})
	if merged.Region != "us" {
		t.Errorf("Region = %q, want %q", merged.Region, "us")
	}
	if merged.Output != "table" {
		t.Errorf("Output = %q, want %q (unchanged)", merged.Output, "table")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Error("ConfigPath() returned empty string")
	}
}
