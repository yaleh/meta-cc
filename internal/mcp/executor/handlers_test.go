package executor

import (
	"testing"

	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
)

// buildToolUseRecord constructs a minimal assistant JSONL record with a tool_use block
// and outer context fields (timestamp, sessionId, turn).
func buildToolUseRecord(sessionID string, turn float64, toolName string) map[string]interface{} {
	return map[string]interface{}{
		"type":      "assistant",
		"timestamp": "2024-01-01T00:00:00Z",
		"sessionId": sessionID,
		"turn":      turn,
		"message": map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type":  "tool_use",
					"id":    "toolu_abc",
					"name":  toolName,
					"input": map[string]interface{}{"file_path": "/tmp/foo"},
				},
			},
		},
	}
}

// buildToolResultRecord constructs a minimal user JSONL record with a tool_result block
// and outer context fields (timestamp, sessionId, turn).
func buildToolResultRecord(sessionID string, turn float64) map[string]interface{} {
	return map[string]interface{}{
		"type":      "user",
		"timestamp": "2024-01-01T01:00:00Z",
		"sessionId": sessionID,
		"turn":      turn,
		"message": map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": "toolu_abc",
					"content":     "file contents here",
				},
			},
		},
	}
}

// toolUseJQFilter returns the jq filter that handleQueryToolBlocks uses for tool_use.
// This mirrors the filter in handlers.go and must be kept in sync.
func toolUseJQFilter() string {
	return `select(.type == "assistant") | . as $rec | .message.content[] | select(.type == "tool_use") | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .`
}

// toolResultJQFilter returns the jq filter that handleQueryToolBlocks uses for tool_result.
func toolResultJQFilter() string {
	return `select(.type == "user" and (.message.content | type == "array")) | . as $rec | .message.content[] | select(.type == "tool_result") | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .`
}

// TestHandleQueryToolBlocks_ToolUse_IncludesTimestamp verifies that tool_use results
// include the timestamp from the outer JSONL record.
func TestHandleQueryToolBlocks_ToolUse_IncludesTimestamp(t *testing.T) {
	record := buildToolUseRecord("sess-123", 1, "Read")
	results, err := runProviderJQ([]map[string]interface{}{record}, toolUseJQFilter(), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	m, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", results[0])
	}
	if m["timestamp"] == nil {
		t.Error("expected timestamp field in result, got nil")
	}
	if m["timestamp"] != "2024-01-01T00:00:00Z" {
		t.Errorf("expected timestamp=2024-01-01T00:00:00Z, got %v", m["timestamp"])
	}
}

// TestHandleQueryToolBlocks_ToolUse_IncludesSessionId verifies that tool_use results
// include the sessionId from the outer JSONL record.
func TestHandleQueryToolBlocks_ToolUse_IncludesSessionId(t *testing.T) {
	record := buildToolUseRecord("sess-456", 2, "Write")
	results, err := runProviderJQ([]map[string]interface{}{record}, toolUseJQFilter(), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	m, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", results[0])
	}
	if m["sessionId"] == nil {
		t.Error("expected sessionId field in result, got nil")
	}
	if m["sessionId"] != "sess-456" {
		t.Errorf("expected sessionId=sess-456, got %v", m["sessionId"])
	}
}

// TestHandleQueryToolBlocks_ToolUse_IncludesTurn verifies that tool_use results
// include the turn from the outer JSONL record.
func TestHandleQueryToolBlocks_ToolUse_IncludesTurn(t *testing.T) {
	record := buildToolUseRecord("sess-789", 5, "Bash")
	results, err := runProviderJQ([]map[string]interface{}{record}, toolUseJQFilter(), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	m, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", results[0])
	}
	if m["turn"] == nil {
		t.Error("expected turn field in result, got nil")
	}
	if m["turn"] != float64(5) {
		t.Errorf("expected turn=5, got %v", m["turn"])
	}
}

// TestHandleQueryToolBlocks_ToolResult_IncludesTimestamp verifies that tool_result results
// include the timestamp from the outer JSONL record.
func TestHandleQueryToolBlocks_ToolResult_IncludesTimestamp(t *testing.T) {
	record := buildToolResultRecord("sess-result-123", 3)
	results, err := runProviderJQ([]map[string]interface{}{record}, toolResultJQFilter(), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	m, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", results[0])
	}
	if m["timestamp"] == nil {
		t.Error("expected timestamp field in tool_result, got nil")
	}
	if m["timestamp"] != "2024-01-01T01:00:00Z" {
		t.Errorf("expected correct timestamp, got %v", m["timestamp"])
	}
}

// toolUseWithNameJQFilter returns the jq filter for tool_use blocks with a tool_name filter.
// This mirrors the filter constructed in handleQueryToolBlocks when tool_name is set.
func toolUseWithNameJQFilter(toolName string) string {
	escaped := EscapeJQ(toolName)
	return `select(.type == "assistant") | . as $rec | .message.content[] | select(.type == "tool_use") | select(.name | test("` + escaped + `")) | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .`
}

// TestHandleQueryToolBlocks_ToolName_FiltersExactMatch verifies that tool_name filters
// tool_use blocks by exact name match (substring/regex via jq test()).
func TestHandleQueryToolBlocks_ToolName_FiltersExactMatch(t *testing.T) {
	records := []map[string]interface{}{
		buildToolUseRecord("sess-1", 1, "Read"),
		buildToolUseRecord("sess-1", 2, "Write"),
		buildToolUseRecord("sess-1", 3, "Bash"),
	}
	results, err := runProviderJQ(records, toolUseWithNameJQFilter("Write"), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	m := results[0].(map[string]interface{})
	if m["name"] != "Write" {
		t.Errorf("expected name=Write, got %v", m["name"])
	}
}

// TestHandleQueryToolBlocks_ToolName_FiltersSubstring verifies that tool_name does
// substring matching (e.g. "Dis" matches "Dispatch").
func TestHandleQueryToolBlocks_ToolName_FiltersSubstring(t *testing.T) {
	records := []map[string]interface{}{
		buildToolUseRecord("sess-1", 1, "Dispatch"),
		buildToolUseRecord("sess-1", 2, "DispatchCancel"),
		buildToolUseRecord("sess-1", 3, "Read"),
	}
	results, err := runProviderJQ(records, toolUseWithNameJQFilter("Dispatch"), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for substring 'Dispatch', got %d", len(results))
	}
}

// TestHandleQueryToolBlocks_ToolName_FiltersRegex verifies that tool_name supports
// regex patterns (consistent with pattern parameter behavior).
func TestHandleQueryToolBlocks_ToolName_FiltersRegex(t *testing.T) {
	records := []map[string]interface{}{
		buildToolUseRecord("sess-1", 1, "Read"),
		buildToolUseRecord("sess-1", 2, "Write"),
		buildToolUseRecord("sess-1", 3, "Bash"),
	}
	// Regex: match Read or Write
	results, err := runProviderJQ(records, toolUseWithNameJQFilter("Read|Write"), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for regex 'Read|Write', got %d", len(results))
	}
}

// TestHandleQueryToolBlocks_ToolName_EmptyReturnsAll verifies that omitting tool_name
// returns all tool_use blocks (no filtering applied).
func TestHandleQueryToolBlocks_ToolName_EmptyReturnsAll(t *testing.T) {
	records := []map[string]interface{}{
		buildToolUseRecord("sess-1", 1, "Read"),
		buildToolUseRecord("sess-1", 2, "Write"),
		buildToolUseRecord("sess-1", 3, "Bash"),
	}
	results, err := runProviderJQ(records, toolUseJQFilter(), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results when no tool_name filter, got %d", len(results))
	}
}

// TestHandleQueryToolBlocks_ToolUse_PreservesToolFields verifies that the original
// tool block fields (type, id, name, input) are still present in results and not overwritten.
func TestHandleQueryToolBlocks_ToolUse_PreservesToolFields(t *testing.T) {
	record := buildToolUseRecord("sess-preserve", 7, "Edit")
	// Override the id in the record to match what we test below
	record["message"].(map[string]interface{})["content"].([]interface{})[0].(map[string]interface{})["id"] = "toolu_preserve"

	results, err := runProviderJQ([]map[string]interface{}{record}, toolUseJQFilter(), 0, mcquery.ParsedTimeRange{})
	if err != nil {
		t.Fatalf("runProviderJQ error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	m, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", results[0])
	}

	// Verify original tool_use fields are preserved and not overwritten by context fields
	if m["type"] != "tool_use" {
		t.Errorf("expected type=tool_use, got %v", m["type"])
	}
	if m["id"] != "toolu_preserve" {
		t.Errorf("expected id=toolu_preserve, got %v", m["id"])
	}
	if m["name"] != "Edit" {
		t.Errorf("expected name=Edit, got %v", m["name"])
	}
	if m["input"] == nil {
		t.Error("expected input field to be preserved")
	}
	// Context fields should also be present
	if m["timestamp"] == nil {
		t.Error("expected timestamp field to be present alongside tool fields")
	}
	if m["sessionId"] == nil {
		t.Error("expected sessionId field to be present alongside tool fields")
	}
}
