package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

// RunInteractiveMode runs the agent in interactive mode (useful for testing)
func RunInteractiveMode(agent *DispatchAgent) {
	fmt.Println("Dispatch Agent Interactive Mode")
	fmt.Println("Type 'exit' to quit")

	// Read user input
	scanner := bufio.NewScanner(os.Stdin)
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
		response, err := agent.ProcessTask(context.Background(), input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Print the response
		fmt.Println(response)
	}
}