package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	logFilePath   string
	accessLogPath string
)

func main() {
	flag.StringVar(&accessLogPath, "log", "/Users/olivier.wulveryck/github.com/owulveryck/gomcptest/examples/samplehttpserver/access.log", "Path to the access log file")

	flag.Parse()
	logFilePath = "/tmp/my_log.txt"

	err := logToFile(logFilePath, "Starting the application")
	if err != nil {
		log.Fatalf("Error logging: %v", err)
	}

	err = logToFile(logFilePath, "Another log entry")
	if err != nil {
		log.Fatalf("Error logging: %v", err)
	}

	fmt.Println("Logs written to:", logFilePath)
	//
	// Create MCP server
	s := server.NewMCPServer(
		"DuckDB 🚀",
		"1.0.0",
	)

	// Add tool
	tool := mcp.NewTool("query_file",
		mcp.WithDescription("runs a SQL query through duckdb to extrat the information of a file"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to execute (compatible with DUCKDB)"),
		),
	)

	// Add tool handler
	s.AddTool(tool, duckDBHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func duckDBHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	queryStr, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return mcp.NewToolResultError("query must be a string"), nil
	}
	logToFile(logFilePath, "running"+queryStr)
	res, err := executeDuckDBQuery(queryStr)
	if err != nil {
		logToFile(logFilePath, err.Error())
		return mcp.NewToolResultError("query_string encountered an error: " + err.Error()), nil
	}
	logToFile(logFilePath, "result"+res)

	return mcp.NewToolResultText(res), nil
}

func executeDuckDBQuery(queryStr string) (string, error) {
	cmd := exec.Command("duckdb")

	cmd.Stdin = bytes.NewBufferString(queryStr + "\n")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing query: %v, stderr: %s", err, stderr.String())
	}

	return out.String(), nil
}

func logToFile(filePath string, elements ...string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush() // Ensure all buffered data is written

	timestamp := time.Now().Format(time.RFC3339)

	for _, element := range elements {
		_, err := fmt.Fprintf(writer, "%s - %s\n", timestamp, element)
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	}

	return nil
}
