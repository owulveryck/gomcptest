# LS MCP Service

This service implements the `LS` function from Claude Function Specifications as an MCP server.

## Description

Lists files and directories in a given path. The path parameter must be an absolute path, not a relative path.

## Parameters

- `path` (string, required): The absolute path to the directory to list (must be absolute, not relative)
- `ignore_pattern` (string, optional): Glob pattern to ignore

## Usage Notes

- You should generally prefer the Glob and Grep tools, if you know which directories to search

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```
{
  "name": "LS",
  "params": {
    "path": "/absolute/path/to/dir",
    "ignore_pattern": "*.tmp"
  }
}
```