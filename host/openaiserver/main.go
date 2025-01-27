package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/gomcptest/internal/vertexai"
)

type configuration struct {
	GCPPRoject      string `envconfig:"GCP_PROJECT" required:"true"`
	GeminiModel     string `envconfig:"GEMINI_MODEL" default:"gemini-2.0-pro"`
	GCPRegion       string `envconfig:"GCP_REGION" default:"us-central1"`
	Port            string `envconfig:"ANALYSE_PDF_PORT" default:"50051"`
	MCPServerSample string `envconfig:"MCP_SERVER" default:"/Users/olivier.wulveryck/github.com/owulveryck/gomcptest/servers/logs/logs"`
	MCPServerArgs   string `envconfig:"MCP_SERVER_ARGS" default:"-log /tmp/access.log"`
}

var (
	config         configuration
	vertexAIClient *vertexai.AI
)

func main() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	vertexAIClient = vertexai.NewAI(ctx, config.GCPPRoject, config.GCPRegion, config.GeminiModel)

	cs := NewChatSession()
	//	cs.AddFunction(NewFindTheaters())
	// cs.AddFunction(NewFindLogs())
	cs.AddFunction(NewMCPFindLogs())

	mux := http.NewServeMux()
	mux.Handle("/v1/chat/completions", http.HandlerFunc(cs.chatCompletionHandler))
	mux.Handle("/v1/models", http.HandlerFunc(modelsHandler))
	mux.Handle("/v1/models/", http.HandlerFunc(modelDetailHandler))

	fmt.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
	_ = vertexAIClient
}
