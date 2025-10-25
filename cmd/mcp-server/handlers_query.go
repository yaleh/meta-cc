package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaleh/meta-cc/internal/config"
)

// handleQuery implements the Layer 2 query tool with jq_filter and jq_transform
func (e *ToolExecutor) handleQuery(cfg *config.Config, scope string, args map[string]interface{}) (string, error) {
	// Extract parameters
	jqFilter := getStringParam(args, "jq_filter", "")
	jqTransform := getStringParam(args, "jq_transform", "")
	limit := getIntParam(args, "limit", 0)

	// Get base directory using pipeline infrastructure
	baseDir, err := getQueryBaseDir(scope)
	if err != nil {
		return "", fmt.Errorf("failed to get base directory: %w", err)
	}

	// Create query executor
	executor := NewQueryExecutor(baseDir)

	// Build combined expression
	expression := executor.buildExpression(jqFilter, jqTransform)

	// Compile expression
	code, err := executor.compileExpression(expression)
	if err != nil {
		return "", fmt.Errorf("invalid jq expression: %w", err)
	}

	// Get all JSONL files in directory
	files, err := getJSONLFiles(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to list JSONL files: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no JSONL files found in %s", baseDir)
	}

	// Execute query with streaming
	ctx := context.Background()
	results := executor.streamFiles(ctx, files, code, limit)

	// Serialize results to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(jsonData), nil
}

// handleQueryRaw implements the Layer 3 raw query tool with jq_expression
func (e *ToolExecutor) handleQueryRaw(cfg *config.Config, scope string, args map[string]interface{}) (string, error) {
	// Extract required parameter
	jqExpression, ok := args["jq_expression"].(string)
	if !ok || jqExpression == "" {
		return "", fmt.Errorf("jq_expression parameter is required")
	}

	limit := getIntParam(args, "limit", 0)

	// Get base directory using pipeline infrastructure
	baseDir, err := getQueryBaseDir(scope)
	if err != nil {
		return "", fmt.Errorf("failed to get base directory: %w", err)
	}

	// Create query executor
	executor := NewQueryExecutor(baseDir)

	// Compile expression
	code, err := executor.compileExpression(jqExpression)
	if err != nil {
		return "", fmt.Errorf("invalid jq expression: %w", err)
	}

	// Get all JSONL files in directory
	files, err := getJSONLFiles(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to list JSONL files: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no JSONL files found in %s", baseDir)
	}

	// Execute query with streaming
	ctx := context.Background()
	results := executor.streamFiles(ctx, files, code, limit)

	// Serialize results to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(jsonData), nil
}

// getQueryBaseDir returns the base directory for the given scope
func getQueryBaseDir(scope string) (string, error) {
	// Use current working directory as project path (same as buildPipelineOptions)
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Get base directory from environment
	// For session scope, we need to use CLAUDE_PROJECT_DIR
	// For project scope, we use the parent directory
	if scope == "session" {
		claudeProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
		if claudeProjectDir != "" {
			return claudeProjectDir, nil
		}
		// Fall back to current directory if env not set
		return cwd, nil
	}

	// Project scope: use parent directory of session dir
	claudeProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	if claudeProjectDir != "" {
		return filepath.Dir(claudeProjectDir), nil
	}

	// Fall back to current directory if env not set
	return cwd, nil
}

// getJSONLFiles returns all .jsonl files in a directory (non-recursive)
func getJSONLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".jsonl" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}
