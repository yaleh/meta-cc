package pipeline_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/filter"
	"github.com/yaleh/meta-cc/internal/mcp/pipeline"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
)

// testConfig returns a minimal config suitable for pipeline tests.
func testConfig() *config.Config {
	return &config.Config{
		Output: config.OutputConfig{
			Mode:            "auto",
			InlineThreshold: 32768,
		},
	}
}

// ─── InjectWarnings ───────────────────────────────────────────────────────────

func TestInjectWarnings_NoWarnings(t *testing.T) {
	input := `{"mode":"inline","data":[]}`
	out, err := pipeline.InjectWarnings(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	w, ok := parsed["warnings"]
	if !ok {
		t.Fatal("expected 'warnings' field")
	}
	// nil warnings should become empty slice
	arr, ok := w.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", w)
	}
	if len(arr) != 0 {
		t.Fatalf("expected empty warnings, got %v", arr)
	}
}

func TestInjectWarnings_WithWarnings(t *testing.T) {
	input := `{"mode":"inline","data":[]}`
	warns := []string{"warn1", "warn2"}
	out, err := pipeline.InjectWarnings(input, warns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	arr, ok := parsed["warnings"].([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", parsed["warnings"])
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(arr))
	}
}

func TestInjectWarnings_NonJSONPassthrough(t *testing.T) {
	input := "plain text stats output"
	out, err := pipeline.InjectWarnings(input, []string{"w"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != input {
		t.Fatalf("expected passthrough, got %q", out)
	}
}

// ─── DataToJSONL ──────────────────────────────────────────────────────────────

func TestDataToJSONL_Empty(t *testing.T) {
	out, err := pipeline.DataToJSONL(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty string, got %q", out)
	}
}

func TestDataToJSONL_SingleRecord(t *testing.T) {
	data := []interface{}{map[string]interface{}{"tool": "Bash", "status": "success"}}
	out, err := pipeline.DataToJSONL(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if obj["tool"] != "Bash" {
		t.Fatalf("unexpected tool: %v", obj["tool"])
	}
}

func TestDataToJSONL_MultipleRecords(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"n": 1},
		map[string]interface{}{"n": 2},
		map[string]interface{}{"n": 3},
	}
	out, err := pipeline.DataToJSONL(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
}

// ─── BuildStatsOnlyResponse ───────────────────────────────────────────────────

func TestBuildStatsOnlyResponse_Empty(t *testing.T) {
	// Should not error on empty data
	out, err := pipeline.BuildStatsOnlyResponse(nil, false, "turn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Result is a stats string (may be empty or contain headers)
	_ = out
}

func TestBuildStatsOnlyResponse_TimestampTool(t *testing.T) {
	// query_user_messages uses timestamp stats
	data := []interface{}{
		map[string]interface{}{"timestamp": "2024-01-01T10:00:00Z", "role": "user", "content": "hello"},
	}
	out, err := pipeline.BuildStatsOnlyResponse(data, true, "turn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out
}

func TestBuildStatsOnlyResponse_SessionLevel(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"sessionId": "abc123", "role": "user", "content": "hello"},
	}
	out, err := pipeline.BuildStatsOnlyResponse(data, true, "session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out
}

func TestBuildStatsOnlyResponse_StandardTool(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
		map[string]interface{}{"tool_name": "Read", "status": "error"},
	}
	out, err := pipeline.BuildStatsOnlyResponse(data, false, "turn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out
}

// ─── TimestampStatsTools ──────────────────────────────────────────────────────

func TestTimestampStatsTools_Contents(t *testing.T) {
	// After Phase D: new consolidated tool names for timestamp stats
	expected := []string{
		"query_session_content",
		"query_session_signals",
	}
	for _, name := range expected {
		if !pipeline.TimestampStatsTools[name] {
			t.Errorf("expected %q to be in TimestampStatsTools", name)
		}
	}
	if pipeline.TimestampStatsTools["query_session_signals_tool_stats"] {
		t.Error("internal subtypes should not be in TimestampStatsTools")
	}
}

// ─── BuildStatsFirstResponse ──────────────────────────────────────────────────

func TestBuildStatsFirstResponse_Basic(t *testing.T) {
	rawData := []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
	}
	parsedData := rawData

	out, err := pipeline.BuildStatsFirstResponse(
		testConfig(),
		rawData, parsedData,
		map[string]interface{}{},
		"query_session_signals", false, "turn", nil, "jsonl",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "---") {
		t.Errorf("expected separator '---' in output, got: %s", out)
	}
}

func TestBuildStatsFirstResponse_TimestampTool(t *testing.T) {
	rawData := []interface{}{
		map[string]interface{}{"timestamp": "2024-01-01T10:00:00Z", "role": "user"},
	}
	out, err := pipeline.BuildStatsFirstResponse(
		testConfig(),
		rawData, rawData,
		map[string]interface{}{},
		"query_session_content", true, "turn", nil, "jsonl",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out
}

func TestBuildStatsFirstResponse_SessionLevel(t *testing.T) {
	rawData := []interface{}{
		map[string]interface{}{"sessionId": "abc", "role": "user", "content": "hi"},
	}
	out, err := pipeline.BuildStatsFirstResponse(
		testConfig(),
		rawData, rawData,
		map[string]interface{}{},
		"query_session_content", true, "session", nil, "jsonl",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out
}

// ─── BuildStandardResponse ────────────────────────────────────────────────────

func TestBuildStandardResponse_Basic(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool_name": "Bash"},
	}
	out, err := pipeline.BuildStandardResponse(
		testConfig(),
		data,
		map[string]interface{}{},
		"query_session_signals", nil, "jsonl",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "inline") {
		t.Errorf("expected 'inline' in output, got: %s", out)
	}
}

// ─── PipelineConfig ───────────────────────────────────────────────────────────

func TestPipelineConfig_Defaults(t *testing.T) {
	pc := pipeline.PipelineConfig{}
	if pc.StatsOnly {
		t.Error("expected StatsOnly=false by default")
	}
	if pc.StatsFirst {
		t.Error("expected StatsFirst=false by default")
	}
	if pc.GroupBySession {
		t.Error("expected GroupBySession=false by default")
	}
}

func TestPipelineConfig_DefaultPreviewLength(t *testing.T) {
	if pipeline.DefaultPreviewLength <= 0 {
		t.Errorf("expected positive DefaultPreviewLength, got %d", pipeline.DefaultPreviewLength)
	}
}

// ─── BuildResponse ────────────────────────────────────────────────────────────

func makeQueryResult(entries ...interface{}) mcquery.QueryResult {
	return mcquery.QueryResult{Entries: entries}
}

func TestBuildResponse_StatsOnly(t *testing.T) {
	pc := pipeline.PipelineConfig{StatsOnly: true, StatsLevel: "turn"}
	result := makeQueryResult(
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
	)
	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out
}

func TestBuildResponse_InvalidStatsLevel(t *testing.T) {
	pc := pipeline.PipelineConfig{StatsOnly: true, StatsLevel: "invalid"}
	result := makeQueryResult()
	_, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err == nil {
		t.Fatal("expected error for invalid stats_level")
	}
	if !strings.Contains(err.Error(), "stats_level") {
		t.Errorf("expected 'stats_level' in error, got: %v", err)
	}
}

func TestBuildResponse_GroupBySessionAndStatsOnlyExclusive(t *testing.T) {
	pc := pipeline.PipelineConfig{StatsOnly: true, GroupBySession: true}
	result := makeQueryResult()
	_, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_content", pc)
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
}

func TestBuildResponse_Standard(t *testing.T) {
	pc := pipeline.PipelineConfig{}
	result := makeQueryResult(
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
	)
	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "inline") {
		t.Errorf("expected 'inline' in output, got: %s", out)
	}
}

func TestBuildResponse_StatsFirst(t *testing.T) {
	pc := pipeline.PipelineConfig{StatsFirst: true, StatsLevel: "turn"}
	result := makeQueryResult(
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
	)
	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "---") {
		t.Errorf("expected '---' separator in stats_first output, got: %s", out)
	}
}

// ─── UseTimestampStats / ApplyMessageFilters fields ──────────────────────────

func TestPipelineConfig_NewFields(t *testing.T) {
	pc := pipeline.PipelineConfig{
		UseTimestampStats:   true,
		ApplyMessageFilters: true,
	}
	if !pc.UseTimestampStats {
		t.Error("UseTimestampStats should be settable to true")
	}
	if !pc.ApplyMessageFilters {
		t.Error("ApplyMessageFilters should be settable to true")
	}
}

func TestBuildStatsOnlyResponse_UseTimestampStats_DifferentOutput(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"timestamp": "2024-01-01T10:00:00Z", "key": "val"},
	}
	outTS, err := pipeline.BuildStatsOnlyResponse(data, true, "turn")
	if err != nil {
		t.Fatalf("unexpected error with useTimestampStats=true: %v", err)
	}
	outStd, err := pipeline.BuildStatsOnlyResponse(data, false, "turn")
	if err != nil {
		t.Fatalf("unexpected error with useTimestampStats=false: %v", err)
	}
	if outTS == outStd {
		t.Error("expected different output for useTimestampStats=true vs false")
	}
}

func TestBuildResponse_WithWarnings(t *testing.T) {
	pc := pipeline.PipelineConfig{}
	result := mcquery.QueryResult{
		Entries:  []interface{}{map[string]interface{}{"x": 1}},
		Warnings: []string{"test warning"},
	}
	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	warns, ok := parsed["warnings"].([]interface{})
	if !ok || len(warns) == 0 {
		t.Errorf("expected non-empty warnings in output, got: %v", parsed["warnings"])
	}
}

// TestPipelineConfig_IncludeSubagents verifies that PipelineConfig.IncludeSubagents
// is a real field and defaults correctly via Go zero value (false) vs explicit true.
func TestPipelineConfig_IncludeSubagents(t *testing.T) {
	// Zero value: IncludeSubagents defaults to false in struct literal
	pc := pipeline.PipelineConfig{}
	if pc.IncludeSubagents {
		t.Error("expected PipelineConfig zero value to have IncludeSubagents=false")
	}

	// Explicit true
	pcTrue := pipeline.PipelineConfig{IncludeSubagents: true}
	if !pcTrue.IncludeSubagents {
		t.Error("expected PipelineConfig{IncludeSubagents: true} to have IncludeSubagents=true")
	}
}

// ─── Pagination ─────────────────────────────────────────────────────────────────

func TestBuildStandardResponse_WithPagination(t *testing.T) {
	data := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]interface{}{"idx": i}
	}

	// page 1: offset=0, pageSize=25
	meta := &filter.PaginationMetadata{
		TotalRecords:    100,
		ReturnedRecords: 25,
		Offset:          0,
		Limit:           25,
		HasMore:         true,
	}
	out, err := pipeline.BuildStandardResponse(
		testConfig(),
		data[:25],
		map[string]interface{}{},
		"query_session_signals", meta, "jsonl",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"pagination"`) {
		t.Errorf("expected 'pagination' key in output, got: %s", out)
	}
	if !strings.Contains(out, `"has_more":true`) {
		t.Errorf("expected has_more=true in output, got: %s", out)
	}

	// last page: offset=75, pageSize=25, no more
	meta2 := &filter.PaginationMetadata{
		TotalRecords:    100,
		ReturnedRecords: 25,
		Offset:          75,
		Limit:           25,
		HasMore:         false,
	}
	out2, err := pipeline.BuildStandardResponse(
		testConfig(),
		data[75:],
		map[string]interface{}{},
		"query_session_signals", meta2, "jsonl",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out2, `"has_more":false`) {
		t.Errorf("expected has_more=false in output, got: %s", out2)
	}
}

func TestBuildResponse_WithPagination(t *testing.T) {
	data := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]interface{}{"idx": i}
	}

	pc := pipeline.PipelineConfig{
		Offset:   10,
		PageSize: 20,
	}
	result := makeQueryResult(data...)
	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"pagination"`) {
		t.Errorf("expected 'pagination' key in output, got: %s", out)
	}
	if !strings.Contains(out, `"total_records":100`) {
		t.Errorf("expected total_records=100, got: %s", out)
	}
	if !strings.Contains(out, `"returned_records":20`) {
		t.Errorf("expected returned_records=20, got: %s", out)
	}
	if !strings.Contains(out, `"has_more":true`) {
		t.Errorf("expected has_more=true, got: %s", out)
	}
}

func TestBuildResponse_DefaultNoPagination(t *testing.T) {
	// Verify default behavior (no page_size) is unchanged
	data := []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
	}
	pc := pipeline.PipelineConfig{} // Offset=0, PageSize=0
	result := makeQueryResult(data...)
	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still include pagination metadata but with full dataset info
	if !strings.Contains(out, `"pagination"`) {
		t.Errorf("expected 'pagination' key in output, got: %s", out)
	}
	if !strings.Contains(out, `"has_more":false`) {
		t.Errorf("expected has_more=false for unpaginated, got: %s", out)
	}
	// Data should still be inline for small result
	if !strings.Contains(out, "inline") {
		t.Errorf("expected 'inline' mode, got: %s", out)
	}
}

func TestPipelineConfig_PaginationDefaults(t *testing.T) {
	pc := pipeline.PipelineConfig{}
	if pc.Offset != 0 {
		t.Error("expected Offset=0 by default")
	}
	if pc.PageSize != 0 {
		t.Error("expected PageSize=0 by default")
	}
}

// ─── DIR-041: jq_filter must be applied as a real post-filter ────────────────
//
// Regression coverage for DIR-041: query_session_content, query_session_signals,
// query_sessions, and query_file_activity all declare a "jq_filter" tool
// parameter, and it was read into PipelineConfig.JQFilter but never applied —
// BuildResponse silently ignored it and returned the full unfiltered result.
// These tests exercise BuildResponse directly (the shared implementation
// underlying all four consolidated tools) and would fail against the old
// no-op behavior.

func jqFilterTestEntries() []interface{} {
	return []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
		map[string]interface{}{"tool_name": "Read", "status": "success"},
		map[string]interface{}{"tool_name": "Edit", "status": "error"},
	}
}

func countRecords(t *testing.T, out string) int {
	t.Helper()
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("expected JSON object response, got: %s (err: %v)", out, err)
	}
	data, ok := parsed["data"].([]interface{})
	if !ok {
		t.Fatalf("expected 'data' array in response, got: %v", parsed["data"])
	}
	return len(data)
}

// TestBuildResponse_JQFilter_SelectFalseEmptiesResult is the core DIR-041
// regression: a jq_filter that can never match anything must produce a
// correspondingly empty result, not the full unfiltered set.
func TestBuildResponse_JQFilter_SelectFalseEmptiesResult(t *testing.T) {
	pc := pipeline.PipelineConfig{JQFilter: "select(false)"}
	result := makeQueryResult(jqFilterTestEntries()...)

	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := countRecords(t, out); got != 0 {
		t.Fatalf("expected 0 records with jq_filter=select(false), got %d in: %s", got, out)
	}

	// Same call without jq_filter must return the full, unfiltered set —
	// proving the difference is due to jq_filter, not some other change.
	pcNoFilter := pipeline.PipelineConfig{}
	outNoFilter, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pcNoFilter)
	if err != nil {
		t.Fatalf("unexpected error (no filter): %v", err)
	}
	if got := countRecords(t, outNoFilter); got != len(jqFilterTestEntries()) {
		t.Fatalf("expected %d records without jq_filter, got %d in: %s", len(jqFilterTestEntries()), got, outNoFilter)
	}
}

// TestBuildResponse_JQFilter_NarrowsResult verifies a targeted jq_filter
// (not just an always-false one) actually narrows the result set as
// promised by the tool schema, e.g. `.[] | select(.status == "error")`.
func TestBuildResponse_JQFilter_NarrowsResult(t *testing.T) {
	pc := pipeline.PipelineConfig{JQFilter: `.[] | select(.status == "error")`}
	result := makeQueryResult(jqFilterTestEntries()...)

	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := countRecords(t, out); got != 1 {
		t.Fatalf("expected 1 record (status==error), got %d in: %s", got, out)
	}
	if !strings.Contains(out, `"tool_name":"Edit"`) {
		t.Errorf("expected filtered record to be the Edit/error entry, got: %s", out)
	}
}

// TestBuildResponse_JQFilter_StatsOnly verifies jq_filter narrows the input
// to stats_only mode too (one of the three modes DIR-041 found affected).
func TestBuildResponse_JQFilter_StatsOnly(t *testing.T) {
	pc := pipeline.PipelineConfig{JQFilter: "select(false)", StatsOnly: true, StatsLevel: "turn"}
	result := makeQueryResult(jqFilterTestEntries()...)

	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pcNoFilter := pipeline.PipelineConfig{StatsOnly: true, StatsLevel: "turn"}
	outNoFilter, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pcNoFilter)
	if err != nil {
		t.Fatalf("unexpected error (no filter): %v", err)
	}
	if out == outNoFilter {
		t.Fatalf("expected stats_only output to differ when jq_filter empties the input; both were: %s", out)
	}
}

// TestBuildResponse_JQFilter_StatsFirst verifies jq_filter narrows the input
// to stats_first mode too.
func TestBuildResponse_JQFilter_StatsFirst(t *testing.T) {
	pc := pipeline.PipelineConfig{JQFilter: "select(false)", StatsFirst: true, StatsLevel: "turn"}
	result := makeQueryResult(jqFilterTestEntries()...)

	out, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pcNoFilter := pipeline.PipelineConfig{StatsFirst: true, StatsLevel: "turn"}
	outNoFilter, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pcNoFilter)
	if err != nil {
		t.Fatalf("unexpected error (no filter): %v", err)
	}
	if out == outNoFilter {
		t.Fatalf("expected stats_first output to differ when jq_filter empties the input; both were: %s", out)
	}
}

// TestBuildResponse_JQFilter_DefaultIsNoOp verifies the documented default
// ('.[]', and an omitted/empty JQFilter) leaves results completely
// unchanged — the fix must not alter default (no custom filter) behavior.
func TestBuildResponse_JQFilter_DefaultIsNoOp(t *testing.T) {
	entries := jqFilterTestEntries()

	pcOmitted := pipeline.PipelineConfig{}
	outOmitted, err := pipeline.BuildResponse(testConfig(), makeQueryResult(entries...), map[string]interface{}{}, "query_session_signals", pcOmitted)
	if err != nil {
		t.Fatalf("unexpected error (omitted): %v", err)
	}

	pcDefault := pipeline.PipelineConfig{JQFilter: ".[]"}
	outDefault, err := pipeline.BuildResponse(testConfig(), makeQueryResult(entries...), map[string]interface{}{}, "query_session_signals", pcDefault)
	if err != nil {
		t.Fatalf("unexpected error (explicit '.[]'): %v", err)
	}

	if outOmitted != outDefault {
		t.Fatalf("expected omitted jq_filter and explicit '.[]' to behave identically:\nomitted: %s\ndefault: %s", outOmitted, outDefault)
	}
	if got := countRecords(t, outDefault); got != len(entries) {
		t.Fatalf("expected default jq_filter to return all %d records, got %d", len(entries), got)
	}
}

// TestBuildResponse_JQFilter_InvalidExpressionFailsClosed verifies an
// unparseable jq_filter surfaces an actionable error rather than being
// silently ignored (which would look identical to the DIR-041 no-op bug).
func TestBuildResponse_JQFilter_InvalidExpressionFailsClosed(t *testing.T) {
	pc := pipeline.PipelineConfig{JQFilter: "!!not valid jq!!"}
	result := makeQueryResult(jqFilterTestEntries()...)

	_, err := pipeline.BuildResponse(testConfig(), result, map[string]interface{}{}, "query_session_signals", pc)
	if err == nil {
		t.Fatal("expected error for invalid jq_filter expression, got nil")
	}
	if !strings.Contains(err.Error(), "jq_filter") {
		t.Errorf("expected error to mention 'jq_filter', got: %v", err)
	}
}

// ─── output_format: tsv (DIR-044) ────────────────────────────────────────────
//
// StandardToolParameters() (internal/mcp/tools/tools.go) declares
// "output_format": "jsonl or tsv" on every one of the 16 MCP tools, but until
// this fix BuildResponse never branched on it — a caller passing
// output_format: "tsv" silently got the same JSON/JSONL shape back. These
// tests assert output_format: "tsv" actually changes the returned response
// format relative to the default, for more than one tool name, both via the
// full BuildResponse entry point and via BuildStandardResponse directly.

// isJSONObject reports whether s parses as a JSON object (used to assert TSV
// output is NOT the same JSON envelope the default jsonl format produces).
func isJSONObject(s string) bool {
	var m map[string]interface{}
	return json.Unmarshal([]byte(s), &m) == nil
}

func TestBuildResponse_OutputFormatTSV_DiffersFromDefault(t *testing.T) {
	entries := []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
		map[string]interface{}{"tool_name": "Read", "status": "error"},
	}

	for _, toolName := range []string{"query_sessions", "query_session_content", "analyze_errors"} {
		t.Run(toolName, func(t *testing.T) {
			pcDefault := pipeline.PipelineConfig{}
			outDefault, err := pipeline.BuildResponse(testConfig(), makeQueryResult(entries...), map[string]interface{}{}, toolName, pcDefault)
			if err != nil {
				t.Fatalf("unexpected error (default): %v", err)
			}
			if !isJSONObject(outDefault) {
				t.Fatalf("expected default output_format to be a JSON object, got: %s", outDefault)
			}

			pcTSV := pipeline.PipelineConfig{OutputFormat: "tsv"}
			args := map[string]interface{}{"output_format": "tsv"}
			outTSV, err := pipeline.BuildResponse(testConfig(), makeQueryResult(entries...), args, toolName, pcTSV)
			if err != nil {
				t.Fatalf("unexpected error (tsv): %v", err)
			}

			if outTSV == outDefault {
				t.Fatalf("expected output_format=tsv to differ from default output, got identical output: %s", outTSV)
			}
			if isJSONObject(outTSV) {
				t.Fatalf("expected output_format=tsv to produce a non-JSON, tab-separated response, got JSON: %s", outTSV)
			}
			if !strings.Contains(outTSV, "\t") {
				t.Fatalf("expected output_format=tsv output to contain tab-separated fields, got: %q", outTSV)
			}
			if !strings.Contains(outTSV, "tool_name") || !strings.Contains(outTSV, "status") {
				t.Fatalf("expected tsv header to include record field names, got: %q", outTSV)
			}
		})
	}
}

func TestBuildStandardResponse_OutputFormatTSV(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
	}
	out, err := pipeline.BuildStandardResponse(testConfig(), data, map[string]interface{}{}, "query_session_signals", nil, "tsv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isJSONObject(out) {
		t.Errorf("expected non-JSON tsv output, got: %s", out)
	}
	if !strings.Contains(out, "tool_name\tBash") && !strings.Contains(out, "Bash") {
		t.Errorf("expected tsv output to include record data, got: %q", out)
	}
}

func TestBuildStatsFirstResponse_OutputFormatTSV(t *testing.T) {
	rawData := []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
	}
	out, err := pipeline.BuildStatsFirstResponse(
		testConfig(),
		rawData, rawData,
		map[string]interface{}{},
		"query_session_signals", false, "turn", nil, "tsv",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "---") {
		t.Fatalf("expected '---' separator in stats_first output, got: %s", out)
	}
	detail := strings.SplitN(out, "\n---\n", 2)[1]
	if isJSONObject(detail) {
		t.Errorf("expected non-JSON tsv detail section, got: %s", detail)
	}
	if !strings.Contains(detail, "tool_name\tBash") && !strings.Contains(detail, "Bash") {
		t.Errorf("expected tsv detail section to include record data, got: %q", detail)
	}
}

func TestBuildResponse_OutputFormatDefault_UnaffectedByTSVSupport(t *testing.T) {
	// Regression guard: adding tsv support must not change default (jsonl)
	// behavior when output_format is omitted entirely.
	entries := []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
	}
	pc := pipeline.PipelineConfig{}
	out, err := pipeline.BuildResponse(testConfig(), makeQueryResult(entries...), map[string]interface{}{}, "query_session_signals", pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isJSONObject(out) {
		t.Fatalf("expected default output to remain a JSON object, got: %s", out)
	}
}
