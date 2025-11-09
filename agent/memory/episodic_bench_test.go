package memory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkEpisodicMemory_Store benchmarks storing single messages
func BenchmarkEpisodicMemory_Store(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		em.Store(ctx, msg, 0.8)
	}
}

// BenchmarkEpisodicMemory_StoreBatch benchmarks batch storage
func BenchmarkEpisodicMemory_StoreBatch(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	batchSize := 100
	messages := make([]Message, batchSize)
	importances := make([]float64, batchSize)

	for i := 0; i < batchSize; i++ {
		messages[i] = Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		importances[i] = 0.8
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.StoreBatch(ctx, messages, importances)
	}
}

// BenchmarkEpisodicMemory_Retrieve benchmarks message retrieval
func BenchmarkEpisodicMemory_Retrieve(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	// Pre-populate with 1000 messages
	for i := 0; i < 1000; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		em.Store(ctx, msg, 0.8)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.Retrieve(ctx, "test query", 10)
	}
}

// BenchmarkEpisodicMemory_RetrieveByTime benchmarks time-based retrieval
func BenchmarkEpisodicMemory_RetrieveByTime(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	now := time.Now()

	// Pre-populate with 1000 messages over 24 hours
	for i := 0; i < 1000; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: now.Add(-24 * time.Hour).Add(time.Duration(i) * time.Minute),
		}
		em.Store(ctx, msg, 0.8)
	}

	start := now.Add(-1 * time.Hour)
	end := now

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.RetrieveByTime(ctx, start, end, 100)
	}
}

// BenchmarkEpisodicMemory_RetrieveByImportance benchmarks importance-based retrieval
func BenchmarkEpisodicMemory_RetrieveByImportance(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	// Pre-populate with 1000 messages with varying importance
	for i := 0; i < 1000; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		importance := float64(i%10) / 10.0 // 0.0 to 0.9
		em.Store(ctx, msg, importance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.RetrieveByImportance(ctx, 0.7, 100)
	}
}

// BenchmarkEpisodicMemory_Search benchmarks search with filters
func BenchmarkEpisodicMemory_Search(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	now := time.Now()

	// Pre-populate with 1000 messages
	for i := 0; i < 1000; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: now.Add(-24 * time.Hour).Add(time.Duration(i) * time.Minute),
			Metadata:  map[string]interface{}{"tags": []string{"test"}},
		}
		importance := float64(i%10) / 10.0
		em.Store(ctx, msg, importance)
	}

	filter := SearchFilter{
		MinImportance: 0.5,
		TimeRange: &TimeRange{
			Start: now.Add(-12 * time.Hour),
			End:   now,
		},
		Limit: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.Search(ctx, filter)
	}
}

// BenchmarkEpisodicMemory_Deduplication benchmarks deduplication check
func BenchmarkEpisodicMemory_Deduplication(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	// Pre-populate with 1000 unique messages
	for i := 0; i < 1000; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Unique message %d", i),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		em.Store(ctx, msg, 0.8)
	}

	// Try to add duplicate of last message
	msg := Message{
		Role:      "user",
		Content:   "Unique message 999",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.isDuplicate(msg)
	}
}

// BenchmarkEpisodicMemory_LargeScale benchmarks with different scales
func BenchmarkEpisodicMemory_LargeScale(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d_messages", size), func(b *testing.B) {
			em := NewEpisodicMemory()
			ctx := context.Background()

			// Pre-populate
			for i := 0; i < size; i++ {
				msg := Message{
					Role:      "user",
					Content:   fmt.Sprintf("Message %d", i),
					Timestamp: time.Now(),
				}
				em.Store(ctx, msg, 0.8)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Mix of operations
				em.Retrieve(ctx, "test", 10)
				em.RetrieveByImportance(ctx, 0.5, 10)

				msg := Message{
					Role:      "user",
					Content:   fmt.Sprintf("New message %d", i),
					Timestamp: time.Now(),
				}
				em.Store(ctx, msg, 0.7)
			}
		})
	}
}

// BenchmarkEpisodicMemory_Clear benchmarks clearing memory
func BenchmarkEpisodicMemory_Clear(b *testing.B) {
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		em := NewEpisodicMemory()

		// Pre-populate with 1000 messages
		for j := 0; j < 1000; j++ {
			msg := Message{
				Role:      "user",
				Content:   fmt.Sprintf("Message %d", j),
				Timestamp: time.Now(),
			}
			em.Store(ctx, msg, 0.8)
		}

		b.StartTimer()
		em.Clear(ctx)
	}
}

// BenchmarkEpisodicMemory_ConcurrentStore benchmarks concurrent storage
func BenchmarkEpisodicMemory_ConcurrentStore(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			msg := Message{
				Role:      "user",
				Content:   fmt.Sprintf("Concurrent message %d", i),
				Timestamp: time.Now(),
			}
			em.Store(ctx, msg, 0.8)
			i++
		}
	})
}

// BenchmarkEpisodicMemory_ConcurrentRetrieve benchmarks concurrent retrieval
func BenchmarkEpisodicMemory_ConcurrentRetrieve(b *testing.B) {
	em := NewEpisodicMemory()
	ctx := context.Background()

	// Pre-populate with 1000 messages
	for i := 0; i < 1000; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		em.Store(ctx, msg, 0.8)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			em.Retrieve(ctx, "test query", 10)
		}
	})
}

// BenchmarkEpisodicMemory_MaxSizeEnforcement benchmarks max size limit
func BenchmarkEpisodicMemory_MaxSizeEnforcement(b *testing.B) {
	config := EpisodicMemoryConfig{
		MaxSize: 1000,
	}
	em := NewEpisodicMemoryWithConfig(config)
	ctx := context.Background()

	// Pre-fill to max
	for i := 0; i < 1000; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
		em.Store(ctx, msg, 0.8)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := Message{
			Role:      "user",
			Content:   fmt.Sprintf("New message %d", i),
			Timestamp: time.Now(),
		}
		em.Store(ctx, msg, 0.8)
	}
}
