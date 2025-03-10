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
		mcp.WithDescription(`Execute SQL queries on files using DuckDB, an in-process analytical database engine that reads directly from files without importing.

**Supported File Formats:**
- CSV
- Parquet
- JSON
- And many others supported by DuckDB

**Usage Examples:**
1. Query local CSV file:
   SELECT * FROM '/path/to/data.csv' LIMIT 10

2. Filter data from Parquet file:
   SELECT column1, column2 FROM '/path/to/data.parquet' WHERE condition

3. Aggregate data across multiple files:
   SELECT category, SUM(amount) 
   FROM '/path/to/data/*.csv' 
   GROUP BY category

4. Join data from different file formats:
   SELECT a.id, a.name, b.value
   FROM '/path/to/users.csv' a
   JOIN '/path/to/transactions.parquet' b
   ON a.id = b.user_id

5. Load remote files (HTTP, S3, etc.):
   SELECT * FROM 'https://example.com/data.csv'

**Capabilities:**
- Powerful SQL analytics directly on files
- Schema inference
- Wildcard path patterns
- Multi-file querying
- Cross-format joins
- Efficient columnar processing

For more details on DuckDB's SQL syntax and functions, visit: https://duckdb.org/docs/sql/introduction`),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("SQL query to execute using DuckDB syntax. Query directly from files like '/path/to/data.csv' without needing to import data first."),
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