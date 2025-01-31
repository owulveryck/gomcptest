package main

import (
	"fmt"
	"log"

	"cloud.google.com/go/vertexai/genai"
)

type callable interface {
	GetGenaiTool() *genai.Tool
	Run(genai.FunctionCall) (*genai.FunctionResponse, error)
	Name() string
}

func (cs *ChatSession) AddFunction(c callable) {
	if cs.model.Tools == nil {
		cs.functionsInventory = make(map[string]callable)
	}
	for _, fn := range c.GetGenaiTool().FunctionDeclarations {
		cs.functionsInventory[fn.Name] = c
	}
	cs.model.Tools = []*genai.Tool{c.GetGenaiTool()}
}

func (cs *ChatSession) CallFunction(f genai.FunctionCall) (*genai.FunctionResponse, error) {
	log.Println("Trying to call", f.Name)
	for k, v := range cs.functionsInventory {
		if k == f.Name {
			log.Printf("Running %v", f.Name)
			return v.Run(f)
		}
	}
	return nil, fmt.Errorf("no function matching the name found")
}
