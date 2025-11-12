package main
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
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set OPENAI_API_KEY environment variable")
	}

	fmt.Println("🚀 Native ReAct Function Calling Demos")
	fmt.Println("=====================================\n")

	// Run demos
	fmt.Println("Demo 1: Simple calculation...")
	runSimpleCalculation(apiKey)

	fmt.Println("\nDemo 2: Multi-step reasoning...")
	runMultiStepReasoning(apiKey)

	fmt.Println("\nDemo 3: Without tools (pure reasoning)...")
	runPureReasoning(apiKey)

	fmt.Println("\n✅ All demos completed!")
}

// Demo 1: Simple calculation using math tool
func runSimpleCalculation(apiKey string) {
	ai := agent.New(agent.ProviderOpenAI, "gpt-4o-mini").
		WithAPIKey(apiKey).
		WithReActMode(true).
		WithReActNativeMode(). // Use native function calling
		WithTools(tools.NewMathTool())

	result, err := ai.Execute(context.Background(), "What is 25 * 17 + 123?")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Answer: %s\n", result.Answer)
	fmt.Printf("📊 Steps: %d, Tool calls: %d\n", 
		len(result.Steps), countToolCalls(result))
}

// Demo 2: Multi-step reasoning with multiple tools
func runMultiStepReasoning(apiKey string) {
	ai := agent.New(agent.ProviderOpenAI, "gpt-4o-mini").
		WithAPIKey(apiKey).
		WithReActMode(true).
		WithReActNativeMode().
		WithTools(
			tools.NewMathTool(),
			tools.NewDateTimeTool(),
		)

	result, err := ai.Execute(context.Background(), 
		"I was born on 1990-05-15. How many days old am I today? Then calculate what 10% of that number would be.")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Answer: %s\n", result.Answer)
	fmt.Printf("📊 Steps: %d, Tool calls: %d\n", 
		len(result.Steps), countToolCalls(result))
		
	// Show reasoning steps
	fmt.Println("🧠 Reasoning steps:")
	for i, step := range result.Steps {
		fmt.Printf("  %d. [%s] %s\n", i+1, step.Type, 
			truncateString(step.Content, 60))
	}
}

// Demo 3: Pure reasoning without tools
func runPureReasoning(apiKey string) {
	ai := agent.New(agent.ProviderOpenAI, "gpt-4o-mini").
		WithAPIKey(apiKey).
		WithReActMode(true).
		WithReActNativeMode()
		// No tools registered

	result, err := ai.Execute(context.Background(), 
		"Explain the concept of compound interest in simple terms with an example.")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Answer: %s\n", result.Answer)
	fmt.Printf("📊 Steps: %d (pure reasoning)\n", len(result.Steps))
}

// Helper functions
func countToolCalls(result *agent.ReActResult) int {
	count := 0
	for _, step := range result.Steps {
		if step.Type == "ACTION" {
			count++
		}
	}
	return count
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}