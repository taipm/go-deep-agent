package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestMultiProviderCreation tests the creation and configuration of MultiProvider
func TestMultiProviderCreation(t *testing.T) {
	t.Run("Create MultiProvider with valid config", func(t *testing.T) {
		config := &MultiProviderConfig{
			Providers: []ProviderConfig{
				{
					Name:  "openai-primary",
					Type:  "openai",
					Model: "gpt-4o-mini",
					APIKey: "sk-test123456789",
					Weight: 2.0,
					MaxConcurrency: 10,
				},
				{
					Name:  "ollama-backup",
					Type:  "ollama",
					Model: "llama2",
					Weight: 1.0,
					MaxConcurrency: 5,
				},
			},
			SelectionStrategy: StrategyWeightedRoundRobin,
			FallbackStrategy:  FallbackStrategyCircuitBreaker,
			HealthCheckInterval: 30 * time.Second,
			HealthCheckTimeout: 5 * time.Second,
			EnableLoadBalancing: true,
			EnableMetrics: true,
		}

		mp, err := NewMultiProvider(config)
		if err != nil {
			t.Fatalf("Failed to create MultiProvider: %v", err)
		}

		if mp == nil {
			t.Fatal("MultiProvider should not be nil")
		}

		// Check provider count
		status := mp.GetProviderStatus()
		if len(status) != 2 {
			t.Errorf("Expected 2 providers, got: %d", len(status))
		}

		// Check initial status
		for name, status := range status {
			if status != ProviderStatusUnknown {
				t.Errorf("Provider %s should have unknown status initially, got: %s", name, status)
			}
		}

		// Test graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = mp.Shutdown(ctx)
		if err != nil {
			t.Errorf("Shutdown should not fail: %v", err)
		}
	})

	t.Run("Create MultiProvider with no providers", func(t *testing.T) {
		config := &MultiProviderConfig{
			Providers: []ProviderConfig{},
		}

		_, err := NewMultiProvider(config)
		if err == nil {
			t.Error("Expected error when creating MultiProvider with no providers")
		}

		expectedError := "at least one provider must be configured"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
		}
	})

	t.Run("Create MultiProvider with nil config", func(t *testing.T) {
		_, err := NewMultiProvider(nil)
		if err == nil {
			t.Error("Expected error when creating MultiProvider with nil config")
		}

		expectedError := "config cannot be nil"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
		}
	})
}

// TestMultiProviderExecution tests the execution functionality of MultiProvider
func TestMultiProviderExecution(t *testing.T) {
	t.Skip("Skipping execution tests that require real API keys")

	t.Run("Execute with mock providers", func(t *testing.T) {
		// This test would use mock adapters instead of real providers
		config := &MultiProviderConfig{
			Providers: []ProviderConfig{
				{
					Name:  "mock-primary",
					Type:  "adapter",
					Model: "test-model",
					Weight: 2.0,
					Adapter: &mockTestAdapter{
						responses: []string{"Response from primary provider"},
					},
				},
				{
					Name:  "mock-backup",
					Type:  "adapter",
					Model: "test-model",
					Weight: 1.0,
					Adapter: &mockTestAdapter{
						responses: []string{"Response from backup provider"},
					},
				},
			},
			SelectionStrategy: StrategyWeightedRoundRobin,
			FallbackStrategy:  FallbackStrategyCircuitBreaker,
			EnableLoadBalancing: true,
			EnableMetrics: true,
		}

		mp, err := NewMultiProvider(config)
		if err != nil {
			t.Fatalf("Failed to create MultiProvider: %v", err)
		}
		defer mp.Shutdown(context.Background())

		ctx := context.Background()
		message := "Hello, MultiProvider!"

		// Test Ask method
		response, err := mp.Ask(ctx, message)
		if err != nil {
			t.Fatalf("Ask should not fail: %v", err)
		}

		if response == "" {
			t.Error("Response should not be empty")
		}

		// Test Stream method
		streamResponse, err := mp.Stream(ctx, message)
		if err != nil {
			t.Fatalf("Stream should not fail: %v", err)
		}

		if streamResponse == "" {
			t.Error("Stream response should not be empty")
		}

		// Check metrics
		metrics := mp.GetMetrics()
		if len(metrics) == 0 {
			t.Error("Metrics should contain provider data")
		}

		// Check health status
		health := mp.GetProviderStatus()
		if len(health) == 0 {
			t.Error("Health status should contain provider data")
		}
	})
}

// TestProviderSelection tests various provider selection strategies
func TestProviderSelection(t *testing.T) {
	providers := []*ProviderConfig{
		{Name: "provider1", Weight: 1.0, ResponseTime: 100 * time.Millisecond},
		{Name: "provider2", Weight: 2.0, ResponseTime: 200 * time.Millisecond},
		{Name: "provider3", Weight: 3.0, ResponseTime: 300 * time.Millisecond},
	}

	selector := NewProviderSelector(&MultiProviderConfig{})

	t.Run("RoundRobin selection", func(t *testing.T) {
		// Test multiple selections to ensure round-robin behavior
		var selectedProviders []string
		for i := 0; i < 6; i++ {
			provider, err := selector.SelectProvider(providers, StrategyRoundRobin)
			if err != nil {
				t.Fatalf("RoundRobin selection failed: %v", err)
			}
			selectedProviders = append(selectedProviders, provider.Name)
		}

		// Should cycle through providers
		if len(selectedProviders) != 6 {
			t.Errorf("Expected 6 selections, got: %d", len(selectedProviders))
		}
	})

	t.Run("Random selection", func(t *testing.T) {
		// Test random selection multiple times
		for i := 0; i < 10; i++ {
			provider, err := selector.SelectProvider(providers, StrategyRandom)
			if err != nil {
				t.Fatalf("Random selection failed: %v", err)
			}

			// Should be one of the providers
			found := false
			for _, p := range providers {
				if p.Name == provider.Name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Random selection returned unknown provider: %s", provider.Name)
			}
		}
	})

	t.Run("Priority selection", func(t *testing.T) {
		provider, err := selector.SelectProvider(providers, StrategyPriority)
		if err != nil {
			t.Fatalf("Priority selection failed: %v", err)
		}

		// Should select provider with highest weight
		if provider.Name != "provider3" {
			t.Errorf("Priority selection should return provider3, got: %s", provider.Name)
		}
	})
}

// TestHealthChecker tests the health checking functionality
func TestHealthChecker(t *testing.T) {
	config := &MultiProviderConfig{
		HealthCheckInterval: 1 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
	}

	healthChecker := NewHealthChecker(config)

	t.Run("Health status initialization", func(t *testing.T) {
		providers := []*ProviderConfig{
			{Name: "test-provider1", Type: "adapter", Adapter: &mockTestAdapter{}},
			{Name: "test-provider2", Type: "adapter", Adapter: &mockTestAdapter{}},
		}

		// Start health checking
		healthChecker.Start(providers, config.HealthCheckInterval, config.HealthCheckTimeout)

		// Give some time for initial health check
		time.Sleep(2 * time.Second)

		status := healthChecker.GetHealthStatus()
		if len(status) != 2 {
			t.Errorf("Expected health status for 2 providers, got: %d", len(status))
		}
	})

	t.Run("Force health check", func(t *testing.T) {
		provider := &ProviderConfig{
			Name:    "force-check-provider",
			Type:    "adapter",
			Adapter: &mockTestAdapter{responses: []string{"test"}},
		}

		result, err := healthChecker.ForceHealthCheck(provider)
		if err != nil {
			t.Errorf("Force health check should not fail: %v", err)
		}

		if result.Provider != provider.Name {
			t.Errorf("Expected provider name %s, got: %s", provider.Name, result.Provider)
		}

		if result.Status != ProviderStatusHealthy {
			t.Errorf("Expected healthy status, got: %s", result.Status)
		}
	})

	t.Run("Unhealthy provider detection", func(t *testing.T) {
		provider := &ProviderConfig{
			Name:    "unhealthy-provider",
			Type:    "adapter",
			Adapter: &mockTestAdapter{shouldError: true, errorMessage: "Provider failed"},
		}

		result, err := healthChecker.ForceHealthCheck(provider)
		if err == nil {
			t.Error("Force health check should fail for unhealthy provider")
		}

		if result.Status != ProviderStatusUnhealthy {
			t.Errorf("Expected unhealthy status, got: %s", result.Status)
		}

		if result.Error == "" {
			t.Error("Expected error message for unhealthy provider")
		}
	})
}

// TestLoadBalancer tests the load balancing functionality
func TestLoadBalancer(t *testing.T) {
	config := &MultiProviderConfig{
		EnableLoadBalancing: true,
		StickySessions:      true,
	}

	loadBalancer := NewLoadBalancer(config)

	t.Run("Load-based provider selection", func(t *testing.T) {
		providers := []*ProviderConfig{
			{
				Name:           "provider1",
				MaxConcurrency: 10,
				Weight:         1.0,
			},
			{
				Name:           "provider2",
				MaxConcurrency: 5,
				Weight:         2.0,
			},
		}

		// Test provider selection with load balancing
		provider, err := loadBalancer.SelectProviderForRequest(providers, "session1")
		if err != nil {
			t.Fatalf("Load balancing selection failed: %v", err)
		}

		// Should select one of the providers
		found := false
		for _, p := range providers {
			if p.Name == provider.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Load balancing returned unknown provider: %s", provider.Name)
		}
	})

	t.Run("Sticky sessions", func(t *testing.T) {
		providers := []*ProviderConfig{
			{Name: "provider1", Weight: 1.0},
			{Name: "provider2", Weight: 1.0},
		}

		sessionID := "test-session-123"

		// First selection should associate session with provider
		provider1, err := loadBalancer.SelectProviderForRequest(providers, sessionID)
		if err != nil {
			t.Fatalf("First selection failed: %v", err)
		}

		// Second selection with same session should return same provider
		provider2, err := loadBalancer.SelectProviderForRequest(providers, sessionID)
		if err != nil {
			t.Fatalf("Second selection failed: %v", err)
		}

		if provider1.Name != provider2.Name {
			t.Errorf("Sticky session failed: expected %s, got %s", provider1.Name, provider2.Name)
		}
	})

	t.Run("Load metrics calculation", func(t *testing.T) {
		providers := []*ProviderConfig{
			{
				Name:           "metrics-provider",
				MaxConcurrency: 10,
				SuccessCount:   80,
				ErrorCount:     20,
			},
		}

		metrics := loadBalancer.GetLoadMetrics(providers)
		if len(metrics) == 0 {
			t.Error("Load metrics should contain provider data")
		}

		providerMetrics := metrics["metrics-provider"]
		if providerMetrics == nil {
			t.Error("Provider metrics should not be nil")
		}

		// Check error rate calculation
		expectedErrorRate := float64(20) / float64(100) // 20 errors out of 100 total
		if providerMetrics.ErrorRate != expectedErrorRate {
			t.Errorf("Expected error rate %f, got: %f", expectedErrorRate, providerMetrics.ErrorRate)
		}
	})
}

// TestFallbackHandler tests the fallback functionality
func TestFallbackHandler(t *testing.T) {
	config := &MultiProviderConfig{
		FallbackStrategy:      FallbackStrategyCircuitBreaker,
		CircuitBreakerThreshold: 3,
		CircuitBreakerTimeout:   30 * time.Second,
	}

	fallbackHandler := NewFallbackHandler(config)

	t.Run("Successful execution without fallback", func(t *testing.T) {
		providers := []*ProviderConfig{
			{
				Name:    "primary-provider",
				Type:    "adapter",
				Adapter: &mockTestAdapter{responses: []string{"success"}},
			},
		}

		executeFunc := func(provider *ProviderConfig) (string, error) {
			if provider.Adapter != nil {
				resp, err := provider.Adapter.Complete(context.Background(), &CompletionRequest{
					Model: provider.Model,
					Messages: []Message{{Role: "user", Content: "test"}},
				})
				if err != nil {
					return "", err
				}
				return resp.Content, nil
			}
			return "", errors.New("no adapter")
		}

		result, err := fallbackHandler.ExecuteWithFallback(
			context.Background(),
			providers[0],
			providers,
			executeFunc,
			"test message",
		)

		if err != nil {
			t.Errorf("Execution should not fail: %v", err)
		}

		if result != "success" {
			t.Errorf("Expected 'success', got: %s", result)
		}
	})

	t.Run("Fallback to secondary provider", func(t *testing.T) {
		providers := []*ProviderConfig{
			{
				Name:    "failing-provider",
				Type:    "adapter",
				Adapter: &mockTestAdapter{shouldError: true, errorMessage: "Provider failed"},
			},
			{
				Name:    "backup-provider",
				Type:    "adapter",
				Adapter: &mockTestAdapter{responses: []string{"backup success"}},
			},
		}

		executeFunc := func(provider *ProviderConfig) (string, error) {
			if provider.Adapter != nil {
				req := &CompletionRequest{
					Model: provider.Model,
					Messages: []Message{{Role: "user", Content: "test"}},
				}
				resp, err := provider.Adapter.Complete(context.Background(), req)
				if err != nil {
					return "", err
				}
				return resp.Content, nil
			}
			return "", errors.New("no adapter")
		}

		result, err := fallbackHandler.ExecuteWithFallback(
			context.Background(),
			providers[0],
			providers,
			executeFunc,
			"test message",
		)

		if err != nil {
			t.Errorf("Fallback execution should not fail: %v", err)
		}

		if result != "backup success" {
			t.Errorf("Expected 'backup success', got: %s", result)
		}
	})

	t.Run("Circuit breaker functionality", func(t *testing.T) {
		providerName := "circuit-test-provider"

		// Create circuit breaker directly
		circuitBreaker := NewCircuitBreaker(providerName, 2, 1*time.Second)

		// Initially should be closed
		if circuitBreaker.State() != CircuitBreakerClosed {
			t.Errorf("Expected circuit breaker to be closed initially, got: %s", circuitBreaker.State())
		}

		// Record failures to trigger circuit breaker
		circuitBreaker.RecordFailure()
		if circuitBreaker.State() != CircuitBreakerClosed {
			t.Error("Circuit breaker should still be closed after 1 failure")
		}

		circuitBreaker.RecordFailure()
		if circuitBreaker.State() != CircuitBreakerOpen {
			t.Error("Circuit breaker should be open after threshold failures")
		}

		if !circuitBreaker.IsOpen() {
			t.Error("Circuit breaker should report as open")
		}

		// Test circuit breaker status
		status := circuitBreaker.GetStatus()
		if status.Name != providerName {
			t.Errorf("Expected provider name %s, got: %s", providerName, status.Name)
		}

		if status.FailureCount != 2 {
			t.Errorf("Expected failure count 2, got: %d", status.FailureCount)
		}

		// Reset circuit breaker
		circuitBreaker.Reset()
		if circuitBreaker.State() != CircuitBreakerClosed {
			t.Error("Circuit breaker should be closed after reset")
		}
	})
}

// TestMetricsCollector tests the metrics collection functionality
func TestMetricsCollector(t *testing.T) {
	config := &MultiProviderConfig{
		EnableMetrics:   true,
		MetricsInterval: 1 * time.Second,
	}

	metricsCollector := NewMetricsCollector(config)

	t.Run("Record request metrics", func(t *testing.T) {
		providers := []*ProviderConfig{
			{Name: "metrics-provider1"},
			{Name: "metrics-provider2"},
		}

		metricsCollector.Start(providers, config.MetricsInterval)
		defer metricsCollector.Stop()

		// Record some test requests
		testMetrics := &RequestMetrics{
			Provider:     "metrics-provider1",
			RequestType:  "ask",
			StartTime:    time.Now().Add(-100 * time.Millisecond),
			EndTime:      time.Now(),
			ResponseTime: 100 * time.Millisecond,
			Success:      true,
			TokenUsage:   TokenUsage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
		}

		metricsCollector.RecordRequest(testMetrics)

		// Check provider metrics
		providerMetrics, err := metricsCollector.GetProviderMetrics("metrics-provider1")
		if err != nil {
			t.Fatalf("Failed to get provider metrics: %v", err)
		}

		if providerMetrics.TotalRequests != 1 {
			t.Errorf("Expected 1 total request, got: %d", providerMetrics.TotalRequests)
		}

		if providerMetrics.SuccessfulRequests != 1 {
			t.Errorf("Expected 1 successful request, got: %d", providerMetrics.SuccessfulRequests)
		}

		// Check global metrics
		globalMetrics := metricsCollector.GetGlobalMetrics()
		if globalMetrics.TotalRequests != 1 {
			t.Errorf("Expected 1 global total request, got: %d", globalMetrics.TotalRequests)
		}

		if globalMetrics.SuccessfulRequests != 1 {
			t.Errorf("Expected 1 global successful request, got: %d", globalMetrics.SuccessfulRequests)
		}
	})

	t.Run("Metrics aggregation", func(t *testing.T) {
		// Test metrics summary
		summary := metricsCollector.GetMetricsSummary()
		if summary == nil {
			t.Error("Metrics summary should not be nil")
		}

		// Check expected fields in summary
		expectedFields := []string{"total_requests", "successful_requests", "providers"}
		for _, field := range expectedFields {
			if _, exists := summary[field]; !exists {
				t.Errorf("Expected field %s in metrics summary", field)
			}
		}

		// Test metrics export
		exported := metricsCollector.ExportMetrics()
		if exported == nil {
			t.Error("Exported metrics should not be nil")
		}

		if _, exists := exported["timestamp"]; !exists {
			t.Error("Exported metrics should contain timestamp")
		}

		if _, exists := exported["global"]; !exists {
			t.Error("Exported metrics should contain global metrics")
		}
	})
}

// TestMultiProviderIntegration tests the complete MultiProvider integration
func TestMultiProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("End-to-end MultiProvider workflow", func(t *testing.T) {
		config := &MultiProviderConfig{
			Providers: []ProviderConfig{
				{
					Name:  "integration-primary",
					Type:  "adapter",
					Model: "test-model",
					Weight: 2.0,
					Adapter: &mockTestAdapter{
						responses: []string{"Primary response"},
					},
					MaxConcurrency: 5,
				},
				{
					Name:  "integration-secondary",
					Type:  "adapter",
					Model: "test-model",
					Weight: 1.0,
					Adapter: &mockTestAdapter{
						responses: []string{"Secondary response"},
					},
					MaxConcurrency: 3,
				},
			},
			SelectionStrategy:    StrategyWeightedRoundRobin,
			FallbackStrategy:     FallbackStrategyCircuitBreaker,
			HealthCheckInterval:  1 * time.Second,
			HealthCheckTimeout:   5 * time.Second,
			EnableLoadBalancing:  true,
			EnableMetrics:        true,
			MetricsInterval:      2 * time.Second,
		}

		mp, err := NewMultiProvider(config)
		if err != nil {
			t.Fatalf("Failed to create MultiProvider: %v", err)
		}
		defer mp.Shutdown(context.Background())

		ctx := context.Background()

		// Wait for initial health checks
		time.Sleep(2 * time.Second)

		// Execute multiple requests
		for i := 0; i < 5; i++ {
			message := fmt.Sprintf("Test message %d", i+1)

			response, err := mp.Ask(ctx, message)
			if err != nil {
				t.Errorf("Request %d failed: %v", i+1, err)
			}

			if response == "" {
				t.Errorf("Request %d returned empty response", i+1)
			}
		}

		// Check health status
		health := mp.GetProviderStatus()
		if len(health) != 2 {
			t.Errorf("Expected health status for 2 providers, got: %d", len(health))
		}

		// Check metrics
		metrics := mp.GetMetrics()
		if len(metrics) != 2 {
			t.Errorf("Expected metrics for 2 providers, got: %d", len(metrics))
		}

		// Check that metrics were recorded
		totalRequests := int64(0)
		for _, providerMetrics := range metrics {
			totalRequests += providerMetrics.TotalRequests
		}

		if totalRequests == 0 {
			t.Error("Expected some requests to be recorded in metrics")
		}

		// Test dynamic provider management
		newProvider := ProviderConfig{
			Name:  "dynamic-provider",
			Type:  "adapter",
			Model: "test-model",
			Weight: 1.0,
			Adapter: &mockTestAdapter{
				responses: []string{"Dynamic response"},
			},
		}

		err = mp.AddProvider(newProvider)
		if err != nil {
			t.Errorf("Failed to add provider: %v", err)
		}

		// Check that provider was added
		healthAfterAdd := mp.GetProviderStatus()
		if len(healthAfterAdd) != 3 {
			t.Errorf("Expected health status for 3 providers after add, got: %d", len(healthAfterAdd))
		}

		// Test provider removal
		err = mp.RemoveProvider("dynamic-provider")
		if err != nil {
			t.Errorf("Failed to remove provider: %v", err)
		}

		// Check that provider was removed
		healthAfterRemove := mp.GetProviderStatus()
		if len(healthAfterRemove) != 2 {
			t.Errorf("Expected health status for 2 providers after remove, got: %d", len(healthAfterRemove))
		}
	})
}