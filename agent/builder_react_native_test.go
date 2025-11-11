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