package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func weatherTool(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	city := args["city"].(string)
	time.Sleep(500 * time.Millisecond)
	return fmt.Sprintf("%s: 22°C, Sunny", city), nil
}

func newsTool(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	topic := args["topic"].(string)
	time.Sleep(500 * time.Millisecond)
	return fmt.Sprintf("Latest on %s: Breaking news update", topic), nil
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY required")
	}

	weatherT := agent.NewTool("weather", "Get weather for a city").
		AddParameter("city", "string", "City name", true)
	weatherT.Handler = weatherTool

	newsT := agent.NewTool("news", "Get news on a topic").
		AddParameter("topic", "string", "News topic", true)
	newsT.Handler = newsTool

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(5).
		WithTool(weatherT).
		WithTool(newsT)

	fmt.Println("Streaming ReAct Example")
	fmt.Println("=======================\n")

	task := "Get the weather in Paris and latest tech news"

	fmt.Printf("Task: %s\n\n", task)

	ctx := context.Background()
	events, err := ai.StreamReAct(ctx, task)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	iterCount := 0
	stepCount := 0

	for event := range events {
		switch event.Type {
		case "start":
			fmt.Println("🚀 Starting ReAct execution...")

		case "thought":
			fmt.Printf("\n💭 [Iteration %d] THOUGHT:\n   %s\n",
				event.Iteration, truncate(event.Content, 80))
			iterCount = event.Iteration

		case "action":
			stepCount++
			fmt.Printf("\n🔧 [Step %d] ACTION:\n   Tool: %s\n",
				stepCount, event.Step.Tool)

		case "observation":
			fmt.Printf("👁️  OBSERVATION:\n   %s\n", truncate(event.Content, 80))

		case "final":
			fmt.Printf("\n✅ FINAL ANSWER:\n   %s\n", truncate(event.Content, 100))

		case "error":
			fmt.Printf("\n❌ ERROR: %v\n", event.Error)

		case "complete":
			fmt.Printf("\n📊 Execution Complete!\n")
			fmt.Printf("   Total iterations: %d\n", iterCount)
			fmt.Printf("   Total steps: %d\n", stepCount)
		}
	}

	fmt.Println("\nDone!")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
