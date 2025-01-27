package vertexai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/vertexai/genai"
)

type AI struct {
	Client   *genai.Client
	Gemini   *genai.GenerativeModel
	Token    string
	location string
	project  string
}

// GenerateImageFromText generates an image from a given text prompt.
// The prompt parameter provides the description of the image to generate.
// It returns the generated image as an image.Image object, and any error encountered.
func (ai *AI) GenerateImageFromText(ctx context.Context, prompt string) (image.Image, error) {
	// Define the API URL
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		ai.location, ai.project, ai.location, "imagen-3.0-generate-001")

	// Define the request payload
	type Instance struct {
		Prompt string `json:"prompt"`
	}
	type Parameters struct {
		SampleCount int `json:"sampleCount"`
	}
	type Payload struct {
		Instances  []Instance `json:"instances"`
		Parameters Parameters `json:"parameters"`
	}

	type Prediction struct {
		MimeType           string `json:"mimeType"`
		BytesBase64Encoded string `json:"bytesBase64Encoded"`
	}

	type ResponsePayload struct {
		Predictions []Prediction `json:"predictions"`
	}
	payload := Payload{
		Instances: []Instance{
			{Prompt: prompt},
		},
		Parameters: Parameters{
			SampleCount: 1,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+ai.Token)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and print the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Parse the JSON payload
	var respPayload ResponsePayload
	err = json.Unmarshal(respBody, &respPayload)
	if err != nil {
		return nil, err
	}

	// Check if predictions exist
	if len(respPayload.Predictions) == 0 {
		return nil, err
	}

	// Decode Base64 string
	encodedBytes := respPayload.Predictions[0].BytesBase64Encoded
	imageBytes, err := base64.StdEncoding.DecodeString(encodedBytes)
	if err != nil {
		return nil, err
	}

	// Decode the image
	img, err := png.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// SimpleQuery to the AI model
func (ai *AI) SimpleQuery(ctx context.Context, prompt string) (string, error) {
	resp, err := ai.Gemini.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}
	result := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if part, ok := part.(genai.Text); ok {
			result += string(part)
		}
	}

	return result, nil
}

// SimpleQuery to the AI model
func (ai *AI) SimpleQueryPng(ctx context.Context, prompt string, image []byte) (string, error) {
	resp, err := ai.Gemini.GenerateContent(ctx, genai.ImageData("png", image), genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}
	result := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if part, ok := part.(genai.Text); ok {
			result += string(part)
		}
	}

	return result, nil
}

func NewAI(ctx context.Context, projectID, locationID, modelID string) *AI {
	client, err := genai.NewClient(ctx, projectID, locationID)
	if err != nil {
		log.Fatalf("Failed to create the client: %v", err)
	}
	tok, err := getAccessToken()
	if err != nil {
		log.Fatalf("Cannot get token: %v", err)
	}
	return &AI{
		location: locationID,
		project:  projectID,
		Client:   client,
		Gemini:   client.GenerativeModel(modelID),
		Token:    tok,
	}
}
