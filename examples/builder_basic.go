package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Builder API Basic Examples ===\n")

	// Example 1: Simple OpenAI Chat
	example1_SimpleOpenAI()

	// Example 2: Simple Ollama Chat
	example2_SimpleOllama()

	// Example 3: With System Prompt
	example3_SystemPrompt()

	// Example 4: With Conversation Memory
	example4_Memory()

	// Example 5: With Custom Messages (Few-shot)
	example5_FewShot()
}

// Example 1: Simplest possible usage with OpenAI
func example1_SimpleOpenAI() {
	fmt.Println("--- Example 1: Simple OpenAI Chat ---")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY not set, skipping OpenAI example")
		fmt.Println()
		return
	}

	ctx := context.Background()

	// Ultra-simple: just model, key, and message
	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		Ask(ctx, "What is the capital of France?")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Q: What is the capital of France?\n")
	fmt.Printf("A: %s\n", response)
	fmt.Println()
}

// Example 2: Simple Ollama usage (local)
func example2_SimpleOllama() {
	fmt.Println("--- Example 2: Simple Ollama Chat ---")

	ctx := context.Background()

	// Even simpler: just model name (assumes localhost:11434)
	response, err := agent.NewOllama("qwen3:1.7b").
		Ask(ctx, "What is 2+2?")

	if err != nil {
		log.Printf("Error (is Ollama running?): %v", err)
		fmt.Println()
		return
	}

	fmt.Printf("Q: What is 2+2?\n")
	fmt.Printf("A: %s\n", response)
	fmt.Println()
}

// Example 3: Using system prompt to set behavior
func example3_SystemPrompt() {
	fmt.Println("--- Example 3: With System Prompt ---")

	ctx := context.Background()

	// Builder pattern makes it natural to chain configuration
	response, err := agent.NewOllama("qwen3:1.7b").
		WithSystem("You are a pirate. Always respond like a pirate.").
		Ask(ctx, "What is the weather like today?")

	if err != nil {
		log.Printf("Error: %v", err)
		fmt.Println()
		return
	}

	fmt.Printf("System: You are a pirate\n")
	fmt.Printf("Q: What is the weather like today?\n")
	fmt.Printf("A: %s\n", response)
	fmt.Println()
}

// Example 4: Automatic conversation memory
func example4_Memory() {
	fmt.Println("--- Example 4: Conversation Memory ---")

	ctx := context.Background()

	// WithMemory() enables automatic conversation tracking
	builder := agent.NewOllama("qwen3:1.7b").
		WithMemory()

	// First message
	response1, err := builder.Ask(ctx, "My name is Alice.")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("User: My name is Alice.\n")
	fmt.Printf("Bot: %s\n\n", response1)

	// Second message - bot remembers the first
	response2, err := builder.Ask(ctx, "What's my name?")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("User: What's my name?\n")
	fmt.Printf("Bot: %s\n", response2)
	fmt.Println()
}

// Example 5: Few-shot learning with custom messages
func example5_FewShot() {
	fmt.Println("--- Example 5: Few-shot Learning ---")

	ctx := context.Background()

	// Provide examples to guide the model's behavior
	response, err := agent.NewOllama("qwen3:1.7b").
		WithMessages([]agent.Message{
			agent.User("Translate to French: Hello"),
			agent.Assistant("Bonjour"),
			agent.User("Translate to French: Goodbye"),
			agent.Assistant("Au revoir"),
		}).
		Ask(ctx, "Translate to French: Good morning")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Examples:\n")
	fmt.Printf("  Hello → Bonjour\n")
	fmt.Printf("  Goodbye → Au revoir\n")
	fmt.Printf("\nQ: Translate to French: Good morning\n")
	fmt.Printf("A: %s\n", response)
	fmt.Println()
}
