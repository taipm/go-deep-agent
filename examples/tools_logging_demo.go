// Example demonstrating Built-in Tools logging
// This example shows how logging works with FileSystem, HTTP, and Math tools
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-deep-agent/agent"
	"github.com/taipm/go-deep-agent/agent/tools"
)

func main() {
	// Initialize agent with DEBUG logging to see all tool operations
	ai := agent.NewOllama("qwen2.5:7b").
		WithDebugLogging().
		WithTool(tools.NewFileSystemTool()).
		WithTool(tools.NewHTTPRequestTool()).
		WithTool(tools.NewMathTool()).
		WithAutoExecute(true).
		WithTimeout(60)

	ctx := context.Background()

	// Example 1: FileSystem operations (logging will show path sanitization, file operations)
	fmt.Println("\n=== Example 1: FileSystem Tool Logging ===")
	response1, err := ai.Ask(ctx, "Create a test file named 'demo.txt' with content 'Hello from Built-in Tools!'")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Response:", response1)

	// Example 2: Math operations (logging will show expression evaluation)
	fmt.Println("\n=== Example 2: Math Tool Logging ===")
	response2, err := ai.Ask(ctx, "Calculate: sqrt(16) + pow(2, 3) * sin(3.14159/2)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Response:", response2)

	// Example 3: HTTP request (logging will show request details, status, duration)
	fmt.Println("\n=== Example 3: HTTP Tool Logging ===")
	response3, err := ai.Ask(ctx, "Make a GET request to https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Response:", response3)

	// Example 4: Path traversal attempt (security logging will catch this)
	fmt.Println("\n=== Example 4: Security Logging (Path Traversal Block) ===")
	response4, err := ai.Ask(ctx, "Read the file at path '../../../etc/passwd'")
	if err != nil {
		fmt.Println("Expected security error:", err)
	} else {
		fmt.Println("Response:", response4)
	}

	fmt.Println("\n=== Logging Demo Complete ===")
	fmt.Println("Check the logs above to see:")
	fmt.Println("  - FileSystem: INFO for operations, WARN for security blocks")
	fmt.Println("  - HTTP: INFO for requests, WARN for slow/4xx, ERROR for 5xx")
	fmt.Println("  - Math: DEBUG for expressions, ERROR for invalid syntax")
}
