package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	ctx := context.Background()

	// Example 1: Using OpenAI
	fmt.Println("=== Example 1: OpenAI Chat ===")
	openaiAgent, err := agent.NewAgent(agent.Config{
		Provider: agent.ProviderOpenAI,
		Model:    "gpt-4o-mini",
		APIKey:   os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create OpenAI agent: %v", err)
	}

	result, err := openaiAgent.Chat(ctx, "What is the capital of Vietnam?", nil)
	if err != nil {
		log.Printf("OpenAI chat error: %v", err)
	} else {
		fmt.Printf("OpenAI Response: %s\n\n", result.Content)
	}

	// Example 2: Using Ollama
	fmt.Println("=== Example 2: Ollama Chat ===")
	ollamaAgent, err := agent.NewAgent(agent.Config{
		Provider: agent.ProviderOllama,
		Model:    "qwen3:1.7b",
		BaseURL:  "http://localhost:11434/v1",
	})
	if err != nil {
		log.Fatalf("Failed to create Ollama agent: %v", err)
	}

	result, err = ollamaAgent.Chat(ctx, "What is the capital of Vietnam?", nil)
	if err != nil {
		log.Printf("Ollama chat error: %v", err)
	} else {
		fmt.Printf("Ollama Response: %s\n\n", result.Content)
	}

	// Example 3: Streaming with OpenAI
	fmt.Println("=== Example 3: Streaming Response ===")
	fmt.Print("Streaming: ")
	result, err = openaiAgent.Chat(ctx, "Write a haiku about AI", &agent.ChatOptions{
		Stream: true,
		OnStream: func(delta string) {
			fmt.Print(delta)
		},
	})
	if err != nil {
		log.Printf("Streaming error: %v", err)
	}
	fmt.Println()

	// Example 4: Chat with conversation history
	fmt.Println("=== Example 4: Conversation History ===")
	result, err = openaiAgent.Chat(ctx, "", &agent.ChatOptions{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant that speaks like a pirate."),
			openai.UserMessage("What is machine learning?"),
		},
	})
	if err != nil {
		log.Printf("Chat with history error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", result.Content)
	}

	// Example 5: Tool calling (function calling)
	fmt.Println("=== Example 5: Tool Calling ===")
	tools := []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "get_weather",
			Description: openai.String("Get the current weather in a given location"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]string{
						"type":        "string",
						"description": "The city and state, e.g. San Francisco, CA",
					},
					"unit": map[string]any{
						"type": "string",
						"enum": []string{"celsius", "fahrenheit"},
					},
				},
				"required": []string{"location"},
			},
		}),
	}

	result, err = openaiAgent.Chat(ctx, "What's the weather like in Hanoi?", &agent.ChatOptions{
		Tools: tools,
	})
	if err != nil {
		log.Printf("Tool calling error: %v", err)
	} else {
		if len(result.Completion.Choices) > 0 && len(result.Completion.Choices[0].Message.ToolCalls) > 0 {
			toolCall := result.Completion.Choices[0].Message.ToolCalls[0]
			fmt.Printf("Tool called: %s\n", toolCall.Function.Name)
			fmt.Printf("Arguments: %s\n\n", toolCall.Function.Arguments)
		} else {
			fmt.Printf("Response: %s\n\n", result.Content)
		}
	}

	// Example 6: Advanced usage with structured outputs
	fmt.Println("=== Example 6: Structured Outputs ===")
	structuredCompletion, err := openaiAgent.GetCompletion(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Extract the name and age from: John is 25 years old"),
		},
		Model: "gpt-4o-mini",
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        "person_info",
					Description: openai.String("Extract person information"),
					Schema: openai.FunctionParameters{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]string{
								"type": "string",
							},
							"age": map[string]string{
								"type": "integer",
							},
						},
						"required":             []string{"name", "age"},
						"additionalProperties": false,
					},
					Strict: openai.Bool(true),
				},
			},
		},
	})
	if err != nil {
		log.Printf("Structured output error: %v", err)
	} else {
		fmt.Printf("Structured Response: %s\n\n", structuredCompletion.Choices[0].Message.Content)
	}
}
