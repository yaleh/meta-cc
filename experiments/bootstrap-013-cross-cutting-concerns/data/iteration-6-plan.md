# Iteration 6: Implementation Plan

**Date**: 2025-10-17
**Phase**: M.plan
**Goal**: Complete error standardization + create linter automation

---

## Iteration Goal

**Primary Objective**: Achieve 80%+ error site standardization and automate enforcement

**Success Criteria**:
1. ✅ Standardize remaining ~35-40 error sites (response_adapter.go, jq_filter.go, capabilities.go top sites)
2. ✅ Create functional linter script (~100 LOC)
3. ✅ All tests passing after changes
4. ✅ V_instance(s₆) ≥ 0.58-0.62
5. ✅ V_meta(s₆) ≥ 0.54-0.58

---

## Current State Assessment

### State s₅ (from Iteration 5 partial completion)

**Value Metrics**:
- V_instance(s₅) ≈ 0.47 (estimated, based on partial completion)
- V_meta(s₅) ≈ 0.46 (estimated)

**Components**:
- V_consistency: 0.45 (error wrapping ~88%, 21 sites using sentinels)
- V_maintainability: 0.48 (config 70% centralized)
- V_enforcement: 0.15 (no automation yet)
- V_documentation: 0.80 (docs good but not updated for new errors)

**Gap to Target** (0.80):
- V_instance gap: 0.33
- V_meta gap: 0.34

**Weakest Components**:
1. **V_enforcement**: 0.15 (gap: 0.65) ← CRITICAL
2. **V_consistency**: 0.45 (gap: 0.35) ← HIGH
3. **V_maintainability**: 0.48 (gap: 0.32) ← MEDIUM

---

## Prioritized Work Plan

### Phase 1: Quick Win Files (response_adapter.go + jq_filter.go)

**Objective**: Standardize 7 error sites in high-visibility files

**Rationale**:
- Quick completion (small files, clear patterns)
- High user impact (response formatting, jq filtering)
- Builds momentum for larger files

**Files**:
1. **response_adapter.go** (4 sites, ~10 LOC)
   - Line 48: File write → ErrFileIO + file path
   - Line 55: Invalid mode → ErrInvalidInput + mode value
   - Line 72: File ref gen → ErrFileIO + file path
   - Line 98: JSON serialize → ErrParseError + context

2. **jq_filter.go** (3 sites, ~8 LOC)
   - Line 21: Invalid jq expr → ErrParseError + expression
   - Line 35: Invalid JSON → ErrParseError + line number
   - Line 62: Marshal error → ErrParseError + context

**Expected LOC**: ~18 lines (7 sites + 2 imports)

**Expected Duration**: 30 minutes

**Expected ΔV**:
- V_consistency: +0.02 (7 more sites standardized)
- V_maintainability: +0.01 (clearer error messages)

---

### Phase 2: High-Impact capabilities.go Sites (Top 25 sites)

**Objective**: Standardize highest-priority error sites in capabilities.go

**Rationale**:
- Largest concentration of errors (38 total sites)
- User-facing (capability loading/discovery)
- Focus on top 25 for token budget management

**Priority Site Selection** (25 of 38):

**A. File I/O Operations (10 sites)**:
- Lines 110, 134, 318, 405, 417, 464, 469, 475, 481, 556
- **Sentinel**: ErrFileIO
- **Context**: Add file paths, operation names

**B. Not Found Errors (6 sites)**:
- Lines 530, 749, 810, 826, 840, 1020
- **Sentinel**: ErrNotFound or ErrMissingParameter
- **Context**: Add search details, source info

**C. Parse/Validation (5 sites)**:
- Lines 269, 277, 282, 307, 563
- **Sentinel**: ErrParseError or ErrConfigError
- **Context**: Add file path, validation failure details

**D. Network Operations (4 sites)**:
- Lines 384, 394, 971, 975
- **Sentinel**: ErrNetworkFailure
- **Context**: Add URL, status code

**Deferred Sites** (13 of 38):
- Less frequently hit paths
- Lower user impact
- Can be addressed in future iteration if needed

**Expected LOC**: ~70 lines (25 sites + import + enhanced context)

**Expected Duration**: 2 hours

**Expected ΔV**:
- V_consistency: +0.08 (25 more sites standardized)
- V_maintainability: +0.05 (much clearer debugging)

---

### Phase 3: Create Simple Linter Script

**Objective**: Automate error pattern enforcement

**Rationale**:
- Addresses largest gap (V_enforcement: 0.15 → 0.60+)
- Prevents regression
- Low complexity (bash/grep, not go/analysis)

**Linter Features**:

1. **Check: Bare fmt.Errorf without %w**
   ```bash
   # Find fmt.Errorf calls without %w wrapping
   grep -rn 'fmt\.Errorf.*"[^%]*"[^%]*$' --include="*.go" | grep -v '%w'
   ```

2. **Check: Short error messages**
   ```bash
   # Find error messages <20 chars (likely lacking context)
   grep -rn 'fmt\.Errorf.*".\{1,20\}"' --include="*.go"
   ```

3. **Check: Missing mcerrors import**
   ```bash
   # Find files with errors but no sentinel import
   grep -l 'fmt\.Errorf\|errors\.New' --include="*.go" \
     | xargs grep -L 'mcerrors "github.com/yaleh/meta-cc/internal/errors"'
   ```

4. **Check: Direct errors.New usage**
   ```bash
   # Should use sentinel errors instead
   grep -rn 'errors\.New' --include="*.go"
   ```

**Output Format**:
```
FILE:LINE:SEVERITY: MESSAGE
cmd/mcp-server/foo.go:42:WARNING: fmt.Errorf without %w wrapping
cmd/mcp-server/bar.go:15:WARNING: Short error message (14 chars)
cmd/mcp-server/baz.go:1:INFO: Missing mcerrors import
```

**Script Location**: `scripts/lint-errors.sh`

**Expected LOC**: ~100 lines (4 checks + report formatting + usage help)

**Expected Duration**: 1 hour

**Expected ΔV**:
- V_enforcement: +0.45 (0.15 → 0.60, linter enables automation)
- V_maintainability: +0.05 (automated checking reduces manual effort)

---

### Phase 4: Update Knowledge Artifacts

**Objective**: Document new sentinel errors and linter usage

**Rationale**:
- Preserve methodology knowledge
- Enable reusability across projects
- Maintain V_documentation score

**Files to Update**:

1. **data/iteration-2-error-conventions.md** (~10 lines)
   - Add ErrFileIO, ErrNetworkFailure, ErrParseError, ErrConfigError
   - Update error wrapping examples

2. **knowledge/conventions/error-handling-conventions.md** (~10 lines)
   - Document linter usage
   - Add linter output interpretation guide

3. **README.md or scripts/README.md** (~5 lines)
   - Document `scripts/lint-errors.sh` usage
   - Add to development workflow

**Expected LOC**: ~25 lines total

**Expected Duration**: 30 minutes

**Expected ΔV**:
- V_documentation: +0.05 (0.80 → 0.85, comprehensive coverage)

---

## Work Breakdown

**Total Planned Phases**: 4
**Total Estimated LOC**: ~213 lines
- Phase 1: 18 LOC (error standardization)
- Phase 2: 70 LOC (capabilities.go standardization)
- Phase 3: 100 LOC (linter script)
- Phase 4: 25 LOC (documentation updates)

**Total Estimated Duration**: ~4 hours

**Scope Validation**:
- ✅ Within 500 LOC phase limit
- ✅ Focused on weakest components (V_enforcement, V_consistency)
- ✅ All phases contribute to value improvement
- ✅ No agent evolution required (coder + doc-writer sufficient)

---

## Agent Selection

### Selected Agents

1. **coder** (Phases 1-3)
   - **Tasks**: Error standardization, linter script creation
   - **Rationale**: Technical implementation, TDD approach proven effective
   - **Expected Output**: Modified files, linter script, all tests passing

2. **doc-writer** (Phase 4)
   - **Tasks**: Documentation updates, iteration report
   - **Rationale**: Knowledge preservation, clear documentation
   - **Expected Output**: Updated conventions, linter docs, iteration-6.md

3. **data-analyst** (M.reflect)
   - **Tasks**: Value calculation, metrics generation
   - **Rationale**: Honest metric assessment
   - **Expected Output**: iteration-6-metrics.json

### Agent Sufficiency Assessment

**Question**: Do we need specialized agents?

**Answer**: NO

**Evidence**:
- **coder**: Proven effective in Iterations 4-5 (TDD, error standardization)
- **Implementation complexity**: LOW-MEDIUM (pattern application, bash scripting)
- **No domain specialization needed**: Error standardization is straightforward
- **Linter complexity**: LOW (bash/grep, not go/analysis framework)

**Specialized Agent Consideration**:
- **linter-generator**: NOT NEEDED (bash script sufficient, go/analysis deferred)
- If future iterations require go/analysis linter, reassess in M.evolve

---

## Risk Assessment

### Identified Risks

1. **Risk**: capabilities.go has 38 error sites (large file, complex logic)
   - **Mitigation**: Focus on top 25 sites, test incrementally
   - **Contingency**: Defer remaining 13 sites if token budget tight

2. **Risk**: Linter false positives may require tuning
   - **Mitigation**: Start with conservative patterns, iterate based on feedback
   - **Contingency**: Mark warnings as INFO level initially

3. **Risk**: Token budget exhaustion (happened in Iteration 5)
   - **Mitigation**: Conservative scope, prioritize phases 1-3
   - **Contingency**: Defer Phase 4 documentation if needed

4. **Risk**: Test failures in capabilities.go (pre-existing issue noted in Iteration 5)
   - **Mitigation**: Check tests before/after, don't introduce new failures
   - **Contingency**: Document any pre-existing failures separately

---

## Expected Value Improvements

### Instance Layer (Cross-Cutting Concerns Quality)

**Current** (s₅):
- V_consistency: 0.45
- V_maintainability: 0.48
- V_enforcement: 0.15
- V_documentation: 0.80

**Expected** (s₆):
- V_consistency: 0.55 (+0.10, from ~32 sites standardized)
- V_maintainability: 0.58 (+0.10, from clearer errors + automation)
- V_enforcement: 0.60 (+0.45, from linter creation)
- V_documentation: 0.85 (+0.05, from docs updates)

**V_instance(s₆) Expected**:
```
V_instance(s₆) = 0.4×0.55 + 0.3×0.58 + 0.2×0.60 + 0.1×0.85
                = 0.220 + 0.174 + 0.120 + 0.085
                = 0.599 ≈ 0.60
```

**ΔV_instance**: +0.13 (+27.7% improvement)

### Meta Layer (Methodology Quality)

**Current** (s₅):
- V_completeness: 0.65
- V_effectiveness: 0.35
- V_reusability: 0.45

**Expected** (s₆):
- V_completeness: 0.72 (+0.07, from complete error + linter methodology)
- V_effectiveness: 0.45 (+0.10, from automation productivity gain)
- V_reusability: 0.52 (+0.07, from transferable linter patterns)

**V_meta(s₆) Expected**:
```
V_meta(s₆) = 0.4×0.72 + 0.3×0.45 + 0.3×0.52
            = 0.288 + 0.135 + 0.156
            = 0.579 ≈ 0.58
```

**ΔV_meta**: +0.12 (+26.1% improvement)

---

## Dependencies & Prerequisites

### Prerequisites (All Met)

1. ✅ **Sentinel errors package exists** (internal/errors/errors.go)
   - 9 sentinel errors available
   - 100% test coverage
   - All tests passing

2. ✅ **Error standardization patterns defined** (Iterations 2, 4, 5)
   - Wrapping pattern established: `fmt.Errorf("context: %w", sentinel)`
   - Test approach proven (TDD)

3. ✅ **Files identified for standardization**
   - response_adapter.go: 4 sites
   - jq_filter.go: 3 sites
   - capabilities.go: 38 sites (top 25 targeted)

4. ✅ **Build system ready**
   - `make all` validates changes
   - All tests currently passing (except pre-existing capabilities_integration_test.go issue)

### Dependencies (Sequential)

**Phase Dependencies**:
- Phase 1 → Phase 2: INDEPENDENT (can run in parallel if needed)
- Phase 2 → Phase 3: WEAK (linter can be created anytime)
- Phase 3 → Phase 4: STRONG (docs reference linter script)

**Execution Strategy**: Sequential (safer for large file changes)

---

## Plan Completion

**Planning Status**: COMPLETE ✅

**Key Decisions**:
1. **Scope**: 4 phases, ~213 LOC (conservative, achievable)
2. **Prioritization**: Enforcement gap (V_enforcement) is primary target
3. **Agent Selection**: Use existing agents (coder, doc-writer, data-analyst)
4. **Risk Mitigation**: Focus on top 25 of 38 capabilities.go sites
5. **Value Target**: V_instance(s₆) ≈ 0.60, V_meta(s₆) ≈ 0.58

**Ready for M.execute**: YES

**Next Phase**: M.execute (Phases 1-4 implementation)

---

**Generated By**: M.plan
**Date**: 2025-10-17
**Version**: 1.0
