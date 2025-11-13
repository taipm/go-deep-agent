package agent

import "github.com/openai/openai-go/v3"

// Message represents a chat message in the conversation.
// This is our own type to avoid users needing to import openai-go.
type Message struct {
	Role       string     // "system", "user", "assistant", or "tool"
	Content    string     // The message content
	ToolCalls  []ToolCall // Tool calls made by assistant (only for assistant messages)
	ToolCallID string     // ID of the tool call this message is responding to (only for tool messages)
}

// System creates a system message.
// System messages set the behavior and context for the assistant.
//
// Example:
//
//	msg := agent.System("You are a helpful assistant that speaks like a pirate.")
func System(content string) Message {
	return Message{
		Role:    "system",
		Content: content,
	}
}

// User creates a user message.
// User messages are the prompts or questions from the end user.
//
// Example:
//
//	msg := agent.User("What is the capital of France?")
func User(content string) Message {
	return Message{
		Role:    "user",
		Content: content,
	}
}

// Assistant creates an assistant message.
// Assistant messages are responses from the AI model.
// Useful for providing conversation history or few-shot examples.
//
// Example:
//
//	msg := agent.Assistant("Paris is the capital of France.")
func Assistant(content string) Message {
	return Message{
		Role:    "assistant",
		Content: content,
	}
}

// convertMessages converts our Message type to OpenAI's message format.
// This internal function allows us to work with our clean API while
// maintaining compatibility with the openai-go library.
func convertMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	result := make([]openai.ChatCompletionMessageParamUnion, len(messages))

	for i, msg := range messages {
		switch msg.Role {
		case "system":
			result[i] = openai.SystemMessage(msg.Content)
		case "user":
			result[i] = openai.UserMessage(msg.Content)
		case "assistant":
			result[i] = openai.AssistantMessage(msg.Content)
		default:
			// Default to user message if role is unknown
			result[i] = openai.UserMessage(msg.Content)
		}
	}

	return result
}
