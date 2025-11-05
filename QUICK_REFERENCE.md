# Quick Reference

Quick reference guide for common operations.

## Installation

```bash
go get github.com/taipm/go-deep-agent
```

## Create Agent

### OpenAI
```go
agent, err := agent.NewAgent(agent.Config{
    Provider: agent.ProviderOpenAI,
    Model:    "gpt-4o-mini",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
})
```

### Ollama
```go
agent, err := agent.NewAgent(agent.Config{
    Provider: agent.ProviderOllama,
    Model:    "qwen3:1.7b",
    BaseURL:  "http://localhost:11434/v1",
})
```

## Chat Operations

### Simple Chat
```go
result, err := agent.Chat(ctx, "Hello", nil)
fmt.Println(result.Content)
```

### Streaming
```go
result, err := agent.Chat(ctx, "Tell a story", &agent.ChatOptions{
    Stream: true,
    OnStream: func(chunk string) {
        fmt.Print(chunk)
    },
})
```

### With History
```go
result, err := agent.Chat(ctx, "Follow-up question", &agent.ChatOptions{
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.SystemMessage("You are helpful."),
        openai.UserMessage("First question"),
        openai.AssistantMessage("First answer"),
    },
})
```

### With Tools
```go
result, err := agent.Chat(ctx, "What's the weather?", &agent.ChatOptions{
    Tools: []openai.ChatCompletionToolUnionParam{
        openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
            Name: "get_weather",
            Parameters: openai.FunctionParameters{
                "type": "object",
                "properties": map[string]any{
                    "location": map[string]string{"type": "string"},
                },
            },
        }),
    },
})

// Check tool calls
if len(result.Completion.Choices[0].Message.ToolCalls) > 0 {
    // Handle tool call
}
```

### Combined
```go
result, err := agent.Chat(ctx, "message", &agent.ChatOptions{
    Messages: history,
    Tools:    tools,
    Stream:   true,
    OnStream: callback,
})
```

## Advanced (Structured Outputs)

```go
completion, err := agent.GetCompletion(ctx, openai.ChatCompletionNewParams{
    Messages: messages,
    Temperature: openai.Float64(0.0),
    ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
        OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
            JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
                Name: "schema_name",
                Schema: openai.FunctionParameters{
                    "type": "object",
                    "properties": map[string]any{
                        "field": map[string]string{"type": "string"},
                    },
                },
            },
        },
    },
})
```

## Error Handling

```go
result, err := agent.Chat(ctx, "Hello", nil)
if err != nil {
    log.Printf("Error: %v", err)
    return
}
```

## Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := agent.Chat(ctx, "Long operation", nil)
```

## Common Patterns

### Building Conversation
```go
history := []openai.ChatCompletionMessageParamUnion{
    openai.SystemMessage("You are helpful."),
}

// First turn
result, _ := agent.Chat(ctx, "Question 1", &agent.ChatOptions{
    Messages: history,
})
history = append(history,
    openai.UserMessage("Question 1"),
    openai.AssistantMessage(result.Content),
)

// Second turn
result, _ = agent.Chat(ctx, "Question 2", &agent.ChatOptions{
    Messages: history,
})
```

### Concurrent Requests
```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        result, _ := agent.Chat(ctx, fmt.Sprintf("Q%d", id), nil)
        fmt.Println(result.Content)
    }(i)
}
wg.Wait()
```

## Full Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/openai/openai-go/v3"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    ctx := context.Background()
    
    // Create agent
    ag, err := agent.NewAgent(agent.Config{
        Provider: agent.ProviderOpenAI,
        Model:    "gpt-4o-mini",
        APIKey:   "your-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Simple chat
    result, err := ag.Chat(ctx, "What is Go?", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.Content)
    
    // Streaming
    fmt.Print("Streaming: ")
    result, err = ag.Chat(ctx, "Count to 5", &agent.ChatOptions{
        Stream: true,
        OnStream: func(s string) { fmt.Print(s) },
    })
    fmt.Println()
    
    // With history
    history := []openai.ChatCompletionMessageParamUnion{
        openai.SystemMessage("You are a helpful assistant."),
        openai.UserMessage("What is AI?"),
    }
    result, err = ag.Chat(ctx, "Tell me more", &agent.ChatOptions{
        Messages: history,
    })
    fmt.Println(result.Content)
}
```

## See Also

- [README.md](README.md) - Full documentation
- [agent/README.md](agent/README.md) - API reference
- [ARCHITECTURE.md](ARCHITECTURE.md) - Design documentation
- [examples/](examples/) - More examples
