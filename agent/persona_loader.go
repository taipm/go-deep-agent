package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadPersona loads a persona from a YAML file
func LoadPersona(path string) (*Persona, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read persona file %s: %w", path, err)
	}

	// Parse YAML
	var persona Persona
	if err := yaml.Unmarshal(data, &persona); err != nil {
		return nil, fmt.Errorf("failed to parse persona YAML from %s: %w", path, err)
	}

	// Validate
	if err := persona.Validate(); err != nil {
		return nil, fmt.Errorf("invalid persona in %s: %w", path, err)
	}

	return &persona, nil
}

// LoadPersonasFromDirectory loads all persona YAML files from a directory
func LoadPersonasFromDirectory(dir string) (map[string]*Persona, error) {
	personas := make(map[string]*Persona)

	// Check directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Load each YAML file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .yaml and .yml files
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		// Load persona
		path := filepath.Join(dir, name)
		persona, err := LoadPersona(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load persona from %s: %w", path, err)
		}

		// Store by persona name
		personas[persona.Name] = persona
	}

	if len(personas) == 0 {
		return nil, fmt.Errorf("no valid persona files found in directory: %s", dir)
	}

	return personas, nil
}

// SavePersona saves a persona to a YAML file
func SavePersona(persona *Persona, path string) error {
	// Validate before saving
	if err := persona.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid persona: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(persona)
	if err != nil {
		return fmt.Errorf("failed to marshal persona to YAML: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write persona file %s: %w", path, err)
	}

	return nil
}

// PersonaRegistry manages multiple personas
type PersonaRegistry struct {
	personas map[string]*Persona
}

// NewPersonaRegistry creates a new empty persona registry
func NewPersonaRegistry() *PersonaRegistry {
	return &PersonaRegistry{
		personas: make(map[string]*Persona),
	}
}

// LoadFromDirectory loads all personas from a directory into the registry
func (r *PersonaRegistry) LoadFromDirectory(dir string) error {
	personas, err := LoadPersonasFromDirectory(dir)
	if err != nil {
		return err
	}

	// Merge with existing personas
	for name, persona := range personas {
		r.personas[name] = persona
	}

	return nil
}

// Add adds a persona to the registry
func (r *PersonaRegistry) Add(persona *Persona) error {
	if err := persona.Validate(); err != nil {
		return fmt.Errorf("cannot add invalid persona: %w", err)
	}

	r.personas[persona.Name] = persona
	return nil
}

// Get retrieves a persona by name
func (r *PersonaRegistry) Get(name string) (*Persona, error) {
	persona, ok := r.personas[name]
	if !ok {
		return nil, fmt.Errorf("persona not found: %s", name)
	}
	return persona, nil
}

// Has checks if a persona exists in the registry
func (r *PersonaRegistry) Has(name string) bool {
	_, ok := r.personas[name]
	return ok
}

// List returns all persona names in the registry
func (r *PersonaRegistry) List() []string {
	names := make([]string, 0, len(r.personas))
	for name := range r.personas {
		names = append(names, name)
	}
	return names
}

// Count returns the number of personas in the registry
func (r *PersonaRegistry) Count() int {
	return len(r.personas)
}

// Remove removes a persona from the registry
func (r *PersonaRegistry) Remove(name string) {
	delete(r.personas, name)
}

// Clear removes all personas from the registry
func (r *PersonaRegistry) Clear() {
	r.personas = make(map[string]*Persona)
}
