package duration

import (
	"fmt"
	"time"
)

// ParseRelativeDate converts a relative date string to a time.Time.
// Supported values: today, yesterday, this-week, last-week, this-month, last-month
// weekStart should be "monday" or "sunday" (default: monday).
func ParseRelativeDate(s string, weekStart string) (time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

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
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil

	case "last-month":
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return firstOfThisMonth.AddDate(0, -1, 0), nil

	default:
		return time.Time{}, fmt.Errorf("unknown relative date %q: supported values are today, yesterday, this-week, last-week, this-month, last-month", s)
	}
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
