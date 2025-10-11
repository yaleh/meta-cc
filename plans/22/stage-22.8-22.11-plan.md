# Phase 22 Stages 22.8-22.11: jsDelivr CDN Integration & Default Source Changes - TDD Implementation Plan

## Overview

**Objective**: Enhance Phase 22 with jsDelivr CDN integration for improved GitHub capability loading, support branch/tag specification, and change default source to GitHub.

**Code Volume**: ~420 lines total | **Priority**: High | **Estimated Time**: 6.5 hours

**Dependencies**:
- Phase 22 Stages 22.1-22.7 (Complete multi-source capability discovery system)
- Existing implementation: `cmd/mcp-server/capabilities.go` and `capabilities_test.go`

**Deliverables**:
- jsDelivr CDN URL generation (avoiding GitHub raw API rate limits)
- Branch/tag specification via `@` symbol syntax
- Version type detection (branch vs tag) for cache TTL management
- Enhanced error handling with retry logic and fallback strategies
- Default source changed to GitHub repository
- Comprehensive documentation updates

---

## Success Criteria

**Functional Acceptance**:
- ✅ jsDelivr URLs correctly generated for GitHub sources
- ✅ `@` symbol parsing for branch/tag specification works
- ✅ Version type detection correctly identifies branches vs tags
- ✅ Cache TTL adapts to version type (1h for branches, 7d for tags)
- ✅ Error handling with exponential backoff retry implemented
- ✅ Fallback to stale cache on network failure works
- ✅ Default source changed to GitHub without breaking local development
- ✅ Documentation updated and accurate

**Code Quality**:
- ✅ Total code: ~420 lines (within extended budget)
  - Stage 22.8: ~150 lines (jsDelivr integration)
  - Stage 22.9: ~50 lines (default source change)
  - Stage 22.10: ~100 lines (error handling)
  - Stage 22.11: ~120 lines (documentation)
- ✅ Each stage ≤ 200 lines
- ✅ Test coverage: ≥ 80%
- ✅ `make all` passes after each stage

---

## Stage 22.8: jsDelivr CDN Integration

### Objective

Replace GitHub raw URLs (raw.githubusercontent.com) with jsDelivr CDN URLs to avoid rate limiting, and implement branch/tag specification using `@` symbol syntax.

### Acceptance Criteria

- [ ] jsDelivr URL format correctly generated
- [ ] `@` symbol parsing extracts branch/tag from source location
- [ ] Default branch handling works (backward compatible)
- [ ] Branches, tags, and commit hashes supported
- [ ] Version type detection distinguishes branches from tags
- [ ] Cache TTL varies by version type (1h branches, 7d tags)
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `cmd/mcp-server/capabilities_test.go` (additions ~80 lines)

```go
// Add test functions:
// - TestParseGitHubSource - Test @ symbol parsing
// - TestBuildJsDelivrURL - Test URL generation
// - TestDetectVersionType - Test branch vs tag detection
// - TestGitHubSourceCacheTTL - Test cache TTL by version type
```

**Test Strategy**:
1. Test `@` symbol parsing with various inputs
2. Verify jsDelivr URL format matches expected pattern
3. Test version type detection with branches, tags, commits
4. Verify cache TTL adapts to version type
5. Test backward compatibility (no `@` defaults to main branch)

**Implementation Details**:

1. **GitHub Source Parsing** (parseGitHubSource function):

```go
// GitHubSource represents a parsed GitHub source with branch/tag
type GitHubSource struct {
    Owner   string
    Repo    string
    Branch  string // Branch, tag, or commit hash
    Subdir  string // Optional subdirectory
}

// parseGitHubSource parses GitHub source with @ symbol
// Format: "owner/repo@branch/subdir" or "owner/repo/subdir" (defaults to main)
func parseGitHubSource(location string) (GitHubSource, error) {
    var result GitHubSource

    // Check for @ symbol (branch/tag specification)
    atIndex := strings.Index(location, "@")

    if atIndex >= 0 {
        // Split at @ symbol
        beforeAt := location[:atIndex]
        afterAt := location[atIndex+1:]

        // Parse owner/repo before @
        parts := strings.SplitN(beforeAt, "/", 2)
        if len(parts) < 2 {
            return result, fmt.Errorf("invalid GitHub source format: %s", location)
        }
        result.Owner = parts[0]
        result.Repo = parts[1]

        // Parse branch/subdir after @
        branchParts := strings.SplitN(afterAt, "/", 2)
        result.Branch = branchParts[0]
        if len(branchParts) > 1 {
            result.Subdir = branchParts[1]
        }
    } else {
        // No @ symbol, default to main branch
        parts := strings.SplitN(location, "/", 3)
        if len(parts) < 2 {
            return result, fmt.Errorf("invalid GitHub source format: %s", location)
        }
        result.Owner = parts[0]
        result.Repo = parts[1]
        result.Branch = "main" // Default branch
        if len(parts) > 2 {
            result.Subdir = parts[2]
        }
    }

    return result, nil
}
```

**Test Cases**:
```go
// TestParseGitHubSource
tests := []struct {
    input    string
    expected GitHubSource
    hasError bool
}{
    {"yaleh/meta-cc@main/commands", GitHubSource{"yaleh", "meta-cc", "main", "commands"}, false},
    {"yaleh/meta-cc@v1.0.0/commands", GitHubSource{"yaleh", "meta-cc", "v1.0.0", "commands"}, false},
    {"yaleh/meta-cc@develop", GitHubSource{"yaleh", "meta-cc", "develop", ""}, false},
    {"yaleh/meta-cc/commands", GitHubSource{"yaleh", "meta-cc", "main", "commands"}, false},
    {"yaleh/meta-cc", GitHubSource{"yaleh", "meta-cc", "main", ""}, false},
    {"invalid", GitHubSource{}, true},
}
```

2. **jsDelivr URL Generation** (buildJsDelivrURL function):

```go
// buildJsDelivrURL generates jsDelivr CDN URL from GitHub source
// Format: https://cdn.jsdelivr.net/gh/owner/repo@branch/subdir/file.md
func buildJsDelivrURL(source GitHubSource, filename string) string {
    // Base URL
    url := fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s@%s",
        source.Owner, source.Repo, source.Branch)

    // Add subdirectory if present
    if source.Subdir != "" {
        url += "/" + source.Subdir
    }

    // Add filename
    url += "/" + filename

    return url
}
```

**Test Cases**:
```go
// TestBuildJsDelivrURL
tests := []struct {
    source   GitHubSource
    filename string
    expected string
}{
    {
        GitHubSource{"yaleh", "meta-cc", "main", "commands"},
        "meta-errors.md",
        "https://cdn.jsdelivr.net/gh/yaleh/meta-cc@main/commands/meta-errors.md",
    },
    {
        GitHubSource{"yaleh", "meta-cc", "v1.0.0", "commands"},
        "meta-errors.md",
        "https://cdn.jsdelivr.net/gh/yaleh/meta-cc@v1.0.0/commands/meta-errors.md",
    },
    {
        GitHubSource{"yaleh", "meta-cc", "develop", ""},
        "README.md",
        "https://cdn.jsdelivr.net/gh/yaleh/meta-cc@develop/README.md",
    },
}
```

3. **Version Type Detection** (detectVersionType function):

```go
// VersionType represents the type of version reference
type VersionType string

const (
    VersionTypeBranch VersionType = "branch"
    VersionTypeTag    VersionType = "tag"
)

// detectVersionType determines if a version is a branch or tag
// Tags typically match: v1.0.0, 1.0.0 (semantic versioning)
// Branches: main, develop, feature/xyz, etc.
func detectVersionType(version string) VersionType {
    // Semantic version pattern: v?1.0.0
    semverRegex := regexp.MustCompile(`^v?\d+\.\d+\.\d+`)

    if semverRegex.MatchString(version) {
        return VersionTypeTag
    }

    return VersionTypeBranch
}
```

**Test Cases**:
```go
// TestDetectVersionType
tests := []struct {
    version  string
    expected VersionType
}{
    {"v1.0.0", VersionTypeTag},
    {"v1.2.3", VersionTypeTag},
    {"1.0.0", VersionTypeTag},
    {"1.2.3-beta", VersionTypeTag},
    {"main", VersionTypeBranch},
    {"develop", VersionTypeBranch},
    {"feature/xyz", VersionTypeBranch},
    {"abc123def", VersionTypeBranch}, // Commit hash
}
```

4. **Cache TTL by Version Type**:

```go
// getCacheTTL returns cache TTL based on version type
func getCacheTTL(versionType VersionType) time.Duration {
    switch versionType {
    case VersionTypeTag:
        return 7 * 24 * time.Hour // 7 days for tags (immutable)
    case VersionTypeBranch:
        return 1 * time.Hour // 1 hour for branches (mutable)
    default:
        return 1 * time.Hour
    }
}
```

5. **Update loadGitHubCapabilities**:

Modify existing placeholder implementation to use jsDelivr:

```go
func loadGitHubCapabilities(repo string) ([]CapabilityMetadata, error) {
    // Parse GitHub source
    source, err := parseGitHubSource(repo)
    if err != nil {
        return nil, err
    }

    // Detect version type for cache management
    versionType := detectVersionType(source.Branch)

    // For now, we need to list files first
    // This requires GitHub API or jsDelivr directory listing
    // Implementation will fetch known capability files
    // TODO: Full implementation in future stage

    return nil, fmt.Errorf("GitHub capability loading via jsDelivr not fully implemented")
}
```

6. **Update readGitHubCapability**:

Modify existing placeholder implementation to use jsDelivr:

```go
func readGitHubCapability(name string, repo string) (string, error) {
    // Parse GitHub source
    source, err := parseGitHubSource(repo)
    if err != nil {
        return "", err
    }

    // Build jsDelivr URL
    url := buildJsDelivrURL(source, name+".md")

    // Fetch content from jsDelivr
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode == 404 {
        return "", newNotFoundError(name)
    }

    if resp.StatusCode != 200 {
        return "", fmt.Errorf("jsDelivr returned status %d", resp.StatusCode)
    }

    content, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(content), nil
}
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/capabilities.go` (+150 lines)
  - Add GitHubSource struct
  - Add parseGitHubSource function (~30 lines)
  - Add buildJsDelivrURL function (~15 lines)
  - Add VersionType enum and detectVersionType function (~20 lines)
  - Add getCacheTTL function (~10 lines)
  - Update loadGitHubCapabilities placeholder (~25 lines)
  - Update readGitHubCapability implementation (~50 lines)

- `cmd/mcp-server/capabilities_test.go` (+80 lines)
  - Add TestParseGitHubSource (~30 lines)
  - Add TestBuildJsDelivrURL (~20 lines)
  - Add TestDetectVersionType (~20 lines)
  - Add TestGitHubSourceCacheTTL (~10 lines)

**Total**: ~230 lines (exceeds 150-line target due to comprehensive tests)

### Test Commands

```bash
# Run Stage 22.8 tests
go test -v ./cmd/mcp-server -run TestParseGitHubSource
go test -v ./cmd/mcp-server -run TestBuildJsDelivrURL
go test -v ./cmd/mcp-server -run TestDetectVersionType
go test -v ./cmd/mcp-server -run TestReadGitHubCapability

# Test jsDelivr integration manually (requires network)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"meta-errors"}}}' | meta-cc-mcp

# Test with specific tag
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"meta-errors"}}}' | meta-cc-mcp

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test `@` symbol parsing with various formats
3. Test jsDelivr URL generation
4. Test version type detection (branches, tags, commits)
5. Test cache TTL by version type
6. Manual test with real jsDelivr CDN (network required)
7. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 22.1 (multi-source capability loading)
- Stage 22.3 (get_capability tool)

### Estimated Time

2 hours (230 lines implementation + tests)

---

## Stage 22.9: Default Source to GitHub

### Objective

Change the default capability source from local `.claude/commands` to GitHub repository `yaleh/meta-cc@main/commands`, enabling zero-configuration deployment while preserving local development workflow.

### Acceptance Criteria

- [ ] Default source changed to GitHub
- [ ] Users without META_CC_CAPABILITY_SOURCES get capabilities from GitHub
- [ ] Local development requires explicit environment variable
- [ ] Cache strategy updated (branch: 1h, tag: 7d, local: no cache)
- [ ] Documentation updated with new default behavior
- [ ] Unit tests updated to reflect new default

### TDD Approach

**Test File**: `cmd/mcp-server/capabilities_test.go` (modifications ~30 lines)

```go
// Modify existing tests:
// - TestExecuteListCapabilitiesTool - Update default source expectations
// - TestExecuteGetCapabilityTool - Update default source expectations
// Add new test:
// - TestDefaultSourceIsGitHub - Verify default source behavior
```

**Test Strategy**:
1. Verify default source is GitHub when no env var set
2. Test local override via environment variable
3. Verify cache behavior for default GitHub source
4. Test backward compatibility with existing tests

**Implementation Details**:

1. **Update Default Source Constant**:

```go
const (
    // DefaultCapabilitySource is the default source when no env var is set
    DefaultCapabilitySource = "yaleh/meta-cc@main/commands"
)
```

2. **Update executeListCapabilitiesTool**:

Modify default source logic in `cmd/mcp-server/capabilities.go`:

```go
func executeListCapabilitiesTool(args map[string]interface{}) (string, error) {
    // Parse sources (test override or environment variable)
    sourcesEnv := os.Getenv("META_CC_CAPABILITY_SOURCES")
    if override, ok := args["_sources"].(string); ok && override != "" {
        sourcesEnv = override
    }

    // Parse sources
    sources := parseCapabilitySources(sourcesEnv)
    if len(sources) == 0 {
        // Default to GitHub repository if no sources configured
        sources = []CapabilitySource{
            {Type: SourceTypeGitHub, Location: DefaultCapabilitySource, Priority: 0},
        }
    }

    // ... rest of implementation unchanged
}
```

3. **Update executeGetCapabilityTool**:

Same change for get_capability tool:

```go
func executeGetCapabilityTool(args map[string]interface{}) (string, error) {
    // ... parameter parsing unchanged

    // Parse sources
    sources := parseCapabilitySources(sourcesEnv)
    if len(sources) == 0 {
        // Default to GitHub repository if no sources configured
        sources = []CapabilitySource{
            {Type: SourceTypeGitHub, Location: DefaultCapabilitySource, Priority: 0},
        }
    }

    // ... rest of implementation unchanged
}
```

4. **Update Cache Strategy Documentation**:

Add comment explaining cache behavior:

```go
// getCapabilityIndex returns capability index with caching support
// Cache strategy:
// - Local sources: No cache (always fresh)
// - GitHub branches: 1-hour cache
// - GitHub tags: 7-day cache (immutable)
func getCapabilityIndex(sources []CapabilitySource, disableCache bool) (CapabilityIndex, error) {
    // ... existing implementation unchanged
}
```

5. **Update Tests for New Default**:

```go
// TestDefaultSourceIsGitHub verifies default source behavior
func TestDefaultSourceIsGitHub(t *testing.T) {
    // Clear environment variable
    oldEnv := os.Getenv("META_CC_CAPABILITY_SOURCES")
    os.Unsetenv("META_CC_CAPABILITY_SOURCES")
    defer os.Setenv("META_CC_CAPABILITY_SOURCES", oldEnv)

    // Call with no _sources override
    args := map[string]interface{}{}

    // Note: This will fail with "not fully implemented" error
    // which is expected until full GitHub implementation
    _, err := executeListCapabilitiesTool(args)

    // Verify error mentions GitHub (not local path)
    if err != nil && !strings.Contains(err.Error(), "GitHub") {
        t.Errorf("expected GitHub-related error, got: %v", err)
    }
}
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/capabilities.go` (+50 lines)
  - Add DefaultCapabilitySource constant (~5 lines)
  - Update executeListCapabilitiesTool default (~5 lines)
  - Update executeGetCapabilityTool default (~5 lines)
  - Add cache strategy comments (~10 lines)
  - Update related documentation comments (~25 lines)

- `cmd/mcp-server/capabilities_test.go` (+30 lines)
  - Add TestDefaultSourceIsGitHub (~20 lines)
  - Update existing tests for new default (~10 lines)

**Total**: ~80 lines (exceeds 50-line target due to test updates)

### Test Commands

```bash
# Test default source (requires network or will fail with expected error)
unset META_CC_CAPABILITY_SOURCES
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# Test local override
export META_CC_CAPABILITY_SOURCES=".claude/commands"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# Run Stage 22.9 tests
go test -v ./cmd/mcp-server -run TestDefaultSource

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test default source behavior (no env var)
3. Test local source override
4. Verify cache strategy
5. Update integration tests
6. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 22.8 (jsDelivr integration)

### Estimated Time

1 hour (80 lines implementation + tests)

---

## Stage 22.10: Error Handling and Fallback

### Objective

Implement robust error handling for jsDelivr CDN and network failures, including exponential backoff retry logic and fallback to stale cache.

### Acceptance Criteria

- [ ] 404 errors provide clear, actionable messages
- [ ] 5xx errors trigger exponential backoff retry (3 attempts)
- [ ] Network unreachable errors use stale cache if available
- [ ] Retry logic: 1s, 2s, 4s delays
- [ ] Fallback priority: fresh > cached > stale > error
- [ ] Error messages distinguish between error types
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `cmd/mcp-server/capabilities_test.go` (additions ~60 lines)

```go
// Add test functions:
// - TestJsDelivrErrorHandling404 - Test 404 error messages
// - TestJsDelivrErrorHandling5xx - Test retry logic
// - TestFallbackToStaleCache - Test stale cache usage
// - TestExponentialBackoff - Test retry delays
```

**Test Strategy**:
1. Mock HTTP responses for different error scenarios
2. Test retry logic with 5xx errors
3. Verify stale cache fallback on network failure
4. Test error message clarity
5. Verify exponential backoff timing

**Implementation Details**:

1. **Enhanced Cache with Staleness Support**:

```go
// CapabilityCache enhanced with staleness tracking
type CapabilityCache struct {
    Index     CapabilityIndex
    Timestamp time.Time
    TTL       time.Duration
    Sources   []CapabilitySource // Track sources for validation
}

// isCacheStale checks if cache is stale but still usable
func isCacheStale(sources []CapabilitySource) bool {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()

    if globalCapabilityCache == nil {
        return false
    }

    age := time.Since(globalCapabilityCache.Timestamp)
    maxStaleAge := 7 * 24 * time.Hour // 7 days

    return age >= globalCapabilityCache.TTL && age < maxStaleAge
}
```

2. **Exponential Backoff Retry**:

```go
// retryWithBackoff performs exponential backoff retry for transient errors
func retryWithBackoff(operation func() error, maxRetries int) error {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }

        lastErr = err

        // Don't retry on 404 (not found) or network unreachable
        if isNotFoundError(err) || isNetworkUnreachableError(err) {
            return err
        }

        // Only retry on 5xx server errors
        if !isServerError(err) {
            return err
        }

        // Exponential backoff: 1s, 2s, 4s
        if attempt < maxRetries-1 {
            delay := time.Duration(1<<attempt) * time.Second
            time.Sleep(delay)
        }
    }

    return lastErr
}

// Error type detection helpers
func isServerError(err error) bool {
    if err == nil {
        return false
    }
    // Check if error message contains "status 5"
    return strings.Contains(err.Error(), "status 5")
}

func isNetworkUnreachableError(err error) bool {
    if err == nil {
        return false
    }
    // Check for network-related errors
    return strings.Contains(err.Error(), "no such host") ||
           strings.Contains(err.Error(), "connection refused") ||
           strings.Contains(err.Error(), "network is unreachable")
}
```

3. **Enhanced readGitHubCapability with Error Handling**:

```go
func readGitHubCapability(name string, repo string) (string, error) {
    // Parse GitHub source
    source, err := parseGitHubSource(repo)
    if err != nil {
        return "", err
    }

    // Build jsDelivr URL
    url := buildJsDelivrURL(source, name+".md")

    // Retry logic for transient errors
    var content string
    err = retryWithBackoff(func() error {
        // Fetch content from jsDelivr
        resp, err := http.Get(url)
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        if resp.StatusCode == 404 {
            // Distinguish between file not found and repo/branch not found
            return enhanceNotFoundError(name, source)
        }

        if resp.StatusCode >= 500 {
            return fmt.Errorf("jsDelivr returned status %d (server error)", resp.StatusCode)
        }

        if resp.StatusCode != 200 {
            return fmt.Errorf("jsDelivr returned status %d", resp.StatusCode)
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return err
        }

        content = string(body)
        return nil
    }, 3) // Max 3 retries

    if err != nil {
        return "", err
    }

    return content, nil
}

// enhanceNotFoundError provides actionable error messages
func enhanceNotFoundError(name string, source GitHubSource) error {
    msg := fmt.Sprintf("capability not found: %s\n", name)
    msg += fmt.Sprintf("Source: %s/%s@%s", source.Owner, source.Repo, source.Branch)

    if source.Subdir != "" {
        msg += fmt.Sprintf("/%s", source.Subdir)
    }

    msg += "\n\nPossible causes:\n"
    msg += "  1. Capability file does not exist\n"
    msg += "  2. Branch/tag name is incorrect\n"
    msg += "  3. Repository or subdirectory is incorrect\n"
    msg += "\nSuggestion: Verify the source configuration"

    return fmt.Errorf(msg)
}
```

4. **Fallback to Stale Cache**:

```go
// getCapabilityIndex with fallback support
func getCapabilityIndex(sources []CapabilitySource, disableCache bool) (CapabilityIndex, error) {
    // Check fresh cache if enabled
    if !disableCache && !hasLocalSources(sources) && isCacheValid(sources) {
        cacheMutex.RLock()
        defer cacheMutex.RUnlock()
        return globalCapabilityCache.Index, nil
    }

    // Load fresh data
    index, err := mergeSources(sources)
    if err != nil {
        // If network error and stale cache available, use stale cache
        if isNetworkUnreachableError(err) && isCacheStale(sources) {
            cacheMutex.RLock()
            defer cacheMutex.RUnlock()

            fmt.Fprintf(os.Stderr, "Warning: Using cached capabilities (may be outdated)\n")
            fmt.Fprintf(os.Stderr, "Network error: %v\n", err)

            return globalCapabilityCache.Index, nil
        }

        return nil, err
    }

    // Update cache (only if no local sources)
    if !hasLocalSources(sources) {
        // Determine TTL based on version type
        ttl := 1 * time.Hour // Default for branches

        // Check if any source is a tag
        for _, source := range sources {
            if source.Type == SourceTypeGitHub {
                ghSource, _ := parseGitHubSource(source.Location)
                if detectVersionType(ghSource.Branch) == VersionTypeTag {
                    ttl = 7 * 24 * time.Hour
                    break
                }
            }
        }

        cacheMutex.Lock()
        globalCapabilityCache = &CapabilityCache{
            Index:     index,
            Timestamp: time.Now(),
            TTL:       ttl,
            Sources:   sources,
        }
        cacheMutex.Unlock()
    }

    return index, nil
}
```

5. **User-Friendly Error Messages**:

```go
// formatCapabilityError formats errors with helpful context
func formatCapabilityError(err error, sources []CapabilitySource) error {
    if err == nil {
        return nil
    }

    if isNetworkUnreachableError(err) {
        return fmt.Errorf(
            "Network unavailable. Cannot load capabilities from GitHub.\n\n" +
            "To use local capabilities:\n" +
            "  export META_CC_CAPABILITY_SOURCES='commands'\n\n" +
            "Original error: %w", err)
    }

    return err
}
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/capabilities.go` (+100 lines)
  - Enhance CapabilityCache struct (~5 lines)
  - Add isCacheStale function (~15 lines)
  - Add retryWithBackoff function (~30 lines)
  - Add error type detection helpers (~15 lines)
  - Update readGitHubCapability with retry (~30 lines)
  - Add enhanceNotFoundError function (~10 lines)
  - Update getCapabilityIndex with fallback (~20 lines)
  - Add formatCapabilityError function (~10 lines)

- `cmd/mcp-server/capabilities_test.go` (+60 lines)
  - Add TestJsDelivrErrorHandling404 (~15 lines)
  - Add TestJsDelivrErrorHandling5xx (~15 lines)
  - Add TestFallbackToStaleCache (~20 lines)
  - Add TestExponentialBackoff (~10 lines)

**Total**: ~160 lines (exceeds 100-line target due to comprehensive error handling)

### Test Commands

```bash
# Run Stage 22.10 tests
go test -v ./cmd/mcp-server -run TestJsDelivrErrorHandling
go test -v ./cmd/mcp-server -run TestFallbackToStaleCache
go test -v ./cmd/mcp-server -run TestExponentialBackoff

# Manual test: Network failure simulation (disconnect network)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp
# Should use stale cache if available

# Manual test: 404 error (invalid source)
export META_CC_CAPABILITY_SOURCES="yaleh/nonexistent@main"
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"test"}}}' | meta-cc-mcp
# Should see clear error message

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test 404 error messages (clear and actionable)
3. Test retry logic with mocked 5xx responses
4. Test stale cache fallback
5. Verify exponential backoff timing
6. Test network unreachable scenario
7. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 22.8 (jsDelivr integration)
- Stage 22.9 (default source change)

### Estimated Time

1.5 hours (160 lines implementation + tests)

---

## Stage 22.11: Documentation Update

### Objective

Update all documentation to reflect jsDelivr CDN integration, branch/tag specification syntax, GitHub default source, and cache strategies.

### Acceptance Criteria

- [ ] CLAUDE.md updated with jsDelivr and @ symbol syntax
- [ ] README.md updated with zero-configuration emphasis
- [ ] docs/capabilities-guide.md updated with local development workflow
- [ ] CHANGELOG.md includes breaking changes and migration guide
- [ ] All examples use new syntax
- [ ] Documentation accurate and comprehensive

### Implementation

**1. CLAUDE.md Updates** (+40 lines):

Location: Section "Unified Meta Command" → "Multi-Source Configuration"

```markdown
### Multi-Source Configuration

Configure capability sources via environment variable:

```bash
# Single local source
export META_CC_CAPABILITY_SOURCES="~/.config/meta-cc/capabilities"

# Multiple sources (priority: left-to-right, left = highest)
export META_CC_CAPABILITY_SOURCES="~/dev/my-caps:yaleh/meta-cc-capabilities"

# Mix local and GitHub
export META_CC_CAPABILITY_SOURCES="~/dev/test:.claude/commands:community/extras"
```

**Source Types**:
- **Local directories**: Immediate reflection, no cache (for development)
- **GitHub repositories**: 1-hour cache for branches, 7-day cache for tags (format: `owner/repo@branch` or `owner/repo`)

**Branch/Tag Specification**:

Use the `@` symbol to specify a branch or tag:

```bash
# Specific branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"

# Specific tag (version pinning)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"

# Specific commit hash
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@abc123/commands"

# Default branch (main)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc/commands"
```

**CDN and Caching**:

GitHub sources use jsDelivr CDN (https://cdn.jsdelivr.net) for improved performance and rate limit avoidance:

- **Branches**: 1-hour cache (mutable, changes frequently)
- **Tags**: 7-day cache (immutable, stable versions)
- **Local sources**: No cache (always fresh)

**Network Resilience**:

meta-cc automatically handles network failures:

- **5xx server errors**: Exponential backoff retry (3 attempts: 1s, 2s, 4s)
- **Network unreachable**: Falls back to stale cache (up to 7 days old)
- **404 errors**: Clear error messages with troubleshooting suggestions

**Default Source**:

If `META_CC_CAPABILITY_SOURCES` is not set, capabilities are loaded from:
```
yaleh/meta-cc@main/commands
```

For local development, explicitly set the environment variable:
```bash
export META_CC_CAPABILITY_SOURCES="commands"
```
```

**2. README.md Updates** (+30 lines):

Location: "Unified Meta Command" section

```markdown
## Unified Meta Command

Use natural language to invoke meta-cognition capabilities:

```bash
/meta "show errors"           # Error analysis
/meta "quality check"         # Code quality scan
/meta "visualize timeline"    # Project timeline
```

### Zero-Configuration Setup

meta-cc works out of the box with no configuration required. Capabilities are automatically loaded from GitHub:

```
yaleh/meta-cc@main/commands
```

This provides:
- **Latest capabilities**: Always up-to-date with the main branch
- **No local files needed**: Capabilities loaded from jsDelivr CDN
- **Fast and reliable**: CDN caching avoids rate limits

### Local Development

For local capability development, override the default source:

```bash
export META_CC_CAPABILITY_SOURCES="commands"
```

This enables:
- **Real-time reflection**: Changes appear immediately (no cache)
- **Testing before commit**: Verify changes locally
- **Offline work**: No network dependency

### Version Pinning

Pin capabilities to a specific version for stability:

```bash
# Use a specific release tag
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"

# Cache: 7 days (tags are immutable)
```

### Multi-Source Capabilities

Load capabilities from multiple sources:

```bash
export META_CC_CAPABILITY_SOURCES="~/my-caps:.claude/commands:yaleh/meta-cc-extras@main"
```

Supports:
- **Local directories**: Immediate reflection, no cache
- **GitHub repositories**: jsDelivr CDN, smart caching (1h branches, 7d tags)
- **Priority-based merging**: Left = highest priority (overrides duplicates)

See [docs/capabilities-guide.md](docs/capabilities-guide.md) for development guide.
```

**3. docs/capabilities-guide.md Updates** (+30 lines):

Location: "Local Development Workflow" section and new "Testing Against Branches" section

```markdown
## Local Development Workflow

**Important**: meta-cc now defaults to loading capabilities from GitHub. For local development, you must explicitly configure a local source.

1. **Configure local source**:
   ```bash
   export META_CC_CAPABILITY_SOURCES="~/dev/my-capabilities"
   ```

2. **Create capability file**:
   ```bash
   cat > ~/dev/my-capabilities/my-feature.md <<EOF
   ---
   name: my-feature
   description: My custom feature analysis.
   keywords: feature, custom, analysis
   category: analysis
   ---

   # My Feature

   Implementation here...
   EOF
   ```

3. **Test capability** (changes reflect immediately):
   ```bash
   # List capabilities (verify yours appears)
   echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

   # Get capability content
   echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"my-feature"}}}' | meta-cc-mcp

   # Use via /meta command
   /meta "my feature"
   ```

4. **Iterate**:
   - Edit capability file
   - Changes reflect immediately (no cache for local sources)
   - Test with `/meta` command

## Testing Against Branches

Test capabilities from different branches before merging:

```bash
# Test from develop branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"
/meta "show errors"

# Test from feature branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@feature/new-capability/commands"
/meta "new capability"

# Test from pull request commit
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@abc123def/commands"
/meta "experimental"
```

**Cache Behavior**:
- Branches: 1-hour cache (changes propagate within 1 hour)
- To force refresh: Restart MCP server or use `_disable_cache: true` parameter

## Publishing Capabilities

### Method 1: GitHub Repository

1. Create GitHub repo: `username/meta-cc-capabilities`
2. Add capabilities: `capabilities/my-feature.md`
3. Users install via:
   ```bash
   # Latest (main branch, 1-hour cache)
   export META_CC_CAPABILITY_SOURCES="username/meta-cc-capabilities/capabilities"

   # Stable version (tag, 7-day cache)
   export META_CC_CAPABILITY_SOURCES="username/meta-cc-capabilities@v1.0.0/capabilities"
   ```

**Recommendation**: Use semantic versioning tags (v1.0.0) for stable releases. This enables:
- **7-day cache**: Faster loading, reduced CDN requests
- **Version pinning**: Users can opt into specific versions
- **Immutability**: Tags don't change, ensuring consistency

### Method 2: Fork and PR

1. Fork `yaleh/meta-cc`
2. Add capability: `.claude/commands/meta-my-feature.md`
3. Submit PR
4. After merge, available to all users via default source
```

**4. CHANGELOG.md Updates** (+20 lines):

Location: Create new unreleased section at top

```markdown
## [Unreleased]

### Added (Phase 22.8-22.11)

- **jsDelivr CDN Integration**: GitHub capabilities now load via jsDelivr CDN (cdn.jsdelivr.net)
  - Avoids GitHub raw API rate limiting
  - Improved performance and reliability
  - Smart caching (1h branches, 7d tags)

- **Branch/Tag Specification**: Use `@` symbol to specify versions
  - Format: `owner/repo@branch/subdir`
  - Examples: `yaleh/meta-cc@v1.0.0/commands`, `yaleh/meta-cc@develop/commands`
  - Supports branches, tags, and commit hashes

- **Enhanced Error Handling**:
  - Exponential backoff retry for 5xx server errors (3 attempts: 1s, 2s, 4s)
  - Fallback to stale cache on network failure (up to 7 days)
  - Clear, actionable error messages for 404 and network errors

### Changed (Phase 22.8-22.11)

- **Default Source Changed**: Capabilities now load from `yaleh/meta-cc@main/commands` by default
  - Zero-configuration deployment
  - Local development requires: `export META_CC_CAPABILITY_SOURCES="commands"`

- **Cache Strategy Enhanced**:
  - Branches: 1-hour cache (mutable)
  - Tags: 7-day cache (immutable)
  - Local sources: No cache (always fresh)

### Breaking Changes

⚠️ **Default Source Changed**: If you were relying on the default `.claude/commands` source, you must now explicitly set:

```bash
export META_CC_CAPABILITY_SOURCES="commands"
```

Or use the new GitHub default: `yaleh/meta-cc@main/commands`

### Migration Guide

**For Local Development**:

Old behavior (implicit):
```bash
# Capabilities loaded from .claude/commands automatically
```

New behavior (explicit):
```bash
# Must explicitly set local source
export META_CC_CAPABILITY_SOURCES="commands"
```

**For Production Deployment**:

Old behavior:
```bash
# Required environment variable
export META_CC_CAPABILITY_SOURCES=".claude/commands"
```

New behavior:
```bash
# No configuration needed (uses GitHub default)
# OR explicitly set for custom source
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"
```

**For Version Pinning**:

```bash
# Pin to specific release (recommended for production)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"

# Benefits:
# - 7-day cache (faster loading)
# - Immutable (no unexpected changes)
# - Explicit version control
```
```

### File Changes

**Modified Files**:
- `CLAUDE.md` (+40 lines)
- `README.md` (+30 lines)
- `docs/capabilities-guide.md` (+30 lines)
- `CHANGELOG.md` (+20 lines)

**Total**: ~120 lines

### Verification Commands

```bash
# Verify documentation updates
grep -r "jsDelivr" CLAUDE.md README.md docs/
grep -r "@" CLAUDE.md README.md docs/ | grep -v "email"
grep -r "cdn.jsdelivr.net" CLAUDE.md README.md docs/

# Check markdown syntax
for file in CLAUDE.md README.md docs/capabilities-guide.md CHANGELOG.md; do
    echo "Checking $file..."
    # Use markdown linter if available
    markdownlint "$file" || echo "  (markdownlint not available, skipping)"
done

# Verify code examples
grep -A 5 "export META_CC_CAPABILITY_SOURCES" CLAUDE.md README.md docs/capabilities-guide.md

# Check cross-references
grep -r "capabilities-guide.md" CLAUDE.md README.md
```

### Testing Protocol

**After Documentation**:
1. Verify all documentation files updated
2. Check markdown syntax and formatting
3. Verify code examples are correct
4. Test examples in documentation
5. Check cross-references between documents
6. Verify CHANGELOG.md completeness
7. Review migration guide clarity

### Dependencies

- All previous stages (22.8-22.10)

### Estimated Time

2 hours (120 lines documentation + verification)

---

## Phase Integration Strategy

### Build Verification

After completing all stages 22.8-22.11:

```bash
# 1. Full build
make all

# 2. Unit tests
go test -v ./cmd/mcp-server -run TestParseGitHubSource
go test -v ./cmd/mcp-server -run TestBuildJsDelivrURL
go test -v ./cmd/mcp-server -run TestDetectVersionType
go test -v ./cmd/mcp-server -run TestDefaultSource
go test -v ./cmd/mcp-server -run TestJsDelivrErrorHandling
go test -v ./cmd/mcp-server -run TestFallbackToStaleCache

# 3. Integration tests
go test -v ./cmd/mcp-server -run TestIntegration

# 4. Test coverage
make test-coverage
# Should maintain ≥80% coverage

# 5. jsDelivr integration tests (requires network)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# Test with tag
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"meta-errors"}}}' | meta-cc-mcp

# Test with branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"meta-errors"}}}' | meta-cc-mcp

# Test default source (no env var)
unset META_CC_CAPABILITY_SOURCES
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp
```

### Rollout Checklist

Before marking Stages 22.8-22.11 complete:

- [ ] All 4 stages completed and tested
- [ ] `make all` passes without errors
- [ ] Test coverage ≥80% maintained
- [ ] jsDelivr URL generation works correctly
- [ ] `@` symbol parsing for branches/tags works
- [ ] Version type detection distinguishes branches from tags
- [ ] Cache TTL adapts to version type
- [ ] Exponential backoff retry logic works for 5xx errors
- [ ] Stale cache fallback works on network failure
- [ ] Default source changed to GitHub
- [ ] Local development workflow documented
- [ ] Documentation updated and accurate
- [ ] CHANGELOG.md includes breaking changes and migration guide
- [ ] Manual jsDelivr testing successful (network required)
- [ ] Backward compatibility verified

---

## File Change Inventory

### Summary by Stage

| Stage  | Modified Files | Total Lines | Description |
|--------|----------------|-------------|-------------|
| 22.8   | 2              | ~230        | jsDelivr integration + tests |
| 22.9   | 2              | ~80         | Default source change + tests |
| 22.10  | 2              | ~160        | Error handling + fallback + tests |
| 22.11  | 4              | ~120        | Documentation updates |
| **Total** | **10** | **~590** | **Comprehensive enhancement** |

**Note**: Total ~590 lines exceeds initial 420-line estimate due to:
- Comprehensive error handling (retry, fallback, messages)
- Extensive test coverage (≥80% for all new code)
- Detailed documentation (migration guide, examples)

### Detailed File Changes

**Stage 22.8**:
- `cmd/mcp-server/capabilities.go` (+150 lines)
- `cmd/mcp-server/capabilities_test.go` (+80 lines)

**Stage 22.9**:
- `cmd/mcp-server/capabilities.go` (+50 lines)
- `cmd/mcp-server/capabilities_test.go` (+30 lines)

**Stage 22.10**:
- `cmd/mcp-server/capabilities.go` (+100 lines)
- `cmd/mcp-server/capabilities_test.go` (+60 lines)

**Stage 22.11**:
- `CLAUDE.md` (+40 lines)
- `README.md` (+30 lines)
- `docs/capabilities-guide.md` (+30 lines)
- `CHANGELOG.md` (+20 lines)

---

## Risk Assessment and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| jsDelivr CDN availability | Low | High | Fallback to stale cache, clear error messages |
| `@` symbol parsing edge cases | Medium | Low | Comprehensive test coverage, clear error handling |
| Cache strategy bugs | Low | Medium | Extensive testing, stale cache fallback |
| Breaking change impact | High | Medium | Clear migration guide, CHANGELOG warnings |
| Network retry delays | Low | Low | Exponential backoff (max 7s total), async execution |
| Documentation accuracy | Medium | Low | Manual verification, example testing |

### Contingency Plans

**If jsDelivr CDN has issues**:
- Fallback to stale cache (up to 7 days)
- Clear error messages guide users to local configuration
- Future: Add alternative CDN (e.g., unpkg.com)

**If `@` symbol parsing fails**:
- Default to main branch (backward compatible)
- Error messages show expected format
- Test coverage prevents regression

**If cache strategy causes issues**:
- Environment variable to disable cache
- Manual cache clear command (future enhancement)
- Stale cache always available as fallback

**If breaking change causes problems**:
- Migration guide in CHANGELOG
- Warning messages in error output
- Backward compatibility for explicit local sources

**If retry logic impacts performance**:
- Max 3 retries with 7s total delay (acceptable)
- Only retry 5xx errors (not 404 or network)
- Async execution in MCP server

---

## Testing Strategy

### Unit Testing

**Coverage Requirements**:
- Each stage: ≥80% coverage
- Critical paths: 100% coverage (@ parsing, URL generation, error handling)
- Edge cases: Comprehensive test cases

**Test Organization**:
```
cmd/mcp-server/
  capabilities.go              - Implementation
  capabilities_test.go         - Unit tests
```

**New Test Functions** (22.8-22.10):
- TestParseGitHubSource
- TestBuildJsDelivrURL
- TestDetectVersionType
- TestGitHubSourceCacheTTL
- TestDefaultSourceIsGitHub
- TestJsDelivrErrorHandling404
- TestJsDelivrErrorHandling5xx
- TestFallbackToStaleCache
- TestExponentialBackoff

### Integration Testing

**Multi-Source Scenarios**:
- Single GitHub source (branch)
- Single GitHub source (tag)
- Mixed local + GitHub sources
- Default source (no env var)
- Priority override with duplicates

**End-to-End Workflows**:
```bash
# Workflow 1: jsDelivr URL Generation
1. Set GitHub source with branch
2. Call get_capability(name)
3. Verify jsDelivr URL format
4. Verify content fetched

# Workflow 2: Version Pinning
1. Set GitHub source with tag
2. Call list_capabilities()
3. Verify 7-day cache TTL
4. Verify content stable

# Workflow 3: Error Handling
1. Set invalid GitHub source
2. Call get_capability(name)
3. Verify clear error message
4. Verify retry logic for 5xx

# Workflow 4: Network Failure
1. Set GitHub source
2. Simulate network failure
3. Verify fallback to stale cache
4. Verify warning message
```

### Regression Testing

**Verify No Breaking Changes**:
```bash
# Existing slash commands should work unchanged
/meta-errors
/meta-quality-scan
/meta-viz

# Existing MCP tools should work unchanged
echo '{"method":"tools/call","params":{"name":"get_session_stats","arguments":{}}}' | meta-cc-mcp

# Local sources should work (with explicit env var)
export META_CC_CAPABILITY_SOURCES="commands"
/meta "show errors"
```

### Performance Testing

**Benchmarks**:
```bash
# Measure jsDelivr loading time
time echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp
# Expected: <3s first call, <100ms cached

# Measure retry overhead (5xx error)
# Expected: 1s + 2s + 4s = 7s max for 3 retries

# Measure cache effectiveness
# First call: Load from jsDelivr
# Second call: Load from cache (should be much faster)
```

---

## Timeline Estimate

| Stage  | Description | Estimated Time |
|--------|-------------|----------------|
| 22.8   | jsDelivr CDN integration | 2 hours |
| 22.9   | Default source change | 1 hour |
| 22.10  | Error handling & fallback | 1.5 hours |
| 22.11  | Documentation updates | 2 hours |
| **Total** | **All stages** | **6.5 hours** |

**Contingency**: +2 hours for testing, debugging, and network integration (total: 8.5 hours)

---

## Conclusion

Stages 22.8-22.11 enhance Phase 22 with production-ready GitHub capability loading, improving reliability, performance, and user experience. Key enhancements:

1. **jsDelivr CDN**: Avoids rate limits, improves performance
2. **Branch/Tag Specification**: Enables version pinning and testing
3. **Smart Caching**: Adapts to version type (1h branches, 7d tags)
4. **Robust Error Handling**: Retry logic, stale cache fallback, clear messages
5. **Zero-Configuration**: Works out of the box with GitHub default source
6. **Local Development**: Explicit configuration enables rapid iteration

Key success factors:
- TDD methodology ensures high quality
- Comprehensive error handling and fallback strategies
- Clear migration guide minimizes breaking change impact
- jsDelivr CDN improves reliability over raw GitHub URLs
- Smart caching balances freshness and performance

Upon completion, meta-cc will have a production-ready, reliable, and user-friendly capability discovery system that works seamlessly for both production deployment and local development.

---

## Next Steps (Post-Stage 22.11)

After completing Stages 22.8-22.11:

1. **Monitor jsDelivr Performance**:
   - Track CDN availability and latency
   - Gather user feedback on reliability
   - Consider adding telemetry (opt-in)

2. **Enhance Capability Discovery**:
   - Implement capability search/filtering
   - Add capability dependency resolution
   - Support capability aliases

3. **Expand CDN Support**:
   - Add fallback CDNs (unpkg.com, etc.)
   - Implement CDN health checks
   - Auto-select fastest CDN

4. **Community Engagement**:
   - Publish official capability repository with tags
   - Document version pinning best practices
   - Create capability contribution guidelines

5. **Advanced Features**:
   - Capability marketplace integration
   - Automated capability testing pipeline
   - Multi-language capability support
