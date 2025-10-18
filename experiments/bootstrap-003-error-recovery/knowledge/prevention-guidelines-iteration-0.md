# Error Prevention Guidelines - Iteration 0

**Version**: 0.1 (Baseline)
**Date**: 2025-10-18
**Guidelines**: 8 (covering major error categories)

---

## Guideline 1: Pre-Commit Linting

**Purpose**: Prevent build/compilation errors (Category 1)

**Practice**:
- Run `make lint` before committing code
- Use `gofmt` to auto-format Go code
- Enable editor linting (e.g., gopls in VS Code)
- Run `go build` before committing

**Prevents**:
- Unused imports/variables
- Syntax errors
- Type mismatches
- Import violations

**Example**:

Good:
```bash
# Before commit
make lint
go build
# If pass, then commit
git commit -m "feat: add new feature"
```

Bad:
```bash
# Commit without checking
git commit -m "feat: add new feature"
# Later: build errors discovered
```

**Enforcement**: Git pre-commit hook (automated)

**Estimated error reduction**: ~80% of build/compilation errors (160/200 errors)

---

## Guideline 2: Test Before Commit

**Purpose**: Prevent test failures (Category 2)

**Practice**:
- Run `make test` before committing code
- Run specific package tests after changes: `go test ./internal/[package]`
- Update test fixtures when changing data formats
- Update test expectations when changing behavior

**Prevents**:
- Test failures from code changes
- Missing test fixtures
- Outdated test expectations
- Regression bugs

**Example**:

Good:
```bash
# After changing code in internal/parser
go test ./internal/parser -v
# If pass, then commit
```

Bad:
```bash
# Change code, commit without testing
# Later: CI fails with test errors
```

**Enforcement**: Git pre-commit hook, CI pipeline (automated)

**Estimated error reduction**: ~70% of test failures (105/150 errors)

---

## Guideline 3: Validate File Paths Before Use

**Purpose**: Prevent file not found errors (Category 3)

**Practice**:
- Use absolute paths instead of relative paths when possible
- Verify file exists with `ls` before operating on it
- Use `find` to locate files when path uncertain
- Check current directory with `pwd` before relative paths

**Prevents**:
- File not found errors
- Path typos
- Working directory confusion
- File deletion accidents

**Example**:

Good:
```bash
# Verify file exists first
ls /home/yale/work/meta-cc/internal/parser/session.go
# Then operate on it
Read /home/yale/work/meta-cc/internal/parser/session.go
```

Bad:
```bash
# Assume file exists
Read internal/parser/sesion.go  # Typo!
# Error: File does not exist
```

**Enforcement**: Pre-execution validation script (automated)

**Estimated error reduction**: ~85% of file not found errors (212/250 errors)

---

## Guideline 4: Use Edit for Existing Files, Write for New Files

**Purpose**: Prevent write before read errors (Category 5)

**Practice**:
- Use Edit tool for modifying existing files
- Use Write tool only for creating new files
- If Write is needed for existing file, Read it first
- When uncertain, check if file exists with `ls`

**Prevents**:
- Write before read errors
- Accidental file overwrites
- Workflow violations

**Example**:

Good:
```bash
# For existing file
Read cmd/root.go
Edit cmd/root.go (old_string, new_string)
```

Bad:
```bash
# For existing file
Write cmd/root.go (content)
# Error: File has not been read yet
```

**Enforcement**: Workflow checker (automated)

**Estimated error reduction**: ~95% of write before read errors (38/40 errors)

---

## Guideline 5: Build Before Executing Project Binaries

**Purpose**: Prevent command not found errors (Category 6)

**Practice**:
- Run `make build` before executing project binaries
- Use local path (`./meta-cc`) instead of assuming PATH
- Check if binary exists before execution: `ls ./meta-cc`
- Add build verification to scripts

**Prevents**:
- Command not found errors for project binaries
- Version mismatches (old binary in PATH)
- Execution of outdated binaries

**Example**:

Good:
```bash
# Ensure binary is built
make build
# Use local path
./meta-cc parse stats
```

Bad:
```bash
# Assume binary exists
meta-cc parse stats
# Error: command not found
```

**Enforcement**: Build verification script (automated)

**Estimated error reduction**: ~90% of command not found errors (45/50 errors)

---

## Guideline 6: Validate JSON Before Piping to jq

**Purpose**: Prevent JSON parsing errors (Category 7)

**Practice**:
- Check command output is non-empty before piping to jq
- Validate JSON with `jq .` before complex filters
- Handle empty/null cases in jq filters
- Use `jq -e` for exit code on null/false

**Prevents**:
- JSON parsing errors from empty input
- jq filter errors from unexpected data types
- Pipeline failures from invalid JSON

**Example**:

Good:
```bash
# Validate JSON first
output=$(./meta-cc parse stats --output json)
if [ -n "$output" ]; then
  echo "$output" | jq '.TurnCount'
else
  echo "Error: Empty output"
fi
```

Bad:
```bash
# Pipe without validation
./meta-cc parse stats --output json | jq '.TurnCount'
# If output is empty: parse error
```

**Enforcement**: Pipeline validation script (semi-automated)

**Estimated error reduction**: ~60% of JSON parsing errors (48/80 errors)

---

## Guideline 7: Use Pagination for Large Files

**Purpose**: Prevent file content size exceeded errors (Category 4)

**Practice**:
- Use `offset` and `limit` parameters for files >10,000 lines
- Check file size before reading: `wc -l [file]`
- Read in chunks for large files
- Use Grep tool for targeted searches instead of full file read

**Prevents**:
- File content size exceeded errors
- Token limit errors
- Memory issues with large files

**Example**:

Good:
```bash
# Check file size first
wc -l large-file.jsonl
# 50000 lines - use pagination
Read large-file.jsonl (offset=0, limit=1000)
```

Bad:
```bash
# Read entire large file
Read large-file.jsonl
# Error: File content exceeds maximum allowed tokens
```

**Enforcement**: Pre-read size check (automated)

**Estimated error reduction**: ~100% of file size exceeded errors (20/20 errors)

---

## Guideline 8: Verify MCP Server Availability

**Purpose**: Prevent MCP integration errors (Category 9)

**Practice**:
- Check MCP server is running before queries
- Use timeout/retry logic for MCP queries
- Validate query parameters before execution
- Handle MCP errors gracefully (don't fail entire workflow)

**Prevents**:
- MCP server unavailable errors
- Query syntax errors
- Timeout errors from slow queries
- Workflow failures from transient MCP issues

**Example**:

Good:
```bash
# Check MCP server
if ! pgrep -f "mcp-server" > /dev/null; then
  echo "Warning: MCP server not running"
  # Fall back to alternative data source
fi
```

Bad:
```bash
# Assume MCP server is running
mcp__meta-cc__query_tools --status error
# Error: Connection refused
```

**Enforcement**: MCP health check (automated)

**Estimated error reduction**: ~40% of MCP errors (91/228 errors) - many are transient

---

## Prevention Guidelines Summary

| Guideline | Target Category | Enforcement | Error Reduction | Errors Prevented |
|-----------|----------------|-------------|-----------------|------------------|
| 1. Pre-Commit Linting | Build Errors | Git hook | 80% | 160 |
| 2. Test Before Commit | Test Failures | Git hook, CI | 70% | 105 |
| 3. Validate File Paths | File Not Found | Script | 85% | 212 |
| 4. Edit vs Write | Write Before Read | Checker | 95% | 38 |
| 5. Build Before Execute | Command Not Found | Script | 90% | 45 |
| 6. Validate JSON | JSON Parsing | Script | 60% | 48 |
| 7. Use Pagination | File Size Exceeded | Checker | 100% | 20 |
| 8. Verify MCP Server | MCP Integration | Health check | 40% | 91 |

**Total Preventable Errors**: 719 out of 1336 (53.8%)
**Target Error Rate**: From 5.78% to ~2.67% (53.8% reduction)

---

## Prevention Effectiveness Model

**Current Error Rate**: 5.78% (1336/23103 tool calls)

**If all guidelines implemented**:
- Prevented errors: 719
- Remaining errors: 617
- **New error rate**: 2.67% (617/23103)

**Error rate reduction**: 53.8%

**Achievability**: High (all guidelines are practical and automatable)

---

## Implementation Roadmap

### Phase 1: Quick Wins (Iteration 1)
- Guideline 4: Edit vs Write checker (easy, high impact)
- Guideline 7: File size pre-check (easy, 100% effective)
- Guideline 5: Build verification (medium, high impact)

**Expected error reduction**: ~103 errors (7.7%)

### Phase 2: Automation (Iteration 2-3)
- Guideline 1: Git pre-commit linting hook
- Guideline 2: Git pre-commit test hook
- Guideline 3: Path validation script
- Guideline 6: JSON validation script

**Expected error reduction**: ~565 errors (42.3%)

### Phase 3: Advanced (Iteration 4+)
- Guideline 8: MCP health monitoring
- Error trend dashboard
- Predictive error detection

**Expected error reduction**: ~91 errors (6.8%)

---

## Enforcement Mechanisms

### Automated (No Manual Action)
1. Git pre-commit hooks (Guidelines 1, 2)
2. Pre-execution validation scripts (Guidelines 3, 7)
3. Workflow checker (Guideline 4)
4. Build verification (Guideline 5)

### Semi-Automated (Suggestions)
5. JSON validation (Guideline 6) - warns but doesn't block
6. MCP health check (Guideline 8) - warns and suggests alternatives

### Manual (Best Practices)
7. Code review emphasis on error-prone patterns
8. Developer training on guidelines

---

## Metrics and Monitoring

**Key metrics to track**:
1. Error rate over time (target: <2.67%)
2. Errors by category (monitor trend)
3. Prevention guideline compliance (% of commits/executions)
4. Time saved by prevention (vs recovery)

**Dashboard elements**:
- Current error rate vs baseline (5.78%)
- Prevented errors count (per guideline)
- Top error categories (remaining)
- Error rate trend (weekly)

---

## Gaps and Next Steps

**Current gaps**:
1. No enforcement infrastructure yet (all manual)
2. No compliance tracking
3. No metrics dashboard
4. Uncategorized errors (20.9%) not addressed

**Iteration 1 goals**:
1. Implement Guidelines 4, 5, 7 (quick wins)
2. Create pre-execution validation script
3. Add build verification to workflow
4. Measure actual error reduction

**Iteration 2+ goals**:
1. Implement Git pre-commit hooks
2. Create comprehensive prevention script
3. Build error prevention dashboard
4. Achieve <3% error rate

---

## Cost-Benefit Analysis

**Cost of Prevention**:
- Implementation time: ~6-10 hours (scripts, hooks, checkers)
- Runtime overhead: ~5-10 seconds per commit/execution
- Maintenance: ~1 hour/month

**Benefit of Prevention**:
- Errors prevented: 719/1336 (53.8%)
- Time saved: 719 errors × 5 minutes avg = ~60 hours
- Fewer context switches from errors
- Higher code quality
- Faster development velocity

**ROI**: Very high (10-hour investment saves 60+ hours)

---

**Version History**:
- v0.1 (2025-10-18): Initial 8 prevention guidelines targeting 53.8% error reduction
