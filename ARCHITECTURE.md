# Architecture & Design

This document explains the design decisions and architecture of the go-deep-agent library.

## Design Goals

1. **Simplicity** - One unified API for 90% of use cases
2. **Flexibility** - Escape hatch for advanced scenarios
3. **Clean Code** - Minimal boilerplate, clear intent
4. **Provider Agnostic** - Same API across different LLM providers
5. **Production Ready** - Proper error handling, context support

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                        User Code                         │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                    agent.Chat()                          │
│  (Unified API - handles all common use cases)           │
└───┬──────────────────┬──────────────────┬───────────────┘
    │                  │                  │
    ▼                  ▼                  ▼
┌────────┐      ┌────────────┐    ┌─────────────┐
│ Simple │      │  Streaming │    │  History    │
│  Chat  │      │            │    │  + Tools    │
└────────┘      └────────────┘    └─────────────┘
    │                  │                  │
    └──────────────────┴──────────────────┘
                       │
                       ▼
            ┌──────────────────────┐
            │   chatStream() or    │
            │ direct API call      │
            └──────────────────────┘
                       │
                       ▼
            ┌──────────────────────┐
            │   OpenAI Go Client   │
            └──────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        ▼                             ▼
┌───────────────┐           ┌─────────────────┐
│  OpenAI API   │           │  Ollama (local) │
└───────────────┘           └─────────────────┘
```

## Package Structure

```
agent/
├── config.go     # Configuration & initialization
│   ├── Provider type & constants
│   ├── Config struct
│   └── NewAgent() factory
│
└── agent.go      # Core implementation
    ├── Agent struct
    ├── ChatOptions & ChatResult
    ├── Chat() - unified method
    ├── chatStream() - private helper
    └── GetCompletion() - advanced API
```

### Why Split into 2 Files?

**config.go (62 lines)**
- Setup & initialization concerns
- Provider configuration
- Easy to extend with new providers

**agent.go (140 lines)**
- Business logic & execution
- Chat operations
- Focused on "what the agent does"

This follows the **Single Responsibility Principle** and makes the codebase easier to navigate.

## API Design Evolution

### Phase 1: Initial Design (Before Refactoring)
```go
Chat(ctx, message string) (string, error)
ChatStream(ctx, message string, callback func(string)) error
ChatWithHistory(ctx, messages []Message) (string, error)
ChatWithToolCalls(ctx, message string, tools []Tool) (*Completion, error)
```

**Problems:**
- 4 different methods for related functionality
- Difficult to combine features (e.g., streaming + history)
- Return types inconsistent
- Not extensible

### Phase 2: Unified API with Options Pattern
```go
Chat(ctx, message string, opts *ChatOptions) (*ChatResult, error)
```

**Benefits:**
- ✅ Single entry point
- ✅ Easy to combine features
- ✅ Consistent return type
- ✅ Extensible (add new options without breaking API)
- ✅ Backward compatible mindset

## Key Design Patterns

### 1. Options Pattern

Used for flexible configuration without breaking changes:

```go
type ChatOptions struct {
    Stream   bool
    OnStream func(string)
    Messages []Message
    Tools    []Tool
}

// Simple: pass nil
Chat(ctx, "Hello", nil)

// Advanced: configure options
Chat(ctx, "Hello", &ChatOptions{Stream: true, OnStream: callback})
```

**Why?**
- Clean API for simple cases
- Flexible for advanced cases
- Easy to add new fields
- No need for multiple constructors

### 2. Factory Pattern

`NewAgent()` encapsulates provider-specific initialization:

```go
func NewAgent(config Config) (*Agent, error) {
    // Validate config
    // Setup provider-specific options
    // Return initialized agent
}
```

**Benefits:**
- Validation in one place
- Provider abstraction
- Easy to test

### 3. Strategy Pattern (Implicit)

Different providers (OpenAI, Ollama) use same interface through unified client:

```go
// Same API works for both
openaiAgent, _ := NewAgent(Config{Provider: ProviderOpenAI, ...})
ollamaAgent, _ := NewAgent(Config{Provider: ProviderOllama, ...})

// Identical usage
openaiAgent.Chat(ctx, "Hello", nil)
ollamaAgent.Chat(ctx, "Hello", nil)
```

## Error Handling Strategy

### Principle: Fail Fast, Clear Messages

```go
if config.Model == "" {
    return nil, fmt.Errorf("model is required")
}

if err != nil {
    return nil, fmt.Errorf("chat completion error: %w", err)
}
```

**Guidelines:**
- Validate early (in `NewAgent`)
- Wrap errors with context
- Use `%w` for error wrapping (Go 1.13+)
- Return meaningful error messages

## Context Usage

All operations accept `context.Context`:

```go
func (a *Agent) Chat(ctx context.Context, ...) (*ChatResult, error)
```

**Purpose:**
- Request cancellation
- Timeouts
- Request-scoped values
- Proper cleanup

**Example:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := agent.Chat(ctx, "Long operation", nil)
```

## Streaming Implementation

### Challenge
Provide both streaming and non-streaming through same API.

### Solution
```go
if opts.Stream {
    return a.chatStream(ctx, params, opts.OnStream)
}
return a.chatSync(ctx, params)
```

### Streaming Flow
```
1. User calls Chat() with Stream: true
2. Create streaming client
3. Use ChatCompletionAccumulator for proper chunk handling
4. Call OnStream callback for each chunk
5. Return full content + completion when done
```

### Why ChatCompletionAccumulator?

OpenAI's official helper for streaming:
- Handles delta accumulation
- Detects when content/tool call finishes
- Proper state management

## Return Types

### ChatResult
```go
type ChatResult struct {
    Content    string                 // Quick access to text
    Completion *openai.ChatCompletion // Full details
}
```

**Design Decision:**
- `Content` for 90% of use cases (simple text access)
- `Completion` for advanced cases (tool calls, token usage, metadata)

### Why Not Just String?

Before: `Chat() (string, error)`
- ❌ No access to tool calls
- ❌ No token usage info
- ❌ Limited metadata

After: `Chat() (*ChatResult, error)`
- ✅ Quick access via `.Content`
- ✅ Full info via `.Completion`
- ✅ Extensible (add fields without breaking API)

## Provider Abstraction

### OpenAI
```go
opts = append(opts, option.WithAPIKey(config.APIKey))
client := openai.NewClient(opts...)
```

### Ollama
```go
opts = append(opts,
    option.WithBaseURL(config.BaseURL),
    option.WithAPIKey("ollama"), // Dummy key
)
client := openai.NewClient(opts...)
```

**Key Insight:** Ollama provides OpenAI-compatible API, so we can use the same client!

## Advanced API: GetCompletion()

### Purpose
Low-level access for:
- Structured outputs (JSON Schema)
- Fine-tuned parameters (temperature, top_p)
- Features not in ChatOptions

### Philosophy
```
Chat()           = High-level, covers 90% of use cases
GetCompletion()  = Escape hatch, covers remaining 10%
```

### Why Keep Both?

**Option A:** Put everything in ChatOptions
```go
type ChatOptions struct {
    Stream bool
    Messages []Message
    Tools []Tool
    Temperature *float64      // Too many fields!
    TopP *float64              // Becomes complex
    MaxTokens *int64           // Not everyone needs these
    ResponseFormat *Format     // Pollutes simple API
    // ... 20+ more fields
}
```

**Option B:** Separate methods (current design)
```go
// 90% of users
Chat(ctx, msg, &ChatOptions{Stream: true})

// 10% power users
GetCompletion(ctx, openai.ChatCompletionNewParams{
    Temperature: 0.7,
    ResponseFormat: jsonSchema,
})
```

## Testing Strategy

### Unit Tests
- Test each provider configuration
- Validate error handling
- Mock OpenAI client for tests

### Integration Tests
- Test against real OpenAI API (with key)
- Test against local Ollama instance
- Verify streaming behavior

### Example Test Structure
```go
func TestChat_Simple(t *testing.T) {
    agent, _ := NewAgent(Config{...})
    result, err := agent.Chat(ctx, "Hello", nil)
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Content)
}

func TestChat_Streaming(t *testing.T) {
    chunks := []string{}
    result, err := agent.Chat(ctx, "Hello", &ChatOptions{
        Stream: true,
        OnStream: func(s string) { chunks = append(chunks, s) },
    })
    assert.NoError(t, err)
    assert.NotEmpty(t, chunks)
}
```

## Performance Considerations

### Agent Reuse
```go
// Good: Create once, reuse
agent, _ := NewAgent(config)
for i := 0; i < 1000; i++ {
    agent.Chat(ctx, fmt.Sprintf("Q%d", i), nil)
}

// Bad: Create for each request
for i := 0; i < 1000; i++ {
    agent, _ := NewAgent(config)  // Expensive!
    agent.Chat(ctx, "Q", nil)
}
```

### Connection Pooling
The underlying `openai.Client` handles connection pooling automatically.

### Memory
- ChatOptions and ChatResult are lightweight
- Streaming doesn't buffer full response in memory

## Future Extensibility

### Adding New Providers
1. Add constant: `ProviderAnthropic = "anthropic"`
2. Add case in `NewAgent()` switch
3. Configure provider-specific client

### Adding New Options
```go
type ChatOptions struct {
    Stream   bool
    OnStream func(string)
    Messages []Message
    Tools    []Tool
    // New field - doesn't break existing code
    Temperature *float64  // Optional
}
```

### Adding New Result Fields
```go
type ChatResult struct {
    Content    string
    Completion *openai.ChatCompletion
    // New field - doesn't break existing code
    TokenUsage *TokenUsage  // Optional
}
```

## Design Trade-offs

### Simplicity vs Flexibility
- **Chose:** Simplicity for common cases, flexibility via options
- **Trade-off:** Slightly more complex implementation, but much better UX

### Single vs Multiple Methods
- **Chose:** Single `Chat()` method
- **Trade-off:** More logic in one method, but much cleaner API

### Abstraction Level
- **Chose:** High-level Chat() + low-level GetCompletion()
- **Trade-off:** Two methods to maintain, but serves both user groups

### Type Safety vs Simplicity
- **Chose:** Typed options structs
- **Trade-off:** More types to define, but compile-time safety

## Lessons Learned

1. **Start simple, refactor as needed** - Initial design had multiple methods, refactored to unified API
2. **Options pattern works well in Go** - Flexible and idiomatic
3. **Provide escape hatches** - GetCompletion() for power users
4. **Provider abstraction is key** - Same API for different backends
5. **Documentation matters** - Clear examples reduce support burden

## References

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://go.dev/doc/effective_go)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Ollama API Documentation](https://github.com/ollama/ollama/blob/main/docs/api.md)
