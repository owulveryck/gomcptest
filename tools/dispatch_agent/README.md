# dispatch_agent MCP Service

This service implements the `dispatch_agent` function from Claude Function Specifications as an MCP server.

## Description

Launches a new agent that has access to the following tools: View, GlobTool, GrepTool, LS, ReadNotebook, WebFetchTool. When you are searching for a keyword or file and are not confident that you will find the right match on the first try, use the Agent tool to perform the search for you. For example, if you're searching for "config" or asking "which file does X?", the dispatch_agent can efficiently perform this search.

## Parameters

- `prompt` (string, required): The task for the agent to perform

## Usage Notes

- Recommended for:
  - Keyword searches like "config" or "logger"
  - Questions like "which file does X?"
  - Open-ended searches requiring multiple rounds of globbing and grepping
- If searching for a specific file path, use View or GlobTool directly instead
- If searching for a specific class definition like "class Foo", use GlobTool instead
- Launch multiple agents concurrently when possible for better performance
- Agent is stateless and returns a single message back to you
- The result returned by the agent is not visible to the user until you show it
- Agent cannot use Bash, Replace, Edit, or NotebookEditCell
- Agent cannot modify files

## Implementation Details

The dispatch_agent implementation includes:
- Intelligent prompt analysis to determine search intent
- Task execution tracking with timestamps and success metrics
- Concurrent tool execution for optimal performance
- Pattern matching for file and code searches
- Comprehensive response composition from multiple tool results

## Output

A single message with the results of the search or analysis.

## Running the Service

```bash
cd cmd
go run main.go
```

## Example

```json
{
  "name": "dispatch_agent",
  "params": {
    "prompt": "Find all config files in the project"
  }
}
```