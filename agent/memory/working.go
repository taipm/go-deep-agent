package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkingMemoryImpl implements the WorkingMemory interface
// Uses FIFO with importance-based retention
type WorkingMemoryImpl struct {
	messages []Message
	capacity int
	mu       sync.RWMutex
}

// NewWorkingMemory creates a new working memory with specified capacity
func NewWorkingMemory(capacity int) *WorkingMemoryImpl {
	if capacity <= 0 {
		capacity = 10 // Default capacity
	}

	return &WorkingMemoryImpl{
		messages: make([]Message, 0, capacity),
		capacity: capacity,
	}
}

// Add implements WorkingMemory.Add
func (w *WorkingMemoryImpl) Add(ctx context.Context, msg Message) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Set timestamp if not already set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	w.messages = append(w.messages, msg)
	return nil
}

// Recent implements WorkingMemory.Recent
func (w *WorkingMemoryImpl) Recent(ctx context.Context, n int) ([]Message, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if n <= 0 {
		return []Message{}, nil
	}

	size := len(w.messages)
	if n > size {
		n = size
	}

	// Return last N messages (most recent)
	start := size - n
	result := make([]Message, n)
	copy(result, w.messages[start:])

	return result, nil
}

// All implements WorkingMemory.All
func (w *WorkingMemoryImpl) All(ctx context.Context) ([]Message, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]Message, len(w.messages))
	copy(result, w.messages)

	return result, nil
}

// Clear implements WorkingMemory.Clear
func (w *WorkingMemoryImpl) Clear(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.messages = make([]Message, 0, w.capacity)
	return nil
}

// Compress implements WorkingMemory.Compress
// Summarizes old messages when capacity is exceeded
func (w *WorkingMemoryImpl) Compress(ctx context.Context) (Message, []Message, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.messages) <= w.capacity {
		// No compression needed
		return Message{}, []Message{}, nil
	}

	// Calculate how many messages to compress
	excess := len(w.messages) - w.capacity
	if excess <= 0 {
		return Message{}, []Message{}, nil
	}

	// Take oldest messages to compress
	toCompress := w.messages[:excess]

	// Create simple summary
	summary := w.createSummary(toCompress)

	// Keep newer messages
	w.messages = w.messages[excess:]

	return summary, toCompress, nil
}

// Size implements WorkingMemory.Size
func (w *WorkingMemoryImpl) Size() int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return len(w.messages)
}

// Capacity implements WorkingMemory.Capacity
func (w *WorkingMemoryImpl) Capacity() int {
	return w.capacity
}

// createSummary creates a simple summary message from multiple messages
func (w *WorkingMemoryImpl) createSummary(messages []Message) Message {
	if len(messages) == 0 {
		return Message{}
	}

	// Simple summarization: indicate how many messages were compressed
	summary := Message{
		Role:      "system",
		Content:   fmt.Sprintf("[Compressed %d older messages]", len(messages)),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"type":             "summary",
			"compressed_count": len(messages),
			"compressed_from":  messages[0].Timestamp,
			"compressed_to":    messages[len(messages)-1].Timestamp,
		},
	}

	return summary
}
