# Iteration 5 Observations

**Date**: 2025-10-17
**Meta-Agent Capability**: M.observe
**Iteration**: 5
**Previous State**: s₄ (V_instance = 0.615, V_meta = 0.545)

---

## Executive Summary

Iteration 5 observation phase analyzed:
- **Error Site Coverage**: 273 total fmt.Errorf across 56 files (165 with %w wrapping)
- **Sentinel Error Usage**: 8 occurrences across 4 files (from Iteration 4)
- **Linter Requirements**: go/analysis framework complex, simpler approach needed
- **High-Impact Targets**: cmd/mcp-server/* has 77 error sites (opportunities for standardization)

**Key Findings**:
1. Error wrapping adoption good (165/273 = 60.4% with %w)
2. Sentinel error adoption low (8 total uses in 4 files only)
3. cmd/mcp-server has concentrated error sites (executor.go: 8, capabilities.go: 44, temp_file_manager.go: 7)
4. Full go/analysis linter likely too complex for single iteration
5. Simpler linter script (bash + grep + go vet) more achievable

---

## Data Collection

### Error Sites Analysis

**Total Error Sites** (fmt.Errorf usage):
- **Total files**: 56 files
- **Total occurrences**: 273 error sites
- **With %w wrapping**: 165 sites (60.4%)
- **Without %w wrapping**: 108 sites (39.6%)

**Sentinel Error Usage** (mcerrors.*):
- **Total files**: 4 files
- **Total occurrences**: 8 uses
- **Files**:
  - internal/query/file_access.go (1)
  - internal/query/sequences.go (1)
  - internal/query/context.go (1)
  - internal/mcp/builder.go (5)

**Error Sites by Package**:

| Package | fmt.Errorf Count | %w Wrapped | Sentinel Errors |
|---------|------------------|------------|-----------------|
| cmd/mcp-server | 77 | ~55 | 0 |
| cmd/* | 48 | ~35 | 0 |
| internal/locator | 14 | ~8 | 0 |
| internal/filter | 12 | ~7 | 0 |
| internal/parser | 6 | ~4 | 0 |
| internal/query | 3 | 3 | 3 |
| internal/mcp | 5 | 5 | 5 |
| internal/errors | 7 | 7 | 0 (test file) |
| pkg/output | 13 | ~10 | 0 |
| pkg/pipeline | 4 | ~3 | 0 |
| internal/stats | 4 | ~2 | 0 |
| internal/config | 4 | 1 | 0 |
| internal/githelper | 8 | 8 | 0 |
| Other | ~74 | ~17 | 0 |

### High-Impact Files (Error Standardization Opportunities)

**cmd/mcp-server/capabilities.go** (44 fmt.Errorf):
- Package download errors (network, file I/O)
- Capability parsing errors (YAML, frontmatter)
- GitHub source errors (URL parsing, fetch)
- Most errors already use %w wrapping
- **Opportunity**: Add sentinel errors (ErrNotFound, ErrInvalidInput, ErrNetworkFailure)

**cmd/mcp-server/executor.go** (8 fmt.Errorf):
- Unknown tool errors
- jq filter errors
- JSONL parse errors
- Response adaptation errors
- **Opportunity**: Add ErrUnknownTool, ErrInvalidInput sentinel errors

**cmd/mcp-server/temp_file_manager.go** (7 fmt.Errorf):
- File creation errors
- Directory creation errors
- File sync/close errors
- **Opportunity**: Add ErrFileIO sentinel error

**cmd/mcp-server/response_adapter.go** (4 fmt.Errorf):
- Temp file write errors
- Unknown output mode errors
- **Opportunity**: Add ErrInvalidInput, ErrFileIO sentinel errors

**internal/locator/*.go** (14 fmt.Errorf):
- Session locating errors
- Environment variable errors
- **Opportunity**: Standardize with ErrNotFound, ErrMissingParameter

### Linter Requirements Analysis

**Option 1: go/analysis Framework (Complex)**

Complexity assessment:
- **Framework Knowledge**: Requires deep understanding of go/analysis API
- **AST Traversal**: Need to traverse Go AST to find error patterns
- **Pattern Matching**: Complex logic to identify:
  - fmt.Errorf without %w
  - fmt.Errorf with %w but no sentinel error
  - Missing error context
- **Integration**: Need to package as analyzer tool
- **Testing**: Requires test fixtures and expected outputs
- **Estimated LOC**: 300-500 lines (analyzer + tests)
- **Estimated Time**: 4-6 hours

**Complexity Score**: 8/10 (HIGH - likely requires specialized agent)

**Option 2: Simple Linter Script (Achievable)**

Approach:
- Bash script using grep + basic pattern matching
- Check for common anti-patterns:
  - `fmt.Errorf` without `%w` (missing error wrapping)
  - `return nil, fmt.Errorf` without sentinel error import
  - Missing operation context in error messages
- Report violations with file:line
- Exit code for CI integration
- **Estimated LOC**: 50-100 lines (bash + documentation)
- **Estimated Time**: 1-2 hours

**Complexity Score**: 3/10 (LOW - achievable with coder agent)

**Option 3: Hybrid Approach (Pragmatic)**

Strategy:
- Phase 1: Error standardization (manual, guided by conventions)
- Phase 2: Simple linter script for common patterns
- Phase 3: (Future) Full go/analysis linter when complexity justified

**Benefits**:
- Incremental progress
- Quick wins (error standardization)
- Foundation for automation (simple linter)
- Defers complex linter until demonstrated need

---

## Pattern Recognition

### Error Handling Patterns Observed

**Pattern 1: Well-Wrapped Errors** (60.4% of sites)
```go
return nil, fmt.Errorf("failed to X: %w", err)
```
**Status**: ✓ Good (preserves error chain)
**Improvement**: Add sentinel error wrapping

**Pattern 2: Missing Error Wrapping** (39.6% of sites)
```go
return nil, fmt.Errorf("error message: %v", err)
```
**Status**: ✗ Bad (loses error chain, can't use errors.Is/As)
**Action**: Convert %v → %w

**Pattern 3: Sentinel Error Usage** (3% of sites)
```go
return nil, fmt.Errorf("context: %w", mcerrors.ErrMissingParameter)
```
**Status**: ✓✓ Excellent (programmatic + context)
**Goal**: Increase adoption to 80%+

**Pattern 4: Multiple Error Wrapping** (rare)
```go
return nil, fmt.Errorf("operation X failed: %w", fmt.Errorf("sub-operation Y: %w", err))
```
**Status**: ⚠ Acceptable but verbose
**Note**: Usually indicates layered error handling

### Identified Gaps

**Gap 1: Low Sentinel Error Adoption**
- Current: 8 uses (3% of wrapped errors)
- Target: 130+ uses (80% of wrapped errors)
- **Impact**: Missing programmatic error handling

**Gap 2: Inconsistent Error Context**
- Some errors have rich context (file, line, operation)
- Others minimal ("error: X")
- **Impact**: Difficult debugging

**Gap 3: No Automated Enforcement**
- Manual review only
- No CI checks
- **Impact**: Patterns degrade over time

**Gap 4: Missing Error Categories**
- Need: ErrFileIO, ErrNetworkFailure, ErrParseError
- Current sentinel errors insufficient for all cases
- **Impact**: Some errors can't use sentinel pattern

---

## Priorities for Iteration 5

### Priority 1: Expand Sentinel Error Set (HIGH)

**Objective**: Add missing sentinel errors for common categories

**Candidates**:
- `ErrFileIO` - File operations (create, write, read, close)
- `ErrNetworkFailure` - HTTP requests, downloads
- `ErrParseError` - YAML, JSON, TOML parsing
- `ErrConfigError` - Configuration validation

**Rationale**:
- cmd/mcp-server has many file I/O and network errors
- Current 5 sentinel errors insufficient
- Low cost (5-10 LOC per error + tests)

**Expected Impact**: V_enforcement +0.05-0.10

### Priority 2: Standardize cmd/mcp-server Errors (HIGH)

**Objective**: Apply sentinel errors to 77 error sites in cmd/mcp-server

**Scope**:
- capabilities.go (44 sites): Network, parsing, file errors
- executor.go (8 sites): Unknown tool, parsing errors
- temp_file_manager.go (7 sites): File I/O errors
- response_adapter.go (4 sites): File I/O, invalid input
- Other files (14 sites): Various

**Approach**:
- Systematic file-by-file standardization
- Test after each file
- Commit after each file (safety)

**Expected Impact**: V_consistency +0.10-0.15

### Priority 3: Create Simple Linter Script (MEDIUM)

**Objective**: Automate basic error pattern checking

**Checks**:
1. `fmt.Errorf` without `%w` (anti-pattern)
2. Error messages without context (too short, <20 chars)
3. Missing sentinel error import in files with errors

**Output**:
- File:Line:Pattern violation
- Exit code 1 if violations found
- JSON output for CI integration

**Integration**:
- Add to Makefile (`make lint-errors`)
- Add to CI workflow (GitHub Actions)

**Expected Impact**: V_enforcement +0.15-0.20

### Priority 4: Update Error Conventions Document (LOW)

**Objective**: Document new sentinel errors and patterns

**Updates**:
- Add new sentinel errors (ErrFileIO, etc.)
- Update examples
- Add linter usage documentation

**Expected Impact**: V_documentation +0.00-0.05

---

## Constraints and Risks

### Constraints

**LOC Limit**: Phase maximum 500 LOC
- Sentinel errors: ~30 LOC
- Error standardization: ~150 LOC (estimates)
- Linter script: ~100 LOC
- **Total estimated**: ~280 LOC (within limit ✓)

**Time Limit**: ~3-4 hours for implementation
- Sentinel errors: 30 minutes
- Error standardization: 2 hours
- Linter script: 1 hour
- Testing: 30 minutes
- **Total estimated**: ~4 hours (achievable)

### Risks

**Risk 1: Scope Creep (Linter Complexity)**
- **Mitigation**: Use simple grep-based linter (not go/analysis)
- **Trigger**: If linter >150 LOC, defer to Iteration 6

**Risk 2: Test Failures After Error Changes**
- **Mitigation**: Test after each file, commit frequently
- **Recovery**: Revert single file if tests fail

**Risk 3: Sentinel Error Category Mismatch**
- **Mitigation**: Review error categories before implementation
- **Validation**: Check against cmd/mcp-server actual error types

---

## Recommendations for M.plan

### Recommended Iteration 5 Objectives

1. **Add 4 new sentinel errors** (ErrFileIO, ErrNetworkFailure, ErrParseError, ErrConfigError)
2. **Standardize 40-50 error sites in cmd/mcp-server** (focus on high-impact files)
3. **Create simple linter script** (grep-based, CI-integrated)
4. **Update error conventions documentation**

### Recommended Phase Structure

**Phase 1: Expand Sentinel Errors** (30 minutes, ~30 LOC)
- Add 4 new sentinel errors to internal/errors/errors.go
- Add tests to internal/errors/errors_test.go
- Run tests, ensure all passing

**Phase 2: Standardize cmd/mcp-server Errors** (2 hours, ~150 LOC)
- File-by-file error standardization
- Order: executor.go → temp_file_manager.go → response_adapter.go → capabilities.go (partial)
- Test after each file
- Target: 40-50 sites (50-65% of cmd/mcp-server errors)

**Phase 3: Create Simple Linter** (1 hour, ~100 LOC)
- Bash script: scripts/lint-errors.sh
- Check 3 patterns (no %w, short messages, missing imports)
- Integrate into Makefile
- Add documentation

**Phase 4: Update Documentation** (30 minutes, ~20 LOC)
- Update iteration-2-error-conventions.md
- Add linter usage examples

### Expected Value Improvements

**V_instance(s₅)**:
- V_consistency: 0.58 → 0.70 (+0.12) - 50+ error sites standardized
- V_maintainability: 0.60 → 0.65 (+0.05) - New sentinel errors + documentation
- V_enforcement: 0.15 → 0.40 (+0.25) - Linter automation
- V_documentation: 0.80 → 0.85 (+0.05) - Updated conventions

**Calculation**:
```
V_instance(s₅) = 0.4×0.70 + 0.3×0.65 + 0.2×0.40 + 0.1×0.85
                = 0.28 + 0.195 + 0.08 + 0.085
                = 0.64
```

**ΔV_instance = +0.025 (+4.1%)**

**V_meta(s₅)**:
- V_completeness: 0.72 → 0.77 (+0.05) - Linter methodology documented
- V_effectiveness: 0.35 → 0.45 (+0.10) - Automation improves efficiency
- V_reusability: 0.52 → 0.58 (+0.06) - Linter pattern transferable

**Calculation**:
```
V_meta(s₅) = 0.4×0.77 + 0.3×0.45 + 0.3×0.58
            = 0.308 + 0.135 + 0.174
            = 0.617
```

**ΔV_meta = +0.072 (+13.2%)**

---

## Data Sources

### Grep Queries Used

```bash
# Total fmt.Errorf usage
grep -r "fmt\.Errorf" **/*.go | wc -l  # 273 total

# Error wrapping with %w
grep -r "fmt\.Errorf.*%w" **/*.go | wc -l  # 165 wrapped

# Sentinel error usage
grep -r "mcerrors\." **/*.go | wc -l  # 8 uses

# cmd/mcp-server errors
grep -n "fmt\.Errorf" cmd/mcp-server/*.go  # 77 sites
```

### Files Analyzed

- internal/errors/errors.go (sentinel errors definition)
- cmd/mcp-server/*.go (77 error sites)
- internal/query/*.go (standardized in Iteration 4)
- internal/mcp/builder.go (standardized in Iteration 4)

---

**Status**: COMPLETE
**Next**: M.plan (create detailed iteration plan)
**Generated By**: M.observe
**Duration**: ~30 minutes
