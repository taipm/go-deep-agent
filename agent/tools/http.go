package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

// NewHTTPRequestTool creates a tool for making HTTP requests.
// Includes timeout protection, header management, and response parsing.
//
// Available methods: GET, POST, PUT, DELETE
// Supports: custom headers, query parameters, request body, JSON parsing
//
// Example:
//
//	httpTool := tools.NewHTTPRequestTool()
//	agent.NewOpenAI("gpt-4o", apiKey).
//	    WithTool(httpTool).
//	    WithAutoExecute().
//	    Ask(ctx, "Get the weather from https://api.weather.com/forecast")
func NewHTTPRequestTool() *agent.Tool {
	return agent.NewTool("http_request", "Make HTTP requests (GET, POST, PUT, DELETE) to APIs and web services").
		AddParameter("method", "string", "HTTP method: GET, POST, PUT, DELETE", true).
		AddParameter("url", "string", "Full URL to request", true).
		AddParameter("headers", "string", "Optional headers as JSON object (e.g., {\"Authorization\": \"Bearer token\"})", false).
		AddParameter("body", "string", "Optional request body (for POST, PUT)", false).
		AddParameter("timeout_seconds", "number", "Optional timeout in seconds (default: 30)", false).
		WithHandler(httpRequestHandler)
}

// httpRequestHandler executes HTTP requests
func httpRequestHandler(args string) (string, error) {
	var params struct {
		Method         string  `json:"method"`
		URL            string  `json:"url"`
		Headers        string  `json:"headers"`
		Body           string  `json:"body"`
		TimeoutSeconds float64 `json:"timeout_seconds"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Validate method
	method := strings.ToUpper(params.Method)
	if !isValidHTTPMethod(method) {
		return "", fmt.Errorf("invalid HTTP method: %s", params.Method)
	}

	// Validate URL
	if params.URL == "" {
		return "", fmt.Errorf("URL is required")
	}
	if !strings.HasPrefix(params.URL, "http://") && !strings.HasPrefix(params.URL, "https://") {
		return "", fmt.Errorf("URL must start with http:// or https://")
	}

	// Set timeout (default 30 seconds)
	timeout := 30 * time.Second
	if params.TimeoutSeconds > 0 {
		timeout = time.Duration(params.TimeoutSeconds) * time.Second
	}

	// Make the request
	return makeHTTPRequest(method, params.URL, params.Headers, params.Body, timeout)
}

// isValidHTTPMethod checks if the HTTP method is supported
func isValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE"}
	for _, m := range validMethods {
		if method == m {
			return true
		}
	}
	return false
}

// makeHTTPRequest performs the actual HTTP request
func makeHTTPRequest(method, url, headersJSON, body string, timeout time.Duration) (string, error) {
	ctx := getContext()

	logInfo(ctx, "Making HTTP request", map[string]interface{}{
		"tool":         "http_request",
		"method":       method,
		"url":          url,
		"timeout_secs": timeout.Seconds(),
		"has_body":     body != "",
	})

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request body
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}

	// Create request
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		logError(ctx, "Failed to create HTTP request", map[string]interface{}{
			"tool":   "http_request",
			"method": method,
			"url":    url,
			"error":  err.Error(),
		})
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", "go-deep-agent/0.5.3")

	// Parse and set custom headers
	if headersJSON != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
			logError(ctx, "Invalid headers JSON", map[string]interface{}{
				"tool":         "http_request",
				"headers_json": headersJSON,
				"error":        err.Error(),
			})
			return "", fmt.Errorf("invalid headers JSON: %w", err)
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		logDebug(ctx, "Custom headers set", map[string]interface{}{
			"tool":          "http_request",
			"header_count":  len(headers),
		})
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		logError(ctx, "HTTP request failed", map[string]interface{}{
			"tool":     "http_request",
			"method":   method,
			"url":      url,
			"error":    err.Error(),
			"duration": time.Since(startTime).Milliseconds(),
		})
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logError(ctx, "Failed to read response body", map[string]interface{}{
			"tool":     "http_request",
			"method":   method,
			"url":      url,
			"status":   resp.StatusCode,
			"error":    err.Error(),
		})
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details
	logLevel := "INFO"
	if resp.StatusCode >= 500 {
		logLevel = "ERROR"
	} else if resp.StatusCode >= 400 {
		logLevel = "WARN"
	} else if duration > 5*time.Second {
		logLevel = "WARN" // Slow request warning
	}

	logFields := map[string]interface{}{
		"tool":          "http_request",
		"method":        method,
		"url":           url,
		"status":        resp.StatusCode,
		"duration_ms":   duration.Milliseconds(),
		"response_size": len(respBody),
		"content_type":  resp.Header.Get("Content-Type"),
	}

	switch logLevel {
	case "ERROR":
		logError(ctx, "HTTP request completed with server error", logFields)
	case "WARN":
		logWarn(ctx, "HTTP request completed with warning", logFields)
	default:
		logInfo(ctx, "HTTP request completed successfully", logFields)
	}

	// Build response
	result := formatHTTPResponse(method, url, resp.StatusCode, resp.Header, respBody, duration)
	return result, nil
}

// formatHTTPResponse formats the HTTP response into a readable string
func formatHTTPResponse(method, url string, statusCode int, headers http.Header, body []byte, duration time.Duration) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("HTTP %s %s\n", method, url))
	result.WriteString(fmt.Sprintf("Status: %d %s\n", statusCode, http.StatusText(statusCode)))
	result.WriteString(fmt.Sprintf("Duration: %v\n", duration))
	result.WriteString(fmt.Sprintf("Content-Length: %d bytes\n", len(body)))

	// Add important headers
	if ct := headers.Get("Content-Type"); ct != "" {
		result.WriteString(fmt.Sprintf("Content-Type: %s\n", ct))
	}

	result.WriteString("\nResponse Body:\n")

	// Try to parse as JSON for better formatting
	if isJSON(headers.Get("Content-Type")) {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			result.WriteString(prettyJSON.String())
		} else {
			result.WriteString(string(body))
		}
	} else {
		// Limit body size for non-JSON responses
		bodyStr := string(body)
		if len(bodyStr) > 1000 {
			result.WriteString(bodyStr[:1000])
			result.WriteString(fmt.Sprintf("\n... (truncated, %d more bytes)", len(bodyStr)-1000))
		} else {
			result.WriteString(bodyStr)
		}
	}

	return result.String()
}

// isJSON checks if content type is JSON
func isJSON(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}
