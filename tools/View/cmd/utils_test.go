package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test file type detection functions
func TestFileTypeDetection(t *testing.T) {
	// Test isImageFile
	t.Run("isImageFile", func(t *testing.T) {
		imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg", ".ico", ".heic"}
		for _, ext := range imageExts {
			if !isImageFile(ext) {
				t.Errorf("%s should be detected as an image file", ext)
			}
		}

		nonImageExts := []string{".txt", ".pdf", ".doc", ".go", ".py"}
		for _, ext := range nonImageExts {
			if isImageFile(ext) {
				t.Errorf("%s should not be detected as an image file", ext)
			}
		}
	})

	// Test isDocumentFile
	t.Run("isDocumentFile", func(t *testing.T) {
		docExts := []string{".pdf", ".doc", ".docx", ".rtf", ".txt", ".odt", ".csv"}
		for _, ext := range docExts {
			if !isDocumentFile(ext) {
				t.Errorf("%s should be detected as a document file", ext)
			}
		}

		nonDocExts := []string{".jpg", ".png", ".go", ".py", ".bin", ".exe"}
		for _, ext := range nonDocExts {
			if isDocumentFile(ext) {
				t.Errorf("%s should not be detected as a document file", ext)
			}
		}
	})

	// Test isMarkdownFile
	t.Run("isMarkdownFile", func(t *testing.T) {
		if !isMarkdownFile(".md") {
			t.Error(".md should be detected as markdown")
		}
		if !isMarkdownFile(".markdown") {
			t.Error(".markdown should be detected as markdown")
		}
		if isMarkdownFile(".txt") {
			t.Error(".txt should not be detected as markdown")
		}
	})

	// Test isCodeFile
	t.Run("isCodeFile", func(t *testing.T) {
		codeExts := []string{".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".cs", ".rb", ".php", ".html", ".css"}
		for _, ext := range codeExts {
			if !isCodeFile(ext) {
				t.Errorf("%s should be detected as a code file", ext)
			}
		}

		nonCodeExts := []string{".jpg", ".png", ".pdf", ".doc", ".bin", ".exe"}
		for _, ext := range nonCodeExts {
			if isCodeFile(ext) {
				t.Errorf("%s should not be detected as a code file", ext)
			}
		}
	})
}

// Test MIME type related tests moved to mime_test.go

// Test language detection
func TestLanguageDetection(t *testing.T) {
	// Test getLanguageFromExt
	t.Run("getLanguageFromExt", func(t *testing.T) {
		testCases := map[string]string{
			".py":   "python",
			".js":   "javascript",
			".ts":   "typescript",
			".go":   "go",
			".java": "java",
			".c":    "c",
			".cpp":  "cpp",
			".cs":   "csharp",
			".rb":   "ruby",
			".php":  "php",
			".html": "html",
			".css":  "css",
			".sql":  "sql",
			".sh":   "bash",
			".ps1":  "powershell",
		}

		for ext, expectedLang := range testCases {
			if lang := getLanguageFromExt(ext); lang != expectedLang {
				t.Errorf("For extension %s expected language %s, got %s", ext, expectedLang, lang)
			}
		}

		// Test unknown extension
		if lang := getLanguageFromExt(".unknown"); lang != "text" {
			t.Errorf("Expected 'text' for unknown extension, got %s", lang)
		}
	})
}

// Test indentation level detection
func TestIndentationLevelDetection(t *testing.T) {
	testCases := []struct {
		line     string
		expected int
	}{
		{"No indentation", 0},
		{"  Two spaces", 2},
		{"    Four spaces", 4},
		{"\tOne tab", 4},
		{"\t\tTwo tabs", 8},
		{"  \tTwo spaces and one tab", 6},
		{"", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.line, func(t *testing.T) {
			result := getIndentationLevel(tc.line)
			if result != tc.expected {
				t.Errorf("For line %q expected indentation level %d, got %d", tc.line, tc.expected, result)
			}
		})
	}
}

// Test hex dump creation
func TestHexDumpCreation(t *testing.T) {
	// Test with empty data
	t.Run("EmptyData", func(t *testing.T) {
		result := createHexDump([]byte{})
		if result != "" {
			t.Errorf("Expected empty string for empty data, got: %s", result)
		}
	})

	// Test with small data
	t.Run("SmallData", func(t *testing.T) {
		data := []byte{0x41, 0x42, 0x43, 0x44} // ABCD
		result := createHexDump(data)

		// Should contain the hex representation
		if !strings.Contains(result, "41 42 43 44") {
			t.Error("Hex dump doesn't contain the expected hex values")
		}

		// Should contain the ASCII representation
		if !strings.Contains(result, "ABCD") {
			t.Error("Hex dump doesn't contain the expected ASCII representation")
		}
	})

	// Test with data that spans multiple lines (more than 16 bytes)
	t.Run("MultiLineData", func(t *testing.T) {
		data := make([]byte, 32)
		for i := 0; i < 32; i++ {
			data[i] = byte(i)
		}

		result := createHexDump(data)

		// Should contain both line markers
		if !strings.Contains(result, "00000000") {
			t.Error("Hex dump doesn't contain the first line offset")
		}

		if !strings.Contains(result, "00000010") {
			t.Error("Hex dump doesn't contain the second line offset")
		}
	})

	// Test with non-printable characters
	t.Run("NonPrintableChars", func(t *testing.T) {
		data := []byte{0x00, 0x01, 0x02, 0x03, 0x41, 0x42, 0x43, 0x44} // \0\1\2\3ABCD
		result := createHexDump(data)

		// Should replace non-printable chars with dots
		if !strings.Contains(result, "....ABCD") {
			t.Error("Hex dump doesn't correctly handle non-printable characters")
		}
	})
}

// Test binary file detection
func TestBinaryFileDetection(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "binary-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a binary file with null bytes
	binFilePath := filepath.Join(tempDir, "test.bin")
	binContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
	if err := os.WriteFile(binFilePath, binContent, 0644); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	// Create a text file without null bytes
	txtFilePath := filepath.Join(tempDir, "test.txt")
	txtContent := []byte("This is a text file with no null bytes")
	if err := os.WriteFile(txtFilePath, txtContent, 0644); err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	// Test detection by content - binary file should be detected
	if !isProbablyBinaryFile(binFilePath, ".bin") {
		t.Error("Failed to detect binary file by content")
	}

	// Test detection by content - text file should not be detected as binary
	if isProbablyBinaryFile(txtFilePath, ".txt") {
		t.Error("Incorrectly detected text file as binary")
	}

	// Test detection by extension - even text content with a binary extension should be detected
	if !isProbablyBinaryFile(txtFilePath, ".exe") {
		t.Error("Failed to detect binary file by extension")
	}

	// Test with different extension but same content
	txtWithBinExt := filepath.Join(tempDir, "text-with-binext.exe")
	err = os.WriteFile(txtWithBinExt, txtContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create text file with binary extension: %v", err)
	}

	// Even with text content, a binary extension should identify it as binary
	if !isProbablyBinaryFile(txtWithBinExt, ".exe") {
		t.Error("Failed to detect binary file by extension")
	}
}

// Test file metadata generation
func TestFileMetadataGeneration(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "metadata-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some content to the file
	content := "Test file for metadata generation"
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Get file info
	fileInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	// Generate metadata
	metadata := generateFileMetadata(tmpFile.Name(), fileInfo)

	// Check that the metadata contains essential information
	expectedFields := []string{
		"File:", tmpFile.Name(),
		"Size:", "bytes",
		"Permissions:",
		"Modified:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(metadata, field) {
			t.Errorf("Expected metadata to contain '%s', but it doesn't", field)
		}
	}

	// Test with different file sizes
	t.Run("FileSizeFormatting", func(t *testing.T) {
		// Create files of different sizes to test size formatting
		sizes := map[string]int{
			"bytes": 500,
			"KB":    1500,
			"MB":    1500000,
			"GB":    1500000000,
		}

		for unit, size := range sizes {
			// Mock a fileinfo with the given size
			tmpFile, _ := os.CreateTemp("", "size-test-*.tmp")
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())

			// Get real fileinfo then create metadata
			info, _ := os.Stat(tmpFile.Name())

			// Create a function to override the size
			getSizeFunc := func() int64 {
				return int64(size)
			}

			// Create a mock fileinfo that returns our size
			mockInfo := mockFileInfo{
				info:     info,
				sizeFunc: getSizeFunc,
			}

			metadata := generateFileMetadata(tmpFile.Name(), mockInfo)

			// Check that the metadata contains the unit
			if !strings.Contains(metadata, unit) {
				t.Errorf("Expected metadata to contain '%s' for size %d, but it doesn't: %s",
					unit, size, metadata)
			}
		}
	})
}

// Mock FileInfo implementation for testing
type mockFileInfo struct {
	info     os.FileInfo
	sizeFunc func() int64
}

func (m mockFileInfo) Name() string       { return m.info.Name() }
func (m mockFileInfo) Size() int64        { return m.sizeFunc() }
func (m mockFileInfo) Mode() os.FileMode  { return m.info.Mode() }
func (m mockFileInfo) ModTime() time.Time { return m.info.ModTime() }
func (m mockFileInfo) IsDir() bool        { return m.info.IsDir() }
func (m mockFileInfo) Sys() interface{}   { return m.info.Sys() }
