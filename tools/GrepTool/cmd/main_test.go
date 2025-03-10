package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestHighlightMatches(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected string
	}{
		{
			name:     "Simple word match",
			text:     "This is a test string",
			pattern:  "test",
			expected: "This is a **test** string",
		},
		{
			name:     "Multiple matches",
			text:     "test this test string test",
			pattern:  "test",
			expected: "**test** this **test** string **test**",
		},
		{
			name:     "Regex pattern",
			text:     "Testing123 test456",
			pattern:  "test\\d+",
			expected: "Testing123 **test456**",
		},
		{
			name:     "No match",
			text:     "No matches here",
			pattern:  "foo",
			expected: "No matches here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex := regexp.MustCompile(tt.pattern)
			result := highlightMatches(tt.text, regex)
			if result != tt.expected {
				t.Errorf("highlightMatches() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsBinaryFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "grepToolTests")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a text file
	textFilePath := filepath.Join(tempDir, "text.txt")
	err = ioutil.WriteFile(textFilePath, []byte("This is a text file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	// Create a binary file (with null bytes)
	binaryFilePath := filepath.Join(tempDir, "binary.bin")
	binaryData := []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x00, 0x77, 0x6f, 0x72, 0x6c, 0x64}
	err = ioutil.WriteFile(binaryFilePath, binaryData, 0644)
	if err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	// Test text file
	if isBinaryFile(textFilePath) {
		t.Errorf("Text file incorrectly identified as binary")
	}

	// Test binary file
	if !isBinaryFile(binaryFilePath) {
		t.Errorf("Binary file not identified as binary")
	}

	// Test non-existent file
	if isBinaryFile(filepath.Join(tempDir, "nonexistent.txt")) {
		t.Errorf("Nonexistent file incorrectly identified as binary")
	}
}

func TestExpandPatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "Simple pattern",
			pattern:  "*.txt",
			expected: []string{"*.txt"},
		},
		{
			name:     "Two options",
			pattern:  "*.{js,ts}",
			expected: []string{"*.js", "*.ts"},
		},
		{
			name:     "Three options",
			pattern:  "*.{js,ts,tsx}",
			expected: []string{"*.js", "*.ts", "*.tsx"},
		},
		{
			name:     "Complex pattern",
			pattern:  "src/*.{js,ts}",
			expected: []string{"src/*.js", "src/*.ts"},
		},
		{
			name:     "Invalid pattern (missing closing brace)",
			pattern:  "*.{js,ts",
			expected: []string{"*.{js,ts"},
		},
		{
			name:     "Invalid pattern (missing opening brace)",
			pattern:  "*js,ts}",
			expected: []string{"*js,ts}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPatterns(tt.pattern)
			
			// Check if lengths match
			if len(result) != len(tt.expected) {
				t.Errorf("expandPatterns() returned %d patterns, want %d", len(result), len(tt.expected))
				return
			}
			
			// Check each pattern
			for i, pattern := range result {
				if pattern != tt.expected[i] {
					t.Errorf("expandPatterns()[%d] = %v, want %v", i, pattern, tt.expected[i])
				}
			}
		})
	}
}

func TestSearchFileContent(t *testing.T) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "grepToolTestSearchContent")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with multiple lines
	testFilePath := filepath.Join(tempDir, "test.txt")
	testContent := `Line 1: Hello world
Line 2: Testing grep
Line 3: Another test line
Line 4: This is a test
Line 5: Final test line
Line 6: No matches here
Line 7: One more test
Line 8: The end`

	err = ioutil.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test cases
	tests := []struct {
		name         string
		pattern      string
		contextLines int
		expectedLen  int // Number of blocks
	}{
		{
			name:         "Simple match no context",
			pattern:      "test",
			contextLines: 0,
			expectedLen:  2, // "test" appears in multiple lines but combined by algorithm
		},
		{
			name:         "Match with context=1",
			pattern:      "test",
			contextLines: 1,
			expectedLen:  1, // With context, blocks merge
		},
		{
			name:         "Match with context=2",
			pattern:      "test",
			contextLines: 2,
			expectedLen:  1, // With more context, all blocks should merge
		},
		{
			name:         "No matches",
			pattern:      "nonexistent",
			contextLines: 0,
			expectedLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex := regexp.MustCompile(tt.pattern)
			blocks, err := searchFileContent(testFilePath, regex, tt.contextLines)
			
			if err != nil {
				t.Errorf("searchFileContent() error = %v", err)
				return
			}
			
			if len(blocks) != tt.expectedLen {
				t.Errorf("searchFileContent() returned %d blocks, want %d", len(blocks), tt.expectedLen)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	// Create a temporary test directory structure
	tempDir, err := ioutil.TempDir("", "grepToolIntegration")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files
	files := map[string]string{
		"file1.txt":           "This is a test file with some text in it.\nIt has multiple lines.\nTesting 123.",
		"file2.js":            "function test() { return 'hello world'; }",
		"file3.ts":            "class Test { constructor() { console.log('testing'); } }",
		"ignored.bin":         string([]byte{0x00, 0x01, 0x02, 0x03}), // Binary file
		".hidden/hidden.txt":  "This is a hidden file that should be ignored.",
	}

	// Create the files
	for filePath, content := range files {
		fullPath := filepath.Join(tempDir, filePath)
		
		// Create directories if needed
		dir := filepath.Dir(fullPath)
		if dir != tempDir {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}
		
		// Write the file
		if err := ioutil.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create git-like directory and file
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}
	gitConfigPath := filepath.Join(gitDir, "config")
	gitConfigContent := "This is a git config file that contains git related information"
	if err := ioutil.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create git config file: %v", err)
	}

	// Test the search function
	tests := []struct {
		name          string
		config        searchConfig
		expectedFiles int
	}{
		{
			name: "Search for 'test' in all files",
			config: searchConfig{
				Path:           tempDir,
				IncludePattern: "*",
				Regex:          regexp.MustCompile("test"),
				ContextLines:   0,
				IgnoreVCS:      true,
			},
			expectedFiles: 3, // file1.txt, file2.js, file3.ts
		},
		{
			name: "Search only in JS files",
			config: searchConfig{
				Path:           tempDir,
				IncludePattern: "*.js",
				Regex:          regexp.MustCompile("function"),
				ContextLines:   0,
				IgnoreVCS:      true,
			},
			expectedFiles: 1, // only file2.js
		},
		{
			name: "Search with case insensitive",
			config: searchConfig{
				Path:           tempDir,
				IncludePattern: "*",
				Regex:          regexp.MustCompile("(?i)TEST"),
				ContextLines:   0,
				IgnoreVCS:      true,
			},
			expectedFiles: 3, // file1.txt, file2.js, file3.ts
		},
		{
			name: "Search with no matches",
			config: searchConfig{
				Path:           tempDir,
				IncludePattern: "*",
				Regex:          regexp.MustCompile("nonexistent"),
				ContextLines:   0,
				IgnoreVCS:      true,
			},
			expectedFiles: 0,
		},
		{
			name: "Don't ignore VCS files",
			config: searchConfig{
				Path:           tempDir,
				IncludePattern: "*",
				Regex:          regexp.MustCompile("git"),
				ContextLines:   0,
				IgnoreVCS:      false,
			},
			expectedFiles: 1, // .git/config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := searchFiles(tt.config)
			
			if err != nil {
				t.Errorf("searchFiles() error = %v", err)
				return
			}
			
			if len(matches) != tt.expectedFiles {
				t.Errorf("searchFiles() returned %d files, want %d", len(matches), tt.expectedFiles)
			}
		})
	}
}