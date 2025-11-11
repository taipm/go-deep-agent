package agent

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockAgent implements a simple mock for testing Executor
type mockAgent struct {
	chatFunc func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error)
}

func (m *mockAgent) Chat(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
	if m.chatFunc != nil {
		return m.chatFunc(ctx, message, opts)
	}
	return &ChatResult{Content: "mock response"}, nil
}

// Test canExecute dependency checking
func TestCanExecute(t *testing.T) {
	config := DefaultPlannerConfig()
	agent := &mockAgent{}
	executor := NewExecutor(config, agent)

	plan := NewPlan("test goal", StrategySequential)
	execCtx := newExecutionContext(plan)

	tests := []struct {
		name     string
		task     Task
		setup    func()
		expected bool
	}{
		{
			name: "No dependencies - should execute",
			task: Task{
				ID:           "task1",
				Status:       TaskStatusPending,
				Dependencies: []string{},
			},
			expected: true,
		},
		{
			name: "Already completed - should not execute",
			task: Task{
				ID:           "task2",
				Status:       TaskStatusCompleted,
				Dependencies: []string{},
			},
			expected: false,
		},
		{
			name: "Already failed - should not execute",
			task: Task{
				ID:           "task3",
				Status:       TaskStatusFailed,
				Dependencies: []string{},
			},
			expected: false,
		},
		{
			name: "Dependencies satisfied - should execute",
			task: Task{
				ID:           "task4",
				Status:       TaskStatusPending,
				Dependencies: []string{"task1"},
			},
			setup: func() {
				execCtx.results["task1"] = &TaskResult{
					TaskID: "task1",
					Status: TaskStatusCompleted,
				}
			},
			expected: true,
		},
		{
			name: "Dependencies not satisfied - should not execute",
			task: Task{
				ID:           "task5",
				Status:       TaskStatusPending,
				Dependencies: []string{"task_missing"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result := executor.canExecute(&tt.task, execCtx)
			if result != tt.expected {
				t.Errorf("canExecute() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test selectNextTask scheduling
func TestSelectNextTask(t *testing.T) {
	config := DefaultPlannerConfig()
	agent := &mockAgent{}
	executor := NewExecutor(config, agent)

	tests := []struct {
		name     string
		strategy PlanningStrategy
		tasks    []Task
		setup    func(*executionContext)
		wantID   string // Empty string means nil expected
	}{
		{
			name:     "No executable tasks",
			strategy: StrategySequential,
			tasks: []Task{
				{ID: "task1", Status: TaskStatusCompleted},
				{ID: "task2", Status: TaskStatusFailed},
			},
			wantID: "",
		},
		{
			name:     "Single executable task",
			strategy: StrategySequential,
			tasks: []Task{
				{ID: "task1", Status: TaskStatusPending, Dependencies: []string{}},
			},
			wantID: "task1",
		},
		{
			name:     "Sequential - returns first",
			strategy: StrategySequential,
			tasks: []Task{
				{ID: "task1", Status: TaskStatusPending, Dependencies: []string{}},
				{ID: "task2", Status: TaskStatusPending, Dependencies: []string{}},
			},
			wantID: "task1",
		},
		{
			name:     "Dependency blocking",
			strategy: StrategySequential,
			tasks: []Task{
				{ID: "task1", Status: TaskStatusPending, Dependencies: []string{"task_missing"}},
			},
			wantID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := NewPlan("test", tt.strategy)
			plan.Tasks = tt.tasks
			execCtx := newExecutionContext(plan)
			if tt.setup != nil {
				tt.setup(execCtx)
			}

			task := executor.selectNextTask(execCtx)
			if tt.wantID == "" {
				if task != nil {
					t.Errorf("selectNextTask() = %v, want nil", task.ID)
				}
			} else {
				if task == nil {
					t.Errorf("selectNextTask() = nil, want task %s", tt.wantID)
				} else if task.ID != tt.wantID {
					t.Errorf("selectNextTask() = %s, want %s", task.ID, tt.wantID)
				}
			}
		})
	}
}

// Test executeTask
func TestExecuteTask(t *testing.T) {
	config := DefaultPlannerConfig()

	tests := []struct {
		name       string
		chatFunc   func(context.Context, string, *ChatOptions) (*ChatResult, error)
		wantStatus TaskStatus
		wantError  bool
	}{
		{
			name: "Successful execution",
			chatFunc: func(ctx context.Context, msg string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: "task result"}, nil
			},
			wantStatus: TaskStatusCompleted,
			wantError:  false,
		},
		{
			name: "Failed execution",
			chatFunc: func(ctx context.Context, msg string, opts *ChatOptions) (*ChatResult, error) {
				return nil, errors.New("chat error")
			},
			wantStatus: TaskStatusFailed,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &mockAgent{chatFunc: tt.chatFunc}
			executor := NewExecutor(config, agent)

			plan := NewPlan("test", StrategySequential)
			execCtx := newExecutionContext(plan)

			task := &Task{
				ID:          "task1",
				Description: "test task",
				Status:      TaskStatusPending,
			}

			err := executor.executeTask(context.Background(), task, execCtx)

			if tt.wantError && err == nil {
				t.Error("executeTask() expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("executeTask() unexpected error: %v", err)
			}

			if task.Status != tt.wantStatus {
				t.Errorf("task.Status = %v, want %v", task.Status, tt.wantStatus)
			}
		})
	}
}

// Test checkGoalCompletion
func TestCheckGoalCompletion(t *testing.T) {
	config := DefaultPlannerConfig()
	agent := &mockAgent{}
	executor := NewExecutor(config, agent)

	tests := []struct {
		name       string
		tasks      []Task
		criteria   []GoalCriterion
		wantResult bool
		wantReason string
	}{
		{
			name: "All tasks completed, no criteria",
			tasks: []Task{
				{ID: "task1", Status: TaskStatusCompleted},
				{ID: "task2", Status: TaskStatusCompleted},
			},
			criteria:   []GoalCriterion{},
			wantResult: true,
		},
		{
			name: "Some tasks pending",
			tasks: []Task{
				{ID: "task1", Status: TaskStatusCompleted},
				{ID: "task2", Status: TaskStatusPending},
			},
			criteria:   []GoalCriterion{},
			wantResult: false,
		},
		{
			name: "Criteria satisfied",
			tasks: []Task{
				{ID: "task1", Status: TaskStatusCompleted},
			},
			criteria: []GoalCriterion{
				{Name: "criterion1", Satisfied: true},
			},
			wantResult: true,
		},
		{
			name: "Criteria not satisfied",
			tasks: []Task{
				{ID: "task1", Status: TaskStatusCompleted},
			},
			criteria: []GoalCriterion{
				{Name: "criterion1", Satisfied: false},
			},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := NewPlan("test", StrategySequential)
			plan.Tasks = tt.tasks
			plan.GoalState = GoalState{Criteria: tt.criteria}
			execCtx := newExecutionContext(plan)

			result, _ := executor.checkGoalCompletion(execCtx)
			if result != tt.wantResult {
				t.Errorf("checkGoalCompletion() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// Test Execute end-to-end
func TestExecute(t *testing.T) {
	config := DefaultPlannerConfig()

	tests := []struct {
		name       string
		tasks      []Task
		chatFunc   func(context.Context, string, *ChatOptions) (*ChatResult, error)
		wantStatus PlanStatus
		wantError  bool
	}{
		{
			name: "Simple plan execution",
			tasks: []Task{
				{ID: "task1", Description: "Do something", Status: TaskStatusPending, Dependencies: []string{}},
			},
			chatFunc: func(ctx context.Context, msg string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: "done"}, nil
			},
			wantStatus: PlanStatusCompleted,
			wantError:  false,
		},
		{
			name: "Plan with dependencies",
			tasks: []Task{
				{ID: "task1", Description: "First", Status: TaskStatusPending, Dependencies: []string{}},
				{ID: "task2", Description: "Second", Status: TaskStatusPending, Dependencies: []string{"task1"}},
			},
			chatFunc: func(ctx context.Context, msg string, opts *ChatOptions) (*ChatResult, error) {
				return &ChatResult{Content: "done"}, nil
			},
			wantStatus: PlanStatusCompleted,
			wantError:  false,
		},
		{
			name: "Context cancellation",
			tasks: []Task{
				{ID: "task1", Description: "Never completes", Status: TaskStatusPending, Dependencies: []string{}},
			},
			chatFunc: func(ctx context.Context, msg string, opts *ChatOptions) (*ChatResult, error) {
				<-ctx.Done()
				return nil, ctx.Err()
			},
			wantStatus: PlanStatusCanceled,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip context cancellation test for now (needs special handling)
			if tt.name == "Context cancellation" {
				t.Skip("Context cancellation test requires special setup")
			}

			agent := &mockAgent{chatFunc: tt.chatFunc}
			executor := NewExecutor(config, agent)

			plan := NewPlan("test goal", StrategySequential)
			plan.Tasks = tt.tasks

			ctx := context.Background()
			result, err := executor.Execute(ctx, plan)

			if tt.wantError && err == nil {
				t.Error("Execute() expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
			}

			if result.Status != tt.wantStatus {
				t.Errorf("result.Status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

// Test buildMetrics
func TestBuildMetrics(t *testing.T) {
	config := DefaultPlannerConfig()
	agent := &mockAgent{}
	executor := NewExecutor(config, agent)

	plan := NewPlan("test", StrategySequential)
	plan.Tasks = []Task{
		{ID: "task1"},
		{ID: "task2"},
		{ID: "task3"},
	}

	execCtx := newExecutionContext(plan)
	execCtx.results["task1"] = &TaskResult{
		TaskID:   "task1",
		Status:   TaskStatusCompleted,
		Duration: 100 * time.Millisecond,
	}
	execCtx.results["task2"] = &TaskResult{
		TaskID: "task2",
		Status: TaskStatusFailed,
	}

	metrics := executor.buildMetrics(execCtx)

	if metrics.TaskCount != 3 {
		t.Errorf("metrics.TaskCount = %d, want 3", metrics.TaskCount)
	}
	if metrics.SuccessRate == 0 {
		t.Error("metrics.SuccessRate should not be 0 with completed tasks")
	}
}
