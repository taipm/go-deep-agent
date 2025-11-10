package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Load configuration from YAML file
	config, err := agent.LoadAgentConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	fmt.Println("📄 Loaded configuration:")
	fmt.Printf("  Model: %s\n", config.Model)
	fmt.Printf("  Temperature: %.1f\n", config.Temperature)
	fmt.Printf("  Max Tokens: %d\n", config.MaxTokens)
	fmt.Printf("  Memory Capacity: %d\n", config.Memory.WorkingCapacity)
	fmt.Println()

	// Create agent with configuration
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	myAgent := agent.NewOpenAI("", apiKey).
		WithAgentConfig(config)

	// Use the agent
	ctx := context.Background()
	fmt.Println("🤖 Asking agent a question...")
	response, err := myAgent.Ask(ctx, "What is the capital of France?")
	if err != nil {
		log.Fatal("Agent error:", err)
	}

	fmt.Printf("\n✅ Response: %s\n\n", response)

	// Export current configuration
	fmt.Println("💾 Exporting current configuration...")
	currentConfig := myAgent.ToAgentConfig()
	if err := agent.SaveAgentConfig(currentConfig, "exported_config.yaml"); err != nil {
		log.Fatal("Failed to save config:", err)
	}

	fmt.Println("✅ Configuration exported to exported_config.yaml")
}
