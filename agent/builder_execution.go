// File: agent/builder_execution.go
// This file contains all execution-related methods for the Builder.
// Extracted from builder.go as part of refactoring to improve maintainability.

package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/taipm/go-deep-agent/agent/memory"
)

func (b *Builder) Ask(ctx context.Context, message string) (string, error) {
	start := time.Now()
	logger := b.getLogger()

	logger.Debug(ctx, "Ask request started",
		F("model", b.model),
		F("message_length", len(message)),
		F("has_cache", b.cacheEnabled),
		F("has_tools", len(b.tools) > 0),
		F("has_rag", b.ragEnabled))

	// Check for multimodal errors
	if b.lastError != nil {
		err := b.lastError
		b.lastError = nil // Clear error
		logger.Error(ctx, "Multimodal error detected", F("error", err.Error()))
		return "", err
	}

	// Ensure client is initialized
	if err := b.ensureClient(); err != nil {
		logger.Error(ctx, "Failed to initialize client", F("error", err.Error()))
		return "", fmt.Errorf("failed to initialize client: %w", err)
	}

	// Check cache first if enabled
	if b.cacheEnabled && b.cache != nil {
		cacheStart := time.Now()
		temp := 0.0
		if b.temperature != nil {
			temp = *b.temperature
		}
		cacheKey := GenerateCacheKey(b.model, message, temp, b.systemPrompt)
		if cached, found, err := b.cache.Get(ctx, cacheKey); err == nil && found {
			cacheDuration := time.Since(cacheStart)
			logger.Info(ctx, "Cache hit",
				F("cache_key", cacheKey),
				F("duration_ms", cacheDuration.Milliseconds()))
			return cached, nil
		} else {
			cacheDuration := time.Since(cacheStart)
			logger.Debug(ctx, "Cache miss",
				F("cache_key", cacheKey),
				F("duration_ms", cacheDuration.Milliseconds()))
		}
	}

	// If auto-execute is enabled and we have tools, use tool execution loop
	if b.autoExecute && len(b.tools) > 0 {
		logger.Debug(ctx, "Using tool execution loop", F("tool_count", len(b.tools)))
		return b.askWithToolExecution(ctx, message)
	}

	// RAG: Retrieve and inject relevant context if enabled
	if b.ragEnabled {
		ragStart := time.Now()
		docs, err := b.retrieveRelevantDocs(ctx, message)
		if err != nil {
			logger.Error(ctx, "RAG retrieval failed", F("error", err.Error()))
			return "", fmt.Errorf("RAG retrieval failed: %w", err)
		}

		b.lastRetrievedDocs = docs
		ragDuration := time.Since(ragStart)

		if len(docs) > 0 {
			logger.Debug(ctx, "RAG documents retrieved",
				F("doc_count", len(docs)),
				F("duration_ms", ragDuration.Milliseconds()))
			// Inject context into the message
			ragContext := b.buildRAGContext(docs)
			message = fmt.Sprintf("Context:\n%s\n\nQuestion: %s", ragContext, message)
		} else {
			logger.Debug(ctx, "No RAG documents found", F("duration_ms", ragDuration.Milliseconds()))
		}
	}

	// Build messages array (includes multimodal content if images added)
	messages := b.buildMessages(message)

	// Clear pending images after building messages
	b.pendingImages = nil

	// Execute request
	requestStart := time.Now()
	completion, err := b.executeSyncRaw(ctx, messages)
	if err != nil {
		requestDuration := time.Since(requestStart)
		logger.Error(ctx, "Request failed",
			F("error", err.Error()),
			F("duration_ms", requestDuration.Milliseconds()))
		return "", err
	}
	requestDuration := time.Since(requestStart)

	result := completion.Choices[0].Message.Content

	// Store in cache if enabled
	if b.cacheEnabled && b.cache != nil {
		temp := 0.0
		if b.temperature != nil {
			temp = *b.temperature
		}
		cacheKey := GenerateCacheKey(b.model, message, temp, b.systemPrompt)
		ttl := b.cacheTTL
		if ttl <= 0 {
			ttl = 5 * time.Minute // Default TTL
		}
		_ = b.cache.Set(ctx, cacheKey, result, ttl)
		logger.Debug(ctx, "Response cached", F("cache_key", cacheKey), F("ttl_seconds", ttl.Seconds()))
	}

	// Track token usage
	b.lastUsage = TokenUsage{
		PromptTokens:     int(completion.Usage.PromptTokens),
		CompletionTokens: int(completion.Usage.CompletionTokens),
		TotalTokens:      int(completion.Usage.TotalTokens),
	}

	totalDuration := time.Since(start)
	logger.Info(ctx, "Ask request completed",
		F("duration_ms", totalDuration.Milliseconds()),
		F("request_ms", requestDuration.Milliseconds()),
		F("prompt_tokens", b.lastUsage.PromptTokens),
		F("completion_tokens", b.lastUsage.CompletionTokens),
		F("total_tokens", b.lastUsage.TotalTokens),
		F("response_length", len(result)))

	// Hierarchical memory: store messages in memory system
	if b.memoryEnabled && b.memory != nil {
		memStart := time.Now()

		// Store user message
		userMsg := memory.Message{
			Role:      "user",
			Content:   message,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"model":      b.model,
				"has_images": len(b.pendingImages) > 0,
				"has_rag":    b.ragEnabled,
			},
		}
		if err := b.memory.Add(ctx, userMsg); err != nil {
			logger.Warn(ctx, "Failed to add user message to memory", F("error", err.Error()))
		}

		// Store assistant response
		assistantMsg := memory.Message{
			Role:      "assistant",
			Content:   result,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"prompt_tokens":     b.lastUsage.PromptTokens,
				"completion_tokens": b.lastUsage.CompletionTokens,
				"total_tokens":      b.lastUsage.TotalTokens,
			},
		}
		if err := b.memory.Add(ctx, assistantMsg); err != nil {
			logger.Warn(ctx, "Failed to add assistant message to memory", F("error", err.Error()))
		}

		memDuration := time.Since(memStart)
		logger.Debug(ctx, "Messages stored in memory",
			F("duration_ms", memDuration.Milliseconds()))
	}

	// Auto-memory: store this conversation turn (legacy FIFO)
	if b.autoMemory {
		b.addMessage(User(message))
		b.addMessage(Assistant(result))
	}

	return result, nil
}

func (b *Builder) askWithToolExecution(ctx context.Context, message string) (string, error) {
	logger := b.getLogger()
	logger.Debug(ctx, "Tool execution loop started", F("max_rounds", b.maxToolRounds))

	// Build messages array (includes multimodal content if images added)
	messages := b.buildMessages(message)

	// Clear pending images after building messages
	b.pendingImages = nil

	// Tool execution loop
	for round := 0; round < b.maxToolRounds; round++ {
		logger.Debug(ctx, "Tool execution round", F("round", round+1))

		// Build params with tools
		params := b.buildParams(messages)

		// Execute request
		completion, err := b.client.Chat.Completions.New(ctx, params)
		if err != nil {
			logger.Error(ctx, "Chat completion failed in tool loop",
				F("round", round+1),
				F("error", err.Error()))
			return "", fmt.Errorf("chat completion failed: %w", err)
		}

		if len(completion.Choices) == 0 {
			logger.Error(ctx, "No response choices returned", F("round", round+1))
			return "", fmt.Errorf("no response choices returned")
		}

		choice := completion.Choices[0]

		// Check if there are tool calls
		if len(choice.Message.ToolCalls) == 0 {
			// No tool calls, return the final response
			result := choice.Message.Content
			logger.Info(ctx, "Tool execution completed",
				F("rounds", round+1),
				F("response_length", len(result)))

			// Hierarchical memory: store messages in memory system
			if b.memoryEnabled && b.memory != nil {
				// Store user message
				userMsg := memory.Message{
					Role:      "user",
					Content:   message,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"tool_execution": true,
						"rounds":         round + 1,
					},
				}
				_ = b.memory.Add(ctx, userMsg)

				// Store assistant response
				assistantMsg := memory.Message{
					Role:      "assistant",
					Content:   result,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"tool_execution": true,
						"rounds":         round + 1,
					},
				}
				_ = b.memory.Add(ctx, assistantMsg)
			}

			// Auto-memory: store conversation
			if b.autoMemory {
				b.addMessage(User(message))
				b.addMessage(Assistant(result))
			}

			return result, nil
		}

		logger.Debug(ctx, "Tool calls received",
			F("round", round+1),
			F("tool_call_count", len(choice.Message.ToolCalls)))

		// Execute tool calls
		// Convert tool calls to param format
		toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			toolCallParams[i] = tc.ToParam()
		}

		// Add assistant message with tool calls
		assistantParam := openai.ChatCompletionAssistantMessageParam{
			ToolCalls: toolCallParams,
		}
		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfAssistant: &assistantParam,
		})

		// Execute tools (parallel or sequential based on config)
		var toolResults []openai.ChatCompletionMessageParamUnion
		var toolErr error

		if b.enableParallel && len(choice.Message.ToolCalls) > 1 {
			// Parallel execution for multiple tools
			toolResults, toolErr = b.executeToolsParallel(ctx, choice.Message.ToolCalls)
		} else {
			// Sequential execution (default or single tool)
			toolResults, toolErr = b.executeToolsSequential(ctx, choice.Message.ToolCalls)
		}

		if toolErr != nil {
			return "", fmt.Errorf("tool execution failed: %w", toolErr)
		}

		// Append all tool results to messages
		messages = append(messages, toolResults...)
	}

	logger.Warn(ctx, "Max tool rounds exceeded", F("max_rounds", b.maxToolRounds))
	return "", fmt.Errorf("max tool rounds (%d) exceeded", b.maxToolRounds)
} // AskMultiple sends a message and returns multiple completion choices.

func (b *Builder) AskMultiple(ctx context.Context, message string) ([]string, error) {
	// Ensure client is initialized
	if err := b.ensureClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	// Build messages array
	messages := b.buildMessages(message)

	// Execute request
	completion, err := b.executeSyncRaw(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Extract all choices
	results := make([]string, len(completion.Choices))
	for i, choice := range completion.Choices {
		results[i] = choice.Message.Content
	}

	// Auto-memory: store first choice only
	if b.autoMemory && len(results) > 0 {
		b.addMessage(User(message))
		b.addMessage(Assistant(results[0]))
	}

	return results, nil
}

func (b *Builder) Stream(ctx context.Context, message string) (string, error) {
	start := time.Now()
	logger := b.getLogger()

	logger.Debug(ctx, "Stream request started",
		F("model", b.model),
		F("message_length", len(message)))

	// Check for multimodal errors
	if b.lastError != nil {
		err := b.lastError
		b.lastError = nil // Clear error
		logger.Error(ctx, "Multimodal error detected in stream", F("error", err.Error()))
		return "", err
	}

	// Ensure client is initialized
	if err := b.ensureClient(); err != nil {
		logger.Error(ctx, "Failed to initialize client for stream", F("error", err.Error()))
		return "", fmt.Errorf("failed to initialize client: %w", err)
	}

	// Build messages array (includes multimodal content if images added)
	messages := b.buildMessages(message)

	// Clear pending images after building messages
	b.pendingImages = nil

	// Build params
	params := b.buildParams(messages)

	// Create streaming request
	logger.Debug(ctx, "Starting stream")
	stream := b.client.Chat.Completions.NewStreaming(ctx, params)

	// Use ChatCompletionAccumulator for full feature support
	acc := openai.ChatCompletionAccumulator{}
	var fullContent string
	chunkCount := 0

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		chunkCount++

		// Check if content just finished
		if content, ok := acc.JustFinishedContent(); ok {
			fullContent = content
			logger.Debug(ctx, "Stream content finished", F("content_length", len(content)))
		}

		// Check if a tool call just finished
		if toolCall, ok := acc.JustFinishedToolCall(); ok {
			logger.Debug(ctx, "Stream tool call finished", F("tool_index", toolCall.Index))
			if b.onToolCall != nil {
				b.onToolCall(toolCall)
			}
		}

		// Check if refusal just finished
		if refusal, ok := acc.JustFinishedRefusal(); ok {
			logger.Warn(ctx, "Stream refusal received", F("refusal", refusal))
			if b.onRefusal != nil {
				b.onRefusal(refusal)
			}
		}

		// Stream delta content in real-time
		if b.onStream != nil && len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			deltaContent := chunk.Choices[0].Delta.Content
			b.onStream(deltaContent)
			// Accumulate content for memory (fallback if JustFinishedContent doesn't work)
			fullContent += deltaContent
		}
	}

	if err := stream.Err(); err != nil {
		duration := time.Since(start)
		logger.Error(ctx, "Stream error",
			F("error", err.Error()),
			F("chunks_received", chunkCount),
			F("duration_ms", duration.Milliseconds()))
		return "", fmt.Errorf("stream error: %w", err)
	}

	duration := time.Since(start)
	logger.Info(ctx, "Stream completed",
		F("duration_ms", duration.Milliseconds()),
		F("chunks", chunkCount),
		F("response_length", len(fullContent)))

	// Hierarchical memory: store messages in memory system
	if b.memoryEnabled && b.memory != nil && fullContent != "" {
		// Store user message
		userMsg := memory.Message{
			Role:      "user",
			Content:   message,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"streaming": true,
			},
		}
		_ = b.memory.Add(ctx, userMsg)

		// Store assistant response
		assistantMsg := memory.Message{
			Role:      "assistant",
			Content:   fullContent,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"streaming": true,
				"chunks":    chunkCount,
			},
		}
		_ = b.memory.Add(ctx, assistantMsg)
	}

	// Auto-memory: store conversation (legacy FIFO)
	if b.autoMemory && fullContent != "" {
		b.addMessage(User(message))
		b.addMessage(Assistant(fullContent))
	}

	return fullContent, nil
}

func (b *Builder) StreamPrint(ctx context.Context, message string) (string, error) {
	return b.OnStream(func(content string) {
		fmt.Print(content)
	}).Stream(ctx, message)
} // ensureClient initializes the OpenAI client if not already initialized.

func (b *Builder) ensureClient() error {
	if b.client != nil {
		return nil
	}

	switch b.provider {
	case ProviderOpenAI:
		if b.apiKey == "" {
			return fmt.Errorf("API key is required for OpenAI")
		}
		client := openai.NewClient(option.WithAPIKey(b.apiKey))
		b.client = &client

	case ProviderOllama:
		if b.baseURL == "" {
			b.baseURL = "http://localhost:11434/v1"
		}
		client := openai.NewClient(
			option.WithBaseURL(b.baseURL),
			option.WithAPIKey("ollama"), // Ollama doesn't require a real key
		)
		b.client = &client

	default:
		return fmt.Errorf("unsupported provider: %s", b.provider)
	}

	return nil
}

func (b *Builder) buildMessages(userMessage string) []openai.ChatCompletionMessageParamUnion {
	result := []openai.ChatCompletionMessageParamUnion{}

	// Add system prompt if set
	if b.systemPrompt != "" {
		result = append(result, openai.SystemMessage(b.systemPrompt))
	}

	// Add conversation history (convert existing messages)
	result = append(result, convertMessages(b.messages)...)

	// Add current user message with multimodal support
	contentParts := b.buildContentParts(userMessage)

	// Check if we have multimodal content (array) or simple text (string)
	switch content := contentParts.(type) {
	case string:
		// Simple text message
		result = append(result, openai.UserMessage(content))
	case []openai.ChatCompletionContentPartUnionParam:
		// Multimodal message with images
		userMsg := openai.ChatCompletionUserMessageParam{
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfArrayOfContentParts: content,
			},
			Role: "user",
		}
		result = append(result, openai.ChatCompletionMessageParamUnion{
			OfUser: &userMsg,
		})
	}

	return result
}

func (b *Builder) buildParams(messages []openai.ChatCompletionMessageParamUnion) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(b.model),
		Messages: messages,
	}

	// Apply all advanced parameters
	if b.temperature != nil {
		params.Temperature = openai.Float(*b.temperature)
	}
	if b.topP != nil {
		params.TopP = openai.Float(*b.topP)
	}
	if b.maxTokens != nil {
		params.MaxTokens = openai.Int(*b.maxTokens)
	}
	if b.presencePenalty != nil {
		params.PresencePenalty = openai.Float(*b.presencePenalty)
	}
	if b.frequencyPenalty != nil {
		params.FrequencyPenalty = openai.Float(*b.frequencyPenalty)
	}
	if b.seed != nil {
		params.Seed = openai.Int(*b.seed)
	}
	if b.logprobs != nil {
		params.Logprobs = openai.Bool(*b.logprobs)
	}
	if b.topLogprobs != nil {
		params.TopLogprobs = openai.Int(*b.topLogprobs)
	}
	if b.n != nil {
		params.N = openai.Int(*b.n)
	}

	// Add tools if any
	if len(b.tools) > 0 {
		toolParams := make([]openai.ChatCompletionToolUnionParam, len(b.tools))
		for i, tool := range b.tools {
			toolParams[i] = tool.toOpenAI()
		}
		params.Tools = toolParams
	}

	// Add response format if set
	if b.responseFormat != nil {
		params.ResponseFormat = *b.responseFormat
	}

	return params
} // executeSyncRaw executes a synchronous (non-streaming) chat completion request and returns the full completion.

func (b *Builder) executeSyncRaw(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	// Use centralized param building to ensure all features (tools, responseFormat, etc.) are included
	params := b.buildParams(messages)

	completion, err := b.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	return completion, nil
}

func (b *Builder) executeWithRetry(ctx context.Context, operation func(context.Context) error) error {
	logger := b.getLogger()

	// Apply timeout if configured
	if b.timeout > 0 {
		logger.Debug(ctx, "Applying timeout", F("timeout_seconds", b.timeout.Seconds()))
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.timeout)
		defer cancel()
	}

	// No retries configured, execute once
	if b.maxRetries == 0 {
		err := operation(ctx)
		if err != nil && ctx.Err() == context.DeadlineExceeded {
			logger.Error(ctx, "Operation timed out", F("timeout", b.timeout.Seconds()))
			return WrapTimeout(err)
		}
		return err
	}

	logger.Debug(ctx, "Retry enabled", F("max_retries", b.maxRetries))

	// Execute with retries
	var lastErr error
	for attempt := 0; attempt <= b.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug(ctx, "Retry attempt", F("attempt", attempt+1), F("max", b.maxRetries+1))
		}

		// Execute operation
		err := operation(ctx)

		// Success
		if err == nil {
			if attempt > 0 {
				logger.Info(ctx, "Retry succeeded", F("attempt", attempt+1))
			}
			return nil
		}

		lastErr = err

		// Check if error is timeout
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error(ctx, "Operation timed out during retry",
				F("attempt", attempt+1),
				F("timeout", b.timeout.Seconds()))
			return WrapTimeout(err)
		}

		// Check if error is retryable
		if !b.isRetryable(err) {
			logger.Debug(ctx, "Error is not retryable",
				F("attempt", attempt+1),
				F("error", err.Error()))
			return err
		}

		// Last attempt failed
		if attempt == b.maxRetries {
			logger.Warn(ctx, "Max retries reached",
				F("attempts", attempt+1),
				F("error", err.Error()))
			break
		}

		// Calculate delay
		delay := b.calculateRetryDelay(attempt)
		logger.Debug(ctx, "Waiting before retry",
			F("attempt", attempt+1),
			F("delay_seconds", delay.Seconds()),
			F("error", err.Error()))

		// Wait before retry
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			logger.Error(ctx, "Context cancelled during retry wait", F("attempt", attempt+1))
			return WrapTimeout(ctx.Err())
		}
	}

	return WrapMaxRetries(b.maxRetries+1, lastErr)
}

func (b *Builder) isRetryable(err error) bool {
	// Retry on rate limit errors
	if IsRateLimitError(err) {
		return true
	}

	// Retry on timeout errors (if not from our timeout)
	if IsTimeoutError(err) {
		return true
	}

	// Don't retry on API key errors
	if IsAPIKeyError(err) {
		return false
	}

	// Don't retry on refusal errors
	if IsRefusalError(err) {
		return false
	}

	// Don't retry on invalid response errors
	if IsInvalidResponseError(err) {
		return false
	}

	// Default: don't retry
	return false
}

func (b *Builder) calculateRetryDelay(attempt int) time.Duration {
	if b.useExpBackoff {
		// Exponential backoff: delay * 2^attempt
		return b.retryDelay * (1 << uint(attempt))
	}
	// Fixed delay
	return b.retryDelay
}
