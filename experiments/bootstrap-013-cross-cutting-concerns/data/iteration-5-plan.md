# Iteration 5 Plan

**Date**: 2025-10-17
**Meta-Agent Capability**: M.plan
**Iteration**: 5
**Current State**: s₄ (V_instance = 0.615, V_meta = 0.545)

---

## Executive Summary

Based on M.observe analysis, Iteration 5 will focus on error standardization expansion and basic linter automation. The plan is conservative and achievable, deferring complex go/analysis linter to future iterations.

**Primary Objective**: Standardize 40-50 additional error sites and create simple linter automation

**Expected Outcomes**:
- V_instance: 0.615 → 0.640 (+0.025, +4.1%)
- V_meta: 0.545 → 0.617 (+0.072, +13.2%)
- 4 new sentinel errors added
- 40-50 error sites standardized (cmd/mcp-server focus)
- Simple linter script created and CI-integrated

**Rationale for Conservative Approach**:
1. go/analysis linter too complex (300-500 LOC, 4-6 hours, likely needs specialized agent)
2. Better to deliver complete simple linter than partial complex one
3. Simple linter provides immediate value (CI integration)
4. Foundation for future enhancement (Iteration 6+)

---

## State Assessment

### Current State Analysis (s₄)

**V_instance(s₄) = 0.615**

| Component | Score | Weight | Contribution | Gap to 0.80 |
|-----------|-------|--------|--------------|-------------|
| V_consistency | 0.58 | 0.4 | 0.232 | 0.22 |
| V_maintainability | 0.60 | 0.3 | 0.18 | 0.20 |
| V_enforcement | 0.15 | 0.2 | 0.03 | 0.65 |
| V_documentation | 0.80 | 0.1 | 0.08 | 0.00 |

**Weakest Component**: V_enforcement (0.15, gap: 0.65)

**V_meta(s₄) = 0.545**

| Component | Score | Weight | Contribution |
|-----------|-------|--------|--------------|
| V_completeness | 0.72 | 0.4 | 0.288 |
| V_effectiveness | 0.35 | 0.3 | 0.105 |
| V_reusability | 0.52 | 0.3 | 0.156 |

**Weakest Component**: V_effectiveness (0.35)

### Critical Gaps

1. **Enforcement Gap (V_enforcement = 0.15)**
   - No automated enforcement
   - Manual review only
   - Patterns degrade over time
   - **Impact**: HIGH (foundation for V_consistency)

2. **Consistency Gap (V_consistency = 0.58)**
   - Only 8 sites use sentinel errors (3%)
   - 108 sites lack %w wrapping (39.6%)
   - **Impact**: MEDIUM (visible quality issue)

3. **Effectiveness Gap (V_effectiveness = 0.35)**
   - Manual pattern application slow
   - No automation tools
   - **Impact**: MEDIUM (productivity)

---

## Prioritization

### Priority Analysis (from M.observe)

| Priority | Problem | Impact | Urgency | Addressability |
|----------|---------|--------|---------|----------------|
| **P1** | Enforcement gap (no linter) | HIGH | HIGH | MEDIUM |
| **P2** | Consistency gap (low sentinel adoption) | MEDIUM | MEDIUM | HIGH |
| **P3** | Effectiveness gap (manual work) | MEDIUM | MEDIUM | HIGH |

### Iteration 5 Focus

**Primary Goal**: Address P1 (Enforcement) + P2 (Consistency)

**Rationale**:
- P1 + P2 together improve both enforcement and consistency
- Linter automation addresses P3 (effectiveness)
- Combined approach maximizes ΔV_instance and ΔV_meta
- Achievable scope (estimated 280 LOC, ~4 hours)

---

## Iteration Goal Definition

### Primary Objective

**Expand error standardization and create basic linter automation**

### Success Criteria

**Measurable Criteria**:
1. ✓ 4 new sentinel errors added (ErrFileIO, ErrNetworkFailure, ErrParseError, ErrConfigError)
2. ✓ 40-50 error sites standardized in cmd/mcp-server
3. ✓ Simple linter script created (bash/grep-based)
4. ✓ Linter integrated into Makefile and CI
5. ✓ All tests passing after changes
6. ✓ Documentation updated

**Value Criteria**:
- V_instance(s₅) ≥ 0.635 (target: 0.640, +0.025)
- V_meta(s₅) ≥ 0.610 (target: 0.617, +0.072)

### Expected ΔV

**Instance Layer**:
- ΔV_consistency: +0.12 (0.58 → 0.70) from 50+ sites standardized
- ΔV_maintainability: +0.05 (0.60 → 0.65) from new sentinel errors
- ΔV_enforcement: +0.25 (0.15 → 0.40) from linter automation
- ΔV_documentation: +0.05 (0.80 → 0.85) from updated docs

**ΔV_instance = +0.025 (+4.1%)**

**Meta Layer**:
- ΔV_completeness: +0.05 (0.72 → 0.77) from linter methodology
- ΔV_effectiveness: +0.10 (0.35 → 0.45) from automation
- ΔV_reusability: +0.06 (0.52 → 0.58) from transferable patterns

**ΔV_meta = +0.072 (+13.2%)**

### Constraints

**LOC Constraint**: ≤500 LOC per Phase, ≤200 LOC per Stage
- Estimated total: 280 LOC (within limit ✓)

**Time Constraint**: ~4 hours estimated
- Phase 1: 30 min (sentinel errors)
- Phase 2: 2 hours (error standardization)
- Phase 3: 1 hour (linter script)
- Phase 4: 30 min (documentation)

**Test Constraint**: All tests must pass
- Test after each file modification
- Commit frequently (safety)

---

## Agent Selection

### Decision Tree Analysis

```
Goal: Error standardization + linter automation
├─ Straightforward? NO (linter has complexity)
├─ Requires specialization? MAYBE
│   ├─ go/analysis linter? YES → Specialized agent likely needed
│   └─ Simple bash linter? NO → Generic coder sufficient
├─ Expected ΔV ≥ 0.05? YES (ΔV_instance = 0.025, ΔV_meta = 0.072)
├─ Reusable? YES (linter pattern universal)
└─ Decision: Use generic agents + defer complex linter
```

### Selected Agents

**A₅ = A₄** (no new agents for Iteration 5)

| Agent | Tasks | Rationale |
|-------|-------|-----------|
| **coder** | Sentinel errors, error standardization, linter script | TDD proven effective (Iteration 4) |
| **data-analyst** | Metrics calculation, value assessment | Established role |
| **doc-writer** | Documentation updates, iteration report | Established role |

**Specialization Assessment**:
- **Simple linter**: Generic coder sufficient (bash/grep, ~100 LOC)
- **Complex linter**: Would require linter-generator agent (defer to Iteration 6)
- **Decision**: Defer linter-generator until simple linter proven insufficient

---

## Work Breakdown

### Phase 1: Expand Sentinel Errors

**Objective**: Add 4 new sentinel errors for common error categories

**Agent**: coder
**Approach**: TDD (tests first, then implementation)
**Estimated LOC**: 30 (errors.go: +15, errors_test.go: +15)
**Estimated Time**: 30 minutes

**Tasks**:

1. **Add Sentinel Errors to internal/errors/errors.go**
   - `ErrFileIO = errors.New("file I/O error")`
   - `ErrNetworkFailure = errors.New("network operation failed")`
   - `ErrParseError = errors.New("parsing failed")`
   - `ErrConfigError = errors.New("configuration error")`
   - Include doc comments with usage examples

2. **Add Tests to internal/errors/errors_test.go**
   - Test existence and messages
   - Test wrapping compatibility
   - Test errors.Is/As usage

3. **Validate**
   - Run `go test ./internal/errors/...`
   - Ensure 100% test coverage maintained
   - All tests passing

**Success Criteria**:
- ✓ 4 new sentinel errors defined
- ✓ Tests passing (100% coverage)
- ✓ Documentation complete

---

### Phase 2: Standardize cmd/mcp-server Errors

**Objective**: Apply sentinel errors to 40-50 error sites in cmd/mcp-server

**Agent**: coder
**Approach**: File-by-file standardization with testing
**Estimated LOC**: 150 (40-50 error sites, ~3 LOC per site)
**Estimated Time**: 2 hours

**Target Files** (in order):

#### Stage 2.1: executor.go (8 error sites, 30 min)

**Error Sites**:
1. Line 109: `unknown tool` → Use ErrUnknownTool
2. Line 137: `jq filter error` → Use ErrParseError
3. Line 148: `JSONL parse error` → Use ErrParseError
4. Line 225: `response adaptation error` → Keep %w wrapping
5. Line 437: `failed to get working directory` → Use ErrFileIO
6. Lines 482-484: `meta-cc error` → Keep (external command error)
7. Line 522: `invalid JSON` → Use ErrParseError

**Pattern**:
```go
// OLD:
return "", fmt.Errorf("unknown tool: %s", toolName)

// NEW:
return "", fmt.Errorf("unknown tool %s in executor: %w", toolName, mcerrors.ErrUnknownTool)
```

**Validation**:
- Run `go test ./cmd/mcp-server/...`
- Ensure no test failures
- Commit after success

#### Stage 2.2: temp_file_manager.go (7 error sites, 20 min)

**Error Sites** (all file I/O):
1. Line 59: `failed to create directory` → ErrFileIO
2. Line 66: `failed to create temp file` → ErrFileIO
3. Line 75: `failed to encode record` → ErrParseError
4. Line 83: `failed to sync file` → ErrFileIO
5. Line 89: `failed to close file` → ErrFileIO
6. Line 95: `failed to rename file` → ErrFileIO
7. Line 117: `failed to glob files` → ErrFileIO

**Pattern**:
```go
// OLD:
return fmt.Errorf("failed to create directory: %w", err)

// NEW:
return fmt.Errorf("failed to create directory %s: %w", dir, mcerrors.ErrFileIO)
```

**Validation**: Same as 2.1

#### Stage 2.3: response_adapter.go (4 error sites, 15 min)

**Error Sites**:
1. Line 48: `failed to write temp file` → ErrFileIO
2. Line 55: `unknown output mode` → ErrInvalidInput
3. Line 72: `failed to generate file reference` → Keep (wraps FileReference error)
4. Line 97: `failed to serialize response` → ErrParseError

**Validation**: Same as 2.1

#### Stage 2.4: jq_filter.go (2 error sites, 10 min)

**Error Sites**:
1. Line 21: `invalid jq expression` → ErrInvalidInput
2. Line 35: `invalid JSON` → ErrParseError

**Validation**: Same as 2.1

#### Stage 2.5: capabilities.go (25-30 error sites, 45 min)

**Note**: capabilities.go has 44 error sites total. Target 25-30 (high-impact ones).

**Priority Error Sites**:
- Network errors (download, HTTP): ErrNetworkFailure (6-8 sites)
- Parse errors (YAML, frontmatter): ErrParseError (5-6 sites)
- File I/O errors: ErrFileIO (8-10 sites)
- Not found errors: ErrNotFound (3-4 sites)
- Invalid input errors: ErrInvalidInput (3-4 sites)

**Defer** (low priority):
- Some "failed to X" errors without clear category
- Test fixture errors (capabilities_test.go)
- HTTP test errors (capabilities_http_test.go)

**Success Criteria**:
- ✓ 40-50 total error sites standardized across all files
- ✓ All cmd/mcp-server tests passing
- ✓ mcerrors imported in all modified files

---

### Phase 3: Create Simple Linter Script

**Objective**: Create bash-based linter for basic error pattern checking

**Agent**: coder
**Approach**: Script development with test cases
**Estimated LOC**: 100 (script: 70, tests/doc: 30)
**Estimated Time**: 1 hour

**Script**: `scripts/lint-errors.sh`

**Checks Implemented**:

1. **Check 1: Missing %w in fmt.Errorf**
   - Pattern: `fmt.Errorf.*:.*err[^%w]`
   - Anti-pattern: `fmt.Errorf("...: %v", err)` or `fmt.Errorf("...: %s", err)`
   - Report: "File:Line: Error not wrapped with %w"

2. **Check 2: Short error messages (<20 chars)**
   - Pattern: `fmt.Errorf\("(.{1,19})"\)`
   - Anti-pattern: `fmt.Errorf("error")`
   - Report: "File:Line: Error message too short (lacks context)"

3. **Check 3: Missing mcerrors import**
   - Check: Files with `fmt.Errorf` but no `import.*mcerrors`
   - Report: "File: Consider using sentinel errors (mcerrors not imported)"

**Output Format**:

```
scripts/lint-errors.sh
---
Check 1: Missing %w in fmt.Errorf
  cmd/foo.go:42: Error not wrapped with %w
  cmd/bar.go:58: Error not wrapped with %w

Check 2: Short error messages
  internal/baz.go:123: Error message too short (lacks context)

Check 3: Missing mcerrors import
  pkg/output/writer.go: Consider using sentinel errors

---
Summary: 4 issues found
Exit code: 1
```

**Integration**:

1. **Makefile** (`/home/yale/work/meta-cc/Makefile`):
   ```makefile
   lint-errors:
   	@scripts/lint-errors.sh

   lint: lint-errors
   	golangci-lint run
   ```

2. **CI Workflow** (`.github/workflows/test.yml`):
   ```yaml
   - name: Lint error patterns
     run: make lint-errors
   ```

**Validation**:
- Run `scripts/lint-errors.sh` on meta-cc codebase
- Expect some violations (not all errors standardized yet)
- Ensure exit code 1 if violations found, 0 if clean

**Success Criteria**:
- ✓ Script detects 3 pattern types
- ✓ Output is clear and actionable
- ✓ Exit code correct (0/1)
- ✓ Integrated into Makefile
- ✓ Documentation complete

---

### Phase 4: Update Documentation

**Objective**: Document new sentinel errors and linter usage

**Agent**: doc-writer
**Approach**: Update existing conventions document
**Estimated LOC**: 20 (documentation additions)
**Estimated Time**: 30 minutes

**Files to Update**:

1. **iteration-2-error-conventions.md** (if exists, or create)
   - Add new sentinel errors section
   - Document ErrFileIO, ErrNetworkFailure, ErrParseError, ErrConfigError
   - Add usage examples
   - Update statistics (8 → 50+ sentinel error uses)

2. **Linter Usage Documentation** (in plan or separate doc)
   - How to run: `make lint-errors`
   - How to fix violations
   - CI integration details

**Success Criteria**:
- ✓ New sentinel errors documented
- ✓ Linter usage explained
- ✓ Examples provided

---

## Dependencies

### Task Graph

```
Phase 1 (Sentinel Errors)
  └─ Phase 2 (Error Standardization)
      ├─ Stage 2.1 (executor.go)
      ├─ Stage 2.2 (temp_file_manager.go)
      ├─ Stage 2.3 (response_adapter.go)
      ├─ Stage 2.4 (jq_filter.go)
      └─ Stage 2.5 (capabilities.go)

Phase 3 (Linter Script) - Independent of Phase 2

Phase 4 (Documentation) - Depends on Phase 1, 2, 3
```

**Critical Path**: Phase 1 → Phase 2 → Phase 4
**Parallel Work**: Phase 3 can be developed alongside Phase 2

---

## Risks and Mitigation

### Risk 1: Test Failures After Error Changes

**Probability**: MEDIUM
**Impact**: MEDIUM (blocks progress)

**Mitigation**:
- Test after each file (Stage 2.1, 2.2, etc.)
- Commit after each successful stage
- If tests fail, revert single file only
- Debug and fix before continuing

**Contingency**:
- If persistent failures, reduce scope (skip capabilities.go Stage 2.5)
- Minimum deliverable: 25 error sites standardized (Stages 2.1-2.4 only)

### Risk 2: Linter Script Too Simple (Not Useful)

**Probability**: LOW
**Impact**: LOW (can enhance later)

**Mitigation**:
- Focus on high-signal checks (missing %w is clear anti-pattern)
- Validate against meta-cc codebase before finalizing
- Get quick feedback on usefulness

**Contingency**:
- If linter too noisy, adjust patterns
- If linter too quiet, add more checks
- Enhancement can happen in Iteration 6

### Risk 3: Scope Creep (Linter Complexity)

**Probability**: LOW (plan explicitly limits scope)
**Impact**: HIGH (if occurs)

**Mitigation**:
- Stick to bash/grep approach (no go/analysis)
- LOC limit: 100 for linter script
- If >150 LOC, stop and defer to Iteration 6

**Trigger**: If linter script exceeds 150 LOC, defer complex features

### Risk 4: Sentinel Error Category Mismatch

**Probability**: LOW
**Impact**: MEDIUM (wrong abstraction)

**Mitigation**:
- Review error types in cmd/mcp-server before defining sentinel errors
- Ensure categories cover 80%+ of error sites
- If mismatch found, adjust categories in Phase 1

**Validation**: Phase 1 includes review of actual error patterns

---

## Expected Value Improvements

### Instance Layer (Cross-Cutting Concerns Quality)

**V_consistency** (weight: 0.4):
- Current: 0.58
- Improvement: 50+ error sites standardized, error wrapping rate 88.6% → 95%
- Expected: 0.70
- **ΔV_consistency = +0.12**

**V_maintainability** (weight: 0.3):
- Current: 0.60
- Improvement: 4 new sentinel errors, better error categorization
- Expected: 0.65
- **ΔV_maintainability = +0.05**

**V_enforcement** (weight: 0.2):
- Current: 0.15
- Improvement: Linter automates checking, CI integration
- Expected: 0.40 (still manual-ish, but automated checking)
- **ΔV_enforcement = +0.25**

**V_documentation** (weight: 0.1):
- Current: 0.80
- Improvement: Updated conventions, linter docs
- Expected: 0.85
- **ΔV_documentation = +0.05**

**V_instance(s₅) Calculation**:
```
V_instance(s₅) = 0.4×0.70 + 0.3×0.65 + 0.2×0.40 + 0.1×0.85
                = 0.28 + 0.195 + 0.08 + 0.085
                = 0.640
```

**ΔV_instance = +0.025 (+4.1%)**

### Meta Layer (Methodology Quality)

**V_completeness** (weight: 0.4):
- Current: 0.72
- Improvement: Linter creation methodology documented
- Expected: 0.77
- **ΔV_completeness = +0.05**

**V_effectiveness** (weight: 0.3):
- Current: 0.35
- Improvement: Linter automation, faster pattern checking
- Expected: 0.45
- **ΔV_effectiveness = +0.10**

**V_reusability** (weight: 0.3):
- Current: 0.52
- Improvement: Linter pattern transferable (bash/grep universal)
- Expected: 0.58
- **ΔV_reusability = +0.06**

**V_meta(s₅) Calculation**:
```
V_meta(s₅) = 0.4×0.77 + 0.3×0.45 + 0.3×0.58
            = 0.308 + 0.135 + 0.174
            = 0.617
```

**ΔV_meta = +0.072 (+13.2%)**

---

## Summary

**Iteration 5 Plan**: Conservative, achievable, focused

**Primary Objective**: Expand error standardization + basic linter automation

**Key Phases**:
1. Expand sentinel errors (4 new, 30 LOC, 30 min)
2. Standardize cmd/mcp-server errors (40-50 sites, 150 LOC, 2 hours)
3. Create simple linter script (bash/grep, 100 LOC, 1 hour)
4. Update documentation (20 LOC, 30 min)

**Total Estimated**:
- LOC: 280 (within 500 limit ✓)
- Time: 4 hours (achievable ✓)

**Expected Outcomes**:
- V_instance: 0.615 → 0.640 (+0.025, +4.1%)
- V_meta: 0.545 → 0.617 (+0.072, +13.2%)

**Agent Set**: A₅ = A₄ (no new agents, defer linter-generator)

**Risks**: Mitigated (test after each stage, scope limits, contingency plans)

**Status**: READY FOR EXECUTION

---

**Generated By**: M.plan
**Duration**: ~30 minutes
**Next**: M.execute (agent orchestration and implementation)
