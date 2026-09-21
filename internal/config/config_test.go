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

func TestLoadMultipleProfiles(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
  output: table
profiles:
  work:
    region: eu-de
    workspace_id: ws-work
    output: json
  personal:
    region: us
    workspace_id: ws-personal
`), 0600)

	workCfg, err := Load("work")
	if err != nil {
		t.Fatalf("Load('work') error = %v", err)
	}
	if workCfg.Region != "eu-de" {
		t.Errorf("work.Region = %q, want %q", workCfg.Region, "eu-de")
	}
	if workCfg.WorkspaceID != "ws-work" {
		t.Errorf("work.WorkspaceID = %q, want %q", workCfg.WorkspaceID, "ws-work")
	}
	if workCfg.Output != "json" {
		t.Errorf("work.Output = %q, want %q", workCfg.Output, "json")
	}

	personalCfg, err := Load("personal")
	if err != nil {
		t.Fatalf("Load('personal') error = %v", err)
	}
	if personalCfg.Region != "us" {
		t.Errorf("personal.Region = %q, want %q", personalCfg.Region, "us")
	}
	if personalCfg.WorkspaceID != "ws-personal" {
		t.Errorf("personal.WorkspaceID = %q, want %q", personalCfg.WorkspaceID, "ws-personal")
	}
	if personalCfg.Output != "table" {
		t.Errorf("personal.Output should inherit default 'table', got %q", personalCfg.Output)
	}

	defaultCfg, err := Load("default")
	if err != nil {
		t.Fatalf("Load('default') error = %v", err)
	}
	if defaultCfg.Region != "global" {
		t.Errorf("default.Region = %q, want %q", defaultCfg.Region, "global")
	}
	if defaultCfg.WorkspaceID != "" {
		t.Errorf("default.WorkspaceID should be empty, got %q", defaultCfg.WorkspaceID)
	}
}

func TestLoadNonexistentProfile(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: eu-de
profiles:
  work:
    region: us
`), 0600)

	cfg, err := Load("nonexistent")
	if err != nil {
		t.Fatalf("Load('nonexistent') error = %v", err)
	}
	if cfg.Region != "eu-de" {
		t.Errorf("nonexistent profile should fall back to default, Region = %q, want %q", cfg.Region, "eu-de")
	}
}

func TestSaveAndLoadWithProfiles(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	original := &ProfileConfig{
		Version: 1,
		Default: Config{Region: "global", Output: "table"},
		Profiles: map[string]Config{
			"work":     {Region: "eu-de", WorkspaceID: "ws-work", Output: "json"},
			"personal": {Region: "us", WorkspaceID: "ws-personal"},
		},
	}
	if err := Save(original); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	workCfg, err := Load("work")
	if err != nil {
		t.Fatalf("Load('work') error = %v", err)
	}
	if workCfg.Region != "eu-de" || workCfg.WorkspaceID != "ws-work" {
		t.Errorf("work profile not preserved: got region=%q workspace=%q", workCfg.Region, workCfg.WorkspaceID)
	}

	personalCfg, err := Load("personal")
	if err != nil {
		t.Fatalf("Load('personal') error = %v", err)
	}
	if personalCfg.Region != "us" || personalCfg.WorkspaceID != "ws-personal" {
		t.Errorf("personal profile not preserved: got region=%q workspace=%q", personalCfg.Region, personalCfg.WorkspaceID)
	}

	defaultCfg, err := Load("default")
	if err != nil {
		t.Fatalf("Load('default') error = %v", err)
	}
	if defaultCfg.Region != "global" {
		t.Errorf("default profile not preserved: got region=%q", defaultCfg.Region)
	}
}

func TestDeleteProfile(t *testing.T) {
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
  personal:
    region: us
`), 0600)

	if err := DeleteProfile("work"); err != nil {
		t.Fatalf("DeleteProfile('work') error = %v", err)
	}

	names, err := LoadProfileNames()
	if err != nil {
		t.Fatalf("LoadProfileNames() error = %v", err)
	}
	for _, n := range names {
		if n == "work" {
			t.Error("profile 'work' should have been deleted but still exists")
		}
	}

	personalCfg, err := Load("personal")
	if err != nil {
		t.Fatalf("Load('personal') error = %v", err)
	}
	if personalCfg.Region != "us" {
		t.Errorf("personal profile affected by work deletion: Region = %q, want %q", personalCfg.Region, "us")
	}
}

func TestDeleteDefaultProfileRejected(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
`), 0600)

	err := DeleteProfile("default")
	if err == nil {
		t.Error("DeleteProfile('default') should return error")
	}
}

func TestDeleteNonexistentProfile(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: global
`), 0600)

	err := DeleteProfile("nonexistent")
	if err == nil {
		t.Error("DeleteProfile('nonexistent') should return error")
	}
}

func TestLoadProfileNames(t *testing.T) {
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
  personal:
    region: us
`), 0600)

	names, err := LoadProfileNames()
	if err != nil {
		t.Fatalf("LoadProfileNames() error = %v", err)
	}
	if len(names) != 3 {
		t.Fatalf("expected 3 profiles (default + 2), got %d: %v", len(names), names)
	}
	hasDefault := false
	hasWork := false
	hasPersonal := false
	for _, n := range names {
		switch n {
		case "default":
			hasDefault = true
		case "work":
			hasWork = true
		case "personal":
			hasPersonal = true
		}
	}
	if !hasDefault || !hasWork || !hasPersonal {
		t.Errorf("missing expected profiles: default=%v work=%v personal=%v (got %v)", hasDefault, hasWork, hasPersonal, names)
	}
}

func TestLoadRaw(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfgPath := filepath.Join(dir, "clockify", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0700)
	os.WriteFile(cfgPath, []byte(`version: 1
default:
  region: eu-de
profiles:
  work:
    region: us
`), 0600)

	pc, err := LoadRaw()
	if err != nil {
		t.Fatalf("LoadRaw() error = %v", err)
	}
	if pc.Version != 1 {
		t.Errorf("Version = %d, want 1", pc.Version)
	}
	if pc.Default.Region != "eu-de" {
		t.Errorf("Default.Region = %q, want %q", pc.Default.Region, "eu-de")
	}
	if len(pc.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(pc.Profiles))
	}
	if pc.Profiles["work"].Region != "us" {
		t.Errorf("Profiles[work].Region = %q, want %q", pc.Profiles["work"].Region, "us")
	}
}
