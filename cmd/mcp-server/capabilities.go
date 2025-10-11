package main

import (
	"encoding/json"
	"fmt"
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

// parseCapabilitySources parses the environment variable into a list of capability sources
func parseCapabilitySources(envVar string) []CapabilitySource {
	if envVar == "" {
		return []CapabilitySource{}
	}

	parts := strings.Split(envVar, ":")
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

// detectSourceType determines if a location is a local path or GitHub repository
func detectSourceType(location string) SourceType {
	// Absolute paths start with "/"
	if strings.HasPrefix(location, "/") {
		return SourceTypeLocal
	}

	// Relative paths start with "." or ".."
	if strings.HasPrefix(location, ".") {
		return SourceTypeLocal
	}

	// GitHub format: "owner/repo" or "owner/repo/subdir"
	// Simple heuristic: contains "/" but doesn't start with "/" or "."
	if strings.Contains(location, "/") {
		return SourceTypeGitHub
	}

	// Default to local for simple directory names
	return SourceTypeLocal
}

// parseFrontmatter extracts capability metadata from markdown content
func parseFrontmatter(content string) (CapabilityMetadata, error) {
	// Regex to extract frontmatter between --- delimiters
	frontmatterRegex := regexp.MustCompile(`(?s)^---\n(.*?)\n---`)
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
			// Log warning but continue processing other files
			continue
		}

		metadata, err := parseFrontmatter(string(content))
		if err != nil {
			// Log warning but continue processing other files
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
}

var globalCapabilityCache *CapabilityCache
var cacheMutex sync.RWMutex

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
		// Default to .claude/commands if no sources configured
		sources = []CapabilitySource{
			{Type: SourceTypeLocal, Location: ".claude/commands", Priority: 0},
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

// readGitHubCapability reads a capability file from GitHub repository
// This is a placeholder for future implementation
func readGitHubCapability(name string, repo string) (string, error) {
	// TODO: Implement GitHub API integration in future stage
	return "", fmt.Errorf("GitHub capability loading not yet implemented")
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
		// Default to .claude/commands if no sources configured
		sources = []CapabilitySource{
			{Type: SourceTypeLocal, Location: ".claude/commands", Priority: 0},
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
