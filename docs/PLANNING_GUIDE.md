# Planning Layer Guide

The Planning Layer in go-deep-agent enables intelligent goal decomposition and execution through three core components: **Decomposer**, **Executor**, and **Monitoring**. This guide covers concepts, strategies, patterns, and best practices.

## Table of Contents

- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Execution Strategies](#execution-strategies)
- [Task Dependencies](#task-dependencies)
- [Goal-Oriented Planning](#goal-oriented-planning)
- [Performance Optimization](#performance-optimization)
- [Common Patterns](#common-patterns)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

### What is Planning?

Planning transforms a high-level goal into a structured execution plan:

```
Goal: "Research AI market trends and write report"
       ↓
[Decomposer] → LLM analyzes complexity
       ↓
Plan: {
  Task 1: "Gather AI market data"
  Task 2: "Analyze trends" (depends on Task 1)
  Task 3: "Write report" (depends on Task 2)
}
       ↓
[Executor] → Executes with chosen strategy
       ↓
Result: Completed report with metrics
```

### Intelligence Progression

The Planning Layer elevates agent intelligence from reactive to proactive:

| Level | Capability | Pattern | Intelligence |
|-------|-----------|---------|--------------|
| 1.0 | Single-turn chat | Basic Q&A | 1.0 |
| 2.0 | Multi-turn conversation | Context retention | 1.5 |
| 2.5 | Tool usage | ReAct pattern | 2.8 |
| **3.0** | **Sequential planning** | **Decompose → Execute** | **3.2** |
| **3.5** | **Parallel/Adaptive planning** | **Optimized execution** | **3.5** |
| 4.0 | Self-reflection | Plan → Execute → Reflect | 4.0 |
| 5.0 | Meta-learning | Continuous improvement | 5.0 |

**Current Status**: v0.7.1 achieves ~3.5/5.0 with parallel and adaptive strategies.

## Core Concepts

### Plans

A Plan represents the complete execution blueprint:

```go
plan := agent.NewPlan(
    "Research and analyze AI industry",
    agent.StrategySequential,
)
```

**Key Properties**:
- `ID`: Unique identifier
- `Goal`: High-level objective
- `Strategy`: Execution approach (Sequential/Parallel/Adaptive)
- `Tasks`: List of task nodes
- `GoalState`: Success criteria
- `Metadata`: Additional context

### Tasks

Tasks are atomic units of work in a plan:

```go
task := agent.Task{
    ID:           "research-1",
    Description:  "Gather AI market data from 2024",
    Type:         agent.TaskTypeObservation,
    Dependencies: []string{"setup"},
    Subtasks:     []agent.Task{...},
}
```

**Task Types**:
- `TaskTypeObservation`: Information gathering
- `TaskTypeAction`: Execution or modification
- `TaskTypeDecision`: Choice-making based on data
- `TaskTypeAggregate`: Combining multiple results

**Task States**:
- `TaskStatusPending`: Not yet started
- `TaskStatusRunning`: Currently executing
- `TaskStatusCompleted`: Successfully finished
- `TaskStatusFailed`: Encountered error

### Decomposer

The Decomposer uses LLM intelligence to break down goals:

```go
decomposer := agent.NewDecomposer(config, llmGenerator)
plan, err := decomposer.Decompose(ctx, "Build a web scraper")
```

**Decomposition Process**:
1. **Complexity Analysis**: Evaluate goal difficulty (1-10 scale)
2. **LLM Generation**: Create task tree with dependencies
3. **Validation**: Check for cycles, validate structure
4. **Optimization**: Identify parallel opportunities

**Complexity Scoring**:
- 1-3: Simple (single task, < 5 minutes)
- 4-6: Moderate (3-5 tasks, < 30 minutes)
- 7-9: Complex (5-10 tasks, multiple phases)
- 10: Very complex (10+ tasks, hours of work)

### Executor

The Executor runs the plan using the chosen strategy:

```go
executor := agent.NewExecutor(config, agentInstance)
result, err := executor.Execute(ctx, plan)
```

**Executor Responsibilities**:
- Dependency resolution (topological sort)
- Strategy selection and switching
- Performance monitoring
- Timeline tracking
- Error handling

## Execution Strategies

### Sequential Strategy

**When to Use**:
- Tasks have strict ordering requirements
- Tasks are very fast (< 100ms each)
- Strong dependencies between all tasks
- Simple, predictable workflows

**Characteristics**:
- One task at a time
- Deterministic execution order
- Minimal overhead
- Easy to debug

**Example**:
```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategySequential

// Good for:
// 1. Setup → 2. Configure → 3. Execute → 4. Cleanup
```

**Performance**:
- Throughput: 1 task/unit time
- Latency: Sum of all task durations
- Overhead: ~0.1% (minimal)

### Parallel Strategy

**When to Use**:
- Many independent tasks
- I/O-bound workloads (API calls, file ops)
- Consistent task complexity
- Maximum throughput critical

**Characteristics**:
- Multiple concurrent tasks
- Respects dependency levels
- Configurable concurrency (MaxParallel)
- Higher throughput

**Example**:
```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategyParallel
config.MaxParallel = 5

// Good for:
// 1. Fetch data from 10 APIs (parallel)
// 2. Aggregate results (sequential)
```

**Performance**:
- Throughput: Up to MaxParallel × sequential rate
- Latency: Similar to sequential per task
- Overhead: ~2-5% (coordination cost)
- Speedup: 2-8x depending on workload

**Dependency Handling**:
```
Level 0: [A] ──────────────────┐
                                ├──> Level 2: [D]
Level 1: [B, C] (parallel) ────┘

Execution:
- A runs alone (level 0)
- B and C run in parallel (level 1) after A completes
- D runs (level 2) after B and C complete
```

### Adaptive Strategy

**When to Use**:
- Unknown workload characteristics
- Mixed task complexity
- Multi-phase pipelines
- Production systems with varying load

**Characteristics**:
- Starts with sequential
- Monitors performance metrics
- Switches strategies mid-execution
- Optimizes automatically

**Example**:
```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategyAdaptive
config.AdaptiveThreshold = 0.6  // Switch if efficiency < 60%
config.MaxParallel = 5

// Automatically adapts to:
// Phase 1: Many simple tasks → switches to parallel
// Phase 2: Complex analysis → switches to sequential
// Phase 3: Multiple reports → switches to parallel
```

**Performance Metrics**:
- **TasksPerSec**: Throughput measurement
- **AvgLatency**: Per-task overhead
- **ParallelEfficiency**: Parallelization effectiveness

**Switching Logic**:
```go
if currentStrategy == Parallel {
    if efficiency < AdaptiveThreshold {
        // Parallel not effective → switch to Sequential
    }
} else if currentStrategy == Sequential {
    if hasIndependentTasks && expectedEfficiency > threshold {
        // Can benefit from parallelism → switch to Parallel
    }
}
```

**Recommended Thresholds**:
- 0.3-0.4: Aggressive parallelism (I/O-bound)
- 0.5-0.6: Balanced (default, most use cases)
- 0.7-0.8: Conservative (CPU-bound, high overhead)

## Task Dependencies

### Dependency Types

**Direct Dependencies**:
```go
task := agent.Task{
    ID: "analyze",
    Dependencies: []string{"fetch-data", "preprocess"},
}
// "analyze" waits for both "fetch-data" and "preprocess"
```

**Transitive Dependencies**:
```
A → B → C
A depends on B, B depends on C
Therefore A transitively depends on C
```

**Diamond Dependencies**:
```
    A
   / \
  B   C
   \ /
    D
```
D waits for both B and C; B and C can run in parallel after A.

### Dependency Validation

The system automatically:
- Detects circular dependencies
- Validates all dependency IDs exist
- Computes topological order (Kahn's algorithm)
- Groups tasks by dependency level (BFS)

**Cycle Detection**:
```go
// This will fail validation:
tasks := []agent.Task{
    {ID: "A", Dependencies: []string{"B"}},
    {ID: "B", Dependencies: []string{"C"}},
    {ID: "C", Dependencies: []string{"A"}},  // Cycle!
}
```

Error: `dependency cycle detected: A → B → C → A`

### Subtasks

Subtasks create hierarchical task structures:

```go
task := agent.Task{
    ID: "build-app",
    Subtasks: []agent.Task{
        {ID: "frontend", Description: "Build UI"},
        {ID: "backend", Description: "Build API"},
        {ID: "database", Description: "Setup DB"},
    },
}
```

**Execution Order**:
- Parent task is marked as started
- Subtasks execute (can be parallel if independent)
- Parent task is marked as completed after all subtasks

**Nesting Limits**:
- Default: 3 levels (configurable via MaxDepth)
- Prevents infinite recursion
- Keeps plans manageable

## Goal-Oriented Planning

### Goal States

Define success criteria for early termination:

```go
plan := &agent.Plan{
    Goal: "Find 3 high-quality research papers",
    GoalState: agent.GoalState{
        Description: "Found 3 papers with relevance >= 0.8",
        Criteria: []agent.GoalCriterion{
            {
                Name:      "papers_found",
                Expected:  "3",
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
    },
}
```

### Periodic Goal Checking

Configure how often to check goal completion:

```go
config := agent.DefaultPlannerConfig()
config.GoalCheckInterval = 5  // Check every 5 tasks
```

**Benefits**:
- Early termination when goal met
- Avoid unnecessary work
- Faster overall execution

**Example**:
```
Plan: Search 20 databases for papers
Goal: Find 3 good papers

Execution:
- Tasks 1-5: Find 2 papers → Check goal (not met)
- Tasks 6-10: Find 1 more paper → Check goal (MET!)
- Tasks 11-20: SKIPPED (goal already satisfied)

Result: 50% time savings
```

### Dynamic Goal Updates

Goals can be updated during execution:

```go
// In task execution callback:
if foundPaper.Relevance > 0.9 {
    plan.GoalState.Criteria[0].Satisfied = true
}
```

## Performance Optimization

### Choosing the Right Strategy

**Decision Tree**:
```
Are all tasks independent?
├─ Yes → Are tasks consistent in complexity?
│  ├─ Yes → Use Parallel
│  └─ No → Use Adaptive
└─ No → Do tasks have strict ordering?
   ├─ Yes → Use Sequential
   └─ No → Use Adaptive
```

**Workload Characteristics**:

| Workload | Strategy | Reason |
|----------|----------|--------|
| 10 API calls | Parallel | Independent, I/O-bound |
| Multi-phase pipeline | Adaptive | Mixed characteristics |
| Setup → Configure → Run | Sequential | Strict ordering |
| Unknown complexity | Adaptive | Learns and optimizes |
| Batch processing | Parallel | Maximize throughput |

### MaxParallel Tuning

**Guidelines**:
- **CPU-bound**: Set to number of CPU cores (e.g., 4-8)
- **I/O-bound**: Higher values (e.g., 10-20)
- **Rate-limited APIs**: Match rate limit (e.g., 3-5)
- **Memory-intensive**: Lower values to avoid OOM

**Measurement**:
```bash
# Benchmark different values:
MaxParallel=5:  20 tasks in 8.2s  (2.4 tasks/sec)
MaxParallel=10: 20 tasks in 5.1s  (3.9 tasks/sec)
MaxParallel=20: 20 tasks in 4.8s  (4.2 tasks/sec) ← Diminishing returns
```

**Optimal Value**: When increasing MaxParallel no longer improves throughput.

### Reducing Overhead

**1. Batch Size**
- Larger batches = less frequent strategy checks
- Smaller batches = quicker adaptation
- Default: Determined by dependency levels

**2. Goal Check Interval**
```go
config.GoalCheckInterval = 10  // Check every 10 tasks
// vs.
config.GoalCheckInterval = 1   // Check every task (higher overhead)
```

**3. Timeline Events**
- Disable if not needed: `config.EnableTimeline = false`
- Reduces memory and processing overhead

## Common Patterns

### 1. ETL Pipeline

Extract → Transform → Load with mixed parallelism:

```go
plan := agent.NewPlan("ETL Pipeline", agent.StrategyAdaptive)

// Extract (parallel - multiple sources)
for _, source := range sources {
    plan.AddTask(agent.Task{
        ID: "extract-" + source,
        Description: "Extract from " + source,
    })
}

// Transform (sequential - data processing)
plan.AddTask(agent.Task{
    ID: "transform",
    Description: "Transform and clean data",
    Dependencies: extractTaskIDs,
})

// Load (parallel - multiple destinations)
for _, dest := range destinations {
    plan.AddTask(agent.Task{
        ID: "load-" + dest,
        Description: "Load to " + dest,
        Dependencies: []string{"transform"},
    })
}
```

### 2. Research & Analysis

Gather data in parallel, synthesize sequentially:

```go
plan := agent.NewPlan("Market Research", agent.StrategyParallel)

// Phase 1: Parallel research
topics := []string{"AI", "ML", "Robotics"}
for _, topic := range topics {
    plan.AddTask(agent.Task{
        ID: "research-" + topic,
        Description: "Research " + topic + " trends",
        Type: agent.TaskTypeObservation,
    })
}

// Phase 2: Sequential synthesis
plan.AddTask(agent.Task{
    ID: "synthesize",
    Description: "Synthesize all research into report",
    Type: agent.TaskTypeAggregate,
    Dependencies: researchTaskIDs,
})
```

### 3. Content Generation

Generate multiple variants in parallel:

```go
plan := agent.NewPlan("Generate Marketing Content", agent.StrategyParallel)

platforms := []string{"Twitter", "LinkedIn", "Blog"}
for _, platform := range platforms {
    plan.AddTask(agent.Task{
        ID: "generate-" + platform,
        Description: "Generate content for " + platform,
        Type: agent.TaskTypeAction,
    })
}

// Review all in parallel too
for _, platform := range platforms {
    plan.AddTask(agent.Task{
        ID: "review-" + platform,
        Description: "Review and edit " + platform + " content",
        Type: agent.TaskTypeDecision,
        Dependencies: []string{"generate-" + platform},
    })
}
```

### 4. Iterative Refinement

Loop-like pattern with goal-based termination:

```go
plan := agent.NewPlan("Optimize Algorithm", agent.StrategySequential)
plan.GoalState = agent.GoalState{
    Description: "Achieve 95% accuracy",
    Criteria: []agent.GoalCriterion{
        {Name: "accuracy", Expected: "0.95", Operator: ">="},
    },
}

// Add multiple iteration tasks
for i := 1; i <= 10; i++ {
    plan.AddTask(agent.Task{
        ID: fmt.Sprintf("iteration-%d", i),
        Description: "Train and evaluate iteration " + fmt.Sprint(i),
    })
}

// Stops early if accuracy >= 95%
```

### 5. Dependency Fan-Out/Fan-In

Diamond pattern for distributed work:

```go
// Fan-out: 1 → N
plan.AddTask(agent.Task{ID: "prepare-data"})

for i := 0; i < N; i++ {
    plan.AddTask(agent.Task{
        ID: fmt.Sprintf("process-%d", i),
        Dependencies: []string{"prepare-data"},
    })
}

// Fan-in: N → 1
plan.AddTask(agent.Task{
    ID: "aggregate-results",
    Dependencies: processingTaskIDs,
})
```

## Best Practices

### 1. Plan Design

**DO**:
- ✅ Keep tasks atomic (single responsibility)
- ✅ Use descriptive task names
- ✅ Specify task types correctly
- ✅ Define clear goal criteria
- ✅ Group related tasks with subtasks

**DON'T**:
- ❌ Create overly granular tasks (< 1 second each)
- ❌ Mix responsibilities in one task
- ❌ Create unnecessary dependencies
- ❌ Exceed 3 levels of nesting
- ❌ Forget to validate dependencies

### 2. Strategy Selection

**DO**:
- ✅ Start with Adaptive for unknown workloads
- ✅ Profile before choosing fixed strategy
- ✅ Consider I/O vs CPU characteristics
- ✅ Test with realistic workloads

**DON'T**:
- ❌ Blindly use Parallel for everything
- ❌ Ignore dependency structures
- ❌ Set MaxParallel too high (resource exhaustion)
- ❌ Forget about external rate limits

### 3. Error Handling

**DO**:
- ✅ Implement task-level error handlers
- ✅ Use context for cancellation
- ✅ Log failures with context
- ✅ Consider retry logic for transient errors

**DON'T**:
- ❌ Ignore task failures in parallel execution
- ❌ Let one failure stop all independent tasks
- ❌ Swallow errors without logging
- ❌ Retry indefinitely without backoff

### 4. Monitoring

**DO**:
- ✅ Track timeline events for debugging
- ✅ Monitor performance metrics
- ✅ Set up alerting for long-running plans
- ✅ Log strategy switches in Adaptive mode

**DON'T**:
- ❌ Enable verbose logging in production
- ❌ Ignore slow tasks that block others
- ❌ Skip performance analysis
- ❌ Forget to clean up resources

### 5. Testing

**DO**:
- ✅ Test with mock LLM generators
- ✅ Verify dependency ordering
- ✅ Test error scenarios
- ✅ Benchmark critical paths

**DON'T**:
- ❌ Test only with real API calls (slow, flaky)
- ❌ Skip edge cases (cycles, missing deps)
- ❌ Ignore performance regressions
- ❌ Test only happy paths

## Troubleshooting

### Slow Execution

**Symptoms**: Plan takes much longer than expected

**Diagnosis**:
1. Check timeline events for bottlenecks
2. Identify slowest tasks
3. Verify strategy is appropriate
4. Check MaxParallel setting

**Solutions**:
- Switch to Parallel if tasks are independent
- Increase MaxParallel for I/O-bound work
- Optimize slow tasks
- Consider breaking large tasks into smaller ones

### High Memory Usage

**Symptoms**: Memory consumption grows during execution

**Diagnosis**:
1. Check number of concurrent tasks
2. Review task result sizes
3. Monitor subtask nesting

**Solutions**:
- Reduce MaxParallel
- Limit subtask depth (MaxDepth)
- Clear task results after aggregation
- Process in smaller batches

### Dependency Errors

**Symptoms**: "dependency cycle detected" or "missing dependency"

**Diagnosis**:
1. Visualize dependency graph
2. Check for typos in dependency IDs
3. Verify task IDs are unique

**Solutions**:
- Use validation before execution
- Draw dependency diagram
- Implement topological sort test
- Use constants for task IDs

### Strategy Not Switching

**Symptoms**: Adaptive strategy stays in one mode

**Diagnosis**:
1. Check AdaptiveThreshold setting
2. Review performance metrics
3. Verify task independence

**Solutions**:
- Lower AdaptiveThreshold for more eager switching
- Ensure tasks have realistic execution times
- Check dependency structure (may force sequential)
- Add logging to strategy decision logic

### Goal Not Terminating Early

**Symptoms**: All tasks execute despite goal being met

**Diagnosis**:
1. Verify GoalCheckInterval is set
2. Check goal criteria are updated
3. Review goal satisfaction logic

**Solutions**:
- Set GoalCheckInterval (default: check at end only)
- Update criteria.Satisfied in task callbacks
- Use correct operators (>=, ==, etc.)
- Add debug logging to goal checks

## Next Steps

- **API Reference**: See [PLANNING_API.md](PLANNING_API.md) for detailed API documentation
- **Performance Guide**: See [PLANNING_PERFORMANCE.md](PLANNING_PERFORMANCE.md) for benchmarks and tuning
- **Examples**: Check [examples/](../examples/) for working code samples

## Version

This guide covers Planning Layer v0.7.1 (November 2025).
