package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY required")
	}

	ctx := context.Background()

	// Example: French translation with few-shot examples
	fmt.Println("=== Few-Shot Translation Example ===")
	
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a French translator.").
		AddFewShotExample("Translate: Hello", "Bonjour").
		AddFewShotExample("Translate: Goodbye", "Au revoir").
		AddFewShotExample("Translate: Thank you", "Merci")

	response, err := ai.Ask(ctx, "Translate: Good morning")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Translation: %s\n", response)
}
