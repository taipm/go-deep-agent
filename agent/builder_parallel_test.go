package agent

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
)

// TestParallelToolExecution tests that parallel execution is faster than sequential
func TestParallelToolExecution(t *testing.T) {
	var executionCount atomic.Int32

	// Create tools with artificial delay
	tool1 := &Tool{
		Name:        "tool1",
		Description: "Test tool 1",
		Handler: func(args string) (string, error) {
			executionCount.Add(1)
			time.Sleep(50 * time.Millisecond)
			return "result1", nil
		},
	}
	tool2 := &Tool{
		Name:        "tool2",
		Description: "Test tool 2",
		Handler: func(args string) (string, error) {
			executionCount.Add(1)
			time.Sleep(50 * time.Millisecond)
			return "result2", nil
		},
	}
	tool3 := &Tool{
		Name:        "tool3",
		Description: "Test tool 3",
		Handler: func(args string) (string, error) {
			executionCount.Add(1)
			time.Sleep(50 * time.Millisecond)
			return "result3", nil
		},
	}

	agent := &Builder{
		tools:          []*Tool{tool1, tool2, tool3},
		enableParallel: true,
		maxWorkers:     10,
		toolTimeout:    30 * time.Second,
	}

	// Create tool calls
	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool1", Arguments: "{}"}},
		{ID: "call2", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool2", Arguments: "{}"}},
		{ID: "call3", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool3", Arguments: "{}"}},
	}

	start := time.Now()
	results, err := agent.executeToolsParallel(context.Background(), toolCalls)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Parallel execution failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	if executionCount.Load() != 3 {
		t.Fatalf("Expected 3 executions, got %d", executionCount.Load())
	}

	// Parallel should take ~50ms (all execute concurrently)
	// Sequential would take ~150ms (one after another)
	if duration > 100*time.Millisecond {
		t.Fatalf("Parallel execution too slow: %v (expected ~50ms)", duration)
	}

	t.Logf("✓ Parallel execution completed in %v (3 tools in ~50ms)", duration)
}

// TestSequentialToolExecution tests sequential execution mode
func TestSequentialToolExecution(t *testing.T) {
	var executionOrder []string
	var mu sync.Mutex

	tool1 := &Tool{
		Name: "tool1",
		Handler: func(args string) (string, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "tool1")
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return "result1", nil
		},
	}
	tool2 := &Tool{
		Name: "tool2",
		Handler: func(args string) (string, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "tool2")
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return "result2", nil
		},
	}

	agent := &Builder{
		tools:          []*Tool{tool1, tool2},
		enableParallel: false,
		toolTimeout:    30 * time.Second,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool1", Arguments: "{}"}},
		{ID: "call2", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "tool2", Arguments: "{}"}},
	}

	results, err := agent.executeToolsSequential(context.Background(), toolCalls)
	if err != nil {
		t.Fatalf("Sequential execution failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Verify sequential order
	if len(executionOrder) != 2 || executionOrder[0] != "tool1" || executionOrder[1] != "tool2" {
		t.Fatalf("Execution order incorrect: %v", executionOrder)
	}

	t.Logf("✓ Sequential execution preserved order: %v", executionOrder)
}

// TestWorkerPoolLimit tests that max workers is enforced
func TestWorkerPoolLimit(t *testing.T) {
	var concurrentCount atomic.Int32
	var maxConcurrent atomic.Int32

	checkConcurrency := func() {
		current := concurrentCount.Add(1)
		defer concurrentCount.Add(-1)

		// Update max if current is higher
		for {
			max := maxConcurrent.Load()
			if current <= max || maxConcurrent.CompareAndSwap(max, current) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Create 5 tools but limit workers to 2
	tools := make([]*Tool, 5)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("tool%d", i+1)
		tools[i] = &Tool{
			Name: name,
			Handler: func(args string) (string, error) {
				checkConcurrency()
				return "result", nil
			},
		}
	}

	agent := &Builder{
		tools:          tools,
		enableParallel: true,
		maxWorkers:     2,
		toolTimeout:    30 * time.Second,
	}

	toolCalls := make([]openai.ChatCompletionMessageToolCallUnion, 5)
	for i := 0; i < 5; i++ {
		toolCalls[i] = openai.ChatCompletionMessageToolCallUnion{
			ID:       fmt.Sprintf("call%d", i+1),
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: fmt.Sprintf("tool%d", i+1), Arguments: "{}"},
		}
	}

	_, err := agent.executeToolsParallel(context.Background(), toolCalls)
	if err != nil {
		t.Fatalf("Parallel execution failed: %v", err)
	}

	max := maxConcurrent.Load()
	if max > 2 {
		t.Fatalf("Max concurrent workers was %d, expected ≤ 2", max)
	}

	t.Logf("✓ Worker pool limit enforced: max %d concurrent (limit: 2)", max)
}

// TestToolTimeout tests per-tool timeout enforcement
func TestToolTimeout(t *testing.T) {
	slowTool := &Tool{
		Name: "slow_tool",
		Handler: func(args string) (string, error) {
			time.Sleep(100 * time.Millisecond) // Longer than timeout
			return "should not reach here", nil
		},
	}

	agent := &Builder{
		tools:          []*Tool{slowTool},
		enableParallel: true,
		toolTimeout:    20 * time.Millisecond,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "slow_tool", Arguments: "{}"}},
	}

	_, err := agent.executeToolsParallel(context.Background(), toolCalls)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !contains(err.Error(), "context deadline exceeded") && !contains(err.Error(), "timeout") {
		t.Fatalf("Expected timeout error, got: %v", err)
	}

	t.Logf("✓ Timeout enforced correctly: %v", err)
}

// TestToolErrorHandling tests error aggregation
func TestToolErrorHandling(t *testing.T) {
	goodTool := &Tool{
		Name: "good_tool",
		Handler: func(args string) (string, error) {
			return "success", nil
		},
	}
	badTool := &Tool{
		Name: "bad_tool",
		Handler: func(args string) (string, error) {
			return "", fmt.Errorf("intentional error")
		},
	}

	agent := &Builder{
		tools:          []*Tool{goodTool, badTool},
		enableParallel: true,
		maxWorkers:     10,
		toolTimeout:    30 * time.Second,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "good_tool", Arguments: "{}"}},
		{ID: "call2", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "bad_tool", Arguments: "{}"}},
	}

	_, err := agent.executeToolsParallel(context.Background(), toolCalls)
	if err == nil {
		t.Fatal("Expected error from bad_tool, got nil")
	}

	if !contains(err.Error(), "bad_tool") || !contains(err.Error(), "intentional error") {
		t.Fatalf("Error message missing tool name or error: %v", err)
	}

	t.Logf("✓ Error handled correctly: %v", err)
}

// TestSingleToolExecution tests that single tool works (no parallelism needed)
func TestSingleToolExecution(t *testing.T) {
	var executed atomic.Bool

	tool := &Tool{
		Name: "single_tool",
		Handler: func(args string) (string, error) {
			executed.Store(true)
			return "result", nil
		},
	}

	agent := &Builder{
		tools:          []*Tool{tool},
		enableParallel: true,
		toolTimeout:    30 * time.Second,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "single_tool", Arguments: "{}"}},
	}

	// Single tool should use sequential path automatically
	results, err := agent.executeToolsSequential(context.Background(), toolCalls)
	if err != nil {
		t.Fatalf("Single tool execution failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if !executed.Load() {
		t.Fatal("Tool was not executed")
	}

	t.Logf("✓ Single tool executed successfully")
}

// TestParallelDisabled tests that parallel is off by default
func TestParallelDisabled(t *testing.T) {
	tool := &Tool{
		Name: "test_tool",
		Handler: func(args string) (string, error) {
			return "result", nil
		},
	}

	agent := &Builder{
		tools: []*Tool{tool},
	}

	if agent.enableParallel {
		t.Fatal("Parallel execution should be disabled by default")
	}

	t.Logf("✓ Parallel disabled by default")
}

// TestContextCancellation tests that context cancellation stops execution
func TestContextCancellation(t *testing.T) {
	slowTool := &Tool{
		Name: "slow_tool",
		Handler: func(args string) (string, error) {
			time.Sleep(200 * time.Millisecond)
			return "result", nil
		},
	}

	agent := &Builder{
		tools:          []*Tool{slowTool},
		enableParallel: true,
		maxWorkers:     10,
		toolTimeout:    30 * time.Second,
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{ID: "call1", Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: "slow_tool", Arguments: "{}"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := agent.executeToolsParallel(ctx, toolCalls)
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	// Context timeout manifests as tool timeout error
	if !contains(err.Error(), "timeout") && !contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("Expected timeout/context error, got: %v", err)
	}

	t.Logf("✓ Context cancellation handled: %v", err)
}

// Helper function
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
