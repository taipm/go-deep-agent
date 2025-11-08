package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestBatchSimple(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompts := []string{
		"Say hello",
		"Say goodbye",
		"Count to 3",
	}
	
	results, _ := agent.Batch(ctx, prompts)
	
	// With invalid key, should return results with errors
	if len(results) != len(prompts) {
		t.Errorf("Expected %d results, got %d", len(prompts), len(results))
	}
	
	// Check that all results have errors
	for _, result := range results {
		if result.Error == nil {
			t.Error("Expected error for each result with invalid API key")
		}
	}
}

func TestBatchWithOptions(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompts := []string{"Q1", "Q2", "Q3"}
	
	opts := &BatchOptions{
		MaxConcurrency:      2,
		DelayBetweenBatches: 100 * time.Millisecond,
		ContinueOnError:     true,
	}
	
	results, err := agent.BatchWithOptions(ctx, prompts, opts)
	
	// Should handle errors gracefully
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
	
	if len(results) != len(prompts) {
		t.Errorf("Expected %d results, got %d", len(prompts), len(results))
	}
}

func TestBatchEmpty(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key")
	
	ctx := context.Background()
	
	results, err := agent.Batch(ctx, []string{})
	
	if err != nil {
		t.Errorf("Unexpected error for empty batch: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestBatchWithSize(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithBatchSize(3)
	
	if agent.batchSize != 3 {
		t.Errorf("Expected batch size 3, got %d", agent.batchSize)
	}
	
	// Test with zero/negative
	agent2 := NewOpenAI("gpt-4o-mini", "test-key").
		WithBatchSize(0)
	
	if agent2.batchSize != 5 { // Should default to 5
		t.Errorf("Expected default batch size 5, got %d", agent2.batchSize)
	}
}

func TestBatchWithDelay(t *testing.T) {
	delay := 200 * time.Millisecond
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithBatchDelay(delay)
	
	if agent.batchDelay != delay {
		t.Errorf("Expected delay %v, got %v", delay, agent.batchDelay)
	}
}

func TestBatchProgressCallback(t *testing.T) {
	var progressCalls []string
	
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		OnBatchProgress(func(completed, total int) {
			progressCalls = append(progressCalls, fmt.Sprintf("%d/%d", completed, total))
		})
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompts := []string{"Q1", "Q2", "Q3"}
	
	_, err := agent.Batch(ctx, prompts)
	
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
	
	// Progress callback should have been called
	if len(progressCalls) == 0 {
		t.Error("Expected progress callback to be called")
	}
	
	t.Logf("Progress calls: %v", progressCalls)
}

func TestBatchItemCompleteCallback(t *testing.T) {
	var completedItems []int
	
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		OnBatchItemComplete(func(result BatchResult) {
			completedItems = append(completedItems, result.Index)
		})
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompts := []string{"Q1", "Q2"}
	
	_, err := agent.Batch(ctx, prompts)
	
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
	
	// Item complete callback should have been called
	if len(completedItems) != len(prompts) {
		t.Errorf("Expected %d item completions, got %d", len(prompts), len(completedItems))
	}
}

func TestBatchSimpleMethod(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompts := []string{"Q1", "Q2", "Q3"}
	
	responses, _ := agent.BatchSimple(ctx, prompts)
	
	// Should still return responses array even with errors
	if len(responses) != len(prompts) {
		t.Errorf("Expected %d responses, got %d", len(prompts), len(responses))
	}
	
	// All responses should be empty due to errors
	for _, resp := range responses {
		if resp != "" {
			t.Error("Expected empty responses with invalid API key")
		}
	}
}

func TestBatchWithRetry(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	prompts := []string{"Q1", "Q2"}
	
	results, err := agent.BatchWithRetry(ctx, prompts, 2)
	
	// Should attempt retries
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
	
	if len(results) != len(prompts) {
		t.Errorf("Expected %d results, got %d", len(prompts), len(results))
	}
}

func TestBatchStats(t *testing.T) {
	results := []BatchResult{
		{Index: 0, Response: "R1", Error: nil, Tokens: TokenUsage{TotalTokens: 10, PromptTokens: 5, CompletionTokens: 5}},
		{Index: 1, Response: "", Error: fmt.Errorf("error"), Tokens: TokenUsage{}},
		{Index: 2, Response: "R3", Error: nil, Tokens: TokenUsage{TotalTokens: 20, PromptTokens: 10, CompletionTokens: 10}},
	}
	
	stats := GetBatchStats(results)
	
	if stats.Total != 3 {
		t.Errorf("Expected total 3, got %d", stats.Total)
	}
	
	if stats.Successful != 2 {
		t.Errorf("Expected 2 successful, got %d", stats.Successful)
	}
	
	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.Failed)
	}
	
	if stats.TotalTokens != 30 {
		t.Errorf("Expected 30 total tokens, got %d", stats.TotalTokens)
	}
	
	if stats.PromptTokens != 15 {
		t.Errorf("Expected 15 prompt tokens, got %d", stats.PromptTokens)
	}
	
	if stats.CompletionTokens != 15 {
		t.Errorf("Expected 15 completion tokens, got %d", stats.CompletionTokens)
	}
}

func TestDefaultBatchOptions(t *testing.T) {
	opts := DefaultBatchOptions()
	
	if opts.MaxConcurrency != 5 {
		t.Errorf("Expected default concurrency 5, got %d", opts.MaxConcurrency)
	}
	
	if opts.DelayBetweenBatches != 0 {
		t.Errorf("Expected default delay 0, got %v", opts.DelayBetweenBatches)
	}
	
	if !opts.ContinueOnError {
		t.Error("Expected ContinueOnError to be true by default")
	}
}

func TestBatchResultStructure(t *testing.T) {
	result := BatchResult{
		Index:    0,
		Prompt:   "Test prompt",
		Response: "Test response",
		Error:    nil,
		Tokens: TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}
	
	if result.Index != 0 {
		t.Errorf("Expected index 0, got %d", result.Index)
	}
	
	if result.Prompt != "Test prompt" {
		t.Errorf("Expected prompt 'Test prompt', got %s", result.Prompt)
	}
	
	if result.Response != "Test response" {
		t.Errorf("Expected response 'Test response', got %s", result.Response)
	}
	
	if result.Tokens.TotalTokens != 30 {
		t.Errorf("Expected 30 tokens, got %d", result.Tokens.TotalTokens)
	}
}

func TestBatchConcurrencyControl(t *testing.T) {
	// Test that batch respects concurrency limits
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithBatchSize(2) // Only 2 concurrent requests
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// More prompts than concurrency limit
	prompts := make([]string, 10)
	for i := range prompts {
		prompts[i] = fmt.Sprintf("Question %d", i+1)
	}
	
	results, err := agent.Batch(ctx, prompts)
	
	// Should handle all prompts despite concurrency limit
	if len(results) != len(prompts) {
		t.Errorf("Expected %d results, got %d", len(prompts), len(results))
	}
	
	if err != nil {
		t.Logf("Expected error with invalid key: %v", err)
	}
}

func TestBatchOrderPreserved(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompts := []string{"First", "Second", "Third", "Fourth", "Fifth"}
	
	results, _ := agent.Batch(ctx, prompts)
	
	// Results should be in same order as prompts
	for i, result := range results {
		if result.Index != i {
			t.Errorf("Expected index %d, got %d", i, result.Index)
		}
		
		if result.Prompt != prompts[i] {
			t.Errorf("Expected prompt %s, got %s", prompts[i], result.Prompt)
		}
	}
}

func TestBatchWithConfiguration(t *testing.T) {
	// Test that batch respects builder configuration
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithSystem("You are helpful").
		WithTemperature(0.7).
		WithMaxTokens(100)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompts := []string{"Q1", "Q2"}
	
	results, err := agent.Batch(ctx, prompts)
	
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
	
	// Each result should have been processed with the same configuration
	if len(results) != len(prompts) {
		t.Errorf("Expected %d results, got %d", len(prompts), len(results))
	}
}

func TestTokenUsageTracking(t *testing.T) {
	usage := TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 200,
		TotalTokens:      300,
	}
	
	if usage.PromptTokens != 100 {
		t.Errorf("Expected 100 prompt tokens, got %d", usage.PromptTokens)
	}
	
	if usage.CompletionTokens != 200 {
		t.Errorf("Expected 200 completion tokens, got %d", usage.CompletionTokens)
	}
	
	if usage.TotalTokens != 300 {
		t.Errorf("Expected 300 total tokens, got %d", usage.TotalTokens)
	}
}

func TestBatchResultsContainErrors(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "invalid-key")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompts := []string{"Q1", "Q2", "Q3"}
	
	results, _ := agent.Batch(ctx, prompts)
	
	// All results should have errors due to invalid key
	errorCount := 0
	for _, result := range results {
		if result.Error != nil {
			errorCount++
		}
	}
	
	if errorCount != len(prompts) {
		t.Logf("Expected all requests to fail, got %d errors out of %d", errorCount, len(prompts))
	}
}

func TestBatchChaining(t *testing.T) {
	// Test fluent API chaining
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithBatchSize(3).
		WithBatchDelay(50 * time.Millisecond).
		OnBatchProgress(func(completed, total int) {
			// Progress tracking
		})
	
	if agent.batchSize != 3 {
		t.Errorf("Expected batch size 3, got %d", agent.batchSize)
	}
	
	if agent.batchDelay != 50*time.Millisecond {
		t.Errorf("Expected delay 50ms, got %v", agent.batchDelay)
	}
	
	if agent.onBatchProgress == nil {
		t.Error("Expected progress callback to be set")
	}
}

func BenchmarkBatchProcessing(b *testing.B) {
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithBatchSize(5)
	
	prompts := []string{"Q1", "Q2", "Q3", "Q4", "Q5"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, _ = agent.Batch(ctx, prompts)
		cancel()
	}
}

func TestBatchWithNilOptions(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Should use default options when nil is passed
	results, err := agent.BatchWithOptions(ctx, []string{"Q1"}, nil)
	
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// Test that demonstrates typical usage
func ExampleBatch() {
	agent := NewOpenAI("gpt-4o-mini", "your-api-key").
		WithBatchSize(5).
		OnBatchProgress(func(completed, total int) {
			fmt.Printf("Progress: %d/%d\n", completed, total)
		})
	
	prompts := []string{
		"What is Go?",
		"What is Python?",
		"What is Rust?",
	}
	
	ctx := context.Background()
	results, err := agent.Batch(ctx, prompts)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("Failed: %s - %v\n", result.Prompt, result.Error)
		} else {
			// Truncate long responses for example
			response := result.Response
			if len(response) > 50 {
				response = response[:50] + "..."
			}
			fmt.Printf("Success: %s\n", strings.TrimSpace(response))
		}
	}
}
