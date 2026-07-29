package executor

import (
	"fmt"

	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
)

// consolidated_handlers.go implements the 3 new consolidated query tools:
//   - query_session_content  (replaces 4 old tools)
//   - query_session_signals  (replaces 5 old tools)
//   - query_file_activity    (replaces 1 old tool)
//
// Each handler switches on role/type and delegates to the existing private
// handler functions — no new filter logic is added here.

func init() {
	registerQueryHandler("query_session_content", func(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
		return handleQuerySessionContent(e, scope, args)
	})
	registerQueryHandler("query_session_signals", func(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
		return handleQuerySessionSignals(e, scope, args)
	})
	registerQueryHandler("query_file_activity", func(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
		return handleQueryFileActivity(e, scope, args)
	})
}

// handleQuerySessionContent routes by 'role' to existing handlers.
//
// The 'contains' literal-substring filter (DIR-047 escaping: QuoteMeta +
// EscapeJQ via the shared containsClause helper) is honored by EVERY role —
// user, assistant, tool, and all (DIR-062; before that fix it was silently
// dropped for all roles except assistant).
//
//	role=user       → handleQueryUserMessages (requires 'pattern' or defaults to ".*")
//	role=assistant  → query assistant messages, optionally filtered by 'contains'
//	role=tool       → handleQueryToolBlocks
//	role=all        → handleQueryConversationFlow
func handleQuerySessionContent(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	role := GetStringParam(args, "role", "")

	switch role {
	case "user":
		// Merge: if pattern not set, default to ".*" (match all string content)
		// Do not set pattern for array content_type since test() can't operate on arrays.
		delegateArgs := copyArgs(args)
		contentType := GetStringParam(delegateArgs, "content_type", "string")
		if GetStringParam(delegateArgs, "pattern", "") == "" && contentType != "array" {
			delegateArgs["pattern"] = ".*"
		}
		return handleQueryUserMessages(e, scope, delegateArgs)

	case "assistant":
		// Query assistant messages; optionally filter by 'contains'
		contains := GetStringParam(args, "contains", "")
		providerName := GetStringParam(args, "provider", "claude")
		limit := GetIntParam(args, "limit", 0)
		workingDir := GetStringParam(args, "working_dir", "")
		includeSubagents := GetBoolParam(args, "include_subagents", true)
		sessionID := GetStringParam(args, "session_id", "")

		jqFilter := `select(.type == "assistant")`
		if contains != "" {
			// Route through the shared containsClause helper (handlers.go),
			// which applies the DIR-047 regexp.QuoteMeta + EscapeJQ + jq
			// test(...; "i") escaping path — the same path every other role
			// uses since DIR-062, so the escaping cannot diverge per branch.
			// The `// empty` guard skips records without a content field
			// rather than producing a jq error (this fixed the query_summaries
			// null return when content is null or absent).
			jqFilter = jqFilter + " | " + containsClause(contains, ".message.content // empty")
		}
		return e.dispatchIndexedContent(providerName, scope, contains, jqFilter, limit, workingDir, sessionID, mcquery.ParsedTimeRange{}, includeSubagents)

	case "tool":
		// Delegate to handleQueryToolBlocks; block_type may be passed through
		return handleQueryToolBlocks(e, scope, args)

	case "all":
		// Delegate to handleQueryConversationFlow (user + assistant flow)
		return handleQueryConversationFlow(e, scope, args)

	default:
		return mcquery.QueryResult{}, fmt.Errorf("invalid role %q: must be one of 'user', 'assistant', 'tool', or 'all'", role)
	}
}

// handleQuerySessionSignals routes by 'type' to existing handlers.
//
//	type=errors        → handleQueryToolErrors
//	type=tokens        → handleQueryTokenUsage
//	type=system_errors → handleQuerySystemErrors
//	type=timestamps    → handleQueryTimestamps
//	type=tool_stats    → handleQueryTools
func handleQuerySessionSignals(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	signalType := GetStringParam(args, "type", "")

	switch signalType {
	case "errors":
		return handleQueryToolErrors(e, scope, args)
	case "tokens":
		return handleQueryTokenUsage(e, scope, args)
	case "system_errors":
		return handleQuerySystemErrors(e, scope, args)
	case "timestamps":
		return handleQueryTimestamps(e, scope, args)
	case "tool_stats":
		return handleQueryTools(e, scope, args)
	default:
		return mcquery.QueryResult{}, fmt.Errorf("invalid type %q: must be one of 'errors', 'tokens', 'system_errors', 'timestamps', or 'tool_stats'", signalType)
	}
}

// handleQueryFileActivity routes by 'type' to existing handlers.
//
//	type=snapshots → handleQueryFileSnapshots
func handleQueryFileActivity(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error) {
	activityType := GetStringParam(args, "type", "")

	switch activityType {
	case "snapshots":
		return handleQueryFileSnapshots(e, scope, args)
	default:
		return mcquery.QueryResult{}, fmt.Errorf("invalid type %q: must be 'snapshots'", activityType)
	}
}

// copyArgs makes a shallow copy of the args map so we can safely mutate defaults.
func copyArgs(args map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(args))
	for k, v := range args {
		out[k] = v
	}
	return out
}
