# Task 3: Documentation Updates Specification

**Agent**: doc-writer
**Date**: 2025-10-15
**Iteration**: 3
**Status**: Design Complete (Ready for Implementation)

---

## Objective

Update MCP documentation to reflect tier-based parameter ordering and add consistency guidelines reference.

---

## Files to Update

### 1. docs/guides/mcp.md

**Current State**: Contains examples with inconsistent parameter ordering
**Target State**: All examples use tier-based ordering

**Changes**:

#### Section: Parameter Ordering Convention (NEW)

Add new section after introduction:

```markdown
## Parameter Ordering Convention

meta-cc MCP tools follow a **tier-based parameter ordering** convention for consistency:

**Tier System**:
1. **Tier 1**: Required parameters (pattern, error_signature, etc.)
2. **Tier 2**: Filtering parameters (tool, status, pattern_target, etc.)
3. **Tier 3**: Range parameters (min_*, max_*, start_*, end_*, threshold, window)
4. **Tier 4**: Output control (limit, offset)
5. **Tier 5**: Standard parameters (scope, jq_filter, stats_only, etc.) - automatic

**Rationale**: Consistent ordering improves readability and reduces cognitive load.

**Note**: JSON parameter order doesn't affect function calls (parameters are key-value pairs).

**Reference**: See `data/api-parameter-convention.md` for complete specification.
```

#### Update: query_tools Examples

**Current** (inconsistent):
```json
{
  "limit": 10,
  "tool": "Bash",
  "status": "error"
}
```

**Updated** (tier-based):
```json
{
  "tool": "Bash",
  "status": "error",
  "limit": 10
}
```

**Locations**:
- Basic usage example (line ~45)
- Advanced filtering example (line ~120)
- Error query example (line ~230)

#### Update: query_user_messages Examples

**Current**:
```json
{
  "pattern": "fix.*bug",
  "limit": 20,
  "max_message_length": 500
}
```

**Updated**:
```json
{
  "pattern": "fix.*bug",
  "max_message_length": 500,
  "limit": 20
}
```

**Locations**:
- Pattern search example (line ~180)
- Message search with limit (line ~310)

#### Update: query_conversation Examples

**Current**:
```json
{
  "start_turn": 10,
  "end_turn": 20,
  "pattern": "error",
  "pattern_target": "any",
  "limit": 5
}
```

**Updated**:
```json
{
  "pattern": "error",
  "pattern_target": "any",
  "start_turn": 10,
  "end_turn": 20,
  "limit": 5
}
```

**Locations**:
- Conversation filtering example (line ~275)

---

### 2. docs/reference/cli.md

**Current State**: Missing `validate-api` command documentation
**Target State**: Complete CLI reference including validation tool

**Changes**:

#### Add: meta-cc validate-api Section

Insert after existing query commands:

```markdown
## meta-cc validate-api

**Purpose**: Validate API consistency according to established conventions

**Usage**:
```bash
meta-cc validate-api [OPTIONS]
```

**Options**:
- `--file <path>` - Path to tools.go (default: cmd/mcp-server/tools.go)
- `--fast` - Run fast checks only (MVP mode, default)
- `--quiet` - Suppress output except errors
- `--json` - Output results as JSON

**Exit Codes**:
- `0` - All checks passed
- `1` - Violations found
- `2` - Invalid usage or error

**Checks Performed** (MVP):
1. **Naming pattern validation**: Verifies tool names use standard prefixes (query_*, get_*, list_*, cleanup_*)
2. **Parameter ordering validation**: Checks tier-based parameter ordering
3. **Description format validation**: Validates description template compliance

**Example Usage**:
```bash
# Validate current API
meta-cc validate-api

# Validate specific file
meta-cc validate-api --file path/to/tools.go

# JSON output for CI integration
meta-cc validate-api --json

# Quiet mode (errors only)
meta-cc validate-api --quiet
```

**Example Output**:
```
API Consistency Validation
==========================

Analyzing cmd/mcp-server/tools.go...
Found 16 tools

Running checks (MVP mode):
  ✓ Naming pattern validation
  ✓ Parameter ordering validation
  ✓ Description format validation

Results:
--------

✗ get_session_stats: Naming pattern violation
  Tool name should use query_* prefix
  Suggestion: Rename to query_session_stats
  Severity: ERROR

✗ query_tools: Parameter ordering violation
  Expected: tool, status, limit
  Actual:   limit, tool, status
  Severity: ERROR

Summary:
--------
Total tools:     16
Checks run:      48
Passed:          45
Failed:          3

Overall Status: FAILED (3 violations found)

Exit code: 1
```

**Integration**:
- **CI Pipeline**: Use `--json` output for automated checking
- **Pre-Commit Hook**: Run `--fast` mode before commits
- **Development**: Run manually before PRs

**References**:
- Naming conventions: `data/api-naming-convention.md`
- Parameter ordering: `data/api-parameter-convention.md`
- Validation methodology: `data/api-consistency-methodology.md`
```

---

### 3. docs/guides/git-hooks.md (NEW)

**Create New File**: Git hooks usage guide

```markdown
# Git Hooks for API Consistency

This guide explains how to use git hooks to automatically validate API consistency before commits.

---

## Pre-Commit Hook

### Purpose

The pre-commit hook runs `meta-cc validate-api` before each git commit to ensure API consistency.

**Benefits**:
- Catches violations early (before commit)
- Prevents inconsistent code from entering repository
- Enforces API conventions automatically

---

## Installation

### Automatic Installation

Run the installation script:

```bash
./scripts/install-consistency-hooks.sh
```

This will:
1. Copy pre-commit hook to `.git/hooks/pre-commit`
2. Make hook executable
3. Test hook with sample validation

### Manual Installation

1. Copy hook template:
   ```bash
   cp scripts/pre-commit.sample .git/hooks/pre-commit
   chmod +x .git/hooks/pre-commit
   ```

2. Verify installation:
   ```bash
   .git/hooks/pre-commit
   ```

---

## Hook Behavior

### What It Does

1. Detects if `cmd/mcp-server/tools.go` was modified
2. Runs `meta-cc validate-api --fast`
3. **If violations found**: Blocks commit, shows errors
4. **If no violations**: Allows commit to proceed

### Example (Passing)

```bash
$ git commit -m "Add new query tool"

Validating API consistency...
API Consistency Validation: PASSED

✓ All checks passed
✓ Commit allowed
```

### Example (Failing)

```bash
$ git commit -m "Add new retrieve_data tool"

Validating API consistency...
API Consistency Validation: FAILED

✗ retrieve_data: Naming pattern violation
  Tool name should use query_* prefix
  Suggestion: Rename to query_data

❌ Commit blocked. Fix violations before committing.
```

---

## Bypassing Hook (Not Recommended)

In exceptional cases, you can bypass the hook:

```bash
git commit --no-verify -m "Emergency fix (bypassing validation)"
```

**Warning**: Only use `--no-verify` for urgent fixes. Always fix violations afterward.

---

## Disabling Hook

To temporarily disable the hook:

```bash
mv .git/hooks/pre-commit .git/hooks/pre-commit.disabled
```

To re-enable:

```bash
mv .git/hooks/pre-commit.disabled .git/hooks/pre-commit
```

---

## Troubleshooting

### Hook Not Running

**Issue**: Commit proceeds without validation
**Cause**: Hook not executable
**Fix**:
```bash
chmod +x .git/hooks/pre-commit
```

### Hook Failing Incorrectly

**Issue**: Hook reports violations but tools.go wasn't changed
**Cause**: Hook detecting unrelated changes
**Fix**: Review hook logic in `.git/hooks/pre-commit`

### Hook Slow

**Issue**: Hook takes too long to run
**Cause**: Running full validation instead of fast mode
**Fix**: Ensure hook uses `--fast` flag

---

## Advanced Configuration

### Custom Validation File

Edit `.git/hooks/pre-commit` to validate different file:

```bash
# Change this line:
./meta-cc validate-api --fast --file cmd/mcp-server/tools.go

# To:
./meta-cc validate-api --fast --file path/to/other/file.go
```

### Additional Checks

Add more validation steps:

```bash
#!/bin/bash

# API consistency
./meta-cc validate-api --fast || exit 1

# Additional checks
go fmt ./cmd/mcp-server/... || exit 1
go vet ./cmd/mcp-server/... || exit 1
```

---

## References

- Validation tool: `docs/reference/cli.md#meta-cc-validate-api`
- API conventions: `data/api-consistency-methodology.md`
- Hook script: `scripts/pre-commit.sample`
```

---

## Expected Outputs

### Updated Files

1. **docs/guides/mcp.md**
   - New section: "Parameter Ordering Convention"
   - Updated examples: query_tools, query_user_messages, query_conversation (10-15 examples total)
   - Consistent tier-based ordering throughout

2. **docs/reference/cli.md**
   - New section: "meta-cc validate-api"
   - Complete command reference with examples
   - Integration guidance (CI, pre-commit, development)

3. **docs/guides/git-hooks.md** (NEW)
   - Installation guide (automatic + manual)
   - Hook behavior explanation
   - Troubleshooting section
   - Advanced configuration

---

## Quality Assurance

### Pre-Update Checklist

- [x] Identify all parameter examples in mcp.md
- [x] Determine correct ordering for each example
- [x] Plan CLI reference structure

### Post-Update Checklist

- [ ] All parameter examples use tier-based ordering
- [ ] Parameter ordering section added to mcp.md
- [ ] validate-api command fully documented in cli.md
- [ ] Git hooks guide complete (installation, usage, troubleshooting)
- [ ] No broken links or references
- [ ] Markdown formatted correctly

---

## Validation

### Example Count Verification

```bash
# Count parameter examples in mcp.md
grep -c '^\s*".*":' docs/guides/mcp.md
# Expected: 50-80 parameter examples

# Verify tier-based ordering
grep -A 5 'query_tools' docs/guides/mcp.md | grep 'limit'
# Expected: limit appears AFTER tool and status
```

### Link Verification

```bash
# Check all internal links
grep -o '\[.*\](.*\.md)' docs/**/*.md | sort | uniq
# Verify all referenced files exist
```

---

## Success Criteria

✅ Parameter ordering section added to mcp.md
✅ 10-15 parameter examples updated (tier-based ordering)
✅ validate-api command documented in cli.md
✅ Git hooks guide created (git-hooks.md)
✅ All links valid
✅ Markdown lint passes

---

## Effort Estimate

**Time**: 2-3 hours
- 1 hour: Update mcp.md examples (find, update, verify)
- 0.5 hour: Add validate-api documentation to cli.md
- 0.5 hour: Create git-hooks.md
- 0.5 hour: Verification and testing
- 0.5 hour: Link checking and formatting

**Complexity**: LOW (documentation updates, find-and-replace)

---

**Specification Status**: ✅ COMPLETE
**Ready for Implementation**: YES
**Next Step**: Update documentation files
