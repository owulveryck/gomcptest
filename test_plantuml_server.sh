#!/bin/bash

# Test script for PlantUML tool with configurable server URL

echo "Testing PlantUML tool with configurable PLANTUML_SERVER..."

# Test 1: Default server (should be http://localhost:9999/plantuml)
echo -e "\n=== Test 1: Default server (localhost:9999) ==="
export GCP_PROJECT=test-project
export LOG_LEVEL=DEBUG
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "render_plantuml", "arguments": {"plantuml_code": "@startuml\nAlice -> Bob: Hello\n@enduml", "output_format": "txt"}}}' | timeout 10s ./bin/plantuml 2>&1 | grep -E "(server|url|Making request)"

# Test 2: Custom server URL
echo -e "\n=== Test 2: Custom server URL ==="
export PLANTUML_SERVER=http://custom-server:8080/plantuml
export LOG_LEVEL=DEBUG
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "render_plantuml", "arguments": {"plantuml_code": "@startuml\nAlice -> Bob: Hello\n@enduml", "output_format": "svg"}}}' | timeout 10s ./bin/plantuml 2>&1 | grep -E "(server|url|Making request)"

# Test 3: Another custom server URL
echo -e "\n=== Test 3: Another custom server URL ==="
export PLANTUML_SERVER=https://my-plantuml.example.com/api/plantuml
export LOG_LEVEL=DEBUG
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "render_plantuml", "arguments": {"plantuml_code": "@startuml\nAlice -> Bob: Hello\n@enduml", "output_format": "svg"}}}' | timeout 10s ./bin/plantuml 2>&1 | grep -E "(server|url|Making request)"

# Test 4: Test with original default but verify it's using the config
echo -e "\n=== Test 4: Verify default server configuration ==="
unset PLANTUML_SERVER
export LOG_LEVEL=DEBUG
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "render_plantuml", "arguments": {"plantuml_code": "@startuml\nAlice -> Bob: Hello\n@enduml", "output_format": "svg"}}}' | timeout 10s ./bin/plantuml 2>&1 | grep -E "(server|url|Making request)"

echo -e "\nTesting completed!"