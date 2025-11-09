package memory

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStatsEnhanced(t *testing.T) {
	config := MemoryConfig{
		WorkingCapacity:   10,
		EpisodicEnabled:   true,
		EpisodicThreshold: 0.7,
		SemanticEnabled:   true,
		ImportanceScoring: true,
		ImportanceWeights: DefaultImportanceWeights(),
	}

	mem := NewWithConfig(config)
	ctx := context.Background()

	// Initially empty
	stats := mem.Stats(ctx)
	if stats.TotalMessages != 0 {
		t.Errorf("Expected 0 total messages, got %d", stats.TotalMessages)
	}
	if stats.WorkingSize != 0 {
		t.Errorf("Expected 0 working size, got %d", stats.WorkingSize)
	}
	if stats.EpisodicSize != 0 {
		t.Errorf("Expected 0 episodic size, got %d", stats.EpisodicSize)
	}

	// Add messages with varying importance
	now := time.Now()
	messages := []struct {
		content   string
		timestamp time.Time
	}{
		{"Just saying hello", now.Add(-10 * time.Second)},                                // Low importance
		{"Remember this is very important information", now.Add(-5 * time.Second)},       // High - has "remember"
		{"I want you to never forget my birthday is May 5th", now.Add(-2 * time.Second)}, // High - has "never forget"
		{"Weather is nice today", now},                                                   // Low importance
	}

	for i, msg := range messages {
		m := Message{
			Role:      "user",
			Content:   msg.content,
			Timestamp: msg.timestamp,
		}
		if err := mem.Add(ctx, m); err != nil {
			t.Fatalf("Failed to add message %d: %v", i, err)
		}
		// Small delay to prevent deduplication issues
		time.Sleep(10 * time.Millisecond)
	}

	// Check stats after adding messages
	stats = mem.Stats(ctx)

	if stats.TotalMessages != 4 {
		t.Errorf("Expected 4 total messages, got %d", stats.TotalMessages)
	}

	if stats.WorkingSize != 4 {
		t.Errorf("Expected 4 working messages, got %d", stats.WorkingSize)
	}

	// High importance messages (>= 0.7) should be in episodic
	if stats.EpisodicSize != 2 {
		t.Errorf("Expected 2 episodic messages (importance >= 0.7), got %d", stats.EpisodicSize)
	}

	// Check episodic timestamps
	if stats.EpisodicOldest.IsZero() {
		t.Error("Expected non-zero episodic oldest timestamp")
	}

	if stats.EpisodicNewest.IsZero() {
		t.Error("Expected non-zero episodic newest timestamp")
	}

	// Oldest should be before newest
	if !stats.EpisodicOldest.Before(stats.EpisodicNewest) {
		t.Error("Oldest timestamp should be before newest timestamp")
	}

	// Check average importance (should be >= 0.7 for messages that passed threshold)
	// Note: With raw scores (not normalized), average can exceed 1.0 when messages
	// match multiple importance features (e.g., "remember" + "personal info")
	if stats.AverageImportance < 0.7 {
		t.Errorf("Expected average importance >= 0.7, got %.2f", stats.AverageImportance)
	}
}

func TestMemoryStatsWithSemantic(t *testing.T) {
	config := MemoryConfig{
		WorkingCapacity:   10,
		EpisodicEnabled:   false,
		SemanticEnabled:   true,
		ImportanceScoring: false,
	}

	mem := NewWithConfig(config)
	ctx := context.Background()

	// Add some facts
	semanticImpl, ok := mem.semantic.(*SemanticMemoryImpl)
	if !ok {
		t.Fatal("Expected SemanticMemoryImpl")
	}

	facts := []Fact{
		{
			Content:    "User prefers dark mode",
			Category:   "preference",
			Confidence: 0.9,
		},
		{
			Content:    "API rate limit is 100/min",
			Category:   "technical",
			Confidence: 1.0,
		},
		{
			Content:    "User's name is John",
			Category:   "personal",
			Confidence: 0.95,
		},
	}

	for _, fact := range facts {
		if err := semanticImpl.StoreFact(ctx, fact); err != nil {
			t.Fatalf("Failed to store fact: %v", err)
		}
	}

	// Check stats
	stats := mem.Stats(ctx)

	if stats.SemanticSize != 3 {
		t.Errorf("Expected 3 semantic facts, got %d", stats.SemanticSize)
	}

	if stats.TotalFacts != 3 {
		t.Errorf("Expected 3 total facts, got %d", stats.TotalFacts)
	}

	// Check categories
	if len(stats.SemanticCategories) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(stats.SemanticCategories))
	}

	// Verify categories contain expected values
	categoryMap := make(map[string]bool)
	for _, cat := range stats.SemanticCategories {
		categoryMap[cat] = true
	}

	expectedCategories := []string{"preference", "technical", "personal"}
	for _, expected := range expectedCategories {
		if !categoryMap[expected] {
			t.Errorf("Expected category '%s' not found in stats", expected)
		}
	}
}

func TestMemoryStatsAfterCompression(t *testing.T) {
	config := MemoryConfig{
		WorkingCapacity:      5,
		AutoCompress:         true,
		CompressionThreshold: 5,
		EpisodicEnabled:      true,
		EpisodicThreshold:    0.5,
		ImportanceScoring:    true,
		ImportanceWeights:    DefaultImportanceWeights(),
	}

	mem := NewWithConfig(config)
	ctx := context.Background()

	// Add more messages than capacity to trigger compression
	for i := 0; i < 10; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: time.Now(),
		}
		mem.Add(ctx, msg)
	}

	// Check stats
	stats := mem.Stats(ctx)

	if stats.CompressionCount == 0 {
		t.Error("Expected at least one compression")
	}

	if stats.LastCompression.IsZero() {
		t.Error("Expected non-zero last compression time")
	}

	// Working memory should be reasonably sized
	// Note: Compression creates a summary and adds it back, so size might be
	// slightly above capacity (capacity + 1) if the last message triggered compression
	if stats.WorkingSize > stats.WorkingCapacity+1 {
		t.Errorf("Working size (%d) significantly exceeds capacity (%d)", stats.WorkingSize, stats.WorkingCapacity)
	}
}

func TestMemoryStatsAfterClear(t *testing.T) {
	config := DefaultMemoryConfig()
	mem := NewWithConfig(config)
	ctx := context.Background()

	// Add some messages
	for i := 0; i < 5; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: time.Now(),
		}
		mem.Add(ctx, msg)
	}

	// Verify messages were added
	stats := mem.Stats(ctx)
	if stats.TotalMessages == 0 {
		t.Error("Expected messages to be added")
	}

	// Clear memory
	if err := mem.Clear(ctx); err != nil {
		t.Fatalf("Failed to clear memory: %v", err)
	}

	// Check stats after clear
	stats = mem.Stats(ctx)

	if stats.TotalMessages != 0 {
		t.Errorf("Expected 0 total messages after clear, got %d", stats.TotalMessages)
	}

	if stats.WorkingSize != 0 {
		t.Errorf("Expected 0 working size after clear, got %d", stats.WorkingSize)
	}

	if stats.EpisodicSize != 0 {
		t.Errorf("Expected 0 episodic size after clear, got %d", stats.EpisodicSize)
	}

	if stats.SemanticSize != 0 {
		t.Errorf("Expected 0 semantic size after clear, got %d", stats.SemanticSize)
	}
}
