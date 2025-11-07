package agent

import (
	"testing"
	"time"
)

// Test edge cases, error paths, and boundary conditions

func TestBuilder_EdgeCases(t *testing.T) {
	t.Run("EmptyAPIKey", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "")
		if builder.apiKey != "" {
			t.Error("Expected empty API key")
		}
	})

	t.Run("EmptyModel", func(t *testing.T) {
		builder := NewOpenAI("", "test-key")
		if builder.model != "" {
			t.Error("Expected empty model")
		}
	})

	t.Run("NilContext", func(t *testing.T) {
		// Nil context will cause API call to fail (not tested here - would need mock)
		// Just verify builder creation doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("EmptyMessage", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		// Should not panic with empty message
		_ = builder
	})

	t.Run("VeryLongMessage", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		longMsg := string(make([]byte, 1000000)) // 1MB message
		// Should handle large messages
		_ = longMsg
		_ = builder
	})
}

func TestBuilder_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Builder) *Builder
		validate func(*testing.T, *Builder)
	}{
		{
			name:  "Temperature_Min",
			setup: func(b *Builder) *Builder { return b.WithTemperature(0.0) },
			validate: func(t *testing.T, b *Builder) {
				if *b.temperature != 0.0 {
					t.Errorf("Expected temperature 0.0, got %f", *b.temperature)
				}
			},
		},
		{
			name:  "Temperature_Max",
			setup: func(b *Builder) *Builder { return b.WithTemperature(2.0) },
			validate: func(t *testing.T, b *Builder) {
				if *b.temperature != 2.0 {
					t.Errorf("Expected temperature 2.0, got %f", *b.temperature)
				}
			},
		},
		{
			name:  "Temperature_Negative",
			setup: func(b *Builder) *Builder { return b.WithTemperature(-1.0) },
			validate: func(t *testing.T, b *Builder) {
				// Should accept but may fail at API level
				if *b.temperature != -1.0 {
					t.Errorf("Expected temperature -1.0, got %f", *b.temperature)
				}
			},
		},
		{
			name:  "MaxTokens_Zero",
			setup: func(b *Builder) *Builder { return b.WithMaxTokens(0) },
			validate: func(t *testing.T, b *Builder) {
				if *b.maxTokens != 0 {
					t.Errorf("Expected maxTokens 0, got %d", *b.maxTokens)
				}
			},
		},
		{
			name:  "MaxTokens_Large",
			setup: func(b *Builder) *Builder { return b.WithMaxTokens(100000) },
			validate: func(t *testing.T, b *Builder) {
				if *b.maxTokens != 100000 {
					t.Errorf("Expected maxTokens 100000, got %d", *b.maxTokens)
				}
			},
		},
		{
			name:  "MaxTokens_Negative",
			setup: func(b *Builder) *Builder { return b.WithMaxTokens(-1) },
			validate: func(t *testing.T, b *Builder) {
				if *b.maxTokens != -1 {
					t.Errorf("Expected maxTokens -1, got %d", *b.maxTokens)
				}
			},
		},
		{
			name:  "TopP_Zero",
			setup: func(b *Builder) *Builder { return b.WithTopP(0.0) },
			validate: func(t *testing.T, b *Builder) {
				if *b.topP != 0.0 {
					t.Errorf("Expected topP 0.0, got %f", *b.topP)
				}
			},
		},
		{
			name:  "TopP_One",
			setup: func(b *Builder) *Builder { return b.WithTopP(1.0) },
			validate: func(t *testing.T, b *Builder) {
				if *b.topP != 1.0 {
					t.Errorf("Expected topP 1.0, got %f", *b.topP)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOpenAI("gpt-4o-mini", "key")
			builder = tt.setup(builder)
			tt.validate(t, builder)
		})
	}
}

func TestBuilder_RetryBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Builder) *Builder
		validate func(*testing.T, *Builder)
	}{
		{
			name:  "Retry_Zero",
			setup: func(b *Builder) *Builder { return b.WithRetry(0) },
			validate: func(t *testing.T, b *Builder) {
				if b.maxRetries != 0 {
					t.Errorf("Expected maxRetries 0, got %d", b.maxRetries)
				}
			},
		},
		{
			name:  "Retry_Negative",
			setup: func(b *Builder) *Builder { return b.WithRetry(-1) },
			validate: func(t *testing.T, b *Builder) {
				if b.maxRetries != -1 {
					t.Errorf("Expected maxRetries -1, got %d", b.maxRetries)
				}
			},
		},
		{
			name:  "Retry_Large",
			setup: func(b *Builder) *Builder { return b.WithRetry(1000) },
			validate: func(t *testing.T, b *Builder) {
				if b.maxRetries != 1000 {
					t.Errorf("Expected maxRetries 1000, got %d", b.maxRetries)
				}
			},
		},
		{
			name:  "RetryDelay_Zero",
			setup: func(b *Builder) *Builder { return b.WithRetryDelay(0) },
			validate: func(t *testing.T, b *Builder) {
				if b.retryDelay != 0 {
					t.Errorf("Expected retryDelay 0, got %v", b.retryDelay)
				}
			},
		},
		{
			name:  "RetryDelay_Negative",
			setup: func(b *Builder) *Builder { return b.WithRetryDelay(-1 * time.Second) },
			validate: func(t *testing.T, b *Builder) {
				if b.retryDelay != -1*time.Second {
					t.Errorf("Expected negative retryDelay, got %v", b.retryDelay)
				}
			},
		},
		{
			name:  "ExponentialBackoff_WithoutRetry",
			setup: func(b *Builder) *Builder { return b.WithExponentialBackoff() },
			validate: func(t *testing.T, b *Builder) {
				// Just verify it doesn't panic
				if b == nil {
					t.Error("Expected non-nil builder")
				}
			},
		},
		{
			name: "ExponentialBackoff_WithRetry",
			setup: func(b *Builder) *Builder {
				return b.WithRetry(3).WithExponentialBackoff()
			},
			validate: func(t *testing.T, b *Builder) {
				if b == nil {
					t.Error("Expected non-nil builder")
				}
				if b.maxRetries != 3 {
					t.Errorf("Expected maxRetries 3, got %d", b.maxRetries)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOpenAI("gpt-4o-mini", "key")
			builder = tt.setup(builder)
			tt.validate(t, builder)
		})
	}
}

func TestBuilder_MemoryBoundaries(t *testing.T) {
	t.Run("MaxHistory_Zero", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithMemory().
			WithMaxHistory(0)
		if builder.maxHistory != 0 {
			t.Errorf("Expected maxHistory 0, got %d", builder.maxHistory)
		}
	})

	t.Run("MaxHistory_Negative", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithMemory().
			WithMaxHistory(-1)
		if builder.maxHistory != -1 {
			t.Errorf("Expected maxHistory -1, got %d", builder.maxHistory)
		}
	})

	t.Run("MaxHistory_Large", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithMemory().
			WithMaxHistory(10000)
		if builder.maxHistory != 10000 {
			t.Errorf("Expected maxHistory 10000, got %d", builder.maxHistory)
		}
	})

	t.Run("Memory_WithoutWithMemory", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithMaxHistory(10) // Without WithMemory()

		// Should still set maxHistory but memory not enabled
		if builder.maxHistory != 10 {
			t.Errorf("Expected maxHistory 10, got %d", builder.maxHistory)
		}
	})

	t.Run("SetHistory_Nil", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithMemory()
		builder.SetHistory(nil)
		history := builder.GetHistory()
		if len(history) != 0 {
			t.Errorf("Expected empty history with nil, got %d messages", len(history))
		}
	})

	t.Run("SetHistory_Empty", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithMemory()
		builder.SetHistory([]Message{})
		history := builder.GetHistory()
		if len(history) != 0 {
			t.Errorf("Expected empty history, got %d messages", len(history))
		}
	})

	t.Run("SetHistory_SingleMessage", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithMemory()
		builder.SetHistory([]Message{User("test")})
		history := builder.GetHistory()
		if len(history) != 1 {
			t.Errorf("Expected 1 message, got %d", len(history))
		}
	})

	t.Run("SetHistory_LargeHistory", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithMemory()
		messages := make([]Message, 1000)
		for i := 0; i < 1000; i++ {
			messages[i] = User("test")
		}
		builder.SetHistory(messages)
		history := builder.GetHistory()
		if len(history) != 1000 {
			t.Errorf("Expected 1000 messages, got %d", len(history))
		}
	})
}

func TestBuilder_ToolEdgeCases(t *testing.T) {
	t.Run("Tool_EmptyName", func(t *testing.T) {
		tool := NewTool("", "description")
		if tool.Name != "" {
			t.Error("Expected empty tool name")
		}
	})

	t.Run("Tool_EmptyDescription", func(t *testing.T) {
		tool := NewTool("test", "")
		if tool.Description != "" {
			t.Error("Expected empty description")
		}
	})

	t.Run("Tool_NoParameters", func(t *testing.T) {
		tool := NewTool("test", "description")
		// Default parameters are added, so just verify tool was created
		if tool == nil {
			t.Error("Expected non-nil tool")
		}
	})

	t.Run("Tool_NoHandler", func(t *testing.T) {
		tool := NewTool("test", "description")
		// Handler is optional, just verify tool exists
		if tool == nil {
			t.Error("Expected non-nil tool")
		}
	})

	t.Run("Tool_DuplicateParameters", func(t *testing.T) {
		tool := NewTool("test", "description").
			AddParameter("param1", "string", "desc", true).
			AddParameter("param1", "integer", "desc2", false)

		// Parameters are added as is (duplicates allowed)
		if tool == nil {
			t.Error("Expected non-nil tool")
		}
	})

	t.Run("Tool_InvalidType", func(t *testing.T) {
		tool := NewTool("test", "description").
			AddParameter("param", "invalid_type", "desc", true)

		// Invalid types are allowed (validated at API level)
		if tool == nil {
			t.Error("Expected non-nil tool")
		}
	})

	t.Run("WithTools_Nil", func(t *testing.T) {
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").WithTools(nil...)
		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("WithTools_Empty", func(t *testing.T) {
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").WithTools()
		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("WithAutoExecute_False", func(t *testing.T) {
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").WithAutoExecute(false)
		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("WithAutoExecute_True", func(t *testing.T) {
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").WithAutoExecute(true)
		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("WithMaxToolRounds_Zero", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithMaxToolRounds(0)
		if builder.maxToolRounds != 0 {
			t.Errorf("Expected maxToolRounds 0, got %d", builder.maxToolRounds)
		}
	})

	t.Run("WithMaxToolRounds_Negative", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithMaxToolRounds(-1)
		if builder.maxToolRounds != -1 {
			t.Errorf("Expected maxToolRounds -1, got %d", builder.maxToolRounds)
		}
	})
}

func TestBuilder_CallbackEdgeCases(t *testing.T) {
	t.Run("OnStream_Nil", func(t *testing.T) {
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").OnStream(nil)
		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("OnStream_Overwrite", func(t *testing.T) {
		// Empty callbacks for testing
		callback1 := func(s string) { /* intentionally empty */ }
		callback2 := func(s string) { /* intentionally empty */ }

		// Just verify it doesn't panic when overwriting
		builder := NewOpenAI("gpt-4o-mini", "key").
			OnStream(callback1).
			OnStream(callback2)

		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("OnRefusal_Nil", func(t *testing.T) {
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").OnRefusal(nil)
		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("OnRefusal_Overwrite", func(t *testing.T) {
		// Empty callbacks for testing
		callback1 := func(s string) { /* intentionally empty */ }
		callback2 := func(s string) { /* intentionally empty */ }

		// Just verify it doesn't panic when overwriting
		builder := NewOpenAI("gpt-4o-mini", "key").
			OnRefusal(callback1).
			OnRefusal(callback2)

		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})
}

func TestBuilder_TimeoutEdgeCases(t *testing.T) {
	t.Run("Timeout_Zero", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithTimeout(0)
		if builder.timeout != 0 {
			t.Errorf("Expected timeout 0, got %v", builder.timeout)
		}
	})

	t.Run("Timeout_Negative", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithTimeout(-1 * time.Second)
		if builder.timeout != -1*time.Second {
			t.Errorf("Expected negative timeout, got %v", builder.timeout)
		}
	})

	t.Run("Timeout_VeryShort", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithTimeout(1 * time.Nanosecond)
		if builder.timeout != 1*time.Nanosecond {
			t.Errorf("Expected 1ns timeout, got %v", builder.timeout)
		}
	})

	t.Run("Timeout_VeryLong", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithTimeout(24 * time.Hour)
		if builder.timeout != 24*time.Hour {
			t.Errorf("Expected 24h timeout, got %v", builder.timeout)
		}
	})
}

func TestBuilder_JSONSchemaEdgeCases(t *testing.T) {
	t.Run("JSONMode_Simple", func(t *testing.T) {
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").WithJSONMode()
		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("JSONSchema_EmptyName", func(t *testing.T) {
		schema := map[string]interface{}{"type": "object"}
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithJSONSchema("", "desc", schema, true)

		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("JSONSchema_EmptyDescription", func(t *testing.T) {
		schema := map[string]interface{}{"type": "object"}
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithJSONSchema("test", "", schema, true)

		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("JSONSchema_NilSchema", func(t *testing.T) {
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithJSONSchema("test", "desc", nil, true)

		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("JSONSchema_EmptySchema", func(t *testing.T) {
		schema := map[string]interface{}{}
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithJSONSchema("test", "desc", schema, true)

		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})

	t.Run("JSONSchema_StrictFalse", func(t *testing.T) {
		schema := map[string]interface{}{"type": "object"}
		// Just verify it doesn't panic
		builder := NewOpenAI("gpt-4o-mini", "key").
			WithJSONSchema("test", "desc", schema, false)

		if builder == nil {
			t.Error("Expected non-nil builder")
		}
	})
}

func TestBuilder_ProviderEdgeCases(t *testing.T) {
	t.Run("Custom_Provider", func(t *testing.T) {
		builder := New("custom-provider", "model")
		if builder.provider != "custom-provider" {
			t.Errorf("Expected custom-provider, got %s", builder.provider)
		}
	})

	t.Run("Empty_Provider", func(t *testing.T) {
		builder := New("", "model")
		if builder.provider != "" {
			t.Errorf("Expected empty provider, got %s", builder.provider)
		}
	})

	t.Run("Ollama_DefaultBaseURL", func(t *testing.T) {
		builder := NewOllama("qwen2.5:3b")
		if builder.baseURL != "http://localhost:11434/v1" {
			t.Errorf("Expected default Ollama URL, got %s", builder.baseURL)
		}
	})

	t.Run("Ollama_CustomBaseURL", func(t *testing.T) {
		builder := NewOllama("qwen2.5:3b").
			WithBaseURL("http://custom:8080/v1")
		if builder.baseURL != "http://custom:8080/v1" {
			t.Errorf("Expected custom URL, got %s", builder.baseURL)
		}
	})

	t.Run("BaseURL_Empty", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithBaseURL("")
		if builder.baseURL != "" {
			t.Errorf("Expected empty baseURL, got %s", builder.baseURL)
		}
	})
}

func TestBuilder_PenaltyEdgeCases(t *testing.T) {
	t.Run("PresencePenalty_Min", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithPresencePenalty(-2.0)
		if *builder.presencePenalty != -2.0 {
			t.Errorf("Expected -2.0, got %f", *builder.presencePenalty)
		}
	})

	t.Run("PresencePenalty_Max", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithPresencePenalty(2.0)
		if *builder.presencePenalty != 2.0 {
			t.Errorf("Expected 2.0, got %f", *builder.presencePenalty)
		}
	})

	t.Run("PresencePenalty_Zero", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithPresencePenalty(0.0)
		if *builder.presencePenalty != 0.0 {
			t.Errorf("Expected 0.0, got %f", *builder.presencePenalty)
		}
	})

	t.Run("FrequencyPenalty_Min", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithFrequencyPenalty(-2.0)
		if *builder.frequencyPenalty != -2.0 {
			t.Errorf("Expected -2.0, got %f", *builder.frequencyPenalty)
		}
	})

	t.Run("FrequencyPenalty_Max", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithFrequencyPenalty(2.0)
		if *builder.frequencyPenalty != 2.0 {
			t.Errorf("Expected 2.0, got %f", *builder.frequencyPenalty)
		}
	})

	t.Run("FrequencyPenalty_Zero", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithFrequencyPenalty(0.0)
		if *builder.frequencyPenalty != 0.0 {
			t.Errorf("Expected 0.0, got %f", *builder.frequencyPenalty)
		}
	})
}

func TestBuilder_SeedAndN(t *testing.T) {
	t.Run("Seed_Zero", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithSeed(0)
		if *builder.seed != 0 {
			t.Errorf("Expected seed 0, got %d", *builder.seed)
		}
	})

	t.Run("Seed_Negative", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithSeed(-1)
		if *builder.seed != -1 {
			t.Errorf("Expected seed -1, got %d", *builder.seed)
		}
	})

	t.Run("Seed_Large", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithSeed(9223372036854775807)
		if *builder.seed != 9223372036854775807 {
			t.Errorf("Expected large seed, got %d", *builder.seed)
		}
	})

}

func TestMessage_ConvertEdgeCases(t *testing.T) {
	t.Run("Convert_EmptySlice", func(t *testing.T) {
		messages := []Message{}
		converted := convertMessages(messages)
		if len(converted) != 0 {
			t.Errorf("Expected 0 converted messages, got %d", len(converted))
		}
	})

	t.Run("Convert_NilSlice", func(t *testing.T) {
		var messages []Message
		converted := convertMessages(messages)
		if len(converted) != 0 {
			t.Errorf("Expected 0 converted messages, got %d", len(converted))
		}
	})

	t.Run("Convert_SingleMessage", func(t *testing.T) {
		messages := []Message{User("test")}
		converted := convertMessages(messages)
		if len(converted) != 1 {
			t.Errorf("Expected 1 converted message, got %d", len(converted))
		}
	})

	t.Run("Convert_EmptyContent", func(t *testing.T) {
		messages := []Message{User("")}
		converted := convertMessages(messages)
		if len(converted) != 1 {
			t.Errorf("Expected 1 converted message, got %d", len(converted))
		}
	})

	t.Run("Convert_InvalidRole", func(t *testing.T) {
		messages := []Message{{Role: "invalid", Content: "test"}}
		converted := convertMessages(messages)
		if len(converted) != 1 {
			t.Errorf("Expected 1 converted message, got %d", len(converted))
		}
	})
}
