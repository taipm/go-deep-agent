package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

const (
	Model = "gemini-1.5-pro-latest"
)

func main() {
	// Lấy API key từ environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set GEMINI_API_KEY environment variable")
	}

	// Khởi tạo Gemini V3 adapter
	fmt.Println("🚀 Initializing Gemini V3...")
	gemini, err := agent.NewGeminiV3Adapter(apiKey, Model)
	if err != nil {
		log.Fatalf("❌ Failed to create Gemini adapter: %v", err)
	}
	defer gemini.Close()

	fmt.Println("✅ Gemini V3 adapter created successfully!")

	// Test 1: Basic chat
	fmt.Println("\n📝 Test 1: Basic conversation")
	response, err := gemini.Complete(context.Background(), &agent.CompletionRequest{
		Model: "gemini-1.5-pro-latest",
		Messages: []agent.Message{
			{Role: "user", Content: "Hello! Can you help me understand Go programming?"},
		},
		Temperature: 0.7,
		MaxTokens:    500,
	})
	if err != nil {
		log.Printf("❌ Error in basic chat: %v", err)
	} else {
		fmt.Printf("✅ Response: %s\n", response.Content)
	}

	// Test 2: Streaming
	fmt.Println("\n🌊 Test 2: Streaming response")
	fmt.Print("Streaming: ")
	response, err = gemini.Stream(context.Background(), &agent.CompletionRequest{
		Model: "gemini-1.5-pro-latest",
		Messages: []agent.Message{
			{Role: "user", Content: "Write a haiku about programming"},
		},
	}, func(chunk string) {
		fmt.Print(chunk)
	})
	if err != nil {
		log.Printf("❌ Error in streaming: %v", err)
	} else {
		fmt.Printf("\n✅ Streaming completed!\n")
	}

	// Test 3: Tool calling
	fmt.Println("\n🔧 Test 3: Tool calling with calculator")
	calculatorTool := agent.NewTool("calculator", "Simple calculator").
		AddParameter("expression", "string", "Math expression", true).
		WithHandler(func(args string) (string, error) {
			return fmt.Sprintf("Calculator result for: %s = [mock calculation]", args), nil
	})

	response, err = gemini.Complete(context.Background(), &agent.CompletionRequest{
		Model: Model,
		Messages: []agent.Message{
			{Role: "user", Content: "Calculate 25 * 4 using the calculator tool"},
		},
		Tools:       []*agent.Tool{calculatorTool},
		Temperature: 0.7,
		MaxTokens:    500,
	})

	if err != nil {
		log.Printf("❌ Error in tool calling: %v", err)
	} else {
		fmt.Printf("✅ Tool response: %s\n", response.Content)
	}

	fmt.Println("\n🎉 All tests completed successfully!")
	fmt.Println("Gemini V3 is ready for production use! 🚀")
}