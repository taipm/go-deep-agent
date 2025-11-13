// Package agent provides a unified interface for multiple LLM providers.
// The adapter pattern allows seamless integration of different LLM providers
// (OpenAI, Gemini, Anthropic, etc.) while maintaining a consistent API.
package agent

import (
	"context"
)

// LLMAdapter abstracts provider-specific LLM implementations.
// This interface provides a thin abstraction layer that allows different
// LLM providers to be used interchangeably without changing the Builder's logic.
//
// Implementations are responsible for:
//   - Converting unified CompletionRequest to provider-specific formats
//   - Calling the provider's SDK with appropriate parameters
//   - Converting provider-specific responses back to CompletionResponse
//
// The interface is intentionally minimal (only 2 methods) to keep implementations
// simple and maintainable. Each adapter should be ~150-200 lines of code.
type LLMAdapter interface {
	// Complete sends a synchronous completion request and returns the full response.
	// This method blocks until the entire response is received from the provider.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - req: Unified completion request with all parameters
	//
	// Returns:
	//   - CompletionResponse: The complete response from the LLM
	//   - error: Any error that occurred during the request
	//
	// Example usage:
	//   resp, err := adapter.Complete(ctx, &CompletionRequest{
	//       Model: "gpt-4",
	//       Messages: []Message{{Role: "user", Content: "Hello"}},
	//   })
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Stream sends a streaming completion request.
	// The onChunk callback is called for each content chunk received from the provider.
	// This enables real-time display of the AI's response as it's being generated.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - req: Unified completion request with all parameters
	//   - onChunk: Callback function invoked for each content chunk
	//
	// Returns:
	//   - CompletionResponse: The complete accumulated response after streaming finishes
	//   - error: Any error that occurred during streaming
	//
	// Example usage:
	//   resp, err := adapter.Stream(ctx, req, func(chunk string) {
	//       fmt.Print(chunk) // Display each chunk as it arrives
	//   })
	Stream(ctx context.Context, req *CompletionRequest, onChunk func(string)) (*CompletionResponse, error)
}

// CompletionRequest contains all parameters for an LLM completion request.
// This unified format is used internally and converted to provider-specific
// formats by each adapter implementation.
//
// Design principles:
//   - Provider-agnostic: Works for all LLM providers
//   - Complete: Includes all common parameters across providers
//   - Extensible: Uses interface{} for provider-specific options
type CompletionRequest struct {
	// Model is the identifier of the model to use.
	// Examples: "gpt-4", "gpt-4o-mini", "gemini-pro", "claude-3-opus"
	//
	// Note: Model names are provider-specific. Each adapter is responsible
	// for using the model name as provided by the user.
	Model string

	// Messages is the conversation history.
	// This is the core input to the LLM, containing the entire conversation context.
	//
	// Each message has a role ("system", "user", "assistant", "tool") and content.
	// The order matters - it represents the chronological flow of the conversation.
	Messages []Message

	// System is the system prompt that sets the behavior of the assistant.
	// Different providers handle this differently:
	//   - OpenAI: Converted to a message with role "system"
	//   - Gemini: Set via SystemInstruction parameter
	//   - Anthropic: Set via separate "system" parameter
	//
	// The adapter is responsible for handling this appropriately.
	System string

	// Temperature controls randomness in the output (typically 0.0 to 2.0).
	// Lower values (e.g., 0.2) make output more focused and deterministic.
	// Higher values (e.g., 1.5) make output more random and creative.
	//
	// Default: Usually 1.0
	// Range varies by provider:
	//   - OpenAI: 0.0 to 2.0
	//   - Gemini: 0.0 to 1.0
	//   - Anthropic: 0.0 to 1.0
	//
	// Adapters should clamp values to their provider's supported range.
	Temperature float64

	// MaxTokens is the maximum number of tokens to generate.
	// This limits the length of the response.
	//
	// Note: Some providers (like Anthropic) require this parameter.
	// A value of 0 typically means "use provider default".
	MaxTokens int

	// TopP controls nucleus sampling (typically 0.0 to 1.0).
	// An alternative to temperature for controlling randomness.
	// Lower values make output more focused, higher values more diverse.
	//
	// Example: 0.1 means only consider tokens with top 10% probability mass.
	// Default: Usually 1.0 (consider all tokens)
	TopP float64

	// Stop is a list of sequences where the model will stop generating.
	// When the model generates any of these sequences, generation stops.
	//
	// Example: []string{"\n\n", "User:", "###"}
	// Useful for structured outputs or preventing unwanted continuations.
	Stop []string

	// Seed is for reproducible outputs (deterministic generation).
	// When set, multiple requests with the same seed should produce similar outputs.
	//
	// Note: Not all providers support this. When supported:
	//   - OpenAI: Full support with seed parameter
	//   - Gemini: Limited support
	//   - Anthropic: Not supported
	//
	// A value of 0 typically means "don't use seeding".
	Seed int64

	// Tools is the list of tools (functions) the model can call.
	// Enables function calling / tool use capabilities.
	//
	// When tools are provided, the model can choose to call these functions
	// instead of (or in addition to) generating text content.
	//
	// See Tool type for structure definition.
	Tools []*Tool

	// ToolChoice controls how the model uses tools.
	// This field is intentionally interface{} because tool choice semantics
	// vary significantly between providers:
	//
	//   - OpenAI: "auto", "none", "required", or {"type": "function", "function": {"name": "..."}}
	//   - Gemini: Different structure
	//   - Anthropic: Different structure
	//
	// The adapter is responsible for converting this appropriately.
	// In most cases, users won't set this directly - they'll use Builder methods.
	ToolChoice interface{}

	// ResponseFormat controls the format of the response.
	// Used for structured outputs (e.g., JSON mode).
	//
	// This is provider-specific and passed through as-is to the adapter.
	// Common use case:
	//   - OpenAI: {"type": "json_object"} for JSON mode
	//   - Other providers may have different formats
	//
	// In most cases, users won't set this directly - they'll use Builder methods.
	ResponseFormat interface{}

	// PresencePenalty penalizes tokens based on whether they appear in the text so far.
	// Range: -2.0 to 2.0
	// Positive values encourage the model to talk about new topics.
	//
	// Note: Not supported by all providers (primarily OpenAI).
	PresencePenalty float64

	// FrequencyPenalty penalizes tokens based on their frequency in the text so far.
	// Range: -2.0 to 2.0
	// Positive values reduce repetition.
	//
	// Note: Not supported by all providers (primarily OpenAI).
	FrequencyPenalty float64

	// LogProbs indicates whether to return log probabilities of output tokens.
	// Useful for understanding model confidence and decision-making.
	//
	// Note: Primarily supported by OpenAI.
	LogProbs bool

	// TopLogProbs is the number of most likely tokens to return log probabilities for.
	// Range: 0 to 20 (OpenAI)
	// Requires LogProbs to be true.
	TopLogProbs int

	// N is the number of completion choices to generate.
	// Default: 1
	//
	// Note: Most use cases only need 1. Higher values increase cost and latency.
	// Not all providers support this.
	N int
}

// CompletionResponse contains the standardized LLM response.
// All adapters must convert their provider-specific responses to this format.
//
// This unified response format allows the Builder and other components
// to work with responses in a consistent way, regardless of the provider.
type CompletionResponse struct {
	// Content is the main text content generated by the model.
	// This is the assistant's response to the user's prompt.
	//
	// For non-streaming requests, this contains the complete response.
	// For streaming requests, this contains the accumulated response.
	Content string

	// ToolCalls contains any tool/function calls requested by the model.
	// When the model decides to use a tool, this array will be populated.
	//
	// The Builder's auto-execute feature can automatically execute these
	// tool calls and feed the results back to the model.
	//
	// Empty if the model didn't request any tool calls.
	ToolCalls []ToolCall

	// Usage contains token consumption statistics.
	// Useful for tracking costs and understanding model behavior.
	//
	// Note: In streaming mode, usage information may only be available
	// at the end of the stream (provider-dependent).
	Usage TokenUsage

	// FinishReason indicates why the model stopped generating.
	// Common values:
	//   - "stop": Natural completion (model decided to stop)
	//   - "length": Reached max_tokens limit
	//   - "tool_calls": Model requested tool calls
	//   - "content_filter": Content was filtered (safety)
	//   - "function_call": (deprecated) Legacy tool call format
	//
	// Values may vary slightly between providers.
	FinishReason string

	// Refusal contains the refusal message if the model refused to respond.
	// Some models (like OpenAI) may refuse requests that violate policies.
	//
	// When non-empty, Content will typically be empty.
	Refusal string

	// ID is the unique identifier for this completion (provider-specific).
	// Useful for logging, debugging, and tracking requests.
	ID string

	// Model is the actual model that was used (may differ from requested model).
	// Some providers may serve requests with a different model version.
	Model string

	// Created is the Unix timestamp when the response was created.
	Created int64
}

// ToolCall represents a request from the model to call a specific tool/function.
// This is returned when the model decides it needs to execute a function
// to gather more information or perform an action.
type ToolCall struct {
	// ID is the unique identifier for this tool call.
	// Used to associate tool results with the original call when sending back to the model.
	//
	// Example: "call_abc123"
	ID string

	// Type is the type of tool call.
	// Currently always "function" for function calling.
	// Future extensions might include other types.
	Type string

	// Name is the name of the function to call.
	// This corresponds to one of the Tool names provided in the request.
	//
	// Example: "get_weather", "search_database"
	Name string

	// Arguments contains the function arguments as a JSON string.
	// The adapter receives this from the provider and passes it through.
	//
	// Example: `{"location": "San Francisco", "units": "celsius"}`
	//
	// To use: json.Unmarshal([]byte(toolCall.Arguments), &params)
	Arguments string
}
