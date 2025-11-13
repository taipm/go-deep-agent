# Memory System Guide

**Complete guide to understanding memory, caching, and storage in go-deep-agent**

> ⚠️ **Important**: This guide clarifies the differences between Memory, Cache, and Vector Storage to avoid confusion.

---

## 🎯 Quick Overview: What's What?

| Component | Purpose | Storage | Persistence | Shared Across Instances |
|-----------|---------|---------|-------------|-------------------------|
| **Conversation Memory** | Remember chat history | RAM | ❌ No (volatile) | ❌ No (per-instance) |
| **Response Cache** | Avoid duplicate API calls | RAM or Redis | ✅ Yes (with Redis) | ✅ Yes (with Redis) |
| **Vector Store** | Knowledge base for RAG | Qdrant/Chroma | ✅ Yes (database) | ✅ Yes (database) |

---

## 📚 Part 1: Conversation Memory (RAM Only)

### What It Does

Stores conversation history so the AI remembers previous messages in the same session.

### Storage Location

**RAM only** - stored in `[]Message` slice inside `Builder` struct.

```go
// agent/builder.go line 35
messages []Message  // ← Lives in memory, lost on restart
```

### How to Enable

```go
agent := agent.NewOpenAI("gpt-4", apiKey).
    WithMemory().        // Enable conversation history
    WithMaxHistory(20)   // Keep last 20 messages
```

### What Happens

```go
// First message
agent.Ask(ctx, "My name is Alice")
// ✅ Stored in b.messages: ["user: My name is Alice", "assistant: Nice to meet you, Alice"]

// Second message
agent.Ask(ctx, "What's my name?")
// ✅ Agent remembers: Checks b.messages, finds "Alice"

// App restarts...
agent = agent.NewOpenAI("gpt-4", apiKey).WithMemory()
agent.Ask(ctx, "What's my name?")
// ❌ Agent forgets: b.messages is empty after restart
```

### Architecture

```
User Message → Builder.Ask()
                  ↓
            if autoMemory == true
                  ↓
            b.messages = append(b.messages, userMsg, assistantMsg)
                  ↓
            Auto-truncate if > maxHistory (FIFO)
```

### Key Methods

```go
// Enable/disable
.WithMemory()                    // Enable autoMemory = true
.DisableMemory()                 // Disable memory

// Configure
.WithMaxHistory(20)              // Keep last 20 messages

// Manual control
.GetHistory() []Message          // Get current conversation
.SetHistory(messages)            // Restore previous conversation
.Clear()                         // Reset conversation
```

### Limitations

| Issue | Impact | Workaround |
|-------|--------|------------|
| **Not persistent** | Lost on restart | Manual save/restore with JSON |
| **Not shared** | Each instance has own history | Use external storage |
| **RAM only** | Memory grows with conversation | Use `WithMaxHistory()` |

### Manual Save/Restore Pattern

```go
import "encoding/json"

// Save to file
history := agent.GetHistory()
data, _ := json.Marshal(history)
os.WriteFile("conversation.json", data, 0644)

// Restore from file
data, _ := os.ReadFile("conversation.json")
var history []agent.Message
json.Unmarshal(data, &history)
agent.SetHistory(history)
```

---

## 💾 Part 2: Hierarchical Memory (Advanced RAM-based)

### What It Does

Three-tier memory system for intelligent conversation management:

1. **Working Memory** - Recent messages (FIFO)
2. **Episodic Memory** - Important messages (importance-scored)
3. **Semantic Memory** - Extracted knowledge

### Storage Location

**All in RAM** - inside `agent/memory/` package structures.

### How to Enable

```go
agent := agent.NewOpenAI("gpt-4", apiKey).
    WithMemory().
    WithHierarchicalMemory(agent.MemoryConfig{
        WorkingCapacity:     20,
        EpisodicEnabled:     true,
        EpisodicThreshold:   0.5,  // Store messages with importance > 0.5
        ImportanceScoring:   true,
    })
```

### How It Works

```
New Message
    ↓
Working Memory (last 20 messages)
    ↓ (importance score > 0.5)
Episodic Memory (important messages)
    ↓ (knowledge extraction)
Semantic Memory (facts, entities)
```

### Example

```go
agent.Ask(ctx, "My birthday is December 25th")
// Working: stores entire message
// Episodic: importance 0.9 → stores in long-term
// Semantic: extracts "birthday = December 25th"

agent.Ask(ctx, "The weather is nice today")
// Working: stores entire message
// Episodic: importance 0.2 → discarded
// Semantic: no fact extracted

// After 100 messages...
agent.Ask(ctx, "When is my birthday?")
// Searches Episodic → finds "December 25th"
```

### Limitations

⚠️ **STILL RAM-based**: All three tiers live in memory, lost on restart!

```go
// ❌ Common misconception
agent.WithHierarchicalMemory(config)
// This does NOT persist to disk/database
// It's just smarter in-memory organization
```

---

## 🚀 Part 3: Response Cache

### What It Does

Caches API responses to avoid duplicate calls for **identical prompts**.

### Storage Options

1. **Memory Cache** (RAM) - Single instance, lost on restart
2. **Redis Cache** (Database) - Multi-instance, persistent

### When It Helps

```go
// First call - Cache MISS
agent.Ask(ctx, "What is 2+2?")  // → Calls OpenAI API ($)

// Second call - Cache HIT
agent.Ask(ctx, "What is 2+2?")  // → Returns cached response (free!)
```

### How to Enable

#### Option A: Memory Cache (Simple)

```go
agent := agent.NewOpenAI("gpt-4", apiKey).
    WithMemoryCache(1000, 5*time.Minute)
    //              ↑    ↑
    //              |    └─ TTL per cache entry
    //              └────── Max 1000 cached responses
```

**Storage**: RAM (lost on restart)

#### Option B: Redis Cache (Production)

```go
agent := agent.NewOpenAI("gpt-4", apiKey).
    WithRedisCache("localhost:6379", "", 0)
```

**Storage**: Redis database (persistent, shared)

#### Option C: Redis with Advanced Config

```go
opts := &agent.RedisCacheOptions{
    Addrs:      []string{"localhost:6379"},
    Password:   "secret",
    DB:         0,
    PoolSize:   20,
    KeyPrefix:  "myapp",
    DefaultTTL: 10 * time.Minute,
}

agent := agent.NewOpenAI("gpt-4", apiKey).
    WithRedisCacheOptions(opts)
```

### Redis Cluster Support

```go
opts := &agent.RedisCacheOptions{
    Addrs: []string{
        "redis-node1:6379",
        "redis-node2:6379",
        "redis-node3:6379",
    },
}
```

### Cache Key Generation

```
Cache Key = Hash(model + system_prompt + temperature + messages + tools)
```

**Example:**
```go
// These are cached separately (different keys)
agent1.WithSystem("You are helpful").Ask(ctx, "Hi")
agent2.WithSystem("You are funny").Ask(ctx, "Hi")

// These share cache (same key)
agent1.Ask(ctx, "What is 2+2?")
agent2.Ask(ctx, "What is 2+2?")  // Cache HIT ✅
```

### Cache Operations

```go
// Get stats
stats := agent.GetCacheStats()
fmt.Printf("Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)

// Clear cache
agent.ClearCache(ctx)

// Check if cached
exists := cache.Exists(ctx, cacheKey)

// Set custom TTL for next request
agent.WithCacheTTL(1 * time.Hour).Ask(ctx, "Important query")
```

### Redis Cache Features

```go
cache := agent.NewRedisCache("localhost:6379", "", 0, 5*time.Minute)

// Distributed lock
locked, _ := cache.SetNX(ctx, "lock:user123", "value", 10*time.Second)

// Batch operations
cache.MSet(ctx, map[string]string{
    "key1": "value1",
    "key2": "value2",
}, 5*time.Minute)

values, _ := cache.MGet(ctx, "key1", "key2")

// Pattern deletion
cache.DeletePattern(ctx, "user:*")

// TTL management
ttl, _ := cache.TTL(ctx, "key1")
cache.Expire(ctx, "key1", 1*time.Hour)
```

---

## 🔍 Part 4: Vector Stores (Qdrant & ChromaDB)

### What It Does

Stores documents/embeddings for **RAG (Retrieval-Augmented Generation)** - knowledge base search.

### Storage Location

**External database** (Qdrant or ChromaDB) - persistent, shared.

### Supported Backends

| Backend | Default URL | Features |
|---------|------------|----------|
| **Qdrant** | `http://localhost:6333` | Production-ready, cloud option |
| **ChromaDB** | `http://localhost:8000` | Lightweight, Python-friendly |

### How to Enable

#### Setup Qdrant

```go
// Create Qdrant client
qdrant, _ := agent.NewQdrantStore("http://localhost:6333")
qdrant.WithAPIKey("your-api-key")

// Create embedding provider
embedding := agent.NewOpenAIEmbedding("text-embedding-3-small", apiKey)

// Connect both to agent
agent := agent.NewOpenAI("gpt-4", apiKey).
    WithVectorRAG(qdrant, "my-knowledge-base", embedding)
```

#### Setup ChromaDB

```go
chroma, _ := agent.NewChromaStore("http://localhost:8000")
embedding := agent.NewOpenAIEmbedding("text-embedding-3-small", apiKey)

agent := agent.NewOpenAI("gpt-4", apiKey).
    WithVectorRAG(chroma, "my-knowledge-base", embedding)
```

### How It Works

```
1. Store Documents
   ↓
   "Go is a programming language" → [0.2, 0.8, 0.1, ...] (embedding)
   ↓
   Save to Qdrant/ChromaDB

2. User Query
   ↓
   "What is Go?" → [0.19, 0.81, 0.09, ...] (embedding)
   ↓
   Search similar vectors in database
   ↓
   Retrieve top 3 relevant documents
   ↓
   Add to prompt context
   ↓
   Send to LLM
```

### Example Usage

```go
// Store knowledge
docs := []agent.Document{
    {
        ID:      "doc1",
        Content: "Go is a statically typed language",
        Metadata: map[string]interface{}{"topic": "programming"},
    },
    {
        ID:      "doc2",
        Content: "Go was created by Google in 2009",
        Metadata: map[string]interface{}{"topic": "history"},
    },
}

agent.AddDocuments(ctx, docs)

// Query with RAG
response, _ := agent.Ask(ctx, "What is Go?")
// Agent automatically:
// 1. Searches vector store for relevant docs
// 2. Adds found docs to prompt
// 3. Sends enriched prompt to LLM
```

### Vector Store Operations

```go
// Create collection
qdrant.CreateCollection(ctx, "products", &agent.CollectionConfig{
    Dimension:      1536,  // text-embedding-3-small dimension
    DistanceMetric: "cosine",
})

// Add documents
docs := []agent.Document{...}
qdrant.Add(ctx, "products", docs)

// Search by vector
results, _ := qdrant.Search(ctx, &agent.SearchRequest{
    Collection: "products",
    Vector:     queryEmbedding,
    Limit:      5,
})

// Search by text (auto-embedding)
results, _ := qdrant.SearchByText(ctx, &agent.TextSearchRequest{
    Collection: "products",
    Text:       "smartphone",
    Limit:      5,
})

// Delete documents
qdrant.Delete(ctx, "products", []string{"doc1", "doc2"})

// Delete collection
qdrant.DeleteCollection(ctx, "products")
```

### Distance Metrics

```go
// Cosine similarity (default) - good for text
agent.DistanceMetricCosine

// Euclidean distance - geometric distance
agent.DistanceMetricEuclidean

// Dot product - fast but not normalized
agent.DistanceMetricDotProduct
```

---

## 🧩 Part 5: Complete Architecture

### Storage Layers Visualization

```
┌─────────────────────────────────────────────────────────────┐
│                     Your Application                        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   Agent Builder       │
         │   (go-deep-agent)     │
         └─┬─────────┬─────────┬─┘
           │         │         │
    ┌──────▼─┐  ┌────▼────┐  ┌▼────────┐
    │ Memory │  │  Cache  │  │ Vector  │
    │ (RAM)  │  │ (Redis) │  │  Store  │
    └────────┘  └─────────┘  └─────────┘
       ↓             ↓             ↓
    Per-session   Shared      Shared
    Volatile    Persistent  Persistent
```

### What Goes Where?

| Data Type | Memory (RAM) | Cache (Redis) | Vector Store (Qdrant) |
|-----------|--------------|---------------|----------------------|
| **User's name** | ✅ Conversation context | ❌ | ❌ |
| **Last 20 messages** | ✅ Working memory | ❌ | ❌ |
| **API response for "2+2"** | ❌ | ✅ Cached result | ❌ |
| **Product documentation** | ❌ | ❌ | ✅ Knowledge base |
| **Company policies** | ❌ | ❌ | ✅ Searchable docs |

### Example: E-commerce Chatbot

```go
// 1. Vector Store: Product catalog
qdrant.Add(ctx, "products", []agent.Document{
    {ID: "p1", Content: "iPhone 15 Pro - $999"},
    {ID: "p2", Content: "Samsung Galaxy S24 - $899"},
})

// 2. Memory: Conversation history
agent := agent.NewOpenAI("gpt-4", apiKey).
    WithMemory().
    WithMaxHistory(20).
    WithVectorRAG(qdrant, "products", embedding)

// 3. Cache: Common queries
agent.WithRedisCache("localhost:6379", "", 0)

// Usage
agent.Ask(ctx, "Hi, my name is Alice")
// → Stored in Memory (RAM)

agent.Ask(ctx, "Show me phones under $1000")
// → Searches Vector Store (Qdrant) for relevant products
// → Stores response in Cache (Redis)

agent.Ask(ctx, "What's my name?")
// → Retrieved from Memory (RAM): "Alice"

agent.Ask(ctx, "Show me phones under $1000")
// → Retrieved from Cache (Redis): cache HIT, no API call
```

---

## 🎓 Part 6: Common Misconceptions

### ❌ Misconception 1: Redis is for conversation memory

**Wrong:**
```go
agent.WithRedisCache("localhost:6379", "", 0)
// "Now my conversations are saved to Redis!"
```

**Reality:**
- Redis Cache stores **API responses**, not conversation history
- Conversation memory is still in RAM (lost on restart)

**Correct approach:**
```go
// Manual save/restore required
history := agent.GetHistory()
redisClient.Set("user:123:history", json.Marshal(history))
```

---

### ❌ Misconception 2: Vector stores remember conversations

**Wrong:**
```go
agent.WithVectorRAG(qdrant, "chats", embedding)
agent.Ask(ctx, "My name is Alice")
// "Now Qdrant remembers my name!"
```

**Reality:**
- Vector stores are for **document search** (RAG)
- Not designed for conversation memory

**Correct use:**
```go
// Store knowledge base
qdrant.Add(ctx, "docs", []agent.Document{
    {Content: "Company founded in 2020"},
})

// Query knowledge
agent.Ask(ctx, "When was the company founded?")
// → Searches Qdrant, finds "2020"
```

---

### ❌ Misconception 3: HierarchicalMemory persists to disk

**Wrong:**
```go
agent.WithHierarchicalMemory(config)
// "Now my memory is persistent!"
```

**Reality:**
- Still RAM-based (3 tiers: Working, Episodic, Semantic)
- Lost on restart
- Just better organization than flat list

---

### ❌ Misconception 4: Cache and Memory are the same

| Aspect | Memory | Cache |
|--------|--------|-------|
| **Purpose** | Remember conversation | Avoid duplicate API calls |
| **Key** | Implicit (order) | Hash(prompt + params) |
| **Lookup** | Sequential (history) | Key-based (O(1)) |
| **TTL** | No expiration | Time-based expiration |
| **Shared** | No (per-instance) | Yes (with Redis) |

---

## 🛠️ Part 7: Best Practices

### For Conversation Memory

```go
// ✅ Good: Short-lived sessions
agent := agent.NewOpenAI("gpt-4", apiKey).
    WithMemory().
    WithMaxHistory(20)

// ✅ Good: Long-lived with manual persistence
history := agent.GetHistory()
saveToDatabase(userID, history)

// Later...
history := loadFromDatabase(userID)
agent.SetHistory(history)
```

### For Caching

```go
// ✅ Good: Redis for production
agent.WithRedisCache("localhost:6379", "", 0)

// ✅ Good: Memory cache for development
agent.WithMemoryCache(100, 5*time.Minute)

// ❌ Bad: No cache for repetitive queries
// Each identical prompt calls API again ($$$)
```

### For RAG (Vector Stores)

```go
// ✅ Good: Store static knowledge
qdrant.Add(ctx, "docs", companyPolicies)
qdrant.Add(ctx, "products", productCatalog)

// ❌ Bad: Store conversation history
// Use Memory system instead
```

### For Multi-Instance Deployment

```go
// Load Balancer
//      ↓
// ┌─────────┬─────────┬─────────┐
// │ Server1 │ Server2 │ Server3 │
// └────┬────┴────┬────┴────┬────┘
//      │         │         │
//      └─────────┴─────────┘
//              ↓
//      ┌──────────────┐
//      │ Redis Cache  │  ← Shared
//      └──────────────┘
//              ↓
//      ┌──────────────┐
//      │ Qdrant Store │  ← Shared
//      └──────────────┘

// Each server has own Memory (RAM)
// ❌ Problem: User switches servers → loses conversation

// ✅ Solution: Store session in Redis
LoadHistory(sessionID) → SetHistory(messages)
```

---

## 📊 Part 8: Performance Comparison

### Memory vs Cache vs Vector Store

| Operation | Memory (RAM) | Redis Cache | Qdrant |
|-----------|--------------|-------------|--------|
| **Read** | ~1 μs | ~1-2 ms | ~5-10 ms |
| **Write** | ~1 μs | ~1-2 ms | ~10-20 ms |
| **Search 1M docs** | N/A | N/A | ~50-100 ms |
| **Persistence** | ❌ | ✅ | ✅ |
| **Shared** | ❌ | ✅ | ✅ |

### When to Use What

```go
// Scenario 1: Simple chatbot (single session)
agent.WithMemory()  // ✅ Fast, sufficient

// Scenario 2: Production chatbot (multi-server)
agent.WithMemory()  // For current session
saveToRedis(sessionID, history)  // Manual persistence

// Scenario 3: Knowledge-based chatbot
agent.WithMemory().  // Conversation
    WithRedisCache("localhost:6379", "", 0).  // Response caching
    WithVectorRAG(qdrant, "kb", embedding)   // Knowledge base

// Scenario 4: High-traffic API
agent.WithRedisCache("localhost:6379", "", 0)  // Reduce API costs
```

---

## 🔮 Part 9: Future Enhancements (Roadmap)

### Potential Features

```go
// Idea 1: Persistent Memory Backend
agent.WithMemory().
    WithMemoryBackend(redisMemory)  // Auto save/restore

// Idea 2: Session Management
agent.WithMemory().
    WithSessionID("user-123")  // Auto-persist per user

// Idea 3: Hybrid Memory
agent.WithMemory().
    WithMemoryTier(agent.MemoryTierRedis, 20).  // Recent 20 in Redis
    WithMemoryTier(agent.MemoryTierPostgres, 1000)  // Long-term in DB
```

---

## 📖 Part 10: Code Examples

### Example 1: Simple Chatbot

```go
func main() {
    agent := agent.NewOpenAI("gpt-4", apiKey).
        WithMemory().
        WithMaxHistory(20)

    agent.Ask(ctx, "My name is Alice")
    agent.Ask(ctx, "What's my name?")  // → "Alice"
}
```

### Example 2: Production Chatbot

```go
func handleChat(sessionID string, message string) string {
    // Load history from Redis
    history := loadHistory(sessionID)
    
    // Create agent with history
    agent := agent.NewOpenAI("gpt-4", apiKey).
        WithMemory().
        WithMaxHistory(20)
    
    agent.SetHistory(history)
    
    // Process message
    response, _ := agent.Ask(ctx, message)
    
    // Save updated history
    saveHistory(sessionID, agent.GetHistory())
    
    return response
}

func loadHistory(sessionID string) []agent.Message {
    data, _ := redisClient.Get(ctx, "session:"+sessionID).Result()
    var history []agent.Message
    json.Unmarshal([]byte(data), &history)
    return history
}

func saveHistory(sessionID string, history []agent.Message) {
    data, _ := json.Marshal(history)
    redisClient.Set(ctx, "session:"+sessionID, data, 24*time.Hour)
}
```

### Example 3: RAG + Memory + Cache

```go
func main() {
    // Setup vector store
    qdrant, _ := agent.NewQdrantStore("http://localhost:6333")
    embedding := agent.NewOpenAIEmbedding("text-embedding-3-small", apiKey)
    
    // Store knowledge base
    qdrant.Add(ctx, "docs", []agent.Document{
        {ID: "1", Content: "Go was created by Google"},
        {ID: "2", Content: "Go is fast and efficient"},
    })
    
    // Create agent with all features
    agent := agent.NewOpenAI("gpt-4", apiKey).
        WithMemory().                                      // Conversation
        WithMaxHistory(20).
        WithRedisCache("localhost:6379", "", 0).          // Response caching
        WithVectorRAG(qdrant, "docs", embedding)          // Knowledge base
    
    // Usage
    agent.Ask(ctx, "What is Go?")
    // 1. Searches vector store for "Go" docs
    // 2. Adds found docs to prompt
    // 3. Calls LLM
    // 4. Caches response in Redis
    // 5. Stores in conversation memory
}
```

---

## 🎯 Summary Table

| Feature | Type | Storage | Persistence | Shared | Use Case |
|---------|------|---------|-------------|--------|----------|
| **WithMemory()** | Conversation | RAM | ❌ | ❌ | Chat history (single session) |
| **WithHierarchicalMemory()** | Conversation | RAM (3-tier) | ❌ | ❌ | Smart memory management |
| **WithMemoryCache()** | API Cache | RAM | ❌ | ❌ | Dev/testing |
| **WithRedisCache()** | API Cache | Redis | ✅ | ✅ | Production caching |
| **WithVectorRAG()** | Knowledge | Qdrant/Chroma | ✅ | ✅ | RAG, semantic search |

---

## 🆘 FAQs

### Q1: How do I persist conversation history?

**A:** Manual save/restore with `GetHistory()` and `SetHistory()`. See [Example 2](#example-2-production-chatbot).

### Q2: Why is my conversation lost after restart?

**A:** Memory is RAM-based. Enable persistence with manual save/restore or external storage.

### Q3: Can I use Redis for conversation memory?

**A:** Not directly. Redis Cache is for API responses. You can manually save conversation to Redis (see examples).

### Q4: What's the difference between Cache and Memory?

**A:** 
- **Memory** = Conversation history (per-instance, RAM)
- **Cache** = API response caching (shared, Redis/RAM)

### Q5: Should I use Qdrant for conversation history?

**A:** No. Qdrant is for document search (RAG), not conversation memory.

### Q6: How to share conversation across servers?

**A:** Save to external storage (Redis, PostgreSQL) and load on each request. See [Example 2](#example-2-production-chatbot).

### Q7: Does WithHierarchicalMemory persist to disk?

**A:** No. It's still RAM-based, just better organized (3 tiers).

### Q8: What's the recommended setup for production?

**A:**
```go
agent.WithMemory().                               // Conversation (RAM)
    WithRedisCache("localhost:6379", "", 0).     // API caching (Redis)
    WithVectorRAG(qdrant, "kb", embedding)       // Knowledge base (Qdrant)

// + Manual session persistence to Redis/DB
```

---

## 📚 Additional Resources

- [Main README](../README.md) - Quick start guide
- [Memory Migration Guide](./MEMORY_MIGRATION.md) - Upgrading from v0.6.x
- [Rate Limiting Guide](./RATE_LIMITING_GUIDE.md) - API rate management
- [Examples Directory](../examples/) - Working code samples

---

## 📝 Changelog

- **v0.7.10** (2025-11-12) - Initial comprehensive memory system guide
- Document created to clarify Memory vs Cache vs Vector Store confusion

---

**Questions?** Open an issue: https://github.com/taipm/go-deep-agent/issues
