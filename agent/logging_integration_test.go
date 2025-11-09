package agent

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestLoggingIntegration verifies that logging is called at key points
func TestLoggingIntegration(t *testing.T) {
	// Create a simple logger that captures log messages
	var logs []string
	captureLogger := &testCaptureLogger{logs: &logs}

	t.Run("Ask_WithLogging", func(t *testing.T) {
		logs = []string{} // Reset
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithLogger(captureLogger).
			WithMemoryCache(100, 5*time.Minute)

		// This will fail due to invalid API key, but should generate logs
		_, _ = builder.Ask(context.Background(), "Hello")

		// Verify we got some log entries
		if len(logs) == 0 {
			t.Error("Expected log entries, got none")
		}

		// Check for key log messages
		hasStartLog := false
		hasErrorLog := false
		for _, log := range logs {
			if strings.Contains(log, "Ask request started") {
				hasStartLog = true
			}
			if strings.Contains(log, "error") || strings.Contains(log, "ERROR") {
				hasErrorLog = true
			}
		}

		if !hasStartLog {
			t.Error("Expected 'Ask request started' log")
		}
		if !hasErrorLog {
			t.Error("Expected error log")
		}
	})

	t.Run("Stream_WithLogging", func(t *testing.T) {
		logs = []string{} // Reset
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithLogger(captureLogger)

		// This will fail due to invalid API key, but should generate logs
		_, _ = builder.Stream(context.Background(), "Hello")

		// Verify we got some log entries
		if len(logs) == 0 {
			t.Error("Expected log entries, got none")
		}

		// Check for stream start log
		hasStreamLog := false
		for _, log := range logs {
			if strings.Contains(log, "Stream request started") {
				hasStreamLog = true
			}
		}

		if !hasStreamLog {
			t.Error("Expected 'Stream request started' log")
		}
	})

	t.Run("Cache_WithLogging", func(t *testing.T) {
		logs = []string{} // Reset
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithLogger(captureLogger).
			WithMemoryCache(100, 5*time.Minute)

		// Get cache stats
		_ = builder.GetCacheStats()

		// Check for cache stats log
		hasCacheLog := false
		for _, log := range logs {
			if strings.Contains(log, "Cache stats") || strings.Contains(log, "No cache") {
				hasCacheLog = true
			}
		}

		if !hasCacheLog {
			t.Error("Expected cache stats log")
		}
	})

	t.Run("ClearCache_WithLogging", func(t *testing.T) {
		logs = []string{} // Reset
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithLogger(captureLogger).
			WithMemoryCache(100, 5*time.Minute)

		// Clear cache
		_ = builder.ClearCache(context.Background())

		// Check for clear cache log
		hasClearLog := false
		for _, log := range logs {
			if strings.Contains(log, "Clearing cache") || strings.Contains(log, "Cache cleared") {
				hasClearLog = true
			}
		}

		if !hasClearLog {
			t.Error("Expected cache clear log")
		}
	})

	t.Run("NoopLogger_NoOutput", func(t *testing.T) {
		logs = []string{} // Reset
		// Default builder uses NoopLogger
		builder := NewOpenAI("gpt-4o-mini", "test-key")

		// This should not produce any logs
		_, _ = builder.Ask(context.Background(), "Hello")

		// Verify no logs were captured (NoopLogger should be used)
		if len(logs) > 0 {
			t.Errorf("Expected no logs with NoopLogger, got %d logs", len(logs))
		}
	})
}

// testCaptureLogger captures log messages for testing
type testCaptureLogger struct {
	logs *[]string
}

func (l *testCaptureLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	*l.logs = append(*l.logs, "DEBUG: "+msg+l.formatFields(fields))
}

func (l *testCaptureLogger) Info(ctx context.Context, msg string, fields ...Field) {
	*l.logs = append(*l.logs, "INFO: "+msg+l.formatFields(fields))
}

func (l *testCaptureLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	*l.logs = append(*l.logs, "WARN: "+msg+l.formatFields(fields))
}

func (l *testCaptureLogger) Error(ctx context.Context, msg string, fields ...Field) {
	*l.logs = append(*l.logs, "ERROR: "+msg+l.formatFields(fields))
}

func (l *testCaptureLogger) formatFields(fields []Field) string {
	if len(fields) == 0 {
		return ""
	}
	var parts []string
	for _, f := range fields {
		parts = append(parts, f.Key+"="+toString(f.Value))
	}
	return " | " + strings.Join(parts, " ")
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int64:
		return "number"
	case float64:
		return "float"
	case bool:
		return "bool"
	default:
		return "value"
	}
}
