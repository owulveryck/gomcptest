package gcp

import (
	"log/slog"
	"strconv"

	"github.com/mark3labs/mcp-go/client"
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
func (chatsession *ChatSession) AddMCPTool(mcpClient client.MCPClient) error {
	// define servername
	mcpServerName := serverPrefix + strconv.Itoa(len(chatsession.servers))
	err := chatsession.addMCPTool(mcpClient, mcpServerName)
	if err != nil {
		slog.Info("cannot register tools for server", "message from MCP Server", err.Error())
	}
	err = chatsession.addMCPResourceTemplate(mcpClient, mcpServerName)
	if err != nil {
		slog.Info("cannot register resources template for server", "message from MCP Server", err.Error())
	}
	err = chatsession.addMCPPromptTemplate(mcpClient, mcpServerName)
	if err != nil {
		slog.Info("cannot register resources template for server", "message from MCP Server", err.Error())
	}
	chatsession.servers = append(chatsession.servers, &MCPServerTool{
		mcpClient: mcpClient,
	})

	return nil
}
