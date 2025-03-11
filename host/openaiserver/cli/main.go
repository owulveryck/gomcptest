package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

// Config holds the configuration parameters.
type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"` // Valid values: DEBUG, INFO, WARN, ERROR
	ImageDir string `envconfig:"IMAGE_DIR" required:"true"`
}

// loadGCPConfig loads and validates the GCP configuration from environment variables.
func loadGCPConfig() (gcp.Configuration, error) {
	var cfg gcp.Configuration
	err := envconfig.Process("", &cfg)
	if err != nil {
		return gcp.Configuration{}, fmt.Errorf("failed to process GCP configuration: %w", err)
	}
	if len(cfg.GeminiModels) == 0 {
		return gcp.Configuration{}, fmt.Errorf("at least one Gemini model must be specified")
	}
	for _, model := range cfg.GeminiModels {
		slog.Info("model", "model", model)
	}
	return cfg, nil
}

// createMCPClient creates an MCP client based on the provided server command and arguments.
func createMCPClient(server []string) (client.MCPClient, error) {
	if len(server) == 0 {
		return nil, fmt.Errorf("server command cannot be empty")
	}
	var mcpClient client.MCPClient
	var err error
	if len(server) > 1 {
		slog.Info("Registering", "command", server[0], "args", server[1:])
		mcpClient, err = client.NewStdioMCPClient(server[0], nil, server[1:]...)
	} else {
		slog.Info("Registering", "command", server[0])
		mcpClient, err = client.NewStdioMCPClient(server[0], nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}
	return mcpClient, nil
}

// extractServers parses the mcpservers flag value.
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

func main() {
	// Set up logging
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing configuration: %v\n", err)
		os.Exit(1)
	}

	// Configure logging
	var logLevel slog.Level
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
		fmt.Printf("Invalid debug level specified (%v), defaulting to INFO\n", cfg.LogLevel)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Parse command line flags
	mcpServers := flag.String("mcpservers", "", "Input string of MCP servers (semicolon-separated)")
	modelFlag := flag.String("model", "", "Specific model to use (overrides environment variable)")
	flag.Parse()

	// Load GCP configuration
	gcpConfig, err := loadGCPConfig()
	if err != nil {
		slog.Error("Failed to load GCP config", "error", err)
		os.Exit(1)
	}

	// Override model if specified in command line
	if *modelFlag != "" {
		gcpConfig.GeminiModels = []string{*modelFlag}
		slog.Info("Using model from command line", "model", *modelFlag)
	}

	// Initialize chat session
	ctx := context.Background()
	slog.Info("Initializing chat session...")
	chatSession := gcp.NewChatSession(ctx, gcpConfig)

	// Register MCP tools
	servers := extractServers(*mcpServers)
	for i, server := range servers {
		slog.Info("Registering MCP server", "index", i, "server", strings.Join(server, " "))
		
		mcpClient, err := createMCPClient(server)
		if err != nil {
			slog.Error("Failed to create MCP client", "error", err)
			os.Exit(1)
		}

		// Initialize the client
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "gomcptest-client",
			Version: "1.0.0",
		}

		_, err = mcpClient.Initialize(ctx, initRequest)
		if err != nil {
			slog.Error("Failed to initialize MCP client", "error", err)
			os.Exit(1)
		}
		slog.Info("MCP client initialized successfully")

		err = chatSession.AddMCPTool(mcpClient)
		if err != nil {
			slog.Error("Failed to add MCP tool", "error", err)
			os.Exit(1)
		}
		slog.Info("MCP tools registered successfully", "server", i)
	}

	// Get default model
	defaultModel := gcpConfig.GeminiModels[0]
	slog.Info("Using default model", "model", defaultModel)

	// Initialize conversation history
	messages := []chatengine.ChatCompletionMessage{}

	// REPL loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\nChat initialized. Type your messages (type 'exit' to quit):")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		userInput := scanner.Text()
		if userInput == "exit" {
			break
		}

		// Add user message to history
		userMessage := chatengine.ChatCompletionMessage{
			Role:    "user",
			Content: userInput,
		}
		messages = append(messages, userMessage)

		// Create chat completion request
		req := chatengine.ChatCompletionRequest{
			Model:       defaultModel,
			Messages:    messages,
			Temperature: 0.7,
			Stream:      false, // Non-streaming mode
		}

		// Send request to chat session
		resp, err := chatSession.HandleCompletionRequest(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Process response
		if len(resp.Choices) > 0 {
			assistantResponse := resp.Choices[0].Message
			fmt.Printf("Assistant: %s\n", assistantResponse.Content)

			// Add assistant response to history
			messages = append(messages, chatengine.ChatCompletionMessage{
				Role:    "assistant",
				Content: assistantResponse.Content,
			})
		} else {
			fmt.Println("No response received")
		}
	}
}