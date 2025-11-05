package main

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Builder API Streaming Examples ===\n")

	// Example 1: Simple Streaming
	example1_SimpleStream()

	// Example 2: Stream with Callbacks
	example2_StreamCallbacks()

	// Example 3: StreamPrint Convenience
	example3_StreamPrint()
}

// Example 1: Simple streaming with OnStream
func example1_SimpleStream() {
	fmt.Println("--- Example 1: Simple Streaming ---")

	ctx := context.Background()

	response, err := agent.NewOllama("qwen3:1.7b").
		OnStream(func(content string) {
			fmt.Print(content) // Print chunks as they arrive
		}).
		Stream(ctx, "Tell me a short joke about programming.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("\n\nComplete response:", response)
	fmt.Println()
}

// Example 2: Stream with multiple callbacks
func example2_StreamCallbacks() {
	fmt.Println("--- Example 2: Stream with Callbacks ---")

	ctx := context.Background()

	chunkCount := 0

	_, err := agent.NewOllama("qwen3:1.7b").
		OnStream(func(content string) {
			chunkCount++
			fmt.Print(content)
		}).
		OnToolCall(func(tool openai.FinishedChatCompletionToolCall) {
			fmt.Printf("\n[Tool Called: %s]\n", tool.Name)
		}).
		OnRefusal(func(refusal string) {
			fmt.Printf("\n[Model Refused: %s]\n", refusal)
		}).
		Stream(ctx, "Write a haiku about Go programming.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("\n\nTotal chunks received: %d\n", chunkCount)
	fmt.Println()
}

// Example 3: StreamPrint convenience method
func example3_StreamPrint() {
	fmt.Println("--- Example 3: StreamPrint Convenience ---")

	ctx := context.Background()

	// StreamPrint automatically prints to stdout
	response, err := agent.NewOllama("qwen3:1.7b").
		StreamPrint(ctx, "What is 2+2?")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("\n\nDone! Response:", response)
}
