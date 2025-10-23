package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/config"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	filterpkg "github.com/yaleh/meta-cc/internal/filter"
	internalOutput "github.com/yaleh/meta-cc/internal/output"
	"github.com/yaleh/meta-cc/internal/parser"
	querypkg "github.com/yaleh/meta-cc/internal/query"
	pkgoutput "github.com/yaleh/meta-cc/pkg/output"
	pipelinepkg "github.com/yaleh/meta-cc/pkg/pipeline"
)

type ToolExecutor struct{}

type toolPipelineConfig struct {
	jqFilter         string
	statsOnly        bool
	statsFirst       bool
	outputFormat     string
	maxMessageLength int
	contentSummary   bool
}

func newToolPipelineConfig(args map[string]interface{}) toolPipelineConfig {
	return toolPipelineConfig{
		jqFilter:         getStringParam(args, "jq_filter", ".[]"),
		statsOnly:        getBoolParam(args, "stats_only", false),
		statsFirst:       getBoolParam(args, "stats_first", false),
		outputFormat:     getStringParam(args, "output_format", "jsonl"),
		maxMessageLength: getIntParam(args, "max_message_length", 0),
		contentSummary:   getBoolParam(args, "content_summary", false),
	}
}

func (c toolPipelineConfig) requiresMessageFilters() bool {
	return c.maxMessageLength > 0 || c.contentSummary
}

func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{}
}

func determineScope(toolName string, args map[string]interface{}) string {
	defaultScope := "project"
	if toolName == "get_session_stats" {
		defaultScope = "session"
	}
	return getStringParam(args, "scope", defaultScope)
}

func recordToolSuccess(toolName, scope string, start time.Time) {
	elapsed := time.Since(start)
	RecordToolCall(toolName, scope, "success")
	RecordToolExecutionDuration(toolName, scope, elapsed)
}

func recordToolFailure(toolName, scope string, start time.Time, errorType string) {
	elapsed := time.Since(start)
	RecordToolCall(toolName, scope, "error")
	RecordToolExecutionDuration(toolName, scope, elapsed)
	RecordError(toolName, errorType, GetErrorSeverity(errorType))
}

func (e *ToolExecutor) executeSpecialTool(cfg *config.Config, toolName, scope string, args map[string]interface{}, start time.Time) (string, bool, error) {
	switch toolName {
	case "cleanup_temp_files":
		output, err := executeCleanupTool(args)
		if err != nil {
			errorType := classifyError(err)
			recordToolFailure(toolName, scope, start, errorType)
			return "", true, err
		}
		recordToolSuccess(toolName, scope, start)
		return output, true, nil

	case "list_capabilities":
		output, err := executeListCapabilitiesTool(cfg, args)
		if err != nil {
			errorType := classifyError(err)
			recordToolFailure(toolName, scope, start, errorType)
			return "", true, err
		}
		recordToolSuccess(toolName, scope, start)
		return output, true, nil

	case "get_capability":
		output, err := executeGetCapabilityTool(cfg, args)
		if err != nil {
			errorType := classifyError(err)
			recordToolFailure(toolName, scope, start, errorType)
			return "", true, err
		}
		recordToolSuccess(toolName, scope, start)
		return output, true, nil

	default:
		return "", false, nil
	}
}

// ExecuteTool executes a meta-cc command and applies jq filtering
func (e *ToolExecutor) ExecuteTool(cfg *config.Config, toolName string, args map[string]interface{}) (string, error) {
	scope := determineScope(toolName, args)
	start := time.Now()

	if output, handled, err := e.executeSpecialTool(cfg, toolName, scope, args, start); handled {
		return output, err
	}

	config := newToolPipelineConfig(args)
	var rawOutput string
	var err error

	switch toolName {
	case "query":
		rawOutput, err = e.executeQuery(scope, config, args)
	case "query_tools":
		rawOutput, err = e.executeQueryTools(scope, config, args)
	case "query_tools_advanced":
		rawOutput, err = e.executeQueryTools(scope, config, args)
	case "query_user_messages":
		rawOutput, err = e.executeQueryUserMessages(scope, config, args)
	case "query_assistant_messages":
		rawOutput, err = e.executeQueryAssistantMessages(scope, config, args)
	case "query_context":
		rawOutput, err = e.executeQueryContext(scope, config, args)
	case "query_tool_sequences":
		rawOutput, err = e.executeQueryToolSequences(scope, config, args)
	case "query_file_access":
		rawOutput, err = e.executeQueryFileAccess(scope, config, args)
	case "query_files":
		rawOutput, err = e.executeQueryFiles(scope, config, args)
	case "query_conversation":
		rawOutput, err = e.executeQueryConversation(scope, config, args)
	case "get_session_stats":
		rawOutput, err = e.executeGetSessionStats(scope, config, args)
	case "query_time_series":
		rawOutput, err = e.executeQueryTimeSeries(scope, config, args)
	case "query_project_state":
		rawOutput, err = e.executeQueryProjectState(scope, config, args)
	case "query_successful_prompts":
		rawOutput, err = e.executeQuerySuccessfulPrompts(scope, config, args)
	default:
		// All query tools must be handled explicitly above.
		// No CLI fallback - all tools use internal/query library.
		recordToolFailure(toolName, scope, start, "validation_error")
		return "", fmt.Errorf("unknown tool %s in executor: %w", toolName, mcerrors.ErrUnknownTool)
	}

	if err != nil {
		errorType := classifyError(err)
		slog.Error("tool execution failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", errorType,
		)
		recordToolFailure(toolName, scope, start, errorType)
		return "", err
	}

	filtered, err := querypkg.ApplyJQFilter(rawOutput, config.jqFilter)
	if err != nil {
		slog.Error("jq filter application failed",
			"tool_name", toolName,
			"jq_filter", config.jqFilter,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", fmt.Errorf("jq filter error for tool %s: %w", toolName, err)
	}

	parsedData, err := e.parseJSONL(filtered)
	if err != nil {
		slog.Error("JSONL parsing failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", fmt.Errorf("JSONL parse error for tool %s: %w", toolName, mcerrors.ErrParseError)
	}

	if toolName == "query_user_messages" && config.requiresMessageFilters() {
		parsedData = e.applyMessageFiltersToData(parsedData, config.maxMessageLength, config.contentSummary)
	}

	output, err := e.buildResponse(cfg, parsedData, args, toolName, config)
	if err != nil {
		return "", err
	}

	slog.Debug("tool execution pipeline completed successfully",
		"tool_name", toolName,
		"output_length", len(output),
	)

	recordToolSuccess(toolName, scope, start)
	return output, nil
}

func buildPipelineOptions(scope string) pipelinepkg.GlobalOptions {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	opts := pipelinepkg.GlobalOptions{
		ProjectPath: cwd,
	}

	if scope == "session" {
		opts.SessionOnly = true
	}

	return opts
}

func (e *ToolExecutor) executeQueryTools(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	options := querypkg.ToolsQueryOptions{
		Pipeline:   buildPipelineOptions(scope),
		Limit:      getIntParam(args, "limit", 0),
		Offset:     getIntParam(args, "offset", 0),
		SortBy:     getStringParam(args, "sort_by", ""),
		Reverse:    getBoolParam(args, "reverse", false),
		Status:     getStringParam(args, "status", ""),
		Tool:       getStringParam(args, "tool", ""),
		Where:      getStringParam(args, "where", ""),
		Expression: getStringParam(args, "filter", ""),
	}

	results, err := querypkg.RunToolsQuery(options)
	if err != nil {
		return "", normalizeQueryError(err)
	}

	format := cfg.outputFormat
	if format == "" {
		format = "jsonl"
	}

	formatted, err := internalOutput.FormatOutput(results, format)
	if err != nil {
		return "", err
	}

	return formatted, nil
}

func (e *ToolExecutor) executeQueryUserMessages(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	contextWindow := getIntParam(args, "with_context", 0)
	if contextWindow == 0 {
		contextWindow = getIntParam(args, "context", 0)
	}

	options := querypkg.UserMessagesQueryOptions{
		Pipeline: buildPipelineOptions(scope),
		Pattern:  getStringParam(args, "pattern", ""),
		Context:  contextWindow,
		Limit:    getIntParam(args, "limit", 0),
		Offset:   getIntParam(args, "offset", 0),
		SortBy:   getStringParam(args, "sort_by", ""),
		Reverse:  getBoolParam(args, "reverse", false),
	}

	messages, err := querypkg.RunUserMessagesQuery(options)
	if err != nil {
		return "", normalizeQueryError(err)
	}

	format := cfg.outputFormat
	if format == "" {
		format = "jsonl"
	}

	switch format {
	case "jsonl":
		return pkgoutput.FormatJSONL(messages)
	case "tsv":
		return pkgoutput.FormatTSV(messages)
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

func (e *ToolExecutor) executeQueryContext(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	errorSig := getStringParam(args, "error_signature", "")
	if errorSig == "" {
		return "", fmt.Errorf("%w: error_signature", mcerrors.ErrMissingParameter)
	}

	window := getIntParam(args, "window", 3)
	result, err := querypkg.BuildContextQuery(entries, errorSig, window)
	if err != nil {
		return "", err
	}

	return pkgoutput.FormatJSONL([]interface{}{result})
}

func (e *ToolExecutor) executeQueryToolSequences(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	minOccurrences := getIntParam(args, "min_occurrences", 3)
	pattern := getStringParam(args, "pattern", "")
	includeBuiltin := getBoolParam(args, "include_builtin_tools", false)

	result, err := querypkg.BuildToolSequenceQuery(entries, minOccurrences, pattern, includeBuiltin)
	if err != nil {
		return "", err
	}

	return pkgoutput.FormatJSONL(result.Sequences)
}

func (e *ToolExecutor) executeQueryFileAccess(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	file := getStringParam(args, "file", "")
	if file == "" {
		return "", fmt.Errorf("%w: file", mcerrors.ErrMissingParameter)
	}

	result, err := querypkg.BuildFileAccessQuery(entries, file)
	if err != nil {
		return "", err
	}

	return pkgoutput.FormatJSONL([]interface{}{result})
}

func (e *ToolExecutor) executeQueryAssistantMessages(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	options := querypkg.AssistantMessagesOptions{
		Pattern:   getStringParam(args, "pattern", ""),
		MinTools:  getIntParam(args, "min_tools", -1),
		MaxTools:  getIntParam(args, "max_tools", -1),
		MinTokens: getIntParam(args, "min_tokens_output", -1),
		MinLength: getIntParam(args, "min_length", -1),
		MaxLength: getIntParam(args, "max_length", -1),
		Limit:     getIntParam(args, "limit", 0),
		Offset:    getIntParam(args, "offset", 0),
		SortBy:    getStringParam(args, "sort_by", ""),
		Reverse:   getBoolParam(args, "reverse", false),
	}

	messages, err := querypkg.BuildAssistantMessages(entries, options)
	if err != nil {
		return "", err
	}

	switch cfg.outputFormat {
	case "", "jsonl":
		return pkgoutput.FormatJSONL(messages)
	case "tsv":
		return pkgoutput.FormatTSV(messages)
	default:
		return "", fmt.Errorf("unsupported output format: %s", cfg.outputFormat)
	}
}

func (e *ToolExecutor) executeQueryConversation(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	options := querypkg.ConversationOptions{
		StartTurn:     getIntParam(args, "start_turn", -1),
		EndTurn:       getIntParam(args, "end_turn", -1),
		Pattern:       getStringParam(args, "pattern", ""),
		PatternTarget: getStringParam(args, "pattern_target", "any"),
		MinDuration:   getIntParam(args, "min_duration", -1),
		MaxDuration:   getIntParam(args, "max_duration", -1),
		Limit:         getIntParam(args, "limit", 0),
		Offset:        getIntParam(args, "offset", 0),
		SortBy:        getStringParam(args, "sort_by", ""),
		Reverse:       getBoolParam(args, "reverse", false),
	}

	turns, err := querypkg.BuildConversationTurns(entries, options)
	if err != nil {
		return "", err
	}

	switch cfg.outputFormat {
	case "", "jsonl":
		return pkgoutput.FormatJSONL(turns)
	case "tsv":
		return pkgoutput.FormatTSV(turns)
	default:
		return "", fmt.Errorf("unsupported output format: %s", cfg.outputFormat)
	}
}

func (e *ToolExecutor) executeQueryFiles(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	threshold := getIntParam(args, "threshold", 5)
	files := querypkg.DetectFileChurn(entries, querypkg.FileChurnOptions{Threshold: threshold})

	switch cfg.outputFormat {
	case "", "jsonl":
		return pkgoutput.FormatJSONL(files)
	case "tsv":
		return pkgoutput.FormatTSV(files)
	default:
		return "", fmt.Errorf("unsupported output format: %s", cfg.outputFormat)
	}
}

func (e *ToolExecutor) executeGetSessionStats(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	toolCalls := parser.ExtractToolCalls(entries)
	stats := querypkg.BuildSessionStats(entries, toolCalls)

	format := cfg.outputFormat
	if format == "" {
		format = "jsonl"
	}

	formatted, err := internalOutput.FormatOutput(stats, format)
	if err != nil {
		return "", err
	}
	return formatted, nil
}

func (e *ToolExecutor) executeQueryTimeSeries(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}
	toolCalls := parser.ExtractToolCalls(entries)

	points, err := querypkg.AnalyzeTimeSeries(toolCalls, getStringParam(args, "metric", "tool-calls"), getStringParam(args, "interval", "hour"), getStringParam(args, "where", ""))
	if err != nil {
		return "", err
	}

	format := cfg.outputFormat
	if format == "" {
		format = "jsonl"
	}

	var outputStr string
	if format == "jsonl" {
		outputStr, err = pkgoutput.FormatJSONL(points)
	} else if format == "tsv" {
		outputStr, err = pkgoutput.FormatTSV(points)
	} else {
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
	if err != nil {
		return "", err
	}
	return outputStr, nil
}

func (e *ToolExecutor) executeQueryProjectState(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	state := querypkg.BuildProjectState(entries, querypkg.ProjectStateOptions{IncludeIncomplete: getBoolParam(args, "include_incomplete_tasks", true)})

	format := cfg.outputFormat
	if format == "" || format == "jsonl" {
		return pkgoutput.FormatJSONL([]interface{}{state})
	}
	return internalOutput.FormatOutput(state, format)
}

func (e *ToolExecutor) executeQuerySuccessfulPrompts(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	minQuality := getFloatParam(args, "min_quality_score", 0)
	limit := getIntParam(args, "limit", 0)
	result := querypkg.BuildSuccessfulPrompts(entries, minQuality, limit)

	format := cfg.outputFormat
	if format == "" {
		format = "jsonl"
	}

	if format == "jsonl" {
		items := make([]interface{}, 0, len(result.Prompts))
		for _, prompt := range result.Prompts {
			items = append(items, prompt)
		}
		return pkgoutput.FormatJSONL(items)
	}
	return internalOutput.FormatOutput(result.Prompts, format)
}

func loadEntries(scope string, args map[string]interface{}) ([]parser.SessionEntry, error) {
	opts := buildPipelineOptions(scope)
	pipe := pipelinepkg.NewSessionPipeline(opts)
	if err := pipe.Load(pipelinepkg.LoadOptions{AutoDetect: true}); err != nil {
		return nil, fmt.Errorf("%w: %v", querypkg.ErrSessionLoad, err)
	}

	entries := pipe.Entries()
	since := getStringParam(args, "since", "")
	last := getIntParam(args, "last_n_turns", 0)
	from := getIntParam(args, "from", 0)
	to := getIntParam(args, "to", 0)

	if since != "" || last > 0 || from > 0 || to > 0 {
		timeFilter := filterpkg.TimeFilter{
			Since:      since,
			LastNTurns: last,
			FromTs:     int64(from),
			ToTs:       int64(to),
		}

		filteredEntries, err := timeFilter.Apply(entries)
		if err != nil {
			return nil, err
		}
		entries = filteredEntries
	}

	return entries, nil
}

func normalizeQueryError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, querypkg.ErrSessionLoad):
		return fmt.Errorf("failed to load session: %w", err)
	case errors.Is(err, querypkg.ErrFilterInvalid):
		return fmt.Errorf("invalid filter: %w", err)
	case errors.Is(err, querypkg.ErrInvalidPattern):
		return fmt.Errorf("invalid regex pattern: %w", err)
	default:
		return err
	}
}

func (e *ToolExecutor) buildResponse(cfg *config.Config, parsedData []interface{}, args map[string]interface{}, toolName string, pipeline toolPipelineConfig) (string, error) {
	if pipeline.statsOnly {
		return e.buildStatsOnlyResponse(parsedData, toolName)
	}

	if pipeline.statsFirst {
		return e.buildStatsFirstResponse(cfg, parsedData, args, toolName)
	}

	return e.buildStandardResponse(cfg, parsedData, args, toolName)
}

func (e *ToolExecutor) buildStatsOnlyResponse(parsedData []interface{}, toolName string) (string, error) {
	jsonlData, err := e.dataToJSONL(parsedData)
	if err != nil {
		slog.Error("dataToJSONL conversion failed (stats_only)",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", err
	}

	output, err := querypkg.GenerateStats(jsonlData)
	if err != nil {
		slog.Error("stats generation failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", err
	}

	return output, nil
}

func (e *ToolExecutor) buildStatsFirstResponse(cfg *config.Config, parsedData []interface{}, args map[string]interface{}, toolName string) (string, error) {
	jsonlData, err := e.dataToJSONL(parsedData)
	if err != nil {
		slog.Error("dataToJSONL conversion failed (stats_first)",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", err
	}

	stats, _ := querypkg.GenerateStats(jsonlData)
	response, err := adaptResponse(cfg, parsedData, args, toolName)
	if err != nil {
		slog.Error("response adaptation failed (stats_first)",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", err
	}

	serialized, err := serializeResponse(response)
	if err != nil {
		slog.Error("response serialization failed (stats_first)",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", err
	}

	return stats + "\n---\n" + serialized, nil
}

func (e *ToolExecutor) buildStandardResponse(cfg *config.Config, parsedData []interface{}, args map[string]interface{}, toolName string) (string, error) {
	response, err := adaptResponse(cfg, parsedData, args, toolName)
	if err != nil {
		slog.Error("response adaptation failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", fmt.Errorf("response adaptation error for tool %s: %w", toolName, err)
	}

	output, err := serializeResponse(response)
	if err != nil {
		slog.Error("response serialization failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", err
	}

	return output, nil
}

// parseJSONL parses JSONL string into array of interfaces
func (e *ToolExecutor) parseJSONL(jsonlData string) ([]interface{}, error) {
	jsonlData = strings.TrimSpace(jsonlData)

	// Handle special cases: empty input or "[]" (exit code 2 scenario)
	if jsonlData == "" || jsonlData == "[]" {
		slog.Debug("parseJSONL: empty input or no results",
			"input", jsonlData,
		)
		return []interface{}{}, nil
	}

	lines := strings.Split(jsonlData, "\n")
	var data []interface{}

	for i, line := range lines {
		if line == "" {
			continue
		}

		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			slog.Error("failed to parse JSONL line",
				"line_number", i+1,
				"line_content", line,
				"error", err.Error(),
				"error_type", "parse_error",
			)
			return nil, fmt.Errorf("invalid JSON on line %d: %w", i+1, mcerrors.ErrParseError)
		}
		data = append(data, obj)
	}

	slog.Debug("parseJSONL completed",
		"record_count", len(data),
	)

	return data, nil
}

// dataToJSONL converts array of interfaces to JSONL string
func (e *ToolExecutor) dataToJSONL(data []interface{}) (string, error) {
	var output strings.Builder
	for i, record := range data {
		jsonBytes, err := json.Marshal(record)
		if err != nil {
			slog.Error("failed to marshal record to JSON",
				"record_index", i,
				"error", err.Error(),
				"error_type", "parse_error",
			)
			return "", err
		}
		output.Write(jsonBytes)
		output.WriteString("\n")
	}
	return output.String(), nil
}

// applyMessageFiltersToData applies content truncation or summary mode to user messages (data array)
func (e *ToolExecutor) applyMessageFiltersToData(messages []interface{}, maxMessageLength int, contentSummary bool) []interface{} {
	if contentSummary {
		return ApplyContentSummary(messages)
	}
	return TruncateMessageContent(messages, maxMessageLength)
}

// Helper functions
func getStringParam(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}

func getBoolParam(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return defaultVal
}

func getIntParam(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	if v, ok := args[key].(int); ok {
		return v
	}
	return defaultVal
}

func getFloatParam(args map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := args[key].(float64); ok {
		return v
	}
	return defaultVal
}

// executeQuery executes the unified query interface
func (e *ToolExecutor) executeQuery(scope string, cfg toolPipelineConfig, args map[string]interface{}) (string, error) {
	// Load entries
	entries, err := loadEntries(scope, args)
	if err != nil {
		return "", err
	}

	// Parse QueryParams from args
	params, err := parseQueryParams(args)
	if err != nil {
		return "", fmt.Errorf("failed to parse query params: %w", err)
	}

	// Set scope
	params.Scope = scope

	// Execute unified query
	results, err := querypkg.Query(entries, params)
	if err != nil {
		return "", fmt.Errorf("query execution failed: %w", err)
	}

	// Format output based on cfg.outputFormat
	format := cfg.outputFormat
	if format == "" {
		format = "jsonl"
	}

	// Convert results to []interface{} for formatting
	var resultSlice []interface{}
	switch v := results.(type) {
	case []interface{}:
		resultSlice = v
	case map[string]interface{}:
		// Single result wrapped in array
		resultSlice = []interface{}{v}
	default:
		resultSlice = []interface{}{v}
	}

	return internalOutput.FormatOutput(resultSlice, format)
}

// parseQueryParams converts MCP tool args to QueryParams
func parseQueryParams(args map[string]interface{}) (querypkg.QueryParams, error) {
	params := querypkg.QueryParams{
		Resource: getStringParam(args, "resource", "entries"),
	}

	// Parse filter object
	if filterArg, ok := args["filter"].(map[string]interface{}); ok {
		params.Filter = parseFilterSpec(filterArg)
	}

	// Parse transform object
	if transformArg, ok := args["transform"].(map[string]interface{}); ok {
		params.Transform = parseTransformSpec(transformArg)
	}

	// Parse aggregate object
	if aggregateArg, ok := args["aggregate"].(map[string]interface{}); ok {
		params.Aggregate = parseAggregateSpec(aggregateArg)
	}

	// Parse output object (use standard params for now)
	params.Output = querypkg.OutputSpec{
		Format: getStringParam(args, "output_format", "jsonl"),
		Limit:  getIntParam(args, "limit", 0),
	}

	// jq_filter passed through
	params.JQFilter = getStringParam(args, "jq_filter", "")

	return params, nil
}

// parseFilterSpec converts filter args to FilterSpec
func parseFilterSpec(filterArgs map[string]interface{}) querypkg.FilterSpec {
	spec := querypkg.FilterSpec{
		Type:         getStringParam(filterArgs, "type", ""),
		SessionID:    getStringParam(filterArgs, "session_id", ""),
		UUID:         getStringParam(filterArgs, "uuid", ""),
		ParentUUID:   getStringParam(filterArgs, "parent_uuid", ""),
		GitBranch:    getStringParam(filterArgs, "git_branch", ""),
		Role:         getStringParam(filterArgs, "role", ""),
		ContentType:  getStringParam(filterArgs, "content_type", ""),
		ContentMatch: getStringParam(filterArgs, "content_match", ""),
		ToolName:     getStringParam(filterArgs, "tool_name", ""),
		ToolStatus:   getStringParam(filterArgs, "tool_status", ""),
	}

	// Handle has_error pointer
	if hasError, ok := filterArgs["has_error"].(bool); ok {
		spec.HasError = &hasError
	}

	// Handle time_range
	if timeRange, ok := filterArgs["time_range"].(map[string]interface{}); ok {
		spec.TimeRange = &querypkg.TimeRange{
			Start: getStringParam(timeRange, "start", ""),
			End:   getStringParam(timeRange, "end", ""),
		}
	}

	return spec
}

// parseTransformSpec converts transform args to TransformSpec
func parseTransformSpec(transformArgs map[string]interface{}) querypkg.TransformSpec {
	spec := querypkg.TransformSpec{
		GroupBy: getStringParam(transformArgs, "group_by", ""),
	}

	// Handle extract array
	if extractRaw, ok := transformArgs["extract"].([]interface{}); ok {
		extract := make([]string, 0, len(extractRaw))
		for _, e := range extractRaw {
			if str, ok := e.(string); ok {
				extract = append(extract, str)
			}
		}
		spec.Extract = extract
	}

	// Handle join object
	if joinArgs, ok := transformArgs["join"].(map[string]interface{}); ok {
		spec.Join = &querypkg.JoinSpec{
			Type: getStringParam(joinArgs, "type", ""),
			On:   getStringParam(joinArgs, "on", ""),
		}
	}

	return spec
}

// parseAggregateSpec converts aggregate args to AggregateSpec
func parseAggregateSpec(aggregateArgs map[string]interface{}) querypkg.AggregateSpec {
	return querypkg.AggregateSpec{
		Function: getStringParam(aggregateArgs, "function", ""),
		Field:    getStringParam(aggregateArgs, "field", ""),
	}
}
