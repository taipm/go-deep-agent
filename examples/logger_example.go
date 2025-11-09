package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

// Example 1: Debug Logging - Detailed tracing for development
func example1_DebugLogging() {
	fmt.Println("=== Example 1: Debug Logging ===")

	builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY")).
		WithDebugLogging(). // Enable debug-level logging
		WithSystem("You are a helpful assistant")

	ctx := context.Background()
	response, err := builder.Ask(ctx, "What is 2+2?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n\n", response)
}

// Example 2: Info Logging - Production-ready logging
func example2_InfoLogging() {
	fmt.Println("=== Example 2: Info Logging (Production) ===")

	builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY")).
		WithInfoLogging(). // Info-level logging (recommended for production)
		WithSystem("You are a helpful assistant").
		WithMemoryCache(100, 5*time.Minute)

	ctx := context.Background()

	// First request - cache miss
	response1, err := builder.Ask(ctx, "What is the capital of France?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Response 1: %s\n", response1)

	// Second request - cache hit
	response2, err := builder.Ask(ctx, "What is the capital of France?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Response 2: %s\n\n", response2)
}

// Example 3: Custom Logger - Implement your own logger
func example3_CustomLogger() {
	fmt.Println("=== Example 3: Custom Logger ===")

	// Create a custom logger that prefixes all messages
	customLogger := &PrefixLogger{prefix: "[MY-APP]"}

	builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY")).
		WithLogger(customLogger).
		WithSystem("You are a helpful assistant")

	ctx := context.Background()
	response, err := builder.Ask(ctx, "Tell me a short joke")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n\n", response)
}

// PrefixLogger is a custom logger that prefixes all log messages
type PrefixLogger struct {
	prefix string
}

func (l *PrefixLogger) Debug(ctx context.Context, msg string, fields ...agent.Field) {
	fmt.Printf("%s [DEBUG] %s %s\n", l.prefix, msg, l.formatFields(fields))
}

func (l *PrefixLogger) Info(ctx context.Context, msg string, fields ...agent.Field) {
	fmt.Printf("%s [INFO] %s %s\n", l.prefix, msg, l.formatFields(fields))
}

func (l *PrefixLogger) Warn(ctx context.Context, msg string, fields ...agent.Field) {
	fmt.Printf("%s [WARN] %s %s\n", l.prefix, msg, l.formatFields(fields))
}

func (l *PrefixLogger) Error(ctx context.Context, msg string, fields ...agent.Field) {
	fmt.Printf("%s [ERROR] %s %s\n", l.prefix, msg, l.formatFields(fields))
}

func (l *PrefixLogger) formatFields(fields []agent.Field) string {
	if len(fields) == 0 {
		return ""
	}
	result := "| "
	for i, f := range fields {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%s=%v", f.Key, f.Value)
	}
	return result
}

// Example 4: Slog Integration with Text Handler
func example4_SlogTextHandler() {
	fmt.Println("=== Example 4: Slog with Text Handler ===")

	// Create slog logger with text handler
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slogLogger := slog.New(handler)

	builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY")).
		WithLogger(agent.NewSlogAdapter(slogLogger)).
		WithSystem("You are a helpful assistant")

	ctx := context.Background()
	response, err := builder.Ask(ctx, "What is Go programming language?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n\n", response)
}

// Example 5: Slog Integration with JSON Handler (Production)
func example5_SlogJSONHandler() {
	fmt.Println("=== Example 5: Slog with JSON Handler (Production) ===")

	// Create slog logger with JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false, // Set to true to include source file/line
	})
	slogLogger := slog.New(handler)

	builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY")).
		WithLogger(agent.NewSlogAdapter(slogLogger)).
		WithSystem("You are a helpful assistant").
		WithTemperature(0.7)

	ctx := context.Background()
	response, err := builder.Ask(ctx, "Explain JSON in one sentence")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nResponse: %s\n\n", response)
}

// Example 6: Streaming with Logging
func example6_StreamingWithLogging() {
	fmt.Println("=== Example 6: Streaming with Info Logging ===")

	builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY")).
		WithInfoLogging().
		WithSystem("You are a helpful assistant").
		OnStream(func(content string) {
			// This callback is called for each chunk
			fmt.Print(content)
		})

	ctx := context.Background()
	response, err := builder.Stream(ctx, "Write a haiku about coding")
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		return
	}

	fmt.Printf("\n\nComplete response: %s\n\n", response)
}

// Example 7: No Logging (Default - Zero Overhead)
func example7_NoLogging() {
	fmt.Println("=== Example 7: No Logging (Default) ===")

	// Default builder uses NoopLogger - zero overhead
	builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY")).
		WithSystem("You are a helpful assistant")

	ctx := context.Background()
	response, err := builder.Ask(ctx, "What is 5+5?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", response)
	fmt.Println("(No logging output - NoopLogger is used by default)\n")
}

// Example 8: Logging with RAG
func example8_RAGWithLogging() {
	fmt.Println("=== Example 8: RAG with Debug Logging ===")

	builder := agent.NewOpenAI("gpt-4o-mini", os.Getenv("OPENAI_API_KEY")).
		WithDebugLogging(). // See RAG retrieval logs
		WithSystem("Answer based on the provided context").
		WithRAG(
			"Go is a statically typed, compiled programming language designed at Google.",
			"Go has built-in concurrency support with goroutines and channels.",
		)

	ctx := context.Background()
	response, err := builder.Ask(ctx, "What are the key features of Go?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nResponse: %s\n\n", response)
}

func main() {
	// Check for API key
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable")
		fmt.Println("Example: export OPENAI_API_KEY=sk-...")
		return
	}

	// Run all examples
	example1_DebugLogging()
	time.Sleep(1 * time.Second)

	example2_InfoLogging()
	time.Sleep(1 * time.Second)

	example3_CustomLogger()
	time.Sleep(1 * time.Second)

	example4_SlogTextHandler()
	time.Sleep(1 * time.Second)

	example5_SlogJSONHandler()
	time.Sleep(1 * time.Second)

	example6_StreamingWithLogging()
	time.Sleep(1 * time.Second)

	example7_NoLogging()
	time.Sleep(1 * time.Second)

	example8_RAGWithLogging()

	fmt.Println("=== All Examples Completed ===")
}
