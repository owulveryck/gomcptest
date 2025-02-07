package gcp

import (
	"cloud.google.com/go/vertexai/genai"
)

type callable interface {
	GetGenaiTool() *genai.Tool
	Run(genai.FunctionCall) (*genai.FunctionResponse, error)
	Name() string
}
