# Release Notes v0.8.0 - Session Persistence

## 🎉 Major Feature: Session Persistence

This release introduces **persistent conversation memory** through the new session persistence system, allowing conversations to survive beyond single program executions.

## ✨ New Features

### Session Persistence System

**Core Components:**
- `MemoryBackend` interface for pluggable storage backends
- `FileBackend` implementation with atomic writes and thread-safety
- Default storage: `~/.go-deep-agent/sessions/{sessionID}.json`

**Builder API Methods:**
```go
// Enable session persistence (auto-loads existing session)
WithSessionID(sessionID string) *Builder

// Use custom storage backend (optional - FileBackend by default)
WithMemoryBackend(backend MemoryBackend) *Builder

// Control auto-save behavior (enabled by default)
WithAutoSave(enabled bool) *Builder

// Manual session management
SaveSession(ctx context.Context) error
LoadSession(ctx context.Context) error
DeleteSession(ctx context.Context) error
ListSessions(ctx context.Context) ([]string, error)
GetSessionID() string
```

### Usage Example

**Simple Session Persistence:**
```go
agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().
    WithSessionID("user-alice") // Auto-load + auto-save enabled

// First conversation
response, _ := agent.Ask(ctx, "What's the capital of France?")
// Automatically saved to ~/.go-deep-agent/sessions/user-alice.json

// Later (new program execution)
agent2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().
    WithSessionID("user-alice") // Auto-loads previous conversation

// Agent remembers!
response, _ := agent2.Ask(ctx, "What did I just ask you?")
```

**Multiple Users:**
```go
// User Alice
aliceAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().
    WithSessionID("user-alice")

// User Bob (separate session)
bobAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().
    WithSessionID("user-bob")
```

**Manual Control:**
```go
agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().
    WithSessionID("user-charlie").
    WithAutoSave(false) // Disable auto-save

// Manually save when needed
agent.SaveSession(ctx)

// Manually load
agent.LoadSession(ctx)

// Delete session
agent.DeleteSession(ctx)

// List all sessions
sessions, _ := agent.ListSessions(ctx)
```

### Key Features

**1. Auto-Save by Default**
- Automatically saves after `Ask()` and `Stream()` calls
- Async operation (non-blocking)
- 5-second timeout per save
- Graceful error handling with logging

**2. Auto-Load on Initialization**
- `WithSessionID()` automatically loads existing session
- New sessions start with empty history
- Graceful degradation on load failures

**3. Thread-Safe**
- `sync.RWMutex` protection for concurrent access
- Safe for multiple goroutines
- Atomic file writes (temp file + rename)

**4. Zero Dependencies**
- File-based storage by default
- No external database required
- Works out-of-the-box

**5. Backward Compatible**
- Existing code without `WithSessionID()` works unchanged
- In-memory conversations still supported
- No breaking changes

**6. Pluggable Backends**
- Implement `MemoryBackend` interface for custom storage
- Easy integration with Redis, PostgreSQL, S3, etc.

## 🧪 Testing

**Comprehensive Test Suite (28 Tests):**

**FileBackend Unit Tests (12 tests):**
- ✅ Basic CRUD operations
- ✅ Concurrent save/load (stress tested)
- ✅ Edge cases (empty IDs, overwrites, large conversations)
- ✅ Error handling (corrupted JSON, non-existent sessions)

**Builder Integration Tests (16 tests):**
- ✅ `WithSessionID()` initialization and auto-load
- ✅ Auto-save after `Ask()` and `Stream()`
- ✅ Manual `SaveSession()`, `LoadSession()`, `DeleteSession()`
- ✅ `ListSessions()` functionality
- ✅ Backward compatibility (without session ID)
- ✅ Custom backend support
- ✅ Method chaining
- ✅ Concurrent access handling
- ✅ Context timeout handling

**Full Test Suite:**
- All 1325 test runs pass ✅
- All 28 session persistence tests pass ✅
- Test duration: ~17s

## 📚 Documentation

**New Files:**
- `examples/session_persistence_basic.go` - 5 working examples
- `agent/memory_backend.go` - Backend implementation (220 lines)
- `agent/memory_backend_test.go` - Unit tests (480 lines)

**Modified Files:**
- `agent/builder.go` - Added 3 fields
- `agent/builder_memory.go` - Added 8 session API methods (~200 lines)
- `agent/builder_execution.go` - Added auto-save hooks (~30 lines)
- `agent/builder_memory_test.go` - Added 16 integration tests (~300 lines)
- `agent/errors.go` - Added 2 error constants

## 🔧 Technical Details

**MemoryBackend Interface:**
```go
type MemoryBackend interface {
    Save(ctx context.Context, sessionID string, messages []Message) error
    Load(ctx context.Context, sessionID string) ([]Message, error)
    Delete(ctx context.Context, sessionID string) error
    List(ctx context.Context) ([]string, error)
}
```

**FileBackend Implementation:**
- **Storage Path**: `~/.go-deep-agent/sessions/`
- **File Format**: Pretty-printed JSON
- **Thread Safety**: `sync.RWMutex`
- **Atomic Writes**: Temp file + `os.Rename()`
- **Auto-Create**: Creates directories automatically

**Auto-Save Hook:**
```go
// In Ask() and Stream() methods
if b.autoSave && b.sessionID != "" && b.memoryBackend != nil {
    go func() {
        saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := b.SaveSession(saveCtx); err != nil {
            b.logger.Error(ctx, "Failed to auto-save session", F("error", err))
        }
    }()
}
```

## 🚀 Migration Guide

**Existing Code (No Changes Required):**
```go
agent := agent.NewOpenAI("gpt-4o-mini", apiKey).WithMemory()
// Works exactly as before (in-memory only)
```

**Enable Session Persistence (One Line):**
```go
agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithMemory().
    WithSessionID("user-123") // That's it!
```

## 🎯 Use Cases

1. **Chatbots**: Maintain conversation context across user sessions
2. **Customer Support**: Resume support conversations seamlessly
3. **Interactive Tutorials**: Remember user progress
4. **Personal Assistants**: Long-term memory of user preferences
5. **Multi-User Applications**: Separate conversation history per user

## 🔐 Privacy & Storage

**Default Storage Location:**
```
~/.go-deep-agent/sessions/
├── user-alice.json
├── user-bob.json
└── session-12345.json
```

**Important Notes:**
- Sessions stored in plain text JSON
- Contains full conversation history
- Consider encryption for sensitive data
- Use custom backend for secure storage

## ⚠️ Breaking Changes

**None** - Fully backward compatible

## 🐛 Bug Fixes

None - New feature release

## 📊 Performance

**FileBackend Performance:**
- Save operation: <1ms (async)
- Load operation: <1ms
- List operations: <1ms for typical use
- Concurrent save stress test: 200 saves in 50ms
- Large conversations: 1000 messages handled efficiently

**Auto-Save Overhead:**
- Async operation (non-blocking)
- 5-second timeout prevents hanging
- Graceful degradation on failures

## 🔮 Future Enhancements (Phase 2 & 3)

**Phase 2 - Advanced Backends:**
- Redis backend for distributed systems
- PostgreSQL/MySQL backend for production
- S3 backend for cloud storage
- MongoDB backend for document storage

**Phase 3 - Advanced Features:**
- Session branching/forking
- Session tagging/metadata
- Time-based expiration
- Session encryption
- Compression for large histories
- Session search/query

## 🙏 Credits

Implemented based on:
- `docs/MEMORY_PERSISTENCE_ROADMAP.md` - Comprehensive 3-phase LEAN roadmap
- `docs/SESSION_ID_EXPLAINED.md` - Deep-dive explanation

## 📝 Summary

v0.8.0 introduces **session persistence**, enabling conversations to persist across program executions with:

- ✅ Simple API (`WithSessionID()`)
- ✅ Auto-save by default
- ✅ Zero dependencies (file-based)
- ✅ Thread-safe
- ✅ Backward compatible
- ✅ Pluggable backends
- ✅ Comprehensive tests (28/28 passing)

**Upgrade now to give your AI agents long-term memory! 🧠**

---

**Installation:**
```bash
go get github.com/taipm/go-deep-agent@v0.8.0
```

**Example:**
```bash
go run examples/session_persistence_basic.go
```
