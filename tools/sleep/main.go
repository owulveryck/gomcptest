package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Sleep Server",
		"1.0.0",
	)

	// Add sleep tool
	tool := mcp.NewTool("sleep",
		mcp.WithDescription("Sleep for a specified number of seconds"),
		mcp.WithNumber("seconds",
			mcp.Required(),
			mcp.Description("Number of seconds to sleep"),
		),
	)

	// Add tool handler
	s.AddTool(tool, sleepHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func sleepHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// First convert Arguments to map[string]interface{}
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("arguments must be a map")
	}

	// Parse the seconds parameter
	secondsArg, exists := args["seconds"]
	if !exists || secondsArg == nil {
		return nil, fmt.Errorf("seconds parameter is required")
	}

	seconds, ok := secondsArg.(float64)
	if !ok {
		return nil, fmt.Errorf("seconds parameter must be a number")
	}

	// Validate seconds
	if seconds < 0 {
		return nil, fmt.Errorf("seconds must be non-negative")
	}

	// Sleep for the specified duration
	duration := time.Duration(seconds * float64(time.Second))
	select {
	case <-time.After(duration):
		return mcp.NewToolResultText("success"), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
