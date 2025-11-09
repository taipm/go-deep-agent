package main

import (
	"context"
	"fmt"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable")
		return
	}

	// Create builder with hierarchical memory (auto-enabled by default)
	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant with long-term memory").
		WithMemory() // Enable auto-memory (stores in conversation history)

	ctx := context.Background()

	// Conversation 1: User provides information
	fmt.Println("=== Conversation 1: Learning about user ===")
	resp1, err := builder.Ask(ctx, "My name is Alice and I love Go programming")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Assistant: %s\n\n", resp1)

	// Conversation 2: Model should remember
	fmt.Println("=== Conversation 2: Recall ===")
	resp2, err := builder.Ask(ctx, "What programming language do I like?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Assistant: %s\n\n", resp2)

	// Check memory stats
	mem := builder.GetMemory()
	stats := mem.Stats(ctx)
	fmt.Println("=== Memory Statistics ===")
	fmt.Printf("Total Messages: %d\n", stats.TotalMessages)
	fmt.Printf("Working Memory Size: %d/%d\n", stats.WorkingSize, stats.WorkingCapacity)
	fmt.Printf("Episodic Memory Size: %d\n", stats.EpisodicSize)
	fmt.Printf("Semantic Memory Size: %d\n", stats.SemanticSize)
}
