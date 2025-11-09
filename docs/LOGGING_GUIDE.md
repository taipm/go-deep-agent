# Logging Guide

Comprehensive guide to logging in go-deep-agent library.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Log Levels](#log-levels)
- [Built-in Loggers](#built-in-loggers)
- [Custom Loggers](#custom-loggers)
- [Slog Integration](#slog-integration)
- [Production Best Practices](#production-best-practices)
- [What Gets Logged](#what-gets-logged)
- [Performance Considerations](#performance-considerations)
- [Examples](#examples)

## Overview

go-deep-agent provides **opt-in, zero-overhead logging** for observability and debugging. By default, no logging occurs (using `NoopLogger`). You can enable logging by:

1. Using built-in loggers: `WithDebugLogging()` or `WithInfoLogging()`
2. Using slog adapter: `WithLogger(agent.NewSlogAdapter(logger))`
3. Implementing custom logger: `WithLogger(customLogger)`

**Key Features:**
- ✅ Zero overhead when disabled (default)
- ✅ Structured logging with fields
- ✅ Context-aware API
- ✅ Interface-based (compatible with any logger)
- ✅ Go 1.21+ slog support
- ✅ Thread-safe
- ✅ Production-ready

## Quick Start

### Enable Debug Logging (Development)

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDebugLogging() // Enable detailed debug logs

response, err := builder.Ask(ctx, "Hello!")
```

Output:
```
[2024-01-15 10:30:45.123] DEBUG: Ask request started | model=gpt-4o-mini message_length=6
[2024-01-15 10:30:45.124] DEBUG: Cache miss | cache_key=... duration_ms=1
[2024-01-15 10:30:45.890] INFO: Ask request completed | duration_ms=767 prompt_tokens=15 completion_tokens=8 total_tokens=23
```

### Enable Info Logging (Production)

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithInfoLogging() // Production-friendly logging

response, err := builder.Ask(ctx, "Hello!")
```

Output:
```
[2024-01-15 10:30:45.123] INFO: Cache hit | cache_key=... duration_ms=2
[2024-01-15 10:30:46.890] INFO: Ask request completed | duration_ms=1767 total_tokens=23
```

## Log Levels

go-deep-agent supports 5 log levels:

| Level | Value | Description | Use Case |
|-------|-------|-------------|----------|
| `LogLevelNone` | 0 | No logging | Default (zero overhead) |
| `LogLevelError` | 1 | Errors only | Critical failures |
| `LogLevelWarn` | 2 | Warnings + Errors | Non-critical issues |
| `LogLevelInfo` | 3 | Info + Warn + Error | Production monitoring |
| `LogLevelDebug` | 4 | All messages | Development & debugging |

### When to Use Each Level

**DEBUG (LogLevelDebug)**
- Development and debugging
- Tracing request flow
- Investigating issues
- Performance profiling

Logs:
- Request start/end
- Cache hits/misses with keys
- Tool execution details
- RAG retrieval metrics
- Retry attempts with delays

**INFO (LogLevelInfo)**
- Production monitoring
- Performance metrics
- Business metrics
- Audit trails

Logs:
- Request completion with duration
- Token usage
- Cache statistics
- RAG results count
- Tool execution success

**WARN (LogLevelWarn)**
- Non-critical issues
- Degraded performance
- Resource limits

Logs:
- Max retries exceeded
- Refusals from model
- Cache evictions

**ERROR (LogLevelError)**
- Failures requiring attention
- System errors
- Integration failures

Logs:
- Request failures
- Tool execution errors
- RAG retrieval failures
- Cache errors

## Built-in Loggers

### 1. NoopLogger (Default)

**Zero overhead** - does nothing. Used by default.

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey)
// No logging - NoopLogger is used automatically
```

**Performance:**
- No allocation
- No I/O
- Inline-able methods
- Literally zero cost

### 2. StdLogger (Standard Library)

Simple logger using `fmt.Println` to stdout.

```go
// Debug level
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDebugLogging()

// Info level
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithInfoLogging()

// Custom level
logger := agent.NewStdLogger(agent.LogLevelWarn)
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(logger)
```

**Output Format:**
```
[timestamp] LEVEL: message | field1=value1 field2=value2
```

**Example:**
```
[2024-01-15 10:30:45.123] INFO: Ask request completed | duration_ms=767 total_tokens=23
```

### 3. SlogAdapter (Go 1.21+)

Integrates with Go's structured logging (`log/slog`).

```go
import "log/slog"

// Create slog logger
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
slogLogger := slog.New(handler)

// Use with agent
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(agent.NewSlogAdapter(slogLogger))
```

**Supports:**
- ✅ TextHandler (human-readable)
- ✅ JSONHandler (machine-readable)
- ✅ Custom handlers
- ✅ Context propagation
- ✅ Level filtering
- ✅ Structured fields

## Custom Loggers

Implement the `Logger` interface to create custom loggers:

```go
type Logger interface {
    Debug(ctx context.Context, msg string, fields ...Field)
    Info(ctx context.Context, msg string, fields ...Field)
    Warn(ctx context.Context, msg string, fields ...Field)
    Error(ctx context.Context, msg string, fields ...Field)
}
```

### Example: File Logger

```go
type FileLogger struct {
    file *os.File
}

func NewFileLogger(filename string) (*FileLogger, error) {
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return nil, err
    }
    return &FileLogger{file: file}, nil
}

func (l *FileLogger) Info(ctx context.Context, msg string, fields ...agent.Field) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    fmt.Fprintf(l.file, "[%s] INFO: %s\n", timestamp, msg)
}

// Implement other methods...

// Usage
fileLogger, _ := NewFileLogger("app.log")
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(fileLogger)
```

### Example: Zap Adapter

```go
import "go.uber.org/zap"

type ZapAdapter struct {
    logger *zap.Logger
}

func NewZapAdapter(logger *zap.Logger) *ZapAdapter {
    return &ZapAdapter{logger: logger}
}

func (z *ZapAdapter) Info(ctx context.Context, msg string, fields ...agent.Field) {
    zapFields := make([]zap.Field, len(fields))
    for i, f := range fields {
        zapFields[i] = zap.Any(f.Key, f.Value)
    }
    z.logger.Info(msg, zapFields...)
}

// Implement other methods...

// Usage
zapLogger, _ := zap.NewProduction()
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(NewZapAdapter(zapLogger))
```

### Example: Logrus Adapter

```go
import "github.com/sirupsen/logrus"

type LogrusAdapter struct {
    logger *logrus.Logger
}

func NewLogrusAdapter(logger *logrus.Logger) *LogrusAdapter {
    return &LogrusAdapter{logger: logger}
}

func (l *LogrusAdapter) Info(ctx context.Context, msg string, fields ...agent.Field) {
    logrusFields := make(logrus.Fields)
    for _, f := range fields {
        logrusFields[f.Key] = f.Value
    }
    l.logger.WithFields(logrusFields).Info(msg)
}

// Implement other methods...

// Usage
logrusLogger := logrus.New()
logrusLogger.SetFormatter(&logrus.JSONFormatter{})
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(NewLogrusAdapter(logrusLogger))
```

## Slog Integration

### Text Handler (Development)

```go
handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
    AddSource: true, // Include file:line
})
logger := slog.New(handler)

builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(agent.NewSlogAdapter(logger))
```

Output:
```
time=2024-01-15T10:30:45.123Z level=INFO msg="Ask request completed" duration_ms=767 total_tokens=23
```

### JSON Handler (Production)

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
    AddSource: false,
})
logger := slog.New(handler)

builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(agent.NewSlogAdapter(logger))
```

Output:
```json
{"time":"2024-01-15T10:30:45.123Z","level":"INFO","msg":"Ask request completed","duration_ms":767,"total_tokens":23}
```

### With Context Values

```go
ctx := context.WithValue(context.Background(), "request_id", "12345")

// slog automatically includes context values if configured
builder.Ask(ctx, "Hello")
```

## Production Best Practices

### 1. Use Info Level in Production

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithInfoLogging() // or slog with LevelInfo
```

**Rationale:**
- Debug logs are too verbose
- Info provides sufficient observability
- Lower overhead
- Cleaner logs

### 2. Use JSON Format

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
logger := slog.New(handler)

builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(agent.NewSlogAdapter(logger))
```

**Benefits:**
- Machine-readable
- Easy to parse
- Integrates with log aggregators
- Structured fields preserved

### 3. Send to Log Aggregator

```go
// Example with CloudWatch, Datadog, etc.
handler := NewCloudWatchHandler(config)
logger := slog.New(handler)

builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithLogger(agent.NewSlogAdapter(logger))
```

### 4. Include Request ID

```go
requestID := uuid.New().String()
ctx := context.WithValue(context.Background(), "request_id", requestID)

response, err := builder.Ask(ctx, message)
```

### 5. Monitor Key Metrics

Track these metrics from logs:

- **Latency**: `duration_ms` field
- **Token Usage**: `total_tokens` field
- **Cache Hit Rate**: Count cache hits vs misses
- **Error Rate**: Count ERROR level logs
- **Retry Rate**: Count retry attempts

### 6. Sampling for High Volume

```go
// Log 1% of requests
if rand.Float64() < 0.01 {
    builder = builder.WithInfoLogging()
}
```

### 7. Separate Log Streams

```go
// Errors to stderr
errorHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelError,
})

// Info to stdout
infoHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})

// Combine handlers
multiHandler := NewMultiHandler(infoHandler, errorHandler)
logger := slog.New(multiHandler)
```

## What Gets Logged

### Ask() Method

**DEBUG:**
- Request start (model, message length, features enabled)
- Cache lookup (hit/miss, duration, cache key)
- Tool execution (round, tool name, args, duration)
- RAG retrieval (doc count, duration, method)
- Request execution (duration)

**INFO:**
- Cache hit (cache key, duration)
- Request completed (total duration, request duration, tokens, response length)
- Tool execution completed (rounds, response length)
- RAG retrieval completed (results count, top_k, min_score)

**WARN:**
- Max tool rounds exceeded

**ERROR:**
- Multimodal errors
- Client initialization failures
- RAG retrieval failures
- Request failures (with error and duration)
- Tool execution failures

### Stream() Method

**DEBUG:**
- Stream start (model, message length)
- Stream content finished (content length)
- Stream tool call finished (tool index)
- Starting stream

**INFO:**
- Stream completed (duration, chunks, response length)

**WARN:**
- Stream refusal received

**ERROR:**
- Multimodal errors
- Client initialization failures
- Stream error (error, chunks received, duration)

### Cache Operations

**DEBUG:**
- Cache stats retrieved (hits, misses, size, hit rate)
- No cache configured
- Clearing cache
- Cache cleared successfully
- No cache to clear

**INFO:**
- Clearing cache
- Cache cleared successfully

**ERROR:**
- Failed to clear cache

### Retry Logic

**DEBUG:**
- Applying timeout (timeout duration)
- Retry enabled (max retries)
- Retry attempt (attempt number)
- Retry succeeded (attempt number)
- Error is not retryable (attempt, error)
- Waiting before retry (attempt, delay, error)

**WARN:**
- Max retries reached (attempts, error)

**ERROR:**
- Operation timed out (attempt, timeout)
- Operation timed out during retry (attempt, timeout)
- Context cancelled during retry wait (attempt)

### RAG Retrieval

**DEBUG:**
- RAG retrieval started (query length)
- Using vector store for retrieval (provider, store)
- Using custom retriever
- No RAG documents available
- Using TF-IDF fallback retrieval (total docs)
- Documents chunked (total chunks, chunk size, overlap)
- No RAG documents found (duration)
- Vector search started (collection, query length)
- Executing vector search (top_k, min_score)

**INFO:**
- Custom retriever completed (doc count)
- RAG retrieval completed (results, top_k, min_score)
- Vector search completed (results, collection)

**ERROR:**
- Custom retriever failed (error)
- RAG retrieval failed (error)
- Vector search configuration missing (error)
- Vector search failed (error)

## Performance Considerations

### NoopLogger (Default)

- **Cost**: Zero
- **Overhead**: None
- **Allocations**: 0
- **CPU**: 0%

### StdLogger

- **Cost**: Minimal
- **Overhead**: fmt.Println + string formatting
- **Allocations**: Per log entry
- **CPU**: <1% for typical workloads

### SlogAdapter

- **Cost**: Low
- **Overhead**: slog internal buffering
- **Allocations**: Per log entry
- **CPU**: <2% for typical workloads

### Benchmark Results

```
BenchmarkNoopLogger-10          1000000000    0.25 ns/op    0 B/op    0 allocs/op
BenchmarkStdLogger-10           5000000       250 ns/op     64 B/op   2 allocs/op
BenchmarkSlogJSON-10            3000000       400 ns/op     128 B/op  3 allocs/op
```

**Recommendations:**
- Use NoopLogger (default) when logging not needed
- Use Info level in production (not Debug)
- Consider sampling for very high traffic
- Use async handlers for better performance

## Examples

See `examples/logger_example.go` for complete examples:

1. **Debug Logging** - Detailed tracing for development
2. **Info Logging** - Production-ready logging with caching
3. **Custom Logger** - Implement custom prefix logger
4. **Slog Text Handler** - Human-readable structured logs
5. **Slog JSON Handler** - Machine-readable JSON logs
6. **Streaming with Logging** - Monitor streaming requests
7. **No Logging** - Default zero-overhead mode
8. **RAG with Logging** - Monitor RAG retrieval

Run examples:
```bash
export OPENAI_API_KEY=sk-...
go run examples/logger_example.go
```

## Troubleshooting

### No logs appearing

**Check:**
1. Is logging enabled? (Default is NoopLogger)
2. Is log level appropriate? (Debug shows more than Info)
3. Is output going to correct stream?

**Solution:**
```go
builder := builder.WithDebugLogging() // or WithInfoLogging()
```

### Too many logs

**Solution:**
```go
// Use Info instead of Debug
builder := builder.WithInfoLogging()

// Or use slog with higher level
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelWarn, // Only WARN and ERROR
})
```

### Performance impact

**Solution:**
1. Use Info level (not Debug)
2. Use async logging
3. Sample high-volume requests
4. Disable logging for non-critical paths

### Integration with existing logger

**Solution:**
Implement the `Logger` interface to wrap your existing logger (see Custom Loggers section).

## Summary

✅ **Default**: Zero overhead (NoopLogger)  
✅ **Development**: Use `WithDebugLogging()`  
✅ **Production**: Use `WithInfoLogging()` or slog with JSON  
✅ **Custom**: Implement `Logger` interface  
✅ **Monitoring**: Track duration, tokens, errors  
✅ **Thread-safe**: All loggers are concurrent-safe  

Happy logging! 🎉
