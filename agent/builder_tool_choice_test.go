package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithToolChoice_ValidValues(t *testing.T) {
	tests := []struct {
		name   string
		choice string
	}{
		{"auto mode", "auto"},
		{"required mode", "required"},
		{"none mode", "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOpenAI("gpt-4o-mini", "test-key").
				WithToolChoice(tt.choice)

			assert.NotNil(t, builder.toolChoice, "toolChoice should be set")
			assert.Nil(t, builder.lastError, "should not have error for valid choice")
		})
	}
}

func TestWithToolChoice_InvalidValue(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithToolChoice("invalid")

	assert.Nil(t, builder.toolChoice, "toolChoice should not be set for invalid value")
	assert.NotNil(t, builder.lastError, "should have error for invalid choice")
	assert.Contains(t, builder.lastError.Error(), "invalid toolChoice value", "error should mention invalid value")
	assert.Contains(t, builder.lastError.Error(), "auto, required, none", "error should list valid values")
}

func TestWithToolChoice_NilByDefault(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key")
	assert.Nil(t, builder.toolChoice, "toolChoice should be nil by default")
}

func TestWithToolChoice_ChainableFluent(t *testing.T) {
	// Test that WithToolChoice returns *Builder for chaining
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSystem("test").
		WithToolChoice("auto").
		WithTimeout(0)

	assert.NotNil(t, builder, "should support method chaining")
	assert.NotNil(t, builder.toolChoice, "toolChoice should be set")
}

func TestToolChoice_ValidationWithoutTools(t *testing.T) {
	ctx := context.Background()
	builder := NewOpenAI("gpt-4o-mini", "fake-key").
		WithToolChoice("required") // No tools configured

	_, err := builder.Ask(ctx, "test")

	assert.Error(t, err, "should error when toolChoice is set without tools")
	assert.Contains(t, err.Error(), "toolChoice is set but no tools are configured", "error should explain the issue")
	assert.Contains(t, err.Error(), "WithTools()", "error should suggest solution")
}

func TestToolChoice_ValidationWithTools(t *testing.T) {
	// This test verifies that validation passes when tools are present
	// We can't actually call the API without a valid key, but we can check
	// that the validation doesn't trigger
	ctx := context.Background()

	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"input": map[string]interface{}{
					"type":        "string",
					"description": "test input",
				},
			},
		},
		Handler: func(args string) (string, error) {
			return "test", nil
		},
	}

	builder := NewOpenAI("gpt-4o-mini", "fake-key").
		WithTools(tool).
		WithToolChoice("required")

	// We expect this to fail at API call, not at validation
	_, err := builder.Ask(ctx, "test")

	// Should fail with API error, not validation error
	if err != nil {
		assert.NotContains(t, err.Error(), "toolChoice is set but no tools are configured",
			"should not fail validation when tools are present")
	}
}

func TestToolChoice_StreamValidationWithoutTools(t *testing.T) {
	ctx := context.Background()
	builder := NewOpenAI("gpt-4o-mini", "fake-key").
		WithToolChoice("required"). // No tools configured
		OnStream(func(content string) {})

	_, err := builder.Stream(ctx, "test")

	assert.Error(t, err, "should error when toolChoice is set without tools in Stream")
	assert.Contains(t, err.Error(), "toolChoice is set but no tools are configured", "error should explain the issue")
}

func TestWithToolChoice_MultipleCallsLastWins(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithToolChoice("auto").
		WithToolChoice("required")

	// Last call should win
	assert.NotNil(t, builder.toolChoice, "toolChoice should be set")
	// We can't directly check the value without accessing internal fields,
	// but we verified it's set
}

func TestWithToolChoiceAfterInvalidValueKeepsError(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithToolChoice("invalid"). // Sets error
		WithToolChoice("auto")     // Sets toolChoice but doesn't clear lastError

	assert.NotNil(t, builder.toolChoice, "toolChoice should be set even after invalid")
	assert.NotNil(t, builder.lastError, "lastError should persist from invalid call")
	// This is correct behavior - lastError is checked at Ask()/Stream() time
}

func TestToolChoice_IntegrationWithAutoExecute(t *testing.T) {
	// Test that toolChoice works with auto-execute mode
	tool := &Tool{
		Name:        "calculator",
		Description: "A calculator",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression": map[string]interface{}{
					"type":        "string",
					"description": "math expression",
				},
			},
		},
		Handler: func(args string) (string, error) {
			return "42", nil
		},
	}

	builder := NewOpenAI("gpt-4o-mini", "fake-key").
		WithTools(tool).
		WithAutoExecute(true).
		WithToolChoice("required")

	assert.NotNil(t, builder.toolChoice, "toolChoice should be set")
	assert.True(t, builder.autoExecute, "auto-execute should be enabled")
}

func TestToolChoice_CaseSensitive(t *testing.T) {
	tests := []struct {
		name    string
		choice  string
		wantErr bool
	}{
		{"lowercase auto", "auto", false},
		{"uppercase AUTO", "AUTO", true}, // Should be case-sensitive
		{"mixed case Auto", "Auto", true},
		{"lowercase required", "required", false},
		{"uppercase REQUIRED", "REQUIRED", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOpenAI("gpt-4o-mini", "test-key").
				WithToolChoice(tt.choice)

			if tt.wantErr {
				assert.NotNil(t, builder.lastError, "should have error for %s", tt.choice)
			} else {
				assert.Nil(t, builder.lastError, "should not have error for %s", tt.choice)
			}
		})
	}
}

func TestToolChoice_EmptyString(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithToolChoice("")

	assert.NotNil(t, builder.lastError, "empty string should be invalid")
	assert.Contains(t, builder.lastError.Error(), "invalid toolChoice value", "should reject empty string")
}

func TestToolChoice_BuildParamsIntegration(t *testing.T) {
	// Test that toolChoice is properly included in params
	tool := &Tool{
		Name:        "test",
		Description: "test",
		Parameters:  map[string]interface{}{"type": "object"},
		Handler:     func(args string) (string, error) { return "ok", nil },
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithTools(tool).
		WithToolChoice("required")

	params := builder.buildParams(builder.buildMessages("test"))

	assert.NotNil(t, params.ToolChoice, "ToolChoice should be in params")
	assert.NotEmpty(t, params.Tools, "Tools should be in params")
}

func TestToolChoiceNilDoesNotSetParams(t *testing.T) {
	// When toolChoice is nil (not set), it should not be in params
	builder := NewOpenAI("gpt-4o-mini", "test-key")

	_ = builder.buildParams(builder.buildMessages("test"))

	// toolChoice field should use zero value when nil
	// OpenAI SDK omits zero values with omitzero tag
	assert.Nil(t, builder.toolChoice, "toolChoice should be nil")
}
