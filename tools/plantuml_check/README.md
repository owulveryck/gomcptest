# PlantUML Check Tool

An MCP (Model Context Protocol) server that validates PlantUML file syntax using the PlantUML jar file.

## Features

- Validates PlantUML syntax using the official PlantUML processor
- Supports common PlantUML file extensions (.puml, .plantuml, .pu)
- Provides detailed error messages for syntax issues
- Uses environment variables for configuration

## Requirements

- Java must be installed and available in PATH
- PlantUML jar file (download from https://plantuml.com/download)
- PLANTUML_JAR environment variable pointing to the jar file

## Environment Variables

- `PLANTUML_JAR`: Path to the PlantUML jar file (required)

## Usage

1. Set the environment variable:
   ```bash
   export PLANTUML_JAR=/path/to/plantuml-mit-1.2024.7.jar
   ```

2. Run the MCP server:
   ```bash
   ./cmd/plantuml_check
   ```

3. The tool provides a `plantuml_check` function that accepts:
   - `file_path`: Absolute path to the PlantUML file to validate

## Example

```json
{
  "tool": "plantuml_check",
  "arguments": {
    "file_path": "/absolute/path/to/diagram.puml"
  }
}
```

## Output

- ✅ Success: Returns confirmation that syntax is valid
- ❌ Error: Returns detailed syntax error information

## Building

```bash
go build -o cmd/plantuml_check cmd/main.go
```

## Testing

The tool includes test files:
- `test.puml`: Valid PlantUML diagram
- `test_invalid.puml`: Invalid PlantUML diagram (missing @enduml)

## Security Notes

- Only validates syntax, does not execute or generate diagrams
- Requires absolute file paths for security
- Validates file extensions before processing