package query

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/query/catalog"
	"github.com/yaleh/meta-cc/internal/query/engine"
	queryfiles "github.com/yaleh/meta-cc/internal/query/files"
)

// DirectoryMetadata holds metadata about a session directory
type DirectoryMetadata struct {
	FileCount  int
	TotalSize  int64
	OldestFile string // RFC3339 timestamp
	NewestFile string // RFC3339 timestamp
}

// HandleGetSessionDirectory implements get_session_directory tool
func HandleGetSessionDirectory(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	scope, ok := args["scope"].(string)
	if !ok || scope == "" {
		return nil, fmt.Errorf("scope parameter is required")
	}

	if scope != "session" && scope != "project" {
		return nil, fmt.Errorf("invalid scope: %s (must be 'session' or 'project')", scope)
	}

	directory, err := GetDirectoryForScope(scope)
	if err != nil {
		return nil, err
	}

	metadata, err := CollectDirectoryMetadata(directory)
	if err != nil {
		return nil, err
	}

	// Count subagent files (always, regardless of include_subagents — so callers know what they're opting out of)
	cwd, _ := os.Getwd()
	mainFiles, _ := GetQueryFiles(scope, cwd, false)
	allFiles, _ := GetQueryFiles(scope, cwd, true)
	subagentFileCount := len(allFiles) - len(mainFiles)
	if subagentFileCount < 0 {
		subagentFileCount = 0
	}

	return map[string]interface{}{
		"directory":           directory,
		"scope":               scope,
		"file_count":          len(mainFiles),
		"total_size_bytes":    metadata.TotalSize,
		"oldest_file":         metadata.OldestFile,
		"newest_file":         metadata.NewestFile,
		"subagent_file_count": subagentFileCount,
	}, nil
}

// GetDirectoryForScope returns the directory path for the given scope
func GetDirectoryForScope(scope string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	loc := locator.NewSessionLocator()

	if scope == "session" {
		sessionFile, err := loc.FromProjectPath(cwd)
		if err != nil {
			return "", fmt.Errorf("failed to locate current session: %w", err)
		}
		return filepath.Dir(sessionFile), nil
	}

	sessionFiles, err := loc.AllSessionsFromProject(cwd)
	if err != nil {
		return "", fmt.Errorf("failed to locate project sessions: %w", err)
	}

	if len(sessionFiles) == 0 {
		return "", fmt.Errorf("no sessions found for project")
	}

	return filepath.Dir(sessionFiles[0]), nil
}

// CollectDirectoryMetadata scans a directory and collects metadata about .jsonl files
func CollectDirectoryMetadata(directory string) (*DirectoryMetadata, error) {
	metadata := &DirectoryMetadata{}

	pattern := filepath.Join(directory, "*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	var oldestTime, newestTime time.Time

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		metadata.FileCount++
		metadata.TotalSize += info.Size()

		modTime := info.ModTime()
		if oldestTime.IsZero() || modTime.Before(oldestTime) {
			oldestTime = modTime
		}
		if newestTime.IsZero() || modTime.After(newestTime) {
			newestTime = modTime
		}
	}

	if !oldestTime.IsZero() {
		metadata.OldestFile = oldestTime.Format(time.RFC3339)
	}
	if !newestTime.IsZero() {
		metadata.NewestFile = newestTime.Format(time.RFC3339)
	}

	return metadata, nil
}

// HandleInspectSessionFiles implements inspect_session_files tool
func HandleInspectSessionFiles(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filesRaw, ok := args["files"]
	if !ok {
		return nil, fmt.Errorf("files parameter is required")
	}

	filesInterface, ok := filesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("files must be an array")
	}

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

	includeSamples := false
	if samplesRaw, ok := args["include_samples"]; ok {
		includeSamples, ok = samplesRaw.(bool)
		if !ok {
			return nil, fmt.Errorf("include_samples must be a boolean")
		}
	}

	result, err := queryfiles.InspectFiles(files, includeSamples)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect files: %w", err)
	}

	return result, nil
}

// HandleExecuteStage2Query implements execute_stage2_query tool
func HandleExecuteStage2Query(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filesRaw, ok := args["files"]
	if !ok {
		return nil, fmt.Errorf("files parameter is required")
	}

	filesInterface, ok := filesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("files must be an array")
	}

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

	filter, ok := args["filter"].(string)
	if !ok || filter == "" {
		return nil, fmt.Errorf("filter parameter is required")
	}

	sort := ""
	if sortRaw, ok := args["sort"]; ok {
		sort, _ = sortRaw.(string)
	}

	transform := ""
	if transformRaw, ok := args["transform"]; ok {
		transform, _ = transformRaw.(string)
	}

	limit := 0
	if limitRaw, ok := args["limit"]; ok {
		switch v := limitRaw.(type) {
		case float64:
			limit = int(v)
		case int:
			limit = v
		}
	}

	stage2Query := &engine.Stage2Query{
		Files:     files,
		Filter:    filter,
		Sort:      sort,
		Transform: transform,
		Limit:     limit,
	}

	result, err := engine.ExecuteStage2Query(stage2Query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute stage 2 query: %w", err)
	}

	warnings := result.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return map[string]interface{}{
		"results": result.Results,
		"metadata": map[string]interface{}{
			"execution_time_ms":     result.Metadata.ExecutionTimeMs,
			"files_processed":       result.Metadata.FilesProcessed,
			"total_records_scanned": result.Metadata.TotalRecordsScanned,
			"results_returned":      result.Metadata.ResultsReturned,
			"truncated":             result.Metadata.Truncated,
		},
		"warnings": warnings,
	}, nil
}

// HandleGetSessionMetadata implements get_session_metadata tool
func HandleGetSessionMetadata(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse scope parameter (defaults to "project")
	scope := "project"
	if scopeRaw, ok := args["scope"].(string); ok && scopeRaw != "" {
		scope = scopeRaw
	}

	// Validate scope
	if scope != "session" && scope != "project" {
		return nil, fmt.Errorf("invalid scope: %s (must be 'session' or 'project')", scope)
	}

	// Get JSONL files for the scope (include subagents by default)
	files, err := GetQueryFiles(scope, "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to get JSONL files for scope %s: %w", scope, err)
	}
	// Derive base directory from first file for schema documentation
	baseDir := ""
	if len(files) > 0 {
		baseDir = filepath.Dir(files[0])
	}

	// Collect file metadata
	fileMetadata := make([]map[string]interface{}, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		recordCount, err := CountLines(file)
		if err != nil {
			recordCount = 0
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
	templateMap, err := catalog.LoadTemplates()
	if err != nil {
		templateMap = make(map[string]catalog.QueryTemplate)
	}

	queryTemplates := make(map[string]interface{})
	for name, template := range templateMap {
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

	return map[string]interface{}{
		"scope":           scope,
		"base_dir":        baseDir,
		"file_count":      len(fileMetadata),
		"files":           fileMetadata,
		"jsonl_schema":    jsonlSchema,
		"query_templates": queryTemplates,
		"timestamp":       time.Now().Format(time.RFC3339),
	}, nil
}

// CountLines counts the number of lines in a file (approximate record count).
func CountLines(filename string) (int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	count := 0
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			count++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
	}
	return count, nil
}
