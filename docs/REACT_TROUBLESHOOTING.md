# ReAct Troubleshooting Guide

## Common Issues and Solutions

This guide helps you diagnose and fix common ReAct execution problems.

---

## Issue 1: "Max iterations reached without final answer"

### Symptoms
```
Error: ReAct max_iterations (iteration 5/5): Maximum iterations reached without calling final_answer()
```

### Root Causes
1. **Wrong task complexity**: Using default settings for complex tasks
2. **Auto-fallback disabled**: No graceful degradation when iterations exhausted
3. **LLM not guided**: No reminders to call final_answer()

### Solutions

#### ✅ Best Solution: Use Task Complexity (v0.7.6+)
```go
// Simple task (3 iterations, 30s timeout)
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActComplexity(agent.ReActTaskSimple).
    WithTools(...)

// Medium task (5 iterations, 60s timeout)
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActComplexity(agent.ReActTaskMedium).
    WithTools(...)

// Complex task (10 iterations, 120s timeout)
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActComplexity(agent.ReActTaskComplex).
    WithTools(...)
```

#### Alternative: Enable Auto-Fallback
```go
// Get best-effort answer instead of error
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActAutoFallback(true).  // Default in v0.7.6+
    WithTools(...)
```

#### Manual: Increase Iterations
```go
// Only if you know the task needs more iterations
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActMaxIterations(10).  // Increase from default 3
    WithTools(...)
```

---

## Issue 2: Task Times Out

### Symptoms
```
Error: ReAct timeout (iteration 3/5): ReAct execution timeout
```

### Root Causes
1. **Too short timeout**: Default timeout insufficient
2. **Slow LLM responses**: Network or API latency
3. **Complex task**: Needs more processing time

### Solutions

#### ✅ Use Appropriate Complexity
```go
// Complex tasks get 120s timeout automatically
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActComplexity(agent.ReActTaskComplex).
    WithTools(...)
```

#### Manual Timeout Adjustment
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActTimeout(180 * time.Second).  // 3 minutes
    WithTools(...)
```

---

## Issue 3: Wrong Task Complexity Choice

### How to Choose Complexity Level

| Complexity | Iterations | Timeout | Use For |
|-----------|-----------|---------|---------|
| **Simple** | 3 | 30s | Single calculation, simple lookup, direct answer |
| **Medium** | 5 | 60s | Multi-step reasoning, 2-3 tool calls, analysis |
| **Complex** | 10 | 120s | Advanced reasoning, multiple tools, complex workflows |

### Examples

#### Simple Tasks
- "What is 25 * 17?"
- "Get current time in UTC"
- "Search for 'Python tutorial'"

```go
.WithReActComplexity(agent.ReActTaskSimple)
```

#### Medium Tasks
- "Calculate average of [1,2,3,4,5] and find 20% of it"
- "Check if today is a weekend, if yes get tomorrow's date"
- "Search for weather and summarize in one sentence"

```go
.WithReActComplexity(agent.ReActTaskMedium)
```

#### Complex Tasks
- "Analyze sales data: calculate mean, median, find outliers, generate report"
- "Multi-step research: search multiple sources, compare, synthesize answer"
- "Financial analysis with multiple calculations and decision logic"

```go
.WithReActComplexity(agent.ReActTaskComplex)
```

---

## Issue 4: LLM Doesn't Call final_answer()

### Symptoms
- Iterations exhausted
- Answer in reasoning but no explicit final_answer()

### Solutions

#### ✅ Enable Iteration Reminders (Default in v0.7.6+)
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActIterationReminders(true).  // Reminds LLM at n-2, n-1, n
    WithTools(...)
```

#### Enable Force Final Answer
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActForceFinalAnswer(true).  // Auto-inject final_answer at max
    WithTools(...)
```

---

## Issue 5: Auto-Fallback Gives Poor Answers

### Symptoms
- Task completes but answer quality is low
- "⚠️ Auto-fallback activated" in response

### Root Cause
Task too complex for iteration budget

### Solutions

#### ✅ Increase Complexity Level
```go
// If using Simple, try Medium
.WithReActComplexity(agent.ReActTaskMedium)

// If using Medium, try Complex
.WithReActComplexity(agent.ReActTaskComplex)
```

#### Disable Auto-Fallback for Strict Enforcement
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActAutoFallback(false).  // Get error instead of poor answer
    WithReActMaxIterations(10).
    WithTools(...)
```

---

## Issue 6: Too Many Tool Calls

### Symptoms
- High iteration count
- Repetitive tool usage
- Slow execution

### Solutions

#### Use Simpler Approach
```go
// Maybe ReAct is overkill - try direct execution
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    // No ReAct mode - direct LLM call
    WithTools(...)

result, _ := ai.Execute(ctx, task)
```

#### Reduce Max Iterations
```go
.WithReActComplexity(agent.ReActTaskSimple)  // Limit to 3 iterations
```

#### Add System Prompt Guidance
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithSystem("Be efficient. Use tools only when necessary. Call final_answer() as soon as you have the answer.").
    WithTools(...)
```

---

## Debugging Tips

### 1. Enable Timeline Tracking
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActTimeline(true).  // Track all events
    WithTools(...)

result, _ := ai.Execute(ctx, task)

// Inspect timeline
for _, event := range result.Timeline.Events {
    fmt.Printf("[%s] %s\n", event.Type, event.Content)
}
```

### 2. Enable Metrics
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActMetrics(true).  // Track performance
    WithTools(...)

result, _ := ai.Execute(ctx, task)

fmt.Printf("Duration: %v\n", result.Metrics.Duration)
fmt.Printf("Iterations: %d\n", result.Metrics.TotalIterations)
fmt.Printf("Tool calls: %d\n", result.Metrics.ToolCalls)
```

### 3. Inspect Error Details
```go
result, err := ai.Execute(ctx, task)
if err != nil {
    var reactErr *agent.ReActError
    if errors.As(err, &reactErr) {
        fmt.Printf("Error type: %s\n", reactErr.Type)
        fmt.Printf("Iteration: %d/%d\n", reactErr.CurrentIteration, reactErr.MaxIterations)
        fmt.Printf("Steps completed: %d\n", len(reactErr.Steps))
        
        for _, suggestion := range reactErr.Suggestions {
            fmt.Printf("  💡 %s\n", suggestion)
        }
    }
}
```

---

## Migration from v0.7.5 to v0.7.6

### Old Approach (v0.7.5)
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActMaxIterations(5).  // Manual guessing
    WithTools(...)
```

### New Recommended Approach (v0.7.6+)
```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActComplexity(agent.ReActTaskMedium).  // Self-documenting
    WithTools(...)
```

### Benefits
- ✅ Self-documenting: Complexity level shows intent
- ✅ Optimal defaults: No guessing on iterations/timeouts
- ✅ Auto-fallback: Graceful degradation instead of errors
- ✅ Iteration reminders: Better LLM guidance
- ✅ Rich errors: Actionable debugging information

---

## Quick Reference

### Default Values (v0.7.6+)

| Setting | Default | Notes |
|---------|---------|-------|
| MaxIterations | 3 | Changed from 5 in v0.7.6 |
| Timeout | 30s | Per complexity level |
| EnableAutoFallback | true | New in v0.7.6 |
| EnableIterationReminders | true | New in v0.7.6 |
| ForceFinalAnswerAtMax | true | New in v0.7.6 |

### Complexity Presets

```go
// Simple: 3 iterations, 30s
agent.ReActTaskSimple

// Medium: 5 iterations, 60s  
agent.ReActTaskMedium

// Complex: 10 iterations, 120s
agent.ReActTaskComplex
```

---

## Still Having Issues?

1. Check examples: `examples/react_math/`, `examples/react_native/`
2. Enable debug mode: `.WithDebug()`
3. Review error suggestions in ReActError
4. Simplify the task or break into smaller steps
5. Consider if ReAct is the right approach (sometimes direct execution is better)

## Related Documentation

- [README.md](../README.md) - Main documentation
- [CHANGELOG.md](../CHANGELOG.md) - Version history
- [examples/](../examples/) - Code examples
