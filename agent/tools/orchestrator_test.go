package tools

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestOrchestrator_Sequential tests sequential execution
func TestOrchestrator_Sequential(t *testing.T) {
	orch := NewOrchestrator()
	orch.SetParallelExecution(false)

	ctx := context.Background()
	calls := []*ToolCall{
		{
			ID:   "1",
			Name: "tool1",
			Args: `{"value":1}`,
			Handler: func(args string) (string, error) {
				time.Sleep(10 * time.Millisecond)
				return "result1", nil
			},
		},
		{
			ID:   "2",
			Name: "tool2",
			Args: `{"value":2}`,
			Handler: func(args string) (string, error) {
				time.Sleep(10 * time.Millisecond)
				return "result2", nil
			},
		},
	}

	start := time.Now()
	results, err := orch.Execute(ctx, calls)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Sequential should take at least 20ms (2 tools * 10ms)
	if duration < 20*time.Millisecond {
		t.Errorf("Sequential execution too fast: %v (expected >= 20ms)", duration)
	}

	// Check results order
	if results[0].ID != "1" || results[1].ID != "2" {
		t.Error("Results not in correct order")
	}
}

// TestOrchestrator_Parallel tests parallel execution
func TestOrchestrator_Parallel(t *testing.T) {
	orch := NewOrchestrator()
	orch.SetParallelExecution(true)
	orch.SetMaxWorkers(5)

	ctx := context.Background()
	calls := []*ToolCall{
		{
			ID:   "1",
			Name: "tool1",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(50 * time.Millisecond)
				return "result1", nil
			},
		},
		{
			ID:   "2",
			Name: "tool2",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(50 * time.Millisecond)
				return "result2", nil
			},
		},
		{
			ID:   "3",
			Name: "tool3",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(50 * time.Millisecond)
				return "result3", nil
			},
		},
	}

	start := time.Now()
	results, err := orch.Execute(ctx, calls)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Parallel should take ~50ms, not 150ms (3 * 50ms)
	if duration > 100*time.Millisecond {
		t.Errorf("Parallel execution too slow: %v (expected ~50ms)", duration)
	}

	t.Logf("Parallel execution: 3 tools in %v (expected ~50ms)", duration)
}

// TestOrchestrator_Dependencies tests dependency-based execution
func TestOrchestrator_Dependencies(t *testing.T) {
	orch := NewOrchestrator()
	orch.SetParallelExecution(true)

	var executionOrder []string
	var mu sync.Mutex

	ctx := context.Background()
	calls := []*ToolCall{
		{
			ID:   "1",
			Name: "tool1",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				mu.Lock()
				executionOrder = append(executionOrder, "tool1")
				mu.Unlock()
				time.Sleep(20 * time.Millisecond)
				return "result1", nil
			},
		},
		{
			ID:        "2",
			Name:      "tool2",
			Args:      `{}`,
			DependsOn: []string{"1"}, // Depends on tool1
			Handler: func(args string) (string, error) {
				mu.Lock()
				executionOrder = append(executionOrder, "tool2")
				mu.Unlock()
				time.Sleep(20 * time.Millisecond)
				return "result2", nil
			},
		},
		{
			ID:        "3",
			Name:      "tool3",
			Args:      `{}`,
			DependsOn: []string{"1"}, // Also depends on tool1
			Handler: func(args string) (string, error) {
				mu.Lock()
				executionOrder = append(executionOrder, "tool3")
				mu.Unlock()
				time.Sleep(20 * time.Millisecond)
				return "result3", nil
			},
		},
	}

	results, err := orch.Execute(ctx, calls)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// tool1 should execute first
	if executionOrder[0] != "tool1" {
		t.Errorf("tool1 should execute first, got %v", executionOrder)
	}

	// tool2 and tool3 should execute after tool1 (in parallel)
	// Order between tool2 and tool3 doesn't matter
	if len(executionOrder) != 3 {
		t.Errorf("Expected 3 executions, got %d", len(executionOrder))
	}

	t.Logf("Execution order: %v", executionOrder)
}

// TestOrchestrator_Timeout tests per-tool timeout
func TestOrchestrator_Timeout(t *testing.T) {
	orch := NewOrchestrator()
	orch.SetToolTimeout(50 * time.Millisecond)

	ctx := context.Background()
	calls := []*ToolCall{
		{
			ID:   "1",
			Name: "slow_tool",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(200 * time.Millisecond) // Longer than timeout
				return "should_not_reach", nil
			},
		},
	}

	results, err := orch.Execute(ctx, calls)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Should timeout
	if results[0].Error == nil {
		t.Error("Expected timeout error, got nil")
	}

	if results[0].Error != nil {
		t.Logf("Got expected timeout error: %v", results[0].Error)
	}
}

// TestOrchestrator_ContextCancellation tests context cancellation
func TestOrchestrator_ContextCancellation(t *testing.T) {
	orch := NewOrchestrator()

	ctx, cancel := context.WithCancel(context.Background())

	calls := []*ToolCall{
		{
			ID:   "1",
			Name: "tool1",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(100 * time.Millisecond)
				return "result1", nil
			},
		},
		{
			ID:   "2",
			Name: "tool2",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(100 * time.Millisecond)
				return "result2", nil
			},
		},
	}

	// Cancel after 30ms
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	results, _ := orch.Execute(ctx, calls)

	// Should have results (may be incomplete)
	if len(results) == 0 {
		t.Error("Expected some results")
	}

	t.Logf("Got %d results after cancellation", len(results))
}

// TestOrchestrator_ErrorHandling tests error aggregation
func TestOrchestrator_ErrorHandling(t *testing.T) {
	orch := NewOrchestrator()

	ctx := context.Background()
	calls := []*ToolCall{
		{
			ID:   "1",
			Name: "success_tool",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				return "success", nil
			},
		},
		{
			ID:   "2",
			Name: "error_tool",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				return "", fmt.Errorf("intentional error")
			},
		},
		{
			ID:   "3",
			Name: "another_success",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				return "success", nil
			},
		},
	}

	results, err := orch.Execute(ctx, calls)
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Check individual results
	if results[0].Error != nil {
		t.Errorf("Tool 1 should succeed, got error: %v", results[0].Error)
	}
	if results[1].Error == nil {
		t.Error("Tool 2 should fail, got success")
	}
	if results[2].Error != nil {
		t.Errorf("Tool 3 should succeed, got error: %v", results[2].Error)
	}
}

// TestOrchestrator_WorkerPool tests worker pool limits
func TestOrchestrator_WorkerPool(t *testing.T) {
	orch := NewOrchestrator()
	orch.SetMaxWorkers(2) // Only 2 concurrent workers

	var concurrentCount int32
	var maxConcurrent int32

	ctx := context.Background()
	calls := make([]*ToolCall, 5)

	for i := 0; i < 5; i++ {
		calls[i] = &ToolCall{
			ID:   fmt.Sprintf("%d", i+1),
			Name: fmt.Sprintf("tool%d", i+1),
			Args: `{}`,
			Handler: func(args string) (string, error) {
				current := atomic.AddInt32(&concurrentCount, 1)

				// Track maximum concurrent executions
				for {
					max := atomic.LoadInt32(&maxConcurrent)
					if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
						break
					}
				}

				time.Sleep(50 * time.Millisecond)
				atomic.AddInt32(&concurrentCount, -1)
				return "success", nil
			},
		}
	}

	results, err := orch.Execute(ctx, calls)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(results) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(results))
	}

	max := atomic.LoadInt32(&maxConcurrent)
	if max > 2 {
		t.Errorf("Max concurrent workers should be 2, got %d", max)
	}

	t.Logf("Max concurrent executions: %d (expected <= 2)", max)
}

// TestOrchestrator_ResultOrder tests that results maintain original order
func TestOrchestrator_ResultOrder(t *testing.T) {
	orch := NewOrchestrator()
	orch.SetParallelExecution(true)

	ctx := context.Background()
	calls := make([]*ToolCall, 10)

	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("tool%d", i)
		calls[i] = &ToolCall{
			ID:   id,
			Name: id,
			Args: `{}`,
			Handler: func(args string) (string, error) {
				// Random delays
				time.Sleep(time.Duration(10+i*5) * time.Millisecond)
				return id, nil
			},
		}
	}

	results, err := orch.Execute(ctx, calls)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(results) != 10 {
		t.Fatalf("Expected 10 results, got %d", len(results))
	}

	// Check order matches input
	for i, result := range results {
		expected := fmt.Sprintf("tool%d", i)
		if result.ID != expected {
			t.Errorf("Result %d: expected ID %s, got %s", i, expected, result.ID)
		}
	}
}

// TestOrchestrator_ComputeStats tests statistics calculation
func TestOrchestrator_ComputeStats(t *testing.T) {
	orch := NewOrchestrator()

	ctx := context.Background()
	calls := []*ToolCall{
		{
			ID:   "1",
			Name: "tool1",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(10 * time.Millisecond)
				return "success", nil
			},
		},
		{
			ID:   "2",
			Name: "tool2",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(20 * time.Millisecond)
				return "", fmt.Errorf("error")
			},
		},
		{
			ID:   "3",
			Name: "tool3",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				time.Sleep(15 * time.Millisecond)
				return "success", nil
			},
		},
	}

	results, err := orch.Execute(ctx, calls)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	plan := orch.buildExecutionPlan(calls)
	stats := ComputeStats(results, plan)

	if stats.TotalTools != 3 {
		t.Errorf("Expected 3 total tools, got %d", stats.TotalTools)
	}

	if stats.SuccessCount != 2 {
		t.Errorf("Expected 2 successes, got %d", stats.SuccessCount)
	}

	if stats.FailureCount != 1 {
		t.Errorf("Expected 1 failure, got %d", stats.FailureCount)
	}

	if stats.MaxDuration < 20*time.Millisecond {
		t.Errorf("Max duration should be >= 20ms, got %v", stats.MaxDuration)
	}

	t.Logf("Stats: %+v", stats)
}

// TestOrchestrator_PanicRecovery tests that panics in tools are handled
func TestOrchestrator_PanicRecovery(t *testing.T) {
	orch := NewOrchestrator()

	ctx := context.Background()
	calls := []*ToolCall{
		{
			ID:   "1",
			Name: "panic_tool",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				panic("intentional panic")
			},
		},
		{
			ID:   "2",
			Name: "normal_tool",
			Args: `{}`,
			Handler: func(args string) (string, error) {
				return "success", nil
			},
		},
	}

	results, err := orch.Execute(ctx, calls)
	if err != nil {
		t.Fatalf("Execute should not return error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// First tool should have panic error
	if results[0].Error == nil {
		t.Error("Expected panic error for tool 1")
	}

	// Second tool should succeed
	if results[1].Error != nil {
		t.Errorf("Tool 2 should succeed, got error: %v", results[1].Error)
	}

	t.Logf("Panic handled: %v", results[0].Error)
}

// TestOrchestrator_EmptyTools tests empty input
func TestOrchestrator_EmptyTools(t *testing.T) {
	orch := NewOrchestrator()

	ctx := context.Background()
	results, err := orch.Execute(ctx, []*ToolCall{})

	if err != nil {
		t.Errorf("Expected no error for empty tools, got: %v", err)
	}

	if results != nil {
		t.Errorf("Expected nil results for empty tools, got: %v", results)
	}
}

// TestOrchestrator_CustomTimeout tests per-tool custom timeout
func TestOrchestrator_CustomTimeout(t *testing.T) {
	orch := NewOrchestrator()
	orch.SetToolTimeout(1 * time.Second) // Default 1s

	ctx := context.Background()
	calls := []*ToolCall{
		{
			ID:      "1",
			Name:    "custom_timeout",
			Args:    `{}`,
			Timeout: 50 * time.Millisecond, // Custom timeout (shorter)
			Handler: func(args string) (string, error) {
				time.Sleep(100 * time.Millisecond)
				return "should_timeout", nil
			},
		},
		{
			ID:   "2",
			Name: "default_timeout",
			Args: `{}`,
			// Uses default 1s timeout
			Handler: func(args string) (string, error) {
				time.Sleep(100 * time.Millisecond)
				return "success", nil
			},
		},
	}

	results, err := orch.Execute(ctx, calls)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Tool 1 should timeout
	if results[0].Error == nil {
		t.Error("Tool 1 should timeout with custom timeout")
	}

	// Tool 2 should succeed with default timeout
	if results[1].Error != nil {
		t.Errorf("Tool 2 should succeed, got error: %v", results[1].Error)
	}
}
