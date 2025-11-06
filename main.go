package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	ctx := context.Background()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Example 1: Simple Chat with Builder API
	fmt.Println("=== Example 1: Simple Chat ===")
	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		Ask(ctx, "What is the capital of Vietnam?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 2: Chat with System Prompt and Temperature
	fmt.Println("=== Example 2: System Prompt & Temperature ===")
	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant that explains concepts in simple terms.").
		WithTemperature(0.7)

	response, err = builder.Ask(ctx, "Explain quantum computing")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 3: Streaming Response
	fmt.Println("=== Example 3: Streaming ===")
	fmt.Print("Streaming: ")
	response, err = agent.NewOpenAI("gpt-4o-mini", apiKey).
		OnStream(func(content string) {
			fmt.Print(content)
		}).
		Stream(ctx, "Write a haiku about AI")
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("\nComplete response: %s\n\n", response)

	// Example 4: Conversation with Memory
	fmt.Println("=== Example 4: Conversation Memory ===")
	builder = agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory()

	// First message
	response, err = builder.Ask(ctx, "My name is John and I'm from Vietnam")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	// Follow-up that requires memory
	response, err = builder.Ask(ctx, "What's my name and where am I from?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Follow-up Response: %s\n\n", response)
	}

	// Example 5: Tool Calling with Auto-Execution
	fmt.Println("=== Example 5: Tool Calling ===")
	weatherTool := agent.NewTool("get_weather", "Get the current weather for a location").
		AddParameter("location", "string", "The city name", true).
		AddParameter("units", "string", "Temperature units (celsius/fahrenheit)", false).
		WithHandler(func(args string) (string, error) {
			return fmt.Sprintf("The weather in Hanoi is sunny, 25°C"), nil
		})

	builder = agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(weatherTool).
		WithAutoExecute(true)

	response, err = builder.Ask(ctx, "What's the weather like in Hanoi?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 6: Structured Outputs with JSON Schema
	fmt.Println("=== Example 6: Structured Outputs ===")
	personSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The person's name",
			},
			"age": map[string]interface{}{
				"type":        "integer",
				"description": "The person's age",
			},
		},
		"required":             []string{"name", "age"},
		"additionalProperties": false,
	}

	response, err = agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithJSONSchema("person_info", "Extract person information", personSchema, true).
		Ask(ctx, "Extract info: John is 25 years old")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Structured Response: %s\n\n", response)
	}

	// Example 7: Error Handling with Timeout and Retry
	fmt.Println("=== Example 7: Error Handling ===")
	response, err = agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTimeout(10*time.Second).
		WithRetry(3).
		WithExponentialBackoff().
		Ask(ctx, "What is machine learning?")
	if err != nil {
		// Check error type
		if agent.IsTimeoutError(err) {
			log.Printf("Request timed out: %v", err)
		} else if agent.IsRateLimitError(err) {
			log.Printf("Rate limit exceeded: %v", err)
		} else {
			log.Printf("Error: %v", err)
		}
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 8: Production-Ready Configuration
	fmt.Println("=== Example 8: Production Configuration ===")
	response, err = agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant").
		WithTemperature(0.7).
		WithMaxTokens(500).
		WithMemory().
		WithMaxHistory(10).
		WithTimeout(30*time.Second).
		WithRetry(3).
		WithExponentialBackoff().
		Ask(ctx, "Explain the benefits of Go programming language")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 9: Managing Conversation History
	fmt.Println("=== Example 9: History Management ===")
	builder = agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory()

	// Have a conversation
	builder.Ask(ctx, "I love programming in Go")
	builder.Ask(ctx, "What are goroutines?")

	// Get conversation history
	history := builder.GetHistory()
	fmt.Printf("Conversation has %d messages\n", len(history))

	// Clear conversation but keep system prompt
	builder.Clear()
	fmt.Printf("After clear: %d messages (system prompt preserved)\n", len(builder.GetHistory()))

	fmt.Println("\n=== All Examples Complete ===")
}
