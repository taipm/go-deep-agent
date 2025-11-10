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
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	ctx := context.Background()

	fmt.Println("=== End-to-End Memory Integration Test with OpenAI ===\n")

	// Configure agent with episodic memory
	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithWorkingMemorySize(5).
		WithEpisodicMemory(0.7).
		WithSystem("You are a helpful assistant with excellent memory. When users ask you to remember something, acknowledge it clearly.")

	mem := builder.GetMemory()

	// Test 1: Important message should be stored in episodic
	fmt.Println("Test 1: User explicitly asks to remember something")
	fmt.Println("User: Remember that my birthday is May 5th")

	response1, err := builder.Ask(ctx, "Remember that my birthday is May 5th")
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}
	fmt.Printf("Assistant: %s\n\n", response1)

	time.Sleep(1 * time.Second) // Give memory time to process

	stats1 := mem.Stats(ctx)
	fmt.Printf("📊 Stats after Test 1:\n")
	fmt.Printf("  Working memory: %d messages\n", stats1.WorkingSize)
	fmt.Printf("  Episodic memory: %d messages\n", stats1.EpisodicSize)
	fmt.Printf("  Average importance: %.2f\n\n", stats1.AverageImportance)

	// Test 2: Casual message (low importance)
	fmt.Println("Test 2: Casual conversation (should not store in episodic)")
	fmt.Println("User: What's the weather like?")

	response2, err := builder.Ask(ctx, "What's the weather like?")
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}
	fmt.Printf("Assistant: %s\n\n", response2)

	time.Sleep(1 * time.Second)

	stats2 := mem.Stats(ctx)
	fmt.Printf("📊 Stats after Test 2:\n")
	fmt.Printf("  Working memory: %d messages\n", stats2.WorkingSize)
	fmt.Printf("  Episodic memory: %d messages (should be same as Test 1)\n", stats2.EpisodicSize)
	fmt.Printf("  Average importance: %.2f\n\n", stats2.AverageImportance)

	// Test 3: Another important message
	fmt.Println("Test 3: Another important piece of information")
	fmt.Println("User: I want you to never forget that I prefer Python over JavaScript")

	response3, err := builder.Ask(ctx, "I want you to never forget that I prefer Python over JavaScript")
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}
	fmt.Printf("Assistant: %s\n\n", response3)

	time.Sleep(1 * time.Second)

	stats3 := mem.Stats(ctx)
	fmt.Printf("📊 Stats after Test 3:\n")
	fmt.Printf("  Working memory: %d messages\n", stats3.WorkingSize)
	fmt.Printf("  Episodic memory: %d messages (should increase)\n", stats3.EpisodicSize)
	fmt.Printf("  Average importance: %.2f\n\n", stats3.AverageImportance)

	// Test 4: Recall - ask about previously stored information
	fmt.Println("Test 4: Testing memory recall")
	fmt.Println("User: What did I tell you to remember about my birthday?")

	response4, err := builder.Ask(ctx, "What did I tell you to remember about my birthday?")
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}
	fmt.Printf("Assistant: %s\n\n", response4)

	time.Sleep(1 * time.Second)

	stats4 := mem.Stats(ctx)
	fmt.Printf("📊 Stats after Test 4:\n")
	fmt.Printf("  Working memory: %d messages\n", stats4.WorkingSize)
	fmt.Printf("  Episodic memory: %d messages\n", stats4.EpisodicSize)
	fmt.Printf("  Total messages processed: %d\n\n", stats4.TotalMessages)

	// Test 5: Fill up working memory to trigger compression
	fmt.Println("Test 5: Filling working memory to test auto-compression")
	for i := 1; i <= 3; i++ {
		msg := fmt.Sprintf("This is casual message number %d", i)
		fmt.Printf("User: %s\n", msg)

		response, err := builder.Ask(ctx, msg)
		if err != nil {
			log.Fatalf("Failed to get response: %v", err)
		}
		fmt.Printf("Assistant: %s\n", response[:min(50, len(response))]+"...\n")
		time.Sleep(500 * time.Millisecond)
	}

	stats5 := mem.Stats(ctx)
	fmt.Printf("\n📊 Final Stats:\n")
	fmt.Printf("  Working memory: %d messages (capacity: %d)\n", stats5.WorkingSize, stats5.WorkingCapacity)
	fmt.Printf("  Episodic memory: %d messages\n", stats5.EpisodicSize)
	fmt.Printf("  Total messages: %d\n", stats5.TotalMessages)
	fmt.Printf("  Compressions: %d\n", stats5.CompressionCount)
	if stats5.EpisodicSize > 0 {
		fmt.Printf("  Average importance: %.2f\n", stats5.AverageImportance)
		fmt.Printf("  Oldest episodic: %s\n", stats5.EpisodicOldest.Format("15:04:05"))
		fmt.Printf("  Newest episodic: %s\n", stats5.EpisodicNewest.Format("15:04:05"))
	}

	// Verification
	fmt.Println("\n=== Verification ===")
	passed := 0
	total := 5

	if stats5.EpisodicSize >= 2 {
		fmt.Println("✅ Test 1 & 3: Important messages stored in episodic memory")
		passed++
	} else {
		fmt.Println("❌ Test 1 & 3 FAILED: Expected at least 2 episodic messages")
	}

	if stats2.EpisodicSize == stats1.EpisodicSize {
		fmt.Println("✅ Test 2: Casual message NOT stored in episodic")
		passed++
	} else {
		fmt.Println("❌ Test 2 FAILED: Casual message incorrectly stored")
	}

	if stats5.TotalMessages >= 8 {
		fmt.Println("✅ All messages processed through memory system")
		passed++
	} else {
		fmt.Println("❌ FAILED: Not all messages processed")
	}

	if stats5.WorkingSize <= stats5.WorkingCapacity+1 {
		fmt.Println("✅ Working memory stays within capacity")
		passed++
	} else {
		fmt.Println("❌ FAILED: Working memory exceeded capacity")
	}

	if stats5.AverageImportance >= 0.7 {
		fmt.Println("✅ Average importance meets threshold")
		passed++
	} else {
		fmt.Printf("❌ FAILED: Average importance %.2f below threshold 0.7\n", stats5.AverageImportance)
	}

	fmt.Printf("\n=== Results: %d/%d tests passed ===\n", passed, total)

	if passed == total {
		fmt.Println("🎉 All end-to-end integration tests PASSED!")
	} else {
		fmt.Println("⚠️  Some tests failed - review above")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
