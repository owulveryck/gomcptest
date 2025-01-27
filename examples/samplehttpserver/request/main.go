 package main

 import (
 	"bufio"
 	"fmt"
 	"log"
 	"os"
 	"regexp"
 	"strings"
 	"time"
 )

 // Function to extract log records within a date range
 func filterLogsByDate(logFilePath string, startDate, endDate time.Time) (string, error) {
 	file, err := os.Open(logFilePath)
 	if err != nil {
 		return "", fmt.Errorf("error opening log file: %w", err)
 	}
 	defer file.Close()

 	var filteredLogs strings.Builder
 	scanner := bufio.NewScanner(file)
 	logLineRegex := regexp.MustCompile(`\[(\d{2}\/\w{3}\/\d{4}:\d{2}:\d{2}:\d{2} [-+]\d{4})\]`) // Regex to extract timestamp

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
 		return "", fmt.Errorf("error reading log file: %w", err)
 	}

 	return filteredLogs.String(), nil
 }

 func main() {
 	// Example Usage:
 	logFilePath := "access.log" // Replace with your log file path
 	startDate, _ := time.Parse("2006-01-02 15:04:05 -0700", "2025-01-24 09:00:00 +0100")
 	endDate, _ := time.Parse("2006-01-02 15:04:05 -0700", "2025-01-24 12:00:00 +0100")

 	filteredLogs, err := filterLogsByDate(logFilePath, startDate, endDate)
 	if err != nil {
 		log.Fatalf("Error filtering logs: %v", err)
 	}

 	fmt.Println("Filtered Logs:\n", filteredLogs)
 }
