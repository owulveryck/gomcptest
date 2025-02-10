package gcp

import (
	"sync"

	"cloud.google.com/go/vertexai/genai"
)

type functionCallStack struct {
	mu    sync.Mutex
	items []genai.FunctionCall
}

// newFunctionCallStack creates and returns a new stack
func newFunctionCallStack() *functionCallStack {
	return &functionCallStack{
		items: make([]genai.FunctionCall, 0),
	}
}

// push adds a genai.FunctionCall to the top of the stack.
func (s *functionCallStack) push(call genai.FunctionCall) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, call)
}

// pop removes and returns the last genai.FunctionCall from the stack (FIFO).
// Returns nil if the stack is empty
func (s *functionCallStack) pop() *genai.FunctionCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) == 0 {
		return nil
	}
	// Get the first element
	call := s.items[0]
	// Remove the first element from the slice
	s.items = s.items[1:]
	return &call
}

// peek returns the last genai.FunctionCall from the stack without removing it (FIFO).
// Returns nil if the stack is empty
func (s *functionCallStack) peek() *genai.FunctionCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) == 0 {
		return nil
	}
	return &s.items[0]
}

// size return the number of items in the stack
func (s *functionCallStack) size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}
