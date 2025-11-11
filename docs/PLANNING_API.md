# Planning Layer API Reference

Complete API documentation for the Planning Layer in go-deep-agent. This reference covers all types, methods, and configuration options.

## Table of Contents

- [Quick Start](#quick-start)
- [Core Types](#core-types)
- [Configuration](#configuration)
- [Decomposer API](#decomposer-api)
- [Executor API](#executor-api)
- [Agent Integration](#agent-integration)
- [Examples](#examples)

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // 1. Create agent
    config := agent.Config{
        Provider: "openai",
        Model:    "gpt-4",
        APIKey:   "your-api-key",
    }
    ag, _ := agent.NewAgent(config)

    // 2. Use high-level API
    result, _ := ag.PlanAndExecute(context.Background(), "Research AI trends")

    // 3. Access results
    println(result.Status)  // "completed"
    println(len(result.Timeline))  // Event count
}
```

### Advanced Usage

```go
// 1. Create custom plan
plan := agent.NewPlan("Custom goal", agent.StrategyParallel)
plan.AddTask(agent.Task{
    ID: "task-1",
    Description: "Do something",
})

// 2. Configure executor
config := agent.DefaultPlannerConfig()
config.MaxParallel = 10
config.Strategy = agent.StrategyAdaptive

// 3. Execute
executor := agent.NewExecutor(config, ag)
result, _ := executor.Execute(context.Background(), plan)
```

## Core Types

### Plan

Represents a complete execution plan.

```go
type Plan struct {
    ID           string
    Goal         string
    Strategy     PlanningStrategy
    Tasks        []Task
    GoalState    GoalState
    Metadata     map[string]interface{}
    CreatedAt    time.Time
}
```

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique identifier (auto-generated) |
| `Goal` | `string` | High-level objective description |
| `Strategy` | `PlanningStrategy` | Execution strategy (Sequential/Parallel/Adaptive) |
| `Tasks` | `[]Task` | List of tasks to execute |
| `GoalState` | `GoalState` | Success criteria (optional) |
| `Metadata` | `map[string]interface{}` | Additional context data |
| `CreatedAt` | `time.Time` | Creation timestamp |

**Constructor**:

```go
func NewPlan(goal string, strategy PlanningStrategy) *Plan
```

Creates a new plan with auto-generated ID and timestamp.

**Example**:
```go
plan := agent.NewPlan(
    "Analyze market trends",
    agent.StrategySequential,
)
```

**Methods**:

```go
// Add a task to the plan
func (p *Plan) AddTask(task Task)

// Get task by ID
func (p *Plan) GetTask(id string) *Task

// Check if all tasks are completed
func (p *Plan) IsComplete() bool

// Get pending tasks (no unmet dependencies)
func (p *Plan) GetReadyTasks() []Task
```

### Task

Represents an atomic unit of work.

```go
type Task struct {
    ID           string
    Description  string
    Type         TaskType
    Status       TaskStatus
    Dependencies []string
    Subtasks     []Task
    Result       interface{}
    Error        error
    StartTime    time.Time
    EndTime      time.Time
    Metadata     map[string]interface{}
}
```

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique task identifier |
| `Description` | `string` | Human-readable task description |
| `Type` | `TaskType` | Task category (Observation/Action/Decision/Aggregate) |
| `Status` | `TaskStatus` | Current state (Pending/Running/Completed/Failed) |
| `Dependencies` | `[]string` | IDs of tasks that must complete first |
| `Subtasks` | `[]Task` | Nested child tasks (optional) |
| `Result` | `interface{}` | Execution result (set after completion) |
| `Error` | `error` | Error if task failed |
| `StartTime` | `time.Time` | Execution start timestamp |
| `EndTime` | `time.Time` | Execution end timestamp |
| `Metadata` | `map[string]interface{}` | Custom task data |

**TaskType Constants**:

```go
const (
    TaskTypeObservation TaskType = "observation"  // Information gathering
    TaskTypeAction      TaskType = "action"       // Execution/modification
    TaskTypeDecision    TaskType = "decision"     // Choice-making
    TaskTypeAggregate   TaskType = "aggregate"    // Result combination
)
```

**TaskStatus Constants**:

```go
const (
    TaskStatusPending   TaskStatus = "pending"    // Not started
    TaskStatusRunning   TaskStatus = "running"    // Currently executing
    TaskStatusCompleted TaskStatus = "completed"  // Successfully finished
    TaskStatusFailed    TaskStatus = "failed"     // Encountered error
)
```

**Example**:
```go
task := agent.Task{
    ID:          "research-ai",
    Description: "Research AI market trends for 2024",
    Type:        agent.TaskTypeObservation,
    Dependencies: []string{"setup-tools"},
    Metadata: map[string]interface{}{
        "source": "web",
        "timeout": 30,
    },
}
```

### GoalState

Defines success criteria for a plan.

```go
type GoalState struct {
    Description string
    Criteria    []GoalCriterion
    MetAt       time.Time
}
```

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `Description` | `string` | Human-readable goal description |
| `Criteria` | `[]GoalCriterion` | List of success conditions |
| `MetAt` | `time.Time` | When goal was achieved (if applicable) |

**GoalCriterion**:

```go
type GoalCriterion struct {
    Name      string
    Expected  string
    Operator  string
    Satisfied bool
}
```

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `Name` | `string` | Criterion identifier | `"items_found"` |
| `Expected` | `string` | Expected value | `"10"` |
| `Operator` | `string` | Comparison operator | `">="`, `"=="`, `"!="` |
| `Satisfied` | `bool` | Whether criterion is met | `false` |

**Example**:
```go
goalState := agent.GoalState{
    Description: "Find 5 relevant research papers",
    Criteria: []agent.GoalCriterion{
        {
            Name:      "papers_count",
            Expected:  "5",
            Operator:  ">=",
            Satisfied: false,
        },
        {
            Name:      "avg_relevance",
            Expected:  "0.8",
            Operator:  ">=",
            Satisfied: false,
        },
    },
}
```

### PlanResult

Contains execution results and metrics.

```go
type PlanResult struct {
    PlanID        string
    Status        PlanStatus
    CompletedAt   time.Time
    Metrics       PlanMetrics
    Timeline      []PlanEvent
    Error         error
}
```

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `PlanID` | `string` | Plan identifier |
| `Status` | `PlanStatus` | Final status (Completed/Failed/Cancelled) |
| `CompletedAt` | `time.Time` | Execution end time |
| `Metrics` | `PlanMetrics` | Performance statistics |
| `Timeline` | `[]PlanEvent` | Execution event log |
| `Error` | `error` | Error if plan failed |

**PlanStatus Constants**:

```go
const (
    PlanStatusPending   PlanStatus = "pending"
    PlanStatusRunning   PlanStatus = "running"
    PlanStatusCompleted PlanStatus = "completed"
    PlanStatusFailed    PlanStatus = "failed"
    PlanStatusCancelled PlanStatus = "cancelled"
)
```

### PlanMetrics

Performance statistics for a plan execution.

```go
type PlanMetrics struct {
    TaskCount        int
    FailedTaskCount  int
    SuccessRate      float64
    ExecutionTime    time.Duration
    AvgTaskDuration  time.Duration
    Strategy         PlanningStrategy
}
```

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `TaskCount` | `int` | Total tasks attempted |
| `FailedTaskCount` | `int` | Number of failed tasks |
| `SuccessRate` | `float64` | Percentage successful (0.0-1.0) |
| `ExecutionTime` | `time.Duration` | Total plan duration |
| `AvgTaskDuration` | `time.Duration` | Mean task execution time |
| `Strategy` | `PlanningStrategy` | Strategy used |

**Example**:
```go
metrics := result.Metrics
fmt.Printf("Success: %.1f%% (%d/%d tasks)\n",
    metrics.SuccessRate*100,
    metrics.TaskCount-metrics.FailedTaskCount,
    metrics.TaskCount,
)
fmt.Printf("Duration: %v (avg: %v per task)\n",
    metrics.ExecutionTime,
    metrics.AvgTaskDuration,
)
```

### PlanEvent

Timeline event during plan execution.

```go
type PlanEvent struct {
    Type        string
    Description string
    Timestamp   time.Time
    Metadata    map[string]interface{}
}
```

**Event Types**:

| Type | Description | When |
|------|-------------|------|
| `task_started` | Task begins execution | Task execution start |
| `task_completed` | Task finishes successfully | Task completes |
| `task_failed` | Task encounters error | Task fails |
| `goal_checked` | Goal criteria evaluated | Periodic check |
| `goal_achieved` | Goal met, plan stopping early | Goal satisfied |
| `strategy_initialized` | Strategy selected | Plan start |
| `strategy_switched` | Strategy changed | Adaptive mode |

**Example**:
```go
for _, event := range result.Timeline {
    fmt.Printf("[%v] %s: %s\n",
        event.Timestamp.Format("15:04:05"),
        event.Type,
        event.Description,
    )
}
```

## Configuration

### PlannerConfig

Configuration for plan execution.

```go
type PlannerConfig struct {
    Strategy          PlanningStrategy
    MaxDepth          int
    MaxSubtasks       int
    MaxParallel       int
    AdaptiveThreshold float64
    GoalCheckInterval int
    Timeout           time.Duration
}
```

**Fields**:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Strategy` | `PlanningStrategy` | `StrategySequential` | Execution strategy |
| `MaxDepth` | `int` | `3` | Maximum subtask nesting levels |
| `MaxSubtasks` | `int` | `10` | Maximum subtasks per task |
| `MaxParallel` | `int` | `5` | Max concurrent tasks (parallel mode) |
| `AdaptiveThreshold` | `float64` | `0.5` | Efficiency threshold for switching (0.0-1.0) |
| `GoalCheckInterval` | `int` | `0` | Check goal every N tasks (0 = at end only) |
| `Timeout` | `time.Duration` | `0` | Max execution time (0 = no limit) |

**PlanningStrategy Constants**:

```go
const (
    StrategySequential PlanningStrategy = "sequential"
    StrategyParallel   PlanningStrategy = "parallel"
    StrategyAdaptive   PlanningStrategy = "adaptive"
)
```

**Default Configuration**:

```go
func DefaultPlannerConfig() PlannerConfig {
    return PlannerConfig{
        Strategy:          StrategySequential,
        MaxDepth:          3,
        MaxSubtasks:       10,
        MaxParallel:       5,
        AdaptiveThreshold: 0.5,
        GoalCheckInterval: 0,
        Timeout:           0,
    }
}
```

**Example**:
```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategyParallel
config.MaxParallel = 10
config.GoalCheckInterval = 5
config.Timeout = 5 * time.Minute
```

## Decomposer API

### NewDecomposer

Creates a new decomposer instance.

```go
func NewDecomposer(config PlannerConfig, llm llmGenerator) *Decomposer
```

**Parameters**:
- `config`: Configuration for decomposition
- `llm`: LLM generator for task breakdown

**Returns**: Configured `*Decomposer`

**Example**:
```go
decomposer := agent.NewDecomposer(
    agent.DefaultPlannerConfig(),
    llmGenerator,
)
```

### Decompose

Breaks down a goal into a plan.

```go
func (d *Decomposer) Decompose(ctx context.Context, goal string) (*Plan, error)
```

**Parameters**:
- `ctx`: Context for cancellation
- `goal`: High-level objective description

**Returns**:
- `*Plan`: Generated execution plan
- `error`: Error if decomposition fails

**Process**:
1. Analyze goal complexity (1-10 scale)
2. Generate task tree with LLM
3. Parse and validate structure
4. Check for dependency cycles
5. Return validated plan

**Example**:
```go
plan, err := decomposer.Decompose(
    context.Background(),
    "Research AI trends and write a report",
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Generated %d tasks\n", len(plan.Tasks))
for _, task := range plan.Tasks {
    fmt.Printf("- %s: %s\n", task.ID, task.Description)
}
```

**Error Cases**:
- Invalid LLM response format
- Dependency cycle detected
- Task count exceeds MaxSubtasks
- Nesting depth exceeds MaxDepth
- Context cancellation

## Executor API

### NewExecutor

Creates a new executor instance.

```go
func NewExecutor(config PlannerConfig, agent agentExecutor) *Executor
```

**Parameters**:
- `config`: Execution configuration
- `agent`: Agent instance for task execution

**Returns**: Configured `*Executor`

**Example**:
```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategyParallel
config.MaxParallel = 10

executor := agent.NewExecutor(config, agentInstance)
```

### Execute

Executes a plan using the configured strategy.

```go
func (e *Executor) Execute(ctx context.Context, plan *Plan) (*PlanResult, error)
```

**Parameters**:
- `ctx`: Context for cancellation and timeouts
- `plan`: Plan to execute

**Returns**:
- `*PlanResult`: Execution results and metrics
- `error`: Error if execution fails

**Process**:
1. Initialize execution context
2. Select strategy (or start with Sequential in Adaptive)
3. Execute tasks according to strategy
4. Monitor performance (Adaptive mode)
5. Check goals periodically (if configured)
6. Build result with metrics and timeline

**Example**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := executor.Execute(ctx, plan)
if err != nil {
    log.Fatalf("Execution failed: %v", err)
}

fmt.Printf("Status: %s\n", result.Status)
fmt.Printf("Tasks: %d/%d successful\n",
    result.Metrics.TaskCount-result.Metrics.FailedTaskCount,
    result.Metrics.TaskCount,
)
```

**Error Cases**:
- Context cancellation
- Timeout exceeded
- Critical task failure
- Dependency cycle detected
- Agent execution error

### Strategy-Specific Methods

#### ExecuteSequential

```go
func (e *Executor) ExecuteSequential(ctx context.Context, plan *Plan) (*PlanResult, error)
```

Executes tasks one at a time in dependency order.

#### ExecuteParallel

```go
func (e *Executor) ExecuteParallel(ctx context.Context, plan *Plan) (*PlanResult, error)
```

Executes independent tasks concurrently, respecting MaxParallel limit.

#### ExecuteAdaptive

```go
func (e *Executor) ExecuteAdaptive(ctx context.Context, plan *Plan) (*PlanResult, error)
```

Dynamically switches between Sequential and Parallel based on performance.

**Note**: These are typically called internally by `Execute()` based on `config.Strategy`. Direct use is for advanced scenarios.

## Agent Integration

### PlanAndExecute

High-level API for planning and execution.

```go
func (a *Agent) PlanAndExecute(ctx context.Context, goal string) (*PlanResult, error)
```

**Parameters**:
- `ctx`: Context for cancellation
- `goal`: High-level objective

**Returns**:
- `*PlanResult`: Execution results
- `error`: Error if planning or execution fails

**Process**:
1. Decompose goal into plan
2. Execute plan with default config
3. Return results

**Example**:
```go
result, err := agent.PlanAndExecute(
    context.Background(),
    "Analyze top 5 tech companies and create comparison report",
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Completed in %v\n", result.Metrics.ExecutionTime)
```

**Customization**:

For more control, use decomposer and executor separately:

```go
// Custom decomposition
decomposer := agent.NewDecomposer(customConfig, llm)
plan, _ := decomposer.Decompose(ctx, goal)

// Modify plan
plan.AddTask(agent.Task{
    ID: "custom-task",
    Description: "Additional step",
})

// Custom execution
executor := agent.NewExecutor(executionConfig, agentInstance)
result, _ := executor.Execute(ctx, plan)
```

## Examples

### Example 1: Sequential Workflow

```go
package main

import (
    "context"
    "fmt"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Setup
    ag, _ := agent.NewAgent(agent.Config{
        Provider: "openai",
        Model:    "gpt-4",
        APIKey:   "sk-...",
    })

    // Create plan
    plan := agent.NewPlan("Setup → Configure → Execute", agent.StrategySequential)
    plan.AddTask(agent.Task{ID: "1", Description: "Setup environment"})
    plan.AddTask(agent.Task{ID: "2", Description: "Configure settings", Dependencies: []string{"1"}})
    plan.AddTask(agent.Task{ID: "3", Description: "Execute main task", Dependencies: []string{"2"}})

    // Execute
    executor := agent.NewExecutor(agent.DefaultPlannerConfig(), ag)
    result, _ := executor.Execute(context.Background(), plan)

    fmt.Printf("Completed %d tasks in %v\n",
        result.Metrics.TaskCount,
        result.Metrics.ExecutionTime,
    )
}
```

### Example 2: Parallel Batch Processing

```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategyParallel
config.MaxParallel = 5

plan := agent.NewPlan("Process 10 items", agent.StrategyParallel)

// Add 10 independent tasks
for i := 1; i <= 10; i++ {
    plan.AddTask(agent.Task{
        ID:          fmt.Sprintf("item-%d", i),
        Description: fmt.Sprintf("Process item %d", i),
        Type:        agent.TaskTypeAction,
    })
}

executor := agent.NewExecutor(config, ag)
result, _ := executor.Execute(context.Background(), plan)

// Expected: ~2x faster than sequential (with MaxParallel=5)
```

### Example 3: Adaptive Strategy

```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategyAdaptive
config.AdaptiveThreshold = 0.6
config.MaxParallel = 5

plan := agent.NewPlan("Mixed workload", agent.StrategyAdaptive)

// Phase 1: Many simple tasks (will trigger parallel)
for i := 1; i <= 10; i++ {
    plan.AddTask(agent.Task{
        ID:          fmt.Sprintf("gather-%d", i),
        Description: "Gather data",
    })
}

// Phase 2: Complex analysis (may switch to sequential)
plan.AddTask(agent.Task{
    ID:          "analyze",
    Description: "Complex analysis of all data",
    Dependencies: gatherTaskIDs,
})

executor := agent.NewExecutor(config, ag)
result, _ := executor.Execute(context.Background(), plan)

// Check strategy switches
for _, event := range result.Timeline {
    if event.Type == "strategy_switched" {
        fmt.Println(event.Description)
    }
}
```

### Example 4: Goal-Oriented Execution

```go
plan := agent.NewPlan("Find papers", agent.StrategySequential)
plan.GoalState = agent.GoalState{
    Description: "Find 3 relevant papers",
    Criteria: []agent.GoalCriterion{
        {Name: "count", Expected: "3", Operator: ">=", Satisfied: false},
    },
}

// Add 20 search tasks
for i := 1; i <= 20; i++ {
    plan.AddTask(agent.Task{
        ID:          fmt.Sprintf("search-%d", i),
        Description: "Search database " + fmt.Sprint(i),
    })
}

config := agent.DefaultPlannerConfig()
config.GoalCheckInterval = 5  // Check every 5 tasks

executor := agent.NewExecutor(config, ag)
result, _ := executor.Execute(context.Background(), plan)

// Will stop early when 3 papers found (likely after ~10-15 tasks)
fmt.Printf("Stopped after %d tasks (goal met)\n", result.Metrics.TaskCount)
```

### Example 5: Error Handling

```go
plan := agent.NewPlan("Robust execution", agent.StrategyParallel)
// Add tasks...

executor := agent.NewExecutor(config, ag)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

result, err := executor.Execute(ctx, plan)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("Execution timeout")
    } else {
        fmt.Printf("Execution error: %v\n", err)
    }
}

// Check individual task failures
for _, task := range plan.Tasks {
    if task.Status == agent.TaskStatusFailed {
        fmt.Printf("Task %s failed: %v\n", task.ID, task.Error)
    }
}

// Metrics still available even on partial failure
fmt.Printf("Success rate: %.1f%%\n", result.Metrics.SuccessRate*100)
```

## Version

API Reference for Planning Layer v0.7.1 (November 2025).

For guides and patterns, see [PLANNING_GUIDE.md](PLANNING_GUIDE.md).
For performance tuning, see [PLANNING_PERFORMANCE.md](PLANNING_PERFORMANCE.md).
