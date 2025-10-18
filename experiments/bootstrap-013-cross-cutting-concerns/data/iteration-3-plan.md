# Iteration 3 Plan: Implementation Phase + Linter Creation

**Date**: 2025-10-17
**Phase**: M.plan
**Based On**: iteration-3-observations.md

---

## State Assessment

### Current State (s₂)

**V_instance(s₂)** = 0.45
- V_consistency: 0.45 (target: 0.80, gap: 0.35)
- V_maintainability: 0.40 (target: 0.80, gap: 0.40)
- V_enforcement: 0.10 (target: 0.80, gap: 0.70)
- V_documentation: 0.80 (target: 0.80, **CONVERGED**)

**V_meta(s₂)** = 0.40
- V_completeness: 0.60 (target: 0.80, gap: 0.20)
- V_effectiveness: 0.20 (target: 0.80, gap: 0.60)
- V_reusability: 0.40 (target: 0.80, gap: 0.40)

**System State**:
- M₂ = M₁ (STABLE, all 5 capabilities sufficient)
- A₂ = A₁ (STABLE, 4 agents: data-analyst, doc-writer, coder, convention-definer)
- Conventions: 100% complete (3/3: logging, error-handling, configuration)
- Implementation: 5% (only MCP server logging, external)

**Weakest Components**:
1. V_enforcement (0.10) - no automation
2. V_maintainability (0.40) - scattered config, inconsistent patterns
3. V_consistency (0.45) - partial standardization

---

## Iteration 3 Goal

### Primary Objective

**Implement cross-cutting conventions and create automated enforcement**

### Success Criteria

1. **Error Handling Standardization**:
   - ✅ 50+ error sites improved (add context, use sentinels)
   - ✅ 10+ sentinel errors defined
   - ✅ All cmd/ errors have proper context

2. **Centralized Configuration**:
   - ✅ internal/config/ package created
   - ✅ Config struct with validation
   - ✅ 8-10 env vars migrated
   - ✅ Fail-fast startup validation

3. **Logging Expansion**:
   - ✅ internal/parser/ has logging
   - ✅ internal/query/ has logging
   - ✅ 20-30 new log statements added
   - ✅ 15-20% logging coverage achieved

4. **Custom Linter Creation**:
   - ✅ Error wrapping analyzer implemented
   - ✅ Error context analyzer implemented
   - ✅ Integrated with golangci-lint or standalone
   - ✅ CI/CD integration documented

### Expected ΔV

**Instance Layer**:
- V_consistency: 0.45 → 0.60 (+0.15, from error standardization)
- V_maintainability: 0.40 → 0.55 (+0.15, from centralized config)
- V_enforcement: 0.10 → 0.45 (+0.35, from linter creation)
- V_documentation: 0.80 → 0.85 (+0.05, from expanded logging)

**V_instance(s₃)**: 0.45 → 0.62 (+0.17, +37.8%)

**Meta Layer**:
- V_completeness: 0.60 → 0.70 (+0.10, implementation validates methodology)
- V_effectiveness: 0.20 → 0.40 (+0.20, linter proves automation value)
- V_reusability: 0.40 → 0.50 (+0.10, patterns refined through implementation)

**V_meta(s₃)**: 0.40 → 0.54 (+0.14, +35.0%)

---

## Agent Selection

### Decision: Use Existing Agents (No Specialization Needed)

**Rationale**:
- Work is well-defined technical implementation (not domain-specific)
- Conventions already defined (clear requirements)
- Templates available (copy-paste ready)
- Generic agents have necessary capabilities

### Agent Assignment

| Agent | Tasks | Rationale |
|-------|-------|-----------|
| **coder** | Implement all code changes | Primary implementation agent, has capability for Go development |
| **data-analyst** | Calculate metrics, measure coverage | Quantify improvements, track statistics |
| **doc-writer** | Create iteration report | Document results, process, learnings |

**Agents NOT Used**:
- convention-definer: No new conventions needed (all defined in Iteration 2)

---

## Work Breakdown

### Phase 1: Error Handling Standardization

**Agent**: coder
**Inputs**:
- iteration-2-error-conventions.md
- error-handling-template.go
- Codebase analysis from observations

**Tasks**:
1. Create sentinel errors in cmd/ package
   - ErrSessionNotFound
   - ErrInvalidFilter
   - ErrParseError
   - ErrOutputError

2. Standardize error wrapping in cmd/ (~50 sites)
   - Add file paths to parse errors
   - Add filter expressions to filter errors
   - Add session IDs where applicable

3. Add sentinel errors in internal/parser/
   - ErrInvalidJSONL
   - ErrMalformedRecord

4. Add sentinel errors in internal/query/
   - ErrEmptyResult
   - ErrInvalidQuery

**Outputs**:
- Modified cmd/ files with improved error handling
- Modified internal/parser/ files with sentinel errors
- Modified internal/query/ files with sentinel errors

**Estimated LOC**: ~200 lines changed (within Phase limit: 500)

---

### Phase 2: Centralized Configuration

**Agent**: coder
**Inputs**:
- iteration-2-config-conventions.md
- config-management-template.go

**Tasks**:
1. Create internal/config/ package structure
   - config.go (Config struct, Load(), Validate())
   - config_test.go (validation tests, default tests)

2. Define Config struct with sections:
   ```go
   type Config struct {
       Log        LogConfig
       Output     OutputConfig
       Capability CapabilityConfig
       Session    SessionConfig
   }
   ```

3. Implement validation with helpful errors

4. Migrate high-priority env var access:
   - cmd/mcp-server/logging.go (LOG_LEVEL → config)
   - cmd/mcp-server/capabilities.go (CAPABILITY_SOURCES → config)
   - internal/mcp/ hybrid output (INLINE_THRESHOLD → config)

**Outputs**:
- internal/config/config.go (~300 lines)
- internal/config/config_test.go (~150 lines)
- Modified cmd/mcp-server/main.go (use config.Load())
- Modified cmd/mcp-server/logging.go (use config)

**Estimated LOC**: ~500 lines new + 50 lines modified (within Phase limit)

---

### Phase 3: Logging Expansion

**Agent**: coder
**Inputs**:
- iteration-1-logging-conventions.md
- logging-template.go (from Iteration 1)
- MCP server logging implementation (reference)

**Tasks**:
1. Add logging to internal/parser/
   - INFO: Parse start/complete
   - DEBUG: Line-by-line parsing
   - ERROR: Parse failures

2. Add logging to internal/query/
   - INFO: Query execution start/complete
   - DEBUG: Filter application
   - WARN: Empty results
   - ERROR: Query failures

3. Add logging to internal/analyzer/
   - INFO: Analysis operations
   - ERROR: Analysis failures

**Outputs**:
- Modified internal/parser/reader.go (~10 new log statements)
- Modified internal/query/ files (~10 new log statements)
- Modified internal/analyzer/ files (~5 new log statements)

**Estimated LOC**: ~50 lines added (within Stage limit: 200)

---

### Phase 4: Custom Linter Creation

**Agent**: coder
**Inputs**:
- iteration-2-error-conventions.md (anti-patterns section)
- go/analysis framework documentation

**Tasks**:
1. Create tools/linters/ package structure

2. Implement error wrapping analyzer:
   ```go
   // Detects:
   // - fmt.Errorf without %w
   // - return err without wrapping
   // - insufficient error context
   ```

3. Implement log-and-throw detector:
   ```go
   // Detects:
   // - log.Error() followed by return err
   ```

4. Create integration:
   - Standalone binary (tools/linters/main.go)
   - golangci-lint plugin OR custom check
   - CI/CD integration documentation

5. Test linter on codebase, measure detection rate

**Outputs**:
- tools/linters/errwrap/analyzer.go (~200 lines)
- tools/linters/logthrow/analyzer.go (~150 lines)
- tools/linters/main.go (CLI wrapper, ~100 lines)
- tools/linters/README.md (usage documentation)

**Estimated LOC**: ~500 lines (within Phase limit)

**Note**: If linter complexity exceeds capabilities, consider creating **linter-generator** specialized agent

---

### Phase 5: Metrics and Documentation

**Agent**: data-analyst
**Inputs**:
- Implementation results
- Before/after code samples
- Linter detection results

**Tasks**:
1. Calculate implementation statistics:
   - Error sites improved (count, percentage)
   - Config vars centralized (count, percentage)
   - Logging statements added (count, coverage percentage)
   - Linter detection rate (anti-patterns found)

2. Calculate V_instance(s₃) components:
   - V_consistency (measure pattern adherence)
   - V_maintainability (assess centralization)
   - V_enforcement (measure linter coverage)
   - V_documentation (calculate logging coverage)

3. Calculate V_meta(s₃) components:
   - V_completeness (implementation completeness)
   - V_effectiveness (automation value)
   - V_reusability (pattern refinement)

**Outputs**:
- data/iteration-3-metrics.json
- data/iteration-3-implementation-summary.yaml

**Estimated Output**: ~300 lines

---

**Agent**: doc-writer
**Inputs**:
- All phase outputs
- Metrics from data-analyst
- Observations and plan documents

**Tasks**:
1. Document iteration-3.md (full iteration report)
2. Update knowledge/INDEX.md (if new patterns extracted)

**Outputs**:
- iteration-3.md (~1000 lines)

---

## Dependencies

```
Phase 1 (Error Handling) ──┐
Phase 2 (Configuration)   ──┼─→ Phase 5 (Metrics) ──→ Phase 5 (Documentation)
Phase 3 (Logging)         ──┤
Phase 4 (Linter)          ──┘
```

**Execution Strategy**:
- Phases 1-4: Can run in parallel (independent)
- Phase 5 (Metrics): Depends on Phases 1-4 completion
- Phase 5 (Documentation): Depends on Metrics

**Recommended Order**:
1. Phase 2 (Configuration) - Foundation for others
2. Phase 1 (Error Handling) - Large impact
3. Phase 3 (Logging) - Quick wins
4. Phase 4 (Linter) - Automation
5. Phase 5 (Metrics + Documentation)

---

## Risks and Mitigations

### Risk 1: Linter Complexity

**Risk**: Custom linter may exceed coder capabilities or time budget

**Mitigation**:
- Start with simple analyzers (error wrapping only)
- If complexity high, trigger M.evolve for linter-generator agent
- Acceptable to defer advanced linters to Iteration 4

**Fallback**: Manual code review with checklist (reduces V_enforcement but still provides value)

---

### Risk 2: Breaking Changes

**Risk**: Centralized config may break existing code

**Mitigation**:
- Run full test suite after config migration
- Provide backward compatibility for one iteration
- Clear migration documentation

**Fallback**: Partial migration (MCP server only), expand in Iteration 4

---

### Risk 3: Scope Creep

**Risk**: Implementation may reveal additional work beyond plan

**Mitigation**:
- Strict adherence to LOC limits (500/phase, 200/stage)
- Defer non-critical improvements to Iteration 4
- Focus on measurable ΔV targets

**Fallback**: Document deferred items for next iteration

---

## Success Metrics

### Quantitative Targets

- Error sites improved: ≥ 50 (target: 23% of 218)
- Sentinel errors created: ≥ 10
- Config vars centralized: ≥ 8 (target: 53% of 15)
- Logging statements added: ≥ 20
- Linter detection rate: ≥ 40% of anti-patterns

### Value Targets

- V_instance(s₃): ≥ 0.60 (stretch: 0.65)
- V_meta(s₃): ≥ 0.50 (stretch: 0.55)
- ΔV_instance: ≥ +0.15
- ΔV_meta: ≥ +0.10

### Qualitative Targets

- All tests pass
- CI/CD pipeline green
- Code reviewable (follows conventions)
- Documentation complete

---

## Plan Summary

**Iteration Goal**: Implement conventions + create automation

**Agent Set**: A₃ = A₂ (no new agents needed)

**Meta-Agent**: M₃ = M₂ (no new capabilities needed)

**Work Phases**: 5 (Error, Config, Logging, Linter, Metrics/Doc)

**Expected ΔV_instance**: +0.17 (+37.8%)

**Expected ΔV_meta**: +0.14 (+35.0%)

**Estimated LOC**: ~1300 lines (within experiment constraints)

**Risks**: Linter complexity (mitigated with fallback)

---

**Status**: COMPLETE
**Next Phase**: M.execute (implementation)
**Generated By**: M.plan (meta-agent)
