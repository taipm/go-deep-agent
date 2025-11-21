package agent

import (
	"context"
	"strings"
	"testing"
)

// TestEnhancedValidation tests the new comprehensive validation methods
func TestEnhancedValidation(t *testing.T) {

	t.Run("ValidateConfig - valid OpenAI configuration", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678")
		err := builder.ValidateConfig()
		if err != nil {
			t.Errorf("Valid configuration should pass validation: %v", err)
		}
	})

	t.Run("ValidateConfig - missing OpenAI API key", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "")
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("Missing API key should fail validation")
		}
		if !strings.Contains(err.Error(), "OpenAI API key is required") {
			t.Errorf("Expected API key error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "export OPENAI_API_KEY") {
			t.Errorf("Should include fix suggestions: %v", err)
		}
	})

	t.Run("ValidateConfig - invalid API key format", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "short")
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("Invalid API key should fail validation")
		}
		if !strings.Contains(err.Error(), "appears to be invalid") {
			t.Errorf("Expected invalid API key error, got: %v", err)
		}
	})

	t.Run("ValidateConfig - valid Gemini configuration", func(t *testing.T) {
		builder := NewGemini("gemini-pro", "AIza1234567890abcdef")
		err := builder.ValidateConfig()
		if err != nil {
			t.Errorf("Valid Gemini configuration should pass validation: %v", err)
		}
	})

	t.Run("ValidateConfig - missing Gemini API key", func(t *testing.T) {
		builder := NewGemini("gemini-pro", "")
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("Missing Gemini API key should fail validation")
		}
		if !strings.Contains(err.Error(), "Gemini API key is required") {
			t.Errorf("Expected Gemini API key error, got: %v", err)
		}
	})

	t.Run("ValidateConfig - valid Ollama configuration", func(t *testing.T) {
		builder := NewOllama("llama2")
		err := builder.ValidateConfig()
		if err != nil {
			t.Errorf("Valid Ollama configuration should pass validation: %v", err)
		}
	})

	t.Run("ValidateConfig - temperature out of range", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithTemperature(3.0) // Too high
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("Temperature out of range should fail validation")
		}
		if !strings.Contains(err.Error(), "temperature 3.0 is out of valid range") {
			t.Errorf("Expected temperature range error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "Creative writing") {
			t.Errorf("Should include usage suggestions: %v", err)
		}
	})

	t.Run("ValidateConfig - topP out of range", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithTopP(1.5) // Too high
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("TopP out of range should fail validation")
		}
		if !strings.Contains(err.Error(), "top_p 1.50 is out of valid range") {
			t.Errorf("Expected top_p range error, got: %v", err)
		}
	})

	t.Run("ValidateConfig - maxTokens too small", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithMaxTokens(0) // Too small
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("MaxTokens too small should fail validation")
		}
		if !strings.Contains(err.Error(), "max_tokens 0 is too small") {
			t.Errorf("Expected maxTokens error, got: %v", err)
		}
	})

	t.Run("ValidateConfig - maxTokens exceeds OpenAI limit", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithMaxTokens(50000) // Exceeds OpenAI limit
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("MaxTokens exceeding limit should fail validation")
		}
		if !strings.Contains(err.Error(), "exceeds OpenAI limit") {
			t.Errorf("Expected OpenAI limit error, got: %v", err)
		}
	})

	t.Run("ValidateConfig - presencePenalty out of range", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithPresencePenalty(3.0) // Too high
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("PresencePenalty out of range should fail validation")
		}
		if !strings.Contains(err.Error(), "presence_penalty 3.0 is out of valid range") {
			t.Errorf("Expected presencePenalty error, got: %v", err)
		}
	})

	t.Run("ValidateConfig - frequencyPenalty out of range", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithFrequencyPenalty(-3.0) // Too low
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("FrequencyPenalty out of range should fail validation")
		}
		if !strings.Contains(err.Error(), "frequency_penalty -3.0 is out of valid range") {
			t.Errorf("Expected frequencyPenalty error, got: %v", err)
		}
	})

	t.Run("ValidateConfig - topLogprobs out of range", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithLogprobs(true).
			WithTopLogprobs(25) // Too high
		err := builder.ValidateConfig()
		if err == nil {
			t.Error("TopLogprobs out of range should fail validation")
		}
		if !strings.Contains(err.Error(), "top_logprobs 25 is out of valid range") {
			t.Errorf("Expected topLogprobs error, got: %v", err)
		}
	})
}

// TestValidateWithDetails tests the detailed validation functionality
func TestValidateWithDetails(t *testing.T) {
	t.Run("ValidateWithDetails - valid configuration", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithTemperature(0.7).
			WithMaxTokens(1000).
			WithTimeout(30 * 1000) // 30 seconds in milliseconds (assuming time.Millisecond)

		details, err := builder.ValidateWithDetails()
		if err != nil {
			t.Errorf("Valid configuration should not have errors: %v", err)
		}

		if details.Provider != "openai" {
			t.Errorf("Expected provider 'openai', got '%s'", details.Provider)
		}
		if details.Model != "gpt-4o-mini" {
			t.Errorf("Expected model 'gpt-4o-mini', got '%s'", details.Model)
		}
		if !details.APIKeySet {
			t.Error("APIKeySet should be true")
		}
		if !details.IsValid() {
			t.Error("Configuration should be valid")
		}
		if details.Summary() != "Validation PASSED - no issues found" {
			t.Errorf("Unexpected summary: %s", details.Summary())
		}
	})

	t.Run("ValidateWithDetails - with warnings", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678")
		// Don't set temperature, maxTokens, or timeout

		details, err := builder.ValidateWithDetails()
		if err != nil {
			t.Errorf("Configuration should be valid despite warnings: %v", err)
		}

		if !details.HasWarnings() {
			t.Error("Should have warnings for unset parameters")
		}

		expectedWarnings := []string{
			"Temperature not set",
			"MaxTokens not set",
			"No timeout set",
		}

		for _, expected := range expectedWarnings {
			found := false
			for _, warning := range details.Warnings {
				if strings.Contains(warning, expected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected warning about '%s', got warnings: %v", expected, details.Warnings)
			}
		}

		if !strings.Contains(details.Summary(), "with 3 warnings") {
			t.Errorf("Summary should mention warnings: %s", details.Summary())
		}
	})

	t.Run("ValidateWithDetails - with errors", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", ""). // Missing API key
			WithTemperature(3.0)                 // Invalid temperature

		details, err := builder.ValidateWithDetails()
		if err == nil {
			t.Error("Invalid configuration should have errors")
		}

		if details.IsValid() {
			t.Error("Configuration should not be valid")
		}

		if len(details.Errors) == 0 {
			t.Error("Should have validation errors")
		}

		if !strings.Contains(details.Summary(), "FAILED") {
			t.Errorf("Summary should indicate failure: %s", details.Summary())
		}
	})

	t.Run("ValidateWithDetails - adapter configuration", func(t *testing.T) {
		// Create a mock adapter for testing
		adapter := &mockAdapter{}
		builder := NewWithAdapter("test-model", adapter)

		details, err := builder.ValidateWithDetails()
		if err != nil {
			t.Errorf("Adapter configuration should be valid: %v", err)
		}

		if !details.AdapterSet {
			t.Error("AdapterSet should be true")
		}
		if !details.IsValid() {
			t.Error("Configuration should be valid with adapter")
		}
	})
}

// TestValidationIntegration tests validation integration with execution methods
func TestValidationIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("Ask method calls validation", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678"). // Valid API key
			WithTemperature(3.0)                 // Invalid temperature

		_, err := builder.Ask(ctx, "Hello")
		if err == nil {
			t.Error("Ask should fail due to validation errors")
		}

		// Should contain validation error messages
		errMsg := err.Error()
		if !strings.Contains(errMsg, "temperature 3.0 is out of valid range") {
			t.Errorf("Expected temperature validation error: %s", errMsg)
		}
		// Should NOT contain API key error since we provided a valid key
		if strings.Contains(errMsg, "OpenAI API key is required") {
			t.Errorf("Should not contain API key error when key is provided: %s", errMsg)
		}
	})

	t.Run("Stream method calls validation", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678"). // Valid API key
			WithTopP(1.5)                         // Invalid topP

		_, err := builder.Stream(ctx, "Hello")
		if err == nil {
			t.Error("Stream should fail due to validation errors")
		}

		// Should contain validation error messages
		errMsg := err.Error()
		if !strings.Contains(errMsg, "top_p 1.50 is out of valid range") {
			t.Errorf("Expected topP validation error: %s", errMsg)
		}
		// Should NOT contain API key error since we provided a valid key
		if strings.Contains(errMsg, "OpenAI API key is required") {
			t.Errorf("Should not contain API key error when key is provided: %s", errMsg)
		}
	})

	t.Run("Valid configuration passes validation in Ask", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
			WithTemperature(0.7).
			WithMaxTokens(100)

		// This should pass validation but fail at API call (since we're using a fake key)
		_, err := builder.Ask(ctx, "Hello")

		// The error should NOT be a validation error
		if err != nil && strings.Contains(err.Error(), "API key is required") {
			t.Errorf("Should not get validation error for valid config: %v", err)
		}
	})
}

// TestFromEnvValidation tests validation of FromEnv constructor
func TestFromEnvValidation(t *testing.T) {
	// This test can't modify environment variables safely in parallel tests,
	// but we can test the validation logic

	t.Run("FromEnv error includes validation guidance", func(t *testing.T) {
		// Create a builder that would come from FromEnv with no config
		// (We can't actually test FromEnv since it requires env var manipulation)

		// Instead, test that the error message format is helpful
		builder := NewOpenAI("gpt-4o-mini", "")
		err := builder.ValidateConfig()

		if err == nil {
			t.Error("Should fail validation")
		}

		errMsg := err.Error()
		expectedPhrases := []string{
			"OpenAI API key is required",
			"Fix options:",
			"export OPENAI_API_KEY",
			"Get your key:",
			"https://platform.openai.com/api-keys",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(errMsg, phrase) {
				t.Errorf("Error message missing '%s': %s", phrase, errMsg)
			}
		}
	})
}

// Mock adapter for testing
type mockAdapter struct{}

func (m *mockAdapter) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{
		Content: "mock response",
		Usage: TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *mockAdapter) Stream(ctx context.Context, req *CompletionRequest, onChunk func(string)) (*CompletionResponse, error) {
	onChunk("mock")
	return &CompletionResponse{
		Content: "mock response",
		Usage: TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *mockAdapter) Close() error {
	return nil
}

// TestValidationEdgeCases tests edge cases and boundary conditions
func TestValidationEdgeCases(t *testing.T) {
	t.Run("Boundary values - temperature", func(t *testing.T) {
		testCases := []struct {
			name        string
			temperature float64
			shouldPass  bool
		}{
			{"exactly 0.0", 0.0, true},
			{"exactly 2.0", 2.0, true},
			{"just below 0.0", -0.1, false},
			{"just above 2.0", 2.1, false},
			{"middle value 1.0", 1.0, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
					WithTemperature(tc.temperature)
				err := builder.ValidateConfig()

				if tc.shouldPass && err != nil {
					t.Errorf("Expected validation to pass for temperature %.1f: %v", tc.temperature, err)
				}
				if !tc.shouldPass && err == nil {
					t.Errorf("Expected validation to fail for temperature %.1f", tc.temperature)
				}
			})
		}
	})

	t.Run("Boundary values - topP", func(t *testing.T) {
		testCases := []struct {
			name       string
			topP       float64
			shouldPass bool
		}{
			{"exactly 0.0", 0.0, true},
			{"exactly 1.0", 1.0, true},
			{"just below 0.0", -0.01, false},
			{"just above 1.0", 1.01, false},
			{"middle value 0.5", 0.5, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := NewOpenAI("gpt-4o-mini", "sk-1234567890abcdef1234567890abcdef12345678").
					WithTopP(tc.topP)
				err := builder.ValidateConfig()

				if tc.shouldPass && err != nil {
					t.Errorf("Expected validation to pass for topP %.2f: %v", tc.topP, err)
				}
				if !tc.shouldPass && err == nil {
					t.Errorf("Expected validation to fail for topP %.2f", tc.topP)
				}
			})
		}
	})

	t.Run("Multiple validation errors", func(t *testing.T) {
		builder := NewOpenAI("gpt-4o-mini", "").           // Missing API key
			WithTemperature(3.0).                            // Invalid temperature
			WithTopP(1.5).                                   // Invalid topP
			WithMaxTokens(0).                               // Invalid maxTokens
			WithPresencePenalty(3.0).                       // Invalid presencePenalty
			WithFrequencyPenalty(-3.0).                     // Invalid frequencyPenalty
			WithLogprobs(true).                             // Requires topLogprobs setting
			WithTopLogprobs(25)                             // Invalid topLogprobs

		details, err := builder.ValidateWithDetails()
		if err == nil {
			t.Error("Should fail validation with multiple errors")
		}

		if len(details.Errors) == 0 {
			t.Error("Should have collected multiple validation errors")
		}

		// Check that we got multiple errors
		if len(details.Errors) < 5 {
			t.Errorf("Expected multiple validation errors, got %d: %v", len(details.Errors), details.Errors)
		}

		// Look for key error types in the collected errors
		allErrors := strings.Join(details.Errors, " ")
		expectedErrors := []string{
			"OpenAI API key is required",
			"temperature 3.0 is out of valid range",
			"top_p 1.50 is out of valid range",
			"max_tokens 0 is too small",
			"presence_penalty 3.0 is out of valid range",
			"frequency_penalty -3.0 is out of valid range",
			"top_logprobs 25 is out of valid range",
		}

		missingCount := 0
		for _, expected := range expectedErrors {
			if !strings.Contains(allErrors, expected) {
				missingCount++
				t.Errorf("Multiple validation error missing '%s': %s", expected, allErrors)
			}
		}

		if missingCount > 0 {
			t.Errorf("Missing %d expected errors out of %d total", missingCount, len(expectedErrors))
		}
	})
}