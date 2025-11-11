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
	fmt.Println("=== ReAct with Built-in MathTool ===")
	fmt.Println("Demonstrates ReAct pattern with professional math operations\n")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable required")
	}

	ctx := context.Background()

	// Example 1: Simple expression evaluation
	example1SimpleCalculation(ctx, apiKey)

	// Example 2: Statistics operations
	example2Statistics(ctx, apiKey)

	// Example 3: Complex multi-step reasoning
	example3ComplexReasoning(ctx, apiKey)

	// Example 4: Unit conversions
	example4UnitConversion(ctx, apiKey)

	// Example 5: Full reasoning trace
	example5FullTrace(ctx, apiKey)

	fmt.Println("\n✅ All ReAct + MathTool examples completed!")
}

// Example 1: Simple expression evaluation
func example1SimpleCalculation(ctx context.Context, apiKey string) {
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("Example 1: Simple Expression Evaluation")
	fmt.Println("─────────────────────────────────────────────────────\n")

	// Create agent with ReAct mode + built-in MathTool
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(3).
		WithTool(tools.NewMathTool()). // Built-in professional math tool
		WithSystem("You are a helpful math assistant. Use the math tool to solve calculations.")

	task := "What is the result of 2 * (15 + 8) - sqrt(16)?"

	fmt.Printf("Task: %s\n\n", task)

	result, err := ai.Execute(ctx, task)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if result.Success {
		fmt.Printf("✅ Answer: %s\n", result.Answer)
		fmt.Printf("📊 Iterations: %d\n\n", result.Iterations)
	}
}

// Example 2: Statistics operations
func example2Statistics(ctx context.Context, apiKey string) {
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("Example 2: Statistical Analysis")
	fmt.Println("─────────────────────────────────────────────────────\n")

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(5).
		WithTool(tools.NewMathTool()).
		WithSystem("You are a data analyst. Use the math tool for statistical calculations.")

	task := "Calculate the mean, median, and standard deviation of: 85, 90, 78, 92, 88, 95, 82"

	fmt.Printf("Task: %s\n\n", task)

	result, err := ai.Execute(ctx, task)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if result.Success {
		fmt.Printf("✅ Answer: %s\n", result.Answer)
		fmt.Printf("📊 Tool calls: %d\n", countToolCalls(result))
		fmt.Printf("📊 Iterations: %d\n\n", result.Iterations)
	}
}

// Example 3: Complex multi-step reasoning
func example3ComplexReasoning(ctx context.Context, apiKey string) {
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("Example 3: Complex Multi-Step Reasoning")
	fmt.Println("─────────────────────────────────────────────────────\n")

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(8).
		WithTool(tools.NewMathTool()).
		WithSystem("You are a problem solver. Break down complex problems step by step.")

	task := `A student scored 75, 82, 90, 88, and 85 on five tests. 
	The final exam is worth 40% of the grade, and the average of the five tests is worth 60%.
	If the student needs a 85% overall to get an A, what minimum score do they need on the final exam?`

	fmt.Printf("Task: %s\n\n", task)

	result, err := ai.Execute(ctx, task)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if result.Success {
		fmt.Printf("✅ Answer: %s\n", result.Answer)
		fmt.Printf("📊 Steps taken: %d\n", len(result.Steps))
		fmt.Printf("📊 Tool calls: %d\n", countToolCalls(result))
		fmt.Printf("📊 Iterations: %d\n\n", result.Iterations)
	}
}

// Example 4: Unit conversions
func example4UnitConversion(ctx context.Context, apiKey string) {
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("Example 4: Unit Conversion")
	fmt.Println("─────────────────────────────────────────────────────\n")

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(5).
		WithTool(tools.NewMathTool()).
		WithSystem("You are a unit conversion expert. Use the math tool for conversions.")

	task := "If I run 5 kilometers, how many meters is that? Also convert 100 celsius to fahrenheit."

	fmt.Printf("Task: %s\n\n", task)

	result, err := ai.Execute(ctx, task)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if result.Success {
		fmt.Printf("✅ Answer: %s\n", result.Answer)
		fmt.Printf("📊 Iterations: %d\n\n", result.Iterations)
	}
}

// Example 5: Full reasoning trace with step-by-step output
func example5FullTrace(ctx context.Context, apiKey string) {
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("Example 5: Full Reasoning Trace")
	fmt.Println("─────────────────────────────────────────────────────\n")

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(5).
		WithTool(tools.NewMathTool()).
		WithSystem("You are a step-by-step problem solver. Show your reasoning clearly.")

	task := "What is 20% of 350, and then add 50 to that result?"

	fmt.Printf("Task: %s\n\n", task)

	result, err := ai.Execute(ctx, task)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	// Print full trace
	fmt.Println("🔍 Reasoning Trace:")
	fmt.Println("════════════════════════════════════════════════════")
	for i, step := range result.Steps {
		switch step.Type {
		case "thought":
			fmt.Printf("\n💭 THOUGHT #%d:\n   %s\n", i+1, step.Content)
		case "action":
			fmt.Printf("\n⚡ ACTION #%d:\n   Tool: %s\n   Args: %v\n", i+1, step.Tool, step.Args)
		case "observation":
			fmt.Printf("\n👁️  OBSERVATION #%d:\n   %s\n", i+1, step.Content)
		case "final":
			fmt.Printf("\n✅ FINAL ANSWER:\n   %s\n", step.Content)
		}
	}
	fmt.Println("\n════════════════════════════════════════════════════")

	if result.Success {
		fmt.Printf("\n📊 Summary:\n")
		fmt.Printf("   Total steps: %d\n", len(result.Steps))
		fmt.Printf("   Tool calls: %d\n", countToolCalls(result))
		fmt.Printf("   Iterations: %d\n\n", result.Iterations)
	}
}

// Helper function to count tool calls in ReAct result
func countToolCalls(result *agent.ReActResult) int {
	count := 0
	for _, step := range result.Steps {
		if step.Type == "action" {
			count++
		}
	}
	return count
}
