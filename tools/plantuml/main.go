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

	// render_plantuml tool - now returns URLs instead of content
	renderPlantUMLTool := mcp.NewTool("render_plantuml",
		mcp.WithDescription("🌱 RENDER: Generate PlantUML diagram URLs from plain text. Returns URLs pointing to the PlantUML server for SVG/PNG rendering. Validates syntax and corrects errors using GenAI if needed."),
		mcp.WithString("plantuml_code",
			mcp.Required(),
			mcp.Description("PlantUML diagram code in plain text format (e.g., '@startuml\nAlice -> Bob: Hello\n@enduml')."),
		),
		mcp.WithString("output_format",
			mcp.Description("Output format: 'svg' for SVG URL (default), 'png' for PNG URL"),
		),
	)

	// decode_plantuml_url tool - decodes URLs back to plain text
	decodePlantUMLTool := mcp.NewTool("decode_plantuml_url",
		mcp.WithDescription("🔍 DECODE: Decode PlantUML URLs or encoded strings back to plain text PlantUML code."),
		mcp.WithString("url_or_encoded",
			mcp.Required(),
			mcp.Description("Either a full PlantUML server URL (e.g., 'http://localhost:9999/plantuml/svg/ENCODED') or just the encoded part (e.g., 'ENCODED')."),
		),
	)

	// Add tool handlers
	s.AddTool(renderPlantUMLTool, renderPlantUMLHandler)
	s.AddTool(decodePlantUMLTool, decodePlantUMLHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
