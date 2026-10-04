package update

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/kim/clockify-cli/internal/config"
)

const (
	updateCheckCacheFile = "update-check.json"
	updateCheckTTL       = 24 * time.Hour
)

// SkipReason explains why an update check should not run.
type SkipReason string

const (
	SkipNone          SkipReason = ""
	SkipQuiet         SkipReason = "quiet"
	SkipJSON          SkipReason = "json-output"
	SkipEnv           SkipReason = "CLOCKIFY_NO_UPDATE_CHECK"
	SkipCI            SkipReason = "ci"
	SkipConfig        SkipReason = "auto_update=false"
	SkipCommand       SkipReason = "exempt-command"
	SkipNonRelease    SkipReason = "dev-build"
	SkipCacheFresh    SkipReason = "cache-fresh"
)

// ShouldSkip reports whether passive update checking is disabled.
func ShouldSkip(quiet bool, outputFormat string, autoUpdateEnabled bool, currentVersion string, cmdName string) SkipReason {
	if quiet {
		return SkipQuiet
	}
	if strings.EqualFold(outputFormat, "json") {
		return SkipJSON
	}
	if os.Getenv("CLOCKIFY_NO_UPDATE_CHECK") != "" {
		return SkipEnv
	}
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		return SkipCI
	}
	if !autoUpdateEnabled {
		return SkipConfig
	}
	switch cmdName {
	case "update", "version", "completion", "help":
		return SkipCommand
	}
	if IsUnsupportedInstall(currentVersion) {
		return SkipNonRelease
	}
	return SkipNone
}

// CachedResult is stored under the profile cache directory.
type CachedResult struct {
	Current         string    `json:"current"`
	Latest          string    `json:"latest"`
	UpdateAvailable bool      `json:"update_available"`
	CheckedAt       time.Time `json:"checked_at"`
}

// MaybeCheck runs a passive update check when due. Returns the result when a
// hint should be printed, or nil when nothing to show. Never returns an error
// that should fail the primary command.
func MaybeCheck(ctx context.Context, client *Client, profile, currentVersion string, quiet bool, outputFormat string, autoUpdateEnabled bool, cmdName string) *CheckResult {
	reason := ShouldSkip(quiet, outputFormat, autoUpdateEnabled, currentVersion, cmdName)
	if reason != SkipNone {
		return nil
	}

	var cached CachedResult
	if ok, _ := config.ReadTTL(profile, updateCheckCacheFile, &cached, updateCheckTTL); ok {
		if cached.Current == NormalizeVersion(currentVersion) && cached.Latest != "" {
			if cached.UpdateAvailable && CompareVersions(cached.Current, cached.Latest) < 0 {
				return &CheckResult{
					Current:         cached.Current,
					Latest:          cached.Latest,
					UpdateAvailable: true,
				}
			}
			return nil
		}
	}

	cctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	res, err := Check(cctx, client, currentVersion)
	if err != nil {
		return nil // silent on network failure
	}

	_ = config.WriteTTL(profile, updateCheckCacheFile, CachedResult{
		Current:         res.Current,
		Latest:          res.Latest,
		UpdateAvailable: res.UpdateAvailable,
		CheckedAt:       time.Now().UTC(),
	})

	if res.UpdateAvailable {
		return res
	}
	return nil
}

// Hint returns the one-line upgrade hint.
func Hint(res *CheckResult) string {
	return "A new version of clockify is available (" + res.Latest + "). Run 'clockify update' to upgrade."
}
