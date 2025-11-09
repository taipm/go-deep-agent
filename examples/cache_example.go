package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	fmt.Println("=== Response Caching Examples ===\n")

	// Example 1: Basic caching
	basicCaching(apiKey)

	// Example 2: Cache with custom TTL
	cachingWithTTL(apiKey)

	// Example 3: Cache statistics
	cacheStats(apiKey)

	// Example 4: Cache clearing
	cacheClear(apiKey)

	// Example 5: Disabling cache
	cacheDisable(apiKey)
}

// basicCaching demonstrates simple response caching
func basicCaching(apiKey string) {
	fmt.Println("1. Basic Caching")
	fmt.Println("----------------")

	// Create agent with memory cache (100 entries, 5 minute TTL)
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemoryCache(100, 5*time.Minute).
		WithTemperature(0.7)

	ctx := context.Background()

	question := "What is 2 + 2?"

	// First call - cache miss, will call API
	fmt.Println("First call (cache miss):")
	start := time.Now()
	response1, err := ai.Ask(ctx, question)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	duration1 := time.Since(start)

	fmt.Printf("Q: %s\n", question)
	fmt.Printf("A: %s\n", response1)
	fmt.Printf("Duration: %v\n\n", duration1)

	// Second call - cache hit, instant response
	fmt.Println("Second call (cache hit):")
	start = time.Now()
	response2, err := ai.Ask(ctx, question)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	duration2 := time.Since(start)

	fmt.Printf("Q: %s\n", question)
	fmt.Printf("A: %s\n", response2)
	fmt.Printf("Duration: %v\n", duration2)
	fmt.Printf("Speedup: %.1fx faster\n\n", float64(duration1)/float64(duration2))
}

// cachingWithTTL demonstrates custom TTL per request
func cachingWithTTL(apiKey string) {
	fmt.Println("2. Caching with Custom TTL")
	fmt.Println("--------------------------")

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemoryCache(100, 1*time.Minute) // Default 1 minute

	ctx := context.Background()

	// Cache with short TTL (2 seconds)
	fmt.Println("Setting cache with 2 second TTL...")
	ai = ai.WithCacheTTL(2 * time.Second)

	response, err := ai.Ask(ctx, "What's the time?")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Response: %s\n\n", response)

	// Immediate second call - should hit cache
	fmt.Println("Immediate second call (should hit cache):")
	start := time.Now()
	response2, _ := ai.Ask(ctx, "What's the time?")
	fmt.Printf("Response: %s (took %v)\n\n", response2, time.Since(start))

	// Wait for expiration
	fmt.Println("Waiting 3 seconds for cache to expire...")
	time.Sleep(3 * time.Second)

	// Third call - cache should be expired
	fmt.Println("Third call after expiration (cache miss):")
	start = time.Now()
	response3, _ := ai.Ask(ctx, "What's the time?")
	fmt.Printf("Response: %s (took %v)\n\n", response3, time.Since(start))
}

// cacheStats demonstrates cache statistics tracking
func cacheStats(apiKey string) {
	fmt.Println("3. Cache Statistics")
	fmt.Println("-------------------")

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemoryCache(10, 5*time.Minute).
		WithTemperature(0.5)

	ctx := context.Background()

	questions := []string{
		"What is the capital of France?",
		"What is 5 + 5?",
		"What is the capital of France?", // Duplicate - cache hit
		"What is the speed of light?",
		"What is 5 + 5?", // Duplicate - cache hit
	}

	for i, q := range questions {
		fmt.Printf("Question %d: %s\n", i+1, q)
		response, err := ai.Ask(ctx, q)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Answer: %s\n\n", response)

		// Small delay to avoid rate limits
		time.Sleep(500 * time.Millisecond)
	}

	// Print cache statistics
	stats := ai.GetCacheStats()
	fmt.Println("\n=== Cache Statistics ===")
	fmt.Printf("Total Requests: %d\n", stats.Hits+stats.Misses)
	fmt.Printf("Cache Hits: %d\n", stats.Hits)
	fmt.Printf("Cache Misses: %d\n", stats.Misses)
	fmt.Printf("Hit Rate: %.1f%%\n", float64(stats.Hits)/float64(stats.Hits+stats.Misses)*100)
	fmt.Printf("Cache Size: %d entries\n", stats.Size)
	fmt.Printf("Total Writes: %d\n", stats.TotalWrites)
	fmt.Printf("Evictions: %d\n\n", stats.Evictions)
}

// cacheClear demonstrates clearing the cache
func cacheClear(apiKey string) {
	fmt.Println("4. Cache Clearing")
	fmt.Println("-----------------")

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemoryCache(10, 5*time.Minute)

	ctx := context.Background()

	// Add some entries
	fmt.Println("Adding entries to cache...")
	ai.Ask(ctx, "Question 1")
	ai.Ask(ctx, "Question 2")
	ai.Ask(ctx, "Question 3")

	stats := ai.GetCacheStats()
	fmt.Printf("Cache size before clear: %d\n\n", stats.Size)

	// Clear cache
	fmt.Println("Clearing cache...")
	err := ai.ClearCache(ctx)
	if err != nil {
		log.Printf("Error clearing cache: %v\n", err)
		return
	}

	stats = ai.GetCacheStats()
	fmt.Printf("Cache size after clear: %d\n", stats.Size)
	fmt.Printf("All stats reset: Hits=%d, Misses=%d, Writes=%d\n\n", stats.Hits, stats.Misses, stats.TotalWrites)
}

// cacheDisable demonstrates disabling cache
func cacheDisable(apiKey string) {
	fmt.Println("5. Disabling Cache")
	fmt.Println("------------------")

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemoryCache(10, 5*time.Minute)

	ctx := context.Background()

	// First call with cache enabled
	fmt.Println("Call with cache enabled:")
	start := time.Now()
	response1, err := ai.Ask(ctx, "What is Go?")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	duration1 := time.Since(start)
	fmt.Printf("Response: %s (took %v)\n\n", response1, duration1)

	// Second call with cache enabled (should be cached)
	fmt.Println("Second call with cache (should be instant):")
	start = time.Now()
	response2, _ := ai.Ask(ctx, "What is Go?")
	duration2 := time.Since(start)
	fmt.Printf("Response: %s (took %v)\n\n", response2, duration2)

	// Disable cache
	ai = ai.DisableCache()

	// Third call with cache disabled (will call API again)
	fmt.Println("Third call with cache DISABLED:")
	start = time.Now()
	response3, _ := ai.Ask(ctx, "What is Go?")
	duration3 := time.Since(start)
	fmt.Printf("Response: %s (took %v)\n\n", response3, duration3)

	fmt.Printf("Cached call was %.1fx faster\n", float64(duration1)/float64(duration2))
	fmt.Printf("Disabled cache call took similar time to first: %v vs %v\n\n", duration3, duration1)
}
