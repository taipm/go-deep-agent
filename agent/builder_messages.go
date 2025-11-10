package agent

// This file contains message and conversation history management methods for Builder.
// Methods: WithMessages, GetHistory, SetHistory, Clear, WithMaxHistory

//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMessages([]agent.Message{
//	        agent.User("What is 2+2?"),
//	        agent.Assistant("4"),
//	        agent.User("What is 3+3?"),
//	    })
func (b *Builder) WithMessages(messages []Message) *Builder {
	b.messages = messages
	return b
}

// GetHistory returns a copy of the current conversation history.
// The system prompt is not included in the returned messages.
//
// Example:
//
//	history := builder.GetHistory()
//	fmt.Printf("Conversation has %d messages\n", len(history))
func (b *Builder) GetHistory() []Message {
	// Return a copy to prevent external modification
	history := make([]Message, len(b.messages))
	copy(history, b.messages)
	return history
}

// SetHistory replaces the conversation history with the provided messages.
// This is useful for restoring a previous conversation state.
// The system prompt is preserved.
//
// Example:
//
//	// Save conversation
//	history := builder.GetHistory()
//
//	// Later, restore it
//	builder.SetHistory(history)
func (b *Builder) SetHistory(messages []Message) *Builder {
	b.messages = messages
	return b
}

// Clear resets the conversation history while preserving the system prompt.
// This is useful for starting a fresh conversation with the same configuration.
//
// Example:
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithSystem("You are a helpful assistant").
//	    WithMemory()
//
//	builder.Ask(ctx, "Hello")
//	builder.Clear() // Start fresh, but keep system prompt
//	builder.Ask(ctx, "Hi") // Model doesn't remember "Hello"
func (b *Builder) Clear() *Builder {
	b.messages = []Message{}
	return b
}

// WithMaxHistory sets the maximum number of messages to keep in history.
// When the limit is reached, old messages are automatically removed (FIFO).
// The system prompt is always preserved and doesn't count toward the limit.
// Set to 0 for unlimited history (default).
//
// Example:
//
//	// Keep only the last 10 messages
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithMemory().
//	    WithMaxHistory(10)
func (b *Builder) WithMaxHistory(max int) *Builder {
	b.maxHistory = max
	return b
}
