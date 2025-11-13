// Package adapters provides LLM provider-specific implementations of the LLMAdapter interface.
package adapters

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/taipm/go-deep-agent/agent"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GeminiAdapter wraps the Google Generative AI Go SDK to implement the LLMAdapter interface.
// This adapter handles all Gemini-specific API calls and data conversions.
//
// Key differences from OpenAI:
//   - System prompt via SystemInstruction (not a message)
//   - Role names: "user" and "model" (not "assistant")
//   - Temperature range: 0.0 to 1.0 (needs clamping)
//   - Content is structured as "parts" not simple strings
//   - Streaming uses iterator pattern (not SSE)
type GeminiAdapter struct {
	client *genai.Client
}

// NewGeminiAdapter creates a new adapter for Google Gemini.
//
// Parameters:
//   - apiKey: Your Google AI API key
//
// Example:
//
//	adapter, err := NewGeminiAdapter("your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewGeminiAdapter(apiKey string) (*GeminiAdapter, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return &GeminiAdapter{client: client}, nil
}

// Close closes the Gemini client and releases resources.
func (a *GeminiAdapter) Close() error {
	if a.client == nil {
		return nil
	}
	return a.client.Close()
}

// Complete sends a synchronous completion request to Gemini.
func (a *GeminiAdapter) Complete(ctx context.Context, req *agent.CompletionRequest) (*agent.CompletionResponse, error) {
	// Create and configure model
	model := a.client.GenerativeModel(req.Model)
	a.configureModel(model, req)

	// Convert messages to Gemini format
	parts := a.convertMessagesToParts(req.Messages)

	// Call Gemini API
	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	// Convert response to unified format
	return a.convertResponse(resp), nil
}

// Stream sends a streaming completion request to Gemini.
func (a *GeminiAdapter) Stream(ctx context.Context, req *agent.CompletionRequest, onChunk func(string)) (*agent.CompletionResponse, error) {
	// Create and configure model
	model := a.client.GenerativeModel(req.Model)
	a.configureModel(model, req)

	// Convert messages to Gemini format
	parts := a.convertMessagesToParts(req.Messages)

	// Create streaming iterator
	iter := model.GenerateContentStream(ctx, parts...)

	var fullContent string
	var usage agent.TokenUsage
	var finishReason string

	// Process stream
	for {
		chunk, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gemini streaming error: %w", err)
		}

		// Extract content from chunk
		if len(chunk.Candidates) > 0 {
			candidate := chunk.Candidates[0]

			// Extract text from parts
			for _, part := range candidate.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					content := string(txt)
					fullContent += content
					if onChunk != nil {
						onChunk(content)
					}
				}
			}

			// Track finish reason
			if candidate.FinishReason != genai.FinishReasonUnspecified {
				finishReason = candidate.FinishReason.String()
			}
		}

		// Track usage (last chunk has final counts)
		if chunk.UsageMetadata != nil {
			usage = agent.TokenUsage{
				PromptTokens:     int(chunk.UsageMetadata.PromptTokenCount),
				CompletionTokens: int(chunk.UsageMetadata.CandidatesTokenCount),
				TotalTokens:      int(chunk.UsageMetadata.TotalTokenCount),
			}
		}
	}

	return &agent.CompletionResponse{
		Content:      fullContent,
		Usage:        usage,
		FinishReason: finishReason,
	}, nil
}

// configureModel sets model parameters from CompletionRequest.
func (a *GeminiAdapter) configureModel(model *genai.GenerativeModel, req *agent.CompletionRequest) {
	// System instruction (Gemini-specific way to set system prompt)
	if req.System != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.System)},
		}
	}

	// Temperature (Gemini supports 0-1, clamp if needed)
	if req.Temperature > 0 {
		temp := float32(req.Temperature)
		if temp > 1.0 {
			temp = 1.0 // Clamp to Gemini's range
		}
		model.SetTemperature(temp)
	}

	// Max output tokens
	if req.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(req.MaxTokens))
	}

	// Top P
	if req.TopP > 0 {
		model.SetTopP(float32(req.TopP))
	}

	// Stop sequences
	if len(req.Stop) > 0 {
		model.StopSequences = req.Stop
	}

	// Tools (if any) - Gemini supports function calling
	if len(req.Tools) > 0 {
		model.Tools = a.convertTools(req.Tools)
	}
}

// convertMessagesToParts converts our Message array to Gemini Parts.
// Gemini doesn't use a messages array like OpenAI - it uses parts directly.
func (a *GeminiAdapter) convertMessagesToParts(messages []agent.Message) []genai.Part {
	parts := []genai.Part{}

	for _, msg := range messages {
		// For Gemini, we primarily care about user messages
		// System prompt is handled separately via SystemInstruction
		// Assistant messages are typically not needed in the prompt (they're in history)

		if msg.Role == "user" || msg.Role == "assistant" {
			parts = append(parts, genai.Text(msg.Content))
		}
	}

	return parts
}

// convertTools converts our Tool definitions to Gemini format.
func (a *GeminiAdapter) convertTools(tools []*agent.Tool) []*genai.Tool {
	geminiTools := make([]*genai.Tool, 0, len(tools))

	for _, tool := range tools {
		// Convert parameters map to Gemini Schema
		// For now, we'll use a simple conversion
		// In production, this should properly convert the JSON schema
		schema := &genai.Schema{
			Type: genai.TypeObject,
		}

		// Gemini uses FunctionDeclaration
		funcDecl := &genai.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  schema,
		}

		geminiTools = append(geminiTools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		})
	}

	return geminiTools
}

// convertResponse converts Gemini's response to our unified format.
func (a *GeminiAdapter) convertResponse(resp *genai.GenerateContentResponse) *agent.CompletionResponse {
	result := &agent.CompletionResponse{}

	// Extract content from candidates
	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]

		// Extract text content from parts
		for _, part := range candidate.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				result.Content += string(txt)
			}
		}

		// Extract finish reason
		if candidate.FinishReason != genai.FinishReasonUnspecified {
			result.FinishReason = candidate.FinishReason.String()
		}

		// Extract tool calls if present
		for _, part := range candidate.Content.Parts {
			if funcCall, ok := part.(genai.FunctionCall); ok {
				// Convert function call to our ToolCall format
				argsJSON := fmt.Sprintf("%v", funcCall.Args) // Simplified - should use proper JSON marshaling
				result.ToolCalls = append(result.ToolCalls, agent.ToolCall{
					ID:        "", // Gemini doesn't provide IDs like OpenAI
					Type:      "function",
					Name:      funcCall.Name,
					Arguments: argsJSON,
				})
			}
		}
	}

	// Extract usage metadata
	if resp.UsageMetadata != nil {
		result.Usage = agent.TokenUsage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return result
}
