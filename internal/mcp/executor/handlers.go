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

	sessionID := GetStringParam(args, "session_id", "")
	return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, tr, includeSubagents)
}

func handleQueryTools(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	toolName := GetStringParam(args, "tool", "")
	status := GetStringParam(args, "status", "")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)
	sessionID := GetStringParam(args, "session_id", "")

	if status != "" && status != "error" && status != "success" {
		return mcquery.QueryResult{}, fmt.Errorf("invalid status %q: must be 'error' or 'success' (or omit status to skip status filtering)", status)
	}

	jqFilter := `select(.type == "assistant") | select(.message.content[] | .type == "tool_use")`

	if toolName != "" {
		escapedTool := EscapeJQ(toolName)
		jqFilter = fmt.Sprintf(`%s | select(.message.content[] | select(.type == "tool_use" and .name == "%s"))`, jqFilter, escapedTool)
	}

	if status == "" {
		return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, mcquery.ParsedTimeRange{}, includeSubagents)
	}

	// status filtering needs the outcome (.is_error) of each tool_use, but
	// that field lives on a *separate* record: the tool_result block in the
	// next user-role JSONL entry, correlated by tool_use_id (see
	// handleQueryToolErrors above, which reads .is_error directly off
	// tool_result blocks for a related but different query shape). The jq
	// pipeline here runs one record at a time with no cross-record join, so
	// fetch both tool_use and tool_result records in a single unbounded pass
	// (limit=0) and correlate them here; the caller's limit is applied after
	// filtering below, not before, so it isn't applied to the wrong (joined,
	// unfiltered) record set.
	joinFilter := fmt.Sprintf(
		`(%s), (select(.type == "user" and (.message.content | type == "array")) | select(.message.content[] | .type == "tool_result"))`,
		jqFilter,
	)

	joined, err := e.dispatchProviderQuery(providerName, scope, joinFilter, 0, workingDir, sessionID, mcquery.ParsedTimeRange{}, includeSubagents)
	if err != nil {
		return mcquery.QueryResult{}, err
	}

	isErrorByToolUseID := make(map[string]bool)
	for _, entry := range joined.Entries {
		rec, ok := entry.(map[string]interface{})
		if !ok || rec["type"] != "user" {
			continue
		}
		message, _ := rec["message"].(map[string]interface{})
		content, _ := message["content"].([]interface{})
		for _, block := range content {
			bm, ok := block.(map[string]interface{})
			if !ok || bm["type"] != "tool_result" {
				continue
			}
			id, _ := bm["tool_use_id"].(string)
			if id == "" {
				continue
			}
			isErr, _ := bm["is_error"].(bool)
			isErrorByToolUseID[id] = isErr
		}
	}

	wantError := status == "error"
	var filtered []interface{}
	for _, entry := range joined.Entries {
		rec, ok := entry.(map[string]interface{})
		if !ok || rec["type"] != "assistant" {
			continue
		}
		message, _ := rec["message"].(map[string]interface{})
		content, _ := message["content"].([]interface{})
		matched := false
		for _, block := range content {
			bm, ok := block.(map[string]interface{})
			if !ok || bm["type"] != "tool_use" {
				continue
			}
			id, _ := bm["id"].(string)
			isErr, found := isErrorByToolUseID[id]
			if !found {
				continue // no matching tool_result observed yet: status unknown, exclude
			}
			if isErr == wantError {
				matched = true
				break
			}
		}
		if matched {
			filtered = append(filtered, entry)
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return mcquery.QueryResult{Entries: filtered, Warnings: joined.Warnings}, nil
}

func handleQueryToolErrors(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "user" and (.message.content | type == "array")) | ` +
		`select(.message.content[] | select(.type == "tool_result" and .is_error == true))`

	sessionID := GetStringParam(args, "session_id", "")
	return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, mcquery.ParsedTimeRange{}, includeSubagents)
}

func handleQueryTokenUsage(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "assistant" and has("message")) | select(.message | has("usage"))`

	sessionID := GetStringParam(args, "session_id", "")
	return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, mcquery.ParsedTimeRange{}, includeSubagents)
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

	sessionID := GetStringParam(args, "session_id", "")
	return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, tr, includeSubagents)
}

func handleQuerySystemErrors(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "system" and .subtype == "api_error")`

	sessionID := GetStringParam(args, "session_id", "")
	return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, mcquery.ParsedTimeRange{}, includeSubagents)
}

func handleQueryFileSnapshots(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	providerName := GetStringParam(args, "provider", "claude")
	limit := GetIntParam(args, "limit", 0)
	workingDir := GetStringParam(args, "working_dir", "")
	includeSubagents := GetBoolParam(args, "include_subagents", true)

	jqFilter := `select(.type == "file-history-snapshot" and has("messageId"))`

	sessionID := GetStringParam(args, "session_id", "")
	return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, mcquery.ParsedTimeRange{}, includeSubagents)
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

	sessionID := GetStringParam(args, "session_id", "")
	return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, tr, includeSubagents)
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

	sessionID := GetStringParam(args, "session_id", "")
	return e.dispatchProviderQuery(providerName, scope, jqFilter, limit, workingDir, sessionID, mcquery.ParsedTimeRange{}, includeSubagents)
}

// EscapeJQ escapes special characters in strings for jq expressions.
func EscapeJQ(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
