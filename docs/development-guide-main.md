# Development Guide - go-deep-agent

## Prerequisites

### Required

- **Go:** 1.25.2 or higher
- **Git:** For version control
- **Make:** (Optional) For build automation

### Optional

- **Docker:** For Redis/Qdrant containers
- **Redis:** For caching and memory persistence
- **Qdrant:** For vector database features

## Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/taipm/go-deep-agent.git
cd go-deep-agent
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Environment Variables

Create `.env` file in project root:

```bash
# OpenAI
OPENAI_API_KEY=your_openai_api_key

# Google Gemini (optional)
GEMINI_API_KEY=your_gemini_api_key

# Redis (optional, for caching)
REDIS_URL=localhost:6379

# Qdrant (optional, for vector search)
QDRANT_URL=http://localhost:6333
```

### 4. Start Optional Services

**Redis (for caching features):**

```bash
docker run -d -p 6379:6379 redis:latest
```

**Qdrant (for vector search features):**

```bash
docker run -d -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

## Local Development

### Run Tests

**All tests:**

```bash
go test ./...
```

**With coverage:**

```bash
go test -cover ./...
```

**Specific package:**

```bash
go test ./agent
go test ./agent/memory
go test ./agent/tools
```

**With verbose output:**

```bash
go test -v ./...
```

**Generate coverage report:**

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Examples

Navigate to any example directory and run:

```bash
cd examples/react_simple
go run main.go
```

**Popular examples:**

```bash
# Simple ReAct pattern
cd examples/react_simple && go run main.go

# Advanced ReAct with tools
cd examples/react_advanced && go run main.go

# Streaming responses
cd examples/react_streaming && go run main.go

# Rate limiting
cd examples/rate_limit_basic && go run main.go

# Planning and reasoning
cd examples/planner_adaptive && go run main.go

# Math teacher agent
cd examples/math_teacher && go run main.go
```

### Build

**Build library:**

```bash
go build ./...
```

**Build specific example:**

```bash
cd examples/react_simple
go build -o react_simple main.go
./react_simple
```

### Linting & Formatting

**Format code:**

```bash
go fmt ./...
```

**Run linter (if golangci-lint installed):**

```bash
golangci-lint run
```

**Vet code:**

```bash
go vet ./...
```

## Project Structure for Development

### Core Package: `agent/`

**Main files to understand:**

1. `builder.go` - Entry point, builder struct
2. `agent.go` - Core Agent type and Chat() method
3. `builder_*.go` - Feature-specific extensions
4. `config.go` - Configuration structures

**Key subsystems:**

- `adapters/` - LLM provider integrations
- `memory/` - Memory management
- `tools/` - Built-in tools

### Testing Strategy

**Unit tests:** Alongside implementation (`*_test.go`)

**Test utilities:**
- `testify/assert` - Assertions
- `miniredis` - Mock Redis server
- Table-driven tests for comprehensive coverage

**Example test:**

```go
func TestBuilderBasic(t *testing.T) {
    ctx := context.Background()
    b := NewOpenAI("gpt-4o-mini", apiKey)

    result := b.Ask(ctx, "Hello")
    assert.NotEmpty(t, result)
}
```

## Common Development Tasks

### Adding a New Feature

1. Create feature branch: `git checkout -b feature/my-feature`
2. Implement in appropriate `builder_*.go` file
3. Add tests in `*_test.go`
4. Update examples if user-facing
5. Update docs
6. Submit PR

### Adding a New Tool

1. Create tool in `agent/tools/`
2. Implement tool interface
3. Add tests
4. Update tools documentation
5. Create example in `examples/`

### Adding a New Adapter

1. Create adapter in `agent/adapters/`
2. Implement adapter interface
3. Add provider-specific logic
4. Add tests
5. Update examples

### Debugging

**Enable logging:**

```go
import "log/slog"

agent.NewOpenAI(model, key).
    WithLogger(slog.Default()).
    Ask(ctx, "test")
```

**Debug tool execution:**

```go
agent.NewOpenAI(model, key).
    WithTools(myTools...).
    WithAutoExecute().
    // Tools will log execution automatically
    Ask(ctx, "use the calculator tool")
```

## Testing Best Practices

### Unit Testing

- Test each builder method independently
- Use table-driven tests for multiple scenarios
- Mock external dependencies (LLM APIs, Redis, etc.)

### Integration Testing

- Use real APIs with test keys (if available)
- Test end-to-end workflows
- Verify streaming, tool calling, RAG

### Test Coverage Goals

- Core library: 80%+ coverage
- Critical paths: 100% coverage
- Examples: Basic smoke tests

## Code Style Guidelines

### Follow Go Conventions

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable names
- Add comments for exported functions

### Builder Pattern Conventions

```go
// ✅ Good: Chainable, returns *Builder
func (b *Builder) WithTemperature(temp float64) *Builder {
    b.temperature = &temp
    return b
}

// ❌ Bad: Not chainable
func (b *Builder) SetTemperature(temp float64) {
    b.temperature = &temp
}
```

### Error Handling

```go
// ✅ Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to execute tool: %w", err)
}

// ❌ Bad: Lost context
if err != nil {
    return err
}
```

## Performance Considerations

### Caching

- Use Redis caching for repeated queries
- Set appropriate TTL values
- Clear cache when needed

### Batch Processing

- Use `Batch()` for multiple requests
- Configure `batchSize` for concurrency
- Monitor rate limits

### Memory Management

- Use `maxHistory` to limit conversation size
- Clear working memory when not needed
- Use semantic memory for long-term storage

## Troubleshooting

### Common Issues

**1. Import errors**

```bash
go mod tidy
go mod vendor
```

**2. API key not found**

```bash
# Check .env file exists and has correct keys
cat .env
```

**3. Redis connection failed**

```bash
# Check Redis is running
redis-cli ping
# Should return: PONG
```

**4. Tests failing**

```bash
# Clear test cache
go clean -testcache
go test ./...
```

## CI/CD

### GitHub Actions

- Automated testing on push/PR
- Coverage reporting
- Linting checks

See [.github/workflows/](.github/workflows/) for pipeline definitions.

## Release Process

1. Update `CHANGELOG.md`
2. Tag version: `git tag v0.x.x`
3. Push tags: `git push --tags`
4. GitHub Actions creates release
5. Update documentation

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for:
- Code of conduct
- PR process
- Issue reporting
- Feature requests

## Resources

### Documentation

- [README.md](../README.md) - Main documentation
- [ARCHITECTURE.md](../ARCHITECTURE.md) - System design
- [API Contracts](api-contracts-main.md) - API reference
- [Examples](../examples/) - Code examples

### External Resources

- [OpenAI API Docs](https://platform.openai.com/docs/)
- [Google Gemini Docs](https://ai.google.dev/docs)
- [Go Documentation](https://golang.org/doc/)

---

**Generated:** 2025-11-14
**Scan Level:** Deep
**Target Audience:** Contributors and developers
