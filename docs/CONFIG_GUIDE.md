# Configuration Guide

**go-deep-agent** supports YAML-based configuration for easy agent setup and deployment. This guide covers everything you need to know about configuring your agents.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration File Structure](#configuration-file-structure)
- [Model Configuration](#model-configuration)
- [Memory Configuration](#memory-configuration)
- [Retry Configuration](#retry-configuration)
- [Tools Configuration](#tools-configuration)
- [Loading Configuration](#loading-configuration)
- [Environment Variable Overrides](#environment-variable-overrides)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

**1. Create a configuration file** (`config.yaml`):

```yaml
model: "gpt-4o-mini"
temperature: 0.7
max_tokens: 2000

memory:
  working_capacity: 20
  episodic_enabled: true

retry:
  max_attempts: 3
  timeout: 30s
```

**2. Load and use the configuration**:

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    // Load configuration
    config, err := agent.LoadAgentConfig("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create agent with configuration
    apiKey := os.Getenv("OPENAI_API_KEY")
    myAgent := agent.NewOpenAI("", apiKey).
        WithAgentConfig(config)
    
    // Use the agent
    response, _ := myAgent.Ask(context.Background(), "Hello!")
    println(response)
}
```

---

## Configuration File Structure

A complete configuration file has four main sections:

```yaml
# Model settings (required)
model: "gpt-4o-mini"
temperature: 0.7
max_tokens: 2000
top_p: 1.0
system_prompt: "You are a helpful assistant"

# Memory settings (optional)
memory:
  working_capacity: 20
  episodic_enabled: true
  episodic_threshold: 0.7
  semantic_enabled: false
  auto_compress: true

# Retry settings (optional)
retry:
  max_attempts: 3
  timeout: 30s
  exponential_backoff: true
  backoff_multiplier: 2.0
  initial_delay: 1s
  max_delay: 30s

# Tools settings (optional)
tools:
  parallel_execution: false
  max_workers: 10
  timeout: 30s
```

---

## Model Configuration

### `model` (required)

The LLM model to use.

**Type**: `string`  
**Examples**: `"gpt-4o-mini"`, `"gpt-4"`, `"gpt-4-turbo"`, `"gpt-3.5-turbo"`

```yaml
model: "gpt-4o-mini"
```

### `temperature`

Controls randomness in responses.

**Type**: `number`  
**Range**: `0.0` to `2.0`  
**Default**: `0.7`

- **0.0**: Deterministic, focused, factual
- **0.7**: Balanced creativity and consistency (recommended)
- **1.5+**: Very creative, more random

```yaml
temperature: 0.3  # For factual answers
# temperature: 1.2  # For creative writing
```

### `max_tokens`

Maximum number of tokens to generate in the response.

**Type**: `integer`  
**Range**: `1` to `128000`  
**Default**: `2000`

```yaml
max_tokens: 4000  # For longer responses
```

### `top_p`

Nucleus sampling parameter (alternative to temperature).

**Type**: `number`  
**Range**: `0.0` to `1.0`  
**Default**: `1.0`

```yaml
top_p: 0.9  # Use top 90% probability mass
```

### `system_prompt`

Defines the agent's behavior and personality.

**Type**: `string`  
**Default**: None

```yaml
system_prompt: |
  You are an expert coding assistant.
  Always provide working, tested code examples.
  Explain your reasoning clearly.
```

---

## Memory Configuration

go-deep-agent has a **hierarchical memory system** with three layers:

1. **Working Memory**: Recent conversation (always enabled)
2. **Episodic Memory**: Important long-term memories
3. **Semantic Memory**: Extracted facts (experimental)

### `working_capacity`

Number of recent messages to keep in working memory.

**Type**: `integer`  
**Range**: `1` to `1000`  
**Default**: `20`

```yaml
memory:
  working_capacity: 50  # Keep last 50 messages
```

### `episodic_enabled`

Enable long-term episodic memory for important conversations.

**Type**: `boolean`  
**Default**: `true`

```yaml
memory:
  episodic_enabled: true
```

### `episodic_threshold`

Importance threshold for storing in episodic memory.

**Type**: `number`  
**Range**: `0.0` to `1.0`  
**Default**: `0.7`

Messages with importance score ≥ threshold are stored long-term.

```yaml
memory:
  episodic_threshold: 0.8  # Higher threshold = fewer memories
```

**What gets high importance scores?**
- Personal information (names, emails, preferences)
- Explicit "remember this" requests
- Important decisions or commitments
- Questions about critical topics

### `semantic_enabled`

Enable semantic fact extraction (experimental).

**Type**: `boolean`  
**Default**: `false`

```yaml
memory:
  semantic_enabled: true
```

### `auto_compress`

Automatically compress old memories to save space.

**Type**: `boolean`  
**Default**: `true`

```yaml
memory:
  auto_compress: true
```

---

## Retry Configuration

Configure automatic retry behavior for failed requests.

### `max_attempts`

Maximum number of retry attempts.

**Type**: `integer`  
**Range**: `1` to `10`  
**Default**: `3`

```yaml
retry:
  max_attempts: 5  # Try up to 5 times
```

### `timeout`

Timeout per request attempt.

**Type**: `string` (duration)  
**Format**: `<number><unit>` where unit is `ns`, `us`, `ms`, `s`, `m`, `h`  
**Default**: `"30s"`

```yaml
retry:
  timeout: 60s   # 60 seconds
  # timeout: 1m  # 1 minute
  # timeout: 500ms  # 500 milliseconds
```

### `exponential_backoff`

Use exponential backoff strategy for retries.

**Type**: `boolean`  
**Default**: `true`

```yaml
retry:
  exponential_backoff: true
```

### `backoff_multiplier`

Multiplier for exponential backoff (only used if `exponential_backoff` is true).

**Type**: `number`  
**Range**: `≥ 1.0`  
**Default**: `2.0`

With multiplier `2.0` and initial delay `1s`:
- 1st retry: 1s delay
- 2nd retry: 2s delay
- 3rd retry: 4s delay
- 4th retry: 8s delay

```yaml
retry:
  exponential_backoff: true
  backoff_multiplier: 2.5
```

### `initial_delay`

Initial delay before first retry.

**Type**: `string` (duration)  
**Default**: `"1s"`

```yaml
retry:
  initial_delay: 2s
```

### `max_delay`

Maximum delay between retries (prevents excessive wait times).

**Type**: `string` (duration)  
**Default**: `"30s"`

```yaml
retry:
  max_delay: 60s  # Cap at 60 seconds
```

---

## Tools Configuration

Configure how tools are executed.

### `parallel_execution`

Execute multiple tools in parallel when possible.

**Type**: `boolean`  
**Default**: `false`

```yaml
tools:
  parallel_execution: true  # Enable parallel execution
```

### `max_workers`

Maximum number of parallel tool workers (only used if `parallel_execution` is true).

**Type**: `integer`  
**Range**: `1` to `100`  
**Default**: `10`

```yaml
tools:
  parallel_execution: true
  max_workers: 20  # Run up to 20 tools concurrently
```

### `timeout`

Timeout per tool execution.

**Type**: `string` (duration)  
**Default**: `"30s"`

```yaml
tools:
  timeout: 45s
```

---

## Loading Configuration

### Basic Loading

```go
config, err := agent.LoadAgentConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

myAgent := agent.NewOpenAI("", apiKey).WithAgentConfig(config)
```

### With Environment Variable Overrides

```go
// Load config with env var overrides
config, err := agent.LoadAgentConfigWithEnvOverrides("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

Supported environment variables:
- `AGENT_MODEL`: Override model name
- `AGENT_TEMPERATURE`: Override temperature
- `AGENT_MAX_TOKENS`: Override max tokens
- `AGENT_MEMORY_CAPACITY`: Override working memory capacity

### Saving Configuration

```go
// Export current agent configuration
config := myAgent.ToAgentConfig()

// Save to file
err := agent.SaveAgentConfig(config, "exported_config.yaml")
if err != nil {
    log.Fatal(err)
}
```

---

## Environment Variable Overrides

You can override specific config values using environment variables:

**Example**:

```bash
# config.yaml has model: "gpt-4o-mini"
export AGENT_MODEL="gpt-4-turbo"
export AGENT_TEMPERATURE="0.3"
export AGENT_MAX_TOKENS="4000"
export AGENT_MEMORY_CAPACITY="50"

go run main.go  # Uses overridden values
```

**Precedence** (highest to lowest):
1. Environment variables
2. YAML file values
3. Default values

---

## Examples

### Example 1: Customer Support Agent

```yaml
model: "gpt-4o-mini"
temperature: 0.7
max_tokens: 1500

system_prompt: |
  You are a friendly customer support specialist.
  Always be empathetic and solution-oriented.
  Ask clarifying questions when needed.

memory:
  working_capacity: 30
  episodic_enabled: true
  episodic_threshold: 0.8  # Remember customer issues

retry:
  max_attempts: 5
  timeout: 60s
  exponential_backoff: true
```

### Example 2: Code Review Agent

```yaml
model: "gpt-4-turbo"
temperature: 0.3  # More analytical
max_tokens: 4000

system_prompt: |
  You are a senior software engineer performing code reviews.
  Focus on security, performance, and best practices.
  Provide constructive, actionable feedback.

memory:
  working_capacity: 50
  episodic_enabled: true
  episodic_threshold: 0.7

retry:
  max_attempts: 3
  timeout: 45s
```

### Example 3: Parallel Tool Execution

```yaml
model: "gpt-4o-mini"
temperature: 0.7

tools:
  parallel_execution: true
  max_workers: 20
  timeout: 30s

retry:
  max_attempts: 5
  timeout: 60s
  exponential_backoff: true
  backoff_multiplier: 2.0
```

### Example 4: Production Deployment

```yaml
model: "gpt-4o-mini"
temperature: 0.5
max_tokens: 3000
top_p: 0.95

memory:
  working_capacity: 100
  episodic_enabled: true
  episodic_threshold: 0.75
  auto_compress: true

retry:
  max_attempts: 5
  timeout: 120s
  exponential_backoff: true
  backoff_multiplier: 2.5
  initial_delay: 2s
  max_delay: 60s

tools:
  parallel_execution: true
  max_workers: 30
  timeout: 60s
```

---

## Best Practices

### 1. **Start with Defaults, Then Tune**

Begin with the default configuration and adjust based on your needs:

```yaml
model: "gpt-4o-mini"
temperature: 0.7  # Default is good for most cases
```

### 2. **Match Temperature to Use Case**

| Use Case | Temperature | Reason |
|----------|-------------|---------|
| Factual Q&A | 0.0 - 0.3 | Deterministic, accurate |
| Chat assistant | 0.5 - 0.8 | Balanced |
| Creative writing | 1.0 - 1.5 | More diverse outputs |

### 3. **Configure Memory Based on Conversation Length**

- **Short sessions** (Q&A): `working_capacity: 10-20`
- **Medium sessions** (support): `working_capacity: 30-50`
- **Long sessions** (therapy, coaching): `working_capacity: 50-100`

### 4. **Use Environment Variables for Secrets**

Never hardcode API keys in config files:

```go
apiKey := os.Getenv("OPENAI_API_KEY")
myAgent := agent.NewOpenAI("", apiKey).WithAgentConfig(config)
```

### 5. **Enable Retries in Production**

Always enable retries for production deployments:

```yaml
retry:
  max_attempts: 5
  timeout: 60s
  exponential_backoff: true
```

### 6. **Use Parallel Tools for Performance**

If your agent uses multiple tools, enable parallel execution:

```yaml
tools:
  parallel_execution: true
  max_workers: 20
```

### 7. **Version Control Your Configs**

Store configuration files in version control:

```
config/
  development.yaml
  staging.yaml
  production.yaml
```

### 8. **Validate Before Deployment**

```go
config, err := agent.LoadAgentConfig("production.yaml")
if err != nil {
    log.Fatal("Invalid config:", err)
}

// Validation happens automatically
```

---

## Troubleshooting

### Error: "model is required"

**Cause**: Missing `model` field in configuration.

**Solution**: Add model to your config:

```yaml
model: "gpt-4o-mini"
```

### Error: "temperature must be between 0 and 2"

**Cause**: Temperature value is out of range.

**Solution**: Use a value between 0.0 and 2.0:

```yaml
temperature: 0.7
```

### Error: "failed to read config file"

**Cause**: Config file not found or no read permission.

**Solution**:
- Check file path is correct
- Verify file exists: `ls -l config.yaml`
- Check read permissions: `chmod 644 config.yaml`

### Error: "failed to parse YAML"

**Cause**: Invalid YAML syntax.

**Solution**: Validate your YAML:
- Check indentation (use spaces, not tabs)
- Validate online: https://www.yamllint.com/
- Use JSON Schema validation in VS Code

### Error: "retry.max_delay must be >= initial_delay"

**Cause**: `max_delay` is less than `initial_delay`.

**Solution**:

```yaml
retry:
  initial_delay: 1s
  max_delay: 30s  # Must be >= initial_delay
```

### Agent Not Using Config Values

**Cause**: Loading config after setting other values.

**Solution**: Call `WithAgentConfig()` before other builder methods:

```go
// ✅ Correct
myAgent := agent.NewOpenAI("", apiKey).
    WithAgentConfig(config).
    WithSystem("Additional prompt")

// ❌ Wrong - config may override system prompt
myAgent := agent.NewOpenAI("", apiKey).
    WithSystem("My prompt").
    WithAgentConfig(config)  // This may override above
```

---

## IDE Support

For IDE autocomplete and validation, use the JSON Schema:

**VS Code**: Install YAML extension and add to settings:

```json
{
  "yaml.schemas": {
    "./config/schema.json": "*.yaml"
  }
}
```

This enables:
- ✅ Auto-completion
- ✅ Inline documentation
- ✅ Real-time validation
- ✅ Error highlighting

---

## Next Steps

- **[Memory Architecture](MEMORY_ARCHITECTURE.md)**: Deep dive into hierarchical memory
- **[Examples](../examples/)**: Complete working examples
- **[API Reference](https://pkg.go.dev/github.com/taipm/go-deep-agent)**: Full API documentation

---

**Questions?** Open an issue on [GitHub](https://github.com/taipm/go-deep-agent/issues)
