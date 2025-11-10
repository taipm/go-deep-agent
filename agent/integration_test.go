//go:build integration
// +build integration

package agent

import (
	"context"
	"os"
	"testing"
	"time"
)

// Integration tests that call real APIs (OpenAI, Ollama)
// These tests are skipped by default. Run with: go test -tags=integration

func TestIntegration_OpenAI_SimpleChat(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	response, err := NewOpenAI("gpt-4o-mini", apiKey).
		Ask(ctx, "Say 'test successful' and nothing else")

	if err != nil {
		t.Fatalf("OpenAI integration test failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response from OpenAI")
	}

	t.Logf("OpenAI response: %s", response)
}

func TestIntegration_OpenAI_Streaming(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	var chunks []string

	response, err := NewOpenAI("gpt-4o-mini", apiKey).
		OnStream(func(content string) {
			chunks = append(chunks, content)
		}).
		Stream(ctx, "Count from 1 to 3")

	if err != nil {
		t.Fatalf("OpenAI streaming test failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response from streaming")
	}

	if len(chunks) == 0 {
		t.Error("Expected to receive streaming chunks")
	}

	t.Logf("Received %d chunks, full response: %s", len(chunks), response)
}

func TestIntegration_OpenAI_ConversationMemory(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	builder := NewOpenAI("gpt-4o-mini", apiKey).WithMemory()

	// First message
	response1, err := builder.Ask(ctx, "My name is Alice")
	if err != nil {
		t.Fatalf("First message failed: %v", err)
	}
	t.Logf("Response 1: %s", response1)

	// Second message that requires memory
	response2, err := builder.Ask(ctx, "What is my name?")
	if err != nil {
		t.Fatalf("Second message failed: %v", err)
	}
	t.Logf("Response 2: %s", response2)

	// Verify conversation history
	history := builder.GetHistory()
	if len(history) != 4 { // 2 user + 2 assistant messages
		t.Errorf("Expected 4 messages in history, got %d", len(history))
	}
}

func TestIntegration_OpenAI_ToolCalling(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()

	// Create a simple calculator tool
	tool := NewTool("calculate", "Perform basic math calculation").
		AddParameter("operation", "string", "Operation: add, subtract, multiply, divide", true).
		AddParameter("a", "number", "First number", true).
		AddParameter("b", "number", "Second number", true).
		WithHandler(func(args string) (string, error) {
			// In real integration test, parse args and calculate
			return `{"result": 42}`, nil
		})

	response, err := NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(tool).
		WithAutoExecute(true).
		Ask(ctx, "What is 20 plus 22?")

	if err != nil {
		t.Fatalf("Tool calling test failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response from tool calling")
	}

	t.Logf("Tool calling response: %s", response)
}

func TestIntegration_OpenAI_JSONSchema(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"age": map[string]interface{}{
				"type": "integer",
			},
		},
		"required":             []string{"name", "age"},
		"additionalProperties": false,
	}

	response, err := NewOpenAI("gpt-4o-mini", apiKey).
		WithJSONSchema("person", "Extract person info", schema, true).
		Ask(ctx, "John is 30 years old")

	if err != nil {
		t.Fatalf("JSON Schema test failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty JSON response")
	}

	t.Logf("JSON Schema response: %s", response)
}

func TestIntegration_OpenAI_Timeout(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()

	// Test with very short timeout (should timeout)
	_, err := NewOpenAI("gpt-4o-mini", apiKey).
		WithTimeout(1*time.Millisecond).
		Ask(ctx, "Write a long essay")

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if !IsTimeoutError(err) {
		t.Errorf("Expected timeout error, got: %v", err)
	}

	t.Logf("Timeout test passed: %v", err)
}

func TestIntegration_OpenAI_Retry(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()

	// Test with retry configuration
	response, err := NewOpenAI("gpt-4o-mini", apiKey).
		WithRetry(3).
		WithRetryDelay(1*time.Second).
		Ask(ctx, "Say 'retry test'")

	if err != nil {
		t.Fatalf("Retry test failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Retry test passed: %s", response)
}

func TestIntegration_Ollama_SimpleChat(t *testing.T) {
	// Check if Ollama is running
	ctx := context.Background()

	response, err := NewOllama("qwen2.5:3b").
		WithTimeout(5*time.Second).
		Ask(ctx, "Say 'test' and nothing else")

	if err != nil {
		t.Skipf("Ollama not available, skipping test: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response from Ollama")
	}

	t.Logf("Ollama response: %s", response)
}

func TestIntegration_Ollama_Streaming(t *testing.T) {
	ctx := context.Background()
	var chunks []string

	response, err := NewOllama("qwen2.5:3b").
		WithTimeout(10*time.Second).
		OnStream(func(content string) {
			chunks = append(chunks, content)
		}).
		Stream(ctx, "Count from 1 to 3")

	if err != nil {
		t.Skipf("Ollama not available, skipping test: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response from streaming")
	}

	if len(chunks) == 0 {
		t.Error("Expected to receive streaming chunks")
	}

	t.Logf("Ollama received %d chunks, full response: %s", len(chunks), response)
}

func TestIntegration_Ollama_ConversationMemory(t *testing.T) {
	ctx := context.Background()
	builder := NewOllama("qwen2.5:3b").
		WithTimeout(10 * time.Second).
		WithMemory()

	// First message
	response1, err := builder.Ask(ctx, "My favorite color is blue")
	if err != nil {
		t.Skipf("Ollama not available, skipping test: %v", err)
	}
	t.Logf("Response 1: %s", response1)

	// Second message that requires memory
	response2, err := builder.Ask(ctx, "What is my favorite color?")
	if err != nil {
		t.Fatalf("Second message failed: %v", err)
	}
	t.Logf("Response 2: %s", response2)

	// Verify conversation history
	history := builder.GetHistory()
	if len(history) != 4 { // 2 user + 2 assistant messages
		t.Errorf("Expected 4 messages in history, got %d", len(history))
	}
}

func TestIntegration_Concurrent_Requests(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	builder := NewOpenAI("gpt-4o-mini", apiKey)

	// Test concurrent requests with same builder (thread-safe)
	const numRequests = 5
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			_, err := builder.Ask(ctx, "Say 'concurrent test'")
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
	close(errors)

	// Check if any errors occurred
	for err := range errors {
		t.Errorf("Concurrent request failed: %v", err)
	}

	t.Logf("Successfully completed %d concurrent requests", numRequests)
}

func TestIntegration_ProductionConfiguration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()

	// Test production-ready configuration
	response, err := NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("You are a helpful assistant").
		WithTemperature(0.7).
		WithMaxTokens(100).
		WithMemory().
		WithMaxHistory(10).
		WithTimeout(30*time.Second).
		WithRetry(3).
		WithExponentialBackoff().
		Ask(ctx, "What is Go programming language?")

	if err != nil {
		t.Fatalf("Production configuration test failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Production config test passed: %s", response[:min(100, len(response))])
}

// TestIntegration_MemorySystem tests hierarchical memory with real API
func TestIntegration_MemorySystem(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()

	builder := NewOpenAI("gpt-4o-mini", apiKey).
		WithWorkingMemorySize(5).
		WithEpisodicMemory(0.7).
		WithSystem("You are a helpful assistant with excellent memory.")

	mem := builder.GetMemory()

	// Test 1: Important message
	t.Log("Test 1: Sending important message...")
	_, err := builder.Ask(ctx, "Remember that my birthday is May 5th")
	if err != nil {
		t.Fatalf("Failed to send important message: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	stats1 := mem.Stats(ctx)

	if stats1.EpisodicSize < 1 {
		t.Error("Expected important message to be stored in episodic memory")
	}
	t.Logf("Stats after important message: Working=%d, Episodic=%d, AvgImportance=%.2f",
		stats1.WorkingSize, stats1.EpisodicSize, stats1.AverageImportance)

	// Test 2: Casual message
	t.Log("Test 2: Sending casual message...")
	_, err = builder.Ask(ctx, "What's the weather like?")
	if err != nil {
		t.Fatalf("Failed to send casual message: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	stats2 := mem.Stats(ctx)

	// Episodic count should not increase significantly
	if stats2.EpisodicSize > stats1.EpisodicSize+1 {
		t.Errorf("Casual message shouldn't significantly increase episodic: was %d, now %d",
			stats1.EpisodicSize, stats2.EpisodicSize)
	}
	t.Logf("Stats after casual message: Working=%d, Episodic=%d",
		stats2.WorkingSize, stats2.EpisodicSize)

	// Test 3: Memory recall
	t.Log("Test 3: Testing memory recall...")
	response, err := builder.Ask(ctx, "What did I tell you about my birthday?")
	if err != nil {
		t.Fatalf("Failed to recall: %v", err)
	}

	t.Logf("Recall response: %s", response)

	// Verify final stats
	stats3 := mem.Stats(ctx)
	if stats3.TotalMessages < 3 {
		t.Errorf("Expected at least 3 messages processed, got %d", stats3.TotalMessages)
	}

	t.Logf("Final stats: Total=%d, Working=%d, Episodic=%d, Compressions=%d",
		stats3.TotalMessages, stats3.WorkingSize, stats3.EpisodicSize, stats3.CompressionCount)
}
