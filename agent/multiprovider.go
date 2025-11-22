// Package agent provides MultiProvider support with fallback mechanisms
// This file contains the core MultiProvider implementation for failover and load balancing
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProviderStatus represents the health status of a provider
type ProviderStatus int

const (
	ProviderStatusUnknown ProviderStatus = iota
	ProviderStatusHealthy
	ProviderStatusDegraded
	ProviderStatusUnhealthy
	ProviderStatusDisabled
)

// String returns string representation of ProviderStatus
func (ps ProviderStatus) String() string {
	switch ps {
	case ProviderStatusUnknown:
		return "unknown"
	case ProviderStatusHealthy:
		return "healthy"
	case ProviderStatusDegraded:
		return "degraded"
	case ProviderStatusUnhealthy:
		return "unhealthy"
	case ProviderStatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// ProviderConfig represents configuration for a single provider
type ProviderConfig struct {
	// Core identification
	Name     string `json:"name"`
	Type     string `json:"type"` // "openai", "ollama", "adapter", "custom"
	Model    string `json:"model"`

	// Connection details
	APIKey  string `json:"api_key,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
	Adapter LLMAdapter `json:"-"` // Custom adapter (if type=adapter)

	// Provider-specific settings
	Timeout      time.Duration `json:"timeout"`
	MaxRetries   int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`

	// Health check settings
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout   time.Duration `json:"health_check_timeout"`
	HealthCheckURL       string        `json:"health_check_url,omitempty"`

	// Load balancing settings
	Weight         float64 `json:"weight"`        // Relative weight for load balancing
	MaxConcurrency int     `json:"max_concurrency"` // Max concurrent requests

	// Runtime status
	Status        ProviderStatus `json:"status"`
	LastCheck     time.Time      `json:"last_check"`
	ErrorCount    int           `json:"error_count"`
	SuccessCount  int           `json:"success_count"`
	ResponseTime  time.Duration `json:"avg_response_time"`

	// Rate limiting
	RequestsPerMinute int `json:"requests_per_minute"`

	// Builder instance (for direct providers)
	Builder *Builder `json:"-"`
}

// ProviderMetrics contains performance metrics for a provider
type ProviderMetrics struct {
	// Performance metrics
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`

	// Health metrics
	LastSuccessTime    time.Time     `json:"last_success_time"`
	LastFailureTime    time.Time     `json:"last_failure_time"`
	UptimePercentage   float64       `json:"uptime_percentage"`

	// Current status
	Status             ProviderStatus `json:"status"`
	LastError          string        `json:"last_error,omitempty"`

	// Rate limiting
	CurrentRPS         float64       `json:"current_rps"`
	AllowedRPS         float64       `json:"allowed_rps"`
}

// SelectionStrategy determines how providers are selected
type SelectionStrategy int

const (
	StrategyRoundRobin SelectionStrategy = iota
	StrategyWeightedRoundRobin
	StrategyLeastConnections
	StrategyFastestResponse
	StrategyRandom
	StrategyPriority
	StrategyCustom
)

// FallbackStrategy determines how fallbacks are handled
type FallbackStrategy int

const (
	FallbackStrategyFailFast FallbackStrategy = iota
	FallbackStrategyRetryWithBackoff
	FallbackStrategyCircuitBreaker
	FallbackStrategyGracefulDegradation
	FallbackStrategyCustom
)

// MultiProviderConfig contains configuration for MultiProvider
type MultiProviderConfig struct {
	// Provider configuration
	Providers []ProviderConfig `json:"providers"`

	// Strategy settings
	SelectionStrategy SelectionStrategy `json:"selection_strategy"`
	FallbackStrategy  FallbackStrategy  `json:"fallback_strategy"`

	// Health check settings
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout   time.Duration `json:"health_check_timeout"`

	// Load balancing settings
	EnableLoadBalancing bool          `json:"enable_load_balancing"`
	StickySessions      bool          `json:"sticky_sessions"`

	// Circuit breaker settings
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration   `json:"circuit_breaker_timeout"`

	// Monitoring settings
	EnableMetrics        bool          `json:"enable_metrics"`
	MetricsInterval      time.Duration `json:"metrics_interval"`

	// Logging
	LogLevel             string        `json:"log_level"`
	EnableDetailedLogging bool         `json:"enable_detailed_logging"`
}

// MultiProvider implements multi-provider support with fallback mechanisms
type MultiProvider struct {
	config      *MultiProviderConfig
	providers   []*ProviderConfig
	healthChecker *HealthChecker
	selector     *ProviderSelector
	balancer     *LoadBalancer
	fallback     *FallbackHandler
	metrics      *MetricsCollector
	logger       Logger

	// Runtime state
	mu           sync.RWMutex
	shutdown     chan struct{}
	circuitBreaker map[string]*CircuitBreaker
}

// NewMultiProvider creates a new MultiProvider instance
func NewMultiProvider(config *MultiProviderConfig) (*MultiProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if len(config.Providers) == 0 {
		return nil, fmt.Errorf("at least one provider must be configured")
	}

	mp := &MultiProvider{
		config:         config,
		providers:      make([]*ProviderConfig, 0, len(config.Providers)),
		healthChecker:  NewHealthChecker(config),
		selector:       NewProviderSelector(config),
		balancer:       NewLoadBalancer(config),
		fallback:       NewFallbackHandler(config),
		metrics:        NewMetricsCollector(config),
		shutdown:       make(chan struct{}),
		circuitBreaker: make(map[string]*CircuitBreaker),
	}

	// Initialize providers
	for i := range config.Providers {
		provider := &config.Providers[i]
		mp.providers = append(mp.providers, provider)

		// Create builder for direct providers
		if provider.Type != "adapter" && provider.Adapter == nil {
			builder, err := mp.createBuilderForProvider(provider)
			if err != nil {
				return nil, fmt.Errorf("failed to create builder for provider %s: %w", provider.Name, err)
			}
			provider.Builder = builder
		}

		// Initialize circuit breaker
		mp.circuitBreaker[provider.Name] = NewCircuitBreaker(
			provider.Name,
			config.CircuitBreakerThreshold,
			config.CircuitBreakerTimeout,
		)
	}

	// Start background services
	if config.HealthCheckInterval > 0 {
		go mp.healthChecker.Start(mp.providers, config.HealthCheckInterval, config.HealthCheckTimeout)
	}

	if config.EnableMetrics {
		go mp.metrics.Start(mp.providers, config.MetricsInterval)
	}

	return mp, nil
}

// createBuilderForProvider creates a Builder instance for a provider configuration
func (mp *MultiProvider) createBuilderForProvider(config *ProviderConfig) (*Builder, error) {
	switch config.Type {
	case "openai":
		if config.APIKey == "" {
			return nil, fmt.Errorf("API key is required for OpenAI provider %s", config.Name)
		}
		return NewOpenAI(config.Model, config.APIKey).
			WithTimeout(config.Timeout).
			WithRetry(config.MaxRetries).
			WithRetryDelay(config.RetryDelay).
			WithBaseURL(config.BaseURL), nil

	case "ollama":
		return NewOllama(config.Model).
			WithTimeout(config.Timeout).
			WithRetry(config.MaxRetries).
			WithRetryDelay(config.RetryDelay).
			WithBaseURL(config.BaseURL), nil

	case "gemini":
		if config.APIKey == "" {
			return nil, fmt.Errorf("API key is required for Gemini provider %s", config.Name)
		}
		// Create production-ready Gemini V3 adapter
		geminiAdapter, err := NewGeminiV3Adapter(config.APIKey, config.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini V3 adapter for provider %s: %w", config.Name, err)
		}
		// Return as custom adapter provider
		config.Adapter = geminiAdapter
		config.Type = "adapter"
		return nil, nil

	case "gemini-v3":
		if config.APIKey == "" {
			return nil, fmt.Errorf("API key is required for Gemini V3 provider %s", config.Name)
		}
		// Explicitly create V3 adapter
		geminiAdapter, err := NewGeminiV3Adapter(config.APIKey, config.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini V3 adapter for provider %s: %w", config.Name, err)
		}
		// Return as custom adapter provider
		config.Adapter = geminiAdapter
		config.Type = "adapter"
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}

// Ask executes a request using MultiProvider logic
func (mp *MultiProvider) Ask(ctx context.Context, message string) (string, error) {
	return mp.executeWithFallback(ctx, func(provider *ProviderConfig) (string, error) {
		// Use adapter if available, otherwise use builder
		if provider.Adapter != nil {
			req := &CompletionRequest{
				Model:    provider.Model,
				Messages: []Message{{Role: "user", Content: message}},
			}
			resp, err := provider.Adapter.Complete(ctx, req)
			if err != nil {
				return "", err
			}
			return resp.Content, nil
		}

		if provider.Builder == nil {
			return "", fmt.Errorf("provider %s has no builder or adapter", provider.Name)
		}

		return provider.Builder.Ask(ctx, message)
	}, message)
}

// Stream executes a streaming request using MultiProvider logic
func (mp *MultiProvider) Stream(ctx context.Context, message string) (string, error) {
	return mp.executeWithFallback(ctx, func(provider *ProviderConfig) (string, error) {
		// Use adapter if available, otherwise use builder
		if provider.Adapter != nil {
			req := &CompletionRequest{
				Model:    provider.Model,
				Messages: []Message{{Role: "user", Content: message}},
			}
			resp, err := provider.Adapter.Stream(ctx, req, nil)
			if err != nil {
				return "", err
			}
			return resp.Content, nil
		}

		if provider.Builder == nil {
			return "", fmt.Errorf("provider %s has no builder or adapter", provider.Name)
		}

		return provider.Builder.Stream(ctx, message)
	}, message)
}

// executeWithFallback executes a function with fallback handling
func (mp *MultiProvider) executeWithFallback(ctx context.Context, fn func(*ProviderConfig) (string, error), message string) (string, error) {
	// Select provider based on strategy
	provider, err := mp.selector.SelectProvider(mp.providers, mp.config.SelectionStrategy)
	if err != nil {
		return "", fmt.Errorf("failed to select provider: %w", err)
	}

	// Check circuit breaker
	cb := mp.circuitBreaker[provider.Name]
	if cb != nil && cb.IsOpen() {
		// Circuit breaker is open, try next provider
		nextProvider, nextErr := mp.selector.SelectNextProvider(mp.providers, provider, mp.config.SelectionStrategy)
		if nextErr != nil {
			return "", fmt.Errorf("all providers are unavailable: %w", nextErr)
		}
		provider = nextProvider
	}

	// Execute request with fallback handling
	response, err := mp.fallback.ExecuteWithFallback(ctx, provider, mp.providers, fn, message)
	return response, err
}

// GetMetrics returns metrics for all providers
func (mp *MultiProvider) GetMetrics() map[string]*ProviderMetrics {
	return mp.metrics.GetAllMetrics()
}

// GetProviderStatus returns the current status of all providers
func (mp *MultiProvider) GetProviderStatus() map[string]ProviderStatus {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	status := make(map[string]ProviderStatus)
	for _, provider := range mp.providers {
		status[provider.Name] = provider.Status
	}

	return status
}

// AddProvider dynamically adds a new provider
func (mp *MultiProvider) AddProvider(config ProviderConfig) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Validate provider config
	if err := mp.validateProviderConfig(config); err != nil {
		return fmt.Errorf("invalid provider configuration: %w", err)
	}

	// Create builder if needed
	if config.Type != "adapter" && config.Adapter == nil {
		builder, err := mp.createBuilderForProvider(&config)
		if err != nil {
			return fmt.Errorf("failed to create builder for provider %s: %w", config.Name, err)
		}
		config.Builder = builder
	}

	// Add to providers list
	mp.providers = append(mp.providers, &config)

	// Initialize circuit breaker
	mp.circuitBreaker[config.Name] = NewCircuitBreaker(
		config.Name,
		mp.config.CircuitBreakerThreshold,
		mp.config.CircuitBreakerTimeout,
	)

	return nil
}

// RemoveProvider removes a provider from the MultiProvider
func (mp *MultiProvider) RemoveProvider(name string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for i, provider := range mp.providers {
		if provider.Name == name {
			// Remove from list
			mp.providers = append(mp.providers[:i], mp.providers[i+1:]...)

			// Remove circuit breaker
			delete(mp.circuitBreaker, name)

			return nil
		}
	}

	return fmt.Errorf("provider %s not found", name)
}

// DisableProvider temporarily disables a provider
func (mp *MultiProvider) DisableProvider(name string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for _, provider := range mp.providers {
		if provider.Name == name {
			provider.Status = ProviderStatusDisabled
			return nil
		}
	}

	return fmt.Errorf("provider %s not found", name)
}

// EnableProvider re-enables a disabled provider
func (mp *MultiProvider) EnableProvider(name string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for _, provider := range mp.providers {
		if provider.Name == name {
			provider.Status = ProviderStatusUnknown // Will be determined by health check
			return nil
		}
	}

	return fmt.Errorf("provider %s not found", name)
}

// Shutdown gracefully shuts down the MultiProvider
func (mp *MultiProvider) Shutdown(ctx context.Context) error {
	close(mp.shutdown)

	// Wait for background services to finish
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		// Give background services time to finish
	}

	return nil
}

// validateProviderConfig validates a provider configuration
func (mp *MultiProvider) validateProviderConfig(config ProviderConfig) error {
	if config.Name == "" {
		return fmt.Errorf("provider name is required")
	}

	if config.Model == "" {
		return fmt.Errorf("provider model is required")
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second // Default timeout
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = 3 // Default retries
	}

	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second // Default retry delay
	}

	if config.Weight == 0 {
		config.Weight = 1.0 // Default weight
	}

	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 10 // Default concurrency
	}

	if config.RequestsPerMinute == 0 {
		config.RequestsPerMinute = 60 // Default rate limit
	}

	return nil
}