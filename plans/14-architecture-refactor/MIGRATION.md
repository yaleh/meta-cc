# Phase 14 Migration Guide

## Overview

Phase 14 introduces architecture improvements that result in some breaking changes, primarily to the `analyze errors` command. This guide helps you migrate from Phase 13 to Phase 14.

---

## Breaking Changes

### 1. `analyze errors` Command Removed

**What Changed**: The `analyze errors` command has been replaced with `query errors`

**Why**: To clarify meta-cc's responsibility (data extraction only, not aggregation)

**Impact**: Medium - Affects users of `/meta-errors` Slash Command and direct CLI usage

#### Migration Path

**Before (Phase 13)**:
```bash
meta-cc analyze errors --window 50
```

**Output (Phase 13)**:
```json
[
  {
    "pattern_id": "err-a1b2c3d4",
    "type": "tool_error",
    "occurrences": 5,
    "signature": "1a2b3c4d5e6f...",
    "context": {
      "tool": "Bash",
      "error_prefix": "command not found: git-lfs",
      "first_turn": 45,
      "last_turn": 89,
      "time_span_seconds": 3600
    }
  }
]
```

**After (Phase 14)**:
```bash
meta-cc query errors | jq '.[-50:]'
```

**Output (Phase 14)**:
```json
[
  {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-10-05T10:30:00Z",
    "turn_index": 45,
    "tool_name": "Bash",
    "error": "command not found: git-lfs",
    "signature": "Bash:command not found: git-lfs"
  },
  {
    "uuid": "550e8400-e29b-41d4-a716-446655440001",
    "timestamp": "2025-10-05T11:45:00Z",
    "turn_index": 89,
    "tool_name": "Bash",
    "error": "command not found: git-lfs",
    "signature": "Bash:command not found: git-lfs"
  }
]
```

#### Equivalent Operations

**Count error patterns**:
```bash
# Phase 13
meta-cc analyze errors --window 100

# Phase 14
meta-cc query errors | jq '.[-100:] | group_by(.signature) | map({
    signature: .[0].signature,
    tool: .[0].tool_name,
    count: length,
    first_occurrence: .[0].timestamp,
    last_occurrence: .[-1].timestamp,
    sample_error: .[0].error
}) | sort_by(-.count)'
```

**Find repeated errors (≥3 occurrences)**:
```bash
# Phase 13
meta-cc analyze errors | jq 'select(.occurrences >= 3)'

# Phase 14
meta-cc query errors | jq 'group_by(.signature) | map({
    signature: .[0].signature,
    count: length,
    occurrences: .
}) | select(.count >= 3)'
```

**Time-based windowing**:
```bash
# Phase 13
meta-cc analyze errors --window 20  # Last 20 turns

# Phase 14
meta-cc query errors | jq '.[-20:]'  # Last 20 errors
```

---

### 2. Error Signature Format Changed

**What Changed**: Error signatures simplified from SHA256 hash to readable format

**Why**: Simpler signatures are easier to understand and debug

#### Comparison

**Phase 13**:
```json
{
  "signature": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d"
}
```

**Phase 14**:
```json
{
  "signature": "Bash:command not found: git-lfs"
}
```

**Format**: `{tool_name}:{error_text[:50]}`

**Impact**: Low - Signatures are more human-readable

---

### 3. Output Sorting is Now Deterministic

**What Changed**: All query commands now output deterministically sorted data

**Why**: Enables reliable CI/CD comparisons and debugging

**Impact**: Low - Output content unchanged, only order

#### Before (Phase 13)

```bash
# Run same query twice
$ meta-cc query tools --limit 10 > out1.jsonl
$ meta-cc query tools --limit 10 > out2.jsonl
$ diff out1.jsonl out2.jsonl
# Files may differ (random map iteration order)
```

#### After (Phase 14)

```bash
# Run same query twice
$ meta-cc query tools --limit 10 > out1.jsonl
$ meta-cc query tools --limit 10 > out2.jsonl
$ diff out1.jsonl out2.jsonl
# Files are identical (sorted by timestamp)
```

**Sorting Rules**:
- `query tools` → sorted by Timestamp
- `query messages` → sorted by TurnSequence
- `query errors` → sorted by Timestamp
- `parse stats` → sorted by metric name

**Impact**: Low - Output is now stable and reproducible

---

## Slash Commands Migration

### `/meta-errors` Command

**File**: `.claude/commands/meta-errors.md`

#### Phase 13 Version (Old)

```markdown
---
name: meta-errors
description: Analyze error patterns in current session
allowed_tools: [Bash]
argument-hint: [window-size]
---

Run error analysis:

```bash
if ! command -v meta-cc &> /dev/null; then
    echo "❌ Error: meta-cc not installed"
    exit 1
fi

WINDOW_SIZE=${1:-20}
meta-cc analyze errors --window "$WINDOW_SIZE" --output json
```
```

#### Phase 14 Version (New)

```markdown
---
name: meta-errors
description: Analyze error patterns in current session
allowed_tools: [Bash]
argument-hint: [window-size]
---

Run error analysis:

```bash
if ! command -v meta-cc &> /dev/null; then
    echo "❌ Error: meta-cc not installed"
    exit 1
fi

# Extract all errors
ERRORS=$(meta-cc query errors)

# Apply window size
WINDOW_SIZE=${1:-20}
WINDOWED=$(echo "$ERRORS" | jq ".[-$WINDOW_SIZE:]")

# Aggregate by signature
PATTERNS=$(echo "$WINDOWED" | jq 'group_by(.signature) | map({
    signature: .[0].signature,
    tool: .[0].tool_name,
    count: length,
    first_occurrence: .[0].timestamp,
    last_occurrence: .[-1].timestamp,
    sample_error: .[0].error,
    all_occurrences: [.[] | {turn: .turn_index, time: .timestamp}]
}) | sort_by(-.count)')

# Display results
echo "## Error Patterns (Last $WINDOW_SIZE errors)"
echo ""
echo "$PATTERNS" | jq -r '.[] | "\n### \(.tool): \(.signature)\n\n**Occurrences**: \(.count)\n**First**: \(.first_occurrence)\n**Last**: \(.last_occurrence)\n\n**Error**:\n```\n\(.sample_error)\n```\n"'
```
```

---

## MCP Server Migration

### Tool Definitions

**Phase 13**:
```json
{
  "name": "analyze_errors",
  "description": "Analyze error patterns in session",
  "inputSchema": {
    "type": "object",
    "properties": {
      "window": {
        "type": "integer",
        "description": "Analysis window size (last N turns)"
      }
    }
  }
}
```

**Phase 14**:
```json
{
  "name": "query_errors",
  "description": "Query all tool errors from session",
  "inputSchema": {
    "type": "object",
    "properties": {
      "limit": {
        "type": "integer",
        "description": "Limit output to N most recent errors"
      },
      "offset": {
        "type": "integer",
        "description": "Skip first M errors"
      }
    }
  }
}
```

### Tool Handler

**Before (Phase 13)**:
```go
func handleAnalyzeErrors(params map[string]interface{}) (interface{}, error) {
    window, _ := params["window"].(float64)

    // Call analyze errors command
    cmd := exec.Command("meta-cc", "analyze", "errors",
        "--window", fmt.Sprintf("%d", int(window)),
        "--output", "json")

    output, err := cmd.Output()
    // ... handle output ...
}
```

**After (Phase 14)**:
```go
func handleQueryErrors(params map[string]interface{}) (interface{}, error) {
    limit, _ := params["limit"].(float64)
    offset, _ := params["offset"].(float64)

    // Call query errors command
    args := []string{"query", "errors", "--output", "json"}
    if limit > 0 {
        args = append(args, "--limit", fmt.Sprintf("%d", int(limit)))
    }
    if offset > 0 {
        args = append(args, "--offset", fmt.Sprintf("%d", int(offset)))
    }

    cmd := exec.Command("meta-cc", args...)
    output, err := cmd.Output()

    // Parse JSONL output
    var errors []map[string]interface{}
    scanner := bufio.NewScanner(bytes.NewReader(output))
    for scanner.Scan() {
        var entry map[string]interface{}
        json.Unmarshal(scanner.Bytes(), &entry)
        errors = append(errors, entry)
    }

    return errors, nil
}
```

---

## Developer Migration

### Adding New Commands

**Phase 13 Pattern (Don't Use)**:
```go
func runNewCommand(cmd *cobra.Command, args []string) error {
    // ❌ Duplicate session location logic
    loc := locator.NewSessionLocator()
    var sessionPath string
    if sessionID != "" {
        sessionPath, _ = loc.FromSessionID(sessionID)
    } else if projectPath != "" {
        sessionPath, _ = loc.FromProjectPath(projectPath)
    } else {
        sessionPath, _ = loc.AutoDetect()
    }

    // ❌ Duplicate JSONL parsing
    entries, _ := parser.ParseJSONL(sessionPath)

    // ❌ Duplicate extraction
    tools, _ := parser.ExtractToolCalls(entries)

    // Command-specific logic
    result := processData(tools)

    // ❌ Duplicate output formatting
    switch outputFormat {
    case "json":
        data, _ := json.MarshalIndent(result, "", "  ")
        fmt.Println(string(data))
    case "tsv":
        // ...
    }
}
```

**Phase 14 Pattern (Use This)**:
```go
func runNewCommand(cmd *cobra.Command, args []string) error {
    // ✅ Use pipeline
    p := pipeline.NewSessionPipeline(getGlobalOptions())
    if err := p.Load(pipeline.LoadOptions{AutoDetect: true}); err != nil {
        return err
    }

    // ✅ Unified extraction
    tools, _ := p.ExtractToolCalls()

    // Command-specific logic
    result := processData(tools)

    // ✅ Deterministic sorting
    output.SortByTimestamp(result)

    // ✅ Unified output
    return output.Format(result, outputFormat)
}
```

**Benefits**:
- 70% less boilerplate code
- Consistent error handling
- Deterministic output
- Easier to test

---

## Testing Your Migration

### 1. Verify Equivalent Output

```bash
# Phase 13 baseline (if available)
meta-cc-phase13 analyze errors --window 50 > phase13-output.json

# Phase 14 equivalent
meta-cc query errors | jq '.[-50:] | group_by(.signature) | map({
    signature: .[0].signature,
    count: length
})' > phase14-output.json

# Compare (counts should match, signatures will differ in format)
jq '.[] | .count' phase13-output.json | sort -n > phase13-counts.txt
jq '.[] | .count' phase14-output.json | sort -n > phase14-counts.txt
diff phase13-counts.txt phase14-counts.txt
```

### 2. Verify Determinism

```bash
# Run same query multiple times
meta-cc query errors --limit 100 > out1.jsonl
meta-cc query errors --limit 100 > out2.jsonl
meta-cc query errors --limit 100 > out3.jsonl

# All should be identical
diff out1.jsonl out2.jsonl
diff out2.jsonl out3.jsonl
```

### 3. Verify Slash Commands

```bash
# Test in Claude Code environment
cd test-workspace

# Run /meta-errors command
# Verify output shows aggregated patterns
# Verify occurrences count correctly
```

### 4. Verify MCP Server

```bash
# List MCP tools
claude mcp list

# Test query_errors tool
# Should return JSONL array of errors
```

---

## Common Patterns

### Pattern 1: Recent Errors

```bash
# Last 50 errors
meta-cc query errors | jq '.[-50:]'

# Last 10 errors with details
meta-cc query errors | jq '.[-10:] | .[] | {
    tool: .tool_name,
    error: .error,
    time: .timestamp
}'
```

### Pattern 2: Error Aggregation

```bash
# Count by tool
meta-cc query errors | jq 'group_by(.tool_name) | map({
    tool: .[0].tool_name,
    error_count: length
})'

# Count by signature
meta-cc query errors | jq 'group_by(.signature) | map({
    signature: .[0].signature,
    count: length,
    sample: .[0].error
}) | sort_by(-.count)'
```

### Pattern 3: Time-Based Analysis

```bash
# Errors after specific time
meta-cc query errors | jq 'select(.timestamp > "2025-10-05T10:00:00Z")'

# Errors in date range
meta-cc query errors | jq 'select(
    .timestamp > "2025-10-05T00:00:00Z" and
    .timestamp < "2025-10-05T23:59:59Z"
)'
```

### Pattern 4: Complex Filtering

```bash
# Bash errors only
meta-cc query errors | jq 'select(.tool_name == "Bash")'

# Long errors (>100 chars)
meta-cc query errors | jq 'select(.error | length > 100)'

# Specific error pattern
meta-cc query errors | jq 'select(.error | contains("not found"))'
```

---

## Troubleshooting

### Issue 1: `analyze errors` command not found

**Error**:
```bash
$ meta-cc analyze errors
Error: unknown command "errors" for "meta-cc analyze"
```

**Solution**: Use `query errors` instead:
```bash
meta-cc query errors
```

### Issue 2: Missing `--window` parameter

**Error**:
```bash
$ meta-cc query errors --window 50
Error: unknown flag: --window
```

**Solution**: Use jq for windowing:
```bash
meta-cc query errors | jq '.[-50:]'
```

### Issue 3: Different signature format

**Issue**: Phase 13 signatures are SHA256 hashes, Phase 14 are readable strings

**Solution**: No action needed - new format is more human-readable

**Example**:
```
Phase 13: "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d"
Phase 14: "Bash:command not found: git-lfs"
```

### Issue 4: Output order changed

**Issue**: Query results appear in different order

**Solution**: No action needed - output is now deterministically sorted

**Note**: This is a feature, not a bug. Sorted output enables:
- Reliable diffs
- CI/CD comparisons
- Easier debugging

### Issue 5: Slash Command errors

**Error**: `/meta-errors` doesn't work or produces incorrect output

**Solution**: Update `.claude/commands/meta-errors.md` using the template above

---

## FAQ

### Q: Why was `analyze errors` removed?

**A**: To clarify responsibility boundaries. meta-cc should extract data, not analyze it. Aggregation and pattern detection should be done by jq (simple cases) or Claude/LLM (complex analysis).

### Q: Do I need to migrate immediately?

**A**: If you're using `analyze errors` directly or in Slash Commands, yes. Otherwise, Phase 14 is backward compatible.

### Q: Will my existing scripts break?

**A**: Only if they use `meta-cc analyze errors`. Use `meta-cc query errors` instead.

### Q: How do I replicate Phase 13 functionality?

**A**: Use jq for aggregation. See "Equivalent Operations" section above.

### Q: Is the new approach more complex?

**A**: Initially yes (learning jq), but long-term benefits:
- More flexible (jq is more powerful than fixed aggregation)
- Follows Unix philosophy (do one thing well)
- Easier to maintain meta-cc (less code, clearer responsibility)

### Q: What about performance?

**A**: Similar performance. jq is fast for most use cases. For very large sessions (>10,000 errors), use `--limit` to reduce jq processing time.

### Q: Can I still use `--window`?

**A**: No, use jq instead: `meta-cc query errors | jq '.[-50:]'`

### Q: Why are signatures different?

**A**: Readable signatures (`Bash:command not found`) are easier to understand than SHA256 hashes. If you need unique IDs, use the `uuid` field.

---

## Rollback Plan

If you need to revert to Phase 13:

```bash
# Checkout Phase 13 branch
git checkout feature/phase-13

# Rebuild
go build -o meta-cc

# Or use Phase 13 binary
cp meta-cc-phase13 meta-cc
```

**Note**: Phase 13 will not receive further updates. Migration to Phase 14 is recommended.

---

## Getting Help

- **Documentation**: `/home/yale/work/meta-cc/plans/14/iteration-14-implementation-plan.md`
- **Examples**: `/home/yale/work/meta-cc/docs/examples-usage.md`
- **Issues**: Check GitHub issues or create a new one

---

**Migration Guide Version**: 1.0
**Last Updated**: 2025-10-05
**Phase 14 Status**: Ready for implementation
