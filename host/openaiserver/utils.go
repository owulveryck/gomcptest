// Logic from point #10 - trying again
package main

import (
	"strings"
)

func parseCommandString(input string) (cmd string, env []string, args []string) {
	fields := strings.Fields(input)
	env = []string{}  // Ensure initialized
	args = []string{} // Ensure initialized

	if len(fields) == 0 {
		return "", env, args
	}

	cmdIndex := len(fields) // Default: assume no command found

	for i, field := range fields {
		isEnvVar := strings.Contains(field, "=") && strings.Index(field, "=") > 0
		if !isEnvVar {
			// Found the first field that is NOT an env var. This is the command.
			cmdIndex = i
			break // Stop searching
		}
	}

	// Assign based on the final cmdIndex
	if cmdIndex == len(fields) { // No command found
		cmd = ""
		env = fields // All fields were env vars
	} else { // Command found at cmdIndex
		env = fields[:cmdIndex]
		cmd = fields[cmdIndex]
		// Args are everything after cmdIndex
		if cmdIndex+1 < len(fields) {
			args = fields[cmdIndex+1:]
		}
	}

	return cmd, env, args
}
