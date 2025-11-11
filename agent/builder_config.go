package agent

import (
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
