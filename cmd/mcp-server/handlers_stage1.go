package main

import (
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
