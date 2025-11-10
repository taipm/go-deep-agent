package agent

import (
	"errors"
	"fmt"
)

// Custom error types for better error handling and recovery

var (
	// ErrAPIKey indicates missing or invalid API key
	ErrAPIKey = errors.New("API key is missing or invalid\n\n" +
		"Fix:\n" +
		"  1. Set environment variable: export OPENAI_API_KEY=\"sk-...\"\n" +
		"  2. Or pass to constructor: agent.NewOpenAI(\"gpt-4\", \"sk-...\")\n" +
		"  3. Get your key: https://platform.openai.com/api-keys")

	// ErrRateLimit indicates rate limit exceeded
	ErrRateLimit = errors.New("rate limit exceeded - too many requests\n\n" +
		"Fix:\n" +
		"  1. Use .WithDefaults() - includes retry with exponential backoff\n" +
		"  2. Or configure: .WithRetry(5).WithRetryDelay(2*time.Second).WithExponentialBackoff()\n" +
		"  3. Upgrade tier: https://platform.openai.com/account/limits\n" +
		"  4. Use caching: .WithRedisCache(\"localhost:6379\", \"\", 0)")

	// ErrTimeout indicates request timeout
	ErrTimeout = errors.New("request timeout - operation took too long\n\n" +
		"Fix:\n" +
		"  1. Increase timeout: .WithTimeout(60 * time.Second)\n" +
		"  2. Use streaming for long responses: .Stream(...)\n" +
		"  3. Check network connection\n" +
		"  4. Check OpenAI status: https://status.openai.com")

	// ErrRefusal indicates content was refused by the model
	ErrRefusal = errors.New("content refused by model - policy violation\n\n" +
		"Fix:\n" +
		"  1. Review policies: https://openai.com/policies/usage-policies\n" +
		"  2. Rephrase your prompt to avoid policy violations\n" +
		"  3. Check content filters and safety settings")

	// ErrInvalidResponse indicates malformed or unexpected response
	ErrInvalidResponse = errors.New("invalid response from API\n\n" +
		"Fix:\n" +
		"  1. Enable debug mode: .WithDebug() to see raw response\n" +
		"  2. Check OpenAI status: https://status.openai.com\n" +
		"  3. Verify API key has proper permissions\n" +
		"  4. Update library: go get -u github.com/taipm/go-deep-agent")

	// ErrMaxRetries indicates maximum retry attempts exceeded
	ErrMaxRetries = errors.New("maximum retry attempts exceeded\n\n" +
		"Fix:\n" +
		"  1. Increase retries: .WithRetry(5) or .WithRetry(10)\n" +
		"  2. Check root cause - enable debug: .WithDebug()\n" +
		"  3. Increase retry delay: .WithRetryDelay(5*time.Second)\n" +
		"  4. Use exponential backoff: .WithExponentialBackoff()")

	// ErrToolExecution indicates tool execution failed
	ErrToolExecution = errors.New("tool execution failed\n\n" +
		"Fix:\n" +
		"  1. Enable debug logging: .WithDebug()\n" +
		"  2. Check tool function implementation\n" +
		"  3. Verify tool parameters match JSON schema\n" +
		"  4. Add error handling in tool function\n" +
		"  5. Increase tool timeout: .WithToolTimeout(60*time.Second)")
)

// APIError wraps API errors with additional context
type APIError struct {
	Type       string // Error type (e.g., "rate_limit", "invalid_api_key")
	Message    string // Error message
	StatusCode int    // HTTP status code (if applicable)
	Err        error  // Underlying error
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s (status %d): %s", e.Type, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// IsAPIKeyError checks if error is API key related
func IsAPIKeyError(err error) bool {
	return errors.Is(err, ErrAPIKey) || isAPIErrorType(err, "invalid_api_key")
}

// IsRateLimitError checks if error is rate limit related
func IsRateLimitError(err error) bool {
	return errors.Is(err, ErrRateLimit) || isAPIErrorType(err, "rate_limit_exceeded")
}

// IsTimeoutError checks if error is timeout related
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsRefusalError checks if error is refusal related
func IsRefusalError(err error) bool {
	return errors.Is(err, ErrRefusal)
}

// IsInvalidResponseError checks if error is invalid response related
func IsInvalidResponseError(err error) bool {
	return errors.Is(err, ErrInvalidResponse)
}

// IsMaxRetriesError checks if error is max retries exceeded
func IsMaxRetriesError(err error) bool {
	return errors.Is(err, ErrMaxRetries)
}

// IsToolExecutionError checks if error is tool execution related
func IsToolExecutionError(err error) bool {
	return errors.Is(err, ErrToolExecution)
}

// isAPIErrorType checks if error is APIError with specific type
func isAPIErrorType(err error, errorType string) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == errorType
	}
	return false
}

// NewAPIError creates a new APIError
func NewAPIError(errorType, message string, statusCode int, err error) *APIError {
	return &APIError{
		Type:       errorType,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// WrapAPIKey wraps an error as API key error
func WrapAPIKey(err error) error {
	return fmt.Errorf("%w: %v", ErrAPIKey, err)
}

// WrapRateLimit wraps an error as rate limit error
func WrapRateLimit(err error) error {
	return fmt.Errorf("%w: %v", ErrRateLimit, err)
}

// WrapTimeout wraps an error as timeout error
func WrapTimeout(err error) error {
	return fmt.Errorf("%w: %v", ErrTimeout, err)
}

// WrapRefusal wraps an error as refusal error
func WrapRefusal(message string) error {
	return fmt.Errorf("%w: %s", ErrRefusal, message)
}

// WrapInvalidResponse wraps an error as invalid response error
func WrapInvalidResponse(err error) error {
	return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
}

// WrapMaxRetries wraps an error as max retries exceeded
func WrapMaxRetries(attempts int, lastErr error) error {
	return fmt.Errorf("%w after %d attempts: %v", ErrMaxRetries, attempts, lastErr)
}

// WrapToolExecution wraps an error as tool execution error
func WrapToolExecution(toolName string, err error) error {
	return fmt.Errorf("%w (%s): %v", ErrToolExecution, toolName, err)
}
