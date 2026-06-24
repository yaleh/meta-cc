# Fixture: content-user-pattern

## Fixture Metadata

```yaml
id: content-user-pattern
tool: query_session_content
decision_class: B
oracle: haiku
```

## Input Parameters

```json
{
  "role": "user",
  "pattern": "fix|bug|error",
  "scope": "project",
  "provider": "claude"
}
```

## Expected Behavior

The tool queries user messages and applies a regex filter `fix|bug|error` to `message.content`.
Only records where the content matches the regex are returned. The pattern is applied via jq's
`test()` function, which uses POSIX Extended Regular Expressions (ERE).

## Required Output Fields

Each returned record MUST contain:

| Field | Type | Value Constraint |
|-------|------|-----------------|
| `type` | string | exactly `"user"` |
| `timestamp` | string | ISO8601 format |
| `sessionId` | string | non-empty UUID |
| `message.role` | string | exactly `"user"` |
| `message.content` | string | matches pattern `fix\|bug\|error` (case-sensitive) |

## Pass/Fail Criteria

### Assertion UP1 (Class A — Field Presence)

**Oracle question**: Does every record in the output contain `timestamp`, `sessionId`,
`message.role: "user"`, and a string `message.content`?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion UP2 (Class B — Pattern Match)

**Oracle question**: Does the `message.content` of every returned record contain at least one of
the words "fix", "bug", or "error" (case-sensitive)?

**Pass**: Oracle answers YES with confidence ≥ 0.7

### Assertion UP3 (Class B — No False Positives)

**Oracle question**: Are there any records in the output where the `message.content` does NOT
contain "fix", "bug", or "error"?

**Pass**: Oracle answers NO (no records lack the pattern)

### Assertion UP4 (Class C — Valid JSON)

**Programmatic check**: Response is valid JSON or JSONL.

**Pass**: JSON parse succeeds

## Example Expected Output Shape

```json
{
  "type": "user",
  "timestamp": "2025-10-24T14:09:22.100Z",
  "sessionId": "abc123-...",
  "message": {
    "role": "user",
    "content": "There's a bug in the authentication module, can you fix it?"
  }
}
```

## Notes on Pattern Semantics

- Pattern is case-sensitive by default (jq's `test()` is case-sensitive unless `"i"` flag is used).
- The meta-cc implementation does NOT add a case-insensitive flag for user messages; `role=assistant`
  with `contains` DOES use case-insensitive matching.
- Alternation `fix|bug|error` in the pattern is standard POSIX ERE.
- If pattern contains special regex characters, they must be properly escaped.

## Edge Cases

- Empty result set is valid if no user messages match the pattern.
- Very long messages are not truncated unless `max_message_length` is specified.
