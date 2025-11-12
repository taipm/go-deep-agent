package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/taipm/go-deep-agent/agent"
	"github.com/taipm/go-deep-agent/agent/tools"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	ctx := context.Background()

	fmt.Println("=" + repeat("=", 78))
	fmt.Println("TOOL CHOICE CONTROL DEMO")
	fmt.Println("Demonstrates how to control when the LLM uses tools")
	fmt.Println("=" + repeat("=", 78) + "\n")

	// Create a calculator tool for demonstrations
	calculator := tools.NewMathTool()

	// ============================================================================
	// SCENARIO 1: AUTO MODE (Default Behavior)
	// ============================================================================
	fmt.Println("┌" + repeat("─", 78) + "┐")
	fmt.Println("│ SCENARIO 1: AUTO MODE (Default)")
	fmt.Println("│ The LLM decides whether to use tools or answer directly")
	fmt.Println("└" + repeat("─", 78) + "┘\n")

	builder1 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(calculator).
		WithAutoExecute(true).
		WithToolChoice("auto") // Optional - this is the default

	// Question that needs a tool
	fmt.Println("Question 1: Calculate 1234 * 5678")
	answer1, err := builder1.Ask(ctx, "Calculate 1234 * 5678")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Answer: %s\n\n", answer1)
	}

	// Question that doesn't need a tool
	fmt.Println("Question 2: What is the capital of France?")
	answer2, err := builder1.Ask(ctx, "What is the capital of France?")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Answer: %s\n\n", answer2)
	}

	// ============================================================================
	// SCENARIO 2: REQUIRED MODE (Force Tool Usage)
	// ============================================================================
	fmt.Println("┌" + repeat("─", 78) + "┐")
	fmt.Println("│ SCENARIO 2: REQUIRED MODE (Compliance & Audit)")
	fmt.Println("│ Force the LLM to call at least one tool - critical for:")
	fmt.Println("│   • Financial calculations (auditable, traceable)")
	fmt.Println("│   • Healthcare data (real-time, accurate)")
	fmt.Println("│   • Legal compliance (verified sources)")
	fmt.Println("└" + repeat("─", 78) + "┘\n")

	builder2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(calculator).
		WithAutoExecute(true).
		WithToolChoice("required") // Force tool usage

	// Financial calculation example
	fmt.Println("Use Case: Financial Compliance")
	fmt.Println("Question: Calculate the total value of 1000 shares at $750.50 each")
	answer3, err := builder2.Ask(ctx, "Calculate the total value of 1000 shares at $750.50 each")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Answer: %s\n", answer3)
		fmt.Println("✓ Calculation verified via tool - audit trail available")
		fmt.Println()
	}

	// What happens if we ask a non-calculation question?
	fmt.Println("Edge Case: Non-calculation question with required mode")
	fmt.Println("Question: What is the capital of France?")
	answer4, err := builder2.Ask(ctx, "What is the capital of France?")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Answer: %s\n", answer4)
		fmt.Println("Note: LLM will try to use tools even for non-math questions")
		fmt.Println()
	}

	// ============================================================================
	// SCENARIO 3: NONE MODE (Disable Tools)
	// ============================================================================
	fmt.Println("┌" + repeat("─", 78) + "┐")
	fmt.Println("│ SCENARIO 3: NONE MODE (Disable Tools)")
	fmt.Println("│ Prevent tool usage even when tools are configured")
	fmt.Println("│ Use cases:")
	fmt.Println("│   • Testing LLM reasoning without tools")
	fmt.Println("│   • Cost optimization (skip tool calls)")
	fmt.Println("│   • Safety checks before actual execution")
	fmt.Println("└" + repeat("─", 78) + "┘\n")

	builder3 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(calculator). // Tools configured but disabled
		WithToolChoice("none") // Disable tool calling

	fmt.Println("Question: Calculate 9 * 8")
	answer5, err := builder3.Ask(ctx, "Calculate 9 * 8")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Answer: %s\n", answer5)
		fmt.Println("Note: LLM answers directly without using the calculator tool")
		fmt.Println()
	}

	// ============================================================================
	// SCENARIO 4: ERROR HANDLING
	// ============================================================================
	fmt.Println("┌" + repeat("─", 78) + "┐")
	fmt.Println("│ SCENARIO 4: ERROR HANDLING")
	fmt.Println("│ What happens when toolChoice is misconfigured?")
	fmt.Println("└" + repeat("─", 78) + "┘\n")

	// Error 1: Invalid choice value
	fmt.Println("Error Case 1: Invalid toolChoice value")
	builder4 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithToolChoice("invalid_value")

	_, err = builder4.Ask(ctx, "test")
	if err != nil {
		fmt.Printf("✓ Caught error: %v\n\n", err)
	}

	// Error 2: toolChoice without tools
	fmt.Println("Error Case 2: toolChoice set but no tools configured")
	builder5 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithToolChoice("required") // No tools!

	_, err = builder5.Ask(ctx, "Calculate 5 + 5")
	if err != nil {
		fmt.Printf("✓ Caught error: %v\n\n", err)
	}

	// ============================================================================
	// SUMMARY
	// ============================================================================
	fmt.Println("┌" + repeat("─", 78) + "┐")
	fmt.Println("│ SUMMARY: When to use each mode")
	fmt.Println("├" + repeat("─", 78) + "┤")
	fmt.Println("│ AUTO (default):")
	fmt.Println("│   • General purpose - LLM decides when to use tools")
	fmt.Println("│   • Best for mixed conversations (some need tools, some don't)")
	fmt.Println("│")
	fmt.Println("│ REQUIRED:")
	fmt.Println("│   • Compliance & audit trails (financial, legal, healthcare)")
	fmt.Println("│   • Quality control - guarantee 100% accurate data")
	fmt.Println("│   • API integration - force real-time data retrieval")
	fmt.Println("│   • Security - mandatory verification steps")
	fmt.Println("│")
	fmt.Println("│ NONE:")
	fmt.Println("│   • Disable tools temporarily")
	fmt.Println("│   • Test LLM reasoning without external data")
	fmt.Println("│   • Cost optimization")
	fmt.Println("└" + repeat("─", 78) + "┘")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
