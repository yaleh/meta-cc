package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/yaleh/meta-cc/internal/locator"
)

// handleQuery and handleQueryRaw deleted in Phase 27 Stage 27.1
// These tools were removed to simplify the query interface
// Users should use the 10 shortcut query tools instead

// executeQuery is an internal helper for convenience tools
// It executes a jq query and returns results as []interface{}
// This allows proper JSONL formatting by response adapters
func (e *ToolExecutor) executeQuery(scope string, jqFilter string, limit int) ([]interface{}, error) {
	// Get base directory using pipeline infrastructure
	baseDir, err := getQueryBaseDir(scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get base directory: %w", err)
	}

	// Create query executor
	executor := NewQueryExecutor(baseDir)

	// Compile expression
	code, err := executor.compileExpression(jqFilter)
	if err != nil {
		return nil, fmt.Errorf("invalid jq expression: %w", err)
	}

	// Get all JSONL files in directory
	files, err := getJSONLFiles(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list JSONL files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no JSONL files found in %s", baseDir)
	}

	// Execute query with streaming
	ctx := context.Background()
	results := executor.streamFiles(ctx, files, code, limit)

	// Return results directly as []interface{}
	// Response adapters will handle serialization (inline or file_ref)
	return results, nil
}

// getQueryBaseDir returns the base directory for the given scope
// For session scope: returns directory of most recently modified session file
// For project scope: returns directory containing all session files
func getQueryBaseDir(scope string) (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	loc := locator.NewSessionLocator()

	// Session scope: return directory of most recently modified session file
	if scope == "session" {
		// Use FromProjectPath to find the newest session file
		sessionFile, err := loc.FromProjectPath(cwd)
		if err != nil {
			return "", fmt.Errorf("failed to locate current session: %w", err)
		}

		// Return the directory containing the session file
		return filepath.Dir(sessionFile), nil
	}

	// Project scope: use SessionLocator to find all session files
	// This matches the behavior of buildPipelineOptions + SessionPipeline.Load

	// Determine project path (same logic as buildPipelineOptions)
	projectPath := cwd

	// AllSessionsFromProject returns the list of session files
	// We need to return the directory containing those files
	sessionFiles, err := loc.AllSessionsFromProject(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to locate project sessions: %w", err)
	}

	if len(sessionFiles) == 0 {
		return "", fmt.Errorf("no sessions found for project: %s", projectPath)
	}

	// All session files should be in the same directory
	// Return the directory of the first session file
	return filepath.Dir(sessionFiles[0]), nil
}

// getJSONLFiles returns all .jsonl files in a directory (non-recursive)
// Files are sorted by modification time (newest first) to prioritize recent sessions
func getJSONLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Collect file info with modification times
	type fileInfo struct {
		path    string
		modTime int64 // Unix timestamp for easier sorting
	}
	var fileInfos []fileInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".jsonl" {
			fullPath := filepath.Join(dir, entry.Name())

			// Get file stat for modification time
			info, err := entry.Info()
			if err != nil {
				// Skip files we can't stat
				continue
			}

			fileInfos = append(fileInfos, fileInfo{
				path:    fullPath,
				modTime: info.ModTime().Unix(),
			})
		}
	}

	// Sort by modification time (newest first = descending order)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime > fileInfos[j].modTime
	})

	// Extract paths
	var files []string
	for _, fi := range fileInfos {
		files = append(files, fi.path)
	}

	return files, nil
}
