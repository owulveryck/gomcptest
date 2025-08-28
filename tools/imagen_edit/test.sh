#!/bin/bash
set -e

echo "Testing imagen_edit MCP server..."

# Convert test image to base64
echo "Converting test image to base64..."
BASE64_IMAGE=$(base64 -i test_data/generative-ai_image_table.png | tr -d '\n')

# Build the imagen_edit tool if needed
if [ ! -f "./bin/imagen_edit" ]; then
    echo "Building imagen_edit tool..."
    cd ../..
    make bin/imagen_edit
    cd tools/imagen_edit
fi

# Initialize and list tools
echo "Initializing MCP server and listing tools..."
(
  cat <<\EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"example-client","version":"1.0.0"},"capabilities":{}}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
EOF
) | ../../bin/imagen_edit

echo
echo "Testing image editing - adding flowers to the table..."
echo "Using GCP_PROJECT: $GCP_PROJECT"
echo "Using GCP_REGION: $GCP_REGION"
# Override region to global for Imagen Edit API
export GCP_REGION=global
export LOG_LEVEL=DEBUG
(
  cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"example-client","version":"1.0.0"},"capabilities":{}}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"imagen_edit","arguments":{"base64_image":"${BASE64_IMAGE}","mime_type":"image/png","edit_instruction":"Add beautiful flowers on the table","temperature":1.0}}}
EOF
) | timeout 30 ../../bin/imagen_edit

echo
echo "Test complete!"