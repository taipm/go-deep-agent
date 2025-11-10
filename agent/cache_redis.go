package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-based cache implementation
type RedisCache struct {
	client     redis.UniversalClient
	prefix     string
	defaultTTL time.Duration
	stats      CacheStats
	statsLock  sync.RWMutex
}

// RedisCacheOptions contains options for Redis cache
type RedisCacheOptions struct {
	// Redis connection
	Addrs    []string // Redis addresses (single: ["localhost:6379"], cluster: multiple)
	Password string   // Redis password
	DB       int      // Database number (0-15, only for single node)

	// Pooling
	PoolSize     int // Connection pool size (default: 10)
	MinIdleConns int // Minimum idle connections (default: 5)

	// Timeouts
	DialTimeout  time.Duration // Dial timeout (default: 5s)
	ReadTimeout  time.Duration // Read timeout (default: 3s)
	WriteTimeout time.Duration // Write timeout (default: 3s)

	// Cache config
	KeyPrefix  string        // Namespace prefix (default: "go-deep-agent")
	DefaultTTL time.Duration // Default TTL (default: 5m)
}

// NewRedisCache creates a new Redis cache with simple configuration
func NewRedisCache(addr, password string, db int, defaultTTL time.Duration) (*RedisCache, error) {
	return NewRedisCacheWithOptions(&RedisCacheOptions{
		Addrs:      []string{addr},
		Password:   password,
		DB:         db,
		DefaultTTL: defaultTTL,
	})
}

// NewRedisCacheWithOptions creates a new Redis cache with advanced options
func NewRedisCacheWithOptions(opts *RedisCacheOptions) (*RedisCache, error) {
	if opts == nil {
		return nil, fmt.Errorf("redis cache options cannot be nil")
	}

	// Set defaults
	if len(opts.Addrs) == 0 {
		opts.Addrs = []string{"localhost:6379"}
	}
	if opts.PoolSize == 0 {
		opts.PoolSize = 10
	}
	if opts.MinIdleConns == 0 {
		opts.MinIdleConns = 5
	}
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 3 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 3 * time.Second
	}
	if opts.KeyPrefix == "" {
		opts.KeyPrefix = "go-deep-agent"
	}
	if opts.DefaultTTL == 0 {
		opts.DefaultTTL = 5 * time.Minute
	}

	// Create Redis client
	var client redis.UniversalClient

	if len(opts.Addrs) == 1 {
		// Single node
		client = redis.NewClient(&redis.Options{
			Addr:         opts.Addrs[0],
			Password:     opts.Password,
			DB:           opts.DB,
			PoolSize:     opts.PoolSize,
			MinIdleConns: opts.MinIdleConns,
			DialTimeout:  opts.DialTimeout,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
		})
	} else {
		// Cluster mode
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        opts.Addrs,
			Password:     opts.Password,
			PoolSize:     opts.PoolSize,
			MinIdleConns: opts.MinIdleConns,
			DialTimeout:  opts.DialTimeout,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), opts.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w\n\n"+
			"Fix:\n"+
			"  1. Check Redis is running: redis-cli ping\n"+
			"  2. Verify connection: redis://localhost:6379\n"+
			"  3. Check firewall/network settings\n"+
			"  4. Start Redis: redis-server or docker run -p 6379:6379 redis\n", err)
	}

	cache := &RedisCache{
		client:     client,
		prefix:     opts.KeyPrefix,
		defaultTTL: opts.DefaultTTL,
		stats:      CacheStats{},
	}

	return cache, nil
}

// makeKey creates a cache key with prefix
func (c *RedisCache) makeKey(key string) string {
	return fmt.Sprintf("%s:cache:%s", c.prefix, key)
}

// statsKey returns the stats key for a given stat type
func (c *RedisCache) statsKey(statType string) string {
	return fmt.Sprintf("%s:stats:%s", c.prefix, statType)
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string) (string, bool, error) {
	redisKey := c.makeKey(key)

	val, err := c.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		// Key not found - cache miss
		c.statsLock.Lock()
		c.stats.Misses++
		c.statsLock.Unlock()

		// Increment miss counter in Redis
		c.client.Incr(ctx, c.statsKey("misses"))

		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("redis get failed: %w\n\n"+
			"Possible causes:\n"+
			"  - Redis connection lost (check redis-cli ping)\n"+
			"  - Network timeout (increase DialTimeout in RedisCacheOptions)\n"+
			"  - Redis server overloaded (check memory/CPU)\n", err)
	}

	// Cache hit
	c.statsLock.Lock()
	c.stats.Hits++
	c.statsLock.Unlock()

	// Increment hit counter in Redis
	c.client.Incr(ctx, c.statsKey("hits"))

	return val, true, nil
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	redisKey := c.makeKey(key)

	if ttl == 0 {
		ttl = c.defaultTTL
	}

	if err := c.client.Set(ctx, redisKey, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	// Update stats
	c.statsLock.Lock()
	c.stats.TotalWrites++
	c.statsLock.Unlock()

	// Increment write counter in Redis
	c.client.Incr(ctx, c.statsKey("writes"))

	return nil
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	redisKey := c.makeKey(key)

	if err := c.client.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}

	return nil
}

// Clear removes all cache keys (using pattern matching)
func (c *RedisCache) Clear(ctx context.Context) error {
	pattern := c.makeKey("*")

	// Use SCAN to find all matching keys
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis scan failed: %w", err)
	}

	// Delete keys in batches
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("redis delete batch failed: %w", err)
		}
	}

	// Clear local stats
	c.statsLock.Lock()
	c.stats = CacheStats{}
	c.statsLock.Unlock()

	// Clear Redis stats
	c.client.Del(ctx,
		c.statsKey("hits"),
		c.statsKey("misses"),
		c.statsKey("writes"),
	)

	return nil
}

// Stats returns cache statistics
func (c *RedisCache) Stats() CacheStats {
	c.statsLock.RLock()
	defer c.statsLock.RUnlock()

	// Get stats from Redis
	ctx := context.Background()

	hits, _ := c.client.Get(ctx, c.statsKey("hits")).Int64()
	misses, _ := c.client.Get(ctx, c.statsKey("misses")).Int64()
	writes, _ := c.client.Get(ctx, c.statsKey("writes")).Int64()

	// Get current size (count of cache keys)
	pattern := c.makeKey("*")
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	size := 0
	for iter.Next(ctx) {
		size++
	}

	return CacheStats{
		Hits:        hits,
		Misses:      misses,
		TotalWrites: writes,
		Size:        size,
		Evictions:   0, // Redis handles evictions automatically
	}
}

// Ping checks if Redis connection is alive
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Exists checks if one or more keys exist
func (c *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = c.makeKey(key)
	}

	count, err := c.client.Exists(ctx, redisKeys...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis exists failed: %w", err)
	}

	return count, nil
}

// TTL returns the remaining time to live of a key
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	redisKey := c.makeKey(key)

	ttl, err := c.client.TTL(ctx, redisKey).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}

	return ttl, nil
}

// Expire sets a timeout on a key
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	redisKey := c.makeKey(key)

	if err := c.client.Expire(ctx, redisKey, ttl).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}

	return nil
}

// SetNX sets a key only if it doesn't exist (distributed lock primitive)
func (c *RedisCache) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	redisKey := c.makeKey(key)

	if ttl == 0 {
		ttl = c.defaultTTL
	}

	success, err := c.client.SetNX(ctx, redisKey, value, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx failed: %w", err)
	}

	return success, nil
}

// MGet retrieves multiple values at once
func (c *RedisCache) MGet(ctx context.Context, keys ...string) ([]string, error) {
	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = c.makeKey(key)
	}

	vals, err := c.client.MGet(ctx, redisKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget failed: %w", err)
	}

	result := make([]string, len(vals))
	for i, val := range vals {
		if val != nil {
			result[i] = val.(string)
		}
	}

	return result, nil
}

// MSet sets multiple key-value pairs at once
func (c *RedisCache) MSet(ctx context.Context, pairs map[string]string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// Use pipeline for atomic operations
	pipe := c.client.Pipeline()

	for key, value := range pairs {
		redisKey := c.makeKey(key)
		pipe.Set(ctx, redisKey, value, ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis mset failed: %w", err)
	}

	// Update stats
	c.statsLock.Lock()
	c.stats.TotalWrites += int64(len(pairs))
	c.statsLock.Unlock()

	return nil
}

// DeletePattern deletes all keys matching a pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.makeKey(pattern)

	// Use SCAN to find all matching keys
	iter := c.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis scan failed: %w", err)
	}

	// Delete keys in batches
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("redis delete batch failed: %w", err)
		}
	}

	return nil
}
