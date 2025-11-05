package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Builder API Advanced Parameters ===\n")

	// Example 1: Temperature Control
	example1_Temperature()

	// Example 2: Max Tokens
	example2_MaxTokens()

	// Example 3: Reduce Repetition
	example3_ReduceRepetition()

	// Example 4: Reproducible Outputs
	example4_Reproducible()

	// Example 5: Multiple Choices
	example5_MultipleChoices()

	// Example 6: Combined Parameters
	example6_Combined()
}

// Example 1: Temperature control for creativity vs precision
func example1_Temperature() {
	fmt.Println("--- Example 1: Temperature Control ---")

	ctx := context.Background()

	// Low temperature (0.2) = Focused, deterministic
	lowTemp, err := agent.NewOllama("qwen3:1.7b").
		WithTemperature(0.2).
		Ask(ctx, "Write a one-sentence fact about Paris.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Low Temperature (0.2) - Focused:")
	fmt.Printf("%s\n\n", lowTemp)

	// High temperature (1.5) = Creative, diverse
	highTemp, err := agent.NewOllama("qwen3:1.7b").
		WithTemperature(1.5).
		Ask(ctx, "Write a one-sentence fact about Paris.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("High Temperature (1.5) - Creative:")
	fmt.Printf("%s\n\n", highTemp)
}

// Example 2: Limiting response length with MaxTokens
func example2_MaxTokens() {
	fmt.Println("--- Example 2: Max Tokens Limit ---")

	ctx := context.Background()

	// Very short response (20 tokens max)
	short, err := agent.NewOllama("qwen3:1.7b").
		WithMaxTokens(20).
		Ask(ctx, "Explain machine learning in detail.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Short (20 tokens max):")
	fmt.Printf("%s\n\n", short)

	// Longer response (100 tokens max)
	long, err := agent.NewOllama("qwen3:1.7b").
		WithMaxTokens(100).
		Ask(ctx, "Explain machine learning in detail.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Longer (100 tokens max):")
	fmt.Printf("%s\n\n", long)
}

// Example 3: Reduce repetition with penalties
func example3_ReduceRepetition() {
	fmt.Println("--- Example 3: Reduce Repetition ---")

	ctx := context.Background()

	// Without penalties
	normal, err := agent.NewOllama("qwen3:1.7b").
		Ask(ctx, "List 5 fruits.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Normal (no penalties):")
	fmt.Printf("%s\n\n", normal)

	// With frequency penalty to reduce repetition
	diverse, err := agent.NewOllama("qwen3:1.7b").
		WithFrequencyPenalty(0.8).
		WithPresencePenalty(0.6).
		Ask(ctx, "List 5 fruits.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("With Penalties (more diverse):")
	fmt.Printf("%s\n\n", diverse)
}

// Example 4: Reproducible outputs with seed
func example4_Reproducible() {
	fmt.Println("--- Example 4: Reproducible Outputs ---")

	ctx := context.Background()

	// Same seed should give same result
	seed := int64(42)

	result1, err := agent.NewOllama("qwen3:1.7b").
		WithSeed(seed).
		WithTemperature(0.7).
		Ask(ctx, "Generate a random number and explain why.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	result2, err := agent.NewOllama("qwen3:1.7b").
		WithSeed(seed).
		WithTemperature(0.7).
		Ask(ctx, "Generate a random number and explain why.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("First call (seed=42):")
	fmt.Printf("%s\n\n", result1)

	fmt.Println("Second call (seed=42) - Should be similar:")
	fmt.Printf("%s\n\n", result2)
}

// Example 5: Multiple choices (N-best responses)
func example5_MultipleChoices() {
	fmt.Println("--- Example 5: Multiple Choices ---")

	ctx := context.Background()

	// Generate 3 different creative responses
	choices, err := agent.NewOllama("qwen3:1.7b").
		WithMultipleChoices(3).
		WithTemperature(0.8).
		AskMultiple(ctx, "Write a one-line joke about programming.")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Generated 3 different jokes:")
	for i, choice := range choices {
		fmt.Printf("%d. %s\n", i+1, choice)
	}
	fmt.Println()
}

// Example 6: Combining multiple parameters
func example6_Combined() {
	fmt.Println("--- Example 6: Combined Parameters ---")

	ctx := context.Background()

	// Precise, short, non-repetitive response
	result, err := agent.NewOllama("qwen3:1.7b").
		WithTemperature(0.3).      // Focused
		WithMaxTokens(50).         // Short
		WithFrequencyPenalty(0.5). // Less repetition
		WithSystem("You are a concise encyclopedia.").
		Ask(ctx, "What is quantum computing?")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Parameters:")
	fmt.Println("  - Temperature: 0.3 (focused)")
	fmt.Println("  - MaxTokens: 50 (short)")
	fmt.Println("  - FrequencyPenalty: 0.5 (less repetition)")
	fmt.Println("  - System: Concise encyclopedia")
	fmt.Println("\nResult:")
	fmt.Printf("%s\n", result)
}
