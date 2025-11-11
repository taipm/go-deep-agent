package agent

import (
	"context"
	"testing"
	"time"
)

func TestPerformanceTracker(t *testing.T) {
	t.Run("basic metrics tracking", func(t *testing.T) {
		tracker := &performanceTracker{}
		tracker.startBatch(StrategySequential, 5)

		// Simulate task completions
		tracker.recordTaskCompletion(100 * time.Millisecond)
		tracker.recordTaskCompletion(150 * time.Millisecond)
		tracker.recordTaskCompletion(120 * time.Millisecond)

		metrics := tracker.getMetrics()

		if metrics.CompletedTasks != 3 {
			t.Errorf("completed tasks = %d, want 3", metrics.CompletedTasks)
		}

		if metrics.TotalTasks != 5 {
			t.Errorf("total tasks = %d, want 5", metrics.TotalTasks)
		}

		if metrics.Strategy != StrategySequential {
			t.Errorf("strategy = %v, want %v", metrics.Strategy, StrategySequential)
		}

		if metrics.AvgLatency == 0 {
			t.Errorf("average latency should not be zero")
		}
	})

	t.Run("parallel efficiency calculation", func(t *testing.T) {
		tracker := &performanceTracker{}
		tracker.startBatch(StrategyParallel, 4)

		// Simulate 4 tasks taking 100ms each, completed in parallel
		for i := 0; i < 4; i++ {
			tracker.recordTaskCompletion(100 * time.Millisecond)
		}

		// Wait a bit to simulate actual wall time
		time.Sleep(150 * time.Millisecond)

		metrics := tracker.getMetrics()

		if metrics.ParallelEfficiency <= 0 {
			t.Errorf("parallel efficiency should be positive, got %f", metrics.ParallelEfficiency)
		}
	})

	t.Run("tasks per second", func(t *testing.T) {
		tracker := &performanceTracker{}
		tracker.startBatch(StrategySequential, 10)

		// Complete 5 tasks
		for i := 0; i < 5; i++ {
			tracker.recordTaskCompletion(50 * time.Millisecond)
		}

		time.Sleep(100 * time.Millisecond) // Ensure some time passes

		metrics := tracker.getMetrics()

		if metrics.TasksPerSec <= 0 {
			t.Errorf("tasks per second should be positive, got %f", metrics.TasksPerSec)
		}
	})
}

func TestShouldSwitchStrategy(t *testing.T) {
	tests := []struct {
		name            string
		metrics         PerformanceMetrics
		currentStrategy PlanningStrategy
		threshold       float64
		wantSwitch      bool
	}{
		{
			name: "parallel with low efficiency - should switch",
			metrics: PerformanceMetrics{
				ParallelEfficiency: 0.2,
				Strategy:           StrategyParallel,
			},
			currentStrategy: StrategyParallel,
			threshold:       0.3,
			wantSwitch:      true,
		},
		{
			name: "parallel with good efficiency - no switch",
			metrics: PerformanceMetrics{
				ParallelEfficiency: 0.8,
				Strategy:           StrategyParallel,
			},
			currentStrategy: StrategyParallel,
			threshold:       0.3,
			wantSwitch:      false,
		},
		{
			name: "sequential - no switch by default",
			metrics: PerformanceMetrics{
				TasksPerSec: 2.0,
				Strategy:    StrategySequential,
			},
			currentStrategy: StrategySequential,
			threshold:       0.3,
			wantSwitch:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultPlannerConfig()
			config.AdaptiveThreshold = tt.threshold
			executor := NewExecutor(config, nil)

			gotSwitch := executor.shouldSwitchStrategy(tt.metrics, tt.currentStrategy)

			if gotSwitch != tt.wantSwitch {
				t.Errorf("shouldSwitchStrategy() = %v, want %v", gotSwitch, tt.wantSwitch)
			}
		})
	}
}

func TestSelectNextStrategy(t *testing.T) {
	tests := []struct {
		name            string
		metrics         PerformanceMetrics
		currentStrategy PlanningStrategy
		remainingTasks  int
		threshold       float64
		wantStrategy    PlanningStrategy
	}{
		{
			name: "parallel with low efficiency - switch to sequential",
			metrics: PerformanceMetrics{
				ParallelEfficiency: 0.15,
			},
			currentStrategy: StrategyParallel,
			remainingTasks:  10,
			threshold:       0.3,
			wantStrategy:    StrategySequential,
		},
		{
			name: "parallel with good efficiency - stay parallel",
			metrics: PerformanceMetrics{
				ParallelEfficiency: 0.85,
			},
			currentStrategy: StrategyParallel,
			remainingTasks:  10,
			threshold:       0.3,
			wantStrategy:    StrategyParallel,
		},
		{
			name: "sequential with many tasks - stay sequential",
			metrics: PerformanceMetrics{
				TasksPerSec: 3.0,
			},
			currentStrategy: StrategySequential,
			remainingTasks:  20,
			threshold:       0.3,
			wantStrategy:    StrategySequential,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultPlannerConfig()
			config.AdaptiveThreshold = tt.threshold
			executor := NewExecutor(config, nil)

			gotStrategy := executor.selectNextStrategy(tt.metrics, tt.currentStrategy, tt.remainingTasks)

			if gotStrategy != tt.wantStrategy {
				t.Errorf("selectNextStrategy() = %v, want %v", gotStrategy, tt.wantStrategy)
			}
		})
	}
}

func TestExecuteAdaptive(t *testing.T) {
	tests := []struct {
		name          string
		plan          *Plan
		mockResponses map[string]string
		wantStatus    PlanStatus
		wantEvents    int // Minimum number of timeline events
	}{
		{
			name: "simple adaptive execution",
			plan: &Plan{
				ID:       "adaptive-test",
				Goal:     "Test adaptive strategy",
				Strategy: StrategyAdaptive,
				Tasks: []Task{
					{ID: "1", Description: "Task 1", Dependencies: []string{}},
					{ID: "2", Description: "Task 2", Dependencies: []string{}},
					{ID: "3", Description: "Task 3", Dependencies: []string{"1", "2"}},
				},
			},
			mockResponses: map[string]string{
				"Task 1": "Result 1",
				"Task 2": "Result 2",
				"Task 3": "Result 3",
			},
			wantStatus: PlanStatusCompleted,
			wantEvents: 1, // At least initial strategy event
		},
		{
			name: "adaptive with multiple levels",
			plan: &Plan{
				ID:       "adaptive-levels",
				Goal:     "Multi-level adaptive",
				Strategy: StrategyAdaptive,
				Tasks: []Task{
					{ID: "1", Description: "Root", Dependencies: []string{}},
					{ID: "2", Description: "L1-A", Dependencies: []string{"1"}},
					{ID: "3", Description: "L1-B", Dependencies: []string{"1"}},
					{ID: "4", Description: "L2", Dependencies: []string{"2", "3"}},
				},
			},
			mockResponses: map[string]string{
				"Root":  "Root done",
				"L1-A":  "L1-A done",
				"L1-B":  "L1-B done",
				"L2":    "L2 done",
			},
			wantStatus: PlanStatusCompleted,
			wantEvents: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock agent
			mock := &mockAgent{
				chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
					response, ok := tt.mockResponses[message]
					if !ok {
						return &ChatResult{Content: "default response"}, nil
					}
					return &ChatResult{Content: response}, nil
				},
			}

			config := DefaultPlannerConfig()
			config.AdaptiveThreshold = 0.3
			executor := NewExecutor(config, mock)

			ctx := context.Background()
			result, err := executor.executeAdaptive(ctx, tt.plan)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", result.Status, tt.wantStatus)
			}

			if len(result.Timeline) < tt.wantEvents {
				t.Errorf("timeline events = %d, want at least %d", len(result.Timeline), tt.wantEvents)
			}

			// Verify timeline has strategy initialized event
			foundInit := false
			for _, event := range result.Timeline {
				if event.Type == "strategy_initialized" {
					foundInit = true
					break
				}
			}

			if !foundInit {
				t.Errorf("timeline should contain strategy_initialized event")
			}
		})
	}
}

func TestExecuteBatchSequential(t *testing.T) {
	plan := &Plan{
		ID:   "batch-test",
		Goal: "Test batch execution",
		Tasks: []Task{
			{ID: "1", Description: "Task 1", Dependencies: []string{}},
			{ID: "2", Description: "Task 2", Dependencies: []string{}},
			{ID: "3", Description: "Task 3", Dependencies: []string{}},
		},
	}

	mock := &mockAgent{
		chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
			return &ChatResult{Content: "done"}, nil
		},
	}

	executor := NewExecutor(DefaultPlannerConfig(), mock)
	execCtx := newExecutionContext(plan)

	ctx := context.Background()
	err := executor.executeBatchSequential(ctx, plan.Tasks, plan, execCtx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify all tasks completed
	if len(execCtx.results) != 3 {
		t.Errorf("completed tasks = %d, want 3", len(execCtx.results))
	}
}

func TestExecuteBatchParallel(t *testing.T) {
	plan := &Plan{
		ID:   "parallel-batch-test",
		Goal: "Test parallel batch",
		Tasks: []Task{
			{ID: "1", Description: "Task 1", Dependencies: []string{}},
			{ID: "2", Description: "Task 2", Dependencies: []string{}},
			{ID: "3", Description: "Task 3", Dependencies: []string{}},
		},
	}

	mock := &mockAgent{
		chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
			time.Sleep(50 * time.Millisecond) // Simulate work
			return &ChatResult{Content: "done"}, nil
		},
	}

	config := DefaultPlannerConfig()
	config.MaxParallel = 3
	executor := NewExecutor(config, mock)
	execCtx := newExecutionContext(plan)

	ctx := context.Background()
	start := time.Now()
	err := executor.executeBatchParallel(ctx, plan.Tasks, plan, execCtx, 0)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify all tasks completed
	if len(execCtx.results) != 3 {
		t.Errorf("completed tasks = %d, want 3", len(execCtx.results))
	}

	// Parallel execution should be faster than sequential
	// 3 tasks * 50ms = 150ms sequential, but parallel should be ~50-100ms
	if duration > 120*time.Millisecond {
		t.Logf("Warning: parallel execution took %v, expected < 120ms", duration)
	}
}
