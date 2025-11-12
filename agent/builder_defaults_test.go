package agent

import (
	"testing"
	"time"
)

// TestWithDefaults_BasicConfiguration verifies that WithDefaults() sets all expected defaults
func TestWithDefaultsBasicConfiguration(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").WithDefaults()

	// Verify maxHistory is set to 20
	if builder.maxHistory != 20 {
		t.Errorf("Expected maxHistory=20, got %d", builder.maxHistory)
	}

	// Verify retry is set to 3
	if builder.maxRetries != 3 {
		t.Errorf("Expected maxRetries=3, got %d", builder.maxRetries)
	}

	// Verify timeout is set to 30 seconds
	if builder.timeout != 30*time.Second {
		t.Errorf("Expected timeout=30s, got %v", builder.timeout)
	}

	// Verify exponential backoff is enabled
	if !builder.useExpBackoff {
		t.Error("Expected exponentialBackoff to be enabled")
	}
}

// TestWithDefaults_Customization verifies that defaults can be overridden via chaining
func TestWithDefaultsCustomization(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").
		WithDefaults().
		WithMaxHistory(50).           // Override memory to 50
		WithRetry(5).                 // Override retry to 5
		WithTimeout(60 * time.Second) // Override timeout to 60s

	// Verify customizations
	if builder.maxHistory != 50 {
		t.Errorf("Expected maxHistory=50, got %d", builder.maxHistory)
	}

	if builder.maxRetries != 5 {
		t.Errorf("Expected maxRetries=5, got %d", builder.maxRetries)
	}

	if builder.timeout != 60*time.Second {
		t.Errorf("Expected timeout=60s, got %v", builder.timeout)
	}

	// Exponential backoff should still be enabled
	if !builder.useExpBackoff {
		t.Error("Expected exponentialBackoff to still be enabled")
	}
}

// TestWithDefaults_Idempotent verifies that calling WithDefaults() twice doesn't duplicate settings
func TestWithDefaultsIdempotent(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").
		WithDefaults().
		WithDefaults() // Call twice

	// Verify settings are still the defaults (not doubled)
	if builder.maxHistory != 20 {
		t.Errorf("Expected maxHistory=20, got %d", builder.maxHistory)
	}

	if builder.maxRetries != 3 {
		t.Errorf("Expected maxRetries=3, got %d", builder.maxRetries)
	}

	if builder.timeout != 30*time.Second {
		t.Errorf("Expected timeout=30s, got %v", builder.timeout)
	}

	if !builder.useExpBackoff {
		t.Error("Expected exponentialBackoff to be enabled")
	}
}

// TestWithDefaults_Override verifies that explicit configuration before WithDefaults() gets overridden
func TestWithDefaultsOverride(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").
		WithMaxHistory(100). // Set custom value first
		WithDefaults()       // Then apply defaults (should override)

	// WithDefaults() should override the previous setting
	expectedMaxHistory := 20
	if builder.maxHistory != expectedMaxHistory {
		t.Errorf("Expected maxHistory=%d (default overrides custom), got %d", expectedMaxHistory, builder.maxHistory)
	}
}

// TestWithDefaults_Chaining verifies that method chaining works correctly
func TestWithDefaultsChaining(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").
		WithDefaults().
		WithTemperature(0.7).
		WithMaxTokens(2000)

	// Verify defaults are set
	expectedMaxHistory := 20
	if builder.maxHistory != expectedMaxHistory {
		t.Errorf("Expected maxHistory=%d, got %d", expectedMaxHistory, builder.maxHistory)
	}

	// Verify chained settings
	if builder.temperature == nil || *builder.temperature != 0.7 {
		t.Errorf("Expected temperature=0.7, got %v", builder.temperature)
	}

	if builder.maxTokens == nil || *builder.maxTokens != 2000 {
		t.Errorf("Expected maxTokens=2000, got %v", builder.maxTokens)
	}
}

// TestWithDefaults_DisableMemory verifies that memory can be disabled after WithDefaults()
func TestWithDefaultsDisableMemory(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").
		WithDefaults().
		DisableMemory()

	// Verify memory is disabled (memoryEnabled flag)
	if builder.memoryEnabled {
		t.Error("Expected memory to be disabled (memoryEnabled=false)")
	}

	// But retry and timeout should still be set
	if builder.maxRetries != 3 {
		t.Errorf("Expected maxRetries=3, got %d", builder.maxRetries)
	}

	if builder.timeout != 30*time.Second {
		t.Errorf("Expected timeout=30s, got %v", builder.timeout)
	}
}

// TestWithDefaults_AllConstructors verifies that WithDefaults() works with all constructors
func TestWithDefaultsAllConstructors(t *testing.T) {
	tests := []struct {
		name    string
		builder *Builder
	}{
		{
			name:    "NewOpenAI",
			builder: NewOpenAI("gpt-4", "sk-test").WithDefaults(),
		},
		{
			name:    "NewOllama",
			builder: NewOllama("llama2").WithDefaults(),
		},
		{
			name:    "New",
			builder: New("openai", "gpt-4").WithAPIKey("sk-test").WithDefaults(),
		},
	}

	expectedMaxHistory := 20
	expectedMaxRetries := 3
	expectedTimeout := 30 * time.Second

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify defaults are set for all constructors
			if tt.builder.maxHistory != expectedMaxHistory {
				t.Errorf("%s: Expected maxHistory=%d, got %d", tt.name, expectedMaxHistory, tt.builder.maxHistory)
			}

			if tt.builder.maxRetries != expectedMaxRetries {
				t.Errorf("%s: Expected maxRetries=%d, got %d", tt.name, expectedMaxRetries, tt.builder.maxRetries)
			}

			if tt.builder.timeout != expectedTimeout {
				t.Errorf("%s: Expected timeout=%v, got %v", tt.name, expectedTimeout, tt.builder.timeout)
			}

			if !tt.builder.useExpBackoff {
				t.Errorf("%s: Expected exponentialBackoff to be enabled", tt.name)
			}
		})
	}
}

// TestWithDefaults_NoSideEffects verifies that WithDefaults() doesn't affect opt-in features
func TestWithDefaultsNoSideEffects(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").WithDefaults()

	// Verify opt-in features remain disabled
	if builder.tools != nil && len(builder.tools) > 0 {
		t.Error("Expected tools to be nil/empty (opt-in feature)")
	}

	if builder.logger != nil {
		t.Error("Expected logger to be nil (opt-in feature)")
	}

	if builder.cache != nil {
		t.Error("Expected cache to be nil (opt-in feature)")
	}

	if builder.enableParallel {
		t.Error("Expected parallel execution to be disabled (opt-in feature)")
	}
}

// TestWithDefaults_EnablesAutoMemory verifies that WithDefaults() enables autoMemory (Bug Fix v0.7.10)
// This test addresses the bug where WithDefaults() was documented to enable memory but didn't
func TestWithDefaultsEnablesAutoMemory(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").WithDefaults()

	// Verify autoMemory is enabled (critical for conversational agents)
	if !builder.autoMemory {
		t.Error("Expected autoMemory to be enabled by WithDefaults()")
	}

	// Verify maxHistory is also set (capacity)
	if builder.maxHistory != 20 {
		t.Errorf("Expected maxHistory=20, got %d", builder.maxHistory)
	}
}

// TestWithDefaults_MemoryCanBeDisabled verifies that memory can be disabled after WithDefaults()
// This ensures users can opt-out if they really don't want memory
func TestWithDefaultsMemoryCanBeDisabled(t *testing.T) {
	builder := NewOpenAI("gpt-4", "sk-test").
		WithDefaults().
		DisableMemory()

	// Verify memory is now disabled
	if builder.memoryEnabled {
		t.Error("Expected memoryEnabled to be false after DisableMemory()")
	}

	// Other defaults should remain
	if builder.maxRetries != 3 {
		t.Errorf("Expected maxRetries=3, got %d", builder.maxRetries)
	}
}
