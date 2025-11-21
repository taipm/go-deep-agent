# Architecture - go-deep-agent Quality Improvement Initiative

## Executive Summary

This architecture document defines the technical approach for transforming go-deep-agent from a high-quality library (92/100 score, 73% coverage) into a production-hardened, security-first, performance-optimized SDK (target: 80%+ coverage, zero vulnerabilities, benchmarked performance).

**Project Type:** Brownfield quality improvement initiative
**Approach:** Additive enhancement - preserve existing proven architecture, add quality layers
**Target State:** Production-ready Go library for AI agents with enterprise-grade quality standards

**Key Principles:**
- ✅ **Keep existing architecture** - Fluent Builder Pattern, multi-provider abstraction (proven, 92/100)
- ✅ **Additive-only changes** - No breaking changes to public API
- ✅ **Quality layering** - Add security, testing, benchmarking infrastructure alongside existing code
- ✅ **Tool-first approach** - Automated quality gates in CI/CD

---

## Decision Summary

| Category | Decision | Version | Affects Epics | Rationale |
| -------- | -------- | ------- | ------------- | --------- |
| **Security Scanning** | gosec + govulncheck + golangci-lint | v2.22.10 + v1.1.4+ + v2.6.1 | Epic 1-3 | Layered security: code patterns + dependencies + comprehensive linting |
| **Benchmarking** | Go testing + benchstat + github-action-benchmark | Built-in + latest + latest | Epic 4-6 | Native foundation with statistical rigor and CI visualization |
| **Test Mocking** | testify/mock + mockery | Latest | Epic 7-9 | Industry standard, auto-generation, maintainable tests |
| **Coverage Reporting** | Codecov | Latest | Epic 7-9 | Best-in-class analytics, free for open source |
| **Race Detection** | go test -race | Built-in | Epic 7-9 | Native, comprehensive, CI-integrated |
| **Structured Logging** | slog + RedactHandler | Go 1.21+ | Epic 3, 13 | Native, secure (secrets protection), structured |
| **Error Handling** | Structured errors + %w wrapping | Go 1.13+ | Epic 13-14 | Context chain, abstraction levels |
| **Observability** | OpenTelemetry (optional hooks) | Latest | Epic 13 | User choice, not forced dependency |
| **Configuration** | Env vars + programmatic + validation | - | Epic 15 | Flexible, validated, secure defaults |

---

## Project Structure

### Current Structure (Preserved)

```
go-deep-agent/
├── agent/                    # Core library - NO CHANGES
│   ├── adapters/            # Provider adapters (OpenAI, Gemini)
│   ├── memory/              # Memory systems (episodic, semantic, working)
│   ├── tools/               # Built-in tools
│   ├── builder.go           # Fluent API builder pattern [ENTRY POINT]
│   ├── agent.go             # Core Agent struct
│   └── builder_*.go         # Builder extensions
├── examples/                 # 20+ usage examples - PRESERVED
├── docs/                     # Documentation - ENHANCED
└── *.md                      # Root docs - PRESERVED
```

### Enhanced Structure (Additions for Quality Initiative)

```
go-deep-agent/
├── agent/
│   ├── security/            # 🆕 Epic 3: Security infrastructure
│   │   ├── validation.go    # Input validation (API keys, prompts, tools, configs)
│   │   ├── tls.go          # TLS 1.2+ enforcement
│   │   ├── secrets.go      # Secrets protection, redaction
│   │   ├── rotation.go     # API key rotation support
│   │   └── audit.go        # Audit logging for sensitive operations
│   │
│   ├── testutil/            # 🆕 Epic 7-9: Test infrastructure
│   │   ├── mocks/          # Generated mocks (mockery)
│   │   │   ├── mock_provider.go
│   │   │   ├── mock_tool.go
│   │   │   └── mock_memory.go
│   │   ├── fixtures/       # Test fixtures
│   │   │   ├── embeddings.go    # RAG test data
│   │   │   ├── documents.go     # Sample documents
│   │   │   └── responses.go     # Mock LLM responses
│   │   └── helpers.go      # Test utilities
│   │
│   └── benchmarks/          # 🆕 Epic 4: Benchmark suite
│       ├── agent_bench_test.go      # Agent creation (<1ms target)
│       ├── tool_bench_test.go       # Tool dispatch (<100μs target)
│       ├── memory_bench_test.go     # Memory ops (<10ms target)
│       ├── rag_bench_test.go        # RAG search (<50ms/10K docs target)
│       ├── provider_bench_test.go   # Provider adapter overhead
│       └── batch_bench_test.go      # Batch throughput (>100 ops/sec target)
│
├── .github/workflows/        # 🔧 Enhanced CI/CD
│   ├── security-scan.yml    # 🆕 Epic 1: gosec + govulncheck + golangci-lint
│   ├── test-coverage.yml    # 🆕 Epic 7-9: Coverage + race detection
│   ├── benchmark.yml        # 🆕 Epic 4-5: Continuous benchmarking
│   └── quality-gate.yml     # 🆕 Epic 10-11: Code quality checks
│
├── docs/
│   ├── security/            # 🆕 Epic 1-3: Security documentation
│   │   ├── security-baseline-report.md
│   │   ├── security-scanning.md
│   │   └── vulnerability-response.md
│   │
│   ├── performance/         # 🆕 Epic 4-6: Performance documentation
│   │   ├── benchmark-results/
│   │   │   └── YYYY-MM-DD-baseline.txt
│   │   └── performance-guide.md
│   │
│   └── adr/                 # 🆕 Epic 12: Architecture Decision Records
│       ├── 0001-security-tooling-stack.md
│       ├── 0002-performance-benchmarking.md
│       ├── 0003-test-infrastructure.md
│       └── 0004-cross-cutting-concerns.md
│
├── .gosec.json              # 🆕 Epic 1: gosec configuration
├── .golangci.yml            # 🆕 Epic 1: golangci-lint v2 config
├── codecov.yml              # 🆕 Epic 9: Codecov configuration
│
└── scripts/                 # 🆕 Quality automation scripts
    ├── run-security-scan.sh     # Local security scanning
    ├── run-benchmarks.sh        # Local benchmark running
    ├── generate-mocks.sh        # Mockery code generation
    └── validate-coverage.sh     # Coverage threshold check
```

---

## Epic to Architecture Mapping

### Phase 1: Security Hardening (Epic 1-3)

**Epic 1: Security Infrastructure Foundation**
- **Architecture:** `agent/security/` package, `.gosec.json`, `.golangci.yml`, `.github/workflows/security-scan.yml`
- **Tools:** gosec v2.22.10, govulncheck v1.1.4+, golangci-lint v2.6.1
- **Integration:** GitHub Actions with SARIF upload to GitHub Security tab
- **Stories:** 1.1-1.6 (6 stories)

**Epic 2: Input Validation & Sanitization**
- **Architecture:** `agent/security/validation.go` (API keys, prompts, tools, configs)
- **Patterns:** Validate() → Sanitize() → Use() pipeline
- **Stories:** 2.1-2.4 (4 stories)

**Epic 3: Secure Defaults & Authentication**
- **Architecture:** `agent/security/tls.go`, `secrets.go`, `rotation.go`, `audit.go`
- **Logging:** slog with custom RedactHandler wrapper (secrets protection)
- **TLS:** Minimum TLS 1.2, strong cipher suites
- **Stories:** 3.1-3.6 (6 stories)

### Phase 2: Performance Optimization (Epic 4-6)

**Epic 4: Benchmark Suite Development**
- **Architecture:** `agent/benchmarks/` directory (centralized benchmarks)
- **Framework:** Go testing package with B.Loop (new standard)
- **Targets:**
  - Agent creation: <1ms
  - Tool dispatch: <100μs
  - Memory ops: <10ms
  - RAG search: <50ms (10K docs)
  - Batch throughput: >100 ops/sec
- **Stories:** 4.1-4.6 (6 stories)

**Epic 5: Performance Baselines & CI Integration** (High-level guidance)
- **Tools:** benchstat (statistical comparison, α=0.05, 20 runs)
- **CI:** github-action-benchmark (automated regression detection, 10% threshold)
- **Storage:** Baselines in gh-pages branch
- **Reporting:** PR comments with performance comparison

**Epic 6: Performance Optimization Implementation** (High-level guidance)
- **Focus Areas:** Memory allocations, goroutine pooling, cache optimization, connection pooling, lazy loading
- **Measurement:** Before/after benchmarks for each optimization

### Phase 3: Test Coverage Enhancement (Epic 7-9) (High-level guidance)

**Epic 7: Unit Test Expansion**
- **Architecture:** `agent/testutil/mocks/` (mockery generated), `agent/testutil/fixtures/`
- **Mocking:** testify/mock + mockery (industry standard)
- **Targets:** Core 85%+, Adapters 80%+, Memory 85%+, Tools 75%+

**Epic 8: Integration & Edge Case Tests**
- **Coverage:** Multi-provider scenarios, error paths, edge cases, concurrent tests
- **Race Detection:** `go test -race` in separate CI job (5-10x memory overhead)

**Epic 9: Test Infrastructure & Reporting**
- **Coverage:** Codecov integration, codecov-action@v4
- **Reporting:** `go test -coverprofile=coverage.out -covermode=atomic`
- **CI:** Upload to Codecov, coverage badges in README

### Phase 4: Code Quality & Technical Debt (Epic 10-12) (High-level guidance)

**Epic 10: Technical Debt Elimination**
- **Focus:** Duplicate code, hardcoded values, dead code, code smells (>50 lines, cyclomatic >10)

**Epic 11: Code Quality Tooling & Standards**
- **Architecture:** `.golangci.yml` v2 schema with comprehensive linters
- **Target:** Go Report Card A+ rating
- **Linters:** gosec, govet, staticcheck, errcheck, exportloopref, noctx, rowserrcheck, sqlclosecheck

**Epic 12: Documentation Enhancement**
- **Architecture:** `docs/adr/` (Architecture Decision Records)
- **Standards:** Package-level godoc (100%), exported symbols (100%), code examples
- **ADR Format:** Context, Decision, Consequences

### Phase 5: Production Hardening (Epic 13-15) (High-level guidance)

**Epic 13: Error Handling & Observability**
- **Errors:** Structured error types with %w wrapping, context chain
- **Logging:** slog with JSONHandler (production), custom fields (request_id, correlation_id)
- **Tracing:** OpenTelemetry optional hooks (user choice, not forced)

**Epic 14: Resilience Patterns**
- **Circuit Breaker:** For provider API calls
- **Retry Logic:** Exponential backoff (max 3 retries, 2x backoff)
- **Timeouts:** Provider 30s, tools 10s (configurable)
- **Context Cancellation:** Throughout all operations

**Epic 15: Configuration & Deployment Readiness**
- **Config:** Environment variables + programmatic + validation
- **Validation:** On startup (fail-fast)
- **Docker:** Example deployment configuration

---

## Technology Stack Details

### Core Technologies (Existing - Preserved)

- **Language:** Go 1.25.2
- **OpenAI SDK:** v3.8.1
- **Gemini SDK:** v0.20.1
- **Redis:** v9 (caching)
- **Qdrant:** Vector store integration
- **Pattern:** Fluent Builder API

### Quality Tooling Stack (New - Additions)

**Security:**
- gosec v2.22.10 - Static security analysis
- govulncheck v1.1.4+ - Dependency vulnerability scanning
- golangci-lint v2.6.1 - Comprehensive linting (50+ linters)

**Performance:**
- Go testing package - Benchmark framework (B.Loop)
- benchstat - Statistical comparison (golang.org/x/perf/cmd/benchstat)
- github-action-benchmark - CI automation + visualization

**Testing:**
- testify/mock - Assertion framework + mocking
- mockery - Mock code generation
- Codecov - Coverage reporting + analytics
- go test -race - Race condition detection

**Observability:**
- log/slog - Structured logging (Go 1.21+)
- OpenTelemetry - Optional tracing hooks (not forced)

---

## Integration Points

### CI/CD Integration (GitHub Actions)

**1. Security Scan Workflow** (Daily + PR)
```yaml
jobs:
  security:
    - Install gosec, govulncheck
    - Run gosec -fmt=sarif (fail on high/critical)
    - Run govulncheck (fail on any vulnerability)
    - Upload SARIF to GitHub Security tab
    - Run golangci-lint (comprehensive checks)
```

**2. Test Coverage Workflow** (PR + main branch)
```yaml
jobs:
  test:
    - go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    - Upload to Codecov (codecov-action@v4)

  race:
    - go test -race ./... (separate job, 2-20x slower)
```

**3. Benchmark Workflow** (PR + main branch)
```yaml
jobs:
  benchmark:
    - go test -bench=. -benchmem (20 iterations)
    - benchstat comparison (current vs baseline)
    - github-action-benchmark (fail on >10% regression)
    - Store baselines in gh-pages
    - Comment on PR with results
```

**4. Quality Gate Workflow** (PR)
```yaml
jobs:
  quality:
    - golangci-lint v2 comprehensive checks
    - Coverage threshold validation (80%+)
    - Go Report Card checks
```

---

## Implementation Patterns

### Security Patterns

**Input Validation Pipeline:**
```go
// agent/security/validation.go

func ValidateAPIKey(provider string, key string) error {
    // Format validation (regex)
    // Length validation (min 32, max 512)
    // Character restrictions
    // Prefix validation (sk-, AI...)
    // No whitespace
    return nil // or error
}

func SanitizePrompt(input string, level PromptProtectionLevel) (string, error) {
    // Detect injection patterns
    // Filter control characters
    // Enforce length limits
    // Log suspicious patterns (audit)
    return sanitized, nil
}
```

**Secrets Protection:**
```go
// agent/security/secrets.go

type RedactHandler struct {
    handler slog.Handler
}

func (h *RedactHandler) Handle(ctx context.Context, r slog.Record) error {
    r.Message = RedactSensitive(r.Message)
    // Redact: sk-[A-Za-z0-9]{48} → sk-***...***
    return h.handler.Handle(ctx, r)
}
```

### Performance Patterns

**Benchmark Structure (B.Loop):**
```go
// agent/benchmarks/agent_bench_test.go

func BenchmarkAgentCreation_Simple(b *testing.B) {
    b.ReportAllocs()
    for b.Loop() { // New standard (not b.N)
        agent, _ := agent.New().
            WithProvider(provider.NewMock()).
            Build()
        _ = agent
    }
}
```

**Statistical Comparison:**
```bash
# 20 runs for statistical significance
go test -bench=. -benchmem -count=20 > new.txt
benchstat baseline.txt new.txt
# α=0.05 threshold, reports statistical significance
```

### Testing Patterns

**Mock Generation:**
```bash
# scripts/generate-mocks.sh
mockery --name=Provider --dir=agent/adapters --output=agent/testutil/mocks/
mockery --name=Tool --dir=agent/tools --output=agent/testutil/mocks/
mockery --name=Memory --dir=agent/memory --output=agent/testutil/mocks/
```

**Test Organization:**
```go
// agent/adapters/openai_adapter_test.go (unit test)
func TestOpenAIAdapter_Chat(t *testing.T) { ... }

// agent/adapters/openai_adapter_integration_test.go (integration test)
// +build integration
func TestOpenAIAdapter_Chat_Integration(t *testing.T) { ... }

// agent/benchmarks/provider_bench_test.go (benchmark)
func BenchmarkProvider_RequestPreparation(b *testing.B) { ... }
```

**Coverage Strategy:**
```yaml
# codecov.yml
coverage:
  status:
    project:
      default:
        target: 80%
    patch:
      default:
        target: 75%

ignore:
  - "**/*_test.go"
  - "**/testutil/**"
  - "**/mocks/**"
```

### Error Handling Patterns

**Structured Errors:**
```go
type AgentError struct {
    Op      string                 // Operation that failed
    Kind    ErrorKind              // Error category
    Err     error                  // Underlying error
    Context map[string]interface{} // Additional context
}

func (e *AgentError) Error() string {
    return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *AgentError) Unwrap() error {
    return e.Err
}
```

**Error Wrapping:**
```go
func (a *Agent) Chat(prompt string) (*Response, error) {
    if err := validatePrompt(prompt); err != nil {
        return nil, fmt.Errorf("agent: chat failed: %w", err)
    }

    resp, err := a.provider.GenerateResponse(prompt)
    if err != nil {
        return nil, fmt.Errorf("agent: provider call failed: %w", err)
    }

    return resp, nil
}
```

### Configuration Pattern

**Environment Variables + Programmatic:**
```go
type Config struct {
    APIKey          string        `env:"OPENAI_API_KEY"`
    Timeout         time.Duration `env:"AGENT_TIMEOUT" default:"30s"`
    MaxRetries      int          `env:"AGENT_MAX_RETRIES" default:"3"`
    EnableTLS       bool         `env:"AGENT_TLS_ENABLED" default:"true"`
    MinTLSVersion   uint16       // tls.VersionTLS12
}

func (c *Config) Validate() error {
    if err := ValidateAPIKey("openai", c.APIKey); err != nil {
        return fmt.Errorf("invalid API key: %w", err)
    }
    if c.Timeout < time.Second || c.Timeout > 10*time.Minute {
        return errors.New("timeout must be between 1s and 10m")
    }
    if c.MaxRetries < 0 || c.MaxRetries > 10 {
        return errors.New("max retries must be between 0 and 10")
    }
    return nil
}
```

---

## Consistency Rules

### Naming Conventions

- **Files:** `snake_case.go` (Go standard)
- **Packages:** lowercase, single word (e.g., `security`, `testutil`, `benchmarks`)
- **Exported symbols:** `PascalCase` (e.g., `Agent`, `Provider`, `ValidateAPIKey`)
- **Unexported symbols:** `camelCase` (e.g., `validatePrompt`, `redactSecrets`)
- **Interfaces:** `-er` suffix when possible (e.g., `Provider`, `Tracer`, `Logger`)
- **Test files:** `*_test.go` (co-located with source)
- **Benchmark files:** `*_bench_test.go` or in `benchmarks/` directory
- **Mock files:** `mock_*.go` in `testutil/mocks/`

### Code Organization

- **Tests:** Co-located `*_test.go` files
- **Integration tests:** `*_integration_test.go` with build tag `// +build integration`
- **Benchmarks:** Centralized in `agent/benchmarks/` directory
- **Mocks:** Generated in `agent/testutil/mocks/` directory
- **Fixtures:** Test data in `agent/testutil/fixtures/` directory
- **Examples:** Separate `examples/` directory (existing)

### Error Messages

- **Format:** Lowercase, no period (e.g., `"failed to initialize provider"`)
- **Include context:** Operation + reason (e.g., `"failed to initialize provider: invalid API key"`)
- **Wrapping:** Use `%w` verb for error chain
- **Redaction:** No secrets in error messages

### Logging Format

**Production (JSON):**
```json
{
  "time": "2025-01-14T10:30:00Z",
  "level": "ERROR",
  "msg": "provider call failed",
  "provider": "openai",
  "model": "gpt-4",
  "error": "timeout exceeded",
  "request_id": "req-abc-123",
  "correlation_id": "corr-xyz-789"
}
```

**Development (Text):**
```
2025-01-14T10:30:00Z ERROR provider call failed provider=openai model=gpt-4 error="timeout exceeded" request_id=req-abc-123
```

**Standard Fields:**
- `request_id` - Unique request identifier
- `correlation_id` - Cross-service correlation
- `user_id` - User identifier (if applicable)
- `operation` - Operation being performed

---

## Performance Considerations

### Performance Targets (From NFRs)

**Agent Operations:**
- Agent creation overhead: <1ms (framework only)
- Tool execution framework overhead: <100μs (excluding tool logic)
- Memory operation latency: <10ms (episodic/semantic/working)

**RAG Performance:**
- Vector search latency: <50ms (10K documents), <200ms (100K documents)
- Document indexing throughput: >100 documents/second
- Cache hit rate: >80% for repeated queries

**Batch Processing:**
- Throughput: >100 operations/second (concurrent batch requests)
- Concurrency: Configurable goroutine pool (default: NumCPU * 2)

**Memory Efficiency:**
- Heap allocation: <10MB overhead for base agent (excluding LLM SDKs)
- GC pressure: Minimal allocations in hot paths
- Memory limits: Configurable (default: 100MB caching, 500MB vector storage)

### Optimization Strategies

**Memory Allocation Minimization:**
- Pre-allocate slices with known capacity
- Reuse buffers where possible
- Use sync.Pool for frequently allocated objects

**Goroutine Pooling:**
- Bounded concurrency with worker pools
- Prevent goroutine leaks with proper cleanup
- Context cancellation throughout

**Caching Strategy:**
- In-memory LRU cache for LLM responses
- Redis for distributed deployments
- Configurable TTL and eviction policies

**Connection Pooling:**
- HTTP keep-alive for provider APIs
- Max idle connections: 10, max per host: 100
- Idle connection timeout: 90s

---

## Deployment Architecture

### Docker Deployment (Example)

```dockerfile
# Multi-stage build for minimal image size
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o agent-app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/agent-app .

# Environment variables
ENV OPENAI_API_KEY=""
ENV AGENT_TIMEOUT="30s"
ENV AGENT_MAX_RETRIES="3"
ENV AGENT_TLS_ENABLED="true"

CMD ["./agent-app"]
```

### Development Environment

**Prerequisites:**
- Go 1.21+ (for slog support)
- Git
- make (optional, for automation)

**Setup Commands:**
```bash
# Clone repository
git clone https://github.com/taipm/go-deep-agent.git
cd go-deep-agent

# Install dependencies
go mod download

# Install quality tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/perf/cmd/benchstat@latest
go install github.com/vektra/mockery/v2@latest

# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./agent/benchmarks/

# Run security scans
gosec -conf=.gosec.json ./...
govulncheck ./...
golangci-lint run

# Generate mocks
./scripts/generate-mocks.sh

# Run coverage
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out
```

---

## Architecture Decision Records (ADRs)

### ADR-0001: Security Tooling Stack

**Status:** Accepted
**Date:** 2025-01-14

**Context:**
go-deep-agent requires zero high/critical vulnerabilities for production readiness. Epic 1-3 mandate automated security scanning.

**Decision:**
Use layered security approach with three tools:
1. gosec v2.22.10 - Static code security patterns
2. govulncheck v1.1.4+ - Dependency vulnerabilities
3. golangci-lint v2.6.1 - Comprehensive linting (includes gosec)

**Consequences:**
- ✅ Comprehensive coverage (code + dependencies + quality)
- ✅ Low overlap, each tool serves distinct purpose
- ✅ Industry standard tools, well-maintained
- ✅ GitHub Actions integration with SARIF reporting
- ⚠️ Three tools to maintain, but all are stable

### ADR-0002: Performance Benchmarking Infrastructure

**Status:** Accepted
**Date:** 2025-01-14

**Context:**
Epic 4-6 requires comprehensive benchmarking with targets (<1ms, <100μs, <10ms, <50ms) and regression detection.

**Decision:**
Layered benchmarking approach:
1. Go testing package with B.Loop - Implementation
2. benchstat - Statistical comparison (α=0.05, 20 runs)
3. github-action-benchmark - CI automation + visualization

**Consequences:**
- ✅ Native Go foundation (zero external deps for benchmarks)
- ✅ Statistical rigor prevents false positives
- ✅ Developer visibility via PR comments + GitHub Pages
- ✅ Performance targets enforced in CI (fail on >10% regression)
- ⚠️ Requires baseline establishment and maintenance

### ADR-0003: Test Infrastructure

**Status:** Accepted
**Date:** 2025-01-14

**Context:**
Epic 7-9 requires 80%+ coverage (85% core, 80% adapters, 75% tools) with mock providers and race detection.

**Decision:**
Comprehensive test stack:
1. testify/mock + mockery - Mocking framework + code generation
2. Codecov - Coverage reporting + analytics
3. go test -race - Race condition detection (separate CI job)
4. Custom testutil package - Test fixtures and helpers

**Consequences:**
- ✅ Industry standard mocking (testify ecosystem)
- ✅ Best-in-class coverage analytics (Codecov)
- ✅ Native race detection (critical for concurrent library)
- ✅ Reusable test infrastructure reduces duplication
- ⚠️ Mockery code generation step required

### ADR-0004: Cross-Cutting Concerns

**Status:** Accepted
**Date:** 2025-01-14

**Context:**
Patterns affecting all epics: logging, errors, observability, configuration.

**Decisions:**
1. **Logging:** slog (Go 1.21+) with custom RedactHandler for secrets protection
2. **Errors:** Structured error types with %w wrapping, context chain
3. **Tracing:** OpenTelemetry optional hooks (user choice, not forced)
4. **Config:** Environment variables + programmatic + validation

**Consequences:**
- ✅ Native slog (no external logging deps), secure by default
- ✅ Error context chain for debugging, abstraction levels maintained
- ✅ Tracing optional (library doesn't force observability stack)
- ✅ Flexible configuration (env vars + code), validated on startup
- ⚠️ Requires Go 1.21+ for slog support

---

## Next Steps

### Immediate Implementation (Phase 1 - Epic 1-4 Ready)

**Week 1-4: Security Hardening (Epic 1-3)**
1. Epic 1: Security Infrastructure Foundation (6 stories)
   - Story 1.1: Security audit baseline
   - Story 1.2-1.6: gosec, govulncheck, golangci-lint, CI/CD, dashboard

2. Epic 2: Input Validation (4 stories)
   - Story 2.1-2.4: API keys, prompts, tools, configs

3. Epic 3: Secure Defaults & Authentication (6 stories)
   - Story 3.1-3.6: TLS, secrets, rotation, credentials, access control, audit

**Week 5-7: Performance Foundation (Epic 4)**
4. Epic 4: Benchmark Suite Development (6 stories)
   - Story 4.1-4.6: Agent, tool, memory, RAG, provider, batch benchmarks

### Workflow Continuation

**After Architecture Review:**
1. ✅ **Validate Architecture** - Use `/bmad:bmm:workflows:validate-architecture` (optional)
2. ✅ **Start Implementation** - Use `/bmad:bmm:workflows:dev-story` for Story 1.1
3. ✅ **Solutioning Gate Check** - Before Phase 4, validate PRD + Architecture alignment

**Remaining Epics (High-Level Guidelines Provided):**
- Epic 5-6: Performance optimization (detailed when Epic 4 completes)
- Epic 7-9: Test coverage enhancement (detailed when Epic 6 completes)
- Epic 10-12: Code quality & technical debt (detailed when Epic 9 completes)
- Epic 13-15: Production hardening (detailed when Epic 12 completes)

---

## Document Metadata

**Version:** 1.0
**Date:** 2025-01-14
**Author:** BMad (Product Owner) + Winston (Architect)
**Status:** Ready for Implementation
**Workflow:** BMad Method - Phase 3 Solutioning - Architecture Workflow

**Updates:**
- Epic 1-4: Fully detailed with technical decisions
- Epic 5-15: High-level guidelines, to be elaborated just-in-time

**Related Documents:**
- [PRD](./PRD.md) - Product requirements (126 FRs)
- [Epics](./epics.md) - Epic & story breakdown (22 stories detailed, 66 estimated)
- [Project Overview](./project-overview.md) - Executive summary
- [Source Tree Analysis](./source-tree-analysis.md) - Current codebase structure

---

_Generated by BMAD Decision Architecture Workflow v1.0_
_For: go-deep-agent Quality Improvement Initiative_
_Approach: Brownfield enhancement - additive quality layers, preserve proven architecture_
