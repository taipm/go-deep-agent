package agent

import (
	"testing"
)

func TestMessageHelpers(t *testing.T) {
	tests := []struct {
		name        string
		message     Message
		wantRole    string
		wantContent string
	}{
		{
			name:        "System message",
			message:     System("You are a helpful assistant"),
			wantRole:    "system",
			wantContent: "You are a helpful assistant",
		},
		{
			name:        "User message",
			message:     User("Hello, world!"),
			wantRole:    "user",
			wantContent: "Hello, world!",
		},
		{
			name:        "Assistant message",
			message:     Assistant("Hi there!"),
			wantRole:    "assistant",
			wantContent: "Hi there!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.message.Role != tt.wantRole {
				t.Errorf("Role = %v, want %v", tt.message.Role, tt.wantRole)
			}
			if tt.message.Content != tt.wantContent {
				t.Errorf("Content = %v, want %v", tt.message.Content, tt.wantContent)
			}
		})
	}
}

func TestConvertMessages(t *testing.T) {
	messages := []Message{
		System("You are helpful"),
		User("Hello"),
		Assistant("Hi!"),
	}

	result := convertMessages(messages)

	if len(result) != 3 {
		t.Errorf("convertMessages() returned %d messages, want 3", len(result))
	}
}

func TestBuilder_New(t *testing.T) {
	builder := New(ProviderOpenAI, "gpt-4o-mini")

	if builder.provider != ProviderOpenAI {
		t.Errorf("provider = %v, want %v", builder.provider, ProviderOpenAI)
	}
	if builder.model != "gpt-4o-mini" {
		t.Errorf("model = %v, want %v", builder.model, "gpt-4o-mini")
	}
	if builder.autoMemory {
		t.Error("autoMemory should be false by default")
	}
}

func TestBuilder_NewOpenAI(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key")

	if builder.provider != ProviderOpenAI {
		t.Errorf("provider = %v, want %v", builder.provider, ProviderOpenAI)
	}
	if builder.model != "gpt-4o-mini" {
		t.Errorf("model = %v, want %v", builder.model, "gpt-4o-mini")
	}
	if builder.apiKey != "test-key" {
		t.Errorf("apiKey = %v, want %v", builder.apiKey, "test-key")
	}
}

func TestBuilder_NewOllama(t *testing.T) {
	builder := NewOllama("qwen3:1.7b")

	if builder.provider != ProviderOllama {
		t.Errorf("provider = %v, want %v", builder.provider, ProviderOllama)
	}
	if builder.model != "qwen3:1.7b" {
		t.Errorf("model = %v, want %v", builder.model, "qwen3:1.7b")
	}
	if builder.baseURL != "http://localhost:11434/v1" {
		t.Errorf("baseURL = %v, want %v", builder.baseURL, "http://localhost:11434/v1")
	}
}

func TestBuilder_WithAPIKey(t *testing.T) {
	builder := New(ProviderOpenAI, "gpt-4o-mini").
		WithAPIKey("new-key")

	if builder.apiKey != "new-key" {
		t.Errorf("apiKey = %v, want %v", builder.apiKey, "new-key")
	}
}

func TestBuilder_WithBaseURL(t *testing.T) {
	builder := NewOllama("qwen3:1.7b").
		WithBaseURL("http://custom:8080/v1")

	if builder.baseURL != "http://custom:8080/v1" {
		t.Errorf("baseURL = %v, want %v", builder.baseURL, "http://custom:8080/v1")
	}
}

func TestBuilder_WithSystem(t *testing.T) {
	builder := NewOllama("qwen3:1.7b").
		WithSystem("You are helpful")

	if builder.systemPrompt != "You are helpful" {
		t.Errorf("systemPrompt = %v, want %v", builder.systemPrompt, "You are helpful")
	}
}

func TestBuilder_WithMemory(t *testing.T) {
	builder := NewOllama("qwen3:1.7b").
		WithMemory()

	if !builder.autoMemory {
		t.Error("autoMemory should be true after WithMemory()")
	}
}

func TestBuilder_WithMessages(t *testing.T) {
	messages := []Message{
		User("Hello"),
		Assistant("Hi!"),
	}

	builder := NewOllama("qwen3:1.7b").
		WithMessages(messages)

	if len(builder.messages) != 2 {
		t.Errorf("messages length = %d, want 2", len(builder.messages))
	}
}

func TestBuilder_Chaining(t *testing.T) {
	// Test that builder methods can be chained
	builder := NewOllama("qwen3:1.7b").
		WithSystem("You are helpful").
		WithMemory().
		WithBaseURL("http://custom:8080/v1")

	if builder.systemPrompt != "You are helpful" {
		t.Error("systemPrompt not set correctly")
	}
	if !builder.autoMemory {
		t.Error("autoMemory not set correctly")
	}
	if builder.baseURL != "http://custom:8080/v1" {
		t.Error("baseURL not set correctly")
	}
}

func TestBuilder_EnsureClientOpenAI_NoAPIKey(t *testing.T) {
	builder := New(ProviderOpenAI, "gpt-4o-mini")
	// Don't set API key

	err := builder.ensureClient()
	if err == nil {
		t.Error("ensureClient() should return error when API key is missing for OpenAI")
	}
}

func TestBuilder_BuildMessages(t *testing.T) {
	builder := NewOllama("qwen3:1.7b").
		WithSystem("System prompt").
		WithMessages([]Message{
			User("Previous message"),
			Assistant("Previous response"),
		})

	messages := builder.buildMessages("New message")

	// Should have: system + 2 history + current user = 4 messages
	if len(messages) != 4 {
		t.Errorf("buildMessages() returned %d messages, want 4", len(messages))
	}
}

func TestBuilder_getToolNames(t *testing.T) {
	t.Run("no tools registered", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini")
		names := builder.getToolNames()

		if len(names) != 0 {
			t.Errorf("getToolNames() returned %d names, expected 0", len(names))
		}
	})

	t.Run("single tool", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini")
		builder.tools = []*Tool{
			{Name: "math", Description: "Math tool"},
		}
		names := builder.getToolNames()

		if len(names) != 1 {
			t.Errorf("getToolNames() returned %d names, expected 1", len(names))
		}
		if names[0] != "math" {
			t.Errorf("getToolNames()[0] = %q, expected %q", names[0], "math")
		}
	})

	t.Run("multiple tools", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini")
		builder.tools = []*Tool{
			{Name: "math", Description: "Math tool"},
			{Name: "datetime", Description: "DateTime tool"},
			{Name: "filesystem", Description: "FileSystem tool"},
		}
		names := builder.getToolNames()

		expected := []string{"math", "datetime", "filesystem"}
		if len(names) != len(expected) {
			t.Errorf("getToolNames() returned %d names, expected %d", len(names), len(expected))
		}

		for i, exp := range expected {
			if names[i] != exp {
				t.Errorf("getToolNames()[%d] = %q, expected %q", i, names[i], exp)
			}
		}
	})
}
