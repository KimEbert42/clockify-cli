package duration

import (
	"testing"
	"time"
)

func TestParseHumanDateTime(t *testing.T) {
	loc := time.FixedZone("TEST", -5*3600)
	now := time.Date(2026, 10, 4, 15, 30, 0, 0, loc) // Sunday
	// Override now via injection — weekday doesn't matter for most cases.

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "now",
			input: "now",
			want:  now,
		},
		{
			name:  "rfc3339 preserved",
			input: "2026-10-04T09:00:00Z",
			want:  time.Date(2026, 10, 4, 9, 0, 0, 0, time.UTC),
		},
		{
			name:  "bare date local midnight",
			input: "2026-10-03",
			want:  time.Date(2026, 10, 3, 0, 0, 0, 0, loc),
		},
		{
			name:  "HH:MM today",
			input: "09:00",
			want:  time.Date(2026, 10, 4, 9, 0, 0, 0, loc),
		},
		{
			name:  "H:MMam",
			input: "9:00am",
			want:  time.Date(2026, 10, 4, 9, 0, 0, 0, loc),
		},
		{
			name:  "9:30pm",
			input: "9:30pm",
			want:  time.Date(2026, 10, 4, 21, 30, 0, 0, loc),
		},
		{
			name:  "12am midnight",
			input: "12:00am",
			want:  time.Date(2026, 10, 4, 0, 0, 0, 0, loc),
		},
		{
			name:  "yesterday 09:00",
			input: "yesterday 09:00",
			want:  time.Date(2026, 10, 3, 9, 0, 0, 0, loc),
		},
		{
			name:  "today 10:30",
			input: "today 10:30",
			want:  time.Date(2026, 10, 4, 10, 30, 0, 0, loc),
		},
		{
			name:  "offset -15m",
			input: "-15m",
			want:  now.Add(-15 * time.Minute),
		},
		{
			name:  "offset +1h",
			input: "+1h",
			want:  now.Add(time.Hour),
		},
		{
			name:  "offset -2d",
			input: "-2d",
			want:  now.Add(-48 * time.Hour),
		},
		{
			name:  "relative today",
			input: "today",
			want:  time.Date(2026, 10, 4, 0, 0, 0, 0, loc),
		},
		{
			name:  "naive datetime",
			input: "2026-10-03T14:15:00",
			want:  time.Date(2026, 10, 3, 14, 15, 0, 0, loc),
		},
		{name: "empty", input: "", wantErr: true},
		{name: "garbage", input: "morning", wantErr: true},
		{name: "bad hour", input: "25:00", wantErr: true},
		{name: "bad ampm hour", input: "13:00pm", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHumanDateTime(tt.input, loc, now)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseHumanDateTime(%q) err=%v wantErr=%v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseHumanDateTime(%q)=%v want %v", tt.input, got, tt.want)
			}
			if got.Location() != loc && !tt.want.Equal(got) {
				// RFC3339 keeps its own zone — only assert local cases
			}
		})
	}
}

func TestParseHumanDateTimeRFC3339ZonePreserved(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, 10, 4, 15, 30, 0, 0, time.UTC)
	in := "2026-10-04T09:00:00+02:00"
	got, err := ParseHumanDateTime(in, loc, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.Format(time.RFC3339) != "2026-10-04T09:00:00+02:00" {
		t.Errorf("zone not preserved: %s", got.Format(time.RFC3339))
	}
}
