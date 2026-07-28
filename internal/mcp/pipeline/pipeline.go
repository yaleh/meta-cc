// Package pipeline provides response-building helpers for the MCP server's tool executor.
// These helpers were extracted from cmd/mcp-server/executor.go to separate response
// construction logic from the orchestration layer.
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaleh/meta-cc/internal/config"
	filterpkg "github.com/yaleh/meta-cc/internal/filter"
	filterspkg "github.com/yaleh/meta-cc/internal/mcp/filters"
	mcquerypkg "github.com/yaleh/meta-cc/internal/mcp/query"
	responsepkg "github.com/yaleh/meta-cc/internal/mcp/response"
	"github.com/yaleh/meta-cc/internal/provider/rawfiles"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
	enginepkg "github.com/yaleh/meta-cc/internal/query/engine"
	querypkg "github.com/yaleh/meta-cc/internal/query/stats"
)

// DefaultPreviewLength is the default rune count for content_preview in content_summary mode.
const DefaultPreviewLength = 100

// PipelineConfig holds configuration for a tool execution pipeline.
type PipelineConfig struct {
	JQFilter                string
	StatsOnly               bool
	StatsFirst              bool
	OutputFormat            string
	MaxMessageLength        int
	ContentSummary          bool
	PreviewLength           int
	GroupBySession          bool
	StatsLevel              string // "turn" (default) or "session"
	ContextTurns            int
	UseTimestampStats       bool // use time-bucketed stats instead of key-count stats
	ApplyMessageFilters     bool // apply message length / content-summary filters
	ExcludeCompactSummaries bool // exclude isCompactSummary=true entries from context_turns
	IncludeSubagents        bool // include subagent JSONL files (default: true)
	Offset                  int  // number of records to skip (pagination offset, default: 0)
	PageSize                int  // max records per page (pagination, default: 0 = no limit)
}

func (c PipelineConfig) requiresMessageFilters() bool {
	return c.MaxMessageLength > 0 || c.ContentSummary
}

// BuildResponse constructs the final response for a query result.
// This is the authoritative implementation; executor.go's BuildResponse was merged here.
func BuildResponse(cfg *config.Config, result mcquerypkg.QueryResult, args map[string]interface{}, toolName string, pc PipelineConfig) (string, error) {
	rawData := result.Entries

	// DIR-041: apply the caller-supplied jq_filter as a post-filter over the
	// tool's own already-produced result set. Each of the four consolidated
	// tools (query_session_content, query_session_signals, query_sessions,
	// query_file_activity) builds its own hard-coded, semantically-fixed jq
	// expression (or, for query_sessions, a Go-native SessionFilter) to
	// produce result.Entries in the first place; pc.JQFilter is a distinct,
	// caller-controlled *second* filtering stage applied on top of that,
	// mirroring the array-input/".[]"-default semantics already promised by
	// the "jq_filter" tool schema description (see internal/mcp/tools/tools.go's
	// StandardToolParameters and internal/query/engine.ApplyJQFilter).
	jqFiltered, jqErr := applyJQPostFilter(rawData, pc.JQFilter)
	if jqErr != nil {
		return "", fmt.Errorf("jq_filter error: %w", jqErr)
	}
	rawData = jqFiltered

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
		providerName := pipelineStringArg(args, "provider", "claude")
		scope := pipelineStringArg(args, "scope", "project")
		workingDir := pipelineStringArg(args, "working_dir", "")

		if providerName == "" || providerName == "claude" {
			// Claude-only queries keep the pre-existing behavior exactly:
			// matched records carry a native Claude uuid, so the direct
			// baseDir JSONL rescan remains correct and backward compatible.
			baseDir, err := mcquerypkg.GetQueryBaseDir(scope, workingDir)
			if err != nil {
				return "", err
			}
			parsedData, err = filterspkg.ExpandContextTurns(parsedData, pc.ContextTurns, baseDir, pc.ExcludeCompactSummaries)
			if err != nil {
				return "", err
			}
		} else {
			// DIR-036: codex/all queries never carry a Claude uuid and are
			// never backed by a Claude project directory, so context must be
			// loaded through the provider/session abstraction instead of
			// rescanning a provider-blind base directory.
			expanded, ctxWarnings, err := expandProviderContext(providerName, workingDir, parsedData, pc.ContextTurns, pc.ExcludeCompactSummaries)
			if err != nil {
				return "", err
			}
			parsedData = expanded
			if len(ctxWarnings) > 0 {
				result.Warnings = append(result.Warnings, ctxWarnings...)
			}
		}
	}

	if pc.GroupBySession && pc.ApplyMessageFilters {
		parsedData = querypkg.GroupBySession(parsedData)
	}

	// Apply pagination (only when PageSize > 0; otherwise pass all data through)
	var paginationMeta *filterpkg.PaginationMetadata
	if pc.PageSize > 0 || pc.Offset > 0 {
		paginated, meta := filterpkg.ApplyPaginationToInterfaces(parsedData, pc.Offset, pc.PageSize)
		parsedData = paginated
		paginationMeta = &meta
	} else {
		// Always compute metadata for the full dataset so every response includes pagination info
		meta := filterpkg.CalculateMetadata(len(parsedData), filterpkg.PaginationConfig{Offset: 0, Limit: 0})
		paginationMeta = &meta
	}

	var output string
	var err error
	if pc.StatsFirst {
		output, err = BuildStatsFirstResponse(cfg, rawData, parsedData, args, toolName, pc.UseTimestampStats, pc.StatsLevel, paginationMeta)
	} else {
		output, err = BuildStandardResponse(cfg, parsedData, args, toolName, paginationMeta)
	}

	if err != nil {
		return "", err
	}

	return InjectWarnings(output, result.Warnings)
}

// expandProviderContext is the DIR-036 provider-neutral context-expansion
// entry point for any provider other than the Claude default ("codex" or
// "all"). It resolves the same Claude+Codex provider registry the original
// query used (internal/provider/rawfiles.NewRegistry), and for each session
// referenced by an already-matched record, reloads that session's full
// canonical, normalized turn stream via providerrecords.BuildForSession —
// the same projection (internal/provider/records.Normalize) the original
// query itself was built from — rather than reaching behind the provider
// abstraction to rescan raw Codex rollout/app-server files directly.
//
// Per the DIR-036 contract, a session that fails to (re)load never causes its
// already-matched records to silently disappear: ExpandContextTurnsCanonical
// retains those matches and returns a warning describing the failure instead.
func expandProviderContext(providerName, workingDir string, parsedData []interface{}, n int, excludeCompactSummaries bool) ([]interface{}, []string, error) {
	projectPath := workingDir
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve working directory for context expansion: %w", err)
		}
		projectPath = cwd
	}
	if abs, err := filepath.Abs(projectPath); err == nil {
		projectPath = abs
	}

	filters, err := rawfiles.ParseProviderFilter(providerName)
	if err != nil {
		return nil, nil, err
	}
	registry := rawfiles.NewRegistry(projectPath)

	loadSession := func(sessionID string) ([]interface{}, error) {
		records, _, err := providerrecords.BuildForSession(context.Background(), registry, filters, sessionID, projectPath)
		if err != nil {
			return nil, err
		}
		out := make([]interface{}, len(records))
		for i, r := range records {
			out[i] = r
		}
		return out, nil
	}

	return filterspkg.ExpandContextTurnsCanonical(parsedData, n, loadSession, excludeCompactSummaries)
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

// applyJQPostFilter applies a caller-supplied jq_filter expression to an
// already-produced result set (entries). ".[]" (the documented default) and
// "" are treated as a no-op identity pass so the zero-value/default
// PipelineConfig behavior is completely unchanged (avoids an unnecessary
// marshal/unmarshal round trip on the hot path where no custom filter was
// requested). Any other expression is compiled and run via
// internal/query/engine.ApplyJQFilter against the full entries array as a
// single JSON value — matching the "Defaults to '.[]'" semantics already
// documented on the jq_filter tool parameter (callers write expressions like
// ".[] | select(...)" to iterate, exactly as with a plain `jq` invocation
// over a JSON array).
func applyJQPostFilter(entries []interface{}, jqExpr string) ([]interface{}, error) {
	if jqExpr == "" || jqExpr == ".[]" {
		return entries, nil
	}

	jsonlData, err := DataToJSONL(entries)
	if err != nil {
		return nil, err
	}

	filteredJSONL, err := enginepkg.ApplyJQFilter(jsonlData, jqExpr)
	if err != nil {
		return nil, err
	}

	filteredJSONL = strings.TrimSpace(filteredJSONL)
	if filteredJSONL == "" {
		return []interface{}{}, nil
	}

	lines := strings.Split(filteredJSONL, "\n")
	out := make([]interface{}, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var v interface{}
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			return nil, fmt.Errorf("jq_filter produced invalid JSON output: %w", err)
		}
		out = append(out, v)
	}
	return out, nil
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
// paginationMeta carries pagination metadata for the response envelope.
func BuildStatsFirstResponse(
	cfg *config.Config,
	rawData []interface{},
	parsedData []interface{},
	args map[string]interface{},
	toolName string,
	useTimestampStats bool,
	statsLevel string,
	paginationMeta *filterpkg.PaginationMetadata,
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
	response, err := responsepkg.AdaptResponse(cfg, parsedData, args, toolName, paginationMeta)
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
// paginationMeta carries pagination metadata for the response envelope.
func BuildStandardResponse(
	cfg *config.Config,
	parsedData []interface{},
	args map[string]interface{},
	toolName string,
	paginationMeta *filterpkg.PaginationMetadata,
) (string, error) {
	response, err := responsepkg.AdaptResponse(cfg, parsedData, args, toolName, paginationMeta)
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
