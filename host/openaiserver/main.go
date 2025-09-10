package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/client"

	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/vertexai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/vertexai/gemini"
)

// Config holds the configuration parameters.
type Config struct {
	Port     int    `envconfig:"PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"` // Valid values: DEBUG, INFO, WARN, ERROR
}

// loadGCPConfig loads and validates the GCP configuration from environment variables.
func loadGCPConfig() (vertexai.Configuration, error) {
	var cfg vertexai.Configuration
	err := envconfig.Process("", &cfg)
	if err != nil {
		return vertexai.Configuration{}, fmt.Errorf("failed to process GCP configuration: %w", err)
	}
	if len(cfg.GeminiModels) == 0 {
		return vertexai.Configuration{}, fmt.Errorf("at least one Gemini model must be specified")
	}
	for _, model := range cfg.GeminiModels {
		slog.Info("model", "model", model)
	}
	return cfg, nil
}

// createMCPClient creates an MCP client based on the provided server command and arguments.
func createMCPClient(server string) (client.MCPClient, error) {
	if len(server) == 0 {
		return nil, fmt.Errorf("server command cannot be empty")
	}
	var mcpClient client.MCPClient
	var err error
	cmd, env, args := parseCommandString(server)
	if len(env) == 0 {
		env = nil
	}
	if len(args) > 0 {
		// TODO: process environment variables
		slog.Info("Registering", "command", cmd, "args", args, "env", env)
		mcpClient, err = client.NewStdioMCPClient(cmd, env, args...)
	} else {
		slog.Info("Registering", "command", cmd, "env", env)
		mcpClient, err = client.NewStdioMCPClient(cmd, env)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}
	return mcpClient, nil
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		slog.Error("Failed to process configuration", "error", err)
		os.Exit(1)
	}

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
		logLevel = slog.LevelInfo // Default to INFO if the value is invalid
		log.Printf("Invalid debug level specified (%v), defaulting to INFO", cfg.LogLevel)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	mcpServers := flag.String("mcpservers", "", "Input string of MCP servers")
	withAllEvents := flag.Bool("withAllEvents", false, "Include all events (tool calls, tool responses) in stream output, not just content chunks")
	flag.Parse()

	gcpconfig, err := loadGCPConfig()
	if err != nil {
		slog.Error("Failed to load GCP config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	openAIHandler := chatengine.NewOpenAIV1WithToolHandlerWithOptions(gemini.NewChatSession(ctx, gcpconfig), *withAllEvents)
	// openAIHandler := chatengine.NewOpenAIV1WithToolHandlerWithOptions(claude.NewChatSession(ctx, gcpconfig), *withAllEvents)

	servers := extractServers(*mcpServers)
	for i := range servers {
		if servers[i] == "" {
			continue
		}
		logger := logger.WithGroup("server" + strconv.Itoa(i))
		slog.SetDefault(logger)

		mcpClient, err := createMCPClient(servers[i])
		if err != nil {
			slog.Error("Failed to create MCP client", "error", err)
			os.Exit(1)
		}

		err = openAIHandler.AddTools(ctx, mcpClient)
		if err != nil {
			slog.Error("Failed to add tools", "error", err)
			os.Exit(1)
		}
	}
	slog.SetDefault(logger)

	slog.Info("Starting web server", "port", cfg.Port)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("/ui", ServeUI)
	mux.HandleFunc("/ui/", ServeUI)
	mux.HandleFunc("/favicon.svg", ServeFavicon)
	mux.HandleFunc("/apple-touch-icon-180x180.png", ServeAppleTouchIcon)
	mux.Handle("/", openAIHandler)

	// Wrap the entire mux with CORS middleware
	corsHandler := CORSMiddleware(mux)

	err = http.ListenAndServe(":"+strconv.Itoa(cfg.Port), corsHandler)
	if err != nil {
		slog.Error("Failed to start web server", "error", err)
		os.Exit(1)
	}
}
