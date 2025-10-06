package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestStreamWriter_WriteRecord(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	record := map[string]interface{}{
		"tool":   "Bash",
		"status": "success",
	}

	err := sw.WriteRecord(record)
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	// Verify output is valid JSON followed by newline
	output := buf.String()
	if !strings.HasSuffix(output, "\n") {
		t.Error("Output should end with newline")
	}

	// Verify JSON is valid
	var decoded map[string]interface{}
	line := strings.TrimSpace(output)
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}

	// Verify content
	if decoded["tool"] != "Bash" {
		t.Errorf("Expected tool='Bash', got '%s'", decoded["tool"])
	}
}

func TestStreamWriter_WriteMultipleRecords(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	records := []map[string]interface{}{
		{"id": float64(1), "tool": "Bash"},
		{"id": float64(2), "tool": "Edit"},
		{"id": float64(3), "tool": "Read"},
	}

	for _, record := range records {
		if err := sw.WriteRecord(record); err != nil {
			t.Fatalf("WriteRecord failed: %v", err)
		}
	}

	// Verify output is JSONL (3 lines)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var decoded map[string]interface{}
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			t.Errorf("Line %d invalid JSON: %v", i+1, err)
		}
	}
}

func TestStreamWriter_EmptyData(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	// No records written
	output := buf.String()
	if output != "" {
		t.Error("Expected empty output for no records")
	}

	// Verify we can still use the writer
	if sw == nil {
		t.Error("StreamWriter should not be nil")
	}
}

func TestStreamWriter_ComplexNestedData(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	record := map[string]interface{}{
		"tool": "Bash",
		"input": map[string]interface{}{
			"command": "ls -la",
			"timeout": float64(5000),
		},
		"tags": []string{"filesystem", "list"},
	}

	err := sw.WriteRecord(record)
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	// Verify nested structure is preserved
	var decoded map[string]interface{}
	line := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify nested input map
	input, ok := decoded["input"].(map[string]interface{})
	if !ok {
		t.Fatal("Input field should be a map")
	}

	if input["command"] != "ls -la" {
		t.Errorf("Expected command='ls -la', got '%v'", input["command"])
	}
}

func TestStreamWriter_NilWriter(t *testing.T) {
	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewStreamWriter should handle nil writer gracefully, but panicked: %v", r)
		}
	}()

	// Create with nil writer (should not panic during creation)
	sw := NewStreamWriter(nil)
	if sw == nil {
		t.Error("NewStreamWriter should not return nil")
	}
}

func TestStreamWriter_WriteError(t *testing.T) {
	// Create a writer that always fails
	failWriter := &failingWriter{}
	sw := NewStreamWriter(failWriter)

	record := map[string]interface{}{
		"tool": "Bash",
	}

	err := sw.WriteRecord(record)
	if err == nil {
		t.Error("Expected error when writing to failing writer")
	}
}

// failingWriter is a test helper that always returns an error
type failingWriter struct{}

func (fw *failingWriter) Write(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

func TestStreamWriter_CompactJSON(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	record := map[string]interface{}{
		"tool":   "Bash",
		"status": "success",
	}

	err := sw.WriteRecord(record)
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	output := strings.TrimSpace(buf.String())

	// Verify it's compact (no extra whitespace)
	if strings.Contains(output, "  ") {
		t.Error("Output should be compact JSON (no extra whitespace)")
	}

	// Verify it's a single line (no internal newlines)
	if strings.Contains(output, "\n") {
		t.Error("Each record should be a single line (no internal newlines)")
	}
}

func TestStreamWriter_SpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	record := map[string]interface{}{
		"error": "File not found: /tmp/test\nLine 2",
		"path":  "C:\\Users\\test\\file.txt",
		"emoji": "🤖",
	}

	err := sw.WriteRecord(record)
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	// Verify special characters are properly escaped
	var decoded map[string]interface{}
	line := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Errorf("Failed to decode JSON with special characters: %v", err)
	}

	// Verify newlines and backslashes are preserved
	if !strings.Contains(decoded["error"].(string), "\n") {
		t.Error("Newlines should be preserved in error field")
	}

	if !strings.Contains(decoded["path"].(string), "\\") {
		t.Error("Backslashes should be preserved in path field")
	}
}
