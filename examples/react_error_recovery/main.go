package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

var attemptCount = 0

func unreliableTool(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation must be a string")
	}

	attemptCount++

	if attemptCount%3 != 0 {
		return "", fmt.Errorf("service unavailable (attempt %d)", attemptCount)
	}

	return fmt.Sprintf("'%s' completed after %d attempts", operation, attemptCount), nil
}

func weirdTool(argsJSON string) (string, error) {
	if argsJSON == `{"mode":"broken"}` {
		return "Plain text response, not JSON!", nil
	}
	return fmt.Sprintf("Processed: %s", argsJSON), nil
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY required")
	}

	unreliableT := agent.NewTool("unreliable_service", "Service that may fail").
		AddParameter("operation", "string", "Operation to perform", true)
	unreliableT.Handler = unreliableTool

	weirdT := agent.NewTool("weird_tool", "Tool with unusual responses").
		AddParameter("mode", "string", "Processing mode", true)
	weirdT.Handler = weirdTool

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(6).
		WithTool(unreliableT).
		WithTool(weirdT)

	fmt.Println("Error Recovery Example")
	fmt.Println("======================\n")

	tasks := []string{
		"Use unreliable_service to process 'data_backup'",
		"Try weird_tool with mode 'broken'",
	}

	ctx := context.Background()

	for i, task := range tasks {
		fmt.Printf("Task %d: %s\n", i+1, task)
		fmt.Println("---")

		attemptCount = 0

		result, err := ai.Execute(ctx, task)

		fmt.Printf("\nResult:\n")
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		}

		if result != nil {
			fmt.Printf("  Success: %v\n", result.Success)
			fmt.Printf("  Iterations: %d\n", result.Iterations)

			errorCount := 0
			for _, step := range result.Steps {
				if step.Type == "ACTION" && step.Error != nil {
					errorCount++
				}
			}

			if errorCount > 0 {
				fmt.Printf("  Errors Encountered: %d\n", errorCount)
				fmt.Printf("\n  Recovery Trace:\n")
				for j, step := range result.Steps {
					if step.Type == "ACTION" {
						if step.Error != nil {
							fmt.Printf("    %d. %s → FAILED: %v\n", j+1, step.Tool, step.Error)
						} else {
							fmt.Printf("    %d. %s → SUCCESS\n", j+1, step.Tool)
						}
					}
				}
			}

			if result.Success {
				fmt.Printf("\n  Answer: %s\n", result.Answer)
			}
		}

		fmt.Println()
	}

	fmt.Println("Done!")
}
