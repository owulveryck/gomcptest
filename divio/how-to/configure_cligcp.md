# How to Configure the cliGCP Command Line Interface

This guide shows you how to configure and customize the cliGCP command line interface for various use cases.

## Prerequisites

- A working installation of gomcptest
- Basic familiarity with the cliGCP tool from the tutorial
- Understanding of environment variables and configuration

## Command Line Arguments

The cliGCP tool accepts the following command line arguments:

```bash
# Specify the MCP servers to use (required)
-mcpservers "tool1;tool2;tool3"

# Example with tool arguments
./cliGCP -mcpservers "./GlobTool;./GrepTool;./dispatch_agent -glob-path ./GlobTool -grep-path ./GrepTool -ls-path ./LS -view-path ./View;./Bash"
```

## Environment Variables Configuration

### GCP Configuration

Configure the Google Cloud Platform integration with these environment variables:

```bash
# GCP Project ID (required)
export GCP_PROJECT=your-gcp-project-id

# GCP Region (default: us-central1)
export GCP_REGION=us-central1

# Comma-separated list of Gemini models (required)
export GEMINI_MODELS=gemini-1.5-pro,gemini-2.0-flash

# Directory to store images (required for image generation)
export IMAGE_DIR=/path/to/image/directory
```

### Advanced Configuration

You can customize the behavior of the cliGCP tool with these additional environment variables:

```bash
# Set a custom system instruction for the model
export SYSTEM_INSTRUCTION="You are a helpful assistant specialized in Go programming."

# Adjust the model's temperature (0.0-1.0, default is 0.2)
# Lower values make output more deterministic, higher values more creative
export MODEL_TEMPERATURE=0.3

# Set a maximum token limit for responses
export MAX_OUTPUT_TOKENS=2048
```

## Creating Shell Aliases

To simplify usage, create shell aliases in your `.bashrc` or `.zshrc`:

```bash
# Add to ~/.bashrc or ~/.zshrc
alias gpt='cd /path/to/gomcptest/bin && ./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash;./Replace"'

# Create specialized aliases for different tasks
alias code-assistant='cd /path/to/gomcptest/bin && GCP_PROJECT=your-project GEMINI_MODELS=gemini-2.0-flash ./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash;./Replace"'

alias security-scanner='cd /path/to/gomcptest/bin && SYSTEM_INSTRUCTION="You are a security expert focused on finding vulnerabilities in code" ./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash"'
```

## Customizing the System Instruction

To modify the default system instruction, edit the `agent.go` file:

```go
// In host/cliGCP/cmd/agent.go
genaimodels[model].SystemInstruction = &genai.Content{
    Role: "user",
    Parts: []genai.Part{
        genai.Text("You are a helpful agent with access to tools. " +
            "Your job is to help the user by performing tasks using these tools. " +
            "You should not make up information. " +
            "If you don't know something, say so and explain what you would need to know to help. " +
            "If not indication, use the current working directory which is " + cwd),
    },
}
```

## Creating Task-Specific Configurations

For different use cases, you can create specialized configuration scripts:

### Code Review Helper

Create a file called `code-reviewer.sh`:

```bash
#!/bin/bash

export GCP_PROJECT=your-gcp-project-id
export GCP_REGION=us-central1
export GEMINI_MODELS=gemini-2.0-flash
export IMAGE_DIR=/tmp/images
export SYSTEM_INSTRUCTION="You are a code review expert. Analyze code for bugs, security issues, and areas for improvement. Focus on providing constructive feedback and detailed explanations."

cd /path/to/gomcptest/bin
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash"
```

Make it executable:

```bash
chmod +x code-reviewer.sh
```

### Documentation Generator

Create a file called `doc-generator.sh`:

```bash
#!/bin/bash

export GCP_PROJECT=your-gcp-project-id
export GCP_REGION=us-central1
export GEMINI_MODELS=gemini-2.0-flash
export IMAGE_DIR=/tmp/images
export SYSTEM_INSTRUCTION="You are a documentation specialist. Your task is to help create clear, comprehensive documentation for code. Analyze code structure and create appropriate documentation following best practices."

cd /path/to/gomcptest/bin
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash;./Replace"
```

## Advanced Tool Configurations

### Configuring dispatch_agent

When using the dispatch_agent tool, you can configure its behavior with additional arguments:

```bash
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./dispatch_agent -glob-path ./GlobTool -grep-path ./GrepTool -ls-path ./LS -view-path ./View -timeout 30s;./Bash;./Replace"
```

### Creating Tool Combinations

You can create specialized tool combinations for different tasks:

```bash
# Web development toolset
./cliGCP -mcpservers "./GlobTool -include '*.{html,css,js}';./GrepTool;./LS;./View;./Bash;./Replace"

# Go development toolset
./cliGCP -mcpservers "./GlobTool -include '*.go';./GrepTool;./LS;./View;./Bash;./Replace"
```

## Troubleshooting Common Issues

### Model Connection Issues

If you're having trouble connecting to the Gemini model:

1. Verify your GCP credentials:
```bash
gcloud auth application-default print-access-token
```

2. Check that the Vertex AI API is enabled:
```bash
gcloud services list --enabled | grep aiplatform
```

3. Verify your project has access to the models you're requesting

### Tool Execution Failures

If tools are failing to execute:

1. Ensure the tool paths are correct
2. Verify the tools are executable
3. Check for permission issues in the directories you're accessing

### Performance Optimization

For better performance:

1. Use more specific tool patterns to reduce search scope
2. Consider creating specialized agents for different tasks
3. Set a lower temperature for more deterministic responses