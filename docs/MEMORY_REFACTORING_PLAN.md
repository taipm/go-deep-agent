# Memory System Refactoring Plan v0.9.0

**Date**: November 12, 2025  
**Status**: PLANNING  
**Breaking Changes**: YES (Major version bump recommended)

## 🎯 Vision: Short-Term vs Long-Term Memory

Redesign memory system theo metaphor **não người**:
- **Short-Term Memory**: RAM only, temporary (hiện tại: `WithMemory()`)
- **Long-Term Memory**: Persistent storage (hiện tại: `WithSessionID()`)

---

## 📊 Current State (v0.8.0)

### Existing Memory APIs

```go
// Basic memory (RAM only)
agent.WithMemory()                          // Enable conversation history
agent.DisableMemory()                       // Disable
agent.WithMaxHistory(20)                    // Limit size

// Hierarchical memory (Working → Episodic → Semantic)
agent.WithHierarchicalMemory(config)        // Custom config
agent.WithEpisodicMemory(0.7)               // Enable episodic tier
agent.WithSemanticMemory()                  // Enable semantic tier
agent.WithWorkingMemorySize(10)             // Working memory capacity
agent.WithImportanceWeights(weights)        // Custom importance scoring

// Session persistence (v0.8.0 - NEW)
agent.WithSessionID("user-123")             // Enable persistence
agent.WithMemoryBackend(backend)            // Custom backend
agent.WithAutoSave(true)                    // Control auto-save
agent.SaveSession(ctx)                      // Manual save
agent.LoadSession(ctx)                      // Manual load
agent.DeleteSession(ctx)                    // Delete session
agent.ListSessions(ctx)                     // List all sessions
agent.GetSessionID()                        // Get current ID
```

### Problems

1. **Naming confusion**: "Memory" vs "Session" - không rõ ràng
2. **API inconsistency**: `WithMemory()` (RAM) vs `WithSessionID()` (persistent)
3. **Missing metaphor**: Không có concept rõ ràng về short-term vs long-term
4. **Backward compatibility**: Breaking change nếu đổi tên

---

## 🚀 Target State (v0.9.0)

### New Memory Architecture

```
┌─────────────────────────────────────────────────────┐
│                    MEMORY SYSTEM                     │
├─────────────────────────────────────────────────────┤
│                                                      │
│  SHORT-TERM MEMORY (RAM)                            │
│  ├── Simple: messages[] array                       │
│  ├── Hierarchical: Working → Episodic → Semantic   │
│  └── Lost on program restart                        │
│                                                      │
│  LONG-TERM MEMORY (Persistent)                      │
│  ├── File: ~/.go-deep-agent/memories/{id}.json     │
│  ├── Redis: go-deep-agent:memories:{id}            │
│  ├── PostgreSQL: agent_memories table              │
│  └── Custom: User-defined backend                   │
│                                                      │
└─────────────────────────────────────────────────────┘
```

### New API Design

```go
// ============================================
// SHORT-TERM MEMORY (RAM)
// ============================================

// Basic (simple messages array)
agent.WithShortMemory()                     // Enable RAM memory
agent.WithShortMemory().WithMaxHistory(20)  // Limit size

// Advanced (hierarchical: Working → Episodic → Semantic)
agent.WithHierarchicalMemory(config)        // Full config
agent.WithEpisodicMemory(0.7)               // Enable episodic tier
agent.WithSemanticMemory()                  // Enable semantic tier

// Control
agent.DisableShortMemory()                  // Disable (explicit)

// ============================================
// LONG-TERM MEMORY (Persistent)
// ============================================

// File backend (default)
agent.WithLongMemory("user-alice")          // Auto: FileBackend + auto-save

// Redis backend
agent.WithLongMemory("user-alice").
    UsingRedis("localhost:6379").
        WithPassword("secret").
        WithTTL(24 * time.Hour).
        WithCompression(true)

// PostgreSQL backend
agent.WithLongMemory("user-alice").
    UsingPostgres("postgres://...")

// Custom backend
agent.WithLongMemory("user-alice").
    UsingBackend(myBackend)

// Operations
agent.SaveLongMemory(ctx)                   // Manual save
agent.LoadLongMemory(ctx)                   // Manual reload
agent.DeleteLongMemory(ctx)                 // Delete current
agent.DeleteLongMemoryByID(ctx, "bob")      // Delete specific
agent.ListLongMemories(ctx)                 // List all
agent.GetLongMemoryID()                     // Get current ID

// Control
agent.WithAutoSaveLongMemory(false)         // Disable auto-save
```

---

## 🔄 Migration Path (Breaking Changes)

### Renamed Methods

| v0.8.0 (Old) | v0.9.0 (New) | Breaking? | Migration |
|--------------|--------------|-----------|-----------|
| `WithMemory()` | `WithShortMemory()` | **YES** | Alias for v0.9 |
| `DisableMemory()` | `DisableShortMemory()` | **YES** | Alias |
| `WithSessionID(id)` | `WithLongMemory(id)` | **YES** | Alias |
| `WithMemoryBackend(b)` | `UsingBackend(b)` | **YES** | Method chaining |
| `WithAutoSave(bool)` | `WithAutoSaveLongMemory(bool)` | **YES** | Rename |
| `SaveSession(ctx)` | `SaveLongMemory(ctx)` | **YES** | Rename |
| `LoadSession(ctx)` | `LoadLongMemory(ctx)` | **YES** | Rename |
| `DeleteSession(ctx)` | `DeleteLongMemory(ctx)` | **YES** | Rename |
| `ListSessions(ctx)` | `ListLongMemories(ctx)` | **YES** | Rename |
| `GetSessionID()` | `GetLongMemoryID()` | **YES** | Rename |

### Backward Compatibility Strategy

**Option A: Deprecate (Recommended for v0.9.0)**
```go
// Keep old methods with deprecation warnings
// Deprecated: Use WithShortMemory() instead
func (b *Builder) WithMemory() *Builder {
    if b.logger != nil {
        b.logger.Warn(context.Background(), 
            "WithMemory() is deprecated, use WithShortMemory()")
    }
    return b.WithShortMemory()
}

// Deprecated: Use WithLongMemory() instead
func (b *Builder) WithSessionID(id string) *Builder {
    if b.logger != nil {
        b.logger.Warn(context.Background(),
            "WithSessionID() is deprecated, use WithLongMemory()")
    }
    return b.WithLongMemory(id)
}
```

**Option B: Hard Break (For v1.0.0)**
- Remove old methods completely
- Force users to migrate
- Clean codebase

**Recommendation**: **Option A** for v0.9.0, then **Option B** for v1.0.0

---

## 📝 Implementation Plan

### Phase 1: Core Refactoring (Week 1)

**Goal**: Rename everything, maintain functionality

#### Task 1.1: Rename Methods in `builder_memory.go`
- [ ] `WithMemory()` → `WithShortMemory()`
- [ ] `DisableMemory()` → `DisableShortMemory()`
- [ ] Add deprecation aliases for old names

#### Task 1.2: Rename Session Persistence Methods
- [ ] `WithSessionID()` → `WithLongMemory()`
- [ ] `SaveSession()` → `SaveLongMemory()`
- [ ] `LoadSession()` → `LoadLongMemory()`
- [ ] `DeleteSession()` → `DeleteLongMemory()`
- [ ] `ListSessions()` → `ListLongMemories()`
- [ ] `GetSessionID()` → `GetLongMemoryID()`
- [ ] `WithAutoSave()` → `WithAutoSaveLongMemory()`

#### Task 1.3: Rename Internal Fields in `builder.go`
```go
// Old fields
sessionID     string
memoryBackend MemoryBackend
autoSave      bool

// New fields
longMemoryID      string        // Clearer than sessionID
longMemoryBackend MemoryBackend // Clearer than memoryBackend
autoSaveLongMemory bool         // Explicit
```

#### Task 1.4: Rename File Backend
```go
// Old: ~/.go-deep-agent/sessions/
// New: ~/.go-deep-agent/memories/

func NewFileBackend(basePath string) (*FileBackend, error) {
    if basePath == "" {
        home, _ := os.UserHomeDir()
        basePath = filepath.Join(home, ".go-deep-agent", "memories") // ← Changed
    }
    // ...
}
```

#### Task 1.5: Update Error Messages
```go
// Old
var ErrSessionIDRequired = errors.New("session ID required")

// New
var ErrLongMemoryIDRequired = errors.New("long memory ID required")
var ErrLongMemoryBackendRequired = errors.New("long memory backend required")
```

**Deliverables**:
- [ ] All methods renamed
- [ ] Deprecation aliases added
- [ ] Internal fields renamed
- [ ] Error messages updated
- [ ] Code compiles successfully

---

### Phase 2: Redis Backend Implementation (Week 1-2)

**Goal**: Add `UsingRedis()` fluent API

#### Task 2.1: Create `memory_backend_redis.go`
```go
package agent

import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisBackend struct {
    client      redis.UniversalClient
    prefix      string
    ttl         time.Duration
    compression bool
    maxMessages int
    mu          sync.RWMutex
}

func NewRedisBackend(addr string) (*RedisBackend, error) {
    // Implementation
}

// Fluent API
func (r *RedisBackend) WithPassword(password string) *RedisBackend {
    // Recreate client with password
    return r
}

func (r *RedisBackend) WithTTL(ttl time.Duration) *RedisBackend {
    r.ttl = ttl
    return r
}

func (r *RedisBackend) WithCompression(enabled bool) *RedisBackend {
    r.compression = enabled
    return r
}

func (r *RedisBackend) WithMaxMessages(max int) *RedisBackend {
    r.maxMessages = max
    return r
}

// MemoryBackend interface implementation
func (r *RedisBackend) Save(ctx context.Context, id string, messages []Message) error
func (r *RedisBackend) Load(ctx context.Context, id string) ([]Message, error)
func (r *RedisBackend) Delete(ctx context.Context, id string) error
func (r *RedisBackend) List(ctx context.Context) ([]string, error)
```

#### Task 2.2: Add `UsingRedis()` to Builder
```go
// In builder_memory.go

func (b *Builder) UsingRedis(addr string) *Builder {
    backend, err := NewRedisBackend(addr)
    if err != nil {
        // Log error but don't fail (graceful degradation)
        if b.logger != nil {
            b.logger.Error(context.Background(), 
                "Failed to connect to Redis, falling back to FileBackend",
                F("addr", addr),
                F("error", err))
        }
        // Keep existing backend (likely FileBackend)
        return b
    }
    b.longMemoryBackend = backend
    return b
}

func (b *Builder) UsingBackend(backend MemoryBackend) *Builder {
    b.longMemoryBackend = backend
    return b
}
```

#### Task 2.3: Redis Backend Tests
- [ ] Unit tests: Basic CRUD (15+ tests)
- [ ] Unit tests: TTL behavior, compression, max messages
- [ ] Unit tests: Connection failures, retry logic
- [ ] Integration tests: Builder + Redis (10+ tests)
- [ ] Load tests: 1000 concurrent operations

**Deliverables**:
- [ ] `memory_backend_redis.go` (300+ lines)
- [ ] `memory_backend_redis_test.go` (500+ lines)
- [ ] `UsingRedis()` builder method
- [ ] All tests passing

---

### Phase 3: Update Tests (Week 2)

**Goal**: Update all existing tests to use new API

#### Task 3.1: Update Unit Tests
```bash
# Find all test files using old API
grep -r "WithMemory()" agent/*_test.go
grep -r "WithSessionID" agent/*_test.go
```

**Files to update**:
- [ ] `agent/builder_test.go`
- [ ] `agent/builder_memory_test.go`
- [ ] `agent/memory_backend_test.go`
- [ ] `agent/builder_defaults_test.go`
- [ ] All other test files

#### Task 3.2: Update Examples
- [ ] `examples/session_persistence_basic.go`
- [ ] All other examples using memory

#### Task 3.3: Test Backward Compatibility
- [ ] Verify old API still works (with warnings)
- [ ] Verify deprecation warnings appear in logs

**Deliverables**:
- [ ] All tests updated to new API
- [ ] All examples updated
- [ ] Backward compat tests passing
- [ ] Full test suite passing (1040+ tests)

---

### Phase 4: Documentation (Week 2)

**Goal**: Update all documentation

#### Task 4.1: Update Core Docs
- [ ] `README.md` - Replace "Session" with "LongMemory"
- [ ] `CHANGELOG.md` - Add v0.9.0 breaking changes section
- [ ] `MIGRATION_v0.9.md` (NEW) - Migration guide

#### Task 4.2: Update Memory Guides
- [ ] `docs/MEMORY_SYSTEM_GUIDE.md` - Update terminology
- [ ] `docs/MEMORY_PERSISTENCE_ROADMAP.md` - Update to "LongMemory"
- [ ] `docs/SESSION_ID_EXPLAINED.md` → `docs/LONG_MEMORY_EXPLAINED.md`
- [ ] `docs/REDIS_BACKEND_DESIGN.md` - Update API examples

#### Task 4.3: Create New Docs
- [ ] `docs/SHORT_VS_LONG_MEMORY.md` (NEW) - Concept guide
- [ ] `docs/REDIS_BACKEND_GUIDE.md` (NEW) - Complete Redis guide

#### Task 4.4: Update Release Notes
- [ ] `RELEASE_NOTES_v0.9.0.md` (NEW) - Feature list + breaking changes

**Deliverables**:
- [ ] All docs updated
- [ ] Migration guide complete
- [ ] New concept guides created

---

### Phase 5: Examples & Tutorials (Week 3)

**Goal**: Show users how to use new API

#### Task 5.1: Update Existing Examples
- [ ] `examples/session_persistence_basic.go` → `examples/long_memory_basic.go`
- [ ] Update all imports and API calls
- [ ] Test all examples compile and run

#### Task 5.2: Create New Examples
- [ ] `examples/short_memory_simple.go` - Basic short-term memory
- [ ] `examples/long_memory_file.go` - File backend
- [ ] `examples/long_memory_redis.go` - Redis backend
- [ ] `examples/long_memory_custom.go` - Custom backend (S3 example)
- [ ] `examples/memory_migration.go` - Old → New API

#### Task 5.3: Create Quickstart Tutorial
- [ ] `docs/QUICKSTART_MEMORY.md` - Step-by-step guide

**Deliverables**:
- [ ] 5+ new examples
- [ ] All examples tested
- [ ] Quickstart tutorial

---

## 📊 Breaking Changes Summary

### API Changes

```go
// ====================================
// SHORT-TERM MEMORY
// ====================================

// OLD → NEW
WithMemory()           → WithShortMemory()
DisableMemory()        → DisableShortMemory()

// UNCHANGED (no concept change)
WithHierarchicalMemory(config)
WithEpisodicMemory(threshold)
WithSemanticMemory()
WithMaxHistory(size)

// ====================================
// LONG-TERM MEMORY
// ====================================

// OLD → NEW
WithSessionID(id)      → WithLongMemory(id)
WithMemoryBackend(b)   → UsingBackend(b)  // Now fluent API
WithAutoSave(bool)     → WithAutoSaveLongMemory(bool)
SaveSession(ctx)       → SaveLongMemory(ctx)
LoadSession(ctx)       → LoadLongMemory(ctx)
DeleteSession(ctx)     → DeleteLongMemory(ctx)
ListSessions(ctx)      → ListLongMemories(ctx)
GetSessionID()         → GetLongMemoryID()

// NEW (Redis backend)
UsingRedis(addr)                          → NEW
  .WithPassword(pass)                     → NEW
  .WithTTL(duration)                      → NEW
  .WithCompression(bool)                  → NEW
  .WithMaxMessages(int)                   → NEW
```

### File Structure Changes

```
OLD:
~/.go-deep-agent/sessions/
  ├── user-alice.json
  └── user-bob.json

NEW:
~/.go-deep-agent/memories/
  ├── user-alice.json
  └── user-bob.json
```

### Error Names

```go
// OLD → NEW
ErrSessionIDRequired        → ErrLongMemoryIDRequired
ErrMemoryBackendRequired    → ErrLongMemoryBackendRequired
```

---

## 🎯 Success Metrics

### Code Quality
- [ ] All tests passing (1040+ tests)
- [ ] Code coverage maintained (>70%)
- [ ] No performance regression
- [ ] Memory leaks checked

### Documentation
- [ ] 100% API documented
- [ ] Migration guide complete
- [ ] 5+ working examples
- [ ] Concept guide clear

### User Experience
- [ ] Deprecation warnings clear
- [ ] Migration path straightforward
- [ ] Breaking changes documented
- [ ] New API more intuitive

---

## 🚦 Release Plan

### v0.9.0-beta.1 (Week 1)
- Core refactoring done
- Deprecation aliases working
- Basic tests passing
- **Beta testing with early adopters**

### v0.9.0-beta.2 (Week 2)
- Redis backend complete
- All tests updated
- Documentation 80% done
- **Extended beta testing**

### v0.9.0 (Week 3)
- All documentation complete
- Examples working
- Migration guide ready
- **Official release**

### v1.0.0 (Future)
- Remove deprecated methods
- Clean codebase
- Full stability

---

## ⚠️ Risks & Mitigation

### Risk 1: Breaking Changes Upset Users
**Mitigation**:
- ✅ Provide deprecation warnings (not hard errors)
- ✅ Maintain old API for 1-2 versions
- ✅ Clear migration guide
- ✅ Announce early on GitHub/Discord

### Risk 2: Redis Dependency Issues
**Mitigation**:
- ✅ Optional dependency (graceful degradation)
- ✅ Fallback to FileBackend on connection failure
- ✅ Clear error messages

### Risk 3: Test Suite Breakage
**Mitigation**:
- ✅ Update tests incrementally (file by file)
- ✅ Run tests after each change
- ✅ Use Git branches for safety

### Risk 4: Documentation Out of Sync
**Mitigation**:
- ✅ Update docs alongside code
- ✅ Use examples as tests
- ✅ CI/CD checks for broken links

---

## 📋 Checklist - Complete Refactoring

### Week 1: Core Refactoring
- [ ] Rename all methods in `builder_memory.go`
- [ ] Rename internal fields in `builder.go`
- [ ] Add deprecation aliases
- [ ] Update error messages
- [ ] File backend: sessions/ → memories/
- [ ] Code compiles successfully
- [ ] Run existing tests (should pass with warnings)

### Week 2: Redis + Tests
- [ ] Implement `memory_backend_redis.go`
- [ ] Add `UsingRedis()` builder method
- [ ] Write Redis backend tests (15+ unit, 10+ integration)
- [ ] Update all existing tests to new API
- [ ] Update all examples
- [ ] Full test suite passing

### Week 3: Documentation + Release
- [ ] Update README.md, CHANGELOG.md
- [ ] Create MIGRATION_v0.9.md
- [ ] Update memory guides
- [ ] Create new examples
- [ ] Write release notes
- [ ] Beta release v0.9.0-beta.1

### Week 4: Beta Testing + Final Release
- [ ] Collect feedback from beta users
- [ ] Fix issues found in beta
- [ ] Final documentation review
- [ ] Release v0.9.0 🚀

---

## 🎓 Learning from Refactoring

### What Went Well (Predicted)
- ✅ Clear metaphor (short-term vs long-term)
- ✅ Backward compatibility maintained
- ✅ Fluent API more intuitive

### What Could Be Better
- ⚠️ Breaking changes painful for users
- ⚠️ Documentation work intensive
- ⚠️ Testing work intensive

### Future Improvements
- 🔮 Consider AI-powered migration tool
- 🔮 Automated example validation in CI
- 🔮 Interactive migration wizard

---

## 🏁 Next Steps

**Immediate Action (Now)**:
1. ✅ Review this plan
2. ✅ Confirm breaking changes acceptable
3. ✅ Decide: v0.9.0 or v1.0.0?
4. ✅ Start implementation (Task 1.1)

**Questions for You**:
1. **Version number**: v0.9.0 (backward compat) or v1.0.0 (clean break)?
2. **Timeline**: 3 weeks OK? Or need faster/slower?
3. **Priority**: Redis backend in v0.9.0 or later?
4. **Testing**: Need help with beta testing? Discord server?

Bạn sẵn sàng bắt đầu chưa? 🚀
