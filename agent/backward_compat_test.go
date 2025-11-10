package agent

import (
	"context"
	"testing"

	"github.com/taipm/go-deep-agent/agent/memory"
)

// TestBackwardCompatibility_V056_SimpleUsage tests v0.5.6 basic usage still works
func TestBackwardCompatibility_V056_SimpleUsage(t *testing.T) {
	// This is exactly how v0.5.6 users would create an agent
	builder := NewOpenAI("gpt-4o-mini", "test-key")

	// Should have memory enabled by default (new behavior, but backward compatible)
	if builder.memory == nil {
		t.Error("Expected memory to be initialized by default")
	}

	// Should be able to access memory
	mem := builder.GetMemory()
	if mem == nil {
		t.Error("GetMemory() should return non-nil")
	}

	// Default config should be sensible
	config := mem.GetConfig()
	if config.WorkingCapacity <= 0 {
		t.Error("Expected positive working capacity")
	}

	// Episodic should be enabled by default in new version
	if !config.EpisodicEnabled {
		t.Error("Expected episodic memory enabled by default")
	}
}

// TestBackwardCompatibility_DisableMemory tests opting out of hierarchical memory
func TestBackwardCompatibility_DisableMemory(t *testing.T) {
	// Users who want v0.5.6 behavior can disable hierarchical memory
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		DisableMemory()

	// memoryEnabled should be false
	if builder.memoryEnabled {
		t.Error("memoryEnabled flag should be false after DisableMemory()")
	}

	// GetMemory should still work
	mem := builder.GetMemory()
	if mem == nil {
		t.Error("GetMemory() should still return memory instance")
	}
}

// TestBackwardCompatibility_WithMessages tests existing message API
func TestBackwardCompatibility_WithMessages(t *testing.T) {
	ctx := context.Background()

	// v0.5.6 users could set messages directly
	messages := []Message{
		User("Hello"),
		Assistant("Hi there!"),
		User("How are you?"),
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithMessages(messages)

	// Messages should be stored in builder
	if len(builder.messages) != 3 {
		t.Errorf("Expected 3 messages in builder, got %d", len(builder.messages))
	}

	// Messages are added to memory when Ask/Stream is called, not in WithMessages
	// So we manually add them to verify memory works
	mem := builder.GetMemory()
	for _, msg := range messages {
		mem.Add(ctx, memory.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	stats := mem.Stats(ctx)
	if stats.WorkingSize != 3 {
		t.Errorf("Expected 3 messages in working memory, got %d", stats.WorkingSize)
	}
}

// TestBackwardCompatibility_WithSystem tests system prompt
func TestBackwardCompatibility_WithSystem(t *testing.T) {
	// v0.5.6 users used WithSystem
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSystem("You are a helpful assistant")

	if builder.systemPrompt == "" {
		t.Error("System prompt should be set")
	}

	// Should still work with new memory features
	builder = builder.WithEpisodicMemory(0.7)

	mem := builder.GetMemory()
	if mem == nil {
		t.Error("Memory should still work after WithSystem")
	}
}

// TestBackwardCompatibility_MethodChaining tests fluent API
func TestBackwardCompatibility_MethodChaining(t *testing.T) {
	// v0.5.6 users could chain methods
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSystem("You are helpful").
		WithTemperature(0.7).
		WithMaxTokens(100)

	if builder.systemPrompt == "" {
		t.Error("System prompt should be set")
	}

	// Should still be chainable with new methods
	builder = builder.
		WithEpisodicMemory(0.8).
		WithWorkingMemorySize(20)

	mem := builder.GetMemory()
	config := mem.GetConfig()

	if config.EpisodicThreshold != 0.8 {
		t.Errorf("Expected threshold 0.8, got %.2f", config.EpisodicThreshold)
	}

	if config.WorkingCapacity != 20 {
		t.Errorf("Expected capacity 20, got %d", config.WorkingCapacity)
	}
}

// TestBackwardCompatibility_DefaultBehavior tests default initialization
func TestBackwardCompatibility_DefaultBehavior(t *testing.T) {
	ctx := context.Background()

	// Simple creation like v0.5.6
	builder := NewOpenAI("gpt-4o-mini", "test-key")

	// Add some messages (new behavior auto-scores importance)
	mem := builder.GetMemory()
	mem.Add(ctx, memory.Message{
		Role:    "user",
		Content: "Hello",
	})

	stats := mem.Stats(ctx)

	// Should have working memory
	if stats.WorkingSize != 1 {
		t.Errorf("Expected 1 message in working memory, got %d", stats.WorkingSize)
	}

	// Total messages tracked (new feature)
	if stats.TotalMessages != 1 {
		t.Errorf("Expected 1 total message, got %d", stats.TotalMessages)
	}
}

// TestBackwardCompatibility_MultipleBuilders tests multiple instances
func TestBackwardCompatibility_MultipleBuilders(t *testing.T) {
	// v0.5.6 users could create multiple builders
	builder1 := NewOpenAI("gpt-4o-mini", "key1")
	builder2 := NewOpenAI("gpt-4o-mini", "key2")

	// Each should have separate memory
	mem1 := builder1.GetMemory()
	mem2 := builder2.GetMemory()

	if mem1 == mem2 {
		t.Error("Each builder should have separate memory instance")
	}

	// Config changes to one shouldn't affect the other
	builder1.WithEpisodicMemory(0.5)
	builder2.WithEpisodicMemory(0.9)

	config1 := mem1.GetConfig()
	config2 := mem2.GetConfig()

	if config1.EpisodicThreshold == config2.EpisodicThreshold {
		t.Error("Configs should be independent")
	}
}

// TestBackwardCompatibility_MessageHelpers tests message helper functions
func TestBackwardCompatibility_MessageHelpers(t *testing.T) {
	// v0.5.6 had User(), Assistant(), System() helpers
	userMsg := User("Hello")
	if userMsg.Role != "user" {
		t.Error("User() should create user message")
	}

	assistantMsg := Assistant("Hi there")
	if assistantMsg.Role != "assistant" {
		t.Error("Assistant() should create assistant message")
	}

	systemMsg := System("You are helpful")
	if systemMsg.Role != "system" {
		t.Error("System() should create system message")
	}

	// These should still work with memory
	ctx := context.Background()
	builder := NewOpenAI("gpt-4o-mini", "test-key")
	mem := builder.GetMemory()

	// Convert and add to memory
	mem.Add(ctx, memory.Message{
		Role:    userMsg.Role,
		Content: userMsg.Content,
	})

	stats := mem.Stats(ctx)
	if stats.WorkingSize != 1 {
		t.Error("Message helpers should work with memory")
	}
}

// TestBackwardCompatibility_ConfigUpdate tests updating configuration
func TestBackwardCompatibility_ConfigUpdate(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key")
	mem := builder.GetMemory()

	// Get initial config
	config := mem.GetConfig()
	originalCapacity := config.WorkingCapacity

	// Update config
	config.WorkingCapacity = 50
	err := mem.SetConfig(config)
	if err != nil {
		t.Fatalf("SetConfig should not error: %v", err)
	}

	// Verify update
	newConfig := mem.GetConfig()
	if newConfig.WorkingCapacity != 50 {
		t.Errorf("Expected capacity 50, got %d", newConfig.WorkingCapacity)
	}

	if newConfig.WorkingCapacity == originalCapacity {
		t.Error("Config should have been updated")
	}
}

// TestBackwardCompatibility_NoAPIKey tests graceful handling
func TestBackwardCompatibility_NoAPIKey(t *testing.T) {
	// v0.5.6 users might create builder without key initially
	builder := NewOpenAI("gpt-4o-mini", "")

	// Should not panic
	mem := builder.GetMemory()
	if mem == nil {
		t.Error("Memory should be initialized even without API key")
	}

	// Memory operations should work
	ctx := context.Background()
	stats := mem.Stats(ctx)

	if stats.WorkingCapacity == 0 {
		t.Error("Memory should have default config")
	}
}
