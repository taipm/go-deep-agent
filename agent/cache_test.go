package agent

import (
	"context"
	"testing"
	"time"
)

const (
	testCacheModel  = "gpt-4o-mini"
	testCacheAPIKey = "test-key"
)

func TestMemoryCacheBasic(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	ctx := context.Background()

	// Test Set and Get
	err := cache.Set(ctx, "key1", "value1", 1*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set: %v", err)
	}

	value, found, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if !found {
		t.Error("Expected to find key1")
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}
}

func TestMemoryCacheMiss(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	ctx := context.Background()

	value, found, err := cache.Get(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if found {
		t.Error("Expected cache miss")
	}

	if value != "" {
		t.Error("Expected empty value on miss")
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	cache := NewMemoryCache(10, 100*time.Millisecond)
	ctx := context.Background()

	// Set with short TTL
	err := cache.Set(ctx, "expiring", "value", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to set: %v", err)
	}

	// Should be found immediately
	_, found, _ := cache.Get(ctx, "expiring")
	if !found {
		t.Error("Expected to find key immediately")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found, _ = cache.Get(ctx, "expiring")
	if found {
		t.Error("Expected key to be expired")
	}
}

func TestMemoryCacheLRU(t *testing.T) {
	cache := NewMemoryCache(3, 1*time.Minute) // Max 3 entries
	ctx := context.Background()

	// Fill cache
	cache.Set(ctx, "key1", "value1", 1*time.Minute)
	cache.Set(ctx, "key2", "value2", 1*time.Minute)
	cache.Set(ctx, "key3", "value3", 1*time.Minute)

	// Access key1 to make it recently used
	cache.Get(ctx, "key1")

	// Add key4 - should evict key2 (least recently used)
	cache.Set(ctx, "key4", "value4", 1*time.Minute)

	// key2 should be evicted
	_, found, _ := cache.Get(ctx, "key2")
	if found {
		t.Error("Expected key2 to be evicted")
	}

	// key1 should still exist (was recently accessed)
	_, found, _ = cache.Get(ctx, "key1")
	if !found {
		t.Error("Expected key1 to still exist")
	}
}

func TestMemoryCacheDelete(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 1*time.Minute)

	// Verify it exists
	_, found, _ := cache.Get(ctx, "key1")
	if !found {
		t.Error("Expected key1 to exist")
	}

	// Delete it
	err := cache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Verify it's gone
	_, found, _ = cache.Get(ctx, "key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	ctx := context.Background()

	// Add multiple keys
	cache.Set(ctx, "key1", "value1", 1*time.Minute)
	cache.Set(ctx, "key2", "value2", 1*time.Minute)
	cache.Set(ctx, "key3", "value3", 1*time.Minute)

	stats := cache.Stats()
	if stats.Size != 3 {
		t.Errorf("Expected size 3, got %d", stats.Size)
	}

	// Clear cache
	err := cache.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear: %v", err)
	}

	stats = cache.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected size 0 after clear, got %d", stats.Size)
	}

	// Verify keys are gone
	_, found, _ := cache.Get(ctx, "key1")
	if found {
		t.Error("Expected all keys to be cleared")
	}
}

func TestMemoryCacheStats(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	ctx := context.Background()

	// Initial stats
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Expected initial stats to be zero")
	}

	// Add and access
	cache.Set(ctx, "key1", "value1", 1*time.Minute)

	// Hit
	cache.Get(ctx, "key1")
	stats = cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}

	// Miss
	cache.Get(ctx, "nonexistent")
	stats = cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}

	if stats.TotalWrites != 1 {
		t.Errorf("Expected 1 write, got %d", stats.TotalWrites)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	key1 := GenerateCacheKey("gpt-4o-mini", "Hello", 0.7, "You are helpful")
	key2 := GenerateCacheKey("gpt-4o-mini", "Hello", 0.7, "You are helpful")
	key3 := GenerateCacheKey("gpt-4o-mini", "Hi", 0.7, "You are helpful")

	// Same inputs should generate same key
	if key1 != key2 {
		t.Error("Expected same keys for same inputs")
	}

	// Different inputs should generate different keys
	if key1 == key3 {
		t.Error("Expected different keys for different prompts")
	}

	// Key should be hex string
	if len(key1) != 64 { // SHA256 hex = 64 chars
		t.Errorf("Expected key length 64, got %d", len(key1))
	}
}

func TestWithCache(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	agent := NewOpenAI(testCacheModel, testCacheAPIKey).
		WithCache(cache)

	if !agent.cacheEnabled {
		t.Error("Expected cache to be enabled")
	}

	if agent.cache == nil {
		t.Error("Expected cache to be set")
	}
}

func TestWithMemoryCache(t *testing.T) {
	agent := NewOpenAI(testCacheModel, testCacheAPIKey).
		WithMemoryCache(100, 10*time.Minute)

	if !agent.cacheEnabled {
		t.Error("Expected cache to be enabled")
	}

	if agent.cache == nil {
		t.Error("Expected cache to be set")
	}
}

func TestWithCacheTTL(t *testing.T) {
	agent := NewOpenAI(testCacheModel, testCacheAPIKey).
		WithMemoryCache(10, 1*time.Minute).
		WithCacheTTL(30 * time.Second)

	if agent.cacheTTL != 30*time.Second {
		t.Errorf("Expected TTL 30s, got %v", agent.cacheTTL)
	}
}

func TestDisableCache(t *testing.T) {
	agent := NewOpenAI(testCacheModel, testCacheAPIKey).
		WithMemoryCache(10, 1*time.Minute).
		DisableCache()

	if agent.cacheEnabled {
		t.Error("Expected cache to be disabled")
	}
}

func TestCacheIntegrationMock(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	agent := NewOpenAI(testCacheModel, testCacheAPIKey).
		WithCache(cache)

	ctx := context.Background()

	// First call - cache miss (will fail due to invalid key, but should attempt cache)
	_, err := agent.Ask(ctx, "Test question")

	// Should have API error
	if err == nil {
		t.Error("Expected API error with invalid key")
	}

	// Check cache stats
	stats := agent.GetCacheStats()
	if stats.Misses == 0 {
		t.Error("Expected at least one cache miss")
	}
}

func TestClearCache(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 1*time.Minute)

	agent := NewOpenAI(testCacheModel, testCacheAPIKey).
		WithCache(cache)

	err := agent.ClearCache(ctx)
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	stats := agent.GetCacheStats()
	if stats.Size != 0 {
		t.Error("Expected cache to be empty after clear")
	}
}

func TestGetCacheStatsNil(t *testing.T) {
	agent := NewOpenAI(testCacheModel, testCacheAPIKey)

	stats := agent.GetCacheStats()

	// Should return zero stats for nil cache
	if stats.Hits != 0 || stats.Misses != 0 || stats.Size != 0 {
		t.Error("Expected zero stats for nil cache")
	}
}

func TestCacheWithDifferentTemperatures(t *testing.T) {
	// Different temperatures should create different cache keys
	key1 := GenerateCacheKey("gpt-4o-mini", "Hello", 0.5, "")
	key2 := GenerateCacheKey("gpt-4o-mini", "Hello", 0.7, "")

	if key1 == key2 {
		t.Error("Expected different keys for different temperatures")
	}
}

func TestCacheWithDifferentModels(t *testing.T) {
	// Different models should create different cache keys
	key1 := GenerateCacheKey("gpt-4o-mini", "Hello", 0.7, "")
	key2 := GenerateCacheKey("gpt-4o", "Hello", 0.7, "")

	if key1 == key2 {
		t.Error("Expected different keys for different models")
	}
}

func TestCacheEntryAccessCount(t *testing.T) {
	cache := NewMemoryCache(10, 1*time.Minute)
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 1*time.Minute)

	// Access multiple times
	cache.Get(ctx, "key1")
	cache.Get(ctx, "key1")
	cache.Get(ctx, "key1")

	// Entry should track access count
	cache.mu.RLock()
	entry := cache.entries["key1"]
	cache.mu.RUnlock()

	if entry.AccessCount != 3 {
		t.Errorf("Expected access count 3, got %d", entry.AccessCount)
	}
}

func TestCacheDefaultValues(t *testing.T) {
	// Test with zero/negative values
	cache := NewMemoryCache(0, 0)

	if cache.maxSize <= 0 {
		t.Error("Expected positive default maxSize")
	}

	if cache.defaultTTL <= 0 {
		t.Error("Expected positive default TTL")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewMemoryCache(100, 1*time.Minute)
	ctx := context.Background()

	// Test concurrent reads and writes
	done := make(chan bool, 20)

	// Writers
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 10; j++ {
				cache.Set(ctx, "key", "value", 1*time.Minute)
			}
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 10; j++ {
				cache.Get(ctx, "key")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not panic and cache should be consistent
	stats := cache.Stats()
	if stats.Size < 0 {
		t.Error("Invalid cache size after concurrent access")
	}
}

func BenchmarkCacheSet(b *testing.B) {
	cache := NewMemoryCache(1000, 1*time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, "key", "value", 1*time.Minute)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewMemoryCache(1000, 1*time.Minute)
	ctx := context.Background()

	cache.Set(ctx, "key", "value", 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, "key")
	}
}

func BenchmarkGenerateCacheKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCacheKey("gpt-4o-mini", "What is the capital of France?", 0.7, "You are helpful")
	}
}
