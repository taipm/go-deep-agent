# Redis Cache Guide

Complete guide to using Redis-based distributed caching in go-deep-agent.

## Table of Contents

- [Overview](#overview)
- [When to Use Redis Cache](#when-to-use-redis-cache)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Advanced Features](#advanced-features)
- [Production Best Practices](#production-best-practices)
- [Performance Tuning](#performance-tuning)
- [Troubleshooting](#troubleshooting)

## Overview

Redis cache provides distributed, persistent caching for AI responses across multiple application instances. Unlike memory cache (which is process-local), Redis cache:

- **Shared**: Multiple instances share the same cache
- **Persistent**: Cache survives application restarts
- **Scalable**: Supports Redis Cluster for horizontal scaling
- **Distributed**: Built-in locking prevents cache stampede

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   App 1     │────▶│   Redis     │◀────│   App 2     │
│ (Instance)  │     │   Server    │     │ (Instance)  │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │   App 3     │
                    │ (Instance)  │
                    └─────────────┘
```

All application instances share cached responses through Redis.

## When to Use Redis Cache

### Use Redis Cache When:

✅ **Multiple Application Instances**: Running multiple servers/containers  
✅ **Persistent Cache Required**: Cache should survive restarts  
✅ **High Traffic**: Same queries from different users/instances  
✅ **Distributed Systems**: Microservices architecture  
✅ **Production Deployments**: Need reliability and scalability  

### Use Memory Cache When:

✅ **Single Instance**: Only one application process  
✅ **Low Latency Critical**: Sub-millisecond response time needed  
✅ **Development/Testing**: Quick prototyping without infrastructure  
✅ **Short-lived Data**: Cache doesn't need to persist  

### Performance Comparison

| Cache Type | Hit Latency | Shared | Persistent | Best For |
|------------|-------------|--------|------------|----------|
| **None** | N/A | No | No | Testing, always-fresh data |
| **Memory** | ~50μs | No | No | Single instance, ultra-low latency |
| **Redis** | ~1-5ms | Yes | Yes | Production, distributed systems |

## Quick Start

### 1. Install Redis

**macOS (Homebrew):**
```bash
brew install redis
brew services start redis
```

**Ubuntu/Debian:**
```bash
sudo apt-get install redis-server
sudo systemctl start redis
```

**Docker:**
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

### 2. Simple Setup

```go
package main

import (
    "context"
    "fmt"
    agent "github.com/taipm/go-deep-agent/agent"
)

func main() {
    apiKey := "your-openai-api-key"
    ctx := context.Background()

    // Create agent with Redis cache
    ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithRedisCache("localhost:6379", "", 0)

    // First call - cache miss (slow)
    resp1, _ := ai.Ask(ctx, "What is Go?")
    fmt.Println(resp1) // ~1-2 seconds

    // Second call - cache hit (fast)
    resp2, _ := ai.Ask(ctx, "What is Go?")
    fmt.Println(resp2) // ~5ms (200x faster!)
}
```

### 3. Verify Cache Hit

```go
stats := ai.GetCacheStats()
fmt.Printf("Hits: %d, Misses: %d, Hit Rate: %.2f%%\n",
    stats.Hits, stats.Misses,
    float64(stats.Hits)/(float64(stats.Hits+stats.Misses))*100)
// Output: Hits: 1, Misses: 1, Hit Rate: 50.00%
```

## Configuration

### Basic Configuration

```go
// Simple setup (defaults)
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCache("localhost:6379", "", 0)
    // Address: localhost:6379
    // Password: "" (no auth)
    // DB: 0
    // Default TTL: 5 minutes
    // Pool size: 10
```

### Advanced Configuration

```go
opts := &agent.RedisCacheOptions{
    // Connection
    Addrs:    []string{"localhost:6379"},        // Single node
    // Addrs: []string{"node1:6379", "node2:6379"}, // Cluster mode
    Password: "your-redis-password",             // Optional auth
    DB:       0,                                 // Database (0-15, single node only)

    // Connection Pooling
    PoolSize:     20,                            // Max connections
    MinIdleConns: 10,                            // Min idle connections
    
    // Timeouts
    DialTimeout:  5 * time.Second,               // Connection timeout
    ReadTimeout:  3 * time.Second,               // Read timeout
    WriteTimeout: 3 * time.Second,               // Write timeout
    
    // Cache Settings
    KeyPrefix:   "myapp",                        // Namespace (default: "go-deep-agent")
    DefaultTTL:  10 * time.Minute,               // Entry expiration (default: 5m)
}

ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCacheOptions(opts)
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `Addrs` | `[]string` | `["localhost:6379"]` | Redis server addresses |
| `Password` | `string` | `""` | Authentication password |
| `DB` | `int` | `0` | Database number (single node only) |
| `PoolSize` | `int` | `10` | Maximum connection pool size |
| `MinIdleConns` | `int` | `5` | Minimum idle connections |
| `DialTimeout` | `time.Duration` | `5s` | Connection timeout |
| `ReadTimeout` | `time.Duration` | `3s` | Read operation timeout |
| `WriteTimeout` | `time.Duration` | `3s` | Write operation timeout |
| `KeyPrefix` | `string` | `"go-deep-agent"` | Cache key namespace |
| `DefaultTTL` | `time.Duration` | `5m` | Default entry expiration |

## Advanced Features

### 1. Custom TTL Per Request

```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCache("localhost:6379", "", 0)

// Cache this response for 1 hour
ai.WithCacheTTL(1 * time.Hour).Ask(ctx, "Latest news?")

// Cache this response for 24 hours
ai.WithCacheTTL(24 * time.Hour).Ask(ctx, "Historical facts?")

// Use default TTL (5 minutes)
ai.Ask(ctx, "Regular query")
```

### 2. Temporarily Disable Cache

```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCache("localhost:6379", "", 0)

// Disable cache temporarily
ai.DisableCache()
resp, _ := ai.Ask(ctx, "Fresh response needed")

// Re-enable cache
ai.EnableCache()
resp2, _ := ai.Ask(ctx, "This will be cached")
```

### 3. Cache Statistics

```go
stats := ai.GetCacheStats()
fmt.Printf("Hits: %d\n", stats.Hits)           // Cache hits
fmt.Printf("Misses: %d\n", stats.Misses)       // Cache misses
fmt.Printf("Writes: %d\n", stats.TotalWrites)  // Total writes
fmt.Printf("Size: %d\n", stats.Size)           // Cached entries count

hitRate := float64(stats.Hits) / float64(stats.Hits + stats.Misses) * 100
fmt.Printf("Hit Rate: %.2f%%\n", hitRate)
```

### 4. Clear Cache

```go
// Clear all cached entries for this application
err := ai.ClearCache(ctx)
if err != nil {
    log.Printf("Failed to clear cache: %v", err)
}
```

### 5. Redis Cluster Mode

```go
opts := &agent.RedisCacheOptions{
    Addrs: []string{
        "redis-node1:6379",
        "redis-node2:6379",
        "redis-node3:6379",
    },
    Password:   "cluster-password",
    PoolSize:   50,                    // Higher for cluster
    KeyPrefix:  "myapp",
    DefaultTTL: 15 * time.Minute,
}

ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCacheOptions(opts)
```

### 6. Multi-Tenant Namespacing

```go
// Tenant 1
opts1 := &agent.RedisCacheOptions{
    Addrs:     []string{"localhost:6379"},
    KeyPrefix: "tenant1",  // Isolated namespace
}
ai1 := agent.NewOpenAI("gpt-4o-mini", apiKey).WithRedisCacheOptions(opts1)

// Tenant 2
opts2 := &agent.RedisCacheOptions{
    Addrs:     []string{"localhost:6379"},
    KeyPrefix: "tenant2",  // Separate namespace
}
ai2 := agent.NewOpenAI("gpt-4o-mini", apiKey).WithRedisCacheOptions(opts2)

// Caches are completely isolated
```

## Production Best Practices

### 1. Connection Pooling

```go
opts := &agent.RedisCacheOptions{
    PoolSize:     50,   // For high-traffic apps
    MinIdleConns: 20,   // Keep warm connections
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
}
```

**Guidelines:**
- **PoolSize**: 10-50 for typical apps, 100+ for high traffic
- **MinIdleConns**: 50% of PoolSize to reduce connection overhead
- **Timeouts**: 3-5 seconds for network operations

### 2. TTL Strategy

```go
// Short TTL for frequently changing data
ai.WithCacheTTL(5 * time.Minute).Ask(ctx, "Current weather?")

// Long TTL for static data
ai.WithCacheTTL(24 * time.Hour).Ask(ctx, "What is the capital of France?")

// No TTL for permanent data (not recommended)
// ai.WithCacheTTL(0).Ask(ctx, "...")
```

**Recommendations:**
- **Real-time data**: 1-5 minutes
- **Semi-static data**: 15-60 minutes
- **Static facts**: 1-24 hours
- **Always set TTL**: Prevents unbounded memory growth

### 3. Error Handling

```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCache("localhost:6379", "", 0)

resp, err := ai.Ask(ctx, "Your question")
if err != nil {
    log.Printf("Request failed: %v", err)
    // Cache errors are logged but don't fail the request
    return
}

// Check cache health
if err := ai.ClearCache(ctx); err != nil {
    log.Printf("Redis connection issue: %v", err)
    // Consider switching to memory cache or alerting
}
```

### 4. Monitoring

```go
// Periodic cache monitoring
ticker := time.NewTicker(1 * time.Minute)
go func() {
    for range ticker.C {
        stats := ai.GetCacheStats()
        hitRate := float64(stats.Hits) / float64(stats.Hits + stats.Misses) * 100
        
        log.Printf("Cache Stats - Hits: %d, Misses: %d, Hit Rate: %.2f%%, Size: %d",
            stats.Hits, stats.Misses, hitRate, stats.Size)
        
        // Alert if hit rate drops below threshold
        if hitRate < 50.0 && stats.Hits+stats.Misses > 100 {
            log.Printf("WARNING: Low cache hit rate: %.2f%%", hitRate)
        }
    }
}()
```

### 5. Redis Security

```go
opts := &agent.RedisCacheOptions{
    Addrs:    []string{"redis.production.com:6379"},
    Password: os.Getenv("REDIS_PASSWORD"),  // From env var
    DB:       0,
    
    // TLS for production (requires redis.Options customization)
    // TLSConfig: &tls.Config{...},
}

ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCacheOptions(opts)
```

**Security Checklist:**
- ✅ Always use password in production
- ✅ Enable TLS for network encryption
- ✅ Use firewall rules to restrict Redis access
- ✅ Store credentials in environment variables, not code
- ✅ Use separate Redis instances per environment (dev/staging/prod)

## Performance Tuning

### 1. Optimize Cache Hit Rate

```go
// Use consistent prompts for better cache hits
basePrompt := "Explain %s in simple terms"
ai.Ask(ctx, fmt.Sprintf(basePrompt, "quantum computing"))
ai.Ask(ctx, fmt.Sprintf(basePrompt, "machine learning"))

// Avoid random variations that prevent cache hits
// ❌ Bad: ai.Ask(ctx, "Explain quantum computing (id: " + uuid.New() + ")")
// ✅ Good: ai.Ask(ctx, "Explain quantum computing")
```

### 2. Batch Operations

```go
// Process multiple queries efficiently
questions := []string{
    "What is Go?",
    "What is Python?",
    "What is Rust?",
}

for _, q := range questions {
    resp, _ := ai.Ask(ctx, q)
    fmt.Println(resp)
    // Subsequent calls are cached automatically
}
```

### 3. Reduce Redis Latency

```go
opts := &agent.RedisCacheOptions{
    Addrs:        []string{"localhost:6379"},  // Use local Redis for lowest latency
    PoolSize:     50,                          // Larger pool = less connection overhead
    MinIdleConns: 25,                          // Keep connections warm
    ReadTimeout:  1 * time.Second,             // Shorter timeout for local Redis
    WriteTimeout: 1 * time.Second,
}
```

**Latency Optimization:**
- **Co-locate Redis**: Same datacenter/region as application
- **Use Redis Cluster**: Distribute load across nodes
- **Monitor slow queries**: Redis `SLOWLOG` command
- **Optimize key prefix**: Shorter = faster serialization

### 4. Memory Management

```go
// Monitor Redis memory usage
// redis-cli INFO memory

opts := &agent.RedisCacheOptions{
    DefaultTTL: 10 * time.Minute,  // Entries auto-expire
    KeyPrefix:  "app",              // Namespace for easy cleanup
}

// Periodic cleanup (if needed)
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        ai.ClearCache(ctx)  // Clear all entries
        log.Println("Cache cleared")
    }
}()
```

**Redis Memory Configuration:**
```bash
# In redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru  # Evict least recently used keys
```

## Troubleshooting

### Connection Errors

**Problem:** `dial tcp: connect: connection refused`

**Solutions:**
```bash
# Check if Redis is running
redis-cli ping
# Expected: PONG

# Start Redis
redis-server

# Check Redis logs
tail -f /usr/local/var/log/redis.log  # macOS
tail -f /var/log/redis/redis-server.log  # Linux
```

### Authentication Errors

**Problem:** `NOAUTH Authentication required`

**Solution:**
```go
// Provide password
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRedisCache("localhost:6379", "your-password", 0)
```

### Slow Cache Performance

**Problem:** Cache hits are slower than expected

**Diagnostics:**
```go
import "time"

start := time.Now()
resp, _ := ai.Ask(ctx, "What is Go?")
duration := time.Since(start)
fmt.Printf("Cache hit took: %v\n", duration)

// Expected for Redis: 1-5ms (local), 5-20ms (remote)
// If > 50ms, investigate network latency or Redis load
```

**Solutions:**
- Use local Redis for development
- Check Redis `SLOWLOG` for slow operations
- Increase `PoolSize` to reduce connection wait
- Use Redis Cluster for horizontal scaling

### Cache Misses

**Problem:** Low cache hit rate

**Debug:**
```go
stats := ai.GetCacheStats()
fmt.Printf("Hit rate: %.2f%%\n",
    float64(stats.Hits)/(float64(stats.Hits+stats.Misses))*100)

// Check cache keys in Redis
// redis-cli KEYS "go-deep-agent:cache:*"
```

**Common Causes:**
- Varying prompts (add/remove whitespace, punctuation)
- Different temperature/model settings
- Cache TTL too short
- Cache cleared between requests

**Solutions:**
- Normalize prompts before caching
- Use consistent model/temperature settings
- Increase TTL for stable data
- Check `ClearCache()` calls

### Memory Issues

**Problem:** Redis using too much memory

**Monitor:**
```bash
redis-cli INFO memory
# used_memory_human: 1.5G
# maxmemory: 2G
# maxmemory_policy: allkeys-lru
```

**Solutions:**
```go
// Reduce TTL
opts := &agent.RedisCacheOptions{
    DefaultTTL: 5 * time.Minute,  // Shorter TTL = less memory
}

// Clear cache periodically
ai.ClearCache(ctx)

// Use Redis eviction policy
// redis.conf: maxmemory-policy allkeys-lru
```

### Cluster Connection Issues

**Problem:** Can't connect to Redis Cluster

**Solution:**
```go
opts := &agent.RedisCacheOptions{
    Addrs: []string{
        "node1:6379",
        "node2:6379",
        "node3:6379",  // Add all cluster nodes
    },
    // Don't specify DB for cluster mode
    Password: "cluster-password",
}
```

## Examples

See [examples/cache_redis_example.go](../examples/cache_redis_example.go) for comprehensive examples:

1. **Simple Setup**: Basic Redis cache with localhost
2. **Advanced Configuration**: Custom pool size, prefix, TTL
3. **Statistics Tracking**: Monitor hits, misses, hit rate
4. **Batch Operations**: Process multiple queries efficiently
5. **Pattern Deletion**: Clear all cache entries
6. **Distributed Locking**: SetNX for cache stampede prevention
7. **Performance Comparison**: Benchmark no cache vs memory vs Redis
8. **TTL Management**: Default, custom, disable/enable cache

## Related Documentation

- [Caching Guide](./CACHING.md) - General caching concepts
- [RAG Guide](./RAG_VECTOR_DATABASES.md) - Combine caching with RAG
- [Batch Processing](../examples/batch_processing.go) - Cache with batch operations

## Summary

Redis cache provides:

✅ **Distributed caching** across multiple instances  
✅ **Persistent cache** surviving restarts  
✅ **Scalability** with Redis Cluster  
✅ **Production-ready** with connection pooling, timeouts  
✅ **Flexible TTL** per request or default  
✅ **Statistics tracking** for monitoring  
✅ **Distributed locking** to prevent cache stampede  

**Use Redis cache for production deployments with multiple instances or when persistence is required.**
