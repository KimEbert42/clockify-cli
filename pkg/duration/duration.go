package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var durationRe = regexp.MustCompile(`^(?:(\d+(?:\.\d+)?)h)?(?:(\d+(?:\.\d+)?)m)?(?:(\d+(?:\.\d+)?)s)?$`)

// Parse converts a human-readable duration string to seconds.
// Supported formats:
//   - Raw seconds: "3600"
//   - Hours: "1h", "1.5h"
//   - Minutes: "30m", "1.5m"
//   - Seconds: "30s"
//   - Combined: "1h30m", "2h15m", "30m15s", "1h30m15s"
func Parse(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Try raw seconds first (all digits)
	if secs, err := strconv.ParseInt(s, 10, 64); err == nil {
		return secs, nil
	}

	matches := durationRe.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration %q: supported formats are seconds (3600), hours (1h, 1.5h), minutes (30m), or combined (1h30m)", s)
	}

	hStr, mStr, sStr := matches[1], matches[2], matches[3]
	if hStr == "" && mStr == "" && sStr == "" {
		return 0, fmt.Errorf("invalid duration %q: no time component found", s)
	}

	var total int64

	if hStr != "" {
		h, err := strconv.ParseFloat(hStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid hours in duration %q: %w", s, err)
		}
		total += int64(h * 3600)
	}

	if mStr != "" {
		m, err := strconv.ParseFloat(mStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid minutes in duration %q: %w", s, err)
		}
		total += int64(m * 60)
	}

	if sStr != "" {
		secs, err := strconv.ParseFloat(sStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid seconds in duration %q: %w", s, err)
		}
		total += int64(secs)
	}

	return total, nil
}

// Format converts seconds to a human-readable duration string.
// Examples: 3600 -> "1h", 5400 -> "1h30m", 1800 -> "30m", 90 -> "1m30s"
func Format(seconds int64) string {
	if seconds == 0 {
		return "0s"
	}

	var parts []string

	h := seconds / 3600
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
		seconds -= h * 3600
	}

	m := seconds / 60
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
		seconds -= m * 60
	}

	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, "")
}
