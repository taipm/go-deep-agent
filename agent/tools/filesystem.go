package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taipm/go-deep-agent/agent"
)

// NewFileSystemTool creates a tool for file system operations.
// Security: Includes path traversal prevention and validates all paths.
//
// Available operations:
//   - read_file: Read contents of a file
//   - write_file: Write content to a file (creates if not exists)
//   - append_file: Append content to a file
//   - delete_file: Delete a file
//   - list_directory: List files and directories
//   - file_exists: Check if file/directory exists
//   - create_directory: Create a directory (with parents)
//
// Example:
//
//	fsTool := tools.NewFileSystemTool()
//	agent.NewOpenAI("gpt-4o", apiKey).
//	    WithTool(fsTool).
//	    WithAutoExecute().
//	    Ask(ctx, "Read the file data.txt")
func NewFileSystemTool() *agent.Tool {
	return agent.NewTool("filesystem", "File system operations: read, write, list files and directories").
		AddParameter("operation", "string", "Operation: read_file, write_file, append_file, delete_file, list_directory, file_exists, create_directory", true).
		AddParameter("path", "string", "File or directory path (relative or absolute)", true).
		AddParameter("content", "string", "Content to write/append (only for write_file, append_file)", false).
		WithHandler(fileSystemHandler)
}

// fileSystemHandler executes file system operations
func fileSystemHandler(args string) (string, error) {
	var params struct {
		Operation string `json:"operation"`
		Path      string `json:"path"`
		Content   string `json:"content"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Validate and sanitize path
	cleanPath, err := sanitizePath(params.Path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Execute operation
	switch params.Operation {
	case "read_file":
		return readFile(cleanPath)
	case "write_file":
		return writeFile(cleanPath, params.Content)
	case "append_file":
		return appendFile(cleanPath, params.Content)
	case "delete_file":
		return deleteFile(cleanPath)
	case "list_directory":
		return listDirectory(cleanPath)
	case "file_exists":
		return fileExists(cleanPath)
	case "create_directory":
		return createDirectory(cleanPath)
	default:
		return "", fmt.Errorf("unknown operation: %s", params.Operation)
	}
}

// sanitizePath prevents path traversal attacks and validates the path
func sanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return "", ErrSecurityViolation
	}

	// Convert to absolute path if relative
	if !filepath.IsAbs(cleanPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		cleanPath = filepath.Join(cwd, cleanPath)
	}

	return cleanPath, nil
}

// readFile reads the contents of a file
func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fmt.Sprintf("File content (%d bytes):\n%s", len(data), string(data)), nil
}

// writeFile writes content to a file (overwrites if exists)
func writeFile(path string, content string) (string, error) {
	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// appendFile appends content to a file
func appendFile(path string, content string) (string, error) {
	// Create file if it doesn't exist
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	n, err := f.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("failed to append to file: %w", err)
	}

	return fmt.Sprintf("Successfully appended %d bytes to %s", n, path), nil
}

// deleteFile deletes a file
func deleteFile(path string) (string, error) {
	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("failed to delete file: %w", err)
	}

	return fmt.Sprintf("Successfully deleted %s", path), nil
}

// listDirectory lists files and directories
func listDirectory(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("Directory %s is empty", path), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Directory %s (%d items):\n", path, len(entries)))

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileType := "FILE"
		if entry.IsDir() {
			fileType = "DIR "
		}

		result.WriteString(fmt.Sprintf("  [%s] %s (%d bytes)\n", fileType, entry.Name(), info.Size()))
	}

	return result.String(), nil
}

// fileExists checks if a file or directory exists
func fileExists(path string) (string, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Sprintf("Path does not exist: %s", path), nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to check path: %w", err)
	}

	fileType := "file"
	if info.IsDir() {
		fileType = "directory"
	}

	return fmt.Sprintf("Path exists: %s (%s, %d bytes)", path, fileType, info.Size()), nil
}

// createDirectory creates a directory and all parent directories
func createDirectory(path string) (string, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return fmt.Sprintf("Successfully created directory: %s", path), nil
}
