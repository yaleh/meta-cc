# Test Scripts

This directory contains validation scripts for verifying `meta-cc` output using traditional Unix commands.

## Overview

The validation script (`validate-meta-cc.sh`) processes Claude Code session history files (JSONL format) using Unix tools like `jq`, `grep`, `sed`, and `awk`. This allows us to independently verify that `meta-cc` is producing correct output.

## Usage

```bash
./validate-meta-cc.sh <command> <path> [args...]
```

### Path Argument

The `<path>` argument can be:
- **A file** (session scope): Process a single `.jsonl` session file
- **A directory** (project scope): Process the latest `.jsonl` file in the directory

### Scope Detection

The script automatically detects the scope:
```bash
# Session scope (single file)
./validate-meta-cc.sh stats ~/.claude/projects/abc-def/session-123.jsonl

# Project scope (directory - uses latest session)
./validate-meta-cc.sh stats ~/.claude/projects/abc-def
./validate-meta-cc.sh stats .  # Current project
```

## Commands

### stats

Calculate session statistics matching `meta-cc parse stats`:

```bash
./validate-meta-cc.sh stats <path>
```

**Output fields:**
- `TurnCount` - Total conversation turns
- `UserTurnCount` - User messages
- `AssistantTurnCount` - Assistant messages
- `ToolCallCount` - Total tool invocations
- `ErrorCount` - Failed tool calls
- `ErrorRate` - Percentage of failed calls
- `DurationSeconds` - Session duration
- `ToolFrequency` - Tool usage counts
- `TopTools` - Top tools with percentages

**Example:**
```bash
$ ./validate-meta-cc.sh stats ~/.claude/projects/-home-yale-work-meta-cc/6a32f273-191a-49c8-a5fc-a5dcba08531a.jsonl
{
  "TurnCount": 2676,
  "UserTurnCount": 1097,
  "AssistantTurnCount": 1579,
  "ToolCallCount": 1012,
  "ErrorCount": 0,
  "ErrorRate": 0,
  "DurationSeconds": 33797,
  "ToolFrequency": {
    "Bash": 495,
    "Read": 162,
    "TodoWrite": 140,
    ...
  },
  "TopTools": [
    {
      "Name": "Bash",
      "Count": 495,
      "Percentage": 48
    },
    ...
  ]
}
```

### errors

Analyze error patterns (not yet implemented):

```bash
./validate-meta-cc.sh errors <path> [window]
```

Args:
- `window` - Number of recent turns to analyze (default: 20)

### query-tools

Query tool calls (not yet implemented):

```bash
./validate-meta-cc.sh query-tools <path> [filter] [limit]
```

Args:
- `filter` - Tool name or status (default: all)
- `limit` - Maximum results (default: 20)

### query-messages

Search user messages (not yet implemented):

```bash
./validate-meta-cc.sh query-messages <path> [pattern] [limit]
```

Args:
- `pattern` - Regex pattern (default: .*)
- `limit` - Maximum results (default: 10)

### timeline

Generate timeline view (not yet implemented):

```bash
./validate-meta-cc.sh timeline <path> [limit]
```

Args:
- `limit` - Number of recent turns (default: 50)

## Validation Workflow

To verify `meta-cc` output:

1. **Run validation script:**
   ```bash
   ./test-scripts/validate-meta-cc.sh stats <session-file> > unix-output.json
   ```

2. **Run meta-cc:**
   ```bash
   ./meta-cc --session <session-id> parse stats > meta-cc-output.json
   ```

3. **Compare outputs:**
   ```bash
   diff <(jq -S . unix-output.json) <(jq -S . meta-cc-output.json)
   ```

Expected differences:
- `DurationSeconds` may differ by ±1 second (rounding)
- Field ordering may differ
- Meta-cc may include additional computed fields

## Dependencies

Required Unix tools:
- `jq` - JSON processor
- `grep` - Pattern matching
- `sed` - Stream editor
- `awk` - Text processing
- `bc` - Calculator (for error rate percentage)

Install on Ubuntu/Debian:
```bash
sudo apt-get install jq bc
```

Install on macOS:
```bash
brew install jq
```

## Implementation Status

- [x] Framework and scope detection
- [x] `stats` command
- [ ] `errors` command
- [ ] `query-tools` command
- [ ] `query-messages` command
- [ ] `timeline` command

## Technical Notes

### JSONL Structure

Claude Code session files use JSONL format with these entry types:

```json
{"type": "user", "message": {...}, "timestamp": "..."}
{"type": "assistant", "message": {...}, "timestamp": "..."}
{"type": "file-history-snapshot", ...}
```

**Relevant types:**
- `user` - User messages
- `assistant` - Assistant messages (may contain tool calls)

**Tool calls** are nested in assistant messages:
```json
{
  "type": "assistant",
  "message": {
    "content": [
      {"type": "text", "text": "..."},
      {"type": "tool_use", "name": "Bash", "id": "...", "input": {...}}
    ]
  }
}
```

### Verification Approach

The validation script uses a different implementation path than meta-cc:

- **meta-cc**: Go-based parser with structured types
- **validation script**: Shell + jq streaming processor

This provides **independent verification** that the parsing logic is correct.

## Future Work

1. Implement remaining commands (errors, query-tools, query-messages, timeline)
2. Add comprehensive test suite with known sessions
3. Create automated comparison tests
4. Add performance benchmarks
5. Support JSON output format alongside JSONL

## Contributing

When implementing new commands:

1. Match the output structure of corresponding `meta-cc` command
2. Test with multiple session files
3. Document expected differences (if any)
4. Update this README with examples
5. Add to the implementation status checklist
