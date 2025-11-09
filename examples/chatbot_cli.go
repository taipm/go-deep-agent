package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Print welcome banner
	printBanner()

	// Get user preferences
	model, provider := selectModel()

	// Get API key from environment (only for OpenAI)
	apiKey := ""
	if provider == "openai" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("\n❌ Error: OPENAI_API_KEY environment variable not set")
			fmt.Println("Please set it with: export OPENAI_API_KEY='your-api-key'")
			os.Exit(1)
		}
	}

	useStreaming := askYesNo("\nEnable streaming mode? (y/n): ")
	useMemory := askYesNo("Enable conversation memory? (y/n): ")

	// Build chatbot with selected options
	var chatbot *agent.Builder

	if provider == "openai" {
		chatbot = agent.NewOpenAI(model, apiKey)
	} else {
		// Ollama (local)
		chatbot = agent.NewOllama(model).WithBaseURL("http://localhost:11434/v1")
	}

	// Configure chatbot
	chatbot = chatbot.
		WithSystem("You are a helpful, friendly assistant. Keep responses concise and clear.").
		WithTemperature(0.7).
		WithTimeout(30 * time.Second)

	if useMemory {
		chatbot = chatbot.WithMemory().WithMaxHistory(20)
		fmt.Println("✅ Conversation memory enabled (max 20 messages)")
	}

	if useStreaming {
		chatbot = chatbot.OnStream(func(chunk string) {
			fmt.Print(chunk)
		})
		fmt.Println("✅ Streaming mode enabled")
	}

	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("🤖 Chatbot ready! Type your message and press Enter.")
	fmt.Println("💡 Commands: /help, /clear, /stats, /exit")
	fmt.Println(strings.Repeat("─", 60) + "\n")

	// Main chat loop
	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(userInput, "/") {
			if handleCommand(userInput, chatbot) {
				break // Exit command
			}
			continue
		}

		// Send message to chatbot
		fmt.Print("AI:  ")
		start := time.Now()

		var response string
		var err error

		if useStreaming {
			response, err = chatbot.Stream(ctx, userInput)
			fmt.Println() // Newline after streaming
		} else {
			response, err = chatbot.Ask(ctx, userInput)
			fmt.Println(response)
		}

		if err != nil {
			fmt.Printf("\n❌ Error: %v\n", err)
			continue
		}

		// Show response time
		elapsed := time.Since(start)
		fmt.Printf("\n⏱️  Response time: %.2fs\n\n", elapsed.Seconds())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}

	fmt.Println("\n👋 Goodbye! Thanks for chatting!")
}

// printBanner prints welcome banner
func printBanner() {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("   🤖 GO-DEEP-AGENT CHATBOT CLI")
	fmt.Println("   Interactive AI Assistant powered by go-deep-agent")
	fmt.Println(strings.Repeat("═", 60) + "\n")
}

// selectModel lets user choose AI model and provider
func selectModel() (string, string) {
	fmt.Println("Select AI Provider:")
	fmt.Println("1. OpenAI (GPT-4o-mini) - Fast, efficient")
	fmt.Println("2. OpenAI (GPT-4o) - Most capable")
	fmt.Println("3. OpenAI (GPT-4-turbo) - Advanced reasoning")
	fmt.Println("4. Ollama (qwen2.5:1.5b) - Local, private, fast")
	fmt.Println("5. Ollama (llama3.2) - Local, private")
	fmt.Print("\nYour choice (1-5): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		return "gpt-4o-mini", "openai"
	case "2":
		return "gpt-4o", "openai"
	case "3":
		return "gpt-4-turbo", "openai"
	case "4":
		return "qwen2.5:1.5b", "ollama"
	case "5":
		return "llama3.2", "ollama"
	default:
		fmt.Println("\nInvalid choice, using default: Ollama (qwen2.5:1.5b)")
		return "qwen2.5:1.5b", "ollama"
	}
}

// askYesNo asks a yes/no question
func askYesNo(prompt string) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}

// handleCommand processes chatbot commands
func handleCommand(cmd string, chatbot *agent.Builder) bool {
	cmd = strings.ToLower(cmd)

	switch cmd {
	case "/help":
		fmt.Println("\n📚 Available Commands:")
		fmt.Println("  /help   - Show this help message")
		fmt.Println("  /clear  - Clear cache")
		fmt.Println("  /stats  - Show cache statistics")
		fmt.Println("  /exit   - Exit the chatbot\n")
		return false

	case "/clear":
		err := chatbot.ClearCache(context.Background())
		if err != nil {
			fmt.Printf("⚠️  Cache clear failed: %v\n\n", err)
		} else {
			fmt.Println("✅ Cache cleared\n")
		}
		return false

	case "/stats":
		stats := chatbot.GetCacheStats()
		total := stats.Hits + stats.Misses
		hitRate := 0.0
		if total > 0 {
			hitRate = float64(stats.Hits) / float64(total) * 100
		}
		fmt.Println("\n📊 Cache Statistics:")
		fmt.Printf("  Hits:       %d\n", stats.Hits)
		fmt.Printf("  Misses:     %d\n", stats.Misses)
		fmt.Printf("  Size:       %d entries\n", stats.Size)
		fmt.Printf("  Evictions:  %d\n", stats.Evictions)
		fmt.Printf("  Hit Rate:   %.2f%%\n\n", hitRate)
		return false

	case "/exit", "/quit", "/q":
		return true

	default:
		fmt.Printf("❌ Unknown command: %s\n", cmd)
		fmt.Println("Type /help for available commands\n")
		return false
	}
}
