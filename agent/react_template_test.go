package agent

import (
	"strings"
	"testing"
)

// TestPromptTemplate tests basic PromptTemplate structure
func TestPromptTemplate(t *testing.T) {
	template := PromptTemplate{
		Template:    "Hello {name}!",
		Description: "Greeting template",
		Variables: map[string]string{
			"name": "Person's name",
		},
	}

	if template.Template != "Hello {name}!" {
		t.Errorf("Expected template 'Hello {name}!', got '%s'", template.Template)
	}
}

// TestRenderTemplate tests template rendering with variable substitution
func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		values   map[string]string
		want     string
	}{
		{
			name:     "Single variable",
			template: "Hello {name}!",
			values:   map[string]string{"name": "Alice"},
			want:     "Hello Alice!",
		},
		{
			name:     "Multiple variables",
			template: "{greeting} {name}, you are {age} years old.",
			values: map[string]string{
				"greeting": "Hi",
				"name":     "Bob",
				"age":      "30",
			},
			want: "Hi Bob, you are 30 years old.",
		},
		{
			name:     "Missing variable",
			template: "Hello {name}, your role is {role}.",
			values:   map[string]string{"name": "Charlie"},
			want:     "Hello Charlie, your role is {role}.",
		},
		{
			name:     "No variables",
			template: "Static text",
			values:   map[string]string{},
			want:     "Static text",
		},
		{
			name:     "Repeated variable",
			template: "{x} + {x} = {result}",
			values:   map[string]string{"x": "5", "result": "10"},
			want:     "5 + 5 = 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderTemplate(tt.template, tt.values)
			if got != tt.want {
				t.Errorf("RenderTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateTemplate tests template validation
func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid template",
			template:    "Hello {name}!",
			shouldError: false,
		},
		{
			name:        "Empty template",
			template:    "",
			shouldError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "Unmatched opening brace",
			template:    "Hello {name!",
			shouldError: true,
			errorMsg:    "unmatched braces",
		},
		{
			name:        "Unmatched closing brace",
			template:    "Hello name}!",
			shouldError: true,
			errorMsg:    "unmatched braces",
		},
		{
			name:        "Empty variable",
			template:    "Hello {}!",
			shouldError: true,
			errorMsg:    "empty variable",
		},
		{
			name:        "Nested braces",
			template:    "Hello {{name}}!",
			shouldError: true,
			errorMsg:    "nested braces",
		},
		{
			name:        "Multiple variables valid",
			template:    "{a} {b} {c}",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.template)
			if (err != nil) != tt.shouldError {
				t.Errorf("ValidateTemplate() error = %v, shouldError = %v", err, tt.shouldError)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("ValidateTemplate() error = %v, want error containing '%s'", err, tt.errorMsg)
			}
		})
	}
}

// TestExtractTemplateVariables tests variable extraction from templates
func TestExtractTemplateVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "Single variable",
			template: "Hello {name}!",
			want:     []string{"name"},
		},
		{
			name:     "Multiple variables",
			template: "{greeting} {name}, you are {age}.",
			want:     []string{"greeting", "name", "age"},
		},
		{
			name:     "No variables",
			template: "Static text",
			want:     []string{},
		},
		{
			name:     "Repeated variable",
			template: "{x} + {x} = {result}",
			want:     []string{"x", "x", "result"},
		},
		{
			name:     "Empty braces",
			template: "Test {} here",
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTemplateVariables(tt.template)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractTemplateVariables() = %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("ExtractTemplateVariables()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

// TestDefaultReActTemplate tests the default template
func TestDefaultReActTemplate(t *testing.T) {
	if DefaultReActTemplate.Template == "" {
		t.Error("DefaultReActTemplate.Template should not be empty")
	}

	if DefaultReActTemplate.Description == "" {
		t.Error("DefaultReActTemplate.Description should not be empty")
	}

	if len(DefaultReActTemplate.Variables) == 0 {
		t.Error("DefaultReActTemplate should have variables")
	}

	// Should contain key ReAct keywords
	if !strings.Contains(DefaultReActTemplate.Template, "THOUGHT") {
		t.Error("Default template should contain THOUGHT")
	}
	if !strings.Contains(DefaultReActTemplate.Template, "ACTION") {
		t.Error("Default template should contain ACTION")
	}
	if !strings.Contains(DefaultReActTemplate.Template, "FINAL") {
		t.Error("Default template should contain FINAL")
	}
}

// TestPredefinedTemplates tests all predefined templates
func TestPredefinedTemplates(t *testing.T) {
	expectedTemplates := []string{"concise", "detailed", "research"}

	for _, name := range expectedTemplates {
		t.Run(name, func(t *testing.T) {
			template, ok := PredefinedTemplates[name]
			if !ok {
				t.Errorf("PredefinedTemplates should contain '%s'", name)
				return
			}

			if template.Template == "" {
				t.Error("Template should not be empty")
			}

			if template.Description == "" {
				t.Error("Template should have description")
			}

			// Validate template format
			err := ValidateTemplate(template.Template)
			if err != nil {
				t.Errorf("Template '%s' is invalid: %v", name, err)
			}

			// Check for required keywords
			if !strings.Contains(template.Template, "THOUGHT") {
				t.Errorf("Template '%s' should contain THOUGHT", name)
			}
		})
	}
}

// TestGetAvailableTemplates tests getting available template names
func TestGetAvailableTemplates(t *testing.T) {
	templates := GetAvailableTemplates()

	if len(templates) == 0 {
		t.Error("Should have at least one predefined template")
	}

	// Check that all returned templates exist
	for _, name := range templates {
		if _, ok := PredefinedTemplates[name]; !ok {
			t.Errorf("GetAvailableTemplates returned '%s' which doesn't exist", name)
		}
	}
}

// TestWithReActPromptTemplate tests custom template builder method
func TestWithReActPromptTemplate(t *testing.T) {
	builder := &Builder{}
	customTemplate := "Custom template with {tools}"

	builder.WithReActPromptTemplate(customTemplate)

	if builder.reactConfig == nil {
		t.Fatal("reactConfig should be initialized")
	}

	if builder.reactConfig.SystemPrompt != customTemplate {
		t.Errorf("Expected SystemPrompt '%s', got '%s'", customTemplate, builder.reactConfig.SystemPrompt)
	}
}

// TestWithReActTemplate tests predefined template builder method
func TestWithReActTemplate(t *testing.T) {
	builder := &Builder{}

	builder.WithReActTemplate("concise")

	if builder.reactConfig == nil {
		t.Fatal("reactConfig should be initialized")
	}

	expectedTemplate := PredefinedTemplates["concise"].Template
	if builder.reactConfig.SystemPrompt != expectedTemplate {
		t.Error("SystemPrompt should be set to concise template")
	}
}

// TestWithReActTemplate_Invalid tests invalid template name
func TestWithReActTemplateInvalid(t *testing.T) {
	builder := &Builder{}

	builder.WithReActTemplate("invalid_template_name")

	// Should not panic, just ignore
	if builder.reactConfig != nil && builder.reactConfig.SystemPrompt != "" {
		t.Error("Invalid template name should not set SystemPrompt")
	}
}

// TestBuildTemplateVariables tests variable building for templates
func TestBuildTemplateVariables(t *testing.T) {
	// Builder with tools
	builder := &Builder{
		tools: []*Tool{
			{Name: "search", Description: "Search the web"},
			{Name: "calc", Description: "Calculate"},
		},
		reactConfig: NewReActConfig(),
	}

	vars := builder.buildTemplateVariables()

	// Should have tools variable
	if _, ok := vars["tools"]; !ok {
		t.Error("Should have 'tools' variable")
	}

	// Tools should contain tool names
	if !strings.Contains(vars["tools"], "search") {
		t.Error("Tools variable should contain 'search'")
	}
	if !strings.Contains(vars["tools"], "calc") {
		t.Error("Tools variable should contain 'calc'")
	}

	// Should have examples variable (empty if no examples)
	if _, ok := vars["examples"]; !ok {
		t.Error("Should have 'examples' variable")
	}

	// Should have rules variable
	if _, ok := vars["rules"]; !ok {
		t.Error("Should have 'rules' variable")
	}
}

// TestBuildTemplateVariables_NoTools tests variables with no tools
func TestBuildTemplateVariablesNoTools(t *testing.T) {
	builder := &Builder{
		tools:       []*Tool{},
		reactConfig: NewReActConfig(),
	}

	vars := builder.buildTemplateVariables()

	if !strings.Contains(vars["tools"], "No tools available") {
		t.Error("Should indicate no tools available")
	}
}

// TestBuildTemplateVariables_WithExamples tests variables with examples
func TestBuildTemplateVariablesWithExamples(t *testing.T) {
	builder := &Builder{
		tools: []*Tool{},
		reactConfig: &ReActConfig{
			Examples: []ReActExample{
				{
					Task: "Test",
					Steps: []string{
						"THOUGHT: test",
						"FINAL: done",
					},
				},
			},
		},
	}

	vars := builder.buildTemplateVariables()

	if vars["examples"] == "" {
		t.Error("Examples variable should not be empty when examples exist")
	}

	if !strings.Contains(vars["examples"], "Test") {
		t.Error("Examples variable should contain example task")
	}
}

// TestBuildReActSystemPrompt_CustomTemplate tests system prompt with custom template
func TestBuildReActSystemPromptCustomTemplate(t *testing.T) {
	builder := &Builder{
		tools: []*Tool{
			{Name: "search", Description: "Search"},
		},
		reactConfig: &ReActConfig{
			SystemPrompt: "Custom: {tools} and {examples}",
			Examples:     []ReActExample{},
		},
	}

	prompt := builder.buildReActSystemPrompt()

	// Should use custom template
	if !strings.Contains(prompt, "Custom:") {
		t.Error("Should use custom template")
	}

	// Should substitute tools
	if !strings.Contains(prompt, "search") {
		t.Error("Should substitute tools variable")
	}

	// Should NOT contain default template text
	if strings.Contains(prompt, "helpful AI assistant") {
		t.Error("Should not contain default template text")
	}
}

// TestBuildReActSystemPrompt_DefaultTemplate tests system prompt with default template
func TestBuildReActSystemPromptDefaultTemplate(t *testing.T) {
	builder := &Builder{
		tools: []*Tool{
			{Name: "search", Description: "Search"},
		},
		reactConfig: &ReActConfig{
			SystemPrompt: "", // Empty means use default
		},
	}

	prompt := builder.buildReActSystemPrompt()

	// Should use default template
	if !strings.Contains(prompt, "helpful AI assistant") {
		t.Error("Should use default template")
	}

	// Should contain tools
	if !strings.Contains(prompt, "search") {
		t.Error("Should contain tools")
	}
}

// TestTemplateIntegration tests full template integration
func TestTemplateIntegration(t *testing.T) {
	// Create agent with custom template
	builder := &Builder{}

	customTemplate := `Research Assistant

Available tools:
{tools}

Examples:
{examples}

Follow THOUGHT → ACTION → FINAL format.`

	searchTool := &Tool{Name: "search", Description: "Search web"}

	builder.WithReActMode(true).
		WithReActPromptTemplate(customTemplate).
		WithTool(searchTool).
		WithReActExamples("research")

	prompt := builder.buildReActSystemPrompt()

	// Check template content
	if !strings.Contains(prompt, "Research Assistant") {
		t.Error("Should contain custom template header")
	}

	// Check tools substituted
	if !strings.Contains(prompt, "search") {
		t.Error("Should substitute tools")
	}

	// Check examples substituted
	if !strings.Contains(prompt, "Example") {
		t.Error("Should substitute examples")
	}
}
