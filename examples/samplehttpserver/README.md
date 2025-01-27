# Simple HTTP Server with Apache Combined Log Format

This is a simple HTTP server written in Go that logs incoming requests in the Apache combined log format. It's designed to generate log samples for use with LLM applications.

## Features

-   Logs requests in Apache combined log format to standard output.
-   Includes request details such as remote IP, user, timestamp, request line, status code, response size, referrer, and user agent.
-   Logs the request processing time.
-   Option to output logs to a file specified by the `LOG_FILE` environment variable.
-   Uses a custom `responseRecorder` to capture response status and size.
-   Implements a basic "Hello, World!" handler for testing purposes.

## Usage

1.  **Build the server**

    ```bash
    go build -o server main.go
    ```

2.  **Run the server**

    ```bash
    ./server
    ```

    The server will start on port 8080.

3.  **Optional: Specify a Log File**

    To write logs to a file, set the `LOG_FILE` environment variable before running the server.

    ```bash
    export LOG_FILE=access.log
    ./server
    ```

4.  **Send Test Requests**

    You can send requests to the server using `curl` or a browser to generate log entries.

    ```bash
    curl http://localhost:8080
    ```

## Log Format

The logs will be printed in the following Apache combined log format:

```
%remote_ip% - %user% [%timestamp%] "%request_line%" %status_code% %response_size% "%referrer%" "%user_agent%"
```

For example:

```
127.0.0.1 - - [07/Nov/2023:19:21:17 +0100] "GET / HTTP/1.1" 200 13 "-" "curl/7.79.1"
```


