package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
	"github.com/taipm/go-deep-agent/agent/memory"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	ctx := context.Background()

	// Example 1: Basic episodic memory configuration
	fmt.Println("=== Example 1: Builder API with Episodic Memory ===")
	builder1 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithWorkingMemorySize(10).
		WithEpisodicMemory(0.7). // Store messages with importance >= 0.7
		WithSystem("You are a helpful assistant with excellent memory.")

	// Add messages using memory package directly
	mem1 := builder1.GetMemory()
	mem1.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "Remember: my birthday is May 5th",
		Timestamp: time.Now(),
	})
	mem1.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "I like coffee in the morning",
		Timestamp: time.Now(),
	})
	mem1.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "The weather is nice today",
		Timestamp: time.Now(),
	})

	// Check memory stats
	stats1 := mem1.Stats(ctx)
	fmt.Printf("Working messages: %d\n", stats1.WorkingSize)
	fmt.Printf("Episodic messages: %d\n", stats1.EpisodicSize)
	fmt.Printf("Average importance: %.2f\n", stats1.AverageImportance)
	fmt.Printf("Config - Episodic enabled: %v, threshold: %.2f\n\n",
		mem1.GetConfig().EpisodicEnabled, mem1.GetConfig().EpisodicThreshold)

	// Example 2: Custom importance weights
	fmt.Println("=== Example 2: Custom Importance Weights ===")
	weights := memory.DefaultImportanceWeights()
	weights.ExplicitRemember = 2.0 // Double weight for "remember this"
	weights.PersonalInfo = 1.5     // Higher weight for personal info
	weights.QuestionAnswer = 0.8   // Higher weight for Q&A

	builder2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithWorkingMemorySize(15).
		WithEpisodicMemory(0.8).
		WithImportanceWeights(weights).
		WithSystem("You are a helpful assistant.")

	mem2 := builder2.GetMemory()
	mem2.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "Remember to always greet me with enthusiasm",
		Timestamp: time.Now(),
	})
	mem2.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "What is the capital of France?",
		Timestamp: time.Now(),
	})

	stats2 := mem2.Stats(ctx)
	fmt.Printf("Working messages: %d\n", stats2.WorkingSize)
	fmt.Printf("Episodic messages: %d\n", stats2.EpisodicSize)
	fmt.Printf("Average importance: %.2f\n", stats2.AverageImportance)
	fmt.Printf("Custom weights - ExplicitRemember: %.1f, PersonalInfo: %.1f\n\n",
		mem2.GetConfig().ImportanceWeights.ExplicitRemember,
		mem2.GetConfig().ImportanceWeights.PersonalInfo)

	// Example 3: Full hierarchical memory (working + episodic + semantic)
	fmt.Println("=== Example 3: Full Hierarchical Memory ===")
	builder3 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithWorkingMemorySize(20).
		WithEpisodicMemory(0.6).
		WithSemanticMemory().
		WithSystem("You are a knowledgeable assistant.")

	mem3 := builder3.GetMemory()
	mem3.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "Remember: I prefer Python over JavaScript",
		Timestamp: time.Now(),
	})
	mem3.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "What is machine learning?",
		Timestamp: time.Now(),
	})

	stats3 := mem3.Stats(ctx)
	fmt.Printf("Working messages: %d\n", stats3.WorkingSize)
	fmt.Printf("Episodic messages: %d\n", stats3.EpisodicSize)
	fmt.Printf("All tiers enabled: working=%v, episodic=%v, semantic=%v\n\n",
		true, mem3.GetConfig().EpisodicEnabled, mem3.GetConfig().SemanticEnabled)

	// Example 4: Using WithHierarchicalMemory for full control
	fmt.Println("=== Example 4: Full Config with WithHierarchicalMemory ===")
	config := memory.DefaultMemoryConfig()
	config.WorkingCapacity = 25
	config.EpisodicEnabled = true
	config.EpisodicThreshold = 0.5
	config.SemanticEnabled = true
	config.AutoCompress = true
	config.ImportanceScoring = true

	builder4 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithHierarchicalMemory(config).
		WithSystem("You are a very capable assistant.")

	mem4 := builder4.GetMemory()
	mem4.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "I need you to remember my preferences",
		Timestamp: time.Now(),
	})
	mem4.Add(ctx, memory.Message{
		Role:      "user",
		Content:   "Never forget important dates",
		Timestamp: time.Now(),
	})

	stats4 := mem4.Stats(ctx)
	fmt.Printf("Working messages: %d\n", stats4.WorkingSize)
	fmt.Printf("Episodic messages: %d\n", stats4.EpisodicSize)
	fmt.Printf("Total messages processed: %d\n", stats4.TotalMessages)
	fmt.Printf("Auto-compress enabled: %v\n\n", mem4.GetConfig().AutoCompress)

	fmt.Println("=== All Examples Completed Successfully ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("1. Use WithEpisodicMemory(threshold) for easy episodic setup")
	fmt.Println("2. Use WithImportanceWeights() to customize what's important")
	fmt.Println("3. Use WithSemanticMemory() to enable fact storage")
	fmt.Println("4. Use WithHierarchicalMemory(config) for full control")
	fmt.Println("5. Use GetMemory() to access memory system directly")
}
