package agent

import (
	"time"
)

// Step type constants for ReAct pattern
const (
	// StepTypeThought represents a reasoning step
	StepTypeThought = "THOUGHT"

	// StepTypeAction represents a tool execution request
	StepTypeAction = "ACTION"

	// StepTypeObservation represents the result from a tool execution
	StepTypeObservation = "OBSERVATION"

	// StepTypeFinal represents the final answer
	StepTypeFinal = "FINAL"
)

// ReActStep represents one step in the ReAct reasoning loop.
// The ReAct pattern alternates between:
// - THOUGHT: reasoning about what to do next
// - ACTION: executing a tool with specific arguments
// - OBSERVATION: receiving the result from the tool (provided by system)
// - FINAL: providing the final answer to the user
type ReActStep struct {
	// Type identifies the step type: THOUGHT, ACTION, OBSERVATION, or FINAL
	Type string

	// Content contains the text content of the step
	// For THOUGHT: the reasoning text
	// For ACTION: the raw action string (e.g., "search(query=\"Paris\")")
	// For OBSERVATION: the tool's result
	// For FINAL: the final answer
	Content string

	// Tool is the name of the tool to execute (only for ACTION type)
	// Example: "search", "calculator", "weather"
	Tool string

	// Args contains the parsed tool arguments (only for ACTION type)
	// Example: {"query": "Paris", "limit": 10}
	Args map[string]interface{}

	// Timestamp records when this step occurred
	Timestamp time.Time

	// Error contains any error that occurred during this step (optional)
	// Typically used for ACTION steps when tool execution fails
	Error error
}

// ReActResult contains the complete outcome of a ReAct execution.
// It includes the final answer, a trace of all reasoning steps,
// and optional metrics and timeline information.
type ReActResult struct {
	// Answer is the final answer to the user's task
	Answer string

	// Steps contains the complete trace of all reasoning steps
	// This provides full transparency into the agent's thinking process
	Steps []ReActStep

	// Iterations is the number of reasoning loops executed
	Iterations int

	// Success indicates whether execution completed successfully
	// true: reached FINAL step with an answer
	// false: stopped due to error, timeout, or max iterations
	Success bool

	// Error contains the error if execution failed
	// Common errors:
	// - ErrMaxIterationsReached: exceeded iteration limit
	// - ErrTimeout: execution timed out
	// - ErrParseFailure: unable to parse LLM output (strict mode)
	Error error

	// Metrics contains execution metrics (optional)
	// Includes iteration count, tool calls, errors, duration, etc.
	Metrics *ReActMetrics

	// Timeline contains timestamped events (optional)
	// Useful for debugging and performance analysis
	Timeline *ReActTimeline
}

// ReActMetrics tracks execution metrics for a ReAct session.
// This helps monitor performance, cost, and behavior.
type ReActMetrics struct {
	// TotalIterations is the number of reasoning loops executed
	TotalIterations int

	// ToolCalls is the number of tools executed
	ToolCalls int

	// Errors is the number of errors encountered
	// Includes tool failures and parse errors
	Errors int

	// Duration is the total execution time
	Duration time.Duration

	// TokensUsed is the total number of LLM tokens consumed
	// Sum of all prompt and completion tokens
	TokensUsed int

	// StartTime records when execution began
	StartTime time.Time

	// EndTime records when execution completed
	EndTime time.Time
}

// TimelineEvent represents a single event in the execution timeline.
type TimelineEvent struct {
	// Timestamp when this event occurred
	Timestamp time.Time

	// Type of event: "step", "tool_call", "error", "complete"
	Type string

	// Content describes the event
	Content string

	// Duration of this event (for events with measurable duration)
	Duration time.Duration

	// Metadata contains additional event-specific data
	Metadata map[string]interface{}
}

// ReActTimeline provides a chronological log of all events during execution.
// This is useful for debugging, performance analysis, and understanding
// the agent's behavior.
type ReActTimeline struct {
	// Events is the chronological list of all events
	Events []TimelineEvent

	// TotalDuration is the total execution time
	TotalDuration time.Duration
}

// ReActCallback defines the interface for observing ReAct execution.
// Implement this interface to receive notifications about execution progress.
//
// Example:
//
//	type MyCallback struct{}
//
//	func (c *MyCallback) OnStep(step ReActStep) {
//	    fmt.Printf("[%s] %s\n", step.Type, step.Content)
//	}
//
//	func (c *MyCallback) OnToolCall(tool string, args map[string]interface{}) {
//	    fmt.Printf("Calling tool: %s with args: %v\n", tool, args)
//	}
//
//	func (c *MyCallback) OnError(err error) {
//	    fmt.Printf("Error: %v\n", err)
//	}
//
//	func (c *MyCallback) OnComplete(result *ReActResult) {
//	    fmt.Printf("Completed in %d iterations\n", result.Iterations)
//	}
type ReActCallback interface {
	// OnStep is called after each reasoning step
	OnStep(step ReActStep)

	// OnToolCall is called before executing a tool
	OnToolCall(tool string, args map[string]interface{})

	// OnError is called when an error occurs
	OnError(err error)

	// OnComplete is called when execution finishes
	OnComplete(result *ReActResult)
}

// AddEvent adds a new event to the timeline.
func (t *ReActTimeline) AddEvent(eventType, content string, duration time.Duration, metadata map[string]interface{}) {
	event := TimelineEvent{
		Timestamp: time.Now(),
		Type:      eventType,
		Content:   content,
		Duration:  duration,
		Metadata:  metadata,
	}
	t.Events = append(t.Events, event)
}

// NewReActMetrics creates a new metrics tracker.
func NewReActMetrics() *ReActMetrics {
	return &ReActMetrics{
		StartTime: time.Now(),
	}
}

// Finalize completes the metrics by setting the end time and calculating duration.
func (m *ReActMetrics) Finalize() {
	m.EndTime = time.Now()
	m.Duration = m.EndTime.Sub(m.StartTime)
}

// NewReActTimeline creates a new timeline tracker.
func NewReActTimeline() *ReActTimeline {
	return &ReActTimeline{
		Events: make([]TimelineEvent, 0),
	}
}

// Finalize completes the timeline by calculating total duration.
func (t *ReActTimeline) Finalize() {
	if len(t.Events) > 0 {
		firstEvent := t.Events[0]
		lastEvent := t.Events[len(t.Events)-1]
		t.TotalDuration = lastEvent.Timestamp.Sub(firstEvent.Timestamp)
	}
}
