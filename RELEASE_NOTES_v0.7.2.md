# Release v0.7.2 - Module Publishing Fix 🔧

## Overview

This is a **hotfix release** that fixes the module publishing issue in v0.7.1. Version v0.7.1 has been **retracted** and should not be used.

## Problem Fixed

Tag v0.7.1 contained a file with an invalid name (`"Quality\n"` with quotes and newline character) that prevented the Go module proxy from creating a zip file:

```
not found: create zip: Quality
: malformed file path "Quality\n": invalid char '\n'
```

This made v0.7.1 inaccessible via `go get`.

## Solution

1. ✅ Removed the problematic tag v0.7.1
2. ✅ Added `retract v0.7.1` directive to go.mod
3. ✅ Created clean tag v0.7.2 without invalid files
4. ✅ Updated documentation and changelog

## Installation

```bash
# Install or update to v0.7.2
go get github.com/taipm/go-deep-agent@v0.7.2

# Or if you encounter cached errors, use direct mode
GOPROXY=direct go get github.com/taipm/go-deep-agent@v0.7.2
```

## Verification

```bash
# Verify version is accessible
go list -m github.com/taipm/go-deep-agent@v0.7.2
# Output: github.com/taipm/go-deep-agent v0.7.2

# Check available versions
GOPROXY=direct go list -m -versions github.com/taipm/go-deep-agent
# Output: v0.3.0 v0.5.0 v0.5.1 v0.5.2 v0.5.7 v0.6.0 v0.6.1 v0.6.2 v0.6.3 v0.6.4 v0.7.0 v0.7.2
```

## Features

This release includes all features from v0.7.1 (Planning Layer) with no functional changes:

### 🧩 Planning Layer - Goal-Oriented Workflows

- **Automatic Decomposition**: LLM-powered goal → task breakdown
- **Dependency Management**: Direct, transitive, diamond patterns with cycle detection
- **3 Execution Strategies**:
  - Sequential: One task at a time, deterministic order
  - Parallel: Topological sort with semaphore-based concurrency (MaxParallel limit)
  - Adaptive: Dynamic strategy switching based on performance metrics
- **Goal-Oriented**: Early termination when success criteria met
- **Performance Monitoring**: Timeline events, metrics (TasksPerSec, AvgLatency, ParallelEfficiency)

### 📚 Documentation

- [Planning Guide](docs/PLANNING_GUIDE.md) - Comprehensive concepts and patterns
- [Planning API](docs/PLANNING_API.md) - Complete API reference
- [Planning Performance](docs/PLANNING_PERFORMANCE.md) - Benchmarks and tuning
- [CHANGELOG](CHANGELOG.md) - Full version history

### 📦 Example Usage

```go
// High-level API - automatic planning and execution
result, _ := agent.NewOpenAI("gpt-4o", apiKey).
    PlanAndExecute(ctx, "Research AI trends and write a report")

// Advanced: Manual control with custom strategies
plan := agent.NewPlan("ETL Pipeline", agent.StrategyParallel)
plan.AddTask(agent.Task{ID: "extract-1", Description: "Extract from DB1"})
plan.AddTask(agent.Task{ID: "extract-2", Description: "Extract from DB2"})
plan.AddTask(agent.Task{
    ID:           "transform",
    Description:  "Transform combined data",
    Dependencies: []string{"extract-1", "extract-2"},
})

config := agent.DefaultPlannerConfig()
config.MaxParallel = 10
config.Strategy = agent.StrategyAdaptive

executor := agent.NewExecutor(config, aiAgent)
result, _ := executor.Execute(ctx, plan)
```

## Migration from v0.7.1

If you were using v0.7.1 (or attempted to), simply update to v0.7.2:

```bash
# Update go.mod
go get github.com/taipm/go-deep-agent@v0.7.2

# Or clean cache and reinstall
go clean -modcache
go get github.com/taipm/go-deep-agent@v0.7.2
```

No code changes required - v0.7.2 is functionally identical to v0.7.1.

## What's Next

See our [CHANGELOG](CHANGELOG.md) for:
- v0.8.0: Enhanced observability & metrics (planned)
- Long-term roadmap: Production features focus

## Links

- **Repository**: https://github.com/taipm/go-deep-agent
- **Documentation**: https://github.com/taipm/go-deep-agent#readme
- **Issues**: https://github.com/taipm/go-deep-agent/issues
- **Changelog**: https://github.com/taipm/go-deep-agent/blob/main/CHANGELOG.md

---

**Full Changelog**: https://github.com/taipm/go-deep-agent/compare/v0.7.0...v0.7.2

**Note**: v0.7.1 is retracted and should not be used.
