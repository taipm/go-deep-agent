# Go Deep Agent 🚀

A powerful yet simple LLM agent library for Go with a modern **Fluent Builder API**. Build AI applications with method chaining, automatic conversation memory, intelligent error handling, and seamless streaming support.

Built with [openai-go v3.8.1](https://github.com/openai/openai-go).

> **Why go-deep-agent?** 60-80% less code than openai-go with 10x better developer experience. [See detailed comparison →](docs/COMPARISON.md)

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
- 🚀 **Batch Processing** - Concurrent request processing with progress tracking (v0.4.0)
- 📚 **RAG Support** - Retrieval-Augmented Generation with document chunking (v0.4.0)
- 💾 **Response Caching** - Memory & Redis caching with TTL management (v0.4.0, v0.5.1 🆕)
- 🔢 **Vector Embeddings** - OpenAI & Ollama embeddings with similarity search (v0.5.0 🆕)
- 🗄️ **Vector Databases** - ChromaDB & Qdrant integration for semantic search (v0.5.0 🆕)
- 🧠 **Vector RAG** - Semantic retrieval with auto-embedding and priority system (v0.5.0 🆕)
- 📊 **Logging & Observability** - Zero-overhead logging with slog support (v0.5.2 🆕)
- 🛠️ **Built-in Tools** - FileSystem, HTTP, DateTime tools ready to use (v0.5.3 🆕)
- ✅ **Well Tested** - 460+ tests, 65%+ coverage, 70+ working examples

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

### 8. Redis Cache - Distributed Caching (v0.5.1 🆕)

```go
// Simple Redis cache setup
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCache("localhost:6379", "", 0)

// First call - cache miss (~1-2s)
resp1, _ := ai.Ask(ctx, "What is Go?")

// Second call - cache hit (~5ms, 200x faster!)
resp2, _ := ai.Ask(ctx, "What is Go?")

// Check cache statistics
stats := ai.GetCacheStats()
fmt.Printf("Hit rate: %.2f%%\n",
    float64(stats.Hits)/(float64(stats.Hits+stats.Misses))*100)

// Advanced configuration
opts := &agent.RedisCacheOptions{
    Addrs:       []string{"localhost:6379"},
    Password:    "your-redis-password",
    PoolSize:    20,                 // Connection pool
    KeyPrefix:   "myapp",            // Namespace
    DefaultTTL:  10 * time.Minute,   // Cache expiration
}

ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCacheOptions(opts)

// Custom TTL per request
ai.WithCacheTTL(1 * time.Hour).Ask(ctx, "Historical facts")

// Redis Cluster support
opts := &agent.RedisCacheOptions{
    Addrs: []string{
        "redis-node1:6379",
        "redis-node2:6379",
        "redis-node3:6379",
    },
    Password: "cluster-password",
}
```

**Benefits:**

- Shared cache across multiple instances
- Persistent cache (survives restarts)
- Distributed locking (prevents cache stampede)
- Scalable with Redis Cluster

### 9. Logging & Observability (v0.5.2 🆕)

```go
// Debug logging for development
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDebugLogging() // Detailed logs with timing, cache, tools, RAG

response, err := builder.Ask(ctx, "Hello!")
// Output:
// [2024-01-15 10:30:45] DEBUG: Ask request started | model=gpt-4o-mini
// [2024-01-15 10:30:45] DEBUG: Cache miss | duration_ms=2
// [2024-01-15 10:30:46] INFO: Ask completed | duration_ms=890 tokens=23

// Info logging for production
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithInfoLogging() // Important events only

// Slog integration (Go 1.21+)
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
logger := slog.New(handler)

builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(agent.NewSlogAdapter(logger))
// Output: {"time":"...","level":"INFO","msg":"Ask completed","duration_ms":890}

// Custom logger (Zap, Logrus, etc.)
type MyLogger struct { /* implement Logger interface */ }
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(&MyLogger{})

// No logging (default - zero overhead)
builder := agent.NewOpenAI("gpt-4o-mini", apiKey)
// NoopLogger - literally zero cost
```

**What gets logged:**

- Request lifecycle (start, duration, completion)
- Token usage (prompt, completion, total)
- Cache operations (hit/miss, duration)
- Tool execution (rounds, tool calls, results)
- RAG retrieval (docs retrieved, method)
- Retry attempts (delays, errors)
- Errors with context

📖 **[Complete Logging Guide](docs/LOGGING_GUIDE.md)** - Custom loggers, slog integration, production best practices

### 10. Built-in Tools

#### 10.1 FileSystem, HTTP, DateTime Tools (v0.5.3 🆕)

Three production-ready tools for common operations: file system, HTTP requests, and date/time.

```go
import "github.com/taipm/go-deep-agent/agent/tools"

// Create built-in tools
fsTool := tools.NewFileSystemTool()      // File operations
httpTool := tools.NewHTTPRequestTool()   // HTTP client
dtTool := tools.NewDateTimeTool()        // Date/time calculations

// Use with agent
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTools(fsTool, httpTool, dtTool).
    WithAutoExecute(true)

// Agent can now use tools automatically
response, _ := ai.Ask(ctx, "Read config.json, get current time in Tokyo, and fetch https://api.github.com/users/github")
```

**FileSystemTool** - 7 operations:
- `read_file`, `write_file`, `append_file`, `delete_file`
- `list_directory`, `file_exists`, `create_directory`
- Security: Path traversal prevention

**HTTPRequestTool** - Full HTTP client:
- Methods: GET, POST, PUT, DELETE
- Features: Headers, timeout, JSON parsing
- Default 30s timeout

**DateTimeTool** - 7 operations:
- `current_time`, `format_date`, `parse_date`
- `add_duration`, `date_diff`, `convert_timezone`, `day_of_week`
- Timezones: UTC, America/New_York, Asia/Tokyo, etc.

#### 10.2 Math Tool (v0.5.4 🆕)

Professional-grade mathematical operations powered by **govaluate** (expression engine) and **gonum** (statistical computing).

```go
mathTool := tools.NewMathTool()

ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithTool(mathTool).
    WithAutoExecute(true)

// Expression evaluation
ai.Ask(ctx, "Calculate: 2 * (3 + 4) + sqrt(16)")
// Uses govaluate with 11 functions: sqrt, pow, sin, cos, tan, log, ln, abs, ceil, floor, round

// Statistics
ai.Ask(ctx, "What's the average of 10, 20, 30, 40, 50?")
// Uses gonum/stat: mean, median, stdev, variance, min, max, sum

// Equation solving
ai.Ask(ctx, "Solve: x+15=42")
// Linear equations (quadratic coming in v0.6.0)

// Unit conversion
ai.Ask(ctx, "Convert 100 km to meters")
// Distance, weight, temperature, time conversions

// Random generation
ai.Ask(ctx, "Generate random number between 1 and 100")
// Integer, float, choice operations
```

**MathTool** - 5 operation categories:
- **evaluate**: Mathematical expressions with 11 functions (80% coverage)
- **statistics**: 7 statistical measures via gonum (15% coverage)
- **solve**: Linear equations, quadratic coming soon (3% coverage)
- **convert**: Distance, weight, temperature, time units (1% coverage)
- **random**: Integer, float, choice generation (1% coverage)

**Dependencies**: +9MB binary for professional accuracy (govaluate, gonum)

📖 **[View builtin_tools_demo.go](examples/builtin_tools_demo.go)** - Complete examples

### 11. History Management

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

### 12. Multimodal - Vision (GPT-4 Vision)

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

### 13. Vector RAG - Semantic Search (v0.5.0 🆕)

```go
// Setup vector database and embeddings
embedding, _ := agent.NewOllamaEmbedding(
    "http://localhost:11434",
    "nomic-embed-text", // Free local embeddings
)

store, _ := agent.NewChromaStore("http://localhost:8000")
store.WithEmbedding(embedding)

// Create collection
config := &agent.CollectionConfig{
    Name:           "company-kb",
    Dimension:      768,
    DistanceMetric: agent.DistanceMetricCosine,
}
store.CreateCollection(ctx, "company-kb", config)

// Create agent with vector RAG
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithVectorRAG(embedding, store, "company-kb").
    WithRAGTopK(3).         // Retrieve top 3 similar docs
    WithMemory()

// Add knowledge base
docs := []string{
    "Our refund policy allows full refunds within 30 days.",
    "Customer support is available 24/7 at support@company.com.",
    "We support integrations with Slack, Teams, and Salesforce.",
}
ai.AddDocumentsToVector(ctx, docs...)

// Ask questions - semantically retrieves relevant context
response, _ := ai.Ask(ctx, "What is your refund policy?")
fmt.Println(response)

// See what was retrieved
retrieved := ai.GetLastRetrievedDocs()
for _, doc := range retrieved {
    fmt.Printf("Score: %.3f | %s\n", doc.Score, doc.Content)
}
```

### 14. Advanced Vector RAG with Metadata

```go
// Add documents with rich metadata
vectorDocs := []*agent.VectorDocument{
    {
        Content: "Python is great for data science and machine learning.",
        Metadata: map[string]interface{}{
            "category":   "programming",
            "language":   "Python",
            "difficulty": "beginner",
            "tags":       []string{"data-science", "ml"},
        },
    },
    {
        Content: "Go is excellent for building high-performance backend services.",
        Metadata: map[string]interface{}{
            "category":   "programming",
            "language":   "Go",
            "difficulty": "intermediate",
            "tags":       []string{"backend", "concurrency"},
        },
    },
}

ai.AddVectorDocuments(ctx, vectorDocs...)

// Query with custom config
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithVectorRAG(embedding, store, "docs").
    WithRAGConfig(&agent.RAGConfig{
        TopK:          5,     // Retrieve top 5
        MinScore:      0.7,   // Only high-confidence results
        IncludeScores: true,  // Show relevance scores
    })

response, _ := ai.Ask(ctx, "Tell me about backend programming")

// Access retrieved metadata
docs := ai.GetLastRetrievedDocs()
for _, doc := range docs {
    fmt.Printf("Language: %s, Difficulty: %s\n",
        doc.Metadata["language"],
        doc.Metadata["difficulty"])
}
```

### 14. Switch Vector Databases - ChromaDB vs Qdrant

```go
// Development: Use ChromaDB (easy setup)
chromaStore, _ := agent.NewChromaStore("http://localhost:8000")
aiDev := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithVectorRAG(embedding, chromaStore, "docs")

// Production: Use Qdrant (high performance)
qdrantStore, _ := agent.NewQdrantStore("http://localhost:6333")
qdrantStore.WithAPIKey("your-api-key") // Optional
aiProd := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithVectorRAG(embedding, qdrantStore, "docs")

// Both use the same API - seamless switching!
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

### Vector RAG (v0.5.0 🆕)

- `WithVectorRAG(embedding, store, collection)` - Enable vector-based RAG
- `AddDocumentsToVector(ctx, docs...)` - Add string documents to vector store
- `AddVectorDocuments(ctx, vectorDocs...)` - Add documents with metadata
- `GetLastRetrievedDocs()` - Get retrieved documents with scores

### Response Caching (v0.4.0, v0.5.1 🆕)

- `WithCache(cache)` - Set custom cache implementation
- `WithMemoryCache(maxSize, defaultTTL)` - In-memory LRU cache
- `WithRedisCache(addr, password, db)` - Redis distributed cache (simple)
- `WithRedisCacheOptions(opts)` - Redis cache with advanced config
- `WithCacheTTL(ttl)` - Set custom TTL for next request
- `DisableCache()` - Temporarily disable caching
- `EnableCache()` - Re-enable caching
- `GetCacheStats()` - Retrieve cache statistics (hits, misses, hit rate)
- `ClearCache(ctx)` - Clear all cached responses

### Embedding Providers (v0.5.0 🆕)

- `NewOllamaEmbedding(baseURL, model)` - Free local embeddings (Ollama)
- `NewOpenAIEmbedding(apiKey, model, dimension)` - OpenAI embeddings
- `Generate(ctx, texts)` - Generate embeddings for texts
- `GenerateQuery(ctx, query)` - Generate embedding for search query

### Vector Stores (v0.5.0 🆕)

- `NewChromaStore(baseURL)` - Create ChromaDB client
- `NewQdrantStore(baseURL)` - Create Qdrant client
- `CreateCollection(ctx, name, config)` - Create collection with config
- `Add(ctx, collection, documents)` - Add documents with auto-embedding
- `Search(ctx, request)` - Vector similarity search
- `SearchByText(ctx, request)` - Text-based semantic search
- `Delete(ctx, collection, ids)` - Delete documents by IDs
- `Count(ctx, collection)` - Get document count
- `Clear(ctx, collection)` - Remove all documents

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

## 📊 Quality Metrics

- ✅ **460+ Tests** passing across all features
- ✅ **65%+ Coverage** with comprehensive test cases
- ✅ **15 Example Files** with 70+ working examples
- ✅ **Production Libraries** (openai-go, govaluate, gonum)
- ✅ **Production Tested** with real OpenAI, Ollama, ChromaDB, Qdrant

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

# Pull embedding model (for vector RAG)
ollama pull nomic-embed-text

# Run Ollama server
ollama serve

# Run Ollama examples
go run examples/ollama_example.go
```

### Vector Database Setup (v0.5.0 🆕)

```bash
# ChromaDB (easiest for development)
docker run -p 8000:8000 chromadb/chroma

# OR Qdrant (production-ready)
docker run -p 6333:6333 qdrant/qdrant

# Run vector RAG examples
go run examples/vector_rag_example.go
go run examples/chroma_example.go
go run examples/qdrant_example.go
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

# Vector RAG examples (v0.5.0 🆕)
go run examples/embedding_example.go      # Embedding basics
go run examples/chroma_example.go         # ChromaDB integration
go run examples/qdrant_example.go         # Qdrant integration
go run examples/vector_rag_example.go     # Complete RAG workflow

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

## 🆚 Why Choose go-deep-agent?

**60-80% less code** than raw openai-go SDK with **10x better developer experience**.

| Feature | openai-go | go-deep-agent | Improvement |
|---------|-----------|---------------|-------------|
| Simple Chat | 26 lines | 14 lines | ⬇️ 46% |
| Streaming | 20+ lines | 5 lines | ⬇️ 75% |
| Memory | 28+ lines (manual) | 6 lines (auto) | ⬇️ 78% |
| Tool Calling | 50+ lines | 14 lines | ⬇️ 72% |
| Multimodal | 25+ lines | 5 lines | ⬇️ 80% |

**[📖 See detailed comparison with code examples →](docs/COMPARISON.md)**

### Key Advantages

- ✅ **Fluent API** - Method chaining reads like natural language
- ✅ **Automatic Features** - Memory, retry, error handling built-in
- ✅ **Production-Ready** - 242 tests, 65.8% coverage, CI/CD
- ✅ **Better DX** - IDE autocomplete, self-documenting code
- ✅ **All openai-go Features** - Plus high-level conveniences

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
- **[COMPARISON.md](docs/COMPARISON.md)** - 🆚 Why go-deep-agent vs openai-go (with code examples)
- **[CHANGELOG.md](CHANGELOG.md)** - Version history and migration guides
- **[RAG_VECTOR_DATABASES.md](docs/RAG_VECTOR_DATABASES.md)** - 🆕 Complete Vector RAG guide (v0.5.0)
- **[LOGGING_GUIDE.md](docs/LOGGING_GUIDE.md)** - 🆕 Comprehensive logging & observability guide (v0.5.2)
- **[examples/](examples/)** - 15 example files with 69+ working examples
- **[agent/README.md](agent/README.md)** - Detailed API documentation
- **[TODO.md](TODO.md)** - Roadmap and implementation progress
- **[ROADMAP.md](ROADMAP.md)** - v0.5.0 Advanced RAG implementation plan

## 🔗 Links

- **GitHub**: [github.com/taipm/go-deep-agent](https://github.com/taipm/go-deep-agent)
- **openai-go**: [github.com/openai/openai-go](https://github.com/openai/openai-go) - Official OpenAI Go library
- **Ollama**: [ollama.com](https://ollama.com) - Run LLMs locally

---

<div align="center">

**Made with ❤️ for the Go community**

⭐ Star us on GitHub if you find this useful!

</div>