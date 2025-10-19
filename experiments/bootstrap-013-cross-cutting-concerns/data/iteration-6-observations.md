# Iteration 6: Data Collection & Observations

**Date**: 2025-10-17
**Phase**: M.observe
**Scope**: Remaining error sites for standardization

---

## Executive Summary

Analysis of remaining error sites identified in Iteration 5 plan. Three files require attention:
- **response_adapter.go**: 4 error sites (lines 48, 55, 72, 98)
- **jq_filter.go**: 3 error sites (lines 21, 35, 62)
- **capabilities.go**: 38 error sites (comprehensive review needed)

**Total Error Sites to Standardize**: ~45 sites

**Sentinel Errors Available** (from Iteration 5):
- ErrNotFound, ErrInvalidInput, ErrMissingParameter, ErrUnknownTool, ErrTimeout (Iteration 4)
- ErrFileIO, ErrNetworkFailure, ErrParseError, ErrConfigError (Iteration 5)

---

## File-by-File Analysis

### 1. response_adapter.go (4 error sites)

**File Size**: 127 lines
**Current Error Sites**: 4

**Error Analysis**:

1. **Line 48**: `fmt.Errorf("failed to write temp file: %w", err)`
   - **Context**: File write operation in adaptResponse
   - **Category**: File I/O
   - **Recommended Sentinel**: `ErrFileIO`
   - **Improvement**: Add file path context
   - **Pattern**: `fmt.Errorf("failed to write temp file %s: %w", filePath, mcerrors.ErrFileIO)`

2. **Line 55**: `fmt.Errorf("unknown output mode: %s", mode)`
   - **Context**: Invalid output mode selection
   - **Category**: Invalid input
   - **Recommended Sentinel**: `ErrInvalidInput`
   - **Improvement**: Wrap with sentinel
   - **Pattern**: `fmt.Errorf("unknown output mode %s in adaptResponse: %w", mode, mcerrors.ErrInvalidInput)`

3. **Line 72**: `fmt.Errorf("failed to generate file reference: %w", err)`
   - **Context**: File reference generation in buildFileRefResponse
   - **Category**: File I/O (metadata generation)
   - **Recommended Sentinel**: `ErrFileIO`
   - **Improvement**: Add file path context
   - **Pattern**: `fmt.Errorf("failed to generate file reference for %s: %w", filePath, mcerrors.ErrFileIO)`

4. **Line 98**: `fmt.Errorf("failed to serialize response: %w", err)`
   - **Context**: JSON serialization in serializeResponse
   - **Category**: Parsing/Serialization
   - **Recommended Sentinel**: `ErrParseError`
   - **Improvement**: Add operation context
   - **Pattern**: `fmt.Errorf("failed to serialize response to JSON: %w", mcerrors.ErrParseError)`

**Estimated LOC**: ~10 lines (4 sites + import)

---

### 2. jq_filter.go (3 error sites)

**File Size**: 114 lines
**Current Error Sites**: 3

**Error Analysis**:

1. **Line 21**: `fmt.Errorf("invalid jq expression: %w", err)`
   - **Context**: jq expression parsing in ApplyJQFilter
   - **Category**: Parse error
   - **Recommended Sentinel**: `ErrParseError`
   - **Improvement**: Add expression context
   - **Pattern**: `fmt.Errorf("invalid jq expression '%s': %w", jqExpr, mcerrors.ErrParseError)`

2. **Line 35**: `fmt.Errorf("invalid JSON: %w", err)`
   - **Context**: JSONL line parsing in ApplyJQFilter
   - **Category**: Parse error
   - **Recommended Sentinel**: `ErrParseError`
   - **Improvement**: Add line number/content context
   - **Pattern**: `fmt.Errorf("invalid JSON at line %d: %w", lineNum, mcerrors.ErrParseError)`

3. **Line 62**: `return "", err` (bare error return)
   - **Context**: JSON marshaling error in ApplyJQFilter
   - **Category**: Parse error (serialization)
   - **Recommended Sentinel**: `ErrParseError`
   - **Improvement**: Wrap with context
   - **Pattern**: `return "", fmt.Errorf("failed to marshal result to JSON: %w", mcerrors.ErrParseError)`

**Estimated LOC**: ~8 lines (3 sites + import + line numbering logic)

---

### 3. capabilities.go (38 error sites)

**File Size**: 1,051 lines
**Current Error Sites**: 38 (high concentration)

**Error Categories**:

#### A. File I/O Operations (13 sites)
- Lines 110, 134, 318, 405, 417, 433, 440, 454, 464, 469, 475, 481, 556
- **Category**: File creation, write, read, directory operations
- **Sentinel**: `ErrFileIO`
- **Pattern**: Add specific file path + operation context

#### B. Network Operations (4 sites)
- Lines 384, 394, 971, 975
- **Category**: HTTP download, jsDelivr API calls
- **Status Codes**: 500, 404, etc.
- **Sentinel**: `ErrNetworkFailure`
- **Pattern**: Add URL + status code context

#### C. Parse/Validation Errors (6 sites)
- Lines 269, 277, 282, 307, 311, 563
- **Category**: Frontmatter parsing, YAML validation, file access
- **Sentinel**: `ErrParseError` or `ErrConfigError`
- **Pattern**: Add file path + specific validation failure

#### D. Missing/Not Found Errors (6 sites)
- Lines 530, 749, 810, 826, 840, 1020
- **Category**: Capability not found, package not found, missing parameters
- **Sentinel**: `ErrNotFound` or `ErrMissingParameter`
- **Pattern**: Add specific item name + search context

#### E. Configuration/Validation Errors (5 sites)
- Lines 898, 913, 1009, 1051
- **Category**: Invalid GitHub source format, general errors
- **Sentinel**: `ErrConfigError` or `ErrInvalidInput`
- **Pattern**: Add configuration details

#### F. Already Good (4 sites)
- Lines 307, 384, 417, 563 (some already have %w wrapping)
- **Action**: Enhance with sentinel errors but preserve existing wrapping

**Prioritization**:
1. **HIGH**: File I/O (13 sites) - Most common, clear sentinel
2. **HIGH**: Not Found errors (6 sites) - User-facing, important for debugging
3. **MEDIUM**: Parse/Validation (6 sites) - Important but less frequent
4. **MEDIUM**: Network (4 sites) - Less frequent but high impact
5. **LOW**: Config errors (5 sites) - Edge cases

**Estimated LOC**: ~85 lines (38 sites + import + enhanced context)

**Strategy**:
- Focus on top 25 highest-impact sites (File I/O + Not Found + Parse errors)
- Defer remaining 13 sites if token budget constrained

---

## Summary Statistics

**Total Error Sites Identified**: 45
- response_adapter.go: 4 sites
- jq_filter.go: 3 sites
- capabilities.go: 38 sites

**Sentinel Error Mapping**:
- `ErrFileIO`: ~15 sites (33%)
- `ErrParseError`: ~10 sites (22%)
- `ErrNotFound`: ~6 sites (13%)
- `ErrNetworkFailure`: ~4 sites (9%)
- `ErrInvalidInput`: ~4 sites (9%)
- `ErrConfigError`: ~3 sites (7%)
- `ErrMissingParameter`: ~3 sites (7%)

**Estimated Total LOC**: ~103 lines for all error standardization

**Priority for Iteration 6**:
1. **Phase 1**: response_adapter.go + jq_filter.go (7 sites, ~18 LOC) - Quick wins
2. **Phase 2**: capabilities.go top 25 sites (~70 LOC) - High impact
3. **Phase 3**: Linter script creation (~100 LOC)
4. **Phase 4**: Documentation updates (~20 LOC)

**Total Planned LOC**: ~208 lines (within 500 LOC phase limit)

---

## Linter Requirements Analysis

Based on error standardization patterns, the linter should detect:

### Anti-Patterns to Detect

1. **Bare `fmt.Errorf` without %w**:
   ```bash
   grep -rn 'fmt\.Errorf.*".*"[^%]*$' --include="*.go" | grep -v '%w'
   ```
   - Finds error creation without wrapping
   - Suggests adding %w + sentinel error

2. **Short error messages** (<20 chars):
   ```bash
   grep -rn 'fmt\.Errorf.*".\{1,20\}"' --include="*.go"
   ```
   - Finds errors lacking context
   - Suggests adding operation details

3. **Missing sentinel error import**:
   ```bash
   grep -l 'fmt\.Errorf' --include="*.go" | xargs grep -L 'mcerrors "github.com/yaleh/meta-cc/internal/errors"'
   ```
   - Finds files with errors but no sentinel imports
   - Suggests adding import

4. **Direct `errors.New` usage**:
   ```bash
   grep -rn 'errors\.New' --include="*.go"
   ```
   - Should use sentinel errors instead
   - Suggests replacing with appropriate sentinel

### Linter Script Design (~100 LOC)

```bash
#!/bin/bash
# scripts/lint-errors.sh
# Simple error linting for meta-cc

# Features:
# 1. Check for bare fmt.Errorf without %w
# 2. Check for short error messages
# 3. Check for missing mcerrors import
# 4. Check for direct errors.New usage
# 5. Generate report with line numbers + suggestions

# Output format:
# FILE:LINE:SEVERITY: MESSAGE
# Example: cmd/mcp-server/foo.go:42:WARNING: fmt.Errorf without %w wrapping
```

**Estimated LOC**: 100 lines (bash script with grep patterns + report formatting)

---

## Observations Completed

**Data Collection**: COMPLETE ✓
- 3 files analyzed
- 45 error sites identified
- Sentinel error mapping defined
- Linter requirements specified

**Ready for M.plan**: YES
- Clear scope identified
- Estimated LOC calculated
- Prioritization defined
- Linter design sketched

---

**Generated By**: M.observe
**Date**: 2025-10-17
**Version**: 1.0
