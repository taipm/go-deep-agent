// Package agent implements provider selection functionality for MultiProvider
// This file contains the provider selection system with various strategies
package agent

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// ProviderSelector implements different strategies for selecting providers
type ProviderSelector struct {
	config      *MultiProviderConfig
	logger      Logger

	// Selection state
	currentIndex int
	mu           sync.Mutex
}

// NewProviderSelector creates a new provider selector instance
func NewProviderSelector(config *MultiProviderConfig) *ProviderSelector {
	rand.Seed(time.Now().UnixNano())
	return &ProviderSelector{
		config: config,
		logger: NewStdLogger(LogLevelInfo),
	}
}

// SelectProvider selects a provider based on the configured strategy
func (ps *ProviderSelector) SelectProvider(providers []*ProviderConfig, strategy SelectionStrategy) (*ProviderConfig, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Filter out disabled providers
	availableProviders := ps.getAvailableProviders(providers)
	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no available providers")
	}

	switch strategy {
	case StrategyRoundRobin:
		return ps.selectRoundRobin(availableProviders)
	case StrategyWeightedRoundRobin:
		return ps.selectWeightedRoundRobin(availableProviders)
	case StrategyLeastConnections:
		return ps.selectLeastConnections(availableProviders)
	case StrategyFastestResponse:
		return ps.selectFastestResponse(availableProviders)
	case StrategyRandom:
		return ps.selectRandom(availableProviders)
	case StrategyPriority:
		return ps.selectPriority(availableProviders)
	default:
		return ps.selectRandom(availableProviders)
	}
}

// SelectNextProvider selects the next provider in line for fallback scenarios
func (ps *ProviderSelector) SelectNextProvider(providers []*ProviderConfig, currentProvider *ProviderConfig, strategy SelectionStrategy) (*ProviderConfig, error) {
	availableProviders := ps.getAvailableProviders(providers)
	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no available providers")
	}

	// Find current provider index
	currentIndex := -1
	for i, provider := range availableProviders {
		if provider.Name == currentProvider.Name {
			currentIndex = i
			break
		}
	}

	// Get next provider in rotation
	var nextProvider *ProviderConfig
	switch strategy {
	case StrategyRoundRobin, StrategyWeightedRoundRobin, StrategyPriority:
		if currentIndex >= 0 && currentIndex < len(availableProviders)-1 {
			nextProvider = availableProviders[currentIndex+1]
		} else {
			// Wrap around to first provider
			nextProvider = availableProviders[0]
		}
	case StrategyLeastConnections, StrategyFastestResponse:
		// For these strategies, just re-select using the same logic
		return ps.SelectProvider(availableProviders, strategy)
	case StrategyRandom:
		// Select a random provider that's not the current one
		var candidates []*ProviderConfig
		for _, provider := range availableProviders {
			if provider.Name != currentProvider.Name {
				candidates = append(candidates, provider)
			}
		}
		if len(candidates) > 0 {
			return ps.selectRandom(candidates)
		}
		return nil, fmt.Errorf("no alternative providers available")
	default:
		return ps.selectRandom(availableProviders)
	}

	return nextProvider, nil
}

// getAvailableProviders filters out disabled and unhealthy providers
func (ps *ProviderSelector) getAvailableProviders(providers []*ProviderConfig) []*ProviderConfig {
	var available []*ProviderConfig

	for _, provider := range providers {
		// Skip disabled providers
		if provider.Status == ProviderStatusDisabled {
			continue
		}

		// Skip unhealthy providers (unless all are unhealthy)
		if provider.Status == ProviderStatusUnhealthy {
			continue
		}

		available = append(available, provider)
	}

	// If no healthy providers, include degraded providers
	if len(available) == 0 {
		for _, provider := range providers {
			if provider.Status == ProviderStatusDegraded && provider.Status != ProviderStatusDisabled {
				available = append(available, provider)
			}
		}
	}

	// If still no providers, include all except disabled (as last resort)
	if len(available) == 0 {
		for _, provider := range providers {
			if provider.Status != ProviderStatusDisabled {
				available = append(available, provider)
			}
		}
	}

	return available
}

// selectRoundRobin implements round-robin selection
func (ps *ProviderSelector) selectRoundRobin(providers []*ProviderConfig) (*ProviderConfig, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.currentIndex >= len(providers) {
		ps.currentIndex = 0
	}

	selected := providers[ps.currentIndex]
	ps.currentIndex++

	ps.logger.Debug(nil, "Selected provider using round-robin",
		F("provider", selected.Name),
		F("index", ps.currentIndex-1),
		F("total_providers", len(providers)))

	return selected, nil
}

// selectWeightedRoundRobin implements weighted round-robin selection
func (ps *ProviderSelector) selectWeightedRoundRobin(providers []*ProviderConfig) (*ProviderConfig, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Calculate total weight
	totalWeight := 0.0
	for _, provider := range providers {
		totalWeight += provider.Weight
	}

	if totalWeight <= 0 {
		// Fallback to simple round-robin if weights are invalid
		return ps.selectRoundRobin(providers)
	}

	// Use weighted selection
	ps.currentIndex++
	runningWeight := 0.0
	targetWeight := float64(ps.currentIndex%int(totalWeight)) + 0.5

	for _, provider := range providers {
		runningWeight += provider.Weight
		if runningWeight >= targetWeight {
			ps.logger.Debug(nil, "Selected provider using weighted round-robin",
				F("provider", provider.Name),
				F("weight", provider.Weight),
				F("target_weight", targetWeight))
			return provider, nil
		}
	}

	// Fallback to first provider
	return providers[0], nil
}

// selectLeastConnections implements least connections selection
func (ps *ProviderSelector) selectLeastConnections(providers []*ProviderConfig) (*ProviderConfig, error) {
	// For this implementation, we'll use error_count as a proxy for connection load
	// In a real implementation, you would track actual concurrent connections

	var selected *ProviderConfig
	minConnections := int(^uint(0) >> 1) // Max int

	for _, provider := range providers {
		connections := provider.ErrorCount // Using error_count as a simple metric
		if connections < minConnections {
			minConnections = connections
			selected = provider
		}
	}

	if selected == nil {
		selected = providers[0] // Fallback
	}

	ps.logger.Debug(nil, "Selected provider using least connections",
		F("provider", selected.Name),
		F("connections", minConnections))

	return selected, nil
}

// selectFastestResponse implements fastest response time selection
func (ps *ProviderSelector) selectFastestResponse(providers []*ProviderConfig) (*ProviderConfig, error) {
	var selected *ProviderConfig
	fastestResponse := time.Hour // Initialize with very large value

	for _, provider := range providers {
		responseTime := provider.ResponseTime
		if responseTime == 0 {
			// If no response time data, give it a moderate priority
			responseTime = 500 * time.Millisecond
		}

		if responseTime < fastestResponse {
			fastestResponse = responseTime
			selected = provider
		}
	}

	if selected == nil {
		selected = providers[0] // Fallback
	}

	ps.logger.Debug(nil, "Selected provider using fastest response",
		F("provider", selected.Name),
		F("response_time", fastestResponse))

	return selected, nil
}

// selectRandom implements random selection
func (ps *ProviderSelector) selectRandom(providers []*ProviderConfig) (*ProviderConfig, error) {
	selected := providers[rand.Intn(len(providers))]

	ps.logger.Debug(nil, "Selected provider using random",
		F("provider", selected.Name),
		F("total_providers", len(providers)))

	return selected, nil
}

// selectPriority implements priority-based selection (by order in list)
func (ps *ProviderSelector) selectPriority(providers []*ProviderConfig) (*ProviderConfig, error) {
	// Sort by weight (higher weight = higher priority) and then by name for stability
	sortedProviders := make([]*ProviderConfig, len(providers))
	copy(sortedProviders, providers)

	sort.Slice(sortedProviders, func(i, j int) bool {
		if sortedProviders[i].Weight != sortedProviders[j].Weight {
			return sortedProviders[i].Weight > sortedProviders[j].Weight
		}
		return sortedProviders[i].Name < sortedProviders[j].Name
	})

	selected := sortedProviders[0]

	ps.logger.Debug(nil, "Selected provider using priority",
		F("provider", selected.Name),
		F("weight", selected.Weight))

	return selected, nil
}

// GetProviderRanking returns providers ranked by the specified strategy
func (ps *ProviderSelector) GetProviderRanking(providers []*ProviderConfig, strategy SelectionStrategy) []ProviderRanking {
	availableProviders := ps.getAvailableProviders(providers)
	rankings := make([]ProviderRanking, len(availableProviders))

	switch strategy {
	case StrategyFastestResponse:
		// Sort by response time
		sort.Slice(availableProviders, func(i, j int) bool {
			return availableProviders[i].ResponseTime < availableProviders[j].ResponseTime
		})
	case StrategyLeastConnections:
		// Sort by error count (fewer errors = better)
		sort.Slice(availableProviders, func(i, j int) bool {
			return availableProviders[i].ErrorCount < availableProviders[j].ErrorCount
		})
	case StrategyPriority, StrategyWeightedRoundRobin:
		// Sort by weight (higher weight = higher priority)
		sort.Slice(availableProviders, func(i, j int) bool {
			if availableProviders[i].Weight != availableProviders[j].Weight {
				return availableProviders[i].Weight > availableProviders[j].Weight
			}
			return availableProviders[i].Name < availableProviders[j].Name
		})
	}

	// Create rankings
	for i, provider := range availableProviders {
		rankings[i] = ProviderRanking{
			Provider:      provider.Name,
			Rank:          i + 1,
			Score:         ps.calculateProviderScore(provider, strategy),
			Status:        provider.Status,
			ResponseTime:  provider.ResponseTime,
			SuccessCount:  provider.SuccessCount,
			ErrorCount:    provider.ErrorCount,
			Weight:        provider.Weight,
		}
	}

	return rankings
}

// ProviderRanking contains ranking information for a provider
type ProviderRanking struct {
	Provider      string        `json:"provider"`
	Rank          int           `json:"rank"`
	Score         float64       `json:"score"`
	Status        ProviderStatus `json:"status"`
	ResponseTime  time.Duration `json:"response_time"`
	SuccessCount  int           `json:"success_count"`
	ErrorCount    int           `json:"error_count"`
	Weight        float64       `json:"weight"`
}

// calculateProviderScore calculates a score for a provider based on the strategy
func (ps *ProviderSelector) calculateProviderScore(provider *ProviderConfig, strategy SelectionStrategy) float64 {
	switch strategy {
	case StrategyFastestResponse:
		// Lower response time = higher score
		if provider.ResponseTime > 0 {
			return 1.0 / float64(provider.ResponseTime.Milliseconds())
		}
		return 0.5 // Default score for unknown response times

	case StrategyLeastConnections:
		// Fewer errors = higher score
		totalRequests := provider.SuccessCount + provider.ErrorCount
		if totalRequests > 0 {
			successRate := float64(provider.SuccessCount) / float64(totalRequests)
			return successRate
		}
		return 0.5 // Default score for unknown success rates

	case StrategyPriority, StrategyWeightedRoundRobin:
		// Use weight as score
		return provider.Weight

	default:
		// Default score based on weight
		return provider.Weight
	}
}

// UpdateProviderMetrics updates the metrics for a provider after a request
func (ps *ProviderSelector) UpdateProviderMetrics(providerName string, responseTime time.Duration, success bool) {
	// This would typically be called by the MultiProvider after each request
	// The actual implementation would update metrics in the provider config
	ps.logger.Debug(nil, "Updating provider metrics",
		F("provider", providerName),
		F("response_time", responseTime),
		F("success", success))
}

// Reset resets the selector state
func (ps *ProviderSelector) Reset() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.currentIndex = 0
}