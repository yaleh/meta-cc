package main

import (
	"fmt"

	querypkg "github.com/yaleh/meta-cc/internal/query"
)

// This file contains adapter functions that demonstrate how legacy MCP tools
// can be mapped to the unified query interface. These adapters are NOT currently
// used in production (legacy tools use their existing implementations), but serve
// as migration examples for future refactoring.
//
// Migration strategy: Each legacy tool can be gradually migrated by:
// 1. Creating an adapter function that maps legacy params to QueryParams
// 2. Testing adapter output matches legacy implementation
// 3. Switching tool executor to use adapter
// 4. Deprecating legacy implementation after 2-3 versions

// adaptQueryTools converts query_tools parameters to unified QueryParams
// Example mapping: query_tools(tool="Read", status="error") →
//
//	query(resource="tools", filter={tool_name:"Read", tool_status:"error"})
func adaptQueryTools(args map[string]interface{}) querypkg.QueryParams {
	params := querypkg.QueryParams{
		Resource: "tools",
		Filter: querypkg.FilterSpec{
			ToolName:   getStringParam(args, "tool", ""),
			ToolStatus: getStringParam(args, "status", ""),
		},
		Output: querypkg.OutputSpec{
			Limit: getIntParam(args, "limit", 0),
		},
	}

	return params
}

// adaptQueryUserMessages converts query_user_messages parameters to QueryParams
// Example mapping: query_user_messages(pattern="error") →
//
//	query(resource="messages", filter={role:"user", content_match:"error"})
func adaptQueryUserMessages(args map[string]interface{}) querypkg.QueryParams {
	params := querypkg.QueryParams{
		Resource: "messages",
		Filter: querypkg.FilterSpec{
			Role:         "user",
			ContentMatch: getStringParam(args, "pattern", ""),
		},
		Output: querypkg.OutputSpec{
			Limit: getIntParam(args, "limit", 0),
		},
	}

	return params
}

// adaptGetSessionStats converts get_session_stats parameters to QueryParams
// Example mapping: get_session_stats() →
//
//	query(resource="entries", aggregate={function:"count"})
//	Note: Session stats require custom aggregation beyond simple count
func adaptGetSessionStats(args map[string]interface{}) querypkg.QueryParams {
	params := querypkg.QueryParams{
		Resource: "entries",
		// Session stats is complex aggregation over multiple resource types
		// Current unified query doesn't support this level of complexity yet
		// Would need multi-stage queries or custom aggregate functions
	}

	return params
}

// Legacy tool name to adapter mapping (for documentation/future use)
var legacyToolAdapters = map[string]func(map[string]interface{}) querypkg.QueryParams{
	"query_tools":         adaptQueryTools,
	"query_user_messages": adaptQueryUserMessages,
	"get_session_stats":   adaptGetSessionStats,
	// Additional adapters can be added here:
	// "query_tool_sequences": adaptQueryToolSequences,
	// "query_file_access": adaptQueryFileAccess,
	// "query_project_state": adaptQueryProjectState,
	// "query_successful_prompts": adaptQuerySuccessfulPrompts,
}

// canAdaptToUnifiedQuery checks if a legacy tool can be adapted to unified query
// This is useful for progressive migration and testing
func canAdaptToUnifiedQuery(toolName string) bool {
	_, exists := legacyToolAdapters[toolName]
	return exists
}

// adaptLegacyTool adapts a legacy tool call to unified query parameters
// Returns error if tool cannot be adapted
func adaptLegacyTool(toolName string, args map[string]interface{}) (querypkg.QueryParams, error) {
	adapter, exists := legacyToolAdapters[toolName]
	if !exists {
		return querypkg.QueryParams{}, fmt.Errorf("no adapter for tool %s", toolName)
	}

	return adapter(args), nil
}
