package agent

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis creates a miniredis server for testing
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *RedisBackend) {
	// Create mini Redis server
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	// Create backend
	backend := NewRedisBackend(mr.Addr())

	return mr, backend
}

func TestRedisBackend_NewRedisBackend(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	// Check defaults
	assert.Equal(t, "go-deep-agent:memories:", backend.prefix)
	assert.Equal(t, 7*24*time.Hour, backend.ttl)
	assert.NotNil(t, backend.client)
}

func TestRedisBackend_NewRedisBackendWithOptions(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	opts := &RedisBackendOptions{
		Addr:     mr.Addr(),
		Password: "secret",
		DB:       2,
		TTL:      24 * time.Hour,
		Prefix:   "myapp:memories:",
		PoolSize: 20,
	}

	backend := NewRedisBackendWithOptions(opts)

	assert.Equal(t, "myapp:memories:", backend.prefix)
	assert.Equal(t, 24*time.Hour, backend.ttl)
	assert.NotNil(t, backend.client)
}

func TestRedisBackend_NewRedisBackendWithClient(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Custom client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	backend := NewRedisBackendWithClient(client)

	assert.Equal(t, "go-deep-agent:memories:", backend.prefix)
	assert.Equal(t, 7*24*time.Hour, backend.ttl)
	assert.NotNil(t, backend.client)
}

func TestRedisBackend_FluentAPI(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	backend := NewRedisBackend(mr.Addr()).
		WithTTL(1 * time.Hour).
		WithPrefix("custom:")

	assert.Equal(t, "custom:", backend.prefix)
	assert.Equal(t, 1*time.Hour, backend.ttl)
}

func TestRedisBackend_SaveAndLoad(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	memoryID := "user-alice"

	// Create test messages
	messages := []Message{
		User("Hello"),
		Assistant("Hi there!"),
	}

	// Save
	err := backend.Save(ctx, memoryID, messages)
	require.NoError(t, err)

	// Load
	loaded, err := backend.Load(ctx, memoryID)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Verify
	assert.Len(t, loaded, 2)
	assert.Equal(t, "Hello", loaded[0].Content)
	assert.Equal(t, "Hi there!", loaded[1].Content)
}

func TestRedisBackend_Load_NonExistent(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Load non-existent memory
	loaded, err := backend.Load(ctx, "non-existent")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestRedisBackend_Load_EmptyID(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Load with empty ID
	_, err := backend.Load(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory ID cannot be empty")
}

func TestRedisBackend_Save_EmptyID(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Save with empty ID
	err := backend.Save(ctx, "", []Message{User("test")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory ID cannot be empty")
}

func TestRedisBackend_Delete(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	memoryID := "user-bob"

	// Save
	messages := []Message{User("Hello")}
	err := backend.Save(ctx, memoryID, messages)
	require.NoError(t, err)

	// Verify exists
	loaded, err := backend.Load(ctx, memoryID)
	require.NoError(t, err)
	assert.NotNil(t, loaded)

	// Delete
	err = backend.Delete(ctx, memoryID)
	require.NoError(t, err)

	// Verify deleted
	loaded, err = backend.Load(ctx, memoryID)
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestRedisBackend_Delete_NonExistent(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Delete non-existent (should not error)
	err := backend.Delete(ctx, "non-existent")
	assert.NoError(t, err)
}

func TestRedisBackend_Delete_EmptyID(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Delete with empty ID
	err := backend.Delete(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory ID cannot be empty")
}

func TestRedisBackend_List(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Save multiple memories
	err := backend.Save(ctx, "user-alice", []Message{User("Hello")})
	require.NoError(t, err)
	err = backend.Save(ctx, "user-bob", []Message{User("Hi")})
	require.NoError(t, err)
	err = backend.Save(ctx, "user-charlie", []Message{User("Hey")})
	require.NoError(t, err)

	// List
	memoryIDs, err := backend.List(ctx)
	require.NoError(t, err)

	// Verify
	assert.Len(t, memoryIDs, 3)
	assert.Contains(t, memoryIDs, "user-alice")
	assert.Contains(t, memoryIDs, "user-bob")
	assert.Contains(t, memoryIDs, "user-charlie")
}

func TestRedisBackend_List_Empty(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// List when empty
	memoryIDs, err := backend.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, memoryIDs)
}

func TestRedisBackend_List_WithPrefix(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Backend with custom prefix
	backend := NewRedisBackend(mr.Addr()).
		WithPrefix("myapp:")

	ctx := context.Background()

	// Save memories
	err = backend.Save(ctx, "user-alice", []Message{User("Hello")})
	require.NoError(t, err)

	// List should only return memories with our prefix
	memoryIDs, err := backend.List(ctx)
	require.NoError(t, err)
	assert.Len(t, memoryIDs, 1)
	assert.Contains(t, memoryIDs, "user-alice")
}

func TestRedisBackend_TTL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Backend with short TTL for testing
	backend := NewRedisBackend(mr.Addr()).
		WithTTL(1 * time.Second)

	ctx := context.Background()
	memoryID := "user-ttl-test"

	// Save
	err = backend.Save(ctx, memoryID, []Message{User("Hello")})
	require.NoError(t, err)

	// Verify exists
	loaded, err := backend.Load(ctx, memoryID)
	require.NoError(t, err)
	assert.NotNil(t, loaded)

	// Fast-forward time in miniredis
	mr.FastForward(2 * time.Second)

	// Verify expired
	loaded, err = backend.Load(ctx, memoryID)
	require.NoError(t, err)
	assert.Nil(t, loaded, "Memory should have expired")
}

func TestRedisBackend_Ping(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Ping should succeed
	err := backend.Ping(ctx)
	assert.NoError(t, err)
}

func TestRedisBackend_Close(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	// Close should not error
	err := backend.Close()
	assert.NoError(t, err)
}

func TestRedisBackend_LargeConversation(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	memoryID := "user-large"

	// Create large conversation (100 messages)
	messages := make([]Message, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			messages[i] = User("User message " + string(rune(i)))
		} else {
			messages[i] = Assistant("Assistant response " + string(rune(i)))
		}
	}

	// Save
	err := backend.Save(ctx, memoryID, messages)
	require.NoError(t, err)

	// Load
	loaded, err := backend.Load(ctx, memoryID)
	require.NoError(t, err)
	assert.Len(t, loaded, 100)
}

func TestRedisBackend_KeyFormat(t *testing.T) {
	mr, backend := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	memoryID := "user-alice"

	// Save
	err := backend.Save(ctx, memoryID, []Message{User("Hello")})
	require.NoError(t, err)

	// Check key in Redis
	expectedKey := "go-deep-agent:memories:user-alice"
	exists := mr.Exists(expectedKey)
	assert.True(t, exists, "Key should exist with correct format")
}

func TestRedisBackend_Integration_WithBuilder(t *testing.T) {
	// Integration test: RedisBackend with Builder
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	backend := NewRedisBackend(mr.Addr())
	ctx := context.Background()

	// Create builder with Redis backend
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithShortMemory().
		WithLongMemory("user-integration").
		UsingBackend(backend)

	// Add some messages
	builder.messages = []Message{
		User("What's the weather?"),
		Assistant("It's sunny today."),
	}

	// Save
	err = builder.SaveLongMemory(ctx)
	require.NoError(t, err)

	// Load in new builder
	builder2 := NewOpenAI("gpt-4o-mini", "test-key").
		WithShortMemory().
		WithLongMemory("user-integration").
		UsingBackend(backend)

	err = builder2.LoadLongMemory(ctx)
	require.NoError(t, err)

	// Verify loaded
	assert.Len(t, builder2.messages, 2)
	assert.Equal(t, "What's the weather?", builder2.messages[0].Content)
	assert.Equal(t, "It's sunny today.", builder2.messages[1].Content)
}
