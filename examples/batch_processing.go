package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	fmt.Println("=== Batch Processing Examples ===\n")

	// Example 1: Simple Batch Processing
	fmt.Println("1. Simple Batch Processing:")
	simpleBatch(apiKey)

	// Example 2: Batch with Progress Tracking
	fmt.Println("\n2. Batch with Progress Tracking:")
	batchWithProgress(apiKey)

	// Example 3: Batch with Concurrency Control
	fmt.Println("\n3. Batch with Concurrency Control:")
	batchWithConcurrency(apiKey)

	// Example 4: Batch with Error Handling
	fmt.Println("\n4. Batch with Individual Item Callbacks:")
	batchWithCallbacks(apiKey)

	// Example 5: Batch Statistics
	fmt.Println("\n5. Batch Statistics:")
	batchStats(apiKey)
}

// Example 1: Simple batch processing
func simpleBatch(apiKey string) {
	ctx := context.Background()

	// Create agent
	myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey)

	// Multiple prompts to process
	prompts := []string{
		"What is Go programming language in one sentence?",
		"What is Python in one sentence?",
		"What is Rust in one sentence?",
	}

	// Process all prompts in parallel
	results, err := myAgent.Batch(ctx, prompts)
	if err != nil {
		log.Printf("Batch error: %v\n", err)
		return
	}

	// Display results
	for i, result := range results {
		if result.Error != nil {
			fmt.Printf("  [%d] Error: %v\n", i+1, result.Error)
		} else {
			fmt.Printf("  [%d] %s\n", i+1, result.Response)
		}
	}
}

// Example 2: Batch with progress tracking
func batchWithProgress(apiKey string) {
	ctx := context.Background()

	// Create agent with progress callback
	myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		OnBatchProgress(func(completed, total int) {
			fmt.Printf("  Progress: %d/%d (%.1f%%)\n",
				completed, total, float64(completed)/float64(total)*100)
		})

	prompts := []string{
		"Say 'Hello'",
		"Say 'World'",
		"Say 'From'",
		"Say 'Go'",
		"Say 'Batch'",
	}

	results, err := myAgent.Batch(ctx, prompts)
	if err != nil {
		log.Printf("Batch error: %v\n", err)
		return
	}

	fmt.Printf("  Completed %d requests\n", len(results))
}

// Example 3: Batch with concurrency control
func batchWithConcurrency(apiKey string) {
	ctx := context.Background()

	// Limit to 3 concurrent requests with delay between batches
	myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithBatchSize(3).
		WithBatchDelay(500 * time.Millisecond)

	prompts := make([]string, 10)
	for i := range prompts {
		prompts[i] = fmt.Sprintf("Count to %d", i+1)
	}

	start := time.Now()
	results, err := myAgent.Batch(ctx, prompts)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Batch error: %v\n", err)
		return
	}

	fmt.Printf("  Processed %d prompts in %v\n", len(results), duration)
	fmt.Printf("  Concurrency limit: 3 requests at a time\n")
}

// Example 4: Batch with individual item callbacks
func batchWithCallbacks(apiKey string) {
	ctx := context.Background()

	successCount := 0
	errorCount := 0

	myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		OnBatchItemComplete(func(result agent.BatchResult) {
			if result.Error != nil {
				errorCount++
				fmt.Printf("  ✗ Item %d failed: %v\n", result.Index+1, result.Error)
			} else {
				successCount++
				fmt.Printf("  ✓ Item %d completed (%d tokens)\n",
					result.Index+1, result.Tokens.TotalTokens)
			}
		})

	prompts := []string{
		"Hello",
		"How are you?",
		"Tell me a joke",
	}

	_, err := myAgent.Batch(ctx, prompts)
	if err != nil {
		log.Printf("Batch error: %v\n", err)
	}

	fmt.Printf("\n  Summary: %d succeeded, %d failed\n", successCount, errorCount)
}

// Example 5: Batch statistics
func batchStats(apiKey string) {
	ctx := context.Background()

	myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey)

	prompts := []string{
		"Explain Go in 5 words",
		"Explain Python in 5 words",
		"Explain Rust in 5 words",
		"Explain Java in 5 words",
		"Explain C++ in 5 words",
	}

	results, err := myAgent.Batch(ctx, prompts)
	if err != nil {
		log.Printf("Batch error: %v\n", err)
		return
	}

	// Get batch statistics
	stats := agent.GetBatchStats(results)

	fmt.Printf("  Total Requests: %d\n", stats.Total)
	fmt.Printf("  Successful: %d\n", stats.Successful)
	fmt.Printf("  Failed: %d\n", stats.Failed)
	fmt.Printf("  Total Tokens: %d\n", stats.TotalTokens)
	fmt.Printf("  Prompt Tokens: %d\n", stats.PromptTokens)
	fmt.Printf("  Completion Tokens: %d\n", stats.CompletionTokens)
	
	if stats.Successful > 0 {
		avgTokens := float64(stats.TotalTokens) / float64(stats.Successful)
		fmt.Printf("  Avg Tokens/Request: %.1f\n", avgTokens)
	}
}

// Example 6: Batch with custom options
func batchWithOptions(apiKey string) {
	ctx := context.Background()

	myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey)

	prompts := []string{
		"Question 1",
		"Question 2",
		"Question 3",
	}

	// Create custom batch options
	opts := &agent.BatchOptions{
		MaxConcurrency:      2,
		DelayBetweenBatches: 1 * time.Second,
		ContinueOnError:     true,
		OnProgress: func(completed, total int) {
			fmt.Printf("Custom progress: %d/%d\n", completed, total)
		},
	}

	results, err := myAgent.BatchWithOptions(ctx, prompts, opts)
	if err != nil {
		log.Printf("Batch error: %v\n", err)
		return
	}

	fmt.Printf("Completed %d requests\n", len(results))
}

// Example 7: BatchSimple for quick results
func batchSimple(apiKey string) {
	ctx := context.Background()

	myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey)

	prompts := []string{
		"Say hello",
		"Say goodbye",
		"Say thank you",
	}

	// BatchSimple returns only responses (empty string for errors)
	responses, err := myAgent.BatchSimple(ctx, prompts)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	for i, response := range responses {
		if response == "" {
			fmt.Printf("[%d] Failed\n", i+1)
		} else {
			fmt.Printf("[%d] %s\n", i+1, response)
		}
	}
}

// Example 8: Batch with retry
func batchWithRetry(apiKey string) {
	ctx := context.Background()

	myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey)

	prompts := []string{
		"Question 1",
		"Question 2",
		"Question 3",
	}

	// Automatically retry failed requests up to 3 times
	results, err := myAgent.BatchWithRetry(ctx, prompts, 3)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Completed %d requests with retries\n", len(results))
}
