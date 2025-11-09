package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// Cache is the interface for response caching
type Cache interface {
	// Get retrieves a cached response
	Get(ctx context.Context, key string) (string, bool, error)

	// Set stores a response in cache with TTL
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Clear removes all keys from cache
	Clear(ctx context.Context) error

	// Stats returns cache statistics
	Stats() CacheStats
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64 // Number of cache hits
	Misses      int64 // Number of cache misses
	Size        int   // Current number of cached items
	Evictions   int64 // Number of evictions (LRU)
	TotalWrites int64 // Total number of writes
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Value       string    // Cached response
	ExpiresAt   time.Time // Expiration time
	CreatedAt   time.Time // Creation time
	AccessedAt  time.Time // Last access time
	AccessCount int64     // Number of times accessed
}

// MemoryCache is an in-memory LRU cache implementation
type MemoryCache struct {
	mu         sync.RWMutex
	entries    map[string]*CacheEntry
	maxSize    int           // Maximum number of entries
	defaultTTL time.Duration // Default TTL
	stats      CacheStats
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int, defaultTTL time.Duration) *MemoryCache {
	if maxSize <= 0 {
		maxSize = 1000 // Default max size
	}
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute // Default TTL
	}

	cache := &MemoryCache{
		entries:    make(map[string]*CacheEntry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(ctx context.Context, key string) (string, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		c.stats.Misses++
		return "", false, nil
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		c.stats.Misses++
		// Don't delete here to avoid deadlock - cleanup will handle it
		return "", false, nil
	}

	// Update access stats (safe because we have read lock and only updating this entry)
	entry.AccessedAt = time.Now()
	entry.AccessCount++

	c.stats.Hits++
	return entry.Value, true, nil
}

// Set stores a value in cache
func (c *MemoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	now := time.Now()

	// Check if we need to evict (LRU)
	if len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	c.entries[key] = &CacheEntry{
		Value:       value,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
		AccessedAt:  now,
		AccessCount: 0,
	}

	c.stats.TotalWrites++
	return nil
}

// Delete removes a key from cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	return nil
}

// Clear removes all entries
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.stats = CacheStats{} // Reset stats
	return nil
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.entries)
	return stats
}

// evictLRU evicts the least recently used entry
func (c *MemoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.AccessedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.stats.Evictions++
	}
}

// cleanup periodically removes expired entries
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// GenerateCacheKey generates a cache key from prompt and configuration
func GenerateCacheKey(model string, prompt string, temperature float64, systemPrompt string) string {
	// Create a deterministic key based on request parameters
	data := struct {
		Model       string
		Prompt      string
		Temperature float64
		System      string
	}{
		Model:       model,
		Prompt:      prompt,
		Temperature: temperature,
		System:      systemPrompt,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// WithCache enables caching with the provided cache implementation
func (b *Builder) WithCache(cache Cache) *Builder {
	b.cache = cache
	b.cacheEnabled = true
	return b
}

// WithMemoryCache enables in-memory caching
func (b *Builder) WithMemoryCache(maxSize int, ttl time.Duration) *Builder {
	cache := NewMemoryCache(maxSize, ttl)
	return b.WithCache(cache)
}

// WithCacheTTL sets the cache TTL for the next request
func (b *Builder) WithCacheTTL(ttl time.Duration) *Builder {
	b.cacheTTL = ttl
	return b
}

// DisableCache disables caching for this builder
func (b *Builder) DisableCache() *Builder {
	b.cacheEnabled = false
	return b
}

// ClearCache clears all cached responses
func (b *Builder) ClearCache(ctx context.Context) error {
	if b.cache == nil {
		return nil
	}
	return b.cache.Clear(ctx)
}

// GetCacheStats returns cache statistics
func (b *Builder) GetCacheStats() CacheStats {
	if b.cache == nil {
		return CacheStats{}
	}
	return b.cache.Stats()
}
