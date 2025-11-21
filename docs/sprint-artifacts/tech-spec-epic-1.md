# Epic Technical Specification: Security Infrastructure Foundation

Date: 2025-11-14
Author: BMad
Epic ID: epic-1
Status: Draft

---

## Overview

Epic 1 thiết lập nền tảng automated security scanning và monitoring infrastructure cho go-deep-agent, là điều kiện tiên quyết cho tất cả công việc security trong project. Epic này triển khai CI/CD pipelines, security tooling (gosec, govulncheck, golangci-lint), và reporting systems để phát hiện vulnerabilities tự động trong quá trình development.

Mục tiêu chính là chuyển từ security reactive (phát hiện sau khi merge) sang security proactive (phát hiện trước khi merge), đảm bảo zero high/critical vulnerabilities trong production code.

## Objectives and Scope

### In Scope
- ✅ Automated security scanning với gosec (code patterns)
- ✅ Vulnerability scanning với govulncheck (dependencies)
- ✅ Comprehensive linting với golangci-lint (security-focused linters)
- ✅ GitHub Actions CI/CD integration
- ✅ SARIF reports upload to GitHub Security tab
- ✅ PR blocking on High/Critical vulnerabilities
- ✅ Security baseline assessment và tracking
- ✅ Security dashboard và reporting infrastructure

### Out of Scope
- ❌ Actual vulnerability remediation (Epic 2-3 responsibility)
- ❌ Runtime security monitoring
- ❌ Dependency update automation
- ❌ Penetration testing
- ❌ Security training materials

## System Architecture Alignment

Epic 1 tạo ra security scanning layer trong CI/CD pipeline, aligned với architecture decisions:

**Architecture Components Referenced:**
- `.github/workflows/security-scan.yml` - GitHub Actions workflow
- `.gosec.json` - gosec configuration
- `.golangci.yml` - golangci-lint v2 configuration
- `docs/security/` - Security documentation directory
  - `security-baseline-report.md`
  - `security-scanning.md`
  - `vulnerability-response.md`

**Tools Stack:**
- **gosec** v2.22.10 - Code security analyzer
- **govulncheck** v1.1.4+ - Vulnerability scanner
- **golangci-lint** v2.6.1 - Comprehensive linter
- **github-action-upload-sarif** v3 - SARIF report uploader

**Constraints:**
- Must not break existing CI/CD pipelines
- Must complete scans in <10 minutes (CI timeout)
- Must be free for open source projects
- Must integrate with GitHub Security tab

## Detailed Design

### Services and Modules

| Module | Responsibility | Inputs | Outputs | Owner |
|--------|---------------|--------|---------|-------|
| **gosec Scanner** | Analyze Go source code for security issues using pattern matching | `.go` files, `.gosec.json` config | JSON/SARIF reports | Security Engineer |
| **govulncheck Scanner** | Scan dependencies for known CVE vulnerabilities | `go.mod`, `go.sum`, Go binary | JSON vulnerability report | Security Engineer |
| **golangci-lint Runner** | Run multiple linters including security-focused ones | `.go` files, `.golangci.yml` | Lint report (colored/JSON) | DevOps Engineer |
| **SARIF Uploader** | Upload security findings to GitHub Security tab | SARIF files | GitHub Security alerts | CI/CD Pipeline |
| **Baseline Generator** | Generate security baseline reports | All scan results | Markdown reports | Security Engineer |
| **CI/CD Orchestrator** | Coordinate all security scans in GitHub Actions | PR events, push events | Pass/Fail status | DevOps Engineer |

### Data Models and Contracts

**gosec Configuration (`.gosec.json`)**
```json
{
  "severity": "high",
  "confidence": "medium",
  "exclude-generated": true,
  "exclude": [
    "G404"  // Random in tests OK
  ],
  "output": "sarif",
  "concurrency": 4
}
```

**golangci-lint Configuration (`.golangci.yml`)**
```yaml
linters:
  enable:
    - gosec       # Security issues
    - govet       # Suspicious constructs
    - staticcheck # Bugs and performance
    - errcheck    # Unchecked errors
    - exportloopref # Loop variable capture
    - noctx       # HTTP requests without context
    - rowserrcheck # SQL rows.Err() checks
    - sqlclosecheck # SQL rows/statements closed

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

**Security Baseline Report Schema**
```markdown
# Security Baseline Report
**Date:** YYYY-MM-DD
**Commit:** abc123

## Summary
- Total Issues: X
- Critical: Y
- High: Z
- Medium: W
- Low: V

## gosec Findings
[Table of findings with CWE IDs]

## govulncheck Findings
[Table of CVEs with severity]

## Remediation Plan
[Prioritized list of fixes]
```

### APIs and Interfaces

**GitHub Actions Workflow Interface**
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
        run: govulncheck -json ./... > govulncheck-report.json

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v2.6.1
          args: --config=.golangci.yml ./...
```

**Command-Line Interface**
```bash
# Local security scan
./scripts/run-security-scan.sh

# Individual tools
gosec -fmt=sarif -out=gosec.sarif -conf=.gosec.json ./...
govulncheck -json ./... > govulncheck-report.json
golangci-lint run --config=.golangci.yml ./...
```

### Workflows and Sequencing

**Security Scan Workflow (on PR)**
```
1. PR Created/Updated
   ↓
2. GitHub Actions Triggered
   ↓
3. Checkout Code
   ↓
4. Setup Go 1.21+ Environment
   ↓
5. Install Security Tools (gosec, govulncheck, golangci-lint)
   ↓
6. Run gosec (parallel)
   ├─ Scan all .go files
   ├─ Apply .gosec.json rules
   └─ Generate SARIF report
   ↓
7. Upload SARIF to GitHub Security
   ↓
8. Run govulncheck (parallel)
   ├─ Scan go.mod dependencies
   ├─ Check vulnerability database
   └─ Generate JSON report
   ↓
9. Run golangci-lint (parallel)
   ├─ Run security linters
   ├─ Apply .golangci.yml config
   └─ Generate report
   ↓
10. Aggregate Results
    ├─ Count Critical/High issues
    ├─ Generate PR comment
    └─ Determine Pass/Fail
    ↓
11. Report Status
    ├─ [PASS] No High/Critical → Allow merge
    └─ [FAIL] High/Critical found → Block merge, comment on PR
```

**Security Baseline Generation Workflow**
```
1. Trigger: Manual or Weekly Schedule
   ↓
2. Run All Security Scans
   ↓
3. Parse JSON/SARIF Results
   ↓
4. Categorize by Severity (Critical, High, Medium, Low)
   ↓
5. Generate Markdown Report
   ↓
6. Store in docs/security/security-reports/YYYY-MM-DD-report.md
   ↓
7. Update Security Badge in README
```

## Non-Functional Requirements

### Performance

**CI/CD Performance Targets:**
- **Total scan time:** <10 minutes (full security scan suite)
- **gosec scan:** <2 minutes for full codebase
- **govulncheck:** <1 minute typical, <5 minutes worst case
- **golangci-lint:** <5 minutes với all enabled linters
- **SARIF upload:** <30 seconds

**Caching Strategy:**
- Go modules cached (actions/setup-go với cache: true)
- Tool binaries cached (reuse between runs)
- Dependency scan results cached (1 hour TTL)

### Security

**Security Requirements for Security Tools:**
- Tools installed from official sources only
- Tool versions pinned in CI/CD (no @latest in production workflows)
- SARIF reports sanitized (no secrets exposed)
- API tokens secured in GitHub Secrets
- Audit logs for security scan failures

**False Positive Handling:**
- Document known false positives in `.gosec.json` exclude list
- Require justification comment for each exclusion
- Review exclusions quarterly

### Reliability/Availability

**Reliability Targets:**
- **CI/CD uptime:** 99.9% (dependent on GitHub Actions SLA)
- **Scan success rate:** >95% (exclude transient network failures)
- **Zero false negatives:** Critical/High vulnerabilities must be detected

**Failure Handling:**
- Transient failures: Auto-retry up to 3 times
- Persistent failures: Notify team via Slack/email
- Scan timeout: Fail-safe default (mark as failed, manual review)

### Observability

**Logging Requirements:**
- Log all security scan starts/completions
- Log severity distribution (Critical: X, High: Y, etc.)
- Log scan durations for performance monitoring
- Log tool versions used

**Metrics to Track:**
- Number of security scans per day
- Average scan duration
- Number of vulnerabilities by severity (trending)
- Time to remediate vulnerabilities (SLA: Critical 48h, High 7d)
- False positive rate

**Alerting:**
- Critical vulnerabilities: Immediate Slack notification
- High vulnerabilities: Daily digest
- Scan failures: Immediate notification
- Scan duration >10min: Alert (performance degradation)

## Dependencies and Integrations

**External Dependencies:**

| Dependency | Version | Purpose | License | Notes |
|-----------|---------|---------|---------|-------|
| **gosec** | v2.22.10+ | Code security scanner | Apache 2.0 | Pinned via go install |
| **govulncheck** | v1.1.4+ | Vulnerability scanner | BSD-3-Clause | Go official tool |
| **golangci-lint** | v2.6.1 | Comprehensive linter | GPL-3.0 | golangci-lint-action v6 |
| **github-codeql-action/upload-sarif** | v3 | SARIF uploader | MIT | GitHub official action |
| **actions/checkout** | v4 | Code checkout | MIT | GitHub official action |
| **actions/setup-go** | v5 | Go environment | MIT | GitHub official action |

**Integration Points:**

1. **GitHub Actions**
   - Workflow triggers: `pull_request`, `push`, `schedule`
   - Secrets: None required for scanning (tools are public)
   - Permissions: `security-events: write` for SARIF upload

2. **GitHub Security Tab**
   - SARIF reports uploaded via codeql-action
   - Alerts created for findings
   - Integration with Dependabot

3. **Go Module System**
   - govulncheck scans `go.mod`, `go.sum`
   - Requires Go 1.18+ for vulnerability database

4. **Local Development**
   - Scripts in `/scripts/` directory
   - Pre-commit hooks (optional)
   - IDE integration (VSCode gosec extension)

**Dependency Manifest Scan Results (go.mod):**
```
✅ Go version: 1.25.2 (latest stable)
✅ Key dependencies:
  - github.com/openai/openai-go v1.12.0
  - github.com/google/generative-ai-go v0.20.1
  - github.com/stretchr/testify v1.11.1
  - github.com/redis/go-redis/v9 v9.16.0

📋 Total dependencies: 42 (8 direct, 34 transitive)
🔍 Security scan priority: Direct dependencies first
```

## Acceptance Criteria (Authoritative)

**AC-1: gosec Automated Scanning**
- ✅ gosec v2.22.10+ installed trong CI/CD
- ✅ `.gosec.json` configuration file created với all rules enabled
- ✅ gosec runs on every PR và push to main
- ✅ SARIF report generated và uploaded to GitHub Security tab
- ✅ High/Critical findings block PR merge

**AC-2: govulncheck Dependency Scanning**
- ✅ govulncheck v1.1.4+ installed trong CI/CD
- ✅ Scans all direct và transitive dependencies
- ✅ JSON report generated với CVE details
- ✅ Any vulnerability blocks PR merge
- ✅ Vulnerability database updated daily

**AC-3: golangci-lint Integration**
- ✅ golangci-lint v2.6.1 configured với security linters
- ✅ `.golangci.yml` configuration file created
- ✅ Runs in CI/CD with 5 minute timeout
- ✅ Security issues reported với clear error messages
- ✅ IDE integration documented

**AC-4: GitHub Actions CI/CD Pipeline**
- ✅ `.github/workflows/security-scan.yml` created
- ✅ Triggers: pull_request, push (main), schedule (daily)
- ✅ Job timeout: 10 minutes maximum
- ✅ Go modules và tool binaries cached
- ✅ Status checks required for PR merge

**AC-5: SARIF Integration với GitHub Security**
- ✅ SARIF reports uploaded via github/codeql-action
- ✅ Findings visible in GitHub Security tab
- ✅ Alerts created for new vulnerabilities
- ✅ Historical tracking of vulnerabilities

**AC-6: Security Baseline Report**
- ✅ Initial security baseline generated
- ✅ Report saved to `docs/security/security-baseline-report.md`
- ✅ Includes: gosec findings, govulncheck CVEs, severity distribution
- ✅ Findings categorized by CWE ID
- ✅ Remediation plan included

**AC-7: Security Dashboard Infrastructure**
- ✅ GitHub Security tab populated với findings
- ✅ Security badge added to README
- ✅ Historical reports stored in `docs/security/security-reports/`
- ✅ Weekly security summary automated

## Traceability Mapping

| AC | PRD FR | Spec Section | Component | Test Idea |
|----|--------|--------------|-----------|-----------|
| AC-1 | FR1 | APIs and Interfaces → gosec | `.gosec.json`, `security-scan.yml` | Run gosec on test code với known vulnerability, verify detection |
| AC-2 | FR2 | APIs and Interfaces → govulncheck | `security-scan.yml` | Add vulnerable dependency, verify govulncheck detects CVE |
| AC-3 | - | APIs and Interfaces → golangci-lint | `.golangci.yml`, `security-scan.yml` | Introduce code smells, verify golangci-lint catches issues |
| AC-4 | FR1, FR2 | Workflows and Sequencing | `.github/workflows/security-scan.yml` | Trigger workflow on PR, verify all steps execute successfully |
| AC-5 | FR1 | Workflows and Sequencing → SARIF | SARIF upload step | Upload SARIF manually, verify appears in GitHub Security |
| AC-6 | FR1, FR2 | Detailed Design → Baseline | `docs/security/security-baseline-report.md` | Generate baseline, verify all sections populated |
| AC-7 | - | Detailed Design → Dashboard | GitHub Security tab, badges | Check Security tab, verify alerts và trends visible |

## Risks, Assumptions, Open Questions

**Risks:**
1. **Risk:** gosec false positives may be high initially
   - **Impact:** HIGH - Team frustration, exclusion list grows unmanageably
   - **Mitigation:** Start với high confidence only, gradually increase sensitivity
   - **Owner:** Security Engineer

2. **Risk:** govulncheck may find vulnerabilities in transitive dependencies we can't easily update
   - **Impact:** MEDIUM - PR blocking without clear fix path
   - **Mitigation:** Document process for requesting dependency updates, consider temporary exceptions
   - **Owner:** DevOps Engineer

3. **Risk:** CI/CD timeout (10 minutes) may be too aggressive for large PRs
   - **Impact:** LOW - Rare occurrence, manual override available
   - **Mitigation:** Monitor scan durations, adjust timeout if pattern emerges
   - **Owner:** DevOps Engineer

**Assumptions:**
1. **Assumption:** GitHub Actions will remain free for public repositories
   - **Validation:** Confirmed by GitHub pricing as of 2025-11-14
   - **Contingency:** If pricing changes, evaluate self-hosted runners

2. **Assumption:** Team has GitHub write access for SARIF upload
   - **Validation:** Required permission: `security-events: write`
   - **Contingency:** Adjust GitHub repository permissions

3. **Assumption:** Go 1.21+ is acceptable minimum version for tools
   - **Validation:** Current project uses Go 1.25.2
   - **Contingency:** N/A - well above minimum

**Open Questions:**
1. **Question:** Should we enable pre-commit hooks for local scanning?
   - **Status:** OPEN
   - **Decision by:** Sprint Planning
   - **Impact:** Developer workflow

2. **Question:** What is acceptable false positive rate for gosec?
   - **Status:** OPEN
   - **Decision by:** After initial baseline (Story 1.1)
   - **Impact:** .gosec.json configuration

3. **Question:** Should security scans run on every commit or only PR?
   - **Status:** OPEN
   - **Decision by:** DevOps Engineer
   - **Impact:** CI/CD resource usage

## Test Strategy Summary

**Test Levels:**

1. **Unit Tests** (Not applicable - Epic 1 is infrastructure)
   - No unit tests for CI/CD workflows
   - Configuration files tested via actual CI/CD runs

2. **Integration Tests**
   - **Test:** Trigger security scan on test PR
   - **Verify:** All three tools (gosec, govulncheck, golangci-lint) execute
   - **Verify:** SARIF uploaded to GitHub Security
   - **Verify:** Status check appears on PR

3. **E2E Tests**
   - **Test:** Create PR với known vulnerability
   - **Verify:** Security scan detects vulnerability
   - **Verify:** PR merge blocked
   - **Verify:** Alert appears in GitHub Security tab

4. **Manual Tests**
   - **Test:** Run security scans locally via scripts
   - **Test:** Verify gosec configuration excludes test files
   - **Test:** Verify golangci-lint catches security issues

**Edge Cases to Test:**
- Very large PR (1000+ files) → verify timeout handling
- PR with binary files → verify tools skip binaries gracefully
- Network failure during govulncheck → verify retry logic
- Malformed SARIF report → verify error handling

**Coverage Target:**
- 100% of CI/CD workflow steps tested via actual runs
- All configuration files validated (gosec.json, golangci.yml)
- All failure paths tested (timeout, network error, tool crash)

**Frameworks:**
- GitHub Actions for CI/CD testing
- Manual testing for initial baseline generation
- Automated testing post-Epic 7 (when test infrastructure available)

---

**Epic 1 Tech Spec Complete** ✅

**Next Steps:**
1. Review với Security Engineer và DevOps Engineer
2. Validate tooling versions availability
3. Begin Story 1.1: Project Security Audit & Baseline Assessment
