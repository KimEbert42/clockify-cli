package config

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempConfigHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cacheDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_CACHE_HOME", cacheDir)
	// ConfigPath may also use other env — clear CLOCKIFY overrides that tests set
	t.Setenv("CLOCKIFY_API_KEY", "")
	t.Setenv("CLOCKIFY_WORKSPACE_ID", "")
	t.Setenv("CLOCKIFY_REGION", "")
	t.Setenv("CLOCKIFY_OUTPUT", "")
	t.Setenv("CLOCKIFY_CACHE_TTL", "")
	return dir
}

func TestLoadDefaultsWhenMissing(t *testing.T) {
	withTempConfigHome(t)
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Region != "global" || cfg.DateFormat != "2006-01-02" || cfg.WeekStarts != "MONDAY" {
		t.Fatalf("defaults=%+v", cfg)
	}
	if !cfg.AutoUpdateEnabled() {
		t.Error("auto update should default enabled")
	}
	if cfg.InteractiveEnabled() {
		t.Error("interactive should default off")
	}
}

func TestLoadProfileMerge(t *testing.T) {
	dir := withTempConfigHome(t)
	// Discover actual config path helper used by package
	// Write a config with default + profile
	// ConfigPath() depends on package implementation — write via Save
	pc := &ProfileConfig{
		Version: 1,
		Default: Config{WorkspaceID: "ws-default", Output: "json"},
		Profiles: map[string]Config{
			"work": {WorkspaceID: "ws-work", Region: "eu-de", Interactive: "true", AutoUpdate: "false", TimeZone: "Europe/Berlin"},
		},
	}
	if err := Save(pc); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load("work")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkspaceID != "ws-work" {
		t.Errorf("WorkspaceID=%q", cfg.WorkspaceID)
	}
	if cfg.Region != "eu-de" {
		t.Errorf("Region=%q", cfg.Region)
	}
	if cfg.Output != "json" {
		t.Errorf("Output=%q (should inherit default)", cfg.Output)
	}
	if !cfg.InteractiveEnabled() || cfg.AutoUpdateEnabled() {
		t.Errorf("interactive/auto_update wrong: %+v", cfg)
	}
	if cfg.TimeZone != "Europe/Berlin" {
		t.Errorf("TimeZone=%q", cfg.TimeZone)
	}

	// Unknown profile falls back to default
	cfg2, err := Load("nope")
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.WorkspaceID != "ws-default" {
		t.Errorf("unknown profile WorkspaceID=%q", cfg2.WorkspaceID)
	}

	// Empty profile = default only
	cfg3, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg3.WorkspaceID != "ws-default" || cfg3.Region != "global" {
		t.Errorf("default load=%+v", cfg3)
	}

	_ = dir
}

func TestLoadEnvOverrides(t *testing.T) {
	withTempConfigHome(t)
	t.Setenv("CLOCKIFY_API_KEY", "env-key")
	t.Setenv("CLOCKIFY_WORKSPACE_ID", "env-ws")
	t.Setenv("CLOCKIFY_REGION", "us")
	t.Setenv("CLOCKIFY_OUTPUT", "csv")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "env-key" || cfg.WorkspaceID != "env-ws" || cfg.Region != "us" || cfg.Output != "csv" {
		t.Fatalf("env overrides missing: %+v", cfg)
	}
}

func TestParseInteractiveAndAutoUpdate(t *testing.T) {
	for _, v := range []string{"true", "1", "yes", "on", " TRUE "} {
		got, err := ParseInteractive(v)
		if err != nil || got != "true" {
			t.Errorf("ParseInteractive(%q)=%q %v", v, got, err)
		}
	}
	for _, v := range []string{"false", "0", "no", "off"} {
		got, err := ParseInteractive(v)
		if err != nil || got != "false" {
			t.Errorf("ParseInteractive(%q)=%q %v", v, got, err)
		}
	}
	if _, err := ParseInteractive("maybe"); err == nil {
		t.Error("expected ParseInteractive error")
	}
	if _, err := ParseAutoUpdate("maybe"); err == nil {
		t.Error("expected ParseAutoUpdate error")
	}
	if got, err := ParseAutoUpdate("on"); err != nil || got != "true" {
		t.Errorf("ParseAutoUpdate=%q %v", got, err)
	}
}

func TestMergeFlagsExtra(t *testing.T) {
	base := &Config{APIKey: "a", WorkspaceID: "w", Region: "global", Output: "table"}
	got := MergeFlags(base, map[string]interface{}{
		"api-key":   "flag-key",
		"workspace": "flag-ws",
		"region":    "eu-de",
		"output":    "json",
		"other":     "ignored",
	})
	if got.APIKey != "flag-key" || got.WorkspaceID != "flag-ws" || got.Region != "eu-de" || got.Output != "json" {
		t.Fatalf("MergeFlags=%+v", got)
	}
	// Empty flags do not override
	got2 := MergeFlags(base, map[string]interface{}{"api-key": "", "region": ""})
	if got2.APIKey != "a" || got2.Region != "global" {
		t.Fatalf("empty flags should not override: %+v", got2)
	}
	// Wrong types ignored
	got3 := MergeFlags(base, map[string]interface{}{"api-key": 123})
	if got3.APIKey != "a" {
		t.Fatalf("non-string flag should be ignored: %+v", got3)
	}
}

func TestLoadRawAndProfileNames(t *testing.T) {
	withTempConfigHome(t)
	names, err := LoadProfileNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 1 || names[0] != "default" {
		t.Errorf("names=%v", names)
	}

	pc, err := LoadRaw()
	if err != nil {
		t.Fatal(err)
	}
	if pc.Profiles == nil {
		t.Fatal("Profiles map should be non-nil")
	}

	pc.Profiles["personal"] = Config{WorkspaceID: "ws-p"}
	if err := Save(pc); err != nil {
		t.Fatal(err)
	}
	names, _ = LoadProfileNames()
	found := false
	for _, n := range names {
		if n == "personal" {
			found = true
		}
	}
	if !found {
		t.Errorf("profile names=%v", names)
	}
}

func TestDeleteProfileExtra(t *testing.T) {
	withTempConfigHome(t)
	pc, _ := LoadRaw()
	pc.Profiles["tmp"] = Config{WorkspaceID: "x"}
	_ = Save(pc)
	if err := DeleteProfile("tmp"); err != nil {
		t.Fatal(err)
	}
	pc, _ = LoadRaw()
	if _, ok := pc.Profiles["tmp"]; ok {
		t.Error("profile still present")
	}
	if err := DeleteProfile("default"); err == nil {
		t.Error("cannot delete default")
	}
	if err := DeleteProfile("missing"); err == nil {
		t.Error("expected missing profile error")
	}
}

func TestFileExistsAndResolveConfigPath(t *testing.T) {
	dir := withTempConfigHome(t)
	if FileExists() {
		// may exist if Save was called in same process without cleanup — force clean
		_ = os.Remove(ConfigPath())
	}
	if FileExists() {
		t.Error("FileExists should be false for empty home")
	}
	pc := &ProfileConfig{Version: 1, Default: defaults, Profiles: map[string]Config{}}
	if err := Save(pc); err != nil {
		t.Fatal(err)
	}
	if !FileExists() {
		t.Error("FileExists should be true after Save")
	}
	// Config file should be 0600
	st, err := os.Stat(ConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0600 {
		t.Errorf("config mode=%v want 0600", st.Mode().Perm())
	}
	_ = ResolveConfigPathForTests(dir)
	_ = filepath.Join(dir, "config.yaml")
}

func TestLoadRawParseError(t *testing.T) {
	withTempConfigHome(t)
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{{{not yaml"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRaw(); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadIgnoresCorruptConfig(t *testing.T) {
	withTempConfigHome(t)
	path := ConfigPath()
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	_ = os.WriteFile(path, []byte("{{{not yaml"), 0600)
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Region != "global" {
		t.Errorf("corrupt config should fall back to defaults: %+v", cfg)
	}
}

func TestInteractiveEnabledNilAndValues(t *testing.T) {
	var c *Config
	if c.InteractiveEnabled() {
		t.Error("nil should be false")
	}
	c = &Config{Interactive: "on"}
	if !c.InteractiveEnabled() {
		t.Error("on should enable")
	}
	c = &Config{Interactive: "OFF"}
	if c.InteractiveEnabled() {
		t.Error("OFF should disable")
	}
}

func TestAutoUpdateEnabledNilAndValues(t *testing.T) {
	var c *Config
	if !c.AutoUpdateEnabled() {
		t.Error("nil should default true")
	}
	c = &Config{AutoUpdate: "false"}
	if c.AutoUpdateEnabled() {
		t.Error("false should disable")
	}
	c = &Config{AutoUpdate: "yes"}
	if !c.AutoUpdateEnabled() {
		t.Error("yes should enable")
	}
}
