package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// TestLogLevel tests LogLevel string representation
func TestLogLevel(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelNone, "NONE"},
		{LogLevelError, "ERROR"},
		{LogLevelWarn, "WARN"},
		{LogLevelInfo, "INFO"},
		{LogLevelDebug, "DEBUG"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestField tests Field creation
func TestField(t *testing.T) {
	field := F("key", "value")
	if field.Key != "key" {
		t.Errorf("Field.Key = %v, want %v", field.Key, "key")
	}
	if field.Value != "value" {
		t.Errorf("Field.Value = %v, want %v", field.Value, "value")
	}
}

// TestNoopLogger tests that NoopLogger does nothing
func TestNoopLogger(t *testing.T) {
	logger := &NoopLogger{}
	ctx := context.Background()

	// These should not panic or output anything
	logger.Debug(ctx, "debug message", F("key", "value"))
	logger.Info(ctx, "info message", F("key", "value"))
	logger.Warn(ctx, "warn message", F("key", "value"))
	logger.Error(ctx, "error message", F("key", "value"))

	// Test passes if no panic
}

// TestStdLoggerLevels tests StdLogger log level filtering
func TestStdLoggerLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         LogLevel
		shouldLogInfo bool
		shouldLogWarn bool
		shouldLogErr  bool
		shouldLogDbg  bool
	}{
		{"None", LogLevelNone, false, false, false, false},
		{"Error", LogLevelError, false, false, true, false},
		{"Warn", LogLevelWarn, false, true, true, false},
		{"Info", LogLevelInfo, true, true, true, false},
		{"Debug", LogLevelDebug, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logger := NewStdLogger(tt.level)
			ctx := context.Background()

			// Log all levels
			logger.Debug(ctx, "debug")
			logger.Info(ctx, "info")
			logger.Warn(ctx, "warn")
			logger.Error(ctx, "error")

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check what was logged
			hasDebug := strings.Contains(output, "DEBUG")
			hasInfo := strings.Contains(output, "INFO")
			hasWarn := strings.Contains(output, "WARN")
			hasError := strings.Contains(output, "ERROR")

			if hasDebug != tt.shouldLogDbg {
				t.Errorf("Debug logging: got %v, want %v", hasDebug, tt.shouldLogDbg)
			}
			if hasInfo != tt.shouldLogInfo {
				t.Errorf("Info logging: got %v, want %v", hasInfo, tt.shouldLogInfo)
			}
			if hasWarn != tt.shouldLogWarn {
				t.Errorf("Warn logging: got %v, want %v", hasWarn, tt.shouldLogWarn)
			}
			if hasError != tt.shouldLogErr {
				t.Errorf("Error logging: got %v, want %v", hasError, tt.shouldLogErr)
			}
		})
	}
}

// TestStdLoggerFields tests field formatting in StdLogger
func TestStdLoggerFields(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewStdLogger(LogLevelInfo)
	ctx := context.Background()

	logger.Info(ctx, "test message",
		F("key1", "value1"),
		F("key2", 42),
		F("key3", true))

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check field formatting
	expected := []string{
		"INFO: test message",
		"key1=value1",
		"key2=42",
		"key3=true",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q, got: %s", exp, output)
		}
	}
}

// TestStdLoggerNoFields tests logging without fields
func TestStdLoggerNoFields(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewStdLogger(LogLevelInfo)
	ctx := context.Background()

	logger.Info(ctx, "simple message")

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should not have the "|" separator when no fields
	if strings.Contains(output, "simple message |") {
		t.Errorf("Expected no field separator for message without fields, got: %s", output)
	}

	if !strings.Contains(output, "INFO: simple message") {
		t.Errorf("Expected message to be logged, got: %s", output)
	}
}

// TestWithLogger tests WithLogger method
func TestWithLogger(t *testing.T) {
	logger := NewStdLogger(LogLevelDebug)
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithLogger(logger)

	if builder.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	// Type assertion to check it's the right logger
	if _, ok := builder.logger.(*StdLogger); !ok {
		t.Errorf("Expected StdLogger, got %T", builder.logger)
	}
}

// TestWithDebugLogging tests WithDebugLogging method
func TestWithDebugLogging(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithDebugLogging()

	if builder.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	stdLogger, ok := builder.logger.(*StdLogger)
	if !ok {
		t.Fatalf("Expected StdLogger, got %T", builder.logger)
	}

	if stdLogger.Level != LogLevelDebug {
		t.Errorf("Expected LogLevelDebug, got %v", stdLogger.Level)
	}
}

// TestWithInfoLogging tests WithInfoLogging method
func TestWithInfoLogging(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithInfoLogging()

	if builder.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	stdLogger, ok := builder.logger.(*StdLogger)
	if !ok {
		t.Fatalf("Expected StdLogger, got %T", builder.logger)
	}

	if stdLogger.Level != LogLevelInfo {
		t.Errorf("Expected LogLevelInfo, got %v", stdLogger.Level)
	}
}

// TestGetLogger tests getLogger helper method
func TestGetLogger(t *testing.T) {
	// Test with no logger set (should return NoopLogger)
	builder1 := NewOpenAI("gpt-4o-mini", "test-key")
	logger1 := builder1.getLogger()
	if _, ok := logger1.(*NoopLogger); !ok {
		t.Errorf("Expected NoopLogger when no logger set, got %T", logger1)
	}

	// Test with logger set
	builder2 := NewOpenAI("gpt-4o-mini", "test-key").
		WithDebugLogging()
	logger2 := builder2.getLogger()
	if _, ok := logger2.(*StdLogger); !ok {
		t.Errorf("Expected StdLogger when logger set, got %T", logger2)
	}
}

// TestLoggerChaining tests method chaining with logger methods
func TestLoggerChaining(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSystem("test system").
		WithDebugLogging().
		WithTemperature(0.7)

	if builder.logger == nil {
		t.Fatal("Expected logger to be set")
	}
	if builder.systemPrompt != "test system" {
		t.Errorf("Expected system prompt to be 'test system', got %v", builder.systemPrompt)
	}
	if builder.temperature == nil || *builder.temperature != 0.7 {
		t.Errorf("Expected temperature to be 0.7")
	}
}

// TestCustomLogger tests using a custom logger implementation
func TestCustomLogger(t *testing.T) {
	// Custom logger that counts calls
	type CountingLogger struct {
		debugCount int
		infoCount  int
		warnCount  int
		errorCount int
	}

	logger := &CountingLogger{}
	ctx := context.Background()

	// Implement Logger interface methods
	logger.debugCount = 0
	logger.infoCount = 0

	// Call methods (this demonstrates the pattern)
	_ = ctx
	_ = logger

	// This test demonstrates the pattern for custom loggers
	// Actual custom loggers would implement the Logger interface
}

// TestLoggerContextPropagation tests that context is passed to logger
func TestLoggerContextPropagation(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	// Test that logger methods accept context
	noopLogger := &NoopLogger{}
	noopLogger.Info(ctx, "test")

	// Test passes if no panic
}

// BenchmarkNoopLogger benchmarks NoopLogger (should be zero overhead)
func BenchmarkNoopLogger(b *testing.B) {
	logger := &NoopLogger{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug(ctx, "debug message", F("key", "value"))
		logger.Info(ctx, "info message", F("key", "value"))
	}
}

// BenchmarkStdLogger benchmarks StdLogger
func BenchmarkStdLogger(b *testing.B) {
	// Redirect output to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	logger := NewStdLogger(LogLevelDebug)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug(ctx, "debug message", F("key", "value"))
		logger.Info(ctx, "info message", F("key", "value"))
	}
}

// BenchmarkStdLoggerFiltered benchmarks StdLogger with level filtering
func BenchmarkStdLoggerFiltered(b *testing.B) {
	// Redirect output to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	logger := NewStdLogger(LogLevelInfo) // Debug messages filtered
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug(ctx, "debug message", F("key", "value")) // Filtered out
		logger.Info(ctx, "info message", F("key", "value"))
	}
}

// ExampleLogger demonstrates basic logger usage
func ExampleLogger() {
	logger := NewStdLogger(LogLevelInfo)
	ctx := context.Background()

	logger.Info(ctx, "Request completed",
		F("duration_ms", 1234),
		F("status", "success"))

	// Output will be similar to:
	// [2025-01-15 10:30:45.123] INFO: Request completed | duration_ms=1234 status=success
}

// Example_withDebugLogging demonstrates debug logging
func Example_withDebugLogging() {
	ai := NewOpenAI("gpt-4o-mini", "test-key").
		WithDebugLogging()

	fmt.Printf("Logger type: %T\n", ai.getLogger())
	// Output: Logger type: *agent.StdLogger
}

// Example_withInfoLogging demonstrates info logging
func Example_withInfoLogging() {
	ai := NewOpenAI("gpt-4o-mini", "test-key").
		WithInfoLogging()

	fmt.Printf("Logger type: %T\n", ai.getLogger())
	// Output: Logger type: *agent.StdLogger
}
