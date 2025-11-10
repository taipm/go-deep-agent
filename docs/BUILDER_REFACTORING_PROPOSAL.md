# Builder.go Refactoring Proposal

**Current Status**: `builder.go` has grown to **1,854 lines** with **74 methods/types**

**Problem**: Single file is becoming difficult to maintain, navigate, and test

**Goal**: Split into focused, maintainable modules while preserving API compatibility

---

## 📊 Current Structure Analysis

### File Metrics
- **Total Lines**: 1,854
- **Methods**: 74 (42 With*, 11 Ask/Stream, 9 Cache, etc.)
- **Cognitive Load**: HIGH (too many responsibilities)

### Functional Groups (by method count)
```
Execution methods:     ~20 methods (Ask, Stream, execute*, build*)
LLM parameters:        ~8 methods (Temperature, TopP, MaxTokens, etc.)
Memory configuration:  ~7 methods (WithMemory, Episodic, Semantic, etc.)
Tool configuration:    ~6 methods (WithTool, Parallel, Workers, etc.)
Cache configuration:   ~9 methods (WithCache, Redis, TTL, Stats, etc.)
Retry/Error handling:  ~5 methods (WithRetry, Backoff, isRetryable, etc.)
Callbacks:             ~3 methods (OnStream, OnToolCall, OnRefusal)
Logging:               ~4 methods (WithLogger, Debug, Info, getLogger)
History/Messages:      ~5 methods (WithMessages, GetHistory, Clear, etc.)
Misc configuration:    ~7 methods (WithTimeout, WithSystem, etc.)
```

### Code Smells Identified

1. **God Object**: Builder does everything (LLM calls, memory, cache, tools, retry, logging)
2. **Feature Envy**: Memory methods should be in memory package
3. **Long Method**: `askWithToolExecution()` is ~137 lines
4. **Primitive Obsession**: Too many individual fields (50+ in Builder struct)
5. **Duplicate Code**: Similar patterns in Ask/AskMultiple/Stream

---

## 🎯 Refactoring Strategy

### Option A: Horizontal Split by Feature (RECOMMENDED ⭐)

**Principle**: Split Builder into feature-focused files, keep single Builder type

**Structure**:
```
agent/
├── builder.go              # Core Builder type + constructors (200 lines)
├── builder_llm.go          # LLM parameters (Temperature, TopP, etc.) (150 lines)
├── builder_memory.go       # Memory configuration methods (150 lines)
├── builder_tools.go        # Tool configuration + parallel execution (200 lines)
├── builder_cache.go        # Cache configuration methods (150 lines)
├── builder_execution.go    # Ask, Stream, execute methods (400 lines)
├── builder_retry.go        # Retry, backoff, error handling (150 lines)
├── builder_callbacks.go    # OnStream, OnToolCall, OnRefusal (100 lines)
├── builder_logging.go      # Logger configuration (100 lines)
└── builder_messages.go     # History, messages management (150 lines)
```

**Pros**:
- ✅ Clear separation of concerns
- ✅ Easy to find specific functionality
- ✅ Zero API changes (all methods still on `*Builder`)
- ✅ Can split tests similarly (builder_llm_test.go, etc.)
- ✅ Backward compatible (same package, same type)

**Cons**:
- ⚠️ Builder struct still in one place (but reduced complexity)
- ⚠️ Some files will have circular dependencies on Builder type

**Migration**: NONE - This is internal refactoring only

---

### Option B: Vertical Split with Composition (ADVANCED)

**Principle**: Compose Builder from smaller, focused components

**Structure**:
```
agent/
├── builder.go              # Main Builder + composition (300 lines)
├── config/
│   ├── llm_config.go       # LLM parameters struct + methods
│   ├── memory_config.go    # Memory configuration
│   ├── tool_config.go      # Tool configuration
│   ├── cache_config.go     # Cache configuration
│   └── retry_config.go     # Retry configuration
└── executor/
    ├── executor.go         # Execution engine interface
    ├── sync_executor.go    # Ask/AskMultiple implementation
    └── stream_executor.go  # Stream implementation
```

**New Builder Structure**:
```go
type Builder struct {
    // Composed configuration objects
    llm      *config.LLMConfig
    memory   *config.MemoryConfig
    tools    *config.ToolConfig
    cache    *config.CacheConfig
    retry    *config.RetryConfig
    
    // Execution engine
    executor executor.Executor
    
    // Minimal fields
    provider Provider
    client   interface{}
}
```

**Pros**:
- ✅ True separation of concerns
- ✅ Each component fully testable in isolation
- ✅ Easy to add new configuration areas
- ✅ Follows composition over inheritance

**Cons**:
- ❌ More complex architecture
- ❌ Potential breaking changes if not careful
- ❌ More files to navigate
- ⚠️ Might need migration guide

**Migration**: Potentially breaking if config structs are exported

---

### Option C: Minimal Split (CONSERVATIVE)

**Principle**: Only extract the most problematic parts

**Structure**:
```
agent/
├── builder.go              # Most of current code (1200 lines)
├── builder_execution.go    # Ask, Stream, execute methods (400 lines)
└── tool_parallel.go        # Already exists (216 lines)
```

**Pros**:
- ✅ Minimal changes
- ✅ Zero risk of breaking changes
- ✅ Easy to implement

**Cons**:
- ❌ Doesn't solve the root problem
- ❌ builder.go still too large (1200 lines)

**Not Recommended** - Doesn't improve maintainability significantly

---

## 🏆 Recommended Approach: Option A (Horizontal Split)

### Phase 1: Preparation (Week 1)
1. **No code changes** - Create plan document (this file)
2. Review with team/stakeholders
3. Set up tracking metrics (file sizes, test coverage)
4. Create test suite to verify no behavior changes

### Phase 2: Extract Execution Logic (Week 2)
**File**: `builder_execution.go` (~400 lines)

**Extract**:
```go
// Move these methods to builder_execution.go
func (b *Builder) Ask(ctx, message) (string, error)
func (b *Builder) askWithToolExecution(ctx, message) (string, error)
func (b *Builder) AskMultiple(ctx, message) ([]string, error)
func (b *Builder) Stream(ctx, message) (string, error)
func (b *Builder) StreamPrint(ctx, message) (string, error)
func (b *Builder) ensureClient() error
func (b *Builder) buildMessages(message) []openai.ChatCompletionMessageParamUnion
func (b *Builder) buildParams(messages) openai.ChatCompletionNewParams
func (b *Builder) executeSyncRaw(ctx, messages) (*openai.ChatCompletion, error)
```

**Benefits**:
- Largest file reduction (~400 lines)
- Clear separation: configuration vs execution
- All execution logic in one place

**Tests**: Create `builder_execution_test.go` with focused execution tests

---

### Phase 3: Extract Feature Configurations (Week 3)

#### 3a. Extract Memory Configuration
**File**: `builder_memory.go` (~150 lines)

```go
// Move these methods:
func (b *Builder) WithMemory() *Builder
func (b *Builder) WithHierarchicalMemory(config) *Builder
func (b *Builder) DisableMemory() *Builder
func (b *Builder) GetMemory() *memory.Memory
func (b *Builder) WithEpisodicMemory(threshold) *Builder
func (b *Builder) WithImportanceWeights(weights) *Builder
func (b *Builder) WithWorkingMemorySize(size) *Builder
func (b *Builder) WithSemanticMemory() *Builder
```

#### 3b. Extract Tool Configuration
**File**: `builder_tools.go` (~200 lines)

```go
// Move these methods:
func (b *Builder) WithTool(tool) *Builder
func (b *Builder) WithTools(tools...) *Builder
func (b *Builder) WithAutoExecute(enable) *Builder
func (b *Builder) WithMaxToolRounds(max) *Builder
func (b *Builder) WithParallelTools(enable) *Builder
func (b *Builder) WithMaxWorkers(max) *Builder
func (b *Builder) WithToolTimeout(timeout) *Builder
func (b *Builder) OnToolCall(callback) *Builder
```

#### 3c. Extract Cache Configuration
**File**: `builder_cache.go` (~150 lines)

```go
// Move these methods:
func (b *Builder) WithCache(cache) *Builder
func (b *Builder) WithMemoryCache(maxSize, ttl) *Builder
func (b *Builder) WithRedisCache(addr, password, db) *Builder
func (b *Builder) WithRedisCacheOptions(opts) *Builder
func (b *Builder) WithCacheTTL(ttl) *Builder
func (b *Builder) DisableCache() *Builder
func (b *Builder) EnableCache() *Builder
func (b *Builder) GetCacheStats() CacheStats
func (b *Builder) ClearCache(ctx) error
```

#### 3d. Extract LLM Parameters
**File**: `builder_llm.go` (~150 lines)

```go
// Move these methods:
func (b *Builder) WithTemperature(temperature) *Builder
func (b *Builder) WithTopP(topP) *Builder
func (b *Builder) WithMaxTokens(maxTokens) *Builder
func (b *Builder) WithPresencePenalty(penalty) *Builder
func (b *Builder) WithFrequencyPenalty(penalty) *Builder
func (b *Builder) WithSeed(seed) *Builder
func (b *Builder) WithLogprobs(enable) *Builder
func (b *Builder) WithTopLogprobs(n) *Builder
func (b *Builder) WithMultipleChoices(n) *Builder
```

---

### Phase 4: Extract Supporting Features (Week 4)

#### 4a. Extract Retry Logic
**File**: `builder_retry.go` (~150 lines)

```go
func (b *Builder) WithRetry(maxRetries) *Builder
func (b *Builder) WithRetryDelay(delay) *Builder
func (b *Builder) WithExponentialBackoff() *Builder
func (b *Builder) executeWithRetry(ctx, operation) error
func (b *Builder) isRetryable(err) bool
func (b *Builder) calculateRetryDelay(attempt) time.Duration
```

#### 4b. Extract Callbacks
**File**: `builder_callbacks.go` (~100 lines)

```go
func (b *Builder) OnStream(callback) *Builder
func (b *Builder) OnToolCall(callback) *Builder
func (b *Builder) OnRefusal(callback) *Builder
```

#### 4c. Extract Logging
**File**: `builder_logging.go` (~100 lines)

```go
func (b *Builder) WithLogger(logger) *Builder
func (b *Builder) WithDebugLogging() *Builder
func (b *Builder) WithInfoLogging() *Builder
func (b *Builder) getLogger() Logger
func (b *Builder) injectLoggerToTools()
```

#### 4d. Extract Message Management
**File**: `builder_messages.go` (~150 lines)

```go
func (b *Builder) WithMessages(messages) *Builder
func (b *Builder) GetHistory() []Message
func (b *Builder) SetHistory(messages) *Builder
func (b *Builder) Clear() *Builder
func (b *Builder) WithMaxHistory(max) *Builder
func (b *Builder) addMessage(message)
```

---

### Phase 5: Core Builder File (Week 5)

**File**: `builder.go` (final ~250 lines)

**Contents**:
```go
// Package declaration and imports
package agent

// Builder struct definition (all fields in one place)
type Builder struct {
    // ~50 fields remain here
    provider Provider
    model    string
    // ... all configuration fields
}

// Constructors
func New(provider, model) *Builder
func NewOpenAI(model, apiKey) *Builder
func NewOllama(model) *Builder

// Basic configuration (doesn't fit elsewhere)
func (b *Builder) WithAPIKey(apiKey) *Builder
func (b *Builder) WithBaseURL(baseURL) *Builder
func (b *Builder) WithSystem(prompt) *Builder
func (b *Builder) WithTimeout(timeout) *Builder

// JSON/Response format
func (b *Builder) WithJSONMode() *Builder
func (b *Builder) WithJSONSchema(...) *Builder
func (b *Builder) WithResponseFormat(format) *Builder
```

---

## 📋 Implementation Checklist

### Pre-Refactoring
- [ ] Create comprehensive test suite for current Builder
- [ ] Document all public APIs
- [ ] Measure baseline metrics (coverage, performance)
- [ ] Create branch: `refactor/builder-split`

### For Each File Split
- [ ] Create new file (e.g., `builder_memory.go`)
- [ ] Move methods (copy, don't delete yet)
- [ ] Add file header comment
- [ ] Run tests - ensure all pass
- [ ] Remove from original `builder.go`
- [ ] Run tests again
- [ ] Create corresponding test file
- [ ] Verify no coverage loss

### Post-Refactoring
- [ ] All tests passing (100% of original tests)
- [ ] No performance regression (benchmarks)
- [ ] Test coverage maintained or improved
- [ ] Documentation updated
- [ ] Code review
- [ ] Merge to main

---

## 🧪 Testing Strategy

### Regression Tests
```go
// Create builder_refactor_test.go
// Test that all original functionality still works

func TestBuilder_AllMethodsPresent(t *testing.T) {
    // Use reflection to verify all methods still exist
}

func TestBuilder_BackwardCompatibility(t *testing.T) {
    // Test all examples from docs still compile and run
}
```

### Coverage Requirements
- Maintain current coverage: **66%**
- Target after refactor: **70%+** (easier to test smaller files)

### Performance Tests
```go
func BenchmarkBuilder_Before(b *testing.B) {
    // Baseline benchmark
}

func BenchmarkBuilder_After(b *testing.B) {
    // Post-refactor benchmark
    // Must be within 5% of baseline
}
```

---

## 📊 Expected Outcomes

### File Size Reduction
```
builder.go:                1854 → 250 lines  (-86%)
builder_execution.go:         0 → 400 lines  (new)
builder_memory.go:            0 → 150 lines  (new)
builder_tools.go:             0 → 200 lines  (new)
builder_cache.go:             0 → 150 lines  (new)
builder_llm.go:               0 → 150 lines  (new)
builder_retry.go:             0 → 150 lines  (new)
builder_callbacks.go:         0 → 100 lines  (new)
builder_logging.go:           0 → 100 lines  (new)
builder_messages.go:          0 → 150 lines  (new)
---
Total:                     1854 → 1800 lines  (slightly less due to reduced comments)
```

### Maintainability Improvements
- ✅ Each file <500 lines (easy to review)
- ✅ Clear responsibility per file
- ✅ Easier to find specific functionality
- ✅ Better test organization
- ✅ Reduced merge conflicts (fewer people editing same file)

### Developer Experience
- ✅ Faster navigation (jump to specific file)
- ✅ Better IDE performance (smaller files)
- ✅ Easier onboarding (clear file names)
- ✅ Focused code reviews (review one feature at a time)

---

## 🚨 Risks & Mitigation

### Risk 1: Breaking Changes
**Likelihood**: LOW  
**Impact**: HIGH  
**Mitigation**:
- All methods stay on `*Builder` type
- Same package (`agent`)
- Comprehensive regression tests
- No changes to public API

### Risk 2: Import Cycles
**Likelihood**: MEDIUM  
**Impact**: LOW  
**Mitigation**:
- All builder_*.go files in same package
- No new imports between them
- Builder struct defined in builder.go only

### Risk 3: Test Maintenance
**Likelihood**: MEDIUM  
**Impact**: MEDIUM  
**Mitigation**:
- Split tests to match file structure
- Use table-driven tests
- Share test fixtures across files

### Risk 4: Lost Functionality
**Likelihood**: LOW  
**Impact**: HIGH  
**Mitigation**:
- Move, don't rewrite
- Run tests after each file split
- Use git to track every change
- Peer review each split

---

## 📅 Timeline

### Conservative Estimate (5 weeks)
- **Week 1**: Planning + setup (this document + test suite)
- **Week 2**: Extract execution logic
- **Week 3**: Extract feature configurations (memory, tools, cache, LLM)
- **Week 4**: Extract supporting features (retry, callbacks, logging, messages)
- **Week 5**: Cleanup, documentation, final testing

### Aggressive Estimate (2 weeks)
- **Week 1**: Extract execution + major features
- **Week 2**: Extract remaining + cleanup

### Recommended: **3 weeks** (balance speed vs risk)

---

## 🎯 Success Criteria

### Must Have (Before Merge)
- ✅ All tests passing (100% of original)
- ✅ Zero API changes (backward compatible)
- ✅ No performance regression (<5% difference)
- ✅ Test coverage maintained (≥66%)
- ✅ All examples still work
- ✅ Documentation updated

### Nice to Have
- ✅ Test coverage improved (>70%)
- ✅ Each file <400 lines
- ✅ New file-specific tests added
- ✅ Code review approved by 2+ people

---

## 🔄 Alternative: Incremental Approach

If full refactor is too risky, we can do incrementally:

### Phase 1 (Low Risk)
1. Extract only `builder_execution.go` (largest file)
2. Test thoroughly
3. Merge to main

### Phase 2 (Medium Risk)
4. Extract `builder_tools.go` + `builder_memory.go`
5. Test thoroughly
6. Merge to main

### Phase 3 (Low Risk)
7. Extract remaining files one by one
8. Merge each independently

**Benefit**: Can stop at any phase if issues arise

---

## 📝 Conclusion

**Recommended**: **Option A - Horizontal Split**

**Why**:
- ✅ Zero breaking changes
- ✅ Clear separation of concerns
- ✅ Easy to implement (3 weeks)
- ✅ Significant maintainability improvement
- ✅ Low risk

**Next Steps**:
1. Review this proposal
2. Get approval from team
3. Create refactoring branch
4. Implement Phase 1 (execution split)
5. Measure results, adjust if needed
6. Continue with remaining phases

**Timeline**: 3 weeks (conservative)  
**Risk Level**: LOW  
**Impact**: HIGH (maintainability)  
**ROI**: Excellent

---

**Document Version**: 1.0  
**Created**: November 10, 2025  
**Author**: AI Assistant  
**Status**: PROPOSAL (awaiting review)
