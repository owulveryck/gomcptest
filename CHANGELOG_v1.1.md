# Changelog: v1.0.0 → v1.1

Release Date: 2025-08-28

## Summary

Version 1.1 represents a significant enhancement to the gomcptest project, focusing on improved Google Gemini integration, comprehensive error handling, and expanded MCP tool capabilities.

## 🚀 New Features

### Enhanced Google Gemini Integration
- **Upgraded Dependencies**: Updated `google.golang.org/genai` from v1.16.0 to v1.22.0
- **Function Calling Configuration**: Added FunctionCallingConfig with validated mode for better tool handling
- **Tool Configuration**: Implemented ToolConfig with FunctionCallingConfigModeValidated for improved reliability

### New MCP Tools
- **PlantUML Check Tool**: Added comprehensive PlantUML syntax validation tool
- **Enhanced Imagen Edit Tool**: Expanded image editing capabilities with HTTP server functionality

### Advanced Error Handling
- **UNEXPECTED_TOOL_CALL Handling**: Added specific error handling for unexpected tool call scenarios
- **Detailed Error Messages**: Implemented comprehensive error reporting with troubleshooting suggestions
- **Stream Processing Improvements**: Enhanced error recovery in both streaming and non-streaming chat processors

## 🔧 Improvements

### Logging Enhancements
- **Content-Based Logging**: Improved logging to show actual content instead of memory pointers
- **Structured Error Reporting**: Added detailed context for debugging tool call failures
- **Better Debug Information**: Enhanced logging with model versions, response IDs, and usage metadata

### Development Workflow
- **Claude Code Permissions**: Updated settings to allow additional git commands and file access
- **Build System**: Enhanced Makefile for better tool management
- **Resource Management**: Improved handling of external dependencies

### Error Recovery
- **MCP Server Failures**: Comprehensive error handling for MCP server connection failures
- **Tool Call Validation**: Better validation and error recovery for malformed tool calls
- **Safety Ratings**: Enhanced handling of content filtering and safety issues

## 📁 File Changes

### Core Components Modified (7 files changed, 147 insertions, 12 deletions)
- `go.mod` & `go.sum`: Dependency updates
- `host/openaiserver/chatengine/gcp/`: Complete overhaul of GCP chat engine
  - `nonstream.go`: Enhanced non-streaming request handling
  - `nonstream_processor.go`: Improved response processing with error handling
  - `stream.go`: Updated streaming functionality
  - `stream_processor.go`: Comprehensive error handling and tool call management

### New Tool Additions
- `tools/imagen_edit/`: Complete image editing tool suite (830+ lines)
- `tools/plantuml_check/`: PlantUML validation tool (131+ lines)

### Configuration Updates
- `.claude/settings.local.json`: Extended permissions for development workflow

## 🐛 Bug Fixes

- **Compilation Errors**: Fixed various compilation issues across the codebase
- **Tool Call Failures**: Improved handling of malformed or invalid tool calls
- **Stream Processing**: Enhanced stability in streaming chat responses
- **Resource Cleanup**: Better management of external resources and connections

## 📊 Statistics

- **Total Files Changed**: 64
- **Lines Added**: 6,214
- **Lines Removed**: 25
- **New Tools**: 2 (imagen_edit, plantuml_check)
- **Core Commits**: 11 commits between versions

## 🔄 Migration Notes

### For Developers
- Update your `go.mod` to use the new genai version (v1.22.0)
- The new function calling configuration provides better tool validation
- Enhanced error messages will provide more debugging context

### For Users
- Improved stability when using tools with the chat engine
- Better error reporting when tool calls fail
- New image editing and PlantUML validation capabilities

## 🎯 Key Commits

1. **81066cf**: feat: Update Google Gemini integration and dependencies
2. **b7571b9**: feat: Add comprehensive error handling and new MCP tools
3. **5ed8692**: Implement comprehensive error handling for MCP server failures
4. **3d76792**: Improve logging to show actual content instead of memory pointers
5. **b36a132**: Fix compilation errors and improve error handling across the codebase

## 🚀 What's Next

This release establishes a more robust foundation for:
- Advanced agent interactions with better error recovery
- Enhanced tool ecosystem with validation capabilities
- Improved developer experience with better debugging information
- Preparation for future MCP protocol enhancements