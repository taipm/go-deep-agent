# Examples

Thư mục này chứa các ví dụ sử dụng go-deep-agent.

## Chạy examples

### Ollama Example

```bash
# Đảm bảo Ollama đang chạy
ollama serve

# Chạy example
cd examples
go run ollama_example.go
```

### OpenAI Example (từ main.go)

```bash
# Set API key
export OPENAI_API_KEY=your-api-key

# Chạy tất cả examples
go run main.go
```

## Các ví dụ có sẵn

1. **Chat cơ bản** - Simple question-answer
2. **Streaming** - Real-time streaming responses
3. **Conversation History** - Multi-turn conversations
4. **Tool Calling** - Function calling với tools
5. **Structured Outputs** - JSON schema validation
6. **Multimodal (Vision)** - Image analysis với GPT-4 Vision

### Multimodal Example

Phân tích ảnh với GPT-4 Vision:

```bash
# Set API key
export OPENAI_API_KEY=your-api-key

# Chạy multimodal examples
go run examples/builder_multimodal.go
```

Các tính năng:

- Mô tả ảnh từ URL
- So sánh nhiều ảnh
- Control detail levels (low/high)
- OCR - extract text từ ảnh
- Phân tích local image files
- Chart/graph analysis
- Multi-turn conversation với ảnh
