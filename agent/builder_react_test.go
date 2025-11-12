package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestReActBasicLoop tests a simple single-iteration ReAct execution
func TestReActBasicLoop(t *testing.T) {
	// Skip if no API key (this is a mock test, but uses real builder structure)
	t.Run("DisabledMode", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx := context.Background()

		// Without enabling ReAct, Execute should fallback to Ask
		result, err := builder.Execute(ctx, "Hello")

		// Expect error due to missing real API key, but test the flow
		if err == nil {
			t.Log("Unexpected success (no real API key)")
		}

		// Result should be created even on error
		if result == nil {
			t.Error("Expected result object even on error")
		}
	})

	t.Run("EnabledWithoutTools", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMaxIterations(3)

		ctx := context.Background()

		// Should attempt ReAct execution even without tools
		_, err := builder.Execute(ctx, "What is 2+2?")

		// Expect error due to missing real API key
		if err == nil {
			t.Log("Unexpected success (no real API key)")
		}

		// Verify ReAct mode was enabled
		if builder.reactConfig == nil || !builder.reactConfig.Enabled {
			t.Error("ReAct should be enabled")
		}
	})

	t.Run("ConfigDefaults", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		// Check default config values
		if builder.reactConfig.MaxIterations != DefaultReActMaxIterations {
			t.Errorf("Expected MaxIterations=%d, got %d",
				DefaultReActMaxIterations, builder.reactConfig.MaxIterations)
		}

		if builder.reactConfig.Timeout != DefaultReActTimeout {
			t.Errorf("Expected Timeout=%v, got %v",
				DefaultReActTimeout, builder.reactConfig.Timeout)
		}

		if builder.reactConfig.Strict != DefaultReActStrict {
			t.Errorf("Expected Strict=%v, got %v",
				DefaultReActStrict, builder.reactConfig.Strict)
		}
	})
}

// TestReActConfiguration tests various configuration options
func TestReActConfiguration(t *testing.T) {
	t.Run("WithReActMaxIterations", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMaxIterations(10)

		if builder.reactConfig.MaxIterations != 10 {
			t.Errorf("Expected MaxIterations=10, got %d", builder.reactConfig.MaxIterations)
		}
	})

	t.Run("WithReActTimeout", func(t *testing.T) {
		timeout := 2 * time.Minute
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActTimeout(timeout)

		if builder.reactConfig.Timeout != timeout {
			t.Errorf("Expected Timeout=%v, got %v", timeout, builder.reactConfig.Timeout)
		}
	})

	t.Run("WithReActStrict", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(true)

		if !builder.reactConfig.Strict {
			t.Error("Expected Strict=true")
		}
	})

	t.Run("WithReActSystemPrompt", func(t *testing.T) {
		customPrompt := "Custom ReAct prompt"
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActSystemPrompt(customPrompt)

		if builder.reactConfig.SystemPrompt != customPrompt {
			t.Errorf("Expected custom prompt, got %q", builder.reactConfig.SystemPrompt)
		}
	})

	t.Run("WithReActMetrics", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMetrics(true)

		if !builder.reactConfig.EnableMetrics {
			t.Error("Expected EnableMetrics=true")
		}
	})

	t.Run("WithReActTimeline", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActTimeline(true)

		if !builder.reactConfig.EnableTimeline {
			t.Error("Expected EnableTimeline=true")
		}
	})

	t.Run("WithReActComplexity_Simple", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActComplexity(ReActTaskSimple)

		if builder.reactConfig.MaxIterations != ReActSimpleMaxIterations {
			t.Errorf("Expected MaxIterations=%d, got %d", ReActSimpleMaxIterations, builder.reactConfig.MaxIterations)
		}
		if builder.reactConfig.Timeout != ReActSimpleTimeout {
			t.Errorf("Expected Timeout=%v, got %v", ReActSimpleTimeout, builder.reactConfig.Timeout)
		}
		if !builder.reactConfig.EnableAutoFallback {
			t.Error("Expected EnableAutoFallback=true for simple tasks")
		}
		if !builder.reactConfig.ForceFinalAnswerAtMax {
			t.Error("Expected ForceFinalAnswerAtMax=true for simple tasks")
		}
		if !builder.reactConfig.EnableIterationReminders {
			t.Error("Expected EnableIterationReminders=true for simple tasks")
		}
	})

	t.Run("WithReActComplexity_Medium", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActComplexity(ReActTaskMedium)

		if builder.reactConfig.MaxIterations != ReActMediumMaxIterations {
			t.Errorf("Expected MaxIterations=%d, got %d", ReActMediumMaxIterations, builder.reactConfig.MaxIterations)
		}
		if builder.reactConfig.Timeout != ReActMediumTimeout {
			t.Errorf("Expected Timeout=%v, got %v", ReActMediumTimeout, builder.reactConfig.Timeout)
		}
		if !builder.reactConfig.EnableAutoFallback {
			t.Error("Expected EnableAutoFallback=true for medium tasks")
		}
	})

	t.Run("WithReActComplexity_Complex", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActComplexity(ReActTaskComplex)

		if builder.reactConfig.MaxIterations != ReActComplexMaxIterations {
			t.Errorf("Expected MaxIterations=%d, got %d", ReActComplexMaxIterations, builder.reactConfig.MaxIterations)
		}
		if builder.reactConfig.Timeout != ReActComplexTimeout {
			t.Errorf("Expected Timeout=%v, got %v", ReActComplexTimeout, builder.reactConfig.Timeout)
		}
		if !builder.reactConfig.EnableAutoFallback {
			t.Error("Expected EnableAutoFallback=true for complex tasks")
		}
	})

	t.Run("WithReActAutoFallback", func(t *testing.T) {
		// Test enabled (default)
		builderEnabled := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActAutoFallback(true)

		if !builderEnabled.reactConfig.EnableAutoFallback {
			t.Error("Expected EnableAutoFallback=true when explicitly enabled")
		}

		// Test disabled
		builderDisabled := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActAutoFallback(false)

		if builderDisabled.reactConfig.EnableAutoFallback {
			t.Error("Expected EnableAutoFallback=false when disabled")
		}
	})

	t.Run("WithReActIterationReminders", func(t *testing.T) {
		// Test enabled (default)
		builderEnabled := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActIterationReminders(true)

		if !builderEnabled.reactConfig.EnableIterationReminders {
			t.Error("Expected EnableIterationReminders=true when explicitly enabled")
		}

		// Test disabled
		builderDisabled := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActIterationReminders(false)

		if builderDisabled.reactConfig.EnableIterationReminders {
			t.Error("Expected EnableIterationReminders=false when disabled")
		}
	})

	t.Run("WithReActForceFinalAnswer", func(t *testing.T) {
		// Test enabled (default)
		builderEnabled := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActForceFinalAnswer(true)

		if !builderEnabled.reactConfig.ForceFinalAnswerAtMax {
			t.Error("Expected ForceFinalAnswerAtMax=true when explicitly enabled")
		}

		// Test disabled
		builderDisabled := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActForceFinalAnswer(false)

		if builderDisabled.reactConfig.ForceFinalAnswerAtMax {
			t.Error("Expected ForceFinalAnswerAtMax=false when disabled")
		}
	})

	t.Run("MethodChaining", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMaxIterations(7).
			WithReActTimeout(90 * time.Second).
			WithReActStrict(false).
			WithReActMetrics(true).
			WithReActTimeline(true)

		if builder.reactConfig.MaxIterations != 7 {
			t.Error("Chaining failed for MaxIterations")
		}
		if builder.reactConfig.Timeout != 90*time.Second {
			t.Error("Chaining failed for Timeout")
		}
		if !builder.reactConfig.EnableMetrics {
			t.Error("Chaining failed for EnableMetrics")
		}
		if !builder.reactConfig.EnableTimeline {
			t.Error("Chaining failed for EnableTimeline")
		}
	})
}

// TestReActSystemPromptGeneration tests the system prompt builder
func TestReActSystemPromptGeneration(t *testing.T) {
	t.Run("WithoutTools", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		prompt := builder.buildReActSystemPrompt()

		// Should contain format instructions
		if !strings.Contains(prompt, "THOUGHT:") {
			t.Error("Prompt should contain THOUGHT format")
		}
		if !strings.Contains(prompt, "ACTION:") {
			t.Error("Prompt should contain ACTION format")
		}
		if !strings.Contains(prompt, "FINAL:") {
			t.Error("Prompt should contain FINAL format")
		}

		// Should indicate no tools
		if !strings.Contains(prompt, "No tools available") {
			t.Error("Prompt should indicate no tools available")
		}
	})

	t.Run("WithTools", func(t *testing.T) {
		searchTool := NewTool("search", "Search the web for information").
			AddParameter("query", "string", "Search query", true)

		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithTool(searchTool)

		prompt := builder.buildReActSystemPrompt()

		// Should list the tool
		if !strings.Contains(prompt, "search") {
			t.Error("Prompt should contain search tool")
		}
		if !strings.Contains(prompt, "Search the web") {
			t.Error("Prompt should contain tool description")
		}
	})

	t.Run("WithMultipleTools", func(t *testing.T) {
		searchTool := NewTool("search", "Search the web")
		calcTool := NewTool("calculator", "Calculate expressions")

		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithTool(searchTool).
			WithTool(calcTool)

		prompt := builder.buildReActSystemPrompt()

		// Should list both tools
		if !strings.Contains(prompt, "search") {
			t.Error("Prompt should contain search tool")
		}
		if !strings.Contains(prompt, "calculator") {
			t.Error("Prompt should contain calculator tool")
		}
	})
}

// TestReActCallback tests the callback system
func TestReActCallback(t *testing.T) {
	t.Run("CallbackInterface", func(t *testing.T) {
		// Mock callback
		type mockCallback struct {
			stepCalls     int
			toolCallCalls int
			errorCalls    int
			completeCalls int
		}

		var mc mockCallback
		callback := &testCallback{
			onStep: func(step ReActStep) {
				mc.stepCalls++
			},
			onToolCall: func(tool string, args map[string]interface{}) {
				mc.toolCallCalls++
			},
			onError: func(err error) {
				mc.errorCalls++
			},
			onComplete: func(result *ReActResult) {
				mc.completeCalls++
			},
		}

		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActCallback(callback)

		if builder.reactConfig.Callback == nil {
			t.Error("Callback should be set")
		}
	})
}

// testCallback is a helper for testing callbacks
type testCallback struct {
	onStep     func(step ReActStep)
	onToolCall func(tool string, args map[string]interface{})
	onError    func(err error)
	onComplete func(result *ReActResult)
}

func (tc *testCallback) OnStep(step ReActStep) {
	if tc.onStep != nil {
		tc.onStep(step)
	}
}

func (tc *testCallback) OnToolCall(tool string, args map[string]interface{}) {
	if tc.onToolCall != nil {
		tc.onToolCall(tool, args)
	}
}

func (tc *testCallback) OnError(err error) {
	if tc.onError != nil {
		tc.onError(err)
	}
}

func (tc *testCallback) OnComplete(result *ReActResult) {
	if tc.onComplete != nil {
		tc.onComplete(result)
	}
}

// TestReActMetricsAndTimeline tests metrics and timeline tracking
func TestReActMetricsAndTimeline(t *testing.T) {
	t.Run("MetricsInitialization", func(t *testing.T) {
		metrics := NewReActMetrics()

		if metrics.TotalIterations != 0 {
			t.Error("TotalIterations should start at 0")
		}
		if metrics.ToolCalls != 0 {
			t.Error("ToolCalls should start at 0")
		}
		if metrics.Errors != 0 {
			t.Error("Errors should start at 0")
		}
		if metrics.StartTime.IsZero() {
			t.Error("StartTime should be set")
		}
	})

	t.Run("MetricsFinalize", func(t *testing.T) {
		metrics := NewReActMetrics()
		time.Sleep(10 * time.Millisecond) // Ensure some time passes

		metrics.Finalize()

		if metrics.EndTime.IsZero() {
			t.Error("EndTime should be set after Finalize")
		}
		if metrics.Duration == 0 {
			t.Error("Duration should be calculated")
		}
		if metrics.Duration < 10*time.Millisecond {
			t.Error("Duration should be at least 10ms")
		}
	})

	t.Run("TimelineInitialization", func(t *testing.T) {
		timeline := NewReActTimeline()

		if timeline.Events == nil {
			t.Error("Events should be initialized")
		}
		if len(timeline.Events) != 0 {
			t.Error("Events should start empty")
		}
	})

	t.Run("TimelineAddEvent", func(t *testing.T) {
		timeline := NewReActTimeline()

		timeline.AddEvent("test", "Test event", 100*time.Millisecond, nil)

		if len(timeline.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(timeline.Events))
		}

		event := timeline.Events[0]
		if event.Type != "test" {
			t.Errorf("Expected type 'test', got %q", event.Type)
		}
		if event.Content != "Test event" {
			t.Errorf("Expected content 'Test event', got %q", event.Content)
		}
		if event.Duration != 100*time.Millisecond {
			t.Errorf("Expected duration 100ms, got %v", event.Duration)
		}
	})

	t.Run("TimelineWithMetadata", func(t *testing.T) {
		timeline := NewReActTimeline()

		metadata := map[string]interface{}{
			"tool":  "search",
			"count": 42,
		}
		timeline.AddEvent("action", "Tool called", 0, metadata)

		event := timeline.Events[0]
		if event.Metadata == nil {
			t.Error("Metadata should be set")
		}
		if event.Metadata["tool"] != "search" {
			t.Error("Metadata should contain tool name")
		}
		if event.Metadata["count"] != 42 {
			t.Error("Metadata should contain count")
		}
	})
}

// TestReActResultStructure tests the result structure
func TestReActResultStructure(t *testing.T) {
	t.Run("EmptyResult", func(t *testing.T) {
		result := &ReActResult{
			Steps: []ReActStep{},
		}

		if result.Answer != "" {
			t.Error("Answer should be empty")
		}
		if result.Success {
			t.Error("Success should be false by default")
		}
		if result.Iterations != 0 {
			t.Error("Iterations should be 0")
		}
		if len(result.Steps) != 0 {
			t.Error("Steps should be empty")
		}
	})

	t.Run("WithSteps", func(t *testing.T) {
		result := &ReActResult{
			Steps: []ReActStep{
				{Type: StepTypeThought, Content: "Thinking..."},
				{Type: StepTypeAction, Content: "search(query='test')", Tool: "search"},
				{Type: StepTypeObservation, Content: "Found results"},
				{Type: StepTypeFinal, Content: "Done"},
			},
			Answer:     "Done",
			Success:    true,
			Iterations: 1,
		}

		if len(result.Steps) != 4 {
			t.Errorf("Expected 4 steps, got %d", len(result.Steps))
		}
		if result.Answer != "Done" {
			t.Error("Answer should be set")
		}
		if !result.Success {
			t.Error("Success should be true")
		}
	})
}

// TestExecuteToolFunction tests the executeTool helper
func TestExecuteToolFunction(t *testing.T) {
	t.Run("ToolNotFound", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx := context.Background()

		_, err := builder.executeTool(ctx, "nonexistent", map[string]interface{}{})

		if err == nil {
			t.Error("Expected error for nonexistent tool")
		}
		if !strings.Contains(err.Error(), "tool not found") {
			t.Errorf("Expected 'tool not found' error, got: %v", err)
		}
	})

	t.Run("ToolWithoutHandler", func(t *testing.T) {
		toolWithoutHandler := &Tool{
			Name:        "test",
			Description: "Test tool",
			Handler:     nil,
		}

		builder := NewOpenAI("gpt-4o-mini", "test-key")
		builder.tools = []*Tool{toolWithoutHandler}
		ctx := context.Background()

		_, err := builder.executeTool(ctx, "test", map[string]interface{}{})

		if err == nil {
			t.Error("Expected error for tool without handler")
		}
		if !strings.Contains(err.Error(), "no handler") {
			t.Errorf("Expected 'no handler' error, got: %v", err)
		}
	})

	t.Run("ToolExecutionSuccess", func(t *testing.T) {
		successTool := NewTool("calculator", "Calculate").
			WithHandler(func(args string) (string, error) {
				return "42", nil
			})

		builder := NewOpenAI("gpt-4o-mini", "test-key")
		builder.tools = []*Tool{successTool}
		ctx := context.Background()

		result, err := builder.executeTool(ctx, "calculator", map[string]interface{}{
			"expression": "2+2",
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "42" {
			t.Errorf("Expected '42', got %q", result)
		}
	})

	t.Run("ToolExecutionError", func(t *testing.T) {
		errorTool := NewTool("error_tool", "Fails").
			WithHandler(func(args string) (string, error) {
				return "", fmt.Errorf("tool failed")
			})

		builder := NewOpenAI("gpt-4o-mini", "test-key")
		builder.tools = []*Tool{errorTool}
		ctx := context.Background()

		_, err := builder.executeTool(ctx, "error_tool", map[string]interface{}{})

		if err == nil {
			t.Error("Expected error from tool execution")
		}
		if !strings.Contains(err.Error(), "tool execution failed") {
			t.Errorf("Expected wrapped error, got: %v", err)
		}
	})

	t.Run("ToolWithArguments", func(t *testing.T) {
		argsTool := NewTool("echo", "Echo arguments").
			WithHandler(func(args string) (string, error) {
				// Verify JSON was received
				if !strings.Contains(args, "message") {
					t.Errorf("Expected JSON with 'message', got: %s", args)
				}
				return args, nil
			})

		builder := NewOpenAI("gpt-4o-mini", "test-key")
		builder.tools = []*Tool{argsTool}
		ctx := context.Background()

		result, err := builder.executeTool(ctx, "echo", map[string]interface{}{
			"message": "hello",
			"count":   3,
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !strings.Contains(result, "hello") {
			t.Errorf("Expected result to contain 'hello', got: %s", result)
		}
	})
}

// TestReActStepTypes tests step type constants
func TestReActStepTypes(t *testing.T) {
	types := []string{
		StepTypeThought,
		StepTypeAction,
		StepTypeObservation,
		StepTypeFinal,
	}

	expectedTypes := []string{"THOUGHT", "ACTION", "OBSERVATION", "FINAL"}

	for i, expected := range expectedTypes {
		if types[i] != expected {
			t.Errorf("Step type %d: expected %q, got %q", i, expected, types[i])
		}
	}
}

// TestReActMaxIterations tests iteration limit enforcement
func TestReActMaxIterations(t *testing.T) {
	t.Run("MaxIterations_1", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMaxIterations(1)

		if builder.reactConfig.MaxIterations != 1 {
			t.Errorf("Expected MaxIterations=1, got %d", builder.reactConfig.MaxIterations)
		}
	})

	t.Run("MaxIterations_3", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMaxIterations(3)

		if builder.reactConfig.MaxIterations != 3 {
			t.Errorf("Expected MaxIterations=3, got %d", builder.reactConfig.MaxIterations)
		}
	})

	t.Run("MaxIterations_5", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMaxIterations(5)

		if builder.reactConfig.MaxIterations != 5 {
			t.Errorf("Expected MaxIterations=5, got %d", builder.reactConfig.MaxIterations)
		}
	})

	t.Run("MaxIterations_10", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMaxIterations(10)

		if builder.reactConfig.MaxIterations != 10 {
			t.Errorf("Expected MaxIterations=10, got %d", builder.reactConfig.MaxIterations)
		}
	})

	t.Run("MaxIterations_Default", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		if builder.reactConfig.MaxIterations != DefaultReActMaxIterations {
			t.Errorf("Expected default MaxIterations=%d, got %d",
				DefaultReActMaxIterations, builder.reactConfig.MaxIterations)
		}
	})
}

// TestReActTimeout tests timeout configuration
func TestReActTimeout(t *testing.T) {
	t.Run("Timeout_10s", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActTimeout(10 * time.Second)

		if builder.reactConfig.Timeout != 10*time.Second {
			t.Errorf("Expected Timeout=10s, got %v", builder.reactConfig.Timeout)
		}
	})

	t.Run("Timeout_30s", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActTimeout(30 * time.Second)

		if builder.reactConfig.Timeout != 30*time.Second {
			t.Errorf("Expected Timeout=30s, got %v", builder.reactConfig.Timeout)
		}
	})

	t.Run("Timeout_1min", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActTimeout(1 * time.Minute)

		if builder.reactConfig.Timeout != 1*time.Minute {
			t.Errorf("Expected Timeout=1min, got %v", builder.reactConfig.Timeout)
		}
	})

	t.Run("Timeout_Default", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		if builder.reactConfig.Timeout != DefaultReActTimeout {
			t.Errorf("Expected default Timeout=%v, got %v",
				DefaultReActTimeout, builder.reactConfig.Timeout)
		}
	})
}

// TestReActStrictMode tests strict vs graceful error handling
func TestReActStrictMode(t *testing.T) {
	t.Run("StrictMode_Enabled", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(true)

		if !builder.reactConfig.Strict {
			t.Error("Expected Strict=true")
		}
	})

	t.Run("StrictMode_Disabled", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(false)

		if builder.reactConfig.Strict {
			t.Error("Expected Strict=false")
		}
	})

	t.Run("StrictMode_Default", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		if builder.reactConfig.Strict != DefaultReActStrict {
			t.Errorf("Expected default Strict=%v, got %v",
				DefaultReActStrict, builder.reactConfig.Strict)
		}
	})
}

// TestReActObservabilityConfig tests metrics and timeline configuration
func TestReActObservabilityConfig(t *testing.T) {
	t.Run("EnableMetrics", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMetrics(true)

		if !builder.reactConfig.EnableMetrics {
			t.Error("Expected EnableMetrics=true")
		}
	})

	t.Run("DisableMetrics", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMetrics(false)

		if builder.reactConfig.EnableMetrics {
			t.Error("Expected EnableMetrics=false")
		}
	})

	t.Run("EnableTimeline", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActTimeline(true)

		if !builder.reactConfig.EnableTimeline {
			t.Error("Expected EnableTimeline=true")
		}
	})

	t.Run("DisableTimeline", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActTimeline(false)

		if builder.reactConfig.EnableTimeline {
			t.Error("Expected EnableTimeline=false")
		}
	})

	t.Run("BothEnabled", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMetrics(true).
			WithReActTimeline(true)

		if !builder.reactConfig.EnableMetrics {
			t.Error("Expected EnableMetrics=true")
		}
		if !builder.reactConfig.EnableTimeline {
			t.Error("Expected EnableTimeline=true")
		}
	})
}

// TestParseReActStep_MultiStepScenarios tests complex multi-step parsing
func TestParseReActStep_MultiStepScenarios(t *testing.T) {
	t.Run("ThoughtOnly", func(t *testing.T) {
		// Each step should be parsed individually
		text := `THOUGHT: I need to search for information`

		stepType, content, _, _, err := parseReActStep(text)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if stepType != StepTypeThought {
			t.Errorf("Expected THOUGHT step, got %q", stepType)
		}
		if !strings.Contains(content, "search for information") {
			t.Errorf("Expected content about searching, got %q", content)
		}
	})

	t.Run("ActionOnly", func(t *testing.T) {
		// Parse single ACTION step
		text := `ACTION: calculator(expression="2+2")`

		stepType, _, tool, _, err := parseReActStep(text)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if stepType != StepTypeAction {
			t.Errorf("Expected ACTION step, got %q", stepType)
		}
		if tool != "calculator" {
			t.Errorf("Expected calculator tool, got %q", tool)
		}
	})

	t.Run("CompleteReActCycle", func(t *testing.T) {
		// Test complete THOUGHT → ACTION → OBSERVATION → FINAL
		steps := []string{
			`THOUGHT: I need to calculate something`,
			`ACTION: calculator(expression="5*5")`,
			`OBSERVATION: Result is 25`,
			`FINAL: The answer is 25`,
		}

		expectedTypes := []string{
			StepTypeThought,
			StepTypeAction,
			StepTypeObservation,
			StepTypeFinal,
		}

		for i, step := range steps {
			stepType, _, _, _, err := parseReActStep(step)
			if err != nil {
				t.Errorf("Step %d error: %v", i, err)
			}
			if stepType != expectedTypes[i] {
				t.Errorf("Step %d: expected %q, got %q", i, expectedTypes[i], stepType)
			}
		}
	})

	t.Run("MultipleThoughts", func(t *testing.T) {
		// Test multiple THOUGHT steps
		thoughts := []string{
			`THOUGHT: First, I need to understand the question`,
			`THOUGHT: Then I should break it down into steps`,
			`THOUGHT: Finally, I can execute the plan`,
		}

		for i, thought := range thoughts {
			stepType, _, _, _, err := parseReActStep(thought)
			if err != nil {
				t.Errorf("Thought %d error: %v", i, err)
			}
			if stepType != StepTypeThought {
				t.Errorf("Thought %d: expected THOUGHT, got %q", i, stepType)
			}
		}
	})

	t.Run("ToolChaining", func(t *testing.T) {
		// Test sequence of different tool calls
		actions := []struct {
			text string
			tool string
		}{
			{`ACTION: search(query="weather")`, "search"},
			{`ACTION: calculator(expression="temp_c*1.8+32")`, "calculator"},
			{`ACTION: formatter(text="The result is {}")`, "formatter"},
		}

		for i, action := range actions {
			stepType, _, tool, _, err := parseReActStep(action.text)
			if err != nil {
				t.Errorf("Action %d error: %v", i, err)
			}
			if stepType != StepTypeAction {
				t.Errorf("Action %d: expected ACTION, got %q", i, stepType)
			}
			if tool != action.tool {
				t.Errorf("Action %d: expected tool %q, got %q", i, action.tool, tool)
			}
		}
	})

	t.Run("ObservationVariations", func(t *testing.T) {
		// Test different OBSERVATION formats
		observations := []string{
			`OBSERVATION: Success`,
			`OBSERVATION: Found 10 results`,
			`OBSERVATION: Error: timeout`,
			`OBSERVATION: {"status": "ok", "data": [1,2,3]}`,
		}

		for i, obs := range observations {
			stepType, _, _, _, err := parseReActStep(obs)
			if err != nil {
				t.Errorf("Observation %d error: %v", i, err)
			}
			if stepType != StepTypeObservation {
				t.Errorf("Observation %d: expected OBSERVATION, got %q", i, stepType)
			}
		}
	})

	t.Run("FinalAnswerVariations", func(t *testing.T) {
		// Test different FINAL answer formats
		finals := []string{
			`FINAL: The answer is 42`,
			`FINAL: I couldn't find an answer`,
			`FINAL: Based on my research, the conclusion is...`,
			`FINAL: Error: unable to complete task`,
		}

		for i, final := range finals {
			stepType, content, _, _, err := parseReActStep(final)
			if err != nil {
				t.Errorf("Final %d error: %v", i, err)
			}
			if stepType != StepTypeFinal {
				t.Errorf("Final %d: expected FINAL, got %q", i, stepType)
			}
			if content == "" {
				t.Errorf("Final %d: content should not be empty", i)
			}
		}
	})
}
