# Go Deep Agent - Builder API Quick Start

This guide provides a quick reference for the **Builder API** - the modern, fluent interface for working with go-deep-agent.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Features](#core-features)
- [Conversation Management](#conversation-management)
- [Error Handling & Recovery](#error-handling--recovery)
- [Advanced Parameters](#advanced-parameters)
- [Streaming](#streaming)
- [Tool Calling](#tool-calling)
- [JSON Schema & Structured Outputs](#json-schema--structured-outputs)
- [Complete Examples](#complete-examples)

## Installation

```bash
go get github.com/taipm/go-deep-agent
```

## Quick Start

### Simple Chat

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
    
    response, err := agent.NewOpenAI("gpt-4o-mini", "your-api-key").
        Ask(ctx, "What is Go?")
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response)
}
```

## Core Features

### 1. Basic Configuration

```go
// OpenAI
builder := agent.NewOpenAI("gpt-4o-mini", apiKey)

// Ollama (local)
builder := agent.NewOllama("qwen2.5:7b")

// Custom endpoint
builder := agent.NewOllama("qwen2.5:7b").
    WithBaseURL("http://localhost:11434/v1")
```

### 2. System Prompts

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful coding assistant").
    Ask(ctx, "Explain quicksort")
```

### 3. Conversation Memory

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory(true) // Auto-save conversation history

// First message
response1, _ := builder.Ask(ctx, "My name is John")

// Second message - remembers context
response2, _ := builder.Ask(ctx, "What's my name?")
// Response: "Your name is John"
```

### 4. Custom Message History

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMessages([]agent.Message{
        agent.System("You are a helpful assistant"),
        agent.User("What is machine learning?"),
        agent.Assistant("Machine learning is..."),
    }).
    Ask(ctx, "Tell me more about neural networks")
```

## Conversation Management

### Get History

Retrieve current conversation history:

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory()

builder.Ask(ctx, "My name is Alice")
builder.Ask(ctx, "I love Go programming")

// Get conversation history
history := builder.GetHistory()
fmt.Printf("Conversation has %d messages\n", len(history))

for _, msg := range history {
    fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
}
```

### Set History

Restore a previous conversation:

```go
// Save conversation
savedHistory := builder.GetHistory()

// Later, restore it
newBuilder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    SetHistory(savedHistory).
    WithMemory()

// Continue from where you left off
response, _ := newBuilder.Ask(ctx, "Where were we?")
```

### Clear Conversation

Reset conversation while keeping system prompt:

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithMemory()

builder.Ask(ctx, "My favorite color is blue")
builder.Clear() // Reset conversation

// AI won't remember the previous message
response, _ := builder.Ask(ctx, "What's my favorite color?")
```

### Limit History (Context Window Management)

Automatically truncate old messages:

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().
    WithMaxHistory(10) // Keep only last 10 messages

// Older messages automatically removed when limit exceeded
for i := 0; i < 20; i++ {
    builder.Ask(ctx, fmt.Sprintf("Message %d", i))
}

// Only last 10 messages kept
history := builder.GetHistory()
fmt.Printf("History size: %d\n", len(history)) // Prints: 10
```

## Error Handling & Recovery

### Timeout

Set request timeout to prevent hanging:

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTimeout(30 * time.Second) // 30 second timeout

response, err := builder.Ask(ctx, "Your question")
if err != nil {
    if agent.IsTimeoutError(err) {
        fmt.Println("Request timed out!")
    }
}
```

### Retry with Fixed Delay

Automatically retry failed requests:

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRetry(3).                    // Retry up to 3 times
    WithRetryDelay(2 * time.Second)  // Wait 2 seconds between retries

response, err := builder.Ask(ctx, "Your question")
if err != nil {
    if agent.IsMaxRetriesError(err) {
        fmt.Println("Failed after all retries")
    }
}
```

### Exponential Backoff

Use exponential backoff for better retry behavior:

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRetry(5).                    // Retry up to 5 times
    WithRetryDelay(time.Second).     // Base delay: 1 second
    WithExponentialBackoff()         // Delays: 1s, 2s, 4s, 8s, 16s
```

### Error Type Checking

Check specific error types for proper handling:

```go
response, err := builder.Ask(ctx, "Your question")
if err != nil {
    switch {
    case agent.IsAPIKeyError(err):
        fmt.Println("Invalid API key")
        // Fatal error - fix configuration
    case agent.IsRateLimitError(err):
        fmt.Println("Rate limited")
        // Wait and retry
    case agent.IsTimeoutError(err):
        fmt.Println("Request timed out")
        // Reduce request complexity or increase timeout
    case agent.IsRefusalError(err):
        fmt.Println("Content refused")
        // Modify prompt to comply with policies
    case agent.IsMaxRetriesError(err):
        fmt.Println("Failed after retries")
        // Service may be down
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

### Production-Ready Configuration

Combine all error handling features:

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTimeout(30 * time.Second).       // Request timeout
    WithRetry(3).                         // Retry attempts
    WithRetryDelay(2 * time.Second).     // Base delay
    WithExponentialBackoff().             // Smart backoff
    WithMemory().                         // Conversation memory
    WithMaxHistory(20)                    // Limit history

// Robust request handling
response, err := builder.Ask(ctx, "Your question")
if err != nil {
    // Handle error appropriately
    log.Printf("Request failed: %v", err)
    return
}
```

## Advanced Parameters

### Temperature & Sampling

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTemperature(0.7).   // Creativity (0-2)
    WithTopP(0.9).          // Nucleus sampling
    WithMaxTokens(1000).    // Max response length
    Ask(ctx, "Write a creative story")
```

### Penalties & Seed

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithPresencePenalty(0.5).    // Encourage new topics
    WithFrequencyPenalty(0.3).   // Reduce repetition
    WithSeed(12345)              // Reproducible outputs
```

### Logprobs

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogprobs(true).
    WithTopLogprobs(5) // Get top 5 token probabilities
```

### Multiple Choices

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMultipleChoices(3) // Generate 3 responses
```

## Streaming

### Basic Streaming

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(chunk string) {
        fmt.Print(chunk) // Print each chunk as it arrives
    }).
    Stream(ctx, "Write a poem")
```

### Stream to Terminal

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    StreamPrint(ctx, "Explain quantum computing")
// Automatically prints to stdout with typing effect
```

### Streaming with Callbacks

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(chunk string) {
        // Handle each chunk
    }).
    OnToolCall(func(tool openai.FinishedChatCompletionToolCall) {
        // Handle tool calls
    }).
    OnRefusal(func(refusal string) {
        // Handle refusals
    }).
    Stream(ctx, "Help me with this task")
```

## Tool Calling

### Define a Tool

```go
weatherTool := agent.NewTool("get_weather", "Get current weather").
    AddParameter("location", "string", "City name", true).
    AddParameter("units", "string", "celsius or fahrenheit", false).
    WithHandler(func(args string) (string, error) {
        var params struct {
            Location string `json:"location"`
            Units    string `json:"units"`
        }
        json.Unmarshal([]byte(args), &params)
        
        // Your logic here
        return fmt.Sprintf("Weather in %s: Sunny, 25°C", params.Location), nil
    })
```

### Use Tools (Manual)

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTool(weatherTool).
    Ask(ctx, "What's the weather in Paris?")

// Model may return a tool call - you handle execution
```

### Auto-Execute Tools

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTool(weatherTool).
    WithAutoExecute(true).     // Automatically execute tools
    WithMaxToolRounds(5).      // Max execution rounds
    Ask(ctx, "What's the weather in Paris?")

// Tool is automatically called and result returned
```

### Multiple Tools

```go
calcTool := agent.NewTool("calculate", "Do math").
    AddParameter("operation", "string", "add/multiply/etc", true).
    AddParameter("a", "number", "First number", true).
    AddParameter("b", "number", "Second number", true).
    WithHandler(calcHandler)

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTools(weatherTool, calcTool, timeTool).
    WithAutoExecute(true).
    Ask(ctx, "What's the weather and what's 123 * 456?")
```

## JSON Schema & Structured Outputs

### JSON Mode (Simple)

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONMode().
    WithSystem("Respond with JSON containing 'answer' and 'confidence' fields").
    Ask(ctx, "What is the capital of France?")

// Response: {"answer":"Paris","confidence":0.99}
```

### JSON Schema (Strict)

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name": map[string]interface{}{
            "type": "string",
            "description": "Person's name",
        },
        "age": map[string]interface{}{
            "type": "integer",
            "description": "Person's age",
        },
        "skills": map[string]interface{}{
            "type": "array",
            "items": map[string]interface{}{
                "type": "string",
            },
        },
    },
    "required": []string{"name", "age", "skills"},
    "additionalProperties": false,
}

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("person_info", "Person information", schema, true).
    Ask(ctx, "Extract: John is 32 and knows Go, Python, AWS")

// Response: {"name":"John","age":32,"skills":["Go","Python","AWS"]}
```

## Complete Examples

### Example 1: Creative Writing with Advanced Parameters

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a creative writer").
    WithTemperature(1.2).        // High creativity
    WithTopP(0.95).
    WithMaxTokens(500).
    WithPresencePenalty(0.6).    // Encourage diverse topics
    Ask(ctx, "Write a short sci-fi story")
```

### Example 2: Code Assistant with Memory

```go
assistant := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a Go programming expert").
    WithMemory(true).
    WithTemperature(0.3) // Lower for more focused responses

// Multi-turn conversation
response1, _ := assistant.Ask(ctx, "How do I create a struct in Go?")
response2, _ := assistant.Ask(ctx, "Now show me how to add methods to it")
response3, _ := assistant.Ask(ctx, "What about interfaces?")
```

### Example 3: Data Extraction with JSON Schema

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "product": map[string]interface{}{"type": "string"},
        "price": map[string]interface{}{"type": "number"},
        "currency": map[string]interface{}{"type": "string"},
        "in_stock": map[string]interface{}{"type": "boolean"},
    },
    "required": []string{"product", "price", "currency", "in_stock"},
    "additionalProperties": false,
}

text := "The MacBook Pro costs $1299 and is currently in stock"

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("product_info", "Product details", schema, true).
    Ask(ctx, fmt.Sprintf("Extract product info: %s", text))

// Parse into struct
var product ProductInfo
json.Unmarshal([]byte(response), &product)
```

### Example 4: AI Agent with Tools

```go
// Define tools
searchTool := agent.NewTool("search", "Search the web").
    AddParameter("query", "string", "Search query", true).
    WithHandler(searchHandler)

calcTool := agent.NewTool("calculate", "Do math").
    AddParameter("expression", "string", "Math expression", true).
    WithHandler(calcHandler)

// Create agent
aiAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful AI assistant with access to tools").
    WithTools(searchTool, calcTool).
    WithAutoExecute(true).
    WithMemory(true)

// Use the agent
response1, _ := aiAgent.Ask(ctx, "Search for Go programming tutorials")
response2, _ := aiAgent.Ask(ctx, "Calculate 15% of 250")
response3, _ := aiAgent.Ask(ctx, "What did I just search for?") // Uses memory
```

### Example 5: Streaming with Progress

```go
var fullResponse string

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a technical writer").
    OnStream(func(chunk string) {
        fmt.Print(chunk)
        fullResponse += chunk
    }).
    Stream(ctx, "Explain Docker containers")

if err != nil {
    log.Fatal(err)
}

// fullResponse contains the complete text
fmt.Printf("\n\nTotal words: %d\n", len(strings.Fields(fullResponse)))
```

## Method Chaining

The Builder API is designed for fluent method chaining:

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithMemory(true).
    WithTemperature(0.7).
    WithMaxTokens(1000).
    WithTool(myTool).
    WithAutoExecute(true).
    WithJSONMode().
    OnStream(func(chunk string) {
        fmt.Print(chunk)
    }).
    Stream(ctx, "Help me with this task")
```

## Best Practices

1. **Use System Prompts** - Set clear instructions for the model's behavior
2. **Enable Memory** - For multi-turn conversations
3. **Lower Temperature** - For factual/deterministic tasks (0.0-0.3)
4. **Higher Temperature** - For creative tasks (0.7-1.5)
5. **JSON Schema Strict Mode** - Always use `strict=true` for reliable structured outputs
6. **Tool Auto-Execute** - Enable for autonomous agent behavior
7. **Streaming** - Use for better UX in interactive applications

## Running Examples

Check out the `examples/` directory:

```bash
cd examples

# Basic Builder usage
go run builder_basic.go

# Advanced parameters
go run builder_advanced.go

# Streaming
go run builder_streaming.go

# Tool calling
go run builder_tools.go

# JSON Schema
go run builder_json_schema.go
```

## More Documentation

- [Full API Reference](./agent/README.md)
- [JSON Schema Guide](./docs/JSON_SCHEMA.md)
- [Examples Directory](./examples/)
- [Architecture Docs](./ARCHITECTURE.md)

## License

MIT
