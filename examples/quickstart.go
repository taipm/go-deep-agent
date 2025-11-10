package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

// Quickstart: The simplest way to use go-deep-agent with production-ready defaults
//
// WithDefaults() configures:
//   - Memory(20): Keep last 20 messages
//   - Retry(3): Retry failed requests 3 times
//   - Timeout(30s): 30-second timeout
//   - ExponentialBackoff: Smart retry delays (1s, 2s, 4s, ...)
//
// This is the recommended starting point for most use cases.

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Create an agent with production-ready defaults
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).WithDefaults()

	// Have a conversation (memory automatically tracks context)
	ctx := context.Background()

	// First message
	resp1, err := ai.Ask(ctx, "Hi! My name is Alice.")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("AI: %s\n\n", resp1)

	// Second message (AI remembers your name from previous message)
	resp2, err := ai.Ask(ctx, "What's my name?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("AI: %s\n\n", resp2)

	// Third message (conversation continues naturally)
	resp3, err := ai.Ask(ctx, "Can you summarize our conversation so far?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("AI: %s\n", resp3)
}
