package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

// Production Agent: WithDefaults() + customization
//
// This example shows how to:
//  1. Start with production-ready defaults via WithDefaults()
//  2. Customize specific settings via method chaining
//  3. Add optional features (tools, logging, etc.)
//
// Philosophy: Start with smart defaults, then customize what you need.

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Start with defaults, then customize
	ai := agent.NewOpenAI("gpt-4", apiKey).
		WithDefaults().                                    // Production-ready: Memory(20), Retry(3), Timeout(30s), ExponentialBackoff
		WithMaxHistory(50).                                // Customize: Increase memory to 50 messages
		WithTemperature(0.7).                              // Add: Creative temperature
		WithMaxTokens(500).                                // Add: Limit response length
		WithSystem("You are a helpful, concise assistant") // Add: System prompt

	ctx := context.Background()

	// Example 1: Simple conversation with customized settings
	fmt.Println("=== Example 1: Customized Defaults ===")
	resp, err := ai.Ask(ctx, "Explain quantum computing in 2 sentences")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("AI: %s\n\n", resp)

	// Example 2: Override defaults for specific needs
	fmt.Println("=== Example 2: Opt-out of Memory ===")
	aiNoMemory := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDefaults().
		DisableMemory() // Remove memory for stateless interactions

	resp2, err := aiNoMemory.Ask(ctx, "What is 2+2?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("AI (no memory): %s\n\n", resp2)

	// Example 3: Progressive enhancement - add tools
	fmt.Println("=== Example 3: Add Tools to Defaults ===")

	// Define a simple calculator tool
	calculator := &agent.Tool{
		Name:        "calculator",
		Description: "Performs basic arithmetic operations",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": map[string]interface{}{
					"type":        "number",
					"description": "First number",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "Second number",
				},
			},
			"required": []string{"operation", "a", "b"},
		},
		Function: func(args map[string]interface{}) (string, error) {
			op := args["operation"].(string)
			a := args["a"].(float64)
			b := args["b"].(float64)

			var result float64
			switch op {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				if b == 0 {
					return "", fmt.Errorf("division by zero")
				}
				result = a / b
			default:
				return "", fmt.Errorf("unknown operation: %s", op)
			}

			return fmt.Sprintf("%.2f", result), nil
		},
	}

	aiWithTools := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDefaults().        // Start with defaults
		WithTools(calculator). // Add tool capability
		WithAutoExecute(true)  // Auto-execute tool calls

	resp3, err := aiWithTools.Ask(ctx, "What is 15.5 multiplied by 3?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("AI (with tools): %s\n\n", resp3)

	// Example 4: Compare Bare vs WithDefaults()
	fmt.Println("=== Example 4: Bare vs WithDefaults() ===")

	// Bare: Full control, no defaults
	aiBare := agent.NewOpenAI("gpt-4o-mini", apiKey)
	fmt.Printf("Bare - MaxHistory: %d, MaxRetries: %d, Timeout: %v\n",
		aiBare.GetMaxHistory(), aiBare.GetMaxRetries(), aiBare.GetTimeout())

	// WithDefaults: Production-ready out-of-the-box
	aiDefaults := agent.NewOpenAI("gpt-4o-mini", apiKey).WithDefaults()
	fmt.Printf("WithDefaults - MaxHistory: %d, MaxRetries: %d, Timeout: %v\n",
		aiDefaults.GetMaxHistory(), aiDefaults.GetMaxRetries(), aiDefaults.GetTimeout())
}

// Note: GetMaxHistory(), GetMaxRetries(), GetTimeout() are placeholder methods
// shown for demonstration. In reality, you would inspect the builder fields
// or observe the behavior during execution.
