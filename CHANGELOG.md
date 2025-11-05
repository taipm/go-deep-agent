# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
