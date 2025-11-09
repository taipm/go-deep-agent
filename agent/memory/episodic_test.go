package memory

import (
	"context"
	"testing"
	"time"
)

// TestEpisodicMemory_Store tests basic storage functionality
func TestEpisodicMemory_Store(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	msg := Message{
		Role:      "user",
		Content:   "Hello, how are you?",
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"test": true},
	}

	err := em.Store(ctx, msg, 0.8)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if em.Size() != 1 {
		t.Errorf("Expected size 1, got %d", em.Size())
	}
}

// TestEpisodicMemory_StoreBatch tests batch storage
func TestEpisodicMemory_StoreBatch(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	messages := []Message{
		{Role: "user", Content: "Message 1", Timestamp: time.Now()},
		{Role: "assistant", Content: "Message 2", Timestamp: time.Now()},
		{Role: "user", Content: "Message 3", Timestamp: time.Now()},
	}

	importances := []float64{0.7, 0.8, 0.9}

	err := em.StoreBatch(ctx, messages, importances)
	if err != nil {
		t.Fatalf("StoreBatch failed: %v", err)
	}

	if em.Size() != 3 {
		t.Errorf("Expected size 3, got %d", em.Size())
	}
}

// TestEpisodicMemory_StoreBatch_MismatchedLengths tests error handling
func TestEpisodicMemory_StoreBatch_MismatchedLengths(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	messages := []Message{
		{Role: "user", Content: "Message 1", Timestamp: time.Now()},
	}
	importances := []float64{0.7, 0.8} // Mismatched length

	err := em.StoreBatch(ctx, messages, importances)
	if err == nil {
		t.Fatal("Expected error for mismatched lengths, got nil")
	}
}

// TestEpisodicMemory_Retrieve tests retrieval functionality
func TestEpisodicMemory_Retrieve(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	// Add 10 messages
	for i := 0; i < 10; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		em.Store(ctx, msg, 0.5)
	}

	// Retrieve top 5
	results, err := em.Retrieve(ctx, "test query", 5)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

// TestEpisodicMemory_RetrieveByTime tests time-based retrieval
func TestEpisodicMemory_RetrieveByTime(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	now := time.Now()

	// Add messages at different times
	messages := []Message{
		{Role: "user", Content: "Old message", Timestamp: now.Add(-2 * time.Hour)},
		{Role: "user", Content: "Recent message 1", Timestamp: now.Add(-30 * time.Minute)},
		{Role: "user", Content: "Recent message 2", Timestamp: now.Add(-15 * time.Minute)},
		{Role: "user", Content: "Latest message", Timestamp: now},
	}

	importances := []float64{0.5, 0.6, 0.7, 0.8}
	em.StoreBatch(ctx, messages, importances)

	// Retrieve messages from last hour
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Minute)

	results, err := em.RetrieveByTime(ctx, start, end, 10)
	if err != nil {
		t.Fatalf("RetrieveByTime failed: %v", err)
	}

	// Should get 3 recent messages (not the old one)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

// TestEpisodicMemory_RetrieveByImportance tests importance-based retrieval
func TestEpisodicMemory_RetrieveByImportance(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	messages := []Message{
		{Role: "user", Content: "Low importance", Timestamp: time.Now()},
		{Role: "user", Content: "Medium importance", Timestamp: time.Now()},
		{Role: "user", Content: "High importance", Timestamp: time.Now()},
	}

	importances := []float64{0.3, 0.6, 0.9}
	em.StoreBatch(ctx, messages, importances)

	// Retrieve only high-importance messages (>= 0.7)
	results, err := em.RetrieveByImportance(ctx, 0.7, 10)
	if err != nil {
		t.Fatalf("RetrieveByImportance failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result with importance >= 0.7, got %d", len(results))
	}

	if results[0].Content != "High importance" {
		t.Errorf("Expected 'High importance', got '%s'", results[0].Content)
	}
}

// TestEpisodicMemory_Search tests search with filters
func TestEpisodicMemory_Search(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	now := time.Now()

	messages := []Message{
		{
			Role:      "user",
			Content:   "Old low importance",
			Timestamp: now.Add(-2 * time.Hour),
			Metadata:  map[string]interface{}{"tags": []string{"old"}},
		},
		{
			Role:      "user",
			Content:   "Recent high importance",
			Timestamp: now.Add(-30 * time.Minute),
			Metadata:  map[string]interface{}{"tags": []string{"recent"}},
		},
		{
			Role:      "user",
			Content:   "Latest high importance",
			Timestamp: now,
			Metadata:  map[string]interface{}{"tags": []string{"recent", "important"}},
		},
	}

	importances := []float64{0.3, 0.8, 0.9}
	em.StoreBatch(ctx, messages, importances)

	// Search for recent high-importance messages
	filter := SearchFilter{
		MinImportance: 0.7,
		TimeRange: &TimeRange{
			Start: now.Add(-1 * time.Hour),
			End:   now.Add(1 * time.Minute),
		},
		Limit: 10,
	}

	results, err := em.Search(ctx, filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should get 2 recent high-importance messages
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// TestEpisodicMemory_Search_WithTags tests tag-based filtering
func TestEpisodicMemory_Search_WithTags(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	messages := []Message{
		{
			Role:      "user",
			Content:   "Message 1",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"tags": []string{"tag1"}},
		},
		{
			Role:      "user",
			Content:   "Message 2",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"tags": []string{"tag1", "tag2"}},
		},
		{
			Role:      "user",
			Content:   "Message 3",
			Timestamp: time.Now(),
			Metadata:  map[string]interface{}{"tags": []string{"tag2"}},
		},
	}

	importances := []float64{0.5, 0.5, 0.5}
	em.StoreBatch(ctx, messages, importances)

	// Search for messages with tag1
	filter := SearchFilter{
		Tags:  []string{"tag1"},
		Limit: 10,
	}

	results, err := em.Search(ctx, filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should get 2 messages with tag1
	if len(results) != 2 {
		t.Errorf("Expected 2 results with tag1, got %d", len(results))
	}
}

// TestEpisodicMemory_Deduplication tests automatic deduplication
func TestEpisodicMemory_Deduplication(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	now := time.Now()

	// Add same message twice within 1 second
	msg := Message{
		Role:      "user",
		Content:   "Duplicate message",
		Timestamp: now,
	}

	em.Store(ctx, msg, 0.8)

	// Try to add duplicate immediately
	msg.Timestamp = now.Add(500 * time.Millisecond)
	em.Store(ctx, msg, 0.8)

	// Should only have 1 message due to deduplication
	if em.Size() != 1 {
		t.Errorf("Expected size 1 after deduplication, got %d", em.Size())
	}

	// Add same content but after 2 seconds (should be allowed)
	msg.Timestamp = now.Add(2 * time.Second)
	em.Store(ctx, msg, 0.8)

	// Should now have 2 messages
	if em.Size() != 2 {
		t.Errorf("Expected size 2 after time gap, got %d", em.Size())
	}
}

// TestEpisodicMemory_MaxSize tests size limit enforcement
func TestEpisodicMemory_MaxSize(t *testing.T) {
	config := EpisodicMemoryConfig{
		MaxSize: 5,
	}
	em := NewEpisodicMemoryWithConfig(config)
	ctx := context.Background()

	// Add 10 messages
	for i := 0; i < 10; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		em.Store(ctx, msg, 0.5)
	}

	// Should only have 5 messages (most recent)
	if em.Size() != 5 {
		t.Errorf("Expected size 5 due to max size limit, got %d", em.Size())
	}
}

// TestEpisodicMemory_Clear tests clearing all memories
func TestEpisodicMemory_Clear(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	// Add some messages
	messages := []Message{
		{Role: "user", Content: "Message 1", Timestamp: time.Now()},
		{Role: "user", Content: "Message 2", Timestamp: time.Now()},
	}
	importances := []float64{0.5, 0.5}
	em.StoreBatch(ctx, messages, importances)

	if em.Size() != 2 {
		t.Errorf("Expected size 2 before clear, got %d", em.Size())
	}

	// Clear all
	err := em.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if em.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", em.Size())
	}
}

// TestEpisodicMemory_EmptyRetrieve tests retrieval from empty memory
func TestEpisodicMemory_EmptyRetrieve(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	results, err := em.Retrieve(ctx, "test query", 5)
	if err != nil {
		t.Fatalf("Retrieve on empty memory failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results from empty memory, got %d", len(results))
	}
}

// TestEpisodicMemory_LargeLimit tests retrieval with limit larger than size
func TestEpisodicMemory_LargeLimit(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	// Add 3 messages
	for i := 0; i < 3; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: time.Now(),
		}
		em.Store(ctx, msg, 0.5)
	}

	// Request 10 but should only get 3
	results, err := em.Retrieve(ctx, "test query", 10)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results (all available), got %d", len(results))
	}
}

// TestHasTags tests the hasTags helper function
func TestHasTags(t *testing.T) {
	tests := []struct {
		name         string
		msg          Message
		requiredTags []string
		expected     bool
	}{
		{
			name: "has all tags",
			msg: Message{
				Metadata: map[string]interface{}{
					"tags": []string{"tag1", "tag2", "tag3"},
				},
			},
			requiredTags: []string{"tag1", "tag2"},
			expected:     true,
		},
		{
			name: "missing some tags",
			msg: Message{
				Metadata: map[string]interface{}{
					"tags": []string{"tag1"},
				},
			},
			requiredTags: []string{"tag1", "tag2"},
			expected:     false,
		},
		{
			name:         "no metadata",
			msg:          Message{},
			requiredTags: []string{"tag1"},
			expected:     false,
		},
		{
			name: "tags not a slice",
			msg: Message{
				Metadata: map[string]interface{}{
					"tags": "not a slice",
				},
			},
			requiredTags: []string{"tag1"},
			expected:     false,
		},
		{
			name: "empty required tags",
			msg: Message{
				Metadata: map[string]interface{}{
					"tags": []string{"tag1"},
				},
			},
			requiredTags: []string{},
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasTags(tt.msg, tt.requiredTags)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEpisodicMemory_ConcurrentAccess tests concurrent store and retrieve
func TestEpisodicMemory_ConcurrentAccess(t *testing.T) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	// Concurrent stores
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			msg := Message{
				Role:      "user",
				Content:   "Concurrent message " + string(rune('0'+n)),
				Timestamp: time.Now(),
			}
			em.Store(ctx, msg, 0.5)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 messages (allowing for potential deduplication)
	size := em.Size()
	if size < 1 || size > 10 {
		t.Errorf("Expected size between 1 and 10, got %d", size)
	}
}
