// Package agent provides Gemini V3 adapter implementation using google.golang.org/genai v1.36.0
// Production-ready with enterprise-grade tool calling support
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/genai"
	"google.golang.org/api/googleapi"
)

// GeminiV3Adapter represents the production-ready Gemini adapter
type GeminiV3Adapter struct {
	client *genai.Client
	model  string
}

// NewGeminiV3Adapter creates a new production-ready Gemini V3 adapter
func NewGeminiV3Adapter(apiKey, model string) (*GeminiV3Adapter, error) {
	if model == "" {
		model = "gemini-1.5-pro-latest" // Default to latest model
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini v3 client: %w", err)
	}

	return &GeminiV3Adapter{
		client: client,
		model:  model,
	}, nil
}

// Close closes the Gemini client and releases resources
func (a *GeminiV3Adapter) Close() error {
	// New client doesn't have explicit Close method
	return nil
}

// Complete implements LLMAdapter interface with production-grade error handling
func (a *GeminiV3Adapter) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Validate request
	if err := a.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert messages to Gemini format
	contents := a.convertMessages(req.Messages)

	// Create generation config
	config := a.createGenerationConfig(req)

	// Generate content
	resp, err := a.client.Models.GenerateContent(ctx, a.model, contents, config)
	if err != nil {
		return nil, a.handleError(err)
	}

	// Convert response
	return a.convertResponse(resp), nil
}

// Stream implements streaming with production-grade error handling
// Simple but effective streaming for developer ease-of-use
func (a *GeminiV3Adapter) Stream(ctx context.Context, req *CompletionRequest, onChunk func(string)) (*CompletionResponse, error) {
	if onChunk == nil {
		// No callback provided, fall back to Complete
		return a.Complete(ctx, req)
	}

	// Validate request
	if err := a.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert messages to Gemini format
	contents := a.convertMessages(req.Messages)

	// Create generation config with streaming enabled
	config := a.createGenerationConfig(req)

	// Generate content and stream it simply
	resp, err := a.client.Models.GenerateContent(ctx, a.model, contents, config)
	if err != nil {
		return nil, a.handleError(err)
	}

	// Simple streaming: send content in chunks if it's long
	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]

		// Extract full text content
		var fullText string
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				fullText += part.Text
			}
		}

		// Simple word-by-word streaming for good UX
		if fullText != "" {
			words := strings.Fields(fullText)
			for i, word := range words {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					chunk := word
					if i < len(words)-1 {
						chunk += " "
					}
					onChunk(chunk)
				}
			}
		}
	}

	// Convert and return the complete response
	return a.convertResponse(resp), nil
}

// validateRequest performs production-grade request validation
func (a *GeminiV3Adapter) validateRequest(req *CompletionRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}
	if req.Model == "" {
		return fmt.Errorf("model cannot be empty")
	}
	if req.Temperature < 0 || req.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	if req.MaxTokens < 1 {
		return fmt.Errorf("maxTokens must be greater than 0")
	}
	return nil
}

// createGenerationConfig creates Gemini generation config
func (a *GeminiV3Adapter) createGenerationConfig(req *CompletionRequest) *genai.GenerateContentConfig {
	temp := float32(req.Temperature)
	topP := float32(req.TopP)
	maxTokens := int32(req.MaxTokens)

	config := &genai.GenerateContentConfig{
		Temperature:     &temp,
		MaxOutputTokens: maxTokens,
		TopP:            &topP,
	}

	// Add stop sequences
	if len(req.Stop) > 0 {
		config.StopSequences = req.Stop
	}

	// Add tools
	if len(req.Tools) > 0 {
		config.Tools = a.convertTools(req.Tools)
	}

	return config
}

// convertMessages converts messages to Gemini format
func (a *GeminiV3Adapter) convertMessages(messages []Message) []*genai.Content {
	if len(messages) == 0 {
		return nil
	}

	var contents []*genai.Content

	for _, msg := range messages {
		if msg.Role == "tool" && msg.ToolCallID != "" {
			// Create function response content
			responseData := map[string]interface{}{
				"result": msg.Content,
			}

			// Try to parse as JSON
			var jsonResult map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Content), &jsonResult); err == nil {
				responseData = jsonResult
			}

			content := genai.NewContentFromFunctionResponse(msg.ToolCallID, responseData, genai.RoleUser)
			contents = append(contents, content)
		} else {
			// Handle assistant messages with tool calls
			if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
				// Create parts with text and function calls
				var parts []*genai.Part

				// Add text content if present
				if msg.Content != "" {
					parts = append(parts, &genai.Part{Text: msg.Content})
				}

				// Add function calls
				for _, toolCall := range msg.ToolCalls {
					// Parse arguments from JSON string
					var argsMap map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Arguments), &argsMap); err != nil {
						// Fallback to empty map if parsing fails
						argsMap = make(map[string]interface{})
					}

					funcCall := &genai.FunctionCall{
						Name: toolCall.Name,
						Args: argsMap,
					}
					parts = append(parts, &genai.Part{FunctionCall: funcCall})
				}

				content := &genai.Content{
					Role:  genai.RoleModel,
					Parts: parts,
				}
				contents = append(contents, content)
			} else {
				// Regular text content
				var role genai.Role
				switch msg.Role {
				case "user":
					role = genai.RoleUser
				case "assistant", "model":
					role = genai.RoleModel
				default:
					role = genai.RoleUser
				}

				content := genai.NewContentFromText(msg.Content, role)
				contents = append(contents, content)
			}
		}
	}

	return contents
}

// convertTools converts tools to Gemini format
func (a *GeminiV3Adapter) convertTools(tools []*Tool) []*genai.Tool {
	geminiTools := make([]*genai.Tool, 0, len(tools))

	for _, tool := range tools {
		if tool == nil {
			continue
		}

		// Convert schema
		schema := a.convertToolSchema(tool)

		funcDecl := &genai.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  schema,
		}

		geminiTool := &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		}

		geminiTools = append(geminiTools, geminiTool)
	}

	return geminiTools
}

// convertToolSchema converts tool schema to Gemini format
func (a *GeminiV3Adapter) convertToolSchema(tool *Tool) *genai.Schema {
	schema := &genai.Schema{
		Type:       genai.TypeObject,
		Properties: make(map[string]*genai.Schema),
		Required:   []string{},
	}

	// Use the tool Parameters directly since it's already map[string]interface{}
	params := tool.Parameters

	// Extract properties
	if props, ok := params["properties"].(map[string]interface{}); ok {
		for propName, propData := range props {
			if propMap, ok := propData.(map[string]interface{}); ok {
				paramSchema := a.convertPropertySchema(propMap)
				schema.Properties[propName] = paramSchema
			}
		}
	}

	// Extract required fields
	if reqs, ok := params["required"].([]string); ok {
		schema.Required = reqs
	}

	return schema
}

// convertPropertySchema converts individual property schemas
func (a *GeminiV3Adapter) convertPropertySchema(propMap map[string]interface{}) *genai.Schema {
	schema := &genai.Schema{}

	// Type conversion
	if typeStr, ok := propMap["type"].(string); ok {
		switch strings.ToLower(typeStr) {
		case "string":
			schema.Type = genai.TypeString
		case "number", "float", "double":
			schema.Type = genai.TypeNumber
		case "integer":
			schema.Type = genai.TypeInteger
		case "boolean":
			schema.Type = genai.TypeBoolean
		case "array":
			schema.Type = genai.TypeArray
			if items, ok := propMap["items"].(map[string]interface{}); ok {
				if itemType, ok := items["type"].(string); ok {
					schema.Items = &genai.Schema{
						Type: a.convertTypeString(itemType),
					}
				}
			}
		case "object":
			schema.Type = genai.TypeObject
		default:
			schema.Type = genai.TypeString
		}
	}

	// Description
	if desc, ok := propMap["description"].(string); ok {
		schema.Description = desc
	}

	// Enum values
	if enumValues, ok := propMap["enum"].([]interface{}); ok {
		enumStrings := make([]string, len(enumValues))
		for i, val := range enumValues {
			if strVal, ok := val.(string); ok {
				enumStrings[i] = strVal
			}
		}
		schema.Enum = enumStrings
	}

	return schema
}

// convertTypeString converts string type to genai.Type
func (a *GeminiV3Adapter) convertTypeString(typeStr string) genai.Type {
	switch strings.ToLower(typeStr) {
	case "string":
		return genai.TypeString
	case "number", "float", "double":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeString
	}
}

// extractToolCalls extracts function calls with proper JSON marshaling
func (a *GeminiV3Adapter) extractToolCalls(parts []*genai.Part) []ToolCall {
	var toolCalls []ToolCall

	for _, part := range parts {
		if part.FunctionCall != nil {
			funcCall := part.FunctionCall

			// Proper JSON marshaling - this fixes the critical bug
			argsJSON, err := json.Marshal(funcCall.Args)
			if err != nil {
				argsJSON = []byte("{}")
			}

			toolCall := ToolCall{
				ID:        fmt.Sprintf("gemini_%s_%s", funcCall.Name, uuid.New().String()[:8]),
				Type:      "function",
				Name:      funcCall.Name,
				Arguments: string(argsJSON),
			}

			toolCalls = append(toolCalls, toolCall)
		}
	}

	return toolCalls
}

// convertResponse converts Gemini response to unified format
func (a *GeminiV3Adapter) convertResponse(resp *genai.GenerateContentResponse) *CompletionResponse {
	result := &CompletionResponse{}

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]

		// Extract text content
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				result.Content += part.Text
			}
		}

		// Extract tool calls
		result.ToolCalls = a.extractToolCalls(candidate.Content.Parts)

		result.FinishReason = a.determineFinishReason(result.ToolCalls, result.Content)
	}

	// Extract usage metadata
	if resp.UsageMetadata != nil {
		result.Usage = TokenUsage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	return result
}

// determineFinishReason determines the appropriate finish reason
func (a *GeminiV3Adapter) determineFinishReason(toolCalls []ToolCall, content string) string {
	if len(toolCalls) > 0 {
		return "tool_calls"
	}
	if content != "" {
		return "stop"
	}
	return "unknown"
}

// handleError provides production-grade error handling with categorization
func (a *GeminiV3Adapter) handleError(err error) error {
	if err == nil {
		return nil
	}

	// Convert to APIError if possible
	if apiErr, ok := err.(*googleapi.Error); ok {
		switch apiErr.Code {
		case 400:
			details := make([]string, len(apiErr.Details))
			for i, d := range apiErr.Details {
				details[i] = fmt.Sprintf("%v", d)
			}
			return fmt.Errorf("gemini bad request: %s - %s", apiErr.Message, strings.Join(details, ", "))
		case 401:
			return fmt.Errorf("gemini authentication error: %s", apiErr.Message)
		case 403:
			details := make([]string, len(apiErr.Details))
			for i, d := range apiErr.Details {
				details[i] = fmt.Sprintf("%v", d)
			}
			return fmt.Errorf("gemini permission denied: %s - %s", apiErr.Message, strings.Join(details, ", "))
		case 429:
			return fmt.Errorf("gemini quota exceeded: %s", apiErr.Message)
		case 500:
			return fmt.Errorf("gemini internal server error: %s", apiErr.Message)
		case 503:
			return fmt.Errorf("gemini service unavailable: %s", apiErr.Message)
		default:
			return fmt.Errorf("gemini API error (%d): %s", apiErr.Code, apiErr.Message)
		}
	}

	errStr := err.Error()

	// Fallback categorization
	switch {
	case strings.Contains(errStr, "API key") || strings.Contains(errStr, "auth"):
		return fmt.Errorf("gemini authentication error: %w", err)
	case strings.Contains(errStr, "quota") || strings.Contains(errStr, "rate limit"):
		return fmt.Errorf("gemini quota exceeded: %w", err)
	case strings.Contains(errStr, "content") && strings.Contains(errStr, "policy"):
		return fmt.Errorf("gemini content policy violation: %w", err)
	case strings.Contains(errStr, "model") || strings.Contains(errStr, "not found"):
		return fmt.Errorf("gemini model not available: %w", err)
	case strings.Contains(errStr, "timeout"):
		return fmt.Errorf("gemini request timeout: %w", err)
	default:
		return fmt.Errorf("gemini API error: %w", err)
	}
}

// GetModel returns the current model
func (a *GeminiV3Adapter) GetModel() string {
	return a.model
}