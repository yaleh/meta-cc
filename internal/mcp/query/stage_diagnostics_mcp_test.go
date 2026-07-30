package query

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// AC4: the uniform diagnostics envelope is surfaced through the MCP Stage 2
// handler response, with bounded skip accounting for a partially corrupt
// corpus (degraded mode) — asserted at the tool boundary callers observe.
func TestHandleExecuteStage2Query_DiagnosticsEnvelopeEmitted(t *testing.T) {
	dir := t.TempDir()
	goodFile := filepath.Join(dir, "good.jsonl")
	if err := os.WriteFile(goodFile, []byte(`{"type":"user","timestamp":"2025-01-15T10:00:00Z","message":{"content":"hi"}}`+"\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	missing := filepath.Join(dir, "missing.jsonl")

	result, err := HandleExecuteStage2Query(context.Background(), map[string]interface{}{
		"files":  []interface{}{goodFile, missing},
		"filter": `select(.type == "user")`,
	})
	if err != nil {
		t.Fatalf("expected degraded success, got: %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	diagRaw, ok := m["diagnostics"]
	if !ok {
		t.Fatal("response must include a diagnostics envelope")
	}
	diag, ok := diagRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("diagnostics must be an object, got %T", diagRaw)
	}

	if diag["backend"] != "stage2_jq" {
		t.Errorf("expected backend=stage2_jq, got %v", diag["backend"])
	}
	if diag["files_considered"] != 2 || diag["files_loaded"] != 1 || diag["files_skipped"] != 1 {
		t.Errorf("unexpected file accounting: %v", diag)
	}
	if diag["degraded"] != true {
		t.Errorf("expected degraded=true, got %v", diag["degraded"])
	}
	if diag["matches_returned"] != 1 {
		t.Errorf("expected matches_returned=1, got %v", diag["matches_returned"])
	}
	sw, ok := diag["skip_warnings"].([]string)
	if !ok || len(sw) == 0 {
		t.Errorf("expected bounded skip_warnings, got %T %v", diag["skip_warnings"], diag["skip_warnings"])
	}
	// warnings key still present and is an array (existing contract preserved).
	if _, ok := m["warnings"].([]string); !ok {
		t.Errorf("warnings must remain a string array, got %T", m["warnings"])
	}
}
