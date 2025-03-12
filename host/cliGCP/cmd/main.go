package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Get default tool paths

	// Parse command line flags
	mcpServers := flag.String("mcpservers", "", "Input string of MCP servers")
	flag.Parse()

	// Initialize agent to verify tools at startup
	agent, err := NewDispatchAgent()
	if err != nil {
		fmt.Printf("Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// Register tools once at startup - exit if it fails
	err = agent.RegisterTools(context.Background(), *mcpServers)
	if err != nil {
		fmt.Printf("Failed to initialize tools: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All tools successfully initialized")

	// If interactive mode is requested, run the agent in interactive mode
	RunInteractiveMode(agent)
}

