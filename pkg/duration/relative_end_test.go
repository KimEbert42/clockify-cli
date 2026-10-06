package duration

import (
	"testing"
	"time"
)

func TestParseRelativeDateEndIn(t *testing.T) {
	loc := time.UTC
	// 2026-10-07 is a Wednesday
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, loc)

	tests := []struct {
		input     string
		weekStart string
		wantDay   string // YYYY-MM-DD
	}{
		{"today", "monday", "2026-10-07"},
		{"yesterday", "monday", "2026-10-06"},
		{"this-week", "monday", "2026-10-11"},   // Sunday
		{"last-week", "monday", "2026-10-04"},   // previous Sunday
		{"this-week", "sunday", "2026-10-10"},   // Saturday if week starts Sunday
		{"this-month", "monday", "2026-10-31"},
		{"last-month", "monday", "2026-09-30"},
	}

	for _, tt := range tests {
		t.Run(tt.input+"_"+tt.weekStart, func(t *testing.T) {
			got, err := ParseRelativeDateEndIn(tt.input, tt.weekStart, loc, now)
			if err != nil {
				t.Fatalf("ParseRelativeDateEndIn(%q) error = %v", tt.input, err)
			}
			if got.Format("2006-01-02") != tt.wantDay {
				t.Errorf("end day = %s, want %s (full=%s)", got.Format("2006-01-02"), tt.wantDay, got)
			}
			if got.Hour() != 23 || got.Minute() != 59 {
				t.Errorf("end time = %s, want ~23:59", got.Format("15:04:05"))
			}
			// End must be strictly after period start
			start, err := ParseRelativeDateIn(tt.input, tt.weekStart, loc, now)
			if err != nil {
				t.Fatal(err)
			}
			if !got.After(start) {
				t.Errorf("end %s not after start %s", got, start)
			}
		})
	}
}

func TestParseRelativeDateEndRejectsUnknown(t *testing.T) {
	if _, err := ParseRelativeDateEndIn("tomorrow", "monday", time.UTC, time.Now()); err == nil {
		t.Fatal("expected error for unknown relative date")
	}
}

func TestRelativeStartAndEndSpanFullWeek(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, loc)
	start, err := ParseRelativeDateIn("this-week", "monday", loc, now)
	if err != nil {
		t.Fatal(err)
	}
	end, err := ParseRelativeDateEndIn("this-week", "monday", loc, now)
	if err != nil {
		t.Fatal(err)
	}
	days := end.Sub(start).Hours() / 24
	if days < 6.9 || days > 7.1 {
		t.Errorf("this-week span = %.2f days, want ~7 (start=%s end=%s)", days, start, end)
	}
}
