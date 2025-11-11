package agent

import (
	"context"
	"fmt"
	"time"
)

// ReActStreamEvent represents a streaming event from ReAct execution.
type ReActStreamEvent struct {
	Type      string
	Content   string
	Step      *ReActStep
	Timestamp time.Time
	Iteration int
	Error     error
}

// StreamReAct executes a task using ReAct pattern with real-time event streaming.
func (b *Builder) StreamReAct(ctx context.Context, task string) (<-chan ReActStreamEvent, error) {
	if b.reactConfig == nil || !b.reactConfig.Enabled {
		return nil, fmt.Errorf("ReAct mode is not enabled")
	}

	if err := b.reactConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ReAct configuration: %w", err)
	}

	events := make(chan ReActStreamEvent, 10)

	go func() {
		defer close(events)
		b.executeReActStream(ctx, task, events)
	}()

	return events, nil
}

// sendEvent sends an event to the channel.
func sendEvent(ctx context.Context, events chan<- ReActStreamEvent, event ReActStreamEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case events <- event:
		return true
	}
}

// executeReActStream runs the ReAct execution and sends events.
func (b *Builder) executeReActStream(ctx context.Context, task string, events chan<- ReActStreamEvent) {
	timeoutCtx, cancel := context.WithTimeout(ctx, b.reactConfig.Timeout)
	defer cancel()

	sendEvent(timeoutCtx, events, ReActStreamEvent{
		Type:      "start",
		Content:   "Starting ReAct execution",
		Timestamp: time.Now(),
	})

	systemPrompt := b.buildReActSystemPrompt()
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: task},
	}

	for iteration := 0; iteration < b.reactConfig.MaxIterations; iteration++ {
		select {
		case <-timeoutCtx.Done():
			sendEvent(timeoutCtx, events, ReActStreamEvent{
				Type:      "error",
				Content:   "Timeout",
				Timestamp: time.Now(),
				Iteration: iteration + 1,
				Error:     fmt.Errorf("timeout"),
			})
			return
		default:
		}

		response, err := b.askWithMessages(timeoutCtx, messages)
		if err != nil {
			sendEvent(timeoutCtx, events, ReActStreamEvent{
				Type:      "error",
				Content:   err.Error(),
				Timestamp: time.Now(),
				Iteration: iteration + 1,
				Error:     err,
			})
			return
		}

		stepType, content, tool, args, parseErr := parseReActStep(response)
		if parseErr != nil {
			if b.reactConfig.Strict {
				sendEvent(timeoutCtx, events, ReActStreamEvent{
					Type:      "error",
					Content:   parseErr.Error(),
					Timestamp: time.Now(),
					Iteration: iteration + 1,
					Error:     parseErr,
				})
				return
			}
			continue
		}

		step := ReActStep{
			Type:      stepType,
			Content:   content,
			Tool:      tool,
			Args:      args,
			Timestamp: time.Now(),
		}

		eventType := map[string]string{
			"THOUGHT":     "thought",
			"ACTION":      "action",
			"OBSERVATION": "observation",
			"FINAL":       "final",
		}[stepType]

		if eventType != "" {
			sendEvent(timeoutCtx, events, ReActStreamEvent{
				Type:      eventType,
				Content:   content,
				Step:      &step,
				Timestamp: step.Timestamp,
				Iteration: iteration + 1,
			})
		}

		if b.reactConfig.Callback != nil {
			b.reactConfig.Callback.OnStep(step)
		}

		messages = append(messages, Message{Role: "assistant", Content: response})

		if stepType == "FINAL" {
			sendEvent(timeoutCtx, events, ReActStreamEvent{
				Type:      "complete",
				Content:   "Success",
				Timestamp: time.Now(),
				Iteration: iteration + 1,
			})
			return
		}

		if stepType == "ACTION" && tool != "" {
			toolResult, toolErr := b.executeTool(timeoutCtx, tool, args)
			if toolErr != nil {
				sendEvent(timeoutCtx, events, ReActStreamEvent{
					Type:      "error",
					Content:   toolErr.Error(),
					Timestamp: time.Now(),
					Iteration: iteration + 1,
					Error:     toolErr,
				})
				errorPrompt := b.buildToolErrorPrompt(tool, toolErr)
				messages = append(messages, Message{Role: "user", Content: errorPrompt})
				continue
			}

			obsStep := ReActStep{
				Type:      "OBSERVATION",
				Content:   toolResult,
				Timestamp: time.Now(),
			}

			sendEvent(timeoutCtx, events, ReActStreamEvent{
				Type:      "observation",
				Content:   toolResult,
				Step:      &obsStep,
				Timestamp: obsStep.Timestamp,
				Iteration: iteration + 1,
			})

			if b.reactConfig.Callback != nil {
				b.reactConfig.Callback.OnStep(obsStep)
				b.reactConfig.Callback.OnToolCall(tool, args)
			}

			obsMessage := fmt.Sprintf("OBSERVATION: %s", toolResult)
			messages = append(messages, Message{Role: "user", Content: obsMessage})
		}
	}

	sendEvent(timeoutCtx, events, ReActStreamEvent{
		Type:      "error",
		Content:   "Max iterations reached",
		Timestamp: time.Now(),
		Iteration: b.reactConfig.MaxIterations,
		Error:     fmt.Errorf("max iterations reached"),
	})
}
