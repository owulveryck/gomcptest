package main

import (
	"os"
	"path/filepath"
)

func getCWD() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Abs(cwd) // Convert to absolute path
}
