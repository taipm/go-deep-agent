# ReAct Pattern API Reference

**go-deep-agent v0.7.0**

Complete API documentation for the ReAct (Reasoning + Acting) pattern implementation.

---

## Table of Contents

1. [Core Types](#core-types)
2. [Configuration](#configuration)
3. [Builder Methods](#builder-methods)
4. [Callbacks](#callbacks)
5. [Errors](#errors)
6. [Advanced Features](#advanced-features)

---

## Core Types

### ReActStep

Represents one step in the ReAct reasoning loop.

```go
type ReActStep struct {
    Type      string                 // Step type: THOUGHT, ACTION, OBSERVATION, FINAL
    Content   string                 // Text content of the step
    Tool      string                 // Tool name (ACTION only)
    Args      map[string]interface{} // Tool arguments (ACTION only)
    Timestamp time.Time              // When this step occurred
    Error     error                  // Error if any (optional)
}
```

**Step Types:**

| Type | Description | Content Example |
|------|-------------|-----------------|
| `THOUGHT` | Reasoning about what to do | "I need to search for Paris weather" |
| `ACTION` | Tool execution request | "search(query=\"Paris weather\")" |
| `OBSERVATION` | Tool result (system-provided) | "Temperature: 15°C, Cloudy" |
| `FINAL` | Final answer | "The weather in Paris is 15°C" |

**Usage:**

```go
result, _ := ai.Ask(ctx, "What's the weather?")
reactResult := result.Metadata["react_result"].(*agent.ReActResult)

for i, step := range reactResult.Steps {
    switch step.Type {
    case agent.StepTypeThought:
        fmt.Printf("Thought: %s\n", step.Content)
    case agent.StepTypeAction:
        fmt.Printf("Action: %s(%v)\n", step.Tool, step.Args)
    case agent.StepTypeObservation:
        fmt.Printf("Result: %s\n", step.Content)
    case agent.StepTypeFinal:
        fmt.Printf("Answer: %s\n", step.Content)
    }
}
```

---

### ReActResult

Complete outcome of a ReAct execution.

```go
type ReActResult struct {
    Answer     string          // Final answer
    Steps      []ReActStep     // Complete reasoning trace
    Iterations int             // Number of loops executed
    Success    bool            // Whether execution succeeded
    Error      error           // Error if failed
    Metrics    *ReActMetrics   // Execution metrics (optional)
    Timeline   *ReActTimeline  // Event timeline (optional)
}
```

**Fields:**

- **Answer** (`string`): The final answer to the user's query
- **Steps** (`[]ReActStep`): Full trace of all reasoning steps
- **Iterations** (`int`): Number of thought-action cycles completed
- **Success** (`bool`):
  - `true`: Reached FINAL step with answer
  - `false`: Stopped due to error, timeout, or max iterations
- **Error** (`error`): Error if execution failed (nil if successful)
- **Metrics** (`*ReActMetrics`): Performance metrics (if enabled)
- **Timeline** (`*ReActTimeline`): Event timeline (if enabled)

**Example:**

```go
result, err := ai.Ask(ctx, "Calculate 15 * 7")
if err != nil {
    log.Fatal(err)
}

reactResult := result.Metadata["react_result"].(*agent.ReActResult)

fmt.Println("Answer:", reactResult.Answer)
fmt.Println("Iterations:", reactResult.Iterations)
fmt.Println("Success:", reactResult.Success)

if reactResult.Metrics != nil {
    fmt.Println("Duration:", reactResult.Metrics.Duration)
    fmt.Println("Tool calls:", reactResult.Metrics.ToolCalls)
}
```

---

### ReActMetrics

Tracks execution metrics for monitoring and optimization.

```go
type ReActMetrics struct {
    TotalIterations int           // Number of reasoning loops
    ToolCalls       int           // Number of tools executed
    Errors          int           // Number of errors encountered
    Duration        time.Duration // Total execution time
    TokensUsed      int           // Total LLM tokens consumed
    StartTime       time.Time     // Execution start
    EndTime         time.Time     // Execution end
}
```

**Usage:**

```go
if reactResult.Metrics != nil {
    m := reactResult.Metrics
    
    // Performance monitoring
    fmt.Printf("Completed in %v\n", m.Duration)
    fmt.Printf("Used %d tokens\n", m.TokensUsed)
    
    // Cost estimation (GPT-4o: $0.005/1K tokens)
    cost := float64(m.TokensUsed) / 1000.0 * 0.005
    fmt.Printf("Estimated cost: $%.4f\n", cost)
    
    // Efficiency check
    if m.TotalIterations > 10 {
        log.Println("⚠️  High iteration count, consider optimization")
    }
}
```

---

### ReActTimeline

Chronological log of all events during execution.

```go
type ReActTimeline struct {
    Events        []TimelineEvent // All events
    TotalDuration time.Duration   // Total execution time
}

type TimelineEvent struct {
    Timestamp time.Time              // When event occurred
    Type      string                 // Event type
    Content   string                 // Event description
    Duration  time.Duration          // Event duration
    Metadata  map[string]interface{} // Additional data
}
```

**Event Types:**

- `"step"`: Reasoning step completed
- `"tool_call"`: Tool execution started
- `"error"`: Error occurred
- `"complete"`: Execution finished

**Example:**

```go
if reactResult.Timeline != nil {
    for _, event := range reactResult.Timeline.Events {
        fmt.Printf("[%s] %s: %s (%v)\n",
            event.Timestamp.Format("15:04:05"),
            event.Type,
            event.Content,
            event.Duration,
        )
    }
}
```

---

## Configuration

### ReActConfig

Main configuration struct for ReAct behavior.

```go
type ReActConfig struct {
    // Execution limits
    MaxIterations  int           // Max thought-action cycles (default: 5)
    TimeoutPerStep time.Duration // Timeout for each step (default: 30s)
    
    // Parsing
    StrictParsing  bool          // Require exact format (default: false)
    
    // Behavior
    StopOnFirstAnswer bool        // Stop when "Answer:" found (default: true)
    IncludeThoughts   bool        // Include reasoning in response (default: true)
    
    // Error handling
    RetryOnError   bool          // Retry on tool failure (default: true)
    MaxRetries     int           // Max retry attempts (default: 2)
}
```

**Field Details:**

#### MaxIterations (int)

Maximum number of thought-action cycles.

- **Default**: `5`
- **Range**: `1-50` (practical limits)
- **Recommendation**:
  - Simple tasks: `3-5`
  - Medium tasks: `5-7`
  - Complex tasks: `7-10`

```go
// Quick task
WithReActMaxIterations(3)

// Production default
WithReActMaxIterations(7)
```

#### TimeoutPerStep (time.Duration)

Maximum time allowed for each step (thought + action).

- **Default**: `30 * time.Second`
- **Recommendation**:
  - Fast tools: `15s`
  - Normal: `30s`
  - Slow APIs: `60s`

```go
WithReActConfig(&agent.ReActConfig{
    TimeoutPerStep: 60 * time.Second,  // 1 minute per step
})
```

#### StrictParsing (bool)

Whether to require exact ReAct format.

- **Default**: `false` (recommended)
- **true**: Only accept exact format, fail on deviations
- **false**: Use fallback parsers for robustness

```go
// Development (catch format issues early)
StrictParsing: true

// Production (handle variations)
StrictParsing: false  // Recommended
```

#### StopOnFirstAnswer (bool)

Stop execution when final answer is detected.

- **Default**: `true`
- **true**: Exit immediately when "Answer:" found (cost efficient)
- **false**: Continue until max iterations (thoroughness)

```go
// Save costs
StopOnFirstAnswer: true  // Recommended

// Ensure completeness
StopOnFirstAnswer: false
```

#### IncludeThoughts (bool)

Include reasoning steps in final response.

- **Default**: `true`
- **true**: Response includes "Thought: ..." text (transparency)
- **false**: Only return final answer (cleaner)

```go
// For debugging/transparency
IncludeThoughts: true

// For production (cleaner output)
IncludeThoughts: false
```

#### RetryOnError (bool)

Retry tool execution on failure.

- **Default**: `true`
- **true**: Attempt retry up to MaxRetries times
- **false**: Fail immediately on tool error

```go
// Production (resilience)
RetryOnError: true

// Development (fail fast)
RetryOnError: false
```

#### MaxRetries (int)

Maximum retry attempts per tool call.

- **Default**: `2`
- **Range**: `0-5` (practical)

```go
// No retries
MaxRetries: 0

// Production default
MaxRetries: 2

// High reliability
MaxRetries: 3
```

---

## Builder Methods

### WithReActMode

Enable ReAct pattern execution.

```go
func (b *Builder) WithReActMode(enabled bool) *Builder
```

**Parameters:**

- `enabled` (bool): `true` to enable ReAct pattern

**Example:**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(calculator, search).
    WithReActMode(true)  // Enable ReAct

result, _ := ai.Ask(ctx, "What is 15 * 7?")
```

---

### WithReActConfig

Configure ReAct behavior.

```go
func (b *Builder) WithReActConfig(config *ReActConfig) *Builder
```

**Parameters:**

- `config` (*ReActConfig): Configuration struct

**Example:**

```go
config := &agent.ReActConfig{
    MaxIterations:     10,
    TimeoutPerStep:    60 * time.Second,
    StrictParsing:     false,
    RetryOnError:      true,
    MaxRetries:        3,
    StopOnFirstAnswer: true,
    IncludeThoughts:   true,
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActConfig(config)
```

---

### WithReActMaxIterations

Shortcut to set max iterations.

```go
func (b *Builder) WithReActMaxIterations(max int) *Builder
```

**Parameters:**

- `max` (int): Maximum iterations (1-50)

**Example:**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActMode(true).
    WithReActMaxIterations(7)  // Allow up to 7 reasoning loops
```

---

### WithReActStrictMode

Enable/disable strict parsing.

```go
func (b *Builder) WithReActStrictMode(strict bool) *Builder
```

**Parameters:**

- `strict` (bool): `true` for strict format validation

**Example:**

```go
// Development: catch format issues
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActMode(true).
    WithReActStrictMode(true)

// Production: allow flexibility
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActMode(true).
    WithReActStrictMode(false)  // Recommended
```

---

### WithReActFewShot

Add few-shot examples to guide the LLM.

```go
func (b *Builder) WithReActFewShot(examples []*ReActExample) *Builder
```

**Parameters:**

- `examples` ([]*ReActExample): List of example executions

**Example:**

```go
examples := []*agent.ReActExample{
    {
        Query: "What's 2+2?",
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

ai := agent.NewOpenAI("gpt-3.5-turbo", apiKey).  // Helps weaker models
    WithReActMode(true).
    WithReActFewShot(examples)
```

---

### WithReActTemplate

Use custom prompt template.

```go
func (b *Builder) WithReActTemplate(template *ReActTemplate) *Builder
```

**Parameters:**

- `template` (*ReActTemplate): Custom template

**Example:**

```go
template := &agent.ReActTemplate{
    SystemPrompt: "You are a research assistant. Use ReAct pattern.",
    InstructionPrompt: `Format:
Thought: <reasoning>
Action: <tool>(<args>)
Observation: <result>
Answer: <final answer>`,
    FewShotExamples: examples,
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActTemplate(template)
```

---

### WithReActCallbacks

Register callbacks for monitoring.

```go
func (b *Builder) WithReActCallbacks(callback *EnhancedReActCallback) *Builder
```

**Parameters:**

- `callback` (*EnhancedReActCallback): Callback handlers

**Example:**

```go
callback := &agent.EnhancedReActCallback{
    OnStepStart: func(iteration int, thought string) {
        log.Printf("[Step %d] %s", iteration, thought)
    },
    OnActionExecute: func(action, input string) {
        log.Printf("[Action] %s(%s)", action, input)
    },
    OnComplete: func(result *agent.ReActResult) {
        log.Printf("[Done] %d steps", len(result.Steps))
    },
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActCallbacks(callback)
```

---

### WithReActStreaming

Enable real-time event streaming.

```go
func (b *Builder) WithReActStreaming(enabled bool) *Builder
```

**Parameters:**

- `enabled` (bool): `true` to enable streaming

**Example:**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActMode(true).
    WithReActStreaming(true)

result, _ := ai.Ask(ctx, "Complex task")

// Listen to stream
for event := range result.ReActStream {
    fmt.Printf("[%s] %s\n", event.Type, event.Content)
}
```

---

## Callbacks

### EnhancedReActCallback

Callback interface for observing execution.

```go
type EnhancedReActCallback struct {
    OnStepStart     func(iteration int, thought string)
    OnActionExecute func(action, input string)
    OnObservation   func(result string)
    OnStepComplete  func(step *ReActStep)
    OnError         func(err error, iteration int)
    OnComplete      func(result *ReActResult)
}
```

**Methods:**

#### OnStepStart

Called when a new reasoning step begins.

```go
OnStepStart: func(iteration int, thought string) {
    fmt.Printf("💭 [Step %d] %s\n", iteration, thought)
}
```

#### OnActionExecute

Called before executing a tool.

```go
OnActionExecute: func(action, input string) {
    fmt.Printf("🔧 Calling %s(%s)\n", action, input)
}
```

#### OnObservation

Called after receiving tool result.

```go
OnObservation: func(result string) {
    fmt.Printf("👁️  Result: %s\n", result)
}
```

#### OnStepComplete

Called after each step finishes.

```go
OnStepComplete: func(step *ReActStep) {
    // Save to database, metrics, etc.
    db.SaveStep(step)
}
```

#### OnError

Called when an error occurs.

```go
OnError: func(err error, iteration int) {
    log.Printf("❌ Error at step %d: %v", iteration, err)
}
```

#### OnComplete

Called when execution finishes.

```go
OnComplete: func(result *ReActResult) {
    fmt.Printf("✅ Complete: %d steps, %v duration\n",
        len(result.Steps),
        result.Metrics.Duration,
    )
}
```

---

## Errors

### ReAct Error Codes

```go
const (
    ErrReActMaxIterations = "REACT_MAX_ITERATIONS"
    ErrReActTimeout       = "REACT_TIMEOUT"
    ErrReActParseFailure  = "REACT_PARSE_FAILURE"
    ErrReActToolNotFound  = "REACT_TOOL_NOT_FOUND"
    ErrReActToolError     = "REACT_TOOL_ERROR"
)
```

### Error Handling

```go
result, err := ai.Ask(ctx, query)
if err != nil {
    if reactErr, ok := err.(*agent.ReActError); ok {
        switch reactErr.Code {
        case agent.ErrReActMaxIterations:
            // Task too complex
            log.Println("Increase MaxIterations or simplify task")
            
        case agent.ErrReActParseFailure:
            // LLM format issue
            log.Println("Add few-shot examples or use better model")
            tool := reactErr.Details["tool"]
            log.Printf("Missing tool: %s", tool)
            
        case agent.ErrReActToolNotFound:
            // Tool not registered
            tool := reactErr.Details["tool"]
            log.Printf("Register tool: %s", tool)
            
        case agent.ErrReActToolError:
            // Tool execution failed
            tool := reactErr.Details["tool"]
            reason := reactErr.Details["reason"]
            log.Printf("Tool %s failed: %s", tool, reason)
            
        case agent.ErrReActTimeout:
            // Execution timeout
            log.Println("Increase TimeoutPerStep or optimize tools")
        }
    }
    return err
}
```

---

## Advanced Features

### Few-Shot Examples

```go
type ReActExample struct {
    Query  string
    Steps  []*ReActStep
    Answer string
}
```

**Example:**

```go
examples := []*agent.ReActExample{
    {
        Query: "Search for Paris population",
        Steps: []*agent.ReActStep{
            {
                Type:        agent.StepTypeThought,
                Content:     "I need to search for Paris population",
                Timestamp:   time.Now(),
            },
            {
                Type:    agent.StepTypeAction,
                Content: "search(\"Paris population 2024\")",
                Tool:    "search",
                Args: map[string]interface{}{
                    "query": "Paris population 2024",
                },
                Timestamp: time.Now(),
            },
            {
                Type:      agent.StepTypeObservation,
                Content:   "Paris has 2.1 million people",
                Timestamp: time.Now(),
            },
        },
        Answer: "Paris has approximately 2.1 million people as of 2024",
    },
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActFewShot(examples)
```

### Custom Templates

```go
type ReActTemplate struct {
    SystemPrompt      string
    InstructionPrompt string
    FewShotExamples   []*ReActExample
}
```

**Example:**

```go
template := &agent.ReActTemplate{
    SystemPrompt: `You are a financial analyst assistant.
Use the ReAct pattern to analyze data systematically.
Always cite your sources.`,
    
    InstructionPrompt: `Think step-by-step:
1. Understand the financial query
2. Identify required data sources
3. Execute calculations carefully
4. Verify results

Format:
Thought: <your analysis>
Action: <tool_name>(<arguments>)
Observation: <tool result>
Answer: <final analysis with citations>`,
    
    FewShotExamples: financialExamples,
}

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActTemplate(template)
```

### Streaming Events

```go
type ReActStreamEvent struct {
    Type        string                 // Event type
    Content     string                 // Event content
    Action      string                 // Tool name (action events)
    ActionInput string                 // Tool input (action events)
    Metadata    map[string]interface{} // Additional data
    Timestamp   time.Time              // Event time
}
```

**Event Types:**

- `"thought"`: Reasoning step
- `"action"`: Tool execution
- `"observation"`: Tool result
- `"answer"`: Final answer
- `"error"`: Error occurred

**Example:**

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithReActMode(true).
    WithReActStreaming(true)

result, _ := ai.Ask(ctx, "Multi-step task")

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

---

## See Also

- [ReAct Guide](../guides/REACT_GUIDE.md) - Conceptual overview and best practices
- [Performance Tuning](../guides/REACT_PERFORMANCE.md) - Optimization strategies
- [Migration Guide](../guides/MIGRATION_v0.7.0.md) - Upgrading from v0.6.0

---

**Questions?** Open an issue at https://github.com/taipm/go-deep-agent/issues
