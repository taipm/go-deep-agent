# Project Overview - go-deep-agent

## Executive Summary

**go-deep-agent** is a production-ready AI agent library for Go, designed to simplify building LLM-powered applications with a modern fluent builder API. It achieves 60-80% code reduction compared to raw OpenAI SDK usage while providing enterprise-grade features.

**Professional Assessment:** 92/100 (See [LIBRARY_ASSESSMENT_REPORT.md](../LIBRARY_ASSESSMENT_REPORT.md))

## Project Metadata

| Property | Value |
|----------|-------|
| **Module** | github.com/taipm/go-deep-agent |
| **Language** | Go 1.25.2 |
| **Type** | Backend Library/SDK |
| **Architecture** | Monolithic (single cohesive library) |
| **License** | MIT |
| **Test Coverage** | 73% |
| **Tests** | 1344+ passing |
| **Primary SDK** | openai-go v3.8.1 |

## Purpose

Provide Go developers with a **production-ready**, **type-safe**, and **developer-friendly** SDK for building AI agents and LLM applications. The library abstracts the complexity of raw LLM APIs while maintaining full control and flexibility.

## Key Value Propositions

1. **60-80% Less Code** - Fluent builder API eliminates boilerplate
2. **Production-Ready** - Built-in retry, timeout, error handling, caching
3. **Type-Safe** - Full Go type safety with compile-time checks
4. **Multi-Provider** - OpenAI, Google Gemini, custom endpoints
5. **Feature-Rich** - Memory, RAG, tools, streaming, batch processing, vector search
6. **Well-Tested** - 1344+ tests with 73% coverage
7. **Developer Experience** - Readable, chainable API with intelligent defaults

## Technology Stack

### Core Technologies

| Category | Technology | Version | Purpose |
|----------|-----------|---------|---------|
| **Language** | Go | 1.25.2 | Primary language |
| **OpenAI SDK** | openai-go | v3.8.1 | OpenAI API integration |
| **Gemini SDK** | google/generative-ai-go | v0.20.1 | Google Gemini integration |
| **Caching** | go-redis | v9.16.0 | Redis caching support |
| **Testing** | miniredis | v2.35.0 | In-memory Redis for tests |
| **Config** | godotenv | v1.5.1 | Environment variables |
| **YAML** | yaml.v3 | v3.0.1 | Configuration parsing |
| **Rate Limiting** | golang.org/x/time | v0.14.0 | Request throttling |
| **Math** | gonum | v0.16.0 | Numerical operations |
| **Testing** | testify | v1.11.1 | Test assertions |

### Architecture Pattern

**Builder Pattern with Fluent API**

```go
// Example: Chainable configuration
response := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithDefaults(). // Memory + Retry + Timeout
    Ask(ctx, "Hello!")
```

## Core Features

### 1. Fluent Builder API
- Method chaining for natural code flow
- Intelligent defaults with `WithDefaults()`
- Over 50 configuration methods

### 2. Memory System (3-Tier Hierarchy)
- **Working Memory** - Active conversation context
- **Episodic Memory** - Sequential event history
- **Semantic Memory** - Long-term knowledge storage
- Automatic importance scoring and retrieval

### 3. Tool Calling & Execution
- Type-safe tool definitions
- Automatic execution with `autoExecute`
- Parallel tool orchestration
- Built-in tools: FileSystem, HTTP, DateTime, Math

### 4. RAG (Retrieval-Augmented Generation)
- Document chunking and retrieval
- Vector embeddings (OpenAI, Ollama)
- Vector databases (ChromaDB, Qdrant)
- Semantic search with similarity scoring

### 5. Production Features
- Smart retry with exponential backoff
- Request timeouts
- Response caching (memory, Redis)
- Batch processing with concurrency control
- Comprehensive error handling

### 6. Streaming Support
- Real-time response streaming
- Content, tool call, and refusal callbacks
- Accumulator pattern for chunk handling

### 7. Multimodal Support
- Vision model support (GPT-4 Vision)
- Images via URL, file path, or base64
- Detail level control (auto, low, high)

### 8. Multi-Provider Support
- OpenAI (GPT models)
- Google Gemini
- Custom base URLs (Ollama, local models)

## Repository Structure

```
go-deep-agent/
├── agent/           # Core library implementation
│   ├── adapters/   # LLM provider adapters
│   ├── memory/     # Memory subsystems
│   └── tools/      # Built-in tools
├── examples/       # 20+ usage examples
├── docs/           # Comprehensive documentation
├── personas/       # Pre-built agent personas
└── config/         # Configuration templates
```

## Project Classification

- **Repository Type:** Monolith
- **Project Type:** Backend Library (Go SDK)
- **Target Users:** Go developers building LLM applications
- **Use Cases:**
  - AI agents and assistants
  - Chatbots with memory
  - RAG applications
  - Tool-calling agents
  - Batch LLM processing
  - Research and reasoning systems

## Development Status

- **Current Version:** v0.11.0+ (active development)
- **Stability:** Production-ready
- **Breaking Changes:** Rare (follows semantic versioning)
- **Maintenance:** Active (see [CHANGELOG.md](../CHANGELOG.md))

## Quick Reference

### Installation

```bash
go get github.com/taipm/go-deep-agent
```

### Minimal Example

```go
import "github.com/taipm/go-deep-agent/agent"

response := agent.NewOpenAI("gpt-4o-mini", apiKey).
    Ask(ctx, "Explain quantum computing")
```

### With All Features

```go
response := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithDefaults().              // Memory + Retry + Timeout
    WithTemperature(0.7).
    WithMaxTokens(500).
    WithTools(myTools...).
    WithAutoExecute().
    Stream(ctx, "Complex question")
```

## Community & Support

- **Documentation:** [docs/](.)
- **Examples:** [examples/](../examples/)
- **Issues:** [GitHub Issues](https://github.com/taipm/go-deep-agent/issues)
- **Contributing:** [CONTRIBUTING.md](../CONTRIBUTING.md)

## Related Documentation

- [Architecture](../ARCHITECTURE.md) - System architecture
- [API Contracts](api-contracts-main.md) - API reference
- [Data Models](data-models-main.md) - Data structures
- [Source Tree](source-tree-analysis.md) - Codebase structure
- [Development Guide](development-guide-main.md) - Setup and development
- [Comparison](../docs/COMPARISON.md) - vs raw OpenAI SDK
- [Assessment](../LIBRARY_ASSESSMENT_REPORT.md) - Quality assessment

---

**Generated:** 2025-11-14
**Scan Level:** Deep
**Documentation Version:** 1.0
