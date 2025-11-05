# Go Deep Agent

A simple yet powerful LLM agent library for Go, supporting multiple providers (OpenAI, Ollama) with a unified, clean API.

Built with [openai-go v3.8.1](https://github.com/openai/openai-go).

## ✨ Features

- ✅ **Unified API** - One `Chat()` method for all use cases
- ✅ **Multi-Provider** - OpenAI, Ollama, and custom endpoints
- ✅ **Streaming** - Real-time response streaming
- ✅ **Conversation History** - Multi-turn conversations
- ✅ **Tool Calling** - Function calling support
- ✅ **Structured Outputs** - JSON Schema validation
- ✅ **Clean & Simple** - Minimal boilerplate code
- ✅ **Production Ready** - Error handling, context support

## 📦 Installation

```bash
go get github.com/taipm/go-deep-agent
```

## 🚀 Quick Start

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
    
    // Create agent
    ag, err := agent.NewAgent(agent.Config{
        Provider: agent.ProviderOpenAI,
        Model:    "gpt-4o-mini",
        APIKey:   "your-api-key",
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
}
```

## 📖 Usage Examples

### 1. Simple Chat (Non-Streaming)

```go
result, err := agent.Chat(ctx, "Explain quantum computing", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Content)
```

### 2. Streaming Response

```go
result, err := agent.Chat(ctx, "Write a haiku about code", &agent.ChatOptions{
    Stream: true,
    OnStream: func(chunk string) {
        fmt.Print(chunk)
    },
})
fmt.Println() // newline after stream
```

### 3. Conversation History

```go
import "github.com/openai/openai-go/v3"

result, err := agent.Chat(ctx, "What about supervised learning?", &agent.ChatOptions{
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.SystemMessage("You are a helpful AI tutor."),
        openai.UserMessage("What is machine learning?"),
        openai.AssistantMessage("Machine learning is..."),
        // New message will be appended automatically
    },
})
```

### 4. Tool Calling (Function Calling)

```go
tools := []openai.ChatCompletionToolUnionParam{
    openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
        Name:        "get_weather",
        Description: openai.String("Get current weather in a location"),
        Parameters: openai.FunctionParameters{
            "type": "object",
            "properties": map[string]any{
                "location": map[string]string{
                    "type":        "string",
                    "description": "City name, e.g. San Francisco",
                },
            },
            "required": []string{"location"},
        },
    }),
}

result, err := agent.Chat(ctx, "What's the weather in Hanoi?", &agent.ChatOptions{
    Tools: tools,
})

// Check if LLM wants to call a tool
if len(result.Completion.Choices) > 0 {
    toolCalls := result.Completion.Choices[0].Message.ToolCalls
    if len(toolCalls) > 0 {
        fmt.Printf("Tool: %s\n", toolCalls[0].Function.Name)
        fmt.Printf("Args: %s\n", toolCalls[0].Function.Arguments)
    }
}
```

### 5. Combining Multiple Features

```go
// Streaming + History + Tools
result, err := agent.Chat(ctx, "Check weather and tell me", &agent.ChatOptions{
    Messages: conversationHistory,
    Tools:    weatherTools,
    Stream:   true,
    OnStream: func(chunk string) {
        fmt.Print(chunk)
    },
})
```

### 6. Using Ollama (Local LLM)

```go
ollamaAgent, err := agent.NewAgent(agent.Config{
    Provider: agent.ProviderOllama,
    Model:    "qwen3:1.7b",
    BaseURL:  "http://localhost:11434/v1",
})

result, err := ollamaAgent.Chat(ctx, "Hello!", nil)
```

### 7. Advanced: Structured Outputs

For advanced use cases requiring full control (temperature, JSON schema, etc.):

```go
completion, err := agent.GetCompletion(ctx, openai.ChatCompletionNewParams{
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Extract name and age: John is 25 years old"),
    },
    Temperature: openai.Float64(0.1),
    ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
        OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
            JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
                Name: "person_info",
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

fmt.Println(completion.Choices[0].Message.Content)
// Output: {"name":"John","age":25}
```

## 🏗️ Project Structure

```
go-deep-agent/
├── agent/
│   ├── config.go         # Configuration & initialization
│   ├── agent.go          # Core agent implementation
│   └── README.md         # API documentation
├── examples/
│   ├── ollama_example.go # Ollama usage examples
│   └── README.md
├── main.go               # Complete examples
├── README.md             # Main documentation (you are here)
├── QUICK_REFERENCE.md    # Quick reference guide
├── ARCHITECTURE.md       # Design & architecture docs
├── CHANGELOG.md          # Version history
├── go.mod
└── go.sum
```

## 📚 API Reference

### Agent Configuration

```go
type Config struct {
    Provider Provider  // ProviderOpenAI or ProviderOllama
    Model    string    // Model name (e.g., "gpt-4o-mini", "qwen3:1.7b")
    APIKey   string    // API key (required for OpenAI)
    BaseURL  string    // Custom endpoint (for Ollama or custom)
}
```

### Chat Options

```go
type ChatOptions struct {
    Stream   bool                                     // Enable streaming
    OnStream func(string)                             // Stream callback
    Messages []openai.ChatCompletionMessageParamUnion // Conversation history
    Tools    []openai.ChatCompletionToolUnionParam    // Function calling tools
}
```

### Chat Result

```go
type ChatResult struct {
    Content    string                 // The response text
    Completion *openai.ChatCompletion // Full completion (for tool calls, metadata)
}
```

### Methods

```go
// Main method - handles all use cases
func (a *Agent) Chat(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error)

// Advanced method - for full control
func (a *Agent) GetCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
```

## 🛠️ Setup

### OpenAI

```bash
export OPENAI_API_KEY=your-api-key-here
```

### Ollama

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull qwen3:1.7b

# Run Ollama (exposes OpenAI-compatible API at :11434)
ollama serve
```

## 🏃 Running Examples

```bash
# Set API key
export OPENAI_API_KEY=your-key

# Run main examples
go run main.go

# Run Ollama examples (requires Ollama running)
go run examples/ollama_example.go
```

## 🎯 Design Philosophy

1. **Simplicity First** - One method for common use cases
2. **Flexibility** - Options pattern for advanced features
3. **Clean API** - Minimal boilerplate, clear intent
4. **Production Ready** - Proper error handling, context support
5. **Provider Agnostic** - Same API for OpenAI, Ollama, custom

## 📋 Requirements

- Go 1.23.3 or higher
- OpenAI API key (for OpenAI provider)
- Ollama running locally (for Ollama provider)

## 🤝 Contributing

Pull requests are welcome! For major changes, please open an issue first.

## 📄 License

MIT

## 📚 Documentation

- **[📋 PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md)** - 🌟 **Start here!** Complete project guide
- **[README.md](README.md)** - Main documentation (getting started, examples)
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Quick reference for common operations
- **[agent/README.md](agent/README.md)** - Complete API reference
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Design decisions and architecture
- **[CHANGELOG.md](CHANGELOG.md)** - Version history and migration guide
- **[TODO.md](TODO.md)** - Roadmap and implementation progress
- **[DESIGN_DECISIONS.md](DESIGN_DECISIONS.md)** - Design decisions log
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution guidelines
- **[examples/](examples/)** - Working code examples

## �🔗 Links

- [openai-go](https://github.com/openai/openai-go) - Official OpenAI Go library
- [Ollama](https://ollama.com) - Run LLMs locally
