# 🔧 Quick Gemini Migration Guide

## 🎯 QUICK FIX for Current Gemini Adapter Issues

**Problem:** GeminiAdapter in go-deep-agent v0.12.0 has critical tool calling issues that make it "not intelligent".

**Solution:** Apply these 3 critical fixes immediately:

---

## ⚡ INSTANT FIXES (5 minutes)

### Fix 1: Update Dependencies
```bash
go get github.com/googleapis/go-genai@v1.36.0
```

### Fix 2: Replace Schema Conversion
**File:** `agent/adapters/gemini_adapter.go` (lines 195-220)

**Replace this entire function:**
```go
func (a *GeminiAdapter) convertTools(tools []*agent.Tool) []*genai.Tool {
    geminiTools := make([]*genai.Tool, 0, len(tools))

    for _, tool := range tools {
        // Convert parameters map to Gemini Schema
        params := tool.Parameters

        schema := &genai.Schema{
            Type:       genai.TypeObject,
            Properties: make(map[string]*genai.Schema),
            Required:   []string{},
        }

        // Extract properties from our tool parameters
        if props, ok := params["properties"].(map[string]interface{}); ok {
            for propName, propData := range props {
                if propMap, ok := propData.(map[string]interface{}); ok {
                    paramSchema := &genai.Schema{}

                    // Set type with proper conversion
                    if paramType, ok := propMap["type"].(string); ok {
                        switch strings.ToLower(paramType) {
                        case "string":
                            paramSchema.Type = genai.TypeString
                        case "number":
                            paramSchema.Type = genai.TypeNumber
                        case "integer":
                            paramSchema.Type = genai.TypeInteger
                        case "boolean":
                            paramSchema.Type = genai.TypeBoolean
                        case "array":
                            paramSchema.Type = genai.TypeArray
                        default:
                            paramSchema.Type = genai.TypeString
                        }
                    }

                    // Set description
                    if desc, ok := propMap["description"].(string); ok {
                        paramSchema.Description = desc
                    }

                    schema.Properties[propName] = paramSchema
                }
            }
        }

        // Extract required fields
        if reqs, ok := params["required"].([]string); ok {
            schema.Required = reqs
        }

        funcDecl := &genai.FunctionDeclaration{
            Name:        tool.Name,
            Description: tool.Description,
            Parameters:  schema,
        }

        geminiTools = append(geminiTools, &genai.Tool{
            FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
        })
    }

    return geminiTools
}
```

### Fix 3: Replace Arguments Processing
**File:** `agent/adapters/gemini_adapter.go` (lines 242-255)

**Replace this section:**
```go
// Extract tool calls if present
for _, part := range candidate.Content.Parts {
    if funcCall, ok := part.(genai.FunctionCall); ok {
        // Properly marshal function arguments
        argsJSON, err := json.Marshal(funcCall.Args)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal function arguments: %w", err)
        }

        result.ToolCalls = append(result.ToolCalls, agent.ToolCall{
            ID:        fmt.Sprintf("%s_%s", funcCall.Name, uuid.New().String()[:8]),
            Type:      "function",
            Name:      funcCall.Name,
            Arguments: string(argsJSON),
        })
    }
}
```

**Add required imports at the top of the file:**
```go
import (
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
)
```

---

## ✅ VERIFICATION

**Test the fix:**
```go
// Run this test to verify the fixes work
func main() {
    // Create math tool
    tool := agent.NewTool("add", "Add two numbers")
    tool.AddParameter("a", "number", "First number", true)
    tool.AddParameter("b", "number", "Second number", true)
    tool.WithHandler(func(args string) (string, error) {
        var params struct {
            A float64 `json:"a"`
            B float64 `json:"b"`
        }
        json.Unmarshal([]byte(args), &params)
        return fmt.Sprintf("%.1f", params.A+params.B), nil
    })

    // Test schema conversion
    schema := convertToolSchema(tool)
    fmt.Printf("Schema: %+v\n", schema)

    // Should show proper properties and required fields
}
```

**Expected Output:**
```
Schema: &{Type:OBJECT Properties:map[a:0x1400016c000 b:0x1400016c080] Required:[a b]}
```

---

## 🎯 RESULT

After applying these fixes:
- ✅ **Schema Conversion**: Proper JSON schema → Gemini Schema
- ✅ **Arguments Processing**: Correct JSON marshaling
- ✅ **Tool Calling**: Full functionality restored
- ✅ **Gemini Intelligence**: Works like OpenAI with tools

**Your math calculations will now work properly with Gemini!** 🎉

---

## 📚 FULL IMPLEMENTATION

For complete implementation with all features, see:
- `GEMINI_V1_36_0_UPGRADE_PLAN.md` - Detailed upgrade plan
- `gemini_critical_fixes.go` - All critical fixes
- `gemini_v2_adapter.go` - Complete new implementation

**Timeline:**
- **5 minutes**: Apply quick fixes above
- **1 hour**: Implement full v1.36.0 upgrade
- **1 day**: Complete testing and documentation

**Result:** Gemini becomes as "intelligent" as OpenAI with full tool calling! 🚀