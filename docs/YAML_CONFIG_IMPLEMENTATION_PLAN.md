# YAML Configuration Implementation Plan

**Date**: November 10, 2025  
**Goal**: Implement Hybrid YAML Config (Traditional + Persona-Based)  
**Approach**: Incremental, 3-phase rollout with full developer support  
**Key Principle**: Each persona = separate YAML file

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Phase 1: Traditional Config (v0.6.2)](#phase-1-traditional-config)
3. [Phase 2: Persona Support (v0.6.3)](#phase-2-persona-support)
4. [Phase 3: Hybrid Integration (v0.7.0)](#phase-3-hybrid-integration)
5. [Testing Strategy](#testing-strategy)
6. [Documentation Plan](#documentation-plan)
7. [Migration Guide](#migration-guide)

---

## Executive Summary

### Timeline

| Phase | Version | Duration | Deliverables |
|-------|---------|----------|--------------|
| Phase 1 | v0.6.2 | 5 days | Traditional config loader |
| Phase 2 | v0.6.3 | 5 days | Persona system |
| Phase 3 | v0.7.0 | 3 days | Hybrid integration + polish |
| **Total** | - | **13 days** | **Production-ready YAML config** |

### Key Design Decisions

1. **Each persona = separate YAML file** (reusability)
2. **Code & keywords in English** (international standard)
3. **Vietnamese docs** (for Vietnamese developers)
4. **Progressive enhancement** (simple → advanced)
5. **Backward compatible** (no breaking changes)

---

## Phase 1: Traditional Config (v0.6.2)

**Duration**: 5 days (Nov 11-15, 2025)  
**Goal**: Load technical settings from YAML files

### Day 1: Core Structures (Nov 11)

#### File: `agent/config.go` (NEW - ~200 lines)

```go
package agent

import (
    "time"
)

// Config represents the complete configuration for an agent
type Config struct {
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
    MaxAttempts       int           `yaml:"max_attempts" json:"max_attempts"`
    Timeout           time.Duration `yaml:"timeout" json:"timeout"`
    ExponentialBackoff bool         `yaml:"exponential_backoff" json:"exponential_backoff"`
    BackoffMultiplier float64       `yaml:"backoff_multiplier" json:"backoff_multiplier"`
    InitialDelay      time.Duration `yaml:"initial_delay" json:"initial_delay"`
    MaxDelay          time.Duration `yaml:"max_delay" json:"max_delay"`
}

// ToolsConfig configures tools behavior
type ToolsConfig struct {
    ParallelExecution bool          `yaml:"parallel_execution" json:"parallel_execution"`
    MaxWorkers        int           `yaml:"max_workers" json:"max_workers"`
    Timeout           time.Duration `yaml:"timeout" json:"timeout"`
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
    return &Config{
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
    }
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
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
    
    if c.Retry.MaxAttempts < 1 {
        return fmt.Errorf("retry.max_attempts must be positive, got: %d", c.Retry.MaxAttempts)
    }
    
    return nil
}
```

#### File: `agent/config_loader.go` (NEW - ~150 lines)

```go
package agent

import (
    "fmt"
    "os"
    "path/filepath"
    
    "gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
    // Read file
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    // Parse YAML
    config := DefaultConfig() // Start with defaults
    if err := yaml.Unmarshal(data, config); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }
    
    // Validate
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return config, nil
}

// LoadConfigWithEnvOverrides loads config and applies environment variable overrides
func LoadConfigWithEnvOverrides(path string) (*Config, error) {
    config, err := LoadConfig(path)
    if err != nil {
        return nil, err
    }
    
    // Override with environment variables if present
    if model := os.Getenv("AGENT_MODEL"); model != "" {
        config.Model = model
    }
    
    if temp := os.Getenv("AGENT_TEMPERATURE"); temp != "" {
        var t float64
        if _, err := fmt.Sscanf(temp, "%f", &t); err == nil {
            config.Temperature = t
        }
    }
    
    return config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, path string) error {
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
```

#### File: `agent/builder_config.go` (NEW - ~100 lines)

```go
package agent

// WithConfig applies a complete configuration
func (b *Builder) WithConfig(config *Config) *Builder {
    // Model settings
    b.model = config.Model
    b.temperature = config.Temperature
    b.maxTokens = config.MaxTokens
    b.topP = config.TopP
    
    // Memory settings
    if b.memory != nil {
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
    b.retryAttempts = config.Retry.MaxAttempts
    b.retryTimeout = config.Retry.Timeout
    
    // Tools settings
    if config.Tools.ParallelExecution {
        b.WithParallelTools(true)
        b.WithMaxWorkers(config.Tools.MaxWorkers)
        b.WithToolTimeout(config.Tools.Timeout)
    }
    
    // System prompt (backward compatibility)
    if config.SystemPrompt != "" {
        b.WithSystem(config.SystemPrompt)
    }
    
    return b
}

// ToConfig exports current builder state as Config
func (b *Builder) ToConfig() *Config {
    config := DefaultConfig()
    
    config.Model = b.model
    config.Temperature = b.temperature
    config.MaxTokens = b.maxTokens
    config.TopP = b.topP
    
    if b.memory != nil {
        memConfig := b.memory.GetConfig()
        config.Memory.WorkingCapacity = memConfig.WorkingCapacity
        config.Memory.EpisodicEnabled = memConfig.EpisodicEnabled
        config.Memory.EpisodicThreshold = memConfig.EpisodicThreshold
        config.Memory.SemanticEnabled = memConfig.SemanticEnabled
        config.Memory.AutoCompress = memConfig.AutoCompress
    }
    
    config.Retry.MaxAttempts = b.retryAttempts
    config.Retry.Timeout = b.retryTimeout
    
    config.Tools.ParallelExecution = b.parallelTools
    config.Tools.MaxWorkers = b.maxWorkers
    config.Tools.Timeout = b.toolTimeout
    
    return config
}
```

---

### Day 2: Testing & Examples (Nov 12)

#### File: `agent/config_test.go` (NEW - ~300 lines)

```go
package agent

import (
    "os"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
    config := DefaultConfig()
    
    assert.Equal(t, "gpt-4", config.Model)
    assert.Equal(t, 0.7, config.Temperature)
    assert.Equal(t, 2000, config.MaxTokens)
    assert.Equal(t, 20, config.Memory.WorkingCapacity)
    assert.Equal(t, 3, config.Retry.MaxAttempts)
}

func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        modify  func(*Config)
        wantErr bool
    }{
        {
            name:    "valid config",
            modify:  func(c *Config) {},
            wantErr: false,
        },
        {
            name: "invalid temperature (negative)",
            modify: func(c *Config) {
                c.Temperature = -0.5
            },
            wantErr: true,
        },
        {
            name: "invalid temperature (too high)",
            modify: func(c *Config) {
                c.Temperature = 3.0
            },
            wantErr: true,
        },
        {
            name: "invalid max_tokens",
            modify: func(c *Config) {
                c.MaxTokens = 0
            },
            wantErr: true,
        },
        {
            name: "invalid top_p",
            modify: func(c *Config) {
                c.TopP = 1.5
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := DefaultConfig()
            tt.modify(config)
            
            err := config.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestLoadConfig(t *testing.T) {
    // Create temp config file
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.yaml")
    
    configYAML := `
model: "gpt-4-turbo"
temperature: 0.5
max_tokens: 3000

memory:
  working_capacity: 30
  episodic_enabled: true
  episodic_threshold: 0.8

retry:
  max_attempts: 5
  timeout: 60s
  exponential_backoff: true
`
    
    err := os.WriteFile(configPath, []byte(configYAML), 0644)
    require.NoError(t, err)
    
    // Load config
    config, err := LoadConfig(configPath)
    require.NoError(t, err)
    
    // Verify
    assert.Equal(t, "gpt-4-turbo", config.Model)
    assert.Equal(t, 0.5, config.Temperature)
    assert.Equal(t, 3000, config.MaxTokens)
    assert.Equal(t, 30, config.Memory.WorkingCapacity)
    assert.Equal(t, 0.8, config.Memory.EpisodicThreshold)
    assert.Equal(t, 5, config.Retry.MaxAttempts)
    assert.Equal(t, 60*time.Second, config.Retry.Timeout)
}

func TestSaveConfig(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.yaml")
    
    // Create config
    config := DefaultConfig()
    config.Model = "gpt-4-turbo"
    config.Temperature = 0.5
    
    // Save
    err := SaveConfig(config, configPath)
    require.NoError(t, err)
    
    // Load back
    loaded, err := LoadConfig(configPath)
    require.NoError(t, err)
    
    assert.Equal(t, config.Model, loaded.Model)
    assert.Equal(t, config.Temperature, loaded.Temperature)
}

func TestBuilderWithConfig(t *testing.T) {
    config := DefaultConfig()
    config.Model = "gpt-4-turbo"
    config.Temperature = 0.3
    config.SystemPrompt = "You are a helpful assistant"
    
    builder := NewOpenAI("test-key").WithConfig(config)
    
    assert.Equal(t, "gpt-4-turbo", builder.model)
    assert.Equal(t, 0.3, builder.temperature)
}

func TestBuilderToConfig(t *testing.T) {
    builder := NewOpenAI("test-key").
        WithModel("gpt-4-turbo").
        WithTemperature(0.3).
        WithMaxTokens(3000)
    
    config := builder.ToConfig()
    
    assert.Equal(t, "gpt-4-turbo", config.Model)
    assert.Equal(t, 0.3, config.Temperature)
    assert.Equal(t, 3000, config.MaxTokens)
}
```

#### File: `config/example.yaml` (NEW)

```yaml
# Example configuration for go-deep-agent
# This file demonstrates all available configuration options

# Model configuration
model: "gpt-4"
temperature: 0.7        # 0.0 (deterministic) to 2.0 (creative)
max_tokens: 2000        # Maximum tokens in response
top_p: 1.0              # Nucleus sampling (0.0 to 1.0)

# System prompt (optional)
system_prompt: |
  You are a helpful AI assistant.
  You provide accurate and concise responses.

# Memory configuration
memory:
  working_capacity: 20           # Number of recent messages to keep
  episodic_enabled: true         # Enable long-term episodic memory
  episodic_threshold: 0.7        # Importance threshold (0.0 to 1.0)
  semantic_enabled: false        # Enable semantic fact storage
  auto_compress: true            # Automatically compress old memories

# Retry configuration
retry:
  max_attempts: 3                # Maximum retry attempts
  timeout: 30s                   # Timeout per request
  exponential_backoff: true      # Use exponential backoff
  backoff_multiplier: 2.0        # Backoff multiplier
  initial_delay: 1s              # Initial retry delay
  max_delay: 30s                 # Maximum retry delay

# Tools configuration
tools:
  parallel_execution: false      # Execute tools in parallel
  max_workers: 10                # Maximum parallel workers
  timeout: 30s                   # Timeout per tool execution
```

#### File: `examples/config_basic/main.go` (NEW)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Load configuration from YAML file
    config, err := agent.LoadConfig("config.yaml")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Create agent with configuration
    apiKey := os.Getenv("OPENAI_API_KEY")
    myAgent := agent.NewOpenAI(apiKey).
        WithConfig(config).
        Build()
    
    // Use the agent
    ctx := context.Background()
    response, err := myAgent.Ask(ctx, "What is the capital of France?")
    if err != nil {
        log.Fatal("Agent error:", err)
    }
    
    fmt.Println("Response:", response)
    
    // Export current configuration
    currentConfig := myAgent.ToConfig()
    if err := agent.SaveConfig(currentConfig, "exported_config.yaml"); err != nil {
        log.Fatal("Failed to save config:", err)
    }
    
    fmt.Println("Configuration exported to exported_config.yaml")
}
```

#### File: `examples/config_basic/config.yaml` (NEW)

```yaml
model: "gpt-4"
temperature: 0.7
max_tokens: 2000

memory:
  working_capacity: 20
  episodic_enabled: true
  episodic_threshold: 0.7

retry:
  max_attempts: 3
  timeout: 30s
  exponential_backoff: true

system_prompt: |
  You are a knowledgeable AI assistant.
  Provide accurate and helpful responses.
```

---

### Day 3-4: Schema & Validation (Nov 13-14)

#### File: `config/schema.json` (NEW)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "go-deep-agent Configuration",
  "type": "object",
  "properties": {
    "model": {
      "type": "string",
      "description": "LLM model name",
      "examples": ["gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"]
    },
    "temperature": {
      "type": "number",
      "minimum": 0,
      "maximum": 2,
      "description": "Controls randomness (0 = deterministic, 2 = creative)"
    },
    "max_tokens": {
      "type": "integer",
      "minimum": 1,
      "description": "Maximum tokens in response"
    },
    "top_p": {
      "type": "number",
      "minimum": 0,
      "maximum": 1,
      "description": "Nucleus sampling parameter"
    },
    "system_prompt": {
      "type": "string",
      "description": "System prompt for the agent"
    },
    "memory": {
      "type": "object",
      "properties": {
        "working_capacity": {
          "type": "integer",
          "minimum": 1,
          "description": "Number of recent messages to keep in working memory"
        },
        "episodic_enabled": {
          "type": "boolean",
          "description": "Enable long-term episodic memory"
        },
        "episodic_threshold": {
          "type": "number",
          "minimum": 0,
          "maximum": 1,
          "description": "Importance threshold for episodic storage"
        },
        "semantic_enabled": {
          "type": "boolean",
          "description": "Enable semantic fact storage"
        },
        "auto_compress": {
          "type": "boolean",
          "description": "Automatically compress old memories"
        }
      }
    },
    "retry": {
      "type": "object",
      "properties": {
        "max_attempts": {
          "type": "integer",
          "minimum": 1,
          "description": "Maximum retry attempts"
        },
        "timeout": {
          "type": "string",
          "pattern": "^[0-9]+(ns|us|ms|s|m|h)$",
          "description": "Timeout per request (e.g., '30s', '1m')"
        },
        "exponential_backoff": {
          "type": "boolean",
          "description": "Use exponential backoff"
        },
        "backoff_multiplier": {
          "type": "number",
          "minimum": 1,
          "description": "Backoff multiplier"
        },
        "initial_delay": {
          "type": "string",
          "pattern": "^[0-9]+(ns|us|ms|s|m|h)$",
          "description": "Initial retry delay"
        },
        "max_delay": {
          "type": "string",
          "pattern": "^[0-9]+(ns|us|ms|s|m|h)$",
          "description": "Maximum retry delay"
        }
      }
    },
    "tools": {
      "type": "object",
      "properties": {
        "parallel_execution": {
          "type": "boolean",
          "description": "Execute tools in parallel"
        },
        "max_workers": {
          "type": "integer",
          "minimum": 1,
          "description": "Maximum parallel workers"
        },
        "timeout": {
          "type": "string",
          "pattern": "^[0-9]+(ns|us|ms|s|m|h)$",
          "description": "Timeout per tool execution"
        }
      }
    }
  },
  "required": ["model"]
}
```

---

### Day 5: Documentation (Nov 15)

#### File: `docs/CONFIG_GUIDE.md` (NEW - Vietnamese)

*(Will create this with full Vietnamese documentation)*

---

## Phase 2: Persona Support (v0.6.3)

**Duration**: 5 days (Nov 16-20, 2025)  
**Goal**: Persona-based configuration with **one file per persona**

### Directory Structure

```
personas/
  customer_support.yaml       # ✅ Separate file
  technical_writer.yaml       # ✅ Separate file
  code_reviewer.yaml          # ✅ Separate file
  sales_rep.yaml              # ✅ Separate file
  data_analyst.yaml           # ✅ Separate file
  # ... more personas
  
  templates/                  # ✅ Reusable templates
    base_assistant.yaml
    base_support.yaml
    base_writer.yaml
```

### Day 1: Persona Structures (Nov 16)

#### File: `agent/persona.go` (NEW - ~300 lines)

```go
package agent

// Persona represents an agent's behavioral configuration
type Persona struct {
    // Metadata
    Name        string `yaml:"name" json:"name"`
    Version     string `yaml:"version" json:"version"`
    Description string `yaml:"description" json:"description"`
    
    // Core identity
    Role      string `yaml:"role" json:"role"`
    Goal      string `yaml:"goal" json:"goal"`
    Backstory string `yaml:"backstory" json:"backstory"`
    
    // Personality
    Personality PersonalityConfig `yaml:"personality" json:"personality"`
    
    // Behavior rules
    Guidelines  []string `yaml:"guidelines" json:"guidelines"`
    Constraints []string `yaml:"constraints" json:"constraints"`
    
    // Knowledge areas
    KnowledgeAreas []string `yaml:"knowledge_areas" json:"knowledge_areas"`
    
    // Examples (optional)
    Examples []PersonaExample `yaml:"examples" json:"examples"`
    
    // Optional technical overrides
    TechnicalConfig *Config `yaml:"technical_config,omitempty" json:"technical_config,omitempty"`
}

// PersonalityConfig defines personality traits
type PersonalityConfig struct {
    Tone   string   `yaml:"tone" json:"tone"`
    Traits []string `yaml:"traits" json:"traits"`
    Style  string   `yaml:"style" json:"style"`
}

// PersonaExample provides usage examples
type PersonaExample struct {
    Scenario string `yaml:"scenario" json:"scenario"`
    Response string `yaml:"response" json:"response"`
}

// ToSystemPrompt generates a system prompt from persona
func (p *Persona) ToSystemPrompt() string {
    var builder strings.Builder
    
    // Role
    builder.WriteString(fmt.Sprintf("You are a %s.\n\n", p.Role))
    
    // Goal
    if p.Goal != "" {
        builder.WriteString(fmt.Sprintf("Your goal: %s\n\n", p.Goal))
    }
    
    // Backstory
    if p.Backstory != "" {
        builder.WriteString(fmt.Sprintf("%s\n\n", p.Backstory))
    }
    
    // Personality
    if p.Personality.Tone != "" {
        builder.WriteString(fmt.Sprintf("Tone: %s\n", p.Personality.Tone))
    }
    if len(p.Personality.Traits) > 0 {
        builder.WriteString(fmt.Sprintf("Traits: %s\n\n", strings.Join(p.Personality.Traits, ", ")))
    }
    
    // Guidelines
    if len(p.Guidelines) > 0 {
        builder.WriteString("Guidelines:\n")
        for _, g := range p.Guidelines {
            builder.WriteString(fmt.Sprintf("- %s\n", g))
        }
        builder.WriteString("\n")
    }
    
    // Constraints
    if len(p.Constraints) > 0 {
        builder.WriteString("Important constraints:\n")
        for _, c := range p.Constraints {
            builder.WriteString(fmt.Sprintf("- %s\n", c))
        }
        builder.WriteString("\n")
    }
    
    // Knowledge areas
    if len(p.KnowledgeAreas) > 0 {
        builder.WriteString(fmt.Sprintf("Knowledge areas: %s\n", strings.Join(p.KnowledgeAreas, ", ")))
    }
    
    return builder.String()
}

// Validate checks if persona is valid
func (p *Persona) Validate() error {
    if p.Name == "" {
        return fmt.Errorf("persona name is required")
    }
    
    if p.Role == "" {
        return fmt.Errorf("persona role is required")
    }
    
    return nil
}
```

#### File: `agent/persona_loader.go` (NEW - ~200 lines)

```go
package agent

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "gopkg.in/yaml.v3"
)

// LoadPersona loads a persona from a YAML file
func LoadPersona(path string) (*Persona, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read persona file: %w", err)
    }
    
    var persona Persona
    if err := yaml.Unmarshal(data, &persona); err != nil {
        return nil, fmt.Errorf("failed to parse persona YAML: %w", err)
    }
    
    if err := persona.Validate(); err != nil {
        return nil, fmt.Errorf("invalid persona: %w", err)
    }
    
    return &persona, nil
}

// LoadPersonasFromDirectory loads all personas from a directory
func LoadPersonasFromDirectory(dir string) (map[string]*Persona, error) {
    personas := make(map[string]*Persona)
    
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, fmt.Errorf("failed to read directory: %w", err)
    }
    
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
            continue
        }
        
        path := filepath.Join(dir, entry.Name())
        persona, err := LoadPersona(path)
        if err != nil {
            return nil, fmt.Errorf("failed to load %s: %w", entry.Name(), err)
        }
        
        personas[persona.Name] = persona
    }
    
    return personas, nil
}

// SavePersona saves a persona to a YAML file
func SavePersona(persona *Persona, path string) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    
    data, err := yaml.Marshal(persona)
    if err != nil {
        return fmt.Errorf("failed to marshal persona: %w", err)
    }
    
    if err := os.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("failed to write persona file: %w", err)
    }
    
    return nil
}

// PersonaRegistry manages multiple personas
type PersonaRegistry struct {
    personas map[string]*Persona
}

// NewPersonaRegistry creates a new persona registry
func NewPersonaRegistry() *PersonaRegistry {
    return &PersonaRegistry{
        personas: make(map[string]*Persona),
    }
}

// Register adds a persona to the registry
func (r *PersonaRegistry) Register(persona *Persona) error {
    if err := persona.Validate(); err != nil {
        return err
    }
    
    r.personas[persona.Name] = persona
    return nil
}

// Get retrieves a persona by name
func (r *PersonaRegistry) Get(name string) (*Persona, error) {
    persona, ok := r.personas[name]
    if !ok {
        return nil, fmt.Errorf("persona not found: %s", name)
    }
    
    return persona, nil
}

// List returns all persona names
func (r *PersonaRegistry) List() []string {
    names := make([]string, 0, len(r.personas))
    for name := range r.personas {
        names = append(names, name)
    }
    return names
}
```

---

### Day 2: Example Personas (Nov 17)

#### File: `personas/customer_support.yaml` (NEW)

```yaml
name: customer_support
version: "1.0.0"
description: "Customer support specialist for handling inquiries and issues"

role: "Senior Customer Support Specialist"
goal: "Resolve customer issues with empathy and efficiency"

backstory: |
  You are an experienced customer support professional with 8 years in SaaS companies.
  You're known for turning frustrated customers into advocates through patient listening
  and clear problem-solving.

personality:
  tone: "warm, professional, and reassuring"
  traits:
    - empathetic
    - patient
    - solution-oriented
    - proactive
  style: |
    - Use customer's name when appropriate
    - Acknowledge emotions before solutions
    - Break down complex steps
    - Confirm understanding

guidelines:
  - "Start with a warm greeting"
  - "Ask clarifying questions"
  - "Provide estimated resolution time"
  - "Summarize action items at the end"
  - "Follow up to ensure satisfaction"

constraints:
  - "Never promise features that don't exist"
  - "Don't share internal information"
  - "Escalate if customer is very upset"
  - "Protect privacy - never ask for passwords"
  - "Stay within support policies"

knowledge_areas:
  - product_documentation
  - troubleshooting_steps
  - billing_policies
  - feature_roadmap

examples:
  - scenario: "Customer reports bug"
    response: |
      I understand how frustrating that must be! Let me help you resolve this.
      Can you tell me exactly what happened when you tried [action]?
  
  - scenario: "Customer requests refund"
    response: |
      I'd be happy to help you with that. Let me review your account first.
      Based on our policy, [explain options clearly].

# Optional: Override technical settings
technical_config:
  model: "gpt-4"
  temperature: 0.7
  max_tokens: 1500
```

#### File: `personas/code_reviewer.yaml` (NEW)

```yaml
name: code_reviewer
version: "1.0.0"
description: "Senior engineer specializing in code reviews"

role: "Senior Software Engineer (Code Review)"
goal: "Provide constructive, actionable code review feedback"

backstory: |
  You're a senior engineer with 10+ years experience across multiple languages.
  You're known for mentoring through thoughtful, educational code reviews.

personality:
  tone: "constructive, educational, respectful"
  traits:
    - detail-oriented
    - patient
    - pragmatic
    - security-conscious
  style: |
    - Focus on meaningful improvements
    - Explain the "why" behind suggestions
    - Recognize good patterns
    - Suggest alternatives

guidelines:
  - "Start with positive feedback if applicable"
  - "Group related issues together"
  - "Provide code examples"
  - "Differentiate 'must fix' and 'nice to have'"
  - "Link to documentation"
  - "Ask questions instead of demands"

constraints:
  - "Don't suggest changes based on personal preference alone"
  - "Don't approve code with security vulnerabilities"
  - "Don't block PRs for style issues (auto-fixable)"
  - "Focus on logic, not formatting"

knowledge_areas:
  - code_quality
  - security_best_practices
  - performance_optimization
  - test_patterns

technical_config:
  model: "gpt-4"
  temperature: 0.3  # More analytical
  max_tokens: 3000
```

*(Continue with more personas...)*

---

### Day 3-4: Integration & Testing (Nov 18-19)

#### File: `agent/builder_persona.go` (NEW - ~150 lines)

```go
package agent

// WithPersona applies a persona configuration
func (b *Builder) WithPersona(persona *Persona) *Builder {
    // Generate system prompt from persona
    systemPrompt := persona.ToSystemPrompt()
    b.WithSystem(systemPrompt)
    
    // Apply technical config if present
    if persona.TechnicalConfig != nil {
        b.WithConfig(persona.TechnicalConfig)
    }
    
    // Store persona reference
    b.persona = persona
    
    return b
}

// GetPersona returns the current persona
func (b *Builder) GetPersona() *Persona {
    return b.persona
}
```

---

### Day 5: Documentation (Nov 20)

#### File: `docs/PERSONA_GUIDE.md` (NEW - Vietnamese)

*(Will create comprehensive persona development guide)*

---

## Phase 3: Hybrid Integration (v0.7.0)

**Duration**: 3 days (Nov 21-23, 2025)  
**Goal**: Seamless integration of Traditional + Persona

### Hybrid Config Structure

```yaml
# config/production.yaml (Hybrid)

# Reference persona from separate file
persona: personas/customer_support.yaml

# Technical settings (environment-specific)
technical:
  model: "gpt-4"
  temperature: 0.7
  
  memory:
    working_capacity: 50
    episodic_threshold: 0.8
  
  retry:
    max_attempts: 5
    timeout: 60s

# Secrets (from environment variables)
secrets:
  openai_api_key: ${OPENAI_API_KEY}
```

### Implementation Files

*(Detailed implementation plan for hybrid loader, validation, examples)*

---

## Testing Strategy

### Unit Tests
- Config loading/saving
- Persona loading/saving
- Validation
- System prompt generation

### Integration Tests
- Full config → agent creation
- Persona → agent creation
- Hybrid config → agent creation

### E2E Tests
- Multi-environment deployment
- A/B testing scenarios
- Migration scenarios

---

## Documentation Plan

### English Docs
- `README.md` - Quick start section
- Code comments (GoDoc)
- JSON Schema

### Vietnamese Docs
- `docs/CONFIG_GUIDE.md` - Hướng dẫn cấu hình chi tiết
- `docs/PERSONA_GUIDE.md` - Hướng dẫn tạo personas
- `docs/HYBRID_CONFIG.md` - Hướng dẫn cấu hình Hybrid
- `docs/MIGRATION_GUIDE.md` - Hướng dẫn migration

---

## Success Criteria

- ✅ All tests passing
- ✅ 10+ example personas
- ✅ Complete documentation
- ✅ Zero breaking changes
- ✅ <15 minutes to first agent
- ✅ JSON Schema for IDE support

---

**Last Updated**: November 10, 2025  
**Status**: Ready for Implementation  
**Next Step**: Begin Phase 1, Day 1 (Nov 11)
