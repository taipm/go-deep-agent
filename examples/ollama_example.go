package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Ollama Builder API Examples ===")
	fmt.Println("Note: Make sure Ollama is running locally on http://localhost:11434")
	fmt.Println("Example: ollama run qwen2.5:3b\n")

	ctx := context.Background()

	// Example 1: Simple Chat with Ollama
	fmt.Println("--- Example 1: Simple Chat ---")
	response, err := agent.NewOllama("qwen2.5:3b").
		Ask(ctx, "What is the capital of Vietnam? Answer in one sentence.")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 2: Streaming Response
	fmt.Println("--- Example 2: Streaming Response ---")
	fmt.Print("Streaming: ")
	response, err = agent.NewOllama("qwen2.5:3b").
		OnStream(func(content string) {
			fmt.Print(content)
		}).
		Stream(ctx, "Count from 1 to 5")
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("\n\n")

	// Example 3: With System Prompt and Temperature
	fmt.Println("--- Example 3: System Prompt & Temperature ---")
	response, err = agent.NewOllama("qwen2.5:3b").
		WithSystem("You are a helpful assistant that speaks concisely.").
		WithTemperature(0.8).
		Ask(ctx, "What is artificial intelligence?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 4: Conversation with Memory
	fmt.Println("--- Example 4: Conversation Memory ---")
	builder := agent.NewOllama("qwen2.5:3b").
		WithMemory().
		WithMaxHistory(10)

	// First message
	response, err = builder.Ask(ctx, "My favorite programming language is Go")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	// Follow-up that requires memory
	response, err = builder.Ask(ctx, "What's my favorite programming language?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Follow-up Response: %s\n\n", response)
	}

	// Example 5: StreamPrint Convenience
	fmt.Println("--- Example 5: StreamPrint Convenience ---")
	response, err = agent.NewOllama("qwen2.5:3b").
		StreamPrint(ctx, "Tell me a short programming joke")
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("\n\n")

	// Example 6: Custom Base URL (if using different port)
	fmt.Println("--- Example 6: Custom Base URL ---")
	response, err = agent.NewOllama("qwen2.5:3b").
		WithBaseURL("http://localhost:11434/v1"). // Default, but can be changed
		Ask(ctx, "What is 2 + 2?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 7: Advanced Configuration
	fmt.Println("--- Example 7: Advanced Configuration ---")
	response, err = agent.NewOllama("qwen2.5:3b").
		WithSystem("You are a concise assistant").
		WithTemperature(0.7).
		WithMaxTokens(200).
		WithTopP(0.9).
		WithMemory().
		WithMaxHistory(10).
		Ask(ctx, "Explain what goroutines are in Go programming")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 8: History Management
	fmt.Println("--- Example 8: History Management ---")
	builder = agent.NewOllama("qwen2.5:3b").WithMemory()

	builder.Ask(ctx, "I'm learning Rust")
	builder.Ask(ctx, "It has a steep learning curve")

	history := builder.GetHistory()
	fmt.Printf("Conversation has %d messages\n", len(history))

	// Clear and start fresh
	builder.Clear()
	response, err = builder.Ask(ctx, "What is your purpose?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("After clear - Response: %s\n\n", response)
	}

	fmt.Println("=== All Ollama Examples Complete ===")
}
