# Changelog

## [v1.3.1] - 2025-01-05

### 🐛 Bug Fixes
- **UI Message Display**: Remove empty assistant message boxes between tool calls for cleaner interface
- **Copy Functionality**: Improve markdown copy functionality and button positioning
- **Button Behavior**: Improve copy button positioning and selection behavior

### ✨ Features
- **Branding**: Rebrand application from "gomcptest simple chat" to "AgentFlow"
- **Image Upload**: Add image upload functionality to chat UI
- **Tool Notifications**: Add clickable tool notifications with popup details and proper alignment
- **Persistent Notifications**: Add persistent tool notification in conversation flow

### 🔧 Technical Improvements
- **Response Display**: Resolve assistant response display issue after tool notifications
- **Tool Ordering**: Correct tool notification ordering to appear before assistant response
- **Message Persistence**: Save assistant messages in streaming responses and adjust popup timeout

## Changes from v1.1 to current version

### ✨ Major New Features

#### 🎉 Complete Chat UI Implementation
- **Brand New Chat Interface**: Added comprehensive web-based chat UI for the OpenAI server
- **Embedded in Server**: Chat UI accessible at `/ui` endpoint, fully integrated with the OpenAI server
- **Professional Design**: Modern, responsive interface with markdown support and syntax highlighting
- **Real-time Streaming**: Live response streaming with visual indicators and typing animations
- **Conversation Management**: Full conversation system with sidebar, auto-titling, rename/delete, and localStorage persistence
- **System Prompt Support**: Customizable system prompts per conversation with save/reset functionality
- **Message Editing**: Edit any message and regenerate responses from that point
- **Dynamic Model Selection**: Choose AI models directly from the interface
- **Tool Execution Visibility**: Real-time display of tool calls and responses with detailed popups
- **SVG Rendering**: Full support for visualizations including Wardley maps and diagrams

#### Enhanced Tool Integration
- **Sleep MCP Tool**: Added new sleep tool for debugging and testing UI interactions
- **Streaming Events**: Comprehensive streaming event system showing tool execution in real-time
- **Tool Call Correlation**: Proper matching of tool calls with their responses using unique IDs

### 🔧 Technical Improvements

#### Streaming Architecture
- **Event-Driven Streaming**: Refactored chat engine with StreamEvent interface for unified event handling
- **Gemini Integration**: Enhanced Gemini provider with tool call and response events
- **Provider Switch**: Changed default provider from Claude to Gemini
- **Enhanced Flags**: Added `--withAllEvents` flag for comprehensive event streaming

#### Server Architecture
- **Improved Routing**: Updated to use `http.ServeMux` for better endpoint management
- **CORS Handling**: Proper cross-origin request support for the chat UI
- **Resource Management**: Better cleanup and error handling throughout the system

### 🎨 User Experience Features

#### Interface Design
- **Responsive Layout**: Optimized for different screen sizes with minimal margins
- **Professional Theme**: Clean color scheme maximizing usable space
- **Sliding Sidebar**: Hover-activated conversation management with arrow indicator
- **Visual Feedback**: Success/error indicators, loading states, and progress animations
- **Accessibility**: Proper contrast, readable fonts, and intuitive navigation

#### Interactive Features
- **Conversation Persistence**: All conversations saved locally across browser sessions
- **Auto-titling**: Conversations automatically named from first message
- **Typing Indicators**: Visual feedback while AI is processing requests
- **Tool Popups**: Detailed execution status and results in overlay windows
- **Message History**: Complete conversation replay and editing capabilities

### 🐛 Bug Fixes

#### SVG Display
- **Object Tag Rendering**: Fixed SVG display using `<object>` tags for better browser compatibility
- **Responsive Sizing**: Improved SVG sizing to use full available space with proper aspect ratios
- **Marked.js Compatibility**: Fixed object parameter handling in custom image renderer

#### Tool Popup System
- **Popup Persistence**: Resolved issues with popups getting stuck or not closing properly
- **Correlation Fixes**: Ensured tool calls properly match with their responses
- **Timeout Handling**: Added appropriate timeouts and fallbacks for unresponsive tools
- **Visual Improvements**: Enhanced popup size, readability, and formatting

### 📋 Development Notes

This release represents a major milestone with the addition of a complete web-based chat interface. The UI provides a professional, feature-rich environment for interacting with the MCP system and demonstrates the full capabilities of the agentic framework.

The chat UI includes advanced features like real-time tool execution visibility, conversation management, and comprehensive streaming support, making it a powerful tool for testing and using MCP-based agents.