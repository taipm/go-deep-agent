package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taipm/go-deep-agent/agent/memory"
)

func main() {
	fmt.Println("=== Episodic Memory Example ===\n")

	// Create episodic memory
	em := memory.NewEpisodicMemory()
	ctx := context.Background()

	// Example 1: Store messages with different importance levels
	fmt.Println("1. Storing messages with importance scores...")

	messages := []struct {
		content    string
		importance float64
		tags       []string
		timestamp  time.Time
	}{
		{
			content:    "User's name is John Smith",
			importance: 0.9, // High importance - personal info
			tags:       []string{"personal", "name"},
			timestamp:  time.Now().Add(-10 * time.Hour),
		},
		{
			content:    "User prefers dark mode in the application",
			importance: 0.8, // High importance - preference
			tags:       []string{"preference", "ui"},
			timestamp:  time.Now().Add(-8 * time.Hour),
		},
		{
			content:    "Discussed pricing: $99/month for Pro plan",
			importance: 0.85, // High importance - business info
			tags:       []string{"pricing", "important"},
			timestamp:  time.Now().Add(-5 * time.Hour),
		},
		{
			content:    "Small talk about the weather",
			importance: 0.2, // Low importance
			tags:       []string{"casual"},
			timestamp:  time.Now().Add(-3 * time.Hour),
		},
		{
			content:    "User asked about API rate limits",
			importance: 0.7, // Medium-high importance
			tags:       []string{"api", "technical"},
			timestamp:  time.Now().Add(-2 * time.Hour),
		},
		{
			content:    "Remember to follow up on the demo next Tuesday",
			importance: 0.95, // Very high importance - explicit reminder
			tags:       []string{"reminder", "important", "demo"},
			timestamp:  time.Now().Add(-1 * time.Hour),
		},
	}

	for _, msg := range messages {
		m := memory.Message{
			Role:      "user",
			Content:   msg.content,
			Timestamp: msg.timestamp,
			Metadata: map[string]interface{}{
				"tags": msg.tags,
			},
		}

		if err := em.Store(ctx, m, msg.importance); err != nil {
			log.Fatalf("Failed to store message: %v", err)
		}

		fmt.Printf("  ✓ Stored (importance: %.2f): %s\n", msg.importance, msg.content)
	}

	fmt.Printf("\nTotal messages stored: %d\n\n", em.Size())

	// Example 2: Retrieve by importance
	fmt.Println("2. Retrieve high-importance messages (>= 0.8)...")

	highImportance, err := em.RetrieveByImportance(ctx, 0.8, 10)
	if err != nil {
		log.Fatalf("Failed to retrieve by importance: %v", err)
	}

	fmt.Printf("Found %d high-importance messages:\n", len(highImportance))
	for i, msg := range highImportance {
		fmt.Printf("  %d. %s\n", i+1, msg.Content)
	}
	fmt.Println()

	// Example 3: Time-based retrieval
	fmt.Println("3. Retrieve messages from last 6 hours...")

	sixHoursAgo := time.Now().Add(-6 * time.Hour)
	now := time.Now()

	recentMessages, err := em.RetrieveByTime(ctx, sixHoursAgo, now, 10)
	if err != nil {
		log.Fatalf("Failed to retrieve by time: %v", err)
	}

	fmt.Printf("Found %d messages in the last 6 hours:\n", len(recentMessages))
	for i, msg := range recentMessages {
		fmt.Printf("  %d. %s\n", i+1, msg.Content)
	}
	fmt.Println()

	// Example 4: Search with multiple filters
	fmt.Println("4. Search with filters (importance >= 0.7, last 24 hours, with 'important' tag)...")

	yesterday := time.Now().Add(-24 * time.Hour)
	filter := memory.SearchFilter{
		MinImportance: 0.7,
		TimeRange: &memory.TimeRange{
			Start: yesterday,
			End:   now,
		},
		Tags:  []string{"important"},
		Limit: 10,
	}

	filtered, err := em.Search(ctx, filter)
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	fmt.Printf("Found %d messages matching all filters:\n", len(filtered))
	for i, msg := range filtered {
		tags := msg.Metadata["tags"]
		fmt.Printf("  %d. %s (tags: %v)\n", i+1, msg.Content, tags)
	}
	fmt.Println()

	// Example 5: Deduplication
	fmt.Println("5. Testing automatic deduplication...")

	duplicateMsg := memory.Message{
		Role:      "user",
		Content:   "User's name is John Smith", // Same as first message
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"tags": []string{"duplicate"}},
	}

	sizeBefore := em.Size()
	em.Store(ctx, duplicateMsg, 0.9)
	sizeAfter := em.Size()

	if sizeBefore == sizeAfter {
		fmt.Println("  ✓ Duplicate message was automatically skipped")
	} else {
		fmt.Println("  ✗ Duplicate was stored (shouldn't happen)")
	}
	fmt.Println()

	// Example 6: Batch storage
	fmt.Println("6. Batch storage of conversation history...")

	conversation := []memory.Message{
		{
			Role:      "user",
			Content:   "How do I integrate the API?",
			Timestamp: time.Now(),
		},
		{
			Role:      "assistant",
			Content:   "You can integrate using our SDK or REST API",
			Timestamp: time.Now(),
		},
		{
			Role:      "user",
			Content:   "Which SDK do you recommend for Go?",
			Timestamp: time.Now(),
		},
		{
			Role:      "assistant",
			Content:   "We recommend go-deep-agent for Go applications",
			Timestamp: time.Now(),
		},
	}

	importances := []float64{0.6, 0.6, 0.7, 0.7}

	if err := em.StoreBatch(ctx, conversation, importances); err != nil {
		log.Fatalf("Failed to batch store: %v", err)
	}

	fmt.Printf("  ✓ Stored %d conversation messages in batch\n", len(conversation))
	fmt.Printf("Total messages now: %d\n\n", em.Size())

	// Example 7: Recent retrieval (without specific query)
	fmt.Println("7. Retrieve most recent 5 messages...")

	recent, err := em.Retrieve(ctx, "", 5)
	if err != nil {
		log.Fatalf("Failed to retrieve recent: %v", err)
	}

	fmt.Printf("Most recent %d messages:\n", len(recent))
	for i, msg := range recent {
		fmt.Printf("  %d. [%s] %s\n", i+1, msg.Role, msg.Content)
	}
	fmt.Println()

	// Example 8: Max size enforcement
	fmt.Println("8. Testing max size enforcement...")

	config := memory.EpisodicMemoryConfig{
		MaxSize: 5, // Only keep last 5 messages
	}
	limitedEM := memory.NewEpisodicMemoryWithConfig(config)

	// Add 10 messages
	for i := 0; i < 10; i++ {
		msg := memory.Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i+1),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		limitedEM.Store(ctx, msg, 0.5)
	}

	finalSize := limitedEM.Size()
	fmt.Printf("  ✓ Added 10 messages, but only kept last %d (max size: 5)\n", finalSize)

	if finalSize == 5 {
		fmt.Println("  ✓ Max size enforcement working correctly!")
	}
	fmt.Println()

	// Example 9: Clear all memories
	fmt.Println("9. Clearing all episodic memories...")

	fmt.Printf("  Size before clear: %d\n", em.Size())

	if err := em.Clear(ctx); err != nil {
		log.Fatalf("Failed to clear: %v", err)
	}

	fmt.Printf("  Size after clear: %d\n", em.Size())
	fmt.Println("  ✓ All memories cleared successfully!")
	fmt.Println()

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("✓ Episodic memory features demonstrated:")
	fmt.Println("  - Store messages with importance scores")
	fmt.Println("  - Retrieve by importance threshold")
	fmt.Println("  - Time-based retrieval")
	fmt.Println("  - Multi-filter search")
	fmt.Println("  - Automatic deduplication")
	fmt.Println("  - Batch storage")
	fmt.Println("  - Recent message retrieval")
	fmt.Println("  - Max size enforcement")
	fmt.Println("  - Clear all memories")
}
