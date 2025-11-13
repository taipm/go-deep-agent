//go:build integration
// +build integration

package adapters

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

// Integration tests for OpenAI adapter with real API
// Run with: go test -tags=integration -v ./agent/adapters/
// Requires: OPENAI_API_KEY environment variable

const openaiIntegrationTestModel = "gpt-4o-mini"

// skipIfNoOpenAIAPIKey skips the test if OPENAI_API_KEY is not set
func skipIfNoOpenAIAPIKey(t *testing.T) string {
	t.Helper()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY environment variable not set")
	}
	return apiKey
}

// TestIntegrationOpenAIAdapterComplete tests Complete() with real OpenAI API
func TestIntegrationOpenAIAdapterComplete(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	tests := []struct {
		name     string
		request  *agent.CompletionRequest
		validate func(*testing.T, *agent.CompletionResponse, error)
	}{
		{
			name: "[P3] simple completion with real API",
			request: &agent.CompletionRequest{
				Model: openaiIntegrationTestModel,
				Messages: []agent.Message{
					{Role: "user", Content: "What is the capital of France? Answer in one word."},
				},
				Temperature: 0.0, // Deterministic for testing
				MaxTokens:   20,
			},
			validate: func(t *testing.T, resp *agent.CompletionResponse, err error) {
				if err != nil {
					t.Fatalf("Complete() error = %v", err)
				}
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
				if resp.Content == "" {
					t.Error("Expected non-empty content")
				}
				t.Logf("Response: %s", resp.Content)
				t.Logf("Usage: %d prompt + %d completion = %d total tokens",
					resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
			},
		},
		{
			name: "[P3] completion with system prompt",
			request: &agent.CompletionRequest{
				Model:  openaiIntegrationTestModel,
				System: "You are a helpful assistant that responds very briefly.",
				Messages: []agent.Message{
					{Role: "user", Content: "What is 2+2?"},
				},
				Temperature: 0.0,
				MaxTokens:   20,
			},
			validate: func(t *testing.T, resp *agent.CompletionResponse, err error) {
				if err != nil {
					t.Fatalf("Complete() with system prompt error = %v", err)
				}
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
				if resp.Content == "" {
					t.Error("Expected non-empty content")
				}
				t.Logf("Response with system prompt: %s", resp.Content)
			},
		},
		{
			name: "[P3] completion with conversation history",
			request: &agent.CompletionRequest{
				Model: openaiIntegrationTestModel,
				Messages: []agent.Message{
					{Role: "user", Content: "My name is Alice."},
					{Role: "assistant", Content: "Hello Alice! Nice to meet you."},
					{Role: "user", Content: "What is my name?"},
				},
				Temperature: 0.0,
				MaxTokens:   20,
			},
			validate: func(t *testing.T, resp *agent.CompletionResponse, err error) {
				if err != nil {
					t.Fatalf("Complete() with history error = %v", err)
				}
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
				if resp.Content == "" {
					t.Error("Expected non-empty content")
				}
				// Response should mention Alice
				t.Logf("Response with history: %s", resp.Content)
			},
		},
		{
			name: "[P3] completion with high temperature",
			request: &agent.CompletionRequest{
				Model: openaiIntegrationTestModel,
				Messages: []agent.Message{
					{Role: "user", Content: "Say 'Test' and nothing else."},
				},
				Temperature: 1.5,
				MaxTokens:   10,
			},
			validate: func(t *testing.T, resp *agent.CompletionResponse, err error) {
				if err != nil {
					t.Fatalf("Complete() with high temperature error = %v", err)
				}
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
				// OpenAI accepts temperatures > 1.0 (unlike Gemini)
				t.Logf("Response with temperature 1.5: %s", resp.Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: OpenAI adapter with real API key
			adapter := NewOpenAIAdapter(apiKey, "")
			// Note: OpenAI client doesn't need explicit Close()

			// WHEN: Sending completion request
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := adapter.Complete(ctx, tt.request)

			// THEN: Validate response
			tt.validate(t, resp, err)
		})
	}
}

// TestIntegrationOpenAIAdapterStream tests Stream() with real OpenAI API
func TestIntegrationOpenAIAdapterStream(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	tests := []struct {
		name     string
		request  *agent.CompletionRequest
		validate func(*testing.T, []string, *agent.CompletionResponse, error)
	}{
		{
			name: "[P3] streaming with real API",
			request: &agent.CompletionRequest{
				Model: openaiIntegrationTestModel,
				Messages: []agent.Message{
					{Role: "user", Content: "Count from 1 to 5, one number per line."},
				},
				Temperature: 0.0,
				MaxTokens:   50,
			},
			validate: func(t *testing.T, chunks []string, resp *agent.CompletionResponse, err error) {
				if err != nil {
					t.Fatalf("Stream() error = %v", err)
				}
				if len(chunks) == 0 {
					t.Error("Expected to receive chunks")
				}
				if resp == nil {
					t.Fatal("Expected non-nil final response")
				}
				if resp.Content == "" {
					t.Error("Expected non-empty accumulated content")
				}
				t.Logf("Received %d chunks", len(chunks))
				t.Logf("Final content: %s", resp.Content)
				t.Logf("Usage: %d total tokens", resp.Usage.TotalTokens)
			},
		},
		{
			name: "[P3] streaming with nil callback",
			request: &agent.CompletionRequest{
				Model: openaiIntegrationTestModel,
				Messages: []agent.Message{
					{Role: "user", Content: "Count from 1 to 3."},
				},
				Temperature: 0.0,
				MaxTokens:   20,
			},
			validate: func(t *testing.T, chunks []string, resp *agent.CompletionResponse, err error) {
				if err != nil {
					t.Fatalf("Stream() with nil callback error = %v", err)
				}
				if len(chunks) != 0 {
					t.Error("Expected no chunks captured with nil callback")
				}
				if resp == nil {
					t.Fatal("Expected non-nil final response")
				}
				if resp.Content == "" {
					t.Error("Expected non-empty content even without callback")
				}
				t.Logf("Content without callback: %s", resp.Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: OpenAI adapter with real API key
			adapter := NewOpenAIAdapter(apiKey, "")
			// Note: OpenAI client doesn't need explicit Close()

			// WHEN: Sending streaming request
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var chunks []string
			var callback func(string)

			// Only set callback for first test case
			if tt.name == "[P3] streaming with real API" {
				callback = func(chunk string) {
					chunks = append(chunks, chunk)
					t.Logf("Chunk received: %s", chunk)
				}
			}

			resp, err := adapter.Stream(ctx, tt.request, callback)

			// THEN: Validate response
			tt.validate(t, chunks, resp, err)
		})
	}
}

// TestIntegrationOpenAIAdapterContextCancellation tests context cancellation
func TestIntegrationOpenAIAdapterContextCancellation(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	t.Run("[P3] context cancellation during Complete", func(t *testing.T) {
		// GIVEN: OpenAI adapter
		adapter := NewOpenAIAdapter(apiKey, "")

		// GIVEN: Context that will be cancelled immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// WHEN: Attempting completion with cancelled context
		req := &agent.CompletionRequest{
			Model: openaiIntegrationTestModel,
			Messages: []agent.Message{
				{Role: "user", Content: "This should not complete"},
			},
		}

		_, err := adapter.Complete(ctx, req)

		// THEN: Should return context error
		if err == nil {
			t.Error("Expected error with cancelled context")
		}
		t.Logf("Context cancellation error (expected): %v", err)
	})

	t.Run("[P3] context timeout during Complete", func(t *testing.T) {
		// GIVEN: OpenAI adapter
		adapter := NewOpenAIAdapter(apiKey, "")

		// GIVEN: Context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// WHEN: Attempting completion with short timeout
		req := &agent.CompletionRequest{
			Model: openaiIntegrationTestModel,
			Messages: []agent.Message{
				{Role: "user", Content: "This should timeout"},
			},
		}

		time.Sleep(2 * time.Millisecond) // Ensure timeout
		_, err := adapter.Complete(ctx, req)

		// THEN: Should return timeout error
		if err == nil {
			t.Error("Expected timeout error")
		}
		t.Logf("Context timeout error (expected): %v", err)
	})
}

// TestIntegrationOpenAIAdapterErrorHandling tests error scenarios with real API
func TestIntegrationOpenAIAdapterErrorHandling(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	tests := []struct {
		name      string
		request   *agent.CompletionRequest
		wantError bool
	}{
		{
			name: "[P3] invalid model name",
			request: &agent.CompletionRequest{
				Model: "invalid-model-name-that-does-not-exist",
				Messages: []agent.Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantError: true,
		},
		{
			name: "[P3] empty messages",
			request: &agent.CompletionRequest{
				Model:    openaiIntegrationTestModel,
				Messages: []agent.Message{},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: OpenAI adapter
			adapter := NewOpenAIAdapter(apiKey, "")

			// WHEN: Sending invalid request
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := adapter.Complete(ctx, tt.request)

			// THEN: Validate error expectation
			if (err != nil) != tt.wantError {
				t.Errorf("Complete() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil {
				t.Logf("Error (expected): %v", err)
			}
		})
	}
}

// TestIntegrationOpenAIAdapterMaxTokens tests MaxTokens parameter
func TestIntegrationOpenAIAdapterMaxTokens(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	t.Run("[P3] respects MaxTokens limit", func(t *testing.T) {
		// GIVEN: OpenAI adapter
		adapter := NewOpenAIAdapter(apiKey, "")

		// WHEN: Requesting completion with low MaxTokens
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req := &agent.CompletionRequest{
			Model: openaiIntegrationTestModel,
			Messages: []agent.Message{
				{Role: "user", Content: "Write a very long story about a cat."},
			},
			MaxTokens:   5, // Very low limit
			Temperature: 0.0,
		}

		resp, err := adapter.Complete(ctx, req)

		// THEN: Should return response with limited tokens
		if err != nil {
			t.Fatalf("Complete() error = %v", err)
		}
		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		// Response should be short due to MaxTokens limit
		t.Logf("Response with MaxTokens=5: %s", resp.Content)
		t.Logf("Completion tokens: %d", resp.Usage.CompletionTokens)

		// Completion tokens should be close to or at limit
		if resp.Usage.CompletionTokens > 10 {
			t.Errorf("Expected completion tokens <= 10, got %d", resp.Usage.CompletionTokens)
		}
	})
}

// TestIntegrationOpenAIAdapterStop tests Stop sequences
func TestIntegrationOpenAIAdapterStop(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	t.Run("[P3] respects stop sequences", func(t *testing.T) {
		// GIVEN: OpenAI adapter
		adapter := NewOpenAIAdapter(apiKey, "")

		// WHEN: Requesting completion with stop sequence
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req := &agent.CompletionRequest{
			Model: openaiIntegrationTestModel,
			Messages: []agent.Message{
				{Role: "user", Content: "Count from 1 to 10. Say STOP when done."},
			},
			Stop:        []string{"STOP"},
			Temperature: 0.0,
			MaxTokens:   100,
		}

		resp, err := adapter.Complete(ctx, req)

		// THEN: Should stop when encountering stop sequence
		if err != nil {
			t.Fatalf("Complete() error = %v", err)
		}
		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		t.Logf("Response with stop sequence: %s", resp.Content)
		t.Logf("Finish reason: %s", resp.FinishReason)
	})
}

// TestIntegrationOpenAIAdapterConcurrent tests concurrent requests
func TestIntegrationOpenAIAdapterConcurrent(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	t.Run("[P3] handles concurrent requests", func(t *testing.T) {
		// GIVEN: OpenAI adapter
		adapter := NewOpenAIAdapter(apiKey, "")

		// WHEN: Sending multiple concurrent requests
		concurrency := 3
		done := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(n int) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				req := &agent.CompletionRequest{
					Model: openaiIntegrationTestModel,
					Messages: []agent.Message{
						{Role: "user", Content: "Say 'Hello' and nothing else."},
					},
					Temperature: 0.0,
					MaxTokens:   10,
				}

				resp, err := adapter.Complete(ctx, req)
				if err != nil {
					done <- err
					return
				}
				if resp == nil || resp.Content == "" {
					done <- err
					return
				}
				t.Logf("Concurrent request %d completed: %s", n, resp.Content)
				done <- nil
			}(i)
		}

		// THEN: All requests should complete successfully
		for i := 0; i < concurrency; i++ {
			if err := <-done; err != nil {
				t.Errorf("Concurrent request failed: %v", err)
			}
		}
	})
}

// TestIntegrationOpenAIAdapterWithOllama tests OpenAI-compatible endpoints (Ollama)
func TestIntegrationOpenAIAdapterWithOllama(t *testing.T) {
	// Only run if OLLAMA_BASE_URL is set
	ollamaURL := os.Getenv("OLLAMA_BASE_URL")
	if ollamaURL == "" {
		t.Skip("Skipping Ollama test: OLLAMA_BASE_URL environment variable not set")
	}

	t.Run("[P3] works with Ollama (OpenAI-compatible API)", func(t *testing.T) {
		// GIVEN: OpenAI adapter configured for Ollama
		adapter := NewOpenAIAdapter("ollama", ollamaURL)

		// WHEN: Sending completion request
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req := &agent.CompletionRequest{
			Model: "qwen3:1.7b", // Ollama reasoning model
			Messages: []agent.Message{
				{Role: "user", Content: "Say hello"},
			},
			Temperature: 0.7,
			MaxTokens:   50,
		}

		resp, err := adapter.Complete(ctx, req)

		// THEN: Should work with OpenAI-compatible API
		if err != nil {
			t.Fatalf("Complete() with Ollama error = %v", err)
		}
		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		// Debug: Log full response
		t.Logf("Ollama response ID: %s", resp.ID)
		t.Logf("Ollama response Model: %s", resp.Model)
		t.Logf("Ollama response Content: '%s'", resp.Content)
		t.Logf("Ollama response FinishReason: %s", resp.FinishReason)
		t.Logf("Ollama response Usage: %+v", resp.Usage)

		if resp.Content == "" {
			t.Error("Expected non-empty content from Ollama")
		}
	})
}

// TestIntegrationOpenAIAdapterSeed tests deterministic generation with seed
func TestIntegrationOpenAIAdapterSeed(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	t.Run("[P3] seed produces deterministic results", func(t *testing.T) {
		// GIVEN: OpenAI adapter
		adapter := NewOpenAIAdapter(apiKey, "")

		// WHEN: Sending same request with same seed twice
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req := &agent.CompletionRequest{
			Model: openaiIntegrationTestModel,
			Messages: []agent.Message{
				{Role: "user", Content: "Generate a random number between 1 and 100."},
			},
			Temperature: 1.0, // High temperature but fixed seed
			Seed:        42,
			MaxTokens:   20,
		}

		// First request
		resp1, err := adapter.Complete(ctx, req)
		if err != nil {
			t.Fatalf("Complete() first request error = %v", err)
		}

		// Second request with same seed
		resp2, err := adapter.Complete(ctx, req)
		if err != nil {
			t.Fatalf("Complete() second request error = %v", err)
		}

		// THEN: Responses should be identical (or very similar) with same seed
		t.Logf("First response: %s", resp1.Content)
		t.Logf("Second response: %s", resp2.Content)
		t.Log("Note: OpenAI seed may not guarantee 100% determinism but should be similar")
	})
}

// TestIntegrationOpenAIAdapterResponseFormat tests structured output
func TestIntegrationOpenAIAdapterResponseFormat(t *testing.T) {
	apiKey := skipIfNoOpenAIAPIKey(t)

	t.Run("[P3] JSON response format", func(t *testing.T) {
		// GIVEN: OpenAI adapter
		adapter := NewOpenAIAdapter(apiKey, "")

		// WHEN: Requesting JSON format response
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req := &agent.CompletionRequest{
			Model: openaiIntegrationTestModel,
			Messages: []agent.Message{
				{Role: "user", Content: "Return a JSON object with name='Alice' and age=30"},
			},
			Temperature: 0.0,
			MaxTokens:   50,
			// Note: ResponseFormat would be set here if implemented
			// ResponseFormat: map[string]interface{}{"type": "json_object"},
		}

		resp, err := adapter.Complete(ctx, req)

		// THEN: Should return valid response (format may or may not be JSON without ResponseFormat set)
		if err != nil {
			t.Fatalf("Complete() error = %v", err)
		}
		if resp == nil {
			t.Fatal("Expected non-nil response")
		}
		t.Logf("Response: %s", resp.Content)
		t.Log("Note: Set ResponseFormat in request for guaranteed JSON output")
	})
}
