package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/taipm/go-deep-agent/agent"
)

func searchTool(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	query := args["query"].(string)
	return fmt.Sprintf("Found: %s is a container orchestration platform", query), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY required")
	}

	searchT := agent.NewTool("search", "Search for information").
		AddParameter("query", "string", "Search query", true)
	searchT.Handler = searchTool

	customTemplate := `You are a research assistant using ReAct.

Available Tools:
{tools}

Examples:
{examples}

Rules:
- Always THOUGHT before ACTION
- Use tools for information
- Provide detailed FINAL answers

Task: {task}`

	examples := []agent.ReActExample{
		{
			Task: "Find info about Go",
			Steps: []string{
				`THOUGHT: Need to search for Go`,
				`ACTION: search(query="golang")`,
				`OBSERVATION: Go is a programming language`,
				`FINAL: Go is a statically typed language`,
			},
		},
	}

	callback := &agent.EnhancedReActCallback{
		OnThought: func(content string, iteration int) {
			fmt.Printf("💭 [%d] Thinking: %s\n", iteration, truncate(content, 60))
		},
		OnAction: func(tool string, args map[string]interface{}, iteration int) {
			fmt.Printf("🔧 [%d] Tool: %s\n", iteration, tool)
		},
		OnObservation: func(content string, iteration int) {
			fmt.Printf("👁️  [%d] Result: %s\n", iteration, truncate(content, 60))
		},
		OnFinal: func(answer string, iteration int) {
			fmt.Printf("✅ [%d] Answer ready\n", iteration)
		},
		OnCompleted: func(result *agent.ReActResult) {
			fmt.Printf("\n📊 Execution complete: %d iterations\n", result.Iterations)
		},
	}

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(5).
		WithReActPromptTemplate(customTemplate).
		WithReActExamples(examples[0]).
		WithReActCallback(callback).
		WithTool(searchT)

	fmt.Println("Advanced Configuration Example")
	fmt.Println("================================\n")

	tasks := []string{
		"Search for information about Kubernetes",
	}

	ctx := context.Background()

	for _, task := range tasks {
		fmt.Printf("Task: %s\n", task)
		fmt.Println(strings.Repeat("-", 70))

		result, err := ai.Execute(ctx, task)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		fmt.Println()
		if result.Success {
			fmt.Printf("Final Answer:\n%s\n", result.Answer)
			fmt.Printf("\nMetrics:\n")
			fmt.Printf("  Iterations: %d\n", result.Iterations)
			fmt.Printf("  Steps: %d\n", len(result.Steps))
		}
	}

	fmt.Println("\nDone!")
}
