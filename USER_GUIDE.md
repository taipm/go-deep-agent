# Go Deep Agent - User Guide

**🎯 Library Status: PRODUCTION READY with Gemini V3, OpenAI, and Ollama support**

---

## 🚀 Quick Start for New Developers

### Installation
```bash
go get github.com/taipm/go-deep-agent
```

### Basic Usage - Choose Your Provider

#### 1. OpenAI (Recommended for Production)
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Create OpenAI adapter
    openai, err := agent.NewOpenAI("gpt-4o-mini", "your-api-key-here")
    if err != nil {
        log.Fatal(err)
    }
    defer openai.Close()

    // Simple chat
    response, err := openai.Complete(context.Background(), &agent.CompletionRequest{
        Messages: []agent.Message{
            {Role: "user", Content: "Hello! Explain Go programming in 100 words"},
        },
        Temperature: 0.7,
        MaxTokens:   200,
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response.Content)
}
```

#### 2. Ollama (Local & Free)
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Create Ollama adapter (no API key needed)
    ollama, err := agent.NewOllama("llama3.1:8b")
    if err != nil {
        log.Fatal(err)
    }
    defer ollama.Close()

    response, err := ollama.Complete(context.Background(), &agent.CompletionRequest{
        Messages: []agent.Message{
            {Role: "user", Content: "What is machine learning?"},
        },
    })

    fmt.Println(response.Content)
}
```

#### 3. Gemini V3 (Latest Google AI)
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Get API key from environment
    apiKey := os.Getenv("GEMINI_API_KEY")

    // Create Gemini V3 adapter
    gemini, err := agent.NewGeminiV3Adapter(apiKey, "gemini-1.5-pro-latest")
    if err != nil {
        log.Fatal(err)
    }
    defer gemini.Close()

    response, err := gemini.Complete(context.Background(), &agent.CompletionRequest{
        Messages: []agent.Message{
            {Role: "user", Content: "Explain quantum computing simply"},
        },
    })

    fmt.Println(response.Content)
}
```

---

## 🔄 For Existing Developers - Migration Guide

### From Previous Versions

#### What Changed:
- ✅ **Gemini V3**: Completely rewritten with production-grade quality
- ✅ **Streaming**: Now available for ALL providers (OpenAI, Ollama, Gemini)
- ✅ **Tool Calling**: Fixed all critical issues with schema conversion
- ✅ **Error Handling**: Enterprise-grade error categorization
- ✅ **MultiProvider**: Enhanced with load balancing and health checks

#### Breaking Changes:
```go
// ❌ OLD WAY (no longer works)
// geminiAdapter := agent.NewGeminiAdapter(apiKey, model)

// ✅ NEW WAY (production ready)
geminiAdapter, err := agent.NewGeminiV3Adapter(apiKey, "gemini-1.5-pro-latest")
```

#### Your Existing Code Still Works:
```go
// ✅ OpenAI - unchanged
openai, _ := agent.NewOpenAI("gpt-4o-mini", apiKey)

// ✅ Ollama - unchanged
ollama, _ := agent.NewOllama("llama3.1:8b")

// ✅ MultiProvider - enhanced but compatible
multiprovider, _ := agent.NewMultiProvider([]agent.ProviderConfig{
    {Type: "openai", APIKey: openaiKey, Model: "gpt-4o-mini"},
    {Type: "ollama", Model: "llama3.1:8b"},
    // ✅ Gemini can now be added!
    {Type: "gemini", APIKey: geminiKey, Model: "gemini-1.5-pro-latest"},
})
```

---

## 🌊 Streaming - NOW AVAILABLE FOR ALL PROVIDERS

### Simple Streaming (Same API for All Providers)
```go
func streamExample() {
    // Works with OpenAI, Ollama, and Gemini!
    adapter, _ := agent.NewOpenAI("gpt-4o-mini", apiKey)

    response, err := adapter.Stream(context.Background(), &agent.CompletionRequest{
        Messages: []agent.Message{
            {Role: "user", Content: "Write a short story about a robot"},
        },
    }, func(chunk string) {
        fmt.Print(chunk) // Real-time output
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n\nComplete response: %s\n", response.Content)
}
```

### Streaming with Different Providers
```go
// OpenAI Streaming
openai, _ := agent.NewOpenAI("gpt-4o-mini", openaiKey)
openai.Stream(ctx, request, onChunk)

// Ollama Streaming
ollama, _ := agent.NewOllama("llama3.1:8b")
ollama.Stream(ctx, request, onChunk)

// Gemini V3 Streaming
gemini, _ := agent.NewGeminiV3Adapter(geminiKey, "gemini-1.5-pro-latest")
gemini.Stream(ctx, request, onChunk)
```

---

## 🔧 Tool Calling - Production Ready

### Define Tools (Same for All Providers)
```go
func setupTools() []*agent.Tool {
    calculator := agent.NewTool("calculator", "Perform mathematical calculations").
        AddParameter("expression", "string", "Math expression to evaluate", true)

    weather := agent.NewTool("get_weather", "Get weather information").
        AddParameter("location", "string", "City name", true).
        AddParameter("units", "string", "Temperature units (celsius/fahrenheit)", false)

    return []*agent.Tool{calculator, weather}
}
```

### Use Tools with Any Provider
```go
func toolExample() {
    // Works with OpenAI, Ollama, and Gemini
    adapter, _ := agent.NewGeminiV3Adapter(geminiKey, "gemini-1.5-pro-latest")

    response, err := adapter.Complete(context.Background(), &agent.CompletionRequest{
        Messages: []agent.Message{
            {Role: "user", Content: "What's 25 * 4?"},
        },
        Tools: setupTools(),
    })

    if err != nil {
        log.Fatal(err)
    }

    // Handle tool calls
    for _, toolCall := range response.ToolCalls {
        fmt.Printf("Tool called: %s with args: %s\n", toolCall.Name, toolCall.Arguments)
    }
}
```

---

## 🏗️ MultiProvider - Enterprise Features

### Setup Multiple Providers
```go
func setupMultiProvider() *agent.MultiProvider {
    config := []agent.ProviderConfig{
        {
            Type:     "openai",
            Name:     "openai-primary",
            APIKey:   os.Getenv("OPENAI_API_KEY"),
            Model:    "gpt-4o-mini",
            Priority: 1, // Highest priority
        },
        {
            Type:  "gemini",
            Name:  "gemini-backup",
            APIKey: os.Getenv("GEMINI_API_KEY"),
            Model: "gemini-1.5-pro-latest",
            Priority: 2,
        },
        {
            Type:     "ollama",
            Name:     "ollama-local",
            Model:    "llama3.1:8b",
            Priority: 3, // Fallback
        },
    }

    multiprovider, err := agent.NewMultiProvider(config)
    if err != nil {
        log.Fatal(err)
    }

    return multiprovider
}
```

### Use with Automatic Load Balancing
```go
func multiProviderExample() {
    mp := setupMultiProvider()

    // Automatically routes to best available provider
    response, err := mp.Complete(context.Background(), &agent.CompletionRequest{
        Messages: []agent.Message{
            {Role: "user", Content: "Explain microservices"},
        },
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response from provider: %s\n", response.Content)
    fmt.Printf("Tokens used: %d\n", response.Usage.TotalTokens)
}
```

---

## 📊 Monitoring & Health Checks

### Provider Health Monitoring
```go
func healthCheckExample() {
    mp := setupMultiProvider()

    // Check health of all providers
    health := mp.GetProviderHealth()
    for provider, status := range health {
        fmt.Printf("%s: %v (response time: %v)\n",
            provider, status.Healthy, status.ResponseTime)
    }

    // Get metrics
    metrics := mp.GetMetrics()
    fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
    fmt.Printf("Success rate: %.2f%%\n", metrics.SuccessRate*100)
}
```

---

## 🛠️ Configuration Options

### Environment Variables
```bash
# Required for OpenAI
export OPENAI_API_KEY="your-openai-key"

# Required for Gemini
export GEMINI_API_KEY="your-gemini-key"

# Optional: Ollama configuration
export OLLAMA_BASE_URL="http://localhost:11434"  # Default
```

### Advanced Configuration
```go
func advancedConfig() {
    // OpenAI with custom settings
    openai, _ := agent.NewOpenAI("gpt-4o-mini", apiKey,
        agent.WithOpenAIBaseURL("https://api.openai.com/v1"),
        agent.WithOpenAITimeout(30*time.Second),
        agent.WithOpenAIRetryAttempts(3),
    )

    // Ollama with custom endpoint
    ollama, _ := agent.NewOllama("llama3.1:8b",
        agent.WithOllamaBaseURL("http://localhost:11434"),
        agent.WithOllamaTimeout(60*time.Second),
    )

    // MultiProvider with load balancing
    mp, _ := agent.NewMultiProvider([]agent.ProviderConfig{
        {Type: "openai", APIKey: openaiKey, Model: "gpt-4o-mini"},
        {Type: "gemini", APIKey: geminiKey, Model: "gemini-1.5-pro-latest"},
    },
        agent.WithLoadBalancing(agent.LoadBalancingStrategyRoundRobin),
        agent.WithHealthChecks(30*time.Second),
        agent.WithMetrics(true),
    )
}
```

---

## 🔍 Error Handling

### Production-Grade Error Handling
```go
func errorHandlingExample() {
    gemini, _ := agent.NewGeminiV3Adapter(apiKey, "gemini-1.5-pro-latest")

    response, err := gemini.Complete(context.Background(), request)
    if err != nil {
        // Errors are properly categorized
        switch {
        case strings.Contains(err.Error(), "authentication"):
            log.Fatal("API key is invalid")
        case strings.Contains(err.Error(), "quota exceeded"):
            log.Fatal("Rate limit reached, try again later")
        case strings.Contains(err.Error(), "content policy"):
            log.Fatal("Request violates content policy")
        default:
            log.Printf("Unexpected error: %v", err)
        }
        return
    }

    fmt.Println(response.Content)
}
```

---

## 📚 Best Practices

### 1. Always Handle Context Cancellation
```go
func bestPractice1() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    response, err := adapter.Complete(ctx, request)
    if err != nil {
        // Handle context timeout
        if errors.Is(err, context.DeadlineExceeded) {
            log.Println("Request timed out")
            return
        }
        log.Fatal(err)
    }
}
```

### 2. Use Streaming for Long Responses
```go
func bestPractice2() {
    response, err := adapter.Stream(ctx, request, func(chunk string) {
        // Show progress to user
        fmt.Print(chunk)
    })
}
```

### 3. Implement Retry Logic
```go
func bestPractice3() {
    var response *agent.CompletionResponse
    var err error

    for attempt := 0; attempt < 3; attempt++ {
        response, err = adapter.Complete(ctx, request)
        if err == nil {
            break // Success
        }

        // Wait before retry
        time.Sleep(time.Duration(attempt+1) * time.Second)
    }

    if err != nil {
        log.Fatal("All attempts failed")
    }
}
```

### 4. Monitor Token Usage
```go
func bestPractice4() {
    response, _ := adapter.Complete(ctx, request)

    // Track costs
    fmt.Printf("Prompt tokens: %d\n", response.Usage.PromptTokens)
    fmt.Printf("Completion tokens: %d\n", response.Usage.CompletionTokens)
    fmt.Printf("Total tokens: %d\n", response.Usage.TotalTokens)
}
```

---

## 🚀 Production Deployment

### Docker Example
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

# Environment variables
ENV OPENAI_API_KEY=""
ENV GEMINI_API_KEY=""

CMD ["./main"]
```

### Kubernetes Configuration
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-deep-agent-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-deep-agent-app
  template:
    metadata:
      labels:
        app: go-deep-agent-app
    spec:
      containers:
      - name: app
        image: your-app:latest
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: openai
        - name: GEMINI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: gemini
        resources:
          limits:
            memory: "512Mi"
            cpu: "500m"
```

---

## 🆘 Support & Troubleshooting

### Common Issues

#### 1. API Key Issues
```bash
# Test API key
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models

# For Gemini
curl -H "x-goog-api-key: $GEMINI_API_KEY" \
     https://generativelanguage.googleapis.com/v1beta/models
```

#### 2. Ollama Connection Issues
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Pull a model
ollama pull llama3.1:8b
```

#### 3. Performance Issues
```go
// Use connection pooling for high throughput
mp, _ := agent.NewMultiProvider(config,
    agent.WithMaxConnections(10),
    agent.WithConnectionTimeout(5*time.Second),
)
```

### Getting Help
- 📖 [Documentation](./docs/)
- 🐛 [Issues](https://github.com/taipm/go-deep-agent/issues)
- 💬 [Discussions](https://github.com/taipm/go-deep-agent/discussions)
- 📧 [Email Support](mailto:support@example.com)

---

## 🎯 Summary

**For New Developers:**
- ✅ Choose OpenAI for production reliability
- ✅ Use Ollama for local development (free)
- ✅ Try Gemini V3 for latest Google AI features
- ✅ All providers support streaming and tool calling

**For Existing Developers:**
- ✅ Your code still works (except old Gemini adapter)
- ✅ Add Gemini V3 to your MultiProvider setup
- ✅ Use streaming for better user experience
- ✅ Monitor provider health with new metrics

**Production Ready Features:**
- ✅ Enterprise-grade error handling
- ✅ Load balancing and failover
- ✅ Health monitoring and metrics
- ✅ Circuit breaker patterns
- ✅ Token usage tracking

**🚀 The library is production-ready for all your AI needs!**