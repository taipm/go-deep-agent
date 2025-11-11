package agent

import (
	"context"
	"fmt"
	"runtime/debug"
)

// PanicError represents an error that was recovered from a panic.
// It includes the panic value and stack trace for debugging.
type PanicError struct {
	Value      interface{} // The value passed to panic()
	StackTrace string      // Full stack trace at panic point
}

// Error implements the error interface.
func (e *PanicError) Error() string {
	return fmt.Sprintf("panic recovered: %v", e.Value)
}

// Unwrap returns nil (PanicError doesn't wrap another error).
func (e *PanicError) Unwrap() error {
	return nil
}

// recoverPanic is a helper function that recovers from panics and converts them to errors.
// It should be called with defer at the beginning of functions that need panic recovery.
//
// Usage:
//
//	func someFunction() (err error) {
//	    defer recoverPanic(&err, "someFunction")
//	    // ... function body that might panic ...
//	}
func recoverPanic(errPtr *error, contextStr string) {
	if r := recover(); r != nil {
		// Capture stack trace
		stackTrace := string(debug.Stack())

		// Create PanicError
		panicErr := &PanicError{
			Value:      r,
			StackTrace: stackTrace,
		}

		// Set the error pointer
		*errPtr = panicErr

		// Log panic if debug logger is available
		// Note: We can't access builder here, so logging happens at call site
	}
}

// recoverPanicWithLogger is like recoverPanic but also logs the panic.
func recoverPanicWithLogger(errPtr *error, contextStr string, logger Logger) {
	if r := recover(); r != nil {
		// Capture stack trace
		stackTrace := string(debug.Stack())

		// Create PanicError
		panicErr := &PanicError{
			Value:      r,
			StackTrace: stackTrace,
		}

		// Set the error pointer
		*errPtr = panicErr

		// Log panic with stack trace
		if logger != nil {
			logger.Error(context.Background(), fmt.Sprintf("PANIC RECOVERED in %s: %v\nStack trace:\n%s",
				contextStr, r, stackTrace))
		}
	}
}

// safeExecute wraps a function call with panic recovery.
// Returns the function result and any error (including recovered panics).
//
// Example:
//
//	result, err := safeExecute("tool execution", func() (string, error) {
//	    return tool.Handler(args)
//	})
func safeExecute(contextStr string, fn func() (string, error)) (result string, err error) {
	defer recoverPanic(&err, contextStr)
	return fn()
}

// safeExecuteWithLogger wraps a function call with panic recovery and logging.
func safeExecuteWithLogger(contextStr string, logger Logger, fn func() (string, error)) (result string, err error) {
	defer recoverPanicWithLogger(&err, contextStr, logger)
	return fn()
}

// safeExecuteVoid wraps a void function call with panic recovery.
// Returns any error from the function or from recovered panics.
//
// Example:
//
//	err := safeExecuteVoid("callback", func() error {
//	    return callback()
//	})
func safeExecuteVoid(contextStr string, fn func() error) (err error) {
	defer recoverPanic(&err, contextStr)
	return fn()
}

// safeExecuteVoidWithLogger wraps a void function call with panic recovery and logging.
func safeExecuteVoidWithLogger(contextStr string, logger Logger, fn func() error) (err error) {
	defer recoverPanicWithLogger(&err, contextStr, logger)
	return fn()
}

// IsPanicError checks if an error is a PanicError.
func IsPanicError(err error) bool {
	_, ok := err.(*PanicError)
	return ok
}

// GetPanicValue extracts the panic value from a PanicError.
// Returns nil if the error is not a PanicError.
func GetPanicValue(err error) interface{} {
	if panicErr, ok := err.(*PanicError); ok {
		return panicErr.Value
	}
	return nil
}

// GetStackTrace extracts the stack trace from a PanicError.
// Returns empty string if the error is not a PanicError.
func GetStackTrace(err error) string {
	if panicErr, ok := err.(*PanicError); ok {
		return panicErr.StackTrace
	}
	return ""
}

// LogFields converts PanicError to structured log fields.
// This enables seamless integration with structured logging libraries.
//
// Example:
//
//	if panicErr, ok := err.(*agent.PanicError); ok {
//	    logger.Error(ctx, "Panic recovered", panicErr.LogFields()...)
//	}
func (e *PanicError) LogFields() []Field {
	fields := []Field{
		{Key: "error_type", Value: "panic"},
		{Key: "panic_value", Value: fmt.Sprintf("%v", e.Value)},
	}

	// Add truncated stack trace (first 500 chars for readability)
	if len(e.StackTrace) > 500 {
		fields = append(fields, Field{
			Key:   "stack_trace",
			Value: e.StackTrace[:500] + "...",
		})
		fields = append(fields, Field{
			Key:   "stack_trace_full_length",
			Value: len(e.StackTrace),
		})
	} else if e.StackTrace != "" {
		fields = append(fields, Field{
			Key:   "stack_trace",
			Value: e.StackTrace,
		})
	}

	return fields
}
