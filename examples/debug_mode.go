package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/taipm/go-deep-agent/agent"
)

// This example demonstrates the enhanced debug mode in go-deep-agent.
// Debug mode provides comprehensive logging of:
// - Requests/responses with secret redaction
// - Errors with full context
// - Token usage (verbose mode)
// - Tool execution (verbose mode)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         Enhanced Debug Mode Examples                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Example 1: Basic Debug Mode
	example1BasicDebug(ctx, apiKey)

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Example 2: Verbose Debug Mode
	example2VerboseDebug(ctx, apiKey)

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Example 3: Custom Debug Configuration
	example3CustomDebug(ctx, apiKey)

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Example 4: Debug Mode with Tools
	example4DebugWithTools(ctx, apiKey)

	fmt.Println()
	fmt.Println("✅ All debug examples completed!")
}

// Example 1: Basic debug mode - logs requests, responses, and errors
func example1BasicDebug(ctx context.Context, apiKey string) {
	fmt.Println("Example 1: Basic Debug Mode")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Basic debug mode logs:")
	fmt.Println("  • Request details")
	fmt.Println("  • Response details")
	fmt.Println("  • Errors with context")
	fmt.Println("  • Secret redaction (API keys masked)")
	fmt.Println()

	// Create agent with basic debug mode
	ag := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDebug(agent.DefaultDebugConfig()).
		WithSystem("You are a helpful assistant.")

	fmt.Println("Making request with debug logging enabled...")
	fmt.Println()

	response, err := ag.Ask(ctx, "What is 2+2?")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println()
	fmt.Println("Response:", response)
}

// Example 2: Verbose debug mode - logs everything including tokens and tools
func example2VerboseDebug(ctx context.Context, apiKey string) {
	fmt.Println("Example 2: Verbose Debug Mode")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Verbose debug mode logs:")
	fmt.Println("  • Everything from basic mode")
	fmt.Println("  • Token usage (prompt, completion, total)")
	fmt.Println("  • Tool execution details")
	fmt.Println()

	// Create agent with verbose debug mode
	ag := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDebug(agent.VerboseDebugConfig()).
		WithSystem("You are a math assistant.")

	fmt.Println("Making request with verbose debug logging...")
	fmt.Println()

	response, err := ag.Ask(ctx, "What is the square root of 144?")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println()
	fmt.Println("Response:", response)
	fmt.Println()
	fmt.Printf("Token usage: %+v\n", ag.GetLastUsage())
}

// Example 3: Custom debug configuration
func example3CustomDebug(ctx context.Context, apiKey string) {
	fmt.Println("Example 3: Custom Debug Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Custom configuration allows:")
	fmt.Println("  • Selective logging (only what you need)")
	fmt.Println("  • Custom truncation limits")
	fmt.Println("  • Security controls")
	fmt.Println()

	// Create custom debug config
	customConfig := agent.DebugConfig{
		Enabled:           true,
		Level:             agent.DebugLevelBasic,
		RedactSecrets:     true, // Always true in production!
		LogRequests:       true,
		LogResponses:      false, // Don't log responses to reduce noise
		LogErrors:         true,
		LogTokenUsage:     true, // Enable token logging even at basic level
		LogToolExecutions: false,
		MaxLogLength:      1000, // Shorter logs
	}

	ag := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDebug(customConfig).
		WithSystem("You are a helpful assistant.")

	fmt.Println("Making request with custom debug config...")
	fmt.Println("(Note: Responses won't be logged)")
	fmt.Println()

	response, err := ag.Ask(ctx, "Explain Go interfaces in one sentence.")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println()
	fmt.Println("Response:", response)
}

// Example 4: Debug mode with tools - see tool execution logging
func example4DebugWithTools(ctx context.Context, apiKey string) {
	fmt.Println("Example 4: Debug Mode with Tools")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Debug mode with tools logs:")
	fmt.Println("  • Tool call arguments")
	fmt.Println("  • Tool execution results")
	fmt.Println("  • Tool errors with context")
	fmt.Println()

	// Define a calculator tool
	calculatorTool := agent.Tool{
		Type: agent.ToolTypeFunction,
		Function: agent.FunctionTool{
			Name:        "calculator",
			Description: "Performs basic arithmetic operations",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "The operation to perform: add, subtract, multiply, divide",
						"enum":        []string{"add", "subtract", "multiply", "divide"},
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "First number",
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Second number",
					},
				},
				"required": []string{"operation", "x", "y"},
			},
			Handler: func(args map[string]interface{}) (string, error) {
				op := args["operation"].(string)
				x := args["x"].(float64)
				y := args["y"].(float64)

				var result float64
				switch op {
				case "add":
					result = x + y
				case "subtract":
					result = x - y
				case "multiply":
					result = x * y
				case "divide":
					if y == 0 {
						return "", fmt.Errorf("division by zero")
					}
					result = x / y
				default:
					return "", fmt.Errorf("unknown operation: %s", op)
				}

				return fmt.Sprintf("%.2f", result), nil
			},
		},
	}

	// Create agent with verbose debug (to see tool execution)
	ag := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDebug(agent.VerboseDebugConfig()).
		WithSystem("You are a math assistant. Use the calculator tool for calculations.").
		WithTool(&calculatorTool).
		WithAutoExecute(true)

	fmt.Println("Asking agent to perform calculation (will use tool)...")
	fmt.Println()

	response, err := ag.Ask(ctx, "What is 156 multiplied by 23?")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println()
	fmt.Println("Response:", response)
}

// Production example: Minimal debug for production monitoring
func productionDebugConfig() agent.DebugConfig {
	return agent.DebugConfig{
		Enabled:           true,
		Level:             agent.DebugLevelBasic,
		RedactSecrets:     true,  // ALWAYS true in production
		LogRequests:       false, // Don't log requests in production
		LogResponses:      false, // Don't log responses in production
		LogErrors:         true,  // Always log errors
		LogTokenUsage:     true,  // Track token costs
		LogToolExecutions: false,
		MaxLogLength:      500, // Keep logs small
	}
}
