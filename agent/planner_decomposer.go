package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// llmGenerator provides an interface for LLM-based text generation.
// This abstraction allows the Decomposer to work with different LLM backends
// and enables easy mocking in tests.
type llmGenerator interface {
	Generate(ctx context.Context, prompt string, opts *ChatOptions) (string, error)
}

// Decomposer breaks down complex goals into structured, actionable task trees.
// It uses LLM-based analysis to decompose goals into subtasks with dependencies,
// respecting configured constraints like maximum depth and subtask count.
type Decomposer struct {
	config *PlannerConfig
	llm    llmGenerator
}

// NewDecomposer creates a new Decomposer with the given configuration and LLM generator.
// If config is nil, it uses DefaultPlannerConfig().
func NewDecomposer(config *PlannerConfig, llm llmGenerator) *Decomposer {
	if config == nil {
		config = DefaultPlannerConfig()
	}
	return &Decomposer{
		config: config,
		llm:    llm,
	}
}

// analyzeComplexity estimates the complexity of a goal to determine if decomposition is needed.
// It returns a complexity score based on goal length, keywords, and structure.
// Higher scores indicate more complex goals requiring decomposition.
func (d *Decomposer) analyzeComplexity(goal string) int {
	score := 0

	// Length-based complexity
	words := strings.Fields(goal)
	wordCount := len(words)
	if wordCount > 20 {
		score += 3
	} else if wordCount > 10 {
		score += 2
	} else if wordCount > 5 {
		score += 1
	}

	// Keywords indicating multiple steps
	multiStepKeywords := []string{
		"and", "then", "after", "before", "also", "additionally",
		"multiple", "several", "all", "each", "every",
		"analyze", "compare", "research", "investigate",
	}
	goalLower := strings.ToLower(goal)
	for _, keyword := range multiStepKeywords {
		if strings.Contains(goalLower, keyword) {
			score++
		}
	}

	// List or enumeration indicators
	if strings.Contains(goal, ",") {
		score += strings.Count(goal, ",")
	}
	if strings.Contains(goalLower, " or ") {
		score++
	}

	return score
}

// createSimplePlan creates a single-task plan for goals that don't require decomposition.
// This is used when the complexity score is below the configured threshold.
func (d *Decomposer) createSimplePlan(goal string) *Plan {
	plan := NewPlan(goal, d.config.Strategy)

	task := Task{
		ID:           generateID(),
		ParentID:     "",
		Description:  goal,
		Type:         TaskTypeAction,
		Dependencies: []string{},
		Status:       TaskStatusPending,
		Subtasks:     []Task{},
		Depth:        0,
	}

	plan.AddTask(task)
	return plan
}

// decompositionPromptTemplate is the template for LLM task decomposition.
const decompositionPromptTemplate = `You are a task planning expert. Break down the following goal into a clear, actionable plan.

GOAL: {{.Goal}}

Please decompose this into subtasks following these rules:
1. Each subtask should be clear and actionable
2. Identify which subtasks can run in parallel (no dependencies)
3. Mark dependencies between subtasks (which tasks must complete before others)
4. Limit nesting to {{.MaxDepth}} levels maximum
5. Aim for {{.MinSubtasks}} to {{.MaxSubtasks}} subtasks per level
6. Use these task types: "action" (execute tool/action), "decision" (make choice), "observation" (gather info), "aggregate" (combine results)

Output ONLY valid JSON in this exact format (no markdown, no extra text):
{
  "tasks": [
    {
      "id": "task_1",
      "description": "Clear description of what to do",
      "type": "action",
      "dependencies": [],
      "subtasks": []
    },
    {
      "id": "task_2",
      "description": "Another task description",
      "type": "observation",
      "dependencies": ["task_1"],
      "subtasks": []
    }
  ]
}

Be specific and practical. Each task should be executable by tools or simple reasoning.`

// buildDecompositionPrompt renders the decomposition prompt with the goal and configuration parameters.
// It substitutes placeholders in the template with actual values from the config.
func (d *Decomposer) buildDecompositionPrompt(goal string) string {
	prompt := decompositionPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{.Goal}}", goal)
	prompt = strings.ReplaceAll(prompt, "{{.MaxDepth}}", fmt.Sprintf("%d", d.config.MaxDepth))
	prompt = strings.ReplaceAll(prompt, "{{.MinSubtasks}}", fmt.Sprintf("%d", d.config.MinSubtaskSplit))
	prompt = strings.ReplaceAll(prompt, "{{.MaxSubtasks}}", fmt.Sprintf("%d", d.config.MaxSubtasks))
	return prompt
}

// taskJSON represents the JSON structure for task parsing.
type taskJSON struct {
	ID           string     `json:"id"`
	Description  string     `json:"description"`
	Type         string     `json:"type"`
	Dependencies []string   `json:"dependencies"`
	Subtasks     []taskJSON `json:"subtasks"`
}

// tasksResponse represents the LLM response structure.
type tasksResponse struct {
	Tasks []taskJSON `json:"tasks"`
}

// parseTasks parses the LLM JSON response into a slice of Task structs.
// It handles markdown code blocks and validates the JSON structure.
// Returns an error if the response is malformed or contains no tasks.
func (d *Decomposer) parseTasks(response string) ([]Task, error) {
	// Clean up response - remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var resp tasksResponse
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if len(resp.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks found in response")
	}

	return d.convertJSONToTasks(resp.Tasks, "", 0), nil
}

// convertJSONToTasks recursively converts taskJSON to Task structs.
func (d *Decomposer) convertJSONToTasks(jsonTasks []taskJSON, parentID string, depth int) []Task {
	tasks := make([]Task, 0, len(jsonTasks))

	for _, jt := range jsonTasks {
		task := Task{
			ID:           jt.ID,
			ParentID:     parentID,
			Description:  jt.Description,
			Type:         TaskType(jt.Type),
			Dependencies: jt.Dependencies,
			Status:       TaskStatusPending,
			Subtasks:     []Task{},
			Depth:        depth,
		}

		// Recursively convert subtasks
		if len(jt.Subtasks) > 0 {
			task.Subtasks = d.convertJSONToTasks(jt.Subtasks, task.ID, depth+1)
		}

		tasks = append(tasks, task)
	}

	return tasks
}

// validateTaskTree validates the task tree structure against configured constraints.
// It checks for: excessive depth, too many subtasks, dependency cycles, and duplicate IDs.
// Returns an error if any validation rule is violated.
func (d *Decomposer) validateTaskTree(tasks []Task) error {
	// Check depth
	maxDepth := d.findMaxDepth(tasks)
	if maxDepth > d.config.MaxDepth {
		return fmt.Errorf("task depth %d exceeds maximum %d", maxDepth, d.config.MaxDepth)
	}

	// Check task count at each level
	if err := d.checkTaskCount(tasks); err != nil {
		return err
	}

	// Check for cycles in dependencies
	if err := d.checkCycles(tasks); err != nil {
		return err
	}

	// Validate task IDs are unique
	if err := d.checkUniqueIDs(tasks); err != nil {
		return err
	}

	return nil
}

// findMaxDepth finds the maximum depth in the task tree.
func (d *Decomposer) findMaxDepth(tasks []Task) int {
	maxDepth := 0
	for _, task := range tasks {
		if task.Depth > maxDepth {
			maxDepth = task.Depth
		}
		if len(task.Subtasks) > 0 {
			subtaskDepth := d.findMaxDepth(task.Subtasks)
			if subtaskDepth > maxDepth {
				maxDepth = subtaskDepth
			}
		}
	}
	return maxDepth
}

// checkTaskCount validates task count doesn't exceed MaxSubtasks.
func (d *Decomposer) checkTaskCount(tasks []Task) error {
	if len(tasks) > d.config.MaxSubtasks {
		return fmt.Errorf("task count %d exceeds maximum %d", len(tasks), d.config.MaxSubtasks)
	}
	for _, task := range tasks {
		if len(task.Subtasks) > 0 {
			if err := d.checkTaskCount(task.Subtasks); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkCycles detects dependency cycles in the task tree.
func (d *Decomposer) checkCycles(tasks []Task) error {
	taskMap := d.buildTaskMap(tasks)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, task := range tasks {
		if err := d.detectCycle(task.ID, taskMap, visited, recStack); err != nil {
			return err
		}
	}
	return nil
}

// buildTaskMap creates a map of all tasks by ID.
func (d *Decomposer) buildTaskMap(tasks []Task) map[string]*Task {
	taskMap := make(map[string]*Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
		if len(tasks[i].Subtasks) > 0 {
			for k, v := range d.buildTaskMap(tasks[i].Subtasks) {
				taskMap[k] = v
			}
		}
	}
	return taskMap
}

// detectCycle performs DFS to detect cycles.
func (d *Decomposer) detectCycle(taskID string, taskMap map[string]*Task, visited, recStack map[string]bool) error {
	if recStack[taskID] {
		return fmt.Errorf("dependency cycle detected involving task %s", taskID)
	}
	if visited[taskID] {
		return nil
	}

	visited[taskID] = true
	recStack[taskID] = true

	if task, exists := taskMap[taskID]; exists {
		for _, depID := range task.Dependencies {
			if err := d.detectCycle(depID, taskMap, visited, recStack); err != nil {
				return err
			}
		}
	}

	recStack[taskID] = false
	return nil
}

// checkUniqueIDs ensures all task IDs are unique.
func (d *Decomposer) checkUniqueIDs(tasks []Task) error {
	seen := make(map[string]bool)
	return d.checkUniqueIDsRecursive(tasks, seen)
}

// checkUniqueIDsRecursive recursively checks for duplicate IDs.
func (d *Decomposer) checkUniqueIDsRecursive(tasks []Task, seen map[string]bool) error {
	for _, task := range tasks {
		if seen[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		seen[task.ID] = true
		if len(task.Subtasks) > 0 {
			if err := d.checkUniqueIDsRecursive(task.Subtasks, seen); err != nil {
				return err
			}
		}
	}
	return nil
}

// identifyDependencies analyzes task relationships.
// Currently uses explicit dependencies from the LLM response.
// This method serves as a placeholder for future implicit dependency detection logic.
func (d *Decomposer) identifyDependencies(tasks []Task) {
	// Dependencies are already populated by LLM response
	// This method is a placeholder for future implicit dependency detection
	// For now, we trust the LLM-provided dependencies
}

// Decompose breaks down a goal into a structured plan with actionable subtasks.
// It analyzes the goal's complexity and either creates a simple single-task plan
// or uses LLM-based decomposition to generate a task tree with dependencies.
// Returns an error if LLM generation, parsing, or validation fails.
func (d *Decomposer) Decompose(ctx context.Context, goal string) (*Plan, error) {
	// Step 1: Analyze goal complexity
	complexity := d.analyzeComplexity(goal)

	// Step 2: If simple, create single-task plan
	if complexity < d.config.MinSubtaskSplit {
		return d.createSimplePlan(goal), nil
	}

	// Step 3: Build decomposition prompt
	prompt := d.buildDecompositionPrompt(goal)

	// Step 4: Call LLM for decomposition
	result, err := d.llm.Generate(ctx, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Step 5: Parse LLM response into tasks
	tasks, err := d.parseTasks(result)
	if err != nil {
		return nil, fmt.Errorf("task parsing failed: %w", err)
	}

	// Step 6: Validate task tree
	if err := d.validateTaskTree(tasks); err != nil {
		return nil, fmt.Errorf("task validation failed: %w", err)
	}

	// Step 7: Identify dependencies (already done by LLM)
	d.identifyDependencies(tasks)

	// Step 8: Create and return plan
	plan := NewPlan(goal, d.config.Strategy)
	for _, task := range tasks {
		plan.AddTask(task)
	}

	return plan, nil
}
