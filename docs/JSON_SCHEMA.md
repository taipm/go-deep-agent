# Builder API - JSON Schema & Structured Outputs

The Builder API provides powerful structured output capabilities through JSON Schema support, enabling type-safe, validated responses from language models.

## Features

- **JSON Mode** - Simple JSON output without strict schema
- **JSON Schema** - Strict structured outputs with validation
- **Nested Structures** - Support for complex nested objects and arrays
- **Type Safety** - Ensures model follows exact schema definition

## Quick Start

### 1. JSON Mode (Simple)

For basic JSON output where you just need valid JSON but don't need strict structure validation:

```go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONMode().
    WithSystem("Respond with JSON containing 'answer' and 'confidence' fields").
    Ask(ctx, "What is the capital of France?")

// Response: {"answer":"Paris","confidence":0.99}
```

**Use JSON Mode when:**
- You need valid JSON but structure can vary
- You want flexibility in response format
- You're using prompt engineering to guide structure

### 2. JSON Schema (Strict)

For structured outputs with guaranteed schema compliance:

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "temperature": map[string]interface{}{
            "type": "number",
            "description": "Temperature in Celsius",
        },
        "condition": map[string]interface{}{
            "type": "string",
            "description": "Weather condition",
        },
    },
    "required": []string{"temperature", "condition"},
    "additionalProperties": false,
}

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("weather", "Weather information", schema, true).
    Ask(ctx, "What's the weather in Tokyo?")

// Response: {"temperature":18,"condition":"cloudy"}
```

**Use JSON Schema when:**
- You need guaranteed structure
- You're parsing into Go structs
- You want type validation
- You need to integrate with external APIs

## API Reference

### WithJSONMode()

Enables JSON object response format (older method).

```go
func (b *Builder) WithJSONMode() *Builder
```

**Example:**
```go
builder.WithJSONMode().
    WithSystem("Return JSON with 'result' field").
    Ask(ctx, "Calculate 10 + 20")
```

**Notes:**
- You must instruct the model to return JSON in your prompt
- Structure is not enforced, just valid JSON
- More flexible but less reliable than JSON Schema

### WithJSONSchema()

Enables strict structured output with schema validation.

```go
func (b *Builder) WithJSONSchema(
    name string,           // Schema name (a-z, A-Z, 0-9, _, -, max 64 chars)
    description string,    // What the response format is for
    schema interface{},    // JSON Schema object
    strict bool,          // Enable strict adherence (recommended: true)
) *Builder
```

**Example:**
```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name": map[string]interface{}{"type": "string"},
        "age": map[string]interface{}{"type": "integer"},
    },
    "required": []string{"name", "age"},
    "additionalProperties": false,
}

builder.WithJSONSchema("person", "Person information", schema, true)
```

**Schema Requirements (Strict Mode):**
- All properties must be in `required` array
- Must include `additionalProperties: false`
- Nested objects must also follow these rules

### WithResponseFormat()

Sets a custom response format directly.

```go
func (b *Builder) WithResponseFormat(
    format *openai.ChatCompletionNewParamsResponseFormatUnion
) *Builder
```

**Example:**
```go
format := &openai.ChatCompletionNewParamsResponseFormatUnion{
    OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
        // ... custom format
    },
}
builder.WithResponseFormat(format)
```

**Use case:** Advanced customization beyond convenience methods.

## Complete Examples

### Example 1: Data Extraction

Extract structured information from unstructured text:

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name": map[string]interface{}{
            "type": "string",
            "description": "Full name",
        },
        "age": map[string]interface{}{
            "type": "integer",
            "description": "Age in years",
        },
        "skills": map[string]interface{}{
            "type": "array",
            "items": map[string]interface{}{
                "type": "string",
            },
            "description": "List of skills",
        },
    },
    "required": []string{"name", "age", "skills"},
    "additionalProperties": false,
}

text := "John Smith is a 32-year-old engineer who knows Go, Python, and AWS."

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("person", "Person info", schema, true).
    Ask(ctx, fmt.Sprintf("Extract data: %s", text))

// Response: {"name":"John Smith","age":32,"skills":["Go","Python","AWS"]}

// Parse into struct
var person struct {
    Name   string   `json:"name"`
    Age    int      `json:"age"`
    Skills []string `json:"skills"`
}
json.Unmarshal([]byte(response), &person)
```

### Example 2: Nested Objects

Handle complex nested structures:

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "book": map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "title": map[string]interface{}{"type": "string"},
                "author": map[string]interface{}{"type": "string"},
                "year": map[string]interface{}{"type": "integer"},
            },
            "required": []string{"title", "author", "year"},
            "additionalProperties": false,
        },
        "review": map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "rating": map[string]interface{}{
                    "type": "integer",
                    "minimum": 1,
                    "maximum": 5,
                },
                "summary": map[string]interface{}{"type": "string"},
            },
            "required": []string{"rating", "summary"},
            "additionalProperties": false,
        },
    },
    "required": []string{"book", "review"},
    "additionalProperties": false,
}

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("book_review", "Book review", schema, true).
    Ask(ctx, "Review '1984' by George Orwell")

// Response: {"book":{"title":"1984","author":"George Orwell","year":1949},"review":{"rating":5,"summary":"..."}}
```

### Example 3: Arrays with Validation

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "cities": map[string]interface{}{
            "type": "array",
            "items": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "name": map[string]interface{}{"type": "string"},
                    "country": map[string]interface{}{"type": "string"},
                    "population": map[string]interface{}{"type": "integer"},
                },
                "required": []string{"name", "country", "population"},
                "additionalProperties": false,
            },
            "minItems": 1,
        },
    },
    "required": []string{"cities"},
    "additionalProperties": false,
}

response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithJSONSchema("cities", "City list", schema, true).
    Ask(ctx, "List 3 largest cities in Europe")
```

## JSON Schema Best Practices

### 1. Always Use Strict Mode

```go
// ✅ Good - Strict enforcement
builder.WithJSONSchema("schema_name", "description", schema, true)

// ❌ Bad - No enforcement
builder.WithJSONSchema("schema_name", "description", schema, false)
```

### 2. Include All Properties in Required

In strict mode, ALL properties must be required:

```go
// ✅ Good
"properties": {
    "name": {...},
    "age": {...},
},
"required": ["name", "age"],  // All properties listed

// ❌ Bad - Will fail in strict mode
"required": ["name"],  // Missing "age"
```

### 3. Set additionalProperties: false

```go
// ✅ Good
map[string]interface{}{
    "type": "object",
    "properties": {...},
    "required": [...],
    "additionalProperties": false,  // Required!
}

// ❌ Bad - Missing additionalProperties
map[string]interface{}{
    "type": "object",
    "properties": {...},
    "required": [...],
    // Missing additionalProperties
}
```

### 4. Apply Rules to Nested Objects

Every nested object must follow the same rules:

```go
// ✅ Good - Nested object has all requirements
"address": {
    "type": "object",
    "properties": {
        "street": {...},
        "city": {...},
    },
    "required": ["street", "city"],
    "additionalProperties": false,  // Required for nested too!
}
```

### 5. Add Descriptions

```go
// ✅ Good - Clear descriptions help the model
"temperature": {
    "type": "number",
    "description": "Temperature in Celsius",  // Helps model understand
}
```

### 6. Use Proper Types

```go
// JSON Schema types:
"string"   // Text
"integer"  // Whole numbers
"number"   // Decimals
"boolean"  // true/false
"array"    // Lists
"object"   // Nested structures
"null"     // Null values
```

### 7. Validate with Constraints

```go
"rating": {
    "type": "integer",
    "minimum": 1,
    "maximum": 5,
},
"email": {
    "type": "string",
    "format": "email",
},
```

## Common Patterns

### Pattern 1: Enum Values

```go
"status": {
    "type": "string",
    "enum": ["active", "pending", "completed"],
}
```

### Pattern 2: Optional Fields (Non-Strict Mode)

When not using strict mode:

```go
"properties": {
    "name": {...},      // Required
    "nickname": {...},  // Optional
},
"required": ["name"],  // Only name is required
```

### Pattern 3: Union Types (oneOf)

```go
"contact": {
    "oneOf": [
        {
            "type": "object",
            "properties": {
                "email": {"type": "string"},
            },
            "required": ["email"],
        },
        {
            "type": "object",
            "properties": {
                "phone": {"type": "string"},
            },
            "required": ["phone"],
        },
    ],
}
```

## Troubleshooting

### Error: "Missing 'property_name'"

**Problem:** Not all properties are in required array (strict mode).

**Solution:**
```go
// Include ALL properties in required
"required": ["field1", "field2", "field3"],  // All of them!
```

### Error: "'additionalProperties' is required"

**Problem:** Missing additionalProperties in object or nested object.

**Solution:**
```go
// Add to every object, including nested ones
"additionalProperties": false,
```

### Model Returns Non-JSON

**Problem:** Using WithJSONMode() without proper instructions.

**Solution:**
```go
// Add clear JSON instruction
builder.WithJSONMode().
    WithSystem("You MUST respond with valid JSON").
    Ask(ctx, prompt)

// Or use WithJSONSchema for guaranteed JSON
```

### Invalid JSON Parsing

**Problem:** Response contains extra text or is malformed.

**Solution:**
- Use WithJSONSchema with strict=true for guaranteed valid JSON
- Check that you're using a model that supports structured outputs (gpt-4o-mini, gpt-4o, etc.)

## Running the Examples

See [builder_json_schema.go](builder_json_schema.go) for complete working examples:

```bash
cd examples
export OPENAI_API_KEY=your-key
go run builder_json_schema.go
```

Examples include:
1. **JSON Mode** - Simple JSON with prompt instructions
2. **Weather Schema** - Basic structured output
3. **Data Extraction** - Extract from unstructured text
4. **Nested Structures** - Complex book review schema

## Related

- [builder_basic.go](builder_basic.go) - Basic Builder usage
- [builder_advanced.go](builder_advanced.go) - Advanced parameters
- [builder_tools.go](builder_tools.go) - Tool calling
- [OpenAI Structured Outputs Guide](https://platform.openai.com/docs/guides/structured-outputs)
