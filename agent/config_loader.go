package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// LoadAgentConfig loads configuration from a YAML file
func LoadAgentConfig(path string) (*AgentConfig, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	config := DefaultAgentConfig() // Start with defaults
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadAgentConfigWithEnvOverrides loads config and applies environment variable overrides
// Environment variables:
//   - AGENT_MODEL: Override model name
//   - AGENT_TEMPERATURE: Override temperature (float)
//   - AGENT_MAX_TOKENS: Override max tokens (int)
//   - AGENT_MEMORY_CAPACITY: Override working memory capacity (int)
func LoadAgentConfigWithEnvOverrides(path string) (*AgentConfig, error) {
	config, err := LoadAgentConfig(path)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if present
	if model := os.Getenv("AGENT_MODEL"); model != "" {
		config.Model = model
	}

	if temp := os.Getenv("AGENT_TEMPERATURE"); temp != "" {
		if t, err := strconv.ParseFloat(temp, 64); err == nil {
			config.Temperature = t
		}
	}

	if maxTokens := os.Getenv("AGENT_MAX_TOKENS"); maxTokens != "" {
		if tokens, err := strconv.Atoi(maxTokens); err == nil {
			config.MaxTokens = tokens
		}
	}

	if capacity := os.Getenv("AGENT_MEMORY_CAPACITY"); capacity != "" {
		if cap, err := strconv.Atoi(capacity); err == nil {
			config.Memory.WorkingCapacity = cap
		}
	}

	// Validate again after overrides
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration after env overrides: %w", err)
	}

	return config, nil
}

// SaveAgentConfig saves configuration to a YAML file
func SaveAgentConfig(config *AgentConfig, path string) error {
	// Validate before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid configuration: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
