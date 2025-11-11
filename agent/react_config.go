package agent

import (
	"fmt"
	"time"
)

// Default configuration values for ReAct pattern
const (
	// DefaultReActMaxIterations is the default maximum number of reasoning loops
	// This prevents infinite loops while allowing for complex multi-step reasoning
	DefaultReActMaxIterations = 5

	// DefaultReActTimeout is the default execution timeout
	// Protects against hanging executions while allowing time for complex tasks
	DefaultReActTimeout = 60 * time.Second

	// DefaultReActStrict is the default strict mode setting
	// false means graceful fallback on parse errors (recommended for production)
	DefaultReActStrict = false
)

// ReActConfig holds configuration for ReAct pattern execution.
// Use the Builder methods to configure these options:
//
//	ai := agent.New().
//	    WithOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActMaxIterations(10).
//	    WithReActTimeout(120 * time.Second).
//	    Build()
type ReActConfig struct {
	// Enabled determines whether ReAct mode is active
	// If false, Execute() will behave like Ask()
	Enabled bool

	// MaxIterations limits the number of reasoning loops
	// Each iteration consists of: THOUGHT → ACTION → OBSERVATION
	// When exceeded, returns ErrMaxIterationsReached
	// Default: 5
	MaxIterations int

	// Timeout for the entire ReAct execution
	// If exceeded, returns partial results with timeout error
	// Default: 60 seconds
	Timeout time.Duration

	// Strict mode controls error handling behavior:
	// - true: fail immediately on parse errors
	// - false: gracefully fallback to normal execution
	// Default: false (graceful)
	Strict bool

	// SystemPrompt override (optional)
	// If empty, uses the default ReAct system prompt
	// Advanced users can customize the prompt format
	// Default: "" (uses built-in prompt)
	SystemPrompt string

	// Callback for execution events (optional)
	// Implement ReActCallback interface to receive notifications
	// Useful for logging, monitoring, and debugging
	// Default: nil (no callbacks)
	Callback ReActCallback

	// EnableMetrics determines whether to collect execution metrics
	// When true, ReActResult.Metrics will be populated
	// Minimal performance overhead
	// Default: false
	EnableMetrics bool

	// EnableTimeline determines whether to track execution timeline
	// When true, ReActResult.Timeline will be populated
	// Useful for debugging and performance analysis
	// Default: false
	EnableTimeline bool

	// Examples are few-shot examples to guide the LLM's reasoning
	// These are included in the system prompt to demonstrate expected format
	// Can be set using WithReActExamples() or WithReActExampleSet()
	// Default: nil (no examples)
	Examples []ReActExample
}

// NewReActConfig creates a new ReActConfig with default values.
func NewReActConfig() *ReActConfig {
	return &ReActConfig{
		Enabled:        false, // Must be explicitly enabled
		MaxIterations:  DefaultReActMaxIterations,
		Timeout:        DefaultReActTimeout,
		Strict:         DefaultReActStrict,
		SystemPrompt:   "", // Use default
		Callback:       nil,
		EnableMetrics:  false,
		EnableTimeline: false,
		Examples:       nil,
	}
}

// Validate checks if the configuration is valid.
// Returns an error if any setting is invalid.
func (c *ReActConfig) Validate() error {
	if c.MaxIterations < 1 {
		return fmt.Errorf("MaxIterations must be >= 1, got %d", c.MaxIterations)
	}

	if c.MaxIterations > 100 {
		return fmt.Errorf("MaxIterations too high (>100), got %d (possible infinite loop)", c.MaxIterations)
	}

	if c.Timeout < 1*time.Second {
		return fmt.Errorf("Timeout must be >= 1s, got %v", c.Timeout)
	}

	if c.Timeout > 10*time.Minute {
		return fmt.Errorf("Timeout too high (>10min), got %v", c.Timeout)
	}

	return nil
}

// Clone creates a deep copy of the configuration.
// This is used internally by the Builder to avoid shared state.
func (c *ReActConfig) Clone() *ReActConfig {
	if c == nil {
		return NewReActConfig()
	}

	// Deep copy Examples slice
	var examples []ReActExample
	if c.Examples != nil {
		examples = make([]ReActExample, len(c.Examples))
		copy(examples, c.Examples)
	}

	return &ReActConfig{
		Enabled:        c.Enabled,
		MaxIterations:  c.MaxIterations,
		Timeout:        c.Timeout,
		Strict:         c.Strict,
		SystemPrompt:   c.SystemPrompt,
		Callback:       c.Callback,
		EnableMetrics:  c.EnableMetrics,
		EnableTimeline: c.EnableTimeline,
		Examples:       examples,
	}
}
