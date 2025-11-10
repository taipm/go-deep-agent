package agent

import (
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

	return config
}
