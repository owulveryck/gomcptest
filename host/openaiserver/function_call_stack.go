package main

import (
	"sync"

	"cloud.google.com/go/vertexai/genai"
)

type FunctionCallStack struct {
	mu    sync.Mutex
	items []genai.FunctionCall
}

// NewFunctionCallStack creates and returns a new stack
func NewFunctionCallStack() *FunctionCallStack {
	return &FunctionCallStack{
		items: make([]genai.FunctionCall, 0),
	}
}

// Push adds a genai.FunctionCall to the top of the stack.
func (s *FunctionCallStack) Push(call genai.FunctionCall) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, call)
}

// Pop removes and returns the last genai.FunctionCall from the stack (FIFO).
// Returns nil if the stack is empty
func (s *FunctionCallStack) Pop() *genai.FunctionCall {
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

// Peek returns the last genai.FunctionCall from the stack without removing it (FIFO).
// Returns nil if the stack is empty
func (s *FunctionCallStack) Peek() *genai.FunctionCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) == 0 {
		return nil
	}
	return &s.items[0]
}

// Size return the number of items in the stack
func (s *FunctionCallStack) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}
