package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/vertexai/genai"

	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/gomcptest/internal/vertexai"
)

type configuration struct {
	GCPPRoject  string `envconfig:"GCP_PROJECT" required:"true"`
	GeminiModel string `envconfig:"GEMINI_MODEL" default:"gemini-2.0-pro"`
	GCPRegion   string `envconfig:"GCP_REGION" default:"us-central1"`
	Port        string `envconfig:"ANALYSE_PDF_PORT" default:"50051"`
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

	client := vertexAIClient.Client
	defer client.Close()

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

	movieTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "find_theaters",
			Description: "find theaters based on location and optionally movie title which is currently playing in theaters",
			Parameters:  schema,
		}},
	}

	model := client.GenerativeModel(config.GeminiModel)

	// Before initiating a conversation, we tell the model which tools it has
	// at its disposal.
	model.Tools = []*genai.Tool{movieTool}

	// For using tools, the chat mode is useful because it provides the required
	// chat context. A model needs to have tools supplied to it in the chat
	// history so it can use them in subsequent conversations.
	//
	// The flow of message expected here is:
	//
	// 1. We send a question to the model
	// 2. The model recognizes that it needs to use a tool to answer the question,
	//    an returns a FunctionCall response asking to use the tool.
	// 3. We send a FunctionResponse message, simulating the return value of
	//    the tool for the model's query.
	// 4. The model provides its text answer in response to this message.
	session := model.StartChat()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// res, err := session.SendMessage(ctx, genai.Text("Which theaters in Mountain View show Barbie movie?"))
		res, err := session.SendMessage(ctx, genai.Text(line))
		if err != nil {
			log.Fatalf("session.SendMessage: %v", err)
		}

		part := res.Candidates[0].Content.Parts[0]
		switch v := part.(type) {
		case genai.FunctionCall:
			funcall := v

			// Expect the model to pass a proper string "location" argument to the tool.
			if _, ok := funcall.Args["location"].(string); !ok {
				log.Fatalf("expected string: %v", funcall.Args["location"])
			}

			// Provide the model with a hard-coded reply.
			res, err = session.SendMessage(ctx, genai.FunctionResponse{
				Name: movieTool.FunctionDeclarations[0].Name,
				Response: map[string]any{
					"theater": "AMC16",
				},
			})
			if err != nil {
				log.Fatal(err)
			}
			printResponse(res)
		default:
			printResponse(res)

		}
	}
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
	fmt.Println("---")
}
