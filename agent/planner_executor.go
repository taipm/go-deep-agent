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

	// Performance tracking for adaptive strategy
	perf performanceTracker

	// Timeline tracking for observability
	timeline      []PlanEvent
	timelineMu    sync.Mutex
	tasksExecuted int // Counter for periodic goal checking
}

// performanceTracker monitors execution performance for adaptive strategy switching.
type performanceTracker struct {
	mu                  sync.RWMutex
	batchStartTime      time.Time
	batchTaskCount      int
	batchCompletedCount int
	totalTaskDuration   time.Duration
	currentStrategy     PlanningStrategy
}

// startBatch initializes performance tracking for a new batch.
func (p *performanceTracker) startBatch(strategy PlanningStrategy, taskCount int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.batchStartTime = time.Now()
	p.batchTaskCount = taskCount
	p.batchCompletedCount = 0
	p.totalTaskDuration = 0
	p.currentStrategy = strategy
}

// recordTaskCompletion records a completed task for performance metrics.
func (p *performanceTracker) recordTaskCompletion(duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.batchCompletedCount++
	p.totalTaskDuration += duration
}

// getMetrics returns current performance metrics.
func (p *performanceTracker) getMetrics() PerformanceMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	batchDuration := time.Since(p.batchStartTime)

	// Tasks per second
	tasksPerSec := 0.0
	if batchDuration > 0 {
		tasksPerSec = float64(p.batchCompletedCount) / batchDuration.Seconds()
	}

	// Average latency per task
	avgLatency := time.Duration(0)
	if p.batchCompletedCount > 0 {
		avgLatency = p.totalTaskDuration / time.Duration(p.batchCompletedCount)
	}

	// Parallel efficiency: ratio of actual time vs ideal parallel time
	// For sequential: efficiency = 1.0 (baseline)
	// For parallel: efficiency = (sum of task durations) / (batch wall time * task count)
	parallelEfficiency := 1.0
	if p.currentStrategy == StrategyParallel && batchDuration > 0 && p.batchTaskCount > 0 {
		idealTime := p.totalTaskDuration
		actualTime := batchDuration * time.Duration(p.batchTaskCount)
		if actualTime > 0 {
			parallelEfficiency = float64(idealTime) / float64(actualTime)
		}
	}

	return PerformanceMetrics{
		TasksPerSec:        tasksPerSec,
		AvgLatency:         avgLatency,
		ParallelEfficiency: parallelEfficiency,
		Strategy:           p.currentStrategy,
		CompletedTasks:     p.batchCompletedCount,
		TotalTasks:         p.batchTaskCount,
	}
}

// PerformanceMetrics captures runtime performance data for adaptive strategy.
type PerformanceMetrics struct {
	// TasksPerSec is the throughput (tasks completed per second).
	TasksPerSec float64

	// AvgLatency is the average time per task.
	AvgLatency time.Duration

	// ParallelEfficiency is the ratio of ideal parallel time to actual time (0.0 to 1.0+).
	// > 0.7: Parallel is efficient
	// < 0.3: Sequential might be better (high overhead)
	ParallelEfficiency float64

	// Strategy is the current execution strategy.
	Strategy PlanningStrategy

	// CompletedTasks is the number of tasks completed in current batch.
	CompletedTasks int

	// TotalTasks is the total tasks in current batch.
	TotalTasks int
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
		plan:          plan,
		results:       make(map[string]*TaskResult),
		startTime:     time.Now(),
		timeline:      []PlanEvent{},
		tasksExecuted: 0,
	}
}

// addEvent records a timeline event for observability.
func (ctx *executionContext) addEvent(eventType, taskID, description string) {
	ctx.timelineMu.Lock()
	defer ctx.timelineMu.Unlock()

	ctx.timeline = append(ctx.timeline, PlanEvent{
		Timestamp:   time.Now(),
		Type:        eventType,
		TaskID:      taskID,
		Description: description,
	})
}

// incrementTaskCounter increments the task execution counter.
func (ctx *executionContext) incrementTaskCounter() int {
	ctx.timelineMu.Lock()
	defer ctx.timelineMu.Unlock()

	ctx.tasksExecuted++
	return ctx.tasksExecuted
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

// topologicalSort performs topological sorting of tasks using Kahn's algorithm.
// Returns ordered task slice or error if cycle detected.
func (e *Executor) topologicalSort(tasks []Task) ([]Task, error) {
	// Build adjacency list and in-degree map
	taskMap := make(map[string]*Task)
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// Initialize
	for i := range tasks {
		task := &tasks[i]
		taskMap[task.ID] = task
		inDegree[task.ID] = 0
		adjList[task.ID] = []string{}
	}

	// Build graph
	for i := range tasks {
		task := &tasks[i]
		for _, depID := range task.Dependencies {
			// Dependency: depID -> task.ID
			adjList[depID] = append(adjList[depID], task.ID)
			inDegree[task.ID]++
		}
	}

	// Find all nodes with in-degree 0
	queue := []string{}
	for taskID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, taskID)
		}
	}

	// Kahn's algorithm
	sorted := []Task{}
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		// Add to sorted list
		sorted = append(sorted, *taskMap[current])

		// Reduce in-degree of neighbors
		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(sorted) != len(tasks) {
		return nil, fmt.Errorf("cycle detected in task dependencies")
	}

	return sorted, nil
}

// groupByDependencyLevel groups tasks into levels for parallel execution.
// Level 0 = no dependencies, Level 1 = depends only on Level 0, etc.
// Returns [][]Task where each inner slice can be executed concurrently.
func (e *Executor) groupByDependencyLevel(tasks []Task) ([][]Task, error) {
	// First perform topological sort to detect cycles
	_, err := e.topologicalSort(tasks)
	if err != nil {
		return nil, err
	}

	// Build task map for quick lookup
	taskMap := make(map[string]*Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Calculate dependency level for each task using BFS
	levels := make(map[string]int)

	// Initialize tasks with no dependencies to level 0
	queue := []string{}
	for _, task := range tasks {
		if len(task.Dependencies) == 0 {
			levels[task.ID] = 0
			queue = append(queue, task.ID)
		}
	}

	// BFS to calculate levels
	processed := make(map[string]bool)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		processed[current] = true

		// Find tasks that depend on current
		for _, task := range tasks {
			if processed[task.ID] {
				continue
			}

			// Check if current is a dependency of task
			isDep := false
			for _, depID := range task.Dependencies {
				if depID == current {
					isDep = true
					break
				}
			}

			if !isDep {
				continue
			}

			// Check if all dependencies of task are processed
			allDepsProcessed := true
			maxDepLevel := 0
			for _, depID := range task.Dependencies {
				if !processed[depID] {
					allDepsProcessed = false
					break
				}
				if levels[depID] > maxDepLevel {
					maxDepLevel = levels[depID]
				}
			}

			if allDepsProcessed {
				levels[task.ID] = maxDepLevel + 1
				queue = append(queue, task.ID)
			}
		}
	}

	// Group tasks by level
	maxLevel := 0
	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	result := make([][]Task, maxLevel+1)
	for _, task := range tasks {
		level := levels[task.ID]
		result[level] = append(result[level], task)
	}

	return result, nil
}

// executeTask executes a single task using the agent's ReAct capabilities.
func (e *Executor) executeTask(ctx context.Context, task *Task, execCtx *executionContext) error {
	startTime := time.Now()

	// Record task start event
	execCtx.addEvent("task_started", task.ID, fmt.Sprintf("Starting task: %s", task.Description))

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

		// Record task failure event
		execCtx.addEvent("task_failed", task.ID, fmt.Sprintf("Task failed: %v", err))

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

	// Record performance metrics for adaptive strategy
	execCtx.perf.recordTaskCompletion(result.Duration)

	// Record task completion event
	execCtx.addEvent("task_completed", task.ID, fmt.Sprintf("Task completed in %v", result.Duration))

	// Increment task counter and check if we should perform goal check
	taskCount := execCtx.incrementTaskCounter()
	if e.config.GoalCheckInterval > 0 && taskCount%e.config.GoalCheckInterval == 0 {
		// Perform periodic goal check
		goalMet, reason := e.checkGoalCompletion(execCtx)
		execCtx.addEvent("goal_checked", "", fmt.Sprintf("Goal check at task %d: %s (met: %v)", taskCount, reason, goalMet))

		if goalMet {
			// Goal achieved early - mark for termination
			execCtx.addEvent("goal_achieved", "", fmt.Sprintf("Goal achieved early after %d tasks", taskCount))
		}
	}

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
				result.Timeline = execCtx.timeline
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
				result.Timeline = execCtx.timeline
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
	result.Timeline = execCtx.timeline
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

// executeParallel executes tasks in parallel by dependency levels.
// Tasks within the same level can run concurrently up to MaxParallel limit.
func (e *Executor) executeParallel(ctx context.Context, plan *Plan) (*PlanResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan cannot be nil")
	}

	// Create execution context
	execCtx := newExecutionContext(plan)
	startTime := time.Now()

	// Create result object
	result := &PlanResult{
		Plan:      *plan,
		Status:    PlanStatusRunning,
		StartedAt: startTime,
	}

	// Group tasks by dependency level
	levels, err := e.groupByDependencyLevel(plan.Tasks)
	if err != nil {
		result.Status = PlanStatusFailed
		result.CompletedAt = time.Now()
		result.Duration = time.Since(startTime)
		result.Error = err
		return result, err
	}

	// Execute level by level
	for levelIdx, levelTasks := range levels {
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

		// Execute all tasks in current level concurrently
		levelErrors := make(chan error, len(levelTasks))
		semaphore := make(chan struct{}, e.config.MaxParallel)

		for i := range levelTasks {
			taskID := levelTasks[i].ID

			// Find task pointer from plan
			var task *Task
			for j := range plan.Tasks {
				if plan.Tasks[j].ID == taskID {
					task = &plan.Tasks[j]
					break
				}
			}

			if task == nil {
				continue
			}

			// Acquire semaphore slot
			semaphore <- struct{}{}

			go func(t *Task) {
				defer func() { <-semaphore }()

				// Execute task
				if err := e.executeTask(ctx, t, execCtx); err != nil {
					levelErrors <- fmt.Errorf("level %d task %s failed: %w", levelIdx, t.ID, err)
					return
				}

				// Execute subtasks sequentially (for simplicity)
				if err := e.executeSubtasks(ctx, t, execCtx); err != nil {
					levelErrors <- fmt.Errorf("level %d task %s subtasks failed: %w", levelIdx, t.ID, err)
					return
				}

				levelErrors <- nil
			}(task)
		}

		// Wait for all tasks in level to complete
		var levelFailed bool
		for i := 0; i < len(levelTasks); i++ {
			if err := <-levelErrors; err != nil {
				levelFailed = true
				// Continue collecting errors instead of early exit
			}
		}

		// If any task in level failed, stop execution
		if levelFailed {
			result.Status = PlanStatusFailed
			result.CompletedAt = time.Now()
			result.Duration = time.Since(startTime)
			result.Error = fmt.Errorf("level %d execution failed", levelIdx)
			result.Metrics = e.buildMetrics(execCtx)
			return result, result.Error
		}
	}

	// Check goal completion
	goalMet, reason := e.checkGoalCompletion(execCtx)
	if goalMet {
		result.Status = PlanStatusCompleted
		result.FinalResult = reason
		result.Metrics = e.buildMetrics(execCtx)
		result.Metrics.GoalAchieved = true
	} else {
		result.Status = PlanStatusFailed
		result.Error = fmt.Errorf("goal not achieved: %s", reason)
	}

	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime)
	return result, nil
}

// shouldSwitchStrategy determines if the execution strategy should change based on performance.
func (e *Executor) shouldSwitchStrategy(metrics PerformanceMetrics, currentStrategy PlanningStrategy) bool {
	// Sequential → Parallel: Switch if we have multiple independent tasks
	if currentStrategy == StrategySequential {
		// Would need to check if remaining tasks have low dependency density
		// For now, this is a placeholder for adaptive logic
		return false
	}

	// Parallel → Sequential: Switch if parallel efficiency is too low
	if currentStrategy == StrategyParallel {
		// If parallel efficiency drops below threshold, sequential might be better
		if metrics.ParallelEfficiency < e.config.AdaptiveThreshold {
			return true
		}
	}

	return false
}

// selectNextStrategy chooses the best strategy based on current performance.
func (e *Executor) selectNextStrategy(metrics PerformanceMetrics, currentStrategy PlanningStrategy, remainingTasks int) PlanningStrategy {
	// If parallel efficiency is low, switch to sequential
	if currentStrategy == StrategyParallel && metrics.ParallelEfficiency < e.config.AdaptiveThreshold {
		return StrategySequential
	}

	// If sequential and we have many independent tasks, consider parallel
	if currentStrategy == StrategySequential && remainingTasks >= e.config.MaxParallel {
		// This would need dependency analysis in full implementation
		// For now, stay sequential unless explicitly parallel
		return StrategySequential
	}

	// Default: keep current strategy
	return currentStrategy
}

// executeAdaptive executes tasks with dynamic strategy switching based on runtime performance.
func (e *Executor) executeAdaptive(ctx context.Context, plan *Plan) (*PlanResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan cannot be nil")
	}

	// Create execution context
	execCtx := newExecutionContext(plan)
	startTime := time.Now()

	// Create result object
	result := &PlanResult{
		Plan:      *plan,
		Status:    PlanStatusRunning,
		StartedAt: startTime,
		Timeline:  []PlanEvent{},
	}

	// Start with sequential strategy (safe default)
	currentStrategy := StrategySequential

	// Record initial strategy
	result.Timeline = append(result.Timeline, PlanEvent{
		Timestamp:   time.Now(),
		Type:        "strategy_initialized",
		Description: fmt.Sprintf("Starting with %s strategy", currentStrategy),
	})

	// Group tasks for batching
	levels, err := e.groupByDependencyLevel(plan.Tasks)
	if err != nil {
		result.Status = PlanStatusFailed
		result.CompletedAt = time.Now()
		result.Duration = time.Since(startTime)
		result.Error = err
		return result, err
	}

	// Execute level by level with adaptive strategy
	for levelIdx, levelTasks := range levels {
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

		// Start performance tracking for this batch
		execCtx.perf.startBatch(currentStrategy, len(levelTasks))

		// Execute batch based on current strategy
		var batchErr error
		if currentStrategy == StrategyParallel && len(levelTasks) > 1 {
			batchErr = e.executeBatchParallel(ctx, levelTasks, plan, execCtx, levelIdx)
		} else {
			batchErr = e.executeBatchSequential(ctx, levelTasks, plan, execCtx)
		}

		if batchErr != nil {
			result.Status = PlanStatusFailed
			result.CompletedAt = time.Now()
			result.Duration = time.Since(startTime)
			result.Error = batchErr
			result.Metrics = e.buildMetrics(execCtx)
			return result, batchErr
		}

		// Get performance metrics for this batch
		perfMetrics := execCtx.perf.getMetrics()

		// Decide if we should switch strategy for next batch
		if e.shouldSwitchStrategy(perfMetrics, currentStrategy) {
			remainingLevels := len(levels) - levelIdx - 1
			if remainingLevels > 0 {
				newStrategy := e.selectNextStrategy(perfMetrics, currentStrategy, remainingLevels)
				if newStrategy != currentStrategy {
					// Record strategy switch
					result.Timeline = append(result.Timeline, PlanEvent{
						Timestamp: time.Now(),
						Type:      "strategy_switched",
						Description: fmt.Sprintf("Switched from %s to %s (efficiency: %.2f, threshold: %.2f)",
							currentStrategy, newStrategy, perfMetrics.ParallelEfficiency, e.config.AdaptiveThreshold),
					})
					currentStrategy = newStrategy
				}
			}
		}
	}

	// Check goal completion
	goalMet, reason := e.checkGoalCompletion(execCtx)
	if goalMet {
		result.Status = PlanStatusCompleted
		result.FinalResult = reason
		result.Metrics = e.buildMetrics(execCtx)
		result.Metrics.GoalAchieved = true
	} else {
		result.Status = PlanStatusFailed
		result.Error = fmt.Errorf("goal not achieved: %s", reason)
	}

	// Merge timeline from execCtx with result timeline
	result.Timeline = append(result.Timeline, execCtx.timeline...)

	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime)
	return result, nil
}

// executeBatchSequential executes a batch of tasks sequentially.
func (e *Executor) executeBatchSequential(ctx context.Context, tasks []Task, plan *Plan, execCtx *executionContext) error {
	for i := range tasks {
		taskID := tasks[i].ID

		// Find task pointer from plan
		var task *Task
		for j := range plan.Tasks {
			if plan.Tasks[j].ID == taskID {
				task = &plan.Tasks[j]
				break
			}
		}

		if task == nil {
			continue
		}

		// Execute task
		if err := e.executeTask(ctx, task, execCtx); err != nil {
			return fmt.Errorf("task %s failed: %w", task.ID, err)
		}

		// Execute subtasks
		if err := e.executeSubtasks(ctx, task, execCtx); err != nil {
			return fmt.Errorf("task %s subtasks failed: %w", task.ID, err)
		}
	}

	return nil
}

// executeBatchParallel executes a batch of tasks in parallel.
func (e *Executor) executeBatchParallel(ctx context.Context, tasks []Task, plan *Plan, execCtx *executionContext, levelIdx int) error {
	levelErrors := make(chan error, len(tasks))
	semaphore := make(chan struct{}, e.config.MaxParallel)

	for i := range tasks {
		taskID := tasks[i].ID

		// Find task pointer from plan
		var task *Task
		for j := range plan.Tasks {
			if plan.Tasks[j].ID == taskID {
				task = &plan.Tasks[j]
				break
			}
		}

		if task == nil {
			continue
		}

		// Acquire semaphore slot
		semaphore <- struct{}{}

		go func(t *Task) {
			defer func() { <-semaphore }()

			// Execute task
			if err := e.executeTask(ctx, t, execCtx); err != nil {
				levelErrors <- fmt.Errorf("level %d task %s failed: %w", levelIdx, t.ID, err)
				return
			}

			// Execute subtasks sequentially
			if err := e.executeSubtasks(ctx, t, execCtx); err != nil {
				levelErrors <- fmt.Errorf("level %d task %s subtasks failed: %w", levelIdx, t.ID, err)
				return
			}

			levelErrors <- nil
		}(task)
	}

	// Wait for all tasks to complete
	var failed bool
	for i := 0; i < len(tasks); i++ {
		if err := <-levelErrors; err != nil {
			failed = true
		}
	}

	if failed {
		return fmt.Errorf("batch execution failed")
	}

	return nil
}
