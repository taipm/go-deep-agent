# Data Models - go-deep-agent

## Overview

This document describes the core data structures and models used in go-deep-agent.

## Core Data Structures

### Message

```go
type Message struct {
    Role    string
    Content string
    // Additional fields for multimodal content
}
```

Represents a conversation message with role (system/user/assistant) and content.

### Tool

```go
type Tool struct {
    Name        string
    Description string
    Parameters  interface{} // JSON schema for parameters
    Handler     func(args map[string]interface{}) (string, error)
}
```

Defines a tool/function that can be called by the LLM.

### Document (RAG)

```go
type Document struct {
    Content  string
    Metadata map[string]interface{}
    Score    float64 // Relevance score for retrieval
}
```

Represents a document for RAG (Retrieval-Augmented Generation).

### ImageContent (Multimodal)

```go
type ImageContent struct {
    URL    string
    Base64 string
    Detail string // "auto", "low", "high"
}
```

Represents image content for vision models.

### BatchResult

```go
type BatchResult struct {
    Index   int
    Content string
    Error   error
}
```

Result from batch processing operations.

## Memory Models

### Episodic Memory Entry

```go
// From memory/episodic.go
type EpisodicEntry struct {
    Timestamp time.Time
    Role      string
    Content   string
    Context   map[string]interface{}
}
```

Stores sequential conversation events.

### Semantic Memory Entry

```go
// From memory/semantic.go
type SemanticEntry struct {
    Key       string
    Value     string
    Category  string
    Embedding []float64 // For vector similarity
}
```

Stores factual knowledge and long-term information.

### Working Memory

```go
// From memory/working.go
type WorkingMemoryState struct {
    CurrentContext map[string]interface{}
    ActiveTasks    []Task
    RecentMessages []Message
}
```

Maintains current conversation state and active context.

## Configuration Models

### Config

```go
type Config struct {
    Model       string
    Temperature float64
    MaxTokens   int
    // Additional configuration fields
}
```

Basic agent configuration.

### RAGConfig

```go
type RAGConfig struct {
    TopK          int     // Number of documents to retrieve
    ScoreThreshold float64 // Minimum relevance score
    MaxTokens     int     // Max tokens from retrieved docs
}
```

Configuration for RAG functionality.

## Provider-Specific Models

### OpenAI Models

Uses official OpenAI SDK types:
- `openai.ChatCompletionMessageParamUnion`
- `openai.ChatCompletion`
- `openai.ChatCompletionToolUnionParam`
- `openai.FinishedChatCompletionToolCall`

### Gemini Models

Uses Google Generative AI SDK types:
- `genai.GenerativeModel`
- `genai.Content`
- `genai.Part`

## Vector Store Models

### VectorStore Interface

```go
type VectorStore interface {
    AddDocuments(ctx context.Context, docs []Document) error
    Search(ctx context.Context, query string, topK int) ([]Document, error)
    Delete(ctx context.Context, id string) error
}
```

Interface for vector database implementations (e.g., Qdrant).

### EmbeddingProvider Interface

```go
type EmbeddingProvider interface {
    Embed(ctx context.Context, text string) ([]float64, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}
```

Interface for generating text embeddings.

## Cache Models

### Cache Interface

```go
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

Interface for caching implementations (Redis, in-memory, etc.).

## Error Types

### Custom Errors

```go
type AgentError struct {
    Code    string
    Message string
    Cause   error
}
```

Structured error handling for agent operations.

## Storage Patterns

### Redis-Backed Storage

- **Memory Persistence**: Episodic and semantic memory can be stored in Redis
- **Caching**: LLM responses cached with configurable TTL
- **Vector Storage**: Integration with Redis vector search (future)

### In-Memory Storage

- **Working Memory**: Primarily in-memory for performance
- **Tool State**: Transient state during execution
- **Batch Processing**: Temporary result accumulation

## Data Flow

```
User Input
    ↓
Builder Configuration
    ↓
Message Construction
    ↓
Memory Retrieval (RAG/Context)
    ↓
LLM API Call (via Adapter)
    ↓
Response Processing
    ↓
Tool Execution (if needed)
    ↓
Memory Storage
    ↓
Result Return
```

## Persistence

### Supported Persistence Layers

1. **Redis** - For caching and memory
2. **Vector Databases** - For RAG (Qdrant integration)
3. **File System** - For configuration and logs

### Data Retention

- **Working Memory**: Session-scoped (cleared after conversation)
- **Episodic Memory**: Configurable retention (can persist to Redis)
- **Semantic Memory**: Long-term (persisted)
- **Cache**: TTL-based expiration

## Schema Validation

Uses JSON Schema for:
- Tool parameter validation
- Response format enforcement (structured outputs)
- Configuration validation

---

**Generated:** 2025-11-14
**Scan Level:** Deep
**Project Type:** Backend Library (Go SDK)
