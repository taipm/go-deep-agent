package agent

import (
	"fmt"
	"strings"
)

// Persona represents an agent's behavioral configuration.
// It defines WHO the agent is, WHAT it does, and HOW it behaves.
type Persona struct {
	// Metadata
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Core Identity
	Role      string `yaml:"role" json:"role"`
	Goal      string `yaml:"goal" json:"goal"`
	Backstory string `yaml:"backstory,omitempty" json:"backstory,omitempty"`

	// Personality
	Personality PersonalityConfig `yaml:"personality" json:"personality"`

	// Behavior Rules
	Guidelines  []string `yaml:"guidelines,omitempty" json:"guidelines,omitempty"`
	Constraints []string `yaml:"constraints,omitempty" json:"constraints,omitempty"`

	// Knowledge Areas
	KnowledgeAreas []string `yaml:"knowledge_areas,omitempty" json:"knowledge_areas,omitempty"`

	// Examples (optional)
	Examples []PersonaExample `yaml:"examples,omitempty" json:"examples,omitempty"`

	// Few-shot Learning (optional) - Phase 1
	FewShot *FewShotConfig `yaml:"fewshot,omitempty" json:"fewshot,omitempty"`

	// Optional Technical Configuration Override
	TechnicalConfig *AgentConfig `yaml:"technical_config,omitempty" json:"technical_config,omitempty"`
}

// PersonalityConfig defines personality traits and communication style
type PersonalityConfig struct {
	Tone   string   `yaml:"tone" json:"tone"`
	Traits []string `yaml:"traits,omitempty" json:"traits,omitempty"`
	Style  string   `yaml:"style,omitempty" json:"style,omitempty"`
}

// PersonaExample provides example scenarios and responses
type PersonaExample struct {
	Scenario string `yaml:"scenario" json:"scenario"`
	Response string `yaml:"response" json:"response"`
}

// ToSystemPrompt generates a system prompt from the persona configuration.
// This converts the structured persona into a natural language prompt.
func (p *Persona) ToSystemPrompt() string {
	var builder strings.Builder

	// Role - WHO you are
	if p.Role != "" {
		builder.WriteString(fmt.Sprintf("You are a %s.\n\n", p.Role))
	}

	// Goal - WHAT you do
	if p.Goal != "" {
		builder.WriteString(fmt.Sprintf("Your goal: %s\n\n", p.Goal))
	}

	// Backstory - Context and experience
	if p.Backstory != "" {
		builder.WriteString(p.Backstory)
		builder.WriteString("\n\n")
	}

	// Personality - HOW you communicate
	builder.WriteString("Communication Style:\n")
	if p.Personality.Tone != "" {
		builder.WriteString(fmt.Sprintf("- Tone: %s\n", p.Personality.Tone))
	}
	if len(p.Personality.Traits) > 0 {
		builder.WriteString(fmt.Sprintf("- Traits: %s\n", strings.Join(p.Personality.Traits, ", ")))
	}
	if p.Personality.Style != "" {
		builder.WriteString(fmt.Sprintf("- Style: %s\n", p.Personality.Style))
	}
	builder.WriteString("\n")

	// Guidelines - What TO do
	if len(p.Guidelines) > 0 {
		builder.WriteString("Guidelines:\n")
		for _, guideline := range p.Guidelines {
			builder.WriteString(fmt.Sprintf("- %s\n", guideline))
		}
		builder.WriteString("\n")
	}

	// Constraints - What NOT to do
	if len(p.Constraints) > 0 {
		builder.WriteString("Important Constraints:\n")
		for _, constraint := range p.Constraints {
			builder.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
		builder.WriteString("\n")
	}

	// Knowledge Areas - Expertise
	if len(p.KnowledgeAreas) > 0 {
		builder.WriteString(fmt.Sprintf("Your expertise includes: %s\n\n", strings.Join(p.KnowledgeAreas, ", ")))
	}

	// Examples - Show expected behavior
	if len(p.Examples) > 0 {
		builder.WriteString("Examples of your responses:\n\n")
		for i, example := range p.Examples {
			builder.WriteString(fmt.Sprintf("%d. Scenario: %s\n", i+1, example.Scenario))
			builder.WriteString(fmt.Sprintf("   Response: %s\n\n", example.Response))
		}
	}

	return builder.String()
}

// Validate checks if the persona configuration is valid
func (p *Persona) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("persona name is required")
	}

	if p.Role == "" {
		return fmt.Errorf("persona role is required (defines who the agent is)")
	}

	if p.Goal == "" {
		return fmt.Errorf("persona goal is required (defines what the agent does)")
	}

	if p.Personality.Tone == "" {
		return fmt.Errorf("personality tone is required")
	}

	// Validate technical config if provided
	if p.TechnicalConfig != nil {
		if err := p.TechnicalConfig.Validate(); err != nil {
			return fmt.Errorf("invalid technical_config: %w", err)
		}
	}

	return nil
}

// GetModel returns the model to use (from technical config or empty)
func (p *Persona) GetModel() string {
	if p.TechnicalConfig != nil {
		return p.TechnicalConfig.Model
	}
	return ""
}

// GetTemperature returns the temperature to use (from technical config or 0.7 default)
func (p *Persona) GetTemperature() float64 {
	if p.TechnicalConfig != nil {
		return p.TechnicalConfig.Temperature
	}
	return 0.7 // Default
}

// String returns a human-readable representation of the persona
func (p *Persona) String() string {
	return fmt.Sprintf("Persona(%s v%s: %s)", p.Name, p.Version, p.Role)
}
