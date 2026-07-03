package executor

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/mcp/tools"
)

// mockAnalysisService is a minimal stub that satisfies the analysis.AnalysisService interface.
// It delegates QueryEditSequences to return a static JSON result for testing.
type mockEditSeqAnalysisSvc struct {
	callArgs map[string]interface{}
}

func (m *mockEditSeqAnalysisSvc) AnalyzeBugs(args map[string]interface{}) (string, error) {
	return `{}`, nil
}
func (m *mockEditSeqAnalysisSvc) AnalyzeErrors(args map[string]interface{}) (string, error) {
	return `{}`, nil
}
func (m *mockEditSeqAnalysisSvc) QualityScan(args map[string]interface{}) (string, error) {
	return `{}`, nil
}
func (m *mockEditSeqAnalysisSvc) GetWorkPatterns(args map[string]interface{}) (string, error) {
	return `{}`, nil
}
func (m *mockEditSeqAnalysisSvc) GetTimeline(args map[string]interface{}) (string, error) {
	return `{}`, nil
}
func (m *mockEditSeqAnalysisSvc) GetTechDebt(args map[string]interface{}) (string, error) {
	return `{}`, nil
}
func (m *mockEditSeqAnalysisSvc) QueryEditSequences(args map[string]interface{}) (string, error) {
	m.callArgs = args
	return `{"files":{},"summary":{"totalFiles":0,"patternDistribution":{"A":0,"B":0,"C":0}}}`, nil
}

func TestHandleQueryEditSequences_MissingFiles(t *testing.T) {
	svc := &mockEditSeqAnalysisSvc{}
	exec := &ToolExecutor{AnalysisSvc: svc}

	_, err := exec.ExecuteTool(nil, "query_edit_sequences", map[string]interface{}{
		"files": []interface{}{},
	})
	if err == nil {
		t.Fatal("expected error when files is empty")
	}
}

func TestHandleQueryEditSequences_ValidInput(t *testing.T) {
	svc := &mockEditSeqAnalysisSvc{}
	exec := &ToolExecutor{AnalysisSvc: svc}

	result, err := exec.ExecuteTool(nil, "query_edit_sequences", map[string]interface{}{
		"files": []interface{}{"/some/file.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Validate JSON response contains 'files' key
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if _, ok := decoded["files"]; !ok {
		t.Errorf("expected 'files' key in result: %s", result)
	}
}

func TestGetToolDefinitions_IncludesQueryEditSequences(t *testing.T) {
	defs := tools.GetToolDefinitions()
	found := false
	for _, tool := range defs {
		if tool.Name == "query_edit_sequences" {
			found = true
			// Verify 'files' is in required fields
			requiresFiles := false
			for _, req := range tool.InputSchema.Required {
				if req == "files" {
					requiresFiles = true
					break
				}
			}
			if !requiresFiles {
				t.Errorf("query_edit_sequences should require 'files' parameter")
			}
			break
		}
	}
	if !found {
		t.Error("expected query_edit_sequences in GetToolDefinitions()")
	}
}

func TestHandleQueryEditSequences_NoFilesParam(t *testing.T) {
	svc := &mockEditSeqAnalysisSvc{}
	exec := &ToolExecutor{AnalysisSvc: svc}

	_, err := exec.ExecuteTool(nil, "query_edit_sequences", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error when 'files' param is missing")
	}
	if !strings.Contains(err.Error(), "files") {
		t.Errorf("expected error message to mention 'files', got: %v", err)
	}
}
