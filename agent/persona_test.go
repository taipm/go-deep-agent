package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPersonaToSystemPrompt verifies system prompt generation
func TestPersonaToSystemPrompt(t *testing.T) {
	persona := &Persona{
		Name:      "TestAssistant",
		Version:   "1.0",
		Role:      "Helpful Assistant",
		Goal:      "Help users with their questions",
		Backstory: "You are experienced in customer support.",
		Personality: PersonalityConfig{
			Tone:   "friendly",
			Traits: []string{"helpful", "patient"},
			Style:  "conversational",
		},
		Guidelines:     []string{"Be clear", "Be concise"},
		Constraints:    []string{"Don't share private info"},
		KnowledgeAreas: []string{"Customer service", "Product knowledge"},
	}

	prompt := persona.ToSystemPrompt()

	// Verify all components are in the prompt
	assert.Contains(t, prompt, "Helpful Assistant", "Should include role")
	assert.Contains(t, prompt, "Help users with their questions", "Should include goal")
	assert.Contains(t, prompt, "experienced in customer support", "Should include backstory")
	assert.Contains(t, prompt, "friendly", "Should include tone")
	assert.Contains(t, prompt, "helpful, patient", "Should include traits")
	assert.Contains(t, prompt, "Be clear", "Should include guidelines")
	assert.Contains(t, prompt, "Don't share private info", "Should include constraints")
	assert.Contains(t, prompt, "Customer service", "Should include knowledge areas")
}

// TestPersonaValidation tests persona validation rules
func TestPersonaValidation(t *testing.T) {
	tests := []struct {
		name      string
		persona   *Persona
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid persona",
			persona: &Persona{
				Name: "Valid",
				Role: "Assistant",
				Goal: "Help users",
				Personality: PersonalityConfig{
					Tone: "friendly",
				},
			},
			wantError: false,
		},
		{
			name: "missing name",
			persona: &Persona{
				Role: "Assistant",
				Goal: "Help users",
				Personality: PersonalityConfig{
					Tone: "friendly",
				},
			},
			wantError: true,
			errorMsg:  "name is required",
		},
		{
			name: "missing role",
			persona: &Persona{
				Name: "Test",
				Goal: "Help users",
				Personality: PersonalityConfig{
					Tone: "friendly",
				},
			},
			wantError: true,
			errorMsg:  "role is required",
		},
		{
			name: "missing goal",
			persona: &Persona{
				Name: "Test",
				Role: "Assistant",
				Personality: PersonalityConfig{
					Tone: "friendly",
				},
			},
			wantError: true,
			errorMsg:  "goal is required",
		},
		{
			name: "missing personality tone",
			persona: &Persona{
				Name:        "Test",
				Role:        "Assistant",
				Goal:        "Help users",
				Personality: PersonalityConfig{},
			},
			wantError: true,
			errorMsg:  "tone is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.persona.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLoadPersona tests loading persona from YAML file
func TestLoadPersona(t *testing.T) {
	// Load an actual persona file
	persona, err := LoadPersona("../personas/customer_support.yaml")
	require.NoError(t, err, "Should load customer support persona")

	assert.Equal(t, "CustomerSupportSpecialist", persona.Name)
	assert.Equal(t, "1.0.0", persona.Version)
	assert.Equal(t, "Senior Customer Support Agent", persona.Role)
	assert.NotEmpty(t, persona.Goal)
	assert.NotEmpty(t, persona.Backstory)
	assert.Equal(t, "friendly and professional", persona.Personality.Tone)
	assert.Contains(t, persona.Personality.Traits, "empathetic")
	assert.NotEmpty(t, persona.Guidelines)
	assert.NotEmpty(t, persona.Constraints)
	assert.NotNil(t, persona.TechnicalConfig)
}

// TestLoadPersona_InvalidFile tests error handling
func TestLoadPersona_InvalidFile(t *testing.T) {
	_, err := LoadPersona("nonexistent.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read persona file")
}

// TestLoadPersona_InvalidYAML tests malformed YAML
func TestLoadPersona_InvalidYAML(t *testing.T) {
	// Create temp file with invalid YAML
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	os.WriteFile(tmpFile, []byte("invalid: yaml: content: bad"), 0644)

	_, err := LoadPersona(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse persona YAML")
}

// TestLoadPersona_InvalidConfig tests invalid persona content
func TestLoadPersona_InvalidConfig(t *testing.T) {
	// Create temp file with incomplete persona
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "incomplete.yaml")
	content := `
name: "Incomplete"
# Missing role and goal - should fail validation
personality:
  tone: "friendly"
`
	os.WriteFile(tmpFile, []byte(content), 0644)

	_, err := LoadPersona(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid persona")
}

// TestLoadPersonasFromDirectory tests loading multiple personas
func TestLoadPersonasFromDirectory(t *testing.T) {
	personas, err := LoadPersonasFromDirectory("../personas")
	require.NoError(t, err, "Should load all personas from directory")

	// Should have at least the 5 personas we created
	assert.GreaterOrEqual(t, len(personas), 5, "Should load at least 5 personas")

	// Verify specific personas
	support, ok := personas["CustomerSupportSpecialist"]
	assert.True(t, ok, "Should have customer support persona")
	assert.NotNil(t, support)

	reviewer, ok := personas["CodeReviewer"]
	assert.True(t, ok, "Should have code reviewer persona")
	assert.NotNil(t, reviewer)

	writer, ok := personas["TechnicalWriter"]
	assert.True(t, ok, "Should have technical writer persona")
	assert.NotNil(t, writer)
}

// TestSavePersona tests saving persona to YAML
func TestSavePersona(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_persona.yaml")

	persona := &Persona{
		Name:    "TestPersona",
		Version: "1.0",
		Role:    "Test Role",
		Goal:    "Test Goal",
		Personality: PersonalityConfig{
			Tone:   "test tone",
			Traits: []string{"trait1", "trait2"},
		},
	}

	// Save
	err := SavePersona(persona, tmpFile)
	require.NoError(t, err, "Should save persona")

	// Verify file exists
	_, err = os.Stat(tmpFile)
	require.NoError(t, err, "File should exist")

	// Load back and verify
	loaded, err := LoadPersona(tmpFile)
	require.NoError(t, err, "Should load saved persona")

	assert.Equal(t, persona.Name, loaded.Name)
	assert.Equal(t, persona.Role, loaded.Role)
	assert.Equal(t, persona.Goal, loaded.Goal)
	assert.Equal(t, persona.Personality.Tone, loaded.Personality.Tone)
}

// TestSavePersona_InvalidPersona tests saving invalid persona
func TestSavePersona_InvalidPersona(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidPersona := &Persona{
		Name: "Invalid",
		// Missing required fields
	}

	err := SavePersona(invalidPersona, tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot save invalid persona")
}

// TestPersonaRegistry tests PersonaRegistry functionality
func TestPersonaRegistry(t *testing.T) {
	registry := NewPersonaRegistry()

	// Initially empty
	assert.Equal(t, 0, registry.Count())
	assert.Empty(t, registry.List())

	// Add persona
	persona := &Persona{
		Name: "TestPersona",
		Role: "Test Role",
		Goal: "Test Goal",
		Personality: PersonalityConfig{
			Tone: "test",
		},
	}

	err := registry.Add(persona)
	require.NoError(t, err)

	// Verify added
	assert.Equal(t, 1, registry.Count())
	assert.True(t, registry.Has("TestPersona"))

	// Get persona
	retrieved, err := registry.Get("TestPersona")
	require.NoError(t, err)
	assert.Equal(t, "TestPersona", retrieved.Name)

	// List personas
	names := registry.List()
	assert.Contains(t, names, "TestPersona")

	// Remove persona
	registry.Remove("TestPersona")
	assert.Equal(t, 0, registry.Count())
	assert.False(t, registry.Has("TestPersona"))

	// Get non-existent persona
	_, err = registry.Get("NonExistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "persona not found")
}

// TestPersonaRegistry_LoadFromDirectory tests loading directory
func TestPersonaRegistry_LoadFromDirectory(t *testing.T) {
	registry := NewPersonaRegistry()

	err := registry.LoadFromDirectory("../personas")
	require.NoError(t, err)

	// Should have loaded multiple personas
	assert.GreaterOrEqual(t, registry.Count(), 5)
	assert.True(t, registry.Has("CustomerSupportSpecialist"))
	assert.True(t, registry.Has("CodeReviewer"))
}

// TestBuilderWithPersona tests builder integration
func TestBuilderWithPersona(t *testing.T) {
	persona, err := LoadPersona("../personas/customer_support.yaml")
	require.NoError(t, err)

	builder := NewOpenAI("", "test-key").WithPersona(persona)

	// Verify persona is set
	assert.NotNil(t, builder.GetPersona())
	assert.Equal(t, "CustomerSupportSpecialist", builder.GetPersona().Name)

	// Verify system prompt was set
	// Note: We can't directly access systemPrompt (private field),
	// but we verified ToSystemPrompt() works in other tests
}

// TestBuilderToPersona tests exporting builder to persona
func TestBuilderToPersona(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithSystem("You are a test assistant").
		WithTemperature(0.7).
		WithMaxTokens(1000)

	persona := builder.ToPersona("TestExport", "1.0")

	assert.Equal(t, "TestExport", persona.Name)
	assert.Equal(t, "1.0", persona.Version)
	assert.Equal(t, "AI Assistant", persona.Role)
	assert.NotNil(t, persona.TechnicalConfig)
	assert.Equal(t, "gpt-4o-mini", persona.TechnicalConfig.Model)
	assert.Equal(t, 0.7, persona.TechnicalConfig.Temperature)
	assert.Equal(t, 1000, persona.TechnicalConfig.MaxTokens)
}

// TestPersonaWithExamples tests persona with examples
func TestPersonaWithExamples(t *testing.T) {
	persona := &Persona{
		Name: "ExamplePersona",
		Role: "Assistant",
		Goal: "Test",
		Personality: PersonalityConfig{
			Tone: "friendly",
		},
		Examples: []PersonaExample{
			{
				Scenario: "User asks for help",
				Response: "I'm here to help!",
			},
			{
				Scenario: "User is confused",
				Response: "Let me clarify that for you.",
			},
		},
	}

	prompt := persona.ToSystemPrompt()

	// Verify examples are in prompt
	assert.Contains(t, prompt, "Examples of your responses")
	assert.Contains(t, prompt, "User asks for help")
	assert.Contains(t, prompt, "I'm here to help!")
	assert.Contains(t, prompt, "User is confused")
	assert.Contains(t, prompt, "Let me clarify")
}

// TestPersonaString tests String() method
func TestPersonaString(t *testing.T) {
	persona := &Persona{
		Name:    "TestPersona",
		Version: "2.0",
		Role:    "Test Assistant",
		Goal:    "Testing",
		Personality: PersonalityConfig{
			Tone: "test",
		},
	}

	str := persona.String()
	assert.Contains(t, str, "TestPersona")
	assert.Contains(t, str, "v2.0")
	assert.Contains(t, str, "Test Assistant")
}

// TestPersonaGetters tests helper methods
func TestPersonaGetters(t *testing.T) {
	// Persona with technical config
	personaWithConfig := &Persona{
		Name: "WithConfig",
		Role: "Test",
		Goal: "Test",
		Personality: PersonalityConfig{
			Tone: "test",
		},
		TechnicalConfig: &AgentConfig{
			Model:       "gpt-4",
			Temperature: 0.5,
		},
	}

	assert.Equal(t, "gpt-4", personaWithConfig.GetModel())
	assert.Equal(t, 0.5, personaWithConfig.GetTemperature())

	// Persona without technical config (defaults)
	personaWithoutConfig := &Persona{
		Name: "WithoutConfig",
		Role: "Test",
		Goal: "Test",
		Personality: PersonalityConfig{
			Tone: "test",
		},
	}

	assert.Equal(t, "", personaWithoutConfig.GetModel())
	assert.Equal(t, 0.7, personaWithoutConfig.GetTemperature()) // Default
}
