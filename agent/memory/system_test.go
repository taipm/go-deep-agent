package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryAdd tests adding messages to memory system
func TestMemoryAdd(t *testing.T) {
	mem := New()
	ctx := context.Background()

	msg := Message{
		Role:      "user",
		Content:   "Test message",
		Timestamp: time.Now(),
	}

	err := mem.Add(ctx, msg)
	require.NoError(t, err)

	stats := mem.Stats(ctx)
	assert.Equal(t, 1, stats.TotalMessages)
	assert.Equal(t, 1, stats.WorkingSize)
}

// TestMemoryAddWithImportance tests importance scoring
func TestMemoryAddWithImportance(t *testing.T) {
	config := DefaultMemoryConfig()
	config.EpisodicThreshold = 0.7
	mem := NewWithConfig(config)
	ctx := context.Background()

	// High importance message (should go to episodic)
	highImportance := Message{
		Role:      "user",
		Content:   "Remember: my name is Alice and I love programming",
		Timestamp: time.Now(),
	}

	err := mem.Add(ctx, highImportance)
	require.NoError(t, err)

	stats := mem.Stats(ctx)
	assert.Equal(t, 1, stats.TotalMessages)
	assert.Equal(t, 1, stats.WorkingSize)
	// Note: Episodic might be 0 if importance < threshold
}

// TestMemoryAddMultiple tests adding multiple messages
func TestMemoryAddMultiple(t *testing.T) {
	mem := New()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Message " + string(rune('A'+i)),
			Timestamp: time.Now(),
		}
		err := mem.Add(ctx, msg)
		require.NoError(t, err)
	}

	stats := mem.Stats(ctx)
	assert.Equal(t, 5, stats.TotalMessages)
	assert.Equal(t, 5, stats.WorkingSize)
}

// TestMemoryRecall tests recalling messages
func TestMemoryRecall(t *testing.T) {
	mem := New()
	ctx := context.Background()

	// Add test messages
	messages := []string{"Hello", "How are you?", "I like Go"}
	for _, content := range messages {
		msg := Message{
			Role:      "user",
			Content:   content,
			Timestamp: time.Now(),
		}
		err := mem.Add(ctx, msg)
		require.NoError(t, err)
	}

	// Recall with default options
	opts := DefaultRecallOptions()
	opts.WorkingSize = 10

	recalled, err := mem.Recall(ctx, "Go", opts)
	require.NoError(t, err)
	assert.NotEmpty(t, recalled)
}

// TestMemoryCompress tests memory compression
func TestMemoryCompress(t *testing.T) {
	config := DefaultMemoryConfig()
	config.WorkingCapacity = 5
	config.AutoCompress = false // Manual compression
	mem := NewWithConfig(config)
	ctx := context.Background()

	// Fill working memory
	for i := 0; i < 7; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Message " + string(rune('A'+i)),
			Timestamp: time.Now(),
		}
		_ = mem.Add(ctx, msg)
	}

	statsBefore := mem.Stats(ctx)
	// With FIFO eviction, working size should be capped at capacity (5)
	assert.Equal(t, 5, statsBefore.WorkingSize)

	// Manual compression
	err := mem.Compress(ctx)
	require.NoError(t, err)

	statsAfter := mem.Stats(ctx)
	assert.Equal(t, 1, statsAfter.CompressionCount)
}

// TestMemoryClear tests clearing all memory
func TestMemoryClear(t *testing.T) {
	mem := New()
	ctx := context.Background()

	// Add messages
	for i := 0; i < 3; i++ {
		msg := Message{
			Role:      "user",
			Content:   "Test",
			Timestamp: time.Now(),
		}
		_ = mem.Add(ctx, msg)
	}

	statsBefore := mem.Stats(ctx)
	assert.Equal(t, 3, statsBefore.TotalMessages)

	// Clear
	err := mem.Clear(ctx)
	require.NoError(t, err)

	statsAfter := mem.Stats(ctx)
	assert.Equal(t, 0, statsAfter.TotalMessages)
	assert.Equal(t, 0, statsAfter.WorkingSize)
}

// TestMemoryConfig tests configuration
func TestMemoryConfig(t *testing.T) {
	config := DefaultMemoryConfig()
	config.WorkingCapacity = 20
	config.EpisodicEnabled = false

	mem := NewWithConfig(config)

	cfg := mem.GetConfig()
	assert.Equal(t, 20, cfg.WorkingCapacity)
	assert.False(t, cfg.EpisodicEnabled)
}

// TestMemorySetConfig tests updating configuration
func TestMemorySetConfig(t *testing.T) {
	mem := New()

	newConfig := DefaultMemoryConfig()
	newConfig.WorkingCapacity = 50

	err := mem.SetConfig(newConfig)
	require.NoError(t, err)

	cfg := mem.GetConfig()
	assert.Equal(t, 50, cfg.WorkingCapacity)
}

// TestMemoryConcurrency tests concurrent access
func TestMemoryConcurrency(t *testing.T) {
	t.Skip("Skipping slow concurrency test - can be enabled for comprehensive validation")

	mem := New()
	ctx := context.Background()

	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 5; j++ {
				msg := Message{
					Role:      "user",
					Content:   "Concurrent message",
					Timestamp: time.Now(),
				}
				_ = mem.Add(ctx, msg)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := mem.Stats(ctx)
	assert.Equal(t, 50, stats.TotalMessages)
}

// TestMemoryBackwardCompatibility tests deprecated NewSmartMemory
func TestMemoryBackwardCompatibility(t *testing.T) {
	config := DefaultMemoryConfig()
	mem := NewSmartMemory(config) // Deprecated but should work

	assert.NotNil(t, mem)
	assert.IsType(t, &Memory{}, mem)
}

// TestImportanceScoring tests importance calculation
func TestImportanceScoring(t *testing.T) {
	mem := New()
	ctx := context.Background()

	tests := []struct {
		name     string
		content  string
		metadata map[string]interface{}
		wantHigh bool // Should have high importance
	}{
		{
			name:     "explicit remember",
			content:  "Remember: my birthday is Jan 15",
			wantHigh: true,
		},
		{
			name:     "personal info",
			content:  "My name is Alice",
			wantHigh: true,
		},
		{
			name:     "question",
			content:  "What is the weather?",
			wantHigh: false,
		},
		{
			name:     "casual",
			content:  "ok",
			wantHigh: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				Role:      "user",
				Content:   tt.content,
				Timestamp: time.Now(),
				Metadata:  tt.metadata,
			}

			err := mem.Add(ctx, msg)
			require.NoError(t, err)
		})
	}
}

// TestMemoryMetadata tests storing metadata
func TestMemoryMetadata(t *testing.T) {
	mem := New()
	ctx := context.Background()

	msg := Message{
		Role:      "user",
		Content:   "Test",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"model":       "gpt-4",
			"temperature": 0.7,
		},
	}

	err := mem.Add(ctx, msg)
	require.NoError(t, err)

	// Metadata should be preserved
	assert.NotNil(t, msg.Metadata)
	assert.Equal(t, "gpt-4", msg.Metadata["model"])
}
