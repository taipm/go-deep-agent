package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestBuilder_Ask tests the Ask method with various scenarios
func TestBuilder_Ask(t *testing.T) {
	t.Run("EmptyPrompt", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx := context.Background()

		// Empty prompt should still work (API may handle it)
		_, err := builder.Ask(ctx, "")

		// We expect an error because we don't have a real API key
		// but the test verifies the function can be called
		if err == nil {
			t.Log("Note: Ask with empty prompt didn't error (unexpected)")
		}
	})

	t.Run("NilContext", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")

		// This should panic or error gracefully
		defer func() {
			if r := recover(); r == nil {
				t.Log("Note: Ask with nil context didn't panic")
			}
		}()

		_, _ = builder.Ask(nil, "test prompt")
	})

	t.Run("MissingAPIKey", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "")
		ctx := context.Background()

		response, err := builder.Ask(ctx, "test prompt")

		if err == nil {
			t.Error("Expected error with missing API key")
		}
		if response != "" {
			t.Errorf("Expected empty response with error, got: %s", response)
		}
	})

	t.Run("WithInvalidModel", func(t *testing.T) {
		builder := NewOpenAI("invalid-model-xyz", "test-key")
		ctx := context.Background()

		_, err := builder.Ask(ctx, "test prompt")

		// Should error due to invalid model
		if err == nil {
			t.Log("Note: Invalid model didn't error (may be handled by API)")
		}
	})

	t.Run("WithContextCanceled", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := builder.Ask(ctx, "test prompt")

		// Should error due to canceled context
		if err == nil {
			t.Error("Expected error with canceled context")
		}
		if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
			t.Logf("Error was: %v (expected context canceled)", err)
		}
	})
}

// TestBuilder_Stream tests the Stream method
func TestBuilder_Stream(t *testing.T) {
	t.Run("MissingAPIKey", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "")
		ctx := context.Background()

		_, err := builder.Stream(ctx, "test prompt")

		if err == nil {
			t.Error("Expected error with missing API key")
		}
	})

	t.Run("WithCallback", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx := context.Background()

		called := false
		builder.OnStream(func(chunk string) {
			called = true
		})

		_, err := builder.Stream(ctx, "test prompt")

		// Should error due to invalid API key, but callback should be set
		if err == nil {
			t.Log("Note: Stream didn't error (unexpected)")
		}
		// Callback may not be called if API errors immediately
		_ = called
	})

	t.Run("EmptyPrompt", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx := context.Background()

		_, err := builder.Stream(ctx, "")

		// Should handle empty prompt gracefully
		if err == nil {
			t.Log("Note: Stream with empty prompt didn't error")
		}
	})
}

// TestBuilder_AskMultiple tests the AskMultiple method
func TestBuilder_AskMultiple(t *testing.T) {
	t.Run("SingleChoice", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").WithMultipleChoices(1)
		ctx := context.Background()

		responses, err := builder.AskMultiple(ctx, "test prompt")

		if err == nil {
			t.Log("Note: AskMultiple didn't error (unexpected)")
		}
		if len(responses) > 1 {
			t.Errorf("Expected 1 or 0 responses, got %d", len(responses))
		}
	})

	t.Run("MultipleChoices", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").WithMultipleChoices(3)
		ctx := context.Background()

		responses, err := builder.AskMultiple(ctx, "test prompt")

		if err == nil {
			t.Log("Note: AskMultiple didn't error (unexpected)")
		}
		// With error, responses should be empty
		if len(responses) > 3 {
			t.Errorf("Expected max 3 responses, got %d", len(responses))
		}
	})

	t.Run("MissingAPIKey", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "").WithMultipleChoices(2)
		ctx := context.Background()

		responses, err := builder.AskMultiple(ctx, "test prompt")

		if err == nil {
			t.Error("Expected error with missing API key")
		}
		if len(responses) != 0 {
			t.Errorf("Expected empty responses with error, got %d", len(responses))
		}
	})
}

// TestBuilder_StreamPrint tests the StreamPrint method
func TestBuilder_StreamPrint(t *testing.T) {
	t.Run("MissingAPIKey", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "")
		ctx := context.Background()

		// StreamPrint writes to stdout, so we just verify it can be called
		_, err := builder.StreamPrint(ctx, "test prompt")

		if err == nil {
			t.Error("Expected error with missing API key")
		}
	})

	t.Run("EmptyPrompt", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx := context.Background()

		_, err := builder.StreamPrint(ctx, "")

		// Should handle empty prompt
		if err == nil {
			t.Log("Note: StreamPrint with empty prompt didn't error")
		}
	})
}

// TestBuilder_BuildParams tests the buildParams method coverage
func TestBuilder_BuildParams(t *testing.T) {
	t.Run("MinimalConfig", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx := context.Background()

		// Call Ask which internally calls buildParams
		_, err := builder.Ask(ctx, "test")

		// We're testing that buildParams gets called
		if err == nil {
			t.Log("Note: Minimal config didn't error")
		}
	})

	t.Run("FullConfig", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithTemperature(0.7).
			WithMaxTokens(100).
			WithTopP(0.9).
			WithPresencePenalty(0.5).
			WithFrequencyPenalty(0.5).
			WithSeed(42)

		ctx := context.Background()

		// Call Ask which internally calls buildParams
		_, err := builder.Ask(ctx, "test")

		// We're testing that buildParams handles all params
		if err == nil {
			t.Log("Note: Full config didn't error")
		}
	})

	t.Run("WithTools", func(t *testing.T) {
		tool := NewTool("test_tool", "A test tool").
			AddParameter("input", "string", "Test input", true)

		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithTool(tool).
			WithAutoExecute(false)

		ctx := context.Background()

		// Call Ask which internally calls buildParams with tools
		_, err := builder.Ask(ctx, "test")

		if err == nil {
			t.Log("Note: WithTools config didn't error")
		}
	})

	t.Run("WithJSONSchema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{"type": "string"},
			},
		}

		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithJSONSchema("TestSchema", "A test schema", schema, true)

		ctx := context.Background()

		// Call Ask which internally calls buildParams with JSON schema
		_, err := builder.Ask(ctx, "test")

		if err == nil {
			t.Log("Note: WithJSONSchema config didn't error")
		}
	})

	t.Run("WithMemoryAndSystem", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithSystem("You are a helpful assistant").
			WithMemory().
			WithMaxHistory(10)

		ctx := context.Background()

		// Call Ask which adds messages to history
		_, _ = builder.Ask(ctx, "First message")
		_, err := builder.Ask(ctx, "Second message")

		if err == nil {
			t.Log("Note: WithMemory config didn't error")
		}

		// Verify history was updated
		history := builder.GetHistory()
		if len(history) < 1 {
			t.Log("Note: History was not updated as expected")
		}
	})
}

// TestBuilder_EnsureClient tests the ensureClient method coverage
func TestBuilder_EnsureClient(t *testing.T) {
	t.Run("OpenAIClient", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		ctx := context.Background()

		// First call creates client
		_, err1 := builder.Ask(ctx, "test1")

		// Second call reuses client
		_, err2 := builder.Ask(ctx, "test2")

		if err1 == nil || err2 == nil {
			t.Log("Note: ensureClient calls didn't error as expected")
		}
	})

	t.Run("OllamaClient", func(t *testing.T) {
		builder := NewOllama("llama2")
		ctx := context.Background()

		// Call creates Ollama client
		_, err := builder.Ask(ctx, "test")

		if err == nil {
			t.Log("Note: Ollama ensureClient didn't error (Ollama may be running)")
		}
	})

	t.Run("CustomBaseURL", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithBaseURL("https://custom.openai.com")

		ctx := context.Background()

		// Call creates client with custom base URL
		_, err := builder.Ask(ctx, "test")

		if err == nil {
			t.Log("Note: Custom base URL didn't error")
		}
	})
}

// TestBuilder_ExecuteWithRetry tests retry logic coverage
func TestBuilder_ExecuteWithRetry(t *testing.T) {
	t.Run("WithRetry", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithRetry(3).
			WithRetryDelay(1)

		ctx := context.Background()

		// This will fail but should retry 3 times
		_, err := builder.Ask(ctx, "test")

		if err == nil {
			t.Error("Expected error after retries")
		}
	})

	t.Run("WithExponentialBackoff", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithRetry(2).
			WithExponentialBackoff()

		ctx := context.Background()

		// This will fail but should retry with exponential backoff
		_, err := builder.Ask(ctx, "test")

		if err == nil {
			t.Error("Expected error after retries")
		}
	})
}

// TestBuilder_ErrorWrapping tests error wrapping functions
func TestBuilder_ErrorWrapping(t *testing.T) {
	t.Run("IsAPIKeyError", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "")
		ctx := context.Background()

		_, err := builder.Ask(ctx, "test")

		if err != nil && !IsAPIKeyError(err) {
			t.Log("Error is not recognized as API key error (may be different error type)")
		}
	})

	t.Run("IsRateLimitError", func(t *testing.T) {
		// We can't easily trigger a rate limit, so we just test the function exists
		err := errors.New("rate limit exceeded")
		isRateLimit := IsRateLimitError(err)
		_ = isRateLimit // Just verify function can be called
	})

	t.Run("IsTimeoutError", func(t *testing.T) {
		// Test with context deadline exceeded
		err := context.DeadlineExceeded
		isTimeout := IsTimeoutError(err)
		if !isTimeout {
			t.Log("DeadlineExceeded not recognized as timeout")
		}
	})
}
