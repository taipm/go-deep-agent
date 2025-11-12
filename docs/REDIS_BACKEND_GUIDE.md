# Redis Backend for Long-Term Memory

Complete guide to using Redis for persistent conversation storage in go-deep-agent.

---

## 🚀 Quick Start (Recommended for 90% of users)

**Get started in 3 lines of code:**

```go
package main

import (
    "context"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    ctx := context.Background()
    
    // ✅ RECOMMENDED: Simple setup with defaults
    backend := agent.NewRedisBackend("localhost:6379")
    defer backend.Close()
    
    // Create agent with persistent memory
    ai := agent.NewOpenAI("gpt-4", apiKey).
        WithShortMemory().
        WithLongMemory("user-alice").
        UsingBackend(backend)
    
    // Conversations automatically saved to Redis!
    ai.Ask(ctx, "My favorite color is blue")
    
    // Restart your app - memory persists automatically!
}
```

---

## Installation

### 1. Install Redis

**macOS:**
```bash
brew install redis
brew services start redis
```

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install redis-server
sudo systemctl start redis-server
```

**Docker:**
```bash
docker run -d --name redis -p 6379:6379 redis:latest
```

**Verify Installation:**
```bash
redis-cli ping
# Should return: PONG
```

### 2. Add Go Dependencies

Redis client is already included in go-deep-agent:
```go
import "github.com/taipm/go-deep-agent/agent"
```

---

## Basic Usage

## 📋 Table of Contents

- [Quick Start](#-quick-start-recommended-for-90-of-users) ← **Start here**
- [Installation](#installation)
- [Common Use Cases](#common-use-cases)
- [Advanced Configuration](#advanced-configuration) (Cluster/Sentinel/Custom)
- [Production Best Practices](#production-best-practices)
- [API Reference](#api-reference)
- [Troubleshooting](#troubleshooting)

---

## Common Use Cases

### ✅ Case 1: Simple Setup (No password)

**This works for 80% of users:**

```go
backend := agent.NewRedisBackend("localhost:6379")

ai := agent.NewOpenAI("gpt-4", apiKey).
    WithShortMemory().
    WithLongMemory("user-123").
    UsingBackend(backend)
```

**Smart defaults applied automatically:**
- ✅ Password: none
- ✅ DB: 0  
- ✅ TTL: 7 days (conversations auto-expire after 7 days of inactivity)
- ✅ Prefix: `go-deep-agent:memories:`
- ✅ Pool: 10 connections

### ✅ Case 2: Production Setup (With password)

**Most common production setup:**

```go
backend := agent.NewRedisBackend("localhost:6379").
    WithPassword("your-redis-password")

ai := agent.NewOpenAI("gpt-4", apiKey).
    WithShortMemory().
    WithLongMemory("user-alice").
    UsingBackend(backend)
```

### ✅ Case 3: Custom TTL (e.g., 24 hours)

**For short-lived conversations:**

```go
backend := agent.NewRedisBackend("localhost:6379").
    WithTTL(24 * time.Hour)  // Expire after 1 day

ai := agent.NewOpenAI("gpt-4", apiKey).
    WithLongMemory("session-xyz").
    UsingBackend(backend)
```

### ✅ Case 4: Multiple Options

**Combine multiple customizations:**

```go
backend := agent.NewRedisBackend("localhost:6379").
    WithPassword("secret").
    WithDB(2).
    WithTTL(24 * time.Hour).
    WithPrefix("myapp:")
```

### Memory Operations

**Auto-Save (Default):**
```go
ai := agent.NewOpenAI("gpt-4", apiKey).
    WithShortMemory().
    WithLongMemory("user-alice").
    UsingBackend(backend)
    // Auto-save enabled by default

ai.Ask(ctx, "Hello")  // Automatically saved to Redis
```

**Manual Save:**
```go
ai := agent.NewOpenAI("gpt-4", apiKey).
    WithShortMemory().
    WithLongMemory("user-bob").
    UsingBackend(backend).
    WithAutoSaveLongMemory(false)  // Disable auto-save

ai.Ask(ctx, "Message 1")
ai.Ask(ctx, "Message 2")
ai.SaveLongMemory(ctx)  // Explicit save
```

**Load Memory:**
```go
// Auto-load on initialization
ai := agent.NewOpenAI("gpt-4", apiKey).
    WithShortMemory().
    WithLongMemory("user-alice").  // Automatically loads existing memory
    UsingBackend(backend)

// Or manual load
ai.LoadLongMemory(ctx)
```

**Delete Memory:**
```go
err := ai.DeleteLongMemory(ctx)
```

**List All Memories:**
```go
memoryIDs, err := ai.ListLongMemories(ctx)
for _, id := range memoryIDs {
    fmt.Println(id)
}
```

---

## Advanced Configuration

<details>
<summary><b>💡 When do I need advanced configuration?</b></summary>

Most users (90%) can skip this section. You only need advanced configuration if:
- ✅ Using Redis Cluster or Sentinel (high availability)
- ✅ Need more than 10 connection pool size
- ✅ Want to share Redis client with other parts of your app
- ✅ Have specific performance tuning requirements

Otherwise, stick with the simple setup above!

</details>

---

### Option A: Fluent API (Recommended)

**Use when**: Customizing 1-3 options

```go
backend := agent.NewRedisBackend("localhost:6379").
    WithPassword("secret").
    WithTTL(24 * time.Hour)
```

### Option B: Options Struct

**Use when**: Customizing 4+ options or prefer struct initialization

```go
opts := &agent.RedisBackendOptions{
    Addr:     "localhost:6379",
    Password: "secret",
    DB:       2,
    TTL:      24 * time.Hour,
    Prefix:   "myapp:conversations:",
    PoolSize: 50,
}

backend := agent.NewRedisBackendWithOptions(opts)
```

**Note**: Both ways work identically. Choose based on preference.

---

### Expert Mode: Redis Cluster/Sentinel

**Use when**: Production deployments with high availability requirements

**Redis Cluster:**
```go
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "node1:6379",
        "node2:6379",
        "node3:6379",
    },
    Password:     "secret",
    PoolSize:     100,
    MinIdleConns: 20,
})

backend := agent.NewRedisBackendWithClient(clusterClient).
    WithTTL(7 * 24 * time.Hour)  // Can still use fluent API
```

**Redis Sentinel:**
```go
sentinelClient := redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{"sentinel1:26379", "sentinel2:26379"},
    Password:      "secret",
})

backend := agent.NewRedisBackendWithClient(sentinelClient)
```

### Configuration Options Reference

| Option | Type | Default | When to Change? |
|--------|------|---------|-----------------|
| `Addr` | string | **required** | Always specify (e.g., `"localhost:6379"`) |
| `Password` | string | `""` (no auth) | Change if Redis requires authentication |
| `DB` | int | `0` | Change to isolate data (0-15 available) |
| `TTL` | Duration | `7 days` | Change for different expiration needs:<br>• 1 hour = anonymous sessions<br>• 7 days = default<br>• 30 days = premium users |
| `Prefix` | string | `"go-deep-agent:memories:"` | Change to avoid key collisions if sharing Redis |
| `PoolSize` | int | `10` | Increase if handling >100 concurrent requests |

**💡 Tip**: For most users, only `Addr` and `Password` need to be set. Other options have sensible defaults.

---

## Production Best Practices

### 1. Connection Management

**Always close connections:**
```go
backend := agent.NewRedisBackend("localhost:6379")
defer backend.Close()
```

**Health checks:**
```go
if err := backend.Ping(ctx); err != nil {
    log.Fatalf("Redis unavailable: %v", err)
}
```

### 2. TTL Strategy

**Choose appropriate TTL based on use case:**

```go
// Anonymous sessions - 1 hour
backend := agent.NewRedisBackend("localhost:6379").
    WithTTL(1 * time.Hour)

// Regular users - 7 days (default)
backend := agent.NewRedisBackend("localhost:6379")
    // Default: 7 * 24 * time.Hour

// Premium users - 30 days
backend := agent.NewRedisBackend("localhost:6379").
    WithTTL(30 * 24 * time.Hour)
```

**TTL extends on every save:**
Active conversations never expire as long as they're being used.

### 3. Key Namespacing

**Use prefixes to avoid collisions:**

```go
// Development
devBackend := agent.NewRedisBackend("localhost:6379").
    WithDB(1).
    WithPrefix("dev:conversations:")

// Production
prodBackend := agent.NewRedisBackend("prod-redis:6379").
    WithDB(0).
    WithPrefix("prod:conversations:")

// Multi-tenant
tenant1Backend := agent.NewRedisBackend("localhost:6379").
    WithPrefix("tenant-abc:conversations:")
```

### 4. Performance Optimization

**Connection pooling for high traffic:**

```go
opts := &agent.RedisBackendOptions{
    Addr:     "localhost:6379",
    PoolSize: 100,        // Increase for high concurrency
    // Default: 10
}

backend := agent.NewRedisBackendWithOptions(opts)
```

**Use Redis Cluster for scale:**

```go
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{"node1:6379", "node2:6379", "node3:6379"},
    // Automatic sharding and failover
})

backend := agent.NewRedisBackendWithClient(clusterClient)
```

### 5. Error Handling

**Graceful degradation:**

```go
backend := agent.NewRedisBackend("localhost:6379")

// Check connection before use
if err := backend.Ping(ctx); err != nil {
    log.Println("Redis unavailable, falling back to in-memory")
    // Use FileBackend or in-memory only
} else {
    ai := agent.NewOpenAI("gpt-4", apiKey).
        WithShortMemory().
        WithLongMemory("user-123").
        UsingBackend(backend)
}
```

### 6. Security

**Use password authentication:**

```go
backend := agent.NewRedisBackend("prod-redis:6379").
    WithPassword(os.Getenv("REDIS_PASSWORD"))
```

**Use TLS for production:**

```go
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
}

client := redis.NewClient(&redis.Options{
    Addr:      "secure-redis:6379",
    Password:  os.Getenv("REDIS_PASSWORD"),
    TLSConfig: tlsConfig,
})

backend := agent.NewRedisBackendWithClient(client)
```

---

## API Reference

### Constructor Functions

#### `NewRedisBackend(addr string) *RedisBackend`

Create backend with defaults.

```go
backend := agent.NewRedisBackend("localhost:6379")
```

#### `NewRedisBackendWithOptions(opts *RedisBackendOptions) *RedisBackend`

Create backend from options struct.

```go
opts := &agent.RedisBackendOptions{
    Addr: "localhost:6379",
    TTL:  24 * time.Hour,
}
backend := agent.NewRedisBackendWithOptions(opts)
```

#### `NewRedisBackendWithClient(client redis.UniversalClient) *RedisBackend`

Create backend with custom Redis client.

```go
client := redis.NewClusterClient(...)
backend := agent.NewRedisBackendWithClient(client)
```

### Fluent API Methods

#### `WithPassword(password string) *RedisBackend`

Set authentication password.

```go
backend.WithPassword("secret")
```

#### `WithDB(db int) *RedisBackend`

Set Redis database number (0-15).

```go
backend.WithDB(2)
```

#### `WithTTL(ttl time.Duration) *RedisBackend`

Set memory expiration time.

```go
backend.WithTTL(24 * time.Hour)
```

#### `WithPrefix(prefix string) *RedisBackend`

Set key prefix for namespacing.

```go
backend.WithPrefix("myapp:conversations:")
```

### Memory Operations

#### `Load(ctx, memoryID) ([]Message, error)`

Load conversation history.

```go
messages, err := backend.Load(ctx, "user-alice")
```

#### `Save(ctx, memoryID, messages) error`

Save conversation history.

```go
err := backend.Save(ctx, "user-alice", messages)
```

#### `Delete(ctx, memoryID) error`

Delete conversation history.

```go
err := backend.Delete(ctx, "user-alice")
```

#### `List(ctx) ([]string, error)`

List all memory IDs.

```go
ids, err := backend.List(ctx)
```

### Utility Methods

#### `Ping(ctx) error`

Check Redis connection health.

```go
if err := backend.Ping(ctx); err != nil {
    log.Fatal("Redis unavailable")
}
```

#### `Close() error`

Close Redis connection.

```go
defer backend.Close()
```

---

## Troubleshooting

### Connection Refused

**Error:**
```
failed to connect to Redis: dial tcp 127.0.0.1:6379: connect: connection refused
```

**Solutions:**
1. Start Redis: `redis-server` or `brew services start redis`
2. Check Redis is running: `redis-cli ping`
3. Verify address and port

### Authentication Failed

**Error:**
```
NOAUTH Authentication required
```

**Solution:**
```go
backend := agent.NewRedisBackend("localhost:6379").
    WithPassword("your-password")
```

### Key Not Found

**Behavior:** `Load()` returns `nil, nil` (not an error)

**Explanation:** First-time memory load - this is expected behavior.

### Memory Not Persisting

**Check auto-save is enabled:**
```go
ai := agent.NewOpenAI("gpt-4", apiKey).
    WithShortMemory().
    WithLongMemory("user-123").
    UsingBackend(backend)
    // Auto-save enabled by default (good!)
```

**If disabled, save manually:**
```go
ai.WithAutoSaveLongMemory(false)
// ... conversations ...
ai.SaveLongMemory(ctx)  // Don't forget this!
```

### TTL Expired

**Check key expiration in Redis:**
```bash
redis-cli TTL "go-deep-agent:memories:user-alice"
# -2 = expired, -1 = no expiration, >0 = seconds remaining
```

**Extend TTL:**
```go
backend := agent.NewRedisBackend("localhost:6379").
    WithTTL(30 * 24 * time.Hour)  // 30 days
```

### Performance Issues

**Increase connection pool:**
```go
opts := &agent.RedisBackendOptions{
    Addr:     "localhost:6379",
    PoolSize: 100,  // Default: 10
}
backend := agent.NewRedisBackendWithOptions(opts)
```

**Use Redis Cluster for scale.**

---

## Examples

See complete examples:
- [Basic Redis Usage](../examples/redis_long_memory_basic.go)
- [Advanced Configuration](../examples/redis_long_memory_advanced.go)

## Related Documentation

- [Memory System Overview](MEMORY_SYSTEM.md)
- [FileBackend Guide](FILE_BACKEND.md)
- [Migration Guide v0.9](MIGRATION_v0.9.md)

---

**Questions?** Open an issue on GitHub or check our [FAQ](FAQ.md).
