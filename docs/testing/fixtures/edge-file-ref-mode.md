# Fixture: edge-file-ref-mode

## Fixture Metadata

```yaml
id: edge-file-ref-mode
tool: query_session_signals, query_session_content
decision_class: A
oracle: haiku
```

## Purpose

Verifies correct behavior when query results exceed the inline threshold (8KB by default),
triggering hybrid output mode where the tool returns a `file_ref` instead of inline JSON.

---

## Case 1: Force file_ref with inline_threshold_bytes=1

### Input Parameters

```json
{
  "type": "timestamps",
  "scope": "project",
  "provider": "claude",
  "inline_threshold_bytes": 1
}
```

### Expected Behavior

Setting `inline_threshold_bytes: 1` forces all non-trivial responses into `file_ref` mode.
The response should contain a `file_ref` field pointing to a readable temp file, not inline JSON records.

### Pass/Fail Criteria

#### Assertion FR1 (Class A — file_ref Structure)

**Oracle question**: Does the output contain a `file_ref` field with a non-empty file path string?

**Pass**: Oracle answers YES with confidence ≥ 0.7

#### Assertion FR2 (Class A — file_ref Readable)

**Oracle question**: Does the output contain metadata about the referenced file, such as
`record_count` or `bytes` fields indicating the size of the referenced data?

**Pass**: Oracle answers YES with confidence ≥ 0.7

#### Assertion FR3 (Class C — file_ref Path Exists)

**Programmatic check**: The file path in `file_ref` points to a readable file on the filesystem.

**Pass**: `test -r <file_ref_path>` succeeds

---

## Case 2: Large project query (natural file_ref)

### Input Parameters

```json
{
  "role": "all",
  "scope": "project",
  "provider": "claude"
}
```

Note: On a large project with many sessions, this query naturally produces >8KB of output.

### Expected Behavior

For projects with significant history, the tool automatically switches to file_ref mode.
For small projects or new sessions, it may return inline JSON — both are valid.

### Pass/Fail Criteria

#### Assertion FR4 (Class A — Either Mode Valid)

**Oracle question**: Does the output either contain well-formed inline JSONL records (type/timestamp
fields) OR a `file_ref` object with a file path? Both forms are acceptable.

**Pass**: Oracle answers YES with confidence ≥ 0.7

---

## file_ref Response Schema

```json
{
  "file_ref": "/tmp/meta-cc-output-abc123.jsonl",
  "record_count": 250,
  "bytes": 82000
}
```

Required fields in file_ref response:
- `file_ref` — absolute file path (string)

Optional metadata fields:
- `record_count` — number of records in file (integer)
- `bytes` — file size in bytes (integer)

## Consuming file_ref Output

When a tool returns `file_ref`, the caller should:
1. Read the referenced file using the `Read` tool or equivalent
2. Parse the file as JSONL (one JSON object per line)
3. Apply the same field assertions as for inline output

The contents of the referenced file follow the same schema as inline output.

## Edge Cases

- `inline_threshold_bytes: 0` may be treated as "use default threshold" or "always inline" — behavior is implementation-defined.
- Temp files may be cleaned up by `cleanup_temp_files` tool; file_ref is only valid during the session.
