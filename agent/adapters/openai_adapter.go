// Package adapters provides LLM provider-specific implementations of the LLMAdapter interface.
package adapters

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/taipm/go-deep-agent/agent"
)

// OpenAIAdapter wraps the OpenAI Go SDK to implement the LLMAdapter interface.
// This adapter handles all OpenAI-specific API calls and data conversions.
//
// It supports:
//   - OpenAI API (api.openai.com)
//   - OpenAI-compatible APIs (via custom baseURL)
//   - Azure OpenAI (via custom baseURL and headers)
//   - Ollama (via local endpoint)
type OpenAIAdapter struct {
	client *openai.Client
}

// NewOpenAIAdapter creates a new adapter for OpenAI or OpenAI-compatible APIs.
//
// Parameters:
//   - apiKey: Your OpenAI API key (e.g., "sk-...")
//   - baseURL: Custom base URL for OpenAI-compatible APIs (empty string for default OpenAI)
//
// Examples:
//
//	// Standard OpenAI
//	adapter := NewOpenAIAdapter("sk-...", "")
//
//	// Ollama (OpenAI-compatible)
//	adapter := NewOpenAIAdapter("ollama", "http://localhost:11434/v1")
//
//	// Azure OpenAI
//	adapter := NewOpenAIAdapter(apiKey, "https://your-resource.openai.azure.com")
func NewOpenAIAdapter(apiKey, baseURL string) *OpenAIAdapter {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := openai.NewClient(opts...)
	return &OpenAIAdapter{client: &client}
}

// Complete sends a synchronous completion request to OpenAI.
// This method converts the unified CompletionRequest to OpenAI's format,
// makes the API call, and converts the response back to our unified format.
func (a *OpenAIAdapter) Complete(ctx context.Context, req *agent.CompletionRequest) (*agent.CompletionResponse, error) {
	// Build OpenAI parameters
	params := a.buildChatCompletionParams(req)

	// Call OpenAI API
	completion, err := a.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
	}

	// Convert response to unified format
	return a.convertResponse(completion), nil
}

// Stream sends a streaming completion request to OpenAI.
// Content chunks are sent to the onChunk callback as they arrive.
// Returns the complete accumulated response after streaming finishes.
func (a *OpenAIAdapter) Stream(ctx context.Context, req *agent.CompletionRequest, onChunk func(string)) (*agent.CompletionResponse, error) {
	// Build OpenAI parameters
	params := a.buildChatCompletionParams(req)

	// Create streaming request
	stream := a.client.Chat.Completions.NewStreaming(ctx, params)
	acc := openai.ChatCompletionAccumulator{}
	var fullContent string

	// Process stream
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		// Check if content just finished
		if content, ok := acc.JustFinishedContent(); ok {
			fullContent = content
		}

		// Check if refusal just finished
		if refusal, ok := acc.JustFinishedRefusal(); ok {
			fullContent += refusal
		}

		// Stream delta content in real-time
		if onChunk != nil && len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			deltaContent := chunk.Choices[0].Delta.Content
			onChunk(deltaContent)
			// Accumulate content for final response (fallback)
			if fullContent == "" {
				fullContent += deltaContent
			}
		}
	}

	// Check for streaming errors
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("openai streaming error: %w", err)
	}

	// Build final response
	resp := &agent.CompletionResponse{
		Content: fullContent,
	}

	return resp, nil
}

// buildChatCompletionParams converts our unified CompletionRequest to OpenAI's parameter format.
func (a *OpenAIAdapter) buildChatCompletionParams(req *agent.CompletionRequest) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(req.Model),
		Messages: a.convertMessages(req),
	}

	// Add optional parameters only if they're set (non-zero values)

	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}

	if req.MaxTokens > 0 {
		params.MaxTokens = openai.Int(int64(req.MaxTokens))
	}

	if req.TopP > 0 {
		params.TopP = openai.Float(req.TopP)
	}

	// Note: Stop sequences currently not supported in this adapter
	// They would require proper parameter union type handling
	_ = req.Stop

	if req.Seed > 0 {
		params.Seed = openai.Int(req.Seed)
	}

	if req.PresencePenalty != 0 {
		params.PresencePenalty = openai.Float(req.PresencePenalty)
	}

	if req.FrequencyPenalty != 0 {
		params.FrequencyPenalty = openai.Float(req.FrequencyPenalty)
	}

	if req.LogProbs {
		params.Logprobs = openai.Bool(true)
		if req.TopLogProbs > 0 {
			params.TopLogprobs = openai.Int(int64(req.TopLogProbs))
		}
	}

	if req.N > 0 {
		params.N = openai.Int(int64(req.N))
	}

	// Add tools if present
	if len(req.Tools) > 0 {
		params.Tools = a.convertTools(req.Tools)
	}

	// Add tool choice if specified (provider-specific, pass through as-is)
	if req.ToolChoice != nil {
		// Tool choice is complex and provider-specific, users should set it directly
		// This is intentionally left simple
	}

	// Add response format if specified (provider-specific, pass through as-is)
	if req.ResponseFormat != nil {
		// Response format is complex and provider-specific
		// This is intentionally left simple
	}

	return params
}

// convertMessages converts our unified Message format to OpenAI's message format.
// Handles system, user, assistant, and tool messages.
func (a *OpenAIAdapter) convertMessages(req *agent.CompletionRequest) []openai.ChatCompletionMessageParamUnion {
	messages := []openai.ChatCompletionMessageParamUnion{}

	// Add system prompt as first message if provided
	if req.System != "" {
		messages = append(messages, openai.SystemMessage(req.System))
	}

	// Convert each message
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))

		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))

		case "assistant":
			// Simple assistant message (tool calls not yet supported in adapter)
			messages = append(messages, openai.AssistantMessage(msg.Content))

		case "tool":
			// Tool result message
			messages = append(messages, openai.ToolMessage(msg.ToolCallID, msg.Content))

		default:
			// Unknown role, default to user message
			messages = append(messages, openai.UserMessage(msg.Content))
		}
	}

	return messages
}

// convertTools converts our Tool definitions to OpenAI's tool parameter format.
func (a *OpenAIAdapter) convertTools(tools []*agent.Tool) []openai.ChatCompletionToolUnionParam {
	result := make([]openai.ChatCompletionToolUnionParam, len(tools))

	for i, tool := range tools {
		// Convert tool parameters to OpenAI's FunctionParameters type
		var funcParams openai.FunctionParameters
		if tool.Parameters != nil {
			funcParams = tool.Parameters
		}

		// Create function tool using OpenAI SDK helper
		result[i] = openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        tool.Name,
			Description: openai.String(tool.Description),
			Parameters:  funcParams,
		})
	}

	return result
}

// convertResponse converts OpenAI's response format to our unified CompletionResponse.
func (a *OpenAIAdapter) convertResponse(completion *openai.ChatCompletion) *agent.CompletionResponse {
	resp := &agent.CompletionResponse{
		ID:      completion.ID,
		Model:   completion.Model,
		Created: completion.Created,
	}

	// Handle empty response
	if len(completion.Choices) == 0 {
		return resp
	}

	// Get first choice (most common case)
	choice := completion.Choices[0]
	message := choice.Message

	// Extract content
	resp.Content = message.Content

	// For reasoning models (e.g., DeepSeek-R1, Qwen3 thinking models via Ollama)
	// If content is empty, try to extract reasoning from ExtraFields
	// Ollama reasoning models return reasoning field instead of content
	if resp.Content == "" && message.JSON.ExtraFields != nil {
		if reasoningField, ok := message.JSON.ExtraFields["reasoning"]; ok {
			// Raw() is a function that returns the raw string value
			resp.Content = reasoningField.Raw()
		}
	}

	// Extract finish reason
	resp.FinishReason = string(choice.FinishReason)

	// Extract refusal if present (OpenAI safety feature)
	resp.Refusal = message.Refusal

	// Convert tool calls if present
	if len(message.ToolCalls) > 0 {
		resp.ToolCalls = make([]agent.ToolCall, len(message.ToolCalls))
		for i, tc := range message.ToolCalls {
			resp.ToolCalls[i] = agent.ToolCall{
				ID:        tc.ID,
				Type:      string(tc.Type),
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	}

	// Extract usage statistics (Usage is a struct, not a pointer)
	resp.Usage = agent.TokenUsage{
		PromptTokens:     int(completion.Usage.PromptTokens),
		CompletionTokens: int(completion.Usage.CompletionTokens),
		TotalTokens:      int(completion.Usage.TotalTokens),
	}

	return resp
}
