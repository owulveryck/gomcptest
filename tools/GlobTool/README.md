# GlobTool MCP Service

This service implements the `GlobTool` function from Claude Function Specifications as an MCP server.

## Description

Fast file pattern matching tool that works with any codebase size. Supports glob patterns like "**/*.js" or "src/**/*.ts". Returns matching file paths sorted by modification time.

## Parameters

- `pattern` (string, required): The glob pattern to match files against
- `path` (string, optional): The directory to search in. Defaults to the current working directory.

## Usage Notes

- Use when you need to find files by name patterns
- For open-ended searches requiring multiple rounds of globbing and grepping, use the Agent tool instead

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```
{
  "name": "GlobTool",
  "params": {
    "pattern": "**/*.go",
    "path": "/path/to/search"
  }
}
```