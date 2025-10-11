# Phase 22: Unified Meta Command & Multi-Source Capability Discovery - TDD Implementation Plan

## Phase Overview

**Objective**: Implement a unified `/meta` slash command that dynamically discovers and executes capabilities from multiple sources (local directories + GitHub repos) using semantic matching.

**Code Volume**: ~800 lines | **Priority**: High | **Status**: Planning

**Dependencies**:
- Phase 0-21 (Complete meta-cc CLI + MCP Server + all existing features)
- Existing 13 slash commands (meta-errors, meta-quality-scan, etc.)
- MCP server capabilities (16 tools)

**Deliverables**:
- Multi-source capability discovery system
- Two new MCP tools: `list_capabilities()` and `get_capability(name)`
- Unified `/meta` slash command replacing 13 individual commands
- Frontmatter metadata for all existing slash commands
- Support for single and composite capability execution
- Local development mode with real-time reflection

---

## Phase Objectives

### Core Problems

**Problem 1: Command Proliferation and Maintenance Overhead**
- Current: 13 separate slash commands (meta-errors, meta-quality-scan, meta-viz, etc.)
- Impact: High maintenance cost, users must remember multiple command names
- Need: Unified entry point with natural language intent matching

**Problem 2: Limited Extensibility**
- Current: Capabilities hardcoded in repository
- Missing: Dynamic capability loading from multiple sources
- Need: Plugin-like architecture supporting local development and community extensions

**Problem 3: Discovery and Composition**
- Current: No semantic capability matching
- Missing: Ability to combine capabilities for complex workflows
- Need: Claude-powered semantic matching and capability orchestration

**Problem 4: Development Workflow Friction**
- Current: Changes require repository commits and cache invalidation
- Missing: Real-time reflection for local development
- Need: Local development mode with instant feedback

### Solution Architecture

```
Phase 22 Implementation Strategy:

1. Multi-Source Capability Loading (Stage 22.1)
   - Environment variable: META_CC_CAPABILITY_SOURCES="source1:source2:source3"
   - Auto-detect source type (local path vs GitHub repo)
   - Priority-based merging (left-to-right, same-name capabilities override)
   - Frontmatter parsing (name, description, keywords, category)

2. MCP Capability Index Tool (Stage 22.2)
   - Tool: list_capabilities()
   - Returns compact capability index (name, description, keywords, category)
   - Local cache (1-hour TTL, disabled for local sources)
   - Hidden test parameters: _sources, _disable_cache

3. MCP Capability Retrieval Tool (Stage 22.3)
   - Tool: get_capability(name)
   - Multi-source search with priority order
   - Returns complete .md file content

4. Frontmatter Enhancement (Stage 22.4)
   - Add metadata to all 13 existing slash commands
   - Fields: name, description, keywords, category
   - Validation: ensure all frontmatter is valid YAML

5. Unified /meta Command (Stage 22.5)
   - Single entry point: /meta "natural language intent"
   - Call list_capabilities() to get capability index
   - Semantic matching using keyword scoring
   - Execute single capability via get_capability()

6. Composite Capability Execution (Stage 22.6)
   - Detect multiple high-scoring capabilities
   - Pipeline patterns: data → visualization
   - Predefined compositions: error+git, quality+viz

7. Testing and Documentation (Stage 22.7)
   - Unit tests for multi-source merging, priority override
   - Integration tests for local, GitHub, mixed sources
   - Documentation: README, developer guide, examples
```

### Design Principles

1. **Natural Language Interface**: Users describe intent, Claude handles capability selection
2. **Multi-Source Support**: Local paths and GitHub repos with unified API
3. **Priority-Based Merging**: Left-to-right source priority, same-name capabilities override
4. **Local Development Mode**: Real-time reflection without cache for local sources
5. **Backward Compatibility**: Existing slash commands continue to work
6. **Semantic Matching**: Claude performs keyword-based scoring and selection
7. **Extensibility**: Community can fork, extend, and contribute capabilities

---

## Success Criteria

**Functional Acceptance**:
- ✅ All stage unit tests pass (TDD methodology)
- ✅ Multi-source capability loading works (local + GitHub)
- ✅ `list_capabilities()` MCP tool returns valid capability index
- ✅ `get_capability(name)` MCP tool retrieves capabilities correctly
- ✅ Source priority merging works (same-name capabilities override)
- ✅ `/meta` command performs semantic matching and execution
- ✅ Local development mode reflects changes immediately
- ✅ Composite capability execution works for common patterns

**Integration Acceptance**:
- ✅ MCP server returns valid responses for new tools
- ✅ `/meta` command works with natural language input
- ✅ Existing slash commands still functional (backward compatibility)
- ✅ Local and GitHub sources load correctly
- ✅ Cache behavior correct (TTL for remote, no cache for local)

**Code Quality**:
- ✅ Total code: ~800 lines (within Phase 22 budget)
  - Stage 22.1: ~200 lines (multi-source foundation)
  - Stage 22.2: ~150 lines (list_capabilities tool)
  - Stage 22.3: ~100 lines (get_capability tool)
  - Stage 22.4: ~80 lines (frontmatter updates)
  - Stage 22.5: ~200 lines (unified /meta command)
  - Stage 22.6: ~100 lines (composite execution)
  - Stage 22.7: ~70 lines (documentation)
- ✅ Each stage ≤ 200 lines
- ✅ Test coverage: ≥ 80%
- ✅ `make all` passes after each stage

---

## Stage 22.1: Multi-Source Capability Discovery Foundation

### Objective

Implement the core capability loading system that supports multiple sources (local directories and GitHub repositories) with priority-based merging and frontmatter parsing.

### Acceptance Criteria

- [ ] Environment variable `META_CC_CAPABILITY_SOURCES` parsed correctly (colon-separated)
- [ ] Auto-detection works for local paths vs GitHub repos (format: `owner/repo` or `owner/repo/subdir`)
- [ ] Frontmatter parsing extracts: name, description, keywords (comma-separated), category
- [ ] Source priority merging works (left-to-right, same-name overrides)
- [ ] Local sources detected and flagged (for cache control)
- [ ] Edge cases handled: empty sources, invalid paths, malformed frontmatter
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `cmd/mcp-server/capabilities_test.go` (~150 lines)

```go
// Test functions:
// - TestParseCapabilitySources - Parse env var into source list
// - TestDetectSourceType - Distinguish local vs GitHub
// - TestParseFrontmatter - Extract metadata from .md files
// - TestSourcePriorityMerging - Verify override behavior
// - TestLoadLocalCapabilities - Load from local directory
// - TestLoadGitHubCapabilities - Load from GitHub repo (mocked)
// - TestInvalidSourceHandling - Error handling for bad sources
// - TestEmptySourcesHandling - Handle empty source list
```

**Test Strategy**:
1. Create test fixtures with sample .md files (valid and invalid frontmatter)
2. Mock GitHub API responses for capability loading
3. Test source parsing with various formats (local paths, GitHub repos)
4. Verify priority merging with duplicate capability names
5. Test edge cases (empty, nil, malformed data)

**Implementation File**: `cmd/mcp-server/capabilities.go` (~200 lines)

```go
// Core structures:
type CapabilitySource struct {
    Type     SourceType  // "local" or "github"
    Location string      // "/path/to/dir" or "owner/repo/subdir"
    Priority int         // Left-to-right priority (0 = highest)
}

type SourceType string
const (
    SourceTypeLocal  SourceType = "local"
    SourceTypeGitHub SourceType = "github"
)

type CapabilityMetadata struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Keywords    []string `yaml:"keywords"`  // Parsed from comma-separated string
    Category    string   `yaml:"category"`
    Source      string   `json:"source"`    // Source identifier for debugging
    FilePath    string   `json:"file_path"` // Relative path within source
}

type CapabilityIndex map[string]CapabilityMetadata  // name → metadata

// Core functions:
// - parseCapabilitySources(envVar string) []CapabilitySource
// - detectSourceType(location string) SourceType
// - parseFrontmatter(content string) (CapabilityMetadata, error)
// - loadLocalCapabilities(path string) ([]CapabilityMetadata, error)
// - loadGitHubCapabilities(repo string) ([]CapabilityMetadata, error)
// - mergeSources(sources []CapabilitySource) (CapabilityIndex, error)

// Example implementation structure:
func parseCapabilitySources(envVar string) []CapabilitySource {
    if envVar == "" {
        return []CapabilitySource{}
    }

    parts := strings.Split(envVar, ":")
    sources := make([]CapabilitySource, 0, len(parts))

    for i, location := range parts {
        location = strings.TrimSpace(location)
        if location == "" {
            continue
        }

        sources = append(sources, CapabilitySource{
            Type:     detectSourceType(location),
            Location: location,
            Priority: i,
        })
    }

    return sources
}

func detectSourceType(location string) SourceType {
    // GitHub format: "owner/repo" or "owner/repo/subdir"
    if !strings.HasPrefix(location, "/") && !strings.HasPrefix(location, ".") {
        // Simple heuristic: contains "/" without leading "/" or "."
        if strings.Contains(location, "/") {
            return SourceTypeGitHub
        }
    }
    return SourceTypeLocal
}

func parseFrontmatter(content string) (CapabilityMetadata, error) {
    // Extract frontmatter (--- ... ---)
    // Parse YAML
    // Parse keywords from comma-separated string to []string
    // Return CapabilityMetadata
}

func loadLocalCapabilities(path string) ([]CapabilityMetadata, error) {
    // Scan directory for *.md files
    // Parse frontmatter from each file
    // Return list of capabilities
}

func loadGitHubCapabilities(repo string) ([]CapabilityMetadata, error) {
    // Use GitHub API to list files
    // Fetch .md files and parse frontmatter
    // Return list of capabilities
}

func mergeSources(sources []CapabilitySource) (CapabilityIndex, error) {
    index := make(CapabilityIndex)

    // Process sources in priority order (low to high)
    // Higher priority (lower index) overwrites lower priority
    for i := len(sources) - 1; i >= 0; i-- {
        source := sources[i]

        var capabilities []CapabilityMetadata
        var err error

        switch source.Type {
        case SourceTypeLocal:
            capabilities, err = loadLocalCapabilities(source.Location)
        case SourceTypeGitHub:
            capabilities, err = loadGitHubCapabilities(source.Location)
        }

        if err != nil {
            return nil, fmt.Errorf("failed to load source %s: %w", source.Location, err)
        }

        // Merge capabilities (same name overrides)
        for _, cap := range capabilities {
            cap.Source = source.Location
            index[cap.Name] = cap
        }
    }

    return index, nil
}
```

**Environment Variable Format**:
```bash
# Single local source
META_CC_CAPABILITY_SOURCES="~/.config/meta-cc/capabilities"

# Multiple sources (local + GitHub)
META_CC_CAPABILITY_SOURCES="~/dev/my-caps:yaleh/meta-cc-capabilities"

# Priority order (left = highest priority)
META_CC_CAPABILITY_SOURCES="~/dev/test:yaleh/meta-cc:community/extras"
```

**Frontmatter Format**:
```yaml
---
name: meta-errors
description: Analyze error patterns and prevention recommendations.
keywords: error, debug, troubleshooting, diagnostics
category: diagnostics
---
```

### File Changes

**New Files**:
- `cmd/mcp-server/capabilities.go` (+200 lines)
- `cmd/mcp-server/capabilities_test.go` (+150 lines)

**Total**: ~350 lines (exceeds 200-line stage target, but foundation stage needs comprehensive implementation)

### Test Commands

```bash
# Run Stage 22.1 tests
go test -v ./cmd/mcp-server -run TestParseCapabilitySources
go test -v ./cmd/mcp-server -run TestDetectSourceType
go test -v ./cmd/mcp-server -run TestParseFrontmatter
go test -v ./cmd/mcp-server -run TestSourcePriorityMerging
go test -v ./cmd/mcp-server -run TestLoad.*Capabilities

# Test with real data
export META_CC_CAPABILITY_SOURCES=".claude/commands"
go test -v ./cmd/mcp-server -run TestMergeSources

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test with various source configurations
3. Verify frontmatter parsing with valid and invalid YAML
4. Test priority merging with duplicate names
5. **HALT if tests fail after 2 fix attempts**

### Dependencies

None (foundation stage)

### Estimated Time

3 hours (350 lines implementation + tests)

---

## Stage 22.2: MCP Capability Index Tool

### Objective

Implement the `list_capabilities()` MCP tool that returns a compact capability index from all configured sources with caching support.

### Acceptance Criteria

- [ ] MCP tool `list_capabilities()` registered and functional
- [ ] Returns compact JSON with name, description, keywords, category for all capabilities
- [ ] Caching implemented with 1-hour TTL for remote sources
- [ ] Local sources bypass cache (always fresh)
- [ ] Hidden test parameters `_sources` and `_disable_cache` work
- [ ] Tool description accurate and complete
- [ ] Integration tests pass
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `cmd/mcp-server/tools_test.go` (additions ~40 lines)

```go
// Add test functions:
// - TestToolsListContainsListCapabilities - Verify tool registration
// - TestListCapabilitiesToolDefinition - Tool schema validation
```

**Test File**: `cmd/mcp-server/executor_test.go` (additions ~30 lines)

```go
// Add test functions:
// - TestExecuteListCapabilities - Basic execution
// - TestListCapabilitiesCaching - Verify cache behavior
// - TestListCapabilitiesLocalBypass - Local sources skip cache
// - TestListCapabilitiesTestParams - Hidden parameters work
```

**Test File**: `cmd/mcp-server/integration_test.go` (additions ~30 lines)

```go
// Add test functions:
// - TestIntegrationListCapabilities - End-to-end test
// - TestIntegrationListCapabilitiesMultiSource - Multiple sources
```

**Implementation Files**:

1. `cmd/mcp-server/capabilities.go` (+80 lines)

```go
// Add caching support:
type CapabilityCache struct {
    Index     CapabilityIndex
    Timestamp time.Time
    TTL       time.Duration
}

var globalCapabilityCache *CapabilityCache
var cacheMutex sync.RWMutex

// Core functions:
// - getCapabilityIndex(sources []CapabilitySource, disableCache bool) (CapabilityIndex, error)
// - isCacheValid(sources []CapabilitySource) bool
// - hasLocalSources(sources []CapabilitySource) bool

func getCapabilityIndex(sources []CapabilitySource, disableCache bool) (CapabilityIndex, error) {
    // Check cache if enabled
    if !disableCache && !hasLocalSources(sources) && isCacheValid(sources) {
        cacheMutex.RLock()
        defer cacheMutex.RUnlock()
        return globalCapabilityCache.Index, nil
    }

    // Load fresh data
    index, err := mergeSources(sources)
    if err != nil {
        return nil, err
    }

    // Update cache (only if no local sources)
    if !hasLocalSources(sources) {
        cacheMutex.Lock()
        globalCapabilityCache = &CapabilityCache{
            Index:     index,
            Timestamp: time.Now(),
            TTL:       1 * time.Hour,
        }
        cacheMutex.Unlock()
    }

    return index, nil
}

func hasLocalSources(sources []CapabilitySource) bool {
    for _, source := range sources {
        if source.Type == SourceTypeLocal {
            return true
        }
    }
    return false
}

func isCacheValid(sources []CapabilitySource) bool {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()

    if globalCapabilityCache == nil {
        return false
    }

    age := time.Since(globalCapabilityCache.Timestamp)
    return age < globalCapabilityCache.TTL
}
```

2. `cmd/mcp-server/tools.go` (+50 lines)

```go
// Add to listTools() function:
{
    Name: "list_capabilities",
    Description: "List all available capabilities from configured sources. Returns compact capability index with name, description, keywords, and category. Supports multi-source loading with priority-based merging.",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "_sources": {
                "type": "string",
                "description": "(Hidden test parameter) Override META_CC_CAPABILITY_SOURCES environment variable"
            },
            "_disable_cache": {
                "type": "boolean",
                "description": "(Hidden test parameter) Bypass cache and force fresh load",
                "default": false
            }
        }
    }`),
},
```

3. `cmd/mcp-server/executor.go` (+30 lines)

```go
// Add case to executeTool() switch statement:
case "list_capabilities":
    // Parse sources (test override or environment variable)
    sourcesEnv := os.Getenv("META_CC_CAPABILITY_SOURCES")
    if override, ok := params["_sources"].(string); ok && override != "" {
        sourcesEnv = override
    }

    // Parse sources
    sources := parseCapabilitySources(sourcesEnv)
    if len(sources) == 0 {
        // Default to .claude/commands if no sources configured
        sources = []CapabilitySource{
            {Type: SourceTypeLocal, Location: ".claude/commands", Priority: 0},
        }
    }

    // Check cache control
    disableCache := false
    if disable, ok := params["_disable_cache"].(bool); ok {
        disableCache = disable
    }

    // Get capability index
    index, err := getCapabilityIndex(sources, disableCache)
    if err != nil {
        return nil, fmt.Errorf("failed to get capability index: %w", err)
    }

    // Convert to array for JSON output
    capabilities := make([]CapabilityMetadata, 0, len(index))
    for _, cap := range index {
        capabilities = append(capabilities, cap)
    }

    // Sort by name for deterministic output
    sort.Slice(capabilities, func(i, j int) bool {
        return capabilities[i].Name < capabilities[j].Name
    })

    // Return inline (capability index is small)
    return map[string]interface{}{
        "mode":         "inline",
        "capabilities": capabilities,
        "source_count": len(sources),
    }, nil
```

**MCP Response Format**:
```json
{
  "mode": "inline",
  "capabilities": [
    {
      "name": "meta-errors",
      "description": "Analyze error patterns and prevention recommendations.",
      "keywords": ["error", "debug", "troubleshooting", "diagnostics"],
      "category": "diagnostics",
      "source": ".claude/commands",
      "file_path": "meta-errors.md"
    },
    {
      "name": "meta-quality-scan",
      "description": "Quick quality assessment of recent work.",
      "keywords": ["quality", "assessment", "code-review"],
      "category": "assessment",
      "source": ".claude/commands",
      "file_path": "meta-quality-scan.md"
    }
  ],
  "source_count": 1
}
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/capabilities.go` (+80 lines)
- `cmd/mcp-server/tools.go` (+50 lines)
- `cmd/mcp-server/executor.go` (+30 lines)
- `cmd/mcp-server/tools_test.go` (+40 lines)
- `cmd/mcp-server/executor_test.go` (+30 lines)
- `cmd/mcp-server/integration_test.go` (+30 lines)

**Total**: ~260 lines (exceeds 150-line target, but includes comprehensive tests)

### Test Commands

```bash
# Run MCP server tests
go test -v ./cmd/mcp-server -run TestToolsListContains.*Capabilities
go test -v ./cmd/mcp-server -run TestListCapabilities.*
go test -v ./cmd/mcp-server -run TestIntegrationListCapabilities

# Manual MCP testing
export META_CC_CAPABILITY_SOURCES=".claude/commands"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# Test with override
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{"_sources":"tests/fixtures/capabilities"}}}' | meta-cc-mcp

# Test cache disable
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{"_disable_cache":true}}}' | meta-cc-mcp

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test MCP tool registration
3. Verify cache behavior (TTL, local bypass)
4. Test with multiple sources
5. Verify hidden test parameters work
6. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 22.1 (multi-source capability loading)

### Estimated Time

2 hours (160 lines implementation + 100 lines tests)

---

## Stage 22.3: MCP Capability Retrieval Tool

### Objective

Implement the `get_capability(name)` MCP tool that retrieves the complete content of a capability from the configured sources.

### Acceptance Criteria

- [ ] MCP tool `get_capability(name)` registered and functional
- [ ] Multi-source search with priority order works
- [ ] Returns complete .md file content
- [ ] Supports both local and GitHub sources
- [ ] Handles missing capabilities gracefully
- [ ] Tool description accurate and complete
- [ ] Integration tests pass
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `cmd/mcp-server/tools_test.go` (additions ~20 lines)

```go
// Add test functions:
// - TestToolsListContainsGetCapability - Verify tool registration
// - TestGetCapabilityToolDefinition - Tool schema validation
```

**Test File**: `cmd/mcp-server/executor_test.go` (additions ~30 lines)

```go
// Add test functions:
// - TestExecuteGetCapability - Basic execution
// - TestGetCapabilityMultiSource - Priority-based search
// - TestGetCapabilityNotFound - Missing capability handling
```

**Test File**: `cmd/mcp-server/integration_test.go` (additions ~30 lines)

```go
// Add test functions:
// - TestIntegrationGetCapability - End-to-end test
// - TestIntegrationGetCapabilityPriority - Verify override behavior
```

**Implementation Files**:

1. `cmd/mcp-server/capabilities.go` (+80 lines)

```go
// Core functions:
// - getCapabilityContent(name string, sources []CapabilitySource) (string, error)
// - readLocalCapability(name string, path string) (string, error)
// - readGitHubCapability(name string, repo string) (string, error)

func getCapabilityContent(name string, sources []CapabilitySource) (string, error) {
    // Search sources in priority order (left-to-right = high-to-low)
    for _, source := range sources {
        var content string
        var err error

        switch source.Type {
        case SourceTypeLocal:
            content, err = readLocalCapability(name, source.Location)
        case SourceTypeGitHub:
            content, err = readGitHubCapability(name, source.Location)
        }

        if err == nil {
            return content, nil
        }

        // If error is "not found", continue to next source
        // If error is other (network, permission), return error
        if !isNotFoundError(err) {
            return "", fmt.Errorf("failed to read from source %s: %w", source.Location, err)
        }
    }

    return "", fmt.Errorf("capability not found: %s", name)
}

func readLocalCapability(name string, path string) (string, error) {
    // Expand path (handle ~ and relative paths)
    expandedPath := expandPath(path)

    // Construct file path
    filePath := filepath.Join(expandedPath, name+".md")

    // Read file
    content, err := os.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return "", newNotFoundError(name)
        }
        return "", err
    }

    return string(content), nil
}

func readGitHubCapability(name string, repo string) (string, error) {
    // Parse repo format: "owner/repo" or "owner/repo/subdir"
    parts := strings.SplitN(repo, "/", 3)
    if len(parts) < 2 {
        return "", fmt.Errorf("invalid GitHub repo format: %s", repo)
    }

    owner := parts[0]
    repoName := parts[1]
    subdir := ""
    if len(parts) > 2 {
        subdir = parts[2]
    }

    // Construct GitHub raw URL
    // https://raw.githubusercontent.com/owner/repo/main/subdir/name.md
    filePath := name + ".md"
    if subdir != "" {
        filePath = subdir + "/" + filePath
    }

    url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repoName, filePath)

    // Fetch content
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode == 404 {
        return "", newNotFoundError(name)
    }

    if resp.StatusCode != 200 {
        return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
    }

    content, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(content), nil
}
```

2. `cmd/mcp-server/tools.go` (+25 lines)

```go
// Add to listTools() function:
{
    Name: "get_capability",
    Description: "Retrieve the complete content of a capability by name. Searches configured sources in priority order (left-to-right) and returns the first match. Returns the complete .md file content including frontmatter.",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "name": {
                "type": "string",
                "description": "Name of the capability to retrieve (without .md extension)"
            },
            "_sources": {
                "type": "string",
                "description": "(Hidden test parameter) Override META_CC_CAPABILITY_SOURCES environment variable"
            }
        },
        "required": ["name"]
    }`),
},
```

3. `cmd/mcp-server/executor.go` (+40 lines)

```go
// Add case to executeTool() switch statement:
case "get_capability":
    // Get capability name
    name, ok := params["name"].(string)
    if !ok || name == "" {
        return nil, fmt.Errorf("missing required parameter: name")
    }

    // Parse sources (test override or environment variable)
    sourcesEnv := os.Getenv("META_CC_CAPABILITY_SOURCES")
    if override, ok := params["_sources"].(string); ok && override != "" {
        sourcesEnv = override
    }

    // Parse sources
    sources := parseCapabilitySources(sourcesEnv)
    if len(sources) == 0 {
        // Default to .claude/commands if no sources configured
        sources = []CapabilitySource{
            {Type: SourceTypeLocal, Location: ".claude/commands", Priority: 0},
        }
    }

    // Get capability content
    content, err := getCapabilityContent(name, sources)
    if err != nil {
        return nil, fmt.Errorf("failed to get capability: %w", err)
    }

    // Return inline (capability content is typically <50KB)
    return map[string]interface{}{
        "mode":    "inline",
        "name":    name,
        "content": content,
    }, nil
```

**MCP Response Format**:
```json
{
  "mode": "inline",
  "name": "meta-errors",
  "content": "---\nname: meta-errors\ndescription: Analyze error patterns...\n---\n\nλ(scope) → error_insights..."
}
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/capabilities.go` (+80 lines)
- `cmd/mcp-server/tools.go` (+25 lines)
- `cmd/mcp-server/executor.go` (+40 lines)
- `cmd/mcp-server/tools_test.go` (+20 lines)
- `cmd/mcp-server/executor_test.go` (+30 lines)
- `cmd/mcp-server/integration_test.go` (+30 lines)

**Total**: ~225 lines (exceeds 100-line target, but includes comprehensive tests)

### Test Commands

```bash
# Run MCP server tests
go test -v ./cmd/mcp-server -run TestGetCapability.*
go test -v ./cmd/mcp-server -run TestIntegrationGetCapability

# Manual MCP testing
export META_CC_CAPABILITY_SOURCES=".claude/commands"
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"meta-errors"}}}' | meta-cc-mcp

# Test priority override
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"meta-errors","_sources":"tests/fixtures/caps1:.claude/commands"}}}' | meta-cc-mcp

# Test not found
echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"nonexistent"}}}' | meta-cc-mcp

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test capability retrieval from local sources
3. Test capability retrieval from GitHub (mocked)
4. Verify priority-based search
5. Test error handling (not found, network errors)
6. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 22.1 (multi-source capability loading)
- Stage 22.2 (capability index tool)

### Estimated Time

1.5 hours (145 lines implementation + 80 lines tests)

---

## Stage 22.4: Frontmatter Enhancement for Existing Slash Commands

### Objective

Add structured metadata (frontmatter) to all 13 existing slash commands to enable semantic matching and capability discovery.

### Acceptance Criteria

- [ ] All 13 .md files in `.claude/commands/` have valid frontmatter
- [ ] Frontmatter includes: name, description, keywords (comma-separated), category
- [ ] Keywords cover key use cases and semantic variations
- [ ] Categories consistent across similar commands
- [ ] Frontmatter validates as correct YAML
- [ ] Existing command functionality unchanged
- [ ] No breaking changes to command execution

### Frontmatter Categories

**Diagnostics**:
- meta-errors
- meta-bugs

**Assessment**:
- meta-quality-scan
- meta-tech-debt

**Visualization**:
- meta-viz
- meta-timeline

**Analysis**:
- meta-architecture
- meta-habits
- meta-focus-analyzer

**Guidance**:
- meta-coach
- meta-guide
- meta-next
- meta-prompt

### Implementation

Update all 13 slash commands with frontmatter. Example structure:

**1. meta-errors.md** (already has frontmatter, verify/enhance):
```yaml
---
name: meta-errors
description: Analyze error patterns and prevention recommendations.
keywords: error, debug, troubleshooting, diagnostics, failure, exception, crash
category: diagnostics
---
```

**2. meta-quality-scan.md**:
```yaml
---
name: meta-quality-scan
description: Quick quality assessment of recent work with scorecard and improvement recommendations.
keywords: quality, assessment, code-review, technical-debt, best-practices, standards
category: assessment
---
```

**3. meta-viz.md**:
```yaml
---
name: meta-viz
description: Create ASCII dashboards and charts from any analysis data.
keywords: visualization, dashboard, chart, graph, metrics, reporting
category: visualization
---
```

**4. meta-timeline.md**:
```yaml
---
name: meta-timeline
description: Visualize project evolution timeline with workflow events.
keywords: timeline, history, evolution, progression, chronology, activity
category: visualization
---
```

**5. meta-architecture.md**:
```yaml
---
name: meta-architecture
description: Analyze architecture evolution, module stability, and structural decisions.
keywords: architecture, design, structure, modularity, dependencies, coupling
category: analysis
---
```

**6. meta-habits.md**:
```yaml
---
name: meta-habits
description: Analyze work patterns and productivity habits insights.
keywords: habits, patterns, productivity, workflow, efficiency, behavior
category: analysis
---
```

**7. meta-focus-analyzer.md**:
```yaml
---
name: meta-focus-analyzer
description: Analyze attention patterns and focus distribution across projects and files.
keywords: focus, attention, concentration, context-switching, multitasking
category: analysis
---
```

**8. meta-coach.md**:
```yaml
---
name: meta-coach
description: Get workflow optimization recommendations and coaching insights.
keywords: coaching, optimization, improvement, guidance, recommendations, advice
category: guidance
---
```

**9. meta-guide.md**:
```yaml
---
name: meta-guide
description: Get intelligent guidance and prioritized action recommendations.
keywords: guide, help, recommendations, actions, next-steps, priorities
category: guidance
---
```

**10. meta-next.md**:
```yaml
---
name: meta-next
description: Generate ready-to-use prompts for natural next steps (no MCP execution).
keywords: next-steps, continuation, prompts, suggestions, follow-up
category: guidance
---
```

**11. meta-prompt.md**:
```yaml
---
name: meta-prompt
description: Refine prompts using successful patterns from project history.
keywords: prompt, refinement, optimization, effectiveness, clarity
category: guidance
---
```

**12. meta-bugs.md**:
```yaml
---
name: meta-bugs
description: Analyze project-level bugs, workflow failures, and fix patterns using meta-cognitive analysis.
keywords: bugs, defects, issues, failures, root-cause, fix-patterns
category: diagnostics
---
```

**13. meta-tech-debt.md**:
```yaml
---
name: meta-tech-debt
description: Track technical debt accumulation, repayment rate, and prioritized remediation plan.
keywords: technical-debt, refactoring, code-smell, maintenance, cleanup
category: assessment
---
```

### File Changes

**Modified Files** (13 files):
1. `.claude/commands/meta-errors.md` (verify existing frontmatter)
2. `.claude/commands/meta-quality-scan.md` (+6 lines frontmatter)
3. `.claude/commands/meta-viz.md` (+6 lines frontmatter)
4. `.claude/commands/meta-timeline.md` (+6 lines frontmatter)
5. `.claude/commands/meta-architecture.md` (+6 lines frontmatter)
6. `.claude/commands/meta-habits.md` (+6 lines frontmatter)
7. `.claude/commands/meta-focus-analyzer.md` (+6 lines frontmatter)
8. `.claude/commands/meta-coach.md` (+6 lines frontmatter)
9. `.claude/commands/meta-guide.md` (+6 lines frontmatter)
10. `.claude/commands/meta-next.md` (+6 lines frontmatter)
11. `.claude/commands/meta-prompt.md` (+6 lines frontmatter)
12. `.claude/commands/meta-bugs.md` (verify existing frontmatter)
13. `.claude/commands/meta-tech-debt.md` (verify existing frontmatter)

**Total**: ~78 lines (13 files × 6 lines average)

### Verification Commands

```bash
# Validate all frontmatter
for file in .claude/commands/meta-*.md; do
    echo "Validating $file..."
    # Extract frontmatter (--- ... ---)
    sed -n '/^---$/,/^---$/p' "$file" | sed '1d;$d' | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)"
done

# Check required fields
for file in .claude/commands/meta-*.md; do
    echo "Checking $file..."
    sed -n '/^---$/,/^---$/p' "$file" | grep -q "name:" && echo "  ✓ name"
    sed -n '/^---$/,/^---$/p' "$file" | grep -q "description:" && echo "  ✓ description"
    sed -n '/^---$/,/^---$/p' "$file" | grep -q "keywords:" && echo "  ✓ keywords"
    sed -n '/^---$/,/^---$/p' "$file" | grep -q "category:" && echo "  ✓ category"
done

# Test capability index loading
export META_CC_CAPABILITY_SOURCES=".claude/commands"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp | jq '.capabilities | length'
# Should return 13
```

### Testing Protocol

**After Implementation**:
1. Validate all frontmatter YAML syntax
2. Verify all required fields present
3. Test `list_capabilities()` returns all 13 capabilities
4. Verify existing slash commands still work
5. Check keywords cover semantic variations
6. **HALT if validation fails after 2 fix attempts**

### Dependencies

- Stage 22.2 (list_capabilities tool)

### Estimated Time

1.5 hours (78 lines + validation)

---

## Stage 22.5: Unified /meta Slash Command

### Objective

Create the unified `/meta` slash command that accepts natural language intent, performs semantic matching against capabilities, and executes the best match.

### Acceptance Criteria

- [ ] `/meta` command accepts natural language input
- [ ] Calls `list_capabilities()` to get capability index
- [ ] Performs keyword-based scoring for semantic matching
- [ ] Executes single capability via `get_capability()`
- [ ] Displays capability output inline
- [ ] Handles no-match case (shows available capabilities)
- [ ] Handles errors gracefully
- [ ] Command description clear and helpful
- [ ] Integration tests pass

### TDD Approach

**Test Strategy**:
1. Test with various natural language inputs
2. Verify semantic matching accuracy
3. Test edge cases (no match, multiple high scores)
4. Verify capability execution
5. Manual testing in Claude Code environment

**Test Cases**:
```
User input: "show errors" → Matches: meta-errors (score: 5)
User input: "quality check" → Matches: meta-quality-scan (score: 4)
User input: "visualize timeline" → Matches: meta-timeline (score: 4)
User input: "architecture analysis" → Matches: meta-architecture (score: 3)
User input: "asdfqwer" → No match (show available capabilities)
```

**Implementation File**: `.claude/commands/meta.md` (~200 lines)

```markdown
---
name: meta
description: Unified meta-cognition command with semantic capability matching. Accepts natural language intent and automatically selects the best capability to execute.
keywords: meta, capability, semantic, match, intent, unified, command
category: unified
---

λ(intent) → capability_execution | ∀capability ∈ available_capabilities:

intent :: string  # Natural language description of user intent

# Step 1: Discover available capabilities
discover :: void → CapabilityIndex
discover() = {
  result: mcp_meta_cc.list_capabilities(),

  if result.error then
    error("Failed to load capabilities: " + result.error)

  capabilities: result.capabilities,

  log("Loaded " + len(capabilities) + " capabilities from " + result.source_count + " sources")
}

# Step 2: Semantic matching using keyword scoring
match :: (intent, CapabilityIndex) → ScoredCapabilities
match(I, C) = {
  # Tokenize intent (lowercase, split on spaces/punctuation)
  intent_tokens: tokenize(I.toLowerCase()),

  # Score each capability
  scores: [],
  for cap in C do {
    score: 0,

    # Check name match (high weight)
    if any(token in cap.name.toLowerCase() for token in intent_tokens):
      score += 3,

    # Check description match (medium weight)
    if any(token in cap.description.toLowerCase() for token in intent_tokens):
      score += 2,

    # Check keywords match (medium weight)
    for keyword in cap.keywords do {
      if any(token in keyword.toLowerCase() for token in intent_tokens):
        score += 1
    },

    # Check category match (low weight)
    if any(token in cap.category.toLowerCase() for token in intent_tokens):
      score += 1,

    if score > 0:
      scores.append({
        capability: cap,
        score: score
      })
  },

  # Sort by score (descending)
  scores.sort(key=lambda x: x.score, reverse=true),

  return scores
}

# Step 3: Execute capability
execute :: (capability_name) → output
execute(name) = {
  # Get full capability content
  result: mcp_meta_cc.get_capability(name=name),

  if result.error then
    error("Failed to get capability: " + result.error)

  content: result.content,

  # Extract and display capability description
  frontmatter: parse_frontmatter(content),
  log("Executing: " + frontmatter.name),
  log("Description: " + frontmatter.description),
  log(""),

  # Execute capability
  # The capability content is a Claude Code slash command
  # We can't directly "execute" it, but we can:
  # 1. Display its instructions
  # 2. Ask Claude to follow the instructions
  # 3. Let Claude call the MCP tools referenced in the capability

  say("I'll execute the **" + frontmatter.name + "** capability:"),
  say(frontmatter.description),
  say(""),
  say("---"),
  say(""),

  # Parse and execute the capability implementation
  # The capability is a markdown file with lambda calculus notation
  # Claude will interpret and execute it
  say(content),

  # Note: The actual execution happens through Claude interpreting
  # the capability content and calling the appropriate MCP tools
}

# Main workflow
main :: intent → void
main(I) = {
  # Step 1: Discover capabilities
  index: discover(),

  # Step 2: Semantic matching
  scored: match(I, index.capabilities),

  if len(scored) == 0 then {
    say("❌ No matching capabilities found for: " + I),
    say(""),
    say("Available capabilities:"),
    for cap in index.capabilities.sort_by(name) do {
      say("  • **" + cap.name + "**: " + cap.description),
      say("    Keywords: " + join(cap.keywords, ", ")),
      say("")
    },
    return
  },

  # Get best match
  best: scored[0],

  say("🎯 Matched capability: **" + best.capability.name + "** (score: " + best.score + ")"),
  say(""),

  # Show alternatives if multiple high scores
  if len(scored) > 1 and scored[1].score >= (best.score * 0.7) then {
    say("Other possible matches:"),
    for i in range(1, min(3, len(scored))) do {
      alt: scored[i],
      say("  • " + alt.capability.name + " (score: " + alt.score + ")")
    },
    say(""),
    say("Proceeding with best match: **" + best.capability.name + "**"),
    say("")
  },

  # Step 3: Execute capability
  execute(best.capability.name)
}

# Entry point
main(intent)
```

**Simplified Markdown Implementation** (for actual .md file):

```markdown
---
name: meta
description: Unified meta-cognition command with semantic capability matching. Accepts natural language intent and automatically selects the best capability to execute.
keywords: meta, capability, semantic, match, intent, unified, command
category: unified
---

# Unified Meta Command

This command accepts natural language intent and automatically discovers, matches, and executes the best capability.

## Step 1: Discover Capabilities

First, let me load all available capabilities:

```
Call mcp_meta_cc.list_capabilities()
```

Expected response format:
```json
{
  "capabilities": [
    {"name": "meta-errors", "description": "...", "keywords": [...], "category": "diagnostics"},
    ...
  ],
  "source_count": 1
}
```

## Step 2: Semantic Matching

I'll analyze your intent: **{{intent}}**

Matching algorithm:
1. Tokenize intent (lowercase, split on spaces)
2. Score each capability:
   - Name match: +3 points
   - Description match: +2 points
   - Keyword match: +1 point per keyword
   - Category match: +1 point
3. Sort by score (descending)
4. Select best match (score > 0)

## Step 3: Execute Capability

If match found:
1. Call `mcp_meta_cc.get_capability(name=best_match)`
2. Display capability info (name, description)
3. Execute capability instructions by interpreting the content

If no match found:
1. Display error message
2. List all available capabilities with descriptions
3. Suggest trying a different intent

## Implementation Notes

- Semantic matching uses keyword-based scoring (not ML)
- Claude performs the actual capability execution by interpreting the .md content
- Capabilities can call MCP tools, read files, analyze data, etc.
- The /meta command is a meta-layer over existing capabilities

## Usage Examples

```
/meta "show errors"
  → Executes meta-errors capability

/meta "quality check"
  → Executes meta-quality-scan capability

/meta "visualize timeline"
  → Executes meta-timeline capability

/meta "help me improve my workflow"
  → Executes meta-coach capability
```
```

### File Changes

**New Files**:
- `.claude/commands/meta.md` (+200 lines)

**Total**: ~200 lines

### Test Commands

```bash
# Validate frontmatter
sed -n '/^---$/,/^---$/p' .claude/commands/meta.md | sed '1d;$d' | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)"

# Manual testing in Claude Code
# 1. Run: /meta "show errors"
# 2. Run: /meta "quality check"
# 3. Run: /meta "visualize timeline"
# 4. Run: /meta "asdfqwer" (should show available capabilities)
```

### Testing Protocol

**After Implementation**:
1. Validate frontmatter syntax
2. Test in Claude Code environment
3. Verify semantic matching accuracy
4. Test with various natural language inputs
5. Verify error handling (no match case)
6. **HALT if functionality broken after 2 fix attempts**

### Dependencies

- Stage 22.2 (list_capabilities tool)
- Stage 22.3 (get_capability tool)
- Stage 22.4 (frontmatter for all commands)

### Estimated Time

3 hours (200 lines + testing in Claude Code)

---

## Stage 22.6: Composite Capability Execution

### Objective

Extend the `/meta` command to detect and execute multiple capabilities in sequence for complex workflows (e.g., error analysis + visualization).

### Acceptance Criteria

- [ ] Detects multiple high-scoring capabilities (score ≥ 70% of best)
- [ ] Implements pipeline patterns (data → visualization)
- [ ] Predefined compositions work: error+viz, quality+viz, bugs+git
- [ ] User can confirm before executing composite
- [ ] Handles errors in multi-step execution
- [ ] Documentation updated with composite examples
- [ ] Integration tests pass

### Pipeline Patterns

**Pattern 1: Analysis + Visualization**
```
User: "show errors with visualization"
  → Matched: meta-errors (5) + meta-viz (3)
  → Pipeline: meta-errors | meta-viz
  → Output: Error analysis + ASCII dashboard
```

**Pattern 2: Assessment + Remediation**
```
User: "quality scan with recommendations"
  → Matched: meta-quality-scan (4) + meta-coach (2)
  → Pipeline: meta-quality-scan | meta-coach
  → Output: Quality report + improvement recommendations
```

**Pattern 3: Diagnostics + Context**
```
User: "analyze bugs with git context"
  → Matched: meta-bugs (4) + (implicit git analysis)
  → Pipeline: meta-bugs (includes git context)
  → Output: Bug analysis with commit history
```

### Implementation

Update `.claude/commands/meta.md` (+100 lines):

```markdown
## Composite Capability Execution

When multiple capabilities match your intent, I can execute them in sequence for comprehensive analysis.

### Detection

Composite execution triggers when:
1. Multiple capabilities score ≥ 70% of best match
2. Intent includes multiple keywords (e.g., "errors" + "visualize")
3. Capabilities belong to compatible categories (analysis + visualization)

### Pipeline Patterns

**Pattern 1: Data → Visualization**
- Categories: diagnostics/analysis/assessment → visualization
- Example: meta-errors → meta-viz
- Workflow: Generate data, then visualize it

**Pattern 2: Analysis → Guidance**
- Categories: diagnostics/assessment → guidance
- Example: meta-quality-scan → meta-coach
- Workflow: Analyze state, then provide recommendations

**Pattern 3: Multi-Source Context**
- Categories: diagnostics + (implicit context)
- Example: meta-bugs (includes git context)
- Workflow: Combine multiple data sources

### Execution Flow

```
1. Detect composite intent
   ↓
2. Rank and filter capabilities (score ≥ 70% of best)
   ↓
3. Determine pipeline pattern
   ↓
4. Ask user for confirmation:
   "I found multiple matching capabilities:
    • meta-errors (score: 5)
    • meta-viz (score: 3)
    Execute them in sequence? (yes/no)"
   ↓
5. Execute pipeline:
   a. Run first capability (collect data)
   b. Pass data to second capability (transform/visualize)
   c. Display combined output
   ↓
6. Handle errors:
   - If step 1 fails, abort
   - If step 2 fails, show partial results
```

### Predefined Compositions

**Composition 1: Error Analysis + Visualization**
```
Intent: "show errors with charts"
Pipeline: meta-errors → meta-viz
Steps:
  1. Run meta-errors (collect error data)
  2. Extract key metrics (error count, patterns, frequency)
  3. Run meta-viz (create ASCII dashboard)
  4. Display combined output
```

**Composition 2: Quality Scan + Coaching**
```
Intent: "quality check with recommendations"
Pipeline: meta-quality-scan → meta-coach
Steps:
  1. Run meta-quality-scan (assess code quality)
  2. Extract improvement areas
  3. Run meta-coach (generate recommendations)
  4. Display combined output
```

**Composition 3: Bug Analysis + Git Context**
```
Intent: "analyze bugs with history"
Pipeline: meta-bugs (built-in git context)
Steps:
  1. Run meta-bugs (includes git commit analysis)
  2. Display bug patterns + related commits
```

### User Confirmation

Before executing composite capabilities:

```
🔍 Detected composite intent: "{{intent}}"

Matched capabilities:
  1. **meta-errors** (score: 5) - Analyze error patterns
  2. **meta-viz** (score: 3) - Create visualizations

Proposed pipeline:
  meta-errors → meta-viz

This will:
  1. Analyze error patterns in your project
  2. Generate an ASCII dashboard with error metrics

Execute this pipeline? (yes/no)
```

### Implementation Code

```python
# Composite detection
def detect_composite(scored_capabilities, threshold=0.7):
    if len(scored_capabilities) < 2:
        return None

    best_score = scored_capabilities[0].score
    candidates = [
        cap for cap in scored_capabilities
        if cap.score >= (best_score * threshold)
    ]

    if len(candidates) < 2:
        return None

    return candidates

# Pipeline pattern detection
def detect_pipeline_pattern(capabilities):
    categories = [cap.capability.category for cap in capabilities]

    # Pattern 1: Data → Visualization
    if any(c in ["diagnostics", "analysis", "assessment"] for c in categories) and \
       "visualization" in categories:
        return "data_to_viz"

    # Pattern 2: Analysis → Guidance
    if any(c in ["diagnostics", "assessment"] for c in categories) and \
       "guidance" in categories:
        return "analysis_to_guidance"

    # Default: Sequential
    return "sequential"

# Execute pipeline
def execute_pipeline(capabilities, pattern):
    if pattern == "data_to_viz":
        # Run data capability first
        data_cap = next(c for c in capabilities if c.capability.category != "visualization")
        viz_cap = next(c for c in capabilities if c.capability.category == "visualization")

        result_data = execute(data_cap.capability.name)
        result_viz = execute_with_data(viz_cap.capability.name, result_data)

        return {
            "data": result_data,
            "visualization": result_viz
        }

    elif pattern == "analysis_to_guidance":
        # Run analysis first, then guidance
        analysis_cap = capabilities[0]
        guidance_cap = capabilities[1]

        result_analysis = execute(analysis_cap.capability.name)
        result_guidance = execute_with_context(guidance_cap.capability.name, result_analysis)

        return {
            "analysis": result_analysis,
            "guidance": result_guidance
        }

    else:
        # Sequential execution
        results = []
        for cap in capabilities:
            result = execute(cap.capability.name)
            results.append(result)
        return results
```

### Error Handling

```
If step 1 fails:
  ❌ Failed to execute meta-errors: <error message>
  Aborting pipeline.

If step 2 fails:
  ⚠️ Warning: meta-viz execution failed: <error message>
  Showing partial results from meta-errors:
  <error analysis output>
```

### Usage Examples

```
/meta "show errors with visualization"
  → Composite: meta-errors + meta-viz
  → Confirmation prompt
  → Execute pipeline
  → Display combined output

/meta "quality check and recommendations"
  → Composite: meta-quality-scan + meta-coach
  → Confirmation prompt
  → Execute pipeline
  → Display combined output
```
```

### File Changes

**Modified Files**:
- `.claude/commands/meta.md` (+100 lines)

**Total**: ~100 lines

### Test Commands

```bash
# Manual testing in Claude Code
# 1. Run: /meta "show errors with charts"
#    - Should detect composite (meta-errors + meta-viz)
#    - Should ask for confirmation
#    - Should execute pipeline

# 2. Run: /meta "quality check with advice"
#    - Should detect composite (meta-quality-scan + meta-coach)
#    - Should execute pipeline

# 3. Run: /meta "simple error check"
#    - Should execute single capability (meta-errors)
#    - Should NOT trigger composite
```

### Testing Protocol

**After Implementation**:
1. Test composite detection with various intents
2. Verify user confirmation prompts
3. Test pipeline execution for each pattern
4. Verify error handling in multi-step execution
5. Test fallback to single capability
6. **HALT if functionality broken after 2 fix attempts**

### Dependencies

- Stage 22.5 (unified /meta command)

### Estimated Time

2 hours (100 lines + testing)

---

## Stage 22.7: Testing and Documentation

### Objective

Create comprehensive tests for the multi-source capability system and update all relevant documentation.

### Acceptance Criteria

- [ ] Unit tests for multi-source merging pass
- [ ] Unit tests for priority override pass
- [ ] Integration tests for local sources pass
- [ ] Integration tests for GitHub sources pass (mocked)
- [ ] Integration tests for mixed sources pass
- [ ] Documentation complete: README, CLAUDE.md, developer guide
- [ ] Usage examples clear and functional
- [ ] Test coverage ≥80% for new code
- [ ] All documentation accurate and consistent

### Test Implementation

**Test File**: `cmd/mcp-server/capabilities_integration_test.go` (~120 lines)

```go
// Integration test functions:
// - TestIntegrationMultiSourceLocal - Load from multiple local directories
// - TestIntegrationMultiSourceGitHub - Load from GitHub repo (mocked)
// - TestIntegrationMixedSources - Combine local + GitHub
// - TestIntegrationPriorityOverride - Verify same-name override
// - TestIntegrationCacheBehavior - Verify cache TTL and local bypass
// - TestIntegrationEndToEnd - Complete workflow: list + get + execute
```

**Test Strategy**:
1. Create test fixtures with sample capabilities in multiple directories
2. Mock GitHub API for remote capability loading
3. Test all combinations of source types
4. Verify priority merging with duplicate names
5. Test cache behavior (TTL, local bypass)
6. End-to-end integration test

### Documentation Updates

**1. CLAUDE.md** (+40 lines)

Add section "Unified Meta Command":

```markdown
## Unified Meta Command

Phase 22 introduces the `/meta` command - a unified entry point for all meta-cognition capabilities.

### Usage

```
/meta "natural language intent"
```

Examples:
```
/meta "show errors"           → Executes meta-errors
/meta "quality check"         → Executes meta-quality-scan
/meta "visualize timeline"    → Executes meta-timeline
/meta "analyze architecture"  → Executes meta-architecture
```

### How It Works

1. **Discovery**: Loads capabilities from configured sources
2. **Matching**: Semantic keyword matching scores each capability
3. **Execution**: Runs best match (or composite of multiple matches)

### Multi-Source Configuration

Configure capability sources via environment variable:

```bash
# Single local source
export META_CC_CAPABILITY_SOURCES="~/.config/meta-cc/capabilities"

# Multiple sources (priority: left-to-right)
export META_CC_CAPABILITY_SOURCES="~/dev/my-caps:yaleh/meta-cc-capabilities"

# Mix local and GitHub
export META_CC_CAPABILITY_SOURCES="~/dev/test:.claude/commands:community/extras"
```

### Composite Execution

The `/meta` command can execute multiple capabilities in sequence:

```
/meta "show errors with visualization"
  → Composite: meta-errors + meta-viz
  → Pipeline: Analyze errors, then create dashboard
```

### MCP Tools

New MCP tools for capability discovery:

- `list_capabilities()` - Get capability index from all sources
- `get_capability(name)` - Retrieve complete capability content

### Local Development

For capability development, use local sources (no cache):

```bash
export META_CC_CAPABILITY_SOURCES="~/dev/capabilities:.claude/commands"
# Changes reflect immediately without cache invalidation
```
```

**2. README.md** (+30 lines)

Add section "Unified Meta Command":

```markdown
## Unified Meta Command

Use natural language to invoke meta-cognition capabilities:

```bash
/meta "show errors"           # Error analysis
/meta "quality check"         # Code quality scan
/meta "visualize timeline"    # Project timeline
```

### Multi-Source Capabilities

Load capabilities from multiple sources:

```bash
export META_CC_CAPABILITY_SOURCES="~/my-caps:.claude/commands:yaleh/meta-cc-extras"
```

Supports:
- Local directories (immediate reflection, no cache)
- GitHub repositories (1-hour cache)
- Priority-based merging (left = highest priority)

See [docs/capabilities-guide.md](docs/capabilities-guide.md) for development guide.
```

**3. docs/capabilities-guide.md** (new, ~150 lines)

```markdown
# Capability Development Guide

This guide explains how to create and extend meta-cc capabilities using the multi-source discovery system.

## Capability Structure

A capability is a markdown file with frontmatter metadata:

```markdown
---
name: my-capability
description: Short description of what this capability does.
keywords: keyword1, keyword2, keyword3
category: diagnostics
---

# Capability Implementation

Your capability implementation here...
Can include:
- MCP tool calls
- File operations
- Data analysis
- Visualization
```

### Frontmatter Fields

- **name**: Unique capability identifier (kebab-case, required)
- **description**: One-sentence description (required)
- **keywords**: Comma-separated keywords for semantic matching (required)
- **category**: Category for grouping (required)
  - Values: diagnostics, assessment, visualization, analysis, guidance

## Local Development Workflow

1. **Create capability directory**:
   ```bash
   mkdir -p ~/dev/my-capabilities
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

3. **Configure source**:
   ```bash
   export META_CC_CAPABILITY_SOURCES="~/dev/my-capabilities:.claude/commands"
   ```

4. **Test capability**:
   ```bash
   # List capabilities (verify yours appears)
   echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

   # Get capability content
   echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"my-feature"}}}' | meta-cc-mcp

   # Use via /meta command
   /meta "my feature"
   ```

5. **Iterate**:
   - Edit capability file
   - Changes reflect immediately (no cache for local sources)
   - Test with `/meta` command

## Publishing Capabilities

### Method 1: GitHub Repository

1. Create GitHub repo: `username/meta-cc-capabilities`
2. Add capabilities: `capabilities/my-feature.md`
3. Users install via:
   ```bash
   export META_CC_CAPABILITY_SOURCES="username/meta-cc-capabilities/capabilities"
   ```

### Method 2: Fork and PR

1. Fork `yaleh/meta-cc`
2. Add capability: `.claude/commands/meta-my-feature.md`
3. Submit PR
4. After merge, available to all users

## Best Practices

1. **Clear frontmatter**: Accurate description and keywords
2. **Keywords**: Include synonyms and common variations
3. **Category**: Choose appropriate category for grouping
4. **Documentation**: Include usage examples in capability
5. **Testing**: Test with various natural language intents
6. **MCP tools**: Use existing MCP tools for data access
7. **Composition**: Design capabilities that can combine with others

## Example Capability

```markdown
---
name: meta-dependencies
description: Analyze project dependencies and detect security issues.
keywords: dependencies, npm, security, vulnerabilities, packages
category: assessment
---

# Dependency Analysis

This capability analyzes project dependencies for:
- Outdated packages
- Security vulnerabilities
- License issues
- Circular dependencies

## Implementation

1. **Detect package manager**:
   - Check for package.json (npm)
   - Check for go.mod (Go)
   - Check for requirements.txt (Python)

2. **Analyze dependencies**:
   ```
   Call mcp_meta_cc.query_tools(tool="Bash", pattern="npm|go|pip")
   ```

3. **Security scan**:
   - Run npm audit (if npm)
   - Check for known CVEs
   - Report vulnerabilities

4. **Recommendations**:
   - List outdated packages
   - Suggest security updates
   - Recommend version pinning
```

## Multi-Source Priority

When same-name capabilities exist in multiple sources, left-most source wins:

```bash
# Priority: my-dev > official
export META_CC_CAPABILITY_SOURCES="~/dev/test:.claude/commands"
```

Use cases:
- Test capability changes before PR
- Override official capability with custom version
- Fork and customize capabilities

## Troubleshooting

**Capability not found**:
- Verify frontmatter is valid YAML
- Check filename matches frontmatter `name` field
- Verify source path in META_CC_CAPABILITY_SOURCES

**Semantic matching fails**:
- Add more keywords to frontmatter
- Use exact capability name: `/meta "meta-my-capability"`
- Check keyword spelling

**Cache not updating**:
- Local sources bypass cache automatically
- GitHub sources: wait 1 hour or use `_disable_cache: true`

## Testing

Test your capability:

```bash
# Unit test: Parse frontmatter
sed -n '/^---$/,/^---$/p' my-capability.md | sed '1d;$d' | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)"

# Integration test: List capabilities
export META_CC_CAPABILITY_SOURCES="~/dev/my-caps"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp | jq '.capabilities[] | select(.name=="my-capability")'

# Semantic matching test
/meta "my capability keywords"
# Should match your capability
```
```

**4. CHANGELOG.md** (+10 lines)

```markdown
## [Unreleased]

### Added (Phase 22)
- Unified `/meta` slash command with semantic capability matching
- Multi-source capability discovery (local directories + GitHub repos)
- MCP tools: `list_capabilities()` and `get_capability(name)`
- Frontmatter metadata for all 13 existing slash commands
- Composite capability execution (pipeline patterns)
- Local development mode with real-time reflection
- Capability development guide (docs/capabilities-guide.md)
```

### File Changes

**New Files**:
- `cmd/mcp-server/capabilities_integration_test.go` (+120 lines)
- `docs/capabilities-guide.md` (+150 lines)

**Modified Files**:
- `CLAUDE.md` (+40 lines)
- `README.md` (+30 lines)
- `CHANGELOG.md` (+10 lines)

**Total**: ~350 lines (exceeds 70-line target significantly, but documentation is critical)

### Verification Commands

```bash
# Run integration tests
go test -v ./cmd/mcp-server -run TestIntegration.*

# Test coverage
go test -cover ./cmd/mcp-server

# Verify documentation
grep -r "list_capabilities" docs/ CLAUDE.md README.md
grep -r "get_capability" docs/ CLAUDE.md README.md

# Check cross-references
grep -r "capabilities-guide.md" docs/ CLAUDE.md README.md
```

### Testing Protocol

**After Documentation**:
1. Run all integration tests
2. Verify test coverage ≥80%
3. Review all documentation for accuracy
4. Test examples in capabilities-guide.md
5. Check cross-references between documents
6. Verify CHANGELOG.md completeness

### Dependencies

- All previous stages (22.1-22.6)

### Estimated Time

2 hours (270 lines implementation + 80 lines documentation)

---

## Phase Integration Strategy

### Build Verification

After completing all stages, verify the complete Phase 22 implementation:

```bash
# 1. Full build
make all

# 2. Unit tests
go test -v ./...

# 3. Integration tests
go test -v ./cmd/mcp-server -run TestIntegration.*

# 4. Test coverage
make test-coverage

# 5. MCP tools functional tests
export META_CC_CAPABILITY_SOURCES=".claude/commands"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp | jq '.capabilities | length'
# Should return 14 (13 existing + 1 meta)

echo '{"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"meta-errors"}}}' | meta-cc-mcp | jq '.content' | head

# 6. Manual testing in Claude Code
# - Test /meta "show errors"
# - Test /meta "quality check"
# - Test /meta "visualize timeline"
# - Test /meta "show errors with charts" (composite)
```

### Multi-Source Testing

Test with various source configurations:

```bash
# Test 1: Single local source
export META_CC_CAPABILITY_SOURCES=".claude/commands"
/meta "show errors"

# Test 2: Multiple local sources
mkdir -p /tmp/test-caps
cp .claude/commands/meta-errors.md /tmp/test-caps/
export META_CC_CAPABILITY_SOURCES="/tmp/test-caps:.claude/commands"
/meta "show errors"  # Should load from /tmp/test-caps (higher priority)

# Test 3: GitHub source (requires network)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc"
/meta "show errors"

# Test 4: Mixed sources
export META_CC_CAPABILITY_SOURCES="~/dev/test:.claude/commands:yaleh/meta-cc-extras"
/meta "show errors"
```

### Rollout Checklist

Before marking Phase 22 complete:

- [ ] All 7 stages completed and tested
- [ ] `make all` passes without errors
- [ ] Test coverage ≥80% (verify with `make test-coverage`)
- [ ] Documentation complete and accurate
- [ ] MCP tools `list_capabilities` and `get_capability` functional
- [ ] Unified `/meta` command works with natural language input
- [ ] Semantic matching accuracy acceptable (manual testing)
- [ ] Composite execution works for common patterns
- [ ] Multi-source loading works (local + GitHub)
- [ ] Priority override behavior correct
- [ ] Local development mode bypasses cache
- [ ] Backward compatibility verified (existing commands still work)
- [ ] CHANGELOG.md updated
- [ ] Git commit includes Phase 22 changes

---

## File Change Inventory

### Summary by Stage

| Stage | New Files | Modified Files | Total Lines |
|-------|-----------|----------------|-------------|
| 22.1  | 2         | 0              | ~350        |
| 22.2  | 0         | 6              | ~260        |
| 22.3  | 0         | 6              | ~225        |
| 22.4  | 0         | 13             | ~78         |
| 22.5  | 1         | 0              | ~200        |
| 22.6  | 0         | 1              | ~100        |
| 22.7  | 2         | 3              | ~350        |
| **Total** | **5** | **29** | **~1563** |

**Note**: Total exceeds 800-line target by ~763 lines. Breakdown:
- Core implementation (22.1-22.3): ~835 lines (foundation + MCP tools + tests)
- Frontmatter updates (22.4): ~78 lines (metadata for 13 commands)
- Unified command (22.5-22.6): ~300 lines (semantic matching + composite)
- Documentation (22.7): ~350 lines (comprehensive guide + integration tests)

**Justification**: Phase 22 is a major architectural addition requiring:
1. Multi-source capability loading system (complex)
2. Comprehensive testing (integration tests for multiple scenarios)
3. Extensive documentation (developer guide for community extensions)

### Detailed File Changes

**New Files (5)**:
1. `cmd/mcp-server/capabilities.go` (280 lines: core + caching + retrieval)
2. `cmd/mcp-server/capabilities_test.go` (150 lines: unit tests)
3. `.claude/commands/meta.md` (300 lines: unified command + composite)
4. `cmd/mcp-server/capabilities_integration_test.go` (120 lines)
5. `docs/capabilities-guide.md` (150 lines)

**Modified Files (29)**:
1. `cmd/mcp-server/tools.go` (+75 lines: 2 new tools)
2. `cmd/mcp-server/executor.go` (+70 lines: 2 tool executors)
3. `cmd/mcp-server/tools_test.go` (+60 lines)
4. `cmd/mcp-server/executor_test.go` (+60 lines)
5. `cmd/mcp-server/integration_test.go` (+60 lines)
6-18. `.claude/commands/meta-*.md` (+78 lines: frontmatter for 13 commands)
19. `CLAUDE.md` (+40 lines)
20. `README.md` (+30 lines)
21. `CHANGELOG.md` (+10 lines)

---

## Risk Assessment and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| GitHub API rate limiting | Medium | Medium | Implement cache (1-hour TTL), add retry logic |
| Frontmatter parsing errors | Medium | Low | Comprehensive YAML validation, clear error messages |
| Semantic matching inaccuracy | High | Medium | Iterative keyword refinement, allow exact name matching |
| Cache invalidation issues | Low | Medium | Local sources bypass cache, test parameter for manual invalidation |
| Composite execution complexity | Medium | Medium | Start with simple patterns, expand iteratively |
| Documentation becomes outdated | Medium | Low | Version-specific docs, automated examples |

### Contingency Plans

**If GitHub API rate limiting occurs**:
- Implement exponential backoff retry
- Increase cache TTL to 2 hours
- Document GitHub token setup for higher rate limits

**If semantic matching fails frequently**:
- Add fuzzy matching for typos
- Allow exact name matching fallback: `/meta "meta-errors"`
- Show top 3 matches for user selection

**If composite execution is too complex**:
- Start with single capability only
- Add composite in Phase 22.6 as optional enhancement
- Defer advanced pipeline patterns to future phases

**If testing fails repeatedly**:
- HALT development per testing protocol
- Document blockers and failure state
- Simplify implementation (remove GitHub support temporarily)

**If cache causes stale data issues**:
- Reduce TTL to 15 minutes
- Add manual cache clear command
- Improve cache key to include source modification time

---

## Testing Strategy

### Unit Testing

**Coverage Requirements**:
- Each stage: ≥80% coverage
- Critical paths: 100% coverage (frontmatter parsing, source merging)
- Edge cases: Comprehensive test cases

**Test Organization**:
```
cmd/mcp-server/
  capabilities.go                   - Core implementation
  capabilities_test.go              - Unit tests
  capabilities_integration_test.go  - Integration tests
  tools_test.go                     - Tool registration tests
  executor_test.go                  - Execution tests
```

### Integration Testing

**Multi-Source Scenarios**:
- Single local source
- Multiple local sources
- Single GitHub source (mocked)
- Mixed local + GitHub sources
- Priority override with duplicate names

**End-to-End Workflows**:
```bash
# Workflow 1: Discovery + Retrieval
1. Set META_CC_CAPABILITY_SOURCES
2. Call list_capabilities()
3. Call get_capability(name)
4. Verify content matches source

# Workflow 2: Semantic Matching
1. User: /meta "show errors"
2. System: list_capabilities()
3. System: Semantic matching (score capabilities)
4. System: get_capability(best_match)
5. System: Execute capability
6. User: See results

# Workflow 3: Composite Execution
1. User: /meta "show errors with charts"
2. System: Detect composite (meta-errors + meta-viz)
3. System: Ask confirmation
4. User: Confirm
5. System: Execute pipeline
6. User: See combined results
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

# Backward compatibility
# (all 16 existing MCP tools + 2 new tools = 18 total)
```

### Performance Testing

**Benchmarks**:
```bash
# Measure capability loading time
time echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# Measure cache effectiveness
# (second call should be significantly faster)

# Measure semantic matching time
# (should be <100ms for 13 capabilities)
```

---

## Post-Phase Verification

### Functional Verification

After completing Phase 22, verify:

1. **Multi-Source Loading**:
   ```bash
   export META_CC_CAPABILITY_SOURCES=".claude/commands:/tmp/test-caps"
   echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp | jq '.capabilities | length'
   # Should return count from both sources
   ```

2. **Semantic Matching**:
   ```bash
   # Test various intents
   /meta "show errors"
   /meta "quality check"
   /meta "visualize timeline"
   /meta "architecture analysis"
   ```

3. **Composite Execution**:
   ```bash
   /meta "show errors with visualization"
   # Should detect composite and ask for confirmation
   ```

4. **Cache Behavior**:
   ```bash
   # First call (should load fresh)
   time echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

   # Second call (should use cache)
   time echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp
   # Should be faster
   ```

5. **Local Development Mode**:
   ```bash
   # Edit local capability
   echo "---\nname: test\ndescription: Test\nkeywords: test\ncategory: diagnostics\n---\n# Test" > /tmp/test-caps/test.md

   # Should reflect immediately
   export META_CC_CAPABILITY_SOURCES="/tmp/test-caps"
   echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp | jq '.capabilities[] | select(.name=="test")'
   ```

### Documentation Verification

1. **Check Tool Count Consistency**:
   ```bash
   grep -r "16 tools" docs/ CLAUDE.md
   # Should find references (needs update to 18 tools)
   ```

2. **Verify Examples Work**:
   ```bash
   # Test examples from docs/capabilities-guide.md
   # Verify they produce expected output
   ```

3. **Check Cross-References**:
   ```bash
   # Verify all documentation cross-references are valid
   grep -r "capabilities-guide.md" docs/ CLAUDE.md README.md
   ```

### Integration Verification

1. **Backward Compatibility**:
   - Existing 13 slash commands work unchanged
   - Existing 16 MCP tools work unchanged
   - Output formats remain consistent

2. **New Functionality**:
   - New MCP tools return expected data
   - Unified `/meta` command performs semantic matching
   - Composite execution works for common patterns

3. **Performance**:
   - Capability loading <1s (local sources)
   - Semantic matching <100ms
   - Cache reduces repeated call latency by >50%

---

## Success Metrics

### Quantitative Metrics

- **Code Quality**:
  - Test coverage ≥ 80%
  - Zero linting errors (`make lint`)
  - Zero test failures (`make test`)

- **Functionality**:
  - 2 new MCP tools functional (list_capabilities, get_capability)
  - 18 total MCP tools available
  - 14 slash commands with frontmatter metadata
  - 1 unified `/meta` command

- **Performance**:
  - Capability loading <1s (local), <3s (GitHub)
  - Semantic matching <100ms
  - Cache hit reduces latency >50%

### Qualitative Metrics

- **Usability**:
  - Natural language interface intuitive
  - Error messages clear and helpful
  - Documentation comprehensive

- **Extensibility**:
  - Easy to add new capabilities
  - Multi-source support enables community extensions
  - Local development mode enables rapid iteration

- **Reliability**:
  - Handles edge cases gracefully (no match, network errors)
  - Cache behavior predictable
  - Backward compatibility maintained

---

## Timeline Estimate

| Stage | Description | Estimated Time |
|-------|-------------|----------------|
| 22.1  | Multi-source foundation | 3 hours |
| 22.2  | list_capabilities tool | 2 hours |
| 22.3  | get_capability tool | 1.5 hours |
| 22.4  | Frontmatter updates | 1.5 hours |
| 22.5  | Unified /meta command | 3 hours |
| 22.6  | Composite execution | 2 hours |
| 22.7  | Testing & documentation | 2 hours |
| **Total** | **All stages** | **15 hours** |

**Contingency**: +5 hours for testing, debugging, and iteration (total: 20 hours)

---

## Conclusion

Phase 22 represents a major architectural enhancement to meta-cc, transforming it from a collection of individual commands into a unified, extensible capability discovery system. The multi-source architecture enables:

1. **Natural Language Interface**: Users describe intent, Claude handles capability selection
2. **Extensibility**: Community can fork, extend, and contribute capabilities
3. **Local Development**: Real-time reflection for rapid iteration
4. **Composition**: Complex workflows through capability pipelines
5. **Backward Compatibility**: Existing commands continue to work

Key success factors:
- TDD methodology ensures high quality
- Multi-source loading with priority merging
- Semantic keyword matching (simple but effective)
- Local development mode (no cache)
- Comprehensive documentation for community

Upon completion, meta-cc will have a flexible, extensible architecture that supports community-driven growth while maintaining the simplicity and power of the existing slash commands.

---

## Next Steps (Post-Phase 22)

After Phase 22 completion:

1. **Community Engagement**:
   - Create example community repository: `meta-cc-capabilities`
   - Document contribution guidelines
   - Encourage capability submissions

2. **Enhanced Semantic Matching**:
   - Add fuzzy matching for typos
   - Implement ML-based intent classification (future phase)
   - Support multi-language intents

3. **Advanced Composition**:
   - Conditional execution (if-then patterns)
   - Parallel execution for independent capabilities
   - Data transformation between pipeline stages

4. **Marketplace**:
   - Curated capability marketplace
   - Rating and review system
   - Automated testing for submitted capabilities

5. **Plugin Architecture**:
   - Binary plugins (not just markdown)
   - Language-specific plugins (Python, Go, Node.js)
   - Sandboxed execution for untrusted capabilities
