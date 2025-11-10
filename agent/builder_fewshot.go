package agent

// WithFewShotExamples adds examples for few-shot learning.
// Examples guide the LLM's behavior through demonstration.
//
// The examples will be included in the system prompt before user messages,
// helping the model understand the expected input/output patterns.
//
// Example:
//
//	examples := []agent.FewShotExample{
//	    {Input: "Translate: Hello", Output: "Bonjour", Quality: 1.0},
//	    {Input: "Translate: Goodbye", Output: "Au revoir", Quality: 1.0},
//	}
//	translator := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithFewShotExamples(examples).
//	    Ask(ctx, "Translate: Good morning")
func (b *Builder) WithFewShotExamples(examples []FewShotExample) *Builder {
	if b.fewshotConfig == nil {
		b.fewshotConfig = &FewShotConfig{}
	}
	b.fewshotConfig.Examples = examples
	b.fewshotConfig.SetDefaults()
	return b
}

// WithFewShotConfig applies a complete few-shot configuration.
// This allows full control over example selection, formatting, and behavior.
//
// Example:
//
//	config := &agent.FewShotConfig{
//	    Examples: examples,
//	    MaxExamples: 3,
//	    SelectionMode: agent.SelectionBest,
//	}
//	agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithFewShotConfig(config)
func (b *Builder) WithFewShotConfig(config *FewShotConfig) *Builder {
	b.fewshotConfig = config
	if b.fewshotConfig != nil {
		b.fewshotConfig.SetDefaults()
	}
	return b
}

// AddFewShotExample adds a single example (convenience method).
// This is useful for inline example definition with method chaining.
//
// Example:
//
//	translator := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    AddFewShotExample("Translate: Hello", "Bonjour").
//	    AddFewShotExample("Translate: Goodbye", "Au revoir").
//	    AddFewShotExample("Translate: Thank you", "Merci")
func (b *Builder) AddFewShotExample(input, output string) *Builder {
	if b.fewshotConfig == nil {
		b.fewshotConfig = &FewShotConfig{
			Examples: []FewShotExample{},
		}
	}

	example := FewShotExample{
		Input:  input,
		Output: output,
	}
	example.SetDefaults()

	b.fewshotConfig.Examples = append(b.fewshotConfig.Examples, example)
	b.fewshotConfig.SetDefaults()
	return b
}

// AddFewShotExampleWithQuality adds an example with explicit quality score.
// Quality (0.0-1.0) affects selection when using SelectionBest mode.
//
// Example:
//
//	agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    AddFewShotExampleWithQuality("Good example", "Perfect output", 1.0).
//	    AddFewShotExampleWithQuality("OK example", "OK output", 0.7)
func (b *Builder) AddFewShotExampleWithQuality(input, output string, quality float64) *Builder {
	if b.fewshotConfig == nil {
		b.fewshotConfig = &FewShotConfig{
			Examples: []FewShotExample{},
		}
	}

	example := FewShotExample{
		Input:   input,
		Output:  output,
		Quality: quality,
	}
	example.SetDefaults()

	b.fewshotConfig.Examples = append(b.fewshotConfig.Examples, example)
	b.fewshotConfig.SetDefaults()
	return b
}

// WithFewShotSelection sets the example selection strategy.
//
// Available modes:
//   - SelectionAll: Use all examples (default)
//   - SelectionRandom: Randomly sample examples
//   - SelectionRecent: Use most recent examples
//   - SelectionBest: Use highest quality examples
//
// Example:
//
//	agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithFewShotExamples(manyExamples).
//	    WithFewShotSelection(agent.SelectionBest). // Use only best examples
//	    WithFewShotMaxExamples(3)                   // Limit to 3
func (b *Builder) WithFewShotSelection(mode SelectionMode) *Builder {
	if b.fewshotConfig == nil {
		b.fewshotConfig = &FewShotConfig{}
	}
	b.fewshotConfig.SelectionMode = mode
	return b
}

// WithFewShotMaxExamples limits the number of examples included in prompts.
// This helps manage token usage while still benefiting from few-shot learning.
//
// Example:
//
//	// Have 100 examples, but only use best 5
//	agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithFewShotExamples(hundredExamples).
//	    WithFewShotMaxExamples(5)
func (b *Builder) WithFewShotMaxExamples(max int) *Builder {
	if b.fewshotConfig == nil {
		b.fewshotConfig = &FewShotConfig{}
	}
	b.fewshotConfig.MaxExamples = max
	return b
}

// GetFewShotExamples returns the current few-shot examples.
// Returns nil if no examples are configured.
//
// Example:
//
//	examples := builder.GetFewShotExamples()
//	fmt.Printf("Using %d examples\n", len(examples))
func (b *Builder) GetFewShotExamples() []FewShotExample {
	if b.fewshotConfig == nil {
		return nil
	}
	return b.fewshotConfig.Examples
}

// GetFewShotConfig returns the current few-shot configuration.
// Returns nil if few-shot learning is not enabled.
//
// Example:
//
//	if config := builder.GetFewShotConfig(); config != nil {
//	    fmt.Printf("Selection mode: %s\n", config.SelectionMode)
//	    fmt.Printf("Will use %d examples\n", config.Count())
//	}
func (b *Builder) GetFewShotConfig() *FewShotConfig {
	return b.fewshotConfig
}

// ClearFewShotExamples removes all few-shot examples.
// Useful when switching tasks or resetting agent behavior.
//
// Example:
//
//	// Use examples for first task
//	agent.WithFewShotExamples(translationExamples)
//	response1, _ := agent.Ask(ctx, "Translate: Hello")
//
//	// Clear and use different examples for second task
//	agent.ClearFewShotExamples().
//	    WithFewShotExamples(codeExamples)
//	response2, _ := agent.Ask(ctx, "Write Go function")
func (b *Builder) ClearFewShotExamples() *Builder {
	b.fewshotConfig = nil
	return b
}

// HasFewShotExamples returns true if few-shot examples are configured.
//
// Example:
//
//	if builder.HasFewShotExamples() {
//	    fmt.Println("Using few-shot learning")
//	}
func (b *Builder) HasFewShotExamples() bool {
	return b.fewshotConfig != nil && len(b.fewshotConfig.Examples) > 0
}
