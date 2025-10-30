package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/query"
)

// handlers_stage1.go implements Stage 1 tools of the two-stage query architecture
// Stage 1: Metadata and directory inspection tools for query planning
// Stage 2: File-level inspection tools (implemented in future stages)

// handleGetSessionDirectory implements get_session_directory tool
// Returns session directory path and metadata based on scope
func handleGetSessionDirectory(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse scope parameter
	scope, ok := args["scope"].(string)
	if !ok || scope == "" {
		return nil, fmt.Errorf("scope parameter is required")
	}

	// Validate scope
	if scope != "session" && scope != "project" {
		return nil, fmt.Errorf("invalid scope: %s (must be 'session' or 'project')", scope)
	}

	// Get directory path based on scope
	directory, err := getDirectoryForScope(scope)
	if err != nil {
		return nil, err
	}

	// Collect metadata about the directory
	metadata, err := collectDirectoryMetadata(directory)
	if err != nil {
		return nil, err
	}

	// Build response
	response := map[string]interface{}{
		"directory":        directory,
		"scope":            scope,
		"file_count":       metadata.FileCount,
		"total_size_bytes": metadata.TotalSize,
		"oldest_file":      metadata.OldestFile,
		"newest_file":      metadata.NewestFile,
	}

	return response, nil
}

// directoryMetadata holds metadata about a session directory
type directoryMetadata struct {
	FileCount  int
	TotalSize  int64
	OldestFile string // RFC3339 timestamp
	NewestFile string // RFC3339 timestamp
}

// getDirectoryForScope returns the directory path for the given scope
func getDirectoryForScope(scope string) (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	loc := locator.NewSessionLocator()

	// For session scope: return directory containing the current session
	if scope == "session" {
		sessionFile, err := loc.FromProjectPath(cwd)
		if err != nil {
			return "", fmt.Errorf("failed to locate current session: %w", err)
		}
		return filepath.Dir(sessionFile), nil
	}

	// For project scope: return directory containing all project sessions
	sessionFiles, err := loc.AllSessionsFromProject(cwd)
	if err != nil {
		return "", fmt.Errorf("failed to locate project sessions: %w", err)
	}

	if len(sessionFiles) == 0 {
		return "", fmt.Errorf("no sessions found for project")
	}

	// All session files are in the same directory
	return filepath.Dir(sessionFiles[0]), nil
}

// collectDirectoryMetadata scans a directory and collects metadata about .jsonl files
func collectDirectoryMetadata(directory string) (*directoryMetadata, error) {
	metadata := &directoryMetadata{
		FileCount:  0,
		TotalSize:  0,
		OldestFile: "",
		NewestFile: "",
	}

	// Find all .jsonl files in the directory
	pattern := filepath.Join(directory, "*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	// Track oldest and newest modification times
	var oldestTime, newestTime time.Time

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue // Skip files we can't stat
		}

		// Count files and accumulate size
		metadata.FileCount++
		metadata.TotalSize += info.Size()

		// Track oldest and newest files
		modTime := info.ModTime()
		if oldestTime.IsZero() || modTime.Before(oldestTime) {
			oldestTime = modTime
		}
		if newestTime.IsZero() || modTime.After(newestTime) {
			newestTime = modTime
		}
	}

	// Format timestamps as RFC3339 (or empty string if no files)
	if !oldestTime.IsZero() {
		metadata.OldestFile = oldestTime.Format(time.RFC3339)
	}
	if !newestTime.IsZero() {
		metadata.NewestFile = newestTime.Format(time.RFC3339)
	}

	return metadata, nil
}

// handleInspectSessionFiles implements inspect_session_files tool
// Returns detailed metadata about specified session files
func handleInspectSessionFiles(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse files parameter
	filesRaw, ok := args["files"]
	if !ok {
		return nil, fmt.Errorf("files parameter is required")
	}

	filesInterface, ok := filesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("files must be an array")
	}

	// Convert to string array
	files := make([]string, 0, len(filesInterface))
	for i, fileRaw := range filesInterface {
		file, ok := fileRaw.(string)
		if !ok {
			return nil, fmt.Errorf("file at index %d is not a string", i)
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("files array cannot be empty")
	}

	// Parse include_samples parameter (optional, defaults to false)
	includeSamples := false
	if samplesRaw, ok := args["include_samples"]; ok {
		includeSamples, ok = samplesRaw.(bool)
		if !ok {
			return nil, fmt.Errorf("include_samples must be a boolean")
		}
	}

	// Call the file inspector
	result, err := query.InspectFiles(files, includeSamples)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect files: %w", err)
	}

	return result, nil
}

// handleGetSessionMetadata implements get_session_metadata tool
// Returns metadata needed for constructing queries including JSONL schema, file info, and query templates
func handleGetSessionMetadata(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse scope parameter (defaults to "project")
	scope := "project"
	if scopeRaw, ok := args["scope"].(string); ok && scopeRaw != "" {
		scope = scopeRaw
	}

	// Validate scope
	if scope != "session" && scope != "project" {
		return nil, fmt.Errorf("invalid scope: %s (must be 'session' or 'project')", scope)
	}

	// Get base directory for the scope
	baseDir, err := getQueryBaseDir(scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get base directory for scope %s: %w", scope, err)
	}

	// Get JSONL files in directory
	files, err := getJSONLFiles(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list JSONL files: %w", err)
	}

	// Collect file metadata
	fileMetadata := make([]map[string]interface{}, 0, len(files))
	for _, file := range files {
		// Get file info
		info, err := os.Stat(file)
		if err != nil {
			// Skip files we can't stat
			continue
		}

		// Estimate record count by counting lines (approximate)
		recordCount, err := countLines(file)
		if err != nil {
			recordCount = 0 // Use 0 if we can't count lines
		}

		fileMetadata = append(fileMetadata, map[string]interface{}{
			"path":        file,
			"size_bytes":  info.Size(),
			"modified_at": info.ModTime().Format(time.RFC3339),
			"records":     recordCount,
		})
	}

	// Define JSONL schema (simplified version)
	jsonlSchema := map[string]interface{}{
		"common_fields": []map[string]string{
			{"name": "type", "description": "Record type (user, assistant, system, summary, etc.)"},
			{"name": "timestamp", "description": "ISO8601 timestamp of the record"},
			{"name": "message", "description": "Message content with structured data"},
			{"name": "cwd", "description": "Current working directory"},
			{"name": "gitBranch", "description": "Git branch at time of record"},
		},
		"user_message_fields": []map[string]string{
			{"name": "message.content", "description": "User message content (string or array of content blocks)"},
			{"name": "message.role", "description": "Always 'user' for user messages"},
		},
		"assistant_message_fields": []map[string]string{
			{"name": "message.content", "description": "Assistant response content (array of content blocks)"},
			{"name": "message.role", "description": "Always 'assistant' for assistant messages"},
			{"name": "message.usage", "description": "Token usage statistics"},
		},
		"tool_fields": []map[string]string{
			{"name": "message.content[].type", "description": "Content block type (text, tool_use, tool_result)"},
			{"name": "message.content[].name", "description": "Tool name (for tool_use blocks)"},
			{"name": "message.content[].input", "description": "Tool input parameters (for tool_use blocks)"},
			{"name": "message.content[].is_error", "description": "Error flag (for tool_result blocks)"},
		},
	}

	// Load query templates
	templateMap, err := query.LoadTemplates()
	if err != nil {
		// If we can't load templates, use basic examples
		templateMap = make(map[string]query.QueryTemplate)
	}

	// Convert templates to the format expected by the response
	queryTemplates := make(map[string]interface{})
	for name, template := range templateMap {
		// Convert examples to simple strings for the response
		examples := make([]string, len(template.Examples))
		for i, example := range template.Examples {
			examples[i] = example.Command
		}

		queryTemplates[name] = map[string]interface{}{
			"description": template.Description,
			"filter":      template.Filter,
			"category":    template.Category,
			"examples":    examples,
			"parameters":  template.Parameters,
		}
	}

	// If we don't have templates loaded, provide basic examples
	if len(queryTemplates) == 0 {
		queryTemplates = map[string]interface{}{
			"user_messages": map[string]interface{}{
				"description": "Filter for user messages",
				"filter":      "select(.type == \"user\")",
				"category":    "message_type",
			},
			"assistant_messages": map[string]interface{}{
				"description": "Filter for assistant messages",
				"filter":      "select(.type == \"assistant\")",
				"category":    "message_type",
			},
			"tool_errors": map[string]interface{}{
				"description": "Filter for tool errors",
				"filter":      "select(.type == \"user\" and (.message.content | type == \"array\")) | select(.message.content[] | select(.type == \"tool_result\" and .is_error == true))",
				"category":    "error_analysis",
			},
			"time_range": map[string]interface{}{
				"description": "Filter by time range (example: last 24 hours)",
				"filter":      "select(.timestamp >= \"2025-10-29T00:00:00Z\")",
				"category":    "time_filtering",
			},
			"smart_file_filter": map[string]interface{}{
				"description": "Smart file filtering based on metadata",
				"filter":      "# Use file metadata to construct efficient file selection",
				"category":    "file_filtering",
			},
		}
	}

	// Construct response
	result := map[string]interface{}{
		"scope":           scope,
		"base_dir":        baseDir,
		"file_count":      len(fileMetadata),
		"files":           fileMetadata,
		"jsonl_schema":    jsonlSchema,
		"query_templates": queryTemplates,
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// countLines counts the number of lines in a file (approximate record count)
func countLines(filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lineCount, nil
}
