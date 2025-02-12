package gcp

import (
	"errors"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

func toGenaiPart(c *chatengine.ChatCompletionMessage) ([]genai.Part, error) {
	if c.Content == nil {
		return nil, errors.New("no content")
	}

	switch v := c.Content.(type) {
	case string:
		return []genai.Part{genai.Text(v)}, nil
	case []interface{}:
		returnedParts := make([]genai.Part, 0)
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if imgurl, ok := m["image_url"].(map[string]interface{}); ok {
					if url, ok := imgurl["url"].(string); ok {
						mime, data, err := chatengine.ExtractImageData(url)
						if err != nil {
							return nil, fmt.Errorf("failed to extract image  %w", err)
						}
						returnedParts = append(returnedParts, genai.Blob{
							Data:     data,
							MIMEType: mime,
						})
					}
				}
				if textMap, ok := m["text"].(map[string]interface{}); ok {
					if value, ok := textMap["value"].(string); ok {
						returnedParts = append(returnedParts, genai.Text(value))
					}
				}
				if text, ok := m["text"].(string); ok {
					returnedParts = append(returnedParts, genai.Text(text))
				}
			}
		}
		return returnedParts, nil
	default:
		return []genai.Part{}, nil
	}
}
