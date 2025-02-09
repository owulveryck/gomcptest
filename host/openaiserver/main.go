package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/client"

	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

type configuration struct {
	MCPServer     string `envconfig:"MCP_SERVER" default:"/Users/olivier.wulveryck/github.com/owulveryck/gomcptest/servers/logs/logs"`
	MCPServerArgs string `envconfig:"MCP_SERVER_ARGS" default:"-log /tmp/access.log"`
}

func main() {
	var config configuration
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}
	openAIHandler := chatengine.NewOpenAIV1WithToolHandler(gcp.NewChatSession())
	if config.MCPServer != "" {
		mcpClient, err := client.NewStdioMCPClientLog(config.MCPServer, strings.Split(config.MCPServerArgs, " "))
		if err != nil {
			log.Fatal(err)
		}
		err = openAIHandler.AddTools(context.Background(), mcpClient)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Server starting on port 8080")
	http.Handle("/", GzipMiddleware(openAIHandler)) // Wrap the handler with the gzip middleware

	log.Fatal(http.ListenAndServe(":8080", nil)) // Use nil to use the default ServeMux
}
