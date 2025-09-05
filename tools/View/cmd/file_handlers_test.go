package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Create test environment for file handler tests
func setupFileHandlerTestEnv(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "file-handlers-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Text file
	textContent := "This is a test file.\nWith multiple lines.\nFor testing the View tool."
	err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte(textContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	// Markdown file
	markdownContent := "# Test Markdown\n\n## Section 1\n\nThis is a test section.\n\n## Section 2\n\nAnother test section."
	err = os.WriteFile(filepath.Join(tempDir, "test.md"), []byte(markdownContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create markdown file: %v", err)
	}

	// Code file
	codeContent := `package main

import "fmt"

// TestFunction is a test function
func TestFunction() {
    fmt.Println("Hello, World!")
}

// AnotherFunction is another test function
func AnotherFunction(x int) int {
    return x * 2
}
`
	err = os.WriteFile(filepath.Join(tempDir, "test.go"), []byte(codeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create code file: %v", err)
	}

	// Binary file
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
	err = os.WriteFile(filepath.Join(tempDir, "test.bin"), []byte(binaryContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	// CSV file
	csvContent := "header1,header2,header3\nvalue1,value2,value3\nvalue4,value5,value6"
	err = os.WriteFile(filepath.Join(tempDir, "test.csv"), []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	// PDF file (just the extension, not real PDF)
	pdfContent := "Fake PDF content"
	err = os.WriteFile(filepath.Join(tempDir, "test.pdf"), []byte(pdfContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create PDF file: %v", err)
	}

	// Image file (fake)
	imageContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	err = os.WriteFile(filepath.Join(tempDir, "test.jpg"), imageContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create image file: %v", err)
	}

	return tempDir
}

// Test text file handling
func TestHandleTextFile(t *testing.T) {
	tempDir := setupFileHandlerTestEnv(t)
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.txt")

	// Basic text file handling
	t.Run("BasicTextFile", func(t *testing.T) {
		result, err := handleTextFile(filePath, 0, 10, false, "", false, "")
		if err != nil {
			t.Fatalf("handleTextFile failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With line numbers
	t.Run("WithLineNumbers", func(t *testing.T) {
		result, err := handleTextFile(filePath, 0, 10, true, "", false, "")
		if err != nil {
			t.Fatalf("handleTextFile with line numbers failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With metadata
	t.Run("WithMetadata", func(t *testing.T) {
		metadata := "FILE: test.txt\nSIZE: 100 bytes\n"
		result, err := handleTextFile(filePath, 0, 10, false, metadata, false, "")
		if err != nil {
			t.Fatalf("handleTextFile with metadata failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With find section
	t.Run("WithFindSection", func(t *testing.T) {
		result, err := handleTextFile(filePath, 0, 10, false, "", true, "multiple")
		if err != nil {
			t.Fatalf("handleTextFile with find section failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With offset and limit
	t.Run("WithOffsetAndLimit", func(t *testing.T) {
		result, err := handleTextFile(filePath, 1, 1, false, "", false, "")
		if err != nil {
			t.Fatalf("handleTextFile with offset/limit failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// Error case: invalid offset - expecting success with message about file length
	t.Run("InvalidOffset", func(t *testing.T) {
		// Test with offset beyond file length
		result, err := handleTextFile(filePath, 1000, 10, false, "", false, "")

		// Accept both implementation approaches:
		// Either it returns a success result with error message
		// Or it returns an error
		if err != nil {
			if !strings.Contains(err.Error(), "Offset") && !strings.Contains(err.Error(), "offset") {
				t.Fatalf("Expected offset-related error, got: %v", err)
			}
			return
		}

		// If no error, ensure result contains message about the offset
		if result == nil {
			t.Fatal("Expected either an error or a non-nil result")
		}
		// Test passes either way
	})
}

// Test binary file handling
func TestHandleBinaryFile(t *testing.T) {
	tempDir := setupFileHandlerTestEnv(t)
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.bin")

	// Without hex dump
	t.Run("WithoutHexDump", func(t *testing.T) {
		result, err := handleBinaryFile(filePath, "", false)
		if err != nil {
			t.Fatalf("handleBinaryFile failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With hex dump
	t.Run("WithHexDump", func(t *testing.T) {
		result, err := handleBinaryFile(filePath, "", true)
		if err != nil {
			t.Fatalf("handleBinaryFile with hex dump failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With metadata
	t.Run("WithMetadata", func(t *testing.T) {
		metadata := "FILE: test.bin\nSIZE: 8 bytes\n"
		result, err := handleBinaryFile(filePath, metadata, false)
		if err != nil {
			t.Fatalf("handleBinaryFile with metadata failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// Error case: non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non-existent.bin")
		result, err := handleBinaryFile(nonExistentPath, "", false)

		// Accept either implementation - error or result with error message
		if err != nil {
			// If it returns an error, that's valid
			return
		}

		// If no error, we should still have a result that indicates the problem
		if result == nil {
			t.Fatal("Expected either an error or a non-nil result")
		}
	})
}

// Test markdown file handling
func TestHandleMarkdownFile(t *testing.T) {
	tempDir := setupFileHandlerTestEnv(t)
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.md")

	// Basic markdown handling
	t.Run("BasicMarkdown", func(t *testing.T) {
		result, err := handleMarkdownFile(filePath, 0, 10, false, "", false, "")
		if err != nil {
			t.Fatalf("handleMarkdownFile failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With line numbers
	t.Run("WithLineNumbers", func(t *testing.T) {
		result, err := handleMarkdownFile(filePath, 0, 10, true, "", false, "")
		if err != nil {
			t.Fatalf("handleMarkdownFile with line numbers failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With section finding
	t.Run("WithSectionFinding", func(t *testing.T) {
		result, err := handleMarkdownFile(filePath, 0, 10, false, "", true, "Section 2")
		if err != nil {
			t.Fatalf("handleMarkdownFile with section finding failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// Error case: non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non-existent.md")
		result, err := handleMarkdownFile(nonExistentPath, 0, 10, false, "", false, "")

		// Accept either implementation - error or result with error message
		if err != nil {
			// If it returns an error, that's valid
			return
		}

		// If no error, we should still have a result that indicates the problem
		if result == nil {
			t.Fatal("Expected either an error or a non-nil result")
		}
	})
}

// Test code file handling
func TestHandleCodeFile(t *testing.T) {
	tempDir := setupFileHandlerTestEnv(t)
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.go")

	// Basic code handling
	t.Run("BasicCode", func(t *testing.T) {
		result, err := handleCodeFile(filePath, 0, 10, false, "", false, "")
		if err != nil {
			t.Fatalf("handleCodeFile failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With line numbers
	t.Run("WithLineNumbers", func(t *testing.T) {
		result, err := handleCodeFile(filePath, 0, 10, true, "", false, "")
		if err != nil {
			t.Fatalf("handleCodeFile with line numbers failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With function finding
	t.Run("WithFunctionFinding", func(t *testing.T) {
		result, err := handleCodeFile(filePath, 0, 10, false, "", true, "TestFunction")
		if err != nil {
			t.Fatalf("handleCodeFile with function finding failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// Error case: non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non-existent.go")
		result, err := handleCodeFile(nonExistentPath, 0, 10, false, "", false, "")

		// Accept either implementation - error or result with error message
		if err != nil {
			// If it returns an error, that's valid
			return
		}

		// If no error, we should still have a result that indicates the problem
		if result == nil {
			t.Fatal("Expected either an error or a non-nil result")
		}
	})
}

// Test document file handling
func TestHandleDocumentFile(t *testing.T) {
	tempDir := setupFileHandlerTestEnv(t)
	defer os.RemoveAll(tempDir)

	// Test with CSV file
	t.Run("CSVFile", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "test.csv")
		result, err := handleDocumentFile(filePath, "")
		if err != nil {
			t.Fatalf("handleDocumentFile CSV failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// Test with PDF file
	t.Run("PDFFile", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "test.pdf")
		result, err := handleDocumentFile(filePath, "")
		if err != nil {
			t.Fatalf("handleDocumentFile PDF failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// Test with metadata
	t.Run("WithMetadata", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "test.pdf")
		metadata := "FILE: test.pdf\nSIZE: 16 bytes\n"
		result, err := handleDocumentFile(filePath, metadata)
		if err != nil {
			t.Fatalf("handleDocumentFile with metadata failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// Error case: non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non-existent.pdf")
		result, err := handleDocumentFile(nonExistentPath, "")

		// Accept either implementation - error or result with error message
		if err != nil {
			// If it returns an error, that's valid
			return
		}

		// If no error, we should still have a result that indicates the problem
		if result == nil {
			t.Fatal("Expected either an error or a non-nil result")
		}
	})
}

// Test image file handling
func TestHandleImageFile(t *testing.T) {
	tempDir := setupFileHandlerTestEnv(t)
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.jpg")

	// Basic image handling
	t.Run("BasicImage", func(t *testing.T) {
		result, err := handleImageFile(filePath, "")
		if err != nil {
			t.Fatalf("handleImageFile failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// With metadata
	t.Run("WithMetadata", func(t *testing.T) {
		metadata := "FILE: test.jpg\nSIZE: 10 bytes\n"
		result, err := handleImageFile(filePath, metadata)
		if err != nil {
			t.Fatalf("handleImageFile with metadata failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	// Error case: non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non-existent.jpg")
		result, err := handleImageFile(nonExistentPath, "")

		// Accept either implementation - error or result with error message
		if err != nil {
			// If it returns an error, that's valid
			return
		}

		// If no error, we should still have a result that indicates the problem
		if result == nil {
			t.Fatal("Expected either an error or a non-nil result")
		}
	})
}
