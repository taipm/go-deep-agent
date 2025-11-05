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

	// Run all examples
	fmt.Println("=== Conversation Management Examples ===\n")

	example1_BasicMemory(apiKey)
	example2_GetAndSetHistory(apiKey)
	example3_ClearConversation(apiKey)
	example4_MaxHistoryLimit(apiKey)
	example5_SaveAndRestoreSession(apiKey)
	example6_MemoryVsNoMemory(apiKey)
}

// Example 1: Basic memory - the simplest use case
func example1_BasicMemory(apiKey string) {
	fmt.Println("--- Example 1: Basic Memory ---")

	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant.").
		WithMemory() // Enable automatic conversation memory

	ctx := context.Background()

	// First message
	response, _ := builder.Ask(ctx, "My name is Alice and I love Go programming.")
	fmt.Printf("User: My name is Alice and I love Go programming.\n")
	fmt.Printf("AI: %s\n\n", response)

	// Second message - AI remembers the context
	response, _ = builder.Ask(ctx, "What's my name?")
	fmt.Printf("User: What's my name?\n")
	fmt.Printf("AI: %s\n\n", response)

	// Third message - AI remembers both name and interest
	response, _ = builder.Ask(ctx, "What programming language do I like?")
	fmt.Printf("User: What programming language do I like?\n")
	fmt.Printf("AI: %s\n\n", response)

	fmt.Println()
}

// Example 2: Get and set history - inspect and manipulate conversation
func example2_GetAndSetHistory(apiKey string) {
	fmt.Println("--- Example 2: Get and Set History ---")

	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory()

	ctx := context.Background()

	// Have a conversation
	builder.Ask(ctx, "I have a dog named Max.")
	builder.Ask(ctx, "He is 5 years old.")

	// Get conversation history
	history := builder.GetHistory()
	fmt.Printf("Conversation has %d messages:\n", len(history))
	for i, msg := range history {
		fmt.Printf("  %d. [%s]: %s\n", i+1, msg.Role, msg.Content)
	}

	// Modify history (e.g., remove a message)
	if len(history) > 0 {
		history = history[1:] // Remove first message
	}

	// Set modified history
	builder.SetHistory(history)
	fmt.Printf("\nAfter removing first message, history has %d messages\n", len(builder.GetHistory()))

	// Continue conversation with modified history
	response, _ := builder.Ask(ctx, "What's my dog's name?")
	fmt.Printf("User: What's my dog's name?\n")
	fmt.Printf("AI: %s\n", response)
	fmt.Printf("(AI might not remember the name since we removed that message)\n\n")

	fmt.Println()
}

// Example 3: Clear conversation - start fresh
func example3_ClearConversation(apiKey string) {
	fmt.Println("--- Example 3: Clear Conversation ---")

	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant.").
		WithMemory()

	ctx := context.Background()

	// First conversation
	builder.Ask(ctx, "My favorite color is blue.")
	response, _ := builder.Ask(ctx, "What's my favorite color?")
	fmt.Printf("Before clear:\n")
	fmt.Printf("User: What's my favorite color?\n")
	fmt.Printf("AI: %s\n\n", response)

	// Clear conversation (keeps system prompt)
	builder.Clear()
	fmt.Printf("Cleared conversation history (kept system prompt)\n\n")

	// New conversation - AI doesn't remember
	response, _ = builder.Ask(ctx, "What's my favorite color?")
	fmt.Printf("After clear:\n")
	fmt.Printf("User: What's my favorite color?\n")
	fmt.Printf("AI: %s\n", response)
	fmt.Printf("(AI doesn't remember because we cleared the history)\n\n")

	fmt.Println()
}

// Example 4: Max history limit - automatic truncation
func example4_MaxHistoryLimit(apiKey string) {
	fmt.Println("--- Example 4: Max History Limit ---")

	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant. Respond briefly.").
		WithMemory().
		WithMaxHistory(4) // Keep only last 4 messages (2 exchanges)

	ctx := context.Background()

	// Have multiple exchanges
	fmt.Println("Having 5 exchanges (10 messages total)...")
	builder.Ask(ctx, "My name is Alice.")
	builder.Ask(ctx, "I'm 25 years old.")
	builder.Ask(ctx, "I live in New York.")
	builder.Ask(ctx, "I work as a developer.")
	builder.Ask(ctx, "I love Go programming.")

	// Check history
	history := builder.GetHistory()
	fmt.Printf("History limited to %d messages (last 2 exchanges)\n", len(history))
	fmt.Println("Recent messages:")
	for i, msg := range history {
		preview := msg.Content
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		fmt.Printf("  %d. [%s]: %s\n", i+1, msg.Role, preview)
	}
	fmt.Println()

	// AI only remembers recent context
	response, _ := builder.Ask(ctx, "What do you know about me?")
	fmt.Printf("User: What do you know about me?\n")
	fmt.Printf("AI: %s\n", response)
	fmt.Printf("(AI only remembers recent messages)\n\n")

	fmt.Println()
}

// Example 5: Save and restore session - persist conversation
func example5_SaveAndRestoreSession(apiKey string) {
	fmt.Println("--- Example 5: Save and Restore Session ---")

	// Session 1: Have a conversation
	builder1 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a travel advisor.").
		WithMemory()

	ctx := context.Background()

	fmt.Println("Session 1:")
	builder1.Ask(ctx, "I want to visit Japan next spring.")
	builder1.Ask(ctx, "I'm interested in cherry blossoms and temples.")

	// Save session
	savedHistory := builder1.GetHistory()
	savedSystem := "You are a travel advisor." // In real app, you'd save this too
	fmt.Printf("Saved session with %d messages\n\n", len(savedHistory))

	// Session 2: Restore conversation in a new builder
	builder2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem(savedSystem).
		SetHistory(savedHistory).
		WithMemory()

	fmt.Println("Session 2 (restored):")
	response, _ := builder2.Ask(ctx, "When is the best time to see cherry blossoms?")
	fmt.Printf("User: When is the best time to see cherry blossoms?\n")
	fmt.Printf("AI: %s\n", response)
	fmt.Printf("(AI remembers the context from saved session)\n\n")

	fmt.Println()
}

// Example 6: Memory vs No Memory - compare behavior
func example6_MemoryVsNoMemory(apiKey string) {
	fmt.Println("--- Example 6: Memory vs No Memory ---")

	ctx := context.Background()

	// Without memory
	fmt.Println("WITHOUT Memory:")
	builderNoMem := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant.")
	// No WithMemory() call

	builderNoMem.Ask(ctx, "My name is Bob.")
	response, _ := builderNoMem.Ask(ctx, "What's my name?")
	fmt.Printf("User: What's my name?\n")
	fmt.Printf("AI: %s\n", response)
	fmt.Printf("(AI doesn't remember without memory)\n\n")

	// With memory
	fmt.Println("WITH Memory:")
	builderWithMem := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant.").
		WithMemory()

	builderWithMem.Ask(ctx, "My name is Bob.")
	response, _ = builderWithMem.Ask(ctx, "What's my name?")
	fmt.Printf("User: What's my name?\n")
	fmt.Printf("AI: %s\n", response)
	fmt.Printf("(AI remembers with memory enabled)\n\n")

	fmt.Println()
}
