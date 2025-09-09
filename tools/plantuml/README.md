# PlantUML Renderer Tool

A Model Context Protocol (MCP) server tool for rendering PlantUML diagrams in multiple output formats.

## Features

- **Single Tool**: `render_plantuml` - Renders PlantUML diagrams from plain text or encoded format
- **Multiple Input Formats**: Supports both plain text PlantUML code and PlantUML encoded strings
- **Multiple Output Formats**:
  - `svg` - SVG vector graphics (default)
  - `png` - PNG image (base64 encoded data URI)
  - `txt` - Plain text representation
  - `encoded` - PlantUML encoded format

## Tool Description

### render_plantuml

**Description**: Render PlantUML diagrams from plain text or encoded format. Supports multiple output formats including SVG, PNG, plain text, and encoded versions.

**Parameters**:
- `plantuml_code` (required): PlantUML diagram code in plain text format (e.g., `@startuml\nAlice -> Bob: Hello\n@enduml`) or in encoded format (base64-like encoded string). The tool will automatically detect the format.
- `output_format` (optional): Output format: `svg` for SVG vector graphics (default), `png` for PNG image, `txt` for plain text representation, `encoded` for PlantUML encoded format

## Usage Examples

### Basic Sequence Diagram

**Input**:
```json
{
  "plantuml_code": "@startuml\nAlice -> Bob: Hello\nBob -> Alice: Hi!\n@enduml",
  "output_format": "svg"
}
```

### Converting to Encoded Format

**Input**:
```json
{
  "plantuml_code": "@startuml\nAlice -> Bob: Hello\nBob -> Alice: Hi!\n@enduml",
  "output_format": "encoded"
}
```

**Output**: `SYWkIImgAStDuNBCoKnELT2rKt3AJx9Iy4ZDoSddSifF0ec0fQmKF38LkHnIyr9AStC00G00__y`

### Decoding to Plain Text

**Input**:
```json
{
  "plantuml_code": "SYWkIImgAStDuNBCoKnELT2rKt3AJx9Iy4ZDoSddSifF0ec0fQmKF38LkHnIyr9AStC00G00__y",
  "output_format": "txt"
}
```

## Implementation Details

- **PlantUML Text Encoding**: Implements the official PlantUML text encoding algorithm using Deflate compression and custom base64-like encoding
- **Online Rendering**: Uses the official PlantUML server (http://www.plantuml.com/plantuml) for SVG and PNG rendering
- **Automatic Format Detection**: Automatically detects if input is plain text or encoded PlantUML
- **Error Handling**: Provides detailed error messages for invalid inputs or network failures

## Building

```bash
go build -o plantuml
```

## Requirements

- Internet connection required for SVG and PNG rendering (uses PlantUML server)
- No external dependencies for encoding/decoding operations

## PlantUML Encoding Format

The tool implements the standard PlantUML text encoding format:
1. UTF-8 encoding
2. Deflate compression
3. Custom base64-like encoding using character set: `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_`

This allows for sharing PlantUML diagrams as compressed URL-safe strings.
