package agent

import (
	"fmt"
)

// Error codes for programmatic error handling
// Only includes the 20 most common error types (LEAN approach: 80/20 rule)
const (
	// API Errors (1xxx) - Most common user-facing errors
	ErrCodeAPIKeyMissing     = "API_KEY_MISSING"
	ErrCodeAPIKeyInvalid     = "API_KEY_INVALID"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeRequestTimeout    = "REQUEST_TIMEOUT"
	ErrCodeInvalidResponse   = "INVALID_RESPONSE"
	ErrCodeContentRefused    = "CONTENT_REFUSED"

	// Tool Errors (2xxx) - Tool execution issues
	ErrCodeToolNotFound        = "TOOL_NOT_FOUND"
	ErrCodeToolExecutionFailed = "TOOL_EXECUTION_FAILED"
	ErrCodeToolPanicked        = "TOOL_PANICKED"
	ErrCodeToolTimeout         = "TOOL_TIMEOUT"

	// RAG/Vector Store Errors (3xxx) - Vector and embedding issues
	ErrCodeVectorStoreNotConfigured = "VECTOR_STORE_NOT_CONFIGURED"
	ErrCodeEmbeddingFailed          = "EMBEDDING_GENERATION_FAILED"
	ErrCodeVectorSearchFailed       = "VECTOR_SEARCH_FAILED"

	// Memory Errors (4xxx) - Conversation memory issues
	ErrCodeMemoryFull = "MEMORY_CAPACITY_FULL"

	// Cache Errors (5xxx) - Caching issues
	ErrCodeCacheConnectionFailed = "CACHE_CONNECTION_FAILED"
	ErrCodeCacheOperationFailed  = "CACHE_OPERATION_FAILED"

	// Configuration Errors (6xxx) - Setup and configuration
	ErrCodeInvalidConfiguration = "INVALID_CONFIGURATION"
	ErrCodeUnsupportedProvider  = "UNSUPPORTED_PROVIDER"

	// Retry Errors (7xxx) - Retry exhaustion
	ErrCodeMaxRetriesExceeded = "MAX_RETRIES_EXCEEDED"

	// Completion Errors (8xxx) - Model response issues
	ErrCodeNoResponseChoices = "NO_RESPONSE_CHOICES"
)

// CodedError provides error codes for programmatic handling
// Simple, lightweight struct - no over-engineering
type CodedError struct {
	Code    string // Error code (e.g., "RATE_LIMIT_EXCEEDED")
	Message string // Human-readable error message
	Err     error  // Underlying error (optional)
}

// Error implements the error interface
func (e *CodedError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As compatibility
func (e *CodedError) Unwrap() error {
	return e.Err
}

// NewCodedError creates a new error with code and message
func NewCodedError(code, message string, err error) *CodedError {
	return &CodedError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Helper functions to create common coded errors

// NewAPIKeyError creates an API key error
func NewAPIKeyError(err error) *CodedError {
	return NewCodedError(ErrCodeAPIKeyMissing, "API key is missing or invalid", err)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(err error) *CodedError {
	return NewCodedError(ErrCodeRateLimitExceeded, "Rate limit exceeded", err)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(err error) *CodedError {
	return NewCodedError(ErrCodeRequestTimeout, "Request timeout", err)
}

// NewToolError creates a tool execution error
func NewToolError(toolName string, err error) *CodedError {
	return NewCodedError(
		ErrCodeToolExecutionFailed,
		fmt.Sprintf("Tool '%s' execution failed", toolName),
		err,
	)
}

// NewToolPanicError creates a tool panic error
func NewToolPanicError(toolName string, panicValue interface{}) *CodedError {
	return NewCodedError(
		ErrCodeToolPanicked,
		fmt.Sprintf("Tool '%s' panicked: %v", toolName, panicValue),
		nil,
	)
}

// NewVectorStoreConfigError creates a vector store configuration error
func NewVectorStoreConfigError(operation string, err error) *CodedError {
	return NewCodedError(
		ErrCodeVectorStoreNotConfigured,
		fmt.Sprintf("Vector store operation '%s' failed", operation),
		err,
	)
}

// NewEmbeddingError creates an embedding generation error
func NewEmbeddingError(err error) *CodedError {
	return NewCodedError(ErrCodeEmbeddingFailed, "Embedding generation failed", err)
}

// NewCacheError creates a cache operation error
func NewCacheError(operation string, err error) *CodedError {
	return NewCodedError(
		ErrCodeCacheOperationFailed,
		fmt.Sprintf("Cache operation '%s' failed", operation),
		err,
	)
}

// Error checking helpers - check if error has specific code

// IsCodedError checks if error is a CodedError
func IsCodedError(err error) bool {
	_, ok := err.(*CodedError)
	return ok
}

// HasErrorCode checks if error has specific error code
func HasErrorCode(err error, code string) bool {
	if codedErr, ok := err.(*CodedError); ok {
		return codedErr.Code == code
	}
	return false
}

// GetErrorCode extracts error code from error, returns empty string if not a CodedError
func GetErrorCode(err error) string {
	if codedErr, ok := err.(*CodedError); ok {
		return codedErr.Code
	}
	return ""
}

// IsRetryable checks if error is retryable based on error code
func IsRetryable(err error) bool {
	code := GetErrorCode(err)
	switch code {
	case ErrCodeRateLimitExceeded,
		ErrCodeRequestTimeout,
		ErrCodeInvalidResponse,
		ErrCodeCacheOperationFailed:
		return true
	default:
		return false
	}
}

// LogFields converts CodedError to structured log fields.
// This enables seamless integration with structured logging libraries.
//
// Example:
//
//	if codedErr, ok := err.(*agent.CodedError); ok {
//	    logger.Error(ctx, "Request failed", codedErr.LogFields()...)
//	}
func (e *CodedError) LogFields() []Field {
	fields := []Field{
		{Key: "error_code", Value: e.Code},
		{Key: "error_message", Value: e.Message},
		{Key: "retryable", Value: IsRetryable(e)},
	}

	// Add underlying error if present
	if e.Err != nil {
		fields = append(fields, Field{Key: "underlying_error", Value: e.Err.Error()})
	}

	return fields
}
