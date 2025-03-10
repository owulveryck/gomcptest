package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Edit ✏️",
		"1.1.0",
	)

	// Add Edit tool
	tool := mcp.NewTool("Edit",
		mcp.WithDescription("This is a tool for editing files. For moving or renaming files, you should generally use the Bash tool with the 'mv' command instead. For larger edits, use the Replace tool to overwrite files. For Jupyter notebooks (.ipynb files), use the NotebookEditCell instead.\n\nBefore using this tool:\n\n1. Use the View tool to understand the file's contents and context\n\n2. Verify the directory path is correct (only applicable when creating new files):\n   - Use the LS tool to verify the parent directory exists and is the correct location\n\nTo make a file edit, provide the following:\n1. file_path: The absolute path to the file to modify (must be absolute, not relative)\n2. old_string: The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)\n3. new_string: The edited text to replace the old_string\n\nThe tool will replace ONE occurrence of old_string with new_string in the specified file.\n\nCRITICAL REQUIREMENTS FOR USING THIS TOOL:\n\n1. UNIQUENESS: The old_string MUST uniquely identify the specific instance you want to change. This means:\n   - Include AT LEAST 3-5 lines of context BEFORE the change point\n   - Include AT LEAST 3-5 lines of context AFTER the change point\n   - Include all whitespace, indentation, and surrounding code exactly as it appears in the file\n\n2. SINGLE INSTANCE: This tool can only change ONE instance at a time. If you need to change multiple instances:\n   - Make separate calls to this tool for each instance\n   - Each call must uniquely identify its specific instance using extensive context\n   - When making multiple edits to the same file, include all edits in a single message with multiple tool calls\n\n3. VERIFICATION: Before using this tool:\n   - Check how many instances of the target text exist in the file using GrepTool\n   - If multiple instances exist, gather enough context to uniquely identify each one\n   - Plan separate tool calls for each instance\n\nWARNING: If you do not follow these requirements:\n   - The tool will fail if old_string matches multiple locations (will report line numbers)\n   - The tool will fail if old_string doesn't match exactly (including whitespace)\n   - You may change the wrong instance if you don't include enough context\n\nWhen making edits:\n   - Ensure the edit results in idiomatic, correct code\n   - Do not leave the code in a broken state\n   - Always use absolute file paths (starting with /)\n   - The tool automatically creates backup files (.bak) before making changes\n   - Files larger than 10MB cannot be edited with this tool\n   - Binary files cannot be edited with this tool\n\nIf you want to create a new file, use:\n   - A new file path, including dir name if needed\n   - An empty old_string\n   - The new file's contents as new_string"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the file to modify"),
		),
		mcp.WithString("old_string",
			mcp.Required(),
			mcp.Description("The text to replace"),
		),
		mcp.WithString("new_string",
			mcp.Required(),
			mcp.Description("The edited text to replace the old_string"),
		),
	)

	// Add tool handler
	s.AddTool(tool, editHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func editHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get start time for performance logging
	startTime := time.Now()
	
	// Extract parameters
	filePath, ok := request.Params.Arguments["file_path"].(string)
	if !ok {
		return mcp.NewToolResultError("file_path must be a string"), nil
	}

	oldString, ok := request.Params.Arguments["old_string"].(string)
	if !ok {
		return mcp.NewToolResultError("old_string must be a string"), nil
	}

	newString, ok := request.Params.Arguments["new_string"].(string)
	if !ok {
		return mcp.NewToolResultError("new_string must be a string"), nil
	}

	// Validate that the path is absolute
	if !filepath.IsAbs(filePath) {
		return mcp.NewToolResultError("file_path must be an absolute path, not a relative path"), nil
	}
	
	// Basic parameter validation
	if len(newString) > 10*1024*1024 {
		return mcp.NewToolResultError("new_string is too large (over 10MB)"), nil
	}

	// Creating a new file
	if oldString == "" {
		return createNewFile(filePath, newString)
	}

	// Modifying an existing file
	result, err := modifyFile(filePath, oldString, newString)
	
	// Log performance metrics
	elapsed := time.Since(startTime)
	if elapsed > 100*time.Millisecond {
		// Only log if operation took more than 100ms
		fmt.Fprintf(os.Stderr, "Edit operation on %s took %v\n", filePath, elapsed)
	}
	
	return result, err
}

// Create a new file with the given content
func createNewFile(filePath, content string) (*mcp.CallToolResult, error) {
	// Validate file path
	if err := validateFilePath(filePath); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	// Check if the directory exists
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Directory does not exist: %s", dir)), nil
	}

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		return mcp.NewToolResultError(fmt.Sprintf("File already exists: %s. Use a non-empty old_string to modify it.", filePath)), nil
	}
	
	// Check if content is too large (limit to 10MB to prevent memory issues)
	if len(content) > 10*1024*1024 {
		return mcp.NewToolResultError(fmt.Sprintf("Content is too large (%d bytes). Maximum size is 10MB.", len(content))), nil
	}
	
	// Check if we have permission to write to the directory
	if err := checkDirWritePermission(dir); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Permission error: %v", err)), nil
	}

	// Create the file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Created new file: %s", filePath)), nil
}

// validateFilePath checks if the file path is valid
func validateFilePath(filePath string) error {
	// Check if path is empty
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	
	// Check if path is too long
	if len(filePath) > 4096 {
		return fmt.Errorf("file path is too long")
	}
	
	// Simple check for potential path traversal
	cleaned := filepath.Clean(filePath)
	if cleaned != filePath {
		return fmt.Errorf("invalid file path: %s", filePath)
	}
	
	return nil
}

// checkDirWritePermission checks if we can write to the directory
func checkDirWritePermission(dirPath string) error {
	// Try to create a temporary file in the directory
	tempFile := filepath.Join(dirPath, ".edit_permission_check")
	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("cannot write to directory: %v", err)
	}
	f.Close()
	
	// Clean up the temporary file
	os.Remove(tempFile)
	return nil
}

// Modify an existing file by replacing oldString with newString
func modifyFile(filePath, oldString, newString string) (*mcp.CallToolResult, error) {
	// Validate file path
	if err := validateFilePath(filePath); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("File does not exist: %s", filePath)), nil
	}
	
	// Check if it's a directory
	if fileInfo.IsDir() {
		return mcp.NewToolResultError(fmt.Sprintf("%s is a directory, not a file", filePath)), nil
	}
	
	// Check if file is too large (limit to 10MB to prevent memory issues)
	if fileInfo.Size() > 10*1024*1024 {
		return mcp.NewToolResultError(fmt.Sprintf("File is too large (%d bytes). Maximum size is 10MB.", fileInfo.Size())), nil
	}
	
	// Check if we have permission to read/write the file
	if err := checkFilePermissions(filePath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Permission error: %v", err)), nil
	}

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}

	contentStr := string(content)
	
	// Detect if the file appears to be binary
	if isBinaryFile(content) {
		return mcp.NewToolResultError("Cannot edit binary files. Use a different tool for binary file manipulation."), nil
	}

	// Count occurrences of oldString
	count := strings.Count(contentStr, oldString)
	if count == 0 {
		return mcp.NewToolResultError("The specified old_string was not found in the file. Please check the string and try again."), nil
	} else if count > 1 {
		// Find line numbers where matches occur to provide better context
		lineNumbers := findLineNumbers(contentStr, oldString)
		return mcp.NewToolResultError(fmt.Sprintf("The specified old_string occurs %d times in the file (around lines %s). Please provide a more specific string with enough context to uniquely identify the instance you want to change.", 
			count, formatLineNumbers(lineNumbers))), nil
	}

	// Create backup file before making changes
	backupPath := filePath + ".bak"
	if err := os.WriteFile(backupPath, content, fileInfo.Mode()); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(os.Stderr, "Warning: Could not create backup file: %v\n", err)
	}

	// Replace the string while preserving original line endings
	newContent := strings.Replace(contentStr, oldString, newString, 1)

	// Write the modified content back to the file with original permissions
	err = os.WriteFile(filePath, []byte(newContent), fileInfo.Mode())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error writing to file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully edited file: %s\nReplaced 1 occurrence of the specified text.\nBackup created at %s", filePath, backupPath)), nil
}

// Check if we have permission to read/write the file
func checkFilePermissions(filePath string) error {
	// Check read permission
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %v", err)
	}
	f.Close()
	
	// Check write permission by opening for append
	f, err = os.OpenFile(filePath, os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("cannot write to file: %v", err)
	}
	f.Close()
	
	return nil
}

// Simple heuristic to detect binary files
func isBinaryFile(content []byte) bool {
	// Check for null bytes which are common in binary files
	for _, b := range content[:min(len(content), 1024)] {
		if b == 0 {
			return true
		}
	}
	return false
}

// Find approximate line numbers where a string occurs
func findLineNumbers(content, searchStr string) []int {
	lines := strings.Split(content, "\n")
	var lineNumbers []int
	
	// Limit to first 5 occurrences to avoid huge error messages
	for i, line := range lines {
		if strings.Contains(line, searchStr) {
			lineNumbers = append(lineNumbers, i+1)
			if len(lineNumbers) >= 5 {
				break
			}
		}
	}
	
	return lineNumbers
}

// Format line numbers for display
func formatLineNumbers(numbers []int) string {
	if len(numbers) == 0 {
		return "unknown"
	}
	
	strNumbers := make([]string, len(numbers))
	for i, num := range numbers {
		strNumbers[i] = fmt.Sprintf("%d", num)
	}
	
	return strings.Join(strNumbers, ", ")
}

// min returns the smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}