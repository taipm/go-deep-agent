package agent

// EnhancedReActCallback provides fine-grained callbacks for each step type.
// This is a convenience wrapper around ReActCallback that splits OnStep into
// separate methods for THOUGHT, ACTION, OBSERVATION, and FINAL.
type EnhancedReActCallback struct {
	// OnThought is called when the agent is reasoning
	OnThought func(content string, iteration int)

	// OnAction is called when the agent plans to execute a tool
	OnAction func(tool string, args map[string]interface{}, iteration int)

	// OnObservation is called when tool results are received
	OnObservation func(content string, iteration int)

	// OnFinal is called when the final answer is ready
	OnFinal func(answer string, iteration int)

	// OnToolExecuted is called after tool execution completes
	OnToolExecuted func(tool string, args map[string]interface{}, result string)

	// OnErrorOccurred is called when any error occurs
	OnErrorOccurred func(err error, iteration int)

	// OnCompleted is called when execution finishes
	OnCompleted func(result *ReActResult)

	// Internal tracking
	currentIteration int
}

// OnStep implements ReActCallback interface by routing to specific methods.
func (e *EnhancedReActCallback) OnStep(step ReActStep) {
	// Increment iteration on THOUGHT steps
	if step.Type == "THOUGHT" {
		e.currentIteration++
	}

	switch step.Type {
	case "THOUGHT":
		if e.OnThought != nil {
			e.OnThought(step.Content, e.currentIteration)
		}

	case "ACTION":
		if e.OnAction != nil {
			e.OnAction(step.Tool, step.Args, e.currentIteration)
		}

	case "OBSERVATION":
		if e.OnObservation != nil {
			e.OnObservation(step.Content, e.currentIteration)
		}

	case "FINAL":
		if e.OnFinal != nil {
			e.OnFinal(step.Content, e.currentIteration)
		}
	}
}

// OnToolCall implements ReActCallback interface.
func (e *EnhancedReActCallback) OnToolCall(tool string, args map[string]interface{}) {
	// This is called before execution, OnAction already handles this
}

// OnError implements ReActCallback interface.
func (e *EnhancedReActCallback) OnError(err error) {
	if e.OnErrorOccurred != nil {
		e.OnErrorOccurred(err, e.currentIteration)
	}
}

// OnComplete implements ReActCallback interface.
func (e *EnhancedReActCallback) OnComplete(result *ReActResult) {
	if e.OnCompleted != nil {
		e.OnCompleted(result)
	}
}

// NewEnhancedCallback creates an EnhancedReActCallback with default no-op implementations.
func NewEnhancedCallback() *EnhancedReActCallback {
	return &EnhancedReActCallback{
		currentIteration: 0,
	}
}

// SimpleProgressCallback provides a simple progress tracking callback.
// It counts steps and can be used to display progress bars or status updates.
type SimpleProgressCallback struct {
	TotalSteps    int
	ThoughtCount  int
	ActionCount   int
	ObservationCount int
	ErrorCount    int
	
	// OnProgress is called after each step with progress percentage
	OnProgress func(percent float64, stepType string, iteration int)
}

// OnStep implements ReActCallback interface.
func (s *SimpleProgressCallback) OnStep(step ReActStep) {
	s.TotalSteps++

	switch step.Type {
	case "THOUGHT":
		s.ThoughtCount++
	case "ACTION":
		s.ActionCount++
	case "OBSERVATION":
		s.ObservationCount++
	}

	if s.OnProgress != nil {
		// Calculate approximate progress (THOUGHT -> ACTION -> OBSERVATION -> FINAL)
		// Each complete cycle is ~25% progress per iteration
		iteration := (s.ThoughtCount + s.ActionCount + s.ObservationCount) / 3
		percent := float64(iteration*25) // Simple approximation
		if percent > 100 {
			percent = 99 // Cap at 99% until FINAL
		}
		if step.Type == "FINAL" {
			percent = 100
		}
		s.OnProgress(percent, step.Type, iteration+1)
	}
}

// OnToolCall implements ReActCallback interface.
func (s *SimpleProgressCallback) OnToolCall(tool string, args map[string]interface{}) {
	// No-op
}

// OnError implements ReActCallback interface.
func (s *SimpleProgressCallback) OnError(err error) {
	s.ErrorCount++
}

// OnComplete implements ReActCallback interface.
func (s *SimpleProgressCallback) OnComplete(result *ReActResult) {
	// Ensure final progress is 100%
	if s.OnProgress != nil && result.Success {
		s.OnProgress(100, "COMPLETE", result.Iterations)
	}
}

// NewSimpleProgressCallback creates a new progress tracking callback.
func NewSimpleProgressCallback(onProgress func(percent float64, stepType string, iteration int)) *SimpleProgressCallback {
	return &SimpleProgressCallback{
		OnProgress: onProgress,
	}
}
