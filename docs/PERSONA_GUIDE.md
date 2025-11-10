# Persona Guide

**go-deep-agent** supports persona-based configuration for defining AI agent behavior. This guide covers everything you need to know about creating and using personas.

---

## Table of Contents

- [What is a Persona?](#what-is-a-persona)
- [Why Use Personas?](#why-use-personas)
- [Persona Structure](#persona-structure)
- [Creating Your First Persona](#creating-your-first-persona)
- [Loading and Using Personas](#loading-and-using-personas)
- [Persona Registry](#persona-registry)
- [Real-World Examples](#real-world-examples)
- [Best Practices](#best-practices)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

---

## What is a Persona?

A **persona** is a YAML configuration file that defines:
- **WHO** the agent is (role, backstory)
- **WHAT** it does (goal, knowledge areas)
- **HOW** it behaves (personality, guidelines, constraints)

Instead of writing system prompts manually, you define behavior declaratively in a structured format.

### Traditional Approach (Manual Prompt)

```go
agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem(`You are a customer support agent. Be friendly and helpful.
    Always greet customers warmly. Provide step-by-step solutions...`)
```

### Persona Approach (Structured Configuration)

```yaml
# personas/customer_support.yaml
name: "CustomerSupport"
role: "Customer Support Agent"
goal: "Help customers resolve issues quickly"
personality:
  tone: "friendly and professional"
  traits:
    - empathetic
    - patient
guidelines:
  - "Always greet customers warmly"
  - "Provide step-by-step solutions"
```

```go
persona, _ := agent.LoadPersona("personas/customer_support.yaml")
supportAgent := agent.NewOpenAI("", apiKey).WithPersona(persona)
```

---

## Why Use Personas?

### ✅ Separation of Concerns
- **Behavior** (YAML) separated from **code** (Go)
- Non-developers can define agent behavior
- Easy to version control and review changes

### ✅ Reusability
- One persona, many projects
- Share personas across teams
- Build a library of proven personas

### ✅ Maintainability
- Update behavior without code changes
- A/B test different personas
- Track persona evolution over time

### ✅ Clarity
- Self-documenting agent behavior
- Clear structure for guidelines and constraints
- Easy to understand what agent will do

---

## Persona Structure

A persona consists of these sections:

```yaml
# === METADATA ===
name: "UniqueIdentifier"        # Required: Unique name
version: "1.0.0"                 # Recommended: Version number
description: "Brief description" # Optional

# === IDENTITY ===
role: "Job Title"                # Required: Who the agent is
goal: "What agent does"          # Required: Agent's purpose
backstory: |                     # Optional: Background/experience
  Multi-line backstory text

# === PERSONALITY ===
personality:
  tone: "communication tone"     # Required: Overall tone
  traits:                        # Optional: Personality traits
    - trait1
    - trait2
  style: "communication style"   # Optional: Style description

# === BEHAVIOR ===
guidelines:                      # Optional: What TO do
  - "Guideline 1"
  - "Guideline 2"

constraints:                     # Optional: What NOT to do
  - "Constraint 1"
  - "Constraint 2"

# === EXPERTISE ===
knowledge_areas:                 # Optional: Areas of knowledge
  - "Area 1"
  - "Area 2"

# === EXAMPLES ===
examples:                        # Optional: Example interactions
  - scenario: "User question"
    response: "Expected response"

# === TECHNICAL CONFIG ===
technical_config:                # Optional: Override technical settings
  model: "gpt-4o-mini"
  temperature: 0.7
  # ... (see Configuration Guide)
```

---

## Creating Your First Persona

Let's create a simple assistant persona step by step.

### Step 1: Create File

Create `personas/my_assistant.yaml`:

```yaml
name: "MyAssistant"
version: "1.0"
```

### Step 2: Define Identity

```yaml
role: "Helpful AI Assistant"
goal: "Answer user questions clearly and accurately"
```

### Step 3: Add Personality

```yaml
personality:
  tone: "friendly and professional"
  traits:
    - helpful
    - clear
    - concise
```

### Step 4: Add Guidelines

```yaml
guidelines:
  - "Provide accurate information"
  - "Cite sources when possible"
  - "Admit when unsure"
```

### Step 5: Complete Persona

```yaml
name: "MyAssistant"
version: "1.0"
description: "A helpful AI assistant for general questions"

role: "Helpful AI Assistant"
goal: "Answer user questions clearly and accurately"

personality:
  tone: "friendly and professional"
  traits:
    - helpful
    - clear
    - concise

guidelines:
  - "Provide accurate information"
  - "Cite sources when possible"
  - "Admit when unsure"

constraints:
  - "Don't make up information"
  - "Don't give medical or legal advice"
```

### Step 6: Use It

```go
package main

import (
    "context"
    "fmt"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    persona, _ := agent.LoadPersona("personas/my_assistant.yaml")
    
    myAgent := agent.NewOpenAI("", apiKey).
        WithPersona(persona)
    
    response, _ := myAgent.Ask(context.Background(), "What is Go?")
    fmt.Println(response)
}
```

---

## Loading and Using Personas

### Load Single Persona

```go
persona, err := agent.LoadPersona("personas/customer_support.yaml")
if err != nil {
    log.Fatal(err)
}

supportAgent := agent.NewOpenAI("", apiKey).
    WithPersona(persona)
```

### Load All Personas from Directory

```go
personas, err := agent.LoadPersonasFromDirectory("personas/")
if err != nil {
    log.Fatal(err)
}

// Access specific persona
supportPersona := personas["CustomerSupportSpecialist"]
codeReviewPersona := personas["CodeReviewer"]

// Create agents
support := agent.NewOpenAI("", apiKey).WithPersona(supportPersona)
reviewer := agent.NewOpenAI("", apiKey).WithPersona(codeReviewPersona)
```

### Save Persona

```go
persona := &agent.Persona{
    Name: "CustomAssistant",
    Role: "Assistant",
    Goal: "Help users",
    Personality: agent.PersonalityConfig{
        Tone: "friendly",
    },
}

err := agent.SavePersona(persona, "personas/custom.yaml")
```

### Export Current Agent as Persona

```go
// Create agent with specific configuration
myAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are helpful").
    WithTemperature(0.7)

// Export as persona for reuse
persona := myAgent.ToPersona("ExportedAssistant", "1.0")
agent.SavePersona(persona, "personas/exported.yaml")
```

---

## Persona Registry

For managing multiple personas in memory:

```go
// Create registry
registry := agent.NewPersonaRegistry()

// Load from directory
registry.LoadFromDirectory("personas/")

// Add persona
persona := &agent.Persona{...}
registry.Add(persona)

// Get persona
support, err := registry.Get("CustomerSupportSpecialist")

// Check existence
if registry.Has("CodeReviewer") {
    reviewer, _ := registry.Get("CodeReviewer")
}

// List all personas
names := registry.List()
for _, name := range names {
    fmt.Println(name)
}

// Count
fmt.Printf("Total personas: %d\n", registry.Count())

// Remove
registry.Remove("OldPersona")

// Clear all
registry.Clear()
```

---

## Real-World Examples

### Example 1: Customer Support Agent

```yaml
name: "CustomerSupportTier1"
version: "1.0.0"
role: "Tier 1 Customer Support Agent"
goal: "Resolve common customer issues quickly and escalate complex ones"

backstory: |
  You're a friendly support agent handling first-line customer inquiries.
  You have access to the knowledge base and ticket system.

personality:
  tone: "warm and professional"
  traits:
    - empathetic
    - patient
    - solution-oriented

guidelines:
  - "Greet customers by name if available"
  - "Acknowledge their frustration"
  - "Provide clear step-by-step solutions"
  - "Ask clarifying questions before assuming"
  - "Escalate to Tier 2 if issue is technical"

constraints:
  - "Never promise refunds without manager approval"
  - "Don't share other customers' information"
  - "Don't speculate about bugs or features"

knowledge_areas:
  - "Account management"
  - "Basic troubleshooting"
  - "Billing questions"
  - "Product features"

examples:
  - scenario: "Customer can't log in"
    response: |
      I understand how frustrating it is when you can't access your account.
      Let's get you back in! Have you tried the password reset option?

technical_config:
  model: "gpt-4o-mini"
  temperature: 0.7
  memory:
    working_capacity: 50
    episodic_enabled: true
```

### Example 2: Code Reviewer

```yaml
name: "SeniorCodeReviewer"
version: "1.0.0"
role: "Senior Software Engineer - Code Review Specialist"
goal: "Provide thorough, constructive code reviews that improve quality"

personality:
  tone: "constructive and educational"
  traits:
    - detail-oriented
    - pragmatic
    - mentoring

guidelines:
  - "Start with positive feedback"
  - "Explain WHY, not just WHAT to change"
  - "Provide code examples"
  - "Categorize: Critical > Important > Nice-to-have"
  - "Consider context (MVP vs production)"

constraints:
  - "Don't be pedantic about style"
  - "Don't suggest rewrites without reason"
  - "Don't assume malice"

knowledge_areas:
  - "Go best practices"
  - "Design patterns"
  - "Security (OWASP)"
  - "Performance optimization"

technical_config:
  model: "gpt-4o-mini"
  temperature: 0.3  # More consistent
  max_tokens: 3000  # Longer reviews
```

### Example 3: Technical Writer

```yaml
name: "DeveloperDocumentationWriter"
version: "1.0.0"
role: "Senior Technical Writer"
goal: "Create clear, accurate developer documentation"

personality:
  tone: "clear and concise"
  traits:
    - precise
    - example-driven
    - developer-empathetic

guidelines:
  - "Use active voice"
  - "Provide working code examples"
  - "Start simple, then show advanced usage"
  - "Include common pitfalls"
  - "Test all code before including"

constraints:
  - "Never include untested code"
  - "Don't assume prior knowledge"
  - "Don't use unexplained jargon"

technical_config:
  model: "gpt-4o-mini"
  temperature: 0.3  # Consistent output
  max_tokens: 4000  # Longer docs
```

### Example 4: Data Analyst

```yaml
name: "DataAnalyst"
version: "1.0.0"
role: "Senior Data Analyst"
goal: "Transform data into actionable insights"

personality:
  tone: "analytical yet accessible"
  traits:
    - methodical
    - curious
    - communicative

guidelines:
  - "Validate data quality first"
  - "Present with visualizations"
  - "Explain methodology clearly"
  - "Highlight trends and outliers"
  - "Provide context and benchmarks"

constraints:
  - "Don't draw conclusions from insufficient data"
  - "Don't present correlation as causation"
  - "Don't ignore data quality issues"

knowledge_areas:
  - "Statistical analysis"
  - "Data visualization"
  - "Business metrics"
  - "A/B testing"

technical_config:
  model: "gpt-4o-mini"
  temperature: 0.4
```

---

## Best Practices

### 1. Version Your Personas

```yaml
name: "CustomerSupport"
version: "2.1.0"  # Semantic versioning
```

Track changes in git:
```bash
git log personas/customer_support.yaml
```

### 2. Be Specific in Guidelines

❌ Bad:
```yaml
guidelines:
  - "Be helpful"
```

✅ Good:
```yaml
guidelines:
  - "Provide step-by-step solutions with numbered lists"
  - "Always offer an alternative if first solution fails"
```

### 3. Use Examples for Complex Scenarios

```yaml
examples:
  - scenario: "Customer is angry about a bug"
    response: |
      I'm really sorry this bug is affecting you. I understand your frustration.
      Let me escalate this immediately and find a workaround for now.
```

### 4. Separate Behavior from Technical Config

```yaml
# Behavior (persona-specific)
personality:
  tone: "friendly"

# Technical (environment-specific)
technical_config:
  model: "gpt-4o-mini"  # Dev: gpt-4o-mini, Prod: gpt-4
```

### 5. Use Constraints to Prevent Errors

```yaml
constraints:
  - "Never make financial commitments"
  - "Don't share API keys or passwords"
  - "Escalate to human if user is suicidal"
```

### 6. Organize Personas by Domain

```
personas/
  support/
    tier1.yaml
    tier2.yaml
    vip.yaml
  engineering/
    code_reviewer.yaml
    architect.yaml
  content/
    writer.yaml
    editor.yaml
```

### 7. Create Persona Templates

```yaml
# personas/templates/base_assistant.yaml
name: "BaseAssistant"
role: "AI Assistant"
personality:
  tone: "helpful and professional"
guidelines:
  - "Provide accurate information"
  - "Cite sources"
  - "Admit when unsure"
```

Extend in specific personas:
```yaml
# personas/math_tutor.yaml
name: "MathTutor"
extends: "templates/base_assistant.yaml"  # (conceptual)
role: "Math Tutor"
knowledge_areas:
  - "Algebra"
  - "Calculus"
```

---

## Advanced Usage

### Combining Persona with Code

```go
// Load base persona
persona, _ := agent.LoadPersona("personas/support.yaml")

// Override or extend with code
supportAgent := agent.NewOpenAI("", apiKey).
    WithPersona(persona).
    WithTemperature(0.8).        // Override persona's temperature
    WithTools(ticketSystem).     // Add tools not in persona
    WithMemory()                 // Enable features
```

### Dynamic Persona Selection

```go
func getPersona(userTier string) (*agent.Persona, error) {
    switch userTier {
    case "free":
        return agent.LoadPersona("personas/support_basic.yaml")
    case "pro":
        return agent.LoadPersona("personas/support_pro.yaml")
    case "enterprise":
        return agent.LoadPersona("personas/support_vip.yaml")
    default:
        return agent.LoadPersona("personas/support_basic.yaml")
    }
}

persona, _ := getPersona(customer.Tier)
supportAgent := agent.NewOpenAI("", apiKey).WithPersona(persona)
```

### Multi-Agent Systems

```go
// Load different personas for different roles
researcher, _ := agent.LoadPersona("personas/researcher.yaml")
writer, _ := agent.LoadPersona("personas/writer.yaml")
editor, _ := agent.LoadPersona("personas/editor.yaml")

// Create specialized agents
researchAgent := agent.NewOpenAI("", apiKey).WithPersona(researcher)
writerAgent := agent.NewOpenAI("", apiKey).WithPersona(writer)
editorAgent := agent.NewOpenAI("", apiKey).WithPersona(editor)

// Workflow
facts := researchAgent.Ask(ctx, "Research quantum computing")
draft := writerAgent.Ask(ctx, "Write article: " + facts)
final := editorAgent.Ask(ctx, "Edit: " + draft)
```

---

## Troubleshooting

### Error: "persona name is required"

**Cause**: YAML file missing `name` field

**Solution**:
```yaml
name: "YourPersonaName"  # Add this
role: "..."
```

### Error: "persona role is required"

**Cause**: YAML file missing `role` field

**Solution**:
```yaml
role: "Job Title or Role"  # Add this
```

### Error: "persona goal is required"

**Cause**: YAML file missing `goal` field

**Solution**:
```yaml
goal: "What the agent does"  # Add this
```

### Error: "personality tone is required"

**Cause**: Missing or empty `personality.tone`

**Solution**:
```yaml
personality:
  tone: "friendly"  # Add tone
```

### Error: "failed to parse persona YAML"

**Cause**: Invalid YAML syntax

**Solution**: Check YAML indentation and syntax:
```yaml
# ✅ Correct
personality:
  tone: "friendly"
  traits:
    - helpful

# ❌ Incorrect (bad indentation)
personality:
tone: "friendly"
```

### Error: "invalid technical_config"

**Cause**: Technical config has invalid values

**Solution**: Check validation rules:
- `temperature`: 0.0 to 2.0
- `max_tokens`: > 0
- `memory.working_capacity`: 1 to 1000
- `retry.max_attempts`: 1 to 10

### Persona Not Behaving as Expected

**Debug steps**:

1. **Check generated system prompt**:
```go
persona, _ := agent.LoadPersona("personas/your_persona.yaml")
prompt := persona.ToSystemPrompt()
fmt.Println(prompt)  // See what's actually sent to LLM
```

2. **Verify technical config**:
```go
if persona.TechnicalConfig != nil {
    fmt.Printf("Model: %s\n", persona.TechnicalConfig.Model)
    fmt.Printf("Temperature: %f\n", persona.TechnicalConfig.Temperature)
}
```

3. **Test with simpler persona**:
```yaml
name: "DebugPersona"
role: "Test Assistant"
goal: "Test if persona loading works"
personality:
  tone: "test"
```

### IDE Not Showing Autocomplete

**Solution**: Add JSON Schema to your YAML file:

```yaml
# yaml-language-server: $schema=./schema.json

name: "YourPersona"
# ... rest of persona
```

Or configure in VS Code settings:
```json
{
  "yaml.schemas": {
    "./personas/schema.json": "personas/*.yaml"
  }
}
```

---

## See Also

- [Configuration Guide](CONFIG_GUIDE.md) - Technical configuration options
- [Examples](../examples/) - Working code examples
- [API Documentation](../README.md) - Full API reference

---

**Need Help?**
- Check [existing personas](../personas/) for examples
- Review [persona tests](../agent/persona_test.go) for usage patterns
- Open an issue on GitHub for support
