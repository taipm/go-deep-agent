# Source Tree Analysis - go-deep-agent

## Project Structure

```
go-deep-agent/
├── agent/                    # 🎯 Core agent implementation
│   ├── adapters/            # LLM provider adapters (OpenAI, Gemini)
│   ├── memory/              # Memory subsystems (episodic, semantic, working, system)
│   ├── tools/               # Built-in tools (datetime, filesystem, http, math, logger)
│   ├── agent.go             # Core Agent type and Chat() method
│   ├── builder.go           # Fluent API builder pattern [ENTRY POINT]
│   ├── builder_*.go         # Builder extensions (config, memory, logging, execution, etc.)
│   ├── config.go            # Configuration structures
│   ├── adapter.go           # Adapter interfaces
│   ├── batch.go             # Batch processing
│   ├── vector_store.go      # Vector database integration
│   ├── embedding.go         # Embedding providers
│   └── qdrant.go            # Qdrant vector store implementation
│
├── examples/                # 📚 Usage examples and demos
│   ├── config_basic/        # Basic configuration examples
│   ├── fewshot_basic/       # Few-shot learning examples
│   ├── persona_basic/       # Persona/system prompt examples
│   ├── planner_*/           # Planning and reasoning examples
│   │   ├── planner_adaptive/   # Adaptive planning
│   │   ├── planner_basic/      # Basic planning
│   │   └── planner_parallel/   # Parallel planning
│   ├── react_*/             # ReAct pattern examples
│   │   ├── react_simple/       # Simple ReAct
│   │   ├── react_advanced/     # Advanced ReAct with tools
│   │   ├── react_math/         # Math reasoning
│   │   ├── react_research/     # Research tasks
│   │   ├── react_native/       # Native tool integration
│   │   ├── react_streaming/    # Streaming responses
│   │   └── react_error_recovery/ # Error handling
│   ├── rate_limit_*/        # Rate limiting examples
│   │   ├── rate_limit_basic/   # Basic rate limiting
│   │   └── rate_limit_advanced/ # Advanced rate limiting
│   ├── tool_choice_demo/    # Tool choice control
│   ├── debug_enhanced/      # Enhanced debugging
│   ├── math_teacher/        # Math teaching agent
│   └── full_config/         # Comprehensive configuration
│
├── docs/                    # 📖 Documentation
│   ├── api/                 # API documentation
│   ├── guides/              # User guides
│   ├── development/         # Development documentation
│   ├── releases/            # Release notes (v0.3.0, v0.7.0, v0.7.1)
│   ├── assessments/         # Project assessments
│   ├── archive/             # Archived documentation
│   │   ├── assessments/
│   │   ├── evaluations/
│   │   ├── planning/
│   │   └── summaries/
│   ├── sprint-artifacts/    # Agile sprint artifacts
│   ├── api-contracts-main.md    # API documentation (generated)
│   ├── data-models-main.md      # Data models (generated)
│   └── project-scan-report.json # Scan state file
│
├── personas/                # 🎭 Pre-built agent personas
│   └── (persona definitions)
│
├── config/                  # ⚙️ Configuration files
│   └── (configuration templates)
│
├── configs/                 # ⚙️ Additional configurations
│   └── (example configurations)
│
├── .github/                 # GitHub configuration
│   ├── workflows/           # CI/CD pipelines
│   └── chatmodes/           # GitHub chat modes
│
├── .claude/                 # Claude Code configuration
│   └── commands/
│       └── bmad/            # BMAD workflow commands
│
├── .gemini/                 # Gemini configuration
│   └── commands/
│
├── .opencode/               # OpenCode configuration
│   ├── agent/
│   └── command/
│
├── .vscode/                 # VS Code settings
│
├── main.go                  # 🚀 Main entry point (if executable)
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
│
└── *.md                     # Root documentation files
    ├── README.md            # Project README
    ├── ARCHITECTURE.md      # Architecture documentation
    ├── CONTRIBUTING.md      # Contribution guidelines
    ├── CHANGELOG.md         # Change log
    ├── CODEBASE_ANALYSIS.md # Codebase analysis
    ├── WHY_REACT.md         # ReAct pattern rationale
    ├── RATE_LIMITING_GUIDE.md # Rate limiting guide
    ├── FEWSHOT_GUIDE.md     # Few-shot learning guide
    ├── JSON_SCHEMA.md       # JSON schema documentation
    ├── COMPARISON.md        # Feature comparisons
    └── (marketing/planning docs)
```

## Critical Directories

### 1. `agent/` - Core Library Implementation

**Purpose:** Contains the entire agent SDK implementation

**Key Files:**
- `builder.go` - Primary entry point for creating agents
- `agent.go` - Core Agent struct and Chat() method
- `builder_*.go` - Modular builder extensions for different features

**Submodules:**
- `adapters/` - Provider-specific implementations (OpenAI, Gemini)
- `memory/` - Memory management system
- `tools/` - Built-in tool implementations

**Entry Points:**
- `agent.NewOpenAI(model, apiKey) *Builder`
- `agent.NewGemini(model, apiKey) *Builder`

### 2. `examples/` - Usage Examples

**Purpose:** Comprehensive examples demonstrating library features

**Categories:**
- **ReAct Pattern** - Reasoning and acting examples (react_*)
- **Planning** - Planning and orchestration (planner_*)
- **Rate Limiting** - Request throttling (rate_limit_*)
- **Configuration** - Config examples (config_basic, full_config)
- **Tools** - Tool usage (tool_choice_demo, math_teacher)
- **Personas** - System prompts and personas (persona_basic)
- **Few-shot** - Few-shot learning (fewshot_basic)

### 3. `docs/` - Documentation

**Purpose:** Project documentation and guides

**Structure:**
- `guides/` - User guides
- `api/` - API documentation
- `releases/` - Version release notes
- `development/` - Development documentation
- Root `.md` files - Various technical guides

### 4. `agent/memory/` - Memory Subsystem

**Purpose:** Agent memory management

**Types:**
- `episodic.go` - Sequential event memory
- `semantic.go` - Factual knowledge storage
- `working.go` - Short-term active context
- `system.go` - System-level state
- `interfaces.go` - Memory abstractions

### 5. `agent/tools/` - Tool System

**Purpose:** Built-in tool implementations

**Tools:**
- `datetime.go` - Time/date operations
- `filesystem.go` - File operations
- `http.go` - HTTP requests
- `math.go` - Mathematical operations
- `logger.go` - Logging functionality
- `orchestrator.go` - Tool coordination

### 6. `agent/adapters/` - LLM Provider Adapters

**Purpose:** Provider-specific integrations

**Adapters:**
- `openai_adapter.go` - OpenAI API integration
- `gemini_adapter.go` - Google Gemini integration

## File Organization Patterns

### Builder Pattern Files

The builder is split into focused modules:

```
builder.go              # Core builder struct and factory methods
builder_config.go       # Configuration methods
builder_memory.go       # Memory configuration
builder_llm.go          # LLM-specific settings
builder_execution.go    # Execution and retry logic
builder_callbacks.go    # Streaming and callback configuration
builder_logging.go      # Logging configuration
builder_messages.go     # Message management
builder_cache.go        # Caching configuration
builder_defaults.go     # Default configurations
builder_fewshot.go      # Few-shot learning
builder_retry.go        # Retry and error recovery
```

### Test Files

- Test files follow `*_test.go` convention
- Located alongside implementation files
- Examples: `agent_test.go`, `builder_memory_test.go`, `unit_test.go`

## Technology Markers

### Go Ecosystem

- **Module:** `github.com/taipm/go-deep-agent`
- **Go Version:** 1.25.2
- **Package Manager:** Go modules (`go.mod`, `go.sum`)

### Key Dependencies (from go.mod)

- `github.com/openai/openai-go/v3` - OpenAI SDK
- `github.com/google/generative-ai-go` - Gemini SDK
- `github.com/redis/go-redis/v9` - Redis client
- `gopkg.in/yaml.v3` - YAML support
- `github.com/stretchr/testify` - Testing
- `golang.org/x/time` - Rate limiting

## Integration Points

### External Services

1. **OpenAI API** - via `agent/adapters/openai_adapter.go`
2. **Google Gemini API** - via `agent/adapters/gemini_adapter.go`
3. **Redis** - For caching and memory persistence
4. **Qdrant** - Vector database (via `agent/qdrant.go`)

### Configuration Files

- `go.mod` - Dependency management
- `.env` files - Environment variables (via godotenv)
- YAML configs - Configuration templates

## Development Workflow

### Build

```bash
go build
```

### Test

```bash
go test ./...
```

### Run Examples

```bash
cd examples/react_simple
go run main.go
```

### Module Management

```bash
go mod tidy
go mod vendor
```

## Asset Locations

No binary assets (images, fonts, etc.) - Pure Go library.

## Code Organization Philosophy

1. **Modular Builder** - Feature-specific builder files
2. **Interface-First** - Abstraction via interfaces (adapters, memory, cache, vector store)
3. **Example-Driven** - Extensive examples for all features
4. **Test Coverage** - Tests alongside implementation
5. **Documentation** - Comprehensive guides and API docs

---

**Generated:** 2025-11-14
**Scan Level:** Deep
**Project Type:** Backend Library (Go SDK)
