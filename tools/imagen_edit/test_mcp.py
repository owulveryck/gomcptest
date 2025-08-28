#!/usr/bin/env python3

import json
import subprocess
import sys
import base64
import os

def test_imagen_edit_mcp():
    # Read the test image and encode it to base64
    image_path = "test_data/generative-ai_image_table.png"
    
    if not os.path.exists(image_path):
        print(f"Error: Test image not found at {image_path}")
        return False
    
    with open(image_path, "rb") as f:
        image_data = base64.b64encode(f.read()).decode('utf-8')
    
    # Create MCP request to add flowers to the table
    mcp_request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "imagen_edit",
            "arguments": {
                "base64_image": image_data,
                "mime_type": "image/png",
                "edit_instruction": "Add beautiful flowers on the table",
                "temperature": 1.0
            }
        }
    }
    
    # Start the MCP server
    try:
        print("Starting MCP server...")
        process = subprocess.Popen(
            ["./bin/imagen_edit"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            cwd="/Users/olivier.wulveryck/github.com/owulveryck/gomcptest"
        )
        
        # Send the request
        print("Sending request to add flowers to the table...")
        request_json = json.dumps(mcp_request) + "\n"
        
        stdout, stderr = process.communicate(input=request_json, timeout=60)
        
        print("STDOUT:")
        print(stdout)
        
        if stderr:
            print("STDERR:")
            print(stderr)
        
        return process.returncode == 0
        
    except subprocess.TimeoutExpired:
        print("Request timed out")
        process.kill()
        return False
    except Exception as e:
        print(f"Error running MCP server: {e}")
        return False

if __name__ == "__main__":
    success = test_imagen_edit_mcp()
    sys.exit(0 if success else 1)