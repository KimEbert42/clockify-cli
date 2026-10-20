package clockify

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FlexibleDuration is a duration in seconds that unmarshals from either a JSON
// number (seconds) or a duration string: ISO-8601 (e.g. "PT400H", "PT1H30M")
// or a plain second count (e.g. "3600"). Clockify uses both shapes across
// endpoints and API versions.
type FlexibleDuration int64

var iso8601DurationRe = regexp.MustCompile(
	`^P(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`,
)

// Int64 returns the duration as seconds.
func (d FlexibleDuration) Int64() int64 { return int64(d) }

func (d *FlexibleDuration) UnmarshalJSON(b []byte) error {
	b = bytesTrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*d = 0
		return nil
	}
	if b[0] != '"' {
		var n int64
		if err := json.Unmarshal(b, &n); err != nil {
			return fmt.Errorf("duration must be a number or string, got %s: %w", b, err)
		}
		*d = FlexibleDuration(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return d.UnmarshalText([]byte(s))
}

func (d *FlexibleDuration) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		*d = 0
		return nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		*d = FlexibleDuration(n)
		return nil
	}
	if secs, ok := parseISO8601Duration(s); ok {
		*d = FlexibleDuration(secs)
		return nil
	}
	return fmt.Errorf("cannot parse duration %q", s)
}

func (d FlexibleDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(d))
}

// ISO8601 renders the duration as an ISO-8601 string (e.g. "PT2H", "PT1H30M").
// Clockify task write endpoints require this form.
func (d FlexibleDuration) ISO8601() string {
	total := int64(d)
	if total <= 0 {
		return "PT0S"
	}
	days := total / 86400
	total %= 86400
	hours := total / 3600
	total %= 3600
	mins := total / 60
	secs := total % 60
	var b strings.Builder
	b.WriteString("P")
	if days > 0 {
		fmt.Fprintf(&b, "%dD", days)
	}
	if hours > 0 || mins > 0 || secs > 0 {
		b.WriteString("T")
		if hours > 0 {
			fmt.Fprintf(&b, "%dH", hours)
		}
		if mins > 0 {
			fmt.Fprintf(&b, "%dM", mins)
		}
		if secs > 0 || hours == 0 && mins == 0 {
			fmt.Fprintf(&b, "%dS", secs)
		}
	}
	if b.Len() == 1 { // just "P"
		return "PT0S"
	}
	return b.String()
}

func parseISO8601Duration(s string) (int64, bool) {
	m := iso8601DurationRe.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	var total float64
	if m[1] != "" {
		days, _ := strconv.ParseFloat(m[1], 64)
		total += days * 24 * 3600
	}
	if m[2] != "" {
		h, _ := strconv.ParseFloat(m[2], 64)
		total += h * 3600
	}
	if m[3] != "" {
		min, _ := strconv.ParseFloat(m[3], 64)
		total += min * 60
	}
	if m[4] != "" {
		sec, _ := strconv.ParseFloat(m[4], 64)
		total += sec
	}
	return int64(total + 0.5), true
}

func bytesTrimSpace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t' || b[0] == '\n' || b[0] == '\r') {
		b = b[1:]
	}
	for len(b) > 0 {
		last := b[len(b)-1]
		if last == ' ' || last == '\t' || last == '\n' || last == '\r' {
			b = b[:len(b)-1]
			continue
		}
		break
	}
	return b
}
