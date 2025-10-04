# Phase 13 Implementation Plan - Quick Reference

## Overview

**Goal**: Simplify output formats from 5 to 2 (JSONL + TSV), align with Unix philosophy

**Total Effort**: ~400 lines of code changes across 4 stages

**Timeline**: 2-3 days (TDD methodology)

---

## Stage Breakdown

### Stage 13.1: Remove Redundant Formats (~100 lines)
- **Duration**: 0.5 day
- **Remove**: JSON pretty, CSV, Markdown formatters
- **Update**: Command parameter validation (12 files)
- **Deliverable**: JSONL default output
- **Tests**: Format validation, backward compatibility checks

### Stage 13.2: Enhance TSV for All Data Types (~120 lines)
- **Duration**: 0.5 day
- **Add**: Generic TSV formatter with reflection
- **Add**: Type-specific extractors (ToolCall, SessionStats, ErrorPattern, etc.)
- **Deliverable**: TSV supports all meta-cc data types
- **Tests**: TSV formatting for all types, field projection

### Stage 13.3: Unify Error Handling (~100 lines)
- **Duration**: 0.5 day
- **Add**: Structured error output (JSONL/TSV)
- **Add**: Exit code standardization (0=success, 1=error, 2=no results)
- **Update**: All commands error handling (12 files)
- **Deliverable**: Consistent error format across all scenarios
- **Tests**: Error output validation, exit code checks

### Stage 13.4: Update Documentation (~80 lines)
- **Duration**: 0.5 day
- **Update**: README.md, Slash Commands, MCP docs
- **Add**: Unix composability guide (200 lines)
- **Add**: Integration tests
- **Deliverable**: Complete migration guide and examples
- **Tests**: Integration tests, Slash Commands verification

---

## Key Changes Summary

### Removed
- `pkg/output/csv.go` (-71 lines)
- `pkg/output/markdown.go` (-204 lines)
- Total: **-275 lines**

### Added
- `internal/output/format.go` (+30 lines)
- `pkg/output/tsv_extractors.go` (+80 lines)
- `internal/output/error.go` (+60 lines)
- `docs/cli-composability.md` (+200 lines)
- Tests and documentation (+375 lines)
- Total: **+745 lines**

### Modified
- 12 command files (error handling)
- TSV formatter enhancement
- README and integration docs
- Total: **~395 lines**

### Net Impact
- Source code: **+430 lines** (including tests)
- Documentation: **+375 lines**
- Total: **+805 lines**

---

## Breaking Changes

### 1. Default Format Change
**Before**: `--output json` (pretty-printed)
**After**: `--output jsonl` (one JSON per line)

**Migration**:
```bash
# Old
meta-cc query tools --output json

# New (automatic)
meta-cc query tools

# For pretty-print
meta-cc query tools | jq '.'
```

### 2. Removed Formats
**Removed**: `--output json|md|csv`
**Kept**: `--output jsonl|tsv`

**Migration**:
```bash
# Old: CSV
meta-cc query tools --output csv

# New: TSV
meta-cc query tools --output tsv

# Old: Markdown
meta-cc query tools --output md

# New: Let Claude Code render
meta-cc query tools  # Claude parses JSONL
```

### 3. Error Output Format
**Before**: Plain text to stderr
**After**: Structured JSON/TSV with exit codes

**Migration**:
```bash
# JSONL format
{"error": "session not found", "code": "SESSION_NOT_FOUND"}

# TSV format (stderr)
Error: session not found (code: SESSION_NOT_FOUND)
```

---

## Design Principles

### 1. Dual Format Principle
- **JSONL** (default): Machine-readable, composable
- **TSV**: CLI-friendly, human-readable

### 2. Format Consistency
- All scenarios output valid format (success/error/no results)
- No format downgrades or fallbacks

### 3. Data-Log Separation
- **stdout**: Data only (JSONL/TSV)
- **stderr**: Diagnostic logs only

### 4. Unix Composability
- meta-cc: Simple retrieval
- jq/awk/grep: Complex filtering
- Principle: Do one thing well

### 5. No Auto-Downgrade
- Client (Claude Code) renders JSONL to Markdown
- meta-cc doesn't make rendering decisions

---

## Testing Strategy

### Unit Tests
```bash
# Stage 13.1: Format removal
go test ./pkg/output -v -run TestFormat

# Stage 13.2: TSV enhancement
go test ./pkg/output -v -run TestFormatTSV

# Stage 13.3: Error handling
go test ./internal/output -v -run TestOutputError

# All tests
go test ./... -v
```

### Integration Tests
```bash
# Format migration tests
./tests/integration/format_migration_test.sh

# Slash Commands tests
./tests/integration/slash_commands_test.sh
```

### Real-World Validation
```bash
# JSONL default
meta-cc query tools --limit 10

# TSV output
meta-cc query tools --limit 10 --output tsv

# jq pipeline
meta-cc query tools | jq 'select(.Status == "error")'

# awk pipeline
meta-cc query tools --output tsv | awk -F'\t' '{print $2}'
```

---

## Success Criteria

### Functional
- ✅ JSONL is default for all commands
- ✅ TSV supports all data types
- ✅ Errors output in consistent format
- ✅ Exit codes standardized (0/1/2)
- ✅ jq/awk pipelines work correctly

### Integration
- ✅ Slash Commands receive JSONL
- ✅ Claude Code renders Markdown
- ✅ MCP Server works with JSONL
- ✅ No regressions in existing tests

### Documentation
- ✅ README updated with format philosophy
- ✅ Unix composability guide complete
- ✅ Migration guide available
- ✅ Integration tests pass

### Code Quality
- ✅ Test coverage ≥ 80%
- ✅ All unit tests pass
- ✅ All integration tests pass
- ✅ Code follows TDD methodology

---

## Common Use Cases

### Pattern 1: Error Analysis
```bash
# JSONL + jq
meta-cc query tools | jq 'select(.Status == "error")'

# TSV + awk
meta-cc query tools --output tsv | awk -F'\t' '$3 == "error"'
```

### Pattern 2: Tool Usage Stats
```bash
# Count by tool
meta-cc query tools | jq -r '.ToolName' | sort | uniq -c

# TSV version
meta-cc query tools --output tsv | awk -F'\t' 'NR>1 {print $2}' | sort | uniq -c
```

### Pattern 3: Complex Filtering
```bash
# Multi-condition filter
meta-cc query tools | jq 'select(.Duration > 5000 and .ToolName == "Bash")'

# Time range filter
meta-cc parse extract --type turns | jq 'select(.timestamp > "2025-10-01")'
```

### Pattern 4: Report Generation
```bash
# CSV report
meta-cc query tools | jq -r '[.UUID, .ToolName, .Status] | @csv' > report.csv

# Markdown report
{
  echo "# Session Report"
  meta-cc parse stats | jq -r '"Turns: \(.TurnCount)"'
  echo "## Tools"
  meta-cc query tools | jq -r '.ToolName' | sort | uniq -c
} > report.md
```

---

## FAQ

### Q: Why remove JSON pretty-print?
**A**: Use `jq '.'` for pretty-print. This follows Unix philosophy (one tool = one job).

### Q: Why remove Markdown output?
**A**: Claude Code should render JSONL to Markdown. meta-cc shouldn't make UI decisions.

### Q: Why remove CSV?
**A**: TSV is simpler (no quoting), faster, and more suitable for Unix tools.

### Q: Will old scripts break?
**A**: Yes, if they use `--output json|md|csv`. Migration is simple: use `jsonl` or `tsv` + jq/awk.

### Q: How to migrate Slash Commands?
**A**: Replace `--output md` with JSONL, let Claude render. See `.claude/commands/meta-*.md` examples.

### Q: Performance impact?
**A**: JSONL is faster (streaming) than JSON pretty. TSV is faster than CSV (no quoting).

---

## Next Steps

1. Review `plan.md` for detailed implementation
2. Start with Stage 13.1 (remove formats)
3. Run tests after each stage
4. Update integration tests
5. Verify Slash Commands work
6. Update documentation

**Estimated Timeline**: 2-3 days with TDD methodology

**Risk Level**: Medium (breaking changes, requires migration)

**Mitigation**: Comprehensive tests, migration guide, rollback plan
