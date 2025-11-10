// Package main demonstrates key error handling patterns in go-deep-agent:
//
//  1. Error Codes - Programmatic error handling via error codes
//  2. Debug Mode - Visibility into requests/responses with secret redaction
//  3. Panic Recovery - Automatic tool panic recovery for stability
//  4. Error Context - Rich error debugging with context and summarization
//  5. Error Chains - Track multiple errors in complex workflows
//
// For complete documentation, see docs/ERROR_HANDLING_BEST_PRACTICES.md
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Error Handling Patterns ===\n")

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable required")
	}

	// Pattern 1: Error code checking
	fmt.Println("Pattern 1: Error Code Checking")
	pattern1ErrorCodes(apiKey)

	// Pattern 2: Debug mode
	fmt.Println("\nPattern 2: Debug Mode")
	pattern2DebugMode(apiKey)

	// Pattern 3: Panic recovery
	fmt.Println("\nPattern 3: Panic Recovery")
	pattern3PanicRecovery(apiKey)

	// Pattern 4: Error context
	fmt.Println("\nPattern 4: Error Context and Summarization")
	pattern4ErrorContext()

	// Pattern 5: Error chains
	fmt.Println("\nPattern 5: Error Chains")
	pattern5ErrorChains()

	fmt.Println("\n✅ All patterns demonstrated!")
}

// Pattern 1: Use error codes for programmatic decisions
func pattern1ErrorCodes(apiKey string) {
	ctx := context.Background()

	// Create agent
	b := agent.NewOpenAI("gpt-4o-mini", apiKey).WithDefaults()

	// Execute request
	resp, err := b.Ask(ctx, "What is 2+2?")
	if err != nil {
		// Check error code
		code := agent.GetErrorCode(err)
		fmt.Printf("Error code: %s\n", code)

		// Make decision based on code
		switch code {
		case agent.ErrCodeRateLimitExceeded:
			fmt.Println("Decision: Wait and retry")
		case agent.ErrCodeRequestTimeout:
			fmt.Println("Decision: Retry with longer timeout")
		case agent.ErrCodeAPIKeyMissing:
			fmt.Println("Decision: Fatal - cannot proceed")
		default:
			fmt.Printf("Decision: Unknown error [%s]\n", code)
		}
		return
	}

	fmt.Printf("Response: %s\n", resp)
}

// Pattern 2: Use debug mode for visibility
func pattern2DebugMode(apiKey string) {
	ctx := context.Background()

	// Enable debug mode (secrets auto-redacted)
	debugConfig := agent.DefaultDebugConfig()

	b := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDefaults().
		WithDebug(debugConfig)

	resp, err := b.Ask(ctx, "What is the capital of France?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", resp)
}

// Pattern 3: Panic recovery (automatic for tools)
func pattern3PanicRecovery(apiKey string) {
	ctx := context.Background()

	b := agent.NewOpenAI("gpt-4o-mini", apiKey).WithDefaults()

	// Add tool that might panic
	b.WithTool("calculator", "Calculate result", func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		// This panic is automatically recovered!
		if args["divide_by_zero"] == true {
			var x int
			return 1 / x, nil // Panic: integer divide by zero
		}
		return "result", nil
	})

	resp, err := b.Ask(ctx, "Calculate something")
	if err != nil {
		// Check if panic occurred
		if agent.IsPanicError(err) {
			fmt.Println("✓ Panic recovered automatically")
			fmt.Printf("Panic value: %v\n", agent.GetPanicValue(err))
			// Stack trace available via GetStackTrace(err)
		} else {
			fmt.Printf("Normal error: %v\n", err)
		}
		return
	}

	fmt.Printf("Response: %s\n", resp)
}

// Pattern 4: Add error context for debugging
func pattern4ErrorContext(apiKey string) {
	// Simulate error from nested function
	err := callNestedFunction("user123")

	if err != nil {
		// Summarize error for logging/monitoring
		summary := agent.SummarizeError(err)
		if summary != nil {
			fmt.Printf("Error Type: %s\n", summary.Type)
			fmt.Printf("Error Code: %s\n", summary.Code)
			fmt.Printf("Message: %s\n", summary.Message)
			fmt.Printf("Retryable: %v\n", summary.Retryable)

			// Context contains operation details
			fmt.Println("Context:")
			for key, value := range summary.Context {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}
}

func callNestedFunction(userID string) error {
	// Simulate error
	baseErr := fmt.Errorf("connection failed")

	// Add context as error bubbles up
	return agent.WithContext(baseErr, "user data fetch", map[string]interface{}{
		"user_id":   userID,
		"timestamp": time.Now().Unix(),
		"retry":     2,
	})
}

// Pattern 5: Smart retry with backoff
func pattern5SmartRetry(apiKey string) {
	ctx := context.Background()

	b := agent.NewOpenAI("gpt-4o-mini", apiKey).WithDefaults()

	// Retry with exponential backoff
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := b.Ask(ctx, "What is 2+2?")
		if err == nil {
			fmt.Printf("✓ Succeeded (attempt %d)\n", attempt+1)
			fmt.Printf("Response: %s\n", resp)
			return
		}

		lastErr = err

		// Check if retryable
		if !agent.IsRetryableError(err) {
			fmt.Println("✗ Non-retryable error")
			break
		}

		// Don't retry on last attempt
		if attempt == maxRetries {
			break
		}

		// Exponential backoff: 1s, 2s, 4s
		backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		fmt.Printf("⏱ Retrying in %v (attempt %d/%d)...\n", backoff, attempt+1, maxRetries)
		time.Sleep(backoff)
	}

	fmt.Printf("✗ Failed after %d attempts: %v\n", maxRetries+1, lastErr)
}
