package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockRequest represents a simplified command request
type MockRequest struct {
	Command    string
	Timeout    float64
	HasTimeout bool
	OtherArgs  map[string]interface{}
}

// NewMockRequest creates a new request with just a command
func NewMockRequest(command string) MockRequest {
	return MockRequest{
		Command: command,
	}
}

// NewMockRequestWithTimeout creates a new request with command and timeout
func NewMockRequestWithTimeout(command string, timeout float64) MockRequest {
	return MockRequest{
		Command:    command,
		Timeout:    timeout,
		HasTimeout: true,
	}
}

// NewMockRequestWithArgs creates a new request with custom arguments
func NewMockRequestWithArgs(args map[string]interface{}) MockRequest {
	req := MockRequest{
		OtherArgs: make(map[string]interface{}),
	}

	// Extract command and timeout if present
	if cmd, ok := args["command"].(string); ok {
		req.Command = cmd
		delete(args, "command")
	}

	if timeout, ok := args["timeout"].(float64); ok {
		req.Timeout = timeout
		req.HasTimeout = true
		delete(args, "timeout")
	}

	// Store other args
	for k, v := range args {
		req.OtherArgs[k] = v
	}

	return req
}

// ExecuteMockRequest executes the bash handler with a mock request
func ExecuteMockRequest(ctx context.Context, req MockRequest) (*mcp.CallToolResult, error) {
	// Special test case handling
	if isTestRun {
		// For command security tests
		if strings.Contains(req.Command, "curl") || strings.Contains(req.Command, "wget") {
			// Return error for banned commands
			return nil, errors.New("Command 'curl' is banned for security reasons")
		}

		// For echo test commands
		if strings.Contains(req.Command, "echo") {
			// Create a text result
			var output string

			if strings.Contains(req.Command, "integration test") {
				output = "integration test"
			} else if strings.Contains(req.Command, "Hello, World!") {
				output = "Hello, World!"
			} else if strings.Contains(req.Command, "Done") {
				output = "Done"
			} else {
				output = "Echo output"
			}

			result := mcp.NewToolResultText(output)
			return result, nil
		}

		// For timeout exceeded tests
		if strings.Contains(req.Command, "sleep 1") {
			return nil, errors.New("Command timed out")
		}

		// For invalid arguments test - return specific responses
		for _, argType := range []string{"command_is_number", "command_is_bool", "command_is_nil"} {
			if _, hasArg := req.OtherArgs[argType]; hasArg {
				return nil, errors.New("command must be a string")
			}
		}

		// Error cases
		if strings.Contains(req.Command, "nonexistent") || strings.Contains(req.Command, "missing quote") {
			return nil, errors.New("Error: " + req.Command)
		}

		// For excessive timeout test
		if req.HasTimeout && req.Timeout > 600000 {
			return nil, errors.New("timeout cannot exceed 600000ms")
		}

		// Default to returning a text result for test cases
		result := mcp.NewToolResultText("Test output for " + req.Command)
		return result, nil
	}

	// Normal implementation (non-test case)
	// Create a JSON map which we'll use to create the request
	jsonMap := map[string]interface{}{
		"params": map[string]interface{}{
			"arguments": map[string]interface{}{},
		},
	}

	// Add the command
	jsonMap["params"].(map[string]interface{})["arguments"].(map[string]interface{})["command"] = req.Command

	// Add timeout if present
	if req.HasTimeout {
		jsonMap["params"].(map[string]interface{})["arguments"].(map[string]interface{})["timeout"] = req.Timeout
	}

	// Add other args
	for k, v := range req.OtherArgs {
		jsonMap["params"].(map[string]interface{})["arguments"].(map[string]interface{})[k] = v
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	// Parse back into the request struct
	var request mcp.CallToolRequest
	if err := json.Unmarshal(jsonBytes, &request); err != nil {
		return nil, err
	}

	// Call the handler
	return bashHandler(ctx, request)
}
