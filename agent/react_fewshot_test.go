package agent

import (
	"strings"
	"testing"
)

// TestReActExample_Basic tests basic ReActExample creation
func TestReActExample_Basic(t *testing.T) {
	example := ReActExample{
		Task: "What is 2+2?",
		Steps: []string{
			`THOUGHT: I need to calculate 2+2.`,
			`ACTION: calculator(expression="2+2")`,
			`OBSERVATION: 4`,
			`FINAL: The answer is 4.`,
		},
		Description: "Simple addition",
	}

	if example.Task != "What is 2+2?" {
		t.Errorf("Expected task 'What is 2+2?', got '%s'", example.Task)
	}

	if len(example.Steps) != 4 {
		t.Errorf("Expected 4 steps, got %d", len(example.Steps))
	}
}

// TestValidateExample_Valid tests validation of valid examples
func TestValidateExample_Valid(t *testing.T) {
	tests := []struct {
		name    string
		example ReActExample
	}{
		{
			name: "Simple example",
			example: ReActExample{
				Task: "What is the capital?",
				Steps: []string{
					`THOUGHT: I need to search.`,
					`ACTION: search(query="capital")`,
					`OBSERVATION: Paris`,
					`FINAL: The capital is Paris.`,
				},
			},
		},
		{
			name: "Multi-step example",
			example: ReActExample{
				Task: "Complex task",
				Steps: []string{
					`THOUGHT: First step.`,
					`ACTION: tool1(arg="value")`,
					`OBSERVATION: Result 1`,
					`THOUGHT: Second step.`,
					`ACTION: tool2(arg="value")`,
					`OBSERVATION: Result 2`,
					`FINAL: Done.`,
				},
			},
		},
		{
			name: "Lowercase keywords",
			example: ReActExample{
				Task: "Test case sensitivity",
				Steps: []string{
					`thought: lowercase works`,
					`action: tool(arg="val")`,
					`observation: ok`,
					`final: success`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExample(tt.example)
			if err != nil {
				t.Errorf("ValidateExample() error = %v, want nil", err)
			}
		})
	}
}

// TestValidateExample_Invalid tests validation of invalid examples
func TestValidateExample_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		example     ReActExample
		expectedErr string
	}{
		{
			name: "Empty task",
			example: ReActExample{
				Task: "",
				Steps: []string{
					`THOUGHT: test`,
					`FINAL: done`,
				},
			},
			expectedErr: "task cannot be empty",
		},
		{
			name: "Empty steps",
			example: ReActExample{
				Task:  "Task",
				Steps: []string{},
			},
			expectedErr: "must have at least one step",
		},
		{
			name: "No THOUGHT",
			example: ReActExample{
				Task: "Task",
				Steps: []string{
					`ACTION: tool(arg="val")`,
					`FINAL: done`,
				},
			},
			expectedErr: "must contain at least one THOUGHT",
		},
		{
			name: "No FINAL",
			example: ReActExample{
				Task: "Task",
				Steps: []string{
					`THOUGHT: thinking`,
					`ACTION: tool(arg="val")`,
				},
			},
			expectedErr: "must end with a FINAL",
		},
		{
			name: "Invalid keyword",
			example: ReActExample{
				Task: "Task",
				Steps: []string{
					`THOUGHT: thinking`,
					`INVALID: wrong keyword`,
					`FINAL: done`,
				},
			},
			expectedErr: "does not start with valid keyword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExample(tt.example)
			if err == nil {
				t.Errorf("ValidateExample() error = nil, want error containing '%s'", tt.expectedErr)
				return
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("ValidateExample() error = %v, want error containing '%s'", err, tt.expectedErr)
			}
		})
	}
}

// TestValidateExamples tests validation of example arrays
func TestValidateExamples(t *testing.T) {
	validExample := ReActExample{
		Task: "Valid",
		Steps: []string{
			`THOUGHT: test`,
			`FINAL: done`,
		},
	}

	invalidExample := ReActExample{
		Task:  "Invalid",
		Steps: []string{},
	}

	tests := []struct {
		name        string
		examples    []ReActExample
		shouldError bool
	}{
		{
			name:        "Empty array",
			examples:    []ReActExample{},
			shouldError: false,
		},
		{
			name:        "All valid",
			examples:    []ReActExample{validExample, validExample},
			shouldError: false,
		},
		{
			name:        "Contains invalid",
			examples:    []ReActExample{validExample, invalidExample},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExamples(tt.examples)
			if (err != nil) != tt.shouldError {
				t.Errorf("ValidateExamples() error = %v, shouldError = %v", err, tt.shouldError)
			}
		})
	}
}

// TestFormatExamples tests example formatting for prompts
func TestFormatExamples(t *testing.T) {
	tests := []struct {
		name     string
		examples []ReActExample
		want     string
	}{
		{
			name:     "Empty examples",
			examples: []ReActExample{},
			want:     "",
		},
		{
			name: "Single example",
			examples: []ReActExample{
				{
					Task: "What is 2+2?",
					Steps: []string{
						`THOUGHT: I need to calculate.`,
						`ACTION: calc(expr="2+2")`,
						`OBSERVATION: 4`,
						`FINAL: Answer is 4.`,
					},
				},
			},
			want: "Here are some examples",
		},
		{
			name: "Multiple examples",
			examples: []ReActExample{
				{
					Task: "Task 1",
					Steps: []string{
						`THOUGHT: Step 1`,
						`FINAL: Done 1`,
					},
				},
				{
					Task: "Task 2",
					Steps: []string{
						`THOUGHT: Step 2`,
						`FINAL: Done 2`,
					},
				},
			},
			want: "Example 1:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatExamples(tt.examples)
			if tt.want == "" {
				if got != "" {
					t.Errorf("FormatExamples() = %v, want empty string", got)
				}
			} else {
				if !strings.Contains(got, tt.want) {
					t.Errorf("FormatExamples() = %v, want to contain '%s'", got, tt.want)
				}
			}
		})
	}
}

// TestFormatExamples_Structure tests the structure of formatted examples
func TestFormatExamples_Structure(t *testing.T) {
	examples := []ReActExample{
		{
			Task: "Calculate 5+3",
			Steps: []string{
				`THOUGHT: I need to add.`,
				`ACTION: calculator(expr="5+3")`,
				`OBSERVATION: 8`,
				`FINAL: The answer is 8.`,
			},
		},
	}

	formatted := FormatExamples(examples)

	// Should contain header
	if !strings.Contains(formatted, "Here are some examples") {
		t.Error("Formatted examples should contain header")
	}

	// Should contain example number
	if !strings.Contains(formatted, "Example 1:") {
		t.Error("Formatted examples should contain example number")
	}

	// Should contain task
	if !strings.Contains(formatted, "Task: Calculate 5+3") {
		t.Error("Formatted examples should contain task")
	}

	// Should contain all steps
	if !strings.Contains(formatted, "THOUGHT: I need to add.") {
		t.Error("Formatted examples should contain THOUGHT step")
	}
	if !strings.Contains(formatted, "ACTION: calculator") {
		t.Error("Formatted examples should contain ACTION step")
	}
	if !strings.Contains(formatted, "OBSERVATION: 8") {
		t.Error("Formatted examples should contain OBSERVATION step")
	}
	if !strings.Contains(formatted, "FINAL: The answer is 8.") {
		t.Error("Formatted examples should contain FINAL step")
	}
}

// TestPredefinedExampleSets tests predefined example sets
func TestPredefinedExampleSets(t *testing.T) {
	expectedSets := []string{"search", "calculation", "research"}

	for _, name := range expectedSets {
		t.Run(name, func(t *testing.T) {
			set, ok := PredefinedExampleSets[name]
			if !ok {
				t.Errorf("PredefinedExampleSets should contain '%s'", name)
				return
			}

			if set.Name != name {
				t.Errorf("Expected set name '%s', got '%s'", name, set.Name)
			}

			if set.Description == "" {
				t.Error("Set should have description")
			}

			if len(set.Examples) == 0 {
				t.Error("Set should have at least one example")
			}

			// Validate all examples in the set
			for i, example := range set.Examples {
				err := ValidateExample(example)
				if err != nil {
					t.Errorf("Example %d in set '%s' is invalid: %v", i, name, err)
				}
			}
		})
	}
}

// TestGetAvailableExampleSets tests getting available sets
func TestGetAvailableExampleSets(t *testing.T) {
	sets := GetAvailableExampleSets()

	if len(sets) == 0 {
		t.Error("Should have at least one predefined set")
	}

	// Check that all returned sets exist
	for _, name := range sets {
		if _, ok := PredefinedExampleSets[name]; !ok {
			t.Errorf("GetAvailableExampleSets returned '%s' which doesn't exist", name)
		}
	}
}

// TestWithReActExamples_String tests adding examples by set name
func TestWithReActExamples_String(t *testing.T) {
	builder := &Builder{}

	builder.WithReActExamples("calculation")

	if builder.reactConfig == nil {
		t.Fatal("reactConfig should be initialized")
	}

	if len(builder.reactConfig.Examples) == 0 {
		t.Error("Examples should be loaded from 'calculation' set")
	}

	// Verify it's the calculation set
	calcSet := PredefinedExampleSets["calculation"]
	if len(builder.reactConfig.Examples) != len(calcSet.Examples) {
		t.Errorf("Expected %d examples, got %d", len(calcSet.Examples), len(builder.reactConfig.Examples))
	}
}

// TestWithReActExamples_Invalid tests invalid set name
func TestWithReActExamples_InvalidName(t *testing.T) {
	builder := &Builder{}

	builder.WithReActExamples("invalid_set_name")

	if builder.reactConfig != nil && len(builder.reactConfig.Examples) > 0 {
		t.Error("Invalid set name should not load examples")
	}
}

// TestWithReActExamples_CustomArray tests adding custom example array
func TestWithReActExamples_CustomArray(t *testing.T) {
	builder := &Builder{}

	customExamples := []ReActExample{
		{
			Task: "Custom task",
			Steps: []string{
				`THOUGHT: Custom thinking`,
				`FINAL: Custom result`,
			},
		},
	}

	builder.WithReActExamples(customExamples)

	if builder.reactConfig == nil {
		t.Fatal("reactConfig should be initialized")
	}

	if len(builder.reactConfig.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(builder.reactConfig.Examples))
	}

	if builder.reactConfig.Examples[0].Task != "Custom task" {
		t.Errorf("Expected custom task, got '%s'", builder.reactConfig.Examples[0].Task)
	}
}

// TestWithReActExamples_SingleExample tests adding single example
func TestWithReActExamples_SingleExample(t *testing.T) {
	builder := &Builder{}

	singleExample := ReActExample{
		Task: "Single task",
		Steps: []string{
			`THOUGHT: thinking`,
			`FINAL: done`,
		},
	}

	builder.WithReActExamples(singleExample)

	if builder.reactConfig == nil {
		t.Fatal("reactConfig should be initialized")
	}

	if len(builder.reactConfig.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(builder.reactConfig.Examples))
	}
}

// TestWithReActExampleSet tests WithReActExampleSet method
func TestWithReActExampleSet(t *testing.T) {
	builder := &Builder{}

	builder.WithReActExampleSet("search")

	if builder.reactConfig == nil {
		t.Fatal("reactConfig should be initialized")
	}

	if len(builder.reactConfig.Examples) == 0 {
		t.Error("Examples should be loaded from 'search' set")
	}

	// Verify it's the search set
	searchSet := PredefinedExampleSets["search"]
	if len(builder.reactConfig.Examples) != len(searchSet.Examples) {
		t.Errorf("Expected %d examples, got %d", len(searchSet.Examples), len(builder.reactConfig.Examples))
	}
}

// TestBuilder_ExamplesInSystemPrompt tests that examples are included in system prompt
func TestBuilder_ExamplesInSystemPrompt(t *testing.T) {
	builder := &Builder{
		reactConfig: NewReActConfig(),
	}

	// Add examples
	builder.WithReActExamples("calculation")

	// Build system prompt
	prompt := builder.buildReActSystemPrompt()

	// Should contain example marker
	if !strings.Contains(prompt, "Here are some examples") {
		t.Error("System prompt should contain examples section")
	}

	// Should contain calculation example content
	if !strings.Contains(prompt, "calculator") {
		t.Error("System prompt should contain calculator example")
	}
}

// TestBuilder_NoExamplesInSystemPrompt tests system prompt without examples
func TestBuilder_NoExamplesInSystemPrompt(t *testing.T) {
	builder := &Builder{
		reactConfig: NewReActConfig(),
	}

	// Build system prompt without examples
	prompt := builder.buildReActSystemPrompt()

	// Should NOT contain example marker
	if strings.Contains(prompt, "Here are some examples") {
		t.Error("System prompt should not contain examples section when no examples provided")
	}
}
