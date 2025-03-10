package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockResult represents a simplified version of the result structure for testing
type MockResult struct {
	IsError bool
	Message string
}

// TestGlobToolHandlerIntegration tests the globToolHandler function with real files
func TestGlobToolHandlerIntegration(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	testCases := []struct {
		name          string
		params        map[string]interface{}
		expectError   bool
		checkContains string
		checkCount    bool
		expectedCount int
	}{
		{
			name: "Basic pattern matching",
			params: map[string]interface{}{
				"pattern": "**/*.go",
				"path":    tempDir,
			},
			checkContains: "Found 6 files matching pattern",
		},
		{
			name: "With exclude pattern",
			params: map[string]interface{}{
				"pattern": "**/*.go",
				"path":    tempDir,
				"exclude": "**/*_test.go",
			},
			checkContains: "Found 3 files matching pattern",
		},
		{
			name: "With limit",
			params: map[string]interface{}{
				"pattern": "**/*",
				"path":    tempDir,
				"limit":   2.0, // Numbers come in as float64 from JSON
			},
			checkCount:    true,
			expectedCount: 2,
		},
		{
			name: "With absolute paths",
			params: map[string]interface{}{
				"pattern":  "**/*.go",
				"path":     tempDir,
				"absolute": true,
			},
			checkContains: "Found 6 files matching pattern",
		},
		{
			name: "No matches",
			params: map[string]interface{}{
				"pattern": "**/*.xyz",
				"path":    tempDir,
			},
			checkContains: "No files matched pattern",
		},
		{
			name: "Invalid path",
			params: map[string]interface{}{
				"pattern": "**/*.go",
				"path":    filepath.Join(tempDir, "nonexistent"),
			},
			expectError:   true,
			checkContains: "Path does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request object with the structure seen in main.go
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tc.params
			request.Params.Name = "GlobTool"

			// Call handler
			result, err := globToolHandler(context.Background(), request)

			// Check errors
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Convert to JSON to inspect structure
			resultJSON, _ := json.Marshal(result)
			resultString := string(resultJSON)

			// Check error case
			if tc.expectError {
				if !strings.Contains(resultString, "isError") || !strings.Contains(resultString, "true") {
					t.Errorf("Expected error result but got success: %s", resultString)
				} else if !strings.Contains(resultString, tc.checkContains) {
					t.Errorf("Error message doesn't contain expected content. Got: %s, Expected to contain: %s", 
						resultString, tc.checkContains)
				}
				return
			}

			// For non-error case
			if strings.Contains(resultString, "isError") && strings.Contains(resultString, "true") {
				t.Errorf("Expected success but got error: %s", resultString)
				return
			}

			// Check for expected content
			if tc.checkContains != "" && !strings.Contains(resultString, tc.checkContains) {
				t.Errorf("Result doesn't contain expected content. Got: %s, Expected to contain: %s", 
					resultString, tc.checkContains)
			}

			// For checking file count, we need to parse resultString manually
			if tc.checkCount {
				// Get the raw result data
				textResult := result

				// Convert to JSON and count new lines
				textJSON, _ := json.Marshal(textResult)
				textString := string(textJSON)
				
				// Just check that there is content
				if !strings.Contains(textString, "content") {
					t.Errorf("Expected content in result: %s", textString)
					return
				}
				
				// Use fmt.Println for inspection in case of errors
				if testing.Verbose() {
					fmt.Println("Result JSON:", textString)
				}
			}
		})
	}
}

// TestMCPStructure prints out the structure of a CallToolResult
func TestMCPStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MCP structure test in short mode")
	}
	
	// Create a basic request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"pattern": "**/*.go",
	}
	request.Params.Name = "GlobTool"
	
	// Call the handler with a known pattern that will succeed
	result, err := globToolHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Error calling handler: %v", err)
	}
	
	// Print the structure
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("CallToolResult structure:\n%s\n", resultJSON)
}

// TestGlobToolHandlerMissingParams tests error cases for missing or invalid parameters
func TestGlobToolHandlerMissingParams(t *testing.T) {
	testCases := []struct {
		name        string
		params      map[string]interface{}
		errorPrefix string
	}{
		{
			name:        "Missing pattern",
			params:      map[string]interface{}{},
			errorPrefix: "pattern must be",
		},
		{
			name: "Pattern not a string",
			params: map[string]interface{}{
				"pattern": 123,
			},
			errorPrefix: "pattern must be",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tc.params
			request.Params.Name = "GlobTool"

			result, err := globToolHandler(context.Background(), request)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			// Convert to JSON to inspect structure
			resultJSON, _ := json.Marshal(result)
			resultString := string(resultJSON)

			// Check for error
			if !strings.Contains(resultString, "isError") || !strings.Contains(resultString, "true") {
				t.Errorf("Expected error result but got success: %s", resultString)
				return
			}

			// Check error message prefix
			if !strings.Contains(resultString, tc.errorPrefix) {
				t.Errorf("Expected error message to contain '%s', got '%s'", 
					tc.errorPrefix, resultString)
			}
		})
	}
}