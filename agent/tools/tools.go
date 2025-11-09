// Package tools provides built-in tools for the go-deep-agent library.
//
// This package includes ready-to-use tools that can be added to agents
// for common operations like file system access, HTTP requests, and date/time handling.
//
// Available Built-in Tools (v0.5.3):
//   - FileSystemTool: Read/write files, list directories, manage file operations
//   - HTTPRequestTool: Make HTTP requests (GET, POST, PUT, DELETE) with full control
//   - DateTimeTool: Parse, format, and manipulate dates and times
//
// Usage Example:
//
//	import "github.com/taipm/go-deep-agent/agent/tools"
//
//	// Create built-in tools
//	fsTool := tools.NewFileSystemTool()
//	httpTool := tools.NewHTTPRequestTool()
//	dtTool := tools.NewDateTimeTool()
//
//	// Add to agent
//	agent.NewOpenAI("gpt-4o", apiKey).
//	    WithTools(fsTool, httpTool, dtTool).
//	    WithAutoExecute(true).
//	    Ask(ctx, "Read the file config.json")
//
// Security Notes:
//   - FileSystemTool includes path traversal prevention
//   - HTTPRequestTool has timeout protection
//   - All tools include proper error handling
package tools

import (
	"fmt"
)

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
