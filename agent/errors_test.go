package agent

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestCustomErrorTypes tests custom error type definitions
func TestCustomErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrAPIKey", ErrAPIKey},
		{"ErrRateLimit", ErrRateLimit},
		{"ErrTimeout", ErrTimeout},
		{"ErrRefusal", ErrRefusal},
		{"ErrInvalidResponse", ErrInvalidResponse},
		{"ErrMaxRetries", ErrMaxRetries},
		{"ErrToolExecution", ErrToolExecution},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("Expected non-nil error for %s", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("Expected non-empty error message for %s", tt.name)
			}
		})
	}
}

// TestAPIError tests APIError struct
func TestAPIError(t *testing.T) {
	err := NewAPIError("rate_limit_exceeded", "Too many requests", 429, nil)

	if err.Type != "rate_limit_exceeded" {
		t.Errorf("Expected type 'rate_limit_exceeded', got %s", err.Type)
	}

	if err.Message != "Too many requests" {
		t.Errorf("Expected message 'Too many requests', got %s", err.Message)
	}

	if err.StatusCode != 429 {
		t.Errorf("Expected status code 429, got %d", err.StatusCode)
	}

	expectedMsg := "rate_limit_exceeded (status 429): Too many requests"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestAPIErrorWithoutStatusCode tests APIError without status code
func TestAPIErrorWithoutStatusCode(t *testing.T) {
	err := NewAPIError("invalid_request", "Bad request", 0, nil)

	expectedMsg := "invalid_request: Bad request"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestErrorCheckers tests error type checking functions
func TestErrorCheckers(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		checker func(error) bool
		want    bool
	}{
		{"IsAPIKeyError with ErrAPIKey", ErrAPIKey, IsAPIKeyError, true},
		{"IsAPIKeyError with wrapped", WrapAPIKey(errors.New("test")), IsAPIKeyError, true},
		{"IsAPIKeyError with other", ErrRateLimit, IsAPIKeyError, false},

		{"IsRateLimitError with ErrRateLimit", ErrRateLimit, IsRateLimitError, true},
		{"IsRateLimitError with wrapped", WrapRateLimit(errors.New("test")), IsRateLimitError, true},
		{"IsRateLimitError with other", ErrAPIKey, IsRateLimitError, false},

		{"IsTimeoutError with ErrTimeout", ErrTimeout, IsTimeoutError, true},
		{"IsTimeoutError with wrapped", WrapTimeout(errors.New("test")), IsTimeoutError, true},
		{"IsTimeoutError with other", ErrAPIKey, IsTimeoutError, false},

		{"IsRefusalError with ErrRefusal", ErrRefusal, IsRefusalError, true},
		{"IsRefusalError with wrapped", WrapRefusal("test"), IsRefusalError, true},
		{"IsRefusalError with other", ErrAPIKey, IsRefusalError, false},

		{"IsMaxRetriesError with ErrMaxRetries", ErrMaxRetries, IsMaxRetriesError, true},
		{"IsMaxRetriesError with wrapped", WrapMaxRetries(3, errors.New("test")), IsMaxRetriesError, true},
		{"IsMaxRetriesError with other", ErrAPIKey, IsMaxRetriesError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checker(tt.err)
			if got != tt.want {
				t.Errorf("Expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestWithTimeout tests timeout configuration
func TestWithTimeout(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTimeout(5 * time.Second)

	if builder.timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", builder.timeout)
	}
}

// TestWithRetry tests retry configuration
func TestWithRetry(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(3)

	if builder.maxRetries != 3 {
		t.Errorf("Expected maxRetries 3, got %d", builder.maxRetries)
	}

	// Should set default retry delay
	if builder.retryDelay != time.Second {
		t.Errorf("Expected default retryDelay 1s, got %v", builder.retryDelay)
	}
}

// TestWithRetryDelay tests retry delay configuration
func TestWithRetryDelay(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(3).
		WithRetryDelay(2 * time.Second)

	if builder.retryDelay != 2*time.Second {
		t.Errorf("Expected retryDelay 2s, got %v", builder.retryDelay)
	}
}

// TestWithExponentialBackoff tests exponential backoff configuration
func TestWithExponentialBackoff(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(3).
		WithExponentialBackoff()

	if !builder.useExpBackoff {
		t.Error("Expected useExpBackoff to be true")
	}

	if builder.retryDelay != time.Second {
		t.Errorf("Expected default retryDelay 1s, got %v", builder.retryDelay)
	}
}

// TestCalculateRetryDelayFixed tests fixed retry delay calculation
func TestCalculateRetryDelayFixed(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(3).
		WithRetryDelay(2 * time.Second)

	// All attempts should have same delay
	for attempt := 0; attempt < 3; attempt++ {
		delay := builder.calculateRetryDelay(attempt)
		if delay != 2*time.Second {
			t.Errorf("Attempt %d: Expected delay 2s, got %v", attempt, delay)
		}
	}
}

// TestCalculateRetryDelayExponential tests exponential backoff calculation
func TestCalculateRetryDelayExponential(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(5).
		WithRetryDelay(time.Second).
		WithExponentialBackoff()

	expected := []time.Duration{
		1 * time.Second,  // 1 * 2^0 = 1s
		2 * time.Second,  // 1 * 2^1 = 2s
		4 * time.Second,  // 1 * 2^2 = 4s
		8 * time.Second,  // 1 * 2^3 = 8s
		16 * time.Second, // 1 * 2^4 = 16s
	}

	for attempt, want := range expected {
		got := builder.calculateRetryDelay(attempt)
		if got != want {
			t.Errorf("Attempt %d: Expected delay %v, got %v", attempt, want, got)
		}
	}
}

// TestIsRetryable tests retryable error detection
func TestIsRetryable(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"RateLimit is retryable", ErrRateLimit, true},
		{"Wrapped RateLimit is retryable", WrapRateLimit(errors.New("test")), true},
		{"Timeout is retryable", ErrTimeout, true},
		{"Wrapped Timeout is retryable", WrapTimeout(errors.New("test")), true},
		{"APIKey is not retryable", ErrAPIKey, false},
		{"Wrapped APIKey is not retryable", WrapAPIKey(errors.New("test")), false},
		{"Refusal is not retryable", ErrRefusal, false},
		{"InvalidResponse is not retryable", ErrInvalidResponse, false},
		{"Generic error is not retryable", errors.New("generic"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := builder.isRetryable(tt.err)
			if got != tt.want {
				t.Errorf("Expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestExecuteWithRetrySuccess tests successful execution without retries
func TestExecuteWithRetrySuccess(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(3)

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return nil // Success on first try
	}

	err := builder.executeWithRetry(context.Background(), operation)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

// TestExecuteWithRetryEventualSuccess tests retry until success
func TestExecuteWithRetryEventualSuccess(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(3).
		WithRetryDelay(10 * time.Millisecond)

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return WrapRateLimit(errors.New("rate limited"))
		}
		return nil // Success on 3rd try
	}

	err := builder.executeWithRetry(context.Background(), operation)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

// TestExecuteWithRetryMaxRetriesExceeded tests max retries exceeded
func TestExecuteWithRetryMaxRetriesExceeded(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(2).
		WithRetryDelay(10 * time.Millisecond)

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return WrapRateLimit(errors.New("rate limited"))
	}

	err := builder.executeWithRetry(context.Background(), operation)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !IsMaxRetriesError(err) {
		t.Errorf("Expected MaxRetriesError, got %v", err)
	}

	// Should call 1 initial + 2 retries = 3 times
	if callCount != 3 {
		t.Errorf("Expected 3 calls (1 + 2 retries), got %d", callCount)
	}
}

// TestExecuteWithRetryNonRetryableError tests non-retryable error
func TestExecuteWithRetryNonRetryableError(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRetry(3).
		WithRetryDelay(10 * time.Millisecond)

	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		return WrapAPIKey(errors.New("invalid key"))
	}

	err := builder.executeWithRetry(context.Background(), operation)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !IsAPIKeyError(err) {
		t.Errorf("Expected APIKeyError, got %v", err)
	}

	// Should only call once (no retries for non-retryable errors)
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retries), got %d", callCount)
	}
}

// TestExecuteWithTimeout tests timeout handling
func TestExecuteWithTimeout(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTimeout(50 * time.Millisecond)

	operation := func(ctx context.Context) error {
		// Simulate slow operation
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	err := builder.executeWithRetry(context.Background(), operation)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if !IsTimeoutError(err) {
		t.Errorf("Expected TimeoutError, got %v", err)
	}
}

// TestErrorHandlingChaining tests method chaining
func TestErrorHandlingChaining(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTimeout(30 * time.Second).
		WithRetry(3).
		WithRetryDelay(time.Second).
		WithExponentialBackoff()

	if builder.timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", builder.timeout)
	}

	if builder.maxRetries != 3 {
		t.Errorf("Expected maxRetries 3, got %d", builder.maxRetries)
	}

	if builder.retryDelay != time.Second {
		t.Errorf("Expected retryDelay 1s, got %v", builder.retryDelay)
	}

	if !builder.useExpBackoff {
		t.Error("Expected useExpBackoff to be true")
	}
}
