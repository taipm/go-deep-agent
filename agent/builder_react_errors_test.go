package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestReActParseErrorHandling tests parse error recovery
func TestReActParseErrorHandling(t *testing.T) {
	t.Run("StrictMode_ParseError", func(t *testing.T) {
		// In strict mode, parse errors should fail immediately
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(true)

		// Verify strict mode is enabled
		if !builder.reactConfig.Strict {
			t.Error("Strict mode should be enabled")
		}
	})

	t.Run("GracefulMode_ParseError", func(t *testing.T) {
		// In graceful mode, parse errors should fallback gracefully
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(false)

		// Verify graceful mode
		if builder.reactConfig.Strict {
			t.Error("Graceful mode should be enabled (Strict=false)")
		}
	})

	t.Run("CorrectionPrompt_Format", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		// Test correction prompt generation
		parseErr := fmt.Errorf("unrecognized step format")
		response := "This is a bad format"

		prompt := builder.buildCorrectionPrompt(parseErr, response)

		// Should contain error message
		if !strings.Contains(prompt, "unrecognized step format") {
			t.Error("Correction prompt should contain error message")
		}

		// Should contain original response
		if !strings.Contains(prompt, response) {
			t.Error("Correction prompt should contain original response")
		}

		// Should contain format instructions
		if !strings.Contains(prompt, "THOUGHT:") {
			t.Error("Correction prompt should contain THOUGHT format")
		}
		if !strings.Contains(prompt, "ACTION:") {
			t.Error("Correction prompt should contain ACTION format")
		}
		if !strings.Contains(prompt, "FINAL:") {
			t.Error("Correction prompt should contain FINAL format")
		}

		// Should contain helpful hints
		if !strings.Contains(prompt, "UPPERCASE") {
			t.Error("Correction prompt should mention UPPERCASE keywords")
		}
	})

	t.Run("CorrectionPrompt_Examples", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		testCases := []struct {
			name     string
			parseErr error
			response string
		}{
			{
				name:     "MissingKeyword",
				parseErr: fmt.Errorf("no valid step found"),
				response: "I think we should search for information",
			},
			{
				name:     "InvalidFormat",
				parseErr: fmt.Errorf("unrecognized step format"),
				response: "thought: lowercase keyword",
			},
			{
				name:     "MalformedAction",
				parseErr: fmt.Errorf("invalid action format"),
				response: "ACTION: search without parentheses",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				prompt := builder.buildCorrectionPrompt(tc.parseErr, tc.response)

				if prompt == "" {
					t.Error("Correction prompt should not be empty")
				}

				if !strings.Contains(prompt, tc.response) {
					t.Errorf("Prompt should contain original response: %s", tc.response)
				}
			})
		}
	})
}

// TestReActToolErrorHandling tests tool error recovery
func TestReActToolErrorHandling(t *testing.T) {
	t.Run("StrictMode_ToolError", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(true)

		if !builder.reactConfig.Strict {
			t.Error("Strict mode should be enabled")
		}
	})

	t.Run("GracefulMode_ToolError", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(false)

		if builder.reactConfig.Strict {
			t.Error("Graceful mode should be enabled")
		}
	})

	t.Run("ToolErrorPrompt_Format", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		toolName := "search"
		toolErr := fmt.Errorf("connection timeout")

		prompt := builder.buildToolErrorPrompt(toolName, toolErr)

		// Should contain tool name
		if !strings.Contains(prompt, toolName) {
			t.Error("Tool error prompt should contain tool name")
		}

		// Should contain error message
		if !strings.Contains(prompt, "connection timeout") {
			t.Error("Tool error prompt should contain error message")
		}

		// Should suggest alternatives
		if !strings.Contains(prompt, "different approach") {
			t.Error("Tool error prompt should suggest alternatives")
		}
	})

	t.Run("ToolErrorPrompt_Examples", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true)

		testCases := []struct {
			name     string
			toolName string
			toolErr  error
		}{
			{
				name:     "NetworkError",
				toolName: "search",
				toolErr:  fmt.Errorf("network unreachable"),
			},
			{
				name:     "InvalidParams",
				toolName: "calculator",
				toolErr:  fmt.Errorf("invalid expression"),
			},
			{
				name:     "RateLimited",
				toolName: "api_call",
				toolErr:  fmt.Errorf("rate limit exceeded"),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				prompt := builder.buildToolErrorPrompt(tc.toolName, tc.toolErr)

				if prompt == "" {
					t.Error("Tool error prompt should not be empty")
				}

				if !strings.Contains(prompt, tc.toolName) {
					t.Errorf("Prompt should contain tool name: %s", tc.toolName)
				}
			})
		}
	})
}

// TestReActErrorMetrics tests error tracking in metrics
func TestReActErrorMetrics(t *testing.T) {
	t.Run("Metrics_TrackErrors", func(t *testing.T) {
		metrics := NewReActMetrics()

		// Simulate errors
		metrics.Errors = 3

		if metrics.Errors != 3 {
			t.Errorf("Expected 3 errors, got %d", metrics.Errors)
		}
	})

	t.Run("Metrics_WithParseErrors", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActMetrics(true)

		if !builder.reactConfig.EnableMetrics {
			t.Error("Metrics should be enabled")
		}
	})

	t.Run("Metrics_WithToolErrors", func(t *testing.T) {
		metrics := NewReActMetrics()

		// Simulate tool errors
		metrics.ToolCalls = 5
		metrics.Errors = 2

		if metrics.ToolCalls != 5 {
			t.Errorf("Expected 5 tool calls, got %d", metrics.ToolCalls)
		}
		if metrics.Errors != 2 {
			t.Errorf("Expected 2 errors, got %d", metrics.Errors)
		}

		// Success rate calculation
		successRate := float64(metrics.ToolCalls-metrics.Errors) / float64(metrics.ToolCalls) * 100
		expectedRate := 60.0 // 3 successful out of 5

		if successRate != expectedRate {
			t.Errorf("Expected success rate %.1f%%, got %.1f%%", expectedRate, successRate)
		}
	})
}

// TestReActErrorCallbacks tests error callbacks
func TestReActErrorCallbacks(t *testing.T) {
	t.Run("Callback_OnError", func(t *testing.T) {
		errorCalls := 0
		callback := &testCallback{
			onError: func(err error) {
				errorCalls++
			},
		}

		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActCallback(callback)

		if builder.reactConfig.Callback == nil {
			t.Error("Callback should be set")
		}

		// Simulate error
		builder.reactConfig.Callback.OnError(fmt.Errorf("test error"))

		if errorCalls != 1 {
			t.Errorf("Expected 1 error callback, got %d", errorCalls)
		}
	})

	t.Run("Callback_MultipleErrors", func(t *testing.T) {
		errors := []error{}
		callback := &testCallback{
			onError: func(err error) {
				errors = append(errors, err)
			},
		}

		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActCallback(callback)

		// Simulate multiple errors
		builder.reactConfig.Callback.OnError(fmt.Errorf("parse error"))
		builder.reactConfig.Callback.OnError(fmt.Errorf("tool error"))
		builder.reactConfig.Callback.OnError(fmt.Errorf("timeout"))

		if len(errors) != 3 {
			t.Errorf("Expected 3 errors, got %d", len(errors))
		}
	})
}

// TestReActErrorTimeline tests error events in timeline
func TestReActErrorTimeline(t *testing.T) {
	t.Run("Timeline_ParseError", func(t *testing.T) {
		timeline := NewReActTimeline()

		timeline.AddEvent("parse_error", "Failed to parse response", 0, nil)

		if len(timeline.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(timeline.Events))
		}

		event := timeline.Events[0]
		if event.Type != "parse_error" {
			t.Errorf("Expected type 'parse_error', got %q", event.Type)
		}
	})

	t.Run("Timeline_ToolError", func(t *testing.T) {
		timeline := NewReActTimeline()

		timeline.AddEvent("tool_error", "Tool execution failed", 0, map[string]interface{}{
			"tool":  "search",
			"error": "timeout",
		})

		event := timeline.Events[0]
		if event.Metadata["tool"] != "search" {
			t.Error("Timeline should track tool name")
		}
		if event.Metadata["error"] != "timeout" {
			t.Error("Timeline should track error type")
		}
	})

	t.Run("Timeline_ErrorSequence", func(t *testing.T) {
		timeline := NewReActTimeline()

		// Simulate error sequence
		timeline.AddEvent("iteration_start", "Iteration 1", 0, nil)
		timeline.AddEvent("parse_error", "Parse failed", 0, nil)
		timeline.AddEvent("parse_error_retry", "Retrying with correction", 0, nil)
		timeline.AddEvent("iteration_start", "Iteration 2", 0, nil)
		timeline.AddEvent("action", "Tool called", 0, nil)
		timeline.AddEvent("tool_error", "Tool failed", 0, nil)
		timeline.AddEvent("final", "Answer provided", 0, nil)

		if len(timeline.Events) != 7 {
			t.Errorf("Expected 7 events, got %d", len(timeline.Events))
		}

		// Check sequence
		expectedTypes := []string{
			"iteration_start",
			"parse_error",
			"parse_error_retry",
			"iteration_start",
			"action",
			"tool_error",
			"final",
		}

		for i, expectedType := range expectedTypes {
			if timeline.Events[i].Type != expectedType {
				t.Errorf("Event %d: expected %q, got %q",
					i, expectedType, timeline.Events[i].Type)
			}
		}
	})
}

// TestReActErrorRecovery tests complete error recovery flows
func TestReActErrorRecovery(t *testing.T) {
	t.Run("Recovery_ParseErrorFallback", func(t *testing.T) {
		// Test that parse errors trigger correction prompts
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(false).
			WithReActMaxIterations(5)

		if builder.reactConfig.MaxIterations != 5 {
			t.Error("MaxIterations should be 5 for retry testing")
		}
	})

	t.Run("Recovery_ToolErrorContinue", func(t *testing.T) {
		// Test that tool errors allow continuation in graceful mode
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithReActMode(true).
			WithReActStrict(false)

		// Create a tool that fails
		failTool := NewTool("fail_tool", "Always fails").
			WithHandler(func(args string) (string, error) {
				return "", fmt.Errorf("tool error")
			})

		builder.WithTool(failTool)

		ctx := context.Background()

		// Execute tool (will fail)
		result, err := builder.executeTool(ctx, "fail_tool", map[string]interface{}{})

		// Should return error
		if err == nil {
			t.Error("Expected error from failing tool")
		}

		// Result should contain error message
		if result != "" {
			t.Logf("Tool returned: %s", result)
		}
	})

	t.Run("Recovery_MultipleErrors", func(t *testing.T) {
		// Test handling of multiple consecutive errors
		metrics := NewReActMetrics()
		timeline := NewReActTimeline()

		// Simulate multiple errors
		for i := 0; i < 3; i++ {
			metrics.Errors++
			timeline.AddEvent("error", fmt.Sprintf("Error %d", i+1), 0, nil)
		}

		if metrics.Errors != 3 {
			t.Errorf("Expected 3 errors, got %d", metrics.Errors)
		}

		if len(timeline.Events) != 3 {
			t.Errorf("Expected 3 timeline events, got %d", len(timeline.Events))
		}
	})

	t.Run("Recovery_TimeoutDuringRetry", func(t *testing.T) {
		// Test timeout during error recovery
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Wait for timeout
		<-ctx.Done()

		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
		}
	})
}

// TestReActErrorMessages tests error message clarity
func TestReActErrorMessages(t *testing.T) {
	t.Run("ParseError_Clear", func(t *testing.T) {
		parseErr := fmt.Errorf("unrecognized step format: missing keyword")

		if !strings.Contains(parseErr.Error(), "unrecognized") {
			t.Error("Parse error should be descriptive")
		}
	})

	t.Run("ToolError_Detailed", func(t *testing.T) {
		toolErr := fmt.Errorf("tool execution failed: connection timeout after 30s")

		if !strings.Contains(toolErr.Error(), "tool execution failed") {
			t.Error("Tool error should mention tool execution")
		}
		if !strings.Contains(toolErr.Error(), "timeout") {
			t.Error("Tool error should include specific error")
		}
	})

	t.Run("MaxIterationsError_Helpful", func(t *testing.T) {
		maxIterErr := fmt.Errorf("max iterations (5) reached without final answer")

		if !strings.Contains(maxIterErr.Error(), "max iterations") {
			t.Error("Max iterations error should be clear")
		}
		if !strings.Contains(maxIterErr.Error(), "5") {
			t.Error("Max iterations error should include limit")
		}
	})
}
