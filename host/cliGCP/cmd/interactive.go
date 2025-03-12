package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/peterh/liner"
)

// RunInteractiveMode runs the agent in interactive mode (useful for testing)
func RunInteractiveMode(agent *DispatchAgent) {
	fmt.Println("Dispatch Agent Interactive Mode")
	fmt.Println("Type 'exit' to quit")
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

		// Process the input
		history = append(history, &genai.Content{
			Role:  "user",
			Parts: []genai.Part{genai.Text(input)},
		})
		response, err := agent.ProcessTask(context.Background(), history)
		if err != nil {
			log.Printf("Error: %v / history: %v\n", err, history)
			continue
		}
		history = append(history, &genai.Content{
			Role:  "model",
			Parts: []genai.Part{genai.Text(response)},
		})

		// Print the response
		// fmt.Println(response)
	}
}
