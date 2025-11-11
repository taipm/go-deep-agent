package agent

import (
	"testing"
)

func TestEnhancedCallback(t *testing.T) {
	thoughtCalled := false
	actionCalled := false
	obsCalled := false
	finalCalled := false

	callback := &EnhancedReActCallback{
		OnThought: func(content string, iteration int) {
			thoughtCalled = true
			if content != "test thought" {
				t.Errorf("Expected 'test thought', got '%s'", content)
			}
		},
		OnAction: func(tool string, args map[string]interface{}, iteration int) {
			actionCalled = true
			if tool != "calculator" {
				t.Errorf("Expected 'calculator', got '%s'", tool)
			}
		},
		OnObservation: func(content string, iteration int) {
			obsCalled = true
		},
		OnFinal: func(answer string, iteration int) {
			finalCalled = true
		},
	}

	// Test THOUGHT
	callback.OnStep(ReActStep{
		Type:    "THOUGHT",
		Content: "test thought",
	})

	if !thoughtCalled {
		t.Error("OnThought should have been called")
	}

	// Test ACTION
	callback.OnStep(ReActStep{
		Type: "ACTION",
		Tool: "calculator",
		Args: map[string]interface{}{"expr": "2+2"},
	})

	if !actionCalled {
		t.Error("OnAction should have been called")
	}

	// Test OBSERVATION
	callback.OnStep(ReActStep{
		Type:    "OBSERVATION",
		Content: "4",
	})

	if !obsCalled {
		t.Error("OnObservation should have been called")
	}

	// Test FINAL
	callback.OnStep(ReActStep{
		Type:    "FINAL",
		Content: "The answer is 4",
	})

	if !finalCalled {
		t.Error("OnFinal should have been called")
	}
}

func TestEnhancedCallbackIteration(t *testing.T) {
	iterations := []int{}

	callback := &EnhancedReActCallback{
		OnThought: func(content string, iteration int) {
			iterations = append(iterations, iteration)
		},
	}

	// Multiple THOUGHT steps should increment iteration
	callback.OnStep(ReActStep{Type: "THOUGHT", Content: "1"})
	callback.OnStep(ReActStep{Type: "THOUGHT", Content: "2"})
	callback.OnStep(ReActStep{Type: "THOUGHT", Content: "3"})

	if len(iterations) != 3 {
		t.Fatalf("Expected 3 iterations, got %d", len(iterations))
	}

	if iterations[0] != 1 || iterations[1] != 2 || iterations[2] != 3 {
		t.Errorf("Expected iterations [1,2,3], got %v", iterations)
	}
}

func TestNewEnhancedCallback(t *testing.T) {
	callback := NewEnhancedCallback()

	if callback == nil {
		t.Fatal("Expected callback, got nil")
	}

	if callback.currentIteration != 0 {
		t.Errorf("Expected iteration 0, got %d", callback.currentIteration)
	}

	// Should not panic with nil handlers
	callback.OnStep(ReActStep{Type: "THOUGHT", Content: "test"})
	callback.OnError(nil)
	callback.OnComplete(&ReActResult{})
}

func TestSimpleProgressCallback(t *testing.T) {
	progressUpdates := []float64{}
	stepTypes := []string{}

	callback := NewSimpleProgressCallback(func(percent float64, stepType string, iteration int) {
		progressUpdates = append(progressUpdates, percent)
		stepTypes = append(stepTypes, stepType)
	})

	// Simulate a complete iteration
	callback.OnStep(ReActStep{Type: "THOUGHT", Content: "thinking"})
	callback.OnStep(ReActStep{Type: "ACTION", Tool: "tool"})
	callback.OnStep(ReActStep{Type: "OBSERVATION", Content: "result"})
	callback.OnStep(ReActStep{Type: "FINAL", Content: "answer"})

	if callback.TotalSteps != 4 {
		t.Errorf("Expected 4 total steps, got %d", callback.TotalSteps)
	}

	if callback.ThoughtCount != 1 {
		t.Errorf("Expected 1 thought, got %d", callback.ThoughtCount)
	}

	if len(progressUpdates) != 4 {
		t.Errorf("Expected 4 progress updates, got %d", len(progressUpdates))
	}

	// Final step should be 100%
	lastProgress := progressUpdates[len(progressUpdates)-1]
	if lastProgress != 100 {
		t.Errorf("Expected final progress 100%%, got %.2f%%", lastProgress)
	}
}

func TestSimpleProgressCallbackErrors(t *testing.T) {
	callback := NewSimpleProgressCallback(nil)

	callback.OnError(nil)
	if callback.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", callback.ErrorCount)
	}

	callback.OnError(nil)
	callback.OnError(nil)
	if callback.ErrorCount != 3 {
		t.Errorf("Expected 3 errors, got %d", callback.ErrorCount)
	}
}

func TestSimpleProgressCallbackComplete(t *testing.T) {
	completeCalled := false
	finalPercent := 0.0

	callback := NewSimpleProgressCallback(func(percent float64, stepType string, iteration int) {
		if stepType == "COMPLETE" {
			completeCalled = true
			finalPercent = percent
		}
	})

	result := &ReActResult{
		Success:    true,
		Iterations: 2,
	}

	callback.OnComplete(result)

	if !completeCalled {
		t.Error("Complete callback should have been called")
	}

	if finalPercent != 100 {
		t.Errorf("Expected 100%% on complete, got %.2f%%", finalPercent)
	}
}

func TestEnhancedCallbackNilHandlers(t *testing.T) {
	callback := &EnhancedReActCallback{}

	// Should not panic with nil handlers
	callback.OnStep(ReActStep{Type: "THOUGHT"})
	callback.OnStep(ReActStep{Type: "ACTION"})
	callback.OnStep(ReActStep{Type: "OBSERVATION"})
	callback.OnStep(ReActStep{Type: "FINAL"})
	callback.OnToolCall("tool", nil)
	callback.OnError(nil)
	callback.OnComplete(nil)
}

func TestCallbackInterfaces(t *testing.T) {
	// Verify EnhancedReActCallback implements ReActCallback
	var _ ReActCallback = (*EnhancedReActCallback)(nil)

	// Verify SimpleProgressCallback implements ReActCallback
	var _ ReActCallback = (*SimpleProgressCallback)(nil)
}
