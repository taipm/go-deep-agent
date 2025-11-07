package agent

import (
	"context"
	"testing"
)

const (
	benchModel  = "gpt-4o-mini"
	benchAPIKey = "test-key"
	benchSystem = "You are helpful"
)

// BenchmarkBuilderCreation measures the overhead of creating a Builder
func BenchmarkBuilderCreation(b *testing.B) {
	b.Run("NewOpenAI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewOpenAI(benchModel, benchAPIKey)
		}
	})

	b.Run("NewOllama", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewOllama("qwen2.5:3b")
		}
	})

	b.Run("WithConfiguration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewOpenAI(benchModel, benchAPIKey).
				WithSystem(benchSystem).
				WithTemperature(0.7).
				WithMaxTokens(500).
				WithMemory()
		}
	})
}

// BenchmarkMemoryOperations measures memory management performance
func BenchmarkMemoryOperations(b *testing.B) {
	b.Run("WithMemory", func(b *testing.B) {
		builder := NewOpenAI("gpt-4o-mini", "test-key")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.WithMemory()
		}
	})

	b.Run("GetHistory_Empty", func(b *testing.B) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").WithMemory()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = builder.GetHistory()
		}
	})

	b.Run("GetHistory_10Messages", func(b *testing.B) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").WithMemory()
		for j := 0; j < 10; j++ {
			builder.messages = append(builder.messages, Message{Role: "user", Content: "test"})
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = builder.GetHistory()
		}
	})

	b.Run("SetHistory", func(b *testing.B) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").WithMemory()
		history := []Message{
			System("You are helpful"),
			User("Hello"),
			Assistant("Hi there!"),
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.SetHistory(history)
		}
	})

	b.Run("Clear", func(b *testing.B) {
		builder := NewOpenAI("gpt-4o-mini", "test-key").
			WithSystem("System").
			WithMemory()
		for j := 0; j < 10; j++ {
			builder.messages = append(builder.messages, Message{Role: "user", Content: "test"})
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.Clear()
			// Restore messages for next iteration
			for j := 0; j < 10; j++ {
				builder.messages = append(builder.messages, Message{Role: "user", Content: "test"})
			}
		}
	})
}

// BenchmarkHistoryManagement measures history management performance
func BenchmarkHistoryManagement(b *testing.B) {
	b.Run("SetHistory_10Messages", func(b *testing.B) {
		builder := NewOpenAI(benchModel, benchAPIKey).WithMemory()
		history := make([]Message, 10)
		for j := 0; j < 10; j++ {
			history[j] = Message{Role: "user", Content: "test message"}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.SetHistory(history)
		}
	})

	b.Run("Clear_WithMessages", func(b *testing.B) {
		builder := NewOpenAI(benchModel, benchAPIKey).WithMemory()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.SetHistory([]Message{
				User("test1"),
				Assistant("response1"),
				User("test2"),
			})
			builder.Clear()
		}
	})

	b.Run("WithMaxHistory", func(b *testing.B) {
		builder := NewOpenAI(benchModel, benchAPIKey).WithMemory()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.WithMaxHistory(20)
		}
	})
}

// BenchmarkToolCreation measures tool building performance
func BenchmarkToolCreation(b *testing.B) {
	b.Run("NewTool_Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewTool("test", "description")
		}
	})

	b.Run("NewTool_WithParameters", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewTool("test", "description").
				AddParameter("param1", "string", "desc1", true).
				AddParameter("param2", "integer", "desc2", false).
				AddParameter("param3", "boolean", "desc3", true)
		}
	})

	b.Run("NewTool_WithHandler", func(b *testing.B) {
		handler := func(args string) (string, error) {
			return "result", nil
		}
		for i := 0; i < b.N; i++ {
			_ = NewTool("test", "description").WithHandler(handler)
		}
	})

	b.Run("Tool_Complete", func(b *testing.B) {
		handler := func(args string) (string, error) {
			return "result", nil
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewTool("test", "description").
				AddParameter("location", "string", "City name", true).
				WithHandler(handler)
		}
	})
}

// BenchmarkConfigurationMethods measures method chaining overhead
func BenchmarkConfigurationMethods(b *testing.B) {
	b.Run("MethodChaining_Short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewOpenAI("gpt-4o-mini", "key").
				WithTemperature(0.7).
				WithMaxTokens(500)
		}
	})

	b.Run("MethodChaining_Long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewOpenAI("gpt-4o-mini", "key").
				WithSystem("System").
				WithTemperature(0.7).
				WithTopP(0.9).
				WithMaxTokens(500).
				WithPresencePenalty(0.5).
				WithFrequencyPenalty(0.5).
				WithSeed(42).
				WithMemory().
				WithMaxHistory(20)
		}
	})
}

// BenchmarkErrorChecking measures error type checking performance
func BenchmarkErrorChecking(b *testing.B) {
	apiErr := &APIError{Type: "api_key_error", Message: "Invalid API key"}
	rateErr := &APIError{Type: "rate_limit", Message: "Rate limit"}
	timeoutErr := &APIError{Type: "timeout", Message: "Timeout"}

	b.Run("IsAPIKeyError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = IsAPIKeyError(apiErr)
		}
	})

	b.Run("IsRateLimitError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = IsRateLimitError(rateErr)
		}
	})

	b.Run("IsTimeoutError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = IsTimeoutError(timeoutErr)
		}
	})
}

// BenchmarkMessageHelpers measures message helper performance
func BenchmarkMessageHelpers(b *testing.B) {
	b.Run("System", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = System("You are helpful")
		}
	})

	b.Run("User", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = User("Hello, world!")
		}
	})

	b.Run("Assistant", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Assistant("Hi there!")
		}
	})

	b.Run("All_Helpers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = System("System")
			_ = User("User")
			_ = Assistant("Assistant")
		}
	})
}

// BenchmarkBuilderCopy measures the cost of copying builders for concurrent use
func BenchmarkBuilderCopy(b *testing.B) {
	b.Run("ShallowCopy", func(b *testing.B) {
		original := NewOpenAI("gpt-4o-mini", "key").
			WithSystem("System").
			WithTemperature(0.7).
			WithMemory()

		for j := 0; j < 10; j++ {
			original.messages = append(original.messages, Message{Role: "user", Content: "test"})
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate copying for concurrent use
			_ = &Builder{
				provider:    original.provider,
				model:       original.model,
				apiKey:      original.apiKey,
				temperature: original.temperature,
			}
		}
	})

	b.Run("DeepCopyMessages", func(b *testing.B) {
		original := NewOpenAI("gpt-4o-mini", "key").WithMemory()
		for j := 0; j < 10; j++ {
			original.messages = append(original.messages, Message{Role: "user", Content: "test"})
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			history := original.GetHistory()
			_ = NewOpenAI("gpt-4o-mini", "key").
				WithMemory().
				SetHistory(history)
		}
	})
}

// BenchmarkStreamingSetup measures streaming configuration overhead
func BenchmarkStreamingSetup(b *testing.B) {
	// Empty callback for benchmarking overhead
	callback := func(content string) {
		// Intentionally empty for performance testing
	}

	b.Run("OnStream", func(b *testing.B) {
		builder := NewOpenAI(benchModel, benchAPIKey)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.OnStream(callback)
		}
	})

	b.Run("OnRefusal", func(b *testing.B) {
		builder := NewOpenAI(benchModel, benchAPIKey)
		// Empty callback for benchmarking overhead
		refusalCallback := func(refusal string) {
			// Intentionally empty for performance testing
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.OnRefusal(refusalCallback)
		}
	})
}

// BenchmarkRetryConfiguration measures retry setup performance
func BenchmarkRetryConfiguration(b *testing.B) {
	b.Run("WithRetry", func(b *testing.B) {
		builder := NewOpenAI("gpt-4o-mini", "key")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.WithRetry(3)
		}
	})

	b.Run("WithExponentialBackoff", func(b *testing.B) {
		builder := NewOpenAI("gpt-4o-mini", "key").WithRetry(5)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder.WithExponentialBackoff()
		}
	})

	b.Run("CompleteRetrySetup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewOpenAI("gpt-4o-mini", "key").
				WithRetry(3).
				WithExponentialBackoff()
		}
	})
}

// BenchmarkContextOperations measures context-related operations
func BenchmarkContextOperations(b *testing.B) {
	b.Run("ContextWithCancel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_ = ctx
		}
	})

	b.Run("BuilderWithContext", func(b *testing.B) {
		builder := NewOpenAI("gpt-4o-mini", "key")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_ = ctx
			_ = builder
		}
	})
}
