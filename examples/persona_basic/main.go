package personabasic
package main

import (
	"context"
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

	// Load persona from YAML file
	persona, err := agent.LoadPersona("../../personas/customer_support.yaml")
	if err != nil {
		log.Fatal("Failed to load persona:", err)
	}

	fmt.Println("📄 Loaded Persona:")
	fmt.Printf("  Name: %s (v%s)\n", persona.Name, persona.Version)
	fmt.Printf("  Role: %s\n", persona.Role)
	fmt.Printf("  Goal: %s\n", persona.Goal)
	fmt.Printf("  Personality Tone: %s\n", persona.Personality.Tone)
	fmt.Println()

	// Create agent with persona
	supportAgent := agent.NewOpenAI("", apiKey).
		WithPersona(persona)

	// Use the agent - it will behave according to persona
	ctx := context.Background()

	fmt.Println("🤖 Customer Support Agent (using persona):")
	fmt.Println("=" + string(make([]byte, 50)) + "=")
	fmt.Println()

	// Example 1: Customer with login issue
	fmt.Println("👤 Customer: I can't log in to my account!")
	fmt.Println()
	response1, err := supportAgent.Ask(ctx, "I can't log in to my account!")
	if err != nil {
		log.Fatal("Agent error:", err)
	}
	fmt.Printf("🎧 Support Agent: %s\n", response1)
	fmt.Println()
	fmt.Println(string(make([]byte, 70)))
	fmt.Println()

	// Example 2: Customer with late order
	fmt.Println("👤 Customer: My order is 3 days late, where is it?")
	fmt.Println()
	response2, err := supportAgent.Ask(ctx, "My order is 3 days late, where is it?")
	if err != nil {
		log.Fatal("Agent error:", err)
	}
	fmt.Printf("🎧 Support Agent: %s\n", response2)
	fmt.Println()

	// Show different personas
	fmt.Println()
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("📚 Trying different personas:")
	fmt.Println()

	// Load and show code reviewer persona
	codeReviewerPersona, err := agent.LoadPersona("../../personas/code_reviewer.yaml")
	if err == nil {
		fmt.Printf("✅ %s - %s\n", codeReviewerPersona.Name, codeReviewerPersona.Role)
	}

	// Load and show technical writer persona
	writerPersona, err := agent.LoadPersona("../../personas/technical_writer.yaml")
	if err == nil {
		fmt.Printf("✅ %s - %s\n", writerPersona.Name, writerPersona.Role)
	}

	// Load all personas from directory
	fmt.Println()
	personas, err := agent.LoadPersonasFromDirectory("../../personas")
	if err == nil {
		fmt.Printf("📦 Total personas available: %d\n", len(personas))
		for name := range personas {
			fmt.Printf("   - %s\n", name)
		}
	}

	fmt.Println()
	fmt.Println("✅ Demo complete!")
}
