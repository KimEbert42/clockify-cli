package duration

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		// Raw seconds
		{"0", 0, false},
		{"3600", 3600, false},
		{"1", 1, false},
		{"86400", 86400, false},

		// Hours
		{"1h", 3600, false},
		{"2h", 7200, false},
		{"24h", 86400, false},

		// Decimal hours
		{"1.5h", 5400, false},
		{"0.5h", 1800, false},
		{"2.25h", 8100, false},

		// Minutes
		{"30m", 1800, false},
		{"1m", 60, false},
		{"59m", 3540, false},

		// Decimal minutes
		{"1.5m", 90, false},
		{"30.5m", 1830, false},

		// Seconds
		{"30s", 30, false},
		{"1s", 1, false},
		{"59s", 59, false},

		// Combined
		{"1h30m", 5400, false},
		{"2h15m", 8100, false},
		{"1h30m15s", 5415, false},
		{"30m15s", 1815, false},
		{"2h30m45s", 9045, false},

		// With spaces
		{" 1h ", 3600, false},
		{" 1h30m ", 5400, false},

		// Errors
		{"", 0, true},
		{"abc", 0, true},
		{"1x", 0, true},
		{"h", 0, true},
		{"1h2x", 0, true},
		{"--1h", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("Parse(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0s"},
		{1, "1s"},
		{30, "30s"},
		{59, "59s"},
		{60, "1m"},
		{90, "1m30s"},
		{3540, "59m"},
		{3600, "1h"},
		{5400, "1h30m"},
		{8100, "2h15m"},
		{9045, "2h30m45s"},
		{86400, "24h"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := Format(tt.input)
			if got != tt.expected {
				t.Errorf("Format(%d) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseFormatRoundtrip(t *testing.T) {
	inputs := []string{"1h", "1h30m", "2h15m", "30m", "45s", "1h30m15s", "2h30m45s"}
	for _, input := range inputs {
		secs, err := Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", input, err)
			continue
		}
		back := Format(secs)
		secs2, err := Parse(back)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", back, err)
			continue
		}
		if secs != secs2 {
			t.Errorf("roundtrip failed: Parse(%q)=%d, Format(%d)=%q, Parse(%q)=%d", input, secs, secs, back, back, secs2)
		}
	}
}
