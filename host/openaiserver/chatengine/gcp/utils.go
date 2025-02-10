package gcp

import (
	"log"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

func toGenaiPart(c *chatengine.ChatCompletionMessage) []genai.Part {
	if c.Content == nil {
		return nil
	}

	switch v := c.Content.(type) {
	case string:
		return []genai.Part{genai.Text(v)}
	case []interface{}:
		returnedParts := make([]genai.Part, 0)
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if imgurl, ok := m["image_url"].(map[string]interface{}); ok {
					if value, ok := imgurl["url"].(string); ok {
						mime, data, err := chatengine.ExtractImageData(value)
						if err != nil {
							log.Fatal(err)
						}
						returnedParts = append(returnedParts, genai.Blob{
							Data:     data,
							MIMEType: mime,
						})
					}
				}
				if text, ok := m["text"].(map[string]interface{}); ok {
					if value, ok := text["value"].(string); ok {
						returnedParts = append(returnedParts, genai.Text(value))
					}
				}
				if text, ok := m["text"].(string); ok {
					returnedParts = append(returnedParts, genai.Text(text))
				}
			}
		}
		return returnedParts
	default:
		return nil
	}
}
