package tools

import (
	"strings"
	"testing"
)

func TestMathTool_Evaluate(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name       string
		expression string
		wantError  bool
	}{
		{"simple addition", `{"operation": "evaluate", "expression": "2 + 3"}`, false},
		{"multiplication", `{"operation": "evaluate", "expression": "2 * (3 + 4)"}`, false},
		{"sqrt function", `{"operation": "evaluate", "expression": "sqrt(16)"}`, false},
		{"pow function", `{"operation": "evaluate", "expression": "pow(2, 3)"}`, false},
		{"sin function", `{"operation": "evaluate", "expression": "sin(0)"}`, false},
		{"cos function", `{"operation": "evaluate", "expression": "cos(0)"}`, false},
		{"complex expression", `{"operation": "evaluate", "expression": "2 * (3 + 4) - sqrt(16) / pow(2, 2)"}`, false},
		{"empty expression", `{"operation": "evaluate", "expression": ""}`, true},
		{"invalid expression", `{"operation": "evaluate", "expression": "2 +"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Handler(tt.expression)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == "" {
					t.Errorf("Expected result but got empty string")
				}
			}
		})
	}
}

func TestMathTool_Statistics_Mean(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "statistics", "numbers": [1, 2, 3, 4, 5], "stat_type": "mean"}`
	result, err := tool.Handler(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Mean of [1,2,3,4,5] = 3.0
	if !strings.HasPrefix(result, "3.0") {
		t.Errorf("Expected mean ~3.0, got %s", result)
	}
}

func TestMathTool_Statistics_Median(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name    string
		numbers string
		want    string
	}{
		{"odd count", `{"operation": "statistics", "numbers": [1, 3, 5], "stat_type": "median"}`, "3.0"},
		{"even count", `{"operation": "statistics", "numbers": [1, 2, 3, 4], "stat_type": "median"}`, "2.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Handler(tt.numbers)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !strings.HasPrefix(result, tt.want) {
				t.Errorf("Expected median ~%s, got %s", tt.want, result)
			}
		})
	}
}

func TestMathTool_Statistics_MinMax(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name     string
		statType string
		want     string
	}{
		{"min", "min", "1.0"},
		{"max", "max", "10.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := `{"operation": "statistics", "numbers": [5, 1, 10, 3, 7], "stat_type": "` + tt.statType + `"}`
			result, err := tool.Handler(args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !strings.HasPrefix(result, tt.want) {
				t.Errorf("Expected %s ~%s, got %s", tt.statType, tt.want, result)
			}
		})
	}
}

func TestMathTool_Statistics_Sum(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "statistics", "numbers": [1, 2, 3, 4, 5], "stat_type": "sum"}`
	result, err := tool.Handler(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Sum of [1,2,3,4,5] = 15.0
	if !strings.HasPrefix(result, "15.0") {
		t.Errorf("Expected sum ~15.0, got %s", result)
	}
}

func TestMathTool_Statistics_StdDev(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "statistics", "numbers": [2, 4, 4, 4, 5, 5, 7, 9], "stat_type": "stdev"}`
	result, err := tool.Handler(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Result should be a positive number
	if result == "" || result == "0.000000" {
		t.Errorf("Expected non-zero stdev, got %s", result)
	}
}

func TestMathTool_Statistics_Variance(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "statistics", "numbers": [1, 2, 3, 4, 5], "stat_type": "variance"}`
	result, err := tool.Handler(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Variance should be positive
	if result == "" || result == "0.000000" {
		t.Errorf("Expected non-zero variance, got %s", result)
	}
}

func TestMathTool_Statistics_Errors(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name string
		args string
	}{
		{"empty numbers", `{"operation": "statistics", "numbers": [], "stat_type": "mean"}`},
		{"missing stat_type", `{"operation": "statistics", "numbers": [1, 2, 3]}`},
		{"invalid stat_type", `{"operation": "statistics", "numbers": [1, 2, 3], "stat_type": "invalid"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Handler(tt.args)
			if err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestMathTool_Solve_Linear(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name     string
		equation string
		want     string
	}{
		{"simple", `{"operation": "solve", "equation": "x+5=10"}`, "x = 5.0"},
		{"subtraction", `{"operation": "solve", "equation": "x-3=7"}`, "x = 10.0"},
		{"identity", `{"operation": "solve", "equation": "x=42"}`, "x = 42.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Handler(tt.equation)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.want) {
				t.Errorf("Expected %s, got %s", tt.want, result)
			}
		})
	}
}

func TestMathTool_Solve_Errors(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name     string
		equation string
	}{
		{"empty equation", `{"operation": "solve", "equation": ""}`},
		{"no equals sign", `{"operation": "solve", "equation": "x + 5"}`},
		{"quadratic (not implemented)", `{"operation": "solve", "equation": "2x^2+3x-5=0"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Handler(tt.equation)
			if err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestMathTool_Convert_Distance(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name     string
		args     string
		expected string
	}{
		{"km to m", `{"operation": "convert", "value": 1, "from_unit": "km", "to_unit": "m"}`, "1000.0"},
		{"m to cm", `{"operation": "convert", "value": 1, "from_unit": "m", "to_unit": "cm"}`, "100.0"},
		{"cm to mm", `{"operation": "convert", "value": 1, "from_unit": "cm", "to_unit": "mm"}`, "10.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Handler(tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMathTool_Convert_Weight(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "convert", "value": 1, "from_unit": "kg", "to_unit": "g"}`
	result, err := tool.Handler(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "1000.0") {
		t.Errorf("Expected 1000.0 g, got %s", result)
	}
}

func TestMathTool_Convert_Temperature(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name string
		args string
		want string
	}{
		{"celsius to fahrenheit", `{"operation": "convert", "value": 0, "from_unit": "celsius", "to_unit": "fahrenheit"}`, "32.0"},
		{"fahrenheit to celsius", `{"operation": "convert", "value": 32, "from_unit": "fahrenheit", "to_unit": "celsius"}`, "0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Handler(tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.want) {
				t.Errorf("Expected %s, got %s", tt.want, result)
			}
		})
	}
}

func TestMathTool_Convert_Time(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name string
		args string
		want string
	}{
		{"hours to minutes", `{"operation": "convert", "value": 1, "from_unit": "hours", "to_unit": "minutes"}`, "60.0"},
		{"minutes to seconds", `{"operation": "convert", "value": 1, "from_unit": "minutes", "to_unit": "seconds"}`, "60.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Handler(tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.want) {
				t.Errorf("Expected %s, got %s", tt.want, result)
			}
		})
	}
}

func TestMathTool_Convert_Errors(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name string
		args string
	}{
		{"missing units", `{"operation": "convert", "value": 1}`},
		{"incompatible units", `{"operation": "convert", "value": 1, "from_unit": "kg", "to_unit": "km"}`},
		{"unknown unit", `{"operation": "convert", "value": 1, "from_unit": "xyz", "to_unit": "abc"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Handler(tt.args)
			if err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestMathTool_Random_Integer(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "random", "random_type": "integer", "min": 1, "max": 10}`
	result, err := tool.Handler(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == "" {
		t.Errorf("Expected random integer but got empty result")
	}
}

func TestMathTool_Random_Float(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "random", "random_type": "float", "min": 0.0, "max": 1.0}`
	result, err := tool.Handler(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == "" {
		t.Errorf("Expected random float but got empty result")
	}
}

func TestMathTool_Random_Choice(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "random", "random_type": "choice", "choices": ["apple", "banana", "cherry"]}`
	result, err := tool.Handler(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	validChoices := []string{"apple", "banana", "cherry"}
	found := false
	for _, choice := range validChoices {
		if result == choice {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected one of %v, got %s", validChoices, result)
	}
}

func TestMathTool_Random_Errors(t *testing.T) {
	tool := NewMathTool()

	tests := []struct {
		name string
		args string
	}{
		{"missing random_type", `{"operation": "random"}`},
		{"invalid range", `{"operation": "random", "random_type": "integer", "min": 10, "max": 5}`},
		{"missing choices", `{"operation": "random", "random_type": "choice"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Handler(tt.args)
			if err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestMathTool_InvalidOperation(t *testing.T) {
	tool := NewMathTool()

	args := `{"operation": "invalid_op"}`
	_, err := tool.Handler(args)

	if err == nil {
		t.Errorf("Expected error for invalid operation but got none")
	}
}

func TestMathTool_InvalidJSON(t *testing.T) {
	tool := NewMathTool()

	args := `{invalid json}`
	_, err := tool.Handler(args)

	if err == nil {
		t.Errorf("Expected error for invalid JSON but got none")
	}
}

func TestMathTool_Metadata(t *testing.T) {
	tool := NewMathTool()

	if tool.Name != "math" {
		t.Errorf("Expected name 'math', got %s", tool.Name)
	}

	if tool.Description == "" {
		t.Errorf("Expected non-empty description")
	}

	if len(tool.Parameters) == 0 {
		t.Errorf("Expected non-empty parameters")
	}
}
