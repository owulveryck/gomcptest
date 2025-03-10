# GlobTool MCP Service

This service implements the `GlobTool` function from Claude Function Specifications as an MCP server.

## Description

Fast file pattern matching tool that works with any codebase size. Supports glob patterns like "**/*.js" or "src/**/*.ts". Returns matching file paths sorted by modification time with detailed metadata.

## Parameters

- `pattern` (string, required): The glob pattern to match files against
- `path` (string, optional): The directory to search in. Defaults to the current working directory.
- `exclude` (string, optional): Glob pattern to exclude from the search results
- `limit` (number, optional): Maximum number of results to return
- `absolute` (boolean, optional): Return absolute paths instead of relative paths

## Features

- Improved glob pattern handling with true `**` support
- Concurrent file walking for better performance
- File metadata including size, modification time, and permissions
- Results can be limited to a specified number
- Exclude patterns to filter unwanted matches
- Option to display absolute or relative paths

## Usage Notes

- Use when you need to find files by name patterns
- For open-ended searches requiring multiple rounds of globbing and grepping, use the Agent tool instead

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```json
{
  "name": "GlobTool",
  "params": {
    "pattern": "**/*.go",
    "path": "/path/to/search",
    "exclude": "**/*_test.go",
    "limit": 50,
    "absolute": true
  }
}
```

## Dependencies

This tool uses the [doublestar](https://github.com/bmatcuk/doublestar) library for improved glob pattern matching.