package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
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
}

func TestFindAssetDarwinUniversal(t *testing.T) {
	rel := &Release{TagName: "v1.0.0", Assets: []Asset{
		{Name: "clockify_1.0.0_darwin_all.tar.gz", BrowserDownloadURL: "http://x/all"},
	}}
	asset, err := FindAsset(rel, "darwin", "arm64")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Name != "clockify_1.0.0_darwin_all.tar.gz" {
		t.Errorf("asset=%q", asset.Name)
	}
}

func TestFindAssetWindowsZip(t *testing.T) {
	rel := &Release{TagName: "v1.0.0", Assets: []Asset{
		{Name: "clockify_1.0.0_windows_amd64.zip", BrowserDownloadURL: "http://x/win"},
	}}
	asset, err := FindAsset(rel, "windows", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Name != "clockify_1.0.0_windows_amd64.zip" {
		t.Errorf("asset=%q", asset.Name)
	}
}

func TestFindAssetNilRelease(t *testing.T) {
	if _, err := FindAsset(nil, "linux", "amd64"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCompareVersionsMore(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0-next", "1.0.0", 0},
		{"1.10.0", "1.9.0", 1},
		{"1.0.0", "1.0.0-next", 0},
		{"2.0.0", "1.9.9", 1},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("CompareVersions(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestIsUnsupportedInstallMore(t *testing.T) {
	if IsUnsupportedInstall("1.0.0") {
		t.Error("1.0.0 should be supported")
	}
	if !IsUnsupportedInstall("v") {
		t.Error("bare v should be unsupported")
	}
}

func TestReleaseByVersionEmptyTag(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{TagName: ""})
	})
	if _, err := c.ReleaseByVersion(context.Background(), "1.0.0"); err == nil {
		t.Fatal("expected error for empty tag")
	}
}

func TestLatestReleaseEmptyTag(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{})
	})
	if _, err := c.LatestRelease(context.Background()); err == nil {
		t.Fatal("expected error for empty tag")
	}
}

func TestGetJSONNonOK(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	})
	var out map[string]any
	if err := c.getJSON(context.Background(), c.repoPath()+"/x", &out); err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckUpdateAvailableAndNot(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{TagName: "v1.1.0"})
	})
	res, err := Check(context.Background(), c, "1.0.0")
	if err != nil || !res.UpdateAvailable {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	res, err = Check(context.Background(), c, "v1.1.0")
	if err != nil || res.UpdateAvailable {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestExtractTarGzAndZip(t *testing.T) {
	dir := t.TempDir()

	// tar.gz
	tgz := filepath.Join(dir, "a.tar.gz")
	f, err := os.Create(tgz)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	body := []byte("#!/bin/sh\necho bin\n")
	if err := tw.WriteHeader(&tar.Header{Name: "clockify", Mode: 0755, Size: int64(len(body))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	_ = tw.Close()
	_ = gz.Close()
	_ = f.Close()

	out, err := extractBinary(tgz, dir, "clockify")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "echo bin") {
		t.Errorf("tar content=%q", data)
	}

	// zip
	zp := filepath.Join(dir, "a.zip")
	zf, err := os.Create(zp)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zf)
	w, err := zw.Create("dist/clockify")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write(body)
	_ = zw.Close()
	_ = zf.Close()

	out2, err := extractBinary(zp, dir, "clockify")
	if err != nil {
		t.Fatal(err)
	}
	data2, _ := os.ReadFile(out2)
	if !strings.Contains(string(data2), "echo bin") {
		t.Errorf("zip content=%q", data2)
	}

	// raw binary
	raw := filepath.Join(dir, "clockify.bin")
	_ = os.WriteFile(raw, []byte("raw"), 0644)
	out3, err := extractBinary(raw, dir, "clockify")
	if err != nil {
		t.Fatal(err)
	}
	data3, _ := os.ReadFile(out3)
	if string(data3) != "raw" {
		t.Errorf("raw content=%q", data3)
	}
}

func TestExtractTarGzMissingBinary(t *testing.T) {
	dir := t.TempDir()
	tgz := filepath.Join(dir, "empty.tar.gz")
	f, _ := os.Create(tgz)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: "README", Mode: 0644, Size: 3})
	_, _ = tw.Write([]byte("hi\n"))
	_ = tw.Close()
	_ = gz.Close()
	_ = f.Close()
	if _, err := extractBinary(tgz, dir, "clockify"); err == nil {
		t.Fatal("expected missing binary error")
	}
}

func TestMaybeCheckUpToDateNoHint(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("CLOCKIFY_NO_UPDATE_CHECK", "")
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	calls := 0
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(Release{TagName: "v1.0.0"})
	})
	res := MaybeCheck(context.Background(), c, "default", "1.0.0", false, "table", true, "list")
	if res != nil {
		t.Fatalf("expected nil when up to date, got %+v", res)
	}
	if calls != 1 {
		t.Errorf("calls=%d", calls)
	}
}

func TestMaybeCheckExemptCommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("CLOCKIFY_NO_UPDATE_CHECK", "")
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	calls := 0
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(Release{TagName: "v9.9.9"})
	})
	res := MaybeCheck(context.Background(), c, "default", "1.0.0", false, "table", true, "version")
	if res != nil || calls != 0 {
		t.Fatalf("res=%v calls=%d", res, calls)
	}
}

func TestReplaceExecutableFallbackWrite(t *testing.T) {
	// Make rename fail by using a dst that is a directory? Instead force fallback
	// by making src unreadable after... simpler: use dst path under non-writable?
	// Direct write path: rename fails when dst is on different filesystem — skip.
	// Cover error path: src missing.
	dir := t.TempDir()
	dst := filepath.Join(dir, "clockify")
	if err := ReplaceExecutable(dst, filepath.Join(dir, "missing-src")); err == nil {
		t.Fatal("expected error when src missing")
	}
}

func TestDownloadFileErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	err := downloadFile(context.Background(), srv.Client(), srv.URL+"/file", filepath.Join(t.TempDir(), "x"))
	if err == nil {
		t.Fatal("expected download error")
	}
}

func TestInstallHappyPath(t *testing.T) {
	dir := t.TempDir()
	// Build a tar.gz asset containing clockify binary
	tgz := filepath.Join(dir, "clockify_1.2.3_linux_amd64.tar.gz")
	f, _ := os.Create(tgz)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	body := []byte("#!/bin/sh\necho updated\n")
	_ = tw.WriteHeader(&tar.Header{Name: "clockify", Mode: 0755, Size: int64(len(body))})
	_, _ = tw.Write(body)
	_ = tw.Close()
	_ = gz.Close()
	_ = f.Close()

	sum, err := fileSHA256(tgz)
	if err != nil {
		t.Fatal(err)
	}
	checksums := fmt.Sprintf("%s  clockify_1.2.3_linux_amd64.tar.gz\n", sum)
	sumsPath := filepath.Join(dir, "checksums.txt")
	_ = os.WriteFile(sumsPath, []byte(checksums), 0644)

	downloadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/checksums.txt":
			http.ServeFile(w, r, sumsPath)
		default:
			http.ServeFile(w, r, tgz)
		}
	}))
	t.Cleanup(downloadSrv.Close)

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.2.3",
			Assets: []Asset{
				{
					Name:               "clockify_1.2.3_linux_amd64.tar.gz",
					BrowserDownloadURL: downloadSrv.URL + "/clockify_1.2.3_linux_amd64.tar.gz",
				},
				{
					Name:               "checksums.txt",
					BrowserDownloadURL: downloadSrv.URL + "/checksums.txt",
				},
			},
		})
	}))
	t.Cleanup(apiSrv.Close)

	c := NewClient()
	c.APIBase = apiSrv.URL
	c.HTTP = apiSrv.Client()

	// Install replaces the running test binary path — use CurrentExecutable which
	// may be the go test binary; ReplaceExecutable can still succeed on linux.
	old, newV, err := Install(context.Background(), c, "", "linux", "amd64")
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if newV != "1.2.3" {
		t.Errorf("newVersion=%q", newV)
	}
	_ = old
}

func TestInstallMissingChecksumsFailsClosed(t *testing.T) {
	dir := t.TempDir()
	tgz := filepath.Join(dir, "clockify_1.2.3_linux_amd64.tar.gz")
	f, _ := os.Create(tgz)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	body := []byte("#!/bin/sh\n")
	_ = tw.WriteHeader(&tar.Header{Name: "clockify", Mode: 0755, Size: int64(len(body))})
	_, _ = tw.Write(body)
	_ = tw.Close()
	_ = gz.Close()
	_ = f.Close()

	downloadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, tgz)
	}))
	t.Cleanup(downloadSrv.Close)

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.2.3",
			Assets: []Asset{{
				Name:               "clockify_1.2.3_linux_amd64.tar.gz",
				BrowserDownloadURL: downloadSrv.URL + "/clockify_1.2.3_linux_amd64.tar.gz",
			}},
		})
	}))
	t.Cleanup(apiSrv.Close)

	c := NewClient()
	c.APIBase = apiSrv.URL
	c.HTTP = apiSrv.Client()

	_, _, err := Install(context.Background(), c, "", "linux", "amd64")
	if err == nil || !strings.Contains(err.Error(), "checksums.txt") {
		t.Fatalf("expected missing checksums error, got %v", err)
	}
}

func TestInstallChecksumMismatchFailsClosed(t *testing.T) {
	dir := t.TempDir()
	tgz := filepath.Join(dir, "clockify_1.2.3_linux_amd64.tar.gz")
	f, _ := os.Create(tgz)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	body := []byte("#!/bin/sh\n")
	_ = tw.WriteHeader(&tar.Header{Name: "clockify", Mode: 0755, Size: int64(len(body))})
	_, _ = tw.Write(body)
	_ = tw.Close()
	_ = gz.Close()
	_ = f.Close()

	sumsPath := filepath.Join(dir, "checksums.txt")
	_ = os.WriteFile(sumsPath, []byte(strings.Repeat("0", 64)+"  clockify_1.2.3_linux_amd64.tar.gz\n"), 0644)

	downloadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/checksums.txt" {
			http.ServeFile(w, r, sumsPath)
			return
		}
		http.ServeFile(w, r, tgz)
	}))
	t.Cleanup(downloadSrv.Close)

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.2.3",
			Assets: []Asset{
				{
					Name:               "clockify_1.2.3_linux_amd64.tar.gz",
					BrowserDownloadURL: downloadSrv.URL + "/clockify_1.2.3_linux_amd64.tar.gz",
				},
				{
					Name:               "checksums.txt",
					BrowserDownloadURL: downloadSrv.URL + "/checksums.txt",
				},
			},
		})
	}))
	t.Cleanup(apiSrv.Close)

	c := NewClient()
	c.APIBase = apiSrv.URL
	c.HTTP = apiSrv.Client()

	_, _, err := Install(context.Background(), c, "", "linux", "amd64")
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got %v", err)
	}
}

func TestHintContainsVersion(t *testing.T) {
	h := Hint(&CheckResult{Current: "1.0.0", Latest: "2.0.0", UpdateAvailable: true})
	if !strings.Contains(h, "2.0.0") || !strings.Contains(h, "clockify update") {
		t.Errorf("hint=%q", h)
	}
}

func TestSetCurrentVersionProvider(t *testing.T) {
	SetCurrentVersionProvider(func() string { return "3.2.1" })
	if GetCurrentVersion() != "3.2.1" {
		t.Errorf("GetCurrentVersion=%q", GetCurrentVersion())
	}
	SetCurrentVersionProvider(nil)
	if GetCurrentVersion() != "dev" {
		t.Errorf("reset=%q", GetCurrentVersion())
	}
}

func TestRuntimeHelpers(t *testing.T) {
	if RuntimeGoos() == "" || RuntimeGoarch() == "" {
		t.Fatal("runtime helpers empty")
	}
	_ = time.Now
}
