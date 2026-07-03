package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/analysis"
	"github.com/yaleh/meta-cc/internal/config"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/mcp/metrics"
	obspkg "github.com/yaleh/meta-cc/internal/mcp/observability"
	pipelinepkg "github.com/yaleh/meta-cc/internal/mcp/pipeline"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
	toolspkg "github.com/yaleh/meta-cc/internal/mcp/tools"
)

// ToolExecutor executes MCP tools for session history analysis.
type ToolExecutor struct {
	AnalysisSvc analysis.AnalysisService
}

// NewToolExecutor creates a new ToolExecutor.
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		AnalysisSvc: analysis.New(),
	}
}

// NewToolPipelineConfig creates a PipelineConfig from args map.
func NewToolPipelineConfig(toolName string, args map[string]interface{}) pipelinepkg.PipelineConfig {
	return pipelinepkg.PipelineConfig{
		JQFilter:            GetStringParam(args, "jq_filter", ".[]"),
		StatsOnly:           GetBoolParam(args, "stats_only", false),
		StatsFirst:          GetBoolParam(args, "stats_first", false),
		OutputFormat:        GetStringParam(args, "output_format", "jsonl"),
		MaxMessageLength:    GetIntParam(args, "max_message_length", 0),
		ContentSummary:      GetBoolParam(args, "content_summary", false),
		PreviewLength:       GetIntParam(args, "preview_length", pipelinepkg.DefaultPreviewLength),
		GroupBySession:      GetBoolParam(args, "group_by_session", false),
		StatsLevel:          GetStringParam(args, "stats_level", "turn"),
		ContextTurns:        GetIntParam(args, "context_turns", 0),
		UseTimestampStats:   pipelinepkg.TimestampStatsTools[toolName],
		ApplyMessageFilters: toolName == "query_session_content",
	}
}

// DetermineScope returns the scope for a tool call.
func DetermineScope(toolName string, args map[string]interface{}) string {
	defaultScope := "project"
	if toolName == "get_session_stats" {
		defaultScope = "session"
	}
	return GetStringParam(args, "scope", defaultScope)
}

// RecordToolSuccess records a successful tool execution.
func RecordToolSuccess(toolName, scope string, start time.Time) {
	elapsed := time.Since(start)
	metrics.RecordToolCall(toolName, scope, "success")
	metrics.RecordToolExecutionDuration(toolName, scope, elapsed)
}

// RecordToolFailure records a failed tool execution.
func RecordToolFailure(toolName, scope string, start time.Time, errorType string) {
	elapsed := time.Since(start)
	metrics.RecordToolCall(toolName, scope, "error")
	metrics.RecordToolExecutionDuration(toolName, scope, elapsed)
	metrics.RecordError(toolName, errorType, metrics.GetErrorSeverity(errorType))
}

// ExecuteSpecialTool handles special tools that don't go through the standard pipeline.
func (e *ToolExecutor) ExecuteSpecialTool(cfg *config.Config, toolName, scope string, args map[string]interface{}, start time.Time) (string, bool, error) {
	handler, ok := specialToolRegistry[toolName]
	if !ok {
		return "", false, nil
	}
	output, err := handler(context.Background(), e, args)
	if err != nil {
		errorType := obspkg.ClassifyError(err)
		RecordToolFailure(toolName, scope, start, errorType)
		return "", true, err
	}
	RecordToolSuccess(toolName, scope, start)
	return output, true, nil
}

// ExecuteToolQuery runs a convenience (query) tool and returns the raw QueryResult.
// Intended for testing and callers that need structured data before serialization.
func (e *ToolExecutor) ExecuteToolQuery(toolName, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	handler, ok := queryHandlerRegistry[toolName]
	if !ok {
		return mcquery.QueryResult{}, fmt.Errorf("unknown tool: %s", toolName)
	}
	return handler(e, scope, args)
}

// ExecuteTool executes a meta-cc command and returns the formatted output.
func (e *ToolExecutor) ExecuteTool(cfg *config.Config, toolName string, args map[string]interface{}) (string, error) {
	scope := DetermineScope(toolName, args)
	start := time.Now()

	if output, handled, err := e.ExecuteSpecialTool(cfg, toolName, scope, args, start); handled {
		return output, err
	}

	if scope != "project" && scope != "session" {
		RecordToolFailure(toolName, scope, start, "validation_error")
		return "", fmt.Errorf("invalid scope %q: must be \"project\" or \"session\"", scope)
	}

	if err := toolspkg.ValidateToolArgs(toolName, args); err != nil {
		RecordToolFailure(toolName, scope, start, "validation_error")
		if strings.Contains(err.Error(), "unknown tool") {
			return "", fmt.Errorf("unknown tool %s in executor: %w", toolName, mcerrors.ErrUnknownTool)
		}
		return "", err
	}

	handler, ok := queryHandlerRegistry[toolName]
	if !ok {
		RecordToolFailure(toolName, scope, start, "validation_error")
		return "", fmt.Errorf("unknown tool %s in executor: %w", toolName, mcerrors.ErrUnknownTool)
	}

	queryResult, err := handler(e, scope, args)
	if err != nil {
		errorType := obspkg.ClassifyError(err)
		slog.Error("tool execution failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", errorType,
		)
		RecordToolFailure(toolName, scope, start, errorType)
		return "", err
	}

	pipeline := NewToolPipelineConfig(toolName, args)
	output, err := pipelinepkg.BuildResponse(cfg, queryResult, args, toolName, pipeline)
	if err != nil {
		return "", err
	}

	slog.Debug("tool execution pipeline completed successfully",
		"tool_name", toolName,
		"output_length", len(output),
	)

	RecordToolSuccess(toolName, scope, start)
	return output, nil
}

// ExecuteQuery is an internal helper for convenience tools.
func (e *ToolExecutor) ExecuteQuery(scope string, jqFilter string, limit int, workingDir string) (mcquery.QueryResult, error) {
	return e.ExecuteQueryWithTimeRange(scope, jqFilter, limit, workingDir, mcquery.ParsedTimeRange{})
}

// ExecuteQueryWithTimeRange is like ExecuteQuery but applies time-range filtering.
func (e *ToolExecutor) ExecuteQueryWithTimeRange(scope string, jqFilter string, limit int, workingDir string, tr mcquery.ParsedTimeRange) (mcquery.QueryResult, error) {
	baseDir, err := mcquery.GetQueryBaseDir(scope, workingDir)
	if err != nil {
		return mcquery.QueryResult{}, fmt.Errorf("failed to get base directory: %w", err)
	}

	executor := mcquery.NewQueryExecutor(baseDir)

	code, err := executor.CompileExpression(jqFilter)
	if err != nil {
		return mcquery.QueryResult{}, fmt.Errorf("invalid jq expression: %w", err)
	}

	files, err := mcquery.GetJSONLFiles(baseDir)
	if err != nil {
		return mcquery.QueryResult{}, fmt.Errorf("failed to list JSONL files: %w", err)
	}

	if len(files) == 0 {
		return mcquery.QueryResult{}, fmt.Errorf("no JSONL files found in %s", baseDir)
	}

	ctx := context.Background()
	result := executor.StreamFilesWithTimeRange(ctx, files, code, limit, tr)
	return result, nil
}

// ParseJSONL parses JSONL string into array of interfaces.
func (e *ToolExecutor) ParseJSONL(jsonlData string) ([]interface{}, error) {
	jsonlData = strings.TrimSpace(jsonlData)

	if jsonlData == "" || jsonlData == "[]" {
		slog.Debug("ParseJSONL: empty input or no results",
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

	slog.Debug("ParseJSONL completed",
		"record_count", len(data),
	)

	return data, nil
}
