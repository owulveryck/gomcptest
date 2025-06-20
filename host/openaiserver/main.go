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
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
	"github.com/owulveryck/gomcptest/host/openaiserver/logging"
)

// Config holds the configuration parameters.
type Config struct {
	Port     int    `envconfig:"PORT" default:"8080"`
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
	if len(args) > 1 {
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
	flag.Parse()

	gcpconfig, err := loadGCPConfig()
	if err != nil {
		slog.Error("Failed to load GCP config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	ctx = logging.WithLogger(ctx, logger)
	openAIHandler := chatengine.NewOpenAIV1WithToolHandler(gcp.NewChatSession(ctx, gcpconfig), cfg.ImageDir)

	servers := extractServers(*mcpServers)
	for i := range servers {
		if servers[i] == "" {
			continue
		}
		serverCtx := logging.WithGroup(ctx, "server"+strconv.Itoa(i))

		mcpClient, err := createMCPClient(servers[i])
		if err != nil {
			logging.Error(serverCtx, "Failed to create MCP client", "error", err)
			os.Exit(1)
		}

		err = openAIHandler.AddTools(serverCtx, mcpClient)
		if err != nil {
			logging.Error(serverCtx, "Failed to add tools", "error", err)
			os.Exit(1)
		}
	}
	logging.Info(ctx, "Starting web server", "port", cfg.Port)
	http.Handle("/", openAIHandler)

	err = http.ListenAndServe(":"+strconv.Itoa(cfg.Port), nil)
	if err != nil {
		logging.Error(ctx, "Failed to start web server", "error", err)
		os.Exit(1)
	}
}
