# Story 1.2: Configure gosec for Comprehensive Security Scanning

Status: ready-for-dev
Date Created: 2025-11-15
Epic: Epic 1 - Security Infrastructure Foundation

## Story

As a **DevOps Engineer**,
I want **gosec configured với comprehensive rule set**,
so that **mọi security vulnerability được detected automatically**.

## Acceptance Criteria

**AC-1: All gosec rules enabled với severity thresholds**
- **Given** gosec v2.22.10+ installed trong development environment
- **When** tạo gosec configuration file `.gosec.json`
- **Then** configuration bao gồm:
  - All gosec rules enabled (G101-G602)
  - Severity thresholds: fail on High và Critical
  - Confidence levels: Medium confidence minimum
  - Rule exclusions documented và justified

**AC-2: Test files có relaxed rules**
- **Given** test files (`*_test.go`) có different security requirements
- **When** configure exclusions cho test files
- **Then** exclusions include:
  - G404 (weak random in tests) - acceptable for test data generation
  - G101 (hardcoded credentials in test fixtures) - acceptable for test mocks
  - Exclusions limited to test files only (`**/*_test.go`)
  - Production code (`agent/**/*.go` excluding tests) has full rules

**AC-3: Multiple output formats configured**
- **Given** gosec configuration file
- **When** define output formats
- **Then** formats include:
  - JSON format: machine-parsable, detailed findings
  - SARIF format: GitHub Security tab integration
  - Text format: human-readable console output
  - All formats include: CWE ID, severity, file location, line number, description

**AC-4: Concurrent scanning enabled**
- **Given** multi-core development và CI environments
- **When** configure scanning performance
- **Then** configuration includes:
  - Concurrency: utilize all CPU cores (auto-detect)
  - Scan target: `./...` (all packages)
  - Performance target: <2 minutes for full scan
  - Exclude patterns: vendor/, testutil/mocks/ (generated code)

**AC-5: Custom rules cho go-deep-agent patterns**
- **Given** go-deep-agent có specific security patterns
- **When** define custom rules
- **Then** custom rules cover:
  - API key validation: detect missing validation before use
  - Secrets logging: detect potential secrets in log calls
  - TLS enforcement: detect missing TLS configuration
  - Tool parameter validation: detect unvalidated tool parameters
  - Custom rules documented trong configuration comments

**AC-6: Configuration file saved và documented**
- **Given** complete gosec configuration
- **When** save configuration file
- **Then** deliverables include:
  - `.gosec.json` in project root với inline comments
  - `docs/security/security-scanning.md` documentation với:
    - Configuration explanation
    - Rule descriptions và rationale
    - Usage examples: `gosec -conf=.gosec.json ./...`
    - VSCode integration instructions
    - CI/CD integration guidance

**AC-7: Configuration validation và testing**
- **Given** gosec configuration file created
- **When** run gosec với configuration
- **Then** validation results:
  - Configuration parses successfully (no syntax errors)
  - All rules load correctly (verify via --list-rules)
  - Scan completes in <2 minutes
  - Output includes findings from Story 1.1 baseline (98 issues)
  - SARIF output validates against SARIF schema 2.1.0

## Tasks / Subtasks

### Task 1: Create base gosec configuration file (AC: AC-1, AC-4)
- [ ] **1.1** Create `.gosec.json` in project root
  - Base structure: `{ "global": {}, "rules": {} }`
  - Enable all rules: G101-G602 (comprehensive security coverage)
  - Set severity threshold: "high" (fail on High và Critical)
  - Set confidence threshold: "medium"
- [ ] **1.2** Configure concurrent scanning
  - Concurrency: 0 (auto-detect CPU cores)
  - Scan paths: `./...` (all packages)
  - Exclude patterns: `["**/vendor/**", "**/testutil/mocks/**"]`
- [ ] **1.3** Verify gosec version compatibility
  - Required: v2.22.10+
  - Current from Story 1.1: v2.22.10 ✅

### Task 2: Configure test file exclusions (AC: AC-2)
- [ ] **2.1** Define test file exclusion rules trong `.gosec.json`
  - Exclude G404 (weak random) cho test files only
  - Exclude G101 (hardcoded credentials) cho test fixtures
  - Path pattern: `**/*_test.go`, `**/testutil/fixtures/**`
- [ ] **2.2** Verify exclusions apply correctly
  - Test: scan production code (`agent/**/*.go`) - full rules
  - Test: scan test code (`agent/**/*_test.go`) - relaxed rules
  - Confirm exclusions don't leak to production code

### Task 3: Configure multiple output formats (AC: AC-3)
- [ ] **3.1** Configure JSON output
  - Default format: JSON với detailed findings
  - Include: CWE ID, severity, confidence, file, line, description
  - Output file: `gosec-report.json`
- [ ] **3.2** Configure SARIF output
  - SARIF schema version: 2.1.0
  - GitHub Security tab compatible
  - Output file: `gosec-results.sarif`
  - Include: rule metadata, locations, fix suggestions
- [ ] **3.3** Configure text output
  - Console-friendly colored output
  - Summary statistics: total issues, by severity
  - Exit codes: 0 (clean), 1 (issues found), 2 (error)
- [ ] **3.4** Document output format usage
  - JSON: `gosec -fmt=json -out=gosec-report.json -conf=.gosec.json ./...`
  - SARIF: `gosec -fmt=sarif -out=gosec-results.sarif -conf=.gosec.json ./...`
  - Text: `gosec -conf=.gosec.json ./...` (default)

### Task 4: Define custom rules cho go-deep-agent (AC: AC-5)
- [ ] **4.1** Analyze go-deep-agent security patterns from Story 1.1 baseline
  - Review 98 findings: identify patterns specific to this project
  - Top CWEs: CWE-703 (72 findings), CWE-276 (12), CWE-22 (8), CWE-338 (3)
  - Prioritize custom rules for high-impact areas
- [ ] **4.2** Create custom rule definitions trong `.gosec.json`
  - API key validation rule: flag usage without validation
  - Secrets logging rule: detect log calls with potential secrets
  - TLS enforcement rule: detect HTTP clients without TLS config
  - Tool validation rule: detect tool.Execute() without parameter validation
- [ ] **4.3** Document custom rules
  - Inline comments trong `.gosec.json` explaining each rule
  - Examples of violations và correct patterns
  - References to architecture.md security patterns

### Task 5: Create security scanning documentation (AC: AC-6)
- [ ] **5.1** Create `docs/security/security-scanning.md`
  - Overview: purpose của gosec configuration
  - Configuration sections explanation
  - Rule descriptions và rationale
- [ ] **5.2** Document usage examples
  - Local development: `gosec -conf=.gosec.json ./...`
  - JSON output: `gosec -fmt=json -out=report.json -conf=.gosec.json ./...`
  - SARIF output: `gosec -fmt=sarif -out=results.sarif -conf=.gosec.json ./...`
  - Fix mode: `gosec -fix -conf=.gosec.json ./...` (future enhancement)
- [ ] **5.3** Document IDE integration
  - VSCode: gosec extension configuration
  - GoLand: External tool setup
  - Pre-commit hook: optional local check before commit
- [ ] **5.4** Document CI/CD integration guidance
  - GitHub Actions integration (teaser for Story 1.4)
  - SARIF upload to GitHub Security tab
  - PR blocking strategy

### Task 6: Validate và test configuration (AC: AC-7)
- [ ] **6.1** Validate configuration syntax
  - Parse `.gosec.json`: `gosec -conf=.gosec.json --list-rules`
  - Verify all rules load correctly
  - Check for configuration errors
- [ ] **6.2** Test full scan with configuration
  - Run: `gosec -conf=.gosec.json ./...`
  - Measure scan duration (target: <2 minutes)
  - Verify findings match Story 1.1 baseline (98 issues expected)
- [ ] **6.3** Test SARIF output validation
  - Run: `gosec -fmt=sarif -out=gosec-results.sarif -conf=.gosec.json ./...`
  - Validate SARIF schema: https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json
  - Test GitHub Security tab upload compatibility
- [ ] **6.4** Test exclusions work correctly
  - Verify test files have G404 excluded
  - Verify production code has all rules active
  - Spot-check 5 files from different packages
- [ ] **6.5** Document actual scan performance
  - Record: scan duration, issues found, CPU usage
  - Compare to Story 1.1 baseline (18 seconds)
  - Confirm <2 minute target met

### Task 7: Commit configuration và documentation (AC: AC-6)
- [ ] **7.1** Review deliverables
  - `.gosec.json`: complete, commented, validated
  - `docs/security/security-scanning.md`: comprehensive, examples included
  - Configuration tested successfully
- [ ] **7.2** Commit to version control
  ```bash
  git add .gosec.json docs/security/security-scanning.md
  git commit -m "security: Configure gosec comprehensive scanning

  - All gosec rules enabled (G101-G602)
  - Severity threshold: High/Critical
  - Test file exclusions: G404, G101
  - Output formats: JSON, SARIF, text
  - Custom rules for go-deep-agent patterns
  - Documentation: docs/security/security-scanning.md
  - Scan performance: <2 minutes (target met)

  Prerequisites: Story 1.1 baseline complete
  Next: Story 1.3 (govulncheck configuration)
  "
  ```

## Dev Notes

### Learnings from Previous Story

**From Story 1.1 (Status: done)**

- **Baseline Metrics Available**: 98 gosec findings (5 critical, 1 high, 20 medium, 72 low)
- **Top CWE IDs**: CWE-703 (Error Handling) 72 findings, CWE-276 (File Permissions) 12, CWE-22 (Path Traversal) 8, CWE-338 (Weak PRNG) 3
- **Scan Performance**: gosec completed in ~18 seconds (well below 2 minute target)
- **Examples Directory Challenge**: Multiple main functions caused build errors in examples/* - excluded from govulncheck in Story 1.1, same exclusion may apply here
- **Tool Versions Confirmed**: gosec v2.22.10, Go 1.25.2 - ready for this story
- **Zero Dependency Vulnerabilities**: All 42 dependencies (8 direct, 34 transitive) clean per govulncheck ✅

**Key Insights for Story 1.2:**
1. **Performance Budget**: Story 1.1 took 18 seconds, we have budget for <2 minutes - configuration overhead minimal
2. **Baseline to Track**: 98 findings from Story 1.1 should be reproduced with .gosec.json configuration to validate config works
3. **Examples Directory**: Consider excluding examples/* from gosec if build issues persist (same pattern as Story 1.1)
4. **Priority CWEs**: Focus custom rules on CWE-703 (error handling) - 73.5% of findings
5. **Test File Pattern**: Story 1.1 showed G404 (weak random) commonly appears in tests - validate exclusion works

**Files Created in Story 1.1:**
- `docs/security/security-baseline-report.md` - Reference for validating configuration reproduces baseline
- `docs/security/` directory - Already exists, reuse for `security-scanning.md`

**Recommended Approach:**
- Start with minimal `.gosec.json` config
- Incrementally add exclusions based on Story 1.1 findings
- Validate each config change reproduces expected baseline
- Document rationale for every exclusion (avoid over-excluding)

[Source: docs/sprint-artifacts/1-1-project-security-audit-baseline-assessment.md#Dev-Agent-Record]

### Architecture Patterns and Constraints

**From Architecture (architecture.md):**

**Epic 1 Guidance:**
- **Tools Stack**: gosec v2.22.10, govulncheck v1.1.4+, golangci-lint v2.6.1
- **Integration**: GitHub Actions with SARIF upload to GitHub Security tab (Story 1.4)
- **Configuration**: `.gosec.json` in project root
- **Documentation**: `docs/security/security-scanning.md`

**Security Patterns:**
- **Additive-only approach**: No changes to existing code, only add configuration
- **Secrets Protection**: RedactHandler pattern for logging (Story 3.2 will implement)
- **Input Validation**: Validation pipeline (Story 2.1-2.4 will implement)
- **TLS Enforcement**: Minimum TLS 1.2 (Story 3.1 will implement)

**Project Structure:**
```
go-deep-agent/
├── .gosec.json                    # 🆕 This story
├── docs/security/
│   ├── security-baseline-report.md # ✅ Story 1.1
│   └── security-scanning.md        # 🆕 This story
└── .github/workflows/
    └── security-scan.yml           # ⏳ Story 1.4
```

**Configuration Standards:**
- **File format**: JSON với inline comments
- **Validation**: gosec --list-rules to verify config loads
- **Documentation**: Every exclusion must be documented với rationale
- **Performance**: <2 minutes scan target (Story 1.1 was 18 seconds baseline)

[Source: docs/architecture.md#Epic-1-Security-Infrastructure-Foundation]

### Source Tree Components

**Files to Create:**
- `.gosec.json` - Main gosec configuration file (project root)
- `docs/security/security-scanning.md` - Configuration documentation

**Files to Reference:**
- `docs/security/security-baseline-report.md` - Story 1.1 baseline for validation
- `docs/architecture.md` - Architecture patterns và security standards

**No Code Changes Required:**
- This story is pure configuration
- No changes to `agent/` packages
- No changes to examples (may exclude from scan if needed)

**Dependencies:**
- gosec v2.22.10+ (already installed in Story 1.1)
- Go 1.25.2 (confirmed in Story 1.1)

### Testing Standards

**Configuration Validation Tests:**

1. **Syntax Validation:**
   - Parse `.gosec.json` successfully
   - Verify: `gosec -conf=.gosec.json --list-rules` succeeds
   - Expected: All rules G101-G602 listed

2. **Rule Enablement Test:**
   - Run: `gosec -conf=.gosec.json ./...`
   - Expected: 98 issues found (matching Story 1.1 baseline)
   - Validate: CWE IDs match baseline (CWE-703, CWE-276, CWE-22, CWE-338)

3. **Exclusion Test:**
   - Scan test file: `gosec -conf=.gosec.json agent/**/*_test.go`
   - Expected: G404 (weak random) excluded for test files
   - Scan production: `gosec -conf=.gosec.json agent/**/*.go`
   - Expected: G404 enforced for production code

4. **SARIF Output Test:**
   - Run: `gosec -fmt=sarif -out=test.sarif -conf=.gosec.json ./...`
   - Validate schema: SARIF 2.1.0 compliant
   - Check: GitHub Security tab compatible format

5. **Performance Test:**
   - Measure: full scan duration with `.gosec.json`
   - Target: <2 minutes
   - Story 1.1 baseline: 18 seconds (expect similar)

**Documentation Quality Test:**
- `security-scanning.md` includes: overview, configuration sections, usage examples, IDE integration
- All code snippets tested và verified
- Links to external references valid (gosec docs, SARIF schema)

### Project Structure Notes

**Alignment with Project Structure:**
- Configuration location: `.gosec.json` (project root, standard gosec convention)
- Documentation: `docs/security/security-scanning.md` (established in Story 1.1)
- No conflicts with existing structure
- Follows Go community conventions (`.gosec.json` is standard)

**Directory Structure After Story:**
```
go-deep-agent/
├── .gosec.json                         # 🆕 NEW
├── docs/
│   └── security/
│       ├── security-baseline-report.md # ✅ Story 1.1
│       └── security-scanning.md        # 🆕 NEW
└── (existing structure unchanged)
```

**gosec Integration Points:**
- **Local Development**: `gosec -conf=.gosec.json ./...` (manual execution)
- **VSCode Extension**: gosec extension auto-detects `.gosec.json`
- **CI/CD**: Story 1.4 will integrate into GitHub Actions
- **golangci-lint**: Story 1.5 will embed gosec within golangci-lint

### References

**Primary Sources:**
1. [Epic 1 Description](../epics.md#Epic-1-Security-Infrastructure-Foundation) - Epic goals và business value
2. [Story 1.2 Details](../epics.md#Story-1.2-Configure-gosec-Comprehensive-Security-Scanning) - Acceptance criteria và technical notes
3. [Architecture - Epic 1](../architecture.md#Epic-1-Security-Infrastructure-Foundation) - Architecture decisions và patterns
4. [Story 1.1 Completion](./1-1-project-security-audit-baseline-assessment.md) - Baseline metrics và learnings

**Tool Documentation:**
- gosec: https://github.com/securego/gosec
- gosec configuration: https://github.com/securego/gosec#configuration
- SARIF specification: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
- GitHub Security tab: https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning

**Related Stories:**
- **Story 1.1** (done): Baseline assessment - provides 98 findings to validate config
- **Story 1.3** (next): govulncheck configuration - parallel security scanning setup
- **Story 1.4** (future): GitHub Actions CI/CD - will use this `.gosec.json` config
- **Story 1.5** (future): golangci-lint - will integrate gosec alongside other linters

## Dev Agent Record

### Context Reference

- [Story 1.2 Technical Context](./1-2-configure-gosec-comprehensive-security-scanning.context.xml) - Generated 2025-11-15

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
