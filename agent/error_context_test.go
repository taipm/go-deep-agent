package agent

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorContext_Error(t *testing.T) {
	baseErr := errors.New("connection failed")

	// Simple context
	err := WithSimpleContext(baseErr, "database query")
	expected := "database query: connection failed"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}

	// With details
	err = WithContext(baseErr, "API request", map[string]interface{}{
		"model":  "gpt-4o-mini",
		"tokens": 150,
	})
	msg := err.Error()
	if !strings.Contains(msg, "API request") {
		t.Error("Expected message to contain operation")
	}
	if !strings.Contains(msg, "connection failed") {
		t.Error("Expected message to contain base error")
	}
	if !strings.Contains(msg, "model") {
		t.Error("Expected message to contain context details")
	}
}

func TestErrorContext_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := WithSimpleContext(baseErr, "operation")

	if unwrapped := errors.Unwrap(wrapped); unwrapped != baseErr {
		t.Errorf("Expected unwrap to return base error, got %v", unwrapped)
	}

	// Test errors.Is
	if !errors.Is(wrapped, baseErr) {
		t.Error("Expected errors.Is to work with wrapped error")
	}
}

func TestWithContext_NilError(t *testing.T) {
	err := WithContext(nil, "operation", nil)
	if err != nil {
		t.Errorf("Expected nil for nil error, got %v", err)
	}

	err = WithSimpleContext(nil, "operation")
	if err != nil {
		t.Errorf("Expected nil for nil error, got %v", err)
	}
}

func TestGetErrorContext(t *testing.T) {
	baseErr := errors.New("test error")
	contextErr := WithContext(baseErr, "test op", map[string]interface{}{
		"key": "value",
	})

	// Should extract context
	ctx := GetErrorContext(contextErr)
	if ctx == nil {
		t.Fatal("Expected to extract error context")
	}
	if ctx.Operation != "test op" {
		t.Errorf("Expected operation 'test op', got %q", ctx.Operation)
	}
	if ctx.Details["key"] != "value" {
		t.Error("Expected to preserve details")
	}

	// Should return nil for regular error
	ctx = GetErrorContext(baseErr)
	if ctx != nil {
		t.Error("Expected nil for non-context error")
	}
}

func TestIsErrorContext(t *testing.T) {
	contextErr := WithSimpleContext(errors.New("test"), "op")
	regularErr := errors.New("test")

	if !IsErrorContext(contextErr) {
		t.Error("Expected IsErrorContext to return true for ErrorContext")
	}

	if IsErrorContext(regularErr) {
		t.Error("Expected IsErrorContext to return false for regular error")
	}

	if IsErrorContext(nil) {
		t.Error("Expected IsErrorContext to return false for nil")
	}
}

func TestErrorChain(t *testing.T) {
	chain := NewErrorChain()

	if chain.HasErrors() {
		t.Error("New chain should have no errors")
	}
	if chain.Count() != 0 {
		t.Errorf("Expected count 0, got %d", chain.Count())
	}

	// Add errors
	chain.AddSimple(errors.New("error 1"), "step 1")
	chain.Add(errors.New("error 2"), "step 2", map[string]interface{}{
		"detail": "value",
	})

	if !chain.HasErrors() {
		t.Error("Chain should have errors")
	}
	if chain.Count() != 2 {
		t.Errorf("Expected count 2, got %d", chain.Count())
	}

	// Test First/Last
	first := chain.First()
	if first == nil {
		t.Fatal("Expected first error")
	}
	if !strings.Contains(first.Error(), "step 1") {
		t.Error("Expected first error to contain step 1")
	}

	last := chain.Last()
	if last == nil {
		t.Fatal("Expected last error")
	}
	if !strings.Contains(last.Error(), "step 2") {
		t.Error("Expected last error to contain step 2")
	}

	// Test Error message
	msg := chain.Error()
	if !strings.Contains(msg, "error chain") {
		t.Error("Expected error message to mention error chain")
	}
	if !strings.Contains(msg, "2 errors") {
		t.Error("Expected error message to mention count")
	}
}

func TestErrorChain_Empty(t *testing.T) {
	chain := NewErrorChain()

	if chain.First() != nil {
		t.Error("Expected nil from empty chain First()")
	}
	if chain.Last() != nil {
		t.Error("Expected nil from empty chain Last()")
	}
	if chain.Error() != "no errors" {
		t.Errorf("Expected 'no errors', got %q", chain.Error())
	}
}

func TestErrorChain_SingleError(t *testing.T) {
	chain := NewErrorChain()
	chain.AddSimple(errors.New("single error"), "operation")

	// Should return the error directly without "error chain" wrapper
	msg := chain.Error()
	if strings.Contains(msg, "error chain") {
		t.Error("Single error should not have 'error chain' prefix")
	}
	if !strings.Contains(msg, "single error") {
		t.Error("Expected message to contain the error")
	}
}

func TestErrorChain_NilErrors(t *testing.T) {
	chain := NewErrorChain()
	chain.AddSimple(nil, "operation 1")
	chain.Add(nil, "operation 2", nil)

	// Nil errors should not be added
	if chain.HasErrors() {
		t.Error("Nil errors should not be added to chain")
	}
	if chain.Count() != 0 {
		t.Errorf("Expected count 0, got %d", chain.Count())
	}
}

func TestErrorChain_All(t *testing.T) {
	chain := NewErrorChain()
	chain.AddSimple(errors.New("error 1"), "op1")
	chain.AddSimple(errors.New("error 2"), "op2")
	chain.AddSimple(errors.New("error 3"), "op3")

	all := chain.All()
	if len(all) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(all))
	}
}

func TestSummarizeError_Nil(t *testing.T) {
	summary := SummarizeError(nil)
	if summary != nil {
		t.Error("Expected nil summary for nil error")
	}
}

func TestSummarizeError_RegularError(t *testing.T) {
	err := errors.New("regular error")
	summary := SummarizeError(err)

	if summary == nil {
		t.Fatal("Expected summary")
	}
	if summary.Type != "error" {
		t.Errorf("Expected type 'error', got %q", summary.Type)
	}
	if summary.Code != "" {
		t.Error("Expected empty code for regular error")
	}
	if summary.Retryable {
		t.Error("Expected non-retryable for regular error")
	}
	if summary.Message != "regular error" {
		t.Errorf("Expected message 'regular error', got %q", summary.Message)
	}
}

func TestSummarizeError_CodedError(t *testing.T) {
	err := NewAPIKeyError(errors.New("missing key"))
	summary := SummarizeError(err)

	if summary == nil {
		t.Fatal("Expected summary")
	}
	if summary.Type != "CodedError" {
		t.Errorf("Expected type 'CodedError', got %q", summary.Type)
	}
	if summary.Code != ErrCodeAPIKeyMissing {
		t.Errorf("Expected code %q, got %q", ErrCodeAPIKeyMissing, summary.Code)
	}
	if summary.Retryable {
		t.Error("Expected non-retryable for API key error")
	}
}

func TestSummarizeError_PanicError(t *testing.T) {
	panicErr := &PanicError{
		Value:      "test panic",
		StackTrace: "goroutine 1 [running]:\ntest stack trace",
	}
	summary := SummarizeError(panicErr)

	if summary == nil {
		t.Fatal("Expected summary")
	}
	if summary.Type != "PanicError" {
		t.Errorf("Expected type 'PanicError', got %q", summary.Type)
	}
	if summary.Retryable {
		t.Error("Expected non-retryable for panic error")
	}
	if summary.Context["panic_value"] != "test panic" {
		t.Error("Expected panic value in context")
	}
	if _, ok := summary.Context["stack_trace"]; !ok {
		t.Error("Expected stack trace in context")
	}
}

func TestSummarizeError_ErrorContext(t *testing.T) {
	baseErr := NewRateLimitError(errors.New("rate limited"))
	contextErr := WithContext(baseErr, "API call", map[string]interface{}{
		"model": "gpt-4o-mini",
		"retry": 2,
	})
	summary := SummarizeError(contextErr)

	if summary == nil {
		t.Fatal("Expected summary")
	}
	if summary.Type != "CodedError" {
		t.Errorf("Expected type 'CodedError' (from underlying), got %q", summary.Type)
	}
	if summary.Code != ErrCodeRateLimitExceeded {
		t.Errorf("Expected code from underlying error, got %q", summary.Code)
	}
	if !summary.Retryable {
		t.Error("Expected retryable for rate limit error")
	}
	if summary.Context["operation"] != "API call" {
		t.Error("Expected operation in context")
	}
	if summary.Context["model"] != "gpt-4o-mini" {
		t.Error("Expected model in context")
	}
}

func TestSummarizeError_SentinelErrors(t *testing.T) {
	tests := []struct {
		err       error
		code      string
		retryable bool
	}{
		{ErrAPIKey, "API_KEY_MISSING", false},
		{ErrRateLimit, "RATE_LIMIT_EXCEEDED", true},
		{ErrTimeout, "REQUEST_TIMEOUT", true},
		{ErrToolExecution, "TOOL_EXECUTION_FAILED", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			summary := SummarizeError(tt.err)
			if summary == nil {
				t.Fatal("Expected summary")
			}
			if summary.Type != "SentinelError" {
				t.Errorf("Expected type 'SentinelError', got %q", summary.Type)
			}
			if summary.Code != tt.code {
				t.Errorf("Expected code %q, got %q", tt.code, summary.Code)
			}
			if summary.Retryable != tt.retryable {
				t.Errorf("Expected retryable=%v, got %v", tt.retryable, summary.Retryable)
			}
		})
	}
}

func TestSummarizeError_LongStackTrace(t *testing.T) {
	// Create panic with very long stack trace
	longTrace := strings.Repeat("goroutine line\n", 100)
	panicErr := &PanicError{
		Value:      "test",
		StackTrace: longTrace,
	}
	summary := SummarizeError(panicErr)

	if summary == nil {
		t.Fatal("Expected summary")
	}

	stackTrace := summary.Context["stack_trace"].(string)
	if len(stackTrace) > 510 { // 500 + "..."
		t.Errorf("Expected truncated stack trace, got length %d", len(stackTrace))
	}
	if !strings.HasSuffix(stackTrace, "...") {
		t.Error("Expected stack trace to be truncated with ...")
	}
}

func TestErrorContext_RealWorldScenario(t *testing.T) {
	// Simulate a real-world error flow
	baseErr := errors.New("network timeout")

	// Wrap with coded error
	codedErr := NewTimeoutError(baseErr)

	// Add context
	contextErr := WithContext(codedErr, "Chat completion", map[string]interface{}{
		"model":         "gpt-4o-mini",
		"retry_attempt": 3,
		"duration_ms":   5000,
	})

	// Summarize for logging
	summary := SummarizeError(contextErr)

	if summary == nil {
		t.Fatal("Expected summary")
	}

	// Should preserve coded error info
	if summary.Code != ErrCodeRequestTimeout {
		t.Errorf("Expected timeout code, got %q", summary.Code)
	}
	if !summary.Retryable {
		t.Error("Expected retryable for timeout")
	}

	// Should have context details
	if summary.Context["operation"] != "Chat completion" {
		t.Error("Expected operation in context")
	}
	if summary.Context["model"] != "gpt-4o-mini" {
		t.Error("Expected model in context")
	}
	if summary.Context["retry_attempt"] != 3 {
		t.Error("Expected retry attempt in context")
	}
}
