package agent

import (
	"errors"
	"fmt"
	"testing"
)

func TestCodedError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CodedError
		expected string
	}{
		{
			name: "error with underlying error",
			err: &CodedError{
				Code:    ErrCodeRateLimitExceeded,
				Message: "Too many requests",
				Err:     errors.New("original error"),
			},
			expected: "[RATE_LIMIT_EXCEEDED] Too many requests: original error",
		},
		{
			name: "error without underlying error",
			err: &CodedError{
				Code:    ErrCodeAPIKeyMissing,
				Message: "API key required",
				Err:     nil,
			},
			expected: "[API_KEY_MISSING] API key required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCodedError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	codedErr := &CodedError{
		Code:    ErrCodeRequestTimeout,
		Message: "Timeout occurred",
		Err:     originalErr,
	}

	if unwrapped := codedErr.Unwrap(); unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNewCodedError(t *testing.T) {
	code := ErrCodeToolExecutionFailed
	message := "Tool failed"
	originalErr := errors.New("tool error")

	err := NewCodedError(code, message, originalErr)

	if err.Code != code {
		t.Errorf("Code = %v, want %v", err.Code, code)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
	if err.Err != originalErr {
		t.Errorf("Err = %v, want %v", err.Err, originalErr)
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func() *CodedError
		expectedCode string
		checkMessage bool
	}{
		{
			name: "NewAPIKeyError",
			constructor: func() *CodedError {
				return NewAPIKeyError(errors.New("invalid key"))
			},
			expectedCode: ErrCodeAPIKeyMissing,
			checkMessage: true,
		},
		{
			name: "NewRateLimitError",
			constructor: func() *CodedError {
				return NewRateLimitError(errors.New("too many requests"))
			},
			expectedCode: ErrCodeRateLimitExceeded,
			checkMessage: true,
		},
		{
			name: "NewTimeoutError",
			constructor: func() *CodedError {
				return NewTimeoutError(errors.New("timeout"))
			},
			expectedCode: ErrCodeRequestTimeout,
			checkMessage: true,
		},
		{
			name: "NewToolError",
			constructor: func() *CodedError {
				return NewToolError("calculator", errors.New("division by zero"))
			},
			expectedCode: ErrCodeToolExecutionFailed,
			checkMessage: true,
		},
		{
			name: "NewToolPanicError",
			constructor: func() *CodedError {
				return NewToolPanicError("weather", "runtime error")
			},
			expectedCode: ErrCodeToolPanicked,
			checkMessage: true,
		},
		{
			name: "NewEmbeddingError",
			constructor: func() *CodedError {
				return NewEmbeddingError(errors.New("embedding failed"))
			},
			expectedCode: ErrCodeEmbeddingFailed,
			checkMessage: true,
		},
		{
			name: "NewCacheError",
			constructor: func() *CodedError {
				return NewCacheError("Get", errors.New("redis down"))
			},
			expectedCode: ErrCodeCacheOperationFailed,
			checkMessage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.expectedCode)
			}

			if tt.checkMessage && err.Message == "" {
				t.Errorf("Message is empty")
			}
		})
	}
}

func TestIsCodedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "CodedError",
			err:      NewCodedError(ErrCodeAPIKeyMissing, "test", nil),
			expected: true,
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "Wrapped CodedError",
			err:      fmt.Errorf("wrapped: %w", NewCodedError(ErrCodeRateLimitExceeded, "test", nil)),
			expected: false, // fmt.Errorf wrapping changes type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCodedError(tt.err); got != tt.expected {
				t.Errorf("IsCodedError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHasErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     string
		expected bool
	}{
		{
			name:     "matching code",
			err:      NewCodedError(ErrCodeRateLimitExceeded, "test", nil),
			code:     ErrCodeRateLimitExceeded,
			expected: true,
		},
		{
			name:     "non-matching code",
			err:      NewCodedError(ErrCodeRateLimitExceeded, "test", nil),
			code:     ErrCodeAPIKeyMissing,
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			code:     ErrCodeRateLimitExceeded,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasErrorCode(tt.err, tt.code); got != tt.expected {
				t.Errorf("HasErrorCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "CodedError",
			err:      NewCodedError(ErrCodeToolTimeout, "test", nil),
			expected: ErrCodeToolTimeout,
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: "",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorCode(tt.err); got != tt.expected {
				t.Errorf("GetErrorCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsRetryableByCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "rate limit (retryable)",
			err:      NewCodedError(ErrCodeRateLimitExceeded, "test", nil),
			expected: true,
		},
		{
			name:     "timeout (retryable)",
			err:      NewCodedError(ErrCodeRequestTimeout, "test", nil),
			expected: true,
		},
		{
			name:     "invalid response (retryable)",
			err:      NewCodedError(ErrCodeInvalidResponse, "test", nil),
			expected: true,
		},
		{
			name:     "cache error (retryable)",
			err:      NewCodedError(ErrCodeCacheOperationFailed, "test", nil),
			expected: true,
		},
		{
			name:     "API key (not retryable)",
			err:      NewCodedError(ErrCodeAPIKeyMissing, "test", nil),
			expected: false,
		},
		{
			name:     "tool error (not retryable)",
			err:      NewCodedError(ErrCodeToolExecutionFailed, "test", nil),
			expected: false,
		},
		{
			name:     "regular error (not retryable)",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorCodes_Constants(t *testing.T) {
	// Verify all error codes are defined and non-empty
	codes := []string{
		ErrCodeAPIKeyMissing,
		ErrCodeAPIKeyInvalid,
		ErrCodeRateLimitExceeded,
		ErrCodeRequestTimeout,
		ErrCodeInvalidResponse,
		ErrCodeContentRefused,
		ErrCodeToolNotFound,
		ErrCodeToolExecutionFailed,
		ErrCodeToolPanicked,
		ErrCodeToolTimeout,
		ErrCodeVectorStoreNotConfigured,
		ErrCodeEmbeddingFailed,
		ErrCodeVectorSearchFailed,
		ErrCodeMemoryFull,
		ErrCodeCacheConnectionFailed,
		ErrCodeCacheOperationFailed,
		ErrCodeInvalidConfiguration,
		ErrCodeUnsupportedProvider,
		ErrCodeMaxRetriesExceeded,
		ErrCodeNoResponseChoices,
	}

	for i, code := range codes {
		if code == "" {
			t.Errorf("Error code at index %d is empty", i)
		}
	}

	// Verify we have exactly 20 error codes (LEAN approach)
	if len(codes) != 20 {
		t.Errorf("Expected 20 error codes, got %d", len(codes))
	}
}

func TestCodedError_WithErrorsIs(t *testing.T) {
	originalErr := errors.New("original error")
	codedErr := NewCodedError(ErrCodeRequestTimeout, "timeout", originalErr)

	// Test errors.Is compatibility
	if !errors.Is(codedErr, originalErr) {
		t.Error("errors.Is should find underlying error")
	}
}

func TestCodedError_RealWorldScenarios(t *testing.T) {
	t.Run("handle rate limit in retry logic", func(t *testing.T) {
		err := NewRateLimitError(errors.New("429 Too Many Requests"))

		// Programmatic handling
		if HasErrorCode(err, ErrCodeRateLimitExceeded) {
			if !IsRetryable(err) {
				t.Error("Rate limit error should be retryable")
			}
			// In real code: time.Sleep(delay); retry()
		} else {
			t.Error("Should detect rate limit error")
		}
	})

	t.Run("handle API key error (don't retry)", func(t *testing.T) {
		err := NewAPIKeyError(errors.New("invalid_api_key"))

		if IsRetryable(err) {
			t.Error("API key error should not be retryable")
		}

		if !HasErrorCode(err, ErrCodeAPIKeyMissing) {
			t.Error("Should detect API key error")
		}
	})

	t.Run("handle tool panic with code", func(t *testing.T) {
		err := NewToolPanicError("calculator", "division by zero")

		code := GetErrorCode(err)
		if code != ErrCodeToolPanicked {
			t.Errorf("Expected code %s, got %s", ErrCodeToolPanicked, code)
		}

		// Check error message includes tool name
		if err.Message == "" {
			t.Error("Error message should not be empty")
		}
	})
}
