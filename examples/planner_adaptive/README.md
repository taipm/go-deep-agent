# Planning Layer - Adaptive Strategy Example

This example demonstrates the **Adaptive Execution Strategy** in the go-deep-agent Planning Layer, showing how the system dynamically switches between sequential and parallel execution based on runtime performance.

## Features Demonstrated

1. **Mixed Workload Adaptation**: Handle varying task complexity with automatic strategy selection
2. **Dynamic Strategy Switching**: Monitor performance and switch strategies mid-execution
3. **Multi-Phase Pipelines**: Optimize execution across different pipeline phases

## Prerequisites

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Run the Example

```bash
cd examples/planner_adaptive
go run main.go
```

## Example Output

```
🧠 Planning Layer - Adaptive Strategy Example
==============================================

✅ Agent initialized with GPT-4

📋 Example 1: Adaptive Strategy with Mixed Workload
----------------------------------------------------
🔄 Executing with adaptive strategy...
   Config: MaxParallel=5, AdaptiveThreshold=0.6

✅ Completed in 15.3s
📊 Metrics:
   - Tasks: 9
   - Success Rate: 100.0%
   - Avg Task Duration: 3.8s

🔀 Strategy Timeline:
   [12:45:01.234] Starting with sequential strategy
   [12:45:06.789] Switched to parallel (efficiency improved)

📋 Example 2: Dynamic Strategy Switching
----------------------------------------
Creating workload designed to trigger strategy switches...

🔄 Starting adaptive execution (threshold=0.5)...

✅ Completed in 28.7s

📅 Detailed Execution Timeline:
   [01] 12:45:15.001 - strategy_initialized: Starting with sequential strategy
   [02] 12:45:15.002 - task_started: Starting task: Quick task 1...
   [03] 12:45:18.345 - task_completed: Task completed in 3.3s
   [04] 12:45:18.346 - task_started: Starting task: Quick task 2...
   ...

📋 Example 3: Multi-Phase Pipeline
-----------------------------------
🔄 Running multi-phase pipeline with adaptive strategy...
   Pipeline: Research (||) → Synthesize (→) → Write (||) → Review (→)

✅ Pipeline completed in 35.2s
📊 Metrics:
   - Total Tasks: 9
   - Success Rate: 100.0%
   - Execution Time: 35.2s
   - Avg Task Duration: 3.9s

📈 Phase Analysis:
   Research: 3 tasks, ~4.2s
   Synthesis: 1 tasks, ~8.1s
   Writing: 4 tasks, ~5.3s
   Review: 1 tasks, ~6.8s
```

## Key Concepts

### Adaptive Strategy

The adaptive strategy automatically selects the best execution approach:

```go
config := agent.DefaultPlannerConfig()
config.Strategy = agent.StrategyAdaptive
config.MaxParallel = 5
config.AdaptiveThreshold = 0.6  // Switch if parallel efficiency < 60%

executor := agent.NewExecutor(config, agentInstance)
```

### How It Works

1. **Initial Strategy**: Starts with sequential execution
2. **Performance Monitoring**: Tracks metrics after each batch:
   - Tasks per second
   - Average latency
   - Parallel efficiency
3. **Switch Decision**: Evaluates if strategy change would improve performance
4. **Dynamic Switching**: Changes strategy mid-execution if beneficial

### Performance Metrics

The system monitors three key metrics:

**Tasks Per Second**
```
TasksPerSec = CompletedTasks / BatchDuration
```
Higher is better. Indicates overall throughput.

**Average Latency**
```
AvgLatency = TotalTaskDuration / CompletedTasks
```
Lower is better. Indicates per-task overhead.

**Parallel Efficiency**
```
ParallelEfficiency = (Sum of task durations) / (Wall time × Task count)
```
Closer to 1.0 is better. Indicates parallelization effectiveness.

### Strategy Switching Logic

**Switch from Sequential to Parallel when:**
- Multiple independent tasks available
- Parallel efficiency would likely be > threshold
- No long sequential dependency chains

**Switch from Parallel to Sequential when:**
- Parallel efficiency < AdaptiveThreshold
- Tasks have many dependencies (little parallelism)
- Coordination overhead exceeds benefits

### Configuration Parameters

**AdaptiveThreshold** (default: 0.5)
- Range: 0.0 to 1.0
- Lower values: More aggressive parallel execution
- Higher values: More conservative, prefers sequential

Recommendations:
- **0.3-0.4**: Aggressive parallelism (I/O-bound workloads)
- **0.5-0.6**: Balanced (most use cases)
- **0.7-0.8**: Conservative (CPU-bound or high overhead)

**MaxParallel** (default: 5)
- Limits concurrent tasks even when parallel is chosen
- See [Parallel Example](../planner_parallel/) for tuning guidance

## Use Cases

### 1. Variable Task Complexity

When tasks have unpredictable execution times:

```go
plan := agent.NewPlan("Mixed complexity workload", agent.StrategyAdaptive)

// Some tasks are quick (100ms)
plan.AddTask(agent.Task{ID: "quick-1", Description: "Quick lookup"})

// Others are slow (5s)
plan.AddTask(agent.Task{ID: "slow-1", Description: "Complex analysis"})

// Adaptive strategy handles both efficiently
```

### 2. Multi-Phase Pipelines

Pipelines with different phase characteristics:

```
Data Collection (parallel) → Aggregation (sequential) → 
Generation (parallel) → Validation (sequential)
```

Each phase gets optimal strategy automatically.

### 3. Unknown Workload Patterns

When you don't know in advance if tasks will be:
- Independent vs dependent
- Fast vs slow
- I/O-bound vs CPU-bound

Adaptive strategy learns and adjusts during execution.

### 4. Production Systems

For production systems with varying load:
- High load: Maximize parallel throughput
- Low load: Minimize overhead with sequential
- System adapts to current conditions

## Performance Characteristics

### Overhead

Adaptive strategy adds minimal overhead:
- **Metric tracking**: < 1ms per task
- **Strategy evaluation**: < 10ms per batch
- **Switching cost**: < 5ms

Total overhead: typically < 1% of execution time

### Benefits

When effective, adaptive strategy provides:
- **10-30% improvement** over wrong strategy choice
- **Near-optimal performance** without manual tuning
- **Robustness** to workload variations

### When to Use

**Use Adaptive when:**
- ✅ Workload characteristics unknown
- ✅ Mix of simple and complex tasks
- ✅ Multi-phase pipelines
- ✅ Production systems with varying load
- ✅ Prototype/experimentation phase

**Use Sequential when:**
- ✅ All tasks have dependencies
- ✅ Tasks are very fast (< 100ms)
- ✅ Strict ordering required

**Use Parallel when:**
- ✅ All tasks are independent
- ✅ Consistent task complexity
- ✅ Maximum throughput critical
- ✅ Well-understood workload

## Monitoring Strategy Switches

Track strategy decisions via timeline:

```go
result, _ := executor.Execute(ctx, plan)

for _, event := range result.Timeline {
    if event.Type == "strategy_initialized" || event.Type == "strategy_switched" {
        fmt.Printf("[%v] %s\n", event.Timestamp, event.Description)
    }
}
```

Event types:
- `strategy_initialized`: Initial strategy chosen
- `strategy_switched`: Strategy changed during execution
- `performance_check`: Metrics evaluated (if enabled)

## Advanced Configuration

### Custom Switching Logic

For advanced use cases, you can implement custom strategy selection:

```go
// This is internal to the executor, but understanding helps
// with AdaptiveThreshold tuning

// Current logic (simplified):
if currentStrategy == Parallel {
    if metrics.ParallelEfficiency < config.AdaptiveThreshold {
        // Switch to Sequential - parallelism not effective
    }
} else if currentStrategy == Sequential {
    if hasIndependentTasks && expectedEfficiency > config.AdaptiveThreshold {
        // Switch to Parallel - can benefit from concurrency
    }
}
```

### Batch Size Impact

Strategy evaluation happens per batch. Smaller batches:
- More frequent evaluation
- Quicker adaptation
- Higher overhead

Larger batches:
- Less frequent evaluation
- More stable strategy
- Lower overhead

Default batch size is determined by dependency levels.

## Troubleshooting

### Too Many Strategy Switches

**Problem**: Excessive switching (> 5 times per execution)

**Solutions**:
- Increase AdaptiveThreshold (more stable)
- Reduce MaxParallel (less aggressive parallelism)
- Use fixed strategy if workload is consistent

### Not Switching When Expected

**Problem**: Strategy stays sequential even with parallel opportunities

**Solutions**:
- Decrease AdaptiveThreshold (more eager to parallelize)
- Check task dependencies (may be more than expected)
- Verify MaxParallel is set appropriately

### Poor Performance

**Problem**: Adaptive performs worse than fixed strategy

**Solutions**:
- Profile to identify bottleneck
- Check if tasks are too fast (< 100ms) - use sequential
- Verify API rate limits aren't causing failures
- Consider fixed strategy for production if workload is predictable

## Related Examples

- [Basic Planning](../planner_basic/) - Sequential execution fundamentals
- [Parallel Execution](../planner_parallel/) - Pure parallel strategy

## Learn More

- [Planning Guide](../../docs/PLANNING_GUIDE.md) - Strategy selection guide
- [Planning API](../../docs/PLANNING_API.md) - Configuration reference
- [Performance Guide](../../docs/PLANNING_PERFORMANCE.md) - Benchmarks and tuning
