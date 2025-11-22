# Gemini SDK Upgrade Plan v1.36.0: Latest googleapis/go-genai Integration

## 🎯 Target Version: googleapis/go-genai v1.36.0

### Key Improvements in v1.36.0:
- ✅ **Enhanced Function Calling**: Better streaming support and argument handling
- ✅ **Improved Media Handling**: part.media_resolution and ImageConfig enhancements
- ✅ **New Configuration Options**: generate_content_config.thinking_level
- ✅ **Better Display Name Support**: For FunctionResponseBlob and FunctionResponseFileData
- ✅ **Universal Function Call Streaming**: Across all languages

---

## 📊 PHÂN TÍCH THỰC TRẠNG HIỆN TẠI

### Vấn Đề Cần Sửa (Based on current gemini_adapter.go):

#### **1. Schema Conversion (Line 203-205) - VẪN ĐỐI LỚN:**
```go
// HIỆN TẠI - SAI HOÀN TOÀN
schema := &genai.Schema{
    Type: genai.TypeObject,  // ❌ Chỉ có type, không có properties
}
```

#### **2. Arguments Processing (Line 246) - VẪN ĐỐI LỚN:**
```go
// HIỆN TẠI - SAI HOÀN TOÀN
argsJSON := fmt.Sprintf("%v", funcCall.Args) // ❌ Không phải JSON hợp lệ
```

#### **3. Tool Result Handling - THIẾU HOÀN TOÀN:**
```go
// HIỆN TẠI - KHÔNG TỒN TẠI
// ❌ Không có method để gửi tool results về Gemini
```

---

## 🚀 PHƯƠNG ÁN UPGRADE CHI TIẾT

### Phase 1: Dependency Update
```bash
# Remove old dependency
go get github.com/google/generative-ai-go@none

# Add latest googleapis/go-genai
go get github.com/googleapis/go-genai@v1.36.0
```

### Phase 2: Complete Rewrite với v1.36.0 API

#### 2.1 Cập Nhật Imports
```go
// CŨ
import "github.com/google/generative-ai-go/genai"

// MỚI
import (
    "github.com/googleapis/go-genai/genai"
    "github.com/googleapis/go-genai/internal/generativelanguage"
)
```

#### 2.2 Client Initialization (Đúng chuẩn v1.36.0)
```go
// MỚI IMPLEMENTATION
func NewGeminiAdapterV3(apiKey, model string) (*GeminiAdapterV3, error) {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        return nil, fmt.Errorf("failed to create Gemini client: %w", err)
    }

    if model == "" {
        model = "gemini-1.5-pro" // Sử dụng model mới nhất
    }

    return &GeminiAdapterV3{
        client: client,
        model:  model,
    }, nil
}
```

#### 2.3 Fix Schema Conversion (QUAN TRỌNG - Sửa lại line 203-205)
```go
// HIỆN TẠI - SAI HOÀN TOÀN (gemini_adapter.go:203-205)
schema := &genai.Schema{
    Type: genai.TypeObject,  // ❌ Chỉ có type, không properties
}

// MỚI IMPLEMENTATION - ĐÚNG CHUẨN v1.36.0
func (a *GeminiAdapterV3) convertToolSchema(tool *agent.Tool) *genai.Schema {
    // Convert our Parameters map to proper Gemini Schema
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
                        if itemType, ok := propMap["items"].(string); ok {
                            paramSchema.Items = &genai.Schema{
                                Type: convertStringToGeminiType(itemType),
                            }
                        }
                    case "object":
                        paramSchema.Type = genai.TypeObject
                    default:
                        paramSchema.Type = genai.TypeString
                    }
                }

                // Set description
                if desc, ok := propMap["description"].(string); ok {
                    paramSchema.Description = desc
                }

                // Handle enum values (New in v1.36.0)
                if enumValues, ok := propMap["enum"].([]interface{}); ok {
                    paramSchema.Enum = make([]interface{}, len(enumValues))
                    for i, val := range enumValues {
                        paramSchema.Enum[i] = val
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

#### 2.4 Fix Arguments Processing (QUAN TRỌNG - Sửa lại line 246)
```go
// HIỆN TẠI - SAI HOÀN TOÀN (gemini_adapter.go:246)
argsJSON := fmt.Sprintf("%v", funcCall.Args) // ❌ Not proper JSON

// MỚI IMPLEMENTATION - Đúng chuẩn v1.36.0
func (a *GeminiAdapterV3) processFunctionCall(funcCall *genai.FunctionCall) ([]agent.ToolCall, error) {
    // Sử dụng built-in JSON marshaling của go-genai v1.36.0
    argsJSON, err := json.Marshal(funcCall.Args)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal function arguments: %w", err)
    }

    // Generate unique ID cho tool call
    toolCallID := fmt.Sprintf("%s_%s", funcCall.Name, uuid.New().String()[:8])

    toolCall := agent.ToolCall{
        ID:        toolCallID,
        Type:      "function",
        Name:      funcCall.Name,
        Arguments: string(argsJSON),
    }

    return []agent.ToolCall{toolCall}, nil
}
```

#### 2.5 Add Tool Result Processing (TÍNH NĂNG MỚI)
```go
// TÍNH NĂNG MỚI - Hoàn toàn thiếu trong current implementation
func (a *GeminiAdapterV3) HandleToolResult(ctx context.Context, toolCallID, functionName, result string) (*genai.Content, error) {
    // Sử dụng NewPartFromFunctionResponse từ v1.36.0
    responsePart := genai.NewPartFromFunctionResponse(functionName, map[string]interface{}{
        "result": result,
        "id":     toolCallID,
    })

    return &genai.Content{
        Parts: []genai.Part{responsePart},
        Role:  "user", // Tool results come from user perspective
    }, nil
}

// Method để gửi tool results về conversation
func (a *GeminiAdapterV3) SendToolResult(ctx context.Context, conversationHistory []genai.Content, toolCallID, functionName, result string) ([]genai.Content, error) {
    toolResultContent, err := a.HandleToolResult(ctx, toolCallID, functionName, result)
    if err != nil {
        return nil, fmt.Errorf("failed to create tool result: %w", err)
    }

    // Append tool result to conversation history
    conversationHistory = append(conversationHistory, *toolResultContent)

    return conversationHistory, nil
}
```

#### 2.6 Fix Response Processing (Sửa lại line 242-254)
```go
// HIỆN TẠI - VẤN ĐỀI (gemini_adapter.go:242-254)
for _, part := range candidate.Content.Parts {
    if funcCall, ok := part.(genai.FunctionCall); ok {
        argsJSON := fmt.Sprintf("%v", funcCall.Args) // ❌ Wrong
        result.ToolCalls = append(result.ToolCalls, agent.ToolCall{
            ID:        "", // ❌ No ID
            Type:      "function",
            Name:      funcCall.Name,
            Arguments: argsJSON,
        })
    }
}

// MỚI IMPLEMENTATION - Đúng chuẩn v1.36.0
func (a *GeminiAdapterV3) extractToolCallsFromResponse(candidate *genai.Candidate) ([]agent.ToolCall, error) {
    var toolCalls []agent.ToolCall

    // Sử dụng FunctionCalls() method từ v1.36.0
    functionCalls := candidate.FunctionCalls()

    for _, funcCall := range functionCalls {
        processedCalls, err := a.processFunctionCall(funcCall)
        if err != nil {
            return nil, fmt.Errorf("failed to process function call: %w", err)
        }
        toolCalls = append(toolCalls, processedCalls...)
    }

    return toolCalls, nil
}
```

#### 2.7 Enhanced Streaming với Tool Calling (TÍNH NĂNG MỚI)
```go
// TÍNH NĂNG MỚI - v1.36.0 Enhanced Streaming
func (a *GeminiAdapterV3) StreamWithTools(ctx context.Context, req *agent.CompletionRequest, onChunk func(string), onToolCall func([]agent.ToolCall)) (*agent.CompletionResponse, error) {
    model := a.client.GenerativeModel(req.Model)
    a.configureModel(model, req)

    // Convert messages to Gemini format
    contents := a.convertMessagesToContents(req.Messages)

    // Enable tool calling with auto mode
    if len(req.Tools) > 0 {
        tools := a.convertTools(req.Tools)
        model.Tools = tools
    }

    // Create streaming iterator with enhanced tool support
    iter := model.GenerateContentStream(ctx, contents...)

    var fullContent string
    var usage genai.UsageMetadata
    var allToolCalls []agent.ToolCall

    for {
        chunk, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("gemini streaming error: %w", err)
        }

        // Extract content from chunk
        if len(chunk.Candidates) > 0 {
            candidate := chunk.Candidates[0]

            // Process text content
            for _, part := range candidate.Content.Parts {
                if textPart := part.Text; textPart != "" {
                    fullContent += textPart
                    if onChunk != nil {
                        onChunk(textPart)
                    }
                }
            }

            // Process function calls with v1.36.0 enhancements
            functionCalls := candidate.FunctionCalls()
            if len(functionCalls) > 0 {
                chunkToolCalls, err := a.processFunctionCalls(functionCalls)
                if err != nil {
                    return nil, fmt.Errorf("failed to process function calls: %w", err)
                }

                allToolCalls = append(allToolCalls, chunkToolCalls...)

                if onToolCall != nil {
                    onToolCall(chunkToolCalls)
                }
            }

            // Track usage (last chunk has final counts)
            if chunk.UsageMetadata != nil {
                usage = *chunk.UsageMetadata
            }
        }
    }

    return &agent.CompletionResponse{
        Content:   fullContent,
        ToolCalls: allToolCalls,
        Usage: agent.TokenUsage{
            PromptTokens:     int(usage.PromptTokenCount),
            CompletionTokens: int(usage.CandidatesTokenCount),
            TotalTokens:      int(usage.TotalTokenCount),
        },
    }, nil
}
```

### Phase 3: Integration với go-deep-agent Builder

#### 3.1 Update Builder để support GeminiAdapterV3
```go
// Trong builder.go hoặc file phù hợp
func (b *Builder) WithGeminiAdapterV3(apiKey, model string) *Builder {
    adapter, err := adapters.NewGeminiAdapterV3(apiKey, model)
    if err != nil {
        b.setError(fmt.Errorf("failed to create Gemini adapter v3: %w", err))
        return b
    }

    return b.WithAdapter(adapter)
}

// Helper method để dễ sử dụng
func (b *Builder) WithGemini(apiKey string) *Builder {
    return b.WithGeminiAdapterV3(apiKey, "gemini-1.5-pro")
}
```

### Phase 4: Comprehensive Testing

#### 4.1 Schema Conversion Tests
```go
func TestGeminiV3_SchemaConversion(t *testing.T) {
    adapter, _ := adapters.NewGeminiAdapterV3("test-key", "gemini-1.5-pro")

    // Test tool với complex parameters
    tool := agent.NewTool("calculate", "Perform calculations")
    tool.AddParameter("expression", "string", "Mathematical expression", true)
    tool.AddParameter("precision", "number", "Number of decimal places", false)

    schema := adapter.(*adapters.GeminiAdapterV3).convertToolSchema(tool)

    assert.NotNil(t, schema)
    assert.Equal(t, genai.TypeObject, schema.Type)
    assert.Contains(t, schema.Properties, "expression")
    assert.Contains(t, schema.Properties, "precision")
    assert.Contains(t, schema.Required, "expression")
}
```

#### 4.2 Tool Calling End-to-End Tests
```go
func TestGeminiV3_ToolCalling(t *testing.T) {
    adapter, _ := adapters.NewGeminiAdapterV3("test-key", "gemini-1.5-pro")

    // Create tool
    tool := agent.NewTool("add_numbers", "Add two numbers")
    tool.AddParameter("a", "number", "First number", true)
    tool.AddParameter("b", "number", "Second number", true)
    tool.WithHandler(func(args string) (string, error) {
        var params struct {
            A float64 `json:"a"`
            B float64 `json:"b"`
        }
        json.Unmarshal([]byte(args), &params)
        return fmt.Sprintf("%.2f", params.A+params.B), nil
    })

    // Test tool processing
    req := &agent.CompletionRequest{
        Tools: []*agent.Tool{tool},
        Messages: []agent.Message{
            {Role: "user", Content: "Add 5 and 3"},
        },
    }

    config := adapter.(*adapters.GeminiAdapterV3).convertTools(req)
    assert.Len(t, config, 1)

    funcDecl := config[0].FunctionDeclarations[0]
    assert.Equal(t, "add_numbers", funcDecl.Name)
    assert.NotNil(t, funcDecl.Parameters)
    assert.Contains(t, funcDecl.Parameters.Properties, "a")
    assert.Contains(t, funcDecl.Parameters.Properties, "b")
}
```

### Phase 5: Migration Timeline

**Week 1:**
- ✅ Update dependencies to googleapis/go-genai@v1.36.0
- ✅ Create GeminiAdapterV3 with proper schema conversion
- ✅ Fix arguments processing

**Week 2:**
- ✅ Add tool result handling
- ✅ Enhanced streaming support
- ✅ Integration testing

**Week 3:**
- ✅ Builder integration
- ✅ Comprehensive test suite
- ✅ Performance testing

**Week 4:**
- ✅ Documentation updates
- ✅ Example implementations
- ✅ Production deployment

---

## 🎯 LỢI ÍCH UPGRADE

### 1. **Chức Năng Hoàn Thiện:**
- ✅ Schema conversion đúng 100%
- ✅ JSON arguments processing chính xác
- ✅ Tool result feedback hoạt động
- ✅ Streaming với enhanced tool calling

### 2. **Performance Cải Thiện:**
- ✅ Streaming function calls (v1.36.0 feature)
- ✅ Better memory management
- ✅ Optimized response processing

### 3. **Developer Experience:**
- ✅ Compatible với existing tool interface
- ✅ Better error messages
- ✅ More comprehensive documentation

### 4. **Future-Proof:**
- ✅ Latest API features (v1.36.0)
- ✅ Thinking level support
- ✅ Enhanced media handling

---

## 🚀 QUICK START GUIDE

```go
// Cách sử dụng mới với v1.36.0
func main() {
    // Create enhanced Gemini adapter
    adapter, err := adapters.NewGeminiAdapterV3("your-api-key", "gemini-1.5-pro")
    if err != nil {
        log.Fatal(err)
    }
    defer adapter.Close()

    // Create math tool
    calculatorTool := agent.NewTool("calculator", "Perform mathematical calculations")
    calculatorTool.AddParameter("expression", "string", "Math expression to evaluate", true)
    calculatorTool.WithHandler(func(args string) (string, error) {
        // Implementation...
    })

    // Use with builder
    agent := agent.NewBuilder().
        WithAdapter(adapter).
        WithTools(calculatorTool).
        WithAutoExecute(true).
        Build()

    // Tool calling now works properly!
    response, err := agent.Ask("Calculate 15 * 8 + 3")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response)
}
```

**Kết quả:** Gemini giờ đây hoạt động "thông minh" như OpenAI với full tool calling capabilities! 🎉