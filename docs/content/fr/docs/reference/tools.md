---
title: "Tools Reference"
linkTitle: "Tools"
weight: 1
description: >
  Comprehensive reference of all available MCP-compatible tools
---

{{% pageinfo %}}
This reference guide documents all available MCP-compatible tools in the gomcptest project, their parameters, and response formats.
{{% /pageinfo %}}

## Bash

Executes bash commands in a persistent shell session.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `command` | string | Yes | The command to execute |
| `timeout` | number | No | Timeout in milliseconds (max 600000) |

### Response

The tool returns the command output as a string.

### Banned Commands

For security reasons, the following commands are banned:
`alias`, `curl`, `curlie`, `wget`, `axel`, `aria2c`, `nc`, `telnet`, `lynx`, `w3m`, `links`, `httpie`, `xh`, `http-prompt`, `chrome`, `firefox`, `safari`

## Edit

Modifies file content by replacing specified text.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_path` | string | Yes | Absolute path to the file to modify |
| `old_string` | string | Yes | Text to replace |
| `new_string` | string | Yes | Replacement text |

### Response

Confirmation message with the updated content.

## GlobTool

Finds files matching glob patterns with metadata.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `pattern` | string | Yes | Glob pattern to match files against |
| `path` | string | No | Directory to search in (default: current directory) |
| `exclude` | string | No | Glob pattern to exclude from results |
| `limit` | number | No | Maximum number of results to return |
| `absolute` | boolean | No | Return absolute paths instead of relative |

### Response

A list of matching files with metadata including path, size, modification time, and permissions.

## GrepTool

Searches file contents using regular expressions.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `pattern` | string | Yes | Regular expression pattern to search for |
| `path` | string | No | Directory to search in (default: current directory) |
| `include` | string | No | File pattern to include in the search |

### Response

A list of matches with file paths, line numbers, and matched content.

## LS

Lists files and directories in a given path.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | Yes | Absolute path to the directory to list |
| `ignore` | array | No | List of glob patterns to ignore |

### Response

A list of files and directories with metadata.

## Replace

Completely replaces a file's contents.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_path` | string | Yes | Absolute path to the file to write |
| `content` | string | Yes | Content to write to the file |

### Response

Confirmation message with the content written.

## View

Reads file contents with optional line range.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_path` | string | Yes | Absolute path to the file to read |
| `offset` | number | No | Line number to start reading from |
| `limit` | number | No | Number of lines to read |

### Response

The file content with line numbers in cat -n format.

## dispatch_agent

Launches a new agent with access to specific tools.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `prompt` | string | Yes | The task for the agent to perform |

### Response

The result of the agent's task execution.

## imagen

Génère et manipule des images en utilisant l'API Google Imagen.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `prompt` | string | Yes | Description de l'image à générer |
| `aspectRatio` | string | No | Ratio d'aspect pour l'image (par défaut: "1:1") |
| `safetyFilterLevel` | string | No | Niveau de filtre de sécurité (par défaut: "block_some") |
| `personGeneration` | string | No | Politique de génération de personnes (par défaut: "dont_allow") |

### Response

Retourne un objet JSON avec le chemin de l'image générée et les métadonnées.

## duckdbserver

Fournit des capacités de traitement de données en utilisant DuckDB.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Requête SQL à exécuter |
| `database` | string | No | Chemin du fichier de base de données (par défaut: en mémoire) |

### Response

Résultats de la requête au format JSON.

## Tool Response Format

Most tools return JSON responses with the following structure:

```json
{
  "result": "...", // String result or
  "results": [...], // Array of results
  "error": "..." // Error message if applicable
}
```

## Error Handling

All tools follow a consistent error reporting format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

Common error codes include:
- `INVALID_PARAMS`: Parameters are missing or invalid
- `EXECUTION_ERROR`: Error executing the requested operation
- `PERMISSION_DENIED`: Permission issues
- `TIMEOUT`: Operation timed out