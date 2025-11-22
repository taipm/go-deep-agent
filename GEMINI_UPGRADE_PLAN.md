# Gemini SDK Upgrade Plan: google/generative-ai-go → cloud.google.com/go/ai

## Current Issues Analysis

### Problems with Current Implementation:

1. **Schema Conversion (Line 203-205):**
```go
// CURRENT - BROKEN
schema := &genai.Schema{
    Type: genai.TypeObject,  // ❌ Only type, no properties
}
```

2. **Arguments Processing (Line 246):**
```go
// CURRENT - BROKEN
argsJSON := fmt.Sprintf("%v", funcCall.Args) // ❌ Not proper JSON
```

3. **Missing Tool Result Handling:**
```go
// CURRENT - MISSING
// ❌ No method to send tool results back to Gemini
```

## Upgrade Strategy

### Phase 1: Dependency Update
```bash
# Remove old dependency
go mod tidy -drop github.com/google/generative-ai-go

# Add new Google Cloud AI SDK
go get cloud.google.com/go/ai@v0.14.0
```

### Phase 2: Adapter Rewrite - Critical Changes

#### 2.1 Import Changes
```go
// BEFORE
import "github.com/google/generative-ai-go/genai"

// AFTER
import (
    "cloud.google.com/go/ai"
    "cloud.google.com/go/ai/apiv1beta"
    genaiproto "cloud.google.com/go/ai/apiv1beta/aipb"
    "google.golang.org/api/option"
)
```

#### 2.2 Client Initialization
```go
// BEFORE (gemini_adapter.go:38-44)
func NewGeminiAdapter(apiKey string) (*GeminiAdapter, error) {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        return nil, fmt.Errorf("failed to create Gemini client: %w", err)
    }
    return &GeminiAdapter{client: client}, nil
}

// AFTER - NEW IMPLEMENTATION
func NewGeminiAdapter(apiKey string) (*GeminiAdapter, error) {
    ctx := context.Background()
    client, err := ai.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        return nil, fmt.Errorf("failed to create Gemini client: %w", err)
    }
    return &GeminiAdapter{client: client}, nil
}
```

#### 2.3 Fix Schema Conversion (CRITICAL)
```go
// BEFORE (gemini_adapter.go:203-205) - BROKEN
schema := &genai.Schema{
    Type: genai.TypeObject,
}

// AFTER - FIXED
func (a *GeminiAdapter) convertToolSchema(tool *agent.Tool) *genaiproto.Schema {
    // Convert our Parameters map to proper Gemini Schema
    params := tool.Parameters

    schema := &genaiproto.Schema{
        Type:        genaiproto.Schema_OBJECT,
        Properties:  make(map[string]*genaiproto.Schema),
        Required:    []string{},
    }

    // Extract properties from our tool parameters
    if props, ok := params["properties"].(map[string]interface{}); ok {
        for propName, propData := range props {
            if propMap, ok := propData.(map[string]interface{}); ok {
                paramSchema := &genaiproto.Schema{}

                // Set type
                if paramType, ok := propMap["type"].(string); ok {
                    switch paramType {
                    case "string":
                        paramSchema.Type = genaiproto.Schema_STRING
                    case "number":
                        paramSchema.Type = genaiproto.Schema_NUMBER
                    case "integer":
                        paramSchema.Type = genaiproto.Schema_INTEGER
                    case "boolean":
                        paramSchema.Type = genaiproto.Schema_BOOLEAN
                    case "array":
                        paramSchema.Type = genaiproto.Schema_ARRAY
                    }
                }

                // Set description
                if desc, ok := propMap["description"].(string); ok {
                    paramSchema.Description = desc
                }

                // Handle enum values
                if enumValues, ok := propMap["enum"].([]interface{}); ok {
                    paramSchema.Enum = make([]string, len(enumValues))
                    for i, val := range enumValues {
                        if strVal, ok := val.(string); ok {
                            paramSchema.Enum[i] = strVal
                        }
                    }
                }

                schema.Properties[propName] = paramSchema
            }
        }
    }

    // Extract required fields
    if reqs, ok := params["required"].([]string); ok {
        schema.Required = reqs
    }

    return schema
}
```

#### 2.4 Fix Arguments Processing (CRITICAL)
```go
// BEFORE (gemini_adapter.go:246) - BROKEN
argsJSON := fmt.Sprintf("%v", funcCall.Args)

// AFTER - FIXED
func (a *GeminiAdapter) processFunctionCall(funcCall *genaiproto.FunctionCall) ([]agent.ToolCall, error) {
    // Properly marshal function arguments
    argsJSON, err := json.Marshal(funcCall.Args)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal function arguments: %w", err)
    }

    toolCall := agent.ToolCall{
        ID:        funcCall.Name + "_" + uuid.New().String()[:8], // Generate ID
        Type:      "function",
        Name:      funcCall.Name,
        Arguments: string(argsJSON),
    }

    return []agent.ToolCall{toolCall}, nil
}
```

#### 2.5 Add Tool Result Processing (CRITICAL - NEW)
```go
// NEW METHOD - Add to GeminiAdapter
func (a *GeminiAdapter) ExecuteToolResult(ctx context.Context, toolCallID, functionName, result string) (*agent.CompletionResponse, error) {
    // Create a content message with tool result
    content := &genaiproto.Content{
        Parts: []*genaiproto.Part{
            {
                Data: &genaiproto.Part_FunctionResponse{
                    FunctionResponse: &genaiproto.FunctionResponse{
                        Name: functionName,
                        Response: map[string]interface{}{
                            "result": result,
                        },
                    },
                },
            },
        },
        Role: "user",
    }

    // This should be part of the conversation history
    // Implementation depends on how we manage conversation state
    return nil, fmt.Errorf("tool result processing requires conversation state management")
}
```

#### 2.6 Complete Response Processing (FIXED)
```go
// BEFORE (gemini_adapter.go:242-254) - LIMITED
for _, part := range candidate.Content.Parts {
    if funcCall, ok := part.(genai.FunctionCall); ok {
        argsJSON := fmt.Sprintf("%v", funcCall.Args) // ❌ Broken
        result.ToolCalls = append(result.ToolCalls, agent.ToolCall{
            ID:        "", // ❌ No ID
            Type:      "function",
            Name:      funcCall.Name,
            Arguments: argsJSON,
        })
    }
}

// AFTER - FIXED
for _, part := range candidate.Content.Parts {
    if funcCall := part.GetFunctionCall(); funcCall != nil {
        toolCalls, err := a.processFunctionCall(funcCall)
        if err != nil {
            return nil, fmt.Errorf("failed to process function call: %w", err)
        }
        result.ToolCalls = append(result.ToolCalls, toolCalls...)
    }
}
```

### Phase 3: Complete Rewrite Plan

#### File Structure Changes:
```
agent/adapters/
├── gemini_adapter.go          # Completely rewritten
├── gemini_tool_processor.go   # NEW - Tool-specific logic
└── gemini_schema_converter.go # NEW - Schema conversion utilities
```

#### Critical Files to Modify:

1. **gemini_adapter.go** - Complete rewrite
2. **go.mod** - Update dependencies
3. **builder_adapter_integration_test.go** - Update tests

### Phase 4: Testing Strategy

#### Test Coverage Required:
1. **Schema Conversion Tests**: Verify proper JSON → Gemini Schema
2. **Tool Execution Tests**: End-to-end tool calling
3. **Arguments Processing Tests**: Proper JSON marshaling
4. **Tool Result Tests**: Feedback to Gemini
5. **Error Handling Tests**: Robust error cases

### Phase 5: Migration Timeline

**Week 1**: Dependency updates and basic structure
**Week 2**: Schema conversion implementation
**Week 3**: Tool execution and result processing
**Week 4**: Testing and integration

## Benefits of Upgrade

1. **Proper Tool Support**: Full function calling capabilities
2. **Better Schema Handling**: Accurate parameter conversion
3. **Improved Error Handling**: Better error messages and recovery
4. **Future-Proof**: Latest Google Cloud AI features
5. **Better Documentation**: More examples and guides

## Risk Assessment

**High Risk**: Complete adapter rewrite required
**Mitigation**:
- Maintain backward compatibility
- Comprehensive testing
- Gradual rollout

**Medium Risk**: API changes breaking existing integrations
**Mitigation**: Version compatibility testing

## Implementation Priority

### P0 (Critical - Must Fix):
1. Schema conversion (convertToolSchema)
2. Arguments processing (processFunctionCall)
3. Tool result handling (ExecuteToolResult)

### P1 (High - Should Fix):
1. Response processing improvements
2. Error handling enhancement
3. Client initialization update

### P2 (Medium - Nice to Have):
1. Performance optimizations
2. Additional Gemini-specific features
3. Advanced tool configurations

## Success Criteria

1. ✅ All existing tests pass
2. ✅ Tool calling works with math examples
3. ✅ Schema conversion produces valid Gemini schemas
4. ✅ Tool results properly fed back to conversation
5. ✅ No performance regression
6. ✅ Comprehensive error handling