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

	// Verify it's valid JSON (array)
	var decoded []parser.ToolCall
	if jsonErr := json.Unmarshal([]byte(output), &decoded); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v", jsonErr)
	}

	if len(decoded) != 2 {
		t.Errorf("expected 2 tool calls, got %d", len(decoded))
	}

	// Verify first tool call
	if decoded[0].UUID != "uuid-1" {
		t.Errorf("expected UUID 'uuid-1', got '%s'", decoded[0].UUID)
	}

	if decoded[0].ToolName != "Bash" {
		t.Errorf("expected ToolName 'Bash', got '%s'", decoded[0].ToolName)
	}

	// Verify second tool call
	if decoded[1].UUID != "uuid-2" {
		t.Errorf("expected UUID 'uuid-2', got '%s'", decoded[1].UUID)
	}

	if decoded[1].Error != "file not found" {
		t.Errorf("expected Error 'file not found', got '%s'", decoded[1].Error)
	}
}

func TestFormatJSONL_EmptyArray(t *testing.T) {
	tools := []parser.ToolCall{}

	output, err := FormatJSONL(tools)
	if err != nil {
		t.Fatalf("FormatJSONL failed: %v", err)
	}

	if strings.TrimSpace(output) != "[]" {
		t.Errorf("expected '[]' for empty array, got '%s'", output)
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

	// JSONL should be compact (no indentation, no extra newlines)
	// Count newlines - should have 0 for compact JSON
	newlineCount := strings.Count(output, "\n")
	if newlineCount > 0 {
		t.Errorf("JSONL should be compact (no newlines), found %d newlines", newlineCount)
	}

	// Should not contain double spaces (indentation indicator)
	if strings.Contains(output, "  ") {
		t.Error("JSONL should be compact (no indentation)")
	}
}
