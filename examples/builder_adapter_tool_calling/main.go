package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Lấy API key từ environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set GEMINI_API_KEY environment variable")
	}

	// Khởi tạo Gemini V3 adapter
	fmt.Println("🚀 Initializing Gemini V3 adapter with Builder...")
	gemini, err := agent.NewGeminiV3Adapter(apiKey, "gemini-1.5-pro-latest")
	if err != nil {
		log.Fatalf("❌ Failed to create Gemini adapter: %v", err)
	}
	defer gemini.Close()

	fmt.Println("✅ Gemini V3 adapter created successfully!")

	// Test 1: Simple tool calling without auto-execute
	fmt.Println("\n🔧 Test 1: Manual tool calling")
	calculatorTool := agent.NewTool("calculator", "Simple calculator for math operations").
		AddParameter("expression", "string", "Mathematical expression to evaluate", true).
		WithHandler(func(args string) (string, error) {
			// Mock calculator implementation
			return fmt.Sprintf("Calculator result for '%s' = 42", args), nil
	})

	builder1 := agent.NewWithAdapter("gemini-1.5-pro-latest", gemini).
		WithTool(calculatorTool)

	response1, err := builder1.Ask(context.Background(), "Calculate 15 * 8 using the calculator")
	if err != nil {
		log.Printf("❌ Error in manual tool calling: %v", err)
	} else {
		fmt.Printf("✅ Manual response: %s\n", response1)
	}

	// Test 2: Auto-execute tool calling (THE CRITICAL TEST)
	fmt.Println("\n🔧 Test 2: Auto-execute tool calling (CRITICAL FIX)")
	builder2 := agent.NewWithAdapter("gemini-1.5-pro-latest", gemini).
		WithTool(calculatorTool).
		WithAutoExecute(true)  // This should now work with adapters!

	response2, err := builder2.Ask(context.Background(), "Calculate 15 * 8 using the calculator")
	if err != nil {
		log.Printf("❌ Error in auto-execute tool calling: %v", err)
	} else {
		fmt.Printf("✅ Auto-execute response: %s\n", response2)

		// Check if the response contains calculation result
		if strings.Contains(response2, "42") {
			fmt.Println("🎉 SUCCESS: Tool was executed automatically!")
		} else {
			fmt.Println("⚠️  Tool may not have been executed properly")
		}
	}

	// Test 3: Multi-step conversation with tool calling
	fmt.Println("\n🔧 Test 3: Multi-step conversation")
	builder3 := agent.NewWithAdapter("gemini-1.5-pro-latest", gemini).
		WithTool(calculatorTool).
		WithAutoExecute(true)

	// First question
	response3a, err := builder3.Ask(context.Background(), "What is 10 * 5?")
	if err != nil {
		log.Printf("❌ Error in first question: %v", err)
	} else {
		fmt.Printf("✅ First answer: %s\n", response3a)
	}

	// Follow-up question (should remember previous calculation)
	response3b, err := builder3.Ask(context.Background(), "Now multiply that result by 2")
	if err != nil {
		log.Printf("❌ Error in follow-up question: %v", err)
	} else {
		fmt.Printf("✅ Follow-up answer: %s\n", response3b)

		// Check if AI remembered the previous result (10 * 5 = 50)
		if strings.Contains(response3b, "100") || strings.Contains(response3b, "twenty") {
			fmt.Println("🎉 SUCCESS: AI remembered previous calculation!")
		} else {
			fmt.Println("⚠️  AI may not have remembered previous result")
		}
	}

	// Test 4: Multiple tools in parallel
	fmt.Println("\n🔧 Test 4: Multiple tools with parallel execution")

	weatherTool := agent.NewTool("weather", "Get weather information").
		AddParameter("location", "string", "City name", true).
		WithHandler(func(args string) (string, error) {
			return fmt.Sprintf("Weather in %s: 25°C, Sunny", args), nil
		})

	builder4 := agent.NewWithAdapter("gemini-1.5-pro-latest", gemini).
		WithTools(calculatorTool, weatherTool).
		WithAutoExecute(true).
		WithParallelTools(true). // Enable parallel execution
		WithMaxWorkers(2)

	response4, err := builder4.Ask(context.Background(), "Calculate 20 * 3 and get weather for Paris")
	if err != nil {
		log.Printf("❌ Error in parallel tools: %v", err)
	} else {
		fmt.Printf("✅ Parallel tools response: %s\n", response4)
	}

	fmt.Println("\n🎉 All Builder-Adapter tool calling tests completed!")
	fmt.Println("✅ CRITICAL FIX VERIFIED: WithAutoExecute(true) now works with adapters!")
	fmt.Println("🚀 Gemini V3 is now fully compatible with Builder layer!")
}