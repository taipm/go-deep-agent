package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	agent "github.com/taipm/go-deep-agent/agent"
)

const (
	modelName = "gpt-4o-mini"
	redisAddr = "localhost:6379"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	fmt.Println("=== Redis Cache Examples ===\n")

	// Example 1: Simple Redis Cache
	example1SimpleRedisCache(ctx, apiKey)

	// Example 2: Redis Cache with Custom Options
	example2RedisCacheWithOptions(ctx, apiKey)

	// Example 3: Cache Statistics Tracking
	example3CacheStatistics(ctx, apiKey)

	// Example 4: Batch Operations
	example4BatchOperations(ctx, apiKey)

	// Example 5: Pattern-based Deletion
	example5PatternDeletion(ctx, apiKey)

	// Example 6: Distributed Locking
	example6DistributedLocking(ctx, apiKey)

	// Example 7: Cache Performance Comparison
	example7PerformanceComparison(ctx, apiKey)

	// Example 8: Cache with TTL Management
	example8TTLManagement(ctx, apiKey)
}

// Example 1: Simple Redis Cache Setup
func example1SimpleRedisCache(ctx context.Context, apiKey string) {
	fmt.Println("--- Example 1: Simple Redis Cache ---")

	// Create AI agent with Redis cache
	// Default: localhost:6379, no password, DB 0
	ai := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0)

	// First call - cache miss
	start := time.Now()
	response1, err := ai.Ask(ctx, "What is 2+2?")
	duration1 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("First call (cache miss): %s\n", response1)
	fmt.Printf("Duration: %v\n", duration1)

	// Second call - cache hit
	start = time.Now()
	response2, err := ai.Ask(ctx, "What is 2+2?")
	duration2 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Second call (cache hit): %s\n", response2)
	fmt.Printf("Duration: %v\n", duration2)
	fmt.Printf("Speed improvement: %.2fx faster\n", float64(duration1)/float64(duration2))

	// Show stats
	stats := ai.GetCacheStats()
	fmt.Printf("Cache Stats - Hits: %d, Misses: %d, Hit Rate: %.2f%%\n\n",
		stats.Hits, stats.Misses,
		float64(stats.Hits)/(float64(stats.Hits+stats.Misses))*100)
}

// Example 2: Redis Cache with Custom Options
func example2RedisCacheWithOptions(ctx context.Context, apiKey string) {
	fmt.Println("--- Example 2: Redis Cache with Custom Options ---")

	// Advanced configuration
	opts := &agent.RedisCacheOptions{
		Addrs:        []string{"localhost:6379"},
		Password:     "", // Set if Redis requires authentication
		DB:           0,
		PoolSize:     20,               // Connection pool size
		MinIdleConns: 10,               // Minimum idle connections
		KeyPrefix:    "myapp",          // Custom key prefix
		DefaultTTL:   10 * time.Minute, // Default TTL for cache entries
		DialTimeout:  5 * time.Second,
	}

	ai := agent.NewOpenAI(modelName, apiKey).
		WithRedisCacheOptions(opts)

	response, err := ai.Ask(ctx, "What is the capital of France?")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", response)
	fmt.Println("Cache configured with custom prefix 'myapp' and 10-minute TTL\n")
}

// Example 3: Cache Statistics Tracking
func example3CacheStatistics(ctx context.Context, apiKey string) {
	fmt.Println("--- Example 3: Cache Statistics Tracking ---")

	ai := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0)

	// Initial stats
	stats := ai.GetCacheStats()
	fmt.Printf("Initial - Hits: %d, Misses: %d, Writes: %d, Size: %d\n",
		stats.Hits, stats.Misses, stats.TotalWrites, stats.Size)

	// Make several requests
	questions := []string{
		"What is Go?",
		"What is Python?",
		"What is Rust?",
	}

	for _, q := range questions {
		ai.Ask(ctx, q)
	}

	// Repeat some questions (cache hits)
	ai.Ask(ctx, "What is Go?")
	ai.Ask(ctx, "What is Python?")

	// Final stats
	stats = ai.GetCacheStats()
	fmt.Printf("Final - Hits: %d, Misses: %d, Writes: %d, Size: %d\n",
		stats.Hits, stats.Misses, stats.TotalWrites, stats.Size)

	hitRate := float64(stats.Hits) / (float64(stats.Hits + stats.Misses)) * 100
	fmt.Printf("Hit Rate: %.2f%%\n\n", hitRate)
}

// Example 4: Batch Operations
func example4BatchOperations(ctx context.Context, apiKey string) {
	fmt.Println("--- Example 4: Batch Operations ---")

	// Note: Batch operations are internal to RedisCache
	// Here we demonstrate the effect of caching on multiple queries

	ai := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0)

	// Process multiple questions
	questions := []string{
		"What is 1+1?",
		"What is 2+2?",
		"What is 3+3?",
		"What is 4+4?",
		"What is 5+5?",
	}

	fmt.Println("Processing 5 questions (cache misses)...")
	start := time.Now()
	for i, q := range questions {
		resp, _ := ai.Ask(ctx, q)
		fmt.Printf("%d. %s -> %s\n", i+1, q, resp)
	}
	duration1 := time.Since(start)

	fmt.Println("\nReprocessing same 5 questions (cache hits)...")
	start = time.Now()
	for i, q := range questions {
		resp, _ := ai.Ask(ctx, q)
		fmt.Printf("%d. %s -> %s\n", i+1, q, resp)
	}
	duration2 := time.Since(start)

	fmt.Printf("\nFirst run (no cache): %v\n", duration1)
	fmt.Printf("Second run (cached): %v\n", duration2)
	fmt.Printf("Speed improvement: %.2fx faster\n\n", float64(duration1)/float64(duration2))
}

// Example 5: Pattern-based Deletion
func example5PatternDeletion(ctx context.Context, apiKey string) {
	fmt.Println("--- Example 5: Pattern-based Deletion ---")

	// Note: Pattern deletion is handled internally by RedisCache
	// This example shows cache clearing

	ai := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0)

	// Add some cached responses
	ai.Ask(ctx, "What is machine learning?")
	ai.Ask(ctx, "What is deep learning?")
	ai.Ask(ctx, "What is neural networks?")

	stats := ai.GetCacheStats()
	fmt.Printf("Before clear - Size: %d, Writes: %d\n", stats.Size, stats.TotalWrites)

	// Clear all cache
	err := ai.ClearCache(ctx)
	if err != nil {
		log.Printf("Error clearing cache: %v", err)
		return
	}

	stats = ai.GetCacheStats()
	fmt.Printf("After clear - Size: %d, Writes: %d\n", stats.Size, stats.TotalWrites)
	fmt.Println("All cache entries cleared\n")
}

// Example 6: Distributed Locking
func example6DistributedLocking(ctx context.Context, apiKey string) {
	fmt.Println("--- Example 6: Distributed Locking (SetNX) ---")

	// Note: SetNX is used internally by RedisCache for cache stampede prevention
	// This example demonstrates the concept

	fmt.Println("Redis SetNX (SET if Not eXists) is used internally for:")
	fmt.Println("1. Preventing cache stampede (multiple processes computing same value)")
	fmt.Println("2. Distributed locking across multiple instances")
	fmt.Println("3. Ensuring only one process fills cache for a given key")
	fmt.Println()

	ai := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0)

	// Simulate multiple concurrent requests for same question
	fmt.Println("Simulating concurrent requests...")
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			start := time.Now()
			resp, _ := ai.Ask(ctx, "What is the meaning of life?")
			duration := time.Since(start)
			fmt.Printf("Goroutine %d: %s (took %v)\n", id, resp, duration)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	fmt.Println("Note: First request computes, others may wait or get cached result\n")
}

// Example 7: Cache Performance Comparison
func example7PerformanceComparison(ctx context.Context, apiKey string) {
	fmt.Println("--- Example 7: Cache Performance Comparison ---")

	question := "Explain quantum computing in one sentence"

	// Without cache
	aiNoCache := agent.NewOpenAI(modelName, apiKey)
	start := time.Now()
	resp1, _ := aiNoCache.Ask(ctx, question)
	noCache1 := time.Since(start)

	start = time.Now()
	_, _ = aiNoCache.Ask(ctx, question)
	noCache2 := time.Since(start)

	// With memory cache
	aiMemCache := agent.NewOpenAI(modelName, apiKey).
		WithMemoryCache(100, 5*time.Minute)
	start = time.Now()
	aiMemCache.Ask(ctx, question)
	memCache1 := time.Since(start)

	start = time.Now()
	aiMemCache.Ask(ctx, question)
	memCache2 := time.Since(start)

	// With Redis cache
	aiRedisCache := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0)
	start = time.Now()
	aiRedisCache.Ask(ctx, question)
	redisCache1 := time.Since(start)

	start = time.Now()
	aiRedisCache.Ask(ctx, question)
	redisCache2 := time.Since(start)

	fmt.Printf("Question: %s\n", question)
	fmt.Printf("Answer: %s\n\n", resp1)

	fmt.Println("Performance Comparison:")
	fmt.Printf("No Cache:      1st: %v, 2nd: %v\n", noCache1, noCache2)
	fmt.Printf("Memory Cache:  1st: %v, 2nd: %v (%.2fx faster)\n",
		memCache1, memCache2, float64(memCache1)/float64(memCache2))
	fmt.Printf("Redis Cache:   1st: %v, 2nd: %v (%.2fx faster)\n",
		redisCache1, redisCache2, float64(redisCache1)/float64(redisCache2))

	fmt.Println("\nComparison:")
	fmt.Printf("Memory vs Redis (2nd call): %.2fx difference\n",
		float64(redisCache2)/float64(memCache2))
	fmt.Println("Note: Memory cache is fastest but not shared across instances")
	fmt.Println("Redis cache is slightly slower but shared and persistent\n")
}

// Example 8: Cache with TTL Management
func example8TTLManagement(ctx context.Context, apiKey string) {
	fmt.Println("--- Example 8: Cache with TTL Management ---")

	// Default TTL
	ai := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0)

	ai.Ask(ctx, "What is Docker?")
	fmt.Println("Cached 'What is Docker?' with default TTL (5 minutes)")

	// Custom TTL for specific request
	aiCustom := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0).
		WithCacheTTL(1 * time.Hour) // This response cached for 1 hour

	aiCustom.Ask(ctx, "What is Kubernetes?")
	fmt.Println("Cached 'What is Kubernetes?' with custom TTL (1 hour)")

	// Disable cache temporarily
	aiNoCache := agent.NewOpenAI(modelName, apiKey).
		WithRedisCache(redisAddr, "", 0).
		DisableCache() // Cache disabled

	aiNoCache.Ask(ctx, "What is Redis?")
	fmt.Println("'What is Redis?' NOT cached (cache disabled)")

	// Re-enable cache
	aiNoCache.EnableCache()
	aiNoCache.Ask(ctx, "What is MongoDB?")
	fmt.Println("'What is MongoDB?' cached (cache re-enabled)")

	stats := ai.GetCacheStats()
	fmt.Printf("\nTotal cached items: %d\n", stats.Size)
	fmt.Println()
}
