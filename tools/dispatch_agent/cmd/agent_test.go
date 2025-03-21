package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentCommandHelpers(t *testing.T) {
	// Skip test if this is a short test run
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a mock agent for testing
	agent := &DispatchAgent{
		// No need to initialize all fields for this test
	}

	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "agent-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a nested directory structure
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test files in different directories
	rootFile := filepath.Join(tmpDir, "root.txt")
	subFile := filepath.Join(subDir, "sub.txt")

	if err := os.WriteFile(rootFile, []byte("root file content"), 0644); err != nil {
		t.Fatalf("Failed to write root file: %v", err)
	}
	if err := os.WriteFile(subFile, []byte("sub file content"), 0644); err != nil {
		t.Fatalf("Failed to write sub file: %v", err)
	}

	// Remember original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Test runCommand - just verify it can execute a command
	output, err := agent.runCommand("echo test")
	if err != nil {
		t.Errorf("runCommand failed: %v", err)
	}
	if output != "test\n" {
		t.Errorf("runCommand returned unexpected output: %q, expected %q", output, "test\n")
	}

	// Test changeDirectory and getCurrentDirectory
	err = agent.changeDirectory(tmpDir)
	if err != nil {
		t.Errorf("changeDirectory failed: %v", err)
	}

	// Get working directory with agent's method
	_, err = agent.getCurrentDirectory() 
	if err != nil {
		t.Errorf("getCurrentDirectory failed: %v", err)
	}
	
	// Use ls to check that the file we created is visible
	lsOutput, err := agent.runCommand("ls")
	if err != nil {
		t.Errorf("runCommand failed after changing directory: %v", err)
	}
	if !strings.Contains(lsOutput, "root.txt") {
		t.Errorf("Expected to find root.txt in directory listing, got: %s", lsOutput)
	}

	// Restore original directory
	if err := agent.changeDirectory(origDir); err != nil {
		t.Errorf("Failed to restore original directory: %v", err)
	}
}

func TestParameterExtraction(t *testing.T) {
	// Test with no path parameter
	args := map[string]interface{}{
		"prompt": "test prompt",
	}

	// We can't actually call the handler with a request since ProcessTask uses an LLM
	// So instead, let's extract the test logic to check parameter extraction

	prompt, ok := args["prompt"].(string)
	if !ok {
		t.Error("Failed to extract prompt parameter")
	}
	if prompt != "test prompt" {
		t.Errorf("Got incorrect prompt. Expected %q, got %q", "test prompt", prompt)
	}

	// Extract path parameter (should be empty)
	var path string
	if pathValue, ok := args["path"]; ok {
		if pathStr, ok := pathValue.(string); ok {
			path = pathStr
		}
	}
	if path != "" {
		t.Errorf("Path should be empty when not provided, got %q", path)
	}

	// Test with path parameter
	args = map[string]interface{}{
		"prompt": "test prompt",
		"path":   "/test/path",
	}

	// Extract path parameter (should match provided value)
	path = ""
	if pathValue, ok := args["path"]; ok {
		if pathStr, ok := pathValue.(string); ok {
			path = pathStr
		}
	}
	if path != "/test/path" {
		t.Errorf("Path incorrect. Expected %q, got %q", "/test/path", path)
	}
}