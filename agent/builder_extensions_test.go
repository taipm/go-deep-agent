package agent

import (
	"testing"

	"github.com/openai/openai-go/v3"
)

// TestWithTemperature tests temperature configuration
func TestWithTemperature(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTemperature(0.7)

	if builder.temperature == nil {
		t.Fatal("Expected temperature to be set")
	}
	if *builder.temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", *builder.temperature)
	}
}

// TestWithTopP tests top_p configuration
func TestWithTopP(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTopP(0.9)

	if builder.topP == nil {
		t.Fatal("Expected topP to be set")
	}
	if *builder.topP != 0.9 {
		t.Errorf("Expected topP 0.9, got %v", *builder.topP)
	}
}

// TestWithMaxTokens tests max_tokens configuration
func TestWithMaxTokens(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMaxTokens(1000)

	if builder.maxTokens == nil {
		t.Fatal("Expected maxTokens to be set")
	}
	if *builder.maxTokens != 1000 {
		t.Errorf("Expected maxTokens 1000, got %v", *builder.maxTokens)
	}
}

// TestWithPresencePenalty tests presence_penalty configuration
func TestWithPresencePenalty(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithPresencePenalty(0.5)

	if builder.presencePenalty == nil {
		t.Fatal("Expected presencePenalty to be set")
	}
	if *builder.presencePenalty != 0.5 {
		t.Errorf("Expected presencePenalty 0.5, got %v", *builder.presencePenalty)
	}
}

// TestWithFrequencyPenalty tests frequency_penalty configuration
func TestWithFrequencyPenalty(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithFrequencyPenalty(0.3)

	if builder.frequencyPenalty == nil {
		t.Fatal("Expected frequencyPenalty to be set")
	}
	if *builder.frequencyPenalty != 0.3 {
		t.Errorf("Expected frequencyPenalty 0.3, got %v", *builder.frequencyPenalty)
	}
}

// TestWithSeed tests seed configuration
func TestWithSeed(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSeed(12345)

	if builder.seed == nil {
		t.Fatal("Expected seed to be set")
	}
	if *builder.seed != 12345 {
		t.Errorf("Expected seed 12345, got %v", *builder.seed)
	}
}

// TestWithLogprobs tests logprobs configuration
func TestWithLogprobs(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithLogprobs(true)

	if builder.logprobs == nil {
		t.Fatal("Expected logprobs to be set")
	}
	if *builder.logprobs != true {
		t.Errorf("Expected logprobs true, got %v", *builder.logprobs)
	}
}

// TestWithTopLogprobs tests top_logprobs configuration
func TestWithTopLogprobs(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTopLogprobs(5)

	if builder.topLogprobs == nil {
		t.Fatal("Expected topLogprobs to be set")
	}
	if *builder.topLogprobs != 5 {
		t.Errorf("Expected topLogprobs 5, got %v", *builder.topLogprobs)
	}
}

// TestWithMultipleChoices tests n (multiple choices) configuration
func TestWithMultipleChoices(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMultipleChoices(3)

	if builder.n == nil {
		t.Fatal("Expected n to be set")
	}
	if *builder.n != 3 {
		t.Errorf("Expected n 3, got %v", *builder.n)
	}
}

// TestOnStream tests stream callback configuration
func TestOnStream(t *testing.T) {
	called := false
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		OnStream(func(content string) {
			called = true
		})

	if builder.onStream == nil {
		t.Fatal("Expected onStream to be set")
	}

	// Test callback
	builder.onStream("test")
	if !called {
		t.Error("Expected onStream callback to be called")
	}
}

// TestOnToolCall tests tool call callback configuration
func TestOnToolCall(t *testing.T) {
	called := false
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		OnToolCall(func(tool openai.FinishedChatCompletionToolCall) {
			called = true
		})

	if builder.onToolCall == nil {
		t.Fatal("Expected onToolCall to be set")
	}

	// Test callback
	builder.onToolCall(openai.FinishedChatCompletionToolCall{})
	if !called {
		t.Error("Expected onToolCall callback to be called")
	}
}

// TestOnRefusal tests refusal callback configuration
func TestOnRefusal(t *testing.T) {
	called := false
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		OnRefusal(func(refusal string) {
			called = true
		})

	if builder.onRefusal == nil {
		t.Fatal("Expected onRefusal to be set")
	}

	// Test callback
	builder.onRefusal("test")
	if !called {
		t.Error("Expected onRefusal callback to be called")
	}
}

// TestWithToolSingle tests adding a single tool
func TestWithToolSingle(t *testing.T) {
	tool := NewTool("test_tool", "A test tool")
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTool(tool)

	if len(builder.tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(builder.tools))
	}
	if builder.tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", builder.tools[0].Name)
	}
}

// TestWithToolsMultiple tests adding multiple tools
func TestWithToolsMultiple(t *testing.T) {
	tool1 := NewTool("tool1", "First tool")
	tool2 := NewTool("tool2", "Second tool")
	tool3 := NewTool("tool3", "Third tool")

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTools(tool1, tool2, tool3)

	if len(builder.tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(builder.tools))
	}
}

// TestWithToolChaining tests adding tools via chaining
func TestWithToolChaining(t *testing.T) {
	tool1 := NewTool("tool1", "First tool")
	tool2 := NewTool("tool2", "Second tool")

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTool(tool1).
		WithTool(tool2)

	if len(builder.tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(builder.tools))
	}
}

// TestWithAutoExecute tests auto-execute configuration
func TestWithAutoExecute(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithAutoExecute(true)

	if !builder.autoExecute {
		t.Error("Expected autoExecute to be true")
	}
	if builder.maxToolRounds != 5 {
		t.Errorf("Expected default maxToolRounds 5, got %d", builder.maxToolRounds)
	}
}

// TestWithMaxToolRounds tests max tool rounds configuration
func TestWithMaxToolRounds(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMaxToolRounds(10)

	if builder.maxToolRounds != 10 {
		t.Errorf("Expected maxToolRounds 10, got %d", builder.maxToolRounds)
	}
}

// TestWithJSONModeConfig tests JSON mode configuration
func TestWithJSONModeConfig(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithJSONMode()

	if builder.responseFormat == nil {
		t.Fatal("Expected responseFormat to be set")
	}
	if builder.responseFormat.OfJSONObject == nil {
		t.Error("Expected OfJSONObject to be set")
	}
}

// TestWithJSONSchemaConfig tests JSON schema configuration
func TestWithJSONSchemaConfig(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
		},
		"required":             []string{"name"},
		"additionalProperties": false,
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithJSONSchema("test_schema", "A test schema", schema, true)

	if builder.responseFormat == nil {
		t.Fatal("Expected responseFormat to be set")
	}
	if builder.responseFormat.OfJSONSchema == nil {
		t.Fatal("Expected OfJSONSchema to be set")
	}

	jsonSchema := builder.responseFormat.OfJSONSchema
	if jsonSchema.JSONSchema.Name != "test_schema" {
		t.Errorf("Expected schema name 'test_schema', got '%s'", jsonSchema.JSONSchema.Name)
	}
}

// TestWithResponseFormat tests custom response format
func TestWithResponseFormat(t *testing.T) {
	format := &openai.ChatCompletionNewParamsResponseFormatUnion{
		OfText: &openai.ResponseFormatTextParam{},
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithResponseFormat(format)

	if builder.responseFormat == nil {
		t.Fatal("Expected responseFormat to be set")
	}
	if builder.responseFormat.OfText == nil {
		t.Error("Expected OfText to be set")
	}
}

// TestAdvancedParametersChaining tests chaining multiple advanced parameters
func TestAdvancedParametersChaining(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTemperature(0.7).
		WithTopP(0.9).
		WithMaxTokens(1000).
		WithPresencePenalty(0.5).
		WithFrequencyPenalty(0.3).
		WithSeed(12345).
		WithLogprobs(true).
		WithTopLogprobs(5).
		WithMultipleChoices(3)

	// Verify all parameters are set
	if builder.temperature == nil || *builder.temperature != 0.7 {
		t.Error("Temperature not set correctly")
	}
	if builder.topP == nil || *builder.topP != 0.9 {
		t.Error("TopP not set correctly")
	}
	if builder.maxTokens == nil || *builder.maxTokens != 1000 {
		t.Error("MaxTokens not set correctly")
	}
	if builder.presencePenalty == nil || *builder.presencePenalty != 0.5 {
		t.Error("PresencePenalty not set correctly")
	}
	if builder.frequencyPenalty == nil || *builder.frequencyPenalty != 0.3 {
		t.Error("FrequencyPenalty not set correctly")
	}
	if builder.seed == nil || *builder.seed != 12345 {
		t.Error("Seed not set correctly")
	}
	if builder.logprobs == nil || *builder.logprobs != true {
		t.Error("Logprobs not set correctly")
	}
	if builder.topLogprobs == nil || *builder.topLogprobs != 5 {
		t.Error("TopLogprobs not set correctly")
	}
	if builder.n == nil || *builder.n != 3 {
		t.Error("N not set correctly")
	}
}

// TestBuildParams tests buildParams method
func TestBuildParams(t *testing.T) {
	tool := NewTool("test_tool", "Test")
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTemperature(0.8).
		WithMaxTokens(500).
		WithTool(tool)

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	params := builder.buildParams(messages)

	// Verify model
	if string(params.Model) != "gpt-4o-mini" {
		t.Errorf("Expected model 'gpt-4o-mini', got '%v'", params.Model)
	}

	// Verify messages
	if len(params.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(params.Messages))
	}

	// We can't easily verify temperature/maxTokens due to SDK's Opt types,
	// but we can verify tools are included
	if len(params.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(params.Tools))
	}
}

// TestBuildMessagesWithSystem tests buildMessages with system prompt
func TestBuildMessagesWithSystem(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSystem("You are helpful")

	messages := builder.buildMessages("Hello")

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// First should be system, second should be user
	// We can't easily check the content due to SDK's union types,
	// but we verified this works in earlier tests
}

// TestBuildMessagesWithHistory tests buildMessages with message history
func TestBuildMessagesWithHistory(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMessages([]Message{
			User("Previous message"),
			Assistant("Previous response"),
		})

	messages := builder.buildMessages("New message")

	// Should have: previous user, previous assistant, new user
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}
}

// TestAddMessage tests the addMessage method
func TestAddMessage(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key")

	if len(builder.messages) != 0 {
		t.Fatal("Expected empty messages initially")
	}

	builder.addMessage(User("Test"))

	if len(builder.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(builder.messages))
	}
}

// TestGetHistory tests retrieving conversation history
func TestGetHistory(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMessages([]Message{
			User("Hello"),
			Assistant("Hi there"),
			User("How are you?"),
		})

	history := builder.GetHistory()

	if len(history) != 3 {
		t.Errorf("Expected 3 messages in history, got %d", len(history))
	}

	// Verify it's a copy, not a reference
	history[0] = User("Modified")
	if builder.messages[0].Content == "Modified" {
		t.Error("GetHistory should return a copy, not a reference")
	}
}

// TestSetHistory tests replacing conversation history
func TestSetHistory(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMessages([]Message{
			User("Old message"),
		})

	newHistory := []Message{
		User("New message 1"),
		Assistant("Response 1"),
		User("New message 2"),
	}

	builder.SetHistory(newHistory)

	if len(builder.messages) != 3 {
		t.Errorf("Expected 3 messages after SetHistory, got %d", len(builder.messages))
	}

	if builder.messages[0].Content != "New message 1" {
		t.Errorf("Expected first message to be 'New message 1', got %s", builder.messages[0].Content)
	}
}

// TestClear tests clearing conversation history
func TestClear(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSystem("You are helpful").
		WithMessages([]Message{
			User("Message 1"),
			Assistant("Response 1"),
			User("Message 2"),
		})

	// Verify messages exist
	if len(builder.messages) != 3 {
		t.Fatalf("Expected 3 messages before clear, got %d", len(builder.messages))
	}

	// Clear messages
	builder.Clear()

	// Verify messages cleared
	if len(builder.messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(builder.messages))
	}

	// Verify system prompt preserved
	if builder.systemPrompt != "You are helpful" {
		t.Error("System prompt should be preserved after Clear()")
	}
}

// TestWithMaxHistory tests setting maximum history limit
func TestWithMaxHistory(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMaxHistory(3)

	if builder.maxHistory != 3 {
		t.Errorf("Expected maxHistory to be 3, got %d", builder.maxHistory)
	}
}

// TestMaxHistoryAutoTruncate tests automatic history truncation
func TestMaxHistoryAutoTruncate(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMaxHistory(3)

	// Add 5 messages
	builder.addMessage(User("Message 1"))
	builder.addMessage(Assistant("Response 1"))
	builder.addMessage(User("Message 2"))
	builder.addMessage(Assistant("Response 2"))
	builder.addMessage(User("Message 3"))

	// Should only keep the last 3
	if len(builder.messages) != 3 {
		t.Errorf("Expected 3 messages (truncated), got %d", len(builder.messages))
	}

	// Verify oldest messages were removed (FIFO)
	if builder.messages[0].Content != "Message 2" {
		t.Errorf("Expected first message to be 'Message 2', got %s", builder.messages[0].Content)
	}

	if builder.messages[2].Content != "Message 3" {
		t.Errorf("Expected last message to be 'Message 3', got %s", builder.messages[2].Content)
	}
}

// TestMaxHistoryUnlimited tests unlimited history (maxHistory = 0)
func TestMaxHistoryUnlimited(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key")
	// Default maxHistory is 0 (unlimited)

	// Add many messages
	for i := 0; i < 10; i++ {
		builder.addMessage(User("Message"))
	}

	// All should be kept
	if len(builder.messages) != 10 {
		t.Errorf("Expected 10 messages (unlimited), got %d", len(builder.messages))
	}
}

// TestConversationManagementChaining tests method chaining
func TestConversationManagementChaining(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMemory().
		WithMaxHistory(5).
		WithSystem("You are helpful").
		Clear().
		SetHistory([]Message{User("Test")})

	if !builder.autoMemory {
		t.Error("Expected autoMemory to be true")
	}

	if builder.maxHistory != 5 {
		t.Errorf("Expected maxHistory 5, got %d", builder.maxHistory)
	}

	if len(builder.messages) != 1 {
		t.Errorf("Expected 1 message after SetHistory, got %d", len(builder.messages))
	}
}
