package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/openai/openai-go/v3"
)

// agentExecutor defines the interface for executing tasks through an agent.
// This allows for easy mocking in tests.
type agentExecutor interface {
	Chat(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error)
}

// Executor executes plans by orchestrating task execution through ReAct cycles.
// It manages dependencies, tracks progress, and evaluates goal completion.
type Executor struct {
	config *PlannerConfig
	agent  agentExecutor
}

// NewExecutor creates a new Executor with the given configuration and agent.
// If config is nil, it uses DefaultPlannerConfig().
func NewExecutor(config *PlannerConfig, agent agentExecutor) *Executor {
	if config == nil {
		config = DefaultPlannerConfig()
	}
	return &Executor{
		config: config,
		agent:  agent,
	}
}

// executionContext tracks the state during plan execution.
type executionContext struct {
	plan      *Plan
	results   map[string]*TaskResult // Map of task ID to execution result
	startTime time.Time
	mu        sync.RWMutex // Protects concurrent access to results
}

// TaskResult captures the outcome of executing a single task.
type TaskResult struct {
	TaskID     string
	Status     TaskStatus
	Output     string
	Error      error
	ReActSteps []ReActStep
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
}

// newExecutionContext creates a new execution context for a plan.
func newExecutionContext(plan *Plan) *executionContext {
	return &executionContext{
		plan:      plan,
		results:   make(map[string]*TaskResult),
		startTime: time.Now(),
	}
}

// canExecute checks if a task's dependencies are satisfied and it's ready to run.
// A task can execute if all its dependencies have completed successfully.
func (e *Executor) canExecute(task *Task, ctx *executionContext) bool {
	// Already completed or failed
	if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed {
		return false
	}

	// Check all dependencies
	for _, depID := range task.Dependencies {
		ctx.mu.RLock()
		result, exists := ctx.results[depID]
		ctx.mu.RUnlock()

		if !exists || result.Status != TaskStatusCompleted {
			return false
		}
	}

	return true
}

// selectNextTask finds the next task to execute based on the planning strategy.
// Returns nil if no executable task is found.
func (e *Executor) selectNextTask(ctx *executionContext) *Task {
	var candidates []*Task

	// Find all executable tasks
	for i := range ctx.plan.Tasks {
		task := &ctx.plan.Tasks[i]
		if e.canExecute(task, ctx) {
			candidates = append(candidates, task)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// For sequential strategy, return first candidate
	if ctx.plan.Strategy == StrategySequential {
		return candidates[0]
	}

	// For parallel/adaptive, return first candidate (parallel execution handled by caller)
	return candidates[0]
}

// executeTask executes a single task using the agent's ReAct capabilities.
func (e *Executor) executeTask(ctx context.Context, task *Task, execCtx *executionContext) error {
	startTime := time.Now()

	// Update task status
	task.Status = TaskStatusRunning
	task.StartedAt = startTime

	// Create task result
	result := &TaskResult{
		TaskID:    task.ID,
		Status:    TaskStatusRunning,
		StartTime: startTime,
	}

	// Build chat options for ReAct execution
	opts := &ChatOptions{
		Tools: []openai.ChatCompletionToolUnionParam{}, // Tools would be populated from agent config
	}

	// Execute task through agent
	chatResult, err := e.agent.Chat(ctx, task.Description, opts)
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err
		task.CompletedAt = time.Now()
		task.Duration = time.Since(startTime)

		result.Status = TaskStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = time.Since(startTime)

		execCtx.mu.Lock()
		execCtx.results[task.ID] = result
		execCtx.mu.Unlock()

		return fmt.Errorf("task %s failed: %w", task.ID, err)
	}

	// Update task with results
	task.Status = TaskStatusCompleted
	task.Result = chatResult.Content
	task.CompletedAt = time.Now()
	task.Duration = time.Since(startTime)

	// Extract ReAct steps if available (placeholder for future ReAct integration)
	// For now, store empty ReActSteps as the structure exists in Task
	task.ReActSteps = []ReActStep{}
	result.ReActSteps = []ReActStep{}

	result.Status = TaskStatusCompleted
	result.Output = chatResult.Content
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)

	execCtx.mu.Lock()
	execCtx.results[task.ID] = result
	execCtx.mu.Unlock()

	return nil
}

// executeSubtasks recursively executes all subtasks of a task.
func (e *Executor) executeSubtasks(ctx context.Context, task *Task, execCtx *executionContext) error {
	if len(task.Subtasks) == 0 {
		return nil
	}

	for i := range task.Subtasks {
		subtask := &task.Subtasks[i]

		// Check if subtask can execute
		if !e.canExecute(subtask, execCtx) {
			continue
		}

		// Execute subtask
		if err := e.executeTask(ctx, subtask, execCtx); err != nil {
			return fmt.Errorf("subtask %s failed: %w", subtask.ID, err)
		}

		// Recursively execute nested subtasks
		if err := e.executeSubtasks(ctx, subtask, execCtx); err != nil {
			return err
		}
	}

	return nil
}

// checkGoalCompletion evaluates whether the plan's goal criteria are satisfied.
func (e *Executor) checkGoalCompletion(execCtx *executionContext) (bool, string) {
	plan := execCtx.plan

	// If no goal state criteria defined, check if all tasks are completed
	if len(plan.GoalState.Criteria) == 0 {
		allCompleted := true
		for _, task := range plan.Tasks {
			if task.Status != TaskStatusCompleted {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			return true, "All tasks completed successfully"
		}
		return false, "Some tasks are still pending or failed"
	}

	// Evaluate each criterion
	for _, criterion := range plan.GoalState.Criteria {
		// Check if criterion is satisfied
		// In a full implementation, this would evaluate Expected vs Actual
		if !criterion.Satisfied {
			return false, fmt.Sprintf("Criterion not met: %s", criterion.Name)
		}
	}

	return true, "All goal criteria satisfied"
}

// Execute runs the plan by executing tasks according to dependencies and strategy.
// It returns a PlanResult with execution details and metrics.
func (e *Executor) Execute(ctx context.Context, plan *Plan) (*PlanResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan cannot be nil")
	}

	// Create execution context
	execCtx := newExecutionContext(plan)
	startTime := time.Now()

	// Create result object to track execution
	result := &PlanResult{
		Plan:      *plan,
		Status:    PlanStatusRunning,
		StartedAt: startTime,
	}

	// Main execution loop
	maxIterations := e.config.MaxSubtasks * 10 // Prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++

		// Check context cancellation
		select {
		case <-ctx.Done():
			result.Status = PlanStatusCanceled
			result.CompletedAt = time.Now()
			result.Duration = time.Since(startTime)
			result.Error = ctx.Err()
			result.Metrics = e.buildMetrics(execCtx)
			return result, ctx.Err()
		default:
		}

		// Select next task to execute
		task := e.selectNextTask(execCtx)
		if task == nil {
			// No more executable tasks - check if we're done
			goalMet, reason := e.checkGoalCompletion(execCtx)
			if goalMet {
				result.Status = PlanStatusCompleted
				result.CompletedAt = time.Now()
				result.Duration = time.Since(startTime)
				result.FinalResult = reason
				result.Metrics = e.buildMetrics(execCtx)
				result.Metrics.GoalAchieved = true
				return result, nil
			}

			// Check if any tasks failed
			anyFailed := false
			for _, task := range plan.Tasks {
				if task.Status == TaskStatusFailed {
					anyFailed = true
					break
				}
			}

			if anyFailed {
				result.Status = PlanStatusFailed
				result.CompletedAt = time.Now()
				result.Duration = time.Since(startTime)
				result.Error = fmt.Errorf("some tasks failed")
				result.Metrics = e.buildMetrics(execCtx)
				return result, fmt.Errorf("plan execution failed: some tasks failed")
			}

			// No executable tasks but not done - might be dependency deadlock
			break
		}

		// Execute the task
		if err := e.executeTask(ctx, task, execCtx); err != nil {
			// Task failed, but continue to see if other tasks can proceed
			continue
		}

		// Execute subtasks if any
		if err := e.executeSubtasks(ctx, task, execCtx); err != nil {
			// Subtask failed, but continue execution
			continue
		}
	}

	// Execution loop ended without completion
	result.Status = PlanStatusFailed
	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime)
	result.Error = fmt.Errorf("execution did not complete within iteration limit")
	result.Metrics = e.buildMetrics(execCtx)
	return result, fmt.Errorf("execution failed to complete")
}

// buildMetrics constructs execution metrics from the execution context.
func (e *Executor) buildMetrics(execCtx *executionContext) PlanMetrics {
	execCtx.mu.RLock()
	defer execCtx.mu.RUnlock()

	totalDuration := time.Since(execCtx.startTime)

	var completed, failed, skipped int
	var totalTaskDuration time.Duration

	for _, result := range execCtx.results {
		switch result.Status {
		case TaskStatusCompleted:
			completed++
			totalTaskDuration += result.Duration
		case TaskStatusFailed:
			failed++
		case TaskStatusSkipped:
			skipped++
		}
	}

	avgDuration := time.Duration(0)
	if completed > 0 {
		avgDuration = totalTaskDuration / time.Duration(completed)
	}

	successRate := 0.0
	totalExecuted := completed + failed + skipped
	if totalExecuted > 0 {
		successRate = float64(completed) / float64(totalExecuted)
	}

	return PlanMetrics{
		TaskCount:       len(execCtx.plan.Tasks),
		ExecutionTime:   totalDuration,
		AvgTaskDuration: avgDuration,
		SuccessRate:     successRate,
	}
}
