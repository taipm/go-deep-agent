package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileSystemTool(t *testing.T) {
	// Create temp directory for tests
	tempDir := t.TempDir()

	t.Run("WriteFile", func(t *testing.T) {
		tool := NewFileSystemTool()
		args := `{"operation": "write_file", "path": "` + filepath.Join(tempDir, "test.txt") + `", "content": "Hello World"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		if !strings.Contains(result, "Successfully wrote") {
			t.Errorf("Unexpected result: %s", result)
		}
	})

	t.Run("ReadFile", func(t *testing.T) {
		// Write a test file first
		testFile := filepath.Join(tempDir, "read_test.txt")
		os.WriteFile(testFile, []byte("Test Content"), 0644)

		tool := NewFileSystemTool()
		args := `{"operation": "read_file", "path": "` + testFile + `"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if !strings.Contains(result, "Test Content") {
			t.Errorf("File content not found in result: %s", result)
		}
	})

	t.Run("AppendFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "append_test.txt")
		os.WriteFile(testFile, []byte("Line 1\n"), 0644)

		tool := NewFileSystemTool()
		args := `{"operation": "append_file", "path": "` + testFile + `", "content": "Line 2\n"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("AppendFile failed: %v", err)
		}
		if !strings.Contains(result, "Successfully appended") {
			t.Errorf("Unexpected result: %s", result)
		}

		// Verify content
		content, _ := os.ReadFile(testFile)
		if !strings.Contains(string(content), "Line 1") || !strings.Contains(string(content), "Line 2") {
			t.Errorf("Append failed, content: %s", string(content))
		}
	})

	t.Run("ListDirectory", func(t *testing.T) {
		// Create some test files
		os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("test"), 0644)

		tool := NewFileSystemTool()
		args := `{"operation": "list_directory", "path": "` + tempDir + `"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("ListDirectory failed: %v", err)
		}
		if !strings.Contains(result, "file1.txt") || !strings.Contains(result, "file2.txt") {
			t.Errorf("Files not found in listing: %s", result)
		}
	})

	t.Run("FileExists", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "exists_test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		tool := NewFileSystemTool()
		args := `{"operation": "file_exists", "path": "` + testFile + `"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("FileExists failed: %v", err)
		}
		if !strings.Contains(result, "Path exists") {
			t.Errorf("Unexpected result: %s", result)
		}
	})

	t.Run("CreateDirectory", func(t *testing.T) {
		newDir := filepath.Join(tempDir, "subdir", "nested")

		tool := NewFileSystemTool()
		args := `{"operation": "create_directory", "path": "` + newDir + `"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("CreateDirectory failed: %v", err)
		}
		if !strings.Contains(result, "Successfully created") {
			t.Errorf("Unexpected result: %s", result)
		}

		// Verify directory exists
		if _, err := os.Stat(newDir); os.IsNotExist(err) {
			t.Errorf("Directory was not created: %s", newDir)
		}
	})

	t.Run("DeleteFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "delete_test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		tool := NewFileSystemTool()
		args := `{"operation": "delete_file", "path": "` + testFile + `"}`

		result, err := tool.Handler(args)
		if err != nil {
			t.Fatalf("DeleteFile failed: %v", err)
		}
		if !strings.Contains(result, "Successfully deleted") {
			t.Errorf("Unexpected result: %s", result)
		}

		// Verify file is deleted
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Errorf("File was not deleted: %s", testFile)
		}
	})

	t.Run("PathTraversalPrevention", func(t *testing.T) {
		tool := NewFileSystemTool()
		args := `{"operation": "read_file", "path": "../../../etc/passwd"}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected error for path traversal attempt")
		}
		if !strings.Contains(err.Error(), "security violation") {
			t.Errorf("Expected security violation error, got: %v", err)
		}
	})

	t.Run("InvalidOperation", func(t *testing.T) {
		tool := NewFileSystemTool()
		args := `{"operation": "invalid_op", "path": "/tmp/test.txt"}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected error for invalid operation")
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		tool := NewFileSystemTool()
		args := `{"operation": "read_file", "path": ""}`

		_, err := tool.Handler(args)
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"Valid relative path", "test.txt", false},
		{"Valid absolute path", "/tmp/test.txt", false},
		{"Path traversal ..", "../test.txt", true},
		{"Path traversal multiple", "../../etc/passwd", true},
		{"Empty path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sanitizePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("sanitizePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
