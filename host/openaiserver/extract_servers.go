package main

import (
	"strings"
)

func extractServers(s string) [][]string {
	// Split the input string by semicolons
	commands := strings.Split(s, ";")
	result := make([][]string, 0, len(commands))

	for _, cmd := range commands {
		// Trim spaces and split each command into parts
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			parts := strings.Fields(cmd)
			result = append(result, parts)
		}
	}

	return result
}
