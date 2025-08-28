# Imagen Edit Tool

The Imagen Edit tool enables image editing using Google's Gemini 2.0 Flash model with image generation capabilities via Vertex AI. It accepts base64 encoded images and text instructions to perform various image editing operations.

## Features

- **Base64 Image Input**: Accepts images as base64 encoded strings instead of file URIs
- **Text-Guided Editing**: Uses natural language instructions to modify images
- **Multiple Operations**: Add objects, modify elements, change colors, remove items, apply effects
- **High-Quality Output**: Powered by Gemini 2.0 Flash with image generation
- **HTTP Serving**: Automatically serves generated images via HTTP
- **Configurable Parameters**: Control generation temperature, top-p, and token limits

## Configuration

Set these environment variables:

- `GCP_PROJECT` (required): Google Cloud Project ID
- `GCP_REGION` (default: "global"): Google Cloud Region
- `IMAGEN_EDIT_TOOL_DIR` (default: "./images_edit"): Directory to save edited images
- `IMAGEN_EDIT_TOOL_PORT` (default: 8081): HTTP server port for serving images

## Authentication

Requires Google Cloud authentication. Use one of:

```bash
# Application Default Credentials
gcloud auth application-default login

# Or set service account key
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

## MCP Tool Usage

The tool provides one main MCP tool:

### `imagen_edit`

Edit images using text instructions with base64 encoded image input.

**Required Parameters:**
- `base64_image`: Base64 encoded image data (without data:image/... prefix)
- `mime_type`: MIME type of the image (e.g., "image/jpeg", "image/png")
- `edit_instruction`: Text describing the edit to perform

**Optional Parameters:**
- `temperature`: Randomness in generation (0.0-2.0, default: 1.0)
- `top_p`: Nucleus sampling parameter (0.0-1.0, default: 0.95)
- `max_output_tokens`: Maximum tokens in response (1-8192, default: 8192)

**Examples:**

```json
{
  "base64_image": "iVBORw0KGgoAAAANSUhEUgAA...",
  "mime_type": "image/jpeg",
  "edit_instruction": "Add chocolate drizzle to the croissants"
}
```

```json
{
  "base64_image": "iVBORw0KGgoAAAANSUhEUgAA...",
  "mime_type": "image/png", 
  "edit_instruction": "Change the car color to blue",
  "temperature": 0.8
}
```

```json
{
  "base64_image": "iVBORw0KGgoAAAANSUhEUgAA...",
  "mime_type": "image/jpeg",
  "edit_instruction": "Remove the person from the background"
}
```

## Supported Edit Operations

- **Add Elements**: "Add flowers to the vase", "Add a hat to the person"
- **Modify Objects**: "Change the car color to red", "Make the building taller"
- **Remove Items**: "Remove the chair from the room", "Delete the watermark"
- **Apply Effects**: "Add dramatic lighting", "Make it look vintage"
- **Style Changes**: "Convert to black and white", "Make it look like a painting"
- **Texture/Material**: "Change wood to metal", "Add rust to the surface"

## Building and Running

### Build
```bash
# Build all tools including imagen_edit
make all

# Build only imagen_edit
make bin/imagen_edit
```

### Run as MCP Server
```bash
# Run directly
./bin/imagen_edit

# Or via make (for development)
make run TOOL=imagen_edit
```

## Image Serving

The tool automatically starts an HTTP server to serve edited images:

- **Default Port**: 8081
- **Image URL Format**: `http://localhost:8081/images/{filename}`
- **Health Check**: `http://localhost:8081/health`

## Example Integration

```python
# Example using the MCP tool with a base64 encoded image
import base64

# Read and encode image
with open("input.jpg", "rb") as f:
    image_data = base64.b64encode(f.read()).decode('utf-8')

# MCP tool call
result = mcp_client.call_tool("imagen_edit", {
    "base64_image": image_data,
    "mime_type": "image/jpeg",
    "edit_instruction": "Add sunglasses to the person",
    "temperature": 0.9
})
```

## Error Handling

The tool validates:
- Base64 image format and decodability
- MIME type validity (image/jpeg, image/png, image/gif, image/webp)
- Parameter ranges (temperature, top_p, max_output_tokens)
- Google Cloud authentication and project access

## Output Format

Returns detailed information about the edited image:

```
Successfully edited image with instruction: "Add chocolate drizzle to the croissants"

AI Response: I've added elegant chocolate drizzle across the croissants...

Edited Image:
  URL: http://localhost:8081/images/edit_a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6.png
  File: ./images_edit/edit_a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6.png
  Size: 245760 bytes
  Format: PNG
```

## Notes

- Uses Gemini 2.0 Flash Preview model specifically designed for image generation
- Images are saved as PNG format for consistency
- Unique filenames prevent conflicts using UUID
- Safety settings are configured to allow creative content generation
- Supports multimodal responses (text + image)