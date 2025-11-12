package agent

import (
	"context"
	"testing"
	"time"
)

func TestStreamReActNotEnabled(t *testing.T) {
	builder := &Builder{}

	_, err := builder.StreamReAct(context.Background(), "test")
	if err == nil {
		t.Error("Expected error when ReAct mode not enabled")
	}
}

func TestStreamReActInvalidConfig(t *testing.T) {
	builder := &Builder{
		reactConfig: &ReActConfig{
			Enabled:       true,
			MaxIterations: 0, // Invalid
		},
	}

	_, err := builder.StreamReAct(context.Background(), "test")
	if err == nil {
		t.Error("Expected error with invalid config")
	}
}

func TestSendEvent(t *testing.T) {
	events := make(chan ReActStreamEvent, 1)
	ctx := context.Background()

	event := ReActStreamEvent{
		Type:    "test",
		Content: "content",
	}

	ok := sendEvent(ctx, events, event)
	if !ok {
		t.Error("Expected sendEvent to return true")
	}

	received := <-events
	if received.Type != "test" {
		t.Errorf("Expected type 'test', got '%s'", received.Type)
	}
}

func TestSendEventCancelled(t *testing.T) {
	events := make(chan ReActStreamEvent)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	event := ReActStreamEvent{
		Type: "test",
	}

	ok := sendEvent(ctx, events, event)
	if ok {
		t.Error("Expected sendEvent to return false when context cancelled")
	}
}

func TestReActStreamEventStructure(t *testing.T) {
	event := ReActStreamEvent{
		Type:      "thought",
		Content:   "thinking",
		Timestamp: time.Now(),
		Iteration: 1,
	}

	if event.Type != "thought" {
		t.Errorf("Expected type 'thought', got '%s'", event.Type)
	}

	if event.Content != "thinking" {
		t.Errorf("Expected content 'thinking', got '%s'", event.Content)
	}

	if event.Iteration != 1 {
		t.Errorf("Expected iteration 1, got %d", event.Iteration)
	}
}
