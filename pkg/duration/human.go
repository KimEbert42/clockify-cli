package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	timeOnlyRe     = regexp.MustCompile(`^(\d{1,2}):(\d{2})(?::(\d{2}))?\s*(?i)(am|pm)?$`)
	offsetRe       = regexp.MustCompile(`^(?i)([+-])(\d+)([smhd])$`)
	relativeTimeRe = regexp.MustCompile(`^(?i)(today|yesterday)\s+(\d{1,2}:\d{2}(?::\d{2})?\s*(?:am|pm)?)$`)
)

// ParseRelativeDateIn parses a relative date using the provided location and now.
// loc defaults to time.Local; now defaults to time.Now() in loc.
// Returns the **start** of the period (midnight local). For inclusive range
// ends use ParseRelativeDateEndIn.
func ParseRelativeDateIn(s, weekStart string, loc *time.Location, now time.Time) (time.Time, error) {
	if loc == nil {
		loc = time.Local
	}
	if now.IsZero() {
		now = time.Now().In(loc)
	} else {
		now = now.In(loc)
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	switch s {
	case "today":
		return today, nil
	case "yesterday":
		return today.AddDate(0, 0, -1), nil
	case "this-week":
		return startOfWeek(today, weekStart), nil
	case "last-week":
		return startOfWeek(today, weekStart).AddDate(0, 0, -7), nil
	case "this-month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc), nil
	case "last-month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, -1, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown relative date %q: supported values are today, yesterday, this-week, last-week, this-month, last-month", s)
	}
}

// ParseRelativeDateEndIn parses a relative date as the inclusive **end** of that
// period (23:59:59.999999999 local on the last day). Use this for --end flags.
func ParseRelativeDateEndIn(s, weekStart string, loc *time.Location, now time.Time) (time.Time, error) {
	start, err := ParseRelativeDateIn(s, weekStart, loc, now)
	if err != nil {
		return time.Time{}, err
	}
	if loc == nil {
		loc = time.Local
	}
	// Normalize to loc
	start = start.In(loc)
	switch s {
	case "today", "yesterday":
		return endOfDay(start, loc), nil
	case "this-week", "last-week":
		// Week runs 7 days from start
		return endOfDay(start.AddDate(0, 0, 6), loc), nil
	case "this-month", "last-month":
		// Last day of the month that contains start
		return endOfDay(time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, -1), loc), nil
	default:
		return endOfDay(start, loc), nil
	}
}

func endOfDay(d time.Time, loc *time.Location) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 999999999, loc)
}

// ParseHumanDateTime parses human-friendly date/time strings into a time.Time.
//
// Grammar (fixed, documented):
//
//	now
//	-15m / +1h / -2d          → offset from now (d = days)
//	HH:MM / H:MMam / H:MMpm   → today at that local time
//	yesterday HH:MM           → yesterday at that local time
//	today HH:MM               → today at that local time
//	YYYY-MM-DD                → that date at 00:00 local
//	YYYY-MM-DDTHH:MM[:SS]     → local naive → attach loc
//	RFC3339                   → unchanged, zone preserved
//	relative date only        → ParseRelativeDateIn (midnight local)
//
// loc defaults to time.Local; now defaults to time.Now() in loc.
func ParseHumanDateTime(s string, loc *time.Location, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date/time value")
	}
	if loc == nil {
		loc = time.Local
	}
	if now.IsZero() {
		now = time.Now().In(loc)
	} else {
		now = now.In(loc)
	}

	if strings.EqualFold(s, "now") {
		return now, nil
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	if IsRelativeDate(s) {
		return ParseRelativeDateIn(s, "monday", loc, now)
	}

	if m := relativeTimeRe.FindStringSubmatch(s); m != nil {
		base, err := ParseRelativeDateIn(strings.ToLower(m[1]), "monday", loc, now)
		if err != nil {
			return time.Time{}, err
		}
		return parseTimeOfDay(m[2], loc, base)
	}

	if m := offsetRe.FindStringSubmatch(s); m != nil {
		sign := 1
		if m[1] == "-" {
			sign = -1
		}
		n, err := strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid offset %q", s)
		}
		var d time.Duration
		switch strings.ToLower(m[3]) {
		case "s":
			d = time.Duration(n) * time.Second
		case "m":
			d = time.Duration(n) * time.Minute
		case "h":
			d = time.Duration(n) * time.Hour
		case "d":
			d = time.Duration(n) * 24 * time.Hour
		}
		return now.Add(time.Duration(sign) * d), nil
	}

	if t, err := time.ParseInLocation("2006-01-02", s, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", s, loc); err == nil {
		return t, nil
	}

	base := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	if t, err := parseTimeOfDay(s, loc, base); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date/time %q: use RFC3339, YYYY-MM-DD, HH:MM, now, ±offsets, or relative dates (e.g. yesterday 09:00)", s)
}

func parseTimeOfDay(s string, loc *time.Location, base time.Time) (time.Time, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	ampm := ""
	if len(s) >= 2 {
		suffix := s[len(s)-2:]
		if suffix == "am" || suffix == "pm" {
			ampm = suffix
			s = strings.TrimSpace(s[:len(s)-2])
		}
	}
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return time.Time{}, fmt.Errorf("invalid time %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time %q", s)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time %q", s)
	}
	sec := 0
	if len(parts) == 3 {
		sec, err = strconv.Atoi(parts[2])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time %q", s)
		}
	}
	if ampm != "" {
		if h < 1 || h > 12 {
			return time.Time{}, fmt.Errorf("invalid 12-hour time %q", s)
		}
		if ampm == "pm" && h != 12 {
			h += 12
		}
		if ampm == "am" && h == 12 {
			h = 0
		}
	} else if h < 0 || h > 23 {
		return time.Time{}, fmt.Errorf("invalid hour in %q", s)
	}
	if m < 0 || m > 59 || sec < 0 || sec > 59 {
		return time.Time{}, fmt.Errorf("invalid time %q", s)
	}
	return time.Date(base.Year(), base.Month(), base.Day(), h, m, sec, 0, loc), nil
}
