package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

// buildReActMetaTools creates the meta-tools for native ReAct function calling.
// These meta-tools provide structured output for the LLM's reasoning process:
//   - think(): Express reasoning steps before taking action
//   - use_tool(): Execute a registered tool with arguments (added in next task)
//   - final_answer(): Provide the final response with confidence (added in next task)
//
// This approach leverages OpenAI's function calling instead of text parsing,
// resulting in more reliable, language-agnostic ReAct execution.
//
// Usage (internal):
//
//	metaTools := builder.buildReActMetaTools()
//	// Returns slice of openai.ChatCompletionToolUnionParam
func (b *Builder) buildReActMetaTools() []openai.ChatCompletionToolUnionParam {
	var tools []openai.ChatCompletionToolUnionParam

	// Meta-tool 1: think() - Express reasoning
	// The LLM uses this to verbalize its thought process before acting
	thinkTool := NewTool("think", "Express your reasoning about the current task. Use this to think through the problem step-by-step before taking action or providing an answer.")
	thinkTool.AddParameter("reasoning", "string", "Your step-by-step thought process. Be explicit about what you know, what you need to find out, and what action you should take next.", true)

	tools = append(tools, thinkTool.toOpenAI())

	// Meta-tool 2: use_tool() - Execute a registered tool
	// The LLM uses this to call one of the available tools with specific arguments
	toolNames := b.getToolNames()
	if len(toolNames) > 0 {
		useTool := NewTool("use_tool", "Execute one of the available tools. Choose the appropriate tool based on what action you need to take.")

		// Add tool_name parameter as enum of available tools
		props := useTool.Parameters["properties"].(map[string]interface{})
		props["tool_name"] = map[string]interface{}{
			"type":        "string",
			"description": "The name of the tool to execute. Must be one of the registered tools.",
			"enum":        toolNames,
		}

		// Add tool_arguments parameter as object (tool-specific arguments)
		props["tool_arguments"] = map[string]interface{}{
			"type":        "object",
			"description": "The arguments to pass to the tool. The structure depends on the tool being called.",
		}

		// Mark both as required
		useTool.Parameters["required"] = []string{"tool_name", "tool_arguments"}

		tools = append(tools, useTool.toOpenAI())
	}

	// Meta-tool 3: final_answer() - Provide the final response
	// The LLM uses this when it has completed reasoning and is ready to answer
	finalAnswerTool := NewTool("final_answer", "Provide the final answer to the user's question. Use this when you have completed all necessary reasoning and tool usage, and are ready to give the final response.")
	finalAnswerTool.AddParameter("answer", "string", "The complete answer to provide to the user. This should be a clear, comprehensive response based on your reasoning and any tool usage.", true)
	finalAnswerTool.AddParameter("confidence", "number", "Your confidence level in this answer, from 0.0 (no confidence) to 1.0 (completely certain). Optional, defaults to 1.0.", false)

	tools = append(tools, finalAnswerTool.toOpenAI())

	return tools
}

// executeReActNative implements ReAct pattern using native function calling.
// This is the modern, recommended approach that leverages OpenAI's structured
// outputs instead of text parsing.
//
// Flow:
//  1. Build meta-tools (think, use_tool, final_answer)
//  2. Send user task with meta-tools to LLM
//  3. Loop: Handle tool calls until final_answer() is reached
//  4. Return result with complete conversation history
//
// Advantages over text-based ReAct:
//   - No regex parsing (more reliable)
//   - Language-agnostic (works with any language)
//   - Structured data (easier to process)
//   - Better error handling
//
// Internal use only.
func (b *Builder) executeReActNative(ctx context.Context, task string) (*ReActResult, error) {
	// Initialize result
	result := &ReActResult{
		Steps: []ReActStep{},
	}

	// Initialize metrics if enabled
	if b.reactConfig.EnableMetrics {
		result.Metrics = NewReActMetrics()
	}

	// Initialize timeline if enabled
	if b.reactConfig.EnableTimeline {
		result.Timeline = NewReActTimeline()
		result.Timeline.AddEvent("start", "Native ReAct execution started", 0, nil)
	}

	// Apply timeout if configured
	if b.reactConfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.reactConfig.Timeout)
		defer cancel()
	}

	// Build meta-tools for native ReAct
	metaTools := b.buildReActMetaTools()
	_ = metaTools // TODO: Use metaTools in LLM call (Task 6)

	// Build conversation history with system prompt
	messages := []Message{}

	// Add Native ReAct system prompt
	systemPrompt := b.reactConfig.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = b.buildReActNativeSystemPrompt()
	}
	messages = append(messages, System(systemPrompt))

	// Add user task
	messages = append(messages, User(task))

	// Execution loop
	for iteration := 0; iteration < b.reactConfig.MaxIterations; iteration++ {
		if b.reactConfig.EnableTimeline {
			result.Timeline.AddEvent("iteration_start", fmt.Sprintf("Iteration %d started", iteration+1), 0, nil)
		}

		// Call LLM with meta-tools
		// TODO: Implement LLM call with tools (Task 6)
		// For now, return error
		result.Success = false
		result.Error = fmt.Errorf("executeReActNative not fully implemented yet")
		return result, result.Error
	}

	// If we reach here, max iterations exceeded without final_answer
	result.Success = false
	result.Error = fmt.Errorf("max iterations (%d) reached without final answer", b.reactConfig.MaxIterations)

	if b.reactConfig.EnableTimeline {
		result.Timeline.AddEvent("max_iterations", "Max iterations reached", 0, nil)
	}

	return result, result.Error
}

// buildReActNativeSystemPrompt creates the system prompt for native ReAct mode.
// This prompt explains how to use the meta-tools effectively.
func (b *Builder) buildReActNativeSystemPrompt() string {
	// TODO: Implement comprehensive system prompt (Phase 5)
	return "You are a helpful assistant. Use the provided tools to think, execute actions, and provide answers."
}
