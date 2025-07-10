package gcp

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/mark3labs/mcp-go/client"
	"github.com/owulveryck/gomcptest/host/openaiserver/logging"
)

const (
	serverPrefix           = "MCP"
	resourcePrefix         = "resource"
	resourceTemplatePrefix = "resourceTemplate"
	toolPrefix             = "tool"
	promptPrefix           = "prompt"
)

type MCPServerTool struct {
	mcpClient client.MCPClient
}

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (chatsession *ChatSession) AddMCPTool(ctx context.Context, mcpClient client.MCPClient) error {
	// define servername
	mcpServerName := serverPrefix + strconv.Itoa(len(chatsession.servers))
	err := chatsession.addMCPTool(ctx, mcpClient, mcpServerName)
	if err != nil {
		logging.Info(ctx, "cannot register tools for server", "message from MCP Server", err.Error())
	}
	err = chatsession.addMCPResourceTemplate(ctx, mcpClient, mcpServerName)
	if err != nil {
		logging.Info(ctx, "cannot register resources template for server", "message from MCP Server", err.Error())
	}
	err = chatsession.addMCPResource(mcpClient, mcpServerName)
	if err != nil {
		slog.Info("cannot register resources for server", "message from MCP Server", err.Error())
	}
	err = chatsession.addMCPPromptTemplate(mcpClient, mcpServerName)
	if err != nil {
		logging.Info(ctx, "cannot register prompt for server", "message from MCP Server", err.Error())
	}
	chatsession.servers = append(chatsession.servers, &MCPServerTool{
		mcpClient: mcpClient,
	})

	return nil
}
