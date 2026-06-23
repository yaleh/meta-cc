package executor

import (
	"context"
	"testing"

	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
)

func TestSpecialToolRegistry_AnalysisHandlers(t *testing.T) {
	analysisTool := []string{
		"analyze_bugs",
		"analyze_errors",
		"quality_scan",
		"get_work_patterns",
		"get_timeline",
		"get_tech_debt",
	}
	for _, tool := range analysisTool {
		if _, ok := specialToolRegistry[tool]; !ok {
			t.Errorf("analysis tool %q not registered in specialToolRegistry", tool)
		}
	}
}

func TestSpecialToolRegistry_QueryHandlers(t *testing.T) {
	queryTools := []string{
		"cleanup_temp_files",
		"get_session_directory",
		"inspect_session_files",
		"execute_stage2_query",
		"get_session_metadata",
	}
	for _, tool := range queryTools {
		if _, ok := specialToolRegistry[tool]; !ok {
			t.Errorf("query tool %q not registered in specialToolRegistry", tool)
		}
	}
}

func TestSpecialToolRegistry_UnknownTool(t *testing.T) {
	_, ok := specialToolRegistry["nonexistent_tool"]
	if ok {
		t.Error("expected nonexistent_tool to not be registered")
	}
}

func TestRegisterHandler_AddsToRegistry(t *testing.T) {
	const testTool = "test_tool_registry_unit"
	called := false
	registerHandler(testTool, func(_ context.Context, _ *ToolExecutor, _ map[string]interface{}) (string, error) {
		called = true
		return "ok", nil
	})
	defer delete(specialToolRegistry, testTool)

	h, ok := specialToolRegistry[testTool]
	if !ok {
		t.Fatal("handler not registered")
	}
	out, err := h(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Errorf("expected 'ok', got %q", out)
	}
	if !called {
		t.Error("handler was not called")
	}
}

// ─── QueryHandlerRegistry ─────────────────────────────────────────────────────

func TestQueryHandlerRegistry_AllConsolidatedTools(t *testing.T) {
	consolidatedTools := []string{
		"query_session_content",
		"query_session_signals",
		"query_file_activity",
	}
	for _, tool := range consolidatedTools {
		if _, ok := queryHandlerRegistry[tool]; !ok {
			t.Errorf("consolidated query tool %q not registered in queryHandlerRegistry", tool)
		}
	}
}

func TestQueryHandlerRegistry_OldToolsRemoved(t *testing.T) {
	removedTools := []string{
		"query_user_messages",
		"query_tools",
		"query_tool_errors",
		"query_token_usage",
		"query_conversation_flow",
		"query_system_errors",
		"query_file_snapshots",
		"query_timestamps",
		"query_summaries",
		"query_tool_blocks",
	}
	for _, tool := range removedTools {
		if _, ok := queryHandlerRegistry[tool]; ok {
			t.Errorf("old query tool %q should have been removed from queryHandlerRegistry in Phase D", tool)
		}
	}
}

func TestQueryHandlerRegistry_UnknownTool(t *testing.T) {
	_, ok := queryHandlerRegistry["nonexistent_query_tool"]
	if ok {
		t.Error("expected nonexistent_query_tool to not be registered")
	}
}

func TestRegisterQueryHandler_AddsToRegistry(t *testing.T) {
	const testTool = "test_query_handler_unit"
	called := false
	registerQueryHandler(testTool, func(_ *ToolExecutor, _ string, _ map[string]interface{}) (mcquery.QueryResult, error) {
		called = true
		return mcquery.QueryResult{}, nil
	})
	defer delete(queryHandlerRegistry, testTool)

	h, ok := queryHandlerRegistry[testTool]
	if !ok {
		t.Fatal("query handler not registered")
	}
	_, err := h(nil, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
