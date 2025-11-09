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
	fmt.Println("=== Built-in Tools Demo (v0.5.3) ===\n")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Note: Comment out examples as needed
	// exampleFileSystemTool(apiKey)
	// exampleHTTPRequestTool(apiKey)
	// exampleDateTimeTool(apiKey)
	exampleCombinedTools(apiKey)
}

func exampleFileSystemTool(apiKey string) {
	fmt.Println("--- Example: FileSystem Tool ---")
	ctx := context.Background()

	fsTool := tools.NewFileSystemTool()
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(fsTool).
		WithAutoExecute(true)

	query := "Create a file called test.txt with content 'Hello AI'"
	fmt.Printf("\nQuery: %s\n", query)

	response, err := ai.Ask(ctx, query)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Response: %s\n\n", response)
	os.Remove("test.txt")
}

func exampleHTTPRequestTool(apiKey string) {
	fmt.Println("--- Example: HTTP Request Tool ---")
	ctx := context.Background()

	httpTool := tools.NewHTTPRequestTool()
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(httpTool).
		WithAutoExecute(true)

	query := "Get data from https://jsonplaceholder.typicode.com/posts/1"
	fmt.Printf("\nQuery: %s\n", query)

	response, err := ai.Ask(ctx, query)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Response: %s\n\n", response)
}

func exampleDateTimeTool(apiKey string) {
	fmt.Println("--- Example: DateTime Tool ---")
	ctx := context.Background()

	dtTool := tools.NewDateTimeTool()
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(dtTool).
		WithAutoExecute(true)

	query := "What day of the week is Christmas 2025?"
	fmt.Printf("\nQuery: %s\n", query)

	response, err := ai.Ask(ctx, query)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Response: %s\n\n", response)
}

func exampleCombinedTools(apiKey string) {
	fmt.Println("--- Example: Combined Tools ---")
	ctx := context.Background()

	fsTool := tools.NewFileSystemTool()
	httpTool := tools.NewHTTPRequestTool()
	dtTool := tools.NewDateTimeTool()

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(fsTool, httpTool, dtTool).
		WithAutoExecute(true).
		WithMaxToolRounds(10)

	query := "Get current time in UTC, then fetch https://jsonplaceholder.typicode.com/posts/1 and save it to api.json"
	fmt.Printf("\nQuery: %s\n", query)

	response, err := ai.Ask(ctx, query)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Response: %s\n\n", response)
	os.Remove("api.json")
}
