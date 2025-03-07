# GrepTool MCP Service

This service implements the `GrepTool` function from Claude Function Specifications as an MCP server.

## Description

Fast content search tool that works with any codebase size. Searches file contents using regular expressions. Supports full regex syntax (eg. "log.*Error", "function\\s+\\w+", etc.). Returns matching file paths sorted by modification time.

## Parameters

- `pattern` (string, required): The regular expression pattern to search for in file contents
- `include` (string, optional): File pattern to include in the search (e.g. "*.js", "*.{ts,tsx}")
- `path` (string, optional): The directory to search in. Defaults to the current working directory.

## Usage Notes

- Use when you need to find files containing specific patterns
- For open-ended searches requiring multiple rounds of globbing and grepping, use the Agent tool instead

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```
{
  "name": "GrepTool",
  "params": {
    "pattern": "func.*Main",
    "include": "*.go",
    "path": "/path/to/search"
  }
}
```