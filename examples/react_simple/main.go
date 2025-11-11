package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/taipm/go-deep-agent/agent"
)

func calculatorTool(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	expr, ok := args["expression"].(string)
	if !ok {
		return "", fmt.Errorf("expression must be a string")
	}

	expr = strings.ReplaceAll(expr, " ", "")

	for _, op := range []string{"+", "-", "*", "/"} {
		if strings.Contains(expr, op) {
			parts := strings.Split(expr, op)
			if len(parts) != 2 {
				return "", fmt.Errorf("invalid expression")
			}

			a, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return "", err
			}

			b, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return "", err
			}

			var result float64
			switch op {
			case "+":
				result = a + b
			case "-":
				result = a - b
			case "*":
				result = a * b
			case "/":
				if b == 0 {
					return "", fmt.Errorf("division by zero")
				}
				result = a / b
			}

			return fmt.Sprintf("%.2f", result), nil
		}
	}

	return "", fmt.Errorf("unsupported expression")
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY required")
	}

	calcTool := agent.NewTool("calculator", "Performs arithmetic").
		AddParameter("expression", "string", "Math expression", true)
	calcTool.Handler = calculatorTool

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(3).
		WithTool(calcTool)

	fmt.Println("Simple ReAct Example")
	fmt.Println("====================\n")

	tasks := []string{
		"What is 25 + 17?",
		"Calculate 100 divided by 4",
	}

	ctx := context.Background()

	for i, task := range tasks {
		fmt.Printf("Task %d: %s\n", i+1, task)
		
		result, err := ai.Execute(ctx, task)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		if result.Success {
			fmt.Printf("Answer: %s\n", result.Answer)
			fmt.Printf("Stats: %d iterations\n\n", result.Iterations)
		}
	}

	fmt.Println("Done!")
}
