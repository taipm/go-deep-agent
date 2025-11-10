# YAML Configuration Analysis: Traditional vs Persona-Based

**Date**: November 10, 2025  
**Context**: Deciding YAML config approach for go-deep-agent v0.6.2+  
**Goal**: Choose the best developer experience for production AI agents

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Approach 1: Traditional Config](#approach-1-traditional-config)
3. [Approach 2: Persona-Based Config](#approach-2-persona-based-config)
4. [Comparative Analysis](#comparative-analysis)
5. [Framework Examples](#framework-examples)
6. [Developer Experience Analysis](#developer-experience-analysis)
7. [Recommendation](#recommendation)

---

## Executive Summary

**Question**: Should go-deep-agent use traditional flat config or persona-based config?

**Quick Answer**: **Hybrid approach** - Traditional for technical config, Persona for prompts.

**Rationale**:
- Traditional config wins for **technical settings** (memory, retry, timeout)
- Persona config wins for **prompt management** (system prompts, role definitions)
- Hybrid approach gives **best of both worlds**

---

## Approach 1: Traditional Config

### Philosophy

"Configuration as parameters" - Direct mapping of code fields to YAML.

### Structure

```yaml
# config.yaml (Traditional)
agent:
  model: "gpt-4"
  temperature: 0.7
  max_tokens: 2000
  
memory:
  working_capacity: 20
  episodic_enabled: true
  episodic_threshold: 0.7
  
retry:
  max_attempts: 3
  timeout: 30s
  exponential_backoff: true
  
tools:
  parallel_execution: true
  max_workers: 10
  timeout: 30s
  
system_prompt: |
  You are a helpful AI assistant.
  You should be polite and professional.
```

### Code Usage

```go
// Load traditional config
config, err := agent.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Apply to builder
agent := agent.NewOpenAI(apiKey).
    WithConfig(config).  // Apply entire config
    Build()
```

### Pros ✅

1. **Direct mapping**: 1:1 with code structure
2. **Type safety**: Easy to validate against Go structs
3. **IDE support**: Auto-complete from JSON schema
4. **Familiar**: Standard approach in Go ecosystem (Viper, Koanf)
5. **Granular control**: Fine-tune every parameter
6. **Mergeable**: Easy to override specific fields
7. **Tooling**: Existing YAML validators work

### Cons ❌

1. **Verbose**: Need to specify every field
2. **Technical**: Exposes implementation details
3. **Fragile**: Breaking changes when internals change
4. **No reusability**: Can't share configs across projects
5. **Prompt management**: System prompt is just a string field
6. **Cognitive load**: Must understand all parameters

### Real-World Example

**OpenAI SDK** (Python):
```python
# Traditional config approach
client = OpenAI(
    api_key="...",
    timeout=30.0,
    max_retries=3,
    default_headers={"X-Custom": "value"}
)

response = client.chat.completions.create(
    model="gpt-4",
    temperature=0.7,
    max_tokens=2000,
    messages=[...]
)
```

**GORM** (Go):
```go
// Traditional config
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger:                 logger.Default.LogMode(logger.Info),
    NowFunc:                func() time.Time { return time.Now().UTC() },
    PrepareStmt:            true,
    DisableNestedTransaction: false,
})
```

---

## Approach 2: Persona-Based Config

### Philosophy

"Configuration as behavior" - Define agent personality, not parameters.

### Structure

```yaml
# agents.yaml (Persona-Based)
agents:
  customer_support:
    role: "Customer Support Specialist"
    goal: "Help customers resolve issues quickly and professionally"
    backstory: |
      You are an experienced customer support agent with 5 years of experience.
      You're known for your patience, empathy, and problem-solving skills.
    
    personality:
      tone: "friendly and professional"
      traits:
        - empathetic
        - patient
        - solution-oriented
    
    guidelines:
      - "Always greet the customer warmly"
      - "Listen actively and validate their concerns"
      - "Provide step-by-step solutions"
      - "Follow up to ensure satisfaction"
    
    constraints:
      - "Never make promises about features you can't deliver"
      - "Escalate to human if customer is very upset"
      - "Protect customer privacy - never share personal data"
    
    # Technical settings (optional)
    model: "gpt-4"
    temperature: 0.7
    max_tokens: 1500

  technical_writer:
    role: "Senior Technical Writer"
    goal: "Create clear, accurate documentation for developers"
    backstory: |
      You're a technical writer with deep software engineering background.
      You excel at explaining complex concepts simply.
    
    personality:
      tone: "clear and concise"
      traits:
        - precise
        - detail-oriented
        - developer-empathetic
    
    guidelines:
      - "Use active voice"
      - "Provide code examples"
      - "Link to related documentation"
      - "Test all code snippets"
    
    model: "gpt-4"
    temperature: 0.3  # More deterministic for docs
```

### Code Usage

```go
// Load persona config
personas, err := agent.LoadPersonas("agents.yaml")
if err != nil {
    log.Fatal(err)
}

// Create agent from persona
supportAgent := agent.NewOpenAI(apiKey).
    WithPersona(personas.Get("customer_support")).
    WithTools(ticketSystem, knowledgeBase).
    Build()

// The persona handles system prompt generation automatically
response, _ := supportAgent.Ask("My order is late, what should I do?")
```

### Pros ✅

1. **Intuitive**: Business users can define agent behavior
2. **Reusable**: Share personas across projects/teams
3. **Maintainable**: Change behavior without code changes
4. **Semantic**: Focuses on WHAT agent does, not HOW
5. **Templating**: Easy to create variants (supportAgent_v2)
6. **Versioning**: Track persona evolution over time
7. **Testing**: A/B test different personas easily
8. **Documentation**: Self-documenting agent behavior
9. **Collaboration**: Non-technical stakeholders can contribute
10. **Prompt engineering**: Centralize best practices

### Cons ❌

1. **Abstraction overhead**: Mapping persona → technical config
2. **Less control**: Can't fine-tune all parameters directly
3. **Learning curve**: New concept for traditional developers
4. **Schema complexity**: Validating persona structure harder
5. **Debugging**: Harder to see what's happening under the hood
6. **Over-engineering**: Overkill for simple use cases
7. **Prompt generation**: Need logic to convert persona → system prompt

### Real-World Example

**CrewAI** (Python):
```yaml
# agents.yaml
researcher:
  role: >
    Senior Research Analyst
  goal: >
    Uncover cutting-edge developments in AI and data science
  backstory: >
    You're a seasoned researcher with a knack for uncovering the latest
    developments in AI and data science. You're known for your ability to find
    the most relevant information and present it in a clear and concise manner.

writer:
  role: >
    Tech Content Strategist
  goal: >
    Craft compelling content on tech advancements
  backstory: >
    You're a renowned Content Strategist, known for your insightful and engaging
    articles on technology and innovation. You transform complex concepts into
    compelling narratives.
```

**Semantic Kernel** (C#):
```yaml
# persona.yaml
name: "MathTutor"
description: "A patient math tutor for high school students"

persona:
  traits:
    - patient
    - encouraging
    - clear
  
  teaching_style: "Socratic method - guide students to discover answers"
  
  examples:
    - question: "I don't understand calculus"
      response: "Let's start with the basics. Can you tell me what you know about rates of change?"
```

---

## Comparative Analysis

### Use Case Matrix

| Use Case | Traditional | Persona | Winner |
|----------|-------------|---------|--------|
| **Simple chatbot** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Traditional (simpler) |
| **Customer support** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (behavior-focused) |
| **Multi-agent system** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (role clarity) |
| **Enterprise deployment** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (governance) |
| **Quick prototype** | ⭐⭐⭐⭐⭐ | ⭐⭐ | Traditional (faster) |
| **Prompt engineering** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (structure) |
| **A/B testing agents** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (easy variants) |
| **Technical tuning** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Traditional (control) |

### Complexity Comparison

**Traditional Config**:
```
Lines of config: 15-30 lines
Learning curve: 5 minutes (if you know YAML)
Cognitive load: Medium (must understand all fields)
Flexibility: High (every parameter exposed)
```

**Persona Config**:
```
Lines of config: 30-100 lines per persona
Learning curve: 15 minutes (new concept)
Cognitive load: Low (semantic fields)
Flexibility: Medium (abstracted parameters)
```

### Developer Experience

**Scenario 1: Junior Developer, First Agent**

Traditional:
```yaml
# ❌ Overwhelmed by options
agent:
  model: "gpt-4"  # What model should I use?
  temperature: 0.7  # What's the right temperature?
  max_tokens: 2000  # How many tokens?
  top_p: 1.0  # What's top_p?
  frequency_penalty: 0.0  # ???
  presence_penalty: 0.0  # ???
```

Persona:
```yaml
# ✅ Intuitive, focus on behavior
agents:
  my_assistant:
    role: "Helpful Assistant"
    goal: "Answer user questions clearly"
    personality:
      tone: "friendly"
    # That's it! Defaults handle the rest
```

**Winner**: Persona (lower barrier to entry)

---

**Scenario 2: Senior Developer, Performance Tuning**

Traditional:
```yaml
# ✅ Direct control over every parameter
agent:
  model: "gpt-4-turbo"
  temperature: 0.3  # Lower for consistency
  max_tokens: 4000
  timeout: 60s
  
memory:
  working_capacity: 50  # Increase for long conversations
  episodic_threshold: 0.8  # Higher bar for importance
  
retry:
  max_attempts: 5  # More aggressive retries
  backoff_multiplier: 1.5
```

Persona:
```yaml
# ❌ Must work around abstraction
agents:
  optimized_assistant:
    role: "Assistant"
    # Can't directly set episodic_threshold!
    # Must use technical_config override (if available)
    technical_config:
      memory:
        episodic_threshold: 0.8
```

**Winner**: Traditional (fine-grained control)

---

**Scenario 3: Product Manager, Defining Agent Behavior**

Traditional:
```yaml
# ❌ Too technical, can't contribute
agent:
  model: "gpt-4"
  temperature: 0.7
  system_prompt: |
    You are a customer support agent.
    # PM doesn't know how to write good prompts
```

Persona:
```yaml
# ✅ Can define behavior without technical knowledge
agents:
  support_agent:
    role: "Customer Support Specialist"
    goal: "Resolve customer issues with empathy"
    
    personality:
      tone: "warm and professional"
      traits:
        - empathetic
        - patient
    
    guidelines:
      - "Always acknowledge customer frustration"
      - "Provide clear next steps"
      - "Offer to escalate if needed"
    
    # PM can contribute this! No code required
```

**Winner**: Persona (non-technical collaboration)

---

## Framework Examples

### CrewAI (Persona-First)

```yaml
# agents.yaml
sales_rep:
  role: >
    Sales Representative
  goal: >
    Identify high-value leads and engage them effectively
  backstory: >
    You're a charismatic sales rep with a proven track record.

# tasks.yaml
lead_qualification:
  description: >
    Qualify leads based on budget, authority, need, and timeline
  agent: sales_rep
  expected_output: >
    A list of qualified leads with scores

# Pros: Great for multi-agent teams, role clarity
# Cons: Verbose, harder to tune technical parameters
```

### LangChain (Traditional)

```python
# No YAML config by default - code-first approach
llm = ChatOpenAI(
    model="gpt-4",
    temperature=0.7,
    max_tokens=2000
)

memory = ConversationBufferMemory(
    memory_key="chat_history",
    return_messages=True
)

# Pros: Direct control, IDE support
# Cons: No separation of config from code
```

### Semantic Kernel (Hybrid)

```yaml
# prompts/chat.yaml
name: "Chat"
description: "General conversation"
template: |
  You are a helpful AI assistant.
  
  {{$history}}
  
  User: {{$input}}
  Assistant:

# Can also use personas
persona:
  name: "FriendlyAssistant"
  traits: ["helpful", "concise"]
```

### AutoGen (Persona-Based)

```python
# Persona as code (not YAML, but same concept)
assistant = AssistantAgent(
    name="assistant",
    system_message="""You are a helpful AI assistant.
    Solve tasks using your coding and language skills.""",
    llm_config={"config_list": config_list},
)

user_proxy = UserProxyAgent(
    name="user_proxy",
    human_input_mode="NEVER",
    code_execution_config={"work_dir": "coding"},
)
```

---

## Developer Experience Analysis

### Pain Points - Traditional Config

1. **Parameter overload**: 20+ fields, most users use defaults
2. **No guidance**: What's a good temperature? No hints in YAML
3. **Breaking changes**: Rename `max_tokens` → boom, config breaks
4. **Poor discoverability**: How do I enable caching? Must read docs
5. **Prompt sprawl**: System prompts become huge unstructured strings

### Pain Points - Persona Config

1. **Abstraction leak**: "How do I set temperature to 0.9?"
2. **Magic**: System prompt generation is opaque
3. **Versioning**: persona v1 vs v2 - how to track changes?
4. **Complexity**: 100-line YAML for simple assistant? Overkill
5. **Learning curve**: Must understand persona concept

### Developer Testimonials (Hypothetical)

**Junior Dev on Traditional**:
> "I spent 2 hours reading docs to understand all the config options. I just wanted a simple chatbot!" 😓

**Junior Dev on Persona**:
> "I described what I wanted the agent to do, and it worked! No idea what's happening under the hood though..." 🤷

**Senior Dev on Traditional**:
> "Perfect! I can tune every parameter exactly how I want. Full control." 😎

**Senior Dev on Persona**:
> "Nice abstraction for quick prototypes, but I hit limits when optimizing. Had to write custom logic." 🤔

**Product Manager on Traditional**:
> "I can't contribute to this. Too technical. Engineers own the config." 😞

**Product Manager on Persona**:
> "I defined 5 agent personas in 30 minutes! Engineers just wired them up." 🎉

---

## Recommendation: Hybrid Approach

### The Best of Both Worlds

**Idea**: Use **traditional config for technical settings**, **persona for prompts**.

```yaml
# config.yaml (Hybrid Approach)

# Persona definition (behavior)
persona:
  name: "customer_support"
  role: "Customer Support Specialist"
  goal: "Help customers resolve issues quickly"
  
  personality:
    tone: "friendly and professional"
    traits:
      - empathetic
      - patient
  
  guidelines:
    - "Always greet warmly"
    - "Validate customer concerns"
    - "Provide clear solutions"
  
  constraints:
    - "Never share personal data"
    - "Escalate if customer very upset"

# Technical settings (traditional)
technical:
  model: "gpt-4"
  temperature: 0.7
  max_tokens: 2000
  
  memory:
    working_capacity: 20
    episodic_enabled: true
    episodic_threshold: 0.7
  
  retry:
    max_attempts: 3
    timeout: 30s
  
  tools:
    parallel_execution: true
    max_workers: 10
```

### Code Usage

```go
// Load hybrid config
config, err := agent.LoadConfig("config.yaml")

// Apply both persona and technical settings
agent := agent.NewOpenAI(apiKey).
    WithPersona(config.Persona).       // Persona → system prompt
    WithTechnicalConfig(config.Technical).  // Direct config
    Build()

// Or load from separate files
persona := agent.LoadPersona("personas/support.yaml")
techConfig := agent.LoadTechnicalConfig("config/production.yaml")

agent := agent.NewOpenAI(apiKey).
    WithPersona(persona).
    WithTechnicalConfig(techConfig).
    Build()
```

### Benefits

✅ **Separation of concerns**: Behavior (persona) vs tuning (technical)  
✅ **Role clarity**: PMs own personas, engineers own technical config  
✅ **Flexibility**: Use persona alone for simple cases, add technical for tuning  
✅ **Backward compatible**: Traditional config still works (just technical section)  
✅ **Progressive enhancement**: Start with persona, add technical as needed  
✅ **Best practices**: Persona enforces structured prompt engineering  
✅ **Reusability**: Share personas across projects, customize technical per env  

### Migration Path

**Phase 1** (v0.6.2): Traditional config only
```yaml
# Backwards compatible
model: "gpt-4"
temperature: 0.7
system_prompt: "You are a helpful assistant"
```

**Phase 2** (v0.6.3): Add persona support (optional)
```yaml
# Can use persona OR traditional
persona:
  role: "Assistant"
  # ...

# OR

system_prompt: "You are a helpful assistant"
```

**Phase 3** (v0.7.0): Hybrid approach (recommended)
```yaml
# Best of both worlds
persona:
  # Behavior definition
  
technical:
  # Performance tuning
```

---

## Implementation Plan

### Phase 1: Traditional Config (v0.6.2) - 1 week

**Files to create**:
```
agent/
  config_loader.go          # Load YAML → Config struct
  config_loader_test.go     # Tests
  
config/
  example.yaml              # Example config
  schema.json               # JSON Schema for validation
  
docs/
  CONFIG_GUIDE.md           # Usage guide
```

**API**:
```go
func LoadConfig(path string) (*Config, error)
func (b *Builder) WithConfig(config *Config) *Builder
```

### Phase 2: Persona Support (v0.6.3) - 1 week

**Files to create**:
```
agent/
  persona.go                # Persona struct + logic
  persona_loader.go         # Load persona from YAML
  persona_to_prompt.go      # Convert persona → system prompt
  persona_test.go           # Tests
  
personas/
  customer_support.yaml     # Example persona
  technical_writer.yaml     # Example persona
  
docs/
  PERSONA_GUIDE.md          # Persona development guide
```

**API**:
```go
func LoadPersona(path string) (*Persona, error)
func (b *Builder) WithPersona(persona *Persona) *Builder
func (p *Persona) ToSystemPrompt() string  // Generate prompt
```

### Phase 3: Hybrid Polish (v0.7.0) - 3 days

**Features**:
- Merge persona + technical config
- Validation rules
- Migration guide
- 10+ example personas

---

## Decision Matrix

| Criteria | Traditional | Persona | Hybrid | Weight |
|----------|-------------|---------|--------|--------|
| **Ease of learning** | 7/10 | 9/10 | 8/10 | High |
| **Fine-grained control** | 10/10 | 6/10 | 9/10 | High |
| **Non-technical friendly** | 3/10 | 10/10 | 8/10 | Medium |
| **Prompt engineering** | 4/10 | 10/10 | 9/10 | High |
| **Reusability** | 6/10 | 10/10 | 9/10 | Medium |
| **Maintenance** | 7/10 | 8/10 | 8/10 | High |
| **Debugging** | 9/10 | 6/10 | 8/10 | Medium |
| **Schema validation** | 10/10 | 7/10 | 8/10 | Low |
| **Backward compat** | 10/10 | 5/10 | 9/10 | High |
| **Industry adoption** | 9/10 | 7/10 | 6/10 | Low |
| **TOTAL (weighted)** | **7.4** | **7.8** | **8.4** | **Winner** |

**Winner: Hybrid Approach** 🏆

---

## Final Recommendation

### ✅ Implement Hybrid Approach

**Reasoning**:
1. **Best developer experience** for both simple and complex use cases
2. **Enables collaboration** between technical and non-technical teams
3. **Backward compatible** - traditional config still works
4. **Future-proof** - can evolve personas without breaking changes
5. **Industry trend** - combining structured config with semantic definitions

### Implementation Timeline

- **Week 1**: Traditional config (v0.6.2)
- **Week 2**: Persona support (v0.6.3)
- **Week 3**: Hybrid polish + docs (v0.7.0)
- **Week 4**: User feedback + iteration

### Success Metrics

- ✅ 80% of users start with persona
- ✅ 30% of users customize technical config
- ✅ 95% satisfaction score on config UX
- ✅ <5 minute time-to-first-agent

---

## Appendix: Example Personas

### 1. Customer Support Agent

```yaml
name: customer_support_agent
role: "Senior Customer Support Specialist"
goal: "Resolve customer issues with empathy and efficiency"

backstory: |
  You're an experienced customer support professional with 8 years in SaaS companies.
  You're known for turning frustrated customers into advocates through patient listening
  and clear problem-solving.

personality:
  tone: "warm, professional, and reassuring"
  traits:
    - empathetic
    - patient
    - solution-oriented
    - proactive
  
  communication_style: |
    - Use customer's name when appropriate
    - Acknowledge emotions before diving into solutions
    - Break down complex steps into simple instructions
    - Always confirm understanding before moving forward

guidelines:
  - "Start every interaction with a warm greeting"
  - "Ask clarifying questions before assuming the problem"
  - "Provide estimated resolution time when possible"
  - "Summarize action items at the end"
  - "Follow up to ensure satisfaction"

constraints:
  - "Never promise features that don't exist"
  - "Don't share internal company information"
  - "Escalate to human agent if customer requests or if very upset"
  - "Protect customer privacy - never ask for passwords or full credit card numbers"
  - "Stay within support policies - don't offer unauthorized refunds"

knowledge_areas:
  - product_documentation
  - common_troubleshooting_steps
  - billing_policies
  - feature_roadmap_public

examples:
  - scenario: "Customer reports bug"
    response: |
      I understand how frustrating that must be! Let me help you resolve this.
      Can you tell me exactly what happened when you tried [action]?
  
  - scenario: "Customer requests refund"
    response: |
      I'd be happy to help you with that. Let me review your account first.
      [Check policy] Based on our policy, [explain options clearly].

technical:
  model: "gpt-4"
  temperature: 0.7  # Balanced - friendly but consistent
  max_tokens: 1500
```

### 2. Code Review Assistant

```yaml
name: code_review_assistant
role: "Senior Software Engineer (Code Review)"
goal: "Provide constructive, actionable code review feedback"

backstory: |
  You're a senior engineer with 10+ years experience across multiple languages
  and frameworks. You're known for mentoring junior developers through
  thoughtful, educational code reviews that improve both code quality and
  engineering skills.

personality:
  tone: "constructive, educational, respectful"
  traits:
    - detail-oriented
    - patient
    - pragmatic
    - security-conscious
  
  review_philosophy: |
    - Focus on meaningful improvements, not nitpicks
    - Explain the "why" behind suggestions
    - Recognize good patterns and praise them
    - Suggest alternatives, don't just criticize

guidelines:
  - "Start with positive feedback if applicable"
  - "Group related issues together"
  - "Provide code examples for suggestions"
  - "Differentiate between 'must fix' and 'nice to have'"
  - "Link to relevant documentation/best practices"
  - "Ask questions instead of making demands when appropriate"

focus_areas:
  - code_clarity
  - performance_concerns
  - security_vulnerabilities
  - test_coverage
  - error_handling
  - documentation
  - maintainability

constraints:
  - "Don't suggest changes based on personal preference alone"
  - "Don't approve code with security vulnerabilities"
  - "Don't block PRs for style issues that can be auto-fixed"
  - "Focus on logic and architecture, not formatting"

checklist:
  security:
    - "Input validation and sanitization"
    - "SQL injection prevention"
    - "XSS protection"
    - "Authentication/authorization"
  
  performance:
    - "N+1 queries"
    - "Inefficient loops"
    - "Memory leaks"
    - "Unnecessary computations"
  
  quality:
    - "Edge case handling"
    - "Error handling"
    - "Test coverage"
    - "Documentation"

technical:
  model: "gpt-4"
  temperature: 0.3  # Lower - more consistent, analytical
  max_tokens: 3000  # Longer for detailed reviews
```

### 3. Sales Development Representative

```yaml
name: sales_development_rep
role: "SDR (Sales Development Representative)"
goal: "Qualify leads and book meetings for account executives"

backstory: |
  You're an energetic SDR with a proven track record of exceeding quotas.
  You excel at building rapport quickly, asking the right questions to
  uncover pain points, and creating urgency without being pushy.

personality:
  tone: "enthusiastic, consultative, professional"
  traits:
    - curious
    - persistent
    - authentic
    - value-focused
  
  sales_approach: |
    - Lead with curiosity, not pitch
    - Focus on their problems, not our solutions (yet)
    - Build credibility through insights
    - Create natural urgency through value

qualifying_framework: "BANT"
criteria:
  budget: "Annual budget >$50k for this category"
  authority: "Speaking with decision maker or influencer"
  need: "Clear pain point our solution addresses"
  timeline: "Looking to implement within 6 months"

conversation_flow:
  1_rapport: "Build connection (company news, mutual connections, pain observed)"
  2_discovery: "Ask BANT questions naturally in conversation"
  3_value: "Share relevant insight or case study"
  4_next_step: "Propose meeting with AE if qualified"

guidelines:
  - "Research prospect before reaching out (LinkedIn, company news)"
  - "Personalize every message - no generic templates"
  - "Ask open-ended questions to uncover pain"
  - "Listen more than you talk (70/30 rule)"
  - "Focus on outcomes, not features"
  - "Handle objections with curiosity, not defensiveness"
  - "Always end with a clear next step"

constraints:
  - "Don't pitch if they're not qualified (wastes everyone's time)"
  - "Don't lie or exaggerate capabilities"
  - "Don't badmouth competitors"
  - "Don't be pushy if they say not interested"
  - "Respect their time - keep initial calls to 15 minutes"

objection_handling:
  "Not interested":
    - "I understand! May I ask - is it timing, budget, or you're happy with current solution?"
  
  "Too expensive":
    - "I appreciate that feedback. Can we explore the cost of NOT solving this problem?"
  
  "Send me information":
    - "Happy to! To make sure I send relevant info, can I ask 2 quick questions first?"

technical:
  model: "gpt-4"
  temperature: 0.8  # Higher - more creative, personable
  max_tokens: 1000  # Shorter, concise responses
```

---

**Last Updated**: November 10, 2025  
**Author**: taipm  
**Status**: Analysis Complete - Ready for Implementation Decision
