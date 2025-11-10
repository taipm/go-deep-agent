package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DebugLevel defines the level of debug logging.
type DebugLevel int

const (
	// DebugLevelNone disables all debug logging.
	DebugLevelNone DebugLevel = iota
	// DebugLevelBasic logs requests, responses, and errors.
	DebugLevelBasic
	// DebugLevelVerbose logs everything including token usage and tool executions.
	DebugLevelVerbose
)

// DebugConfig configures debug logging behavior.
type DebugConfig struct {
	// Enabled turns debug logging on/off.
	Enabled bool

	// Level controls what gets logged.
	Level DebugLevel

	// RedactSecrets masks API keys and sensitive data in logs.
	RedactSecrets bool

	// LogRequests logs outgoing API requests.
	LogRequests bool

	// LogResponses logs incoming API responses.
	LogResponses bool

	// LogErrors logs errors with full context.
	LogErrors bool

	// LogTokenUsage logs token consumption per request.
	LogTokenUsage bool

	// LogToolExecutions logs tool calls and results.
	LogToolExecutions bool

	// MaxLogLength limits the size of logged messages (0 = unlimited).
	MaxLogLength int
}

// DefaultDebugConfig returns a sensible default debug configuration.
func DefaultDebugConfig() DebugConfig {
	return DebugConfig{
		Enabled:           true,
		Level:             DebugLevelBasic,
		RedactSecrets:     true,
		LogRequests:       true,
		LogResponses:      true,
		LogErrors:         true,
		LogTokenUsage:     false,
		LogToolExecutions: false,
		MaxLogLength:      2000,
	}
}

// VerboseDebugConfig returns a configuration that logs everything.
func VerboseDebugConfig() DebugConfig {
	return DebugConfig{
		Enabled:           true,
		Level:             DebugLevelVerbose,
		RedactSecrets:     true,
		LogRequests:       true,
		LogResponses:      true,
		LogErrors:         true,
		LogTokenUsage:     true,
		LogToolExecutions: true,
		MaxLogLength:      5000,
	}
}

// debugLogger handles debug output for the agent.
type debugLogger struct {
	config DebugConfig
	logger Logger
}

// newDebugLogger creates a new debug logger.
func newDebugLogger(config DebugConfig, logger Logger) *debugLogger {
	return &debugLogger{
		config: config,
		logger: logger,
	}
}

// isEnabled checks if debug logging is enabled.
func (d *debugLogger) isEnabled() bool {
	return d.config.Enabled && d.logger != nil
}

// shouldLog checks if a specific log type should be logged based on level.
func (d *debugLogger) shouldLog(logType string) bool {
	if !d.isEnabled() {
		return false
	}

	switch logType {
	case "request", "response", "error":
		return d.config.Level >= DebugLevelBasic
	case "token", "tool":
		return d.config.Level >= DebugLevelVerbose
	default:
		return false
	}
}

// logRequest logs an outgoing API request.
func (d *debugLogger) logRequest(method, url string, body interface{}) {
	if !d.isEnabled() || !d.config.LogRequests || !d.shouldLog("request") {
		return
	}

	var bodyStr string
	if body != nil {
		if jsonBytes, err := json.Marshal(body); err == nil {
			bodyStr = string(jsonBytes)
			if d.config.RedactSecrets {
				bodyStr = d.redactSecrets(bodyStr)
			}
			bodyStr = d.truncate(bodyStr)
		}
	}

	d.logger.Debug(context.Background(), fmt.Sprintf("[DEBUG] %s %s\nBody: %s", method, url, bodyStr))
}

// logResponse logs an incoming API response.
func (d *debugLogger) logResponse(status int, body interface{}, duration time.Duration) {
	if !d.isEnabled() || !d.config.LogResponses || !d.shouldLog("response") {
		return
	}

	var bodyStr string
	if body != nil {
		if jsonBytes, err := json.Marshal(body); err == nil {
			bodyStr = string(jsonBytes)
			if d.config.RedactSecrets {
				bodyStr = d.redactSecrets(bodyStr)
			}
			bodyStr = d.truncate(bodyStr)
		}
	}

	d.logger.Debug(context.Background(), fmt.Sprintf("[DEBUG] Response (status: %d, duration: %s)\nBody: %s",
		status, duration, bodyStr))
}

// logError logs an error with full context.
func (d *debugLogger) logError(err error, contextStr string) {
	if !d.isEnabled() || !d.config.LogErrors || !d.shouldLog("error") {
		return
	}

	errMsg := err.Error()
	if d.config.RedactSecrets {
		errMsg = d.redactSecrets(errMsg)
	}

	var contextInfo string
	if contextStr != "" {
		contextInfo = fmt.Sprintf(" [%s]", contextStr)
	}

	d.logger.Error(context.Background(), fmt.Sprintf("[DEBUG] Error%s: %s", contextInfo, errMsg))
}

// logTokenUsage logs token consumption.
func (d *debugLogger) logTokenUsage(promptTokens, completionTokens, totalTokens int) {
	if !d.isEnabled() || !d.config.LogTokenUsage || !d.shouldLog("token") {
		return
	}

	d.logger.Debug(context.Background(), fmt.Sprintf("[DEBUG] Token Usage - Prompt: %d, Completion: %d, Total: %d",
		promptTokens, completionTokens, totalTokens))
}

// logToolExecution logs a tool call and its result.
func (d *debugLogger) logToolExecution(toolName string, args interface{}, result string, err error) {
	if !d.isEnabled() || !d.config.LogToolExecutions || !d.shouldLog("tool") {
		return
	}

	var argsStr string
	if args != nil {
		if jsonBytes, err := json.Marshal(args); err == nil {
			argsStr = string(jsonBytes)
			if d.config.RedactSecrets {
				argsStr = d.redactSecrets(argsStr)
			}
			argsStr = d.truncate(argsStr)
		}
	}

	if err != nil {
		errMsg := err.Error()
		if d.config.RedactSecrets {
			errMsg = d.redactSecrets(errMsg)
		}
		d.logger.Debug(context.Background(), fmt.Sprintf("[DEBUG] Tool '%s' failed\nArgs: %s\nError: %s",
			toolName, argsStr, errMsg))
	} else {
		resultStr := d.truncate(result)
		if d.config.RedactSecrets {
			resultStr = d.redactSecrets(resultStr)
		}
		d.logger.Debug(context.Background(), fmt.Sprintf("[DEBUG] Tool '%s' executed\nArgs: %s\nResult: %s",
			toolName, argsStr, resultStr))
	}
}

// redactSecrets masks sensitive information in strings.
func (d *debugLogger) redactSecrets(s string) string {
	// Redact API keys (common patterns)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`sk-[a-zA-Z0-9]{48}`),                      // OpenAI keys
		regexp.MustCompile(`sk-proj-[a-zA-Z0-9_-]{43,}`),              // OpenAI project keys
		regexp.MustCompile(`Bearer\s+[a-zA-Z0-9_\-\.]{20,}`),          // Bearer tokens
		regexp.MustCompile(`api[_-]?key["']?\s*[:=]\s*["']?[^\s"']+`), // Generic API keys in JSON/params
		regexp.MustCompile(`token["']?\s*[:=]\s*["']?[^\s"']+`),       // Generic tokens
		regexp.MustCompile(`password["']?\s*[:=]\s*["']?[^\s"']+`),    // Passwords
	}

	result := s
	for _, pattern := range patterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// Keep first 10 chars visible, redact the rest
			if len(match) > 15 {
				return match[:10] + "***REDACTED***"
			}
			return "***REDACTED***"
		})
	}

	return result
}

// truncate limits string length for logging.
func (d *debugLogger) truncate(s string) string {
	if d.config.MaxLogLength <= 0 || len(s) <= d.config.MaxLogLength {
		return s
	}

	// Show beginning and end
	half := d.config.MaxLogLength / 2
	return s[:half] + fmt.Sprintf("\n... [%d chars truncated] ...\n", len(s)-d.config.MaxLogLength) + s[len(s)-half:]
}

// logInfo logs general information messages.
func (d *debugLogger) logInfo(format string, args ...interface{}) {
	if !d.isEnabled() {
		return
	}
	d.logger.Info(context.Background(), fmt.Sprintf("[DEBUG] "+format, args...))
}

// Helper function to format duration nicely.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Milliseconds()))
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// Helper to sanitize message content for logging.
func sanitizeForLogging(content string, redactSecrets bool, maxLength int) string {
	result := content

	// Redact secrets if enabled
	if redactSecrets {
		// Simple redaction - can be enhanced
		result = strings.ReplaceAll(result, "\n", " ")
		if strings.Contains(strings.ToLower(result), "api") {
			// Be conservative and redact potential API keys
			words := strings.Fields(result)
			for i, word := range words {
				if len(word) > 20 && !strings.Contains(word, " ") {
					words[i] = word[:5] + "***"
				}
			}
			result = strings.Join(words, " ")
		}
	}

	// Truncate if needed
	if maxLength > 0 && len(result) > maxLength {
		return result[:maxLength] + "..."
	}

	return result
}
