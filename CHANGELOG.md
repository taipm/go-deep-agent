# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2025-11-07 🚀 Major Release: Builder API Rewrite

### 🎯 Complete Rewrite with Fluent Builder Pattern

This is a **major rewrite** introducing a fluent Builder API that maximizes code readability and developer experience. The library is now production-ready with comprehensive testing and CI/CD.

### ✨ Added - Core Features

- **🎯 Fluent Builder API** - Natural method chaining for all operations
  - `NewOpenAI(model, apiKey)` - OpenAI provider
  - `NewOllama(model)` - Ollama provider (local LLMs)
  - `New(provider, model)` - Generic constructor

- **🧠 Automatic Conversation Memory**
  - `WithMemory()` - Enable automatic history tracking
  - `WithMaxHistory(n)` - FIFO truncation for long conversations
  - `GetHistory()` / `SetHistory()` - Session persistence
  - `Clear()` - Reset conversation

- **📡 Enhanced Streaming**
  - `Stream(ctx, message)` - Stream responses
  - `StreamPrint(ctx, message)` - Stream and print
  - `OnStream(callback)` - Custom stream handlers
  - `OnRefusal(callback)` - Content refusal detection

- **🛠️ Tool Calling with Auto-Execution**
  - `WithTools(tools...)` - Register multiple tools
  - `WithAutoExecute(true)` - Automatic tool call execution
  - `WithMaxToolRounds(n)` - Control execution loops
  - `OnToolCall(callback)` - Tool call monitoring
  - Type-safe tool definitions with `NewTool()`

- **📋 Structured Outputs (JSON Schema)**
  - `WithJSONMode()` - Force JSON responses
  - `WithJSONSchema(name, desc, schema, strict)` - Schema validation
  - Strict mode support for guaranteed schema compliance

- **🖼️ Multimodal Support (Vision)** ⭐ NEW
  - `WithImage(url)` - Add images from URLs
  - `WithImageURL(url, detail)` - Control detail level (Low/High/Auto)
  - `WithImageFile(filePath, detail)` - Load local images
  - `WithImageBase64(base64Data, mimeType, detail)` - Base64 images
  - `ClearImages()` - Remove pending images
  - Supports: GPT-4o, GPT-4o-mini, GPT-4 Turbo, GPT-4 Vision
  - Image formats: JPEG, PNG, GIF, WebP

- **⚡ Error Handling & Recovery**
  - `WithTimeout(duration)` - Request timeouts
  - `WithRetry(maxRetries)` - Automatic retries
  - `WithRetryDelay(duration)` - Fixed retry delay
  - `WithExponentialBackoff()` - Smart retry strategy (1s, 2s, 4s, 8s...)
  - Error type checkers: `IsTimeoutError()`, `IsRateLimitError()`, `IsAPIKeyError()`, etc.

- **🎛️ Advanced Parameters**
  - `WithSystem(prompt)` - System prompts
  - `WithTemperature(t)` - Creativity control (0-2)
  - `WithTopP(p)` - Nucleus sampling (0-1)
  - `WithMaxTokens(n)` - Output length limits
  - `WithPresencePenalty(p)` - Topic diversity (-2 to 2)
  - `WithFrequencyPenalty(p)` - Repetition control (-2 to 2)
  - `WithSeed(n)` - Reproducible outputs
  - `WithN(n)` - Multiple completions

### 📊 Quality Metrics

- ✅ **242 tests** (all passing)
- ✅ **65.8% code coverage** (exceeded 60% goal)
- ✅ **13 benchmarks** (0.3-10 ns/op)
- ✅ **8 example files** with 41+ working examples
- ✅ **Full CI/CD pipeline** (test, lint, build, security scan)
- ✅ **Multi-version Go support** (1.21, 1.22, 1.23)
- ✅ **Cross-platform builds** (Linux, macOS, Windows; amd64, arm64)

### 🔄 Changed - Breaking Changes

- **BREAKING**: Complete API redesign
  - Old: `agent.Chat(ctx, message, stream)` 
  - New: `agent.NewOpenAI(model, key).Ask(ctx, message)`
  
- **BREAKING**: Builder pattern replaces functional options
  - Fluent method chaining instead of variadic options
  - More discoverable API with IDE autocomplete

- **BREAKING**: Package structure reorganized
  - `agent.Builder` is now the main entry point
  - All configuration via method chaining
  - Cleaner imports: just `github.com/taipm/go-deep-agent/agent`

### 📚 Documentation

- **README.md** - Complete rewrite with 9 usage examples
- **TODO.md** - 11 phases documented (11/12 complete)
- **examples/** - 8 comprehensive example files:
  - `builder_basic.go` - Basic usage patterns
  - `builder_streaming.go` - Streaming examples
  - `builder_tools.go` - Tool calling demos
  - `builder_json_schema.go` - Structured outputs
  - `builder_conversation.go` - Memory management
  - `builder_errors.go` - Error handling
  - `builder_multimodal.go` - Vision/image analysis ⭐ NEW
  - `ollama_example.go` - Local LLM usage

### 🚀 Implementation Phases

All 11 phases completed:

1. ✅ **Phase 1**: Core Builder (12 tests)
2. ✅ **Phase 2**: Advanced Parameters (9 tests)
3. ✅ **Phase 3**: Full Streaming (3 tests)
4. ✅ **Phase 4**: Tool Calling (19 tests)
5. ✅ **Phase 5**: JSON Schema (3 tests)
6. ✅ **Phase 6**: Testing & Documentation (55 tests, 39.2% coverage)
7. ✅ **Phase 7**: Conversation Management (7 tests, 6 examples)
8. ✅ **Phase 8**: Error Handling & Recovery (14 tests, 6 examples)
9. ✅ **Phase 9**: Examples & Documentation (SKIPPED - already complete)
10. ✅ **Phase 10**: Testing & Quality (229 tests, 62.6% coverage, CI/CD)
11. ✅ **Phase 11**: Multimodal Support (13 tests, 7 examples)

### 🎓 Migration Guide from v0.2.0

See detailed migration examples in [Migration Guide](#migration-guide-1) below.

**Quick comparison:**
```go
// OLD v0.2.0
result, err := agent.Chat(ctx, "Hello", &agent.ChatOptions{
    Stream: true,
    OnStream: func(s string) { fmt.Print(s) },
})

// NEW v0.3.0
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(s string) { fmt.Print(s) }).
    Stream(ctx, "Hello")
```

## [0.2.0] - Previous Release

### Added
- Comprehensive documentation in README.md
- API documentation in agent/README.md
- Architecture documentation in ARCHITECTURE.md
- Examples in examples/ directory

### Changed
- **BREAKING**: Unified `Chat()`, `ChatStream()`, `ChatWithHistory()`, and `ChatWithToolCalls()` into single `Chat()` method with options pattern
- **BREAKING**: `Chat()` now returns `*ChatResult` instead of `string`
- Refactored package structure:
  - Split agent package into `config.go` (configuration) and `agent.go` (implementation)
  - Total: 202 lines across 2 files (down from 165 lines in single file)

### Removed
- Removed `ChatStream()` method (merged into `Chat()`)
- Removed `ChatWithHistory()` method (merged into `Chat()`)
- Removed `ChatWithToolCalls()` method (merged into `Chat()`)

## [0.1.0] - Initial Release

### Added
- Basic agent implementation supporting OpenAI and Ollama
- Multiple chat methods:
  - `Chat()` - Simple chat completion
  - `ChatStream()` - Streaming responses
  - `ChatWithHistory()` - Conversation history support
  - `ChatWithToolCalls()` - Function calling
- `GetCompletion()` for advanced use cases
- Support for structured outputs via JSON Schema
- OpenAI-compatible API for Ollama
- Example implementations

### Implementation Details
- Built on openai-go v3.8.1
- Provider abstraction layer
- ChatCompletionAccumulator for streaming
- Context support for cancellation and timeouts

---

## Migration Guide

### Migrating from v0.2.0 to v0.3.0 (Builder API)

v0.3.0 introduces a complete rewrite with fluent Builder pattern. The migration is straightforward once you understand the pattern.

#### Simple Chat

**Before (v0.2.0):**
```go
result, err := agent.Chat(ctx, "Hello", nil)
fmt.Println(result.Content)
```

**After (v0.3.0):**
```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    Ask(ctx, "Hello")
fmt.Println(response)
```

#### Streaming

**Before:**
```go
result, err := agent.Chat(ctx, "Hello", &agent.ChatOptions{
    Stream: true,
    OnStream: func(delta string) { fmt.Print(delta) },
})
```

**After:**
```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(delta string) { fmt.Print(delta) }).
    Stream(ctx, "Hello")
```

#### Conversation Memory

**Before:**
```go
result, err := agent.Chat(ctx, "", &agent.ChatOptions{
    Messages: conversationHistory,
})
```

**After:**
```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).WithMemory()

builder.Ask(ctx, "First question")
builder.Ask(ctx, "Second question") // Remembers context automatically
```

#### Tool Calling

**Before:**
```go
result, err := agent.Chat(ctx, "Weather?", &agent.ChatOptions{
    Tools: tools,
})
```

**After:**
```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTools(weatherTool).
    WithAutoExecute(true).
    Ask(ctx, "What's the weather?")
```

#### Advanced Configuration

**Before:**
```go
result, err := agent.Chat(ctx, "Explain Go", &agent.ChatOptions{
    Temperature: 0.7,
    MaxTokens: 500,
    Stream: true,
    OnStream: streamHandler,
})
```

**After:**
```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithTemperature(0.7).
    WithMaxTokens(500).
    OnStream(streamHandler).
    Stream(ctx, "Explain Go")
```

#### New Features in v0.3.0

**Multimodal (Vision):**
```go
// Analyze images with GPT-4 Vision
response, err := agent.NewOpenAI("gpt-4o", apiKey).
    WithImage("https://example.com/photo.jpg").
    Ask(ctx, "What's in this image?")
```

**Error Handling with Retry:**
```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithTimeout(30 * time.Second).
    WithRetry(3).
    WithExponentialBackoff().
    Ask(ctx, "Your question")
```

**JSON Schema:**
```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("person", "A person object", personSchema, true).
    Ask(ctx, "Generate a person")
```

### Key Benefits of v0.3.0

1. **More Readable** - Fluent API reads like English
2. **Better IDE Support** - Method chaining with autocomplete
3. **Type Safety** - Compile-time checks
4. **Composable** - Chain any methods together
5. **Discoverable** - All options visible in IDE
6. **Flexible** - Reuse builders, modify on the fly

### Migrating from v0.1.0 to v0.2.0

#### Simple Chat
**Before:**
```go
response, err := agent.Chat(ctx, "Hello", false)
fmt.Println(response)
```

**After:**
```go
result, err := agent.Chat(ctx, "Hello", nil)
fmt.Println(result.Content)
```

#### Streaming
**Before:**
```go
err := agent.ChatStream(ctx, "Hello", func(delta string) {
    fmt.Print(delta)
})
```

**After:**
```go
result, err := agent.Chat(ctx, "Hello", &agent.ChatOptions{
    Stream: true,
    OnStream: func(delta string) {
        fmt.Print(delta)
    },
})
```

#### Conversation History
**Before:**
```go
response, err := agent.ChatWithHistory(ctx, messages)
```

**After:**
```go
result, err := agent.Chat(ctx, "", &agent.ChatOptions{
    Messages: messages,
})
```

#### Tool Calling
**Before:**
```go
completion, err := agent.ChatWithToolCalls(ctx, "Weather?", tools)
```

**After:**
```go
result, err := agent.Chat(ctx, "Weather?", &agent.ChatOptions{
    Tools: tools,
})
// Access full completion: result.Completion
```

#### Combined Features (NEW!)
```go
// Now you can combine streaming + history + tools!
result, err := agent.Chat(ctx, "next question", &agent.ChatOptions{
    Messages: conversationHistory,
    Tools:    tools,
    Stream:   true,
    OnStream: func(s string) { fmt.Print(s) },
})
```

### Benefits of Migration

1. **Single API** - One method to learn instead of four
2. **Composable** - Easily combine features (streaming + history + tools)
3. **Consistent** - All operations return same type (`*ChatResult`)
4. **Extensible** - Easy to add new options without breaking changes
5. **Cleaner Code** - Less method pollution, clearer intent

### GetCompletion() Unchanged

The advanced `GetCompletion()` method remains unchanged for power users who need full control over OpenAI API parameters.
