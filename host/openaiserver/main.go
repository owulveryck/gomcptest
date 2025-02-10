package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/mark3labs/mcp-go/client"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

func main() {
	mcpServers := flag.String("mcpservers", "", "Input string of MCP servers")
	flag.Parse()

	openAIHandler := chatengine.NewOpenAIV1WithToolHandler(gcp.NewChatSession())
	servers := extractServers(*mcpServers)
	for i := range servers {
		var mcpClient client.MCPClient
		var err error
		if len(servers[i]) > 1 {
			mcpClient, err = client.NewStdioMCPClientLog(servers[i][0], nil, servers[i][1:]...)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			mcpClient, err = client.NewStdioMCPClientLog(servers[i][0], nil)
			if err != nil {
				log.Fatal(err)
			}

		}
		err = openAIHandler.AddTools(context.Background(), mcpClient)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Server starting on port 8080")
	// http.Handle("/", GzipMiddleware(openAIHandler)) // Wrap the handler with the gzip middleware
	http.Handle("/", openAIHandler) // Wrap the handler with the gzip middleware

	log.Fatal(http.ListenAndServe(":8080", nil)) // Use nil to use the default ServeMux
}
