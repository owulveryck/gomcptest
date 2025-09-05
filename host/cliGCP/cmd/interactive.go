package main

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/peterh/liner"
	"google.golang.org/genai"
)

// RunInteractiveMode runs the agent in interactive mode (useful for testing)
func RunInteractiveMode(agent *DispatchAgent) {
	titleColor := color.New(color.FgCyan, color.Bold)
	promptColor := color.New(color.FgGreen, color.Bold)
	errorColor := color.New(color.FgRed, color.Bold)

	titleColor.Println("Dispatch Agent Interactive Mode")
	titleColor.Println("Type 'exit' to quit")
	// Initialize liner for command history
	line := liner.NewLiner()
	defer line.Close()

	// Read user input
	history := make([]*genai.Content, 0)
	for {
		// Use plain prompt and colorize it separately
		promptColor.Print("> ")
		input, err := line.Prompt("")

		if err == io.EOF || err == liner.ErrPromptAborted {
			titleColor.Println("\nExiting...")
			break
		}

		if err != nil {
			errorColor.Fprintf(os.Stderr, "Error reading input: %s\n", err)
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
			titleColor.Println("Exiting...")
			break
		}

		// Process the input
		history = append(history, &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{genai.NewPartFromText(input)},
		})
		response, err := agent.ProcessTask(context.Background(), history)
		if err != nil {
			errorColor.Printf("Error: %v / history: %v\n", err, history)
			continue
		}
		history = append(history, &genai.Content{
			Role:  "model",
			Parts: []*genai.Part{genai.NewPartFromText(response)},
		})

		// Print the response is handled in ProcessTask
	}
}
