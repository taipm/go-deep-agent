package agent

import (
	"testing"
)

// TestParseThought tests the parseThought function
func TestParseThought(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid THOUGHT uppercase",
			input:    "THOUGHT: I need to search for information",
			expected: "I need to search for information",
		},
		{
			name:     "Valid THOUGHT lowercase",
			input:    "thought: Let me analyze this problem",
			expected: "Let me analyze this problem",
		},
		{
			name:     "Valid THOUGHT mixed case",
			input:    "Thought: What should I do next?",
			expected: "What should I do next?",
		},
		{
			name:     "THOUGHT with extra whitespace",
			input:    "THOUGHT:    Multiple spaces here   ",
			expected: "Multiple spaces here",
		},
		{
			name:     "THOUGHT with leading whitespace",
			input:    "   THOUGHT: Trimmed input",
			expected: "Trimmed input",
		},
		{
			name:     "Not a THOUGHT - ACTION",
			input:    "ACTION: search(query='test')",
			expected: "",
		},
		{
			name:     "Not a THOUGHT - FINAL",
			input:    "FINAL: The answer is 42",
			expected: "",
		},
		{
			name:     "Not a THOUGHT - plain text",
			input:    "Just some random text",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "THOUGHT with colon in content",
			input:    "THOUGHT: Time is 10:30 AM",
			expected: "Time is 10:30 AM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseThought(tt.input)
			if result != tt.expected {
				t.Errorf("parseThought(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseAction tests the parseAction function
func TestParseAction(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTool    string
		wantArgsStr string
		wantOK      bool
	}{
		{
			name:        "Valid ACTION with args",
			input:       "ACTION: search(query='Paris')",
			wantTool:    "search",
			wantArgsStr: "query='Paris'",
			wantOK:      true,
		},
		{
			name:        "Valid ACTION uppercase",
			input:       "ACTION: calculator(expression='2+2')",
			wantTool:    "calculator",
			wantArgsStr: "expression='2+2'",
			wantOK:      true,
		},
		{
			name:        "Valid ACTION lowercase",
			input:       "action: get_weather(city=\"Paris\")",
			wantTool:    "get_weather",
			wantArgsStr: `city="Paris"`,
			wantOK:      true,
		},
		{
			name:        "Valid ACTION no args",
			input:       "ACTION: simple_tool",
			wantTool:    "simple_tool",
			wantArgsStr: "",
			wantOK:      true,
		},
		{
			name:        "Valid ACTION with spaces",
			input:       "ACTION:   search  (  query='test'  )",
			wantTool:    "search",
			wantArgsStr: "query='test'",
			wantOK:      true,
		},
		{
			name:        "Valid ACTION complex args",
			input:       `ACTION: api_call(url="https://api.com", method="POST", headers="{}")`,
			wantTool:    "api_call",
			wantArgsStr: `url="https://api.com", method="POST", headers="{}"`,
			wantOK:      true,
		},
		{
			name:        "Valid ACTION underscore in name",
			input:       "ACTION: fetch_data(id=123)",
			wantTool:    "fetch_data",
			wantArgsStr: "id=123",
			wantOK:      true,
		},
		{
			name:        "Valid ACTION numbers in name",
			input:       "ACTION: tool123(param=value)",
			wantTool:    "tool123",
			wantArgsStr: "param=value",
			wantOK:      true,
		},
		{
			name:        "Not an ACTION - THOUGHT",
			input:       "THOUGHT: I should search",
			wantTool:    "",
			wantArgsStr: "",
			wantOK:      false,
		},
		{
			name:        "Not an ACTION - FINAL",
			input:       "FINAL: The answer",
			wantTool:    "",
			wantArgsStr: "",
			wantOK:      false,
		},
		{
			name:        "Not an ACTION - plain text",
			input:       "Just text",
			wantTool:    "",
			wantArgsStr: "",
			wantOK:      false,
		},
		{
			name:        "Invalid ACTION - starts with number",
			input:       "ACTION: 123tool()",
			wantTool:    "",
			wantArgsStr: "",
			wantOK:      false,
		},
		{
			name:        "Invalid ACTION - special chars in name",
			input:       "ACTION: tool-name()",
			wantTool:    "",
			wantArgsStr: "",
			wantOK:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, argsStr, ok := parseAction(tt.input)
			if ok != tt.wantOK {
				t.Errorf("parseAction(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if tool != tt.wantTool {
				t.Errorf("parseAction(%q) tool = %q, want %q", tt.input, tool, tt.wantTool)
			}
			if argsStr != tt.wantArgsStr {
				t.Errorf("parseAction(%q) argsStr = %q, want %q", tt.input, argsStr, tt.wantArgsStr)
			}
		})
	}
}

// TestParseActionArgs tests the parseActionArgs function
func TestParseActionArgs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      map[string]interface{}
		wantError bool
	}{
		{
			name:  "Empty args",
			input: "",
			want:  map[string]interface{}{},
		},
		{
			name:  "Single quoted string",
			input: "query='Paris'",
			want:  map[string]interface{}{"query": "Paris"},
		},
		{
			name:  "Double quoted string",
			input: `query="Paris"`,
			want:  map[string]interface{}{"query": "Paris"},
		},
		{
			name:  "Multiple args",
			input: `query="Paris", limit=10`,
			want:  map[string]interface{}{"query": "Paris", "limit": 10},
		},
		{
			name:  "Integer value",
			input: "count=42",
			want:  map[string]interface{}{"count": 42},
		},
		{
			name:  "Float value",
			input: "temperature=20.5",
			want:  map[string]interface{}{"temperature": 20.5},
		},
		{
			name:  "Boolean true",
			input: "enabled=true",
			want:  map[string]interface{}{"enabled": true},
		},
		{
			name:  "Boolean false",
			input: "enabled=false",
			want:  map[string]interface{}{"enabled": false},
		},
		{
			name:  "Mixed types",
			input: `city="Paris", temp=20.5, count=3, active=true`,
			want: map[string]interface{}{
				"city":   "Paris",
				"temp":   20.5,
				"count":  3,
				"active": true,
			},
		},
		{
			name:  "Unquoted string value",
			input: "mode=fast",
			want:  map[string]interface{}{"mode": "fast"},
		},
		{
			name:  "Args with spaces",
			input: `query = "test" , limit = 5`,
			want:  map[string]interface{}{"query": "test", "limit": 5},
		},
		{
			name:  "JSON object",
			input: `{"query": "Paris", "limit": 10}`,
			want:  map[string]interface{}{"query": "Paris", "limit": float64(10)},
		},
		{
			name:  "JSON with nested object",
			input: `{"user": {"name": "John", "age": 30}}`,
			want: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  float64(30),
				},
			},
		},
		{
			name:  "String with special chars",
			input: `url="https://api.example.com/search?q=test"`,
			want:  map[string]interface{}{"url": "https://api.example.com/search?q=test"},
		},
		{
			name:      "Invalid format",
			input:     "not-valid-format",
			wantError: true,
		},
		{
			name:      "Malformed JSON",
			input:     `{query: "Paris"}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseActionArgs(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("parseActionArgs(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseActionArgs(%q) unexpected error: %v", tt.input, err)
				return
			}

			if !mapsEqual(got, tt.want) {
				t.Errorf("parseActionArgs(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseFinal tests the parseFinal function
func TestParseFinal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid FINAL uppercase",
			input:    "FINAL: The answer is 42",
			expected: "The answer is 42",
		},
		{
			name:     "Valid FINAL lowercase",
			input:    "final: Paris is the capital",
			expected: "Paris is the capital",
		},
		{
			name:     "Valid FINAL mixed case",
			input:    "Final: Done processing",
			expected: "Done processing",
		},
		{
			name:     "FINAL with extra whitespace",
			input:    "FINAL:    Answer here   ",
			expected: "Answer here",
		},
		{
			name:     "Not a FINAL - THOUGHT",
			input:    "THOUGHT: Thinking",
			expected: "",
		},
		{
			name:     "Not a FINAL - ACTION",
			input:    "ACTION: search()",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFinal(tt.input)
			if result != tt.expected {
				t.Errorf("parseFinal(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseObservation tests the parseObservation function
func TestParseObservation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid OBSERVATION",
			input:    "OBSERVATION: Temperature is 20°C",
			expected: "Temperature is 20°C",
		},
		{
			name:     "Valid observation lowercase",
			input:    "observation: Result here",
			expected: "Result here",
		},
		{
			name:     "Not an OBSERVATION",
			input:    "THOUGHT: Thinking",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseObservation(tt.input)
			if result != tt.expected {
				t.Errorf("parseObservation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseReActStep tests the main parseReActStep function
func TestParseReActStep(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantType     string
		wantContent  string
		wantTool     string
		wantArgs     map[string]interface{}
		wantError    bool
		errorContain string
	}{
		{
			name:        "Parse THOUGHT",
			input:       "THOUGHT: I need to search for Paris",
			wantType:    StepTypeThought,
			wantContent: "I need to search for Paris",
			wantTool:    "",
			wantArgs:    nil,
		},
		{
			name:        "Parse ACTION with args",
			input:       `ACTION: search(query="Paris")`,
			wantType:    StepTypeAction,
			wantContent: `ACTION: search(query="Paris")`,
			wantTool:    "search",
			wantArgs:    map[string]interface{}{"query": "Paris"},
		},
		{
			name:        "Parse ACTION no args",
			input:       "ACTION: get_time",
			wantType:    StepTypeAction,
			wantContent: "ACTION: get_time",
			wantTool:    "get_time",
			wantArgs:    map[string]interface{}{},
		},
		{
			name:        "Parse FINAL",
			input:       "FINAL: The capital is Paris",
			wantType:    StepTypeFinal,
			wantContent: "The capital is Paris",
			wantTool:    "",
			wantArgs:    nil,
		},
		{
			name:        "Parse OBSERVATION",
			input:       "OBSERVATION: Temperature: 20°C",
			wantType:    StepTypeObservation,
			wantContent: "Temperature: 20°C",
			wantTool:    "",
			wantArgs:    nil,
		},
		{
			name:         "Parse ACTION with invalid args",
			input:        "ACTION: search(malformed-args)",
			wantType:     StepTypeAction,
			wantContent:  "ACTION: search(malformed-args)",
			wantTool:     "search",
			wantArgs:     nil,
			wantError:    true,
			errorContain: "failed to parse action arguments",
		},
		{
			name:         "Unrecognized format",
			input:        "Random text without format",
			wantType:     "",
			wantContent:  "",
			wantTool:     "",
			wantArgs:     nil,
			wantError:    true,
			errorContain: "unrecognized step format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stepType, content, tool, args, err := parseReActStep(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("parseReActStep(%q) expected error, got nil", tt.input)
					return
				}
				if tt.errorContain != "" && !containsString(err.Error(), tt.errorContain) {
					t.Errorf("parseReActStep(%q) error = %q, want to contain %q", tt.input, err.Error(), tt.errorContain)
				}
			} else if err != nil {
				t.Errorf("parseReActStep(%q) unexpected error: %v", tt.input, err)
				return
			}

			if stepType != tt.wantType {
				t.Errorf("parseReActStep(%q) stepType = %q, want %q", tt.input, stepType, tt.wantType)
			}
			if content != tt.wantContent {
				t.Errorf("parseReActStep(%q) content = %q, want %q", tt.input, content, tt.wantContent)
			}
			if tool != tt.wantTool {
				t.Errorf("parseReActStep(%q) tool = %q, want %q", tt.input, tool, tt.wantTool)
			}
			if !mapsEqual(args, tt.wantArgs) {
				t.Errorf("parseReActStep(%q) args = %v, want %v", tt.input, args, tt.wantArgs)
			}
		})
	}
}

// TestParseReActStep_RealLLMOutputs tests with realistic GPT-4 style outputs
func TestParseReActStep_RealLLMOutputs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantTool string
	}{
		{
			name:     "GPT-4 style THOUGHT",
			input:    "THOUGHT: To find the current weather in Paris, I'll use the weather search tool.",
			wantType: StepTypeThought,
		},
		{
			name:     "GPT-4 style ACTION",
			input:    `ACTION: weather_search(city="Paris", units="metric")`,
			wantType: StepTypeAction,
			wantTool: "weather_search",
		},
		{
			name:     "GPT-4 style FINAL with explanation",
			input:    "FINAL: Based on the search results, the weather in Paris is currently 18°C with partly cloudy skies.",
			wantType: StepTypeFinal,
		},
		{
			name:     "Lowercase thought",
			input:    "thought: I should calculate the total first",
			wantType: StepTypeThought,
		},
		{
			name:     "Action with complex JSON",
			input:    `ACTION: api_call(endpoint="https://api.com/data", method="POST", body={"key": "value"})`,
			wantType: StepTypeAction,
			wantTool: "api_call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stepType, _, tool, _, err := parseReActStep(tt.input)

			if err != nil && tt.wantType != "" {
				t.Errorf("parseReActStep(%q) unexpected error: %v", tt.input, err)
				return
			}

			if stepType != tt.wantType {
				t.Errorf("parseReActStep(%q) stepType = %q, want %q", tt.input, stepType, tt.wantType)
			}

			if tt.wantTool != "" && tool != tt.wantTool {
				t.Errorf("parseReActStep(%q) tool = %q, want %q", tt.input, tool, tt.wantTool)
			}
		})
	}
}

// TestParseActionArgs_EdgeCases tests edge cases in argument parsing
func TestParseActionArgs_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]interface{}
	}{
		{
			name:  "URL with query params",
			input: `url="https://api.com?q=test&limit=10"`,
			want:  map[string]interface{}{"url": "https://api.com?q=test&limit=10"},
		},
		{
			name:  "String with commas",
			input: `text="Hello, world, how are you?"`,
			want:  map[string]interface{}{"text": "Hello, world, how are you?"},
		},
		{
			name:  "Negative number",
			input: "offset=-5",
			want:  map[string]interface{}{"offset": -5},
		},
		{
			name:  "Zero value",
			input: "count=0",
			want:  map[string]interface{}{"count": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseActionArgs(tt.input)
			if err != nil {
				t.Errorf("parseActionArgs(%q) unexpected error: %v", tt.input, err)
				return
			}

			if !mapsEqual(got, tt.want) {
				t.Errorf("parseActionArgs(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// Helper function to compare maps
func mapsEqual(a, b map[string]interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}

		// Handle nested maps
		if aMap, aOk := v.(map[string]interface{}); aOk {
			if bMap, bOk := bv.(map[string]interface{}); bOk {
				if !mapsEqual(aMap, bMap) {
					return false
				}
				continue
			}
			return false
		}

		// Direct comparison
		if v != bv {
			return false
		}
	}

	return true
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
