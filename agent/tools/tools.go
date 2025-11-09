// Package tools provides built-in tools for the go-deep-agent library.
//
// This package includes ready-to-use tools that can be added to agents
// for common operations like file system access, HTTP requests, and date/time handling.
//
// Available Built-in Tools (v0.5.5):
//   - FileSystemTool: Read/write files, list directories, manage file operations
//   - HTTPRequestTool: Make HTTP requests (GET, POST, PUT, DELETE) with full control
//   - DateTimeTool: Parse, format, and manipulate dates and times (SAFE - auto-loadable)
//   - MathTool: Evaluate expressions, statistics, solve equations, conversions (SAFE - auto-loadable)
//
// Usage Example:
//
//	import "github.com/taipm/go-deep-agent/agent/tools"
//
//	// Option 1: Load safe tools (DateTime + Math) by default
//	ai := tools.WithDefaults(agent.NewOpenAI("gpt-4o", apiKey)).
//	    WithAutoExecute(true)
//
//	// Option 2: Load specific tools
//	ai := agent.NewOpenAI("gpt-4o", apiKey).
//	    WithTools(tools.NewFileSystemTool(), tools.NewHTTPRequestTool()).
//	    WithAutoExecute(true)
//
//	// Option 3: Load all built-in tools
//	ai := tools.WithAll(agent.NewOpenAI("gpt-4o", apiKey)).
//	    WithAutoExecute(true)
//
// Security Notes:
//   - DateTimeTool and MathTool are SAFE (no file access, no network calls)
//   - FileSystemTool includes path traversal prevention (use with caution)
//   - HTTPRequestTool has timeout protection (use with caution)
//   - All tools include proper error handling
package tools

import (
	"fmt"

	"github.com/taipm/go-deep-agent/agent"
)

// WithDefaults adds DateTime and Math tools to the builder.
// These tools are safe (no file system access, no network calls),
// have no side effects, and enhance agent capabilities from the core.
//
// Example:
//
//	ai := tools.WithDefaults(agent.NewOpenAI("gpt-4o-mini", apiKey)).
//	    WithAutoExecute(true).
//	    Ask(ctx, "What day of the week is Christmas 2025?")
func WithDefaults(b *agent.Builder) *agent.Builder {
	return b.WithTools(
		NewDateTimeTool(),
		NewMathTool(),
	)
}

// WithAll adds all built-in tools (FileSystem, HTTP, DateTime, Math) to the builder.
// Use with caution: FileSystemTool and HTTPRequestTool have security implications.
//
// Example:
//
//	ai := tools.WithAll(agent.NewOpenAI("gpt-4o-mini", apiKey)).
//	    WithAutoExecute(true).
//	    Ask(ctx, "Fetch https://api.example.com/data and save to file.json")
func WithAll(b *agent.Builder) *agent.Builder {
	return b.WithTools(
		NewFileSystemTool(),
		NewHTTPRequestTool(),
		NewDateTimeTool(),
		NewMathTool(),
	)
}

// Common error messages
var (
	ErrInvalidInput      = fmt.Errorf("invalid input parameters")
	ErrOperationFailed   = fmt.Errorf("operation failed")
	ErrSecurityViolation = fmt.Errorf("security violation detected")
	ErrTimeout           = fmt.Errorf("operation timeout")
)

// Version information
const (
	Version             = "0.5.3"
	ToolsPackageVersion = "1.0.0"
)
