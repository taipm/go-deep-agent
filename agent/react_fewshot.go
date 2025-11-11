package agent

import (
	"fmt"
	"strings"
)

// ReActExample represents a few-shot example for ReAct pattern.
// Examples help the LLM understand the expected format and reasoning style.
type ReActExample struct {
	Task        string   // The task/question
	Steps       []string // The reasoning steps (THOUGHT, ACTION, OBSERVATION, FINAL)
	Description string   // Optional description of what this example demonstrates
}

// ReActExampleSet is a collection of related examples for a specific domain.
type ReActExampleSet struct {
	Name        string         // Name of the example set (e.g., "search", "calculation")
	Description string         // Description of when to use this set
	Examples    []ReActExample // The examples in this set
}

// PredefinedExampleSets contains domain-specific example sets.
var PredefinedExampleSets = map[string]*ReActExampleSet{
	"search": {
		Name:        "search",
		Description: "Examples for web search and information retrieval tasks",
		Examples: []ReActExample{
			{
				Task: "What is the capital of France?",
				Steps: []string{
					`THOUGHT: I need to search for information about France's capital.`,
					`ACTION: search(query="capital of France")`,
					`OBSERVATION: Paris is the capital and largest city of France.`,
					`FINAL: The capital of France is Paris.`,
				},
				Description: "Simple fact lookup",
			},
		},
	},
	"calculation": {
		Name:        "calculation",
		Description: "Examples for mathematical calculations and problem solving",
		Examples: []ReActExample{
			{
				Task: "What is 25% of 80?",
				Steps: []string{
					`THOUGHT: I need to calculate 25 percent of 80. This is 0.25 * 80.`,
					`ACTION: calculator(expression="0.25 * 80")`,
					`OBSERVATION: 20`,
					`FINAL: 25% of 80 is 20.`,
				},
				Description: "Percentage calculation",
			},
			{
				Task: "Solve: 2x + 5 = 15",
				Steps: []string{
					`THOUGHT: I need to solve for x. First, subtract 5 from both sides: 2x = 10.`,
					`ACTION: calculator(expression="15 - 5")`,
					`OBSERVATION: 10`,
					`THOUGHT: Now divide both sides by 2: x = 10/2.`,
					`ACTION: calculator(expression="10 / 2")`,
					`OBSERVATION: 5`,
					`FINAL: x = 5`,
				},
				Description: "Multi-step equation solving",
			},
		},
	},
	"research": {
		Name:        "research",
		Description: "Examples for research tasks requiring multiple sources",
		Examples: []ReActExample{
			{
				Task: "Compare the populations of Tokyo and New York",
				Steps: []string{
					`THOUGHT: I need to find the population of both cities and compare them.`,
					`ACTION: search(query="Tokyo population 2024")`,
					`OBSERVATION: Tokyo has a population of approximately 37 million in the metropolitan area.`,
					`THOUGHT: Now I need to search for New York's population.`,
					`ACTION: search(query="New York population 2024")`,
					`OBSERVATION: New York has a population of approximately 20 million in the metropolitan area.`,
					`FINAL: Tokyo has a larger population (37 million) compared to New York (20 million). Tokyo is approximately 1.85 times more populous than New York.`,
				},
				Description: "Multi-source comparison",
			},
		},
	},
}

// FormatExamples formats a list of examples into a string for the system prompt.
func FormatExamples(examples []ReActExample) string {
	if len(examples) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "Here are some examples to guide your reasoning:")
	parts = append(parts, "")

	for i, example := range examples {
		parts = append(parts, fmt.Sprintf("Example %d:", i+1))
		parts = append(parts, fmt.Sprintf("Task: %s", example.Task))
		parts = append(parts, "")

		for _, step := range example.Steps {
			parts = append(parts, step)
		}

		parts = append(parts, "")
	}

	return strings.Join(parts, "\n")
}

// WithReActExamples adds few-shot examples to the ReAct configuration.
// Examples help the LLM understand the expected format and reasoning style.
//
// Usage:
//
//	// Use predefined example set
//	agent := agent.NewOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActExamples("calculation")
//
//	// Use custom examples
//	examples := []agent.ReActExample{
//	    {
//	        Task: "What is 2+2?",
//	        Steps: []string{
//	            `THOUGHT: I need to add 2 and 2.`,
//	            `ACTION: calculator(expression="2+2")`,
//	            `OBSERVATION: 4`,
//	            `FINAL: The answer is 4.`,
//	        },
//	    },
//	}
//	agent := agent.NewOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActExamples(examples)
func (b *Builder) WithReActExamples(examples interface{}) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}

	switch v := examples.(type) {
	case string:
		// Load predefined example set by name
		if exampleSet, ok := PredefinedExampleSets[v]; ok {
			b.reactConfig.Examples = exampleSet.Examples
		} else {
			// Invalid example set name - ignore
			return b
		}
	case []ReActExample:
		// Use custom examples
		b.reactConfig.Examples = v
	case ReActExample:
		// Single example
		b.reactConfig.Examples = []ReActExample{v}
	default:
		// Unsupported type - ignore
		return b
	}

	return b
}

// WithReActExampleSet adds a complete example set to the configuration.
// This is useful when you want to use a predefined set with all its examples.
func (b *Builder) WithReActExampleSet(setName string) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}

	if exampleSet, ok := PredefinedExampleSets[setName]; ok {
		b.reactConfig.Examples = exampleSet.Examples
	}

	return b
}

// GetAvailableExampleSets returns the names of all predefined example sets.
func GetAvailableExampleSets() []string {
	sets := make([]string, 0, len(PredefinedExampleSets))
	for name := range PredefinedExampleSets {
		sets = append(sets, name)
	}
	return sets
}

// ValidateExample checks if a ReActExample is well-formed.
func ValidateExample(example ReActExample) error {
	if example.Task == "" {
		return fmt.Errorf("example task cannot be empty")
	}

	if len(example.Steps) == 0 {
		return fmt.Errorf("example must have at least one step")
	}

	return validateExampleSteps(example.Steps)
}

// validateExampleSteps checks if steps contain valid ReAct keywords and structure.
func validateExampleSteps(steps []string) error {
	validKeywords := []string{"THOUGHT:", "ACTION:", "OBSERVATION:", "FINAL:"}
	hasThought := false
	hasFinal := false

	for _, step := range steps {
		keyword, err := checkStepKeyword(step, validKeywords)
		if err != nil {
			return err
		}

		if keyword == "THOUGHT:" {
			hasThought = true
		}
		if keyword == "FINAL:" {
			hasFinal = true
		}
	}

	return validateExampleStructure(hasThought, hasFinal)
}

// checkStepKeyword validates that a step starts with a valid keyword.
func checkStepKeyword(step string, validKeywords []string) (string, error) {
	step = strings.TrimSpace(step)
	stepUpper := strings.ToUpper(step)

	for _, keyword := range validKeywords {
		if strings.HasPrefix(stepUpper, keyword) {
			return keyword, nil
		}
	}

	return "", fmt.Errorf("step does not start with valid keyword (THOUGHT, ACTION, OBSERVATION, FINAL): %s", step)
}

// validateExampleStructure checks that example has required THOUGHT and FINAL steps.
func validateExampleStructure(hasThought, hasFinal bool) error {
	if !hasThought {
		return fmt.Errorf("example must contain at least one THOUGHT step")
	}

	if !hasFinal {
		return fmt.Errorf("example must end with a FINAL step")
	}

	return nil
}

// ValidateExamples validates a list of examples.
func ValidateExamples(examples []ReActExample) error {
	for i, example := range examples {
		if err := ValidateExample(example); err != nil {
			return fmt.Errorf("example %d: %w", i+1, err)
		}
	}
	return nil
}
