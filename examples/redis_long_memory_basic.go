// Redis Long-Term Memory - Quick Start
//
// This example demonstrates persistent conversation storage using Redis.
// Conversations are automatically saved to Redis and survive application restarts.
//
// Prerequisites:
// 1. Install Redis:
//    - macOS: brew install redis
//    - Ubuntu: sudo apt-get install redis
//    - Docker: docker run -d -p 6379:6379 redis
//
// 2. Start Redis server:
//    redis-server
//
// 3. Verify Redis is running:
//    redis-cli ping  (should return "PONG")
//
// Features demonstrated:
// - Zero-config Redis setup
// - Automatic memory persistence
// - Conversation continuity across restarts
// - TTL-based expiration
//
// Run: go run examples/redis_long_memory_basic.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	ctx := context.Background()

	// Demo Part 1: First conversation - Save to Redis
	fmt.Println("=== PART 1: First Conversation ===")
	fmt.Println("Creating agent with Redis backend...")

	// Create Redis backend (simplest form - just address)
	redisBackend := agent.NewRedisBackend("localhost:6379")
	defer redisBackend.Close()

	// Verify Redis connection
	if err := redisBackend.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to Redis: %v\n"+
			"Make sure Redis is running: redis-server", err)
	}
	fmt.Println("✓ Connected to Redis")

	// Create agent with Redis-backed long-term memory
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithShortMemory().                     // RAM memory for this conversation
		WithLongMemory("user-alice").          // Persistent memory ID
		UsingBackend(redisBackend).            // Use Redis for storage
		WithAutoSaveLongMemory(true)           // Auto-save after each message (default)

	// First conversation
	fmt.Println("\n👤 User: My favorite color is blue, and I love cats.")
	resp1, err := ai.Ask(ctx, "My favorite color is blue, and I love cats.")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("🤖 Assistant: %s\n", resp1)
	// Memory automatically saved to Redis: key = "go-deep-agent:memories:user-alice"

	// Second message
	fmt.Println("\n👤 User: What's my favorite color?")
	resp2, err := ai.Ask(ctx, "What's my favorite color?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("🤖 Assistant: %s\n", resp2)

	fmt.Println("\n✓ Conversation saved to Redis")
	fmt.Println("  Key: go-deep-agent:memories:user-alice")
	fmt.Println("  TTL: 7 days (default)")

	// Simulate application restart
	fmt.Println("\n=== Simulating Application Restart ===")
	fmt.Println("(In real scenario, restart your program)")
	time.Sleep(1 * time.Second)

	// Demo Part 2: New session - Load from Redis
	fmt.Println("\n=== PART 2: After Restart - Load from Redis ===")

	// Create NEW agent instance (simulates restart)
	// Same Redis backend and memory ID
	ai2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithShortMemory().
		WithLongMemory("user-alice").          // Same user ID
		UsingBackend(redisBackend)
	// Previous conversation automatically loaded from Redis!

	// Ask question about previous conversation
	fmt.Println("\n👤 User: What do I love?")
	resp3, err := ai2.Ask(ctx, "What do I love?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("🤖 Assistant: %s\n", resp3)
	// Should mention "cats" from first conversation

	// Verify memory was loaded
	history := ai2.GetHistory()
	fmt.Printf("\n✓ Loaded %d messages from Redis\n", len(history))
	fmt.Println("  Agent remembers previous conversation!")

	// Demo Part 3: List all memories
	fmt.Println("\n=== PART 3: List All Memories ===")

	memoryIDs, err := ai2.ListLongMemories(ctx)
	if err != nil {
		log.Fatalf("Error listing memories: %v", err)
	}

	fmt.Printf("Found %d memory(ies) in Redis:\n", len(memoryIDs))
	for i, id := range memoryIDs {
		fmt.Printf("  %d. %s\n", i+1, id)
	}

	// Optional: Clean up (delete memory)
	fmt.Println("\n=== Optional: Clean Up ===")
	fmt.Println("Do you want to delete the memory? (y/n)")
	// Uncomment to enable interactive cleanup:
	// var answer string
	// fmt.Scanln(&answer)
	// if answer == "y" {
	// 	err = ai2.DeleteLongMemory(ctx)
	// 	if err != nil {
	// 		log.Fatalf("Error deleting memory: %v", err)
	// 	}
	// 	fmt.Println("✓ Memory deleted from Redis")
	// }

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("✓ Redis backend enables persistent memory across restarts")
	fmt.Println("✓ Zero configuration - just provide Redis address")
	fmt.Println("✓ Automatic save/load - no manual intervention needed")
	fmt.Println("✓ TTL ensures memories don't live forever (7 days default)")
}
