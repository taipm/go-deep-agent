package plannerbasic
// Package main demonstrates basic planning capabilities of go-deep-agent.
// This example shows how to use PlanAndExecute to decompose and execute complex goals.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create agent configuration
	config := agent.Config{
		Provider: "openai",
		Model:    "gpt-4",
		APIKey:   apiKey,
	}

	// Create agent
	ag, err := agent.NewAgent(config)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Define a goal that requires planning
	goal := "Research the top 3 programming trends in 2024 and summarize each in one sentence"

	fmt.Printf("🎯 Goal: %s\n\n", goal)
	fmt.Println("🔄 Planning and executing...")

	// Execute with planning
	ctx := context.Background()
	result, err := ag.PlanAndExecute(ctx, goal)
	if err != nil {
		log.Fatalf("Execution failed: %v", err)
	}

	// Display results
	fmt.Println("\n✅ Execution Complete!")
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Tasks in Plan: %d\n", result.Metrics.TaskCount)
	fmt.Printf("Success Rate: %.1f%%\n", result.Metrics.SuccessRate*100)

	// Show plan structure
	fmt.Println("\n📋 Plan Structure:")
	planJSON, _ := json.MarshalIndent(result.Plan, "", "  ")
	fmt.Println(string(planJSON))

	// Show final result
	if result.FinalResult != nil {
		fmt.Printf("\n🎉 Result: %v\n", result.FinalResult)
	}
}
