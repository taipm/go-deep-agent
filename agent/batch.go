package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenUsage represents token usage statistics
type TokenUsage struct {
	PromptTokens       int
	CompletionTokens   int
	TotalTokens        int
	PromptCachedTokens int // Cached tokens (for providers like Anthropic that support caching)
}

// EstimateCost estimates the cost in USD for this token usage.
// This is a helper method for common cost calculation.
//
// Parameters:
//   - promptPricePer1M: Price per 1 million prompt tokens in USD
//   - completionPricePer1M: Price per 1 million completion tokens in USD
//
// Returns estimated cost in USD.
//
// Example:
//
//	cost := usage.EstimateCost(0.50, 1.50) // Gemini Pro pricing
//	fmt.Printf("Estimated cost: $%.4f\n", cost)
func (u TokenUsage) EstimateCost(promptPricePer1M, completionPricePer1M float64) float64 {
	promptCost := float64(u.PromptTokens) / 1_000_000.0 * promptPricePer1M
	completionCost := float64(u.CompletionTokens) / 1_000_000.0 * completionPricePer1M
	return promptCost + completionCost
}

// String returns a human-readable representation of token usage.
//
// Example output: "Tokens: 150 prompt + 75 completion = 225 total"
func (u TokenUsage) String() string {
	return fmt.Sprintf("Tokens: %d prompt + %d completion = %d total",
		u.PromptTokens, u.CompletionTokens, u.TotalTokens)
}

// BatchResult represents the result of a single batch request
type BatchResult struct {
	Index    int        // Original index in the batch
	Prompt   string     // Original prompt
	Response string     // LLM response
	Error    error      // Error if request failed
	Tokens   TokenUsage // Token usage for this request
}

// BatchOptions configures batch processing behavior
type BatchOptions struct {
	// MaxConcurrency limits concurrent requests (default: 5)
	MaxConcurrency int

	// DelayBetweenBatches adds delay between batch chunks (default: 0)
	DelayBetweenBatches time.Duration

	// ContinueOnError continues processing if individual requests fail (default: true)
	ContinueOnError bool

	// OnProgress callback for tracking progress (optional)
	OnProgress func(completed, total int)

	// OnItemComplete callback when each item completes (optional)
	OnItemComplete func(result BatchResult)
}

// DefaultBatchOptions returns default batch processing options
func DefaultBatchOptions() *BatchOptions {
	return &BatchOptions{
		MaxConcurrency:      5,
		DelayBetweenBatches: 0,
		ContinueOnError:     true,
		OnProgress:          nil,
		OnItemComplete:      nil,
	}
}

// WithBatchSize sets the maximum concurrent requests for batch processing
func (b *Builder) WithBatchSize(size int) *Builder {
	if size <= 0 {
		size = 5 // Default to 5
	}
	b.batchSize = size
	return b
}

// WithBatchDelay sets delay between batch chunks
func (b *Builder) WithBatchDelay(delay time.Duration) *Builder {
	b.batchDelay = delay
	return b
}

// OnBatchProgress sets a callback for batch progress tracking
func (b *Builder) OnBatchProgress(fn func(completed, total int)) *Builder {
	b.onBatchProgress = fn
	return b
}

// OnBatchItemComplete sets a callback when each batch item completes
func (b *Builder) OnBatchItemComplete(fn func(result BatchResult)) *Builder {
	b.onBatchItemComplete = fn
	return b
}

// Batch processes multiple prompts concurrently and returns all results
// Results are returned in the same order as the input prompts
func (b *Builder) Batch(ctx context.Context, prompts []string) ([]BatchResult, error) {
	if len(prompts) == 0 {
		return []BatchResult{}, nil
	}

	opts := DefaultBatchOptions()
	if b.batchSize > 0 {
		opts.MaxConcurrency = b.batchSize
	}
	if b.batchDelay > 0 {
		opts.DelayBetweenBatches = b.batchDelay
	}
	if b.onBatchProgress != nil {
		opts.OnProgress = b.onBatchProgress
	}
	if b.onBatchItemComplete != nil {
		opts.OnItemComplete = b.onBatchItemComplete
	}

	return b.BatchWithOptions(ctx, prompts, opts)
}

// BatchWithOptions processes multiple prompts with custom options
func (b *Builder) BatchWithOptions(ctx context.Context, prompts []string, opts *BatchOptions) ([]BatchResult, error) {
	if opts == nil {
		opts = DefaultBatchOptions()
	}

	total := len(prompts)
	results := make([]BatchResult, total)

	// Channel for work items
	type workItem struct {
		index  int
		prompt string
	}

	workChan := make(chan workItem, total)
	resultChan := make(chan BatchResult, total)

	// Fill work channel
	for i, prompt := range prompts {
		workChan <- workItem{index: i, prompt: prompt}
	}
	close(workChan)

	// Worker pool
	var wg sync.WaitGroup
	completed := 0
	var completedMu sync.Mutex

	for i := 0; i < opts.MaxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for work := range workChan {
				// Add delay if specified and not first batch
				if opts.DelayBetweenBatches > 0 && work.index > 0 {
					time.Sleep(opts.DelayBetweenBatches)
				}

				// Create a new builder for this request to ensure thread-safety
				// Copy only the essential configuration
				agentCopy := NewOpenAI(b.model, b.apiKey)
				if b.baseURL != "" {
					agentCopy.baseURL = b.baseURL
				}
				if b.systemPrompt != "" {
					agentCopy = agentCopy.WithSystem(b.systemPrompt)
				}
				if b.temperature != nil {
					agentCopy = agentCopy.WithTemperature(*b.temperature)
				}
				if b.maxTokens != nil {
					agentCopy = agentCopy.WithMaxTokens(*b.maxTokens)
				}
				if b.topP != nil {
					agentCopy = agentCopy.WithTopP(*b.topP)
				}
				if b.maxRetries > 0 {
					agentCopy = agentCopy.WithRetry(b.maxRetries)
				}

				// Copy tools
				for _, tool := range b.tools {
					agentCopy.tools = append(agentCopy.tools, tool)
				}
				if b.autoExecute {
					agentCopy = agentCopy.WithAutoExecute(true)
				}

				// Execute request
				response, err := agentCopy.Ask(ctx, work.prompt)

				result := BatchResult{
					Index:    work.index,
					Prompt:   work.prompt,
					Response: response,
					Error:    err,
					Tokens:   agentCopy.lastUsage,
				}

				resultChan <- result

				// Update progress
				completedMu.Lock()
				completed++
				currentCompleted := completed
				completedMu.Unlock()

				// Call progress callback
				if opts.OnProgress != nil {
					opts.OnProgress(currentCompleted, total)
				}

				// Call item complete callback
				if opts.OnItemComplete != nil {
					opts.OnItemComplete(result)
				}

				// Check if we should stop on error
				if err != nil && !opts.ContinueOnError {
					// Drain work channel to stop other workers
					go func() {
						for range workChan {
							// Drain remaining work
						}
					}()
					return
				}
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	for result := range resultChan {
		results[result.Index] = result
	}

	// Check if any errors occurred and ContinueOnError is false
	if !opts.ContinueOnError {
		for _, result := range results {
			if result.Error != nil {
				return results, fmt.Errorf("batch processing failed: %w", result.Error)
			}
		}
	}

	return results, nil
}

// BatchSimple is a convenience method that processes prompts and returns only successful responses
// Failed requests are skipped. Returns responses in order (empty string for failed requests)
func (b *Builder) BatchSimple(ctx context.Context, prompts []string) ([]string, error) {
	results, err := b.Batch(ctx, prompts)
	if err != nil {
		return nil, err
	}

	responses := make([]string, len(results))
	for i, result := range results {
		if result.Error == nil {
			responses[i] = result.Response
		} else {
			responses[i] = "" // Empty string for failed requests
		}
	}

	return responses, nil
}

// BatchWithRetry processes prompts with automatic retry for failed requests
func (b *Builder) BatchWithRetry(ctx context.Context, prompts []string, maxRetries int) ([]BatchResult, error) {
	if maxRetries <= 0 {
		return b.Batch(ctx, prompts)
	}

	results, err := b.Batch(ctx, prompts)
	if err != nil {
		return results, err
	}

	// Retry failed requests
	for retry := 0; retry < maxRetries; retry++ {
		var failedPrompts []string
		var failedIndices []int

		for i, result := range results {
			if result.Error != nil {
				failedPrompts = append(failedPrompts, result.Prompt)
				failedIndices = append(failedIndices, i)
			}
		}

		if len(failedPrompts) == 0 {
			break // All succeeded
		}

		// Retry failed requests
		retryResults, _ := b.Batch(ctx, failedPrompts)

		// Update results
		for i, idx := range failedIndices {
			if retryResults[i].Error == nil {
				results[idx] = retryResults[i]
				results[idx].Index = idx // Preserve original index
			}
		}
	}

	return results, nil
}

// GetBatchStats returns statistics about batch results
type BatchStats struct {
	Total            int
	Successful       int
	Failed           int
	TotalTokens      int
	PromptTokens     int
	CompletionTokens int
	AverageTime      time.Duration
}

// GetBatchStats computes statistics from batch results
func GetBatchStats(results []BatchResult) BatchStats {
	stats := BatchStats{
		Total: len(results),
	}

	for _, result := range results {
		if result.Error == nil {
			stats.Successful++
			stats.TotalTokens += result.Tokens.TotalTokens
			stats.PromptTokens += result.Tokens.PromptTokens
			stats.CompletionTokens += result.Tokens.CompletionTokens
		} else {
			stats.Failed++
		}
	}

	return stats
}
