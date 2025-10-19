# Iteration 7: Observations (Data Collection Phase)

**Date**: 2025-10-17
**Phase**: M.observe
**Status**: COMPLETE

---

## Data Collection

### Context from Iteration 6

**Carried Forward**:
- V_instance(s₆) = 0.55 (+0.08 from s₅)
- V_meta(s₆) = 0.56 (+0.10 from s₅)
- Error linter created (scripts/lint-errors.sh, 161 LOC, 4 checks)
- 7 error sites standardized (response_adapter.go: 4, jq_filter.go: 3)
- **Deferred Work**: capabilities.go standardization, linter CI integration, documentation

**System State**:
- M₆ = M₅ (no meta-agent evolution)
- A₆ = A₅ (no new agents)
- Accelerating progress (+17.0% V_instance, +21.7% V_meta)

---

## Error Site Analysis

### File: cmd/mcp-server/capabilities.go

**Complexity**: HIGH (1074 LOC, diverse error types, high user impact)

**Linter Results**:
- **12 warnings**: fmt.Errorf without %w wrapping
- **1 info**: Missing mcerrors import
- **Total**: 13 issues detected

**Error Categories Identified**:

#### 1. Validation Errors (2 sites)
**Lines**: 311, 1020
```go
// Line 311: Path validation
return nil, fmt.Errorf("path %s is not a directory", path)

// Line 1020: Parameter validation
return "", fmt.Errorf("missing required parameter: name")
```
**Sentinel**: ErrInvalidInput
**Context Needed**: Path value, parameter name

#### 2. File I/O Errors (5 sites)
**Lines**: 307, 110, 134, 433, 405
```go
// Line 307: Path access
return nil, fmt.Errorf("failed to access path %s: %w", path, err)

// Line 110: Session cache creation
sessionCacheInitErr = fmt.Errorf("failed to create session cache dir: %w", err)

// Line 134: Cache cleanup
return fmt.Errorf("failed to cleanup session cache: %w", err)

// Line 433: Package open
return fmt.Errorf("failed to open package: %w", err)

// Line 405: File creation
return fmt.Errorf("failed to create file: %w", err)
```
**Sentinel**: ErrFileIO
**Already Has %w**: Yes, but missing sentinel
**Context Needed**: File paths, operation details

#### 3. Network Errors (3 sites)
**Lines**: 384, 394, 971, 975
```go
// Line 384: Package download
return fmt.Errorf("failed to download package: %w", err)

// Line 394: HTTP status error
return fmt.Errorf("download failed with status %d", resp.StatusCode)

// Line 971: Server error
return fmt.Errorf("jsDelivr returned status %d (server error)", resp.StatusCode)

// Line 975: HTTP error
return fmt.Errorf("jsDelivr returned status %d", resp.StatusCode)
```
**Sentinel**: ErrNetworkError (needs creation)
**Context Needed**: URLs, status codes, operation details

#### 4. Parse/Format Errors (3 sites)
**Lines**: 277, 898, 913
```go
// Line 277: YAML parse
return CapabilityMetadata{}, fmt.Errorf("failed to parse frontmatter YAML: %w", err)

// Line 898, 913: GitHub source parsing
return result, fmt.Errorf("invalid GitHub source format: %s", location)
```
**Sentinel**: ErrParseError
**Context Needed**: Input strings, formats expected

#### 5. Not Found Errors (3 sites)
**Lines**: 530, 810, 840
```go
// Line 530: Package file not found
return fmt.Errorf("package file not found: %s", packagePath)

// Line 810: No sources configured
return "", fmt.Errorf("capability not found: %s (no sources configured)", name)

// Line 840: Capability not found
return "", fmt.Errorf("capability not found: %s", name)
```
**Sentinel**: ErrNotFound (needs creation)
**Context Needed**: Capability names, file paths

#### 6. Unknown Type Errors (2 sites)
**Lines**: 589, 826
```go
// Line 589: Unknown source type (in mergeSources)
return nil, fmt.Errorf("unknown source type: %s", source.Type)

// Line 826: Unknown source type (in getCapabilityContent)
return "", fmt.Errorf("unknown source type: %s", source.Type)
```
**Sentinel**: ErrInvalidInput
**Context Needed**: Type values, valid options

#### 7. Archive Extraction Errors (4 sites)
**Lines**: 439, 454, 464, 469, 475, 481
```go
// Line 439: gzip reader creation
return fmt.Errorf("failed to create gzip reader: %w", err)

// Line 454: tar read
return fmt.Errorf("tar read error: %w", err)

// Line 464, 469: Directory creation
return fmt.Errorf("failed to create directory: %w", err)
return fmt.Errorf("failed to create parent directory: %w", err)

// Line 475, 481: File operations
return fmt.Errorf("failed to create file: %w", err)
return fmt.Errorf("failed to write file: %w", err)
```
**Sentinel**: ErrFileIO (archives are files)
**Already Has %w**: Yes, but missing sentinel
**Context Needed**: Archive paths, file names

#### 8. Propagated Errors (Multiple sites)
**Lines**: 318, 521, 556, 563, 593, 632, 749, 836, 874, 1051
```go
// Already have %w wrapping, need sentinel enrichment:
return nil, fmt.Errorf("failed to glob .md files: %w", err)
return nil, fmt.Errorf("failed to load source %s: %w", source.Location, err)
return nil, fmt.Errorf("failed to download package: %w", err)
return nil, fmt.Errorf("failed to load package capabilities: %w", err)
return "", fmt.Errorf("failed to read from source %s: %w", source.Location, err)
return "", fmt.Errorf("failed to get capability: %w", err)
```

---

## Pattern Recognition

### High-Frequency Patterns

1. **File I/O Operations** (~12 sites)
   - Category: Infrastructure
   - Impact: Blocking (users can't load capabilities)
   - Standardization: ErrFileIO + file paths

2. **Network Operations** (~4 sites)
   - Category: Integration
   - Impact: Blocking (offline users, transient failures)
   - Standardization: ErrNetworkError + URLs + status codes

3. **Not Found Errors** (~3 sites)
   - Category: User Input
   - Impact: Degrading (unclear which capability to use)
   - Standardization: ErrNotFound + resource identifiers

4. **Parse Errors** (~3 sites)
   - Category: Data Processing
   - Impact: Degrading (capability metadata unusable)
   - Standardization: ErrParseError + format details

5. **Validation Errors** (~3 sites)
   - Category: User Input
   - Impact: Blocking (invalid parameters)
   - Standardization: ErrInvalidInput + validation details

---

## Gap Analysis

### Detection Gaps

1. **No Sentinel Errors for Network Operations**
   - Current: Generic fmt.Errorf
   - Impact: Can't distinguish network errors from other errors
   - Solution: Create ErrNetworkError sentinel

2. **No Sentinel Errors for Not Found**
   - Current: Custom notFoundError type exists, but not used consistently
   - Impact: Inconsistent error detection
   - Solution: Promote notFoundError to sentinel error

3. **Missing mcerrors Import**
   - Current: No access to sentinel errors
   - Impact: Can't standardize errors
   - Solution: Add import alias

### Diagnosis Gaps

1. **Network Errors Lack URL Context**
   - Lines 394, 971, 975: HTTP status code present, but URL missing
   - Impact: Can't troubleshoot which endpoint failed
   - Solution: Add URL to error messages

2. **File I/O Errors Lack Operation Context**
   - Lines 433, 464, 469, 475, 481: Generic "failed to X"
   - Impact: Hard to distinguish between create/read/write failures
   - Solution: Add operation type + file paths

3. **Parse Errors Lack Input Context**
   - Line 277: YAML parse error doesn't show which capability
   - Impact: Hard to identify malformed capability files
   - Solution: Add capability name + line numbers if available

### Prevention Gaps

1. **No Automated Enforcement**
   - Linter exists but not in CI
   - Impact: New errors regress without review
   - Solution: Add to Makefile + GitHub Actions

2. **No Documentation of Error Conventions**
   - Current: Patterns exist but not documented
   - Impact: Contributors don't know standards
   - Solution: Add to knowledge/best-practices/

---

## Prioritization

### Top 25 Error Sites (Priority Order)

**PRIORITY 1: User-Facing Errors** (High visibility, 8 sites):
1. Line 810: Capability not found (no sources)
2. Line 840: Capability not found (general)
3. Line 1020: Missing parameter validation
4. Line 311: Path not directory
5. Line 898, 913: Invalid GitHub source format (2 sites)
6. Line 530: Package file not found
7. Line 1009: Enhanced not found error

**PRIORITY 2: Network Operations** (Transient failure clarity, 4 sites):
8. Line 384: Package download failed
9. Line 394: Download status code error
10. Line 971: jsDelivr server error
11. Line 975: jsDelivr HTTP error

**PRIORITY 3: File I/O** (Infrastructure reliability, 10 sites):
12. Line 307: Path access failed
13. Line 110: Session cache creation
14. Line 134: Cache cleanup
15. Line 433: Package open failed
16. Line 439: Gzip reader creation
17. Line 454: Tar read error
18. Line 464, 469: Directory creation (2 sites)
19. Line 475, 481: File create/write (2 sites)

**PRIORITY 4: Parse Errors** (Data quality, 3 sites):
20. Line 277: YAML parse failed
21. Line 282: Missing name field
22. Line 269: No frontmatter found

**Coverage**: 25/38 sites (66% of file, meets target)

---

## CI Integration Requirements

### Makefile Target

**Requirement**: Add `make lint-errors` target

**Dependencies**:
- scripts/lint-errors.sh (exists)
- Bash shell (available on all platforms)

**Integration Point**: Add to `make all` or `make lint`

**Expected Output**: Exit code 0 (pass) or 1 (fail)

### GitHub Actions Workflow

**Requirement**: Add error-linting.yml workflow

**Trigger Events**:
- push (to main, feature branches)
- pull_request (all PRs)

**Jobs**:
1. checkout code
2. run `make lint-errors`
3. fail build if issues found

**Expected**: ~10-15 LOC YAML configuration

---

## Documentation Requirements

### Error Conventions Documentation

**Location**: knowledge/best-practices/error-handling.md (new file)

**Content**:
1. Sentinel error usage guide
2. Error wrapping patterns
3. Context enrichment examples
4. Linter usage instructions

**Expected**: ~50-100 LOC markdown

### Linter Usage Guide

**Location**: scripts/README.md or inline in lint-errors.sh

**Content**:
1. How to run linter locally
2. How to interpret warnings
3. How to fix common issues
4. CI integration status

**Expected**: ~20-30 LOC markdown

---

## Metrics

**Total Error Sites in capabilities.go**: ~38
**Linter-Detected Sites**: 13 (targeted for this iteration)
**Top 25 Priority Sites**: 25 (66% coverage target)
**New Sentinel Errors Needed**: 2 (ErrNetworkError, ErrNotFound)
**Files to Modify**: 3 (capabilities.go, internal/errors/errors.go, Makefile)
**Documentation Files to Create**: 2 (error conventions, linter guide)

---

## Next Phase Input

**For M.plan**:
1. Standardize top 25 error sites in capabilities.go
2. Create 2 new sentinel errors (ErrNetworkError, ErrNotFound)
3. Integrate linter into Makefile + GitHub Actions
4. Document error conventions
5. Expected ΔV_instance: +0.15-0.20
6. Expected ΔV_meta: +0.10-0.15

**Constraints**:
- Token budget management (learned from Iterations 5-6)
- Build validation after each change
- TDD approach for new sentinel errors

---

**Document Version**: 1.0
**Created**: 2025-10-17
**Generated By**: M.observe (Meta-Agent)
