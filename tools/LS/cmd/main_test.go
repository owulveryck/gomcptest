package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestLsHandler(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ls-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files and directories
	os.Mkdir(filepath.Join(tempDir, "testdir"), 0755)
	os.WriteFile(filepath.Join(tempDir, "testfile.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "ignoreme.tmp"), []byte("ignore"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".hidden"), []byte("hidden"), 0644)

	// Context for our tests
	ctx := context.Background()

	t.Run("ValidPath", func(t *testing.T) {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"path": tempDir,
		}
		request.Params.Name = "LS"

		result, err := lsHandler(ctx, request)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Convert to JSON to inspect result
		resultJSON, _ := json.Marshal(result)
		resultString := string(resultJSON)

		if strings.Contains(resultString, "isError") && strings.Contains(resultString, "true") {
			t.Errorf("Expected success but got error: %s", resultString)
			return
		}

		if !strings.Contains(resultString, "testdir") {
			t.Errorf("Result should contain 'testdir', got: %s", resultString)
		}
		if !strings.Contains(resultString, "testfile.txt") {
			t.Errorf("Result should contain 'testfile.txt', got: %s", resultString)
		}
		if !strings.Contains(resultString, "ignoreme.tmp") {
			t.Errorf("Result should contain 'ignoreme.tmp', got: %s", resultString)
		}
		if strings.Contains(resultString, ".hidden") {
			t.Errorf("Result should not contain '.hidden', got: %s", resultString)
		}
	})

	t.Run("WithIgnorePattern", func(t *testing.T) {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"path":           tempDir,
			"ignore_pattern": "*.tmp",
		}
		request.Params.Name = "LS"

		result, err := lsHandler(ctx, request)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Convert to JSON to inspect result
		resultJSON, _ := json.Marshal(result)
		resultString := string(resultJSON)

		if strings.Contains(resultString, "isError") && strings.Contains(resultString, "true") {
			t.Errorf("Expected success but got error: %s", resultString)
			return
		}

		if !strings.Contains(resultString, "testdir") {
			t.Errorf("Result should contain 'testdir', got: %s", resultString)
		}
		if !strings.Contains(resultString, "testfile.txt") {
			t.Errorf("Result should contain 'testfile.txt', got: %s", resultString)
		}
		if strings.Contains(resultString, "ignoreme.tmp") {
			t.Errorf("Result should not contain 'ignoreme.tmp', got: %s", resultString)
		}
		if strings.Contains(resultString, ".hidden") {
			t.Errorf("Result should not contain '.hidden', got: %s", resultString)
		}
	})

	t.Run("RelativePath", func(t *testing.T) {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"path": "relative/path",
		}
		request.Params.Name = "LS"

		result, err := lsHandler(ctx, request)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Convert to JSON to inspect result
		resultJSON, _ := json.Marshal(result)
		resultString := string(resultJSON)

		if !strings.Contains(resultString, "isError") || !strings.Contains(resultString, "true") {
			t.Errorf("Expected error result but got success: %s", resultString)
			return
		}

		if !strings.Contains(resultString, "path must be an absolute path") {
			t.Errorf("Result should contain 'path must be an absolute path', got: %s", resultString)
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"path": filepath.Join(tempDir, "nonexistent"),
		}
		request.Params.Name = "LS"

		result, err := lsHandler(ctx, request)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Convert to JSON to inspect result
		resultJSON, _ := json.Marshal(result)
		resultString := string(resultJSON)

		if !strings.Contains(resultString, "isError") || !strings.Contains(resultString, "true") {
			t.Errorf("Expected error result but got success: %s", resultString)
			return
		}

		if !strings.Contains(resultString, "Path does not exist") {
			t.Errorf("Result should contain 'Path does not exist', got: %s", resultString)
		}
	})

	t.Run("EmptyDirectory", func(t *testing.T) {
		emptyDir := filepath.Join(tempDir, "empty")
		os.Mkdir(emptyDir, 0755)

		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"path": emptyDir,
		}
		request.Params.Name = "LS"

		result, err := lsHandler(ctx, request)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Convert to JSON to inspect result
		resultJSON, _ := json.Marshal(result)
		resultString := string(resultJSON)

		if strings.Contains(resultString, "isError") && strings.Contains(resultString, "true") {
			t.Errorf("Expected success but got error: %s", resultString)
			return
		}

		if !strings.Contains(resultString, "Directory is empty") {
			t.Errorf("Result should contain 'Directory is empty', got: %s", resultString)
		}
	})
}

func TestListDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "listdir-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files and directories
	os.Mkdir(filepath.Join(tempDir, "dir1"), 0755)
	os.Mkdir(filepath.Join(tempDir, "dir2"), 0755)
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test1"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.log"), []byte("test2"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".hidden"), []byte("hidden"), 0644)

	t.Run("ListAll", func(t *testing.T) {
		entries, err := listDirectory(tempDir, nil)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(entries.Dirs) != 2 {
			t.Errorf("Expected 2 directories, got %d", len(entries.Dirs))
		}
		if len(entries.Files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(entries.Files))
		}

		// Check if directory entries contain expected values
		foundDir1 := false
		foundDir2 := false
		for _, dir := range entries.Dirs {
			if dir == "dir1" {
				foundDir1 = true
			}
			if dir == "dir2" {
				foundDir2 = true
			}
		}
		if !foundDir1 {
			t.Errorf("Expected to find 'dir1' in directory listing")
		}
		if !foundDir2 {
			t.Errorf("Expected to find 'dir2' in directory listing")
		}

		// Check if file entries contain expected values
		foundFile1 := false
		foundFile2 := false
		for _, file := range entries.Files {
			if file == "file1.txt" {
				foundFile1 = true
			}
			if file == "file2.log" {
				foundFile2 = true
			}
		}
		if !foundFile1 {
			t.Errorf("Expected to find 'file1.txt' in file listing")
		}
		if !foundFile2 {
			t.Errorf("Expected to find 'file2.log' in file listing")
		}
	})

	t.Run("WithIgnore", func(t *testing.T) {
		entries, err := listDirectory(tempDir, []string{"*.log"})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(entries.Dirs) != 2 {
			t.Errorf("Expected 2 directories, got %d", len(entries.Dirs))
		}
		if len(entries.Files) != 1 {
			t.Errorf("Expected 1 file, got %d", len(entries.Files))
		}

		// Check if file entries contain expected values
		foundFile1 := false
		foundFile2 := false
		for _, file := range entries.Files {
			if file == "file1.txt" {
				foundFile1 = true
			}
			if file == "file2.log" {
				foundFile2 = true
			}
		}
		if !foundFile1 {
			t.Errorf("Expected to find 'file1.txt' in file listing")
		}
		if foundFile2 {
			t.Errorf("Expected 'file2.log' to be filtered out")
		}
	})

	t.Run("SortedOutput", func(t *testing.T) {
		// Add more files in non-alphabetical order
		os.WriteFile(filepath.Join(tempDir, "afile.txt"), []byte("a"), 0644)
		os.WriteFile(filepath.Join(tempDir, "zfile.txt"), []byte("z"), 0644)
		os.Mkdir(filepath.Join(tempDir, "adir"), 0755)
		os.Mkdir(filepath.Join(tempDir, "zdir"), 0755)

		entries, err := listDirectory(tempDir, nil)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(entries.Dirs) != 4 {
			t.Errorf("Expected 4 directories, got %d", len(entries.Dirs))
		}
		if len(entries.Files) != 4 {
			t.Errorf("Expected 4 files, got %d", len(entries.Files))
		}

		// Check that arrays are sorted alphabetically
		if len(entries.Dirs) > 0 && entries.Dirs[0] != "adir" {
			t.Errorf("Expected first directory to be 'adir', got '%s'", entries.Dirs[0])
		}
		if len(entries.Dirs) > 3 && entries.Dirs[3] != "zdir" {
			t.Errorf("Expected last directory to be 'zdir', got '%s'", entries.Dirs[3])
		}
		if len(entries.Files) > 0 && entries.Files[0] != "afile.txt" {
			t.Errorf("Expected first file to be 'afile.txt', got '%s'", entries.Files[0])
		}
		if len(entries.Files) > 3 && entries.Files[3] != "zfile.txt" {
			t.Errorf("Expected last file to be 'zfile.txt', got '%s'", entries.Files[3])
		}
	})
}
