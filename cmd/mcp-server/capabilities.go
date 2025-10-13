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
	// Uses GitHub Release package for fast, reliable, offline-friendly distribution
	DefaultCapabilitySource = "https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"

	// LocalCapabilitySource defines the local capability source for development
	LocalCapabilitySource = "capabilities/commands"
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

// Session cache variables
var (
	sessionCacheDir     string
	sessionCacheInitErr error
	sessionCacheOnce    sync.Once
)

// GitHubSource represents a parsed GitHub source with branch/tag
type GitHubSource struct {
	Owner  string // Repository owner
	Repo   string // Repository name
	Branch string // Branch, tag, or commit hash
	Subdir string // Optional subdirectory
}

// getSessionCacheDir returns the session-scoped cache directory
// Creates temp directory on first call, reuses for subsequent calls in same session
func getSessionCacheDir() (string, error) {
	sessionCacheOnce.Do(func() {
		// Try to get session ID from environment
		sessionID := os.Getenv("CLAUDE_CODE_SESSION_ID")
		if sessionID == "" {
			// Fallback: use process ID
			sessionID = fmt.Sprintf("mcp-%d", os.Getpid())
		}

		// Create session temp directory
		tempBase := os.TempDir()
		sessionDir := filepath.Join(tempBase, fmt.Sprintf("claude-session-%s", sessionID))

		// Create cache directory within session dir
		cacheDir := filepath.Join(sessionDir, ".meta-cc-capabilities")

		// Create directory if it doesn't exist
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			sessionCacheInitErr = fmt.Errorf("failed to create session cache dir: %w", err)
			return
		}

		sessionCacheDir = cacheDir
	})

	if sessionCacheInitErr != nil {
		return "", sessionCacheInitErr
	}

	return sessionCacheDir, nil
}

// CleanupSessionCache removes the session cache directory
// Should be called on MCP server shutdown
func CleanupSessionCache() error {
	if sessionCacheDir == "" {
		return nil
	}

	// Remove the entire session directory (parent of cache dir)
	sessionDir := filepath.Dir(sessionCacheDir)
	if err := os.RemoveAll(sessionDir); err != nil {
		return fmt.Errorf("failed to cleanup session cache: %w", err)
	}

	return nil
}

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
	// But check if it's actually a local directory first
	if strings.Contains(location, "/") {
		// Check if it's a local directory that exists
		if _, err := os.Stat(location); err == nil {
			return SourceTypeLocal
		}
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

// getPackageCacheDir returns the session cache directory for a package source
func getPackageCacheDir(packageLocation string) (string, error) {
	// Get session cache base directory
	sessionCache, err := getSessionCacheDir()
	if err != nil {
		return "", err
	}

	// Compute hash of package location for subdirectory name
	hash := sha256.Sum256([]byte(packageLocation))
	hashStr := hex.EncodeToString(hash[:])[:16] // Use first 16 chars

	// Cache directory: /tmp/claude-session-<id>/.meta-cc-capabilities/packages/<hash>/
	cacheDir := filepath.Join(sessionCache, "packages", hashStr)
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
// Downloads once per session, uses session cache for subsequent calls
func loadPackageCapabilities(packageLocation string) ([]CapabilityMetadata, error) {
	// Get cache directory
	cacheDir, err := getPackageCacheDir(packageLocation)
	if err != nil {
		return nil, err
	}

	// Check if already downloaded in this session
	commandsDir := filepath.Join(cacheDir, "commands")
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		// Not yet downloaded, download and extract
		if err := downloadAndExtractPackage(packageLocation, cacheDir); err != nil {
			return nil, fmt.Errorf("failed to download package: %w", err)
		}
	}

	// Load capabilities from extracted package
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

// SessionCapabilityCache represents the session-scoped capability index cache
// Simple in-memory cache, valid for the entire session (no TTL)
type SessionCapabilityCache struct {
	Index   CapabilityIndex
	Sources []CapabilitySource // Track sources for validation
}

var sessionCapabilityCache *SessionCapabilityCache
var sessionCacheMutex sync.RWMutex

// getCapabilityIndex returns capability index with session-scoped caching
// Local sources bypass cache, package sources use session temp directory
func getCapabilityIndex(sources []CapabilitySource, disableCache bool) (CapabilityIndex, error) {
	// Local sources always bypass cache (for development workflow)
	hasLocal := hasLocalSources(sources)

	// Check session cache if enabled and no local sources
	if !disableCache && !hasLocal && sessionCapabilityCache != nil {
		sessionCacheMutex.RLock()
		defer sessionCacheMutex.RUnlock()
		return sessionCapabilityCache.Index, nil
	}

	// Load fresh data
	index, err := mergeSources(sources)
	if err != nil {
		return nil, err
	}

	// Update session cache (skip if local sources present)
	if !disableCache && !hasLocal {
		sessionCacheMutex.Lock()
		sessionCapabilityCache = &SessionCapabilityCache{
			Index:   index,
			Sources: sources,
		}
		sessionCacheMutex.Unlock()
	}

	return index, nil
}

// hasLocalSources checks if any of the sources is a local filesystem source
func hasLocalSources(sources []CapabilitySource) bool {
	for _, src := range sources {
		if src.Type == SourceTypeLocal {
			return true
		}
	}
	return false
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
		// Default to GitHub Release package if no sources configured
		sources = []CapabilitySource{
			{Type: SourceTypePackage, Location: DefaultCapabilitySource, Priority: 0},
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
		case SourceTypePackage:
			content, err = readPackageCapability(name, source.Location)
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

// readPackageCapability reads a capability file from a package source
// Uses session cache, downloads once per session
func readPackageCapability(name string, packageLocation string) (string, error) {
	// Get cache directory
	cacheDir, err := getPackageCacheDir(packageLocation)
	if err != nil {
		return "", err
	}

	// Check if already downloaded in this session
	commandsDir := filepath.Join(cacheDir, "commands")
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		// Not yet downloaded, download and extract
		if err := downloadAndExtractPackage(packageLocation, cacheDir); err != nil {
			return "", fmt.Errorf("failed to download package: %w", err)
		}
	}

	// Read capability file from extracted package
	return readLocalCapability(name, commandsDir)
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
		// Default to GitHub Release package if no sources configured
		sources = []CapabilitySource{
			{Type: SourceTypePackage, Location: DefaultCapabilitySource, Priority: 0},
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
