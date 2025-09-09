#!/bin/bash

# Test script for PlantUML tool with different logging configurations

echo "Testing PlantUML tool with different log configurations..."

# Test 1: Default logging (INFO level to STDERR)
echo -e "\n=== Test 1: Default logging (INFO to STDERR) ==="
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "render_plantuml", "arguments": {"plantuml_code": "@startuml\nAlice -> Bob: Hello\n@enduml", "output_format": "txt"}}}' | timeout 10s ./bin/plantuml

# Test 2: DEBUG logging to STDERR
echo -e "\n=== Test 2: DEBUG logging to STDERR ==="
export LOG_LEVEL=DEBUG
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "render_plantuml", "arguments": {"plantuml_code": "@startuml\nAlice -> Bob: Hello\n@enduml", "output_format": "txt"}}}' | timeout 10s ./bin/plantuml

# Test 3: ERROR logging to file
echo -e "\n=== Test 3: ERROR logging to file ==="
export LOG_LEVEL=ERROR
export LOG_OUTPUT=/tmp/plantuml_test.log
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "render_plantuml", "arguments": {"plantuml_code": "@startuml\nAlice -> Bob: Hello\n@enduml", "output_format": "txt"}}}' | timeout 10s ./bin/plantuml

echo -e "\n=== Checking log file content ==="
cat /tmp/plantuml_test.log

# Test 4: Test invalid PlantUML (should not call plantuml.com)
echo -e "\n=== Test 4: Invalid PlantUML (testing retry protection) ==="
export LOG_LEVEL=DEBUG
export LOG_OUTPUT=STDERR
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "render_plantuml", "arguments": {"plantuml_code": "@startuml\nInvalidSyntax -> : test\n@enduml", "output_format": "svg"}}}' | timeout 15s ./bin/plantuml

echo -e "\nTesting completed!"