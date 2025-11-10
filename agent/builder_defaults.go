package agent

import "time"

// WithDefaults configures the agent with production-ready default settings.
// This is the recommended starting point for most production use cases.
//
// Default Configuration:
//   - Memory(20): Keeps last 20 messages in conversation history
//   - Retry(3): Retries failed requests up to 3 times
//   - Timeout(30s): Sets 30-second timeout for API requests
//   - ExponentialBackoff: Uses exponential backoff for retries (1s, 2s, 4s, ...)
//
// Philosophy:
//   - Covers 80% of production use cases out-of-the-box
//   - Users can customize any setting via method chaining
//   - Opt-in features (Tools, Logging, Cache) remain disabled by default
//
// Example - Basic usage:
//
//	agent := agent.NewOpenAI(apiKey).WithDefaults()
//	resp, _ := agent.Ask("Hello")
//
// Example - Customize defaults:
//
//	agent := agent.NewOpenAI(apiKey).
//	    WithDefaults().
//	    Memory(50).              // Override memory to 50
//	    WithTools(search).       // Add tools
//	    WithLogging(logger)      // Add logging
//
// Example - Opt-out of specific defaults:
//
//	agent := agent.NewOpenAI(apiKey).
//	    WithDefaults().
//	    DisableMemory()          // Remove memory
//
// Returns:
//   - *Builder: The builder instance with defaults configured (chainable)
func (b *Builder) WithDefaults() *Builder {
	// Memory: Keep last 20 messages
	b.WithMaxHistory(20)

	// Retry: Retry failed requests up to 3 times
	b.WithRetry(3)

	// Timeout: 30-second timeout for API requests
	b.WithTimeout(30 * time.Second)

	// ExponentialBackoff: Smart retry delays (1s, 2s, 4s, 8s, ...)
	b.WithExponentialBackoff()

	return b
}
