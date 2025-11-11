package agent

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Benchmarks for Planning Layer performance measurement

func BenchmarkExecuteSequential(b *testing.B) {
	benchmarkExecutor(b, StrategySequential, 5)
}

func BenchmarkExecuteSequential10Tasks(b *testing.B) {
	benchmarkExecutor(b, StrategySequential, 10)
}

func BenchmarkExecuteSequential20Tasks(b *testing.B) {
	benchmarkExecutor(b, StrategySequential, 20)
}

func BenchmarkExecuteParallel(b *testing.B) {
	benchmarkExecutor(b, StrategyParallel, 5)
}

func BenchmarkExecuteParallel10Tasks(b *testing.B) {
	benchmarkExecutor(b, StrategyParallel, 10)
}

func BenchmarkExecuteParallel20Tasks(b *testing.B) {
	benchmarkExecutor(b, StrategyParallel, 20)
}

func BenchmarkExecuteAdaptive(b *testing.B) {
	benchmarkExecutor(b, StrategyAdaptive, 5)
}

func BenchmarkExecuteAdaptive10Tasks(b *testing.B) {
	benchmarkExecutor(b, StrategyAdaptive, 10)
}

func BenchmarkExecuteAdaptive20Tasks(b *testing.B) {
	benchmarkExecutor(b, StrategyAdaptive, 20)
}

func benchmarkExecutor(b *testing.B, strategy PlanningStrategy, taskCount int) {
	// Create plan with independent tasks
	tasks := make([]Task, taskCount)
	for i := 0; i < taskCount; i++ {
		tasks[i] = Task{
			ID:          fmt.Sprintf("task-%d", i+1),
			Description: fmt.Sprintf("Execute task %d", i+1),
			Type:        TaskTypeAction,
		}
	}

	plan := &Plan{
		ID:       fmt.Sprintf("benchmark-%s-%d", strategy, taskCount),
		Goal:     "Performance testing",
		Strategy: strategy,
		Tasks:    tasks,
	}

	mock := &mockAgent{
		chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
			// Simulate realistic task execution time (5ms)
			time.Sleep(5 * time.Millisecond)
			return &ChatResult{Content: "done"}, nil
		},
	}

	config := DefaultPlannerConfig()
	config.MaxParallel = 5
	executor := NewExecutor(config, mock)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(context.Background(), plan)
		if err != nil {
			b.Fatal(err)
		}

		// Reset task statuses for next iteration
		for j := range plan.Tasks {
			plan.Tasks[j].Status = TaskStatusPending
			plan.Tasks[j].Result = nil
		}
	}
}

func BenchmarkGoalChecking(b *testing.B) {
	plan := &Plan{
		ID:   "goal-check-bench",
		Goal: "Measure goal checking overhead",
		GoalState: GoalState{
			Description: "Complete 5 tasks",
			Criteria: []GoalCriterion{
				{
					Name:      "tasks_completed",
					Expected:  "5",
					Operator:  ">=",
					Satisfied: false,
				},
			},
		},
		Tasks: []Task{
			{ID: "1", Description: "Task 1"},
			{ID: "2", Description: "Task 2"},
			{ID: "3", Description: "Task 3"},
			{ID: "4", Description: "Task 4"},
			{ID: "5", Description: "Task 5"},
		},
	}

	mock := &mockAgent{
		chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
			return &ChatResult{Content: "done"}, nil
		},
	}

	config := DefaultPlannerConfig()
	config.GoalCheckInterval = 1 // Check after every task
	executor := NewExecutor(config, mock)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(context.Background(), plan)
		if err != nil {
			b.Fatal(err)
		}

		// Reset for next iteration
		for j := range plan.Tasks {
			plan.Tasks[j].Status = TaskStatusPending
			plan.Tasks[j].Result = nil
		}
		plan.GoalState.Criteria[0].Satisfied = false
	}
}

func BenchmarkTopologicalSort(b *testing.B) {
	// Create complex dependency graph: 20 tasks
	tasks := []Task{
		{ID: "1", Description: "Root"},
		{ID: "2", Description: "L1-A", Dependencies: []string{"1"}},
		{ID: "3", Description: "L1-B", Dependencies: []string{"1"}},
		{ID: "4", Description: "L2-A", Dependencies: []string{"2"}},
		{ID: "5", Description: "L2-B", Dependencies: []string{"2", "3"}},
		{ID: "6", Description: "L2-C", Dependencies: []string{"3"}},
		{ID: "7", Description: "L3-A", Dependencies: []string{"4", "5"}},
		{ID: "8", Description: "L3-B", Dependencies: []string{"5", "6"}},
		{ID: "9", Description: "L4", Dependencies: []string{"7", "8"}},
		{ID: "10", Description: "L1-C", Dependencies: []string{"1"}},
		{ID: "11", Description: "L2-D", Dependencies: []string{"10"}},
		{ID: "12", Description: "L3-C", Dependencies: []string{"11"}},
		{ID: "13", Description: "L4-B", Dependencies: []string{"12", "9"}},
		{ID: "14", Description: "L1-D", Dependencies: []string{"1"}},
		{ID: "15", Description: "L2-E", Dependencies: []string{"14"}},
		{ID: "16", Description: "L3-D", Dependencies: []string{"15"}},
		{ID: "17", Description: "L4-C", Dependencies: []string{"16"}},
		{ID: "18", Description: "L5", Dependencies: []string{"13", "17"}},
		{ID: "19", Description: "L6-A", Dependencies: []string{"18"}},
		{ID: "20", Description: "L6-B", Dependencies: []string{"18"}},
	}

	config := DefaultPlannerConfig()
	executor := NewExecutor(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.topologicalSort(tasks)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGroupByDependencyLevel(b *testing.B) {
	// Same complex graph as topological sort
	tasks := []Task{
		{ID: "1", Description: "Root"},
		{ID: "2", Description: "L1-A", Dependencies: []string{"1"}},
		{ID: "3", Description: "L1-B", Dependencies: []string{"1"}},
		{ID: "4", Description: "L2-A", Dependencies: []string{"2"}},
		{ID: "5", Description: "L2-B", Dependencies: []string{"2", "3"}},
		{ID: "6", Description: "L2-C", Dependencies: []string{"3"}},
		{ID: "7", Description: "L3-A", Dependencies: []string{"4", "5"}},
		{ID: "8", Description: "L3-B", Dependencies: []string{"5", "6"}},
		{ID: "9", Description: "L4", Dependencies: []string{"7", "8"}},
		{ID: "10", Description: "L1-C", Dependencies: []string{"1"}},
		{ID: "11", Description: "L2-D", Dependencies: []string{"10"}},
		{ID: "12", Description: "L3-C", Dependencies: []string{"11"}},
		{ID: "13", Description: "L4-B", Dependencies: []string{"12", "9"}},
		{ID: "14", Description: "L1-D", Dependencies: []string{"1"}},
		{ID: "15", Description: "L2-E", Dependencies: []string{"14"}},
		{ID: "16", Description: "L3-D", Dependencies: []string{"15"}},
		{ID: "17", Description: "L4-C", Dependencies: []string{"16"}},
		{ID: "18", Description: "L5", Dependencies: []string{"13", "17"}},
		{ID: "19", Description: "L6-A", Dependencies: []string{"18"}},
		{ID: "20", Description: "L6-B", Dependencies: []string{"18"}},
	}

	config := DefaultPlannerConfig()
	executor := NewExecutor(config, nil)

	// Pre-sort for grouping
	sorted, _ := executor.topologicalSort(tasks)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.groupByDependencyLevel(sorted)
	}
}

func BenchmarkParallelVsSequential(b *testing.B) {
	// Compare parallel vs sequential for 20 independent tasks
	taskCount := 20

	b.Run("Sequential", func(b *testing.B) {
		benchmarkExecutor(b, StrategySequential, taskCount)
	})

	b.Run("Parallel-MaxParallel5", func(b *testing.B) {
		benchmarkExecutor(b, StrategyParallel, taskCount)
	})

	b.Run("Parallel-MaxParallel10", func(b *testing.B) {
		tasks := make([]Task, taskCount)
		for i := 0; i < taskCount; i++ {
			tasks[i] = Task{
				ID:          fmt.Sprintf("task-%d", i+1),
				Description: fmt.Sprintf("Execute task %d", i+1),
				Type:        TaskTypeAction,
			}
		}

		plan := &Plan{
			ID:       "parallel-bench-10",
			Goal:     "Performance testing",
			Strategy: StrategyParallel,
			Tasks:    tasks,
		}

		mock := &mockAgent{
			chatFunc: func(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
				time.Sleep(5 * time.Millisecond)
				return &ChatResult{Content: "done"}, nil
			},
		}

		config := DefaultPlannerConfig()
		config.MaxParallel = 10
		executor := NewExecutor(config, mock)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := executor.Execute(context.Background(), plan)
			if err != nil {
				b.Fatal(err)
			}

			for j := range plan.Tasks {
				plan.Tasks[j].Status = TaskStatusPending
				plan.Tasks[j].Result = nil
			}
		}
	})
}
