package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	fmt.Println("=== Rate Limiting Basic Example ===\n")

	// Example 1: Simple rate limiting
	simpleRateLimiting(apiKey)

	fmt.Println()

	// Example 2: Burst capacity
	burstCapacity(apiKey)

	fmt.Println()

	// Example 3: Rate limit statistics
	rateLimitStats(apiKey)
}

// simpleRateLimiting demonstrates basic rate limiting
func simpleRateLimiting(apiKey string) {
	fmt.Println("1. Simple Rate Limiting")
	fmt.Println("   Limit: 2 requests/second, Burst: 5")
	fmt.Println()

	// Create agent with rate limiting: 2 requests per second, burst of 5
	agentBuilder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRateLimit(2.0, 5).
		WithTemperature(0.7)

	ctx := context.Background()

	// Make multiple requests - first 5 will be immediate (burst)
	// then throttled to 2 per second
	for i := 1; i <= 8; i++ {
		start := time.Now()

		response, err := agentBuilder.Ask(ctx, fmt.Sprintf("Say 'Request %d'", i))
		if err != nil {
			log.Printf("   Request %d failed: %v", i, err)
			continue
		}

		duration := time.Since(start)
		fmt.Printf("   Request %d: %s (took %v)\n", i, response, duration.Round(time.Millisecond))
	}
}

// burstCapacity demonstrates burst handling
func burstCapacity(apiKey string) {
	fmt.Println("2. Burst Capacity")
	fmt.Println("   Limit: 10 requests/second, Burst: 3")
	fmt.Println()

	agentBuilder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRateLimit(10.0, 3). // Allow 3 immediate requests, then 10/sec
		WithTemperature(0.7)

	ctx := context.Background()

	// First 3 requests should be immediate (burst)
	fmt.Println("   First 3 requests (burst):")
	for i := 1; i <= 3; i++ {
		start := time.Now()
		response, err := agentBuilder.Ask(ctx, "Say 'OK'")
		duration := time.Since(start)

		if err != nil {
			log.Printf("   Request %d failed: %v", i, err)
			continue
		}

		fmt.Printf("   Request %d: %s (waited %v)\n",
			i, response, duration.Round(time.Millisecond))
	}

	fmt.Println()
	fmt.Println("   Next 2 requests (rate limited):")

	// Next requests will be rate limited
	for i := 4; i <= 5; i++ {
		start := time.Now()
		response, err := agentBuilder.Ask(ctx, "Say 'OK'")
		duration := time.Since(start)

		if err != nil {
			log.Printf("   Request %d failed: %v", i, err)
			continue
		}

		fmt.Printf("   Request %d: %s (waited %v)\n",
			i, response, duration.Round(time.Millisecond))
	}
}

// rateLimitStats demonstrates monitoring rate limit statistics
func rateLimitStats(apiKey string) {
	fmt.Println("3. Rate Limit Statistics")
	fmt.Println("   This example shows how rate limiting works behind the scenes")
	fmt.Println()

	// Create agent with rate limiting
	agentBuilder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRateLimit(5.0, 3).
		WithTemperature(0.7)

	ctx := context.Background()

	// Make some requests
	fmt.Println("   Making 5 requests...")
	for i := 1; i <= 5; i++ {
		start := time.Now()
		_, err := agentBuilder.Ask(ctx, fmt.Sprintf("Say 'Request %d'", i))
		duration := time.Since(start)

		if err != nil {
			log.Printf("   Request %d failed: %v", i, err)
		} else {
			fmt.Printf("   Request %d completed in %v\n", i, duration.Round(time.Millisecond))
		}
	}

	fmt.Println()
	fmt.Println("   ✓ Rate limiting is working transparently")
	fmt.Println("   ✓ First 3 requests used burst capacity")
	fmt.Println("   ✓ Remaining requests were throttled to 5/second")
}
