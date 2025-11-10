package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

// Example demonstrating programmatic error handling with error codes

func main() {
	fmt.Println("=== Error Codes Example ===\n")

	// Example 1: Basic error code checking
	example1BasicErrorHandling()

	// Example 2: Retry logic with error codes
	example2RetryLogic()

	// Example 3: Error code routing
	example3ErrorRouting()

	// Example 4: Tool error handling
	example4ToolErrors()
}

// Example 1: Basic error code checking
func example1BasicErrorHandling() {
	fmt.Println("--- Example 1: Basic Error Code Checking ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Simulate API key error
		err := agent.NewAPIKeyError(nil)
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Error Code: %s\n", agent.GetErrorCode(err))
		fmt.Printf("Has API Key error: %v\n", agent.HasErrorCode(err, agent.ErrCodeAPIKeyMissing))
		fmt.Printf("Is Retryable: %v\n\n", agent.IsRetryable(err))
		return
	}

	ai := agent.NewOpenAI("gpt-4", apiKey)
	ctx := context.Background()

	resp, err := ai.Ask(ctx, "Hello")
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Response: %s\n\n", resp)
}

// Example 2: Retry logic based on error codes
func example2RetryLogic() {
	fmt.Println("--- Example 2: Smart Retry Logic ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Skipping: OPENAI_API_KEY not set\n")
		return
	}

	ai := agent.NewOpenAI("gpt-4", apiKey)
	ctx := context.Background()

	message := "What is Go?"

	// Retry with exponential backoff for retryable errors
	var resp string
	var err error
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = ai.Ask(ctx, message)

		if err == nil {
			fmt.Printf("Success on attempt %d\n", attempt+1)
			fmt.Printf("Response: %s\n\n", resp)
			return
		}

		// Check if error is retryable
		if !agent.IsRetryable(err) {
			fmt.Printf("Error is not retryable: %v\n\n", err)
			return
		}

		if attempt < maxRetries {
			delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
			fmt.Printf("Attempt %d failed (code: %s), retrying in %v...\n",
				attempt+1, agent.GetErrorCode(err), delay)
			time.Sleep(delay)
		}
	}

	fmt.Printf("Failed after %d attempts: %v\n\n", maxRetries+1, err)
}

// Example 3: Route errors to different handlers
func example3ErrorRouting() {
	fmt.Println("--- Example 3: Error Routing ---")

	// Simulate different error types
	errors := []error{
		agent.NewAPIKeyError(nil),
		agent.NewRateLimitError(nil),
		agent.NewTimeoutError(nil),
		agent.NewToolError("calculator", nil),
		agent.NewEmbeddingError(nil),
	}

	for _, err := range errors {
		routeError(err)
	}

	fmt.Println()
}

func routeError(err error) {
	code := agent.GetErrorCode(err)

	fmt.Printf("Error: %v\n", err)

	switch code {
	case agent.ErrCodeAPIKeyMissing, agent.ErrCodeAPIKeyInvalid:
		fmt.Println("→ Action: Check API key configuration")
		fmt.Println("  - Verify OPENAI_API_KEY environment variable")
		fmt.Println("  - Ensure key starts with 'sk-'")

	case agent.ErrCodeRateLimitExceeded:
		fmt.Println("→ Action: Implement rate limiting")
		fmt.Println("  - Add retry with exponential backoff")
		fmt.Println("  - Consider upgrading OpenAI tier")
		fmt.Println("  - Enable caching to reduce requests")

	case agent.ErrCodeRequestTimeout:
		fmt.Println("→ Action: Increase timeout or use streaming")
		fmt.Println("  - Use .WithTimeout(60 * time.Second)")
		fmt.Println("  - Or use .Stream() for long responses")

	case agent.ErrCodeToolExecutionFailed, agent.ErrCodeToolPanicked:
		fmt.Println("→ Action: Debug tool implementation")
		fmt.Println("  - Enable .WithDebug() to see tool logs")
		fmt.Println("  - Check tool function for errors")
		fmt.Println("  - Add error handling to tool")

	case agent.ErrCodeEmbeddingFailed:
		fmt.Println("→ Action: Check embedding configuration")
		fmt.Println("  - Verify embedding API key")
		fmt.Println("  - Check text length (max ~8,000 tokens)")
		fmt.Println("  - Ensure embedding provider is configured")

	default:
		fmt.Println("→ Action: General error handling")
		fmt.Println("  - Enable debug mode for more details")
		fmt.Println("  - Check OpenAI status: https://status.openai.com")
	}

	fmt.Println()
}

// Example 4: Tool error handling with codes
func example4ToolErrors() {
	fmt.Println("--- Example 4: Tool Error Handling ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Skipping: OPENAI_API_KEY not set\n")
		return
	}

	// Create tool that might fail
	calculatorTool := agent.Tool{
		Name:        "calculator",
		Description: "Performs calculations",
		Handler: func(args string) (string, error) {
			// Simulate error
			return "", fmt.Errorf("division by zero")
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{"type": "string"},
				"a":         map[string]interface{}{"type": "number"},
				"b":         map[string]interface{}{"type": "number"},
			},
		},
	}

	ai := agent.NewOpenAI("gpt-4", apiKey).
		WithTool(&calculatorTool)

	ctx := context.Background()
	resp, err := ai.Ask(ctx, "Calculate 10 divided by 0")

	if err != nil {
		code := agent.GetErrorCode(err)
		fmt.Printf("Error occurred (code: %s)\n", code)

		if code == agent.ErrCodeToolExecutionFailed || code == agent.ErrCodeToolPanicked {
			fmt.Println("Tool execution failed!")
			fmt.Println("Recommendation:")
			fmt.Println("1. Check tool implementation for errors")
			fmt.Println("2. Add input validation")
			fmt.Println("3. Handle edge cases (e.g., division by zero)")
			fmt.Println()
			return
		}
	}

	fmt.Printf("Response: %s\n\n", resp)
}

// Helper function to handle errors with codes
func handleError(err error) {
	if err == nil {
		return
	}

	// Check if it's a coded error
	if agent.IsCodedError(err) {
		code := agent.GetErrorCode(err)
		isRetryable := agent.IsRetryable(err)

		fmt.Printf("Error Code: %s\n", code)
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Retryable: %v\n\n", isRetryable)

		// Handle based on code
		if isRetryable {
			fmt.Println("→ This error is transient, retry recommended")
		} else {
			fmt.Println("→ This error requires manual intervention")
		}
	} else {
		// Regular error without code
		fmt.Printf("Error (no code): %v\n\n", err)
	}
}

// Real-world example: Production error handler
func productionErrorHandler(err error) {
	if err == nil {
		return
	}

	// Log error with code
	code := agent.GetErrorCode(err)
	log.Printf("[ERROR] Code: %s, Error: %v", code, err)

	// Metrics/monitoring
	if code != "" {
		// In production: increment error counter by code
		// metrics.IncrementCounter("agent.errors", map[string]string{"code": code})
		fmt.Printf("📊 Metrics: Incremented error counter for code: %s\n", code)
	}

	// Alerting for critical errors
	criticalCodes := []string{
		agent.ErrCodeAPIKeyMissing,
		agent.ErrCodeAPIKeyInvalid,
		agent.ErrCodeUnsupportedProvider,
	}

	for _, criticalCode := range criticalCodes {
		if code == criticalCode {
			// In production: send alert to PagerDuty/Slack
			// alerting.SendAlert("Critical error: " + code)
			fmt.Printf("🚨 ALERT: Critical error detected: %s\n", code)
			return
		}
	}

	// Auto-retry for retryable errors
	if agent.IsRetryable(err) {
		// In production: add to retry queue
		// retryQueue.Add(task)
		fmt.Printf("🔄 Retry: Error is retryable, adding to retry queue\n")
	}
}
