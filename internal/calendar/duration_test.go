package calendar

import (
	"testing"
	"time"
)

func TestParseTaskDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		// ISO 8601 (what Taskwarrior exports for duration UDAs)
		{"iso minutes", "PT30M", 30 * time.Minute, false},
		{"iso hour", "PT1H", time.Hour, false},
		{"iso hour minutes", "PT1H30M", 90 * time.Minute, false},
		{"iso seconds", "PT45S", 45 * time.Second, false},
		{"iso day", "P1D", 24 * time.Hour, false},
		{"iso day and time", "P1DT2H", 26 * time.Hour, false},
		{"iso week", "P1W", 7 * 24 * time.Hour, false},
		{"iso month before T", "P1M", durMonth, false},
		{"iso minute after T", "PT1M", time.Minute, false},
		{"iso lowercase", "pt15m", 15 * time.Minute, false},

		// Shorthand fallback
		{"shorthand min", "30min", 30 * time.Minute, false},
		{"shorthand h", "2h", 2 * time.Hour, false},
		{"shorthand combined", "1h30m", 90 * time.Minute, false},
		{"shorthand day", "2d", 48 * time.Hour, false},
		{"shorthand week", "1w", 7 * 24 * time.Hour, false},
		{"bare number is seconds", "90", 90 * time.Second, false},
		{"shorthand spaces", "1h 30min", 90 * time.Minute, false},

		// Errors
		{"empty", "", 0, true},
		{"garbage", "abc", 0, true},
		{"unknown unit", "5z", 0, true},
		{"iso missing value", "PTM", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTaskDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseTaskDuration(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTaskDuration(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseTaskDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
