package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"PlantUML Check 🌱",
		"1.0.0",
	)

	// Add PlantUML check tool
	tool := mcp.NewTool("plantuml_check",
		mcp.WithDescription("Validates PlantUML file syntax using the PlantUML jar file. This tool checks if a PlantUML file has valid syntax by attempting to process it with the PlantUML processor.\n\nEnvironment Variables:\n- PLANTUML_JAR: Path to the PlantUML jar file (required)\n\nUsage:\n1. file_path: The absolute path to the PlantUML file to validate (.puml, .plantuml, or .pu files)\n\nThe tool will:\n- Check if the PlantUML jar file exists at the specified path\n- Validate the syntax of the provided PlantUML file\n- Return validation results with any syntax errors or confirmation of valid syntax\n\nRequirements:\n- Java must be installed and available in PATH\n- PLANTUML_JAR environment variable must point to a valid PlantUML jar file\n- The file to check must exist and be readable"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the PlantUML file to validate"),
		),
	)

	// Add tool handler
	s.AddTool(tool, plantumlCheckHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func plantumlCheckHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// First convert Arguments to map[string]interface{}
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, errors.New("arguments must be a map")
	}

	// Extract parameters
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, errors.New("file_path must be a string")
	}

	// Validate that the path is absolute
	if !filepath.IsAbs(filePath) {
		return nil, errors.New("file_path must be an absolute path, not a relative path")
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("File does not exist: %s", filePath))
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return nil, errors.New(fmt.Sprintf("%s is a directory, not a file", filePath))
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".puml" && ext != ".plantuml" && ext != ".pu" {
		return nil, errors.New(fmt.Sprintf("File must have a PlantUML extension (.puml, .plantuml, or .pu), got: %s", ext))
	}

	// Get PlantUML jar path from environment
	jarPath := os.Getenv("PLANTUML_JAR")
	if jarPath == "" {
		return nil, errors.New("PLANTUML_JAR environment variable is required and must point to the PlantUML jar file")
	}

	// Check if jar file exists
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("PlantUML jar file not found at: %s", jarPath))
	}

	// Check if java is available
	if _, err := exec.LookPath("java"); err != nil {
		return nil, errors.New("Java is not installed or not available in PATH")
	}

	// Validate PlantUML syntax
	result, err := validatePlantUMLSyntax(jarPath, filePath)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

func validatePlantUMLSyntax(jarPath, filePath string) (string, error) {
	// Use PlantUML's syntax check mode
	// The -checkonly flag makes PlantUML only check syntax without generating output
	cmd := exec.Command("java", "-jar", jarPath, "-checkonly", filePath)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		// PlantUML returns non-zero exit code for syntax errors
		if exitError, ok := err.(*exec.ExitError); ok {
			// Parse the output to provide meaningful error messages
			if strings.Contains(outputStr, "Syntax Error") ||
				strings.Contains(outputStr, "error") ||
				strings.Contains(outputStr, "Error") {
				return fmt.Sprintf("❌ PlantUML syntax validation failed for %s:\n\n%s", filepath.Base(filePath), outputStr), nil
			}
			return fmt.Sprintf("❌ PlantUML validation failed with exit code %d:\n\n%s", exitError.ExitCode(), outputStr), nil
		}
		// Other execution errors
		return "", fmt.Errorf("failed to execute PlantUML: %v", err)
	}

	// If we get here, validation was successful
	if outputStr != "" {
		return fmt.Sprintf("✅ PlantUML syntax is valid for %s\n\nOutput:\n%s", filepath.Base(filePath), outputStr), nil
	}

	return fmt.Sprintf("✅ PlantUML syntax is valid for %s", filepath.Base(filePath)), nil
}
