#!/bin/bash
set -e

echo "Testing imagen_edit tool directly..."

# Read the base64 encoded image
BASE64_IMAGE=$(cat base64_image.txt)

# Set environment variables for the test
export GCP_REGION=global
export LOG_LEVEL=DEBUG

echo "Starting MCP server and sending image edit request..."

# Create a single session with initialize and tool call
(
  cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test-client","version":"1.0.0"},"capabilities":{}}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"imagen_edit","arguments":{"base64_image":"${BASE64_IMAGE}","mime_type":"image/png","edit_instruction":"Add beautiful flowers on the table","temperature":0.8}}}
EOF
) | timeout 180 ../../bin/imagen_edit

echo "Test completed."