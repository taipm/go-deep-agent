package agent

import (
	"testing"
	"time"
)

// TestDefaultPlannerConfig verifies that default configuration has correct values.
func TestDefaultPlannerConfig(t *testing.T) {
	config := DefaultPlannerConfig()

	if config == nil {
		t.Fatal("DefaultPlannerConfig returned nil")
	}

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"MaxDepth", config.MaxDepth, 3},
		{"MaxSubtasks", config.MaxSubtasks, 5},
		{"MinSubtaskSplit", config.MinSubtaskSplit, 2},
		{"Strategy", config.Strategy, StrategySequential},
		{"MaxParallel", config.MaxParallel, 3},
		{"AdaptiveThreshold", config.AdaptiveThreshold, 0.7},
		{"GoalCheckInterval", config.GoalCheckInterval, 5},
		{"GoalTimeout", config.GoalTimeout, 5 * time.Minute},
		{"ReActEnabled", config.ReActEnabled, true},
		{"FallbackToReAct", config.FallbackToReAct, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestPlannerConfigValidate_ValidConfig tests that valid configs pass validation.
func TestPlannerConfigValidate_ValidConfig(t *testing.T) {
	config := DefaultPlannerConfig()

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate() failed for default config: %v", err)
	}
}

// TestPlannerConfigValidate_InvalidMaxDepth tests MaxDepth validation.
func TestPlannerConfigValidate_InvalidMaxDepth(t *testing.T) {
	config := DefaultPlannerConfig()
	config.MaxDepth = 0

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail for MaxDepth = 0")
	}

	config.MaxDepth = -1
	err = config.Validate()
	if err == nil {
		t.Error("Validate() should fail for MaxDepth < 0")
	}
}

// TestPlannerConfigValidate_InvalidMaxSubtasks tests MaxSubtasks validation.
func TestPlannerConfigValidate_InvalidMaxSubtasks(t *testing.T) {
	config := DefaultPlannerConfig()
	config.MaxSubtasks = 0

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail for MaxSubtasks = 0")
	}

	config.MaxSubtasks = -1
	err = config.Validate()
	if err == nil {
		t.Error("Validate() should fail for MaxSubtasks < 0")
	}
}

// TestPlannerConfigValidate_InvalidMinSubtaskSplit tests MinSubtaskSplit validation.
func TestPlannerConfigValidate_InvalidMinSubtaskSplit(t *testing.T) {
	config := DefaultPlannerConfig()
	config.MinSubtaskSplit = 1

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail for MinSubtaskSplit < 2")
	}

	config.MinSubtaskSplit = 0
	err = config.Validate()
	if err == nil {
		t.Error("Validate() should fail for MinSubtaskSplit = 0")
	}
}

// TestPlannerConfigValidate_InvalidMaxParallel tests MaxParallel validation.
func TestPlannerConfigValidate_InvalidMaxParallel(t *testing.T) {
	config := DefaultPlannerConfig()
	config.MaxParallel = 0

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail for MaxParallel = 0")
	}

	config.MaxParallel = -1
	err = config.Validate()
	if err == nil {
		t.Error("Validate() should fail for MaxParallel < 0")
	}
}

// TestPlannerConfigValidate_InvalidAdaptiveThreshold tests AdaptiveThreshold validation.
func TestPlannerConfigValidate_InvalidAdaptiveThreshold(t *testing.T) {
	tests := []struct {
		name      string
		threshold float64
		shouldErr bool
	}{
		{"Valid 0.0", 0.0, false},
		{"Valid 0.5", 0.5, false},
		{"Valid 1.0", 1.0, false},
		{"Invalid -0.1", -0.1, true},
		{"Invalid 1.1", 1.1, true},
		{"Invalid -1.0", -1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultPlannerConfig()
			config.AdaptiveThreshold = tt.threshold

			err := config.Validate()
			if tt.shouldErr && err == nil {
				t.Errorf("Validate() should fail for AdaptiveThreshold = %v", tt.threshold)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Validate() should pass for AdaptiveThreshold = %v, got error: %v", tt.threshold, err)
			}
		})
	}
}

// TestPlannerConfigValidate_InvalidGoalCheckInterval tests GoalCheckInterval validation.
func TestPlannerConfigValidate_InvalidGoalCheckInterval(t *testing.T) {
	config := DefaultPlannerConfig()
	config.GoalCheckInterval = 0

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail for GoalCheckInterval = 0")
	}

	config.GoalCheckInterval = -1
	err = config.Validate()
	if err == nil {
		t.Error("Validate() should fail for GoalCheckInterval < 0")
	}
}

// TestPlannerConfigValidate_InvalidGoalTimeout tests GoalTimeout validation.
func TestPlannerConfigValidate_InvalidGoalTimeout(t *testing.T) {
	config := DefaultPlannerConfig()
	config.GoalTimeout = 0

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail for GoalTimeout = 0")
	}

	config.GoalTimeout = -1 * time.Second
	err = config.Validate()
	if err == nil {
		t.Error("Validate() should fail for GoalTimeout < 0")
	}
}

// TestPlannerConfigValidate_InvalidStrategy tests Strategy validation.
func TestPlannerConfigValidate_InvalidStrategy(t *testing.T) {
	config := DefaultPlannerConfig()
	config.Strategy = "invalid"

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should fail for invalid strategy")
	}
}

// TestPlannerConfigValidate_AllValidStrategies tests all valid strategy values.
func TestPlannerConfigValidate_AllValidStrategies(t *testing.T) {
	strategies := []PlanningStrategy{
		StrategySequential,
		StrategyParallel,
		StrategyAdaptive,
	}

	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			config := DefaultPlannerConfig()
			config.Strategy = strategy

			err := config.Validate()
			if err != nil {
				t.Errorf("Validate() should pass for strategy %v, got error: %v", strategy, err)
			}
		})
	}
}

// TestPlannerConfigValidate_CustomValidConfig tests a custom valid configuration.
func TestPlannerConfigValidate_CustomValidConfig(t *testing.T) {
	config := &PlannerConfig{
		MaxDepth:          5,
		MaxSubtasks:       10,
		MinSubtaskSplit:   3,
		Strategy:          StrategyParallel,
		MaxParallel:       5,
		AdaptiveThreshold: 0.8,
		GoalCheckInterval: 10,
		GoalTimeout:       10 * time.Minute,
		ReActEnabled:      false,
		FallbackToReAct:   false,
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate() failed for custom valid config: %v", err)
	}
}
