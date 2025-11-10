# Error Handling Best Practices

Comprehensive guide for error handling in go-deep-agent.

## Table of Contents

- [Quick Start](#quick-start)
- [Error Codes](#error-codes)
- [Debug Mode](#debug-mode)
- [Panic Recovery](#panic-recovery)
- [Error Context](#error-context)
- [Production Patterns](#production-patterns)
- [Common Mistakes](#common-mistakes)

## Quick Start

### Basic Error Handling

```go
agent, err := agent.NewBuilder().
    WithDefaults("YOUR_API_KEY").
    Build()
if err != nil {
    log.Fatalf("Failed to create agent: %v", err)
}

// Execute with error handling
resp, err := agent.Execute(ctx, "Your request")
if err != nil {
    // Check error type/code for appropriate handling
    if agent.IsRateLimitError(err) {
        time.Sleep(time.Second)
        return // Retry
    }
    log.Printf("Error: %v", err)
    return
}
```

### Recommended Setup

```go
// Development
agent, err := agent.NewBuilder().
    WithDefaults(apiKey).
    WithDebug(agent.VerboseDebugConfig()).  // Full logging
    Build()

// Production
agent, err := agent.NewBuilder().
    WithDefaults(apiKey).
    WithDebug(agent.DefaultDebugConfig()).  // Basic logging, secrets redacted
    Build()
```

## Error Codes

### When to Use Error Codes

Error codes help you make programmatic decisions:

```go
resp, err := agent.Execute(ctx, prompt)
if err != nil {
    // Get error code
    code := agent.GetErrorCode(err)
    
    switch code {
    case agent.ErrCodeRateLimitExceeded:
        // Exponential backoff
        time.Sleep(time.Duration(math.Pow(2, retryCount)) * time.Second)
        return retry()
        
    case agent.ErrCodeRequestTimeout:
        // Retry with longer timeout
        return retryWithTimeout(ctx, 30*time.Second)
        
    case agent.ErrCodeAPIKeyMissing:
        // Fatal: cannot recover
        return fmt.Errorf("configuration error: %w", err)
        
    default:
        // Unknown error: log and fail
        log.Printf("Unexpected error [%s]: %v", code, err)
        return err
    }
}
```

### Error Code Categories

**Configuration Errors** (Not Retryable):
- `ErrCodeAPIKeyMissing` - API key not provided
- `ErrCodeInvalidModel` - Invalid model name
- `ErrCodeInvalidConfig` - Invalid configuration

**Transient Errors** (Retryable):
- `ErrCodeRateLimitExceeded` - Rate limit hit (wait and retry)
- `ErrCodeRequestTimeout` - Request timed out (retry)
- `ErrCodeServiceUnavailable` - Service temporarily down

**Request Errors** (Fix Required):
- `ErrCodeInvalidRequest` - Malformed request
- `ErrCodeInvalidJSONSchema` - Invalid JSON schema
- `ErrCodeContextLengthExceeded` - Input too long

**Tool Errors**:
- `ErrCodeToolNotFound` - Tool doesn't exist
- `ErrCodeToolExecutionFailed` - Tool crashed or failed
- `ErrCodeInvalidToolCall` - Malformed tool call from LLM

### Creating Custom Coded Errors

```go
// Create custom error with code and retry behavior
err := agent.NewCodedError(
    baseErr,
    "CUSTOM_ERROR_CODE",
    false, // not retryable
    "Detailed error message",
)

// Or use existing constructors
err = agent.NewRateLimitError(fmt.Errorf("quota exceeded"))
err = agent.NewTimeoutError(fmt.Errorf("deadline exceeded"))
```

### Checking Error Properties

```go
if agent.IsCodedError(err) {
    code := agent.GetErrorCode(err)
    if agent.IsRetryableError(err) {
        // Implement retry logic
    }
}

// Check specific error types
if agent.IsRateLimitError(err) {
    // Rate limit specific handling
}

if agent.IsTimeoutError(err) {
    // Timeout specific handling
}
```

## Debug Mode

### Debug Levels

```go
// Level 1: None (production default)
config := agent.DebugConfig{
    Enabled: false, // No overhead
}

// Level 2: Basic (recommended for production)
config := agent.DefaultDebugConfig()
// Logs: requests, responses, errors
// Secrets: automatically redacted
// Performance: minimal overhead

// Level 3: Verbose (development only)
config := agent.VerboseDebugConfig()
// Logs: everything + token usage + tool execution
// Use for debugging complex issues
```

### Custom Debug Configuration

```go
config := agent.DebugConfig{
    Enabled:         true,
    Level:           agent.DebugLevelBasic,
    RedactSecrets:   true,  // Always true in production
    LogRequests:     true,
    LogResponses:    true,
    LogErrors:       true,
    LogTokenUsage:   false, // Enable for cost monitoring
    LogToolExecution: false, // Enable for tool debugging
    MaxLogLength:    1000,  // Truncate long outputs
}

agent, err := agent.NewBuilder().
    WithDefaults(apiKey).
    WithDebug(config).
    Build()
```

### Secret Redaction

Debug mode automatically redacts:
- OpenAI API keys (`sk-*`, `sk-proj-*`)
- Bearer tokens
- Passwords (field named "password")
- Credentials (field named "credential")
- API keys (field named "api_key", "apikey")

```go
// Before redaction:
// {"api_key": "sk-proj-abc123xyz789"}
// After redaction:
// {"api_key": "sk-proj-ab..."}
```

### Production Debug Setup

```go
// Option 1: Environment-based
var debugConfig agent.DebugConfig
if os.Getenv("ENV") == "production" {
    debugConfig = agent.DefaultDebugConfig()
} else {
    debugConfig = agent.VerboseDebugConfig()
}

// Option 2: Logger-based
logger := log.New(os.Stdout, "[AGENT] ", log.LstdFlags)
debugConfig := agent.DebugConfig{
    Enabled:       true,
    Level:         agent.DebugLevelBasic,
    RedactSecrets: true,
}

agent, err := agent.NewBuilder().
    WithDefaults(apiKey).
    WithLogger(logger).
    WithDebug(debugConfig).
    Build()
```

## Panic Recovery

### Why Panic Recovery Matters

Tools can panic due to:
- Nil pointer dereferences
- Array index out of bounds
- Division by zero
- Third-party library panics

Without recovery, one bad tool crashes your entire application.

### Automatic Tool Panic Recovery

go-deep-agent automatically recovers from tool panics:

```go
agent.WithTool("risky_tool", "May panic", func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // This panic is automatically caught
    panic("something went wrong")
})

resp, err := agent.Execute(ctx, "use risky tool")
if err != nil {
    if agent.IsPanicError(err) {
        // Tool panicked - recovered automatically
        panicValue := agent.GetPanicValue(err)
        stackTrace := agent.GetStackTrace(err)
        
        log.Printf("Tool panic recovered: %v", panicValue)
        log.Printf("Stack trace:\n%s", stackTrace)
    }
}
```

### Manual Panic Recovery

For custom code that might panic:

```go
import "github.com/taipm/go-deep-agent/agent"

func riskyOperation() error {
    var err error
    defer agent.RecoverPanic(&err, "risky operation")
    
    // Code that might panic
    data := processData(input)
    
    return nil
}

// With logging
func riskyOperationWithLogging() error {
    var err error
    defer agent.RecoverPanicWithLogger(&err, "risky operation", logger)
    
    // Code that might panic
    
    return nil
}
```

### Safe Execution Wrappers

```go
// Wrap function that returns value
result, err := agent.SafeExecute("data processing", func() (interface{}, error) {
    // Code that might panic
    return processData(input), nil
})

// Wrap void function
err := agent.SafeExecuteVoid("cleanup", func() error {
    // Code that might panic
    cleanup()
    return nil
})
```

### Production Panic Handling

```go
resp, err := agent.Execute(ctx, prompt)
if err != nil {
    if agent.IsPanicError(err) {
        // Tool panicked - this is serious
        panicValue := agent.GetPanicValue(err)
        stackTrace := agent.GetStackTrace(err)
        
        // Log to monitoring system
        monitoring.RecordPanic(map[string]interface{}{
            "panic_value": panicValue,
            "stack_trace": stackTrace,
            "timestamp":   time.Now(),
        })
        
        // Alert on-call engineer
        alerting.SendAlert("Tool panic detected", stackTrace)
        
        // Return user-friendly error
        return fmt.Errorf("internal error occurred, please try again")
    }
}
```

## Error Context

### Adding Context to Errors

Error context helps debugging by providing operation context:

```go
// Simple context
err = agent.WithSimpleContext(err, "database query")
// Output: "database query: connection failed"

// Rich context with details
err = agent.WithContext(err, "API call", map[string]interface{}{
    "model":         "gpt-4o-mini",
    "retry_attempt": 2,
    "duration_ms":   1500,
})
// Output includes operation + all details
```

### Error Context Chain

Build context as errors propagate:

```go
func processRequest(ctx context.Context, userID string) error {
    data, err := fetchUserData(userID)
    if err != nil {
        return agent.WithContext(err, "fetch user data", map[string]interface{}{
            "user_id": userID,
        })
    }
    
    result, err := callLLM(ctx, data)
    if err != nil {
        return agent.WithContext(err, "LLM call", map[string]interface{}{
            "user_id": userID,
            "data_size": len(data),
        })
    }
    
    return nil
}

// Caller sees full context chain
if err := processRequest(ctx, "user123"); err != nil {
    // Error includes both "fetch user data" and "LLM call" context
    log.Printf("Request failed: %v", err)
}
```

### Error Summarization

Get comprehensive error information:

```go
summary := agent.SummarizeError(err)
if summary != nil {
    log.Printf("Error Type: %s", summary.Type)
    log.Printf("Error Code: %s", summary.Code)
    log.Printf("Retryable: %v", summary.Retryable)
    log.Printf("Message: %s", summary.Message)
    
    // Context details
    for key, value := range summary.Context {
        log.Printf("  %s: %v", key, value)
    }
    
    // Panic information
    if summary.Type == "PanicError" {
        log.Printf("Panic Value: %v", summary.Context["panic_value"])
        log.Printf("Stack Trace:\n%s", summary.Context["stack_trace"])
    }
}
```

### Error Chains

Track multiple errors in a workflow:

```go
chain := agent.NewErrorChain()

// Step 1
if err := validateInput(input); err != nil {
    chain.AddSimple(err, "input validation")
}

// Step 2
if err := checkPermissions(user); err != nil {
    chain.Add(err, "permission check", map[string]interface{}{
        "user_id": user.ID,
        "role":    user.Role,
    })
}

// Step 3
if err := processData(data); err != nil {
    chain.AddSimple(err, "data processing")
}

// Check if any errors occurred
if chain.HasErrors() {
    log.Printf("Workflow failed with %d errors:", chain.Count())
    
    // Get first error
    if first := chain.First(); first != nil {
        log.Printf("First error: %v", first)
    }
    
    // Get all errors
    for i, err := range chain.All() {
        log.Printf("  %d. %v", i+1, err)
    }
    
    // Return chain as error
    return chain
}
```

## Production Patterns

### Complete Error Handling Flow

```go
func executeAgent(ctx context.Context, prompt string) (string, error) {
    // 1. Setup with debug mode
    config := agent.DebugConfig{
        Enabled:          true,
        Level:            agent.DebugLevelBasic,
        RedactSecrets:    true,
        LogRequests:      true,
        LogResponses:     true,
        LogErrors:        true,
        LogTokenUsage:    false,
        LogToolExecution: false,
        MaxLogLength:     2000,
    }
    
    // 2. Create agent with proper configuration
    ag, err := agent.NewBuilder().
        WithDefaults(os.Getenv("OPENAI_API_KEY")).
        WithDebug(config).
        WithLogger(logger).
        Build()
    if err != nil {
        return "", agent.WithContext(err, "agent setup", map[string]interface{}{
            "config_valid": config.Enabled,
        })
    }
    
    // 3. Execute with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    resp, err := ag.Execute(ctx, prompt)
    if err != nil {
        return "", handleAgentError(err, prompt)
    }
    
    return resp.Content, nil
}

func handleAgentError(err error, prompt string) error {
    // Add context
    err = agent.WithContext(err, "agent execution", map[string]interface{}{
        "prompt_length": len(prompt),
        "timestamp":     time.Now().Unix(),
    })
    
    // Get summary for monitoring
    summary := agent.SummarizeError(err)
    if summary != nil {
        // Send to monitoring system
        monitoring.RecordError(map[string]interface{}{
            "error_type": summary.Type,
            "error_code": summary.Code,
            "retryable":  summary.Retryable,
            "context":    summary.Context,
        })
    }
    
    // Check for panic
    if agent.IsPanicError(err) {
        // Critical: alert on-call
        alerting.SendCriticalAlert("Agent panic", map[string]interface{}{
            "panic_value": agent.GetPanicValue(err),
            "stack_trace": agent.GetStackTrace(err),
        })
        return fmt.Errorf("internal error occurred")
    }
    
    // Check error code for retry logic
    if agent.IsRetryableError(err) {
        code := agent.GetErrorCode(err)
        
        switch code {
        case agent.ErrCodeRateLimitExceeded:
            // Specific rate limit handling
            return fmt.Errorf("rate limited, please try again in a few seconds")
            
        case agent.ErrCodeRequestTimeout:
            // Timeout handling
            return fmt.Errorf("request timed out, please try again")
            
        default:
            // Generic retryable error
            return fmt.Errorf("temporary error, please retry")
        }
    }
    
    // Non-retryable error
    return err
}
```

### Retry Logic with Error Codes

```go
func executeWithRetry(ctx context.Context, ag *agent.Agent, prompt string, maxRetries int) (*agent.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        resp, err := ag.Execute(ctx, prompt)
        if err == nil {
            return resp, nil
        }
        
        lastErr = err
        
        // Don't retry on non-retryable errors
        if !agent.IsRetryableError(err) {
            break
        }
        
        // Don't retry on last attempt
        if attempt == maxRetries {
            break
        }
        
        // Calculate backoff based on error code
        var backoff time.Duration
        switch agent.GetErrorCode(err) {
        case agent.ErrCodeRateLimitExceeded:
            // Exponential backoff for rate limits
            backoff = time.Duration(math.Pow(2, float64(attempt))) * time.Second
            
        case agent.ErrCodeRequestTimeout:
            // Linear backoff for timeouts
            backoff = time.Duration(attempt+1) * 2 * time.Second
            
        default:
            // Default backoff
            backoff = time.Second
        }
        
        log.Printf("Attempt %d failed, retrying in %v: %v", attempt+1, backoff, err)
        
        select {
        case <-time.After(backoff):
            continue
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    // Add context before returning
    return nil, agent.WithContext(lastErr, "execute with retry", map[string]interface{}{
        "max_retries": maxRetries,
        "attempts":    maxRetries + 1,
    })
}
```

### Error Logging for Monitoring

```go
func logErrorForMonitoring(err error) {
    summary := agent.SummarizeError(err)
    if summary == nil {
        return
    }
    
    // Build structured log entry
    entry := map[string]interface{}{
        "timestamp":   time.Now().UTC(),
        "error_type":  summary.Type,
        "error_code":  summary.Code,
        "message":     summary.Message,
        "retryable":   summary.Retryable,
    }
    
    // Add context
    for key, value := range summary.Context {
        entry[fmt.Sprintf("ctx_%s", key)] = value
    }
    
    // Different log levels based on error type
    switch summary.Type {
    case "PanicError":
        // Critical: tool panicked
        logger.Error("PANIC RECOVERED", entry)
        
    case "CodedError":
        if summary.Retryable {
            // Warning: transient error
            logger.Warn("Retryable error", entry)
        } else {
            // Error: non-retryable
            logger.Error("Non-retryable error", entry)
        }
        
    default:
        // Info: unexpected error
        logger.Info("Unexpected error", entry)
    }
}
```

## Common Mistakes

### ❌ Mistake 1: Ignoring Error Codes

```go
// BAD
resp, err := agent.Execute(ctx, prompt)
if err != nil {
    return err // Don't know if retryable
}
```

```go
// GOOD
resp, err := agent.Execute(ctx, prompt)
if err != nil {
    if agent.IsRetryableError(err) {
        // Implement retry logic
        return retry(ctx, prompt)
    }
    return err
}
```

### ❌ Mistake 2: Not Using Debug Mode

```go
// BAD - No visibility into what's happening
agent, _ := agent.NewBuilder().
    WithDefaults(apiKey).
    Build()
```

```go
// GOOD - Debug mode for visibility
agent, _ := agent.NewBuilder().
    WithDefaults(apiKey).
    WithDebug(agent.DefaultDebugConfig()). // Secrets auto-redacted
    Build()
```

### ❌ Mistake 3: Not Handling Panics

```go
// BAD - Panic crashes the application
agent.WithTool("tool", "", func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // If this panics, app crashes
    return processUnsafeData(args), nil
})
```

```go
// GOOD - Panic is automatically recovered
agent.WithTool("tool", "", func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // Panic is caught and returned as PanicError
    return processUnsafeData(args), nil
})

// In caller
resp, err := agent.Execute(ctx, prompt)
if err != nil {
    if agent.IsPanicError(err) {
        log.Printf("Tool panicked: %v", agent.GetPanicValue(err))
        // Handle gracefully
    }
}
```

### ❌ Mistake 4: Exposing Secrets in Logs

```go
// BAD - Secrets visible in logs
config := agent.DebugConfig{
    Enabled:       true,
    RedactSecrets: false, // DANGER!
}
```

```go
// GOOD - Secrets automatically redacted
config := agent.DefaultDebugConfig() // RedactSecrets=true
// or
config := agent.DebugConfig{
    Enabled:       true,
    RedactSecrets: true, // Safe
}
```

### ❌ Mistake 5: Not Adding Error Context

```go
// BAD - No context for debugging
func processRequest(userID string) error {
    data, err := fetchData(userID)
    if err != nil {
        return err // Which user? What operation?
    }
    return nil
}
```

```go
// GOOD - Rich context for debugging
func processRequest(userID string) error {
    data, err := fetchData(userID)
    if err != nil {
        return agent.WithContext(err, "fetch user data", map[string]interface{}{
            "user_id":   userID,
            "timestamp": time.Now(),
        })
    }
    return nil
}
```

### ❌ Mistake 6: Wrong Retry Logic

```go
// BAD - Retry all errors including non-retryable ones
for i := 0; i < 3; i++ {
    resp, err := agent.Execute(ctx, prompt)
    if err == nil {
        return resp, nil
    }
    time.Sleep(time.Second) // Retry configuration errors too!
}
```

```go
// GOOD - Only retry retryable errors
for i := 0; i < 3; i++ {
    resp, err := agent.Execute(ctx, prompt)
    if err == nil {
        return resp, nil
    }
    
    if !agent.IsRetryableError(err) {
        return nil, err // Don't retry configuration errors
    }
    
    time.Sleep(time.Second)
}
```

### ❌ Mistake 7: Not Using Error Summarization

```go
// BAD - Manual error inspection
if err != nil {
    // Complicated type assertions
    if codedErr, ok := err.(*agent.CodedError); ok {
        // ...
    } else if panicErr, ok := err.(*agent.PanicError); ok {
        // ...
    }
}
```

```go
// GOOD - Use SummarizeError
if err != nil {
    summary := agent.SummarizeError(err)
    if summary != nil {
        // All information in one place
        log.Printf("Error [%s]: %s (retryable=%v)", 
            summary.Code, summary.Message, summary.Retryable)
        
        if summary.Type == "PanicError" {
            alerting.SendAlert(summary.Context["panic_value"])
        }
    }
}
```

## Summary

**Quick Checklist for Production:**

- ✅ Use `WithDefaults()` for easy setup
- ✅ Enable debug mode with `DefaultDebugConfig()` (secrets auto-redacted)
- ✅ Check error codes with `IsRetryableError()` for retry logic
- ✅ Panics are automatically recovered - check with `IsPanicError()`
- ✅ Add context with `WithContext()` for better debugging
- ✅ Use `SummarizeError()` for comprehensive error analysis
- ✅ Log errors to monitoring system with error codes and context
- ✅ Implement smart retry with exponential backoff for rate limits
- ✅ Alert on panics and non-retryable errors

**See Also:**

- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues and solutions
- [Error Codes Reference](../agent/errors.go) - All error codes
- [Examples](../examples/) - Real-world usage examples
