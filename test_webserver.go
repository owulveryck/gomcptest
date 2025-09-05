package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// Simple test to verify web server functionality
func main() {
	imageDir := "./images_edit"
	port := 8081

	// Ensure image directory exists
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		fmt.Printf("Failed to create image directory: %v\n", err)
		return
	}

	// Set up endpoints
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Test Web Server Running on Port %d", port)
	})

	portStr := strconv.Itoa(port)
	fmt.Printf("Starting test web server on port %s...\n", portStr)

	if err := http.ListenAndServe(":"+portStr, nil); err != nil {
		fmt.Printf("Failed to start web server: %v\n", err)
	}
}
