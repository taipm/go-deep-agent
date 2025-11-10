package memory

import (
	"context"
	"testing"
	"time"
)

// TestMemoryIntegration_WorkingToEpisodic tests the flow from Working to Episodic memory
func TestMemoryIntegration_WorkingToEpisodic(t *testing.T) {
	ctx := context.Background()

	// Create memory system with small working capacity to trigger compression
	config := MemoryConfig{
		WorkingCapacity:   3, // Small capacity to test compression
		EpisodicEnabled:   true,
		EpisodicThreshold: 0.5,
		AutoCompress:      true,
		ImportanceScoring: true,
	}

	mem := NewWithConfig(config)

	// Add messages with varying importance
	messages := []struct {
		content    string
		importance float64 // Expected importance
	}{
		{"Hello, how are you?", 0.3},                    // Low importance
		{"My name is John and I'm allergic to peanuts", 0.8}, // High importance (personal info)
		{"What's the weather like?", 0.3},               // Low importance
		{"Remember: my birthday is March 15th", 0.9},    // High importance (explicit remember)
		{"Thanks for your help!", 0.3},                  // Low importance
	}

	for _, msg := range messages {
		err := mem.Add(ctx, Message{
			Role:      "user",
			Content:   msg.content,
			Timestamp: time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}
	}

	// Working memory should have last 3 messages (capacity = 3)
	stats := mem.Stats(ctx)
	if stats.WorkingSize != 3 {
		t.Errorf("Expected working size 3, got %d", stats.WorkingSize)
	}

	// Episodic memory should have messages with importance >= 0.5
	// Expected: "allergic to peanuts" (0.8) and "birthday" (0.9)
	if stats.EpisodicSize < 2 {
		t.Errorf("Expected at least 2 messages in episodic memory, got %d", stats.EpisodicSize)
	}

	// Test recall - should retrieve from both Working and Episodic
	opts := RecallOptions{
		MaxMessages:  10,
		WorkingSize:  3,
		EpisodicTopK: 2,
		Deduplicate:  true,
	}

	recalled, err := mem.Recall(ctx, "personal information", opts)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	// Should have messages from both tiers
	if len(recalled) == 0 {
		t.Error("Expected recalled messages, got none")
	}

	t.Logf("Recalled %d messages from %d working + %d episodic",
		len(recalled), stats.WorkingSize, stats.EpisodicSize)
}

// TestMemoryIntegration_ImportanceScoring tests automatic importance calculation
func TestMemoryIntegration_ImportanceScoring(t *testing.T) {
	ctx := context.Background()

	config := DefaultMemoryConfig()
	config.WorkingCapacity = 10
	config.EpisodicEnabled = true
	config.EpisodicThreshold = 0.6
	config.ImportanceScoring = true

	mem := NewWithConfig(config)

	testCases := []struct {
		message         string
		expectedHighImp bool // Should be >= 0.6 (threshold)
		reason          string
	}{
		{
			message:         "Remember this: I prefer Python over Go",
			expectedHighImp: true,
			reason:          "Explicit 'remember' keyword",
		},
		{
			message:         "My email is john@example.com and my phone is 555-1234",
			expectedHighImp: true,
			reason:          "Personal information (email, phone)",
		},
		{
			message:         "What time is it?",
			expectedHighImp: false,
			reason:          "Generic question, low importance",
		},
		{
			message:         "How's the weather?",
			expectedHighImp: false,
			reason:          "Casual chat, low importance",
		},
	}

	for i, tc := range testCases {
		msg := Message{
			Role:      "user",
			Content:   tc.message,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Metadata:  make(map[string]interface{}),
		}

		err := mem.Add(ctx, msg)
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}

		// Check calculated importance
		if msg.Metadata != nil {
			if imp, ok := msg.Metadata["importance"].(float64); ok {
				isHigh := imp >= config.EpisodicThreshold
				if isHigh != tc.expectedHighImp {
					t.Errorf("Case %d (%s): expected high=%v, got importance=%.2f\nMessage: %s",
						i, tc.reason, tc.expectedHighImp, imp, tc.message)
				} else {
					t.Logf("✓ Case %d: importance=%.2f (expected high=%v) - %s",
						i, imp, tc.expectedHighImp, tc.reason)
				}
			}
		}
	}

	stats := mem.Stats(ctx)
	t.Logf("Final stats: Working=%d, Episodic=%d", stats.WorkingSize, stats.EpisodicSize)
}

// TestMemoryIntegration_Compression tests memory compression triggers
func TestMemoryIntegration_Compression(t *testing.T) {
	ctx := context.Background()

	config := MemoryConfig{
		WorkingCapacity:      5,
		AutoCompress:         true,
		CompressionThreshold: 5,
		EpisodicEnabled:      true,
		EpisodicThreshold:    0.5,
		ImportanceScoring:    true,
	}

	mem := NewWithConfig(config)

	// Add 10 messages (should trigger compression at 5)
	for i := 0; i < 10; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Message number " + string(rune('0'+i)),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}

		err := mem.Add(ctx, msg)
		if err != nil {
			t.Fatalf("Failed to add message %d: %v", i, err)
		}
	}

	stats := mem.Stats(ctx)

	// Working memory should not exceed capacity
	if stats.WorkingSize > config.WorkingCapacity {
		t.Errorf("Working memory exceeded capacity: %d > %d",
			stats.WorkingSize, config.WorkingCapacity)
	}

	// Compression should have occurred
	if stats.CompressionCount == 0 {
		t.Error("Expected compression to occur, but compression count is 0")
	}

	t.Logf("Compression occurred %d times, working size: %d",
		stats.CompressionCount, stats.WorkingSize)
}

// TestMemoryIntegration_VectorRetrieval tests semantic similarity search
func TestMemoryIntegration_VectorRetrieval(t *testing.T) {
	ctx := context.Background()

	config := DefaultMemoryConfig()
	config.EpisodicEnabled = true
	config.EpisodicThreshold = 0.0 // Store everything for this test

	mem := NewWithConfig(config)

	// Add messages about different topics
	topics := map[string][]string{
		"programming": {
			"I love coding in Go",
			"Python is great for data science",
			"JavaScript is essential for web development",
		},
		"food": {
			"I'm allergic to peanuts",
			"I prefer Italian cuisine",
			"My favorite dish is pasta carbonara",
		},
		"weather": {
			"It's sunny today",
			"I love rainy weather",
			"Winter is my favorite season",
		},
	}

	for topic, messages := range topics {
		for _, content := range messages {
			msg := Message{
				Role:      "user",
				Content:   content,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"topic": topic,
				},
			}

			// Force high importance to ensure episodic storage
			msg.Metadata["importance"] = 0.8

			err := mem.episodic.Store(ctx, msg, 0.8)
			if err != nil {
				t.Fatalf("Failed to store message: %v", err)
			}
		}
	}

	// Test retrieval by similarity
	testQueries := []struct {
		query          string
		expectedTopic  string
		expectedInTop3 bool
	}{
		{
			query:          "programming languages",
			expectedTopic:  "programming",
			expectedInTop3: true,
		},
		{
			query:          "dietary restrictions",
			expectedTopic:  "food",
			expectedInTop3: true,
		},
		{
			query:          "climate preferences",
			expectedTopic:  "weather",
			expectedInTop3: true,
		},
	}

	for _, tc := range testQueries {
		results, err := mem.episodic.Retrieve(ctx, tc.query, 3)
		if err != nil {
			t.Fatalf("Retrieval failed for query '%s': %v", tc.query, err)
		}

		if len(results) == 0 {
			t.Errorf("No results for query '%s'", tc.query)
			continue
		}

		// Check if expected topic appears in top results
		found := false
		for _, msg := range results {
			if topic, ok := msg.Metadata["topic"].(string); ok {
				if topic == tc.expectedTopic {
					found = true
					break
				}
			}
		}

		if tc.expectedInTop3 && !found {
			t.Errorf("Query '%s': expected topic '%s' in top 3, but not found",
				tc.query, tc.expectedTopic)
		} else if found {
			t.Logf("✓ Query '%s': found expected topic '%s'",
				tc.query, tc.expectedTopic)
		}
	}
}

// TestMemoryIntegration_ConcurrentAccess tests thread safety
func TestMemoryIntegration_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	config := DefaultMemoryConfig()
	mem := NewWithConfig(config)

	const goroutines = 10
	const messagesPerGoroutine = 20

	// Concurrent writes
	done := make(chan bool, goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			for i := 0; i < messagesPerGoroutine; i++ {
				msg := Message{
					Role:      "user",
					Content:   "Concurrent message",
					Timestamp: time.Now(),
				}
				if err := mem.Add(ctx, msg); err != nil {
					t.Errorf("Goroutine %d: failed to add message: %v", id, err)
				}
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines
	for g := 0; g < goroutines; g++ {
		<-done
	}

	// Verify data integrity
	stats := mem.Stats(ctx)
	totalExpected := goroutines * messagesPerGoroutine

	// We expect working memory to have at most WorkingCapacity messages
	// and episodic to have the rest (based on importance threshold)
	totalMessages := stats.WorkingSize + stats.EpisodicSize
	if totalMessages > totalExpected {
		t.Errorf("Data corruption: expected max %d messages, got %d",
			totalExpected, totalMessages)
	}

	t.Logf("Concurrent test passed: %d goroutines, %d total messages, final: Working=%d, Episodic=%d",
		goroutines, totalExpected, stats.WorkingSize, stats.EpisodicSize)
}

// TestMemoryIntegration_FullCycle tests complete memory lifecycle
func TestMemoryIntegration_FullCycle(t *testing.T) {
	ctx := context.Background()

	config := MemoryConfig{
		WorkingCapacity:   5,
		EpisodicEnabled:   true,
		EpisodicThreshold: 0.6,
		AutoCompress:      true,
		ImportanceScoring: true,
	}

	mem := NewWithConfig(config)

	// Phase 1: Add initial messages
	t.Log("Phase 1: Adding initial messages...")
	for i := 0; i < 10; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Initial message " + string(rune('A'+i)),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := mem.Add(ctx, msg); err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}
	}

	stats1 := mem.Stats(ctx)
	t.Logf("After Phase 1: Working=%d, Episodic=%d, Compressions=%d",
		stats1.WorkingSize, stats1.EpisodicSize, stats1.CompressionCount)

	// Phase 2: Recall with different queries
	t.Log("Phase 2: Testing recall...")
	recalled, err := mem.Recall(ctx, "messages", RecallOptions{
		MaxMessages:  10,
		WorkingSize:  5,
		EpisodicTopK: 3,
	})
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}
	t.Logf("Recalled %d messages", len(recalled))

	// Phase 3: Manual compression
	t.Log("Phase 3: Manual compression...")
	if err := mem.Compress(ctx); err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	stats2 := mem.Stats(ctx)
	if stats2.CompressionCount <= stats1.CompressionCount {
		t.Error("Expected compression count to increase")
	}

	// Phase 4: Clear and verify
	t.Log("Phase 4: Clear memory...")
	if err := mem.Clear(ctx); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	stats3 := mem.Stats(ctx)
	if stats3.WorkingSize != 0 || stats3.EpisodicSize != 0 {
		t.Errorf("Memory not cleared: Working=%d, Episodic=%d",
			stats3.WorkingSize, stats3.EpisodicSize)
	}

	t.Log("✓ Full cycle test completed successfully")
}
