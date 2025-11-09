package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

// NewDateTimeTool creates a tool for date and time operations.
// Supports multiple formats, timezones, and common date/time calculations.
//
// Available operations:
//   - current_time: Get current time in specified timezone and format
//   - format_date: Format a date string to another format
//   - parse_date: Parse a date string and get details
//   - add_duration: Add duration to a date (days, hours, minutes)
//   - date_diff: Calculate difference between two dates
//   - convert_timezone: Convert time from one timezone to another
//   - day_of_week: Get day of week for a date
//
// Example:
//
//	dtTool := tools.NewDateTimeTool()
//	agent.NewOpenAI("gpt-4o", apiKey).
//	    WithTool(dtTool).
//	    WithAutoExecute().
//	    Ask(ctx, "What day of the week is Christmas 2025?")
func NewDateTimeTool() *agent.Tool {
	return agent.NewTool("datetime", "Date and time operations: current time, formatting, parsing, calculations, timezone conversion").
		AddParameter("operation", "string", "Operation: current_time, format_date, parse_date, add_duration, date_diff, convert_timezone, day_of_week", true).
		AddParameter("date", "string", "Date string (format: 2006-01-02 or 2006-01-02 15:04:05)", false).
		AddParameter("format", "string", "Output format: RFC3339, RFC1123, Unix, or custom Go format", false).
		AddParameter("timezone", "string", "Timezone (e.g., UTC, America/New_York, Asia/Tokyo)", false).
		AddParameter("duration", "string", "Duration to add (e.g., 24h, 30m, 7d for days)", false).
		AddParameter("date2", "string", "Second date for comparison (date_diff)", false).
		WithHandler(dateTimeHandler)
}

// dateTimeHandler executes date/time operations
func dateTimeHandler(args string) (string, error) {
	var params struct {
		Operation string `json:"operation"`
		Date      string `json:"date"`
		Format    string `json:"format"`
		Timezone  string `json:"timezone"`
		Duration  string `json:"duration"`
		Date2     string `json:"date2"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Execute operation
	switch params.Operation {
	case "current_time":
		return getCurrentTime(params.Timezone, params.Format)
	case "format_date":
		return formatDate(params.Date, params.Format, params.Timezone)
	case "parse_date":
		return parseDate(params.Date, params.Timezone)
	case "add_duration":
		return addDuration(params.Date, params.Duration, params.Timezone)
	case "date_diff":
		return dateDiff(params.Date, params.Date2)
	case "convert_timezone":
		return convertTimezone(params.Date, params.Timezone)
	case "day_of_week":
		return dayOfWeek(params.Date)
	default:
		return "", fmt.Errorf("unknown operation: %s", params.Operation)
	}
}

// getCurrentTime returns the current time
func getCurrentTime(tz, format string) (string, error) {
	loc, err := getLocation(tz)
	if err != nil {
		return "", err
	}

	now := time.Now().In(loc)
	formatted := formatTime(now, format)

	return fmt.Sprintf("Current time in %s:\n%s\nUnix: %d", loc.String(), formatted, now.Unix()), nil
}

// formatDate formats a date string to another format
func formatDate(dateStr, format, tz string) (string, error) {
	t, err := parseDateTime(dateStr)
	if err != nil {
		return "", err
	}

	if tz != "" {
		loc, err := getLocation(tz)
		if err != nil {
			return "", err
		}
		t = t.In(loc)
	}

	formatted := formatTime(t, format)
	return fmt.Sprintf("Formatted date:\n%s", formatted), nil
}

// parseDate parses a date string and returns details
func parseDate(dateStr, tz string) (string, error) {
	t, err := parseDateTime(dateStr)
	if err != nil {
		return "", err
	}

	if tz != "" {
		loc, err := getLocation(tz)
		if err != nil {
			return "", err
		}
		t = t.In(loc)
	}

	var result strings.Builder
	result.WriteString("Parsed date details:\n")
	result.WriteString(fmt.Sprintf("  Date: %s\n", t.Format("2006-01-02")))
	result.WriteString(fmt.Sprintf("  Time: %s\n", t.Format("15:04:05")))
	result.WriteString(fmt.Sprintf("  Timezone: %s\n", t.Location()))
	result.WriteString(fmt.Sprintf("  Day of week: %s\n", t.Weekday()))
	result.WriteString(fmt.Sprintf("  Day of year: %d\n", t.YearDay()))
	result.WriteString(fmt.Sprintf("  Week number: %d\n", getWeekNumber(t)))
	result.WriteString(fmt.Sprintf("  Unix timestamp: %d\n", t.Unix()))
	result.WriteString(fmt.Sprintf("  RFC3339: %s\n", t.Format(time.RFC3339)))

	return result.String(), nil
}

// addDuration adds duration to a date
func addDuration(dateStr, duration, tz string) (string, error) {
	t, err := parseDateTime(dateStr)
	if err != nil {
		return "", err
	}

	// Parse duration (support days as well)
	d, err := parseDuration(duration)
	if err != nil {
		return "", err
	}

	newTime := t.Add(d)

	if tz != "" {
		loc, err := getLocation(tz)
		if err != nil {
			return "", err
		}
		newTime = newTime.In(loc)
	}

	return fmt.Sprintf("Original: %s\nDuration: %s\nResult: %s",
		t.Format(time.RFC3339), duration, newTime.Format(time.RFC3339)), nil
}

// dateDiff calculates difference between two dates
func dateDiff(date1Str, date2Str string) (string, error) {
	t1, err := parseDateTime(date1Str)
	if err != nil {
		return "", fmt.Errorf("invalid date1: %w", err)
	}

	t2, err := parseDateTime(date2Str)
	if err != nil {
		return "", fmt.Errorf("invalid date2: %w", err)
	}

	diff := t2.Sub(t1)
	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Date 1: %s\n", t1.Format(time.RFC3339)))
	result.WriteString(fmt.Sprintf("Date 2: %s\n", t2.Format(time.RFC3339)))
	result.WriteString(fmt.Sprintf("Difference: %d days, %d hours, %d minutes\n", days, hours, minutes))
	result.WriteString(fmt.Sprintf("Total hours: %.2f\n", diff.Hours()))
	result.WriteString(fmt.Sprintf("Total minutes: %.0f\n", diff.Minutes()))

	return result.String(), nil
}

// convertTimezone converts time from one timezone to another
func convertTimezone(dateStr, targetTZ string) (string, error) {
	t, err := parseDateTime(dateStr)
	if err != nil {
		return "", err
	}

	targetLoc, err := getLocation(targetTZ)
	if err != nil {
		return "", err
	}

	converted := t.In(targetLoc)

	return fmt.Sprintf("Original: %s (%s)\nConverted: %s (%s)",
		t.Format(time.RFC3339), t.Location(),
		converted.Format(time.RFC3339), targetLoc), nil
}

// dayOfWeek returns the day of week for a date
func dayOfWeek(dateStr string) (string, error) {
	t, err := parseDateTime(dateStr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Date: %s\nDay of week: %s\nWeek number: %d",
		t.Format("2006-01-02"), t.Weekday(), getWeekNumber(t)), nil
}

// Helper functions

// parseDateTime parses a date string with multiple format support
func parseDateTime(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}

	// Try common formats
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02-01-2006",
		time.RFC1123,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s (try format: 2006-01-02 or 2006-01-02 15:04:05)", dateStr)
}

// getLocation gets timezone location
func getLocation(tz string) (*time.Location, error) {
	if tz == "" {
		return time.UTC, nil
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %s (try: UTC, America/New_York, Asia/Tokyo)", tz)
	}

	return loc, nil
}

// formatTime formats time according to specified format
func formatTime(t time.Time, format string) string {
	if format == "" {
		format = time.RFC3339
	}

	switch strings.ToLower(format) {
	case "rfc3339":
		return t.Format(time.RFC3339)
	case "rfc1123":
		return t.Format(time.RFC1123)
	case "unix":
		return fmt.Sprintf("%d", t.Unix())
	default:
		// Try custom format
		return t.Format(format)
	}
}

// parseDuration parses duration with day support
func parseDuration(s string) (time.Duration, error) {
	// Check for days (e.g., "7d")
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		var days int
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// Standard duration parsing
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %s (use: 24h, 30m, 7d)", s)
	}

	return d, nil
}

// getWeekNumber returns ISO week number
func getWeekNumber(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}
