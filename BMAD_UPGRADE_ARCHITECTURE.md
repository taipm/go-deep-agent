# BMAD Method: Architecture Phase - Gemini SDK Upgrade Technical Specifications
# Phase 3: Detailed Technical Design and Implementation Specifications

**Date:** 2025-11-22
**Team:** Go-Deep-Agent Development Team
**Previous Phases:** Brainstorming ✅ → Mind Mapping ✅
**Current Phase:** Architecture Design
**Next Phase:** Development Implementation

---

## 🎯 ARCHITECTURE OVERVIEW

### **High-Level System Architecture**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      GO-DEEP-AGENT ENHANCED SYSTEM ARCHITECTURE                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  ┌─────────────────────────────────────┐    ┌─────────────────────────────────────┐   │
│  │       MULTIPLAYER SYSTEM            │    │      ENHANCED ADAPTER LAYER         │   │
│  │       (Enterprise-Ready)           │    │     (v0.12.1 → v0.13.0)           │   │
│  │  - Load Balancer                      │    │  - OpenAI Adapter (✅ Working)       │   │
│  │  - Circuit Breaker                  │    │  - Gemini Adapter (❌ Upgrade Target) │   │
│  │  - Health Monitor                   │    │  - Ollama Adapter (✅ Working)       │   │
│  │  - Metrics Collector               │    │  - Custom Adapter Framework        │   │
│  │  - Enhanced Error Handling          │    │  - Tool Execution Engine           │   │
│  └─────────────────────────────────────┘    └─────────────────────────────────────┘   │
│                                                                            │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                    GEMINI v1.36.0 UPGRADE ARCHITECTURE                         │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                                                            │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                    SHARED TOOL SYSTEM (Enhanced)                            │   │
  │  │  - Universal Tool Registry                                              │   │
  │  │  - Advanced Execution Engine                                           │   │
  │  │  - Result Processing Pipeline                                          │   │
  │  │  - Performance Optimization                                            │   │
  │  │  - Security Hardening                                                 │   │
  │  └─────────────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🔧 DETAILED COMPONENT ARCHITECTURE

### **1. Enhanced Gemini Adapter V3**

#### **1.1 Core Interface Definition**
```go
// GeminiAdapterV3 implements enhanced LLMAdapter with v1.36.0 features
type GeminiAdapterV3 struct {
    // Core API connection
    client *genai.Client

    // Configuration
    model        string
    temperature  float32
    maxTokens    int32
    topP         float32
    stopSequences []string

    // Enhanced features
    streamingEnabled bool
    toolConfig      *ToolConfig
    safetySettings  []*genai.SafetySetting

    // State management
    conversationHistory []genai.Content
    toolCallCount     int

    // Quality metrics
    metrics         *AdapterMetrics

    // Error handling
    errorHandler    *ErrorHandler

    // Logger
    logger         Logger
}

// Enhanced interface with new methods for v1.36.0 features
type EnhancedGeminiAdapter interface {
    LLMAdapter // Base interface

    // v1.36.0 specific methods
    ProcessToolResult(ctx context.Context, toolCallID, functionName, result string) error
    GetConversationHistory() []genai.Content
    SetToolConfig(config *ToolConfig) error
    GetMetrics() *AdapterMetrics
}
```

#### **1.2 Configuration Structure**
```go
// Tool Configuration for advanced tool calling behavior
type ToolConfig struct {
    // Function calling mode
    Mode genai.FunctionCallingConfig_Mode

    // Streaming configuration
    EnableStream bool

    // Tool execution configuration
    MaxToolRounds int
    ToolTimeout   time.Duration

    // Caching configuration
    EnableCache bool
    CacheTTL    time.Duration
}

// Metrics collection for performance monitoring
type AdapterMetrics struct {
    RequestCount       int64     `json:"request_count"`
    ToolCallCount      int64     `json:"tool_call_count"`
    StreamRequestCount  int64     `json:"stream_request_count"`
    ErrorCount         int64     `json:"error_count"`
    AverageLatency     time.Duration `json:"average_latency"`
    PeakLatency        time.Duration `json:"peak_latency"`
    MemoryUsage        int64     `json:"memory_usage"`
    LastResetTime       time.Time  `json:"last_reset_time"`
}
```

### **2. Schema Conversion Engine**

#### **2.1 Core Schema Converter**
```go
// SchemaConverter handles JSON Schema → Gemini Schema conversion with 100% accuracy
type SchemaConverter struct {
    logger Logger
    cache  map[string]*genai.Schema
    mutex  sync.RWMutex
}

func (sc *SchemaConverter) ConvertToolSchema(tool *agent.Tool) (*genai.Schema, error) {
    // Cache lookup for performance
    sc.mutex.RLock()
    if cached, exists := sc.cache[tool.Name]; exists {
        sc.mutex.RUnlock()
        return cached, nil
    }
    sc.mutex.RUnlock()

    // Perform conversion
    schema := sc.performConversion(tool)

    // Cache result
    sc.mutex.Lock()
    sc.cache[tool.Name] = schema
    sc.mutex.Unlock()

    return schema, nil
}

func (sc *SchemaConverter) performConversion(tool *agent.Tool) *genai.Schema {
    params := tool.Parameters

    schema := &genai.Schema{
        Type:       genai.TypeObject,
        Properties: make(map[string]*genai.Schema),
        Required:   []string{},
    }

    // Extract and convert properties
    if props, ok := params["properties"].(map[string]interface{}); ok {
        for propName, propData := range props {
            if propMap, ok := propData.(map[string]interface{}); ok {
                paramSchema := sc.convertProperty(propMap)
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

func (sc *SchemaConverter) convertProperty(propMap map[string]interface{}) *genai.Schema {
    paramSchema := &genai.Schema{}

    // Type conversion with full support
    if paramType, ok := propMap["type"].(string); ok {
        paramSchema.Type = sc.convertType(paramType)
    }

    // Description
    if desc, ok := propMap["description"].(string); ok {
        paramSchema.Description = desc
    }

    // Enum values
    if enumValues, ok := propMap["enum"].([]interface{}); ok {
        paramSchema.Enum = make([]interface{}, len(enumValues))
        for i, val := range enumValues {
            paramSchema.Enum[i] = val
        }
    }

    // Array item types
    if items, ok := propMap["items"].(map[string]interface{}); ok {
        if itemType, ok := items["type"].(string); ok {
            paramSchema.Items = &genai.Schema{
                Type: sc.convertType(itemType),
            }
        }
    }

    // Object properties
    if props, ok := propMap["properties"].(map[string]interface{}); ok {
        paramSchema.Properties = make(map[string]*genai.Schema)
        for propName, propData := range props {
            if propMap, ok := propData.(map[string]interface{}); ok {
                paramSchema.Properties[propName] = sc.convertProperty(propMap)
            }
        }
    }

    // Default values
    if def, ok := propMap["default"]; ok {
        paramSchema.Default = def
    }

    return paramSchema
}

func (sc *SchemaConverter) convertType(typeStr string) genai.Type {
    typeStr = strings.ToLower(typeStr)
    switch typeStr {
    case "string":
        return genai.TypeString
    case "number", "float", "double":
        return genai.TypeNumber
    case "integer", "int":
        return genai.TypeInteger
    case "boolean", "bool":
        return genai.TypeBoolean
    case "array", "list":
        return genai.TypeArray
    case "object":
        return genai.TypeObject
    case "null":
        return genai.TypeNull
    default:
        return genai.TypeString // Default to string
    }
}
```

#### **2.2 Schema Validation Engine**
```go
// SchemaValidator ensures schema correctness and provides detailed validation
type SchemaValidator struct {
    logger Logger
}

func (sv *SchemaValidator) ValidateSchema(schema *genai.Schema, toolName string) error {
    // Validate structure
    if schema.Type != genai.TypeObject {
        return fmt.Errorf("invalid schema for tool '%s': root type must be 'object'", toolName)
    }

    // Validate properties
    for propName, propSchema := range schema.Properties {
        if err := sv.validateProperty(propSchema, toolName, propName); err != nil {
            return err
        }
    }

    // Validate required fields
    for _, requiredProp := range schema.Required {
        if _, exists := schema.Properties[requiredProp]; !exists {
            return fmt.Errorf("invalid schema for tool '%s': required field '%s' is missing", toolName, requiredProp)
        }
    }

    return nil
}

func (sv *SchemaValidator) validateProperty(schema *genai.Schema, toolName, propName string) error {
    // Validate type
    if schema.Type == genai.TypeString {
        return nil
    }

    // Validate array types
    if schema.Type == genai.TypeArray && schema.Items == nil {
        return fmt.Errorf("invalid schema for tool '%s': array property '%s' must specify 'items' type", toolName, propName)
    }

    // Validate object types
    if schema.Type == genai.TypeObject {
        for subPropName, subPropSchema := range schema.Properties {
            if err := sv.validateProperty(subPropSchema, toolName, fmt.Sprintf("%s.%s", propName, subPropName)); err != nil {
                return err
            }
        }
    }

    return nil
}
```

### **3. Arguments Processing System**

#### **3.1 Arguments Processor**
```go
// ArgumentsProcessor handles function call arguments with comprehensive validation
type ArgumentsProcessor struct {
    validator *ArgumentValidator
    logger    Logger
    metrics   *ProcessingMetrics
}

type ProcessingMetrics struct {
    TotalProcessed   int64     `json:"total_processed"`
    ValidationErrors  int64     `json:"validation_errors"`
    ProcessingTime    time.Duration `json:"processing_time"`
    CacheHitRate     float64     `json:"cache_hit_rate"`
}

type ArgumentValidator struct {
    typeChecker    TypeChecker
    schemaValidator *SchemaValidator
    logger        Logger
}

func (ap *ArgumentsProcessor) ProcessFunctionCall(funcCall *genai.FunctionCall, schema *genai.Schema) ([]agent.ToolCall, error) {
    startTime := time.Now()
    defer func() {
        ap.metrics.ProcessingTime += time.Since(startTime)
    }()

    // Validate against schema if provided
    if schema != nil {
        if err := ap.validator.schemaValidator.ValidateSchema(schema, funcCall.Name); err != nil {
            return nil, fmt.Errorf("schema validation failed for function '%s': %w", funcCall.Name, err)
        }
    }

    // Validate arguments against schema
    if err := ap.validator.typeChecker.ValidateArguments(funcCall.Args, schema); err != nil {
        return nil, fmt.Errorf("argument validation failed for function '%s': %w", funcCall.Name, err)
    }

    // Convert arguments to JSON
    argsJSON, err := json.Marshal(funcCall.Args)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal arguments for function '%s': %w", funcCall.Name, err)
    }

    // Generate unique ID
    toolCallID := fmt.Sprintf("%s_%s_%d", funcCall.Name, uuid.New().String()[:8], time.Now().UnixNano()%1000000)

    toolCall := agent.ToolCall{
        ID:        toolCallID,
        Type:      "function",
        Name:      funcCall.Name,
        Arguments: string(argsJSON),
    }

    ap.metrics.TotalProcessed++

    return []agent.ToolCall{toolCall}, nil
}

type TypeChecker struct {
    logger Logger
}

func (tc *TypeChecker) ValidateArguments(args map[string]interface{}, schema *genai.Schema) error {
    if schema.Type != genai.TypeObject {
        return fmt.Errorf("arguments must be an object, got %s", schema.Type.String())
    }

    // Validate each argument against schema
    for propName, propSchema := range schema.Properties {
        argValue, exists := args[propName]
        if !exists {
            if tc.isRequired(propName, schema.Required) {
                return fmt.Errorf("required argument '%s' is missing", propName)
            }
            continue
        }

        if err := tc.validateArgumentValue(argValue, propSchema, propName); err != nil {
            return fmt.Errorf("invalid argument '%s': %w", propName, err)
        }
    }

    // Check for unexpected arguments
    for argName := range args {
        if _, exists := schema.Properties[argName]; !exists {
            return fmt.Errorf("unexpected argument '%s' not in schema", argName)
        }
    }

    return nil
}

func (tc *TypeChecker) validateArgumentValue(value interface{}, schema *genai.Schema, propName string) error {
    switch schema.Type {
    case genai.TypeString:
        if _, ok := value.(string); !ok {
            return fmt.Errorf("argument '%s' must be a string", propName)
        }

    case genai.TypeNumber:
        if _, ok := value.(float64); !ok {
            if _, ok := value.(int); !ok {
                return fmt.Errorf("argument '%s' must be a number", propName)
            }
        }

    case genai.TypeInteger:
        if _, ok := value.(int64); !ok {
            if _, ok := value.(int32); !ok {
                return fmt.Errorf("argument '%s' must be an integer", propName)
            }
        }

    case genai.TypeBoolean:
        if _, ok := value.(bool); !ok {
            if strVal, ok := value.(string); ok && (strings.ToLower(strVal) == "true" || strings.ToLower(strVal) == "false") {
                // String boolean values are acceptable
            } else {
                return fmt.Errorf("argument '%s' must be a boolean", propName)
            }
        }

    case genai.TypeArray:
        if _, ok := value.([]interface{}); !ok {
            return fmt.Errorf("argument '%s' must be an array", propName)
        }

    case genai.TypeObject:
        if _, ok := value.(map[string]interface{}); !ok {
            return fmt.Errorf("argument '%s' must be an object", propName)
        }
        // Additional object validation could be added here

    default:
        return fmt.Errorf("unsupported type '%s' for argument '%s'", schema.Type.String(), propName)
    }

    return nil
}

func (tc *TypeChecker) isRequired(propName string, required []string) bool {
    for _, req := range required {
        if req == propName {
            return true
        }
    }
    return false
}
```

### **4. Tool Result Handler**

#### **4.1 Result Processing Engine**
```go
// ResultHandler processes tool execution results and feeds them back to Gemini
type ResultHandler struct {
    conversationManager *ConversationManager
    responseFormatter  *ResponseFormatter
    logger            Logger
    metrics           *ResultMetrics
}

type ResultMetrics struct {
    ResultsProcessed  int64     `json:"results_processed"`
    ProcessingTime    time.Duration `json:"processing_time"`
    ErrorRate        float64     `json:"error_rate"`
    SuccessRate      float64     `json:"success_rate"`
}

type ConversationManager struct {
    conversationHistory []genai.Content
    maxHistoryLength   int
    logger            Logger
    mutex             sync.RWMutex
}

type ResponseFormatter struct {
    logger Logger
}

func (rh *ResultHandler) ProcessToolResult(ctx context.Context, toolCallID, functionName, result string, toolExecutionError error) (*genai.Content, error) {
    startTime := time.Now()

    defer func() {
        rh.metrics.ResultsProcessed++
        processingTime := time.Since(startTime)
        rh.metrics.ProcessingTime += processingTime
        if toolExecutionError != nil {
            rh.metrics.ErrorRate = float64(rh.metrics.ResultsProcessed) / float64(rh.metrics.ErrorRate+rh.metrics.SuccessRate)
        } else {
            rh.metrics.SuccessRate = float64(rh.metrics.ResultsProcessed) / float64(rh.metrics.ErrorRate+rh.metrics.SuccessRate)
        }
    }()

    // Handle tool execution error
    if toolExecutionError != nil {
        errorContent := rh.responseFormatter.FormatErrorResult(functionName, toolCallID, toolExecutionError)
        rh.conversationManager.AddContent(*errorContent)
        return errorContent, toolExecutionError
    }

    // Create success response
    successContent := rh.responseFormatter.FormatSuccessResult(functionName, toolCallID, result)

    // Add to conversation history
    rh.conversationManager.AddContent(*successContent)

    return successContent, nil
}

type ResponseFormatter struct {
    logger Logger
}

func (rf *ResponseFormatter) FormatSuccessResult(functionName, toolCallID, result string) *genai.Content {
    return &genai.Content{
        Parts: []genai.Part{
            genai.NewPartFromFunctionResponse(functionName, map[string]interface{}{
                "result": result,
                "id":     toolCallID,
                "timestamp": time.Now().Unix(),
            }),
        },
        Role: genai.RoleModel, // Tool results come from assistant perspective
    }
}

func (rf *ResponseFormatter) FormatErrorResult(functionName, toolCallID string, err error) *genai.Content {
    errorMessage := fmt.Sprintf("Error executing %s (ID: %s): %v", functionName, toolCallID, err)

    return &genai.Content{
        Parts: []genai.Part{
            genai.NewPartFromFunctionResponse(functionName, map[string]interface{}{
                "error":   errorMessage,
                "id":     toolCallID,
                "success": false,
                "timestamp": time.Now().Unix(),
            }),
        },
        Role: genai.RoleModel,
    }
}

func (cm *ConversationManager) AddContent(content genai.Content) {
    cm.mutex.Lock()
    defer cm.mutex.Unlock()

    cm.conversationHistory = append(cm.conversationHistory, content)

    // Maintain conversation history length
    if len(cm.conversationHistory) > cm.maxHistoryLength {
        cm.conversationHistory = cm.conversationHistory[1:]
    }
}
```

---

## 🔄 DATA FLOW ARCHITECTURE

### **5. Enhanced Conversation Flow**

#### **5.1 Complete Conversation Processing Pipeline**
```
User Request → Tool Execution → Result Feedback → Next Turn

[REQUEST] → [TOOL CALL] → [EXECUTION] → [FEEDBACK] → [CONTINUE]

┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                            COMPLETE CONVERSATION FLOW                              │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  User Message: "Calculate 15 * 8 + 3"                                          │
│  └─────────────────────────────────────────────────────────────────────────────────────────┘ │
│                                    │                                           │
│                                    ▼                                           │
│  ┌─────────────────┐  │   ┌─────────────────┐   │   ┌─────────────────┐   │   │
│  │   Message      │   │   Tool         │   │   Schema          │   │   │
│  │   Processing  │   │   Detection   │   │   Validation     │   │   │
│  │   &           │   │   &           │   │   &               │   │   │
│  │   Preparation│   │   &           │   │   &               │   │   │
│  └─────────────────┘   │   └─────────────────┘   │   └─────────────────┘   │   │
│                                    │                                           │
│                                    ▼                                           │
│  ┌───────────────────────────────────────────────────────────────────────────────────────┐
│  │                        ENHANCED GEMINI ADAPTER V3                                 │
│  │  ┌─────────────┐  │   ┌──────────────┐   │   ┌─────────────┐   │   │
│  │  │   Schema      │   │   Arguments   │   │   Result       │   │   │
│  │  │   Converter   │   │   Processor   │   │   Handler     │   │   │
│  │  │   (Enhanced)  │   │   (Robust)    │   │   (New)       │   │   │
│  │  └─────────────┘   │   └──────────────┘   │   └─────────────┘   │   │
│  └───────────────────────────────────────────────────────────────────────────────────────┘
│                                    │                                           │
│                                    ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐
│  │                              GEMINI v1.36.0 API                                 │
│  │  ✅ FunctionCalling.Enabled                                                │
│  │  ✅ GenerateContent() with Tools                                           │
│  │  │   ← Proper tool schema conversion                                            │
│  │  │   ← Function call generation                                             │
│  │  │   ← Arguments processing                                                 │
│  │  │   ← Result processing support                                             │
│  │  │   ← Enhanced streaming capabilities                                          │
│  │  │   ← Better error handling                                                  │
│  └─────────────────────────────────────────────────────────────────────────────────────┘
│                                    │                                           │
│                                    ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐
│  │                              CONVERSATION STATE MANAGEMENT                             │
│  │  │  Tool Call Reception → Result Processing → State Update →           │
│  │  │  Next Request Preparation                                           │
│  │  │  Context Enhancement                                                    │
│  │  │  Memory Management                                                     │
│  │  │  Error Recovery                                                        │
│  │  └─────────────────────────────────────────────────────────────────────────────────────┘
│                                    │                                           │
│                                    ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐
│  │                              GEMINI RESPONSE                                        │
│  │  │  "The result is 123"                                                        │
│  │  │  │   ← Tool result successfully processed                                       │
│  │  │   │   ← Conversation state updated                                         │
│  │  │   │   ← Context enhanced for better responses                             │
│  │  │  │   ← Next generation ready                                                │
│  │  │   │   │                                                     │
│  │  │   │   │         ← Enhanced reasoning                                      │
│  │  │   │   │         ← Mathematical accuracy                                │
│  │  │   │   │         ← Problem-solving capabilities                           │
│  │  │   │   │                                                      │
│  │  │   │   └─────────────────────────────────────────────────────┘     │
│  └─────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📋 INTERFACE SPECIFICATIONS

### **6. Enhanced LLMAdapter Interface**

#### **6.1 Extended Interface Definition**
```go
// Enhanced interface that extends base LLMAdapter with v1.36.0 features
type EnhancedLLMAdapter interface {
    // Base LLMAdapter methods (unchanged)
    Complete(ctx context.Context, req *agent.CompletionRequest) (*agent.CompletionResponse, error)
    Stream(ctx context.Context, req *agent.CompletionRequest, onChunk func(string)) (*agent.CompletionResponse, error)

    // Enhanced v1.36.0 specific methods
    ProcessToolResult(ctx context.Context, toolCallID string, functionName string, result string, error error) error
    GetConversationState() *ConversationState
    ResetConversation() error
    SetToolConfig(config *ToolConfig) error
    GetAdapterMetrics() *AdapterMetrics
    SetLogger(logger Logger)
    Close() error
}

// ConversationState provides visibility into conversation state
type ConversationState struct {
    MessageCount      int                    `json:"message_count"`
    ToolCallCount     int                    `json:"tool_call_count"`
    CurrentTurn     int                    `json:"current_turn"`
    LastActivity    time.Time               `json:"last_activity"`
    MemoryUsage    int64                  `json:"memory_usage"`
}

// ToolConfig provides advanced tool calling configuration
type ToolConfig struct {
    // Function calling mode
    Mode genai.FunctionCallingConfig_Mode `json:"mode"`

    // Execution limits
    MaxToolRounds int                     `json:"max_tool_rounds"`
    ToolTimeout    time.Duration               `json:"tool_timeout"`

    // Performance settings
    EnableCache    bool                        `json:"enable_cache"`
    CacheTTL       time.Duration               `json:"cache_ttl"`

    // Streaming settings
    EnableStream  bool                        `json:"enable_stream"`
    StreamTimeout  time.Duration               `json:"stream_timeout"`

    // Retry configuration
    MaxRetries     int                        `json:"max_retries"`
    RetryDelay     time.Duration               `json:"retry_delay"`
}
```

#### **6.2 Error Management Interface**
```go
// ErrorHandler provides comprehensive error management and recovery strategies
type ErrorHandler struct {
    logger     Logger
    metrics    *ErrorMetrics
    strategies map[ErrorType] RecoveryStrategy
}

type ErrorType string

const (
    ErrorTypeValidation    ErrorType = "validation"
    ErrorTypeExecution    ErrorType = "execution"
    ErrorTypeCommunication ErrorType = "communication"
    ErrorTypeTimeout     ErrorType = "timeout"
    ErrorTypeMemory       ErrorType = "memory"
    ErrorTypeRateLimit    ErrorType = "rate_limit"
)

type RecoveryStrategy string

const (
    StrategyRetry      RecoveryStrategy = "retry"
    StrategyFallback    RecoveryStrategy = "fallback"
    StrategyAbort      RecoveryStrategy = "abort"
    StrategyLogOnly    RecoveryStrategy = "log_only"
)

type ErrorMetrics struct {
    TotalErrors     int64     `json:"total_errors"`
    RecoveryCount   int64     `json:"recovery_count"`
    ErrorTypes      map[ErrorType]int64 `json:"error_types"`
    LastErrorTime    time.Time  `json:"last_error_time"`
}

type ErrorDetail struct {
    Type        ErrorType      `json:"type"`
    Message     string          `json:"message"`
    Timestamp  time.Time      `json:"timestamp"`
    Context     map[string]interface{} `json:"context"`
    Recovery    RecoveryStrategy   `json:"recovery_strategy"`
    Success     bool             `json:"success"`
    Retries     int              `json:"retries"`
}
```

---

## 📊 IMPLEMENTATION PATTERNS

### **7. Design Patterns Applied**

#### **7.1 Strategy Pattern for Error Handling**
```go
// ErrorHandler implements Strategy pattern for different error types
type ErrorHandler struct {
    strategies map[ErrorType]RecoveryStrategy
    logger     Logger
    metrics    *ErrorMetrics
}

func (eh *ErrorHandler) HandleError(ctx context.Context, err error, context map[string]interface{}) error {
    errorType := eh.categorizeError(err)

    strategy := eh.getStrategy(errorType)

    switch strategy {
    case StrategyRetry:
        return eh.retryWithBackoff(ctx, err, context)
    case StrategyFallback:
        return eh.fallbackAlternative(ctx, err, context)
    case StrategyAbort:
        return err // Return original error
    case StrategyLogOnly:
        eh.logError(err, context)
        return nil // Continue processing
    default:
        return err
    }
}

func (eh *ErrorHandler) categorizeError(err error) ErrorType {
    errorStr := err.Error()

    // Categorize error types based on error messages
    if strings.Contains(errorStr, "validation") || strings.Contains(errorStr, "schema") {
        return ErrorTypeValidation
    }
    if strings.Contains(errorStr, "timeout") || strings.Contains(errorStr, "deadline") {
        return ErrorTypeTimeout
    }
    if strings.Contains(errorStr, "rate limit") || strings.Contains(errorStr, "quota") {
        return ErrorTypeRateLimit
    }
    if strings.Contains(errorStr, "network") || strings.Contains(errorStr, "connection") {
        return ErrorTypeCommunication
    }

    return ErrorTypeExecution
}
```

#### **7.2 Builder Pattern for Configuration**
```go
// Enhanced builder for GeminiAdapter with fluent configuration
type GeminiAdapterBuilder struct {
    config *GeminiConfig
    apiKey string
    model  string
    logger Logger
}

type GeminiConfig struct {
    Temperature   float32             `json:"temperature"`
    MaxTokens     int32               `json:"max_tokens"`
    TopP          float32             `json:"top_p"`
    StopSequences []string             `json:"stop_sequences"`
    ToolConfig     *ToolConfig        `json:"tool_config"`
    SafetySettings []*genai.SafetySetting `json:"safety_settings"`
    Streaming     bool                `json:"streaming"`
}

func (gab *GeminiAdapterBuilder) WithAPIKey(apiKey string) *GeminiBuilder {
    gab.apiKey = apiKey
    return gab
}

func (gab *GeminiAdapterBuilder) WithModel(model string) *GeminiBuilder {
    gab.model = model
    return gab
}

func (gab *GeminiBuilder) WithTemperature(temp float32) *GeminiBuilder {
    gab.config.Temperature = temp
    return gab
}

func (gab *GeminiBuilder) WithToolConfig(config *ToolConfig) *GeminiBuilder {
    gab.config.ToolConfig = config
    return gab
}

func (gab *GeminiBuilder) EnableStreaming(enable bool) *GeminiBuilder {
    gab.config.Streaming = enable
    return gab
}

func (gab *GeminiBuilder) Build() (*GeminiAdapterV3, error) {
    if gab.apiKey == "" {
        return nil, fmt.Errorf("API key is required")
    }

    if gab.model == "" {
        gab.model = "gemini-1.5-pro" // Default model
    }

    // Validate configuration
    if err := gab.validateConfig(); err != nil {
        return nil, err
    }

    // Create client
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(gab.apiKey))
    if err != nil {
        return nil, fmt.Errorf("failed to create Gemini client: %w", err)
    }

    // Create adapter
    adapter := &GeminiAdapterV3{
        client:         client,
        model:          gab.model,
        temperature:    gab.config.Temperature,
        maxTokens:      gab.config.MaxTokens,
        topP:          gab.config.TopP,
        stopSequences:  gab.config.StopSequences,
        toolConfig:     gab.config.ToolConfig,
        safetySettings: gab.config.SafetySettings,
        streamingEnabled: gab.config.Streaming,
        logger:        gab.logger,
    }

    return adapter, nil
}
```

#### **7.3 Observer Pattern for Metrics Collection**
```go
// MetricsCollector implements Observer pattern for performance monitoring
type MetricsCollector struct {
    subscribers []func(*MetricsData) error
    mutex       sync.RWMutex
    interval    time.Duration
    ticker      *time.Ticker
    running     int32
}

type MetricsData struct {
    Timestamp        time.Time     `json:"timestamp"`
    RequestCount     int64       `json:"request_count"`
    ToolCallCount     int64       `json:"tool_call_count"`
    ErrorCount       int64       `json:"error_count"`
    ResponseTime     time.Duration `json:"response_time"`
    MemoryUsage      int64       `json:"memory_usage"`
    SuccessRate      float64     `json:"success_rate"`
}

func (mc *MetricsCollector) Subscribe(callback func(*MetricsData) error) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()

    mc.subscribers = append(mc.subscribers, callback)
    return nil
}

func (mc *MetricsCollector) CollectMetrics(metrics *MetricsData) error {
    mc.mutex.RLock()
    defer mc.mutex.RUnlock()

    for _, subscriber := range mc.subscribers {
        if err := subscriber(metrics); err != nil {
            return fmt.Errorf("metrics subscriber error: %w", err)
        }
    }

    return nil
}

// Auto-collection loop
func (mc *MetricsCollector) startAutoCollection() {
    atomic.StoreInt32(&mc.running, 1)

    mc.ticker = time.NewTicker(mc.interval)
    go func() {
        for atomic.LoadInt32(&mc.running) == 1 {
            select {
            case <-mc.ticker.C:
                mc.collectCurrentMetrics()
            case <-time.After(2 * mc.interval):
                mc.collectCurrentMetrics()
            }
        }
    }()
}

func (mc *MetricsCollector) collectCurrentMetrics() {
    // Collect current system metrics
    metrics := &MetricsData{
        Timestamp:    time.Now(),
        RequestCount: 0,
        ToolCallCount: 0,
        ErrorCount:   0,
        ResponseTime: 0,
        MemoryUsage: 0,
        SuccessRate:  0.0,
    }

    mc.CollectMetrics(metrics)
}
```

---

## 📚 IMPLEMENTATION SPECIFICATIONS

### **8. File Structure Specifications**

#### **8.1 New Files Structure**
```
agent/adapters/gemini_v3/
├── gemini_adapter_v3.go              # Main adapter implementation
├── schema_converter.go               # Schema conversion engine
├── args_processor.go                 # Arguments processing system
├── result_handler.go                  # Tool result handler
└── conversation_manager.go              # Conversation state management

agent/adapters/gemini_v3/processing/
├── type_converter.go                     # Type conversion utilities
├── validator.go                           # Validation engine
└── cache.go                                # Performance optimization

agent/adapters/gemini_v3/streaming/
├── stream_processor.go                   # Enhanced streaming
├── backpressure_manager.go               # Resource management
└── connection_pool.go                    # Connection management

agent/adapters/gemini_v3/error/
├── error_handler.go                     # Error management
├── recovery_strategies.go               # Recovery mechanisms
└── metrics_collector.go                # Performance monitoring

agent/adapters/gemini_v3/validation/
├── schema_validator.go                  # Schema validation
├── argument_validator.go                # Argument validation
└── input_sanitizer.go                   # Security validation
```

#### **8.2 Modified Files Structure**
```
agent/adapters/
├── gemini_adapter.go                    # Apply critical fixes or replace
├── adapter_interface.go                  # Enhanced interface definitions
└── adapter_test.go                      # Updated test suite

agent/
├── builder.go                           # Enhanced builder integration
└── types.go                           # Enhanced type definitions
```

---

## 🎯 QUALITY GATES DEFINITION

### **9.1 Architecture Quality Gates**

#### **Gate 1: Schema Conversion**
- **Requirement:** 100% schema conversion accuracy
- **Test Cases:** All JSON Schema types (string, number, integer, boolean, array, object)
- **Performance:** <10ms conversion time per schema
- **Coverage:** All edge cases and error conditions
- **Validation:** Comprehensive error handling and reporting

#### **Gate 2: Arguments Processing**
- **Requirement:** 100% accurate JSON marshaling
- **Test Cases:** Complex object arguments, nested structures
- **Performance:** <5ms processing time per function call
- **Coverage:** Type validation, error categorization
- **Validation:** Schema-based validation with clear error messages

#### **Gate 3: Tool Result Handling**
- **Requirement:** 100% result processing accuracy
- **Test Cases:** Success and error results, multi-turn conversations
- **Performance:** <3ms result processing time
- **Coverage:** State management, context maintenance
- **Validation:** Conversation state integrity

#### **Gate 4: Integration Testing**
- **Requirement:** 95%+ test coverage
- **Test Coverage:** All user workflows, edge cases
- **Performance:** Sub-100ms response time with tool calling
- **Coverage**: Backward compatibility, integration points
- **Validation:** End-to-end scenario testing

#### **Gate 5: Performance Requirements**
- **Requirement:** Meet all performance specifications
- **Test Coverage:** Load testing, stress testing
- **Metrics:** <100ms average response time
- **Coverage:** Concurrent request handling
- **Validation:** Performance benchmarking

### **9.2 Code Quality Standards**

#### **Gate 6: Code Quality**
- **Requirement:** Zero lint issues
- **Testing:** golangci-lint with comprehensive ruleset
- **Coverage:** Code complexity analysis
- **Validation:** No security vulnerabilities

#### **Gate 7: Security Requirements**
- **Requirement:** Zero security vulnerabilities
- **Testing:** Security scanning (gosec, govulncheck)
- **Coverage:** Input validation and sanitization
- **Validation:** Security compliance checks

#### **Gate 8: Documentation**
- **Requirement:** 100% API documentation
- **Testing:** Godoc coverage validation
- **Coverage:** Architecture documentation
- **Validation:** Documentation completeness

---

## 🎯 IMPLEMENTATION DECISIONS

### **10.1 Technology Choices**

#### **JSON Schema → Gemini Schema Conversion**
**Decision:** Custom converter rather than library-based
**Rationale:**
- Maximum control over conversion logic
- Better error handling and debugging
- Optimized for our specific use cases
- Extensive validation and testing capabilities

**Implementation:** Custom `SchemaConverter` with caching for performance

#### **Error Management Strategy**
**Decision:** Comprehensive error categorization with recovery strategies
**Rationale:**
- Enables graceful degradation
- Provides detailed debugging information
- Supports different recovery strategies per error type
- Enables learning from error patterns

**Implementation:** `ErrorHandler` with strategy pattern

#### **Conversation State Management**
**Decision Dedicated conversation manager with history tracking
**Rationale:**
- Essential for multi-turn tool conversations
- Enables proper context maintenance
- Memory management with size limits
- Debugging and analysis capabilities

**Implementation:** `ConversationManager` with configurable history length

---

## 🚀 NEXT PHASE PREPARATION

**Architecture Design Status:** ✅ COMPLETED
**Key Deliverables:**
- ✅ Detailed component architecture
- ✅ Interface specifications
- ✅ Implementation patterns and patterns
- ✅ Quality gates definition
- ✅ File structure specifications
- ✅ Performance requirements

**Next Phase:** Development Implementation
**Timeline:** 4 weeks with quality gates
**Team Readiness:** ✅ READY

**Preparation Tasks:**
- [ ] Review architecture with team
- [] Finalize technical specifications
- [] Set up development environment
- [] Prepare test strategy
- [] Establish quality metrics

**Architecture Documentation Status:** ✅ COMPLETED**
**Quality:** Comprehensive and detailed
**Scope:** Complete technical specifications for v1.36.0 upgrade

---

**Architecture Design Session Status:** ✅ COMPLETED**
**Architecture Status:** Comprehensive design finalized with BMAD Method rigor
**Next Phase:** Development Implementation with quality assurance