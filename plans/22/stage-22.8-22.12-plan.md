# Phase 22 Stages 22.8-22.12: Capability Package Distribution - TDD Implementation Plan

## Overview

**Objective**: Implement package file distribution for capabilities, enabling offline-friendly, reliable capability loading via prebuilt `.tar.gz` bundles.

**Code Volume**: ~750 lines total | **Priority**: High | **Estimated Time**: 12 hours

**Dependencies**:
- Phase 22 Stages 22.1-22.7 (Complete multi-source capability discovery system)
- Existing implementation: `cmd/mcp-server/capabilities.go` and `capabilities_test.go`

**Key Design Decision**:
Instead of jsDelivr CDN approach, we use prebuilt package files (`.tar.gz`) distributed via GitHub Releases. This provides:
- **Offline-friendly**: Download once, cache locally
- **Reliable**: No CDN dependencies, no rate limits
- **Fast**: No network calls after initial download
- **Simple**: Standard tar.gz format, easy to verify and debug

**Package File Structure**:
```
capabilities-latest.tar.gz
├── commands/
│   ├── meta-errors.md
│   ├── meta-quality-scan.md
│   └── ...
└── agents/
    └── ... (if any)
```

**Source Format Examples**:
```bash
# GitHub Release package (default)
export META_CC_CAPABILITY_SOURCES="https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"

# Local package file
export META_CC_CAPABILITY_SOURCES="/path/to/capabilities.tar.gz"

# Mix package and directory sources
export META_CC_CAPABILITY_SOURCES="/local/caps.tar.gz:~/dev/caps"

# Still support GitHub repo (for development)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
```

**Deliverables**:
- Makefile target for building capability packages
- GitHub Actions workflow for automatic release uploads
- MCP support for package file sources
- Cache strategy with metadata tracking
- Smart fallback: package → GitHub raw → local cache
- Comprehensive documentation updates

---

## Success Criteria

**Functional Acceptance**:
- ✅ `make bundle-capabilities` creates valid `.tar.gz` package
- ✅ GitHub Actions automatically uploads package to releases
- ✅ Package URLs recognized and parsed correctly
- ✅ Package download, extraction, and caching works
- ✅ Cache metadata tracks package versions and TTL
- ✅ Default source changed to GitHub Release package
- ✅ Fallback strategy works: package → GitHub → local cache
- ✅ Documentation updated and accurate

**Code Quality**:
- ✅ Total code: ~750 lines (within budget)
  - Stage 22.8: ~100 lines (build + CI)
  - Stage 22.9: ~250 lines (package source support)
  - Stage 22.10: ~100 lines (cache strategy)
  - Stage 22.11: ~80 lines (default source + fallback)
  - Stage 22.12: ~120 lines (documentation)
- ✅ Each stage ≤ 250 lines
- ✅ Test coverage: ≥ 80%
- ✅ `make all` passes after each stage

---

## Stage 22.8: Capability Package Build & Release Automation

### Objective

Implement build tooling to package capabilities into `.tar.gz` bundles and automate uploading to GitHub Releases.

### Acceptance Criteria

- [ ] `make bundle-capabilities` creates valid `.tar.gz` package
- [ ] Package contains `commands/` and `agents/` directories
- [ ] Package filename: `capabilities-latest.tar.gz`
- [ ] GitHub Actions uploads package to releases
- [ ] Release asset available at stable URL: `releases/latest/download/capabilities-latest.tar.gz`
- [ ] Package extraction works correctly
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `Makefile` (manual verification), `.github/workflows/release.yml` (CI test)

**Test Strategy**:
1. Test `make bundle-capabilities` creates expected archive
2. Verify archive structure (commands/, agents/)
3. Test archive extraction produces correct files
4. Verify GitHub Actions workflow syntax
5. Integration test with real GitHub Release (manual)

**Implementation Details**:

#### 1. Makefile Target for Capability Packaging

**Location**: `Makefile` (add new targets)

```makefile
CAPABILITIES_DIR := capabilities
CAPABILITIES_ARCHIVE := capabilities-latest.tar.gz

.PHONY: bundle-capabilities clean-capabilities

bundle-capabilities:
	@echo "Creating capability package: $(CAPABILITIES_ARCHIVE)..."
	@if [ ! -d "$(CAPABILITIES_DIR)/commands" ]; then \
		echo "ERROR: $(CAPABILITIES_DIR)/commands/ directory not found"; \
		exit 1; \
	fi
	@mkdir -p $(BUILD_DIR)
	@tar -czf $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE) -C $(CAPABILITIES_DIR) commands agents 2>/dev/null || \
		tar -czf $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE) -C $(CAPABILITIES_DIR) commands
	@echo "✓ Package created: $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE)"
	@echo "  Size: $$(du -h $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE) | cut -f1)"
	@echo "  Files: $$(tar -tzf $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE) | wc -l)"

clean-capabilities:
	@echo "Cleaning capability packages..."
	@rm -f $(BUILD_DIR)/$(CAPABILITIES_ARCHIVE)

# Update clean target to include capabilities
clean: clean-capabilities
```

**Test Commands**:
```bash
# Test package creation
make bundle-capabilities

# Verify package structure
tar -tzf build/capabilities-latest.tar.gz | head -20

# Test extraction
mkdir -p /tmp/test-caps
tar -xzf build/capabilities-latest.tar.gz -C /tmp/test-caps
ls -la /tmp/test-caps/commands/
rm -rf /tmp/test-caps

# Verify package contains expected files
tar -tzf build/capabilities-latest.tar.gz | grep -E "commands/meta-.*\.md" | wc -l
# Should match number of capabilities in capabilities/commands/

# Test clean
make clean-capabilities
[ ! -f build/capabilities-latest.tar.gz ] && echo "✓ Clean successful"
```

#### 2. GitHub Actions Workflow Update

**Location**: `.github/workflows/release.yml` (modify existing workflow)

Add step to upload capability package to releases:

```yaml
# Insert after binary cross-compilation step
- name: Package Capabilities
  run: |
    make bundle-capabilities
    echo "Capability package created: build/capabilities-latest.tar.gz"
    ls -lh build/capabilities-latest.tar.gz

- name: Upload Release Assets
  uses: softprops/action-gh-release@v1
  if: startsWith(github.ref, 'refs/tags/')
  with:
    files: |
      build/*.tar.gz
      build/capabilities-latest.tar.gz
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Test Strategy**:
```bash
# Validate workflow syntax
yamllint .github/workflows/release.yml

# Test locally with act (if available)
act -j release --secret GITHUB_TOKEN=<token>

# Manual integration test:
# 1. Create test tag: git tag v0.0.1-test && git push origin v0.0.1-test
# 2. Verify GitHub Actions runs
# 3. Check release assets include capabilities-latest.tar.gz
# 4. Download and verify: curl -L https://github.com/yaleh/meta-cc/releases/download/v0.0.1-test/capabilities-latest.tar.gz -o test.tar.gz
# 5. Clean up: git tag -d v0.0.1-test && git push --delete origin v0.0.1-test
```

#### 3. Integration Test Script

**Location**: `tests/integration/test-capability-package.sh` (new file)

```bash
#!/bin/bash
# Integration test for capability package creation and extraction

set -e

BUILD_DIR="build"
PACKAGE="capabilities-latest.tar.gz"
TEST_DIR="/tmp/test-capability-package-$$"

echo "=== Testing Capability Package Creation ==="

# Clean previous builds
make clean-capabilities

# Build package
make bundle-capabilities

# Verify package exists
if [ ! -f "$BUILD_DIR/$PACKAGE" ]; then
    echo "ERROR: Package not created"
    exit 1
fi
echo "✓ Package created: $BUILD_DIR/$PACKAGE"

# Extract package
mkdir -p "$TEST_DIR"
tar -xzf "$BUILD_DIR/$PACKAGE" -C "$TEST_DIR"

# Verify structure
if [ ! -d "$TEST_DIR/commands" ]; then
    echo "ERROR: commands/ directory not found in package"
    exit 1
fi
echo "✓ Package structure valid"

# Count files
COMMAND_COUNT=$(find "$TEST_DIR/commands" -name "*.md" | wc -l)
if [ "$COMMAND_COUNT" -lt 1 ]; then
    echo "ERROR: No capability files found in package"
    exit 1
fi
echo "✓ Found $COMMAND_COUNT capability files"

# Verify frontmatter in sample file
SAMPLE_FILE=$(find "$TEST_DIR/commands" -name "*.md" | head -1)
if ! grep -q "^---$" "$SAMPLE_FILE"; then
    echo "WARNING: Sample file missing frontmatter: $SAMPLE_FILE"
fi

# Clean up
rm -rf "$TEST_DIR"
echo "✓ All tests passed"
```

**Make Integration**:
```makefile
.PHONY: test-capability-package

test-capability-package:
	@bash tests/integration/test-capability-package.sh
```

### File Changes

**Modified Files**:
- `Makefile` (+30 lines)
  - Add `bundle-capabilities` target (~15 lines)
  - Add `clean-capabilities` target (~5 lines)
  - Update `clean` target (~2 lines)
  - Add `test-capability-package` target (~3 lines)
  - Add documentation comments (~5 lines)

- `.github/workflows/release.yml` (+20 lines)
  - Add "Package Capabilities" step (~5 lines)
  - Update "Upload Release Assets" step (~10 lines)
  - Add comments (~5 lines)

**New Files**:
- `tests/integration/test-capability-package.sh` (+50 lines)

**Total**: ~100 lines

### Testing Protocol

**After Implementation**:
1. Run `make bundle-capabilities` and verify package created
2. Test extraction and verify structure
3. Run `tests/integration/test-capability-package.sh`
4. Validate GitHub Actions workflow syntax
5. Verify package size reasonable (<1MB)
6. **Manual**: Create test tag and verify GitHub Actions uploads package
7. **HALT if tests fail after 2 fix attempts**

### Dependencies

None (builds on existing infrastructure)

### Estimated Time

3 hours (build tooling + CI integration + testing)

---

## Stage 22.9: MCP Package File Source Support

### Objective

Extend MCP server to recognize, download, extract, and cache capability packages from URLs or local paths.

### Acceptance Criteria

- [ ] Package URLs (`.tar.gz`) recognized as `SourceTypePackage`
- [ ] Local package paths recognized
- [ ] Package download works (HTTP/HTTPS)
- [ ] Package extraction to cache directory
- [ ] `loadPackageCapabilities` reads capabilities from extracted package
- [ ] Cache directory: `.capabilities-cache/packages/<hash>/`
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `cmd/mcp-server/capabilities_test.go` (additions ~150 lines)

**Test Strategy**:
1. Test package URL detection (`.tar.gz` suffix)
2. Test package download (mock HTTP)
3. Test package extraction (tar.gz)
4. Test cache directory structure
5. Test loading capabilities from extracted package
6. Test local package file paths
7. Test mixed sources (package + directory)

**Implementation Details**:

#### 1. Extend SourceType Enum

**Location**: `cmd/mcp-server/capabilities.go`

```go
const (
	// SourceTypeLocal represents a local filesystem directory
	SourceTypeLocal SourceType = "local"
	// SourceTypeGitHub represents a GitHub repository
	SourceTypeGitHub SourceType = "github"
	// SourceTypePackage represents a tar.gz package file (URL or local path)
	SourceTypePackage SourceType = "package"
)
```

#### 2. Package Detection in parseCapabilitySources

**Location**: `cmd/mcp-server/capabilities.go`

```go
// detectSourceType determines the type of capability source
func detectSourceType(location string) SourceType {
	// Check for package file (tar.gz)
	if strings.HasSuffix(location, ".tar.gz") || strings.HasSuffix(location, ".tgz") {
		return SourceTypePackage
	}

	// Check for GitHub repository format
	if !strings.HasPrefix(location, "/") &&
		!strings.HasPrefix(location, "~") &&
		!strings.HasPrefix(location, ".") &&
		strings.Contains(location, "/") {
		return SourceTypeGitHub
	}

	// Default to local path
	return SourceTypeLocal
}
```

**Test Cases**:
```go
func TestDetectSourceType(t *testing.T) {
	tests := []struct {
		location string
		expected SourceType
	}{
		// Package sources
		{"https://example.com/caps.tar.gz", SourceTypePackage},
		{"/path/to/caps.tar.gz", SourceTypePackage},
		{"~/downloads/caps.tgz", SourceTypePackage},

		// GitHub sources
		{"yaleh/meta-cc", SourceTypeGitHub},
		{"yaleh/meta-cc@main/commands", SourceTypeGitHub},

		// Local sources
		{"/path/to/dir", SourceTypeLocal},
		{"~/config/caps", SourceTypeLocal},
		{"./relative/path", SourceTypeLocal},
	}

	for _, tt := range tests {
		result := detectSourceType(tt.location)
		if result != tt.expected {
			t.Errorf("detectSourceType(%q) = %v, want %v",
				tt.location, result, tt.expected)
		}
	}
}
```

#### 3. Package Cache Directory Management

**Location**: `cmd/mcp-server/capabilities.go`

```go
import (
	"crypto/sha256"
	"encoding/hex"
)

// getPackageCacheDir returns the cache directory for a package source
func getPackageCacheDir(packageLocation string) (string, error) {
	// Compute hash of package location for cache directory name
	hash := sha256.Sum256([]byte(packageLocation))
	hashStr := hex.EncodeToString(hash[:])[:16] // Use first 16 chars

	// Cache directory: ~/.capabilities-cache/packages/<hash>/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, CapabilityCacheDir, "packages", hashStr)
	return cacheDir, nil
}
```

**Test Cases**:
```go
func TestGetPackageCacheDir(t *testing.T) {
	tests := []struct {
		location string
		wantErr  bool
	}{
		{"https://example.com/caps.tar.gz", false},
		{"/local/path/caps.tar.gz", false},
	}

	for _, tt := range tests {
		cacheDir, err := getPackageCacheDir(tt.location)
		if (err != nil) != tt.wantErr {
			t.Errorf("getPackageCacheDir(%q) error = %v, wantErr %v",
				tt.location, err, tt.wantErr)
		}
		if !tt.wantErr && cacheDir == "" {
			t.Errorf("getPackageCacheDir(%q) returned empty path", tt.location)
		}
	}
}
```

#### 4. Package Download Function

**Location**: `cmd/mcp-server/capabilities.go`

```go
import (
	"net/http"
	"io"
)

// downloadPackage downloads a package from URL to destination path
func downloadPackage(url string, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
```

**Test Cases** (mock HTTP):
```go
func TestDownloadPackage(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	// Test download
	tmpFile := filepath.Join(t.TempDir(), "test.tar.gz")
	err := downloadPackage(server.URL, tmpFile)
	if err != nil {
		t.Fatalf("downloadPackage failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("downloaded content = %q, want %q", string(content), "test content")
	}
}
```

#### 5. Package Extraction Function

**Location**: `cmd/mcp-server/capabilities.go`

```go
import (
	"archive/tar"
	"compress/gzip"
)

// extractPackage extracts a tar.gz package to destination directory
func extractPackage(packagePath string, destDir string) error {
	// Open package file
	file, err := os.Open(packagePath)
	if err != nil {
		return fmt.Errorf("failed to open package: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read error: %w", err)
		}

		// Construct destination path
		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Create parent directory
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Create file
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// Copy content
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		}
	}

	return nil
}
```

**Test Cases**:
```go
func TestExtractPackage(t *testing.T) {
	// Create test tar.gz package
	tmpDir := t.TempDir()
	packagePath := filepath.Join(tmpDir, "test.tar.gz")

	// Create package with test content
	createTestPackage(t, packagePath, map[string]string{
		"commands/test.md": "# Test Capability",
		"agents/test.md":   "# Test Agent",
	})

	// Extract package
	extractDir := filepath.Join(tmpDir, "extract")
	err := extractPackage(packagePath, extractDir)
	if err != nil {
		t.Fatalf("extractPackage failed: %v", err)
	}

	// Verify extracted files
	testFile := filepath.Join(extractDir, "commands", "test.md")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("expected file not found: %s", testFile)
	}
}

// Helper: create test package
func createTestPackage(t *testing.T, path string, files map[string]string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for name, content := range files {
		// Write header
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("failed to write header: %v", err)
		}

		// Write content
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write content: %v", err)
		}
	}
}
```

#### 6. loadPackageCapabilities Function

**Location**: `cmd/mcp-server/capabilities.go`

```go
// loadPackageCapabilities loads capabilities from a package source
func loadPackageCapabilities(packageLocation string) ([]CapabilityMetadata, error) {
	// Get cache directory
	cacheDir, err := getPackageCacheDir(packageLocation)
	if err != nil {
		return nil, err
	}

	// Check if already cached and extracted
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		// Download and extract package
		if err := downloadAndExtractPackage(packageLocation, cacheDir); err != nil {
			return nil, err
		}
	}

	// Load capabilities from extracted package
	commandsDir := filepath.Join(cacheDir, "commands")
	caps, err := loadLocalCapabilities(commandsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load package capabilities: %w", err)
	}

	return caps, nil
}

// downloadAndExtractPackage downloads (if URL) and extracts a package
func downloadAndExtractPackage(packageLocation string, cacheDir string) error {
	var packagePath string

	// Check if URL or local path
	if strings.HasPrefix(packageLocation, "http://") || strings.HasPrefix(packageLocation, "https://") {
		// Download package
		packagePath = filepath.Join(cacheDir, "package.tar.gz")
		if err := os.MkdirAll(filepath.Dir(packagePath), 0755); err != nil {
			return err
		}
		if err := downloadPackage(packageLocation, packagePath); err != nil {
			return err
		}
	} else {
		// Local path, expand tilde
		packagePath = expandTilde(packageLocation)
		if _, err := os.Stat(packagePath); os.IsNotExist(err) {
			return fmt.Errorf("package file not found: %s", packagePath)
		}
	}

	// Extract package
	if err := extractPackage(packagePath, cacheDir); err != nil {
		return err
	}

	return nil
}
```

**Test Cases**:
```go
func TestLoadPackageCapabilities(t *testing.T) {
	// Create test package
	tmpDir := t.TempDir()
	packagePath := filepath.Join(tmpDir, "test.tar.gz")
	createTestPackage(t, packagePath, map[string]string{
		"commands/meta-test.md": `---
name: meta-test
description: Test capability.
keywords: test
category: test
---
# Test`,
	})

	// Load capabilities from package
	caps, err := loadPackageCapabilities(packagePath)
	if err != nil {
		t.Fatalf("loadPackageCapabilities failed: %v", err)
	}

	// Verify loaded capabilities
	if len(caps) != 1 {
		t.Errorf("expected 1 capability, got %d", len(caps))
	}
	if caps[0].Name != "meta-test" {
		t.Errorf("expected name 'meta-test', got %q", caps[0].Name)
	}
}
```

#### 7. Integrate Package Source into mergeSources

**Location**: `cmd/mcp-server/capabilities.go`

```go
// mergeSources loads capabilities from multiple sources and merges them
func mergeSources(sources []CapabilitySource) (CapabilityIndex, error) {
	index := make(CapabilityIndex)

	for _, source := range sources {
		var caps []CapabilityMetadata
		var err error

		switch source.Type {
		case SourceTypeLocal:
			caps, err = loadLocalCapabilities(source.Location)
		case SourceTypeGitHub:
			caps, err = loadGitHubCapabilities(source.Location)
		case SourceTypePackage:
			caps, err = loadPackageCapabilities(source.Location)
		default:
			err = fmt.Errorf("unknown source type: %s", source.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to load source %s: %w", source.Location, err)
		}

		// Merge capabilities (lower priority = earlier in list, gets overridden)
		for _, cap := range caps {
			cap.Source = source.Location
			cap.FilePath = cap.Name + ".md"

			// Only add if not already present (priority check)
			if _, exists := index[cap.Name]; !exists {
				index[cap.Name] = cap
			}
		}
	}

	return index, nil
}
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/capabilities.go` (+200 lines)
  - Add SourceTypePackage constant (~2 lines)
  - Update detectSourceType function (~10 lines)
  - Add getPackageCacheDir function (~15 lines)
  - Add downloadPackage function (~25 lines)
  - Add extractPackage function (~60 lines)
  - Add downloadAndExtractPackage function (~30 lines)
  - Add loadPackageCapabilities function (~25 lines)
  - Update mergeSources function (~10 lines)
  - Add imports and comments (~23 lines)

- `cmd/mcp-server/capabilities_test.go` (+150 lines)
  - Add TestDetectSourceType (~30 lines)
  - Add TestGetPackageCacheDir (~20 lines)
  - Add TestDownloadPackage (~25 lines)
  - Add TestExtractPackage (~30 lines)
  - Add TestLoadPackageCapabilities (~25 lines)
  - Add createTestPackage helper (~20 lines)

**Total**: ~350 lines (exceeds 250-line target due to comprehensive tar.gz handling)

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test package URL detection
3. Test package download (mock HTTP)
4. Test package extraction
5. Test loading capabilities from package
6. Test local package paths
7. Test mixed sources (package + directory)
8. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 22.8 (capability package build)

### Estimated Time

4 hours (350 lines implementation + comprehensive tests)

---

## Stage 22.10: Package Cache Metadata & TTL Strategy

### Objective

Implement cache metadata tracking for packages, including TTL validation and cleanup mechanisms.

### Acceptance Criteria

- [ ] Cache metadata stored in `.meta-cc-cache.json`
- [ ] Metadata tracks: package URL, download time, extracted path, TTL
- [ ] TTL validation: 7 days for release packages, 1 hour for branch packages
- [ ] Cache cleanup mechanism removes expired entries
- [ ] Stale cache detection and handling
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `cmd/mcp-server/capabilities_test.go` (additions ~50 lines)

**Test Strategy**:
1. Test metadata creation and persistence
2. Test TTL calculation (releases vs branches)
3. Test cache expiration detection
4. Test cache cleanup
5. Test stale cache usage

**Implementation Details**:

#### 1. Cache Metadata Structure

**Location**: `cmd/mcp-server/capabilities.go`

```go
// PackageCacheMetadata tracks cached package information
type PackageCacheMetadata struct {
	PackageURL    string    `json:"package_url"`
	DownloadTime  time.Time `json:"download_time"`
	ExtractedPath string    `json:"extracted_path"`
	TTL           int64     `json:"ttl_seconds"` // TTL in seconds
	IsRelease     bool      `json:"is_release"`  // Release vs branch
}

// CacheMetadataFile represents the cache metadata file structure
type CacheMetadataFile struct {
	Version  string                          `json:"version"`
	Packages map[string]PackageCacheMetadata `json:"packages"` // Key: cache hash
}
```

#### 2. Cache Metadata Persistence

**Location**: `cmd/mcp-server/capabilities.go`

```go
// loadCacheMetadata loads cache metadata from disk
func loadCacheMetadata() (*CacheMetadataFile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	metadataPath := filepath.Join(homeDir, CapabilityCacheDir, ".meta-cc-cache.json")

	// Check if file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		// Return empty metadata
		return &CacheMetadataFile{
			Version:  "1.0",
			Packages: make(map[string]PackageCacheMetadata),
		}, nil
	}

	// Read metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata CacheMetadataFile
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// saveCacheMetadata saves cache metadata to disk
func saveCacheMetadata(metadata *CacheMetadataFile) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(homeDir, CapabilityCacheDir)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	metadataPath := filepath.Join(cacheDir, ".meta-cc-cache.json")

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}
```

#### 3. TTL Calculation

**Location**: `cmd/mcp-server/capabilities.go`

```go
// getPackageTTL returns TTL for a package based on URL type
func getPackageTTL(packageURL string) time.Duration {
	// Release packages: 7 days (immutable)
	if strings.Contains(packageURL, "/releases/") {
		return 7 * 24 * time.Hour
	}

	// Branch packages: 1 hour (mutable)
	return 1 * time.Hour
}

// isReleasePackage checks if package URL points to a release
func isReleasePackage(packageURL string) bool {
	return strings.Contains(packageURL, "/releases/")
}
```

**Test Cases**:
```go
func TestGetPackageTTL(t *testing.T) {
	tests := []struct {
		packageURL string
		wantTTL    time.Duration
	}{
		{
			"https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz",
			7 * 24 * time.Hour,
		},
		{
			"https://github.com/yaleh/meta-cc/releases/download/v1.0.0/capabilities-v1.0.0.tar.gz",
			7 * 24 * time.Hour,
		},
		{
			"https://example.com/caps.tar.gz",
			1 * time.Hour,
		},
	}

	for _, tt := range tests {
		result := getPackageTTL(tt.packageURL)
		if result != tt.wantTTL {
			t.Errorf("getPackageTTL(%q) = %v, want %v",
				tt.packageURL, result, tt.wantTTL)
		}
	}
}
```

#### 4. Cache Validation

**Location**: `cmd/mcp-server/capabilities.go`

```go
// isCacheValid checks if cached package is still valid
func isCacheValid(packageURL string) bool {
	metadata, err := loadCacheMetadata()
	if err != nil {
		return false
	}

	// Get cache hash
	hash := sha256.Sum256([]byte(packageURL))
	hashStr := hex.EncodeToString(hash[:])[:16]

	pkg, exists := metadata.Packages[hashStr]
	if !exists {
		return false
	}

	// Check TTL
	age := time.Since(pkg.DownloadTime)
	ttl := time.Duration(pkg.TTL) * time.Second

	return age < ttl
}

// isCacheStale checks if cache is expired but still usable (< 7 days old)
func isCacheStale(packageURL string) bool {
	metadata, err := loadCacheMetadata()
	if err != nil {
		return false
	}

	hash := sha256.Sum256([]byte(packageURL))
	hashStr := hex.EncodeToString(hash[:])[:16]

	pkg, exists := metadata.Packages[hashStr]
	if !exists {
		return false
	}

	age := time.Since(pkg.DownloadTime)
	ttl := time.Duration(pkg.TTL) * time.Second
	maxStaleAge := 7 * 24 * time.Hour

	return age >= ttl && age < maxStaleAge
}
```

#### 5. Update loadPackageCapabilities with Metadata

**Location**: `cmd/mcp-server/capabilities.go`

```go
// loadPackageCapabilities loads capabilities from a package source
func loadPackageCapabilities(packageLocation string) ([]CapabilityMetadata, error) {
	// Get cache directory
	cacheDir, err := getPackageCacheDir(packageLocation)
	if err != nil {
		return nil, err
	}

	// Check if cache is valid
	if !isCacheValid(packageLocation) {
		// Cache expired or doesn't exist, download and extract
		if err := downloadAndExtractPackage(packageLocation, cacheDir); err != nil {
			// If network error, try using stale cache
			if isCacheStale(packageLocation) {
				fmt.Fprintf(os.Stderr, "Warning: Using stale cached package (network error)\n")
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			} else {
				return nil, err
			}
		} else {
			// Update cache metadata
			if err := updateCacheMetadata(packageLocation, cacheDir); err != nil {
				// Non-fatal, log warning
				fmt.Fprintf(os.Stderr, "Warning: Failed to update cache metadata: %v\n", err)
			}
		}
	}

	// Load capabilities from extracted package
	commandsDir := filepath.Join(cacheDir, "commands")
	caps, err := loadLocalCapabilities(commandsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load package capabilities: %w", err)
	}

	return caps, nil
}

// updateCacheMetadata updates cache metadata after downloading a package
func updateCacheMetadata(packageURL string, cacheDir string) error {
	metadata, err := loadCacheMetadata()
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(packageURL))
	hashStr := hex.EncodeToString(hash[:])[:16]

	ttl := getPackageTTL(packageURL)

	metadata.Packages[hashStr] = PackageCacheMetadata{
		PackageURL:    packageURL,
		DownloadTime:  time.Now(),
		ExtractedPath: cacheDir,
		TTL:           int64(ttl.Seconds()),
		IsRelease:     isReleasePackage(packageURL),
	}

	return saveCacheMetadata(metadata)
}
```

#### 6. Cache Cleanup Function

**Location**: `cmd/mcp-server/capabilities.go`

```go
// cleanupExpiredCache removes expired cache entries
func cleanupExpiredCache() error {
	metadata, err := loadCacheMetadata()
	if err != nil {
		return err
	}

	maxStaleAge := 7 * 24 * time.Hour
	modified := false

	for hash, pkg := range metadata.Packages {
		age := time.Since(pkg.DownloadTime)

		// Remove if older than max stale age
		if age >= maxStaleAge {
			// Remove cache directory
			if err := os.RemoveAll(pkg.ExtractedPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to remove cache directory: %v\n", err)
			}

			// Remove from metadata
			delete(metadata.Packages, hash)
			modified = true
		}
	}

	if modified {
		return saveCacheMetadata(metadata)
	}

	return nil
}
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/capabilities.go` (+100 lines)
  - Add PackageCacheMetadata struct (~8 lines)
  - Add CacheMetadataFile struct (~5 lines)
  - Add loadCacheMetadata function (~25 lines)
  - Add saveCacheMetadata function (~15 lines)
  - Add getPackageTTL function (~8 lines)
  - Add isReleasePackage function (~3 lines)
  - Add isCacheValid function (~20 lines)
  - Add isCacheStale function (~20 lines)
  - Add updateCacheMetadata function (~20 lines)
  - Add cleanupExpiredCache function (~25 lines)
  - Update loadPackageCapabilities (~15 lines)

- `cmd/mcp-server/capabilities_test.go` (+50 lines)
  - Add TestGetPackageTTL (~20 lines)
  - Add TestCacheMetadataPersistence (~20 lines)
  - Add TestCacheValidation (~10 lines)

**Total**: ~150 lines (exceeds 100-line target due to comprehensive metadata handling)

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test cache metadata creation and persistence
3. Test TTL calculation (releases vs branches)
4. Test cache expiration detection
5. Test stale cache usage
6. Test cache cleanup
7. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 22.9 (package source support)

### Estimated Time

2 hours (150 lines implementation + tests)

---

## Stage 22.11: Default Source to Package File & Fallback Strategy

### Objective

Change default capability source to GitHub Release package file and implement intelligent fallback strategy.

### Acceptance Criteria

- [ ] Default source changed to GitHub Release package URL
- [ ] Fallback strategy: package → GitHub raw → local cache
- [ ] `latest` redirect handling works
- [ ] Local development requires explicit configuration
- [ ] Documentation updated
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `cmd/mcp-server/capabilities_test.go` (additions ~30 lines)

**Test Strategy**:
1. Test default source is package URL
2. Test fallback to GitHub raw on package failure
3. Test fallback to local cache on all failures
4. Test `latest` redirect handling

**Implementation Details**:

#### 1. Update Default Source Constant

**Location**: `cmd/mcp-server/capabilities.go`

```go
const (
	// DefaultCapabilitySource is the default source when no env var is set
	// Uses GitHub Release package for zero-configuration deployment
	DefaultCapabilitySource = "https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"

	// FallbackGitHubSource is the fallback source when package fails
	FallbackGitHubSource = "yaleh/meta-cc@main/commands"

	// LocalCapabilitySource defines the local capability source for development
	LocalCapabilitySource = "capabilities/commands"

	// CapabilityCacheDir defines the directory for caching downloaded capabilities
	CapabilityCacheDir = ".capabilities-cache"
)
```

#### 2. Implement Fallback Strategy

**Location**: `cmd/mcp-server/capabilities.go`

```go
// getCapabilityIndexWithFallback loads capability index with fallback strategy
func getCapabilityIndexWithFallback(sources []CapabilitySource, disableCache bool) (CapabilityIndex, error) {
	// Try primary sources first
	index, err := mergeSources(sources)
	if err == nil {
		return index, nil
	}

	// If all sources failed, try fallback strategy
	primaryErr := err

	// Check if primary source was a package
	hasPackageSource := false
	for _, source := range sources {
		if source.Type == SourceTypePackage {
			hasPackageSource = true
			break
		}
	}

	if hasPackageSource {
		fmt.Fprintf(os.Stderr, "Warning: Package source failed, trying fallback to GitHub...\n")
		fmt.Fprintf(os.Stderr, "Primary error: %v\n", primaryErr)

		// Fallback to GitHub raw
		fallbackSources := []CapabilitySource{
			{Type: SourceTypeGitHub, Location: FallbackGitHubSource, Priority: 0},
		}

		index, err := mergeSources(fallbackSources)
		if err == nil {
			return index, nil
		}

		fmt.Fprintf(os.Stderr, "Warning: GitHub fallback failed: %v\n", err)

		// Try stale cache as last resort
		// (Note: This requires package source to have been cached previously)
		for _, source := range sources {
			if source.Type == SourceTypePackage && isCacheStale(source.Location) {
				fmt.Fprintf(os.Stderr, "Using stale cached package as last resort\n")

				// Force load from cache (ignore TTL)
				index, err := forceLoadFromCache(source.Location)
				if err == nil {
					return index, nil
				}
			}
		}
	}

	// All strategies failed, return original error
	return nil, primaryErr
}

// forceLoadFromCache loads capabilities from cache ignoring TTL
func forceLoadFromCache(packageLocation string) (CapabilityIndex, error) {
	cacheDir, err := getPackageCacheDir(packageLocation)
	if err != nil {
		return nil, err
	}

	// Check if cache exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no cache available for package")
	}

	// Load capabilities from cache
	commandsDir := filepath.Join(cacheDir, "commands")
	caps, err := loadLocalCapabilities(commandsDir)
	if err != nil {
		return nil, err
	}

	// Convert to index
	index := make(CapabilityIndex)
	for _, cap := range caps {
		cap.Source = packageLocation
		index[cap.Name] = cap
	}

	return index, nil
}
```

#### 3. Update Tool Executors with Fallback

**Location**: `cmd/mcp-server/capabilities.go`

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
		// Default to GitHub Release package if no sources configured
		sources = []CapabilitySource{
			{Type: SourceTypePackage, Location: DefaultCapabilitySource, Priority: 0},
		}
	}

	// Check cache disable flag
	disableCache := false
	if disable, ok := args["_disable_cache"].(bool); ok {
		disableCache = disable
	}

	// Get capability index with fallback
	index, err := getCapabilityIndexWithFallback(sources, disableCache)
	if err != nil {
		return "", err
	}

	// Format output as JSON
	result := map[string]interface{}{
		"capabilities": index,
		"sources":      formatSources(sources),
		"count":        len(index),
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(output), nil
}
```

#### 4. Handle `latest` Redirects

**Location**: `cmd/mcp-server/capabilities.go`

```go
// downloadPackageWithRedirect downloads a package handling redirects
func downloadPackageWithRedirect(url string, destPath string) error {
	// Create HTTP client with redirect following
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Update downloadPackage to use redirect-aware version
func downloadPackage(url string, destPath string) error {
	return downloadPackageWithRedirect(url, destPath)
}
```

**Test Cases**:
```go
func TestDefaultSourceIsPackage(t *testing.T) {
	// Verify default source is package URL
	if !strings.HasSuffix(DefaultCapabilitySource, ".tar.gz") {
		t.Errorf("DefaultCapabilitySource should be a package URL, got: %s", DefaultCapabilitySource)
	}
	if !strings.Contains(DefaultCapabilitySource, "/releases/") {
		t.Errorf("DefaultCapabilitySource should point to a release, got: %s", DefaultCapabilitySource)
	}
}

func TestFallbackStrategy(t *testing.T) {
	// Test fallback to GitHub when package fails
	// (This requires mocking network calls, simplified version shown)
	sources := []CapabilitySource{
		{Type: SourceTypePackage, Location: "https://example.com/nonexistent.tar.gz", Priority: 0},
	}

	// This should fail but try fallback
	_, err := getCapabilityIndexWithFallback(sources, false)

	// Error expected (no real GitHub access in test)
	if err == nil {
		t.Errorf("Expected error when all sources fail")
	}
}
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/capabilities.go` (+80 lines)
  - Update DefaultCapabilitySource constant (~3 lines)
  - Add FallbackGitHubSource constant (~2 lines)
  - Add getCapabilityIndexWithFallback function (~40 lines)
  - Add forceLoadFromCache function (~20 lines)
  - Add downloadPackageWithRedirect function (~20 lines)
  - Update executeListCapabilitiesTool (~5 lines)
  - Update executeGetCapabilityTool (~5 lines)

- `cmd/mcp-server/capabilities_test.go` (+30 lines)
  - Add TestDefaultSourceIsPackage (~10 lines)
  - Add TestFallbackStrategy (~20 lines)

**Total**: ~110 lines (exceeds 80-line target due to comprehensive fallback logic)

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test default source is package URL
3. Test fallback to GitHub on package failure
4. Test stale cache usage as last resort
5. Test `latest` redirect handling
6. **Manual**: Test with real GitHub Release package
7. **HALT if tests fail after 2 fix attempts**

### Dependencies

- Stage 22.10 (cache metadata)

### Estimated Time

2 hours (110 lines implementation + tests)

---

## Stage 22.12: Documentation Updates

### Objective

Update all documentation to reflect package file distribution strategy, cache behavior, and fallback mechanisms.

### Acceptance Criteria

- [ ] CLAUDE.md updated with package file approach
- [ ] README.md emphasizes zero-configuration
- [ ] docs/capabilities-guide.md updated with packaging workflow
- [ ] CHANGELOG.md includes breaking changes
- [ ] All examples use package file syntax
- [ ] Documentation accurate and comprehensive

### Implementation

#### 1. CLAUDE.md Updates (+40 lines)

**Location**: Section "Unified Meta Command" → "Multi-Source Configuration"

```markdown
### Multi-Source Configuration

Configure capability sources via environment variable:

```bash
# Single local source
export META_CC_CAPABILITY_SOURCES="~/.config/meta-cc/capabilities"

# Package file (GitHub Release)
export META_CC_CAPABILITY_SOURCES="https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"

# Local package file
export META_CC_CAPABILITY_SOURCES="/path/to/capabilities.tar.gz"

# Multiple sources (priority: left-to-right, left = highest)
export META_CC_CAPABILITY_SOURCES="~/dev/my-caps:/local/caps.tar.gz:https://example.com/caps.tar.gz"

# Mix package, directory, and GitHub sources
export META_CC_CAPABILITY_SOURCES="~/dev/test:./capabilities.tar.gz:yaleh/meta-cc@main/commands"
```

**Source Types**:
- **Local directories**: Immediate reflection, no cache (for development)
- **Package files** (`.tar.gz`): Cached with TTL (7 days for releases, 1 hour for branches)
- **GitHub repositories**: GitHub API access (format: `owner/repo@branch` or `owner/repo`)

**Package File Distribution**:

Capabilities are distributed as prebuilt `.tar.gz` packages for:
- **Offline-friendly**: Download once, cache locally
- **Reliable**: No CDN dependencies, no rate limits
- **Fast**: No network calls after initial download

Cache directory: `~/.capabilities-cache/packages/<hash>/`

**Cache Strategy**:

- **Release packages**: 7-day cache (immutable releases from `/releases/`)
- **Branch packages**: 1-hour cache (mutable, custom builds)
- **Local sources**: No cache (always fresh)

**Fallback Strategy**:

When package loading fails, meta-cc automatically tries:
1. **Primary**: Package file (if configured)
2. **Fallback 1**: GitHub raw access (`yaleh/meta-cc@main/commands`)
3. **Fallback 2**: Stale cache (up to 7 days old)

**Default Source**:

If `META_CC_CAPABILITY_SOURCES` is not set, capabilities are loaded from:
```
https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz
```

For local development, explicitly set the environment variable:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```
```

#### 2. README.md Updates (+30 lines)

**Location**: "Unified Meta Command" section

```markdown
## Unified Meta Command

Use natural language to invoke meta-cognition capabilities:

```bash
/meta "show errors"           # Error analysis
/meta "quality check"         # Code quality scan
/meta "visualize timeline"    # Project timeline
```

### Zero-Configuration Setup

meta-cc works out of the box with no configuration required. Capabilities are automatically loaded from the latest GitHub Release package:

```
https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz
```

This provides:
- **Latest capabilities**: Always up-to-date with the latest release
- **Offline-friendly**: Download once, cache for 7 days
- **Reliable**: No CDN dependencies, no rate limits
- **Fast**: Cached package loading (~10ms)

### Local Development

For local capability development, override the default source:

```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

This enables:
- **Real-time reflection**: Changes appear immediately (no cache)
- **Testing before commit**: Verify changes locally
- **Offline work**: No network dependency

### Custom Package Files

Distribute custom capabilities as package files:

```bash
# Build package
cd my-capabilities
tar -czf ../my-caps.tar.gz commands/ agents/

# Use package
export META_CC_CAPABILITY_SOURCES="/path/to/my-caps.tar.gz"
```

### Multi-Source Capabilities

Load capabilities from multiple sources:

```bash
export META_CC_CAPABILITY_SOURCES="~/my-caps:/local/caps.tar.gz:https://example.com/caps.tar.gz"
```

Supports:
- **Local directories**: Immediate reflection, no cache
- **Package files**: Cached with TTL (7d releases, 1h custom)
- **GitHub repositories**: GitHub API access
- **Priority-based merging**: Left = highest priority (overrides duplicates)

See [docs/capabilities-guide.md](docs/capabilities-guide.md) for development guide.
```

#### 3. docs/capabilities-guide.md Updates (+30 lines)

**Location**: Add new section "Building Capability Packages"

```markdown
## Building Capability Packages

### Creating a Package

Package your capabilities for distribution:

```bash
# 1. Organize capabilities
my-capabilities/
├── commands/
│   ├── capability1.md
│   └── capability2.md
└── agents/
    └── agent1.md

# 2. Build package
cd my-capabilities
tar -czf ../my-capabilities.tar.gz commands/ agents/

# Or use the Makefile target (for meta-cc itself)
make bundle-capabilities
# Creates: build/capabilities-latest.tar.gz
```

### Package Structure

Valid package structure:

```
capabilities.tar.gz
├── commands/          # Required
│   ├── *.md          # Capability files
│   └── ...
└── agents/           # Optional
    ├── *.md          # Agent files
    └── ...
```

### Testing Packages

Test your package before distribution:

```bash
# 1. Use local package
export META_CC_CAPABILITY_SOURCES="/path/to/my-capabilities.tar.gz"

# 2. List capabilities
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# 3. Test capability
/meta "my capability"
```

### Distributing Packages

**Option 1: GitHub Releases**

1. Create GitHub Release with tag (e.g., `v1.0.0`)
2. Upload package as release asset: `capabilities-v1.0.0.tar.gz`
3. Users install via:
   ```bash
   export META_CC_CAPABILITY_SOURCES="https://github.com/username/repo/releases/download/v1.0.0/capabilities-v1.0.0.tar.gz"
   ```

**Option 2: Web Server**

1. Upload package to web server: `https://example.com/capabilities.tar.gz`
2. Users install via:
   ```bash
   export META_CC_CAPABILITY_SOURCES="https://example.com/capabilities.tar.gz"
   ```

**Option 3: Local Distribution**

1. Share package file directly
2. Users install via:
   ```bash
   export META_CC_CAPABILITY_SOURCES="/path/to/capabilities.tar.gz"
   ```

### Cache Behavior

- **Release packages** (`/releases/`): 7-day cache (immutable)
- **Custom packages**: 1-hour cache (may change)
- **Local packages**: Re-extracted on each use

To force refresh:
```bash
# Clear cache
rm -rf ~/.capabilities-cache/packages/

# Or use _disable_cache parameter (MCP)
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{"_disable_cache":true}}}' | meta-cc-mcp
```

## Publishing Capabilities

### Method 1: GitHub Release Package (Recommended)

1. Create GitHub repo: `username/meta-cc-capabilities`
2. Add capabilities: `capabilities/commands/my-feature.md`
3. Build package:
   ```bash
   tar -czf capabilities-latest.tar.gz -C capabilities commands agents
   ```
4. Create GitHub Release with tag (e.g., `v1.0.0`)
5. Upload `capabilities-latest.tar.gz` as release asset
6. Users install via:
   ```bash
   # Latest release
   export META_CC_CAPABILITY_SOURCES="https://github.com/username/meta-cc-capabilities/releases/latest/download/capabilities-latest.tar.gz"

   # Specific version
   export META_CC_CAPABILITY_SOURCES="https://github.com/username/meta-cc-capabilities/releases/download/v1.0.0/capabilities-v1.0.0.tar.gz"
   ```

**Benefits**:
- **7-day cache**: Faster loading, reduced network calls
- **Version pinning**: Users can opt into specific versions
- **Immutable**: Release packages don't change
- **Automatic updates**: Users using `latest` get new versions automatically

### Method 2: GitHub Repository (Development)

1. Fork `yaleh/meta-cc`
2. Add capability: `capabilities/commands/meta-my-feature.md`
3. Submit PR
4. After merge, available to all users via package or GitHub source

### Method 3: Fork and PR

1. Fork `yaleh/meta-cc`
2. Add capability: `capabilities/commands/meta-my-feature.md`
3. Submit PR
4. After merge, included in next release package
```

#### 4. CHANGELOG.md Updates (+20 lines)

**Location**: Create new unreleased section at top

```markdown
## [Unreleased]

### Added (Phase 22.8-22.12)

- **Package File Distribution**: Capabilities distributed as prebuilt `.tar.gz` packages
  - Build target: `make bundle-capabilities` creates `capabilities-latest.tar.gz`
  - GitHub Actions automatically uploads packages to releases
  - Package format: `commands/` and `agents/` directories
  - Cache directory: `~/.capabilities-cache/packages/<hash>/`

- **Package Source Support**:
  - MCP recognizes `.tar.gz` URLs and local paths
  - Automatic download and extraction
  - Smart caching with metadata tracking
  - Format: `export META_CC_CAPABILITY_SOURCES="/path/to/caps.tar.gz"`

- **Cache Strategy**:
  - Release packages: 7-day cache (immutable)
  - Custom packages: 1-hour cache (mutable)
  - Cache metadata: `.meta-cc-cache.json` tracks TTL and versions
  - Automatic cleanup of expired cache (>7 days)

- **Intelligent Fallback**:
  - Primary: Package file (if configured)
  - Fallback 1: GitHub raw access (`yaleh/meta-cc@main/commands`)
  - Fallback 2: Stale cache (up to 7 days old)
  - Clear error messages guide users to solutions

### Changed (Phase 22.8-22.12)

- **Default Source Changed**: Capabilities now load from GitHub Release package by default
  - Zero-configuration deployment
  - Local development requires: `export META_CC_CAPABILITY_SOURCES="capabilities/commands"`

### Breaking Changes

⚠️ **Default Source Changed**: If you were relying on the default GitHub raw source, you must now explicitly set:

```bash
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
```

Or use the new package default:
```bash
# Uses package file (default, no configuration needed)
unset META_CC_CAPABILITY_SOURCES
```

### Migration Guide

**For Local Development**:

Old behavior (implicit):
```bash
# Capabilities loaded from GitHub raw automatically
```

New behavior (explicit):
```bash
# Must explicitly set local source
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

**For Production Deployment**:

Old behavior:
```bash
# Required environment variable for GitHub source
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
```

New behavior:
```bash
# No configuration needed (uses release package)
# OR explicitly set for custom source
export META_CC_CAPABILITY_SOURCES="https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"
```

**For Custom Distributions**:

```bash
# Build package
make bundle-capabilities

# Distribute package
# Option 1: GitHub Release (recommended)
gh release create v1.0.0 build/capabilities-latest.tar.gz

# Option 2: Web server
cp build/capabilities-latest.tar.gz /var/www/html/

# Option 3: Local file
export META_CC_CAPABILITY_SOURCES="/path/to/capabilities-latest.tar.gz"
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
grep -r "capabilities-latest.tar.gz" CLAUDE.md README.md docs/
grep -r "Package File Distribution" CLAUDE.md README.md docs/
grep -r "bundle-capabilities" CLAUDE.md README.md docs/

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

- All previous stages (22.8-22.11)

### Estimated Time

2 hours (120 lines documentation + verification)

---

## Phase Integration Strategy

### Build Verification

After completing all stages 22.8-22.12:

```bash
# 1. Full build
make all

# 2. Build capability package
make bundle-capabilities

# 3. Verify package structure
tar -tzf build/capabilities-latest.tar.gz | head -20

# 4. Run integration test
bash tests/integration/test-capability-package.sh

# 5. Unit tests
go test -v ./cmd/mcp-server -run TestDetectSourceType
go test -v ./cmd/mcp-server -run TestGetPackageCacheDir
go test -v ./cmd/mcp-server -run TestDownloadPackage
go test -v ./cmd/mcp-server -run TestExtractPackage
go test -v ./cmd/mcp-server -run TestLoadPackageCapabilities
go test -v ./cmd/mcp-server -run TestGetPackageTTL
go test -v ./cmd/mcp-server -run TestCacheMetadata
go test -v ./cmd/mcp-server -run TestDefaultSourceIsPackage
go test -v ./cmd/mcp-server -run TestFallbackStrategy

# 6. Test coverage
make test-coverage
# Should maintain ≥80% coverage

# 7. Integration test with local package
export META_CC_CAPABILITY_SOURCES="build/capabilities-latest.tar.gz"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# 8. Test with URL (if release exists)
export META_CC_CAPABILITY_SOURCES="https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# 9. Test default source (no env var)
unset META_CC_CAPABILITY_SOURCES
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp

# 10. Test fallback strategy
export META_CC_CAPABILITY_SOURCES="https://nonexistent.example.com/caps.tar.gz"
echo '{"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | meta-cc-mcp
# Should fallback to GitHub, then stale cache
```

### Rollout Checklist

Before marking Stages 22.8-22.12 complete:

- [ ] All 5 stages completed and tested
- [ ] `make all` passes without errors
- [ ] Test coverage ≥80% maintained
- [ ] `make bundle-capabilities` works correctly
- [ ] Package structure valid (commands/, agents/)
- [ ] GitHub Actions workflow syntax valid
- [ ] Package URL detection works
- [ ] Package download works (HTTP/HTTPS)
- [ ] Package extraction works
- [ ] Cache metadata persists correctly
- [ ] TTL calculation correct (7d releases, 1h custom)
- [ ] Cache expiration detection works
- [ ] Default source changed to package URL
- [ ] Fallback strategy works (package → GitHub → stale cache)
- [ ] Local development workflow documented
- [ ] Documentation updated and accurate
- [ ] CHANGELOG.md includes breaking changes and migration guide
- [ ] Manual package testing successful (local and URL)
- [ ] Backward compatibility verified

---

## File Change Inventory

### Summary by Stage

| Stage  | Modified Files | Total Lines | Description |
|--------|----------------|-------------|-------------|
| 22.8   | 3              | ~100        | Build + CI + integration test |
| 22.9   | 2              | ~350        | Package source support + tests |
| 22.10  | 2              | ~150        | Cache metadata + TTL + tests |
| 22.11  | 2              | ~110        | Default source + fallback + tests |
| 22.12  | 4              | ~120        | Documentation updates |
| **Total** | **13** | **~830** | **Complete package distribution** |

**Note**: Total ~830 lines exceeds initial 750-line estimate due to:
- Comprehensive tar.gz handling (extraction, validation)
- Robust cache metadata system
- Intelligent fallback strategy with stale cache support
- Extensive test coverage (≥80% for all new code)
- Detailed documentation (migration guide, examples)

### Detailed File Changes

**Stage 22.8**:
- `Makefile` (+30 lines)
- `.github/workflows/release.yml` (+20 lines)
- `tests/integration/test-capability-package.sh` (+50 lines, new file)

**Stage 22.9**:
- `cmd/mcp-server/capabilities.go` (+200 lines)
- `cmd/mcp-server/capabilities_test.go` (+150 lines)

**Stage 22.10**:
- `cmd/mcp-server/capabilities.go` (+100 lines)
- `cmd/mcp-server/capabilities_test.go` (+50 lines)

**Stage 22.11**:
- `cmd/mcp-server/capabilities.go` (+80 lines)
- `cmd/mcp-server/capabilities_test.go` (+30 lines)

**Stage 22.12**:
- `CLAUDE.md` (+40 lines)
- `README.md` (+30 lines)
- `docs/capabilities-guide.md` (+30 lines)
- `CHANGELOG.md` (+20 lines)

---

## Risk Assessment and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| tar.gz extraction bugs | Medium | High | Comprehensive test coverage, test with various package structures |
| Package file corruption | Low | High | Checksum validation (future enhancement), retry download |
| Cache directory conflicts | Low | Medium | Use hash-based cache directories, cleanup expired entries |
| Breaking change impact | High | Medium | Clear migration guide, CHANGELOG warnings |
| Network download failures | Medium | High | Fallback strategy (GitHub → stale cache), clear error messages |
| GitHub Release availability | Low | High | Fallback to GitHub raw, stale cache support |
| Documentation accuracy | Medium | Low | Manual verification, example testing |

### Contingency Plans

**If tar.gz extraction fails**:
- Comprehensive error messages show extraction failure reason
- Fallback to GitHub raw access automatically
- Use stale cache if available
- Test with various package structures before release

**If package file corrupted**:
- Re-download on corruption detection
- Future: Add checksum validation (SHA256)
- Fallback to GitHub raw access

**If cache directory issues**:
- Hash-based directories prevent conflicts
- Automatic cleanup of expired entries
- Manual cleanup command (future enhancement)

**If breaking change causes problems**:
- Migration guide in CHANGELOG
- Warning messages in error output
- Backward compatibility for explicit sources

**If network failures persist**:
- Stale cache support (up to 7 days)
- Clear error messages guide to local sources
- Fallback to GitHub raw access

**If GitHub Release unavailable**:
- Automatic fallback to GitHub raw
- Stale cache as last resort
- Error messages suggest alternatives

---

## Testing Strategy

### Unit Testing

**Coverage Requirements**:
- Each stage: ≥80% coverage
- Critical paths: 100% coverage (extraction, cache, fallback)
- Edge cases: Comprehensive test cases

**Test Organization**:
```
cmd/mcp-server/
  capabilities.go              - Implementation
  capabilities_test.go         - Unit tests
tests/integration/
  test-capability-package.sh   - Integration test script
```

**New Test Functions** (22.8-22.11):
- TestDetectSourceType
- TestGetPackageCacheDir
- TestDownloadPackage
- TestExtractPackage
- TestLoadPackageCapabilities
- TestGetPackageTTL
- TestCacheMetadataPersistence
- TestCacheValidation
- TestDefaultSourceIsPackage
- TestFallbackStrategy

### Integration Testing

**Multi-Source Scenarios**:
- Single package source (URL)
- Single package source (local path)
- Mixed package + directory sources
- Default source (no env var)
- Priority override with duplicates

**End-to-End Workflows**:
```bash
# Workflow 1: Package Download and Extraction
1. Set package URL source
2. Call list_capabilities()
3. Verify package downloaded
4. Verify package extracted
5. Verify capabilities loaded

# Workflow 2: Cache TTL
1. Set release package source
2. Load capabilities (first call)
3. Verify cache created
4. Load capabilities (second call)
5. Verify cache used (no download)

# Workflow 3: Fallback Strategy
1. Set invalid package source
2. Call list_capabilities()
3. Verify fallback to GitHub
4. Verify warning message
5. Verify capabilities loaded

# Workflow 4: Stale Cache
1. Set package source
2. Load capabilities
3. Expire cache (mock time)
4. Simulate network failure
5. Verify stale cache used
6. Verify warning message
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
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
/meta "show errors"

# GitHub sources should still work
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
/meta "show errors"
```

### Performance Testing

**Benchmarks**:
```bash
# Measure package download time
time bash -c 'export META_CC_CAPABILITY_SOURCES="https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz" && echo "{\"method\":\"tools/call\",\"params\":{\"name\":\"list_capabilities\",\"arguments\":{}}}" | meta-cc-mcp'
# Expected: <3s first call, <100ms cached

# Measure extraction time
time tar -xzf build/capabilities-latest.tar.gz -C /tmp/test-extract
# Expected: <500ms for typical package

# Measure cache effectiveness
# First call: Download and extract
# Second call: Load from cache (should be much faster)
```

---

## Timeline Estimate

| Stage  | Description | Estimated Time |
|--------|-------------|----------------|
| 22.8   | Build tooling + CI | 3 hours |
| 22.9   | Package source support | 4 hours |
| 22.10  | Cache metadata + TTL | 2 hours |
| 22.11  | Default source + fallback | 2 hours |
| 22.12  | Documentation updates | 2 hours |
| **Total** | **All stages** | **13 hours** |

**Contingency**: +3 hours for testing, debugging, and integration (total: 16 hours)

---

## Conclusion

Stages 22.8-22.12 implement production-ready capability package distribution, providing:

1. **Offline-Friendly**: Download once, cache locally (7 days for releases)
2. **Reliable**: No CDN dependencies, no rate limits
3. **Fast**: Cached package loading, no network calls after download
4. **Simple**: Standard tar.gz format, easy to verify and debug
5. **Robust**: Intelligent fallback (package → GitHub → stale cache)
6. **Zero-Configuration**: Works out of the box with GitHub Release package

Key success factors:
- TDD methodology ensures high quality
- Comprehensive tar.gz handling with extensive tests
- Smart cache strategy with metadata tracking
- Intelligent fallback minimizes failures
- Clear migration guide reduces breaking change impact
- Package distribution scales better than CDN approach

Upon completion, meta-cc will have a production-ready, reliable, and user-friendly capability distribution system that works seamlessly offline and online, with minimal network dependencies.

---

## Next Steps (Post-Stage 22.12)

After completing Stages 22.8-22.12:

1. **Package Verification**:
   - Implement checksum validation (SHA256)
   - Add package signature verification
   - Verify package integrity before extraction

2. **Cache Management**:
   - Add manual cache clear command
   - Implement cache size limits
   - Add cache statistics/status command

3. **Distribution Enhancements**:
   - Support multiple package formats (zip, etc.)
   - Add package versioning metadata
   - Implement package delta updates

4. **Community Engagement**:
   - Publish official capability packages with releases
   - Create capability package registry
   - Document package creation best practices

5. **Advanced Features**:
   - Capability marketplace integration
   - Automated package testing pipeline
   - Multi-language capability support
