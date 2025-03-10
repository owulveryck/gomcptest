package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/mark3labs/mcp-go/mcp"
)

// setupTestFiles creates a temporary directory with test files
func setupTestFiles(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "globtool-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test directory structure
	dirs := []string{
		"src",
		"src/utils",
		"src/components",
		"docs",
		"test",
		".hidden",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}
	}

	// Create test files
	files := []string{
		"README.md",
		"src/main.go",
		"src/utils/helpers.go",
		"src/utils/helpers_test.go",
		"src/components/widget.go",
		"src/components/widget_test.go",
		"docs/index.html",
		"docs/style.css",
		"test/main_test.go",
		".hidden/config.json",
	}

	for i, file := range files {
		filePath := filepath.Join(tempDir, file)
		// Create file with unique content and size
		content := make([]byte, (i+1)*100) // Different sizes
		for j := range content {
			content[j] = byte(i + j%256)
		}
		
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
		
		// Add delay to ensure different modification times
		time.Sleep(100 * time.Millisecond)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestGetFileInfo(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	testFile := filepath.Join(tempDir, "README.md")
	
	// Test with relative paths
	fileInfo := getFileInfo(testFile, false)
	if fileInfo.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, fileInfo.Path)
	}
	if fileInfo.Size != 100 {
		t.Errorf("Expected size 100, got %d", fileInfo.Size)
	}
	
	// Test with absolute paths
	fileInfo = getFileInfo(testFile, true)
	absPath, _ := filepath.Abs(testFile)
	if fileInfo.Path != absPath {
		t.Errorf("Expected absolute path %s, got %s", absPath, fileInfo.Path)
	}
}

// TestMCPStructs prints out the structure of MCP types
func TestMCPStructs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MCP structure test in short mode")
	}
	
	// Create and print a basic result
	textResult := mcp.NewToolResultText("test text")
	bytes, _ := json.MarshalIndent(textResult, "", "  ")
	fmt.Printf("Text Result struct: %s\n", bytes)
	
	// Create and print an error result
	errorResult := mcp.NewToolResultError("test error")
	bytes, _ = json.MarshalIndent(errorResult, "", "  ")
	fmt.Printf("Error Result struct: %s\n", bytes)
}

func TestFindMatchingFiles(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	testCases := []struct {
		name           string
		pattern        string
		excludePattern string
		useAbsolute    bool
		expectedCount  int
		expectError    bool
	}{
		{
			name:          "Find all Go files",
			pattern:       "**/*.go",
			expectedCount: 6,
		},
		{
			name:           "Find Go files excluding tests",
			pattern:        "**/*.go",
			excludePattern: "**/*_test.go",
			expectedCount:  3,
		},
		{
			name:          "Find files in src directory",
			pattern:       "**/src/**/*",
			expectedCount: 5,
		},
		{
			name:          "Find only test files",
			pattern:       "**/*_test.go",
			expectedCount: 3,
		},
		{
			name:          "Find files with absolute paths",
			pattern:       "**/*.go",
			useAbsolute:   true,
			expectedCount: 6,
		},
		{
			name:          "Find files with invalid pattern",
			pattern:       "[",
			expectError:   true,
		},
		{
			name:          "Find no matches",
			pattern:       "**/*.java",
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := findMatchingFiles(tempDir, tc.pattern, tc.excludePattern, tc.useAbsolute)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if len(files) != tc.expectedCount {
				t.Errorf("Expected %d files, got %d", tc.expectedCount, len(files))
			}
			
			// Check absolute path setting
			if tc.useAbsolute && len(files) > 0 {
				if !filepath.IsAbs(files[0].Path) {
					t.Errorf("Expected absolute path but got relative: %s", files[0].Path)
				}
			}
		})
	}
}

func TestSortFilesByModTime(t *testing.T) {
	// Create a test slice of FileInfo with known modification times
	now := time.Now()
	files := []FileInfo{
		{Path: "file1.txt", ModTime: now.Add(-2 * time.Hour)},
		{Path: "file2.txt", ModTime: now.Add(-1 * time.Hour)},
		{Path: "file3.txt", ModTime: now},
	}
	
	// Sort the files
	sortFilesByModTime(files)
	
	// Check the sorting order (newest first)
	if files[0].Path != "file3.txt" || files[1].Path != "file2.txt" || files[2].Path != "file1.txt" {
		t.Errorf("Files not sorted correctly by modification time")
	}
}

func TestFormatFileSize(t *testing.T) {
	testCases := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}
	
	for _, tc := range testCases {
		result := formatFileSize(tc.size)
		if result != tc.expected {
			t.Errorf("formatFileSize(%d) = %s; want %s", tc.size, result, tc.expected)
		}
	}
}