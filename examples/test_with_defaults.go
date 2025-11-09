package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
	"github.com/taipm/go-deep-agent/agent/tools"
)

func main() {
	fmt.Println("=== Test tools.WithDefaults() - v0.5.5 ===\n")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	ctx := context.Background()

	// Test 1: WithDefaults() - DateTime + Math auto-loaded
	fmt.Println("--- Test 1: WithDefaults() (DateTime + Math) ---")
	ai1 := tools.WithDefaults(agent.NewOpenAI("gpt-4o-mini", apiKey)).
		WithAutoExecute(true)

	response1, err := ai1.Ask(ctx, "What day of the week is Christmas 2025 and what is 2 * (3 + 4)?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response1)
	}

	// Test 2: WithAll() - All built-in tools
	fmt.Println("--- Test 2: WithAll() (FileSystem + HTTP + DateTime + Math) ---")
	ai2 := tools.WithAll(agent.NewOpenAI("gpt-4o-mini", apiKey)).
		WithAutoExecute(true)

	response2, err := ai2.Ask(ctx, "Get current UTC time and calculate 2^10")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response2)
	}

	// Test 3: Manual selection - Only Math
	fmt.Println("--- Test 3: Manual selection (Math only) ---")
	ai3 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(tools.NewMathTool()).
		WithAutoExecute(true)

	response3, err := ai3.Ask(ctx, "Calculate the mean of: 10, 20, 30, 40, 50")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response3)
	}

	// Test 4: No tools - Pure chatbot
	fmt.Println("--- Test 4: No tools (Pure chatbot) ---")
	ai4 := agent.NewOpenAI("gpt-4o-mini", apiKey)

	response4, err := ai4.Ask(ctx, "Hello, how are you?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response4)
	}

	fmt.Println("=== All tests completed successfully! ===")
}
