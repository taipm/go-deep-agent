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
