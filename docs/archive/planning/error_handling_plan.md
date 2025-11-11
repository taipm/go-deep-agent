# 🔥 ERROR HANDLING & DEBUGGING - PHÂN TÍCH CHI TIẾT & KẾ HOẠCH CẢI THIỆN

**Ngày phân tích**: 10/11/2025  
**Phiên bản hiện tại**: v0.5.8  
**Điểm hiện tại**: 85/100  
**Mục tiêu**: 95/100 (World-class error handling)

---

## 📊 ĐÁNH GIÁ HIỆN TRẠNG

### ✅ Điểm mạnh (Đã có)

1. **Custom Error Types** ✅
   ```go
   // errors.go - 7 sentinel errors
   var (
       ErrAPIKey           = errors.New("API key is missing or invalid")
       ErrRateLimit        = errors.New("rate limit exceeded")
       ErrTimeout          = errors.New("request timeout")
       ErrRefusal          = errors.New("content refused by model")
       ErrInvalidResponse  = errors.New("invalid response from API")
       ErrMaxRetries       = errors.New("maximum retry attempts exceeded")
       ErrToolExecution    = errors.New("tool execution failed")
   )
   ```

2. **Error Type Checking** ✅
   ```go
   // Helper functions
   IsAPIKeyError(err)
   IsRateLimitError(err)
   IsTimeoutError(err)
   IsRefusalError(err)
   IsMaxRetriesError(err)
   IsToolExecutionError(err)
   ```

3. **APIError Struct** ✅
   ```go
   type APIError struct {
       Type       string  // Error type
       Message    string  // Error message
       StatusCode int     // HTTP status code
       Err        error   // Underlying error
   }
   ```

4. **Error Wrapping** ✅
   ```go
   WrapAPIKey(err)
   WrapRateLimit(err)
   WrapTimeout(err)
   WrapRefusal(message)
   WrapMaxRetries(attempts, lastErr)
   WrapToolExecution(toolName, err)
   ```

5. **Comprehensive Tests** ✅
   - `errors_test.go`: 380 lines, 15+ test functions
   - Error type checking tests
   - Error wrapping tests
   - Retry logic tests

### ⚠️ Điểm yếu (Cần cải thiện)

#### 1. **Error Messages không user-friendly** (⭐⭐⭐)

**Vấn đề:**
```go
// Hiện tại - Technical, khó hiểu
err := fmt.Errorf("failed to generate embedding: %w", err)
// User nhìn thấy: "failed to generate embedding: context deadline exceeded"
// Không biết phải làm gì!

err := fmt.Errorf("vector store must be configured with WithVectorRAG")
// User: "WithVectorRAG là gì? Làm sao config?"
```

**Cần có:**
```go
// User-friendly với actionable advice
type UserFriendlyError struct {
    Code        string   // "EMBEDDING_TIMEOUT"
    Message     string   // "Embedding generation timed out"
    Cause       string   // "The embedding API took too long to respond"
    Solution    string   // "Try: 1) Increase timeout, 2) Reduce text length, 3) Check network"
    DocsURL     string   // "https://go-deep-agent.dev/docs/errors#EMBEDDING_TIMEOUT"
    Err         error    // Original error
}
```

---

#### 2. **Thiếu Error Codes** (⭐⭐⭐⭐)

**Vấn đề:**
```go
// Không có error codes
if err != nil {
    // Làm sao phân biệt errors?
    // Phải parse string message? BAD!
}
```

**Cần có:**
```go
// Error codes cho programmatic handling
const (
    // API Errors (1000-1999)
    ErrCodeAPIKeyMissing    = "API_KEY_MISSING"      // 1001
    ErrCodeAPIKeyInvalid    = "API_KEY_INVALID"      // 1002
    ErrCodeRateLimit        = "RATE_LIMIT_EXCEEDED"  // 1003
    ErrCodeTimeout          = "REQUEST_TIMEOUT"      // 1004
    
    // Tool Errors (2000-2999)
    ErrCodeToolNotFound     = "TOOL_NOT_FOUND"       // 2001
    ErrCodeToolPanic        = "TOOL_PANICKED"        // 2002
    ErrCodeToolTimeout      = "TOOL_TIMEOUT"         // 2003
    
    // RAG Errors (3000-3999)
    ErrCodeVectorStoreNotConfigured = "VECTOR_STORE_NOT_CONFIGURED"  // 3001
    ErrCodeEmbeddingFailed          = "EMBEDDING_GENERATION_FAILED"   // 3002
    
    // Memory Errors (4000-4999)
    ErrCodeMemoryFull       = "MEMORY_CAPACITY_FULL" // 4001
    
    // Cache Errors (5000-5999)
    ErrCodeCacheConnection  = "CACHE_CONNECTION_FAILED"  // 5001
    ErrCodeCacheOperation   = "CACHE_OPERATION_FAILED"   // 5002
)
```

---

#### 3. **Thiếu Error Context** (⭐⭐⭐⭐)

**Vấn đề:**
```go
// Khi error xảy ra, không biết context
err := ai.Ask(ctx, "Hello")
// Error: "rate limit exceeded"
// Không biết:
// - Request nào bị limit?
// - Model nào?
// - Limit là bao nhiêu?
// - Khi nào có thể retry?
```

**Cần có:**
```go
type RichError struct {
    Code        string
    Message     string
    Context     map[string]interface{}{
        "model":         "gpt-4",
        "request_id":    "req_abc123",
        "retry_after":   60,  // seconds
        "limit":         "90 requests/min",
        "usage":         "85/90",
        "reset_time":    time.Now().Add(60*time.Second),
    }
    Stack       []string  // Stack trace
    Timestamp   time.Time
    Err         error
}
```

---

#### 4. **Thiếu Error Recovery Guide** (⭐⭐⭐⭐⭐)

**Vấn đề:**
- Không có troubleshooting docs
- User không biết cách fix errors
- Không có common errors list

**Cần có:**
```markdown
# docs/TROUBLESHOOTING.md

## Common Errors & Solutions

### 1. RATE_LIMIT_EXCEEDED

**Symptoms:**
```
Error: rate limit exceeded (status 429)
```

**Causes:**
- Too many requests in short time
- Exceeding tier limits
- Multiple agents sharing same key

**Solutions:**
1. Add retry with backoff (already done with WithDefaults())
2. Increase WithRetryDelay(5 * time.Second)
3. Implement request batching
4. Upgrade OpenAI tier
5. Use multiple API keys with load balancing

**Code example:**
```go
ai := agent.NewOpenAI("gpt-4", key).
    WithDefaults().              // Has retry + backoff
    WithRetryDelay(5*time.Second) // Longer delay
```

**See also:**
- [Rate Limits Docs](https://platform.openai.com/docs/guides/rate-limits)
- [Error Handling Best Practices](./ERROR_HANDLING.md)
```

---

#### 5. **Debug Mode chưa tối ưu** (⭐⭐⭐)

**Vấn đề:**
```go
// Debug logging không đủ chi tiết
builder := agent.NewOpenAI("gpt-4", key).WithDebugLogging()
// Log gì? Ở đâu? Format thế nào?
```

**Cần có:**
```go
// Rich debug mode
type DebugConfig struct {
    LogLevel        string    // "debug", "trace", "verbose"
    LogRequests     bool      // Log all API requests
    LogResponses    bool      // Log all API responses
    LogTokens       bool      // Log token usage
    LogErrors       bool      // Log all errors with stack trace
    LogTools        bool      // Log tool executions
    LogMemory       bool      // Log memory operations
    LogCache        bool      // Log cache hits/misses
    RedactSecrets   bool      // Redact API keys in logs
    Output          io.Writer // Where to write logs
}

// Usage:
ai := agent.NewOpenAI("gpt-4", key).
    WithDebug(DebugConfig{
        LogLevel:      "trace",
        LogRequests:   true,
        LogResponses:  true,
        LogTokens:     true,
        RedactSecrets: true,  // Don't leak API keys!
        Output:        os.Stderr,
    })
```

---

#### 6. **Error Metrics & Monitoring** (⭐⭐⭐⭐)

**Vấn đề:**
- Không track error rates
- Không có error analytics
- Không có alerts

**Cần có:**
```go
// Error metrics
type ErrorMetrics struct {
    sync.RWMutex
    
    TotalErrors       int64
    ErrorsByType      map[string]int64  // "RATE_LIMIT": 15, "TIMEOUT": 3
    ErrorsByModel     map[string]int64  // "gpt-4": 10, "gpt-4o": 8
    RetriedRequests   int64
    FailedAfterRetry  int64
    AverageRetries    float64
    
    LastError         error
    LastErrorTime     time.Time
    
    // Time-series data (last hour)
    ErrorTimeSeries   []ErrorEvent
}

type ErrorEvent struct {
    Timestamp time.Time
    Code      string
    Message   string
    Model     string
    Retried   bool
}

// Usage:
metrics := ai.GetErrorMetrics()
fmt.Printf("Error rate: %.2f%%\n", metrics.ErrorRate())
fmt.Printf("Most common: %s (%d times)\n", metrics.MostCommonError())
```

---

#### 7. **Panic Recovery không đầy đủ** (⭐⭐⭐)

**Vấn đề:**
```go
// Hiện tại - Chỉ có trong tool execution
func executeToolWithRecovery() (result string, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("tool panicked: %v", r)
        }
    }()
    // ...
}

// Thiếu panic recovery ở:
// - Ask() / Stream()
// - Batch processing
// - Memory operations
// - Cache operations
```

**Cần có:**
```go
// Global panic recovery
func (b *Builder) withPanicRecovery(fn func() error) (err error) {
    defer func() {
        if r := recover(); r != nil {
            stack := debug.Stack()
            err = &PanicError{
                Value:      r,
                Stack:      string(stack),
                Timestamp:  time.Now(),
            }
            
            // Log panic
            b.getLogger().Error(context.Background(), "PANIC RECOVERED", 
                Field("panic", r),
                Field("stack", string(stack)),
            )
        }
    }()
    return fn()
}

// Usage:
func (b *Builder) Ask(ctx context.Context, message string) (string, error) {
    var result string
    err := b.withPanicRecovery(func() error {
        var e error
        result, e = b.askInternal(ctx, message)
        return e
    })
    return result, err
}
```

---

## 🎯 KẾ HOẠCH CẢI THIỆN (3 Tuần)

### 📅 Week 1: Foundation (Error Codes & Rich Errors)

**Goal**: Thêm error codes và rich error context

#### Day 1-2: Error Codes System (16h)

**Files to create:**
```
agent/
├── error_codes.go          # Error code constants
├── error_rich.go           # RichError struct
└── error_codes_test.go     # Tests
```

**Tasks:**
1. Define error code constants (100+ codes)
2. Create RichError struct with context
3. Update all error returns to use codes
4. Write comprehensive tests

**Code:**
```go
// agent/error_codes.go
package agent

const (
    // Category: API Errors (1xxx)
    ErrCodeAPIKeyMissing        = "API_KEY_MISSING"
    ErrCodeAPIKeyInvalid        = "API_KEY_INVALID"
    ErrCodeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"
    ErrCodeRequestTimeout       = "REQUEST_TIMEOUT"
    ErrCodeInvalidRequest       = "INVALID_REQUEST"
    ErrCodeInsufficientQuota    = "INSUFFICIENT_QUOTA"
    
    // Category: Tool Errors (2xxx)
    ErrCodeToolNotFound         = "TOOL_NOT_FOUND"
    ErrCodeToolExecutionFailed  = "TOOL_EXECUTION_FAILED"
    ErrCodeToolPanicked         = "TOOL_PANICKED"
    ErrCodeToolTimeout          = "TOOL_TIMEOUT"
    ErrCodeToolInvalidArgs      = "TOOL_INVALID_ARGS"
    
    // Category: RAG Errors (3xxx)
    ErrCodeVectorStoreNotConfigured  = "VECTOR_STORE_NOT_CONFIGURED"
    ErrCodeEmbeddingFailed           = "EMBEDDING_GENERATION_FAILED"
    ErrCodeDocumentChunkingFailed    = "DOCUMENT_CHUNKING_FAILED"
    ErrCodeVectorSearchFailed        = "VECTOR_SEARCH_FAILED"
    
    // Category: Memory Errors (4xxx)
    ErrCodeMemoryFull           = "MEMORY_CAPACITY_FULL"
    ErrCodeMemoryCorrupted      = "MEMORY_CORRUPTED"
    
    // Category: Cache Errors (5xxx)
    ErrCodeCacheConnectionFailed = "CACHE_CONNECTION_FAILED"
    ErrCodeCacheOperationFailed  = "CACHE_OPERATION_FAILED"
    ErrCodeCacheKeyNotFound      = "CACHE_KEY_NOT_FOUND"
    
    // Category: Configuration Errors (6xxx)
    ErrCodeInvalidConfiguration = "INVALID_CONFIGURATION"
    ErrCodeMissingConfiguration = "MISSING_CONFIGURATION"
)

// Error metadata
var ErrorMetadata = map[string]ErrorMeta{
    ErrCodeAPIKeyMissing: {
        HTTPStatus: 401,
        Severity:   "critical",
        Category:   "authentication",
        Retryable:  false,
        DocsURL:    "https://go-deep-agent.dev/docs/errors#api-key-missing",
    },
    ErrCodeRateLimitExceeded: {
        HTTPStatus: 429,
        Severity:   "warning",
        Category:   "rate-limit",
        Retryable:  true,
        DocsURL:    "https://go-deep-agent.dev/docs/errors#rate-limit",
    },
    // ... 100+ more
}

type ErrorMeta struct {
    HTTPStatus int
    Severity   string  // "critical", "error", "warning", "info"
    Category   string
    Retryable  bool
    DocsURL    string
}
```

```go
// agent/error_rich.go
package agent

import (
    "fmt"
    "runtime/debug"
    "time"
)

type RichError struct {
    // Core fields
    Code        string                 // Error code (e.g., "RATE_LIMIT_EXCEEDED")
    Message     string                 // Human-readable message
    Cause       string                 // What caused this error
    Solution    string                 // How to fix it
    
    // Context
    Context     map[string]interface{} // Additional context
    Timestamp   time.Time              // When error occurred
    
    // Debug info
    Stack       string                 // Stack trace (if available)
    RequestID   string                 // Request ID for tracing
    
    // Metadata
    Severity    string                 // "critical", "error", "warning"
    Retryable   bool                   // Can retry?
    DocsURL     string                 // Link to documentation
    
    // Underlying error
    Err         error                  // Wrapped error
}

func NewRichError(code string, err error) *RichError {
    meta := ErrorMetadata[code]
    
    return &RichError{
        Code:      code,
        Message:   generateMessage(code, err),
        Cause:     generateCause(code, err),
        Solution:  generateSolution(code),
        Context:   make(map[string]interface{}),
        Timestamp: time.Now(),
        Stack:     string(debug.Stack()),
        Severity:  meta.Severity,
        Retryable: meta.Retryable,
        DocsURL:   meta.DocsURL,
        Err:       err,
    }
}

func (e *RichError) Error() string {
    return fmt.Sprintf("[%s] %s (see: %s)", e.Code, e.Message, e.DocsURL)
}

func (e *RichError) Unwrap() error {
    return e.Err
}

func (e *RichError) WithContext(key string, value interface{}) *RichError {
    e.Context[key] = value
    return e
}

func (e *RichError) WithRequestID(id string) *RichError {
    e.RequestID = id
    return e
}

// User-friendly formatted output
func (e *RichError) Format() string {
    return fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════
║ ERROR: %s
╠══════════════════════════════════════════════════════════════
║ 
║ ❌ What happened:
║    %s
║ 
║ 🔍 Why it happened:
║    %s
║ 
║ ✅ How to fix:
║    %s
║ 
║ 📚 Documentation:
║    %s
║ 
║ ⏰ Time: %s
║ 🆔 Request ID: %s
╚══════════════════════════════════════════════════════════════
`, e.Code, e.Message, e.Cause, e.Solution, e.DocsURL, 
   e.Timestamp.Format(time.RFC3339), e.RequestID)
}
```

**Deliverables:**
- ✅ 100+ error codes defined
- ✅ RichError struct implemented
- ✅ Error metadata complete
- ✅ 20+ tests passing

---

#### Day 3-4: Update Existing Code (16h)

**Tasks:**
1. Update all `fmt.Errorf()` to use RichError
2. Add context to all errors
3. Update error wrapping functions

**Example migration:**
```go
// Before:
func (b *Builder) Ask(ctx context.Context, msg string) (string, error) {
    if b.apiKey == "" {
        return "", fmt.Errorf("API key is missing")
    }
    // ...
}

// After:
func (b *Builder) Ask(ctx context.Context, msg string) (string, error) {
    if b.apiKey == "" {
        return "", NewRichError(ErrCodeAPIKeyMissing, nil).
            WithContext("method", "Ask").
            WithContext("model", b.model)
    }
    // ...
}
```

**Deliverables:**
- ✅ 50+ files updated
- ✅ All errors have codes
- ✅ All errors have context
- ✅ Tests passing

---

#### Day 5: Error Messages Helper (8h)

**Tasks:**
1. Create message generation helpers
2. Create actionable solutions database

**Code:**
```go
// agent/error_messages.go
package agent

var ErrorSolutions = map[string]string{
    ErrCodeAPIKeyMissing: `
1. Set OPENAI_API_KEY environment variable
2. Or pass API key to NewOpenAI("model", "sk-...")
3. Get API key from: https://platform.openai.com/api-keys
`,
    ErrCodeRateLimitExceeded: `
1. Use WithDefaults() - includes retry with backoff
2. Increase retry delay: WithRetryDelay(5 * time.Second)
3. Reduce request rate
4. Upgrade OpenAI tier: https://platform.openai.com/account/limits
5. Use multiple API keys with load balancing
`,
    ErrCodeToolTimeout: `
1. Increase tool timeout: WithToolTimeout(60 * time.Second)
2. Optimize tool function (reduce complexity)
3. Enable parallel execution: WithParallelTools(true)
4. Check tool dependencies (network, database, etc.)
`,
    // ... 100+ more solutions
}

func generateSolution(code string) string {
    if solution, ok := ErrorSolutions[code]; ok {
        return solution
    }
    return "Check documentation for more information."
}
```

**Deliverables:**
- ✅ 100+ error solutions
- ✅ Helper functions complete
- ✅ User-friendly messages

---

### 📅 Week 2: Debug Tools & Monitoring (40h)

#### Day 1-2: Enhanced Debug Mode (16h)

**Files:**
```
agent/
├── debug.go              # Debug configuration
├── debug_interceptor.go  # Request/response logging
└── debug_test.go         # Tests
```

**Code:**
```go
// agent/debug.go
package agent

import (
    "context"
    "io"
    "os"
    "time"
)

type DebugConfig struct {
    // What to log
    LogRequests       bool  // Log API requests
    LogResponses      bool  // Log API responses  
    LogTokens         bool  // Log token usage
    LogErrors         bool  // Log errors with stack traces
    LogTools          bool  // Log tool executions
    LogMemory         bool  // Log memory operations
    LogCache          bool  // Log cache hits/misses
    LogRetries        bool  // Log retry attempts
    
    // How to log
    LogLevel          string    // "debug", "trace", "verbose"
    PrettyPrint       bool      // Pretty-print JSON
    RedactSecrets     bool      // Redact API keys (default: true)
    MaxBodySize       int       // Max request/response body to log (bytes)
    
    // Where to log
    Output            io.Writer // Where to write (default: os.Stderr)
    
    // Performance
    EnableProfiling   bool      // Enable CPU/memory profiling
    ProfileOutput     string    // Profiling output file
}

func DefaultDebugConfig() DebugConfig {
    return DebugConfig{
        LogRequests:    true,
        LogResponses:   true,
        LogErrors:      true,
        LogLevel:       "debug",
        PrettyPrint:    true,
        RedactSecrets:  true,  // SECURITY: Always redact by default
        MaxBodySize:    10000, // 10KB
        Output:         os.Stderr,
    }
}

func (b *Builder) WithDebug(config DebugConfig) *Builder {
    b.debugConfig = config
    b.debugMode = true
    return b
}

// Debug interceptor
func (b *Builder) debugLogRequest(ctx context.Context, req interface{}) {
    if !b.debugMode || !b.debugConfig.LogRequests {
        return
    }
    
    // Redact secrets
    safeReq := b.redactSecrets(req)
    
    // Pretty print
    if b.debugConfig.PrettyPrint {
        json, _ := json.MarshalIndent(safeReq, "", "  ")
        fmt.Fprintf(b.debugConfig.Output, "🔵 REQUEST:\n%s\n", json)
    } else {
        fmt.Fprintf(b.debugConfig.Output, "🔵 REQUEST: %+v\n", safeReq)
    }
}
```

**Deliverables:**
- ✅ Full debug config
- ✅ Request/response logging
- ✅ Secret redaction
- ✅ Performance profiling

---

#### Day 3-4: Error Metrics & Analytics (16h)

**Code:**
```go
// agent/error_metrics.go
package agent

import (
    "sync"
    "time"
)

type ErrorMetrics struct {
    mu sync.RWMutex
    
    // Counters
    TotalErrors       int64
    TotalRequests     int64
    ErrorsByCode      map[string]int64
    ErrorsByModel     map[string]int64
    ErrorsBySeverity  map[string]int64
    
    // Retry stats
    TotalRetries      int64
    SuccessfulRetries int64
    FailedRetries     int64
    
    // Timing
    AverageRetryDelay time.Duration
    
    // Recent errors (last 100)
    RecentErrors      []ErrorEvent
    
    // Time series (last hour, 60 buckets)
    ErrorTimeSeries   [60]int
    CurrentBucket     int
    LastBucketTime    time.Time
}

type ErrorEvent struct {
    Timestamp time.Time
    Code      string
    Message   string
    Model     string
    Severity  string
    Context   map[string]interface{}
}

func (m *ErrorMetrics) RecordError(err *RichError) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.TotalErrors++
    m.ErrorsByCode[err.Code]++
    m.ErrorsBySeverity[err.Severity]++
    
    if model, ok := err.Context["model"].(string); ok {
        m.ErrorsByModel[model]++
    }
    
    // Add to recent errors (max 100)
    event := ErrorEvent{
        Timestamp: err.Timestamp,
        Code:      err.Code,
        Message:   err.Message,
        Severity:  err.Severity,
        Context:   err.Context,
    }
    
    m.RecentErrors = append(m.RecentErrors, event)
    if len(m.RecentErrors) > 100 {
        m.RecentErrors = m.RecentErrors[1:]
    }
    
    // Update time series
    m.updateTimeSeries()
}

func (m *ErrorMetrics) ErrorRate() float64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if m.TotalRequests == 0 {
        return 0
    }
    return float64(m.TotalErrors) / float64(m.TotalRequests) * 100
}

func (m *ErrorMetrics) MostCommonError() (string, int64) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    var maxCode string
    var maxCount int64
    
    for code, count := range m.ErrorsByCode {
        if count > maxCount {
            maxCode = code
            maxCount = count
        }
    }
    
    return maxCode, maxCount
}

func (m *ErrorMetrics) Report() string {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    commonCode, commonCount := m.MostCommonError()
    
    return fmt.Sprintf(`
ERROR METRICS REPORT
====================
Total Errors:      %d
Total Requests:    %d
Error Rate:        %.2f%%
Most Common:       %s (%d times)
Successful Retries: %d / %d (%.1f%%)

Errors by Severity:
- Critical:        %d
- Error:           %d  
- Warning:         %d

Top 5 Errors:
%s
`, m.TotalErrors, m.TotalRequests, m.ErrorRate(), 
   commonCode, commonCount,
   m.SuccessfulRetries, m.TotalRetries, 
   float64(m.SuccessfulRetries)/float64(m.TotalRetries)*100,
   m.ErrorsBySeverity["critical"],
   m.ErrorsBySeverity["error"],
   m.ErrorsBySeverity["warning"],
   m.topErrorsReport())
}

// Usage:
ai := agent.NewOpenAI("gpt-4", key).WithDefaults()
metrics := ai.GetErrorMetrics()

// After some operations...
fmt.Println(metrics.Report())
```

**Deliverables:**
- ✅ Error metrics tracking
- ✅ Time-series data
- ✅ Analytics reports
- ✅ Integration with Builder

---

#### Day 5: Panic Recovery (8h)

**Code:**
```go
// agent/panic_recovery.go
package agent

import (
    "context"
    "fmt"
    "runtime/debug"
    "time"
)

type PanicError struct {
    Value     interface{}
    Stack     string
    Timestamp time.Time
    Context   map[string]interface{}
}

func (e *PanicError) Error() string {
    return fmt.Sprintf("panic recovered: %v", e.Value)
}

func (b *Builder) withPanicRecovery(fn func() error) (err error) {
    defer func() {
        if r := recover(); r != nil {
            panicErr := &PanicError{
                Value:     r,
                Stack:     string(debug.Stack()),
                Timestamp: time.Now(),
                Context:   make(map[string]interface{}),
            }
            
            // Log panic
            b.getLogger().Error(context.Background(), "PANIC RECOVERED",
                Field("panic_value", r),
                Field("stack_trace", panicErr.Stack),
            )
            
            // Record in metrics
            if b.errorMetrics != nil {
                b.errorMetrics.RecordPanic(panicErr)
            }
            
            err = panicErr
        }
    }()
    
    return fn()
}

// Update all public methods
func (b *Builder) Ask(ctx context.Context, message string) (string, error) {
    var result string
    err := b.withPanicRecovery(func() error {
        var e error
        result, e = b.askInternal(ctx, message)
        return e
    })
    return result, err
}
```

**Deliverables:**
- ✅ Global panic recovery
- ✅ Applied to all public methods
- ✅ Panic logging
- ✅ Tests with panic scenarios

---

### 📅 Week 3: Documentation & Testing (40h)

#### Day 1-3: TROUBLESHOOTING.md (24h)

**File:** `docs/TROUBLESHOOTING.md` (500+ lines)

**Structure:**
```markdown
# Troubleshooting Guide

## Table of Contents
1. Quick Diagnostic Checklist
2. Common Errors (A-Z)
3. Error Code Reference
4. Debug Tools
5. Performance Issues
6. FAQ

## 1. Quick Diagnostic Checklist

Before diving into specific errors, run this checklist:

- [ ] API key is set correctly
- [ ] Network connection is working
- [ ] Using latest version: `go get -u github.com/taipm/go-deep-agent`
- [ ] Check OpenAI status: https://status.openai.com
- [ ] Enable debug mode: `.WithDebug(DefaultDebugConfig())`
- [ ] Check error metrics: `ai.GetErrorMetrics().Report()`

## 2. Common Errors

### API_KEY_MISSING

**Error message:**
```
[API_KEY_MISSING] API key is missing or invalid
```

**Symptoms:**
- 401 Unauthorized responses
- "invalid_api_key" errors

**Root causes:**
1. OPENAI_API_KEY environment variable not set
2. API key passed to NewOpenAI() is empty
3. API key passed is incorrect format

**Solutions:**

Option 1: Environment variable
```bash
export OPENAI_API_KEY="sk-..."
```

Option 2: Code
```go
ai := agent.NewOpenAI("gpt-4", "sk-your-key-here")
```

Option 3: Check key format
- Must start with "sk-"
- Must be 51 characters long
- Get from: https://platform.openai.com/api-keys

**Verification:**
```go
// Test your key
ai := agent.NewOpenAI("gpt-4", key).WithDebug(DefaultDebugConfig())
resp, err := ai.Ask(context.Background(), "test")
if err != nil {
    fmt.Println(ai.GetErrorMetrics().Report())
}
```

### RATE_LIMIT_EXCEEDED

**Error message:**
```
[RATE_LIMIT_EXCEEDED] Rate limit exceeded: 90 requests/min
```

**Symptoms:**
- 429 Too Many Requests
- Intermittent failures during high load

**Root causes:**
1. Too many requests in short time
2. Exceeding tier limits (Free/Tier 1/2/3/4/5)
3. Multiple agents sharing same key
4. No retry logic

**Solutions:**

✅ **Best practice:** Use WithDefaults() (includes retry + backoff)
```go
ai := agent.NewOpenAI("gpt-4", key).WithDefaults()
```

Advanced configuration:
```go
ai := agent.NewOpenAI("gpt-4", key).
    WithRetry(5).                        // Retry up to 5 times
    WithRetryDelay(2 * time.Second).     // Start with 2s delay
    WithExponentialBackoff().            // Increase: 2s, 4s, 8s, 16s, 32s
    WithTimeout(60 * time.Second)        // Max 60s total
```

Check limits:
```go
// After error, check when you can retry
if err != nil {
    if richErr, ok := err.(*RichError); ok {
        if retryAfter, ok := richErr.Context["retry_after"].(int); ok {
            fmt.Printf("Retry after %d seconds\n", retryAfter)
        }
    }
}
```

**Long-term solutions:**
1. Upgrade tier: https://platform.openai.com/account/limits
2. Implement request batching
3. Use caching: `.WithRedisCache("localhost:6379", "", 0)`
4. Load balance across multiple keys

### TOOL_TIMEOUT

[... 50+ more errors documented ...]

## 3. Error Code Reference

[Full table of 100+ error codes]

## 4. Debug Tools

### Enable Debug Mode
[...]

### Analyze Error Metrics
[...]

## 5. Performance Issues

### Slow Responses
[...]

### High Memory Usage
[...]

## 6. FAQ

Q: How to handle errors gracefully?
Q: Should I retry on all errors?
Q: How to log errors for debugging?
[... 20+ FAQs ...]
```

**Deliverables:**
- ✅ 500+ lines of docs
- ✅ 50+ common errors
- ✅ 100+ error codes documented
- ✅ 20+ FAQs

---

#### Day 4-5: ERROR_HANDLING_GUIDE.md (16h)

**File:** `docs/ERROR_HANDLING_GUIDE.md` (300+ lines)

**Structure:**
```markdown
# Error Handling Best Practices

## Philosophy

go-deep-agent follows Go's error handling philosophy:
- Errors are values
- Handle errors explicitly
- Return errors, don't panic
- Add context to errors
- Make errors actionable

## Error Types

### 1. Sentinel Errors
[...]

### 2. Rich Errors
[...]

### 3. Error Codes
[...]

## Best Practices

### 1. Always Check Errors
[...]

### 2. Add Context
[...]

### 3. Use Error Codes for Programmatic Handling
[...]

### 4. Log Errors Properly
[...]

### 5. Retry Strategically
[...]

## Patterns

### Pattern 1: Graceful Degradation
[...]

### Pattern 2: Circuit Breaker
[...]

### Pattern 3: Error Aggregation
[...]

## Examples

### Example 1: Production Error Handling
[...]

### Example 2: Multi-tier Error Recovery
[...]
```

**Deliverables:**
- ✅ Complete best practices guide
- ✅ 10+ code examples
- ✅ Production patterns

---

## 📊 SUCCESS METRICS

### Before (v0.5.8 - Score: 85/100)

- ✅ Custom error types (7 sentinel errors)
- ✅ Error type checking functions
- ✅ Basic error wrapping
- ❌ No error codes
- ❌ No rich error context
- ❌ No error metrics
- ❌ No troubleshooting guide
- ❌ Debug mode basic
- ❌ Panic recovery partial (tools only)

### After (v0.6.0 - Target Score: 95/100)

- ✅ 100+ error codes
- ✅ Rich errors with context
- ✅ User-friendly error messages
- ✅ Actionable solutions
- ✅ Error metrics & analytics
- ✅ Enhanced debug mode
- ✅ Comprehensive troubleshooting guide
- ✅ Global panic recovery
- ✅ Error documentation (500+ lines)
- ✅ Best practices guide

**Improvement: +10 points (85 → 95)**

---

## 📅 TIMELINE

| Week | Focus | Hours | Deliverables |
|------|-------|-------|--------------|
| Week 1 | Error Codes & Rich Errors | 40h | Codes, RichError, migration |
| Week 2 | Debug & Monitoring | 40h | Debug mode, metrics, panic recovery |
| Week 3 | Documentation | 40h | TROUBLESHOOTING.md, guides |
| **Total** | | **120h** | **World-class error handling** |

**Start**: Week 7 (Nov 17, 2025)  
**End**: Week 9 (Dec 7, 2025)

---

## 🎯 EXPECTED IMPACT

### Developer Experience
- ⏱️ Debug time: 30 min → 5 min (-83%)
- 📚 Error understanding: 60% → 95% (+35%)
- 🔧 Fix success rate: 70% → 95% (+25%)

### Production Reliability
- 🛡️ Error recovery: 80% → 95% (+15%)
- 📊 Error visibility: 40% → 95% (+55%)
- ⚡ MTTR (Mean Time To Recovery): 2h → 30min (-75%)

### User Satisfaction
- ⭐ Error handling rating: 8.5/10 → 9.5/10
- 📖 Documentation rating: 8.0/10 → 9.5/10
- 💡 Self-service success: 60% → 90% (+30%)

---

## ✅ NEXT ACTIONS

1. **Review this plan** - Get feedback
2. **Week 1 Day 1** - Start error codes implementation
3. **Daily standup** - Track progress
4. **Weekly demo** - Show improvements
5. **v0.6.0 release** - Launch with world-class error handling

**Ready to start?** 🚀
