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
	fmt.Println("🚀 Planning Layer - Parallel Execution Example")
	fmt.Println("================================================\n")

	// Check for API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable not set")
	}

	// Create agent configuration
	config := agent.Config{
		Provider: "openai",
		Model:    "gpt-4",
		APIKey:   apiKey,
	}

	// Create agent
	agentInstance, err := agent.NewAgent(config)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Println("✅ Agent initialized with GPT-4\n")

	// Example 1: Parallel batch processing
	fmt.Println("📋 Example 1: Parallel Batch Processing")
	fmt.Println("----------------------------------------")
	runBatchProcessingExample(agentInstance)

	// Example 2: Dependency-aware parallel execution
	fmt.Println("\n📋 Example 2: Dependency-Aware Parallel Execution")
	fmt.Println("--------------------------------------------------")
	runDependencyParallelExample(agentInstance)

	// Example 3: Performance comparison
	fmt.Println("\n📋 Example 3: Sequential vs Parallel Performance")
	fmt.Println("-------------------------------------------------")
	runPerformanceComparisonExample(agentInstance)
}

// runBatchProcessingExample demonstrates processing 10 independent tasks in parallel
func runBatchProcessingExample(agentInstance *agent.Agent) {
	// Create a plan to analyze 10 tech companies in parallel
	plan := agent.NewPlan(
		"Analyze top 10 tech companies for investment",
		agent.StrategyParallel,
	)

	companies := []string{
		"Apple", "Microsoft", "Google", "Amazon", "Meta",
		"Tesla", "NVIDIA", "Netflix", "Adobe", "Salesforce",
	}

	for _, company := range companies {
		plan.AddTask(agent.Task{
			ID:          fmt.Sprintf("analyze-%s", company),
			Description: fmt.Sprintf("Analyze %s: market cap, revenue growth, key products", company),
			Type:        agent.TaskTypeObservation,
		})
	}

	// Configure parallel execution with max 5 concurrent tasks
	config := agent.DefaultPlannerConfig()
	config.MaxParallel = 5
	config.Strategy = agent.StrategyParallel

	executor := agent.NewExecutor(config, agentInstance)

	fmt.Printf("🔄 Processing %d companies with MaxParallel=%d...\n", len(companies), config.MaxParallel)

	start := time.Now()
	result, err := executor.Execute(context.Background(), plan)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Execution failed: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Completed in %v\n", duration)
	fmt.Printf("📊 Metrics:\n")
	fmt.Printf("   - Tasks: %d\n", result.Metrics.TaskCount)
	fmt.Printf("   - Success Rate: %.1f%%\n", result.Metrics.SuccessRate*100)
	fmt.Printf("   - Avg Task Duration: %v\n", result.Metrics.AvgTaskDuration)
	fmt.Printf("   - Throughput: %.1f tasks/sec\n", float64(result.Metrics.TaskCount)/duration.Seconds())

	// Show sample results
	fmt.Println("\n📝 Sample Analysis Results:")
	for i, task := range plan.Tasks[:3] {
		if task.Result != nil {
			fmt.Printf("\n%d. %s:\n", i+1, task.Description)
			if chatResult, ok := task.Result.(*agent.ChatResult); ok {
				// Truncate long responses
				content := chatResult.Content
				if len(content) > 200 {
					content = content[:200] + "..."
				}
				fmt.Printf("   %s\n", content)
			}
		}
	}
}

// runDependencyParallelExample demonstrates parallel execution with dependencies
func runDependencyParallelExample(agentInstance *agent.Agent) {
	// Create a plan with dependency structure:
	// Research (root) -> Analysis A, B, C (parallel) -> Final Report (aggregation)
	plan := agent.NewPlan(
		"Market research report with parallel analysis",
		agent.StrategyParallel,
	)

	plan.AddTask(agent.Task{
		ID:          "research",
		Description: "Gather market data for AI industry in 2024",
		Type:        agent.TaskTypeObservation,
	})

	plan.AddTask(agent.Task{
		ID:           "analyze-tech",
		Description:  "Analyze technology trends from research data",
		Type:         agent.TaskTypeAction,
		Dependencies: []string{"research"},
	})

	plan.AddTask(agent.Task{
		ID:           "analyze-market",
		Description:  "Analyze market size and growth from research data",
		Type:         agent.TaskTypeAction,
		Dependencies: []string{"research"},
	})

	plan.AddTask(agent.Task{
		ID:           "analyze-competition",
		Description:  "Analyze competitive landscape from research data",
		Type:         agent.TaskTypeAction,
		Dependencies: []string{"research"},
	})

	plan.AddTask(agent.Task{
		ID:           "final-report",
		Description:  "Synthesize tech, market, and competition analysis into executive summary",
		Type:         agent.TaskTypeAggregate,
		Dependencies: []string{"analyze-tech", "analyze-market", "analyze-competition"},
	})

	config := agent.DefaultPlannerConfig()
	config.MaxParallel = 3
	config.Strategy = agent.StrategyParallel

	executor := agent.NewExecutor(config, agentInstance)

	fmt.Println("🔄 Executing dependency-aware parallel plan...")
	fmt.Println("   Structure: Research → [3 parallel analyses] → Final Report")

	start := time.Now()
	result, err := executor.Execute(context.Background(), plan)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Execution failed: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Completed in %v\n", duration)
	fmt.Printf("📊 Timeline Events: %d\n", len(result.Timeline))

	// Show execution timeline
	fmt.Println("\n📅 Execution Timeline:")
	for _, event := range result.Timeline {
		fmt.Printf("   [%v] %s: %s\n",
			event.Timestamp.Format("15:04:05.000"),
			event.Type,
			event.Description,
		)
	}
}

// runPerformanceComparisonExample compares sequential vs parallel execution
func runPerformanceComparisonExample(agentInstance *agent.Agent) {
	taskCount := 8
	tasks := []agent.Task{}

	for i := 1; i <= taskCount; i++ {
		tasks = append(tasks, agent.Task{
			ID:          fmt.Sprintf("task-%d", i),
			Description: fmt.Sprintf("Summarize key points about AI trend #%d", i),
			Type:        agent.TaskTypeObservation,
		})
	}

	// Test 1: Sequential execution
	planSeq := agent.NewPlan("Sequential test", agent.StrategySequential)
	for _, task := range tasks {
		planSeq.AddTask(task)
	}

	configSeq := agent.DefaultPlannerConfig()
	configSeq.Strategy = agent.StrategySequential
	executorSeq := agent.NewExecutor(configSeq, agentInstance)

	fmt.Printf("🔄 Running %d tasks sequentially...\n", taskCount)
	startSeq := time.Now()
	resultSeq, err := executorSeq.Execute(context.Background(), planSeq)
	durationSeq := time.Since(startSeq)

	if err != nil {
		log.Printf("❌ Sequential execution failed: %v\n", err)
		return
	}

	// Test 2: Parallel execution
	planPar := agent.NewPlan("Parallel test", agent.StrategyParallel)
	for _, task := range tasks {
		planPar.AddTask(task)
	}

	configPar := agent.DefaultPlannerConfig()
	configPar.Strategy = agent.StrategyParallel
	configPar.MaxParallel = 4
	executorPar := agent.NewExecutor(configPar, agentInstance)

	fmt.Printf("🔄 Running %d tasks in parallel (MaxParallel=4)...\n", taskCount)
	startPar := time.Now()
	resultPar, err := executorPar.Execute(context.Background(), planPar)
	durationPar := time.Since(startPar)

	if err != nil {
		log.Printf("❌ Parallel execution failed: %v\n", err)
		return
	}

	// Compare results
	fmt.Println("\n📊 Performance Comparison:")
	fmt.Printf("   Sequential: %v (%.1f tasks/sec)\n",
		durationSeq,
		float64(taskCount)/durationSeq.Seconds(),
	)
	fmt.Printf("   Parallel:   %v (%.1f tasks/sec)\n",
		durationPar,
		float64(taskCount)/durationPar.Seconds(),
	)

	speedup := float64(durationSeq) / float64(durationPar)
	fmt.Printf("\n⚡ Speedup: %.2fx faster\n", speedup)

	fmt.Printf("\n📈 Metrics Comparison:\n")
	fmt.Printf("   Sequential - Avg Duration: %v\n", resultSeq.Metrics.AvgTaskDuration)
	fmt.Printf("   Parallel   - Avg Duration: %v\n", resultPar.Metrics.AvgTaskDuration)
}
