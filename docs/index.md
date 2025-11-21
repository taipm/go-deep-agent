# go-deep-agent Documentation Index

**Master navigation for AI-assisted development**

## Project Overview

- **Type:** Monolith (Go Backend Library/SDK)
- **Module:** github.com/taipm/go-deep-agent
- **Language:** Go 1.25.2
- **Architecture:** Fluent Builder Pattern
- **Purpose:** Production-ready AI agent library with multi-provider support

## Quick Reference

| Aspect | Details |
|--------|---------|
| **Entry Point** | `main.go`, `agent/builder.go` |
| **Core Package** | `agent/` |
| **Tech Stack** | Go 1.25.2, OpenAI SDK v3.8.1, Gemini SDK v0.20.1, Redis v9, Qdrant |
| **Pattern** | Builder pattern with fluent API |
| **Test Coverage** | 73% (1344+ tests) |
| **Professional Score** | 92/100 |

## 📚 Generated Documentation

### Core Documentation

1. **[Project Overview](project-overview.md)** ⭐
   - Executive summary and value propositions
   - Technology stack overview
   - Feature highlights
   - Repository structure

2. **[API Contracts](api-contracts-main.md)** ⭐
   - Core Agent and Builder APIs
   - Tool system interfaces
   - Memory system APIs
   - Provider adapters (OpenAI, Gemini)
   - Complete method signatures

3. **[Data Models](data-models-main.md)** ⭐
   - Message and conversation structures
   - Tool definitions
   - Memory models (Episodic, Semantic, Working)
   - RAG document structures
   - Cache and vector store interfaces

4. **[Source Tree Analysis](source-tree-analysis.md)** ⭐
   - Complete directory structure with annotations
   - Critical directories explained
   - File organization patterns
   - Integration points

5. **[Development Guide](development-guide-main.md)** ⭐
   - Environment setup
   - Local development workflow
   - Testing strategy
   - Build and deployment
   - Contribution guidelines

## 📖 Existing Documentation

### Root Level Documentation

#### Primary Docs
- **[README.md](../README.md)** - Main project documentation
  - Features overview
  - Quick start guide
  - Installation instructions
  - Basic examples

- **[ARCHITECTURE.md](../ARCHITECTURE.md)** - System architecture
  - Design principles
  - Component relationships
  - Architecture patterns

- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Contribution guidelines
  - How to contribute
  - PR process
  - Code style
  - Testing requirements

- **[CHANGELOG.md](../CHANGELOG.md)** - Version history
  - Release notes
  - Breaking changes
  - New features by version

#### Analysis & Assessment
- **[CODEBASE_ANALYSIS.md](../CODEBASE_ANALYSIS.md)** - Codebase analysis
- **[LIBRARY_ASSESSMENT_REPORT.md](../LIBRARY_ASSESSMENT_REPORT.md)** - Professional assessment (92/100)
- **[AGENT_USAGE_STRATEGY.md](../AGENT_USAGE_STRATEGY.md)** - Usage strategies

#### Technical Guides
- **[WHY_REACT.md](WHY_REACT.md)** - ReAct pattern rationale
- **[RATE_LIMITING_GUIDE.md](RATE_LIMITING_GUIDE.md)** - Rate limiting implementation
- **[FEWSHOT_GUIDE.md](FEWSHOT_GUIDE.md)** - Few-shot learning
- **[JSON_SCHEMA.md](JSON_SCHEMA.md)** - JSON schema usage
- **[ERROR_HANDLING_BEST_PRACTICES.md](ERROR_HANDLING_BEST_PRACTICES.md)** - Error handling
- **[MEMORY_MIGRATION.md](MEMORY_MIGRATION.md)** - Memory system migration

#### Design & Planning
- **[COMPARISON.md](COMPARISON.md)** - vs OpenAI SDK comparison
- **[LLM_PROVIDERS_INTEGRATION_DESIGN.md](../LLM_PROVIDERS_INTEGRATION_DESIGN.md)** - Multi-provider design
- **[ADVANCED_RAG_PLAN.md](ADVANCED_RAG_PLAN.md)** - RAG implementation plan
- **[YAML_CONFIG_ANALYSIS.md](YAML_CONFIG_ANALYSIS.md)** - YAML configuration

#### BMAD Method Documentation

- **[BMAD_METHOD_WORKFLOW.md](BMAD_METHOD_WORKFLOW.md)** ⭐ **NEW** - BMAD Method workflow and principles
- **[BMAD_IMPLEMENTATION_GUIDE.md](BMAD_IMPLEMENTATION_GUIDE.md)** ⭐ **NEW** - Detailed BMAD implementation guide
- **[BMAD_PROJECT_RETROSPECTIVE.md](BMAD_PROJECT_RETROSPECTIVE.md)** ⭐ **NEW** - Project retrospective and lessons learned

### Release Notes

Located in `releases/`:
- [v0.7.1](releases/v0.7.1.md)
- [v0.7.0](releases/v0.7.0.md)
- [v0.3.0](releases/v0.3.0.md)

### Module-Specific Documentation

- **[agent/README.md](../agent/README.md)** - Core agent package
- **[agent/adapters/README.md](../agent/adapters/README.md)** - LLM adapters

## 🚀 Getting Started

### For New Users

1. Read [README.md](../README.md) for overview
2. Check [Project Overview](project-overview.md) for context
3. Review [API Contracts](api-contracts-main.md) for available APIs
4. Try [examples/](../examples/) for hands-on learning

### For Contributors

1. Read [CONTRIBUTING.md](../CONTRIBUTING.md)
2. Follow [Development Guide](development-guide-main.md)
3. Review [ARCHITECTURE.md](../ARCHITECTURE.md)
4. Check [Source Tree](source-tree-analysis.md) for codebase structure

### For AI-Assisted Development (Brownfield PRD)

**Start here for planning new features:**

1. **[Project Overview](project-overview.md)** - Understand the project
2. **[API Contracts](api-contracts-main.md)** - Available APIs
3. **[Data Models](data-models-main.md)** - Data structures
4. **[Architecture](../ARCHITECTURE.md)** - System design
5. **[Source Tree](source-tree-analysis.md)** - Codebase navigation

**Key context for AI:**
- This is a **Go library/SDK** for building AI agents
- Uses **Builder pattern** for fluent API
- Supports **multi-provider** (OpenAI, Gemini)
- Has **comprehensive test coverage** (73%, 1344+ tests)
- **Production-ready** with retry, timeout, caching, memory

## 📁 Repository Structure

```
go-deep-agent/
├── agent/              # Core library (Builder, Agent, Tools, Memory, Adapters)
├── examples/           # 20+ usage examples (ReAct, Planning, Tools, etc.)
├── docs/               # Documentation (you are here)
├── personas/           # Pre-built agent personas
├── config/             # Configuration templates
├── configs/            # Example configurations
├── main.go             # Entry point
└── *.md                # Root documentation files
```

### Critical Directories

| Directory | Purpose | Key Files |
|-----------|---------|-----------|
| `agent/` | Core library | `builder.go`, `agent.go`, `builder_*.go` |
| `agent/adapters/` | LLM providers | `openai_adapter.go`, `gemini_adapter.go` |
| `agent/memory/` | Memory system | `episodic.go`, `semantic.go`, `working.go` |
| `agent/tools/` | Built-in tools | `filesystem.go`, `http.go`, `math.go`, etc. |
| `examples/` | Examples | `react_*/`, `planner_*/`, `rate_limit_*/`, etc. |

## 🎯 Common Use Cases

### Building a Simple Agent

See: [README.md](../README.md) Quick Start section

### Building with Memory

See: [MEMORY_MIGRATION.md](MEMORY_MIGRATION.md)

### Using Tools

See: [examples/react_advanced/](../examples/react_advanced/)

### Implementing RAG

See: [ADVANCED_RAG_PLAN.md](ADVANCED_RAG_PLAN.md)

### Rate Limiting

See: [RATE_LIMITING_GUIDE.md](RATE_LIMITING_GUIDE.md)

### Streaming Responses

See: [examples/react_streaming/](../examples/react_streaming/)

## 🔧 Development Workflows

### Run Tests

```bash
go test ./...
```

### Run Examples

```bash
cd examples/react_simple && go run main.go
```

### Build

```bash
go build ./...
```

### Format Code

```bash
go fmt ./...
```

See [Development Guide](development-guide-main.md) for complete details.

## 🌟 Key Features

- ✅ Fluent Builder API
- ✅ Multi-Provider (OpenAI, Gemini)
- ✅ 3-Tier Memory System
- ✅ Tool Calling & Auto-Execution
- ✅ RAG with Vector Search
- ✅ Streaming Support
- ✅ Batch Processing
- ✅ Caching (Memory, Redis)
- ✅ Error Recovery
- ✅ Production-Ready

## 📊 Quality Metrics

- **Tests:** 1344+ passing
- **Coverage:** 73%
- **Professional Assessment:** 92/100
- **Go Report Card:** A+
- **Documentation:** Comprehensive

## 🔗 External Resources

- [OpenAI API Documentation](https://platform.openai.com/docs/)
- [Google Gemini Documentation](https://ai.google.dev/docs)
- [Go Language Documentation](https://golang.org/doc/)

---

**Documentation Index Generated:** 2025-11-14
**Scan Level:** Deep
**Project Type:** Backend Library (Go SDK)
**Repository Type:** Monolith

**For AI-Assisted Development:** This index provides complete context for understanding the go-deep-agent codebase. All generated documentation is optimized for AI comprehension and brownfield feature planning.
