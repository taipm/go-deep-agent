package agent

import (
	"encoding/json"
	"testing"
)

// TestNewTool tests the NewTool factory function
func TestNewTool(t *testing.T) {
	tool := NewTool("test_tool", "A test tool")

	if tool.Name != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", tool.Name)
	}
	if tool.Description != "A test tool" {
		t.Errorf("Expected description 'A test tool', got '%s'", tool.Description)
	}
	if tool.Parameters == nil {
		t.Error("Expected Parameters to be initialized")
	}
	if tool.Handler != nil {
		t.Error("Expected Handler to be nil initially")
	}
}

// TestAddParameter tests adding parameters to a tool
func TestAddParameter(t *testing.T) {
	tool := NewTool("calc", "Calculator").
		AddParameter("operation", "string", "The operation to perform", true).
		AddParameter("value", "number", "The value", false)

	// Check parameters exist
	props, ok := tool.Parameters["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	// Check operation parameter
	if _, exists := props["operation"]; !exists {
		t.Error("Expected 'operation' parameter to exist")
	}

	// Check required array
	required, ok := tool.Parameters["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be []string")
	}
	if len(required) != 1 || required[0] != "operation" {
		t.Errorf("Expected required=['operation'], got %v", required)
	}
}

// TestAddParameterChaining tests method chaining
func TestAddParameterChaining(t *testing.T) {
	tool := NewTool("test", "Test tool").
		AddParameter("a", "string", "First param", true).
		AddParameter("b", "number", "Second param", true).
		AddParameter("c", "boolean", "Third param", false)

	props := tool.Parameters["properties"].(map[string]interface{})
	required := tool.Parameters["required"].([]string)

	if len(props) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(props))
	}
	if len(required) != 2 {
		t.Errorf("Expected 2 required parameters, got %d", len(required))
	}
}

// TestWithHandler tests setting a handler function
func TestWithHandler(t *testing.T) {
	handlerCalled := false
	tool := NewTool("test", "Test").WithHandler(func(args string) (string, error) {
		handlerCalled = true
		return "result", nil
	})

	if tool.Handler == nil {
		t.Fatal("Expected Handler to be set")
	}

	// Test handler execution
	result, err := tool.Handler("{}")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "result" {
		t.Errorf("Expected 'result', got '%s'", result)
	}
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

// TestToOpenAI tests conversion to OpenAI format
func TestToOpenAI(t *testing.T) {
	tool := NewTool("get_weather", "Get weather for a location").
		AddParameter("location", "string", "City name", true).
		AddParameter("units", "string", "Temperature units", false)

	// Test that toOpenAI doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("toOpenAI panicked: %v", r)
		}
	}()

	openAITool := tool.toOpenAI()
	_ = openAITool // Just verify it doesn't panic
}

// TestStringParam tests StringParam helper
func TestStringParam(t *testing.T) {
	param := StringParam("test description")

	if param["type"] != "string" {
		t.Errorf("Expected type 'string', got '%v'", param["type"])
	}
	if param["description"] != "test description" {
		t.Errorf("Expected description 'test description', got '%v'", param["description"])
	}
}

// TestNumberParam tests NumberParam helper
func TestNumberParam(t *testing.T) {
	param := NumberParam("test number")

	if param["type"] != "number" {
		t.Errorf("Expected type 'number', got '%v'", param["type"])
	}
	if param["description"] != "test number" {
		t.Errorf("Expected description 'test number', got '%v'", param["description"])
	}
}

// TestBoolParam tests BoolParam helper
func TestBoolParam(t *testing.T) {
	param := BoolParam("test bool")

	if param["type"] != "boolean" {
		t.Errorf("Expected type 'boolean', got '%v'", param["type"])
	}
	if param["description"] != "test bool" {
		t.Errorf("Expected description 'test bool', got '%v'", param["description"])
	}
}

// TestArrayParam tests ArrayParam helper
func TestArrayParam(t *testing.T) {
	param := ArrayParam("test array", "string")

	if param["type"] != "array" {
		t.Errorf("Expected type 'array', got '%v'", param["type"])
	}
	items, ok := param["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected items to be map[string]interface{}")
	}
	if items["type"] != "string" {
		t.Errorf("Expected items type 'string', got '%v'", items["type"])
	}
}

// TestEnumParam tests EnumParam helper
func TestEnumParam(t *testing.T) {
	param := EnumParam("test enum", "red", "green", "blue")

	if param["type"] != "string" {
		t.Errorf("Expected type 'string', got '%v'", param["type"])
	}
	enum, ok := param["enum"].([]string)
	if !ok {
		t.Fatal("Expected enum to be []string")
	}
	if len(enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(enum))
	}
	if enum[0] != "red" || enum[1] != "green" || enum[2] != "blue" {
		t.Errorf("Expected ['red','green','blue'], got %v", enum)
	}
}

// TestToolComplexParameters tests a tool with complex parameters
func TestToolComplexParameters(t *testing.T) {
	tool := NewTool("complex_tool", "A complex tool").
		AddParameter("name", "string", "Name", true).
		AddParameter("age", "number", "Age", true).
		AddParameter("active", "boolean", "Is active", false)

	// Verify all parameters are present
	props := tool.Parameters["properties"].(map[string]interface{})
	if len(props) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(props))
	}

	// Verify required list
	required := tool.Parameters["required"].([]string)
	if len(required) != 2 {
		t.Errorf("Expected 2 required params, got %d", len(required))
	}

	// Verify specific parameter types
	nameParam := props["name"].(map[string]interface{})
	if nameParam["type"] != "string" {
		t.Error("Expected name to be string type")
	}

	ageParam := props["age"].(map[string]interface{})
	if ageParam["type"] != "number" {
		t.Error("Expected age to be number type")
	}

	activeParam := props["active"].(map[string]interface{})
	if activeParam["type"] != "boolean" {
		t.Error("Expected active to be boolean type")
	}
}

// TestToolHandlerWithJSON tests handler with JSON arguments
func TestToolHandlerWithJSON(t *testing.T) {
	tool := NewTool("multiply", "Multiply two numbers").
		AddParameter("a", "number", "First number", true).
		AddParameter("b", "number", "Second number", true).
		WithHandler(func(args string) (string, error) {
			var params struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			result := params.A * params.B
			resultBytes, err := json.Marshal(map[string]float64{"result": result})
			if err != nil {
				return "", err
			}
			return string(resultBytes), nil
		})

	// Test the handler
	result, err := tool.Handler(`{"a": 10, "b": 5}`)
	if err != nil {
		t.Fatalf("Handler error: %v", err)
	}

	var resultMap map[string]float64
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if resultMap["result"] != 50 {
		t.Errorf("Expected result 50, got %v", resultMap["result"])
	}
}

// TestToolNoParameters tests a tool with no parameters
func TestToolNoParameters(t *testing.T) {
	tool := NewTool("get_time", "Get current time")

	props := tool.Parameters["properties"].(map[string]interface{})
	required := tool.Parameters["required"].([]string)

	if len(props) != 0 {
		t.Errorf("Expected 0 properties, got %d", len(props))
	}
	if len(required) != 0 {
		t.Errorf("Expected 0 required params, got %d", len(required))
	}
}

// TestMultipleToolsIndependence tests that multiple tools don't interfere
func TestMultipleToolsIndependence(t *testing.T) {
	tool1 := NewTool("tool1", "First tool").
		AddParameter("param1", "string", "Param 1", true)

	tool2 := NewTool("tool2", "Second tool").
		AddParameter("param2", "number", "Param 2", true)

	// Verify tool1 parameters
	props1 := tool1.Parameters["properties"].(map[string]interface{})
	if _, exists := props1["param2"]; exists {
		t.Error("tool1 should not have param2")
	}

	// Verify tool2 parameters
	props2 := tool2.Parameters["properties"].(map[string]interface{})
	if _, exists := props2["param1"]; exists {
		t.Error("tool2 should not have param1")
	}
}
