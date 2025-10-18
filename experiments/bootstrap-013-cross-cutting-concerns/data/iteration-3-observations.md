# Iteration 3 Observations: Implementation Opportunities

**Date**: 2025-10-17
**Phase**: M.observe
**Focus**: Analyzing codebase for cross-cutting concerns implementation

---

## Observation Summary

### Error Handling Analysis

**Current State**:
- **fmt.Errorf usage**: 218 occurrences across cmd/ and internal/
- **Error wrapping with %w**: ~70% (majority already using correct pattern)
- **errors.New usage**: Minimal (mostly in tests)
- **Custom error types**: Present in internal/output/error.go (ErrorCode enum)

**Patterns Found**:

1. **Good Existing Patterns** (cmd/stats_aggregate.go, cmd/query_context.go):
   ```go
   return fmt.Errorf("failed to locate session: %w", err)
   return fmt.Errorf("invalid filter: %w", err)
   return fmt.Errorf("filter evaluation error: %w", err)
   ```
   - ✅ Uses %w for wrapping
   - ✅ Provides operation context
   - ⚠️ Could add more entity identifiers

2. **Areas for Improvement**:
   - Missing sentinel errors (no package-level error constants)
   - Inconsistent context depth (some errors lack file/line info)
   - No error classification (could benefit from Iteration 2's classifyError pattern)
   - Missing entity identifiers in some wrappers

**Implementation Targets**:
- **cmd/** package: ~100 error sites (standardize wrapping, add sentinel errors)
- **internal/parser/**: ~30 error sites (add context)
- **internal/query/**: ~25 error sites (add context)
- **internal/locator/**: ~20 error sites (add sentinels)

**Estimated Impact**: 30-40% of error handling can be improved in this iteration

---

### Configuration Management Analysis

**Current State**:
- **os.Getenv usage**: 15 occurrences across cmd/ and internal/
- **Centralized config**: None (scattered access)
- **Naming consistency**: 50% (mixed LOG_LEVEL vs META_CC_*)
- **Validation**: Partial (some env vars checked, others not)

**Environment Variables Detected**:
1. `LOG_LEVEL` (MCP server) - deprecated form
2. `META_CC_LOG_LEVEL` (proposed convention)
3. `META_CC_LOG_FORMAT` (proposed)
4. `META_CC_CAPABILITY_SOURCES` (MCP server)
5. `META_CC_INLINE_THRESHOLD` (MCP server)
6. `CC_SESSION_ID` (Claude Code)
7. `CC_PROJECT_HASH` (Claude Code)

**Implementation Targets**:
- Create `internal/config/` package
- Define Config struct with nested sections (Log, Output, Capability, Session)
- Migrate 8-10 high-priority env var accesses
- Add validation at startup
- Provide defaults for all non-sensitive values

**Estimated Impact**: 50% of configuration centralized in this iteration

---

### Logging Coverage Analysis

**Current State**:
- **MCP server logging**: 51 log statements (excellent coverage)
- **cmd/ logging**: ~5% coverage (minimal)
- **internal/parser/ logging**: 0% coverage
- **internal/query/ logging**: 0% coverage
- **internal/analyzer/ logging**: 0% coverage

**External Validation** (from Iteration 2):
- MCP server implemented logging between iterations
- 90% adherence to Iteration 1 conventions
- Validates convention quality and practicality

**Implementation Targets**:
- Add logging to internal/parser/ (parse operations, errors)
- Add logging to internal/query/ (query execution, results)
- Add logging to internal/analyzer/ (analysis operations)
- Target: 20-30 new log statements (5% → 15-20% coverage)

**Estimated Impact**: 3x increase in logging coverage

---

### Automation Opportunities

**Linter Requirements** (from Iteration 2 conventions):

1. **Error Wrapping Linter**:
   - Detect `fmt.Errorf` without `%w` (breaks unwrapping)
   - Detect `return err` without wrapping (missing context)
   - Detect insufficient error context (operation, entity, identifiers)
   - Suggest fixes automatically

2. **Error Logging Linter**:
   - Detect log-and-throw anti-pattern
   - Detect errors logged without error_type classification
   - Ensure errors logged at appropriate level

3. **Configuration Linter**:
   - Detect scattered `os.Getenv` calls (should use centralized config)
   - Detect env vars without META_CC_ prefix
   - Detect missing validation for required vars

**Implementation Approach**:
- Use go/analysis framework (standard library)
- Create custom analyzer in `tools/linters/`
- Integration with golangci-lint
- Provide auto-fix suggestions where possible

**Estimated Coverage**: Linter can detect 40-50% of anti-patterns

---

## Implementation Priorities

### Priority 1: Error Handling Standardization (High Impact)

**Why**:
- 218 error sites exist (large surface area)
- 70% already use %w (good foundation)
- Missing sentinel errors cause poor error handling
- Improves V_consistency directly

**Scope**:
- Standardize ~50 error wrapping sites in cmd/
- Add 5-10 sentinel errors per package
- Improve error context (add file paths, line numbers, IDs)

**Expected ΔV_consistency**: +0.10-0.15 (45% → 55-60%)

---

### Priority 2: Centralized Configuration (Medium-High Impact)

**Why**:
- 15 scattered os.Getenv calls (maintenance burden)
- No validation (runtime failures)
- Inconsistent naming (50% adherence)
- Improves V_maintainability directly

**Scope**:
- Create internal/config/ package
- Migrate 8-10 high-priority env vars
- Add validation with helpful error messages
- Provide sensible defaults

**Expected ΔV_maintainability**: +0.10-0.15 (40% → 50-55%)

---

### Priority 3: Logging Expansion (Medium Impact)

**Why**:
- MCP server shows 90% convention adherence (validates quality)
- Critical packages (parser, query) have 0% coverage
- Improved observability for debugging
- Incremental value (can expand further in Iteration 4)

**Scope**:
- Add logging to internal/parser/ (~10 log sites)
- Add logging to internal/query/ (~10 log sites)
- Add logging to internal/analyzer/ (~5 log sites)

**Expected ΔV_documentation**: +0.05 (80% → 85%, exceeds target)

---

### Priority 4: Custom Linter Creation (Automation Foundation)

**Why**:
- Automation prevents regressions
- Enforces conventions in CI/CD
- Improves V_enforcement significantly
- Foundation for future iterations

**Scope**:
- Error wrapping analyzer (detect missing %w)
- Error context analyzer (detect insufficient context)
- Log-and-throw detector
- Integration with golangci-lint

**Expected ΔV_enforcement**: +0.30-0.40 (10% → 40-50%)

---

## Data Collection Complete

**Metrics**:
- Error sites analyzed: 218
- Configuration sites analyzed: 15
- Logging sites analyzed: 51 (MCP server) + ~5 (cmd/)
- Packages analyzed: 14 (internal/) + 1 (cmd/)

**Patterns Identified**:
- Error wrapping: 70% good, 30% needs improvement
- Configuration: 0% centralized, 50% naming consistent
- Logging: MCP server excellent (90% adherence), other packages minimal

**Gaps Found**:
- No sentinel errors defined
- No centralized configuration
- Minimal logging outside MCP server
- No automated enforcement (linters)

**Quality**: Observations based on actual code analysis, reproducible

---

**Status**: COMPLETE
**Next Phase**: M.plan (define Iteration 3 goals and agent selection)
**Generated By**: M.observe (meta-agent)
