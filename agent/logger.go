package agent

import (
	"context"
	"fmt"
	"time"
)

// LogLevel defines the severity level for logging
type LogLevel int

const (
	// LogLevelNone disables all logging
	LogLevelNone LogLevel = iota
	// LogLevelError logs only errors
	LogLevelError
	// LogLevelWarn logs warnings and errors
	LogLevelWarn
	// LogLevelInfo logs informational messages, warnings, and errors
	LogLevelInfo
	// LogLevelDebug logs all messages including debug information
	LogLevelDebug
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelNone:
		return "NONE"
	case LogLevelError:
		return "ERROR"
	case LogLevelWarn:
		return "WARN"
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

// Logger defines the interface for structured logging.
// Implementations can integrate with any logging library (slog, zap, logrus, etc.)
//
// Example custom logger:
//
//	type MyLogger struct {
//	    logger *zap.Logger
//	}
//
//	func (l *MyLogger) Debug(ctx context.Context, msg string, fields ...Field) {
//	    l.logger.Debug(msg, fieldsToZap(fields)...)
//	}
type Logger interface {
	// Debug logs a debug-level message with optional structured fields
	Debug(ctx context.Context, msg string, fields ...Field)

	// Info logs an info-level message with optional structured fields
	Info(ctx context.Context, msg string, fields ...Field)

	// Warn logs a warning-level message with optional structured fields
	Warn(ctx context.Context, msg string, fields ...Field)

	// Error logs an error-level message with optional structured fields
	Error(ctx context.Context, msg string, fields ...Field)
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// F creates a new Field (shorthand helper function)
//
// Example:
//
//	logger.Info(ctx, "Request completed",
//	    agent.F("duration_ms", 1234),
//	    agent.F("status", "success"))
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// NoopLogger is a Logger implementation that discards all log messages.
// This is the default logger to ensure zero performance overhead when logging is not configured.
//
// Example:
//
//	// Default behavior - no logging
//	ai := agent.NewOpenAI("gpt-4o-mini", apiKey)
type NoopLogger struct{}

// Debug implements Logger interface (no-op)
func (l *NoopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}

// Info implements Logger interface (no-op)
func (l *NoopLogger) Info(ctx context.Context, msg string, fields ...Field) {}

// Warn implements Logger interface (no-op)
func (l *NoopLogger) Warn(ctx context.Context, msg string, fields ...Field) {}

// Error implements Logger interface (no-op)
func (l *NoopLogger) Error(ctx context.Context, msg string, fields ...Field) {}

// StdLogger implements Logger using Go's standard library (fmt package).
// It outputs human-readable logs to stdout with timestamp and log level.
//
// Example:
//
//	logger := agent.NewStdLogger(agent.LogLevelDebug)
//	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithLogger(logger)
type StdLogger struct {
	Level LogLevel
}

// NewStdLogger creates a new StdLogger with the specified log level
//
// Example:
//
//	// Debug logging (most verbose)
//	logger := agent.NewStdLogger(agent.LogLevelDebug)
//
//	// Info logging (production default)
//	logger := agent.NewStdLogger(agent.LogLevelInfo)
//
//	// Error logging only
//	logger := agent.NewStdLogger(agent.LogLevelError)
func NewStdLogger(level LogLevel) *StdLogger {
	return &StdLogger{Level: level}
}

// Debug logs a debug-level message
func (l *StdLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	if l.Level >= LogLevelDebug {
		l.log("DEBUG", msg, fields)
	}
}

// Info logs an info-level message
func (l *StdLogger) Info(ctx context.Context, msg string, fields ...Field) {
	if l.Level >= LogLevelInfo {
		l.log("INFO", msg, fields)
	}
}

// Warn logs a warning-level message
func (l *StdLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	if l.Level >= LogLevelWarn {
		l.log("WARN", msg, fields)
	}
}

// Error logs an error-level message
func (l *StdLogger) Error(ctx context.Context, msg string, fields ...Field) {
	if l.Level >= LogLevelError {
		l.log("ERROR", msg, fields)
	}
}

// log formats and outputs the log message
func (l *StdLogger) log(level, msg string, fields []Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	output := fmt.Sprintf("[%s] %s: %s", timestamp, level, msg)

	if len(fields) > 0 {
		output += " |"
		for _, f := range fields {
			output += fmt.Sprintf(" %s=%v", f.Key, f.Value)
		}
	}

	fmt.Println(output)
}
