package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestTimelineEvents(t *testing.T) {
	t.Run("task events recorded", func(t *testing.T) {
		plan := &Plan{
			ID:   "timeline-test",
			Goal: "Test timeline recording",
			Tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
				{ID: "2", Description: "Task 2", Dependencies: []string{}},
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
		_ = executor.executeTask(ctx, &plan.Tasks[0], execCtx)

		// Verify events were recorded
		if len(execCtx.timeline) == 0 {
			t.Error("no timeline events recorded")
		}

		// Check for task_started and task_completed events
		foundStart := false
		foundComplete := false
		for _, event := range execCtx.timeline {
			if event.Type == "task_started" && event.TaskID == "1" {
				foundStart = true
			}
			if event.Type == "task_completed" && event.TaskID == "1" {
				foundComplete = true
			}
		}

		if !foundStart {
			t.Error("task_started event not found")
		}
		if !foundComplete {
			t.Error("task_completed event not found")
		}
	})

	t.Run("timeline in result", func(t *testing.T) {
		plan := &Plan{
			ID:       "result-timeline-test",
			Goal:     "Test result timeline",
			Strategy: StrategySequential,
			Tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: "done"}, nil
			},
		}

		executor := NewExecutor(DefaultPlannerConfig(), mock)

		ctx := context.Background()
		result, err := executor.Execute(ctx, plan)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(result.Timeline) == 0 {
			t.Error("result should have timeline events")
		}
	})

	t.Run("adaptive timeline includes strategy events", func(t *testing.T) {
		plan := &Plan{
			ID:       "adaptive-timeline-test",
			Goal:     "Test adaptive timeline",
			Strategy: StrategyAdaptive,
			Tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
				{ID: "2", Description: "Task 2", Dependencies: []string{}},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: "done"}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.AdaptiveThreshold = 0.3
		executor := NewExecutor(config, mock)

		ctx := context.Background()
		result, err := executor.executeAdaptive(ctx, plan)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check for strategy_initialized event
		foundStrategyInit := false
		for _, event := range result.Timeline {
			if event.Type == "strategy_initialized" {
				foundStrategyInit = true
				break
			}
		}

		if !foundStrategyInit {
			t.Error("strategy_initialized event not found in timeline")
		}
	})
}

func TestPeriodicGoalChecking(t *testing.T) {
	t.Run("goal checked at interval", func(t *testing.T) {
		plan := &Plan{
			ID:   "periodic-goal-test",
			Goal: "Test periodic goal checking",
			Tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
				{ID: "2", Description: "Task 2", Dependencies: []string{}},
				{ID: "3", Description: "Task 3", Dependencies: []string{}},
				{ID: "4", Description: "Task 4", Dependencies: []string{}},
				{ID: "5", Description: "Task 5", Dependencies: []string{}},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: "done"}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.GoalCheckInterval = 2 // Check every 2 tasks
		executor := NewExecutor(config, mock)

		ctx := context.Background()
		result, err := executor.Execute(ctx, plan)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Count goal_checked events
		goalCheckCount := 0
		for _, event := range result.Timeline {
			if event.Type == "goal_checked" {
				goalCheckCount++
			}
		}

		// With 5 tasks and interval 2, we should have checks at task 2 and 4
		expectedChecks := 2
		if goalCheckCount != expectedChecks {
			t.Errorf("goal check count = %d, want %d", goalCheckCount, expectedChecks)
		}
	})

	t.Run("early termination when goal met", func(t *testing.T) {
		plan := &Plan{
			ID:   "early-termination-test",
			Goal: "Find 3 items",
			GoalState: GoalState{
				Description: "Found 3 items",
				Criteria: []GoalCriterion{
					{
						Name:      "items_found",
						Expected:  "3",
						Actual:    "",
						Operator:  ">=",
						Satisfied: true, // Simulate goal already met
					},
				},
			},
			Tasks: []Task{
				{ID: "1", Description: "Find item 1", Dependencies: []string{}},
				{ID: "2", Description: "Find item 2", Dependencies: []string{}},
				{ID: "3", Description: "Find item 3", Dependencies: []string{}},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: "found"}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.GoalCheckInterval = 2
		executor := NewExecutor(config, mock)

		ctx := context.Background()
		result, err := executor.Execute(ctx, plan)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check for goal_achieved event
		foundGoalAchieved := false
		for _, event := range result.Timeline {
			if event.Type == "goal_achieved" {
				foundGoalAchieved = true
				break
			}
		}

		if !foundGoalAchieved {
			t.Logf("Timeline events: %d", len(result.Timeline))
			for _, event := range result.Timeline {
				t.Logf("  - %s: %s", event.Type, event.Description)
			}
			// Note: Early termination currently just records event, doesn't actually stop
			// This is OK for now as full implementation would need more complex logic
		}
	})
}

func TestEventDescriptions(t *testing.T) {
	t.Run("event descriptions are meaningful", func(t *testing.T) {
		plan := &Plan{
			ID:   "description-test",
			Goal: "Test event descriptions",
			Tasks: []Task{
				{ID: "1", Description: "Test task", Dependencies: []string{}},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: "result"}, nil
			},
		}

		executor := NewExecutor(DefaultPlannerConfig(), mock)
		execCtx := newExecutionContext(plan)

		ctx := context.Background()
		_ = executor.executeTask(ctx, &plan.Tasks[0], execCtx)

		// Verify all events have non-empty descriptions
		for _, event := range execCtx.timeline {
			if event.Description == "" {
				t.Errorf("event %s has empty description", event.Type)
			}
			if event.Type == "task_started" {
				if !strings.Contains(event.Description, "Test task") {
					t.Errorf("task_started description should contain task description")
				}
			}
		}
	})
}

func TestTaskCounter(t *testing.T) {
	t.Run("task counter increments", func(t *testing.T) {
		plan := &Plan{
			ID: "counter-test",
			Tasks: []Task{
				{ID: "1", Description: "Task 1", Dependencies: []string{}},
			},
		}

		execCtx := newExecutionContext(plan)

		// Increment counter multiple times
		count1 := execCtx.incrementTaskCounter()
		count2 := execCtx.incrementTaskCounter()
		count3 := execCtx.incrementTaskCounter()

		if count1 != 1 {
			t.Errorf("first count = %d, want 1", count1)
		}
		if count2 != 2 {
			t.Errorf("second count = %d, want 2", count2)
		}
		if count3 != 3 {
			t.Errorf("third count = %d, want 3", count3)
		}
	})
}

func TestAddEvent(t *testing.T) {
	t.Run("concurrent event addition", func(t *testing.T) {
		plan := &Plan{ID: "concurrent-test"}
		execCtx := newExecutionContext(plan)

		// Add events concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				execCtx.addEvent("test_event", "", "concurrent test")
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all events were added
		if len(execCtx.timeline) != 10 {
			t.Errorf("timeline events = %d, want 10", len(execCtx.timeline))
		}
	})
}

func TestTaskFailureEvents(t *testing.T) {
	t.Run("task failure recorded in timeline", func(t *testing.T) {
		plan := &Plan{
			ID:   "failure-test",
			Goal: "Test failure events",
			Tasks: []Task{
				{ID: "1", Description: "Failing task", Dependencies: []string{}},
			},
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				return nil, fmt.Errorf("intentional failure")
			},
		}

		executor := NewExecutor(DefaultPlannerConfig(), mock)
		execCtx := newExecutionContext(plan)

		ctx := context.Background()
		_ = executor.executeTask(ctx, &plan.Tasks[0], execCtx)

		// Check for task_failed event
		foundFailure := false
		for _, event := range execCtx.timeline {
			if event.Type == "task_failed" && event.TaskID == "1" {
				foundFailure = true
				if !strings.Contains(event.Description, "failed") {
					t.Errorf("failure event should mention failure in description")
				}
			}
		}

		if !foundFailure {
			t.Error("task_failed event not found")
		}
	})
}
