package chatengine

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func (o *OpenAIV1WithToolHandler) AddTools(ctx context.Context, client client.MCPClient) error {
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "gomcptest-client",
		Version: "1.0.0",
	}

	initResult, err := client.Initialize(ctx, initRequest)
	if err != nil {
		return err
	}
	slog.Info(
		"Initialized",
		slog.String("name", initResult.ServerInfo.Name),
		slog.String("version", initResult.ServerInfo.Version),
	)

	return o.c.AddMCPTool(client)
}
