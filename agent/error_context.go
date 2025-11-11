package agent

import (
	"fmt"
)

// ErrorContext provides rich context for errors to aid debugging and monitoring.
// Use this to wrap errors with additional information about what was being attempted.
type ErrorContext struct {
	Operation string                 // What operation was being performed
	Details   map[string]interface{} // Additional context (model, tool name, etc.)
	Err       error                  // The underlying error
}

// Error implements the error interface.
func (e *ErrorContext) Error() string {
	if len(e.Details) == 0 {
		return fmt.Sprintf("%s: %v", e.Operation, e.Err)
	}

	// Build detailed error message
	msg := fmt.Sprintf("%s: %v\nContext:", e.Operation, e.Err)
	for k, v := range e.Details {
		msg += fmt.Sprintf("\n  - %s: %v", k, v)
	}
	return msg
}

// Unwrap returns the underlying error for errors.Is/As compatibility.
func (e *ErrorContext) Unwrap() error {
	return e.Err
}

// WithContext wraps an error with contextual information.
// This helps with debugging by providing information about what was being attempted.
//
// Example:
//
//	err := agent.WithContext(err, "API request", map[string]interface{}{
//	    "model": "gpt-4o-mini",
//	    "tokens": 150,
//	    "retry_attempt": 2,
//	})
func WithContext(err error, operation string, details map[string]interface{}) error {
	if err == nil {
		return nil
	}
	return &ErrorContext{
		Operation: operation,
		Details:   details,
		Err:       err,
	}
}

// WithSimpleContext wraps an error with a simple operation description.
//
// Example:
//
//	err := agent.WithSimpleContext(err, "parsing tool arguments")
func WithSimpleContext(err error, operation string) error {
	if err == nil {
		return nil
	}
	return &ErrorContext{
		Operation: operation,
		Details:   nil,
		Err:       err,
	}
}

// GetErrorContext extracts ErrorContext from an error if present.
// Returns nil if the error is not an ErrorContext.
func GetErrorContext(err error) *ErrorContext {
	if errCtx, ok := err.(*ErrorContext); ok {
		return errCtx
	}
	return nil
}

// IsErrorContext checks if an error is an ErrorContext.
func IsErrorContext(err error) bool {
	_, ok := err.(*ErrorContext)
	return ok
}

// LogFields converts ErrorContext to structured log fields.
// This enables seamless integration with structured logging libraries (slog, zap, logrus, etc.)
//
// Example with agent.Logger:
//
//	if err != nil {
//	    logger.Error(ctx, "Operation failed", errCtx.LogFields()...)
//	}
//
// Example with slog:
//
//	if errCtx := agent.GetErrorContext(err); errCtx != nil {
//	    slog.Error("Operation failed", errCtx.LogFields()...)
//	}
func (e *ErrorContext) LogFields() []Field {
	fields := []Field{
		{Key: "error", Value: e.Err.Error()},
		{Key: "operation", Value: e.Operation},
	}

	// Add all details as separate fields
	for k, v := range e.Details {
		fields = append(fields, Field{Key: k, Value: v})
	}

	return fields
}

// ErrorChain represents a chain of errors with context at each level.
// Useful for tracking errors through multiple layers of the application.
type ErrorChain struct {
	errors []error
}

// NewErrorChain creates a new error chain.
func NewErrorChain() *ErrorChain {
	return &ErrorChain{
		errors: make([]error, 0),
	}
}

// Add adds an error to the chain with context.
func (ec *ErrorChain) Add(err error, operation string, details map[string]interface{}) *ErrorChain {
	if err != nil {
		contextErr := WithContext(err, operation, details)
		ec.errors = append(ec.errors, contextErr)
	}
	return ec
}

// AddSimple adds an error to the chain with simple context.
func (ec *ErrorChain) AddSimple(err error, operation string) *ErrorChain {
	if err != nil {
		contextErr := WithSimpleContext(err, operation)
		ec.errors = append(ec.errors, contextErr)
	}
	return ec
}

// Error returns a formatted string of all errors in the chain.
func (ec *ErrorChain) Error() string {
	if len(ec.errors) == 0 {
		return "no errors"
	}

	if len(ec.errors) == 1 {
		return ec.errors[0].Error()
	}

	msg := fmt.Sprintf("error chain (%d errors):", len(ec.errors))
	for i, err := range ec.errors {
		msg += fmt.Sprintf("\n%d. %v", i+1, err)
	}
	return msg
}

// First returns the first error in the chain, or nil if empty.
func (ec *ErrorChain) First() error {
	if len(ec.errors) == 0 {
		return nil
	}
	return ec.errors[0]
}

// Last returns the last error in the chain, or nil if empty.
func (ec *ErrorChain) Last() error {
	if len(ec.errors) == 0 {
		return nil
	}
	return ec.errors[len(ec.errors)-1]
}

// HasErrors returns true if the chain contains any errors.
func (ec *ErrorChain) HasErrors() bool {
	return len(ec.errors) > 0
}

// Count returns the number of errors in the chain.
func (ec *ErrorChain) Count() int {
	return len(ec.errors)
}

// All returns all errors in the chain.
func (ec *ErrorChain) All() []error {
	return ec.errors
}

// ErrorSummary provides a high-level summary of an error for logging/monitoring.
type ErrorSummary struct {
	Type      string                 // Error type (CodedError, PanicError, etc.)
	Code      string                 // Error code if available
	Message   string                 // Error message
	Retryable bool                   // Whether the error is retryable
	Context   map[string]interface{} // Additional context
}

// SummarizeError creates a summary of an error for structured logging/monitoring.
// This is useful for sending errors to monitoring systems (DataDog, New Relic, etc.)
//
// Example:
//
//	summary := agent.SummarizeError(err)
//	log.Printf("[ERROR] type=%s code=%s retryable=%v msg=%s",
//	    summary.Type, summary.Code, summary.Retryable, summary.Message)
func SummarizeError(err error) *ErrorSummary {
	if err == nil {
		return nil
	}

	summary := &ErrorSummary{
		Type:      "error",
		Code:      "",
		Message:   err.Error(),
		Retryable: false,
		Context:   make(map[string]interface{}),
	}

	// Check for PanicError
	if IsPanicError(err) {
		summary.Type = "PanicError"
		summary.Retryable = false
		panicValue := GetPanicValue(err)
		if panicValue != nil {
			summary.Context["panic_value"] = panicValue
		}
		stackTrace := GetStackTrace(err)
		if stackTrace != "" {
			// Store first 500 chars of stack trace
			if len(stackTrace) > 500 {
				summary.Context["stack_trace"] = stackTrace[:500] + "..."
			} else {
				summary.Context["stack_trace"] = stackTrace
			}
		}
		return summary
	}

	// Check for CodedError
	if IsCodedError(err) {
		summary.Type = "CodedError"
		summary.Code = GetErrorCode(err)
		summary.Retryable = IsRetryable(err)
		return summary
	}

	// Check for ErrorContext
	if errCtx := GetErrorContext(err); errCtx != nil {
		summary.Type = "ErrorContext"
		summary.Context["operation"] = errCtx.Operation
		if errCtx.Details != nil {
			for k, v := range errCtx.Details {
				summary.Context[k] = v
			}
		}
		// Recursively summarize underlying error
		if underlying := SummarizeError(errCtx.Err); underlying != nil {
			summary.Code = underlying.Code
			summary.Retryable = underlying.Retryable
			if underlying.Type != "error" {
				summary.Type = underlying.Type
			}
		}
		return summary
	}

	// Check for common sentinel errors
	switch err {
	case ErrAPIKey:
		summary.Type = "SentinelError"
		summary.Code = "API_KEY_MISSING"
		summary.Retryable = false
	case ErrRateLimit:
		summary.Type = "SentinelError"
		summary.Code = "RATE_LIMIT_EXCEEDED"
		summary.Retryable = true
	case ErrTimeout:
		summary.Type = "SentinelError"
		summary.Code = "REQUEST_TIMEOUT"
		summary.Retryable = true
	case ErrToolExecution:
		summary.Type = "SentinelError"
		summary.Code = "TOOL_EXECUTION_FAILED"
		summary.Retryable = false
	}

	return summary
}

// ExtractLogFields extracts structured log fields from any error type.
// This is the universal helper that works with all error types in go-deep-agent.
// It intelligently detects the error type and returns appropriate fields.
//
// Supported error types:
//   - *CodedError: Returns error_code, error_message, retryable
//   - *PanicError: Returns error_type, panic_value, stack_trace
//   - *ErrorContext: Returns operation, all details, and underlying error fields
//   - Other errors: Returns basic error field
//
// Example with agent.Logger:
//
//	if err != nil {
//	    fields := agent.ExtractLogFields(err)
//	    logger.Error(ctx, "Operation failed", fields...)
//	}
//
// Example with slog:
//
//	import "log/slog"
//	if err != nil {
//	    fields := agent.ExtractLogFields(err)
//	    slog.Error("Operation failed", toSlogAttrs(fields)...)
//	}
//
// Example with zap:
//
//	import "go.uber.org/zap"
//	if err != nil {
//	    fields := agent.ExtractLogFields(err)
//	    logger.Error("Operation failed", toZapFields(fields)...)
//	}
func ExtractLogFields(err error) []Field {
	if err == nil {
		return []Field{}
	}

	// Check for PanicError first (most specific)
	if panicErr, ok := err.(*PanicError); ok {
		return panicErr.LogFields()
	}

	// Check for CodedError
	if codedErr, ok := err.(*CodedError); ok {
		return codedErr.LogFields()
	}

	// Check for ErrorContext
	if errCtx, ok := err.(*ErrorContext); ok {
		return errCtx.LogFields()
	}

	// Fallback: basic error field
	return []Field{
		{Key: "error", Value: err.Error()},
	}
}

// ExtractLogFieldsWithSummary extracts log fields and adds error summary information.
// This is useful for monitoring systems that need categorization and retry information.
//
// Example:
//
//	if err != nil {
//	    fields := agent.ExtractLogFieldsWithSummary(err)
//	    logger.Error(ctx, "Request failed", fields...)
//	    // Fields include: error, error_type, error_code, retryable, etc.
//	}
func ExtractLogFieldsWithSummary(err error) []Field {
	if err == nil {
		return []Field{}
	}

	// Start with base fields
	fields := ExtractLogFields(err)

	// Add summary information
	summary := SummarizeError(err)
	if summary != nil {
		fields = append(fields, Field{Key: "error_type", Value: summary.Type})
		if summary.Code != "" {
			fields = append(fields, Field{Key: "error_code", Value: summary.Code})
		}
		fields = append(fields, Field{Key: "retryable", Value: summary.Retryable})

		// Add context from summary if available
		for k, v := range summary.Context {
			// Avoid duplicates (base fields take precedence)
			isDuplicate := false
			for _, f := range fields {
				if f.Key == k {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				fields = append(fields, Field{Key: k, Value: v})
			}
		}
	}

	return fields
}
