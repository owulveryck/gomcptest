package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Setup creates temporary test files
func setupTestFiles(t *testing.T) (string, func()) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "replace-tool-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "existing.txt")
	err = os.WriteFile(testFile, []byte("This is an existing file.\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a directory we can't write to
	noPermDir := filepath.Join(tempDir, "noperm")
	if err := os.Mkdir(noPermDir, 0500); err != nil {
		t.Fatalf("Failed to create no-perm dir: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Chmod(noPermDir, 0700) // Restore permissions so we can delete
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// Test replacing an existing file
func TestReplaceExistingFile(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	t.Run("ReplaceExistingFile", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "existing.txt")
		newContent := "This is the new content."
		
		// First verify the original file content
		origContent, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file before replace: %v", err)
		}
		
		if string(origContent) != "This is an existing file.\n" {
			t.Fatalf("test file doesn't contain the expected text")
		}
		
		// Create arguments map
		args := map[string]interface{}{
			"file_path": filePath,
			"content":   newContent,
		}

		// Create a minimal request
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		
		// Call the handler
		result, err := replaceHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Check result - it should contain success message
		fmt.Printf("Result: %#v\n", result)
		resultStr := fmt.Sprintf("%v", result)
		if !strings.Contains(resultStr, "Successfully replaced existing file") {
			t.Errorf("expected success message for replacing file, got: %v", resultStr)
		}

		// Verify file content was changed
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file after replace: %v", err)
		}

		if string(content) != newContent {
			t.Errorf("file content doesn't match expected:\ngot: %s\nwant: %s", string(content), newContent)
		}
	})
}

// Test creating a new file
func TestCreateNewFile(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	t.Run("CreateNewFile", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "newfile.txt")
		content := "This is a new file content"
		
		// Verify file doesn't exist yet
		if _, err := os.Stat(filePath); err == nil {
			t.Fatalf("test file already exists before creation")
		}

		// Create arguments map
		args := map[string]interface{}{
			"file_path": filePath,
			"content":   content,
		}

		// Create a minimal request
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		
		// Call the handler
		result, err := replaceHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Check result - it should contain success message
		resultStr := fmt.Sprintf("%v", result)
		if !strings.Contains(resultStr, "Successfully created new file") {
			t.Errorf("expected success message for creating file, got: %v", resultStr)
		}

		// Verify file was created with correct content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read created file: %v", err)
		}

		if string(fileContent) != content {
			t.Errorf("file content doesn't match expected:\ngot: %s\nwant: %s", string(fileContent), content)
		}
	})
}

// Test relative path error
func TestRelativePath(t *testing.T) {
	t.Run("RelativePath", func(t *testing.T) {
		// Create arguments map
		args := map[string]interface{}{
			"file_path": "relative/path.txt",
			"content":   "content",
		}

		// Create a minimal request
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		
		// Call the handler
		result, err := replaceHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Should be an error
		resultStr := fmt.Sprintf("%v", result)
		if !strings.Contains(resultStr, "absolute path") {
			t.Errorf("expected error about relative path, got: %+v", result)
		}
	})
}

// Test non-existent directory
func TestNonExistentDirectory(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	t.Run("NonExistentDirectory", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "nonexistent", "file.txt")
		
		// Create arguments map
		args := map[string]interface{}{
			"file_path": filePath,
			"content":   "test content",
		}

		// Create a minimal request
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		
		// Call the handler
		result, err := replaceHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Should be an error
		resultStr := fmt.Sprintf("%v", result)
		if !strings.Contains(resultStr, "Directory does not exist") {
			t.Errorf("expected error about directory not existing, got: %+v", result)
		}
		
		// Verify file wasn't created
		if _, err := os.Stat(filePath); err == nil {
			t.Errorf("file was created in non-existent directory")
		}
	})
}

// Test invalid arguments
func TestInvalidArguments(t *testing.T) {
	t.Run("InvalidFilePathArg", func(t *testing.T) {
		// Create arguments map with non-string file_path
		args := map[string]interface{}{
			"file_path": 123, // Not a string
			"content":   "test",
		}

		var req mcp.CallToolRequest
		req.Params.Arguments = args
		
		result, _ := replaceHandler(context.Background(), req)
		
		// Should have error about file_path type
		resultStr := fmt.Sprintf("%v", result)
		if !strings.Contains(resultStr, "file_path must be a string") {
			t.Errorf("expected error about file_path type, got: %+v", result)
		}
	})

	t.Run("InvalidContentArg", func(t *testing.T) {
		// Create arguments map with non-string content
		args := map[string]interface{}{
			"file_path": "/test/path",
			"content":   123, // Not a string
		}

		var req mcp.CallToolRequest
		req.Params.Arguments = args
		
		result, _ := replaceHandler(context.Background(), req)
		
		// Should have error about content type
		resultStr := fmt.Sprintf("%v", result)
		if !strings.Contains(resultStr, "content must be a string") {
			t.Errorf("expected error about content type, got: %+v", result)
		}
	})
}

// Test write errors
func TestWriteError(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	t.Run("NoPermissionDir", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "noperm", "file.txt")
		
		// Create arguments map
		args := map[string]interface{}{
			"file_path": filePath,
			"content":   "test content",
		}

		// Create a minimal request
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		
		// Call the handler
		result, err := replaceHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Should be an error
		resultStr := fmt.Sprintf("%v", result)
		if !strings.Contains(resultStr, "Error writing file") {
			t.Errorf("expected error about writing file, got: %v", result)
		}
		
		// Verify file wasn't created
		if _, err := os.Stat(filePath); err == nil {
			t.Errorf("file was created in no permission directory")
		}
	})
}

// Test empty path
func TestEmptyPath(t *testing.T) {
	t.Run("EmptyPath", func(t *testing.T) {
		// Create arguments map
		args := map[string]interface{}{
			"file_path": "",
			"content":   "test content",
		}

		// Create a minimal request
		var req mcp.CallToolRequest
		req.Params.Arguments = args
		
		// Call the handler
		result, err := replaceHandler(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Should be an error
		resultStr := fmt.Sprintf("%v", result)
		if !strings.Contains(resultStr, "absolute path") {
			t.Errorf("expected error about absolute path, got: %v", result)
		}
	})
}