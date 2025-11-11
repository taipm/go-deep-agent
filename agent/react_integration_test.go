package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

// Integration tests require OPENAI_API_KEY environment variable
// Run with: OPENAI_API_KEY=xxx go test -v -run TestReActIntegration

func skipIfNoAPIKey(t *testing.T) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}
	return apiKey
}

// Mock calculator tool for integration tests
func integrationCalculator(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	expr, ok := args["expression"].(string)
	if !ok {
		return "", fmt.Errorf("expression required")
	}

	// Simple calculator: only handles "X + Y" format
	var a, b float64
	if n, _ := fmt.Sscanf(expr, "%f + %f", &a, &b); n == 2 {
		return fmt.Sprintf("%.2f", a+b), nil
	}
	if n, _ := fmt.Sscanf(expr, "%f - %f", &a, &b); n == 2 {
		return fmt.Sprintf("%.2f", a-b), nil
	}
	if n, _ := fmt.Sscanf(expr, "%f * %f", &a, &b); n == 2 {
		return fmt.Sprintf("%.2f", a*b), nil
	}
	if n, _ := fmt.Sscanf(expr, "%f / %f", &a, &b); n == 2 {
		if b == 0 {
			return "", fmt.Errorf("division by zero")
		}
		return fmt.Sprintf("%.2f", a/b), nil
	}

	return "", fmt.Errorf("unsupported expression format")
}

func TestReActIntegration_SimpleCalculation(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	calcTool := NewTool("calculator", "Performs basic arithmetic operations").
		AddParameter("expression", "string", "Math expression like '5 + 3'", true)
	calcTool.Handler = integrationCalculator

	ai := NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(3).
		WithTool(calcTool).
		WithTimeout(30 * time.Second)

	ctx := context.Background()
	result, err := ai.Execute(ctx, "What is 25 + 17?")

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure")
	}

	if result.Iterations == 0 {
		t.Errorf("Expected at least 1 iteration")
	}

	if len(result.Steps) == 0 {
		t.Errorf("Expected steps in result")
	}

	t.Logf("Result: %s", result.Answer)
	t.Logf("Iterations: %d", result.Iterations)
	t.Logf("Steps: %d", len(result.Steps))
}

func TestReActIntegration_MultiStep(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	calcTool := NewTool("calculator", "Performs arithmetic").
		AddParameter("expression", "string", "Math expression", true)
	calcTool.Handler = integrationCalculator

	ai := NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(5).
		WithTool(calcTool).
		WithTimeout(30 * time.Second)

	ctx := context.Background()
	result, err := ai.Execute(ctx, "Calculate 10 + 5, then multiply that result by 2")

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success")
	}

	// Should require multiple iterations
	if result.Iterations < 2 {
		t.Errorf("Expected at least 2 iterations for multi-step, got %d", result.Iterations)
	}

	t.Logf("Multi-step result: %s", result.Answer)
	t.Logf("Iterations: %d", result.Iterations)
}

func TestReActIntegration_ErrorRecovery(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	attempts := 0
	unreliableTool := NewTool("unreliable", "May fail on first attempt").
		AddParameter("data", "string", "Data to process", true)
	unreliableTool.Handler = func(argsJSON string) (string, error) {
		attempts++
		if attempts == 1 {
			return "", fmt.Errorf("temporary failure")
		}
		return "success after retry", nil
	}

	ai := NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(5).
		WithTool(unreliableTool).
		WithTimeout(30 * time.Second)

	ctx := context.Background()
	result, err := ai.Execute(ctx, "Use unreliable tool to process 'test data'. If it fails, try again.")

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success after retry")
	}

	if attempts < 2 {
		t.Errorf("Expected at least 2 attempts (with retry), got %d", attempts)
	}

	t.Logf("Error recovery result: %s", result.Answer)
	t.Logf("Total attempts: %d", attempts)
}

func TestReActIntegration_WithCallback(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	calcTool := NewTool("calculator", "Arithmetic").
		AddParameter("expression", "string", "Math expression", true)
	calcTool.Handler = integrationCalculator

	thoughtCount := 0
	actionCount := 0

	callback := &EnhancedReActCallback{
		OnThought: func(content string, iteration int) {
			thoughtCount++
			t.Logf("Thought [%d]: %s", iteration, content[:min(50, len(content))])
		},
		OnAction: func(tool string, args map[string]interface{}, iteration int) {
			actionCount++
			t.Logf("Action [%d]: %s", iteration, tool)
		},
		OnObservation: func(content string, iteration int) {
			t.Logf("Observation [%d]: %s", iteration, content)
		},
		OnFinal: func(answer string, iteration int) {
			t.Logf("Final [%d]: %s", iteration, answer[:min(50, len(answer))])
		},
		OnCompleted: func(result *ReActResult) {
			t.Logf("Completed: %d iterations", result.Iterations)
		},
	}

	ai := NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(3).
		WithTool(calcTool).
		WithReActCallback(callback).
		WithTimeout(30 * time.Second)

	ctx := context.Background()
	result, err := ai.Execute(ctx, "Calculate 100 / 4")

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success")
	}

	if thoughtCount == 0 {
		t.Errorf("Expected OnThought to be called")
	}

	if actionCount == 0 {
		t.Errorf("Expected OnAction to be called")
	}

	t.Logf("Callback counts: thoughts=%d, actions=%d", thoughtCount, actionCount)
}

func TestReActIntegration_Streaming(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	calcTool := NewTool("calculator", "Arithmetic").
		AddParameter("expression", "string", "Math expression", true)
	calcTool.Handler = integrationCalculator

	ai := NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(3).
		WithTool(calcTool).
		WithTimeout(30 * time.Second)

	ctx := context.Background()
	events, err := ai.StreamReAct(ctx, "What is 7 * 8?")

	if err != nil {
		t.Fatalf("StreamReAct failed: %v", err)
	}

	eventCount := 0
	hasThought := false
	hasAction := false
	hasFinal := false

	for event := range events {
		eventCount++
		t.Logf("Event [%d]: type=%s, iteration=%d", eventCount, event.Type, event.Iteration)

		switch event.Type {
		case "thought":
			hasThought = true
		case "action":
			hasAction = true
		case "final":
			hasFinal = true
		case "error":
			t.Errorf("Unexpected error event: %v", event.Error)
		}
	}

	if eventCount == 0 {
		t.Errorf("Expected events from stream")
	}

	if !hasThought {
		t.Errorf("Expected at least one thought event")
	}

	if !hasAction {
		t.Errorf("Expected at least one action event")
	}

	if !hasFinal {
		t.Errorf("Expected final event")
	}

	t.Logf("Total events: %d", eventCount)
}

func TestReActIntegration_WithExamples(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	calcTool := NewTool("calculator", "Arithmetic").
		AddParameter("expression", "string", "Math expression", true)
	calcTool.Handler = integrationCalculator

	example := ReActExample{
		Task: "What is 10 + 5?",
		Steps: []string{
			`THOUGHT: I need to add 10 and 5`,
			`ACTION: calculator(expression="10 + 5")`,
			`OBSERVATION: 15.00`,
			`FINAL: The answer is 15`,
		},
		Description: "Simple addition",
	}

	ai := NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(3).
		WithTool(calcTool).
		WithReActExamples(example).
		WithTimeout(30 * time.Second)

	ctx := context.Background()
	result, err := ai.Execute(ctx, "What is 20 + 30?")

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success with examples")
	}

	t.Logf("Result with examples: %s", result.Answer)
}

func TestReActIntegration_MaxIterations(t *testing.T) {
	apiKey := skipIfNoAPIKey(t)

	// Tool that always returns incomplete info to force iterations
	searchTool := NewTool("search", "Search for info").
		AddParameter("query", "string", "Search query", true)
	searchTool.Handler = func(argsJSON string) (string, error) {
		return "partial information", nil
	}

	ai := NewOpenAI("gpt-4o-mini", apiKey).
		WithReActMode(true).
		WithReActMaxIterations(2). // Very low limit
		WithTool(searchTool).
		WithTimeout(30 * time.Second)

	ctx := context.Background()
	result, err := ai.Execute(ctx, "Search for information about quantum computing, then molecular biology, then astrophysics")

	// Should hit max iterations
	if err == nil && result != nil {
		t.Logf("Iterations: %d (max was 2)", result.Iterations)
		if result.Iterations > 2 {
			t.Errorf("Exceeded max iterations: got %d, max 2", result.Iterations)
		}
	}
}
