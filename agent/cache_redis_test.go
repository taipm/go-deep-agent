package agent

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// setupMiniRedis creates a miniredis server for testing
func setupMiniRedis(t *testing.T) (*miniredis.Miniredis, *RedisCache) {
	t.Helper()

	// Start miniredis
	mr := miniredis.RunT(t)

	// Create Redis cache
	cache, err := NewRedisCache(mr.Addr(), "", 0, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create Redis cache: %v", err)
	}

	return mr, cache
}

func TestNewRedisCache(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cache, err := NewRedisCache(mr.Addr(), "", 0, 5*time.Minute)
	if err != nil {
		t.Fatalf("NewRedisCache failed: %v", err)
	}
	defer cache.Close()

	if cache == nil {
		t.Fatal("Expected cache to be created")
	}

	// Test connection
	ctx := context.Background()
	if err := cache.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestNewRedisCacheWithOptions(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	opts := &RedisCacheOptions{
		Addrs:      []string{mr.Addr()},
		Password:   "",
		DB:         0,
		PoolSize:   5,
		KeyPrefix:  "test-prefix",
		DefaultTTL: 10 * time.Minute,
	}

	cache, err := NewRedisCacheWithOptions(opts)
	if err != nil {
		t.Fatalf("NewRedisCacheWithOptions failed: %v", err)
	}
	defer cache.Close()

	if cache.prefix != "test-prefix" {
		t.Errorf("Expected prefix 'test-prefix', got '%s'", cache.prefix)
	}

	if cache.defaultTTL != 10*time.Minute {
		t.Errorf("Expected defaultTTL 10m, got %v", cache.defaultTTL)
	}
}

func TestNewRedisCacheWithNilOptions(t *testing.T) {
	_, err := NewRedisCacheWithOptions(nil)
	if err == nil {
		t.Error("Expected error with nil options")
	}
}

func TestNewRedisCacheConnectionFailed(t *testing.T) {
	// Use invalid address
	_, err := NewRedisCache("localhost:9999", "", 0, 5*time.Minute)
	if err == nil {
		t.Error("Expected error with invalid address")
	}
}

func TestRedisCacheSetGet(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set value
	err := cache.Set(ctx, "key1", "value1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get value
	val, found, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !found {
		t.Error("Expected key to be found")
	}

	if val != "value1" {
		t.Errorf("Expected 'value1', got '%s'", val)
	}

	// Check stats
	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.TotalWrites != 1 {
		t.Errorf("Expected 1 write, got %d", stats.TotalWrites)
	}
}

func TestRedisCacheGetMiss(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Get non-existent key
	val, found, err := cache.Get(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if found {
		t.Error("Expected key not to be found")
	}

	if val != "" {
		t.Errorf("Expected empty value, got '%s'", val)
	}

	// Check stats
	stats := cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

func TestRedisCacheSetWithDefaultTTL(t *testing.T) {
	mr, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set with zero TTL (should use default)
	err := cache.Set(ctx, "key1", "value1", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Check TTL in miniredis
	ttl := mr.TTL("go-deep-agent:cache:key1")
	if ttl == 0 {
		t.Error("Expected TTL to be set")
	}
	if ttl > 5*time.Minute {
		t.Errorf("Expected TTL <= 5m, got %v", ttl)
	}
}

func TestRedisCacheDelete(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set value
	cache.Set(ctx, "key1", "value1", 5*time.Minute)

	// Delete
	err := cache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, found, _ := cache.Get(ctx, "key1")
	if found {
		t.Error("Expected key to be deleted")
	}
}

func TestRedisCacheClear(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set multiple values
	cache.Set(ctx, "key1", "value1", 5*time.Minute)
	cache.Set(ctx, "key2", "value2", 5*time.Minute)
	cache.Set(ctx, "key3", "value3", 5*time.Minute)

	// Clear all
	err := cache.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify all deleted
	_, found1, _ := cache.Get(ctx, "key1")
	_, found2, _ := cache.Get(ctx, "key2")
	_, found3, _ := cache.Get(ctx, "key3")

	if found1 || found2 || found3 {
		t.Error("Expected all keys to be deleted")
	}

	// Check stats cleared
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 3 {
		t.Errorf("Expected stats to be cleared, got Hits=%d, Misses=%d", stats.Hits, stats.Misses)
	}
}

func TestRedisCacheStats(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Initial stats
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 || stats.TotalWrites != 0 {
		t.Error("Expected initial stats to be zero")
	}

	// Perform operations
	cache.Set(ctx, "key1", "value1", 5*time.Minute)
	cache.Set(ctx, "key2", "value2", 5*time.Minute)
	cache.Get(ctx, "key1")      // hit
	cache.Get(ctx, "key1")      // hit
	cache.Get(ctx, "nonexist")  // miss

	// Check stats
	stats = cache.Stats()
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.TotalWrites != 2 {
		t.Errorf("Expected 2 writes, got %d", stats.TotalWrites)
	}
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}
}

func TestRedisCacheExists(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set values
	cache.Set(ctx, "key1", "value1", 5*time.Minute)
	cache.Set(ctx, "key2", "value2", 5*time.Minute)

	// Check existence
	count, err := cache.Exists(ctx, "key1", "key2", "key3")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 keys to exist, got %d", count)
	}
}

func TestRedisCacheTTL(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set with specific TTL
	cache.Set(ctx, "key1", "value1", 10*time.Minute)

	// Get TTL
	ttl, err := cache.TTL(ctx, "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}

	if ttl <= 0 {
		t.Error("Expected positive TTL")
	}

	if ttl > 10*time.Minute {
		t.Errorf("Expected TTL <= 10m, got %v", ttl)
	}
}

func TestRedisCacheExpire(t *testing.T) {
	mr, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set value
	cache.Set(ctx, "key1", "value1", 10*time.Minute)

	// Change TTL
	err := cache.Expire(ctx, "key1", 1*time.Minute)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}

	// Check new TTL
	ttl := mr.TTL("go-deep-agent:cache:key1")
	if ttl > 1*time.Minute {
		t.Errorf("Expected TTL <= 1m, got %v", ttl)
	}
}

func TestRedisCacheSetNX(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// First SetNX should succeed
	success, err := cache.SetNX(ctx, "lock1", "value1", 1*time.Minute)
	if err != nil {
		t.Fatalf("SetNX failed: %v", err)
	}
	if !success {
		t.Error("Expected first SetNX to succeed")
	}

	// Second SetNX with same key should fail
	success, err = cache.SetNX(ctx, "lock1", "value2", 1*time.Minute)
	if err != nil {
		t.Fatalf("SetNX failed: %v", err)
	}
	if success {
		t.Error("Expected second SetNX to fail")
	}

	// Verify original value unchanged
	val, found, _ := cache.Get(ctx, "lock1")
	if !found || val != "value1" {
		t.Error("Expected original value to remain")
	}
}

func TestRedisCacheMGet(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set multiple values
	cache.Set(ctx, "key1", "value1", 5*time.Minute)
	cache.Set(ctx, "key2", "value2", 5*time.Minute)
	cache.Set(ctx, "key3", "value3", 5*time.Minute)

	// MGet
	vals, err := cache.MGet(ctx, "key1", "key2", "key3", "nonexist")
	if err != nil {
		t.Fatalf("MGet failed: %v", err)
	}

	if len(vals) != 4 {
		t.Errorf("Expected 4 values, got %d", len(vals))
	}

	if vals[0] != "value1" || vals[1] != "value2" || vals[2] != "value3" {
		t.Errorf("Unexpected values: %v", vals)
	}

	if vals[3] != "" {
		t.Errorf("Expected empty string for nonexistent key, got '%s'", vals[3])
	}
}

func TestRedisCacheMSet(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// MSet
	pairs := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	err := cache.MSet(ctx, pairs, 5*time.Minute)
	if err != nil {
		t.Fatalf("MSet failed: %v", err)
	}

	// Verify all set
	val1, found1, _ := cache.Get(ctx, "key1")
	val2, found2, _ := cache.Get(ctx, "key2")
	val3, found3, _ := cache.Get(ctx, "key3")

	if !found1 || !found2 || !found3 {
		t.Error("Expected all keys to be set")
	}

	if val1 != "value1" || val2 != "value2" || val3 != "value3" {
		t.Errorf("Unexpected values: %s, %s, %s", val1, val2, val3)
	}
}

func TestRedisCacheDeletePattern(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set multiple values with pattern
	cache.Set(ctx, "user:1:name", "Alice", 5*time.Minute)
	cache.Set(ctx, "user:1:age", "30", 5*time.Minute)
	cache.Set(ctx, "user:2:name", "Bob", 5*time.Minute)
	cache.Set(ctx, "product:1", "Item", 5*time.Minute)

	// Delete pattern user:1:*
	err := cache.DeletePattern(ctx, "user:1:*")
	if err != nil {
		t.Fatalf("DeletePattern failed: %v", err)
	}

	// Verify user:1:* deleted
	_, found1, _ := cache.Get(ctx, "user:1:name")
	_, found2, _ := cache.Get(ctx, "user:1:age")
	if found1 || found2 {
		t.Error("Expected user:1:* to be deleted")
	}

	// Verify others still exist
	_, found3, _ := cache.Get(ctx, "user:2:name")
	_, found4, _ := cache.Get(ctx, "product:1")
	if !found3 || !found4 {
		t.Error("Expected other keys to remain")
	}
}

func TestRedisCachePing(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	err := cache.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestRedisCacheClose(t *testing.T) {
	_, cache := setupMiniRedis(t)

	err := cache.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestRedisCacheKeyPrefix(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	opts := &RedisCacheOptions{
		Addrs:      []string{mr.Addr()},
		KeyPrefix:  "myapp",
		DefaultTTL: 5 * time.Minute,
	}

	cache, _ := NewRedisCacheWithOptions(opts)
	defer cache.Close()

	ctx := context.Background()

	// Set value
	cache.Set(ctx, "key1", "value1", 5*time.Minute)

	// Check key in miniredis has correct prefix
	if !mr.Exists("myapp:cache:key1") {
		t.Error("Expected key with 'myapp' prefix")
	}
}

func TestRedisCacheMultipleOperations(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Perform multiple operations
	for i := 0; i < 10; i++ {
		key := "key" + string(rune('0'+i))
		val := "value" + string(rune('0'+i))
		cache.Set(ctx, key, val, 5*time.Minute)
	}

	// Get all
	for i := 0; i < 10; i++ {
		key := "key" + string(rune('0'+i))
		expectedVal := "value" + string(rune('0'+i))
		val, found, _ := cache.Get(ctx, key)
		if !found {
			t.Errorf("Expected key %s to be found", key)
		}
		if val != expectedVal {
			t.Errorf("Expected %s, got %s", expectedVal, val)
		}
	}

	// Check stats
	stats := cache.Stats()
	if stats.TotalWrites != 10 {
		t.Errorf("Expected 10 writes, got %d", stats.TotalWrites)
	}
	if stats.Hits != 10 {
		t.Errorf("Expected 10 hits, got %d", stats.Hits)
	}
	if stats.Size != 10 {
		t.Errorf("Expected size 10, got %d", stats.Size)
	}
}

func TestRedisCacheEmptyValue(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set empty value
	err := cache.Set(ctx, "emptykey", "", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get empty value
	val, found, err := cache.Get(ctx, "emptykey")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !found {
		t.Error("Expected key to be found")
	}

	if val != "" {
		t.Errorf("Expected empty string, got '%s'", val)
	}
}

func TestRedisCacheConcurrentAccess(t *testing.T) {
	_, cache := setupMiniRedis(t)
	defer cache.Close()

	ctx := context.Background()

	// Set initial value
	cache.Set(ctx, "counter", "0", 5*time.Minute)

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, found, err := cache.Get(ctx, "counter")
			if err != nil || !found {
				t.Errorf("Concurrent get failed: err=%v, found=%v", err, found)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
