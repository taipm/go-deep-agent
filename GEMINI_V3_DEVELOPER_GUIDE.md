# Gemini V3 Developer Guide - Production Ready! 🚀

## 🎯 Overview

Gemini V3 đã sẵn sàng production với enterprise-grade tool calling support. Sử dụng Google GenAI SDK v1.36.0 với tất cả critical fixes đã được áp dụng.

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Basic Usage](#basic-usage)
- [Tool Calling](#tool-calling)
- [MultiProvider Integration](#multiprovider-integration)
- [Streaming Support](#streaming-support)
- [Error Handling](#error-handling)
- [Advanced Examples](#advanced-examples)

## ⚡ Quick Start

### Cách 1: Sử dụng trực tiếp (Recommended)

```go
package main

import (
    "context"
    "fmt"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Khởi tạo Gemini V3 adapter
    gemini, err := agent.NewGeminiV3Adapter("your-gemini-api-key", "gemini-1.5-pro-latest")
    if err != nil {
        panic(err)
    }
    defer gemini.Close()

    // Chat đơn giản
    response, err := gemini.Complete(context.Background(), &agent.CompletionRequest{
        Model: "gemini-1.5-pro-latest",
        Messages: []agent.Message{
            {Role: "user", Content: "Hello! How are you?"},
        },
        Temperature: 0.7,
        MaxTokens:    1000,
    })

    if err != nil {
        panic(err)
    }

    fmt.Printf("Response: %s\n", response.Content)
}
```

### Cách 2: Sử dụng Builder Pattern (Easy & Popular)

```go
package main

import (
    "context"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Sử dụng builder với Gemini V3
    gemini := agent.NewGeminiV3("your-gemini-api-key", "gemini-1.5-pro-latest")

    response, err := gemini.
        WithSystem("You are a helpful assistant").
        WithTemperature(0.7).
        Ask(context.Background(), "What is the meaning of life?")

    if err != nil {
        panic(err)
    }

    fmt.Printf("Response: %s\n", response)
}
```

## 🛠️ Tool Calling

### Calculator Tool Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    gemini := agent.NewGeminiV3("your-api-key", "gemini-1.5-pro-latest")

    // Tạo tool calculator
    calculatorTool := agent.NewTool("calculator", "Perform mathematical operations").
        AddParameter("a", "number", "First number", true).
        AddParameter("b", "number", "Second number", true).
        AddParameter("operation", "string", "Operation (add, subtract, multiply, divide)", true).
        AddEnum("operation", []string{"add", "subtract", "multiply", "divide"}).
        WithHandler(func(args string) (string, error) {
            // Logic calculator ở đây
            // Parse JSON args và thực hiện calculation
            return "Result: 42", nil
        })

    response, err := gemini.
        WithTools(calculatorTool).
        WithAutoExecute(true). // Tự động execute tools
        Ask(context.Background(), "Calculate 15 + 27")

    if err != nil {
        panic(err)
    }

    fmt.Printf("Response: %s\n", response)
}
```

## 🔄 Streaming Support

### Simple Streaming (Easy to Use)

```go
package main

import (
    "context"
    "fmt"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    gemini := agent.NewGeminiV3("your-api-key", "gemini-1.5-pro-latest")

    // Streaming response
    response, err := gemini.Stream(context.Background(), &agent.CompletionRequest{
        Model: "gemini-1.5-pro-latest",
        Messages: []agent.Message{
            {Role: "user", Content: "Tell me a short story about AI"},
        },
    }, func(chunk string) {
        // Callback nhận từng chunk
        fmt.Printf("Chunk: %s", chunk)
    })

    if err != nil {
        panic(err)
    }

    fmt.Printf("\nComplete response: %s\n", response.Content)
}
```

### Builder Pattern with Streaming

```go
// Sử dụng streaming với builder pattern
response, err := gemini.
    WithSystem("You are a storyteller").
    WithTemperature(0.8).
    Stream(context.Background(), "Write a poem about technology", func(chunk string) {
        fmt.Print(chunk)
    })
```

## 🏢 MultiProvider Integration

### Production Setup với Fallback

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // MultiProvider với Gemini V3 + OpenAI
    config := &agent.MultiProviderConfig{
        Providers: []agent.ProviderConfig{
            {
                Name:     "gemini-primary",
                Type:     "gemini-v3",
                Model:    "gemini-1.5-pro-latest",
                APIKey:   "your-gemini-key",
                Timeout:  30 * time.Second,
                Weight:   0.7, // 70% traffic
            },
            {
                Name:     "openai-fallback",
                Type:     "openai",
                Model:    "gpt-4o-mini",
                APIKey:   "your-openai-key",
                Timeout:  30 * time.Second,
                Weight:   0.3, // 30% traffic
            },
        },
        SelectionStrategy: agent.StrategyWeightedRoundRobin,
        FallbackStrategy:  agent.FallbackStrategyRetryWithBackoff,
        HealthCheckInterval: 5 * time.Minute,
    }

    multiprovider, err := agent.NewMultiProvider(config)
    if err != nil {
        panic(err)
    }
    defer multiprovider.Shutdown(context.Background())

    // Sử dụng với auto-fallback
    response, err := multiprovider.Ask(context.Background(), "What's the weather like?")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Response: %s\n", response)
}
```

## 🛡️ Error Handling

### Production-Grade Error Handling

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    gemini, err := agent.NewGeminiV3Adapter("your-api-key", "gemini-1.5-pro-latest")
    if err != nil {
        log.Fatalf("Failed to create Gemini adapter: %v", err)
    }

    for i := 0; i < 3; i++ {
        response, err := gemini.Complete(context.Background(), &agent.CompletionRequest{
            Model: "gemini-1.5-pro-latest",
            Messages: []agent.Message{
                {Role: "user", Content: fmt.Sprintf("Attempt %d: Tell me a joke", i+1)},
            },
        })

        if err != nil {
            // Gemini adapter provides categorized errors
            switch {
            case strings.Contains(err.Error(), "authentication"):
                log.Fatalf("Authentication failed: %v", err)
            case strings.Contains(err.Error(), "quota exceeded"):
                log.Printf("Quota exceeded, retrying...: %v", err)
                time.Sleep(time.Second * 5)
                continue
            case strings.Contains(err.Error(), "content policy"):
                log.Printf("Content policy violation: %v", err)
                break
            default:
                log.Printf("Other error: %v", err)
                continue
            }
        } else {
            fmt.Printf("Success: %s\n", response.Content)
            break
        }
    }
}
```

## 🔧 Advanced Examples

### Custom Tool with Validation

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    gemini := agent.NewGeminiV3("your-api-key", "gemini-1.5-pro-latest")

    // Custom validation tool
    validationTool := agent.NewTool("validate_email", "Validate email address").
        AddParameter("email", "string", "Email address to validate", true).
        AddParameter("strict", "boolean", "Strict validation", false).
        WithHandler(func(args string) (string, error) {
            var params struct {
                Email  string `json:"email"`
                Strict bool   `json:"strict"`
            }

            if err := json.Unmarshal([]byte(args), &params); err != nil {
                return "", fmt.Errorf("invalid arguments: %w", err)
            }

            // Email validation logic
            if !strings.Contains(params.Email, "@") {
                return "", fmt.Errorf("invalid email format")
            }

            return fmt.Sprintf("Email %s is valid", params.Email), nil
        })

    response, err := gemini.
        WithTools(validationTool).
        WithAutoExecute(true).
        Ask(context.Background(), "Validate this email: user@example.com")

    if err != nil {
        panic(err)
    }

    fmt.Printf("Response: %s\n", response)
}
```

### System Instructions with Tools

```go
// System instruction + tools
response, err := gemini.
    WithSystem("You are a helpful math tutor. Always show your work step by step.").
    WithTools(calculatorTool).
    WithAutoExecute(true).
    Ask(context.Background(), "What is the result of 15 * 4 + 23?")
```

## 📊 Best Practices

### ✅ Do's

1. **Always validate API keys** before using
2. **Set reasonable timeouts** for production
3. **Use AutoExecute** for simple tool usage
4. **Handle errors appropriately** based on type
5. **Use Builder pattern** for readability
6. **Test with different models** and parameters

### ❌ Don'ts

1. **Don't use empty API keys** - will cause runtime errors
2. **Don't ignore errors** - they provide valuable information
3. **Don't set extremely high MaxTokens** - can be costly
4. **Don't use streaming without context cancellation**
5. **Don't forget to close adapters** when done

## 🔍 Troubleshooting

### Common Issues and Solutions

#### 1. "API key authentication error"
```go
// Solution: Check API key and permissions
gemini, err := agent.NewGeminiV3Adapter("valid-api-key", "gemini-1.5-pro-latest")
if err != nil {
    log.Printf("API key error: %v", err)
    return
}
```

#### 2. "Quota exceeded"
```go
// Solution: Implement retry with exponential backoff
response, err := gemini.Complete(ctx, req)
if err != nil && strings.Contains(err.Error(), "quota exceeded") {
    // Implement retry logic
    time.Sleep(time.Second * 5)
    response, err = gemini.Complete(ctx, req)
}
```

#### 3. "Content policy violation"
```go
// Solution: Handle policy violations gracefully
if err != nil && strings.Contains(err.Error(), "content policy") {
    // Use different prompt or model
    log.Printf("Content policy violation, trying different approach")
    // Try again with different content
}
```

## 🚀 Performance Tips

### For High-Performance Applications

1. **Reuse adapter instances** instead of creating new ones
2. **Use streaming** for long responses
3. **Set appropriate MaxTokens** based on expected response length
4. **Use MultiProvider** for load balancing and redundancy
5. **Monitor token usage** to control costs

### Example: High-Performance Setup

```go
// Reuse adapter
var gemini *agent.GeminiV3Adapter

func init() {
    var err error
    gemini, err = agent.NewGeminiV3Adapter("your-api-key", "gemini-1.5-pro-latest")
    if err != nil {
        log.Fatal(err)
    }
}

func handleRequest(message string) (string, error) {
    // Reuse existing adapter
    return gemini.Ask(context.Background(), message)
}
```

## 📝 Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Initialize Gemini V3
    gemini, err := agent.NewGeminiV3Adapter("your-api-key", "gemini-1.5-pro-latest")
    if err != nil {
        log.Fatalf("Failed to create Gemini adapter: %v", err)
    }
    defer gemini.Close()

    // Create tools
    calculator := agent.NewTool("calculator", "Perform math operations").
        AddParameter("expression", "string", "Math expression", true).
        WithHandler(func(args string) (string, error) {
            // Simple calculator implementation
            return fmt.Sprintf("Result: %s", args), nil
        })

    // Complete example with streaming, tools, and error handling
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    response, err := gemini.
        WithSystem("You are a helpful assistant with math capabilities.").
        WithTools(calculator).
        WithAutoExecute(true).
        Stream(ctx, &agent.CompletionRequest{
            Model: "gemini-1.5-pro-latest",
            Messages: []agent.Message{
                {Role: "user", Content: "Calculate the area of a circle with radius 5"},
            },
        }, func(chunk string) {
            fmt.Printf("📝 %s", chunk)
        })

    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    fmt.Printf("\n✅ Success: %s\n", response.Content)
}
```

---

**🎉 Chúc mừng! Bạn đã sẵn sàng sử dụng Gemini V3 production-grade với go-deep-agent!**

**Cần hỗ trợ?** Check documentation hoặc tạo issue trên GitHub repository.