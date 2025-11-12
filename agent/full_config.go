package agent

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// FullConfig represents a complete agent configuration combining persona and settings.
// This is the Phase 3 implementation that allows flexible configuration where behavior
// (persona) and technical settings can be defined together or separately.
//
// Priority order for conflicting settings:
//
//	Code > FullConfig.Settings > Persona.TechnicalConfig > Defaults
type FullConfig struct {
	// Persona defines the agent's behavior, personality, and role
	Persona *Persona `yaml:"persona,omitempty"`

	// Settings defines technical configuration (model, memory, retry, etc.)
	// These settings override any technical_config defined in the persona
	Settings *AgentSettings `yaml:"settings,omitempty"`

	// Metadata provides additional information about the configuration
	Metadata *ConfigMetadata `yaml:"metadata,omitempty"`
}

// ConfigMetadata contains descriptive information about the configuration.
type ConfigMetadata struct {
	Name        string `yaml:"name,omitempty"`
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
	Environment string `yaml:"environment,omitempty"` // e.g., "development", "staging", "production"
	Author      string `yaml:"author,omitempty"`
	CreatedAt   string `yaml:"created_at,omitempty"`
	UpdatedAt   string `yaml:"updated_at,omitempty"`
}

// AgentSettings defines technical configuration for the agent.
// This is separate from persona to allow technical settings to be defined
// independently and reused across different personas.
type AgentSettings struct {
	// Model configuration
	Model       string        `yaml:"model,omitempty"`
	Temperature float64       `yaml:"temperature,omitempty"`
	MaxTokens   int           `yaml:"max_tokens,omitempty"`
	TopP        float64       `yaml:"top_p,omitempty"`
	Timeout     time.Duration `yaml:"timeout,omitempty"`

	// Memory configuration
	Memory *MemorySettings `yaml:"memory,omitempty"`

	// Retry configuration
	Retry *RetrySettings `yaml:"retry,omitempty"`

	// Tools configuration
	Tools *ToolsSettings `yaml:"tools,omitempty"`
}

// MemorySettings configures memory behavior
type MemorySettings struct {
	WorkingCapacity   int     `yaml:"working_capacity,omitempty"`
	EpisodicEnabled   bool    `yaml:"episodic_enabled,omitempty"`
	EpisodicThreshold float64 `yaml:"episodic_threshold,omitempty"`
	SemanticEnabled   bool    `yaml:"semantic_enabled,omitempty"`
	AutoCompress      bool    `yaml:"auto_compress,omitempty"`
}

// RetrySettings configures retry behavior
type RetrySettings struct {
	MaxAttempts        int           `yaml:"max_attempts,omitempty"`
	Timeout            time.Duration `yaml:"timeout,omitempty"`
	ExponentialBackoff bool          `yaml:"exponential_backoff,omitempty"`
	BackoffMultiplier  float64       `yaml:"backoff_multiplier,omitempty"`
	InitialDelay       time.Duration `yaml:"initial_delay,omitempty"`
	MaxDelay           time.Duration `yaml:"max_delay,omitempty"`
}

// ToolsSettings configures tools behavior
type ToolsSettings struct {
	ParallelExecution bool          `yaml:"parallel_execution,omitempty"`
	MaxWorkers        int           `yaml:"max_workers,omitempty"`
	Timeout           time.Duration `yaml:"timeout,omitempty"`
}

// LoadFullConfig loads a complete agent configuration from a YAML file.
// The file can contain persona, settings, or both.
func LoadFullConfig(path string) (*FullConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config FullConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// LoadSettings loads agent settings from a YAML file.
// This allows loading just the technical settings without a persona.
func LoadSettings(path string) (*AgentSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings AgentSettings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings YAML: %w", err)
	}

	if err := settings.Validate(); err != nil {
		return nil, fmt.Errorf("settings validation failed: %w", err)
	}

	return &settings, nil
}

// SaveFullConfig saves the full configuration to a YAML file.
func SaveFullConfig(config *FullConfig, path string) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid config: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveSettings saves the settings to a YAML file.
func SaveSettings(settings *AgentSettings, path string) error {
	if err := settings.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid settings: %w", err)
	}

	data, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// Validate checks if the full configuration is valid.
func (c *FullConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// At least one section must be present
	if c.Persona == nil && c.Settings == nil {
		return fmt.Errorf("config must have at least persona or settings")
	}

	// Validate persona if present
	if c.Persona != nil {
		if err := c.Persona.Validate(); err != nil {
			return fmt.Errorf("persona validation failed: %w", err)
		}
	}

	// Validate settings if present
	if c.Settings != nil {
		if err := c.Settings.Validate(); err != nil {
			return fmt.Errorf("settings validation failed: %w", err)
		}
	}

	return nil
}

// Validate checks if the agent settings are valid.
func (s *AgentSettings) Validate() error {
	if s == nil {
		return nil // nil settings is valid (use defaults)
	}

	// Model validation (optional - can be empty if using persona's model)
	if s.Temperature < 0 || s.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2, got: %f", s.Temperature)
	}

	if s.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative, got: %d", s.MaxTokens)
	}

	if s.TopP < 0 || s.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1, got: %f", s.TopP)
	}

	// Validate memory if present
	if s.Memory != nil {
		if s.Memory.WorkingCapacity < 0 {
			return fmt.Errorf("memory.working_capacity must be non-negative, got: %d", s.Memory.WorkingCapacity)
		}
		if s.Memory.EpisodicThreshold < 0 || s.Memory.EpisodicThreshold > 1 {
			return fmt.Errorf("memory.episodic_threshold must be between 0 and 1, got: %f", s.Memory.EpisodicThreshold)
		}
	}

	// Validate retry if present
	if s.Retry != nil {
		if s.Retry.MaxAttempts < 0 {
			return fmt.Errorf("retry.max_attempts must be non-negative, got: %d", s.Retry.MaxAttempts)
		}
	}

	// Validate tools if present
	if s.Tools != nil {
		if s.Tools.MaxWorkers < 0 {
			return fmt.Errorf("tools.max_workers must be non-negative, got: %d", s.Tools.MaxWorkers)
		}
	}

	return nil
}

// ApplyDefaults applies default values to missing fields.
func (s *AgentSettings) ApplyDefaults() {
	if s == nil {
		return
	}

	if s.Temperature == 0 {
		s.Temperature = 0.7
	}

	if s.Memory != nil && s.Memory.WorkingCapacity == 0 {
		s.Memory.WorkingCapacity = 20
	}

	if s.Retry != nil {
		if s.Retry.MaxAttempts == 0 {
			s.Retry.MaxAttempts = 3
		}
		if s.Retry.Timeout == 0 {
			s.Retry.Timeout = 30 * time.Second
		}
	}

	if s.Tools != nil && s.Tools.MaxWorkers == 0 {
		s.Tools.MaxWorkers = 10
	}
}

// HasPersona returns true if the configuration includes a persona.
func (c *FullConfig) HasPersona() bool {
	return c != nil && c.Persona != nil
}

// HasSettings returns true if the configuration includes settings.
func (c *FullConfig) HasSettings() bool {
	return c != nil && c.Settings != nil
}

// GetModel returns the model name from settings or persona's technical config.
// Settings take precedence over persona.
func (c *FullConfig) GetModel() string {
	if c.Settings != nil && c.Settings.Model != "" {
		return c.Settings.Model
	}
	if c.Persona != nil && c.Persona.TechnicalConfig != nil {
		return c.Persona.TechnicalConfig.Model
	}
	return ""
}

// GetTemperature returns the temperature from settings or persona's technical config.
// Settings take precedence over persona. Returns 0 if not set.
func (c *FullConfig) GetTemperature() float64 {
	if c.Settings != nil && c.Settings.Temperature > 0 {
		return c.Settings.Temperature
	}
	if c.Persona != nil && c.Persona.TechnicalConfig != nil {
		return c.Persona.TechnicalConfig.Temperature
	}
	return 0
}

// String returns a human-readable representation of the configuration.
func (c *FullConfig) String() string {
	if c == nil {
		return "FullConfig(nil)"
	}

	hasPersona := "no"
	if c.HasPersona() {
		hasPersona = fmt.Sprintf("yes (%s)", c.Persona.Name)
	}

	hasSettings := "no"
	if c.HasSettings() {
		hasSettings = fmt.Sprintf("yes (model: %s)", c.GetModel())
	}

	metadata := ""
	if c.Metadata != nil && c.Metadata.Name != "" {
		metadata = fmt.Sprintf(", metadata: %s", c.Metadata.Name)
	}

	return fmt.Sprintf("FullConfig(persona: %s, settings: %s%s)", hasPersona, hasSettings, metadata)
}
