package agent

import (
	"fmt"
	"time"
)

// ReActMode defines the execution mode for ReAct pattern.
// Determines whether to use function calling (native) or text parsing.
type ReActMode string

const (
	// ReActModeNative uses OpenAI's function calling with meta-tools.
	// This is the recommended mode as it's more reliable and language-agnostic.
	// Uses structured output: think(), use_tool(), final_answer()
	// Advantages: no text parsing, fewer errors, better multilingual support
	ReActModeNative ReActMode = "native"

	// ReActModeText uses text parsing with regex patterns (legacy).
	// Expects format: THOUGHT: ... ACTION: tool(args) FINAL: ...
	// This mode is maintained for backward compatibility.
	// Deprecated: Use ReActModeNative for new applications.
	ReActModeText ReActMode = "text"

	// ReActModeHybrid tries native mode first, falls back to text mode.
	// Useful during migration or when dealing with unpredictable LLM behavior.
	// Not implemented yet - reserved for future use.
	ReActModeHybrid ReActMode = "hybrid"
)

// ReActTaskComplexity defines the complexity level of a ReAct task.
// This helps users choose appropriate settings for different task types.
// Use WithReActComplexity() to automatically configure optimal settings.
type ReActTaskComplexity string

const (
	// ReActTaskSimple is for simple tasks: classification, review, yes/no decisions.
	// Recommended settings: MaxIterations=3, Timeout=30s
	// Examples: Document review, sentiment analysis, simple validation
	ReActTaskSimple ReActTaskComplexity = "simple"

	// ReActTaskMedium is for moderate reasoning tasks.
	// Recommended settings: MaxIterations=5, Timeout=60s
	// Examples: Multi-step calculations, data analysis, moderate research
	ReActTaskMedium ReActTaskComplexity = "medium"

	// ReActTaskComplex is for complex reasoning tasks.
	// Recommended settings: MaxIterations=10, Timeout=120s
	// Examples: Deep research, planning, multi-source analysis
	ReActTaskComplex ReActTaskComplexity = "complex"
)

// Recommended settings for each task complexity level
const (
	// Simple task settings
	ReActSimpleMaxIterations = 3
	ReActSimpleTimeout       = 30 * time.Second

	// Medium task settings
	ReActMediumMaxIterations = 5
	ReActMediumTimeout       = 60 * time.Second

	// Complex task settings
	ReActComplexMaxIterations = 10
	ReActComplexTimeout       = 120 * time.Second
)

// Default configuration values for ReAct pattern
const (
	// DefaultReActMaxIterations is the default maximum number of reasoning loops.
	// Changed from 5 to 3 in v0.7.6 for better UX with simple tasks.
	// Most tasks don't need more than 3 iterations. Use WithReActComplexity()
	// or WithReActMaxIterations() to customize for complex tasks.
	DefaultReActMaxIterations = 3

	// DefaultReActTimeout is the default execution timeout
	// Protects against hanging executions while allowing time for complex tasks
	DefaultReActTimeout = 60 * time.Second

	// DefaultReActStrict is the default strict mode setting
	// false means graceful fallback on parse errors (recommended for production)
	DefaultReActStrict = false

	// DefaultReActMode is the default execution mode
	// Native mode is preferred for reliability and language support
	DefaultReActMode = ReActModeNative
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

	// Mode determines the execution strategy
	// ReActModeNative: Use function calling (recommended)
	// ReActModeText: Use text parsing (legacy)
	// ReActModeHybrid: Try native, fallback to text (future)
	// Default: ReActModeNative
	Mode ReActMode

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

	// EnableAutoFallback enables graceful degradation when max iterations reached.
	// When true and final_answer() not called:
	// - Synthesizes answer from completed reasoning steps
	// - Returns success with warning instead of error
	// - Prevents losing all analysis work
	// Default: true (recommended for better UX)
	// Added in v0.7.6
	EnableAutoFallback bool

	// ForceFinalAnswerAtMax forces LLM to provide final_answer() at last iteration.
	// When true, injects special prompt at final iteration requiring conclusion.
	// Prevents timeout errors by ensuring LLM provides answer.
	// Works in conjunction with EnableIterationReminders.
	// Default: true (recommended)
	// Added in v0.7.6
	ForceFinalAnswerAtMax bool

	// EnableIterationReminders adds progressive urgency to system prompts.
	// Reminds LLM to wrap up analysis as max iterations approaches:
	// - At n-2: "You have 2 iterations remaining"
	// - At n-1: "This is your second-to-last iteration"
	// - At n: "FINAL iteration - you MUST call final_answer() now"
	// Reduces analysis paralysis and timeout errors.
	// Default: true (recommended)
	// Added in v0.7.6
	EnableIterationReminders bool

	// Examples are few-shot examples to guide the LLM's reasoning
	// These are included in the system prompt to demonstrate expected format
	// Can be set using WithReActExamples() or WithReActExampleSet()
	// Default: nil (no examples)
	Examples []ReActExample
}

// NewReActConfig creates a new ReActConfig with default values.
func NewReActConfig() *ReActConfig {
	return &ReActConfig{
		Enabled:                  false,            // Must be explicitly enabled
		Mode:                     DefaultReActMode, // Native mode by default
		MaxIterations:            DefaultReActMaxIterations,
		Timeout:                  DefaultReActTimeout,
		Strict:                   DefaultReActStrict,
		SystemPrompt:             "", // Use default
		Callback:                 nil,
		EnableMetrics:            false,
		EnableTimeline:           false,
		EnableAutoFallback:       true, // v0.7.6: Better UX
		ForceFinalAnswerAtMax:    true, // v0.7.6: Prevent timeouts
		EnableIterationReminders: true, // v0.7.6: Guide LLM
		Examples:                 nil,
	}
}

// Validate checks if the configuration is valid.
// Returns an error if any setting is invalid.
func (c *ReActConfig) Validate() error {
	// Validate Mode
	switch c.Mode {
	case ReActModeNative, ReActModeText:
		// Valid modes
	case ReActModeHybrid:
		return fmt.Errorf("ReActModeHybrid not implemented yet")
	case "":
		return fmt.Errorf("Mode cannot be empty, must be one of: native, text")
	default:
		return fmt.Errorf("invalid ReActMode: %q, must be one of: native, text", c.Mode)
	}

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
		Mode:           c.Mode,
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
