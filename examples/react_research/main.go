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

// Mock search tool that returns predefined results
func searchTool(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query must be a string")
	}

	// Simulated search results
	knowledge := map[string]string{
		"golang":     "Go is a statically typed, compiled language created at Google in 2009. Known for concurrency via goroutines.",
		"react":      "React is a JavaScript library for building user interfaces, created by Facebook in 2013.",
		"python":     "Python is a high-level interpreted language created by Guido van Rossum in 1991.",
		"kubernetes": "Kubernetes is an open-source container orchestration platform originally developed by Google.",
		"ai":         "Artificial Intelligence refers to systems that can perform tasks requiring human intelligence.",
		"llm":        "Large Language Models are AI systems trained on vast text data to understand and generate human language.",
		"gpt":        "GPT (Generative Pre-trained Transformer) is a type of LLM developed by OpenAI.",
		"agent":      "AI agents are autonomous systems that perceive their environment and take actions to achieve goals.",
	}

	queryLower := strings.ToLower(query)
	for key, value := range knowledge {
		if strings.Contains(queryLower, key) {
			return value, nil
		}
	}

	return "No relevant information found.", nil
}

// Summarize tool that combines multiple pieces of information
func summarizeTool(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	text, ok := args["text"].(string)
	if !ok {
		return "", fmt.Errorf("text must be a string")
	}

	// Simple summarization: first sentence + word count
	sentences := strings.Split(text, ".")
	summary := strings.TrimSpace(sentences[0])
	words := len(strings.Fields(text))

	return fmt.Sprintf("Summary: %s. (Original: %d words)", summary, words), nil
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY required")
	}

	// Create tools
	searchT := agent.NewTool("search", "Search for information on a topic").
		AddParameter("query", "string", "Search query", true)
	searchT.Handler = searchTool

	summarizeT := agent.NewTool("summarize", "Summarize text").
		AddParameter("text", "string", "Text to summarize", true)
	summarizeT.Handler = summarizeTool

	// Configure agent with both tools
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(5).
		WithTool(searchT).
		WithTool(summarizeT)

	fmt.Println("Multi-Step Research Example")
	fmt.Println("============================\n")

	tasks := []string{
		"What is Go programming language and who created it?",
		"Compare AI agents and LLMs, then summarize the differences",
	}

	ctx := context.Background()

	for i, task := range tasks {
		fmt.Printf("Task %d: %s\n", i+1, task)
		fmt.Println(strings.Repeat("-", 60))

		result, err := ai.Execute(ctx, task)
		if err != nil {
			log.Printf("Error: %v\n\n", err)
			continue
		}

		if result.Success {
			fmt.Printf("\nFinal Answer:\n%s\n", result.Answer)
			fmt.Printf("\nStats:\n")
			fmt.Printf("  Iterations: %d\n", result.Iterations)
			fmt.Printf("  Steps: %d\n", len(result.Steps))

			// Count tool calls from steps
			toolCalls := 0
			for _, step := range result.Steps {
				if step.Type == "ACTION" {
					toolCalls++
				}
			}

			if toolCalls > 0 {
				fmt.Printf("\nTool Calls: %d\n", toolCalls)
				fmt.Printf("Tool Sequence:\n")
				callNum := 1
				for _, step := range result.Steps {
					if step.Type == "ACTION" {
						fmt.Printf("  %d. %s\n", callNum, step.Tool)
						callNum++
					}
				}
			}

			fmt.Println()
		}
	}

	fmt.Println("Research complete!")
}
