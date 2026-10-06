package duration

import (
	"time"
)

// ParseRelativeDate converts a relative date string to the **start** of that
// period (midnight local). For inclusive range ends use ParseRelativeDateEnd.
// Supported values: today, yesterday, this-week, last-week, this-month, last-month
// weekStart should be "monday" or "sunday" (default: monday).
func ParseRelativeDate(s, weekStart string) (time.Time, error) {
	return ParseRelativeDateIn(s, weekStart, time.Local, time.Now())
}

// ParseRelativeDateEnd converts a relative date string to the inclusive **end**
// of that period (23:59:59.999… local on the last day). weekStart "monday"/"sunday".
func ParseRelativeDateEnd(s, weekStart string) (time.Time, error) {
	return ParseRelativeDateEndIn(s, weekStart, time.Local, time.Now())
}

// startOfWeek returns the start of the week containing d.
// weekStart should be "monday" or "sunday".
func startOfWeek(d time.Time, weekStart string) time.Time {
	weekday := d.Weekday()

	var offset int
	if weekStart == "sunday" {
		offset = int(weekday)
	} else {
		// Monday start (default)
		offset = int(weekday) - 1
		if offset < 0 {
			offset = 6 // Sunday -> 6 days back to Monday
		}
	}

	return d.AddDate(0, 0, -offset)
}

// IsRelativeDate returns true if the string is a known relative date.
func IsRelativeDate(s string) bool {
	switch s {
	case "today", "yesterday", "this-week", "last-week", "this-month", "last-month":
		return true
	}
	return false
}
