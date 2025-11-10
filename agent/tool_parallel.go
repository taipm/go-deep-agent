package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/openai/openai-go/v3"
)

// executeToolsParallel executes tools in parallel using a worker pool.
func (b *Builder) executeToolsParallel(ctx context.Context, toolCalls []openai.ChatCompletionMessageToolCallUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	logger := b.getLogger()

	if len(toolCalls) == 0 {
		return nil, nil
	}

	// Sequential execution if only 1 tool or parallel disabled
	if !b.enableParallel || len(toolCalls) == 1 {
		return b.executeToolsSequential(ctx, toolCalls)
	}

	// Tool execution result
	type toolResult struct {
		index    int
		toolCall openai.ChatCompletionMessageToolCallUnion
		result   string
		err      error
		duration time.Duration
	}

	// Create result channel
	results := make(chan toolResult, len(toolCalls))

	// Worker pool semaphore
	maxWorkers := b.maxWorkers
	if maxWorkers == 0 {
		maxWorkers = 10
	}
	if len(toolCalls) < maxWorkers {
		maxWorkers = len(toolCalls)
	}

	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	// Launch tool executions
	for i, tc := range toolCalls {
		wg.Add(1)
		sem <- struct{}{} // Acquire worker

		go func(index int, toolCall openai.ChatCompletionMessageToolCallUnion) {
			defer wg.Done()
			defer func() { <-sem }() // Release worker

			start := time.Now()
			result, err := b.executeOneTool(ctx, toolCall)
			duration := time.Since(start)

			results <- toolResult{
				index:    index,
				toolCall: toolCall,
				result:   result,
				err:      err,
				duration: duration,
			}
		}(i, tc)
	}

	// Wait for all to complete
	wg.Wait()
	close(results)

	// Collect results in original order
	resultMap := make(map[int]toolResult)
	var totalDuration time.Duration
	successCount := 0
	failureCount := 0

	for r := range results {
		resultMap[r.index] = r
		totalDuration += r.duration
		if r.err == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	// Log parallel execution stats
	logger.Info(ctx, "Parallel tool execution completed",
		F("total_tools", len(toolCalls)),
		F("success_count", successCount),
		F("failure_count", failureCount),
		F("total_duration_ms", totalDuration.Milliseconds()),
		F("max_workers", maxWorkers))

	// Convert to messages in original order
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(toolCalls))
	for i := 0; i < len(toolCalls); i++ {
		r := resultMap[i]

		if r.err != nil {
			logger.Error(ctx, "Tool execution failed",
				F("tool_name", r.toolCall.Function.Name),
				F("error", r.err.Error()),
				F("duration_ms", r.duration.Milliseconds()))

			// Return error immediately (consistent with sequential behavior)
			return nil, fmt.Errorf("tool execution failed (%s): %w", r.toolCall.Function.Name, r.err)
		}

		logger.Debug(ctx, "Tool execution succeeded",
			F("tool_name", r.toolCall.Function.Name),
			F("result_length", len(r.result)),
			F("duration_ms", r.duration.Milliseconds()))

		messages = append(messages, openai.ToolMessage(r.result, r.toolCall.ID))
	}

	return messages, nil
}

// executeToolsSequential executes tools one by one.
func (b *Builder) executeToolsSequential(ctx context.Context, toolCalls []openai.ChatCompletionMessageToolCallUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	logger := b.getLogger()
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(toolCalls))

	for _, toolCall := range toolCalls {
		start := time.Now()
		result, err := b.executeOneTool(ctx, toolCall)
		duration := time.Since(start)

		toolName := toolCall.Function.Name

		if err != nil {
			logger.Error(ctx, "Tool execution failed",
				F("tool_name", toolName),
				F("error", err.Error()),
				F("duration_ms", duration.Milliseconds()))
			return nil, fmt.Errorf("tool execution failed (%s): %w", toolName, err)
		}

		logger.Debug(ctx, "Tool execution succeeded",
			F("tool_name", toolName),
			F("result_length", len(result)),
			F("duration_ms", duration.Milliseconds()))

		messages = append(messages, openai.ToolMessage(result, toolCall.ID))
	}

	return messages, nil
}

// executeOneTool executes a single tool with timeout.
func (b *Builder) executeOneTool(ctx context.Context, toolCall openai.ChatCompletionMessageToolCallUnion) (string, error) {
	logger := b.getLogger()
	toolName := toolCall.Function.Name

	// Find handler
	var handler func(string) (string, error)
	for _, tool := range b.tools {
		if tool.Name == toolName {
			handler = tool.Handler
			break
		}
	}

	if handler == nil {
		return "", fmt.Errorf("no handler found for tool: %s", toolName)
	}

	// Apply timeout if configured
	timeout := b.toolTimeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute in goroutine to support timeout
	done := make(chan struct{})
	var result string
	var err error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("tool panicked: %v", r)
			}
			close(done)
		}()

		logger.Debug(execCtx, "Executing tool",
			F("tool_name", toolName),
			F("args_length", len(toolCall.Function.Arguments)))

		result, err = handler(toolCall.Function.Arguments)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		return result, err
	case <-execCtx.Done():
		return "", fmt.Errorf("tool execution timeout after %v", timeout)
	}
}
