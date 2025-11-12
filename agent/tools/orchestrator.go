package tools

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Orchestrator coordinates parallel tool execution with dependency management,
// worker pools, timeouts, and result aggregation.
//
// Features:
// - Automatic parallel execution of independent tools
// - Dependency detection and sequential execution when needed
// - Configurable worker pool size
// - Per-tool timeout enforcement
// - Context cancellation support
// - Error aggregation
// - Result ordering preservation
type Orchestrator struct {
	maxWorkers     int           // Maximum concurrent workers
	toolTimeout    time.Duration // Default timeout per tool
	enableParallel bool          // Enable parallel execution
	mu             sync.RWMutex  // Protects configuration
}

// OrchestratorConfig holds configuration for the orchestrator.
type OrchestratorConfig struct {
	MaxWorkers     int           // Max concurrent workers (default: 10)
	ToolTimeout    time.Duration // Timeout per tool (default: 30s)
	EnableParallel bool          // Enable parallel execution (default: true)
}

// DefaultOrchestratorConfig returns sensible defaults.
func DefaultOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{
		MaxWorkers:     10,
		ToolTimeout:    30 * time.Second,
		EnableParallel: true,
	}
}

// NewOrchestrator creates a new tool orchestrator with default config.
func NewOrchestrator() *Orchestrator {
	config := DefaultOrchestratorConfig()
	return &Orchestrator{
		maxWorkers:     config.MaxWorkers,
		toolTimeout:    config.ToolTimeout,
		enableParallel: config.EnableParallel,
	}
}

// NewOrchestratorWithConfig creates orchestrator with custom configuration.
func NewOrchestratorWithConfig(config OrchestratorConfig) *Orchestrator {
	return &Orchestrator{
		maxWorkers:     config.MaxWorkers,
		toolTimeout:    config.ToolTimeout,
		enableParallel: config.EnableParallel,
	}
}

// SetMaxWorkers configures the maximum number of concurrent workers.
func (o *Orchestrator) SetMaxWorkers(max int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if max < 1 {
		max = 1
	}
	o.maxWorkers = max
}

// SetToolTimeout sets the default timeout for tool execution.
func (o *Orchestrator) SetToolTimeout(timeout time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.toolTimeout = timeout
}

// SetParallelExecution enables or disables parallel execution.
func (o *Orchestrator) SetParallelExecution(enable bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.enableParallel = enable
}

// ToolCall represents a tool to be executed with its metadata.
type ToolCall struct {
	ID        string                       // Unique identifier
	Name      string                       // Tool name
	Args      string                       // JSON arguments
	Handler   func(string) (string, error) // Tool handler function
	Timeout   time.Duration                // Execution timeout (0 = use default)
	DependsOn []string                     // IDs of tools this depends on
}

// ToolResult contains the result of a tool execution.
type ToolResult struct {
	ID        string        // Tool call ID
	Name      string        // Tool name
	Result    string        // Success result
	Error     error         // Error if execution failed
	Duration  time.Duration // Execution duration
	StartTime time.Time     // When execution started
	EndTime   time.Time     // When execution finished
}

// ExecutionPlan represents a plan for executing tools with dependencies.
type ExecutionPlan struct {
	Batches [][]string          // Tool IDs grouped by execution batch
	Graph   map[string][]string // Dependency graph (tool ID -> depends on IDs)
}

// Execute runs multiple tools with automatic parallelization.
// Tools without dependencies run in parallel, respecting worker pool limits.
// Tools with dependencies run sequentially after their dependencies complete.
//
// Returns results in the same order as input toolCalls (for LLM context preservation).
func (o *Orchestrator) Execute(ctx context.Context, toolCalls []*ToolCall) ([]*ToolResult, error) {
	if len(toolCalls) == 0 {
		return nil, nil
	}

	// Check if parallel execution is enabled
	o.mu.RLock()
	parallel := o.enableParallel
	o.mu.RUnlock()

	if !parallel || len(toolCalls) == 1 {
		// Sequential execution
		return o.executeSequential(ctx, toolCalls)
	}

	// Build execution plan based on dependencies
	plan := o.buildExecutionPlan(toolCalls)

	// Execute plan in batches
	return o.executePlan(ctx, toolCalls, plan)
}

// executeSequential runs tools one by one (used when parallel is disabled).
func (o *Orchestrator) executeSequential(ctx context.Context, toolCalls []*ToolCall) ([]*ToolResult, error) {
	results := make([]*ToolResult, len(toolCalls))

	for i, tc := range toolCalls {
		result := o.executeOne(ctx, tc)
		results[i] = result

		// Stop on first error if context is canceled
		if result.Error != nil && ctx.Err() != nil {
			return results, ctx.Err()
		}
	}

	return results, nil
}

// buildExecutionPlan analyzes dependencies and creates execution batches.
// Tools in the same batch can run in parallel.
func (o *Orchestrator) buildExecutionPlan(toolCalls []*ToolCall) *ExecutionPlan {
	// Build dependency graph
	graph := make(map[string][]string)
	idToIndex := make(map[string]int)

	for i, tc := range toolCalls {
		idToIndex[tc.ID] = i
		if len(tc.DependsOn) > 0 {
			graph[tc.ID] = tc.DependsOn
		}
	}

	// Group tools into batches using topological sort
	batches := [][]string{}
	executed := make(map[string]bool)

	for len(executed) < len(toolCalls) {
		batch := []string{}

		for _, tc := range toolCalls {
			if executed[tc.ID] {
				continue
			}

			// Check if all dependencies are executed
			canExecute := true
			for _, depID := range tc.DependsOn {
				if !executed[depID] {
					canExecute = false
					break
				}
			}

			if canExecute {
				batch = append(batch, tc.ID)
			}
		}

		if len(batch) == 0 {
			// Circular dependency detected or invalid dependency
			// Add remaining tools to final batch
			for _, tc := range toolCalls {
				if !executed[tc.ID] {
					batch = append(batch, tc.ID)
				}
			}
		}

		batches = append(batches, batch)
		for _, id := range batch {
			executed[id] = true
		}
	}

	return &ExecutionPlan{
		Batches: batches,
		Graph:   graph,
	}
}

// executePlan executes tools according to the plan, respecting dependencies.
func (o *Orchestrator) executePlan(ctx context.Context, toolCalls []*ToolCall, plan *ExecutionPlan) ([]*ToolResult, error) {
	// Map tool IDs to ToolCall objects
	idToTool := make(map[string]*ToolCall)
	for _, tc := range toolCalls {
		idToTool[tc.ID] = tc
	}

	// Map tool IDs to results
	idToResult := make(map[string]*ToolResult)

	// Execute each batch in sequence, but tools within batch in parallel
	for batchNum, batch := range plan.Batches {
		batchResults := o.executeBatch(ctx, batch, idToTool)

		// Store results
		for id, result := range batchResults {
			idToResult[id] = result
		}

		// Check for context cancellation
		if ctx.Err() != nil {
			break
		}

		// Log batch completion
		logInfo(ctx, "Tool batch completed", map[string]interface{}{
			"batch":      batchNum + 1,
			"total":      len(plan.Batches),
			"tool_count": len(batch),
		})
	}

	// Return results in original order
	results := make([]*ToolResult, len(toolCalls))
	for i, tc := range toolCalls {
		if result, ok := idToResult[tc.ID]; ok {
			results[i] = result
		} else {
			// Tool was not executed (likely due to error or cancellation)
			results[i] = &ToolResult{
				ID:    tc.ID,
				Name:  tc.Name,
				Error: fmt.Errorf("tool not executed"),
			}
		}
	}

	return results, nil
}

// executeBatch runs a batch of tools in parallel using worker pool.
func (o *Orchestrator) executeBatch(ctx context.Context, batch []string, idToTool map[string]*ToolCall) map[string]*ToolResult {
	results := make(map[string]*ToolResult)
	resultsMu := sync.Mutex{}

	// Create worker pool
	o.mu.RLock()
	maxWorkers := o.maxWorkers
	o.mu.RUnlock()

	// Limit workers to batch size
	if len(batch) < maxWorkers {
		maxWorkers = len(batch)
	}

	// Create semaphore for worker pool
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, toolID := range batch {
		tc := idToTool[toolID]
		if tc == nil {
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire worker

		go func(call *ToolCall) {
			defer wg.Done()
			defer func() { <-sem }() // Release worker

			result := o.executeOne(ctx, call)

			resultsMu.Lock()
			results[call.ID] = result
			resultsMu.Unlock()
		}(tc)
	}

	wg.Wait()
	return results
}

// executeOne executes a single tool with timeout and error handling.
func (o *Orchestrator) executeOne(ctx context.Context, tc *ToolCall) *ToolResult {
	result := &ToolResult{
		ID:        tc.ID,
		Name:      tc.Name,
		StartTime: time.Now(),
	}

	// Determine timeout
	timeout := tc.Timeout
	if timeout == 0 {
		o.mu.RLock()
		timeout = o.toolTimeout
		o.mu.RUnlock()
	}

	// Create context with timeout
	toolCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute in goroutine to support timeout
	done := make(chan struct{})
	var output string
	var err error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("tool panicked: %v", r)
			}
			close(done)
		}()

		output, err = tc.Handler(tc.Args)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		result.Result = output
		result.Error = err
	case <-toolCtx.Done():
		result.Error = fmt.Errorf("tool execution timeout after %v", timeout)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Log execution
	if result.Error != nil {
		logError(toolCtx, "Tool execution failed", map[string]interface{}{
			"tool_id":  tc.ID,
			"tool":     tc.Name,
			"duration": result.Duration,
			"error":    result.Error.Error(),
		})
	} else {
		logDebug(toolCtx, "Tool executed successfully", map[string]interface{}{
			"tool_id":  tc.ID,
			"tool":     tc.Name,
			"duration": result.Duration,
		})
	}

	return result
}

// Stats returns statistics about an execution.
type ExecutionStats struct {
	TotalTools      int           // Total number of tools
	SuccessCount    int           // Successfully executed
	FailureCount    int           // Failed executions
	TotalDuration   time.Duration // Total execution time
	ParallelBatches int           // Number of parallel batches
	AvgDuration     time.Duration // Average tool duration
	MaxDuration     time.Duration // Longest tool duration
}

// ComputeStats calculates statistics from execution results.
func ComputeStats(results []*ToolResult, plan *ExecutionPlan) *ExecutionStats {
	stats := &ExecutionStats{
		TotalTools:      len(results),
		ParallelBatches: len(plan.Batches),
	}

	for _, r := range results {
		if r.Error == nil {
			stats.SuccessCount++
		} else {
			stats.FailureCount++
		}

		stats.TotalDuration += r.Duration
		if r.Duration > stats.MaxDuration {
			stats.MaxDuration = r.Duration
		}
	}

	if stats.TotalTools > 0 {
		stats.AvgDuration = stats.TotalDuration / time.Duration(stats.TotalTools)
	}

	return stats
}
