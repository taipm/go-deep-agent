// Package agent implements metrics collection for MultiProvider
// This file contains the metrics collection and monitoring system
package agent

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector collects and aggregates metrics for MultiProvider
type MetricsCollector struct {
	config      *MultiProviderConfig
	logger      Logger

	// Metrics storage
	providerMetrics map[string]*ProviderMetrics
	globalMetrics   *GlobalMetrics
	mu              sync.RWMutex

	// Aggregation state
	interval     time.Duration
	shutdown     chan struct{}
	running      int32
}

// GlobalMetrics contains global metrics across all providers
type GlobalMetrics struct {
	// Request metrics
	TotalRequests         int64     `json:"total_requests"`
	SuccessfulRequests    int64     `json:"successful_requests"`
	FailedRequests        int64     `json:"failed_requests"`
	SuccessRate           float64   `json:"success_rate"`

	// Performance metrics
	AverageResponseTime   time.Duration `json:"average_response_time"`
	MinResponseTime       time.Duration `json:"min_response_time"`
	MaxResponseTime       time.Duration `json:"max_response_time"`
	P95ResponseTime       time.Duration `json:"p95_response_time"`
	P99ResponseTime       time.Duration `json:"p99_response_time"`

	// Availability metrics
	TotalUptime           time.Duration `json:"total_uptime"`
	TotalDowntime         time.Duration `json:"total_downtime"`
	UptimePercentage      float64       `json:"uptime_percentage"`

	// Provider distribution
	ProviderDistribution  map[string]int64 `json:"provider_distribution"`

	// Timestamps
	StartTime             time.Time      `json:"start_time"`
	LastUpdateTime        time.Time      `json:"last_update_time"`

	// Active metrics
	ActiveRequests        int64          `json:"active_requests"`
	ConcurrentConnections int64          `json:"concurrent_connections"`
}

// RequestMetrics contains metrics for a single request
type RequestMetrics struct {
	Provider       string        `json:"provider"`
	RequestType    string        `json:"request_type"` // "ask", "stream"
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	ResponseTime   time.Duration `json:"response_time"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
	TokenUsage     TokenUsage    `json:"token_usage"`
	FallbackUsed   bool          `json:"fallback_used"`
	FallbackCount  int           `json:"fallback_count"`
	SessionID      string        `json:"session_id,omitempty"`
}

// AggregatedMetrics contains aggregated metrics over a time period
type AggregatedMetrics struct {
	Period         time.Duration            `json:"period"`
	StartTime      time.Time                `json:"start_time"`
	EndTime        time.Time                `json:"end_time"`
	ProviderMetrics map[string]*ProviderMetrics `json:"provider_metrics"`
	GlobalMetrics  *GlobalMetrics           `json:"global_metrics"`
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector(config *MultiProviderConfig) *MetricsCollector {
	return &MetricsCollector{
		config:         config,
		logger:         NewStdLogger(LogLevelInfo),
		providerMetrics: make(map[string]*ProviderMetrics),
		globalMetrics: &GlobalMetrics{
			ProviderDistribution: make(map[string]int64),
			StartTime:           time.Now(),
		},
		shutdown: make(chan struct{}),
	}
}

// Start begins the metrics collection process
func (mc *MetricsCollector) Start(providers []*ProviderConfig, interval time.Duration) {
	if !atomic.CompareAndSwapInt32(&mc.running, 0, 1) {
		mc.logger.Info(nil, "Metrics collector already running")
		return
	}

	// Set default interval if zero or negative
	if interval <= 0 {
		interval = 30 * time.Second // Default 30 seconds
	}

	mc.interval = interval
	mc.logger.Info(nil, "Starting metrics collector",
		F("interval", interval),
		F("providers", len(providers)))

	// Initialize metrics for each provider
	for _, provider := range providers {
		mc.providerMetrics[provider.Name] = &ProviderMetrics{
			Status: ProviderStatusUnknown,
		}
	}

	// Start aggregation goroutine
	go mc.aggregationLoop()
}

// Stop stops the metrics collection process
func (mc *MetricsCollector) Stop() {
	if !atomic.CompareAndSwapInt32(&mc.running, 1, 0) {
		return
	}

	mc.logger.Info(nil, "Stopping metrics collector")
	close(mc.shutdown)
}

// RecordRequest records metrics for a completed request
func (mc *MetricsCollector) RecordRequest(metrics *RequestMetrics) {
	if metrics == nil {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Update provider metrics
	providerMetrics := mc.getOrCreateProviderMetrics(metrics.Provider)
	mc.updateProviderMetrics(providerMetrics, metrics)

	// Update global metrics
	mc.updateGlobalMetrics(metrics)

	mc.logger.Debug(nil, "Recorded request metrics",
		F("provider", metrics.Provider),
		F("response_time", metrics.ResponseTime),
		F("success", metrics.Success))
}

// GetMetrics returns current metrics for all providers
func (mc *MetricsCollector) GetAllMetrics() map[string]*ProviderMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := make(map[string]*ProviderMetrics)
	for name, providerMetrics := range mc.providerMetrics {
		metrics[name] = mc.copyProviderMetrics(providerMetrics)
	}

	return metrics
}

// GetProviderMetrics returns metrics for a specific provider
func (mc *MetricsCollector) GetProviderMetrics(providerName string) (*ProviderMetrics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	providerMetrics, exists := mc.providerMetrics[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	return mc.copyProviderMetrics(providerMetrics), nil
}

// GetGlobalMetrics returns global metrics
func (mc *MetricsCollector) GetGlobalMetrics() *GlobalMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy to avoid race conditions
	globalCopy := *mc.globalMetrics
	globalCopy.ProviderDistribution = make(map[string]int64)
	for k, v := range mc.globalMetrics.ProviderDistribution {
		globalCopy.ProviderDistribution[k] = v
	}

	return &globalCopy
}

// GetAggregatedMetrics returns aggregated metrics over a time period
func (mc *MetricsCollector) GetAggregatedMetrics(period time.Duration) *AggregatedMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	endTime := time.Now()
	startTime := endTime.Add(-period)

	// Filter provider metrics by time period
	filteredProviderMetrics := make(map[string]*ProviderMetrics)
	for name, providerMetrics := range mc.providerMetrics {
		// In a real implementation, you would filter by time
		// For now, return current metrics
		filteredProviderMetrics[name] = mc.copyProviderMetrics(providerMetrics)
	}

	return &AggregatedMetrics{
		Period:          period,
		StartTime:       startTime,
		EndTime:         endTime,
		ProviderMetrics: filteredProviderMetrics,
		GlobalMetrics:   mc.copyGlobalMetrics(),
	}
}

// ResetMetrics resets all metrics
func (mc *MetricsCollector) ResetMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()

	// Reset provider metrics
	for name := range mc.providerMetrics {
		mc.providerMetrics[name] = &ProviderMetrics{
			Status: ProviderStatusUnknown,
		}
	}

	// Reset global metrics
	mc.globalMetrics = &GlobalMetrics{
		ProviderDistribution: make(map[string]int64),
		StartTime:           now,
		LastUpdateTime:      now,
	}

	mc.logger.Info(nil, "Metrics reset")
}

// aggregationLoop runs periodic aggregation and cleanup
func (mc *MetricsCollector) aggregationLoop() {
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.performAggregation()
		case <-mc.shutdown:
			return
		}
	}
}

// performAggregation performs periodic aggregation tasks
func (mc *MetricsCollector) performAggregation() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	mc.globalMetrics.LastUpdateTime = now

	// Update uptime metrics
	mc.updateUptimeMetrics()

	// Calculate response time percentiles
	mc.calculateResponseTimePercentiles()

	// Update provider statuses
	mc.updateProviderStatuses()

	mc.logger.Debug(nil, "Metrics aggregation completed")
}

// getOrCreateProviderMetrics gets or creates provider metrics
func (mc *MetricsCollector) getOrCreateProviderMetrics(providerName string) *ProviderMetrics {
	providerMetrics, exists := mc.providerMetrics[providerName]
	if !exists {
		providerMetrics = &ProviderMetrics{
			Status: ProviderStatusUnknown,
		}
		mc.providerMetrics[providerName] = providerMetrics
	}

	return providerMetrics
}

// updateProviderMetrics updates provider-specific metrics
func (mc *MetricsCollector) updateProviderMetrics(providerMetrics *ProviderMetrics, requestMetrics *RequestMetrics) {
	atomic.AddInt64(&providerMetrics.TotalRequests, 1)

	if requestMetrics.Success {
		atomic.AddInt64(&providerMetrics.SuccessfulRequests, 1)
		providerMetrics.LastSuccessTime = requestMetrics.EndTime
	} else {
		atomic.AddInt64(&providerMetrics.FailedRequests, 1)
		providerMetrics.LastFailureTime = requestMetrics.EndTime
		if requestMetrics.Error != "" {
			providerMetrics.LastError = requestMetrics.Error
		}
	}

	// Update response time metrics
	mc.updateResponseTimeMetrics(providerMetrics, requestMetrics.ResponseTime)

	// Update current RPS (requests per second)
	if !requestMetrics.StartTime.IsZero() {
		elapsed := time.Since(requestMetrics.StartTime).Seconds()
		if elapsed > 0 {
			providerMetrics.CurrentRPS = 1.0 / elapsed
		}
	}

	// Update status based on recent performance
	mc.updateProviderStatus(providerMetrics)
}

// updateGlobalMetrics updates global metrics
func (mc *MetricsCollector) updateGlobalMetrics(requestMetrics *RequestMetrics) {
	atomic.AddInt64(&mc.globalMetrics.TotalRequests, 1)

	if requestMetrics.Success {
		atomic.AddInt64(&mc.globalMetrics.SuccessfulRequests, 1)
	} else {
		atomic.AddInt64(&mc.globalMetrics.FailedRequests, 1)
	}

	// Update provider distribution
	count, exists := mc.globalMetrics.ProviderDistribution[requestMetrics.Provider]
	if !exists {
		count = 0
	}
	mc.globalMetrics.ProviderDistribution[requestMetrics.Provider] = count + 1

	// Update global response time metrics
	mc.updateGlobalResponseTimeMetrics(requestMetrics.ResponseTime)

	// Calculate success rate
	total := atomic.LoadInt64(&mc.globalMetrics.TotalRequests)
	successful := atomic.LoadInt64(&mc.globalMetrics.SuccessfulRequests)
	if total > 0 {
		mc.globalMetrics.SuccessRate = float64(successful) / float64(total) * 100.0
	}
}

// updateResponseTimeMetrics updates response time metrics for a provider
func (mc *MetricsCollector) updateResponseTimeMetrics(providerMetrics *ProviderMetrics, responseTime time.Duration) {
	// Update average response time (exponential moving average)
	if providerMetrics.AverageResponseTime == 0 {
		providerMetrics.AverageResponseTime = responseTime
	} else {
		// EMA with alpha = 0.1
		providerMetrics.AverageResponseTime = (providerMetrics.AverageResponseTime*9 + responseTime) / 10
	}

	// Update min/max response times
	if providerMetrics.AverageResponseTime == 0 || responseTime < providerMetrics.AverageResponseTime {
		// This is a simplification - in reality, you'd track actual min/max
	}
}

// updateGlobalResponseTimeMetrics updates global response time metrics
func (mc *MetricsCollector) updateGlobalResponseTimeMetrics(responseTime time.Duration) {
	// Update average response time (exponential moving average)
	if mc.globalMetrics.AverageResponseTime == 0 {
		mc.globalMetrics.AverageResponseTime = responseTime
		mc.globalMetrics.MinResponseTime = responseTime
		mc.globalMetrics.MaxResponseTime = responseTime
	} else {
		// EMA with alpha = 0.1
		mc.globalMetrics.AverageResponseTime = (mc.globalMetrics.AverageResponseTime*9 + responseTime) / 10

		// Update min/max (simplified)
		if responseTime < mc.globalMetrics.MinResponseTime {
			mc.globalMetrics.MinResponseTime = responseTime
		}
		if responseTime > mc.globalMetrics.MaxResponseTime {
			mc.globalMetrics.MaxResponseTime = responseTime
		}
	}
}

// updateProviderStatus updates provider status based on metrics
func (mc *MetricsCollector) updateProviderStatus(providerMetrics *ProviderMetrics) {
	totalRequests := atomic.LoadInt64(&providerMetrics.TotalRequests)
	totalFailures := atomic.LoadInt64(&providerMetrics.FailedRequests)

	if totalRequests == 0 {
		providerMetrics.Status = ProviderStatusUnknown
		return
	}

	failureRate := float64(totalFailures) / float64(totalRequests)

	// Update status based on failure rate and response time
	if failureRate > 0.5 {
		providerMetrics.Status = ProviderStatusUnhealthy
	} else if failureRate > 0.1 || providerMetrics.AverageResponseTime > 10*time.Second {
		providerMetrics.Status = ProviderStatusDegraded
	} else {
		providerMetrics.Status = ProviderStatusHealthy
	}

	// Update uptime percentage
	providerMetrics.UptimePercentage = (1.0 - failureRate) * 100.0
}

// updateUptimeMetrics updates uptime metrics
func (mc *MetricsCollector) updateUptimeMetrics() {
	now := time.Now()
	totalRuntime := now.Sub(mc.globalMetrics.StartTime)

	// This is a simplified calculation
	// In a real implementation, you would track actual downtime periods
	mc.globalMetrics.TotalUptime = totalRuntime
	mc.globalMetrics.UptimePercentage = 100.0 // Assume 100% unless we detect downtime
}

// calculateResponseTimePercentiles calculates response time percentiles
func (mc *MetricsCollector) calculateResponseTimePercentiles() {
	// This is a simplified calculation
	// In a real implementation, you would store response time history and calculate actual percentiles
	if mc.globalMetrics.AverageResponseTime > 0 {
		mc.globalMetrics.P95ResponseTime = time.Duration(float64(mc.globalMetrics.AverageResponseTime) * 1.5)
		mc.globalMetrics.P99ResponseTime = time.Duration(float64(mc.globalMetrics.AverageResponseTime) * 2.0)
	}
}

// updateProviderStatuses updates all provider statuses
func (mc *MetricsCollector) updateProviderStatuses() {
	for _, providerMetrics := range mc.providerMetrics {
		mc.updateProviderStatus(providerMetrics)
	}
}

// copyProviderMetrics creates a deep copy of provider metrics
func (mc *MetricsCollector) copyProviderMetrics(original *ProviderMetrics) *ProviderMetrics {
	if original == nil {
		return nil
	}

	copy := *original
	return &copy
}

// copyGlobalMetrics creates a deep copy of global metrics
func (mc *MetricsCollector) copyGlobalMetrics() *GlobalMetrics {
	if mc.globalMetrics == nil {
		return nil
	}

	copy := *mc.globalMetrics
	copy.ProviderDistribution = make(map[string]int64)
	for k, v := range mc.globalMetrics.ProviderDistribution {
		copy.ProviderDistribution[k] = v
	}

	return &copy
}

// GetMetricsSummary returns a summary of current metrics
func (mc *MetricsCollector) GetMetricsSummary() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	global := mc.globalMetrics
	total := atomic.LoadInt64(&global.TotalRequests)
	successful := atomic.LoadInt64(&global.SuccessfulRequests)
	failed := atomic.LoadInt64(&global.FailedRequests)

	return map[string]interface{}{
		"total_requests":      total,
		"successful_requests": successful,
		"failed_requests":     failed,
		"success_rate":        global.SuccessRate,
		"average_response_time": global.AverageResponseTime.String(),
		"uptime_percentage":   global.UptimePercentage,
		"providers":           len(mc.providerMetrics),
		"healthy_providers":   mc.countHealthyProviders(),
		"uptime":              time.Since(global.StartTime).String(),
	}
}

// countHealthyProviders counts the number of healthy providers
func (mc *MetricsCollector) countHealthyProviders() int {
	count := 0
	for _, providerMetrics := range mc.providerMetrics {
		if providerMetrics.Status == ProviderStatusHealthy {
			count++
		}
	}
	return count
}

// ExportMetrics exports metrics in a format suitable for monitoring systems
func (mc *MetricsCollector) ExportMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	export := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"global":    mc.globalMetrics,
		"providers": mc.providerMetrics,
	}

	return export
}