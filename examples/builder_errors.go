package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable")
		return
	}

	// Run all examples
	fmt.Println("=== Error Handling & Recovery Examples ===\n")

	example1_BasicTimeout(apiKey)
	example2_RetryWithFixedDelay(apiKey)
	example3_RetryWithExponentialBackoff(apiKey)
	example4_ErrorTypeChecking(apiKey)
	example5_TimeoutWithRetry(apiKey)
	example6_ProductionReadyExample(apiKey)
}

// Example 1: Basic timeout handling
func example1_BasicTimeout(apiKey string) {
	fmt.Println("--- Example 1: Basic Timeout ---")

	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTimeout(5 * time.Second) // 5 second timeout

	ctx := context.Background()

	// This should work fine (quick response expected)
	response, err := builder.Ask(ctx, "Say 'Hello' in one word")
	if err != nil {
		if agent.IsTimeoutError(err) {
			fmt.Println("Request timed out!")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	fmt.Println()
}

// Example 2: Retry with fixed delay
func example2_RetryWithFixedDelay(apiKey string) {
	fmt.Println("--- Example 2: Retry with Fixed Delay ---")

	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRetry(3).                       // Retry up to 3 times
		WithRetryDelay(2 * time.Second).    // Wait 2 seconds between retries
		WithTimeout(30 * time.Second)       // Overall timeout

	ctx := context.Background()

	fmt.Println("Asking a question (will retry on transient errors)...")
	response, err := builder.Ask(ctx, "What is the capital of France?")
	if err != nil {
		if agent.IsMaxRetriesError(err) {
			fmt.Println("Failed after all retry attempts")
		} else if agent.IsTimeoutError(err) {
			fmt.Println("Request timed out")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	fmt.Println()
}

// Example 3: Retry with exponential backoff
func example3_RetryWithExponentialBackoff(apiKey string) {
	fmt.Println("--- Example 3: Retry with Exponential Backoff ---")

	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRetry(5).                       // Retry up to 5 times
		WithRetryDelay(time.Second).        // Base delay: 1 second
		WithExponentialBackoff().           // Enable exponential backoff (1s, 2s, 4s, 8s, 16s)
		WithTimeout(60 * time.Second)       // Overall timeout

	ctx := context.Background()

	fmt.Println("Asking a question with exponential backoff...")
	fmt.Println("Retry delays: 1s, 2s, 4s, 8s, 16s")
	
	response, err := builder.Ask(ctx, "Explain quantum computing briefly")
	if err != nil {
		if agent.IsMaxRetriesError(err) {
			fmt.Println("Failed after exponential backoff retries")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	fmt.Println()
}

// Example 4: Error type checking
func example4_ErrorTypeChecking(apiKey string) {
	fmt.Println("--- Example 4: Error Type Checking ---")

	// Test with invalid API key
	fmt.Println("Testing with invalid API key...")
	builder := agent.NewOpenAI("gpt-4o-mini", "invalid-key").
		WithRetry(2).
		WithTimeout(10 * time.Second)

	_, err := builder.Ask(context.Background(), "Hello")

	if err != nil {
		// Check different error types
		if agent.IsAPIKeyError(err) {
			fmt.Println("✗ API key error detected (as expected)")
		} else if agent.IsRateLimitError(err) {
			fmt.Println("✗ Rate limit error")
		} else if agent.IsTimeoutError(err) {
			fmt.Println("✗ Timeout error")
		} else if agent.IsRefusalError(err) {
			fmt.Println("✗ Content refused")
		} else if agent.IsMaxRetriesError(err) {
			fmt.Println("✗ Max retries exceeded")
		} else {
			fmt.Printf("✗ Other error: %v\n", err)
		}
	}

	fmt.Println()
}

// Example 5: Timeout with retry (realistic scenario)
func example5_TimeoutWithRetry(apiKey string) {
	fmt.Println("--- Example 5: Timeout with Retry ---")

	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTimeout(10 * time.Second).      // 10s timeout per request
		WithRetry(2).                        // Retry up to 2 times
		WithRetryDelay(time.Second)          // 1s delay between retries

	ctx := context.Background()

	fmt.Println("Making request with timeout + retry...")
	start := time.Now()

	response, err := builder.Ask(ctx, "List 3 programming languages")
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Failed after %.2f seconds\n", elapsed.Seconds())
		if agent.IsTimeoutError(err) {
			fmt.Println("Reason: Timeout")
		} else if agent.IsMaxRetriesError(err) {
			fmt.Println("Reason: Max retries exceeded")
		} else {
			fmt.Printf("Reason: %v\n", err)
		}
	} else {
		fmt.Printf("Success in %.2f seconds\n", elapsed.Seconds())
		fmt.Printf("Response: %s\n", response)
	}

	fmt.Println()
}

// Example 6: Production-ready configuration
func example6_ProductionReadyExample(apiKey string) {
	fmt.Println("--- Example 6: Production-Ready Configuration ---")

	// Production-ready builder with comprehensive error handling
	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant. Be concise.").
		WithMemory().
		WithMaxHistory(20).                  // Limit memory
		WithTimeout(30 * time.Second).       // 30s timeout
		WithRetry(3).                        // 3 retries
		WithRetryDelay(2 * time.Second).     // 2s base delay
		WithExponentialBackoff().            // Exponential backoff
		WithTemperature(0.7)                 // Balanced creativity

	ctx := context.Background()

	fmt.Println("Using production-ready configuration:")
	fmt.Println("  - Memory: enabled (max 20 messages)")
	fmt.Println("  - Timeout: 30 seconds")
	fmt.Println("  - Retry: 3 attempts with exponential backoff")
	fmt.Println("  - Temperature: 0.7")
	fmt.Println()

	// Make requests with error handling
	questions := []string{
		"What is Go programming?",
		"Why is it popular?",
		"Give me one advantage",
	}

	for i, question := range questions {
		fmt.Printf("Q%d: %s\n", i+1, question)
		
		response, err := builder.Ask(ctx, question)
		
		if err != nil {
			// Handle different error types
			fmt.Printf("  Error: ")
			switch {
			case agent.IsAPIKeyError(err):
				fmt.Println("Invalid API key - check configuration")
				return // Fatal error, stop
			case agent.IsRateLimitError(err):
				fmt.Println("Rate limited - waiting before retry...")
				time.Sleep(5 * time.Second)
				continue // Skip this question, try next
			case agent.IsTimeoutError(err):
				fmt.Println("Request timed out - question too complex?")
				continue
			case agent.IsMaxRetriesError(err):
				fmt.Println("Failed after all retries - service may be down")
				continue
			default:
				fmt.Printf("Unexpected error: %v\n", err)
				continue
			}
		}

		fmt.Printf("  A: %s\n", response)
		fmt.Println()
	}

	fmt.Println("Production example completed!")
	fmt.Println()
}

// Example 7: Custom error handling wrapper (bonus)
func robustAsk(builder *agent.Builder, ctx context.Context, message string) (string, error) {
	response, err := builder.Ask(ctx, message)
	
	if err != nil {
		// Log error with context
		fmt.Printf("[ERROR] %v\n", err)
		
		// Handle specific cases
		switch {
		case agent.IsAPIKeyError(err):
			return "", fmt.Errorf("authentication failed: %w", err)
		case agent.IsRateLimitError(err):
			return "", fmt.Errorf("rate limit exceeded, try again later: %w", err)
		case agent.IsTimeoutError(err):
			return "", fmt.Errorf("request took too long: %w", err)
		case agent.IsMaxRetriesError(err):
			return "", fmt.Errorf("service unavailable after retries: %w", err)
		default:
			return "", fmt.Errorf("request failed: %w", err)
		}
	}
	
	return response, nil
}
