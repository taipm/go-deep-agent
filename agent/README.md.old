# Agent Package Documentation

Package `agent` provides a unified interface for interacting with multiple LLM providers (OpenAI, Ollama) through a clean, simple API.

## Core Concepts

### Agent

The main type for interacting with LLMs. Create an agent using `NewAgent()` with appropriate configuration.

```go
agent, err := agent.NewAgent(agent.Config{
    Provider: agent.ProviderOpenAI,
    Model:    "gpt-4o-mini",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
})
```

## Types

### Config

Configuration for creating a new agent.

```go
type Config struct {
    Provider Provider  // LLM provider (ProviderOpenAI or ProviderOllama)
    Model    string    // Model name
    APIKey   string    // API key (required for OpenAI)
    BaseURL  string    // Custom endpoint (optional, default for Ollama: http://localhost:11434/v1)
}
```

**Supported Providers:**
- `ProviderOpenAI` - OpenAI models (gpt-4o, gpt-4o-mini, etc.)
- `ProviderOllama` - Local models via Ollama (llama3.2, qwen3, etc.)

### ChatOptions

Options for configuring chat behavior.

```go
type ChatOptions struct {
    Stream   bool                                     // Enable streaming responses
    OnStream func(string)                             // Callback function for streaming chunks
    Messages []openai.ChatCompletionMessageParamUnion // Conversation history
    Tools    []openai.ChatCompletionToolUnionParam    // Tools for function calling
}
```

**Fields:**
- `Stream`: If `true`, responses are streamed in real-time
- `OnStream`: Callback function called for each streaming chunk
- `Messages`: Full conversation history. If provided, the `message` parameter is appended
- `Tools`: Array of tools the LLM can call (function calling)

### ChatResult

Result returned from a chat operation.

```go
type ChatResult struct {
    Content    string                 // The text response
    Completion *openai.ChatCompletion // Full completion object (for metadata, tool calls)
}
```

**Fields:**
- `Content`: The main text response from the LLM
- `Completion`: Full OpenAI completion object, useful for:
  - Tool calls information
  - Token usage stats
  - Model information
  - Finish reason

## Methods

### NewAgent

Creates a new agent with the specified configuration.

```go
func NewAgent(config Config) (*Agent, error)
```

**Parameters:**
- `config`: Agent configuration

**Returns:**
- `*Agent`: Initialized agent
- `error`: Error if configuration is invalid

**Example:**
```go
agent, err := agent.NewAgent(agent.Config{
    Provider: agent.ProviderOllama,
    Model:    "qwen3:1.7b",
    BaseURL:  "http://localhost:11434/v1",
})
```

### Chat

Main method for all chat operations. Supports simple chat, streaming, conversation history, and tool calling.

```go
func (a *Agent) Chat(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `message`: User message. Can be empty if `opts.Messages` contains the full conversation
- `opts`: Optional configuration. Pass `nil` for simple chat

**Returns:**
- `*ChatResult`: Chat result containing content and full completion
- `error`: Error if the operation fails

**Use Cases:**

#### 1. Simple Chat
```go
result, err := agent.Chat(ctx, "Hello!", nil)
fmt.Println(result.Content)
```

#### 2. Streaming
```go
result, err := agent.Chat(ctx, "Tell me a story", &agent.ChatOptions{
    Stream: true,
    OnStream: func(chunk string) {
        fmt.Print(chunk)
    },
})
```

#### 3. Conversation History
```go
result, err := agent.Chat(ctx, "What about neural networks?", &agent.ChatOptions{
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.SystemMessage("You are an AI expert."),
        openai.UserMessage("What is deep learning?"),
        openai.AssistantMessage("Deep learning is..."),
    },
})
```

#### 4. Tool Calling
```go
result, err := agent.Chat(ctx, "What's the weather?", &agent.ChatOptions{
    Tools: []openai.ChatCompletionToolUnionParam{
        openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
            Name:        "get_weather",
            Description: openai.String("Get current weather"),
            Parameters: openai.FunctionParameters{
                "type": "object",
                "properties": map[string]any{
                    "location": map[string]string{"type": "string"},
                },
                "required": []string{"location"},
            },
        }),
    },
})

// Check for tool calls
if len(result.Completion.Choices[0].Message.ToolCalls) > 0 {
    toolCall := result.Completion.Choices[0].Message.ToolCalls[0]
    fmt.Printf("Tool: %s, Args: %s\n", toolCall.Function.Name, toolCall.Function.Arguments)
}
```

#### 5. Combined Features
```go
result, err := agent.Chat(ctx, "next message", &agent.ChatOptions{
    Messages: history,
    Tools:    tools,
    Stream:   true,
    OnStream: func(s string) { fmt.Print(s) },
})
```

### GetCompletion

Low-level method for advanced use cases requiring full control over OpenAI API parameters.

```go
func (a *Agent) GetCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
```

**Parameters:**
- `ctx`: Context
- `params`: Full OpenAI chat completion parameters

**Returns:**
- `*openai.ChatCompletion`: Raw OpenAI completion
- `error`: Error if operation fails

**When to use:**
- Structured outputs with JSON Schema
- Fine-tuned parameter control (temperature, top_p, etc.)
- Multiple choices (N parameter)
- Token probability analysis (logprobs)
- Custom parameters not exposed by `Chat()`

**Example: Structured Output**
```go
completion, err := agent.GetCompletion(ctx, openai.ChatCompletionNewParams{
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Extract: Alice is 30 years old"),
    },
    Temperature: openai.Float64(0.0),
    ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
        OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
            JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
                Name: "person",
                Schema: openai.FunctionParameters{
                    "type": "object",
                    "properties": map[string]any{
                        "name": map[string]string{"type": "string"},
                        "age":  map[string]string{"type": "integer"},
                    },
                    "required": []string{"name", "age"},
                },
                Strict: openai.Bool(true),
            },
        },
    },
})
```

## Error Handling

All methods return errors for:
- Invalid configuration
- API errors
- Network issues
- Context cancellation

Always check errors:

```go
result, err := agent.Chat(ctx, "Hello", nil)
if err != nil {
    log.Printf("Chat error: %v", err)
    return
}
```

## Context Usage

Use context for:
- Cancellation
- Timeouts
- Request scoping

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := agent.Chat(ctx, "Long task...", nil)
```

## Best Practices

### 1. Reuse Agents
Create agent once, reuse for multiple requests:

```go
agent, _ := agent.NewAgent(config)

// Multiple requests
result1, _ := agent.Chat(ctx, "Question 1", nil)
result2, _ := agent.Chat(ctx, "Question 2", nil)
```

### 2. Handle Streaming Properly
Always handle streaming errors:

```go
result, err := agent.Chat(ctx, "Long story", &agent.ChatOptions{
    Stream: true,
    OnStream: func(chunk string) {
        fmt.Print(chunk) // Print as it streams
    },
})
if err != nil {
    log.Printf("Streaming error: %v", err)
}
fmt.Println() // Newline after stream
```

### 3. Build Conversation History
Maintain conversation state:

```go
history := []openai.ChatCompletionMessageParamUnion{
    openai.SystemMessage("You are a helpful assistant."),
}

// First exchange
result, _ := agent.Chat(ctx, "What is AI?", &agent.ChatOptions{
    Messages: history,
})
history = append(history, 
    openai.UserMessage("What is AI?"),
    openai.AssistantMessage(result.Content),
)

// Next exchange
result, _ = agent.Chat(ctx, "Tell me more", &agent.ChatOptions{
    Messages: history,
})
```

### 4. Check Tool Calls
Always verify tool calls:

```go
result, _ := agent.Chat(ctx, "Weather?", &agent.ChatOptions{Tools: tools})

if len(result.Completion.Choices) > 0 {
    toolCalls := result.Completion.Choices[0].Message.ToolCalls
    if len(toolCalls) > 0 {
        // Handle tool call
        for _, tc := range toolCalls {
            fmt.Printf("Calling: %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
        }
    } else {
        // Normal text response
        fmt.Println(result.Content)
    }
}
```

## Examples

See:
- `/examples/ollama_example.go` - Ollama usage examples
- `/main.go` - Complete OpenAI examples with all features

## Provider-Specific Notes

### OpenAI
- Requires valid API key
- Supports all features (streaming, tools, structured outputs)
- Default models: gpt-4o, gpt-4o-mini, gpt-3.5-turbo

### Ollama
- Requires Ollama running locally
- Default endpoint: `http://localhost:11434/v1`
- OpenAI-compatible API
- Supports streaming and basic chat
- Tool calling support depends on model

## Troubleshooting

### "API key is required for OpenAI"
Set your OpenAI API key:
```bash
export OPENAI_API_KEY=sk-...
```

### "Connection refused" with Ollama
Ensure Ollama is running:
```bash
ollama serve
```

### "Model is required"
Specify model in config:
```go
Config{
    Provider: agent.ProviderOllama,
    Model:    "qwen3:1.7b", // Required
}
```

## Performance Tips

1. **Reuse agents** - Avoid creating new agents for each request
2. **Use streaming** - For better UX in interactive applications
3. **Set timeouts** - Use context with timeout for long operations
4. **Batch requests** - Process multiple prompts efficiently
5. **Monitor tokens** - Check `result.Completion.Usage` for token counts

## Thread Safety

Agents are safe for concurrent use. You can use the same agent from multiple goroutines:

```go
agent, _ := agent.NewAgent(config)

var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        result, _ := agent.Chat(ctx, fmt.Sprintf("Question %d", id), nil)
        fmt.Printf("Response %d: %s\n", id, result.Content)
    }(i)
}
wg.Wait()
```
