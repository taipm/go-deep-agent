package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// tokenBucketLimiter implements RateLimiter using the token bucket algorithm.
// It uses golang.org/x/time/rate internally for efficient rate limiting.
type tokenBucketLimiter struct {
	config RateLimitConfig

	// For global rate limiting
	globalLimiter *rate.Limiter
	globalStats   *rateLimitStats

	// For per-key rate limiting
	perKeyLimiters map[string]*perKeyLimiter
	mu             sync.RWMutex

	// Cleanup goroutine control
	stopCleanup chan struct{}
	cleanupOnce sync.Once
}

// perKeyLimiter holds a rate limiter and stats for a specific key.
type perKeyLimiter struct {
	limiter    *rate.Limiter
	stats      *rateLimitStats
	lastAccess time.Time
	mu         sync.RWMutex
}

// rateLimitStats holds thread-safe statistics.
type rateLimitStats struct {
	allowed       int64
	denied        int64
	waited        int64
	totalWaitTime time.Duration
	lastUpdate    time.Time
	mu            sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with the given configuration.
// Returns an error if the configuration is invalid.
func NewRateLimiter(config RateLimitConfig) (RateLimiter, error) {
	// Validate configuration
	if config.RequestsPerSecond <= 0 {
		return nil, fmt.Errorf("RequestsPerSecond must be positive, got %f", config.RequestsPerSecond)
	}
	if config.BurstSize < 1 {
		return nil, fmt.Errorf("BurstSize must be >= 1, got %d", config.BurstSize)
	}

	// Set default values
	if config.KeyTimeout == 0 {
		config.KeyTimeout = 5 * time.Minute
	}
	if config.WaitTimeout == 0 {
		config.WaitTimeout = 30 * time.Second
	}

	limiter := &tokenBucketLimiter{
		config:         config,
		globalStats:    &rateLimitStats{lastUpdate: time.Now()},
		perKeyLimiters: make(map[string]*perKeyLimiter),
		stopCleanup:    make(chan struct{}),
	}

	// Create global limiter if not using per-key limiting
	if !config.PerKey {
		limiter.globalLimiter = rate.NewLimiter(
			rate.Limit(config.RequestsPerSecond),
			config.BurstSize,
		)
	} else {
		// Start cleanup goroutine for per-key limiters
		go limiter.cleanupUnusedLimiters()
	}

	return limiter, nil
}

// Allow checks if a request is allowed under the current rate limit.
func (tb *tokenBucketLimiter) Allow(key string) bool {
	limiter, stats := tb.getLimiterAndStats(key)

	allowed := limiter.Allow()

	stats.mu.Lock()
	if allowed {
		stats.allowed++
	} else {
		stats.denied++
	}
	stats.lastUpdate = time.Now()
	stats.mu.Unlock()

	// Update last access time for per-key limiters
	if tb.config.PerKey && key != "" {
		tb.updateLastAccess(key)
	}

	return allowed
}

// Wait blocks until the rate limiter allows the request to proceed.
func (tb *tokenBucketLimiter) Wait(ctx context.Context, key string) error {
	limiter, stats := tb.getLimiterAndStats(key)

	// Apply wait timeout if configured
	if tb.config.WaitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, tb.config.WaitTimeout)
		defer cancel()
	}

	start := time.Now()
	err := limiter.Wait(ctx)
	waitDuration := time.Since(start)

	stats.mu.Lock()
	if err == nil {
		stats.waited++
		stats.totalWaitTime += waitDuration
		stats.allowed++
	}
	stats.lastUpdate = time.Now()
	stats.mu.Unlock()

	// Update last access time for per-key limiters
	if tb.config.PerKey && key != "" {
		tb.updateLastAccess(key)
	}

	return err
}

// Reserve reserves capacity for a request.
func (tb *tokenBucketLimiter) Reserve(key string) *Reservation {
	limiter, stats := tb.getLimiterAndStats(key)

	res := limiter.Reserve()
	if !res.OK() {
		return &Reservation{ok: false}
	}

	delay := res.Delay()
	timeToAct := time.Now().Add(delay)

	stats.mu.Lock()
	if delay > 0 {
		stats.waited++
		stats.totalWaitTime += delay
	}
	stats.allowed++
	stats.lastUpdate = time.Now()
	stats.mu.Unlock()

	// Update last access time for per-key limiters
	if tb.config.PerKey && key != "" {
		tb.updateLastAccess(key)
	}

	return &Reservation{
		ok:        true,
		delay:     delay,
		timeToAct: timeToAct,
		cancel: func() {
			res.Cancel()
			stats.mu.Lock()
			stats.allowed--
			stats.mu.Unlock()
		},
	}
}

// Stats returns current rate limiting statistics.
func (tb *tokenBucketLimiter) Stats(key string) RateLimitStats {
	limiter, stats := tb.getLimiterAndStats(key)

	stats.mu.RLock()
	defer stats.mu.RUnlock()

	result := RateLimitStats{
		Allowed:         stats.allowed,
		Denied:          stats.denied,
		Waited:          stats.waited,
		TotalWaitTime:   stats.totalWaitTime,
		LastUpdate:      stats.lastUpdate,
		AvailableTokens: float64(limiter.Tokens()),
	}

	if tb.config.PerKey {
		tb.mu.RLock()
		result.ActiveKeys = len(tb.perKeyLimiters)
		tb.mu.RUnlock()
	}

	return result
}

// getLimiterAndStats returns the appropriate limiter and stats for the given key.
func (tb *tokenBucketLimiter) getLimiterAndStats(key string) (*rate.Limiter, *rateLimitStats) {
	if !tb.config.PerKey {
		return tb.globalLimiter, tb.globalStats
	}

	// Per-key rate limiting
	tb.mu.RLock()
	pkl, exists := tb.perKeyLimiters[key]
	tb.mu.RUnlock()

	if exists {
		return pkl.limiter, pkl.stats
	}

	// Create new per-key limiter
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Double-check after acquiring write lock
	if pkl, exists := tb.perKeyLimiters[key]; exists {
		return pkl.limiter, pkl.stats
	}

	pkl = &perKeyLimiter{
		limiter: rate.NewLimiter(
			rate.Limit(tb.config.RequestsPerSecond),
			tb.config.BurstSize,
		),
		stats:      &rateLimitStats{lastUpdate: time.Now()},
		lastAccess: time.Now(),
	}
	tb.perKeyLimiters[key] = pkl

	return pkl.limiter, pkl.stats
}

// updateLastAccess updates the last access time for a per-key limiter.
func (tb *tokenBucketLimiter) updateLastAccess(key string) {
	tb.mu.RLock()
	pkl, exists := tb.perKeyLimiters[key]
	tb.mu.RUnlock()

	if exists {
		pkl.mu.Lock()
		pkl.lastAccess = time.Now()
		pkl.mu.Unlock()
	}
}

// cleanupUnusedLimiters periodically removes unused per-key limiters.
func (tb *tokenBucketLimiter) cleanupUnusedLimiters() {
	ticker := time.NewTicker(tb.config.KeyTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tb.performCleanup()
		case <-tb.stopCleanup:
			return
		}
	}
}

// performCleanup removes limiters that haven't been accessed recently.
func (tb *tokenBucketLimiter) performCleanup() {
	now := time.Now()
	keysToDelete := []string{}

	tb.mu.RLock()
	for key, pkl := range tb.perKeyLimiters {
		pkl.mu.RLock()
		if now.Sub(pkl.lastAccess) > tb.config.KeyTimeout {
			keysToDelete = append(keysToDelete, key)
		}
		pkl.mu.RUnlock()
	}
	tb.mu.RUnlock()

	if len(keysToDelete) > 0 {
		tb.mu.Lock()
		for _, key := range keysToDelete {
			delete(tb.perKeyLimiters, key)
		}
		tb.mu.Unlock()
	}
}

// Stop stops the cleanup goroutine for per-key limiters.
// Should be called when the rate limiter is no longer needed.
func (tb *tokenBucketLimiter) Stop() {
	tb.cleanupOnce.Do(func() {
		close(tb.stopCleanup)
	})
}
