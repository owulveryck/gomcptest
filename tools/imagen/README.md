# Imagen Tool Suite

A comprehensive Model Context Protocol (MCP) tool suite for generating images using Google Cloud Vertex AI Imagen API. Provides three specialized tools for different quality/speed trade-offs with integrated HTTP server for serving generated images.

## Features

- **Vertex AI Integration** - Uses Google Cloud Vertex AI backend like openaiserver
- **Three specialized tools** for different quality/speed trade-offs
- **HTTP Image Server** - Automatic web server to serve generated images
- **URL-based Output** - Returns HTTP URLs instead of file paths
- **Complete API coverage** with all supported Imagen parameters
- **Consistent configuration** using envconfig like other gomcptest tools
- **Person generation controls** for ethical AI use
- **Multiple aspect ratios** and image resolutions
- **Batch generation** (1-4 images per request)
- **Automatic image saving** with organized naming
- **MCP compatible** for seamless AI agent integration

## Available Tools

### 1. `imagen_generate_standard`
High-quality image generation using the standard Imagen 4.0 model.
- **Model**: `imagen-4.0-generate-001`
- **Quality**: High
- **Speed**: Standard
- **Resolution**: 1K or 2K options
- **Use case**: Balanced quality and speed

### 2. `imagen_generate_ultra`
Ultra high-quality image generation with enhanced detail and photorealism.
- **Model**: `imagen-4.0-ultra-generate-001`
- **Quality**: Ultra High
- **Speed**: Slower (highest quality)
- **Resolution**: 1K or 2K options
- **Use case**: Professional/marketing content

### 3. `imagen_generate_fast`
Fast image generation optimized for speed and rapid iteration.
- **Model**: `imagen-4.0-fast-generate-001`
- **Quality**: Good
- **Speed**: Fast
- **Resolution**: Fixed size (no size options)
- **Use case**: Rapid prototyping and concepts

## Configuration

Set these environment variables:

### Required
- `GCP_PROJECT`: Google Cloud Project ID

### Optional (with defaults)
- `GCP_REGION`: Google Cloud Region (default: us-central1)
- `IMAGEN_TOOL_DIR`: Directory to save images (default: ./images)
- `IMAGEN_TOOL_PORT`: HTTP server port for serving images (default: 8080)
- `LOG_LEVEL`: Logging level (default: INFO)

## HTTP Image Server

The tool automatically starts an HTTP server to serve generated images:

- **Endpoint**: `http://localhost:{IMAGEN_TOOL_PORT}/images/{filename}`
- **Health Check**: `http://localhost:{IMAGEN_TOOL_PORT}/health`
- **Content-Type**: Automatically detected based on file extension
- **Security**: Directory traversal protection
- **Format**: Generated images are served as PNG files

### Example URLs
```
http://localhost:8080/images/imagen_std_a1b2c3d4-e5f6-7890-abcd-ef1234567890.png
http://localhost:8080/images/imagen_ultra_f6e5d4c3-b2a1-9876-5432-10fedcba9876.png
http://localhost:8080/images/imagen_fast_9a8b7c6d-5e4f-3210-9876-543210fedcba.png
```

## Authentication

Uses Google Cloud authentication like other gomcptest tools:

**Option 1: Application Default Credentials**
```bash
gcloud auth application-default login
```

**Option 2: Service Account Key**
```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

**Option 3: GCP Environment** (GKE, Cloud Run, etc.)
- Automatically uses attached service account

## Supported Parameters

| Parameter | Type | Description | Standard | Ultra | Fast |
|-----------|------|-------------|----------|-------|------|
| `prompt` | string | Text description (max 480 tokens) | ✅ | ✅ | ✅ |
| `number_of_images` | number | Number of images (1-4, default: 1) | ✅ | ✅ | ✅ |
| `sample_image_size` | string | Resolution: "1K" or "2K" (default: "1K") | ✅ | ✅ | ❌ |
| `aspect_ratio` | string | Ratio: "1:1", "3:4", "4:3", "9:16", "16:9" (default: "1:1") | ✅ | ✅ | ✅ |
| `person_generation` | string | Policy: "dont_allow", "allow_adult", "allow_all" (default: "allow_adult") | ✅ | ✅ | ✅ |

## Usage Examples

### Standard Quality Generation
```json
{
  "tool": "imagen_generate_standard",
  "arguments": {
    "prompt": "A serene mountain landscape at sunset",
    "number_of_images": 2,
    "aspect_ratio": "16:9"
  }
}
```

**Response:**
```
Generated 2 image(s) using imagen-4.0-generate-001 via Vertex AI for prompt: "A serene mountain landscape at sunset"

Image 1:
  URL: http://localhost:8080/images/imagen_std_a1b2c3d4-e5f6-7890-abcd-ef1234567890.png
  File: ./images/imagen_std_a1b2c3d4-e5f6-7890-abcd-ef1234567890.png
  Size: 1024576 bytes
  Format: PNG

Image 2:
  URL: http://localhost:8080/images/imagen_std_f6e5d4c3-b2a1-9876-5432-10fedcba9876.png
  File: ./images/imagen_std_f6e5d4c3-b2a1-9876-5432-10fedcba9876.png
  Size: 987432 bytes
  Format: PNG
```

### Ultra High-Quality Portrait
```json
{
  "tool": "imagen_generate_ultra",
  "arguments": {
    "prompt": "Professional headshot of a business executive",
    "sample_image_size": "2K",
    "aspect_ratio": "3:4",
    "person_generation": "allow_adult"
  }
}
```

### Fast Concept Iteration
```json
{
  "tool": "imagen_generate_fast",
  "arguments": {
    "prompt": "Logo design concepts for a tech startup",
    "number_of_images": 4,
    "aspect_ratio": "1:1"
  }
}
```

## Building and Running

### Build (via Makefile)
```bash
cd tools
make build
```

### Run as MCP Server with Web Server
```bash
# Set environment variables
export GCP_PROJECT="your-project-id"
export GCP_REGION="us-central1"
export IMAGEN_TOOL_DIR="./images"
export IMAGEN_TOOL_PORT="8080"

# Run the tool (starts both MCP and HTTP servers)
./bin/imagen
```

### Test Tool Availability
```bash
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | ./bin/imagen
```

### Test Web Server
```bash
# Health check
curl http://localhost:8080/health

# Access generated image (after generation)
curl http://localhost:8080/images/imagen_std_your-image-id.png
```

## Integration with gomcptest

This tool follows the same patterns as other gomcptest tools while adding HTTP serving capabilities:

- **Configuration**: Uses envconfig like openaiserver
- **Authentication**: Google Cloud credentials like cliGCP  
- **Build System**: Integrated with tools/Makefile
- **Client Pattern**: Vertex AI backend like openaiserver
- **Error Handling**: Consistent error patterns
- **Testing**: Unit tests with environment cleanup
- **HTTP Server**: Same image serving pattern as openaiserver

## Output Structure

### Generated Image URLs
Images are served via HTTP with unique filenames:
- **Format**: `imagen_{model}_{uuid}.png`
- **Examples**:
  - `imagen_std_a1b2c3d4-e5f6-7890-abcd-ef1234567890.png` (standard)
  - `imagen_ultra_f6e5d4c3-b2a1-9876-5432-10fedcba9876.png` (ultra)
  - `imagen_fast_9a8b7c6d-5e4f-3210-9876-543210fedcba.png` (fast)

### Tool Response Format
```
Generated N image(s) using {model} via Vertex AI for prompt: "{prompt}"

Image 1:
  URL: http://localhost:{port}/images/{filename}
  File: {local_path}
  Size: {bytes} bytes
  Format: PNG
```

## Web Server Features

### Security
- **Directory traversal protection**: Prevents access to files outside image directory
- **Method restriction**: Only GET requests allowed
- **File validation**: Ensures requested files exist and are not directories
- **MIME type detection**: Proper content-type headers

### Performance
- **Direct file serving**: Efficient `io.Copy` streaming
- **Proper error handling**: HTTP status codes for different error conditions
- **Minimal overhead**: Lightweight HTTP server running in background

### Endpoints
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/images/{filename}` | GET | Serve generated image files |
| `/health` | GET | Health check (returns "OK") |

## Error Handling

The tool provides comprehensive error handling for:
- Missing or invalid Google Cloud credentials
- Invalid or missing GCP_PROJECT configuration
- HTTP server startup failures
- Image file access errors
- Invalid parameters and values
- Vertex AI API errors and rate limits
- Network connectivity issues
- File system errors
- Prompt length validation (480 token limit)

## Environment Variable Setup

```bash
# Required
export GCP_PROJECT="your-gcp-project-id"

# Optional (with defaults)
export GCP_REGION="us-central1"
export IMAGEN_TOOL_DIR="./images"
export IMAGEN_TOOL_PORT="8080"
export LOG_LEVEL="INFO"
```

## Troubleshooting

### Common Issues

1. **Missing GCP_PROJECT**: Ensure `GCP_PROJECT` environment variable is set
2. **Authentication**: Run `gcloud auth application-default login`
3. **Port conflicts**: Change `IMAGEN_TOOL_PORT` if 8080 is in use
4. **File permissions**: Ensure write access to `IMAGEN_TOOL_DIR`
5. **Image access**: Verify web server is running on configured port

### Debug Mode
```bash
LOG_LEVEL=DEBUG ./bin/imagen
```

### Test Web Server
```bash
# Check if server is running
curl -s http://localhost:8080/health

# Test with custom port
IMAGEN_TOOL_PORT=9090 ./bin/imagen &
curl -s http://localhost:9090/health
```

## Differences from AI Studio API

This tool uses **Vertex AI** (not AI Studio) for enterprise-grade features:
- **Enhanced Security**: VPC controls and enterprise authentication
- **Better Scaling**: Higher quotas and rate limits
- **Compliance**: GDPR/SOC compliance for enterprise use
- **Integration**: Native GCP service integration
- **Monitoring**: Cloud Logging and Monitoring integration

The integrated HTTP server provides immediate web access to generated images, making it ideal for web applications and AI agents that need direct image URLs.