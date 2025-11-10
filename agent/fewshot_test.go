package agent

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ============================================================================
// FewShotExample Tests
// ============================================================================

func TestFewShotExample_Validate(t *testing.T) {
	tests := []struct {
		name    string
		example FewShotExample
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid example",
			example: FewShotExample{
				Input:   "Test input",
				Output:  "Test output",
				Quality: 0.9,
			},
			wantErr: false,
		},
		{
			name: "empty input",
			example: FewShotExample{
				Input:  "",
				Output: "Test output",
			},
			wantErr: true,
			errMsg:  "input cannot be empty",
		},
		{
			name: "empty output",
			example: FewShotExample{
				Input:  "Test input",
				Output: "",
			},
			wantErr: true,
			errMsg:  "output cannot be empty",
		},
		{
			name: "negative quality",
			example: FewShotExample{
				Input:   "Test input",
				Output:  "Test output",
				Quality: -0.1,
			},
			wantErr: true,
			errMsg:  "quality must be between 0.0 and 1.0",
		},
		{
			name: "quality too high",
			example: FewShotExample{
				Input:   "Test input",
				Output:  "Test output",
				Quality: 1.5,
			},
			wantErr: true,
			errMsg:  "quality must be between 0.0 and 1.0",
		},
		{
			name: "whitespace input",
			example: FewShotExample{
				Input:  "   ",
				Output: "Test output",
			},
			wantErr: true,
			errMsg:  "input cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.example.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFewShotExample_SetDefaults(t *testing.T) {
	example := FewShotExample{
		Input:  "Test",
		Output: "Result",
	}

	example.SetDefaults()

	assert.NotEmpty(t, example.ID, "ID should be generated")
	assert.False(t, example.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.Equal(t, 1.0, example.Quality, "Quality should default to 1.0")
}

func TestFewShotExample_HasTag(t *testing.T) {
	example := FewShotExample{
		Input:  "Test",
		Output: "Result",
		Tags:   []string{"tag1", "tag2", "tag3"},
	}

	assert.True(t, example.HasTag("tag1"))
	assert.True(t, example.HasTag("tag2"))
	assert.False(t, example.HasTag("tag4"))
}

func TestFewShotExample_String(t *testing.T) {
	example := FewShotExample{
		Input:   "Hello",
		Output:  "Bonjour",
		Quality: 0.95,
		Tags:    []string{"translation", "french"},
	}

	str := example.String()
	assert.Contains(t, str, "Hello")
	assert.Contains(t, str, "Bonjour")
	assert.Contains(t, str, "0.95")
	assert.Contains(t, str, "translation, french")
}

func TestFewShotExample_JSON(t *testing.T) {
	example := FewShotExample{
		Input:   "Test input",
		Output:  "Test output",
		Quality: 0.8,
		Tags:    []string{"test"},
	}

	// Marshal
	data, err := json.Marshal(example)
	require.NoError(t, err)

	// Unmarshal
	var decoded FewShotExample
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, example.Input, decoded.Input)
	assert.Equal(t, example.Output, decoded.Output)
	assert.Equal(t, example.Quality, decoded.Quality)
	assert.Equal(t, example.Tags, decoded.Tags)
}

func TestFewShotExample_YAML(t *testing.T) {
	example := FewShotExample{
		Input:   "Test input",
		Output:  "Test output",
		Quality: 0.8,
		Tags:    []string{"test"},
	}

	// Marshal
	data, err := yaml.Marshal(example)
	require.NoError(t, err)

	// Unmarshal
	var decoded FewShotExample
	err = yaml.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, example.Input, decoded.Input)
	assert.Equal(t, example.Output, decoded.Output)
	assert.Equal(t, example.Quality, decoded.Quality)
	assert.Equal(t, example.Tags, decoded.Tags)
}

// ============================================================================
// FewShotConfig Tests
// ============================================================================

func TestFewShotConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *FewShotConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &FewShotConfig{
				Examples: []FewShotExample{
					{Input: "A", Output: "B"},
					{Input: "C", Output: "D"},
				},
				MaxExamples:   2,
				SelectionMode: SelectionAll,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config is nil",
		},
		{
			name: "no examples",
			config: &FewShotConfig{
				Examples: []FewShotExample{},
			},
			wantErr: true,
			errMsg:  "no examples provided",
		},
		{
			name: "negative max examples",
			config: &FewShotConfig{
				Examples: []FewShotExample{
					{Input: "A", Output: "B"},
				},
				MaxExamples: -1,
			},
			wantErr: true,
			errMsg:  "max_examples cannot be negative",
		},
		{
			name: "invalid example",
			config: &FewShotConfig{
				Examples: []FewShotExample{
					{Input: "A", Output: "B"},
					{Input: "", Output: "D"}, // Invalid
				},
			},
			wantErr: true,
			errMsg:  "example 1 invalid",
		},
		{
			name: "invalid selection mode",
			config: &FewShotConfig{
				Examples: []FewShotExample{
					{Input: "A", Output: "B"},
				},
				SelectionMode: "invalid_mode",
			},
			wantErr: true,
			errMsg:  "invalid selection mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFewShotConfig_SetDefaults(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "A", Output: "B"},
		},
	}

	config.SetDefaults()

	assert.Equal(t, 5, config.MaxExamples, "MaxExamples should default to 5")
	assert.Equal(t, SelectionAll, config.SelectionMode, "SelectionMode should default to 'all'")
	assert.NotEmpty(t, config.Examples[0].ID, "Example ID should be generated")
}

func TestFewShotConfig_ToPrompt(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "Translate: Hello", Output: "Bonjour"},
			{Input: "Translate: Goodbye", Output: "Au revoir"},
		},
		MaxExamples: 2,
	}
	config.SetDefaults()

	prompt := config.ToPrompt()

	assert.Contains(t, prompt, "Here are examples")
	assert.Contains(t, prompt, "Hello")
	assert.Contains(t, prompt, "Bonjour")
	assert.Contains(t, prompt, "Goodbye")
	assert.Contains(t, prompt, "Au revoir")
}

func TestFewShotConfig_ToPromptWithTemplate(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "A", Output: "B"},
			{Input: "C", Output: "D"},
		},
		PromptTemplate: "Custom format:\n{{.Examples}}",
	}
	config.SetDefaults()

	prompt := config.ToPrompt()

	assert.Contains(t, prompt, "Custom format:")
	assert.Contains(t, prompt, "A")
	assert.Contains(t, prompt, "B")
}

func TestFewShotConfig_Count(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "1", Output: "A"},
			{Input: "2", Output: "B"},
			{Input: "3", Output: "C"},
			{Input: "4", Output: "D"},
			{Input: "5", Output: "E"},
		},
		MaxExamples: 3,
	}
	config.SetDefaults()

	count := config.Count()
	assert.Equal(t, 3, count, "Should return max_examples count")
}

// ============================================================================
// Selection Mode Tests
// ============================================================================

func TestFewShotConfig_SelectionAll(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "1", Output: "A"},
			{Input: "2", Output: "B"},
			{Input: "3", Output: "C"},
		},
		MaxExamples:   2,
		SelectionMode: SelectionAll,
	}

	selected := config.SelectExamples()

	assert.Len(t, selected, 2, "Should return max_examples count")
	assert.Equal(t, "1", selected[0].Input)
	assert.Equal(t, "2", selected[1].Input)
}

func TestFewShotConfig_SelectionBest(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "Low", Output: "A", Quality: 0.5},
			{Input: "High", Output: "B", Quality: 1.0},
			{Input: "Med", Output: "C", Quality: 0.8},
		},
		MaxExamples:   2,
		SelectionMode: SelectionBest,
	}

	selected := config.SelectExamples()

	assert.Len(t, selected, 2)
	assert.Equal(t, "High", selected[0].Input, "Should select highest quality first")
	assert.Equal(t, "Med", selected[1].Input, "Should select second highest")
}

func TestFewShotConfig_SelectionRecent(t *testing.T) {
	now := time.Now()
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "Old", Output: "A", CreatedAt: now.Add(-2 * time.Hour)},
			{Input: "New", Output: "B", CreatedAt: now},
			{Input: "Mid", Output: "C", CreatedAt: now.Add(-1 * time.Hour)},
		},
		MaxExamples:   2,
		SelectionMode: SelectionRecent,
	}

	selected := config.SelectExamples()

	assert.Len(t, selected, 2)
	assert.Equal(t, "New", selected[0].Input, "Should select most recent first")
	assert.Equal(t, "Mid", selected[1].Input, "Should select second most recent")
}

func TestFewShotConfig_SelectionRandom(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "1", Output: "A"},
			{Input: "2", Output: "B"},
			{Input: "3", Output: "C"},
			{Input: "4", Output: "D"},
			{Input: "5", Output: "E"},
		},
		MaxExamples:   3,
		SelectionMode: SelectionRandom,
	}

	selected := config.SelectExamples()

	// Should return 3 examples (randomized)
	assert.Len(t, selected, 3)

	// All should be valid examples
	for _, ex := range selected {
		assert.NotEmpty(t, ex.Input)
		assert.NotEmpty(t, ex.Output)
	}
}

func TestFewShotConfig_SelectionSimilar_FallsBackToAll(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "1", Output: "A"},
			{Input: "2", Output: "B"},
		},
		MaxExamples:   2,
		SelectionMode: SelectionSimilar, // Phase 2 feature, should fall back
	}

	selected := config.SelectExamples()

	assert.Len(t, selected, 2, "Should fall back to 'all' mode")
}

func TestFewShotConfig_MaxExamplesUnlimited(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "1", Output: "A"},
			{Input: "2", Output: "B"},
			{Input: "3", Output: "C"},
		},
		MaxExamples: 0, // Unlimited
	}

	selected := config.SelectExamples()

	assert.Len(t, selected, 3, "Should return all examples when max=0")
}

// ============================================================================
// JSON/YAML Marshaling Tests
// ============================================================================

func TestFewShotConfig_JSON(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "Test", Output: "Result", Quality: 0.9},
		},
		MaxExamples:   5,
		SelectionMode: SelectionBest,
	}

	// Marshal
	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Unmarshal
	var decoded FewShotConfig
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Examples, 1)
	assert.Equal(t, "Test", decoded.Examples[0].Input)
	assert.Equal(t, 5, decoded.MaxExamples)
	assert.Equal(t, SelectionBest, decoded.SelectionMode)
}

func TestFewShotConfig_YAML(t *testing.T) {
	config := &FewShotConfig{
		Examples: []FewShotExample{
			{Input: "Test", Output: "Result", Quality: 0.9},
		},
		MaxExamples:   5,
		SelectionMode: SelectionBest,
	}

	// Marshal
	data, err := yaml.Marshal(config)
	require.NoError(t, err)

	// Unmarshal
	var decoded FewShotConfig
	err = yaml.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Examples, 1)
	assert.Equal(t, "Test", decoded.Examples[0].Input)
	assert.Equal(t, 5, decoded.MaxExamples)
	assert.Equal(t, SelectionBest, decoded.SelectionMode)
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestFewShotConfig_EmptyToPrompt(t *testing.T) {
	var config *FewShotConfig
	prompt := config.ToPrompt()
	assert.Empty(t, prompt, "Nil config should return empty prompt")

	config = &FewShotConfig{}
	prompt = config.ToPrompt()
	assert.Empty(t, prompt, "Config with no examples should return empty prompt")
}

func TestFewShotExample_IsValid(t *testing.T) {
	valid := FewShotExample{
		Input:  "Test",
		Output: "Result",
	}
	assert.True(t, valid.IsValid())

	invalid := FewShotExample{
		Input:  "",
		Output: "Result",
	}
	assert.False(t, invalid.IsValid())
}
