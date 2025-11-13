package agent

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// Test NewFileBackend with default path
func TestNewFileBackend_DefaultPath(t *testing.T) {
	backend, err := NewFileBackend("")
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	if backend.GetBasePath() == "" {
		t.Error("Expected non-empty base path")
	}

	// Should contain ".go-deep-agent/sessions"
	if !containsSubstring(backend.GetBasePath(), ".go-deep-agent") {
		t.Errorf("Expected path to contain '.go-deep-agent', got: %s", backend.GetBasePath())
	}
}

// Test NewFileBackend with custom path
func TestNewFileBackend_CustomPath(t *testing.T) {
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "custom-sessions")

	backend, err := NewFileBackend(customPath)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	if backend.GetBasePath() != customPath {
		t.Errorf("Expected path %s, got %s", customPath, backend.GetBasePath())
	}

	// Directory should be created
	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

// Test Save and Load basic functionality
func TestFileBackend_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()
	sessionID := "test-session-1"

	// Prepare test messages
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
		{Role: "assistant", Content: "I'm doing great!"},
	}

	// Save messages
	err = backend.Save(ctx, sessionID, messages)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load messages
	loaded, err := backend.Load(ctx, sessionID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded messages
	if len(loaded) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(loaded))
	}

	for i, msg := range messages {
		if loaded[i].Role != msg.Role {
			t.Errorf("Message %d: expected role %s, got %s", i, msg.Role, loaded[i].Role)
		}
		if loaded[i].Content != msg.Content {
			t.Errorf("Message %d: expected content %s, got %s", i, msg.Content, loaded[i].Content)
		}
	}
}

// Test Load non-existent session returns nil (not error)
func TestFileBackend_LoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()

	// Load non-existent session
	loaded, err := backend.Load(ctx, "non-existent-session")
	if err != nil {
		t.Errorf("Expected nil error for non-existent session, got: %v", err)
	}

	if loaded != nil {
		t.Errorf("Expected nil messages for non-existent session, got: %v", loaded)
	}
}

// Test Delete functionality
func TestFileBackend_Delete(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()
	sessionID := "test-session-delete"

	// Save a session
	messages := []Message{
		{Role: "user", Content: "Test message"},
	}
	err = backend.Save(ctx, sessionID, messages)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify session exists
	loaded, err := backend.Load(ctx, sessionID)
	if err != nil || loaded == nil {
		t.Fatal("Session should exist before deletion")
	}

	// Delete session
	err = backend.Delete(ctx, sessionID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify session no longer exists
	loaded, err = backend.Load(ctx, sessionID)
	if err != nil {
		t.Errorf("Load after delete should not error, got: %v", err)
	}
	if loaded != nil {
		t.Error("Session should not exist after deletion")
	}
}

// Test Delete non-existent session (idempotent)
func TestFileBackend_DeleteNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()

	// Delete non-existent session should not error
	err = backend.Delete(ctx, "non-existent-session")
	if err != nil {
		t.Errorf("Expected nil error for deleting non-existent session, got: %v", err)
	}
}

// Test List functionality
func TestFileBackend_List(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()

	// List empty directory
	sessions, err := backend.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}

	// Create multiple sessions
	sessionIDs := []string{"session-1", "session-2", "session-3"}
	messages := []Message{{Role: "user", Content: "Test"}}

	for _, sessionID := range sessionIDs {
		err = backend.Save(ctx, sessionID, messages)
		if err != nil {
			t.Fatalf("Save failed for %s: %v", sessionID, err)
		}
	}

	// List sessions
	sessions, err = backend.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(sessions) != len(sessionIDs) {
		t.Errorf("Expected %d sessions, got %d", len(sessionIDs), len(sessions))
	}

	// Verify all sessions are listed
	for _, expectedID := range sessionIDs {
		found := false
		for _, actualID := range sessions {
			if actualID == expectedID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find session %s in list", expectedID)
		}
	}
}

// Test List with non-existent directory
func TestFileBackend_ListNonExistentDir(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "does-not-exist")

	backend := &FileBackend{basePath: nonExistentPath}
	ctx := context.Background()

	// Should return empty list, not error
	sessions, err := backend.List(ctx)
	if err != nil {
		t.Errorf("Expected nil error for non-existent directory, got: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected empty list, got %d sessions", len(sessions))
	}
}

// Test concurrent Save operations (thread safety)
func TestFileBackend_ConcurrentSave(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()
	numGoroutines := 10
	numSavesPerGoroutine := 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines to save concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			sessionID := "concurrent-session"
			for j := 0; j < numSavesPerGoroutine; j++ {
				messages := []Message{
					{Role: "user", Content: "Message from goroutine"},
					{Role: "assistant", Content: "Response"},
				}
				err := backend.Save(ctx, sessionID, messages)
				if err != nil {
					t.Errorf("Goroutine %d: Save failed: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state is valid (no corruption)
	loaded, err := backend.Load(ctx, "concurrent-session")
	if err != nil {
		t.Fatalf("Load after concurrent saves failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("Expected messages after concurrent saves")
	}
}

// Test concurrent Read operations
func TestFileBackend_ConcurrentLoad(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()
	sessionID := "concurrent-load-session"

	// Save initial data
	messages := []Message{
		{Role: "user", Content: "Test message"},
		{Role: "assistant", Content: "Test response"},
	}
	err = backend.Save(ctx, sessionID, messages)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Launch multiple concurrent reads
	numGoroutines := 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			loaded, err := backend.Load(ctx, sessionID)
			if err != nil {
				t.Errorf("Goroutine %d: Load failed: %v", id, err)
				return
			}
			if len(loaded) != len(messages) {
				t.Errorf("Goroutine %d: Expected %d messages, got %d", id, len(messages), len(loaded))
			}
		}(i)
	}

	wg.Wait()
}

// Test empty session ID validation
func TestFileBackend_EmptySessionID(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()
	messages := []Message{{Role: "user", Content: "Test"}}

	// Load with empty session ID
	_, err = backend.Load(ctx, "")
	if err == nil {
		t.Error("Expected error for empty session ID in Load")
	}

	// Save with empty session ID
	err = backend.Save(ctx, "", messages)
	if err == nil {
		t.Error("Expected error for empty session ID in Save")
	}

	// Delete with empty session ID
	err = backend.Delete(ctx, "")
	if err == nil {
		t.Error("Expected error for empty session ID in Delete")
	}
}

// Test Save overwrites existing session
func TestFileBackend_SaveOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()
	sessionID := "overwrite-session"

	// First save
	messages1 := []Message{
		{Role: "user", Content: "First message"},
	}
	err = backend.Save(ctx, sessionID, messages1)
	if err != nil {
		t.Fatalf("First save failed: %v", err)
	}

	// Second save (overwrite)
	messages2 := []Message{
		{Role: "user", Content: "Second message"},
		{Role: "assistant", Content: "Second response"},
	}
	err = backend.Save(ctx, sessionID, messages2)
	if err != nil {
		t.Fatalf("Second save failed: %v", err)
	}

	// Load and verify
	loaded, err := backend.Load(ctx, sessionID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded) != len(messages2) {
		t.Errorf("Expected %d messages, got %d", len(messages2), len(loaded))
	}

	if loaded[0].Content != "Second message" {
		t.Errorf("Expected 'Second message', got '%s'", loaded[0].Content)
	}
}

// Test large conversation (stress test)
func TestFileBackend_LargeConversation(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()
	sessionID := "large-session"

	// Create 1000 messages
	messages := make([]Message, 1000)
	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			messages[i] = Message{Role: "user", Content: "User message " + string(rune(i))}
		} else {
			messages[i] = Message{Role: "assistant", Content: "Assistant message " + string(rune(i))}
		}
	}

	// Save
	err = backend.Save(ctx, sessionID, messages)
	if err != nil {
		t.Fatalf("Save large conversation failed: %v", err)
	}

	// Load
	loaded, err := backend.Load(ctx, sessionID)
	if err != nil {
		t.Fatalf("Load large conversation failed: %v", err)
	}

	if len(loaded) != 1000 {
		t.Errorf("Expected 1000 messages, got %d", len(loaded))
	}
}

// Test JSON corruption handling
func TestFileBackend_CorruptedJSON(t *testing.T) {
	tempDir := t.TempDir()
	backend, err := NewFileBackend(tempDir)
	if err != nil {
		t.Fatalf("NewFileBackend failed: %v", err)
	}

	ctx := context.Background()
	sessionID := "corrupted-session"

	// Manually create corrupted JSON file
	filePath := filepath.Join(tempDir, sessionID+".json")
	err = os.WriteFile(filePath, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Load should return error
	_, err = backend.Load(ctx, sessionID)
	if err == nil {
		t.Error("Expected error when loading corrupted JSON")
	}
}

// Helper function for string contains check
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
