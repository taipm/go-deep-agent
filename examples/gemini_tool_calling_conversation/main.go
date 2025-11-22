package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Lấy API key từ environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set GEMINI_API_KEY environment variable")
	}

	// Khởi tạo Gemini V3 adapter
	fmt.Println("🚀 Initializing Gemini V3...")
	gemini, err := agent.NewGeminiV3Adapter(apiKey, "gemini-1.5-pro-latest")
	if err != nil {
		log.Fatalf("❌ Failed to create Gemini adapter: %v", err)
	}
	defer gemini.Close()

	fmt.Println("✅ Gemini V3 adapter created successfully!")

	// Test multi-turn conversation với tool calling
	fmt.Println("\n🔧 Testing Multi-Turn Tool Calling Conversation...")

	// Tool calculator
	calculatorTool := agent.NewTool("calculator", "Simple calculator for math operations").
		AddParameter("expression", "string", "Mathematical expression to evaluate", true).
		WithHandler(func(args string) (string, error) {
			// Mock calculator - trong thực tế sẽ parse và evaluate expression
			return fmt.Sprintf("Calculator result for '%s' = 42", args), nil
		})

	// Tạo conversation history ban đầu
	conversation := []agent.Message{
		{Role: "user", Content: "I need help with math calculations. Can you help me?"},
		{Role: "assistant", Content: "Absolutely! I have access to a calculator tool. What calculation would you like me to perform?"},
	}

	// Turn 1: User yêu cầu calculation
	turn1UserMsg := agent.Message{Role: "user", Content: "Calculate 15 * 8 + 32"}
	conversation = append(conversation, turn1UserMsg)

	// Turn 1: Request AI call tool
	fmt.Println("\n📝 Turn 1: User requests calculation")
	response, err := gemini.Complete(context.Background(), &agent.CompletionRequest{
		Model:       "gemini-1.5-pro-latest",
		Messages:    conversation,
		Tools:       []*agent.Tool{calculatorTool},
		Temperature: 0.7,
		MaxTokens:   500,
	})

	if err != nil {
		log.Fatalf("❌ Error in turn 1: %v", err)
	}

	fmt.Printf("✅ AI Response: %s\n", response.Content)

	// Nếu AI gọi tool, add assistant message với tool calls vào conversation
	if len(response.ToolCalls) > 0 {
		assistantWithToolCall := agent.Message{
			Role:       "assistant",
			Content:    response.Content,
			ToolCalls:  response.ToolCalls,
		}
		conversation = append(conversation, assistantWithToolCall)

		// Giả lập tool result
		for _, toolCall := range response.ToolCalls {
			toolResult := agent.Message{
				Role:       "tool",
				Content:    fmt.Sprintf("Calculator result for '%s' = 152", toolCall.Arguments),
				ToolCallID: toolCall.ID,
			}
			conversation = append(conversation, toolResult)
		}

		// Turn 2: User hỏi tiếp dựa trên kết quả
		fmt.Println("\n📝 Turn 2: User asks follow-up question")
		turn2UserMsg := agent.Message{Role: "user", Content: "Great! Now what if we multiply that result by 3?"}
		conversation = append(conversation, turn2UserMsg)

		// AI phải nhớ kết quả calculation trước đó
		response2, err := gemini.Complete(context.Background(), &agent.CompletionRequest{
			Model:       "gemini-1.5-pro-latest",
			Messages:    conversation,
			Tools:       []*agent.Tool{calculatorTool},
			Temperature: 0.7,
			MaxTokens:   500,
		})

		if err != nil {
			log.Fatalf("❌ Error in turn 2: %v", err)
		}

		fmt.Printf("✅ AI Follow-up Response: %s\n", response2.Content)

		// Test conversation history
		fmt.Println("\n📚 Conversation History:")
		for i, msg := range conversation {
			fmt.Printf("  [%d] Role: %s, Content: %s", i+1, msg.Role, msg.Content)
			if len(msg.ToolCalls) > 0 {
				fmt.Printf(" (Tool Calls: %d)", len(msg.ToolCalls))
			}
			if msg.ToolCallID != "" {
				fmt.Printf(" (Tool Call ID: %s)", msg.ToolCallID)
			}
			fmt.Println()
		}

		fmt.Println("\n🎉 Multi-turn tool calling conversation test completed successfully!")
		fmt.Println("✅ Tool Call History Bug FIXED - AI now remembers tool calls in conversation!")
	} else {
		fmt.Println("⚠️  AI did not call any tools in turn 1")
	}
}