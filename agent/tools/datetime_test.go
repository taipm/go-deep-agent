package tools

import (
	"strings"
	"testing"
	"time"
)

func TestDateTimeTool(t *testing.T) {
	t.Run("CurrentTime", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "current_time", "timezone": "UTC"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("CurrentTime failed: %v", err)
		}
		if !strings.Contains(result, "Current time") {
			t.Errorf("Unexpected result: %s", result)
		}
		if !strings.Contains(result, "UTC") {
			t.Errorf("Timezone not found in result: %s", result)
		}
	})

	t.Run("FormatDate", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "format_date", "date": "2025-01-15", "format": "RFC3339"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("FormatDate failed: %v", err)
		}
		if !strings.Contains(result, "2025-01-15") {
			t.Errorf("Date not formatted correctly: %s", result)
		}
	})

	t.Run("ParseDate", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "parse_date", "date": "2025-12-25"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("ParseDate failed: %v", err)
		}
		if !strings.Contains(result, "Parsed date details") {
			t.Errorf("Parse details not found: %s", result)
		}
		if !strings.Contains(result, "2025-12-25") {
			t.Errorf("Date not found in result: %s", result)
		}
	})

	t.Run("AddDuration", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "add_duration", "date": "2025-01-01", "duration": "7d"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("AddDuration failed: %v", err)
		}
		if !strings.Contains(result, "2025-01-08") {
			t.Errorf("Duration not added correctly: %s", result)
		}
	})

	t.Run("DateDiff", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "date_diff", "date": "2025-01-01", "date2": "2025-01-08"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("DateDiff failed: %v", err)
		}
		if !strings.Contains(result, "7 days") {
			t.Errorf("Difference not calculated correctly: %s", result)
		}
	})

	t.Run("ConvertTimezone", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "convert_timezone", "date": "2025-01-01 12:00:00", "timezone": "America/New_York"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("ConvertTimezone failed: %v", err)
		}
		if !strings.Contains(result, "America/New_York") {
			t.Errorf("Timezone conversion failed: %s", result)
		}
	})

	t.Run("DayOfWeek", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "day_of_week", "date": "2025-12-25"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("DayOfWeek failed: %v", err)
		}
		if !strings.Contains(result, "Day of week") {
			t.Errorf("Day of week not found: %s", result)
		}
		if !strings.Contains(result, "Thursday") {
			t.Errorf("Expected Thursday, got: %s", result)
		}
	})

	t.Run("InvalidOperation", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "invalid_op"}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected error for invalid operation")
		}
	})

	t.Run("InvalidDate", func(t *testing.T) {
		tool := NewDateTimeTool()
		args := `{"operation": "parse_date", "date": "invalid-date"}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected error for invalid date")
		}
	})
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"RFC3339", "2025-01-15T12:00:00Z", false},
		{"Date only", "2025-01-15", false},
		{"Date time", "2025-01-15 12:00:00", false},
		{"Slash format", "2025/01/15", false},
		{"US format", "01/15/2025", false},
		{"Invalid", "invalid", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDateTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"Days", "7d", 7 * 24 * time.Hour, false},
		{"Hours", "24h", 24 * time.Hour, false},
		{"Minutes", "30m", 30 * time.Minute, false},
		{"Seconds", "60s", 60 * time.Second, false},
		{"Invalid", "invalid", 0, true},
		{"Negative days", "-1d", -24 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetLocation(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{"Empty (UTC)", "", false},
		{"UTC", "UTC", false},
		{"America/New_York", "America/New_York", false},
		{"Asia/Tokyo", "Asia/Tokyo", false},
		{"Invalid", "Invalid/Timezone", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := getLocation(tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && loc == nil {
				t.Error("getLocation() returned nil location")
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2025, 1, 15, 12, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"Default (RFC3339)", "", "2025-01-15T12:30:45Z"},
		{"RFC3339", "RFC3339", "2025-01-15T12:30:45Z"},
		{"Unix", "unix", "1736944245"},
		{"Custom", "2006-01-02", "2025-01-15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(testTime, tt.format)
			if tt.name != "RFC1123" && result != tt.expected {
				t.Errorf("formatTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetWeekNumber(t *testing.T) {
	testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	week := getWeekNumber(testDate)

	if week < 1 || week > 53 {
		t.Errorf("getWeekNumber() = %d, want between 1 and 53", week)
	}
}
