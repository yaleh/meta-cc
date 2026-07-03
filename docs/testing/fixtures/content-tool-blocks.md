# Fixture: content-tool-blocks

## Fixture Metadata

```yaml
id: content-tool-blocks
tool: query_session_content
decision_class: A
oracle: haiku
```

## Input Parameters

```json
{
  "role": "tool",
  "block_type": "tool_use",
  "scope": "project",
  "provider": "claude"
}
```

## Expected Behavior

The tool extracts individual `tool_use` blocks from assistant records. Unlike other query types,
this tool explodes one assistant record into multiple output records (one per tool_use block).
Each output record is a flattened object combining the outer record's context fields (`timestamp`,
`sessionId`, `turn`) with the tool_use block fields.

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"tool_use"` |
| `id` | string | non-empty tool call ID |
| `name` | string | non-empty tool name |
| `input` | object | tool input parameters |
| `timestamp` | string | ISO8601 format (from outer record) |
| `sessionId` | string | non-empty UUID (from outer record) |
| `turn` | number | integer ≥ 0 (from outer record) |

## Pass/Fail Criteria

### Assertion TB1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `type: "tool_use"`, `id`, `name`,
`input`, `timestamp`, `sessionId`, and `turn`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion TB2 (Class A — Context Fields Injected)

**Oracle question**: Does every returned tool_use record contain `timestamp` and `sessionId` fields
that appear to come from the outer session context (not from the tool_use block itself)?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion TB3 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "tool_use",
  "id": "toolu_abc123",
  "name": "Read",
  "input": {
    "file_path": "/home/user/project/main.go"
  },
  "timestamp": "2025-10-24T14:07:38.500Z",
  "sessionId": "abc123-...",
  "turn": 5
}
```

## Variant: tool_result blocks

### Input Parameters (variant)

```json
{
  "role": "tool",
  "block_type": "tool_result",
  "scope": "project",
  "provider": "claude"
}
```

### Required Output Fields (tool_result)

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"tool_result"` |
| `tool_use_id` | string | non-empty, matches a tool_use ID |
| `content` | string or array | result content |
| `timestamp` | string | ISO8601 format (from outer record) |
| `sessionId` | string | non-empty UUID (from outer record) |
| `turn` | number | integer ≥ 0 (from outer record) |

### Assertion TB4 (Class A — tool_result fields)

**Oracle question**: Does every tool_result record contain `type: "tool_result"`, `tool_use_id`,
`content`, `timestamp`, `sessionId`, and `turn`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

## Edge Cases

- An assistant record with N tool_use blocks produces N output records.
- The `is_error` field may be present on `tool_result` blocks (boolean).
- `input` field shape varies per tool; only its presence (as an object) is required.
