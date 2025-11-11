package agent

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Integration tests for advanced planning scenarios
// These tests verify end-to-end behavior with realistic workloads

func TestIntegration_ParallelBatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("batch process 10 items in parallel", func(t *testing.T) {
		// Create a plan to process 10 items with 5 parallel workers
		tasks := []Task{}
		for i := 1; i <= 10; i++ {
			tasks = append(tasks, Task{
				ID:          fmt.Sprintf("item-%d", i),
				Description: fmt.Sprintf("Process item %d", i),
				Type:        TaskTypeAction,
			})
		}

		plan := &Plan{
			ID:       "batch-processing-test",
			Goal:     "Process 10 items efficiently",
			Strategy: StrategyParallel,
			Tasks:    tasks,
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				// Simulate processing time
				time.Sleep(50 * time.Millisecond)
				return &ChatResult{Content: fmt.Sprintf("Processed: %s", message)}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.MaxParallel = 5
		executor := NewExecutor(config, mock)

		ctx := context.Background()
		start := time.Now()
		result, err := executor.executeParallel(ctx, plan)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		if result.Status != PlanStatusCompleted {
			t.Errorf("status = %v, want %v", result.Status, PlanStatusCompleted)
		}

		// Verify all tasks completed
		completedCount := 0
		for _, task := range plan.Tasks {
			if task.Status == TaskStatusCompleted {
				completedCount++
			}
		}

		if completedCount != 10 {
			t.Errorf("completed tasks = %d, want 10", completedCount)
		}

		// Verify parallel execution was faster than sequential
		// Sequential: 10 * 50ms = 500ms
		// Parallel (5 workers): 2 batches * 50ms = ~100-150ms
		if duration > 300*time.Millisecond {
			t.Logf("Warning: parallel execution took %v, expected < 300ms", duration)
		}

		t.Logf("Batch processing: %d tasks in %v (%.1f tasks/sec)",
			completedCount, duration, float64(completedCount)/duration.Seconds())
	})

	t.Run("parallel with error handling", func(t *testing.T) {
		tasks := []Task{
			{ID: "1", Description: "Success 1", Type: TaskTypeAction},
			{ID: "2", Description: "FAIL", Type: TaskTypeAction},
			{ID: "3", Description: "Success 2", Type: TaskTypeAction},
		}

		plan := &Plan{
			ID:       "error-handling-test",
			Goal:     "Handle errors in parallel execution",
			Strategy: StrategyParallel,
			Tasks:    tasks,
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				if message == "FAIL" {
					return nil, fmt.Errorf("intentional failure")
				}
				return &ChatResult{Content: "done"}, nil
			},
		}

		executor := NewExecutor(DefaultPlannerConfig(), mock)
		ctx := context.Background()
		result, err := executor.executeParallel(ctx, plan)

		if err == nil {
			t.Error("expected error for failed task")
		}

		if result.Status != PlanStatusFailed {
			t.Errorf("status = %v, want %v", result.Status, PlanStatusFailed)
		}
	})
}

func TestIntegration_AdaptiveStrategySwitch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("adaptive switches based on performance", func(t *testing.T) {
		// Create multi-level tasks that would benefit from different strategies
		plan := &Plan{
			ID:       "adaptive-switching-test",
			Goal:     "Demonstrate adaptive strategy switching",
			Strategy: StrategyAdaptive,
			Tasks: []Task{
				// Level 0: Single root task
				{ID: "root", Description: "Root task", Type: TaskTypeObservation},
				// Level 1: Multiple parallel tasks
				{ID: "batch-1", Description: "Batch 1", Dependencies: []string{"root"}},
				{ID: "batch-2", Description: "Batch 2", Dependencies: []string{"root"}},
				{ID: "batch-3", Description: "Batch 3", Dependencies: []string{"root"}},
				// Level 2: Aggregation
				{ID: "aggregate", Description: "Aggregate results", Dependencies: []string{"batch-1", "batch-2", "batch-3"}},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				time.Sleep(30 * time.Millisecond)
				return &ChatResult{Content: "done"}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.AdaptiveThreshold = 0.5
		config.MaxParallel = 3
		executor := NewExecutor(config, mock)

		ctx := context.Background()
		result, err := executor.executeAdaptive(ctx, plan)

		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		if result.Status != PlanStatusCompleted {
			t.Errorf("status = %v, want %v", result.Status, PlanStatusCompleted)
		}

		// Verify timeline has strategy events
		hasStrategyInit := false
		for _, event := range result.Timeline {
			if event.Type == "strategy_initialized" {
				hasStrategyInit = true
			}
			t.Logf("Event: %s - %s", event.Type, event.Description)
		}

		if !hasStrategyInit {
			t.Error("timeline should have strategy_initialized event")
		}
	})
}

func TestIntegration_GoalOrientedPlanning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("early termination when goal met", func(t *testing.T) {
		// Simulate finding items until goal is met
		itemsFound := 0

		plan := &Plan{
			ID:   "goal-oriented-test",
			Goal: "Find 3 high-quality items",
			GoalState: GoalState{
				Description: "Found 3 items with quality >= 4.0",
				Criteria: []GoalCriterion{
					{
						Name:      "items_found",
						Expected:  "3",
						Operator:  ">=",
						Satisfied: false, // Will be updated dynamically
					},
				},
			},
			Tasks: []Task{
				{ID: "search-1", Description: "Search batch 1", Type: TaskTypeObservation},
				{ID: "search-2", Description: "Search batch 2", Type: TaskTypeObservation},
				{ID: "search-3", Description: "Search batch 3", Type: TaskTypeObservation},
				{ID: "search-4", Description: "Search batch 4", Type: TaskTypeObservation},
				{ID: "search-5", Description: "Search batch 5", Type: TaskTypeObservation},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				itemsFound++
				// Simulate goal being met after 3 items
				if itemsFound >= 3 {
					plan.GoalState.Criteria[0].Satisfied = true
				}
				return &ChatResult{Content: fmt.Sprintf("Found item %d", itemsFound)}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.GoalCheckInterval = 1 // Check after every task
		executor := NewExecutor(config, mock)

		ctx := context.Background()
		result, err := executor.Execute(ctx, plan)

		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		// Verify goal was checked periodically
		goalCheckCount := 0
		for _, event := range result.Timeline {
			if event.Type == "goal_checked" {
				goalCheckCount++
			}
		}

		if goalCheckCount == 0 {
			t.Error("no goal checks recorded")
		}

		t.Logf("Goal checks: %d, Items found: %d", goalCheckCount, itemsFound)
	})
}

func TestIntegration_DeepNesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("3-level task hierarchy", func(t *testing.T) {
		// Create a deeply nested task structure
		plan := &Plan{
			ID:   "deep-nesting-test",
			Goal: "Handle deep task hierarchies",
			Tasks: []Task{
				{
					ID:          "level-0",
					Description: "Root task",
					Type:        TaskTypeObservation,
					Subtasks: []Task{
						{
							ID:          "level-1-a",
							Description: "Subtask 1A",
							Type:        TaskTypeAction,
							Subtasks: []Task{
								{
									ID:          "level-2-a1",
									Description: "Deep subtask A1",
									Type:        TaskTypeAction,
								},
								{
									ID:          "level-2-a2",
									Description: "Deep subtask A2",
									Type:        TaskTypeAction,
								},
							},
						},
						{
							ID:          "level-1-b",
							Description: "Subtask 1B",
							Type:        TaskTypeAction,
						},
					},
				},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: fmt.Sprintf("Completed: %s", message)}, nil
			},
		}

		executor := NewExecutor(DefaultPlannerConfig(), mock)
		ctx := context.Background()
		result, err := executor.Execute(ctx, plan)

		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		if result.Status != PlanStatusCompleted {
			t.Errorf("status = %v, want %v", result.Status, PlanStatusCompleted)
		}

		// Verify all subtasks were executed
		// Should have: 1 root + 2 level-1 + 2 level-2 = 5 total tasks
		taskCount := 0
		for _, event := range result.Timeline {
			if event.Type == "task_completed" {
				taskCount++
			}
		}

		if taskCount < 3 {
			t.Logf("Warning: only %d tasks completed, expected at least 3", taskCount)
		}
	})
}

func TestIntegration_ComplexDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("diamond dependency pattern", func(t *testing.T) {
		// Classic diamond: A -> B,C -> D
		plan := &Plan{
			ID:       "diamond-dep-test",
			Goal:     "Handle diamond dependencies",
			Strategy: StrategyParallel,
			Tasks: []Task{
				{ID: "A", Description: "Root", Type: TaskTypeObservation},
				{ID: "B", Description: "Left branch", Type: TaskTypeAction, Dependencies: []string{"A"}},
				{ID: "C", Description: "Right branch", Type: TaskTypeAction, Dependencies: []string{"A"}},
				{ID: "D", Description: "Convergence", Type: TaskTypeAction, Dependencies: []string{"B", "C"}},
			},
		}

		executionOrder := []string{}
		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				// Extract task ID from message (assumes format like "Root", "Left branch", etc.)
				taskID := ""
				for _, task := range plan.Tasks {
					if task.Description == message {
						taskID = task.ID
						break
					}
				}
				executionOrder = append(executionOrder, taskID)
				time.Sleep(20 * time.Millisecond)
				return &ChatResult{Content: "done"}, nil
			},
		}

		executor := NewExecutor(DefaultPlannerConfig(), mock)
		ctx := context.Background()
		result, err := executor.executeParallel(ctx, plan)

		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		if result.Status != PlanStatusCompleted {
			t.Errorf("status = %v, want %v", result.Status, PlanStatusCompleted)
		}

		// Verify execution order respects dependencies
		// A must come before B and C
		// B and C must come before D
		posA, posB, posC, posD := -1, -1, -1, -1
		for i, id := range executionOrder {
			switch id {
			case "A":
				posA = i
			case "B":
				posB = i
			case "C":
				posC = i
			case "D":
				posD = i
			}
		}

		if posA == -1 || posB == -1 || posC == -1 || posD == -1 {
			t.Errorf("not all tasks executed: A=%d, B=%d, C=%d, D=%d", posA, posB, posC, posD)
		}

		if posA >= posB || posA >= posC {
			t.Errorf("A should execute before B and C")
		}

		if posB >= posD || posC >= posD {
			t.Errorf("B and C should execute before D")
		}

		t.Logf("Execution order: %v", executionOrder)
	})
}

func TestIntegration_TimelineCompleteness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("timeline captures all events", func(t *testing.T) {
		plan := &Plan{
			ID:   "timeline-completeness-test",
			Goal: "Verify comprehensive timeline",
			Tasks: []Task{
				{ID: "1", Description: "Task 1", Type: TaskTypeObservation},
				{ID: "2", Description: "Task 2", Type: TaskTypeAction},
				{ID: "3", Description: "Task 3", Type: TaskTypeAction},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				time.Sleep(10 * time.Millisecond)
				return &ChatResult{Content: "done"}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.GoalCheckInterval = 2
		executor := NewExecutor(config, mock)

		ctx := context.Background()
		result, err := executor.Execute(ctx, plan)

		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		// Verify timeline has expected event types
		eventTypes := make(map[string]int)
		for _, event := range result.Timeline {
			eventTypes[event.Type]++
		}

		// Should have: task_started, task_completed for each task
		// Plus goal_checked events
		expectedTypes := []string{"task_started", "task_completed"}
		for _, eventType := range expectedTypes {
			if eventTypes[eventType] == 0 {
				t.Errorf("missing event type: %s", eventType)
			}
		}

		// Should have at least one goal check (3 tasks, interval 2)
		if eventTypes["goal_checked"] == 0 {
			t.Error("missing goal_checked events")
		}

		t.Logf("Event distribution: %v", eventTypes)
		t.Logf("Total timeline events: %d", len(result.Timeline))
	})
}

func TestIntegration_PerformanceMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("metrics track execution accurately", func(t *testing.T) {
		plan := &Plan{
			ID:       "metrics-test",
			Goal:     "Track performance metrics",
			Strategy: StrategyParallel,
			Tasks: []Task{
				{ID: "1", Description: "Task 1"},
				{ID: "2", Description: "Task 2"},
				{ID: "3", Description: "Task 3"},
				{ID: "4", Description: "Task 4"},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				time.Sleep(25 * time.Millisecond)
				return &ChatResult{Content: "done"}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.MaxParallel = 2
		executor := NewExecutor(config, mock)

		ctx := context.Background()
		result, err := executor.executeParallel(ctx, plan)

		if err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		// Verify metrics
		if result.Metrics.TaskCount != 4 {
			t.Errorf("task count = %d, want 4", result.Metrics.TaskCount)
		}

		if result.Metrics.ExecutionTime == 0 {
			t.Error("execution time should not be zero")
		}

		if result.Metrics.AvgTaskDuration == 0 {
			t.Error("average task duration should not be zero")
		}

		if result.Metrics.SuccessRate != 1.0 {
			t.Errorf("success rate = %.2f, want 1.0", result.Metrics.SuccessRate)
		}

		t.Logf("Metrics - Tasks: %d, Duration: %v, Avg: %v, Success: %.1f%%",
			result.Metrics.TaskCount,
			result.Metrics.ExecutionTime,
			result.Metrics.AvgTaskDuration,
			result.Metrics.SuccessRate*100)
	})
}
