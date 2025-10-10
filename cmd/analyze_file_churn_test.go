package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/analyzer"
)

func TestAnalyzeFileChurnCommand_OutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in CI - requires real git history")
	}

	// Test: verify that analyze file-churn outputs array of objects (not wrapped)
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"analyze", "file-churn", "--project", "/home/yale/work/meta-cc", "--threshold", "100", "--output", "jsonl"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Empty output is acceptable (no files meet threshold)
	if output == "" || strings.TrimSpace(output) == "" {
		t.Skip("No output (no files meet threshold) - skipping format validation")
	}

	// Parse output - should be JSONL array format (one object per line)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		if line == "" {
			continue
		}

		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("Line %d is not valid JSON: %v\nLine: %s", i+1, err, line)
		}

		// Check: Should have file-level fields directly (not wrapped in "high_churn_files")
		// Expected fields: file, total_accesses, read_count, edit_count, write_count, time_span_minutes, first_access, last_access
		if _, hasFile := obj["file"]; hasFile {
			// ✅ Correct format: direct file object
			expectedFields := []string{"file", "total_accesses", "read_count", "edit_count", "write_count"}
			for _, field := range expectedFields {
				if _, exists := obj[field]; !exists {
					t.Errorf("Line %d missing expected field '%s': %v", i+1, field, obj)
				}
			}
		} else if _, hasWrapped := obj["high_churn_files"]; hasWrapped {
			// ❌ Wrong format: wrapped in "high_churn_files" object
			t.Errorf("Line %d: Output should be array of file objects, not wrapped object with 'high_churn_files' key\nGot: %s", i+1, line)
		} else {
			t.Errorf("Line %d: Unexpected JSON structure (no 'file' or 'high_churn_files' field): %v", i+1, obj)
		}
	}
}

func TestAnalyzeFileChurnCommand_MatchesSequencesFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in CI - requires real git history")
	}

	// Test: verify file-churn output format matches sequences format (array of objects)
	// Both should output JSONL without wrapper objects

	var fileChurnBuf, sequencesBuf bytes.Buffer

	// Run file-churn
	rootCmd.SetOut(&fileChurnBuf)
	rootCmd.SetErr(&fileChurnBuf)
	rootCmd.SetArgs([]string{"analyze", "file-churn", "--project", "/home/yale/work/meta-cc", "--threshold", "1", "--output", "jsonl"})
	_ = rootCmd.Execute()

	// Run sequences
	rootCmd.SetOut(&sequencesBuf)
	rootCmd.SetErr(&sequencesBuf)
	rootCmd.SetArgs([]string{"analyze", "sequences", "--project", "/home/yale/work/meta-cc", "--min-occurrences", "2", "--output", "jsonl"})
	_ = rootCmd.Execute()

	fileChurnOutput := strings.TrimSpace(fileChurnBuf.String())
	sequencesOutput := strings.TrimSpace(sequencesBuf.String())

	// Both should produce JSONL (one JSON object per line)
	// or empty output if no results

	if fileChurnOutput != "" {
		fileChurnLines := strings.Split(fileChurnOutput, "\n")
		for i, line := range fileChurnLines {
			if line == "" {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				t.Fatalf("file-churn line %d is not valid JSON: %v", i+1, err)
			}

			// Should NOT have wrapper keys like "high_churn_files" or "sequences"
			if _, hasWrapper := obj["high_churn_files"]; hasWrapper {
				t.Errorf("file-churn output should not be wrapped (line %d)", i+1)
			}
		}
	}

	if sequencesOutput != "" {
		sequencesLines := strings.Split(sequencesOutput, "\n")
		for i, line := range sequencesLines {
			if line == "" {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				t.Fatalf("sequences line %d is not valid JSON: %v", i+1, err)
			}

			// Verify sequences is NOT wrapped
			if _, hasWrapper := obj["sequences"]; hasWrapper {
				t.Errorf("sequences output should not be wrapped (line %d)", i+1)
			}
		}
	}
}

func TestFileChurnAnalysis_StructureHasWrapper(t *testing.T) {
	// Test: Document that FileChurnAnalysis has a wrapper field
	// This test documents the current structure (will help track if it changes)

	result := analyzer.FileChurnAnalysis{
		HighChurnFiles: []analyzer.FileChurnDetail{
			{
				File:          "/test/file.go",
				TotalAccesses: 10,
				ReadCount:     5,
				EditCount:     3,
				WriteCount:    2,
			},
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// This documents the current structure: FileChurnAnalysis has "high_churn_files" wrapper
	if _, hasWrapper := obj["high_churn_files"]; !hasWrapper {
		t.Error("FileChurnAnalysis structure changed - no longer has 'high_churn_files' wrapper")
	}

	// Note: This is the structure we need to unwrap before outputting to JSONL
}

func TestAnalyzeFileChurnCommand_Exists(t *testing.T) {
	// Test: analyze file-churn command is registered under analyze
	cmd := analyzeCmd
	found := false
	for _, subcmd := range cmd.Commands() {
		if subcmd.Name() == "file-churn" {
			found = true
			break
		}
	}

	if !found {
		t.Error("analyze file-churn command not found")
	}
}

func TestAnalyzeFileChurnCommand_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"analyze", "file-churn", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Verify help mentions key functionality
	expectedContent := []string{
		"Detect files",
		"--threshold",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Expected '%s' in help output, got: %s", content, output)
		}
	}
}
