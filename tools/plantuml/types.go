package main

// PlantUMLRequest represents a request to render PlantUML
type PlantUMLRequest struct {
	PlantUMLCode string `json:"plantuml_code"`
	OutputFormat string `json:"output_format"`
}

// PlantUMLResponse represents a response from PlantUML rendering
type PlantUMLResponse struct {
	Format  string `json:"format"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}
