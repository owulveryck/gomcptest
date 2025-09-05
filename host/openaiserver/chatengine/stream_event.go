package chatengine

// StreamEvent is an interface for all events that can be sent in a stream
type StreamEvent interface {
	// IsStreamEvent is a marker method to identify stream events
	IsStreamEvent() bool
}

// Ensure ChatCompletionStreamResponse implements StreamEvent
func (ChatCompletionStreamResponse) IsStreamEvent() bool {
	return true
}
