package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

// Example demonstrating parallel tool execution with the Builder API
func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Create tools that simulate real work (e.g., API calls, database queries)
	weatherTool := &agent.Tool{
		Name:        "get_weather",
		Description: "Get current weather for a city",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"city": map[string]interface{}{
					"type":        "string",
					"description": "City name",
				},
			},
			"required": []string{"city"},
		},
		Handler: func(args string) (string, error) {
			// Simulate API call delay
			time.Sleep(500 * time.Millisecond)
			return fmt.Sprintf("Weather data for %s: 72°F, Sunny", args), nil
		},
	}

	stockTool := &agent.Tool{
		Name:        "get_stock_price",
		Description: "Get current stock price",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol": map[string]interface{}{
					"type":        "string",
					"description": "Stock symbol (e.g., AAPL)",
				},
			},
			"required": []string{"symbol"},
		},
		Handler: func(args string) (string, error) {
			// Simulate API call delay
			time.Sleep(500 * time.Millisecond)
			return fmt.Sprintf("Stock price for %s: $150.25", args), nil
		},
	}

	newsTool := &agent.Tool{
		Name:        "get_news",
		Description: "Get latest news headlines",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"topic": map[string]interface{}{
					"type":        "string",
					"description": "News topic",
				},
			},
			"required": []string{"topic"},
		},
		Handler: func(args string) (string, error) {
			// Simulate API call delay
			time.Sleep(500 * time.Millisecond)
			return fmt.Sprintf("Latest news about %s: Breaking developments...", args), nil
		},
	}

	// Example 1: Sequential execution (default)
	fmt.Println("=== Example 1: Sequential Tool Execution ===")
	sequentialAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(weatherTool, stockTool, newsTool).
		WithAutoExecute(true).
		WithMaxToolRounds(3)

	start := time.Now()
	ctx := context.Background()
	response, err := sequentialAgent.Ask(ctx, "Give me the weather in San Francisco, stock price of AAPL, and latest tech news")
	if err != nil {
		log.Printf("Sequential execution error: %v", err)
	} else {
		fmt.Printf("Response: %s\n", response)
		fmt.Printf("Sequential execution time: %v\n\n", time.Since(start))
	}

	// Example 2: Parallel execution (3x faster for independent tools)
	fmt.Println("=== Example 2: Parallel Tool Execution ===")
	parallelAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(weatherTool, stockTool, newsTool).
		WithAutoExecute(true).
		WithMaxToolRounds(3).
		WithParallelTools(true).         // Enable parallel execution
		WithMaxWorkers(10).              // Allow up to 10 concurrent tools
		WithToolTimeout(5 * time.Second) // Timeout per tool

	start = time.Now()
	response, err = parallelAgent.Ask(ctx, "Give me the weather in San Francisco, stock price of AAPL, and latest tech news")
	if err != nil {
		log.Printf("Parallel execution error: %v", err)
	} else {
		fmt.Printf("Response: %s\n", response)
		fmt.Printf("Parallel execution time: %v (should be ~500ms vs ~1500ms sequential)\n\n", time.Since(start))
	}

	// Example 3: Limited concurrency (max 2 workers)
	fmt.Println("=== Example 3: Limited Concurrency (2 workers) ===")
	limitedAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(weatherTool, stockTool, newsTool).
		WithAutoExecute(true).
		WithMaxToolRounds(3).
		WithParallelTools(true).
		WithMaxWorkers(2). // Limit to 2 concurrent tools
		WithToolTimeout(5 * time.Second)

	start = time.Now()
	response, err = limitedAgent.Ask(ctx, "Give me the weather in NYC, stock price of GOOGL, and latest AI news")
	if err != nil {
		log.Printf("Limited concurrency error: %v", err)
	} else {
		fmt.Printf("Response: %s\n", response)
		fmt.Printf("Limited concurrency time: %v (should be ~1000ms with 2 workers)\n\n", time.Since(start))
	}

	// Example 4: Custom timeout handling
	fmt.Println("=== Example 4: Custom Timeout Handling ===")
	timeoutAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(weatherTool, stockTool, newsTool).
		WithAutoExecute(true).
		WithMaxToolRounds(3).
		WithParallelTools(true).
		WithMaxWorkers(10).
		WithToolTimeout(100 * time.Millisecond) // Very short timeout

	start = time.Now()
	response, err = timeoutAgent.Ask(ctx, "Give me the weather in London, stock price of MSFT, and latest space news")
	if err != nil {
		fmt.Printf("Expected timeout error: %v\n", err)
		fmt.Printf("Execution time before timeout: %v\n\n", time.Since(start))
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	// Performance comparison
	fmt.Println("=== Performance Summary ===")
	fmt.Println("Sequential: ~1500ms (3 tools × 500ms each)")
	fmt.Println("Parallel (10 workers): ~500ms (all 3 execute concurrently)")
	fmt.Println("Parallel (2 workers): ~1000ms (2 concurrent + 1 queued)")
	fmt.Println("\nConfiguration Options:")
	fmt.Println("  WithParallelTools(true)         - Enable parallel execution")
	fmt.Println("  WithMaxWorkers(n)               - Limit concurrent tools (default: 10)")
	fmt.Println("  WithToolTimeout(duration)       - Set per-tool timeout (default: 30s)")
}
