# System-Level Test Design - go-deep-agent

**Date:** 2025-11-14
**Author:** Murat (Master Test Architect)
**Project Phase:** Phase 3 Solutioning - Testability Review
**Status:** Ready for Solutioning Gate Check

---

## Executive Summary

This system-level testability review evaluates go-deep-agent architecture against testability criteria before solutioning gate check. Project is a **Go library** for building AI agents (not an application), requiring specialized test strategy for SDK/library validation.

**Testability Assessment:**
- **Controllability:** PASS - Factory patterns, mock providers, test helpers enable state control
- **Observability:** PASS - Comprehensive logging, error types, structured responses
- **Reliability:** CONCERNS - Flaky test risk from external LLM APIs, no benchmark regression detection

**Key Finding:** Architecture is highly testable but current test coverage (71.2%) and missing performance benchmarks create production risk. PRD Phase 1-3 (Security, Performance, Testing) will address gaps.

---

## 1. Testability Assessment

### 1.1 Controllability (PASS with minor concerns)

**Definition:** Can we control system state for testing?

**Evidence:**

✅ **Factory Patterns for Test Data**
- Test helpers in `agent/testutil/` provide factories for agents, tools, messages
- Mock providers (`NewMockProvider()`) enable deterministic testing
- Builder pattern allows precise configuration control in tests

✅ **Dependency Injection**
- Provider adapters are interfaces - easily mockable
- Memory backends pluggable (in-memory, file, Redis)
- Tools registered dynamically - test isolation possible

✅ **State Management**
- Memory can be seeded with known messages
- Configuration exposed for validation
- Agent state queryable for assertions

⚠️ **External Dependencies**
- Real OpenAI/Gemini calls in integration tests (slow, flaky, costly)
- **Mitigation (PRD Epic 7-9):** Mock provider adapters, fixture responses, offline mode

**Recommendation:**
- Expand mock provider coverage (Story 7.1 in PRD)
- Add `TestMode` configuration for deterministic behavior
- Create `agent/testutil/fixtures/` for canned LLM responses

---

### 1.2 Observability (PASS)

**Definition:** Can we inspect system state and validate results?

**Evidence:**

✅ **Structured Error Types**
- `ErrInvalidConfig`, `ErrProviderFailed`, `ErrToolExecutionFailed`
- Error wrapping with context (`fmt.Errorf(...%w)`)
- Error messages actionable and debuggable

✅ **Comprehensive Logging**
- `slog` integration for structured logging (PRD Epic 3 enhances)
- Tool execution logs (inputs, outputs, duration)
- ReAct pattern logs (Thought → Action → Observation)

✅ **Test-Friendly APIs**
- Public APIs well-documented with examples
- Builder pattern exposes configuration for validation
- Agent state queryable (memory, tools, config)

✅ **Deterministic Results**
- Test assertions validate message structure, tool calls, memory state
- No hidden randomness (except LLM outputs - mocked in unit tests)

**Recommendation:**
- Add telemetry hooks for OpenTelemetry (PRD Epic 13 - FR54-FR55)
- Enhance test assertions with structured error validation
- Add debug mode with detailed traces

---

### 1.3 Reliability (CONCERNS - Requires Mitigation)

**Definition:** Are tests isolated, reproducible, and stable?

**Evidence:**

⚠️ **Test Isolation**
- Current: 1,344 tests, ratio 1.41:1 (good)
- **Concern:** No explicit cleanup strategy documented in tests
- **Mitigation (PRD Epic 7-9):** Fixtures with auto-cleanup, parallel-safe data factories

⚠️ **External API Flakiness**
- Integration tests call real OpenAI/Gemini APIs
- **Risk:** Rate limits, network failures, API changes
- **Mitigation (PRD Epic 7-9):** Offline mode, mock responses, contract testing

⚠️ **Performance Regression Detection**
- **Critical Gap:** No benchmarks (identified in TECHNICAL_ASSESSMENT)
- **Risk:** Performance degradation undetected until production
- **Mitigation (PRD Epic 4-6):** Benchmark suite, CI regression detection

✅ **Loose Coupling**
- Provider adapters, memory backends, tools are interfaces
- Components testable in isolation
- No circular dependencies (except tools - PRD Epic 10 fixes)

**Recommendation:**
- **HIGH PRIORITY:** Implement benchmark suite (Epic 4-6 in PRD)
- **HIGH PRIORITY:** Create offline test mode with fixture responses
- **MEDIUM PRIORITY:** Add contract tests for provider API changes

---

## 2. Architecturally Significant Requirements (ASRs)

ASRs are quality requirements driving architecture and testability decisions.

### ASR-1: Multi-Provider Abstraction (Score: 6 - High Priority)

**Requirement:** Library must support multiple LLM providers (OpenAI, Gemini, Ollama) with unified interface.

**Architecture Decision:** Provider adapter pattern with interface `Provider`

**Testability Challenge:**
- Provider-specific quirks (error codes, rate limits, response formats)
- Real API calls expensive and slow in tests

**Test Strategy:**
- **Unit:** Mock provider interface, validate request formatting (Epic 7 - FR28)
- **Integration:** Contract tests against provider API schemas (Epic 8 - FR32)
- **E2E:** Smoke tests with real APIs (limited, gated behind env var)

**Risk:** **Probability 2 × Impact 3 = Score 6 (HIGH)**
- Failure: Provider API changes break integration silently
- Mitigation: Contract tests, versioned API mocks, automated provider smoke tests

---

### ASR-2: Memory System Persistence (Score: 4 - Medium Priority)

**Requirement:** Memory must survive restarts, support multi-user isolation, scale to 1M+ messages.

**Architecture Decision:** 3-tier memory (Working, Episodic, Semantic) with pluggable backends

**Testability Challenge:**
- File-based persistence creates test state pollution
- Redis backend requires infrastructure
- Large corpus testing (1M messages) slow in CI

**Test Strategy:**
- **Unit:** In-memory backend, validate memory operations (Epic 7 - FR29)
- **Integration:** Test database (sqlite), fixture data (Epic 8 - FR32)
- **Performance:** Benchmark memory scaling (1K, 10K, 100K, 1M messages) (Epic 4 - FR15)

**Risk:** **Probability 2 × Impact 2 = Score 4 (MEDIUM)**
- Failure: Memory growth unbounded, OOM in production
- Mitigation: LRU eviction policy (PRD Epic 6 - FR24), memory cleanup tests

---

### ASR-3: ReAct Pattern Correctness (Score: 9 - CRITICAL)

**Requirement:** ReAct (Reasoning + Acting) pattern must implement Yao et al. 2022 paper correctly.

**Architecture Decision:** State machine (Think → Act → Observe → Done) with streaming support

**Testability Challenge:**
- ReAct logic complex - multi-iteration loops, error recovery, termination conditions
- LLM non-determinism makes E2E tests flaky
- Tool execution failures must not crash agent

**Test Strategy:**
- **Unit:** Mock LLM responses, validate state transitions (Epic 7 - FR31)
- **Integration:** Mock tools, validate ReAct loop with canned scenarios (Epic 8 - FR33)
- **E2E:** Real LLM + real tools (smoke only, behind feature flag)

**Risk:** **Probability 3 × Impact 3 = Score 9 (CRITICAL)**
- Failure: ReAct infinite loops, incorrect tool calls, crashes on errors
- Mitigation: **IMMEDIATE** - Expand ReAct test coverage (currently 72%, target 85%+)
  - Epic 7: Unit tests for state machine transitions
  - Epic 8: Error path tests (timeout, retry, max iterations)
  - Epic 9: Integration tests with deterministic mock tools

---

### ASR-4: Builder API Type Safety (Score: 6 - High Priority)

**Requirement:** Fluent builder must prevent invalid configurations at compile time.

**Architecture Decision:** 74 fluent methods with compile-time type checking

**Testability Challenge:**
- Builder state mutations complex (74 methods)
- Invalid configuration combinations possible (e.g., `WithReActMode(true)` without tools)
- Error messages must be actionable

**Test Strategy:**
- **Unit:** Test all 74 builder methods (Epic 7 - FR27)
- **Integration:** Test invalid configuration detection (Epic 8 - FR34)
- **E2E:** Test production-like configurations (Epic 9 - FR38)

**Risk:** **Probability 2 × Impact 3 = Score 6 (HIGH)**
- Failure: Runtime errors for invalid configs, users confused
- Mitigation: Early validation in `Build()` (PRD Epic 2 - FR3-FR6)

---

### ASR-5: Performance Overhead <1ms (Score: 6 - High Priority)

**Requirement:** Framework overhead must be <1ms for agent creation, <100μs for tool dispatch.

**Architecture Decision:** Lightweight builder, no reflection, minimal allocations

**Testability Challenge:**
- **CRITICAL GAP:** No benchmarks currently (TECHNICAL_ASSESSMENT finding)
- Performance regressions undetected
- Production impact unknown

**Test Strategy:**
- **Benchmarks:** Comprehensive suite (Epic 4 - FR13-FR18)
  - Agent creation: <1ms (FR13)
  - Tool dispatch: <100μs (FR14)
  - Memory ops: <10ms (FR15)
  - RAG search: <50ms/10K docs (FR16)
- **CI:** Automated regression detection (Epic 5 - FR19-FR21)
- **Profiling:** CPU/memory profiles for hot paths (Epic 6 - FR22-FR26)

**Risk:** **Probability 2 × Impact 3 = Score 6 (HIGH)**
- Failure: Performance degrades unnoticed, production SLOs breached
- Mitigation: **IMMEDIATE** - Epic 4-6 implementation (benchmark suite + CI)

---

## 3. Test Levels Strategy

Based on architecture (Go library SDK, not application), test pyramid differs from typical web app:

```
      E2E (5%)
      --------
   Integration (25%)
   ----------------
    Unit (70%)
    -----------
```

**Rationale:** Libraries need deep unit coverage (70%) for all public APIs and edge cases. Integration tests (25%) validate component interaction. E2E tests (5%) are minimal - smoke tests with real providers.

### 3.1 Unit Tests (70% - Target 85%+ for core)

**Scope:**
- All public APIs (74 builder methods)
- Business logic (ReAct state machine, memory operations, tool execution)
- Error handling (all error paths)
- Edge cases (nil inputs, empty data, boundary conditions)

**Approach:**
- Mock provider interface (`NewMockProvider()`)
- In-memory backends (no file/Redis dependencies)
- Deterministic tool responses
- Fast execution (<1s per test file)

**Epic Coverage:** Epic 7 (FR27-FR31)

---

### 3.2 Integration Tests (25%)

**Scope:**
- Multi-component interaction (agent + provider + memory + tools)
- Backend integrations (file persistence, Redis)
- Provider adapter contract validation
- Error propagation across components

**Approach:**
- Test databases (sqlite in-memory)
- Mock HTTP servers for provider APIs
- Fixture-based LLM responses
- Contract testing (Pact for provider APIs)

**Epic Coverage:** Epic 8 (FR32-FR35)

---

### 3.3 E2E Tests (5%)

**Scope:**
- Smoke tests with real OpenAI/Gemini APIs
- Production configuration validation
- Real Redis backend
- Actual tool execution (safe tools only)

**Approach:**
- Gated behind `E2E_TESTS=true` env var
- Run only in pre-release pipeline
- Limited to critical paths (basic chat, ReAct, memory persistence)
- Acceptance: Flakiness expected, retries allowed (max 3)

**Epic Coverage:** Epic 9 (FR36-FR39)

---

### 3.4 Performance Tests (Benchmarks)

**Scope:**
- Framework overhead measurement
- Scalability validation (1K, 10K, 100K operations)
- Regression detection

**Approach:**
- Go benchmarking (`testing.B`)
- Benchstat for statistical comparison
- CI integration with baseline tracking

**Epic Coverage:** Epic 4-6 (FR13-FR26)

---

## 4. NFR Testing Approach

### 4.1 Security (Epic 1-3 in PRD)

**Approach:**
- **Static Analysis:** gosec, golangci-lint (FR1-FR2)
- **Input Validation:** Unit tests for API key, prompt, tool parameter validation (FR3-FR6)
- **Secrets Protection:** Verify no logging of sensitive data (FR7-FR8)
- **Access Control:** Tool permission scope tests (FR11)

**Validation:**
- ✅ PASS: All security tests green, zero high/critical vulnerabilities
- ⚠️ CONCERNS: Minor gaps with mitigation plan
- ❌ FAIL: Critical exposure (API key logged, SQL injection possible)

---

### 4.2 Performance (Epic 4-6 in PRD)

**Approach:**
- **Go Benchmarks:** Measure framework overhead (not LLM latency)
- **Scaling Tests:** Validate memory with 1K, 10K, 100K, 1M messages
- **Profiling:** `pprof` for CPU/memory hot paths
- **CI Integration:** Automated regression detection

**Validation:**
- ✅ PASS: All targets met (<1ms agent, <100μs tools, <10ms memory, <50ms RAG)
- ⚠️ CONCERNS: Trending toward limits (e.g., 480ms approaching 500ms)
- ❌ FAIL: SLO breached (e.g., agent creation >1ms)

**Note:** k6 NOT applicable (library, not web service). Use Go benchmarks instead.

---

### 4.3 Reliability (Epic 13-14 in PRD)

**Approach:**
- **Error Handling Tests:** Validate retry, fallback, graceful degradation (FR56-FR60)
- **Circuit Breaker:** Test failure threshold behavior (FR56)
- **Context Cancellation:** Validate timeout enforcement (FR58-FR61)
- **Panic Recovery:** Validate tool panic handling (FR60)

**Validation:**
- ✅ PASS: Error handling comprehensive, retries validated
- ⚠️ CONCERNS: Partial coverage, missing telemetry
- ❌ FAIL: No recovery path (500 error crashes agent)

---

### 4.4 Maintainability (Epic 10-12 in PRD)

**Approach:**
- **Coverage:** CI jobs (not Playwright) - target 80%+ (FR27-FR30)
- **Code Quality:** golangci-lint, Go Report Card A+ (FR44-FR46)
- **Duplication:** <5% (FR40)
- **Documentation:** 100% godoc coverage (FR47-FR49)

**Validation:**
- ✅ PASS: 80%+ coverage, <5% duplication, A+ rating, full docs
- ⚠️ CONCERNS: Coverage 60-79%, duplication >5%
- ❌ FAIL: <60% coverage, tangled code (>10% duplication)

**Current Status:** 71.2% coverage (CONCERNS - needs +8.8% to PASS)

---

## 5. Test Environment Requirements

### 5.1 Local Development

**Dependencies:**
- Go 1.21+ (slog support)
- No external services required (in-memory backends)
- Optional: Redis for backend integration tests

**Setup Time:** <5 minutes (`go mod download`)

---

### 5.2 CI/CD Pipeline

**Requirements:**
- GitHub Actions runners (ubuntu-latest)
- Go 1.21+ installed
- Redis container (for integration tests)
- OpenAI API key (for E2E smoke tests, optional)

**Tools:**
- gosec v2.22.10 (security scanning)
- golangci-lint v2.6.1 (code quality)
- govulncheck v1.1.4+ (vulnerability scanning)
- benchstat (benchmark comparison)

**Runtime:** <15 minutes (unit 5min, integration 5min, security 3min, quality 2min)

---

### 5.3 Pre-Release Validation

**Requirements:**
- Real OpenAI API access (smoke tests)
- Real Gemini API access (smoke tests)
- Redis production-like instance
- Benchmark baseline comparison

**Frequency:** Weekly or before release

---

## 6. Testability Concerns

### 6.1 CRITICAL: No Performance Benchmarks

**Issue:** Currently 0 benchmarks, performance regressions undetected.

**Impact:** Production risk - performance degradation unknown until deployed.

**Recommendation:** **BLOCKER for solutioning gate check**
- Epic 4 (Benchmark Suite): Implement comprehensive benchmarks (FR13-FR18)
- Epic 5 (CI Integration): Automated regression detection (FR19-FR21)
- Epic 6 (Optimization): Address identified bottlenecks (FR22-FR26)

**Gate Decision:** **CONCERNS** - Can proceed to implementation if Epic 4-6 prioritized in Sprint 0.

---

### 6.2 MEDIUM: Test Coverage 71.2% (Target 80%+)

**Issue:** Current coverage below library best practices (80%+ for SDKs).

**Gaps:**
- Tool calling error paths: 65% (needs +10%)
- Planning layer edge cases: 60% (needs +15%)
- Memory cleanup scenarios: 78% (needs +7%)

**Recommendation:** Medium priority, address in Epic 7-9.

**Gate Decision:** **CONCERNS** - Acceptable if gap acknowledged and Epic 7-9 scheduled.

---

### 6.3 LOW: Tool Type Safety (map[string]interface{})

**Issue:** Tools use `map[string]interface{}` instead of type-safe parameters.

**Impact:** Runtime errors from type mismatches, poor DX.

**Recommendation:** Low priority, address in Epic 11 (Code Quality).

**Gate Decision:** **PASS with note** - Workaround exists (JSON Schema validation).

---

### 6.4 LOW: Cyclic Dependencies (agent ↔ tools)

**Issue:** `agent` and `tools` packages have circular dependency via `go:linkname` hack.

**Impact:** Fragile, breaks with Go version changes.

**Recommendation:** Low priority, refactor in v0.12.0 (PRD Epic 10).

**Gate Decision:** **PASS with note** - Technical debt, not testability blocker.

---

## 7. Recommendations for Sprint 0

Before implementation (Phase 4), complete these foundational tasks:

### 7.1 Test Infrastructure (*framework workflow)

**Tasks:**
- Set up benchmark infrastructure (Epic 4 - 1 week)
- Configure CI/CD security scanning (Epic 1 - 3 days)
- Create test fixture library (Epic 9 - 3 days)
- Establish coverage baselines (Epic 7 - 2 days)

**Owner:** Test Architect (TEA)
**Timeline:** Week 1-2 of Sprint 0

---

### 7.2 CI/CD Quality Gates (*ci workflow)

**Tasks:**
- gosec + govulncheck + golangci-lint pipeline (Epic 1 - 2 days)
- Benchmark regression detection (Epic 5 - 3 days)
- Coverage threshold enforcement (Epic 9 - 1 day)
- Security SARIF reporting (Epic 1 - 1 day)

**Owner:** DevOps + Test Architect
**Timeline:** Week 2 of Sprint 0

---

### 7.3 Mock Provider Framework (Epic 7)

**Tasks:**
- Create `MockProvider` with fixture responses
- Add `TestMode` configuration flag
- Generate canned LLM responses for tests
- Document offline testing guide

**Owner:** Dev Team + Test Architect
**Timeline:** Week 3 of Sprint 0

---

## 8. Traceability Matrix (Sample)

Full matrix in Epic-to-FR mapping (docs/epics.md), sample below:

| ASR | Risk Score | Epic | Stories | Test Level | Coverage Target |
|-----|-----------|------|---------|------------|----------------|
| ASR-3 (ReAct) | 9 (CRITICAL) | Epic 7-8 | 7.1-7.5, 8.1-8.4 | Unit + Integration | 85%+ |
| ASR-5 (Performance) | 6 (HIGH) | Epic 4-6 | 4.1-4.6, 5.1-5.3 | Benchmarks | 100% (all targets) |
| ASR-1 (Multi-Provider) | 6 (HIGH) | Epic 7-8 | 7.2, 8.2 | Unit + Contract | 80%+ |
| ASR-2 (Memory) | 4 (MEDIUM) | Epic 7, 4 | 7.3, 4.3 | Unit + Perf | 85%+ |
| ASR-4 (Builder) | 6 (HIGH) | Epic 7-8 | 7.1, 8.4 | Unit + Integration | 85%+ |

---

## 9. Gate Decision

### 9.1 Testability Assessment Summary

| Criterion | Status | Details |
|-----------|--------|---------|
| **Controllability** | ✅ PASS | Factory patterns, mock providers, DI architecture |
| **Observability** | ✅ PASS | Structured errors, comprehensive logging, test-friendly APIs |
| **Reliability** | ⚠️ CONCERNS | No benchmarks (BLOCKER), test isolation needs fixtures |

### 9.2 Overall Gate Decision: **CONCERNS**

**Rationale:**
- Architecture is highly testable (PASS controllability and observability)
- Critical gap: Zero performance benchmarks (identified in TECHNICAL_ASSESSMENT)
- Test coverage 71.2% below library best practices (80%+)

**Conditions to Proceed:**
1. **MANDATORY:** Epic 4-6 (Benchmarks) scheduled in Sprint 0 or Sprint 1
2. **MANDATORY:** Epic 7-9 (Test Coverage) included in roadmap
3. **RECOMMENDED:** Test infrastructure setup before Sprint 1

**Waiver:** Not required - concerns addressable in implementation phase.

**Approval:** Ready for solutioning gate check if Epic 4-9 committed to roadmap.

---

## 10. Next Steps

### 10.1 Immediate (Before Solutioning Gate Check)

1. Review this document with architect and PM
2. Confirm Epic 4-9 prioritization in PRD
3. Validate test strategy aligns with implementation plan
4. Run solutioning gate check workflow (`*solutioning-gate-check`)

### 10.2 Sprint 0 (Foundation)

1. Execute `*framework` workflow to scaffold test infrastructure
2. Execute `*ci` workflow to set up quality gates
3. Create mock provider framework (3 days)
4. Establish benchmark baselines (2 days)

### 10.3 Sprint 1+ (Implementation)

1. Follow Epic 1-15 sequence from PRD
2. Maintain 80%+ coverage gate for each Epic
3. Track ASR risk scores (Epic completion should reduce scores)
4. Run `*trace` workflow Phase 2 before each Epic completion

---

## Document Metadata

**Version:** 1.0
**Workflow:** BMad Method - Phase 3 Solutioning - Test Design (System-Level)
**Generated By:** TEA Agent (Master Test Architect)
**Confidence:** High (based on PRD, Architecture, TECHNICAL_ASSESSMENT analysis)

**Related Documents:**
- [PRD](./PRD.md) - 126 Functional Requirements
- [Architecture](./architecture.md) - Technical architecture decisions
- [Epics](./epics.md) - Epic-to-FR traceability (Epic 1-15)
- [TECHNICAL_ASSESSMENT_v0.10.1.md](../TECHNICAL_ASSESSMENT_v0.10.1.md) - Professional assessment (92/100)

---

**Generated by BMAD TEA Agent - Test Architect Module**
**Workflow:** `.bmad/bmm/testarch/test-design`
**Version:** 4.0 (BMad v6)
**Knowledge Base:** nfr-criteria.md, test-levels-framework.md, risk-governance.md, test-quality.md
