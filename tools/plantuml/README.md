# PlantUML Renderer Tool

A Model Context Protocol (MCP) server tool for generating PlantUML diagram URLs and decoding PlantUML from URLs or encoded strings.

## Features

- **Two Main Tools**:
  - `render_plantuml` - Generates PlantUML diagram URLs with syntax validation and error correction
  - `decode_plantuml_url` - Decodes PlantUML URLs or encoded strings back to plain text
- **URL-Based Architecture**: Returns URLs pointing to PlantUML server instead of raw content
- **Automatic Error Correction**: Uses GenAI to fix PlantUML syntax errors automatically
- **Local PlantUML Server**: Works with local PlantUML server for validation and rendering

## Tool Descriptions

### render_plantuml

**Description**: Generate PlantUML diagram URLs from plain text. Returns URLs pointing to the PlantUML server for SVG/PNG rendering. Validates syntax and corrects errors using GenAI if needed.

**Parameters**:
- `plantuml_code` (required): PlantUML diagram code in plain text format (e.g., `@startuml\nAlice -> Bob: Hello\n@enduml`)
- `output_format` (optional): Output format: `svg` for SVG URL (default), `png` for PNG URL

**Returns**: URL in format `http://localhost:9999/plantuml/[svg|png]/ENCODED_PLANTUML`

### decode_plantuml_url

**Description**: Decode PlantUML URLs or encoded strings back to plain text PlantUML code.

**Parameters**:
- `url_or_encoded` (required): Either a full PlantUML server URL (e.g., `http://localhost:9999/plantuml/svg/ENCODED`) or just the encoded part (e.g., `ENCODED`)

**Returns**: Original PlantUML code in plain text format

## Usage Examples

### Generating a PlantUML SVG URL

**Input**:
```json
{
  "plantuml_code": "@startuml\nAlice -> Bob: Hello\nBob -> Alice: Hi!\n@enduml",
  "output_format": "svg"
}
```

**Output**: `http://localhost:9999/plantuml/svg/SYWkIImgAStDuNBCoKnELT2rKt3AJx9Iy4ZDoSddSifF0ec0fQmKF38LkHnIyr9AStC00G00__y`

### Generating a PlantUML PNG URL

**Input**:
```json
{
  "plantuml_code": "@startuml\nAlice -> Bob: Hello\nBob -> Alice: Hi!\n@enduml",
  "output_format": "png"
}
```

**Output**: `http://localhost:9999/plantuml/png/SYWkIImgAStDuNBCoKnELT2rKt3AJx9Iy4ZDoSddSifF0ec0fQmKF38LkHnIyr9AStC00G00__y`

### Decoding a PlantUML URL

**Input**:
```json
{
  "url_or_encoded": "http://localhost:9999/plantuml/svg/SYWkIImgAStDuNBCoKnELT2rKt3AJx9Iy4ZDoSddSifF0ec0fQmKF38LkHnIyr9AStC00G00__y"
}
```

**Output**: 
```
@startuml
Alice -> Bob: Hello
Bob -> Alice: Hi!
@enduml
```

### Decoding just the Encoded Part

**Input**:
```json
{
  "url_or_encoded": "SYWkIImgAStDuNBCoKnELT2rKt3AJx9Iy4ZDoSddSifF0ec0fQmKF38LkHnIyr9AStC00G00__y"
}
```

**Output**: 
```
@startuml
Alice -> Bob: Hello
Bob -> Alice: Hi!
@enduml
```

## Implementation Details

- **URL-Based Architecture**: Returns URLs instead of diagram content for better performance and client-side caching
- **Local PlantUML Server**: Uses configurable local PlantUML server (default: `http://localhost:9999/plantuml`)
- **Syntax Validation**: Validates PlantUML syntax by calling the `/txt/` endpoint before generating URLs
- **Automatic Error Correction**: Uses Google Gemini AI to automatically fix PlantUML syntax errors
- **PlantUML Text Encoding**: Implements the official PlantUML text encoding algorithm using Deflate compression and custom base64-like encoding
- **URL Pattern Recognition**: Supports decoding from full URLs or just encoded strings
- **Error Handling**: Provides detailed error messages for invalid inputs or server failures

## Configuration

The tool uses environment variables for configuration:

- `PLANTUML_SERVER`: PlantUML server URL (default: `http://localhost:9999/plantuml`)
- `GCP_PROJECT`: Google Cloud Project ID (required for GenAI error correction)
- `GCP_REGION`: Google Cloud Region (default: `us-central1`)
- `LOG_LEVEL`: Logging level (default: `ERROR`)

## Building

```bash
go build -o plantuml
```

## Requirements

- Local PlantUML server running (for validation and rendering)
- Google Cloud Project with Vertex AI enabled (for error correction)
- No external dependencies for encoding/decoding operations

## URL Format

Generated URLs follow the pattern:
```
http://localhost:9999/plantuml/[svg|png]/ENCODED_PLANTUML
```

Where `ENCODED_PLANTUML` is the PlantUML code encoded using:
1. UTF-8 encoding
2. Deflate compression  
3. Custom base64-like encoding using character set: `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_`

## Error Correction

When the tool detects PlantUML syntax errors, it automatically:
1. Sends the code and error message to Google Gemini AI
2. Gets a corrected version of the PlantUML code
3. Re-validates the corrected code
4. Returns the URL for the corrected diagram

This ensures that even malformed PlantUML input produces valid, renderable diagrams.
