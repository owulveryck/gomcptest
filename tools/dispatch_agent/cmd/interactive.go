package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/vertexai/genai"
)

// RunInteractiveMode runs the agent in interactive mode (useful for testing)
func RunInteractiveMode(agent *DispatchAgent) {
	fmt.Println("Dispatch Agent Interactive Mode")
	fmt.Println("Type 'exit' to quit")

	// Read user input
	scanner := bufio.NewScanner(os.Stdin)
	history := make([]*genai.Content, 0)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "exit" {
			break
		}

		// Process the input
		history = append(history, &genai.Content{
			Role:  "user",
			Parts: []genai.Part{genai.Text(input)},
		})
		response, err := agent.ProcessTask(context.Background(), history)
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
