# Go Deep Agent 🚀

A powerful yet simple LLM agent library for Go with a modern **Fluent Builder API**. Build AI applications with method chaining, automatic conversation memory, intelligent error handling, and seamless streaming support.

Built with [openai-go v3.8.1](https://github.com/openai/openai-go).

## ✨ Features

- 🎯 **Fluent Builder API** - Natural, readable method chaining
- 🤖 **Multi-Provider** - OpenAI, Ollama, and custom endpoints
- 🧠 **Conversation Memory** - Automatic history management with FIFO truncation
- 📡 **Streaming** - Real-time response streaming with callbacks
- 🛠️ **Tool Calling** - Auto-execution with type-safe function definitions
- 📋 **Structured Outputs** - JSON Schema with strict mode
- ⚡ **Error Recovery** - Smart retries with exponential backoff
- 🎛️ **Advanced Controls** - Temperature, top-p, tokens, penalties, seed
- 🧪 **Production Ready** - Timeouts, retries, comprehensive error handling
- 🖼️ **Multimodal** - Vision support for GPT-4 Vision (images via URL/file/base64)
- ✅ **Well Tested** - 242 tests, 62.6% coverage, 41+ working examples

## 📦 Installation

```bash
go get github.com/taipm/go-deep-agent
```

## 🚀 Quick Start

### Simple Chat - One Line

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).Ask(ctx, "What is Go?")
```

### With Streaming

```go
agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(content string) {
        fmt.Print(content)
    }).
    Stream(ctx, "Write a haiku about code")
```

### With Conversation Memory

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).WithMemory()

builder.Ask(ctx, "My name is John")
builder.Ask(ctx, "What's my name?")  // AI remembers: "Your name is John"
```

### Production-Ready Configuration

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithTemperature(0.7).
    WithMaxTokens(500).
    WithMemory().
    WithMaxHistory(10).
    WithTimeout(30 * time.Second).
    WithRetry(3).
    WithExponentialBackoff().
    Ask(ctx, "Explain Go concurrency")
```

## 📖 Builder API Examples

### 1. OpenAI with System Prompt

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithTemperature(0.7).
    Ask(ctx, "Explain quantum computing")
```

### 2. Streaming with Callbacks

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(content string) {
        fmt.Print(content)  // Print each chunk as it arrives
    }).
    Stream(ctx, "Write a haiku about AI")
```

### 3. Conversation Memory

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().           // Enable automatic memory
    WithMaxHistory(10)      // Keep last 10 messages (auto-truncate)

// First message
builder.Ask(ctx, "My name is John and I'm from Vietnam")

// AI remembers previous context
builder.Ask(ctx, "What's my name and where am I from?")
// Response: "Your name is John and you're from Vietnam"
```

### 4. Tool Calling with Auto-Execution

```go
weatherTool := agent.NewTool("get_weather", "Get current weather").
    AddParameter("location", "string", "City name", true).
    WithHandler(func(args string) (string, error) {
        return `{"temp": 25, "condition": "sunny"}`, nil
    })

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTools(weatherTool).
    WithAutoExecute(true).  // Automatically execute tool calls
    Ask(ctx, "What's the weather in Hanoi?")
```

### 5. Structured Outputs (JSON Schema)

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name": map[string]interface{}{"type": "string"},
        "age":  map[string]interface{}{"type": "integer"},
    },
    "required":             []string{"name", "age"},
    "additionalProperties": false,
}

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("person_info", "Extract person info", schema, true).
    Ask(ctx, "John is 25 years old")
// Response: {"name":"John","age":25}
```

### 6. Error Handling with Retry

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTimeout(10 * time.Second).
    WithRetry(3).
    WithExponentialBackoff().  // 1s, 2s, 4s, 8s delays
    Ask(ctx, "What is Go?")

if err != nil {
    if agent.IsTimeoutError(err) {
        log.Println("Request timed out")
    } else if agent.IsRateLimitError(err) {
        log.Println("Rate limit exceeded")
    }
}
```

### 7. Using Ollama (Local LLM)

```go
// Simple usage - default base URL is http://localhost:11434/v1
response, err := agent.NewOllama("qwen2.5:3b").
    Ask(ctx, "What is Go?")

// With configuration
response, err := agent.NewOllama("qwen2.5:3b").
    WithSystem("You are a concise assistant").
    WithTemperature(0.8).
    WithMemory().
    Ask(ctx, "Explain goroutines")
```

### 8. History Management

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).WithMemory()

builder.Ask(ctx, "I love Go programming")
builder.Ask(ctx, "What are channels?")

// Get conversation history
history := builder.GetHistory()
fmt.Printf("Messages: %d\n", len(history))

// Clear conversation (keeps system prompt)
builder.Clear()

// Save and restore sessions
savedHistory := builder.GetHistory()
// ... later ...
builder.SetHistory(savedHistory)
```

### 9. Multimodal - Vision (GPT-4 Vision)

```go
// Analyze image from URL
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithImage("https://example.com/photo.jpg").
    Ask(ctx, "What do you see in this image?")

// Compare multiple images with detail control
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithImageURL("https://example.com/image1.jpg", agent.ImageDetailLow).
    WithImageURL("https://example.com/image2.jpg", agent.ImageDetailHigh).
    Ask(ctx, "Compare these two images")

// Analyze local image file
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithImageFile("./chart.png", agent.ImageDetailHigh).
    Ask(ctx, "Extract data from this chart")

// Conversation with images
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).WithMemory()
builder.WithImage("https://example.com/photo.jpg").
    Ask(ctx, "What's in this image?")
builder.Ask(ctx, "What colors are prominent?") // Remembers the image
```

## 📖 Builder API Methods

### Core Methods

- `NewOpenAI(model, apiKey)` - Create OpenAI builder
- `NewOllama(model)` - Create Ollama builder (localhost:11434)
- `New(provider, model)` - Generic constructor
- `Ask(ctx, message)` - Send message, get response
- `Stream(ctx, message)` - Stream response with callbacks
- `StreamPrint(ctx, message)` - Stream and print to stdout

### Configuration

- `WithAPIKey(key)` - Set API key
- `WithBaseURL(url)` - Custom endpoint
- `WithModel(model)` - Change model
- `WithSystem(prompt)` - System prompt
- `WithTemperature(temp)` - Sampling temperature (0-2)
- `WithTopP(topP)` - Nucleus sampling (0-1)
- `WithMaxTokens(max)` - Maximum tokens to generate
- `WithPresencePenalty(penalty)` - Presence penalty (-2 to 2)
- `WithFrequencyPenalty(penalty)` - Frequency penalty (-2 to 2)
- `WithSeed(seed)` - For reproducible outputs
- `WithN(n)` - Number of completions to generate

### Conversation Management

- `WithMemory()` - Enable automatic conversation memory
- `WithMaxHistory(max)` - Limit messages (FIFO truncation)
- `GetHistory()` - Get conversation messages
- `SetHistory(messages)` - Restore conversation
- `Clear()` - Reset conversation (keeps system prompt)

### Tool Calling

- `WithTools(tools...)` - Register tools/functions
- `WithAutoExecute(enable)` - Auto-execute tool calls
- `WithMaxToolRounds(max)` - Max execution rounds (default 5)
- `OnToolCall(callback)` - Tool call callback

### Multimodal Support (Vision)

- `WithImage(url)` - Add image with auto detail level
- `WithImageURL(url, detail)` - Add image with specific detail (Low/High)
- `WithImageFile(filePath, detail)` - Add local image file
- `WithImageBase64(base64Data, mimeType, detail)` - Add base64-encoded image
- `ClearImages()` - Remove pending images

Supported models: `gpt-4o`, `gpt-4o-mini`, `gpt-4-turbo`, `gpt-4-vision-preview`

### Structured Outputs

- `WithJSONMode()` - Force JSON output
- `WithJSONSchema(name, desc, schema, strict)` - Structured JSON

### Streaming Callbacks

- `OnStream(callback)` - Content chunk callback
- `OnRefusal(callback)` - Refusal detection callback

### Error Handling & Recovery

- `WithTimeout(duration)` - Request timeout
- `WithRetry(maxRetries)` - Retry failed requests
- `WithRetryDelay(delay)` - Fixed delay between retries
- `WithExponentialBackoff()` - Use exponential backoff

### Error Type Checking

- `IsAPIKeyError(err)` - Check for API key errors
- `IsRateLimitError(err)` - Check for rate limits
- `IsTimeoutError(err)` - Check for timeouts
- `IsRefusalError(err)` - Check for content refusals
- `IsInvalidResponseError(err)` - Check for invalid responses
- `IsMaxRetriesError(err)` - Check if retries exhausted
- `IsToolExecutionError(err)` - Check for tool errors

## 🏗️ Project Structure

```plaintext
go-deep-agent/
├── agent/
│   ├── builder.go              # Fluent Builder API
│   ├── errors.go               # Custom error types
│   ├── tools.go                # Tool calling support
│   ├── *_test.go               # Comprehensive tests (76 tests)
│   └── README.md               # API documentation
├── examples/
│   ├── builder_basic.go        # Basic examples
│   ├── builder_streaming.go   # Streaming examples
│   ├── builder_tools.go        # Tool calling examples
│   ├── builder_json_schema.go # JSON Schema examples
│   ├── builder_conversation.go # Memory management
│   ├── builder_errors.go       # Error handling
│   ├── ollama_example.go       # Ollama examples
│   └── README.md
├── main.go                     # Quick start demo
├── README.md                   # You are here
├── BUILDER_API.md              # Complete Builder API guide
├── TODO.md                     # Development roadmap
└── go.mod
```

## � Quality Metrics

- ✅ **76 Tests** passing across all features
- ✅ **50.9% Coverage** with comprehensive test cases
- ✅ **8 Example Files** with 34+ working examples
- ✅ **Zero External Dependencies** (except openai-go)
- ✅ **Production Tested** with real OpenAI and Ollama APIs

## 🛠️ Setup & Usage

### OpenAI Setup

```bash
# Set your API key
export OPENAI_API_KEY=your-api-key-here

# Run examples
go run main.go
```

### Ollama Setup

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull qwen2.5:3b

# Run Ollama server
ollama serve

# Run Ollama examples
go run examples/ollama_example.go
```

## 🏃 Running Examples

```bash
# Basic examples
go run examples/builder_basic.go

# Streaming examples
go run examples/builder_streaming.go

# Tool calling examples
go run examples/builder_tools.go

# JSON Schema examples (requires OPENAI_API_KEY)
go run examples/builder_json_schema.go

# Conversation management
go run examples/builder_conversation.go

# Error handling examples
go run examples/builder_errors.go

# Ollama examples (requires Ollama running)
go run examples/ollama_example.go

# Quick demo with all features
go run main.go
```

## 🎯 Design Philosophy

1. **Fluent API** - Method chaining for natural, readable code
2. **Smart Defaults** - Works out of the box, customize as needed
3. **Memory Management** - Automatic conversation history with FIFO truncation
4. **Error Recovery** - Intelligent retries with exponential backoff
5. **Type Safety** - Leverages Go's type system for safety
6. **Zero Surprises** - Predictable behavior, no hidden magic
7. **Production Ready** - Timeouts, retries, comprehensive error handling

## 🧩 Advanced Use Cases

### Multi-Round Tool Execution

```go
calculateTool := agent.NewTool("calculate", "Perform math calculations").
    AddParameter("expression", "string", "Math expression", true).
    WithHandler(func(args string) (string, error) {
        // ... calculation logic ...
        return result, nil
    })

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTools(calculateTool).
    WithAutoExecute(true).
    WithMaxToolRounds(5).  // Allow multiple tool calls
    Ask(ctx, "Calculate (10 + 20) * 3, then add 50")
```

### Session Persistence

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).WithMemory()

// Have conversation
builder.Ask(ctx, "I'm learning Go")
builder.Ask(ctx, "Tell me about channels")

// Save session
session := builder.GetHistory()
saveToDatabase(session)

// Later: restore session
loadedSession := loadFromDatabase()
builder.SetHistory(loadedSession)
builder.Ask(ctx, "What were we talking about?")
```

### Production Error Handling

```go
func robustAsk(ctx context.Context, prompt string) (string, error) {
    builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithTimeout(30 * time.Second).
        WithRetry(3).
        WithExponentialBackoff()

    response, err := builder.Ask(ctx, prompt)
    if err != nil {
        if agent.IsTimeoutError(err) {
            return "", fmt.Errorf("request timed out after 30s")
        }
        if agent.IsRateLimitError(err) {
            time.Sleep(60 * time.Second) // Wait and retry
            return robustAsk(ctx, prompt)
        }
        return "", err
    }
    return response, nil
}
```

## 📋 Requirements

- **Go 1.23.3** or higher
- **OpenAI API key** (for OpenAI provider)
- **Ollama** running locally (for Ollama provider)

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new features
4. Ensure all tests pass: `go test ./...`
5. Submit a pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details

## 📚 Documentation

- **[README.md](README.md)** - Main documentation (you are here)
- **[BUILDER_API.md](BUILDER_API.md)** - Complete Builder API reference with examples
- **[agent/README.md](agent/README.md)** - Detailed API documentation
- **[examples/](examples/)** - 8 example files with 34+ working examples
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Quick reference guide
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Design decisions and architecture
- **[TODO.md](TODO.md)** - Roadmap and implementation progress (8/12 phases complete)

## 🔗 Links

- **GitHub**: [github.com/taipm/go-deep-agent](https://github.com/taipm/go-deep-agent)
- **openai-go**: [github.com/openai/openai-go](https://github.com/openai/openai-go) - Official OpenAI Go library
- **Ollama**: [ollama.com](https://ollama.com) - Run LLMs locally

---

<div align="center">

**Made with ❤️ for the Go community**

⭐ Star us on GitHub if you find this useful!

</div>