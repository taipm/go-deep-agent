// Package agent implements load balancing functionality for MultiProvider
// This file contains the load balancing system with various algorithms
package agent

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancer manages load balancing across multiple providers
type LoadBalancer struct {
	config      *MultiProviderConfig
	logger      Logger

	// Load balancing state
	activeRequests  map[string]*int64 // provider -> active request count
	requestHistory  map[string][]time.Duration // provider -> response time history
	mu              sync.RWMutex

	// Sticky session support
	sessionMap      map[string]string // session_id -> provider_name
	sessionMu       sync.RWMutex
	sessionTimeout  time.Duration
}

// LoadBalancingMetrics contains metrics for load balancing decisions
type LoadBalancingMetrics struct {
	Provider          string        `json:"provider"`
	ActiveRequests    int64         `json:"active_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	ErrorRate         float64       `json:"error_rate"`
	LoadScore         float64       `json:"load_score"`
	Capacity          int           `json:"capacity"`
	Utilization       float64       `json:"utilization"`
}

// NewLoadBalancer creates a new load balancer instance
func NewLoadBalancer(config *MultiProviderConfig) *LoadBalancer {
	return &LoadBalancer{
		config:          config,
		logger:          NewStdLogger(LogLevelInfo),
		activeRequests:  make(map[string]*int64),
		requestHistory:  make(map[string][]time.Duration),
		sessionMap:      make(map[string]string),
		sessionTimeout:  30 * time.Minute, // Default session timeout
	}
}

// SelectProviderForRequest selects the best provider for a specific request
func (lb *LoadBalancer) SelectProviderForRequest(providers []*ProviderConfig, sessionID string) (*ProviderConfig, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Check sticky sessions if enabled
	if lb.config.StickySessions && sessionID != "" {
		if provider := lb.getStickyProvider(providers, sessionID); provider != nil {
			lb.logger.Debug(nil, "Selected provider using sticky session",
				F("session_id", sessionID),
				F("provider", provider.Name))
			return provider, nil
		}
	}

	// Filter available providers
	availableProviders := lb.getAvailableProviders(providers)
	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no available providers for load balancing")
	}

	// Select provider based on load balancing strategy
	selected, err := lb.selectBasedOnLoad(availableProviders)
	if err != nil {
		return nil, fmt.Errorf("load balancing failed: %w", err)
	}

	// Update sticky session if enabled
	if lb.config.StickySessions && sessionID != "" {
		lb.setStickyProvider(sessionID, selected.Name)
	}

	lb.logger.Debug(nil, "Selected provider for request",
		F("provider", selected.Name),
		F("strategy", "load_based"),
		F("session_id", sessionID))

	return selected, nil
}

// getStickyProvider returns the provider associated with a session
func (lb *LoadBalancer) getStickyProvider(providers []*ProviderConfig, sessionID string) *ProviderConfig {
	lb.sessionMu.RLock()
	providerName, exists := lb.sessionMap[sessionID]
	lb.sessionMu.RUnlock()

	if !exists {
		return nil
	}

	// Find the provider in the current provider list
	for _, provider := range providers {
		if provider.Name == providerName {
			// Check if provider is still available
			if provider.Status == ProviderStatusHealthy || provider.Status == ProviderStatusDegraded {
				return provider
			}
		}
	}

	// Provider not found or not available, remove from session map
	lb.sessionMu.Lock()
	delete(lb.sessionMap, sessionID)
	lb.sessionMu.Unlock()

	return nil
}

// setStickyProvider associates a session with a provider
func (lb *LoadBalancer) setStickyProvider(sessionID, providerName string) {
	lb.sessionMu.Lock()
	defer lb.sessionMu.Unlock()

	lb.sessionMap[sessionID] = providerName

	// Clean up old sessions periodically (simple cleanup)
	if len(lb.sessionMap) > 1000 {
		lb.cleanupOldSessions()
	}
}

// cleanupOldSessions removes expired session mappings
func (lb *LoadBalancer) cleanupOldSessions() {
	// In a real implementation, you would track session creation time
	// For now, this is a simple cleanup when the map gets too large
	if len(lb.sessionMap) > 500 {
		// Clear half the entries (simple LRU-like cleanup)
		count := 0
		for sessionID := range lb.sessionMap {
			delete(lb.sessionMap, sessionID)
			count++
			if count >= len(lb.sessionMap)/2 {
				break
			}
		}
	}
}

// selectBasedOnLoad selects provider based on current load
func (lb *LoadBalancer) selectBasedOnLoad(providers []*ProviderConfig) (*ProviderConfig, error) {
	if !lb.config.EnableLoadBalancing {
		// If load balancing is disabled, return the first healthy provider
		for _, provider := range providers {
			if provider.Status == ProviderStatusHealthy {
				return provider, nil
			}
		}
		return providers[0], nil
	}

	// Calculate load scores for each provider
	type providerScore struct {
		provider *ProviderConfig
		score    float64
		metrics  *LoadBalancingMetrics
	}

	var scoredProviders []providerScore

	for _, provider := range providers {
		metrics := lb.calculateLoadMetrics(provider)
		score := lb.calculateLoadScore(metrics)

		scoredProviders = append(scoredProviders, providerScore{
			provider: provider,
			score:    score,
			metrics:  metrics,
		})
	}

	// Select provider with best score (lowest load)
	var bestProvider *ProviderConfig
	var bestScore float64 = -1

	for _, scored := range scoredProviders {
		if scored.score > bestScore && (scored.provider.Status == ProviderStatusHealthy || scored.provider.Status == ProviderStatusDegraded) {
			bestScore = scored.score
			bestProvider = scored.provider
		}
	}

	if bestProvider == nil {
		return nil, fmt.Errorf("no suitable providers found")
	}

	lb.logger.Debug(nil, "Load balancing decision",
		F("selected_provider", bestProvider.Name),
		F("load_score", bestScore),
		F("total_providers", len(scoredProviders)))

	return bestProvider, nil
}

// calculateLoadMetrics calculates detailed load metrics for a provider
func (lb *LoadBalancer) calculateLoadMetrics(provider *ProviderConfig) *LoadBalancingMetrics {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	metrics := &LoadBalancingMetrics{
		Provider: provider.Name,
		Capacity: provider.MaxConcurrency,
	}

	// Get active requests
	if activeReqs, exists := lb.activeRequests[provider.Name]; exists {
		metrics.ActiveRequests = atomic.LoadInt64(activeReqs)
	}

	// Calculate average response time
	if history, exists := lb.requestHistory[provider.Name]; exists && len(history) > 0 {
		var total time.Duration
		for _, rt := range history {
			total += rt
		}
		metrics.AverageResponseTime = total / time.Duration(len(history))
	}

	// Calculate utilization
	if metrics.Capacity > 0 {
		metrics.Utilization = float64(metrics.ActiveRequests) / float64(metrics.Capacity)
	}

	// Calculate error rate
	totalRequests := provider.SuccessCount + provider.ErrorCount
	if totalRequests > 0 {
		metrics.ErrorRate = float64(provider.ErrorCount) / float64(totalRequests)
	}

	// Estimate requests per second (simplified)
	// In a real implementation, you would track this more accurately
	metrics.RequestsPerSecond = float64(totalRequests) / time.Since(time.Now().Add(-time.Hour)).Seconds()

	// Calculate load score (lower is better)
	metrics.LoadScore = lb.calculateLoadScore(metrics)

	return metrics
}

// calculateLoadScore calculates a load score for a provider
func (lb *LoadBalancer) calculateLoadScore(metrics *LoadBalancingMetrics) float64 {
	if metrics.Capacity == 0 {
		return 0.0
	}

	// Factors that contribute to load score:
	// 1. Current utilization (0-1, lower is better)
	// 2. Error rate (0-1, lower is better)
	// 3. Response time (normalized, lower is better)
	// 4. Available capacity (higher is better)

	// Weight factors (can be tuned)
	const (
		utilizationWeight = 0.4
		errorWeight        = 0.3
		responseTimeWeight = 0.2
		capacityWeight     = 0.1
	)

	// Normalize response time (0-1 scale, assuming 5 seconds as "slow")
	normalizedResponseTime := float64(metrics.AverageResponseTime) / float64(5*time.Second)
	if normalizedResponseTime > 1.0 {
		normalizedResponseTime = 1.0
	}

	// Available capacity ratio
	availableCapacity := float64(metrics.Capacity-int(metrics.ActiveRequests)) / float64(metrics.Capacity)
	if availableCapacity < 0 {
		availableCapacity = 0
	}

	// Calculate weighted score (higher is better)
	score := (1.0-metrics.Utilization)*utilizationWeight +
		(1.0-metrics.ErrorRate)*errorWeight +
		(1.0-normalizedResponseTime)*responseTimeWeight +
		availableCapacity*capacityWeight

	return score
}

// getAvailableProviders filters providers based on current load
func (lb *LoadBalancer) getAvailableProviders(providers []*ProviderConfig) []*ProviderConfig {
	var available []*ProviderConfig

	for _, provider := range providers {
		// Skip disabled providers
		if provider.Status == ProviderStatusDisabled {
			continue
		}

		// Check if provider has capacity
		if lb.hasCapacity(provider) {
			available = append(available, provider)
		}
	}

	// If no providers with capacity, include degraded providers
	if len(available) == 0 {
		for _, provider := range providers {
			if provider.Status == ProviderStatusDegraded && lb.hasCapacity(provider) {
				available = append(available, provider)
			}
		}
	}

	return available
}

// hasCapacity checks if a provider has available capacity
func (lb *LoadBalancer) hasCapacity(provider *ProviderConfig) bool {
	if provider.MaxConcurrency <= 0 {
		// Unlimited capacity
		return true
	}

	lb.mu.RLock()
	activeReqs := int64(0)
	if active, exists := lb.activeRequests[provider.Name]; exists {
		activeReqs = atomic.LoadInt64(active)
	}
	lb.mu.RUnlock()

	return activeReqs < int64(provider.MaxConcurrency)
}

// StartRequest marks the start of a request for load tracking
func (lb *LoadBalancer) StartRequest(providerName string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if _, exists := lb.activeRequests[providerName]; !exists {
		lb.activeRequests[providerName] = new(int64)
	}

	atomic.AddInt64(lb.activeRequests[providerName], 1)
}

// EndRequest marks the end of a request and records response time
func (lb *LoadBalancer) EndRequest(providerName string, responseTime time.Duration, success bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Decrement active requests
	if active, exists := lb.activeRequests[providerName]; exists {
		atomic.AddInt64(active, -1)
	}

	// Record response time in history
	if _, exists := lb.requestHistory[providerName]; !exists {
		lb.requestHistory[providerName] = make([]time.Duration, 0, 100)
	}

	history := lb.requestHistory[providerName]
	history = append(history, responseTime)

	// Keep only recent history (last 100 requests)
	if len(history) > 100 {
		history = history[len(history)-100:]
	}
	lb.requestHistory[providerName] = history
}

// GetLoadMetrics returns current load metrics for all providers
func (lb *LoadBalancer) GetLoadMetrics(providers []*ProviderConfig) map[string]*LoadBalancingMetrics {
	metrics := make(map[string]*LoadBalancingMetrics)

	for _, provider := range providers {
		metrics[provider.Name] = lb.calculateLoadMetrics(provider)
	}

	return metrics
}

// IsOverloaded checks if a provider is currently overloaded
func (lb *LoadBalancer) IsOverloaded(provider *ProviderConfig) bool {
	if provider.MaxConcurrency <= 0 {
		return false // No limit
	}

	metrics := lb.calculateLoadMetrics(provider)

	// Consider overloaded if utilization > 90% or error rate > 50%
	return metrics.Utilization > 0.9 || metrics.ErrorRate > 0.5
}

// GetRecommendedConcurrency returns recommended concurrency settings for providers
func (lb *LoadBalancer) GetRecommendedConcurrency(providers []*ProviderConfig) map[string]int {
	recommendations := make(map[string]int)

	for _, provider := range providers {
		metrics := lb.calculateLoadMetrics(provider)

		// Base recommendation on current load and performance
		if metrics.AverageResponseTime > 5*time.Second {
			// Slow response times, recommend reducing concurrency
			recommendations[provider.Name] = int(float64(provider.MaxConcurrency) * 0.7)
		} else if metrics.Utilization < 0.5 && metrics.ErrorRate < 0.1 {
			// Low utilization and good performance, can increase concurrency
			recommendations[provider.Name] = int(float64(provider.MaxConcurrency) * 1.2)
		} else {
			// Keep current concurrency
			recommendations[provider.Name] = provider.MaxConcurrency
		}

		// Ensure minimum concurrency
		if recommendations[provider.Name] < 1 {
			recommendations[provider.Name] = 1
		}
	}

	return recommendations
}

// ClearSession clears a sticky session mapping
func (lb *LoadBalancer) ClearSession(sessionID string) {
	lb.sessionMu.Lock()
	defer lb.sessionMu.Unlock()
	delete(lb.sessionMap, sessionID)
}

// ClearAllSessions clears all sticky session mappings
func (lb *LoadBalancer) ClearAllSessions() {
	lb.sessionMu.Lock()
	defer lb.sessionMu.Unlock()
	lb.sessionMap = make(map[string]string)
}

// GetSessionCount returns the number of active sticky sessions
func (lb *LoadBalancer) GetSessionCount() int {
	lb.sessionMu.RLock()
	defer lb.sessionMu.RUnlock()
	return len(lb.sessionMap)
}

// Reset resets all load balancing state
func (lb *LoadBalancer) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.activeRequests = make(map[string]*int64)
	lb.requestHistory = make(map[string][]time.Duration)

	lb.sessionMu.Lock()
	lb.sessionMap = make(map[string]string)
	lb.sessionMu.Unlock()
}