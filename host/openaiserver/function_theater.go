package main

import "cloud.google.com/go/vertexai/genai"

type FindTheaters struct {
	name string
}

func NewFindTheaters() *FindTheaters {
	return &FindTheaters{
		name: "find_theaters",
	}
}

func (findtheaters *FindTheaters) GetGenaiTool() *genai.Tool {
	// To use functions / tools, we have to first define a schema that describes
	// the function to the model. The schema is similar to OpenAPI 3.0.
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"location": {
				Type:        genai.TypeString,
				Description: "The city and state, e.g. San Francisco, CA or a zip code e.g. 95616",
			},
			"title": {
				Type:        genai.TypeString,
				Description: "Any movie title",
			},
		},
		Required: []string{"location"},
	}

	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        findtheaters.name,
			Description: "find theaters based on location and optionally movie title which is currently playing in theaters",
			Parameters:  schema,
		}},
	}
}

func (findtheaters *FindTheaters) Name() string {
	return findtheaters.name
}

func (findtheaters *FindTheaters) Run(f genai.FunctionCall) (*genai.FunctionResponse, error) {
	return &genai.FunctionResponse{
		Name: findtheaters.name,
		Response: map[string]any{
			"theater": "AMC16",
		},
	}, nil
}
