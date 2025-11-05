package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Builder API Tool Calling Examples ===\n")

	// Example 1: Simple Tool
	example1_SimpleTool()

	// Example 2: Tool with Auto-Execution
	example2_AutoExecution()

	// Example 3: Multiple Tools
	example3_MultipleTools()
}

// Example 1: Define and use a simple tool
func example1_SimpleTool() {
	fmt.Println("--- Example 1: Simple Tool Definition ---")

	ctx := context.Background()

	// Define a weather tool
	weatherTool := agent.NewTool("get_weather", "Get the current weather for a location").
		AddParameter("location", "string", "The city name", true).
		AddParameter("units", "string", "Temperature units (celsius/fahrenheit)", false).
		WithHandler(func(args string) (string, error) {
			var params struct {
				Location string `json:"location"`
				Units    string `json:"units"`
			}
			json.Unmarshal([]byte(args), &params)

			// Simulate weather API call
			return fmt.Sprintf("Weather in %s: Sunny, 25°C", params.Location), nil
		})

	// Use the tool WITHOUT auto-execution (manual)
	response, err := agent.NewOllama("qwen3:1.7b").
		WithTool(weatherTool).
		Ask(ctx, "What's the weather like in Paris?")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", response)
	fmt.Println()
}

// Example 2: Auto-execute tool calls
func example2_AutoExecution() {
	fmt.Println("--- Example 2: Auto-Execution ---")

	ctx := context.Background()

	// Define a calculator tool
	calculatorTool := agent.NewTool("calculate", "Perform arithmetic calculations").
		AddParameter("operation", "string", "Operation: add, subtract, multiply, divide", true).
		AddParameter("a", "number", "First number", true).
		AddParameter("b", "number", "Second number", true).
		WithHandler(func(args string) (string, error) {
			var params struct {
				Operation string  `json:"operation"`
				A         float64 `json:"a"`
				B         float64 `json:"b"`
			}
			json.Unmarshal([]byte(args), &params)

			var result float64
			switch params.Operation {
			case "add":
				result = params.A + params.B
			case "subtract":
				result = params.A - params.B
			case "multiply":
				result = params.A * params.B
			case "divide":
				if params.B == 0 {
					return "", fmt.Errorf("division by zero")
				}
				result = params.A / params.B
			default:
				return "", fmt.Errorf("unknown operation: %s", params.Operation)
			}

			return fmt.Sprintf("%.2f", result), nil
		})

	// With auto-execution, tool calls are handled automatically
	response, err := agent.NewOllama("qwen3:1.7b").
		WithTool(calculatorTool).
		WithAutoExecute(true).
		Ask(ctx, "What is 123 multiplied by 456?")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Q: What is 123 multiplied by 456?\n")
	fmt.Printf("A: %s\n", response)
	fmt.Println()
}

// Example 3: Multiple tools working together
func example3_MultipleTools() {
	fmt.Println("--- Example 3: Multiple Tools ---")

	ctx := context.Background()

	// Weather tool
	weatherTool := agent.NewTool("get_weather", "Get weather for a location").
		AddParameter("location", "string", "City name", true).
		WithHandler(func(args string) (string, error) {
			var params struct {
				Location string `json:"location"`
			}
			json.Unmarshal([]byte(args), &params)
			return fmt.Sprintf("Weather in %s: Sunny, 25°C", params.Location), nil
		})

	// Time tool
	timeTool := agent.NewTool("get_time", "Get current time for a location").
		AddParameter("location", "string", "City name", true).
		WithHandler(func(args string) (string, error) {
			var params struct {
				Location string `json:"location"`
			}
			json.Unmarshal([]byte(args), &params)
			return fmt.Sprintf("Current time in %s: 14:30 PM", params.Location), nil
		})

	// Use multiple tools
	response, err := agent.NewOllama("qwen3:1.7b").
		WithTools(weatherTool, timeTool).
		WithAutoExecute(true).
		Ask(ctx, "What's the weather and time in Tokyo?")

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Q: What's the weather and time in Tokyo?\n")
	fmt.Printf("A: %s\n", response)
	fmt.Println()
}
