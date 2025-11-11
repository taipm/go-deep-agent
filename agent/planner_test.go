package agent

import (
	"testing"
	"time"
)

// Test enum values
func TestPlanningStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy PlanningStrategy
		expected string
	}{
		{"Sequential", StrategySequential, "sequential"},
		{"Parallel", StrategyParallel, "parallel"},
		{"Adaptive", StrategyAdaptive, "adaptive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.strategy) != tt.expected {
				t.Errorf("Strategy = %v, want %v", tt.strategy, tt.expected)
			}
		})
	}
}

func TestPlanStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   PlanStatus
		expected string
	}{
		{"Created", PlanStatusCreated, "created"},
		{"Running", PlanStatusRunning, "running"},
		{"Completed", PlanStatusCompleted, "completed"},
		{"Failed", PlanStatusFailed, "failed"},
		{"Canceled", PlanStatusCanceled, "canceled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Status = %v, want %v", tt.status, tt.expected)
			}
		})
	}
}

func TestTaskStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected string
	}{
		{"Pending", TaskStatusPending, "pending"},
		{"Running", TaskStatusRunning, "running"},
		{"Completed", TaskStatusCompleted, "completed"},
		{"Failed", TaskStatusFailed, "failed"},
		{"Skipped", TaskStatusSkipped, "skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Status = %v, want %v", tt.status, tt.expected)
			}
		})
	}
}

func TestTaskType(t *testing.T) {
	tests := []struct {
		name     string
		taskType TaskType
		expected string
	}{
		{"Action", TaskTypeAction, "action"},
		{"Decision", TaskTypeDecision, "decision"},
		{"Observation", TaskTypeObservation, "observation"},
		{"Aggregate", TaskTypeAggregate, "aggregate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.taskType) != tt.expected {
				t.Errorf("Type = %v, want %v", tt.taskType, tt.expected)
			}
		})
	}
}

// Test Plan creation
func TestNewPlan(t *testing.T) {
	goal := "Test goal"
	strategy := StrategySequential

	plan := NewPlan(goal, strategy)

	if plan == nil {
		t.Fatal("NewPlan returned nil")
	}
	if plan.Goal != goal {
		t.Errorf("Goal = %v, want %v", plan.Goal, goal)
	}
	if plan.Strategy != strategy {
		t.Errorf("Strategy = %v, want %v", plan.Strategy, strategy)
	}
	if plan.ID == "" {
		t.Error("ID should not be empty")
	}
	if plan.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if plan.Tasks == nil {
		t.Error("Tasks should be initialized")
	}
	if plan.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

// Test generateID uniqueness
func TestGenerateID(t *testing.T) {
	seen := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id := generateID()
		if id == "" {
			t.Error("generateID returned empty string")
		}
		if seen[id] {
			t.Errorf("generateID produced duplicate: %s", id)
		}
		seen[id] = true
	}

	if len(seen) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(seen))
	}
}

// Test Task methods
func TestTaskIsCompleted(t *testing.T) {
	task := &Task{Status: TaskStatusCompleted}
	if !task.IsCompleted() {
		t.Error("IsCompleted should return true for completed task")
	}

	task.Status = TaskStatusPending
	if task.IsCompleted() {
		t.Error("IsCompleted should return false for pending task")
	}
}

func TestTaskIsFailed(t *testing.T) {
	task := &Task{Status: TaskStatusFailed}
	if !task.IsFailed() {
		t.Error("IsFailed should return true for failed task")
	}

	task.Status = TaskStatusCompleted
	if task.IsFailed() {
		t.Error("IsFailed should return false for completed task")
	}
}

// Test Plan.AddTask
func TestPlanAddTask(t *testing.T) {
	plan := NewPlan("test", StrategySequential)

	task1 := Task{ID: "task1", Description: "First task"}
	task2 := Task{ID: "task2", Description: "Second task"}

	plan.AddTask(task1)
	if len(plan.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(plan.Tasks))
	}

	plan.AddTask(task2)
	if len(plan.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(plan.Tasks))
	}

	if plan.Tasks[0].ID != "task1" {
		t.Errorf("First task ID = %v, want task1", plan.Tasks[0].ID)
	}
	if plan.Tasks[1].ID != "task2" {
		t.Errorf("Second task ID = %v, want task2", plan.Tasks[1].ID)
	}
}

// Test Plan.GetTaskByID
func TestPlanGetTaskByID(t *testing.T) {
	plan := NewPlan("test", StrategySequential)

	task1 := Task{ID: "task1", Description: "Root task"}
	task2 := Task{
		ID:          "task2",
		Description: "Parent task",
		Subtasks: []Task{
			{ID: "task3", Description: "Child task"},
		},
	}

	plan.AddTask(task1)
	plan.AddTask(task2)

	// Test finding root task
	found := plan.GetTaskByID("task1")
	if found == nil {
		t.Fatal("GetTaskByID failed to find task1")
	}
	if found.ID != "task1" {
		t.Errorf("Found task ID = %v, want task1", found.ID)
	}

	// Test finding nested task
	found = plan.GetTaskByID("task3")
	if found == nil {
		t.Fatal("GetTaskByID failed to find nested task3")
	}
	if found.ID != "task3" {
		t.Errorf("Found task ID = %v, want task3", found.ID)
	}

	// Test not found
	found = plan.GetTaskByID("nonexistent")
	if found != nil {
		t.Error("GetTaskByID should return nil for nonexistent task")
	}
}

// Test GoalState and GoalCriterion
func TestGoalState(t *testing.T) {
	goal := GoalState{
		Description: "Complete all tasks",
		Criteria: []GoalCriterion{
			{Name: "tasks_done", Operator: "==", Expected: 5, Actual: 5, Satisfied: true, Weight: 1.0},
			{Name: "quality", Operator: ">=", Expected: 0.8, Actual: 0.9, Satisfied: true, Weight: 0.8},
		},
		Satisfied: true,
		Progress:  1.0,
	}

	if goal.Description == "" {
		t.Error("Description should not be empty")
	}
	if len(goal.Criteria) != 2 {
		t.Errorf("Expected 2 criteria, got %d", len(goal.Criteria))
	}
	if !goal.Satisfied {
		t.Error("Goal should be satisfied")
	}
	if goal.Progress != 1.0 {
		t.Errorf("Progress = %v, want 1.0", goal.Progress)
	}
}

// Test PlanResult creation
func TestPlanResult(t *testing.T) {
	plan := NewPlan("test goal", StrategySequential)
	result := &PlanResult{
		Plan:           *plan,
		Status:         PlanStatusCompleted,
		Progress:       1.0,
		CompletedTasks: 5,
		FailedTasks:    0,
		SkippedTasks:   0,
		TotalTasks:     5,
		StartedAt:      time.Now().Add(-1 * time.Minute),
		CompletedAt:    time.Now(),
		Duration:       1 * time.Minute,
		FinalResult:    "Success",
		Error:          nil,
	}

	if result.Status != PlanStatusCompleted {
		t.Errorf("Status = %v, want %v", result.Status, PlanStatusCompleted)
	}
	if result.Progress != 1.0 {
		t.Errorf("Progress = %v, want 1.0", result.Progress)
	}
	if result.CompletedTasks != 5 {
		t.Errorf("CompletedTasks = %v, want 5", result.CompletedTasks)
	}
	if result.FinalResult != "Success" {
		t.Errorf("FinalResult = %v, want Success", result.FinalResult)
	}
}

// Test PlanMetrics
func TestPlanMetrics(t *testing.T) {
	metrics := PlanMetrics{
		DecompositionTime: 500 * time.Millisecond,
		TaskCount:         10,
		AvgTaskDepth:      2.5,
		ExecutionTime:     5 * time.Second,
		AvgTaskDuration:   500 * time.Millisecond,
		ParallelTasks:     3,
		SuccessRate:       0.9,
		GoalAchieved:      true,
		GoalProgress:      1.0,
		TotalTokens:       1000,
		EstimatedCost:     0.05,
		ActualCost:        0.048,
	}

	if metrics.TaskCount != 10 {
		t.Errorf("TaskCount = %v, want 10", metrics.TaskCount)
	}
	if metrics.SuccessRate != 0.9 {
		t.Errorf("SuccessRate = %v, want 0.9", metrics.SuccessRate)
	}
	if !metrics.GoalAchieved {
		t.Error("GoalAchieved should be true")
	}
}

// Test PlanEvent
func TestPlanEvent(t *testing.T) {
	event := PlanEvent{
		Timestamp:   time.Now(),
		Type:        "task_completed",
		TaskID:      "task1",
		Description: "Task 1 completed successfully",
		Data: map[string]interface{}{
			"duration": 2.5,
			"tokens":   100,
		},
	}

	if event.Type != "task_completed" {
		t.Errorf("Type = %v, want task_completed", event.Type)
	}
	if event.TaskID != "task1" {
		t.Errorf("TaskID = %v, want task1", event.TaskID)
	}
	if event.Data == nil {
		t.Error("Data should not be nil")
	}
	if event.Data["duration"] != 2.5 {
		t.Errorf("Duration = %v, want 2.5", event.Data["duration"])
	}
}
