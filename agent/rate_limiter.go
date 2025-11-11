package agent

import (
	"context"
	"time"
)

// RateLimiter defines the interface for rate limiting operations.
// It provides methods to check, wait for, and reserve capacity for requests.
type RateLimiter interface {
	// Allow checks if a request is allowed under the current rate limit.
	// Returns true if the request can proceed, false otherwise.
	// The key parameter allows for per-key rate limiting (e.g., per-user, per-API-key).
	// Use empty string for global rate limiting.
	Allow(key string) bool

	// Wait blocks until the rate limiter allows the request to proceed.
	// Returns an error if the context is cancelled or deadline exceeded.
	// The key parameter allows for per-key rate limiting.
	Wait(ctx context.Context, key string) error

	// Reserve reserves capacity for a request and returns when it can proceed.
	// Returns a Reservation that can be used to check delay or cancel the reservation.
	// The key parameter allows for per-key rate limiting.
	Reserve(key string) *Reservation

	// Stats returns current rate limiting statistics.
	// The key parameter allows getting stats for a specific key.
	// Use empty string for global stats.
	Stats(key string) RateLimitStats
}

// RateLimitConfig holds the configuration for rate limiting.
type RateLimitConfig struct {
	// Enabled determines if rate limiting is active.
	Enabled bool

	// RequestsPerSecond is the sustained rate of requests allowed.
	// For example, 10.0 means 10 requests per second on average.
	RequestsPerSecond float64

	// BurstSize is the maximum number of requests that can be made at once.
	// This allows for short bursts above the sustained rate.
	// Must be >= 1. Recommended: 2-5x RequestsPerSecond for API calls.
	BurstSize int

	// PerKey enables per-key rate limiting when true.
	// When false, rate limiting applies globally across all keys.
	PerKey bool

	// KeyTimeout is the duration after which unused per-key limiters are cleaned up.
	// Only applies when PerKey is true.
	// Default: 5 minutes.
	KeyTimeout time.Duration

	// WaitTimeout is the maximum duration to wait for rate limit availability.
	// If zero, waits indefinitely (subject to context cancellation).
	WaitTimeout time.Duration
}

// RateLimitStats contains statistics about rate limiting.
type RateLimitStats struct {
	// Allowed is the total number of requests that were allowed.
	Allowed int64

	// Denied is the total number of requests that were denied (via Allow).
	Denied int64

	// Waited is the total number of requests that had to wait.
	Waited int64

	// TotalWaitTime is the cumulative time spent waiting.
	TotalWaitTime time.Duration

	// ActiveKeys is the number of active per-key limiters (when PerKey is enabled).
	ActiveKeys int

	// AvailableTokens is the current number of available tokens in the bucket.
	AvailableTokens float64

	// LastUpdate is the timestamp of the last rate limit check.
	LastUpdate time.Time
}

// Reservation holds information about a reserved rate limit token.
type Reservation struct {
	// ok indicates if the reservation is valid.
	ok bool

	// delay is the duration to wait before the request can proceed.
	delay time.Duration

	// timeToAct is the time when the request can proceed.
	timeToAct time.Time

	// cancel is a function to cancel the reservation and return the token.
	cancel func()
}

// OK returns whether the reservation is valid.
func (r *Reservation) OK() bool {
	return r.ok
}

// Delay returns the duration to wait before the request can proceed.
// Returns 0 if the request can proceed immediately.
func (r *Reservation) Delay() time.Duration {
	if !r.ok {
		return 0
	}
	now := time.Now()
	if r.timeToAct.After(now) {
		return r.timeToAct.Sub(now)
	}
	return 0
}

// DelayFrom returns the duration to wait from the given time.
func (r *Reservation) DelayFrom(t time.Time) time.Duration {
	if !r.ok {
		return 0
	}
	if r.timeToAct.After(t) {
		return r.timeToAct.Sub(t)
	}
	return 0
}

// Cancel cancels the reservation and returns the reserved token to the bucket.
// This should be called if the request is not going to be made.
func (r *Reservation) Cancel() {
	if r.cancel != nil {
		r.cancel()
	}
}

// RateLimitContext provides per-request rate limiting information.
// It can be used to track metrics and make rate limiting decisions.
type RateLimitContext struct {
	// Key is the rate limit key for this request.
	Key string

	// WaitDuration is how long the request had to wait.
	WaitDuration time.Duration

	// Allowed indicates if the request was allowed.
	Allowed bool

	// Timestamp is when the rate limit check occurred.
	Timestamp time.Time
}

// DefaultRateLimitConfig returns a sensible default configuration.
// Rate limiting is disabled by default to maintain backward compatibility.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:           false,
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		PerKey:            false,
		KeyTimeout:        5 * time.Minute,
		WaitTimeout:       30 * time.Second,
	}
}
