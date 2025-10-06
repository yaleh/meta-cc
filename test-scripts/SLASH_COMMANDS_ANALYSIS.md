# Slash Commands Analysis: meta-cc Calls

This document lists all existing slash commands and subagents, analyzing their `meta-cc` calls and scope (session vs project).

## Summary

**Total Slash Commands:** 8
**meta-cc calls:** Multiple commands per slash command
**Default Scope (Phase 13):** Project-level (using `--project .`)
**Session-only mode:** `--session-only` flag

## Slash Commands

### 1. /meta-stats

**Description:** Display current project's latest session statistics (Phase 13: project-level by default)

**meta-cc Calls:**
1. `meta-cc parse stats`
   - **Output:** JSONL (default)
   - **Scope:** Project-level (default), session-only with `--session-only`
   - **Processing:** jq renders JSONL to Markdown

2. `meta-cc stats aggregate --group-by tool --metrics "count,error_rate"`
   - **Output:** JSONL
   - **Scope:** Project-level
   - **Processing:** jq renders aggregated statistics

**Exit Code Handling:**
- 0: Success
- 1: Error
- 2: No data available

---

### 2. /meta-errors

**Description:** Analyze error patterns in current project's latest session (Phase 13: project-level by default)

**Arguments:**
- `window-size` (optional, default: 20)

**meta-cc Calls:**
1. `meta-cc query tools --where "status='error'" --stream`
   - **Output:** JSONL
   - **Scope:** Project-level
   - **Processing:** Count errors, stream to temp file

2. `meta-cc analyze errors --window <window-size>`
   - **Output:** JSONL
   - **Scope:** Project-level
   - **Processing:** jq renders error patterns to Markdown

**Features:**
- Detects repeated errors (≥3 occurrences)
- Provides optimization suggestions
- Large error set handling (>10 errors → show top 10)

**Exit Code Handling:**
- 0: Success
- 1: Error
- 2: No errors detected

---

### 3. /meta-query-tools

**Description:** Quick query tool calls in current project's latest session (Phase 13: project-level by default)

**Arguments:**
- `filter-expr` (optional): Tool name or WHERE expression
- `limit` (optional, default: 20)

**meta-cc Calls:**
1. `meta-cc query tools [--tool <tool-name>] [--where <expression>] --limit <limit> --stream`
   - **Output:** JSONL
   - **Scope:** Project-level
   - **Processing:** jq renders to summary and list

**Features:**
- Phase 10: Advanced filtering with WHERE expressions
- SQL-like syntax: `tool='Bash' AND status='error'`
- IN operator: `tool IN ('Bash','Edit')`
- Streaming output

**Exit Code Handling:**
- 0: Success
- 1: Error
- 2: No matching tool calls

---

### 4. /meta-query-messages

**Description:** Search user messages in current project's latest session (Phase 13: project-level by default)

**Arguments:**
- `pattern` (optional, default: `.*`)
- `limit` (optional, default: 10)

**meta-cc Calls:**
1. `meta-cc query user-messages --match <pattern> --limit <limit> --sort-by timestamp --reverse --output json`
   - **Output:** JSON
   - **Scope:** Project-level
   - **Processing:** jq renders messages with truncation

**Features:**
- Regex pattern matching
- Reverse chronological order
- Content truncation (300 characters)

---

### 5. /meta-timeline

**Description:** Generate timeline view of current project's latest session (Phase 13: project-level by default)

**Arguments:**
- `limit` (optional, default: 50)

**meta-cc Calls:**
1. `meta-cc query tools --limit <limit> --output json`
   - **Output:** JSON
   - **Scope:** Project-level
   - **Processing:** jq renders timeline with status indicators

2. `meta-cc analyze errors --window <limit> --output md` (if errors detected)
   - **Output:** Markdown
   - **Scope:** Project-level
   - **Processing:** Show error analysis

**Features:**
- Tool call sequence visualization
- Error detection and analysis
- Top tools summary
- Phase 8: Efficient pagination support

---

### 6. /meta-compare

**Description:** Compare current session with other project sessions

**Arguments:**
- `project-path` (optional)

**meta-cc Calls:**
1. `meta-cc parse stats --output md`
   - **Output:** Markdown
   - **Scope:** Current project

2. `meta-cc --project <project-path> parse stats --output md`
   - **Output:** Markdown
   - **Scope:** Specified project

**Features:**
- Cross-project comparison
- Efficiency analysis
- Tool usage pattern comparison

---

### 7. /meta-help

**Description:** Display all meta-cc commands and usage help

**meta-cc Calls:**
1. `meta-cc --help`
   - **Output:** Text
   - **Scope:** N/A (shows CLI help)

**Features:**
- Complete command reference
- Installation instructions
- Troubleshooting guide
- MCP server configuration

---

### 8. /meta-verify-build

**Description:** Verify meta-cc build with real session data

**Arguments:**
- `session` (optional, default: MVP session `6a32f273-191a-49c8-a5fc-a5dcba08531a`)

**meta-cc Calls:**
1. `meta-cc --session <session-id> parse stats --output json`
   - **Output:** JSON
   - **Scope:** Specific session
   - **Validation:** Check TurnCount field

2. `meta-cc --session <session-id> parse extract --type turns --output json`
   - **Output:** JSON
   - **Scope:** Specific session
   - **Validation:** Check array length

3. `meta-cc --session <session-id> parse extract --type tools --output json`
   - **Output:** JSON
   - **Scope:** Specific session
   - **Validation:** Check array length

4. `meta-cc --session <session-id> analyze errors --output json`
   - **Output:** JSON
   - **Scope:** Specific session
   - **Validation:** Check pattern count

**Features:**
- Automated build verification
- Real session testing
- Multi-command validation

---

## Subagents

### @meta-coach

**Status:** Not found in `.claude/subagents/` directory

**Expected Functionality:**
- Interactive meta-cognition coaching
- Workflow optimization suggestions
- Pattern detection and reflection
- Hook/command creation assistance

**Note:** Subagent implementation may be pending or located elsewhere.

---

## Scope Summary

| Command | Default Scope | Can Use Session? | Session Flag |
|---------|---------------|------------------|--------------|
| /meta-stats | Project | Yes | `--session-only` |
| /meta-errors | Project | Yes | `--session-only` |
| /meta-query-tools | Project | Yes | `--session-only` |
| /meta-query-messages | Project | Yes | `--session-only` |
| /meta-timeline | Project | Yes | `--session-only` |
| /meta-compare | Both | N/A | Uses `--project` |
| /meta-help | N/A | N/A | N/A |
| /meta-verify-build | Session | N/A | Uses `--session` |

---

## Output Formats

| Format | Commands | Purpose |
|--------|----------|---------|
| JSONL | stats, errors, query-tools | Default, Claude renders to Markdown |
| JSON | query-messages, timeline | Direct JSON output |
| Markdown | compare, help | Human-readable output |
| TSV | All query commands | Alternative with `--output tsv` |

---

## Phase 13 Changes

**Default Behavior:**
- All slash commands now default to **project-level** analysis
- Uses current working directory as project path
- Analyzes the latest session in the project

**Opt-Out:**
- Use `--session-only` flag to analyze only current session
- Reverts to environment variable detection (`$CC_SESSION_ID`)

**Migration Impact:**
- Old: `meta-cc parse stats` → analyzed current session only
- New: `meta-cc parse stats` → analyzes latest session in project (current directory)
- Explicit: `meta-cc --session-only parse stats` → current session only

---

## Validation Script Mapping

The validation script (`validate-meta-cc.sh`) implements Unix-based equivalents:

| meta-cc Command | Validation Script Command | Status |
|-----------------|---------------------------|--------|
| `parse stats` | `validate-meta-cc.sh stats` | ✅ Implemented |
| `analyze errors` | `validate-meta-cc.sh errors` | ⏳ Pending |
| `query tools` | `validate-meta-cc.sh query-tools` | ⏳ Pending |
| `query user-messages` | `validate-meta-cc.sh query-messages` | ⏳ Pending |
| N/A | `validate-meta-cc.sh timeline` | ⏳ Pending |

See `test-scripts/README.md` for validation workflow details.
