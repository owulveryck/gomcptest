package main

import (
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		return
	}

	// Setup logger
	if err := setupLogger(cfg); err != nil {
		fmt.Printf("Failed to setup logger: %v\n", err)
		return
	}

	slog.Info("Starting PlantUML Renderer", "version", "1.0.0")

	// Create a new MCP server for PlantUML rendering
	s := server.NewMCPServer(
		"PlantUML Renderer 🌱",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// render_plantuml tool
	renderPlantUMLTool := mcp.NewTool("render_plantuml",
		mcp.WithDescription("🌱 RENDER: Render PlantUML diagrams from plain text or encoded format. Supports multiple output formats including SVG, PNG, plain text, and encoded versions."),
		mcp.WithString("plantuml_code",
			mcp.Required(),
			mcp.Description("PlantUML diagram code in plain text format (e.g., '@startuml\nAlice -> Bob: Hello\n@enduml') or in encoded format (base64-like encoded string). The tool will automatically detect the format."),
		),
		mcp.WithString("output_format",
			mcp.Description("Output format: 'svg' for SVG vector graphics (default), 'png' for PNG image, 'txt' for plain text representation, 'encoded' for PlantUML encoded format"),
		),
	)

	// Add tool handler
	s.AddTool(renderPlantUMLTool, renderPlantUMLHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
