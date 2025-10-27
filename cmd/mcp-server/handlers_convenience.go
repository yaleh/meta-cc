package main

import (
	"fmt"
	"strings"

	"github.com/yaleh/meta-cc/internal/config"
)

// handlers_convenience.go implements the 10 convenience tools (Layer 1)
// These tools wrap executeQuery() with pre-configured jq expressions
// Phase 27 Stage 27.1: Updated to use executeQuery() instead of handleQuery()

// handleQueryUserMessages implements query_user_messages convenience tool
// Maps to Query 1 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQueryUserMessages(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	pattern := getStringParam(args, "pattern", "")
	contentType := getStringParam(args, "content_type", "string")
	limit := getIntParam(args, "limit", 0)

	// Build jq filter based on content type
	var jqFilter string
	if contentType == "string" {
		jqFilter = `select(.type == "user" and (.message.content | type == "string"))`
	} else {
		jqFilter = `select(.type == "user" and (.message.content | type == "array"))`
	}

	// Add pattern filter if provided
	if pattern != "" {
		escapedPattern := escapeJQ(pattern)
		jqFilter = fmt.Sprintf(`%s | select(.message.content | test("%s"))`, jqFilter, escapedPattern)
	}

	// Call executeQuery directly
	return e.executeQuery(scope, jqFilter, limit)
}

// handleQueryTools implements query_tools convenience tool
// Maps to Query 2 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQueryTools(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	toolName := getStringParam(args, "tool_name", "")
	limit := getIntParam(args, "limit", 0)

	// Base filter for all tool_use blocks
	jqFilter := `select(.type == "assistant") | select(.message.content[] | .type == "tool_use")`

	// Add tool name filter if provided
	if toolName != "" {
		escapedTool := escapeJQ(toolName)
		jqFilter = fmt.Sprintf(`%s | select(.message.content[] | select(.type == "tool_use" and .name == "%s"))`, jqFilter, escapedTool)
	}

	return e.executeQuery(scope, jqFilter, limit)
}

// handleQueryToolErrors implements query_tool_errors convenience tool
// Maps to Query 3 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQueryToolErrors(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	limit := getIntParam(args, "limit", 0)

	// Fixed jq filter for tool errors
	jqFilter := `select(.type == "user" and (.message.content | type == "array")) | ` +
		`select(.message.content[] | select(.type == "tool_result" and .is_error == true))`

	return e.executeQuery(scope, jqFilter, limit)
}

// handleQueryTokenUsage implements query_token_usage convenience tool
// Maps to Query 4 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQueryTokenUsage(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	limit := getIntParam(args, "limit", 0)

	// Filter for assistant messages with usage information
	jqFilter := `select(.type == "assistant" and has("message")) | select(.message | has("usage"))`

	return e.executeQuery(scope, jqFilter, limit)
}

// handleQueryConversationFlow implements query_conversation_flow convenience tool
// Maps to Query 5 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQueryConversationFlow(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	limit := getIntParam(args, "limit", 0)

	// Filter for user and assistant messages only
	jqFilter := `select(.type == "user" or .type == "assistant")`

	// Note: jq_transform was removed in Phase 27 - transform parameter is ignored
	// Users should use jq_filter for transformations instead

	return e.executeQuery(scope, jqFilter, limit)
}

// handleQuerySystemErrors implements query_system_errors convenience tool
// Maps to Query 6 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQuerySystemErrors(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	limit := getIntParam(args, "limit", 0)

	// Filter for system API errors
	jqFilter := `select(.type == "system" and .subtype == "api_error")`

	return e.executeQuery(scope, jqFilter, limit)
}

// handleQueryFileSnapshots implements query_file_snapshots convenience tool
// Maps to Query 7 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQueryFileSnapshots(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	limit := getIntParam(args, "limit", 0)

	// Filter for file history snapshots with messageId
	jqFilter := `select(.type == "file-history-snapshot" and has("messageId"))`

	return e.executeQuery(scope, jqFilter, limit)
}

// handleQueryTimestamps implements query_timestamps convenience tool
// Maps to Query 8 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQueryTimestamps(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	limit := getIntParam(args, "limit", 0)

	// Filter for entries with timestamp
	jqFilter := `select(.timestamp != null)`

	return e.executeQuery(scope, jqFilter, limit)
}

// handleQuerySummaries implements query_summaries convenience tool
// Maps to Query 9 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQuerySummaries(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	keyword := getStringParam(args, "keyword", "")
	limit := getIntParam(args, "limit", 0)

	// Base filter for summary entries
	jqFilter := `select(.type == "summary")`

	// Add keyword filter if provided
	if keyword != "" {
		escapedKeyword := escapeJQ(keyword)
		jqFilter = fmt.Sprintf(`%s | select(.summary | test("%s"; "i"))`, jqFilter, escapedKeyword)
	}

	return e.executeQuery(scope, jqFilter, limit)
}

// handleQueryToolBlocks implements query_tool_blocks convenience tool
// Maps to Query 10 from frequent-jsonl-queries.md
func (e *ToolExecutor) handleQueryToolBlocks(cfg *config.Config, scope string, args map[string]interface{}) ([]interface{}, error) {
	blockType := getStringParam(args, "block_type", "tool_use")
	limit := getIntParam(args, "limit", 0)

	// Validate block_type
	if blockType != "tool_use" && blockType != "tool_result" {
		return nil, fmt.Errorf("invalid block_type: %s (must be 'tool_use' or 'tool_result')", blockType)
	}

	var jqFilter string
	if blockType == "tool_use" {
		// Extract tool_use blocks from assistant messages
		jqFilter = `select(.type == "assistant") | .message.content[] | select(.type == "tool_use")`
	} else {
		// Extract tool_result blocks from user messages
		jqFilter = `select(.type == "user" and (.message.content | type == "array")) | .message.content[] | select(.type == "tool_result")`
	}

	return e.executeQuery(scope, jqFilter, limit)
}

// escapeJQ escapes special characters in strings for jq expressions
func escapeJQ(s string) string {
	// Escape backslashes first
	s = strings.ReplaceAll(s, `\`, `\\`)
	// Escape double quotes
	s = strings.ReplaceAll(s, `"`, `\"`)
	// Escape regex special chars for test() function
	// Note: This is basic escaping - for complex patterns, users should use query_raw
	return s
}
