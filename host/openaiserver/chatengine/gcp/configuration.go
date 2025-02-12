package gcp

type Configuration struct {
	GCPProject   string   `envconfig:"GCP_PROJECT" required:"true"`
	GeminiModels []string `envconfig:"GEMINI_MODELS" default:"gemini-1.5-pro,gemini-2.0-flash"`
	GCPRegion    string   `envconfig:"GCP_REGION" default:"us-central1"`
}
