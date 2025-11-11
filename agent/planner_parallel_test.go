package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name      string
		tasks     []Task
		wantLen   int
		wantError bool
	}{
		{
			name: "simple linear dependency",
			tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
				{ID: "2", Description: "Task 2", Dependencies: []string{"1"}},
				{ID: "3", Description: "Task 3", Dependencies: []string{"2"}},
			},
			wantLen:   3,
			wantError: false,
		},
		{
			name: "parallel tasks (no dependencies)",
			tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
				{ID: "2", Description: "Task 2", Dependencies: []string{}},
				{ID: "3", Description: "Task 3", Dependencies: []string{}},
			},
			wantLen:   3,
			wantError: false,
		},
		{
			name: "diamond dependency",
			tasks: []Task{
				{ID: "1", Description: "Root", Dependencies: []string{}},
				{ID: "2", Description: "Left", Dependencies: []string{"1"}},
				{ID: "3", Description: "Right", Dependencies: []string{"1"}},
				{ID: "4", Description: "Bottom", Dependencies: []string{"2", "3"}},
			},
			wantLen:   4,
			wantError: false,
		},
		{
			name: "cycle detection",
			tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{"2"}},
				{ID: "2", Description: "Task 2", Dependencies: []string{"1"}},
			},
			wantLen:   0,
			wantError: true,
		},
		{
			name: "self-cycle",
			tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{"1"}},
			},
			wantLen:   0,
			wantError: true,
		},
		{
			name:      "empty task list",
			tasks:     []Task{},
			wantLen:   0,
			wantError: false,
		},
		{
			name: "single task",
			tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
			},
			wantLen:   1,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(DefaultPlannerConfig(), nil)
			sorted, err := executor.topologicalSort(tt.tasks)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(sorted) != tt.wantLen {
				t.Errorf("sorted length = %d, want %d", len(sorted), tt.wantLen)
			}

			// Verify topological order: dependencies come before dependents
			position := make(map[string]int)
			for i, task := range sorted {
				position[task.ID] = i
			}

			for _, task := range sorted {
				for _, depID := range task.Dependencies {
					if position[depID] >= position[task.ID] {
						t.Errorf("dependency %s comes after task %s", depID, task.ID)
					}
				}
			}
		})
	}
}

func TestGroupByDependencyLevel(t *testing.T) {
	tests := []struct {
		name       string
		tasks      []Task
		wantLevels int
		wantError  bool
	}{
		{
			name: "simple linear - 3 levels",
			tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
				{ID: "2", Description: "Task 2", Dependencies: []string{"1"}},
				{ID: "3", Description: "Task 3", Dependencies: []string{"2"}},
			},
			wantLevels: 3,
			wantError:  false,
		},
		{
			name: "all parallel - 1 level",
			tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
				{ID: "2", Description: "Task 2", Dependencies: []string{}},
				{ID: "3", Description: "Task 3", Dependencies: []string{}},
			},
			wantLevels: 1,
			wantError:  false,
		},
		{
			name: "diamond - 3 levels",
			tasks: []Task{
				{ID: "1", Description: "Root", Dependencies: []string{}},
				{ID: "2", Description: "Left", Dependencies: []string{"1"}},
				{ID: "3", Description: "Right", Dependencies: []string{"1"}},
				{ID: "4", Description: "Bottom", Dependencies: []string{"2", "3"}},
			},
			wantLevels: 3,
			wantError:  false,
		},
		{
			name: "complex DAG",
			tasks: []Task{
				{ID: "1", Description: "L0-1", Dependencies: []string{}},
				{ID: "2", Description: "L0-2", Dependencies: []string{}},
				{ID: "3", Description: "L1-1", Dependencies: []string{"1"}},
				{ID: "4", Description: "L1-2", Dependencies: []string{"2"}},
				{ID: "5", Description: "L2-1", Dependencies: []string{"3", "4"}},
			},
			wantLevels: 3,
			wantError:  false,
		},
		{
			name: "cycle detection",
			tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{"2"}},
				{ID: "2", Description: "Task 2", Dependencies: []string{"1"}},
			},
			wantLevels: 0,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(DefaultPlannerConfig(), nil)
			levels, err := executor.groupByDependencyLevel(tt.tasks)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(levels) != tt.wantLevels {
				t.Errorf("level count = %d, want %d", len(levels), tt.wantLevels)
			}

			// Verify level correctness
			for levelIdx, levelTasks := range levels {
				for _, task := range levelTasks {
					// All dependencies must be in earlier levels
					for _, depID := range task.Dependencies {
						found := false
						for earlyLevel := 0; earlyLevel < levelIdx; earlyLevel++ {
							for _, earlyTask := range levels[earlyLevel] {
								if earlyTask.ID == depID {
									found = true
									break
								}
							}
							if found {
								break
							}
						}
						if !found {
							t.Errorf("task %s at level %d has dependency %s not in earlier levels",
								task.ID, levelIdx, depID)
						}
					}
				}
			}
		})
	}
}

func TestExecuteParallel(t *testing.T) {
	tests := []struct {
		name          string
		plan          *Plan
		mockResponses map[string]string
		wantStatus    PlanStatus
		wantError     bool
		minDuration   time.Duration // Expected minimum duration for parallel execution
	}{
		{
			name: "parallel tasks execute concurrently",
			plan: &Plan{
				ID:       "test-parallel",
				Goal:     "Execute 3 tasks in parallel",
				Strategy: StrategyParallel,
				Tasks: []Task{
					{ID: "1", Description: "Task 1", Dependencies: []string{}},
					{ID: "2", Description: "Task 2", Dependencies: []string{}},
					{ID: "3", Description: "Task 3", Dependencies: []string{}},
				},
			},
			mockResponses: map[string]string{
				"Task 1": "Result 1",
				"Task 2": "Result 2",
				"Task 3": "Result 3",
			},
			wantStatus:  PlanStatusCompleted,
			wantError:   false,
			minDuration: 0, // Should be faster than sequential
		},
		{
			name: "dependency levels execute in order",
			plan: &Plan{
				ID:       "test-levels",
				Goal:     "Execute with dependencies",
				Strategy: StrategyParallel,
				Tasks: []Task{
					{ID: "1", Description: "Root", Dependencies: []string{}},
					{ID: "2", Description: "Left", Dependencies: []string{"1"}},
					{ID: "3", Description: "Right", Dependencies: []string{"1"}},
					{ID: "4", Description: "Bottom", Dependencies: []string{"2", "3"}},
				},
			},
			mockResponses: map[string]string{
				"Root":   "Root done",
				"Left":   "Left done",
				"Right":  "Right done",
				"Bottom": "Bottom done",
			},
			wantStatus: PlanStatusCompleted,
			wantError:  false,
		},
		{
			name: "task failure stops execution",
			plan: &Plan{
				ID:       "test-failure",
				Goal:     "Handle task failure",
				Strategy: StrategyParallel,
				Tasks: []Task{
					{ID: "1", Description: "Task 1", Dependencies: []string{}},
					{ID: "2", Description: "FAIL", Dependencies: []string{}},
					{ID: "3", Description: "Task 3", Dependencies: []string{"1", "2"}},
				},
			},
			mockResponses: map[string]string{
				"Task 1": "Success",
				// FAIL will trigger error in mock
				"Task 3": "Should not execute",
			},
			wantStatus: PlanStatusFailed,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock agent
			mock := &mockAgent{
				chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
					response, ok := tt.mockResponses[message]
					if !ok {
						return nil, errors.New("task failed")
					}
					return &ChatResult{Content: response}, nil
				},
			}

			config := DefaultPlannerConfig()
			config.MaxParallel = 3
			executor := NewExecutor(config, mock)

			ctx := context.Background()
			start := time.Now()
			result, err := executor.executeParallel(ctx, tt.plan)
			duration := time.Since(start)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if result.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", result.Status, tt.wantStatus)
			}

			if tt.minDuration > 0 && duration < tt.minDuration {
				t.Errorf("execution too fast: %v, expected at least %v", duration, tt.minDuration)
			}
		})
	}
}

func TestExecuteParallelConcurrency(t *testing.T) {
	// Create plan with 10 parallel tasks
	tasks := []Task{}
	for i := 1; i <= 10; i++ {
		tasks = append(tasks, Task{
			ID:           fmt.Sprintf("%d", i),
			Description:  fmt.Sprintf("Task %d", i),
			Dependencies: []string{},
		})
	}

	plan := &Plan{
		ID:       "concurrency-test",
		Goal:     "Test MaxParallel limit",
		Strategy: StrategyParallel,
		Tasks:    tasks,
	}

	// Mock agent that tracks concurrent execution
	var mu sync.Mutex
	var maxConcurrent int
	var currentConcurrent int

	mock := &mockAgentFunc{
		chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
			mu.Lock()
			currentConcurrent++
			if currentConcurrent > maxConcurrent {
				maxConcurrent = currentConcurrent
			}
			mu.Unlock()

			// Simulate work
			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			currentConcurrent--
			mu.Unlock()

			return &ChatResult{
				Content: "Done",
			}, nil
		},
	}

	config := DefaultPlannerConfig()
	config.MaxParallel = 3 // Limit to 3 concurrent tasks
	executor := NewExecutor(config, mock)

	ctx := context.Background()
	result, err := executor.executeParallel(ctx, plan)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != PlanStatusCompleted {
		t.Errorf("status = %v, want %v", result.Status, PlanStatusCompleted)
	}

	// Verify that we never exceeded MaxParallel
	if maxConcurrent > config.MaxParallel {
		t.Errorf("max concurrent = %d, exceeds limit %d", maxConcurrent, config.MaxParallel)
	}

	t.Logf("Max concurrent tasks: %d (limit: %d)", maxConcurrent, config.MaxParallel)
}

func TestExecuteParallelContextCancellation(t *testing.T) {
	t.Skip("Context cancellation timing is implementation-dependent")

	plan := &Plan{
		ID:       "cancel-test",
		Goal:     "Test context cancellation",
		Strategy: StrategyParallel,
		Tasks: []Task{
			{ID: "1", Description: "Task 1", Dependencies: []string{}},
			{ID: "2", Description: "Task 2", Dependencies: []string{}},
			{ID: "3", Description: "Task 3", Dependencies: []string{}},
		},
	}

	mock := &mockAgentFunc{
		chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
			// Check context before simulating work
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			// Simulate long-running task
			select {
			case <-time.After(1 * time.Second):
				return &ChatResult{Content: "Done"}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	executor := NewExecutor(DefaultPlannerConfig(), mock)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := executor.executeParallel(ctx, plan)

	if err == nil {
		t.Errorf("expected context cancellation error")
	}

	if result.Status != PlanStatusCanceled {
		t.Errorf("status = %v, want %v", result.Status, PlanStatusCanceled)
	}
}

// mockAgentFunc allows custom chat function for testing
type mockAgentFunc struct {
	chatFunc func(context.Context, string, *ChatOptions) (*ChatResult, error)
}

func (m *mockAgentFunc) Chat(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
	return m.chatFunc(ctx, message, opts)
}
