package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Get default tool paths
	toolPaths := DefaultToolPaths()

	// Parse command line flags
	var interactive bool
	flag.BoolVar(&interactive, "interactive", false, "Run in interactive mode")
	flag.StringVar(&toolPaths.ViewPath, "view-path", toolPaths.ViewPath, "Path to View tool executable")
	flag.StringVar(&toolPaths.GlobPath, "glob-path", toolPaths.GlobPath, "Path to GlobTool executable")
	flag.StringVar(&toolPaths.GrepPath, "grep-path", toolPaths.GrepPath, "Path to GrepTool executable")
	flag.StringVar(&toolPaths.LSPath, "ls-path", toolPaths.LSPath, "Path to LS executable")
	flag.Parse()

	// Log the tool paths that will be used
	fmt.Printf("Using tool paths:\n")
	fmt.Printf("  View: %s\n", toolPaths.ViewPath)
	fmt.Printf("  GlobTool: %s\n", toolPaths.GlobPath)
	fmt.Printf("  GrepTool: %s\n", toolPaths.GrepPath)
	fmt.Printf("  LS: %s\n", toolPaths.LSPath)

	// Initialize agent to verify tools at startup
	agent, err := NewDispatchAgent()
	if err != nil {
		fmt.Printf("Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// Register tools once at startup - exit if it fails
	err = agent.RegisterTools(context.Background(), toolPaths)
	if err != nil {
		fmt.Printf("Failed to initialize tools: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All tools successfully initialized")

	// If interactive mode is requested, run the agent in interactive mode
	if interactive {
		RunInteractiveMode(agent)
		return
	}

	// Create MCP server
	s := server.NewMCPServer(
		"dispatch_agent 🤖",
		"1.0.0",
	)

	// Create dispatch_agent tool
	tool := CreateDispatchTool(agent)

	// Create handler and add tool to server
	dispatchHandler := CreateDispatchHandler(agent)
	s.AddTool(tool, dispatchHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
