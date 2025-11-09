package tools

import "context"

// LogFunc is a callback function for logging from tools
// Signature: func(level, msg string, fields map[string]interface{})
type LogFunc func(level, msg string, fields map[string]interface{})

// Global log function (set by Builder via SetLogFunc)
var globalLogFunc LogFunc = func(level, msg string, fields map[string]interface{}) {
	// No-op by default
}

// SetLogFunc sets the global logging function for all built-in tools.
// This is called by Builder to propagate its logger to tools.
//
// Example from Builder:
//
//	tools.SetLogFunc(func(level, msg string, fields map[string]interface{}) {
//	    b.logger.Debug/Info/Warn/Error(ctx, msg, convertFieldsToLogger(fields)...)
//	})
func SetLogFunc(fn LogFunc) {
	if fn != nil {
		globalLogFunc = fn
	}
}

// getContext returns a background context for logging
// Tools don't have access to user context, so we use Background
func getContext() context.Context {
	return context.Background()
}

// logDebug logs a debug-level message
func logDebug(ctx context.Context, msg string, fields map[string]interface{}) {
	globalLogFunc("DEBUG", msg, fields)
}

// logInfo logs an info-level message
func logInfo(ctx context.Context, msg string, fields map[string]interface{}) {
	globalLogFunc("INFO", msg, fields)
}

// logWarn logs a warning-level message
func logWarn(ctx context.Context, msg string, fields map[string]interface{}) {
	globalLogFunc("WARN", msg, fields)
}

// logError logs an error-level message
func logError(ctx context.Context, msg string, fields map[string]interface{}) {
	globalLogFunc("ERROR", msg, fields)
}
