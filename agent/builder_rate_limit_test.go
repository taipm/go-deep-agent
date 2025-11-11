package agent

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestBuilderWithRateLimit tests basic rate limiting integration
func TestBuilderWithRateLimit(t *testing.T) {
	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimit(10.0, 5)

	if !builder.rateLimitEnabled {
		t.Error("rate limiting should be enabled")
	}

	if builder.rateLimitConfig.RequestsPerSecond != 10.0 {
		t.Errorf("expected 10.0 req/s, got %f", builder.rateLimitConfig.RequestsPerSecond)
	}

	if builder.rateLimitConfig.BurstSize != 5 {
		t.Errorf("expected burst size 5, got %d", builder.rateLimitConfig.BurstSize)
	}
}

// TestBuilderWithRateLimitConfig tests advanced rate limit configuration
func TestBuilderWithRateLimitConfig(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 20.0,
		BurstSize:         10,
		PerKey:            true,
		KeyTimeout:        10 * time.Minute,
		WaitTimeout:       60 * time.Second,
	}

	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimitConfig(config)

	if !builder.rateLimitEnabled {
		t.Error("rate limiting should be enabled")
	}

	if builder.rateLimitConfig.PerKey != true {
		t.Error("per-key rate limiting should be enabled")
	}

	if builder.rateLimitConfig.KeyTimeout != 10*time.Minute {
		t.Errorf("expected 10m key timeout, got %v", builder.rateLimitConfig.KeyTimeout)
	}
}

// TestBuilderWithRateLimitKey tests per-key rate limiting
func TestBuilderWithRateLimitKey(t *testing.T) {
	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimit(10.0, 5).
		WithRateLimitKey("user-123")

	if builder.rateLimitKey != "user-123" {
		t.Errorf("expected key 'user-123', got '%s'", builder.rateLimitKey)
	}
}

// TestBuilderRateLimitInitialization tests lazy initialization of rate limiter
func TestBuilderRateLimitInitialization(t *testing.T) {
	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimit(10.0, 5)

	if builder.rateLimiter != nil {
		t.Error("rate limiter should not be initialized until first use")
	}

	// Initialize rate limiter
	err := builder.ensureRateLimiter()
	if err != nil {
		t.Fatalf("failed to initialize rate limiter: %v", err)
	}

	if builder.rateLimiter == nil {
		t.Error("rate limiter should be initialized")
	}

	// Second call should not recreate
	limiter1 := builder.rateLimiter
	err = builder.ensureRateLimiter()
	if err != nil {
		t.Fatalf("failed on second initialization: %v", err)
	}
	limiter2 := builder.rateLimiter

	if limiter1 != limiter2 {
		t.Error("rate limiter should not be recreated on second call")
	}
}

// TestBuilderRateLimitAllowBurst tests burst capacity
func TestBuilderRateLimitAllowBurst(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         3,
	}

	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimitConfig(config)

	err := builder.ensureRateLimiter()
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Should allow burst requests
	for i := 0; i < config.BurstSize; i++ {
		if !builder.rateLimiter.Allow("") {
			t.Errorf("request %d should be allowed (within burst)", i)
		}
	}

	// Next request should be denied
	if builder.rateLimiter.Allow("") {
		t.Error("request beyond burst should be denied")
	}

	// Check stats
	stats := builder.rateLimiter.Stats("")
	if stats.Allowed != int64(config.BurstSize) {
		t.Errorf("expected %d allowed, got %d", config.BurstSize, stats.Allowed)
	}
	if stats.Denied != 1 {
		t.Errorf("expected 1 denied, got %d", stats.Denied)
	}
}

// TestBuilderRateLimitConcurrent tests concurrent requests with rate limiting
func TestBuilderRateLimitConcurrent(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 100.0, // High rate for faster test
		BurstSize:         10,
		WaitTimeout:       5 * time.Second,
	}

	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimitConfig(config)

	err := builder.ensureRateLimiter()
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Launch concurrent requests
	numRequests := 20
	var wg sync.WaitGroup
	var successful atomic.Int32
	var failed atomic.Int32

	ctx := context.Background()
	start := time.Now()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := builder.rateLimiter.Wait(ctx, ""); err != nil {
				failed.Add(1)
			} else {
				successful.Add(1)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	if failed.Load() > 0 {
		t.Errorf("expected all requests to succeed, got %d failures", failed.Load())
	}

	if successful.Load() != int32(numRequests) {
		t.Errorf("expected %d successful, got %d", numRequests, successful.Load())
	}

	// Should take some time due to rate limiting
	t.Logf("Completed %d requests in %v", numRequests, duration)

	// Verify stats
	stats := builder.rateLimiter.Stats("")
	if stats.Allowed != int64(numRequests) {
		t.Errorf("expected %d allowed in stats, got %d", numRequests, stats.Allowed)
	}
}

// TestBuilderRateLimitPerKey tests independent per-key rate limits
func TestBuilderRateLimitPerKey(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         3,
		PerKey:            true,
		KeyTimeout:        1 * time.Second,
	}

	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimitConfig(config)

	err := builder.ensureRateLimiter()
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Exhaust key1's burst
	key1 := "user-1"
	for i := 0; i < config.BurstSize; i++ {
		if !builder.rateLimiter.Allow(key1) {
			t.Errorf("key1 request %d should be allowed", i)
		}
	}
	if builder.rateLimiter.Allow(key1) {
		t.Error("key1 should be rate limited")
	}

	// key2 should have independent limit
	key2 := "user-2"
	for i := 0; i < config.BurstSize; i++ {
		if !builder.rateLimiter.Allow(key2) {
			t.Errorf("key2 request %d should be allowed", i)
		}
	}
	if builder.rateLimiter.Allow(key2) {
		t.Error("key2 should be rate limited")
	}

	// Verify independent stats
	stats1 := builder.rateLimiter.Stats(key1)
	stats2 := builder.rateLimiter.Stats(key2)

	if stats1.Allowed != int64(config.BurstSize) {
		t.Errorf("key1: expected %d allowed, got %d", config.BurstSize, stats1.Allowed)
	}
	if stats2.Allowed != int64(config.BurstSize) {
		t.Errorf("key2: expected %d allowed, got %d", config.BurstSize, stats2.Allowed)
	}

	if stats1.Denied != 1 || stats2.Denied != 1 {
		t.Errorf("each key should have 1 denied: key1=%d, key2=%d", stats1.Denied, stats2.Denied)
	}
}

// TestBuilderRateLimitRefill tests token refill behavior
func TestBuilderRateLimitRefill(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 100.0, // Fast refill for quick test
		BurstSize:         2,
	}

	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimitConfig(config)

	err := builder.ensureRateLimiter()
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Exhaust burst
	for i := 0; i < config.BurstSize; i++ {
		if !builder.rateLimiter.Allow("") {
			t.Errorf("initial request %d should be allowed", i)
		}
	}
	if builder.rateLimiter.Allow("") {
		t.Error("should be rate limited after burst")
	}

	// Wait for refill (at 100 req/s, ~20ms should add 2 tokens)
	time.Sleep(30 * time.Millisecond)

	// Should allow more requests after refill
	if !builder.rateLimiter.Allow("") {
		t.Error("request should be allowed after refill")
	}
}

// TestBuilderRateLimitContextCancellation tests context cancellation during Wait
func TestBuilderRateLimitContextCancellation(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 1.0, // Very slow
		BurstSize:         1,
		WaitTimeout:       10 * time.Second,
	}

	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimitConfig(config)

	err := builder.ensureRateLimiter()
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Exhaust burst
	builder.rateLimiter.Allow("")

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Should fail due to context timeout
	start := time.Now()
	err = builder.rateLimiter.Wait(ctx, "")
	duration := time.Since(start)

	if err == nil {
		t.Error("expected context timeout error")
	}

	// Should timeout quickly
	if duration > 100*time.Millisecond {
		t.Errorf("timeout took too long: %v", duration)
	}
}

// TestBuilderRateLimitDisabled tests that disabled rate limiting doesn't affect requests
func TestBuilderRateLimitDisabled(t *testing.T) {
	// Builder without rate limiting
	builder := NewOpenAI("gpt-4", "test-key")

	if builder.rateLimitEnabled {
		t.Error("rate limiting should be disabled by default")
	}

	// Should not have rate limiter
	if builder.rateLimiter != nil {
		t.Error("rate limiter should be nil when disabled")
	}

	// ensureRateLimiter should not create one if disabled
	err := builder.ensureRateLimiter()
	if err != nil {
		t.Errorf("ensureRateLimiter failed: %v", err)
	}

	// Still should be nil
	if builder.rateLimiter != nil {
		t.Error("rate limiter should remain nil when disabled")
	}
}

// TestBuilderRateLimitStats tests statistics tracking
func TestBuilderRateLimitStats(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 100.0,
		BurstSize:         5,
		WaitTimeout:       5 * time.Second,
	}

	builder := NewOpenAI("gpt-4", "test-key").
		WithRateLimitConfig(config)

	err := builder.ensureRateLimiter()
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Make some Allow requests (within burst)
	allowCount := 0
	for i := 0; i < 3; i++ {
		if builder.rateLimiter.Allow("") {
			allowCount++
		}
	}

	// Make a Wait request
	ctx := context.Background()
	if err := builder.rateLimiter.Wait(ctx, ""); err == nil {
		allowCount++
	}

	// Exhaust remaining burst
	for builder.rateLimiter.Allow("") {
		allowCount++
	}

	// This should be denied
	if builder.rateLimiter.Allow("") {
		t.Error("should be denied after burst exhaustion")
	}

	stats := builder.rateLimiter.Stats("")

	if stats.Allowed < int64(allowCount) {
		t.Errorf("expected at least %d allowed, got %d", allowCount, stats.Allowed)
	}

	if stats.Denied < 1 {
		t.Errorf("expected at least 1 denied, got %d", stats.Denied)
	}

	if stats.LastUpdate.IsZero() {
		t.Error("last update should be set")
	}

	if stats.AvailableTokens < 0 {
		t.Errorf("available tokens should be non-negative, got %f", stats.AvailableTokens)
	}
}

// TestBuilderRateLimitInvalidConfig tests validation of invalid configurations
func TestBuilderRateLimitInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config RateLimitConfig
	}{
		{
			name: "zero requests per second",
			config: RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: 0,
				BurstSize:         10,
			},
		},
		{
			name: "negative requests per second",
			config: RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: -5.0,
				BurstSize:         10,
			},
		},
		{
			name: "zero burst size",
			config: RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: 10.0,
				BurstSize:         0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOpenAI("gpt-4", "test-key").
				WithRateLimitConfig(tt.config)

			err := builder.ensureRateLimiter()
			if err == nil {
				t.Error("expected error for invalid config, got nil")
			}
		})
	}
}
