package agent

import (
	"errors"
	"time"
)

// PlannerConfig controls planning behavior and execution strategies.
type PlannerConfig struct {
	// MaxDepth is the maximum task nesting level (default: 3).
	MaxDepth int

	// MaxSubtasks is the maximum number of subtasks per level (default: 5).
	MaxSubtasks int

	// MinSubtaskSplit is the minimum number of subtasks to decompose (default: 2).
	MinSubtaskSplit int

	// Strategy determines how tasks are executed (default: sequential).
	Strategy PlanningStrategy

	// MaxParallel is the maximum number of concurrent subtasks (default: 3).
	MaxParallel int

	// AdaptiveThreshold is the success rate threshold for switching strategy (default: 0.7).
	AdaptiveThreshold float64

	// GoalCheckInterval is the number of steps between goal state checks (default: 5).
	GoalCheckInterval int

	// GoalTimeout is the maximum time allowed per goal (default: 5 minutes).
	GoalTimeout time.Duration

	// ReActEnabled determines if ReAct pattern is used for task execution (default: true).
	ReActEnabled bool

	// FallbackToReAct allows falling back to ReAct if planning fails (default: true).
	FallbackToReAct bool
}

// DefaultPlannerConfig returns a configuration with sensible defaults.
func DefaultPlannerConfig() *PlannerConfig {
	return &PlannerConfig{
		MaxDepth:          3,
		MaxSubtasks:       5,
		MinSubtaskSplit:   2,
		Strategy:          StrategySequential,
		MaxParallel:       3,
		AdaptiveThreshold: 0.7,
		GoalCheckInterval: 5,
		GoalTimeout:       5 * time.Minute,
		ReActEnabled:      true,
		FallbackToReAct:   true,
	}
}

// Validate checks if the configuration is valid and returns an error if not.
func (c *PlannerConfig) Validate() error {
	if c.MaxDepth <= 0 {
		return errors.New("MaxDepth must be greater than 0")
	}
	if c.MaxSubtasks <= 0 {
		return errors.New("MaxSubtasks must be greater than 0")
	}
	if c.MinSubtaskSplit < 2 {
		return errors.New("MinSubtaskSplit must be at least 2")
	}
	if c.MaxParallel <= 0 {
		return errors.New("MaxParallel must be greater than 0")
	}
	if c.AdaptiveThreshold < 0.0 || c.AdaptiveThreshold > 1.0 {
		return errors.New("AdaptiveThreshold must be between 0.0 and 1.0")
	}
	if c.GoalCheckInterval <= 0 {
		return errors.New("GoalCheckInterval must be greater than 0")
	}
	if c.GoalTimeout <= 0 {
		return errors.New("GoalTimeout must be greater than 0")
	}

	// Validate strategy
	switch c.Strategy {
	case StrategySequential, StrategyParallel, StrategyAdaptive:
		// Valid
	default:
		return errors.New("Strategy must be sequential, parallel, or adaptive")
	}

	return nil
}
