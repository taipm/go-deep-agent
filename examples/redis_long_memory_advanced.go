// Redis Long-Term Memory - Advanced Configuration
//
// This example demonstrates advanced Redis backend features:
// - Custom configuration (password, DB, TTL, prefix)
// - Fluent API vs Options struct
// - Redis Cluster support
// - Multiple users with different TTLs
//
// Prerequisites: Redis running on localhost:6379
//
// Run: go run examples/redis_long_memory_advanced.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	_ = godotenv.Load()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	ctx := context.Background()

	// ===================================================================
	// Example 1: Fluent API Configuration
	// ===================================================================
	fmt.Println("=== Example 1: Fluent API Configuration ===\n")

	backend1 := agent.NewRedisBackend("localhost:6379").
		WithDB(1).                              // Use DB 1 instead of default 0
		WithTTL(24 * time.Hour).                // 1 day expiration
		WithPrefix("myapp:conversations:")      // Custom prefix

	ai1 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithShortMemory().
		WithLongMemory("user-bob").
		UsingBackend(backend1)

	resp1, _ := ai1.Ask(ctx, "Remember: I'm Bob and I like pizza")
	fmt.Printf("Bob: %s\n", resp1)
	fmt.Println("✓ Saved to Redis DB 1 with custom prefix")
	fmt.Println("  Key: myapp:conversations:user-bob")
	fmt.Println("  TTL: 24 hours\n")

	// ===================================================================
	// Example 2: Options Struct Configuration
	// ===================================================================
	fmt.Println("=== Example 2: Options Struct Configuration ===\n")

	opts := &agent.RedisBackendOptions{
		Addr:     "localhost:6379",
		Password: "",                           // No password
		DB:       2,                            // Use DB 2
		TTL:      48 * time.Hour,               // 2 days
		Prefix:   "chatbot:memories:",
		PoolSize: 20,                           // Larger pool for high traffic
	}

	backend2 := agent.NewRedisBackendWithOptions(opts)

	ai2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithShortMemory().
		WithLongMemory("user-charlie").
		UsingBackend(backend2)

	resp2, _ := ai2.Ask(ctx, "Remember: I'm Charlie and I love coffee")
	fmt.Printf("Charlie: %s\n", resp2)
	fmt.Println("✓ Saved to Redis DB 2 with options struct")
	fmt.Println("  Key: chatbot:memories:user-charlie")
	fmt.Println("  TTL: 48 hours")
	fmt.Println("  Pool: 20 connections\n")

	// ===================================================================
	// Example 3: Expert Mode - Custom Redis Client (Cluster)
	// ===================================================================
	fmt.Println("=== Example 3: Custom Redis Client ===\n")
	fmt.Println("(Using single-node client for demo)")
	fmt.Println("In production, use ClusterClient for multi-node setup\n")

	// Expert: Full control over Redis client
	customClient := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     50,
		MinIdleConns: 10,
	})

	backend3 := agent.NewRedisBackendWithClient(customClient).
		WithPrefix("enterprise:sessions:")

	ai3 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithShortMemory().
		WithLongMemory("user-diana").
		UsingBackend(backend3)

	resp3, _ := ai3.Ask(ctx, "Remember: I'm Diana and I work in enterprise")
	fmt.Printf("Diana: %s\n", resp3)
	fmt.Println("✓ Saved using custom Redis client")
	fmt.Println("  Key: enterprise:sessions:user-diana")
	fmt.Println("  Custom timeouts and pool settings applied\n")

	// ===================================================================
	// Example 4: Multi-User Scenario
	// ===================================================================
	fmt.Println("=== Example 4: Multi-User Management ===\n")

	// Shared backend for all users
	sharedBackend := agent.NewRedisBackend("localhost:6379")

	// User 1: Short-lived session (1 hour)
	user1 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithShortMemory().
		WithLongMemory("session-123").
		UsingBackend(
			agent.NewRedisBackend("localhost:6379").
				WithTTL(1 * time.Hour).
				WithPrefix("sessions:"),
		)

	// User 2: Long-lived conversation (7 days)
	user2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithShortMemory().
		WithLongMemory("user-premium-456").
		UsingBackend(
			agent.NewRedisBackend("localhost:6379").
				WithTTL(7 * 24 * time.Hour).
				WithPrefix("premium:"),
		)

	user1.Ask(ctx, "Quick question about weather")
	user2.Ask(ctx, "Start of long conversation about AI")

	fmt.Println("✓ User 1: Anonymous session (1 hour TTL)")
	fmt.Println("✓ User 2: Premium user (7 days TTL)")

	// ===================================================================
	// Example 5: List All Memories Across Different Prefixes
	// ===================================================================
	fmt.Println("\n=== Example 5: List Memories ===\n")

	// List from different backends
	memories1, _ := ai1.ListLongMemories(ctx)
	fmt.Printf("Backend 1 (myapp:conversations:): %d memories\n", len(memories1))

	memories2, _ := ai2.ListLongMemories(ctx)
	fmt.Printf("Backend 2 (chatbot:memories:): %d memories\n", len(memories2))

	memories3, _ := ai3.ListLongMemories(ctx)
	fmt.Printf("Backend 3 (enterprise:sessions:): %d memories\n", len(memories3))

	// ===================================================================
	// Example 6: Manual Save/Load Control
	// ===================================================================
	fmt.Println("\n=== Example 6: Manual Save Control ===\n")

	manualBackend := agent.NewRedisBackend("localhost:6379").
		WithPrefix("manual:")

	aiManual := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithShortMemory().
		WithLongMemory("user-manual").
		UsingBackend(manualBackend).
		WithAutoSaveLongMemory(false)           // Disable auto-save

	// Messages not saved automatically
	aiManual.Ask(ctx, "Message 1")
	aiManual.Ask(ctx, "Message 2")
	aiManual.Ask(ctx, "Message 3")

	fmt.Println("Added 3 messages (not saved yet)")

	// Manual save
	if err := aiManual.SaveLongMemory(ctx); err != nil {
		log.Fatalf("Failed to save: %v", err)
	}
	fmt.Println("✓ Manually saved all 3 messages to Redis")

	// ===================================================================
	// Cleanup
	// ===================================================================
	fmt.Println("\n=== Cleanup ===\n")

	ai1.DeleteLongMemory(ctx)
	ai2.DeleteLongMemory(ctx)
	ai3.DeleteLongMemory(ctx)
	aiManual.DeleteLongMemory(ctx)

	backend1.Close()
	backend2.Close()
	backend3.Close()
	sharedBackend.Close()
	manualBackend.Close()
	customClient.Close()

	fmt.Println("✓ All memories deleted and connections closed")

	fmt.Println("\n=== Advanced Features Summary ===")
	fmt.Println("✓ Fluent API: Chain methods for simple configs")
	fmt.Println("✓ Options struct: All-in-one configuration")
	fmt.Println("✓ Custom client: Full control for experts")
	fmt.Println("✓ Multiple backends: Different configs per use case")
	fmt.Println("✓ TTL per user: Short sessions vs long conversations")
	fmt.Println("✓ Manual control: Disable auto-save for batch operations")
}
