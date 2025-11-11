package agent

import (
	"context"
	"testing"
	"time"
)

func TestNewRateLimiter_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      RateLimitConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: 10.0,
				BurstSize:         20,
			},
			expectError: false,
		},
		{
			name: "zero requests per second",
			config: RateLimitConfig{
				RequestsPerSecond: 0,
				BurstSize:         10,
			},
			expectError: true,
		},
		{
			name: "negative requests per second",
			config: RateLimitConfig{
				RequestsPerSecond: -5,
				BurstSize:         10,
			},
			expectError: true,
		},
		{
			name: "zero burst size",
			config: RateLimitConfig{
				RequestsPerSecond: 10,
				BurstSize:         0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRateLimiter(tt.config)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         5,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	// Should allow burst requests
	for i := 0; i < config.BurstSize; i++ {
		if !limiter.Allow("") {
			t.Errorf("request %d should be allowed (within burst)", i)
		}
	}

	// Next request should be denied (burst exhausted)
	if limiter.Allow("") {
		t.Error("request should be denied (burst exhausted)")
	}

	// Check stats
	stats := limiter.Stats("")
	if stats.Allowed != int64(config.BurstSize) {
		t.Errorf("expected %d allowed, got %d", config.BurstSize, stats.Allowed)
	}
	if stats.Denied != 1 {
		t.Errorf("expected 1 denied, got %d", stats.Denied)
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 100.0, // High rate for faster test
		BurstSize:         2,
		WaitTimeout:       5 * time.Second,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	ctx := context.Background()

	// First two requests should not wait (burst)
	start := time.Now()
	for i := 0; i < 2; i++ {
		if err := limiter.Wait(ctx, ""); err != nil {
			t.Errorf("request %d failed: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		t.Errorf("burst requests took too long: %v", elapsed)
	}

	// Third request should wait
	start = time.Now()
	if err := limiter.Wait(ctx, ""); err != nil {
		t.Errorf("wait failed: %v", err)
	}
	elapsed = time.Since(start)
	if elapsed < 5*time.Millisecond { // Should wait at least a bit
		t.Errorf("expected to wait, but completed in %v", elapsed)
	}

	// Check stats
	stats := limiter.Stats("")
	if stats.Waited < 1 {
		t.Errorf("expected at least 1 wait, got %d", stats.Waited)
	}
	if stats.TotalWaitTime == 0 {
		t.Error("expected non-zero total wait time")
	}
}

func TestRateLimiter_Wait_ContextCancellation(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 1.0, // Very slow rate
		BurstSize:         1,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	// Exhaust burst
	limiter.Allow("")

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Should fail due to context timeout
	err = limiter.Wait(ctx, "")
	if err == nil {
		t.Error("expected context deadline error, got nil")
	}
}

func TestRateLimiter_Reserve(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         2,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	// First reservation should be immediate
	r1 := limiter.Reserve("")
	if !r1.OK() {
		t.Error("reservation should be ok")
	}
	if r1.Delay() > 0 {
		t.Errorf("first reservation should have no delay, got %v", r1.Delay())
	}

	// Second reservation should also be immediate (within burst)
	r2 := limiter.Reserve("")
	if !r2.OK() {
		t.Error("reservation should be ok")
	}
	if r2.Delay() > 0 {
		t.Errorf("second reservation should have no delay, got %v", r2.Delay())
	}

	// Third reservation should have delay
	r3 := limiter.Reserve("")
	if !r3.OK() {
		t.Error("reservation should be ok")
	}
	if r3.Delay() == 0 {
		t.Error("third reservation should have delay")
	}

	// Test cancellation
	r3.Cancel()
	// After cancellation, the token is returned to the bucket
}

func TestRateLimiter_PerKey(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         3,
		PerKey:            true,
		KeyTimeout:        1 * time.Second,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	// Different keys should have independent limits
	key1, key2 := "user1", "user2"

	// Exhaust key1's burst
	for i := 0; i < config.BurstSize; i++ {
		if !limiter.Allow(key1) {
			t.Errorf("key1 request %d should be allowed", i)
		}
	}
	if limiter.Allow(key1) {
		t.Error("key1 should be rate limited")
	}

	// key2 should still have full burst available
	for i := 0; i < config.BurstSize; i++ {
		if !limiter.Allow(key2) {
			t.Errorf("key2 request %d should be allowed", i)
		}
	}

	// Check stats are independent
	stats1 := limiter.Stats(key1)
	stats2 := limiter.Stats(key2)

	if stats1.Allowed != int64(config.BurstSize) {
		t.Errorf("key1: expected %d allowed, got %d", config.BurstSize, stats1.Allowed)
	}
	if stats1.Denied != 1 {
		t.Errorf("key1: expected 1 denied, got %d", stats1.Denied)
	}

	if stats2.Allowed != int64(config.BurstSize) {
		t.Errorf("key2: expected %d allowed, got %d", config.BurstSize, stats2.Allowed)
	}
	if stats2.Denied != 0 {
		t.Errorf("key2: expected 0 denied, got %d", stats2.Denied)
	}

	// Check active keys
	if stats1.ActiveKeys < 2 {
		t.Errorf("expected at least 2 active keys, got %d", stats1.ActiveKeys)
	}
}

func TestRateLimiter_Stats(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         5,
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	// Initial stats
	stats := limiter.Stats("")
	if stats.Allowed != 0 || stats.Denied != 0 || stats.Waited != 0 {
		t.Error("initial stats should be zero")
	}

	// Make some requests
	for i := 0; i < 3; i++ {
		limiter.Allow("")
	}

	stats = limiter.Stats("")
	if stats.Allowed != 3 {
		t.Errorf("expected 3 allowed, got %d", stats.Allowed)
	}
	if stats.AvailableTokens <= 0 {
		t.Errorf("expected positive available tokens, got %f", stats.AvailableTokens)
	}
	if stats.LastUpdate.IsZero() {
		t.Error("last update should be set")
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	if config.Enabled {
		t.Error("rate limiting should be disabled by default")
	}
	if config.RequestsPerSecond <= 0 {
		t.Error("RequestsPerSecond should be positive")
	}
	if config.BurstSize < 1 {
		t.Error("BurstSize should be >= 1")
	}
	if config.KeyTimeout == 0 {
		t.Error("KeyTimeout should be set")
	}
	if config.WaitTimeout == 0 {
		t.Error("WaitTimeout should be set")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	config := RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 10.0,
		BurstSize:         5,
		PerKey:            true,
		KeyTimeout:        100 * time.Millisecond, // Short timeout for test
	}

	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	tb, ok := limiter.(*tokenBucketLimiter)
	if !ok {
		t.Fatal("limiter is not tokenBucketLimiter")
	}
	defer tb.Stop()

	// Create some per-key limiters
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		limiter.Allow(key)
	}

	// Verify they exist
	tb.mu.RLock()
	activeCount := len(tb.perKeyLimiters)
	tb.mu.RUnlock()

	if activeCount != len(keys) {
		t.Errorf("expected %d active keys, got %d", len(keys), activeCount)
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Keys should be cleaned up
	tb.mu.RLock()
	activeCount = len(tb.perKeyLimiters)
	tb.mu.RUnlock()

	if activeCount != 0 {
		t.Errorf("expected 0 active keys after cleanup, got %d", activeCount)
	}
}
