#!/bin/bash
set -e

# Create test data
echo "Creating test data..."
cat > test_data.csv << EOF
id,name,sales
1,Alpha,1000
2,Beta,2500
3,Gamma,1800
4,Delta,3200
5,Epsilon,950
EOF

echo "Test data created as test_data.csv"

# Build the duckdb server
echo "Building duckdb server..."
go build -o duckdbserver

# Start the server with various tests using JSON-RPC format
echo "Testing duckdb server..."

# Initialize and list tools
(
  cat <<\EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"example-client","version":"1.0.0"},"capabilities":{}}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
EOF
) | ./duckdbserver

echo
echo "Running simple query on local test data..."
(
  cat <<\EOF
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"duckdb","arguments":{"query":"SELECT * FROM 'test_data.csv' WHERE sales > 1500"}}}
EOF
) | ./duckdbserver

echo
echo "Running aggregate query..."
(
  cat <<\EOF
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"duckdb","arguments":{"query":"SELECT SUM(sales) as total_sales FROM 'test_data.csv'"}}}
EOF
) | ./duckdbserver

echo
echo "Running query with filter and order..."
(
  cat <<\EOF
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"duckdb","arguments":{"query":"SELECT name, sales FROM 'test_data.csv' ORDER BY sales DESC LIMIT 3"}}}
EOF
) | ./duckdbserver

echo
echo "Testing with invalid syntax (should show error)..."
(
  cat <<\EOF
{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"duckdb","arguments":{"query":"SELECT BROKEN QUERY"}}}
EOF
) | ./duckdbserver

# Test with remote data if available
echo
echo "Testing with Hugging Face dataset (if accessible)..."
(
  cat <<\EOF
{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"duckdb","arguments":{"query":"SELECT * FROM 'hf://datasets/fka/awesome-chatgpt-prompts/prompts.csv' LIMIT 5;"}}}
EOF
) | ./duckdbserver

echo
echo "Tests complete!"
