package agent

import (
	"context"
	"testing"
	"time"

	"github.com/taipm/go-deep-agent/agent/memory"
)

func TestBuilder_WithEpisodicMemory(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithEpisodicMemory(0.8)

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()
	if !config.EpisodicEnabled {
		t.Error("Expected episodic memory to be enabled")
	}

	if config.EpisodicThreshold != 0.8 {
		t.Errorf("Expected threshold 0.8, got %.2f", config.EpisodicThreshold)
	}

	if !config.ImportanceScoring {
		t.Error("Expected importance scoring to be enabled")
	}

	if !builder.memoryEnabled {
		t.Error("Expected memory to be enabled in builder")
	}
}

func TestBuilder_WithImportanceWeights(t *testing.T) {
	weights := memory.DefaultImportanceWeights()
	weights.ExplicitRemember = 2.0
	weights.PersonalInfo = 1.5

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithImportanceWeights(weights)

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()
	if config.ImportanceWeights.ExplicitRemember != 2.0 {
		t.Errorf("Expected ExplicitRemember weight 2.0, got %.2f", config.ImportanceWeights.ExplicitRemember)
	}

	if config.ImportanceWeights.PersonalInfo != 1.5 {
		t.Errorf("Expected PersonalInfo weight 1.5, got %.2f", config.ImportanceWeights.PersonalInfo)
	}

	if !config.ImportanceScoring {
		t.Error("Expected importance scoring to be enabled")
	}
}

func TestBuilder_WithWorkingMemorySize(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithWorkingMemorySize(50)

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()
	if config.WorkingCapacity != 50 {
		t.Errorf("Expected working capacity 50, got %d", config.WorkingCapacity)
	}
}

func TestBuilder_WithSemanticMemory(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSemanticMemory()

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()
	if !config.SemanticEnabled {
		t.Error("Expected semantic memory to be enabled")
	}
}

func TestBuilder_MemoryMethodChaining(t *testing.T) {
	weights := memory.DefaultImportanceWeights()
	weights.ExplicitRemember = 2.0

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithWorkingMemorySize(30).
		WithEpisodicMemory(0.7).
		WithImportanceWeights(weights).
		WithSemanticMemory()

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()

	// Verify all configs were applied
	if config.WorkingCapacity != 30 {
		t.Errorf("Expected working capacity 30, got %d", config.WorkingCapacity)
	}

	if !config.EpisodicEnabled {
		t.Error("Expected episodic memory to be enabled")
	}

	if config.EpisodicThreshold != 0.7 {
		t.Errorf("Expected threshold 0.7, got %.2f", config.EpisodicThreshold)
	}

	if config.ImportanceWeights.ExplicitRemember != 2.0 {
		t.Errorf("Expected ExplicitRemember weight 2.0, got %.2f", config.ImportanceWeights.ExplicitRemember)
	}

	if !config.SemanticEnabled {
		t.Error("Expected semantic memory to be enabled")
	}
}

func TestBuilder_MemoryInitialization(t *testing.T) {
	// Test that memory is initialized even when not explicitly created
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithEpisodicMemory(0.5)

	if builder.memory == nil {
		t.Error("Expected memory to be auto-initialized")
	}

	// Test multiple config calls on same builder
	builder.WithWorkingMemorySize(100)

	config := builder.GetMemory().GetConfig()
	if config.WorkingCapacity != 100 {
		t.Errorf("Expected capacity 100 after second config call, got %d", config.WorkingCapacity)
	}

	if config.EpisodicThreshold != 0.5 {
		t.Errorf("Expected threshold 0.5 to be preserved, got %.2f", config.EpisodicThreshold)
	}
}

// ============================================================================
// Session Persistence Tests (v0.8.0+)
// ============================================================================

func TestBuilder_WithSessionID_Basic(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("test-session")

	if builder.longMemoryID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", builder.longMemoryID)
	}

	if builder.longMemoryBackend == nil {
		t.Error("Expected memory backend to be set")
	}

	if !builder.autoSaveLongMemory {
		t.Error("Expected auto-save to be enabled by default")
	}
}

func TestBuilder_WithSessionID_DefaultBackend(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithSessionID("test-session")

	// Should auto-initialize FileBackend
	if builder.longMemoryBackend == nil {
		t.Error("Expected default FileBackend to be initialized")
	}

	if builder.longMemoryID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", builder.longMemoryID)
	}
}

func TestBuilder_WithAutoSave(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithSessionID("test-session").
		WithAutoSave(false)

	if builder.autoSaveLongMemory {
		t.Error("Expected auto-save to be disabled")
	}

	// Re-enable
	builder.WithAutoSave(true)
	if !builder.autoSaveLongMemory {
		t.Error("Expected auto-save to be enabled")
	}
}

func TestBuilder_SaveSession_RequiresSessionID(t *testing.T) {
	ctx := context.Background()
	builder := NewOpenAI("gpt-4o-mini", "test-key").WithMemory()

	// No session ID set
	err := builder.SaveSession(ctx)
	if err != ErrLongMemoryIDRequired {
		t.Errorf("Expected ErrLongMemoryIDRequired, got %v", err)
	}
}

func TestBuilder_SaveSession_RequiresBackend(t *testing.T) {
	ctx := context.Background()
	builder := NewOpenAI("gpt-4o-mini", "test-key").WithMemory()
	builder.longMemoryID = "test" // Set directly to bypass auto-init

	err := builder.SaveSession(ctx)
	if err != ErrLongMemoryBackendRequired {
		t.Errorf("Expected ErrLongMemoryBackendRequired, got %v", err)
	}
}

func TestBuilder_LoadSession(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)
	ctx := context.Background()

	// Create and save messages
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}
	backend.Save(ctx, "test-session", messages)

	// Create builder and load
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("test-session")

	// Messages should be auto-loaded
	history := builder.GetHistory()
	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}
}

func TestBuilder_DeleteSession(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)
	ctx := context.Background()

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("delete-test")

	// Add some messages
	builder.addMessage(User("Test message"))

	// Save
	builder.SaveSession(ctx)

	// Verify saved
	loaded, _ := backend.Load(ctx, "delete-test")
	if loaded == nil {
		t.Fatal("Expected session to exist before deletion")
	}

	// Delete
	err := builder.DeleteSession(ctx)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify deleted
	loaded, _ = backend.Load(ctx, "delete-test")
	if loaded != nil {
		t.Error("Expected session to be deleted")
	}
}

func TestBuilder_ListSessions(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)
	ctx := context.Background()

	// Create multiple sessions
	sessions := []string{"session-1", "session-2", "session-3"}
	for _, sessionID := range sessions {
		backend.Save(ctx, sessionID, []Message{{Role: "user", Content: "Test"}})
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend)

	// List sessions
	listed, err := builder.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(listed) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(listed))
	}
}

func TestBuilder_GetSessionID(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithSessionID("my-session")

	sessionID := builder.GetSessionID()
	if sessionID != "my-session" {
		t.Errorf("Expected 'my-session', got '%s'", sessionID)
	}

	// Builder without session ID
	builder2 := NewOpenAI("gpt-4o-mini", "test-key").WithMemory()
	if builder2.GetSessionID() != "" {
		t.Error("Expected empty session ID")
	}
}

func TestBuilder_SessionPersistence_EndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)
	ctx := context.Background()

	// First agent instance - create session
	agent1 := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("e2e-session")

	// Add messages
	agent1.addMessage(User("Message 1"))
	agent1.addMessage(Assistant("Response 1"))
	agent1.addMessage(User("Message 2"))

	// Manually save
	err := agent1.SaveSession(ctx)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Second agent instance - load session
	agent2 := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("e2e-session")

	// Should auto-load previous messages
	history := agent2.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 messages from loaded session, got %d", len(history))
	}

	if history[0].Content != "Message 1" {
		t.Errorf("Expected 'Message 1', got '%s'", history[0].Content)
	}

	// Add more messages
	agent2.addMessage(User("Message 3"))

	// Save again
	agent2.SaveSession(ctx)

	// Third agent instance - verify cumulative
	agent3 := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("e2e-session")

	history3 := agent3.GetHistory()
	if len(history3) != 4 {
		t.Errorf("Expected 4 messages total, got %d", len(history3))
	}
}

func TestBuilder_BackwardCompatibility_WithoutSessionID(t *testing.T) {
	// Old code without session ID should work exactly as before
	builder := NewOpenAI("gpt-4o-mini", "test-key").WithMemory()

	// Add messages (in-memory only)
	builder.addMessage(User("Test message"))
	builder.addMessage(Assistant("Test response"))

	// Should work fine (backward compatible)
	history := builder.GetHistory()
	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}

	// No session ID
	if builder.GetSessionID() != "" {
		t.Error("Expected no session ID for backward compatibility")
	}

	// No backend
	if builder.longMemoryBackend != nil {
		t.Error("Expected no backend for backward compatibility")
	}
}

func TestBuilder_AutoLoad_NonExistentSession(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)

	// Load non-existent session should not error
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("non-existent")

	// Should start with empty history
	history := builder.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history for new session, got %d messages", len(history))
	}
}

func TestBuilder_ManualSaveLoad_WithAutoSaveDisabled(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)
	ctx := context.Background()

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("manual-session").
		WithAutoSave(false) // Disable auto-save

	// Add messages
	builder.addMessage(User("Message 1"))
	builder.addMessage(Assistant("Response 1"))

	// Manually save
	err := builder.SaveSession(ctx)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Create new builder and load
	builder2 := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("manual-session")

	history := builder2.GetHistory()
	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}
}

func TestBuilder_SessionID_MethodChaining(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)

	// Test fluent API
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMaxHistory(20).
		WithMemoryBackend(backend).
		WithSessionID("chain-test").
		WithAutoSave(true).
		WithSystem("You are helpful")

	if builder.GetSessionID() != "chain-test" {
		t.Error("Expected session ID to be set")
	}

	if builder.maxHistory != 20 {
		t.Error("Expected maxHistory to be set")
	}

	if builder.systemPrompt != "You are helpful" {
		t.Error("Expected system prompt to be set")
	}
}

func TestBuilder_SessionPersistence_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)
	ctx := context.Background()

	// Multiple builders accessing same session
	builder1 := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("concurrent-session")

	builder2 := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("concurrent-session")

	// Builder 1 adds messages and saves
	builder1.addMessage(User("From builder 1"))
	builder1.SaveSession(ctx)

	// Builder 2 loads (should see builder 1's messages)
	builder2.LoadSession(ctx)

	history := builder2.GetHistory()
	if len(history) < 1 {
		t.Error("Expected builder 2 to load builder 1's messages")
	}

	// Builder 2 adds more and saves
	builder2.addMessage(User("From builder 2"))
	builder2.SaveSession(ctx)

	// Builder 1 reloads (should see combined messages)
	builder1.LoadSession(ctx)

	history1 := builder1.GetHistory()
	if len(history1) < 2 {
		t.Errorf("Expected at least 2 messages after reload, got %d", len(history1))
	}
}

func TestBuilder_SessionPersistence_WithTimeout(t *testing.T) {
	tempDir := t.TempDir()
	backend, _ := NewFileBackend(tempDir)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMemoryBackend(backend).
		WithSessionID("timeout-test")

	builder.addMessage(User("Test"))

	// Save with expired context - should handle gracefully
	err := builder.SaveSession(ctx)
	// May or may not error depending on timing, but shouldn't crash
	_ = err
}
