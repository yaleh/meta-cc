# Iteration 7: Plan (Strategy Formation Phase)

**Date**: 2025-10-17
**Phase**: M.plan
**Status**: COMPLETE

---

## State Assessment

### Current State (s₆)

**Value Function**:
```yaml
V_instance(s₆) = 0.55
  V_consistency: 0.52 (28/60+ sites standardized, 47% coverage)
  V_maintainability: 0.53 (error context improved in 2 files)
  V_enforcement: 0.50 (linter exists, not CI-integrated)
  V_documentation: 0.80 (CONVERGED)

V_meta(s₆) = 0.56
  V_completeness: 0.70 (linter methodology validated)
  V_effectiveness: 0.42 (automation demonstrated)
  V_reusability: 0.52 (bash/grep patterns transferable)

Gap to Target:
  V_instance: 0.25 (need +45% improvement)
  V_meta: 0.24 (need +43% improvement)
```

**Weakest Components**:
1. **V_enforcement (0.50)**: Linter exists but not automated
2. **V_effectiveness (0.42)**: Methodology partially proven
3. **V_consistency (0.52)**: Only 47% of error sites standardized

**Critical Issues** (from observations):
1. capabilities.go: 38 error sites unstandardized (high user impact)
2. No CI enforcement (manual review required)
3. Missing sentinel errors (ErrNetworkError, ErrNotFound)
4. No documented error conventions

---

## Objectives Definition

### Primary Objective

**Complete capabilities.go error standardization + CI automation**

**Success Criteria**:
1. ✅ Top 25 error sites in capabilities.go standardized (66%+ coverage)
2. ✅ 2 new sentinel errors created (ErrNetworkError, ErrNotFound)
3. ✅ Linter integrated into Makefile + GitHub Actions CI
4. ✅ Error conventions documented
5. ✅ All tests passing, build successful
6. ✅ V_instance(s₇) ≥ 0.70-0.75 (+27-36% improvement)
7. ✅ V_meta(s₇) ≥ 0.66-0.71 (+18-27% improvement)

### Expected Value Impact

**Instance Layer**:
```yaml
V_consistency: 0.52 → 0.65 (+0.13)
  Rationale: 25 more sites standardized (53 total / ~60 files = 88% coverage)

V_maintainability: 0.53 → 0.62 (+0.09)
  Rationale: Richer context in high-visibility file (capabilities.go)

V_enforcement: 0.50 → 0.80 (+0.30)
  Rationale: CI integration completes automation

V_documentation: 0.80 → 0.85 (+0.05)
  Rationale: Error conventions + linter guide added

Expected V_instance(s₇):
  0.4 × 0.65 + 0.3 × 0.62 + 0.2 × 0.80 + 0.1 × 0.85
  = 0.260 + 0.186 + 0.160 + 0.085
  = 0.691 ≈ 0.70 (+27% from s₆)
```

**Meta Layer**:
```yaml
V_completeness: 0.70 → 0.78 (+0.08)
  Rationale: Full methodology cycle (detect → standardize → automate → document)

V_effectiveness: 0.42 → 0.55 (+0.13)
  Rationale: CI automation proves end-to-end methodology

V_reusability: 0.52 → 0.60 (+0.08)
  Rationale: Documented patterns transferable across projects

Expected V_meta(s₇):
  0.4 × 0.78 + 0.3 × 0.55 + 0.3 × 0.60
  = 0.312 + 0.165 + 0.180
  = 0.657 ≈ 0.66 (+18% from s₆)
```

**Total Expected Improvement**:
- ΔV_instance: +0.15 (+27%)
- ΔV_meta: +0.10 (+18%)
- Combined: Strong progress toward convergence

---

## Work Breakdown

### Task 1: Create New Sentinel Errors

**Agent**: coder
**Inputs**: internal/errors/errors.go, error pattern analysis
**Outputs**: 2 new sentinel errors

**Work**:
1. Add `ErrNetworkError` sentinel (network operations, HTTP errors)
2. Add `ErrNotFound` sentinel (capability/file not found)
3. Update comments/documentation in errors.go
4. **Expected LOC**: ~8 lines

**Rationale**: Required before capabilities.go standardization

**Dependencies**: None (standalone)

**Expected Duration**: ~10 minutes

---

### Task 2: Standardize capabilities.go Error Sites (Top 25)

**Agent**: coder
**Inputs**: capabilities.go, sentinel errors, observation analysis
**Outputs**: capabilities.go with 25 standardized error sites

**Work**:
1. Add mcerrors import alias
2. Standardize **Priority 1** (User-Facing Errors, 8 sites):
   - Lines 810, 840: Capability not found → ErrNotFound
   - Line 1020: Parameter validation → ErrInvalidInput
   - Line 311: Path not directory → ErrInvalidInput
   - Lines 898, 913: GitHub format → ErrParseError
   - Line 530: Package not found → ErrNotFound
   - Line 1009: Enhanced error → ErrNotFound

3. Standardize **Priority 2** (Network Operations, 4 sites):
   - Line 384: Download failed → ErrNetworkError
   - Line 394: Status code → ErrNetworkError
   - Line 971, 975: jsDelivr errors → ErrNetworkError

4. Standardize **Priority 3** (File I/O, 10 sites):
   - Line 307: Path access → ErrFileIO
   - Line 110: Cache creation → ErrFileIO
   - Line 134: Cache cleanup → ErrFileIO
   - Lines 433, 439, 454: Archive operations → ErrFileIO
   - Lines 464, 469, 475, 481: File operations → ErrFileIO

5. Standardize **Priority 4** (Parse Errors, 3 sites):
   - Line 277: YAML parse → ErrParseError
   - Line 282: Missing field → ErrParseError
   - Line 269: No frontmatter → ErrParseError

**Expected LOC**: ~70 lines (25 sites × ~2.8 lines/site average)

**Rationale**: Highest user impact file, 66% coverage meets target

**Dependencies**: Task 1 (sentinel errors)

**Expected Duration**: ~45-60 minutes

**Validation**: Build + linter passes

---

### Task 3: Integrate Linter into Makefile

**Agent**: coder
**Inputs**: Makefile, scripts/lint-errors.sh
**Outputs**: Updated Makefile with lint-errors target

**Work**:
1. Add `lint-errors` target:
   ```makefile
   .PHONY: lint-errors
   lint-errors:
   	@echo "Running error linter..."
   	@./scripts/lint-errors.sh cmd/ internal/
   ```

2. Add to `all` or `lint` target:
   ```makefile
   lint: lint-errors lint-go
   ```

**Expected LOC**: ~5 lines

**Rationale**: Enables `make lint-errors` for local development

**Dependencies**: None

**Expected Duration**: ~5 minutes

---

### Task 4: Create GitHub Actions Workflow

**Agent**: coder
**Inputs**: .github/workflows/test.yml (reference), lint-errors target
**Outputs**: .github/workflows/error-linting.yml

**Work**:
1. Create error-linting.yml:
   ```yaml
   name: Error Linting

   on:
     push:
       branches: [ main, develop ]
     pull_request:

   jobs:
     lint-errors:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         - name: Run error linter
           run: make lint-errors
   ```

**Expected LOC**: ~15 lines

**Rationale**: Automates enforcement on every PR/push

**Dependencies**: Task 3 (Makefile target)

**Expected Duration**: ~10 minutes

---

### Task 5: Document Error Conventions

**Agent**: doc-writer
**Inputs**: Standardization patterns, sentinel errors, linter usage
**Outputs**: knowledge/best-practices/error-handling.md

**Work**:
1. Sentinel error usage guide:
   - When to use each sentinel
   - How to wrap with context
   - Examples from codebase

2. Error wrapping patterns:
   - Always use %w for wrapping
   - Add operation context
   - Add resource identifiers (paths, URLs, names)

3. Context enrichment examples:
   - File I/O: Add file paths
   - Network: Add URLs + status codes
   - Parse: Add input strings + line numbers
   - Validation: Add parameter names + expected values

4. Linter usage:
   - How to run locally
   - How to interpret warnings
   - How to fix common issues
   - CI integration status

**Expected LOC**: ~80 lines markdown

**Rationale**: Preserves methodology knowledge, guides contributors

**Dependencies**: Tasks 1-4 (patterns established)

**Expected Duration**: ~20 minutes

---

### Task 6: Update Iteration Documentation

**Agent**: doc-writer
**Inputs**: All task outputs, metrics
**Outputs**: iteration-7.md report

**Work**:
1. Document all changes made
2. Calculate honest metrics (V_instance, V_meta)
3. Perform convergence check
4. Identify remaining gaps
5. Recommend next iteration focus (if needed)

**Expected LOC**: ~800-1000 lines markdown

**Rationale**: Tracks experiment progress and methodology

**Dependencies**: Tasks 1-5 (all work complete)

**Expected Duration**: ~30 minutes

---

## Agent Selection

### Selected Agents

**Agent Set (A₇ = A₆)**:
1. **coder**: Tasks 1-4 (implementation work)
2. **doc-writer**: Tasks 5-6 (documentation)
3. **data-analyst**: Metrics calculation (integrated into doc-writer)

**Rationale**:
- **coder**: Well-suited for error standardization (proven in Iterations 5-6)
- **doc-writer**: Effective for structured documentation (proven in all iterations)
- **No new agents needed**: Tasks within existing capabilities

### Specialization Assessment

**Question**: Do we need specialized agents?

**Answer**: NO

**Evidence**:
- Error standardization: Pattern application (coder effective)
- CI integration: Simple YAML + Makefile (coder effective)
- Documentation: Structured writing (doc-writer effective)
- No complex domain knowledge required
- Generic agents have 100% success rate in this experiment

**Conclusion**: A₇ = A₆ (no evolution)

---

## Task Sequencing

### Sequential Dependencies

```mermaid
Task 1 (Sentinel Errors)
  ↓
Task 2 (Standardize capabilities.go) ← depends on Task 1
  ↓
Task 3 (Makefile) ← standalone, but logically after Task 2
  ↓
Task 4 (GitHub Actions) ← depends on Task 3
  ↓
Task 5 (Documentation) ← depends on Tasks 1-4 (patterns established)
  ↓
Task 6 (Iteration Report) ← depends on all tasks
```

**Execution Order**:
1. Task 1 (sentinel errors) - prerequisite
2. Task 2 (standardize) - main work
3. Task 3 (Makefile) - CI preparation
4. Task 4 (GitHub Actions) - CI automation
5. Task 5 (conventions doc) - knowledge preservation
6. Task 6 (iteration report) - final documentation

**Parallelization Opportunities**: None (sequential dependencies)

---

## Risks and Mitigations

### Risk 1: Token Budget Exhaustion

**Probability**: MEDIUM (occurred in Iterations 5-6)
**Impact**: HIGH (incomplete work, deferred objectives)

**Mitigation**:
- Prioritize high-value work first (Tasks 1-2-3-4)
- Task 5 (documentation) can be deferred if needed
- Task 6 (iteration report) always completed (experiment protocol)
- Conservative scoping: 25 sites (not all 38)

**Contingency**: Defer Task 5 to Iteration 8 if budget tight

---

### Risk 2: Build Failures

**Probability**: LOW (TDD approach working well)
**Impact**: MEDIUM (blocks progress, debugging time)

**Mitigation**:
- Run `go build` after each task
- Run `make test` before finalizing
- Use linter to validate changes
- Small, incremental changes

**Contingency**: Revert last change, debug, retry

---

### Risk 3: Linter False Positives

**Probability**: LOW (linter validated in Iteration 6)
**Impact**: LOW (CI noise, contributor friction)

**Mitigation**:
- Test linter on full codebase before CI integration
- Document how to interpret warnings
- Start with advisory mode (warnings, not errors)

**Contingency**: Refine linter patterns, add exclusions

---

### Risk 4: Pre-existing Test Failures

**Probability**: HIGH (parser_test.go failing in Iteration 6)
**Impact**: LOW (not our regression, build succeeds)

**Mitigation**:
- Document pre-existing failures
- Verify no new test failures introduced
- Fix pre-existing failures in separate iteration (technical debt)

**Contingency**: Acknowledge in iteration report, continue

---

## Success Metrics

### Instance Layer Targets

```yaml
V_consistency ≥ 0.65:
  Measure: (standardized_sites / total_sites_in_scope)
  Target: 53 / 60 = 88% coverage
  Method: Count grep -r "mcerrors" cmd/ internal/

V_maintainability ≥ 0.62:
  Measure: Error context richness in capabilities.go
  Target: All errors have paths/URLs/names
  Method: Manual review of 25 standardized sites

V_enforcement ≥ 0.80:
  Measure: CI integration completeness
  Target: Makefile target + GitHub Actions workflow
  Method: Verify make lint-errors works + CI runs

V_documentation ≥ 0.85:
  Measure: Error conventions coverage
  Target: All patterns documented with examples
  Method: knowledge/best-practices/error-handling.md exists
```

### Meta Layer Targets

```yaml
V_completeness ≥ 0.78:
  Measure: Methodology lifecycle completion
  Target: Detect → Standardize → Automate → Document
  Method: All phases evidenced in artifacts

V_effectiveness ≥ 0.55:
  Measure: Automation productivity gain
  Target: CI prevents regressions automatically
  Method: Linter catches issues in CI

V_reusability ≥ 0.60:
  Measure: Pattern transferability
  Target: Documentation enables other projects
  Method: Error conventions doc usable outside meta-cc
```

---

## Expected Outcomes

### Artifacts

1. **internal/errors/errors.go** (+8 LOC, 2 new sentinels)
2. **cmd/mcp-server/capabilities.go** (~70 LOC changed, 25 sites)
3. **Makefile** (+5 LOC, lint-errors target)
4. **.github/workflows/error-linting.yml** (+15 LOC, new file)
5. **knowledge/best-practices/error-handling.md** (+80 LOC, new file)
6. **iteration-7.md** (~900 LOC, final report)

**Total LOC**: ~180 lines (within phase limits)

### Value Improvements

**Instance Layer**: +0.15 (+27%)
**Meta Layer**: +0.10 (+18%)
**Combined**: Approaching convergence threshold (V ≥ 0.80)

### Convergence Trajectory

**After Iteration 7**:
- V_instance(s₇) ≈ 0.70 (gap: 0.10)
- V_meta(s₇) ≈ 0.66 (gap: 0.14)
- Estimated iterations to convergence: 1-2 more

**Key Gaps Remaining** (for Iteration 8):
- V_consistency: 12% coverage gap (remaining 7 sites in capabilities.go)
- V_effectiveness: 25% gap (methodology validation in diverse files)
- V_reusability: 20% gap (cross-project testing)

---

## Execution Readiness

**Prerequisites**: ✅ All met
- Linter script exists (scripts/lint-errors.sh)
- Sentinel errors framework exists (internal/errors/errors.go)
- Build environment ready (go, make, git)
- Observation data complete (iteration-7-observations.md)

**Agent Availability**: ✅ All ready
- coder: Available for implementation
- doc-writer: Available for documentation
- data-analyst: Available for metrics

**Resources**: ✅ All available
- Token budget: ~140K remaining (sufficient)
- Time estimate: ~2-3 hours total
- Tools: All present and validated

**Proceed to M.execute**: ✅ READY

---

**Document Version**: 1.0
**Created**: 2025-10-17
**Generated By**: M.plan (Meta-Agent)
