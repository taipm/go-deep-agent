package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Parser regular expressions for ReAct format
var (
	// thoughtRegex matches "THOUGHT: <reasoning>"
	// Supports both single-line and multi-line thoughts
	thoughtRegex = regexp.MustCompile(`(?i)^THOUGHT:\s*(.+)$`)

	// actionRegex matches "ACTION: tool(args)" or "ACTION: tool"
	// Captures: tool name and optional arguments
	actionRegex = regexp.MustCompile(`(?i)^ACTION:\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:\((.*)\))?$`)

	// finalRegex matches "FINAL: <answer>"
	// Supports multi-line answers
	finalRegex = regexp.MustCompile(`(?i)^FINAL:\s*(.+)$`)

	// observationRegex matches "OBSERVATION: <result>"
	// Usually added by system, not LLM, but included for completeness
	observationRegex = regexp.MustCompile(`(?i)^OBSERVATION:\s*(.+)$`)
)

// parseThought extracts a THOUGHT step from the LLM response.
// Supports formats:
//   - "THOUGHT: I need to search for information"
//   - "thought: Let me analyze this problem"
//
// Returns the reasoning text, or empty string if not a THOUGHT step.
//
// Example:
//
//	thought := parseThought("THOUGHT: I should search for weather")
//	// Returns: "I should search for weather"
func parseThought(text string) string {
	text = strings.TrimSpace(text)
	if matches := thoughtRegex.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// parseAction extracts an ACTION step from the LLM response.
// Supports formats:
//   - "ACTION: search(query='Paris weather')"
//   - "ACTION: calculator(expression='2+2')"
//   - "ACTION: get_weather(city=\"Paris\", units=\"metric\")"
//   - "ACTION: simple_tool"  (no arguments)
//
// Returns:
//   - tool: the tool name
//   - argsStr: the raw arguments string (unparsed)
//   - ok: true if this is a valid ACTION step
//
// Example:
//
//	tool, args, ok := parseAction("ACTION: search(query='Paris')")
//	// Returns: "search", "query='Paris'", true
func parseAction(text string) (tool string, argsStr string, ok bool) {
	text = strings.TrimSpace(text)
	matches := actionRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		return "", "", false
	}

	tool = matches[1]
	if len(matches) > 2 {
		argsStr = strings.TrimSpace(matches[2])
	}

	return tool, argsStr, true
}

// parseActionArgs parses the arguments string from an ACTION step.
// Supports multiple formats:
//   - JSON object: {query: "Paris", limit: 10}
//   - Key-value pairs: query="Paris", limit=10
//   - Single quoted: query='Paris', limit=10
//   - Python-style: query="Paris", units="metric"
//
// Returns a map of argument name to value.
// Returns error if parsing fails.
//
// Example:
//
//	args, err := parseActionArgs(`query="Paris", limit=10`)
//	// Returns: map[string]interface{}{"query": "Paris", "limit": 10}, nil
func parseActionArgs(argsStr string) (map[string]interface{}, error) {
	if argsStr == "" {
		return map[string]interface{}{}, nil
	}

	argsStr = strings.TrimSpace(argsStr)

	// Try parsing as JSON first
	if strings.HasPrefix(argsStr, "{") {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
			return args, nil
		}
	}

	// Parse key-value pairs: key="value", key='value', key=value
	args := make(map[string]interface{})
	// Updated regex to handle quoted strings with commas inside
	kvRegex := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^,]+))`)
	matches := kvRegex.FindAllStringSubmatch(argsStr, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("could not parse arguments: %s", argsStr)
	}

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		key := match[1]
		// Extract value from the appropriate capture group
		// Group 2: double-quoted, Group 3: single-quoted, Group 4: unquoted
		var value string
		if match[2] != "" {
			value = match[2] // Double-quoted
		} else if match[3] != "" {
			value = match[3] // Single-quoted
		} else if len(match) > 4 {
			value = strings.TrimSpace(match[4]) // Unquoted
		}

		// Try to parse as number
		var parsedValue interface{} = value
		if num, err := parseNumber(value); err == nil {
			parsedValue = num
		} else if value == "true" {
			parsedValue = true
		} else if value == "false" {
			parsedValue = false
		}

		args[key] = parsedValue
	}

	return args, nil
}

// parseFinal extracts the FINAL answer from the LLM response.
// Supports formats:
//   - "FINAL: The answer is 42"
//   - "final: Paris is the capital of France"
//
// Returns the final answer text, or empty string if not a FINAL step.
//
// Example:
//
//	answer := parseFinal("FINAL: The weather in Paris is sunny")
//	// Returns: "The weather in Paris is sunny"
func parseFinal(text string) string {
	text = strings.TrimSpace(text)
	if matches := finalRegex.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// parseObservation extracts an OBSERVATION step from text.
// This is typically added by the system, not the LLM, but included for completeness.
//
// Returns the observation text, or empty string if not an OBSERVATION step.
//
// Example:
//
//	obs := parseObservation("OBSERVATION: Temperature is 20°C")
//	// Returns: "Temperature is 20°C"
func parseObservation(text string) string {
	text = strings.TrimSpace(text)
	if matches := observationRegex.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// parseReActStep is the main parser that tries to identify the step type
// and extract the relevant information from the LLM response.
//
// It tries parsers in this order:
//  1. THOUGHT - reasoning step
//  2. ACTION - tool execution request
//  3. FINAL - final answer
//  4. OBSERVATION - tool result (system-generated)
//
// Returns:
//   - stepType: THOUGHT, ACTION, FINAL, OBSERVATION, or empty string
//   - content: the extracted content
//   - tool: tool name (only for ACTION)
//   - args: parsed arguments (only for ACTION)
//   - err: parsing error (only for ACTION with malformed args)
//
// Example:
//
//	stepType, content, tool, args, err := parseReActStep("THOUGHT: I need more info")
//	// Returns: "THOUGHT", "I need more info", "", nil, nil
//
//	stepType, content, tool, args, err := parseReActStep("ACTION: search(query='Paris')")
//	// Returns: "ACTION", "search(query='Paris')", "search", {"query": "Paris"}, nil
func parseReActStep(text string) (stepType string, content string, tool string, args map[string]interface{}, err error) {
	text = strings.TrimSpace(text)

	// Try THOUGHT
	if thought := parseThought(text); thought != "" {
		return StepTypeThought, thought, "", nil, nil
	}

	// Try ACTION
	if toolName, argsStr, ok := parseAction(text); ok {
		parsedArgs, parseErr := parseActionArgs(argsStr)
		if parseErr != nil {
			// Return error but still identify as ACTION
			return StepTypeAction, text, toolName, nil, fmt.Errorf("failed to parse action arguments: %w", parseErr)
		}
		return StepTypeAction, text, toolName, parsedArgs, nil
	}

	// Try FINAL
	if final := parseFinal(text); final != "" {
		return StepTypeFinal, final, "", nil, nil
	}

	// Try OBSERVATION
	if obs := parseObservation(text); obs != "" {
		return StepTypeObservation, obs, "", nil, nil
	}

	// Not a recognized step format
	return "", "", "", nil, fmt.Errorf("unrecognized step format: %s", text)
}

// parseNumber attempts to parse a string as a number (int or float).
// Returns the number as interface{} (either int or float64), or error if not a number.
func parseNumber(s string) (interface{}, error) {
	// Try float first (handles both int and float)
	var f float64
	if n, err := fmt.Sscanf(s, "%f", &f); err == nil && n == 1 {
		// Check if it's actually an integer
		if f == float64(int(f)) {
			return int(f), nil
		}
		return f, nil
	}

	return nil, fmt.Errorf("not a number: %s", s)
}
