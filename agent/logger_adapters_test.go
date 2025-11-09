package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// TestSlogAdapter_Creation tests SlogAdapter creation
func TestSlogAdapter_Creation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	adapter := NewSlogAdapter(logger)

	if adapter == nil {
		t.Fatal("Expected non-nil adapter")
	}

	if adapter.logger != logger {
		t.Error("Adapter logger mismatch")
	}
}

// TestSlogAdapter_DebugLevel tests debug-level logging
func TestSlogAdapter_DebugLevel(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Debug(ctx, "debug message", F("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Expected 'debug message' in output, got: %s", output)
	}
	if !strings.Contains(output, "key") {
		t.Errorf("Expected 'key' in output, got: %s", output)
	}
}

// TestSlogAdapter_InfoLevel tests info-level logging
func TestSlogAdapter_InfoLevel(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Info(ctx, "info message", F("count", 42))

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("Expected 'info message' in output, got: %s", output)
	}
	if !strings.Contains(output, "count") {
		t.Errorf("Expected 'count' in output, got: %s", output)
	}
}

// TestSlogAdapter_WarnLevel tests warn-level logging
func TestSlogAdapter_WarnLevel(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Warn(ctx, "warning message", F("retry", 3))

	output := buf.String()
	if !strings.Contains(output, "warning message") {
		t.Errorf("Expected 'warning message' in output, got: %s", output)
	}
	if !strings.Contains(output, "retry") {
		t.Errorf("Expected 'retry' in output, got: %s", output)
	}
}

// TestSlogAdapter_ErrorLevel tests error-level logging
func TestSlogAdapter_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Error(ctx, "error message", F("error_code", 500))

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("Expected 'error message' in output, got: %s", output)
	}
	if !strings.Contains(output, "error_code") {
		t.Errorf("Expected 'error_code' in output, got: %s", output)
	}
}

// TestSlogAdapter_JSONHandler tests JSON output format
func TestSlogAdapter_JSONHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Info(ctx, "test message", 
		F("string_field", "value"),
		F("int_field", 42),
		F("bool_field", true))

	output := buf.String()
	
	// Parse JSON to verify structure
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got: %v", logEntry["msg"])
	}

	if logEntry["string_field"] != "value" {
		t.Errorf("Expected string_field='value', got: %v", logEntry["string_field"])
	}

	// JSON numbers are float64
	if logEntry["int_field"] != float64(42) {
		t.Errorf("Expected int_field=42, got: %v", logEntry["int_field"])
	}

	if logEntry["bool_field"] != true {
		t.Errorf("Expected bool_field=true, got: %v", logEntry["bool_field"])
	}
}

// TestSlogAdapter_MultipleFields tests logging with multiple fields
func TestSlogAdapter_MultipleFields(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Info(ctx, "multi-field test",
		F("field1", "value1"),
		F("field2", 123),
		F("field3", true),
		F("field4", 3.14))

	output := buf.String()
	
	expectedFields := []string{"field1", "field2", "field3", "field4"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output, got: %s", field, output)
		}
	}
}

// TestSlogAdapter_NoFields tests logging without fields
func TestSlogAdapter_NoFields(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Info(ctx, "message without fields")

	output := buf.String()
	if !strings.Contains(output, "message without fields") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

// TestSlogAdapter_LevelFiltering tests that slog respects level filtering
func TestSlogAdapter_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn, // Only WARN and ERROR
	})
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	
	// These should be filtered out
	adapter.Debug(ctx, "debug message")
	adapter.Info(ctx, "info message")
	
	// These should appear
	adapter.Warn(ctx, "warn message")
	adapter.Error(ctx, "error message")

	output := buf.String()
	
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Expected warn message in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Expected error message in output")
	}
}

// TestSlogAdapter_ContextPropagation tests context propagation
func TestSlogAdapter_ContextPropagation(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	// Create context with value
	ctx := context.WithValue(context.Background(), "request_id", "12345")
	
	// Log with context (slog should handle context)
	adapter.Info(ctx, "test with context", F("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "test with context") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

// TestSlogAdapter_WithBuilder tests SlogAdapter integration with Builder
func TestSlogAdapter_WithBuilder(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithLogger(adapter)

	// Verify logger is set
	if builder.logger != adapter {
		t.Error("Expected SlogAdapter to be set as logger")
	}

	// Trigger logging by getting cache stats
	_ = builder.GetCacheStats()

	output := buf.String()
	if !strings.Contains(output, "No cache configured") {
		t.Errorf("Expected cache log message, got: %s", output)
	}
}

// TestSlogAdapter_FieldTypes tests different field value types
func TestSlogAdapter_FieldTypes(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Info(ctx, "type test",
		F("string", "text"),
		F("int", 42),
		F("int64", int64(9999999999)),
		F("float64", 3.14159),
		F("bool", true),
		F("nil", nil))

	output := buf.String()
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify all field types
	if logEntry["string"] != "text" {
		t.Errorf("String field mismatch: %v", logEntry["string"])
	}
	if logEntry["int"] != float64(42) {
		t.Errorf("Int field mismatch: %v", logEntry["int"])
	}
	if logEntry["bool"] != true {
		t.Errorf("Bool field mismatch: %v", logEntry["bool"])
	}
}

// TestSlogAdapter_ConcurrentLogging tests concurrent logging safety
func TestSlogAdapter_ConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	done := make(chan bool)

	// Spawn multiple goroutines logging concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			adapter.Info(ctx, "concurrent log", F("goroutine", id))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify we got logs (exact count may vary due to buffering)
	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected some log output from concurrent logging")
	}
}

// TestSlogAdapter_EmptyMessage tests logging with empty message
func TestSlogAdapter_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	adapter.Info(ctx, "", F("key", "value"))

	output := buf.String()
	// Should still log the field
	if !strings.Contains(output, "key") {
		t.Errorf("Expected field in output even with empty message, got: %s", output)
	}
}

// TestSlogAdapter_LargeFields tests logging with many fields
func TestSlogAdapter_LargeFields(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	adapter := NewSlogAdapter(logger)

	ctx := context.Background()
	
	// Create 50 fields
	fields := make([]Field, 50)
	for i := 0; i < 50; i++ {
		fields[i] = F("field"+string(rune('A'+i%26)), i)
	}

	adapter.Info(ctx, "large field test", fields...)

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON with large fields: %v", err)
	}

	// Verify message is present
	if logEntry["msg"] != "large field test" {
		t.Error("Expected message in large field output")
	}
}
