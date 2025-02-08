package gcp

import (
	"github.com/mark3labs/mcp-go/client"
)

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (chatsession *ChatSession) AddMCPTool(c client.MCPClient) error {
	return nil
}
