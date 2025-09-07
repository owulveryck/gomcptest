package chatengine

import (
	"bytes"
	"html/template"
	"log/slog"
	"time"
)

// TemplateMiddleware provides template processing for system messages
type TemplateMiddleware struct {
	funcMap template.FuncMap
}

// NewTemplateMiddleware creates a new template middleware with default functions
func NewTemplateMiddleware() *TemplateMiddleware {
	funcMap := template.FuncMap{
		"now":                  time.Now,
		"loadLocation":         time.LoadLocation,
		"formatTimeInLocation": formatTimeInLocation,
	}

	return &TemplateMiddleware{
		funcMap: funcMap,
	}
}

// formatTimeInLocation takes a timezone name, format string, and time (from pipeline), and returns formatted time in that timezone
// Note: In template pipelines, the time.Time is passed as the last argument from the pipeline
func formatTimeInLocation(locationName, format string, t time.Time) string {
	location, err := time.LoadLocation(locationName)
	if err != nil {
		// Fallback to UTC if timezone loading fails
		return t.UTC().Format(format)
	}
	return t.In(location).Format(format)
}

// ProcessRequest processes the chat completion request and applies templates to system messages
func (tm *TemplateMiddleware) ProcessRequest(request *ChatCompletionRequest) error {
	for i := range request.Messages {
		if request.Messages[i].Role == "system" {
			processed, err := tm.processSystemMessage(&request.Messages[i])
			if err != nil {
				slog.Error("Failed to process system message template",
					slog.String("error", err.Error()),
					slog.Int("message_index", i))
				return err
			}
			request.Messages[i] = *processed
		}
	}
	return nil
}

// processSystemMessage processes a single system message through the template engine
func (tm *TemplateMiddleware) processSystemMessage(message *ChatCompletionMessage) (*ChatCompletionMessage, error) {
	// Get the string content from the message
	content := message.GetContent()
	if content == "" {
		// No content to process, return as-is
		return message, nil
	}

	// Create and parse the template
	tmpl, err := template.New("systemPrompt").Funcs(tm.funcMap).Parse(content)
	if err != nil {
		return nil, err
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, err
	}

	// Create a new message with the processed content
	processedMessage := *message
	processedMessage.Content = buf.String()

	slog.Debug("Processed system message template",
		slog.String("original", content),
		slog.String("processed", buf.String()))

	return &processedMessage, nil
}

// AddTemplateFunc adds a custom function to the template function map
func (tm *TemplateMiddleware) AddTemplateFunc(name string, fn interface{}) {
	tm.funcMap[name] = fn
}
