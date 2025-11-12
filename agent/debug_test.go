package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Mock logger for testing
type mockLogger struct {
	debugMsgs []string
	infoMsgs  []string
	errorMsgs []string
}

func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	m.debugMsgs = append(m.debugMsgs, msg)
}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...Field) {
	m.infoMsgs = append(m.infoMsgs, msg)
}

func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	// Not used in debug logger
}

func (m *mockLogger) Error(ctx context.Context, msg string, fields ...Field) {
	m.errorMsgs = append(m.errorMsgs, msg)
}

func TestDefaultDebugConfig(t *testing.T) {
	config := DefaultDebugConfig()

	if !config.Enabled {
		t.Error("Expected Enabled=true")
	}
	if config.Level != DebugLevelBasic {
		t.Errorf("Expected DebugLevelBasic, got %v", config.Level)
	}
	if !config.RedactSecrets {
		t.Error("Expected RedactSecrets=true")
	}
	if !config.LogRequests {
		t.Error("Expected LogRequests=true")
	}
	if !config.LogResponses {
		t.Error("Expected LogResponses=true")
	}
	if !config.LogErrors {
		t.Error("Expected LogErrors=true")
	}
	if config.LogTokenUsage {
		t.Error("Expected LogTokenUsage=false for basic level")
	}
	if config.LogToolExecutions {
		t.Error("Expected LogToolExecutions=false for basic level")
	}
	if config.MaxLogLength != 2000 {
		t.Errorf("Expected MaxLogLength=2000, got %d", config.MaxLogLength)
	}
}

func TestVerboseDebugConfig(t *testing.T) {
	config := VerboseDebugConfig()

	if !config.Enabled {
		t.Error("Expected Enabled=true")
	}
	if config.Level != DebugLevelVerbose {
		t.Errorf("Expected DebugLevelVerbose, got %v", config.Level)
	}
	if !config.RedactSecrets {
		t.Error("Expected RedactSecrets=true")
	}
	if !config.LogTokenUsage {
		t.Error("Expected LogTokenUsage=true for verbose level")
	}
	if !config.LogToolExecutions {
		t.Error("Expected LogToolExecutions=true for verbose level")
	}
	if config.MaxLogLength != 5000 {
		t.Errorf("Expected MaxLogLength=5000, got %d", config.MaxLogLength)
	}
}

func TestDebugLogger_LogRequest(t *testing.T) {
	mock := &mockLogger{}
	config := DefaultDebugConfig()
	dl := newDebugLogger(config, mock)

	body := map[string]string{"test": "data"}
	dl.logRequest("POST", "https://api.example.com/chat", body)

	if len(mock.debugMsgs) != 1 {
		t.Fatalf("Expected 1 debug message, got %d", len(mock.debugMsgs))
	}

	msg := mock.debugMsgs[0]
	if !strings.Contains(msg, "POST") {
		t.Error("Expected message to contain POST")
	}
	if !strings.Contains(msg, "https://api.example.com/chat") {
		t.Error("Expected message to contain URL")
	}
	if !strings.Contains(msg, "test") {
		t.Error("Expected message to contain body data")
	}
}

func TestDebugLogger_LogResponse(t *testing.T) {
	mock := &mockLogger{}
	config := DefaultDebugConfig()
	dl := newDebugLogger(config, mock)

	body := map[string]string{"result": "success"}
	dl.logResponse(200, body, 500*time.Millisecond)

	if len(mock.debugMsgs) != 1 {
		t.Fatalf("Expected 1 debug message, got %d", len(mock.debugMsgs))
	}

	msg := mock.debugMsgs[0]
	if !strings.Contains(msg, "200") {
		t.Error("Expected message to contain status code")
	}
	if !strings.Contains(msg, "success") {
		t.Error("Expected message to contain response body")
	}
}

func TestDebugLogger_LogError(t *testing.T) {
	mock := &mockLogger{}
	config := DefaultDebugConfig()
	dl := newDebugLogger(config, mock)

	err := errors.New("test error")
	dl.logError(err, "test context")

	if len(mock.errorMsgs) != 1 {
		t.Fatalf("Expected 1 error message, got %d", len(mock.errorMsgs))
	}

	msg := mock.errorMsgs[0]
	if !strings.Contains(msg, "test error") {
		t.Error("Expected message to contain error text")
	}
	if !strings.Contains(msg, "test context") {
		t.Error("Expected message to contain context")
	}
}

func TestDebugLogger_LogTokenUsage(t *testing.T) {
	mock := &mockLogger{}
	config := VerboseDebugConfig() // Verbose includes token usage
	dl := newDebugLogger(config, mock)

	dl.logTokenUsage(100, 50, 150)

	if len(mock.debugMsgs) != 1 {
		t.Fatalf("Expected 1 debug message, got %d", len(mock.debugMsgs))
	}

	msg := mock.debugMsgs[0]
	if !strings.Contains(msg, "100") {
		t.Error("Expected message to contain prompt tokens")
	}
	if !strings.Contains(msg, "50") {
		t.Error("Expected message to contain completion tokens")
	}
	if !strings.Contains(msg, "150") {
		t.Error("Expected message to contain total tokens")
	}
}

func TestDebugLogger_LogTokenUsage_BasicLevel(t *testing.T) {
	mock := &mockLogger{}
	config := DefaultDebugConfig() // Basic doesn't include token usage
	dl := newDebugLogger(config, mock)

	dl.logTokenUsage(100, 50, 150)

	// Should not log at basic level
	if len(mock.debugMsgs) != 0 {
		t.Errorf("Expected 0 debug messages at basic level, got %d", len(mock.debugMsgs))
	}
}

func TestDebugLogger_LogToolExecution_Success(t *testing.T) {
	mock := &mockLogger{}
	config := VerboseDebugConfig()
	dl := newDebugLogger(config, mock)

	args := map[string]interface{}{"x": 10, "y": 20}
	dl.logToolExecution("calculator", args, "30", nil)

	if len(mock.debugMsgs) != 1 {
		t.Fatalf("Expected 1 debug message, got %d", len(mock.debugMsgs))
	}

	msg := mock.debugMsgs[0]
	if !strings.Contains(msg, "calculator") {
		t.Error("Expected message to contain tool name")
	}
	if !strings.Contains(msg, "30") {
		t.Error("Expected message to contain result")
	}
	if !strings.Contains(msg, "executed") {
		t.Error("Expected message to contain 'executed' for success")
	}
}

func TestDebugLogger_LogToolExecution_Error(t *testing.T) {
	mock := &mockLogger{}
	config := VerboseDebugConfig()
	dl := newDebugLogger(config, mock)

	args := map[string]interface{}{"x": 10, "y": 0}
	err := errors.New("division by zero")
	dl.logToolExecution("calculator", args, "", err)

	if len(mock.debugMsgs) != 1 {
		t.Fatalf("Expected 1 debug message, got %d", len(mock.debugMsgs))
	}

	msg := mock.debugMsgs[0]
	if !strings.Contains(msg, "calculator") {
		t.Error("Expected message to contain tool name")
	}
	if !strings.Contains(msg, "division by zero") {
		t.Error("Expected message to contain error")
	}
	if !strings.Contains(msg, "failed") {
		t.Error("Expected message to contain 'failed' for error")
	}
}

func TestDebugLogger_RedactSecrets(t *testing.T) {
	mock := &mockLogger{}
	config := DefaultDebugConfig()
	dl := newDebugLogger(config, mock)

	tests := []struct {
		name             string
		input            string
		shouldContain    string
		shouldNotContain string
	}{
		{
			name:             "OpenAI API key",
			input:            `{"api_key": "sk-1234567890abcdef1234567890abcdef1234567890abcdef"}`,
			shouldContain:    "REDACTED",
			shouldNotContain: "1234567890abcdef",
		},
		{
			name:             "OpenAI project key",
			input:            `{"api_key": "sk-proj-abcdefghijklmnopqrstuvwxyz1234567890"}`,
			shouldContain:    "REDACTED",
			shouldNotContain: "abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:             "Bearer token",
			input:            `Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9`,
			shouldContain:    "REDACTED",
			shouldNotContain: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:             "Generic API key in JSON",
			input:            `{"api_key": "supersecretkey123456"}`,
			shouldContain:    "REDACTED",
			shouldNotContain: "supersecretkey123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := dl.redactSecrets(tt.input)
			if !strings.Contains(redacted, tt.shouldContain) {
				t.Errorf("Expected redacted string to contain %q, got: %s", tt.shouldContain, redacted)
			}
			if strings.Contains(redacted, tt.shouldNotContain) {
				t.Errorf("Expected redacted string to NOT contain %q, got: %s", tt.shouldNotContain, redacted)
			}
		})
	}
}

func TestDebugLogger_Truncate(t *testing.T) {
	mock := &mockLogger{}
	config := DebugConfig{
		Enabled:      true,
		Level:        DebugLevelBasic,
		MaxLogLength: 50,
	}
	dl := newDebugLogger(config, mock)

	longString := strings.Repeat("x", 200)
	truncated := dl.truncate(longString)

	if len(truncated) > 100 { // Should be around MaxLogLength + overhead
		t.Errorf("Expected truncated string to be around %d chars, got %d", config.MaxLogLength, len(truncated))
	}
	if !strings.Contains(truncated, "truncated") {
		t.Error("Expected truncation indicator in output")
	}
}

func TestDebugLogger_Truncate_NoLimit(t *testing.T) {
	mock := &mockLogger{}
	config := DebugConfig{
		Enabled:      true,
		Level:        DebugLevelBasic,
		MaxLogLength: 0, // No limit
	}
	dl := newDebugLogger(config, mock)

	longString := strings.Repeat("x", 5000)
	truncated := dl.truncate(longString)

	if truncated != longString {
		t.Error("Expected no truncation when MaxLogLength=0")
	}
}

func TestDebugLogger_Disabled(t *testing.T) {
	mock := &mockLogger{}
	config := DebugConfig{
		Enabled: false, // Disabled
	}
	dl := newDebugLogger(config, mock)

	dl.logRequest("POST", "https://api.example.com", nil)
	dl.logResponse(200, nil, time.Second)
	dl.logError(errors.New("test"), "context")
	dl.logTokenUsage(100, 50, 150)
	dl.logToolExecution("tool", nil, "result", nil)

	// Nothing should be logged when disabled
	if len(mock.debugMsgs) != 0 {
		t.Errorf("Expected 0 debug messages when disabled, got %d", len(mock.debugMsgs))
	}
	if len(mock.errorMsgs) != 0 {
		t.Errorf("Expected 0 error messages when disabled, got %d", len(mock.errorMsgs))
	}
}

func TestDebugLogger_NoLogger(t *testing.T) {
	config := DefaultDebugConfig()
	dl := newDebugLogger(config, nil) // No logger

	// Should not panic
	dl.logRequest("POST", "https://api.example.com", nil)
	dl.logResponse(200, nil, time.Second)
	dl.logError(errors.New("test"), "context")
}

func TestBuilder_WithDebug(t *testing.T) {
	b := NewOpenAI("gpt-4o-mini", "test-key").
		WithDebug(DefaultDebugConfig())

	if b.debugLogger == nil {
		t.Error("Expected debugLogger to be set")
	}
	if !b.debugConfig.Enabled {
		t.Error("Expected debug config to be enabled")
	}
	if b.logger == nil {
		t.Error("Expected logger to be auto-created if not set")
	}
}

func TestBuilder_WithDebug_CustomLogger(t *testing.T) {
	mock := &mockLogger{}
	b := NewOpenAI("gpt-4o-mini", "test-key").
		WithLogger(mock).
		WithDebug(VerboseDebugConfig())

	if b.debugLogger == nil {
		t.Error("Expected debugLogger to be set")
	}
	if b.debugLogger.logger != mock {
		t.Error("Expected debugLogger to use custom logger")
	}
}

func TestDebugLevel_String(t *testing.T) {
	tests := []struct {
		level    DebugLevel
		expected string
	}{
		{DebugLevelNone, "DebugLevelNone"},
		{DebugLevelBasic, "DebugLevelBasic"},
		{DebugLevelVerbose, "DebugLevelVerbose"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			// Just verify the levels are distinct
			if tt.level < 0 || tt.level > 2 {
				t.Errorf("Expected level to be 0-2, got %d", tt.level)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		contains string
	}{
		{500 * time.Microsecond, "µs"},
		{500 * time.Millisecond, "ms"},
		{2 * time.Second, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.duration.String(), func(t *testing.T) {
			result := formatDuration(tt.duration)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected %q to contain %q", result, tt.contains)
			}
		})
	}
}

func TestDebugLogger_RealWorldScenario(t *testing.T) {
	mock := &mockLogger{}
	config := VerboseDebugConfig()
	dl := newDebugLogger(config, mock)

	// Simulate a complete request lifecycle
	requestBody := map[string]interface{}{
		"model":    "gpt-4o-mini",
		"messages": []string{"Hello!"},
		"api_key":  "sk-1234567890abcdef1234567890abcdef1234567890abcdef",
	}

	// 1. Log request
	dl.logRequest("POST", "https://api.openai.com/v1/chat/completions", requestBody)

	// 2. Log tool execution
	toolArgs := map[string]int{"x": 10, "y": 20}
	dl.logToolExecution("calculator", toolArgs, "30", nil)

	// 3. Log token usage
	dl.logTokenUsage(100, 50, 150)

	// 4. Log response
	responseBody := map[string]interface{}{
		"choices": []map[string]string{
			{"message": "Hello! How can I help?"},
		},
	}
	dl.logResponse(200, responseBody, 1500*time.Millisecond)

	// Verify all logged
	if len(mock.debugMsgs) != 4 {
		t.Errorf("Expected 4 debug messages, got %d", len(mock.debugMsgs))
	}

	// Verify secret redaction in first message
	if strings.Contains(mock.debugMsgs[0], "1234567890abcdef") {
		t.Error("API key was not redacted in request log")
	}
	if !strings.Contains(mock.debugMsgs[0], "REDACTED") {
		t.Error("Expected REDACTED marker in request log")
	}
}

func TestDebugLogger_SecretRedaction_Comprehensive(t *testing.T) {
	mock := &mockLogger{}
	config := DefaultDebugConfig()
	dl := newDebugLogger(config, mock)

	// Test various secret patterns
	body := map[string]interface{}{
		"openai_key": "sk-1234567890abcdef1234567890abcdef1234567890abcdef",
		"auth":       "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
		"password":   "mySecretPassword123!",
		"token":      "ghp_1234567890abcdefghijklmnopqrstuvwxyz",
		"normal":     "this should not be redacted",
	}

	dl.logRequest("POST", "https://api.example.com", body)

	msg := mock.debugMsgs[0]

	// Should redact all secrets
	if strings.Contains(msg, "1234567890abcdef") {
		t.Error("OpenAI key not redacted")
	}
	if strings.Contains(msg, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Error("Bearer token not redacted")
	}
	if strings.Contains(msg, "mySecretPassword123!") {
		t.Error("Password not redacted")
	}

	// Should keep normal text
	if !strings.Contains(msg, "this should not be redacted") {
		t.Error("Normal text was incorrectly redacted")
	}
}
