package main

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	ctx := context.Background()

	// Example: Using Ollama (local model)
	fmt.Println("=== Testing with Ollama ===")
	ollamaAgent, err := agent.NewAgent(agent.Config{
		Provider: agent.ProviderOllama,
		Model:    "qwen3:1.7b",
		BaseURL:  "http://localhost:11434/v1",
	})
	if err != nil {
		log.Fatalf("Failed to create Ollama agent: %v", err)
	}

	// Simple chat (non-streaming)
	result, err := ollamaAgent.Chat(ctx, "What is the capital of Vietnam? Answer in one sentence.", nil)
	if err != nil {
		log.Printf("Ollama chat error: %v", err)
	} else {
		fmt.Printf("Ollama Response: %s\n\n", result.Content)
	}

	// Streaming example
	fmt.Println("=== Streaming Response ===")
	fmt.Print("Streaming: ")
	result, err = ollamaAgent.Chat(ctx, "Count from 1 to 5", &agent.ChatOptions{
		Stream: true,
		OnStream: func(delta string) {
			fmt.Print(delta)
		},
	})
	if err != nil {
		log.Printf("Streaming error: %v", err)
	}
	fmt.Println("\n")

	// Conversation history
	fmt.Println("=== Conversation History ===")
	result, err = ollamaAgent.Chat(ctx, "", &agent.ChatOptions{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant."),
			openai.UserMessage("What is AI?"),
		},
	})
	if err != nil {
		log.Printf("Chat with history error: %v", err)
	} else {
		fmt.Printf("Response: %s\n", result.Content)
	}
}
