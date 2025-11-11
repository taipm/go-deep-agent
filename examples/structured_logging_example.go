// Package main demonstrates structured logging integration with go-deep-agent errors.
//
// This example shows how to use the new LogFields() methods and ExtractLogFields()
// function to integrate error handling with structured logging libraries.
//
// Supported integrations:
//   - agent.Logger (built-in)
//   - log/slog (Go 1.21+)
//   - go.uber.org/zap
//   - github.com/sirupsen/logrus
//   - Any structured logger
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Structured Logging Integration Examples ===\n")

	// Example 1: Using with agent.Logger
	example1AgentLogger()

	// Example 2: Using with slog (Go standard library)
	example2Slog()

	// Example 3: ExtractLogFields universal helper
	example3UniversalHelper()

	// Example 4: Production error handling pattern
	example4ProductionPattern()

	fmt.Println("\n✅ All examples completed!")
}

// Example 1: Using LogFields() with agent.Logger
func example1AgentLogger() {
	fmt.Println("Example 1: Agent Logger Integration")

	logger := agent.NewStdLogger(agent.LogLevelInfo)
	ctx := context.Background()

	// Simulate an error
	baseErr := errors.New("database connection timeout")
	codedErr := agent.NewTimeoutError(baseErr)
	contextErr := agent.WithContext(codedErr, "user query", map[string]interface{}{
		"user_id":  12345,
		"query":    "SELECT * FROM orders",
		"duration": 5000,
	})

	// Log with structured fields
	if errCtx := agent.GetErrorContext(contextErr); errCtx != nil {
		logger.Error(ctx, "Query failed", errCtx.LogFields()...)
	}

	fmt.Println("  ✓ Logged error with structured fields\n")
}

// Example 2: Using with slog (Go 1.21+ standard library)
func example2Slog() {
	fmt.Println("Example 2: slog Integration")

	// Create slog logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a panic error
	panicErr := &agent.PanicError{
		Value:      "nil pointer dereference",
		StackTrace: "goroutine 1 [running]:\nmain.processRequest()\n\t/app/main.go:42",
	}

	// Convert to slog attributes
	fields := panicErr.LogFields()
	attrs := make([]slog.Attr, len(fields))
	for i, f := range fields {
		attrs[i] = slog.Any(f.Key, f.Value)
	}

	logger.LogAttrs(context.Background(), slog.LevelError, "Tool panic recovered", attrs...)
	fmt.Println("  ✓ Logged panic error with slog\n")
}

// Example 3: Universal helper for any error type
func example3UniversalHelper() {
	fmt.Println("Example 3: ExtractLogFields Universal Helper")

	logger := agent.NewStdLogger(agent.LogLevelDebug)
	ctx := context.Background()

	// Different error types
	errors := []error{
		agent.NewAPIKeyError(errors.New("missing key")),
		agent.NewRateLimitError(errors.New("429 Too Many Requests")),
		&agent.PanicError{Value: "array index out of range", StackTrace: "..."},
		agent.WithContext(errors.New("timeout"), "API call", map[string]interface{}{
			"endpoint": "/api/users",
			"timeout":  30,
		}),
	}

	// Log all errors with ExtractLogFields
	for i, err := range errors {
		fields := agent.ExtractLogFields(err)
		logger.Error(ctx, fmt.Sprintf("Error %d", i+1), fields...)
	}

	fmt.Println("  ✓ Logged multiple error types with universal helper\n")
}

// Example 4: Production error handling pattern
func example4ProductionPattern() {
	fmt.Println("Example 4: Production Error Handling Pattern")

	// Setup logger
	logger := agent.NewStdLogger(agent.LogLevelInfo)
	ctx := context.Background()

	// Simulate production error
	err := simulateProductionError()

	if err != nil {
		// Use ExtractLogFieldsWithSummary for complete error information
		fields := agent.ExtractLogFieldsWithSummary(err)
		logger.Error(ctx, "Production error occurred", fields...)

		// Make retry decision based on error
		if agent.IsRetryable(err) {
			fmt.Println("  → Retrying operation...")
		} else {
			fmt.Println("  → Non-retryable error, failing...")
		}
	}

	fmt.Println("  ✓ Production pattern demonstrated\n")
}

// simulateProductionError creates a realistic nested error
func simulateProductionError() error {
	// Layer 1: Base error
	baseErr := errors.New("network unreachable")

	// Layer 2: Coded error
	codedErr := agent.NewCodedError(
		agent.ErrCodeRequestTimeout,
		"API request timeout",
		baseErr,
	)

	// Layer 3: Error context
	contextErr := agent.WithContext(codedErr, "fetch user profile", map[string]interface{}{
		"user_id":       "usr_123456",
		"endpoint":      "https://api.example.com/v1/users/123456",
		"retry_count":   3,
		"timeout_ms":    5000,
		"correlation_id": "req_abc123",
	})

	return contextErr
}

// Helper function to convert agent.Field to slog.Attr
func fieldsToSlogAttrs(fields []agent.Field) []slog.Attr {
	attrs := make([]slog.Attr, len(fields))
	for i, f := range fields {
		attrs[i] = slog.Any(f.Key, f.Value)
	}
	return attrs
}
