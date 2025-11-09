package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPRequestTool(t *testing.T) {
	t.Run("GET Request", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success"}`))
		}))
		defer server.Close()

		tool := NewHTTPRequestTool()
		args := `{"method": "GET", "url": "` + server.URL + `"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		if !strings.Contains(result, "200 OK") {
			t.Errorf("Expected 200 OK, got: %s", result)
		}
		if !strings.Contains(result, "success") {
			t.Errorf("Response body not found: %s", result)
		}
	})

	t.Run("POST Request with Body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"created": true}`))
		}))
		defer server.Close()

		tool := NewHTTPRequestTool()
		args := `{"method": "POST", "url": "` + server.URL + `", "body": "{\"name\": \"test\"}"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("POST request failed: %v", err)
		}
		if !strings.Contains(result, "201") {
			t.Errorf("Expected 201 status, got: %s", result)
		}
	})

	t.Run("Custom Headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("Authorization header not found")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		tool := NewHTTPRequestTool()
		headers := `{"Authorization": "Bearer test-token"}`
		args := `{"method": "GET", "url": "` + server.URL + `", "headers": "` + strings.ReplaceAll(headers, `"`, `\"`) + `"}`

		_, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("Request with headers failed: %v", err)
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(3 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		tool := NewHTTPRequestTool()
		args := `{"method": "GET", "url": "` + server.URL + `", "timeout_seconds": 1}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected timeout error")
		}
		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") && !strings.Contains(err.Error(), "context") {
			t.Logf("Got error (may be valid timeout error): %v", err)
		}
	})

	t.Run("Invalid Method", func(t *testing.T) {
		tool := NewHTTPRequestTool()
		args := `{"method": "INVALID", "url": "http://example.com"}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected error for invalid method")
		}
	})

	t.Run("Invalid URL", func(t *testing.T) {
		tool := NewHTTPRequestTool()
		args := `{"method": "GET", "url": "not-a-url"}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("Empty URL", func(t *testing.T) {
		tool := NewHTTPRequestTool()
		args := `{"method": "GET", "url": ""}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected error for empty URL")
		}
	})
}

func TestIsValidHTTPMethod(t *testing.T) {
	tests := []struct {
		method string
		valid  bool
	}{
		{"GET", true},
		{"POST", true},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", false},
		{"HEAD", false},
		{"OPTIONS", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := isValidHTTPMethod(tt.method)
			if result != tt.valid {
				t.Errorf("isValidHTTPMethod(%s) = %v, want %v", tt.method, result, tt.valid)
			}
		})
	}
}

func TestFormatHTTPResponse(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	body := []byte(`{"test": "data"}`)
	duration := 100 * time.Millisecond

	result := formatHTTPResponse("GET", "http://example.com", 200, headers, body, duration)

	if !strings.Contains(result, "GET http://example.com") {
		t.Error("Result missing method and URL")
	}
	if !strings.Contains(result, "200 OK") {
		t.Error("Result missing status code")
	}
	if !strings.Contains(result, "application/json") {
		t.Error("Result missing content type")
	}
	if !strings.Contains(result, "test") {
		t.Error("Result missing body content")
	}
}

func TestIsJSON(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"text/plain", false},
		{"text/html", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := isJSON(tt.contentType)
			if result != tt.expected {
				t.Errorf("isJSON(%s) = %v, want %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

func TestHTTPRequestHandler(t *testing.T) {
	t.Run("Invalid JSON", func(t *testing.T) {
		_, err := httpRequestHandler("invalid json")
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("Valid Arguments", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		args := map[string]interface{}{
			"method": "GET",
			"url":    server.URL,
		}
		argsJSON, _ := json.Marshal(args)

		result, err := httpRequestHandler(string(argsJSON))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if !strings.Contains(result, "200 OK") {
			t.Errorf("Unexpected result: %s", result)
		}
	})
}
