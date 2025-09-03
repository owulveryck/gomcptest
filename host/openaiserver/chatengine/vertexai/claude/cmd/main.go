package main

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/vertex"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/vertexai"
)

func main() {
	config, _ := vertexai.LoadConfig()
	client := anthropic.NewClient(
		vertex.WithGoogleAuth(context.Background(), config.GCPRegion, config.GCPProject),
	)
	content := "What is a quaternion?"
	stream := client.Messages.NewStreaming(context.TODO(), anthropic.MessageNewParams{
		Model:     "claude-opus-4-1@20250805",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(content)),
		},
	})

	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		err := message.Accumulate(event)
		if err != nil {
			panic(err)
		}

		switch eventVariant := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch deltaVariant := eventVariant.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				print(deltaVariant.Text)
			}
		}
	}

	if stream.Err() != nil {
		panic(stream.Err())
	}
}
