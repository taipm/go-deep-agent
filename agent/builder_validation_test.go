package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestConfigValidation tests comprehensive configuration validation at execution time
func TestConfigValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("toolChoice without tools should fail", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "fake-key").
			WithToolChoice("required")

		_, err := builder.Ask(ctx, "Hello")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if !errors.Is(err, ErrToolChoiceRequiresTools) {
			t.Errorf("Expected ErrToolChoiceRequiresTools, got: %v", err)
		}

		// Check error message contains helpful guidance
		errMsg := err.Error()
		expectedPhrases := []string{
			"tool choice requires tools",
			"WithTools",
			"Example:",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(errMsg, phrase) {
				t.Errorf("Error message missing '%s':\n%s", phrase, errMsg)
			}
		}
	})

	t.Run("toolChoice with tools should validate successfully", func(t *testing.T) {
		tool := &Tool{
			Name:        "test_tool",
			Description: "A test tool",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]interface{}{
						"type":        "string",
						"description": "Input text",
					},
				},
				"required": []string{"input"},
			},
		}

		builder := NewOpenAI("gpt-4o-mini", "fake-key").
			WithTools(tool).
			WithToolChoice("required")

		// Should not get validation error (will fail on API call, but that's expected)
		_, err := builder.Ask(ctx, "test")
		if err != nil && errors.Is(err, ErrToolChoiceRequiresTools) {
			t.Errorf("Should not get validation error with tools configured: %v", err)
		}
	})

	t.Run("validation works in Stream method", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "fake-key").
			WithToolChoice("auto")

		_, err := builder.Stream(ctx, "Hello")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if !errors.Is(err, ErrToolChoiceRequiresTools) {
			t.Errorf("Expected ErrToolChoiceRequiresTools, got: %v", err)
		}
	})

	t.Run("no toolChoice is valid", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "fake-key")

		// Should not get validation error for toolChoice
		// (will fail on API call for other reasons)
		_, err := builder.Ask(ctx, "Hello")
		if err != nil && errors.Is(err, ErrToolChoiceRequiresTools) {
			t.Errorf("Should not get toolChoice validation error when toolChoice not set: %v", err)
		}
	})
}
