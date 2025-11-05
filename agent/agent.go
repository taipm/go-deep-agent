// Package agent provides a deep-agent implementation supporting multiple LLM providers
package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

// Agent represents a deep-agent that can interact with LLMs
type Agent struct {
	config Config
	client *openai.Client
}

// ChatOptions configures the behavior of Chat method
type ChatOptions struct {
	Stream   bool                                     // Enable streaming mode
	OnStream func(string)                             // Callback for streaming chunks
	Messages []openai.ChatCompletionMessageParamUnion // Full conversation history (if provided, message param is appended)
	Tools    []openai.ChatCompletionToolUnionParam    // Tools for function calling
}

// ChatResult contains the response from Chat
type ChatResult struct {
	Content    string                 // The response content
	Completion *openai.ChatCompletion // Full completion object (useful for tool calls)
}

// Chat sends a message and returns the response
// Supports simple chat, conversation history, streaming, and tool calls via options
// If opts is nil, performs a simple non-streaming chat
func (a *Agent) Chat(ctx context.Context, message string, opts *ChatOptions) (*ChatResult, error) {
	// Default options for simple chat
	if opts == nil {
		opts = &ChatOptions{}
	}

	// Build messages array
	messages := opts.Messages
	if message != "" {
		messages = append(messages, openai.UserMessage(message))
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	// Prepare completion params
	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModel(a.config.Model),
	}
	if len(opts.Tools) > 0 {
		params.Tools = opts.Tools
	}

	// Streaming mode
	if opts.Stream {
		content, completion, err := a.chatStream(ctx, params, opts.OnStream)
		if err != nil {
			return nil, err
		}
		return &ChatResult{Content: content, Completion: completion}, nil
	}

	// Non-streaming mode
	completion, err := a.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("chat completion error: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	return &ChatResult{
		Content:    completion.Choices[0].Message.Content,
		Completion: completion,
	}, nil
}

// chatStream performs a streaming chat completion (private helper)
func (a *Agent) chatStream(ctx context.Context, params openai.ChatCompletionNewParams, callback func(string)) (string, *openai.ChatCompletion, error) {
	stream := a.client.Chat.Completions.NewStreaming(ctx, params)

	// Using ChatCompletionAccumulator for better streaming handling
	acc := openai.ChatCompletionAccumulator{}
	var fullContent string

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		// Check if content stream just finished
		if content, ok := acc.JustFinishedContent(); ok {
			fullContent = content
			if callback != nil {
				callback(content)
			}
			break
		}

		// Stream delta content
		if callback != nil && len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			callback(chunk.Choices[0].Delta.Content)
		}
	}

	if err := stream.Err(); err != nil {
		return "", nil, fmt.Errorf("stream error: %w", err)
	}

	// Build completion result from accumulated data
	completion := &openai.ChatCompletion{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: fullContent,
				},
			},
		},
	}
	return fullContent, completion, nil
}

// GetCompletion returns the full completion object for advanced usage
func (a *Agent) GetCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	// Set model from config if not specified
	if params.Model == "" {
		params.Model = openai.ChatModel(a.config.Model)
	}

	completion, err := a.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("chat completion error: %w", err)
	}

	return completion, nil
}
