# Few-Shot Learning Guide

**Version**: v0.7.0 (Phase 1 - Static Examples)

Few-shot learning allows you to provide example input-output pairs to guide your AI agent's behavior, improving consistency and quality without fine-tuning.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Builder API Reference](#builder-api-reference)
- [Selection Modes](#selection-modes)
- [YAML Persona Integration](#yaml-persona-integration)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Migration Guide](#migration-guide)

## Introduction

### What is Few-Shot Learning?

Few-shot learning is a technique where you provide a few examples to demonstrate the desired behavior:

```
System: You are a translator.

Example 1:
User: Translate: Hello
Assistant: Bonjour

Example 2:
User: Translate: Goodbye
Assistant: Au revoir

Current Request:
User: Translate: Good morning
Assistant: [LLM generates following the pattern]
```

### Benefits

- **Improved Consistency**: Examples guide the model's output format
- **Better Quality**: Shows the exact style and tone you want
- **No Fine-Tuning**: Works with any LLM without training
- **Reusable**: Store examples in YAML personas for reuse

## Quick Start

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    ai := agent.NewOpenAI("gpt-4o-mini", "your-api-key").
        WithSystem("You are a French translator.").
        AddFewShotExample("Translate: Hello", "Bonjour").
        AddFewShotExample("Translate: Goodbye", "Au revoir").
        AddFewShotExample("Translate: Thank you", "Merci")

    response, _ := ai.Ask(context.Background(), "Translate: Good morning")
    fmt.Println(response) // Expected: Bonjour (le matin) or similar
}
```

### Array-Based Examples

```go
examples := []agent.FewShotExample{
    {Input: "Hello", Output: "Bonjour", Quality: 1.0},
    {Input: "Goodbye", Output: "Au revoir", Quality: 1.0},
    {Input: "Thank you", Output: "Merci", Quality: 0.9},
}

ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a translator.").
    WithFewShotExamples(examples)
```

## Builder API Reference

### 1. AddFewShotExample(input, output string)

Add a single example with default quality (1.0).

```go
ai := agent.NewOpenAI(model, apiKey).
    AddFewShotExample("What is 2+2?", "4").
    AddFewShotExample("What is 3+3?", "6")
```

**Use when**: Adding examples one by one during development.

### 2. AddFewShotExampleWithQuality(input, output string, quality float64)

Add an example with a custom quality score (0.0-1.0).

```go
ai := agent.NewOpenAI(model, apiKey).
    AddFewShotExample("Perfect example", "Perfect output").            // quality: 1.0
    AddFewShotExampleWithQuality("Good example", "Good output", 0.8).  // quality: 0.8
    AddFewShotExampleWithQuality("Okay example", "Okay output", 0.6)   // quality: 0.6
```

**Use when**: You want to prioritize certain examples with `SelectionBest` mode.

### 3. WithFewShotExamples(examples []FewShotExample)

Add multiple examples at once.

```go
examples := []agent.FewShotExample{
    {
        Input:   "Translate: Hello",
        Output:  "Bonjour",
        Quality: 1.0,
        Tags:    []string{"greeting", "basic"},
    },
    {
        Input:   "Translate: Good morning",
        Output:  "Bonjour (le matin)",
        Quality: 0.9,
        Tags:    []string{"greeting", "time-specific"},
    },
}

ai := agent.NewOpenAI(model, apiKey).
    WithFewShotExamples(examples)
```

**Use when**: Loading examples from a database or configuration.

### 4. WithFewShotConfig(config *FewShotConfig)

Apply a complete few-shot configuration.

```go
config := &agent.FewShotConfig{
    Examples:      examples,
    MaxExamples:   3,
    SelectionMode: agent.SelectionBest,
}

ai := agent.NewOpenAI(model, apiKey).
    WithFewShotConfig(config)
```

**Use when**: You need full control over selection behavior.

### 5. WithFewShotSelectionMode(mode SelectionMode)

Set the strategy for selecting examples.

```go
ai := agent.NewOpenAI(model, apiKey).
    WithFewShotExamples(manyExamples).
    WithFewShotSelectionMode(agent.SelectionBest).
    WithFewShotConfig(&agent.FewShotConfig{MaxExamples: 3})
```

**Available modes**: `SelectionAll`, `SelectionRandom`, `SelectionRecent`, `SelectionBest`.

### 6. GetFewShotExamples() []FewShotExample

Retrieve current examples.

```go
ai := agent.NewOpenAI(model, apiKey).
    AddFewShotExample("input1", "output1").
    AddFewShotExample("input2", "output2")

examples := ai.GetBuilder().GetFewShotExamples()
fmt.Printf("Total examples: %d\n", len(examples))
```

**Use when**: Debugging or inspecting current configuration.

### 7. ClearFewShotExamples()

Remove all examples.

```go
ai := agent.NewOpenAI(model, apiKey).
    AddFewShotExample("temp", "example").
    ClearFewShotExamples() // Now empty
```

**Use when**: Resetting configuration between tests.

## Selection Modes

When you have many examples, selection modes determine which ones are used in prompts.

### SelectionAll (Default)

Use all examples up to `MaxExamples`.

```go
config := &agent.FewShotConfig{
    Examples:      tenExamples,
    MaxExamples:   5,
    SelectionMode: agent.SelectionAll, // Use first 5 examples
}
```

**Best for**: Small, curated example sets.

### SelectionRandom

Randomly select examples each time.

```go
ai := agent.NewOpenAI(model, apiKey).
    WithFewShotExamples(manyExamples).
    WithFewShotSelectionMode(agent.SelectionRandom).
    WithFewShotConfig(&agent.FewShotConfig{MaxExamples: 3})
```

**Best for**: Preventing overfitting to specific examples, testing variety.

### SelectionRecent

Use the most recently created examples.

```go
config := &agent.FewShotConfig{
    Examples:      examplesWithTimestamps,
    MaxExamples:   5,
    SelectionMode: agent.SelectionRecent, // Newest 5
}
```

**Best for**: Evolving example sets where recent examples are more relevant.

### SelectionBest

Use examples with highest quality scores.

```go
examples := []agent.FewShotExample{
    {Input: "ex1", Output: "out1", Quality: 1.0},  // Selected
    {Input: "ex2", Output: "out2", Quality: 0.9},  // Selected
    {Input: "ex3", Output: "out3", Quality: 0.8},  // Selected
    {Input: "ex4", Output: "out4", Quality: 0.5},  // Not selected
}

ai := agent.NewOpenAI(model, apiKey).
    WithFewShotExamples(examples).
    WithFewShotSelectionMode(agent.SelectionBest).
    WithFewShotConfig(&agent.FewShotConfig{MaxExamples: 3})
```

**Best for**: Production systems where you want to use only verified, high-quality examples.

### SelectionSimilar (Coming in Phase 2)

Select examples semantically similar to the current query (requires embeddings).

## YAML Persona Integration

### Loading a Persona with Few-Shot Examples

```go
persona, err := agent.LoadPersona("personas/translator_fewshot.yaml")
if err != nil {
    log.Fatal(err)
}

ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithPersona(persona)

// Persona automatically includes:
// - System prompt from role/personality
// - Few-shot examples from fewshot section
// - Technical config (model, temperature, memory, retry)
```

### Example YAML Persona

```yaml
name: "FrenchTranslator"
version: "1.0.0"
role: "French Translator"
goal: "Translate English phrases to French accurately"

personality:
  tone: "precise and consistent"
  traits:
    - "accurate"
    - "clear"
  style: "direct"

# Few-shot examples
fewshot:
  examples:
    - input: "Translate: Hello"
      output: "Bonjour"
      quality: 1.0
      tags: ["greeting", "basic"]
    
    - input: "Translate: Goodbye"
      output: "Au revoir"
      quality: 1.0
      tags: ["greeting"]
    
    - input: "Translate: Thank you"
      output: "Merci"
      quality: 1.0
      tags: ["polite"]
  
  max_examples: 3
  selection_mode: "best"

technical_config:
  model: "gpt-4o-mini"
  temperature: 0.3
  memory:
    working_capacity: 30
    episodic_enabled: true
  retry:
    max_attempts: 3
    initial_delay: "1s"
```

### Benefits of YAML Personas

- **Reusability**: Share personas across applications
- **Version Control**: Track changes to examples and behavior
- **Team Collaboration**: Non-developers can edit YAML files
- **A/B Testing**: Easily swap personas to compare performance

## Use Cases

### 1. Translation Services

```go
ai := agent.NewOpenAI(model, apiKey).
    WithSystem("You are a translator.").
    AddFewShotExample("Translate to French: Hello", "Bonjour").
    AddFewShotExample("Translate to French: Goodbye", "Au revoir").
    AddFewShotExample("Translate to Spanish: Hello", "Hola").
    AddFewShotExample("Translate to Spanish: Goodbye", "Adiós")
```

### 2. Code Generation

```go
ai := agent.NewOpenAI(model, apiKey).
    WithSystem("Generate idiomatic Go code.").
    WithTemperature(0.2).
    AddFewShotExample(
        "Generate error variable for database connection",
        `var ErrDatabaseConnection = errors.New("failed to connect to database")`,
    ).
    AddFewShotExample(
        "Generate HTTP client with timeout",
        `client := &http.Client{
    Timeout: 10 * time.Second,
}`,
    )
```

### 3. Customer Support

```go
ai := agent.NewOpenAI(model, apiKey).
    WithSystem("You are a customer support agent.").
    AddFewShotExample(
        "My order hasn't arrived",
        "I understand your concern. Let me check your order status. Could you provide your order number?",
    ).
    AddFewShotExample(
        "How do I return an item?",
        "Returns are easy! You can initiate a return within 30 days. Visit your Orders page and click 'Return Item'.",
    )
```

### 4. Data Extraction

```go
ai := agent.NewOpenAI(model, apiKey).
    WithSystem("Extract structured data from text.").
    AddFewShotExample(
        "John Smith, age 30, lives in New York",
        `{"name": "John Smith", "age": 30, "city": "New York"}`,
    ).
    AddFewShotExample(
        "Sarah Johnson, 25 years old, from Los Angeles",
        `{"name": "Sarah Johnson", "age": 25, "city": "Los Angeles"}`,
    )
```

## Best Practices

### 1. Example Quality Over Quantity

- **Start small**: 3-5 examples is often enough
- **High quality**: Perfect examples are better than many mediocre ones
- **Representative**: Cover the most common use cases

### 2. Use Quality Scores

```go
// Mark verified, production-ready examples
AddFewShotExampleWithQuality("verified input", "verified output", 1.0)

// Mark experimental examples
AddFewShotExampleWithQuality("experimental input", "experimental output", 0.7)
```

Then use `SelectionBest` to prioritize high-quality examples.

### 3. Tag Your Examples

```go
examples := []agent.FewShotExample{
    {
        Input:  "Hello",
        Output: "Bonjour",
        Tags:   []string{"greeting", "basic", "common"},
    },
    {
        Input:  "Good evening",
        Output: "Bonsoir",
        Tags:   []string{"greeting", "time-specific", "formal"},
    },
}
```

Tags help with:
- Organization and categorization
- Future filtering (Phase 2+)
- Documentation and searchability

### 4. Keep Examples Focused

Bad (too vague):
```go
AddFewShotExample("translate", "translated")
```

Good (specific):
```go
AddFewShotExample("Translate to French: Hello", "Bonjour")
```

### 5. Use YAML for Production

Development:
```go
// Quick prototyping
ai.AddFewShotExample("input", "output")
```

Production:
```yaml
# personas/my_agent.yaml
fewshot:
  examples:
    - input: "input"
      output: "output"
      quality: 1.0
```

### 6. Combine with Memory

```go
ai := agent.NewOpenAI(model, apiKey).
    WithSystem("You are a helpful assistant.").
    WithMemory(10).                              // Remember conversation
    AddFewShotExample("What is 2+2?", "4").      // Example format
    AddFewShotExample("What is 3+3?", "6")

// Now agent has:
// - Examples showing how to answer math questions
// - Memory to remember what user asked before
```

### 7. Test Different Selection Modes

```go
// Development: Use all examples
config.SelectionMode = agent.SelectionAll

// Testing: Random variety
config.SelectionMode = agent.SelectionRandom

// Production: Best quality only
config.SelectionMode = agent.SelectionBest
```

### 8. Monitor Example Effectiveness

```go
examples := ai.GetBuilder().GetFewShotExamples()
for _, ex := range examples {
    fmt.Printf("Example (quality %.1f): %s -> %s\n", 
        ex.Quality, ex.Input, ex.Output)
}
```

## Migration Guide

### From WithMessages() to Few-Shot

**Before** (v0.6.x):
```go
ai := agent.NewOpenAI(model, apiKey).
    WithSystem("You are a translator.").
    WithMessages([]agent.Message{
        {Role: "user", Content: "Translate: Hello"},
        {Role: "assistant", Content: "Bonjour"},
        {Role: "user", Content: "Translate: Goodbye"},
        {Role: "assistant", Content: "Au revoir"},
    })
```

**After** (v0.7.0):
```go
ai := agent.NewOpenAI(model, apiKey).
    WithSystem("You are a translator.").
    AddFewShotExample("Translate: Hello", "Bonjour").
    AddFewShotExample("Translate: Goodbye", "Au revoir")
```

### Benefits of Migration

1. **Clearer Intent**: Code explicitly shows examples vs conversation
2. **Selection Modes**: Can use `SelectionBest`, `SelectionRandom`, etc.
3. **Quality Tracking**: Assign quality scores for prioritization
4. **YAML Support**: Store examples in reusable personas
5. **Future Features**: Phase 2+ will add semantic similarity, learning from feedback

### Backward Compatibility

Both approaches work! Few-shot is recommended for:
- Teaching consistent behavior patterns
- Reusable example sets
- Production systems

Use `WithMessages()` for:
- Actual conversation history
- One-off custom contexts

## What's Next?

### Phase 2 (v0.7.1) - Dynamic Selection
- Semantic similarity selection
- Auto-select examples based on query
- Embedding-based retrieval

### Phase 3 (v0.7.2) - Learning from Feedback
- Collect user feedback on outputs
- Automatically improve example quality
- Persistence to database

### Phase 4 (v0.7.3) - Production Features
- Example clustering and deduplication
- A/B testing different example sets
- Analytics and performance tracking

## Support

- **GitHub Issues**: https://github.com/taipm/go-deep-agent/issues
- **Examples**: See `examples/fewshot_basic/`
- **Schema**: See `personas/schema.json`
- **Tests**: See `agent/fewshot_test.go`

## License

MIT License - See LICENSE file for details.
