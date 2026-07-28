package response

import (
	"strings"
	"testing"
)

func TestRecordsToTSV_Empty(t *testing.T) {
	out := RecordsToTSV(nil)
	if out != "" {
		t.Errorf("expected empty string for no records, got: %q", out)
	}
}

func TestRecordsToTSV_HeaderAndRows(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tool_name": "Bash", "status": "success"},
		map[string]interface{}{"tool_name": "Read", "status": "error"},
	}
	out := RecordsToTSV(data)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d: %q", len(lines), out)
	}
	// Header should be the sorted union of keys.
	if lines[0] != "status\ttool_name" {
		t.Errorf("expected sorted header 'status\\ttool_name', got: %q", lines[0])
	}
	if lines[1] != "success\tBash" {
		t.Errorf("unexpected row 1: %q", lines[1])
	}
	if lines[2] != "error\tRead" {
		t.Errorf("unexpected row 2: %q", lines[2])
	}
}

func TestRecordsToTSV_UnionOfKeysAcrossRecords(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"a": "1"},
		map[string]interface{}{"a": "2", "b": "3"},
	}
	out := RecordsToTSV(data)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if lines[0] != "a\tb" {
		t.Fatalf("expected header 'a\\tb', got: %q", lines[0])
	}
	if lines[1] != "1\t" {
		t.Errorf("expected missing field to render as empty cell, got: %q", lines[1])
	}
	if lines[2] != "2\t3" {
		t.Errorf("unexpected row: %q", lines[2])
	}
}

func TestRecordsToTSV_NestedValueJSONEncoded(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"tags": []interface{}{"x", "y"}},
	}
	out := RecordsToTSV(data)
	if !strings.Contains(out, `["x","y"]`) {
		t.Errorf("expected nested array to be JSON-encoded inline, got: %q", out)
	}
}

func TestRecordsToTSV_EscapesTabsAndNewlines(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"msg": "line1\tline2\nline3"},
	}
	out := RecordsToTSV(data)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header + 1 row despite embedded tab/newline, got %d lines: %q", len(lines), out)
	}
	if !strings.Contains(lines[1], `\t`) || !strings.Contains(lines[1], `\n`) {
		t.Errorf("expected escaped \\t and \\n sequences, got: %q", lines[1])
	}
}

func TestSerializeResponseTSV_InlineMode(t *testing.T) {
	response := map[string]interface{}{
		"mode": OutputModeInline,
		"data": []interface{}{
			map[string]interface{}{"tool_name": "Bash"},
		},
	}
	out, ok := SerializeResponseTSV(response)
	if !ok {
		t.Fatal("expected SerializeResponseTSV to handle inline mode")
	}
	if !strings.Contains(out, "tool_name") {
		t.Errorf("expected header with tool_name, got: %q", out)
	}
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Errorf("expected non-JSON TSV output, got JSON-looking output: %q", out)
	}
}

func TestSerializeResponseTSV_FileRefModeFallsBack(t *testing.T) {
	response := map[string]interface{}{
		"mode": OutputModeFileRef,
		"file_ref": map[string]interface{}{
			"path": "/tmp/foo.jsonl",
		},
	}
	_, ok := SerializeResponseTSV(response)
	if ok {
		t.Error("expected SerializeResponseTSV to decline file_ref mode (fall back to JSON)")
	}
}
