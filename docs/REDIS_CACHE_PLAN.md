# Redis Cache Backend Implementation Plan

## Overview
Add Redis-based caching backend for distributed, persistent response caching in production environments.

## Current State (v0.4.0)
- ✅ `Cache` interface defined
- ✅ `MemoryCache` implementation (LRU, TTL)
- ✅ Cache statistics tracking
- ✅ Integrated into Ask() method
- ✅ Builder methods: WithCache(), WithMemoryCache(), WithCacheTTL()

**Limitations:**
- Single-process only (no sharing across instances)
- Lost on restart (no persistence)
- Limited by RAM
- No distributed locking

## Goals (v0.5.0)

### 1. Redis Cache Implementation

#### 1.1 Core Features
```go
type RedisCache struct {
    client     redis.UniversalClient
    prefix     string        // Key prefix for namespacing
    defaultTTL time.Duration
    stats      *CacheStats
    statsLock  sync.RWMutex
}

// Options for Redis cache
type RedisCacheOptions struct {
    // Redis connection
    Addrs      []string      // Redis addresses (cluster mode)
    Password   string        // Redis password
    DB         int           // Database number
    
    // Pooling
    PoolSize      int         // Connection pool size
    MinIdleConns  int         // Minimum idle connections
    
    // Timeouts
    DialTimeout  time.Duration
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    
    // Cache config
    KeyPrefix  string        // Namespace prefix
    DefaultTTL time.Duration
    
    // Advanced
    EnableTracing  bool       // OpenTelemetry tracing
    EnableMetrics  bool       // Prometheus metrics
}
```

#### 1.2 Builder API
```go
// Simple Redis cache
builder.WithRedisCache("localhost:6379", "", 0)

// With options
builder.WithRedisCacheOptions(&agent.RedisCacheOptions{
    Addrs:       []string{"localhost:6379"},
    Password:    "",
    DB:          0,
    PoolSize:    10,
    KeyPrefix:   "go-deep-agent:",
    DefaultTTL:  10 * time.Minute,
})

// Redis Cluster
builder.WithRedisCacheOptions(&agent.RedisCacheOptions{
    Addrs: []string{
        "redis-1:6379",
        "redis-2:6379",
        "redis-3:6379",
    },
    DefaultTTL: 5 * time.Minute,
})

// Redis Sentinel
builder.WithRedisSentinelCache(&agent.RedisSentinelOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{"sentinel-1:26379", "sentinel-2:26379"},
    DefaultTTL:    5 * time.Minute,
})
```

### 2. Implementation Details

#### 2.1 Files Structure

**agent/cache_redis.go** (~400 LOC)
```go
// RedisCache implementation
type RedisCache struct {
    client     redis.UniversalClient
    prefix     string
    defaultTTL time.Duration
    stats      *CacheStats
    statsLock  sync.RWMutex
}

// Constructor
func NewRedisCache(addr, password string, db int, defaultTTL time.Duration) (*RedisCache, error)
func NewRedisCacheWithOptions(opts *RedisCacheOptions) (*RedisCache, error)

// Cache interface implementation
func (c *RedisCache) Get(ctx context.Context, key string) (string, bool, error)
func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error
func (c *RedisCache) Delete(ctx context.Context, key string) error
func (c *RedisCache) Clear(ctx context.Context) error
func (c *RedisCache) Stats() CacheStats

// Advanced operations
func (c *RedisCache) GetWithTTL(ctx context.Context, key string) (string, time.Duration, bool, error)
func (c *RedisCache) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
func (c *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error)
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error)
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error

// Batch operations
func (c *RedisCache) MGet(ctx context.Context, keys ...string) ([]string, error)
func (c *RedisCache) MSet(ctx context.Context, pairs map[string]string, ttl time.Duration) error

// Health check
func (c *RedisCache) Ping(ctx context.Context) error
func (c *RedisCache) Close() error
```

**agent/cache_redis_test.go** (~500 LOC)
- Unit tests with miniredis (in-memory mock)
- Integration tests with real Redis (Docker)
- Cluster mode tests
- Sentinel mode tests
- Performance benchmarks

#### 2.2 Key Design Decisions

**Key Naming Convention:**
```
{prefix}:cache:{hash}
Example: go-deep-agent:cache:a3f5b2c1d4e5f6...
```

**Value Format:**
```json
{
    "response": "The actual LLM response...",
    "model": "gpt-4o-mini",
    "timestamp": 1699564800,
    "metadata": {
        "temperature": 0.7,
        "tokens": 150
    }
}
```

**Statistics Tracking:**
- Use Redis INCR for atomic counters
- Separate keys for stats:
  - `{prefix}:stats:hits`
  - `{prefix}:stats:misses`
  - `{prefix}:stats:writes`

### 3. Advanced Features

#### 3.1 Cache Warming
```go
// Preload frequently used responses
func (b *Builder) WarmCache(ctx context.Context, prompts []string) error {
    for _, prompt := range prompts {
        _, _ = b.Ask(ctx, prompt)
    }
    return nil
}
```

#### 3.2 Cache Invalidation Patterns
```go
// Invalidate by pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
    // Use SCAN to find matching keys
    // Delete in batches
}

// Example: Clear all GPT-4 responses
cache.DeletePattern(ctx, "*:model:gpt-4:*")
```

#### 3.3 Distributed Locking (Optional)
```go
// Prevent cache stampede
func (c *RedisCache) GetOrCompute(ctx context.Context, key string, ttl time.Duration, 
    compute func() (string, error)) (string, error) {
    
    // Try to get from cache
    if val, found, _ := c.Get(ctx, key); found {
        return val, nil
    }
    
    // Acquire lock
    lockKey := key + ":lock"
    locked, _ := c.client.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
    
    if locked {
        defer c.client.Del(ctx, lockKey)
        
        // Compute value
        val, err := compute()
        if err != nil {
            return "", err
        }
        
        // Store in cache
        c.Set(ctx, key, val, ttl)
        return val, nil
    }
    
    // Wait for other process to compute
    time.Sleep(100 * time.Millisecond)
    return c.GetOrCompute(ctx, key, ttl, compute)
}
```

### 4. Testing Strategy

#### 4.1 Unit Tests (with miniredis)
```go
func TestRedisCacheBasic(t *testing.T) {
    s := miniredis.RunT(t)
    cache := NewRedisCache(s.Addr(), "", 0, 1*time.Minute)
    
    // Test Set/Get
    cache.Set(ctx, "key1", "value1", 1*time.Minute)
    val, found, _ := cache.Get(ctx, "key1")
    // assertions...
}
```

#### 4.2 Integration Tests (with Docker)
```go
func TestRedisCacheIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Start Redis container
    cache := NewRedisCache("localhost:6379", "", 0, 1*time.Minute)
    
    // Run comprehensive tests
}
```

#### 4.3 Performance Tests
```go
func BenchmarkRedisCacheGet(b *testing.B) {
    // Benchmark cache retrieval
}

func BenchmarkRedisCacheMGet(b *testing.B) {
    // Benchmark batch retrieval
}
```

### 5. Examples

**examples/cache_redis.go** (~300 LOC)
```go
func main() {
    // Example 1: Simple Redis cache
    simpleRedisCache()
    
    // Example 2: Redis with custom options
    redisWithOptions()
    
    // Example 3: Redis Cluster
    redisCluster()
    
    // Example 4: Cache warming
    cacheWarming()
    
    // Example 5: Pattern-based invalidation
    patternInvalidation()
}
```

**examples/cache_comparison.go** (~200 LOC)
- Compare MemoryCache vs RedisCache performance
- Show when to use each
- Demonstrate fallback strategies

### 6. Dependencies

```go
require (
    github.com/redis/go-redis/v9 v9.3.0
    github.com/alicebob/miniredis/v2 v2.31.0 // Testing
)
```

### 7. Documentation

#### 7.1 New Documentation Files
**docs/CACHING_GUIDE.md** - Comprehensive caching guide
- When to use MemoryCache vs RedisCache
- Configuration best practices
- Performance tuning
- Common patterns

**docs/REDIS_DEPLOYMENT.md** - Production Redis setup
- Single instance
- Redis Cluster
- Redis Sentinel
- Cloud providers (AWS ElastiCache, Redis Cloud)

#### 7.2 README Updates
```markdown
### Response Caching

#### In-Memory Cache (Single Process)
```go
builder.WithMemoryCache(1000, 5*time.Minute)
```

#### Redis Cache (Distributed, Persistent)
```go
builder.WithRedisCache("localhost:6379", "", 0)
```

#### Redis Cluster
```go
builder.WithRedisCacheOptions(&agent.RedisCacheOptions{
    Addrs: []string{"redis-1:6379", "redis-2:6379", "redis-3:6379"},
})
```
```

### 8. Migration Path

#### 8.1 Backward Compatibility
- Existing `WithMemoryCache()` continues to work
- `WithCache()` accepts any Cache implementation
- No breaking changes

#### 8.2 Switching Caches
```go
// Development: Use memory cache
if os.Getenv("ENV") == "development" {
    builder.WithMemoryCache(100, 5*time.Minute)
} else {
    // Production: Use Redis
    builder.WithRedisCache(os.Getenv("REDIS_URL"), "", 0)
}
```

#### 8.3 Gradual Rollout
```go
// Try Redis, fallback to memory
redisCache, err := agent.NewRedisCache("localhost:6379", "", 0, 5*time.Minute)
if err != nil {
    log.Printf("Redis unavailable, using memory cache: %v", err)
    builder.WithMemoryCache(100, 5*time.Minute)
} else {
    builder.WithCache(redisCache)
}
```

### 9. Implementation Timeline

#### Sprint 1 (Week 1): Foundation
- [ ] Design RedisCache struct and interfaces
- [ ] Implement basic Get/Set/Delete
- [ ] Add connection pooling
- [ ] Unit tests with miniredis

#### Sprint 2 (Week 2): Advanced Features
- [ ] Implement MGet/MSet (batch operations)
- [ ] Add pattern-based deletion
- [ ] Implement cache statistics in Redis
- [ ] Health check and Close methods

#### Sprint 3 (Week 3): Production Readiness
- [ ] Redis Cluster support
- [ ] Redis Sentinel support
- [ ] Integration tests with Docker
- [ ] Performance benchmarks
- [ ] Error handling and retries

#### Sprint 4 (Week 4): Documentation & Examples
- [ ] Create cache_redis.go example
- [ ] Write CACHING_GUIDE.md
- [ ] Write REDIS_DEPLOYMENT.md
- [ ] Update README.md
- [ ] Add cache comparison example

### 10. Success Metrics

#### Functional
- ✅ Full Cache interface implementation
- ✅ Support for standalone, cluster, and sentinel Redis
- ✅ 30+ tests, all passing
- ✅ Connection pooling and health checks
- ✅ 3+ working examples

#### Performance
- Get operation: <5ms (local Redis)
- Set operation: <10ms (local Redis)
- Batch operations: 100 items in <50ms
- Connection pool: Handle 100+ concurrent requests

#### Reliability
- Automatic reconnection on failure
- Graceful degradation (fallback to memory)
- Comprehensive error handling
- Circuit breaker for failed connections

### 11. Risk Mitigation

#### Technical Risks
1. **Redis unavailable** - Fallback to MemoryCache
2. **Network latency** - Connection pooling, local Redis
3. **Memory usage** - TTL enforcement, pattern-based cleanup
4. **Breaking changes** - Maintain backward compatibility

#### Operational Risks
1. **Redis crash** - Auto-reconnect, health checks
2. **Network partition** - Timeout configuration
3. **Redis eviction** - Monitor memory, adjust TTL
4. **Cost** - Use local Redis for dev, managed for prod

### 12. Future Enhancements (v0.6.0+)

#### Advanced Caching Strategies
- **Cache Aside** - Current implementation
- **Write Through** - Update cache on every write
- **Write Behind** - Async cache updates
- **Read Through** - Auto-populate on miss

#### Multi-Tier Caching
```go
builder.WithMultiTierCache(
    agent.NewMemoryCache(100, 1*time.Minute),   // L1: Fast, small
    agent.NewRedisCache("localhost:6379", "", 0), // L2: Slower, large
)
```

#### Cache Analytics
- Track hit/miss rates per model
- Monitor cache size trends
- Alert on low hit rates
- Cost analysis (API calls saved)

#### Smart Cache Invalidation
- Time-based (current implementation)
- Event-based (invalidate on new data)
- Version-based (invalidate on model update)
- Dependency-based (invalidate related keys)
