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

	// ErrReActMaxIterations indicates max iterations reached without final answer
	// Added in v0.7.6 - this is now a structured error instead of generic fmt.Errorf
	ErrReActMaxIterations = errors.New("max iterations reached without final answer\n\n" +
		"Fix:\n" +
		"  1. Use task complexity: .WithReActComplexity(agent.ReActTaskSimple/Medium/Complex)\n" +
		"  2. Enable auto-fallback (default): .WithReActAutoFallback(true)\n" +
		"  3. Increase iterations: .WithReActMaxIterations(10)\n" +
		"  4. Enable reminders (default): .WithReActIterationReminders(true)\n" +
		"  5. Simplify the task or break it into smaller steps")
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

// ReActError provides detailed context when ReAct execution fails.
// Added in v0.7.6 to give users actionable debugging information.
//
// This error type is returned when:
//   - Max iterations reached without final_answer() (and auto-fallback disabled)
//   - Timeout occurs during ReAct execution
//   - Critical parsing errors in ReAct loop
//
// Usage:
//
//	result, err := ai.Execute(ctx, task)
//	if err != nil {
//	    var reactErr *agent.ReActError
//	    if errors.As(err, &reactErr) {
//	        fmt.Printf("Failed after %d iterations\n", reactErr.CurrentIteration)
//	        fmt.Printf("Steps: %v\n", reactErr.Steps)
//	        fmt.Printf("Suggestions: %v\n", reactErr.Suggestions)
//	    }
//	}
type ReActError struct {
	// Type indicates error category: "max_iterations", "timeout", "parse_error"
	Type string

	// Message is the human-readable error description
	Message string

	// CurrentIteration is the iteration where error occurred (1-based)
	CurrentIteration int

	// MaxIterations is the configured maximum
	MaxIterations int

	// Steps contains the reasoning steps completed before error
	Steps []ReActStep

	// Suggestions contains actionable fix recommendations
	Suggestions []string

	// Err is the underlying error (if any)
	Err error
}

func (e *ReActError) Error() string {
	baseMsg := fmt.Sprintf("ReAct %s (iteration %d/%d): %s",
		e.Type, e.CurrentIteration, e.MaxIterations, e.Message)

	if len(e.Suggestions) > 0 {
		baseMsg += "\n\nSuggestions:"
		for i, suggestion := range e.Suggestions {
			baseMsg += fmt.Sprintf("\n  %d. %s", i+1, suggestion)
		}
	}

	if len(e.Steps) > 0 {
		baseMsg += fmt.Sprintf("\n\nCompleted %d steps before failure:", len(e.Steps))
		thoughtCount := 0
		actionCount := 0
		for _, step := range e.Steps {
			if step.Type == StepTypeThought {
				thoughtCount++
			} else if step.Type == StepTypeAction {
				actionCount++
			}
		}
		baseMsg += fmt.Sprintf("\n  - %d thoughts, %d actions", thoughtCount, actionCount)
	}

	return baseMsg
}

func (e *ReActError) Unwrap() error {
	return e.Err
}

// NewReActMaxIterationsError creates a ReActError for max iterations scenario.
// This provides much better UX than generic "max iterations reached" message.
//
// Added in: v0.7.6
func NewReActMaxIterationsError(currentIteration, maxIterations int, steps []ReActStep) *ReActError {
	suggestions := []string{
		"Use .WithReActComplexity(agent.ReActTaskMedium) or ReActTaskComplex for harder tasks",
		"Enable auto-fallback (default): .WithReActAutoFallback(true) to get best-effort answers",
		"Increase max iterations: .WithReActMaxIterations(10) or higher",
		"Enable iteration reminders (default): .WithReActIterationReminders(true)",
		"Simplify the task or break it into multiple smaller agent calls",
	}

	return &ReActError{
		Type:             "max_iterations",
		Message:          "Maximum iterations reached without calling final_answer()",
		CurrentIteration: currentIteration,
		MaxIterations:    maxIterations,
		Steps:            steps,
		Suggestions:      suggestions,
		Err:              ErrReActMaxIterations,
	}
}

// NewReActTimeoutError creates a ReActError for timeout scenario.
//
// Added in: v0.7.6
func NewReActTimeoutError(currentIteration, maxIterations int, steps []ReActStep, timeout string) *ReActError {
	suggestions := []string{
		fmt.Sprintf("Increase timeout: .WithReActTimeout(120*time.Second) (current: %s)", timeout),
		"Use .WithReActComplexity(agent.ReActTaskComplex) for longer timeouts",
		"Simplify the task to require fewer steps",
		"Check network connectivity and LLM API status",
	}

	return &ReActError{
		Type:             "timeout",
		Message:          "ReAct execution timeout",
		CurrentIteration: currentIteration,
		MaxIterations:    maxIterations,
		Steps:            steps,
		Suggestions:      suggestions,
		Err:              ErrTimeout,
	}
}

// IsReActMaxIterationsError checks if error is ReAct max iterations related
func IsReActMaxIterationsError(err error) bool {
	var reactErr *ReActError
	if errors.As(err, &reactErr) {
		return reactErr.Type == "max_iterations"
	}
	return errors.Is(err, ErrReActMaxIterations)
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
