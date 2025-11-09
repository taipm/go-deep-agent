# Hướng dẫn test Chatbot CLI với Ollama

## Chuẩn bị

### 1. Pull model qwen2.5:1.5b (recommended - nhanh & nhẹ)

```bash
# Pull model (lần đầu tiên, ~900MB)
ollama pull qwen2.5:1.5b

# Hoặc pull llama3.2 (lớn hơn, ~2GB)
ollama pull llama3.2

# Kiểm tra models đã có
ollama list
```

### 2. Đảm bảo Ollama đang chạy

```bash
# Terminal 1: Start Ollama
ollama serve

# Terminal 2: Test Ollama
curl http://localhost:11434/api/tags
```

## Chạy Chatbot

### Cách 1: Dùng script test tự động

```bash
cd examples
./test_ollama_chatbot.sh
```

### Cách 2: Chạy trực tiếp

```bash
cd examples
go run chatbot_cli.go

# Khi được hỏi:
# Your choice (1-5): 4            <- Chọn qwen2.5:1.5b
# Enable streaming mode? (y/n): y  <- Bật streaming
# Enable conversation memory? (y/n): y <- Bật memory
```

### Cách 3: Build và chạy

```bash
cd examples
go build chatbot_cli.go
./chatbot_cli
```

## Test scenarios

### Test 1: Simple conversation
```
You: Hello, what is Go?
AI: [streaming response about Go programming language]

You: What are its main features?
AI: [response with memory of previous context]
```

### Test 2: Cache statistics
```
You: What is 2+2?
AI: 4

You: /stats
📊 Cache Statistics:
  Hits:       0
  Misses:     1
  Size:       1 entries

You: What is 2+2?  # Same question
AI: 4

You: /stats
📊 Cache Statistics:
  Hits:       1      <- Cache hit!
  Misses:     1
  Hit Rate:   50.00%
```

### Test 3: Commands
```
You: /help
📚 Available Commands:
  /help   - Show this help message
  /clear  - Clear cache
  /stats  - Show cache statistics
  /exit   - Exit the chatbot

You: /clear
✅ Cache cleared

You: /exit
👋 Goodbye!
```

## Troubleshooting

### Error: model not found
```bash
# Pull the model
ollama pull qwen2.5:1.5b
```

### Error: connection refused
```bash
# Start Ollama service
ollama serve
```

### Slow responses
- qwen2.5:1.5b is fastest (~500ms typical)
- llama3.2 is slower but better quality (~1-2s)
- First request is slower (model loading)
- Subsequent requests are faster (model cached)

## Model comparison

| Model | Size | Speed | Quality | Memory |
|-------|------|-------|---------|--------|
| qwen2.5:1.5b | ~900MB | ⚡⚡⚡⚡ | ⭐⭐⭐ | ~2GB RAM |
| llama3.2 | ~2GB | ⚡⚡⚡ | ⭐⭐⭐⭐ | ~4GB RAM |

**Recommendation**: Start with qwen2.5:1.5b for fast, local testing!
