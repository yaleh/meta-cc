# Metadata-Driven Query Architecture

## Overview

The Metadata-Driven Query Architecture is a new approach to querying session data that provides lightweight metadata access with minimal data transfer overhead. Instead of returning large result sets, this architecture returns compact metadata that enables Claude to construct efficient queries.

## Components

### 1. get_session_metadata Tool

The core component of this architecture is the `get_session_metadata` tool which provides:

- **JSONL Schema Information**: Structure and field definitions for session data
- **File Metadata**: Details about session files including paths, sizes, timestamps, and record counts
- **Query Templates**: Pre-defined jq filter templates for common query patterns

#### Usage

```json
{
  "method": "tools/call",
  "params": {
    "name": "get_session_metadata",
    "arguments": {
      "scope": "project"
    }
  }
}
```

#### Response Structure

```json
{
  "base_dir": "/path/to/session/files",
  "file_count": 770,
  "files": [
    {
      "path": "/path/to/session/file.jsonl",
      "size_bytes": 1515670,
      "modified_at": "2025-10-30T04:06:04Z",
      "records": 0
    }
  ],
  "jsonl_schema": {
    "common_fields": [...],
    "user_message_fields": [...],
    "assistant_message_fields": [...],
    "tool_fields": [...]
  },
  "query_templates": {
    "user_messages": {...},
    "tool_errors": {...},
    "time_range": {...},
    "smart_file_filter": {...}
  }
}
```

### 2. Query Template Library

YAML-based template definitions that provide reusable jq filter patterns:

- **user_messages.yaml**: Filter for user messages
- **tool_errors.yaml**: Filter for tool execution errors
- **time_range.yaml**: Filter by timestamp range
- **smart_file_filter.yaml**: Smart file filtering based on metadata

### 3. Shell Helper Tools

Lightweight shell scripts that can be used to execute queries:

- **query_helper.sh**: Generic helper for all query tools
- **query_user_messages.sh**: Specific helper for user messages
- **query_tool_errors.sh**: Specific helper for tool errors

## Benefits

### Zero-Code Extensibility
Claude can create new queries without requiring code changes by combining metadata with jq expressions.

### Minimal Data Transfer
Returns 1-5 KB of metadata instead of 10-1000 KB of actual results, reducing network overhead.

### Complete Transparency
Users see full query commands, making debugging straightforward.

### Extreme Maintainability
Less than 100 lines of core code versus 2000+ lines in the previous implementation.

### Maximum Flexibility
Claude can freely combine jq, grep, sort, and other command-line tools.

## Example Workflow

```
User: "Show me errors from yesterday mentioning 'timeout'"

Claude:
1. Calls get_session_metadata(scope: "project")
2. Filters files by timestamp (yesterday's files only)
3. Constructs optimized query:
   cat file1.jsonl file2.jsonl \
     | grep '"is_error":true' \
     | jq -c 'select(.type == "user") |
              select(.message.content[]? |
              select(.type == "tool_result" and .is_error)) |
              select(.message.content | test("timeout"))' \
     | jq -s 'sort_by(.timestamp) | reverse'
4. Executes via Bash tool
```

## Security Considerations

- **Path Sandboxing**: Restricts access to session directories only
- **Metadata Exclusion**: Sensitive content is excluded from metadata
- **Parameterization**: Query templates use parameterization to prevent injection
- **Audit Trail**: Full audit trail through MCP and Bash tool logs

## Backward Compatibility

Existing query tools remain functional to ensure backward compatibility while providing the new metadata-driven approach as an alternative.
