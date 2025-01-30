# MCP Log Seeker

This is a dummy implementation of an MCP server that allows you to extract log records from a specified log file within a given time range.

## Overview

The server provides a single tool `find_logs` that accepts the following parameters:

- `start_date`: The start date of the log extraction in the format `YYYY-MM-DD HH:mm:ss +ZZZZ` (e.g., `2025-01-24 12:00:00 +0100`).
- `end_date`: The end date of the log extraction in the format `YYYY-MM-DD HH:mm:ss +ZZZZ` (e.g., `2025-01-24 13:00:00 +0100`).
- `server_name`: The name of the server to get the logs from. The server accepts only `myserver` as a valid value.

The `find_logs` tool reads a specified log file line by line, parses the timestamp of a line using a regular expression against the log from the `access.log` file, and returns lines that fall within the provided start and end date.
The result is a string of all matched log lines concatenated together.

The server logs all actions into `/tmp/my_log.txt`.

## Getting Started

### Prerequisites
- Go >= 1.21
- `github.com/mark3labs/mcp-go`

### Usage
1. Run the server: `go run main.go -log <path-to-your-access.log>`

   - Replace `<path-to-your-access.log>` with the actual path to your access log file.
   - If the `-log` parameter is omitted the default value is `/Users/olivier.wulveryck/github.com/owulveryck/gomcptest/examples/samplehttpserver/access.log`

2. Interact with the server using an MCP client by calling the `find_logs` tool.

```bash
mcp call find_logs --server_name myserver --start_date "2024-11-26 08:00:00 +0000"  --end_date "2024-11-27 11:00:00 +0000"
```

### Example Log File
The server reads logs from the file path provided by the `-log` parameter. The default value is  `/Users/olivier.wulveryck/github.com/owulveryck/gomcptest/examples/samplehttpserver/access.log`

The log entries must follow the format:
```text
127.0.0.1 - - [26/Nov/2024:08:05:02 +0000] "GET / HTTP/1.1" 200 10 "-" "curl/8.1.2"
```

### Example Output

```text
127.0.0.1 - - [26/Nov/2024:08:05:02 +0000] "GET / HTTP/1.1" 200 10 "-" "curl/8.1.2"
127.0.0.1 - - [26/Nov/2024:09:15:13 +0000] "GET / HTTP/1.1" 200 10 "-" "curl/8.1.2"
```

## Notes
- The server only allows `myserver` as server name.
- Errors are logged to `/tmp/my_log.txt`.

TODO:

Change the name of the function
