# Iteration 4: Plan - Configuration Migration & Error Standardization

**Date**: 2025-10-17
**Phase**: M.plan (Planning)
**Status**: COMPLETED

---

## Iteration 4 Objective

**Goal**: Apply centralized configuration and standardize error handling across MCP server and core packages

**Scope**: Focused, high-impact implementation
- Config migration: MCP server (7 env var accesses → centralized)
- Error standardization: ~25 error sites across internal/mcp, internal/query
- Create sentinel errors package

**Expected Value Improvement**:
- V_instance: 0.465 → 0.60-0.65 (+0.135-0.185)
  - V_consistency: 0.45 → 0.60 (+0.15)
  - V_maintainability: 0.45 → 0.60 (+0.15)
- V_meta: 0.455 → 0.55-0.60 (+0.095-0.145)

---

## Work Breakdown

### Phase 1: Create Sentinel Errors Package (10 LOC)

**Agent**: coder
**Approach**: TDD (create tests first)

**File**: `internal/errors/errors.go` (NEW)

**Implementation**:
```go
package errors

import "errors"

// Sentinel errors for common conditions across meta-cc
var (
    // ErrNotFound indicates a requested resource was not found
    ErrNotFound = errors.New("not found")

    // ErrInvalidInput indicates input validation failed
    ErrInvalidInput = errors.New("invalid input")

    // ErrMissingParameter indicates a required parameter was not provided
    ErrMissingParameter = errors.New("required parameter missing")

    // ErrUnknownTool indicates an unsupported tool was requested
    ErrUnknownTool = errors.New("unknown tool")

    // ErrTimeout indicates an operation exceeded its time limit
    ErrTimeout = errors.New("operation timeout")
)
```

**Tests**: `internal/errors/errors_test.go`
- Test errors.Is compatibility
- Test error wrapping
- ~20 LOC

**Total**: ~30 LOC

---

### Phase 2: Migrate MCP Server to Centralized Config (50 LOC)

**Agent**: coder
**Approach**: Incremental migration with testing

#### Step 1: Add Config Loading to main.go (10 LOC)

**File**: `cmd/mcp-server/main.go`

**Change**:
```go
import (
    // ... existing imports
    "github.com/yaleh/meta-cc/internal/config"
)

func main() {
    // Load configuration with fail-fast validation
    cfg, err := config.Load()
    if err != nil {
        slog.Error("configuration error", "error", err)
        os.Exit(1)
    }

    // Initialize structured logging with config
    InitLogger(cfg)  // Pass config to InitLogger

    // ... rest of main
}
```

---

#### Step 2: Update logging.go to Use Config (15 LOC)

**File**: `cmd/mcp-server/logging.go`

**Change**:
```go
-// InitLogger initializes the global slog logger with configuration
-func InitLogger() {
-    // Determine log level from environment variable (default: INFO)
-    logLevel := slog.LevelInfo
-    if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
-        switch strings.ToUpper(envLevel) {
-        case "DEBUG":
-            logLevel = slog.LevelDebug
-        // ... etc
-        }
-    }
+// InitLogger initializes the global slog logger with centralized configuration
+func InitLogger(cfg *config.Config) {
+    logLevel := cfg.Log.Level

    // Create JSON handler for structured logging
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level:     logLevel,
        AddSource: true,
    })

    logger := slog.New(handler)
    slog.SetDefault(logger)
}
```

**Benefit**: ~15 lines removed (env parsing logic), cleaner code

---

#### Step 3: Update output_mode.go to Use Config (10 LOC)

**File**: `cmd/mcp-server/output_mode.go`

**Change**:
```go
-// getOutputModeConfig returns output mode configuration from parameters or environment.
-func getOutputModeConfig(params map[string]interface{}) *OutputModeConfig {
+// getOutputModeConfig returns output mode configuration from centralized config and parameters.
+func getOutputModeConfig(cfg *config.Config, params map[string]interface{}) *OutputModeConfig {
     config := DefaultOutputModeConfig()

     // Check parameter first (highest priority)
     if thresholdParam, ok := params["inline_threshold_bytes"]; ok {
         if threshold, ok := thresholdParam.(float64); ok {
             config.InlineThresholdBytes = int(threshold)
             return config
         }
     }

-    // Check environment variable
-    if envThreshold := os.Getenv("META_CC_INLINE_THRESHOLD"); envThreshold != "" {
-        if threshold, err := strconv.Atoi(envThreshold); err == nil && threshold > 0 {
-            config.InlineThresholdBytes = threshold
-            return config
-        }
-    }
+    // Use centralized config
+    config.InlineThresholdBytes = cfg.Output.InlineThreshold
+    return config

-    // Use default
-    return config
 }
```

---

#### Step 4: Update response_adapter.go (5 LOC)

**File**: `cmd/mcp-server/response_adapter.go`

**Change**: Pass cfg as parameter, use cfg.Session.SessionID, cfg.Session.ProjectHash

---

#### Step 5: Update capabilities.go (10 LOC)

**File**: `cmd/mcp-server/capabilities.go`

**Change**: Use cfg.Capability.SourcesSlice() instead of os.Getenv("META_CC_CAPABILITY_SOURCES")

---

**Total Phase 2**: ~50 LOC changes

---

### Phase 3: Standardize Errors in internal/mcp/builder.go (15 LOC)

**Agent**: coder

**File**: `internal/mcp/builder.go`

**Target Lines**: 177, 191, 220, 250, 313

**Changes**:
```go
import (
    mcerrors "github.com/yaleh/meta-cc/internal/errors"
)

// Line 177:
-return nil, fmt.Errorf("pattern parameter is required")
+return nil, fmt.Errorf("pattern parameter required for query_user_messages tool: %w", mcerrors.ErrMissingParameter)

// Line 191:
-return nil, fmt.Errorf("error_signature parameter is required")
+return nil, fmt.Errorf("error_signature parameter required for query_context tool: %w", mcerrors.ErrMissingParameter)

// Line 220:
-return nil, fmt.Errorf("file parameter is required")
+return nil, fmt.Errorf("file parameter required for query_file_access tool: %w", mcerrors.ErrMissingParameter)

// Line 250:
-return nil, fmt.Errorf("where parameter is required")
+return nil, fmt.Errorf("where parameter required for query_tools_advanced tool: %w", mcerrors.ErrMissingParameter)

// Line 313:
-return nil, fmt.Errorf("unknown tool: %s", toolName)
+return nil, fmt.Errorf("unknown tool %s in BuildToolCommand: %w", toolName, mcerrors.ErrUnknownTool)
```

**Total**: ~15 LOC changes (5 error messages improved)

---

### Phase 4: Standardize Errors in internal/query/*.go (10 LOC)

**Agent**: coder

**Files**: context.go, sequences.go, file_access.go

**Changes**:
```go
import (
    mcerrors "github.com/yaleh/meta-cc/internal/errors"
)

// query/context.go:
-return nil, fmt.Errorf("window size must be non-negative")
+return nil, fmt.Errorf("window size must be non-negative for query_context (got: %d): %w", window, mcerrors.ErrInvalidInput)

// query/sequences.go:
-return nil, fmt.Errorf("minOccurrences must be at least 1")
+return nil, fmt.Errorf("minOccurrences must be at least 1 for query_tool_sequences (got: %d): %w", minOccurrences, mcerrors.ErrInvalidInput)

// query/file_access.go:
-return nil, fmt.Errorf("file path is required")
+return nil, fmt.Errorf("file path required for query_file_access: %w", mcerrors.ErrMissingParameter)
```

**Total**: ~10 LOC changes (3 error messages improved)

---

## LOC Summary

| Phase | Component | LOC | Priority |
|-------|-----------|-----|----------|
| 1 | Sentinel errors (code + tests) | 30 | HIGH |
| 2 | MCP server config migration | 50 | HIGH |
| 3 | builder.go error standardization | 15 | HIGH |
| 4 | query/*.go error standardization | 10 | MEDIUM |
| **Total** | **All phases** | **105** | - |

**Assessment**: 105 LOC well within 500 LOC Phase limit ✅

---

## Testing Strategy

### Unit Tests

**Phase 1: Sentinel Errors**
- Test errors.Is compatibility
- Test error wrapping with fmt.Errorf
- Verify sentinel error behavior

**Phase 2: Config Migration**
- Test main.go config loading
- Test InitLogger with config
- Test output mode config resolution
- Verify backward compatibility (LOG_LEVEL fallback)

**Phase 3-4: Error Standardization**
- Test error messages contain context
- Test sentinel error wrapping
- Verify errors.Is works correctly

---

### Integration Tests

**MCP Server**:
- Start server with various env var configurations
- Verify LOG_LEVEL fallback works
- Verify META_CC_INLINE_THRESHOLD respected
- Verify session info loaded correctly

**Error Handling**:
- Test missing parameters return ErrMissingParameter
- Test unknown tools return ErrUnknownTool
- Test invalid input returns ErrInvalidInput

---

## Risk Mitigation

### Risk 1: Breaking MCP Server Startup
**Mitigation**:
- Config package already tested (100% coverage)
- Backward compatibility built-in (LOG_LEVEL fallback)
- Test MCP server startup locally before commit

### Risk 2: Test Failures
**Mitigation**:
- Run `make test` after each phase
- Fix failures immediately before proceeding
- Use TDD approach (tests first)

### Risk 3: Scope Creep
**Mitigation**:
- Strict adherence to 4 phases
- No additional error sites beyond plan
- Defer logging expansion to Iteration 5

---

## Success Criteria

### Technical Criteria
- [ ] All tests pass (`make test`)
- [ ] No lint errors (`make lint`)
- [ ] MCP server starts successfully
- [ ] Config loading works with env vars
- [ ] Sentinel errors work with errors.Is

### Metrics Criteria
- [ ] V_consistency ≥ 0.55 (from 0.45)
- [ ] V_maintainability ≥ 0.55 (from 0.45)
- [ ] V_instance ≥ 0.58 (from 0.465)
- [ ] V_meta ≥ 0.52 (from 0.455)

### Quality Criteria
- [ ] Error wrapping rate: 70% → 85%+
- [ ] Config centralization: 0% → 70%+
- [ ] Code reduction: ~20 lines removed (env parsing logic)

---

## Execution Schedule

**Phase 1** (30 min):
1. Create internal/errors/errors.go
2. Create internal/errors/errors_test.go
3. Run tests, verify passing

**Phase 2** (60 min):
1. Update main.go (config loading)
2. Update logging.go (use config)
3. Update output_mode.go (use config)
4. Update response_adapter.go, capabilities.go
5. Run tests, fix failures

**Phase 3** (20 min):
1. Update builder.go error messages
2. Run tests

**Phase 4** (20 min):
1. Update query/*.go error messages
2. Run tests

**Reflection & Documentation** (30 min):
1. Calculate V_instance(s₄), V_meta(s₄)
2. Create metrics.json
3. Create iteration-4.md

**Total Estimated Time**: ~2.5 hours

---

## Agent Selection

### Phase 1-4: coder (TDD Implementation)
- **Rationale**: Generic coder sufficient for straightforward refactoring
- **Approach**: Test-driven development
- **Specialization**: NOT needed (no complex domain logic)

### Reflection: data-analyst + M.reflect
- **Rationale**: Calculate metrics, analyze improvements
- **Approach**: Evidence-based value calculation

### Documentation: doc-writer
- **Rationale**: Create iteration-4.md report
- **Approach**: Structured documentation template

**Agent Set**: A₄ = A₃ (no new agents needed)

---

## Expected Outcomes

### Instance Layer (Cross-Cutting Concerns Quality)

**V_consistency** (0.45 → 0.60, +0.15):
- Error wrapping: 70% → 85% (+15 error sites improved)
- Config centralization: 0% → 70% (7/10 env var accesses centralized)
- Sentinel errors enable future linting

**V_maintainability** (0.45 → 0.60, +0.15):
- Single source of truth (internal/config)
- Reduced duplication (~20 lines env parsing removed)
- Centralized error definitions (internal/errors)

**V_enforcement** (0.10 → 0.15, +0.05):
- Sentinel errors prepare for linting
- Still manual enforcement (linter deferred to Iteration 5)

**V_documentation** (0.80 → 0.80, 0.00):
- Maintain current level (no regression)

**V_instance(s₄) Calculation**:
```
V_instance(s₄) = 0.4×0.60 + 0.3×0.60 + 0.2×0.15 + 0.1×0.80
                = 0.24 + 0.18 + 0.03 + 0.08
                = 0.63 (conservative estimate)
```

**Expected Range**: 0.60-0.65

---

### Meta Layer (Methodology Quality)

**V_completeness** (0.65 → 0.70, +0.05):
- Config migration pattern validated in practice
- Error standardization workflow refined
- More implementation experience

**V_effectiveness** (0.25 → 0.30, +0.05):
- Applying patterns (not just defining)
- Measuring actual benefit (code reduction, consistency improvement)

**V_reusability** (0.45 → 0.50, +0.05):
- Config pattern highly transferable
- Sentinel error pattern universal

**V_meta(s₄) Calculation**:
```
V_meta(s₄) = 0.4×0.70 + 0.3×0.30 + 0.3×0.50
            = 0.28 + 0.09 + 0.15
            = 0.52 (conservative estimate)
```

**Expected Range**: 0.50-0.55

---

## Convergence Assessment (Preliminary)

**Expected Status**: NOT CONVERGED (need more iterations)

**Analysis**:
- V_instance(s₄): ~0.63 (target: 0.80, gap: 0.17)
- V_meta(s₄): ~0.52 (target: 0.80, gap: 0.28)
- M₄ == M₃: YES (no new meta-agent capabilities)
- A₄ == A₃: YES (no new agents needed)

**Next Iteration Focus** (likely Iteration 5):
- Logging expansion (internal/parser, internal/query)
- Custom linter creation (likely trigger linter-generator agent)
- Further error standardization (cmd/mcp-server)

**Estimated Iterations to Convergence**: 2-3 more iterations

---

## Summary

**Iteration 4 Plan**: ✅ COMPLETE

**Scope**: Focused, high-impact
- Config migration: 7 env var accesses
- Error standardization: ~20 error sites
- Sentinel errors: 5 common errors

**LOC**: 105 (within limits)

**Expected Value**:
- V_instance: +0.135-0.185 (significant)
- V_meta: +0.095-0.145 (moderate)

**Next Phase**: M.execute (Implementation)

---

**Generated By**: M.plan (Meta-Agent)
**Agent**: data-analyst + doc-writer
**Status**: Ready for Execution Phase
