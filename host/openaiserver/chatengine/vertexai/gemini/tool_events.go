package gemini

import (
	"encoding/json"
	"time"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

// ToolCallEvent represents a tool call being made by the model
type ToolCallEvent struct {
	ID        string          `json:"id"`
	Object    string          `json:"object"`
	Created   int64           `json:"created"`
	EventType string          `json:"event_type"`
	ToolCall  ToolCallDetails `json:"tool_call"`
}

type ToolCallDetails struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResponseEvent represents the response from a tool call
type ToolResponseEvent struct {
	ID           string              `json:"id"`
	Object       string              `json:"object"`
	Created      int64               `json:"created"`
	EventType    string              `json:"event_type"`
	ToolResponse ToolResponseDetails `json:"tool_response"`
}

type ToolResponseDetails struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Response interface{} `json:"response"`
	Error    string      `json:"error,omitempty"`
}

// NewToolCallEvent creates a new tool call event
func NewToolCallEvent(completionID string, toolName string, args map[string]interface{}) ToolCallEvent {
	return ToolCallEvent{
		ID:        completionID,
		Object:    "tool.call",
		Created:   time.Now().Unix(),
		EventType: "tool_call",
		ToolCall: ToolCallDetails{
			ID:        generateToolCallID(),
			Name:      toolName,
			Arguments: args,
		},
	}
}

// NewToolResponseEvent creates a new tool response event
func NewToolResponseEvent(completionID string, toolCallID string, toolName string, response interface{}, err error) ToolResponseEvent {
	event := ToolResponseEvent{
		ID:        completionID,
		Object:    "tool.response",
		Created:   time.Now().Unix(),
		EventType: "tool_response",
		ToolResponse: ToolResponseDetails{
			ID:       toolCallID,
			Name:     toolName,
			Response: response,
		},
	}

	if err != nil {
		event.ToolResponse.Error = err.Error()
	}

	return event
}

// generateToolCallID generates a unique ID for a tool call
func generateToolCallID() string {
	return "call_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ToJSON converts the event to JSON bytes
func (e ToolCallEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// ToJSON converts the event to JSON bytes
func (e ToolResponseEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Implement StreamEvent interface for ToolCallEvent
func (ToolCallEvent) IsStreamEvent() bool {
	return true
}

// Implement StreamEvent interface for ToolResponseEvent
func (ToolResponseEvent) IsStreamEvent() bool {
	return true
}

// Ensure our events implement the interface
var _ chatengine.StreamEvent = (*ToolCallEvent)(nil)
var _ chatengine.StreamEvent = (*ToolResponseEvent)(nil)
