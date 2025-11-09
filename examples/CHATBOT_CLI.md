# 🤖 Interactive Chatbot CLI

A fully interactive command-line chatbot powered by go-deep-agent, featuring conversation memory, real-time streaming, and support for multiple AI providers.

## ✨ Features

- **🎙️ Interactive Chat Loop** - Natural conversation with real-time input/output
- **🤖 Multiple AI Providers**:
  - OpenAI GPT-4o-mini (fast, efficient)
  - OpenAI GPT-4o (most capable)
  - OpenAI GPT-4-turbo (advanced reasoning)
  - Ollama (local, private)
- **💬 Conversation Memory** - Optional memory to remember context (max 20 messages)
- **⚡ Streaming Mode** - Real-time response streaming with chunk-by-chunk display
- **📊 Cache Statistics** - Monitor cache hits, misses, and performance
- **⏱️ Response Time Tracking** - See how fast each response is generated
- **🎮 Built-in Commands** - Help, stats, clear, exit

## 🚀 Quick Start

### Prerequisites

```bash
# For OpenAI models
export OPENAI_API_KEY="your-api-key-here"

# For Ollama (optional, if using local models)
ollama serve
```

### Run

```bash
# From go-deep-agent root directory
go run examples/chatbot_cli.go

# Or build and run
cd examples
go build chatbot_cli.go
./chatbot_cli
```

## 📖 Usage Guide

### Startup Flow

When you run the chatbot, you'll be prompted to configure it:

```
═══════════════════════════════════════════════════════════
   🤖 GO-DEEP-AGENT CHATBOT CLI
   Interactive AI Assistant powered by go-deep-agent
═══════════════════════════════════════════════════════════

Select AI Provider:
1. OpenAI (GPT-4o-mini) - Fast, efficient
2. OpenAI (GPT-4o) - Most capable
3. OpenAI (GPT-4-turbo) - Advanced reasoning
4. Ollama (llama3.2) - Local, private

Your choice (1-4): 1
Enable streaming mode? (y/n): y
Enable conversation memory? (y/n): y

✅ Conversation memory enabled (max 20 messages)
✅ Streaming mode enabled

────────────────────────────────────────────────────────────
🤖 Chatbot ready! Type your message and press Enter.
💡 Commands: /help, /stats, /clear, /exit
────────────────────────────────────────────────────────────
```

### Chat Examples

**Simple Question:**
```
You: What is Go programming language?
AI:  Go is a statically typed, compiled programming language designed 
at Google. It's known for its simplicity, efficiency, and excellent 
support for concurrent programming through goroutines and channels.

⏱️  Response time: 1.23s
```

**With Conversation Memory:**
```
You: My name is Alice
AI:  Nice to meet you, Alice! How can I help you today?

⏱️  Response time: 0.89s

You: What's my name?
AI:  Your name is Alice!

⏱️  Response time: 0.67s
```

**Code Generation:**
```
You: Write a simple HTTP server in Go
AI:  Here's a simple HTTP server in Go:

```go
package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}
```

⏱️  Response time: 1.45s
```

### Built-in Commands

| Command | Description | Example Output |
|---------|-------------|----------------|
| `/help` | Show available commands | Lists all commands |
| `/stats` | Display cache statistics | Hits, misses, hit rate |
| `/clear` | Clear cache | Confirmation message |
| `/exit` | Exit chatbot | Also: `/quit`, `/q` |

**Examples:**

```
You: /help

📚 Available Commands:
  /help   - Show this help message
  /stats  - Show cache statistics
  /clear  - Clear cache
  /exit   - Exit the chatbot

You: /stats

📊 Cache Statistics:
  Hits:       3
  Misses:     2
  Size:       2 entries
  Evictions:  0
  Hit Rate:   60.00%

You: /clear
✅ Cache cleared

You: /exit

👋 Goodbye! Thanks for chatting!
```

## ⚙️ Configuration Options

### Model Selection

| Model | Provider | Speed | Capability | Cost | Use Case |
|-------|----------|-------|------------|------|----------|
| gpt-4o-mini | OpenAI | ⚡⚡⚡ | ⭐⭐⭐ | $ | Daily tasks, quick questions |
| gpt-4o | OpenAI | ⚡⚡ | ⭐⭐⭐⭐⭐ | $$$ | Complex reasoning, coding |
| gpt-4-turbo | OpenAI | ⚡⚡ | ⭐⭐⭐⭐ | $$ | Balanced performance |
| llama3.2 | Ollama | ⚡⚡⚡ | ⭐⭐⭐ | Free | Local, privacy-focused |

### Streaming Mode

**Enabled (Recommended):**
- Real-time response streaming
- See AI "thinking" as it generates
- Better user experience for long responses

**Disabled:**
- Wait for complete response
- Cleaner output
- Better for short, factual queries

### Conversation Memory

**Enabled:**
- AI remembers last 20 messages
- Natural multi-turn conversations
- Context-aware responses

**Disabled:**
- Each message is independent
- No memory overhead
- Better for one-off questions

## 🎯 Use Cases

### 1. Daily Assistant
```
You: Summarize this email: [paste email]
You: Draft a professional reply
You: Make it more formal
```

### 2. Coding Help
```
You: How do I read a file in Go?
You: Show me error handling too
You: What about reading line by line?
```

### 3. Learning & Research
```
You: Explain quantum computing in simple terms
You: How does it differ from classical computing?
You: What are real-world applications?
```

### 4. Brainstorming
```
You: I need ideas for a mobile app
You: Focus on productivity apps
You: What features should it have?
```

## 🔧 Advanced Customization

You can modify `chatbot_cli.go` to add more features:

### Add Custom System Prompt

```go
chatbot = chatbot.
    WithSystem("You are a Go expert. Focus on idiomatic Go patterns.").
    WithTemperature(0.7)
```

### Enable Caching

```go
chatbot = chatbot.
    WithMemoryCache(5 * time.Minute). // Cache responses for 5 minutes
    WithMemory()
```

### Add More Models

```go
fmt.Println("5. Claude (Anthropic)")
fmt.Println("6. Gemini (Google)")

// In selectModel():
case "5":
    return "claude-3-5-sonnet-20241022", "claude"
```

### Custom Temperature

```go
temp := askFloat("Set temperature (0.0-2.0): ")
chatbot = chatbot.WithTemperature(temp)
```

## 📊 Performance Tips

### Optimize for Speed
- Use `gpt-4o-mini` (fastest OpenAI model)
- Enable memory caching
- Disable streaming for short queries
- Use Ollama for local/offline access

### Optimize for Quality
- Use `gpt-4o` for complex tasks
- Enable conversation memory
- Set temperature to 0.3 for factual responses
- Set temperature to 0.9 for creative responses

### Optimize for Cost
- Enable caching (repeated questions are free)
- Use memory to avoid re-stating context
- Use `gpt-4o-mini` instead of `gpt-4o` when possible

## 🐛 Troubleshooting

### Error: OPENAI_API_KEY not set
```bash
export OPENAI_API_KEY="sk-your-key-here"
```

### Error: Connection refused (Ollama)
```bash
# Start Ollama server
ollama serve

# In another terminal
ollama pull llama3.2
```

### Slow responses
- Try `gpt-4o-mini` instead of `gpt-4o`
- Check your internet connection
- Enable caching for repeated questions

### Memory issues
- Reduce `WithMaxHistory(20)` to lower number
- Disable memory for simple Q&A

## 🎓 Learning Resources

- [Go Deep Agent Documentation](../README.md)
- [Builder API Guide](../BUILDER_API.md)
- [More Examples](./README.md)

## 🤝 Contributing

Found a bug or have a feature idea? Open an issue or submit a PR!

## 📝 License

Same as go-deep-agent (see root LICENSE file)

---

**Built with ❤️ using [go-deep-agent](https://github.com/taipm/go-deep-agent)**
