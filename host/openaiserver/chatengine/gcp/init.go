package gcp

import (
	"context"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/gomcptest/internal/vertexai"
)

type configuration struct {
	GCPPRoject  string `envconfig:"GCP_PROJECT" required:"true"`
	GeminiModel string `envconfig:"GEMINI_MODEL" default:"gemini-2.0-pro"`
	GCPRegion   string `envconfig:"GCP_REGION" default:"us-central1"`
}

var (
	config         configuration
	vertexAIClient *vertexai.AI
)

func init() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	vertexAIClient = vertexai.NewAI(ctx, config.GCPPRoject, config.GCPRegion, config.GeminiModel)
}
