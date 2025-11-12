package agent

import (
	"time"
	_ "unsafe" // For go:linkname

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/openai/openai-go/v3/shared/constant"
	"github.com/taipm/go-deep-agent/agent/memory"
)

// Link to tools.SetLogFunc using go:linkname to avoid import cycle
//
//go:linkname toolsSetLogFunc github.com/taipm/go-deep-agent/agent/tools.SetLogFunc
func toolsSetLogFunc(fn func(level, msg string, fields map[string]interface{}))

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
	tools         []*Tool                                          // Registered tools
	autoExecute   bool                                             // If true, automatically execute tool calls
	maxToolRounds int                                              // Maximum number of tool execution rounds (default 5)
	toolChoice    *openai.ChatCompletionToolChoiceOptionUnionParam // Tool choice control ("auto", "required", "none")

	// Tool orchestration (parallel execution)
	enableParallel bool          // Enable parallel tool execution (default: false)
	maxWorkers     int           // Max concurrent tool workers (default: 10)
	toolTimeout    time.Duration // Timeout per tool (default: 30s)

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

	// Enhanced debug mode
	debugConfig DebugConfig  // Debug configuration
	debugLogger *debugLogger // Debug logger instance

	// Memory (Hierarchical memory system)
	memory        *memory.Memory // Hierarchical memory system (default: enabled)
	memoryEnabled bool           // Whether memory system is enabled

	// Persona (Behavioral configuration)
	persona *Persona // Active persona configuration

	// Few-shot Learning
	fewshotConfig *FewShotConfig // Few-shot examples configuration

	// ReAct (Reasoning + Acting)
	reactConfig *ReActConfig // ReAct pattern configuration

	// Rate Limiting
	rateLimiter      RateLimiter     // Rate limiter instance
	rateLimitConfig  RateLimitConfig // Rate limit configuration
	rateLimitEnabled bool            // Whether rate limiting is enabled
	rateLimitKey     string          // Key for per-key rate limiting

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
		provider:      provider,
		model:         model,
		autoMemory:    false, // Opt-in for auto-memory
		messages:      []Message{},
		memory:        memory.New(), // Auto-enable hierarchical memory
		memoryEnabled: true,
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
		provider:      ProviderOpenAI,
		model:         model,
		apiKey:        apiKey,
		autoMemory:    false,
		messages:      []Message{},
		memory:        memory.New(),
		memoryEnabled: true,
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
		provider:      ProviderOllama,
		model:         model,
		baseURL:       "http://localhost:11434/v1",
		autoMemory:    false,
		messages:      []Message{},
		memory:        memory.New(),
		memoryEnabled: true,
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

// WithSemanticMemory enables semantic memory for fact storage.
// Semantic memory stores and retrieves factual information by category.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithSemanticMemory()

// WithMessages sets the conversation history directly.
// Useful for continuing a previous conversation or providing few-shot examples.

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

// WithRetryDelay sets the base delay between retry attempts.
// Default is 1 second.
// Use WithExponentialBackoff for exponential backoff instead of fixed delay.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithRetry(3).
//	    WithRetryDelay(2 * time.Second)

// WithExponentialBackoff enables exponential backoff for retries.
// Delay doubles after each retry: 1s, 2s, 4s, 8s, etc.
// Must be used with WithRetry.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithRetry(5).
//	    WithExponentialBackoff()

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

// WithTopP sets nucleus sampling probability (0-1).
// The model considers tokens with top_p probability mass.
// Lower values make output more focused. Use either temperature OR top_p, not both.
// Default is typically 1.0.
//
// Example:
//
//	builder.WithTopP(0.9)

// WithMaxTokens sets the maximum number of tokens to generate.
// Useful for controlling response length and costs.
//
// Example:
//
//	// Short responses only
//	builder.WithMaxTokens(100)

// WithPresencePenalty penalizes tokens based on whether they appear in the text so far (-2.0 to 2.0).
// Positive values encourage the model to talk about new topics.
// Default is 0.
//
// Example:
//
//	// Encourage diversity
//	builder.WithPresencePenalty(0.6)

// WithFrequencyPenalty penalizes tokens based on their frequency in the text so far (-2.0 to 2.0).
// Positive values reduce repetition.
// Default is 0.
//
// Example:
//
//	// Reduce repetition
//	builder.WithFrequencyPenalty(0.5)

// WithSeed sets a seed for deterministic sampling.
// When set, the model will attempt to make repeat requests with the same parameters
// return the same result. This is useful for reproducible testing.
//
// Example:
//
//	// Reproducible outputs
//	builder.WithSeed(42)

// WithLogprobs enables returning log probability information for output tokens.
// This is useful for understanding the model's confidence in its predictions.
//
// Example:
//
//	builder.WithLogprobs(true).WithTopLogprobs(5)

// WithTopLogprobs sets the number of most likely tokens to return at each position (0-20).
// Requires WithLogprobs(true) to be set.
//
// Example:
//
//	builder.WithLogprobs(true).WithTopLogprobs(5)

// WithMultipleChoices generates N different completion choices.
// Use AskMultiple() to get all choices, or Ask() to get just the first one.
//
// Example:
//
//	// Generate 3 different responses
//	builder.WithMultipleChoices(3)

// OnStream sets a callback function to receive streaming content chunks.
// Use with Stream() method for real-time response streaming.
//
// Example:
//
//	builder.OnStream(func(content string) {
//	    fmt.Print(content)
//	})

// OnToolCall sets a callback for when a tool call is detected during streaming.
//
// Example:
//
//	builder.OnToolCall(func(tool openai.FinishedChatCompletionToolCall) {
//	    fmt.Printf("Tool called: %s\n", tool.Function.Name)
//	})

// OnRefusal sets a callback for when the model refuses to respond.
//
// Example:
//
//	builder.OnRefusal(func(refusal string) {
//	    fmt.Printf("Model refused: %s\n", refusal)
//	})

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

// WithTools adds multiple tools at once.
//
// Example:
//
//	builder.WithTools(weatherTool, calculatorTool, searchTool)

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

// WithMaxToolRounds sets the maximum number of tool execution rounds.
// Prevents infinite loops. Default is 5.
//
// Example:
//
//	builder.WithAutoExecute(true).WithMaxToolRounds(3)

// WithParallelTools enables parallel execution of independent tools.
// Tools without dependencies run concurrently, respecting worker pool limits.
// This can significantly speed up multi-tool executions (3x+ faster).
//
// Example:
//
//	builder.WithTools(tool1, tool2, tool3).
//	    WithAutoExecute(true).
//	    WithParallelTools(true).  // Execute tools in parallel
//	    Ask(ctx, "Analyze this data")

// WithMaxWorkers sets the maximum number of concurrent tool workers.
// Only applies when parallel tool execution is enabled.
// Default is 10 workers.
//
// Example:
//
//	builder.WithParallelTools(true).
//	    WithMaxWorkers(5)  // Max 5 tools running concurrently

// WithToolTimeout sets the timeout for individual tool executions.
// Only applies when parallel tool execution is enabled.
// Default is 30 seconds.
//
// Example:
//
//	builder.WithParallelTools(true).
//	    WithToolTimeout(60 * time.Second)  // 60s timeout per tool

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

// askWithToolExecution handles the tool execution loop.
// Use WithMultipleChoices(n) to set the number of choices.
//
// Example:
//
//	choices, err := builder.WithMultipleChoices(3).
//	    AskMultiple(ctx, "Write a haiku about coding")
//	for i, choice := range choices {
//	    fmt.Printf("Choice %d: %s\n", i+1, choice)
//	}

// Stream sends a message and streams the response using the registered callbacks.
// Returns the complete response text after streaming finishes.
// Uses ALL ChatCompletionAccumulator features: content, tool calls, refusals, usage.
//
// Example:
//
//	response, err := builder.OnStream(func(content string) {
//	    fmt.Print(content)
//	}).Stream(ctx, "Tell me a story")

// StreamPrint is a convenience method that streams and prints the response to stdout.
// Returns the complete response text.
//
// Example:
//
//	response, err := builder.StreamPrint(ctx, "Tell me a joke")

// buildMessages constructs the full message array for the request.

// buildParams builds ChatCompletionNewParams with all configured options.

// executeWithRetry wraps an operation with retry logic and timeout handling

// isRetryable checks if an error is retryable

// calculateRetryDelay calculates the delay before next retry

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

// WithCacheTTL sets the TTL for the next cached response.
// If not set, the cache's default TTL is used.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemoryCache(1000, 5*time.Minute).
//	    WithCacheTTL(1*time.Hour)

// DisableCache disables caching for this builder.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemoryCache(1000, 5*time.Minute).
//	    DisableCache() // Temporarily disable

// EnableCache enables caching for this builder (if cache is set).
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemoryCache(1000, 5*time.Minute).
//	    DisableCache().
//	    EnableCache() // Re-enable

// GetCacheStats returns cache statistics if caching is enabled.
//
// Example:
//
//	stats := builder.GetCacheStats()
//	fmt.Printf("Hits: %d, Misses: %d, Hit Rate: %.2f%%\n",
//	    stats.Hits, stats.Misses,
//	    float64(stats.Hits)/(float64(stats.Hits+stats.Misses))*100)

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

// getToolNames extracts the names of all registered tools as a string slice.
// Used internally for building enum schemas in native ReAct mode.
// Returns empty slice if no tools registered.
//
// Example:
//
//	builder.tools = []*Tool{
//	    NewMathTool(),
//	    NewDateTimeTool(),
//	}
//	names := builder.getToolNames()  // Returns: ["math", "datetime"]
func (b *Builder) getToolNames() []string {
	if len(b.tools) == 0 {
		return []string{}
	}

	names := make([]string, len(b.tools))
	for i, tool := range b.tools {
		names[i] = tool.Name
	}
	return names
}
