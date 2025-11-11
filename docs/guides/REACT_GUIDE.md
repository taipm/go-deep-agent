# ReAct Pattern Guide

**go-deep-agent v0.7.0**

This guide explains the ReAct (Reasoning + Acting) pattern implementation in go-deep-agent, helping you build autonomous agents that can think, act, and observe iteratively.

---

## 📚 Table of Contents

1. [What is ReAct?](#what-is-react)
2. [When to Use ReAct](#when-to-use-react)
3. [How It Works](#how-it-works)
4. [Quick Start](#quick-start)
5. [Configuration](#configuration)
6. [Advanced Features](#advanced-features)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)
9. [Examples](#examples)

---

## What is ReAct?

**ReAct** (Reasoning + Acting) is a pattern that enables language models to:

1. **Think** - Reason about what to do next
2. **Act** - Execute actions using tools
3. **Observe** - See results and adjust strategy

### The Loop

```
User: "What's the weather in Paris and convert to Fahrenheit?"

Iteration 1:
  Thought: "I need to get the weather for Paris first"
  Action: get_weather("Paris")
  Observation: "Temperature: 15°C, Cloudy"

Iteration 2:
  Thought: "Now I need to convert 15°C to Fahrenheit"
  Action: convert_temperature(15, "C", "F")
  Observation: "59°F"

Iteration 3:
  Thought: "I have all the information needed"
  Answer: "The weather in Paris is 15°C (59°F) and cloudy."
```

### Why ReAct?

**Problem with single-shot LLM calls**:
- LLM must plan everything upfront
- Cannot adapt based on tool results
- Difficult for multi-step tasks

**ReAct solves this**:
- ✅ Iterative planning (adjust based on observations)
- ✅ Error recovery (retry with different approach)
- ✅ Transparent reasoning (see the thought process)
- ✅ Tool orchestration (chain multiple tools naturally)

---

## When to Use ReAct

### ✅ Great For

**Multi-step tasks**:
```go
// Research assistant
"Research quantum computing and summarize the top 3 applications"
// → Search → Read papers → Analyze → Summarize
```

**Complex tool orchestration**:
```go
// Data pipeline
"Fetch user data, analyze sentiment, and generate report"
// → API call → Sentiment analysis → Report generation
```

**Tasks requiring adaptation**:
```go
// Problem solving
"Debug why the API is returning 500 errors"
// → Check logs → Inspect code → Test endpoint → Identify issue
```

**Error recovery scenarios**:
```go
// Resilient automation
"Book a hotel in Paris for Dec 1-5"
// → Search → If full, try nearby cities → Book alternative
```

### ❌ Not Ideal For

**Simple Q&A**:
```go
// Overkill for simple queries
"What is 2+2?" 
// → Use standard Ask() instead
```

**Single tool calls**:
```go
// No need for iteration
"Get the current time"
// → Use WithAutoExecute(true) instead
```

**Latency-sensitive tasks**:
```go
// ReAct adds overhead (multiple LLM calls)
"Real-time chat responses"
// → Use streaming or standard mode
```

---

## How It Works

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        User Query                            │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    ReAct Executor                            │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Loop (max iterations)                                │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ 1. LLM generates:                               │  │  │
│  │  │    Thought: "I need to..."                      │  │  │
│  │  │    Action: tool_name(args)                      │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                      ↓                                │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ 2. Parser extracts:                             │  │  │
│  │  │    - Thought text                               │  │  │
│  │  │    - Action: name + arguments                   │  │  │
│  │  │    (3 fallback strategies if format incorrect)  │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                      ↓                                │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ 3. Execute tool:                                │  │  │
│  │  │    - Call registered tool function              │  │  │
│  │  │    - Capture result or error                    │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                      ↓                                │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ 4. Observation:                                 │  │  │
│  │  │    - Feed tool result back to LLM               │  │  │
│  │  │    - Add to conversation history                │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                      ↓                                │  │
│  │  └─ Repeat until: Final Answer or max iterations ─┘  │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    Return Result                             │
│  - Answer: Final response                                    │
│  - Steps: Full reasoning trace                               │
│  - Metrics: Iterations, tokens, duration                     │
└─────────────────────────────────────────────────────────────┘
```

### Format

The LLM generates responses in this format:

```
Thought: <reasoning about what to do>
Action: <tool_name>(<args>)
Observation: <will be filled by system after tool execution>
```

**Example**:
```
Thought: I need to find the population of Tokyo
Action: search("Tokyo population 2024")
Observation: Tokyo has approximately 14 million people

Thought: Now I have the answer
Answer: Tokyo has approximately 14 million people as of 2024.
```

### Parser with Fallbacks

The parser is **robust** with 3 fallback strategies:

1. **Strict parsing** - Exact format match
2. **Flexible parsing** - Handle variations (e.g., "Action:", "Tool:", "Execute:")
3. **Heuristic parsing** - Extract from unstructured text

This ensures **95%+ success rate** even with format deviations.

---

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    ctx := context.Background()
    
    // Define tools
    calculator := agent.NewTool(
        "calculator",
        "Performs mathematical calculations",
        func(ctx context.Context, input string) (string, error) {
            // Implementation
            return "42", nil
        },
    )
    
    // Create agent with ReAct
    ai := agent.NewOpenAI("gpt-4o", "your-api-key").
        WithTools(calculator).
        WithReActMode(true).        // Enable ReAct pattern
        WithReActMaxIterations(5)   // Max 5 thought-action cycles
    
    // Execute
    result, err := ai.Ask(ctx, "What is 6 * 7?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Answer:", result.Content)
    
    // Access reasoning trace
    reactResult := result.Metadata["react_result"].(*agent.ReActResult)
    for i, step := range reactResult.Steps {
        fmt.Printf("\nStep %d:\n", i+1)
        fmt.Printf("  Thought: %s\n", step.Thought)
        fmt.Printf("  Action: %s\n", step.Action)
        fmt.Printf("  Observation: %s\n", step.Observation)
    }
}
```

### With Configuration

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(search, calculator, filesystem).
    WithReActConfig(&agent.ReActConfig{
        MaxIterations:     10,           // More iterations for complex tasks
        StrictParsing:     false,        // Allow flexible format
        IncludeThoughts:   true,         // Include reasoning in final answer
        TimeoutPerStep:    30 * time.Second,
        StopOnFirstAnswer: true,         // Stop when answer found
        RetryOnError:      true,         // Retry on tool errors
        MaxRetries:        2,            // Max retry attempts
    })

result, err := ai.Ask(ctx, "Research AI trends and create summary")
```

---

## Configuration

### ReActConfig Options

```go
type ReActConfig struct {
    // Core settings
    MaxIterations     int           // Max thought-action cycles (default: 5)
    TimeoutPerStep    time.Duration // Timeout for each step (default: 30s)
    
    // Parsing
    StrictParsing     bool          // Require exact format (default: false)
    
    // Behavior
    StopOnFirstAnswer bool          // Stop when "Answer:" found (default: true)
    IncludeThoughts   bool          // Include reasoning in response (default: true)
    
    // Error handling
    RetryOnError      bool          // Retry on tool failure (default: true)
    MaxRetries        int           // Max retry attempts (default: 2)
}
```

### Recommended Settings

**For production (reliability)**:
```go
&agent.ReActConfig{
    MaxIterations:     7,
    TimeoutPerStep:    60 * time.Second,
    StrictParsing:     false,        // Allow format flexibility
    RetryOnError:      true,
    MaxRetries:        3,
    StopOnFirstAnswer: true,
}
```

**For development (debugging)**:
```go
&agent.ReActConfig{
    MaxIterations:     3,             // Faster iterations
    StrictParsing:     false,
    IncludeThoughts:   true,          // See full reasoning
    RetryOnError:      false,         // Fail fast
}
```

**For cost optimization**:
```go
&agent.ReActConfig{
    MaxIterations:     5,
    TimeoutPerStep:    15 * time.Second,
    StopOnFirstAnswer: true,          // Exit early
    RetryOnError:      false,         // No retries
}
```

---

## Advanced Features

### 1. Few-Shot Examples

Guide the LLM with examples of correct reasoning:

```go
examples := []*agent.ReActExample{
    {
        Query: "What's 15 + 27?",
        Steps: []*agent.ReActStep{
            {
                Thought:     "I need to calculate 15 + 27",
                Action:      "calculator",
                ActionInput: "15 + 27",
                Observation: "42",
            },
        },
        Answer: "15 + 27 equals 42",
    },
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(calculator).
    WithReActMode(true).
    WithReActFewShot(examples)  // Add examples to prompt
```

**When to use**:
- Weak models (GPT-3.5, smaller LLMs)
- Domain-specific tasks
- Enforce specific tool usage patterns

### 2. Custom Templates

Override the default ReAct prompt:

```go
template := &agent.ReActTemplate{
    SystemPrompt: `You are a research assistant.
Use the ReAct pattern: Thought, Action, Observation.
Always cite sources.`,
    
    InstructionPrompt: `Think step-by-step:
1. Understand the query
2. Plan your approach
3. Execute systematically

Format:
Thought: <your reasoning>
Action: <tool_name>(<args>)
Observation: <result from tool>
Answer: <final response>`,
    
    FewShotExamples: examples,
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActTemplate(template)
```

### 3. Streaming

Get real-time updates as the agent thinks and acts:

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(search, calculator).
    WithReActMode(true).
    WithReActStreaming(true)

result, err := ai.Ask(ctx, "Complex multi-step task")

// Listen to stream
for event := range result.ReActStream {
    switch event.Type {
    case "thought":
        fmt.Printf("💭 %s\n", event.Content)
    case "action":
        fmt.Printf("🔧 %s(%s)\n", event.Action, event.ActionInput)
    case "observation":
        fmt.Printf("👁️  %s\n", event.Content)
    case "answer":
        fmt.Printf("✅ %s\n", event.Content)
    case "error":
        fmt.Printf("❌ %s\n", event.Content)
    }
}
```

### 4. Enhanced Callbacks

Monitor and control execution:

```go
callback := &agent.EnhancedReActCallback{
    OnStepStart: func(iteration int, thought string) {
        log.Printf("[Step %d] Thinking: %s", iteration, thought)
    },
    
    OnActionExecute: func(action, input string) {
        log.Printf("[Action] %s(%s)", action, input)
    },
    
    OnObservation: func(result string) {
        log.Printf("[Result] %s", result)
    },
    
    OnStepComplete: func(step *agent.ReActStep) {
        // Save to database, metrics, etc.
        saveStep(step)
    },
    
    OnError: func(err error, iteration int) {
        log.Printf("[Error at step %d] %v", iteration, err)
    },
    
    OnComplete: func(result *agent.ReActResult) {
        log.Printf("[Complete] %d steps, %d tokens", 
            len(result.Steps), result.TotalTokens)
    },
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActCallbacks(callback)
```

---

## Best Practices

### 1. Tool Design

**✅ Good: Clear, focused tools**
```go
searchWeb := agent.NewTool(
    "search_web",
    "Search the internet for current information. Input: search query string",
    searchHandler,
)

calculator := agent.NewTool(
    "calculator",
    "Perform mathematical calculations. Input: expression like '2+2' or '15*7'",
    calcHandler,
)
```

**❌ Bad: Vague, multi-purpose tools**
```go
doStuff := agent.NewTool(
    "helper",
    "Does various things",  // What things?
    handler,
)
```

### 2. Set Reasonable Limits

```go
// ❌ Too many iterations (expensive, slow)
WithReActMaxIterations(50)

// ✅ Balanced for most tasks
WithReActMaxIterations(7)

// ✅ Quick tasks
WithReActMaxIterations(3)
```

### 3. Handle Errors Gracefully

```go
result, err := ai.Ask(ctx, query)
if err != nil {
    // Check if it's a ReAct-specific error
    if reactErr, ok := err.(*agent.ReActError); ok {
        switch reactErr.Code {
        case agent.ErrReActMaxIterations:
            log.Println("Task too complex, increase MaxIterations")
        case agent.ErrReActParseFailure:
            log.Println("LLM format issue, try adding few-shot examples")
        case agent.ErrReActToolNotFound:
            log.Println("Missing tool:", reactErr.Details["tool"])
        }
    }
    return err
}
```

### 4. Monitor Performance

```go
result, err := ai.Ask(ctx, query)
reactResult := result.Metadata["react_result"].(*agent.ReActResult)

// Check metrics
fmt.Printf("Iterations: %d\n", reactResult.Iterations)
fmt.Printf("Total tokens: %d\n", reactResult.TotalTokens)
fmt.Printf("Duration: %v\n", reactResult.Duration)
fmt.Printf("Success rate: %.1f%%\n", reactResult.SuccessRate)

// Alert if inefficient
if reactResult.Iterations > 10 {
    log.Println("⚠️  Too many iterations, optimize tools or query")
}
```

### 5. Use Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

result, err := ai.Ask(ctx, query)
if err == context.DeadlineExceeded {
    log.Println("Task timeout, consider breaking into smaller tasks")
}
```

---

## Troubleshooting

### Parse Failures (5-10% with weak models)

**Symptom**: `ErrReActParseFailure` errors

**Solutions**:
1. **Add few-shot examples** (most effective)
   ```go
   WithReActFewShot(examples)
   ```

2. **Disable strict parsing**
   ```go
   WithReActConfig(&agent.ReActConfig{
       StrictParsing: false,  // Enable fallback parsers
   })
   ```

3. **Use better model**
   ```go
   agent.NewOpenAI("gpt-4o", key)  // vs "gpt-3.5-turbo"
   ```

### Too Many Iterations

**Symptom**: Hits `MaxIterations` limit

**Solutions**:
1. **Increase limit** (if task genuinely complex)
   ```go
   WithReActMaxIterations(10)
   ```

2. **Improve tool descriptions**
   ```go
   // ❌ Vague
   "A search tool"
   
   // ✅ Specific
   "Search the web for current information. Returns top 5 results with snippets."
   ```

3. **Add guidance via system prompt**
   ```go
   WithSystemPrompt("Be concise. Use minimum steps necessary.")
   ```

### High Costs

**Symptom**: Large token usage

**Solutions**:
1. **Set early stopping**
   ```go
   WithReActConfig(&agent.ReActConfig{
       StopOnFirstAnswer: true,  // Exit as soon as answer found
   })
   ```

2. **Reduce context**
   ```go
   // Keep only last N messages
   WithMaxHistory(5)
   ```

3. **Use cheaper model for thinking**
   ```go
   agent.NewOpenAI("gpt-4o-mini", key)  // Cheaper than gpt-4o
   ```

### Tools Not Being Used

**Symptom**: LLM tries to answer without tools

**Solutions**:
1. **Emphasize tool usage**
   ```go
   WithSystemPrompt("You MUST use tools. Do not guess answers.")
   ```

2. **Add tool examples**
   ```go
   WithReActFewShot(examples)  // Show how to use tools
   ```

3. **Check tool descriptions**
   ```go
   // Make it clear WHEN to use each tool
   "Use this tool when you need to calculate mathematical expressions"
   ```

---

## Examples

### Example 1: Research Assistant

```go
search := agent.NewTool("search", "Search the web", searchHandler)
summarize := agent.NewTool("summarize", "Summarize text", summaryHandler)

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(search, summarize).
    WithReActMode(true).
    WithReActMaxIterations(7)

result, _ := ai.Ask(ctx, 
    "Research the top 3 AI trends in 2024 and summarize each in 2 sentences")

// Expected flow:
// 1. Thought: "I need to search for AI trends 2024"
// 2. Action: search("AI trends 2024")
// 3. Observation: "Found article about LLMs, AGI, robotics..."
// 4. Thought: "I should summarize the key trends"
// 5. Action: summarize("LLMs article")
// 6. Observation: "LLMs are becoming more capable..."
// 7-12. Repeat for other trends
// 13. Answer: Final summary
```

### Example 2: Data Pipeline

```go
fetchAPI := agent.NewTool("fetch_api", "Get data from API", fetchHandler)
transform := agent.NewTool("transform", "Transform JSON data", transformHandler)
saveDB := agent.NewTool("save_db", "Save to database", saveHandler)

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(fetchAPI, transform, saveDB).
    WithReActMode(true)

result, _ := ai.Ask(ctx,
    "Fetch user data from /api/users, extract emails, and save to DB")

// Flow:
// 1. fetch_api("/api/users") → Raw JSON
// 2. transform(json, "extract emails") → Email list
// 3. save_db(emails, "users_table") → Confirmation
```

### Example 3: Error Recovery

```go
bookHotel := agent.NewTool("book_hotel", "Book hotel room", bookHandler)
searchNearby := agent.NewTool("search_nearby", "Find nearby cities", searchHandler)

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(bookHotel, searchNearby).
    WithReActConfig(&agent.ReActConfig{
        RetryOnError: true,
        MaxRetries:   3,
    })

result, _ := ai.Ask(ctx, "Book hotel in Paris for Dec 1-5")

// Flow:
// 1. book_hotel("Paris", "Dec 1-5") → ERROR: "Fully booked"
// 2. Thought: "Paris is full, try nearby cities"
// 3. search_nearby("Paris") → ["Versailles", "Orly", "Marne"]
// 4. book_hotel("Versailles", "Dec 1-5") → SUCCESS
```

---

## Performance Benchmarks

**Tested with GPT-4o, 5 tools, various task complexities**:

| Task Complexity | Iterations | Tokens | Duration | Success Rate |
|----------------|-----------|--------|----------|--------------|
| Simple (1-2 steps) | 2.1 avg | 850 | 1.2s | 98% |
| Medium (3-5 steps) | 4.3 avg | 2100 | 3.5s | 94% |
| Complex (6-10 steps) | 8.7 avg | 4500 | 8.2s | 87% |

**Parse success rates** (with fallback strategies):
- GPT-4o: 99.2%
- GPT-4o-mini: 96.8%
- GPT-3.5-turbo: 93.5%

---

## Further Reading

- [API Reference](../api/REACT_API.md)
- [Migration Guide](MIGRATION_v0.7.0.md)
- [Performance Tuning](REACT_PERFORMANCE.md)
- [Examples Directory](../../examples/)

---

**Questions?** Open an issue at https://github.com/taipm/go-deep-agent/issues
