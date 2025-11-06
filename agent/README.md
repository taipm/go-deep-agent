# Agent Package - Builder API Reference

Complete API reference for the `agent` package featuring the **Fluent Builder API** for building LLM-powered applications.

## Table of Contents

- [Quick Start](#quick-start)
- [Constructors](#constructors)
- [Core Methods](#core-methods)
- [Configuration Methods](#configuration-methods)
- [Conversation Management](#conversation-management)
- [Tool Calling](#tool-calling)
- [Structured Outputs](#structured-outputs)
- [Streaming](#streaming)
- [Error Handling & Recovery](#error-handling--recovery)
- [Types](#types)
- [Error Types](#error-types)

## Quick Start

```go
import "github.com/taipm/go-deep-agent/agent"

// Simple usage
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    Ask(ctx, "Hello, world!")

// With configuration
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithTemperature(0.7).
    WithMemory().
    Ask(ctx, "Explain Go concurrency")

// Streaming
agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(content string) {
        fmt.Print(content)
    }).
    Stream(ctx, "Write a story")
```

---

## Constructors

### NewOpenAI

Create a Builder for OpenAI with model and API key.

```go
func NewOpenAI(model, apiKey string) *Builder
```

**Parameters:**
- `model` - Model name (e.g., "gpt-4o-mini", "gpt-4", "gpt-3.5-turbo")
- `apiKey` - OpenAI API key

**Example:**
```go
builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY"))
```

### NewOllama

Create a Builder for Ollama with model name. Uses `http://localhost:11434/v1` by default.

```go
func NewOllama(model string) *Builder
```

**Parameters:**
- `model` - Ollama model name (e.g., "qwen2.5:3b", "llama3.2", "phi3")

**Example:**
```go
builder := agent.NewOllama("qwen2.5:3b")
```

### New

Generic constructor for any provider.

```go
func New(provider Provider, model string) *Builder
```

**Parameters:**
- `provider` - `ProviderOpenAI` or `ProviderOllama`
- `model` - Model name

**Example:**
```go
builder := agent.New(agent.ProviderOpenAI, "gpt-4o-mini").
    WithAPIKey(apiKey)
```

---

## Core Methods

### Ask

Send a message and get a response.

```go
func (b *Builder) Ask(ctx context.Context, message string) (string, error)
```

**Parameters:**
- `ctx` - Context for cancellation and timeouts
- `message` - User message to send

**Returns:**
- `string` - AI response text
- `error` - Error if request fails

**Example:**
```go
response, err := builder.Ask(ctx, "What is Go?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

### Stream

Stream response with callbacks.

```go
func (b *Builder) Stream(ctx context.Context, message string) (string, error)
```

**Parameters:**
- `ctx` - Context for cancellation
- `message` - User message

**Returns:**
- `string` - Complete response text
- `error` - Error if request fails

**Note:** Use `OnStream()` to set callback before calling `Stream()`.

**Example:**
```go
response, err := builder.
    OnStream(func(content string) {
        fmt.Print(content)
    }).
    Stream(ctx, "Tell me a story")
```

### StreamPrint

Convenience method that streams and prints to stdout.

```go
func (b *Builder) StreamPrint(ctx context.Context, message string) (string, error)
```

**Example:**
```go
response, err := builder.StreamPrint(ctx, "Write a haiku")
```

---

## Configuration Methods

### WithAPIKey

Set the API key (required for OpenAI).

```go
func (b *Builder) WithAPIKey(apiKey string) *Builder
```

**Example:**
```go
builder.WithAPIKey(os.Getenv("OPENAI_API_KEY"))
```

### WithModel

Change the model.

```go
func (b *Builder) WithModel(model string) *Builder
```

**Example:**
```go
builder.WithModel("gpt-4")
```

### WithBaseURL

Set custom endpoint URL.

```go
func (b *Builder) WithBaseURL(baseURL string) *Builder
```

**Example:**
```go
// For custom Ollama port
builder.WithBaseURL("http://localhost:11434/v1")

// For Azure OpenAI
builder.WithBaseURL("https://your-resource.openai.azure.com/v1")
```

### WithSystem

Set system prompt.

```go
func (b *Builder) WithSystem(prompt string) *Builder
```

**Example:**
```go
builder.WithSystem("You are a helpful assistant that explains concepts concisely")
```

### WithTemperature

Set sampling temperature (0.0 - 2.0). Lower = more focused, Higher = more creative.

```go
func (b *Builder) WithTemperature(temperature float64) *Builder
```

**Example:**
```go
builder.WithTemperature(0.7)  // Balanced
builder.WithTemperature(0.0)  // Deterministic
builder.WithTemperature(1.5)  // Very creative
```

### WithTopP

Set nucleus sampling (0.0 - 1.0). Alternative to temperature.

```go
func (b *Builder) WithTopP(topP float64) *Builder
```

**Example:**
```go
builder.WithTopP(0.9)
```

### WithMaxTokens

Set maximum tokens to generate.

```go
func (b *Builder) WithMaxTokens(maxTokens int64) *Builder
```

**Example:**
```go
builder.WithMaxTokens(500)  // Limit to 500 tokens
```

### WithPresencePenalty

Set presence penalty (-2.0 to 2.0). Positive values encourage new topics.

```go
func (b *Builder) WithPresencePenalty(penalty float64) *Builder
```

**Example:**
```go
builder.WithPresencePenalty(0.6)  // Encourage diverse topics
```

### WithFrequencyPenalty

Set frequency penalty (-2.0 to 2.0). Positive values reduce repetition.

```go
func (b *Builder) WithFrequencyPenalty(penalty float64) *Builder
```

**Example:**
```go
builder.WithFrequencyPenalty(0.5)  // Reduce repetitive words
```

### WithSeed

Set seed for reproducible outputs.

```go
func (b *Builder) WithSeed(seed int64) *Builder
```

**Example:**
```go
builder.WithSeed(42)  // Same seed = same output
```

### WithN

Set number of completions to generate.

```go
func (b *Builder) WithN(n int64) *Builder
```

**Example:**
```go
builder.WithN(3)  // Generate 3 different responses
```

---

## Conversation Management

### WithMemory

Enable automatic conversation memory. Messages are stored and sent with each request.

```go
func (b *Builder) WithMemory() *Builder
```

**Example:**
```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).WithMemory()

builder.Ask(ctx, "My name is Alice")
builder.Ask(ctx, "What's my name?")  // AI remembers: "Alice"
```

### WithMaxHistory

Limit conversation history with FIFO (First In, First Out) truncation.

```go
func (b *Builder) WithMaxHistory(max int) *Builder
```

**Parameters:**
- `max` - Maximum number of messages to keep (0 = unlimited)

**Example:**
```go
builder.WithMaxHistory(10)  // Keep last 10 messages only
```

**Note:** System prompt is never removed.

### GetHistory

Get current conversation history.

```go
func (b *Builder) GetHistory() []Message
```

**Returns:**
- `[]Message` - Copy of conversation history

**Example:**
```go
history := builder.GetHistory()
fmt.Printf("Conversation has %d messages\n", len(history))

// Iterate messages
for _, msg := range history {
    fmt.Printf("%s: %s\n", msg.Role, msg.Content)
}
```

### SetHistory

Set conversation history (for restoring sessions).

```go
func (b *Builder) SetHistory(messages []Message) *Builder
```

**Example:**
```go
// Save session
savedHistory := builder.GetHistory()
saveToDatabase(savedHistory)

// Later: restore session
loadedHistory := loadFromDatabase()
builder.SetHistory(loadedHistory)
```

### Clear

Clear conversation history while preserving system prompt.

```go
func (b *Builder) Clear() *Builder
```

**Example:**
```go
builder.Clear()  // Start fresh conversation
```

---

## Tool Calling

### NewTool

Create a new tool for function calling.

```go
func NewTool(name, description string) *Tool
```

**Example:**
```go
tool := agent.NewTool("get_weather", "Get current weather for a location")
```

### AddParameter

Add parameter to tool definition.

```go
func (t *Tool) AddParameter(name, paramType, description string, required bool) *Tool
```

**Parameters:**
- `name` - Parameter name
- `paramType` - Type: "string", "integer", "number", "boolean", "object", "array"
- `description` - Parameter description
- `required` - Whether parameter is required

**Example:**
```go
tool.AddParameter("location", "string", "City name", true).
    AddParameter("units", "string", "celsius or fahrenheit", false)
```

### WithHandler

Set tool execution handler.

```go
func (t *Tool) WithHandler(handler func(args string) (string, error)) *Tool
```

**Parameters:**
- `handler` - Function that receives JSON args and returns result

**Example:**
```go
tool.WithHandler(func(args string) (string, error) {
    var params struct {
        Location string `json:"location"`
        Units    string `json:"units"`
    }
    json.Unmarshal([]byte(args), &params)
    
    // Call weather API...
    return `{"temp": 25, "condition": "sunny"}`, nil
})
```

### WithTools

Register tools with Builder.

```go
func (b *Builder) WithTools(tools ...*Tool) *Builder
```

**Example:**
```go
builder.WithTools(weatherTool, calculatorTool)
```

### WithAutoExecute

Enable automatic tool execution.

```go
func (b *Builder) WithAutoExecute(enable bool) *Builder
```

**Example:**
```go
builder.WithAutoExecute(true)  // Automatically execute tool calls
```

### WithMaxToolRounds

Set maximum tool execution rounds (default: 5).

```go
func (b *Builder) WithMaxToolRounds(max int) *Builder
```

**Example:**
```go
builder.WithMaxToolRounds(10)  // Allow up to 10 tool calls
```

### Complete Tool Example

```go
// Define tool
weatherTool := agent.NewTool("get_weather", "Get current weather").
    AddParameter("location", "string", "City name", true).
    AddParameter("units", "string", "celsius/fahrenheit", false).
    WithHandler(func(args string) (string, error) {
        var params struct {
            Location string `json:"location"`
            Units    string `json:"units"`
        }
        json.Unmarshal([]byte(args), &params)
        // ... fetch weather ...
        return `{"temp": 25, "condition": "sunny"}`, nil
    })

// Use tool
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTools(weatherTool).
    WithAutoExecute(true).
    Ask(ctx, "What's the weather in Tokyo?")
```

---

## Structured Outputs

### WithJSONMode

Force model to return valid JSON.

```go
func (b *Builder) WithJSONMode() *Builder
```

**Example:**
```go
response, err := builder.
    WithJSONMode().
    WithSystem("Return JSON with 'answer' and 'confidence' fields").
    Ask(ctx, "What is 2+2?")
```

### WithJSONSchema

Enforce strict JSON schema.

```go
func (b *Builder) WithJSONSchema(name, description string, schema interface{}, strict bool) *Builder
```

**Parameters:**
- `name` - Schema name
- `description` - Schema description
- `schema` - JSON Schema (map[string]interface{})
- `strict` - Enable strict mode (recommended)

**Example:**
```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name": map[string]interface{}{
            "type":        "string",
            "description": "Person's name",
        },
        "age": map[string]interface{}{
            "type":        "integer",
            "description": "Person's age",
        },
    },
    "required":             []string{"name", "age"},
    "additionalProperties": false,
}

response, err := builder.
    WithJSONSchema("person_info", "Extract person information", schema, true).
    Ask(ctx, "John is 25 years old")

// Response: {"name":"John","age":25}
```

---

## Streaming

### OnStream

Set content streaming callback.

```go
func (b *Builder) OnStream(callback func(content string)) *Builder
```

**Example:**
```go
builder.OnStream(func(content string) {
    fmt.Print(content)  // Print each chunk
})
```

### OnToolCall

Set tool call callback (for streaming).

```go
func (b *Builder) OnToolCall(callback func(tool openai.FinishedChatCompletionToolCall)) *Builder
```

**Example:**
```go
builder.OnToolCall(func(tool openai.FinishedChatCompletionToolCall) {
    fmt.Printf("Tool called: %s\n", tool.Name)
})
```

### OnRefusal

Set refusal callback (when model refuses to answer).

```go
func (b *Builder) OnRefusal(callback func(refusal string)) *Builder
```

**Example:**
```go
builder.OnRefusal(func(refusal string) {
    log.Printf("Model refused: %s\n", refusal)
})
```

---

## Error Handling & Recovery

### WithTimeout

Set request timeout.

```go
func (b *Builder) WithTimeout(timeout time.Duration) *Builder
```

**Example:**
```go
builder.WithTimeout(30 * time.Second)
```

### WithRetry

Set maximum retry attempts for failed requests.

```go
func (b *Builder) WithRetry(maxRetries int) *Builder
```

**Example:**
```go
builder.WithRetry(3)  // Retry up to 3 times
```

### WithRetryDelay

Set fixed delay between retries.

```go
func (b *Builder) WithRetryDelay(delay time.Duration) *Builder
```

**Example:**
```go
builder.WithRetryDelay(2 * time.Second)  // Wait 2s between retries
```

### WithExponentialBackoff

Use exponential backoff for retries (1s, 2s, 4s, 8s, 16s...).

```go
func (b *Builder) WithExponentialBackoff() *Builder
```

**Example:**
```go
builder.WithRetry(5).WithExponentialBackoff()
```

### Error Type Checkers

```go
func IsAPIKeyError(err error) bool         // Check for API key errors
func IsRateLimitError(err error) bool      // Check for rate limit errors
func IsTimeoutError(err error) bool        // Check for timeout errors
func IsRefusalError(err error) bool        // Check for content refusals
func IsInvalidResponseError(err error) bool // Check for invalid responses
func IsMaxRetriesError(err error) bool     // Check if retries exhausted
func IsToolExecutionError(err error) bool  // Check for tool errors
```

**Example:**
```go
response, err := builder.Ask(ctx, "Hello")
if err != nil {
    if agent.IsTimeoutError(err) {
        log.Println("Request timed out")
    } else if agent.IsRateLimitError(err) {
        log.Println("Rate limit exceeded, please wait")
    } else if agent.IsAPIKeyError(err) {
        log.Println("Invalid API key")
    } else {
        log.Printf("Other error: %v", err)
    }
}
```

### Production Configuration Example

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithTemperature(0.7).
    WithMaxTokens(500).
    WithMemory().
    WithMaxHistory(20).
    WithTimeout(30 * time.Second).
    WithRetry(3).
    WithExponentialBackoff()

response, err := builder.Ask(ctx, prompt)
if err != nil {
    // Handle with proper error types
    if agent.IsTimeoutError(err) {
        return "", fmt.Errorf("request timed out after 30s")
    }
    if agent.IsRateLimitError(err) {
        time.Sleep(60 * time.Second)
        return builder.Ask(ctx, prompt)  // Retry after cooldown
    }
    return "", err
}
```

---

## Types

### Builder

Main type for building LLM requests with method chaining.

```go
type Builder struct {
    // (fields are private, use methods to configure)
}
```

### Message

Represents a conversation message.

```go
type Message struct {
    Role    string  // "system", "user", "assistant"
    Content string  // Message content
}
```

**Helper Functions:**
```go
func System(content string) Message     // Create system message
func User(content string) Message       // Create user message
func Assistant(content string) Message  // Create assistant message
```

**Example:**
```go
messages := []agent.Message{
    agent.System("You are helpful"),
    agent.User("Hello"),
    agent.Assistant("Hi! How can I help?"),
    agent.User("What's the weather?"),
}
builder.SetHistory(messages)
```

### Tool

Represents a callable tool/function.

```go
type Tool struct {
    Name        string
    Description string
    Parameters  []ToolParameter
    Handler     func(args string) (string, error)
}
```

### Provider

Provider type constant.

```go
type Provider string

const (
    ProviderOpenAI Provider = "openai"
    ProviderOllama Provider = "ollama"
)
```

---

## Error Types

### APIError

Custom error type with additional context.

```go
type APIError struct {
    Type       error   // Error type (ErrRateLimit, ErrTimeout, etc.)
    Message    string  // Error message
    StatusCode int     // HTTP status code (if applicable)
    Err        error   // Underlying error
}
```

### Error Constants

```go
var (
    ErrAPIKey           = errors.New("API key error")
    ErrRateLimit        = errors.New("rate limit exceeded")
    ErrTimeout          = errors.New("request timeout")
    ErrRefusal          = errors.New("content refusal")
    ErrInvalidResponse  = errors.New("invalid response")
    ErrMaxRetries       = errors.New("max retries exceeded")
    ErrToolExecution    = errors.New("tool execution failed")
)
```

---

## See Also

- **[BUILDER_API.md](../BUILDER_API.md)** - Complete guide with detailed examples
- **[examples/](../examples/)** - 8 example files with 34+ working examples
- **[README.md](../README.md)** - Main documentation
- **[openai-go Documentation](https://pkg.go.dev/github.com/openai/openai-go/v3)** - Underlying library

---

**Version:** 2.0.0 (Builder API)  
**Go Version:** 1.23.3+  
**License:** MIT
