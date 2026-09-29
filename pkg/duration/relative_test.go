package duration

import (
	"testing"
	"time"
)

func TestParseRelativeDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	tests := []struct {
		input     string
		weekStart string
		expected  time.Time
		wantErr   bool
	}{
		{"today", "monday", today, false},
		{"yesterday", "monday", today.AddDate(0, 0, -1), false},
		{"this-week", "monday", startOfWeek(today, "monday"), false},
		{"last-week", "monday", startOfWeek(today, "monday").AddDate(0, 0, -7), false},
		{"this-month", "monday", time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), false},
		{"last-month", "monday", time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -1, 0), false},

		// Sunday week start
		{"this-week", "sunday", startOfWeek(today, "sunday"), false},
		{"last-week", "sunday", startOfWeek(today, "sunday").AddDate(0, 0, -7), false},

		// Errors
		{"", "monday", time.Time{}, true},
		{"tomorrow", "monday", time.Time{}, true},
		{"next-week", "monday", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input+"_"+tt.weekStart, func(t *testing.T) {
			got, err := ParseRelativeDate(tt.input, tt.weekStart)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRelativeDate(%q, %q) error = %v, wantErr %v", tt.input, tt.weekStart, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.expected) {
				t.Errorf("ParseRelativeDate(%q, %q) = %v, want %v", tt.input, tt.weekStart, got, tt.expected)
			}
		})
	}
}

func TestIsRelativeDate(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"today", true},
		{"yesterday", true},
		{"this-week", true},
		{"last-week", true},
		{"this-month", true},
		{"last-month", true},
		{"tomorrow", false},
		{"next-week", false},
		{"2024-01-15", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsRelativeDate(tt.input)
			if got != tt.expected {
				t.Errorf("IsRelativeDate(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStartOfWeekMonday(t *testing.T) {
	// 2024-01-10 is a Wednesday
	d := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	got := startOfWeek(d, "monday")
	want := time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC) // Monday
	if !got.Equal(want) {
		t.Errorf("startOfWeek(Wednesday, monday) = %v, want %v", got, want)
	}

	// 2024-01-07 is a Sunday
	d = time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)
	got = startOfWeek(d, "monday")
	want = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // Previous Monday
	if !got.Equal(want) {
		t.Errorf("startOfWeek(Sunday, monday) = %v, want %v", got, want)
	}

	// 2024-01-01 is a Monday
	d = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	got = startOfWeek(d, "monday")
	want = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("startOfWeek(Monday, monday) = %v, want %v", got, want)
	}
}

func TestStartOfWeekSunday(t *testing.T) {
	// 2024-01-10 is a Wednesday
	d := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	got := startOfWeek(d, "sunday")
	want := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC) // Sunday
	if !got.Equal(want) {
		t.Errorf("startOfWeek(Wednesday, sunday) = %v, want %v", got, want)
	}

	// 2024-01-07 is a Sunday
	d = time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)
	got = startOfWeek(d, "sunday")
	want = time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("startOfWeek(Sunday, sunday) = %v, want %v", got, want)
	}
}
