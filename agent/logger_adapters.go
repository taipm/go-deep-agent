package agent

import (
	"context"
	"log/slog"
)

// SlogAdapter adapts the standard library's slog.Logger to the Logger interface.
// This allows using Go 1.21+ structured logging with go-deep-agent.
//
// Example:
//
//	import "log/slog"
//	
//	// Create slog logger with JSON handler
//	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
//	    Level: slog.LevelDebug,
//	})
//	slogLogger := slog.New(handler)
//	
//	// Use with builder
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithLogger(agent.NewSlogAdapter(slogLogger))
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new SlogAdapter that wraps a slog.Logger.
//
// Example:
//
//	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
//	adapter := agent.NewSlogAdapter(slogLogger)
//	builder.WithLogger(adapter)
func NewSlogAdapter(logger *slog.Logger) *SlogAdapter {
	return &SlogAdapter{logger: logger}
}

// Debug logs a debug-level message with structured fields.
func (s *SlogAdapter) Debug(ctx context.Context, msg string, fields ...Field) {
	s.logger.DebugContext(ctx, msg, s.convertFields(fields)...)
}

// Info logs an info-level message with structured fields.
func (s *SlogAdapter) Info(ctx context.Context, msg string, fields ...Field) {
	s.logger.InfoContext(ctx, msg, s.convertFields(fields)...)
}

// Warn logs a warning-level message with structured fields.
func (s *SlogAdapter) Warn(ctx context.Context, msg string, fields ...Field) {
	s.logger.WarnContext(ctx, msg, s.convertFields(fields)...)
}

// Error logs an error-level message with structured fields.
func (s *SlogAdapter) Error(ctx context.Context, msg string, fields ...Field) {
	s.logger.ErrorContext(ctx, msg, s.convertFields(fields)...)
}

// convertFields converts our Field slice to slog.Attr slice
func (s *SlogAdapter) convertFields(fields []Field) []any {
	attrs := make([]any, len(fields))
	for i, field := range fields {
		attrs[i] = slog.Any(field.Key, field.Value)
	}
	return attrs
}
