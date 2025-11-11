# Planning Layer - Parallel Execution Example

This example demonstrates the **Parallel Execution Strategy** in the go-deep-agent Planning Layer, showing how to efficiently process multiple independent tasks concurrently.

## Features Demonstrated

1. **Batch Processing**: Process 10 independent tasks with configurable concurrency
2. **Dependency-Aware Parallelism**: Execute tasks in parallel while respecting dependencies
3. **Performance Comparison**: Measure speedup from parallel vs sequential execution

## Prerequisites

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Run the Example

```bash
cd examples/planner_parallel
go run main.go
```

## Example Output

```
🚀 Planning Layer - Parallel Execution Example
================================================

✅ Agent initialized with GPT-4

📋 Example 1: Parallel Batch Processing
----------------------------------------
🔄 Processing 10 companies with MaxParallel=5...

✅ Completed in 8.2s
📊 Metrics:
   - Tasks: 10
   - Success Rate: 100.0%
   - Avg Task Duration: 4.1s
   - Throughput: 1.2 tasks/sec

📝 Sample Analysis Results:

1. Analyze Apple: market cap, revenue growth, key products:
   Apple Inc. (AAPL) - Market Cap: $2.8T, Revenue Growth: 8% YoY...

📋 Example 2: Dependency-Aware Parallel Execution
--------------------------------------------------
🔄 Executing dependency-aware parallel plan...
   Structure: Research → [3 parallel analyses] → Final Report

✅ Completed in 12.5s
📊 Timeline Events: 10

📅 Execution Timeline:
   [12:30:01.123] task_started: Starting task: Gather market data...
   [12:30:05.456] task_completed: Task completed in 4.3s
   [12:30:05.457] task_started: Starting task: Analyze technology trends...
   [12:30:05.458] task_started: Starting task: Analyze market size...
   [12:30:05.459] task_started: Starting task: Analyze competitive landscape...

📋 Example 3: Sequential vs Parallel Performance
-------------------------------------------------
🔄 Running 8 tasks sequentially...
🔄 Running 8 tasks in parallel (MaxParallel=4)...

📊 Performance Comparison:
   Sequential: 32.1s (0.25 tasks/sec)
   Parallel:   8.5s (0.94 tasks/sec)

⚡ Speedup: 3.78x faster

📈 Metrics Comparison:
   Sequential - Avg Duration: 4.0s
   Parallel   - Avg Duration: 4.2s
```

## Key Concepts

### Parallel Strategy

The parallel strategy executes tasks concurrently while respecting dependencies:

```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategyParallel
config.MaxParallel = 5  // Max 5 concurrent tasks

executor := agent.NewExecutor(config, agentInstance)
result, err := executor.Execute(ctx, plan)
```

### Independent Tasks

Tasks without dependencies can run in parallel:

```go
plan := agent.NewPlan("Batch processing", agent.StrategyParallel)

for i := 1; i <= 10; i++ {
    plan.AddTask(agent.Task{
        ID:          fmt.Sprintf("task-%d", i),
        Description: fmt.Sprintf("Process item %d", i),
        Type:        agent.TaskTypeAction,
        // No dependencies - can run in parallel
    })
}
```

### Dependency Graph

Tasks with dependencies execute in topologically sorted order:

```go
// Diamond dependency: A → B,C → D
plan.AddTask(agent.Task{ID: "A", Description: "Root"})
plan.AddTask(agent.Task{ID: "B", Dependencies: []string{"A"}})
plan.AddTask(agent.Task{ID: "C", Dependencies: []string{"A"}})
plan.AddTask(agent.Task{ID: "D", Dependencies: []string{"B", "C"}})
```

Execution timeline:
- Level 0: A executes alone
- Level 1: B and C execute in parallel (both depend on A)
- Level 2: D executes after B and C complete

### MaxParallel Configuration

Controls concurrency to avoid overwhelming the system:

```go
config.MaxParallel = 3  // At most 3 tasks run simultaneously
```

Guidelines:
- **CPU-bound tasks**: Set to number of CPU cores
- **I/O-bound tasks** (API calls): Higher values (5-10)
- **Rate-limited APIs**: Lower values (1-3) to respect limits

## Performance Benefits

Parallel execution provides speedup for:

1. **Independent batch operations** (e.g., analyzing multiple items)
2. **I/O-bound workloads** (e.g., API calls, web scraping)
3. **Data processing pipelines** with parallelizable stages

Expected speedup (independent tasks):
- MaxParallel=2: ~2x faster
- MaxParallel=5: ~3-4x faster
- MaxParallel=10: ~5-8x faster

Actual speedup depends on:
- Task execution time variance
- System resources (CPU, memory, network)
- External API rate limits

## Timeline Monitoring

Track execution progress with timeline events:

```go
result, _ := executor.Execute(ctx, plan)

for _, event := range result.Timeline {
    fmt.Printf("[%v] %s: %s\n", 
        event.Timestamp.Format("15:04:05"),
        event.Type,
        event.Description,
    )
}
```

Event types:
- `task_started`: Task begins execution
- `task_completed`: Task finishes successfully
- `task_failed`: Task encounters error

## Use Cases

### 1. Data Collection
Process multiple data sources in parallel:
- Fetch from 10 different APIs simultaneously
- Scrape 20 web pages concurrently
- Query multiple databases in parallel

### 2. Analysis Pipelines
Multi-stage processing with parallel phases:
- Stage 1: Gather data (sequential)
- Stage 2: Analyze each data point (parallel)
- Stage 3: Aggregate results (sequential)

### 3. Batch Operations
Apply same operation to many items:
- Generate summaries for 100 documents
- Classify 50 customer reviews
- Translate 30 text snippets

## Related Examples

- [Basic Planning](../planner_basic/) - Sequential execution fundamentals
- [Adaptive Strategy](../planner_adaptive/) - Dynamic strategy switching

## Learn More

- [Planning Guide](../../docs/PLANNING_GUIDE.md) - Comprehensive concepts
- [Planning API](../../docs/PLANNING_API.md) - API reference
- [Performance Guide](../../docs/PLANNING_PERFORMANCE.md) - Optimization tips
