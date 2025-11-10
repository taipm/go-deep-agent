package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

// This example demonstrates panic recovery in go-deep-agent.
// The agent automatically recovers from panics in tool handlers and converts them to errors.

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Panic Recovery Examples                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Example 1: Tool that panics is recovered gracefully
	example1ToolPanic(ctx, apiKey)

	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println()

	// Example 2: Detecting panic errors
	example2DetectPanic(ctx, apiKey)

	fmt.Println()
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println()

	// Example 3: Extracting panic details
	example3PanicDetails(ctx, apiKey)

	fmt.Println()
	fmt.Println("✅ All panic recovery examples completed!")
}

// Example 1: Tool panic is recovered and converted to error
func example1ToolPanic(ctx context.Context, apiKey string) {
	fmt.Println("Example 1: Graceful Panic Recovery")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Creating a tool that will panic...")
	fmt.Println()

	// Define a tool that panics
	panicTool := agent.NewTool("buggy_calculator", "A calculator with a bug").
		AddParameter("x", "number", "First number", true).
		AddParameter("y", "number", "Second number", true)

	panicTool.Handler = func(argsJSON string) (string, error) {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", err
		}

		// This will panic!
		var slice []int
		_ = slice[10] // Index out of range panic

		return "never reached", nil
	}

	// Create agent with panic recovery (automatic)
	ag := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDebug(agent.VerboseDebugConfig()).
		WithSystem("You are a math assistant. Use the buggy_calculator tool.").
		WithTool(panicTool).
		WithAutoExecute(true)

	fmt.Println("Asking agent to use the buggy tool...")
	fmt.Println()

	response, err := ag.Ask(ctx, "Calculate 5 + 3")

	if err != nil {
		fmt.Println("✅ Panic was recovered gracefully!")
		fmt.Printf("Error: %v\n", err)

		// Check if it's a panic error
		if agent.IsPanicError(err) {
			fmt.Println("✅ Error is a PanicError (panic was caught)")
		}
	} else {
		fmt.Printf("Response: %s\n", response)
	}
}

// Example 2: Detecting and handling panic errors
func example2DetectPanic(ctx context.Context, apiKey string) {
	fmt.Println("Example 2: Detecting Panic Errors")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Create a tool that panics with nil pointer
	nilTool := agent.NewTool("nil_tool", "A tool with nil pointer bug").
		AddParameter("data", "string", "Some data", true)

	nilTool.Handler = func(argsJSON string) (string, error) {
		var str *string
		return *str, nil // Nil pointer dereference panic
	}

	ag := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("Use the nil_tool.").
		WithTool(nilTool).
		WithAutoExecute(true)

	_, err := ag.Ask(ctx, "Use the tool")

	if err != nil {
		// Check if it's a panic
		if agent.IsPanicError(err) {
			fmt.Println("✅ Detected panic error!")
			fmt.Printf("Error type: PanicError\n")
			fmt.Printf("Error message: %v\n", err)

			// Get panic value
			panicValue := agent.GetPanicValue(err)
			fmt.Printf("Panic value: %v\n", panicValue)
		} else {
			fmt.Println("Regular error (not a panic)")
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// Example 3: Extracting panic details for logging/debugging
func example3PanicDetails(ctx context.Context, apiKey string) {
	fmt.Println("Example 3: Extracting Panic Details")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Create a tool that panics with custom value
	customPanicTool := agent.NewTool("custom_panic", "Tool with custom panic").
		AddParameter("action", "string", "Action to perform", true)

	customPanicTool.Handler = func(argsJSON string) (string, error) {
		panic("CUSTOM PANIC MESSAGE: Something went very wrong!")
	}

	ag := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithSystem("Use the custom_panic tool.").
		WithTool(customPanicTool).
		WithAutoExecute(true)

	_, err := ag.Ask(ctx, "Do something")

	if err != nil && agent.IsPanicError(err) {
		fmt.Println("✅ Panic recovered! Extracting details...")
		fmt.Println()

		// Get panic value
		panicValue := agent.GetPanicValue(err)
		fmt.Printf("Panic Value: %v\n", panicValue)
		fmt.Println()

		// Get stack trace
		stackTrace := agent.GetStackTrace(err)
		fmt.Println("Stack Trace (first 500 chars):")
		if len(stackTrace) > 500 {
			fmt.Println(stackTrace[:500] + "...")
		} else {
			fmt.Println(stackTrace)
		}
		fmt.Println()

		fmt.Println("💡 Use stack trace for debugging in production!")
		fmt.Println("   - Log to monitoring system (DataDog, New Relic, etc.)")
		fmt.Println("   - Alert on repeated panics")
		fmt.Println("   - Track panic patterns")
	}
}

// Production example: Comprehensive error handling with panics
func productionErrorHandler(err error) {
	if err == nil {
		return
	}

	// Check if it's a panic
	if agent.IsPanicError(err) {
		// High priority - tool code has a bug
		log.Printf("[CRITICAL] Tool panic detected: %v", err)

		// Get details
		panicValue := agent.GetPanicValue(err)
		stackTrace := agent.GetStackTrace(err)

		// Log to monitoring system
		log.Printf("[PANIC VALUE] %v", panicValue)
		log.Printf("[STACK TRACE]\n%s", stackTrace)

		// Send alert
		// alerting.SendPagerDuty("Tool panic in production", panicValue, stackTrace)

		return
	}

	// Check if it's a coded error
	if agent.IsCodedError(err) {
		code := agent.GetErrorCode(err)
		log.Printf("[ERROR CODE] %s: %v", code, err)

		// Handle by code
		switch code {
		case agent.ErrCodeRateLimitExceeded:
			// Backoff and retry
		case agent.ErrCodeAPIKeyMissing:
			// Alert - configuration issue
		default:
			// Generic handling
		}

		return
	}

	// Regular error
	log.Printf("[ERROR] %v", err)
}
