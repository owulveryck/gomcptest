# Bash MCP Service

This service implements the `Bash` function from Claude Function Specifications as an MCP server.

## Description

Executes a given bash command in a persistent shell session with optional timeout, ensuring proper handling and security measures.

## Parameters

- `command` (string, required): The command to execute
- `timeout` (number, optional): Optional timeout in milliseconds (max 600000)

## Usage Notes

- Verify directory exists before creating files/directories
- Some commands are limited or banned for security reasons
- All commands share the same shell session (environment persists)
- Commands will timeout after 30 minutes if no timeout specified
- Output truncated if exceeding 30000 characters
- Avoid using search commands like `find` and `grep`
- Avoid read tools like `cat`, `head`, `tail`, and `ls`

## Banned Commands

For security reasons, the following commands are banned:
alias, curl, curlie, wget, axel, aria2c, nc, telnet, lynx, w3m, links, httpie, xh, http-prompt, chrome, firefox, safari

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```
{
  "name": "Bash",
  "params": {
    "command": "ls -la",
    "timeout": 30000
  }
}
```