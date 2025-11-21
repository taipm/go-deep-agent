package agent

import (
	"fmt"
	"time"

	"github.com/taipm/go-deep-agent/agent/memory"
)

// WithAgentConfig applies a complete AgentConfig to the builder
// This allows loading configuration from YAML files
//
// Example:
//
//	config, _ := agent.LoadAgentConfig("config.yaml")
//	builder := agent.NewOpenAI("", apiKey).WithAgentConfig(config)
func (b *Builder) WithAgentConfig(config *AgentConfig) *Builder {
	// Model settings
	if config.Model != "" {
		b.model = config.Model
	}
	b.WithTemperature(config.Temperature)
	b.WithMaxTokens(int64(config.MaxTokens))
	b.WithTopP(config.TopP)

	// System prompt (backward compatibility)
	if config.SystemPrompt != "" {
		b.WithSystem(config.SystemPrompt)
	}

	// Memory settings
	if b.memoryEnabled && b.memory != nil {
		memConfig := memory.MemoryConfig{
			WorkingCapacity:   config.Memory.WorkingCapacity,
			EpisodicEnabled:   config.Memory.EpisodicEnabled,
			EpisodicThreshold: config.Memory.EpisodicThreshold,
			SemanticEnabled:   config.Memory.SemanticEnabled,
			AutoCompress:      config.Memory.AutoCompress,
		}
		b.memory.SetConfig(memConfig)
	}

	// Retry settings
	if config.Retry.MaxAttempts > 0 {
		b.WithRetry(config.Retry.MaxAttempts)
	}
	if config.Retry.Timeout > 0 {
		b.WithTimeout(config.Retry.Timeout)
	}
	if config.Retry.ExponentialBackoff {
		b.WithExponentialBackoff()
	}

	// Tools settings
	if config.Tools.ParallelExecution {
		b.WithParallelTools(true)
		if config.Tools.MaxWorkers > 0 {
			b.WithMaxWorkers(config.Tools.MaxWorkers)
		}
		if config.Tools.Timeout > 0 {
			b.WithToolTimeout(config.Tools.Timeout)
		}
	}

	// Rate limiting settings
	if config.RateLimit.Enabled {
		b.WithRateLimitConfig(config.RateLimit)
	}

	return b
}

// ToAgentConfig exports the current builder state as an AgentConfig
// This allows saving the current configuration to a YAML file
//
// Example:
//
//	config := builder.ToAgentConfig()
//	agent.SaveAgentConfig(config, "exported_config.yaml")
func (b *Builder) ToAgentConfig() *AgentConfig {
	config := DefaultAgentConfig()

	// Model settings
	config.Model = b.model
	if b.temperature != nil {
		config.Temperature = *b.temperature
	}
	if b.maxTokens != nil {
		config.MaxTokens = int(*b.maxTokens)
	}
	if b.topP != nil {
		config.TopP = *b.topP
	}

	// System prompt
	config.SystemPrompt = b.systemPrompt

	// Memory settings
	if b.memoryEnabled && b.memory != nil {
		memConfig := b.memory.GetConfig()
		config.Memory.WorkingCapacity = memConfig.WorkingCapacity
		config.Memory.EpisodicEnabled = memConfig.EpisodicEnabled
		config.Memory.EpisodicThreshold = memConfig.EpisodicThreshold
		config.Memory.SemanticEnabled = memConfig.SemanticEnabled
		config.Memory.AutoCompress = memConfig.AutoCompress
	}

	// Retry settings
	config.Retry.MaxAttempts = b.maxRetries
	config.Retry.Timeout = b.timeout
	config.Retry.ExponentialBackoff = b.useExpBackoff
	if b.retryDelay > 0 {
		config.Retry.InitialDelay = b.retryDelay
	}

	// Tools settings
	config.Tools.ParallelExecution = b.enableParallel
	config.Tools.MaxWorkers = b.maxWorkers
	config.Tools.Timeout = b.toolTimeout

	// Rate limiting settings
	if b.rateLimitEnabled {
		config.RateLimit = b.rateLimitConfig
	} else {
		config.RateLimit = DefaultRateLimitConfig()
	}

	return config
}

// WithFullConfig applies a complete FullConfig (persona + settings) to the builder.
// This is the Phase 3 method that combines persona and technical settings.
//
// Priority: Settings override Persona's TechnicalConfig
//
// Example:
//
//	config, _ := agent.LoadFullConfig("configs/agent.yaml")
//	builder := agent.NewOpenAI("", apiKey).WithFullConfig(config)
func (b *Builder) WithFullConfig(config *FullConfig) *Builder {
	if config == nil {
		return b
	}

	// Apply persona first (generates system prompt + applies persona's technical config)
	if config.Persona != nil {
		b.WithPersona(config.Persona)
	}

	// Apply settings second (overrides persona's technical config)
	if config.Settings != nil {
		b.WithSettings(config.Settings)
	}

	return b
}

// WithSettings applies AgentSettings to the builder.
// This allows applying just technical settings without a persona.
//
// Example:
//
//	settings, _ := agent.LoadSettings("configs/production.yaml")
//	builder := agent.NewOpenAI("gpt-4", apiKey).WithSettings(settings)
func (b *Builder) WithSettings(settings *AgentSettings) *Builder {
	if settings == nil {
		return b
	}

	// Model settings
	if settings.Model != "" {
		b.model = settings.Model
	}
	if settings.Temperature > 0 {
		b.WithTemperature(settings.Temperature)
	}
	if settings.MaxTokens > 0 {
		b.WithMaxTokens(int64(settings.MaxTokens))
	}
	if settings.TopP > 0 {
		b.WithTopP(settings.TopP)
	}
	if settings.Timeout > 0 {
		b.WithTimeout(settings.Timeout)
	}

	// Memory settings
	if settings.Memory != nil && b.memoryEnabled && b.memory != nil {
		memConfig := memory.MemoryConfig{
			WorkingCapacity:   settings.Memory.WorkingCapacity,
			EpisodicEnabled:   settings.Memory.EpisodicEnabled,
			EpisodicThreshold: settings.Memory.EpisodicThreshold,
			SemanticEnabled:   settings.Memory.SemanticEnabled,
			AutoCompress:      settings.Memory.AutoCompress,
		}
		b.memory.SetConfig(memConfig)
	}

	// Retry settings
	if settings.Retry != nil {
		if settings.Retry.MaxAttempts > 0 {
			b.WithRetry(settings.Retry.MaxAttempts)
		}
		if settings.Retry.Timeout > 0 {
			b.WithTimeout(settings.Retry.Timeout)
		}
		if settings.Retry.ExponentialBackoff {
			b.WithExponentialBackoff()
		}
		if settings.Retry.BackoffMultiplier > 0 {
			// Store for exponential backoff calculation
			// Note: Builder doesn't have direct field for this yet
		}
	}

	// Tools settings
	if settings.Tools != nil {
		if settings.Tools.ParallelExecution {
			b.WithParallelTools(true)
		}
		if settings.Tools.MaxWorkers > 0 {
			b.WithMaxWorkers(settings.Tools.MaxWorkers)
		}
		if settings.Tools.Timeout > 0 {
			b.WithToolTimeout(settings.Tools.Timeout)
		}
	}

	return b
}

// GetFullConfig exports the current builder state as a FullConfig.
// This includes both persona and settings.
//
// Example:
//
//	config := builder.GetFullConfig()
//	agent.SaveFullConfig(config, "exported_config.yaml")
func (b *Builder) GetFullConfig() *FullConfig {
	config := &FullConfig{
		Persona:  b.persona,
		Settings: b.ToAgentSettings(),
	}
	return config
}

// ToAgentSettings exports the current builder state as AgentSettings.
// This extracts just the technical settings without persona.
//
// Example:
//
//	settings := builder.ToAgentSettings()
//	agent.SaveSettings(settings, "settings.yaml")
func (b *Builder) ToAgentSettings() *AgentSettings {
	settings := &AgentSettings{}

	// Model settings
	settings.Model = b.model
	if b.temperature != nil {
		settings.Temperature = *b.temperature
	}
	if b.maxTokens != nil {
		settings.MaxTokens = int(*b.maxTokens)
	}
	if b.topP != nil {
		settings.TopP = *b.topP
	}
	settings.Timeout = b.timeout

	// Memory settings
	if b.memoryEnabled && b.memory != nil {
		memConfig := b.memory.GetConfig()
		settings.Memory = &MemorySettings{
			WorkingCapacity:   memConfig.WorkingCapacity,
			EpisodicEnabled:   memConfig.EpisodicEnabled,
			EpisodicThreshold: memConfig.EpisodicThreshold,
			SemanticEnabled:   memConfig.SemanticEnabled,
			AutoCompress:      memConfig.AutoCompress,
		}
	}

	// Retry settings
	if b.maxRetries > 0 || b.timeout > 0 {
		settings.Retry = &RetrySettings{
			MaxAttempts:        b.maxRetries,
			Timeout:            b.timeout,
			ExponentialBackoff: b.useExpBackoff,
			InitialDelay:       b.retryDelay,
		}
	}

	// Tools settings
	if b.enableParallel || b.maxWorkers > 0 {
		settings.Tools = &ToolsSettings{
			ParallelExecution: b.enableParallel,
			MaxWorkers:        b.maxWorkers,
			Timeout:           b.toolTimeout,
		}
	}

	return settings
}

// WithRateLimit enables rate limiting with simple configuration.
// This is the basic method for enabling rate limiting with sensible defaults.
//
// Parameters:
//   - requestsPerSecond: sustained rate of requests allowed (e.g., 10.0 = 10 requests/sec)
//   - burstSize: maximum burst of requests (e.g., 20 = allow 20 requests at once)
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4", apiKey).
//	    WithRateLimit(10.0, 20) // 10 req/s, burst of 20
func (b *Builder) WithRateLimit(requestsPerSecond float64, burstSize int) *Builder {
	b.rateLimitConfig = RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		PerKey:            false,
		KeyTimeout:        5 * time.Minute,
		WaitTimeout:       30 * time.Second,
	}
	b.rateLimitEnabled = true
	return b
}

// WithRateLimitConfig enables rate limiting with advanced configuration.
// Use this for more control over rate limiting behavior (per-key, timeouts, etc.).
//
// Example:
//
//	config := agent.RateLimitConfig{
//	    Enabled:           true,
//	    RequestsPerSecond: 10.0,
//	    BurstSize:         20,
//	    PerKey:            true,  // Enable per-key rate limiting
//	    KeyTimeout:        5 * time.Minute,
//	}
//	builder := agent.NewOpenAI("gpt-4", apiKey).
//	    WithRateLimitConfig(config)
func (b *Builder) WithRateLimitConfig(config RateLimitConfig) *Builder {
	b.rateLimitConfig = config
	b.rateLimitEnabled = config.Enabled
	return b
}

// WithRateLimitKey sets the key for per-key rate limiting.
// This is useful for implementing per-user or per-API-key rate limits.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4", apiKey).
//	    WithRateLimit(10.0, 20).
//	    WithRateLimitKey("user-123") // Different limit per user
func (b *Builder) WithRateLimitKey(key string) *Builder {
	b.rateLimitKey = key
	return b
}

// validateConfiguration checks for invalid or conflicting configuration
// Added in v0.7.9 - comprehensive validation at execution time
//
// This method is called automatically by Ask(), Stream(), and other execution methods
// to catch configuration errors early with clear, actionable error messages.
//
// Validation checks:
//   - Tool choice requirements (toolChoice set without tools)
//   - API key requirements for different providers
//   - Adapter vs client configuration consistency
//   - Invalid parameter ranges
//   - Conflicting settings
//
// Returns nil if configuration is valid, or a detailed error with fixes.
// For comprehensive validation with all errors, use ValidateWithDetails().
func (b *Builder) validateConfiguration() error {
	// Check tool choice requires tools
	if b.toolChoice != nil && len(b.tools) == 0 {
		return ErrToolChoiceRequiresTools
	}

	// Validate provider-specific configuration
	if err := b.validateProviderConfig(); err != nil {
		return err
	}

	// Validate parameter ranges
	if err := b.validateParameterRanges(); err != nil {
		return err
	}

	// Validate conflicting settings
	if err := b.validateConflictingSettings(); err != nil {
		return err
	}

	return nil
}

// validateConfigurationAll returns all validation errors, not just the first one
// This is used internally by ValidateWithDetails() to collect comprehensive errors
func (b *Builder) validateConfigurationAll() []error {
	var errors []error

	// Check tool choice requires tools
	if b.toolChoice != nil && len(b.tools) == 0 {
		errors = append(errors, ErrToolChoiceRequiresTools)
	}

	// Validate provider-specific configuration
	if err := b.validateProviderConfigAll(); err != nil {
		errors = append(errors, err...)
	}

	// Validate parameter ranges
	if errs := b.validateParameterRangesAll(); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	// Validate conflicting settings
	if err := b.validateConflictingSettings(); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// validateProviderConfigAll returns all provider-specific validation errors
func (b *Builder) validateProviderConfigAll() []error {
	var errors []error

	switch b.provider {
	case ProviderOpenAI:
		if b.apiKey == "" && b.adapter == nil {
			errors = append(errors, fmt.Errorf("OpenAI API key is required\n\n"+
				"Fix options:\n"+
				"  1. Set environment variable: export OPENAI_API_KEY=\"sk-...\"\n"+
				"  2. Use constructor: agent.NewOpenAI(\"gpt-4o-mini\", \"sk-...\")\n"+
				"  3. Set API key: .WithAPIKey(\"sk-...\")\n"+
				"  4. Use FromEnv(): agent.FromEnv() for auto-detection\n\n"+
				"Get your key: https://platform.openai.com/api-keys"))
		}
		// Validate API key format
		if b.apiKey != "" && len(b.apiKey) < 10 {
			errors = append(errors, fmt.Errorf("OpenAI API key appears to be invalid (too short)\n\n"+
				"Valid API keys start with 'sk-' and are typically 50+ characters\n"+
				"Get your key: https://platform.openai.com/api-keys"))
		}

	case ProviderGemini:
		if b.apiKey == "" && b.adapter == nil {
			errors = append(errors, fmt.Errorf("Gemini API key is required\n\n"+
				"Fix options:\n"+
				"  1. Set environment variable: export GEMINI_API_KEY=\"AIza...\"\n"+
				"  2. Use constructor: agent.NewGemini(\"gemini-pro\", \"AIza...\")\n"+
				"  3. Set API key: .WithAPIKey(\"AIza...\")\n\n"+
				"Get your key: https://aistudio.google.com/apikey"))
		}
	}

	return errors
}

// validateParameterRangesAll returns all parameter range validation errors
func (b *Builder) validateParameterRangesAll() []error {
	var errors []error

	// Temperature range: 0.0 to 2.0 for OpenAI, 0.0 to 1.0 for Gemini
	if b.temperature != nil {
		temp := *b.temperature
		if temp < 0.0 || temp > 2.0 {
			errors = append(errors, fmt.Errorf("temperature %.1f is out of valid range [0.0, 2.0]\n\n"+
				"Fix:\n"+
				"  • Creative writing: .WithTemperature(1.2) (0.8-1.5)\n"+
				"  • Balanced responses: .WithTemperature(0.7) (0.5-1.0)\n"+
				"  • Factual answers: .WithTemperature(0.2) (0.0-0.4)\n"+
				"  • Deterministic output: .WithTemperature(0.0)", temp))
		}
	}

	// TopP range: 0.0 to 1.0
	if b.topP != nil {
		topP := *b.topP
		if topP < 0.0 || topP > 1.0 {
			errors = append(errors, fmt.Errorf("top_p %.2f is out of valid range [0.0, 1.0]\n\n"+
				"Fix:\n"+
				"  • Focused output: .WithTopP(0.8) or .WithTopP(0.9)\n"+
				"  • Diverse output: .WithTopP(0.95) or .WithTopP(1.0)\n"+
				"  • Note: Use temperature OR top_p, not both", topP))
		}
	}

	// MaxTokens should be reasonable
	if b.maxTokens != nil {
		maxTokens := *b.maxTokens
		if maxTokens < 1 {
			errors = append(errors, fmt.Errorf("max_tokens %d is too small (minimum: 1)\n\n"+
				"Fix:\n"+
				"  • Short responses: .WithMaxTokens(100)\n"+
				"  • Medium responses: .WithMaxTokens(500)\n"+
				"  • Long responses: .WithMaxTokens(2000) or higher", maxTokens))
		}
		if maxTokens > 32768 && b.provider == ProviderOpenAI {
			errors = append(errors, fmt.Errorf("max_tokens %d exceeds OpenAI limit (32768)\n\n"+
				"Fix:\n"+
				"  • Use shorter response: .WithMaxTokens(4000)\n"+
				"  • Split into multiple requests\n"+
				"  • For longer content, consider using 32k models", maxTokens))
		}
	}

	// Penalty ranges: -2.0 to 2.0
	if b.presencePenalty != nil {
		penalty := *b.presencePenalty
		if penalty < -2.0 || penalty > 2.0 {
			errors = append(errors, fmt.Errorf("presence_penalty %.1f is out of valid range [-2.0, 2.0]\n\n"+
				"Fix:\n"+
				"  • Reduce repetition: .WithPresencePenalty(0.6)\n"+
				"  • Encourage new topics: .WithPresencePenalty(1.0)\n"+
				"  • Default behavior: .WithPresencePenalty(0.0)", penalty))
		}
	}

	if b.frequencyPenalty != nil {
		penalty := *b.frequencyPenalty
		if penalty < -2.0 || penalty > 2.0 {
			errors = append(errors, fmt.Errorf("frequency_penalty %.1f is out of valid range [-2.0, 2.0]\n\n"+
				"Fix:\n"+
				"  • Reduce repetition: .WithFrequencyPenalty(0.5)\n"+
				"  • Strong reduction: .WithFrequencyPenalty(1.0)\n"+
				"  • Default behavior: .WithFrequencyPenalty(0.0)", penalty))
		}
	}

	// TopLogprobs range: 0 to 20
	if b.topLogprobs != nil {
		topLogprobs := *b.topLogprobs
		if topLogprobs < 0 || topLogprobs > 20 {
			errors = append(errors, fmt.Errorf("top_logprobs %d is out of valid range [0, 20]\n\n"+
				"Fix:\n"+
				"  • Top 5 tokens: .WithTopLogprobs(5)\n"+
				"  • Top 10 tokens: .WithTopLogprobs(10)\n"+
				"  • Must enable: .WithLogprobs(true)", topLogprobs))
		}
	}

	return errors
}

// validateProviderConfig checks provider-specific requirements
func (b *Builder) validateProviderConfig() error {
	switch b.provider {
	case ProviderOpenAI:
		if b.apiKey == "" && b.adapter == nil {
			return fmt.Errorf("OpenAI API key is required\n\n"+
				"Fix options:\n"+
				"  1. Set environment variable: export OPENAI_API_KEY=\"sk-...\"\n"+
				"  2. Use constructor: agent.NewOpenAI(\"gpt-4o-mini\", \"sk-...\")\n"+
				"  3. Set API key: .WithAPIKey(\"sk-...\")\n"+
				"  4. Use FromEnv(): agent.FromEnv() for auto-detection\n\n"+
				"Get your key: https://platform.openai.com/api-keys")
		}
		// Validate API key format
		if b.apiKey != "" && len(b.apiKey) < 10 {
			return fmt.Errorf("OpenAI API key appears to be invalid (too short)\n\n"+
				"Valid API keys start with 'sk-' and are typically 50+ characters\n"+
				"Get your key: https://platform.openai.com/api-keys")
		}

	case ProviderGemini:
		if b.apiKey == "" && b.adapter == nil {
			return fmt.Errorf("Gemini API key is required\n\n"+
				"Fix options:\n"+
				"  1. Set environment variable: export GEMINI_API_KEY=\"AIza...\"\n"+
				"  2. Use constructor: agent.NewGemini(\"gemini-pro\", \"AIza...\")\n"+
				"  3. Set API key: .WithAPIKey(\"AIza...\")\n\n"+
				"Get your key: https://aistudio.google.com/apikey")
		}

	case ProviderOllama:
		// Ollama doesn't require API key but validate base URL if set
		if b.baseURL != "" && b.baseURL == "" {
			return fmt.Errorf("Ollama base URL is invalid or empty\n\n"+
				"Fix:\n"+
				"  1. Use default: Removes WithBaseURL() call (uses http://localhost:11434/v1)\n"+
				"  2. Set correct URL: .WithBaseURL(\"http://localhost:11434/v1\")\n"+
				"  3. Check Ollama is running: ollama list")
		}
	}

	return nil
}

// validateParameterRanges checks parameter values are within acceptable ranges
func (b *Builder) validateParameterRanges() error {
	// Temperature range: 0.0 to 2.0 for OpenAI, 0.0 to 1.0 for Gemini
	if b.temperature != nil {
		temp := *b.temperature
		if temp < 0.0 || temp > 2.0 {
			return fmt.Errorf("temperature %.1f is out of valid range [0.0, 2.0]\n\n"+
				"Fix:\n"+
				"  • Creative writing: .WithTemperature(1.2) (0.8-1.5)\n"+
				"  • Balanced responses: .WithTemperature(0.7) (0.5-1.0)\n"+
				"  • Factual answers: .WithTemperature(0.2) (0.0-0.4)\n"+
				"  • Deterministic output: .WithTemperature(0.0)", temp)
		}
	}

	// TopP range: 0.0 to 1.0
	if b.topP != nil {
		topP := *b.topP
		if topP < 0.0 || topP > 1.0 {
			return fmt.Errorf("top_p %.2f is out of valid range [0.0, 1.0]\n\n"+
				"Fix:\n"+
				"  • Focused output: .WithTopP(0.8) or .WithTopP(0.9)\n"+
				"  • Diverse output: .WithTopP(0.95) or .WithTopP(1.0)\n"+
				"  • Note: Use temperature OR top_p, not both", topP)
		}
	}

	// MaxTokens should be reasonable
	if b.maxTokens != nil {
		maxTokens := *b.maxTokens
		if maxTokens < 1 {
			return fmt.Errorf("max_tokens %d is too small (minimum: 1)\n\n"+
				"Fix:\n"+
				"  • Short responses: .WithMaxTokens(100)\n"+
				"  • Medium responses: .WithMaxTokens(500)\n"+
				"  • Long responses: .WithMaxTokens(2000) or higher", maxTokens)
		}
		if maxTokens > 32768 && b.provider == ProviderOpenAI {
			return fmt.Errorf("max_tokens %d exceeds OpenAI limit (32768)\n\n"+
				"Fix:\n"+
				"  • Use shorter response: .WithMaxTokens(4000)\n"+
				"  • Split into multiple requests\n"+
				"  • For longer content, consider using 32k models", maxTokens)
		}
	}

	// Penalty ranges: -2.0 to 2.0
	if b.presencePenalty != nil {
		penalty := *b.presencePenalty
		if penalty < -2.0 || penalty > 2.0 {
			return fmt.Errorf("presence_penalty %.1f is out of valid range [-2.0, 2.0]\n\n"+
				"Fix:\n"+
				"  • Reduce repetition: .WithPresencePenalty(0.6)\n"+
				"  • Encourage new topics: .WithPresencePenalty(1.0)\n"+
				"  • Default behavior: .WithPresencePenalty(0.0)", penalty)
		}
	}

	if b.frequencyPenalty != nil {
		penalty := *b.frequencyPenalty
		if penalty < -2.0 || penalty > 2.0 {
			return fmt.Errorf("frequency_penalty %.1f is out of valid range [-2.0, 2.0]\n\n"+
				"Fix:\n"+
				"  • Reduce repetition: .WithFrequencyPenalty(0.5)\n"+
				"  • Strong reduction: .WithFrequencyPenalty(1.0)\n"+
				"  • Default behavior: .WithFrequencyPenalty(0.0)", penalty)
		}
	}

	// TopLogprobs range: 0 to 20
	if b.topLogprobs != nil {
		topLogprobs := *b.topLogprobs
		if topLogprobs < 0 || topLogprobs > 20 {
			return fmt.Errorf("top_logprobs %d is out of valid range [0, 20]\n\n"+
				"Fix:\n"+
				"  • Top 5 tokens: .WithTopLogprobs(5)\n"+
				"  • Top 10 tokens: .WithTopLogprobs(10)\n"+
				"  • Must enable: .WithLogprobs(true)", topLogprobs)
		}
	}

	return nil
}

// validateConflictingSettings checks for mutually exclusive configurations
func (b *Builder) validateConflictingSettings() error {
	// Check for tool choice conflicts with auto-execute
	// Note: This is a simplified check since toolChoice is a complex union type
	if b.toolChoice != nil && b.autoExecute {
		// We'll do a basic check - in practice this would need more sophisticated handling
		// of the OpenAI union type to determine if it's set to "none"
	}

	// Check for multiple response format conflicts
	if b.responseFormat != nil && b.reactConfig != nil && b.reactConfig.Enabled {
		// This is not necessarily an error, but worth warning about
		// ReAct with structured outputs might have unexpected behavior
	}

	return nil
}

// ValidateConfig is a public method for manual validation
// This allows users to validate their configuration before making API calls
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey)
//	if err := builder.ValidateConfig(); err != nil {
//		log.Fatalf("Configuration error: %v", err)
//	}
func (b *Builder) ValidateConfig() error {
	return b.validateConfiguration()
}

// ValidateWithDetails returns detailed validation information
// This provides more comprehensive feedback for debugging
//
// Example:
//
//	details, err := builder.ValidateWithDetails()
//	if err != nil {
//		fmt.Printf("Validation failed: %v\n", err)
//		fmt.Printf("Provider: %s\n", details.Provider)
//		fmt.Printf("API Key Set: %t\n", details.APIKeySet)
//		fmt.Printf("Warnings: %v\n", details.Warnings)
//	}
func (b *Builder) ValidateWithDetails() (*ValidationDetails, error) {
	details := &ValidationDetails{
		Provider:      string(b.provider),
		Model:         b.model,
		APIKeySet:     b.apiKey != "",
		AdapterSet:    b.adapter != nil,
		ToolCount:     len(b.tools),
		AutoExecute:   b.autoExecute,
		MemoryEnabled: b.memoryEnabled,
		Warnings:      []string{},
		Errors:        []string{},
	}

	var validationErr error

	// Run comprehensive validation and collect all errors
	if errs := b.validateConfigurationAll(); len(errs) > 0 {
		for _, err := range errs {
			details.Errors = append(details.Errors, err.Error())
		}
		validationErr = errs[0] // Return the first error for compatibility
	}

	// Add informational warnings
	if b.temperature == nil {
		details.Warnings = append(details.Warnings, "Temperature not set (will use model default)")
	}

	if b.maxTokens == nil {
		details.Warnings = append(details.Warnings, "MaxTokens not set (response length unlimited)")
	}

	if b.timeout == 0 {
		details.Warnings = append(details.Warnings, "No timeout set (requests may hang indefinitely)")
	}

	if len(b.tools) > 0 && !b.autoExecute {
		details.Warnings = append(details.Warnings, "Tools configured but auto-execute disabled")
	}

	return details, validationErr
}

// ValidationDetails provides comprehensive validation information
type ValidationDetails struct {
	// Core configuration
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	APIKeySet bool   `json:"api_key_set"`
	AdapterSet bool  `json:"adapter_set"`

	// Feature configuration
	ToolCount     int    `json:"tool_count"`
	AutoExecute   bool   `json:"auto_execute"`
	MemoryEnabled bool   `json:"memory_enabled"`

	// Validation results
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

// IsValid returns true if there are no validation errors
func (v *ValidationDetails) IsValid() bool {
	return len(v.Errors) == 0
}

// HasWarnings returns true if there are validation warnings
func (v *ValidationDetails) HasWarnings() bool {
	return len(v.Warnings) > 0
}

// Summary returns a human-readable summary
func (v *ValidationDetails) Summary() string {
	if len(v.Errors) > 0 {
		return fmt.Sprintf("Validation FAILED with %d errors, %d warnings", len(v.Errors), len(v.Warnings))
	}
	if len(v.Warnings) > 0 {
		return fmt.Sprintf("Validation PASSED with %d warnings", len(v.Warnings))
	}
	return "Validation PASSED - no issues found"
}
