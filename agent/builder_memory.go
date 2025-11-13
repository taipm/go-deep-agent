package agent

import (
	"context"

	"github.com/taipm/go-deep-agent/agent/memory"
)

// Memory configuration methods for Builder
// This file contains all methods related to memory management,
// including hierarchical memory, episodic memory, semantic memory,
// and long-term memory persistence (v0.9.0+).
//
// Memory Architecture:
// - SHORT-TERM MEMORY: Stored in RAM, lost on restart (WithShortMemory)
// - LONG-TERM MEMORY: Persistent storage across restarts (WithLongMemory)

// ============================================================================
// SHORT-TERM MEMORY (RAM-based)
// ============================================================================

// WithShortMemory enables short-term conversation memory stored in RAM.
// Messages are kept in memory during program execution but lost on restart.
//
// For persistent storage across restarts, use WithLongMemory().
//
// Example:
//
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().                    // Enable RAM memory
//	    WithMaxHistory(20)                    // Keep last 20 messages
//
// See also: WithLongMemory() for persistent storage
func (b *Builder) WithShortMemory() *Builder {
	b.autoMemory = true
	return b
}

// WithMemory is deprecated. Use WithShortMemory() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use WithShortMemory() for RAM-based memory or WithLongMemory() for persistent storage.
func (b *Builder) WithMemory() *Builder {
	if b.logger != nil {
		ctx := context.Background()
		b.logger.Warn(ctx, "WithMemory() is deprecated, use WithShortMemory() for RAM memory or WithLongMemory() for persistent storage")
	}
	return b.WithShortMemory()
}

func (b *Builder) WithHierarchicalMemory(config memory.MemoryConfig) *Builder {
	b.memory = memory.NewWithConfig(config)
	b.memoryEnabled = true
	return b
}

// DisableShortMemory disables short-term memory.
// Messages will not be kept in RAM.
func (b *Builder) DisableShortMemory() *Builder {
	b.memoryEnabled = false
	return b
}

// DisableMemory is deprecated. Use DisableShortMemory() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use DisableShortMemory() instead.
func (b *Builder) DisableMemory() *Builder {
	if b.logger != nil {
		ctx := context.Background()
		b.logger.Warn(ctx, "DisableMemory() is deprecated, use DisableShortMemory()")
	}
	return b.DisableShortMemory()
}

func (b *Builder) GetMemory() *memory.Memory {
	return b.memory
}

func (b *Builder) WithEpisodicMemory(threshold float64) *Builder {
	if b.memory == nil {
		b.memory = memory.New()
	}
	config := b.memory.GetConfig()
	config.EpisodicEnabled = true
	config.EpisodicThreshold = threshold
	config.ImportanceScoring = true
	_ = b.memory.SetConfig(config) // Error only occurs if memory is nil, which we check above
	b.memoryEnabled = true
	return b
}

func (b *Builder) WithImportanceWeights(weights memory.ImportanceWeights) *Builder {
	if b.memory == nil {
		b.memory = memory.New()
	}
	config := b.memory.GetConfig()
	config.ImportanceWeights = weights
	config.ImportanceScoring = true
	_ = b.memory.SetConfig(config)
	b.memoryEnabled = true
	return b
}

func (b *Builder) WithWorkingMemorySize(size int) *Builder {
	if b.memory == nil {
		b.memory = memory.New()
	}
	config := b.memory.GetConfig()
	config.WorkingCapacity = size
	_ = b.memory.SetConfig(config)
	b.memoryEnabled = true
	return b
}

func (b *Builder) WithSemanticMemory() *Builder {
	if b.memory == nil {
		b.memory = memory.New()
	}
	config := b.memory.GetConfig()
	config.SemanticEnabled = true
	_ = b.memory.SetConfig(config)
	b.memoryEnabled = true
	return b
}

// ============================================================================
// LONG-TERM MEMORY (Persistent Storage) - v0.9.0+
// ============================================================================

// WithLongMemory enables long-term persistent memory with a unique identifier.
// Conversations are automatically saved to disk/Redis and restored across program restarts.
//
// Default backend: FileBackend (~/.go-deep-agent/memories/)
// Default auto-save: enabled
//
// Example - File backend (default):
//
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().                   // Enable RAM memory
//	    WithLongMemory("user-alice")         // Enable persistence
//
// Example - Redis backend:
//
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().
//	    WithLongMemory("user-alice").
//	        UsingRedis("localhost:6379")
//
// The agent will:
//  1. Check if "user-alice" memory exists in storage
//  2. If yes, load previous conversation history
//  3. After each Ask()/Stream(), auto-save to storage
//
// See also: SaveLongMemory(), LoadLongMemory(), DeleteLongMemory()
func (b *Builder) WithLongMemory(id string) *Builder {
	b.longMemoryID = id

	// Initialize default backend if not set
	if b.longMemoryBackend == nil {
		backend, err := NewFileBackend("")
		if err != nil {
			// Log error but don't fail - gracefully degrade to in-memory
			if b.logger != nil {
				ctx := context.Background()
				b.logger.Error(ctx, "Failed to initialize default file backend",
					F("error", err.Error()),
					F("fallback", "in-memory mode"))
			}
			return b
		}
		b.longMemoryBackend = backend
	}

	// Enable auto-save by default
	if !b.autoSaveLongMemory {
		b.autoSaveLongMemory = true
	}

	// Auto-load existing memory if available
	if id != "" && b.longMemoryBackend != nil {
		ctx := context.Background()
		messages, err := b.longMemoryBackend.Load(ctx, id)
		if err != nil {
			// Log error but don't fail - start with empty history
			if b.logger != nil {
				b.logger.Error(ctx, "Failed to load long-term memory",
					F("memory_id", id),
					F("error", err.Error()),
					F("fallback", "empty history"))
			}
		} else if messages != nil {
			// Successfully loaded previous memory
			b.messages = messages
			if b.logger != nil {
				b.logger.Info(ctx, "Long-term memory loaded",
					F("memory_id", id),
					F("message_count", len(messages)))
			}
		}
	}

	return b
}

// WithSessionID is deprecated. Use WithLongMemory() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use WithLongMemory() for persistent conversation storage.
func (b *Builder) WithSessionID(sessionID string) *Builder {
	if b.logger != nil {
		ctx := context.Background()
		b.logger.Warn(ctx, "WithSessionID() is deprecated, use WithLongMemory() instead")
	}
	return b.WithLongMemory(sessionID)
}

// UsingBackend sets a custom backend for long-term memory storage.
// Use this with WithLongMemory() to specify custom storage.
//
// Available backends:
//   - FileBackend: Local file storage (default)
//   - RedisBackend: Distributed cache (v0.9.0+)
//   - PostgresBackend: Database storage (future)
//   - Custom: Implement MemoryBackend interface
//
// Example with custom file path:
//
//	backend, _ := NewFileBackend("/custom/path")
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().
//	    WithLongMemory("user-alice").
//	        UsingBackend(backend)
//
// Example with custom implementation:
//
//	type MyS3Backend struct {}
//	func (m *MyS3Backend) Load(...) ([]Message, error) { ... }
//	func (m *MyS3Backend) Save(...) error { ... }
//	func (m *MyS3Backend) Delete(...) error { ... }
//	func (m *MyS3Backend) List(...) ([]string, error) { ... }
//
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().
//	    WithLongMemory("user-alice").
//	        UsingBackend(&MyS3Backend{})
func (b *Builder) UsingBackend(backend MemoryBackend) *Builder {
	b.longMemoryBackend = backend
	return b
}

// WithMemoryBackend is deprecated. Use UsingBackend() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use UsingBackend() for a more fluent API.
func (b *Builder) WithMemoryBackend(backend MemoryBackend) *Builder {
	if b.logger != nil {
		ctx := context.Background()
		b.logger.Warn(ctx, "WithMemoryBackend() is deprecated, use UsingBackend() instead")
	}
	return b.UsingBackend(backend)
}

// WithAutoSaveLongMemory controls whether conversation history is automatically saved
// after each message. Default: true when using WithLongMemory().
//
// Use WithAutoSaveLongMemory(false) for manual control:
//
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().
//	    WithLongMemory("user-alice").
//	    WithAutoSaveLongMemory(false)  // Manual save mode
//
//	agent.Ask(ctx, "Hello")
//	agent.SaveLongMemory(ctx)  // Explicit save
func (b *Builder) WithAutoSaveLongMemory(enabled bool) *Builder {
	b.autoSaveLongMemory = enabled
	return b
}

// WithAutoSave is deprecated. Use WithAutoSaveLongMemory() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use WithAutoSaveLongMemory() for clearer intent.
func (b *Builder) WithAutoSave(enabled bool) *Builder {
	if b.logger != nil {
		ctx := context.Background()
		b.logger.Warn(ctx, "WithAutoSave() is deprecated, use WithAutoSaveLongMemory() instead")
	}
	return b.WithAutoSaveLongMemory(enabled)
}

// SaveLongMemory explicitly saves the current conversation history to persistent storage.
// Only needed when WithAutoSaveLongMemory(false) is set.
//
// Returns error if:
//   - No memory ID is set (use WithLongMemory first)
//   - No memory backend is configured
//   - Backend save operation fails
//
// Example:
//
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().
//	    WithLongMemory("user-alice").
//	    WithAutoSaveLongMemory(false)  // Manual save mode
//
//	agent.Ask(ctx, "Message 1")
//	agent.Ask(ctx, "Message 2")
//	err := agent.SaveLongMemory(ctx)  // Save both messages
func (b *Builder) SaveLongMemory(ctx context.Context) error {
	if b.longMemoryID == "" {
		return ErrLongMemoryIDRequired
	}

	if b.longMemoryBackend == nil {
		return ErrLongMemoryBackendRequired
	}

	return b.longMemoryBackend.Save(ctx, b.longMemoryID, b.messages)
}

// SaveSession is deprecated. Use SaveLongMemory() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use SaveLongMemory() for clearer intent.
func (b *Builder) SaveSession(ctx context.Context) error {
	if b.logger != nil {
		b.logger.Warn(ctx, "SaveSession() is deprecated, use SaveLongMemory() instead")
	}
	return b.SaveLongMemory(ctx)
}

// LoadLongMemory explicitly loads conversation history from persistent storage.
// Normally not needed as WithLongMemory() auto-loads on initialization.
//
// Use this to reload memory after external modifications:
//
//	agent.LoadLongMemory(ctx)  // Refresh from storage
func (b *Builder) LoadLongMemory(ctx context.Context) error {
	if b.longMemoryID == "" {
		return ErrLongMemoryIDRequired
	}

	if b.longMemoryBackend == nil {
		return ErrLongMemoryBackendRequired
	}

	messages, err := b.longMemoryBackend.Load(ctx, b.longMemoryID)
	if err != nil {
		return err
	}

	if messages != nil {
		b.messages = messages
	}

	return nil
}

// LoadSession is deprecated. Use LoadLongMemory() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use LoadLongMemory() for clearer intent.
func (b *Builder) LoadSession(ctx context.Context) error {
	if b.logger != nil {
		b.logger.Warn(ctx, "LoadSession() is deprecated, use LoadLongMemory() instead")
	}
	return b.LoadLongMemory(ctx)
}

// DeleteLongMemory removes the current memory from persistent storage.
// Does not clear in-memory messages (use Clear() for that).
//
// Example:
//
//	agent.DeleteLongMemory(ctx)  // Remove from storage
//	agent.Clear()                // Clear RAM messages
func (b *Builder) DeleteLongMemory(ctx context.Context) error {
	if b.longMemoryID == "" {
		return ErrLongMemoryIDRequired
	}

	if b.longMemoryBackend == nil {
		return ErrLongMemoryBackendRequired
	}

	return b.longMemoryBackend.Delete(ctx, b.longMemoryID)
}

// DeleteSession is deprecated. Use DeleteLongMemory() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use DeleteLongMemory() for clearer intent.
func (b *Builder) DeleteSession(ctx context.Context) error {
	if b.logger != nil {
		b.logger.Warn(ctx, "DeleteSession() is deprecated, use DeleteLongMemory() instead")
	}
	return b.DeleteLongMemory(ctx)
}

// ListLongMemories returns all available memory IDs from the backend.
// Requires a memory backend to be configured.
//
// Example:
//
//	memories, _ := agent.ListLongMemories(ctx)
//	for _, memoryID := range memories {
//	    fmt.Println(memoryID)
//	}
func (b *Builder) ListLongMemories(ctx context.Context) ([]string, error) {
	if b.longMemoryBackend == nil {
		return nil, ErrLongMemoryBackendRequired
	}

	return b.longMemoryBackend.List(ctx)
}

// ListSessions is deprecated. Use ListLongMemories() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use ListLongMemories() for clearer intent.
func (b *Builder) ListSessions(ctx context.Context) ([]string, error) {
	if b.logger != nil {
		b.logger.Warn(ctx, "ListSessions() is deprecated, use ListLongMemories() instead")
	}
	return b.ListLongMemories(ctx)
}

// GetLongMemoryID returns the current long-term memory ID, or empty string if not set.
func (b *Builder) GetLongMemoryID() string {
	return b.longMemoryID
}

// GetSessionID is deprecated. Use GetLongMemoryID() instead.
//
// Deprecated: This method will be removed in v1.0.0.
// Use GetLongMemoryID() for clearer intent.
func (b *Builder) GetSessionID() string {
	if b.logger != nil {
		ctx := context.Background()
		b.logger.Warn(ctx, "GetSessionID() is deprecated, use GetLongMemoryID() instead")
	}
	return b.GetLongMemoryID()
}
