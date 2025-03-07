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

const (
	dateTimeFormat = "2006-01-02 15:04:05 -0700"
	logTimestampFormat = "02/Jan/2006:15:04:05 -0700"
)

// logFilePath is the path to the application log file.
var logFilePath = "/tmp/my_log.txt"

// accessLogPath is the path to the access log file to be parsed.
var accessLogPath string

func main() {
	flag.StringVar(&accessLogPath, "log", "/Users/olivier.wulveryck/github.com/owulveryck/gomcptest/examples/samplehttpserver/access.log", "Path to the access log file")
	flag.Parse()

	if err := logToFile("Starting the application"); err != nil {
		log.Fatalf("Error logging: %v", err)
	}

	if err := logToFile("Another log entry"); err != nil {
		log.Fatalf("Error logging: %v", err)
	}

	fmt.Println("Logs written to:", logFilePath)

	// Create MCP server
	s := server.NewMCPServer(
		"Log Seeker 🚀",
		"1.0.0",
	)

	// Add tool
	tool := mcp.NewTool("find_logs",
		mcp.WithDescription("extracts log records from a specified log file that fall within a given time range."),
		mcp.WithString("start_date",
			mcp.Required(),
			mcp.Description("The start date of the log extraction in the format "+dateTimeFormat),
		),
		mcp.WithString("end_date",
			mcp.Required(),
			mcp.Description("The end date of the log extraction in the format "+dateTimeFormat),
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
	if err := logToFile("opening log file"); err != nil {
		return nil, fmt.Errorf("error logging: %w", err)
	}

	params := request.Params.Arguments
	serverName, ok := params["server_name"].(string)
	if !ok {
		return mcp.NewToolResultError("server_name must be a string"), nil
	}

	startDateStr, ok := params["start_date"].(string)
	if !ok {
		return mcp.NewToolResultError("start_date must be a string"), nil
	}

	endDateStr, ok := params["end_date"].(string)
	if !ok {
		return mcp.NewToolResultError("end_date must be a string"), nil
	}

	startDate, err := time.Parse(dateTimeFormat, startDateStr)
	if err != nil {
		return mcp.NewToolResultError("wrong start_date format: " + err.Error()), nil
	}

	endDate, err := time.Parse(dateTimeFormat, endDateStr)
	if err != nil {
		return mcp.NewToolResultError("wrong end_date format: " + err.Error()), nil
	}

	if serverName != "myserver" {
		return mcp.NewToolResultError("Unknown server " + serverName), nil
	}

	if err := logToFile(startDateStr, endDateStr, serverName); err != nil {
		return nil, fmt.Errorf("error logging: %w", err)
	}

	filteredLogs, err := filterLogs(serverName, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error filtering logs: %w", err)
	}

	if err := logToFile(filteredLogs); err != nil {
		return nil, fmt.Errorf("error logging: %w", err)
	}

	return mcp.NewToolResultText(filteredLogs), nil
}

// filterLogs filters the logs based on server name and date range.
func filterLogs(serverName string, startDate, endDate time.Time) (string, error) {
	file, err := os.Open(accessLogPath)
	if err != nil {
		return "", fmt.Errorf("error opening log file: %w", err)
	}
	defer file.Close()

	logLineRegex := regexp.MustCompile(`\[(\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2} [-+]\d{4})\]`)
	var filteredLogs strings.Builder
	scanner := bufio.NewScanner(file)

	log.Println("Scanning logs")
	for scanner.Scan() {
		line := scanner.Text()
		matches := logLineRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			logTimeStr := matches[1]
			logTime, err := time.Parse(logTimestampFormat, logTimeStr)
			if err != nil {
				log.Printf("Error parsing log timestamp: %v, skipping line: %s\n", err, line)
				continue
			}

			if logTime.After(startDate) && logTime.Before(endDate) {
				filteredLogs.WriteString(line)
				filteredLogs.WriteString("\n")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading log file: %w", err)
	}

	return filteredLogs.String(), nil
}

// logToFile logs messages to the specified log file.
func logToFile(elements ...string) error {
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	timestamp := time.Now().Format(time.RFC3339)
	for _, element := range elements {
		if _, err := fmt.Fprintf(writer, "%s - %s\n", timestamp, element); err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	}

	return nil
}
