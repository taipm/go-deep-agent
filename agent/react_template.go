package agent

import (
	"fmt"
	"strings"
)

// PromptTemplate represents a customizable system prompt template for ReAct.
// Templates support variable substitution for dynamic content.
type PromptTemplate struct {
	Template    string            // The template string with {variable} placeholders
	Description string            // Description of what this template is for
	Variables   map[string]string // Available variables and their descriptions
}

// DefaultReActTemplate is the default ReAct system prompt template.
var DefaultReActTemplate = PromptTemplate{
	Template: `You are a helpful AI assistant that uses the ReAct (Reasoning + Acting) pattern to solve problems.

Follow this format EXACTLY:

THOUGHT: [Your reasoning about what to do next]
ACTION: tool_name(arg1="value1", arg2="value2")
OBSERVATION: [Tool result will be provided by the system]
... (repeat THOUGHT/ACTION/OBSERVATION as needed)
FINAL: [Your final answer to the user]

Rules:
1. Always start with a THOUGHT to reason about the problem
2. Use ACTION to call available tools when you need information
3. Wait for OBSERVATION before continuing
4. Use FINAL when you have enough information to answer
5. Be concise and focused in your reasoning

{tools}

{examples}`,
	Description: "Standard ReAct pattern template",
	Variables: map[string]string{
		"tools":    "Available tools list",
		"examples": "Few-shot examples",
		"rules":    "Additional rules",
	},
}

// TemplateVariable represents a variable that can be substituted in templates.
type TemplateVariable struct {
	Name        string // Variable name (without braces)
	Value       string // The actual value to substitute
	Required    bool   // Whether this variable is required
	Description string // Description of what this variable contains
}

// WithReActPromptTemplate sets a custom prompt template for ReAct execution.
// The template can include variables like {tools}, {examples}, {rules} which will be
// automatically substituted with appropriate content.
//
// Example:
//
//	customTemplate := `You are a research assistant using ReAct pattern.
//
//	Available tools:
//	{tools}
//
//	Examples:
//	{examples}
//
//	Follow the THOUGHT → ACTION → OBSERVATION → FINAL format strictly.`
//
//	agent := agent.NewOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActPromptTemplate(customTemplate)
func (b *Builder) WithReActPromptTemplate(template string) *Builder {
	if b.reactConfig == nil {
		b.reactConfig = NewReActConfig()
	}
	b.reactConfig.SystemPrompt = template
	return b
}

// RenderTemplate substitutes variables in a template string.
// Variables should be in the format {variable_name}.
// If a variable is not found in the values map, it's left unchanged.
func RenderTemplate(template string, values map[string]string) string {
	result := template
	for key, value := range values {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// ValidateTemplate checks if a template string is well-formed.
// It verifies that all variable placeholders are properly formatted.
func ValidateTemplate(template string) error {
	if template == "" {
		return fmt.Errorf("template cannot be empty")
	}

	// Check for unmatched braces
	openCount := strings.Count(template, "{")
	closeCount := strings.Count(template, "}")

	if openCount != closeCount {
		return fmt.Errorf("template has unmatched braces: %d opening, %d closing", openCount, closeCount)
	}

	// Check for empty variables {}
	if strings.Contains(template, "{}") {
		return fmt.Errorf("template contains empty variable placeholder {}")
	}

	// Check for nested braces
	if strings.Contains(template, "{{") || strings.Contains(template, "}}") {
		return fmt.Errorf("template contains nested braces (not supported)")
	}

	return nil
}

// ExtractTemplateVariables extracts all variable names from a template.
// Returns a slice of variable names (without braces).
func ExtractTemplateVariables(template string) []string {
	var variables []string
	inVar := false
	varName := ""

	for _, char := range template {
		if char == '{' {
			inVar = true
			varName = ""
		} else if char == '}' {
			if inVar && varName != "" {
				variables = append(variables, varName)
			}
			inVar = false
		} else if inVar {
			varName += string(char)
		}
	}

	return variables
}

// buildTemplateVariables creates the variable map for template rendering.
// This is used internally by buildReActSystemPrompt when a custom template is set.
func (b *Builder) buildTemplateVariables() map[string]string {
	vars := make(map[string]string)

	// Build tools list
	toolsList := ""
	if len(b.tools) == 0 {
		toolsList = "(No tools available)\n"
	} else {
		toolsList = "Available tools:\n"
		for _, tool := range b.tools {
			toolsList += fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description)
		}
	}
	vars["tools"] = toolsList

	// Build examples list
	examplesList := ""
	if b.reactConfig != nil && len(b.reactConfig.Examples) > 0 {
		examplesList = FormatExamples(b.reactConfig.Examples)
	}
	vars["examples"] = examplesList

	// Additional rules (can be extended in the future)
	vars["rules"] = ""

	return vars
}

// PredefinedTemplates contains commonly used template variations.
var PredefinedTemplates = map[string]PromptTemplate{
	"concise": {
		Template: `ReAct Pattern - Be concise and direct.

Format:
THOUGHT: [reasoning]
ACTION: tool(args)
OBSERVATION: [result]
FINAL: [answer]

{tools}

{examples}`,
		Description: "Minimalist template for quick responses",
	},
	"detailed": {
		Template: `You are an advanced AI assistant using the ReAct (Reasoning + Acting) pattern.

Your task is to solve complex problems through careful reasoning and tool usage.

STRICT FORMAT:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
THOUGHT: Explain your reasoning step-by-step
ACTION: tool_name(arg1="value1", arg2="value2")
OBSERVATION: [System provides tool result]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Repeat as needed until you have sufficient information.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
FINAL: Provide your comprehensive answer
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

GUIDELINES:
• Start with THOUGHT to plan your approach
• Use ACTIONS to gather necessary information
• Analyze each OBSERVATION carefully
• Only use FINAL when you have complete information
• Be thorough and accurate

{tools}

{examples}

{rules}`,
		Description: "Comprehensive template with detailed instructions",
	},
	"research": {
		Template: `You are a research AI assistant using systematic ReAct reasoning.

RESEARCH PROTOCOL:
1. Analyze the research question (THOUGHT)
2. Identify information sources (ACTION with tools)
3. Evaluate findings (OBSERVATION analysis)
4. Synthesize conclusions (FINAL answer)

FORMAT:
THOUGHT: [Your analytical reasoning]
ACTION: tool_name(parameters)
OBSERVATION: [Research data]
FINAL: [Evidence-based conclusion]

{tools}

EXAMPLE RESEARCH APPROACH:
{examples}

Remember: Back claims with evidence from observations.`,
		Description: "Template optimized for research tasks",
	},
}

// WithReActTemplate sets a predefined template by name.
//
// Example:
//
//	agent := agent.NewOpenAI(apiKey).
//	    WithReActMode(true).
//	    WithReActTemplate("research")
func (b *Builder) WithReActTemplate(templateName string) *Builder {
	if template, ok := PredefinedTemplates[templateName]; ok {
		b.WithReActPromptTemplate(template.Template)
	}
	return b
}

// GetAvailableTemplates returns the names of all predefined templates.
func GetAvailableTemplates() []string {
	templates := make([]string, 0, len(PredefinedTemplates))
	for name := range PredefinedTemplates {
		templates = append(templates, name)
	}
	return templates
}
