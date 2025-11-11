# Migration Guide: v0.6.0 → v0.7.0

**go-deep-agent**

This guide helps you upgrade from v0.6.0 to v0.7.0, which introduces the **ReAct pattern** for autonomous multi-step reasoning.

---

## Overview

**What's New in v0.7.0:**

- ✨ **ReAct Pattern** - Thought → Action → Observation loop for multi-step tasks
- ✨ **Enhanced Observability** - Metrics, timeline, and callbacks
- ✨ **Streaming Support** - Real-time progress updates
- ✨ **Few-Shot Examples** - Guide LLM behavior with examples
- ✨ **Custom Templates** - Override default prompts
- ✨ **Robust Parsing** - 3 fallback strategies for 95%+ success rate

**Breaking Changes:**

- ❌ None - v0.7.0 is **100% backward compatible**

**Upgrade Effort:**

- ⏱️ **0 minutes** - If you don't need ReAct (existing code works unchanged)
- ⏱️ **5-10 minutes** - To enable basic ReAct mode
- ⏱️ **30-60 minutes** - To leverage advanced ReAct features

---

## Quick Start (5 Minutes)

### Before (v0.6.0)

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(calculator, search).
    WithAutoExecute(true)

result, err := ai.Ask(ctx, "What is 6 * 7 and what's the weather in Paris?")
// Single-shot execution
```

### After (v0.7.0)

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(calculator, search).
    WithReActMode(true)  // Enable ReAct pattern

result, err := ai.Ask(ctx, "What is 6 * 7 and what's the weather in Paris?")
// Multi-step reasoning:
// 1. Thought: "First calculate 6*7"
// 2. Action: calculator("6*7")
// 3. Observation: "42"
// 4. Thought: "Now get Paris weather"
// 5. Action: search("Paris weather")
// 6. Observation: "15°C, Cloudy"
// 7. Answer: "6*7 = 42. Weather in Paris is 15°C and cloudy."
```

**That's it!** Just add `.WithReActMode(true)`.

---

## Migration Scenarios

### Scenario 1: Keep Existing Behavior (No Migration)

If you're happy with current behavior, **do nothing**.

```go
// v0.6.0 code
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithAutoExecute(true)

result, _ := ai.Ask(ctx, "Simple question")
```

✅ **Works identically in v0.7.0** - No changes needed.

---

### Scenario 2: Enable ReAct for Multi-Step Tasks

For tasks requiring multiple tools or reasoning steps:

**Before (v0.6.0):**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(search, calculator, filesystem).
    WithAutoExecute(true)

// Relies on LLM to plan everything in one shot
result, _ := ai.Ask(ctx, "Research quantum computing and summarize top 3 trends")
```

**After (v0.7.0):**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(search, calculator, filesystem).
    WithReActMode(true).            // Enable iterative reasoning
    WithReActMaxIterations(7)       // Allow up to 7 steps

result, _ := ai.Ask(ctx, "Research quantum computing and summarize top 3 trends")

// Access reasoning trace
reactResult := result.Metadata["react_result"].(*agent.ReActResult)
for _, step := range reactResult.Steps {
    fmt.Printf("%s: %s\n", step.Type, step.Content)
}
```

**Benefits:**

- Better handling of complex tasks
- Transparent reasoning process
- Error recovery via retry logic
- Natural tool orchestration

---

### Scenario 3: Add Monitoring and Observability

Track execution metrics and events:

**Before (v0.6.0):**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...)

result, _ := ai.Ask(ctx, query)
// No visibility into execution
```

**After (v0.7.0):**

```go
// Option 1: Metrics
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithReActMode(true)

result, _ := ai.Ask(ctx, query)

reactResult := result.Metadata["react_result"].(*agent.ReActResult)
if reactResult.Metrics != nil {
    fmt.Printf("Iterations: %d\n", reactResult.Metrics.TotalIterations)
    fmt.Printf("Duration: %v\n", reactResult.Metrics.Duration)
    fmt.Printf("Tokens: %d\n", reactResult.Metrics.TokensUsed)
    fmt.Printf("Tool calls: %d\n", reactResult.Metrics.ToolCalls)
}

// Option 2: Callbacks
callback := &agent.EnhancedReActCallback{
    OnStepStart: func(iteration int, thought string) {
        log.Printf("[Step %d] %s", iteration, thought)
    },
    OnComplete: func(result *agent.ReActResult) {
        log.Printf("Completed in %d steps", len(result.Steps))
    },
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActCallbacks(callback)
```

---

### Scenario 4: Real-Time Progress Updates

Stream execution events in real-time:

**Before (v0.6.0):**

```go
result, _ := ai.Ask(ctx, query)
// Wait for completion, no progress updates
```

**After (v0.7.0):**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActMode(true).
    WithReActStreaming(true)

result, _ := ai.Ask(ctx, "Complex multi-step task")

// Real-time progress
for event := range result.ReActStream {
    switch event.Type {
    case "thought":
        fmt.Printf("💭 Thinking: %s\n", event.Content)
    case "action":
        fmt.Printf("🔧 Action: %s\n", event.Action)
    case "observation":
        fmt.Printf("👁️  Result: %s\n", event.Content)
    case "answer":
        fmt.Printf("✅ Answer: %s\n", event.Content)
    }
}
```

---

### Scenario 5: Improve Weak Model Performance

Guide smaller/cheaper models with examples:

**Before (v0.6.0):**

```go
// GPT-3.5 sometimes fails on complex tasks
ai := agent.NewOpenAI("gpt-3.5-turbo", apiKey).
    WithTools(calculator)

result, _ := ai.Ask(ctx, "Calculate (15 * 7) + (22 / 2)")
// Inconsistent results
```

**After (v0.7.0):**

```go
// Add few-shot examples to guide behavior
examples := []*agent.ReActExample{
    {
        Query: "What is 2 + 2?",
        Steps: []*agent.ReActStep{
            {
                Thought:     "I need to calculate 2+2",
                Action:      "calculator",
                ActionInput: "2+2",
                Observation: "4",
            },
        },
        Answer: "2+2 equals 4",
    },
}

ai := agent.NewOpenAI("gpt-3.5-turbo", apiKey).
    WithTools(calculator).
    WithReActMode(true).
    WithReActFewShot(examples)

result, _ := ai.Ask(ctx, "Calculate (15 * 7) + (22 / 2)")
// Much more reliable with examples
```

---

## Configuration Migration

### Simple Configuration

**v0.6.0 equivalent:**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithAutoExecute(true)
```

**v0.7.0 with ReAct:**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithReActMode(true)  // Simple enable
```

### Advanced Configuration

**v0.7.0 full control:**

```go
config := &agent.ReActConfig{
    MaxIterations:     10,                    // Allow up to 10 reasoning loops
    TimeoutPerStep:    60 * time.Second,      // 1 minute per step
    StrictParsing:     false,                 // Allow flexible format
    RetryOnError:      true,                  // Retry on tool failures
    MaxRetries:        3,                     // Up to 3 retry attempts
    StopOnFirstAnswer: true,                  // Exit when answer found
    IncludeThoughts:   true,                  // Include reasoning in response
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithReActConfig(config)
```

---

## API Changes

### New Methods

| Method | Purpose | Example |
|--------|---------|---------|
| `WithReActMode(bool)` | Enable ReAct pattern | `WithReActMode(true)` |
| `WithReActConfig(*ReActConfig)` | Full configuration | See above |
| `WithReActMaxIterations(int)` | Set iteration limit | `WithReActMaxIterations(7)` |
| `WithReActStrictMode(bool)` | Strict parsing | `WithReActStrictMode(false)` |
| `WithReActFewShot([]*ReActExample)` | Add examples | See Scenario 5 |
| `WithReActTemplate(*ReActTemplate)` | Custom prompts | See API docs |
| `WithReActCallbacks(*EnhancedReActCallback)` | Monitoring | See Scenario 3 |
| `WithReActStreaming(bool)` | Real-time events | See Scenario 4 |

### New Types

| Type | Purpose |
|------|---------|
| `ReActStep` | One reasoning step |
| `ReActResult` | Complete execution result |
| `ReActConfig` | Configuration options |
| `ReActMetrics` | Performance metrics |
| `ReActTimeline` | Event timeline |
| `ReActExample` | Few-shot example |
| `ReActTemplate` | Custom prompt template |
| `EnhancedReActCallback` | Execution callbacks |
| `ReActStreamEvent` | Streaming event |

### Deprecated (None)

✅ **All v0.6.0 APIs still work** - No deprecations.

---

## Breaking Changes

### ❌ NONE

v0.7.0 is **100% backward compatible** with v0.6.0.

All existing code continues to work without modifications.

---

## Common Patterns

### Pattern 1: Basic ReAct

```go
// Minimal change from v0.6.0
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithReActMode(true)  // Only change
```

### Pattern 2: Production Setup

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithReActConfig(&agent.ReActConfig{
        MaxIterations:     7,
        TimeoutPerStep:    60 * time.Second,
        StrictParsing:     false,
        RetryOnError:      true,
        MaxRetries:        2,
        StopOnFirstAnswer: true,
    })
```

### Pattern 3: Development/Debugging

```go
callback := &agent.EnhancedReActCallback{
    OnStepStart: func(iteration int, thought string) {
        log.Printf("[%d] Thought: %s", iteration, thought)
    },
    OnActionExecute: func(action, input string) {
        log.Printf("→ %s(%s)", action, input)
    },
    OnObservation: func(result string) {
        log.Printf("← %s", result)
    },
    OnError: func(err error, iteration int) {
        log.Printf("✗ Error at step %d: %v", iteration, err)
    },
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActMode(true).
    WithReActStrictMode(true).      // Catch format issues
    WithReActCallbacks(callback).   // Full logging
    WithReActStreaming(true)        // Real-time events
```

### Pattern 4: Cost Optimization

```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).  // Cheaper model
    WithReActConfig(&agent.ReActConfig{
        MaxIterations:     5,           // Fewer iterations
        TimeoutPerStep:    15 * time.Second,
        StopOnFirstAnswer: true,        // Exit early
        RetryOnError:      false,       // No retries
    })
```

---

## Testing Your Migration

### Step 1: Add Tests

```go
func TestReActMigration(t *testing.T) {
    ai := agent.NewOpenAI("gpt-4o", apiKey).
        WithTools(testTools...).
        WithReActMode(true).
        WithReActMaxIterations(5)
    
    result, err := ai.Ask(context.Background(), "Test query")
    require.NoError(t, err)
    
    // Verify ReAct result exists
    reactResult, ok := result.Metadata["react_result"].(*agent.ReActResult)
    require.True(t, ok)
    require.NotNil(t, reactResult)
    
    // Check execution
    assert.True(t, reactResult.Success)
    assert.NotEmpty(t, reactResult.Steps)
    assert.NotEmpty(t, reactResult.Answer)
}
```

### Step 2: Compare Results

```go
// v0.6.0 baseline
oldAI := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithAutoExecute(true)

oldResult, _ := oldAI.Ask(ctx, query)

// v0.7.0 with ReAct
newAI := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(tools...).
    WithReActMode(true)

newResult, _ := newAI.Ask(ctx, query)

// Both should work, ReAct provides more transparency
fmt.Println("Old:", oldResult.Content)
fmt.Println("New:", newResult.Content)
```

---

## Performance Impact

**Expectations:**

| Metric | Change | Reason |
|--------|--------|--------|
| **Latency** | +10-30% | Multiple LLM calls (thought + action per iteration) |
| **Token usage** | +20-50% | More conversation turns |
| **Cost** | +20-50% | More tokens consumed |
| **Success rate** | +15-40% | Better handling of complex tasks |
| **Reliability** | +25% | Retry logic and error recovery |

**When is it worth it?**

- ✅ Complex multi-step tasks
- ✅ Tasks requiring tool orchestration
- ✅ Need for transparency/debugging
- ✅ Error recovery important
- ❌ Simple Q&A (stick with standard mode)
- ❌ Ultra-low latency requirements

---

## Rollback Plan

If you need to roll back:

```go
// v0.7.0 with ReAct
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActMode(true)  // ← Remove this line

// Back to v0.6.0 behavior
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithAutoExecute(true)
```

Or downgrade package:

```bash
go get github.com/taipm/go-deep-agent@v0.6.0
```

---

## Troubleshooting

### Issue 1: Parse Failures

**Symptom:** `ErrReActParseFailure` errors

**Solution:**

```go
// Add few-shot examples
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActFewShot(examples).
    WithReActStrictMode(false)  // Allow flexible parsing
```

### Issue 2: Too Many Iterations

**Symptom:** Hits `MaxIterations` limit

**Solution:**

```go
// Increase limit
WithReActMaxIterations(10)

// Or improve tool descriptions
calculator := agent.NewTool(
    "calculator",
    "Performs math calculations. Input: expression like '2+2' or '15*7'",  // More specific
    handler,
)
```

### Issue 3: High Costs

**Symptom:** Large token usage

**Solution:**

```go
// Option 1: Early stopping
WithReActConfig(&agent.ReActConfig{
    StopOnFirstAnswer: true,  // Exit as soon as answer found
})

// Option 2: Cheaper model
agent.NewOpenAI("gpt-4o-mini", apiKey)

// Option 3: Fewer iterations
WithReActMaxIterations(3)
```

---

## Next Steps

1. **Read the guides:**
   - [ReAct Pattern Guide](REACT_GUIDE.md)
   - [API Reference](../api/REACT_API.md)
   - [Performance Tuning](REACT_PERFORMANCE.md)

2. **Try examples:**
   - `examples/react_simple/` - Basic usage
   - `examples/react_research/` - Multi-tool orchestration
   - `examples/react_streaming/` - Real-time events

3. **Join the community:**
   - GitHub Issues: <https://github.com/taipm/go-deep-agent/issues>
   - Discussions: <https://github.com/taipm/go-deep-agent/discussions>

---

## Summary

**Migration Checklist:**

- [ ] Update to v0.7.0: `go get github.com/taipm/go-deep-agent@v0.7.0`
- [ ] Review new ReAct APIs
- [ ] Add `.WithReActMode(true)` for multi-step tasks
- [ ] Configure `ReActConfig` for production
- [ ] Add callbacks/streaming if needed
- [ ] Test with your workload
- [ ] Monitor metrics and costs
- [ ] Adjust configuration based on results

**Key Takeaway:**

v0.7.0 is **100% backward compatible**. Adopt ReAct at your own pace, starting with the tasks that benefit most from multi-step reasoning.

---

**Questions?** Open an issue or discussion on GitHub!
