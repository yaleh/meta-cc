package executor

import (
	"fmt"
	"strings"

	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
)

// handlers.go implements the 10 convenience tools (Layer 1).
// Each tool is registered via init() so executor.go needs no switch statement.
// Exported Handle* methods are kept for backward compatibility with cmd/mcp-server.

// Old tool registrations removed in Phase D.
// The private handler functions (handleQueryUserMessages, handleQueryTools, etc.)
// are kept because consolidated_handlers.go delegates to them.

// ─── Private implementations ──────────────────────────────────────────────────

func handleQueryUserMessages(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	pattern := GetStringParam(args, "pattern", "")
	contentType := GetStringParam(args, "content_type", "string")
	limit := GetIntParam(args, "limit", 0)
	minContentLength := GetIntParam(args, "min_content_length", 0)
	maxContentLength := GetIntParam(args, "max_content_length", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	sinceStr := GetStringParam(args, "since", "")
	untilStr := GetStringParam(args, "until", "")
	excludeSystem := GetBoolParam(args, "exclude_system_messages", false)
	excludeCompact := GetBoolParam(args, "exclude_compact_summaries", true)
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	tr, err := mcquery.ParseTimeRange(sinceStr, untilStr)
	if err != nil {
		return mcquery.QueryResult{}, err
	}

	if contentType != "string" && (minContentLength > 0 || maxContentLength > 0) {
		return mcquery.QueryResult{}, fmt.Errorf("content length filtering (min_content_length/max_content_length) only applies to string content type, not %q", contentType)
	}

	var jqFilter string
	if contentType == "string" {
		jqFilter = `select(.type == "user" and (.message.content | type == "string"))`
	} else {
		jqFilter = `select(.type == "user" and (.message.content | type == "array"))`
	}

	if pattern != "" {
		escapedPattern := EscapeJQ(pattern)
		jqFilter = fmt.Sprintf(`%s | select(.message.content | test("%s"))`, jqFilter, escapedPattern)
	}

	if minContentLength > 0 {
		jqFilter = fmt.Sprintf(`%s | select(.message.content | length >= %d)`, jqFilter, minContentLength)
	}
	if maxContentLength > 0 {
		jqFilter = fmt.Sprintf(`%s | select(.message.content | length <= %d)`, jqFilter, maxContentLength)
	}

	if excludeSystem && (contentType == "string" || contentType == "") {
		jqFilter += ` | select(.message.content | (startswith("<local-command-caveat>") or startswith("<command-name>") or startswith("<local-command-stdout>") or startswith("<task-notification>")) | not)`
	}

	if excludeCompact {
		jqFilter += ` | select(.isCompactSummary != true)`
	}

	return e.ExecuteQueryWithTimeRangeForProvider(providerName, scope, jqFilter, limit, workingDir, tr, includeSubagents)
}

func handleQueryTools(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	toolName := GetStringParam(args, "tool", "")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "assistant") | select(.message.content[] | .type == "tool_use")`

	if toolName != "" {
		escapedTool := EscapeJQ(toolName)
		jqFilter = fmt.Sprintf(`%s | select(.message.content[] | select(.type == "tool_use" and .name == "%s"))`, jqFilter, escapedTool)
	}

	return e.ExecuteQueryForProvider(providerName, scope, jqFilter, limit, workingDir, includeSubagents)
}

func handleQueryToolErrors(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "user" and (.message.content | type == "array")) | ` +
		`select(.message.content[] | select(.type == "tool_result" and .is_error == true))`

	return e.ExecuteQueryForProvider(providerName, scope, jqFilter, limit, workingDir, includeSubagents)
}

func handleQueryTokenUsage(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "assistant" and has("message")) | select(.message | has("usage"))`

	return e.ExecuteQueryForProvider(providerName, scope, jqFilter, limit, workingDir, includeSubagents)
}

func handleQueryConversationFlow(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	sinceStr := GetStringParam(args, "since", "")
	untilStr := GetStringParam(args, "until", "")
	excludeCompact := GetBoolParam(args, "exclude_compact_summaries", true)
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	tr, err := mcquery.ParseTimeRange(sinceStr, untilStr)
	if err != nil {
		return mcquery.QueryResult{}, err
	}

	jqFilter := `select(.type == "user" or .type == "assistant")`

	if excludeCompact {
		jqFilter += ` | select(.isCompactSummary != true)`
	}

	return e.ExecuteQueryWithTimeRangeForProvider(providerName, scope, jqFilter, limit, workingDir, tr, includeSubagents)
}

func handleQuerySystemErrors(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "system" and .subtype == "api_error")`

	return e.ExecuteQueryForProvider(providerName, scope, jqFilter, limit, workingDir, includeSubagents)
}

func handleQueryFileSnapshots(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "file-history-snapshot" and has("messageId"))`

	return e.ExecuteQueryForProvider(providerName, scope, jqFilter, limit, workingDir, includeSubagents)
}

func handleQueryTimestamps(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	sinceStr := GetStringParam(args, "since", "")
	untilStr := GetStringParam(args, "until", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	tr, err := mcquery.ParseTimeRange(sinceStr, untilStr)
	if err != nil {
		return mcquery.QueryResult{}, err
	}

	jqFilter := `select(.timestamp != null)`

	return e.ExecuteQueryWithTimeRangeForProvider(providerName, scope, jqFilter, limit, workingDir, tr, includeSubagents)
}

func handleQueryToolBlocks(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	blockType := GetStringParam(args, "block_type", "tool_use")
	toolName := GetStringParam(args, "tool_name", "")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	if blockType != "tool_use" && blockType != "tool_result" {
		return mcquery.QueryResult{}, fmt.Errorf("invalid block_type: %s (must be 'tool_use' or 'tool_result')", blockType)
	}

	var jqFilter string
	if blockType == "tool_use" {
		jqFilter = `select(.type == "assistant") | . as $rec | .message.content[] | select(.type == "tool_use")`
		if toolName != "" {
			jqFilter = fmt.Sprintf(`%s | select(.name | test("%s"))`, jqFilter, EscapeJQ(toolName))
		}
		jqFilter += ` | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .`
	} else {
		jqFilter = `select(.type == "user" and (.message.content | type == "array")) | . as $rec | .message.content[] | select(.type == "tool_result") | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .`
	}

	return e.ExecuteQueryForProvider(providerName, scope, jqFilter, limit, workingDir, includeSubagents)
}

// EscapeJQ escapes special characters in strings for jq expressions.
func EscapeJQ(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
