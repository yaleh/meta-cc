package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/config"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	querypkg "github.com/yaleh/meta-cc/internal/query"
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
		rawOutput, err = e.handleQuery(cfg, scope, args)
	case "query_raw":
		rawOutput, err = e.handleQueryRaw(cfg, scope, args)

	// Layer 1: Convenience Tools (10 high-frequency queries)
	case "query_user_messages":
		rawOutput, err = e.handleQueryUserMessages(cfg, scope, args)
	case "query_tools":
		rawOutput, err = e.handleQueryTools(cfg, scope, args)
	case "query_tool_errors":
		rawOutput, err = e.handleQueryToolErrors(cfg, scope, args)
	case "query_token_usage":
		rawOutput, err = e.handleQueryTokenUsage(cfg, scope, args)
	case "query_conversation_flow":
		rawOutput, err = e.handleQueryConversationFlow(cfg, scope, args)
	case "query_system_errors":
		rawOutput, err = e.handleQuerySystemErrors(cfg, scope, args)
	case "query_file_snapshots":
		rawOutput, err = e.handleQueryFileSnapshots(cfg, scope, args)
	case "query_timestamps":
		rawOutput, err = e.handleQueryTimestamps(cfg, scope, args)
	case "query_summaries":
		rawOutput, err = e.handleQuerySummaries(cfg, scope, args)
	case "query_tool_blocks":
		rawOutput, err = e.handleQueryToolBlocks(cfg, scope, args)
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

	// Phase 25 Fix: All Phase 25 tools execute jq internally, so we should
	// NOT apply jq_filter again to avoid double application.
	//
	// Phase 25 tools (12 total):
	// - query, query_raw: Execute jq in handleQuery/handleQueryRaw
	// - 10 convenience tools: Call handleQuery internally, which executes jq
	//
	// Therefore, NO Phase 25 tool should have jq_filter applied post-processing.
	phase25Tools := map[string]bool{
		"query": true, "query_raw": true,
		"query_user_messages": true, "query_tools": true, "query_tool_errors": true,
		"query_token_usage": true, "query_conversation_flow": true, "query_system_errors": true,
		"query_file_snapshots": true, "query_timestamps": true, "query_summaries": true,
		"query_tool_blocks": true,
	}

	filtered := rawOutput
	isPhase25Tool := phase25Tools[toolName]
	shouldApplyJQFilter := !isPhase25Tool

	if shouldApplyJQFilter {
		var err error
		filtered, err = querypkg.ApplyJQFilter(rawOutput, config.jqFilter)
		if err != nil {
			slog.Error("jq filter application failed",
				"tool_name", toolName,
				"jq_filter", config.jqFilter,
				"error", err.Error(),
				"error_type", "execution_error",
			)
			return "", fmt.Errorf("jq filter error for tool %s: %w", toolName, err)
		}
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
