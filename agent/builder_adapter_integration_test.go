package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// TestBuilderAdapterIntegration tests the integration between Builder and adapters
// This is critical to ensure the adapter integration bug fix works correctly
func TestBuilderAdapterIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("Builder with adapter - basic completion", func(t *testing.T) {
		// Create a mock adapter
		adapter := &mockTestAdapter{
			responses: []string{"Hello from adapter!"},
		}

		// Create builder with adapter
		builder := NewWithAdapter("test-model", adapter).
			WithSystem("You are a helpful assistant")

		// Test basic completion
		response, err := builder.Ask(ctx, "Hello")
		if err != nil {
			t.Fatalf("Ask should not fail with adapter: %v", err)
		}

		if response != "Hello from adapter!" {
			t.Errorf("Expected 'Hello from adapter!', got: %s", response)
		}

		// Verify adapter was called correctly
		if !adapter.wasCalled {
			t.Error("Adapter should have been called")
		}
	})

	t.Run("Builder with adapter - streaming", func(t *testing.T) {
		adapter := &mockTestAdapter{
			streamResponses: []string{"Hello ", "from ", "stream!"},
		}

		builder := NewWithAdapter("test-model", adapter)

		var chunks []string
		response, err := builder.OnStream(func(chunk string) {
			chunks = append(chunks, chunk)
		}).Stream(ctx, "Hello")

		if err != nil {
			t.Fatalf("Stream should not fail with adapter: %v", err)
		}

		// Verify chunks were received
		if len(chunks) != 3 {
			t.Errorf("Expected 3 chunks, got %d: %v", len(chunks), chunks)
		}

		// Verify final response
		expectedResponse := "Hello from stream!"
		if response != expectedResponse {
			t.Errorf("Expected '%s', got: %s", expectedResponse, response)
		}
	})

	t.Run("Builder with adapter - parameter passing", func(t *testing.T) {
		adapter := &mockTestAdapter{
			responses: []string{"Parameter test passed"},
		}

		builder := NewWithAdapter("test-model", adapter).
			WithTemperature(0.8).
			WithMaxTokens(1000).
			WithTopP(0.9).
			WithPresencePenalty(0.5).
			WithFrequencyPenalty(0.3).
			WithSeed(42)

		_, err := builder.Ask(ctx, "Test parameters")
		if err != nil {
			t.Fatalf("Ask should not fail: %v", err)
		}

		// Verify parameters were passed to adapter
		if adapter.lastRequest == nil {
			t.Fatal("Adapter should have received request")
		}

		req := adapter.lastRequest
		if req.Temperature != 0.8 {
			t.Errorf("Expected temperature 0.8, got: %f", req.Temperature)
		}
		if req.MaxTokens != 1000 {
			t.Errorf("Expected maxTokens 1000, got: %d", req.MaxTokens)
		}
		if req.TopP != 0.9 {
			t.Errorf("Expected topP 0.9, got: %f", req.TopP)
		}
		if req.PresencePenalty != 0.5 {
			t.Errorf("Expected presencePenalty 0.5, got: %f", req.PresencePenalty)
		}
		if req.FrequencyPenalty != 0.3 {
			t.Errorf("Expected frequencyPenalty 0.3, got: %f", req.FrequencyPenalty)
		}
		if req.Seed != 42 {
			t.Errorf("Expected seed 42, got: %d", req.Seed)
		}
	})

	t.Run("Builder with adapter - system prompt and messages", func(t *testing.T) {
		adapter := &mockTestAdapter{
			responses: []string{"System and messages test passed"},
		}

		builder := NewWithAdapter("test-model", adapter).
			WithSystem("You are a helpful assistant").
			WithMessages([]Message{
				{Role: "user", Content: "Previous question"},
				{Role: "assistant", Content: "Previous answer"},
			})

		_, err := builder.Ask(ctx, "New question")
		if err != nil {
			t.Fatalf("Ask should not fail: %v", err)
		}

		// Verify request contains correct messages
		req := adapter.lastRequest
		if req.System != "You are a helpful assistant" {
			t.Errorf("Expected system prompt 'You are a helpful assistant', got: %s", req.System)
		}

		// Should have: previous user, previous assistant, new user (system is in separate field)
		expectedMessageCount := 3
		if len(req.Messages) != expectedMessageCount {
			t.Errorf("Expected %d messages, got %d: %v", expectedMessageCount, len(req.Messages), req.Messages)
		}

		// Verify last message is our new question
		lastMessage := req.Messages[len(req.Messages)-1]
		if lastMessage.Content != "New question" {
			t.Errorf("Expected last message content 'New question', got: %s", lastMessage.Content)
		}
	})
}

// TestBuilderAdapterVsClient tests that adapter takes precedence over OpenAI client
func TestBuilderAdapterVsClient(t *testing.T) {
	ctx := context.Background()

	t.Run("Adapter takes precedence over client initialization", func(t *testing.T) {
		adapter := &mockTestAdapter{
			responses: []string{"Adapter should be used"},
		}

		// Create builder with adapter but also set OpenAI configuration
		builder := NewWithAdapter("test-model", adapter).
			WithAPIKey("sk-fake-key") // This should be ignored due to adapter

		response, err := builder.Ask(ctx, "Test")
		if err != nil {
			t.Fatalf("Ask should not fail: %v", err)
		}

		if response != "Adapter should be used" {
			t.Errorf("Expected adapter response, got: %s", response)
		}

		// Verify adapter was used, not OpenAI client
		if !adapter.wasCalled {
			t.Error("Adapter should have been called instead of initializing OpenAI client")
		}
	})

	t.Run("No adapter - uses OpenAI client", func(t *testing.T) {
		// This test verifies that without adapter, the builder tries to use OpenAI
		// (it will fail with API key error, but that proves it's trying to use OpenAI)
		builder := NewOpenAI("gpt-4o-mini", ""). // Missing API key
			WithSystem("Test system")

		_, err := builder.Ask(ctx, "Test")
		if err == nil {
			t.Error("Should fail with missing API key when no adapter")
		}

		// Should contain OpenAI validation error
		if !strings.Contains(err.Error(), "OpenAI API key is required") {
			t.Errorf("Expected API key error, got: %v", err)
		}
	})
}

// TestBuilderAdapterEdgeCases tests edge cases and error conditions
func TestBuilderAdapterEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Adapter returns error", func(t *testing.T) {
		adapter := &mockTestAdapter{
			shouldError: true,
			errorMessage: "Adapter intentionally failed",
		}

		builder := NewWithAdapter("test-model", adapter)

		_, err := builder.Ask(ctx, "Test")
		if err == nil {
			t.Error("Ask should fail when adapter returns error")
		}

		if !strings.Contains(err.Error(), "Adapter intentionally failed") {
			t.Errorf("Expected adapter error message, got: %v", err)
		}
	})

	t.Run("Adapter returns error in stream", func(t *testing.T) {
		adapter := &mockTestAdapter{
			shouldError: true,
			errorMessage: "Stream adapter failed",
		}

		builder := NewWithAdapter("test-model", adapter)

		var chunks []string
		_, err := builder.OnStream(func(chunk string) {
			chunks = append(chunks, chunk)
		}).Stream(ctx, "Test")

		if err == nil {
			t.Error("Stream should fail when adapter returns error")
		}

		if !strings.Contains(err.Error(), "Stream adapter failed") {
			t.Errorf("Expected adapter error message, got: %v", err)
		}
	})

	t.Run("Adapter with timeout", func(t *testing.T) {
		adapter := &mockTestAdapter{
			responses: []string{"Response after delay"},
			delay:      100 * time.Millisecond,
		}

		builder := NewWithAdapter("test-model", adapter).
			WithTimeout(50 * time.Millisecond) // Shorter than adapter delay

		start := time.Now()
		_, err := builder.Ask(ctx, "Test")
		duration := time.Since(start)

		if err == nil {
			t.Error("Ask should fail due to timeout")
		}

		// Should timeout quickly, not wait for adapter delay
		if duration > 200*time.Millisecond {
			t.Errorf("Should have timed out quickly, but took %v", duration)
		}
	})

	t.Run("Adapter with nil callback in stream", func(t *testing.T) {
		adapter := &mockTestAdapter{
			responses: []string{"Test response"},
		}

		builder := NewWithAdapter("test-model", adapter)

		// Should not panic even with nil callback
		_, err := builder.OnStream(nil).Stream(ctx, "Test")
		if err != nil {
			t.Errorf("Stream should work with nil callback: %v", err)
		}
	})
}

// TestBuilderAdapterTools tests adapter integration with tools
func TestBuilderAdapterTools(t *testing.T) {
	ctx := context.Background()

	t.Run("Builder with adapter and tools", func(t *testing.T) {
		adapter := &mockTestAdapter{
			responses: []string{"Tools integration working"},
			toolCalls: []ToolCall{
				{
					ID:        "call_123",
					Type:      "function",
					Name:      "test_function",
					Arguments: `{"param": "value"}`,
				},
			},
		}

		tool := &Tool{
			Name:        "test_function",
			Description: "A test function",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param": map[string]interface{}{
						"type": "string",
						"description": "A parameter",
					},
				},
				"required": []string{"param"},
			},
		}

		builder := NewWithAdapter("test-model", adapter).
			WithTools(tool).
			WithAutoExecute(false) // Don't auto-execute for this test

		response, err := builder.Ask(ctx, "Use the test function")
		if err != nil {
			t.Fatalf("Ask should not fail: %v", err)
		}

		if response != "Tools integration working" {
			t.Errorf("Expected 'Tools integration working', got: %s", response)
		}

		// Verify tools were passed to adapter
		req := adapter.lastRequest
		if len(req.Tools) != 1 {
			t.Errorf("Expected 1 tool, got: %d", len(req.Tools))
		}

		if req.Tools[0].Name != "test_function" {
			t.Errorf("Expected tool name 'test_function', got: %s", req.Tools[0].Name)
		}
	})
}

// TestBuilderAdapterMemory tests adapter integration with memory systems
func TestBuilderAdapterMemory(t *testing.T) {
	ctx := context.Background()

	t.Run("Builder with adapter and short memory", func(t *testing.T) {
		adapter := &mockTestAdapter{
			responses: []string{"Response 1", "Response 2"},
		}

		builder := NewWithAdapter("test-model", adapter).
			WithShortMemory().
			WithMaxHistory(10)

		// First message
		response1, err := builder.Ask(ctx, "First message")
		if err != nil {
			t.Fatalf("First Ask should not fail: %v", err)
		}

		// Second message - should include first conversation
		response2, err := builder.Ask(ctx, "Second message")
		if err != nil {
			t.Fatalf("Second Ask should not fail: %v", err)
		}

		if response1 != "Response 1" {
			t.Errorf("Expected 'Response 1', got: %s", response1)
		}

		if response2 != "Response 2" {
			t.Errorf("Expected 'Response 2', got: %s", response2)
		}

		// Verify memory was maintained
		req := adapter.lastRequest
		// Should have: user: First, assistant: Response 1, user: Second (system is in separate field)
		expectedMessageCount := 3
		if len(req.Messages) != expectedMessageCount {
			t.Errorf("Expected %d messages with memory, got %d: %v", expectedMessageCount, len(req.Messages), req.Messages)
		}
	})
}

// TestBuilderAdapterFromEnv tests FromEnv with adapter scenarios
func TestBuilderAdapterFromEnv(t *testing.T) {
	// Note: This test doesn't modify environment variables to avoid test interference
	// It tests the logic that would be used in FromEnv

	t.Run("FromEnv adapter behavior", func(t *testing.T) {
		// Test that FromEnv() creates builders that can use adapters
		adapter := &mockTestAdapter{
			responses: []string{"FromEnv adapter test"},
		}

		// This simulates what would happen if FromEnv() created an adapter-based builder
		builder := NewWithAdapter("test-model", adapter)

		response, err := builder.Ask(context.Background(), "Test")
		if err != nil {
			t.Fatalf("Ask should not fail: %v", err)
		}

		if response != "FromEnv adapter test" {
			t.Errorf("Expected 'FromEnv adapter test', got: %s", response)
		}
	})
}

// mockTestAdapter is a test implementation of LLMAdapter
type mockTestAdapter struct {
	responses       []string
	streamResponses []string
	toolCalls       []ToolCall
	shouldError     bool
	errorMessage    string
	delay           time.Duration
	wasCalled       bool
	lastRequest     *CompletionRequest
	callCount       int
}

func (m *mockTestAdapter) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	m.wasCalled = true
	m.callCount++
	m.lastRequest = req

	if m.delay > 0 {
		// Respect context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.delay):
			// Continue with execution
		}
	}

	if m.shouldError {
		return nil, errors.New(m.errorMessage)
	}

	content := "Default mock response"
	if len(m.responses) > 0 {
		content = m.responses[0]
	}

	response := &CompletionResponse{
		Content:    content,
		ToolCalls:  m.toolCalls,
		Usage: TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
		FinishReason: "stop",
		ID:          "test-response-id",
		Model:       req.Model,
		Created:     time.Now().Unix(),
	}

	// Rotate responses for multiple calls
	if len(m.responses) > 1 {
		m.responses = m.responses[1:]
	}

	return response, nil
}

func (m *mockTestAdapter) Stream(ctx context.Context, req *CompletionRequest, onChunk func(string)) (*CompletionResponse, error) {
	m.wasCalled = true
	m.callCount++
	m.lastRequest = req

	if m.delay > 0 {
		// Respect context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.delay):
			// Continue with execution
		}
	}

	if m.shouldError {
		return nil, errors.New(m.errorMessage)
	}

	var fullContent strings.Builder
	for _, chunk := range m.streamResponses {
		if onChunk != nil {
			onChunk(chunk)
		}
		fullContent.WriteString(chunk)
	}

	response := &CompletionResponse{
		Content: fullContent.String(),
		ToolCalls: m.toolCalls,
		Usage: TokenUsage{
			PromptTokens:     10,
			CompletionTokens: len(m.streamResponses),
			TotalTokens:      10 + len(m.streamResponses),
		},
		FinishReason: "stop",
		ID:          "test-stream-id",
		Model:       req.Model,
		Created:     time.Now().Unix(),
	}

	return response, nil
}

func (m *mockTestAdapter) Close() error {
	return nil
}

// TestBuilderAdapterRealWorldScenario tests realistic usage patterns
func TestBuilderAdapterRealWorldScenario(t *testing.T) {
	ctx := context.Background()

	t.Run("Complete conversation flow with adapter", func(t *testing.T) {
		adapter := &mockTestAdapter{
			responses: []string{
				"Hello! How can I help you today?",
				"The weather in Paris is currently 22°C and sunny.",
				"You're welcome! Have a great day.",
			},
		}

		tool := &Tool{
			Name:        "get_weather",
			Description: "Get weather information for a location",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "City name",
					},
				},
				"required": []string{"location"},
			},
		}

		builder := NewWithAdapter("test-model", adapter).
			WithSystem("You are a helpful weather assistant").
			WithTools(tool).
			WithTemperature(0.7).
			WithShortMemory()

		// Simulate complete conversation
		questions := []string{
			"Hello!",
			"What's the weather in Paris?",
			"Thank you!",
		}

		for i, question := range questions {
			response, err := builder.Ask(ctx, question)
			if err != nil {
				t.Fatalf("Question %d should not fail: %v", i+1, err)
			}

			expectedResponses := []string{
				"Hello! How can I help you today?",
				"The weather in Paris is currently 22°C and sunny.",
				"You're welcome! Have a great day.",
			}

			if response != expectedResponses[i] {
				t.Errorf("Question %d: Expected '%s', got: %s", i+1, expectedResponses[i], response)
			}
		}

		// Verify conversation context was maintained
		req := adapter.lastRequest
		// Should have: user: Hello!, assistant: Hello!, user: What's weather?, assistant: Weather, user: Thank you! (system is in separate field)
		expectedMessages := 5
		if len(req.Messages) != expectedMessages {
			t.Errorf("Expected %d messages in conversation, got %d", expectedMessages, len(req.Messages))
		}

		// Verify all calls went to adapter
		if adapter.callCount != 3 {
			t.Errorf("Expected 3 adapter calls, got %d", adapter.callCount)
		}
	})
}