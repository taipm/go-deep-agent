package agent

import (
	"context"
	"fmt"
)

// AgentLLMWrapper wraps an Agent to implement the llmGenerator interface.
// This allows the Decomposer to use the Agent's LLM capabilities for task decomposition.
type AgentLLMWrapper struct {
	agent *Agent
}

// NewAgentLLMWrapper creates a new wrapper around an Agent.
func NewAgentLLMWrapper(agent *Agent) *AgentLLMWrapper {
	return &AgentLLMWrapper{agent: agent}
}

// Generate implements the llmGenerator interface by calling the Agent's Chat method.
// It converts the prompt into a chat message and returns the LLM's response.
func (w *AgentLLMWrapper) Generate(ctx context.Context, prompt string, opts *ChatOptions) (string, error) {
	// Use the provided options or create default ones
	if opts == nil {
		opts = &ChatOptions{}
	}

	// Call the agent's Chat method
	result, err := w.agent.Chat(ctx, prompt, opts)
	if err != nil {
		return "", fmt.Errorf("agent chat failed: %w", err)
	}

	return result.Content, nil
}

// PlanAndExecute decomposes a goal into tasks and executes them using the planning system.
// It returns detailed results about the plan execution including metrics and task outcomes.
func (a *Agent) PlanAndExecute(ctx context.Context, goal string) (*PlanResult, error) {
	// Get planner configuration (use default if not configured)
	plannerConfig := DefaultPlannerConfig()

	// Check if agent has custom planner config
	// This will be added to Agent struct in the next step
	// For now, use default config

	// Create LLM wrapper for decomposer
	llmWrapper := NewAgentLLMWrapper(a)

	// Create decomposer
	decomposer := NewDecomposer(plannerConfig, llmWrapper)

	// Decompose goal into plan
	plan, err := decomposer.Decompose(ctx, goal)
	if err != nil {
		return nil, fmt.Errorf("goal decomposition failed: %w", err)
	}

	// Create executor
	executor := NewExecutor(plannerConfig, a)

	// Execute the plan
	result, err := executor.Execute(ctx, plan)
	if err != nil {
		return result, fmt.Errorf("plan execution failed: %w", err)
	}

	return result, nil
}

// PlanAndExecuteWithConfig is like PlanAndExecute but allows custom planner configuration.
func (a *Agent) PlanAndExecuteWithConfig(ctx context.Context, goal string, config *PlannerConfig) (*PlanResult, error) {
	if config == nil {
		config = DefaultPlannerConfig()
	}

	// Create LLM wrapper for decomposer
	llmWrapper := NewAgentLLMWrapper(a)

	// Create decomposer with custom config
	decomposer := NewDecomposer(config, llmWrapper)

	// Decompose goal into plan
	plan, err := decomposer.Decompose(ctx, goal)
	if err != nil {
		return nil, fmt.Errorf("goal decomposition failed: %w", err)
	}

	// Create executor with custom config
	executor := NewExecutor(config, a)

	// Execute the plan
	result, err := executor.Execute(ctx, plan)
	if err != nil {
		return result, fmt.Errorf("plan execution failed: %w", err)
	}

	return result, nil
}
