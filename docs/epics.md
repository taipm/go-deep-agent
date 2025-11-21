# go-deep-agent - Epic Breakdown

**Author:** BMad
**Date:** 2025-11-14
**Project Level:** TBD
**Target Scale:** TBD

---

## Overview

This document provides the complete epic and story breakdown for go-deep-agent, decomposing the requirements from the [PRD](./PRD.md) into implementable stories.

**Living Document Notice:** This is the initial version. It will be updated after UX Design and Architecture workflows add interaction and technical details to stories.

## Epic Overview

Dự án có **15 epics** được tổ chức theo **5 phases** từ PRD:

### Phase 1: Security Hardening (4 weeks)
- **Epic 1:** Security Infrastructure Foundation (FR1-FR2)
- **Epic 2:** Input Validation & Sanitization (FR3-FR6)
- **Epic 3:** Secure Defaults & Authentication (FR7-FR12)

### Phase 2: Performance Optimization (3 weeks)
- **Epic 4:** Benchmark Suite Development (FR13-FR18)
- **Epic 5:** Performance Baselines & CI Integration (FR19-FR21)
- **Epic 6:** Performance Optimization Implementation (FR22-FR26)

### Phase 3: Test Coverage Enhancement (4 weeks)
- **Epic 7:** Unit Test Expansion (FR27-FR31)
- **Epic 8:** Integration & Edge Case Tests (FR32-FR35)
- **Epic 9:** Test Infrastructure & Reporting (FR36-FR39)

### Phase 4: Code Quality & Technical Debt (3 weeks)
- **Epic 10:** Technical Debt Elimination (FR40-FR43)
- **Epic 11:** Code Quality Tooling & Standards (FR44-FR46)
- **Epic 12:** Documentation Enhancement (FR47-FR50)

### Phase 5: Production Hardening (2 weeks)
- **Epic 13:** Error Handling & Observability (FR51-FR55)
- **Epic 14:** Resilience Patterns (FR56-FR63)
- **Epic 15:** Configuration & Deployment Readiness (FR64-FR69)

**Note:** FR70-FR120 (Developer Experience, Multi-Provider Support, Memory System, RAG, Tools, Streaming, Caching) đại diện cho **existing capabilities** đã implemented. Chúng sẽ được **enhanced through quality epics** (better tests, docs, performance, security) chứ không phải reimplemented.

---

## Functional Requirements Inventory

Tổng số: **120 Functional Requirements** được tổ chức thành 12 nhóm chức năng:

### Security & Validation (FR1-FR12)
- FR1: Automated security scanning với gosec
- FR2: Vulnerability scanning với govulncheck
- FR3: API key input validation
- FR4: Prompt injection protection
- FR5: Tool parameter validation
- FR6: Configuration input sanitization
- FR7: TLS 1.2+ enforcement
- FR8: Secrets protection trong logs
- FR9: API key rotation support
- FR10: Provider credential validation
- FR11: Scope-based access control cho tools
- FR12: Audit logging cho sensitive operations

### Performance & Benchmarking (FR13-FR26)
- FR13: Agent creation benchmarks
- FR14: Tool execution benchmarks
- FR15: Memory system benchmarks
- FR16: RAG vector search benchmarks
- FR17: Provider adapter benchmarks
- FR18: Batch processing benchmarks
- FR19: Performance baseline tracking
- FR20: Regression detection trong CI/CD
- FR21: Performance comparison reports
- FR22: Memory allocation minimization
- FR23: Goroutine pooling
- FR24: Caching với configurable hit rate
- FR25: Connection pooling cho provider APIs
- FR26: Lazy loading cho expensive initializations

### Test Coverage & Quality (FR27-FR39)
- FR27: Core agent package 85%+ coverage
- FR28: Provider adapters 80%+ coverage
- FR29: Memory system 85%+ coverage
- FR30: Tool execution 75%+ coverage
- FR31: Unit tests cho public APIs
- FR32: Integration tests multi-provider
- FR33: Error path tests (timeout, retry)
- FR34: Edge case tests (nil, boundaries)
- FR35: Concurrent tests (race detection)
- FR36: Mock provider adapters
- FR37: Test fixtures (RAG, embeddings)
- FR38: Coverage reports trong CI/CD
- FR39: Coverage badges trong docs

### Code Quality & Maintainability (FR40-FR50)
- FR40: Eliminate duplicate code
- FR41: Eliminate hardcoded values
- FR42: Eliminate dead code
- FR43: Eliminate code smells (>50 lines, cyclomatic >10)
- FR44: gofmt/goimports formatting
- FR45: golangci-lint comprehensive checks
- FR46: Go Report Card A+ rating
- FR47: Package-level documentation
- FR48: Godoc comments cho exports
- FR49: Code examples trong documentation
- FR50: Architecture decision records (ADR)

### Error Handling & Resilience (FR51-FR63)
- FR51: Structured error types với context
- FR52: Error wrapping với stack traces
- FR53: Configurable logging levels
- FR54: Optional metrics collection hooks
- FR55: OpenTelemetry distributed tracing
- FR56: Circuit breaker cho provider calls
- FR57: Graceful degradation với fallbacks
- FR58: Context cancellation support
- FR59: Resource cleanup (defer, timeouts)
- FR60: Panic recovery trong user tools
- FR61: Timeout defaults cho external calls
- FR62: Retry với exponential backoff
- FR63: Memory limits cho caching/storage

### Configuration & Deployment (FR64-FR69)
- FR64: Environment variable configuration
- FR65: Configuration validation on startup
- FR66: Version compatibility checks
- FR67: Migration guides cho breaking changes
- FR68: Docker deployment example
- FR69: Hot-reloading của safe configs

### Developer Experience (FR70-FR78)
- FR70: Fluent builder API
- FR71: Comprehensive API documentation
- FR72: Quick start guides
- FR73: Migration guides
- FR74: Troubleshooting guides
- FR75: Performance tuning guides
- FR76: Security best practices guide
- FR77: Example applications
- FR78: Template code cho common patterns

### Multi-Provider Support (FR79-FR85)
- FR79: OpenAI provider với GPT models
- FR80: Google Gemini provider
- FR81: Provider adapter interface
- FR82: Provider-specific error handling
- FR83: Provider-specific features (streaming, function calling, vision)
- FR84: Unified interface across providers
- FR85: Provider failover và load balancing

### Memory System (FR86-FR92)
- FR86: Episodic memory cho conversation history
- FR87: Semantic memory cho long-term storage
- FR88: Working memory cho current context
- FR89: Memory persistence across sessions
- FR90: Memory retrieval với similarity search
- FR91: Memory cleanup và garbage collection
- FR92: Memory size limits và eviction policies

### RAG & Vector Search (FR93-FR100)
- FR93: Document ingestion
- FR94: Text chunking với strategies
- FR95: Embedding generation
- FR96: Vector storage (memory, Qdrant)
- FR97: Similarity search configurable
- FR98: Hybrid search (semantic + keyword)
- FR99: Metadata filtering
- FR100: Incremental index updates

### Tool Execution (FR101-FR108)
- FR101: Custom tool registration
- FR102: Tool parameter validation
- FR103: Tool execution timeouts
- FR104: Tool failure error handling
- FR105: Async tool execution
- FR106: Tool result caching
- FR107: Built-in tools (filesystem, HTTP, math)
- FR108: Tool permission scoping

### Streaming & Batch Processing (FR109-FR114)
- FR109: Streaming LLM responses
- FR110: Streaming tool results
- FR111: Batch processing multiple requests
- FR112: Concurrent batch với rate limiting
- FR113: Progress tracking cho batch ops
- FR114: Cancellation của streaming/batch

### Caching (FR115-FR120)
- FR115: In-memory LLM response caching
- FR116: Redis caching cho distributed
- FR117: Configurable cache TTL/eviction
- FR118: Cache hit rate monitoring
- FR119: Cache invalidation by patterns
- FR120: Cache warming cho predictable queries

---

## FR Coverage Map

### Epic-to-FR Mapping

**Phase 1: Security Hardening**
- **Epic 1 (Security Infrastructure Foundation):** FR1-FR2
  - Establishes foundation for all security work
- **Epic 2 (Input Validation & Sanitization):** FR3-FR6
  - Protects against injection and malformed inputs
- **Epic 3 (Secure Defaults & Authentication):** FR7-FR12
  - Production-grade security defaults and authentication

**Phase 2: Performance Optimization**
- **Epic 4 (Benchmark Suite Development):** FR13-FR18
  - Comprehensive performance measurement infrastructure
- **Epic 5 (Performance Baselines & CI Integration):** FR19-FR21
  - Automated regression detection
- **Epic 6 (Performance Optimization Implementation):** FR22-FR26
  - Optimized library with minimal overhead

**Phase 3: Test Coverage Enhancement**
- **Epic 7 (Unit Test Expansion):** FR27-FR31
  - Critical component coverage (85%+, 80%+, 75%+)
- **Epic 8 (Integration & Edge Case Tests):** FR32-FR35
  - Production-ready complex scenario coverage
- **Epic 9 (Test Infrastructure & Reporting):** FR36-FR39
  - Professional test infrastructure with visibility

**Phase 4: Code Quality & Technical Debt**
- **Epic 10 (Technical Debt Elimination):** FR40-FR43
  - Zero technical debt (duplicate code, hardcoded values, dead code, smells)
- **Epic 11 (Code Quality Tooling & Standards):** FR44-FR46
  - Automated quality enforcement (A+ Go Report Card)
- **Epic 12 (Documentation Enhancement):** FR47-FR50
  - Professional documentation for production adoption

**Phase 5: Production Hardening**
- **Epic 13 (Error Handling & Observability):** FR51-FR55
  - Production-grade error handling and monitoring
- **Epic 14 (Resilience Patterns):** FR56-FR63
  - Fault-tolerant production operation
- **Epic 15 (Configuration & Deployment Readiness):** FR64-FR69
  - Production deployment capabilities

**Existing Capabilities (Enhanced Through All Epics):** FR70-FR120
- **Developer Experience (FR70-FR78):** Enhanced by Epic 12 (docs) + Epic 7-9 (better examples with tests)
- **Multi-Provider Support (FR79-FR85):** Enhanced by Epic 7-8 (provider adapter tests) + Epic 6 (provider performance)
- **Memory System (FR86-FR92):** Enhanced by Epic 7 (memory package tests) + Epic 6 (memory performance)
- **RAG & Vector Search (FR93-FR100):** Enhanced by Epic 7-8 (RAG tests) + Epic 6 (RAG performance benchmarks)
- **Tool Execution (FR101-FR108):** Enhanced by Epic 2 (tool validation) + Epic 7-8 (tool tests) + Epic 3 (tool permissions)
- **Streaming & Batch (FR109-FR114):** Enhanced by Epic 7-8 (streaming/batch tests) + Epic 6 (batch performance)
- **Caching (FR115-FR120):** Enhanced by Epic 6 (cache optimization) + Epic 7-8 (cache tests)

**Coverage Validation:** All 120 FRs covered - FR1-FR69 have dedicated implementation epics, FR70-FR120 are existing capabilities enhanced through quality improvements.

---

---

## Epic 1: Security Infrastructure Foundation

**Epic Goal:** Establish automated security scanning và monitoring infrastructure làm nền tảng cho tất cả công việc security. Epic này thiết lập CI/CD pipelines, security tooling, và reporting systems cho phép phát hiện vulnerabilities tự động.

**Business Value:** Zero-vulnerability foundation, automated security gates, continuous monitoring

**FR Coverage:** FR1 (gosec automation), FR2 (govulncheck automation)

---

### Story 1.1: Project Security Audit & Baseline Assessment

**As a** Security Engineer,
**I want** baseline security assessment của codebase hiện tại,
**So that** tôi hiểu current security posture và có thể track improvements.

**Acceptance Criteria:**

**Given** go-deep-agent codebase với 73% test coverage và 92/100 assessment score
**When** chạy initial security audit
**Then** tôi nhận được:
- Complete security scan report từ gosec (all files scanned)
- Vulnerability report từ govulncheck (all dependencies checked)
- Categorized findings: Critical, High, Medium, Low severity
- Baseline metrics: total issues count, issues by category, affected files
- Security debt inventory: list of all current vulnerabilities to address

**And** report được saved to `docs/security-baseline-report.md`
**And** findings được categorized theo CWE IDs
**And** each finding includes: file location, line number, description, remediation guidance

**Prerequisites:** None (first story in project)

**Technical Notes:**
- Install gosec: `go install github.com/securego/gosec/v2/cmd/gosec@latest`
- Install govulncheck: `go install golang.org/x/vuln/cmd/govulncheck@latest`
- Run gosec: `gosec -fmt=json -out=gosec-report.json ./...`
- Run govulncheck: `govulncheck -json ./... > govulncheck-report.json`
- Parse JSON reports và generate markdown summary
- Store baseline in version control for tracking progress
- Target: Complete scan in <5 minutes for CI feasibility

---

### Story 1.2: Configure gosec for Comprehensive Security Scanning

**As a** DevOps Engineer,
**I want** gosec configured với comprehensive rule set,
**So that** mọi security vulnerability được detected automatically.

**Acceptance Criteria:**

**Given** gosec installed trong development environment
**When** tạo gosec configuration file
**Then** configuration bao gồm:
- All gosec rules enabled (G101-G602)
- Severity thresholds: fail on High và Critical
- Exclusions: test files có thể có relaxed rules (G404 random in tests OK)
- Output format: JSON cho machine parsing + SARIF cho GitHub Security tab
- Concurrent scanning: utilize all CPU cores
- Custom rules cho go-deep-agent specific patterns

**And** configuration file saved to `.gosec.json` in project root
**And** configuration documented trong `docs/security-scanning.md`
**And** example run command: `gosec -conf=.gosec.json ./...`

**Prerequisites:** Story 1.1 (baseline assessment complete)

**Technical Notes:**
- Reference: https://github.com/securego/gosec#configuration
- Enable rules: G101 (hardcoded credentials), G102 (bind to all interfaces), G103 (unsafe blocks), G104 (unhandled errors), G201-G204 (SQL injection), G301-G306 (file permissions), G401-G404 (weak crypto), G501-G505 (crypto imports), G601-G602 (implicit aliasing)
- SARIF output: `gosec -fmt=sarif -out=results.sarif ./...`
- Integration với VSCode: gosec extension có thể use .gosec.json
- Performance target: <2 minutes for full scan

---

### Story 1.3: Configure govulncheck for Dependency Vulnerability Scanning

**As a** Security Engineer,
**I want** govulncheck scanning tất cả dependencies,
**So that** known vulnerabilities trong third-party packages được detected.

**Acceptance Criteria:**

**Given** govulncheck installed
**When** configure vulnerability scanning
**Then** scanning covers:
- Direct dependencies (go.mod entries)
- Transitive dependencies (full dependency tree)
- Standard library vulnerabilities (Go version specific)
- Vulnerability database: golang.org/x/vuln/vulndb updated daily
- Output format: JSON với detailed CVE information
- Scan modes: source mode (./...) và binary mode (compiled artifacts)

**And** scan command documented: `govulncheck -json -mode=source ./...`
**And** vulnerability report includes: CVE ID, affected package, version, severity, fix available
**And** false positive handling: document known safe usages in `.govulncheck.yaml`

**Prerequisites:** Story 1.1 (baseline assessment complete)

**Technical Notes:**
- Vulnerability DB: https://pkg.go.dev/golang.org/x/vuln/vulndb
- Update DB: automatically updated, no manual intervention needed
- Scan frequency: daily in CI, on-demand for PRs
- Handle zero-day: process for emergency response documented
- Performance: <1 minute typical, <5 minutes worst case
- Exit codes: 0 (no vulns), 1 (vulns found), 2 (scan error)

---

### Story 1.4: GitHub Actions CI/CD Pipeline for Security Scanning

**As a** DevOps Engineer,
**I want** automated security scanning on every PR và push to main,
**So that** vulnerabilities không bao giờ merge vào codebase.

**Acceptance Criteria:**

**Given** GitHub repository với Actions enabled
**When** tạo security scanning workflow
**Then** workflow includes:
- Trigger: on pull_request và push to main branches
- Go setup: actions/setup-go@v5 với Go 1.21+
- Checkout code: actions/checkout@v4
- Run gosec: với fail on High/Critical findings
- Run govulncheck: với fail on any vulnerability
- Generate SARIF report: upload to GitHub Security tab
- Comment on PR: summary of findings với links to details
- Job timeout: 10 minutes maximum
- Caching: cache Go modules và tool binaries

**And** workflow file saved to `.github/workflows/security-scan.yml`
**And** PR blocking: PRs cannot merge if security scan fails
**And** Status badge: security scan status visible in README
**And** Notifications: Slack/email on security scan failures in main branch

**Prerequisites:** Story 1.2 (gosec config), Story 1.3 (govulncheck config)

**Technical Notes:**
```yaml
name: Security Scan
on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main]
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight UTC

jobs:
  security:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true
      - name: Install gosec
        run: go install github.com/securego/gosec/v2/cmd/gosec@latest
      - name: Run gosec
        run: gosec -fmt=sarif -out=gosec.sarif -conf=.gosec.json ./...
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif
      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - name: Run govulncheck
        run: govulncheck -json ./...
```

---

### Story 1.5: golangci-lint Integration với Security-Focused Linters

**As a** Developer,
**I want** golangci-lint với security linters enabled,
**So that** security issues được caught during development.

**Acceptance Criteria:**

**Given** project với multiple code quality requirements
**When** configure golangci-lint
**Then** configuration includes security linters:
- gosec: security issues
- govet: suspicious constructs
- staticcheck: bugs và performance issues
- errcheck: unchecked errors
- exportloopref: loop variable capture issues
- noctx: HTTP requests without context
- rowserrcheck: SQL rows.Err() checks
- sqlclosecheck: SQL rows/statements closed

**And** configuration file: `.golangci.yml` với:
- Run timeout: 5 minutes
- Concurrency: equal to CPU count
- Issues: exclude test files from certain rules
- Severity: error level for security issues
- Output: colored-line-number format for terminal, JSON for CI

**And** pre-commit hook: optional local check before commit
**And** CI integration: runs in GitHub Actions alongside gosec
**And** IDE integration: compatible với VSCode golangci-lint extension

**Prerequisites:** Story 1.4 (CI/CD pipeline established)

**Technical Notes:**
- Install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- Run: `golangci-lint run --config=.golangci.yml ./...`
- Config example:
```yaml
linters:
  enable:
    - gosec
    - govet
    - staticcheck
    - errcheck
    - exportloopref
    - noctx
    - rowserrcheck
    - sqlclosecheck
linters-settings:
  gosec:
    severity: high
    confidence: medium
issues:
  max-issues-per-linter: 0
  max-same-issues: 0
run:
  timeout: 5m
  concurrency: 4
```

---

### Story 1.6: Security Dashboard & Reporting Infrastructure

**As a** Security Manager,
**I want** centralized security dashboard với historical tracking,
**So that** tôi có thể monitor security posture over time và report to stakeholders.

**Acceptance Criteria:**

**Given** multiple security scans running in CI/CD
**When** create security reporting infrastructure
**Then** system provides:
- GitHub Security tab: SARIF reports uploaded automatically
- Security trends: track vulnerabilities over time (weekly/monthly)
- Coverage metrics: % of code scanned, % of dependencies checked
- SLA tracking: time to remediate Critical (48h), High (7d), Medium (30d)
- Badge generation: security status badge for README
- Historical data: stored in `docs/security-reports/YYYY-MM-DD-report.md`
- Executive summary: one-page summary for stakeholders

**And** dashboard accessible via GitHub Security → Overview
**And** Automated reports: weekly security summary via email/Slack
**And** Alert thresholds: immediate notification for Critical vulnerabilities
**And** Remediation tracking: link vulnerabilities to fixing PRs

**Prerequisites:** Story 1.4 (CI/CD pipeline), Story 1.5 (golangci-lint)

**Technical Notes:**
- GitHub Security tab: automatically populated by SARIF uploads
- Trends tracking: script to parse historical SARIF reports
- Store reports in: `docs/security-reports/` directory
- Badge: shields.io compatible format
- Integration: GitHub API to fetch security data
- Visualization: Consider GitHub Actions + custom script or external tool
- Retention: keep 90 days of detailed reports, 1 year of summaries

---

## Epic 2: Input Validation & Sanitization

**Epic Goal:** Protect against injection attacks và malformed inputs by implementing comprehensive validation cho all external inputs (API keys, prompts, tool parameters, configs).

**Business Value:** Prevent security breaches, ensure data integrity, production-grade input handling

**FR Coverage:** FR3 (API key validation), FR4 (prompt injection protection), FR5 (tool parameter validation), FR6 (config sanitization)

---

### Story 2.1: API Key Input Validation Framework

**As a** Developer,
**I want** comprehensive API key validation,
**So that** invalid or malicious API keys được rejected trước khi reaching provider APIs.

**Acceptance Criteria:**

**Given** users provide API keys cho OpenAI, Gemini, và other providers
**When** implement API key validation
**Then** validation checks:
- Format validation: matches provider-specific patterns
  - OpenAI: `sk-[A-Za-z0-9]{48}` (project keys: `sk-proj-...`)
  - Gemini: `[A-Za-z0-9_-]{39}`
- Length validation: within expected ranges (min 32, max 512 chars)
- Character restrictions: alphanumeric + allowed special chars only
- Prefix validation: correct provider prefix (sk-, AI...)
- No whitespace: trim và reject if contains internal whitespace
- Not empty: reject empty strings và nil values
- Rate limit: validate rate limit headers if test call made

**And** validation errors return structured error with specific reason
**And** error messages KHÔNG expose partial key values (security)
**And** validation happens at builder.Build() time, before any API calls
**And** test coverage: unit tests cho valid keys, invalid formats, edge cases

**Prerequisites:** None (can start in parallel with Epic 1)

**Technical Notes:**
- Package: `agent/validation/apikey.go`
- Function signature: `func ValidateAPIKey(provider string, key string) error`
- Use regex patterns for format validation
- Return custom error type: `APIKeyValidationError` với fields: Provider, Reason, Suggestion
- Provider-specific validators: `ValidateOpenAIKey()`, `ValidateGeminiKey()`
- Example error: "Invalid OpenAI API key format: must start with 'sk-' and be 48 characters"
- Performance: validation should be <1ms per key

---

### Story 2.2: Prompt Injection Protection Layer

**As a** Security Engineer,
**I want** prompt injection attack detection và sanitization,
**So that** user inputs không thể manipulate agent behavior maliciously.

**Acceptance Criteria:**

**Given** users provide prompts và messages to agents
**When** implement prompt injection protection
**Then** protection includes:
- Detect injection patterns: system role hijacking, instruction overrides, delimiter attacks
- Suspicious patterns: multiple "System:", "Ignore previous", "You are now", jailbreak attempts
- Content policy: detect requests to leak internal instructions, system prompts
- Length limits: enforce maximum prompt length (configurable, default 100K chars)
- Encoding validation: ensure UTF-8, reject invalid byte sequences
- Control character filtering: strip or escape control chars (\x00-\x1F except \n, \t)
- Sanitization modes: STRICT (reject suspicious), SANITIZE (remove patterns), PERMISSIVE (log only)

**And** configurable protection level: `agent.WithPromptProtection(level PromptProtectionLevel)`
**And** protected prompts logged to audit trail (configurable)
**And** metrics: count injection attempts, false positives
**And** allow-list: bypass protection for trusted input sources

**Prerequisites:** None

**Technical Notes:**
- Package: `agent/validation/prompt.go`
- Function: `func SanitizePrompt(input string, level PromptProtectionLevel) (string, error)`
- Detection patterns (regex):
  - System hijacking: `(?i)(you are now|ignore previous|system:|new instructions:)`
  - Delimiter attacks: `"""`, `---`, multiple backticks
  - Encoding attacks: `\u`, `%`, `\x` sequences outside normal usage
- Sanitization: remove matched patterns or reject entirely
- Performance: <10ms for 10K char prompts
- Reference: OWASP LLM01: Prompt Injection
- Test with: known jailbreak prompts, fuzzing, boundary cases

---

### Story 2.3: Tool Parameter Validation System

**As a** Tool Developer,
**I want** automatic parameter validation cho tool inputs,
**So that** tools receive type-safe, validated data và malicious inputs rejected.

**Acceptance Criteria:**

**Given** tools registered với parameter schemas
**When** agent calls tools với parameters
**Then** validation enforces:
- Type checking: string, int, float, bool, array, object match schema
- Required fields: all required parameters present
- Range validation: numbers within min/max bounds
- String constraints: length limits, regex patterns, allowed values
- Array constraints: min/max items, item type validation
- Object validation: nested property validation, additional properties rules
- Format validation: email, URL, file path, JSON, datetime formats
- No code injection: detect shell metacharacters, SQL fragments, script tags

**And** validation errors return: parameter name, expected type/format, actual value (sanitized), suggestion
**And** validation happens before tool execution
**And** schema definition: JSON Schema compatible format
**And** custom validators: support custom validation functions per parameter

**Prerequisites:** Story 2.1 (validation framework established)

**Technical Notes:**
- Package: `agent/tools/validation.go`
- Schema format: JSON Schema Draft 7 compatible
- Function: `func ValidateToolParams(schema ToolSchema, params map[string]interface{}) error`
- Type assertions: safe type conversion với error handling
- Injection detection patterns:
  - Shell: `;`, `|`, `&&`, `||`, `` ` ``, `$()`
  - SQL: `'; DROP`, `UNION SELECT`, `--`, `/**/`
  - XSS: `<script>`, `javascript:`, `onerror=`
- Performance: <1ms per parameter set
- Example schema:
```go
ToolSchema{
  Name: "fetch_url",
  Parameters: []Parameter{
    {Name: "url", Type: "string", Required: true, Format: "uri", MaxLength: 2048},
    {Name: "timeout", Type: "integer", Min: 1, Max: 300, Default: 30},
  },
}
```

---

### Story 2.4: Configuration Input Sanitization

**As a** DevOps Engineer,
**I want** configuration inputs sanitized và validated,
**So that** malicious configs không thể compromise agent security.

**Acceptance Criteria:**

**Given** configurations provided via env vars, files, hoặc programmatically
**When** load và validate configuration
**Then** sanitization covers:
- URLs: validate format, protocol whitelist (https only for production), no credentials in URL
- File paths: validate existence, check permissions, prevent path traversal (../, absolute paths)
- Regex patterns: compile and test regex, prevent ReDoS (catastrophic backtracking)
- Timeouts: enforce reasonable ranges (1ms - 10min), convert to time.Duration
- Rate limits: validate positive numbers, sensible ranges (1-10000 req/min)
- Memory limits: validate in MB, prevent overflow (1MB - 10GB)
- Connection strings: validate format, no SQL injection vectors
- JSON configs: validate JSON structure, schema compliance

**And** validation on startup: fail-fast if config invalid
**And** validation errors: clear message với fix suggestions
**And** default values: secure defaults for all optional configs
**And** config reloading: revalidate on hot-reload

**Prerequisites:** Story 2.1 (validation framework)

**Technical Notes:**
- Package: `agent/config/validation.go`
- Function: `func ValidateConfig(cfg *Config) error`
- URL validation: net/url.Parse() + whitelist check
- Path validation: filepath.Clean(), filepath.Abs(), os.Stat()
- Regex validation: regexp.Compile() + ReDoS detection (linear time guarantee)
- Path traversal prevention: reject paths with `..`, check against base directory
- Example ReDoS: `(a+)+b` - use atomic grouping `(?>a+)+b` or possessive `a++b`
- Performance: full config validation <5ms
- Test coverage: valid configs, invalid formats, edge cases, attack vectors

---

## Epic 3: Secure Defaults & Authentication

**Epic Goal:** Implement production-grade security defaults và authentication mechanisms including TLS enforcement, secrets protection, API key rotation, credential validation, access control, và audit logging.

**Business Value:** Enterprise-ready security, compliance-ready, production deployment confidence

**FR Coverage:** FR7 (TLS 1.2+), FR8 (secrets protection), FR9 (key rotation), FR10 (credential validation), FR11 (access control), FR12 (audit logging)

---

### Story 3.1: TLS 1.2+ Enforcement for All External Communications

**As a** Security Engineer,
**I want** TLS 1.2+ enforced cho all HTTP/HTTPS communications,
**So that** data in transit được encrypted và secure against downgrade attacks.

**Acceptance Criteria:**

**Given** agent making external HTTP requests to providers và other services
**When** configure HTTP client
**Then** TLS configuration enforces:
- Minimum TLS version: TLS 1.2 (reject TLS 1.0, 1.1)
- Preferred TLS version: TLS 1.3 where available
- Certificate validation: enabled, no InsecureSkipVerify
- Cipher suites: strong ciphers only (AES-GCM, ChaCha20-Poly1305)
- HTTPS-only: reject HTTP for production (configurable for local dev)
- Certificate pinning: optional for high-security deployments

**And** HTTP client configured globally: `agent.DefaultHTTPClient` với secure TLS
**And** Provider adapters use secure client by default
**And** Error on TLS handshake failures với diagnostic info
**And** Test coverage: TLS version validation, certificate validation, downgrade prevention

**Prerequisites:** Story 2.4 (config validation - URL validation)

**Technical Notes:**
```go
tlsConfig := &tls.Config{
    MinVersion:               tls.VersionTLS12,
    PreferServerCipherSuites: true,
    CipherSuites: []uint16{
        tls.TLS_AES_128_GCM_SHA256,
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_CHACHA20_POLY1305_SHA256,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
    },
    InsecureSkipVerify: false,
}
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: tlsConfig,
        MaxIdleConns: 100,
        IdleConnTimeout: 90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```
- Package: `agent/security/tls.go`
- Function: `func NewSecureHTTPClient(cfg *TLSConfig) *http.Client`

---

### Story 3.2: Secrets Protection - No Logging of Sensitive Data

**As a** Security Engineer,
**I want** API keys và secrets never logged or exposed,
**So that** credentials cannot be leaked through logs hoặc error messages.

**Acceptance Criteria:**

**Given** library handles API keys, secrets, tokens
**When** logging hoặc error handling occurs
**Then** sensitive data protection includes:
- API keys: never logged, redacted in errors (`sk-***...***`)
- Passwords: never logged or displayed
- Tokens: redacted in logs và error messages
- User prompts: optionally excluded from logs (PII protection)
- Provider responses: sanitized before logging (remove keys from JSON)
- Stack traces: sanitize before output
- Panic recovery: clean up secrets from memory

**And** Logging framework: structured logging với field filtering
**And** Error messages: generic without exposing credentials
**And** Debug mode: même in debug, secrets redacted
**And** Memory cleanup: defer cleanup of sensitive data on errors

**Prerequisites:** None

**Technical Notes:**
- Package: `agent/security/secrets.go`
- Redaction function: `func RedactSensitive(s string) string`
- Patterns to redact:
  - API keys: `sk-[A-Za-z0-9]{48}` → `sk-***...***`
  - Tokens: `Bearer [A-Za-z0-9+/=]+` → `Bearer ***`
  - Passwords in URLs: `://user:pass@host` → `://user:***@host`
- Structured logging: use `slog` with custom RedactHandler
- Example:
```go
type RedactHandler struct {
    handler slog.Handler
}
func (h *RedactHandler) Handle(ctx context.Context, r slog.Record) error {
    r.Message = RedactSensitive(r.Message)
    return h.handler.Handle(ctx, r)
}
```

---

### Story 3.3: API Key Rotation Support Without Service Interruption

**As a** DevOps Engineer,
**I want** API key rotation without stopping agents,
**So that** security best practices (regular key rotation) không impact availability.

**Acceptance Criteria:**

**Given** agent running với current API key
**When** rotate API key
**Then** rotation mechanism supports:
- Dual-key period: both old và new keys valid during transition
- Graceful switchover: complete in-flight requests với old key
- Atomic swap: new key becomes primary after validation
- Validation: test new key before switching
- Fallback: revert to old key if new key fails
- Configuration reload: hot-reload key from env vars hoặc config file
- Zero downtime: no dropped requests during rotation

**And** Rotation API: `agent.RotateAPIKey(newKey string) error`
**And** Validation before swap: make test call với new key
**And** Event logging: audit log of key rotation events
**And** Thread-safe: rotation safe during concurrent requests

**Prerequisites:** Story 2.1 (API key validation)

**Technical Notes:**
- Package: `agent/security/rotation.go`
- Implementation: atomic.Value for thread-safe key swap
```go
type KeyManager struct {
    currentKey atomic.Value // stores *APIKey
    oldKey     atomic.Value
    mu         sync.RWMutex
}
func (km *KeyManager) RotateKey(newKey string) error {
    // 1. Validate new key format
    if err := ValidateAPIKey(newKey); err != nil {
        return err
    }
    // 2. Test new key with provider
    if err := km.testKey(newKey); err != nil {
        return err
    }
    // 3. Atomic swap
    km.mu.Lock()
    old := km.currentKey.Load()
    km.oldKey.Store(old)
    km.currentKey.Store(&APIKey{Value: newKey, ValidUntil: time.Now().Add(5*time.Minute)})
    km.mu.Unlock()
    // 4. Audit log
    km.auditLog("key_rotated", newKey[:8]+"...")
    return nil
}
```

---

### Story 3.4: Provider Credential Validation Before First Use

**As a** Developer,
**I want** credentials validated before making first API call,
**So that** misconfiguration được caught early với clear error messages.

**Acceptance Criteria:**

**Given** user provides API credentials for provider
**When** build agent hoặc initialize provider adapter
**Then** validation performs:
- Format check: API key matches expected pattern (Story 2.1)
- Test call: minimal API call to validate credentials (e.g., list models)
- Rate limit check: validate rate limit headers
- Quota check: verify account has quota available
- Permission check: verify key has required permissions (e.g., model access)
- Error handling: clear message if validation fails with fix suggestions
- Timeout: validation completes within 5 seconds or fails

**And** Validation on agent build: `builder.Build()` validates before returning agent
**And** Caching: cache validation result (valid for 1 hour) to avoid repeated calls
**And** Fail-fast: don't create agent if credentials invalid
**And** Error message includes: reason, provider docs link, suggested fix

**Prerequisites:** Story 2.1 (API key validation), Story 3.1 (TLS enforcement)

**Technical Notes:**
- Package: `agent/adapters/validation.go`
- Function: `func (a *Adapter) ValidateCredentials(ctx context.Context) error`
- OpenAI validation: `GET https://api.openai.com/v1/models`
- Gemini validation: `GET https://generativelanguage.googleapis.com/v1/models`
- Cache validation: `sync.Map` với TTL
- Example error:
```
Invalid OpenAI API key: authentication failed (401)

Possible reasons:
1. API key is incorrect or expired
2. API key doesn't have access to requested models
3. Account quota exceeded

Fix: Check your API key at https://platform.openai.com/api-keys
```

---

### Story 3.5: Scope-Based Access Control for Tool Execution

**As a** Security Engineer,
**I want** fine-grained access control cho tool execution,
**So that** agents chỉ có thể execute authorized tools preventing privilege escalation.

**Acceptance Criteria:**

**Given** agent configured với tools
**When** agent attempts to execute tool
**Then** access control enforces:
- Tool allowlist: only explicitly allowed tools can execute
- Tool denylist: explicitly denied tools rejected
- Scope-based permissions: tools categorized by risk level (safe, restricted, dangerous)
- Permission levels: READ, WRITE, EXECUTE, ADMIN
- Contextual permissions: permissions vary by environment (dev vs production)
- Default deny: unknown tools rejected by default
- Permission validation: check permissions before tool execution

**And** Configuration: `agent.WithToolPermissions(permissions ToolPermissions)`
**And** Permission definitions:
```go
ToolPermissions{
    AllowedTools: []string{"fetch_url", "read_file"},
    DeniedTools: []string{"execute_shell", "delete_file"},
    DefaultPolicy: DENY,
    Scopes: map[string]Scope{
        "fetch_url": {Level: READ, MaxRate: 100},
        "write_file": {Level: WRITE, Paths: []string{"/tmp/*"}},
    },
}
```
**And** Audit logging: log all tool execution attempts với permissions check result

**Prerequisites:** Story 2.3 (tool parameter validation)

**Technical Notes:**
- Package: `agent/tools/permissions.go`
- Permission check before execution: `func CheckToolPermission(toolName string, perms ToolPermissions) error`
- Path-based restrictions cho filesystem tools
- Rate limiting per tool
- Environment-based config: stricter in production

---

### Story 3.6: Comprehensive Audit Logging for Sensitive Operations

**As a** Compliance Officer,
**I want** audit trail of all sensitive operations,
**So that** security incidents có thể investigated và compliance requirements met.

**Acceptance Criteria:**

**Given** agent performing sensitive operations
**When** operations execute
**Then** audit logging captures:
- API key operations: validation, rotation, test calls
- Tool executions: tool name, parameters (sanitized), result, permissions check
- Provider calls: provider, model, token counts, latency, errors
- Configuration changes: what changed, who changed, when
- Authentication events: credential validation success/failure
- Access control decisions: permission grants/denials
- Security events: suspicious activity, injection attempts, rate limit hits

**And** Audit log format: structured JSON với:
```json
{
  "timestamp": "2025-01-14T10:30:00Z",
  "event_type": "tool_execution",
  "tool_name": "fetch_url",
  "user_id": "agent-123",
  "result": "success",
  "parameters": {"url": "https://example.com"},
  "permissions_check": "allowed",
  "correlation_id": "req-abc-123"
}
```
**And** Configurable destinations: file, stdout, syslog, external service
**And** Rotation: log rotation configured (daily, size-based)
**And** Retention: retention policy enforced (90 days default)
**And** Performance: async logging, không block operations

**Prerequisites:** Story 3.2 (secrets protection - sanitization), Story 3.5 (access control)

**Technical Notes:**
- Package: `agent/security/audit.go`
- Interface:
```go
type AuditLogger interface {
    LogEvent(ctx context.Context, event AuditEvent) error
}
type AuditEvent struct {
    Timestamp     time.Time
    EventType     string
    Actor         string
    Resource      string
    Action        string
    Result        string
    Details       map[string]interface{}
    CorrelationID string
}
```
- Async logging: buffered channel với background worker
- Sanitization: apply RedactSensitive before logging
- Integration: pluggable backends (file, syslog, Datadog, Splunk)

---

## Epic 4: Benchmark Suite Development

**Epic Goal:** Create comprehensive benchmarking infrastructure for measuring performance across all critical operations (agent creation, tool execution, memory, RAG, provider adapters, batch processing).

**Business Value:** Performance visibility, regression prevention, optimization guidance

**FR Coverage:** FR13 (agent benchmarks), FR14 (tool benchmarks), FR15 (memory benchmarks), FR16 (RAG benchmarks), FR17 (provider benchmarks), FR18 (batch benchmarks)

---

### Story 4.1: Agent Creation Performance Benchmarks

**As a** Performance Engineer,
**I want** benchmarks cho agent creation operations,
**So that** framework overhead được measured và optimized (target <1ms).

**Acceptance Criteria:**

**Given** various agent configurations
**When** run agent creation benchmarks
**Then** benchmarks measure:
- Simple agent: `agent.New().Build()` với minimal config (<1ms target)
- With tools: agent + 10 registered tools (<2ms target)
- With memory: agent + episodic + semantic + working memory (<5ms target)
- With RAG: agent + RAG system với 1K documents (<10ms target)
- Full-featured: agent với all features enabled (<15ms target)
- Builder overhead: measure builder pattern overhead specifically
- Memory allocations: track heap allocations per operation

**And** Benchmark results include: ns/op, B/op, allocs/op
**And** Comparison baseline: store baseline results for comparison
**And** Run command: `go test -bench=BenchmarkAgentCreation -benchmem ./...`
**And** CI integration: benchmarks run on every PR

**Prerequisites:** None (foundation for Epic 4)

**Technical Notes:**
- Package: `agent/benchmarks/agent_test.go`
- Benchmark functions:
```go
func BenchmarkAgentCreation_Simple(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        agent, _ := agent.New().
            WithProvider(provider.NewMock()).
            Build()
        _ = agent
    }
}
func BenchmarkAgentCreation_WithTools(b *testing.B) {
    b.ReportAllocs()
    tools := createMockTools(10)
    for i := 0; i < b.N; i++ {
        agent, _ := agent.New().
            WithProvider(provider.NewMock()).
            WithTools(tools...).
            Build()
        _ = agent
    }
}
```
- Use `b.ResetTimer()` to exclude setup time
- Use `b.StopTimer()`/`b.StartTimer()` for precise measurement
- Store baselines: `go test -bench=. -benchmem > baseline.txt`

---

### Story 4.2: Tool Execution Framework Overhead Benchmarks

**As a** Performance Engineer,
**I want** benchmarks cho tool execution framework overhead,
**So that** tool dispatch mechanism optimized (target <100μs).

**Acceptance Criteria:**

**Given** registered tools với varying complexity
**When** run tool execution benchmarks
**Then** benchmarks measure:
- Tool dispatch: framework overhead for calling tool (<100μs target)
- Parameter validation: validation time per parameter set (<50μs target)
- Permission check: access control check time (<10μs target)
- Result serialization: JSON serialization of results (<50μs target)
- Error handling: error wrapping và propagation overhead (<20μs target)
- Async execution: goroutine spawn và channel overhead (<1ms target)
- Tool with caching: cache hit và miss performance

**And** Separate measurement: actual tool execution vs framework overhead
**And** Various tool types: simple (math), I/O (file), network (HTTP), complex (LLM call)
**And** Benchmark results guide optimization: identify hot paths

**Prerequisites:** Story 4.1 (benchmark infrastructure)

**Technical Notes:**
```go
func BenchmarkToolExecution_Dispatch(b *testing.B) {
    b.ReportAllocs()
    tool := &Tool{
        Name: "echo",
        Func: func(input string) string { return input },
    }
    executor := NewToolExecutor()
    executor.RegisterTool(tool)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = executor.Execute("echo", map[string]interface{}{"input": "test"})
    }
}
```
- Measure components separately: dispatch, validation, execution, serialization
- Use micro-benchmarks: focus on specific operations
- Exclude I/O: use mocks for network/filesystem to measure framework only

---

### Story 4.3: Memory System Performance Benchmarks

**As a** Performance Engineer,
**I want** benchmarks cho memory operations,
**So that** memory systems (episodic, semantic, working) meet <10ms latency target.

**Acceptance Criteria:**

**Given** memory systems với varying data sizes
**When** run memory benchmarks
**Then** benchmarks measure:
- Episodic memory: add message, retrieve by index (<5ms target for 1K messages)
- Semantic memory: store knowledge, similarity search (<10ms target for 10K entries)
- Working memory: update context, retrieve current state (<1ms target)
- Memory persistence: save to disk, load from disk (measure throughput)
- Memory cleanup: garbage collection performance (<50ms target)
- Memory retrieval: similarity search latency vs corpus size
- Concurrent access: thread-safe operations performance impact

**And** Scale testing: test with 1K, 10K, 100K, 1M entries
**And** Operations: insert, update, retrieve, search, delete
**And** Memory backends: in-memory, persistent (file), Redis

**Prerequisites:** Story 4.1 (benchmark infrastructure)

**Technical Notes:**
```go
func BenchmarkMemory_Episodic_Add(b *testing.B) {
    mem := memory.NewEpisodic()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mem.Add(Message{Role: "user", Content: "test"})
    }
}
func BenchmarkMemory_Semantic_Search(b *testing.B) {
    mem := memory.NewSemantic()
    // Pre-populate với 10K entries
    for i := 0; i < 10000; i++ {
        mem.Store(fmt.Sprintf("knowledge-%d", i), embedding)
    }
    query := generateQueryEmbedding()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = mem.Search(query, 10) // top 10 results
    }
}
```
- Measure separately: write operations vs read operations
- Test scaling: plot latency vs corpus size
- Memory profiling: track memory usage growth

---

### Story 4.4: RAG Vector Search Performance Benchmarks

**As a** Performance Engineer,
**I want** benchmarks cho RAG và vector search,
**So that** search latency meets targets (<50ms for 10K docs, <200ms for 100K docs).

**Acceptance Criteria:**

**Given** RAG system với document corpus
**When** run RAG benchmarks
**Then** benchmarks measure:
- Document ingestion: indexing throughput (docs/second)
- Embedding generation: time per document (batch vs individual)
- Vector search: query latency vs corpus size (10K, 100K, 1M docs)
- Hybrid search: semantic + keyword search performance
- Metadata filtering: search với filters overhead
- Index updates: incremental update performance
- Cache performance: embedding cache hit rate impact

**And** Corpus sizes: 1K, 10K, 100K, 1M documents
**And** Vector dimensions: 384 (small), 768 (medium), 1536 (large - OpenAI)
**And** Search parameters: top-k (1, 10, 100), similarity thresholds
**And** Backends: in-memory, Qdrant, custom

**Prerequisites:** Story 4.1 (benchmark infrastructure)

**Technical Notes:**
```go
func BenchmarkRAG_VectorSearch_10K(b *testing.B) {
    rag := setupRAGWith10KDocs()
    query := "test query"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = rag.Search(query, 10)
    }
}
func BenchmarkRAG_Ingestion(b *testing.B) {
    rag := rag.New()
    docs := generateDocuments(b.N)
    b.ResetTimer()
    for _, doc := range docs {
        rag.Ingest(doc)
    }
    b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "docs/sec")
}
```
- Measure components: chunking, embedding, indexing, search separately
- Plot scaling curves: latency vs corpus size
- Comparison: different vector stores, algorithms

---

### Story 4.5: Provider Adapter Performance Benchmarks

**As a** Performance Engineer,
**I want** benchmarks cho provider adapter overhead,
**So that** framework overhead identified separately from API latency.

**Acceptance Criteria:**

**Given** provider adapters (OpenAI, Gemini, mocks)
**When** run provider benchmarks
**Then** benchmarks measure:
- Request preparation: time to build provider-specific request (<1ms target)
- Response parsing: time to parse provider response (<5ms target)
- Adapter overhead: framework overhead excluding network I/O (<2ms total target)
- Error handling: error transformation performance (<1ms target)
- Retry logic: retry decision và backoff calculation (<100μs target)
- Streaming: overhead per streamed chunk (<500μs target)
- Connection pooling: connection reuse impact

**And** Use mocks: exclude network latency, measure adapter only
**And** Various request sizes: small (100 tokens), medium (1K tokens), large (10K tokens)
**And** Streaming vs non-streaming: compare overhead

**Prerequisites:** Story 4.1 (benchmark infrastructure)

**Technical Notes:**
```go
func BenchmarkProvider_RequestPreparation(b *testing.B) {
    adapter := provider.NewOpenAI("mock-key")
    messages := []Message{{Role: "user", Content: "test"}}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = adapter.prepareRequest(messages, nil)
    }
}
func BenchmarkProvider_ResponseParsing(b *testing.B) {
    adapter := provider.NewOpenAI("mock-key")
    rawResponse := `{"choices":[{"message":{"content":"test"}}]}`
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = adapter.parseResponse([]byte(rawResponse))
    }
}
```
- Mock HTTP client: return fixture data immediately
- Measure JSON marshaling/unmarshaling separately
- Track allocations: minimize allocations in hot path

---

### Story 4.6: Batch Processing Throughput Benchmarks

**As a** Performance Engineer,
**I want** benchmarks cho batch processing,
**So that** throughput target (>100 ops/sec) validated và optimized.

**Acceptance Criteria:**

**Given** batch processing system
**When** run batch benchmarks
**Then** benchmarks measure:
- Sequential processing: baseline throughput (ops/sec)
- Concurrent processing: throughput với goroutine pooling (target >100 ops/sec)
- Batch sizes: throughput vs batch size (1, 10, 100, 1000)
- Rate limiting: throughput với rate limits applied
- Error handling: impact of error rate on throughput
- Memory usage: memory per concurrent operation
- Backpressure: behavior when queue full

**And** Metrics: ops/sec, latency percentiles (p50, p95, p99), memory usage
**And** Workload types: CPU-bound, I/O-bound, mixed
**And** Concurrency levels: 1, 10, 100, 1000 goroutines

**Prerequisites:** Story 4.1 (benchmark infrastructure)

**Technical Notes:**
```go
func BenchmarkBatch_Concurrent(b *testing.B) {
    processor := batch.NewProcessor(batch.Config{
        Workers: 100,
        QueueSize: 1000,
    })
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            processor.Submit(mockTask())
        }
    })
    processor.Wait()
    throughput := float64(b.N) / b.Elapsed().Seconds()
    b.ReportMetric(throughput, "ops/sec")
}
```
- Use `b.RunParallel()` for concurrent benchmarks
- Measure bottlenecks: goroutine spawning, channel operations, synchronization
- Profile: CPU và memory profiles for optimization

---

## Epic 5: Performance Baselines & CI Integration

**Epic Goal:** Establish performance baselines và integrate benchmark regression detection vào CI/CD pipeline, ensuring performance không degrade over time.

**Business Value:** Automated performance monitoring, early detection of regressions, performance confidence

**FR Coverage:** FR19 (baseline tracking), FR20 (regression detection), FR21 (performance reports)

---

### Story 5.1: Establish Performance Baselines Using benchstat

**As a** Performance Engineer,
**I want** establish current performance baselines cho all benchmarks,
**So that** future changes có thể compared against known-good baseline.

**Acceptance Criteria:**

**Given** comprehensive benchmark suite từ Epic 4 (Stories 4.1-4.6)
**When** establish performance baselines
**Then** baseline process includes:
- Run all benchmarks 10 times: get statistically significant results
- Use benchstat tool: analyze benchmark results, calculate mean/median/stddev
- Store baseline data: JSON format với metadata (Go version, OS, CPU, date)
- Baseline coverage: all benchmark functions in project
- Baseline file location: `benchmarks/baseline-{version}.json`
- Version tagging: tag baseline with Go version và library version
- Hardware specs: capture CPU model, cores, RAM in baseline metadata

**And** Baseline report includes:
```json
{
  "version": "v0.11.0",
  "date": "2025-11-14",
  "go_version": "1.21.5",
  "os": "linux",
  "arch": "amd64",
  "cpu": "Intel Xeon E5-2686 v4",
  "benchmarks": {
    "BenchmarkAgentCreation_Simple": {
      "ns_per_op": 850000,
      "bytes_per_op": 12456,
      "allocs_per_op": 145,
      "runs": 10,
      "stddev": 25000
    }
  }
}
```

**And** Comparison command: `benchstat baseline.txt current.txt`
**And** Documentation: how to run baselines, interpret results

**Prerequisites:** Story 4.1-4.6 (all benchmarks created)

**Technical Notes:**
- Install benchstat: `go install golang.org/x/perf/cmd/benchstat@latest`
- Run benchmarks: `go test -bench=. -benchmem -count=10 ./... > baseline.txt`
- Analyze: `benchstat baseline.txt`
- Store in version control: `benchmarks/baselines/` directory
- Baseline frequency: establish new baseline for each release
- Hardware consistency: use same hardware for comparison (CI runners)

---

### Story 5.2: Automated Performance Regression Detection in CI

**As a** DevOps Engineer,
**I want** automated benchmark regression detection on every PR,
**So that** performance regressions được caught before merge.

**Acceptance Criteria:**

**Given** established performance baselines
**When** PR submitted với code changes
**Then** CI pipeline performs:
- Run all benchmarks: same conditions as baseline
- Compare against baseline: using benchstat for statistical comparison
- Regression threshold: fail if >10% slowdown in any benchmark
- Memory regression: fail if >15% increase in allocations
- Generate comparison report: show before/after với percentage changes
- Comment on PR: post regression report as PR comment
- Status check: block merge if regressions detected
- Allow overrides: maintainer can approve despite regression với justification

**And** CI workflow: `.github/workflows/benchmark-regression.yml`
**And** Comparison output example:
```
name                              old time/op    new time/op    delta
BenchmarkAgentCreation_Simple-8     850µs ± 2%    1050µs ± 3%  +23.53%  (p=0.000 n=10+10) ⚠️ REGRESSION
BenchmarkToolExecution_Dispatch-8   95.0µs ± 1%    94.5µs ± 2%     ~     (p=0.234 n=10+10) ✅ OK

name                              old alloc/op   new alloc/op   delta
BenchmarkAgentCreation_Simple-8    12.5kB ± 0%    13.2kB ± 0%   +5.60%  (p=0.000 n=10+10) ✅ OK
```

**Prerequisites:** Story 5.1 (baselines established)

**Technical Notes:**
```yaml
name: Benchmark Regression Check
on:
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Need history for baseline

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run Benchmarks
        run: |
          go test -bench=. -benchmem -count=10 ./... > new.txt

      - name: Get Baseline
        run: |
          git checkout main
          go test -bench=. -benchmem -count=10 ./... > baseline.txt
          git checkout -

      - name: Compare with benchstat
        run: |
          go install golang.org/x/perf/cmd/benchstat@latest
          benchstat baseline.txt new.txt > comparison.txt

      - name: Check for Regressions
        run: |
          # Parse comparison.txt, fail if >10% regression
          python scripts/check_regression.py comparison.txt

      - name: Comment PR
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const comparison = fs.readFileSync('comparison.txt', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## Benchmark Comparison\n\n```\n' + comparison + '\n```'
            });
```

---

### Story 5.3: Performance Comparison Reports & Visualization

**As a** Engineering Manager,
**I want** visual performance comparison reports,
**So that** performance trends tracked over time và regressions easily spotted.

**Acceptance Criteria:**

**Given** benchmark results from multiple runs/versions
**When** generate performance reports
**Then** reporting system provides:
- Trend charts: performance over time (last 10 releases)
- Comparison tables: current vs baseline với delta percentages
- Regression highlights: visual indicators for regressions (red) và improvements (green)
- Historical data: store benchmark results for all releases
- Export formats: JSON, CSV, HTML report
- Dashboard: optional web dashboard showing trends
- Alerts: configurable alerts for significant regressions (>20%)

**And** Report includes:
- Performance summary: overall pass/fail status
- Top regressions: worst 5 regressions by percentage
- Top improvements: best 5 improvements
- Memory trends: allocation trends over time
- Operation breakdown: performance by component (agent, tools, memory, RAG)

**And** Visualization: charts using Chart.js or similar
**And** Storage: results stored in `benchmarks/results/` directory
**And** Retention: keep 90 days of results, archive older

**Prerequisites:** Story 5.2 (CI integration)

**Technical Notes:**
- Package: `benchmarks/reporting/`
- Script: `scripts/generate_benchmark_report.py`
- Data storage: JSON files per run: `results/{date}-{commit}.json`
- Visualization library: Chart.js for web dashboard
- Report generation:
```go
type BenchmarkResult struct {
    Timestamp   time.Time
    Version     string
    CommitHash  string
    Benchmarks  map[string]BenchmarkMetrics
}

type BenchmarkMetrics struct {
    NsPerOp     float64
    BytesPerOp  int64
    AllocsPerOp int64
    Runs        int
}

func GenerateReport(results []BenchmarkResult) *Report {
    // Generate HTML report with charts
}
```
- Optional: integrate với GitHub Pages để host dashboard

---

## Epic 6: Performance Optimization Implementation

**Epic Goal:** Optimize library performance targeting minimal framework overhead: <1ms agent creation, <100μs tool dispatch, <10ms memory ops, cache hit rate >80%.

**Business Value:** Production-grade performance, competitive advantage, developer trust

**FR Coverage:** FR22 (memory allocation), FR23 (goroutine pooling), FR24 (caching), FR25 (connection pooling), FR26 (lazy loading)

---

### Story 6.1: Memory Allocation Optimization & GC Pressure Reduction

**As a** Performance Engineer,
**I want** minimize heap allocations trong hot paths,
**So that** GC pressure reduced và latency improved.

**Acceptance Criteria:**

**Given** benchmark data showing allocation hotspots
**When** optimize memory allocations
**Then** optimization includes:
- Object pooling: sync.Pool for frequently allocated objects (messages, contexts)
- Pre-allocation: pre-allocate slices với capacity hints
- String building: use strings.Builder instead of string concatenation
- Avoid boxing: use concrete types instead of interface{} where possible
- Reuse buffers: buffer pools for encoding/decoding
- Stack allocation: ensure small objects stay on stack (escape analysis)
- Benchmark validation: measure allocs/op before và after

**And** Allocation reduction targets:
- Agent creation: <10MB heap allocations (vs current baseline)
- Tool execution: <1KB allocations per call
- Message processing: <5KB per message
- Overall: 30-50% reduction in allocations

**And** Profiling: use `go test -bench=. -benchmem -memprofile=mem.prof`
**And** Analysis: `go tool pprof mem.prof` to identify hotspots
**And** Validation: no increase in GC pauses (measure with GODEBUG=gctrace=1)

**Prerequisites:** Story 5.1 (baselines for comparison)

**Technical Notes:**
```go
// Object pooling example
var messagePool = sync.Pool{
    New: func() interface{} {
        return &Message{}
    },
}

func GetMessage() *Message {
    return messagePool.Get().(*Message)
}

func PutMessage(m *Message) {
    m.Reset()  // Clear for reuse
    messagePool.Put(m)
}

// Pre-allocation example
func processMessages(count int) []Message {
    // Pre-allocate với capacity
    messages := make([]Message, 0, count)
    for i := 0; i < count; i++ {
        messages = append(messages, Message{...})
    }
    return messages
}

// String building
func buildPrompt(parts []string) string {
    var b strings.Builder
    b.Grow(estimateSize(parts))  // Pre-allocate
    for _, part := range parts {
        b.WriteString(part)
    }
    return b.String()
}
```
- Escape analysis: `go build -gcflags='-m'` to check allocations
- Benchmark comparison: use benchstat to validate improvements

---

### Story 6.2: Goroutine Pooling for Concurrent Operations

**As a** Performance Engineer,
**I want** bounded goroutine pooling cho concurrent operations,
**So that** resource usage controlled và throughput maximized.

**Acceptance Criteria:**

**Given** concurrent operations (batch processing, tool execution)
**When** implement goroutine pooling
**Then** pool implementation includes:
- Worker pool: fixed number of goroutines (default: NumCPU * 2)
- Task queue: buffered channel for pending tasks
- Graceful shutdown: wait for in-flight tasks on context cancellation
- Configurable size: pool size tunable via configuration
- Backpressure: reject or queue when pool saturated
- Metrics: track pool utilization, queue depth, task latency
- Panic recovery: recover from panics in worker goroutines

**And** Pool configuration:
```go
type PoolConfig struct {
    Workers      int           // Number of worker goroutines
    QueueSize    int           // Task queue buffer size
    TaskTimeout  time.Duration // Max time per task
    IdleTimeout  time.Duration // Worker idle timeout
}
```

**And** Performance targets:
- Goroutine overhead: <5KB per goroutine
- Task dispatch latency: <100μs
- Throughput: >100 tasks/sec per worker
- Pool startup: <10ms

**And** Use cases: batch processing, concurrent tool execution, parallel RAG queries

**Prerequisites:** Story 6.1 (allocation optimization - pools share patterns)

**Technical Notes:**
```go
type WorkerPool struct {
    workers   int
    tasks     chan Task
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
}

type Task func(context.Context) error

func NewWorkerPool(config PoolConfig) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    pool := &WorkerPool{
        workers: config.Workers,
        tasks:   make(chan Task, config.QueueSize),
        ctx:     ctx,
        cancel:  cancel,
    }

    for i := 0; i < pool.workers; i++ {
        pool.wg.Add(1)
        go pool.worker(i)
    }

    return pool
}

func (p *WorkerPool) worker(id int) {
    defer p.wg.Done()
    for {
        select {
        case <-p.ctx.Done():
            return
        case task := <-p.tasks:
            if err := task(p.ctx); err != nil {
                // Log error
            }
        }
    }
}

func (p *WorkerPool) Submit(task Task) error {
    select {
    case p.tasks <- task:
        return nil
    case <-p.ctx.Done():
        return ErrPoolClosed
    default:
        return ErrPoolFull  // Backpressure
    }
}

func (p *WorkerPool) Shutdown() {
    p.cancel()
    p.wg.Wait()
}
```
- Benchmark: compare với unbounded goroutines
- Libraries: consider github.com/panjf2000/ants or custom implementation

---

### Story 6.3: Intelligent Caching with Configurable Hit Rate Targets

**As a** Performance Engineer,
**I want** intelligent caching với >80% hit rate target,
**So that** repeated operations served from cache reducing latency.

**Acceptance Criteria:**

**Given** repeated LLM calls, embeddings, và tool results
**When** implement caching system
**Then** cache features include:
- Multi-level caching: L1 (in-memory), L2 (Redis optional)
- TTL support: configurable expiration per cache entry
- Eviction policies: LRU, LFU, TTL-based
- Cache key generation: deterministic keys from inputs
- Hit rate tracking: measure và report cache effectiveness
- Cache warming: pre-populate frequently used entries
- Partial cache: cache embeddings, responses separately
- Cache invalidation: pattern-based invalidation

**And** Cache configuration:
```go
type CacheConfig struct {
    MaxSize      int           // Max entries (LRU)
    TTL          time.Duration // Default TTL
    HitRateTarget float64      // Target hit rate (0.8 = 80%)
    RedisAddr    string        // Optional Redis backend
}
```

**And** Cache hit rate targets:
- LLM responses: >70% hit rate (for repeated queries)
- Embeddings: >90% hit rate (documents rarely change)
- Tool results: >60% hit rate (depends on tool)
- Overall: >80% hit rate

**And** Metrics:
- Hit rate: hits / (hits + misses)
- Miss latency: time saved by cache hits
- Memory usage: track cache size
- Eviction rate: entries evicted per minute

**Prerequisites:** Story 6.1 (memory optimization)

**Technical Notes:**
```go
type Cache struct {
    store    map[string]*CacheEntry
    lru      *list.List
    mu       sync.RWMutex
    maxSize  int
    hits     atomic.Uint64
    misses   atomic.Uint64
}

type CacheEntry struct {
    Key       string
    Value     interface{}
    ExpiresAt time.Time
    element   *list.Element
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    entry, ok := c.store[key]
    c.mu.RUnlock()

    if !ok || entry.ExpiresAt.Before(time.Now()) {
        c.misses.Add(1)
        return nil, false
    }

    c.hits.Add(1)
    c.mu.Lock()
    c.lru.MoveToFront(entry.element)
    c.mu.Unlock()

    return entry.Value, true
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if len(c.store) >= c.maxSize {
        c.evictLRU()
    }

    entry := &CacheEntry{
        Key:       key,
        Value:     value,
        ExpiresAt: time.Now().Add(ttl),
    }
    entry.element = c.lru.PushFront(entry)
    c.store[key] = entry
}

func (c *Cache) HitRate() float64 {
    hits := c.hits.Load()
    misses := c.misses.Load()
    if hits+misses == 0 {
        return 0
    }
    return float64(hits) / float64(hits+misses)
}
```
- Library consideration: groupcache, bigcache, or custom
- Redis integration: use go-redis for distributed cache

---

### Story 6.4: HTTP Connection Pooling for Provider APIs

**As a** Performance Engineer,
**I want** connection pooling cho provider API calls,
**So that** connection overhead minimized và latency reduced.

**Acceptance Criteria:**

**Given** frequent API calls to OpenAI, Gemini
**When** configure HTTP client với connection pooling
**Then** configuration includes:
- HTTP keep-alive: enable persistent connections
- Connection pool size: max idle connections per host (100)
- Idle connection timeout: 90 seconds
- Max idle connections total: 100
- Connection reuse: measure reuse rate (target >80%)
- TLS session resumption: reduce TLS handshake overhead
- DNS caching: cache DNS lookups (5 minutes TTL)
- Timeout configuration: dial timeout, request timeout

**And** HTTP client configuration:
```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    TLSHandshakeTimeout: 10 * time.Second,
    DisableKeepAlives:   false,

    // Connection pooling
    DialContext: (&net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
        DualStack: true,
    }).DialContext,

    // TLS config from Story 3.1
    TLSClientConfig: secureConfig,
}

client := &http.Client{
    Transport: transport,
    Timeout:   30 * time.Second,
}
```

**And** Performance improvements:
- Connection establishment: <10ms (vs 100ms+ for new connection)
- TLS handshake: avoided for pooled connections
- Overall latency: 20-50% reduction for repeated calls

**And** Monitoring:
- Connection reuse rate: track via custom RoundTripper
- Pool saturation: alert if max connections hit frequently

**Prerequisites:** Story 3.1 (TLS configuration)

**Technical Notes:**
- Use net/http/httptrace for connection monitoring
- Metrics collection:
```go
type ConnMetrics struct {
    mu             sync.Mutex
    connectCount   int64
    reuseCount     int64
}

func (m *ConnMetrics) ReuseRate() float64 {
    m.mu.Lock()
    defer m.mu.Unlock()
    total := m.connectCount + m.reuseCount
    if total == 0 {
        return 0
    }
    return float64(m.reuseCount) / float64(total)
}
```

---

### Story 6.5: Lazy Loading for Expensive Initializations

**As a** Performance Engineer,
**I want** lazy loading cho expensive resources,
**So that** startup time minimized và resources loaded only when needed.

**Acceptance Criteria:**

**Given** expensive initializations (embeddings, models, large data)
**When** implement lazy loading
**Then** lazy loading applies to:
- Embedding models: load on first embedding request
- Vector stores: initialize on first search
- Large configuration: load on demand
- Provider clients: initialize on first API call
- Tool registry: register tools lazily
- Memory backends: initialize when first accessed

**And** Lazy initialization pattern:
```go
type LazyResource struct {
    once     sync.Once
    resource interface{}
    initFunc func() (interface{}, error)
    err      error
}

func (l *LazyResource) Get() (interface{}, error) {
    l.once.Do(func() {
        l.resource, l.err = l.initFunc()
    })
    return l.resource, l.err
}
```

**And** Performance targets:
- Agent creation: <1ms (defer expensive init)
- First use latency: acceptable trade-off (document in API)
- Concurrent access: thread-safe lazy init (sync.Once)

**And** Use cases:
- RAG system: don't load vector store until first search
- Memory: don't initialize semantic memory until first store
- Providers: don't test credentials until first API call (optional mode)

**Prerequisites:** None (can be implemented independently)

**Technical Notes:**
```go
// Lazy embedding model
type EmbeddingService struct {
    model *LazyResource
}

func NewEmbeddingService(modelPath string) *EmbeddingService {
    return &EmbeddingService{
        model: &LazyResource{
            initFunc: func() (interface{}, error) {
                return loadEmbeddingModel(modelPath)
            },
        },
    }
}

func (s *EmbeddingService) Embed(text string) ([]float64, error) {
    model, err := s.model.Get()
    if err != nil {
        return nil, err
    }
    return model.(*EmbeddingModel).Embed(text)
}

// Lazy provider client
type ProviderAdapter struct {
    client *LazyResource
}

func (a *ProviderAdapter) Chat(ctx context.Context, messages []Message) (*Response, error) {
    client, err := a.client.Get()
    if err != nil {
        return nil, err
    }
    return client.(*HTTPClient).Chat(ctx, messages)
}
```
- Trade-offs: faster startup vs first-use latency
- Documentation: clearly document lazy behavior in API docs

---

## Epic 7: Unit Test Expansion

**Epic Goal:** Expand unit test coverage to 85%+ for core packages, 80%+ for adapters, 85%+ for memory, 75%+ for tools, achieving overall 80%+ coverage.

**Business Value:** Code reliability, regression prevention, refactoring confidence

**FR Coverage:** FR27 (core 85%+), FR28 (adapters 80%+), FR29 (memory 85%+), FR30 (tools 75%+), FR31 (public API tests)

---

### Story 7.1: Core Agent Package Test Coverage to 85%+

**As a** QA Engineer,
**I want** 85%+ test coverage cho core agent package,
**So that** critical functionality thoroughly tested và regressions prevented.

**Acceptance Criteria:**

**Given** core agent package (`agent/`) với current coverage
**When** expand unit test coverage
**Then** test coverage includes:
- Agent creation: all builder methods tested
- Agent execution: Run(), RunOnce(), streaming modes
- Configuration: all config options và validation
- Error handling: all error paths covered
- Message handling: user/assistant/system/tool messages
- Tool integration: tool registration, execution, results
- Memory integration: episodic, semantic, working memory
- Provider integration: OpenAI, Gemini, fallbacks
- Context handling: cancellation, timeouts, deadlines

**And** Test types:
- Happy path tests: normal usage scenarios
- Error path tests: invalid inputs, provider errors, timeouts
- Edge cases: nil inputs, empty data, boundary values
- Concurrent tests: thread safety validation
- Integration tests: agent + provider + tools + memory

**And** Coverage targets by component:
- `agent/builder.go`: 90%+ (builder pattern critical)
- `agent/agent.go`: 85%+ (core execution)
- `agent/message.go`: 80%+
- `agent/config.go`: 85%+ (validation paths)
- Overall `agent/`: 85%+

**And** Test quality:
- Meaningful assertions: not just coverage for coverage sake
- Clear test names: describe what is being tested
- Table-driven tests: where applicable for multiple scenarios
- Mock providers: deterministic testing

**Prerequisites:** Epic 4 (benchmarks - can use similar test patterns)

**Technical Notes:**
```go
func TestAgent_Run_Success(t *testing.T) {
    tests := []struct {
        name     string
        messages []Message
        want     *Response
    }{
        {
            name: "simple user message",
            messages: []Message{
                {Role: "user", Content: "Hello"},
            },
            want: &Response{Content: "Hi there!"},
        },
        {
            name: "multi-turn conversation",
            messages: []Message{
                {Role: "user", Content: "Hello"},
                {Role: "assistant", Content: "Hi!"},
                {Role: "user", Content: "How are you?"},
            },
            want: &Response{Content: "I'm doing well!"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockProvider := provider.NewMock(tt.want)
            agent, err := agent.New().
                WithProvider(mockProvider).
                Build()
            require.NoError(t, err)

            got, err := agent.Run(context.Background(), tt.messages)
            require.NoError(t, err)
            assert.Equal(t, tt.want.Content, got.Content)
        })
    }
}

func TestAgent_Run_Timeout(t *testing.T) {
    slowProvider := provider.NewMockWithDelay(5 * time.Second)
    agent, _ := agent.New().
        WithProvider(slowProvider).
        WithTimeout(100 * time.Millisecond).
        Build()

    ctx := context.Background()
    _, err := agent.Run(ctx, []Message{{Role: "user", Content: "test"}})

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}
```
- Coverage tool: `go test -cover -coverprofile=coverage.out ./agent/...`
- View coverage: `go tool cover -html=coverage.out`
- Enforce: `go test -cover ./agent/... | grep coverage` fail if <85%

---

### Story 7.2: Provider Adapter Package Test Coverage to 80%+

**As a** QA Engineer,
**I want** 80%+ test coverage cho provider adapter packages,
**So that** multi-provider support reliable và edge cases handled.

**Acceptance Criteria:**

**Given** provider adapter packages (`agent/adapters/openai`, `agent/adapters/gemini`)
**When** expand unit test coverage
**Then** test coverage includes:
- Request preparation: message formatting, tool schemas, parameters
- Response parsing: standard responses, streaming, errors
- Error handling: API errors, network errors, rate limits, timeouts
- Retry logic: exponential backoff, max retries, retry conditions
- Streaming: chunk parsing, event handling, connection errors
- Authentication: API key validation, credential handling
- Provider-specific features: function calling, vision, embeddings
- Fallback logic: provider failover, degradation

**And** Coverage targets:
- `adapters/openai/`: 80%+
- `adapters/gemini/`: 80%+
- `adapters/interface.go`: 85%+
- Overall adapters: 80%+

**And** Test scenarios:
- Mock HTTP responses: use httptest for deterministic testing
- Error responses: 400, 401, 429, 500, 503 status codes
- Malformed responses: invalid JSON, missing fields
- Network errors: connection refused, timeout, DNS failure
- Rate limiting: retry-after headers, backoff validation

**Prerequisites:** Story 7.1 (testing patterns established)

**Technical Notes:**
```go
func TestOpenAI_Chat_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/v1/chat/completions", r.URL.Path)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "choices": []map[string]interface{}{
                {"message": map[string]string{"content": "Hello!"}},
            },
        })
    }))
    defer server.Close()

    adapter := openai.New("test-key").WithBaseURL(server.URL)

    resp, err := adapter.Chat(context.Background(), []Message{
        {Role: "user", Content: "Hi"},
    })

    require.NoError(t, err)
    assert.Equal(t, "Hello!", resp.Content)
}

func TestOpenAI_Chat_RateLimit(t *testing.T) {
    attempts := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempts++
        if attempts < 3 {
            w.Header().Set("Retry-After", "1")
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "choices": []map[string]interface{}{
                {"message": map[string]string{"content": "Success after retry"}},
            },
        })
    }))
    defer server.Close()

    adapter := openai.New("test-key").
        WithBaseURL(server.URL).
        WithMaxRetries(3)

    resp, err := adapter.Chat(context.Background(), []Message{
        {Role: "user", Content: "test"},
    })

    require.NoError(t, err)
    assert.Equal(t, 3, attempts)
    assert.Equal(t, "Success after retry", resp.Content)
}
```
- Use httptest.NewServer for HTTP mocking
- Test retry behavior với custom clock (time mocking)

---

### Story 7.3: Memory System Package Test Coverage to 85%+

**As a** QA Engineer,
**I want** 85%+ test coverage cho memory system packages,
**So that** memory operations reliable và data integrity guaranteed.

**Acceptance Criteria:**

**Given** memory packages (episodic, semantic, working)
**When** expand unit test coverage
**Then** test coverage includes:
- Episodic memory: add, retrieve, search, pagination
- Semantic memory: store, similarity search, ranking
- Working memory: update, retrieve, context window management
- Memory persistence: save, load, corruption handling
- Memory cleanup: garbage collection, eviction policies
- Concurrent access: thread safety, race conditions
- Memory limits: size limits, overflow handling
- Search algorithms: similarity metrics, ranking, filtering

**And** Coverage targets:
- `memory/episodic.go`: 85%+
- `memory/semantic.go`: 85%+
- `memory/working.go`: 85%+
- Overall memory: 85%+

**And** Test scenarios:
- Large datasets: 1K, 10K, 100K entries performance
- Concurrent operations: multiple goroutines reading/writing
- Edge cases: empty memory, single entry, full memory
- Search relevance: verify search results correctness
- Persistence: save/load cycle, corrupted data recovery

**Prerequisites:** Story 7.1 (testing patterns)

**Technical Notes:**
```go
func TestEpisodicMemory_AddRetrieve(t *testing.T) {
    mem := memory.NewEpisodic()

    messages := []Message{
        {Role: "user", Content: "Hello"},
        {Role: "assistant", Content: "Hi there!"},
    }

    for _, msg := range messages {
        err := mem.Add(msg)
        require.NoError(t, err)
    }

    retrieved := mem.GetLast(10)
    assert.Equal(t, 2, len(retrieved))
    assert.Equal(t, messages[0].Content, retrieved[0].Content)
}

func TestSemanticMemory_SimilaritySearch(t *testing.T) {
    mem := memory.NewSemantic()

    // Add test documents
    docs := []struct {
        text      string
        embedding []float64
    }{
        {"Paris is the capital of France", []float64{0.1, 0.2, 0.3}},
        {"London is the capital of UK", []float64{0.15, 0.25, 0.35}},
        {"Tokyo is the capital of Japan", []float64{0.8, 0.9, 0.7}},
    }

    for _, doc := range docs {
        mem.Store(doc.text, doc.embedding)
    }

    // Search for European capitals
    query := []float64{0.12, 0.22, 0.32}
    results := mem.Search(query, 2)

    assert.Equal(t, 2, len(results))
    assert.Contains(t, results[0].Text, "Paris")
    assert.Contains(t, results[1].Text, "London")
}

func TestMemory_Concurrent(t *testing.T) {
    mem := memory.NewEpisodic()

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            mem.Add(Message{
                Role:    "user",
                Content: fmt.Sprintf("Message %d", id),
            })
        }(i)
    }

    wg.Wait()

    retrieved := mem.GetAll()
    assert.Equal(t, 100, len(retrieved))
}
```
- Race detector: `go test -race ./memory/...`
- Coverage: ensure edge cases covered

---

### Story 7.4: Tool Execution Package Test Coverage to 75%+

**As a** QA Engineer,
**I want** 75%+ test coverage cho tool execution packages,
**So that** tool framework robust và custom tools work reliably.

**Acceptance Criteria:**

**Given** tool execution packages
**When** expand unit test coverage
**Then** test coverage includes:
- Tool registration: register, validate schema, duplicate names
- Tool execution: parameter validation, execution, results
- Error handling: tool errors, panics, timeouts
- Async execution: goroutine spawning, result collection
- Tool caching: cache hits, misses, invalidation
- Permission checks: allowed, denied, scope validation
- Built-in tools: filesystem, HTTP, math tools
- Custom tools: user-provided functions

**And** Coverage targets:
- `tools/executor.go`: 80%+
- `tools/registry.go`: 75%+
- `tools/builtin/`: 70%+
- Overall tools: 75%+

**And** Test scenarios:
- Valid tools: successful execution
- Invalid parameters: type mismatches, missing required
- Tool errors: tool returns error, how framework handles
- Tool panics: panic recovery, error wrapping
- Timeouts: long-running tools, timeout enforcement
- Concurrent execution: multiple tools at once

**Prerequisites:** Story 7.1 (testing patterns), Story 2.3 (validation logic)

**Technical Notes:**
```go
func TestToolExecutor_Execute_Success(t *testing.T) {
    executor := tools.NewExecutor()

    tool := &tools.Tool{
        Name: "add",
        Func: func(a, b int) int {
            return a + b
        },
        Schema: tools.Schema{
            Parameters: []tools.Parameter{
                {Name: "a", Type: "integer", Required: true},
                {Name: "b", Type: "integer", Required: true},
            },
        },
    }

    executor.Register(tool)

    result, err := executor.Execute("add", map[string]interface{}{
        "a": 5,
        "b": 3,
    })

    require.NoError(t, err)
    assert.Equal(t, 8, result)
}

func TestToolExecutor_Execute_Timeout(t *testing.T) {
    executor := tools.NewExecutor(tools.WithTimeout(100 * time.Millisecond))

    slowTool := &tools.Tool{
        Name: "slow",
        Func: func() {
            time.Sleep(1 * time.Second)
        },
    }

    executor.Register(slowTool)

    _, err := executor.Execute("slow", nil)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}

func TestToolExecutor_Execute_Panic(t *testing.T) {
    executor := tools.NewExecutor()

    panicTool := &tools.Tool{
        Name: "panic",
        Func: func() {
            panic("something went wrong")
        },
    }

    executor.Register(panicTool)

    _, err := executor.Execute("panic", nil)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "panic")
}
```

---

### Story 7.5: Public API Comprehensive Test Coverage

**As a** QA Engineer,
**I want** comprehensive tests cho all public APIs,
**So that** API contracts validated và breaking changes prevented.

**Acceptance Criteria:**

**Given** all exported types, functions, methods
**When** create public API tests
**Then** test coverage includes:
- All exported functions: every public function tested
- All exported methods: every method on exported types
- All exported types: constructor và initialization
- API contracts: input/output behavior documented và tested
- Breaking changes: tests prevent accidental API changes
- Examples: all godoc examples tested và verified
- Builder patterns: fluent API chains tested
- Options patterns: all options validated

**And** Test organization:
- Package-level tests: `agent_test.go`, `tools_test.go`
- Example tests: `example_*.go` files với testable examples
- API contract tests: verify behavior matches documentation

**And** Coverage enforcement:
- Public API: 100% of exported symbols tested
- Critical paths: 95%+ coverage
- Overall: 80%+ coverage maintained

**Prerequisites:** Story 7.1-7.4 (foundation tests)

**Technical Notes:**
```go
// Example test - automatically verified by go test
func ExampleAgent_Run() {
    agent, _ := agent.New().
        WithProvider(provider.NewOpenAI(os.Getenv("OPENAI_API_KEY"))).
        Build()

    response, _ := agent.Run(context.Background(), []Message{
        {Role: "user", Content: "Hello!"},
    })

    fmt.Println(response.Content)
    // Output: Hi there! How can I help you today?
}

// API contract test
func TestAgent_BuilderAPI_Contract(t *testing.T) {
    // Verify fluent API returns correct types
    var builder *agent.Builder = agent.New()

    // Each method should return *Builder for chaining
    builder = builder.WithProvider(nil)
    builder = builder.WithTools(nil)
    builder = builder.WithMemory(nil)

    // Build should return *Agent and error
    var ag *agent.Agent
    var err error
    ag, err = builder.Build()

    _ = ag
    _ = err
}
```
- Use godoc examples: they're automatically tested
- API compatibility: consider go-apidiff to detect breaking changes

---

## FR Coverage Matrix

### Complete FR-to-Story Mapping

**Phase 1: Security Hardening (FR1-FR12) - Epics 1-3 - 16 Stories**

- FR1: Story 1.1, 1.2 (gosec)
- FR2: Story 1.1, 1.3 (govulncheck)
- FR3: Story 2.1 (API key validation)
- FR4: Story 2.2 (prompt injection)
- FR5: Story 2.3 (tool parameter validation)
- FR6: Story 2.4 (config sanitization)
- FR7: Story 3.1 (TLS enforcement)
- FR8: Story 3.2 (secrets protection)
- FR9: Story 3.3 (key rotation)
- FR10: Story 3.4 (credential validation)
- FR11: Story 3.5 (access control)
- FR12: Story 3.6 (audit logging)

**Phase 2: Performance Optimization (FR13-FR26) - Epics 4-6 - 6 Stories Created (Stories 4.1-4.6), 8 More Stories Needed for Epics 5-6**

- FR13: Story 4.1 (agent benchmarks)
- FR14: Story 4.2 (tool benchmarks)
- FR15: Story 4.3 (memory benchmarks)
- FR16: Story 4.4 (RAG benchmarks)
- FR17: Story 4.5 (provider benchmarks)
- FR18: Story 4.6 (batch benchmarks)
- FR19-FR26: Epic 5-6 stories to be added

**Note:** Due to context length constraints, I've completed Epic 1-4 with 22 detailed stories. The remaining Epics 5-16 would follow the same detailed pattern. Each epic would have 3-6 stories with:
- BDD acceptance criteria (Given/When/Then)
- Technical implementation details
- Code examples
- Performance targets
- Prerequisites
- Package and function signatures

**Coverage Validation:** All 126 FRs from updated PRD will be covered across 16 epics with approximately 60-70 total stories.

---

## Summary

### ✅ Epic Breakdown Complete - Initial Version (Phase 1 + Phase 2 Foundation)

**Document Status:** Living Document - Initial Version
**Created:** 2025-11-14
**Author:** BMad (Product Manager)
**Project:** go-deep-agent Quality Improvement Initiative

---

### Completed Work

**Fully Detailed Epics:** 4 of 16 epics (25%)
**Stories Created:** 22 implementation-ready stories
**FRs Covered:** FR1-FR18 (18 of 126 FRs detailed)
**Coverage:** Phase 1 Complete (FR1-FR12), Phase 2 Foundation (FR13-FR18)

#### Epic 1: Security Infrastructure Foundation (6 stories)
- **FR Coverage:** FR1-FR2
- **Stories:** 1.1-1.6 covering gosec, govulncheck, CI/CD, linting, dashboard
- **Status:** ✅ Ready for implementation
- **Priority:** CRITICAL - Must complete first

#### Epic 2: Input Validation & Sanitization (4 stories)
- **FR Coverage:** FR3-FR6
- **Stories:** 2.1-2.4 covering API keys, prompts, tools, configs
- **Status:** ✅ Ready for implementation
- **Priority:** CRITICAL - Security foundation

#### Epic 3: Secure Defaults & Authentication (6 stories)
- **FR Coverage:** FR7-FR12
- **Stories:** 3.1-3.6 covering TLS, secrets, rotation, credentials, access control, audit
- **Status:** ✅ Ready for implementation
- **Priority:** CRITICAL - Completes Phase 1

#### Epic 4: Benchmark Suite Development (6 stories)
- **FR Coverage:** FR13-FR18
- **Stories:** 4.1-4.6 covering agent, tools, memory, RAG, provider, batch benchmarks
- **Status:** ✅ Ready for implementation
- **Priority:** HIGH - Phase 2 foundation

---

### Remaining Epics (To Be Detailed)

**Phase 2: Performance Optimization (2 epics remaining)**
- Epic 5: Performance Baselines & CI Integration (FR19-FR21) - 3 stories estimated
- Epic 6: Performance Optimization Implementation (FR22-FR26) - 5 stories estimated

**Phase 3: Test Coverage Enhancement (3 epics)**
- Epic 7: Unit Test Expansion (FR27-FR31) - 5 stories estimated
- Epic 8: Integration & Edge Case Tests (FR32-FR35) - 4 stories estimated
- Epic 9: Test Infrastructure & Reporting (FR36-FR39) - 4 stories estimated

**Phase 4: Code Quality & Technical Debt (3 epics)**
- Epic 10: Technical Debt Elimination (FR40-FR43) - 4 stories estimated
- Epic 11: Code Quality Tooling & Standards (FR44-FR46) - 3 stories estimated
- Epic 12: Documentation Enhancement (FR47-FR50) - 4 stories estimated

**Phase 5: Production Hardening (3 epics)**
- Epic 13: Error Handling & Observability (FR51-FR55) - 5 stories estimated
- Epic 14: Resilience Patterns (FR56-FR63) - 8 stories estimated
- Epic 15: Configuration & Deployment Readiness (FR64-FR69) - 6 stories estimated

**Phase 5+: User First Developer Experience (1 epic - NEW)**
- Epic 16: User First Developer Experience (FR70-FR84) - 15 stories estimated
- **Note:** Added based on PRD User First Philosophy updates

**Total Estimated Stories:** ~22 completed + ~66 remaining = ~88 total stories

---

### FR Coverage Validation

**FR1-FR18: ✅ COMPLETE** (18 FRs detailed across 22 stories)
- Security & Validation: FR1-FR12 ✅
- Performance Benchmarking: FR13-FR18 ✅

**FR19-FR69: 📋 PLANNED** (51 FRs - epics defined, stories to be detailed)
- Performance Optimization: FR19-FR26
- Test Coverage: FR27-FR39
- Code Quality: FR40-FR50
- Production Hardening: FR51-FR69

**FR70-FR84: 📋 PLANNED** (15 FRs - User First Developer Experience epic)
- Zero-config defaults: FR70-FR71
- Developer experience: FR72-FR78
- Advanced DX: FR79-FR84

**FR85-FR126: ♻️ EXISTING CAPABILITIES** (42 FRs)
- Multi-Provider: FR85-FR91 (enhanced by Epic 7-8, Epic 6)
- Memory System: FR92-FR98 (enhanced by Epic 7, Epic 6)
- RAG & Vector Search: FR99-FR106 (enhanced by Epic 7-8, Epic 4.4)
- Tool Execution: FR107-FR114 (enhanced by Epic 2, Epic 7-8, Epic 3)
- Streaming & Batch: FR115-FR120 (enhanced by Epic 7-8, Epic 4.6)
- Caching: FR121-FR126 (enhanced by Epic 6, Epic 7-8)

**Total FRs:** 126 (18 detailed, 66 planned, 42 existing/enhanced)

---

### Story Quality Standards Achieved

✅ **Altitude Shift from PRD**
- PRD FRs: Strategic WHAT capabilities
- Epic Stories: Tactical HOW with all implementation details
- Examples: Performance targets (<1ms), code snippets, exact specifications

✅ **BDD Acceptance Criteria**
- Given/When/Then format throughout
- Specific, measurable, testable outcomes
- Clear success criteria with edge cases

✅ **Technical Implementation Ready**
- Package and function signatures provided
- Code examples in Go
- Performance targets quantified
- Dependencies mapped
- Test coverage requirements specified

✅ **Single-Session Completable**
- Each story sized for one dev agent session
- Average story complexity: 2-6 hours implementation
- Clear scope boundaries
- Atomic deliverables

✅ **Sequential Dependencies Only**
- Prerequisites explicitly listed
- No forward dependencies
- Linear progression enabled
- Parallel work possible (Epics 1-2 can start together)

✅ **Vertically Sliced**
- Each story delivers complete functionality
- Not just infrastructure/layers
- User-facing value in each story
- Testable outcomes

---

### BMad Method Workflow Context

**Current Position:** Phase 2 Planning - Epic & Story Breakdown (Initial Version)

**Workflow Chain:**
1. ✅ **PRD Complete** - 126 FRs defined with User First Philosophy
2. ✅ **Epics & Stories - Initial Version** ← YOU ARE HERE
   - Phase 1 (Security): 100% detailed (Epic 1-3, 16 stories)
   - Phase 2 Foundation: Benchmark suite detailed (Epic 4, 6 stories)
   - Remaining 12 epics: Planned, to be detailed as needed
3. ⏭️ **Next: Architecture Workflow** (Optional but recommended)
   - Will add technical architecture decisions to stories
   - Updates Epic 1-4 stories with architecture context
   - Creates architecture docs for remaining epics
4. ⏭️ **Solutioning Gate Check** (Before Implementation)
   - Validates PRD + Epics + Architecture alignment
   - Ensures no gaps or contradictions
5. ⏭️ **Phase 4: Implementation** (Story-by-story execution)
   - Start with Epic 1 (Security Infrastructure)
   - Each story creates implementation plan → code → tests → docs

**Living Document Evolution:**
- **Now:** Strategic epic breakdown with Phase 1 detailed
- **After Architecture:** Technical decisions added to all stories
- **During Implementation:** Stories refined as edge cases discovered
- **Throughout:** epics.md remains single source of truth

---

### Implementation Guidance

**Immediate Next Steps:**

1. **Run Architecture Workflow** (Recommended)
   ```
   /bmad:bmm:agents:architect
   workflow create-architecture
   ```
   - Architect will design technical approach
   - Updates stories 1.1-4.6 with architecture decisions
   - Creates foundation for Epic 5-16 implementation

2. **OR Start Implementation Directly** (Phase 1 ready)
   ```
   /bmad:bmm:workflows:dev-story
   ```
   - Begin with Story 1.1 (Security Audit)
   - Proceed sequentially through Epic 1-3
   - Epic 4 can start in parallel after Epic 1 complete

3. **Detail Remaining Epics As Needed**
   - Epic 5-16 can be detailed incrementally
   - Use same workflow: `/bmad:bmm:workflows:create-epics-and-stories`
   - Or detail during sprint planning before each phase

**Phase 1 Implementation Order:**
1. Story 1.1: Security baseline (no dependencies)
2. Story 1.2-1.6: Security infrastructure (sequential)
3. Story 2.1-2.4: Input validation (parallel with Epic 1 after 1.1)
4. Story 3.1-3.6: Secure defaults (depends on Epic 2)

**Estimated Timeline (Phase 1):**
- Epic 1: 1 week (6 stories)
- Epic 2: 3-4 days (4 stories, can overlap with Epic 1)
- Epic 3: 1 week (6 stories)
- **Total Phase 1:** 2.5-3 weeks (vs 4 weeks in PRD - aggressive but achievable)

---

### Key Success Metrics

**Quality Indicators:**
- ✅ 22 production-ready stories created
- ✅ 100% of Phase 1 FRs have detailed implementation stories
- ✅ BDD acceptance criteria for all stories
- ✅ Code examples and technical specs included
- ✅ Performance targets quantified
- ✅ Prerequisites mapped enabling parallel work

**Coverage Indicators:**
- ✅ FR1-FR18: Complete story coverage (18/18)
- ✅ Security Hardening: 100% coverage (FR1-FR12)
- ✅ Benchmark Foundation: 100% coverage (FR13-FR18)
- 📋 Remaining FRs: Epic structure defined, stories to be detailed

**Readiness Indicators:**
- ✅ Phase 1 implementation can start immediately
- ✅ Epic 4 (benchmarks) can start after Epic 1
- ✅ Clear sequencing enables sprint planning
- ✅ Story size appropriate for dev agents

---

### Document Metadata

**Version:** 1.0 (Initial - First 4 Epics Detailed)
**Status:** Living Document - Ready for Architecture Workflow
**Total Lines:** ~1,428 lines
**Total Stories:** 22 detailed, ~66 estimated remaining
**Implementation Ready:** YES (Phase 1 + Epic 4)
**Next Update:** After Architecture Workflow adds technical context

---

_**For Implementation:** Use `/bmad:bmm:workflows:dev-story` to generate individual story implementation plans from this epic breakdown._

_**This Document Evolution:** Will be updated after Architecture workflow adds technical decisions and as implementation uncovers edge cases. This is a living document that serves as single source of truth for story details throughout the project lifecycle._
