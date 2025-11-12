package agent

import (
	"testing"
)

func TestBuildReActMetaTools(t *testing.T) {
	t.Run("think and final_answer without registered tools", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini")

		tools := builder.buildReActMetaTools()

		// Should return 2 tools: think() + final_answer()
		// (no use_tool() since no tools registered)
		if len(tools) != 2 {
			t.Errorf("buildReActMetaTools() returned %d tools, expected 2", len(tools))
		}
	})

	t.Run("all three meta-tools with registered tools", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini")
		builder.tools = []*Tool{
			{Name: "math", Description: "Math tool"},
			{Name: "datetime", Description: "DateTime tool"},
		}

		tools := builder.buildReActMetaTools()

		// Should return 3 tools: think() + use_tool() + final_answer()
		if len(tools) != 3 {
			t.Errorf("buildReActMetaTools() returned %d tools, expected 3", len(tools))
		}
	})
}

func TestReActModeBuilderMethods(t *testing.T) {
	t.Run("WithReActNativeMode sets native mode", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini").
			WithReActMode(true).
			WithReActNativeMode()

		if builder.reactConfig.Mode != ReActModeNative {
			t.Errorf("Mode = %v, expected ReActModeNative", builder.reactConfig.Mode)
		}
	})

	t.Run("WithReActTextMode sets text mode", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini").
			WithReActMode(true).
			WithReActTextMode()

		if builder.reactConfig.Mode != ReActModeText {
			t.Errorf("Mode = %v, expected ReActModeText", builder.reactConfig.Mode)
		}
	})

	t.Run("default mode is native", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini").
			WithReActMode(true)

		if builder.reactConfig.Mode != ReActModeNative {
			t.Errorf("Default Mode = %v, expected ReActModeNative", builder.reactConfig.Mode)
		}
	})
}

func TestProgressiveUrgencyReminders(t *testing.T) {
	t.Run("reminders enabled with proper message injection", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini").
			WithReActMode(true).
			WithReActIterationReminders(true).
			WithReActMaxIterations(3)

		// Verify config
		if !builder.reactConfig.EnableIterationReminders {
			t.Error("Expected EnableIterationReminders=true")
		}

		if builder.reactConfig.MaxIterations != 3 {
			t.Errorf("Expected MaxIterations=3, got %d", builder.reactConfig.MaxIterations)
		}

		// Note: Full integration test with message injection would require mock LLM
		// This validates configuration; actual message injection tested in integration tests
	})

	t.Run("reminders disabled", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini").
			WithReActMode(true).
			WithReActIterationReminders(false).
			WithReActMaxIterations(3)

		if builder.reactConfig.EnableIterationReminders {
			t.Error("Expected EnableIterationReminders=false")
		}
	})
}

func TestAutoFallbackMechanism(t *testing.T) {
	t.Run("auto-fallback enabled", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini").
			WithReActMode(true).
			WithReActAutoFallback(true).
			WithReActMaxIterations(3)

		if !builder.reactConfig.EnableAutoFallback {
			t.Error("Expected EnableAutoFallback=true")
		}
	})

	t.Run("synthesizeFallbackAnswer with no steps", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini").
			WithReActMode(true)

		answer := builder.synthesizeFallbackAnswer([]ReActStep{}, 5)

		if answer == "" {
			t.Error("Expected non-empty fallback answer")
		}

		// Should mention max iterations
		if len(answer) < 20 {
			t.Error("Fallback answer seems too short")
		}
	})

	t.Run("synthesizeFallbackAnswer with steps", func(t *testing.T) {
		builder := New(ProviderOpenAI, "gpt-4o-mini").
			WithReActMode(true)

		steps := []ReActStep{
			{Type: StepTypeThought, Content: "I need to analyze this"},
			{Type: StepTypeAction, Content: "search(query)", Tool: "search"},
			{Type: StepTypeObservation, Content: "Found some data"},
			{Type: StepTypeThought, Content: "Based on the data..."},
		}

		answer := builder.synthesizeFallbackAnswer(steps, 5)

		if answer == "" {
			t.Error("Expected non-empty fallback answer")
		}

		// Should mention thoughts and actions
		if len(answer) < 50 {
			t.Error("Fallback answer seems too short given 4 steps")
		}
	})
}
