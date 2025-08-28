package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/genai"
)

func main() {
	// Read the test image
	imageBytes, err := ioutil.ReadFile("test_data/generative-ai_image_table.png")
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create Vertex AI client (matching updated main.go)
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  "bsjxygz-gcp-octo-lille",
		Location: "global",
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Note: genai.Client doesn't have Close method

	// Create parts for multimodal content (matching Go example exactly)
	imagePart := genai.NewPartFromBytes(imageBytes, "image/png")
	textPart := genai.NewPartFromText("Add beautiful flowers on the table")

	// Create content with image and text parts
	content := genai.NewContentFromParts([]*genai.Part{imagePart, textPart}, genai.RoleUser)

	// Configure generation parameters (matching Go example exactly)
	generateConfig := &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
	}

	fmt.Println("Making API call to generate content...")
	
	// Make the API call using direct method like Go example (not streaming)
	response, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash-preview-image-generation", []*genai.Content{content}, generateConfig)
	if err != nil {
		log.Fatalf("API call failed: %v", err)
	}

	fmt.Printf("Success! Generated %d candidates\n", len(response.Candidates))
	
	// Process response (similar to processEditResponse)
	for i, candidate := range response.Candidates {
		if candidate.Content == nil {
			continue
		}
		
		fmt.Printf("Candidate %d:\n", i+1)
		for j, part := range candidate.Content.Parts {
			fmt.Printf("  Part %d: Text='%s', InlineData=%v\n", j+1, part.Text, part.InlineData != nil)
			
			if part.Text != "" {
				fmt.Printf("  Text Response: %s\n", part.Text)
			}
			
			if part.InlineData != nil && len(part.InlineData.Data) > 0 {
				fmt.Printf("  Image found, size: %d bytes\n", len(part.InlineData.Data))
				// Save the image
				err := ioutil.WriteFile(fmt.Sprintf("output_test_%d_%d.png", i+1, j+1), part.InlineData.Data, 0644)
				if err != nil {
					log.Printf("Failed to save image: %v", err)
				} else {
					fmt.Printf("  Saved image to output_test_%d_%d.png\n", i+1, j+1)
				}
			}
		}
	}
}