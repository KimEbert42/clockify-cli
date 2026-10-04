package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kim/clockify-cli/internal/config"
	"github.com/spf13/cobra"
)

func TestCompletionGenerateAllShells(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		buf := new(strings.Builder)
		cmd := newRootCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"completion", shell})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("completion %s error = %v", shell, err)
		}
		if buf.Len() == 0 {
			t.Errorf("completion %s produced empty output", shell)
		}
	}
}

func TestCompletionBackcompatPositional(t *testing.T) {
	buf := new(strings.Builder)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion", "bash"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("completion bash error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("positional completion bash produced empty output")
	}
}

func TestCompletionInstallWritesFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, "xdg-data"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg-config"))
	t.Setenv("USERPROFILE", home)

	buf := new(strings.Builder)
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion", "install", "--shell", "zsh", "--quiet"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("install error = %v", err)
	}

	path := filepath.Join(home, "xdg-data", "zsh", "site-functions", "_clockify")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected installed file at %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Error("installed completion file is empty")
	}

	// Second install without --force should fail
	cmd2 := newRootCmd()
	cmd2.SetOut(new(strings.Builder))
	cmd2.SetErr(new(strings.Builder))
	cmd2.SetArgs([]string{"completion", "install", "--shell", "zsh"})
	if err := cmd2.Execute(); err == nil {
		t.Error("expected error when overwriting without --force")
	}
}

func TestCompletionInstallBashPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BASH_COMP_INSTALL_DIR", dir)
	t.Setenv("HOME", t.TempDir())

	cmd := newRootCmd()
	cmd.SetOut(new(strings.Builder))
	cmd.SetErr(new(strings.Builder))
	cmd.SetArgs([]string{"completion", "install", "--shell", "bash", "--force", "--quiet"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("install bash error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "clockify")); err != nil {
		t.Errorf("expected bash completion at %s: %v", filepath.Join(dir, "clockify"), err)
	}
}

func TestTTLCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))

	type payload struct {
		Items []string `json:"items"`
	}
	in := payload{Items: []string{"a", "b"}}

	profile := "default"
	if err := config.WriteTTL(profile, "test.json", in); err != nil {
		t.Fatalf("WriteTTL error = %v", err)
	}

	var out payload
	ok, err := config.ReadTTL(profile, "test.json", &out, config.DefaultListCacheTTL)
	if err != nil || !ok {
		t.Fatalf("ReadTTL ok=%v err=%v", ok, err)
	}
	if len(out.Items) != 2 || out.Items[0] != "a" {
		t.Errorf("unexpected payload: %+v", out)
	}

	// Stale entry
	if err := config.WriteTTL(profile, "stale.json", in); err != nil {
		t.Fatal(err)
	}
	// Rewrite with zero time by writing raw envelope
	raw := []byte(`{"fetched_at":"2020-01-01T00:00:00Z","data":{"items":["old"]}}`)
	if err := config.WriteCache(profile, "stale.json", raw); err != nil {
		t.Fatal(err)
	}
	var stale payload
	ok, err = config.ReadTTL(profile, "stale.json", &stale, config.DefaultListCacheTTL)
	if err != nil || ok {
		t.Errorf("stale cache should not be fresh: ok=%v err=%v", ok, err)
	}
}

func TestCachedCompletionNamesFiltered(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	cmd := newRootCmd()
	items := []completionItem{
		{ID: "1", Name: "Website"},
		{ID: "2", Name: "Mobile App"},
		{ID: "3", Name: "Website Redesign"},
	}
	rememberCompletionNames(cmd, "ws1", compResourceProjects, items)

	got, directive := completeCachedNames(cmd, "ws1", compResourceProjects, "Web")
	if directive == cobra.ShellCompDirectiveError || len(got) != 2 {
		t.Fatalf("expected 2 matches for Web, got %v (dir=%v)", got, directive)
	}
	if got[0] != "Website" || got[1] != "Website Redesign" {
		t.Errorf("unexpected matches: %v", got)
	}

	// Prefix match is case-insensitive
	got, _ = completeCachedNames(cmd, "ws1", compResourceProjects, "web")
	if len(got) != 2 {
		t.Errorf("case-insensitive prefix failed: %v", got)
	}

	// No cache miss should not call API / should error directive
	miss, missDir := completeCachedNames(cmd, "ws1", compResourceTags, "x")
	if miss != nil || missDir != cobra.ShellCompDirectiveError {
		t.Errorf("expected empty+error for missing cache, got %v dir=%v", miss, missDir)
	}
}

func TestScopedCacheKey(t *testing.T) {
	if got := scopedCacheKey("", "projects"); got != "projects.json" {
		t.Errorf("got %q", got)
	}
	if got := scopedCacheKey("ws1", "projects"); got != "projects.ws1.json" {
		t.Errorf("got %q", got)
	}
}

func TestCacheDisabledFromCmd(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--no-cache", "version"})
	// Execute to bind persistent flags onto the parsed command tree is complex;
	// test helper directly with a command that has the flag.
	root := newRootCmd()
	root.SetArgs([]string{"version"})
	_ = root.Execute()
	// Simulate flag lookup path used by tests that set flags manually.
	if cacheDisabledFromCmd(root) {
		t.Error("expected no-cache false by default")
	}
	root.PersistentFlags().Set("no-cache", "true")
	// Persistent flags on root may not be visible after execute on child;
	// verify via a fresh command with flag set.
	fresh := newRootCmd()
	if err := fresh.PersistentFlags().Set("no-cache", "true"); err != nil {
		t.Fatal(err)
	}
	if !cacheDisabledFromCmd(fresh) {
		t.Error("expected no-cache true after set")
	}
	_ = cmd
}
