package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/mark3labs/mcp-go/client"

	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

func main() {
	var gcpconfig gcp.Configuration
	ctx := context.Background()

	err := envconfig.Process("", &gcpconfig)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	if len(gcpconfig.GeminiModels) == 0 {
		slog.Error("please specify at least one model")
		os.Exit(1)
	}
	for _, model := range gcpconfig.GeminiModels {
		slog.Info("model", "model", model)
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// handler := slog.NewJSONHandler(os.Stdout, opts)
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	mcpServers := flag.String("mcpservers", "", "Input string of MCP servers")
	flag.Parse()

	openAIHandler := chatengine.NewOpenAIV1WithToolHandler(gcp.NewChatSession(ctx, gcpconfig))
	// openAIHandler := chatengine.NewOpenAIV1WithToolHandler(ollama.NewEngine())
	servers := extractServers(*mcpServers)
	for i := range servers {
		logger := logger.WithGroup("server" + strconv.Itoa(i))
		slog.SetDefault(logger)
		var mcpClient client.MCPClient
		var err error
		if len(servers[i]) > 1 {
			logger.Info("Registering", "command", servers[i][0], "args", servers[i][1:])
			mcpClient, err = client.NewStdioMCPClientLog(servers[i][0], nil, servers[i][1:]...)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}
		} else {
			logger.Info("Registering", "command", servers[i][0])
			mcpClient, err = client.NewStdioMCPClientLog(servers[i][0], nil)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}

		}
		err = openAIHandler.AddTools(ctx, mcpClient)
		if err != nil {
			logger.Error("Failed to add tools", "error", err)
			os.Exit(1)
		}
	}
	slog.SetDefault(logger)

	port := "8080"
	slog.Info("Starting web server", "port", 8080)
	// http.Handle("/", GzipMiddleware(openAIHandler)) // Wrap the handler with the gzip middleware
	http.Handle("/", openAIHandler) // Wrap the handler with the gzip middleware

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		slog.Error("Failed to start web server", "error", err)
		os.Exit(1)
	}
}
