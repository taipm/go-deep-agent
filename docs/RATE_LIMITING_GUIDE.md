# Rate Limiting Guide - go-deep-agent

**Status:** 🚧 Not Implemented (Planned for v0.8.0)  
**Priority:** P1 (High)  
**Complexity:** Medium  
**Estimated Time:** 1-2 weeks

---

## 📚 Mục lục

1. [Rate Limiting là gì?](#rate-limiting-là-gì)
2. [Tại sao cần Rate Limiting?](#tại-sao-cần-rate-limiting)
3. [Các loại Rate Limiting](#các-loại-rate-limiting)
4. [Thiết kế đề xuất](#thiết-kế-đề-xuất)
5. [Ví dụ triển khai](#ví-dụ-triển-khai)
6. [Best Practices](#best-practices)

---

## Rate Limiting là gì?

**Rate Limiting** là kỹ thuật giới hạn số lượng requests mà một client có thể thực hiện trong một khoảng thời gian nhất định.

### Ví dụ đơn giản:

```
Giới hạn: 100 requests/phút
- Request 1-100: ✅ Cho phép
- Request 101:   ❌ Từ chối với lỗi "Rate limit exceeded"
- Sau 1 phút:    ✅ Counter reset về 0, lại cho phép 100 requests mới
```

### Thuật ngữ quan trọng:

- **Rate**: Tốc độ cho phép (ví dụ: 100 requests/minute)
- **Burst**: Số lượng requests tối đa trong thời điểm ngắn (ví dụ: 10 requests/second)
- **Window**: Khoảng thời gian tính toán (sliding window, fixed window, token bucket)
- **Quota**: Tổng số lượng tài nguyên được phép sử dụng (ví dụ: 1M tokens/month)

---

## Tại sao cần Rate Limiting?

### 🛡️ 1. Bảo vệ khỏi lạm dụng (Abuse Protection)

**Vấn đề:**
```go
// Kẻ tấn công có thể gửi vô số requests
for i := 0; i < 1000000; i++ {
    ai.Ask(ctx, "Spam request")  // Không bị giới hạn!
}
```

**Hậu quả:**
- Chi phí API tăng vọt (OpenAI tính phí theo token)
- Server quá tải
- Service bị chậm cho users khác
- Có thể bị OpenAI ban account

**Giải pháp với Rate Limiting:**
```go
// Với rate limiting: Chỉ cho phép 100 req/phút
for i := 0; i < 1000000; i++ {
    err := ai.Ask(ctx, "Request")
    if err == agent.ErrRateLimitExceeded {
        // Request 101+ bị từ chối
        time.Sleep(1 * time.Minute)  // Phải đợi
    }
}
```

### 💰 2. Kiểm soát chi phí (Cost Control)

**Tình huống thực tế:**

```
OpenAI Pricing (GPT-4):
- Input:  $0.03 / 1K tokens
- Output: $0.06 / 1K tokens

Không có rate limiting:
- User A gửi 10,000 requests (bug trong code)
- Mỗi request: 500 tokens input + 500 tokens output = 1,000 tokens
- Tổng: 10,000 * 1,000 = 10M tokens
- Chi phí: 10M / 1000 * $0.045 = $450 trong 1 giờ!
```

**Với rate limiting:**
```
Giới hạn: 1,000 requests/hour per user
- User A bị block sau 1,000 requests
- Chi phí tối đa: $45/hour thay vì $450/hour
- Tiết kiệm: 90%!
```

### ⚖️ 3. Phân bổ tài nguyên công bằng (Fair Usage)

**Kịch bản Multi-Tenant SaaS:**

```
Hệ thống có 100 users:
- Không rate limiting: User A gửi 90% traffic → 99 users khác bị chậm
- Có rate limiting: Mỗi user tối đa 100 req/min → Fair cho tất cả
```

### 🔒 4. Tuân thủ API Provider Limits

**OpenAI API Limits:**
```
Tier 1 (Free):
- 3 requests/minute
- 200 requests/day
- 40,000 tokens/day

Tier 5 (Enterprise):
- 10,000 requests/minute
- 2,000,000 tokens/minute
```

**Nếu không có rate limiting:**
```go
// Code này sẽ bị OpenAI reject
for i := 0; i < 100; i++ {
    ai.Ask(ctx, "Question")  // Request 4+ → 429 Error
}
```

### 🚨 5. Phòng chống DoS/DDoS

**Denial of Service Attack:**
```
Attacker gửi 1 triệu requests/giây
→ Hệ thống quá tải
→ Service down cho users hợp lệ
→ Doanh thu bị mất
```

**Với rate limiting:**
```
Mỗi IP chỉ cho phép 100 req/min
→ Attacker bị block sau 100 requests
→ Service vẫn hoạt động bình thường
```

---

## Các loại Rate Limiting

### 1. **Fixed Window** (Cửa sổ cố định)

**Cách hoạt động:**
```
Window: 1 phút
Limit: 100 requests

Minute 1 (00:00-00:59):
- Request 1-100:  ✅ Allowed
- Request 101+:   ❌ Blocked

Minute 2 (01:00-01:59):
- Counter reset về 0
- Request 1-100:  ✅ Allowed again
```

**Ưu điểm:**
- ✅ Đơn giản, dễ implement
- ✅ Hiệu năng cao (chỉ cần 1 counter)
- ✅ Dễ hiểu với users

**Nhược điểm:**
- ❌ Có thể bị burst at window boundary
  ```
  00:59 → 50 requests
  01:00 → 50 requests
  = 100 requests trong 1 giây!
  ```

**Use case:**
- API đơn giản
- Không quan trọng việc burst ngắn hạn

### 2. **Sliding Window** (Cửa sổ trượt)

**Cách hoạt động:**
```
Limit: 100 requests/phút
Current time: 12:00:30

Count requests trong 60 giây trước:
- Từ 11:59:30 đến 12:00:30
- Nếu < 100: Allow
- Nếu >= 100: Block

12:00:31 → Window trượt: 11:59:31 - 12:00:31
12:00:32 → Window trượt: 11:59:32 - 12:00:32
```

**Ưu điểm:**
- ✅ Chính xác hơn Fixed Window
- ✅ Ngăn chặn burst at boundary
- ✅ Phân phối đều traffic

**Nhược điểm:**
- ❌ Phức tạp hơn (cần lưu timestamp mỗi request)
- ❌ Tốn memory (lưu lịch sử requests)
- ❌ Hiệu năng thấp hơn Fixed Window

**Use case:**
- API quan trọng cần chính xác
- Multi-tenant SaaS

### 3. **Token Bucket** (Thùng token) - ⭐ Khuyên dùng

**Cách hoạt động:**

```
Bucket capacity: 100 tokens
Refill rate: 10 tokens/giây

Trạng thái ban đầu: 100 tokens

Request 1: Consume 1 token → 99 tokens còn lại ✅
Request 2: Consume 1 token → 98 tokens ✅
...
Request 101: No tokens left → ❌ BLOCKED

Sau 1 giây: +10 tokens → 10 tokens
Request 102-111: Consume 10 tokens → ✅ Allowed
Request 112: No tokens → ❌ BLOCKED
```

**Công thức:**
```go
// Cập nhật số tokens
tokensToAdd = (currentTime - lastRefill) * refillRate
currentTokens = min(currentTokens + tokensToAdd, bucketCapacity)

// Kiểm tra request
if currentTokens >= requestCost {
    currentTokens -= requestCost
    return ALLOW
} else {
    return BLOCK
}
```

**Ưu điểm:**
- ✅ Cho phép burst ngắn hạn (khi bucket đầy)
- ✅ Smooth traffic over time
- ✅ Hiệu năng tốt (chỉ cần 2 variables: tokens, lastRefill)
- ✅ Linh hoạt (có thể set burst capacity)

**Nhược điểm:**
- ❌ Hơi phức tạp để hiểu
- ❌ Cần tính toán refill

**Use case:** ⭐ **Khuyên dùng cho go-deep-agent**
- Cân bằng giữa hiệu năng và chính xác
- Cho phép burst hợp lý
- Phù hợp với LLM API (có peak traffic)

### 4. **Leaky Bucket** (Thùng dò)

**Cách hoạt động:**

```
Queue capacity: 100 requests
Processing rate: 10 requests/giây

Request đến → Vào queue
Queue → Process với tốc độ cố định

Queue: [R1, R2, R3, ...]
       ↓ 10 req/s
     Process
```

**Ưu điểm:**
- ✅ Traffic smoothing (output rate đều)
- ✅ Bảo vệ downstream services

**Nhược điểm:**
- ❌ Latency cao (phải queue)
- ❌ Không phù hợp real-time

**Use case:**
- Message queue systems
- Batch processing

---

## Thiết kế đề xuất cho go-deep-agent

### Architecture Overview

```
User Request
    ↓
┌─────────────────────┐
│  Rate Limiter       │
│  (Token Bucket)     │
└─────────────────────┘
    ↓ (if allowed)
┌─────────────────────┐
│  Agent.Ask()        │
│  Agent.Stream()     │
└─────────────────────┘
    ↓
┌─────────────────────┐
│  OpenAI API         │
└─────────────────────┘
```

### API Design

```go
package agent

// RateLimiter interface
type RateLimiter interface {
    // Allow checks if request is allowed
    Allow(ctx context.Context, key string) (bool, error)
    
    // Wait waits until request is allowed (blocking)
    Wait(ctx context.Context, key string) error
    
    // Reserve reserves permission for future use
    Reserve(ctx context.Context, key string) (*Reservation, error)
    
    // Stats returns current rate limit statistics
    Stats(ctx context.Context, key string) (*RateLimitStats, error)
}

// RateLimitConfig configuration
type RateLimitConfig struct {
    // Strategy: "token-bucket", "fixed-window", "sliding-window"
    Strategy string
    
    // Rate: requests per second
    Rate int
    
    // Burst: maximum burst size (for token bucket)
    Burst int
    
    // KeyFunc: function to extract rate limit key from context
    // Default: per IP, can be per user, per API key, etc.
    KeyFunc func(ctx context.Context) string
    
    // OnRateLimitExceeded: callback when limit exceeded
    OnRateLimitExceeded func(ctx context.Context, key string)
    
    // Storage: where to store counters (memory, redis)
    Storage RateLimitStorage
}

// RateLimitStats statistics
type RateLimitStats struct {
    Key           string
    Limit         int
    Remaining     int
    ResetAt       time.Time
    RetryAfter    time.Duration
}

// Builder methods
func (b *Builder) WithRateLimit(config *RateLimitConfig) *Builder
func (b *Builder) WithRateLimitPerSecond(rate, burst int) *Builder
func (b *Builder) WithRateLimitPerMinute(rate, burst int) *Builder
func (b *Builder) WithRateLimitPerHour(rate, burst int) *Builder
```

### Implementation với Token Bucket

```go
package agent

import (
    "context"
    "sync"
    "time"
    "golang.org/x/time/rate"
)

// TokenBucketLimiter implementation
type TokenBucketLimiter struct {
    limiters sync.Map // map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
    mu       sync.RWMutex
}

func NewTokenBucketLimiter(rps int, burst int) *TokenBucketLimiter {
    return &TokenBucketLimiter{
        rate:  rate.Limit(rps),
        burst: burst,
    }
}

func (l *TokenBucketLimiter) getLimiter(key string) *rate.Limiter {
    limiter, exists := l.limiters.Load(key)
    if !exists {
        limiter = rate.NewLimiter(l.rate, l.burst)
        l.limiters.Store(key, limiter)
    }
    return limiter.(*rate.Limiter)
}

func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
    limiter := l.getLimiter(key)
    return limiter.Allow(), nil
}

func (l *TokenBucketLimiter) Wait(ctx context.Context, key string) error {
    limiter := l.getLimiter(key)
    return limiter.Wait(ctx)
}

func (l *TokenBucketLimiter) Reserve(ctx context.Context, key string) (*Reservation, error) {
    limiter := l.getLimiter(key)
    r := limiter.Reserve()
    
    if !r.OK() {
        return nil, ErrRateLimitExceeded
    }
    
    return &Reservation{
        reservation: r,
        delay:       r.Delay(),
    }, nil
}

func (l *TokenBucketLimiter) Stats(ctx context.Context, key string) (*RateLimitStats, error) {
    limiter := l.getLimiter(key)
    
    // Calculate remaining tokens
    r := limiter.Reserve()
    defer r.Cancel()
    
    remaining := l.burst
    if !r.OK() {
        remaining = 0
    }
    
    return &RateLimitStats{
        Key:        key,
        Limit:      l.burst,
        Remaining:  remaining,
        ResetAt:    time.Now().Add(r.Delay()),
        RetryAfter: r.Delay(),
    }, nil
}
```

### Redis-backed Rate Limiter (Distributed)

```go
package agent

import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

// RedisRateLimiter for distributed rate limiting
type RedisRateLimiter struct {
    client *redis.Client
    rate   int
    window time.Duration
}

func NewRedisRateLimiter(client *redis.Client, rate int, window time.Duration) *RedisRateLimiter {
    return &RedisRateLimiter{
        client: client,
        rate:   rate,
        window: window,
    }
}

func (l *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
    // Sliding window with Redis
    now := time.Now()
    windowStart := now.Add(-l.window)
    
    pipe := l.client.Pipeline()
    
    // Remove old entries
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))
    
    // Count entries in window
    countCmd := pipe.ZCard(ctx, key)
    
    // Add current request
    pipe.ZAdd(ctx, key, redis.Z{
        Score:  float64(now.UnixNano()),
        Member: now.UnixNano(),
    })
    
    // Set expiry
    pipe.Expire(ctx, key, l.window)
    
    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }
    
    count := countCmd.Val()
    return count < int64(l.rate), nil
}
```

---

## Ví dụ triển khai

### Example 1: Basic Rate Limiting

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    ctx := context.Background()
    
    // Create agent with rate limiting
    ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithRateLimitPerMinute(100, 10).  // 100 req/min, burst 10
        WithAutoExecute(true)
    
    // Normal usage - first 10 requests are fast (burst)
    for i := 0; i < 15; i++ {
        start := time.Now()
        
        resp, err := ai.Ask(ctx, "Hello")
        if err != nil {
            if err == agent.ErrRateLimitExceeded {
                fmt.Printf("Request %d: Rate limited!\n", i+1)
                continue
            }
            panic(err)
        }
        
        fmt.Printf("Request %d: %s (took %v)\n", i+1, resp, time.Since(start))
    }
    
    // Output:
    // Request 1: Hello! (took 500ms)   - From burst
    // Request 2: Hello! (took 501ms)   - From burst
    // ...
    // Request 10: Hello! (took 502ms)  - From burst
    // Request 11: Rate limited!        - Burst exhausted, waiting for refill
    // Request 12: Hello! (took 1.5s)   - After refill
}
```

### Example 2: Per-User Rate Limiting

```go
package main

import (
    "context"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Rate limit by user ID
    config := &agent.RateLimitConfig{
        Strategy: "token-bucket",
        Rate:     100,  // 100 req/min
        Burst:    10,
        KeyFunc: func(ctx context.Context) string {
            // Extract user ID from context
            userID := ctx.Value("user_id").(string)
            return fmt.Sprintf("user:%s", userID)
        },
        OnRateLimitExceeded: func(ctx context.Context, key string) {
            log.Printf("Rate limit exceeded for %s", key)
            // Send alert, update metrics, etc.
        },
    }
    
    ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithRateLimit(config)
    
    // Different users have separate limits
    ctx1 := context.WithValue(context.Background(), "user_id", "user123")
    ctx2 := context.WithValue(context.Background(), "user_id", "user456")
    
    ai.Ask(ctx1, "Question 1")  // User 123's quota
    ai.Ask(ctx2, "Question 2")  // User 456's quota (separate)
}
```

### Example 3: Graceful Degradation

```go
package main

import (
    "context"
    "time"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    ctx := context.Background()
    
    ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithRateLimitPerSecond(10, 5)
    
    // Strategy 1: Wait for availability
    err := ai.WaitForRateLimit(ctx, "user123")
    if err == nil {
        resp, _ := ai.Ask(ctx, "Question")
        fmt.Println(resp)
    }
    
    // Strategy 2: Check before making request
    allowed, _ := ai.CheckRateLimit(ctx, "user123")
    if allowed {
        resp, _ := ai.Ask(ctx, "Question")
        fmt.Println(resp)
    } else {
        // Fallback to cached response
        resp := getCachedResponse("Question")
        fmt.Println("Cached:", resp)
    }
    
    // Strategy 3: Reserve slot
    reservation, err := ai.ReserveRateLimit(ctx, "user123")
    if err == nil {
        time.Sleep(reservation.Delay())  // Wait if needed
        resp, _ := ai.Ask(ctx, "Question")
        fmt.Println(resp)
    }
}
```

### Example 4: Multi-Tier Rate Limiting

```go
package main

import (
    "context"
    "github.com/taipm/go-deep-agent/agent"
)

func createAgentForTier(tier string, apiKey string) *agent.Builder {
    var rate, burst int
    
    switch tier {
    case "free":
        rate = 10    // 10 req/min
        burst = 2    // 2 burst
    case "basic":
        rate = 100   // 100 req/min
        burst = 10   // 10 burst
    case "premium":
        rate = 1000  // 1000 req/min
        burst = 100  // 100 burst
    case "enterprise":
        rate = 10000 // Unlimited (high limit)
        burst = 1000
    }
    
    return agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithRateLimitPerMinute(rate, burst)
}

func main() {
    ctx := context.Background()
    
    // Free tier user
    freeAI := createAgentForTier("free", apiKey)
    freeAI.Ask(ctx, "Question")  // Limited to 10/min
    
    // Premium tier user
    premiumAI := createAgentForTier("premium", apiKey)
    premiumAI.Ask(ctx, "Question")  // Can do 1000/min
}
```

### Example 5: Rate Limit with Metrics

```go
package main

import (
    "context"
    "github.com/taipm/go-deep-agent/agent"
    "github.com/prometheus/client_golang/prometheus"
)

var (
    rateLimitCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rate_limit_exceeded_total",
            Help: "Total number of rate limit exceeded errors",
        },
        []string{"user_id"},
    )
)

func main() {
    ctx := context.Background()
    
    config := &agent.RateLimitConfig{
        Strategy: "token-bucket",
        Rate:     100,
        Burst:    10,
        OnRateLimitExceeded: func(ctx context.Context, key string) {
            // Emit metric
            rateLimitCounter.WithLabelValues(key).Inc()
            
            // Log
            log.Printf("Rate limit exceeded: %s", key)
            
            // Alert (if threshold exceeded)
            if getExceededCount(key) > 100 {
                sendAlert("Possible abuse detected", key)
            }
        },
    }
    
    ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithRateLimit(config)
    
    // Usage
    _, err := ai.Ask(ctx, "Question")
    if err == agent.ErrRateLimitExceeded {
        // Handle rate limit
        stats, _ := ai.GetRateLimitStats(ctx)
        fmt.Printf("Retry after: %v\n", stats.RetryAfter)
    }
}
```

---

## Best Practices

### 1. Chọn chiến lược phù hợp

```
Fixed Window:
- ✅ Use: Simple APIs, non-critical applications
- ❌ Avoid: Paid APIs, multi-tenant systems

Sliding Window:
- ✅ Use: High-accuracy requirements
- ❌ Avoid: High-throughput systems (memory intensive)

Token Bucket: ⭐ RECOMMENDED
- ✅ Use: Most production scenarios
- ✅ Allows burst, smooth traffic
- ✅ Good performance

Leaky Bucket:
- ✅ Use: When need strict output rate
- ❌ Avoid: Real-time applications (high latency)
```

### 2. Thiết lập limits hợp lý

```go
// ❌ BAD: Quá restrictive
ai.WithRateLimitPerSecond(1, 1)  // Users sẽ frustrate

// ❌ BAD: Quá lỏng lẻo
ai.WithRateLimitPerSecond(10000, 5000)  // Không có tác dụng

// ✅ GOOD: Cân bằng
ai.WithRateLimitPerMinute(100, 10)  // Reasonable cho most users
```

**Cách tính limits:**

```
OpenAI Tier 5 limit: 10,000 RPM
Expected users: 100
Safety margin: 20%

Per-user limit: 10,000 * 0.8 / 100 = 80 req/min
Burst: 10-20% of rate = 8-16 (chọn 10)
```

### 3. Implement graceful degradation

```go
// ✅ GOOD: Fallback strategy
resp, err := ai.Ask(ctx, question)
if err == agent.ErrRateLimitExceeded {
    // Strategy 1: Wait and retry
    stats, _ := ai.GetRateLimitStats(ctx)
    time.Sleep(stats.RetryAfter)
    resp, err = ai.Ask(ctx, question)
}

if err == agent.ErrRateLimitExceeded {
    // Strategy 2: Use cache
    resp = getCachedResponse(question)
}

if resp == "" {
    // Strategy 3: Return error with helpful message
    return fmt.Errorf("service busy, retry after %v", stats.RetryAfter)
}
```

### 4. Thông báo rõ ràng cho users

```go
// ✅ GOOD: Clear error message
if err == agent.ErrRateLimitExceeded {
    stats, _ := ai.GetRateLimitStats(ctx)
    
    return &APIResponse{
        Error: "Rate limit exceeded",
        Message: fmt.Sprintf(
            "You've used %d/%d requests. Please retry after %v",
            stats.Limit - stats.Remaining,
            stats.Limit,
            stats.RetryAfter,
        ),
        RetryAfter: stats.ResetAt,
    }
}
```

### 5. Monitor và alert

```go
// Track metrics
rateLimitHits.Inc()
rateLimitRemaining.Set(float64(stats.Remaining))

// Alert on high rate limit hits
if hitRate > 0.8 {
    alert("80% of users hitting rate limit - consider increasing")
}

// Alert on abuse
if userHitCount > 1000 {
    alert("Possible abuse detected for user: " + userID)
}
```

### 6. Testing

```go
func TestRateLimit(t *testing.T) {
    ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithRateLimitPerSecond(10, 5)
    
    // Test burst
    for i := 0; i < 5; i++ {
        _, err := ai.Ask(ctx, "Question")
        assert.NoError(t, err)  // Should succeed
    }
    
    // Test rate limit
    for i := 0; i < 10; i++ {
        _, err := ai.Ask(ctx, "Question")
        if err == agent.ErrRateLimitExceeded {
            t.Log("Rate limited as expected")
            return
        }
    }
    
    t.Error("Should have hit rate limit")
}
```

---

## Implementation Roadmap

### Phase 1: Core Implementation (Week 1)

- [ ] `RateLimiter` interface
- [ ] Token Bucket implementation (memory)
- [ ] Builder API methods
- [ ] Error types and codes
- [ ] Unit tests

### Phase 2: Advanced Features (Week 2)

- [ ] Redis-backed rate limiter
- [ ] Per-user, per-IP, per-API-key strategies
- [ ] Rate limit statistics
- [ ] Metrics integration
- [ ] Documentation

### Phase 3: Testing & Optimization

- [ ] Load testing
- [ ] Benchmark suite
- [ ] Examples
- [ ] Production deployment guide

---

## Tài liệu tham khảo

1. **golang.org/x/time/rate** - Go standard rate limiting library
2. **NGINX Rate Limiting** - Best practices guide
3. **Stripe API Rate Limiting** - Real-world example
4. **Redis Rate Limiting Patterns** - Distributed rate limiting
5. **RFC 6585** - HTTP Status Code 429 (Too Many Requests)

---

## FAQ

**Q: Rate limiting khác gì với throttling?**

A: 
- **Rate Limiting**: Hard limit, block requests khi vượt quota
- **Throttling**: Slow down requests, vẫn process nhưng chậm hơn

**Q: Nên dùng memory hay Redis?**

A:
- **Memory**: Single instance, đơn giản, nhanh
- **Redis**: Multi-instance, distributed, persistent

**Q: Token bucket vs leaky bucket?**

A:
- **Token Bucket**: Allows burst, smooth over time (RECOMMENDED)
- **Leaky Bucket**: Strict output rate, higher latency

**Q: Làm sao test rate limiting?**

A: 
```go
// Use time.Sleep hoặc mock time
ai.WithRateLimitPerSecond(10, 5)
// Send 20 requests rapidly
// Expect: 5 succeed immediately, 5 succeed after delay, 10 fail
```

---

**Status:** Draft - Ready for implementation  
**Next Steps:** Create GitHub issue for v0.8.0  
**Owner:** TBD  
**Target Release:** v0.8.0
