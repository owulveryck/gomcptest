package gemini

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/genai"
)

func toGenaiPart(c *chatengine.ChatCompletionMessage) ([]*genai.Part, error) {
	if c.Content == nil {
		return nil, errors.New("no content")
	}

	switch v := c.Content.(type) {
	case string:
		return []*genai.Part{genai.NewPartFromText(v)}, nil
	case []interface{}:
		returnedParts := make([]*genai.Part, 0)
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				// Handle image_url content type
				if imgurl, ok := m["image_url"].(map[string]interface{}); ok {
					if url, ok := imgurl["url"].(string); ok {
						mime, data, err := chatengine.ExtractImageData(url)
						if err != nil {
							return nil, fmt.Errorf("failed to extract image  %w", err)
						}
						returnedParts = append(returnedParts, genai.NewPartFromBytes(data, mime))
					}
				}
				// Handle audio content type (inline audio data)
				if audio, ok := m["audio"].(map[string]interface{}); ok {
					if audioData, ok := audio["data"].(string); ok {
						mime, data, err := chatengine.ExtractFileData(audioData)
						if err != nil {
							return nil, fmt.Errorf("failed to extract audio data: %w", err)
						}
						slog.Debug("Processing audio", "mime_type", mime, "size", len(data))
						returnedParts = append(returnedParts, genai.NewPartFromBytes(data, mime))
					}
				}
				// Handle file content type (general file including audio)
				if file, ok := m["file"].(map[string]interface{}); ok {
					if fileData, ok := file["file_data"].(string); ok {
						mime, data, err := chatengine.ExtractFileData(fileData)
						if err != nil {
							return nil, fmt.Errorf("failed to extract file data: %w", err)
						}
						slog.Debug("Processing file", "mime_type", mime, "size", len(data))
						returnedParts = append(returnedParts, genai.NewPartFromBytes(data, mime))
					}
				}
				// Handle text content type (nested object)
				if textMap, ok := m["text"].(map[string]interface{}); ok {
					if value, ok := textMap["value"].(string); ok {
						returnedParts = append(returnedParts, genai.NewPartFromText(value))
					}
				}
				// Handle text content type (direct string)
				if text, ok := m["text"].(string); ok {
					returnedParts = append(returnedParts, genai.NewPartFromText(text))
				}
			}
		}
		return returnedParts, nil
	default:
		return []*genai.Part{}, nil
	}
}

func extractGenaiSchemaFromMCPProperty(p interface{}) (*genai.Schema, error) {
	switch p := p.(type) {
	case map[string]interface{}:
		return extractGenaiSchemaFromMCPPRopertyMap(p)
	default:
		return nil, fmt.Errorf("unhandled type for property %T (%v)", p, p)
	}
}

func extractGenaiSchemaFromMCPPRopertyMap(p map[string]interface{}) (*genai.Schema, error) {
	var propertyType, propertyDescription string
	var ok bool
	// first check if we have type and description
	if propertyType, ok = p["type"].(string); !ok {
		return nil, fmt.Errorf("expected type in the property details (%v)", p)
	}
	if propertyDescription, ok = p["description"].(string); !ok {
		slog.Debug("properties", "no description found", p)
	}
	switch propertyType {
	case "string":
		return &genai.Schema{
			Type:        genai.TypeString,
			Description: propertyDescription,
		}, nil
	case "number":
		return &genai.Schema{
			Type:        genai.TypeNumber,
			Description: propertyDescription,
		}, nil
	case "boolean":
		return &genai.Schema{
			Type:        genai.TypeBoolean,
			Description: propertyDescription,
		}, nil
	case "integer":
		return &genai.Schema{
			Type:        genai.TypeInteger,
			Description: propertyDescription,
		}, nil
	case "object":
		var properties map[string]interface{}
		var ok bool
		if properties, ok = p["properties"].(map[string]interface{}); !ok {
			return nil, fmt.Errorf("expected items in the property details for a type array (%v)", p)
		}
		var required []string
		if r, ok := p["required"].([]interface{}); ok {
			for _, r := range r {
				required = append(required, r.(string))
			}
		}
		genaiProperties := make(map[string]*genai.Schema, len(properties))
		for p, prop := range properties {
			var err error
			genaiProperties[p], err = extractGenaiSchemaFromMCPProperty(prop)
			if err != nil {
				return nil, err
			}
		}
		return &genai.Schema{
			Type:        genai.TypeObject,
			Description: propertyDescription,
			Properties:  genaiProperties,
			Required:    required,
		}, nil
	case "array":
		var items interface{}
		var ok bool
		if items, ok = p["items"]; !ok {
			return nil, fmt.Errorf("expected items in the property details for a type array (%v)", p)
		}
		schema, err := extractGenaiSchemaFromMCPProperty(items)
		if err != nil {
			return nil, err
		}
		return &genai.Schema{
			Type:        genai.TypeArray,
			Description: propertyDescription,
			Items:       schema,
		}, nil
	default:
		return nil, fmt.Errorf("unhandled type")
	}

	return nil, nil
}
