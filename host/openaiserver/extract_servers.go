package main

import (
	"strings"
)

func extractServers(s string) []string {
	// Split the input string by semicolons
	return strings.Split(s, ";")
}
