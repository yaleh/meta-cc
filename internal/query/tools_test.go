package query

import (
	"os"
	"path/filepath"
	"testing"

	pipelinepkg "github.com/yaleh/meta-cc/pkg/pipeline"
)

func createToolsTestSession(t *testing.T, sessionID, projectHash string, content string) {
	t.Helper()

	projectsRoot := os.Getenv("META_CC_PROJECTS_ROOT")
	if projectsRoot == "" {
		t.Fatal("META_CC_PROJECTS_ROOT must be set for tests")
	}

	sessionDir := filepath.Join(projectsRoot, projectHash)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(sessionDir) })

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write session fixture: %v", err)
	}
}

func TestRunToolsQuery_FiltersAndPagination(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-tools-library"
	projectHash := "-home-yale-work-test-tools-library"
	fixture := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Grep","input":{"pattern":"test"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"match found"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/test.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:03:00.000Z","uuid":"uuid-4","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","is_error":true,"content":"file not found"}]}}
`

	createToolsTestSession(t, sessionID, projectHash, fixture)

	opts := ToolsQueryOptions{
		Pipeline: pipelinepkg.GlobalOptions{
			SessionID: sessionID,
		},
		Status: "error",
		Limit:  1,
	}

	calls, err := RunToolsQuery(opts)
	if err != nil {
		t.Fatalf("RunToolsQuery failed: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call after limit/filter, got %d", len(calls))
	}

	if calls[0].ToolName != "Read" {
		t.Fatalf("expected Read tool, got %s", calls[0].ToolName)
	}
}

func TestRunToolsQuery_FilterByToolName(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-tools-library-by-tool"
	projectHash := "-home-yale-work-test-tools-library-by-tool"
	fixture := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"file.txt"}]}}
`

	createToolsTestSession(t, sessionID, projectHash, fixture)

	opts := ToolsQueryOptions{
		Pipeline: pipelinepkg.GlobalOptions{SessionID: sessionID},
		Tool:     "Bash",
	}

	calls, err := RunToolsQuery(opts)
	if err != nil {
		t.Fatalf("RunToolsQuery failed: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	if calls[0].ToolName != "Bash" {
		t.Fatalf("expected Bash tool, got %s", calls[0].ToolName)
	}
}

func TestRunToolsQuery_WhereLike(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-tools-library-like"
	projectHash := "-home-yale-work-test-tools-library-like"
	fixture := `{"type":"assistant","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"meta-cc-run","input":{"command":"meta"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"ok"}]}}
`

	createToolsTestSession(t, sessionID, projectHash, fixture)

	opts := ToolsQueryOptions{
		Pipeline: pipelinepkg.GlobalOptions{SessionID: sessionID},
		Where:    "tool LIKE 'meta-cc%'",
	}

	calls, err := RunToolsQuery(opts)
	if err != nil {
		t.Fatalf("RunToolsQuery failed: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].ToolName != "meta-cc-run" {
		t.Fatalf("expected meta-cc-run tool, got %s", calls[0].ToolName)
	}
}
