package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

		if b.reactConfig.EnableMetrics {
			result.Metrics.TotalIterations = iteration + 1
		}

		// Progressive urgency reminders (v0.7.6+)
		// Inject reminder messages at critical points to guide LLM toward final_answer()
		if b.reactConfig.EnableIterationReminders {
			remainingIterations := b.reactConfig.MaxIterations - (iteration + 1)

			// At n-2 iterations: Gentle reminder
			if remainingIterations == 2 {
				reminder := System("⚠️ REMINDER: You have 2 iterations remaining before max iterations is reached. Please start wrapping up your reasoning and prepare to call final_answer().")
				messages = append(messages, reminder)

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("reminder", "2 iterations remaining (gentle reminder)", 0, nil)
				}
			}

			// At n-1 iterations: Urgent reminder
			if remainingIterations == 1 {
				reminder := System("⚠️ URGENT: This is your LAST iteration before max iterations is reached. You MUST call final_answer() now with your best response based on the work completed so far.")
				messages = append(messages, reminder)

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("reminder", "1 iteration remaining (urgent reminder)", 0, nil)
				}
			}

			// At n iterations (last one): Critical reminder
			if remainingIterations == 0 {
				reminder := System("🚨 CRITICAL: This is the FINAL iteration. You absolutely MUST call final_answer() in this iteration. If you don't, your work will be lost. Provide your best answer based on your reasoning so far.")
				messages = append(messages, reminder)

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("reminder", "0 iterations remaining (critical reminder)", 0, nil)
				}
			}
		}

		// Call LLM with meta-tools
		completion, err := b.callLLMWithMetaTools(ctx, messages, metaTools)
		if err != nil {
			result.Success = false
			result.Error = fmt.Errorf("LLM call failed at iteration %d: %w", iteration+1, err)

			if b.reactConfig.EnableTimeline {
				result.Timeline.AddEvent("error", fmt.Sprintf("LLM error: %v", err), 0, nil)
			}

			// Callback: onError
			if b.reactConfig.Callback != nil {
				b.reactConfig.Callback.OnError(result.Error)
			}

			return result, result.Error
		}

		// Check if LLM made any tool calls
		if len(completion.Choices) == 0 {
			result.Success = false
			result.Error = fmt.Errorf("no response choices at iteration %d", iteration+1)
			return result, result.Error
		}

		choice := completion.Choices[0]

		// If no tool calls, this is an error (LLM should always use meta-tools)
		if len(choice.Message.ToolCalls) == 0 {
			// LLM didn't use any tools - treat as implicit final answer
			content := choice.Message.Content
			result.Answer = content
			result.Success = true
			result.Iterations = iteration + 1

			// Add implicit final step
			step := ReActStep{
				Type:      StepTypeFinal,
				Content:   content,
				Timestamp: time.Now(),
			}
			result.Steps = append(result.Steps, step)

			if b.reactConfig.EnableTimeline {
				result.Timeline.AddEvent("final_implicit", "Implicit final answer", 0, nil)
			}

			return result, nil
		}

		// Process tool calls
		for _, toolCall := range choice.Message.ToolCalls {
			funcName := toolCall.Function.Name
			funcArgs := toolCall.Function.Arguments

			// Handle each meta-tool
			switch funcName {
			case "think":
				// Parse reasoning from arguments
				var args struct {
					Reasoning string `json:"reasoning"`
				}
				if err := json.Unmarshal([]byte(funcArgs), &args); err != nil {
					result.Success = false
					result.Error = fmt.Errorf("failed to parse think() arguments: %w", err)
					return result, result.Error
				}

				// Record THOUGHT step
				step := ReActStep{
					Type:      StepTypeThought,
					Content:   args.Reasoning,
					Timestamp: time.Now(),
				}
				result.Steps = append(result.Steps, step)

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("thought", args.Reasoning, 0, nil)
				}

				// Callback: onStep
				if b.reactConfig.Callback != nil {
					b.reactConfig.Callback.OnStep(step)
				}

				// Add assistant's tool call to conversation
				messages = append(messages, Assistant(fmt.Sprintf("THOUGHT: %s", args.Reasoning)))

			case "use_tool":
				// Parse tool name and arguments
				var args struct {
					ToolName      string                 `json:"tool_name"`
					ToolArguments map[string]interface{} `json:"tool_arguments"`
				}
				if err := json.Unmarshal([]byte(funcArgs), &args); err != nil {
					result.Success = false
					result.Error = fmt.Errorf("failed to parse use_tool() arguments: %w", err)
					return result, result.Error
				}

				// Record ACTION step
				actionStep := ReActStep{
					Type:      StepTypeAction,
					Content:   fmt.Sprintf("%s(%v)", args.ToolName, args.ToolArguments),
					Tool:      args.ToolName,
					Args:      args.ToolArguments,
					Timestamp: time.Now(),
				}
				result.Steps = append(result.Steps, actionStep)

				if b.reactConfig.EnableMetrics {
					result.Metrics.ToolCalls++
				}

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("action", fmt.Sprintf("Tool: %s", args.ToolName), 0, nil)
				}

				// Callback: onStep
				if b.reactConfig.Callback != nil {
					b.reactConfig.Callback.OnStep(actionStep)
				}

				// Execute the actual tool
				toolResult, toolErr := b.executeTool(ctx, args.ToolName, args.ToolArguments)

				// Record OBSERVATION step
				obsContent := toolResult
				if toolErr != nil {
					obsContent = fmt.Sprintf("ERROR: %v", toolErr)
					if b.reactConfig.EnableMetrics {
						result.Metrics.Errors++
					}
				}

				obsStep := ReActStep{
					Type:      StepTypeObservation,
					Content:   obsContent,
					Tool:      args.ToolName,
					Timestamp: time.Now(),
					Error:     toolErr,
				}
				result.Steps = append(result.Steps, obsStep)

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("observation", obsContent, 0, nil)
				}

				// Callback: onStep
				if b.reactConfig.Callback != nil {
					b.reactConfig.Callback.OnStep(obsStep)
				}

				// Add tool execution to conversation
				messages = append(messages, Assistant(fmt.Sprintf("ACTION: %s(%v)", args.ToolName, args.ToolArguments)))
				messages = append(messages, User(fmt.Sprintf("OBSERVATION: %s", obsContent)))

			case "final_answer":
				// Parse final answer
				var args struct {
					Answer     string   `json:"answer"`
					Confidence *float64 `json:"confidence,omitempty"`
				}
				if err := json.Unmarshal([]byte(funcArgs), &args); err != nil {
					result.Success = false
					result.Error = fmt.Errorf("failed to parse final_answer() arguments: %w", err)
					return result, result.Error
				}

				// Record FINAL step
				step := ReActStep{
					Type:      StepTypeFinal,
					Content:   args.Answer,
					Timestamp: time.Now(),
				}
				result.Steps = append(result.Steps, step)

				if b.reactConfig.EnableTimeline {
					confidenceStr := ""
					if args.Confidence != nil {
						confidenceStr = fmt.Sprintf(" (confidence: %.2f)", *args.Confidence)
					}
					result.Timeline.AddEvent("final", fmt.Sprintf("Final answer%s", confidenceStr), 0, nil)
				}

				// Callback: onStep
				if b.reactConfig.Callback != nil {
					b.reactConfig.Callback.OnStep(step)
				}

				// Set result
				result.Answer = args.Answer
				result.Success = true
				result.Iterations = iteration + 1

				return result, nil

			default:
				// Unknown meta-tool
				result.Success = false
				result.Error = fmt.Errorf("unknown meta-tool called: %s", funcName)
				return result, result.Error
			}
		}
	}

	// If we reach here, max iterations exceeded without final_answer
	// Check if auto-fallback is enabled (v0.7.6+)
	if b.reactConfig.EnableAutoFallback || b.reactConfig.ForceFinalAnswerAtMax {
		// Synthesize a final answer from the collected reasoning steps
		fallbackAnswer := b.synthesizeFallbackAnswer(result.Steps, b.reactConfig.MaxIterations)

		result.Answer = fallbackAnswer
		result.Success = true
		result.Iterations = b.reactConfig.MaxIterations

		// Add a fallback step to document what happened
		fallbackStep := ReActStep{
			Type:      StepTypeFinal,
			Content:   fallbackAnswer,
			Timestamp: time.Now(),
		}
		result.Steps = append(result.Steps, fallbackStep)

		if b.reactConfig.EnableTimeline {
			result.Timeline.AddEvent("auto_fallback", "Auto-fallback: synthesized answer from steps", 0, nil)
		}

		// Callback: onComplete (even though we hit max iterations)
		if b.reactConfig.Callback != nil {
			b.reactConfig.Callback.OnComplete(result)
		}

		return result, nil
	}

	// Auto-fallback disabled - return rich error with debugging info (v0.7.6+)
	result.Success = false
	result.Error = NewReActMaxIterationsError(b.reactConfig.MaxIterations, b.reactConfig.MaxIterations, result.Steps)

	if b.reactConfig.EnableTimeline {
		result.Timeline.AddEvent("max_iterations", "Max iterations reached", 0, nil)
	}

	// Callback: onError
	if b.reactConfig.Callback != nil {
		b.reactConfig.Callback.OnError(result.Error)
	}

	return result, result.Error
}

// synthesizeFallbackAnswer creates a best-effort answer from the reasoning steps
// when max iterations is reached without an explicit final_answer() call.
//
// This is part of the auto-fallback mechanism (v0.7.6+) that provides graceful
// degradation instead of hard errors.
func (b *Builder) synthesizeFallbackAnswer(steps []ReActStep, maxIterations int) string {
	if len(steps) == 0 {
		return fmt.Sprintf("⚠️ Max iterations (%d) reached with no reasoning steps. Unable to provide an answer.", maxIterations)
	}

	// Build summary of what was accomplished
	thoughtCount := 0
	actionCount := 0
	lastThought := ""
	lastObservation := ""

	for _, step := range steps {
		switch step.Type {
		case StepTypeThought:
			thoughtCount++
			lastThought = step.Content
		case StepTypeAction:
			actionCount++
		case StepTypeObservation:
			lastObservation = step.Content
		}
	}

	// Synthesize answer based on available information
	var answer string
	answer += fmt.Sprintf("⚠️ Auto-fallback activated: Max iterations (%d) reached without explicit final answer.\n\n", maxIterations)
	answer += fmt.Sprintf("**Work Summary**: %d thoughts, %d actions taken.\n\n", thoughtCount, actionCount)

	if lastThought != "" {
		answer += fmt.Sprintf("**Last Reasoning**: %s\n\n", lastThought)
	}

	if lastObservation != "" {
		answer += fmt.Sprintf("**Last Observation**: %s\n\n", lastObservation)
	}

	answer += "**Note**: This answer was synthesized from partial work. For better results, consider:\n"
	answer += "1. Increasing MaxIterations for this task complexity\n"
	answer += "2. Simplifying the task\n"
	answer += "3. Using WithReActComplexity() to auto-configure settings"

	return answer
}

// callLLMWithMetaTools calls the LLM with meta-tools and returns the completion.
// This is a helper for executeReActNative() to keep the main loop cleaner.
func (b *Builder) callLLMWithMetaTools(ctx context.Context, messages []Message, metaTools []openai.ChatCompletionToolUnionParam) (*openai.ChatCompletion, error) {
	// Ensure client is initialized
	if err := b.ensureClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	// Convert messages to OpenAI format
	openaiMessages := convertMessages(messages)

	// Build params with meta-tools
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(b.model),
		Messages: openaiMessages,
		Tools:    metaTools,
	}

	// Apply temperature if set
	if b.temperature != nil {
		params.Temperature = openai.Float(*b.temperature)
	}

	// Execute request
	completion, err := b.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	return completion, nil
}

// buildReActNativeSystemPrompt creates the system prompt for native ReAct mode.
// This prompt explains how to use the meta-tools effectively.
func (b *Builder) buildReActNativeSystemPrompt() string {
	toolCount := len(b.tools)
	toolsList := ""
	if toolCount > 0 {
		names := b.getToolNames()
		toolsList = fmt.Sprintf("Available tools: %v", names)
	} else {
		toolsList = "No tools available."
	}

	return fmt.Sprintf(`You are an intelligent assistant that uses structured function calling to solve problems step-by-step.

AVAILABLE FUNCTIONS:
- think(reasoning): Express your step-by-step reasoning before taking action
- use_tool(tool_name, tool_arguments): Execute one of the registered tools (only if tools available)
- final_answer(answer, confidence): Provide your final response with optional confidence (0.0-1.0)

%s

WORKFLOW:
1. Start by calling think() to reason about the problem
2. If you need information or computation, use use_tool() to execute appropriate tools
3. Continue thinking and using tools as needed
4. End with final_answer() when you have a complete response

IMPORTANT RULES:
- Always use think() before taking any action
- Only use tools that are actually registered and available
- For tool arguments, follow each tool's expected parameter structure
- Use final_answer() to conclude - this ends the conversation
- Be thorough in your reasoning but concise in your answers

Example flow:
1. think("I need to understand what the user is asking...")
2. use_tool("search", {"query": "..."}) [if search tool available]
3. think("Based on the results, I can now provide...")
4. final_answer("Here is the answer...", 0.9)`, toolsList)
}
