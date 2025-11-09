package main

import (
	"context"
	"fmt"
	"os"

	"github.com/taipm/go-deep-agent/agent"
	"github.com/taipm/go-deep-agent/agent/memory"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable")
		return
	}

	// Custom memory configuration with higher importance threshold
	config := memory.DefaultMemoryConfig()
	config.EpisodicThreshold = 0.6 // Only store messages with importance >= 0.6
	config.WorkingCapacity = 20
	config.EpisodicEnabled = true
	config.SemanticEnabled = true

	// Create builder with custom memory
	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a personal assistant with excellent memory").
		WithHierarchicalMemory(config).
		WithMemory()

	ctx := context.Background()

	// Test 1: Important information (should be stored)
	fmt.Println("=== Test 1: Important Information ===")
	resp1, _ := builder.Ask(ctx, "Remember: My birthday is January 15, 1990 and I prefer dark mode")
	fmt.Printf("Assistant: %s\n\n", resp1)

	// Test 2: Casual chat (low importance, might not be stored in episodic)
	fmt.Println("=== Test 2: Casual Chat ===")
	resp2, _ := builder.Ask(ctx, "ok")
	fmt.Printf("Assistant: %s\n\n", resp2)

	// Test 3: Question (medium importance)
	fmt.Println("=== Test 3: Question ===")
	resp3, _ := builder.Ask(ctx, "What's the weather like today?")
	fmt.Printf("Assistant: %s\n\n", resp3)

	// Test 4: Recall important info
	fmt.Println("=== Test 4: Recall Birthday ===")
	resp4, _ := builder.Ask(ctx, "When is my birthday?")
	fmt.Printf("Assistant: %s\n\n", resp4)

	// Check memory stats
	mem := builder.GetMemory()
	stats := mem.Stats(ctx)

	fmt.Println("=== Memory Statistics ===")
	fmt.Printf("Total Messages: %d\n", stats.TotalMessages)
	fmt.Printf("Working Memory: %d/%d\n", stats.WorkingSize, stats.WorkingCapacity)
	fmt.Printf("Episodic Memory: %d (important messages)\n", stats.EpisodicSize)
	fmt.Printf("Semantic Memory: %d (facts)\n", stats.SemanticSize)
	fmt.Printf("Compression Count: %d\n", stats.CompressionCount)

	// Get memory config
	cfg := mem.GetConfig()
	fmt.Printf("\nMemory Config:\n")
	fmt.Printf("- Episodic Threshold: %.2f\n", cfg.EpisodicThreshold)
	fmt.Printf("- Working Capacity: %d\n", cfg.WorkingCapacity)
	fmt.Printf("- Auto Compress: %v\n", cfg.AutoCompress)
	fmt.Printf("- Episodic Enabled: %v\n", cfg.EpisodicEnabled)
}
