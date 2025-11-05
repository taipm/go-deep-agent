package main
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
	fmt.Println("=== OpenAI Tool Calling Test ===\n")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	ctx := context.Background()

	// Test 1: Simple calculator tool
	test1_Calculator(ctx, apiKey)

	// Test 2: Weather tool
	test2_Weather(ctx, apiKey)

	// Test 3: Multiple tools
	test3_MultipleTool(ctx, apiKey)
}

func test1_Calculator(ctx context.Context, apiKey string) {
	fmt.Println("--- Test 1: Calculator Tool ---")

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
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}

			fmt.Printf("  🔧 Tool called: calculate(%s, %.0f, %.0f)\n", params.Operation, params.A, params.B)

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

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(calculatorTool).
		WithAutoExecute(true).
		WithMaxToolRounds(3).
		Ask(ctx, "What is 123 multiplied by 456?")

	if err != nil {
		log.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Final answer: %s\n\n", response)
}

func test2_Weather(ctx context.Context, apiKey string) {
	fmt.Println("--- Test 2: Weather Tool ---")

	weatherTool := agent.NewTool("get_weather", "Get the current weather for a location").
		AddParameter("location", "string", "The city name", true).
		AddParameter("units", "string", "Temperature units: celsius or fahrenheit", false).
		WithHandler(func(args string) (string, error) {
			var params struct {
				Location string `json:"location"`
				Units    string `json:"units"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}

			if params.Units == "" {
				params.Units = "celsius"
			}

			fmt.Printf("  🔧 Tool called: get_weather(%s, %s)\n", params.Location, params.Units)

			// Simulate weather API
			temp := 25
			if params.Units == "fahrenheit" {
				temp = 77
			}

			return fmt.Sprintf("The weather in %s is sunny with a temperature of %d°%s",
				params.Location, temp, map[string]string{"celsius": "C", "fahrenheit": "F"}[params.Units]), nil
		})

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(weatherTool).
		WithAutoExecute(true).
		Ask(ctx, "What's the weather like in Paris?")

	if err != nil {
		log.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Final answer: %s\n\n", response)
}

func test3_MultipleTool(ctx context.Context, apiKey string) {
	fmt.Println("--- Test 3: Multiple Tools ---")

	// Weather tool
	weatherTool := agent.NewTool("get_weather", "Get weather for a location").
		AddParameter("location", "string", "City name", true).
		WithHandler(func(args string) (string, error) {
			var params struct {
				Location string `json:"location"`
			}
			json.Unmarshal([]byte(args), &params)
			fmt.Printf("  🔧 Tool called: get_weather(%s)\n", params.Location)
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
			fmt.Printf("  🔧 Tool called: get_time(%s)\n", params.Location)
			return fmt.Sprintf("Current time in %s: 14:30", params.Location), nil
		})

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(weatherTool, timeTool).
		WithAutoExecute(true).
		Ask(ctx, "What's the weather and time in Tokyo?")

	if err != nil {
		log.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Final answer: %s\n\n", response)
}
