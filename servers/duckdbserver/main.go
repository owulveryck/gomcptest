package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	logFilePath   string
	accessLogPath string
)

func main() {
	//
	// Create MCP server
	s := server.NewMCPServer(
		"DuckDB 🚀",
		"1.0.0",
	)

	// Add tool
	tool := mcp.NewTool("duckdb",
		mcp.WithDescription(`The duckdb tool allows you to execute SQL queries against data stored in various file formats. It uses DuckDB, an in-process SQL OLAP database management system, to efficiently process and extract information from these files. This tool is particularly useful for analyzing and manipulating data directly from files without needing to load them into a separate database.

**File Access and Formats:**

The duckdb tool can access data from the following file types and locations:

*   **Local Files:** Specify the complete path to a file on the local filesystem (e.g., /path/to/my_data.csv).
*   **Remote Files on Hugging Face:**  Access files stored on the Hugging Face Hub by prefixing the file path with hf: (e.g., hf://username/repository/data/my_data.parquet).
*   **Supported File Formats:** The tool automatically detects and handles the following file formats:
    *   CSV
    *   Parquet
    *   JSON

*   **Wildcards:** You can use wildcards (*) in the file path to query multiple files at once (e.g., /path/to/data/sales_*.csv). This is useful for querying data that is split across multiple files with a consistent naming pattern.

**Important Considerations for Querying:**

*   **DuckDB Compatibility:** Ensure that your SQL query is compatible with DuckDB's SQL dialect. Refer to the DuckDB documentation for specific syntax and supported functions.
*   **File Structure:** You should be aware of the structure of the data within the file(s) you are querying (e.g., column names, data types) in order to write effective SQL queries.
*   **Error Handling:** If the query is invalid or if there are issues accessing the file, the tool will return an error.

**Output:**

The tool returns a dictionary containing the results of the SQL query. The structure of the dictionary will depend on the specific query that was executed. It will typically include column names and the corresponding data retrieved by the query.
			`),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to execute (compatible with DUCKDB). A SQL query (compatible with DuckDB syntax) that you want to execute. This query should be designed to retrieve the specific information you need from the target file(s)."),
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
	res, err := executeDuckDBQuery(queryStr)
	if err != nil {
		return mcp.NewToolResultError("duckdb encountered an error: " + err.Error()), nil
	}

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
