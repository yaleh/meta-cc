// Package pipeline provides response-building helpers for the MCP server's tool executor.
// These helpers were extracted from cmd/mcp-server/executor.go to separate response
// construction logic from the orchestration layer.
package pipeline

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/yaleh/meta-cc/internal/config"
	filterspkg "github.com/yaleh/meta-cc/internal/mcp/filters"
	mcquerypkg "github.com/yaleh/meta-cc/internal/mcp/query"
	responsepkg "github.com/yaleh/meta-cc/internal/mcp/response"
	querypkg "github.com/yaleh/meta-cc/internal/query/stats"
)

// DefaultPreviewLength is the default rune count for content_preview in content_summary mode.
const DefaultPreviewLength = 100

// PipelineConfig holds configuration for a tool execution pipeline.
type PipelineConfig struct {
	JQFilter            string
	StatsOnly           bool
	StatsFirst          bool
	OutputFormat        string
	MaxMessageLength    int
	ContentSummary      bool
	PreviewLength       int
	GroupBySession      bool
	StatsLevel          string // "turn" (default) or "session"
	ContextTurns        int
	UseTimestampStats   bool // use time-bucketed stats instead of key-count stats
	ApplyMessageFilters bool // apply message length / content-summary filters
}

func (c PipelineConfig) requiresMessageFilters() bool {
	return c.MaxMessageLength > 0 || c.ContentSummary
}

// BuildResponse constructs the final response for a query result.
// This is the authoritative implementation; executor.go's BuildResponse was merged here.
func BuildResponse(cfg *config.Config, result mcquerypkg.QueryResult, args map[string]interface{}, toolName string, pc PipelineConfig) (string, error) {
	rawData := result.Entries

	if pc.StatsLevel != "" && pc.StatsLevel != "turn" && pc.StatsLevel != "session" {
		return "", fmt.Errorf("invalid stats_level: must be 'turn' or 'session'")
	}

	if pc.GroupBySession && pc.StatsOnly {
		return "", fmt.Errorf("group_by_session and stats_only are mutually exclusive")
	}

	if pc.StatsOnly && !result.BypassStats {
		output, err := BuildStatsOnlyResponse(rawData, pc.UseTimestampStats, pc.StatsLevel)
		if err != nil {
			return "", err
		}
		return InjectWarnings(output, result.Warnings)
	}

	parsedData := rawData
	if pc.ApplyMessageFilters && pc.requiresMessageFilters() {
		parsedData = filterspkg.ApplyMessageFiltersToData(rawData, pc.MaxMessageLength, pc.ContentSummary, pc.PreviewLength)
	}

	if pc.ContextTurns > 0 && pc.ApplyMessageFilters &&
		pipelineStringArg(args, "content_type") != "array" {
		baseDir, err := mcquerypkg.GetQueryBaseDir(
			pipelineStringArg(args, "scope", "project"),
			pipelineStringArg(args, "working_dir", ""),
		)
		if err != nil {
			return "", err
		}
		parsedData, err = filterspkg.ExpandContextTurns(parsedData, pc.ContextTurns, baseDir)
		if err != nil {
			return "", err
		}
	}

	if pc.GroupBySession && pc.ApplyMessageFilters {
		parsedData = querypkg.GroupBySession(parsedData)
	}

	var output string
	var err error
	if pc.StatsFirst {
		output, err = BuildStatsFirstResponse(cfg, rawData, parsedData, args, toolName, pc.UseTimestampStats, pc.StatsLevel)
	} else {
		output, err = BuildStandardResponse(cfg, parsedData, args, toolName)
	}

	if err != nil {
		return "", err
	}

	return InjectWarnings(output, result.Warnings)
}

// pipelineStringArg extracts a string value from args map with an optional default.
func pipelineStringArg(args map[string]interface{}, key string, defaultVals ...string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	if len(defaultVals) > 0 {
		return defaultVals[0]
	}
	return ""
}

// TimestampStatsTools is the set of tool names that should use GenerateTimestampStats
// instead of GenerateStats when producing stats_only or stats_first output.
// These tools return records that lack a tool/ToolName field but have timestamp data,
// so time-bucketed stats are more meaningful than the meaningless "unknown" key.
var TimestampStatsTools = map[string]bool{
	// New consolidated tools that use time-bucketed stats:
	// query_session_content covers: user messages, conversation flow, summaries
	// query_session_signals covers: timestamps (when type=timestamps)
	"query_session_content": true,
	"query_session_signals": true,

	// query_summaries was removed in Phase D (it used select(.type=="summary") which never
	// matched any Claude Code records — that type does not exist in the JSONL schema).
	// The correct replacement is: query_session_content(role=assistant, contains="## Summary")
	// Root cause documented in: docs/tasks/fix-query-summaries-root-cause.md
	"query_summaries": true,
}

// InjectWarnings adds a "warnings" field to a JSON response string.
// If the output is valid JSON object, it adds the field. Otherwise returns as-is.
func InjectWarnings(output string, warnings []string) (string, error) {
	if warnings == nil {
		warnings = []string{}
	}

	// Try to parse as JSON object
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		// Not a JSON object (e.g., stats_only plain text) — skip injection
		return output, nil
	}

	parsed["warnings"] = warnings

	result, err := json.Marshal(parsed)
	if err != nil {
		return "", fmt.Errorf("failed to re-serialize response with warnings: %w", err)
	}
	return string(result), nil
}

// DataToJSONL converts array of interfaces to JSONL string.
func DataToJSONL(data []interface{}) (string, error) {
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

// BuildStatsOnlyResponse generates a stats-only response for the given data.
// statsLevel may be "turn" (default) or "session".
// useTimestampStats selects time-bucketed stats; when false, key-count stats are used.
func BuildStatsOnlyResponse(parsedData []interface{}, useTimestampStats bool, statsLevel string) (string, error) {
	jsonlData, err := DataToJSONL(parsedData)
	if err != nil {
		slog.Error("DataToJSONL conversion failed (stats_only)",
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", err
	}

	var output string
	if statsLevel == "session" && useTimestampStats {
		output, err = querypkg.GenerateSessionStats(jsonlData)
	} else if useTimestampStats {
		output, err = querypkg.GenerateTimestampStats(jsonlData)
	} else {
		output, err = querypkg.GenerateStats(jsonlData)
	}
	if err != nil {
		slog.Error("stats generation failed",
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", err
	}

	return output, nil
}

// BuildStatsFirstResponse generates a stats-first response: stats header followed by
// serialized detail data.
// useTimestampStats selects time-bucketed stats; when false, key-count stats are used.
// toolName is passed through to AdaptResponse for output formatting only.
func BuildStatsFirstResponse(
	cfg *config.Config,
	rawData []interface{},
	parsedData []interface{},
	args map[string]interface{},
	toolName string,
	useTimestampStats bool,
	statsLevel string,
) (string, error) {
	// Use rawData for stats (sessionId field preserved, not renamed by content_summary)
	jsonlData, err := DataToJSONL(rawData)
	if err != nil {
		slog.Error("DataToJSONL conversion failed (stats_first)",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", err
	}

	var stats string
	if statsLevel == "session" && useTimestampStats {
		stats, _ = querypkg.GenerateSessionStats(jsonlData)
	} else if useTimestampStats {
		stats, _ = querypkg.GenerateTimestampStats(jsonlData)
	} else {
		stats, _ = querypkg.GenerateStats(jsonlData)
	}

	// Use parsedData for detail rendering (may have content_summary applied)
	response, err := responsepkg.AdaptResponse(cfg, parsedData, args, toolName)
	if err != nil {
		slog.Error("response adaptation failed (stats_first)",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", err
	}

	serialized, err := responsepkg.SerializeResponse(response)
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

// BuildStandardResponse generates a standard (non-stats) response for the given data.
func BuildStandardResponse(
	cfg *config.Config,
	parsedData []interface{},
	args map[string]interface{},
	toolName string,
) (string, error) {
	response, err := responsepkg.AdaptResponse(cfg, parsedData, args, toolName)
	if err != nil {
		slog.Error("response adaptation failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", fmt.Errorf("response adaptation error for tool %s: %w", toolName, err)
	}

	output, err := responsepkg.SerializeResponse(response)
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
