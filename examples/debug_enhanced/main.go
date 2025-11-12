package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
	"github.com/taipm/go-deep-agent/agent/tools"
)

// Example: Enhanced Debug Output (v0.7.7)
//
// This example demonstrates the tree-style debug logging introduced in v0.7.7.
// Debug output shows ReAct iterations with beautiful tree formatting:
//
//   [DEBUG] ReAct Iteration 1/5
//   [DEBUG] ├─ THOUGHT: I need to calculate 500000 + (500000 * 0.5)
//   [DEBUG] ├─ ACTION: math(operation="evaluate", expression="500000+(500000*0.5)")
//   [DEBUG] ├─ OBSERVATION: 750000.000000
//   [DEBUG] └─ Duration: 1.2s
//
// Enable debug logging with .WithDebug() or configure verbosity with .WithDebugConfig()

func main() {
	fmt.Println("=== Enhanced Debug Output (v0.7.7) ===")
	fmt.Println("Tree-style logging for ReAct iterations\n")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable required")
	}

	ctx := context.Background()

	// Example 1: Basic debug mode
	example1BasicDebug(ctx, apiKey)

	// Example 2: Verbose debug with tool executions
	example2VerboseDebug(ctx, apiKey)

	fmt.Println("\n✅ Debug examples completed!")
}

// Example 1: Basic debug mode
func example1BasicDebug(ctx context.Context, apiKey string) {
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("Example 1: Basic Debug Mode")
	fmt.Println("─────────────────────────────────────────────────────\n")

	// Enable debug logging - shows ReAct iterations with tree structure
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDebugLogging(). // Enable debug logging (v0.7.7: tree-style output)
		WithReActMode(true).
		WithReActComplexity(agent.ReActTaskSimple). // 3 iterations
		WithTool(tools.NewMathTool()).
		WithSystem("You are a helpful math assistant.")

	task := "Calculate: If a company had $500,000 revenue and grew by 50%, what's the new revenue?"

	fmt.Printf("Task: %s\n\n", task)
	fmt.Println("Debug output will show each ReAct iteration:\n")

	result, err := ai.Execute(ctx, task)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if result.Success {
		fmt.Printf("\n✅ Final Answer: %s\n", result.Answer)
		fmt.Printf("📊 Completed in %d iterations\n\n", result.Iterations)
	}
}

// Example 2: Verbose debug with tool execution details
func example2VerboseDebug(ctx context.Context, apiKey string) {
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("Example 2: Verbose Debug Mode")
	fmt.Println("─────────────────────────────────────────────────────\n")

	// Verbose debug shows tool execution details
	debugConfig := agent.VerboseDebugConfig() // LogToolExecutions = true

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDebug(debugConfig). // Verbose mode with tool details
		WithReActMode(true).
		WithReActComplexity(agent.ReActTaskMedium). // 5 iterations
		WithTool(tools.NewMathTool()).
		WithSystem("You are a data analyst.")

	task := "Calculate the average of these sales figures: $12,500, $18,300, $15,700"

	fmt.Printf("Task: %s\n\n", task)
	fmt.Println("Verbose debug shows tool inputs/outputs:\n")

	result, err := ai.Execute(ctx, task)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if result.Success {
		fmt.Printf("\n✅ Final Answer: %s\n", result.Answer)
		fmt.Printf("📊 Completed in %d iterations\n", result.Iterations)

		// Show the reasoning steps
		fmt.Println("\n📝 Reasoning steps:")
		for i, step := range result.Steps {
			fmt.Printf("  %d. [%s] %s\n", i+1, step.Type, truncate(step.Content, 80))
		}
		fmt.Println()
	}
}

// Helper: Truncate string for display
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
