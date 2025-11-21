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
	ctx := context.Background()

	fmt.Println("🚀 MultiProvider Basic Example")
	fmt.Println("=============================")

	// Define multiple providers with different characteristics
	providers := []agent.ProviderConfig{
		{
			Name:    "openai-primary",
			Type:    "openai",
			Model:   "gpt-4o-mini",
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			Weight:  3.0, // Higher weight = more traffic
			MaxConcurrency: 10,
			Timeout: 10 * time.Second,
		},
		{
			Name:    "ollama-backup",
			Type:    "ollama",
			Model:   "llama2",
			BaseURL: "http://localhost:11434",
			Weight:  2.0, // Medium weight
			MaxConcurrency: 5,
			Timeout: 15 * time.Second,
		},
		{
			Name:    "openai-budget",
			Type:    "openai",
			Model:   "gpt-3.5-turbo",
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			Weight:  1.0, // Lower weight = less traffic
			MaxConcurrency: 3,
			Timeout: 20 * time.Second,
		},
	}

	// Create MultiProvider configuration
	config := &agent.MultiProviderConfig{
		Providers: providers,

		// Load balancing strategy
		SelectionStrategy: agent.StrategyWeightedRoundRobin,

		// Fallback strategy with circuit breaker
		FallbackStrategy:      agent.FallbackStrategyCircuitBreaker,
		CircuitBreakerThreshold: 3, // Open after 3 failures
		CircuitBreakerTimeout:   30 * time.Second,

		// Health monitoring
		HealthCheckInterval: 15 * time.Second,
		HealthCheckTimeout:  5 * time.Second,

		// Load balancing
		EnableLoadBalancing: true,
		StickySessions:      false,

		// Monitoring
		EnableMetrics: true,
	}

	// Create MultiProvider instance
	mp, err := agent.NewMultiProvider(config)
	if err != nil {
		log.Fatal("Failed to create MultiProvider:", err)
	}
	defer mp.Shutdown(ctx)

	fmt.Printf("✅ MultiProvider created with %d providers\n", len(providers))

	// Test the MultiProvider
	fmt.Println("\n📝 Testing Basic MultiProvider Functionality")
	fmt.Println("------------------------------------------")

	// Make several requests to see load balancing in action
	for i := 0; i < 5; i++ {
		message := fmt.Sprintf("This is test message #%d", i+1)

		response, err := mp.Ask(ctx, message)
		if err != nil {
			fmt.Printf("❌ Request #%d failed: %v\n", i+1, err)
			continue
		}

		fmt.Printf("✅ Request #%d: %s\n", i+1, response)
	}

	// Display provider health status
	fmt.Println("\n🏥 Provider Health Status")
	fmt.Println("------------------------")

	healthStatus := mp.GetProviderStatus()
	for name, status := range healthStatus {
		fmt.Printf("%-20s: %s\n", name, status.String())
	}

	// Display metrics
	fmt.Println("\n📊 Performance Metrics")
	fmt.Println("---------------------")

	metrics := mp.GetMetrics()
	for name, providerMetrics := range metrics {
		if providerMetrics.TotalRequests > 0 {
			fmt.Printf("%-20s:\n", name)
			fmt.Printf("  📈 Total Requests: %d\n", providerMetrics.TotalRequests)
			fmt.Printf("  ✅ Successful: %d\n", providerMetrics.SuccessfulRequests)
			fmt.Printf("  ❌ Failed: %d\n", providerMetrics.FailedRequests)
			fmt.Printf("  📈 Success Rate: %.2f%%\n", providerMetrics.UptimePercentage)
			if providerMetrics.AverageResponseTime > 0 {
				fmt.Printf("  ⏱️  Avg Response: %s\n", providerMetrics.AverageResponseTime)
			}
		}
	}

	// Demonstrate dynamic provider management
	fmt.Println("\n🔧 Dynamic Provider Management")
	fmt.Println("------------------------------")

	// Add a new provider
	newProvider := agent.ProviderConfig{
		Name:    "emergency-provider",
		Type:    "openai",
		Model:   "gpt-4o",
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Weight:  0.5, // Low priority
		MaxConcurrency: 1,
	}

	err = mp.AddProvider(newProvider)
	if err != nil {
		fmt.Printf("❌ Failed to add provider: %v\n", err)
	} else {
		fmt.Println("✅ Added emergency provider")
	}

	// Test with the new provider
	response, err := mp.Ask(ctx, "Test with emergency provider")
	if err != nil {
		fmt.Printf("❌ Emergency provider test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Emergency provider response: %s\n", response)
	}

	// Remove the emergency provider
	err = mp.RemoveProvider("emergency-provider")
	if err != nil {
		fmt.Printf("❌ Failed to remove provider: %v\n", err)
	} else {
		fmt.Println("✅ Removed emergency provider")
	}

	fmt.Println("\n🎉 MultiProvider Example Completed Successfully!")
	fmt.Println("========================================")

	// Final metrics summary
	finalMetrics := mp.GetMetrics()
	totalRequests := int64(0)
	totalSuccessful := int64(0)

	for _, providerMetrics := range finalMetrics {
		totalRequests += providerMetrics.TotalRequests
		totalSuccessful += providerMetrics.SuccessfulRequests
	}

	fmt.Printf("📈 Summary: %d total requests, %d successful (%.1f%% success rate)\n",
		totalRequests, totalSuccessful,
		float64(totalSuccessful)/float64(totalRequests)*100)
}