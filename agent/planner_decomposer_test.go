package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockLLMGenerator is a mock implementation of llmGenerator for testing.
type mockLLMGenerator struct {
	response string
	err      error
}

func (m *mockLLMGenerator) Generate(ctx context.Context, prompt string, opts *ChatOptions) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

// Test analyzeComplexity with different goal types
func TestAnalyzeComplexity(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	tests := []struct {
		name     string
		goal     string
		minScore int
		maxScore int
	}{
		{
			name:     "Simple goal",
			goal:     "Calculate 2 + 2",
			minScore: 0,
			maxScore: 2,
		},
		{
			name:     "Medium complexity",
			goal:     "Research AI trends and summarize findings",
			minScore: 2,
			maxScore: 5,
		},
		{
			name:     "Complex goal with multiple steps",
			goal:     "Analyze competitors A, B, C, then compare features, pricing, and market position, and recommend best strategy",
			minScore: 5,
			maxScore: 15,
		},
		{
			name:     "Goal with list",
			goal:     "Check stock prices for AAPL, GOOGL, MSFT, TSLA, AMZN",
			minScore: 4,
			maxScore: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := decomposer.analyzeComplexity(tt.goal)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("analyzeComplexity() = %d, want between %d and %d", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

// Test createSimplePlan
func TestCreateSimplePlan(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	goal := "Simple task"
	plan := decomposer.createSimplePlan(goal)

	if plan == nil {
		t.Fatal("createSimplePlan returned nil")
	}
	if plan.Goal != goal {
		t.Errorf("plan.Goal = %s, want %s", plan.Goal, goal)
	}
	if len(plan.Tasks) != 1 {
		t.Errorf("plan.Tasks length = %d, want 1", len(plan.Tasks))
	}
	if plan.Tasks[0].Description != goal {
		t.Errorf("task.Description = %s, want %s", plan.Tasks[0].Description, goal)
	}
	if plan.Tasks[0].Status != TaskStatusPending {
		t.Errorf("task.Status = %s, want %s", plan.Tasks[0].Status, TaskStatusPending)
	}
}

// Test buildDecompositionPrompt
func TestBuildDecompositionPrompt(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	goal := "Test goal"
	prompt := decomposer.buildDecompositionPrompt(goal)

	// Check that template variables are replaced
	if !strings.Contains(prompt, goal) {
		t.Error("Prompt should contain the goal")
	}
	if !strings.Contains(prompt, "3") { // MaxDepth default
		t.Error("Prompt should contain MaxDepth value")
	}
	if !strings.Contains(prompt, "5") { // MaxSubtasks default
		t.Error("Prompt should contain MaxSubtasks value")
	}
	if !strings.Contains(prompt, "2") { // MinSubtaskSplit default
		t.Error("Prompt should contain MinSubtaskSplit value")
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("Prompt should mention JSON format")
	}
}

// Test parseTasks with valid JSON
func TestParseTasksValidJSON(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	response := `{
		"tasks": [
			{
				"id": "task_1",
				"description": "First task",
				"type": "action",
				"dependencies": [],
				"subtasks": []
			},
			{
				"id": "task_2",
				"description": "Second task",
				"type": "observation",
				"dependencies": ["task_1"],
				"subtasks": []
			}
		]
	}`

	tasks, err := decomposer.parseTasks(response)
	if err != nil {
		t.Fatalf("parseTasks() failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("parseTasks() returned %d tasks, want 2", len(tasks))
	}

	if tasks[0].ID != "task_1" {
		t.Errorf("tasks[0].ID = %s, want task_1", tasks[0].ID)
	}
	if tasks[1].ID != "task_2" {
		t.Errorf("tasks[1].ID = %s, want task_2", tasks[1].ID)
	}
	if len(tasks[1].Dependencies) != 1 {
		t.Errorf("tasks[1].Dependencies length = %d, want 1", len(tasks[1].Dependencies))
	}
}

// Test parseTasks with nested subtasks
func TestParseTasksWithSubtasks(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	response := `{
		"tasks": [
			{
				"id": "task_1",
				"description": "Parent task",
				"type": "aggregate",
				"dependencies": [],
				"subtasks": [
					{
						"id": "task_1_1",
						"description": "Child task",
						"type": "action",
						"dependencies": [],
						"subtasks": []
					}
				]
			}
		]
	}`

	tasks, err := decomposer.parseTasks(response)
	if err != nil {
		t.Fatalf("parseTasks() failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("parseTasks() returned %d tasks, want 1", len(tasks))
	}
	if len(tasks[0].Subtasks) != 1 {
		t.Errorf("tasks[0].Subtasks length = %d, want 1", len(tasks[0].Subtasks))
	}
	if tasks[0].Subtasks[0].Depth != 1 {
		t.Errorf("subtask.Depth = %d, want 1", tasks[0].Subtasks[0].Depth)
	}
}

// Test parseTasks with malformed JSON
func TestParseTasksMalformedJSON(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	response := `{ "tasks": [ invalid json ] }`

	_, err := decomposer.parseTasks(response)
	if err == nil {
		t.Error("parseTasks() should fail with malformed JSON")
	}
}

// Test parseTasks with markdown code blocks
func TestParseTasksWithMarkdown(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	response := "```json\n{\"tasks\":[{\"id\":\"task_1\",\"description\":\"Test\",\"type\":\"action\",\"dependencies\":[],\"subtasks\":[]}]}\n```"

	tasks, err := decomposer.parseTasks(response)
	if err != nil {
		t.Fatalf("parseTasks() should handle markdown blocks: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("parseTasks() returned %d tasks, want 1", len(tasks))
	}
}

// Test validateTaskTree with valid tree
func TestValidateTaskTreeValid(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	tasks := []Task{
		{ID: "task_1", Depth: 0, Dependencies: []string{}, Subtasks: []Task{}},
		{ID: "task_2", Depth: 0, Dependencies: []string{"task_1"}, Subtasks: []Task{}},
	}

	err := decomposer.validateTaskTree(tasks)
	if err != nil {
		t.Errorf("validateTaskTree() failed for valid tree: %v", err)
	}
}

// Test validateTaskTree catches depth violation
func TestValidateTaskTreeDepthViolation(t *testing.T) {
	config := DefaultPlannerConfig()
	config.MaxDepth = 2
	decomposer := NewDecomposer(config, nil)

	tasks := []Task{
		{ID: "task_1", Depth: 0, Subtasks: []Task{
			{ID: "task_2", Depth: 1, Subtasks: []Task{
				{ID: "task_3", Depth: 2, Subtasks: []Task{
					{ID: "task_4", Depth: 3, Subtasks: []Task{}}, // Exceeds MaxDepth
				}},
			}},
		}},
	}

	err := decomposer.validateTaskTree(tasks)
	if err == nil {
		t.Error("validateTaskTree() should fail when depth exceeds MaxDepth")
	}
}

// Test validateTaskTree catches too many tasks
func TestValidateTaskTreeTooManyTasks(t *testing.T) {
	config := DefaultPlannerConfig()
	config.MaxSubtasks = 3
	decomposer := NewDecomposer(config, nil)

	tasks := []Task{
		{ID: "task_1", Depth: 0},
		{ID: "task_2", Depth: 0},
		{ID: "task_3", Depth: 0},
		{ID: "task_4", Depth: 0}, // Exceeds MaxSubtasks
	}

	err := decomposer.validateTaskTree(tasks)
	if err == nil {
		t.Error("validateTaskTree() should fail when task count exceeds MaxSubtasks")
	}
}

// Test validateTaskTree catches duplicate IDs
func TestValidateTaskTreeDuplicateIDs(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	tasks := []Task{
		{ID: "task_1", Depth: 0, Subtasks: []Task{}},
		{ID: "task_1", Depth: 0, Subtasks: []Task{}}, // Duplicate ID
	}

	err := decomposer.validateTaskTree(tasks)
	if err == nil {
		t.Error("validateTaskTree() should fail with duplicate IDs")
	}
}

// Test validateTaskTree catches dependency cycles
func TestValidateTaskTreeCycles(t *testing.T) {
	config := DefaultPlannerConfig()
	decomposer := NewDecomposer(config, nil)

	tasks := []Task{
		{ID: "task_1", Depth: 0, Dependencies: []string{"task_2"}, Subtasks: []Task{}},
		{ID: "task_2", Depth: 0, Dependencies: []string{"task_1"}, Subtasks: []Task{}}, // Cycle
	}

	err := decomposer.validateTaskTree(tasks)
	if err == nil {
		t.Error("validateTaskTree() should detect dependency cycles")
	}
}

// Test Decompose with simple goal (no decomposition needed)
func TestDecomposeSimpleGoal(t *testing.T) {
	config := DefaultPlannerConfig()
	mockLLM := &mockLLMGenerator{}
	decomposer := NewDecomposer(config, mockLLM)

	ctx := context.Background()
	goal := "Simple" // Short goal, low complexity

	plan, err := decomposer.Decompose(ctx, goal)
	if err != nil {
		t.Fatalf("Decompose() failed: %v", err)
	}

	if plan == nil {
		t.Fatal("Decompose() returned nil plan")
	}
	if len(plan.Tasks) != 1 {
		t.Errorf("Decompose() returned %d tasks, want 1 for simple goal", len(plan.Tasks))
	}
}

// Test Decompose with complex goal
func TestDecomposeComplexGoal(t *testing.T) {
	config := DefaultPlannerConfig()
	mockLLM := &mockLLMGenerator{
		response: `{
			"tasks": [
				{
					"id": "task_1",
					"description": "First step",
					"type": "action",
					"dependencies": [],
					"subtasks": []
				},
				{
					"id": "task_2",
					"description": "Second step",
					"type": "observation",
					"dependencies": ["task_1"],
					"subtasks": []
				}
			]
		}`,
	}
	decomposer := NewDecomposer(config, mockLLM)

	ctx := context.Background()
	goal := "This is a complex goal that requires multiple steps and coordination between different tasks"

	plan, err := decomposer.Decompose(ctx, goal)
	if err != nil {
		t.Fatalf("Decompose() failed: %v", err)
	}

	if plan == nil {
		t.Fatal("Decompose() returned nil plan")
	}
	if len(plan.Tasks) != 2 {
		t.Errorf("Decompose() returned %d tasks, want 2", len(plan.Tasks))
	}
	if plan.Goal != goal {
		t.Errorf("plan.Goal = %s, want %s", plan.Goal, goal)
	}
}

// Test Decompose with LLM error
func TestDecomposeLLMError(t *testing.T) {
	config := DefaultPlannerConfig()
	mockLLM := &mockLLMGenerator{
		err: errors.New("LLM API error"),
	}
	decomposer := NewDecomposer(config, mockLLM)

	ctx := context.Background()
	goal := "This is a complex goal requiring multi-step decomposition and coordination between different teams"

	_, err := decomposer.Decompose(ctx, goal)
	if err == nil {
		t.Error("Decompose() should fail when LLM returns error")
	}
}

// Test Decompose with invalid JSON response
func TestDecomposeInvalidJSON(t *testing.T) {
	config := DefaultPlannerConfig()
	mockLLM := &mockLLMGenerator{
		response: "invalid json response",
	}
	decomposer := NewDecomposer(config, mockLLM)

	ctx := context.Background()
	goal := "This is a complex goal requiring multi-step decomposition and coordination between different teams"

	_, err := decomposer.Decompose(ctx, goal)
	if err == nil {
		t.Error("Decompose() should fail with invalid JSON")
	}
}
