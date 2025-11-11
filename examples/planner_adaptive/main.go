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
	fmt.Println("🧠 Planning Layer - Adaptive Strategy Example")
	fmt.Println("==============================================\n")

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

	// Example 1: Adaptive strategy with mixed workload
	fmt.Println("📋 Example 1: Adaptive Strategy with Mixed Workload")
	fmt.Println("----------------------------------------------------")
	runMixedWorkloadExample(agentInstance)

	// Example 2: Strategy switching based on performance
	fmt.Println("\n📋 Example 2: Dynamic Strategy Switching")
	fmt.Println("----------------------------------------")
	runDynamicSwitchingExample(agentInstance)

	// Example 3: Multi-phase pipeline with adaptive execution
	fmt.Println("\n📋 Example 3: Multi-Phase Pipeline")
	fmt.Println("-----------------------------------")
	runMultiPhasePipelineExample(agentInstance)
}

// runMixedWorkloadExample demonstrates adaptive strategy with varying task complexity
func runMixedWorkloadExample(agentInstance *agent.Agent) {
	plan := agent.NewPlan(
		"Research and analysis with adaptive execution",
		agent.StrategyAdaptive,
	)

	// Phase 1: Simple data gathering (good for parallel)
	for i := 1; i <= 5; i++ {
		plan.AddTask(agent.Task{
			ID:          fmt.Sprintf("gather-%d", i),
			Description: fmt.Sprintf("Gather basic facts about AI trend #%d", i),
			Type:        agent.TaskTypeObservation,
		})
	}

	// Phase 2: Complex analysis (may benefit from sequential)
	plan.AddTask(agent.Task{
		ID:           "deep-analysis",
		Description:  "Perform deep analysis of all gathered data: correlations, patterns, insights",
		Type:         agent.TaskTypeAction,
		Dependencies: []string{"gather-1", "gather-2", "gather-3", "gather-4", "gather-5"},
	})

	// Phase 3: More parallel tasks
	for i := 1; i <= 3; i++ {
		plan.AddTask(agent.Task{
			ID:           fmt.Sprintf("report-%d", i),
			Description:  fmt.Sprintf("Generate report section #%d from analysis", i),
			Type:         agent.TaskTypeAction,
			Dependencies: []string{"deep-analysis"},
		})
	}

	// Configure adaptive execution
	config := agent.DefaultPlannerConfig()
	config.Strategy = agent.StrategyAdaptive
	config.MaxParallel = 5
	config.AdaptiveThreshold = 0.6 // Switch if parallel efficiency < 60%

	executor := agent.NewExecutor(config, agentInstance)

	fmt.Println("🔄 Executing with adaptive strategy...")
	fmt.Printf("   Config: MaxParallel=%d, AdaptiveThreshold=%.1f\n",
		config.MaxParallel, config.AdaptiveThreshold)

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

	// Analyze strategy switches in timeline
	fmt.Println("\n🔀 Strategy Timeline:")
	strategyEvents := 0
	for _, event := range result.Timeline {
		if event.Type == "strategy_initialized" || event.Type == "strategy_switched" {
			fmt.Printf("   [%v] %s\n",
				event.Timestamp.Format("15:04:05.000"),
				event.Description,
			)
			strategyEvents++
		}
	}

	if strategyEvents == 0 {
		fmt.Println("   No strategy switches (single strategy used throughout)")
	}
}

// runDynamicSwitchingExample demonstrates when and why strategy switches occur
func runDynamicSwitchingExample(agentInstance *agent.Agent) {
	plan := agent.NewPlan(
		"Demonstrate strategy switching triggers",
		agent.StrategyAdaptive,
	)

	// Create tasks that will trigger strategy evaluation
	// Start with many simple parallel tasks
	fmt.Println("Creating workload designed to trigger strategy switches...")

	for i := 1; i <= 8; i++ {
		plan.AddTask(agent.Task{
			ID:          fmt.Sprintf("batch-1-task-%d", i),
			Description: fmt.Sprintf("Quick task %d: summarize a short text", i),
			Type:        agent.TaskTypeObservation,
		})
	}

	// Add a complex sequential dependency chain
	plan.AddTask(agent.Task{
		ID:           "complex-1",
		Description:  "Complex analysis step 1: requires detailed reasoning",
		Type:         agent.TaskTypeAction,
		Dependencies: []string{"batch-1-task-1"},
	})

	plan.AddTask(agent.Task{
		ID:           "complex-2",
		Description:  "Complex analysis step 2: builds on step 1",
		Type:         agent.TaskTypeAction,
		Dependencies: []string{"complex-1"},
	})

	// More parallel tasks
	for i := 1; i <= 6; i++ {
		plan.AddTask(agent.Task{
			ID:           fmt.Sprintf("batch-2-task-%d", i),
			Description:  fmt.Sprintf("Another quick task %d", i),
			Type:         agent.TaskTypeObservation,
			Dependencies: []string{"complex-2"},
		})
	}

	config := agent.DefaultPlannerConfig()
	config.Strategy = agent.StrategyAdaptive
	config.MaxParallel = 4
	config.AdaptiveThreshold = 0.5

	executor := agent.NewExecutor(config, agentInstance)

	fmt.Printf("\n🔄 Starting adaptive execution (threshold=%.1f)...\n", config.AdaptiveThreshold)

	start := time.Now()
	result, err := executor.Execute(context.Background(), plan)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Execution failed: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Completed in %v\n", duration)

	// Show detailed timeline with strategy events
	fmt.Println("\n📅 Detailed Execution Timeline:")
	for i, event := range result.Timeline {
		if i < 20 { // Show first 20 events
			fmt.Printf("   [%02d] %v - %s: %s\n",
				i+1,
				event.Timestamp.Format("15:04:05.000"),
				event.Type,
				event.Description,
			)
		}
	}

	if len(result.Timeline) > 20 {
		fmt.Printf("   ... and %d more events\n", len(result.Timeline)-20)
	}
}

// runMultiPhasePipelineExample shows adaptive strategy in a real-world pipeline
func runMultiPhasePipelineExample(agentInstance *agent.Agent) {
	plan := agent.NewPlan(
		"Multi-phase content generation pipeline",
		agent.StrategyAdaptive,
	)

	// Phase 1: Research (parallel)
	topics := []string{"AI Ethics", "AI Safety", "AI Regulations"}
	for i, topic := range topics {
		plan.AddTask(agent.Task{
			ID:          fmt.Sprintf("research-%d", i+1),
			Description: fmt.Sprintf("Research current state of %s", topic),
			Type:        agent.TaskTypeObservation,
		})
	}

	// Phase 2: Synthesis (sequential - needs all research)
	plan.AddTask(agent.Task{
		ID:           "synthesize",
		Description:  "Synthesize research findings into coherent narrative",
		Type:         agent.TaskTypeAggregate,
		Dependencies: []string{"research-1", "research-2", "research-3"},
	})

	// Phase 3: Content generation (parallel - independent pieces)
	sections := []string{"Introduction", "Main Body", "Conclusion", "References"}
	for i, section := range sections {
		plan.AddTask(agent.Task{
			ID:           fmt.Sprintf("write-%d", i+1),
			Description:  fmt.Sprintf("Write %s section based on synthesis", section),
			Type:         agent.TaskTypeAction,
			Dependencies: []string{"synthesize"},
		})
	}

	// Phase 4: Quality check (sequential - final review)
	plan.AddTask(agent.Task{
		ID:           "review",
		Description:  "Review and edit all sections for consistency and quality",
		Type:         agent.TaskTypeDecision,
		Dependencies: []string{"write-1", "write-2", "write-3", "write-4"},
	})

	config := agent.DefaultPlannerConfig()
	config.Strategy = agent.StrategyAdaptive
	config.MaxParallel = 3
	config.AdaptiveThreshold = 0.6

	executor := agent.NewExecutor(config, agentInstance)

	fmt.Println("🔄 Running multi-phase pipeline with adaptive strategy...")
	fmt.Println("   Pipeline: Research (||) → Synthesize (→) → Write (||) → Review (→)")

	start := time.Now()
	result, err := executor.Execute(context.Background(), plan)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Execution failed: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Pipeline completed in %v\n", duration)
	fmt.Printf("📊 Metrics:\n")
	fmt.Printf("   - Total Tasks: %d\n", result.Metrics.TaskCount)
	fmt.Printf("   - Success Rate: %.1f%%\n", result.Metrics.SuccessRate*100)
	fmt.Printf("   - Execution Time: %v\n", result.Metrics.ExecutionTime)
	fmt.Printf("   - Avg Task Duration: %v\n", result.Metrics.AvgTaskDuration)

	// Analyze phase performance
	fmt.Println("\n📈 Phase Analysis:")
	phases := map[string][]string{
		"Research":  {"research-1", "research-2", "research-3"},
		"Synthesis": {"synthesize"},
		"Writing":   {"write-1", "write-2", "write-3", "write-4"},
		"Review":    {"review"},
	}

	for phaseName, taskIDs := range phases {
		phaseTime := time.Duration(0)
		for _, taskID := range taskIDs {
			for _, task := range plan.Tasks {
				if task.ID == taskID && task.Status == agent.TaskStatusCompleted {
					// Estimate from timeline events
					for i, event := range result.Timeline {
						if event.Type == "task_started" && event.Description == fmt.Sprintf("Starting task: %s", task.Description) {
							if i+1 < len(result.Timeline) {
								nextEvent := result.Timeline[i+1]
								if nextEvent.Type == "task_completed" {
									phaseTime += nextEvent.Timestamp.Sub(event.Timestamp)
								}
							}
						}
					}
				}
			}
		}
		fmt.Printf("   %s: %d tasks, ~%v\n", phaseName, len(taskIDs), phaseTime)
	}
}
