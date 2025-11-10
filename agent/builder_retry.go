package agent

import (
	"time"
)

// Retry configuration methods for Builder
// This file contains methods for configuring retry behavior.
// Retry execution logic is in builder_execution.go

func (b *Builder) WithRetry(maxRetries int) *Builder {
	b.maxRetries = maxRetries
	if b.retryDelay == 0 {
		b.retryDelay = time.Second // Default 1s
	}
	return b
}

func (b *Builder) WithRetryDelay(delay time.Duration) *Builder {
	b.retryDelay = delay
	return b
}

func (b *Builder) WithExponentialBackoff() *Builder {
	b.useExpBackoff = true
	if b.retryDelay == 0 {
		b.retryDelay = time.Second // Default 1s base
	}
	return b
}
