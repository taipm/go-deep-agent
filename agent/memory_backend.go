package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// MemoryBackend defines the interface for long-term memory persistence backends.
// Implementations can store conversation history in files, databases, Redis, or other storage systems.
//
// Example usage:
//
//	backend := NewFileBackend("")  // Uses default path
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().
//	    WithLongMemory("user-123").
//	        UsingBackend(backend)
type MemoryBackend interface {
	// Load retrieves conversation history for a given memory ID.
	// Returns nil if memory doesn't exist (first time).
	// Returns error only for actual failures (not for missing memories).
	Load(ctx context.Context, memoryID string) ([]Message, error)

	// Save stores conversation history for a given memory ID.
	// Should use atomic writes to prevent corruption.
	Save(ctx context.Context, memoryID string, messages []Message) error

	// Delete removes a memory's conversation history.
	// Returns nil if memory doesn't exist.
	Delete(ctx context.Context, memoryID string) error

	// List returns all available memory IDs.
	// Returns empty slice if no memories exist.
	List(ctx context.Context) ([]string, error)
}

// FileBackend implements MemoryBackend using local file storage.
// It stores long-term memories as JSON files in a configurable directory.
// Default path: ~/.go-deep-agent/memories/
//
// Features:
//   - Atomic writes (temp file + rename) to prevent corruption
//   - Thread-safe operations (mutex protection)
//   - Automatic directory creation
//   - JSON pretty-printing for easy debugging
//
// Example:
//
//	backend := NewFileBackend("")  // Uses default path
//	backend := NewFileBackend("/custom/path/memories")  // Custom path
type FileBackend struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileBackend creates a new file-based long-term memory storage backend.
//
// If basePath is empty, uses default: ~/.go-deep-agent/memories/
// Creates directory if it doesn't exist.
//
// Returns error if directory creation fails.
func NewFileBackend(basePath string) (*FileBackend, error) {
	// Use default path if not specified
	if basePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		basePath = filepath.Join(home, ".go-deep-agent", "memories")
	}

	// Create directory if not exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memories directory: %w", err)
	}

	return &FileBackend{
		basePath: basePath,
	}, nil
}

// Load retrieves conversation history from a JSON file.
//
// Returns:
//   - nil, nil if memory file doesn't exist (first time)
//   - messages, nil if successfully loaded
//   - nil, error if file read or JSON parsing fails
func (f *FileBackend) Load(ctx context.Context, memoryID string) ([]Message, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Validate memory ID
	if memoryID == "" {
		return nil, fmt.Errorf("memory ID cannot be empty")
	}

	// Construct file path
	filePath := f.getFilePath(memoryID)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Memory doesn't exist yet - this is normal for first time
		return nil, nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read memory file: %w", err)
	}

	// Parse JSON
	var messages []Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse memory JSON: %w", err)
	}

	return messages, nil
}

// Save stores conversation history to a JSON file.
//
// Uses atomic write strategy (temp file + rename) to prevent corruption.
// Pretty-prints JSON for easy debugging.
func (f *FileBackend) Save(ctx context.Context, memoryID string, messages []Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Validate memory ID
	if memoryID == "" {
		return fmt.Errorf("memory ID cannot be empty")
	}

	// Construct file path
	filePath := f.getFilePath(memoryID)

	// Marshal to pretty JSON
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal messages to JSON: %w", err)
	}

	// Write atomically: temp file + rename
	// This prevents corruption if process crashes during write
	tempPath := filePath + ".tmp"

	// Write to temp file
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp memory file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, filePath); err != nil {
		// Clean up temp file on failure
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp memory file: %w", err)
	}

	return nil
}

// Delete removes a memory's conversation history.
//
// Returns nil if memory file doesn't exist (idempotent).
func (f *FileBackend) Delete(ctx context.Context, memoryID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Validate memory ID
	if memoryID == "" {
		return fmt.Errorf("memory ID cannot be empty")
	}

	// Construct file path
	filePath := f.getFilePath(memoryID)

	// Remove file (ignore NotExist error)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete memory file: %w", err)
	}

	return nil
}

// List returns all available memory IDs.
//
// Returns empty slice if memories directory doesn't exist or is empty.
func (f *FileBackend) List(ctx context.Context) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Read directory
	entries, err := os.ReadDir(f.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist yet - return empty list
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read memories directory: %w", err)
	}

	// Filter JSON files and extract memory IDs
	var memories []string
	for _, entry := range entries {
		// Skip directories and non-JSON files
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) == ".json" {
			// Extract memory ID (remove .json extension)
			memoryID := name[:len(name)-5]
			memories = append(memories, memoryID)
		}
	}

	return memories, nil
}

// getFilePath constructs the full file path for a memory ID.
// Internal helper method.
func (f *FileBackend) getFilePath(memoryID string) string {
	return filepath.Join(f.basePath, memoryID+".json")
}

// GetBasePath returns the base directory path used for memory storage.
// Useful for debugging and testing.
func (f *FileBackend) GetBasePath() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.basePath
}
