package tools_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/mcp/tools"
)

func TestStandardToolParameters(t *testing.T) {
	params := tools.StandardToolParameters()

	requiredParams := []string{
		"scope", "provider", "jq_filter", "stats_only",
		"stats_first", "inline_threshold_bytes",
	}

	for _, param := range requiredParams {
		if _, ok := params[param]; !ok {
			t.Errorf("missing standard parameter: %s", param)
		}
	}
}

// TestStandardToolParameters_OutputFormatNotUniversal is the DIR-044
// regression: output_format used to be declared here (merged onto all 16
// tools) even though pipeline.BuildResponse never consulted it for most of
// them. It's now added per-tool (see OutputFormatProperty) only on the four
// tools whose data flows through pipeline.BuildResponse, not as a standard
// parameter merged onto every tool.
func TestStandardToolParameters_OutputFormatNotUniversal(t *testing.T) {
	params := tools.StandardToolParameters()
	if _, ok := params["output_format"]; ok {
		t.Error("output_format must not be a universal StandardToolParameters entry (DIR-044): " +
			"it's only implemented for the 4 tools routed through pipeline.BuildResponse, " +
			"declare it per-tool via tools.OutputFormatProperty() instead")
	}
}

// TestOutputFormat_ScopedToSupportedTools is the DIR-044 regression: only
// the 4 tools whose results are built via internal/mcp/pipeline.BuildResponse
// (and therefore actually honor output_format=tsv) declare "output_format" in
// their MCP schema. The other 12 tools marshal their own typed,
// non-flat-record analyzer results directly and never touch PipelineConfig,
// so output_format is not advertised on them (previously it was declared but
// silently ignored for every one of the 16 tools).
func TestOutputFormat_ScopedToSupportedTools(t *testing.T) {
	supported := map[string]bool{
		"query_sessions":        true,
		"query_session_content": true,
		"query_session_signals": true,
		"query_file_activity":   true,
	}

	for _, def := range tools.GetToolDefinitions() {
		_, hasOutputFormat := def.InputSchema.Properties["output_format"]
		if supported[def.Name] {
			if !hasOutputFormat {
				t.Errorf("tool %q must declare output_format (it is routed through pipeline.BuildResponse)", def.Name)
			}
		} else {
			if hasOutputFormat {
				t.Errorf("tool %q must not declare output_format (it never reaches pipeline.BuildResponse, "+
					"so output_format would be silently ignored)", def.Name)
			}
		}
	}
}

// analysisServiceTools is the set of six MCP tools backed by
// internal/analysis.Service, dispatched through
// internal/mcp/executor.ExecuteSpecialTool via registerHandler
// (analysis_handlers.go). That dispatch path returns before
// internal/mcp/executor.NewToolPipelineConfig / internal/mcp/pipeline.BuildResponse
// is ever reached, so jq_filter/stats_first/inline_threshold_bytes/offset/page_size
// (only consulted there) are always inert for these six tools (DIR-048/DIR-045).
var analysisServiceTools = []string{
	"analyze_bugs",
	"analyze_errors",
	"quality_scan",
	"get_work_patterns",
	"get_timeline",
	"get_tech_debt",
}

// postProcessingParams are the six StandardToolParameters() entries that
// only have meaning on the internal/mcp/query response path: jq_filter
// (post-filter), stats_first (stats-then-details mode),
// inline_threshold_bytes (pipeline hybrid response threshold), and
// offset/page_size (record-array pagination).
var postProcessingParams = []string{"jq_filter", "stats_first", "inline_threshold_bytes", "offset", "output_mode", "page_size"}

// TestAnalysisToolParameters_ExcludesPostProcessingParams is the DIR-048
// regression: AnalysisStandardToolParameters() (the base parameter set used
// for the six analysis.Service-backed tools) must not declare jq_filter,
// stats_first, inline_threshold_bytes, offset, or page_size, since none is
// ever consulted on that dispatch path. scope/provider/stats_only/
// include_subagents are unaffected -- stats_only is genuinely wired
// per-tool (DIR-042), and include_subagents is a data-scope parameter
// (which JSONL files get read at all) rather than a response-formatting
// parameter, so it stays out of this task's scope.
func TestAnalysisToolParameters_ExcludesPostProcessingParams(t *testing.T) {
	params := tools.AnalysisStandardToolParameters()

	for _, p := range postProcessingParams {
		if _, ok := params[p]; ok {
			t.Errorf("AnalysisStandardToolParameters() must not declare %q: it is never consulted "+
				"for analysis.Service tools (see DIR-048)", p)
		}
	}

	for _, p := range []string{"scope", "provider", "stats_only", "include_subagents"} {
		if _, ok := params[p]; !ok {
			t.Errorf("AnalysisStandardToolParameters() must still declare %q", p)
		}
	}
}

// TestAnalysisTools_SchemaExcludesPostProcessingParams is the DIR-048
// regression proving the fix at the live schema level: analyze_bugs,
// analyze_errors, quality_scan, get_work_patterns, get_timeline, and
// get_tech_debt must not advertise jq_filter/stats_first/
// inline_threshold_bytes/offset/page_size in their MCP tool schema. Before
// this fix, all five were merged in via
// StandardToolParameters() with fully-specified, convincing descriptions even
// though internal/analysis/*.go never reads any of them (confirmed by
// `grep -rn '"jq_filter"\|"stats_first"\|"offset"\|"page_size"' internal/analysis/*.go`
// returning zero hits) and the ExecuteSpecialTool dispatch path for these six
// tools never reaches the pipeline code that would have consulted them.
func TestAnalysisTools_SchemaExcludesPostProcessingParams(t *testing.T) {
	index := tools.BuildToolSchemaIndex()

	for _, name := range analysisServiceTools {
		s, err := tools.GetToolSchemaByName(index, name)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", name, err)
		}
		for _, p := range postProcessingParams {
			if _, ok := s.Properties[p]; ok {
				t.Errorf("tool %q must not declare %q (DIR-048: inert for analysis.Service tools)", name, p)
			}
		}
	}
}

// TestPostProcessingParams_StillPresentOnPipelineRoutedTools guards against
// an overly broad fix: jq_filter/stats_first/inline_threshold_bytes/offset/
// page_size must remain on the four tools that actually route through
// internal/mcp/pipeline.BuildResponse.
func TestPostProcessingParams_StillPresentOnPipelineRoutedTools(t *testing.T) {
	pipelineRoutedTools := []string{
		"query_sessions",
		"query_session_content",
		"query_session_signals",
		"query_file_activity",
	}

	index := tools.BuildToolSchemaIndex()
	for _, name := range pipelineRoutedTools {
		s, err := tools.GetToolSchemaByName(index, name)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", name, err)
		}
		for _, p := range postProcessingParams {
			if _, ok := s.Properties[p]; !ok {
				t.Errorf("tool %q must still declare %q (it is routed through pipeline.BuildResponse)", name, p)
			}
		}
	}
}

func TestMergeParameters(t *testing.T) {
	specific := map[string]tools.Property{
		"limit": {
			Type:        "number",
			Description: "Max results",
		},
		"scope": {
			Type:        "string",
			Description: "Custom scope description",
		},
	}

	merged := tools.MergeParameters(specific)

	if _, ok := merged["limit"]; !ok {
		t.Error("specific parameter 'limit' missing")
	}
	if _, ok := merged["jq_filter"]; !ok {
		t.Error("standard parameter 'jq_filter' missing")
	}
	if merged["scope"].Description != "Custom scope description" {
		t.Errorf("parameter override failed, got: %s", merged["scope"].Description)
	}
}

func TestBuildToolScopesSharedParametersToExecutionPath(t *testing.T) {
	defs := tools.GetToolDefinitions()
	byName := make(map[string]tools.Tool, len(defs))
	for _, def := range defs {
		byName[def.Name] = def
	}

	analysis := byName["analyze_errors"].InputSchema.Properties
	if _, ok := analysis["stats_only"]; !ok {
		t.Fatal("analysis tools must retain stats_only, which their handlers read")
	}
	for _, unread := range []string{"jq_filter", "stats_first", "offset", "output_mode", "page_size", "inline_threshold_bytes"} {
		if _, ok := analysis[unread]; ok {
			t.Errorf("analysis tool must not advertise unread query-pipeline parameter %q", unread)
		}
	}

	query := byName["query_session_signals"].InputSchema.Properties
	for _, wired := range []string{"jq_filter", "stats_first", "offset", "output_mode", "page_size", "inline_threshold_bytes"} {
		if _, ok := query[wired]; !ok {
			t.Errorf("query-pipeline tool must retain wired shared parameter %q", wired)
		}
	}
}

func TestGetToolDefinitions(t *testing.T) {
	defs := tools.GetToolDefinitions()
	if len(defs) == 0 {
		t.Error("expected non-empty tool definitions")
	}

	// Verify all tools serialize to JSON
	for _, tool := range defs {
		_, err := json.Marshal(tool)
		if err != nil {
			t.Errorf("tool %s failed to serialize: %v", tool.Name, err)
		}
	}
}

func TestBuildToolSchemaIndex(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	if len(index) == 0 {
		t.Error("expected non-empty schema index")
	}

	// Check a known tool (new consolidated tools)
	if _, ok := index["query_session_content"]; !ok {
		t.Error("expected query_session_content in index")
	}
}

func TestGetToolSchemaByName(t *testing.T) {
	index := tools.BuildToolSchemaIndex()

	schema, err := tools.GetToolSchemaByName(index, "query_session_content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "object" {
		t.Errorf("expected object type, got %s", schema.Type)
	}

	_, err = tools.GetToolSchemaByName(index, "nonexistent_tool")
	if err == nil {
		t.Fatal("expected error for nonexistent tool")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Errorf("expected 'unknown tool' in error, got: %v", err)
	}
}

func TestValidateToolArgs_ValidTool_ValidArgs(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_content", map[string]interface{}{
		"role":  "user",
		"limit": float64(10),
	})
	if err != nil {
		t.Fatalf("unexpected error for valid tool+args: %v", err)
	}
}

func TestValidateToolArgs_ValidTool_EmptyArgs(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_signals", map[string]interface{}{
		"type": "errors",
	})
	if err != nil {
		t.Fatalf("unexpected error for valid args: %v", err)
	}
}

func TestValidateToolArgs_ValidTool_InvalidArgKey(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_content", map[string]interface{}{
		"unknown_key": "value",
	})
	if err == nil {
		t.Fatal("expected error for invalid arg key")
	}
	if !strings.Contains(err.Error(), "unknown_key") {
		t.Errorf("expected 'unknown_key' in error, got: %v", err)
	}
}

func TestValidateToolArgs_UnknownTool(t *testing.T) {
	err := tools.ValidateToolArgs("no_such_tool", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !strings.Contains(err.Error(), "no_such_tool") {
		t.Errorf("expected tool name in error, got: %v", err)
	}
}

// Phase A: New consolidated tool schema tests

func TestNewConsolidatedToolsPresent(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	newTools := []string{
		"query_session_content",
		"query_session_signals",
		"query_file_activity",
	}
	for _, name := range newTools {
		if _, ok := index[name]; !ok {
			t.Errorf("expected new tool %q in schema index", name)
		}
	}
}

func TestQuerySessionContentSchema(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	s, err := tools.GetToolSchemaByName(index, "query_session_content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Type != "object" {
		t.Errorf("expected object type, got %s", s.Type)
	}
	// must have role parameter
	if _, ok := s.Properties["role"]; !ok {
		t.Error("query_session_content must have 'role' parameter")
	}
	// role is required
	found := false
	for _, r := range s.Required {
		if r == "role" {
			found = true
		}
	}
	if !found {
		t.Error("query_session_content: 'role' must be required")
	}
}

func TestQuerySessionSignalsSchema(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	s, err := tools.GetToolSchemaByName(index, "query_session_signals")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.Properties["type"]; !ok {
		t.Error("query_session_signals must have 'type' parameter")
	}
	found := false
	for _, r := range s.Required {
		if r == "type" {
			found = true
		}
	}
	if !found {
		t.Error("query_session_signals: 'type' must be required")
	}
}

func TestQueryFileActivitySchema(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	s, err := tools.GetToolSchemaByName(index, "query_file_activity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.Properties["type"]; !ok {
		t.Error("query_file_activity must have 'type' parameter")
	}
	found := false
	for _, r := range s.Required {
		if r == "type" {
			found = true
		}
	}
	if !found {
		t.Error("query_file_activity: 'type' must be required")
	}
}

func TestValidateToolArgs_QuerySessionContent(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_content", map[string]interface{}{
		"role":  "user",
		"limit": float64(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateToolArgs_QuerySessionSignals(t *testing.T) {
	err := tools.ValidateToolArgs("query_session_signals", map[string]interface{}{
		"type": "errors",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateToolArgs_QueryFileActivity(t *testing.T) {
	err := tools.ValidateToolArgs("query_file_activity", map[string]interface{}{
		"type": "snapshots",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestQuerySessionContentSchema_ToolRole_DescribesContextFields verifies that the
// block_type parameter description mentions context fields (timestamp/sessionId/turn).
func TestQuerySessionContentSchema_ToolRole_DescribesContextFields(t *testing.T) {
	index := tools.BuildToolSchemaIndex()
	s, err := tools.GetToolSchemaByName(index, "query_session_content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	prop, ok := s.Properties["block_type"]
	if !ok {
		t.Fatal("query_session_content must have 'block_type' parameter")
	}
	desc := prop.Description
	if !strings.Contains(desc, "timestamp") && !strings.Contains(desc, "sessionId") {
		t.Errorf("block_type description should mention context fields (timestamp/sessionId), got: %q", desc)
	}
}

// Phase D-description: Description hints tests (TASK-13)

func TestQuerySessionContentDescriptionHints(t *testing.T) {
	defs := tools.GetToolDefinitions()
	for _, tool := range defs {
		if tool.Name == "query_session_content" {
			if !strings.Contains(tool.Description, "timestamp") {
				t.Errorf("query_session_content description should contain 'timestamp', got: %q", tool.Description)
			}
			if !strings.Contains(tool.Description, "inspect_session_files") {
				t.Errorf("query_session_content description should contain 'inspect_session_files', got: %q", tool.Description)
			}
			return
		}
	}
	t.Fatal("query_session_content not found in tool definitions")
}

func TestQuerySessionSignalsDescriptionHints(t *testing.T) {
	defs := tools.GetToolDefinitions()
	for _, tool := range defs {
		if tool.Name == "query_session_signals" {
			if !strings.Contains(tool.Description, "since") {
				t.Errorf("query_session_signals description should contain 'since', got: %q", tool.Description)
			}
			if !strings.Contains(tool.Description, "until") {
				t.Errorf("query_session_signals description should contain 'until', got: %q", tool.Description)
			}
			return
		}
	}
	t.Fatal("query_session_signals not found in tool definitions")
}

func TestExecuteStage2QueryDescriptionHints(t *testing.T) {
	defs := tools.GetToolDefinitions()
	for _, tool := range defs {
		if tool.Name == "execute_stage2_query" {
			if !strings.Contains(tool.Description, "inspect_session_files") {
				t.Errorf("execute_stage2_query description should contain 'inspect_session_files', got: %q", tool.Description)
			}
			if !strings.Contains(tool.Description, "transform") {
				t.Errorf("execute_stage2_query description should contain 'transform', got: %q", tool.Description)
			}
			return
		}
	}
	t.Fatal("execute_stage2_query not found in tool definitions")
}

// DIR-024: get_session_directory and get_session_metadata must advertise
// "provider" and "working_dir" so the Stage 1 discovery tools describe the
// same provider-aware contract as the convenience query tools.
func TestGetSessionDirectoryAndMetadataSchema_AdvertiseProviderAndWorkingDir(t *testing.T) {
	index := tools.BuildToolSchemaIndex()

	for _, name := range []string{"get_session_directory", "get_session_metadata"} {
		s, err := tools.GetToolSchemaByName(index, name)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", name, err)
		}
		if _, ok := s.Properties["provider"]; !ok {
			t.Errorf("%s must have 'provider' parameter", name)
		}
		if _, ok := s.Properties["working_dir"]; !ok {
			t.Errorf("%s must have 'working_dir' parameter", name)
		}
	}
}

// Phase D: Old tools must be removed from tool definitions.
func TestOldToolsRemovedFromDefinitions(t *testing.T) {
	removedTools := []string{
		"query_tool_errors",
		"query_token_usage",
		"query_system_errors",
		"query_timestamps",
		"query_tools",
		"query_summaries",
		"query_tool_blocks",
		"query_user_messages",
		"query_conversation_flow",
		"query_file_snapshots",
	}
	index := tools.BuildToolSchemaIndex()
	for _, name := range removedTools {
		if _, ok := index[name]; ok {
			t.Errorf("old tool %q should have been removed from definitions in Phase D", name)
		}
	}
}
