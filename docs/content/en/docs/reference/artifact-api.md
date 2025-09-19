---
title: "Artifact Storage API Reference"
linkTitle: "Artifact API"
weight: 5
description: >-
  Complete API reference for the artifact storage endpoints in the OpenAI server
---

The OpenAI server provides a RESTful API for storing and retrieving generic artifacts (files). This API allows you to upload any type of file and retrieve it later using a unique identifier.

## Authentication

The artifact API endpoints do not require authentication and are publicly accessible. In production environments, consider implementing authentication middleware as needed.

## Content Types

The API supports any content type. Common examples include:
- `text/plain` - Text files
- `application/json` - JSON documents
- `image/jpeg`, `image/png` - Images
- `application/pdf` - PDF documents
- `audio/webm`, `audio/wav` - Audio files
- `application/octet-stream` - Binary files

## Endpoints

### Upload Artifact

Uploads a new artifact to the server.

**Request:**
```
POST /artifact/
```

**Headers:**
- `Content-Type` (required): MIME type of the file being uploaded
- `X-Original-Filename` (required): Original filename including extension

**Request Body:**
- Binary file data

**Response:**

Success (201 Created):
```json
{
  "artifactId": "7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf"
}
```

**Response Headers:**
- `Location`: URL where the artifact can be retrieved
- `Content-Type`: `application/json`

**Error Responses:**

400 Bad Request:
```
Missing 'Content-Type' or 'X-Original-Filename' header
```

413 Payload Too Large:
```
Error saving file: http: request body too large
```

500 Internal Server Error:
```
Could not create file on server
```

**Example:**
```bash
curl -X POST http://localhost:8080/artifact/ \
  -H "Content-Type: text/plain" \
  -H "X-Original-Filename: example.txt" \
  --data-binary @example.txt
```

### Retrieve Artifact

Downloads an artifact by its unique identifier.

**Request:**
```
GET /artifact/{artifactId}
```

**Path Parameters:**
- `artifactId` (required): UUID of the artifact to retrieve

**Response:**

Success (200 OK):
- Returns the original file content as binary data

**Response Headers:**
- `Content-Type`: Original MIME type of the file
- `Content-Disposition`: `inline; filename="original-filename.ext"`
- `Content-Length`: Size of the file in bytes
- `Accept-Ranges`: `bytes` (supports range requests)
- `Last-Modified`: Timestamp when the file was uploaded

**Error Responses:**

400 Bad Request:
```
Invalid artifact ID format
```

404 Not Found:
```
404 page not found
```

500 Internal Server Error:
```
Could not read artifact metadata
Corrupted artifact metadata
```

**Example:**
```bash
curl http://localhost:8080/artifact/7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf
```

## Data Models

### Artifact Metadata

Each uploaded artifact has associated metadata stored in a `.meta.json` file:

```json
{
  "originalFilename": "example.txt",
  "contentType": "text/plain",
  "size": 1024,
  "uploadTimestamp": "2025-09-19T12:01:11.277651Z"
}
```

**Fields:**
- `originalFilename` (string): The original name of the uploaded file
- `contentType` (string): MIME type of the file
- `size` (number): Size of the file in bytes
- `uploadTimestamp` (string): ISO 8601 timestamp of when the file was uploaded

## Configuration

The artifact storage behavior can be configured using environment variables:

- `ARTIFACT_PATH`: Directory where artifacts are stored (default: `~/openaiserver/artifacts`)
- `MAX_UPLOAD_SIZE`: Maximum file size in bytes (default: `52428800` = 50MB)

## File Storage

### Storage Structure

Artifacts are stored using the following directory structure:

```
${ARTIFACT_PATH}/
├── 7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf           # Binary file content
├── 7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf.meta.json # Metadata file
├── 123e4567-e89b-12d3-a456-426614174000           # Another file
└── 123e4567-e89b-12d3-a456-426614174000.meta.json # Its metadata
```

### File Naming

- **Artifact files**: Named using the UUID (no extension)
- **Metadata files**: Named using the UUID + `.meta.json` suffix
- **UUIDs**: Generated using UUID v4 standard (RFC 4122)

## Security Considerations

### File Size Limits

- Default maximum upload size: 50MB
- Configurable via `MAX_UPLOAD_SIZE` environment variable
- Requests exceeding the limit return HTTP 413 (Payload Too Large)

### File Type Validation

- The API accepts any content type
- Content type validation is based on the `Content-Type` header
- No server-side file content inspection is performed

### Path Traversal Protection

- Artifact IDs must be valid UUIDs
- Invalid UUID format returns HTTP 400 (Bad Request)
- File paths are constructed using secure `filepath.Join()`

### Storage Directory

- Default storage path uses user home directory
- Tilde (`~`) expansion is supported
- Directory is created automatically with 0755 permissions
- Metadata files are created with 0644 permissions

## Error Handling

### Client Errors (4xx)

- **400 Bad Request**: Invalid UUID format or missing required headers
- **404 Not Found**: Artifact with the specified ID does not exist
- **413 Payload Too Large**: File exceeds maximum upload size

### Server Errors (5xx)

- **500 Internal Server Error**: File system errors, metadata corruption, or server misconfiguration

### Logging

All artifact operations are logged with structured logging:

```
INFO Artifact uploaded successfully artifactID=7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf filename=example.txt size=1024
DEBUG Artifact served artifactID=7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf filename=example.txt
ERROR Could not create artifact file error="permission denied" path=/artifacts/uuid
```

## Rate Limiting

The artifact API does not implement built-in rate limiting. For production environments, consider implementing rate limiting at the reverse proxy level or using middleware.

## CORS Support

The artifact API includes CORS headers to support web-based clients:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With
Access-Control-Allow-Credentials: true
```

## Integration Examples

### JavaScript/Fetch API

```javascript
// Upload file
const uploadFile = async (file) => {
  const response = await fetch('/artifact/', {
    method: 'POST',
    headers: {
      'Content-Type': file.type,
      'X-Original-Filename': file.name
    },
    body: file
  });

  const result = await response.json();
  return result.artifactId;
};

// Download file
const downloadFile = async (artifactId) => {
  const response = await fetch(`/artifact/${artifactId}`);
  return response.blob();
};
```

### Python/Requests

```python
import requests

# Upload file
def upload_file(file_path, content_type):
    with open(file_path, 'rb') as f:
        headers = {
            'Content-Type': content_type,
            'X-Original-Filename': os.path.basename(file_path)
        }
        response = requests.post('http://localhost:8080/artifact/',
                               headers=headers, data=f)
        return response.json()['artifactId']

# Download file
def download_file(artifact_id, output_path):
    response = requests.get(f'http://localhost:8080/artifact/{artifact_id}')
    with open(output_path, 'wb') as f:
        f.write(response.content)
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

// Upload file
func uploadFile(filePath, contentType string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    req, err := http.NewRequest("POST", "http://localhost:8080/artifact/", file)
    if err != nil {
        return "", err
    }

    req.Header.Set("Content-Type", contentType)
    req.Header.Set("X-Original-Filename", filepath.Base(filePath))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]string
    json.NewDecoder(resp.Body).Decode(&result)
    return result["artifactId"], nil
}

// Download file
func downloadFile(artifactID, outputPath string) error {
    resp, err := http.Get(fmt.Sprintf("http://localhost:8080/artifact/%s", artifactID))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.Copy(file, resp.Body)
    return err
}
```