package memory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkMemoryAdd tests performance of adding messages
func BenchmarkMemoryAdd(b *testing.B) {
	mem := New()
	ctx := context.Background()

	msg := Message{
		Role:      "user",
		Content:   "This is a benchmark message",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mem.Add(ctx, msg)
	}
}

// BenchmarkMemoryAddWithMetadata tests performance with rich metadata
func BenchmarkMemoryAddWithMetadata(b *testing.B) {
	mem := New()
	ctx := context.Background()

	msg := Message{
		Role:      "user",
		Content:   "Benchmark with metadata",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"tokens":    150,
			"streaming": true,
			"tools":     []string{"search", "calculator"},
			"rag":       true,
			"images":    2,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mem.Add(ctx, msg)
	}
}

// BenchmarkMemoryRecall tests retrieval performance
func BenchmarkMemoryRecall(b *testing.B) {
	mem := New()
	ctx := context.Background()

	// Pre-populate with 100 messages
	for i := 0; i < 100; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		_ = mem.Add(ctx, msg)
	}

	opts := RecallOptions{
		MaxMessages: 10,
		WorkingSize: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mem.Recall(ctx, "Message", opts)
	}
}

// BenchmarkMemoryCompress tests compression performance
func BenchmarkMemoryCompress(b *testing.B) {
	mem := NewWithConfig(MemoryConfig{
		WorkingCapacity:   50,
		EpisodicThreshold: 0.5,
	})
	ctx := context.Background()

	// Pre-populate with 60 messages (triggers compression)
	for i := 0; i < 60; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		_ = mem.Add(ctx, msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mem.Compress(ctx)
	}
}

// BenchmarkWorkingMemoryFIFO tests FIFO operations
func BenchmarkWorkingMemoryFIFO(b *testing.B) {
	wm := NewWorkingMemory(10)
	ctx := context.Background()

	msg := Message{
		Role:      "user",
		Content:   "FIFO benchmark",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wm.Add(ctx, msg)
	}
}

// BenchmarkEpisodicMemoryAdd tests important message storage
func BenchmarkEpisodicMemoryAdd(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	msg := Message{
		Role:      "user",
		Content:   "Important memory",
		Timestamp: time.Now(),
	}
	importance := 0.8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = em.Store(ctx, msg, importance)
	}
}

// BenchmarkMemoryStats tests stats collection performance
func BenchmarkMemoryStats(b *testing.B) {
	mem := New()
	ctx := context.Background()

	// Pre-populate with messages
	for i := 0; i < 50; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Stats message %d", i),
			Timestamp: time.Now(),
		}
		_ = mem.Add(ctx, msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mem.Stats(ctx)
	}
}

// BenchmarkMemoryLargeScale tests performance with large message volumes
func BenchmarkMemoryLargeScale(b *testing.B) {
	benchmarks := []struct {
		name     string
		messages int
	}{
		{"100_messages", 100},
		{"1000_messages", 1000},
		{"10000_messages", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				mem := NewWithConfig(MemoryConfig{
					WorkingCapacity:   100,
					EpisodicThreshold: 0.6,
				})
				ctx := context.Background()
				b.StartTimer()

				for j := 0; j < bm.messages; j++ {
					msg := Message{
						Role:      "user",
						Content:   fmt.Sprintf("Large scale message %d", j),
						Timestamp: time.Now(),
					}
					_ = mem.Add(ctx, msg)
				}
			}
		})
	}
}

// BenchmarkImportanceCalculation tests scoring performance
func BenchmarkImportanceCalculation(b *testing.B) {
	mem := New()
	ctx := context.Background()

	messages := []Message{
		{Role: "user", Content: "Remember my birthday is March 15th", Timestamp: time.Now()},
		{Role: "user", Content: "What's the weather?", Timestamp: time.Now()},
		{Role: "user", Content: "Can you help me with this problem?", Timestamp: time.Now()},
		{Role: "user", Content: "Just saying hello", Timestamp: time.Now()},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, msg := range messages {
			_ = mem.Add(ctx, msg)
		}
	}
}
