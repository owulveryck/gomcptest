# Edit MCP Service

This service implements the `Edit` function from Claude Function Specifications as an MCP server.

## Description

This is a tool for editing files. For moving or renaming files, you should generally use the Bash tool with the 'mv' command instead. For larger edits, use the Write tool to overwrite files.

## Parameters

- `file_path` (string, required): The absolute path to the file to modify
- `old_string` (string, required): The text to replace
- `new_string` (string, required): The edited text to replace the old_string

## Usage Notes

- The old_string must uniquely identify the specific instance you want to change
- Include at least 3-5 lines of context before and after the change point
- This tool can only change one instance at a time
- Before using, check how many instances of the target text exist in the file
- For new files, use a new file path, empty old_string, and new file's contents as new_string

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```
{
  "name": "Edit",
  "params": {
    "file_path": "/absolute/path/to/file.txt",
    "old_string": "function oldFunction() {\n  console.log('old');\n}",
    "new_string": "function newFunction() {\n  console.log('new');\n}"
  }
}
```