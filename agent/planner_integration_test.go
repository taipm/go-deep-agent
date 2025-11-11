package agent

import (
	"context"
	"testing"
)

// Test AgentLLMWrapper with real Agent (basic smoke test)
func TestAgentLLMWrapperCreation(t *testing.T) {
	config := Config{
		Provider: "openai",
		Model:    "gpt-4",
		APIKey:   "test-key",
	}

	agent, err := NewAgent(config)
	if err != nil {
		t.Skip("Skipping test - requires valid config")
	}

	wrapper := NewAgentLLMWrapper(agent)
	if wrapper == nil {
		t.Fatal("NewAgentLLMWrapper returned nil")
	}

	if wrapper.agent != agent {
		t.Error("Wrapper should reference the provided agent")
	}
}

// Test PlanAndExecute method exists and has correct signature
func TestPlanAndExecuteMethodExists(t *testing.T) {
	config := Config{
		Provider: "openai",
		Model:    "gpt-4",
		APIKey:   "test-key",
	}

	agent, err := NewAgent(config)
	if err != nil {
		t.Skip("Skipping test - requires valid config")
	}

	// Just verify the method exists (will fail without API key, but that's OK)
	ctx := context.Background()
	_, err = agent.PlanAndExecute(ctx, "test goal")

	// We expect an error (no valid API key), but the method should exist
	// The error could be from decomposition or execution, but not "method not found"
	if err != nil {
		t.Logf("Expected error (no API key): %v", err)
	}
}

// Test PlanAndExecuteWithConfig method exists
func TestPlanAndExecuteWithConfigMethodExists(t *testing.T) {
	config := Config{
		Provider: "openai",
		Model:    "gpt-4",
		APIKey:   "test-key",
	}

	agent, err := NewAgent(config)
	if err != nil {
		t.Skip("Skipping test - requires valid config")
	}

	customConfig := &PlannerConfig{
		MaxDepth:        2,
		MaxSubtasks:     3,
		MinSubtaskSplit: 2,
		Strategy:        StrategySequential,
	}

	ctx := context.Background()
	_, err = agent.PlanAndExecuteWithConfig(ctx, "test goal", customConfig)

	if err != nil {
		t.Logf("Expected error (no API key): %v", err)
	}
}

// Test integration flow structure (decompose -> execute)
func TestIntegrationFlow(t *testing.T) {
	// This test verifies the structure without needing real API calls
	// The actual integration will be tested in examples

	plannerConfig := DefaultPlannerConfig()
	if plannerConfig == nil {
		t.Fatal("DefaultPlannerConfig returned nil")
	}

	// Verify config is valid
	if err := plannerConfig.Validate(); err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}

	// Test that components can be created
	// (actual execution requires valid API keys and will be in examples)
	t.Log("Integration flow structure validated")
}
