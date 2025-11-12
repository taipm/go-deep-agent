package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBackend stores long-term memories (conversation history) in Redis.
//
// Features:
//   - TTL support (automatic expiration)
//   - Key prefix for namespacing
//   - Connection pooling built-in
//   - Cluster/Sentinel support via custom client
//
// Default Configuration:
//   - TTL: 7 days (reasonable for conversations)
//   - Prefix: "go-deep-agent:memories:"
//   - DB: 0
//   - Pool: 10 connections
//
// Example (Simple - Beginner):
//
//	backend := NewRedisBackend("localhost:6379")
//	agent := NewOpenAI("gpt-4", apiKey).
//	    WithShortMemory().
//	    WithLongMemory("user-alice").
//	        UsingBackend(backend)
//
// Example (Advanced - Custom Config):
//
//	backend := NewRedisBackend("localhost:6379").
//	    WithPassword("secret").
//	    WithDB(2).
//	    WithTTL(24 * time.Hour).
//	    WithPrefix("myapp:conversations:")
//
// Example (Expert - Custom Client):
//
//	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
//	    Addrs: []string{"node1:6379", "node2:6379"},
//	})
//	backend := NewRedisBackendWithClient(redisClient)
type RedisBackend struct {
	client redis.UniversalClient
	prefix string
	ttl    time.Duration
}

// RedisBackendOptions provides configuration for Redis backend.
//
// Use this when you need to configure multiple options at once:
//
//	opts := &RedisBackendOptions{
//	    Addr:     "localhost:6379",
//	    Password: "secret",
//	    DB:       2,
//	    TTL:      24 * time.Hour,
//	    Prefix:   "myapp:memories:",
//	}
//	backend := NewRedisBackendWithOptions(opts)
type RedisBackendOptions struct {
	// Addr is the Redis server address (host:port)
	// Example: "localhost:6379"
	Addr string

	// Password for Redis authentication (optional)
	// Leave empty if Redis has no password
	Password string

	// DB is the Redis database number (0-15)
	// Default: 0
	DB int

	// TTL is how long memories are kept before auto-expiration
	// Default: 7 days (7 * 24 * time.Hour)
	// Set to 0 for no expiration (not recommended)
	TTL time.Duration

	// Prefix is prepended to all Redis keys for namespacing
	// Default: "go-deep-agent:memories:"
	// Example: "myapp:conversations:"
	Prefix string

	// PoolSize is the maximum number of socket connections
	// Default: 10
	PoolSize int
}

// NewRedisBackend creates a Redis backend with smart defaults.
// This is the recommended way to use Redis for 90% of use cases.
//
// Parameters:
//   - addr: Redis server address (e.g., "localhost:6379")
//
// Smart defaults applied automatically:
//   - Password: none (use WithPassword() if needed)
//   - DB: 0 (use WithDB() to change)
//   - TTL: 7 days (use WithTTL() to change)
//   - Prefix: "go-deep-agent:memories:"
//   - Pool: 10 connections
//
// Quick Start (simplest):
//
//	backend := NewRedisBackend("localhost:6379")
//	defer backend.Close()
//
// With password (common):
//
//	backend := NewRedisBackend("localhost:6379").
//	    WithPassword("secret")
//
// Custom TTL (advanced):
//
//	backend := NewRedisBackend("localhost:6379").
//	    WithPassword("secret").
//	    WithTTL(24 * time.Hour)
//
// For Redis Cluster/Sentinel, use NewRedisBackendWithClient() instead.
func NewRedisBackend(addr string) *RedisBackend {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // No password by default
		DB:       0,  // Default DB
		PoolSize: 10, // Default pool size
	})

	return &RedisBackend{
		client: client,
		prefix: "go-deep-agent:memories:",
		ttl:    7 * 24 * time.Hour, // 7 days default
	}
}

// NewRedisBackendWithOptions creates a Redis backend from options struct.
//
// Use this when configuring multiple options:
//
//	opts := &RedisBackendOptions{
//	    Addr:     "localhost:6379",
//	    Password: "secret",
//	    DB:       2,
//	    TTL:      24 * time.Hour,
//	    Prefix:   "myapp:memories:",
//	    PoolSize: 20,
//	}
//	backend := NewRedisBackendWithOptions(opts)
func NewRedisBackendWithOptions(opts *RedisBackendOptions) *RedisBackend {
	// Apply defaults
	if opts.TTL == 0 {
		opts.TTL = 7 * 24 * time.Hour
	}
	if opts.Prefix == "" {
		opts.Prefix = "go-deep-agent:memories:"
	}
	if opts.PoolSize == 0 {
		opts.PoolSize = 10
	}

	client := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
		PoolSize: opts.PoolSize,
	})

	return &RedisBackend{
		client: client,
		prefix: opts.Prefix,
		ttl:    opts.TTL,
	}
}

// NewRedisBackendWithClient creates a Redis backend with a custom client.
//
// Expert mode: Use this for advanced configurations like:
//   - Redis Cluster
//   - Redis Sentinel
//   - Custom connection settings
//   - Shared client instance
//
// Example (Cluster):
//
//	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
//	    Addrs: []string{"node1:6379", "node2:6379", "node3:6379"},
//	})
//	backend := NewRedisBackendWithClient(redisClient)
func NewRedisBackendWithClient(client redis.UniversalClient) *RedisBackend {
	return &RedisBackend{
		client: client,
		prefix: "go-deep-agent:memories:",
		ttl:    7 * 24 * time.Hour,
	}
}

// WithPassword sets the Redis authentication password.
// Default: "" (no authentication)
//
// Use this when your Redis server requires a password.
//
// Example:
//
//	backend := NewRedisBackend("localhost:6379").
//	    WithPassword("mypassword")
func (r *RedisBackend) WithPassword(password string) *RedisBackend {
	// Recreate client with password
	if client, ok := r.client.(*redis.Client); ok {
		opts := client.Options()
		opts.Password = password
		r.client = redis.NewClient(opts)
	}
	return r
}

// WithDB sets the Redis database number (0-15).
// Default: 0
//
// Use this to isolate memories from other data in Redis.
//
// Common use cases:
//   - DB 0: Production memories (default)
//   - DB 1: Staging/test memories
//   - DB 2: Development memories
//
// Fluent API:
//
//	backend := NewRedisBackend("localhost:6379").
//	    WithDB(2)
func (r *RedisBackend) WithDB(db int) *RedisBackend {
	// Recreate client with new DB
	if client, ok := r.client.(*redis.Client); ok {
		opts := client.Options()
		opts.DB = db
		r.client = redis.NewClient(opts)
	}
	return r
}

// WithTTL sets how long memories are kept before auto-expiration.
// Default: 7 days (168 hours)
//
// TTL (Time To Live) determines when inactive memories expire from Redis.
// Note: TTL is extended on every save, so active conversations never expire.
//
// Common values:
//   - 1 * time.Hour         = 1 hour (anonymous sessions)
//   - 24 * time.Hour        = 1 day (temporary chats)
//   - 7 * 24 * time.Hour    = 7 days (default - recommended)
//   - 30 * 24 * time.Hour   = 30 days (premium users)
//   - 0                     = never expire (not recommended - use with caution)
//
// Example:
//
//	// Expire after 24 hours of inactivity
//	backend := NewRedisBackend("localhost:6379").
//	    WithTTL(24 * time.Hour)
//
// Default: 7 days
// Set to 0 for no expiration (not recommended)
//
// Fluent API:
//
//	backend := NewRedisBackend("localhost:6379").
//	    WithTTL(24 * time.Hour)  // 1 day
func (r *RedisBackend) WithTTL(ttl time.Duration) *RedisBackend {
	r.ttl = ttl
	return r
}

// WithPrefix sets the Redis key prefix for namespacing.
// Default: "go-deep-agent:memories:"
//
// Use this to avoid key collisions when:
//   - Sharing Redis with other apps
//   - Running multiple environments (dev/staging/prod)
//   - Organizing different memory types
//
// Key format: {prefix}{memoryID}
// Example: "myapp:conversations:user-123"
//
// Common prefixes:
//   - "go-deep-agent:memories:"      = default
//   - "myapp:prod:conversations:"    = production environment
//   - "myapp:staging:conversations:" = staging environment
//
// Example:
//
//	backend := NewRedisBackend("localhost:6379").
//	    WithPrefix("myapp:prod:")
func (r *RedisBackend) WithPrefix(prefix string) *RedisBackend {
	r.prefix = prefix
	return r
}

// Load retrieves conversation history from Redis.
//
// Returns:
//   - nil, nil if memory doesn't exist (first time)
//   - messages, nil if successfully loaded
//   - nil, error if Redis operation fails
func (r *RedisBackend) Load(ctx context.Context, memoryID string) ([]Message, error) {
	// Validate memory ID
	if memoryID == "" {
		return nil, fmt.Errorf("memory ID cannot be empty")
	}

	// Construct Redis key
	key := r.prefix + memoryID

	// Get from Redis
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key doesn't exist - this is normal for first time
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get memory from Redis: %w", err)
	}

	// Parse JSON
	var messages []Message
	if err := json.Unmarshal([]byte(data), &messages); err != nil {
		return nil, fmt.Errorf("failed to parse memory JSON: %w", err)
	}

	return messages, nil
}

// Save stores conversation history to Redis with TTL.
//
// Features:
//   - Automatic JSON serialization
//   - TTL extension on every save (active conversations never expire)
//   - Atomic operation
func (r *RedisBackend) Save(ctx context.Context, memoryID string, messages []Message) error {
	// Validate memory ID
	if memoryID == "" {
		return fmt.Errorf("memory ID cannot be empty")
	}

	// Construct Redis key
	key := r.prefix + memoryID

	// Marshal to JSON
	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages to JSON: %w", err)
	}

	// Save to Redis with TTL
	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to save memory to Redis: %w", err)
	}

	return nil
}

// Delete removes a memory from Redis.
//
// Returns nil if memory doesn't exist (idempotent).
func (r *RedisBackend) Delete(ctx context.Context, memoryID string) error {
	// Validate memory ID
	if memoryID == "" {
		return fmt.Errorf("memory ID cannot be empty")
	}

	// Construct Redis key
	key := r.prefix + memoryID

	// Delete from Redis (ignore NotFound)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete memory from Redis: %w", err)
	}

	return nil
}

// List returns all available memory IDs from Redis.
//
// Uses SCAN for safe iteration (doesn't block Redis).
// Returns empty slice if no memories exist.
func (r *RedisBackend) List(ctx context.Context) ([]string, error) {
	// Scan for keys matching prefix
	pattern := r.prefix + "*"
	var memoryIDs []string

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Extract memory ID by removing prefix
		if len(key) > len(r.prefix) {
			memoryID := key[len(r.prefix):]
			memoryIDs = append(memoryIDs, memoryID)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan Redis keys: %w", err)
	}

	return memoryIDs, nil
}

// Ping checks if Redis connection is healthy.
//
// Returns nil if connection is OK, error otherwise.
// Useful for health checks and initialization validation.
func (r *RedisBackend) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection.
//
// Call this when shutting down your application:
//
//	defer backend.Close()
func (r *RedisBackend) Close() error {
	return r.client.Close()
}
