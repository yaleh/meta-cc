# Phase 14 Implementation Plan - Quick Reference

## Overview

**Goal**: Eliminate code duplication and clarify responsibility boundaries

**Total Effort**: ~600 lines of refactoring (net reduction of 854 lines, -72%)

**Timeline**: 2-3 days (TDD methodology)

---

## Stage Breakdown

### Stage 14.1: Pipeline Abstraction Layer (~120 lines)
- **Duration**: 0.5-1 day
- **Create**: `pkg/pipeline/session.go` - unified session processing
- **Deliverable**: SessionPipeline abstraction (locate → load → extract → output)
- **Tests**: ≥90% coverage
- **Key**: Eliminates 50 lines of duplicate session location code

### Stage 14.2: Simplify errors Command (~80 lines)
- **Duration**: 0.5 day
- **Replace**: `cmd/analyze.go` errors logic with `cmd/query_errors.go`
- **Deliverable**: Simple error list (no aggregation, let jq/LLM handle)
- **Breaking**: `--window` parameter removed, output format changes
- **Key**: Reduces errors command from 317 lines to 80 lines (-75%)

### Stage 14.3: Output Sorting Standardization (~50 lines)
- **Duration**: 0.5 day
- **Create**: `pkg/output/sort.go` - deterministic sorting utilities
- **Deliverable**: All query commands output sorted data
- **Tests**: Verify sorting is idempotent
- **Key**: Fixes non-deterministic Go map iteration

### Stage 14.4: Code Deduplication (~0 lines, -854 net)
- **Duration**: 0.5-1 day
- **Refactor**: 5 commands to use SessionPipeline
- **Deliverable**: parse stats, query tools, query messages, analyze sequences, analyze file-churn
- **Tests**: Integration tests verify behavior unchanged
- **Key**: Removes ~345 lines of duplicate code across commands

---

## Key Changes Summary

### Code Size Targets

```
Command              Before    After     Reduction
-----------------------------------------------------
parse stats          170 lines 60 lines  -65%
query tools          307 lines 80 lines  -74%
query messages       280 lines 70 lines  -75%
analyze errors       317 lines 80 lines  -75%
analyze sequences    120 lines 50 lines  -58%
-----------------------------------------------------
Total                1194 lines 340 lines -72%

Plus pipeline        0 lines   120 lines +120
-----------------------------------------------------
Net Change           1194 lines 460 lines -734 (-61.5%)
```

### Breaking Changes

#### 1. `analyze errors` → `query errors`

```bash
# ❌ Old (Phase 13)
meta-cc analyze errors --window 50

# ✅ New (Phase 14)
meta-cc query errors | jq '.[-50:]'
```

#### 2. Error Output Format

**Before**: Aggregated patterns with counts
```json
[
  {
    "pattern_id": "err-a1b2",
    "occurrences": 5,
    "signature": "sha256-hash",
    "first_seen": "...",
    "last_seen": "..."
  }
]
```

**After**: Simple error list (aggregate with jq)
```json
[
  {
    "uuid": "...",
    "timestamp": "...",
    "tool_name": "Bash",
    "error": "command not found",
    "signature": "Bash:command not found"
  }
]
```

#### 3. Output Sorting

All query commands now output deterministically sorted data:
- `query tools` → sorted by Timestamp
- `query messages` → sorted by TurnSequence
- `query errors` → sorted by Timestamp

---

## Design Principles

### 1. Responsibility Minimization
**meta-cc**: Extract data only, no analysis decisions
**LLM/tools**: Perform aggregation, filtering, pattern detection

### 2. Pipeline Pattern
```
SessionPipeline:
  locate → load → parse → extract → output
```

### 3. Output Determinism
Stable, sorted output for CI/CD comparisons

### 4. Code Reuse First
No duplicate session location/parsing code

### 5. Delayed Decision
Push filtering/windowing to downstream tools (jq, awk, LLM)

---

## Testing Strategy

### Unit Tests
```bash
# Run all tests
go test ./... -v

# Test specific stage
go test ./pkg/pipeline -v          # Stage 14.1
go test ./cmd -run TestQueryErrors # Stage 14.2
go test ./pkg/output -run TestSort # Stage 14.3

# Check coverage
go test ./pkg/pipeline -cover
```

### Integration Tests
```bash
# Validation script
./test-scripts/validate-phase-14.sh

# Verify behavior equivalence
meta-cc query tools --limit 50  # Should match Phase 13 output (sorted)
```

### Regression Tests
```bash
# Existing tests should still pass
go test ./...
./tests/integration/slash_commands_test.sh
```

---

## Migration Guide

### For Users

**Aggregate errors with jq**:
```bash
# Count patterns
meta-cc query errors | jq 'group_by(.signature) | map({
    signature: .[0].signature,
    count: length,
    tool: .[0].tool_name,
    sample: .[0].error
})'

# Last 50 errors
meta-cc query errors | jq '.[-50:]'

# Errors in time window
meta-cc query errors | jq 'select(.timestamp > "2025-10-01")'
```

### For Slash Commands

**Update `.claude/commands/meta-errors.md`**:
```bash
# Extract errors (new command)
ERRORS=$(meta-cc query errors)

# Aggregate with jq
PATTERNS=$(echo "$ERRORS" | jq 'group_by(.signature) | map({
    signature: .[0].signature,
    count: length,
    sample: .[0].error
})')
```

### For Developers

**Use pipeline for new commands**:
```go
func runNewCommand(cmd *cobra.Command, args []string) error {
    // Use pipeline
    p := pipeline.NewSessionPipeline(getGlobalOptions())
    if err := p.Load(pipeline.LoadOptions{AutoDetect: true}); err != nil {
        return err
    }

    // Extract data
    tools, _ := p.ExtractToolCalls()

    // Process
    result := processData(tools)

    // Sort (deterministic output)
    output.SortByTimestamp(result)

    // Unified output
    return output.Format(result, outputFormat)
}
```

---

## Success Criteria

### Functional
- ✅ All commands produce identical output to Phase 13 (except analyze errors)
- ✅ Output is deterministically sorted
- ✅ `query errors` replaces `analyze errors` with simpler output
- ✅ Pipeline handles all 3 session location methods
- ✅ No regressions in existing functionality

### Code Quality
- ✅ Code reduction ≥60% (1194 → ≤460 lines including pipeline)
- ✅ Unit test coverage ≥80%
- ✅ All unit tests pass
- ✅ Integration tests pass
- ✅ No duplicate session location/parsing code

### Documentation
- ✅ Migration guide in README.md
- ✅ Breaking changes documented
- ✅ Slash Commands updated
- ✅ Pipeline usage examples
- ✅ `query errors` command documented

### Integration
- ✅ Slash Commands work (except meta-errors needs update)
- ✅ MCP Server returns deterministic output
- ✅ Phase 13 integration tests pass
- ✅ Real-world validation (3 projects)

---

## Common Use Cases

### Pattern 1: Error Analysis
```bash
# Extract errors
meta-cc query errors

# Aggregate by signature
meta-cc query errors | jq 'group_by(.signature) | map({
    sig: .[0].signature,
    count: length
}) | sort_by(-.count)'

# Recent errors (last 50)
meta-cc query errors | jq '.[-50:]'
```

### Pattern 2: Deterministic Output
```bash
# Sorted by timestamp (deterministic)
meta-cc query tools --limit 100

# Same query produces same output
OUTPUT1=$(meta-cc query tools --limit 50)
OUTPUT2=$(meta-cc query tools --limit 50)
# $OUTPUT1 == $OUTPUT2
```

### Pattern 3: Pipeline Usage
```go
// New command template
func runMyCommand(cmd *cobra.Command, args []string) error {
    p := pipeline.NewSessionPipeline(getGlobalOptions())
    p.Load(pipeline.LoadOptions{AutoDetect: true})

    tools, _ := p.ExtractToolCalls()

    // Your logic here
    result := analyze(tools)

    output.SortByTimestamp(result)
    return output.Format(result, outputFormat)
}
```

---

## Risk Assessment

### High Risk: Breaking Changes
**Mitigation**: Migration guide, CHANGELOG warnings, jq examples

### Medium Risk: Code Size Underestimation
**Mitigation**: Refactor one command at a time, keep tests passing

### Medium Risk: Pipeline Bugs
**Mitigation**: ≥90% test coverage, canary testing (refactor one command first)

### Low Risk: Sorting Performance
**Mitigation**: Benchmark (expect +2% overhead, acceptable)

---

## Timeline

**Day 1**: Pipeline foundation (Stage 14.1)
**Day 2**: Simplification & sorting (Stages 14.2, 14.3)
**Day 3**: Deduplication & validation (Stage 14.4)

---

## Quick Commands

### Build
```bash
go build -o meta-cc
```

### Test
```bash
# All tests
go test ./... -v

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Validation
./test-scripts/validate-phase-14.sh
```

### Verify Code Size
```bash
wc -l cmd/*.go pkg/pipeline/*.go
```

### Check Determinism
```bash
meta-cc query tools --limit 100 > out1.jsonl
meta-cc query tools --limit 100 > out2.jsonl
diff out1.jsonl out2.jsonl  # Should be identical
```

---

## Next Steps

1. Review `iteration-14-implementation-plan.md` for detailed specs
2. Start with Stage 14.1 (pipeline abstraction)
3. Run tests after each stage
4. Update Slash Commands and documentation
5. Validate with real projects

**Estimated Timeline**: 2-3 days

**Risk Level**: Medium (breaking changes, extensive refactoring)

**Success Probability**: High (clear design, comprehensive tests, incremental approach)

**Ready to Begin**: ✅
