package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// WithReActMode enables or disables ReAct pattern execution.
// When enabled, the Execute() method will use the ReAct reasoning loop.
// When disabled, Execute() behaves like Ask().
//
// Example:
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).  // Enable ReAct
//	    Build()
//
//	result, err := ai.Execute(ctx, "Search for weather in Paris")
func (b *Builder) WithReActMode(enabled bool) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.Enabled = enabled
	return b
}

// WithReActMaxIterations sets the maximum number of reasoning loops.
// Each iteration consists of: THOUGHT → ACTION → OBSERVATION.
// When exceeded, execution stops and returns ErrMaxIterationsReached.
//
// Default: 5
// Valid range: 1-100
//
// Example:
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActMaxIterations(10).  // Allow up to 10 reasoning steps
//	    Build()
func (b *Builder) WithReActMaxIterations(n int) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.MaxIterations = n
	return b
}

// WithReActTimeout sets the execution timeout for the entire ReAct session.
// If exceeded, execution stops and returns partial results with a timeout error.
//
// Default: 60 seconds
// Valid range: 1s - 10 minutes
//
// Example:
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActTimeout(2 * time.Minute).  // 2 minute timeout
//	    Build()
func (b *Builder) WithReActTimeout(d time.Duration) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.Timeout = d
	return b
}

// WithReActStrict enables or disables strict mode.
// In strict mode, parse errors cause execution to fail immediately.
// In non-strict mode (default), parse errors trigger graceful fallback to normal execution.
//
// Default: false (graceful)
//
// Example:
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActStrict(true).  // Fail fast on parse errors
//	    Build()
func (b *Builder) WithReActStrict(strict bool) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.Strict = strict
	return b
}

// WithReActSystemPrompt sets a custom system prompt for ReAct execution.
// If empty (default), uses the built-in ReAct prompt template.
// Advanced users can customize the prompt format and examples.
//
// Default: "" (uses built-in prompt)
//
// Example:
//
//	customPrompt := `You are an agent. Use this format:
//	THOUGHT: [reasoning]
//	ACTION: [tool(args)]
//	FINAL: [answer]`
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActSystemPrompt(customPrompt).
//	    Build()
func (b *Builder) WithReActSystemPrompt(prompt string) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.SystemPrompt = prompt
	return b
}

// WithReActCallback sets a callback handler for execution events.
// The callback receives notifications for each step, tool call, error, and completion.
// Useful for logging, monitoring, and real-time progress updates.
//
// Default: nil (no callbacks)
//
// Example:
//
//	type MyCallback struct{}
//
//	func (c *MyCallback) OnStep(step ReActStep) {
//	    fmt.Printf("[%s] %s\n", step.Type, step.Content)
//	}
//	// ... implement other methods
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActCallback(&MyCallback{}).
//	    Build()
func (b *Builder) WithReActCallback(callback ReActCallback) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.Callback = callback
	return b
}

// WithReActMetrics enables or disables execution metrics collection.
// When enabled, ReActResult.Metrics will contain detailed execution statistics.
// Minimal performance overhead.
//
// Default: false
//
// Example:
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActMetrics(true).  // Collect metrics
//	    Build()
//
//	result, _ := ai.Execute(ctx, "task")
//	fmt.Printf("Iterations: %d, Duration: %v\n",
//	    result.Metrics.TotalIterations,
//	    result.Metrics.Duration)
func (b *Builder) WithReActMetrics(enabled bool) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.EnableMetrics = enabled
	return b
}

// WithReActTimeline enables or disables execution timeline tracking.
// When enabled, ReActResult.Timeline will contain a chronological log of all events.
// Useful for debugging and performance analysis.
//
// Default: false
//
// Example:
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActTimeline(true).  // Track timeline
//	    Build()
//
//	result, _ := ai.Execute(ctx, "task")
//	for _, event := range result.Timeline.Events {
//	    fmt.Printf("[%s] %s\n", event.Type, event.Content)
//	}
func (b *Builder) WithReActTimeline(enabled bool) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.EnableTimeline = enabled
	return b
}

// Execute runs a task using the ReAct pattern (if enabled).
// If ReAct is disabled, behaves like Ask().
//
// The ReAct pattern alternates between reasoning and acting:
// 1. THOUGHT: Agent reasons about what to do
// 2. ACTION: Agent calls a tool
// 3. OBSERVATION: System provides tool result
// 4. Repeat until FINAL answer
//
// Example:
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithTool(searchTool).
//	    WithReActMode(true).
//	    Build()
//
//	result, err := ai.Execute(ctx, "Search for weather in Paris and summarize")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Answer: %s\n", result.Answer)
//	fmt.Printf("Steps taken: %d\n", len(result.Steps))
//	for _, step := range result.Steps {
//	    fmt.Printf("[%s] %s\n", step.Type, step.Content)
//	}
func (b *Builder) Execute(ctx context.Context, task string) (*ReActResult, error) {
	// Check if ReAct is enabled
	if b.reactConfig == nil || !b.reactConfig.Enabled {
		// Fallback to normal Ask()
		response, err := b.Ask(ctx, task)
		if err != nil {
			return &ReActResult{
				Success: false,
				Error:   err,
			}, err
		}

		// Return as single-step result
		return &ReActResult{
			Answer:  response,
			Success: true,
			Steps: []ReActStep{
				{
					Type:      StepTypeFinal,
					Content:   response,
					Timestamp: time.Now(),
				},
			},
			Iterations: 1,
		}, nil
	}

	// ReAct execution - delegate to core loop
	return b.executeReAct(ctx, task)
}

// executeReAct implements the core ReAct reasoning + acting loop.
// This is the main execution logic for ReAct pattern.
func (b *Builder) executeReAct(ctx context.Context, task string) (*ReActResult, error) {
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
		result.Timeline.AddEvent("start", "ReAct execution started", 0, nil)
	}

	// Apply timeout if configured
	if b.reactConfig.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.reactConfig.Timeout)
		defer cancel()
	}

	// Build conversation history with system prompt
	messages := []Message{}

	// Add ReAct system prompt
	systemPrompt := b.reactConfig.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = b.buildReActSystemPrompt()
	}
	messages = append(messages, System(systemPrompt))

	// Add user task
	messages = append(messages, User(task))

	// Execution loop
	for iteration := 0; iteration < b.reactConfig.MaxIterations; iteration++ {
		if b.reactConfig.EnableTimeline {
			result.Timeline.AddEvent("iteration_start", fmt.Sprintf("Iteration %d started", iteration+1), 0, nil)
		}

		// Call LLM
		response, err := b.askWithMessages(ctx, messages)
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

		// Parse the response
		stepType, content, tool, args, parseErr := parseReActStep(response)

		// Handle parse errors
		if parseErr != nil {
			if b.reactConfig.Strict {
				// Strict mode: fail immediately
				result.Success = false
				result.Error = fmt.Errorf("parse error at iteration %d: %w", iteration+1, parseErr)

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("parse_error", fmt.Sprintf("Parse error: %v", parseErr), 0, nil)
				}

				if b.reactConfig.Callback != nil {
					b.reactConfig.Callback.OnError(result.Error)
				}

				return result, result.Error
			}

			// Graceful mode: try self-correction first (one retry)
			if iteration < b.reactConfig.MaxIterations-1 {
				// Add assistant's malformed response
				messages = append(messages, Assistant(response))

				// Add correction prompt
				correctionPrompt := b.buildCorrectionPrompt(parseErr, response)
				messages = append(messages, User(correctionPrompt))

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("parse_error_retry", "Attempting self-correction", 0, nil)
				}

				if b.reactConfig.EnableMetrics {
					result.Metrics.Errors++
				}

				// Continue to next iteration for retry
				continue
			}

			// Max iterations or retry failed: treat as FINAL answer
			stepType = StepTypeFinal
			content = response

			if b.reactConfig.EnableMetrics {
				result.Metrics.Errors++
			}
		}

		// Create step
		step := ReActStep{
			Type:      stepType,
			Content:   content,
			Tool:      tool,
			Args:      args,
			Timestamp: time.Now(),
			Error:     parseErr,
		}

		result.Steps = append(result.Steps, step)

		// Callback: onStep
		if b.reactConfig.Callback != nil {
			b.reactConfig.Callback.OnStep(step)
		}

		// Handle different step types
		switch stepType {
		case StepTypeThought:
			// Add THOUGHT to conversation
			messages = append(messages, Assistant(response))

			if b.reactConfig.EnableTimeline {
				result.Timeline.AddEvent("thought", content, 0, nil)
			}

			// Continue to next iteration (LLM should produce ACTION next)

		case StepTypeAction:
			// Add ACTION to conversation
			messages = append(messages, Assistant(response))

			if b.reactConfig.EnableMetrics {
				result.Metrics.ToolCalls++
			}

			if b.reactConfig.EnableTimeline {
				result.Timeline.AddEvent("action", fmt.Sprintf("Tool: %s, Args: %v", tool, args), 0, nil)
			}

			// Execute the tool
			observation, toolErr := b.executeTool(ctx, tool, args)

			if toolErr != nil {
				// Handle tool error
				if b.reactConfig.Strict {
					result.Success = false
					result.Error = fmt.Errorf("tool execution failed: %w", toolErr)

					if b.reactConfig.EnableMetrics {
						result.Metrics.Errors++
					}

					if b.reactConfig.Callback != nil {
						b.reactConfig.Callback.OnError(result.Error)
					}

					return result, result.Error
				}

				// Graceful mode: inject error as observation with guidance
				observation = fmt.Sprintf("Tool execution failed: %v\n\nPlease try:\n1. Use different parameters\n2. Use a different tool\n3. Provide a final answer based on available information", toolErr)

				if b.reactConfig.EnableMetrics {
					result.Metrics.Errors++
				}

				if b.reactConfig.EnableTimeline {
					result.Timeline.AddEvent("tool_error", fmt.Sprintf("Tool %s failed: %v", tool, toolErr), 0, nil)
				}
			}

			// Add OBSERVATION to conversation
			obsMessage := fmt.Sprintf("OBSERVATION: %s", observation)
			messages = append(messages, User(obsMessage))

			// Record observation step
			obsStep := ReActStep{
				Type:      StepTypeObservation,
				Content:   observation,
				Timestamp: time.Now(),
				Error:     toolErr,
			}
			result.Steps = append(result.Steps, obsStep)

			if b.reactConfig.EnableTimeline {
				result.Timeline.AddEvent("observation", observation, 0, nil)
			}

			// Callback: onToolCall
			if b.reactConfig.Callback != nil {
				// Note: We're adapting our callback to match OpenAI's FinishedChatCompletionToolCall
				// For now, just notify about the tool call
				b.reactConfig.Callback.OnStep(obsStep)
			}

		case StepTypeFinal:
			// Final answer reached
			result.Answer = content
			result.Success = true
			result.Iterations = iteration + 1

			if b.reactConfig.EnableTimeline {
				result.Timeline.AddEvent("final", content, 0, nil)
			}

			// Finalize metrics
			if b.reactConfig.EnableMetrics {
				result.Metrics.Finalize()
			}

			// Callback: onComplete
			if b.reactConfig.Callback != nil {
				b.reactConfig.Callback.OnComplete(result)
			}

			return result, nil

		default:
			// Unknown step type
			result.Success = false
			result.Error = fmt.Errorf("unknown step type: %s", stepType)
			return result, result.Error
		}
	}

	// Max iterations reached without FINAL answer
	result.Success = false
	result.Error = fmt.Errorf("max iterations (%d) reached without final answer", b.reactConfig.MaxIterations)
	result.Iterations = b.reactConfig.MaxIterations

	if b.reactConfig.EnableMetrics {
		result.Metrics.Finalize()
	}

	if b.reactConfig.EnableTimeline {
		result.Timeline.AddEvent("max_iterations", "Maximum iterations reached", 0, nil)
	}

	if b.reactConfig.Callback != nil {
		b.reactConfig.Callback.OnError(result.Error)
	}

	return result, result.Error
}

// buildReActSystemPrompt generates the default ReAct system prompt with tool descriptions.
func (b *Builder) buildReActSystemPrompt() string {
	// Use custom template if provided
	if b.reactConfig != nil && b.reactConfig.SystemPrompt != "" {
		// Build variable substitution map
		vars := b.buildTemplateVariables()
		// Render template with variables
		return RenderTemplate(b.reactConfig.SystemPrompt, vars)
	}

	// Use default template
	prompt := `You are a helpful AI assistant that uses the ReAct (Reasoning + Acting) pattern to solve problems.

Follow this format EXACTLY:

THOUGHT: [Your reasoning about what to do next]
ACTION: tool_name(arg1="value1", arg2="value2")
OBSERVATION: [Tool result will be provided by the system]
... (repeat THOUGHT/ACTION/OBSERVATION as needed)
FINAL: [Your final answer to the user]

Rules:
1. Always start with a THOUGHT to reason about the problem
2. Use ACTION to call available tools when you need information
3. Wait for OBSERVATION before continuing
4. Use FINAL when you have enough information to answer
5. Be concise and focused in your reasoning

Available tools:
`

	// Add tool descriptions
	if len(b.tools) == 0 {
		prompt += "(No tools available)\n"
	} else {
		for _, tool := range b.tools {
			prompt += fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description)
		}
	}

	// Add few-shot examples if provided
	if b.reactConfig != nil && len(b.reactConfig.Examples) > 0 {
		prompt += "\n" + FormatExamples(b.reactConfig.Examples)
	}

	return prompt
}

// buildCorrectionPrompt generates a self-correction prompt when parsing fails.
// This helps the LLM understand what went wrong and fix its format.
func (b *Builder) buildCorrectionPrompt(parseErr error, response string) string {
	prompt := fmt.Sprintf(`Your previous response could not be parsed correctly.

Error: %v

Your response was:
%s

Please follow the EXACT format:

THOUGHT: [your reasoning]
ACTION: tool_name(arg1="value1", arg2="value2")
FINAL: [your final answer]

Make sure to:
1. Use UPPERCASE for keywords (THOUGHT, ACTION, FINAL)
2. Put tool arguments in parentheses with quotes
3. Use only ONE keyword per response
4. Keep the format simple and clear

Try again:`, parseErr, response)

	return prompt
}

// buildToolErrorPrompt generates a prompt when tool execution fails.
// This helps the LLM understand the error and try alternative approaches.
func (b *Builder) buildToolErrorPrompt(toolName string, toolErr error) string {
	prompt := fmt.Sprintf(`The tool "%s" encountered an error:

Error: %v

Please either:
1. Try a different approach or different tool
2. Modify your parameters and try again
3. Provide a FINAL answer based on available information

Continue with your reasoning:`, toolName, toolErr)

	return prompt
}

// askWithMessages sends a message array to the LLM and returns the text response.
// This is a helper for ReAct execution loop.
func (b *Builder) askWithMessages(ctx context.Context, messages []Message) (string, error) {
	// Ensure client is initialized
	if err := b.ensureClient(); err != nil {
		return "", fmt.Errorf("failed to initialize client: %w", err)
	}

	// Convert to OpenAI format
	openaiMessages := convertMessages(messages)

	// Execute request
	completion, err := b.executeSyncRaw(ctx, openaiMessages)
	if err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return completion.Choices[0].Message.Content, nil
}

// executeTool executes a single tool by name with the given arguments.
// Returns the tool result or an error.
func (b *Builder) executeTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
	// Find the tool
	var targetTool *Tool
	for _, tool := range b.tools {
		if tool.Name == toolName {
			targetTool = tool
			break
		}
	}

	if targetTool == nil {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	// Check if handler is set
	if targetTool.Handler == nil {
		return "", fmt.Errorf("tool %s has no handler", toolName)
	}

	// Convert args to JSON string for handler
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool arguments: %w", err)
	}

	// Execute the tool handler
	result, err := targetTool.Handler(string(argsJSON))
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}
