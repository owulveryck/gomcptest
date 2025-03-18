---
title: "How to Create a Custom MCP Tool"
linkTitle: "Create Custom Tool"
weight: 1
description: >-
  Build your own Model Context Protocol (MCP) compatible tools
---

This guide shows you how to create a new custom tool that's compatible with the Model Context Protocol (MCP) in gomcptest.

## Prerequisites

- A working installation of gomcptest
- Go programming knowledge
- Understanding of the MCP protocol basics

## Steps to create a custom tool

### 1. Create the tool directory structure

```bash
mkdir -p tools/YourToolName/cmd
```

### 2. Create the README.md file

Create a `README.md` in the tool directory with documentation:

```bash
touch tools/YourToolName/README.md
```

Include the following sections:
- Tool description
- Parameters
- Usage notes
- Example

### 3. Create the main.go file

Create a `main.go` file in the cmd directory:

```bash
touch tools/YourToolName/cmd/main.go
```

### 4. Implement the tool functionality

Here's a template to start with:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go"
)

// Define your tool's parameters structure
type Params struct {
	// Add your parameters here
	// Example:
	InputParam string `json:"input_param"`
}

func main() {
	server := mcp.NewServer()

	// Register your tool function
	server.RegisterFunction("YourToolName", func(params json.RawMessage) (any, error) {
		var p Params
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("failed to parse parameters: %w", err)
		}

		// Implement your tool's logic here
		result := doSomethingWithParams(p)

		return result, nil
	})

	if err := server.Run(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func doSomethingWithParams(p Params) interface{} {
	// Your tool's core functionality
	// ...
	
	// Return the result
	return map[string]interface{}{
		"result": "Your processed result",
	}
}
```

### 5. Add the tool to the Makefile

Open the Makefile in the root directory and add your tool:

```makefile
YourToolName:
	go build -o bin/YourToolName tools/YourToolName/cmd/main.go
```

Also add it to the `all` target.

### 6. Build your tool

```bash
make YourToolName
```

### 7. Test your tool

Test the tool directly:

```bash
echo '{"name":"YourToolName","params":{"input_param":"test"}}' | ./bin/YourToolName
```

### 8. Use with the CLI

Add your tool to the CLI command:

```bash
./bin/cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./YourToolName;./dispatch_agent;./Bash;./Replace"
```

## Tips for effective tool development

- Focus on a single, well-defined purpose
- Provide clear error messages
- Include meaningful response formatting
- Implement proper parameter validation
- Handle edge cases gracefully
- Consider adding unit tests in a _test.go file