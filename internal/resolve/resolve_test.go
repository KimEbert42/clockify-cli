package resolve

import (
	"testing"
)

func TestIsUUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"550e8400e29b41d4a716446655440000", true},
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"abc123", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isUUID(tt.input); got != tt.want {
			t.Errorf("isUUID(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
