package memory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMemory(t *testing.T) {
	mem := New()

	if mem == nil {
		t.Fatal("New returned nil")
	}

	ctx := context.Background()
	msg := Message{
		Role:      "user",
		Content:   "Remember: my name is John",
		Timestamp: time.Now(),
	}

	// Add message
	err := mem.Add(ctx, msg)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Recall
	opts := DefaultRecallOptions()
	recalled, err := mem.Recall(ctx, "John", opts)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	if len(recalled) == 0 {
		t.Fatal("Expected recalled messages, got none")
	}

	// Stats
	stats := mem.Stats(ctx)
	if stats.TotalMessages == 0 {
		t.Error("Expected total messages > 0")
	}
}

func TestWorkingMemory(t *testing.T) {
	wm := NewWorkingMemory(5)

	if wm.Capacity() != 5 {
		t.Errorf("Expected capacity 5, got %d", wm.Capacity())
	}

	if wm.Size() != 0 {
		t.Errorf("Expected size 0, got %d", wm.Size())
	}

	ctx := context.Background()
	msg := Message{
		Role:      "user",
		Content:   "Test message",
		Timestamp: time.Now(),
	}

	// Add message
	err := wm.Add(ctx, msg)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if wm.Size() != 1 {
		t.Errorf("Expected size 1, got %d", wm.Size())
	}

	// Recent
	recent, err := wm.Recent(ctx, 1)
	if err != nil {
		t.Fatalf("Recent failed: %v", err)
	}

	if len(recent) != 1 {
		t.Errorf("Expected 1 recent message, got %d", len(recent))
	}

	// All
	all, err := wm.All(ctx)
	if err != nil {
		t.Fatalf("All failed: %v", err)
	}

	if len(all) != 1 {
		t.Errorf("Expected 1 message, got %d", len(all))
	}

	// Clear
	err = wm.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if wm.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", wm.Size())
	}
}

func TestEpisodicMemory(t *testing.T) {
	em := NewEpisodicMemory()

	ctx := context.Background()
	msg := Message{
		Role:      "user",
		Content:   "Important message",
		Timestamp: time.Now(),
	}

	// Store
	err := em.Store(ctx, msg, 0.9)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if em.Size() != 1 {
		t.Errorf("Expected size 1, got %d", em.Size())
	}

	// Retrieve
	retrieved, err := em.Retrieve(ctx, "", 10)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("Expected 1 retrieved message, got %d", len(retrieved))
	}

	// Clear
	err = em.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if em.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", em.Size())
	}
}

func TestSemanticMemory(t *testing.T) {
	sm := NewSemanticMemory()

	ctx := context.Background()
	fact := Fact{
		Content:    "User prefers dark mode",
		Category:   "preference",
		Confidence: 0.9,
		Source:     "user_input",
		Metadata:   map[string]interface{}{"theme": "dark"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// StoreFact
	err := sm.StoreFact(ctx, fact)
	if err != nil {
		t.Fatalf("StoreFact failed: %v", err)
	}

	if sm.Size() != 1 {
		t.Errorf("Expected size 1, got %d", sm.Size())
	}

	// ListFacts
	facts, err := sm.ListFacts(ctx, "", 10)
	if err != nil {
		t.Fatalf("ListFacts failed: %v", err)
	}

	if len(facts) != 1 {
		t.Errorf("Expected 1 fact, got %d", len(facts))
	}

	// UpdateFact
	factID := facts[0].ID
	updatedFact := fact
	updatedFact.Content = "User prefers light mode"
	err = sm.UpdateFact(ctx, factID, updatedFact)
	if err != nil {
		t.Fatalf("UpdateFact failed: %v", err)
	}

	// DeleteFact
	err = sm.DeleteFact(ctx, factID)
	if err != nil {
		t.Fatalf("DeleteFact failed: %v", err)
	}

	if sm.Size() != 0 {
		t.Errorf("Expected size 0 after delete, got %d", sm.Size())
	}
}

// TestWorkingMemory_CapacityEnforcement verifies FIFO eviction when capacity is exceeded
func TestWorkingMemory_CapacityEnforcement(t *testing.T) {
	// Create working memory with capacity of 3
	wm := NewWorkingMemory(3)
	ctx := context.Background()

	// Add 5 messages (exceeds capacity)
	for i := 0; i < 5; i++ {
		msg := Message{
			Role:    "user",
			Content: fmt.Sprintf("msg%d", i),
		}
		err := wm.Add(ctx, msg)
		if err != nil {
			t.Fatalf("Add failed for message %d: %v", i, err)
		}
	}

	// Should have exactly 3 messages (capacity limit)
	if wm.Size() != 3 {
		t.Errorf("Expected size 3, got %d", wm.Size())
	}

	// Should have messages 2, 3, 4 (oldest msg0 and msg1 evicted via FIFO)
	all, err := wm.All(ctx)
	if err != nil {
		t.Fatalf("All failed: %v", err)
	}

	if len(all) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(all))
	}

	// Verify FIFO order - should have msg2, msg3, msg4
	expectedContents := []string{"msg2", "msg3", "msg4"}
	for i, msg := range all {
		if msg.Content != expectedContents[i] {
			t.Errorf("Message %d: expected %s, got %s", i, expectedContents[i], msg.Content)
		}
	}
}

// TestWorkingMemory_CapacityNeverExceeded verifies capacity is never exceeded even with concurrent adds
func TestWorkingMemory_CapacityNeverExceeded(t *testing.T) {
	wm := NewWorkingMemory(5)
	ctx := context.Background()

	// Add 100 messages
	for i := 0; i < 100; i++ {
		msg := Message{
			Role:    "user",
			Content: fmt.Sprintf("message number %d", i),
		}
		err := wm.Add(ctx, msg)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Verify size never exceeds capacity
		if wm.Size() > wm.Capacity() {
			t.Fatalf("Size %d exceeds capacity %d after adding message %d", wm.Size(), wm.Capacity(), i)
		}
	}

	// Final size should equal capacity
	if wm.Size() != 5 {
		t.Errorf("Final size should be 5, got %d", wm.Size())
	}

	// Should have last 5 messages (95-99)
	all, _ := wm.All(ctx)
	if all[0].Content != "message number 95" {
		t.Errorf("Oldest message should be 'message number 95', got '%s'", all[0].Content)
	}
	if all[4].Content != "message number 99" {
		t.Errorf("Newest message should be 'message number 99', got '%s'", all[4].Content)
	}
}
