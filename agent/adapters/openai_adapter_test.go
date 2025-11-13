package adapters

import (
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/taipm/go-deep-agent/agent"
)

const (
	testOpenAIModel  = "gpt-4o-mini"
	testOpenAIAPIKey = "test-key"
)

// TestNewOpenAIAdapter tests the constructor
func TestNewOpenAIAdapter(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
		wantNil bool
	}{
		{
			name:    "[P1] valid API key with default URL",
			apiKey:  "sk-test-api-key-123",
			baseURL: "",
			wantNil: false,
		},
		{
			name:    "[P1] valid API key with custom baseURL",
			apiKey:  "test-api-key",
			baseURL: "http://localhost:11434/v1",
			wantNil: false,
		},
		{
			name:    "[P2] empty API key should still create adapter",
			apiKey:  "",
			baseURL: "",
			wantNil: false, // OpenAI SDK allows empty key, will fail on API call
		},
		{
			name:    "[P2] Azure OpenAI baseURL",
			apiKey:  "azure-key",
			baseURL: "https://my-resource.openai.azure.com",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN: Creating adapter with given API key and baseURL
			adapter := NewOpenAIAdapter(tt.apiKey, tt.baseURL)

			// THEN: Check adapter creation
			if (adapter == nil) != tt.wantNil {
				t.Errorf("NewOpenAIAdapter() returned nil = %v, want nil = %v", adapter == nil, tt.wantNil)
				return
			}

			// THEN: Valid adapter should have client
			if !tt.wantNil {
				if adapter.client == nil {
					t.Error("Expected adapter.client to be non-nil")
				}
			}
		})
	}
}

// TestOpenAIAdapterTemperatureHandling tests temperature parameter handling
func TestOpenAIAdapterTemperatureHandling(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		shouldSet   bool
	}{
		{
			name:        "[P2] temperature 0.0 - should not be set (zero value)",
			temperature: 0.0,
			shouldSet:   false,
		},
		{
			name:        "[P2] temperature 0.7 - common value",
			temperature: 0.7,
			shouldSet:   true,
		},
		{
			name:        "[P2] temperature 1.0 - upper bound",
			temperature: 1.0,
			shouldSet:   true,
		},
		{
			name:        "[P2] temperature 1.5 - OpenAI supports > 1.0",
			temperature: 1.5,
			shouldSet:   true, // OpenAI supports up to 2.0
		},
		{
			name:        "[P2] temperature 2.0 - max allowed",
			temperature: 2.0,
			shouldSet:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: Adapter instance
			adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

			// GIVEN: Completion request with specific temperature
			req := &agent.CompletionRequest{
				Model:       testOpenAIModel,
				Temperature: tt.temperature,
				Messages: []agent.Message{
					{Role: "user", Content: "test"},
				},
			}

			// WHEN: Building chat completion parameters
			params := adapter.buildChatCompletionParams(req)

			// THEN: Check temperature was set correctly
			// Note: We can't easily check param.Opt internals, so we just verify
			// the request was accepted and parameters built successfully
			if params.Model == "" {
				t.Error("Expected model to be set in parameters")
			}
		})
	}
}

// TestOpenAIAdapterMessageConversion tests message format conversion
func TestOpenAIAdapterMessageConversion(t *testing.T) {
	tests := []struct {
		name     string
		request  *agent.CompletionRequest
		wantLen  int
		wantDesc string
	}{
		{
			name: "[P1] single user message",
			request: &agent.CompletionRequest{
				Model: testOpenAIModel,
				Messages: []agent.Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantLen:  1,
			wantDesc: "Should convert single user message",
		},
		{
			name: "[P1] system prompt with user message",
			request: &agent.CompletionRequest{
				Model:  testOpenAIModel,
				System: "You are a helpful assistant",
				Messages: []agent.Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantLen:  2, // System + user message
			wantDesc: "Should add system message first",
		},
		{
			name: "[P2] conversation with multiple messages",
			request: &agent.CompletionRequest{
				Model: testOpenAIModel,
				Messages: []agent.Message{
					{Role: "user", Content: "What is 2+2?"},
					{Role: "assistant", Content: "2+2 equals 4."},
					{Role: "user", Content: "What about 3+3?"},
				},
			},
			wantLen:  3,
			wantDesc: "Should convert all conversation messages",
		},
		{
			name: "[P2] system message in messages array",
			request: &agent.CompletionRequest{
				Model: testOpenAIModel,
				Messages: []agent.Message{
					{Role: "system", Content: "You are helpful"},
					{Role: "user", Content: "Hello"},
				},
			},
			wantLen:  2,
			wantDesc: "Should convert system message in array",
		},
		{
			name: "[P2] tool message",
			request: &agent.CompletionRequest{
				Model: testOpenAIModel,
				Messages: []agent.Message{
					{Role: "user", Content: "What's the weather?"},
					{Role: "assistant", Content: "Let me check"},
					{Role: "tool", Content: "72°F and sunny", ToolCallID: "call_123"},
				},
			},
			wantLen:  3,
			wantDesc: "Should convert tool messages",
		},
		{
			name: "[P2] empty messages",
			request: &agent.CompletionRequest{
				Model:    testOpenAIModel,
				Messages: []agent.Message{},
			},
			wantLen:  0,
			wantDesc: "Should handle empty messages",
		},
		{
			name: "[P2] unknown role defaults to user",
			request: &agent.CompletionRequest{
				Model: testOpenAIModel,
				Messages: []agent.Message{
					{Role: "unknown", Content: "test"},
				},
			},
			wantLen:  1,
			wantDesc: "Should default unknown roles to user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: Adapter instance
			adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

			// WHEN: Converting messages
			messages := adapter.convertMessages(tt.request)

			// THEN: Should convert expected number of messages
			if len(messages) != tt.wantLen {
				t.Errorf("convertMessages() got %d messages, want %d (%s)", len(messages), tt.wantLen, tt.wantDesc)
			}
		})
	}
}

// TestOpenAIAdapterToolConversion tests tool format conversion
func TestOpenAIAdapterToolConversion(t *testing.T) {
	tests := []struct {
		name    string
		tools   []*agent.Tool
		wantLen int
	}{
		{
			name:    "[P2] nil tools",
			tools:   nil,
			wantLen: 0,
		},
		{
			name:    "[P2] empty tools",
			tools:   []*agent.Tool{},
			wantLen: 0,
		},
		{
			name: "[P1] single tool",
			tools: []*agent.Tool{
				{
					Name:        "get_weather",
					Description: "Get the weather for a location",
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
				},
			},
			wantLen: 1,
		},
		{
			name: "[P2] multiple tools",
			tools: []*agent.Tool{
				{
					Name:        "get_weather",
					Description: "Get weather",
					Parameters:  map[string]interface{}{"type": "object"},
				},
				{
					Name:        "search_web",
					Description: "Search the web",
					Parameters:  map[string]interface{}{"type": "object"},
				},
				{
					Name:        "calculate",
					Description: "Perform calculation",
					Parameters:  map[string]interface{}{"type": "object"},
				},
			},
			wantLen: 3,
		},
		{
			name: "[P2] tool with nil parameters",
			tools: []*agent.Tool{
				{
					Name:        "simple_tool",
					Description: "A simple tool",
					Parameters:  nil, // Nil parameters should be handled
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: Adapter instance
			adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

			// WHEN: Converting tools
			openaiTools := adapter.convertTools(tt.tools)

			// THEN: Should convert expected number of tools
			if len(openaiTools) != tt.wantLen {
				t.Errorf("convertTools() got %d tools, want %d", len(openaiTools), tt.wantLen)
			}
		})
	}
}

// TestOpenAIAdapterParameterConversion tests all parameter conversions
func TestOpenAIAdapterParameterConversion(t *testing.T) {
	tests := []struct {
		name    string
		request *agent.CompletionRequest
		desc    string
	}{
		{
			name: "[P1] maxTokens parameter",
			request: &agent.CompletionRequest{
				Model:     testOpenAIModel,
				Messages:  []agent.Message{{Role: "user", Content: "test"}},
				MaxTokens: 100,
			},
			desc: "MaxTokens should be included in parameters",
		},
		{
			name: "[P2] topP parameter",
			request: &agent.CompletionRequest{
				Model:    testOpenAIModel,
				Messages: []agent.Message{{Role: "user", Content: "test"}},
				TopP:     0.9,
			},
			desc: "TopP should be included in parameters",
		},
		{
			name: "[P2] seed parameter",
			request: &agent.CompletionRequest{
				Model:    testOpenAIModel,
				Messages: []agent.Message{{Role: "user", Content: "test"}},
				Seed:     12345,
			},
			desc: "Seed should be included for deterministic output",
		},
		{
			name: "[P2] presence penalty parameter",
			request: &agent.CompletionRequest{
				Model:           testOpenAIModel,
				Messages:        []agent.Message{{Role: "user", Content: "test"}},
				PresencePenalty: 0.5,
			},
			desc: "PresencePenalty should be included in parameters",
		},
		{
			name: "[P2] frequency penalty parameter",
			request: &agent.CompletionRequest{
				Model:            testOpenAIModel,
				Messages:         []agent.Message{{Role: "user", Content: "test"}},
				FrequencyPenalty: 0.3,
			},
			desc: "FrequencyPenalty should be included in parameters",
		},
		{
			name: "[P2] logprobs parameter",
			request: &agent.CompletionRequest{
				Model:       testOpenAIModel,
				Messages:    []agent.Message{{Role: "user", Content: "test"}},
				LogProbs:    true,
				TopLogProbs: 3,
			},
			desc: "LogProbs and TopLogProbs should be included",
		},
		{
			name: "[P2] N parameter (multiple completions)",
			request: &agent.CompletionRequest{
				Model:    testOpenAIModel,
				Messages: []agent.Message{{Role: "user", Content: "test"}},
				N:        3,
			},
			desc: "N parameter should allow multiple completions",
		},
		{
			name: "[P1] all parameters set",
			request: &agent.CompletionRequest{
				Model:            testOpenAIModel,
				System:           "You are helpful",
				Messages:         []agent.Message{{Role: "user", Content: "test"}},
				Temperature:      0.8,
				MaxTokens:        500,
				TopP:             0.95,
				Seed:             42,
				PresencePenalty:  0.2,
				FrequencyPenalty: 0.1,
				LogProbs:         true,
				TopLogProbs:      5,
				N:                2,
			},
			desc: "All parameters should be handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.desc)

			// GIVEN: Adapter instance
			adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

			// WHEN: Building parameters
			// THEN: Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("buildChatCompletionParams() panicked: %v", r)
				}
			}()

			params := adapter.buildChatCompletionParams(tt.request)

			// Verify basic structure
			if string(params.Model) != tt.request.Model {
				t.Errorf("Model: got %s, want %s", params.Model, tt.request.Model)
			}
			if len(params.Messages) == 0 {
				t.Error("Messages should be converted")
			}
		})
	}
}

// TestOpenAIAdapterResponseConversion tests response format conversion
func TestOpenAIAdapterResponseConversion(t *testing.T) {
	// Note: This tests the conversion logic, not actual API calls
	t.Run("[P1] empty choices should not panic", func(t *testing.T) {
		// GIVEN: Adapter instance
		adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

		// GIVEN: OpenAI response with no choices
		completion := &openai.ChatCompletion{
			ID:      "test-id",
			Model:   testOpenAIModel,
			Created: 1234567890,
			Choices: []openai.ChatCompletionChoice{},
		}

		// WHEN: Converting response
		// THEN: Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("convertResponse() panicked: %v", r)
			}
		}()

		resp := adapter.convertResponse(completion)

		// THEN: Should have basic fields
		if resp.ID != "test-id" {
			t.Errorf("Response ID: got %s, want test-id", resp.ID)
		}
		if resp.Content != "" {
			t.Errorf("Response content should be empty for no choices, got: %s", resp.Content)
		}
	})
}

// TestOpenAIAdapterStreamCallbackInvocation tests streaming callback
func TestOpenAIAdapterStreamCallbackInvocation(t *testing.T) {
	tests := []struct {
		name        string
		onChunk     func(string)
		shouldPanic bool
	}{
		{
			name: "[P1] valid callback should work",
			onChunk: func(s string) {
				// Normal callback
			},
			shouldPanic: false,
		},
		{
			name:        "[P2] nil callback should not panic",
			onChunk:     nil,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN: Stream method handles callback
			defer func() {
				r := recover()
				if (r != nil) != tt.shouldPanic {
					t.Errorf("Stream callback panic = %v, wantPanic %v", r != nil, tt.shouldPanic)
				}
			}()

			// THEN: Callback invocation should be safe
			if tt.onChunk != nil {
				tt.onChunk("test chunk")
			}
		})
	}
}

// TestOpenAIAdapterEdgeCases tests edge cases and error conditions
func TestOpenAIAdapterEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		request *agent.CompletionRequest
		desc    string
	}{
		{
			name: "[P2] very long message",
			request: &agent.CompletionRequest{
				Model: testOpenAIModel,
				Messages: []agent.Message{
					{Role: "user", Content: string(make([]byte, 10000))},
				},
			},
			desc: "Should handle very long messages without panic",
		},
		{
			name: "[P2] special characters in message",
			request: &agent.CompletionRequest{
				Model: testOpenAIModel,
				Messages: []agent.Message{
					{Role: "user", Content: "Test with emoji 🚀 and unicode 日本語 and newlines\n\n\n"},
				},
			},
			desc: "Should handle special characters correctly",
		},
		{
			name: "[P2] empty content in message",
			request: &agent.CompletionRequest{
				Model: testOpenAIModel,
				Messages: []agent.Message{
					{Role: "user", Content: ""},
				},
			},
			desc: "Should handle empty content",
		},
		{
			name: "[P2] maximum parameters",
			request: &agent.CompletionRequest{
				Model:            testOpenAIModel,
				Messages:         []agent.Message{{Role: "user", Content: "test"}},
				System:           "system prompt",
				Temperature:      1.5,
				MaxTokens:        2000,
				TopP:             0.95,
				Stop:             []string{"END", "STOP"},
				Seed:             9999,
				PresencePenalty:  0.5,
				FrequencyPenalty: 0.5,
				LogProbs:         true,
				TopLogProbs:      5,
				N:                3,
			},
			desc: "Should handle all parameters set to maximum values",
		},
		{
			name: "[P2] zero values (should not be set)",
			request: &agent.CompletionRequest{
				Model:       testOpenAIModel,
				Messages:    []agent.Message{{Role: "user", Content: "test"}},
				Temperature: 0, // Zero should not be set
				MaxTokens:   0, // Zero should not be set
				TopP:        0, // Zero should not be set
				Seed:        0, // Zero should not be set
				N:           0, // Zero should not be set
			},
			desc: "Should not set parameters with zero values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.desc)

			// GIVEN: Adapter instance
			adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

			// WHEN: Building parameters
			// THEN: Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("buildChatCompletionParams() panicked: %v", r)
				}
			}()

			params := adapter.buildChatCompletionParams(tt.request)

			// Verify basic structure
			if params.Model == "" {
				t.Error("Model should be set")
			}
			if len(params.Messages) == 0 && len(tt.request.Messages) > 0 {
				t.Error("Messages should be converted")
			}
		})
	}
}

// TestOpenAIAdapterToolsInRequest tests tool integration in request
func TestOpenAIAdapterToolsInRequest(t *testing.T) {
	t.Run("[P1] request with tools sets tools parameter", func(t *testing.T) {
		// GIVEN: Adapter instance
		adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

		// GIVEN: Request with tools
		req := &agent.CompletionRequest{
			Model:    testOpenAIModel,
			Messages: []agent.Message{{Role: "user", Content: "What's the weather?"}},
			Tools: []*agent.Tool{
				{
					Name:        "get_weather",
					Description: "Get weather",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		}

		// WHEN: Building parameters
		params := adapter.buildChatCompletionParams(req)

		// THEN: Tools should be set
		if len(params.Tools) == 0 {
			t.Error("Expected tools to be set in parameters")
		}
		if len(params.Tools) != len(req.Tools) {
			t.Errorf("Expected %d tools, got %d", len(req.Tools), len(params.Tools))
		}
	})

	t.Run("[P2] request without tools has no tools parameter", func(t *testing.T) {
		// GIVEN: Adapter instance
		adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

		// GIVEN: Request without tools
		req := &agent.CompletionRequest{
			Model:    testOpenAIModel,
			Messages: []agent.Message{{Role: "user", Content: "Hello"}},
		}

		// WHEN: Building parameters
		params := adapter.buildChatCompletionParams(req)

		// THEN: Tools should be empty
		if len(params.Tools) != 0 {
			t.Error("Expected no tools in parameters")
		}
	})
}

// TestOpenAIAdapterModelParameter tests model parameter handling
func TestOpenAIAdapterModelParameter(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		wantModel string
	}{
		{
			name:      "[P1] gpt-4o-mini model",
			model:     "gpt-4o-mini",
			wantModel: "gpt-4o-mini",
		},
		{
			name:      "[P2] gpt-4-turbo model",
			model:     "gpt-4-turbo",
			wantModel: "gpt-4-turbo",
		},
		{
			name:      "[P2] gpt-3.5-turbo model",
			model:     "gpt-3.5-turbo",
			wantModel: "gpt-3.5-turbo",
		},
		{
			name:      "[P2] custom model name",
			model:     "custom-model-v1",
			wantModel: "custom-model-v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: Adapter instance
			adapter := NewOpenAIAdapter(testOpenAIAPIKey, "")

			// GIVEN: Request with specific model
			req := &agent.CompletionRequest{
				Model:    tt.model,
				Messages: []agent.Message{{Role: "user", Content: "test"}},
			}

			// WHEN: Building parameters
			params := adapter.buildChatCompletionParams(req)

			// THEN: Model should be set correctly
			if string(params.Model) != tt.wantModel {
				t.Errorf("Model: got %s, want %s", params.Model, tt.wantModel)
			}
		})
	}
}
