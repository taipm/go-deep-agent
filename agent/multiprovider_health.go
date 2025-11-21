// Package agent implements health checking functionality for MultiProvider
// This file contains the health checking system that monitors provider availability
package agent

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthChecker performs health checks on LLM providers
type HealthChecker struct {
	config      *MultiProviderConfig
	logger      Logger
	httpClient  *http.Client

	// Health check state
	healthStatus map[string]*ProviderHealthStatus
	mu           sync.RWMutex
}

// ProviderHealthStatus contains the health status of a provider
type ProviderHealthStatus struct {
	// Current status
	Status        ProviderStatus `json:"status"`
	LastCheck     time.Time      `json:"last_check"`
	LastSuccess   time.Time      `json:"last_success,omitempty"`
	LastFailure   time.Time      `json:"last_failure,omitempty"`

	// Health metrics
	ConsecutiveSuccesses int           `json:"consecutive_successes"`
	ConsecutiveFailures  int           `json:"consecutive_failures"`
	TotalSuccesses       int64         `json:"total_successes"`
	TotalFailures        int64         `json:"total_failures"`
	UptimePercentage     float64       `json:"uptime_percentage"`

	// Response time metrics
	LastResponseTime     time.Duration `json:"last_response_time"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	MinResponseTime      time.Duration `json:"min_response_time"`
	MaxResponseTime      time.Duration `json:"max_response_time"`

	// Error information
	LastError            string        `json:"last_error,omitempty"`
	ErrorCount           int           `json:"error_count"`

	// Health check configuration
	CheckInterval        time.Duration `json:"check_interval"`
	CheckTimeout         time.Duration `json:"check_timeout"`
	HealthyThreshold     int           `json:"healthy_threshold"`
	UnhealthyThreshold   int           `json:"unhealthy_threshold"`
}

// HealthCheckResult represents the result of a single health check
type HealthCheckResult struct {
	Provider       string        `json:"provider"`
	Status         ProviderStatus `json:"status"`
	ResponseTime   time.Duration `json:"response_time"`
	Error          string        `json:"error,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(config *MultiProviderConfig) *HealthChecker {
	return &HealthChecker{
		config:      config,
		logger:      NewStdLogger(LogLevelInfo),
		httpClient: &http.Client{
			Timeout: config.HealthCheckTimeout,
		},
		healthStatus: make(map[string]*ProviderHealthStatus),
	}
}

// Start begins the health checking process for all providers
func (hc *HealthChecker) Start(providers []*ProviderConfig, interval, timeout time.Duration) {
	hc.logger.Info(context.Background(), "Starting health checker",
		F("interval", interval),
		F("timeout", timeout),
		F("providers", len(providers)))

	// Initialize health status for all providers
	hc.mu.Lock()
	for _, provider := range providers {
		hc.healthStatus[provider.Name] = &ProviderHealthStatus{
			Status:             ProviderStatusUnknown,
			CheckInterval:      interval,
			CheckTimeout:       timeout,
			HealthyThreshold:   2, // Need 2 consecutive successes to be healthy
			UnhealthyThreshold: 3, // Need 3 consecutive failures to be unhealthy
		}
	}
	hc.mu.Unlock()

	// Start health checking goroutine
	go hc.healthCheckLoop(providers, interval, timeout)
}

// Stop stops the health checking process
func (hc *HealthChecker) Stop() {
	hc.logger.Info(context.Background(), "Stopping health checker")
	// In a real implementation, you might use context cancellation or a stop channel
}

// GetHealthStatus returns the current health status of all providers
func (hc *HealthChecker) GetHealthStatus() map[string]*ProviderHealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	// Return a copy to avoid race conditions
	status := make(map[string]*ProviderHealthStatus)
	for name, health := range hc.healthStatus {
		statusCopy := *health
		status[name] = &statusCopy
	}

	return status
}

// GetProviderHealth returns the health status of a specific provider
func (hc *HealthChecker) GetProviderHealth(providerName string) (*ProviderHealthStatus, error) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	health, exists := hc.healthStatus[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	// Return a copy to avoid race conditions
	healthCopy := *health
	return &healthCopy, nil
}

// IsHealthy checks if a provider is currently healthy
func (hc *HealthChecker) IsHealthy(providerName string) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	health, exists := hc.healthStatus[providerName]
	if !exists {
		return false
	}

	return health.Status == ProviderStatusHealthy
}

// healthCheckLoop runs the health checking process at regular intervals
func (hc *HealthChecker) healthCheckLoop(providers []*ProviderConfig, interval, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial health check
	hc.checkAllProviders(providers, timeout)

	for range ticker.C {
		hc.checkAllProviders(providers, timeout)
	}
}

// checkAllProviders performs health checks on all providers
func (hc *HealthChecker) checkAllProviders(providers []*ProviderConfig, timeout time.Duration) {
	var wg sync.WaitGroup

	for _, provider := range providers {
		wg.Add(1)
		go func(p *ProviderConfig) {
			defer wg.Done()
			hc.checkProvider(p, timeout)
		}(provider)
	}

	wg.Wait()
}

// checkProvider performs a health check on a single provider
func (hc *HealthChecker) checkProvider(provider *ProviderConfig, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	result := &HealthCheckResult{
		Provider:  provider.Name,
		Timestamp: start,
	}

	// Perform different health checks based on provider type
	var err error
	switch provider.Type {
	case "openai":
		err = hc.checkOpenAIProvider(ctx, provider)
	case "ollama":
		err = hc.checkOllamaProvider(ctx, provider)
	case "gemini":
		err = hc.checkGeminiProvider(ctx, provider)
	case "adapter":
		err = hc.checkAdapterProvider(ctx, provider)
	default:
		err = fmt.Errorf("unsupported provider type: %s", provider.Type)
	}

	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Status = ProviderStatusUnhealthy
		result.Error = err.Error()
		hc.logger.Error(ctx, "Health check failed",
			F("provider", provider.Name),
			F("error", err.Error()),
			F("response_time", result.ResponseTime))
	} else {
		result.Status = ProviderStatusHealthy
		hc.logger.Debug(ctx, "Health check passed",
			F("provider", provider.Name),
			F("response_time", result.ResponseTime))
	}

	// Update health status
	hc.updateHealthStatus(provider.Name, result)
}

// checkOpenAIProvider performs health check for OpenAI provider
func (hc *HealthChecker) checkOpenAIProvider(ctx context.Context, provider *ProviderConfig) error {
	if provider.APIKey == "" {
		return fmt.Errorf("API key is not configured")
	}

	// Create a simple request to check OpenAI availability
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenAI API returned status %d", resp.StatusCode)
	}

	return nil
}

// checkOllamaProvider performs health check for Ollama provider
func (hc *HealthChecker) checkOllamaProvider(ctx context.Context, provider *ProviderConfig) error {
	baseURL := "http://localhost:11434"
	if provider.BaseURL != "" {
		baseURL = provider.BaseURL
	}

	// Check Ollama availability by listing models
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	return nil
}

// checkGeminiProvider performs health check for Gemini provider
func (hc *HealthChecker) checkGeminiProvider(ctx context.Context, provider *ProviderConfig) error {
	if provider.APIKey == "" {
		return fmt.Errorf("API key is not configured")
	}

	// For now, simulate Gemini health check (in real implementation, would call Gemini API)
	// This is a placeholder until Gemini adapter is fully implemented
	if provider.APIKey == "invalid-key" {
		return fmt.Errorf("invalid API key")
	}

	// Simulate some latency
	time.Sleep(50 * time.Millisecond)

	return nil
}

// checkAdapterProvider performs health check for adapter provider
func (hc *HealthChecker) checkAdapterProvider(ctx context.Context, provider *ProviderConfig) error {
	if provider.Adapter == nil {
		return fmt.Errorf("adapter is not configured")
	}

	// Create a simple completion request to test adapter
	req := &CompletionRequest{
		Model:   provider.Model,
		Messages: []Message{{Role: "user", Content: "health check"}},
		MaxTokens: 1,
	}

	// Test adapter with a very short timeout
	adapterCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := provider.Adapter.Complete(adapterCtx, req)
	if err != nil {
		return fmt.Errorf("adapter health check failed: %w", err)
	}

	return nil
}

// updateHealthStatus updates the health status of a provider based on check result
func (hc *HealthChecker) updateHealthStatus(providerName string, result *HealthCheckResult) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	health := hc.healthStatus[providerName]
	if health == nil {
		// Initialize if not exists
		health = &ProviderHealthStatus{
			HealthyThreshold:   2,
			UnhealthyThreshold: 3,
		}
		hc.healthStatus[providerName] = health
	}

	// Update timestamps
	health.LastCheck = result.Timestamp

	if result.Status == ProviderStatusHealthy {
		health.LastSuccess = result.Timestamp
		health.ConsecutiveSuccesses++
		health.ConsecutiveFailures = 0
		health.TotalSuccesses++

		// Update response time metrics
		health.LastResponseTime = result.ResponseTime
		if health.AverageResponseTime == 0 {
			health.AverageResponseTime = result.ResponseTime
			health.MinResponseTime = result.ResponseTime
			health.MaxResponseTime = result.ResponseTime
		} else {
			// Simple exponential moving average for average response time
			health.AverageResponseTime = (health.AverageResponseTime*9 + result.ResponseTime) / 10
			if result.ResponseTime < health.MinResponseTime {
				health.MinResponseTime = result.ResponseTime
			}
			if result.ResponseTime > health.MaxResponseTime {
				health.MaxResponseTime = result.ResponseTime
			}
		}

		// Clear error information
		health.LastError = ""

		// Check if provider should be marked as healthy
		if health.Status != ProviderStatusHealthy && health.ConsecutiveSuccesses >= health.HealthyThreshold {
			health.Status = ProviderStatusHealthy
			hc.logger.Info(context.Background(), "Provider marked as healthy",
				F("provider", providerName),
				F("consecutive_successes", health.ConsecutiveSuccesses))
		}
	} else {
		health.LastFailure = result.Timestamp
		health.ConsecutiveFailures++
		health.ConsecutiveSuccesses = 0
		health.TotalFailures++
		health.LastError = result.Error
		health.ErrorCount++

		// Check if provider should be marked as unhealthy
		if health.Status != ProviderStatusUnhealthy && health.ConsecutiveFailures >= health.UnhealthyThreshold {
			health.Status = ProviderStatusUnhealthy
			hc.logger.Info(context.Background(), "Provider marked as unhealthy",
				F("provider", providerName),
				F("consecutive_failures", health.ConsecutiveFailures),
				F("error", result.Error))
		}
	}

	// Calculate uptime percentage
	totalChecks := health.TotalSuccesses + health.TotalFailures
	if totalChecks > 0 {
		health.UptimePercentage = float64(health.TotalSuccesses) / float64(totalChecks) * 100.0
	}
}

// GetHealthyProviders returns a list of currently healthy providers
func (hc *HealthChecker) GetHealthyProviders() []string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var healthy []string
	for name, health := range hc.healthStatus {
		if health.Status == ProviderStatusHealthy {
			healthy = append(healthy, name)
		}
	}

	return healthy
}

// ForceHealthCheck forces an immediate health check for a specific provider
func (hc *HealthChecker) ForceHealthCheck(provider *ProviderConfig) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		Provider:  provider.Name,
		Timestamp: time.Now(),
	}

	start := time.Now()

	// Perform health check
	var err error
	switch provider.Type {
	case "openai":
		err = hc.checkOpenAIProvider(context.Background(), provider)
	case "ollama":
		err = hc.checkOllamaProvider(context.Background(), provider)
	case "gemini":
		err = hc.checkGeminiProvider(context.Background(), provider)
	case "adapter":
		err = hc.checkAdapterProvider(context.Background(), provider)
	default:
		err = fmt.Errorf("unsupported provider type: %s", provider.Type)
	}

	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Status = ProviderStatusUnhealthy
		result.Error = err.Error()
	} else {
		result.Status = ProviderStatusHealthy
	}

	// Update health status
	hc.updateHealthStatus(provider.Name, result)

	return result, err
}