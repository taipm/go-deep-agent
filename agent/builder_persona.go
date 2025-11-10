package agent

// WithPersona applies a persona configuration to the builder.
// The persona defines the agent's behavior, personality, and communication style.
//
// This method:
//  1. Generates a system prompt from the persona's role, personality, and guidelines
//  2. Applies the technical configuration if provided in the persona
//
// Example:
//
//	persona, _ := agent.LoadPersona("personas/customer_support.yaml")
//
//	supportAgent := agent.NewOpenAI("", apiKey).
//	    WithPersona(persona).
//	    WithTools(ticketTool)
//
//	response, _ := supportAgent.Ask(ctx, "My order is late!")
func (b *Builder) WithPersona(persona *Persona) *Builder {
	if persona == nil {
		return b
	}

	// Generate and apply system prompt from persona
	systemPrompt := persona.ToSystemPrompt()
	b.WithSystem(systemPrompt)

	// Apply technical configuration if provided
	if persona.TechnicalConfig != nil {
		b.WithAgentConfig(persona.TechnicalConfig)
	}

	// Store persona reference for later retrieval
	b.persona = persona

	return b
}

// GetPersona returns the currently active persona, if any.
// Returns nil if no persona has been set.
//
// Example:
//
//	if persona := builder.GetPersona(); persona != nil {
//	    fmt.Printf("Using persona: %s\n", persona.Name)
//	}
func (b *Builder) GetPersona() *Persona {
	return b.persona
}

// ToPersona exports the current builder configuration as a Persona.
// This is useful for:
//   - Saving the current agent configuration as a reusable persona
//   - Sharing agent configurations with others
//   - Version controlling agent behavior
//
// Note: This creates a basic persona from technical settings.
// For rich personas with guidelines and examples, create them manually.
//
// Example:
//
//	// Create agent with specific configuration
//	agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithSystem("You are helpful assistant").
//	    WithTemperature(0.7).
//	    WithMemory()
//
//	// Export as persona
//	persona := agent.ToPersona("MyAssistant", "v1.0")
//
//	// Save for reuse
//	agent.SavePersona(persona, "personas/my_assistant.yaml")
func (b *Builder) ToPersona(name, version string) *Persona {
	persona := &Persona{
		Name:    name,
		Version: version,
		Role:    "AI Assistant", // Default role
	}

	// Extract system prompt if set
	if b.systemPrompt != "" {
		persona.Goal = b.systemPrompt
	}

	// Create technical config from builder settings
	persona.TechnicalConfig = b.ToAgentConfig()

	// Set default personality
	persona.Personality = PersonalityConfig{
		Tone: "helpful and professional",
	}

	return persona
}
