# Project Overview - go-deep-agent

A comprehensive Go library for building LLM-powered applications with support for OpenAI and Ollama.

## 🎯 Current State

**Version:** 0.1.0 (Current API)
**Next Version:** 2.0.0 (Builder API - In Planning)

**Status:** Planning phase for complete rewrite with Builder Pattern (Option B)

## 📁 Project Structure

```
go-deep-agent/
├── agent/
│   ├── config.go (62 lines)      # Configuration & initialization
│   ├── agent.go (140 lines)      # Current implementation
│   └── README.md                 # API documentation
├── examples/
│   ├── ollama_example.go         # Ollama usage examples
│   └── README.md
├── main.go                       # Complete examples
├── go.mod
├── go.sum
│
├── README.md                     # Main documentation
├── QUICK_REFERENCE.md            # Quick reference guide
├── ARCHITECTURE.md               # Design & architecture
├── CHANGELOG.md                  # Version history
├── TODO.md                       # ⭐ Implementation roadmap
├── DESIGN_DECISIONS.md           # ⭐ Design decisions log
└── CONTRIBUTING.md               # ⭐ Contribution guidelines
```

## 📚 Documentation Map

### For Users

1. **[README.md](README.md)** - Start here
   - Installation
   - Quick start
   - Usage examples
   - Features overview

2. **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Fast lookup
   - Common operations
   - Code snippets
   - Patterns

3. **[agent/README.md](agent/README.md)** - Complete reference
   - All API methods
   - Parameters
   - Examples
   - Best practices

### For Contributors

4. **[CONTRIBUTING.md](CONTRIBUTING.md)** - Start here for contributing
   - Setup guide
   - Code style
   - Testing
   - PR process

5. **[TODO.md](TODO.md)** - Implementation roadmap
   - 12 phases planned
   - Task breakdown
   - Timeline estimate
   - Current progress

6. **[DESIGN_DECISIONS.md](DESIGN_DECISIONS.md)** - Design rationale
   - Key decisions explained
   - Options considered
   - Trade-offs
   - Open questions

7. **[ARCHITECTURE.md](ARCHITECTURE.md)** - Deep dive
   - Architecture overview
   - Design patterns
   - Evolution history
   - Future extensibility

8. **[CHANGELOG.md](CHANGELOG.md)** - Version history
   - Release notes
   - Breaking changes
   - Migration guide

## 🚀 Vision: Builder API (v2.0)

### Current API (v0.1.0)
```go
// Verbose, requires nil, imports openai
agent, _ := agent.NewAgent(agent.Config{
    Provider: agent.ProviderOpenAI,
    Model:    "gpt-4o-mini",
    APIKey:   key,
})
result, _ := agent.Chat(ctx, "Hello", nil)
fmt.Println(result.Content)
```

### New Builder API (v2.0 - Planned)
```go
// Fluent, natural, no openai import needed
response := agent.NewOpenAI("gpt-4o-mini", key).
    Ask(ctx, "Hello")
fmt.Println(response)
```

## 🎯 Goals

### User Experience
- ✅ Fluent, chainable API
- ✅ Simple for beginners
- ✅ Powerful for experts
- ✅ No dependency exposure
- ✅ Auto conversation memory

### Technical
- ✅ 100% openai-go feature utilization
- ✅ Clean, maintainable code
- ✅ >80% test coverage
- ✅ Comprehensive documentation
- ✅ Zero breaking changes after v2.0

### Community
- ✅ Open source (MIT)
- ✅ Welcoming to contributors
- ✅ Active maintenance
- ✅ Clear communication

## 📊 Metrics

### Current (v0.1.0)
- **Code:** 202 lines (agent package)
- **Features:** 3/15 openai-go features (20%)
- **Tests:** Basic coverage
- **Docs:** Complete

### Target (v2.0)
- **Code:** ~500-800 lines (agent package)
- **Features:** 15/15 openai-go features (100%)
- **Tests:** >80% coverage
- **Docs:** Complete + examples

## 🗓️ Timeline

### Phase 0: Planning (Current)
- ✅ Design decisions documented
- ✅ TODO created
- ✅ Contributing guidelines
- 📅 Target: Completed

### Phase 1-8: Core Implementation
- Builder pattern
- All openai-go features
- Testing
- 📅 Target: 15-20 days

### Phase 9-10: Documentation & Testing
- Examples
- Documentation
- Quality assurance
- 📅 Target: 5 days

### Phase 11: Advanced Features (Optional)
- RAG support
- Caching
- Multimodal
- 📅 Target: TBD

### Phase 12: Release v2.0
- Final polish
- Community announcement
- 📅 Target: TBD

**Estimated Total:** 20-27 days for v2.0 (without advanced features)

## 🔥 Why This Project?

### Problem
- Most Go LLM libraries are either:
  - Too simple (just API wrappers)
  - Too complex (over-engineered)
  - Provider-specific (not extensible)

### Solution
- **Simple for 90%** - Natural API for common cases
- **Powerful for 10%** - Access to all advanced features
- **Provider agnostic** - Same API for OpenAI, Ollama, others
- **Production ready** - Error handling, retries, observability

### Unique Value
1. **Fluent Builder API** - Most natural Go API for LLMs
2. **100% openai-go** - Maximum feature utilization
3. **No leaky abstraction** - Users don't import openai-go
4. **Auto-memory** - Conversation management built-in
5. **Comprehensive docs** - Every feature explained

## 🎓 Learning Path

### Beginner
1. Read README.md Quick Start
2. Try examples/ollama_example.go
3. Read QUICK_REFERENCE.md
4. Build first app

### Intermediate
1. Read agent/README.md
2. Explore all features
3. Read best practices
4. Contribute examples

### Advanced
1. Read ARCHITECTURE.md
2. Review DESIGN_DECISIONS.md
3. Contribute code
4. Propose new features

## 🤝 How to Get Involved

### Users
- Try the library
- Report bugs
- Request features
- Share projects built with it
- Write tutorials

### Contributors
- Check TODO.md for tasks
- Read CONTRIBUTING.md
- Pick a task
- Submit PR
- Get recognized!

### Maintainers
- Review PRs
- Triage issues
- Guide contributors
- Make decisions
- Release versions

## 📞 Contact & Links

- **GitHub:** https://github.com/taipm/go-deep-agent
- **Issues:** https://github.com/taipm/go-deep-agent/issues
- **Discussions:** https://github.com/taipm/go-deep-agent/discussions
- **License:** MIT

## 🙏 Acknowledgments

- [openai-go](https://github.com/openai/openai-go) - Excellent official Go client
- [Ollama](https://ollama.com) - Amazing local LLM runtime
- Go community - For great tools and libraries

---

**Last Updated:** November 5, 2025
**Maintainer:** taipm
**Status:** Active Development (Planning Phase for v2.0)
