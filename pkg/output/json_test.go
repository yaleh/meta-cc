package output

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/yale/meta-cc/internal/parser"
)

func TestFormatErrorJSON(t *testing.T) {
	err := errors.New("test error message")

	output, formatErr := FormatErrorJSON(err, "TEST_ERROR_CODE")
	if formatErr != nil {
		t.Fatalf("FormatErrorJSON failed: %v", formatErr)
	}

	// Verify JSON structure
	var errObj map[string]string
	if jsonErr := json.Unmarshal([]byte(output), &errObj); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v", jsonErr)
	}

	if errObj["error"] != "test error message" {
		t.Errorf("expected error='test error message', got '%s'", errObj["error"])
	}

	if errObj["code"] != "TEST_ERROR_CODE" {
		t.Errorf("expected code='TEST_ERROR_CODE', got '%s'", errObj["code"])
	}

	// Should be compact JSON (no indentation)
	if strings.Contains(output, "\n") {
		t.Error("error JSON should be compact (no newlines)")
	}
}

func TestFormatJSONL_ToolCalls(t *testing.T) {
	tools := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Bash",
			Status:   "success",
			Error:    "",
		},
		{
			UUID:     "uuid-2",
			ToolName: "Edit",
			Status:   "error",
			Error:    "file not found",
		},
	}

	output, err := FormatJSONL(tools)
	if err != nil {
		t.Fatalf("FormatJSONL failed: %v", err)
	}

	// JSONL format: one object per line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (JSONL), got %d", len(lines))
	}

	// Verify first line
	var first parser.ToolCall
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("line 1 is not valid JSON: %v", err)
	}

	if first.UUID != "uuid-1" {
		t.Errorf("expected UUID 'uuid-1', got '%s'", first.UUID)
	}

	if first.ToolName != "Bash" {
		t.Errorf("expected ToolName 'Bash', got '%s'", first.ToolName)
	}

	// Verify second line
	var second parser.ToolCall
	if err := json.Unmarshal([]byte(lines[1]), &second); err != nil {
		t.Fatalf("line 2 is not valid JSON: %v", err)
	}

	if second.UUID != "uuid-2" {
		t.Errorf("expected UUID 'uuid-2', got '%s'", second.UUID)
	}

	if second.Error != "file not found" {
		t.Errorf("expected Error 'file not found', got '%s'", second.Error)
	}
}

func TestFormatJSONL_EmptyArray(t *testing.T) {
	tools := []parser.ToolCall{}

	output, err := FormatJSONL(tools)
	if err != nil {
		t.Fatalf("FormatJSONL failed: %v", err)
	}

	// Empty JSONL should be empty string
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty string for empty array, got '%s'", output)
	}
}

func TestFormatJSONL_Struct(t *testing.T) {
	data := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}{
		Name:  "test",
		Value: 42,
	}

	output, err := FormatJSONL(data)
	if err != nil {
		t.Fatalf("FormatJSONL failed: %v", err)
	}

	// Verify it's valid JSON
	var decoded map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(output), &decoded); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v", jsonErr)
	}

	if decoded["name"] != "test" {
		t.Errorf("expected name='test', got '%v'", decoded["name"])
	}

	if int(decoded["value"].(float64)) != 42 {
		t.Errorf("expected value=42, got '%v'", decoded["value"])
	}
}

func TestFormatJSONL_CompactOutput(t *testing.T) {
	tools := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "success"},
	}

	output, err := FormatJSONL(tools)
	if err != nil {
		t.Fatalf("FormatJSONL failed: %v", err)
	}

	// JSONL has one newline at the end (one line per object)
	// For single object, expect exactly 1 newline
	newlineCount := strings.Count(output, "\n")
	if newlineCount != 1 {
		t.Errorf("JSONL with 1 object should have exactly 1 newline, found %d", newlineCount)
	}

	// Should not contain double spaces (indentation indicator)
	if strings.Contains(output, "  ") {
		t.Error("JSONL should be compact (no indentation)")
	}
}

// TestFormatJSONL_ProperJSONLFormat tests that JSONL output is one object per line
// This documents the expected behavior per Phase 13 principles
func TestFormatJSONL_ProperJSONLFormat(t *testing.T) {
	type ErrorEntry struct {
		UUID     string `json:"uuid"`
		ToolName string `json:"tool_name"`
		Error    string `json:"error"`
	}

	errors := []ErrorEntry{
		{UUID: "uuid-1", ToolName: "Bash", Error: "command not found"},
		{UUID: "uuid-2", ToolName: "Read", Error: "file not found"},
		{UUID: "uuid-3", ToolName: "Edit", Error: "invalid syntax"},
	}

	output, err := FormatJSONL(errors)
	if err != nil {
		t.Fatalf("FormatJSONL failed: %v", err)
	}

	// CRITICAL: JSONL format should be ONE JSON OBJECT PER LINE, not a JSON array
	// Expected output:
	// {"uuid":"uuid-1","tool_name":"Bash","error":"command not found"}
	// {"uuid":"uuid-2","tool_name":"Read","error":"file not found"}
	// {"uuid":"uuid-3","tool_name":"Edit","error":"invalid syntax"}

	// NOT expected (JSON Array format):
	// [{"uuid":"uuid-1",...},{"uuid":"uuid-2",...},{"uuid":"uuid-3",...}]

	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Fatalf("JSONL should have 3 lines (one per object), got %d lines. Output:\n%s", len(lines), output)
	}

	// Verify each line is valid JSON object
	for i, line := range lines {
		var decoded ErrorEntry
		if jsonErr := json.Unmarshal([]byte(line), &decoded); jsonErr != nil {
			t.Errorf("line %d is not valid JSON object: %v\nLine: %s", i+1, jsonErr, line)
		}
	}

	// Verify first line
	var first ErrorEntry
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("failed to unmarshal first line: %v", err)
	}
	if first.UUID != "uuid-1" || first.ToolName != "Bash" {
		t.Errorf("first line incorrect: %+v", first)
	}

	// Verify it does NOT start with '[' (JSON Array marker)
	if strings.HasPrefix(output, "[") {
		t.Error("BUG: JSONL output should NOT be JSON Array format (should not start with '[')")
	}
}
