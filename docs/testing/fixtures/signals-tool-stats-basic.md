# Fixture: signals-tool-stats-basic

## Fixture Metadata

```yaml
id: signals-tool-stats-basic
tool: query_session_signals
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "type": "tool_stats",
  "scope": "project",
  "provider": "claude"
}
```

## Expected Behavior

The tool queries JSONL records for assistant entries that contain `tool_use` blocks in their
`message.content` array. Each returned record is a full JSONL assistant record that triggered
at least one tool call.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"assistant"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |
| `message.content` | array | length ≥ 1 |
| `message.content[].type` | string | at least one element is `"tool_use"` |
| `message.content[].name` | string | non-empty tool name |
| `message.content[].id` | string | non-empty tool call ID |

## Pass/Fail Criteria

### Assertion TS1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `timestamp`, `sessionId`, `type: "assistant"`,
and at least one `message.content` element with `type: "tool_use"`, `name`, and `id`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion TS2 (Class A — Tool Name Non-empty)

**Oracle question**: Is the `name` field in every `tool_use` block a non-empty string?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion TS3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "assistant",
  "timestamp": "2025-10-24T14:07:38.500Z",
  "sessionId": "abc123-...",
  "message": {
    "role": "assistant",
    "content": [
      {
        "type": "text",
        "text": "I'll read that file for you."
      },
      {
        "type": "tool_use",
        "id": "toolu_abc123",
        "name": "Read",
        "input": {
          "file_path": "/path/to/file.go"
        }
      }
    ]
  }
}
```

## Variant: Filtered by Tool Name

### Input Parameters (variant)

```json
{
  "type": "tool_stats",
  "tool": "Bash",
  "scope": "project",
  "provider": "claude"
}
```

### Additional Assertion TS4 (Class B — Filter Semantics)

**Oracle question**: In the filtered output, does every record contain at least one `tool_use` block
where `name` equals exactly `"Bash"`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

## Edge Cases

- Records with multiple tool_use blocks in one turn are still one returned record.
- Tool `input` field content varies by tool — only `type`, `name`, and `id` are required in assertions.
