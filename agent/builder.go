package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
	"github.com/openai/openai-go/v3/shared/constant"
)

// Builder provides a fluent API for building and executing LLM requests.
// It supports method chaining for a natural, readable API.
//
// Example:
//
//	response := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithSystem("You are a helpful assistant").
//	    Ask(ctx, "Hello!")
type Builder struct {
	// Core configuration
	provider Provider
	model    string
	apiKey   string
	baseURL  string

	// Conversation state
	systemPrompt string
	messages     []Message
	autoMemory   bool // If true, automatically add messages to conversation history
	maxHistory   int  // Maximum number of messages to keep (0 = unlimited)

	// Advanced parameters
	temperature      *float64 // Sampling temperature (0-2)
	topP             *float64 // Nucleus sampling (0-1)
	maxTokens        *int64   // Maximum tokens to generate
	presencePenalty  *float64 // Presence penalty (-2.0 to 2.0)
	frequencyPenalty *float64 // Frequency penalty (-2.0 to 2.0)
	seed             *int64   // For reproducible outputs
	logprobs         *bool    // Return log probabilities
	topLogprobs      *int64   // Number of most likely tokens with log probs (0-20)
	n                *int64   // Number of chat completion choices to generate

	// Streaming callbacks
	onStream   func(content string)                             // Called for each content chunk
	onToolCall func(tool openai.FinishedChatCompletionToolCall) // Called when tool call finishes
	onRefusal  func(refusal string)                             // Called when refusal detected

	// Tool calling
	tools         []*Tool // Registered tools
	autoExecute   bool    // If true, automatically execute tool calls
	maxToolRounds int     // Maximum number of tool execution rounds (default 5)

	// Response format (structured outputs)
	responseFormat *openai.ChatCompletionNewParamsResponseFormatUnion

	// Error handling & recovery
	timeout       time.Duration // Request timeout (0 = no timeout)
	maxRetries    int           // Maximum retry attempts (0 = no retries)
	retryDelay    time.Duration // Base delay between retries (default 1s)
	useExpBackoff bool          // Use exponential backoff for retries

	// Multimodal support
	pendingImages []ImageContent // Images to include in next message
	lastError     error          // Last error from multimodal operations

	// Batch processing
	batchSize           int                        // Max concurrent requests in batch (default: 5)
	batchDelay          time.Duration              // Delay between batch chunks
	onBatchProgress     func(completed, total int) // Batch progress callback
	onBatchItemComplete func(result BatchResult)   // Individual item completion callback

	// RAG (Retrieval-Augmented Generation)
	ragEnabled        bool         // Whether RAG is enabled
	ragDocuments      []Document   // Documents for RAG
	ragRetriever      RAGRetriever // Custom retriever function
	ragConfig         *RAGConfig   // RAG configuration
	lastRetrievedDocs []Document   // Last retrieved documents

	// Vector RAG
	vectorStore       VectorStore       // Vector database for semantic search
	embeddingProvider EmbeddingProvider // Embedding provider for vector RAG
	vectorCollection  string            // Collection name for vector RAG

	// Caching
	cache        Cache         // Cache implementation
	cacheEnabled bool          // Whether caching is enabled
	cacheTTL     time.Duration // Cache TTL for next request

	// Logging
	logger Logger // Logger for observability (default: NoopLogger)

	// OpenAI client (lazy initialized)
	client *openai.Client

	// Usage tracking
	lastUsage TokenUsage // Last request token usage
}

// New creates a new Builder with the specified provider and model.
// Use NewOpenAI() or NewOllama() for convenience constructors.
//
// Example:
//
//	builder := agent.New(agent.ProviderOpenAI, "gpt-4o-mini").
//	    WithAPIKey(apiKey)
func New(provider Provider, model string) *Builder {
	return &Builder{
		provider:   provider,
		model:      model,
		autoMemory: false, // Opt-in for auto-memory
		messages:   []Message{},
	}
}

// NewOpenAI creates a new Builder for OpenAI with the specified model and API key.
// This is the most convenient constructor for OpenAI.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey)
func NewOpenAI(model, apiKey string) *Builder {
	return &Builder{
		provider:   ProviderOpenAI,
		model:      model,
		apiKey:     apiKey,
		autoMemory: false,
		messages:   []Message{},
	}
}

// NewOllama creates a new Builder for Ollama with the specified model.
// By default, it uses http://localhost:11434/v1 as the base URL.
//
// Example:
//
//	builder := agent.NewOllama("qwen2.5:7b")
func NewOllama(model string) *Builder {
	return &Builder{
		provider:   ProviderOllama,
		model:      model,
		baseURL:    "http://localhost:11434/v1",
		autoMemory: false,
		messages:   []Message{},
	}
}

// WithAPIKey sets the API key for the provider.
// Required for OpenAI, not needed for Ollama.
//
// Example:
//
//	builder := agent.New(agent.ProviderOpenAI, "gpt-4o-mini").
//	    WithAPIKey(apiKey)
func (b *Builder) WithAPIKey(apiKey string) *Builder {
	b.apiKey = apiKey
	return b
}

// WithBaseURL sets a custom base URL for the provider.
// Useful for custom endpoints or Ollama installations.
//
// Example:
//
//	builder := agent.NewOllama("qwen2.5:7b").
//	    WithBaseURL("http://192.168.1.100:11434/v1")
func (b *Builder) WithBaseURL(baseURL string) *Builder {
	b.baseURL = baseURL
	return b
}

// WithSystem sets the system prompt that defines the assistant's behavior.
// This is equivalent to adding a system message at the start of the conversation.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithSystem("You are a helpful coding assistant")
func (b *Builder) WithSystem(prompt string) *Builder {
	b.systemPrompt = prompt
	return b
}

// WithMemory enables automatic conversation memory.
// When enabled, all user messages and assistant responses are automatically
// stored in the conversation history for context.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemory()
//	builder.Ask(ctx, "My name is Alice") // Stored in memory
//	builder.Ask(ctx, "What's my name?")  // Model remembers: "Alice"
func (b *Builder) WithMemory() *Builder {
	b.autoMemory = true
	return b
}

// WithMessages sets the conversation history directly.
// Useful for continuing a previous conversation or providing few-shot examples.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMessages([]agent.Message{
//	        agent.User("What is 2+2?"),
//	        agent.Assistant("4"),
//	        agent.User("What is 3+3?"),
//	    })
func (b *Builder) WithMessages(messages []Message) *Builder {
	b.messages = messages
	return b
}

// GetHistory returns a copy of the current conversation history.
// The system prompt is not included in the returned messages.
//
// Example:
//
//	history := builder.GetHistory()
//	fmt.Printf("Conversation has %d messages\n", len(history))
func (b *Builder) GetHistory() []Message {
	// Return a copy to prevent external modification
	history := make([]Message, len(b.messages))
	copy(history, b.messages)
	return history
}

// SetHistory replaces the conversation history with the provided messages.
// This is useful for restoring a previous conversation state.
// The system prompt is preserved.
//
// Example:
//
//	// Save conversation
//	history := builder.GetHistory()
//
//	// Later, restore it
//	builder.SetHistory(history)
func (b *Builder) SetHistory(messages []Message) *Builder {
	b.messages = messages
	return b
}

// Clear resets the conversation history while preserving the system prompt.
// This is useful for starting a fresh conversation with the same configuration.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithSystem("You are a helpful assistant").
//	    WithMemory()
//
//	builder.Ask(ctx, "Hello")
//	builder.Clear() // Start fresh, but keep system prompt
//	builder.Ask(ctx, "Hi") // Model doesn't remember "Hello"
func (b *Builder) Clear() *Builder {
	b.messages = []Message{}
	return b
}

// WithMaxHistory sets the maximum number of messages to keep in history.
// When the limit is reached, old messages are automatically removed (FIFO).
// The system prompt is always preserved and doesn't count toward the limit.
// Set to 0 for unlimited history (default).
//
// Example:
//
//	// Keep only the last 10 messages
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemory().
//	    WithMaxHistory(10)
func (b *Builder) WithMaxHistory(max int) *Builder {
	b.maxHistory = max
	return b
}

// WithTimeout sets the request timeout.
// If set, all API requests will be wrapped with a context timeout.
// Set to 0 for no timeout (default).
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithTimeout(30 * time.Second)
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.timeout = timeout
	return b
}

// WithRetry sets the maximum number of retry attempts.
// When an API request fails, it will be retried up to maxRetries times.
// Set to 0 for no retries (default).
// Use WithRetryDelay to configure delay between retries.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithRetry(3).
//	    WithRetryDelay(2 * time.Second)
func (b *Builder) WithRetry(maxRetries int) *Builder {
	b.maxRetries = maxRetries
	if b.retryDelay == 0 {
		b.retryDelay = time.Second // Default 1s
	}
	return b
}

// WithRetryDelay sets the base delay between retry attempts.
// Default is 1 second.
// Use WithExponentialBackoff for exponential backoff instead of fixed delay.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithRetry(3).
//	    WithRetryDelay(2 * time.Second)
func (b *Builder) WithRetryDelay(delay time.Duration) *Builder {
	b.retryDelay = delay
	return b
}

// WithExponentialBackoff enables exponential backoff for retries.
// Delay doubles after each retry: 1s, 2s, 4s, 8s, etc.
// Must be used with WithRetry.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithRetry(5).
//	    WithExponentialBackoff()
func (b *Builder) WithExponentialBackoff() *Builder {
	b.useExpBackoff = true
	if b.retryDelay == 0 {
		b.retryDelay = time.Second // Default 1s base
	}
	return b
}

// WithTemperature sets the sampling temperature (0-2).
// Higher values (e.g., 1.0+) make output more random and creative.
// Lower values (e.g., 0.2) make output more focused and deterministic.
// Default is typically 1.0.
//
// Example:
//
//	// Creative writing
//	builder.WithTemperature(1.5)
//
//	// Factual answers
//	builder.WithTemperature(0.2)
func (b *Builder) WithTemperature(temperature float64) *Builder {
	b.temperature = &temperature
	return b
}

// WithTopP sets nucleus sampling probability (0-1).
// The model considers tokens with top_p probability mass.
// Lower values make output more focused. Use either temperature OR top_p, not both.
// Default is typically 1.0.
//
// Example:
//
//	builder.WithTopP(0.9)
func (b *Builder) WithTopP(topP float64) *Builder {
	b.topP = &topP
	return b
}

// WithMaxTokens sets the maximum number of tokens to generate.
// Useful for controlling response length and costs.
//
// Example:
//
//	// Short responses only
//	builder.WithMaxTokens(100)
func (b *Builder) WithMaxTokens(maxTokens int64) *Builder {
	b.maxTokens = &maxTokens
	return b
}

// WithPresencePenalty penalizes tokens based on whether they appear in the text so far (-2.0 to 2.0).
// Positive values encourage the model to talk about new topics.
// Default is 0.
//
// Example:
//
//	// Encourage diversity
//	builder.WithPresencePenalty(0.6)
func (b *Builder) WithPresencePenalty(penalty float64) *Builder {
	b.presencePenalty = &penalty
	return b
}

// WithFrequencyPenalty penalizes tokens based on their frequency in the text so far (-2.0 to 2.0).
// Positive values reduce repetition.
// Default is 0.
//
// Example:
//
//	// Reduce repetition
//	builder.WithFrequencyPenalty(0.5)
func (b *Builder) WithFrequencyPenalty(penalty float64) *Builder {
	b.frequencyPenalty = &penalty
	return b
}

// WithSeed sets a seed for deterministic sampling.
// When set, the model will attempt to make repeat requests with the same parameters
// return the same result. This is useful for reproducible testing.
//
// Example:
//
//	// Reproducible outputs
//	builder.WithSeed(42)
func (b *Builder) WithSeed(seed int64) *Builder {
	b.seed = &seed
	return b
}

// WithLogprobs enables returning log probability information for output tokens.
// This is useful for understanding the model's confidence in its predictions.
//
// Example:
//
//	builder.WithLogprobs(true).WithTopLogprobs(5)
func (b *Builder) WithLogprobs(enable bool) *Builder {
	b.logprobs = &enable
	return b
}

// WithTopLogprobs sets the number of most likely tokens to return at each position (0-20).
// Requires WithLogprobs(true) to be set.
//
// Example:
//
//	builder.WithLogprobs(true).WithTopLogprobs(5)
func (b *Builder) WithTopLogprobs(n int64) *Builder {
	b.topLogprobs = &n
	return b
}

// WithMultipleChoices generates N different completion choices.
// Use AskMultiple() to get all choices, or Ask() to get just the first one.
//
// Example:
//
//	// Generate 3 different responses
//	builder.WithMultipleChoices(3)
func (b *Builder) WithMultipleChoices(n int64) *Builder {
	b.n = &n
	return b
}

// OnStream sets a callback function to receive streaming content chunks.
// Use with Stream() method for real-time response streaming.
//
// Example:
//
//	builder.OnStream(func(content string) {
//	    fmt.Print(content)
//	})
func (b *Builder) OnStream(callback func(string)) *Builder {
	b.onStream = callback
	return b
}

// OnToolCall sets a callback for when a tool call is detected during streaming.
//
// Example:
//
//	builder.OnToolCall(func(tool openai.FinishedChatCompletionToolCall) {
//	    fmt.Printf("Tool called: %s\n", tool.Function.Name)
//	})
func (b *Builder) OnToolCall(callback func(openai.FinishedChatCompletionToolCall)) *Builder {
	b.onToolCall = callback
	return b
}

// OnRefusal sets a callback for when the model refuses to respond.
//
// Example:
//
//	builder.OnRefusal(func(refusal string) {
//	    fmt.Printf("Model refused: %s\n", refusal)
//	})
func (b *Builder) OnRefusal(callback func(string)) *Builder {
	b.onRefusal = callback
	return b
}

// WithTool adds a tool that the model can call.
// Tools allow the model to execute functions and use the results.
//
// Example:
//
//	tool := agent.NewTool("get_weather", "Get weather for a location").
//	    AddParameter("location", "string", "City name", true).
//	    WithHandler(func(args string) (string, error) {
//	        return "Sunny, 25°C", nil
//	    })
//	builder.WithTool(tool)
func (b *Builder) WithTool(tool *Tool) *Builder {
	b.tools = append(b.tools, tool)
	return b
}

// WithTools adds multiple tools at once.
//
// Example:
//
//	builder.WithTools(weatherTool, calculatorTool, searchTool)
func (b *Builder) WithTools(tools ...*Tool) *Builder {
	b.tools = append(b.tools, tools...)
	return b
}

// WithAutoExecute enables automatic execution of tool calls.
// When enabled, the builder will automatically call tool handlers and
// continue the conversation with the results.
//
// Example:
//
//	builder.WithTool(weatherTool).
//	    WithAutoExecute(true).
//	    Ask(ctx, "What's the weather in Paris?")
//	// Automatically calls weatherTool and returns final answer
func (b *Builder) WithAutoExecute(enable bool) *Builder {
	b.autoExecute = enable
	if b.maxToolRounds == 0 {
		b.maxToolRounds = 5 // Default max rounds
	}
	return b
}

// WithMaxToolRounds sets the maximum number of tool execution rounds.
// Prevents infinite loops. Default is 5.
//
// Example:
//
//	builder.WithAutoExecute(true).WithMaxToolRounds(3)
func (b *Builder) WithMaxToolRounds(max int) *Builder {
	b.maxToolRounds = max
	return b
}

// WithJSONMode enables JSON object response format.
// This is an older method of generating JSON responses.
// The model will return valid JSON, but you need to instruct it
// in your system or user message to generate JSON.
//
// Example:
//
//	builder.WithJSONMode().
//	    WithSystem("Return your response as JSON").
//	    Ask(ctx, "Get weather for Paris")
func (b *Builder) WithJSONMode() *Builder {
	b.responseFormat = &openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONObject: &shared.ResponseFormatJSONObjectParam{
			Type: constant.JSONObject("json_object"),
		},
	}
	return b
}

// WithJSONSchema enables structured JSON output with a schema.
// The model will always follow the exact schema defined.
// This is the recommended way to get structured outputs.
//
// Parameters:
//   - name: Schema name (a-z, A-Z, 0-9, underscores, dashes, max 64 chars)
//   - description: What the response format is for
//   - schema: JSON Schema object defining the structure
//   - strict: If true, enables strict schema adherence
//
// Example:
//
//	schema := map[string]interface{}{
//	    "type": "object",
//	    "properties": map[string]interface{}{
//	        "temperature": map[string]interface{}{"type": "number"},
//	        "condition": map[string]interface{}{"type": "string"},
//	    },
//	    "required": []string{"temperature", "condition"},
//	}
//	builder.WithJSONSchema("weather_response", "Weather information", schema, true)
func (b *Builder) WithJSONSchema(name, description string, schema interface{}, strict bool) *Builder {
	b.responseFormat = &openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
			Type: constant.JSONSchema("json_schema"),
			JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        name,
				Description: openai.String(description),
				Schema:      schema,
				Strict:      openai.Bool(strict),
			},
		},
	}
	return b
}

// WithResponseFormat sets a custom response format.
// Use WithJSONMode() or WithJSONSchema() for convenience methods.
//
// Example:
//
//	format := &openai.ChatCompletionNewParamsResponseFormatUnion{
//	    OfText: &openai.ResponseFormatTextParam{},
//	}
//	builder.WithResponseFormat(format)
func (b *Builder) WithResponseFormat(format *openai.ChatCompletionNewParamsResponseFormatUnion) *Builder {
	b.responseFormat = format
	return b
}

// Ask sends a message and returns the response as a string.
// This is the simplest method for getting a response.
// If tools are registered and autoExecute is enabled, automatically handles tool calls.
//
// Example:
//
//	response := builder.Ask(ctx, "What is the capital of France?")
//	fmt.Println(response) // "Paris is the capital of France."
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

	// Auto-memory: store this conversation turn
	if b.autoMemory {
		b.addMessage(User(message))
		b.addMessage(Assistant(result))
	}

	return result, nil
}

// askWithToolExecution handles the tool execution loop.
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

		for _, toolCall := range choice.Message.ToolCalls {
			toolStart := time.Now()
			// Find the tool handler
			var handler func(string) (string, error)
			toolName := toolCall.Function.Name

			for _, tool := range b.tools {
				if tool.Name == toolName {
					handler = tool.Handler
					break
				}
			}

			if handler == nil {
				logger.Error(ctx, "No handler found for tool", F("tool_name", toolName))
				return "", fmt.Errorf("no handler found for tool: %s", toolName)
			}

			logger.Debug(ctx, "Executing tool",
				F("tool_name", toolName),
				F("args_length", len(toolCall.Function.Arguments)))

			// Execute the tool
			result, err := handler(toolCall.Function.Arguments)
			toolDuration := time.Since(toolStart)

			if err != nil {
				logger.Error(ctx, "Tool execution failed",
					F("tool_name", toolName),
					F("error", err.Error()),
					F("duration_ms", toolDuration.Milliseconds()))
				return "", fmt.Errorf("tool execution failed: %w", err)
			}

			logger.Debug(ctx, "Tool execution succeeded",
				F("tool_name", toolName),
				F("result_length", len(result)),
				F("duration_ms", toolDuration.Milliseconds()))

			// Add tool result using the helper function
			messages = append(messages, openai.ToolMessage(result, toolCall.ID))
		}
	}

	logger.Warn(ctx, "Max tool rounds exceeded", F("max_rounds", b.maxToolRounds))
	return "", fmt.Errorf("max tool rounds (%d) exceeded", b.maxToolRounds)
} // AskMultiple sends a message and returns multiple completion choices.
// Use WithMultipleChoices(n) to set the number of choices.
//
// Example:
//
//	choices, err := builder.WithMultipleChoices(3).
//	    AskMultiple(ctx, "Write a haiku about coding")
//	for i, choice := range choices {
//	    fmt.Printf("Choice %d: %s\n", i+1, choice)
//	}
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

// Stream sends a message and streams the response using the registered callbacks.
// Returns the complete response text after streaming finishes.
// Uses ALL ChatCompletionAccumulator features: content, tool calls, refusals, usage.
//
// Example:
//
//	response, err := builder.OnStream(func(content string) {
//	    fmt.Print(content)
//	}).Stream(ctx, "Tell me a story")
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

	// Auto-memory: store conversation
	if b.autoMemory && fullContent != "" {
		b.addMessage(User(message))
		b.addMessage(Assistant(fullContent))
	}

	return fullContent, nil
}

// StreamPrint is a convenience method that streams and prints the response to stdout.
// Returns the complete response text.
//
// Example:
//
//	response, err := builder.StreamPrint(ctx, "Tell me a joke")
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

// buildMessages constructs the full message array for the request.
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

// buildParams builds ChatCompletionNewParams with all configured options.
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

// executeWithRetry wraps an operation with retry logic and timeout handling
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

// isRetryable checks if an error is retryable
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

// calculateRetryDelay calculates the delay before next retry
func (b *Builder) calculateRetryDelay(attempt int) time.Duration {
	if b.useExpBackoff {
		// Exponential backoff: delay * 2^attempt
		return b.retryDelay * (1 << uint(attempt))
	}
	// Fixed delay
	return b.retryDelay
}

// addMessage adds a message to the conversation history.
func (b *Builder) addMessage(message Message) {
	b.messages = append(b.messages, message)

	// Auto-truncate if maxHistory is set and exceeded
	if b.maxHistory > 0 && len(b.messages) > b.maxHistory {
		// Remove oldest messages to stay within limit (FIFO)
		excess := len(b.messages) - b.maxHistory
		b.messages = b.messages[excess:]
	}
}

// WithCache sets a custom cache implementation for response caching.
//
// Example:
//
//	cache := agent.NewMemoryCache(1000, 5*time.Minute)
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithCache(cache)
func (b *Builder) WithCache(cache Cache) *Builder {
	b.cache = cache
	b.cacheEnabled = true
	return b
}

// WithMemoryCache enables in-memory caching with LRU eviction.
//
// Parameters:
//   - maxSize: Maximum number of cached responses (default: 1000)
//   - defaultTTL: Default time-to-live for cache entries (default: 5 minutes)
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemoryCache(500, 10*time.Minute)
func (b *Builder) WithMemoryCache(maxSize int, defaultTTL time.Duration) *Builder {
	b.cache = NewMemoryCache(maxSize, defaultTTL)
	b.cacheEnabled = true
	return b
}

// WithRedisCache enables Redis-based caching with simple configuration.
//
// Parameters:
//   - addr: Redis server address (e.g., "localhost:6379")
//   - password: Redis password (use "" if no password)
//   - db: Redis database number (0-15)
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithRedisCache("localhost:6379", "", 0)
func (b *Builder) WithRedisCache(addr, password string, db int) *Builder {
	cache, err := NewRedisCache(addr, password, db, 5*time.Minute)
	if err != nil {
		// Log error but don't fail - fall back to no caching
		return b
	}
	b.cache = cache
	b.cacheEnabled = true
	return b
}

// WithRedisCacheOptions enables Redis-based caching with advanced configuration.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithRedisCacheOptions(&agent.RedisCacheOptions{
//	        Addrs:      []string{"localhost:6379"},
//	        Password:   "",
//	        DB:         0,
//	        PoolSize:   10,
//	        KeyPrefix:  "my-app",
//	        DefaultTTL: 10 * time.Minute,
//	    })
func (b *Builder) WithRedisCacheOptions(opts *RedisCacheOptions) *Builder {
	cache, err := NewRedisCacheWithOptions(opts)
	if err != nil {
		// Log error but don't fail - fall back to no caching
		return b
	}
	b.cache = cache
	b.cacheEnabled = true
	return b
}

// WithCacheTTL sets the TTL for the next cached response.
// If not set, the cache's default TTL is used.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemoryCache(1000, 5*time.Minute).
//	    WithCacheTTL(1*time.Hour)
func (b *Builder) WithCacheTTL(ttl time.Duration) *Builder {
	b.cacheTTL = ttl
	return b
}

// DisableCache disables caching for this builder.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemoryCache(1000, 5*time.Minute).
//	    DisableCache() // Temporarily disable
func (b *Builder) DisableCache() *Builder {
	b.cacheEnabled = false
	return b
}

// EnableCache enables caching for this builder (if cache is set).
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemoryCache(1000, 5*time.Minute).
//	    DisableCache().
//	    EnableCache() // Re-enable
func (b *Builder) EnableCache() *Builder {
	if b.cache != nil {
		b.cacheEnabled = true
	}
	return b
}

// GetCacheStats returns cache statistics if caching is enabled.
//
// Example:
//
//	stats := builder.GetCacheStats()
//	fmt.Printf("Hits: %d, Misses: %d, Hit Rate: %.2f%%\n",
//	    stats.Hits, stats.Misses,
//	    float64(stats.Hits)/(float64(stats.Hits+stats.Misses))*100)
func (b *Builder) GetCacheStats() CacheStats {
	logger := b.getLogger()
	if b.cache != nil {
		stats := b.cache.Stats()
		hitRate := 0.0
		if stats.Hits+stats.Misses > 0 {
			hitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
		}
		logger.Debug(context.Background(), "Cache stats retrieved",
			F("hits", stats.Hits),
			F("misses", stats.Misses),
			F("size", stats.Size),
			F("hit_rate", hitRate))
		return stats
	}
	logger.Debug(context.Background(), "No cache configured")
	return CacheStats{}
}

// ClearCache clears all cached responses.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemoryCache(1000, 5*time.Minute)
//
//	// ... use builder ...
//
//	builder.ClearCache(ctx) // Clear all cached responses
func (b *Builder) ClearCache(ctx context.Context) error {
	logger := b.getLogger()
	if b.cache != nil {
		logger.Info(ctx, "Clearing cache")
		err := b.cache.Clear(ctx)
		if err != nil {
			logger.Error(ctx, "Failed to clear cache", F("error", err.Error()))
			return err
		}
		logger.Info(ctx, "Cache cleared successfully")
		return nil
	}
	logger.Debug(ctx, "No cache to clear")
	return nil
}

// ===== Logging Methods =====

// WithLogger sets a custom logger for observability and debugging.
// By default, a NoopLogger is used (no logging overhead).
//
// The logger can be any implementation of the Logger interface, making it
// compatible with popular logging libraries (slog, zap, logrus, etc.)
//
// Example with custom logger:
//
//	type MyLogger struct{}
//	func (l *MyLogger) Debug(ctx context.Context, msg string, fields ...Field) { ... }
//	func (l *MyLogger) Info(ctx context.Context, msg string, fields ...Field) { ... }
//	func (l *MyLogger) Warn(ctx context.Context, msg string, fields ...Field) { ... }
//	func (l *MyLogger) Error(ctx context.Context, msg string, fields ...Field) { ... }
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithLogger(&MyLogger{})
//
// Example with slog (Go 1.21+):
//
//	import "log/slog"
//	logger := slog.Default()
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithLogger(agent.NewSlogAdapter(logger))
func (b *Builder) WithLogger(logger Logger) *Builder {
	b.logger = logger
	return b
}

// WithDebugLogging enables debug-level logging using the standard library logger.
// This is useful for development and troubleshooting.
//
// Debug logging includes:
//   - Request details (model, message length, cache status)
//   - Cache hits/misses with keys
//   - Tool execution details
//   - RAG retrieval information
//   - Retry attempts and delays
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithDebugLogging()
//
//	// Output example:
//	// [2025-01-15 10:30:45.123] DEBUG: Starting request | model=gpt-4o-mini message_length=50
//	// [2025-01-15 10:30:45.124] DEBUG: Cache miss | cache_key=abc123
//	// [2025-01-15 10:30:46.456] INFO: Request completed | duration_ms=1332 tokens_prompt=12
func (b *Builder) WithDebugLogging() *Builder {
	b.logger = NewStdLogger(LogLevelDebug)
	return b
}

// WithInfoLogging enables info-level logging using the standard library logger.
// This is recommended for production use.
//
// Info logging includes:
//   - Request completion with duration and token usage
//   - Cache hits (but not detailed cache misses)
//   - Tool execution results
//   - Warnings and errors
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithInfoLogging()
//
//	// Output example:
//	// [2025-01-15 10:30:46.456] INFO: Request completed | duration_ms=1332 tokens_prompt=12 tokens_completion=45
//	// [2025-01-15 10:30:47.789] INFO: Cache hit | cache_key=abc123
func (b *Builder) WithInfoLogging() *Builder {
	b.logger = NewStdLogger(LogLevelInfo)
	return b
}

// getLogger returns the configured logger or NoopLogger if none is set.
// This ensures zero overhead when logging is not enabled.
func (b *Builder) getLogger() Logger {
	if b.logger == nil {
		return &NoopLogger{}
	}
	return b.logger
}
