# View MCP Service

This service implements the `View` function from Claude Function Specifications as an MCP server.

## Description

Reads a file from the local filesystem. The file_path parameter must be an absolute path, not a relative path.

## Parameters

- `file_path` (string, required): The absolute path to the file to read
- `offset` (number, optional): The line number to start reading from
- `limit` (number, optional): The number of lines to read

## Usage Notes

- By default, reads up to 2000 lines from the beginning of the file
- Any lines longer than 2000 characters will be truncated
- For image files, the tool will display the image as a base64 encoded string
- For Jupyter notebooks (.ipynb files), use the ReadNotebook instead

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```
{
  "name": "View",
  "params": {
    "file_path": "/absolute/path/to/file.txt",
    "offset": 10,
    "limit": 100
  }
}
```