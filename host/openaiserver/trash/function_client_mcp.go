package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/vertexai/genai"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPServerTool struct {
	name   string
	client *client.StdioMCPClient
}

func (findmcplogs *MCPServerTool) GetGenaiTool() *genai.Tool {
	ctx := context.Background()
	var err error
	findmcplogs.client, err = client.NewStdioMCPClientLog(config.MCPServerSample, strings.Split(config.MCPServerArgs, " "))
	if err != nil {
		log.Fatal(err)
	}
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "gomcptest-client",
		Version: "1.0.0",
	}

	initResult, err := findmcplogs.client.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	fmt.Printf(
		"Initialized with server: %s %s\n\n",
		initResult.ServerInfo.Name,
		initResult.ServerInfo.Version,
	)
	// List Tools
	fmt.Println("Listing available tools...")
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := findmcplogs.client.ListTools(ctx, toolsRequest)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}
	schema := &genai.Schema{
		Type:       genai.TypeObject,
		Properties: make(map[string]*genai.Schema),
	}
	if len(tools.Tools) != 1 {
		log.Fatal("only one tool supported so fat")
	}
	for _, tool := range tools.Tools {
		fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
		for k, v := range tool.InputSchema.Properties {
			v := v.(map[string]interface{})
			schema.Properties[k] = &genai.Schema{
				Type:        genai.TypeString,
				Description: v["description"].(string),
			}
		}
		schema.Required = tool.InputSchema.Required
		// Creating schema
		// Warning, we return only one function for the POC
		return &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  schema,
			}},
		}
	}
	return nil
}

func (findmcplogs *MCPServerTool) Run(f genai.FunctionCall) (*genai.FunctionResponse, error) {
	request := mcp.CallToolRequest{}
	request.Params.Name = f.Name
	request.Params.Arguments = make(map[string]interface{})
	for k, v := range f.Args {
		request.Params.Arguments[k] = v.(string)
	}

	result, err := findmcplogs.client.CallTool(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to call service: %v", err)
	}
	var content string
	res := result.Content[0].(map[string]interface{})
	content = res["text"].(string)
	log.Println(content)
	return &genai.FunctionResponse{
		Name: f.Name,
		Response: map[string]any{
			"logs": content,
		},
	}, nil
}

func (findmcplogs *MCPServerTool) Name() string {
	return findmcplogs.name
}

func NewMCPServerTool() *MCPServerTool {
	return &MCPServerTool{
		name: "my_tooling",
	}
}
