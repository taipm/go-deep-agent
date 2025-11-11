package ratelimitadvanced
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	fmt.Println("=== Rate Limiting Advanced Example ===\n")

	// Example 1: Per-key rate limiting
	perKeyRateLimiting(apiKey)

	fmt.Println()

	// Example 2: Advanced configuration
	advancedConfiguration(apiKey)

	fmt.Println()

	// Example 3: Concurrent requests with rate limiting
	concurrentRequests(apiKey)
}

// perKeyRateLimiting demonstrates per-user/per-key rate limiting
func perKeyRateLimiting(apiKey string) {
	fmt.Println("1. Per-Key Rate Limiting (e.g., per-user limits)")
	fmt.Println("   Limit: 3 requests/second per user, Burst: 5")
	fmt.Println()

	// Enable per-key rate limiting
	config := agent.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 3.0,
		BurstSize:         5,
		PerKey:            true,  // Enable per-key limits
		KeyTimeout:        5 * time.Minute,
		WaitTimeout:       30 * time.Second,
	}

	ctx := context.Background()

	// Simulate different users making requests
	users := []string{"user-alice", "user-bob", "user-charlie"}

	fmt.Println("   Simulating 3 users making requests simultaneously:")
	var wg sync.WaitGroup

	for _, user := range users {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()

			// Each user gets their own rate limit
			agentBuilder := agent.NewOpenAI("gpt-4o-mini", apiKey).
				WithRateLimitConfig(config).
				WithRateLimitKey(userID). // Set the user ID as rate limit key
				WithTemperature(0.7)

			// Make 3 requests per user
			for i := 1; i <= 3; i++ {
				start := time.Now()
				response, err := agentBuilder.Ask(ctx, 
					fmt.Sprintf("Say 'Hello from %s, request %d'", userID, i))
				duration := time.Since(start)

				if err != nil {
					log.Printf("   [%s] Request %d failed: %v", userID, i, err)
					continue
				}

				fmt.Printf("   [%s] Request %d: %s (took %v)\n", 
					userID, i, response, duration.Round(time.Millisecond))
			}
		}(user)
	}

	wg.Wait()
	fmt.Println()
	fmt.Println("   ✓ Each user has independent rate limits")
	fmt.Println("   ✓ Users don't affect each other's quotas")
}

// advancedConfiguration demonstrates custom rate limit configuration
func advancedConfiguration(apiKey string) {
	fmt.Println("2. Advanced Configuration")
	fmt.Println("   Custom timeouts and behavior")
	fmt.Println()

	// Create custom configuration
	config := agent.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 5.0,
		BurstSize:         10,
		PerKey:            false,
		KeyTimeout:        10 * time.Minute,  // Cleanup unused keys after 10 min
		WaitTimeout:       5 * time.Second,   // Max wait time per request
	}

	agentBuilder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRateLimitConfig(config).
		WithTemperature(0.7)

	ctx := context.Background()

	fmt.Println("   Configuration:")
	fmt.Printf("   - Sustained Rate: %.1f requests/second\n", config.RequestsPerSecond)
	fmt.Printf("   - Burst Capacity: %d requests\n", config.BurstSize)
	fmt.Printf("   - Wait Timeout: %v\n", config.WaitTimeout)
	fmt.Printf("   - Per-Key Enabled: %v\n", config.PerKey)
	fmt.Println()

	// Make requests
	fmt.Println("   Making requests:")
	for i := 1; i <= 5; i++ {
		start := time.Now()
		response, err := agentBuilder.Ask(ctx, fmt.Sprintf("Say 'Request %d'", i))
		duration := time.Since(start)

		if err != nil {
			log.Printf("   Request %d failed: %v", i, err)
			continue
		}

		fmt.Printf("   Request %d: %s (waited %v)\n", 
			i, response, duration.Round(time.Millisecond))
	}
}

// concurrentRequests demonstrates handling concurrent requests with rate limiting
func concurrentRequests(apiKey string) {
	fmt.Println("3. Concurrent Requests with Rate Limiting")
	fmt.Println("   Limit: 10 requests/second, Burst: 5")
	fmt.Println()

	agentBuilder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRateLimit(10.0, 5).
		WithTemperature(0.7)

	ctx := context.Background()

	// Launch 15 concurrent requests
	numRequests := 15
	var wg sync.WaitGroup
	results := make(chan string, numRequests)
	startTime := time.Now()

	fmt.Printf("   Launching %d concurrent requests...\n", numRequests)

	for i := 1; i <= numRequests; i++ {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()

			start := time.Now()
			response, err := agentBuilder.Ask(ctx, 
				fmt.Sprintf("Say 'Request %d'", requestID))
			duration := time.Since(start)

			if err != nil {
				results <- fmt.Sprintf("   Request %d: FAILED (%v)", requestID, err)
			} else {
				results <- fmt.Sprintf("   Request %d: %s (waited %v)", 
					requestID, response, duration.Round(time.Millisecond))
			}
		}(i)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Print results as they come in
	for result := range results {
		fmt.Println(result)
	}

	totalDuration := time.Since(startTime)
	fmt.Println()
	fmt.Printf("   Total time: %v\n", totalDuration.Round(time.Millisecond))
	fmt.Printf("   Average: %.2f requests/second\n", 
		float64(numRequests)/totalDuration.Seconds())
	fmt.Println()
	fmt.Println("   ✓ Rate limiting automatically queued concurrent requests")
	fmt.Println("   ✓ First 5 requests used burst capacity")
	fmt.Println("   ✓ Remaining requests throttled to 10/second")
}
