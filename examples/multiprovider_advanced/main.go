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

	fmt.Println("🚀 MultiProvider Advanced Example")
	fmt.Println("===============================")

	// Create a custom adapter for demonstration
	customAdapter := &mockAdapter{
		responses: []string{
			"Custom adapter response 1",
			"Custom adapter response 2",
			"Custom adapter response 3",
			"Custom adapter response 4",
			"Custom adapter response 5",
		},
	}

	// Advanced multi-provider setup with different strategies
	providers := []agent.ProviderConfig{
		// Primary: High-performance OpenAI
		{
			Name:    "openai-gpt4",
			Type:    "openai",
			Model:   "gpt-4",
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			Weight:  4.0, // Highest priority
			MaxConcurrency: 5,
			Timeout: 15 * time.Second,
		},
		// Secondary: Cost-effective OpenAI
		{
			Name:    "openai-gpt35",
			Type:    "openai",
			Model:   "gpt-3.5-turbo",
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			Weight:  2.0,
			MaxConcurrency: 10,
			Timeout: 10 * time.Second,
		},
		// Local: Custom adapter
		{
			Name:    "custom-adapter",
			Type:    "adapter",
			Model:   "custom-model",
			Adapter: customAdapter,
			Weight:  1.0,
			MaxConcurrency: 3,
			Timeout: 5 * time.Second,
		},
	}

	// Advanced configuration with all features enabled
	config := &agent.MultiProviderConfig{
		Providers: providers,

		// Weighted round-robin for optimal load distribution
		SelectionStrategy: agent.StrategyWeightedRoundRobin,

		// Circuit breaker with graceful degradation
		FallbackStrategy:       agent.FallbackStrategyGracefulDegradation,
		CircuitBreakerThreshold: 3,
		CircuitBreakerTimeout:   45 * time.Second,

		// Aggressive health monitoring
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  3 * time.Second,

		// Advanced load balancing
		EnableLoadBalancing: true,
		StickySessions:      true, // Important for user experience

		// Comprehensive monitoring
		EnableMetrics:        true,
		MetricsInterval:      5 * time.Second,
		EnableDetailedLogging: true,
		LogLevel:             "info",
	}

	// Create advanced MultiProvider
	mp, err := agent.NewMultiProvider(config)
	if err != nil {
		log.Fatal("Failed to create MultiProvider:", err)
	}
	defer mp.Shutdown(ctx)

	fmt.Printf("✅ Advanced MultiProvider created with %d providers\n", len(providers))

	// Create load balancer and selector for manual control
	loadBalancer := agent.NewLoadBalancer(config)
	selector := agent.NewProviderSelector(config)

	fmt.Println("\n🎯 Testing Advanced Features")
	fmt.Println("==========================")

	// Test 1: Load Balancing with Session Management
	fmt.Println("\n1️⃣ Load Balancing with Sticky Sessions")
	fmt.Println("---------------------------------------")

	// Simulate multiple users with sticky sessions
	users := []string{"alice", "bob", "charlie", "diana", "eve"}

	for _, user := range users {
		provider, err := loadBalancer.SelectProviderForRequest(providers, user)
		if err != nil {
			fmt.Printf("❌ Failed to select provider for %s: %v\n", user, err)
			continue
		}

		response, err := mp.Ask(ctx, fmt.Sprintf("Hello from %s", user))
		if err != nil {
			fmt.Printf("❌ Request from %s failed: %v\n", user, err)
		} else {
			fmt.Printf("✅ %s → %s: %s\n", user, provider.Name, response)
		}
	}

	// Test 2: Provider Selection Strategies
	fmt.Println("\n2️⃣ Testing Different Selection Strategies")
	fmt.Println("------------------------------------------")

	strategies := []agent.SelectionStrategy{
		agent.StrategyRoundRobin,
		agent.StrategyWeightedRoundRobin,
		agent.StrategyLeastConnections,
		agent.StrategyRandom,
		agent.StrategyPriority,
	}

	for _, strategy := range strategies {
		fmt.Printf("\n📍 Strategy: %s\n", strategyToString(strategy))

		for i := 0; i < 3; i++ {
			provider, err := selector.SelectProvider(providers, strategy)
			if err != nil {
				fmt.Printf("   ❌ Selection failed: %v\n", err)
				continue
			}

			fmt.Printf("   📦 Request %d → %s\n", i+1, provider.Name)
		}
	}

	// Test 3: Health Monitoring and Circuit Breaker
	fmt.Println("\n3️⃣ Health Monitoring & Circuit Breaker")
	fmt.Println("-----------------------------------------")

	// Monitor health status over time
	fmt.Println("Monitoring provider health for 20 seconds...")

	for i := 0; i < 4; i++ {
		time.Sleep(5 * time.Second)

		healthStatus := mp.GetProviderStatus()
		fmt.Printf("\n🏥 Health Check #%d:\n", i+1)
		for name, status := range healthStatus {
			statusIcon := getStatusIcon(status)
			fmt.Printf("   %s %s: %s\n", statusIcon, name, status.String())
		}

		// Note: Circuit breaker status is reflected in provider health
		// Unhealthy providers are likely experiencing circuit breaker issues
		fmt.Printf("   💡 Provider health reflects circuit breaker status\n")
	}

	// Test 4: Performance Metrics
	fmt.Println("\n4️⃣ Performance Metrics Analysis")
	fmt.Println("------------------------------")

	// Get detailed metrics
	metrics := mp.GetMetrics()

	// Calculate global metrics from provider metrics
	var totalRequests int64
	var totalSuccessful int64
	for _, providerMetrics := range metrics {
		totalRequests += providerMetrics.TotalRequests
		totalSuccessful += providerMetrics.SuccessfulRequests
	}

	var globalSuccessRate float64
	if totalRequests > 0 {
		globalSuccessRate = float64(totalSuccessful) / float64(totalRequests) * 100
	}

	fmt.Printf("🌍 Global Statistics:\n")
	fmt.Printf("   📈 Total Requests: %d\n", totalRequests)
	fmt.Printf("   ✅ Successful: %d\n", totalSuccessful)
	fmt.Printf("   ❌ Failed: %d\n", totalRequests-totalSuccessful)
	fmt.Printf("   📊 Success Rate: %.2f%%\n", globalSuccessRate)

	fmt.Printf("\n📊 Provider Details:\n")
	for name, providerMetrics := range metrics {
		if providerMetrics.TotalRequests > 0 {
			fmt.Printf("   🏭 %s:\n", name)
			fmt.Printf("      📈 Total: %d | ✅ Success: %d | ❌ Failed: %d\n",
				providerMetrics.TotalRequests,
				providerMetrics.SuccessfulRequests,
				providerMetrics.FailedRequests)
			fmt.Printf("      📊 Success Rate: %.2f%% | 🏥 Status: %s\n",
				providerMetrics.UptimePercentage, providerMetrics.Status.String())

			if providerMetrics.AverageResponseTime > 0 {
				fmt.Printf("      ⏱️  Avg Response: %s\n", providerMetrics.AverageResponseTime)
			}
		}
	}

	// Test 5: Load Balancer Metrics
	fmt.Println("\n5️⃣ Load Balancer Metrics")
	fmt.Println("-------------------------")

	loadMetrics := loadBalancer.GetLoadMetrics(providers)
	for name, lbMetrics := range loadMetrics {
		fmt.Printf("⚖️  %s:\n", name)
		fmt.Printf("   🔄 Active Requests: %d\n", lbMetrics.ActiveRequests)
		fmt.Printf("   📈 Load Score: %.2f\n", lbMetrics.LoadScore)
		fmt.Printf("   🔧 Capacity: %d\n", lbMetrics.Capacity)
		fmt.Printf("   📊 Utilization: %.1f%%\n", lbMetrics.Utilization*100)
	}

	// Test 6: Dynamic Provider Management
	fmt.Println("\n6️⃣ Dynamic Provider Management")
	fmt.Println("------------------------------")

	// Add a temporary provider
	tempProvider := agent.ProviderConfig{
		Name:    "temp-high-perf",
		Type:    "openai",
		Model:   "gpt-4",
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Weight:  5.0, // Highest weight
		MaxConcurrency: 1,
		Timeout: 5 * time.Second,
	}

	fmt.Println("➕ Adding temporary high-performance provider...")
	err = mp.AddProvider(tempProvider)
	if err != nil {
		fmt.Printf("❌ Failed to add provider: %v\n", err)
	} else {
		fmt.Println("✅ Temporary provider added successfully")

		// Test with the new provider
		response, err := mp.Ask(ctx, "Test with temporary provider")
		if err != nil {
			fmt.Printf("❌ Temporary provider test failed: %v\n", err)
		} else {
			fmt.Printf("✅ Temporary provider response: %s\n", response)
		}

		// Remove the temporary provider
		fmt.Println("➖ Removing temporary provider...")
		err = mp.RemoveProvider("temp-high-perf")
		if err != nil {
			fmt.Printf("❌ Failed to remove provider: %v\n", err)
		} else {
			fmt.Println("✅ Temporary provider removed successfully")
		}
	}

	fmt.Println("\n🎉 Advanced MultiProvider Demo Completed!")
	fmt.Println("==================================")

	// Export metrics for external monitoring
	// Note: Metrics export is handled internally by MultiProvider
	fmt.Printf("📤 Metrics collection enabled for external monitoring\n")
}

// Helper functions for demo visualization

func strategyToString(strategy agent.SelectionStrategy) string {
	switch strategy {
	case agent.StrategyRoundRobin:
		return "Round Robin"
	case agent.StrategyWeightedRoundRobin:
		return "Weighted Round Robin"
	case agent.StrategyLeastConnections:
		return "Least Connections"
	case agent.StrategyFastestResponse:
		return "Fastest Response"
	case agent.StrategyRandom:
		return "Random"
	case agent.StrategyPriority:
		return "Priority"
	default:
		return "Unknown"
	}
}

func getStatusIcon(status agent.ProviderStatus) string {
	switch status {
	case agent.ProviderStatusHealthy:
		return "🟢"
	case agent.ProviderStatusDegraded:
		return "🟡"
	case agent.ProviderStatusUnhealthy:
		return "🔴"
	case agent.ProviderStatusDisabled:
		return "⚫"
	case agent.ProviderStatusUnknown:
		return "⚪"
	default:
		return "❓"
	}
}

func getCircuitBreakerIcon(state agent.CircuitBreakerState) string {
	switch state {
	case agent.CircuitBreakerClosed:
		return "🟢"
	case agent.CircuitBreakerOpen:
		return "🔴"
	case agent.CircuitBreakerHalfOpen:
		return "🟡"
	default:
		return "❓"
	}
}

// Mock adapter for demonstration
type mockAdapter struct {
	responses []string
	index     int
}

func (m *mockAdapter) Complete(ctx context.Context, req *agent.CompletionRequest) (*agent.CompletionResponse, error) {
	response := &agent.CompletionResponse{
		Content: m.responses[m.index%len(m.responses)],
		Usage: agent.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		FinishReason: "stop",
		Model:       req.Model,
		Created:     time.Now().Unix(),
	}

	m.index++
	return response, nil
}

func (m *mockAdapter) Stream(ctx context.Context, req *agent.CompletionRequest, onChunk func(string)) (*agent.CompletionResponse, error) {
	content := m.responses[m.index%len(m.responses)]
	m.index++

	// Simulate streaming by calling the callback multiple times
	if onChunk != nil {
		for i, char := range content {
			onChunk(string(char))
			time.Sleep(10 * time.Millisecond)
		}
	}

	return &agent.CompletionResponse{
		Content: content,
		Usage: agent.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: int64(len(content)),
			TotalTokens:      10 + int64(len(content)),
		},
		FinishReason: "stop",
		Model:       req.Model,
		Created:     time.Now().Unix(),
	}, nil
}

func (m *mockAdapter) Close() error {
	return nil
}