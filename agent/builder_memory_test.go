package agent

import (
	"testing"

	"github.com/taipm/go-deep-agent/agent/memory"
)

func TestBuilder_WithEpisodicMemory(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithEpisodicMemory(0.8)

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()
	if !config.EpisodicEnabled {
		t.Error("Expected episodic memory to be enabled")
	}

	if config.EpisodicThreshold != 0.8 {
		t.Errorf("Expected threshold 0.8, got %.2f", config.EpisodicThreshold)
	}

	if !config.ImportanceScoring {
		t.Error("Expected importance scoring to be enabled")
	}

	if !builder.memoryEnabled {
		t.Error("Expected memory to be enabled in builder")
	}
}

func TestBuilder_WithImportanceWeights(t *testing.T) {
	weights := memory.DefaultImportanceWeights()
	weights.ExplicitRemember = 2.0
	weights.PersonalInfo = 1.5

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithImportanceWeights(weights)

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()
	if config.ImportanceWeights.ExplicitRemember != 2.0 {
		t.Errorf("Expected ExplicitRemember weight 2.0, got %.2f", config.ImportanceWeights.ExplicitRemember)
	}

	if config.ImportanceWeights.PersonalInfo != 1.5 {
		t.Errorf("Expected PersonalInfo weight 1.5, got %.2f", config.ImportanceWeights.PersonalInfo)
	}

	if !config.ImportanceScoring {
		t.Error("Expected importance scoring to be enabled")
	}
}

func TestBuilder_WithWorkingMemorySize(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithWorkingMemorySize(50)

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()
	if config.WorkingCapacity != 50 {
		t.Errorf("Expected working capacity 50, got %d", config.WorkingCapacity)
	}
}

func TestBuilder_WithSemanticMemory(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSemanticMemory()

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()
	if !config.SemanticEnabled {
		t.Error("Expected semantic memory to be enabled")
	}
}

func TestBuilder_MemoryMethodChaining(t *testing.T) {
	weights := memory.DefaultImportanceWeights()
	weights.ExplicitRemember = 2.0

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithWorkingMemorySize(30).
		WithEpisodicMemory(0.7).
		WithImportanceWeights(weights).
		WithSemanticMemory()

	mem := builder.GetMemory()
	if mem == nil {
		t.Fatal("Expected memory to be initialized")
	}

	config := mem.GetConfig()

	// Verify all configs were applied
	if config.WorkingCapacity != 30 {
		t.Errorf("Expected working capacity 30, got %d", config.WorkingCapacity)
	}

	if !config.EpisodicEnabled {
		t.Error("Expected episodic memory to be enabled")
	}

	if config.EpisodicThreshold != 0.7 {
		t.Errorf("Expected threshold 0.7, got %.2f", config.EpisodicThreshold)
	}

	if config.ImportanceWeights.ExplicitRemember != 2.0 {
		t.Errorf("Expected ExplicitRemember weight 2.0, got %.2f", config.ImportanceWeights.ExplicitRemember)
	}

	if !config.SemanticEnabled {
		t.Error("Expected semantic memory to be enabled")
	}
}

func TestBuilder_MemoryInitialization(t *testing.T) {
	// Test that memory is initialized even when not explicitly created
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithEpisodicMemory(0.5)

	if builder.memory == nil {
		t.Error("Expected memory to be auto-initialized")
	}

	// Test multiple config calls on same builder
	builder.WithWorkingMemorySize(100)

	config := builder.GetMemory().GetConfig()
	if config.WorkingCapacity != 100 {
		t.Errorf("Expected capacity 100 after second config call, got %d", config.WorkingCapacity)
	}

	if config.EpisodicThreshold != 0.5 {
		t.Errorf("Expected threshold 0.5 to be preserved, got %.2f", config.EpisodicThreshold)
	}
}
