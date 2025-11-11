package agent

import (
	"fmt"
	"time"
)

// AgentConfig represents the complete configuration for an agent
// This includes model settings, memory, retry, and tools configuration
type AgentConfig struct {
	// Model configuration
	Model       string  `yaml:"model" json:"model"`
	Temperature float64 `yaml:"temperature" json:"temperature"`
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
	TopP        float64 `yaml:"top_p" json:"top_p"`

	// Memory configuration
	Memory MemoryConfig `yaml:"memory" json:"memory"`

	// Retry configuration
	Retry RetryConfig `yaml:"retry" json:"retry"`

	// Tools configuration
	Tools ToolsConfig `yaml:"tools" json:"tools"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`

	// System prompt (for backward compatibility)
	SystemPrompt string `yaml:"system_prompt" json:"system_prompt"`
}

// MemoryConfig configures memory behavior
type MemoryConfig struct {
	WorkingCapacity   int     `yaml:"working_capacity" json:"working_capacity"`
	EpisodicEnabled   bool    `yaml:"episodic_enabled" json:"episodic_enabled"`
	EpisodicThreshold float64 `yaml:"episodic_threshold" json:"episodic_threshold"`
	SemanticEnabled   bool    `yaml:"semantic_enabled" json:"semantic_enabled"`
	AutoCompress      bool    `yaml:"auto_compress" json:"auto_compress"`
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts        int           `yaml:"max_attempts" json:"max_attempts"`
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`
	ExponentialBackoff bool          `yaml:"exponential_backoff" json:"exponential_backoff"`
	BackoffMultiplier  float64       `yaml:"backoff_multiplier" json:"backoff_multiplier"`
	InitialDelay       time.Duration `yaml:"initial_delay" json:"initial_delay"`
	MaxDelay           time.Duration `yaml:"max_delay" json:"max_delay"`
}

// ToolsConfig configures tools behavior
type ToolsConfig struct {
	ParallelExecution bool          `yaml:"parallel_execution" json:"parallel_execution"`
	MaxWorkers        int           `yaml:"max_workers" json:"max_workers"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout"`
}

// DefaultAgentConfig returns configuration with sensible defaults
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2000,
		TopP:        1.0,

		Memory: MemoryConfig{
			WorkingCapacity:   20,
			EpisodicEnabled:   true,
			EpisodicThreshold: 0.7,
			SemanticEnabled:   false,
			AutoCompress:      true,
		},

		Retry: RetryConfig{
			MaxAttempts:        3,
			Timeout:            30 * time.Second,
			ExponentialBackoff: true,
			BackoffMultiplier:  2.0,
			InitialDelay:       1 * time.Second,
			MaxDelay:           30 * time.Second,
		},

		Tools: ToolsConfig{
			ParallelExecution: false,
			MaxWorkers:        10,
			Timeout:           30 * time.Second,
		},

		RateLimit: DefaultRateLimitConfig(),
	}
}

// Validate checks if configuration is valid
func (c *AgentConfig) Validate() error {
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	if c.Temperature < 0 || c.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2, got: %f", c.Temperature)
	}

	if c.MaxTokens < 1 {
		return fmt.Errorf("max_tokens must be positive, got: %d", c.MaxTokens)
	}

	if c.TopP < 0 || c.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1, got: %f", c.TopP)
	}

	if c.Memory.WorkingCapacity < 1 {
		return fmt.Errorf("memory.working_capacity must be positive, got: %d", c.Memory.WorkingCapacity)
	}

	if c.Memory.EpisodicThreshold < 0 || c.Memory.EpisodicThreshold > 1 {
		return fmt.Errorf("memory.episodic_threshold must be between 0 and 1, got: %f", c.Memory.EpisodicThreshold)
	}

	if c.Retry.MaxAttempts < 1 {
		return fmt.Errorf("retry.max_attempts must be positive, got: %d", c.Retry.MaxAttempts)
	}

	if c.Retry.Timeout < 0 {
		return fmt.Errorf("retry.timeout must be non-negative, got: %v", c.Retry.Timeout)
	}

	if c.Retry.ExponentialBackoff {
		if c.Retry.BackoffMultiplier < 1 {
			return fmt.Errorf("retry.backoff_multiplier must be >= 1, got: %f", c.Retry.BackoffMultiplier)
		}

		if c.Retry.InitialDelay < 0 {
			return fmt.Errorf("retry.initial_delay must be non-negative, got: %v", c.Retry.InitialDelay)
		}

		if c.Retry.MaxDelay < c.Retry.InitialDelay {
			return fmt.Errorf("retry.max_delay must be >= initial_delay, got max: %v, initial: %v",
				c.Retry.MaxDelay, c.Retry.InitialDelay)
		}
	}

	if c.Tools.ParallelExecution {
		if c.Tools.MaxWorkers < 1 {
			return fmt.Errorf("tools.max_workers must be positive when parallel_execution is enabled, got: %d",
				c.Tools.MaxWorkers)
		}

		if c.Tools.Timeout < 0 {
			return fmt.Errorf("tools.timeout must be non-negative, got: %v", c.Tools.Timeout)
		}
	}

	// Validate rate limit configuration if enabled
	if c.RateLimit.Enabled {
		if c.RateLimit.RequestsPerSecond <= 0 {
			return fmt.Errorf("rate_limit.requests_per_second must be positive, got: %f",
				c.RateLimit.RequestsPerSecond)
		}

		if c.RateLimit.BurstSize < 1 {
			return fmt.Errorf("rate_limit.burst_size must be >= 1, got: %d", c.RateLimit.BurstSize)
		}

		if c.RateLimit.KeyTimeout < 0 {
			return fmt.Errorf("rate_limit.key_timeout must be non-negative, got: %v",
				c.RateLimit.KeyTimeout)
		}

		if c.RateLimit.WaitTimeout < 0 {
			return fmt.Errorf("rate_limit.wait_timeout must be non-negative, got: %v",
				c.RateLimit.WaitTimeout)
		}
	}

	return nil
}
