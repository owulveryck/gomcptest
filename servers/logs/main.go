package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
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
		"Log Seeker 🚀",
		"1.0.0",
	)

	// Add tool
	tool := mcp.NewTool("find_logs",
		mcp.WithDescription("extracts log records from a specified log file that fall within a given time range. It reads the log file line by line, identifies log entries by parsing their timestamp, and includes only those entries that occur after the provided start date and before the provided end date in the returned string."),
		mcp.WithString("start_date",
			mcp.Required(),
			mcp.Description("The start date of the log extraction in the format 2025-01-24 12:00:00 +0100"),
		),
		mcp.WithString("end_date",
			mcp.Required(),
			mcp.Description("The end date of the log extraction in the format 2025-01-24 12:00:00 +0100"),
		),
		mcp.WithString("server_name",
			mcp.Required(),
			mcp.Description("The name of the server to get the logs from"),
		),
	)

	// Add tool handler
	s.AddTool(tool, logHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func logHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logToFile(logFilePath, "opening log file")
	file, err := os.Open(accessLogPath)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("error opening log file: %w", err)
	}
	defer file.Close()

	serverNameStr, ok := request.Params.Arguments["server_name"].(string)
	if !ok {
		return mcp.NewToolResultError("server_name must be a string"), nil
	}
	startDateStr, ok := request.Params.Arguments["start_date"].(string)
	if !ok {
		return mcp.NewToolResultError("start_date must be a string"), nil
	}
	endDateStr, ok := request.Params.Arguments["end_date"].(string)
	if !ok {
		return mcp.NewToolResultError("end_date must be a string"), nil
	}
	startDate, err := time.Parse("2006-01-02 15:04:05 -0700", startDateStr)
	if err != nil {
		return mcp.NewToolResultError("wrong start_date format:" + err.Error()), nil
	}
	endDate, err := time.Parse("2006-01-02 15:04:05 -0700", endDateStr)
	if err != nil {
		return mcp.NewToolResultError("wrong end_date format:" + err.Error()), nil
	}
	if serverNameStr != "myserver" {
		return mcp.NewToolResultError("Unknown server " + serverNameStr), nil
	}
	logToFile(logFilePath, startDateStr)
	logToFile(logFilePath, endDateStr)
	logToFile(logFilePath, serverNameStr)
	var filteredLogs strings.Builder
	scanner := bufio.NewScanner(file)
	logLineRegex := regexp.MustCompile(`\[(\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2} [-+]\d{4})\]`) // Regex to extract timestamp

	log.Println("Scanning logs")
	for scanner.Scan() {
		line := scanner.Text()
		matches := logLineRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			logTimeStr := matches[1]
			logTime, err := time.Parse("02/Jan/2006:15:04:05 -0700", logTimeStr)
			if err != nil {
				log.Printf("Error parsing log timestamp: %v, skipping line: %s\n", err, line)
				continue // Skip lines with unparseable timestamps
			}

			if logTime.After(startDate) && logTime.Before(endDate) {
				filteredLogs.WriteString(line)
				filteredLogs.WriteString("\n")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}
	logToFile(logFilePath, filteredLogs.String())

	return mcp.NewToolResultText(filteredLogs.String()), nil
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
