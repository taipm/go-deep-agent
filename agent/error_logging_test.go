package agent

import (
	"errors"
	"fmt"
	"testing"
)

// TestErrorContext_LogFields tests ErrorContext.LogFields()
func TestErrorContext_LogFields(t *testing.T) {
	baseErr := errors.New("database connection failed")
	ctx := &ErrorContext{
		Operation: "database query",
		Details: map[string]interface{}{
			"query":        "SELECT * FROM users",
			"timeout_ms":   5000,
			"retry_count":  3,
			"host":         "db.example.com",
			"success_rate": 0.95,
		},
		Err: baseErr,
	}

	fields := ctx.LogFields()

	// Verify basic fields
	if len(fields) < 2 {
		t.Fatalf("Expected at least 2 fields, got %d", len(fields))
	}

	// Check error field
	found := false
	for _, f := range fields {
		if f.Key == "error" && f.Value == baseErr.Error() {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'error' field with base error message")
	}

	// Check operation field
	found = false
	for _, f := range fields {
		if f.Key == "operation" && f.Value == "database query" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'operation' field")
	}

	// Check all detail fields are present
	expectedKeys := map[string]bool{
		"query":        false,
		"timeout_ms":   false,
		"retry_count":  false,
		"host":         false,
		"success_rate": false,
	}
	for _, f := range fields {
		if _, ok := expectedKeys[f.Key]; ok {
			expectedKeys[f.Key] = true
		}
	}
	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Expected field '%s' not found in LogFields output", key)
		}
	}
}

// TestCodedError_LogFields tests CodedError.LogFields()
func TestCodedError_LogFields(t *testing.T) {
	baseErr := errors.New("connection refused")
	codedErr := &CodedError{
		Code:    ErrCodeRequestTimeout,
		Message: "Request timeout after 30s",
		Err:     baseErr,
	}

	fields := codedErr.LogFields()

	// Verify minimum fields
	if len(fields) < 3 {
		t.Fatalf("Expected at least 3 fields, got %d", len(fields))
	}

	// Check error_code
	found := false
	for _, f := range fields {
		if f.Key == "error_code" && f.Value == ErrCodeRequestTimeout {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'error_code' field")
	}

	// Check error_message
	found = false
	for _, f := range fields {
		if f.Key == "error_message" && f.Value == "Request timeout after 30s" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'error_message' field")
	}

	// Check retryable
	found = false
	for _, f := range fields {
		if f.Key == "retryable" {
			found = true
			if retryable, ok := f.Value.(bool); !ok || !retryable {
				t.Error("Expected 'retryable' to be true for timeout errors")
			}
			break
		}
	}
	if !found {
		t.Error("Expected 'retryable' field")
	}

	// Check underlying_error
	found = false
	for _, f := range fields {
		if f.Key == "underlying_error" && f.Value == baseErr.Error() {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'underlying_error' field")
	}
}

// TestCodedError_LogFields_NoUnderlying tests CodedError without underlying error
func TestCodedError_LogFields_NoUnderlying(t *testing.T) {
	codedErr := &CodedError{
		Code:    ErrCodeAPIKeyMissing,
		Message: "API key not configured",
		Err:     nil,
	}

	fields := codedErr.LogFields()

	// Should not have underlying_error field
	for _, f := range fields {
		if f.Key == "underlying_error" {
			t.Error("Should not have 'underlying_error' field when Err is nil")
		}
	}

	// But should have other fields
	if len(fields) < 3 {
		t.Errorf("Expected at least 3 fields, got %d", len(fields))
	}
}

// TestPanicError_LogFields tests PanicError.LogFields()
func TestPanicError_LogFields(t *testing.T) {
	panicErr := &PanicError{
		Value:      "nil pointer dereference",
		StackTrace: "goroutine 1 [running]:\nmain.function()\n\t/path/to/file.go:42\n",
	}

	fields := panicErr.LogFields()

	// Check error_type
	found := false
	for _, f := range fields {
		if f.Key == "error_type" && f.Value == "panic" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'error_type' field with value 'panic'")
	}

	// Check panic_value
	found = false
	for _, f := range fields {
		if f.Key == "panic_value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'panic_value' field")
	}

	// Check stack_trace
	found = false
	for _, f := range fields {
		if f.Key == "stack_trace" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'stack_trace' field")
	}
}

// TestPanicError_LogFields_LongStackTrace tests truncation of long stack traces
func TestPanicError_LogFields_LongStackTrace(t *testing.T) {
	// Create a very long stack trace (> 500 chars)
	longTrace := ""
	for i := 0; i < 100; i++ {
		longTrace += fmt.Sprintf("goroutine %d [running]:\n", i)
	}

	panicErr := &PanicError{
		Value:      "test panic",
		StackTrace: longTrace,
	}

	fields := panicErr.LogFields()

	// Check for stack_trace truncation
	foundTruncated := false
	foundFullLength := false
	for _, f := range fields {
		if f.Key == "stack_trace" {
			foundTruncated = true
			str, ok := f.Value.(string)
			if !ok {
				t.Error("stack_trace value should be string")
			}
			if len(str) > 503 { // 500 + "..."
				t.Error("stack_trace should be truncated to ~500 chars")
			}
		}
		if f.Key == "stack_trace_full_length" {
			foundFullLength = true
			length, ok := f.Value.(int)
			if !ok {
				t.Error("stack_trace_full_length should be int")
			}
			if length != len(longTrace) {
				t.Errorf("Expected full length %d, got %d", len(longTrace), length)
			}
		}
	}

	if !foundTruncated {
		t.Error("Expected truncated stack_trace field")
	}
	if !foundFullLength {
		t.Error("Expected stack_trace_full_length field for long traces")
	}
}

// TestExtractLogFields tests the universal ExtractLogFields function
func TestExtractLogFields(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedKeys  []string
		expectedCount int
	}{
		{
			name:          "nil error",
			err:           nil,
			expectedKeys:  []string{},
			expectedCount: 0,
		},
		{
			name:          "plain error",
			err:           errors.New("plain error"),
			expectedKeys:  []string{"error"},
			expectedCount: 1,
		},
		{
			name: "CodedError",
			err: &CodedError{
				Code:    ErrCodeRateLimitExceeded,
				Message: "Rate limit hit",
				Err:     nil,
			},
			expectedKeys:  []string{"error_code", "error_message", "retryable"},
			expectedCount: 3,
		},
		{
			name: "PanicError",
			err: &PanicError{
				Value:      "test panic",
				StackTrace: "short trace",
			},
			expectedKeys:  []string{"error_type", "panic_value", "stack_trace"},
			expectedCount: 3,
		},
		{
			name: "ErrorContext",
			err: &ErrorContext{
				Operation: "test op",
				Details: map[string]interface{}{
					"key1": "value1",
					"key2": 42,
				},
				Err: errors.New("base error"),
			},
			expectedKeys:  []string{"error", "operation", "key1", "key2"},
			expectedCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := ExtractLogFields(tt.err)

			if len(fields) != tt.expectedCount {
				t.Errorf("Expected %d fields, got %d", tt.expectedCount, len(fields))
			}

			for _, expectedKey := range tt.expectedKeys {
				found := false
				for _, f := range fields {
					if f.Key == expectedKey {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected key '%s' not found in fields", expectedKey)
				}
			}
		})
	}
}

// TestExtractLogFieldsWithSummary tests the summary-enhanced field extraction
func TestExtractLogFieldsWithSummary(t *testing.T) {
	codedErr := &CodedError{
		Code:    ErrCodeRateLimitExceeded,
		Message: "Rate limit exceeded",
		Err:     errors.New("429 Too Many Requests"),
	}

	fields := ExtractLogFieldsWithSummary(codedErr)

	// Should have base fields + summary fields
	expectedKeys := map[string]bool{
		"error_code":    false,
		"error_message": false,
		"error_type":    false,
		"retryable":     false,
	}

	for _, f := range fields {
		if _, ok := expectedKeys[f.Key]; ok {
			expectedKeys[f.Key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Expected key '%s' not found in summary fields", key)
		}
	}

	// Verify retryable is true for rate limit
	for _, f := range fields {
		if f.Key == "retryable" {
			if retryable, ok := f.Value.(bool); !ok || !retryable {
				t.Error("Expected retryable=true for rate limit error")
			}
		}
	}
}

// TestExtractLogFields_Integration tests real-world scenario
func TestExtractLogFields_Integration(t *testing.T) {
	// Simulate a real error scenario
	baseErr := errors.New("connection timeout")
	codedErr := NewTimeoutError(baseErr)
	contextErr := WithContext(codedErr, "API request", map[string]interface{}{
		"endpoint": "/api/users",
		"method":   "GET",
		"duration": 30000,
	})

	fields := ExtractLogFieldsWithSummary(contextErr)

	// Should have fields from all layers
	expectedKeys := []string{
		"error",      // From base error
		"operation",  // From ErrorContext
		"endpoint",   // From ErrorContext details
		"method",     // From ErrorContext details
		"duration",   // From ErrorContext details
		"error_type", // From summary
		"error_code", // From summary
		"retryable",  // From summary
	}

	foundKeys := make(map[string]bool)
	for _, f := range fields {
		foundKeys[f.Key] = true
	}

	for _, key := range expectedKeys {
		if !foundKeys[key] {
			t.Errorf("Expected key '%s' not found in integrated error fields", key)
		}
	}
}
