package main

import (
	"testing"
)

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	// Test file type detection
	if !isImageFile(".jpg") {
		t.Error(".jpg should be detected as an image file")
	}
	if isImageFile(".txt") {
		t.Error(".txt should not be detected as an image file")
	}
	
	if !isDocumentFile(".pdf") {
		t.Error(".pdf should be detected as a document file")
	}
	if isDocumentFile(".jpg") {
		t.Error(".jpg should not be detected as a document file")
	}
	
	if !isMarkdownFile(".md") {
		t.Error(".md should be detected as markdown")
	}
	if isMarkdownFile(".txt") {
		t.Error(".txt should not be detected as markdown")
	}
	
	if !isCodeFile(".go") {
		t.Error(".go should be detected as a code file")
	}
	if isCodeFile(".jpg") {
		t.Error(".jpg should not be detected as a code file")
	}
	
	// Test MIME type detection
	if mime := getMIMETypeByExt(".jpg"); mime != "image/jpeg" {
		t.Errorf("Expected image/jpeg for .jpg, got %s", mime)
	}
	if mime := getMIMETypeByExt(".unknown"); mime != "application/octet-stream" {
		t.Errorf("Expected application/octet-stream for unknown ext, got %s", mime)
	}
	
	// Test language detection
	if lang := getLanguageFromExt(".py"); lang != "python" {
		t.Errorf("Expected python for .py, got %s", lang)
	}
	if lang := getLanguageFromExt(".unknown"); lang != "text" {
		t.Errorf("Expected text for unknown ext, got %s", lang)
	}
	
	// Test indentation level detection
	if level := getIndentationLevel("No indentation"); level != 0 {
		t.Errorf("Expected indentation level 0, got %d", level)
	}
	if level := getIndentationLevel("  Two spaces"); level != 2 {
		t.Errorf("Expected indentation level 2, got %d", level)
	}
	if level := getIndentationLevel("\tOne tab"); level != 4 {
		t.Errorf("Expected indentation level 4, got %d", level)
	}
	
	// Test hex dump creation
	emptyDump := createHexDump([]byte{})
	if emptyDump != "" {
		t.Errorf("Expected empty string for empty data, got: %s", emptyDump)
	}
	
	data := []byte{0x41, 0x42, 0x43, 0x44}  // "ABCD"
	dump := createHexDump(data)
	if dump == "" {
		t.Error("Expected non-empty hex dump")
	}
}

// Test section finding
func TestSectionFinding(t *testing.T) {
	// Test markdown section finding
	markdown := []string{
		"# Main Title",
		"",
		"Some content here.",
		"",
		"## Section One",
		"",
		"Content in section one.",
		"",
		"## Section Two",
		"",
		"Content in section two.",
	}
	
	// Test finding a section
	section, found := findMarkdownSection(markdown, "Section Two")
	if !found {
		t.Error("Failed to find 'Section Two'")
	}
	if len(section) == 0 {
		t.Error("Empty section returned")
	}
	
	// Test not finding a section
	_, found = findMarkdownSection(markdown, "Nonexistent Section")
	if found {
		t.Error("Found nonexistent section")
	}
	
	// Test code section finding
	code := []string{
		"package main",
		"",
		"// TestFunction does something",
		"func TestFunction() {",
		"    // Some code here",
		"}",
	}
	
	// Test finding a function
	section, found = findCodeSection(code, "TestFunction", "go")
	if !found {
		t.Error("Failed to find 'TestFunction'")
	}
	if len(section) == 0 {
		t.Error("Empty section returned")
	}
	
	// Test extraction of definition blocks
	block := extractDefinitionBlock(code, 3, 10)  // Start at "func TestFunction"
	if len(block) <= 0 {
		t.Error("Expected non-empty block")
	}
}