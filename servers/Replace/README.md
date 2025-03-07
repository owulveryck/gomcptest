# Replace MCP Service

This service implements the `Replace` function from Claude Function Specifications as an MCP server.

## Description

Write a file to the local filesystem. Overwrites the existing file if there is one.

## Parameters

- `file_path` (string, required): The absolute path to the file to write (must be absolute, not relative)
- `content` (string, required): The content to write to the file

## Usage Notes

- Use the ReadFile tool to understand the file's contents and context before replacing
- For new files, verify the parent directory exists using the LS tool

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```
{
  "name": "Replace",
  "params": {
    "file_path": "/absolute/path/to/file.txt",
    "content": "This is the new content of the file.\nIt will completely replace any existing content."
  }
}
```