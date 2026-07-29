package query

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/provider/rawfiles"
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

// stageProviderArg resolves the stage-1 `provider` argument (DIR-073): an
// explicit claude/codex/all passes through unchanged; an omitted/empty value
// resolves to the host that launched this MCP process via
// config.OmittedProviderDefault — the same single source of truth the
// executor handlers, pipeline, raw-file selection, and analysis dispatch use.
func stageProviderArg(args map[string]interface{}) string {
	if providerName, ok := args["provider"].(string); ok && providerName != "" {
		return providerName
	}
	return config.OmittedProviderDefault()
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

	providerName := stageProviderArg(args)
	workingDir, _ := args["working_dir"].(string)

	switch providerName {
	case "claude":
		return buildClaudeDirectoryResult(scope, workingDir)
	case "codex":
		return buildCodexDirectoryResult(ctx, scope, workingDir)
	case "all":
		return buildAllDirectoryResult(ctx, scope, workingDir)
	default:
		return nil, fmt.Errorf("invalid provider %q: must be \"claude\", \"codex\", or \"all\"", providerName)
	}
}

// buildClaudeDirectoryResult implements the original (pre-DIR-024) Claude-only
// get_session_directory behavior unchanged, so an explicit provider="claude"
// (and the host default when launched from Claude Code) remains backward
// compatible.
func buildClaudeDirectoryResult(scope, workingDir string) (map[string]interface{}, error) {
	directory, err := GetDirectoryForScope(scope, workingDir)
	if err != nil {
		return nil, err
	}

	metadata, err := CollectDirectoryMetadata(directory)
	if err != nil {
		return nil, err
	}

	// Count subagent files (always, regardless of include_subagents — so callers know what they're opting out of)
	mainFiles, _ := GetQueryFiles(scope, workingDir, false)
	allFiles, _ := GetQueryFiles(scope, workingDir, true)
	subagentFileCount := len(allFiles) - len(mainFiles)
	if subagentFileCount < 0 {
		subagentFileCount = 0
	}

	return map[string]interface{}{
		"provider":            string(conversation.ProviderClaude),
		"directory":           directory,
		"scope":               scope,
		"file_count":          len(mainFiles),
		"total_size_bytes":    metadata.TotalSize,
		"oldest_file":         metadata.OldestFile,
		"newest_file":         metadata.NewestFile,
		"subagent_file_count": subagentFileCount,
	}, nil
}

// buildCodexDirectoryResult resolves raw Codex rollout files for the given
// scope/project and summarizes them. Codex sessions are not necessarily
// stored under one flat directory (unlike Claude's per-project session
// directory), so the response exposes an explicit "files" list; "directory"
// is only populated when every selected file happens to share one parent.
func buildCodexDirectoryResult(ctx context.Context, scope, workingDir string) (map[string]interface{}, error) {
	projectPath, err := resolveProjectPath(workingDir)
	if err != nil {
		return nil, err
	}

	registry := rawfiles.NewRegistry(projectPath)
	files, err := rawfiles.SelectCodexFiles(ctx, registry, scope, projectPath)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(files))
	var totalSize int64
	var oldest, newest time.Time
	for _, f := range files {
		info, statErr := os.Stat(f.Path)
		if statErr != nil {
			continue
		}
		paths = append(paths, f.Path)
		totalSize += info.Size()
		modTime := info.ModTime()
		if oldest.IsZero() || modTime.Before(oldest) {
			oldest = modTime
		}
		if newest.IsZero() || modTime.After(newest) {
			newest = modTime
		}
	}

	result := map[string]interface{}{
		"provider":         string(conversation.ProviderCodex),
		"scope":            scope,
		"files":            paths,
		"file_count":       len(paths),
		"total_size_bytes": totalSize,
		"directory":        commonDirectory(paths),
	}
	if !oldest.IsZero() {
		result["oldest_file"] = oldest.Format(time.RFC3339)
	}
	if !newest.IsZero() {
		result["newest_file"] = newest.Format(time.RFC3339)
	}
	return result, nil
}

// buildAllDirectoryResult returns an explicit per-provider breakdown so
// callers never mistake Claude's directory-based corpus for Codex's raw-file
// corpus (or vice versa). A provider that legitimately has no data produces a
// warning rather than being silently dropped; the call only fails outright
// when *no* provider produced data.
func buildAllDirectoryResult(ctx context.Context, scope, workingDir string) (map[string]interface{}, error) {
	providers := map[string]interface{}{}
	var warnings []string

	if claudeResult, err := buildClaudeDirectoryResult(scope, workingDir); err != nil {
		warnings = append(warnings, fmt.Sprintf("claude: %v", err))
	} else {
		providers["claude"] = claudeResult
	}

	if codexResult, err := buildCodexDirectoryResult(ctx, scope, workingDir); err != nil {
		warnings = append(warnings, fmt.Sprintf("codex: %v", err))
	} else {
		providers["codex"] = codexResult
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no session data available for any provider: %s", strings.Join(warnings, "; "))
	}

	return map[string]interface{}{
		"provider":  "all",
		"scope":     scope,
		"providers": providers,
		"warnings":  warnings,
	}, nil
}

// resolveProjectPath resolves the effective project path for provider-aware
// discovery: an explicit working_dir argument, absolutized, or the server's
// current working directory when none was supplied.
func resolveProjectPath(workingDir string) (string, error) {
	projectPath := workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to determine working directory: %w", err)
		}
		projectPath = cwd
	}
	if abs, err := filepath.Abs(projectPath); err == nil {
		return abs, nil
	}
	return projectPath, nil
}

// commonDirectory returns the shared parent directory of paths when every
// path lives in the same directory, or nil when they diverge (Codex rollout
// files may be scattered across multiple directories, e.g. one per day).
func commonDirectory(paths []string) interface{} {
	if len(paths) == 0 {
		return nil
	}
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		if filepath.Dir(p) != dir {
			return nil
		}
	}
	return dir
}

// GetDirectoryForScope returns the Claude session directory path for the
// given scope. workingDir overrides the server's current working directory
// when non-empty.
func GetDirectoryForScope(scope, workingDir string) (string, error) {
	projectPath := workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		projectPath = cwd
	}

	loc := locator.NewSessionLocator()

	if scope == "session" {
		sessionFile, err := loc.FromProjectPath(projectPath)
		if err != nil {
			return "", fmt.Errorf("failed to locate current session: %w", err)
		}
		return filepath.Dir(sessionFile), nil
	}

	sessionFiles, err := loc.AllSessionsFromProject(projectPath)
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

	providerName := stageProviderArg(args)
	workingDir, _ := args["working_dir"].(string)

	switch providerName {
	case "claude":
		return buildClaudeMetadataResult(scope, workingDir)
	case "codex":
		return buildCodexMetadataResult(ctx, scope, workingDir)
	case "all":
		return buildAllMetadataResult(ctx, scope, workingDir)
	default:
		return nil, fmt.Errorf("invalid provider %q: must be \"claude\", \"codex\", or \"all\"", providerName)
	}
}

// buildClaudeMetadataResult implements the original (pre-DIR-024) Claude-only
// get_session_metadata behavior unchanged, so an explicit provider="claude"
// (and the host default when launched from Claude Code) remains backward
// compatible.
func buildClaudeMetadataResult(scope, workingDir string) (map[string]interface{}, error) {
	// Get JSONL files for the scope (include subagents by default)
	files, err := GetQueryFiles(scope, workingDir, true)
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
		queryTemplates = defaultClaudeQueryTemplates()
	}

	return map[string]interface{}{
		"provider":        string(conversation.ProviderClaude),
		"scope":           scope,
		"base_dir":        baseDir,
		"file_count":      len(fileMetadata),
		"files":           fileMetadata,
		"jsonl_schema":    claudeJSONLSchema(),
		"query_templates": queryTemplates,
		"timestamp":       time.Now().Format(time.RFC3339),
	}, nil
}

// buildCodexMetadataResult resolves raw Codex rollout files for the given
// scope/project and describes them using Codex's own raw event schema.
// Codex rollout files use a fundamentally different shape than Claude's
// JSONL (session_meta/turn_context/event_msg/response_item for the legacy
// schema, or thread.started/turn.started/item.* for the new schema), so the
// schema hints and query templates returned here are Codex-specific rather
// than a lossy projection onto Claude's schema.
func buildCodexMetadataResult(ctx context.Context, scope, workingDir string) (map[string]interface{}, error) {
	projectPath, err := resolveProjectPath(workingDir)
	if err != nil {
		return nil, err
	}

	registry := rawfiles.NewRegistry(projectPath)
	files, err := rawfiles.SelectCodexFiles(ctx, registry, scope, projectPath)
	if err != nil {
		return nil, err
	}

	fileMetadata := make([]map[string]interface{}, 0, len(files))
	for _, f := range files {
		info, statErr := os.Stat(f.Path)
		if statErr != nil {
			continue
		}

		recordCount, countErr := CountLines(f.Path)
		if countErr != nil {
			recordCount = 0
		}

		fileMetadata = append(fileMetadata, map[string]interface{}{
			"path":        f.Path,
			"session_id":  f.SessionID,
			"size_bytes":  info.Size(),
			"modified_at": info.ModTime().Format(time.RFC3339),
			"records":     recordCount,
		})
	}

	return map[string]interface{}{
		"provider":        string(conversation.ProviderCodex),
		"scope":           scope,
		"file_count":      len(fileMetadata),
		"files":           fileMetadata,
		"jsonl_schema":    codexJSONLSchema(),
		"query_templates": codexQueryTemplates(),
		"timestamp":       time.Now().Format(time.RFC3339),
	}, nil
}

// buildAllMetadataResult returns an explicit per-provider breakdown of file
// metadata and schema hints. It never merges Claude and Codex file lists or
// schemas — see rawfiles package docs for why that would be misleading.
func buildAllMetadataResult(ctx context.Context, scope, workingDir string) (map[string]interface{}, error) {
	providers := map[string]interface{}{}
	var warnings []string

	if claudeResult, err := buildClaudeMetadataResult(scope, workingDir); err != nil {
		warnings = append(warnings, fmt.Sprintf("claude: %v", err))
	} else {
		providers["claude"] = claudeResult
	}

	if codexResult, err := buildCodexMetadataResult(ctx, scope, workingDir); err != nil {
		warnings = append(warnings, fmt.Sprintf("codex: %v", err))
	} else {
		providers["codex"] = codexResult
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no session data available for any provider: %s", strings.Join(warnings, "; "))
	}

	return map[string]interface{}{
		"provider":  "all",
		"scope":     scope,
		"providers": providers,
		"warnings":  warnings,
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}

// claudeJSONLSchema documents the Claude Code session JSONL record shape.
func claudeJSONLSchema() map[string]interface{} {
	return map[string]interface{}{
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
}

// defaultClaudeQueryTemplates is the fallback template set used when
// catalog.LoadTemplates() finds nothing on disk.
func defaultClaudeQueryTemplates() map[string]interface{} {
	return map[string]interface{}{
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

// codexJSONLSchema documents the raw Codex rollout record shape. Codex has
// shipped two on-disk schema versions; both are documented here since a
// single rollout file only ever uses one, detected from its first line.
func codexJSONLSchema() map[string]interface{} {
	return map[string]interface{}{
		"common_fields": []map[string]string{
			{"name": "timestamp", "description": "ISO8601 timestamp of the record"},
			{"name": "type", "description": "Record type. Legacy schema: session_meta, turn_context, event_msg, response_item. New schema: thread.started, turn.started, item.message, item.tool_call, item.tool_result, turn.completed"},
			{"name": "payload", "description": "Event-specific payload; shape depends on type (see legacy_response_item_fields / new_schema_fields)"},
		},
		"legacy_response_item_fields": []map[string]string{
			{"name": "payload.type", "description": "message | function_call | function_call_output | custom_tool_call | custom_tool_call_output | reasoning (for type == response_item)"},
			{"name": "payload.role", "description": "user | assistant | developer | system (for payload.type == message)"},
			{"name": "payload.content", "description": "Array of content blocks with .type and .text (for payload.type == message)"},
			{"name": "payload.name", "description": "Tool name (for payload.type == function_call / custom_tool_call)"},
			{"name": "payload.output", "description": "Tool output (for payload.type == function_call_output / custom_tool_call_output)"},
			{"name": "payload.is_error", "description": "Error flag (for payload.type == function_call_output / custom_tool_call_output)"},
		},
		"legacy_event_msg_fields": []map[string]string{
			{"name": "payload.type", "description": "user_message | agent_message | token_count | task_started (for type == event_msg)"},
			{"name": "payload.message", "description": "Message text (for payload.type == user_message / agent_message)"},
		},
		"new_schema_fields": []map[string]string{
			{"name": "payload.role", "description": "user | assistant (for type == item.message)"},
			{"name": "payload.content", "description": "Message text (for type == item.message)"},
			{"name": "payload.name", "description": "Tool name (for type == item.tool_call)"},
			{"name": "payload.input", "description": "Tool input (for type == item.tool_call)"},
			{"name": "payload.output", "description": "Tool output (for type == item.tool_result)"},
			{"name": "payload.is_error", "description": "Error flag (for type == item.tool_result)"},
		},
	}
}

// codexQueryTemplates provides jq starting points against Codex's raw
// (non-normalized) rollout schema.
func codexQueryTemplates() map[string]interface{} {
	return map[string]interface{}{
		"user_messages": map[string]interface{}{
			"description": "Filter for user messages (legacy + new schema)",
			"filter":      `select((.type == "response_item" and .payload.type == "message" and .payload.role == "user") or (.type == "event_msg" and .payload.type == "user_message") or (.type == "item.message" and .payload.role == "user"))`,
			"category":    "message_type",
		},
		"assistant_messages": map[string]interface{}{
			"description": "Filter for assistant messages (legacy + new schema)",
			"filter":      `select((.type == "response_item" and .payload.type == "message" and .payload.role == "assistant") or (.type == "event_msg" and .payload.type == "agent_message") or (.type == "item.message" and .payload.role == "assistant"))`,
			"category":    "message_type",
		},
		"tool_errors": map[string]interface{}{
			"description": "Filter for failed tool calls (legacy + new schema)",
			"filter":      `select((.type == "response_item" and (.payload.type == "function_call_output" or .payload.type == "custom_tool_call_output") and .payload.is_error == true) or (.type == "item.tool_result" and .payload.is_error == true))`,
			"category":    "error_analysis",
		},
		"time_range": map[string]interface{}{
			"description": "Filter by time range (example: last 24 hours)",
			"filter":      "select(.timestamp >= \"2025-10-29T00:00:00Z\")",
			"category":    "time_filtering",
		},
	}
}

// CountLines counts the number of lines in a file (approximate record count).
//
// NOTE(DIR-038): this hand-rolled bufio.NewReader + ReadBytes('\n') loop
// predates and duplicates the shape now crystallized in
// parser.ReadLineBounded (internal/parser/bounded_reader.go). Migrating it
// is tracked as optional/best-effort follow-up, not done here, to keep this
// task's diff scoped to closing the check-no-scanner violation in
// internal/provider/codex/appserver/client.go.
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
