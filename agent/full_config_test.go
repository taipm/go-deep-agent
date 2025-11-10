package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFullConfig(t *testing.T) {
	// Load the full customer support config
	config, err := LoadFullConfig("../configs/full_customer_support.yaml")
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify metadata
	assert.Equal(t, "Production Customer Support Agent", config.Metadata.Name)
	assert.Equal(t, "production", config.Metadata.Environment)

	// Verify persona
	assert.True(t, config.HasPersona())
	assert.Equal(t, "CustomerSupportSpecialist", config.Persona.Name)
	assert.Equal(t, "Senior Customer Support Agent", config.Persona.Role)

	// Verify settings
	assert.True(t, config.HasSettings())
	assert.Equal(t, "gpt-4", config.Settings.Model)
	assert.Equal(t, 0.6, config.Settings.Temperature)
	assert.Equal(t, 3000, config.Settings.MaxTokens)

	// Verify settings override persona's technical_config
	assert.Equal(t, "gpt-4", config.GetModel())   // Settings override persona's gpt-4o-mini
	assert.Equal(t, 0.6, config.GetTemperature()) // Settings override persona's 0.7
}

func TestLoadFullConfig_InvalidFile(t *testing.T) {
	_, err := LoadFullConfig("nonexistent.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadFullConfig_InvalidYAML(t *testing.T) {
	// Create temp file with invalid YAML
	tmpfile, err := os.CreateTemp("", "invalid-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("invalid: yaml: content:\n  - broken")
	require.NoError(t, err)
	tmpfile.Close()

	_, err = LoadFullConfig(tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config YAML")
}

func TestLoadFullConfig_EmptyConfig(t *testing.T) {
	// Create temp file with empty config (should fail validation)
	tmpfile, err := os.CreateTemp("", "empty-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("metadata:\n  name: \"Empty\"\n")
	require.NoError(t, err)
	tmpfile.Close()

	_, err = LoadFullConfig(tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config must have at least persona or settings")
}

func TestLoadSettings(t *testing.T) {
	// Load production settings
	settings, err := LoadSettings("../configs/settings_production.yaml")
	require.NoError(t, err)
	require.NotNil(t, settings)

	assert.Equal(t, "gpt-4", settings.Model)
	assert.Equal(t, 0.6, settings.Temperature)
	assert.Equal(t, 3000, settings.MaxTokens)
	assert.Equal(t, 60*time.Second, settings.Timeout)

	// Verify memory settings
	require.NotNil(t, settings.Memory)
	assert.Equal(t, 100, settings.Memory.WorkingCapacity)
	assert.True(t, settings.Memory.EpisodicEnabled)

	// Verify retry settings
	require.NotNil(t, settings.Retry)
	assert.Equal(t, 5, settings.Retry.MaxAttempts)
	assert.True(t, settings.Retry.ExponentialBackoff)

	// Verify tools settings
	require.NotNil(t, settings.Tools)
	assert.True(t, settings.Tools.ParallelExecution)
	assert.Equal(t, 15, settings.Tools.MaxWorkers)
}

func TestLoadSettings_Development(t *testing.T) {
	// Load development settings
	settings, err := LoadSettings("../configs/settings_development.yaml")
	require.NoError(t, err)

	assert.Equal(t, "gpt-4o-mini", settings.Model)
	assert.Equal(t, 0.7, settings.Temperature)
	assert.Equal(t, 20*time.Second, settings.Timeout)

	// Development has lower capacity
	assert.Equal(t, 20, settings.Memory.WorkingCapacity)
	assert.Equal(t, 2, settings.Retry.MaxAttempts)
	assert.False(t, settings.Tools.ParallelExecution)
}

func TestFullConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *FullConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name:    "empty config",
			config:  &FullConfig{},
			wantErr: true,
			errMsg:  "config must have at least persona or settings",
		},
		{
			name: "valid persona only",
			config: &FullConfig{
				Persona: &Persona{
					Name: "Test",
					Role: "Assistant",
					Goal: "Help users",
					Personality: PersonalityConfig{
						Tone: "friendly",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid settings only",
			config: &FullConfig{
				Settings: &AgentSettings{
					Model:       "gpt-4",
					Temperature: 0.7,
				},
			},
			wantErr: false,
		},
		{
			name: "valid persona and settings",
			config: &FullConfig{
				Persona: &Persona{
					Name: "Test",
					Role: "Assistant",
					Goal: "Help users",
					Personality: PersonalityConfig{
						Tone: "friendly",
					},
				},
				Settings: &AgentSettings{
					Model:       "gpt-4",
					Temperature: 0.7,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid persona",
			config: &FullConfig{
				Persona: &Persona{
					// Missing required fields
					Name: "Test",
				},
			},
			wantErr: true,
			errMsg:  "persona validation failed",
		},
		{
			name: "invalid settings - bad temperature",
			config: &FullConfig{
				Settings: &AgentSettings{
					Temperature: 3.0, // Invalid: must be 0-2
				},
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
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

func TestAgentSettingsValidation(t *testing.T) {
	tests := []struct {
		name     string
		settings *AgentSettings
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  false, // nil is valid
		},
		{
			name: "valid settings",
			settings: &AgentSettings{
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   1000,
			},
			wantErr: false,
		},
		{
			name: "invalid temperature - too high",
			settings: &AgentSettings{
				Temperature: 2.5,
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2",
		},
		{
			name: "invalid temperature - negative",
			settings: &AgentSettings{
				Temperature: -0.1,
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 2",
		},
		{
			name: "invalid max_tokens",
			settings: &AgentSettings{
				MaxTokens: -100,
			},
			wantErr: true,
			errMsg:  "max_tokens must be non-negative",
		},
		{
			name: "invalid memory capacity",
			settings: &AgentSettings{
				Memory: &MemorySettings{
					WorkingCapacity: -10,
				},
			},
			wantErr: true,
			errMsg:  "memory.working_capacity must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
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

func TestBuilderWithFullConfig(t *testing.T) {
	config, err := LoadFullConfig("../configs/full_customer_support.yaml")
	require.NoError(t, err)

	builder := NewOpenAI("", "test-key").WithFullConfig(config)

	// Verify persona was applied
	assert.NotNil(t, builder.persona)
	assert.Equal(t, "CustomerSupportSpecialist", builder.persona.Name)

	// Verify settings override persona
	assert.Equal(t, "gpt-4", builder.model) // From settings, not persona's gpt-4o-mini
	assert.NotNil(t, builder.temperature)
	assert.Equal(t, 0.6, *builder.temperature) // From settings, not persona's 0.7
}

func TestBuilderWithSettings(t *testing.T) {
	settings, err := LoadSettings("../configs/settings_production.yaml")
	require.NoError(t, err)

	builder := NewOpenAI("gpt-4o-mini", "test-key").WithSettings(settings)

	// Verify settings were applied
	assert.Equal(t, "gpt-4", builder.model) // Override from settings
	assert.NotNil(t, builder.temperature)
	assert.Equal(t, 0.6, *builder.temperature)
	assert.NotNil(t, builder.maxTokens)
	assert.Equal(t, int64(3000), *builder.maxTokens)
}

func TestSettingsOverridePersona(t *testing.T) {
	// Load persona (has gpt-4o-mini, temp 0.7)
	persona, err := LoadPersona("../personas/customer_support.yaml")
	require.NoError(t, err)

	// Load settings (has gpt-4, temp 0.6)
	settings, err := LoadSettings("../configs/settings_production.yaml")
	require.NoError(t, err)

	// Apply persona first, then settings
	builder := NewOpenAI("", "test-key").
		WithPersona(persona).
		WithSettings(settings)

	// Settings should win
	assert.Equal(t, "gpt-4", builder.model)
	assert.NotNil(t, builder.temperature)
	assert.Equal(t, 0.6, *builder.temperature)
}

func TestCodeOverridesConfig(t *testing.T) {
	config, err := LoadFullConfig("../configs/full_customer_support.yaml")
	require.NoError(t, err)

	// Code overrides take precedence
	builder := NewOpenAI("", "test-key").
		WithFullConfig(config).
		WithTemperature(0.9) // Code override

	// Code override should win
	assert.NotNil(t, builder.temperature)
	assert.Equal(t, 0.9, *builder.temperature) // Not 0.6 from config
}

func TestSaveAndLoadFullConfig(t *testing.T) {
	// Create a full config
	original := &FullConfig{
		Metadata: &ConfigMetadata{
			Name:        "Test Config",
			Version:     "1.0.0",
			Environment: "test",
		},
		Persona: &Persona{
			Name: "TestAgent",
			Role: "Test Assistant",
			Goal: "Help with testing",
			Personality: PersonalityConfig{
				Tone: "friendly",
			},
		},
		Settings: &AgentSettings{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	}

	// Save to temp file
	tmpfile := filepath.Join(os.TempDir(), "test-config.yaml")
	defer os.Remove(tmpfile)

	err := SaveFullConfig(original, tmpfile)
	require.NoError(t, err)

	// Load back
	loaded, err := LoadFullConfig(tmpfile)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, original.Metadata.Name, loaded.Metadata.Name)
	assert.Equal(t, original.Persona.Name, loaded.Persona.Name)
	assert.Equal(t, original.Settings.Model, loaded.Settings.Model)
	assert.Equal(t, original.Settings.Temperature, loaded.Settings.Temperature)
}

func TestSaveAndLoadSettings(t *testing.T) {
	original := &AgentSettings{
		Model:       "gpt-4",
		Temperature: 0.8,
		MaxTokens:   2000,
		Memory: &MemorySettings{
			WorkingCapacity: 50,
			EpisodicEnabled: true,
		},
		Retry: &RetrySettings{
			MaxAttempts:        3,
			ExponentialBackoff: true,
		},
	}

	// Save to temp file
	tmpfile := filepath.Join(os.TempDir(), "test-settings.yaml")
	defer os.Remove(tmpfile)

	err := SaveSettings(original, tmpfile)
	require.NoError(t, err)

	// Load back
	loaded, err := LoadSettings(tmpfile)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, original.Model, loaded.Model)
	assert.Equal(t, original.Temperature, loaded.Temperature)
	assert.Equal(t, original.MaxTokens, loaded.MaxTokens)
	assert.Equal(t, original.Memory.WorkingCapacity, loaded.Memory.WorkingCapacity)
	assert.Equal(t, original.Retry.MaxAttempts, loaded.Retry.MaxAttempts)
}

func TestGetFullConfig(t *testing.T) {
	persona, err := LoadPersona("../personas/customer_support.yaml")
	require.NoError(t, err)

	// Start with gpt-4, then apply persona (which has gpt-4o-mini in technical_config)
	builder := NewOpenAI("gpt-4", "test-key").
		WithPersona(persona). // This will override model to gpt-4o-mini
		WithTemperature(0.8). // This overrides persona's temperature
		WithMaxTokens(2000)

	config := builder.GetFullConfig()

	assert.NotNil(t, config)
	assert.True(t, config.HasPersona())
	assert.True(t, config.HasSettings())
	assert.Equal(t, persona.Name, config.Persona.Name)
	// Persona's technical_config overrode builder's initial model
	assert.Equal(t, "gpt-4o-mini", config.Settings.Model)
	// But code overrides (WithTemperature) take precedence
	assert.Equal(t, 0.8, config.Settings.Temperature)
}

func TestToAgentSettings(t *testing.T) {
	builder := NewOpenAI("gpt-4", "test-key").
		WithTemperature(0.7).
		WithMaxTokens(1500).
		WithRetry(3).
		WithMaxWorkers(10)

	settings := builder.ToAgentSettings()

	assert.Equal(t, "gpt-4", settings.Model)
	assert.Equal(t, 0.7, settings.Temperature)
	assert.Equal(t, 1500, settings.MaxTokens)
	assert.NotNil(t, settings.Retry)
	assert.Equal(t, 3, settings.Retry.MaxAttempts)
	assert.NotNil(t, settings.Tools)
	assert.Equal(t, 10, settings.Tools.MaxWorkers)
}

func TestFullConfigString(t *testing.T) {
	config := &FullConfig{
		Persona: &Persona{
			Name: "TestAgent",
			Role: "Assistant",
			Goal: "Help",
			Personality: PersonalityConfig{
				Tone: "friendly",
			},
		},
		Settings: &AgentSettings{
			Model: "gpt-4",
		},
		Metadata: &ConfigMetadata{
			Name: "My Config",
		},
	}

	str := config.String()
	assert.Contains(t, str, "TestAgent")
	assert.Contains(t, str, "gpt-4")
	assert.Contains(t, str, "My Config")
}
