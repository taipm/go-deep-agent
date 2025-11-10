package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAgentConfig(t *testing.T) {
	config := DefaultAgentConfig()

	// Model settings
	assert.Equal(t, "gpt-4", config.Model)
	assert.Equal(t, 0.7, config.Temperature)
	assert.Equal(t, 2000, config.MaxTokens)
	assert.Equal(t, 1.0, config.TopP)

	// Memory settings
	assert.Equal(t, 20, config.Memory.WorkingCapacity)
	assert.True(t, config.Memory.EpisodicEnabled)
	assert.Equal(t, 0.7, config.Memory.EpisodicThreshold)
	assert.False(t, config.Memory.SemanticEnabled)
	assert.True(t, config.Memory.AutoCompress)

	// Retry settings
	assert.Equal(t, 3, config.Retry.MaxAttempts)
	assert.Equal(t, 30*time.Second, config.Retry.Timeout)
	assert.True(t, config.Retry.ExponentialBackoff)
	assert.Equal(t, 2.0, config.Retry.BackoffMultiplier)
	assert.Equal(t, 1*time.Second, config.Retry.InitialDelay)
	assert.Equal(t, 30*time.Second, config.Retry.MaxDelay)

	// Tools settings
	assert.False(t, config.Tools.ParallelExecution)
	assert.Equal(t, 10, config.Tools.MaxWorkers)
	assert.Equal(t, 30*time.Second, config.Tools.Timeout)
}

func TestAgentConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*AgentConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *AgentConfig) {},
			wantErr: false,
		},
		{
			name: "missing model",
			modify: func(c *AgentConfig) {
				c.Model = ""
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "invalid temperature (negative)",
			modify: func(c *AgentConfig) {
				c.Temperature = -0.5
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2",
		},
		{
			name: "invalid temperature (too high)",
			modify: func(c *AgentConfig) {
				c.Temperature = 3.0
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2",
		},
		{
			name: "invalid max_tokens",
			modify: func(c *AgentConfig) {
				c.MaxTokens = 0
			},
			wantErr: true,
			errMsg:  "max_tokens must be positive",
		},
		{
			name: "invalid top_p (negative)",
			modify: func(c *AgentConfig) {
				c.TopP = -0.1
			},
			wantErr: true,
			errMsg:  "top_p must be between 0 and 1",
		},
		{
			name: "invalid top_p (too high)",
			modify: func(c *AgentConfig) {
				c.TopP = 1.5
			},
			wantErr: true,
			errMsg:  "top_p must be between 0 and 1",
		},
		{
			name: "invalid working_capacity",
			modify: func(c *AgentConfig) {
				c.Memory.WorkingCapacity = 0
			},
			wantErr: true,
			errMsg:  "memory.working_capacity must be positive",
		},
		{
			name: "invalid episodic_threshold (negative)",
			modify: func(c *AgentConfig) {
				c.Memory.EpisodicThreshold = -0.1
			},
			wantErr: true,
			errMsg:  "memory.episodic_threshold must be between 0 and 1",
		},
		{
			name: "invalid episodic_threshold (too high)",
			modify: func(c *AgentConfig) {
				c.Memory.EpisodicThreshold = 1.5
			},
			wantErr: true,
			errMsg:  "memory.episodic_threshold must be between 0 and 1",
		},
		{
			name: "invalid max_attempts",
			modify: func(c *AgentConfig) {
				c.Retry.MaxAttempts = 0
			},
			wantErr: true,
			errMsg:  "retry.max_attempts must be positive",
		},
		{
			name: "invalid timeout",
			modify: func(c *AgentConfig) {
				c.Retry.Timeout = -1 * time.Second
			},
			wantErr: true,
			errMsg:  "retry.timeout must be non-negative",
		},
		{
			name: "invalid backoff_multiplier",
			modify: func(c *AgentConfig) {
				c.Retry.ExponentialBackoff = true
				c.Retry.BackoffMultiplier = 0.5
			},
			wantErr: true,
			errMsg:  "retry.backoff_multiplier must be >= 1",
		},
		{
			name: "invalid initial_delay",
			modify: func(c *AgentConfig) {
				c.Retry.ExponentialBackoff = true
				c.Retry.InitialDelay = -1 * time.Second
			},
			wantErr: true,
			errMsg:  "retry.initial_delay must be non-negative",
		},
		{
			name: "invalid max_delay (less than initial)",
			modify: func(c *AgentConfig) {
				c.Retry.ExponentialBackoff = true
				c.Retry.InitialDelay = 10 * time.Second
				c.Retry.MaxDelay = 5 * time.Second
			},
			wantErr: true,
			errMsg:  "retry.max_delay must be >= initial_delay",
		},
		{
			name: "invalid max_workers",
			modify: func(c *AgentConfig) {
				c.Tools.ParallelExecution = true
				c.Tools.MaxWorkers = 0
			},
			wantErr: true,
			errMsg:  "tools.max_workers must be positive",
		},
		{
			name: "invalid tools timeout",
			modify: func(c *AgentConfig) {
				c.Tools.ParallelExecution = true
				c.Tools.Timeout = -1 * time.Second
			},
			wantErr: true,
			errMsg:  "tools.timeout must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultAgentConfig()
			tt.modify(config)

			err := config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadAgentConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `model: "gpt-4-turbo"
temperature: 0.5
max_tokens: 3000
top_p: 0.9

memory:
  working_capacity: 30
  episodic_enabled: true
  episodic_threshold: 0.8
  semantic_enabled: true
  auto_compress: false

retry:
  max_attempts: 5
  timeout: 60s
  exponential_backoff: true
  backoff_multiplier: 2.5
  initial_delay: 2s
  max_delay: 60s

tools:
  parallel_execution: true
  max_workers: 20
  timeout: 45s

system_prompt: "You are a helpful assistant"
`

	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	require.NoError(t, err)

	// Load config
	config, err := LoadAgentConfig(configPath)
	require.NoError(t, err)

	// Verify model settings
	assert.Equal(t, "gpt-4-turbo", config.Model)
	assert.Equal(t, 0.5, config.Temperature)
	assert.Equal(t, 3000, config.MaxTokens)
	assert.Equal(t, 0.9, config.TopP)

	// Verify memory settings
	assert.Equal(t, 30, config.Memory.WorkingCapacity)
	assert.True(t, config.Memory.EpisodicEnabled)
	assert.Equal(t, 0.8, config.Memory.EpisodicThreshold)
	assert.True(t, config.Memory.SemanticEnabled)
	assert.False(t, config.Memory.AutoCompress)

	// Verify retry settings
	assert.Equal(t, 5, config.Retry.MaxAttempts)
	assert.Equal(t, 60*time.Second, config.Retry.Timeout)
	assert.True(t, config.Retry.ExponentialBackoff)
	assert.Equal(t, 2.5, config.Retry.BackoffMultiplier)
	assert.Equal(t, 2*time.Second, config.Retry.InitialDelay)
	assert.Equal(t, 60*time.Second, config.Retry.MaxDelay)

	// Verify tools settings
	assert.True(t, config.Tools.ParallelExecution)
	assert.Equal(t, 20, config.Tools.MaxWorkers)
	assert.Equal(t, 45*time.Second, config.Tools.Timeout)

	// Verify system prompt
	assert.Equal(t, "You are a helpful assistant", config.SystemPrompt)
}

func TestLoadAgentConfig_InvalidFile(t *testing.T) {
	_, err := LoadAgentConfig("nonexistent.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadAgentConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "bad.yaml")

	badYAML := `model: "gpt-4"
temperature: invalid_number
`
	err := os.WriteFile(configPath, []byte(badYAML), 0644)
	require.NoError(t, err)

	_, err = LoadAgentConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestLoadAgentConfig_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `model: ""
temperature: 0.5
`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	_, err = LoadAgentConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration")
}

func TestLoadAgentConfigWithEnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configYAML := `model: "gpt-4"
temperature: 0.7
max_tokens: 2000
memory:
  working_capacity: 20
`
	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	require.NoError(t, err)

	// Set environment variables
	os.Setenv("AGENT_MODEL", "gpt-4-turbo")
	os.Setenv("AGENT_TEMPERATURE", "0.3")
	os.Setenv("AGENT_MAX_TOKENS", "4000")
	os.Setenv("AGENT_MEMORY_CAPACITY", "50")
	defer func() {
		os.Unsetenv("AGENT_MODEL")
		os.Unsetenv("AGENT_TEMPERATURE")
		os.Unsetenv("AGENT_MAX_TOKENS")
		os.Unsetenv("AGENT_MEMORY_CAPACITY")
	}()

	// Load with overrides
	config, err := LoadAgentConfigWithEnvOverrides(configPath)
	require.NoError(t, err)

	// Verify overrides
	assert.Equal(t, "gpt-4-turbo", config.Model)
	assert.Equal(t, 0.3, config.Temperature)
	assert.Equal(t, 4000, config.MaxTokens)
	assert.Equal(t, 50, config.Memory.WorkingCapacity)
}

func TestSaveAgentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "output.yaml")

	// Create config
	config := DefaultAgentConfig()
	config.Model = "gpt-4-turbo"
	config.Temperature = 0.3
	config.SystemPrompt = "Test prompt"

	// Save
	err := SaveAgentConfig(config, configPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load back and verify
	loaded, err := LoadAgentConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, config.Model, loaded.Model)
	assert.Equal(t, config.Temperature, loaded.Temperature)
	assert.Equal(t, config.SystemPrompt, loaded.SystemPrompt)
}

func TestSaveAgentConfig_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "output.yaml")

	// Create invalid config
	config := DefaultAgentConfig()
	config.Model = "" // Invalid

	// Save should fail
	err := SaveAgentConfig(config, configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot save invalid configuration")
}

func TestBuilderWithAgentConfig(t *testing.T) {
	config := DefaultAgentConfig()
	config.Model = "gpt-4-turbo"
	config.Temperature = 0.3
	config.MaxTokens = 3000
	config.SystemPrompt = "You are a helpful assistant"
	config.Memory.WorkingCapacity = 50
	config.Retry.MaxAttempts = 5

	builder := NewOpenAI("", "test-key").WithAgentConfig(config)

	assert.Equal(t, "gpt-4-turbo", builder.model)
	assert.NotNil(t, builder.temperature)
	assert.Equal(t, 0.3, *builder.temperature)
	assert.NotNil(t, builder.maxTokens)
	assert.Equal(t, int64(3000), *builder.maxTokens)
	assert.Equal(t, "You are a helpful assistant", builder.systemPrompt)
}

func TestBuilderToAgentConfig(t *testing.T) {
	temp := 0.3
	maxTokens := int64(3000)
	topP := 0.9

	builder := NewOpenAI("gpt-4-turbo", "test-key")
	builder.temperature = &temp
	builder.maxTokens = &maxTokens
	builder.topP = &topP
	builder.systemPrompt = "Test prompt"
	builder.maxRetries = 5

	config := builder.ToAgentConfig()

	assert.Equal(t, "gpt-4-turbo", config.Model)
	assert.Equal(t, 0.3, config.Temperature)
	assert.Equal(t, 3000, config.MaxTokens)
	assert.Equal(t, 0.9, config.TopP)
	assert.Equal(t, "Test prompt", config.SystemPrompt)
	assert.Equal(t, 5, config.Retry.MaxAttempts)
}

func TestConfigRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "roundtrip.yaml")

	// Create original config
	original := DefaultAgentConfig()
	original.Model = "gpt-4-turbo"
	original.Temperature = 0.5
	original.SystemPrompt = "Test"

	// Save
	err := SaveAgentConfig(original, configPath)
	require.NoError(t, err)

	// Load
	loaded, err := LoadAgentConfig(configPath)
	require.NoError(t, err)

	// Save again
	configPath2 := filepath.Join(tmpDir, "roundtrip2.yaml")
	err = SaveAgentConfig(loaded, configPath2)
	require.NoError(t, err)

	// Load again
	loaded2, err := LoadAgentConfig(configPath2)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.Model, loaded2.Model)
	assert.Equal(t, original.Temperature, loaded2.Temperature)
	assert.Equal(t, original.SystemPrompt, loaded2.SystemPrompt)
}
