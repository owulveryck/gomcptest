package gcp

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	GCPProject   string   `envconfig:"GCP_PROJECT" required:"true"`
	GeminiModels []string `envconfig:"GEMINI_MODELS" default:"gemini-1.5-pro,gemini-2.0-flash"`
	GCPRegion    string   `envconfig:"GCP_REGION" default:"us-central1"`
	ImagenModels []string `envconfig:"IMAGEN_MODELS"`
	ImageDir     string   `envconfig:"IMAGE_DIR" required:"true"`
	Port         string   `envconfig:"PORT" default:"8080"`
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (Configuration, error) {
	var cfg Configuration
	err := envconfig.Process("", &cfg)
	if err != nil {
		return Configuration{}, fmt.Errorf("error processing configuration: %v", err)
	}
	return cfg, nil
}
