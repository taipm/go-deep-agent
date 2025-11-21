// Package agent implements fallback mechanisms for MultiProvider
// This file contains the fallback system with circuit breaker and automatic switching
package agent

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// FallbackHandler manages fallback logic and provider switching
type FallbackHandler struct {
	config      *MultiProviderConfig
	logger      Logger

	// Circuit breaker state
	circuitBreakers map[string]*CircuitBreaker
	mu              sync.RWMutex
}

// NewFallbackHandler creates a new fallback handler instance
func NewFallbackHandler(config *MultiProviderConfig) *FallbackHandler {
	return &FallbackHandler{
		config:          config,
		logger:          NewStdLogger(LogLevelInfo),
		circuitBreakers: make(map[string]*CircuitBreaker),
	}
}

// ExecuteWithFallback executes a request with automatic fallback to other providers
func (fh *FallbackHandler) ExecuteWithFallback(ctx context.Context, primaryProvider *ProviderConfig, providers []*ProviderConfig, executeFunc func(*ProviderConfig) (string, error), message string) (string, error) {
	// Filter providers to exclude disabled ones
	availableProviders := fh.getAvailableProviders(providers, primaryProvider)
	if len(availableProviders) == 0 {
		return "", fmt.Errorf("no available providers for fallback")
	}

	var lastError error
	var attemptedProviders []string

	// Try each provider in sequence according to the fallback strategy
	for _, provider := range availableProviders {
		// Check circuit breaker
		cb := fh.getCircuitBreaker(provider.Name)
		if cb.IsOpen() {
			fh.logger.Info(ctx, "Provider circuit breaker is open, skipping",
				F("provider", provider.Name),
				F("state", cb.State()))
			continue
		}

		// Attempt execution with this provider
		fh.logger.Debug(ctx, "Attempting request with provider",
			F("provider", provider.Name),
			F("attempted_providers", attemptedProviders))

		result, err := fh.executeWithTimeout(ctx, provider, executeFunc, message)
		attemptedProviders = append(attemptedProviders, provider.Name)

		if err == nil {
			// Success - record success and return result
			cb.RecordSuccess()
			fh.logger.Info(ctx, "Request succeeded with provider",
				F("provider", provider.Name),
				F("attempts", len(attemptedProviders)))

			return result, nil
		}

		// Failure - record failure and continue to next provider
		lastError = err
		cb.RecordFailure()
		fh.logger.Warn(ctx, "Request failed with provider, trying fallback",
			F("provider", provider.Name),
			F("error", err.Error()),
			F("remaining_providers", len(availableProviders)-len(attemptedProviders)))

		// Check if context is cancelled
		if ctx.Err() != nil {
			return "", fmt.Errorf("request cancelled: %w", ctx.Err())
		}
	}

	// All providers failed
	return "", fmt.Errorf("all providers failed. Last error: %w. Attempted providers: %v", lastError, attemptedProviders)
}

// getAvailableProviders returns providers sorted by fallback priority
func (fh *FallbackHandler) getAvailableProviders(providers []*ProviderConfig, excludeProvider *ProviderConfig) []*ProviderConfig {
	var available []*ProviderConfig

	for _, provider := range providers {
		// Skip excluded provider (usually the one that just failed)
		if provider.Name == excludeProvider.Name {
			continue
		}

		// Skip disabled providers
		if provider.Status == ProviderStatusDisabled {
			continue
		}

		// For fallback strategies, we might want to prioritize healthy providers
		available = append(available, provider)
	}

	// Sort providers based on fallback strategy
	switch fh.config.FallbackStrategy {
	case FallbackStrategyFailFast:
		// For fail-fast, keep original order
		return available

	case FallbackStrategyRetryWithBackoff, FallbackStrategyCircuitBreaker:
		// Prioritize healthy providers
		fh.sortProvidersByHealth(available)

	case FallbackStrategyGracefulDegradation:
		// Sort by weight and health
		fh.sortProvidersByWeightAndHealth(available)

	default:
		return available
	}

	return available
}

// sortProvidersByHealth sorts providers by health status (healthy first)
func (fh *FallbackHandler) sortProvidersByHealth(providers []*ProviderConfig) {
	// Simple bubble sort by health status
	for i := 0; i < len(providers)-1; i++ {
		for j := 0; j < len(providers)-i-1; j++ {
			if fh.getProviderHealthScore(providers[j]) < fh.getProviderHealthScore(providers[j+1]) {
				providers[j], providers[j+1] = providers[j+1], providers[j]
			}
		}
	}
}

// sortProvidersByWeightAndHealth sorts providers by weight and health
func (fh *FallbackHandler) sortProvidersByWeightAndHealth(providers []*ProviderConfig) {
	for i := 0; i < len(providers)-1; i++ {
		for j := 0; j < len(providers)-i-1; j++ {
			score1 := fh.getProviderWeightScore(providers[j])
			score2 := fh.getProviderWeightScore(providers[j+1])
			if score1 < score2 {
				providers[j], providers[j+1] = providers[j+1], providers[j]
			}
		}
	}
}

// getProviderHealthScore returns a score based on provider health
func (fh *FallbackHandler) getProviderHealthScore(provider *ProviderConfig) float64 {
	switch provider.Status {
	case ProviderStatusHealthy:
		return 3.0
	case ProviderStatusDegraded:
		return 2.0
	case ProviderStatusUnknown:
		return 1.0
	case ProviderStatusUnhealthy:
		return 0.0
	default:
		return 0.0
	}
}

// getProviderWeightScore returns a score based on provider weight and health
func (fh *FallbackHandler) getProviderWeightScore(provider *ProviderConfig) float64 {
	healthScore := fh.getProviderHealthScore(provider)
	return provider.Weight * healthScore
}

// executeWithTimeout executes a function with timeout and backoff logic
func (fh *FallbackHandler) executeWithTimeout(ctx context.Context, provider *ProviderConfig, executeFunc func(*ProviderConfig) (string, error), message string) (string, error) {
	// Apply retry with backoff if configured
	if fh.config.FallbackStrategy == FallbackStrategyRetryWithBackoff {
		return fh.executeWithRetry(ctx, provider, executeFunc, message)
	}

	// Simple execution with timeout
	return executeFunc(provider)
}

// executeWithRetry executes a function with retry logic and exponential backoff
func (fh *FallbackHandler) executeWithRetry(ctx context.Context, provider *ProviderConfig, executeFunc func(*ProviderConfig) (string, error), message string) (string, error) {
	maxRetries := provider.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3 // Default
	}

	baseDelay := provider.RetryDelay
	if baseDelay <= 0 {
		baseDelay = 1 * time.Second // Default
	}

	var lastError error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			if delay > 30*time.Second {
				delay = 30 * time.Second // Cap at 30 seconds
			}

			fh.logger.Debug(ctx, "Retrying request with backoff",
				F("provider", provider.Name),
				F("attempt", attempt),
				F("delay", delay))

			select {
			case <-ctx.Done():
				return "", fmt.Errorf("request cancelled during retry: %w", ctx.Err())
			case <-time.After(delay):
				// Continue with retry
			}
		}

		result, err := executeFunc(provider)
		if err == nil {
			if attempt > 0 {
				fh.logger.Info(ctx, "Request succeeded after retries",
					F("provider", provider.Name),
					F("attempts", attempt+1))
			}
			return result, nil
		}

		lastError = err

		fh.logger.Warn(ctx, "Request attempt failed",
			F("provider", provider.Name),
			F("attempt", attempt+1),
			F("error", err.Error()))

		// Check if we should retry based on error type
		if !fh.shouldRetryError(err) {
			break
		}
	}

	return "", fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastError)
}

// shouldRetryError determines if an error should trigger a retry
func (fh *FallbackHandler) shouldRetryError(err error) bool {
	// Don't retry on certain error types
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Don't retry on authentication errors
	if strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "authentication") || strings.Contains(errStr, "api key") {
		return false
	}

	// Don't retry on validation errors
	if strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid") {
		return false
	}

	// Retry on network errors and timeouts
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") {
		return true
	}

	// Default: retry on other errors
	return true
}

// getCircuitBreaker returns or creates a circuit breaker for a provider
func (fh *FallbackHandler) getCircuitBreaker(providerName string) *CircuitBreaker {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	cb, exists := fh.circuitBreakers[providerName]
	if !exists {
		cb = NewCircuitBreaker(providerName, fh.config.CircuitBreakerThreshold, fh.config.CircuitBreakerTimeout)
		fh.circuitBreakers[providerName] = cb
	}

	return cb
}

// GetCircuitBreakerStatus returns the status of all circuit breakers
func (fh *FallbackHandler) GetCircuitBreakerStatus() map[string]*CircuitBreakerStatus {
	fh.mu.RLock()
	defer fh.mu.RUnlock()

	status := make(map[string]*CircuitBreakerStatus)
	for name, cb := range fh.circuitBreakers {
		status[name] = cb.GetStatus()
	}

	return status
}

// ResetCircuitBreaker resets a specific circuit breaker
func (fh *FallbackHandler) ResetCircuitBreaker(providerName string) error {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	cb, exists := fh.circuitBreakers[providerName]
	if !exists {
		return fmt.Errorf("circuit breaker for provider %s not found", providerName)
	}

	cb.Reset()
	fh.logger.Info(nil, "Circuit breaker reset", F("provider", providerName))

	return nil
}

// GetAllCircuitBreakers returns all circuit breakers
func (fh *FallbackHandler) GetAllCircuitBreakers() map[string]*CircuitBreaker {
	fh.mu.RLock()
	defer fh.mu.RUnlock()

	// Return a copy to avoid race conditions
	circuitBreakers := make(map[string]*CircuitBreaker)
	for name, cb := range fh.circuitBreakers {
		circuitBreakers[name] = cb
	}

	return circuitBreakers
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	threshold      int
	timeout        time.Duration

	// State
	state          CircuitBreakerState
	failureCount   int64
	lastFailureTime int64
	mu             sync.RWMutex

	// Metrics
	requests       int64
	successes      int64
	failures       int64
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreakerStatus contains the status of a circuit breaker
type CircuitBreakerStatus struct {
	Name            string              `json:"name"`
	State           CircuitBreakerState `json:"state"`
	FailureCount    int64               `json:"failure_count"`
	Threshold       int                 `json:"threshold"`
	LastFailureTime int64               `json:"last_failure_time"`
	Timeout         time.Duration       `json:"timeout"`
	Requests        int64               `json:"requests"`
	Successes       int64               `json:"successes"`
	Failures        int64               `json:"failures"`
	SuccessRate     float64             `json:"success_rate"`
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:      name,
		threshold: threshold,
		timeout:   timeout,
		state:     CircuitBreakerClosed,
	}
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state == CircuitBreakerOpen
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.AddInt64(&cb.requests, 1)
	atomic.AddInt64(&cb.successes, 1)

	// Reset failure count on success
	cb.failureCount = 0

	// If in half-open state, move to closed on first success
	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerClosed
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.AddInt64(&cb.requests, 1)
	atomic.AddInt64(&cb.failures, 1)

	cb.failureCount++
	cb.lastFailureTime = time.Now().Unix()

	// Check if we should open the circuit
	if cb.state == CircuitBreakerClosed && cb.failureCount >= int64(cb.threshold) {
		cb.state = CircuitBreakerOpen
	}
}

// GetStatus returns the current status of the circuit breaker
func (cb *CircuitBreaker) GetStatus() *CircuitBreakerStatus {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	var successRate float64
	if cb.requests > 0 {
		successRate = float64(cb.successes) / float64(cb.requests) * 100.0
	}

	return &CircuitBreakerStatus{
		Name:            cb.name,
		State:           cb.state,
		FailureCount:    cb.failureCount,
		Threshold:       cb.threshold,
		LastFailureTime: cb.lastFailureTime,
		Timeout:         cb.timeout,
		Requests:        cb.requests,
		Successes:       cb.successes,
		Failures:        cb.failures,
		SuccessRate:     successRate,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitBreakerClosed
	cb.failureCount = 0
	cb.lastFailureTime = 0

	// Reset metrics
	cb.requests = 0
	cb.successes = 0
	cb.failures = 0
}

// ShouldAllowRequest determines if a request should be allowed through the circuit breaker
func (cb *CircuitBreaker) ShouldAllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// Check if timeout has elapsed
		if time.Now().Unix()-cb.lastFailureTime > int64(cb.timeout.Seconds()) {
			cb.state = CircuitBreakerHalfOpen
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// String returns a string representation of the circuit breaker state
func (cbs CircuitBreakerState) String() string {
	switch cbs {
	case CircuitBreakerClosed:
		return "closed"
	case CircuitBreakerOpen:
		return "open"
	case CircuitBreakerHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}