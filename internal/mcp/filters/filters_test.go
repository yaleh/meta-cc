package filters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- TruncateMessageContent tests ---

func TestTruncateMessageContent_Basic(t *testing.T) {
	tests := []struct {
		name           string
		messages       []interface{}
		maxLen         int
		expectTruncate bool
		expectContent  string
	}{
		{
			name: "truncate long content",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(1),
					"timestamp":     "2025-10-06T12:00:00Z",
					"content":       strings.Repeat("a", 1000),
				},
			},
			maxLen:         500,
			expectTruncate: true,
			expectContent:  strings.Repeat("a", 500) + "... [TRUNCATED]",
		},
		{
			name: "short content not truncated",
			messages: []interface{}{
				map[string]interface{}{
					"content": "short content",
				},
			},
			maxLen:         500,
			expectTruncate: false,
			expectContent:  "short content",
		},
		{
			name: "zero maxLen returns original",
			messages: []interface{}{
				map[string]interface{}{
					"content": strings.Repeat("a", 1000),
				},
			},
			maxLen:         0,
			expectTruncate: false,
			expectContent:  strings.Repeat("a", 1000),
		},
		{
			name:           "empty messages",
			messages:       []interface{}{},
			maxLen:         500,
			expectTruncate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateMessageContent(tt.messages, tt.maxLen)
			if len(result) != len(tt.messages) {
				t.Fatalf("expected %d messages, got %d", len(tt.messages), len(result))
			}
			if len(tt.messages) == 0 {
				return
			}
			msgMap, ok := result[0].(map[string]interface{})
			if !ok {
				t.Fatal("expected map result")
			}
			if tt.expectContent != "" {
				content, _ := msgMap["content"].(string)
				if content != tt.expectContent {
					t.Errorf("expected content=%q, got %q", tt.expectContent, content)
				}
			}
			if tt.expectTruncate {
				truncated, _ := msgMap["content_truncated"].(bool)
				if !truncated {
					t.Error("expected content_truncated=true")
				}
			}
		})
	}
}

func TestTruncateMessageContent_NestedStructure(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{
			"type":      "user",
			"timestamp": "2026-03-08T07:57:25Z",
			"message": map[string]interface{}{
				"content": strings.Repeat("x", 500),
				"role":    "user",
			},
		},
	}

	result := TruncateMessageContent(messages, 100)

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	msgMap := result[0].(map[string]interface{})
	nested, ok := msgMap["message"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nested message map to exist")
	}

	content, ok := nested["content"].(string)
	if !ok {
		t.Fatal("expected content to be string")
	}

	if len(content) > 120 { // 100 + "... [TRUNCATED]"
		t.Errorf("content not truncated, length=%d", len(content))
	}

	if truncated, ok := msgMap["content_truncated"].(bool); !ok || !truncated {
		t.Error("expected content_truncated=true")
	}
}

func TestTruncateMessageContent_Immutability(t *testing.T) {
	original := []interface{}{
		map[string]interface{}{
			"content": strings.Repeat("a", 1000),
		},
	}
	originalMap := original[0].(map[string]interface{})
	originalContent := originalMap["content"].(string)

	TruncateMessageContent(original, 500)

	afterContent := originalMap["content"].(string)
	if afterContent != originalContent {
		t.Error("TruncateMessageContent mutated the original message")
	}
}

// --- ApplyContentSummary tests ---

func TestApplyContentSummary_Basic(t *testing.T) {
	tests := []struct {
		name           string
		messages       []interface{}
		expectPreview  string
		expectFields   []string
		unexpectFields []string
	}{
		{
			name: "long content creates preview",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(42),
					"timestamp":     "2025-10-06T12:00:00Z",
					"content":       strings.Repeat("a", 200),
					"extra_field":   "should be removed",
				},
			},
			expectPreview:  strings.Repeat("a", 100) + "...",
			expectFields:   []string{"turn_sequence", "timestamp", "content_preview"},
			unexpectFields: []string{"content", "extra_field"},
		},
		{
			name: "short content no ellipsis",
			messages: []interface{}{
				map[string]interface{}{
					"content": "short",
				},
			},
			expectPreview: "short",
		},
		{
			name:     "empty messages",
			messages: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyContentSummary(tt.messages, 100)
			if len(result) != len(tt.messages) {
				t.Fatalf("expected %d messages, got %d", len(tt.messages), len(result))
			}
			if len(tt.messages) == 0 {
				return
			}
			msgMap, ok := result[0].(map[string]interface{})
			if !ok {
				t.Fatal("expected map result")
			}
			for _, field := range tt.expectFields {
				if _, exists := msgMap[field]; !exists {
					t.Errorf("expected field %q to exist", field)
				}
			}
			for _, field := range tt.unexpectFields {
				if _, exists := msgMap[field]; exists {
					t.Errorf("field %q should not exist", field)
				}
			}
			if tt.expectPreview != "" {
				preview, _ := msgMap["content_preview"].(string)
				if preview != tt.expectPreview {
					t.Errorf("expected preview=%q, got %q", tt.expectPreview, preview)
				}
			}
		})
	}
}

func TestApplyContentSummary_IncludesSessionID(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{
			"sessionId": "abc-session-123",
			"uuid":      "uuid-001",
			"timestamp": "2026-03-09T10:00:00Z",
			"message": map[string]interface{}{
				"role":    "user",
				"content": "hello world",
			},
		},
	}
	result := ApplyContentSummary(messages, 100)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m := result[0].(map[string]interface{})
	sessionID, ok := m["session_id"]
	if !ok {
		t.Error("missing session_id field")
	}
	if sessionID != "abc-session-123" {
		t.Errorf("session_id = %q, want %q", sessionID, "abc-session-123")
	}
}

// --- ApplyMessageFiltersToData tests ---

func TestApplyMessageFiltersToData_ContentSummary(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{
			"content": strings.Repeat("a", 200),
			"extra":   "value",
		},
	}
	result := ApplyMessageFiltersToData(messages, 0, true, 50)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m := result[0].(map[string]interface{})
	if _, exists := m["extra"]; exists {
		t.Error("extra field should be removed by content summary")
	}
	preview, _ := m["content_preview"].(string)
	if !strings.HasSuffix(preview, "...") {
		t.Errorf("expected ellipsis in preview, got %q", preview)
	}
}

func TestApplyMessageFiltersToData_Truncate(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{
			"content": strings.Repeat("b", 1000),
		},
	}
	result := ApplyMessageFiltersToData(messages, 100, false, 0)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m := result[0].(map[string]interface{})
	content, _ := m["content"].(string)
	expected := strings.Repeat("b", 100) + "... [TRUNCATED]"
	if content != expected {
		t.Errorf("expected truncated content, got %q", content)
	}
}

// --- ExpandContextTurns tests ---

// writeJSONLFile writes a slice of objects as JSONL to a temp file.
func writeJSONLFile(t *testing.T, dir string, name string, objects []map[string]interface{}) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create JSONL file: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, obj := range objects {
		if err := enc.Encode(obj); err != nil {
			t.Fatalf("failed to write JSONL: %v", err)
		}
	}
	return path
}

// --- ApplyContentSummary tool_use tests ---

func TestApplyContentSummary_ToolUse_NonEmpty(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{
			"type":      "tool_use",
			"name":      "mcp__manda__Dispatch",
			"input":     map[string]interface{}{"id": "p-1", "action": "run"},
			"sessionId": "s1",
			"uuid":      "u1",
			"timestamp": "2026-07-15T00:00:00Z",
		},
	}
	result := ApplyContentSummary(messages, 100)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m := result[0].(map[string]interface{})
	preview, _ := m["content_preview"].(string)
	if preview == "" {
		t.Error("expected non-empty content_preview for tool_use block, got empty string")
	}
}

func TestApplyContentSummary_ToolUse_Format(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{
			"type":      "tool_use",
			"name":      "mcp__manda__Dispatch",
			"input":     map[string]interface{}{"id": "p-1"},
			"sessionId": "s1",
			"uuid":      "u1",
			"timestamp": "2026-07-15T00:00:00Z",
		},
	}
	result := ApplyContentSummary(messages, 200)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m := result[0].(map[string]interface{})
	preview, _ := m["content_preview"].(string)
	// Must match: name followed by space and JSON object opening brace
	// e.g. "mcp__manda__Dispatch {"
	if !strings.HasPrefix(preview, "mcp__manda__Dispatch {") {
		t.Errorf("content_preview format incorrect, got %q, want prefix %q", preview, "mcp__manda__Dispatch {")
	}
}

func TestApplyContentSummary_ToolUse_PreviewLength(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{
			"type":      "tool_use",
			"name":      "very_long_tool_name",
			"input":     map[string]interface{}{"key": strings.Repeat("v", 200)},
			"sessionId": "s1",
			"uuid":      "u1",
			"timestamp": "2026-07-15T00:00:00Z",
		},
	}
	previewLength := 20
	result := ApplyContentSummary(messages, previewLength)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m := result[0].(map[string]interface{})
	preview, _ := m["content_preview"].(string)
	runeLen := len([]rune(preview))
	// max is previewLength runes + "..." = 23 runes
	if runeLen > previewLength+3 {
		t.Errorf("content_preview rune length %d exceeds max %d", runeLen, previewLength+3)
	}
}

func TestApplyContentSummary_ToolResult_Unchanged(t *testing.T) {
	// tool_result records have a flat "content" string field
	messages := []interface{}{
		map[string]interface{}{
			"type":      "tool_result",
			"content":   "some result text",
			"sessionId": "s1",
			"uuid":      "u2",
			"timestamp": "2026-07-15T00:00:00Z",
		},
	}
	result := ApplyContentSummary(messages, 100)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	m := result[0].(map[string]interface{})
	preview, _ := m["content_preview"].(string)
	if preview == "" {
		t.Error("expected non-empty content_preview for tool_result with content field")
	}
	if preview != "some result text" {
		t.Errorf("tool_result preview changed, got %q, want %q", preview, "some result text")
	}
}

func TestExpandContextTurns_EmptyInput(t *testing.T) {
	result, err := ExpandContextTurns([]interface{}{}, 2, "/tmp", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestExpandContextTurns_ZeroN(t *testing.T) {
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u1", "sessionId": "s1"},
	}
	result, err := ExpandContextTurns(rawData, 0, "/tmp", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result (passthrough), got %d", len(result))
	}
}

func TestExpandContextTurns_Basic(t *testing.T) {
	dir := t.TempDir()

	// Create a session with 5 turns
	sessionID := "test-session-001"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID, "turn": float64(0)},
		{"uuid": "u1", "sessionId": sessionID, "turn": float64(1)},
		{"uuid": "u2", "sessionId": sessionID, "turn": float64(2)}, // matched
		{"uuid": "u3", "sessionId": sessionID, "turn": float64(3)},
		{"uuid": "u4", "sessionId": sessionID, "turn": float64(4)},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// rawData: only turn 2 matched
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u2", "sessionId": sessionID, "turn": float64(2)},
	}

	result, err := ExpandContextTurns(rawData, 1, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With N=1 around index 2: turns 1,2,3 should be included
	if len(result) != 3 {
		t.Fatalf("expected 3 results (turns 1,2,3), got %d", len(result))
	}

	// Verify context flags
	for _, entry := range result {
		obj := entry.(map[string]interface{})
		uuid := obj["uuid"].(string)
		ctx := obj["context"].(bool)
		if uuid == "u2" && ctx != false {
			t.Errorf("matched turn u2 should have context=false, got %v", ctx)
		}
		if (uuid == "u1" || uuid == "u3") && ctx != true {
			t.Errorf("context turn %s should have context=true, got %v", uuid, ctx)
		}
	}
}

func TestExpandContextTurns_WindowClampAtStart(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-clamp-start"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID},
		{"uuid": "u1", "sessionId": sessionID},
		{"uuid": "u2", "sessionId": sessionID},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// Match turn 0: window should be [0,1] (clamped)
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u0", "sessionId": sessionID},
	}

	result, err := ExpandContextTurns(rawData, 2, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
}

func TestExpandContextTurns_WindowClampAtEnd(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-clamp-end"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID},
		{"uuid": "u1", "sessionId": sessionID},
		{"uuid": "u2", "sessionId": sessionID},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// Match last turn: window should be clamped
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u2", "sessionId": sessionID},
	}

	result, err := ExpandContextTurns(rawData, 2, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
}

func TestExpandContextTurns_OverlappingWindows(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-overlap"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID},
		{"uuid": "u1", "sessionId": sessionID},
		{"uuid": "u2", "sessionId": sessionID},
		{"uuid": "u3", "sessionId": sessionID},
		{"uuid": "u4", "sessionId": sessionID},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// Match turns 1 and 3 with N=1 -> windows [0,1,2] and [2,3,4] overlap at 2
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u1", "sessionId": sessionID},
		map[string]interface{}{"uuid": "u3", "sessionId": sessionID},
	}

	result, err := ExpandContextTurns(rawData, 1, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All 5 turns should appear exactly once
	if len(result) != 5 {
		t.Fatalf("expected 5 results (no duplicates), got %d", len(result))
	}

	uuidSeen := make(map[string]int)
	for _, entry := range result {
		obj := entry.(map[string]interface{})
		uuid := obj["uuid"].(string)
		uuidSeen[uuid]++
	}
	for uuid, count := range uuidSeen {
		if count != 1 {
			t.Errorf("uuid %s appeared %d times, expected 1", uuid, count)
		}
	}
}

func TestExpandContextTurns_MultipleSessionOrder(t *testing.T) {
	dir := t.TempDir()

	// Two sessions
	sessA := "session-alpha"
	sessB := "session-beta"

	turnsA := []map[string]interface{}{
		{"uuid": "a0", "sessionId": sessA},
		{"uuid": "a1", "sessionId": sessA},
		{"uuid": "a2", "sessionId": sessA},
	}
	turnsB := []map[string]interface{}{
		{"uuid": "b0", "sessionId": sessB},
		{"uuid": "b1", "sessionId": sessB},
	}
	writeJSONLFile(t, dir, fmt.Sprintf("%s.jsonl", sessA), turnsA)
	writeJSONLFile(t, dir, fmt.Sprintf("%s.jsonl", sessB), turnsB)

	rawData := []interface{}{
		map[string]interface{}{"uuid": "a1", "sessionId": sessA},
		map[string]interface{}{"uuid": "b0", "sessionId": sessB},
	}

	result, err := ExpandContextTurns(rawData, 1, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// sessA: a1 matched -> window [a0,a1,a2] = 3 turns
	// sessB: b0 matched -> window [b0,b1] = 2 turns (clamped)
	if len(result) != 5 {
		t.Fatalf("expected 5 results, got %d", len(result))
	}
}

func TestExpandContextTurns_SnakeCaseSessionID(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-snake"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID},
		{"uuid": "u1", "sessionId": sessionID},
		{"uuid": "u2", "sessionId": sessionID},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// rawData uses snake_case session_id
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u1", "session_id": sessionID},
	}

	result, err := ExpandContextTurns(rawData, 1, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
}

// TestLoadTurnsForSession_LargeImageLine_NoError verifies that loadTurnsForSession
// does NOT return an error when a JSONL file contains a large image line (>10MB),
// and that the matching turn is still returned with binary content replaced.
func TestLoadTurnsForSession_LargeImageLine_NoError(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-large-image"
	largeBase64 := strings.Repeat("A", 11*1024*1024) // 11 MB, exceeds old 10 MB Scanner limit

	// Write JSONL using actual Claude Code image structure (triggers stripImageData)
	imageLine := `{"uuid":"img-turn","sessionId":"` + sessionID + `","message":{"content":[{"type":"tool_result","content":[{"type":"image","source":{"type":"base64","media_type":"image/png","data":"` + largeBase64 + `"}}]}]}}`
	normalLine := `{"uuid":"normal-turn","sessionId":"` + sessionID + `","content":"hello"}`

	path := filepath.Join(dir, "session.jsonl")
	if err := os.WriteFile(path, []byte(imageLine+"\n"+normalLine+"\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	turns, err := loadTurnsForSession(dir, sessionID)
	if err != nil {
		t.Fatalf("loadTurnsForSession returned unexpected error: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}

	// First turn: large base64 should be stripped (data replacement tested at stripImageData level)
	imgJSON, _ := json.Marshal(turns[0])
	if bytes.Contains(imgJSON, []byte(largeBase64)) {
		t.Error("large base64 data should have been stripped from image turn")
	}
	if !bytes.Contains(imgJSON, []byte("binary-omitted")) {
		t.Error("expected binary-omitted placeholder in image turn")
	}

	// Second turn: normal content preserved
	normalObj, ok := turns[1].(map[string]interface{})
	if !ok {
		t.Fatalf("turns[1] is not a map")
	}
	if normalObj["content"] != "hello" {
		t.Errorf("expected content=\"hello\", got %v", normalObj["content"])
	}
}

func TestExpandContextTurns_InvalidBaseDir(t *testing.T) {
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u1", "sessionId": "s1"},
	}

	_, err := ExpandContextTurns(rawData, 1, "/nonexistent/path/that/does/not/exist", false)
	if err == nil {
		t.Error("expected error for invalid base dir")
	}
}

func TestExpandContextTurns_ContextFieldAdded(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-ctx"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID},
		{"uuid": "u1", "sessionId": sessionID},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	rawData := []interface{}{
		map[string]interface{}{"uuid": "u0", "sessionId": sessionID},
	}

	result, err := ExpandContextTurns(rawData, 1, dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, entry := range result {
		obj := entry.(map[string]interface{})
		if _, exists := obj["context"]; !exists {
			t.Errorf("turn %s missing 'context' field", obj["uuid"])
		}
	}
}

// TestExpandContextTurns_SkipsCompactSummary verifies that compact summary entries
// are excluded from results even when they fall within the context window.
func TestExpandContextTurns_SkipsCompactSummary(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-compact-skip"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID},
		{"uuid": "u1", "sessionId": sessionID, "isCompactSummary": true},
		{"uuid": "u2", "sessionId": sessionID},
		{"uuid": "u3", "sessionId": sessionID},
		{"uuid": "u4", "sessionId": sessionID},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// Match u2, N=2 -> window covers u0..u4, but u1 is compact and must be excluded
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u2", "sessionId": sessionID},
	}

	result, err := ExpandContextTurns(rawData, 2, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, entry := range result {
		obj, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		isCompact, _ := obj["isCompactSummary"].(bool)
		if isCompact {
			t.Errorf("result must not contain any isCompactSummary=true entry, but found uuid=%v", obj["uuid"])
		}
	}
}

// TestExpandContextTurns_ExcludeCompactSummaries verifies that when excludeCompactSummaries=true,
// entries with isCompactSummary=true are excluded from context_turns results.
func TestExpandContextTurns_ExcludeCompactSummaries(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-compact-excl"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID},
		{"uuid": "u1", "sessionId": sessionID, "isCompactSummary": true},
		{"uuid": "u2", "sessionId": sessionID},
		{"uuid": "u3", "sessionId": sessionID},
		{"uuid": "u4", "sessionId": sessionID},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// Match u2 with N=2 -> window covers u0..u4; excludeCompactSummaries=true
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u2", "sessionId": sessionID},
	}

	result, err := ExpandContextTurns(rawData, 2, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No isCompactSummary=true entry must appear
	for _, entry := range result {
		obj, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		isCompact, _ := obj["isCompactSummary"].(bool)
		if isCompact {
			t.Errorf("result must not contain any isCompactSummary=true entry, but found uuid=%v", obj["uuid"])
		}
	}

	// u1 must not appear
	for _, entry := range result {
		obj, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if obj["uuid"] == "u1" {
			t.Error("compact summary u1 must not appear in context results")
		}
	}
}

// TestExpandContextTurns_CompactSummaryNotCountedAsTurn verifies that compact summary
// entries are not counted as turns in the window calculation. Session [u0, u1(compact),
// u2, u3, u4], match u2, N=1: u1 is not in uuidToIndex so u2's physical index is 2,
// window covers turns[1..3]. Emit skips compact turns[1] (u1), so result is [u2, u3] (len=2).
func TestExpandContextTurns_CompactSummaryNotCountedAsTurn(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-compact-count"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID},
		{"uuid": "u1", "sessionId": sessionID, "isCompactSummary": true},
		{"uuid": "u2", "sessionId": sessionID},
		{"uuid": "u3", "sessionId": sessionID},
		{"uuid": "u4", "sessionId": sessionID},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// Match u2, N=1: compact u1 is removed before index build.
	// Filtered sequence: [u0, u2, u3, u4]; u2 is at index 1.
	// Window: lo=max(0,1-1)=0, hi=min(3,1+1)=2 -> [u0, u2, u3].
	// Result: 3 entries (u0, u2, u3).
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u2", "sessionId": sessionID},
	}

	result, err := ExpandContextTurns(rawData, 1, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 results (u0, u2, u3), got %d", len(result))
	}

	for _, entry := range result {
		obj := entry.(map[string]interface{})
		if obj["uuid"] == "u1" {
			t.Errorf("compact summary u1 must not appear in result")
		}
	}
}

// TestExpandContextTurns_MatchedCompactSummarySkipped verifies that when the matched
// entry itself is a compact summary it is also excluded from results.
func TestExpandContextTurns_MatchedCompactSummarySkipped(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-compact-matched"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID, "isCompactSummary": true},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// rawData contains only the compact summary entry itself
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u0", "sessionId": sessionID, "isCompactSummary": true},
	}

	result, err := ExpandContextTurns(rawData, 1, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("expected empty result when matched entry is compact summary, got %d", len(result))
	}
}

// TestExpandContextTurns_AllCompactSummarySession verifies that when an entire session
// consists of compact summary entries, the result is empty.
func TestExpandContextTurns_AllCompactSummarySession(t *testing.T) {
	dir := t.TempDir()

	sessionID := "session-all-compact"
	turns := []map[string]interface{}{
		{"uuid": "u0", "sessionId": sessionID, "isCompactSummary": true},
		{"uuid": "u1", "sessionId": sessionID, "isCompactSummary": true},
		{"uuid": "u2", "sessionId": sessionID, "isCompactSummary": true},
	}
	writeJSONLFile(t, dir, "session.jsonl", turns)

	// Match u1 (which is compact)
	rawData := []interface{}{
		map[string]interface{}{"uuid": "u1", "sessionId": sessionID, "isCompactSummary": true},
	}

	result, err := ExpandContextTurns(rawData, 1, dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("expected empty result for all-compact session, got %d", len(result))
	}
}
