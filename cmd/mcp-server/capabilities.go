package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// SourceType represents the type of capability source
type SourceType string

const (
	// SourceTypeLocal represents a local filesystem directory
	SourceTypeLocal SourceType = "local"
	// SourceTypeGitHub represents a GitHub repository
	SourceTypeGitHub SourceType = "github"
	// SourceTypePackage represents a tar.gz package file (URL or local path)
	SourceTypePackage SourceType = "package"

	// DefaultCapabilitySource is the default source when no env var is set
	// Uses GitHub repository with main branch for zero-configuration deployment
	DefaultCapabilitySource = "yaleh/meta-cc@main/commands"

	// LocalCapabilitySource defines the local capability source for development
	LocalCapabilitySource = "capabilities/commands"

	// CapabilityCacheDir defines the directory for caching downloaded capabilities
	CapabilityCacheDir = ".capabilities-cache"
)

// CapabilitySource represents a source of capabilities
type CapabilitySource struct {
	Type     SourceType // "local" or "github"
	Location string     // "/path/to/dir" or "owner/repo/subdir"
	Priority int        // Left-to-right priority (0 = highest)
}

// CapabilityMetadata represents metadata extracted from capability frontmatter
type CapabilityMetadata struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Keywords    []string `yaml:"-"` // Parsed from comma-separated string
	Category    string   `yaml:"category"`
	Source      string   `json:"source"`    // Source identifier for debugging
	FilePath    string   `json:"file_path"` // Relative path within source

	// Internal field for parsing keywords from YAML
	KeywordsRaw string `yaml:"keywords"`
}

// CapabilityIndex maps capability names to their metadata
type CapabilityIndex map[string]CapabilityMetadata

// GitHubSource represents a parsed GitHub source with branch/tag
type GitHubSource struct {
	Owner  string // Repository owner
	Repo   string // Repository name
	Branch string // Branch, tag, or commit hash
	Subdir string // Optional subdirectory
}

// VersionType represents the type of version reference
type VersionType string

const (
	// VersionTypeBranch represents a branch reference (mutable)
	VersionTypeBranch VersionType = "branch"
	// VersionTypeTag represents a tag reference (immutable)
	VersionTypeTag VersionType = "tag"
)

// parseCapabilitySources parses the environment variable into a list of capability sources
func parseCapabilitySources(envVar string) []CapabilitySource {
	if envVar == "" {
		return []CapabilitySource{}
	}

	// Split by path separator, but handle URLs specially
	// URLs contain : which conflicts with path separator on Unix
	parts := smartSplitSources(envVar)
	sources := make([]CapabilitySource, 0, len(parts))

	priority := 0
	for _, location := range parts {
		location = strings.TrimSpace(location)
		if location == "" {
			continue
		}

		sources = append(sources, CapabilitySource{
			Type:     detectSourceType(location),
			Location: location,
			Priority: priority,
		})
		priority++
	}

	return sources
}

// smartSplitSources splits sources by path separator, but handles URLs correctly
func smartSplitSources(envVar string) []string {
	if envVar == "" {
		return []string{}
	}

	// On Unix, path separator is :, but URLs also contain :
	// We need to avoid splitting on : that's part of http:// or https://
	sep := string(os.PathListSeparator)
	if sep != ":" {
		// On Windows, path separator is ;, no conflict with URLs
		return strings.Split(envVar, sep)
	}

	// Unix: split on : but not in URLs
	var parts []string
	current := ""
	i := 0
	for i < len(envVar) {
		if envVar[i] == ':' {
			// Check if this is part of a URL scheme (http:// or https://)
			if i+2 < len(envVar) && envVar[i+1] == '/' && envVar[i+2] == '/' {
				// This is a URL scheme, include it in current part
				current += "://"
				i += 3
			} else {
				// This is a separator, save current part and start new one
				if current != "" {
					parts = append(parts, current)
				}
				current = ""
				i++
			}
		} else {
			current += string(envVar[i])
			i++
		}
	}
	// Don't forget the last part
	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// detectSourceType determines if a location is a local path, package file, or GitHub repository
func detectSourceType(location string) SourceType {
	// Check for URLs first (before checking for / in path)
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		// Check if URL points to package file
		if strings.HasSuffix(location, ".tar.gz") || strings.HasSuffix(location, ".tgz") {
			return SourceTypePackage
		}
		return SourceTypeGitHub // URLs default to GitHub
	}

	// Check for package file (tar.gz or tgz)
	if strings.HasSuffix(location, ".tar.gz") || strings.HasSuffix(location, ".tgz") {
		return SourceTypePackage
	}

	// Absolute paths start with "/"
	if strings.HasPrefix(location, "/") {
		return SourceTypeLocal
	}

	// Relative paths start with "." or ".."
	if strings.HasPrefix(location, ".") {
		return SourceTypeLocal
	}

	// Tilde-prefixed paths (home directory)
	if strings.HasPrefix(location, "~") {
		return SourceTypeLocal
	}

	// GitHub format: "owner/repo" or "owner/repo/subdir"
	// Simple heuristic: contains "/" but doesn't start with "/" or "." or "~"
	if strings.Contains(location, "/") {
		return SourceTypeGitHub
	}

	// Default to local for simple directory names
	return SourceTypeLocal
}

// parseFrontmatter extracts capability metadata from markdown content
func parseFrontmatter(content string) (CapabilityMetadata, error) {
	// Regex to extract frontmatter between --- delimiters
	// Support both LF (\n) and CRLF (\r\n) line endings
	frontmatterRegex := regexp.MustCompile(`(?s)^---\r?\n(.*?)\r?\n---`)
	matches := frontmatterRegex.FindStringSubmatch(content)

	if len(matches) < 2 {
		return CapabilityMetadata{}, fmt.Errorf("no frontmatter found")
	}

	frontmatterYAML := matches[1]

	// Parse YAML
	var metadata CapabilityMetadata
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &metadata); err != nil {
		return CapabilityMetadata{}, fmt.Errorf("failed to parse frontmatter YAML: %w", err)
	}

	// Validate required fields
	if metadata.Name == "" {
		return CapabilityMetadata{}, fmt.Errorf("name field is required")
	}

	// Parse keywords from comma-separated string
	if metadata.KeywordsRaw != "" {
		keywords := strings.Split(metadata.KeywordsRaw, ",")
		metadata.Keywords = make([]string, 0, len(keywords))
		for _, kw := range keywords {
			kw = strings.TrimSpace(kw)
			if kw != "" {
				metadata.Keywords = append(metadata.Keywords, kw)
			}
		}
	} else {
		metadata.Keywords = []string{}
	}

	return metadata, nil
}

// loadLocalCapabilities loads all capability files from a local directory
func loadLocalCapabilities(path string) ([]CapabilityMetadata, error) {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", path, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", path)
	}

	// Find all .md files in directory
	pattern := filepath.Join(path, "*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob .md files: %w", err)
	}

	capabilities := make([]CapabilityMetadata, 0, len(matches))

	for _, filePath := range matches {
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		metadata, err := parseFrontmatter(string(content))
		if err != nil {
			// Skip files with invalid frontmatter
			continue
		}

		// Set file path relative to source directory
		relPath, _ := filepath.Rel(path, filePath)
		metadata.FilePath = relPath

		capabilities = append(capabilities, metadata)
	}

	return capabilities, nil
}

// loadGitHubCapabilities loads capabilities from a GitHub repository
// This is a placeholder for future implementation
func loadGitHubCapabilities(repo string) ([]CapabilityMetadata, error) {
	// TODO: Implement GitHub API integration in future stage
	return nil, fmt.Errorf("GitHub capability loading not yet implemented")
}

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

// expandTilde expands ~ to the user's home directory
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path // Return unchanged if home dir can't be determined
	}

	if path == "~" {
		return homeDir
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}

	return path // Return unchanged if it starts with ~ but not ~/
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

// loadPackageCapabilities loads capabilities from a package source
func loadPackageCapabilities(packageLocation string) ([]CapabilityMetadata, error) {
	// Get cache directory
	cacheDir, err := getPackageCacheDir(packageLocation)
	if err != nil {
		return nil, err
	}

	// Check if cache is valid
	if !isCacheValidForPackage(packageLocation) {
		// Cache expired or doesn't exist, download and extract
		if err := downloadAndExtractPackage(packageLocation, cacheDir); err != nil {
			// If network error, try using stale cache
			if isCacheStaleForPackage(packageLocation) {
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

// mergeSources merges capabilities from multiple sources with priority-based overriding
func mergeSources(sources []CapabilitySource) (CapabilityIndex, error) {
	index := make(CapabilityIndex)

	// Process sources in reverse priority order (lower priority first)
	// So that higher priority sources (lower index) overwrite later
	for i := len(sources) - 1; i >= 0; i-- {
		source := sources[i]

		var capabilities []CapabilityMetadata
		var err error

		switch source.Type {
		case SourceTypeLocal:
			capabilities, err = loadLocalCapabilities(source.Location)
		case SourceTypeGitHub:
			capabilities, err = loadGitHubCapabilities(source.Location)
		case SourceTypePackage:
			capabilities, err = loadPackageCapabilities(source.Location)
		default:
			return nil, fmt.Errorf("unknown source type: %s", source.Type)
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

// CapabilityCache represents cached capability index with TTL
type CapabilityCache struct {
	Index     CapabilityIndex
	Timestamp time.Time
	TTL       time.Duration
	Sources   []CapabilitySource // Track sources for validation
}

var globalCapabilityCache *CapabilityCache
var cacheMutex sync.RWMutex

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

// getCapabilityIndex returns capability index with caching support
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

// hasLocalSources checks if any source is a local filesystem source
func hasLocalSources(sources []CapabilitySource) bool {
	for _, source := range sources {
		if source.Type == SourceTypeLocal {
			return true
		}
	}
	return false
}

// isCacheValid checks if the cache is still valid based on TTL
func isCacheValid(sources []CapabilitySource) bool {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if globalCapabilityCache == nil {
		return false
	}

	age := time.Since(globalCapabilityCache.Timestamp)
	return age < globalCapabilityCache.TTL
}

// isCacheStale checks if cache is stale but still usable (expired but within 7 days)
func isCacheStale(sources []CapabilitySource) bool {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if globalCapabilityCache == nil {
		return false
	}

	age := time.Since(globalCapabilityCache.Timestamp)
	maxStaleAge := 7 * 24 * time.Hour // 7 days

	// Cache is stale if: expired (age >= TTL) but within maxStaleAge
	return age >= globalCapabilityCache.TTL && age < maxStaleAge
}

// getPackageHash computes the hash for a package URL (for cache directory naming)
func getPackageHash(packageURL string) string {
	hash := sha256.Sum256([]byte(packageURL))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars
}

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

// getPackageTTL returns TTL for a package based on URL type
func getPackageTTL(packageURL string) time.Duration {
	// Release packages: 7 days (immutable)
	if isReleasePackage(packageURL) {
		return 7 * 24 * time.Hour
	}

	// Branch packages: 1 hour (mutable)
	return 1 * time.Hour
}

// isReleasePackage checks if package URL points to a release
func isReleasePackage(packageURL string) bool {
	return strings.Contains(packageURL, "/releases/")
}

// isCacheValidForPackage checks if cached package is still valid
func isCacheValidForPackage(packageURL string) bool {
	metadata, err := loadCacheMetadata()
	if err != nil {
		return false
	}

	// Get cache hash
	hash := getPackageHash(packageURL)

	pkg, exists := metadata.Packages[hash]
	if !exists {
		return false
	}

	// Check if cache directory actually exists
	cacheDir, err := getPackageCacheDir(packageURL)
	if err != nil {
		return false
	}
	commandsDir := filepath.Join(cacheDir, "commands")
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		// Cache metadata exists but directory is missing - invalidate cache
		return false
	}

	// Check TTL
	age := time.Since(pkg.DownloadTime)
	ttl := time.Duration(pkg.TTL) * time.Second

	return age < ttl
}

// isCacheStaleForPackage checks if cache is expired but still usable (< 7 days old)
func isCacheStaleForPackage(packageURL string) bool {
	metadata, err := loadCacheMetadata()
	if err != nil {
		return false
	}

	hash := getPackageHash(packageURL)

	pkg, exists := metadata.Packages[hash]
	if !exists {
		return false
	}

	age := time.Since(pkg.DownloadTime)
	ttl := time.Duration(pkg.TTL) * time.Second
	maxStaleAge := 7 * 24 * time.Hour

	return age >= ttl && age < maxStaleAge
}

// updateCacheMetadata updates cache metadata after downloading a package
func updateCacheMetadata(packageURL string, cacheDir string) error {
	metadata, err := loadCacheMetadata()
	if err != nil {
		return err
	}

	hash := getPackageHash(packageURL)
	ttl := getPackageTTL(packageURL)

	metadata.Packages[hash] = PackageCacheMetadata{
		PackageURL:    packageURL,
		DownloadTime:  time.Now(),
		ExtractedPath: cacheDir,
		TTL:           int64(ttl.Seconds()),
		IsRelease:     isReleasePackage(packageURL),
	}

	return saveCacheMetadata(metadata)
}

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

// isServerError checks if error is a 5xx server error
func isServerError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains "status 5"
	return strings.Contains(err.Error(), "status 5")
}

// isNetworkUnreachableError checks if error is a network unreachable error
func isNetworkUnreachableError(err error) bool {
	if err == nil {
		return false
	}
	// Check for network-related errors
	return strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "network is unreachable")
}

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

// executeListCapabilitiesTool handles the list_capabilities MCP tool
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

	// Check cache control (hidden test parameter)
	disableCache := false
	if disable, ok := args["_disable_cache"].(bool); ok {
		disableCache = disable
	}

	// Get capability index
	index, err := getCapabilityIndex(sources, disableCache)
	if err != nil {
		return "", fmt.Errorf("failed to get capability index: %w", err)
	}

	// Convert to array for JSON output
	capabilities := make([]CapabilityMetadata, 0, len(index))
	for _, cap := range index {
		capabilities = append(capabilities, cap)
	}

	// Sort by name for deterministic output
	// Note: We can't use sort.Slice here without importing "sort"
	// Let's add it in a simple way
	for i := 0; i < len(capabilities); i++ {
		for j := i + 1; j < len(capabilities); j++ {
			if capabilities[i].Name > capabilities[j].Name {
				capabilities[i], capabilities[j] = capabilities[j], capabilities[i]
			}
		}
	}

	// Build response
	result := map[string]interface{}{
		"mode":         "inline",
		"capabilities": capabilities,
		"source_count": len(sources),
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// notFoundError represents an error for capability not found
type notFoundError struct {
	name string
}

func (e *notFoundError) Error() string {
	return fmt.Sprintf("capability not found: %s", e.name)
}

func newNotFoundError(name string) error {
	return &notFoundError{name: name}
}

func isNotFoundError(err error) bool {
	_, ok := err.(*notFoundError)
	return ok
}

// getCapabilityContent retrieves the complete content of a capability from sources
func getCapabilityContent(name string, sources []CapabilitySource) (string, error) {
	if len(sources) == 0 {
		return "", fmt.Errorf("capability not found: %s (no sources configured)", name)
	}

	// Search sources in priority order (left-to-right = high-to-low)
	for _, source := range sources {
		var content string
		var err error

		switch source.Type {
		case SourceTypeLocal:
			content, err = readLocalCapability(name, source.Location)
		case SourceTypeGitHub:
			content, err = readGitHubCapability(name, source.Location)
		default:
			return "", fmt.Errorf("unknown source type: %s", source.Type)
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

// readLocalCapability reads a capability file from local filesystem
func readLocalCapability(name string, path string) (string, error) {
	// Construct file path
	filePath := filepath.Join(path, name+".md")

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

// buildCachePath constructs the cache file path for a GitHub source
// Cache directory structure: .capabilities-cache/github/{owner}/{repo}/{branch}/{subdir}/{filename}
func buildCachePath(source GitHubSource, filename string) string {
	parts := []string{
		CapabilityCacheDir,
		"github",
		source.Owner,
		source.Repo,
		source.Branch,
	}

	if source.Subdir != "" {
		parts = append(parts, source.Subdir)
	}

	parts = append(parts, filename)
	return filepath.Join(parts...)
}

// readGitHubCapability reads a capability file from GitHub repository via jsDelivr
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

// enhanceNotFoundError provides actionable error messages for 404 errors
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
	msg += "\nSuggestion: Run /meta to see available capabilities"

	return fmt.Errorf("%s", msg)
}

// executeGetCapabilityTool handles the get_capability MCP tool
func executeGetCapabilityTool(args map[string]interface{}) (string, error) {
	// Get capability name
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("missing required parameter: name")
	}

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

	// Get capability content
	content, err := getCapabilityContent(name, sources)
	if err != nil {
		return "", fmt.Errorf("failed to get capability: %w", err)
	}

	// Build response (inline mode)
	result := map[string]interface{}{
		"mode":    "inline",
		"name":    name,
		"content": content,
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
