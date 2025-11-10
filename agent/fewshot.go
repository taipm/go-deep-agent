package agent

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// SelectionMode defines how examples are selected for few-shot learning
type SelectionMode string

const (
	// SelectionAll includes all examples (up to MaxExamples)
	SelectionAll SelectionMode = "all"

	// SelectionRandom randomly samples examples
	SelectionRandom SelectionMode = "random"

	// SelectionRecent uses most recently created examples
	SelectionRecent SelectionMode = "recent"

	// SelectionBest uses highest quality examples
	SelectionBest SelectionMode = "best"

	// SelectionSimilar uses semantic similarity (Phase 2)
	// Reserved for future implementation with embeddings
	SelectionSimilar SelectionMode = "similar"
)

// FewShotExample represents a single training example for few-shot learning.
// Examples guide the LLM's behavior through demonstration.
type FewShotExample struct {
	// ID uniquely identifies this example (optional, auto-generated if empty)
	ID string `yaml:"id,omitempty" json:"id,omitempty"`

	// Input is the example input/query
	Input string `yaml:"input" json:"input"`

	// Output is the expected/desired output
	Output string `yaml:"output" json:"output"`

	// Quality score (0.0 to 1.0) indicating example effectiveness
	// Higher quality examples may be prioritized in selection
	Quality float64 `yaml:"quality,omitempty" json:"quality,omitempty"`

	// Tags for categorization and filtering
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// Context stores additional metadata (flexible map)
	Context map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"`

	// CreatedAt timestamp (auto-set if not provided)
	CreatedAt time.Time `yaml:"created_at,omitempty" json:"created_at,omitempty"`
}

// Validate checks if the example is valid
func (e *FewShotExample) Validate() error {
	if strings.TrimSpace(e.Input) == "" {
		return errors.New("fewshot example: input cannot be empty")
	}
	if strings.TrimSpace(e.Output) == "" {
		return errors.New("fewshot example: output cannot be empty")
	}
	if e.Quality < 0.0 || e.Quality > 1.0 {
		return fmt.Errorf("fewshot example: quality must be between 0.0 and 1.0, got %.2f", e.Quality)
	}
	return nil
}

// String returns a string representation of the example
func (e *FewShotExample) String() string {
	quality := ""
	if e.Quality > 0 {
		quality = fmt.Sprintf(" [Quality: %.2f]", e.Quality)
	}
	tags := ""
	if len(e.Tags) > 0 {
		tags = fmt.Sprintf(" [Tags: %s]", strings.Join(e.Tags, ", "))
	}
	return fmt.Sprintf("Example%s%s:\n  Input: %s\n  Output: %s",
		quality, tags, e.Input, e.Output)
}

// IsValid returns true if the example is valid (convenience method)
func (e *FewShotExample) IsValid() bool {
	return e.Validate() == nil
}

// HasTag checks if the example has a specific tag
func (e *FewShotExample) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// SetDefaults sets default values for optional fields
func (e *FewShotExample) SetDefaults() {
	if e.ID == "" {
		// Generate simple ID from timestamp and input hash
		e.ID = fmt.Sprintf("ex_%d", time.Now().UnixNano())
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	if e.Quality == 0 {
		e.Quality = 1.0 // Default to highest quality
	}
}

// FewShotConfig configures few-shot learning behavior
type FewShotConfig struct {
	// Examples is the list of training examples
	Examples []FewShotExample `yaml:"examples" json:"examples"`

	// MaxExamples limits how many examples to include in the prompt
	// Default: 5 (0 = unlimited, but model context limits still apply)
	MaxExamples int `yaml:"max_examples,omitempty" json:"max_examples,omitempty"`

	// SelectionMode determines how examples are chosen
	// Default: SelectionAll
	SelectionMode SelectionMode `yaml:"selection_mode,omitempty" json:"selection_mode,omitempty"`

	// PromptTemplate customizes how examples are formatted in the prompt
	// Variables: {{.Examples}} (selected examples)
	// Leave empty for default formatting
	PromptTemplate string `yaml:"prompt_template,omitempty" json:"prompt_template,omitempty"`
}

// Validate checks if the configuration is valid
func (c *FewShotConfig) Validate() error {
	if c == nil {
		return errors.New("fewshot config: config is nil")
	}
	if len(c.Examples) == 0 {
		return errors.New("fewshot config: no examples provided")
	}
	if c.MaxExamples < 0 {
		return fmt.Errorf("fewshot config: max_examples cannot be negative, got %d", c.MaxExamples)
	}

	// Validate each example
	for i, example := range c.Examples {
		if err := example.Validate(); err != nil {
			return fmt.Errorf("fewshot config: example %d invalid: %w", i, err)
		}
	}

	// Validate selection mode
	validModes := map[SelectionMode]bool{
		SelectionAll:     true,
		SelectionRandom:  true,
		SelectionRecent:  true,
		SelectionBest:    true,
		SelectionSimilar: true, // Reserved for Phase 2
	}
	if c.SelectionMode != "" && !validModes[c.SelectionMode] {
		return fmt.Errorf("fewshot config: invalid selection mode '%s'", c.SelectionMode)
	}

	return nil
}

// SetDefaults sets default values for optional fields
func (c *FewShotConfig) SetDefaults() {
	if c.MaxExamples == 0 {
		c.MaxExamples = 5 // Default to 5 examples
	}
	if c.SelectionMode == "" {
		c.SelectionMode = SelectionAll
	}

	// Set defaults for each example
	for i := range c.Examples {
		c.Examples[i].SetDefaults()
	}
}

// ToPrompt converts selected examples into a formatted prompt string
func (c *FewShotConfig) ToPrompt() string {
	if c == nil || len(c.Examples) == 0 {
		return ""
	}

	// Use custom template if provided
	if c.PromptTemplate != "" {
		// Simple variable replacement (Phase 1: basic implementation)
		// Phase 2+ could use text/template for more sophistication
		return c.formatWithTemplate()
	}

	// Default formatting
	var builder strings.Builder
	builder.WriteString("Here are examples of expected behavior:\n\n")

	// Get selected examples (will implement selection in Task 1.3)
	selected := c.SelectExamples()

	for i, example := range selected {
		builder.WriteString(fmt.Sprintf("Example %d:\n", i+1))
		builder.WriteString(fmt.Sprintf("  Input: %s\n", example.Input))
		builder.WriteString(fmt.Sprintf("  Output: %s\n", example.Output))
		builder.WriteString("\n")
	}

	return builder.String()
}

// formatWithTemplate applies custom template (basic implementation)
func (c *FewShotConfig) formatWithTemplate() string {
	// Phase 1: Simple {{.Examples}} replacement
	// Phase 2+: Use text/template for more power
	template := c.PromptTemplate

	var examplesStr strings.Builder
	selected := c.SelectExamples()
	for i, example := range selected {
		examplesStr.WriteString(fmt.Sprintf("Example %d: %s → %s\n", i+1, example.Input, example.Output))
	}

	// Simple string replacement
	result := strings.ReplaceAll(template, "{{.Examples}}", examplesStr.String())
	return result
}

// SelectExamples returns the examples to use based on selection mode
func (c *FewShotConfig) SelectExamples() []FewShotExample {
	if c == nil || len(c.Examples) == 0 {
		return nil
	}

	// Make a copy to avoid modifying original
	examples := make([]FewShotExample, len(c.Examples))
	copy(examples, c.Examples)

	// Apply selection mode
	switch c.SelectionMode {
	case SelectionAll, "": // Default: all examples
		examples = c.selectAll(examples)

	case SelectionRandom:
		examples = c.selectRandom(examples)

	case SelectionRecent:
		examples = c.selectRecent(examples)

	case SelectionBest:
		examples = c.selectBest(examples)

	case SelectionSimilar:
		// Phase 2: Semantic similarity
		// For now, fall back to "all"
		examples = c.selectAll(examples)

	default:
		// Unknown mode, use "all"
		examples = c.selectAll(examples)
	}

	// Apply MaxExamples limit
	maxExamples := c.MaxExamples
	if maxExamples == 0 || maxExamples > len(examples) {
		maxExamples = len(examples)
	}

	return examples[:maxExamples]
}

// selectAll returns all examples (no special ordering)
func (c *FewShotConfig) selectAll(examples []FewShotExample) []FewShotExample {
	return examples
}

// selectRandom randomly shuffles examples
func (c *FewShotConfig) selectRandom(examples []FewShotExample) []FewShotExample {
	// Fisher-Yates shuffle
	for i := len(examples) - 1; i > 0; i-- {
		// Simple pseudo-random based on nanosecond time
		// Good enough for Phase 1; Phase 2+ could use math/rand with seed
		j := int(time.Now().UnixNano() % int64(i+1))
		examples[i], examples[j] = examples[j], examples[i]
	}
	return examples
}

// selectRecent sorts by CreatedAt (most recent first)
func (c *FewShotConfig) selectRecent(examples []FewShotExample) []FewShotExample {
	// Bubble sort (simple, good for small N in Phase 1)
	// Phase 2+ could use sort.Slice for performance
	for i := 0; i < len(examples)-1; i++ {
		for j := 0; j < len(examples)-i-1; j++ {
			if examples[j].CreatedAt.Before(examples[j+1].CreatedAt) {
				examples[j], examples[j+1] = examples[j+1], examples[j]
			}
		}
	}
	return examples
}

// selectBest sorts by Quality (highest first)
func (c *FewShotConfig) selectBest(examples []FewShotExample) []FewShotExample {
	// Bubble sort by quality (descending)
	for i := 0; i < len(examples)-1; i++ {
		for j := 0; j < len(examples)-i-1; j++ {
			if examples[j].Quality < examples[j+1].Quality {
				examples[j], examples[j+1] = examples[j+1], examples[j]
			}
		}
	}
	return examples
}

// Count returns the number of examples that would be selected
func (c *FewShotConfig) Count() int {
	if c == nil {
		return 0
	}
	selected := c.SelectExamples()
	return len(selected)
}
