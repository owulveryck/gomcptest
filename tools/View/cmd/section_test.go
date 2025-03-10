package main

import (
    "testing"
)

// Test markdown section finding functionality
func TestFindMarkdownSection(t *testing.T) {
    // Sample markdown document
    markdown := []string{
        "# Main Heading",
        "",
        "Some content here.",
        "",
        "## Section One",
        "",
        "Content for section one.",
        "",
        "## Section Two",
        "",
        "Content for section two.",
        "",
        "### Subsection",
        "",
        "Subsection content.",
        "",
        "## Section Three",
        "",
        "Final section content.",
    }
    
    // Test finding a section by heading
    t.Run("FindSectionByHeading", func(t *testing.T) {
        section, found := findMarkdownSection(markdown, "Section Two")
        if !found {
            t.Error("Failed to find section 'Section Two'")
        }
        if len(section) == 0 {
            t.Error("Found section is empty")
        }
        if section[0] != "## Section Two" {
            t.Errorf("Expected first line to be '## Section Two', got '%s'", section[0])
        }
        
        // Check that it includes content and subsection
        hasContent := false
        hasSubsection := false
        for _, line := range section {
            if line == "Content for section two." {
                hasContent = true
            }
            if line == "### Subsection" {
                hasSubsection = true
            }
        }
        
        if !hasContent {
            t.Error("Section doesn't contain section content")
        }
        if !hasSubsection {
            t.Error("Section doesn't contain subsection heading")
        }
    })
    
    // Test finding a section by content
    t.Run("FindSectionByContent", func(t *testing.T) {
        section, found := findMarkdownSection(markdown, "Subsection content")
        if !found {
            t.Error("Failed to find section containing 'Subsection content'")
        }
        if len(section) == 0 {
            t.Error("Found section is empty")
        }
        
        // Should include the content line
        hasContent := false
        for _, line := range section {
            if line == "Subsection content." {
                hasContent = true
                break
            }
        }
        if !hasContent {
            t.Error("Section doesn't contain the search text")
        }
    })
    
    // Test with a non-existent section
    t.Run("SectionNotFound", func(t *testing.T) {
        section, found := findMarkdownSection(markdown, "Non-existent Section")
        if found {
            t.Error("Incorrectly found non-existent section")
        }
        if len(section) > 0 {
            t.Errorf("Expected empty section, got %d lines", len(section))
        }
    })
    
    // Test with empty markdown
    t.Run("EmptyMarkdown", func(t *testing.T) {
        section, found := findMarkdownSection([]string{}, "Anything")
        if found {
            t.Error("Incorrectly found section in empty markdown")
        }
        if len(section) > 0 {
            t.Error("Expected empty section")
        }
    })
}

// Test code section finding functionality
func TestFindCodeSection(t *testing.T) {
    // Sample code
    code := []string{
        "package main",
        "",
        "import \"fmt\"",
        "",
        "// TestFunction is a test function",
        "func TestFunction() {",
        "    fmt.Println(\"Hello, World!\")",
        "}",
        "",
        "// AnotherFunction is another test function",
        "func AnotherFunction(x int) int {",
        "    return x * 2",
        "}",
    }
    
    // Test finding a function by name
    t.Run("FindFunctionByName", func(t *testing.T) {
        section, found := findCodeSection(code, "TestFunction", "go")
        if !found {
            t.Error("Failed to find function 'TestFunction'")
        }
        if len(section) == 0 {
            t.Error("Found section is empty")
        }
        
        // Check that it has the function definition and body
        hasDefinition := false
        hasBody := false
        for _, line := range section {
            if line == "func TestFunction() {" {
                hasDefinition = true
            }
            if line == "    fmt.Println(\"Hello, World!\")" {
                hasBody = true
            }
        }
        
        if !hasDefinition {
            t.Error("Section doesn't contain function definition")
        }
        if !hasBody {
            t.Error("Section doesn't contain function body")
        }
    })
    
    // Test finding code by content
    t.Run("FindByContent", func(t *testing.T) {
        section, found := findCodeSection(code, "Hello, World", "go")
        if !found {
            t.Error("Failed to find content 'Hello, World'")
        }
        if len(section) == 0 {
            t.Error("Found section is empty")
        }
        
        // Should include the content line
        hasContent := false
        for _, line := range section {
            if line == "    fmt.Println(\"Hello, World!\")" {
                hasContent = true
                break
            }
        }
        if !hasContent {
            t.Error("Section doesn't contain the search text")
        }
    })
    
    // Test with a non-existent function
    t.Run("FunctionNotFound", func(t *testing.T) {
        section, found := findCodeSection(code, "NonExistentFunction", "go")
        if found {
            t.Error("Incorrectly found non-existent function")
        }
        if len(section) > 0 {
            t.Errorf("Expected empty section, got %d lines", len(section))
        }
    })
    
    // Test with different language patterns
    t.Run("DifferentLanguagePatterns", func(t *testing.T) {
        // Python code
        pythonCode := []string{
            "#!/usr/bin/env python",
            "# Python example",
            "",
            "def hello_function():",
            "    print('Hello, World!')",
            "    return True",
            "",
            "class TestClass:",
            "    def __init__(self):",
            "        self.value = 42",
        }
        
        // Find Python function
        section, found := findCodeSection(pythonCode, "hello_function", "python")
        if !found {
            t.Error("Failed to find Python function")
        }
        if len(section) == 0 {
            t.Error("Found section is empty")
        }
        
        // Find Python class
        section, found = findCodeSection(pythonCode, "TestClass", "python")
        if !found {
            t.Error("Failed to find Python class")
        }
        if len(section) == 0 {
            t.Error("Found section is empty")
        }
    })
}

// Test definition block extraction
func TestExtractDefinitionBlock(t *testing.T) {
    // Test with Python-style indentation
    t.Run("PythonIndentation", func(t *testing.T) {
        code := []string{
            "def test_function():",
            "    print('Hello')",
            "    if True:",
            "        print('World')",
            "    print('End')",
            "",
            "# Another function",
            "def another_function():",
            "    pass",
        }
        
        // Extract the first function
        block := extractDefinitionBlock(code, 0, 10)
        
        // Should include the function and its contents
        if len(block) < 4 || len(block) > 7 {
            t.Errorf("Expected 4-7 lines for a function block, got %d", len(block))
        }
        
        // Extract with limit
        limitedBlock := extractDefinitionBlock(code, 0, 3)
        if len(limitedBlock) != 3 {
            t.Errorf("Expected 3 lines with limit, got %d", len(limitedBlock))
        }
    })
    
    // Test with brace-style code
    t.Run("BraceStyle", func(t *testing.T) {
        code := []string{
            "function testFunction() {",
            "  console.log('Hello');",
            "  if (true) {",
            "    console.log('World');",
            "  }",
            "  console.log('End');",
            "}",
            "",
            "function anotherFunction() {",
            "  return true;",
            "}",
        }
        
        // Extract the first function with braces
        block := extractDefinitionBlock(code, 0, 20)
        
        // Should include the function and its contents
        if len(block) < 5 || len(block) > 9 {
            t.Errorf("Expected 5-9 lines for a brace-style function, got %d", len(block))
        }
    })
}