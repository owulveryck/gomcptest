package chatengine

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/owulveryck/gomcptest/host/openaiserver/logging"
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
	logging.Info(ctx,
		"Initialized",
		"name", initResult.ServerInfo.Name,
		"version", initResult.ServerInfo.Version,
	)

	return o.c.AddMCPTool(ctx, client)
}
