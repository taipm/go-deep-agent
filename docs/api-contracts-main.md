# API Contracts - go-deep-agent

## Overview

This document describes the public API interfaces of go-deep-agent, a Go library for building AI agents with support for multiple LLM providers.

## Core Types

### Agent

```go
type Agent struct {
    config Config
    client *openai.Client
}
```

**Primary Method:**

```go
func (a *Agent) Chat(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error)
```

Sends a message to the LLM and returns the response. Supports:
- Simple chat
- Conversation history
- Streaming responses
- Tool/function calling

**ChatOptions:**
```go
type ChatOptions struct {
    Stream   bool                                     // Enable streaming
    OnStream func(string)                             // Streaming callback
    Messages []openai.ChatCompletionMessageParamUnion // Conversation history
    Tools    []openai.ChatCompletionToolUnionParam    // Tools for function calling
}
```

**ChatResult:**
```go
type ChatResult struct {
    Content    string                 // Response content
    Completion *openai.ChatCompletion // Full completion object
}
```

### Builder

The `Builder` provides a fluent API for constructing and executing LLM requests with extensive configuration options.

```go
type Builder struct {
    // Core configuration
    provider Provider
    model    string
    apiKey   string
    baseURL  string

    // Conversation state
    systemPrompt string
    messages     []Message
    autoMemory   bool
    maxHistory   int

    // Advanced parameters
    temperature      *float64
    topP             *float64
    maxTokens        *int64
    presencePenalty  *float64
    frequencyPenalty *float64
    seed             *int64
    logprobs         *bool
    topLogprobs      *int64
    n                *int64

    // Streaming callbacks
    onStream   func(content string)
    onToolCall func(tool openai.FinishedChatCompletionToolCall)
    onRefusal  func(refusal string)

    // Tool calling
    tools         []*Tool
    autoExecute   bool
    maxToolRounds int
    toolChoice    *openai.ChatCompletionToolChoiceOptionUnionParam

    // Tool orchestration
    enableParallel bool
    maxWorkers     int
    toolTimeout    time.Duration

    // Response format
    responseFormat *openai.ChatCompletionNewParamsResponseFormatUnion

    // Error handling & recovery
    timeout       time.Duration
    maxRetries    int
    retryDelay    time.Duration
    useExpBackoff bool

    // Multimodal support
    pendingImages []ImageContent
    lastError     error

    // Batch processing
    batchSize           int
    batchDelay          time.Duration
    onBatchProgress     func(completed, total int)
    onBatchItemComplete func(result BatchResult)

    // RAG (Retrieval-Augmented Generation)
    ragEnabled        bool
    ragDocuments      []Document
    ragRetriever      RAGRetriever
    ragConfig         *RAGConfig
    lastRetrievedDocs []Document

    // Vector RAG
    vectorStore       VectorStore
    embeddingProvider EmbeddingProvider
    vectorCollection  string

    // Caching
    cache        Cache
    cacheEnabled bool
    cacheTTL     time.Duration
}
```

**Example Usage:**

```go
response := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    Ask(ctx, "Hello!")
```

## Tools System

Located in `agent/tools/`

### Available Built-in Tools

1. **DateTime Tools** (`datetime.go`)
   - Time/date operations

2. **Filesystem Tools** (`filesystem.go`)
   - File operations

3. **HTTP Tools** (`http.go`)
   - HTTP requests

4. **Logger Tools** (`logger.go`)
   - Logging functionality

5. **Math Tools** (`math.go`)
   - Mathematical operations

6. **Orchestrator** (`orchestrator.go`)
   - Tool coordination and parallel execution

### Tool Registration

```go
type Tool struct {
    // Tool definition and execution logic
}
```

Tools can be registered with the builder and executed automatically when the LLM requests them.

## Memory System

Located in `agent/memory/`

### Memory Types

1. **Episodic Memory** (`episodic.go`)
   - Event-based memory storage
   - Sequential conversation history

2. **Semantic Memory** (`semantic.go`)
   - Knowledge and facts storage
   - Long-term factual information

3. **Working Memory** (`working.go`)
   - Short-term active context
   - Current conversation state

4. **System Memory** (`system.go`)
   - System-level configurations and state

### Memory Interfaces

```go
// Defined in memory/interfaces.go
type Memory interface {
    // Memory operations
}
```

## Adapters

Located in `agent/adapters/`

### OpenAI Adapter

```go
// openai_adapter.go
// Implements OpenAI API integration
```

### Gemini Adapter

```go
// gemini_adapter.go
// Implements Google Gemini API integration
```

## Key Features

### 1. Multi-Provider Support
- OpenAI (GPT models)
- Google Gemini

### 2. Advanced Capabilities
- **Streaming**: Real-time response streaming with callbacks
- **Tool Calling**: Automatic tool execution with parallel support
- **RAG**: Document retrieval and vector search integration
- **Memory**: Multiple memory types for context management
- **Caching**: Response caching for performance
- **Batch Processing**: Concurrent request processing
- **Error Recovery**: Retry logic with exponential backoff
- **Multimodal**: Image support for vision models

### 3. Configuration Options
- Temperature, Top-P, Max Tokens
- Presence/Frequency penalties
- Seed for reproducibility
- Response format (JSON schema)
- Timeout and retry settings

## Provider Enumeration

```go
type Provider int

const (
    ProviderOpenAI Provider = iota
    ProviderGemini
)
```

## Entry Points

Primary entry points for creating agents:

1. `NewOpenAI(model, apiKey string) *Builder`
2. `NewGemini(model, apiKey string) *Builder`

## Rate Limiting

Uses `golang.org/x/time/rate` for request rate limiting.

## Testing

Comprehensive test coverage including:
- Unit tests (`*_test.go` files)
- Integration tests
- Mock support via `miniredis` for Redis-backed features

## Dependencies

Key external dependencies:
- `github.com/openai/openai-go/v3` - OpenAI SDK
- `github.com/google/generative-ai-go` - Google Gemini SDK
- `github.com/redis/go-redis/v9` - Redis client (for caching/memory)
- `gopkg.in/yaml.v3` - YAML configuration
- `github.com/stretchr/testify` - Testing utilities

---

**Generated:** 2025-11-14
**Scan Level:** Deep
**Project Type:** Backend Library (Go SDK)
