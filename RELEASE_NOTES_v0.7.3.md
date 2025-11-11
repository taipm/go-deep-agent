# Release Notes: v0.7.3 - Rate Limiting 🚦

**Release Date**: November 11, 2025

## Overview

Version 0.7.3 introduces comprehensive **Rate Limiting** support to help you stay within API limits, control costs, and implement fair usage policies in multi-tenant applications.

## 🎯 Key Features

### Token Bucket Algorithm
- Industry-standard token bucket implementation
- Configurable requests per second (sustained rate)
- Burst capacity for temporary spikes
- Automatic token refill

### Per-Key Rate Limiting
- Independent rate limits per key (user ID, API key, etc.)
- Automatic cleanup of unused limiters
- Configurable cleanup timeout

### Seamless Integration
- Works transparently with `Ask()` and `Stream()` methods
- Zero overhead when disabled (default)
- Fluent builder API

## 📦 Installation

```bash
go get github.com/taipm/go-deep-agent@v0.7.3
```

## 🚀 Quick Start

### Simple Rate Limiting

```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRateLimit(10.0, 20) // 10 req/s, burst of 20

// First 20 requests use burst (immediate)
// Remaining requests throttled to 10/second
for i := 0; i < 100; i++ {
    ai.Ask(ctx, fmt.Sprintf("Question %d", i))
}
```

### Per-User Rate Limiting

```go
config := agent.RateLimitConfig{
    Enabled:           true,
    RequestsPerSecond: 5.0,
    BurstSize:         10,
    PerKey:            true, // Enable per-key limits
}

// Each user gets independent quota
aiUser1 := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRateLimitConfig(config).
    WithRateLimitKey("user-123")

aiUser2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRateLimitConfig(config).
    WithRateLimitKey("user-456")
```

### Advanced Configuration

```go
config := agent.RateLimitConfig{
    Enabled:           true,
    RequestsPerSecond: 10.0,       // Sustained rate
    BurstSize:         20,          // Burst capacity
    PerKey:            false,       // Global or per-key
    KeyTimeout:        5 * time.Minute,  // Cleanup unused keys
    WaitTimeout:       30 * time.Second, // Max wait per request
}
```

## 📊 What's New

### Core Implementation
- `RateLimiter` interface with `Allow`, `Wait`, `Reserve`, `Stats` methods
- Token bucket implementation using `golang.org/x/time/rate`
- Thread-safe statistics tracking
- Context-aware waiting with timeouts

### Builder Methods
- `WithRateLimit(requestsPerSecond, burstSize)` - Simple configuration
- `WithRateLimitConfig(config)` - Advanced configuration  
- `WithRateLimitKey(key)` - Set key for per-key limiting

### AgentConfig Integration
- Added `RateLimitConfig` field
- YAML/JSON serialization support
- Configuration validation

### Statistics
- Track allowed, denied, and waited requests
- Total wait time monitoring
- Available tokens and active keys count

## 🧪 Testing

**21 Test Cases** - 100% passing:
- 8 unit tests (core functionality)
- 12 integration tests (builder, concurrency)
- 1 configuration test

**Coverage**: 73.7% of statements in agent package

## 📚 Documentation

### Examples
- `examples/rate_limit_basic` - Simple usage, burst capacity
- `examples/rate_limit_advanced` - Per-key limits, concurrent requests

### Guides
- [Rate Limiting Guide](../docs/RATE_LIMITING_GUIDE.md) - 957 lines, comprehensive
- [Quick Reference](../RATE_LIMITING_QUICKSTART.md) - Visual overview
- [Implementation TODO](../TODO_RATE_LIMITING.md) - Development roadmap

## 🔄 Migration Guide

### From v0.7.2 or earlier

**No breaking changes** - Rate limiting is disabled by default:

```go
// Existing code works unchanged
ai := agent.NewOpenAI("gpt-4o-mini", apiKey)
ai.Ask(ctx, "Hello") // No rate limiting

// Enable when needed
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithRateLimit(10.0, 20) // Now rate limited
```

### Enabling Rate Limiting

Choose your configuration level:

**Level 1: Simple** (for most use cases)
```go
ai.WithRateLimit(10.0, 20) // 10 req/s, burst 20
```

**Level 2: Advanced** (for custom behavior)
```go
config := agent.RateLimitConfig{
    Enabled:           true,
    RequestsPerSecond: 10.0,
    BurstSize:         20,
    WaitTimeout:       30 * time.Second,
}
ai.WithRateLimitConfig(config)
```

**Level 3: Per-Key** (for multi-tenant apps)
```go
config := agent.RateLimitConfig{
    Enabled:           true,
    RequestsPerSecond: 5.0,
    BurstSize:         10,
    PerKey:            true, // Enable per-key
}
ai.WithRateLimitConfig(config).
   WithRateLimitKey(userID)
```

## 📋 Configuration Reference

```go
type RateLimitConfig struct {
    Enabled           bool          // Enable/disable rate limiting
    RequestsPerSecond float64       // Sustained request rate (> 0)
    BurstSize         int           // Maximum burst requests (>= 1)
    PerKey            bool          // Enable per-key rate limiting
    KeyTimeout        time.Duration // Cleanup timeout for unused keys
    WaitTimeout       time.Duration // Max wait time per request
}
```

### Default Values
- `Enabled`: `false` (backward compatible)
- `RequestsPerSecond`: `10.0`
- `BurstSize`: `20`
- `KeyTimeout`: `5 * time.Minute`
- `WaitTimeout`: `30 * time.Second`

## 🎯 Use Cases

### 1. API Compliance
Stay within OpenAI/provider rate limits:
```go
// OpenAI: 10,000 req/min = ~166 req/s
ai.WithRateLimit(160.0, 200) // Stay safely under limit
```

### 2. Cost Control
Prevent accidental quota exhaustion:
```go
// Limit to 100 requests per second
ai.WithRateLimit(100.0, 150)
```

### 3. Multi-Tenant Fair Usage
Implement per-user quotas:
```go
config := agent.RateLimitConfig{
    RequestsPerSecond: 5.0,  // 5 req/s per user
    BurstSize:         10,    // Allow small bursts
    PerKey:            true,  // Independent per user
}
```

### 4. Burst Handling
Allow spikes while maintaining average rate:
```go
// Sustained: 10 req/s, Burst: 50
ai.WithRateLimit(10.0, 50)
// First 50 requests immediate
// Then throttled to 10/second
```

## 🔧 Dependencies

### New
- `golang.org/x/time v0.14.0` - Token bucket implementation

### Updated
- No dependency updates required

## ⚡ Performance

- **Zero overhead** when rate limiting is disabled
- **Minimal overhead** when enabled (~10μs per request)
- **Thread-safe** - supports concurrent requests
- **Efficient cleanup** - automatic removal of unused per-key limiters

## 🐛 Known Limitations

1. **In-Memory Only**: Current implementation uses in-memory storage
   - Not suitable for distributed systems (multiple instances)
   - Redis backend planned for future release (optional)

2. **Per-Instance Limits**: Each application instance has independent limits
   - For distributed rate limiting, wait for Redis backend

3. **No Persistent State**: Limits reset on application restart
   - Acceptable for most use cases
   - Persistent state planned for Redis backend

## 🔮 Future Enhancements

Planned for future releases:
- Redis backend for distributed rate limiting
- Persistent state across restarts
- Rate limit statistics dashboard
- Adaptive rate limiting based on API response headers

## 📈 Statistics

### Code Metrics
- **New Files**: 3 (`rate_limiter.go`, `rate_limiter_token_bucket.go`, `rate_limiter_test.go`)
- **Modified Files**: 4 (`agent_config.go`, `builder.go`, `builder_config.go`, `builder_execution.go`)
- **Lines Added**: ~1,900
- **Tests Added**: 21 (all passing)

### Coverage
- Overall: 73.7% (up from 73.6%)
- Rate limiter: 100% (all code paths tested)

## 🙏 Acknowledgments

- Token bucket algorithm based on `golang.org/x/time/rate`
- Design inspired by industry best practices
- Community feedback from production users

## 📞 Support

- **Documentation**: [Rate Limiting Guide](../docs/RATE_LIMITING_GUIDE.md)
- **Examples**: `examples/rate_limit_basic`, `examples/rate_limit_advanced`
- **Issues**: [GitHub Issues](https://github.com/taipm/go-deep-agent/issues)
- **Discussions**: [GitHub Discussions](https://github.com/taipm/go-deep-agent/discussions)

## ✅ Verification

Verify the release:

```bash
# Check version
go list -m github.com/taipm/go-deep-agent@v0.7.3
# Output: github.com/taipm/go-deep-agent v0.7.3

# Run tests
go test -v github.com/taipm/go-deep-agent/agent -run TestRateLimit
# All 21 tests should pass

# Try it out
go get github.com/taipm/go-deep-agent@v0.7.3
```

---

**Full Changelog**: [v0.7.2...v0.7.3](https://github.com/taipm/go-deep-agent/compare/v0.7.2...v0.7.3)
