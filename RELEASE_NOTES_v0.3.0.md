# Release v0.3.0: Builder API with Multimodal Support 🚀

**Major Release** - Complete rewrite with fluent Builder pattern

---

## 🎯 Highlights

- **🎨 Fluent Builder API** - Natural, readable method chaining
- **🖼️ Multimodal Support** - GPT-4 Vision with image analysis (NEW!)
- **🛠️ Tool Calling** - Auto-execution with type-safe definitions
- **📡 Advanced Streaming** - Real-time responses with callbacks
- **🧠 Smart Memory** - Automatic conversation history management
- **⚡ Error Recovery** - Retry with exponential backoff
- **📋 JSON Schema** - Structured outputs with validation

## 📊 Quality Metrics

- ✅ **242 tests** (all passing)
- ✅ **65.8% code coverage** (exceeded 60% goal)
- ✅ **13 benchmarks** (0.3-10 ns/op)
- ✅ **8 example files** with 41+ working examples
- ✅ **Full CI/CD pipeline**
- ✅ **Cross-platform** (Linux, macOS, Windows)

---

## 🆕 What's New in v0.3.0

### 1. Fluent Builder API

Natural, readable code with method chaining:

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithTemperature(0.7).
    WithMaxTokens(500).
    WithMemory().
    Ask(ctx, "Explain quantum computing")
```

### 2. Multimodal Support (Vision) 🖼️

Analyze images with GPT-4 Vision:

```go
// Image from URL
response, err := agent.NewOpenAI("gpt-4o", apiKey).
    WithImage("https://example.com/photo.jpg").
    Ask(ctx, "What's in this image?")

// Local image file
response, err := agent.NewOpenAI("gpt-4o", apiKey).
    WithImageFile("./chart.png", agent.ImageDetailHigh).
    Ask(ctx, "Extract data from this chart")

// Compare multiple images
response, err := agent.NewOpenAI("gpt-4o", apiKey).
    WithImageURL("image1.jpg", agent.ImageDetailLow).
    WithImageURL("image2.jpg", agent.ImageDetailHigh).
    Ask(ctx, "Compare these images")
```

**Supported:**
- Models: `gpt-4o`, `gpt-4o-mini`, `gpt-4-turbo`, `gpt-4-vision-preview`
- Formats: JPEG, PNG, GIF, WebP
- Sources: URL, local file, base64
- Detail levels: Auto, Low (512x512), High (2048x2048)

### 3. Tool Calling with Auto-Execution

Type-safe tool definitions with automatic execution:

```go
weatherTool := agent.NewTool("get_weather", "Get weather for a city").
    AddParameter("city", "string", "City name", true).
    WithHandler(func(args string) (string, error) {
        // Your implementation
        return "Sunny, 25°C", nil
    })

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTools(weatherTool).
    WithAutoExecute(true).
    WithMaxToolRounds(5).
    Ask(ctx, "What's the weather in Paris?")
```

### 4. Automatic Conversation Memory

No manual history management needed:

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).WithMemory()

builder.Ask(ctx, "My name is John")
builder.Ask(ctx, "What's my name?") // Remembers: "Your name is John"

// With max history (FIFO)
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().
    WithMaxHistory(10) // Keep last 10 messages
```

### 5. Error Handling & Recovery

Production-ready error handling:

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTimeout(30 * time.Second).
    WithRetry(3).
    WithExponentialBackoff(). // 1s, 2s, 4s, 8s delays
    Ask(ctx, "Your question")

if err != nil {
    if agent.IsTimeoutError(err) {
        // Handle timeout
    } else if agent.IsRateLimitError(err) {
        // Handle rate limit
    }
}
```

### 6. Streaming with Callbacks

Real-time streaming with type-safe callbacks:

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(content string) {
        fmt.Print(content) // Print as it arrives
    }).
    OnRefusal(func(refusal string) {
        log.Printf("Content refused: %s", refusal)
    }).
    Stream(ctx, "Write a haiku about code")
```

### 7. JSON Schema Validation

Structured outputs with schema validation:

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name": map[string]interface{}{"type": "string"},
        "age":  map[string]interface{}{"type": "number"},
    },
    "required": []string{"name", "age"},
}

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("person", "A person object", schema, true).
    Ask(ctx, "Generate a random person")
```

---

## 📦 Installation

```bash
go get github.com/taipm/go-deep-agent@v0.3.0
```

## 🚀 Quick Start

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

---

## 🔄 Migration from v0.2.0

**Before:**
```go
result, err := agent.Chat(ctx, "Hello", &agent.ChatOptions{
    Temperature: 0.7,
    Stream: true,
    OnStream: streamHandler,
})
```

**After:**
```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTemperature(0.7).
    OnStream(streamHandler).
    Stream(ctx, "Hello")
```

See [CHANGELOG.md](https://github.com/taipm/go-deep-agent/blob/main/CHANGELOG.md) for complete migration guide.

---

## 🎓 Examples

Check out 8 comprehensive example files in [`examples/`](https://github.com/taipm/go-deep-agent/tree/main/examples):

1. **builder_basic.go** - Basic usage patterns
2. **builder_streaming.go** - Streaming examples
3. **builder_tools.go** - Tool calling demos
4. **builder_json_schema.go** - Structured outputs
5. **builder_conversation.go** - Memory management
6. **builder_errors.go** - Error handling
7. **builder_multimodal.go** - Vision/image analysis ⭐ NEW
8. **ollama_example.go** - Local LLM usage

---

## 🏗️ Implementation

Completed 11 phases of development:

1. ✅ Core Builder (12 tests)
2. ✅ Advanced Parameters (9 tests)
3. ✅ Full Streaming (3 tests)
4. ✅ Tool Calling (19 tests)
5. ✅ JSON Schema (3 tests)
6. ✅ Testing & Documentation (55 tests)
7. ✅ Conversation Management (7 tests)
8. ✅ Error Handling & Recovery (14 tests)
9. ✅ Examples & Documentation
10. ✅ Testing & Quality (229 tests, CI/CD)
11. ✅ Multimodal Support (13 tests) ⭐ NEW

---

## 🔗 Links

- **Documentation**: [README.md](https://github.com/taipm/go-deep-agent/blob/main/README.md)
- **Changelog**: [CHANGELOG.md](https://github.com/taipm/go-deep-agent/blob/main/CHANGELOG.md)
- **Examples**: [examples/](https://github.com/taipm/go-deep-agent/tree/main/examples)
- **API Reference**: [pkg.go.dev](https://pkg.go.dev/github.com/taipm/go-deep-agent/agent)

---

## ⚠️ Breaking Changes

This is a **complete API rewrite**. The old API is deprecated. See [CHANGELOG.md](https://github.com/taipm/go-deep-agent/blob/main/CHANGELOG.md) for detailed migration guide.

**Key changes:**
- Builder pattern replaces functional options
- Fluent method chaining
- Cleaner, more intuitive API
- Better IDE autocomplete support

---

## 🙏 Credits

Built with:
- [openai-go v3.8.1](https://github.com/openai/openai-go) - Official OpenAI Go SDK
- Go 1.23.3

---

**Full Changelog**: https://github.com/taipm/go-deep-agent/blob/main/CHANGELOG.md

---

Made with ❤️ for the Go community
