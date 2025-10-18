# Iteration 4: Observations - Configuration Migration & Error Standardization Opportunities

**Date**: 2025-10-17
**Phase**: M.observe (Observation)
**Status**: COMPLETED

---

## Executive Summary

Analysis reveals **10 os.Getenv sites** ready for config migration and **25+ error sites** needing standardization. MCP server is the primary target with 3 config files (logging.go, output_mode.go, main.go) and ~15 error sites lacking consistent wrapping.

**Key Findings**:
- ✅ Config package ready (589 LOC, 100% test coverage)
- ✅ Error conventions defined (Iteration 2)
- 🎯 **10 os.Getenv calls** to migrate (MCP server: 7, other: 3)
- 🎯 **25+ error sites** need %w wrapping improvement
- 🎯 **3 sentinel errors** to create (ErrNotFound, ErrInvalidInput, ErrTimeout)

**Effort Estimation**:
- Config migration: ~50-75 LOC changes
- Error standardization: ~30-40 error sites
- Total: **80-115 LOC** (well within Phase limit)

---

## Configuration Migration Opportunities

### Current State Analysis

**Total os.Getenv Calls**: 12 total (10 non-test)

**Distribution**:
- **cmd/mcp-server/**: 7 calls (HIGH priority)
  - `logging.go`: LOG_LEVEL parsing (line 19)
  - `output_mode.go`: META_CC_INLINE_THRESHOLD (line 142)
  - `response_adapter.go`: CC_SESSION_ID (line 107), CC_PROJECT_HASH (line 116)
  - `capabilities.go`: CLAUDE_CODE_SESSION_ID (line 85), META_CC_CAPABILITY_SOURCES (lines 713, 1024)

- **Test files**: 2 calls (LOW priority - testing)
  - `capabilities_test.go`: META_CC_CAPABILITY_SOURCES
  - `executor_test.go`: CC_SESSION_ID, CC_PROJECT_HASH

### Migration Targets (Priority Order)

#### 1. MCP Server Logging (HIGH - Complete Migration)

**File**: `cmd/mcp-server/logging.go`

**Current** (lines 18-30):
```go
logLevel := slog.LevelInfo
if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
    switch strings.ToUpper(envLevel) {
    case "DEBUG":
        logLevel = slog.LevelDebug
    // ... etc
    }
}
```

**Target**:
```go
cfg, err := config.Load()
if err != nil {
    // Fail-fast on config error
    slog.Error("configuration error", "error", err)
    os.Exit(1)
}

logLevel := cfg.Log.Level
```

**Benefits**:
- ✅ Validation included (no invalid levels)
- ✅ Backward compatible (LOG_LEVEL fallback in config)
- ✅ Single source of truth
- ✅ Reduced code: ~15 lines → ~5 lines

---

#### 2. Output Mode Configuration (HIGH)

**File**: `cmd/mcp-server/output_mode.go`

**Current** (lines 141-147):
```go
if envThreshold := os.Getenv("META_CC_INLINE_THRESHOLD"); envThreshold != "" {
    if threshold, err := strconv.Atoi(envThreshold); err == nil && threshold > 0 {
        config.InlineThresholdBytes = threshold
        return config
    }
}
```

**Target**:
```go
// Use config from centralized config.Load()
// Config already includes InlineThreshold with validation
```

**Benefits**:
- ✅ Type conversion handled
- ✅ Validation included (positive integers only)
- ✅ Centralized defaults

---

#### 3. Session Configuration (MEDIUM)

**File**: `cmd/mcp-server/response_adapter.go`

**Current** (lines 107, 116):
```go
if sessionID := os.Getenv("CC_SESSION_ID"); sessionID != "" {
    // ... use sessionID
}

if projectHash := os.Getenv("CC_PROJECT_HASH"); projectHash != "" {
    // ... use projectHash
}
```

**Target**:
```go
// From global cfg
sessionID := cfg.Session.SessionID
projectHash := cfg.Session.ProjectHash
```

**Benefits**:
- ✅ Single read point
- ✅ Available throughout application
- ✅ Testable (can inject config)

---

#### 4. Capability Sources (MEDIUM)

**File**: `cmd/mcp-server/capabilities.go`

**Current** (lines 713, 1024):
```go
sourcesEnv := os.Getenv("META_CC_CAPABILITY_SOURCES")
```

**Target**:
```go
sources := cfg.Capability.SourcesSlice()
```

**Benefits**:
- ✅ Already parsed into slice
- ✅ Single source of truth
- ✅ Easier to test

---

### Migration Plan Summary

**Phase 1**: MCP Server Configuration Migration
- Target files: logging.go, output_mode.go, response_adapter.go, capabilities.go
- Approach:
  1. Add `cfg, _ := config.Load()` to main.go startup
  2. Pass cfg to functions needing config
  3. Replace os.Getenv with cfg.* accesses
- LOC: ~50-60 changes

**Phase 2**: Test Updates (if time permits)
- Update test setup to use config
- LOC: ~15-20 changes

**Total Estimated LOC**: ~50-75

---

## Error Handling Standardization Opportunities

### Current State Analysis

**Error Handling Quality**:
- **Using %w correctly**: ~124 error sites (good wrapping)
- **NOT using %w**: ~25 error sites (needs improvement)
- **Missing sentinel errors**: No ErrNotFound, ErrInvalidInput, etc.

**Quality Metrics**:
- Wrapping rate: **83%** (124/149)
- Target: **95%** (142/149)
- Gap: **~18 error sites** to improve

### Error Pattern Analysis

#### Pattern 1: Simple Error Creation (Needs Context)

**Found in**: `internal/mcp/builder.go` (lines 177, 191, 220, 250, 313)

**Current**:
```go
return nil, fmt.Errorf("pattern parameter is required")
return nil, fmt.Errorf("error_signature parameter is required")
return nil, fmt.Errorf("file parameter is required")
return nil, fmt.Errorf("where parameter is required")
return nil, fmt.Errorf("unknown tool: %s", toolName)
```

**Issues**:
- ❌ No context (which function? which operation?)
- ❌ Not using sentinel errors (for programmatic checking)

**Target**:
```go
// Define sentinel errors
var (
    ErrMissingParameter = errors.New("required parameter missing")
    ErrUnknownTool     = errors.New("unknown tool")
)

// Usage with context
if pattern == "" {
    return nil, fmt.Errorf("pattern parameter required for query_user_messages: %w", ErrMissingParameter)
}

return nil, fmt.Errorf("unknown tool %s in BuildToolCommand: %w", toolName, ErrUnknownTool)
```

---

#### Pattern 2: Validation Errors (Needs Consistency)

**Found in**: `internal/query/*.go`

**Current**:
```go
// query/context.go:
return nil, fmt.Errorf("window size must be non-negative")

// query/sequences.go:
return nil, fmt.Errorf("minOccurrences must be at least 1")

// query/file_access.go:
return nil, fmt.Errorf("file path is required")
```

**Issues**:
- ❌ Inconsistent message format
- ❌ No operation context
- ❌ Not wrappable (no sentinel errors)

**Target**:
```go
var ErrInvalidInput = errors.New("invalid input")

if window < 0 {
    return nil, fmt.Errorf("window size must be non-negative for query_context (got: %d): %w", window, ErrInvalidInput)
}
```

---

#### Pattern 3: Parser Errors (Already Good!)

**Found in**: `internal/parser/*.go`

**Current** (GOOD):
```go
return nil, fmt.Errorf("failed to open session file: %w", err)
return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
return fmt.Errorf("failed to unmarshal tool_result content: %w", err)
```

**Assessment**: ✅ **Already follows best practices** (operation + context + %w)

---

### Sentinel Errors to Create

**Package**: `internal/errors/errors.go` (new file)

```go
package errors

import "errors"

// Sentinel errors for common conditions
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

**Usage**:
- Wrap with context: `fmt.Errorf("failed to load user %s: %w", id, errors.ErrNotFound)`
- Check programmatically: `if errors.Is(err, errors.ErrNotFound) { ... }`

---

### Error Standardization Targets (Priority Order)

#### 1. internal/mcp/builder.go (HIGH - 5 errors)
- Lines: 177, 191, 220, 250, 313
- Add operation context + sentinel errors
- Estimated: ~5 error messages to improve

#### 2. internal/query/*.go (MEDIUM - 3 errors)
- context.go, sequences.go, file_access.go
- Add validation context + sentinel errors
- Estimated: ~3 error messages to improve

#### 3. cmd/mcp-server/*.go (MEDIUM - ~15 errors)
- Survey for errors not using %w
- Add wrapping where missing
- Estimated: ~10-15 error sites

#### 4. internal/parser/*.go (LOW - already good)
- ✅ Already using %w correctly
- No changes needed

---

### Error Standardization Plan

**Phase 1**: Create Sentinel Errors (5-10 LOC)
- File: `internal/errors/errors.go`
- Define 5 common sentinel errors
- Add package documentation

**Phase 2**: Improve builder.go Errors (10-15 LOC)
- Wrap with sentinel errors
- Add operation context
- Improve error messages

**Phase 3**: Improve query/*.go Errors (5-10 LOC)
- Add context
- Use sentinel errors
- Consistent formatting

**Phase 4**: Survey MCP Server (10-15 LOC)
- Find errors without %w
- Add wrapping
- Improve context

**Total Estimated LOC**: ~30-40 changes

---

## Logging Expansion (Deferred)

**Rationale**: Iteration 3 deferred logging expansion. Still valid target for future iterations.

**Targets**:
- internal/parser/: ~10 log statements
- internal/query/: ~10 log statements

**Deferral Reason**: Focus on config migration + error standardization first (higher ROI).

---

## Implementation Priorities

### Iteration 4 Focus (Scoped)

**Priority 1: Configuration Migration** (50-75 LOC)
- ✅ Centralized config ready
- ✅ High ROI (reduces duplication)
- ✅ Improves testability
- 🎯 Target: 80% config centralized

**Priority 2: Error Standardization** (30-40 LOC)
- ✅ Conventions defined
- ✅ Improves consistency
- ✅ Enables programmatic error handling
- 🎯 Target: 95% error wrapping

**Priority 3: Logging Expansion** (DEFERRED)
- ⏳ Defer to Iteration 5
- Lower ROI than config/errors

**Total Estimated LOC**: 80-115 (well within 500 LOC Phase limit)

---

## Evidence for Metrics

### V_consistency Evidence
- **Current**: ~70% (124/149 errors use %w correctly)
- **Target**: ~95% (improve 18 error sites)
- **Config**: 0% centralized → 80% centralized (10 env var accesses)

### V_maintainability Evidence
- **Current**: Scattered config (7 files touch env vars)
- **Target**: Centralized config (1 file)
- **Benefit**: Single source of truth, easier updates

### V_enforcement Evidence
- **Current**: Manual (no linting)
- **Target**: Manual + sentinel errors (prepare for linting)
- **Improvement**: Sentinel errors enable future linting

### V_documentation Evidence
- **Current**: 0.80 (already converged)
- **Target**: 0.80 (maintain, no regression)

---

## Risks and Mitigations

### Risk 1: Breaking Changes
**Risk**: Config migration breaks MCP server
**Mitigation**: Backward compatibility built into config.Load() (LOG_LEVEL fallback)

### Risk 2: Scope Creep
**Risk**: Trying to fix all errors → exceeds LOC limit
**Mitigation**: Focus on high-priority files (builder.go, query/*.go)

### Risk 3: Test Failures
**Risk**: Config changes break tests
**Mitigation**: TDD approach, run tests frequently

---

## Summary

**Observation Complete**: ✅

**Key Targets Identified**:
1. **10 os.Getenv sites** → migrate to internal/config
2. **25+ error sites** → improve wrapping + add sentinel errors
3. **MCP server** → primary target for both

**Estimated LOC**: 80-115 (safe margin within 500 LOC limit)

**Next Phase**: M.plan (Define detailed plan for config migration + error standardization)

---

**Generated By**: M.observe (Meta-Agent)
**Agent**: data-analyst
**Status**: Ready for Planning Phase
