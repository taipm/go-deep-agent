package agent

import (
	"encoding/json"

	"github.com/openai/openai-go/v3"
)

// Tool represents a function that the LLM can call.
// This is a simplified wrapper around OpenAI's tool definition.
type Tool struct {
	Name        string                            // Function name
	Description string                            // What the function does
	Parameters  map[string]interface{}            // JSON schema for parameters
	Handler     func(args string) (string, error) // Function implementation
}

// NewTool creates a new tool with the given name and description.
// You can then add parameters using AddParameter() or set them directly.
//
// Example:
//
//	tool := agent.NewTool("get_weather", "Get weather for a location").
//	    AddParameter("location", "string", "City name", true)
func NewTool(name, description string) *Tool {
	return &Tool{
		Name:        name,
		Description: description,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
	}
}

// AddParameter adds a parameter to the tool's schema.
//
// Example:
//
//	tool.AddParameter("location", "string", "The city name", true).
//	    AddParameter("units", "string", "Temperature units (celsius/fahrenheit)", false)
func (t *Tool) AddParameter(name, paramType, description string, required bool) *Tool {
	props := t.Parameters["properties"].(map[string]interface{})
	props[name] = map[string]interface{}{
		"type":        paramType,
		"description": description,
	}

	if required {
		reqs := t.Parameters["required"].([]string)
		t.Parameters["required"] = append(reqs, name)
	}

	return t
}

// WithHandler sets the function handler for this tool.
// The handler receives the arguments as a JSON string and should return a result string.
//
// Example:
//
//	tool.WithHandler(func(args string) (string, error) {
//	    var params struct {
//	        Location string `json:"location"`
//	    }
//	    json.Unmarshal([]byte(args), &params)
//	    return fmt.Sprintf("Weather in %s: Sunny, 25°C", params.Location), nil
//	})
func (t *Tool) WithHandler(handler func(string) (string, error)) *Tool {
	t.Handler = handler
	return t
}

// toOpenAI converts our Tool to OpenAI's ChatCompletionToolUnionParam format.
func (t *Tool) toOpenAI() openai.ChatCompletionToolUnionParam {
	// Create function parameters from our schema
	var funcParams openai.FunctionParameters
	paramsJSON, _ := json.Marshal(t.Parameters)
	json.Unmarshal(paramsJSON, &funcParams)

	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        t.Name,
		Description: openai.String(t.Description),
		Parameters:  funcParams,
	})
}

// Common tool parameter helpers

// StringParam creates a string parameter definition.
func StringParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

// NumberParam creates a number parameter definition.
func NumberParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "number",
		"description": description,
	}
}

// BoolParam creates a boolean parameter definition.
func BoolParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

// ArrayParam creates an array parameter definition.
func ArrayParam(description, itemType string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items": map[string]interface{}{
			"type": itemType,
		},
	}
}

// EnumParam creates an enum parameter definition.
func EnumParam(description string, values ...string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
		"enum":        values,
	}
}
