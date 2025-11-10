# Error Handling Usage Examples

This document provides code examples for error handling patterns. For complete guidance, see [ERROR_HANDLING_BEST_PRACTICES.md](../docs/ERROR_HANDLING_BEST_PRACTICES.md).

## Pattern 1: Error Code Checking

Use error codes to make programmatic decisions:

```go
package main

import (
    "context"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    ctx := context.Background()
    b := agent.NewOpenAI("gpt-4o-mini", apiKey).WithDefaults()
    
    resp, err := b.Ask(ctx, "Your prompt")
    if err != nil {
        // Get error code
        code := agent.GetErrorCode(err)
        
        // Make decision based on code
        switch code {
        case agent.ErrCodeRateLimitExceeded:
            // Wait and retry
            time.Sleep(2 * time.Second)
            return retry()
            
        case agent.ErrCodeRequestTimeout:
            // Retry with longer timeout
            return retryWithTimeout(ctx)
            
        case agent.ErrCodeAPIKeyMissing:
            // Fatal: cannot proceed
            log.Fatal("API key configuration error")
            
        default:
            log.Printf("Unexpected error [%s]: %v", code, err)
        }
        return
    }
    
    fmt.Println(resp)
}
```

## Pattern 2: Debug Mode

Enable debug mode for visibility (secrets are auto-redacted):

```go
// Development: Verbose logging
config := agent.VerboseDebugConfig()

b := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDefaults().
    WithDebug(config)

resp, err := b.Ask(ctx, "Your prompt")
```

```go
// Production: Basic logging
config := agent.DefaultDebugConfig() // Secrets redacted

b := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDefaults().
    WithDebug(config)

resp, err := b.Ask(ctx, "Your prompt")
```

## Pattern 3: Panic Recovery

Panic recovery is automatic for tools. Check for panics:

```go
resp, err := b.Ask(ctx, "Use a tool")
if err != nil {
    if agent.IsPanicError(err) {
        // Tool panicked - recovered automatically
        panicValue := agent.GetPanicValue(err)
        stackTrace := agent.GetStackTrace(err)
        
        log.Printf("Tool panic: %v", panicValue)
        log.Printf("Stack trace:\n%s", stackTrace)
        
        // Alert monitoring system
        alertOnPanic(panicValue, stackTrace)
    }
}
```

## Pattern 4: Error Context

Add context to errors as they bubble up:

```go
func processUser(userID string) error {
    data, err := fetchUserData(userID)
    if err != nil {
        // Add context
        return agent.WithContext(err, "fetch user data", map[string]interface{}{
            "user_id": userID,
            "timestamp": time.Now().Unix(),
        })
    }
    
    result, err := callLLM(ctx, data)
    if err != nil {
        // Add more context
        return agent.WithContext(err, "LLM call", map[string]interface{}{
            "user_id": userID,
            "data_size": len(data),
        })
    }
    
    return nil
}

// In caller
if err := processUser("user123"); err != nil {
    // Summarize for logging
    summary := agent.SummarizeError(err)
    if summary != nil {
        log.Printf("Type: %s, Code: %s", summary.Type, summary.Code)
        log.Printf("Message: %s", summary.Message)
        log.Printf("Retryable: %v", summary.Retryable)
        
        // Context contains all details
        for key, value := range summary.Context {
            log.Printf("  %s: %v", key, value)
        }
    }
}
```

## Pattern 5: Error Chains

Track multiple errors in complex workflows:

```go
func processWorkflow(userID string) error {
    chain := agent.NewErrorChain()
    
    // Step 1
    if err := validateInput(input); err != nil {
        chain.AddSimple(err, "input validation")
        return chain // Stop on validation error
    }
    
    // Step 2
    if err := checkPermissions(userID); err != nil {
        chain.Add(err, "permission check", map[string]interface{}{
            "user_id": userID,
        })
        // Continue even if this fails
    }
    
    // Step 3
    if err := processData(data); err != nil {
        chain.AddSimple(err, "data processing")
    }
    
    // Check results
    if chain.HasErrors() {
        log.Printf("Workflow failed with %d errors", chain.Count())
        
        // Get first error (primary cause)
        first := chain.First()
        
        // Get all errors
        for i, err := range chain.All() {
            log.Printf("  %d. %v", i+1, err)
        }
        
        return chain
    }
    
    return nil
}
```

## Pattern 6: Production Error Handler

Complete production error handling:

```go
func handleProductionError(err error, operation string) {
    // Add context
    err = agent.WithContext(err, operation, map[string]interface{}{
        "timestamp": time.Now().Unix(),
    })
    
    // Get summary
    summary := agent.SummarizeError(err)
    if summary == nil {
        log.Printf("Error: %v", err)
        return
    }
    
    // Log structured data
    logEntry := map[string]interface{}{
        "error_type": summary.Type,
        "error_code": summary.Code,
        "message":    summary.Message,
        "retryable":  summary.Retryable,
        "context":    summary.Context,
    }
    
    // Send to monitoring system
    monitoring.RecordError(logEntry)
    
    // Check for panic (critical)
    if summary.Type == "PanicError" {
        alerting.SendCriticalAlert("Panic recovered", logEntry)
    }
    
    // Check for non-retryable errors
    if !summary.Retryable {
        code := summary.Code
        if code == agent.ErrCodeAPIKeyMissing || 
           code == agent.ErrCodeInvalidConfig {
            alerting.SendAlert("Configuration error", logEntry)
        }
    }
}
```

## Pattern 7: Smart Retry Logic

Retry with backoff based on error code:

```go
func executeWithRetry(ctx context.Context, b *agent.Builder, prompt string, maxRetries int) (string, error) {
    var lastErr error
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        resp, err := b.Ask(ctx, prompt)
        if err == nil {
            return resp, nil
        }
        
        lastErr = err
        
        // Check if error has a code
        code := agent.GetErrorCode(err)
        if code == "" {
            // No code - unknown error, don't retry
            break
        }
        
        // Check if specific error is retryable
        isRetryable := false
        var backoff time.Duration
        
        switch code {
        case agent.ErrCodeRateLimitExceeded:
            isRetryable = true
            // Exponential backoff: 1s, 2s, 4s, 8s...
            backoff = time.Duration(math.Pow(2, float64(attempt))) * time.Second
            
        case agent.ErrCodeRequestTimeout:
            isRetryable = true
            // Linear backoff: 2s, 4s, 6s...
            backoff = time.Duration(attempt+1) * 2 * time.Second
            
        case agent.ErrCodeServiceUnavailable:
            isRetryable = true
            // Fixed backoff
            backoff = 5 * time.Second
            
        default:
            // Other errors not retryable
            isRetryable = false
        }
        
        if !isRetryable {
            break
        }
        
        if attempt == maxRetries {
            break
        }
        
        log.Printf("Retrying in %v (attempt %d/%d)...", backoff, attempt+1, maxRetries)
        
        select {
        case <-time.After(backoff):
            continue
        case <-ctx.Done():
            return "", ctx.Err()
        }
    }
    
    return "", agent.WithContext(lastErr, "retry logic", map[string]interface{}{
        "max_retries": maxRetries,
        "attempts":    maxRetries + 1,
        "final_code":  agent.GetErrorCode(lastErr),
    })
}
```

## Summary

Key patterns:

1. **Error Codes** - Use `GetErrorCode()` for programmatic decisions
2. **Debug Mode** - Use `DefaultDebugConfig()` or `VerboseDebugConfig()`
3. **Panic Recovery** - Check with `IsPanicError()`, get details with `GetPanicValue()` and `GetStackTrace()`
4. **Error Context** - Use `WithContext()` to add operation context
5. **Error Chains** - Use `NewErrorChain()` for multi-step workflows
6. **Error Summary** - Use `SummarizeError()` for comprehensive error info
7. **Smart Retry** - Check error code and retry appropriately

See [ERROR_HANDLING_BEST_PRACTICES.md](../docs/ERROR_HANDLING_BEST_PRACTICES.md) for complete documentation.
