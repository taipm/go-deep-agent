package agent

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// PlanningStrategy defines the execution approach for a plan.
type PlanningStrategy string

const (
	// StrategySequential executes tasks one-by-one in dependency order (safest, default).
	StrategySequential PlanningStrategy = "sequential"

	// StrategyParallel executes independent tasks concurrently (fastest).
	StrategyParallel PlanningStrategy = "parallel"

	// StrategyAdaptive switches strategy based on runtime performance (smartest).
	StrategyAdaptive PlanningStrategy = "adaptive"
)

// PlanStatus tracks the overall state of a plan execution.
type PlanStatus string

const (
	// PlanStatusCreated indicates the plan has been created but not yet started.
	PlanStatusCreated PlanStatus = "created"

	// PlanStatusRunning indicates the plan execution is in progress.
	PlanStatusRunning PlanStatus = "running"

	// PlanStatusCompleted indicates all tasks have been successfully completed.
	PlanStatusCompleted PlanStatus = "completed"

	// PlanStatusFailed indicates the plan failed due to a critical error.
	PlanStatusFailed PlanStatus = "failed"

	// PlanStatusCanceled indicates the plan was canceled by the user.
	PlanStatusCanceled PlanStatus = "canceled"
)

// TaskStatus tracks the lifecycle state of a single task.
type TaskStatus string

const (
	// TaskStatusPending indicates the task has not yet started.
	TaskStatusPending TaskStatus = "pending"

	// TaskStatusRunning indicates the task is currently executing.
	TaskStatusRunning TaskStatus = "running"

	// TaskStatusCompleted indicates the task finished successfully.
	TaskStatusCompleted TaskStatus = "completed"

	// TaskStatusFailed indicates the task failed with an error.
	TaskStatusFailed TaskStatus = "failed"

	// TaskStatusSkipped indicates the task was skipped due to dependency failure.
	TaskStatusSkipped TaskStatus = "skipped"
)

// TaskType categorizes the purpose of a task.
type TaskType string

const (
	// TaskTypeAction represents a task that executes a tool or action.
	TaskTypeAction TaskType = "action"

	// TaskTypeDecision represents a task that makes a choice based on data.
	TaskTypeDecision TaskType = "decision"

	// TaskTypeObservation represents a task that gathers information.
	TaskTypeObservation TaskType = "observation"

	// TaskTypeAggregate represents a task that combines results from other tasks.
	TaskTypeAggregate TaskType = "aggregate"
)

// Plan represents a decomposed task plan with strategy and goal state.
type Plan struct {
	// ID is a unique identifier for the plan.
	ID string

	// Goal is the high-level objective to accomplish.
	Goal string

	// GoalState defines measurable success criteria for the plan.
	GoalState GoalState

	// Tasks contains all decomposed subtasks in a tree structure.
	Tasks []Task

	// Strategy determines how tasks are executed (sequential, parallel, adaptive).
	Strategy PlanningStrategy

	// CreatedAt is the timestamp when the plan was created.
	CreatedAt time.Time

	// EstimatedCost is the predicted token/API cost for executing this plan.
	EstimatedCost float64

	// Metadata stores additional custom plan data.
	Metadata map[string]interface{}
}

// NewPlan creates a new plan with the given goal and strategy.
func NewPlan(goal string, strategy PlanningStrategy) *Plan {
	return &Plan{
		ID:        generateID(),
		Goal:      goal,
		Strategy:  strategy,
		CreatedAt: time.Now(),
		Tasks:     []Task{},
		Metadata:  make(map[string]interface{}),
	}
}

// Task represents a single subtask in a plan.
type Task struct {
	// ID is a unique identifier for this task.
	ID string

	// ParentID is the ID of the parent task (empty for root tasks).
	ParentID string

	// Description explains what this task should accomplish.
	Description string

	// Type categorizes the task's purpose (action, decision, observation, aggregate).
	Type TaskType

	// Dependencies contains IDs of tasks that must complete before this one starts.
	Dependencies []string

	// Status tracks the task lifecycle (pending, running, completed, failed, skipped).
	Status TaskStatus

	// ReActSteps contains the ReAct execution steps if this task was executed.
	ReActSteps []ReActStep

	// Result stores the task's output value.
	Result interface{}

	// Error contains any error that occurred during execution.
	Error error

	// StartedAt is when task execution began.
	StartedAt time.Time

	// CompletedAt is when task execution finished.
	CompletedAt time.Time

	// Duration is how long the task took to execute.
	Duration time.Duration

	// Subtasks contains child tasks if this task was further decomposed.
	Subtasks []Task

	// Depth indicates the nesting level (0 = root task).
	Depth int
}

// GoalState defines measurable success criteria for a plan.
type GoalState struct {
	// Description is a human-readable explanation of the goal.
	Description string

	// Criteria contains all conditions that must be satisfied.
	Criteria []GoalCriterion

	// Satisfied indicates whether the goal has been achieved.
	Satisfied bool

	// Progress represents completion percentage (0.0 to 1.0).
	Progress float64

	// CheckedAt is when the goal was last evaluated.
	CheckedAt time.Time
}

// GoalCriterion represents a single measurable condition.
type GoalCriterion struct {
	// Name identifies this criterion.
	Name string

	// Operator defines the comparison type (==, !=, >, <, >=, <=, contains, matches).
	Operator string

	// Expected is the target value for this criterion.
	Expected interface{}

	// Actual is the current observed value.
	Actual interface{}

	// Satisfied indicates whether this criterion is met.
	Satisfied bool

	// Weight represents the importance of this criterion (0.0 to 1.0, default: 1.0).
	Weight float64
}

// PlanResult contains the results of plan execution.
type PlanResult struct {
	// Plan is the executed plan with all its tasks.
	Plan Plan

	// Status indicates the overall plan outcome.
	Status PlanStatus

	// Progress represents overall completion (0.0 to 1.0).
	Progress float64

	// CompletedTasks is the number of successfully completed tasks.
	CompletedTasks int

	// FailedTasks is the number of tasks that failed.
	FailedTasks int

	// SkippedTasks is the number of tasks that were skipped.
	SkippedTasks int

	// TotalTasks is the total number of tasks in the plan.
	TotalTasks int

	// StartedAt is when plan execution began.
	StartedAt time.Time

	// CompletedAt is when plan execution finished.
	CompletedAt time.Time

	// Duration is how long the plan took to execute.
	Duration time.Duration

	// FinalResult contains the final output of the plan.
	FinalResult interface{}

	// Error contains any critical error that stopped the plan.
	Error error

	// Metrics provides detailed performance statistics.
	Metrics PlanMetrics

	// Timeline contains all execution events for observability.
	Timeline []PlanEvent
}

// PlanMetrics tracks performance statistics for plan execution.
type PlanMetrics struct {
	// DecompositionTime is how long it took to decompose the goal into tasks.
	DecompositionTime time.Duration

	// TaskCount is the total number of tasks created.
	TaskCount int

	// AvgTaskDepth is the average nesting level of tasks.
	AvgTaskDepth float64

	// ExecutionTime is the total time spent executing tasks.
	ExecutionTime time.Duration

	// AvgTaskDuration is the average time per task.
	AvgTaskDuration time.Duration

	// ParallelTasks is the number of tasks executed concurrently.
	ParallelTasks int

	// SuccessRate is the percentage of tasks completed successfully (0.0 to 1.0).
	SuccessRate float64

	// GoalAchieved indicates whether the plan's goal was met.
	GoalAchieved bool

	// GoalProgress is the final goal completion percentage (0.0 to 1.0).
	GoalProgress float64

	// TotalTokens is the total number of LLM tokens consumed.
	TotalTokens int

	// EstimatedCost is the predicted cost before execution.
	EstimatedCost float64

	// ActualCost is the actual cost after execution.
	ActualCost float64
}

// PlanEvent records a significant event during plan execution.
type PlanEvent struct {
	// Timestamp is when the event occurred.
	Timestamp time.Time

	// Type categorizes the event (plan_created, task_started, task_completed, goal_checked, strategy_switched, etc).
	Type string

	// TaskID identifies the related task (if applicable).
	TaskID string

	// Description provides human-readable details about the event.
	Description string

	// Data contains additional event-specific information.
	Data map[string]interface{}
}

// generateID creates a unique identifier for plans and tasks.
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// IsCompleted returns true if the task finished successfully.
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

// IsFailed returns true if the task failed with an error.
func (t *Task) IsFailed() bool {
	return t.Status == TaskStatusFailed
}

// AddTask appends a new task to the plan.
func (p *Plan) AddTask(task Task) {
	p.Tasks = append(p.Tasks, task)
}

// GetTaskByID finds a task by its ID (searches recursively through subtasks).
func (p *Plan) GetTaskByID(id string) *Task {
	for i := range p.Tasks {
		if task := findTaskByID(&p.Tasks[i], id); task != nil {
			return task
		}
	}
	return nil
}

// findTaskByID recursively searches for a task by ID.
func findTaskByID(task *Task, id string) *Task {
	if task.ID == id {
		return task
	}
	for i := range task.Subtasks {
		if found := findTaskByID(&task.Subtasks[i], id); found != nil {
			return found
		}
	}
	return nil
}
