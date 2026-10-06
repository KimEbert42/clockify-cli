// Package update implements GitHub Releases-based self-update and passive update checks.
package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultRepoOwner = "kim"
	DefaultRepoName  = "clockify-cli"
	defaultAPIBase   = "https://api.github.com"
	defaultTimeout   = 15 * time.Second
	checkTimeout     = 3 * time.Second
)

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type Client struct {
	APIBase   string
	RepoOwner string
	RepoName  string
	HTTP      *http.Client
}

func NewClient() *Client {
	return &Client{
		APIBase:   defaultAPIBase,
		RepoOwner: DefaultRepoOwner,
		RepoName:  DefaultRepoName,
		HTTP:      &http.Client{Timeout: defaultTimeout},
	}
}

func (c *Client) repoPath() string {
	return fmt.Sprintf("%s/repos/%s/%s", c.APIBase, c.RepoOwner, c.RepoName)
}

func (c *Client) getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("github api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// LatestRelease fetches the most recent non-draft release.
func (c *Client) LatestRelease(ctx context.Context) (*Release, error) {
	var rel Release
	url := c.repoPath() + "/releases/latest"
	if err := c.getJSON(ctx, url, &rel); err != nil {
		return nil, err
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("no release found")
	}
	return &rel, nil
}

// ReleaseByVersion fetches a release by tag (accepts "1.2.3" or "v1.2.3").
func (c *Client) ReleaseByVersion(ctx context.Context, version string) (*Release, error) {
	tag := NormalizeVersion(version)
	var rel Release
	url := c.repoPath() + "/releases/tags/" + tag
	if err := c.getJSON(ctx, url, &rel); err != nil {
		return nil, err
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("release %q not found", tag)
	}
	return &rel, nil
}

// FindAsset returns the archive asset for the given platform.
func FindAsset(rel *Release, goos, goarch string) (*Asset, error) {
	if rel == nil {
		return nil, fmt.Errorf("nil release")
	}
	prefs := assetNamePrefs(goos, goarch)
	for _, pref := range prefs {
		for i := range rel.Assets {
			if rel.Assets[i].Name == pref {
				return &rel.Assets[i], nil
			}
		}
	}
	// Fuzzy fallback: contain os and arch tokens
	for _, pref := range prefs {
		for i := range rel.Assets {
			n := strings.ToLower(rel.Assets[i].Name)
			if strings.Contains(n, strings.ToLower(pref)) ||
				(strings.Contains(n, goos) && (strings.Contains(n, goarch) || strings.Contains(n, "all") || strings.Contains(n, "universal")) && hasArchiveExt(n)) {
				return &rel.Assets[i], nil
			}
		}
	}
	names := make([]string, 0, len(rel.Assets))
	for _, a := range rel.Assets {
		names = append(names, a.Name)
	}
	return nil, fmt.Errorf("no release asset for %s/%s (have: %s)", goos, goarch, strings.Join(names, ", "))
}

func assetNamePrefs(goos, goarch string) []string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	prefs := []string{
		fmt.Sprintf("clockify_%s_%s%s", goos, goarch, ext),
		fmt.Sprintf("clockify_%s_%s_%s", goos, goarch, ext),
	}
	// GoReleaser universal macOS binaries publish as darwin_all.
	if goos == "darwin" {
		prefs = append(prefs,
			fmt.Sprintf("clockify_darwin_all%s", ext),
			fmt.Sprintf("clockify_macos_all%s", ext),
			fmt.Sprintf("clockify_darwin_universal%s", ext),
		)
	}
	return prefs
}

func hasArchiveExt(name string) bool {
	return strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip")
}

// NormalizeVersion strips a leading "v".
func NormalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// CompareVersions compares semver-ish strings. Returns -1, 0, or 1.
func CompareVersions(a, b string) int {
	an, bn := NormalizeVersion(a), NormalizeVersion(b)
	if an == bn {
		return 0
	}
	ap, bp := splitVersion(an), splitVersion(bn)
	for i := 0; i < 3; i++ {
		if ap[i] != bp[i] {
			if ap[i] < bp[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func splitVersion(v string) [3]int {
	v = strings.TrimSuffix(v, "-next")
	parts := strings.SplitN(v, ".", 3)
	var out [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		n := 0
		for _, c := range parts[i] {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		out[i] = n
	}
	return out
}

type CheckResult struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"update_available"`
}

// Check queries the latest release and compares to currentVersion.
func Check(ctx context.Context, client *Client, currentVersion string) (*CheckResult, error) {
	rel, err := client.LatestRelease(ctx)
	if err != nil {
		return nil, err
	}
	res := &CheckResult{
		Current: NormalizeVersion(currentVersion),
		Latest:  NormalizeVersion(rel.TagName),
	}
	res.UpdateAvailable = CompareVersions(res.Current, res.Latest) < 0
	return res, nil
}

// Install downloads and replaces the running binary with the given release asset.
// version may be empty for latest.
// Assets are verified against the release checksums.txt (sha256) before extract.
func Install(ctx context.Context, client *Client, version, goos, goarch string) (oldVersion, newVersion string, err error) {
	oldVersion = NormalizeVersion(GetCurrentVersion())
	var rel *Release
	if version == "" {
		rel, err = client.LatestRelease(ctx)
	} else {
		rel, err = client.ReleaseByVersion(ctx, version)
	}
	if err != nil {
		return oldVersion, "", err
	}
	newVersion = NormalizeVersion(rel.TagName)

	asset, err := FindAsset(rel, goos, goarch)
	if err != nil {
		return oldVersion, newVersion, err
	}

	binName := "clockify"
	if goos == "windows" {
		binName = "clockify.exe"
	}

	tmpDir, err := os.MkdirTemp("", "clockify-update-*")
	if err != nil {
		return oldVersion, newVersion, err
	}
	defer os.RemoveAll(tmpDir)

	assetPath := filepath.Join(tmpDir, asset.Name)
	if err := downloadFile(ctx, client.HTTP, asset.BrowserDownloadURL, assetPath); err != nil {
		return oldVersion, newVersion, err
	}

	// Fail closed: require checksums.txt match before replacing the binary.
	if err := verifyChecksums(ctx, client, rel, *asset, assetPath); err != nil {
		return oldVersion, newVersion, err
	}

	extracted, err := extractBinary(assetPath, tmpDir, binName)
	if err != nil {
		return oldVersion, newVersion, err
	}
	if err := os.Chmod(extracted, 0755); err != nil {
		return oldVersion, newVersion, err
	}

	exe, err := CurrentExecutable()
	if err != nil {
		return oldVersion, newVersion, err
	}
	if err := ReplaceExecutable(exe, extracted); err != nil {
		return oldVersion, newVersion, err
	}
	return oldVersion, newVersion, nil
}

// verifyChecksums downloads checksums.txt from the release and checks asset sha256.
func verifyChecksums(ctx context.Context, client *Client, rel *Release, asset Asset, assetPath string) error {
	var sumsAsset *Asset
	for i := range rel.Assets {
		if strings.EqualFold(rel.Assets[i].Name, "checksums.txt") {
			sumsAsset = &rel.Assets[i]
			break
		}
	}
	if sumsAsset == nil {
		return fmt.Errorf("release %s is missing checksums.txt; refusing to install unverified binary", rel.TagName)
	}

	sumsPath := filepath.Join(filepath.Dir(assetPath), "checksums.txt")
	if err := downloadFile(ctx, client.HTTP, sumsAsset.BrowserDownloadURL, sumsPath); err != nil {
		return fmt.Errorf("failed to download checksums.txt: %w", err)
	}
	sums, err := os.ReadFile(sumsPath)
	if err != nil {
		return fmt.Errorf("failed to read checksums.txt: %w", err)
	}

	want, err := lookupSHA256(string(sums), asset.Name)
	if err != nil {
		return err
	}
	got, err := fileSHA256(assetPath)
	if err != nil {
		return fmt.Errorf("failed to hash %s: %w", asset.Name, err)
	}
	if !strings.EqualFold(want, got) {
		return fmt.Errorf("checksum mismatch for %s: want %s got %s (refusing install)", asset.Name, want, got)
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func lookupSHA256(sumsBody, filename string) (string, error) {
	for _, line := range strings.Split(sumsBody, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// format: <sha256>  <filename>  (or single space)
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[1] == filename || strings.HasSuffix(fields[1], "/"+filename) || filepath.Base(fields[1]) == filename {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("no checksum entry for %s in checksums.txt", filename)
}

func downloadFile(ctx context.Context, client *http.Client, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func extractBinary(archivePath, destDir, binName string) (string, error) {
	target := filepath.Join(destDir, binName)
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(archivePath, destDir, binName)
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(archivePath, destDir, binName)
	default:
		// Assume raw binary asset
		data, err := os.ReadFile(archivePath)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(target, data, 0755); err != nil {
			return "", err
		}
		return target, nil
	}
}

func extractTarGz(path, destDir, binName string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	target := filepath.Join(destDir, binName)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		name := filepath.Base(hdr.Name)
		if name != binName && name != strings.TrimSuffix(binName, ".exe") {
			// also accept bare "clockify" when looking for .exe
			if name != "clockify" && name != "clockify.exe" {
				continue
			}
		}
		out, err := os.Create(target)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return "", err
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return target, nil
	}
	return "", fmt.Errorf("binary %q not found in archive", binName)
}

func extractZip(path, destDir, binName string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()
	target := filepath.Join(destDir, binName)
	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if name != binName && name != "clockify" && name != "clockify.exe" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return "", err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			return "", err
		}
		return target, nil
	}
	return "", fmt.Errorf("binary %q not found in zip", binName)
}

// CurrentExecutable returns the path to the running binary.
func CurrentExecutable() (string, error) {
	return os.Executable()
}

// ReplaceExecutable replaces dst with the file at src.
func ReplaceExecutable(dst, src string) error {
	// Prefer atomic rename when possible.
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Fallback: write alongside then rename.
	tmp := dst + ".new"
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0755); err != nil {
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		// Last resort: direct write
		if werr := os.WriteFile(dst, data, 0755); werr != nil {
			return fmt.Errorf("failed to replace binary: %w (rename: %v)", werr, err)
		}
	}
	return nil
}

// GetCurrentVersion returns the build version baked into this binary.
func GetCurrentVersion() string {
	return currentVersionFn()
}

var currentVersionFn = func() string { return "dev" }

// SetCurrentVersionProvider overrides how the running version is read (tests).
func SetCurrentVersionProvider(fn func() string) {
	if fn == nil {
		currentVersionFn = func() string { return "dev" }
		return
	}
	currentVersionFn = fn
}

// IsUnsupportedInstall reports whether this build cannot self-update.
func IsUnsupportedInstall(version string) bool {
	v := NormalizeVersion(version)
	return v == "" || v == "dev" || v == "unknown" || !hasNumericVersion(v)
}

func hasNumericVersion(v string) bool {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) == 0 {
		return false
	}
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return parts[0] != ""
}

// DefaultGoos/goarch helpers for tests.
func RuntimeGoos() string { return runtime.GOOS }
func RuntimeGoarch() string { return runtime.GOARCH }
