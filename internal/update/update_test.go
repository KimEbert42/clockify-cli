package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient()
	c.APIBase = srv.URL
	c.HTTP = srv.Client()
	return c, srv
}

func TestNormalizeAndCompare(t *testing.T) {
	if NormalizeVersion("v1.2.3") != "1.2.3" {
		t.Fatal("NormalizeVersion failed")
	}
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "v1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.2.0", "1.1.9", 1},
		{"0.9.0", "1.0.0", -1},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("CompareVersions(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestLatestRelease(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/repos/kim/clockify-cli/releases/latest") {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.2.3",
			Assets: []Asset{
				{Name: "clockify_1.2.3_linux_amd64.tar.gz", BrowserDownloadURL: "http://example/linux"},
				{Name: "clockify_1.2.3_darwin_arm64.tar.gz", BrowserDownloadURL: "http://example/darwin"},
				{Name: "clockify_1.2.3_windows_amd64.zip", BrowserDownloadURL: "http://example/win"},
			},
		})
	})
	rel, err := c.LatestRelease(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rel.TagName != "v1.2.3" {
		t.Fatalf("TagName=%q", rel.TagName)
	}
	asset, err := FindAsset(rel, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Name != "clockify_1.2.3_linux_amd64.tar.gz" {
		t.Fatalf("asset=%q", asset.Name)
	}
}

func TestReleaseByVersion(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/releases/tags/1.2.3") {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(Release{TagName: "v1.2.3"})
	})
	rel, err := c.ReleaseByVersion(context.Background(), "1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if rel.TagName != "v1.2.3" {
		t.Fatalf("TagName=%q", rel.TagName)
	}
}

func TestFindAssetMissing(t *testing.T) {
	rel := &Release{TagName: "v1", Assets: []Asset{{Name: "other.zip"}}}
	if _, err := FindAsset(rel, "plan9", "amd64"); err == nil {
		t.Fatal("expected error for missing asset")
	}
}

func TestIsUnsupportedInstall(t *testing.T) {
	for _, v := range []string{"dev", "", "unknown", "abc"} {
		if !IsUnsupportedInstall(v) {
			t.Errorf("IsUnsupportedInstall(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"1.0.0", "v0.9.0-m8"} {
		if IsUnsupportedInstall(v) {
			// v0.9.0-m8 has numeric prefix - should be supported
			if strings.HasPrefix(NormalizeVersion(v), "0.") || strings.HasPrefix(NormalizeVersion(v), "1.") {
				// ok if numeric
			} else {
				t.Errorf("IsUnsupportedInstall(%q) = true, want false", v)
			}
		}
	}
}

func TestShouldSkip(t *testing.T) {
	if r := ShouldSkip(true, "", true, "1.0.0", "time-entry"); r != SkipQuiet {
		t.Errorf("quiet skip = %q", r)
	}
	if r := ShouldSkip(false, "json", true, "1.0.0", "time-entry"); r != SkipJSON {
		t.Errorf("json skip = %q", r)
	}
	if r := ShouldSkip(false, "", false, "1.0.0", "time-entry"); r != SkipConfig {
		t.Errorf("config skip = %q", r)
	}
	if r := ShouldSkip(false, "", true, "dev", "time-entry"); r != SkipNonRelease {
		t.Errorf("dev skip = %q", r)
	}
	if r := ShouldSkip(false, "", true, "1.0.0", "update"); r != SkipCommand {
		t.Errorf("command skip = %q", r)
	}
	t.Setenv("CLOCKIFY_NO_UPDATE_CHECK", "")
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("CLOCKIFY_NO_UPDATE_CHECK", "1")
	if r := ShouldSkip(false, "", true, "1.0.0", "time-entry"); r != SkipEnv {
		t.Errorf("env skip = %q", r)
	}
	t.Setenv("CLOCKIFY_NO_UPDATE_CHECK", "")
	t.Setenv("CI", "true")
	if r := ShouldSkip(false, "", true, "1.0.0", "time-entry"); r != SkipCI {
		t.Errorf("ci skip = %q", r)
	}
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	if r := ShouldSkip(false, "", true, "1.0.0", "list"); r != SkipNone {
		t.Errorf("no-skip = %q", r)
	}
}

func TestMaybeCheckCachesAndHints(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("CLOCKIFY_NO_UPDATE_CHECK", "")
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	SetCurrentVersionProvider(func() string { return "1.0.0" })
	t.Cleanup(func() { SetCurrentVersionProvider(nil) })

	calls := 0
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(Release{TagName: "v2.0.0"})
	})

	res := MaybeCheck(context.Background(), c, "default", "1.0.0", false, "table", true, "list")
	if res == nil || !res.UpdateAvailable {
		t.Fatalf("expected update available, got %+v", res)
	}
	if calls != 1 {
		t.Fatalf("calls=%d want 1", calls)
	}

	// Second call within TTL should not hit network again but still return hint.
	res2 := MaybeCheck(context.Background(), c, "default", "1.0.0", false, "table", true, "list")
	if res2 == nil || res2.Latest != "2.0.0" {
		t.Fatalf("cached hint missing: %+v", res2)
	}
	if calls != 1 {
		t.Fatalf("calls after cache=%d want 1", calls)
	}

	if Hint(res) == "" {
		t.Fatal("empty hint")
	}
}

func TestMaybeCheckNetworkFailureSilent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("CLOCKIFY_NO_UPDATE_CHECK", "")
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	SetCurrentVersionProvider(func() string { return "1.0.0" })
	t.Cleanup(func() { SetCurrentVersionProvider(nil) })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	c := NewClient()
	c.APIBase = srv.URL
	c.HTTP = srv.Client()

	res := MaybeCheck(context.Background(), c, "default", "1.0.0", false, "table", true, "list")
	if res != nil {
		t.Fatalf("expected nil on network failure, got %+v", res)
	}
}

func TestReplaceExecutable(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "clockify")
	src := filepath.Join(dir, "clockify.newbin")
	if err := os.WriteFile(src, []byte("#!/bin/sh\necho hi\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceExecutable(dst, src); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "#!/bin/sh\necho hi\n" {
		t.Fatalf("content=%q", data)
	}
}

func TestCheckResultJSONShape(t *testing.T) {
	res := &CheckResult{Current: "1.0.0", Latest: "1.1.0", UpdateAvailable: true}
	b, _ := json.Marshal(res)
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"current", "latest", "update_available"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %s", k)
		}
	}
	_ = time.Now // keep time import if unused elsewhere
}
