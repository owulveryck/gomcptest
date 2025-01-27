package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/kr/pretty"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type FindMCPLogs struct {
	name   string
	client *client.StdioMCPClient
}

func (findmcplogs *FindMCPLogs) GetGenaiTool() *genai.Tool {
	ctx := context.Background()
	var err error
	findmcplogs.client, err = client.NewStdioMCPClient(config.MCPServerSample, strings.Split(config.MCPServerArgs, " "))
	if err != nil {
		log.Fatal(err)
	}
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "example-client",
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
		/*
				* mcp.ToolInputSchema{
			    Type:       "object",
			    Properties: {
			        "end_date": map[string]interface {}{
			            "description": "The end date of the log extraction in the format 2025-01-24 12:00:00 +0100",
			            "type":        "string",
			        },
			        "server_name": map[string]interface {}{
			            "description": "The name of the server to get the logs from",
			            "type":        "string",
			        },
			        "start_date": map[string]interface {}{
			            "description": "The start date of the log extraction in the format 2025-01-24 12:00:00 +0100",
			            "type":        "string",
			        },
			    },
			    Required: {"start_date", "end_date", "server_name"},
			}*/
		// Creating schema
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

func (findmcplogs *FindMCPLogs) Run(f genai.FunctionCall) (*genai.FunctionResponse, error) {
	getLogRequest := mcp.CallToolRequest{}
	serverName := f.Args["server_name"].(string)
	startDate := f.Args["start_date"].(string)
	endDate := f.Args["end_date"].(string)
	getLogRequest.Params.Name = f.Name
	getLogRequest.Params.Arguments = map[string]interface{}{
		"server_name": serverName,
		"start_date":  startDate,
		"end_date":    endDate,
	}
	pretty.Print(getLogRequest)

	result, err := findmcplogs.client.CallTool(context.Background(), getLogRequest)
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

func (findmcplogs *FindMCPLogs) Name() string {
	return findmcplogs.name
}

func NewMCPFindLogs() *FindMCPLogs {
	return &FindMCPLogs{
		name: "find_logs",
	}
}
