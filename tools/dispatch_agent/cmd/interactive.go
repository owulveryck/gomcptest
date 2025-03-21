package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/peterh/liner"
)

// RunInteractiveMode runs the agent in interactive mode (useful for testing)
func RunInteractiveMode(agent *DispatchAgent) {
	fmt.Println("Dispatch Agent Interactive Mode")
	fmt.Println("Type 'exit' to quit")
	fmt.Println("You can specify a working directory with '--path=/your/path prompt'")
	// Initialize liner for command history
	line := liner.NewLiner()
	defer line.Close()

	// Read user input
	history := make([]*genai.Content, 0)
	for {

		input, err := line.Prompt("> ")

		if err == io.EOF || err == liner.ErrPromptAborted {
			fmt.Println("\nExiting...")
			break
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %s\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Add non-empty inputs to liner history
		line.AppendHistory(input)

		// Handle exit command
		if input == "exit" {
			fmt.Println("Exiting...")
			break
		}

		// Check for path flag
		var workingPath string
		if strings.HasPrefix(input, "--path=") {
			parts := strings.SplitN(input, " ", 2)
			if len(parts) == 2 {
				pathFlag := parts[0]
				workingPath = strings.TrimPrefix(pathFlag, "--path=")
				input = parts[1]
			} else {
				fmt.Println("Please provide a prompt after the --path flag")
				continue
			}
		}

		// Process the input
		history = append(history, &genai.Content{
			Role:  "user",
			Parts: []genai.Part{genai.Text(input)},
		})
		response, err := agent.ProcessTask(context.Background(), history, workingPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		history = append(history, &genai.Content{
			Role:  "model",
			Parts: []genai.Part{genai.Text(response)},
		})

		// Print the response
		fmt.Println(response)
	}
}
