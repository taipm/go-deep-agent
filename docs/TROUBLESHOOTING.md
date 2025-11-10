# Troubleshooting Guide

Quick reference for common errors and how to fix them.

## Table of Contents

1. [Quick Diagnostic Checklist](#quick-diagnostic-checklist)
2. [Top 10 Common Errors](#top-10-common-errors)
3. [Error Handling Best Practices](#error-handling-best-practices)
4. [Getting Help](#getting-help)

---

## Quick Diagnostic Checklist

Run this checklist before diving into specific errors:

- [ ] **API key is set**: `echo $OPENAI_API_KEY`
- [ ] **Network connection works**: `ping api.openai.com`
- [ ] **Latest version installed**: `go get -u github.com/taipm/go-deep-agent`
- [ ] **OpenAI service status**: https://status.openai.com
- [ ] **Enable debug mode**: Add `.WithDebug()` to see detailed logs
- [ ] **Check Go version**: `go version` (requires Go 1.19+)

---

## Top 10 Common Errors

### 1. API_KEY_MISSING ⭐⭐⭐⭐⭐

**Error message:**

```
API key is missing or invalid

Fix:
  1. Set environment variable: export OPENAI_API_KEY="sk-..."
  2. Or pass to constructor: agent.NewOpenAI("gpt-4", "sk-...")
  3. Get your key: https://platform.openai.com/api-keys
```

**Symptoms:**

- 401 Unauthorized responses
- "invalid_api_key" errors
- Authentication failures

**Root causes:**

1. `OPENAI_API_KEY` environment variable not set
2. API key passed to `NewOpenAI()` is empty or wrong
3. API key has incorrect format

**Solutions:**

**Option 1: Environment variable** (Recommended for production)

```bash
# Linux/Mac
export OPENAI_API_KEY="sk-proj-..."

# Windows PowerShell
$env:OPENAI_API_KEY="sk-proj-..."

# Add to ~/.bashrc or ~/.zshrc for persistence
echo 'export OPENAI_API_KEY="sk-proj-..."' >> ~/.bashrc
```

**Option 2: Pass directly in code** (For development/testing)

```go
ai := agent.NewOpenAI("gpt-4", "sk-proj-...")
```

**Option 3: Load from .env file**

```go
import "github.com/joho/godotenv"

godotenv.Load()
apiKey := os.Getenv("OPENAI_API_KEY")
ai := agent.NewOpenAI("gpt-4", apiKey)
```

**Verify your API key:**

- Must start with `sk-proj-` (new format) or `sk-` (old format)
- Get from: https://platform.openai.com/api-keys
- Test with: `curl https://api.openai.com/v1/models -H "Authorization: Bearer $OPENAI_API_KEY"`

---

### 2. RATE_LIMIT_EXCEEDED ⭐⭐⭐⭐⭐

**Error message:**

```
rate limit exceeded - too many requests

Fix:
  1. Use .WithDefaults() - includes retry with exponential backoff
  2. Or configure: .WithRetry(5).WithRetryDelay(2*time.Second).WithExponentialBackoff()
  3. Upgrade tier: https://platform.openai.com/account/limits
  4. Use caching: .WithRedisCache("localhost:6379", "", 0)
```

**Symptoms:**

- 429 Too Many Requests
- Intermittent failures during high load
- "Rate limit reached" messages

**Root causes:**

1. Too many requests in short time
2. Exceeding tier limits (Free: 3 RPM, Tier 1: 500 RPM, etc.)
3. Multiple agents/services sharing same API key
4. No retry logic configured

**Solutions:**

**✅ Best practice: Use WithDefaults()** (Recommended)

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithDefaults()  // Includes: retry(3) + backoff + timeout(30s)

resp, err := ai.Ask(ctx, "Hello")
// Automatically retries on rate limit with exponential backoff
```

**Advanced retry configuration:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithRetry(5).                        // Retry up to 5 times
    WithRetryDelay(2 * time.Second).     // Start with 2s delay
    WithExponentialBackoff().            // Increase: 2s, 4s, 8s, 16s, 32s
    WithTimeout(60 * time.Second)        // Max 60s total
```

**Check when you can retry:**

```go
resp, err := ai.Ask(ctx, "Hello")
if err != nil {
    if agent.IsRateLimitError(err) {
        fmt.Println("Rate limited - will retry automatically")
        // WithDefaults() handles this automatically
    }
}
```

**Long-term solutions:**

1. **Upgrade OpenAI tier**: https://platform.openai.com/account/limits
   - Free: 3 RPM, 200 RPD
   - Tier 1: 500 RPM, 10,000 TPM
   - Tier 2: 5,000 RPM, 450,000 TPM

2. **Implement caching**:

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithDefaults().
    WithRedisCache("localhost:6379", "", 0)  // Cache responses

// Identical requests use cache (no API call)
resp1, _ := ai.Ask(ctx, "What is Go?")  // API call
resp2, _ := ai.Ask(ctx, "What is Go?")  // From cache
```

3. **Batch similar requests**:

```go
// Instead of 100 individual calls
results, _ := ai.Batch(ctx, messages, agent.BatchConfig{
    MaxConcurrency: 5,  // Limit concurrent requests
})
```

4. **Load balance across multiple keys**:

```go
keys := []string{"sk-key1", "sk-key2", "sk-key3"}
agents := make([]*agent.Builder, len(keys))
for i, key := range keys {
    agents[i] = agent.NewOpenAI("gpt-4", key).WithDefaults()
}

// Round-robin or random selection
agent := agents[requestCount % len(agents)]
```

---

### 3. REQUEST_TIMEOUT ⭐⭐⭐⭐

**Error message:**

```
request timeout - operation took too long

Fix:
  1. Increase timeout: .WithTimeout(60 * time.Second)
  2. Use streaming for long responses: .Stream(...)
  3. Check network connection
  4. Check OpenAI status: https://status.openai.com
```

**Symptoms:**

- Requests hang and timeout
- "context deadline exceeded" errors
- Slow responses

**Root causes:**

1. Default timeout (30s) too short for complex requests
2. Network latency
3. Large responses taking long to generate
4. OpenAI API experiencing issues

**Solutions:**

**Increase timeout:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithTimeout(60 * time.Second)  // 60 seconds

// Or use context with deadline
ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
defer cancel()

resp, err := ai.Ask(ctx, "Explain quantum computing in detail")
```

**Use streaming for long responses:**

```go
ai := agent.NewOpenAI("gpt-4", key).WithDefaults()

// Stream incrementally (no timeout issues)
err := ai.Stream(ctx, "Write a long essay", func(chunk string) {
    fmt.Print(chunk)  // Print as it generates
})
```

**Debug timeout issues:**

```go
start := time.Now()
resp, err := ai.Ask(ctx, message)
duration := time.Since(start)

if err != nil {
    if agent.IsTimeoutError(err) {
        fmt.Printf("Timed out after %v\n", duration)
        // Try with longer timeout or streaming
    }
}
```

**Check network:**

```bash
# Test connectivity to OpenAI
curl -w "@-" -o /dev/null -s "https://api.openai.com/v1/models" \
  -H "Authorization: Bearer $OPENAI_API_KEY" <<'EOF'
    time_total:  %{time_total}s\n
EOF

# Should complete in < 2 seconds
```

---

### 4. TOOL_EXECUTION_FAILED ⭐⭐⭐⭐

**Error message:**

```
tool execution failed

Fix:
  1. Enable debug logging: .WithDebug()
  2. Check tool function implementation
  3. Verify tool parameters match JSON schema
  4. Add error handling in tool function
  5. Increase tool timeout: .WithToolTimeout(60*time.Second)
```

**Symptoms:**

- Tool calls fail with errors
- Panic in tool functions
- "tool panicked" messages
- Tool timeout errors

**Root causes:**

1. Tool function has bugs or panics
2. Invalid parameters passed to tool
3. Tool timeout (default: 30s)
4. Missing error handling in tool function

**Solutions:**

**Enable debug mode to see details:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithDebug().  // See tool execution details
    WithTool(agent.Tool{
        Name: "get_weather",
        Description: "Get current weather",
        Handler: func(args string) (string, error) {
            // Your tool logic
            return "Sunny, 72°F", nil
        },
    })
```

**Add proper error handling:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithTool(agent.Tool{
        Name: "search_database",
        Handler: func(args string) (string, error) {
            // Parse arguments safely
            var params struct {
                Query string `json:"query"`
            }
            if err := json.Unmarshal([]byte(args), &params); err != nil {
                return "", fmt.Errorf("invalid arguments: %w", err)
            }
            
            // Validate input
            if params.Query == "" {
                return "", fmt.Errorf("query cannot be empty")
            }
            
            // Execute with error handling
            result, err := db.Search(params.Query)
            if err != nil {
                return "", fmt.Errorf("search failed: %w", err)
            }
            
            return result, nil
        },
    })
```

**Handle tool timeout:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithToolTimeout(60 * time.Second).  // Increase timeout
    WithTool(agent.Tool{
        Name: "slow_operation",
        Handler: func(args string) (string, error) {
            // Long-running operation
            time.Sleep(45 * time.Second)
            return "Done", nil
        },
    })
```

**Test tools independently:**

```go
// Test tool before using with agent
tool := agent.Tool{
    Name: "calculator",
    Handler: func(args string) (string, error) {
        var params struct {
            Operation string `json:"operation"`
            A, B      float64 `json:"a,b"`
        }
        json.Unmarshal([]byte(args), &params)
        
        switch params.Operation {
        case "add":
            return fmt.Sprintf("%.2f", params.A + params.B), nil
        default:
            return "", fmt.Errorf("unknown operation: %s", params.Operation)
        }
    },
}

// Test directly
result, err := tool.Handler(`{"operation":"add","a":5,"b":3}`)
if err != nil {
    log.Fatal("Tool test failed:", err)
}
fmt.Println("Tool works:", result)  // "8.00"
```

---

### 5. CACHE_CONNECTION_FAILED ⭐⭐⭐

**Error message:**

```
failed to connect to Redis: connection refused

Fix:
  1. Check Redis is running: redis-cli ping
  2. Verify connection: redis://localhost:6379
  3. Check firewall/network settings
  4. Start Redis: redis-server or docker run -p 6379:6379 redis
```

**Symptoms:**

- Cache initialization fails
- "connection refused" errors
- Redis not reachable

**Root causes:**

1. Redis server not running
2. Wrong connection string
3. Firewall blocking port 6379
4. Redis requires authentication

**Solutions:**

**Check Redis status:**

```bash
# Test Redis connection
redis-cli ping
# Expected: PONG

# Check if Redis is running
ps aux | grep redis-server

# Check port
netstat -an | grep 6379
```

**Start Redis:**

```bash
# Option 1: Direct
redis-server

# Option 2: Docker
docker run --name redis -p 6379:6379 -d redis

# Option 3: Docker Compose
docker-compose up -d redis
```

**Configure Redis cache:**

```go
import "time"

ai := agent.NewOpenAI("gpt-4", key).
    WithDefaults().
    WithRedisCache("localhost:6379", "", 0)  // host:port, password, db

// With password
ai.WithRedisCache("localhost:6379", "mypassword", 0)

// With custom TTL
ai.WithRedisCacheWithTTL("localhost:6379", "", 0, 1*time.Hour)
```

**Handle cache errors gracefully:**

```go
ai := agent.NewOpenAI("gpt-4", key).WithDefaults()

// Try to enable cache, but don't fail if Redis unavailable
err := ai.WithRedisCache("localhost:6379", "", 0)
if err != nil {
    log.Printf("Warning: Cache unavailable, continuing without cache: %v", err)
    // Agent still works, just without caching
}

resp, err := ai.Ask(ctx, "Hello")  // Works with or without cache
```

---

### 6. VECTOR_STORE_NOT_CONFIGURED ⭐⭐⭐

**Error message:**

```
no embedding provided and no embedding provider configured

Fix:
  1. Set embedding provider: NewChromaStore(url).WithEmbedding(embedder)
  2. Or provide embeddings: doc.Embedding = []float32{...}
  3. Example: embedder, _ := agent.NewOpenAIEmbedding("text-embedding-3-small", apiKey)
```

**Symptoms:**

- Vector store operations fail
- "no embedding provider configured" errors
- Cannot add documents to RAG

**Root causes:**

1. Vector store created without embedding provider
2. Trying to add documents without embeddings
3. Missing OpenAI embedding API key

**Solutions:**

**Configure vector store with embedding:**

```go
// 1. Create embedding provider
embedder, err := agent.NewOpenAIEmbedding(
    agent.EmbeddingModelSmall,  // "text-embedding-3-small"
    os.Getenv("OPENAI_API_KEY"),
)
if err != nil {
    log.Fatal(err)
}

// 2. Create vector store with embedder
store := agent.NewChromaStore("http://localhost:8000").
    WithEmbedding(embedder)

// 3. Add documents (embeddings auto-generated)
docs := []*agent.VectorDocument{
    {Content: "Go is a programming language"},
    {Content: "Python is popular for AI"},
}

ids, err := store.Add(context.Background(), "my_collection", docs)
```

**Use with RAG agent:**

```go
embedder, _ := agent.NewOpenAIEmbedding(agent.EmbeddingModelSmall, apiKey)
store := agent.NewChromaStore("http://localhost:8000").WithEmbedding(embedder)

ai := agent.NewOpenAI("gpt-4", apiKey).
    WithDefaults().
    WithVectorRAG(store, embedder, agent.DefaultRAGConfig())

// Now can do RAG queries
resp, err := ai.Ask(ctx, "What is Go?")
// Automatically retrieves relevant docs and includes in context
```

**Start ChromaDB:**

```bash
# Option 1: Docker
docker run -p 8000:8000 chromadb/chroma

# Option 2: Python
pip install chromadb
chroma run --path ./chroma_data
```

**Alternative: Provide embeddings manually:**

```go
// If you have pre-computed embeddings
docs := []*agent.VectorDocument{
    {
        Content: "Document text",
        Embedding: []float32{0.1, 0.2, 0.3, ...},  // 1536 dimensions
    },
}

store.Add(ctx, "collection", docs)  // No embedder needed
```

---

### 7. CONTENT_REFUSED ⭐⭐⭐

**Error message:**

```
content refused by model - policy violation

Fix:
  1. Review policies: https://openai.com/policies/usage-policies
  2. Rephrase your prompt to avoid policy violations
  3. Check content filters and safety settings
```

**Symptoms:**

- Model refuses to respond
- "content_policy_violation" errors
- Responses about harmful content

**Root causes:**

1. Prompt violates OpenAI's usage policies
2. Content involves harmful/illegal activities
3. Sensitive personal information requested

**Solutions:**

**Review and rephrase prompt:**

```go
// ❌ Likely to be refused
resp, err := ai.Ask(ctx, "How to hack into a system")
if err != nil {
    if agent.IsRefusalError(err) {
        // Content refused
    }
}

// ✅ Rephrased to educational context
resp, err = ai.Ask(ctx, "Explain cybersecurity concepts for educational purposes")
```

**Handle refusals gracefully:**

```go
resp, err := ai.Ask(ctx, userInput)
if err != nil {
    if agent.IsRefusalError(err) {
        return "I cannot assist with that request. Please try rephrasing."
    }
    return fmt.Sprintf("Error: %v", err)
}
return resp
```

**Review OpenAI policies:**

- Usage Policies: https://openai.com/policies/usage-policies
- Avoid: Illegal activities, harmful content, privacy violations
- Safe: Educational content, general information, creative writing

---

### 8. INVALID_RESPONSE ⭐⭐

**Error message:**

```
invalid response from API

Fix:
  1. Enable debug mode: .WithDebug() to see raw response
  2. Check OpenAI status: https://status.openai.com
  3. Verify API key has proper permissions
  4. Update library: go get -u github.com/taipm/go-deep-agent
```

**Symptoms:**

- Unexpected API response format
- JSON parsing errors
- Null or empty responses

**Root causes:**

1. OpenAI API changes or issues
2. Outdated library version
3. Network corruption
4. API key lacks permissions

**Solutions:**

**Enable debug to see raw response:**

```go
ai := agent.NewOpenAI("gpt-4", key).WithDebug()

resp, err := ai.Ask(ctx, "Hello")
if err != nil {
    if agent.IsInvalidResponseError(err) {
        // Check debug output for raw API response
        log.Printf("Invalid response: %v", err)
    }
}
```

**Update library:**

```bash
# Update to latest version
go get -u github.com/taipm/go-deep-agent

# Clear module cache if needed
go clean -modcache
go mod tidy
```

**Check OpenAI status:**

- Status page: https://status.openai.com
- If there's an outage, wait and retry

**Verify API key permissions:**

```bash
# Test with curl
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "test"}]
  }'
```

---

### 9. MAX_RETRIES_EXCEEDED ⭐⭐

**Error message:**

```
maximum retry attempts exceeded

Fix:
  1. Increase retries: .WithRetry(5) or .WithRetry(10)
  2. Check root cause - enable debug: .WithDebug()
  3. Increase retry delay: .WithRetryDelay(5*time.Second)
  4. Use exponential backoff: .WithExponentialBackoff()
```

**Symptoms:**

- All retry attempts fail
- Persistent errors despite retries
- Long delays before final failure

**Root causes:**

1. Underlying issue not transient (e.g., invalid API key)
2. Insufficient retry attempts
3. Retry delay too short
4. No exponential backoff

**Solutions:**

**Increase retry attempts:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithRetry(10).                       // More retries
    WithRetryDelay(5 * time.Second).     // Longer initial delay
    WithExponentialBackoff()             // Backoff: 5s, 10s, 20s, 40s...
```

**Debug root cause:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithDebug().  // See each retry attempt
    WithRetry(5)

resp, err := ai.Ask(ctx, "Hello")
if err != nil {
    if agent.IsMaxRetriesError(err) {
        // Check debug logs to see why retries failed
        // If same error every time, it's not transient
        log.Printf("All retries failed: %v", err)
    }
}
```

**Distinguish transient vs persistent errors:**

```go
resp, err := ai.Ask(ctx, message)
if err != nil {
    switch {
    case agent.IsRateLimitError(err):
        // Transient - retry with longer delay
        time.Sleep(10 * time.Second)
        
    case agent.IsAPIKeyError(err):
        // Persistent - fix API key, don't retry
        return fmt.Errorf("invalid API key - please check configuration")
        
    case agent.IsTimeoutError(err):
        // Transient - retry with longer timeout
        // Use context with increased deadline
        
    default:
        return err
    }
}
```

---

### 10. MEMORY_FULL ⭐⭐

**Error message:**

```
memory capacity full

Fix:
  - Increase capacity: .WithMaxHistory(100)
  - Or disable memory: .WithMemory(nil)
```

**Symptoms:**

- Cannot add more messages to history
- "memory capacity full" warnings
- Old messages not appearing in context

**Root causes:**

1. Default memory limit (20 messages) reached
2. Long conversations exceed capacity
3. Memory not being cleared between sessions

**Solutions:**

**Increase memory capacity:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithMaxHistory(100)  // Store up to 100 messages
```

**Use WithDefaults() (includes 20 message memory):**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithDefaults()  // Includes WithMaxHistory(20)
```

**Disable memory if not needed:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithMemory(nil)  // Stateless - no memory

// Each request is independent
resp1, _ := ai.Ask(ctx, "My name is Alice")
resp2, _ := ai.Ask(ctx, "What's my name?")  
// Won't know name (no memory)
```

**Clear memory manually:**

```go
ai := agent.NewOpenAI("gpt-4", key).
    WithMaxHistory(20)

// After each conversation
ai.ClearMemory()

// Or clear specific ranges
ai.ClearMemoryRange(0, 10)  // Clear first 10 messages
```

**Use sliding window:**

```go
// Keep only last N messages
ai := agent.NewOpenAI("gpt-4", key).
    WithMaxHistory(10)  // Only last 10 messages

// Older messages automatically dropped
for i := 0; i < 100; i++ {
    ai.Ask(ctx, fmt.Sprintf("Message %d", i))
    // Only last 10 kept in memory
}
```

---

## Error Handling Best Practices

### 1. Always Check Errors

```go
// ❌ Bad - ignoring errors
resp, _ := ai.Ask(ctx, "Hello")

// ✅ Good - handle errors
resp, err := ai.Ask(ctx, "Hello")
if err != nil {
    log.Printf("Error: %v", err)
    return err
}
```

### 2. Use Error Type Checking

```go
resp, err := ai.Ask(ctx, message)
if err != nil {
    switch {
    case agent.IsRateLimitError(err):
        // Handle rate limit specifically
        log.Println("Rate limited - retrying with delay")
        
    case agent.IsTimeoutError(err):
        // Handle timeout
        log.Println("Timed out - try with longer timeout")
        
    case agent.IsAPIKeyError(err):
        // Handle auth errors
        log.Println("Authentication failed - check API key")
        
    default:
        // Generic error handling
        log.Printf("Request failed: %v", err)
    }
    return err
}
```

### 3. Use WithDefaults() for Production

```go
// ✅ Production-ready configuration
ai := agent.NewOpenAI("gpt-4", apiKey).
    WithDefaults()  // Includes: retry, backoff, timeout, memory

// Handles most common errors automatically:
// - Rate limits: Retries with exponential backoff
// - Timeouts: 30s default
// - Memory: 20 messages
```

### 4. Enable Debug Mode During Development

```go
// During development
ai := agent.NewOpenAI("gpt-4", apiKey).
    WithDebug().     // See detailed logs
    WithDefaults()

// Remove .WithDebug() in production
```

### 5. Implement Graceful Degradation

```go
func askWithFallback(ai *agent.Builder, ctx context.Context, msg string) string {
    // Try with GPT-4
    resp, err := ai.Ask(ctx, msg)
    if err != nil {
        log.Printf("GPT-4 failed: %v, trying GPT-3.5", err)
        
        // Fallback to GPT-3.5 Turbo
        fallbackAI := agent.NewOpenAI("gpt-3.5-turbo", apiKey).WithDefaults()
        resp, err = fallbackAI.Ask(ctx, msg)
        if err != nil {
            return "Sorry, I'm experiencing technical difficulties."
        }
    }
    return resp
}
```

---

## Getting Help

### Community Support

- **GitHub Issues**: https://github.com/taipm/go-deep-agent/issues
- **Discussions**: https://github.com/taipm/go-deep-agent/discussions
- **Examples**: https://github.com/taipm/go-deep-agent/tree/main/examples

### Documentation

- **README**: https://github.com/taipm/go-deep-agent#readme
- **API Reference**: https://pkg.go.dev/github.com/taipm/go-deep-agent/agent
- **Changelog**: https://github.com/taipm/go-deep-agent/blob/main/CHANGELOG.md

### OpenAI Resources

- **OpenAI Documentation**: https://platform.openai.com/docs
- **Status Page**: https://status.openai.com
- **Community Forum**: https://community.openai.com
- **Rate Limits**: https://platform.openai.com/account/limits
- **Usage Policies**: https://openai.com/policies/usage-policies

### Before Asking for Help

Include this information when reporting issues:

1. **Go version**: `go version`
2. **Library version**: Check `go.mod`
3. **Error message**: Full error output
4. **Code sample**: Minimal reproducible example
5. **Debug logs**: Output with `.WithDebug()` enabled
6. **OpenAI status**: Check https://status.openai.com

**Example issue template:**

```markdown
**Describe the issue:**
[Clear description of what's happening]

**Environment:**
- Go version: 1.21.0
- go-deep-agent version: v0.5.8
- OS: macOS 14.0

**Code to reproduce:**
```go
ai := agent.NewOpenAI("gpt-4", apiKey).WithDefaults()
resp, err := ai.Ask(ctx, "Hello")
// Error: ...
```

**Error output:**
```
[Full error message and stack trace]
```

**Additional context:**
- OpenAI status: Normal
- Tried solutions: [What you've already tried]
```

---

**Last updated**: November 10, 2025  
**Version**: v0.5.9 (LEAN error handling improvements)

Found this helpful? ⭐ Star the repo: https://github.com/taipm/go-deep-agent
