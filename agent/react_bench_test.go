package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Benchmarks for ReAct pattern performance
// Run with: go test -bench=BenchmarkReAct -benchmem

// Fast mock tool for benchmarking
func benchCalculator(argsJSON string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	expr, _ := args["expression"].(string)
	return fmt.Sprintf("result: %s", expr), nil
}

func BenchmarkReActParser_Simple(b *testing.B) {
	output := `THOUGHT: I need to calculate this
ACTION: calculator(expression="25 + 17")
OBSERVATION: 42
FINAL: The answer is 42`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseReActSteps(output)
	}
}

func BenchmarkReActParser_Complex(b *testing.B) {
	output := `THOUGHT: This is a complex multi-line thought
that spans several lines and contains
detailed reasoning about the problem

ACTION: search(query="complex topic with lots of details")

OBSERVATION: Here is a very long observation
that contains multiple paragraphs
of information retrieved from the tool
with lots of details and context

THOUGHT: Now I need to process this information
and think about what to do next

ACTION: summarize(text="all the previous information")

OBSERVATION: Summary of everything

FINAL: Based on all the analysis, here is
a comprehensive multi-line answer
that addresses the question fully`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseReActSteps(output)
	}
}

func BenchmarkReActParser_LargeOutput(b *testing.B) {
	var sb strings.Builder

	// Generate large output with 10 iterations
	for iter := 0; iter < 10; iter++ {
		sb.WriteString(fmt.Sprintf("THOUGHT: Iteration %d thinking\n", iter))
		sb.WriteString(fmt.Sprintf("ACTION: tool_%d(data=\"value\")\n", iter))
		sb.WriteString(fmt.Sprintf("OBSERVATION: Result from iteration %d\n", iter))
	}
	sb.WriteString("FINAL: Final answer after 10 iterations")

	output := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseReActSteps(output)
	}
}

func BenchmarkToolExecution_Simple(b *testing.B) {
	calcTool := NewTool("calculator", "Fast calculator").
		AddParameter("expression", "string", "Math expression", true)
	calcTool.Handler = benchCalculator

	args := `{"expression":"10 + 20"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calcTool.Handler(args)
	}
}

func BenchmarkToolExecution_JSONParsing(b *testing.B) {
	args := `{"expression":"10 + 20","precision":"2","unit":"meters"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var parsed map[string]interface{}
		_ = json.Unmarshal([]byte(args), &parsed)
	}
}

func BenchmarkToolExecution_ComplexArgs(b *testing.B) {
	calcTool := NewTool("calculator", "Fast calculator").
		AddParameter("expression", "string", "Math expression", true).
		AddParameter("precision", "number", "Decimal precision", false).
		AddParameter("unit", "string", "Unit of measurement", false)

	calcTool.Handler = func(argsJSON string) (string, error) {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", err
		}
		expr := args["expression"].(string)
		precision := 2
		if p, ok := args["precision"].(float64); ok {
			precision = int(p)
		}
		return fmt.Sprintf("result: %s (precision: %d)", expr, precision), nil
	}

	args := `{"expression":"(10 + 20) * 3.14159","precision":4,"unit":"radians"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calcTool.Handler(args)
	}
}

func BenchmarkReActStepAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		step := ReActStep{
			Type:    "ACTION",
			Content: "Performing calculation",
			Tool:    "calculator",
			Args:    map[string]interface{}{"expression": "10 + 20"},
		}
		_ = step
	}
}

func BenchmarkReActResultAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := &ReActResult{
			Answer:     "The answer is 42",
			Iterations: 2,
			Success:    true,
			Steps: []ReActStep{
				{Type: "THOUGHT", Content: "Thinking..."},
				{Type: "ACTION", Tool: "calculator"},
				{Type: "OBSERVATION", Content: "42"},
				{Type: "FINAL", Content: "Answer is 42"},
			},
		}
		_ = result
	}
}

func BenchmarkReActExampleFormatting(b *testing.B) {
	examples := []ReActExample{
		{
			Task: "What is 10 + 5?",
			Steps: []string{
				`THOUGHT: Need to add`,
				`ACTION: calculator(expression="10 + 5")`,
				`OBSERVATION: 15`,
				`FINAL: Answer is 15`,
			},
		},
		{
			Task: "Search for Go",
			Steps: []string{
				`THOUGHT: Need to search`,
				`ACTION: search(query="golang")`,
				`OBSERVATION: Go is a language`,
				`FINAL: Found info`,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatExamples(examples)
	}
}

func BenchmarkPromptTemplateRendering(b *testing.B) {
	template := `You are a helpful assistant using ReAct.

Available Tools:
{tools}

Examples:
{examples}

Task: {task}

Think step by step using the ReAct pattern.`

	vars := map[string]string{
		"tools":    "calculator, search, summarize",
		"examples": "Example 1: ...\nExample 2: ...",
		"task":     "Calculate 25 + 17",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := template
		for key, val := range vars {
			result = strings.ReplaceAll(result, "{"+key+"}", val)
		}
		_ = result
	}
}

func BenchmarkCallbackInvocation(b *testing.B) {
	callback := &EnhancedReActCallback{
		OnThought: func(content string, iteration int) {
			_ = content
			_ = iteration
		},
		OnAction: func(tool string, args map[string]interface{}, iteration int) {
			_ = tool
			_ = args
			_ = iteration
		},
		OnObservation: func(content string, iteration int) {
			_ = content
			_ = iteration
		},
		OnFinal: func(answer string, iteration int) {
			_ = answer
			_ = iteration
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if callback.OnThought != nil {
			callback.OnThought("Thinking about the problem", 1)
		}
		if callback.OnAction != nil {
			callback.OnAction("calculator", map[string]interface{}{"expr": "10+20"}, 1)
		}
		if callback.OnObservation != nil {
			callback.OnObservation("Result: 30", 1)
		}
		if callback.OnFinal != nil {
			callback.OnFinal("The answer is 30", 1)
		}
	}
}

// Helper function for parsing (simplified version)
func parseReActSteps(output string) []ReActStep {
	var steps []ReActStep
	lines := strings.Split(output, "\n")

	var currentStep *ReActStep
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "THOUGHT:") {
			if currentStep != nil {
				steps = append(steps, *currentStep)
			}
			currentStep = &ReActStep{
				Type:    "THOUGHT",
				Content: strings.TrimSpace(strings.TrimPrefix(line, "THOUGHT:")),
			}
		} else if strings.HasPrefix(line, "ACTION:") {
			if currentStep != nil {
				steps = append(steps, *currentStep)
			}
			currentStep = &ReActStep{
				Type:    "ACTION",
				Content: strings.TrimSpace(strings.TrimPrefix(line, "ACTION:")),
			}
		} else if strings.HasPrefix(line, "OBSERVATION:") {
			if currentStep != nil {
				steps = append(steps, *currentStep)
			}
			currentStep = &ReActStep{
				Type:    "OBSERVATION",
				Content: strings.TrimSpace(strings.TrimPrefix(line, "OBSERVATION:")),
			}
		} else if strings.HasPrefix(line, "FINAL:") {
			if currentStep != nil {
				steps = append(steps, *currentStep)
			}
			currentStep = &ReActStep{
				Type:    "FINAL",
				Content: strings.TrimSpace(strings.TrimPrefix(line, "FINAL:")),
			}
		} else if currentStep != nil {
			currentStep.Content += " " + line
		}
	}

	if currentStep != nil {
		steps = append(steps, *currentStep)
	}

	return steps
}
